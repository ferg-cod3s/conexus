package protocol

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONRPCVersion is the JSON-RPC protocol version
const JSONRPCVersion = "2.0"

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// UnmarshalJSON implements custom unmarshaling to normalize ID types
func (r *Request) UnmarshalJSON(data []byte) error {
	type Alias Request
	aux := &struct {
		ID json.RawMessage `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Normalize the ID type
	r.ID = normalizeID(aux.ID)
	return nil
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

// UnmarshalJSON implements custom unmarshaling to normalize ID types
func (r *Response) UnmarshalJSON(data []byte) error {
	type Alias Response
	aux := &struct {
		ID json.RawMessage `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Normalize the ID type
	r.ID = normalizeID(aux.ID)
	return nil
}

// normalizeID converts JSON number IDs to int to ensure type consistency
func normalizeID(raw json.RawMessage) interface{} {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	// Try to parse as string (quoted)
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str
	}

	// Try to parse as number
	var num float64
	if err := json.Unmarshal(raw, &num); err == nil {
		// Convert to int if it's a whole number
		if num == float64(int(num)) {
			return int(num)
		}
		// Keep as float64 for non-integer numbers (rare but spec-compliant)
		return num
	}

	// Fallback: return as-is (shouldn't happen with valid JSON-RPC)
	return raw
}

// Error represents a JSON-RPC 2.0 error
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// Standard JSON-RPC error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// Handler handles JSON-RPC method calls
type Handler interface {
	Handle(method string, params json.RawMessage) (interface{}, error)
}

// Server handles JSON-RPC communication over stdio
type Server struct {
	reader  io.Reader
	writer  io.Writer
	handler Handler
}

// NewServer creates a new JSON-RPC server
func NewServer(reader io.Reader, writer io.Writer, handler Handler) *Server {
	return &Server{
		reader:  reader,
		writer:  writer,
		handler: handler,
	}
}

// Serve starts processing JSON-RPC requests
func (s *Server) Serve() error {
	decoder := json.NewDecoder(s.reader)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			// After a parse error, we cannot reliably continue reading from the stream
			return s.sendError(nil, ParseError, fmt.Sprintf("parse error: %v", err), nil)
		}

		// Validate request
		if req.JSONRPC != JSONRPCVersion {
			s.sendError(req.ID, InvalidRequest, "invalid jsonrpc version", nil)
			continue
		}

		if req.Method == "" {
			s.sendError(req.ID, InvalidRequest, "method required", nil)
			continue
		}

		// Handle the request
		result, err := s.handler.Handle(req.Method, req.Params)
		if err != nil {
			s.sendError(req.ID, InternalError, err.Error(), nil)
			continue
		}

		// Send successful response
		if err := s.sendResult(req.ID, result); err != nil {
			return fmt.Errorf("failed to send response: %w", err)
		}
	}
}

// sendResult sends a successful JSON-RPC response
func (s *Server) sendResult(id interface{}, result interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return s.sendError(id, InternalError, "failed to marshal result", nil)
	}

	resp := Response{
		JSONRPC: JSONRPCVersion,
		Result:  resultJSON,
		ID:      id,
	}

	return json.NewEncoder(s.writer).Encode(resp)
}

// sendError sends an error JSON-RPC response
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) error {
	var dataJSON json.RawMessage
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal error data: %w", err)
		}
	}

	resp := Response{
		JSONRPC: JSONRPCVersion,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    dataJSON,
		},
		ID: id,
	}

	return json.NewEncoder(s.writer).Encode(resp)
}

// Client represents a JSON-RPC client
type Client struct {
	reader io.Reader
	writer io.Writer
	nextID int
}

// NewClient creates a new JSON-RPC client
func NewClient(reader io.Reader, writer io.Writer) *Client {
	return &Client{
		reader: reader,
		writer: writer,
		nextID: 1,
	}
}

// Call makes a JSON-RPC method call
func (c *Client) Call(method string, params interface{}) (json.RawMessage, error) {
	// Marshal params
	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create request
	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  paramsJSON,
		ID:      c.nextID,
	}
	c.nextID++

	// Send request
	if err := json.NewEncoder(c.writer).Encode(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	var resp Response
	if err := json.NewDecoder(c.reader).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}

// Notify sends a JSON-RPC notification (no response expected)
func (c *Client) Notify(method string, params interface{}) error {
	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  method,
		Params:  paramsJSON,
		ID:      nil, // Null ID indicates notification
	}

	return json.NewEncoder(c.writer).Encode(req)
}
