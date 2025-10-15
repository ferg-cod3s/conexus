package sqlite

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchHybrid_Basic tests basic hybrid search functionality.
func TestSearchHybrid_Basic(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add test documents with both text content and vectors
	docs := []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "function calculate total price",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
			Metadata: map[string]interface{}{
				"language": "go",
				"path":     "calculator.go",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:      "doc2",
			Content: "compute the sum of numbers",
			Vector:  normalizeVector([]float32{0.9, 0.1, 0.0}),
			Metadata: map[string]interface{}{
				"language": "go",
				"path":     "math.go",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:      "doc3",
			Content: "database connection string",
			Vector:  normalizeVector([]float32{0.0, 1.0, 0.0}),
			Metadata: map[string]interface{}{
				"language": "go",
				"path":     "db.go",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	// Test hybrid search with both query and vector
	query := "calculate"
	queryVector := normalizeVector([]float32{1.0, 0.0, 0.0})
	
	results, err := store.SearchHybrid(context.Background(), query, queryVector, vectorstore.SearchOptions{
		Limit: 10,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, results, "Should find results")
	
	// doc1 should rank highly (matches both keyword "calculate" and vector similarity)
	assert.Equal(t, "doc1", results[0].Document.ID, "doc1 should rank first")
	assert.Equal(t, "hybrid", results[0].Method, "Method should be 'hybrid'")
	
	// All results should have scores
	for _, result := range results {
		assert.Greater(t, result.Score, float32(0), "Score should be positive")
	}
}

// TestSearchHybrid_OnlyQuery tests hybrid search with only text query (no vector).
func TestSearchHybrid_OnlyQuery(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add test documents
	docs := []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "golang programming language",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
		},
		{
			ID:      "doc2",
			Content: "python scripting tool",
			Vector:  normalizeVector([]float32{0.0, 1.0, 0.0}),
		},
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	// Search with only query text (no vector)
	results, err := store.SearchHybrid(context.Background(), "golang", nil, vectorstore.SearchOptions{
		Limit: 10,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "doc1", results[0].Document.ID, "Should find golang document")
}

// TestSearchHybrid_OnlyVector tests hybrid search with only vector (no query).
func TestSearchHybrid_OnlyVector(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add test documents
	docs := []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "some content",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
		},
		{
			ID:      "doc2",
			Content: "other content",
			Vector:  normalizeVector([]float32{0.0, 1.0, 0.0}),
		},
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	// Search with only vector (no query text)
	queryVector := normalizeVector([]float32{0.99, 0.01, 0.0})
	results, err := store.SearchHybrid(context.Background(), "", queryVector, vectorstore.SearchOptions{
		Limit: 10,
	})

	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, "doc1", results[0].Document.ID, "Should find similar vector")
}

// TestSearchHybrid_EmptyInputs tests error handling for empty inputs.
func TestSearchHybrid_EmptyInputs(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Both query and vector empty should error
	_, err := store.SearchHybrid(context.Background(), "", nil, vectorstore.SearchOptions{})
	assert.Error(t, err, "Should error with no query or vector")
	assert.Contains(t, err.Error(), "must provide either query text or query vector")
}

// TestSearchHybrid_NoResults tests hybrid search when no documents match.
func TestSearchHybrid_NoResults(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Empty store - no documents
	results, err := store.SearchHybrid(context.Background(), "test", normalizeVector([]float32{1.0, 0.0}), vectorstore.SearchOptions{
		Limit: 10,
	})

	require.NoError(t, err)
	assert.Empty(t, results, "Should return empty results for empty store")
}

// TestSearchHybrid_Ranking tests that hybrid search ranks results correctly.
func TestSearchHybrid_Ranking(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add documents with varying relevance to query and vector
	docs := []vectorstore.Document{
		{
			ID:      "exact_match",
			Content: "python programming language tutorial",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
		},
		{
			ID:      "keyword_only",
			Content: "python snake reptile animal",
			Vector:  normalizeVector([]float32{0.0, 0.0, 1.0}),
		},
		{
			ID:      "vector_only",
			Content: "coding development software",
			Vector:  normalizeVector([]float32{0.98, 0.02, 0.0}),
		},
		{
			ID:      "neither",
			Content: "unrelated content about cooking",
			Vector:  normalizeVector([]float32{0.0, 1.0, 0.0}),
		},
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	// Search with query "python programming" and vector similar to [1, 0, 0]
	query := "python programming"
	queryVector := normalizeVector([]float32{1.0, 0.0, 0.0})

	results, err := store.SearchHybrid(context.Background(), query, queryVector, vectorstore.SearchOptions{
		Limit: 10,
	})

	require.NoError(t, err)
	require.NotEmpty(t, results)

	// "exact_match" should rank highest (matches both keyword and vector)
	assert.Equal(t, "exact_match", results[0].Document.ID, "Document matching both should rank first")

	// "neither" should rank lowest or not appear
	for i, result := range results {
		if result.Document.ID == "neither" {
			assert.Greater(t, i, 0, "Unrelated document should not rank first")
		}
	}

	// Scores should be in descending order
	for i := 1; i < len(results); i++ {
		assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score, "Scores should be descending")
	}
}

// TestSearchHybrid_Limit tests that limit parameter is respected.
func TestSearchHybrid_Limit(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add many documents
	docs := make([]vectorstore.Document, 20)
	for i := 0; i < 20; i++ {
		docs[i] = vectorstore.Document{
			ID:      fmt.Sprintf("doc%d", i),
			Content: "test content with keyword search",
			Vector:  normalizeVector([]float32{float32(i) / 20.0, 1.0 - float32(i)/20.0, 0.0}),
		}
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		limit         int
		expectedCount int
	}{
		{"limit 5", 5, 5},
		{"limit 10", 10, 10},
		{"limit 15", 15, 15},
		{"limit 0 (default)", 0, 10}, // Default limit is 10
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := store.SearchHybrid(context.Background(), "keyword", normalizeVector([]float32{0.5, 0.5, 0.0}), vectorstore.SearchOptions{
				Limit: tc.limit,
			})

			require.NoError(t, err)
			assert.Len(t, results, tc.expectedCount, "Should return correct number of results")
		})
	}
}

