// Package integration provides MCP tool validation with real-world Conexus data
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/mcp"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getProjectRoot returns absolute path to project root
func getProjectRoot(t *testing.T) string {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err, "Should get project root")
	return root
}

// indexRealCodebase indexes actual Conexus source files into vector store
func indexRealCodebase(t *testing.T, ctx context.Context, store vectorstore.VectorStore, embedder embedding.Embedder, dir string) int {
	indexed := 0

	// Walk directory and index Go files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files and test files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			t.Logf("Warning: failed to read %s: %v", path, err)
			return nil // Skip this file
		}

		// Create embedding (mock embedder generates deterministic vectors)
		vec, err := embedder.Embed(ctx, string(content))
		if err != nil {
			t.Logf("Warning: failed to embed %s: %v", path, err)
			return nil
		}

		// Create document
		now := time.Now()
		doc := vectorstore.Document{
			ID:      path,
			Content: string(content),
			Vector:  vec.Vector,
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   path,
				"indexed_at":  now.Unix(),
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Upsert to store
		if err := store.Upsert(ctx, doc); err != nil {
			t.Logf("Warning: failed to index %s: %v", path, err)
			return nil
		}

		indexed++
		return nil
	})

	require.NoError(t, err, "Should walk directory")
	t.Logf("Indexed %d Go files from %s", indexed, dir)
	return indexed
}

// callMCPTool is a helper to call MCP tools via JSON-RPC
func callMCPTool(t *testing.T, toolName string, args map[string]interface{}) (map[string]interface{}, *protocol.Response) {
	// Create tool call arguments
	argsJSON, err := json.Marshal(args)
	require.NoError(t, err)

	// Create tool call request
	toolCallReq := map[string]interface{}{
		"name":      "context.index_control",
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
	store := vectorstore.NewMemoryStore()
	embedder := embedding.NewMock(384)
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()

	loggerCfg := observability.LoggerConfig{Level: "error"}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-realworld")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

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
			t.Fatalf("Serve error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for response")
	}

	// Parse response
	responseData := writer.Bytes()
	t.Logf("Response data: %s", string(responseData))
	if len(responseData) == 0 {
		t.Fatal("No response data received")
	}
	var response protocol.Response
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)
	t.Logf("Response result: %s", string(response.Result))

	var result map[string]interface{}
	err = json.Unmarshal(response.Result, &result)
	require.NoError(t, err)

	if response.Error != nil {
		t.Fatalf("Tool call error: %s", response.Error.Message)
	}

	return result, &response
}

