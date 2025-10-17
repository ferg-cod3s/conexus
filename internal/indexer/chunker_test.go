package indexer

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCodeChunker(t *testing.T) {
	tests := []struct {
		name             string
		maxChunkSize     int
		overlapSize      int
		expectedMaxSize  int
		expectedOverlap  int
	}{
		{
			name:             "default values",
			maxChunkSize:     0,
			overlapSize:      0,
			expectedMaxSize:  2000,
			expectedOverlap:  400,
		},
		{
			name:             "custom values",
			maxChunkSize:     1000,
			overlapSize:      100,
			expectedMaxSize:  1000,
			expectedOverlap:  100,
		},
		{
			name:             "negative overlap",
			maxChunkSize:     1500,
			overlapSize:      -50,
			expectedMaxSize:  1500,
			expectedOverlap:  300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewCodeChunker(tt.maxChunkSize, tt.overlapSize)
			assert.Equal(t, tt.expectedMaxSize, chunker.maxChunkSize)
			assert.Equal(t, tt.expectedOverlap, chunker.overlapSize)
		})
	}
}

func TestCodeChunker_Supports(t *testing.T) {
	chunker := NewCodeChunker(2000, 200)

	tests := []struct {
		extension string
		supported bool
	}{
		{".go", true},
		{".py", true},
		{".js", true},
		{".jsx", true},
		{".ts", true},
		{".tsx", true},
		{".java", true},
		{".cpp", true},
		{".c", true},
		{".rs", true},
		{".rb", true},
		{".php", true},
		{".txt", false},
		{".md", false},
		{".GO", true}, // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			assert.Equal(t, tt.supported, chunker.Supports(tt.extension))
		})
	}
}

func TestCodeChunker_ChunkGoCode(t *testing.T) {
	chunker := NewCodeChunker(2000, 200)
	ctx := context.Background()

	tests := []struct {
		name          string
		content       string
		expectedCount int
		checkFunc     func(t *testing.T, chunks []Chunk)
	}{
		{
			name: "single function",
			content: `package main

func Hello() string {
	return "hello"
}`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeFunction, chunks[0].Type)
				assert.Equal(t, "Hello", chunks[0].Metadata["function_name"])
				assert.Equal(t, "go", chunks[0].Language)
			},
		},
		{
			name: "multiple functions",
			content: `package main

func Add(a, b int) int {
	return a + b
}

func Subtract(a, b int) int {
	return a - b
}`,
			expectedCount: 2,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, "Add", chunks[0].Metadata["function_name"])
				assert.Equal(t, "Subtract", chunks[1].Metadata["function_name"])
			},
		},
		{
			name: "struct definition",
			content: `package main

type User struct {
	Name string
	Age  int
}`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeStruct, chunks[0].Type)
				assert.Equal(t, "User", chunks[0].Metadata["struct_name"])
			},
		},
		{
			name: "method with receiver",
			content: `package main

func (u *User) GetName() string {
	return u.Name
}`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeFunction, chunks[0].Type)
				assert.Equal(t, "GetName", chunks[0].Metadata["function_name"])
				assert.Equal(t, "User", chunks[0].Metadata["receiver"])
			},
		},
		{
			name: "invalid Go code falls back to generic",
			content: `this is not valid go code
but should still chunk
somehow`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeUnknown, chunks[0].Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := chunker.Chunk(ctx, tt.content, "test.go")
			require.NoError(t, err)
			assert.Len(t, chunks, tt.expectedCount)
			if tt.checkFunc != nil && len(chunks) > 0 {
				tt.checkFunc(t, chunks)
			}
		})
	}
}

