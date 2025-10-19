// Package sqlite provides hybrid search implementation combining BM25 and vector search.
package sqlite

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// HybridSearchOptions extends SearchOptions with hybrid-specific parameters.
type HybridSearchOptions struct {
	vectorstore.SearchOptions
	Alpha float32 // Weight for vector search vs BM25 (0=BM25 only, 1=vector only, 0.5=equal weight)
	K     int     // RRF constant (default: 60)
}

// SearchHybrid combines BM25 and vector search using Reciprocal Rank Fusion (RRF).
// The alpha parameter controls the balance between semantic (vector) and keyword (BM25) search:
//   - alpha=0.0: BM25 only (pure keyword search)
//   - alpha=0.5: Equal weighting (default)
//   - alpha=1.0: Vector only (pure semantic search)
//
// RRF formula: score = α/(k+rank_vector) + (1-α)/(k+rank_bm25)
// where k is typically 60 (standard RRF constant).
func (s *Store) SearchHybrid(ctx context.Context, query string, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	// Validate inputs
	if query == "" && len(vector) == 0 {
		return nil, fmt.Errorf("must provide either query text or query vector")
	}

	// Build hybrid options with defaults
	hybridOpts := HybridSearchOptions{
		SearchOptions: opts,
		Alpha:         0.5, // Default to equal weighting
		K:             60,  // Standard RRF constant
	}

	// Set default limit if not specified
	if hybridOpts.Limit <= 0 {
		hybridOpts.Limit = 10
	}

	// Fetch more results from each method to ensure we have enough after fusion
	// We request 2x the limit to account for overlaps and get better coverage
	searchLimit := (hybridOpts.Limit + hybridOpts.Offset) * 2
	searchOpts := vectorstore.SearchOptions{
		Limit:     searchLimit,
		Offset:    hybridOpts.Offset,
		Threshold: opts.Threshold,
		Filters:   opts.Filters,
	}

	var bm25Results []vectorstore.SearchResult
	var vectorResults []vectorstore.SearchResult
	var err error

	// Execute BM25 search if query is provided
	if query != "" {
		bm25Results, err = s.SearchBM25(ctx, query, searchOpts)
		if err != nil {
			return nil, fmt.Errorf("bm25 search failed: %w", err)
		}
	}

	// Execute vector search if vector is provided
	if len(vector) > 0 {
		vectorResults, err = s.SearchVector(ctx, vector, searchOpts)
		if err != nil {
			return nil, fmt.Errorf("vector search failed: %w", err)
		}
	}

	// If only one method produced results, return those
	if len(bm25Results) == 0 && len(vectorResults) == 0 {
		return []vectorstore.SearchResult{}, nil
	}
	if len(bm25Results) == 0 {
		return limitResults(vectorResults, hybridOpts.Limit), nil
	}
	if len(vectorResults) == 0 {
		return limitResults(bm25Results, hybridOpts.Limit), nil
	}

	// Apply Reciprocal Rank Fusion
	fusedResults := applyRRF(bm25Results, vectorResults, hybridOpts)

	// Apply metadata-aware reranking if requested
	if opts.Rerank {
		// Compute and apply metadata boosts
		for i := range fusedResults {
			boost := computeMetadataBoost(fusedResults[i].Document, query)
			fusedResults[i].Score += boost
		}
		// Re-sort by boosted scores
		sort.Slice(fusedResults, func(i, j int) bool {
			return fusedResults[i].Score > fusedResults[j].Score
		})
	}

	// Apply threshold filter if specified
	if opts.Threshold > 0 {
		filtered := make([]vectorstore.SearchResult, 0, len(fusedResults))
		for _, result := range fusedResults {
			if result.Score >= opts.Threshold {
				filtered = append(filtered, result)
			}
		}
		fusedResults = filtered
	}

	// Limit final results
	return limitResults(fusedResults, hybridOpts.Limit), nil
}

