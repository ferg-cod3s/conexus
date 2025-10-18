// Package federation implements multi-source result federation for search queries.
package federation

import (
	"strings"

	"github.com/ferg-cod3s/conexus/internal/schema"
)

// Detector finds relationships between search results from different sources
type Detector struct {
}

func NewDetector() *Detector {
	return &Detector{}
}

// DetectRelationships finds relationships between results from different sources
func (d *Detector) DetectRelationships(results []schema.SearchResultItem) []Relationship {
	var relationships []Relationship

	// Group results by source type
	sourceGroups := make(map[string][]schema.SearchResultItem)
	for _, result := range results {
		sourceGroups[result.SourceType] = append(sourceGroups[result.SourceType], result)
	}

	// Detect relationships within and between source groups
	for sourceType, group := range sourceGroups {
		// Check relationships within the same source type
		rels := d.detectWithinSourceRelationships(sourceType, group)
		relationships = append(relationships, rels...)
	}

	// Detect cross-source relationships
	sourceTypes := make([]string, 0, len(sourceGroups))
	for sourceType := range sourceGroups {
		sourceTypes = append(sourceTypes, sourceType)
	}

	for i, sourceType1 := range sourceTypes {
		for j := i + 1; j < len(sourceTypes); j++ {
			sourceType2 := sourceTypes[j]
			group1 := sourceGroups[sourceType1]
			group2 := sourceGroups[sourceType2]

			// Find relationships between these two source groups
			rels := d.detectCrossSourceRelationships(sourceType1, group1, sourceType2, group2)
			relationships = append(relationships, rels...)
		}
	}

	return relationships
}

// detectWithinSourceRelationships finds relationships within the same source type
func (d *Detector) detectWithinSourceRelationships(sourceType string, results []schema.SearchResultItem) []Relationship {
	var relationships []Relationship

	for i, item1 := range results {
		for j := i + 1; j < len(results); j++ {
			item2 := results[j]
			if rel := d.detectRelationship(item1, item2); rel != nil {
				relationships = append(relationships, *rel)
			}
		}
	}

	return relationships
}

// detectCrossSourceRelationships finds relationships between two different source types
func (d *Detector) detectCrossSourceRelationships(sourceType1 string, group1 []schema.SearchResultItem, sourceType2 string, group2 []schema.SearchResultItem) []Relationship {
	var relationships []Relationship

	for _, item1 := range group1 {
		for _, item2 := range group2 {
			if rel := d.detectRelationship(item1, item2); rel != nil {
				relationships = append(relationships, *rel)
			}
		}
	}

	return relationships
}

// detectRelationship determines if two search results are related
func (d *Detector) detectRelationship(item1, item2 schema.SearchResultItem) *Relationship {
	// Check for file path relationships
	filePath1, hasFile1 := item1.Metadata["file_path"].(string)
	filePath2, hasFile2 := item2.Metadata["file_path"].(string)

	if hasFile1 && hasFile2 {
		if rel := d.detectFileRelationship(filePath1, filePath2, item1, item2); rel != nil {
			return rel
		}
	}

	// Check for ticket ID relationships
	ticketID1, hasTicket1 := item1.Metadata["ticket_id"].(string)
	ticketID2, hasTicket2 := item2.Metadata["ticket_id"].(string)

	if hasTicket1 && hasTicket2 && ticketID1 == ticketID2 {
		return &Relationship{
			Type:       RelationTypeSameTicket,
			Item1ID:    item1.ID,
			Item2ID:    item2.ID,
			Confidence: 1.0,
			Metadata: map[string]interface{}{
				"ticket_id": ticketID1,
			},
		}
	}

	// Check for content similarity (same entity mentioned)
	if d.contentSimilarity(item1.Content, item2.Content) > 0.7 {
		return &Relationship{
			Type:       RelationTypeSameEntity,
			Item1ID:    item1.ID,
			Item2ID:    item2.ID,
			Confidence: 0.8,
		}
	}

	return nil
}

// detectFileRelationship checks for file-based relationships
func (d *Detector) detectFileRelationship(path1, path2 string, item1, item2 schema.SearchResultItem) *Relationship {
	// Same file
	if path1 == path2 {
		return &Relationship{
			Type:       RelationTypeSameFile,
			Item1ID:    item1.ID,
			Item2ID:    item2.ID,
			Confidence: 1.0,
			Metadata: map[string]interface{}{
				"file_path": path1,
			},
		}
	}

	// Test file relationship
	if d.isTestFileRelationship(path1, path2) {
		return &Relationship{
			Type:       RelationTypeTestFile,
			Item1ID:    item1.ID,
			Item2ID:    item2.ID,
			Confidence: 0.9,
			Metadata: map[string]interface{}{
				"test_file":   path1,
				"source_file": path2,
			},
		}
	}

	// Documentation relationship
	if d.isDocumentationRelationship(path1, path2) {
		return &Relationship{
			Type:       RelationTypeDocumentation,
			Item1ID:    item1.ID,
			Item2ID:    item2.ID,
			Confidence: 0.8,
			Metadata: map[string]interface{}{
				"doc_file":     path1,
				"source_file":  path2,
			},
		}
	}

	return nil
}

// isTestFileRelationship checks if files have a test relationship
func (d *Detector) isTestFileRelationship(path1, path2 string) bool {
	// Use simple heuristics - could be enhanced with the existing relationship detector
	return strings.Contains(path1, "_test.") || strings.Contains(path2, "_test.") ||
		   strings.Contains(path1, ".test.") || strings.Contains(path2, ".test.") ||
		   strings.Contains(path1, ".spec.") || strings.Contains(path2, ".spec.")
}

// isDocumentationRelationship checks if one file is documentation for another
func (d *Detector) isDocumentationRelationship(path1, path2 string) bool {
	docExts := []string{".md", ".rst", ".txt", ".adoc"}
	for _, ext := range docExts {
		if strings.HasSuffix(path1, ext) || strings.HasSuffix(path2, ext) {
			return true
		}
	}
	return strings.Contains(strings.ToLower(path1), "readme") ||
		   strings.Contains(strings.ToLower(path2), "readme") ||
		   strings.Contains(path1, "/docs/") || strings.Contains(path2, "/docs/")
}

// contentSimilarity calculates simple text similarity
func (d *Detector) contentSimilarity(text1, text2 string) float32 {
	// Simple word overlap similarity
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	if len(words1) == 0 || len(words2) == 0 {
		return 0
	}

	wordSet1 := make(map[string]bool)
	wordSet2 := make(map[string]bool)

	for _, word := range words1 {
		wordSet1[word] = true
	}
	for _, word := range words2 {
		wordSet2[word] = true
	}

	intersection := 0
	for word := range wordSet1 {
		if wordSet2[word] {
			intersection++
		}
	}

	union := len(wordSet1) + len(wordSet2) - intersection
	if union == 0 {
		return 0
	}

	return float32(intersection) / float32(union)
}

// Relationship represents a detected relationship between two search results
type Relationship struct {
	Type       string                 `json:"type"`
	Item1ID    string                 `json:"item1_id"`
	Item2ID    string                 `json:"item2_id"`
	Confidence float32                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Relationship types
const (
	RelationTypeSameFile       = "same_file"
	RelationTypeSameTicket     = "same_ticket"
	RelationTypeSameEntity     = "same_entity"
	RelationTypeTestFile       = "test_file"
	RelationTypeDocumentation  = "documentation"
)
