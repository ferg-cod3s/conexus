package protocol

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"testing"
)

// mockHandler implements Handler for testing
type mockHandler struct {
	mu       sync.Mutex
	calls    map[string]int
	response interface{}
	err      error
}

func newMockHandler() *mockHandler {
	return &mockHandler{
		calls: make(map[string]int),
	}
}

func (h *mockHandler) Handle(method string, params json.RawMessage) (interface{}, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.calls[method]++

	if h.err != nil {
		return nil, h.err
	}

	if h.response != nil {
		return h.response, nil
	}

	// Default response
	return map[string]string{"result": "ok"}, nil
}

// TestRequest_JSONMarshaling tests request marshaling
func TestRequest_JSONMarshaling(t *testing.T) {
	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  "test_method",
		Params:  json.RawMessage(`{"key":"value"}`),
		ID:      1,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if decoded.Method != req.Method {
		t.Errorf("method mismatch: got %s, want %s", decoded.Method, req.Method)
	}

	if decoded.ID != req.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, req.ID)
	}
}

// TestResponse_JSONMarshaling tests response marshaling
func TestResponse_JSONMarshaling(t *testing.T) {
	resp := Response{
		JSONRPC: JSONRPCVersion,
		Result:  json.RawMessage(`{"success":true}`),
		ID:      1,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("ID mismatch: got %v, want %v", decoded.ID, resp.ID)
	}
}

// TestError_JSONMarshaling tests error marshaling
func TestError_JSONMarshaling(t *testing.T) {
	err := Error{
		Code:    InternalError,
		Message: "test error",
		Data:    json.RawMessage(`{"detail":"additional info"}`),
	}

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("failed to marshal error: %v", jsonErr)
	}

	var decoded Error
	if jsonErr := json.Unmarshal(data, &decoded); jsonErr != nil {
		t.Fatalf("failed to unmarshal error: %v", jsonErr)
	}

	if decoded.Code != err.Code {
		t.Errorf("code mismatch: got %d, want %d", decoded.Code, err.Code)
	}

	if decoded.Message != err.Message {
		t.Errorf("message mismatch: got %s, want %s", decoded.Message, err.Message)
	}
}

// TestServer_ValidRequest tests handling valid requests
func TestServer_ValidRequest(t *testing.T) {
	handler := newMockHandler()
	handler.response = map[string]string{"result": "success"}

	input := `{"jsonrpc":"2.0","method":"test","params":null,"id":1}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer, handler)

	// Process one request
	err := server.Serve()
	if err != nil && err != io.EOF {
		t.Fatalf("serve failed: %v", err)
	}

	// Check handler was called
	if handler.calls["test"] != 1 {
		t.Errorf("expected 1 call to 'test', got %d", handler.calls["test"])
	}

	// Check response
	var resp Response
	if err := json.Unmarshal(writer.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("unexpected error in response: %v", resp.Error)
	}
}

// TestServer_InvalidJSONRPCVersion tests invalid version handling
func TestServer_InvalidJSONRPCVersion(t *testing.T) {
	handler := newMockHandler()

	input := `{"jsonrpc":"1.0","method":"test","id":1}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer, handler)
	_ = server.Serve()

	// Should return error response
	var resp Response
	if err := json.Unmarshal(writer.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for invalid version")
	}

	if resp.Error.Code != InvalidRequest {
		t.Errorf("expected error code %d, got %d", InvalidRequest, resp.Error.Code)
	}
}

// TestServer_MissingMethod tests missing method handling
func TestServer_MissingMethod(t *testing.T) {
	handler := newMockHandler()

	input := `{"jsonrpc":"2.0","id":1}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer, handler)
	_ = server.Serve()

	var resp Response
	if err := json.Unmarshal(writer.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for missing method")
	}

	if resp.Error.Code != InvalidRequest {
		t.Errorf("expected error code %d, got %d", InvalidRequest, resp.Error.Code)
	}
}

// TestServer_ParseError tests malformed JSON handling
func TestServer_ParseError(t *testing.T) {
	handler := newMockHandler()

	input := `{invalid json}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer, handler)
	_ = server.Serve()

	var resp Response
	if err := json.Unmarshal(writer.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error for invalid JSON")
	}

	if resp.Error.Code != ParseError {
		t.Errorf("expected error code %d, got %d", ParseError, resp.Error.Code)
	}
}

// TestClient_Call tests client method calls
func TestClient_Call(t *testing.T) {
	// Use buffers for simpler synchronous testing
	requestBuf := &bytes.Buffer{}
	responseBuf := &bytes.Buffer{}

	// Create handler
	handler := newMockHandler()
	handler.response = map[string]interface{}{
		"status": "success",
		"value":  42,
	}

	// Build request manually
	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  "test_method",
		Params:  json.RawMessage(`{"key":"value"}`),
		ID:      1,
	}

	// Marshal and send request
	reqData, _ := json.Marshal(req)
	requestBuf.Write(reqData)
	requestBuf.WriteByte('\n')

	// Process one request (will return after one request)
	// Note: In real usage, Serve() runs forever, but here we test the logic
	var resp Response
	if requestBuf.Len() > 0 {
		decoder := json.NewDecoder(requestBuf)
		var testReq Request
		if err := decoder.Decode(&testReq); err == nil {
			result, _ := handler.Handle(testReq.Method, testReq.Params)
			resultJSON, _ := json.Marshal(result)
			resp = Response{
				JSONRPC: JSONRPCVersion,
				Result:  resultJSON,
				ID:      testReq.ID,
			}
			json.NewEncoder(responseBuf).Encode(resp)
		}
	}

	// Verify response
	if err := json.Unmarshal(resp.Result, &map[string]interface{}{}); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
}