// TestSearchHybrid_Threshold tests minimum score threshold filtering.
func TestSearchHybrid_Threshold(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add test documents
	docs := []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "exact match content",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
		},
		{
			ID:      "doc2",
			Content: "partial match",
			Vector:  normalizeVector([]float32{0.5, 0.5, 0.0}),
		},
		{
			ID:      "doc3",
			Content: "weak match",
			Vector:  normalizeVector([]float32{0.1, 0.1, 0.98}),
		},
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	// Search with high threshold
	results, err := store.SearchHybrid(context.Background(), "exact match", normalizeVector([]float32{1.0, 0.0, 0.0}), vectorstore.SearchOptions{
		Limit:     10,
		Threshold: 0.015, // High threshold filters out weak matches
	})

	require.NoError(t, err)
	assert.NotEmpty(t, results)

	// All results should meet threshold
	for _, result := range results {
		assert.GreaterOrEqual(t, result.Score, float32(0.015), "All results should meet threshold")
	}
}

// TestSearchHybrid_Filters tests metadata filtering.
func TestSearchHybrid_Filters(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add documents with different languages
	docs := []vectorstore.Document{
		{
			ID:      "go_doc",
			Content: "golang function implementation",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
			Metadata: map[string]interface{}{
				"language": "go",
			},
		},
		{
			ID:      "py_doc",
			Content: "python function implementation",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
			Metadata: map[string]interface{}{
				"language": "python",
			},
		},
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	// Search with language filter
	results, err := store.SearchHybrid(context.Background(), "function", normalizeVector([]float32{1.0, 0.0, 0.0}), vectorstore.SearchOptions{
		Limit: 10,
		Filters: map[string]interface{}{
			"language": "go",
		},
	})

	require.NoError(t, err)
	require.Len(t, results, 1, "Should only find filtered documents")
	assert.Equal(t, "go_doc", results[0].Document.ID)
	assert.Equal(t, "go", results[0].Document.Metadata["language"])
}

// TestSearchHybrid_Overlapping tests deduplication when documents appear in both BM25 and vector results.
func TestSearchHybrid_Overlapping(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add documents that will match both BM25 and vector search
	docs := []vectorstore.Document{
		{
			ID:      "overlap_doc",
			Content: "machine learning algorithm",
			Vector:  normalizeVector([]float32{1.0, 0.0, 0.0}),
		},
		{
			ID:      "other_doc",
			Content: "database query optimization",
			Vector:  normalizeVector([]float32{0.0, 1.0, 0.0}),
		},
	}

	err := store.UpsertBatch(context.Background(), docs)
	require.NoError(t, err)

	// Search with query and vector that both match "overlap_doc"
	query := "machine learning"
	queryVector := normalizeVector([]float32{1.0, 0.0, 0.0})

	results, err := store.SearchHybrid(context.Background(), query, queryVector, vectorstore.SearchOptions{
		Limit: 10,
	})

	require.NoError(t, err)
	
	// Count occurrences of each document ID
	idCount := make(map[string]int)
	for _, result := range results {
		idCount[result.Document.ID]++
	}

	// Each document should appear exactly once
	for id, count := range idCount {
		assert.Equal(t, 1, count, "Document %s should appear exactly once", id)
	}
}

// TestSearchHybrid_ContextCancellation tests context cancellation during search.
func TestSearchHybrid_ContextCancellation(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Add a document
	err := store.UpsertBatch(context.Background(), []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "test content",
			Vector:  normalizeVector([]float32{1.0, 0.0}),
		},
	})
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Search should handle cancellation
	_, err = store.SearchHybrid(ctx, "test", normalizeVector([]float32{1.0, 0.0}), vectorstore.SearchOptions{})
	assert.Error(t, err, "Should error on cancelled context")
}

