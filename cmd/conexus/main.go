package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ferg-cod3s/conexus/internal/config"
	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/mcp"
	"github.com/ferg-cod3s/conexus/internal/mcp/webhooks"
	"github.com/ferg-cod3s/conexus/internal/middleware"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/security"
	"github.com/ferg-cod3s/conexus/internal/security/auth"
	"github.com/ferg-cod3s/conexus/internal/security/ratelimit"
	"github.com/ferg-cod3s/conexus/internal/tls"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"
)

const Version = "0.1.3-alpha"

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger (replaces old setupLogging)
	// In stdio mode (MCP), logs must go to stderr to avoid interfering with JSON-RPC
	logOutput := os.Stdout
	if os.Getenv("CONEXUS_PORT") == "" || cfg.Server.Port == 0 {
		logOutput = os.Stderr
	}
	logger := observability.NewLogger(observability.LoggerConfig{
		Level:         cfg.Logging.Level,
		Format:        cfg.Logging.Format,
		Output:        logOutput,
		AddSource:     true,
		SentryEnabled: cfg.Observability.Sentry.Enabled,
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

	// Initialize Sentry if enabled
	if cfg.Observability.Sentry.Enabled {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Observability.Sentry.DSN,
			Environment:      cfg.Observability.Sentry.Environment,
			Release:          cfg.Observability.Sentry.Release,
			TracesSampleRate: cfg.Observability.Sentry.SampleRate,
			EnableTracing:    true,
			EnableLogs:       true,
		})
		if err != nil {
			logger.Error("Failed to initialize Sentry", "error", err)
			os.Exit(1)
		}
		defer sentry.Flush(2 * time.Second)
		logger.Info("Sentry enabled",
			"environment", cfg.Observability.Sentry.Environment,
			"sample_rate", cfg.Observability.Sentry.SampleRate,
		)
	} else {
		logger.Info("Sentry disabled")
	}

	// Initialize vector store (SQLite)
	vectorStore, err := sqlite.NewStore(cfg.Database.Path)
	if err != nil {
		logger.Error("Failed to initialize vector store", "error", err)
		os.Exit(1)
	}
	defer vectorStore.Close()

	// Initialize connector store (SQLite)
	connectorStore, err := connectors.NewStore(cfg.Database.Path)
	if err != nil {
		logger.Error("Failed to initialize connector store", "error", err)
		os.Exit(1)
	}
	defer connectorStore.Close()

	// Initialize embedder from configuration
	provider, err := embedding.Get(cfg.Embedding.Provider)
	if err != nil {
		logger.Error("Failed to get embedding provider", "provider", cfg.Embedding.Provider, "error", err)
		os.Exit(1)
	}

	// Prepare provider configuration
	providerConfig := make(map[string]interface{})
	for k, v := range cfg.Embedding.Config {
		providerConfig[k] = v
	}

	// Add common config fields
	providerConfig["model"] = cfg.Embedding.Model
	providerConfig["dimensions"] = cfg.Embedding.Dimensions

	// Create embedder instance
	embedder, err := provider.Create(providerConfig)
	if err != nil {
		logger.Error("Failed to create embedder", "provider", cfg.Embedding.Provider, "error", err)
		os.Exit(1)
	}

	logger.Info("Embedder initialized",
		"provider", cfg.Embedding.Provider,
		"model", embedder.Model(),
		"dimensions", embedder.Dimensions(),
	)

	// Initialize indexer controller
	idx := indexer.NewIndexController("./data/indexer_state.json")

	// Initialize error handler
	errorHandler := observability.NewErrorHandler(logger, metrics, cfg.Observability.Sentry.Enabled)

	// Check if we're running in HTTP mode (explicit CONEXUS_PORT env var)
	// Default is stdio mode for MCP compatibility
	if os.Getenv("CONEXUS_PORT") != "" && cfg.Server.Port > 0 {
		runHTTPServer(ctx, cfg, vectorStore, connectorStore, embedder, logger, metrics, tracerProvider, idx)
	} else {
		// Run in stdio mode (default MCP behavior)
		logger.Info("Running in stdio mode (MCP over stdin/stdout)")
		mcpServer := mcp.NewServer(os.Stdin, os.Stdout, cfg.Indexer.RootPath, vectorStore, connectorStore, embedder, metrics, errorHandler, idx)
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
	connectorStore connectors.ConnectorStore,
	embedder embedding.Embedder,
	logger *observability.Logger,
	metrics *observability.MetricsCollector,
	tracerProvider *observability.TracerProvider,
	idx indexer.IndexController,
) {
	// Initialize TLS manager if TLS is enabled
	var tlsManager *tls.Manager
	if cfg.TLS.Enabled {
		var err error
		tlsManager, err = tls.NewManager(&cfg.TLS, logger)
		if err != nil {
			logger.Error("Failed to initialize TLS manager", "error", err)
			os.Exit(1)
		}

		// Validate certificates
		if err := tlsManager.ValidateCertificates(); err != nil {
			logger.Error("Certificate validation failed", "error", err)
			os.Exit(1)
		}

		logger.Info("TLS enabled",
			"auto_cert", cfg.TLS.AutoCert,
			"min_version", cfg.TLS.MinVersion)
	}
	// Setup HTTP server with health check
	mux := http.NewServeMux()

	// Initialize webhook handler
	errorHandler := observability.NewErrorHandler(logger, metrics, false)
	webhookHandler := webhooks.NewWebhookHandler(connectorStore, embedder, vectorStore, errorHandler)

	// Initialize JWT manager if authentication is enabled
	var jwtManager *auth.JWTManager
	var authMiddleware *middleware.AuthMiddleware
	if cfg.Auth.Enabled {
		var err error
		jwtManager, err = auth.NewJWTManager(
			cfg.Auth.PrivateKey,
			cfg.Auth.PublicKey,
			cfg.Auth.Issuer,
			cfg.Auth.Audience,
			cfg.Auth.TokenExpiry,
		)
		if err != nil {
			logger.Error("Failed to initialize JWT manager", "error", err)
			os.Exit(1)
		}
		authMiddleware = middleware.NewAuthMiddleware(jwtManager)
		logger.Info("JWT authentication enabled",
			"issuer", cfg.Auth.Issuer,
			"audience", cfg.Auth.Audience,
			"token_expiry_minutes", cfg.Auth.TokenExpiry,
		)
	} else {
		logger.Info("JWT authentication disabled")
	}

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
		handleJSONRPC(w, r.WithContext(requestCtx), vectorStore, connectorStore, embedder, logger, metrics, tracerProvider, idx, cfg.Indexer.RootPath)
	})

	// GitHub webhook endpoint
	mux.HandleFunc("/webhooks/github", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Start trace span if tracing is enabled
		var span trace.Span
		requestCtx := r.Context()
		if tracerProvider != nil {
			requestCtx, span = tracerProvider.StartSpan(requestCtx, "webhook.github")
			defer span.End()
		}

		// Handle GitHub webhook
		webhookHandler.HandleGitHubWebhook(w, r.WithContext(requestCtx))
	})

	// Root info endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"name":"conexus","version":"%s","mcp_endpoint":"/mcp","webhook_endpoint":"/webhooks/github"}`, Version)
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	// Initialize rate limiting middleware if enabled
	var rateLimitMiddleware *middleware.RateLimitMiddleware
	if cfg.RateLimit.Enabled {
		// Helper function to parse duration strings
		parseDuration := func(s string) time.Duration {
			if s == "" {
				return time.Minute // default to 1 minute
			}
			d, err := time.ParseDuration(s)
			if err != nil {
				logger.Warn("Invalid duration string, using default", "duration", s, "error", err)
				return time.Minute
			}
			return d
		}

		// Convert config to ratelimit package format
		rateLimitConfig := ratelimit.Config{
			Enabled: cfg.RateLimit.Enabled,
			Algorithm: func() ratelimit.Algorithm {
				switch cfg.RateLimit.Algorithm {
				case "token_bucket":
					return ratelimit.TokenBucket
				case "sliding_window":
					return ratelimit.SlidingWindow
				default:
					return ratelimit.SlidingWindow // default
				}
			}(),
			Redis: ratelimit.RedisConfig{
				Enabled:   cfg.RateLimit.Redis.Enabled,
				Addr:      cfg.RateLimit.Redis.Addr,
				Password:  cfg.RateLimit.Redis.Password,
				DB:        cfg.RateLimit.Redis.DB,
				KeyPrefix: cfg.RateLimit.Redis.KeyPrefix,
			},
			Default: ratelimit.LimitConfig{
				Requests: cfg.RateLimit.Default.Requests,
				Window:   parseDuration(cfg.RateLimit.Default.Window),
			},
			Health: ratelimit.LimitConfig{
				Requests: cfg.RateLimit.Health.Requests,
				Window:   parseDuration(cfg.RateLimit.Health.Window),
			},
			Webhook: ratelimit.LimitConfig{
				Requests: cfg.RateLimit.Webhook.Requests,
				Window:   parseDuration(cfg.RateLimit.Webhook.Window),
			},
			Auth: ratelimit.LimitConfig{
				Requests: cfg.RateLimit.Auth.Requests,
				Window:   parseDuration(cfg.RateLimit.Auth.Window),
			},
			BurstMultiplier: cfg.RateLimit.BurstMultiplier,
			CleanupInterval: parseDuration(cfg.RateLimit.CleanupInterval),
		}

		// Initialize rate limiter
		rateLimiter, err := ratelimit.NewRateLimiter(rateLimitConfig)
		if err != nil {
			logger.Error("Failed to initialize rate limiter", "error", err)
			os.Exit(1)
		}

		// Initialize rate limiting middleware
		rateLimitMiddleware = middleware.NewRateLimitMiddleware(middleware.RateLimitConfig{
			RateLimiter:      rateLimiter,
			MetricsCollector: metrics,
			SkipPaths:        cfg.RateLimit.SkipPaths,
			SkipIPs:          cfg.RateLimit.SkipIPs,
			TrustedProxies:   cfg.RateLimit.TrustedProxies,
		}, logger)

		logger.Info("Rate limiting enabled",
			"algorithm", cfg.RateLimit.Algorithm,
			"redis_enabled", cfg.RateLimit.Redis.Enabled,
			"default_requests", cfg.RateLimit.Default.Requests,
			"default_window", cfg.RateLimit.Default.Window,
		)
	} else {
		logger.Info("Rate limiting disabled")
	}

	// Initialize security middleware
	xContentTypeOptions := ""
	if cfg.Security.ContentType {
		xContentTypeOptions = "nosniff"
	}

	securityMiddleware := middleware.NewSecurityMiddleware(middleware.SecurityConfig{
		CSP: middleware.CSPConfig{
			Enabled: cfg.Security.CSP.Enabled,
			Default: cfg.Security.CSP.DefaultSrc,
			Script:  cfg.Security.CSP.ScriptSrc,
			Style:   cfg.Security.CSP.StyleSrc,
			Image:   cfg.Security.CSP.ImgSrc,
			Font:    cfg.Security.CSP.FontSrc,
			Connect: cfg.Security.CSP.ConnectSrc,
			Media:   cfg.Security.CSP.MediaSrc,
			Object:  cfg.Security.CSP.ObjectSrc,
			Frame:   cfg.Security.CSP.FrameSrc,
			Report:  cfg.Security.CSP.ReportURI,
		},
		HSTS: middleware.HSTSConfig{
			Enabled:           cfg.Security.HSTS.Enabled,
			MaxAge:            cfg.Security.HSTS.MaxAge,
			IncludeSubdomains: cfg.Security.HSTS.IncludeSubdomains,
			Preload:           cfg.Security.HSTS.Preload,
		},
		XFrameOptions:       cfg.Security.FrameOptions,
		XContentTypeOptions: xContentTypeOptions,
		ReferrerPolicy:      cfg.Security.ReferrerPolicy,
	}, logger)

	// Initialize CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware(middleware.CORSConfig{
		Enabled:          cfg.CORS.Enabled,
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}, logger)

	// Apply middleware in correct order: rate limiting first, then CORS, then security headers, then auth
	var handler http.Handler = mux
	if rateLimitMiddleware != nil {
		handler = rateLimitMiddleware.Middleware(handler)
	}
	handler = corsMiddleware.Middleware(handler)
	handler = securityMiddleware.Middleware(handler)
	if authMiddleware != nil {
		handler = authMiddleware.Middleware(handler)
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Configure TLS if enabled
	if tlsManager != nil {
		server.TLSConfig = tlsManager.GetTLSConfig()
		logger.Info("HTTPS server configured with TLS")
	}

	// Start HTTP redirect server if TLS is enabled
	if tlsManager != nil {
		httpsPort := cfg.Server.Port
		if httpsPort == 443 {
			httpsPort = 0 // Use default HTTPS port
		}
		if err := tlsManager.StartHTTPRedirect(ctx, httpsPort); err != nil {
			logger.Error("Failed to start HTTP redirect server", "error", err)
			os.Exit(1)
		}
	}

	// Start server in goroutine
	go func() {
		scheme := "http"
		if tlsManager != nil {
			scheme = "https"
		}
		logger.Info("Server starting",
			"scheme", scheme,
			"addr", addr,
			"health_endpoint", fmt.Sprintf("%s://%s/health", scheme, addr),
			"mcp_endpoint", fmt.Sprintf("%s://%s/mcp", scheme, addr),
		)

		var err error
		if tlsManager != nil {
			// Use auto-cert or manual certs
			if cfg.TLS.AutoCert {
				err = server.ListenAndServeTLS("", "")
			} else {
				err = server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
			}
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
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
	connectorStore connectors.ConnectorStore,
	embedder embedding.Embedder,
	logger *observability.Logger,
	metrics *observability.MetricsCollector,
	tracerProvider *observability.TracerProvider,
	idx indexer.IndexController,
	rootPath string,
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
		indexer:        idx,
		vectorStore:    vectorStore,
		connectorStore: connectorStore,
		embedder:       embedder,
		logger:         logger,
		metrics:        metrics,
		tracerProvider: tracerProvider,
		rootPath:       rootPath,
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
	indexer        indexer.IndexController
	vectorStore    *sqlite.Store
	connectorStore connectors.ConnectorStore
	embedder       embedding.Embedder
	logger         *observability.Logger
	metrics        *observability.MetricsCollector
	tracerProvider *observability.TracerProvider
	rootPath       string
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

		// Route to appropriate handler
		switch req.Name {
		case mcp.ToolContextSearch:
			return h.handleContextSearch(ctx, req.Arguments)
		case mcp.ToolContextGetRelatedInfo:
			return h.handleGetRelatedInfo(ctx, req.Arguments)
		case mcp.ToolContextIndexControl:
			return h.handleIndexControl(ctx, req.Arguments)
		case mcp.ToolContextConnectorManagement:
			return h.handleConnectorManagement(ctx, req.Arguments)
		default:
			return nil, &protocol.Error{
				Code:    protocol.MethodNotFound,
				Message: fmt.Sprintf("unknown tool: %s", req.Name),
			}
		}

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

func (h *mcpHTTPHandler) handleContextSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req mcp.SearchRequest
	startTime := time.Now()

	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid search request: %v", err),
		}
	}

	// Validate required fields
	if req.Query == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "query is required",
		}
	}

	// Set defaults
	topK := req.TopK
	if topK <= 0 {
		topK = 20
	}
	if topK > 100 {
		topK = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Generate query embedding
	queryVec, err := h.embedder.Embed(ctx, req.Query)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to generate query embedding: %v", err),
		}
	}

	// Prepare search options
	opts := vectorstore.SearchOptions{
		Limit:   topK,
		Offset:  offset,
		Filters: make(map[string]interface{}),
	}

	// Apply filters
	if req.Filters != nil {
		if len(req.Filters.SourceTypes) > 0 {
			opts.Filters["source_types"] = req.Filters.SourceTypes
		}
		if req.Filters.DateRange != nil {
			opts.Filters["date_range"] = map[string]string{
				"from": req.Filters.DateRange.From,
				"to":   req.Filters.DateRange.To,
			}
		}
		// Apply work context filters
		if req.Filters.WorkContext != nil {
			if req.Filters.WorkContext.ActiveFile != "" {
				opts.Filters["related_files"] = req.Filters.WorkContext.ActiveFile
			}
			if req.Filters.WorkContext.GitBranch != "" {
				opts.Filters["git_branch"] = req.Filters.WorkContext.GitBranch
			}
			if len(req.Filters.WorkContext.OpenTicketIDs) > 0 {
				opts.Filters["ticket_ids"] = req.Filters.WorkContext.OpenTicketIDs
			}
		}
	}

	// Apply work context from request (overrides filter)
	if req.WorkContext != nil {
		if req.WorkContext.ActiveFile != "" {
			opts.Filters["boost_file"] = req.WorkContext.ActiveFile
		}
		if req.WorkContext.GitBranch != "" {
			opts.Filters["git_branch"] = req.WorkContext.GitBranch
		}
		if len(req.WorkContext.OpenTicketIDs) > 0 {
			opts.Filters["boost_tickets"] = req.WorkContext.OpenTicketIDs
		}
	}

	// Perform hybrid search (combines vector + BM25)
	results, searchErr := h.vectorStore.SearchHybrid(ctx, req.Query, queryVec.Vector, opts)
	if searchErr != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("search failed: %v", searchErr),
		}
	}

	queryTime := float64(time.Since(startTime).Milliseconds())

	// Get total count for pagination
	totalCount, countErr := h.vectorStore.Count(ctx)
	if countErr != nil {
		// Log error but don't fail the request
		totalCount = int64(len(results))
	}

	// Convert results to response format
	searchResults := make([]mcp.SearchResultItem, 0, len(results))
	for _, r := range results {
		// Extract source type from metadata
		sourceType := "file" // default
		if st, ok := r.Document.Metadata["source_type"].(string); ok {
			sourceType = st
		}

		searchResults = append(searchResults, mcp.SearchResultItem{
			ID:         r.Document.ID,
			Content:    r.Document.Content,
			Score:      r.Score,
			SourceType: sourceType,
			Metadata:   r.Document.Metadata,
		})
	}

	return mcp.SearchResponse{
		Results:    searchResults,
		TotalCount: len(searchResults),
		QueryTime:  queryTime,
		Offset:     offset,
		Limit:      topK,
		HasMore:    int64(offset+len(results)) < totalCount,
	}, nil
}