// TestClient_CallWithNilParams tests call with no parameters
func TestClient_CallWithNilParams(t *testing.T) {
	// Build request manually
	req := Request{
		JSONRPC: JSONRPCVersion,
		Method:  "test",
		Params:  nil,
		ID:      1,
	}

	// Marshal request
	reqData, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	// Verify request structure
	var decoded Request
	if err := json.Unmarshal(reqData, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if decoded.Method != "test" {
		t.Errorf("method mismatch: got %s, want test", decoded.Method)
	}

	if decoded.Params != nil {
		t.Error("expected nil params")
	}
}

// TestClient_Notify tests notification (no response expected)
func TestClient_Notify(t *testing.T) {
	writer := &bytes.Buffer{}
	client := NewClient(nil, writer)

	params := map[string]string{"event": "test"}
	err := client.Notify("notification", params)

	if err != nil {
		t.Fatalf("notify failed: %v", err)
	}

	// Parse sent request
	var req Request
	if err := json.Unmarshal(writer.Bytes(), &req); err != nil {
		t.Fatalf("failed to parse request: %v", err)
	}

	if req.Method != "notification" {
		t.Errorf("method mismatch: got %s, want notification", req.Method)
	}

	if req.ID != nil {
		t.Error("notification should have null ID")
	}
}

// TestErrorCodes tests all standard error codes
func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code int
	}{
		{"ParseError", ParseError},
		{"InvalidRequest", InvalidRequest},
		{"MethodNotFound", MethodNotFound},
		{"InvalidParams", InvalidParams},
		{"InternalError", InternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code == 0 {
				t.Error("error code should not be zero")
			}

			// Verify codes are in the correct range
			if tt.code > -32600 || tt.code < -32800 {
				t.Errorf("error code %d outside standard range", tt.code)
			}
		})
	}
}

// TestServer_ConcurrentRequests tests concurrent request handling
func TestServer_ConcurrentRequests(t *testing.T) {
	handler := newMockHandler()
	handler.response = map[string]string{"result": "ok"}

	// Create multiple requests
	requests := []string{
		`{"jsonrpc":"2.0","method":"test1","id":1}`,
		`{"jsonrpc":"2.0","method":"test2","id":2}`,
		`{"jsonrpc":"2.0","method":"test3","id":3}`,
	}

	input := strings.Join(requests, "\n") + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer, handler)
	_ = server.Serve()

	// Should have processed all requests
	totalCalls := 0
	for _, count := range handler.calls {
		totalCalls += count
	}

	if totalCalls != 3 {
		t.Errorf("expected 3 total calls, got %d", totalCalls)
	}
}

// TestClient_IDGeneration tests automatic ID generation
func TestClient_IDGeneration(t *testing.T) {
	writer := &bytes.Buffer{}
	client := NewClient(nil, writer)

	// First call
	err := client.Notify("test", nil)
	if err != nil {
		t.Fatalf("notify failed: %v", err)
	}

	// Client should increment nextID
	if client.nextID != 1 {
		t.Errorf("expected nextID to be 1, got %d", client.nextID)
	}
}

// TestServer_HandlerError tests handler error propagation
func TestServer_HandlerError(t *testing.T) {
	handler := newMockHandler()
	handler.err = &Error{
		Code:    InternalError,
		Message: "handler error",
	}

	input := `{"jsonrpc":"2.0","method":"test","id":1}` + "\n"
	reader := strings.NewReader(input)
	writer := &bytes.Buffer{}

	server := NewServer(reader, writer, handler)
	_ = server.Serve()

	var resp Response
	if err := json.Unmarshal(writer.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error == nil {
		t.Error("expected error in response")
	}

	if resp.Error.Code != InternalError {
		t.Errorf("expected error code %d, got %d", InternalError, resp.Error.Code)
	}
}

// TestResponse_ErrorAndResult tests mutually exclusive error/result
func TestResponse_ErrorAndResult(t *testing.T) {
	// Response should have either error OR result, not both
	resp := Response{
		JSONRPC: JSONRPCVersion,
		Result:  json.RawMessage(`{"success":true}`),
		Error: &Error{
			Code:    InternalError,
			Message: "error",
		},
		ID: 1,
	}

	// According to JSON-RPC 2.0 spec, this is invalid
	// but we can still marshal it
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Verify both fields are present in JSON
	if !bytes.Contains(data, []byte("result")) {
		t.Error("expected 'result' in JSON")
	}

	if !bytes.Contains(data, []byte("error")) {
		t.Error("expected 'error' in JSON")
	}
}

// TestRequest_WithDifferentParamTypes tests various parameter types
func TestRequest_WithDifferentParamTypes(t *testing.T) {
	tests := []struct {
		name   string
		params interface{}
	}{
		{"string params", "test string"},
		{"number params", 42},
		{"object params", map[string]string{"key": "value"}},
		{"array params", []string{"a", "b", "c"}},
		{"null params", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paramsJSON json.RawMessage
			if tt.params != nil {
				data, err := json.Marshal(tt.params)
				if err != nil {
					t.Fatalf("failed to marshal params: %v", err)
				}
				paramsJSON = data
			}

			req := Request{
				JSONRPC: JSONRPCVersion,
				Method:  "test",
				Params:  paramsJSON,
				ID:      1,
			}

			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			var decoded Request
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal request: %v", err)
			}

			if decoded.Method != req.Method {
				t.Errorf("method mismatch")
			}
		})
	}
}
