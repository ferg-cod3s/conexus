// Package vectorstore provides in-memory vector storage for POC and testing.
package vectorstore

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
)

// MemoryStore is an in-memory implementation of VectorStore for POC and testing.
// Thread-safe with RWMutex for concurrent access.
type MemoryStore struct {
	mu        sync.RWMutex
	documents map[string]Document // ID -> Document mapping
	index     []string            // Ordered list of document IDs for iteration
}

// NewMemoryStore creates a new in-memory vector store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		documents: make(map[string]Document),
		index:     make([]string, 0),
	}
}

// Upsert inserts or updates a document with its vector.
func (m *MemoryStore) Upsert(ctx context.Context, doc Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document ID cannot be empty")
	}
	if len(doc.Vector) == 0 {
		return fmt.Errorf("document vector cannot be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if document exists
	if existing, exists := m.documents[doc.ID]; exists {
		// Update timestamp
		doc.CreatedAt = existing.CreatedAt
		doc.UpdatedAt = time.Now()
	} else {
		// New document
		now := time.Now()
		if doc.CreatedAt.IsZero() {
			doc.CreatedAt = now
		}
		if doc.UpdatedAt.IsZero() {
			doc.UpdatedAt = now
		}
		m.index = append(m.index, doc.ID)
	}

	m.documents[doc.ID] = doc
	return nil
}

// UpsertBatch efficiently inserts or updates multiple documents.
func (m *MemoryStore) UpsertBatch(ctx context.Context, docs []Document) error {
	for _, doc := range docs {
		if err := m.Upsert(ctx, doc); err != nil {
			return fmt.Errorf("upsert document %s: %w", doc.ID, err)
		}
	}
	return nil
}

// Delete removes a document by ID.
func (m *MemoryStore) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.documents[id]; !exists {
		return fmt.Errorf("document %s not found", id)
	}

	delete(m.documents, id)

	// Remove from index
	for i, docID := range m.index {
		if docID == id {
			m.index = append(m.index[:i], m.index[i+1:]...)
			break
		}
	}

	return nil
}

// Get retrieves a document by ID.
func (m *MemoryStore) Get(ctx context.Context, id string) (*Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	doc, exists := m.documents[id]
	if !exists {
		return nil, fmt.Errorf("document %s not found", id)
	}

	return &doc, nil
}

