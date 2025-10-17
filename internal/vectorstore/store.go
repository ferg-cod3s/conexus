// Package vectorstore provides storage abstractions for vectors and metadata with hybrid search.
package vectorstore

import (
	"context"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
)

// Document represents a stored chunk with its vector embedding.
type Document struct {
	ID        string                 // Unique document identifier
	Content   string                 // Original text content
	Vector    embedding.Vector       // Dense embedding vector
	Metadata  map[string]interface{} // Arbitrary metadata (file path, language, etc.)
	CreatedAt time.Time              // When the document was stored
	UpdatedAt time.Time              // Last update timestamp
}

// SearchResult represents a single search result with relevance score.
type SearchResult struct {
	Document Document // The matched document
	Score    float32  // Relevance score (higher is better)
	Method   string   // Search method used ("bm25", "vector", "hybrid")
}

// SearchOptions configures search behavior.
type SearchOptions struct {
	Limit     int                    // Maximum number of results
	Offset    int                    // Number of results to skip (for pagination)
	Threshold float32                // Minimum score threshold
	Filters   map[string]interface{} // Metadata filters (e.g., language="go")
	Rerank    bool                   // Apply reranking to results
}

// VectorStore provides hybrid search over stored documents.
type VectorStore interface {
	// Upsert inserts or updates a document with its vector.
	Upsert(ctx context.Context, doc Document) error

	// UpsertBatch efficiently inserts or updates multiple documents.
	UpsertBatch(ctx context.Context, docs []Document) error

	// Delete removes a document by ID.
	Delete(ctx context.Context, id string) error

	// Get retrieves a document by ID.
	Get(ctx context.Context, id string) (*Document, error)

	// SearchVector performs dense vector similarity search.
	SearchVector(ctx context.Context, vector embedding.Vector, opts SearchOptions) ([]SearchResult, error)

	// SearchBM25 performs sparse keyword search using BM25.
	SearchBM25(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error)

	// SearchHybrid combines vector and BM25 search with fusion.
	SearchHybrid(ctx context.Context, query string, vector embedding.Vector, opts SearchOptions) ([]SearchResult, error)

	// Count returns the total number of documents.
	Count(ctx context.Context) (int64, error)

	// ListIndexedFiles returns a list of all unique file paths that have been indexed.
	ListIndexedFiles(ctx context.Context) ([]string, error)

	// GetFileChunks returns all chunks for a specific file path, sorted by start_line.
	GetFileChunks(ctx context.Context, filePath string) ([]Document, error)

	// Close releases resources.
	Close() error
}

// IndexStats provides statistics about the vector store.
type IndexStats struct {
	TotalDocuments int64            // Total documents indexed
	TotalChunks    int64            // Total chunks (same as documents for now)
	Languages      map[string]int64 // Document count per language
	LastIndexedAt  time.Time        // Timestamp of last indexing operation
	IndexSize      int64            // Storage size in bytes
}

// StatsProvider provides statistics about stored data.
type StatsProvider interface {
	// Stats returns current index statistics.
	Stats(ctx context.Context) (*IndexStats, error)
}