func (h *mcpHTTPHandler) handleGetRelatedInfo(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req mcp.GetRelatedInfoRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate that at least one identifier is provided
	if req.FilePath == "" && req.TicketID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "either file_path or ticket_id must be provided",
		}
	}

	// Validate file path if provided
	if req.FilePath != "" {
		cleanedPath, err := security.ValidatePath(req.FilePath, "")
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: fmt.Sprintf("invalid file path: %v", err),
			}
		}
		req.FilePath = cleanedPath
	}

	// Build search query and filters based on provided identifiers
	var query string
	opts := vectorstore.SearchOptions{
		Limit:   20,
		Filters: make(map[string]interface{}),
	}

	if req.FilePath != "" {
		// File path flow: search for chunks from the same file or related files
		query = fmt.Sprintf("content related to %s", filepath.Base(req.FilePath))
		opts.Filters["file_path"] = req.FilePath // Exact file matches
	} else {
		// Ticket ID flow: search for ticket-related content
		query = fmt.Sprintf("ticket:%s", req.TicketID)
		opts.Filters["ticket_ids"] = []string{req.TicketID}
	}

	// Generate query embedding
	queryVec, err := h.embedder.Embed(ctx, query)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to generate query embedding: %v", err),
		}
	}

	// Perform hybrid search
	results, err := h.vectorStore.SearchHybrid(ctx, query, queryVec.Vector, opts)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("search failed: %v", err),
		}
	}

	// Convert results to response format
	relatedItems := make([]mcp.RelatedItem, 0, len(results))
	var relatedPRs, relatedIssues []string
	var discussions []mcp.DiscussionSummary

	for _, r := range results {
		// Add to related items
		item := mcp.RelatedItem{
			ID:       r.Document.ID,
			Content:  r.Document.Content,
			Score:    r.Score,
			Metadata: r.Document.Metadata,
		}

		// Extract additional fields from metadata
		if sourceType, ok := r.Document.Metadata["source_type"].(string); ok {
			item.SourceType = sourceType
		}
		if filePath, ok := r.Document.Metadata["file_path"].(string); ok {
			item.FilePath = filePath
		}
		if startLine, ok := r.Document.Metadata["start_line"].(float64); ok {
			item.StartLine = int(startLine)
		}
		if endLine, ok := r.Document.Metadata["end_line"].(float64); ok {
			item.EndLine = int(endLine)
		}

		relatedItems = append(relatedItems, item)

		// Categorize by source type for backward compatibility
		sourceType := item.SourceType
		switch sourceType {
		case "github_pr":
			if prNum, ok := r.Document.Metadata["pr_number"].(string); ok {
				relatedPRs = append(relatedPRs, prNum)
			}
		case "github_issue", "jira":
			if issueID, ok := r.Document.Metadata["issue_id"].(string); ok {
				relatedIssues = append(relatedIssues, issueID)
			}
		case "slack":
			channel, _ := r.Document.Metadata["channel"].(string)
			timestamp, _ := r.Document.Metadata["timestamp"].(string)
			discussions = append(discussions, mcp.DiscussionSummary{
				Channel:   channel,
				Timestamp: timestamp,
				Summary:   truncateString(r.Document.Content, 200),
			})
		}
	}

	// Generate summary
	summary := fmt.Sprintf("Found %d related items", len(relatedItems))
	if req.FilePath != "" {
		summary = fmt.Sprintf("Related information for %s: %d items (%d PRs, %d issues, %d discussions)",
			req.FilePath, len(relatedItems), len(relatedPRs), len(relatedIssues), len(discussions))
	} else {
		summary = fmt.Sprintf("Related information for ticket %s: %d items (%d PRs, %d issues, %d discussions)",
			req.TicketID, len(relatedItems), len(relatedPRs), len(relatedIssues), len(discussions))
	}

	return mcp.GetRelatedInfoResponse{
		Summary:       summary,
		RelatedItems:  relatedItems,
		RelatedPRs:    relatedPRs,
		RelatedIssues: relatedIssues,
		Discussions:   discussions,
	}, nil
}