// TestApplyRRF_AlphaWeights tests different alpha values for weighting.
func TestApplyRRF_AlphaWeights(t *testing.T) {
	// Mock results
	bm25Results := []vectorstore.SearchResult{
		{Document: vectorstore.Document{ID: "doc1"}, Score: 10.0, Method: "bm25"},
		{Document: vectorstore.Document{ID: "doc2"}, Score: 8.0, Method: "bm25"},
	}
	vectorResults := []vectorstore.SearchResult{
		{Document: vectorstore.Document{ID: "doc2"}, Score: 0.9, Method: "vector"},
		{Document: vectorstore.Document{ID: "doc3"}, Score: 0.8, Method: "vector"},
	}

	testCases := []struct {
		name        string
		alpha       float32
		expectFirst string
	}{
		{"alpha=0.0 (BM25 only)", 0.0, "doc1"},
		{"alpha=0.5 (equal)", 0.5, "doc2"},
		{"alpha=1.0 (vector only)", 1.0, "doc2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := HybridSearchOptions{
				Alpha: tc.alpha,
				K:     60,
			}

			results := applyRRF(bm25Results, vectorResults, opts)
			
			require.NotEmpty(t, results)
			assert.Equal(t, tc.expectFirst, results[0].Document.ID, "First result should match expected")
			
			// All results should have method="hybrid"
			for _, result := range results {
				assert.Equal(t, "hybrid", result.Method)
			}
		})
	}
}

// TestApplyRRF_EmptyResults tests RRF with empty result sets.
func TestApplyRRF_EmptyResults(t *testing.T) {
	bm25Results := []vectorstore.SearchResult{
		{Document: vectorstore.Document{ID: "doc1"}, Score: 10.0},
	}
	vectorResults := []vectorstore.SearchResult{}

	opts := HybridSearchOptions{
		Alpha: 0.5,
		K:     60,
	}

	// RRF should handle empty vector results gracefully
	results := applyRRF(bm25Results, vectorResults, opts)
	
	require.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)
}

// TestLimitResults tests the limitResults helper function.
func TestLimitResults(t *testing.T) {
	results := []vectorstore.SearchResult{
		{Document: vectorstore.Document{ID: "doc1"}},
		{Document: vectorstore.Document{ID: "doc2"}},
		{Document: vectorstore.Document{ID: "doc3"}},
		{Document: vectorstore.Document{ID: "doc4"}},
		{Document: vectorstore.Document{ID: "doc5"}},
	}

	testCases := []struct {
		name     string
		limit    int
		expected int
	}{
		{"limit less than results", 3, 3},
		{"limit equal to results", 5, 5},
		{"limit greater than results", 10, 5},
		{"limit zero", 0, 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limited := limitResults(results, tc.limit)
			assert.Len(t, limited, tc.expected)
		})
	}
}

// normalizeVector normalizes a vector to unit length for cosine similarity.
func normalizeVector(v []float32) []float32 {
	var magnitude float32
	for _, val := range v {
		magnitude += val * val
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))
	
	if magnitude == 0 {
		return v
	}
	
	normalized := make([]float32, len(v))
	for i, val := range v {
		normalized[i] = val / magnitude
	}
	return normalized
}

// setupTestStore creates a new test store with a temporary database.
func setupTestStore(t *testing.T) *Store {
	t.Helper()
	
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	store, err := NewStore(dbPath)
	require.NoError(t, err, "Failed to create test store")
	
	return store
}
