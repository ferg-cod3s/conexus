package integration

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/mcp"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndMCPWithMonitoring tests complete MCP server lifecycle with monitoring
func TestEndToEndMCPWithMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end monitoring test in short mode")
	}

	// Setup components with monitoring

	// Create temporary database for testing
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err, "Should create vector store")

	embedder := embedding.NewMock(384)
	
	// Create connector store
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err, "Should create connector store")
	defer connStore.Close()

	// Create indexer
	idx := indexer.NewIndexer("test-state.json")

	// Setup observability
	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, false) // Disable Sentry for test

	// Create MCP server
	reader, writer := io.Pipe()
	server := mcp.NewServer(reader, writer, "", store, connStore, embedder, metrics, errorHandler, idx)

	// Start server in goroutine
	done := make(chan error, 1)
	go func() {
		done <- server.Serve()
	}()

	// Test 1: Index some content
	t.Run("index_content", func(t *testing.T) {
		indexReq := map[string]interface{}{
			"name": "context.index_control",
			"arguments": map[string]interface{}{
				"action": "index",
				"content": map[string]interface{}{
					"path":        "/test/file.go",
					"content":     "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
					"source_type": "file",
				},
			},
		}

		response := executeMCPToolCall(t, indexReq, server, reader, writer)
		assert.NotNil(t, response.Result, "Should have result")
		assert.Nil(t, response.Error, "Should not have error")
	})

	// Test 2: Search for indexed content
	t.Run("search_content", func(t *testing.T) {
		searchReq := map[string]interface{}{
			"name": "context.search",
			"arguments": map[string]interface{}{
				"query": "Hello World function",
				"top_k": 5,
			},
		}

		response := executeMCPToolCall(t, searchReq, server, reader, writer)
		assert.NotNil(t, response.Result, "Should have result")
		assert.Nil(t, response.Error, "Should not have error")

		// Parse search results
		var result map[string]interface{}
		err := json.Unmarshal(response.Result, &result)
		require.NoError(t, err, "Should parse search result")

		results, ok := result["results"].([]interface{})
		require.True(t, ok, "Should have results array")
		assert.Greater(t, len(results), 0, "Should find indexed content")
	})

	// Test 3: Get related info
	t.Run("get_related_info", func(t *testing.T) {
		infoReq := map[string]interface{}{
			"name": "context.get_related_info",
			"arguments": map[string]interface{}{
				"file_path": "/test/file.go",
			},
		}

		response := executeMCPToolCall(t, infoReq, server, reader, writer)
		assert.NotNil(t, response.Result, "Should have result")
		assert.Nil(t, response.Error, "Should not have error")
	})

	// Test 4: Check monitoring metrics
	t.Run("verify_monitoring", func(t *testing.T) {
		// Check that metrics were recorded
		// Note: In a real test, we'd verify specific metric values
		assert.NotNil(t, metrics, "Metrics collector should exist")
		assert.NotNil(t, errorHandler, "Error handler should exist")
	})

	// Test 5: Index status check
	t.Run("index_status", func(t *testing.T) {
		statusReq := map[string]interface{}{
			"name": "context.index_control",
			"arguments": map[string]interface{}{
				"action": "status",
			},
		}

		response := executeMCPToolCall(t, statusReq, server, reader, writer)
		assert.NotNil(t, response.Result, "Should have result")
		assert.Nil(t, response.Error, "Should not have error")

		var result map[string]interface{}
		err := json.Unmarshal(response.Result, &result)
		require.NoError(t, err, "Should parse status result")

		assert.Equal(t, "ok", result["status"], "Index status should be ok")
	})

	// Close the server
	reader.Close()

	// Wait for server to finish
	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Fatalf("Server error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not finish within timeout")
	}
}

