package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/search"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// Server implements the MCP protocol server
type Server struct {
	vectorStore    vectorstore.VectorStore
	rootPath       string
	connectorStore connectors.ConnectorStore
	embedder       embedding.Embedder
	searchCache    *search.SearchCache
	metrics        *observability.MetricsCollector
	errorHandler   *observability.ErrorHandler
	jsonrpcSrv     *protocol.Server
	indexer        indexer.IndexController
}

// NewServer creates a new MCP server
func NewServer(
	reader io.Reader,
	writer io.Writer,
	rootPath string,
	vectorStore vectorstore.VectorStore,
	connectorStore connectors.ConnectorStore,
	embedder embedding.Embedder,
	metrics *observability.MetricsCollector,
	errorHandler *observability.ErrorHandler,
	indexer indexer.IndexController,
) *Server {
	// Initialize search cache (max 100 entries, 5 minute TTL)
	searchCache := search.NewSearchCache(100, 5*time.Minute)

	s := &Server{
		vectorStore:    vectorStore,
		rootPath:       rootPath,
		connectorStore: connectorStore,
		embedder:       embedder,
		searchCache:    searchCache,
		metrics:        metrics,
		errorHandler:   errorHandler,
		indexer:        indexer,
	}

	// Create JSON-RPC server with this server as handler
	s.jsonrpcSrv = protocol.NewServer(reader, writer, s)

	return s
}

// Handle implements protocol.Handler interface
func (s *Server) Handle(method string, params json.RawMessage) (interface{}, error) {
	ctx := context.Background()

	// Add method context for tracing and logging
	ctx = observability.WithRequestContext(ctx, fmt.Sprintf("mcp_%s_%d", method, time.Now().UnixNano()))

	switch method {
	case "initialize":
		return s.handleInitialize(ctx, params)
	case "tools/list":
		return s.handleToolsList(ctx)
	case "tools/call":
		return s.handleToolsCall(ctx, params)
	case "resources/list":
		return s.handleResourcesList(ctx, params)
	case "resources/read":
		return s.handleResourcesRead(ctx, params)
	default:
		errorCtx := observability.ExtractErrorContext(ctx, method)
		errorCtx.ErrorType = "method_not_found"
		errorCtx.ErrorCode = protocol.MethodNotFound

		if s.errorHandler != nil {
			s.errorHandler.HandleError(ctx, fmt.Errorf("method not found: %s", method), errorCtx)
		}

		return nil, &protocol.Error{
			Code:    protocol.MethodNotFound,
			Message: fmt.Sprintf("method not found: %s", method),
		}
	}
}

// Serve starts the MCP server
func (s *Server) Serve() error {
	return s.jsonrpcSrv.Serve()
}

// Close releases resources
func (s *Server) Close() error {
	if s.vectorStore != nil {
		return s.vectorStore.Close()
	}
	return nil
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	// Parse initialize request
	var req map[string]interface{}
	if err := json.Unmarshal(params, &req); err != nil {
		errorCtx := observability.ExtractErrorContext(ctx, "initialize")
		errorCtx.ErrorType = "invalid_params"
		errorCtx.ErrorCode = protocol.InvalidParams
		errorCtx.Params = params

		if s.errorHandler != nil {
			s.errorHandler.HandleError(ctx, err, errorCtx)
		}

		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid parameters: %v", err),
		}
	}

	// Return initialize response with server capabilities
	return map[string]interface{}{
		"protocolVersion": "2025-06-18",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "conexus",
			"version": "0.2.1-alpha",
		},
	}, nil
}

// handleToolsList returns the list of available tools
func (s *Server) handleToolsList(ctx context.Context) (interface{}, error) {
	return map[string]interface{}{
		"tools": GetToolDefinitions(),
	}, nil
}

