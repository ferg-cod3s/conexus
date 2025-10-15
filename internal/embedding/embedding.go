// Package embedding provides pluggable text embedding generation with provider abstractions.
package embedding

import (
	"context"
)

// Vector represents a dense embedding vector.
type Vector []float32

// Embedding is a text embedding with metadata.
type Embedding struct {
	Text   string  // Original text that was embedded
	Vector Vector  // Dense vector representation
	Model  string  // Model used for embedding (e.g., "mock", "openai/text-embedding-3-small")
}

// Embedder generates embeddings for text inputs.
type Embedder interface {
	// Embed generates an embedding for a single text input.
	Embed(ctx context.Context, text string) (*Embedding, error)
	
	// EmbedBatch generates embeddings for multiple texts efficiently.
	EmbedBatch(ctx context.Context, texts []string) ([]*Embedding, error)
	
	// Dimensions returns the dimensionality of vectors produced by this embedder.
	Dimensions() int
	
	// Model returns the identifier of the embedding model.
	Model() string
}

// Provider is a factory for creating embedders with specific configurations.
type Provider interface {
	// Name returns the provider identifier (e.g., "openai", "voyage", "mock").
	Name() string
	
	// Create instantiates an embedder with the given configuration.
	Create(config map[string]interface{}) (Embedder, error)
}

// ProviderRegistry manages available embedding providers.
type ProviderRegistry interface {
	// Register adds a provider to the registry.
	Register(provider Provider) error
	
	// Get retrieves a provider by name.
	Get(name string) (Provider, error)
	
	// List returns all registered provider names.
	List() []string
}
