package sqlite

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

func TestSearchVector_Basic(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add test documents with known vectors
	docs := []vectorstore.Document{
		{ID: "doc1", Content: "document about cats", Vector: embedding.Vector{1.0, 0.0, 0.0}},
		{ID: "doc2", Content: "document about dogs", Vector: embedding.Vector{0.9, 0.1, 0.0}},
		{ID: "doc3", Content: "document about birds", Vector: embedding.Vector{0.0, 1.0, 0.0}},
		{ID: "doc4", Content: "document about fish", Vector: embedding.Vector{0.0, 0.0, 1.0}},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Query with vector similar to "cats"
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{
		Limit: 2,
	})

	require.NoError(t, err)
	require.Len(t, results, 2)

	// First result should be "cats" with perfect similarity
	assert.Equal(t, "document about cats", results[0].Document.Content)
	assert.InDelta(t, 1.0, results[0].Score, 0.001)

	// Second result should be "dogs" with high similarity
	assert.Equal(t, "document about dogs", results[1].Document.Content)
	assert.Greater(t, results[1].Score, float32(0.8))
	assert.Less(t, results[1].Score, float32(1.0))
}

func TestSearchVector_IdenticalVectors(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add document with specific vector
	testVector := embedding.Vector{0.6, 0.8, 0.0}
	doc := vectorstore.Document{
		ID:      "doc1",
		Content: "test document",
		Vector:  testVector,
	}
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Query with identical vector
	results, err := store.SearchVector(ctx, testVector, vectorstore.SearchOptions{
		Limit: 1,
	})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.InDelta(t, 1.0, results[0].Score, 0.001, "identical vectors should have similarity = 1.0")
}

func TestSearchVector_OrthogonalVectors(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add documents with orthogonal vectors
	docs := []vectorstore.Document{
		{ID: "x", Content: "x-axis", Vector: embedding.Vector{1.0, 0.0, 0.0}},
		{ID: "y", Content: "y-axis", Vector: embedding.Vector{0.0, 1.0, 0.0}},
		{ID: "z", Content: "z-axis", Vector: embedding.Vector{0.0, 0.0, 1.0}},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Query with x-axis vector
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{
		Limit: 3,
	})

	require.NoError(t, err)
	require.Len(t, results, 3)

	// First should be x-axis with similarity 1.0
	assert.Equal(t, "x-axis", results[0].Document.Content)
	assert.InDelta(t, 1.0, results[0].Score, 0.001)

	// Others should be orthogonal with similarity ~0.0
	assert.InDelta(t, 0.0, results[1].Score, 0.001, "orthogonal vectors should have similarity = 0.0")
	assert.InDelta(t, 0.0, results[2].Score, 0.001, "orthogonal vectors should have similarity = 0.0")
}

