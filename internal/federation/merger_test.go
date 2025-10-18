package federation

import (
	"testing"

	"github.com/ferg-cod3s/conexus/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestMerger_Merge(t *testing.T) {
	tests := []struct {
		name             string
		connectorResults []ConnectorResult
		expectedCount    int
		expectedOrder    []string // Expected order of result IDs
	}{
		{
			name: "merge results from multiple connectors",
			connectorResults: []ConnectorResult{
				{
					ConnectorID:   "filesystem",
					ConnectorType: "filesystem",
					Results: []schema.SearchResultItem{
						{ID: "fs1", Content: "file content 1", Score: 0.9, SourceType: "file"},
						{ID: "fs2", Content: "file content 2", Score: 0.7, SourceType: "file"},
					},
				},
				{
					ConnectorID:   "github",
					ConnectorType: "github",
					Results: []schema.SearchResultItem{
						{ID: "gh1", Content: "github content 1", Score: 0.8, SourceType: "github"},
					},
				},
			},
			expectedCount: 3,
			expectedOrder: []string{"fs1", "gh1", "fs2"}, // Should be sorted by score
		},
		{
			name: "deduplicate identical content",
			connectorResults: []ConnectorResult{
				{
					ConnectorID:   "filesystem",
					ConnectorType: "filesystem",
					Results: []schema.SearchResultItem{
						{ID: "fs1", Content: "identical content", Score: 0.8, SourceType: "file"},
					},
				},
				{
					ConnectorID:   "github",
					ConnectorType: "github",
					Results: []schema.SearchResultItem{
						{ID: "gh1", Content: "identical content", Score: 0.9, SourceType: "github"},
					},
				},
			},
			expectedCount: 1, // Should deduplicate, keep higher score
			expectedOrder: []string{"gh1"},
		},
		{
			name: "empty results",
			connectorResults: []ConnectorResult{
				{
					ConnectorID:   "filesystem",
					ConnectorType: "filesystem",
					Results:       []schema.SearchResultItem{},
				},
			},
			expectedCount: 0,
			expectedOrder: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := NewMerger()
			results := merger.Merge(tt.connectorResults)

			assert.Len(t, results, tt.expectedCount)

			if len(tt.expectedOrder) > 0 {
				var actualOrder []string
				for _, result := range results {
					actualOrder = append(actualOrder, result.ID)
				}
				assert.Equal(t, tt.expectedOrder, actualOrder)
			}

			// Verify metadata was added
			for _, result := range results {
				assert.NotNil(t, result.Metadata)
				assert.Contains(t, result.Metadata, "connector_id")
				assert.Contains(t, result.Metadata, "connector_type")
			}
		})
	}
}

func TestMerger_deduplicate(t *testing.T) {
	tests := []struct {
		name          string
		input         []schema.SearchResultItem
		expectedCount int
		expectedIDs   []string
	}{
		{
			name: "no duplicates",
			input: []schema.SearchResultItem{
				{ID: "1", Content: "content 1", Score: 0.9},
				{ID: "2", Content: "content 2", Score: 0.8},
			},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name: "exact duplicates - keep higher score",
			input: []schema.SearchResultItem{
				{ID: "1", Content: "duplicate content", Score: 0.7},
				{ID: "2", Content: "duplicate content", Score: 0.9},
			},
			expectedCount: 1,
			expectedIDs:   []string{"2"}, // Higher score wins
		},
		{
			name: "similar content deduplicated",
			input: []schema.SearchResultItem{
				{ID: "1", Content: "this is a test", Score: 0.8},
				{ID: "2", Content: "this is a test", Score: 0.6},
			},
			expectedCount: 1,
			expectedIDs:   []string{"1"}, // First one wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := NewMerger()
			results := merger.deduplicate(tt.input)

			assert.Len(t, results, tt.expectedCount)

			var actualIDs []string
			for _, result := range results {
				actualIDs = append(actualIDs, result.ID)
			}
			assert.Equal(t, tt.expectedIDs, actualIDs)
		})
	}
}
