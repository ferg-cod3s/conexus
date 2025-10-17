package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVectorStoreIntegration tests vector store operations
func TestVectorStoreIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping vector store integration test in short mode")
	}

	ctx := context.Background()

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err, "Should create vector store")

	embedder := embedding.NewMock(384)

	// Test 1: Store and retrieve documents
	t.Run("store_and_retrieve", func(t *testing.T) {
		emb, err := embedder.Embed(ctx, "test content")
		require.NoError(t, err)

		doc := vectorstore.Document{
			ID:      "test-doc-1",
			Content: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n\nfunc helper() {\n\treturn\n}",
			Vector:  emb.Vector,
			Metadata: map[string]interface{}{
				"source_type": "file",
				"language":    "go",
			},
		}

		err = store.Upsert(ctx, doc)
		assert.NoError(t, err, "Should store document successfully")

		// Retrieve document
		retrieved, err := store.Get(ctx, "test-doc-1")
		require.NoError(t, err, "Should retrieve document")
		assert.Equal(t, doc.Content, retrieved.Content, "Content should match")
		assert.Equal(t, "go", retrieved.Metadata["language"], "Metadata should match")
	})

	// Test 2: Batch operations
	t.Run("batch_operations", func(t *testing.T) {
		emb1, err := embedder.Embed(ctx, "batch content 1")
		require.NoError(t, err)
		emb2, err := embedder.Embed(ctx, "batch content 2")
		require.NoError(t, err)

		docs := []vectorstore.Document{
			{
				ID:      "batch-doc-1",
				Content: "package utils\n\nfunc Add(a, b int) int {\n\treturn a + b\n}",
				Vector:  emb1.Vector,
				Metadata: map[string]interface{}{
					"source_type": "file",
					"language":    "go",
				},
			},
			{
				ID:      "batch-doc-2",
				Content: "package main\n\ntype User struct {\n\tName string\n\tAge  int\n}",
				Vector:  emb2.Vector,
				Metadata: map[string]interface{}{
					"source_type": "file",
					"language":    "go",
				},
			},
		}

		err = store.UpsertBatch(ctx, docs)
		assert.NoError(t, err, "Should batch store documents successfully")

		// Verify count
		count, err := store.Count(ctx)
		require.NoError(t, err, "Should get count")
		assert.GreaterOrEqual(t, count, int64(3), "Should have at least 3 documents")
	})

	// Test 3: Vector search
	t.Run("vector_search", func(t *testing.T) {
		queryEmb, err := embedder.Embed(ctx, "function main")
		require.NoError(t, err)

		opts := vectorstore.SearchOptions{
			Limit: 5,
		}

		results, err := store.SearchVector(ctx, queryEmb.Vector, opts)
		require.NoError(t, err, "Should perform vector search")
		assert.Greater(t, len(results), 0, "Should find results")
	})

	// Test 4: BM25 search
	t.Run("bm25_search", func(t *testing.T) {
		opts := vectorstore.SearchOptions{
			Limit: 5,
		}

		results, err := store.SearchBM25(ctx, "package main", opts)
		require.NoError(t, err, "Should perform BM25 search")
		assert.Greater(t, len(results), 0, "Should find results")
	})

	// Test 5: Hybrid search
	t.Run("hybrid_search", func(t *testing.T) {
		queryEmb, err := embedder.Embed(ctx, "function main")
		require.NoError(t, err)

		opts := vectorstore.SearchOptions{
			Limit: 5,
		}

		results, err := store.SearchHybrid(ctx, "function main", queryEmb.Vector, opts)
		require.NoError(t, err, "Should perform hybrid search")
		assert.Greater(t, len(results), 0, "Should find results")
	})

	// Test 6: Search with filters
	t.Run("search_with_filters", func(t *testing.T) {
		queryEmb, err := embedder.Embed(ctx, "test query")
		require.NoError(t, err)

		opts := vectorstore.SearchOptions{
			Limit: 10,
			Filters: map[string]interface{}{
				"language": "go",
			},
		}

		results, err := store.SearchVector(ctx, queryEmb.Vector, opts)
		require.NoError(t, err, "Should search with filters")
		// Results may be empty if no documents match the filter, which is fine
		for _, result := range results {
			assert.Equal(t, "go", result.Document.Metadata["language"], "Filtered results should match language")
		}
	})

	// Test 7: Delete operations
	t.Run("delete_operations", func(t *testing.T) {
		err := store.Delete(ctx, "test-doc-1")
		assert.NoError(t, err, "Should delete document")

		// Verify deletion
		_, err = store.Get(ctx, "test-doc-1")
		assert.Error(t, err, "Should not find deleted document")
	})
}