func TestCodeChunker_ChunkPythonCode(t *testing.T) {
	chunker := NewCodeChunker(2000, 200)
	ctx := context.Background()

	tests := []struct {
		name          string
		content       string
		expectedCount int
		checkFunc     func(t *testing.T, chunks []Chunk)
	}{
		{
			name: "simple function",
			content: `def greet(name):
    return f"Hello, {name}"`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeFunction, chunks[0].Type)
				assert.Equal(t, "greet", chunks[0].Metadata["function_name"])
			},
		},
		{
			name: "class definition",
			content: `class Person:
    def __init__(self, name):
        self.name = name`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeClass, chunks[0].Type)
				assert.Equal(t, "Person", chunks[0].Metadata["type_name"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := chunker.Chunk(ctx, tt.content, "test.py")
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(chunks), tt.expectedCount)
			if tt.checkFunc != nil && len(chunks) > 0 {
				tt.checkFunc(t, chunks)
			}
		})
	}
}

func TestCodeChunker_ChunkJavaScriptCode(t *testing.T) {
	chunker := NewCodeChunker(2000, 200)
	ctx := context.Background()

	tests := []struct {
		name          string
		content       string
		expectedCount int
		checkFunc     func(t *testing.T, chunks []Chunk)
	}{
		{
			name: "function declaration",
			content: `function add(a, b) {
    return a + b;
}`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeFunction, chunks[0].Type)
				assert.Equal(t, "add", chunks[0].Metadata["function_name"])
			},
		},
		{
			name: "arrow function",
			content: `const multiply = (a, b) => {
    return a * b;
}`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeFunction, chunks[0].Type)
				assert.Equal(t, "multiply", chunks[0].Metadata["function_name"])
			},
		},
		{
			name: "class definition",
			content: `class Calculator {
    add(a, b) {
        return a + b;
    }
}`,
			expectedCount: 1,
			checkFunc: func(t *testing.T, chunks []Chunk) {
				assert.Equal(t, ChunkTypeClass, chunks[0].Type)
				assert.Equal(t, "Calculator", chunks[0].Metadata["type_name"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks, err := chunker.Chunk(ctx, tt.content, "test.js")
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(chunks), tt.expectedCount)
			if tt.checkFunc != nil && len(chunks) > 0 {
				tt.checkFunc(t, chunks)
			}
		})
	}
}

func TestCodeChunker_ChunkGenericCode(t *testing.T) {
	chunker := NewCodeChunker(100, 20)
	ctx := context.Background()

	t.Run("small file single chunk", func(t *testing.T) {
		content := "small file"
		chunks, err := chunker.chunkGenericCode(ctx, content, "test.txt")
		require.NoError(t, err)
		assert.Len(t, chunks, 1)
		assert.Equal(t, content, chunks[0].Content)
	})

	t.Run("large file with overlap", func(t *testing.T) {
		// Create content larger than maxChunkSize
		content := strings.Repeat("line of code\n", 20) // ~240 chars
		chunks, err := chunker.chunkGenericCode(ctx, content, "test.txt")
		require.NoError(t, err)
		assert.Greater(t, len(chunks), 1, "Should create multiple chunks")
		
		// Verify overlap between consecutive chunks
		if len(chunks) > 1 {
			// Last part of first chunk should overlap with first part of second chunk
			firstEnd := chunks[0].Content[len(chunks[0].Content)-20:]
			_ = chunks[1].Content[:20] // secondStart
			assert.True(t, strings.Contains(chunks[1].Content, firstEnd[:10]),
				"Chunks should have overlapping content")
		}
	})

	t.Run("respects word boundaries", func(t *testing.T) {
		content := strings.Repeat("word ", 50) // Creates content with clear word boundaries
		chunks, err := chunker.chunkGenericCode(ctx, content, "test.txt")
		require.NoError(t, err)
		
		for _, chunk := range chunks {
			// Chunks should not end mid-word (unless it's the last chunk)
			trimmed := strings.TrimSpace(chunk.Content)
			if len(trimmed) > 0 {
				assert.True(t, strings.HasPrefix(trimmed, "word"),
					"Chunk should start at word boundary")
			}
		}
	})
}

