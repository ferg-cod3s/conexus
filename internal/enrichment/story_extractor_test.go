package enrichment

import (
	"testing"
)

func TestNewStoryExtractor(t *testing.T) {
	extractor := NewStoryExtractor()
	if extractor == nil {
		t.Fatal("Expected non-nil extractor")
	}
}

func TestExtractStoryReferences(t *testing.T) {
	extractor := NewStoryExtractor()

	tests := []struct {
		name     string
		content  string
		expected map[string][]string
	}{
		{
			name:     "empty content",
			content:  "",
			expected: map[string][]string{},
		},
		{
			name:     "no references",
			content:  "This is just a regular commit message",
			expected: map[string][]string{},
		},
		{
			name:    "issue references",
			content: "Fixes #123 and PROJ-456, also JIRA-789",
			expected: map[string][]string{
				"issues": {"123", "456", "789"},
				"prs":    {"123"}, // #123 also matches PR pattern
			},
		},
		{
			name:    "PR references",
			content: "Merges pull/42 and #99",
			expected: map[string][]string{
				"prs":    {"42", "99"},
				"issues": {"99"}, // #99 also matches issue pattern
			},
		},
		{
			name:    "branch references",
			content: "From feature/PROJ-123 and bugfix/JIRA-456",
			expected: map[string][]string{
				"branches": {"PROJ-123", "JIRA-456"},
				"issues":   {"123", "456"}, // PROJ-123 and JIRA-456 also match issue pattern
			},
		},
		{
			name:    "mixed references",
			content: "Fixes #123 from feature/PROJ-456, merges pull/789",
			expected: map[string][]string{
				"issues":   {"123", "456"}, // #123 and PROJ-456 match issue pattern
				"prs":      {"123", "789"}, // #123 and pull/789 match PR pattern
				"branches": {"PROJ-456"},
			},
		},
		{
			name:    "duplicate references",
			content: "Fixes #123 and #123 again",
			expected: map[string][]string{
				"issues": {"123", "123"}, // #123 matches issue pattern twice
				"prs":    {"123", "123"}, // #123 also matches PR pattern twice
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := extractor.ExtractStoryReferences(test.content)

			// Check that all expected keys are present
			for expectedKey := range test.expected {
				if _, exists := result[expectedKey]; !exists {
					t.Errorf("Expected key %s not found in result", expectedKey)
				}
			}

			// Check that values match
			for key, expectedValues := range test.expected {
				if resultValues, exists := result[key]; exists {
					if len(resultValues) != len(expectedValues) {
						t.Errorf("Key %s: expected %d values, got %d", key, len(expectedValues), len(resultValues))
					}
					for i, expected := range expectedValues {
						if i < len(resultValues) && resultValues[i] != expected {
							t.Errorf("Key %s: expected %v, got %v", key, expectedValues, resultValues)
						}
					}
				}
			}

			// Only check for unexpected keys if they have values (empty slices are ok)
			for resultKey, resultValues := range result {
				if _, exists := test.expected[resultKey]; !exists && len(resultValues) > 0 {
					t.Errorf("Unexpected key %s with values %v found in result", resultKey, resultValues)
				}
			}
		})
	}
}
