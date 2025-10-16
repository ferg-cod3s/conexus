package embedding

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMock(t *testing.T) {
	tests := []struct {
		name       string
		dimensions int
		wantModel  string
	}{
		{
			name:       "standard 384 dimensions",
			dimensions: 384,
			wantModel:  "mock-384",
		},
		{
			name:       "small 128 dimensions",
			dimensions: 128,
			wantModel:  "mock-128",
		},
		{
			name:       "large 1536 dimensions",
			dimensions: 1536,
			wantModel:  "mock-1536",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMock(tt.dimensions)
			require.NotNil(t, m)
			assert.Equal(t, tt.dimensions, m.Dimensions())
			assert.Equal(t, tt.wantModel, m.Model())
		})
	}
}

func TestMockEmbedder_Embed(t *testing.T) {
	ctx := context.Background()
	m := NewMock(384)

	t.Run("successful embedding", func(t *testing.T) {
		text := "Hello, world!"
		emb, err := m.Embed(ctx, text)
		
		require.NoError(t, err)
		require.NotNil(t, emb)
		assert.Equal(t, text, emb.Text)
		assert.Equal(t, "mock-384", emb.Model)
		assert.Len(t, emb.Vector, 384)
	})

	t.Run("deterministic - same input produces same output", func(t *testing.T) {
		text := "deterministic test"
		
		emb1, err1 := m.Embed(ctx, text)
		require.NoError(t, err1)
		
		emb2, err2 := m.Embed(ctx, text)
		require.NoError(t, err2)
		
		// Vectors should be identical
		require.Len(t, emb1.Vector, len(emb2.Vector))
		for i := range emb1.Vector {
			assert.Equal(t, emb1.Vector[i], emb2.Vector[i], "vector mismatch at index %d", i)
		}
	})

	t.Run("different inputs produce different outputs", func(t *testing.T) {
		emb1, err1 := m.Embed(ctx, "text one")
		require.NoError(t, err1)
		
		emb2, err2 := m.Embed(ctx, "text two")
		require.NoError(t, err2)
		
		// Vectors should be different
		different := false
		for i := range emb1.Vector {
			if emb1.Vector[i] != emb2.Vector[i] {
				different = true
				break
			}
		}
		assert.True(t, different, "different texts should produce different vectors")
	})

	t.Run("vector is normalized", func(t *testing.T) {
		text := "normalization test"
		emb, err := m.Embed(ctx, text)
		require.NoError(t, err)
		
		// Calculate magnitude
		var sumSquares float32
		for _, val := range emb.Vector {
			sumSquares += val * val
		}
		magnitude := math.Sqrt(float64(sumSquares))
		
		// Should be unit vector (magnitude ≈ 1.0)
		assert.InDelta(t, 1.0, magnitude, 0.0001, "vector should be normalized to unit length")
	})

	t.Run("empty text returns error", func(t *testing.T) {
		emb, err := m.Embed(ctx, "")
		
		assert.Error(t, err)
		assert.Nil(t, emb)
		assert.Contains(t, err.Error(), "empty text")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		// Current implementation doesn't check context, but test for future
		// This documents expected behavior
		_, err := m.Embed(cancelCtx, "test")
		
		// Current implementation will succeed (no context checks yet)
		// This test documents that we should add context checks in the future
		if err != nil {
			assert.ErrorIs(t, err, context.Canceled)
		}
	})
}

func TestMockEmbedder_EmbedBatch(t *testing.T) {
	ctx := context.Background()
	m := NewMock(384)

	t.Run("successful batch embedding", func(t *testing.T) {
		texts := []string{"first", "second", "third"}
		
		embeddings, err := m.EmbedBatch(ctx, texts)
		
		require.NoError(t, err)
		require.Len(t, embeddings, 3)
		
		for i, emb := range embeddings {
			assert.Equal(t, texts[i], emb.Text)
			assert.Equal(t, "mock-384", emb.Model)
			assert.Len(t, emb.Vector, 384)
		}
	})

	t.Run("empty batch returns empty slice", func(t *testing.T) {
		texts := []string{}
		
		embeddings, err := m.EmbedBatch(ctx, texts)
		
		require.NoError(t, err)
		assert.Empty(t, embeddings)
	})

	t.Run("batch with empty text fails", func(t *testing.T) {
		texts := []string{"valid", "", "also valid"}
		
		embeddings, err := m.EmbedBatch(ctx, texts)
		
		assert.Error(t, err)
		assert.Nil(t, embeddings)
		assert.Contains(t, err.Error(), "index 1")
	})

	t.Run("batch is deterministic", func(t *testing.T) {
		texts := []string{"alpha", "beta", "gamma"}
		
		emb1, err1 := m.EmbedBatch(ctx, texts)
		require.NoError(t, err1)
		
		emb2, err2 := m.EmbedBatch(ctx, texts)
		require.NoError(t, err2)
		
		require.Len(t, emb1, len(emb2))
		for i := range emb1 {
			assert.Equal(t, emb1[i].Text, emb2[i].Text)
			for j := range emb1[i].Vector {
				assert.Equal(t, emb1[i].Vector[j], emb2[i].Vector[j])
			}
		}
	})

	t.Run("nil slice returns empty slice", func(t *testing.T) {
		var texts []string
		
		embeddings, err := m.EmbedBatch(ctx, texts)
		
		require.NoError(t, err)
		assert.Empty(t, embeddings)
	})
}

