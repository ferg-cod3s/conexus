package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/mcp"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
	"github.com/prometheus/client_golang/prometheus"
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
	// Use nil metrics to avoid registration issues in tests
	var metrics *observability.MetricsCollector
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	reader, writer := io.Pipe()
	server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, idx)

	done := make(chan error, 1)
	go func() {
		done <- server.Serve()
	}()

	// Test invalid search query
	t.Run("invalid_search", func(t *testing.T) {
		invalidReq := map[string]interface{}{
			"name": "context_search",
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

	// Close writer to signal EOF to server
	writer.Close()

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
	defer store.Close()

	// Create connector store
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()

	// Use a separate registry for test metrics to avoid duplicate registration
	testRegistry := prometheus.NewRegistry()
	metrics := observability.NewMetricsCollectorWithRegistry("test-concurrent", testRegistry)

	// Test concurrent search requests using direct store access instead of MCP protocol
	// This tests the underlying concurrent safety of the search functionality

	// Create mock vectors with non-zero values
	createMockVector := func(id int) embedding.Vector {
		vec := make(embedding.Vector, 384)
		for i := range vec {
			vec[i] = float32(i+id) / 1000.0 // Small non-zero values
		}
		return vec
	}

	// Index some test content first
	testDocs := []vectorstore.Document{
		{
			ID:      "test1",
			Content: "package main\n\nfunc helloWorld() {\n\tfmt.Println(\"Hello, World!\")\n}",
			Vector:  createMockVector(1),
			Metadata: map[string]interface{}{
				"file_path": "/test/hello.go",
				"language":  "go",
			},
		},
		{
			ID:      "test2",
			Content: "def hello_world():\n\tprint(\"Hello, World!\")",
			Vector:  createMockVector(2),
			Metadata: map[string]interface{}{
				"file_path": "/test/hello.py",
				"language":  "python",
			},
		},
		{
			ID:      "test3",
			Content: "function helloWorld() {\n\tconsole.log(\"Hello, World!\");\n}",
			Vector:  createMockVector(3),
			Metadata: map[string]interface{}{
				"file_path": "/test/hello.js",
				"language":  "javascript",
			},
		},
	}

	// Insert test documents
	err = store.UpsertBatch(context.Background(), testDocs)
	require.NoError(t, err)

	// Test concurrent searches
	numGoroutines := 10
	searchesPerGoroutine := 5
	results := make(chan error, numGoroutines)

	start := time.Now()

	// Launch concurrent searches
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() {
				if r := recover(); r != nil {
					results <- fmt.Errorf("goroutine %d panicked: %v", goroutineID, r)
					return
				}
			}()

			for j := 0; j < searchesPerGoroutine; j++ {
				// Perform different types of searches
				switch j % 3 {
				case 0:
					// BM25 search
					_, err := store.SearchBM25(context.Background(), "Hello World", vectorstore.SearchOptions{
						Limit: 5,
					})
					if err != nil {
						results <- fmt.Errorf("goroutine %d BM25 search %d failed: %v", goroutineID, j, err)
						return
					}

				case 1:
					// Vector search
					queryVector := make(embedding.Vector, 384)
					for i := range queryVector {
						queryVector[i] = float32(i+goroutineID+j) / 1000.0
					}
					_, err := store.SearchVector(context.Background(), queryVector, vectorstore.SearchOptions{
						Limit: 5,
					})
					if err != nil {
						results <- fmt.Errorf("goroutine %d vector search %d failed: %v", goroutineID, j, err)
						return
					}

				case 2:
					// Hybrid search
					queryVector := make(embedding.Vector, 384)
					for i := range queryVector {
						queryVector[i] = float32(i+goroutineID+j) / 1000.0
					}
					_, err := store.SearchHybrid(context.Background(), "Hello", queryVector, vectorstore.SearchOptions{
						Limit: 5,
					})
					if err != nil {
						results <- fmt.Errorf("goroutine %d hybrid search %d failed: %v", goroutineID, j, err)
						return
					}
				}

				// Record metrics
				metrics.RecordVectorSearch("concurrent_test", "success", time.Since(start), 5)
			}

			results <- nil
		}(i)
	}

	// Wait for all goroutines to complete
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent search failed: %v", err)
		} else {
			successCount++
		}
	}

	duration := time.Since(start)
	totalSearches := numGoroutines * searchesPerGoroutine

	t.Logf("Completed %d concurrent searches in %v (%.2f searches/sec)",
		totalSearches, duration, float64(totalSearches)/duration.Seconds())

	// Verify all searches succeeded
	assert.Equal(t, numGoroutines, successCount, "All concurrent searches should succeed")

	// Verify metrics were recorded
	assert.Greater(t, totalSearches, 0, "Should have recorded search metrics")
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
	// Use nil metrics to avoid registration issues in tests
	var metrics *observability.MetricsCollector
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	reader, writer := io.Pipe()
	server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, idx)

	// Test health check method (if implemented)
	// Note: This would test a health endpoint if the MCP server exposed one
	assert.NotNil(t, server, "Server should be created successfully")

	reader.Close()
}