// truncateString truncates a string to the specified maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (h *mcpHTTPHandler) handleIndexControl(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req mcp.IndexControlRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate action
	validActions := map[string]bool{
		"start":         true,
		"stop":          true,
		"status":        true,
		"force_reindex": true,
		"reindex_paths": true,
	}

	if !validActions[req.Action] {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid action: %s", req.Action),
		}
	}

	// Check if indexer is available
	if h.indexer == nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: "index controller not available",
		}
	}

	switch req.Action {
	case "status":
		// Get document count
		count, err := h.vectorStore.Count(ctx)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get document count: %v", err),
			}
		}

		// Get indexer status
		idxStatus := h.indexer.GetStatus()

		// Convert to response format
		var startTime, estimatedEnd string
		if !idxStatus.StartTime.IsZero() {
			startTime = idxStatus.StartTime.Format(time.RFC3339)
		}
		if !idxStatus.EstimatedEnd.IsZero() {
			estimatedEnd = idxStatus.EstimatedEnd.Format(time.RFC3339)
		}

		var metrics *mcp.IndexMetrics
		if idxStatus.Metrics.TotalFiles > 0 {
			metrics = &mcp.IndexMetrics{
				TotalFiles:      idxStatus.Metrics.TotalFiles,
				IndexedFiles:    idxStatus.Metrics.IndexedFiles,
				SkippedFiles:    idxStatus.Metrics.SkippedFiles,
				TotalChunks:     idxStatus.Metrics.TotalChunks,
				Duration:        idxStatus.Metrics.Duration.Seconds(),
				BytesProcessed:  idxStatus.Metrics.BytesProcessed,
				StateSize:       idxStatus.Metrics.StateSize,
				IncrementalSave: idxStatus.Metrics.IncrementalSave.Seconds(),
			}
		}

		details := map[string]interface{}{}
		details["documents_indexed"] = count
		details["indexer_available"] = true

		return mcp.IndexControlResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Index contains %d documents", count),
			Details: details,
			IndexStatus: &mcp.IndexStatus{
				IsIndexing:     idxStatus.IsIndexing,
				Phase:          idxStatus.Phase,
				Progress:       idxStatus.Progress,
				FilesProcessed: idxStatus.FilesProcessed,
				TotalFiles:     idxStatus.TotalFiles,
				ChunksCreated:  idxStatus.ChunksCreated,
				StartTime:      startTime,
				EstimatedEnd:   estimatedEnd,
				LastError:      idxStatus.LastError,
				Metrics:        metrics,
			},
		}, nil

	case "start":
		// Use configured root path
		rootPath := h.rootPath
		if rootPath == "" {
			rootPath = "."
		}

		// Load ignore patterns
		ignorePatterns := []string{".git"}
		if gitignore, err := indexer.LoadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
			ignorePatterns = append(ignorePatterns, gitignore...)
		}

		opts := indexer.IndexOptions{
			RootPath:       rootPath,
			IgnorePatterns: ignorePatterns,
			MaxFileSize:    1024 * 1024, // 1MB
			IncludeGitInfo: true,
			Embedder:       h.embedder,
			VectorStore:    h.vectorStore,
		}

		if err := h.indexer.Start(ctx, opts); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to start indexing: %v", err),
			}
		}

		return mcp.IndexControlResponse{
			Status:  "ok",
			Message: "Background indexing started",
		}, nil

	case "stop":
		if err := h.indexer.Stop(ctx); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to stop indexing: %v", err),
			}
		}

		return mcp.IndexControlResponse{
			Status:  "ok",
			Message: "Indexing stopped",
		}, nil

	case "force_reindex":
		// Use configured root path
		rootPath := h.rootPath
		if rootPath == "" {
			rootPath = "."
		}

		// Load ignore patterns
		ignorePatterns := []string{".git"}
		if gitignore, err := indexer.LoadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
			ignorePatterns = append(ignorePatterns, gitignore...)
		}

		opts := indexer.IndexOptions{
			RootPath:       rootPath,
			IgnorePatterns: ignorePatterns,
			MaxFileSize:    1024 * 1024, // 1MB
			IncludeGitInfo: true,
			Embedder:       h.embedder,
			VectorStore:    h.vectorStore,
		}

		if err := h.indexer.ForceReindex(ctx, opts); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to start force reindex: %v", err),
			}
		}

		return mcp.IndexControlResponse{
			Status:  "ok",
			Message: "Force reindex started",
		}, nil

	case "reindex_paths":
		if len(req.Paths) == 0 {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "paths are required for reindex_paths action",
			}
		}

		// Use configured root path
		rootPath := h.rootPath
		if rootPath == "" {
			rootPath = "."
		}

		// Load ignore patterns
		ignorePatterns := []string{".git"}
		if gitignore, err := indexer.LoadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
			ignorePatterns = append(ignorePatterns, gitignore...)
		}

		opts := indexer.IndexOptions{
			RootPath:       rootPath,
			IgnorePatterns: ignorePatterns,
			MaxFileSize:    1024 * 1024, // 1MB
			IncludeGitInfo: true,
			Embedder:       h.embedder,
			VectorStore:    h.vectorStore,
		}

		if err := h.indexer.ReindexPaths(ctx, opts, req.Paths); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to start selective reindex: %v", err),
			}
		}

		return mcp.IndexControlResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Reindexing %d paths", len(req.Paths)),
		}, nil

	default:
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("unimplemented action: %s", req.Action),
		}
	}
}