// ToolCallRequest represents a tool call request
type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// handleToolsCall executes a tool call
func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req ToolCallRequest
	if err := json.Unmarshal(params, &req); err != nil {
		errorCtx := observability.ExtractErrorContext(ctx, "tools/call")
		errorCtx.ErrorType = "invalid_params"
		errorCtx.ErrorCode = protocol.InvalidParams
		errorCtx.Params = params

		if s.errorHandler != nil {
			s.errorHandler.HandleError(ctx, err, errorCtx)
		}

		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid parameters: %v", err),
		}
	}

	// Add tool context for tracing and logging
	ctx = observability.WithToolContext(ctx, req.Name, "1.0.0")

	switch req.Name {
	case ToolContextSearch:
		return s.handleContextSearch(ctx, req.Arguments)
	case ToolContextGetRelatedInfo:
		return s.handleGetRelatedInfo(ctx, req.Arguments)
	case ToolContextIndexControl:
		return s.handleIndexControl(ctx, req.Arguments)
	case ToolContextConnectorManagement:
		return s.handleConnectorManagement(ctx, req.Arguments)
	default:
		errorCtx := observability.ExtractErrorContext(ctx, "tools/call")
		errorCtx.ErrorType = "tool_not_found"
		errorCtx.ErrorCode = protocol.MethodNotFound
		errorCtx.ToolName = req.Name

		if s.errorHandler != nil {
			s.errorHandler.HandleError(ctx, fmt.Errorf("unknown tool: %s", req.Name), errorCtx)
		}

		return nil, &protocol.Error{
			Code:    protocol.MethodNotFound,
			Message: fmt.Sprintf("unknown tool: %s", req.Name),
		}
	}
}

