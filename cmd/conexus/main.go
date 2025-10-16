package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ferg-cod3s/conexus/internal/config"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/mcp"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
	"go.opentelemetry.io/otel/trace"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const Version = "0.1.0-alpha"

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger (replaces old setupLogging)
	logger := observability.NewLogger(observability.LoggerConfig{
		Level:     cfg.Logging.Level,
		Format:    cfg.Logging.Format,
		Output:    os.Stdout,
		AddSource: true,
	})

	logger.Info("Conexus MCP Server starting",
		"version", Version,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"database", cfg.Database.Path,
		"metrics_enabled", cfg.Observability.Metrics.Enabled,
		"tracing_enabled", cfg.Observability.Tracing.Enabled,
	)

	// Initialize metrics collector if enabled
	var metrics *observability.MetricsCollector
	if cfg.Observability.Metrics.Enabled {
		metrics = observability.NewMetricsCollector("conexus")
		logger.Info("Metrics collection enabled",
			"port", cfg.Observability.Metrics.Port,
			"path", cfg.Observability.Metrics.Path,
		)

		// Start metrics HTTP server on separate port
		go startMetricsServer(ctx, cfg.Observability.Metrics, logger)
	} else {
		logger.Info("Metrics collection disabled")
	}

	// Initialize tracing provider if enabled
	var tracerProvider *observability.TracerProvider
	if cfg.Observability.Tracing.Enabled {
		tracerProvider, err = observability.NewTracerProvider(observability.TracerConfig{
			ServiceName:    "conexus",
			ServiceVersion: Version,
			Environment:    "development", // TODO: Make configurable
			OTLPEndpoint:   cfg.Observability.Tracing.Endpoint,
			SamplingRate:   cfg.Observability.Tracing.SampleRate,
			Enabled:        true,
		})
		if err != nil {
			logger.Error("Failed to initialize tracing provider", "error", err)
			os.Exit(1)
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				logger.Error("Failed to shutdown tracer provider", "error", err)
			}
		}()
		logger.Info("Tracing enabled",
			"endpoint", cfg.Observability.Tracing.Endpoint,
			"sample_rate", cfg.Observability.Tracing.SampleRate,
		)
	} else {
		logger.Info("Tracing disabled")
	}

	// Initialize vector store (SQLite)
	vectorStore, err := sqlite.NewStore(cfg.Database.Path)
	if err != nil {
		logger.Error("Failed to initialize vector store", "error", err)
		os.Exit(1)
	}
	defer vectorStore.Close()

	// Initialize embedder (mock for now - would be real implementation)
	embedder := embedding.NewMock(768) // Standard embedding dimension

	// Check if we're running in HTTP mode (has PORT env or config)
	if cfg.Server.Port > 0 {
		runHTTPServer(ctx, cfg, vectorStore, embedder, logger, metrics, tracerProvider)
	} else {
		// Run in stdio mode (default MCP behavior)
		logger.Info("Running in stdio mode (MCP over stdin/stdout)")
		mcpServer := mcp.NewServer(os.Stdin, os.Stdout, vectorStore, embedder)
		if err := mcpServer.Serve(); err != nil {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}
}

// startMetricsServer starts the Prometheus metrics HTTP server on a separate port.
func startMetricsServer(ctx context.Context, cfg config.MetricsConfig, logger *observability.Logger) {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle(cfg.Path, promhttp.Handler())

	// Health check for metrics server
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","component":"metrics"}`)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Starting metrics server",
		"addr", addr,
		"path", cfg.Path,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Metrics server failed", "error", err)
	}
}

func runHTTPServer(
	ctx context.Context,
	cfg *config.Config,
	vectorStore *sqlite.Store,
	embedder embedding.Embedder,
	logger *observability.Logger,
	metrics *observability.MetricsCollector,
	tracerProvider *observability.TracerProvider,
) {
	// Setup HTTP server with health check
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","version":"%s"}`, Version)
	})

	// MCP endpoint (JSON-RPC over HTTP) with observability
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Start trace span if tracing is enabled
		var span trace.Span
		requestCtx := r.Context()
		if tracerProvider != nil {
			requestCtx, span = tracerProvider.StartSpan(requestCtx, "mcp.request")
			defer span.End()
		}

		// Handle JSON-RPC request/response with observability
		handleJSONRPC(w, r.WithContext(requestCtx), vectorStore, embedder, logger, metrics, tracerProvider)
	})

	// Root info endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"name":"conexus","version":"%s","mcp_endpoint":"/mcp"}`, Version)
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting",
			"addr", addr,
			"health_endpoint", fmt.Sprintf("http://%s/health", addr),
			"mcp_endpoint", fmt.Sprintf("http://%s/mcp", addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server stopped")
}

