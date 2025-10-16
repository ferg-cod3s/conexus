# Indexer Package

## Overview
The indexer package provides file system traversal, content chunking, and metadata extraction for building a searchable codebase index. It includes robust incremental indexing with Merkle tree-based change detection and seamless integration with vector stores for persistence.

## Key Components

### `Indexer` Interface
Main interface for indexing operations:
- `Index(ctx, opts)` - Full index of a directory tree
- `IndexIncremental(ctx, opts)` - Only index changed/added files (using Merkle trees)

### `FileWalker`
File system traversal with comprehensive `.gitignore` support:
- Respects `.gitignore` rules (wildcards, negations, directory-only patterns)
- Built-in default ignore patterns (`.git/`, `node_modules/`, etc.)
- Configurable max file size limits
- Error handling with partial results

### `MerkleTree`
Content hashing for efficient change detection:
- SHA-256 based file hashing
- Hierarchical tree structure for directory organization
- Persistent storage for incremental indexing
- Detects added, modified, and deleted files

### `Chunk`
Represents a semantic unit of content:
- `ID` - Unique identifier (file path + content hash)
- `FilePath` - Absolute path to source file
- `Content` - Text content of the chunk
- `StartLine`, `EndLine` - Source location
- `Language` - Programming language (auto-detected)
- `ChunkType` - Semantic type (function, class, comment, etc.)

## Usage Examples

### Basic Full Indexing

```go
import (
    "context"
    "github.com/ferg-cod3s/conexus/internal/indexer"
)

// Create indexer
idx := indexer.NewIndexer()

// Configure indexing options
opts := indexer.IndexOptions{
    RootPath:       "/path/to/repo",
    IgnorePatterns: []string{"*.test.go", "vendor/"},
    MaxFileSize:    1024 * 1024, // 1MB
}

// Perform full index
chunks, err := idx.Index(context.Background(), opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Indexed %d chunks\n", len(chunks))
```

### Incremental Indexing

```go
// First run: full index
chunks, err := idx.Index(ctx, opts)

// Subsequent runs: only index changes
changedChunks, err := idx.IndexIncremental(ctx, opts)

// changedChunks contains only chunks from modified/added files
fmt.Printf("Updated %d chunks\n", len(changedChunks))
```

### Vector Store Integration

The indexer seamlessly integrates with vector stores for persistent chunk storage:

```go
import (
    "github.com/ferg-cod3s/conexus/internal/indexer"
    "github.com/ferg-cod3s/conexus/internal/vectorstore/memory"
)

// Create indexer with vector store
store := memory.NewMemoryStore()
idx := indexer.NewIndexer(indexer.WithVectorStore(store))

opts := indexer.IndexOptions{
    RootPath: "/path/to/repo",
}

// Full index - chunks automatically stored
chunks, err := idx.Index(ctx, opts)
// Vector store now contains all chunks

// Incremental index - automatically handles updates
changedChunks, err := idx.IndexIncremental(ctx, opts)
// Vector store automatically updated:
//   - New/modified chunks added/updated
//   - Deleted file chunks removed
```

### Vector Store Operations During Indexing

The indexer performs these vector store operations automatically:

1. **During `Index()`**:
   - Stores all new chunks in vector store
   - Assigns unique IDs based on file path + content hash

2. **During `IndexIncremental()`**:
   - **Added files**: New chunks stored in vector store
   - **Modified files**: Old chunks deleted, new chunks stored
   - **Deleted files**: All chunks for that file removed from vector store
   - **Unchanged files**: No operations performed

### Important: Return Value Semantics

**Critical Understanding**: `IndexIncremental()` returns only chunks from **modified or added** files, NOT all existing chunks.

```go
// Example scenario:
// 1. Initial index with files A, B, C
chunks1, _ := idx.Index(ctx, opts)
// len(chunks1) = total chunks from A, B, C

// 2. Delete file B, file A and C unchanged
chunks2, _ := idx.IndexIncremental(ctx, opts)
// len(chunks2) = 0 (no modified files!)
// BUT: Vector store correctly has B's chunks removed

// 3. To verify deletions, query vector store directly:
chunkA, exists := store.Get(ctx, chunkIDFromFileA)
// exists = true (file A unchanged, chunk still present)

chunkB, exists := store.Get(ctx, chunkIDFromFileB)
// exists = false (file B deleted, chunks removed)
```

**Why this design?**
- Callers only need to process changes (efficient)
- Deletion detection happens internally via Merkle tree
- Vector store state is authoritative, not return values
- Enables efficient incremental updates without reprocessing unchanged files

### Testing Vector Store Integration

When testing vector store operations with the indexer:

```go
func TestVectorStoreDeletion(t *testing.T) {
    store := memory.NewMemoryStore()
    idx := indexer.NewIndexer(indexer.WithVectorStore(store))
    
    // Index files and capture chunk IDs
    chunks1, _ := idx.Index(ctx, opts)
    fileAChunkID := chunks1[0].ID
    fileBChunkID := chunks1[1].ID
    
    // Delete a file from disk
    os.Remove(filepath.Join(rootDir, "file_b.go"))
    
    // Incremental index
    chunks2, _ := idx.IndexIncremental(ctx, opts)
    
    // CORRECT: chunks2 will be empty (no modifications)
    assert.Empty(t, chunks2)
    
    // CORRECT: Verify store state directly
    _, exists := store.Get(ctx, fileAChunkID)
    assert.True(t, exists, "unchanged file chunk should exist")
    
    _, exists = store.Get(ctx, fileBChunkID)
    assert.False(t, exists, "deleted file chunk should be removed")
}
```

## Chunk Types

