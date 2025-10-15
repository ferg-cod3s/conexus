# Search Package

## Overview
Hybrid search orchestration and reranking capabilities.

## Key Interfaces

### `Retriever`
High-level search interface.

### `FusionStrategy`
Combines sparse (BM25) and dense (vector) results:
- **RRF (Reciprocal Rank Fusion)**: `1 / (k + rank)`
- **Weighted**: `α * sparse_score + (1-α) * dense_score`

### `Reranker`
Re-scores and re-orders results based on query-document relevance.

### `Pipeline`
Orchestrates: search → fuse → rerank.

## Hybrid Modes

| Mode       | Description                        |
|------------|-----------------------------------|
| `rrf`      | Reciprocal Rank Fusion (default)  |
| `weighted` | Weighted score combination        |
| `sparse`   | BM25 only                         |
| `dense`    | Vector similarity only            |

## Usage Example

```go
import "github.com/ferg-cod3s/conexus/internal/search"

pipeline := search.NewPipeline(store, embedder, fusion, reranker)

query := search.Query{
    Text:       "authentication middleware",
    Limit:      10,
    HybridMode: search.HybridModeRRF,
}

results, err := pipeline.Search(ctx, query)
```

## Reranking

### Lexical Reranker (MVP)
- Exact match bonus
- Term overlap scoring
- Low cost, fast

### Future: Cross-Encoder
- BERT-based reranker
- Highest quality, slower
- Optional for production

## Implementation Status
- [ ] RRF fusion strategy
- [ ] Weighted fusion strategy
- [ ] Lexical reranker
- [ ] Search pipeline
- [ ] Unit tests
