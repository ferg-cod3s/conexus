# Embedding Package

## Overview
Provides pluggable text embedding generation with a provider abstraction for future integrations (OpenAI, Voyage, Cohere, local models).

## Key Interfaces

### `Embedder`
Generates embeddings for text:
- `Embed()` - Single text embedding
- `EmbedBatch()` - Batch embedding for efficiency
- `Dimensions()` - Vector dimensionality
- `Model()` - Model identifier

### `Provider`
Factory for creating embedders with specific configs.

### `ProviderRegistry`
Manages available embedding providers.

## Usage Example

```go
import "github.com/ferg-cod3s/conexus/internal/embedding"

// Use mock embedder (default)
embedder := embedding.NewMock(384) // 384 dimensions

emb, err := embedder.Embed(ctx, "func main() { fmt.Println(\"Hello\") }")
// emb.Vector is []float32 with 384 dimensions
```

## Providers

### Mock (default)
- Deterministic embeddings from hash(text)
- Zero cost, reproducible tests
- 384 or configurable dimensions

### Future Providers
- OpenAI (`text-embedding-3-small`, `text-embedding-3-large`)
- Voyage AI (`voyage-code-2`)
- Cohere (`embed-multilingual-v3`)
- Local models (sentence-transformers via HTTP)

## Implementation Status
- [ ] Embedder interface
- [ ] Provider registry
- [ ] Mock embedder (deterministic)
- [ ] Unit tests
