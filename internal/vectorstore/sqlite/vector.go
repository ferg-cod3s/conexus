// Package sqlite provides vector similarity search implementation.
package sqlite

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// SearchVector performs dense vector similarity search using brute-force K-nearest neighbors.
// This is a simple but correct implementation suitable for small to medium datasets (< 100k docs).
// Future optimizations: HNSW index, product quantization, or GPU acceleration.
func (s *Store) SearchVector(ctx context.Context, queryVector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	// Validate input
	if len(queryVector) == 0 {
		return nil, fmt.Errorf("query vector cannot be empty")
	}

	// Check for zero-length vector (would cause division by zero in cosine similarity)
	if vectorMagnitude(queryVector) == 0 {
		return nil, fmt.Errorf("query vector has zero magnitude")
	}

	// Set default limit if not specified
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := opts.Offset

	// Build SQL query to fetch all documents
	// Note: In production, we'd use spatial indexes or approximate nearest neighbor algorithms
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

	// Execute query
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	// Calculate similarity for each document
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

		// Calculate cosine similarity
		similarity := cosineSimilarity(queryVector, doc.Vector)

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
