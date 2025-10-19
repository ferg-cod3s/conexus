package federation

import (
"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDetector tests detector initialization
func TestNewDetector(t *testing.T) {
	d := NewDetector()
	require.NotNil(t, d)
	assert.NotNil(t, d.idPatterns)
	assert.Equal(t, 5, len(d.idPatterns))
	assert.NotNil(t, d.idPatterns["uuid"])
	assert.NotNil(t, d.idPatterns["number"])
	assert.NotNil(t, d.idPatterns["github_issue"])
	assert.NotNil(t, d.idPatterns["jira_ticket"])
	assert.NotNil(t, d.idPatterns["slug"])
}

// TestDetectRelationships_EmptyResults tests with no results
func TestDetectRelationships_EmptyResults(t *testing.T) {
	d := NewDetector()
	relationships := d.DetectRelationships([]*QueryResult{}, []interface{}{})
	assert.Equal(t, 0, len(relationships))
}

// TestDetectRelationships_SingleSource tests with only one source
func TestDetectRelationships_SingleSource(t *testing.T) {
	d := NewDetector()
	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"id": "item1"},
				map[string]interface{}{"id": "item2"},
			},
		},
	}
	relationships := d.DetectRelationships(results, nil)
	assert.Equal(t, 0, len(relationships))
}

// TestDetectRelationships_ExactIDMatch tests exact ID matching across sources
func TestDetectRelationships_ExactIDMatch(t *testing.T) {
	d := NewDetector()
	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"id": "PROJ-123"},
			},
		},
		{
			Source: "source2",
			Items: []interface{}{
				map[string]interface{}{"id": "PROJ-123"},
			},
		},
	}
	relationships := d.DetectRelationships(results, nil)
	assert.Greater(t, len(relationships), 0)
	assert.Contains(t, relationships["PROJ-123"], "PROJ-123")
}

// TestDetectRelationships_TicketIDMatch tests ticket ID matching
func TestDetectRelationships_TicketIDMatch(t *testing.T) {
	d := NewDetector()
	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"ticket_id": "JIRA-456"},
			},
		},
		{
			Source: "source2",
			Items: []interface{}{
				map[string]interface{}{"ticket_id": "JIRA-456"},
			},
		},
	}
	relationships := d.DetectRelationships(results, nil)
	assert.Greater(t, len(relationships), 0)
}

// TestDetectRelationships_IssueIDMatch tests issue ID matching
func TestDetectRelationships_IssueIDMatch(t *testing.T) {
	d := NewDetector()
	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"issue_id": "789"},
			},
		},
		{
			Source: "source2",
			Items: []interface{}{
				map[string]interface{}{"issue_id": "789"},
			},
		},
	}
	relationships := d.DetectRelationships(results, nil)
	assert.Greater(t, len(relationships), 0)
}

// TestDetectRelationships_MultipleMatches tests multiple matching items
func TestDetectRelationships_MultipleMatches(t *testing.T) {
	d := NewDetector()
	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"id": "item1"},
				map[string]interface{}{"id": "item2"},
				map[string]interface{}{"id": "item3"},
			},
		},
		{
			Source: "source2",
			Items: []interface{}{
				map[string]interface{}{"id": "item1"},
				map[string]interface{}{"id": "item2"},
				map[string]interface{}{"id": "item4"},
			},
		},
	}
	relationships := d.DetectRelationships(results, nil)
	assert.Greater(t, len(relationships), 0)
	assert.Contains(t, relationships["item1"], "item1")
	assert.Contains(t, relationships["item2"], "item2")
}

// TestDetectRelationships_ErrorSourceSkipped tests that error sources are skipped
func TestDetectRelationships_ErrorSourceSkipped(t *testing.T) {
	d := NewDetector()
	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"id": "item1"},
			},
		},
		{
			Source:  "source2",
			Error:   fmt.Errorf("connection error"),
			Items:   nil,
		},
	}
	relationships := d.DetectRelationships(results, nil)
	// Should not crash, relationships may be empty
	assert.NotNil(t, relationships)
}

// TestDetectRelationships_ThreeSources tests relationship detection across three sources
func TestDetectRelationships_ThreeSources(t *testing.T) {
	d := NewDetector()
	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"id": "ISSUE-100"},
			},
		},
		{
			Source: "source2",
			Items: []interface{}{
				map[string]interface{}{"ticket_id": "ISSUE-100"},
			},
		},
		{
			Source: "source3",
			Items: []interface{}{
				map[string]interface{}{"issue_id": "ISSUE-100"},
			},
		},
	}
	relationships := d.DetectRelationships(results, nil)
	assert.Greater(t, len(relationships), 0)
}