// computeMetadataBoost calculates a small additive boost based on document metadata.
// Boosts are conservative to avoid disrupting core relevance signals.
func computeMetadataBoost(doc vectorstore.Document, query string) float32 {
	var boost float32
	const maxBoost = 0.006 // Cap total boost to avoid destabilizing rankings

	// Path-based boost: prefer documents whose filename contains query terms
	if filePath, ok := doc.Metadata["path"].(string); ok {
		filename := strings.ToLower(filepath.Base(filePath))
		queryTerms := strings.Fields(strings.ToLower(query))
		for _, term := range queryTerms {
			if strings.Contains(filename, term) {
				boost += 0.0015
				break // Only apply once per document
			}
		}
	} else if filePath, ok := doc.Metadata["file_path"].(string); ok {
		filename := strings.ToLower(filepath.Base(filePath))
		queryTerms := strings.Fields(strings.ToLower(query))
		for _, term := range queryTerms {
			if strings.Contains(filename, term) {
				boost += 0.0015
				break
			}
		}
	}

	// Recency boost: prefer recently updated documents
	if !doc.UpdatedAt.IsZero() {
		daysSinceUpdate := time.Since(doc.UpdatedAt).Hours() / 24
		if daysSinceUpdate <= 7 {
			boost += 0.003
		} else if daysSinceUpdate <= 30 {
			boost += 0.0015
		}
	}

	// Language hint boost: small boost if query mentions the document's language
	if lang, ok := doc.Metadata["language"].(string); ok && lang != "" {
		lowerQuery := strings.ToLower(query)
		lowerLang := strings.ToLower(lang)
		if strings.Contains(lowerQuery, lowerLang) {
			boost += 0.001
		}
	}

	// Cap the boost to prevent it from overwhelming core relevance
	if boost > maxBoost {
		boost = maxBoost
	}
	return boost
}

// applyRRF implements Reciprocal Rank Fusion to combine results from multiple search methods.
// RRF formula: RRF_score(d) = Σ 1/(k + rank_i(d)) for each ranker i
// With weighting: score = α * vector_rrf + (1-α) * bm25_rrf
func applyRRF(bm25Results, vectorResults []vectorstore.SearchResult, opts HybridSearchOptions) []vectorstore.SearchResult {
	k := float32(opts.K)
	alpha := opts.Alpha

	// Validate alpha is in [0, 1]
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}

	// Create rank maps: document ID -> rank position (0-based)
	bm25Ranks := make(map[string]int, len(bm25Results))
	for i, result := range bm25Results {
		bm25Ranks[result.Document.ID] = i
	}

	vectorRanks := make(map[string]int, len(vectorResults))
	for i, result := range vectorResults {
		vectorRanks[result.Document.ID] = i
	}

	// Collect all unique document IDs
	docMap := make(map[string]*vectorstore.SearchResult)
	for _, result := range bm25Results {
		r := result // Copy to avoid pointer issues
		docMap[r.Document.ID] = &r
	}
	for _, result := range vectorResults {
		if _, exists := docMap[result.Document.ID]; !exists {
			r := result
			docMap[r.Document.ID] = &r
		}
	}

	// Calculate RRF scores for each document
	fusedResults := make([]vectorstore.SearchResult, 0, len(docMap))
	for docID, result := range docMap {
		var rrfScore float32

		// Add BM25 RRF contribution
		if rank, found := bm25Ranks[docID]; found {
			bm25RRF := 1.0 / (k + float32(rank))
			rrfScore += (1 - alpha) * bm25RRF
		}

		// Add vector RRF contribution
		if rank, found := vectorRanks[docID]; found {
			vectorRRF := 1.0 / (k + float32(rank))
			rrfScore += alpha * vectorRRF
		}

		fusedResults = append(fusedResults, vectorstore.SearchResult{
			Document: result.Document,
			Score:    rrfScore,
			Method:   "hybrid",
		})
	}

	// Sort by RRF score (descending)
	sort.Slice(fusedResults, func(i, j int) bool {
		return fusedResults[i].Score > fusedResults[j].Score
	})

	return fusedResults
}

// limitResults truncates results to the specified limit.
func limitResults(results []vectorstore.SearchResult, limit int) []vectorstore.SearchResult {
	if limit <= 0 {
		return results
	}
	if len(results) <= limit {
		return results
	}
	return results[:limit]
}
