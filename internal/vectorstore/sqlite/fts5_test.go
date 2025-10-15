package sqlite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

func TestSearchBM25_Basic(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Insert test documents
	docs := []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "Go is a programming language designed for simplicity",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"language": "go",
				"file":     "intro.go",
			},
		},
		{
			ID:      "doc2",
			Content: "Python is a versatile programming language",
			Vector:  embedding.Vector{0.2, 0.3, 0.4},
			Metadata: map[string]interface{}{
				"language": "python",
				"file":     "intro.py",
			},
		},
		{
			ID:      "doc3",
			Content: "Go excels at concurrent programming with goroutines",
			Vector:  embedding.Vector{0.3, 0.4, 0.5},
			Metadata: map[string]interface{}{
				"language": "go",
				"file":     "concurrency.go",
			},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	t.Run("simple query", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "programming", vectorstore.SearchOptions{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.Len(t, results, 3) // All docs contain "programming"
		
		// Verify all results have positive scores
		for _, r := range results {
			assert.Greater(t, r.Score, float32(0))
			assert.Equal(t, "bm25", r.Method)
		}
	})

	t.Run("specific term query", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "goroutines", vectorstore.SearchOptions{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "doc3", results[0].Document.ID)
		assert.Contains(t, results[0].Document.Content, "goroutines")
	})

	t.Run("multi-word query", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "Go programming", vectorstore.SearchOptions{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2) // At least doc1 and doc3
		
		// First result should have highest score
		if len(results) > 1 {
			assert.GreaterOrEqual(t, results[0].Score, results[1].Score)
		}
	})
}

