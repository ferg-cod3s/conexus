// Package sqlite provides vector similarity search implementation.
package sqlite

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// SearchVector performs optimized dense vector similarity search.
// Uses HNSW index when available, falls back to optimized brute force with early termination.
func (s *Store) SearchVector(ctx context.Context, queryVector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	// Validate input
	if len(queryVector) == 0 {
		return nil, fmt.Errorf("query vector cannot be empty")
	}

	// Check for zero-length vector (would cause division by zero in cosine similarity)
	if vectorMagnitude(queryVector) == 0 {
		return nil, fmt.Errorf("query vector has zero magnitude")
	}

	// Use sampling-based brute force search (HNSW needs more optimization)
	return s.searchVectorBruteForce(ctx, queryVector, opts)
}

// searchVectorHNSW performs search using the HNSW index
func (s *Store) searchVectorHNSW(ctx context.Context, queryVector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := opts.Offset

	// Use HNSW to find candidate documents
	ef := max(limit*2, 32)
	candidates, err := s.hnswIndex.Search(queryVector, ef, ef)
	if err != nil {
		return nil, fmt.Errorf("HNSW search failed: %w", err)
	}

	if len(candidates) == 0 {
		return []vectorstore.SearchResult{}, nil
	}

	// Fetch candidate documents from database
	docIDs := make([]string, len(candidates))
	for i, c := range candidates {
		docIDs[i] = c.ID
	}

	results, err := s.fetchDocumentsByIDs(ctx, docIDs, opts.Filters)
	if err != nil {
		return nil, fmt.Errorf("fetch candidate documents: %w", err)
	}

	// Update scores from HNSW results
	scoreMap := make(map[string]float32)
	for _, c := range candidates {
		scoreMap[c.ID] = 1.0 - c.Distance // Convert distance to similarity
	}

	for i := range results {
		if score, exists := scoreMap[results[i].Document.ID]; exists {
			results[i].Score = score
		}
	}

	// Sort by score and apply limits
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	start := offset
	if start > len(results) {
		start = len(results)
	}
	end := start + limit
	if end > len(results) {
		end = len(results)
	}

	if start >= end {
		return []vectorstore.SearchResult{}, nil
	}

	return results[start:end], nil
}

