package federation

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"
)

// Merger handles deduplication and merging of results from multiple sources
type Merger struct {
	results        map[string][]interface{}
	contentHashes  map[string]string // content -> hash
	sourceTracking map[string][]string // item ID -> list of sources
	attributions   map[string]map[string]interface{} // item ID -> source metadata
}

// NewMerger creates a new result merger
func NewMerger() *Merger {
	return &Merger{
		results:        make(map[string][]interface{}),
		contentHashes:  make(map[string]string),
		sourceTracking: make(map[string][]string),
		attributions:   make(map[string]map[string]interface{}),
	}
}

// AddResults adds results from a source to the merger
func (m *Merger) AddResults(source string, items []interface{}) {
	if _, exists := m.results[source]; !exists {
		m.results[source] = []interface{}{}
	}
	m.results[source] = append(m.results[source], items...)
}

// MergeAndDeduplicate merges and deduplicates results from all sources
func (m *Merger) MergeAndDeduplicate() ([]interface{}, DeduplicationStats) {
	stats := DeduplicationStats{
		TotalResults: m.countTotalResults(),
	}

	// Map to track unique items
	uniqueItems := make(map[string]interface{})
	duplicateMap := make(map[string]int) // hash -> count

	for source, items := range m.results {
		for _, item := range items {
			itemStr := m.itemToString(item)
			hash := m.hashContent(itemStr)

			if _, exists := uniqueItems[hash]; !exists {
				// First occurrence of this item
				itemID := fmt.Sprintf("%s_%d", source, len(uniqueItems))
				uniqueItems[hash] = item
				m.sourceTracking[itemID] = []string{source}
				m.attributions[itemID] = map[string]interface{}{
					"source": source,
					"hash":   hash,
				}
			} else {
				// Duplicate found
				stats.DuplicatesFound++
				// Find the item ID and add source
				for id, tracked := range m.sourceTracking {
					if !contains(tracked, source) {
						tracked = append(tracked, source)
						m.sourceTracking[id] = tracked
					}
				}
			}

			duplicateMap[hash]++
		}
	}

	// Build merged result list
	mergedItems := make([]interface{}, 0, len(uniqueItems))
	for _, item := range uniqueItems {
		mergedItems = append(mergedItems, item)
	}

	// Sort by most common (appears in most sources)
	sort.Slice(mergedItems, func(i, j int) bool {
		hashI := m.hashContent(m.itemToString(mergedItems[i]))
		hashJ := m.hashContent(m.itemToString(mergedItems[j]))
		return duplicateMap[hashI] > duplicateMap[hashJ]
	})

	stats.UniqueResults = len(uniqueItems)
	stats.MergedResults = len(mergedItems)

	return mergedItems, stats
}

// GetSourceAttributions returns source attribution metadata
func (m *Merger) GetSourceAttributions() map[string]map[string]interface{} {
	return m.attributions
}

// countTotalResults counts total results across all sources
func (m *Merger) countTotalResults() int {
	count := 0
	for _, items := range m.results {
		count += len(items)
	}
	return count
}

// itemToString converts an item to a string representation for comparison
func (m *Merger) itemToString(item interface{}) string {
	switch v := item.(type) {
	case string:
		return v
	case map[string]interface{}:
		// For maps, use content field if available
		if content, ok := v["content"]; ok {
			if s, ok := content.(string); ok {
				return s
			}
		}
		// Otherwise use file_path or id
		if filePath, ok := v["file_path"]; ok {
			if s, ok := filePath.(string); ok {
				return s
			}
		}
		if id, ok := v["id"]; ok {
			if s, ok := id.(string); ok {
				return s
			}
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", item)
	}
}

// hashContent computes a hash of content for deduplication
func (m *Merger) hashContent(content string) string {
	h := md5.New()
	_, _ = io.WriteString(h, content)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// CalculateSimilarity calculates similarity between two items (0-100)
func CalculateSimilarity(item1, item2 interface{}) int {
	str1 := fmt.Sprintf("%v", item1)
	str2 := fmt.Sprintf("%v", item2)

	// Simple string similarity: count common characters
	if str1 == str2 {
		return 100
	}

	common := 0
	maxLen := len(str1)
	if len(str2) > maxLen {
		maxLen = len(str2)
	}

	for i := 0; i < len(str1) && i < len(str2); i++ {
		if str1[i] == str2[i] {
			common++
		}
	}

	if maxLen == 0 {
		return 0
	}

	return (common * 100) / maxLen
}

// helper function
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
