package federation

import (
	"fmt"
	"regexp"
	"strings"
)

// Detector handles cross-source relationship detection
type Detector struct {
	idPatterns map[string]*regexp.Regexp
}

// NewDetector creates a new relationship detector
func NewDetector() *Detector {
	return &Detector{
		idPatterns: map[string]*regexp.Regexp{
			// Common ID patterns
			"uuid":        regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
			"number":      regexp.MustCompile(`^\d+$`),
			"github_issue": regexp.MustCompile(`^#?\d+$`),
			"jira_ticket": regexp.MustCompile(`^[A-Z]+-\d+$`),
			"slug":        regexp.MustCompile(`^[a-z0-9\-_]+$`),
		},
	}
}

// DetectRelationships detects cross-source relationships
func (d *Detector) DetectRelationships(results []*QueryResult, mergedItems []interface{}) map[string][]string {
	relationships := make(map[string][]string)

	if len(results) <= 1 {
		return relationships
	}

	// Extract IDs from results grouped by source
	sourceIDs := make(map[string]map[string]interface{}) // source -> id -> item
	for _, result := range results {
		if result.Error != nil {
			continue
		}
		sourceIDs[result.Source] = d.extractIDs(result.Items)
	}

	// Find matching IDs across sources
	sources := make([]string, 0, len(sourceIDs))
	for source := range sourceIDs {
		sources = append(sources, source)
	}

	for i := 0; i < len(sources); i++ {
		for j := i + 1; j < len(sources); j++ {
			sourceA := sources[i]
			sourceB := sources[j]

			idsA := sourceIDs[sourceA]
			idsB := sourceIDs[sourceB]

			// Find matching IDs
			for idA, itemA := range idsA {
				for idB, itemB := range idsB {
					if d.isRelated(itemA, itemB) {
						relationships[idA] = append(relationships[idA], idB)
						relationships[idB] = append(relationships[idB], idA)
					}
				}
			}
		}
	}

	return relationships
}

// extractIDs extracts IDs from items for relationship detection
func (d *Detector) extractIDs(items []interface{}) map[string]interface{} {
	idMap := make(map[string]interface{})

	for i, item := range items {
		var id string

		switch v := item.(type) {
		case map[string]interface{}:
			// Try common ID fields
			if idVal, ok := v["id"]; ok {
				id = fmt.Sprintf("%v", idVal)
			} else if idVal, ok := v["ticket_id"]; ok {
				id = fmt.Sprintf("%v", idVal)
			} else if idVal, ok := v["issue_id"]; ok {
				id = fmt.Sprintf("%v", idVal)
			} else if idVal, ok := v["file_path"]; ok {
				id = fmt.Sprintf("%v", idVal)
			} else {
				id = fmt.Sprintf("item_%d", i)
			}
		default:
			id = fmt.Sprintf("item_%d", i)
		}

		if id != "" {
			idMap[id] = item
		}
	}

	return idMap
}

// isRelated checks if two items are related
func (d *Detector) isRelated(item1, item2 interface{}) bool {
	str1 := d.itemToComparable(item1)
	str2 := d.itemToComparable(item2)

	if str1 == "" || str2 == "" {
		return false
	}

	// Exact match
	if str1 == str2 {
		return true
	}

	// Check for common substrings (indicates possible relationship)
	if d.hasSimilarContent(str1, str2) {
		return true
	}

	// Check if one ID is contained in the other
	if strings.Contains(str1, str2) || strings.Contains(str2, str1) {
		return true
	}

	return false
}

// itemToComparable extracts a comparable string from an item
func (d *Detector) itemToComparable(item interface{}) string {
	switch v := item.(type) {
	case map[string]interface{}:
		// Try to extract a unique identifier
		if id, ok := v["id"]; ok {
			return fmt.Sprintf("%v", id)
		}
		if id, ok := v["ticket_id"]; ok {
			return fmt.Sprintf("%v", id)
		}
		if id, ok := v["issue_id"]; ok {
			return fmt.Sprintf("%v", id)
		}
		// Use file path if available
		if fp, ok := v["file_path"]; ok {
			return fmt.Sprintf("%v", fp)
		}
		return ""
	default:
		return fmt.Sprintf("%v", item)
	}
}

// hasSimilarContent checks if two strings have similar content
func (d *Detector) hasSimilarContent(str1, str2 string) bool {
	// Normalize strings
	normalized1 := strings.ToLower(strings.TrimSpace(str1))
	normalized2 := strings.ToLower(strings.TrimSpace(str2))

	// Check for significant overlap (70%+ common characters in order)
	commonLen := d.longestCommonSubstring(normalized1, normalized2)
	minLen := len(normalized1)
	if len(normalized2) < minLen {
		minLen = len(normalized2)
	}

	if minLen == 0 {
		return false
	}

	return float64(commonLen)/float64(minLen) > 0.7
}

// longestCommonSubstring finds the longest common substring
func (d *Detector) longestCommonSubstring(s1, s2 string) int {
	max := 0
	m := len(s1)
	n := len(s2)

	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			k := 0
			for i+k < m && j+k < n && s1[i+k] == s2[j+k] {
				k++
			}
			if k > max {
				max = k
			}
		}
	}

	return max
}

// BuildRelationshipGraph creates a graph of relationships
func (d *Detector) BuildRelationshipGraph(relationships map[string][]string) map[string]map[string]bool {
	graph := make(map[string]map[string]bool)

	for id, related := range relationships {
		if _, exists := graph[id]; !exists {
			graph[id] = make(map[string]bool)
		}

		for _, relID := range related {
			graph[id][relID] = true
		}
	}

	return graph
}