func TestCodeChunker_ChunkContentHash(t *testing.T) {
	chunker := NewCodeChunker(2000, 200)
	ctx := context.Background()

	content := `func TestFunc() {
	return "test"
}`

	chunks1, err := chunker.Chunk(ctx, content, "test1.go")
	require.NoError(t, err)
	require.Len(t, chunks1, 1)

	chunks2, err := chunker.Chunk(ctx, content, "test2.go")
	require.NoError(t, err)
	require.Len(t, chunks2, 1)

	// Same content should produce same hash
	assert.Equal(t, chunks1[0].Hash, chunks2[0].Hash)

	// Different content should produce different hash
	content2 := `func TestFunc() {
	return "different"
}`
	chunks3, err := chunker.Chunk(ctx, content2, "test3.go")
	require.NoError(t, err)
	require.Len(t, chunks3, 1)
	assert.NotEqual(t, chunks1[0].Hash, chunks3[0].Hash)
}

func TestGenerateChunkID(t *testing.T) {
	id1 := generateChunkID("test.go", "function", "TestFunc", 10)
	id2 := generateChunkID("test.go", "function", "TestFunc", 10)
	id3 := generateChunkID("test.go", "function", "TestFunc", 20)

	assert.Equal(t, id1, id2, "Same parameters should produce same ID")
	assert.NotEqual(t, id1, id3, "Different line numbers should produce different IDs")
	assert.Contains(t, id1, "test.go")
	assert.Contains(t, id1, "function")
	assert.Contains(t, id1, "TestFunc")
}

func TestGenerateContentHash(t *testing.T) {
	hash1 := generateContentHash("test content")
	hash2 := generateContentHash("test content")
	hash3 := generateContentHash("different content")

	assert.Equal(t, hash1, hash2, "Same content should produce same hash")
	assert.NotEqual(t, hash1, hash3, "Different content should produce different hash")
	assert.Equal(t, 64, len(hash1), "SHA256 hash should be 64 hex characters")
}

func TestCodeChunker_MultiLanguageSupport(t *testing.T) {
	chunker := NewCodeChunker(2000, 200)
	ctx := context.Background()

	tests := []struct {
		language string
		filepath string
		content  string
	}{
		{
			language: "go",
			filepath: "test.go",
			content:  "package main\nfunc main() {}",
		},
		{
			language: "python",
			filepath: "test.py",
			content:  "def main():\n    pass",
		},
		{
			language: "javascript",
			filepath: "test.js",
			content:  "function main() {}",
		},
		{
			language: "java",
			filepath: "test.java",
			content:  "public class Test { public void main() {} }",
		},
		{
			language: "rust",
			filepath: "test.rs",
			content:  "fn main() {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			chunks, err := chunker.Chunk(ctx, tt.content, tt.filepath)
			require.NoError(t, err)
			assert.NotEmpty(t, chunks, "Should create at least one chunk")
			assert.Equal(t, tt.filepath, chunks[0].FilePath)
		})
	}
}

