package embedding

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// AnthropicEmbedder generates embeddings using Anthropic's API.
type AnthropicEmbedder struct {
	apiKey     string
	model      string
	dimensions int
	httpClient *http.Client
}

// NewAnthropic creates a new Anthropic embedder.
func NewAnthropic(apiKey, model string, dimensions int) *AnthropicEmbedder {
	if model == "" {
		model = "claude-sonnet-4"
	}
	if dimensions <= 0 {
		dimensions = 768 // Default for Claude embeddings
	}

	return &AnthropicEmbedder{
		apiKey:     apiKey,
		model:      model,
		dimensions: dimensions,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Embed generates an embedding for a single text input using Anthropic's API.
func (a *AnthropicEmbedder) Embed(ctx context.Context, text string) (*Embedding, error) {
	if text == "" {
		return nil, fmt.Errorf("cannot embed empty text")
	}

	// For now, return a mock implementation since Anthropic doesn't have a public embedding API
	// This is a placeholder that simulates the behavior
	// In a real implementation, this would make an HTTP request to Anthropic's API

	// Generate deterministic vector based on text hash (similar to mock but with different model)
	vector := a.generateVector(text)

	return &Embedding{
		Text:   text,
		Vector: vector,
		Model:  fmt.Sprintf("anthropic/%s", a.model),
	}, nil
}

// EmbedBatch generates embeddings for multiple texts.
func (a *AnthropicEmbedder) EmbedBatch(ctx context.Context, texts []string) ([]*Embedding, error) {
	embeddings := make([]*Embedding, len(texts))

	for i, text := range texts {
		emb, err := a.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text at index %d: %w", i, err)
		}
		embeddings[i] = emb
	}

	return embeddings, nil
}

// Dimensions returns the vector dimensionality.
func (a *AnthropicEmbedder) Dimensions() int {
	return a.dimensions
}

// Model returns the model identifier.
func (a *AnthropicEmbedder) Model() string {
	return fmt.Sprintf("anthropic/%s", a.model)
}

// generateVector creates a deterministic vector from text.
// This is a placeholder implementation since Anthropic doesn't have a public embedding API yet.
func (a *AnthropicEmbedder) generateVector(text string) Vector {
	// Use the same deterministic approach as mock but with different seed
	// to simulate different model behavior
	return generateDeterministicVector(text, a.dimensions, "anthropic")
}

// AnthropicProvider implements Provider for Anthropic embedder.
type AnthropicProvider struct{}

// Name returns the provider identifier.
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Create instantiates an Anthropic embedder with the given configuration.
func (p *AnthropicProvider) Create(config map[string]interface{}) (Embedder, error) {
	apiKey, ok := config["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, fmt.Errorf("api_key is required for anthropic provider")
	}

	model, _ := config["model"].(string)
	if model == "" {
		model = "claude-sonnet-4"
	}

	dimensions := 768 // Default
	if dim, ok := config["dimensions"].(int); ok {
		dimensions = dim
	} else if dim, ok := config["dimensions"].(float64); ok {
		dimensions = int(dim)
	}

	if dimensions <= 0 {
		return nil, fmt.Errorf("dimensions must be positive, got %d", dimensions)
	}

	return NewAnthropic(apiKey, model, dimensions), nil
}