func TestSearchVector_SimilarityRanking(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Normalize helper
	normalize := func(v embedding.Vector) embedding.Vector {
		mag := vectorMagnitude(v)
		result := make(embedding.Vector, len(v))
		for i := range v {
			result[i] = v[i] / mag
		}
		return result
	}

	// Add documents with varying similarity to query
	baseVector := normalize(embedding.Vector{1.0, 0.0, 0.0})
	docs := []vectorstore.Document{
		{ID: "very", Content: "very similar", Vector: normalize(embedding.Vector{1.0, 0.1, 0.0})},     // ~0.995 similarity
		{ID: "some", Content: "somewhat similar", Vector: normalize(embedding.Vector{1.0, 1.0, 0.0})}, // ~0.707 similarity
		{ID: "less", Content: "less similar", Vector: normalize(embedding.Vector{0.5, 1.0, 0.0})},     // ~0.447 similarity
		{ID: "not", Content: "not similar", Vector: normalize(embedding.Vector{0.0, 1.0, 0.0})},       // 0.0 similarity
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Query
	results, err := store.SearchVector(ctx, baseVector, vectorstore.SearchOptions{
		Limit: 4,
	})

	require.NoError(t, err)
	require.Len(t, results, 4)

	// Verify descending order
	assert.Equal(t, "very similar", results[0].Document.Content)
	assert.Equal(t, "somewhat similar", results[1].Document.Content)
	assert.Equal(t, "less similar", results[2].Document.Content)
	assert.Equal(t, "not similar", results[3].Document.Content)

	// Verify scores are in descending order
	for i := 0; i < len(results)-1; i++ {
		assert.GreaterOrEqual(t, results[i].Score, results[i+1].Score,
			"results should be sorted by similarity (descending)")
	}
}

func TestSearchVector_LimitEnforcement(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add 10 documents
	for i := 0; i < 10; i++ {
		doc := vectorstore.Document{
			ID:      fmt.Sprintf("doc%d", i),
			Content: fmt.Sprintf("document %d", i),
			Vector:  embedding.Vector{float32(i) / 10.0, 0.5, 0.5},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Query with limit 3
	queryVector := embedding.Vector{0.5, 0.5, 0.5}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{
		Limit: 3,
	})

	require.NoError(t, err)
	assert.Len(t, results, 3, "should return exactly limit number of results")
}

func TestSearchVector_DefaultLimit(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add 15 documents
	for i := 0; i < 15; i++ {
		doc := vectorstore.Document{
			ID:      fmt.Sprintf("doc%d", i),
			Content: fmt.Sprintf("document %d", i),
			Vector:  embedding.Vector{float32(i) / 15.0, 0.5, 0.5},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Query without specifying limit (should default to 10)
	queryVector := embedding.Vector{0.5, 0.5, 0.5}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{})

	require.NoError(t, err)
	assert.Len(t, results, 10, "should default to 10 results when limit not specified")
}

func TestSearchVector_ThresholdFiltering(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add documents with varying similarity
	docs := []vectorstore.Document{
		{ID: "high", Content: "high similarity", Vector: embedding.Vector{1.0, 0.0, 0.0}},
		{ID: "med", Content: "medium similarity", Vector: embedding.Vector{0.7, 0.7, 0.0}},
		{ID: "low", Content: "low similarity", Vector: embedding.Vector{0.0, 1.0, 0.0}},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Query with high threshold
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{
		Threshold: 0.9,
		Limit:     10,
	})

	require.NoError(t, err)
	assert.Len(t, results, 1, "only documents above threshold should be returned")
	assert.Equal(t, "high similarity", results[0].Document.Content)
	assert.GreaterOrEqual(t, results[0].Score, float32(0.9))
}

func TestSearchVector_MetadataFiltering(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add documents with metadata
	docs := []vectorstore.Document{
		{ID: "cat", Content: "cat doc", Vector: embedding.Vector{1.0, 0.0, 0.0}, Metadata: map[string]interface{}{"category": "animals"}},
		{ID: "dog", Content: "dog doc", Vector: embedding.Vector{0.9, 0.1, 0.0}, Metadata: map[string]interface{}{"category": "animals"}},
		{ID: "car", Content: "car doc", Vector: embedding.Vector{0.8, 0.2, 0.0}, Metadata: map[string]interface{}{"category": "vehicles"}},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Query with metadata filter
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{
		Filters: map[string]interface{}{"category": "animals"},
		Limit:   10,
	})

	require.NoError(t, err)
	assert.Len(t, results, 2, "should only return documents matching metadata filter")
	
	// Verify all results have correct category
	// Metadata is already map[string]interface{}, no need to unmarshal
	for _, result := range results {
		require.NotNil(t, result.Document.Metadata)
		assert.Equal(t, "animals", result.Document.Metadata["category"])
	}
}

func TestSearchVector_ContextCancellation(t *testing.T) {
	store := newTestStore(t)

	// Add some documents
	for i := 0; i < 5; i++ {
		doc := vectorstore.Document{
			ID:      fmt.Sprintf("doc%d", i),
			Content: fmt.Sprintf("document %d", i),
			Vector:  embedding.Vector{float32(i), 0.5, 0.5},
		}
		err := store.Upsert(context.Background(), doc)
		require.NoError(t, err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Query with cancelled context
	queryVector := embedding.Vector{0.5, 0.5, 0.5}
	_, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

func TestSearchVector_EmptyQueryVector(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Try to query with empty vector
	emptyVector := embedding.Vector{}
	_, err := store.SearchVector(ctx, emptyVector, vectorstore.SearchOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "query vector cannot be empty")
}

func TestSearchVector_ZeroMagnitudeVector(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Try to query with zero-magnitude vector
	zeroVector := embedding.Vector{0.0, 0.0, 0.0}
	_, err := store.SearchVector(ctx, zeroVector, vectorstore.SearchOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "zero magnitude")
}

func TestSearchVector_DimensionMismatch(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add document with 3D vector
	doc := vectorstore.Document{
		ID:      "doc1",
		Content: "test doc",
		Vector:  embedding.Vector{1.0, 0.0, 0.0},
	}
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Query with 4D vector (dimension mismatch)
	queryVector := embedding.Vector{1.0, 0.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{})

	require.NoError(t, err)
	// Should return 0 results due to dimension mismatch
	assert.Len(t, results, 0, "dimension mismatch should exclude document from results")
}

func TestSearchVector_NoDocuments(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Query empty store
	queryVector := embedding.Vector{1.0, 0.0, 0.0}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{})

	require.NoError(t, err)
	assert.Empty(t, results, "should return empty results when no documents exist")
}

func TestSearchVector_SingleDocument(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add single document
	testVector := embedding.Vector{0.6, 0.8, 0.0}
	doc := vectorstore.Document{
		ID:      "only",
		Content: "only document",
		Vector:  testVector,
	}
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Query
	queryVector := embedding.Vector{0.6, 0.8, 0.0}
	results, err := store.SearchVector(ctx, queryVector, vectorstore.SearchOptions{})

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "only document", results[0].Document.Content)
	assert.InDelta(t, 1.0, results[0].Score, 0.001)
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		vec1     embedding.Vector
		vec2     embedding.Vector
		expected float64
		delta    float64
	}{
		{
			name:     "identical vectors",
			vec1:     embedding.Vector{1.0, 0.0, 0.0},
			vec2:     embedding.Vector{1.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "orthogonal vectors",
			vec1:     embedding.Vector{1.0, 0.0, 0.0},
			vec2:     embedding.Vector{0.0, 1.0, 0.0},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "45 degree angle",
			vec1:     embedding.Vector{1.0, 0.0, 0.0},
			vec2:     embedding.Vector{1.0, 1.0, 0.0},
			expected: 0.707, // cos(45°) ≈ 0.707
			delta:    0.01,
		},
		{
			name:     "opposite direction",
			vec1:     embedding.Vector{1.0, 0.0, 0.0},
			vec2:     embedding.Vector{-1.0, 0.0, 0.0},
			expected: 0.0, // clamped from -1.0
			delta:    0.001,
		},
		{
			name:     "different magnitudes same direction",
			vec1:     embedding.Vector{2.0, 0.0, 0.0},
			vec2:     embedding.Vector{5.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := cosineSimilarity(tt.vec1, tt.vec2)
			assert.InDelta(t, tt.expected, similarity, tt.delta)
		})
	}
}

func TestCosineSimilarity_DimensionMismatch(t *testing.T) {
	vec1 := embedding.Vector{1.0, 0.0, 0.0}
	vec2 := embedding.Vector{1.0, 0.0} // different dimension

	similarity := cosineSimilarity(vec1, vec2)
	assert.Equal(t, float32(0.0), similarity, "dimension mismatch should return 0.0")
}

func TestCosineSimilarity_ZeroMagnitude(t *testing.T) {
	vec1 := embedding.Vector{1.0, 0.0, 0.0}
	vec2 := embedding.Vector{0.0, 0.0, 0.0}

	similarity := cosineSimilarity(vec1, vec2)
	assert.Equal(t, float32(0.0), similarity, "zero magnitude vector should return 0.0")
}

func TestVectorMagnitude(t *testing.T) {
	tests := []struct {
		name     string
		vector   embedding.Vector
		expected float32
		delta    float64
	}{
		{
			name:     "unit vector x-axis",
			vector:   embedding.Vector{1.0, 0.0, 0.0},
			expected: 1.0,
			delta:    0.001,
		},
		{
			name:     "unit vector diagonal",
			vector:   embedding.Vector{0.707, 0.707, 0.0},
			expected: 1.0,
			delta:    0.01,
		},
		{
			name:     "3-4-5 triangle",
			vector:   embedding.Vector{3.0, 4.0, 0.0},
			expected: 5.0,
			delta:    0.001,
		},
		{
			name:     "zero vector",
			vector:   embedding.Vector{0.0, 0.0, 0.0},
			expected: 0.0,
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			magnitude := vectorMagnitude(tt.vector)
			assert.InDelta(t, tt.expected, magnitude, tt.delta)
		})
	}
}

func TestSearchVector_ResultStructure(t *testing.T) {
	store := newTestStore(t)

	ctx := context.Background()

	// Add document with metadata
	metadata := map[string]interface{}{
		"author": "test",
		"tags":   []string{"tag1", "tag2"},
	}
	testVector := embedding.Vector{1.0, 0.0, 0.0}
	doc := vectorstore.Document{
		ID:       "doc1",
		Content:  "test content",
		Vector:   testVector,
		Metadata: metadata,
	}
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Query
	results, err := store.SearchVector(ctx, testVector, vectorstore.SearchOptions{})

	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	
	// Verify result structure
	assert.NotEmpty(t, result.Document.ID, "ID should be populated")
	assert.Equal(t, "test content", result.Document.Content)
	assert.NotNil(t, result.Document.Vector)
	assert.Equal(t, testVector, result.Document.Vector)
	assert.NotNil(t, result.Document.Metadata)
	assert.InDelta(t, 1.0, result.Score, 0.001)
	assert.False(t, result.Document.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, result.Document.UpdatedAt.IsZero(), "UpdatedAt should be set")
}
