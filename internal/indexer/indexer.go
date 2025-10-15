// Package indexer provides file system traversal, content chunking, and metadata extraction
// for building a searchable codebase index.
package indexer

import (
	"context"
	"io/fs"
	"time"
)

// Chunk represents a unit of indexed content with metadata.
type Chunk struct {
	ID          string            // Unique identifier (hash-based)
	Content     string            // Raw text content
	FilePath    string            // Relative path from repo root
	Language    string            // Programming language or "markdown", "text"
	Type        ChunkType         // Function, class, doc paragraph, etc.
	StartLine   int               // Starting line number in source file
	EndLine     int               // Ending line number in source file
	Metadata    map[string]string // Additional metadata (git commit, author, etc.)
	Hash        string            // Content hash (for deduplication/incremental updates)
	IndexedAt   time.Time         // When this chunk was indexed
}

// ChunkType categorizes the semantic type of a chunk.
type ChunkType string

const (
	ChunkTypeFunction   ChunkType = "function"
	ChunkTypeClass      ChunkType = "class"
	ChunkTypeStruct     ChunkType = "struct"
	ChunkTypeInterface  ChunkType = "interface"
	ChunkTypeComment    ChunkType = "comment"
	ChunkTypeParagraph  ChunkType = "paragraph"  // For docs
	ChunkTypeCodeBlock  ChunkType = "code_block" // For embedded code in docs
	ChunkTypeUnknown    ChunkType = "unknown"
)

// IndexOptions configures indexing behavior.
type IndexOptions struct {
	RootPath      string   // Root directory to index
	IgnorePatterns []string // .gitignore-style patterns to exclude
	MaxFileSize   int64    // Skip files larger than this (bytes)
	IncludeGitInfo bool    // Extract git metadata (commit hash, author)
	ChunkSize     int      // Target chunk size in tokens (for sliding window)
	ChunkOverlap  int      // Overlap between chunks (tokens)
}

// Indexer walks a file system and produces chunks with metadata.
type Indexer interface {
	// Index walks the file system and returns all chunks.
	Index(ctx context.Context, opts IndexOptions) ([]Chunk, error)
	
	// IndexIncremental only indexes files that have changed since last run.
	// Uses Merkle tree hashing to detect changes efficiently.
	IndexIncremental(ctx context.Context, opts IndexOptions, previousState []byte) ([]Chunk, []byte, error)
}

// Chunker splits file content into semantic chunks.
type Chunker interface {
	// Chunk splits content into chunks based on the file type and language.
	Chunk(ctx context.Context, content string, filePath string) ([]Chunk, error)
	
	// Supports returns true if this chunker handles the given file extension.
	Supports(fileExtension string) bool
}

// Walker traverses a file system respecting ignore patterns.
type Walker interface {
	// Walk traverses the directory tree and calls fn for each file.
	Walk(ctx context.Context, root string, ignorePatterns []string, fn func(path string, info fs.FileInfo) error) error
}

// MerkleTree provides content hashing for incremental indexing.
type MerkleTree interface {
	// Hash computes a Merkle tree hash for the given directory.
	// Returns a compact representation of the tree state for later comparison.
	Hash(ctx context.Context, root string, ignorePatterns []string) ([]byte, error)
	
	// Diff compares two tree states and returns paths that changed.
	Diff(ctx context.Context, oldState, newState []byte) ([]string, error)
}