// TestExtractIDs_MapWithID tests extracting ID from map with id field
func TestExtractIDs_MapWithID(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		map[string]interface{}{"id": "item1", "name": "First"},
		map[string]interface{}{"id": "item2", "name": "Second"},
	}
	idMap := d.extractIDs(items)
	assert.Equal(t, 2, len(idMap))
	assert.Contains(t, idMap, "item1")
	assert.Contains(t, idMap, "item2")
}

// TestExtractIDs_MapWithTicketID tests extracting ID from map with ticket_id field
func TestExtractIDs_MapWithTicketID(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		map[string]interface{}{"ticket_id": "TICKET-1"},
		map[string]interface{}{"ticket_id": "TICKET-2"},
	}
	idMap := d.extractIDs(items)
	assert.Equal(t, 2, len(idMap))
	assert.Contains(t, idMap, "TICKET-1")
	assert.Contains(t, idMap, "TICKET-2")
}

// TestExtractIDs_MapWithIssueID tests extracting ID from map with issue_id field
func TestExtractIDs_MapWithIssueID(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		map[string]interface{}{"issue_id": "ISSUE-100"},
		map[string]interface{}{"issue_id": "ISSUE-200"},
	}
	idMap := d.extractIDs(items)
	assert.Equal(t, 2, len(idMap))
	assert.Contains(t, idMap, "ISSUE-100")
	assert.Contains(t, idMap, "ISSUE-200")
}

// TestExtractIDs_MapWithFilePath tests extracting ID from map with file_path field
func TestExtractIDs_MapWithFilePath(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		map[string]interface{}{"file_path": "/path/to/file1.txt"},
		map[string]interface{}{"file_path": "/path/to/file2.txt"},
	}
	idMap := d.extractIDs(items)
	assert.Equal(t, 2, len(idMap))
	assert.Contains(t, idMap, "/path/to/file1.txt")
	assert.Contains(t, idMap, "/path/to/file2.txt")
}

// TestExtractIDs_MapWithoutID tests extracting ID from map without ID fields
func TestExtractIDs_MapWithoutID(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		map[string]interface{}{"name": "First"},
		map[string]interface{}{"name": "Second"},
	}
	idMap := d.extractIDs(items)
	assert.Equal(t, 2, len(idMap))
	assert.Contains(t, idMap, "item_0")
	assert.Contains(t, idMap, "item_1")
}

// TestExtractIDs_NonMapItems tests extracting ID from non-map items
func TestExtractIDs_NonMapItems(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		"string_item_1",
		"string_item_2",
		42,
	}
	idMap := d.extractIDs(items)
	assert.Equal(t, 3, len(idMap))
	assert.Contains(t, idMap, "item_0")
	assert.Contains(t, idMap, "item_1")
	assert.Contains(t, idMap, "item_2")
}

// TestExtractIDs_EmptyList tests extracting IDs from empty list
func TestExtractIDs_EmptyList(t *testing.T) {
	d := NewDetector()
	items := []interface{}{}
	idMap := d.extractIDs(items)
	assert.Equal(t, 0, len(idMap))
}

// TestExtractIDs_MixedIDFields tests extracting IDs with mixed field names
func TestExtractIDs_MixedIDFields(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		map[string]interface{}{"id": "id1"},
		map[string]interface{}{"ticket_id": "ticket1"},
		map[string]interface{}{"issue_id": "issue1"},
		map[string]interface{}{"file_path": "/file1"},
	}
	idMap := d.extractIDs(items)
	assert.Equal(t, 4, len(idMap))
	assert.Contains(t, idMap, "id1")
	assert.Contains(t, idMap, "ticket1")
	assert.Contains(t, idMap, "issue1")
	assert.Contains(t, idMap, "/file1")
}

// TestIsRelated_ExactMatch tests exact match detection
func TestIsRelated_ExactMatch(t *testing.T) {
	d := NewDetector()
	item1 := map[string]interface{}{"id": "PROJ-123"}
	item2 := map[string]interface{}{"id": "PROJ-123"}
	assert.True(t, d.isRelated(item1, item2))
}

// TestIsRelated_Substring tests substring detection
func TestIsRelated_Substring(t *testing.T) {
	d := NewDetector()
	item1 := map[string]interface{}{"id": "PROJ-123"}
	item2 := map[string]interface{}{"id": "123"}
	assert.True(t, d.isRelated(item1, item2))
}

