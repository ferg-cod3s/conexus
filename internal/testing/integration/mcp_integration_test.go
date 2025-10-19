package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/mcp"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPServerConnection tests MCP server stdio transport connection
func TestMCPServerConnection(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func() (io.Reader, io.Writer)
		expectedError bool
		description   string
	}{
		{
			name: "valid_stdio_connection",
			setupFunc: func() (io.Reader, io.Writer) {
				return bytes.NewReader([]byte{}), &bytes.Buffer{}
			},
			expectedError: false,
			description:   "Should create server with valid stdio streams",
		},
		{
			name: "nil_reader",
			setupFunc: func() (io.Reader, io.Writer) {
				return nil, &bytes.Buffer{}
			},
			expectedError: false, // Server creation allows nil, Serve() will fail
			description:   "Should handle nil reader gracefully",
		},
		{
			name: "nil_writer",
			setupFunc: func() (io.Reader, io.Writer) {
				return bytes.NewReader([]byte{}), nil
			},
			expectedError: false, // Server creation allows nil, Serve() will fail
			description:   "Should handle nil writer gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, writer := tt.setupFunc()
			store := vectorstore.NewMemoryStore()
			embedder := embedding.NewMock(384)

			connStore, err := connectors.NewStore(":memory:")
			require.NoError(t, err)
			defer connStore.Close()

			loggerCfg := observability.LoggerConfig{
				Level: "error",
			}
			logger := observability.NewLogger(loggerCfg)

			// Use nil metrics to avoid registration issues in tests
			var metrics *observability.MetricsCollector
			errorHandler := observability.NewErrorHandler(logger, metrics, false)

			server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, nil)

			if !tt.expectedError {
				assert.NotNil(t, server, tt.description)
			}
		})
	}
}

// TestMCPToolDiscovery tests the tools/list endpoint
func TestMCPToolDiscovery(t *testing.T) {
	// Create JSON-RPC request for tools/list
	request := protocol.Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/list",
	}

	requestJSON, err := json.Marshal(request)
	require.NoError(t, err)

	// Add newline for JSON-RPC line protocol
	requestJSON = append(requestJSON, '\n')

	reader := bytes.NewReader(requestJSON)
	writer := &bytes.Buffer{}

	store := vectorstore.NewMemoryStore()
	embedder := embedding.NewMock(384)

	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()

	loggerCfg := observability.LoggerConfig{
		Level: "error",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-tools")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, nil)
	require.NotNil(t, server)

	// Run server in goroutine (it will process one request and EOF)
	done := make(chan error, 1)
	go func() {
		done <- server.Serve()
	}()

	// Wait for response or timeout
	select {
	case err := <-done:
		// EOF is expected after processing one request
		if err != nil && err != io.EOF {
			t.Fatalf("Serve returned unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for server response")
	}

	// Parse response
	responseData := writer.Bytes()
	if len(responseData) == 0 {
		t.Fatal("No response data received")
	}

	var response protocol.Response
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err, "Response should be valid JSON-RPC")

	// Verify response structure
	assert.Equal(t, "2.0", response.JSONRPC, "Should be JSON-RPC 2.0")
	assert.Nil(t, response.Error, "Should not have error")
	assert.NotNil(t, response.Result, "Should have result")

	// Parse result as tools list
	var result map[string]interface{}
	err = json.Unmarshal(response.Result, &result)
	require.NoError(t, err)

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok, "Result should contain 'tools' array")
	assert.Len(t, tools, 4, "Should discover 4 MCP tools")

	// Verify each tool has required fields
	expectedTools := map[string]bool{
		"context.search":               false,
		"context.get_related_info":     false,
		"context.index_control":        false,
		"context.connector_management": false,
	}

	for _, toolInterface := range tools {
		tool, ok := toolInterface.(map[string]interface{})
		require.True(t, ok, "Tool should be object")

		name, ok := tool["name"].(string)
		require.True(t, ok, "Tool should have name")

		description, ok := tool["description"].(string)
		require.True(t, ok, "Tool should have description")
		assert.NotEmpty(t, description, "Description should not be empty")

		inputSchema, ok := tool["inputSchema"].(map[string]interface{})
		require.True(t, ok, "Tool should have inputSchema")
		assert.NotNil(t, inputSchema, "InputSchema should not be nil")

		// Mark tool as found
		if _, exists := expectedTools[name]; exists {
			expectedTools[name] = true
		}
	}

	// Verify all expected tools were found
	for toolName, found := range expectedTools {
		assert.True(t, found, "Tool %s should be discovered", toolName)
	}
}

