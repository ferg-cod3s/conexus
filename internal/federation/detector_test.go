package federation

import (
	"testing"

	"github.com/ferg-cod3s/conexus/internal/schema"
	"github.com/stretchr/testify/assert"
)


func TestDetector_DetectRelationships(t *testing.T) {
	tests := []struct {
		name              string
		results           []schema.SearchResultItem
		expectedRelations int
		expectedTypes     []string
	}{
		{
			name: "same ticket relationship",
			results: []schema.SearchResultItem{
				{
					ID: "1", SourceType: "github", Content: "Issue description",
					Metadata: map[string]interface{}{"ticket_id": "PROJ-123"},
				},
				{
					ID: "2", SourceType: "slack", Content: "Discussion about issue",
					Metadata: map[string]interface{}{"ticket_id": "PROJ-123"},
				},
			},
			expectedRelations: 1,
			expectedTypes:     []string{RelationTypeSameTicket},
		},
		{
			name: "same file relationship",
			results: []schema.SearchResultItem{
				{
					ID: "1", SourceType: "file", Content: "Function definition",
					Metadata: map[string]interface{}{"file_path": "/src/main.go"},
				},
				{
					ID: "2", SourceType: "github", Content: "PR comment",
					Metadata: map[string]interface{}{"file_path": "/src/main.go"},
				},
			},
			expectedRelations: 1,
			expectedTypes:     []string{RelationTypeSameFile},
		},
		{
			name: "test file relationship",
			results: []schema.SearchResultItem{
				{
					ID: "1", SourceType: "file", Content: "Test function",
					Metadata: map[string]interface{}{"file_path": "/src/main_test.go"},
				},
				{
					ID: "2", SourceType: "file", Content: "Implementation",
					Metadata: map[string]interface{}{"file_path": "/src/main.go"},
				},
			},
			expectedRelations: 1,
			expectedTypes:     []string{RelationTypeTestFile},
		},
		{
			name: "documentation relationship",
			results: []schema.SearchResultItem{
				{
					ID: "1", SourceType: "file", Content: "README content",
					Metadata: map[string]interface{}{"file_path": "/docs/README.md"},
				},
				{
					ID: "2", SourceType: "file", Content: "Code implementation",
					Metadata: map[string]interface{}{"file_path": "/src/main.go"},
				},
			},
			expectedRelations: 1,
			expectedTypes:     []string{RelationTypeDocumentation},
		},
		{
			name: "content similarity relationship",
			results: []schema.SearchResultItem{
				{
					ID: "1", SourceType: "github", Content: "This is a very specific error message that should match",
				},
				{
					ID: "2", SourceType: "slack", Content: "This is a very specific error message that should match",
				},
			},
			expectedRelations: 1,
			expectedTypes:     []string{RelationTypeSameEntity},
		},
		{
			name: "no relationships",
			results: []schema.SearchResultItem{
				{
					ID: "1", SourceType: "file", Content: "Unrelated content 1",
					Metadata: map[string]interface{}{"file_path": "/file1.go"},
				},
				{
					ID: "2", SourceType: "github", Content: "Unrelated content 2",
					Metadata: map[string]interface{}{"file_path": "/file2.go"},
				},
			},
			expectedRelations: 0,
			expectedTypes:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewDetector()
			relationships := detector.DetectRelationships(tt.results)

			assert.Len(t, relationships, tt.expectedRelations)

			if len(tt.expectedTypes) > 0 {
				var actualTypes []string
				for _, rel := range relationships {
					actualTypes = append(actualTypes, rel.Type)
				}
				assert.Equal(t, tt.expectedTypes, actualTypes)
			}
		})
	}
}

func TestDetector_detectRelationship(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name        string
		item1       schema.SearchResultItem
		item2       schema.SearchResultItem
		expectRel   bool
		expectedType string
	}{
		{
			name: "same ticket",
			item1: schema.SearchResultItem{
				ID: "1", SourceType: "github",
				Metadata: map[string]interface{}{"ticket_id": "PROJ-123"},
			},
			item2: schema.SearchResultItem{
				ID: "2", SourceType: "slack",
				Metadata: map[string]interface{}{"ticket_id": "PROJ-123"},
			},
			expectRel:   true,
			expectedType: RelationTypeSameTicket,
		},
		{
			name: "same file",
			item1: schema.SearchResultItem{
				ID: "1", SourceType: "file",
				Metadata: map[string]interface{}{"file_path": "/main.go"},
			},
			item2: schema.SearchResultItem{
				ID: "2", SourceType: "github",
				Metadata: map[string]interface{}{"file_path": "/main.go"},
			},
			expectRel:   true,
			expectedType: RelationTypeSameFile,
		},
		{
			name: "no relationship",
			item1: schema.SearchResultItem{
				ID: "1", SourceType: "file",
				Metadata: map[string]interface{}{"file_path": "/file1.go"},
			},
			item2: schema.SearchResultItem{
				ID: "2", SourceType: "github",
				Metadata: map[string]interface{}{"file_path": "/file2.go"},
			},
			expectRel: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := detector.detectRelationship(tt.item1, tt.item2)

			if tt.expectRel {
				assert.NotNil(t, rel)
				assert.Equal(t, tt.expectedType, rel.Type)
				assert.Equal(t, tt.item1.ID, rel.Item1ID)
				assert.Equal(t, tt.item2.ID, rel.Item2ID)
			} else {
				assert.Nil(t, rel)
			}
		})
	}
}

func TestDetector_contentSimilarity(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float32
	}{
		{
			name:     "identical text",
			text1:    "this is a test",
			text2:    "this is a test",
			expected: 1.0,
		},
		{
			name:     "partial overlap",
			text1:    "this is a test",
			text2:    "this is another test",
			expected: 0.6, // 3 out of 5 words overlap
		},
		{
			name:     "no overlap",
			text1:    "hello world",
			text2:    "goodbye universe",
			expected: 0.0,
		},
		{
			name:     "empty text",
			text1:    "",
			text2:    "some text",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := detector.contentSimilarity(tt.text1, tt.text2)
			assert.InDelta(t, tt.expected, similarity, 0.1)
		})
	}
}

func TestDetector_isTestFileRelationship(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name   string
		path1  string
		path2  string
		expected bool
	}{
		{
			name:     "go test file",
			path1:    "/main.go",
			path2:    "/main_test.go",
			expected: true,
		},
		{
			name:     "javascript test file",
			path1:    "/component.js",
			path2:    "/component.test.js",
			expected: true,
		},
		{
			name:     "spec file",
			path1:    "/service.ts",
			path2:    "/service.spec.ts",
			expected: true,
		},
		{
			name:     "no test relationship",
			path1:    "/main.go",
			path2:    "/utils.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isTestFileRelationship(tt.path1, tt.path2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetector_isDocumentationRelationship(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name   string
		path1  string
		path2  string
		expected bool
	}{
		{
			name:     "markdown file",
			path1:    "/README.md",
			path2:    "/main.go",
			expected: true,
		},
		{
			name:     "docs directory",
			path1:    "/docs/guide.txt",
			path2:    "/src/main.go",
			expected: true,
		},
		{
			name:     "readme file",
			path1:    "/readme.rst",
			path2:    "/lib/code.py",
			expected: true,
		},
		{
			name:     "no documentation",
			path1:    "/main.go",
			path2:    "/utils.go",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.isDocumentationRelationship(tt.path1, tt.path2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewDetector(t *testing.T) {
	detector := NewDetector()
	assert.NotNil(t, detector)
}