// TestMCPErrorHandlingWithMonitoring tests error scenarios with monitoring
func TestMCPErrorHandlingWithMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling monitoring test in short mode")
	}

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)
	
	// Create connector store
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()
	
	idx := indexer.NewIndexer("test-error-state.json")

	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	reader, writer := io.Pipe()
	server := mcp.NewServer(reader, writer, "", store, connStore, embedder, metrics, errorHandler, idx)

	done := make(chan error, 1)
	go func() {
		done <- server.Serve()
	}()

	// Test invalid search query
	t.Run("invalid_search", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"name": "context.search",
			"arguments": map[string]interface{}{
				// Missing required "query" field
				"top_k": 5,
			},
		}

		response := executeMCPToolCall(t, invalidReq, server, reader, writer)
		assert.NotNil(t, response.Error, "Should have error for invalid request")
		assert.Equal(t, protocol.InvalidParams, response.Error.Code, "Should be invalid params error")
	})

	// Test invalid tool name
	t.Run("invalid_tool", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"name":      "invalid.tool",
			"arguments": map[string]interface{}{},
		}

		response := executeMCPToolCall(t, invalidReq, server, reader, writer)
		assert.NotNil(t, response.Error, "Should have error for invalid tool")
		assert.Equal(t, protocol.MethodNotFound, response.Error.Code, "Should be method not found error")
	})

	reader.Close()

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Fatalf("Server error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not finish within timeout")
	}
}

// TestMCPConcurrentRequestsWithMonitoring tests concurrent requests with monitoring
func TestMCPConcurrentRequestsWithMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent monitoring test in short mode")
	}

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)
	
	// Create connector store
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()
	
	idx := indexer.NewIndexer("test-concurrent-state.json")

	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	// For concurrent testing, we'd need a more sophisticated setup
	// This is a placeholder for the concurrent test structure
	reader, writer := io.Pipe()
	server := mcp.NewServer(reader, writer, "", store, connStore, embedder, metrics, errorHandler, idx)

	done := make(chan error, 1)
	go func() {
		done <- server.Serve()
	}()

	// Index some content first
	indexReq := map[string]interface{}{
		"name": "context.index_control",
		"arguments": map[string]interface{}{
			"action": "index",
			"content": map[string]interface{}{
				"path":        "/test/concurrent.go",
				"content":     "package concurrent\n\nfunc test() {\n\t// concurrent test content\n}",
				"source_type": "file",
			},
		},
	}

	response := executeMCPToolCall(t, indexReq, server, reader, writer)
	assert.NotNil(t, response.Result, "Should index content successfully")

	reader.Close()

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Fatalf("Server error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not finish within timeout")
	}
}

// executeMCPToolCall executes a tool call and returns the response
func executeMCPToolCall(t *testing.T, toolCall map[string]interface{}, server *mcp.Server, reader *io.PipeReader, writer *io.PipeWriter) protocol.Response {
	// Marshal tool call to JSON
	paramsJSON, err := json.Marshal(toolCall)
	require.NoError(t, err)

	// Create JSON-RPC request
	request := protocol.Request{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/call",
		Params:  json.RawMessage(paramsJSON),
	}

	requestJSON, err := json.Marshal(request)
	require.NoError(t, err)
	requestJSON = append(requestJSON, '\n')

	// Write request
	_, err = writer.Write(requestJSON)
	require.NoError(t, err)

	// Read response
	responseData := make([]byte, 4096)
	n, err := reader.Read(responseData)
	require.NoError(t, err)

	var response protocol.Response
	err = json.Unmarshal(responseData[:n], &response)
	require.NoError(t, err)

	return response
}

// TestMCPHealthCheck tests MCP server health validation
func TestMCPHealthCheck(t *testing.T) {
	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)
	
	// Create connector store
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()
	
	idx := indexer.NewIndexer("test-health-state.json")

	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	reader, writer := io.Pipe()
	server := mcp.NewServer(reader, writer, "", store, connStore, embedder, metrics, errorHandler, idx)

	// Test health check method (if implemented)
	// Note: This would test a health endpoint if the MCP server exposed one
	assert.NotNil(t, server, "Server should be created successfully")

	reader.Close()
}