// TestMCPToolExecution tests individual tool call execution
func TestMCPToolExecution(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		args         interface{}
		setupStore   func(*vectorstore.MemoryStore) error
		validateResp func(*testing.T, interface{})
		expectError  bool
		description  string
	}{
		{
			name:     "context_search_basic",
			toolName: "context.search",
			args: map[string]interface{}{
				"query": "authentication implementation",
				"top_k": 5,
			},
			setupStore: func(store *vectorstore.MemoryStore) error {
				ctx := context.Background()
				now := time.Now()
				doc := vectorstore.Document{
					ID:        "test-doc-1",
					Content:   "authentication implementation details",
					Vector:    make(embedding.Vector, 384),
					Metadata:  map[string]interface{}{"source_type": "file"},
					CreatedAt: now,
					UpdatedAt: now,
				}
				return store.Upsert(ctx, doc)
			},
			validateResp: func(t *testing.T, result interface{}) {
				resp, ok := result.(map[string]interface{})
				require.True(t, ok, "Result should be object")

				assert.Contains(t, resp, "results", "Should have results field")
				assert.Contains(t, resp, "total_count", "Should have total_count field")
				assert.Contains(t, resp, "query_time_ms", "Should have query_time_ms field")

				totalCount, ok := resp["total_count"].(float64)
				require.True(t, ok, "total_count should be number")
				assert.GreaterOrEqual(t, int(totalCount), 0, "Should have non-negative count")
			},
			expectError: false,
			description: "Should execute basic search successfully",
		},
		{
			name:     "context_search_with_filters",
			toolName: "context.search",
			args: map[string]interface{}{
				"query": "test query",
				"top_k": 10,
				"filters": map[string]interface{}{
					"source_types": []string{"file", "github"},
				},
			},
			setupStore: func(store *vectorstore.MemoryStore) error {
				return nil // No docs needed for this test
			},
			validateResp: func(t *testing.T, result interface{}) {
				resp, ok := result.(map[string]interface{})
				require.True(t, ok, "Result should be object")
				assert.Contains(t, resp, "results", "Should have results field")
			},
			expectError: false,
			description: "Should handle search with filters",
		},
		{
			name:     "context_get_related_info_file",
			toolName: "context.get_related_info",
			args: map[string]interface{}{
				"file_path": "/path/to/file.go",
			},
			setupStore: func(store *vectorstore.MemoryStore) error {
				return nil
			},
			validateResp: func(t *testing.T, result interface{}) {
				resp, ok := result.(map[string]interface{})
				require.True(t, ok, "Result should be object")

				assert.Contains(t, resp, "summary", "Should have summary field")
				summary, ok := resp["summary"].(string)
				require.True(t, ok, "Summary should be string")
				assert.NotEmpty(t, summary, "Summary should not be empty")
			},
			expectError: false,
			description: "Should get related info for file",
		},
		{
			name:     "context_get_related_info_ticket",
			toolName: "context.get_related_info",
			args: map[string]interface{}{
				"ticket_id": "JIRA-123",
			},
			setupStore: func(store *vectorstore.MemoryStore) error {
				return nil
			},
			validateResp: func(t *testing.T, result interface{}) {
				resp, ok := result.(map[string]interface{})
				require.True(t, ok, "Result should be object")
				assert.Contains(t, resp, "summary", "Should have summary field")
			},
			expectError: false,
			description: "Should get related info for ticket",
		},
		{
			name:     "context_index_control_status",
			toolName: "context.index_control",
			args: map[string]interface{}{
				"action": "status",
			},
			setupStore: func(store *vectorstore.MemoryStore) error {
				ctx := context.Background()
				now := time.Now()
				doc := vectorstore.Document{
					ID:        "status-test-doc",
					Content:   "test content",
					Vector:    make(embedding.Vector, 384),
					Metadata:  map[string]interface{}{},
					CreatedAt: now,
					UpdatedAt: now,
				}
				return store.Upsert(ctx, doc)
			},
			validateResp: func(t *testing.T, result interface{}) {
				resp, ok := result.(map[string]interface{})
				require.True(t, ok, "Result should be object")

				assert.Contains(t, resp, "status", "Should have status field")
				assert.Contains(t, resp, "message", "Should have message field")

				status, ok := resp["status"].(string)
				require.True(t, ok, "Status should be string")
				assert.Equal(t, "ok", status, "Status should be 'ok'")
			},
			expectError: false,
			description: "Should return index status",
		},
		{
			name:     "context_connector_management_list",
			toolName: "context.connector_management",
			args: map[string]interface{}{
				"action": "list",
			},
			setupStore: func(store *vectorstore.MemoryStore) error {
				return nil
			},
			validateResp: func(t *testing.T, result interface{}) {
				resp, ok := result.(map[string]interface{})
				require.True(t, ok, "Result should be object")

				assert.Contains(t, resp, "connectors", "Should have connectors field")
				assert.Contains(t, resp, "status", "Should have status field")

				connectors, ok := resp["connectors"].([]interface{})
				require.True(t, ok, "Connectors should be array")
				assert.GreaterOrEqual(t, len(connectors), 1, "Should have at least one connector")
			},
			expectError: false,
			description: "Should list connectors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup store
			store := vectorstore.NewMemoryStore()
			if tt.setupStore != nil {
				err := tt.setupStore(store)
				require.NoError(t, err, "Store setup should succeed")
			}

			embedder := embedding.NewMock(384)

			connStore, err := connectors.NewStore(":memory:")
			require.NoError(t, err)
			defer connStore.Close()

			loggerCfg := observability.LoggerConfig{
				Level: "error",
			}
			logger := observability.NewLogger(loggerCfg)
			metrics := observability.NewMetricsCollector("test-tool-call")
			errorHandler := observability.NewErrorHandler(logger, metrics, false)

			// Create tool call request
			argsJSON, err := json.Marshal(tt.args)
			require.NoError(t, err)

			toolCallReq := map[string]interface{}{
				"name":      tt.toolName,
				"arguments": json.RawMessage(argsJSON),
			}

			toolCallJSON, err := json.Marshal(toolCallReq)
			require.NoError(t, err)

			// Create JSON-RPC request
			request := protocol.Request{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Method:  "tools/call",
				Params:  json.RawMessage(toolCallJSON),
			}

			requestJSON, err := json.Marshal(request)
			require.NoError(t, err)
			requestJSON = append(requestJSON, '\n')

			// Setup server
			reader := bytes.NewReader(requestJSON)
			writer := &bytes.Buffer{}
			server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, nil)

			// Run server
			done := make(chan error, 1)
			go func() {
				done <- server.Serve()
			}()

			select {
			case err := <-done:
				if err != nil && err != io.EOF {
					if !tt.expectError {
						t.Fatalf("Serve returned unexpected error: %v", err)
					}
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Timeout waiting for server response")
			}

			// Parse response
			responseData := writer.Bytes()
			var response protocol.Response
			err = json.Unmarshal(responseData, &response)
			require.NoError(t, err, "Response should be valid JSON-RPC")

			if tt.expectError {
				assert.NotNil(t, response.Error, "Should have error")
			} else {
				assert.Nil(t, response.Error, "Should not have error: %v", response.Error)
				assert.NotNil(t, response.Result, "Should have result")

				// Parse and validate result
				var result interface{}
				err = json.Unmarshal(response.Result, &result)
				require.NoError(t, err, "Result should be valid JSON")

				if tt.validateResp != nil {
					tt.validateResp(t, result)
				}
			}
		})
	}
}

