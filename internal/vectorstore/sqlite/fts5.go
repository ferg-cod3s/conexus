// Package sqlite provides FTS5 full-text search implementation.
package sqlite

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// SearchBM25 performs sparse keyword search using SQLite FTS5 with BM25 ranking.
// The query is automatically parsed and escaped for FTS5 syntax.
func (s *Store) SearchBM25(ctx context.Context, query string, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Set default limit if not specified
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := opts.Offset

	// Parse and escape query for FTS5
	fts5Query := parseFTS5Query(query)

	// Build the SQL query with metadata filters
	sqlQuery, args := buildBM25Query(fts5Query, opts.Filters, limit, offset)

	// Execute search
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("execute search: %w", err)
	}
	defer rows.Close()

	// Collect results
	var results []vectorstore.SearchResult
	for rows.Next() {
		var doc vectorstore.Document
		var vectorJSON, metadataJSON []byte
		var createdAt, updatedAt int64
		var score float32

		err := rows.Scan(
			&doc.ID,
			&doc.Content,
			&vectorJSON,
			&metadataJSON,
			&createdAt,
			&updatedAt,
			&score,
		)
		if err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}

		// Deserialize vector and metadata (reuse existing logic)
		if err := deserializeDocument(&doc, vectorJSON, metadataJSON, createdAt, updatedAt); err != nil {
			return nil, fmt.Errorf("deserialize document: %w", err)
		}

		// Normalize BM25 score to [0, 1] range
		// FTS5 rank is negative (lower is better), we invert it
		// and normalize. Typical BM25 scores range from -10 to 0.
		normalizedScore := normalizeRank(score)

		// Apply threshold filter
		if opts.Threshold > 0 && normalizedScore < opts.Threshold {
			continue
		}

		results = append(results, vectorstore.SearchResult{
			Document: doc,
			Score:    normalizedScore,
			Method:   "bm25",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate results: %w", err)
	}

	return results, nil
}

// parseFTS5Query converts a user query into FTS5 syntax.
// Handles:
// - Escaping special characters
// - Converting spaces to AND operators
// - Supporting quoted phrases
// - Supporting basic boolean operators (AND, OR, NOT)
func parseFTS5Query(query string) string {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Handle quoted phrases - preserve them
	phrases := extractPhrases(query)

	// Replace phrases with placeholders
	for i, phrase := range phrases {
		placeholder := fmt.Sprintf("__PHRASE_%d__", i)
		query = strings.Replace(query, fmt.Sprintf(`"%s"`, phrase), placeholder, 1)
	}

	// Escape special FTS5 characters in remaining text
	// FTS5 special chars: " ( ) AND OR NOT
	query = escapeFTS5Special(query)

	// Restore phrases with proper quoting
	for i, phrase := range phrases {
		placeholder := fmt.Sprintf("__PHRASE_%d__", i)
		escapedPhrase := escapeFTS5Special(phrase)
		query = strings.Replace(query, placeholder, fmt.Sprintf(`"%s"`, escapedPhrase), 1)
	}

	// Convert boolean operators to uppercase
	query = normalizeOperators(query)

	// If no explicit operators, convert spaces to AND
	// If no explicit operators, convert spaces to AND
	if !containsExplicitOperators(query) {
		words := splitPreservingQuotes(query)
		query = strings.Join(words, " AND ")
	}

	return query
}

// extractPhrases finds all quoted phrases in the query
func extractPhrases(query string) []string {
	re := regexp.MustCompile(`"([^"]+)"`)
	matches := re.FindAllStringSubmatch(query, -1)

	phrases := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			phrases = append(phrases, match[1])
		}
	}
	return phrases
}

// escapeFTS5Special escapes special characters for FTS5
func escapeFTS5Special(s string) string {
	// Escape characters that have special meaning in FTS5
	// FTS5 special characters: " ( ) AND OR NOT, plus we handle / and other punctuation
	// Note: @ is preserved for email addresses and other identifiers
	replacer := strings.NewReplacer(
		`"`, `""`, // Double quotes are escaped by doubling
		`/`, " ", // Replace slashes with spaces to separate path components
		`(`, " ", // Replace parentheses with spaces
		`)`, " ", // Replace parentheses with spaces
		`-`, " ", // Replace hyphens with spaces
	)
	return replacer.Replace(s)
}

// normalizeOperators converts boolean operators to uppercase
func normalizeOperators(query string) string {
	// Use word boundaries to avoid matching substrings
	re := regexp.MustCompile(`\b(and|or|not)\b`)
	return re.ReplaceAllStringFunc(query, func(s string) string {
		return strings.ToUpper(s)
	})
}

// containsExplicitOperators checks if query contains explicit boolean operators
func containsExplicitOperators(query string) bool {
	return strings.Contains(query, " AND ") ||
		strings.Contains(query, " OR ") ||
		strings.Contains(query, " NOT ")
}

// splitPreservingQuotes splits query on spaces but keeps quoted phrases intact
func splitPreservingQuotes(query string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for _, r := range query {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case ' ':
			if inQuotes {
				// Space inside quotes - keep it
				current.WriteRune(r)
			} else if current.Len() > 0 {
				// Space outside quotes - token boundary
				tokens = append(tokens, strings.TrimSpace(current.String()))
				current.Reset()
			}
			// Else: multiple spaces in a row, skip
		default:
			current.WriteRune(r)
		}
	}

	// Don't forget the last token
	if current.Len() > 0 {
		tokens = append(tokens, strings.TrimSpace(current.String()))
	}

	return tokens
}

// buildBM25Query constructs the SQL query for BM25 search with filters
func buildBM25Query(fts5Query string, filters map[string]interface{}, limit int, offset int) (string, []interface{}) {
	baseQuery := `
		SELECT 
			d.id,
			d.content,
			d.vector,
			d.metadata,
			d.created_at,
			d.updated_at,
			fts.rank as score
		FROM documents_fts fts
		JOIN documents d ON fts.id = d.id
		WHERE fts.content MATCH ?
	`

	args := []interface{}{fts5Query}

	// Add metadata filters
	if len(filters) > 0 {
		for key, value := range filters {
			// Use JSON extraction for metadata filtering
			// SQLite JSON functions: json_extract(metadata, '$.key')
			baseQuery += fmt.Sprintf(" AND json_extract(d.metadata, '$.%s') = ?", key)
			args = append(args, value)
		}
	}

	// Order by relevance (rank is negative, lower is better)
	baseQuery += " ORDER BY fts.rank ASC"

	// Add limit and offset
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	return baseQuery, args
}

// normalizeRank converts FTS5 negative rank to positive score in [0, 1]
// FTS5 BM25 rank is negative (typically -10 to 0, where 0 is best match)
func normalizeRank(rank float32) float32 {
	// Invert and clamp to reasonable range
	// Most relevant results are near 0, less relevant go to -10 or lower
	score := -rank

	// Normalize to [0, 1] scale
	// Assume typical range is 0 to 10
	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}

	return score / 10.0
}