func TestSearchBM25_Filters(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Insert test documents with different languages
	docs := []vectorstore.Document{
		{
			ID:      "go1",
			Content: "Go programming language",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"language": "go",
				"type":     "tutorial",
			},
		},
		{
			ID:      "py1",
			Content: "Python programming language",
			Vector:  embedding.Vector{0.2, 0.3, 0.4},
			Metadata: map[string]interface{}{
				"language": "python",
				"type":     "tutorial",
			},
		},
		{
			ID:      "go2",
			Content: "Advanced Go patterns",
			Vector:  embedding.Vector{0.3, 0.4, 0.5},
			Metadata: map[string]interface{}{
				"language": "go",
				"type":     "advanced",
			},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	t.Run("filter by language", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "programming", vectorstore.SearchOptions{
			Limit: 10,
			Filters: map[string]interface{}{
				"language": "go",
			},
		})
		require.NoError(t, err)
		assert.Len(t, results, 1) // Only go1 matches
		assert.Equal(t, "go1", results[0].Document.ID)
	})

	t.Run("multiple filters", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "language", vectorstore.SearchOptions{
			Limit: 10,
			Filters: map[string]interface{}{
				"language": "go",
				"type":     "tutorial",
			},
		})
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "go1", results[0].Document.ID)
	})

	t.Run("no matches with filter", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "programming", vectorstore.SearchOptions{
			Limit: 10,
			Filters: map[string]interface{}{
				"language": "rust",
			},
		})
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestSearchBM25_Phrases(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Insert documents with similar but different phrases
	docs := []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "The quick brown fox jumps over the lazy dog",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
		},
		{
			ID:      "doc2",
			Content: "A lazy dog sleeps under the tree",
			Vector:  embedding.Vector{0.2, 0.3, 0.4},
		},
		{
			ID:      "doc3",
			Content: "The brown dog is very lazy",
			Vector:  embedding.Vector{0.3, 0.4, 0.5},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	t.Run("exact phrase match", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, `"lazy dog"`, vectorstore.SearchOptions{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
		
		// doc1 and doc2 contain exact phrase "lazy dog"
		foundExact := false
		for _, r := range results {
			if r.Document.ID == "doc1" || r.Document.ID == "doc2" {
				foundExact = true
				break
			}
		}
		assert.True(t, foundExact, "should find documents with exact phrase")
	})

	t.Run("phrase with quotes in content", func(t *testing.T) {
		// Add document with quotes
		doc := vectorstore.Document{
			ID:      "doc4",
			Content: `He said "hello world" and smiled`,
			Vector:  embedding.Vector{0.4, 0.5, 0.6},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		results, err := store.SearchBM25(ctx, `"hello world"`, vectorstore.SearchOptions{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})
}

func TestSearchBM25_EdgeCases(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Insert test document
	doc := vectorstore.Document{
		ID:      "doc1",
		Content: "Test document for edge cases",
		Vector:  embedding.Vector{0.1, 0.2, 0.3},
	}
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	t.Run("empty query", func(t *testing.T) {
		_, err := store.SearchBM25(ctx, "", vectorstore.SearchOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("whitespace only query", func(t *testing.T) {
		_, err := store.SearchBM25(ctx, "   ", vectorstore.SearchOptions{})
		assert.Error(t, err)
	})

	t.Run("no results", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "nonexistent", vectorstore.SearchOptions{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("default limit", func(t *testing.T) {
		// Add many documents
		var docs []vectorstore.Document
		for i := 0; i < 20; i++ {
			docs = append(docs, vectorstore.Document{
				ID:      string(rune('a' + i)),
				Content: "test document",
				Vector:  embedding.Vector{0.1, 0.2, 0.3},
			})
		}
		err := store.UpsertBatch(ctx, docs)
		require.NoError(t, err)

		results, err := store.SearchBM25(ctx, "test", vectorstore.SearchOptions{
			Limit: 0, // Should default to 10
		})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 10)
	})

	t.Run("limit enforcement", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "test", vectorstore.SearchOptions{
			Limit: 5,
		})
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 5)
	})
}

func TestSearchBM25_Threshold(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Insert documents with varying relevance
	docs := []vectorstore.Document{
		{
			ID:      "doc1",
			Content: "Go programming Go programming Go programming", // High relevance
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
		},
		{
			ID:      "doc2",
			Content: "Go programming language", // Medium relevance
			Vector:  embedding.Vector{0.2, 0.3, 0.4},
		},
		{
			ID:      "doc3",
			Content: "This document mentions Go once", // Low relevance
			Vector:  embedding.Vector{0.3, 0.4, 0.5},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	t.Run("no threshold", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "Go", vectorstore.SearchOptions{
			Limit:     10,
			Threshold: 0,
		})
		require.NoError(t, err)
		assert.Len(t, results, 3) // All documents
	})

	t.Run("high threshold", func(t *testing.T) {
		results, err := store.SearchBM25(ctx, "Go programming", vectorstore.SearchOptions{
			Limit:     10,
			Threshold: 0.5, // Only high-relevance docs
		})
		require.NoError(t, err)
		
		// Should return fewer results with high threshold
		assert.LessOrEqual(t, len(results), 2)
		
		// All returned results should have score >= threshold
		for _, r := range results {
			assert.GreaterOrEqual(t, r.Score, float32(0.5))
		}
	})
}

func TestSearchBM25_Ranking(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Insert documents with different relevance levels
	docs := []vectorstore.Document{
		{
			ID:      "high",
			Content: "Golang Golang Golang programming in Go",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
		},
		{
			ID:      "medium",
			Content: "Programming in Golang with standard library",
			Vector:  embedding.Vector{0.2, 0.3, 0.4},
		},
		{
			ID:      "low",
			Content: "This document briefly mentions Golang",
			Vector:  embedding.Vector{0.3, 0.4, 0.5},
		},
	}

	err := store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	results, err := store.SearchBM25(ctx, "Golang programming", vectorstore.SearchOptions{
		Limit: 10,
	})
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// Verify results are ranked by relevance (descending score)
	for i := 1; i < len(results); i++ {
		assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
			"results should be ranked by score (highest first)")
	}

	// High relevance document should rank first
	assert.Equal(t, "high", results[0].Document.ID)
}

func TestSearchBM25_ContextCancellation(t *testing.T) {
	store := newTestStore(t)

	// Insert a document
	doc := vectorstore.Document{
		ID:      "doc1",
		Content: "Test document",
		Vector:  embedding.Vector{0.1, 0.2, 0.3},
	}
	err := store.Upsert(context.Background(), doc)
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Search should respect cancellation
	_, err = store.SearchBM25(ctx, "test", vectorstore.SearchOptions{
		Limit: 10,
	})
	assert.Error(t, err)
}

func TestParseFTS5Query(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple word",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "two words",
			input:    "hello world",
			expected: "hello AND world",
		},
		{
			name:     "quoted phrase",
			input:    `"hello world"`,
			expected: `"hello world"`,
		},
		{
			name:     "phrase and word",
			input:    `"hello world" test`,
			expected: `"hello world" AND test`,
		},
		{
			name:     "explicit AND",
			input:    "hello and world",
			expected: "hello AND world",
		},
		{
			name:     "explicit OR",
			input:    "hello or world",
			expected: "hello OR world",
		},
		{
			name:     "explicit NOT",
			input:    "hello not world",
			expected: "hello NOT world",
		},
		{
			name:     "mixed operators",
			input:    "go and programming or rust",
			expected: "go AND programming OR rust",
		},
		{
			name:     "special characters in phrase",
			input:    `"test@example.com"`,
			expected: `"test@example.com"`,
		},
		{
			name:     "multiple spaces",
			input:    "hello    world",
			expected: "hello AND world",
		},
		{
			name:     "leading/trailing spaces",
			input:    "  hello world  ",
			expected: "hello AND world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFTS5Query(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeRank(t *testing.T) {
	tests := []struct {
		name     string
		rank     float32
		expected float32
	}{
		{
			name:     "best match (0)",
			rank:     0,
			expected: 0.0,
		},
		{
			name:     "typical match (-5)",
			rank:     -5.0,
			expected: 0.5,
		},
		{
			name:     "weak match (-10)",
			rank:     -10.0,
			expected: 1.0,
		},
		{
			name:     "very weak match (-15)",
			rank:     -15.0,
			expected: 1.0, // Clamped to max
		},
		{
			name:     "positive rank (shouldn't happen)",
			rank:     5.0,
			expected: 0.0, // Clamped to min
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeRank(tt.rank)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}
