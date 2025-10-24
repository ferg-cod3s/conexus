// Package indexer provides file system traversal, content chunking, and metadata extraction
// for building a searchable codebase index.
package indexer

import (
	"context"
	"io/fs"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// Chunk represents a unit of indexed content with metadata.
type Chunk struct {
	ID        string            // Unique identifier (hash-based)
	Content   string            // Raw text content
	FilePath  string            // Relative path from repo root
	Language  string            // Programming language or "markdown", "text"
	Type      ChunkType         // Function, class, doc paragraph, etc.
	StartLine int               // Starting line number in source file
	EndLine   int               // Ending line number in source file
	Metadata  map[string]string // Additional metadata (git commit, author, etc.)
	Hash      string            // Content hash (for deduplication/incremental updates)
	IndexedAt time.Time         // When this chunk was indexed

	// New fields for story context
	StoryIDs     []string `json:"story_ids,omitempty"`
	TicketIDs    []string `json:"ticket_ids,omitempty"`
	PRNumbers    []string `json:"pr_numbers,omitempty"`
	DiscussionID string   `json:"discussion_id,omitempty"`
	BranchName   string   `json:"branch_name,omitempty"`
}

// ChunkType categorizes the semantic type of a chunk.
type ChunkType string

const (
	ChunkTypeFunction  ChunkType = "function"
	ChunkTypeClass     ChunkType = "class"
	ChunkTypeStruct    ChunkType = "struct"
	ChunkTypeInterface ChunkType = "interface"
	ChunkTypeComment   ChunkType = "comment"
	ChunkTypeParagraph ChunkType = "paragraph"  // For docs
	ChunkTypeCodeBlock ChunkType = "code_block" // For embedded code in docs
	ChunkTypeUnknown   ChunkType = "unknown"
)

// IndexOptions configures indexing behavior.
type IndexOptions struct {
	RootPath       string                  // Root directory to index
	IgnorePatterns []string                // .gitignore-style patterns to exclude
	MaxFileSize    int64                   // Skip files larger than this (bytes)
	IncludeGitInfo bool                    // Extract git metadata (commit hash, author)
	ChunkSize      int                     // Target chunk size in tokens (for sliding window)
	ChunkOverlap   int                     // Overlap between chunks (tokens)
	Embedder       embedding.Embedder      // Optional: Embedder for generating vectors
	VectorStore    vectorstore.VectorStore // Optional: VectorStore for storing vectors
}

// Indexer walks a file system and produces chunks with metadata.
type Indexer interface {
	// Index walks the file system and returns all chunks.
	Index(ctx context.Context, opts IndexOptions) ([]Chunk, error)

	// IndexIncremental only indexes files that have changed since last run.
	// Uses Merkle tree hashing to detect changes efficiently.
	IndexIncremental(ctx context.Context, opts IndexOptions, previousState []byte) ([]Chunk, []byte, error)

	// GetStatus returns current indexing status and progress.
	GetStatus() IndexStatus
}

// IndexController manages background indexing operations.
type IndexController interface {
	// Start begins background indexing with the given options.
	Start(ctx context.Context, opts IndexOptions) error

	// Stop gracefully stops background indexing.
	Stop(ctx context.Context) error

	// ForceReindex performs a complete reindex of the codebase.
	ForceReindex(ctx context.Context, opts IndexOptions) error

	// ReindexPaths reindexes only the specified paths.
	ReindexPaths(ctx context.Context, opts IndexOptions, paths []string) error

	// GetStatus returns current indexing status.
	GetStatus() IndexStatus

	// HealthCheck performs health validation of the index.
	HealthCheck(ctx context.Context) error
}

// IndexStatus represents the current status of indexing operations.
type IndexStatus struct {
	IsIndexing     bool         // Whether indexing is currently running
	Phase          string       // Current phase (scanning, chunking, embedding, etc.)
	Progress       float64      // Progress as percentage (0-100)
	FilesProcessed int          // Number of files processed so far
	TotalFiles     int          // Total number of files to process (if known)
	ChunksCreated  int          // Number of chunks created so far
	StartTime      time.Time    // When indexing started
	EstimatedEnd   time.Time    // Estimated completion time
	LastError      string       // Last error encountered
	Metrics        IndexMetrics // Current metrics
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