func handleJSONRPC(
	w http.ResponseWriter,
	r *http.Request,
	vectorStore *sqlite.Store,
	embedder embedding.Embedder,
	logger *observability.Logger,
	metrics *observability.MetricsCollector,
	tracerProvider *observability.TracerProvider,
) {
	ctx := r.Context()
	startTime := time.Now()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed to read request body", "error", err)
		sendJSONRPCError(w, nil, protocol.ParseError, "Failed to read request", nil)
		return
	}
	defer r.Body.Close()

	// Parse JSON-RPC request
	var req protocol.Request
	if err := json.Unmarshal(body, &req); err != nil {
		logger.Error("Invalid JSON in request", "error", err)
		sendJSONRPCError(w, nil, protocol.ParseError, "Invalid JSON", nil)
		return
	}

	// Validate request
	if req.JSONRPC != protocol.JSONRPCVersion {
		logger.Warn("Invalid JSON-RPC version", "version", req.JSONRPC)
		sendJSONRPCError(w, req.ID, protocol.InvalidRequest, "Invalid JSON-RPC version", nil)
		return
	}

	if req.Method == "" {
		logger.Warn("Missing method in request")
		sendJSONRPCError(w, req.ID, protocol.InvalidRequest, "Method required", nil)
		return
	}

	logger.Debug("Handling MCP request", "method", req.Method)

	// Track in-flight requests if metrics enabled
	if metrics != nil {
		metrics.MCPRequestsInFlight.WithLabelValues(req.Method).Inc()
		defer metrics.MCPRequestsInFlight.WithLabelValues(req.Method).Dec()
	}

	// Create MCP handler (using dummy reader/writer since we handle HTTP directly)
	mcpHandler := &mcpHTTPHandler{
		vectorStore:    vectorStore,
		embedder:       embedder,
		logger:         logger,
		metrics:        metrics,
		tracerProvider: tracerProvider,
	}

	// Handle the method
	result, err := mcpHandler.Handle(ctx, req.Method, req.Params)

	// Record metrics
	duration := time.Since(startTime).Seconds()
	if metrics != nil {
		metrics.MCPRequestDuration.WithLabelValues(req.Method).Observe(duration)
		if err != nil {
			metrics.MCPRequestsTotal.WithLabelValues(req.Method, "error").Inc()
			metrics.MCPErrors.WithLabelValues(req.Method, "handler_error").Inc()
		} else {
			metrics.MCPRequestsTotal.WithLabelValues(req.Method, "success").Inc()
		}
	}

	if err != nil {
		logger.Error("Handler error", "method", req.Method, "error", err, "duration_ms", duration*1000)
		// Check if it's a protocol error
		if protoErr, ok := err.(*protocol.Error); ok {
			sendJSONRPCError(w, req.ID, protoErr.Code, protoErr.Message, protoErr.Data)
		} else {
			sendJSONRPCError(w, req.ID, protocol.InternalError, err.Error(), nil)
		}
		return
	}

	logger.Debug("Request handled successfully", "method", req.Method, "duration_ms", duration*1000)

	// Send success response
	sendJSONRPCResult(w, req.ID, result)
}

// mcpHTTPHandler implements protocol.Handler for HTTP transport with observability.
type mcpHTTPHandler struct {
	vectorStore    *sqlite.Store
	embedder       embedding.Embedder
	logger         *observability.Logger
	metrics        *observability.MetricsCollector
	tracerProvider *observability.TracerProvider
}

func (h *mcpHTTPHandler) Handle(ctx context.Context, method string, params json.RawMessage) (interface{}, error) {
	// Start span if tracing enabled
	if h.tracerProvider != nil {
		var span trace.Span
		ctx, span = h.tracerProvider.StartSpan(ctx, fmt.Sprintf("mcp.%s", method))
		defer span.End()
	}

	switch method {
	case "tools/list":
		h.logger.Debug("Listing tools")
		return map[string]interface{}{
			"tools": mcp.GetToolDefinitions(),
		}, nil

	case "tools/call":
		var req mcp.ToolCallRequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: fmt.Sprintf("invalid parameters: %v", err),
			}
		}
		h.logger.Debug("Tool call", "tool", req.Name)
		// For now, return placeholder - full implementation would route to handlers
		return map[string]interface{}{
			"tool":   req.Name,
			"status": "not_implemented",
		}, nil

	case "resources/list":
		h.logger.Debug("Listing resources")
		return map[string]interface{}{
			"resources": []mcp.ResourceDefinition{
				{
					URI:         fmt.Sprintf("%s://%s/", mcp.ResourceScheme, mcp.ResourceFiles),
					Name:        "Indexed Files",
					Description: "Browse indexed project files",
					MimeType:    "application/x-directory",
				},
			},
		}, nil

	case "resources/read":
		var req struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: fmt.Sprintf("invalid parameters: %v", err),
			}
		}
		h.logger.Debug("Reading resource", "uri", req.URI)
		return map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      req.URI,
					"mimeType": "text/plain",
					"text":     "Resource content not yet implemented",
				},
			},
		}, nil

	default:
		h.logger.Warn("Method not found", "method", method)
		return nil, &protocol.Error{
			Code:    protocol.MethodNotFound,
			Message: fmt.Sprintf("method not found: %s", method),
		}
	}
}

func sendJSONRPCResult(w http.ResponseWriter, id interface{}, result interface{}) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		sendJSONRPCError(w, id, protocol.InternalError, "Failed to marshal result", nil)
		return
	}

	resp := protocol.Response{
		JSONRPC: protocol.JSONRPCVersion,
		Result:  resultJSON,
		ID:      id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// #nosec G104 - Error encoding after WriteHeader means broken connection, no recovery possible
	json.NewEncoder(w).Encode(resp)
}

func sendJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	var dataJSON json.RawMessage
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			// Fallback to simple error
			dataJSON = nil
		}
	}

	resp := protocol.Response{
		JSONRPC: protocol.JSONRPCVersion,
		Error: &protocol.Error{
			Code:    code,
			Message: message,
			Data:    dataJSON,
		},
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors still use 200 OK
	// #nosec G104 - Error encoding after WriteHeader means broken connection, no recovery possible
	json.NewEncoder(w).Encode(resp)
}