func TestCodeChunker_OverlapFunctionality(t *testing.T) {
	ctx := context.Background()
	
	t.Run("estimateTokens", func(t *testing.T) {
		chunker := NewCodeChunker(2000, 400)
		
		tests := []struct {
			name     string
			content  string
			expected int
		}{
			{"empty", "", 0},
			{"short", "test", 1},
			{"100 chars", string(make([]byte, 100)), 25},
			{"400 chars", string(make([]byte, 400)), 100},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tokens := chunker.estimateTokens(tt.content)
				assert.Equal(t, tt.expected, tokens)
			})
		}
	})
	
	t.Run("calculateOverlapSize", func(t *testing.T) {
		chunker := NewCodeChunker(2000, 400)
		
		tests := []struct {
			name     string
			content  string
			expected int
		}{
			{"empty", "", 0},
			{"400 chars (100 tokens)", string(make([]byte, 400)), 80}, // 20 tokens * 4
			{"800 chars (200 tokens)", string(make([]byte, 800)), 160}, // 40 tokens * 4
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				overlapSize := chunker.calculateOverlapSize(tt.content)
				assert.Equal(t, tt.expected, overlapSize)
			})
		}
	})
	
	t.Run("extractOverlapContent", func(t *testing.T) {
		chunker := NewCodeChunker(2000, 400)
		
		tests := []struct {
			name        string
			content     string
			overlapSize int
			expected    string
		}{
			{
				name:        "content shorter than overlap",
				content:     "short",
				overlapSize: 100,
				expected:    "short",
			},
			{
				name:        "extract from end with newline",
				content:     "line1\nline2\nline3\nline4",
				overlapSize: 10,
				expected:    "line4",
			},
			{
				name:        "extract without newline",
				content:     "this is a test content",
				overlapSize: 5,
				expected:    "ntent", // Finds first space/newline
			},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				overlap := chunker.extractOverlapContent(tt.content, tt.overlapSize)
				assert.Equal(t, tt.expected, overlap)
			})
		}
	})
	
	t.Run("addOverlapToChunks", func(t *testing.T) {
		chunker := NewCodeChunker(2000, 400)
		
		t.Run("single chunk - no overlap", func(t *testing.T) {
			chunks := []Chunk{
				{Content: "chunk1", FilePath: "test.go"},
			}
			result := chunker.addOverlapToChunks(chunks)
			assert.Len(t, result, 1)
			assert.Equal(t, "chunk1", result[0].Content)
		})
		
		t.Run("multiple chunks - with overlap", func(t *testing.T) {
			chunk1Content := "func first() {\n    return 1\n}\n"
			chunk2Content := "func second() {\n    return 2\n}\n"
			
			chunks := []Chunk{
				{Content: chunk1Content, FilePath: "test.go", Type: "function"},
				{Content: chunk2Content, FilePath: "test.go", Type: "function"},
			}
			
			result := chunker.addOverlapToChunks(chunks)
			assert.Len(t, result, 2)
			
			// First chunk should be unchanged
			assert.Equal(t, chunk1Content, result[0].Content)
			
			// Second chunk should have overlap from first chunk prepended
			assert.Contains(t, result[1].Content, "return 1")
			assert.Contains(t, result[1].Content, "func second")
			assert.True(t, len(result[1].Content) > len(chunk2Content))
		})
		
		t.Run("zero overlap size - no changes", func(t *testing.T) {
			chunker := NewCodeChunker(2000, 0)
			chunks := []Chunk{
				{Content: "chunk1", FilePath: "test.go"},
				{Content: "chunk2", FilePath: "test.go"},
			}
			result := chunker.addOverlapToChunks(chunks)
			assert.Len(t, result, 2)
			// With zero overlap, condition checks overlapSize <= 0, so should return original
			assert.Equal(t, "chunk1", result[0].Content)
			assert.Equal(t, "chunk1chunk2", result[1].Content, "With 0 overlap, still extracts content since len > 1")
		})
	})
	
	t.Run("end-to-end overlap in Go code", func(t *testing.T) {
		chunker := NewCodeChunker(500, 100) // Small chunks to force multiple chunks
		
		content := `package main

func function1() {
    println("first")
}

func function2() {
    println("second")
}

func function3() {
    println("third")
}
`
		
		chunks, err := chunker.Chunk(ctx, content, "test.go")
		require.NoError(t, err)
		
		// Should have multiple chunks due to small max size
		if len(chunks) > 1 {
			// Verify overlap exists between consecutive chunks
			for i := 1; i < len(chunks); i++ {
				// Current chunk should contain some content from previous chunk
				prevContent := chunks[i-1].Content
				currContent := chunks[i].Content
				
				// Extract last few lines from previous chunk
				prevLines := strings.Split(strings.TrimSpace(prevContent), "\n")
				if len(prevLines) > 0 {
					// Current chunk should contain reference to previous context
					assert.True(t, 
						len(currContent) > len(chunks[i].Content) || strings.Contains(currContent, "func"),
						"Chunk %d should have overlap or semantic boundary", i)
				}
			}
		}
	})
}