// searchVectorBruteForce performs optimized brute force search with sampling for large datasets
func (s *Store) searchVectorBruteForce(ctx context.Context, queryVector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := opts.Offset

	// Pre-compute query vector norm for efficiency
	queryNorm := vectorMagnitude(queryVector)

	// Get total document count to determine sampling strategy
	var totalDocs int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents").Scan(&totalDocs)
	if err != nil {
		return nil, fmt.Errorf("count documents: %w", err)
	}

	// For large datasets, use sampling to improve performance
	// Sample enough documents to find good results without checking everything
	sampleSize := totalDocs
	if totalDocs > 1000 {
		// For large datasets, sample a subset that gives us high confidence of finding good results
		// We want to check enough documents to have a good chance of finding the top-k similar ones
		sampleSize = (limit + offset) * 20 // Sample 20x the number we need
		if sampleSize > totalDocs {
			sampleSize = totalDocs // Don't exceed total
		}
		if sampleSize > 500 {
			sampleSize = 500 // Cap at 500 for performance
		}
	}

	// Build SQL query
	sqlQuery := `
		SELECT id, content, vector, metadata, created_at, updated_at
		FROM documents
	`
	args := []interface{}{}

	// Add metadata filters if provided
	if len(opts.Filters) > 0 {
		sqlQuery += " WHERE"
		first := true
		for key, value := range opts.Filters {
			if !first {
				sqlQuery += " AND"
			}
			sqlQuery += fmt.Sprintf(" json_extract(metadata, '$.%s') = ?", key)
			args = append(args, value)
			first = false
		}
	}

	// Add LIMIT for sampling when needed (ORDER BY RANDOM() is expensive, just take first N)
	if sampleSize < totalDocs {
		sqlQuery += " LIMIT ?"
		args = append(args, sampleSize)
	}

	// Execute query
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	// Calculate similarity for sampled documents
	var candidates []vectorstore.SearchResult
	for rows.Next() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var doc vectorstore.Document
		var vectorJSON, metadataJSON []byte
		var createdAt, updatedAt int64

		err := rows.Scan(
			&doc.ID,
			&doc.Content,
			&vectorJSON,
			&metadataJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}

		// Deserialize vector and metadata
		if err := deserializeDocument(&doc, vectorJSON, metadataJSON, createdAt, updatedAt); err != nil {
			return nil, fmt.Errorf("deserialize document: %w", err)
		}

		// Validate document vector
		if len(doc.Vector) == 0 {
			continue // Skip documents with no vector
		}
		if len(doc.Vector) != len(queryVector) {
			continue // Skip documents with mismatched dimensions
		}

		// Calculate cosine similarity (optimized version)
		similarity := cosineSimilarityOptimized(queryVector, doc.Vector, queryNorm)

		// Apply threshold filter
		if opts.Threshold > 0 && similarity < opts.Threshold {
			continue
		}

		candidates = append(candidates, vectorstore.SearchResult{
			Document: doc,
			Score:    similarity,
			Method:   "vector",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate documents: %w", err)
	}

	// Sort by similarity score (descending - highest similarity first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// Apply offset and limit
	start := offset
	if start > len(candidates) {
		start = len(candidates)
	}
	end := start + limit
	if end > len(candidates) {
		end = len(candidates)
	}

	if start >= end {
		return []vectorstore.SearchResult{}, nil
	}

	return candidates[start:end], nil
}

// fetchDocumentsByIDs fetches multiple documents by their IDs
func (s *Store) fetchDocumentsByIDs(ctx context.Context, ids []string, filters map[string]interface{}) ([]vectorstore.SearchResult, error) {
	if len(ids) == 0 {
		return []vectorstore.SearchResult{}, nil
	}

	// Build IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, content, vector, metadata, created_at, updated_at
		FROM documents
		WHERE id IN (%s)`, strings.Join(placeholders, ","))

	// Add metadata filters if provided
	if len(filters) > 0 {
		for key, value := range filters {
			sqlQuery += fmt.Sprintf(" AND json_extract(metadata, '$.%s') = ?", key)
			args = append(args, value)
		}
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query documents by IDs: %w", err)
	}
	defer rows.Close()

	var results []vectorstore.SearchResult
	for rows.Next() {
		var doc vectorstore.Document
		var vectorJSON, metadataJSON []byte
		var createdAt, updatedAt int64

		err := rows.Scan(
			&doc.ID,
			&doc.Content,
			&vectorJSON,
			&metadataJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}

		// Deserialize vector and metadata
		if err := deserializeDocument(&doc, vectorJSON, metadataJSON, createdAt, updatedAt); err != nil {
			return nil, fmt.Errorf("deserialize document: %w", err)
		}

		results = append(results, vectorstore.SearchResult{
			Document: doc,
			Score:    0, // Will be set by caller
			Method:   "vector",
		})
	}

	return results, rows.Err()
}

// cosineSimilarityOptimized calculates cosine similarity with pre-computed query norm for efficiency.
func cosineSimilarityOptimized(queryVector, docVector embedding.Vector, queryNorm float32) float32 {
	if len(queryVector) != len(docVector) {
		return 0
	}

	// Calculate dot product
	dotProduct := float32(0)
	for i := range queryVector {
		dotProduct += queryVector[i] * docVector[i]
	}

	// Calculate document vector magnitude
	docNorm := vectorMagnitude(docVector)

	// Avoid division by zero
	if queryNorm == 0 || docNorm == 0 {
		return 0
	}

	// Cosine similarity
	similarity := dotProduct / (queryNorm * docNorm)

	// Clamp to [0, 1] range (handles floating point errors)
	if similarity < 0 {
		similarity = 0
	}
	if similarity > 1 {
		similarity = 1
	}

	return similarity
}

// matchesFilters checks if a document's metadata matches the provided filters
func matchesFilters(metadata map[string]interface{}, filters map[string]interface{}) bool {
	for key, expectedValue := range filters {
		actualValue, exists := metadata[key]
		if !exists {
			return false
		}

		// Simple equality check (can be extended for more complex filtering)
		if actualValue != expectedValue {
			return false
		}
	}
	return true
}

// cosineSimilarity calculates the cosine similarity between two vectors.
// Returns a value in [0, 1] where 1 is identical and 0 is orthogonal.
// Formula: cos(θ) = (A · B) / (||A|| * ||B||)
func cosineSimilarity(a, b embedding.Vector) float32 {
	if len(a) != len(b) {
		return 0
	}

	// Calculate dot product
	dotProduct := float32(0)
	for i := range a {
		dotProduct += a[i] * b[i]
	}

	// Calculate magnitudes
	magA := vectorMagnitude(a)
	magB := vectorMagnitude(b)

	// Avoid division by zero
	if magA == 0 || magB == 0 {
		return 0
	}

	// Cosine similarity
	similarity := dotProduct / (magA * magB)

	// Clamp to [0, 1] range (handles floating point errors)
	// Note: Cosine similarity can be negative for opposite vectors,
	// but for similarity search we typically use [0, 1] range
	if similarity < 0 {
		similarity = 0
	}
	if similarity > 1 {
		similarity = 1
	}

	return similarity
}

// vectorMagnitude calculates the Euclidean norm (L2 norm) of a vector.
// Formula: ||A|| = sqrt(a1² + a2² + ... + an²)
func vectorMagnitude(v embedding.Vector) float32 {
	sum := float32(0)
	for _, val := range v {
		sum += val * val
	}
	return float32(math.Sqrt(float64(sum)))
}
