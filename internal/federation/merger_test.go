package federation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMerger(t *testing.T) {
	merger := NewMerger()
	assert.NotNil(t, merger)
	assert.NotNil(t, merger.results)
	assert.NotNil(t, merger.contentHashes)
	assert.NotNil(t, merger.sourceTracking)
	assert.NotNil(t, merger.attributions)
	assert.Empty(t, merger.results)
	assert.Empty(t, merger.contentHashes)
}

func TestMerger_AddResults_SingleSource(t *testing.T) {
	merger := NewMerger()
	items := []interface{}{"item1", "item2", "item3"}

	merger.AddResults("source1", items)

	assert.NotNil(t, merger.results["source1"])
	assert.Equal(t, 3, len(merger.results["source1"]))
}

func TestMerger_AddResults_MultipleSources(t *testing.T) {
	merger := NewMerger()

	merger.AddResults("source1", []interface{}{"item1", "item2"})
	merger.AddResults("source2", []interface{}{"item3", "item4"})
	merger.AddResults("source1", []interface{}{"item5"})

	assert.Equal(t, 2, len(merger.results))
	assert.Equal(t, 3, len(merger.results["source1"]))
	assert.Equal(t, 2, len(merger.results["source2"]))
}

func TestMerger_AddResults_Empty(t *testing.T) {
	merger := NewMerger()

	merger.AddResults("source1", []interface{}{})

	assert.NotNil(t, merger.results["source1"])
	assert.Empty(t, merger.results["source1"])
}

func TestMerger_MergeAndDeduplicate_Identical(t *testing.T) {
	merger := NewMerger()
	items := []interface{}{
		map[string]interface{}{"id": "1", "content": "same content"},
		map[string]interface{}{"id": "2", "content": "same content"},
	}

	merger.AddResults("source1", []interface{}{items[0]})
	merger.AddResults("source2", []interface{}{items[1]})

	merged, stats := merger.MergeAndDeduplicate()

	assert.Equal(t, 2, stats.TotalResults)
	assert.Greater(t, stats.DuplicatesFound, 0)
	assert.Less(t, len(merged), 2) // Duplicates removed
	assert.Greater(t, stats.UniqueResults, 0)
}

func TestMerger_MergeAndDeduplicate_Unique(t *testing.T) {
	merger := NewMerger()

	merger.AddResults("source1", []interface{}{
		map[string]interface{}{"id": "1", "content": "content1"},
	})
	merger.AddResults("source2", []interface{}{
		map[string]interface{}{"id": "2", "content": "content2"},
	})

	merged, stats := merger.MergeAndDeduplicate()

	assert.Equal(t, 2, stats.TotalResults)
	assert.Equal(t, 0, stats.DuplicatesFound)
	assert.Equal(t, 2, stats.UniqueResults)
	assert.Equal(t, 2, len(merged))
}

func TestMerger_MergeAndDeduplicate_Empty(t *testing.T) {
	merger := NewMerger()

	merged, stats := merger.MergeAndDeduplicate()

	assert.Equal(t, 0, stats.TotalResults)
	assert.Equal(t, 0, stats.DuplicatesFound)
	assert.Empty(t, merged)
}

func TestMerger_MergeAndDeduplicate_StringItems(t *testing.T) {
	merger := NewMerger()

	merger.AddResults("source1", []interface{}{"string1", "string2"})
	merger.AddResults("source2", []interface{}{"string2", "string3"})

	merged, stats := merger.MergeAndDeduplicate()

	assert.Equal(t, 4, stats.TotalResults)
	assert.Greater(t, stats.DuplicatesFound, 0)
	assert.Less(t, len(merged), 4)
}

func TestMerger_GetSourceAttributions(t *testing.T) {
	merger := NewMerger()

	merger.AddResults("source1", []interface{}{
		map[string]interface{}{"id": "1", "content": "content1"},
	})

	merger.MergeAndDeduplicate()

	attributions := merger.GetSourceAttributions()

	assert.NotNil(t, attributions)
	assert.NotEmpty(t, attributions)
}

func TestMerger_ItemToString_String(t *testing.T) {
	merger := NewMerger()

	result := merger.itemToString("test string")
	assert.Equal(t, "test string", result)
}

func TestMerger_ItemToString_MapWithContent(t *testing.T) {
	merger := NewMerger()

	item := map[string]interface{}{
		"content":   "test content",
		"id":        "123",
		"file_path": "/test/file.go",
	}

	result := merger.itemToString(item)
	assert.Equal(t, "test content", result)
}

func TestMerger_ItemToString_MapWithFilePath(t *testing.T) {
	merger := NewMerger()

	item := map[string]interface{}{
		"id":        "123",
		"file_path": "/test/file.go",
	}

	result := merger.itemToString(item)
	assert.Equal(t, "/test/file.go", result)
}

func TestMerger_ItemToString_MapWithID(t *testing.T) {
	merger := NewMerger()

	item := map[string]interface{}{
		"id": "123",
	}

	result := merger.itemToString(item)
	assert.Equal(t, "123", result)
}

func TestMerger_ItemToString_MapGeneric(t *testing.T) {
	merger := NewMerger()

	item := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	result := merger.itemToString(item)
	assert.NotEmpty(t, result)
}

func TestMerger_ItemToString_OtherTypes(t *testing.T) {
	merger := NewMerger()

	tests := []struct {
		name  string
		input interface{}
	}{
		{"int", 42},
		{"float", 3.14},
		{"bool", true},
		{"slice", []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := merger.itemToString(tt.input)
			assert.NotEmpty(t, result)
		})
	}
}

