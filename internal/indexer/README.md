# Indexer Package

## Overview
The indexer package provides file system traversal, content chunking, and metadata extraction for building a searchable codebase index.

## Key Interfaces

### `Indexer`
Main interface for indexing operations:
- `Index()` - Full index of a directory tree
- `IndexIncremental()` - Only index changed files (using Merkle trees)

### `Chunker`
Splits file content into semantic chunks:
- Code chunkers: AST-based (functions, classes, structs)
- Doc chunkers: Sliding window with paragraph boundaries

### `Walker`
File system traversal with `.gitignore` support.

### `MerkleTree`
Content hashing for detecting changes in incremental indexing.

## Usage Example

```go
import "github.com/ferg-cod3s/conexus/internal/indexer"

opts := indexer.IndexOptions{
    RootPath:       "/path/to/repo",
    IgnorePatterns: []string{"*.test.go", "vendor/"},
    MaxFileSize:    1024 * 1024, // 1MB
    ChunkSize:      512,          // tokens
    ChunkOverlap:   50,           // tokens
}

idx := indexer.New() // Implementation TBD
chunks, err := idx.Index(ctx, opts)
```

## Chunk Types
- `function` - Function definitions
- `class` - Class definitions
- `struct` - Go structs
- `interface` - Go interfaces
- `comment` - Doc comments
- `paragraph` - Documentation paragraphs
- `code_block` - Code snippets in docs

## Implementation Status
- [ ] File walker with `.gitignore` support
- [ ] Merkle tree for incremental indexing
- [ ] Code chunker (AST-based)
- [ ] Doc chunker (sliding window)
- [ ] Unit tests