// TestMCPErrorHandling tests error response formats
func TestMCPErrorHandling(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		params          interface{}
		expectedErrCode int
		description     string
	}{
		{
			name:            "invalid_method",
			method:          "invalid.method",
			params:          json.RawMessage(`{}`),
			expectedErrCode: protocol.MethodNotFound,
			description:     "Should return MethodNotFound for invalid method",
		},
		{
			name:            "invalid_params_structure",
			method:          "tools/call",
			params:          json.RawMessage(`"invalid"`),
			expectedErrCode: protocol.InvalidParams,
			description:     "Should return InvalidParams for malformed params",
		},
		{
			name:   "missing_required_field_search",
			method: "tools/call",
			params: map[string]interface{}{
				"name": "context.search",
				"arguments": map[string]interface{}{
					// Missing required "query" field
					"top_k": 10,
				},
			},
			expectedErrCode: protocol.InvalidParams,
			description:     "Should return InvalidParams for missing query",
		},
		{
			name:   "invalid_tool_name",
			method: "tools/call",
			params: map[string]interface{}{
				"name":      "invalid.tool",
				"arguments": map[string]interface{}{},
			},
			expectedErrCode: protocol.MethodNotFound,
			description:     "Should return MethodNotFound for unknown tool",
		},
		{
			name:   "missing_file_and_ticket",
			method: "tools/call",
			params: map[string]interface{}{
				"name":      "context.get_related_info",
				"arguments": map[string]interface{}{
					// Missing both file_path and ticket_id
				},
			},
			expectedErrCode: protocol.InvalidParams,
			description:     "Should require either file_path or ticket_id",
		},
		{
			name:   "invalid_index_action",
			method: "tools/call",
			params: map[string]interface{}{
				"name": "context.index_control",
				"arguments": map[string]interface{}{
					"action": "invalid_action",
				},
			},
			expectedErrCode: protocol.InvalidParams,
			description:     "Should reject invalid index control action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := vectorstore.NewMemoryStore()
			embedder := embedding.NewMock(384)

			connStore, err := connectors.NewStore(":memory:")
			require.NoError(t, err)
			defer connStore.Close()

			loggerCfg := observability.LoggerConfig{
				Level: "error",
			}
			logger := observability.NewLogger(loggerCfg)
			metrics := observability.NewMetricsCollector("test-method")
			errorHandler := observability.NewErrorHandler(logger, metrics, false)

			var paramsJSON json.RawMessage
			var err2 error

			switch v := tt.params.(type) {
			case json.RawMessage:
				paramsJSON = v
			case string:
				paramsJSON = json.RawMessage(v)
			default:
				paramsJSON, err2 = json.Marshal(v)
				require.NoError(t, err2)
			}

			request := protocol.Request{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Method:  tt.method,
				Params:  paramsJSON,
			}

			requestJSON, err2 := json.Marshal(request)
			require.NoError(t, err2)
			requestJSON = append(requestJSON, '\n')

			reader := bytes.NewReader(requestJSON)
			writer := &bytes.Buffer{}
			server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, nil)

			done := make(chan error, 1)
			go func() {
				done <- server.Serve()
			}()

			select {
			case err := <-done:
				if err != nil && err != io.EOF {
					t.Logf("Serve error (may be expected): %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Timeout waiting for server response")
			}

			responseData := writer.Bytes()
			if len(responseData) == 0 {
				t.Fatal("No response data received")
			}

			var response protocol.Response
			err2 = json.Unmarshal(responseData, &response)
			require.NoError(t, err2, "Response should be valid JSON-RPC")

			assert.NotNil(t, response.Error, "Should have error: %s", tt.description)
			if response.Error != nil {
				assert.Equal(t, tt.expectedErrCode, response.Error.Code,
					"Error code should match: %s", tt.description)
				assert.NotEmpty(t, response.Error.Message, "Error message should not be empty")
			}
		})
	}
}