func (h *mcpHTTPHandler) handleConnectorManagement(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req mcp.ConnectorManagementRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate action
	validActions := map[string]bool{
		"list":   true,
		"add":    true,
		"update": true,
		"remove": true,
	}

	if !validActions[req.Action] {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid action: %s", req.Action),
		}
	}

	switch req.Action {
	case "list":
		connectors, err := h.connectorStore.List(ctx)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to list connectors: %v", err),
			}
		}

		connectorInfos := make([]mcp.ConnectorInfo, len(connectors))
		for i, conn := range connectors {
			connectorInfos[i] = mcp.ConnectorInfo{
				ID:     conn.ID,
				Type:   conn.Type,
				Name:   conn.Name,
				Status: conn.Status,
				Config: conn.Config,
			}
		}

		return mcp.ConnectorManagementResponse{
			Connectors: connectorInfos,
			Status:     "ok",
			Message:    "Retrieved connector list",
		}, nil

	case "add":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		connector := &connectors.Connector{
			ID:     req.ConnectorID,
			Name:   req.ConnectorID, // Default name to ID, can be updated later
			Type:   "filesystem",    // Default type, should be specified in config
			Config: req.ConnectorConfig,
			Status: "active",
		}

		// Extract type and name from config if provided
		if configType, ok := req.ConnectorConfig["type"].(string); ok {
			connector.Type = configType
		}
		if configName, ok := req.ConnectorConfig["name"].(string); ok {
			connector.Name = configName
		}

		if err := h.connectorStore.Add(ctx, connector); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to add connector: %v", err),
			}
		}

		return mcp.ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s added successfully", req.ConnectorID),
		}, nil

	case "update":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		connector := &connectors.Connector{
			Type:   "filesystem", // Default type
			Config: req.ConnectorConfig,
			Status: "active",
		}

		// Extract type and name from config if provided
		if configType, ok := req.ConnectorConfig["type"].(string); ok {
			connector.Type = configType
		}
		if configName, ok := req.ConnectorConfig["name"].(string); ok {
			connector.Name = configName
		}

		if err := h.connectorStore.Update(ctx, req.ConnectorID, connector); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to update connector: %v", err),
			}
		}

		return mcp.ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s updated successfully", req.ConnectorID),
		}, nil

	case "remove":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		if err := h.connectorStore.Remove(ctx, req.ConnectorID); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to remove connector: %v", err),
			}
		}

		return mcp.ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s removed successfully", req.ConnectorID),
		}, nil

	default:
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: "unexpected error",
		}
	}
}