// TestIsRelated_NoMatch tests non-matching items
func TestIsRelated_NoMatch(t *testing.T) {
	d := NewDetector()
	item1 := map[string]interface{}{"id": "ABC-111"}
	item2 := map[string]interface{}{"id": "XYZ-999"}
	assert.False(t, d.isRelated(item1, item2))
}

// TestIsRelated_EmptyComparable tests with empty comparable strings
func TestIsRelated_EmptyComparable(t *testing.T) {
	d := NewDetector()
	item1 := map[string]interface{}{}
	item2 := map[string]interface{}{"id": "PROJ-123"}
	assert.False(t, d.isRelated(item1, item2))
}

// TestItemToComparable_MapWithID tests extracting comparable from map with id
func TestItemToComparable_MapWithID(t *testing.T) {
	d := NewDetector()
	item := map[string]interface{}{"id": "ITEM-100"}
	result := d.itemToComparable(item)
	assert.Equal(t, "ITEM-100", result)
}

// TestItemToComparable_MapWithTicketID tests extracting comparable from map with ticket_id
func TestItemToComparable_MapWithTicketID(t *testing.T) {
	d := NewDetector()
	item := map[string]interface{}{"ticket_id": "TICKET-50"}
	result := d.itemToComparable(item)
	assert.Equal(t, "TICKET-50", result)
}

// TestItemToComparable_MapWithIssueID tests extracting comparable from map with issue_id
func TestItemToComparable_MapWithIssueID(t *testing.T) {
	d := NewDetector()
	item := map[string]interface{}{"issue_id": "ISSUE-75"}
	result := d.itemToComparable(item)
	assert.Equal(t, "ISSUE-75", result)
}

// TestItemToComparable_MapWithFilePath tests extracting comparable from map with file_path
func TestItemToComparable_MapWithFilePath(t *testing.T) {
	d := NewDetector()
	item := map[string]interface{}{"file_path": "/path/to/file"}
	result := d.itemToComparable(item)
	assert.Equal(t, "/path/to/file", result)
}

// TestItemToComparable_MapWithoutIDFields tests extracting comparable from map without ID fields
func TestItemToComparable_MapWithoutIDFields(t *testing.T) {
	d := NewDetector()
	item := map[string]interface{}{"name": "Something"}
	result := d.itemToComparable(item)
	assert.Equal(t, "", result)
}

// TestItemToComparable_String tests extracting comparable from string
func TestItemToComparable_String(t *testing.T) {
	d := NewDetector()
	result := d.itemToComparable("test_string")
	assert.Equal(t, "test_string", result)
}

// TestItemToComparable_Number tests extracting comparable from number
func TestItemToComparable_Number(t *testing.T) {
	d := NewDetector()
	result := d.itemToComparable(42)
	assert.Equal(t, "42", result)
}

// TestHasSimilarContent_HighSimilarity tests high similarity detection
func TestHasSimilarContent_HighSimilarity(t *testing.T) {
	d := NewDetector()
	str1 := "project-management-system"
	str2 := "project-management-system"
	assert.True(t, d.hasSimilarContent(str1, str2))
}

// TestHasSimilarContent_PartialSimilarity tests partial similarity detection
func TestHasSimilarContent_PartialSimilarity(t *testing.T) {
	d := NewDetector()
	str1 := "project-management-tool"
	str2 := "project-management-system"
	assert.True(t, d.hasSimilarContent(str1, str2))
}

// TestHasSimilarContent_LowSimilarity tests low similarity detection
func TestHasSimilarContent_LowSimilarity(t *testing.T) {
	d := NewDetector()
	str1 := "abcdefghij"
	str2 := "zyxwvutsrq"
	assert.False(t, d.hasSimilarContent(str1, str2))
}

// TestHasSimilarContent_CaseInsensitive tests case-insensitive comparison
func TestHasSimilarContent_CaseInsensitive(t *testing.T) {
	d := NewDetector()
	str1 := "PROJECT-123"
	str2 := "project-123"
	assert.True(t, d.hasSimilarContent(str1, str2))
}

// TestHasSimilarContent_WithWhitespace tests whitespace handling
func TestHasSimilarContent_WithWhitespace(t *testing.T) {
	d := NewDetector()
	str1 := "  project-123  "
	str2 := "project-123"
	assert.True(t, d.hasSimilarContent(str1, str2))
}

// TestHasSimilarContent_EmptyStrings tests with empty strings
func TestHasSimilarContent_EmptyStrings(t *testing.T) {
	d := NewDetector()
	assert.False(t, d.hasSimilarContent("", ""))
	assert.False(t, d.hasSimilarContent("", "something"))
	assert.False(t, d.hasSimilarContent("something", ""))
}