// TestMCPProtocolCompliance tests JSON-RPC 2.0 format validation
func TestMCPProtocolCompliance(t *testing.T) {
	tests := []struct {
		name        string
		request     string
		shouldError bool
		description string
	}{
		{
			name:        "valid_jsonrpc_request",
			request:     `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
			shouldError: false,
			description: "Valid JSON-RPC 2.0 request should be accepted",
		},
		{
			name:        "missing_jsonrpc_field",
			request:     `{"id":1,"method":"tools/list"}`,
			shouldError: true,
			description: "Request without jsonrpc field should fail",
		},
		{
			name:        "wrong_jsonrpc_version",
			request:     `{"jsonrpc":"1.0","id":1,"method":"tools/list"}`,
			shouldError: true,
			description: "Wrong JSON-RPC version should fail",
		},
		{
			name:        "missing_method",
			request:     `{"jsonrpc":"2.0","id":1}`,
			shouldError: true,
			description: "Request without method should fail",
		},
		{
			name:        "malformed_json",
			request:     `{"jsonrpc":"2.0",invalid}`,
			shouldError: true,
			description: "Malformed JSON should fail gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := vectorstore.NewMemoryStore()
			embedder := embedding.NewMock(384)

			connStore, err := connectors.NewStore(":memory:")
			require.NoError(t, err)
			defer connStore.Close()

			loggerCfg := observability.LoggerConfig{
				Level: "error",
			}
			logger := observability.NewLogger(loggerCfg)
			metrics := observability.NewMetricsCollector("test-protocol")
			errorHandler := observability.NewErrorHandler(logger, metrics, false)

			requestData := []byte(tt.request + "\n")
			reader := bytes.NewReader(requestData)
			writer := &bytes.Buffer{}
			server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, nil)

			done := make(chan error, 1)
			go func() {
				done <- server.Serve()
			}()

			select {
			case err := <-done:
				if tt.shouldError {
					// For protocol errors, we might get error from Serve or in response
					// Both are acceptable
					responseData := writer.Bytes()
					if len(responseData) > 0 {
						var response protocol.Response
						jsonErr := json.Unmarshal(responseData, &response)
						if jsonErr == nil && response.Error != nil {
							// Error in response is expected
							return
						}
					}
					// Or error from Serve
					if err != nil && err != io.EOF {
						return
					}
					t.Errorf("Expected error but got none: %s", tt.description)
				} else {
					if err != nil && err != io.EOF {
						t.Errorf("Unexpected error: %v", err)
					}
				}
			case <-time.After(2 * time.Second):
				t.Fatal("Timeout waiting for server response")
			}

			// For valid requests, verify response format
			if !tt.shouldError {
				responseData := writer.Bytes()
				if len(responseData) > 0 {
					var response protocol.Response
					err := json.Unmarshal(responseData, &response)
					require.NoError(t, err, "Response should be valid JSON")
					assert.Equal(t, "2.0", response.JSONRPC, "Response should be JSON-RPC 2.0")
				}
			}
		})
	}
}

// TestMCPConcurrentRequests tests handling multiple concurrent tool calls
func TestMCPConcurrentRequests(t *testing.T) {
	t.Skip("Skipping concurrent requests test - requires multi-request handling")

	// This test would verify that multiple requests can be processed
	// Currently, our test setup processes one request at a time
	// In production with Claude Desktop, requests come sequentially
}

// TestMCPTimeoutHandling tests timeout behavior
func TestMCPTimeoutHandling(t *testing.T) {
	t.Skip("Skipping timeout test - requires slow operation simulation")

	// This test would verify timeout handling for slow operations
	// Would need to create a slow embedder or slow vectorstore mock
}

// Helper function to validate JSON-RPC response format
func validateJSONRPCResponse(t *testing.T, data []byte) protocol.Response {
	var response protocol.Response
	err := json.Unmarshal(data, &response)
	require.NoError(t, err, "Response should be valid JSON")
	assert.Equal(t, "2.0", response.JSONRPC, "Should be JSON-RPC 2.0")
	return response
}

// Helper function to create test document
func createTestDocument(id, content, sourceType string) vectorstore.Document {
	now := time.Now()
	return vectorstore.Document{
		ID:      id,
		Content: content,
		Vector:  make(embedding.Vector, 384),
		Metadata: map[string]interface{}{
			"source_type": sourceType,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