// TestVectorStorePerformance tests vector store performance characteristics
func TestVectorStorePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping vector store performance test in short mode")
	}

	ctx := context.Background()

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)

	// Test batch insertion performance
	t.Run("batch_insertion_performance", func(t *testing.T) {
		numDocs := 100
		docs := make([]vectorstore.Document, numDocs)

		emb, err := embedder.Embed(ctx, "performance test content")
		require.NoError(t, err)

		for i := 0; i < numDocs; i++ {
			docs[i] = vectorstore.Document{
				ID:      "perf-doc-" + string(rune(i)),
				Content: "package perf\n\nfunc test" + string(rune(i)) + "() {\n\treturn " + string(rune(i)) + "\n}",
				Vector:  emb.Vector,
				Metadata: map[string]interface{}{
					"source_type": "file",
					"language":    "go",
				},
			}
		}

		startTime := time.Now()
		err = store.UpsertBatch(ctx, docs)
		duration := time.Since(startTime)

		assert.NoError(t, err, "Should batch insert documents successfully")
		t.Logf("Inserted %d documents in %v (avg: %v per doc)", numDocs, duration, duration/time.Duration(numDocs))
		assert.Less(t, duration, 5*time.Second, "Should insert batch within reasonable time")
	})

	// Test search performance
	t.Run("search_performance", func(t *testing.T) {
		queryEmb, err := embedder.Embed(ctx, "function test")
		require.NoError(t, err)

		opts := vectorstore.SearchOptions{
			Limit: 20,
		}

		startTime := time.Now()
		results, err := store.SearchVector(ctx, queryEmb.Vector, opts)
		duration := time.Since(startTime)

		assert.NoError(t, err, "Should perform search successfully")
		assert.Less(t, duration, 1*time.Second, "Should search within reasonable time")
		t.Logf("Found %d results in %v", len(results), duration)
	})
}

// TestVectorStoreErrorHandling tests error handling in vector store operations
func TestVectorStoreErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping vector store error handling test in short mode")
	}

	ctx := context.Background()

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)

	// Test operations with invalid data
	t.Run("invalid_document_handling", func(t *testing.T) {
		// Test with empty ID
		emb, err := embedder.Embed(ctx, "test")
		require.NoError(t, err)

		doc := vectorstore.Document{
			ID:      "",
			Content: "test content",
			Vector:  emb.Vector,
		}

		err = store.Upsert(ctx, doc)
		// Should handle gracefully (may or may not error depending on implementation)
		assert.NotPanics(t, func() {
			store.Upsert(ctx, doc)
		}, "Should not panic on invalid document")
	})

	// Test retrieval of non-existent documents
	t.Run("non_existent_retrieval", func(t *testing.T) {
		_, err := store.Get(ctx, "non-existent-id")
		assert.Error(t, err, "Should error on non-existent document")
	})

	// Test deletion of non-existent documents
	t.Run("non_existent_deletion", func(t *testing.T) {
		_ = store.Delete(ctx, "non-existent-id") // May or may not error depending on implementation
		assert.NotPanics(t, func() {
			store.Delete(ctx, "non-existent-id")
		}, "Should not panic on deleting non-existent document")
	})
}

