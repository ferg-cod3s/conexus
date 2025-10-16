package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// Server implements the MCP protocol server
type Server struct {
	vectorStore vectorstore.VectorStore
	embedder    embedding.Embedder
	jsonrpcSrv  *protocol.Server
}

// NewServer creates a new MCP server
func NewServer(
	reader io.Reader,
	writer io.Writer,
	vectorStore vectorstore.VectorStore,
	embedder embedding.Embedder,
) *Server {
	s := &Server{
		vectorStore: vectorStore,
		embedder:    embedder,
	}
	
	// Create JSON-RPC server with this server as handler
	s.jsonrpcSrv = protocol.NewServer(reader, writer, s)
	
	return s
}

// Handle implements protocol.Handler interface
func (s *Server) Handle(method string, params json.RawMessage) (interface{}, error) {
	ctx := context.Background()
	
	switch method {
	case "tools/list":
		return s.handleToolsList(ctx)
	case "tools/call":
		return s.handleToolsCall(ctx, params)
	case "resources/list":
		return s.handleResourcesList(ctx, params)
	case "resources/read":
		return s.handleResourcesRead(ctx, params)
	default:
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
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid parameters: %v", err),
		}
	}
	
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
		return nil, &protocol.Error{
			Code:    protocol.MethodNotFound,
			Message: fmt.Sprintf("unknown tool: %s", req.Name),
		}
	}
}

// ResourcesListRequest represents a resources/list request
type ResourcesListRequest struct {
	URI string `json:"uri,omitempty"`
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
	
	// For now, return placeholder - will be implemented when indexer provides file listing
	return map[string]interface{}{
		"resources": []ResourceDefinition{
			{
				URI:         fmt.Sprintf("%s://%s/", ResourceScheme, ResourceFiles),
				Name:        "Indexed Files",
				Description: "Browse indexed project files",
				MimeType:    "application/x-directory",
			},
		},
	}, nil
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
	
	// For now, return placeholder - will be implemented when indexer provides file content
	return map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      req.URI,
				"mimeType": "text/plain",
				"text":     "Resource content not yet implemented",
			},
		},
	}, nil
}
