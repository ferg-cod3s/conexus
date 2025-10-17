package vectorstore

import (
	"fmt"
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	assert.NotNil(t, store)
	assert.NotNil(t, store.documents)
	assert.NotNil(t, store.index)

	count, err := store.Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestMemoryStore_Upsert(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	doc := Document{
		ID:      "doc1",
		Content: "Hello world",
		Vector:  embedding.Vector{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{
			"language": "go",
		},
	}

	// Insert
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Verify count
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Retrieve
	retrieved, err := store.Get(ctx, "doc1")
	require.NoError(t, err)
	assert.Equal(t, "doc1", retrieved.ID)
	assert.Equal(t, "Hello world", retrieved.Content)
	assert.Equal(t, "go", retrieved.Metadata["language"])

	// Update
	doc.Content = "Updated content"
	err = store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Verify update
	retrieved, err = store.Get(ctx, "doc1")
	require.NoError(t, err)
	assert.Equal(t, "Updated content", retrieved.Content)

	// Count should still be 1
	count, err = store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMemoryStore_Upsert_Validation(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	tests := []struct {
		name    string
		doc     Document
		wantErr string
	}{
		{
			name: "empty ID",
			doc: Document{
				Content: "test",
				Vector:  embedding.Vector{0.1, 0.2},
			},
			wantErr: "document ID cannot be empty",
		},
		{
			name: "empty vector",
			doc: Document{
				ID:      "doc1",
				Content: "test",
				Vector:  embedding.Vector{},
			},
			wantErr: "document vector cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Upsert(ctx, tt.doc)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMemoryStore_UpsertBatch(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	docs := []Document{
		{
			ID:      "doc1",
			Content: "First document",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
		},
		{
			ID:      "doc2",
			Content: "Second document",
			Vector:  embedding.Vector{0.4, 0.5, 0.6},
		},
		{
			ID:      "doc3",
			Content: "Third document",
			Vector:  embedding.Vector{0.7, 0.8, 0.9},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	doc := Document{
		ID:      "doc1",
		Content: "To be deleted",
		Vector:  embedding.Vector{0.1, 0.2, 0.3},
	}

	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Delete
	err = store.Delete(ctx, "doc1")
	require.NoError(t, err)

	// Verify deletion
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Try to retrieve deleted doc
	_, err = store.Get(ctx, "doc1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Delete non-existent doc
	err = store.Delete(ctx, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemoryStore_Get(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	doc := Document{
		ID:      "doc1",
		Content: "Test content",
		Vector:  embedding.Vector{0.1, 0.2, 0.3},
		Metadata: map[string]interface{}{
			"language": "go",
			"path":     "/test/file.go",
		},
	}

	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	retrieved, err := store.Get(ctx, "doc1")
	require.NoError(t, err)
	assert.Equal(t, doc.ID, retrieved.ID)
	assert.Equal(t, doc.Content, retrieved.Content)
	assert.Equal(t, doc.Vector, retrieved.Vector)
	assert.Equal(t, "go", retrieved.Metadata["language"])

	// Get non-existent
	_, err = store.Get(ctx, "nonexistent")
	require.Error(t, err)
}

func TestMemoryStore_SearchVector(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Insert test documents with similar vectors
	docs := []Document{
		{
			ID:      "doc1",
			Content: "Machine learning algorithms",
			Vector:  embedding.Vector{0.9, 0.1, 0.1}, // Similar to query
			Metadata: map[string]interface{}{
				"language": "go",
			},
		},
		{
			ID:      "doc2",
			Content: "Deep neural networks",
			Vector:  embedding.Vector{0.8, 0.2, 0.1}, // Somewhat similar
			Metadata: map[string]interface{}{
				"language": "python",
			},
		},
		{
			ID:      "doc3",
			Content: "Cooking recipes",
			Vector:  embedding.Vector{0.1, 0.9, 0.1}, // Different
			Metadata: map[string]interface{}{
				"language": "go",
			},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	// Search
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, SearchOptions{
		Limit: 2,
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// Results should be ordered by similarity
	assert.Equal(t, "doc1", results[0].Document.ID)
	assert.Greater(t, results[0].Score, results[1].Score)
	assert.Equal(t, "vector", results[0].Method)
}

func TestMemoryStore_SearchVector_WithFilters(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	docs := []Document{
		{
			ID:      "doc1",
			Content: "Go code",
			Vector:  embedding.Vector{0.9, 0.1, 0.1},
			Metadata: map[string]interface{}{
				"language": "go",
			},
		},
		{
			ID:      "doc2",
			Content: "Python code",
			Vector:  embedding.Vector{0.8, 0.2, 0.1},
			Metadata: map[string]interface{}{
				"language": "python",
			},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	// Search with language filter
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, SearchOptions{
		Limit: 10,
		Filters: map[string]interface{}{
			"language": "go",
		},
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)
}

func TestMemoryStore_SearchVector_WithThreshold(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	docs := []Document{
		{
			ID:      "doc1",
			Content: "Very similar",
			Vector:  embedding.Vector{0.99, 0.01, 0.01},
		},
		{
			ID:      "doc2",
			Content: "Not similar",
			Vector:  embedding.Vector{0.1, 0.9, 0.1},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	// High threshold should filter out low-scoring results
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, SearchOptions{
		Limit:     10,
		Threshold: 0.9, // High threshold
	})
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)
}

func TestMemoryStore_SearchBM25(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	docs := []Document{
		{
			ID:      "doc1",
			Content: "machine learning is a subset of artificial intelligence",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
		},
		{
			ID:      "doc2",
			Content: "deep learning uses neural networks for machine learning",
			Vector:  embedding.Vector{0.4, 0.5, 0.6},
		},
		{
			ID:      "doc3",
			Content: "cooking pasta with tomato sauce",
			Vector:  embedding.Vector{0.7, 0.8, 0.9},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	// Search for machine learning
	results, err := store.SearchBM25(ctx, "machine learning", SearchOptions{
		Limit: 2,
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)

	// doc2 has "machine learning" and more context, should rank higher or similar to doc1
	assert.Contains(t, []string{"doc1", "doc2"}, results[0].Document.ID)
	assert.Equal(t, "bm25", results[0].Method)
}

func TestMemoryStore_SearchBM25_EmptyQuery(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	_, err := store.SearchBM25(ctx, "", SearchOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "query cannot be empty")
}

func TestMemoryStore_SearchHybrid(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	docs := []Document{
		{
			ID:      "doc1",
			Content: "machine learning algorithms",
			Vector:  embedding.Vector{0.9, 0.1, 0.1}, // Good vector match
		},
		{
			ID:      "doc2",
			Content: "deep learning neural networks",
			Vector:  embedding.Vector{0.1, 0.9, 0.1}, // Poor vector match, but has keywords
		},
		{
			ID:      "doc3",
			Content: "cooking recipes",
			Vector:  embedding.Vector{0.1, 0.1, 0.9}, // Poor match overall
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	// Hybrid search should combine vector and keyword signals
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchHybrid(ctx, "machine learning", queryVector, SearchOptions{
		Limit: 2,
	})
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "hybrid", results[0].Method)

	// Should rank doc1 higher (good on both vector and keywords)
	assert.Equal(t, "doc1", results[0].Document.ID)
}

func TestMemoryStore_Stats(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Empty store
	stats, err := store.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.TotalDocuments)
	assert.Equal(t, int64(0), stats.TotalChunks)

	// Add documents
	docs := []Document{
		{
			ID:      "doc1",
			Content: "Go code",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"language": "go",
			},
		},
		{
			ID:      "doc2",
			Content: "Python code",
			Vector:  embedding.Vector{0.4, 0.5, 0.6},
			Metadata: map[string]interface{}{
				"language": "python",
			},
		},
		{
			ID:      "doc3",
			Content: "More Go code",
			Vector:  embedding.Vector{0.7, 0.8, 0.9},
			Metadata: map[string]interface{}{
				"language": "go",
			},
		},
	}

	err = store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	// Check stats
	stats, err = store.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.TotalDocuments)
	assert.Equal(t, int64(3), stats.TotalChunks)
	assert.Equal(t, int64(2), stats.Languages["go"])
	assert.Equal(t, int64(1), stats.Languages["python"])
	assert.False(t, stats.LastIndexedAt.IsZero())
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        embedding.Vector
		b        embedding.Vector
		expected float32
	}{
		{
			name:     "identical vectors",
			a:        embedding.Vector{1.0, 0.0, 0.0},
			b:        embedding.Vector{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        embedding.Vector{1.0, 0.0, 0.0},
			b:        embedding.Vector{0.0, 1.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "opposite vectors",
			a:        embedding.Vector{1.0, 0.0, 0.0},
			b:        embedding.Vector{-1.0, 0.0, 0.0},
			expected: -1.0,
		},
		{
			name:     "different lengths",
			a:        embedding.Vector{1.0, 0.0},
			b:        embedding.Vector{1.0, 0.0, 0.0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cosineSimilarity(tt.a, tt.b)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestMatchesFilters(t *testing.T) {
	doc := Document{
		Metadata: map[string]interface{}{
			"language": "go",
			"path":     "/test/file.go",
		},
	}

	tests := []struct {
		name     string
		filters  map[string]interface{}
		expected bool
	}{
		{
			name:     "no filters",
			filters:  map[string]interface{}{},
			expected: true,
		},
		{
			name: "matching filter",
			filters: map[string]interface{}{
				"language": "go",
			},
			expected: true,
		},
		{
			name: "non-matching filter",
			filters: map[string]interface{}{
				"language": "python",
			},
			expected: false,
		},
		{
			name: "multiple matching filters",
			filters: map[string]interface{}{
				"language": "go",
				"path":     "/test/file.go",
			},
			expected: true,
		},
		{
			name: "one non-matching filter",
			filters: map[string]interface{}{
				"language": "go",
				"path":     "/other/file.go",
			},
			expected: false,
		},
		{
			name: "filter key not in metadata",
			filters: map[string]interface{}{
				"nonexistent": "value",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilters(doc, tt.filters)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "simple text",
			text:     "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with punctuation",
			text:     "Hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "multiple spaces",
			text:     "hello    world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "empty string",
			text:     "",
			expected: nil, // Accept nil for empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tokenize(tt.text)
			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMemoryStore_Concurrency(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Concurrent writes
	const numGoroutines = 10
	const docsPerGoroutine = 10

	errChan := make(chan error, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(offset int) {
			for j := 0; j < docsPerGoroutine; j++ {
				doc := Document{
					ID:      fmt.Sprintf("doc-%d-%d", offset, j),
					Content: "Test content",
					Vector:  embedding.Vector{0.1, 0.2, 0.3},
				}
				if err := store.Upsert(ctx, doc); err != nil {
					errChan <- err
					return
				}
			}
			errChan <- nil
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		require.NoError(t, err)
	}

	// Verify count
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(numGoroutines*docsPerGoroutine), count)
}

func TestMemoryStore_Timestamps(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	doc := Document{
		ID:      "doc1",
		Content: "Test",
		Vector:  embedding.Vector{0.1, 0.2, 0.3},
	}

	// Insert
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	retrieved, err := store.Get(ctx, "doc1")
	require.NoError(t, err)
	assert.False(t, retrieved.CreatedAt.IsZero())
	assert.False(t, retrieved.UpdatedAt.IsZero())
	createdAt := retrieved.CreatedAt

	// Update after small delay
	time.Sleep(10 * time.Millisecond)
	doc.Content = "Updated"
	err = store.Upsert(ctx, doc)
	require.NoError(t, err)

	retrieved, err = store.Get(ctx, "doc1")
	require.NoError(t, err)
	assert.Equal(t, createdAt, retrieved.CreatedAt) // Should not change
	assert.True(t, retrieved.UpdatedAt.After(createdAt))
}