// TestLongestCommonSubstring_IdenticalStrings tests with identical strings
func TestLongestCommonSubstring_IdenticalStrings(t *testing.T) {
	d := NewDetector()
	result := d.longestCommonSubstring("hello", "hello")
	assert.Equal(t, 5, result)
}

// TestLongestCommonSubstring_PartialMatch tests with partial match
func TestLongestCommonSubstring_PartialMatch(t *testing.T) {
	d := NewDetector()
	result := d.longestCommonSubstring("hello", "helloworld")
	assert.Equal(t, 5, result)
}

// TestLongestCommonSubstring_NoMatch tests with no common substring
func TestLongestCommonSubstring_NoMatch(t *testing.T) {
	d := NewDetector()
	result := d.longestCommonSubstring("abc", "xyz")
	assert.Equal(t, 0, result)
}

// TestLongestCommonSubstring_SingleCharMatch tests with single character match
func TestLongestCommonSubstring_SingleCharMatch(t *testing.T) {
	d := NewDetector()
	result := d.longestCommonSubstring("abc", "def")
	assert.Equal(t, 0, result)
}

// TestLongestCommonSubstring_ComplexMatch tests with complex strings
func TestLongestCommonSubstring_ComplexMatch(t *testing.T) {
	d := NewDetector()
	result := d.longestCommonSubstring("programming", "grammar")
	assert.Greater(t, result, 0)
}

// TestLongestCommonSubstring_EmptyString tests with empty strings
func TestLongestCommonSubstring_EmptyString(t *testing.T) {
	d := NewDetector()
	assert.Equal(t, 0, d.longestCommonSubstring("", ""))
	assert.Equal(t, 0, d.longestCommonSubstring("", "hello"))
	assert.Equal(t, 0, d.longestCommonSubstring("hello", ""))
}

// TestBuildRelationshipGraph_Empty tests graph building with empty relationships
func TestBuildRelationshipGraph_Empty(t *testing.T) {
	d := NewDetector()
	graph := d.BuildRelationshipGraph(map[string][]string{})
	assert.Equal(t, 0, len(graph))
}

// TestBuildRelationshipGraph_SingleNode tests graph with single node
func TestBuildRelationshipGraph_SingleNode(t *testing.T) {
	d := NewDetector()
	relationships := map[string][]string{
		"item1": {"item2", "item3"},
	}
	graph := d.BuildRelationshipGraph(relationships)
	assert.Equal(t, 1, len(graph))
	assert.NotNil(t, graph["item1"])
	assert.True(t, graph["item1"]["item2"])
	assert.True(t, graph["item1"]["item3"])
}

// TestBuildRelationshipGraph_MultipleNodes tests graph with multiple nodes
func TestBuildRelationshipGraph_MultipleNodes(t *testing.T) {
	d := NewDetector()
	relationships := map[string][]string{
		"item1": {"item2", "item3"},
		"item2": {"item1"},
		"item3": {"item1"},
	}
	graph := d.BuildRelationshipGraph(relationships)
	assert.Equal(t, 3, len(graph))
	assert.True(t, graph["item1"]["item2"])
	assert.True(t, graph["item2"]["item1"])
	assert.True(t, graph["item3"]["item1"])
}

// TestBuildRelationshipGraph_ComplexGraph tests graph with complex relationships
func TestBuildRelationshipGraph_ComplexGraph(t *testing.T) {
	d := NewDetector()
	relationships := map[string][]string{
		"A": {"B", "C"},
		"B": {"A", "C"},
		"C": {"A", "B"},
		"D": {"E"},
		"E": {"D"},
	}
	graph := d.BuildRelationshipGraph(relationships)
	assert.Equal(t, 5, len(graph))
	assert.Equal(t, 2, len(graph["A"]))
	assert.Equal(t, 2, len(graph["B"]))
	assert.Equal(t, 2, len(graph["C"]))
	assert.Equal(t, 1, len(graph["D"]))
	assert.Equal(t, 1, len(graph["E"]))
}

// TestBuildRelationshipGraph_DuplicateRelationships tests handling of duplicate relationships
func TestBuildRelationshipGraph_DuplicateRelationships(t *testing.T) {
	d := NewDetector()
	relationships := map[string][]string{
		"item1": {"item2", "item2", "item3"},
	}
	graph := d.BuildRelationshipGraph(relationships)
	assert.Equal(t, 1, len(graph))
	assert.True(t, graph["item1"]["item2"])
	assert.True(t, graph["item1"]["item3"])
}

