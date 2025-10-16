package embedding

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
)

// MockEmbedder generates deterministic embeddings from text hashes.
// Useful for testing and development without external API dependencies.
type MockEmbedder struct {
	dimensions int
	model      string
}

// NewMock creates a new mock embedder with the specified dimensions.
func NewMock(dimensions int) *MockEmbedder {
	return &MockEmbedder{
		dimensions: dimensions,
		model:      fmt.Sprintf("mock-%d", dimensions),
	}
}

// Embed generates a deterministic embedding from text hash.
func (m *MockEmbedder) Embed(ctx context.Context, text string) (*Embedding, error) {
	if text == "" {
		return nil, fmt.Errorf("cannot embed empty text")
	}
	
	vector := m.generateVector(text)
	
	return &Embedding{
		Text:   text,
		Vector: vector,
		Model:  m.model,
	}, nil
}

// EmbedBatch generates embeddings for multiple texts.
func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([]*Embedding, error) {
	embeddings := make([]*Embedding, len(texts))
	
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text at index %d: %w", i, err)
		}
		embeddings[i] = emb
	}
	
	return embeddings, nil
}

// Dimensions returns the vector dimensionality.
func (m *MockEmbedder) Dimensions() int {
	return m.dimensions
}

// Model returns the model identifier.
func (m *MockEmbedder) Model() string {
	return m.model
}

// generateVector creates a deterministic normalized vector from text.
// Uses SHA256 hash as seed for reproducible pseudo-random values.
func (m *MockEmbedder) generateVector(text string) Vector {
	// Hash the text to get a deterministic seed
	hash := sha256.Sum256([]byte(text))
	
	vector := make(Vector, m.dimensions)
	
	// Generate pseudo-random values from hash
	for i := 0; i < m.dimensions; i++ {
		// Use different parts of the hash cyclically
		offset := (i * 4) % len(hash)
		seed := binary.BigEndian.Uint32(hash[offset:])
		
		// Convert to float in range [-1, 1]
		// Guard against integer overflow (G115): use int64 intermediate to avoid overflow
		// when converting uint32 to signed type
		seed64 := int64(seed)
		if seed64 > math.MaxInt32 {
			seed64 = seed64 % math.MaxInt32
		}
		vector[i] = float32(seed64) / float32(math.MaxInt32) // #nosec G115 -- seed64 is guaranteed <= MaxInt32
	}
	
	// Normalize to unit vector
	return normalize(vector)
}

// normalize scales a vector to unit length.
func normalize(v Vector) Vector {
	var sumSquares float32
	for _, val := range v {
		sumSquares += val * val
	}
	
	if sumSquares == 0 {
		return v
	}
	
	magnitude := float32(math.Sqrt(float64(sumSquares)))
	
	normalized := make(Vector, len(v))
	for i, val := range v {
		normalized[i] = val / magnitude
	}
	
	return normalized
}

// MockProvider implements Provider for the mock embedder.
type MockProvider struct{}

// Name returns the provider identifier.
func (p *MockProvider) Name() string {
	return "mock"
}

// Create instantiates a mock embedder with the given config.
func (p *MockProvider) Create(config map[string]interface{}) (Embedder, error) {
	dimensions := 384 // Default
	
	if dim, ok := config["dimensions"].(int); ok {
		dimensions = dim
	} else if dim, ok := config["dimensions"].(float64); ok {
		dimensions = int(dim)
	}
	
	if dimensions <= 0 {
		return nil, fmt.Errorf("dimensions must be positive, got %d", dimensions)
	}
	
	return NewMock(dimensions), nil
}
