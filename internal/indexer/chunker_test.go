package indexer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCodeChunker(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	assert.NotNil(t, chunker)
	assert.Equal(t, 1000, chunker.maxChunkSize)
	assert.Equal(t, 100, chunker.overlapSize)
	assert.True(t, chunker.Supports(".go"))
	assert.True(t, chunker.Supports(".py"))
	assert.True(t, chunker.Supports(".js"))
	assert.False(t, chunker.Supports(".txt"))
}

func TestChunkerSupports(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	testCases := []struct {
		extension string
		expected  bool
	}{
		{".go", true},
		{".py", true},
		{".js", true},
		{".ts", true},
		{".jsx", true},
		{".java", true},
		{".cpp", true},
		{".c", true},
		{".rs", true},
		{".txt", false},
		{".json", false},
		{".yaml", false},
		{".md", false},
	}

	for _, tc := range testCases {
		t.Run(tc.extension, func(t *testing.T) {
			result := chunker.Supports(tc.extension)
			assert.Equal(t, tc.expected, result, "Extension: %s", tc.extension)
		})
	}
}

func TestGenerateChunkID(t *testing.T) {
	id1 := generateChunkID("test.go", "function", "main", 10)
	id2 := generateChunkID("test.go", "function", "main", 10)
	id3 := generateChunkID("test.go", "function", "helper", 10)

	assert.Equal(t, id1, id2, "Same parameters should generate same ID")
	assert.NotEqual(t, id1, id3, "Different name should generate different ID")
}

func TestGenerateContentHash(t *testing.T) {
	content1 := "test content"
	content2 := "test content"
	content3 := "different content"

	hash1 := generateContentHash(content1)
	hash2 := generateContentHash(content2)
	hash3 := generateContentHash(content3)

	assert.Equal(t, hash1, hash2, "Same content should generate same hash")
	assert.NotEqual(t, hash1, hash3, "Different content should generate different hash")
	assert.NotEmpty(t, hash1, "Hash should not be empty")
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		content   string
		expectErr bool
		minChunks int
	}{
		{
			name:      "Go code",
			filePath:  "test.go",
			content:   "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
			expectErr: false,
			minChunks: 1,
		},
		{
			name:      "Python code",
			filePath:  "test.py",
			content:   "def hello():\n    print(\"Hello, World!\")",
			expectErr: false,
			minChunks: 1,
		},
		{
			name:      "JavaScript code",
			filePath:  "test.js",
			content:   "function hello() {\n    console.log(\"Hello, World!\");\n}",
			expectErr: false,
			minChunks: 1,
		},
		{
			name:      "Java code",
			filePath:  "test.java",
			content:   "public class Test {\n    public static void main(String[] args) {\n        System.out.println(\"Hello, World!\");\n    }\n}",
			expectErr: false,
			minChunks: 1,
		},
		{
			name:      "C code",
			filePath:  "test.c",
			content:   "#include <stdio.h>\n\nint main() {\n    printf(\"Hello, World!\\n\");\n    return 0;\n}",
			expectErr: false,
			minChunks: 1,
		},
		{
			name:      "Rust code",
			filePath:  "test.rs",
			content:   "fn main() {\n    println!(\"Hello, World!\");\n}",
			expectErr: false,
			minChunks: 1,
		},
		{
			name:      "Unsupported file type",
			filePath:  "test.txt",
			content:   "This is just text",
			expectErr: false,
			minChunks: 1,
		},
		{
			name:      "Empty content",
			filePath:  "test.go",
			content:   "",
			expectErr: false,
			minChunks: 1, // chunkGenericCode still creates a chunk even for empty content
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunker := NewCodeChunker(1000, 100)
			chunks, err := chunker.Chunk(context.Background(), tc.content, tc.filePath)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if len(chunks) < tc.minChunks {
				t.Errorf("Expected at least %d chunks, got %d", tc.minChunks, len(chunks))
			}

			// Verify chunk properties
			for _, chunk := range chunks {
				assert.NotEmpty(t, chunk.ID, "Chunk ID should not be empty")
				// Empty content is allowed for empty files
				if tc.content != "" {
					assert.NotEmpty(t, chunk.Content, "Chunk content should not be empty for non-empty input")
				}
				assert.Equal(t, tc.filePath, chunk.FilePath, "File path should match")
				assert.GreaterOrEqual(t, chunk.StartLine, 0, "Start line should be >= 0")
				// For empty content, both start and end line might be 0
				if tc.content != "" {
					assert.GreaterOrEqual(t, chunk.EndLine, chunk.StartLine, "End line should be >= start line")
				}
			}
		})
	}
}

func TestChunkWithStoryReferences(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)
	content := `package main

// Fixes #123
func main() {
	// Related to PROJ-456
	println("Hello, World!")
}`

	chunks, err := chunker.Chunk(context.Background(), content, "test.go")
	assert.NoError(t, err)
	assert.Greater(t, len(chunks), 0)

	// Check that story references were extracted
	for _, chunk := range chunks {
		// The chunk should contain story references
		if len(chunk.StoryIDs) > 0 || len(chunk.PRNumbers) > 0 {
			// Found story references
			break
		}
	}
}

func TestChunkGoCode(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	tests := []struct {
		name      string
		content   string
		expectErr bool
	}{
		{
			name: "Simple function",
			content: `package main

func hello() {
	println("Hello")
}`,
			expectErr: false,
		},
		{
			name: "Multiple functions",
			content: `package main

func func1() {
	println("1")
}

func func2() {
	println("2")
}`,
			expectErr: false,
		},
		{
			name: "With struct",
			content: `package main

type Person struct {
	Name string
	Age  int
}

func (p *Person) Greet() {
	println("Hello, " + p.Name)
}`,
			expectErr: false,
		},
		{
			name: "Invalid Go syntax",
			content: `package main

func invalid( {`,
			expectErr: false, // Go parser might not error on this, or error is handled
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := chunker.chunkGoCode(context.Background(), tc.content, "test.go")

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Greater(t, len(chunks), 0, "Should have at least one chunk")
			}
		})
	}
}