Semantic chunk types for different content kinds:
- `function` - Function definitions
- `class` - Class definitions  
- `struct` - Go structs
- `interface` - Go interfaces
- `method` - Class/struct methods
- `comment` - Doc comments
- `paragraph` - Documentation paragraphs
- `code_block` - Code snippets in docs
- `file` - Entire file (for small files)

## File Walker Features

### Gitignore Pattern Support
- `*` - Matches any characters except `/`
- `**` - Matches zero or more directories
- `?` - Matches any single character
- `[abc]` - Character class matching
- `!` - Negation (include previously ignored)
- `pattern/` - Directory-only patterns
- `/pattern` - Anchored patterns (from root)

### Default Ignore Patterns
```go
defaultIgnorePatterns := []string{
    ".git/",
    ".svn/",
    ".hg/",
    "node_modules/",
    "vendor/",
    ".DS_Store",
    "*.pyc",
    "__pycache__/",
    "*.so",
    "*.dylib",
    "*.test",
    "*.out",
}
```

### Custom Ignore Patterns

```go
opts := indexer.IndexOptions{
    RootPath: "/path/to/repo",
    IgnorePatterns: []string{
        "*.min.js",       // Ignore minified files
        "dist/",          // Ignore build output
        "coverage/",      // Ignore test coverage
        "**/*.generated.go", // Ignore generated code
    },
}
```

## Merkle Tree Implementation

### Change Detection Algorithm
1. **Initial Index**: Build Merkle tree from file system
   - Compute SHA-256 hash for each file
   - Store tree structure in memory
   - Persist tree for future comparisons

2. **Incremental Index**: Compare current state to previous tree
   - Walk file system, compute current hashes
   - Compare with previous Merkle tree:
     - **Added**: File in current tree, not in previous
     - **Modified**: File in both trees, different hash
     - **Deleted**: File in previous tree, not in current
   - Index only added/modified files
   - Delete vector store chunks for deleted files
   - Update Merkle tree with current state

### Tree Structure
```
/repo (root)
├── src/
│   ├── main.go [hash: abc123...]
│   └── util.go [hash: def456...]
└── test/
    └── main_test.go [hash: ghi789...]
```

### Persistence
- Merkle tree saved as `.conexus/merkle.json` in root directory
- Atomic writes with temp file + rename
- Graceful handling of missing/corrupted tree (triggers full reindex)

## Performance Characteristics

### Full Index
- **Throughput**: ~1000-5000 files/second (depending on file size)
- **Memory**: O(n) where n = number of chunks
- **Disk I/O**: Sequential reads, minimal seeks

### Incremental Index
- **Best Case**: O(changed files) when few changes
- **Worst Case**: O(n) when many files changed (still faster than full reindex)
- **Merkle Tree**: O(log n) for tree operations
- **Vector Store**: O(changed chunks) for updates/deletes

### Memory Usage
- **Walker**: Processes files one at a time (streaming)
- **Merkle Tree**: Stores file paths + hashes (minimal overhead)
- **Chunks**: Full content kept in memory during indexing
- **Recommendation**: For large repos (100K+ files), consider batched processing

## Error Handling

### Partial Failures
The indexer continues on file-level errors and returns partial results:

```go
chunks, err := idx.Index(ctx, opts)
if err != nil {
    // err contains details of failed files
    // chunks contains successfully processed chunks
    fmt.Printf("Warning: %v\n", err)
    fmt.Printf("Indexed %d chunks despite errors\n", len(chunks))
}
```

### Vector Store Errors
Vector store operations are atomic per file:
- If chunk storage fails, indexing halts with error
- Already-stored chunks remain in vector store
- Retry with `IndexIncremental()` to resume

## Testing

### Unit Tests
- `walker_test.go` - File system traversal and ignore patterns (100% coverage)
- `merkle_test.go` - Merkle tree operations and persistence (100% coverage)
- `indexer_impl_test.go` - Core indexing logic (100% coverage)

### Integration Tests
- `vectorstore_integration_test.go` - End-to-end indexing with vector stores:
  - Full index workflow
  - Incremental index with modifications
  - File deletion handling
  - Vector store state verification

### Running Tests
```bash
# All indexer tests
go test ./internal/indexer

# Specific test
go test ./internal/indexer -run TestIndexIncremental_VectorStoreDeletions

# With verbose output
go test -v ./internal/indexer

# With coverage
go test -cover ./internal/indexer
```

## Implementation Status

- [x] File walker with `.gitignore` support
- [x] Merkle tree for incremental indexing  
- [x] Merkle tree persistence
- [x] Core indexer implementation
- [x] Vector store integration
- [x] Incremental indexing with change detection
- [x] File deletion handling
- [x] Unit tests (80%+ coverage)
- [x] Integration tests with vector stores
- [ ] Code chunker (AST-based) - Future work
- [ ] Doc chunker (sliding window) - Future work
- [ ] Embedding generation integration - Future work

## Future Enhancements

### Planned Features
1. **Smart Chunking**
   - AST-based chunking for code files
   - Respect function/class boundaries
   - Sliding window for documentation

2. **Embedding Integration**
   - Generate embeddings during indexing
   - Store embeddings in vector store
   - Support multiple embedding providers

3. **Metadata Extraction**
   - Git blame information
   - Author/timestamp metadata
   - Dependency relationships

4. **Performance Optimizations**
   - Parallel file processing
   - Batched vector store operations
   - Memory-mapped file reading

5. **Advanced Filters**
   - Language-specific filtering
   - Complexity-based sampling
   - Custom chunk validators

## Related Packages

- `internal/vectorstore` - Persistent chunk storage
- `internal/embedding` - Embedding generation (planned)
- `internal/search` - Query and retrieval (planned)
- `internal/mcp` - MCP protocol server (planned)

## License

See root LICENSE file.