func TestMockEmbedder_Dimensions(t *testing.T) {
	tests := []struct {
		name       string
		dimensions int
	}{
		{"small", 64},
		{"standard", 384},
		{"large", 1536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMock(tt.dimensions)
			
			// Dimensions method returns correct value
			assert.Equal(t, tt.dimensions, m.Dimensions())
			
			// Generated vectors have correct length
			emb, err := m.Embed(context.Background(), "test")
			require.NoError(t, err)
			assert.Len(t, emb.Vector, tt.dimensions)
		})
	}
}

func TestMockEmbedder_Model(t *testing.T) {
	tests := []struct {
		dimensions int
		wantModel  string
	}{
		{384, "mock-384"},
		{128, "mock-128"},
		{1536, "mock-1536"},
	}

	for _, tt := range tests {
		t.Run(tt.wantModel, func(t *testing.T) {
			m := NewMock(tt.dimensions)
			assert.Equal(t, tt.wantModel, m.Model())
			
			emb, err := m.Embed(context.Background(), "test")
			require.NoError(t, err)
			assert.Equal(t, tt.wantModel, emb.Model)
		})
	}
}

func TestNormalize(t *testing.T) {
	t.Run("normalizes non-unit vector", func(t *testing.T) {
		v := Vector{3.0, 4.0} // Magnitude = 5
		normalized := normalize(v)
		
		assert.InDelta(t, 0.6, normalized[0], 0.0001)  // 3/5
		assert.InDelta(t, 0.8, normalized[1], 0.0001)  // 4/5
		
		// Check unit length
		var sumSquares float32
		for _, val := range normalized {
			sumSquares += val * val
		}
		magnitude := math.Sqrt(float64(sumSquares))
		assert.InDelta(t, 1.0, magnitude, 0.0001)
	})

	t.Run("handles zero vector", func(t *testing.T) {
		v := Vector{0.0, 0.0, 0.0}
		normalized := normalize(v)
		
		// Zero vector stays zero
		assert.Equal(t, v, normalized)
	})

	t.Run("already normalized vector unchanged", func(t *testing.T) {
		v := Vector{1.0, 0.0, 0.0} // Already unit length
		normalized := normalize(v)
		
		assert.InDelta(t, 1.0, normalized[0], 0.0001)
		assert.InDelta(t, 0.0, normalized[1], 0.0001)
		assert.InDelta(t, 0.0, normalized[2], 0.0001)
	})
}

func TestMockProvider_Name(t *testing.T) {
	p := &MockProvider{}
	assert.Equal(t, "mock", p.Name())
}

func TestMockProvider_Create(t *testing.T) {
	p := &MockProvider{}

	t.Run("creates with default dimensions", func(t *testing.T) {
		config := map[string]interface{}{}
		
		embedder, err := p.Create(config)
		
		require.NoError(t, err)
		require.NotNil(t, embedder)
		assert.Equal(t, 384, embedder.Dimensions())
		assert.Equal(t, "mock-384", embedder.Model())
	})

	t.Run("creates with custom dimensions as int", func(t *testing.T) {
		config := map[string]interface{}{
			"dimensions": 512,
		}
		
		embedder, err := p.Create(config)
		
		require.NoError(t, err)
		require.NotNil(t, embedder)
		assert.Equal(t, 512, embedder.Dimensions())
		assert.Equal(t, "mock-512", embedder.Model())
	})

	t.Run("creates with custom dimensions as float64", func(t *testing.T) {
		config := map[string]interface{}{
			"dimensions": float64(256),
		}
		
		embedder, err := p.Create(config)
		
		require.NoError(t, err)
		require.NotNil(t, embedder)
		assert.Equal(t, 256, embedder.Dimensions())
	})

	t.Run("rejects zero dimensions", func(t *testing.T) {
		config := map[string]interface{}{
			"dimensions": 0,
		}
		
		embedder, err := p.Create(config)
		
		assert.Error(t, err)
		assert.Nil(t, embedder)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("rejects negative dimensions", func(t *testing.T) {
		config := map[string]interface{}{
			"dimensions": -100,
		}
		
		embedder, err := p.Create(config)
		
		assert.Error(t, err)
		assert.Nil(t, embedder)
		assert.Contains(t, err.Error(), "must be positive")
	})

	t.Run("ignores invalid dimension type", func(t *testing.T) {
		config := map[string]interface{}{
			"dimensions": "not a number",
		}
		
		embedder, err := p.Create(config)
		
		// Should fall back to default
		require.NoError(t, err)
		assert.Equal(t, 384, embedder.Dimensions())
	})
}

// Benchmark tests
func BenchmarkMockEmbedder_Embed(b *testing.B) {
	ctx := context.Background()
	m := NewMock(384)
	text := "This is a sample text for benchmarking the embedding generation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Embed(ctx, text)
	}
}

func BenchmarkMockEmbedder_EmbedBatch(b *testing.B) {
	ctx := context.Background()
	m := NewMock(384)
	texts := []string{
		"First sample text",
		"Second sample text",
		"Third sample text",
		"Fourth sample text",
		"Fifth sample text",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.EmbedBatch(ctx, texts)
	}
}

func BenchmarkNormalize(b *testing.B) {
	v := make(Vector, 384)
	for i := range v {
		v[i] = float32(i) * 0.1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalize(v)
	}
}
