package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRelationshipDetector(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
	}{
		{
			name:           "go file",
			targetFilePath: "internal/mcp/handlers.go",
		},
		{
			name:           "test file",
			targetFilePath: "internal/mcp/handlers_test.go",
		},
		{
			name:           "nested directory",
			targetFilePath: "internal/vectorstore/sqlite/store.go",
		},
		{
			name:           "empty path",
			targetFilePath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			require.NotNil(t, detector)
		})
	}
}

func TestDetectRelationType_TestFiles(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
	}{
		// Go test files
		{
			name:           "go test file for implementation",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/handlers_test.go",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		{
			name:           "go implementation for test file",
			targetFilePath: "internal/mcp/handlers_test.go",
			relatedPath:    "internal/mcp/handlers.go",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		// Java test files
		{
			name:           "java test with Test suffix",
			targetFilePath: "src/main/Handler.java",
			relatedPath:    "src/test/HandlerTest.java",
			chunkType:      "class",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		{
			name:           "java test with Test prefix",
			targetFilePath: "src/main/Handler.java",
			relatedPath:    "src/test/TestHandler.java",
			chunkType:      "class",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		// Python test files
		{
			name:           "python test with test_ prefix",
			targetFilePath: "src/handler.py",
			relatedPath:    "tests/test_handler.py",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		{
			name:           "python test with _test suffix",
			targetFilePath: "src/handler.py",
			relatedPath:    "tests/handler_test.py",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		// JavaScript test files
		{
			name:           "js test file with .test.js",
			targetFilePath: "src/handler.js",
			relatedPath:    "src/handler.test.js",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		{
			name:           "js spec file with .spec.js",
			targetFilePath: "src/handler.js",
			relatedPath:    "src/handler.spec.js",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		{
			name:           "typescript test file",
			targetFilePath: "src/handler.ts",
			relatedPath:    "src/handler.test.ts",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
		// Rust test files
		{
			name:           "rust file in tests directory",
			targetFilePath: "src/handler.rs",
			relatedPath:    "tests/handler.rs",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got, "Expected %s relationship between %s and %s", tt.want, tt.targetFilePath, tt.relatedPath)
		})
	}
}

func TestDetectRelationType_Documentation(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
	}{
		{
			name:           "markdown file",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "docs/api-reference.md",
			chunkType:      "paragraph",
			metadata:       map[string]interface{}{},
			want:           RelationTypeDocumentation,
		},
		{
			name:           "readme file",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "README.md",
			chunkType:      "paragraph",
			metadata:       map[string]interface{}{},
			want:           RelationTypeDocumentation,
		},
		{
			name:           "rst file",
			targetFilePath: "src/handler.py",
			relatedPath:    "docs/guide.rst",
			chunkType:      "paragraph",
			metadata:       map[string]interface{}{},
			want:           RelationTypeDocumentation,
		},
		{
			name:           "docs directory",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "docs/architecture/design.txt",
			chunkType:      "paragraph",
			metadata:       map[string]interface{}{},
			want:           RelationTypeDocumentation,
		},
		{
			name:           "documentation directory",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "documentation/api.txt",
			chunkType:      "paragraph",
			metadata:       map[string]interface{}{},
			want:           RelationTypeDocumentation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got, "Expected documentation relationship")
		})
	}
}

func TestDetectRelationType_SymbolReferences(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
	}{
		{
			name:           "function chunk type",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSymbolRef,
		},
		{
			name:           "class chunk type",
			targetFilePath: "src/Handler.java",
			relatedPath:    "src/Server.java",
			chunkType:      "class",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSymbolRef,
		},
		{
			name:           "struct chunk type",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/types.go",
			chunkType:      "struct",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSymbolRef,
		},
		{
			name:           "interface chunk type",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/interfaces.go",
			chunkType:      "interface",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSymbolRef,
		},
		{
			name:           "method chunk type",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "method",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSymbolRef,
		},
		{
			name:           "symbol_name in metadata",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "code",
			metadata:       map[string]interface{}{"symbol_name": "NewServer"},
			want:           RelationTypeSymbolRef,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got, "Expected symbol_ref relationship")
		})
	}
}

func TestDetectRelationType_Imports(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
	}{
		{
			name:           "same directory go files",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeImport,
		},
		{
			name:           "same directory python files",
			targetFilePath: "src/handlers.py",
			relatedPath:    "src/server.py",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeImport,
		},
		{
			name:           "same directory js files",
			targetFilePath: "src/handlers.js",
			relatedPath:    "src/server.js",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeImport,
		},
		{
			name:           "parent/child go package",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/types/request.go",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeImport,
		},
		{
			name:           "child/parent go package",
			targetFilePath: "internal/mcp/types/request.go",
			relatedPath:    "internal/mcp/handlers.go",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeImport,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got, "Expected import relationship")
		})
	}
}

func TestDetectRelationType_SimilarCode(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
	}{
		{
			name:           "same language different directory",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/vectorstore/handlers.go",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSimilarCode,
		},
		{
			name:           "same extension different package",
			targetFilePath: "src/api/handlers.py",
			relatedPath:    "src/db/handlers.py",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSimilarCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got, "Expected similar_code relationship")
		})
	}
}

func TestDetectRelationType_Unknown(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
	}{
		{
			name:           "different languages",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "scripts/test.py",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeUnknown,
		},
		{
			name:           "non-code files",
			targetFilePath: "config.yml",
			relatedPath:    "data.json",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeUnknown,
		},
		{
			name:           "empty target path",
			targetFilePath: "",
			relatedPath:    "internal/mcp/handlers.go",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeUnknown,
		},
		{
			name:           "empty related path",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got, "Expected unknown relationship")
		})
	}
}

func TestDetectRelationType_PriorityOrder(t *testing.T) {
	// Test that relationship detection follows the correct priority order
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
		description    string
	}{
		{
			name:           "test file beats documentation",
			targetFilePath: "handler.go",
			relatedPath:    "handler_test.go",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
			description:    "Test file relationship has higher priority than function symbol",
		},
		{
			name:           "documentation beats symbol ref",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "docs/handlers.md",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeDocumentation,
			description:    "Documentation has higher priority than symbol reference",
		},
		{
			name:           "symbol ref beats import",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSymbolRef,
			description:    "Symbol reference has higher priority than import (same directory)",
		},
		{
			name:           "import beats similar code",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/utils.go",
			chunkType:      "code",
			metadata:       map[string]interface{}{},
			want:           RelationTypeImport,
			description:    "Import relationship has higher priority than similar code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got, tt.description)
		})
	}
}

func TestDetectRelationType_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		targetFilePath string
		relatedPath    string
		chunkType      string
		metadata       map[string]interface{}
		want           string
	}{
		{
			name:           "nil metadata",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "function",
			metadata:       nil,
			want:           RelationTypeSymbolRef,
		},
		{
			name:           "empty metadata",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeSymbolRef,
		},
		{
			name:           "empty chunk type",
			targetFilePath: "internal/mcp/handlers.go",
			relatedPath:    "internal/mcp/server.go",
			chunkType:      "",
			metadata:       map[string]interface{}{},
			want:           RelationTypeImport,
		},
		{
			name:           "case sensitivity in paths",
			targetFilePath: "Internal/MCP/Handlers.go",
			relatedPath:    "Internal/MCP/Handlers_Test.go",
			chunkType:      "function",
			metadata:       map[string]interface{}{},
			want:           RelationTypeTestFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := NewRelationshipDetector(tt.targetFilePath)
			got := detector.DetectRelationType(tt.relatedPath, tt.chunkType, tt.metadata)
			assert.Equal(t, tt.want, got)
		})
	}
}
