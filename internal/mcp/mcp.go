// Package mcp implements the Model Context Protocol server for exposing Conexus to LLM agents.
package mcp

import (
	"context"
	"io"
)

// Server is an MCP protocol server that exposes tools and resources.
type Server interface {
	// Serve starts the MCP server (blocking).
	Serve(ctx context.Context) error
	
	// RegisterTool adds a tool handler.
	RegisterTool(tool Tool) error
	
	// RegisterResource adds a resource handler.
	RegisterResource(resource Resource) error
}

// Tool represents an MCP tool (function/action).
type Tool interface {
	// Name returns the tool identifier (e.g., "context.search").
	Name() string
	
	// Description returns a human-readable description for LLM agents.
	Description() string
	
	// Schema returns the JSON schema for tool parameters.
	Schema() map[string]interface{}
	
	// Execute runs the tool with the given parameters.
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// Resource represents an MCP resource (read-only data source).
type Resource interface {
	// URI returns the resource URI pattern (e.g., "codebase://{path}").
	URI() string
	
	// Description returns a human-readable description.
	Description() string
	
	// MimeType returns the MIME type of the resource content.
	MimeType() string
	
	// Read retrieves the resource content.
	Read(ctx context.Context, uri string) ([]byte, error)
}

// Transport abstracts the MCP message transport layer.
type Transport interface {
	// Read reads a JSON-RPC message from the transport.
	Read() ([]byte, error)
	
	// Write sends a JSON-RPC message to the transport.
	Write(message []byte) error
	
	// Close closes the transport.
	Close() error
}

// StdioTransport implements Transport over stdin/stdout.
type StdioTransport struct {
	In  io.Reader
	Out io.Writer
}

// Message represents a JSON-RPC 2.0 message.
type Message struct {
	JSONRPC string                 `json:"jsonrpc"` // Always "2.0"
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *ErrorObject           `json:"error,omitempty"`
}

// ErrorObject represents a JSON-RPC error.
type ErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON-RPC error codes.
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)
