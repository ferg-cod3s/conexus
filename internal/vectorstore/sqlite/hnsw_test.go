package sqlite

import (
	"testing"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHNSWBasic(t *testing.T) {
	// Create HNSW index
	hnsw := NewHNSWIndex(DefaultHNSWConfig())

	// Create test vectors (simple 3D for testing)
	vec1 := embedding.Vector{1.0, 0.0, 0.0}
	vec2 := embedding.Vector{0.0, 1.0, 0.0}
	vec3 := embedding.Vector{0.9, 0.1, 0.0} // Close to vec1

	// Insert vectors
	err := hnsw.Insert("doc1", vec1)
	require.NoError(t, err)

	err = hnsw.Insert("doc2", vec2)
	require.NoError(t, err)

	err = hnsw.Insert("doc3", vec3)
	require.NoError(t, err)

	// Search for nearest neighbors to vec1
	query := embedding.Vector{1.0, 0.0, 0.0}
	results, err := hnsw.Search(query, 2, 32)
	require.NoError(t, err)

	// Should find doc1 and doc3 as closest
	assert.Len(t, results, 2)
	assert.Equal(t, "doc1", results[0].ID) // Exact match should be first
	assert.Equal(t, "doc3", results[1].ID) // Close vector should be second

	// Check distances (should be very small)
	assert.True(t, results[0].Distance < 0.01) // doc1 should be very close to query
	assert.True(t, results[1].Distance < 0.1)  // doc3 should be close but not as close
}

func TestHNSWSize(t *testing.T) {
	hnsw := NewHNSWIndex(DefaultHNSWConfig())

	// Initially empty
	assert.Equal(t, 0, hnsw.Size())

	// Add some vectors
	vec := embedding.Vector{1.0, 0.0, 0.0}
	err := hnsw.Insert("doc1", vec)
	require.NoError(t, err)

	assert.Equal(t, 1, hnsw.Size())

	// Add another
	err = hnsw.Insert("doc2", vec)
	require.NoError(t, err)

	assert.Equal(t, 2, hnsw.Size())

	// Remove one
	err = hnsw.Remove("doc1")
	require.NoError(t, err)

	assert.Equal(t, 1, hnsw.Size())
}

func TestHnswNormalizeVector(t *testing.T) {
	// Test vector normalization
	vec := embedding.Vector{3.0, 4.0, 0.0} // Magnitude = 5
	normalized := hnswNormalizeVector(vec)

	// Check magnitude is 1
	mag := vectorMagnitude(normalized)
	assert.True(t, mag > 0.999 && mag < 1.001) // Should be approximately 1

	// Check direction is preserved
	expected := embedding.Vector{0.6, 0.8, 0.0} // 3/5, 4/5, 0
	for i := range normalized {
		assert.True(t, normalized[i] > expected[i]-0.001 && normalized[i] < expected[i]+0.001)
	}
}