// TestMCPRealWorldDataValidation validates MCP tools with real Conexus codebase
func TestMCPRealWorldDataValidation(t *testing.T) {
	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	// Create store with real data
	store := vectorstore.NewMemoryStore()
	embedder := embedding.NewMock(384)

	// Index internal/agent directory (small corpus)
	agentDir := filepath.Join(projectRoot, "internal", "agent")
	indexed := indexRealCodebase(t, ctx, store, embedder, agentDir)
	require.Greater(t, indexed, 0, "Should index some files")

	// Setup MCP server
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()

	loggerCfg := observability.LoggerConfig{Level: "error"}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-realworld-validation")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	// Test 1: context.search with real indexed data
	t.Run("search_with_real_data", func(t *testing.T) {
		// Create search request
		argsJSON, err := json.Marshal(map[string]interface{}{
			"query": "Execute method",
			"top_k": 5,
		})
		require.NoError(t, err)

		toolCallReq := map[string]interface{}{
			"name":      "context.search",
			"arguments": json.RawMessage(argsJSON),
		}

		toolCallJSON, err := json.Marshal(toolCallReq)
		require.NoError(t, err)

		request := protocol.Request{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "tools/call",
			Params:  json.RawMessage(toolCallJSON),
		}

		requestJSON, err := json.Marshal(request)
		require.NoError(t, err)
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
				t.Fatalf("Serve error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		// Parse response
		responseData := writer.Bytes()
		var response protocol.Response
		err = json.Unmarshal(responseData, &response)
		require.NoError(t, err)

		assert.Nil(t, response.Error, "Should not have error")
		assert.NotNil(t, response.Result, "Should have result")

		var result map[string]interface{}
		err = json.Unmarshal(response.Result, &result)
		require.NoError(t, err)

		// Verify result structure
		assert.Contains(t, result, "results")
		results := result["results"].([]interface{})

		// With real indexed data, we should find results
		if len(results) > 0 {
			t.Logf("✓ Found %d results for 'Execute method'", len(results))

			// Verify first result has expected fields
			firstResult := results[0].(map[string]interface{})
			assert.Contains(t, firstResult, "content")
			assert.Contains(t, firstResult, "score")
			assert.Contains(t, firstResult, "metadata")
		} else {
			t.Log("⚠ No results found (embeddings may not match)")
		}
	})

	// Test 2: context.index_control status
	t.Run("index_status_with_real_data", func(t *testing.T) {
		argsJSON, err := json.Marshal(map[string]interface{}{
			"action": "status",
		})
		require.NoError(t, err)

		toolCallReq := map[string]interface{}{
			"name":      "context.index_control",
			"arguments": json.RawMessage(argsJSON),
		}

		toolCallJSON, err := json.Marshal(toolCallReq)
		require.NoError(t, err)

		request := protocol.Request{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "tools/call",
			Params:  json.RawMessage(toolCallJSON),
		}

		requestJSON, err := json.Marshal(request)
		require.NoError(t, err)
		requestJSON = append(requestJSON, '\n')

		reader := bytes.NewReader(requestJSON)
		writer := &bytes.Buffer{}

		// Create indexer for status testing
		tempDir := t.TempDir()
		idx := indexer.NewIndexer(filepath.Join(tempDir, "test-state.json"))

		server := mcp.NewServer(reader, writer, store, connStore, embedder, metrics, errorHandler, idx)

		done := make(chan error, 1)
		go func() {
			done <- server.Serve()
		}()

		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				t.Fatalf("Serve error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		responseData := writer.Bytes()
		var response protocol.Response
		err = json.Unmarshal(responseData, &response)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(response.Result, &result)
		require.NoError(t, err)

		// Verify status
		assert.Contains(t, result, "status")
		assert.Equal(t, "ok", result["status"])

		// Should show document count
		assert.Contains(t, result, "message")
		message := result["message"].(string)
		assert.Contains(t, message, "documents")

		t.Logf("✓ Index status: %s", message)
	})

	// Test 3: context.get_related_info with real file
	t.Run("get_related_info_real_file", func(t *testing.T) {
		// Use actual analyzer.go file
		analyzerPath := filepath.Join(projectRoot, "internal", "agent", "analyzer", "analyzer.go")

		argsJSON, err := json.Marshal(map[string]interface{}{
			"file_path": analyzerPath,
		})
		require.NoError(t, err)

		toolCallReq := map[string]interface{}{
			"name":      "context.get_related_info",
			"arguments": json.RawMessage(argsJSON),
		}

		toolCallJSON, err := json.Marshal(toolCallReq)
		require.NoError(t, err)

		request := protocol.Request{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "tools/call",
			Params:  json.RawMessage(toolCallJSON),
		}

		requestJSON, err := json.Marshal(request)
		require.NoError(t, err)
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
				t.Fatalf("Serve error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		responseData := writer.Bytes()
		var response protocol.Response
		err = json.Unmarshal(responseData, &response)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(response.Result, &result)
		require.NoError(t, err)

		// Verify structure
		assert.Contains(t, result, "summary")
		summary := result["summary"].(string)
		assert.NotEmpty(t, summary)

		t.Logf("✓ Got related info for analyzer.go: %s", summary[:min(50, len(summary))])
	})
}

// TestMCPMultiStepWorkflow tests chaining MCP tools together
func TestMCPMultiStepWorkflow(t *testing.T) {
	t.Skip("Multi-step workflow requires more complex test orchestration")

	// This would test:
	// 1. Search for relevant files
	// 2. Use search results to call get_related_info
	// 3. Verify the workflow produces coherent output
}

// TestMCPEdgeCases tests edge cases with real data
func TestMCPEdgeCases(t *testing.T) {
	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	store := vectorstore.NewMemoryStore()
	embedder := embedding.NewMock(384)

	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()

	loggerCfg := observability.LoggerConfig{Level: "error"}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-edge-cases")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	// Test 1: Empty index search
	t.Run("search_empty_index", func(t *testing.T) {
		// Don't index anything - search empty store
		argsJSON, err := json.Marshal(map[string]interface{}{
			"query": "anything",
			"top_k": 5,
		})
		require.NoError(t, err)

		toolCallReq := map[string]interface{}{
			"name":      "context.search",
			"arguments": json.RawMessage(argsJSON),
		}

		toolCallJSON, err := json.Marshal(toolCallReq)
		require.NoError(t, err)

		request := protocol.Request{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "tools/call",
			Params:  json.RawMessage(toolCallJSON),
		}

		requestJSON, err := json.Marshal(request)
		require.NoError(t, err)
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
				t.Fatalf("Serve error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		responseData := writer.Bytes()
		var response protocol.Response
		err = json.Unmarshal(responseData, &response)
		require.NoError(t, err)

		// Should not error on empty results
		assert.Nil(t, response.Error)

		var result map[string]interface{}
		err = json.Unmarshal(response.Result, &result)
		require.NoError(t, err)

		results := result["results"].([]interface{})
		assert.Empty(t, results, "Empty index should return 0 results")

		t.Log("✓ Empty index handled gracefully")
	})

	// Test 2: Non-existent file path
	t.Run("get_related_info_nonexistent_file", func(t *testing.T) {
		argsJSON, err := json.Marshal(map[string]interface{}{
			"file_path": "/nonexistent/file.go",
		})
		require.NoError(t, err)

		toolCallReq := map[string]interface{}{
			"name":      "context.get_related_info",
			"arguments": json.RawMessage(argsJSON),
		}

		toolCallJSON, err := json.Marshal(toolCallReq)
		require.NoError(t, err)

		request := protocol.Request{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "tools/call",
			Params:  json.RawMessage(toolCallJSON),
		}

		requestJSON, err := json.Marshal(request)
		require.NoError(t, err)
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
				// May error or return graceful message
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		responseData := writer.Bytes()
		var response protocol.Response
		err = json.Unmarshal(responseData, &response)
		require.NoError(t, err)

		// Either error or empty result acceptable
		if response.Error == nil {
			var result map[string]interface{}
			err = json.Unmarshal(response.Result, &result)
			require.NoError(t, err)
			t.Log("✓ Non-existent file handled gracefully")
		} else {
			t.Logf("✓ Non-existent file returned error: %s", response.Error.Message)
		}
	})

	// Test 3: Large result set
	t.Run("search_large_results", func(t *testing.T) {
		// Index multiple files
		vectorstoreDir := filepath.Join(projectRoot, "internal", "vectorstore")
		indexed := indexRealCodebase(t, ctx, store, embedder, vectorstoreDir)
		require.GreaterOrEqual(t, indexed, 5, "Need at least 5 files for this test")

		argsJSON, err := json.Marshal(map[string]interface{}{
			"query": "agent",
			"top_k": 100, // Request many results
		})
		require.NoError(t, err)

		toolCallReq := map[string]interface{}{
			"name":      "context.search",
			"arguments": json.RawMessage(argsJSON),
		}

		toolCallJSON, err := json.Marshal(toolCallReq)
		require.NoError(t, err)

		request := protocol.Request{
			JSONRPC: "2.0",
			ID:      json.RawMessage(`1`),
			Method:  "tools/call",
			Params:  json.RawMessage(toolCallJSON),
		}

		requestJSON, err := json.Marshal(request)
		require.NoError(t, err)
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
				t.Fatalf("Serve error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout")
		}

		responseData := writer.Bytes()
		var response protocol.Response
		err = json.Unmarshal(responseData, &response)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(response.Result, &result)
		require.NoError(t, err)

		results := result["results"].([]interface{})
		t.Logf("✓ Large query returned %d results", len(results))

		// Verify reasonable limits applied
		assert.LessOrEqual(t, len(results), indexed, "Results should not exceed indexed docs")
	})
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