// ResourcesListRequest represents a resources/list request
type ResourcesListRequest struct {
	URI    string `json:"uri,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

// handleResourcesList returns available resources
func (s *Server) handleResourcesList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req ResourcesListRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: fmt.Sprintf("invalid parameters: %v", err),
			}
		}
	}

	// If URI is specified, validate it
	if req.URI != "" {
		if !strings.HasPrefix(req.URI, fmt.Sprintf("%s://", ResourceScheme)) {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "invalid URI scheme",
			}
		}
		// For now, we only support listing all files, not subdirectories
		if req.URI != fmt.Sprintf("%s://%s/", ResourceScheme, ResourceFiles) {
			return map[string]interface{}{
				"resources": []ResourceDefinition{},
			}, nil
		}
	}

	// Get all indexed files from vectorstore
	files, err := s.vectorStore.ListIndexedFiles(ctx)
	if err != nil {
		errorCtx := observability.ExtractErrorContext(ctx, "resources/list")
		errorCtx.ErrorType = "vectorstore_error"
		errorCtx.ErrorCode = protocol.InternalError

		if s.errorHandler != nil {
			s.errorHandler.HandleError(ctx, err, errorCtx)
		}

		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to list indexed files: %v", err),
		}
	}

	// Convert to resource definitions
	resources := make([]ResourceDefinition, 0, len(files)+1)

	// Add the root directory resource
	resources = append(resources, ResourceDefinition{
		URI:         fmt.Sprintf("%s://%s/", ResourceScheme, ResourceFiles),
		Name:        "Indexed Files",
		Description: "Browse indexed project files",
		MimeType:    "application/x-directory",
	})

	// Add individual file resources with pagination
	const pageSize = 50
	startIdx := 0

	// Handle cursor for pagination
	if req.Cursor != "" {
		// Simple cursor implementation: cursor is the index as string
		if cursorIdx, err := strconv.Atoi(req.Cursor); err == nil && cursorIdx >= 0 {
			startIdx = cursorIdx
		}
	}

	// Add individual file resources (paginated)
	endIdx := startIdx + pageSize
	if endIdx > len(files) {
		endIdx = len(files)
	}

	for i := startIdx; i < endIdx; i++ {
		filePath := files[i]

		// Validate path for security
		if err := s.validateFilePath(filePath); err != nil {
			continue // Skip invalid paths
		}

		// Determine MIME type based on file extension
		mimeType := s.getMimeType(filePath)

		resources = append(resources, ResourceDefinition{
			URI:         fmt.Sprintf("%s://file/%s", ResourceScheme, filePath),
			Name:        filepath.Base(filePath),
			Description: fmt.Sprintf("Indexed file: %s", filePath),
			MimeType:    mimeType,
		})
	}

	response := map[string]interface{}{
		"resources": resources,
	}

	// Add next cursor if there are more results
	if endIdx < len(files) {
		response["nextCursor"] = strconv.Itoa(endIdx)
	}

	return response, nil
}

// ResourcesReadRequest represents a resources/read request
type ResourcesReadRequest struct {
	URI string `json:"uri"`
}

// handleResourcesRead returns resource content
func (s *Server) handleResourcesRead(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req ResourcesReadRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid parameters: %v", err),
		}
	}

	// Validate URI format
	if !strings.HasPrefix(req.URI, fmt.Sprintf("%s://file/", ResourceScheme)) {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "invalid URI format, expected engine://file/{path}",
		}
	}

	// Extract file path from URI
	filePath := strings.TrimPrefix(req.URI, fmt.Sprintf("%s://file/", ResourceScheme))

	// Validate path for security
	if err := s.validateFilePath(filePath); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid file path: %v", err),
		}
	}

	// Get all chunks for this file
	chunks, err := s.vectorStore.GetFileChunks(ctx, filePath)
	if err != nil {
		errorCtx := observability.ExtractErrorContext(ctx, "resources/read")
		errorCtx.ErrorType = "vectorstore_error"
		errorCtx.ErrorCode = protocol.InternalError

		if s.errorHandler != nil {
			s.errorHandler.HandleError(ctx, err, errorCtx)
		}

		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get file chunks: %v", err),
		}
	}

	if len(chunks) == 0 {
		return nil, &protocol.Error{
			Code:    protocol.InvalidRequest,
			Message: "file not found or not indexed",
		}
	}

	// Determine MIME type
	mimeType := s.getMimeType(filePath)

	// For single chunk files, return the content directly
	if len(chunks) == 1 {
		return map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      req.URI,
					"mimeType": mimeType,
					"text":     chunks[0].Content,
				},
			},
		}, nil
	}

	// For multi-chunk files, return each chunk with line ranges
	contents := make([]map[string]interface{}, 0, len(chunks))
	for _, chunk := range chunks {
		metadata := chunk.Metadata
		startLine, _ := metadata["start_line"].(float64)
		endLine, _ := metadata["end_line"].(float64)

		content := map[string]interface{}{
			"uri":      req.URI,
			"mimeType": mimeType,
			"text":     chunk.Content,
		}

		// Add line range information for multi-chunk files
		if startLine > 0 && endLine > 0 {
			content["startLineNumber"] = int(startLine)
			content["endLineNumber"] = int(endLine)
		}

		contents = append(contents, content)
	}

	return map[string]interface{}{
		"contents": contents,
	}, nil
}

// validateFilePath validates that a file path is safe and doesn't contain traversal attempts
func (s *Server) validateFilePath(filePath string) error {
	// Check for empty path
	if filePath == "" {
		return fmt.Errorf("empty file path")
	}

	// Check for path traversal attempts
	if strings.Contains(filePath, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Check for absolute paths
	if filepath.IsAbs(filePath) {
		return fmt.Errorf("absolute paths not allowed")
	}

	// Additional validation can be added here as needed
	return nil
}

// getMimeType determines the MIME type based on file extension
func (s *Server) getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	// Text/Code files
	case ".go":
		return "text/x-go"
	case ".js", ".jsx":
		return "application/javascript"
	case ".ts", ".tsx":
		return "application/typescript"
	case ".py":
		return "text/x-python"
	case ".rs":
		return "text/x-rust"
	case ".java":
		return "text/x-java"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "text/x-c++"
	case ".c":
		return "text/x-c"
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"
	case ".xml":
		return "application/xml"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	// Image files
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	// Document files
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		return "application/vnd.ms-powerpoint"
	// Archive files
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	// Default for unknown files
	default:
		return "application/octet-stream"
	}
}
