# Vector Store Package

## Overview
Storage abstractions for vectors and metadata with hybrid search (BM25 + vector similarity).

## Key Interfaces

### `VectorStore`
Main storage interface:
- `Upsert()` / `UpsertBatch()` - Insert or update documents
- `SearchVector()` - Dense vector similarity search
- `SearchBM25()` - Sparse keyword search (BM25)
- `SearchHybrid()` - Combined search with fusion
- `Count()`, `Get()`, `Delete()` - CRUD operations

### `StatsProvider`
Provides index statistics (document count, size, languages).

## Implementation: SQLite

### BM25 (Sparse)
- SQLite FTS5 extension for full-text search
- Configurable BM25 parameters (k1, b)

### Vector (Dense)
- In-memory vector index for MVP
- SQLite blob storage for persistence
- Future: HNSW/IVF for large-scale

## Usage Example

```go
import "github.com/ferg-cod3s/conexus/internal/vectorstore"

store, err := sqlite.New("/path/to/index.db")

doc := vectorstore.Document{
    ID:      "file.go:10-20",
    Content: "func main() { ... }",
    Vector:  embedding.Vector{0.1, 0.2, ...},
    Metadata: map[string]interface{}{
        "file_path": "main.go",
        "language":  "go",
    },
}

err = store.Upsert(ctx, doc)

results, err := store.SearchHybrid(ctx, "main function", queryVector, opts)
```

## Schema

### `documents` table
- `id` TEXT PRIMARY KEY
- `content` TEXT
- `vector` BLOB (serialized float32 array)
- `metadata` JSON
- `created_at`, `updated_at` TIMESTAMP

### `documents_fts` FTS5 table
- Virtual table for BM25 search
- Indexes `content` column

## Implementation Status
- [ ] SQLite store implementation
- [ ] FTS5 BM25 search
- [ ] In-memory vector index
- [ ] Hybrid search with fusion
- [ ] Unit tests
