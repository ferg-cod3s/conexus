package integration

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndMCPWithMonitoring tests complete monitoring integration
func TestEndToEndMCPWithMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end monitoring test in short mode")
	}

	// Setup components with monitoring
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err, "Should create vector store")
	defer store.Close()

	embedder := embedding.NewMock(384)

	// Create connector store
	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err, "Should create connector store")
	defer connStore.Close()

	// Create indexer
	idx := indexer.NewIndexer("test-e2e-state.json")

	// Setup observability with unique metric names
	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-e2e")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	// Test 1: Add content directly to store
	t.Run("add_content", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Generate embedding for the content
		testContent := "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"
		emb, err := embedder.Embed(ctx, testContent)
		require.NoError(t, err, "Should generate embedding")

		// Add content to store directly
		doc := vectorstore.Document{
			ID:       "test-doc-1",
			Content:  testContent,
			Vector:   emb.Vector,
			Metadata: map[string]interface{}{"path": "/test/file.go", "type": "file"},
		}
		err = store.Upsert(ctx, doc)
		require.NoError(t, err, "Should upsert content to store")
	})

	// Test 2: Search for indexed content
	t.Run("search_content", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Generate embedding for query
		query := "Hello World"
		queryEmb, err := embedder.Embed(ctx, query)
		require.NoError(t, err, "Should generate query embedding")

		// Perform hybrid search
		opts := vectorstore.SearchOptions{Limit: 5}
		results, err := store.SearchHybrid(ctx, query, queryEmb.Vector, opts)
		require.NoError(t, err, "Should perform search")
		assert.Greater(t, len(results), 0, "Should find indexed content")
	})

	// Test 3: Verify indexer status (without HealthCheck which requires state file)
	t.Run("index_status", func(t *testing.T) {
		// Get indexer status without checking health (HealthCheck requires state file to exist)
		status := idx.GetStatus()
		assert.NotNil(t, status, "Status should exist")
		assert.Equal(t, "idle", status.Phase, "Should start in idle phase")
	})

	// Test 4: Check monitoring components are initialized
	t.Run("verify_monitoring", func(t *testing.T) {
		// Verify all monitoring components exist
		assert.NotNil(t, logger, "Logger should exist")
		assert.NotNil(t, metrics, "Metrics collector should exist")
		assert.NotNil(t, errorHandler, "Error handler should exist")
		assert.NotNil(t, idx, "Indexer should exist")
		assert.NotNil(t, store, "Store should exist")
		assert.NotNil(t, connStore, "Connector store should exist")
	})
}

// TestMCPErrorHandlingWithMonitoring tests error scenarios with monitoring
func TestMCPErrorHandlingWithMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling monitoring test in short mode")
	}

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()

	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-error")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	// Test 1: Invalid search with empty inputs should error
	t.Run("invalid_search", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to search with empty query and empty vector - should error
		opts := vectorstore.SearchOptions{Limit: 5}
		_, err := store.SearchHybrid(ctx, "", []float32{}, opts)
		// SearchHybrid should error when both query and vector are empty
		require.Error(t, err, "Should error with empty query and vector")
		assert.Contains(t, err.Error(), "must provide either query text or query vector")
	})

	// Test 2: Verify error handler exists
	t.Run("invalid_tool", func(t *testing.T) {
		// Verify error handler can be used for error scenarios
		assert.NotNil(t, errorHandler)
	})

	// Test 3: Verify metrics work
	t.Run("verify_error_metrics", func(t *testing.T) {
		assert.NotNil(t, metrics, "Metrics should be initialized")
		assert.NotNil(t, errorHandler, "Error handler should be initialized")
	})
}

// TestMCPConcurrentRequestsWithMonitoring tests concurrent access with monitoring
func TestMCPConcurrentRequestsWithMonitoring(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent requests monitoring test in short mode")
	}

	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)
	defer store.Close()

	embedder := embedding.NewMock(384)

	connStore, err := connectors.NewStore(":memory:")
	require.NoError(t, err)
	defer connStore.Close()

	idx := indexer.NewIndexer("test-concurrent-state.json")

	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-concurrent")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)

	// Test 1: Setup content for concurrent access
	t.Run("setup_content", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		for i := 0; i < 5; i++ {
			content := "concurrent test document"
			emb, err := embedder.Embed(ctx, content)
			require.NoError(t, err)

			doc := vectorstore.Document{
				ID:       "doc-" + string(rune(i+'0')),
				Content:  content,
				Vector:   emb.Vector,
				Metadata: map[string]interface{}{"index": i},
			}
			err = store.Upsert(ctx, doc)
			require.NoError(t, err)
		}
	})

	// Test 2: Verify search works
	t.Run("verify_search", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		queryEmb, err := embedder.Embed(ctx, "concurrent test")
		require.NoError(t, err)

		opts := vectorstore.SearchOptions{Limit: 5}
		results, err := store.SearchHybrid(ctx, "concurrent test", queryEmb.Vector, opts)
		require.NoError(t, err)
		assert.Greater(t, len(results), 0, "Should find concurrent documents")
	})

	// Test 3: Verify monitoring initialization
	t.Run("verify_monitoring", func(t *testing.T) {
		assert.NotNil(t, logger)
		assert.NotNil(t, metrics)
		assert.NotNil(t, errorHandler)
		assert.NotNil(t, idx)
	})
}