// TestDetectorIntegration_CompleteWorkflow tests complete detector workflow
func TestDetectorIntegration_CompleteWorkflow(t *testing.T) {
	d := NewDetector()

	// Create multi-source results with related items
	results := []*QueryResult{
		{
			Source: "github",
			Items: []interface{}{
				map[string]interface{}{"id": "github-issue-123", "title": "Bug report"},
			},
		},
		{
			Source: "jira",
			Items: []interface{}{
				map[string]interface{}{"ticket_id": "PROJ-456", "title": "Bug report"},
			},
		},
		{
			Source: "local",
			Items: []interface{}{
				map[string]interface{}{"file_path": "/repo/issue-123.md"},
			},
		},
	}

	// Extract relationships
	relationships := d.DetectRelationships(results, nil)
	assert.NotNil(t, relationships)

	// Build relationship graph
	graph := d.BuildRelationshipGraph(relationships)
	assert.NotNil(t, graph)
}

// TestDetectorIntegration_LargeScale tests detector with many sources and items
func TestDetectorIntegration_LargeScale(t *testing.T) {
	d := NewDetector()

	// Create 5 sources with 20 items each, some matching
	var results []*QueryResult
	for src := 1; src <= 5; src++ {
		var items []interface{}
		for i := 1; i <= 20; i++ {
			items = append(items, map[string]interface{}{
				"id": i,
				"source": src,
			})
		}
		results = append(results, &QueryResult{
			Source: "source",
			Items:  items,
		})
	}

	relationships := d.DetectRelationships(results, nil)
	assert.NotNil(t, relationships)

	graph := d.BuildRelationshipGraph(relationships)
	assert.NotNil(t, graph)
}

// TestDetectorIntegration_MixedIDTypes tests detector with different ID field types
func TestDetectorIntegration_MixedIDTypes(t *testing.T) {
	d := NewDetector()

	results := []*QueryResult{
		{
			Source: "source1",
			Items: []interface{}{
				map[string]interface{}{"id": "item-001"},
				map[string]interface{}{"ticket_id": "TICKET-100"},
				map[string]interface{}{"issue_id": "ISSUE-50"},
			},
		},
		{
			Source: "source2",
			Items: []interface{}{
				map[string]interface{}{"id": "item-001"},
				map[string]interface{}{"file_path": "/code/item-001.txt"},
			},
		},
	}

	relationships := d.DetectRelationships(results, nil)
	graph := d.BuildRelationshipGraph(relationships)
	assert.Greater(t, len(graph), 0)
}

// TestDetectorEdgeCase_SpecialCharacters tests with special characters in IDs
func TestDetectorEdgeCase_SpecialCharacters(t *testing.T) {
	d := NewDetector()
	item1 := map[string]interface{}{"id": "proj_123-abc"}
	item2 := map[string]interface{}{"id": "proj_123-abc"}
	assert.True(t, d.isRelated(item1, item2))
}

// TestDetectorEdgeCase_UnicodeCharacters tests with unicode characters
func TestDetectorEdgeCase_UnicodeCharacters(t *testing.T) {
	d := NewDetector()
	item1 := map[string]interface{}{"id": "プロジェクト-123"}
	item2 := map[string]interface{}{"id": "プロジェクト-123"}
	assert.True(t, d.isRelated(item1, item2))
}

// TestDetectorEdgeCase_VeryLongStrings tests with very long strings
func TestDetectorEdgeCase_VeryLongStrings(t *testing.T) {
	d := NewDetector()
	long1 := "a very long string that contains a lot of characters for testing purposes " +
		"and should still work correctly when comparing similar long strings with slight differences"
	long2 := "a very long string that contains a lot of characters for testing purposes " +
		"and should still work correctly when comparing similar long strings with small divergences"
	result := d.hasSimilarContent(long1, long2)
	assert.True(t, result)
}

// TestDetectorEdgeCase_NumericStrings tests with numeric string comparison
func TestDetectorEdgeCase_NumericStrings(t *testing.T) {
	d := NewDetector()
	assert.True(t, d.isRelated(
		map[string]interface{}{"id": "12345"},
		map[string]interface{}{"id": "12345"},
	))
}

// TestDetectorEdgeCase_NilValues tests handling of nil values
func TestDetectorEdgeCase_NilValues(t *testing.T) {
	d := NewDetector()
	items := []interface{}{
		nil,
		map[string]interface{}{"id": "valid"},
		nil,
	}
	idMap := d.extractIDs(items)
	assert.NotNil(t, idMap)
}