// SearchVector performs dense vector similarity search using cosine similarity.
func (m *MemoryStore) SearchVector(ctx context.Context, vector embedding.Vector, opts SearchOptions) ([]SearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []SearchResult

	// Compute similarity for each document
	for _, docID := range m.index {
		doc := m.documents[docID]

		// Apply metadata filters
		if !matchesFilters(doc, opts.Filters) {
			continue
		}

		var score float32
		// If no vector provided, this is a metadata-only query
		if len(vector) == 0 {
			score = 1.0 // All matching documents have equal score
		} else {
			// Compute cosine similarity
			score = cosineSimilarity(vector, doc.Vector)

			// Apply threshold
			if opts.Threshold > 0 && score < opts.Threshold {
				continue
			}
		}
		results = append(results, SearchResult{
			Document: doc,
			Score:    score,
			Method:   "vector",
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// SearchBM25 performs sparse keyword search using simple BM25 implementation.
// Note: This is a simplified BM25 for POC. Production should use proper text search engine.
func (m *MemoryStore) SearchBM25(ctx context.Context, query string, opts SearchOptions) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	queryTerms := tokenize(query)
	if len(queryTerms) == 0 {
		return nil, fmt.Errorf("query contains no valid terms")
	}

	// Compute IDF for query terms
	idf := m.computeIDF(queryTerms)

	var results []SearchResult

	// Score each document
	for _, docID := range m.index {
		doc := m.documents[docID]

		// Apply metadata filters
		if !matchesFilters(doc, opts.Filters) {
			continue
		}

		// Compute BM25 score
		score := m.bm25Score(doc.Content, queryTerms, idf)

		// Apply threshold
		if opts.Threshold > 0 && score < opts.Threshold {
			continue
		}

		results = append(results, SearchResult{
			Document: doc,
			Score:    score,
			Method:   "bm25",
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// SearchHybrid combines vector and BM25 search with reciprocal rank fusion.
func (m *MemoryStore) SearchHybrid(ctx context.Context, query string, vector embedding.Vector, opts SearchOptions) ([]SearchResult, error) {
	// Get vector search results
	vectorResults, err := m.SearchVector(ctx, vector, SearchOptions{
		Limit:     opts.Limit * 2, // Get more for fusion
		Threshold: 0,              // Apply threshold after fusion
		Filters:   opts.Filters,
	})
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}

	// Get BM25 search results
	bm25Results, err := m.SearchBM25(ctx, query, SearchOptions{
		Limit:     opts.Limit * 2,
		Threshold: 0,
		Filters:   opts.Filters,
	})
	if err != nil {
		return nil, fmt.Errorf("bm25 search: %w", err)
	}

	// Reciprocal Rank Fusion
	fusedScores := make(map[string]float32)
	const k = 60.0 // RRF constant

	// Add vector search scores
	for rank, result := range vectorResults {
		fusedScores[result.Document.ID] += 1.0 / (k + float32(rank+1))
	}

	// Add BM25 scores
	for rank, result := range bm25Results {
		fusedScores[result.Document.ID] += 1.0 / (k + float32(rank+1))
	}

	// Build result list
	var results []SearchResult
	for docID, score := range fusedScores {
		if opts.Threshold > 0 && score < opts.Threshold {
			continue
		}

		doc := m.documents[docID]
		results = append(results, SearchResult{
			Document: doc,
			Score:    score,
			Method:   "hybrid",
		})
	}

	// Sort by fused score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// Count returns the total number of documents.
func (m *MemoryStore) Count(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.documents)), nil
}

// Close releases resources (no-op for memory store).
func (m *MemoryStore) Close() error {
	return nil
}

// Stats returns index statistics.
func (m *MemoryStore) Stats(ctx context.Context) (*IndexStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &IndexStats{
		TotalDocuments: int64(len(m.documents)),
		TotalChunks:    int64(len(m.documents)),
		Languages:      make(map[string]int64),
	}

	var lastIndexed time.Time
	for _, doc := range m.documents {
		// Count by language
		if lang, ok := doc.Metadata["language"].(string); ok {
			stats.Languages[lang]++
		}

		// Track latest indexed time
		if doc.UpdatedAt.After(lastIndexed) {
			lastIndexed = doc.UpdatedAt
		}
	}

	stats.LastIndexedAt = lastIndexed
	return stats, nil
}

// Helper functions

// cosineSimilarity computes the cosine similarity between two vectors.
func cosineSimilarity(a, b embedding.Vector) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, magA, magB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(magA))) * float32(math.Sqrt(float64(magB))))
}

// matchesFilters checks if a document matches all metadata filters.
func matchesFilters(doc Document, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	for key, expectedValue := range filters {
		actualValue, exists := doc.Metadata[key]
		if !exists || actualValue != expectedValue {
			return false
		}
	}

	return true
}

// tokenize splits text into lowercase terms.
func tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	var terms []string
	for _, word := range words {
		// Remove common punctuation
		word = strings.Trim(word, ".,!?;:\"'()[]{}") 
		if len(word) > 0 {
			terms = append(terms, word)
		}
	}
	return terms
}

// computeIDF calculates inverse document frequency for query terms.
func (m *MemoryStore) computeIDF(terms []string) map[string]float32 {
	idf := make(map[string]float32)
	totalDocs := float32(len(m.documents))

	for _, term := range terms {
		docsWithTerm := 0
		for _, doc := range m.documents {
			if strings.Contains(strings.ToLower(doc.Content), term) {
				docsWithTerm++
			}
		}
		if docsWithTerm > 0 {
			idf[term] = float32(math.Log(float64(totalDocs / float32(docsWithTerm))))
		}
	}

	return idf
}

// bm25Score computes simplified BM25 score for a document.
func (m *MemoryStore) bm25Score(content string, queryTerms []string, idf map[string]float32) float32 {
	const k1 = 1.5
	const b = 0.75

	contentLower := strings.ToLower(content)
	docTerms := tokenize(contentLower)
	docLength := float32(len(docTerms))

	// Compute average document length
	avgDocLength := m.avgDocLength()

	// Term frequency
	tf := make(map[string]int)
	for _, term := range docTerms {
		tf[term]++
	}

	// Compute BM25 score
	var score float32
	for _, term := range queryTerms {
		termFreq := float32(tf[term])
		if termFreq == 0 {
			continue
		}

		idfScore := idf[term]
		numerator := termFreq * (k1 + 1)
		denominator := termFreq + k1*(1-b+b*(docLength/avgDocLength))

		score += idfScore * (numerator / denominator)
	}

	return score
}

// avgDocLength computes average document length across all documents.
func (m *MemoryStore) avgDocLength() float32 {
	if len(m.documents) == 0 {
		return 1
	}

	var totalLength int
	for _, doc := range m.documents {
		totalLength += len(tokenize(doc.Content))
	}

	return float32(totalLength) / float32(len(m.documents))
}
