// Package federation implements multi-source result federation for search queries.
package federation

import (
	"sort"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/schema"
)


// Merger combines and deduplicates search results from multiple connectors
type Merger struct {
	deduplicationThreshold float32
}

// NewMerger creates a new result merger
func NewMerger() *Merger {
	return &Merger{
		deduplicationThreshold: 0.85, // 85% similarity threshold for deduplication
	}
}

// Merge combines and ranks results from multiple connectors
func (m *Merger) Merge(connectorResults []ConnectorResult) []schema.SearchResultItem {
	if len(connectorResults) == 0 {
		return []schema.SearchResultItem{}
	}

	// Flatten all results
	var allResults []schema.SearchResultItem
	for _, cr := range connectorResults {
		for _, result := range cr.Results {
			// Add connector metadata
			if result.Metadata == nil {
				result.Metadata = make(map[string]interface{})
			}
			result.Metadata["connector_id"] = cr.ConnectorID
			result.Metadata["connector_type"] = cr.ConnectorType
			allResults = append(allResults, result)
		}
	}

	if len(allResults) == 0 {
		return allResults
	}

	// Deduplicate results
	dedupedResults := m.deduplicate(allResults)

	// Apply cross-source ranking
	rankedResults := m.rankBySourceDiversity(dedupedResults)

	// Normalize scores to [0, 1] range
	normalizedResults := m.normalizeScores(rankedResults)

	// Sort by final score
	sort.Slice(normalizedResults, func(i, j int) bool {
		return normalizedResults[i].Score > normalizedResults[j].Score
	})

	return normalizedResults
}

// deduplicate removes duplicate or very similar results
func (m *Merger) deduplicate(results []schema.SearchResultItem) []schema.SearchResultItem {
	if len(results) <= 1 {
		return results
	}

	var deduped []schema.SearchResultItem
	seen := make(map[string]bool)

	for _, result := range results {
		// Create a signature based on content similarity
		signature := m.createContentSignature(result.Content)

		if !seen[signature] {
			deduped = append(deduped, result)
			seen[signature] = true
		} else {
			// If duplicate, keep the one with higher score
			for i, existing := range deduped {
				if m.createContentSignature(existing.Content) == signature {
					if result.Score > existing.Score {
						deduped[i] = result
					}
					break
				}
			}
		}
	}

	return deduped
}

// rankBySourceDiversity boosts results that come from diverse sources
func (m *Merger) rankBySourceDiversity(results []schema.SearchResultItem) []schema.SearchResultItem {
	if len(results) <= 1 {
		return results
	}

	// Count results per source type
	sourceCounts := make(map[string]int)
	for _, result := range results {
		sourceCounts[result.SourceType]++
	}

	// Apply diversity bonus
	diversityBonus := float32(0.1) // 10% bonus for underrepresented sources
	totalSources := len(sourceCounts)

	for i := range results {
		sourceCount := sourceCounts[results[i].SourceType]
		// Bonus is higher when this source type is less common
		bonus := diversityBonus * float32(totalSources-sourceCount+1) / float32(totalSources)
		results[i].Score *= (1.0 + bonus)
	}

	return results
}

// normalizeScores normalizes all scores to the [0, 1] range
func (m *Merger) normalizeScores(results []schema.SearchResultItem) []schema.SearchResultItem {
	if len(results) == 0 {
		return results
	}

	// Find max score
	maxScore := results[0].Score
	for _, result := range results[1:] {
		if result.Score > maxScore {
			maxScore = result.Score
		}
	}

	// If all scores are 0, return as is (all normalized to 0)
	if maxScore == 0 {
		return results
	}

	// Normalize all scores to [0, 1]
	for i := range results {
		results[i].Score = results[i].Score / maxScore
		// Ensure score is within bounds (handle floating point errors)
		if results[i].Score < 0 {
			results[i].Score = 0
		}
		if results[i].Score > 1 {
			results[i].Score = 1
		}
	}

	return results
}

// createContentSignature generates a simplified signature for content deduplication
func (m *Merger) createContentSignature(content string) string {
	// Simple approach: normalize whitespace and take first 100 characters
	normalized := strings.Fields(content)
	if len(normalized) > 20 { // Limit to first 20 words
		normalized = normalized[:20]
	}
	return strings.Join(normalized, " ")
}
