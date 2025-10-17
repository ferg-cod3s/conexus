// Package search provides hybrid search and reranking capabilities.
package search

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// Query represents a search query with optional filters and parameters.
type Query struct {
	Text       string                 // Search query text
	Filters    map[string]interface{} // Metadata filters
	Limit      int                    // Maximum results to return
	Threshold  float32                // Minimum relevance score
	HybridMode HybridMode             // How to combine sparse and dense results
}

// HybridMode controls how sparse (BM25) and dense (vector) results are combined.
type HybridMode string

const (
	HybridModeRRF      HybridMode = "rrf"      // Reciprocal Rank Fusion
	HybridModeWeighted HybridMode = "weighted" // Weighted sum of scores
	HybridModeSparse   HybridMode = "sparse"   // BM25 only
	HybridModeDense    HybridMode = "dense"    // Vector only
)

// Result represents a search result with provenance information.
type Result struct {
	Document     vectorstore.Document // The matched document
	Score        float32              // Final relevance score
	SparseScore  float32              // BM25 score (if applicable)
	DenseScore   float32              // Vector similarity score (if applicable)
	RerankedFrom int                  // Original rank before reranking (-1 if not reranked)
}

// Retriever performs hybrid search over a vector store.
type Retriever interface {
	// Retrieve performs a hybrid search query.
	Retrieve(ctx context.Context, query Query) ([]Result, error)
}

// Reranker re-scores and re-orders search results.
type Reranker interface {
	// Rerank re-scores results based on the original query.
	// Returns re-ordered results (may change the list length if filtering).
	Rerank(ctx context.Context, query string, results []Result) ([]Result, error)
}

// FusionStrategy combines multiple ranked lists into a single ranking.
type FusionStrategy interface {
	// Fuse combines sparse and dense search results.
	Fuse(ctx context.Context, sparseResults, denseResults []vectorstore.SearchResult, mode HybridMode) ([]Result, error)
}

// Pipeline orchestrates the full retrieval pipeline: search → fuse → rerank.
type Pipeline struct {
	Store    vectorstore.VectorStore
	Embedder embedding.Embedder
	Fusion   FusionStrategy
	Reranker Reranker // Optional
}

// NewPipeline creates a new search pipeline.
func NewPipeline(store vectorstore.VectorStore, embedder embedding.Embedder, fusion FusionStrategy, reranker Reranker) *Pipeline {
	return &Pipeline{
		Store:    store,
		Embedder: embedder,
		Fusion:   fusion,
		Reranker: reranker,
	}
}

// Search executes the full search pipeline.
func (p *Pipeline) Search(ctx context.Context, query Query) ([]Result, error) {
	// Generate query embedding
	emb, err := p.Embedder.Embed(ctx, query.Text)
	if err != nil {
		return nil, err
	}

	// Perform hybrid search based on mode
	searchOpts := vectorstore.SearchOptions{
		Limit:     query.Limit * 2, // Get more candidates for reranking
		Threshold: query.Threshold,
		Filters:   query.Filters,
	}

	var sparseResults, denseResults []vectorstore.SearchResult

	if query.HybridMode == HybridModeSparse || query.HybridMode == HybridModeRRF || query.HybridMode == HybridModeWeighted {
		sparseResults, err = p.Store.SearchBM25(ctx, query.Text, searchOpts)
		if err != nil {
			return nil, err
		}
	}

	if query.HybridMode == HybridModeDense || query.HybridMode == HybridModeRRF || query.HybridMode == HybridModeWeighted {
		denseResults, err = p.Store.SearchVector(ctx, emb.Vector, searchOpts)
		if err != nil {
			return nil, err
		}
	}

	// Fuse results
	results, err := p.Fusion.Fuse(ctx, sparseResults, denseResults, query.HybridMode)
	if err != nil {
		return nil, err
	}

	// Apply reranking if configured
	if p.Reranker != nil {
		results, err = p.Reranker.Rerank(ctx, query.Text, results)
		if err != nil {
			return nil, err
		}
	}

	// Trim to requested limit
	if len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

// SearchCache provides caching for search results to improve performance.
type SearchCache struct {
	mu      sync.RWMutex
	entries map[string]*CachedSearchResult
	maxSize int
	ttl     time.Duration
}

// CachedSearchResult represents a cached search result.
type CachedSearchResult struct {
	Query     string
	Filters   map[string]interface{}
	Results   []vectorstore.SearchResult
	QueryTime float64
	Timestamp time.Time
}

// NewSearchCache creates a new search cache with the specified size and TTL.
func NewSearchCache(maxSize int, ttl time.Duration) *SearchCache {
	return &SearchCache{
		entries: make(map[string]*CachedSearchResult),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// generateKey creates a cache key from query and filters.
func (c *SearchCache) generateKey(query string, filters map[string]interface{}) string {
	// Create a deterministic key from query and filters
	key := fmt.Sprintf("%s|%v", query, filters)
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// Get retrieves a cached search result if available and not expired.
func (c *SearchCache) Get(query string, filters map[string]interface{}) (*CachedSearchResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateKey(query, filters)
	result, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Since(result.Timestamp) > c.ttl {
		// Remove expired entry
		delete(c.entries, key)
		return nil, false
	}

	return result, true
}

// Set stores a search result in the cache.
func (c *SearchCache) Set(query string, filters map[string]interface{}, results []vectorstore.SearchResult, queryTime float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if at capacity
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	key := c.generateKey(query, filters)
	c.entries[key] = &CachedSearchResult{
		Query:     query,
		Filters:   filters,
		Results:   results,
		QueryTime: queryTime,
		Timestamp: time.Now(),
	}
}

// evictOldest removes the oldest cache entry.
func (c *SearchCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// Clear removes all cached entries.
func (c *SearchCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CachedSearchResult)
}

// Size returns the current number of cached entries.
func (c *SearchCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