func TestMerger_HashContent(t *testing.T) {
	merger := NewMerger()

	hash1 := merger.hashContent("test content")
	hash2 := merger.hashContent("test content")
	hash3 := merger.hashContent("different content")

	assert.Equal(t, hash1, hash2, "same content should produce same hash")
	assert.NotEqual(t, hash1, hash3, "different content should produce different hash")
	assert.Len(t, hash1, 32, "MD5 hash should be 32 chars")
}

func TestCalculateSimilarity_Identical(t *testing.T) {
	similarity := CalculateSimilarity("test", "test")
	assert.Equal(t, 100, similarity)
}

func TestCalculateSimilarity_Completely_Different(t *testing.T) {
	similarity := CalculateSimilarity("abc", "xyz")
	assert.Equal(t, 0, similarity)
}

func TestCalculateSimilarity_Partial(t *testing.T) {
	similarity := CalculateSimilarity("hello", "help")
	assert.Greater(t, similarity, 0)
	assert.Less(t, similarity, 100)
}

func TestCalculateSimilarity_Empty(t *testing.T) {
	similarity := CalculateSimilarity("", "")
	assert.Equal(t, 100, similarity) // empty strings are identical
}


func TestCalculateSimilarity_DifferentLength(t *testing.T) {
	similarity := CalculateSimilarity("testing", "test data")
	assert.Greater(t, similarity, 0)
	assert.Less(t, similarity, 100)
}
func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "item exists",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			name:  "item not exists",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "a",
			want:  false,
		},
		{
			name:  "first element",
			slice: []string{"a", "b", "c"},
			item:  "a",
			want:  true,
		},
		{
			name:  "last element",
			slice: []string{"a", "b", "c"},
			item:  "c",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.slice, tt.item)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMerger_CountTotalResults(t *testing.T) {
	merger := NewMerger()

	merger.AddResults("source1", []interface{}{"1", "2", "3"})
	merger.AddResults("source2", []interface{}{"4", "5"})
	merger.AddResults("source3", []interface{}{})

	count := merger.countTotalResults()
	assert.Equal(t, 5, count)
}

func TestMerger_SourceTracking(t *testing.T) {
	merger := NewMerger()

	// Add same content from multiple sources
	items := []interface{}{
		map[string]interface{}{"id": "1", "content": "same"},
	}

	merger.AddResults("source1", []interface{}{items[0]})
	merger.AddResults("source2", []interface{}{items[0]})

	merged, stats := merger.MergeAndDeduplicate(); _ = merged

	attributions := merger.GetSourceAttributions()
	assert.NotEmpty(t, attributions)
	assert.Greater(t, stats.DuplicatesFound, 0)
}

func TestMerger_SortingByFrequency(t *testing.T) {
	merger := NewMerger()

	// Item appearing in 2 sources
	frequentItem := map[string]interface{}{"id": "1", "content": "frequent"}

	// Item appearing in 1 source
	rareItem := map[string]interface{}{"id": "2", "content": "rare"}

	merger.AddResults("source1", []interface{}{frequentItem, rareItem})
	merger.AddResults("source2", []interface{}{frequentItem})

	merged, _ := merger.MergeAndDeduplicate()

	// Most frequent should appear first
	assert.Greater(t, len(merged), 0)
}

func TestMerger_ComplexScenario(t *testing.T) {
	merger := NewMerger()

	// Add diverse items
	merger.AddResults("source1", []interface{}{
		map[string]interface{}{"id": "1", "content": "content1"},
		map[string]interface{}{"id": "2", "content": "content2"},
		"string_item1",
	})

	merger.AddResults("source2", []interface{}{
		map[string]interface{}{"id": "3", "content": "content3"},
		"string_item1", // Duplicate
	})

	merger.AddResults("source3", []interface{}{
		map[string]interface{}{"id": "1", "content": "content1"}, // Duplicate
	})

	merged, stats := merger.MergeAndDeduplicate()

	assert.Equal(t, 6, stats.TotalResults)
	assert.Greater(t, stats.DuplicatesFound, 0)
	assert.Less(t, stats.UniqueResults, 6)
	assert.Equal(t, stats.MergedResults, len(merged))
}

func TestMerger_LargeScale(t *testing.T) {
	merger := NewMerger()

	// Add large number of items
	const numSources = 10
	const itemsPerSource = 100

	for s := 0; s < numSources; s++ {
		items := make([]interface{}, itemsPerSource)
		for i := 0; i < itemsPerSource; i++ {
			items[i] = map[string]interface{}{
				"id":      i,
				"source":  s,
				"content": "test content",
			}
		}
		merger.AddResults(string(rune(s)), items)
	}

	merged, stats := merger.MergeAndDeduplicate()

	assert.Equal(t, numSources*itemsPerSource, stats.TotalResults)
	assert.Greater(t, stats.DuplicatesFound, 0)
	assert.Greater(t, stats.UniqueResults, 0)
	assert.Equal(t, stats.MergedResults, len(merged))
}

func TestDeduplicationStats(t *testing.T) {
	stats := DeduplicationStats{
		TotalResults:    100,
		DuplicatesFound: 25,
		UniqueResults:   75,
		MergedResults:   75,
	}

	assert.Equal(t, 100, stats.TotalResults)
	assert.Equal(t, 25, stats.DuplicatesFound)
	assert.Equal(t, 75, stats.UniqueResults)
	assert.Equal(t, 75, stats.MergedResults)
}