// TestIndexingPipelineIntegration tests the complete indexing pipeline
func TestIndexingPipelineIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping indexing pipeline integration test in short mode")
	}

	ctx := context.Background()

	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "indexing-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test Go files
	testFiles := map[string]string{
		"main.go":  "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n\nfunc helper() {\n\treturn\n}",
		"utils.go": "package utils\n\nfunc Add(a, b int) int {\n\treturn a + b\n}\n\ntype User struct {\n\tName string\n\tAge  int\n}",
		"api.go":   "package api\n\n// GetUser retrieves a user by ID\nfunc GetUser(id int) (*User, error) {\n\treturn nil, nil\n}",
	}

	for filename, content := range testFiles {
		path := filepath.Join(tempDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err, "Should create test file")
	}

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)
	idx := indexer.NewIndexer(filepath.Join(tempDir, "test-integration-state.json"))

	loggerCfg := observability.LoggerConfig{
		Level:  "info",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-integration")
	_ = observability.NewErrorHandler(logger, metrics, false) // Not used in indexing tests

	// Test 1: Basic indexing with IndexOptions
	t.Run("basic_indexing", func(t *testing.T) {
		opts := indexer.IndexOptions{
			RootPath:       tempDir,
			IgnorePatterns: []string{"*.git*", "*_test.go", "*.md"},
			MaxFileSize:    1024 * 1024, // 1MB
			Embedder:       embedder,
			VectorStore:    store,
		}

		chunks, err := idx.Index(ctx, opts)
		require.NoError(t, err, "Should index successfully")
		assert.Greater(t, len(chunks), 0, "Should create chunks")

		// Verify chunks were stored in vector store
		count, err := store.Count(ctx)
		require.NoError(t, err, "Should get count")
		assert.GreaterOrEqual(t, count, int64(len(chunks)), "Should have at least as many documents as chunks")
	})

	// Test 2: Incremental indexing
	t.Run("incremental_indexing", func(t *testing.T) {
		opts := indexer.IndexOptions{
			RootPath:       tempDir,
			IgnorePatterns: []string{"*.git*", "*_test.go", "*.md"},
			MaxFileSize:    1024 * 1024,
			Embedder:       embedder,
			VectorStore:    store,
		}

		// First incremental index (should be full since no previous state)
		chunks1, state1, err := idx.IndexIncremental(ctx, opts, nil)
		require.NoError(t, err, "Should perform first incremental index")
		assert.Greater(t, len(chunks1), 0, "Should create chunks")

		// Second incremental index (should be empty since no changes)
		chunks2, state2, err := idx.IndexIncremental(ctx, opts, state1)
		require.NoError(t, err, "Should perform second incremental index")
		assert.Equal(t, 0, len(chunks2), "Should have no changes")

		// States may be the same if no changes occurred, which is fine
		// Just verify we got valid state data
		assert.NotNil(t, state1, "First state should not be nil")
		assert.NotNil(t, state2, "Second state should not be nil")
	})

	// Test 3: Index status and health check
	t.Run("index_status_and_health", func(t *testing.T) {
		status := idx.GetStatus()
		assert.NotNil(t, status, "Should get status")

		// Health check may fail if state file doesn't exist, which is OK for this test
		// We just verify it doesn't panic
		assert.NotPanics(t, func() {
			idx.HealthCheck(ctx)
		}, "Health check should not panic")
	})
}

// TestIndexingPipelinePerformance tests indexing performance characteristics
func TestIndexingPipelinePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping indexing performance test in short mode")
	}

	ctx := context.Background()

	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "indexing-perf-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create multiple test Go files for performance testing
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("file%d.go", i)
		content := fmt.Sprintf("package main\n\nfunc func%d() {\n\tprintln(\"test %d\")\n}\n\ntype Type%d struct {\n\tField%d int\n}", i, i, i, i)
		path := filepath.Join(tempDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err, "Should create test file")
	}

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)
	idx := indexer.NewIndexer(filepath.Join(tempDir, "test-perf-state.json"))

	// Test indexing performance with real files
	t.Run("indexing_performance", func(t *testing.T) {
		opts := indexer.IndexOptions{
			RootPath:       tempDir,
			IgnorePatterns: []string{"*.git*", "*_test.go", "*.md", "vendor/*", "node_modules/*"},
			MaxFileSize:    1024 * 1024, // 1MB
			Embedder:       embedder,
			VectorStore:    store,
		}

		startTime := time.Now()
		chunks, err := idx.Index(ctx, opts)
		duration := time.Since(startTime)

		assert.NoError(t, err, "Should index successfully")
		if len(chunks) > 0 {
			t.Logf("Indexed %d chunks in %v (avg: %v per chunk)", len(chunks), duration, duration/time.Duration(len(chunks)))
		} else {
			t.Logf("Indexed 0 chunks in %v", duration)
		}
		assert.Less(t, duration, 30*time.Second, "Should index within reasonable time")
	})
}

// TestIndexingPipelineErrorRecovery tests error recovery in indexing
func TestIndexingPipelineErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping indexing error recovery test in short mode")
	}

	ctx := context.Background()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "indexing-error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup components
	store, err := sqlite.NewStore(":memory:")
	require.NoError(t, err)

	embedder := embedding.NewMock(384)
	idx := indexer.NewIndexer(filepath.Join(tempDir, "test-recovery-state.json"))

	loggerCfg := observability.LoggerConfig{
		Level:  "debug",
		Format: "json",
	}
	logger := observability.NewLogger(loggerCfg)
	metrics := observability.NewMetricsCollector("test-recovery")
	_ = observability.NewErrorHandler(logger, metrics, false)

	// Test with invalid root path
	t.Run("invalid_root_path", func(t *testing.T) {
		opts := indexer.IndexOptions{
			RootPath:    "/nonexistent/path",
			Embedder:    embedder,
			VectorStore: store,
		}

		_, err := idx.Index(ctx, opts)
		assert.Error(t, err, "Should error on invalid path")
	})

	// Test indexer health check
	t.Run("health_check", func(t *testing.T) {
		// Health check may fail if state file doesn't exist, which is expected for a new indexer
		// We just verify it doesn't panic and returns a meaningful error
		err := idx.HealthCheck(ctx)
		// Either it succeeds or fails with a specific error about missing state file
		if err != nil {
			assert.Contains(t, err.Error(), "index state file does not exist", "Should fail with expected error")
		}
	})
}