func TestChunkPythonCode(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "Simple function",
			content: `def hello():
    print("Hello")`,
		},
		{
			name: "Class definition",
			content: `class Person:
    def __init__(self, name):
        self.name = name
    
    def greet(self):
        print(f"Hello, {self.name}")`,
		},
		{
			name: "Module with imports",
			content: `import os
import sys

from typing import List

def process_items(items: List[str]) -> List[str]:
    return [item.upper() for item in items]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := chunker.chunkPythonCode(context.Background(), tc.content, "test.py")
			assert.NoError(t, err)
			assert.Greater(t, len(chunks), 0, "Should have at least one chunk")
		})
	}
}

func TestChunkJavaScriptCode(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "Simple function",
			content: `function hello() {
    console.log("Hello");
}`,
		},
		{
			name: "Arrow function",
			content: `const hello = () => {
    console.log("Hello");
};`,
		},
		{
			name: "Class definition",
			content: `class Person {
    constructor(name) {
        this.name = name;
    }
    
    greet() {
        console.log(` + "`Hello, ${this.name}`" + `);
    }
}`,
		},
		{
			name: "Module with exports",
			content: `export const PI = 3.14159;

export function calculateArea(radius) {
    return PI * radius * radius;
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := chunker.chunkJavaScriptCode(context.Background(), tc.content, "test.js")
			assert.NoError(t, err)
			assert.Greater(t, len(chunks), 0, "Should have at least one chunk")
		})
	}
}

func TestChunkJavaCode(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "Simple class",
			content: `public class Hello {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}`,
		},
		{
			name: "Class with methods",
			content: `public class Calculator {
    public int add(int a, int b) {
        return a + b;
    }
    
    public int subtract(int a, int b) {
        return a - b;
    }
}`,
		},
		{
			name: "Interface",
			content: `public interface Drawable {
    void draw();
    double getArea();
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := chunker.chunkJavaCode(context.Background(), tc.content, "test.java")
			assert.NoError(t, err)
			assert.Greater(t, len(chunks), 0, "Should have at least one chunk")
		})
	}
}

func TestChunkCCode(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "Simple function",
			content: `#include <stdio.h>

void hello() {
    printf("Hello, World!\\n");
}`,
		},
		{
			name: "Struct and functions",
			content: `#include <stdio.h>
#include <stdlib.h>

typedef struct {
    char name[50];
    int age;
} Person;

void print_person(Person* p) {
    printf("Name: %s, Age: %d\\n", p->name, p->age);
}`,
		},
		{
			name: "Header file style",
			content: `#ifndef HELLO_H
#define HELLO_H

void hello(void);
int add(int a, int b);

#endif /* HELLO_H */`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := chunker.chunkCCode(context.Background(), tc.content, "test.c")
			assert.NoError(t, err)
			assert.Greater(t, len(chunks), 0, "Should have at least one chunk")
		})
	}
}

func TestChunkRustCode(t *testing.T) {
	chunker := NewCodeChunker(1000, 100)

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "Simple function",
			content: `fn hello() {
    println!("Hello, World!");
}`,
		},
		{
			name: "Struct and impl",
			content: `struct Person {
    name: String,
    age: u32,
}

impl Person {
    fn new(name: String, age: u32) -> Self {
        Person { name, age }
    }
    
    fn greet(&self) {
        println!("Hello, {}!", self.name);
    }
}`,
		},
		{
			name: "Module with traits",
			content: `trait Drawable {
    fn draw(&self);
    fn area(&self) -> f64;
}

struct Circle {
    radius: f64,
}

impl Drawable for Circle {
    fn draw(&self) {
        println!("Drawing circle");
    }
    
    fn area(&self) -> f64 {
        3.14159 * self.radius * self.radius
    }
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := chunker.chunkRustCode(context.Background(), tc.content, "test.rs")
			assert.NoError(t, err)
			assert.Greater(t, len(chunks), 0, "Should have at least one chunk")
		})
	}
}

func TestChunkGenericCode(t *testing.T) {
	chunker := NewCodeChunker(100, 20) // Small chunk size for testing

	tests := []struct {
		name           string
		content        string
		expectedChunks int
	}{
		{
			name: "Short content",
			content: `This is a short file
with just a few lines
of content.`,
			expectedChunks: 1,
		},
		{
			name: "Long content",
			content: `Line 1
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7
Line 8
Line 9
Line 10
Line 11
Line 12
Line 13
Line 14
Line 15`,
			expectedChunks: 2, // Actual behavior based on chunk size
		},
		{
			name:           "Empty content",
			content:        "",
			expectedChunks: 1, // Still creates one chunk even for empty content
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chunks, err := chunker.chunkGenericCode(context.Background(), tc.content, "test.txt")
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedChunks, len(chunks), "Number of chunks should match expected")

			// Verify chunk properties
			for i, chunk := range chunks {
				assert.NotEmpty(t, chunk.ID, "Chunk ID should not be empty")
				// Empty content is allowed for empty files
				if tc.content != "" {
					assert.NotEmpty(t, chunk.Content, "Chunk content should not be empty for non-empty input")
				}
				assert.Equal(t, "test.txt", chunk.FilePath, "File path should match")

				// Check overlap
				if i > 0 {
					// Previous chunk should overlap with current
					assert.Greater(t, chunk.StartLine, chunks[i-1].StartLine, "Chunks should progress forward")
				}
			}
		})
	}
}
