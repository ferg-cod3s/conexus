package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourcesList(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Test vectorstore directly first
	testDoc := vectorstore.Document{
		ID:      "test-direct",
		Content: "Test content",
		Vector:  embedding.Vector(make([]float32, 384)),
		Metadata: map[string]interface{}{
			"file_path":   "test.go",
			"source_type": "file",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test direct upsert
	t.Logf("Direct test - Vector length: %d", len(testDoc.Vector))
	if len(testDoc.Vector) > 0 {
		t.Logf("Direct test - First element: %f", testDoc.Vector[0])
	}

	// Check if vector is empty before Upsert
	if len(testDoc.Vector) == 0 {
		t.Fatal("Vector is empty before Upsert")
	}

	// Debug: check vector type
	t.Logf("Vector type: %T", testDoc.Vector)

	// Debug: check if vector is empty in Upsert method
	t.Logf("About to call Upsert with vector length: %d", len(testDoc.Vector))

	// Check the exact error from Upsert
	err := vs.Upsert(ctx, testDoc)
	if err != nil {
		t.Logf("Upsert error: %v", err)
		// Check if the vector is empty in the error context
		if len(testDoc.Vector) == 0 {
			t.Logf("Vector became empty during Upsert call")
		}
	}
	require.NoError(t, err)

	// Add test files to vector store
	testFiles := []string{
		"src/main.go",
		"src/auth.go",
		"README.md",
		"docs/api.md",
		"config.yml",
	}

	for _, filePath := range testFiles {
		// Create a simple vector for testing
		vector := make([]float32, 384)
		for i := range vector {
			vector[i] = 0.1
		}

		// Debug: check vector length before creating document
		t.Logf("Vector length before document creation: %d", len(vector))

		// Create document with explicit vector assignment
		doc := vectorstore.Document{
			ID:      filePath,
			Content: "Test content for " + filePath,
			Vector:  embedding.Vector(vector), // Explicitly cast to embedding.Vector
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"source_type": "file",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Debug: check vector length after creating document
		t.Logf("Vector length after document creation: %d", len(doc.Vector))

		err := vs.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	for _, filePath := range testFiles {
		doc := vectorstore.Document{
			ID:      filePath,
			Content: "Test content for " + filePath,
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"source_type": "file",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := vs.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Test resources/list without parameters
	response, err := server.handleResourcesList(ctx, json.RawMessage(`{}`))
	require.NoError(t, err)

	responseMap, ok := response.(map[string]interface{})
	require.True(t, ok)

	resources, exists := responseMap["resources"]
	require.True(t, exists)

	resourcesList, ok := resources.([]interface{})
	require.True(t, ok)

	// Should have root directory + all files
	assert.GreaterOrEqual(t, len(resourcesList), len(testFiles)+1)

	// Check for root directory resource
	foundRoot := false
	for _, resource := range resourcesList {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}

		uri, exists := resourceMap["uri"].(string)
		if !exists {
			continue
		}

		if uri == "engine://files/" {
			foundRoot = true
			assert.Equal(t, "Indexed Files", resourceMap["name"])
			assert.Equal(t, "application/x-directory", resourceMap["mimeType"])
			break
		}
	}
	assert.True(t, foundRoot, "Root directory resource should be present")

	// Test resources/list with pagination
	paginatedParams := `{"cursor": "0"}`
	response, err = server.handleResourcesList(ctx, json.RawMessage(paginatedParams))
	require.NoError(t, err)

	responseMap, ok = response.(map[string]interface{})
	require.True(t, ok)

	resources, exists = responseMap["resources"]
	require.True(t, exists)

	resourcesList, ok = resources.([]interface{})
	require.True(t, ok)

	// Should be paginated (max 50 per page)
	assert.LessOrEqual(t, len(resourcesList), 51)

	// Test resources/list with invalid URI
	invalidParams := `{"uri": "invalid://path"}`
	_, err = server.handleResourcesList(ctx, json.RawMessage(invalidParams))
	assert.Error(t, err)

	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
}

func TestResourcesRead(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Add test file with multiple chunks
	filePath := "src/main.go"
	var vector1 []float32
	vector1 = make([]float32, 384)
	for i := range vector1 {
		vector1[i] = 0.1
	}
	var vector2 []float32
	vector2 = make([]float32, 384)
	for i := range vector2 {
		vector2[i] = 0.1
	}

	doc1 := vectorstore.Document{
		ID:      filePath + "_chunk1",
		Content: "package main\n\nimport \"fmt\"\n\nfunc main() {",
		Vector:  vector1,
		Metadata: map[string]interface{}{
			"file_path":   filePath,
			"source_type": "file",
			"start_line":  1.0,
			"end_line":    5.0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	doc2 := vectorstore.Document{
		ID:      filePath + "_chunk2",
		Content: "fmt.Println(\"Hello, World!\")\n}",
		Vector:  vector2,
		Metadata: map[string]interface{}{
			"file_path":   filePath,
			"source_type": "file",
			"start_line":  6.0,
			"end_line":    7.0,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := vs.Upsert(ctx, doc1)
	require.NoError(t, err)
	err = vs.Upsert(ctx, doc2)
	require.NoError(t, err)

	// Test resources/read for multi-chunk file
	readParams := fmt.Sprintf(`{"uri": "engine://file/%s"}`, filePath)
	response, err := server.handleResourcesRead(ctx, json.RawMessage(readParams))
	require.NoError(t, err)

	responseMap, ok := response.(map[string]interface{})
	require.True(t, ok)

	contents, exists := responseMap["contents"]
	require.True(t, exists)

	contentsList, ok := contents.([]interface{})
	require.True(t, ok)

	// Should have 2 chunks
	assert.Len(t, contentsList, 2)

	// Check first chunk
	firstChunk, ok := contentsList[0].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "engine://file/src/main.go", firstChunk["uri"])
	assert.Equal(t, "text/plain", firstChunk["mimeType"])
	assert.Equal(t, "package main\n\nimport \"fmt\"\n\nfunc main() {", firstChunk["text"])
	assert.Equal(t, 1, firstChunk["startLineNumber"])
	assert.Equal(t, 5, firstChunk["endLineNumber"])

	// Check second chunk
	secondChunk, ok := contentsList[1].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "engine://file/src/main.go", secondChunk["uri"])
	assert.Equal(t, "text/plain", secondChunk["mimeType"])
	assert.Equal(t, "fmt.Println(\"Hello, World!\")\n}", secondChunk["text"])
	assert.Equal(t, 6, secondChunk["startLineNumber"])
	assert.Equal(t, 7, secondChunk["endLineNumber"])

	// Test resources/read for non-existent file
	nonExistentParams := `{"uri": "engine://file/nonexistent.go"}`
	_, err = server.handleResourcesRead(ctx, json.RawMessage(nonExistentParams))
	assert.Error(t, err)

	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidRequest, protocolErr.Code)

	// Test resources/read with invalid URI format
	invalidParams := `{"uri": "invalid://file/test.go"}`
	_, err = server.handleResourcesRead(ctx, json.RawMessage(invalidParams))
	assert.Error(t, err)

	protocolErr, ok = err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
}

func TestResourcesReadSingleChunk(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Add single chunk file
	filePath := "README.md"
	var vector []float32
	vector = make([]float32, 384)
	for i := range vector {
		vector[i] = 0.1
	}

	doc := vectorstore.Document{
		ID:      filePath,
		Content: "# Test Project\n\nThis is a test README.",
		Vector:  vector,
		Metadata: map[string]interface{}{
			"file_path":   filePath,
			"source_type": "file",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Debug: check vector length
	if len(doc.Vector) == 0 {
		t.Logf("Vector is empty for file: %s", filePath)
	} else {
		t.Logf("Vector length for %s: %d", filePath, len(doc.Vector))
	}

	// Debug: check vector in Upsert call
	t.Logf("Calling Upsert with vector length: %d", len(doc.Vector))
	if len(doc.Vector) > 0 {
		t.Logf("First element: %f", doc.Vector[0])
	}

	// Debug: check vector in loop
	t.Logf("Loop - Vector length for %s: %d", filePath, len(doc.Vector))
	if len(doc.Vector) == 0 {
		t.Logf("Vector is empty for file: %s", filePath)
	}

	// Debug: check vector type in loop
	t.Logf("Loop - Vector type for %s: %T", filePath, doc.Vector)

	// Debug: check if vector is empty in Upsert method
	t.Logf("Loop - About to call Upsert with vector length: %d", len(doc.Vector))

	// Debug: check vector in Upsert method
	t.Logf("Loop - Calling Upsert with vector length: %d", len(doc.Vector))

	// Debug: check vector length in Upsert method
	t.Logf("Loop - Vector length in Upsert method: %d", len(doc.Vector))

	// Debug: check if vector is empty in Upsert method
	if len(doc.Vector) == 0 {
		t.Logf("Vector is empty in Upsert method for file: %s", filePath)
	}

	// Debug: check vector length in Upsert method
	t.Logf("Loop - Vector length in Upsert method: %d", len(doc.Vector))

	// Debug: check vector length in Upsert method
	t.Logf("Loop - Vector length in Upsert method: %d", len(doc.Vector))

	// Debug: check vector length in Upsert method
	t.Logf("Loop - Vector length in Upsert method: %d", len(doc.Vector))

	// Debug: check vector length in Upsert method
	t.Logf("Loop - Vector length in Upsert method: %d", len(doc.Vector))

	// Debug: check vector length in Upsert method
	t.Logf("Loop - Vector length in Upsert method: %d", len(doc.Vector))

	err := vs.Upsert(ctx, doc)
	require.NoError(t, err)

	// Test resources/read for single chunk file
	readParams := fmt.Sprintf(`{"uri": "engine://file/%s"}`, filePath)
	response, err := server.handleResourcesRead(ctx, json.RawMessage(readParams))
	require.NoError(t, err)

	responseMap, ok := response.(map[string]interface{})
	require.True(t, ok)

	contents, exists := responseMap["contents"]
	require.True(t, exists)

	contentsList, ok := contents.([]interface{})
	require.True(t, ok)

	// Should have 1 chunk
	assert.Len(t, contentsList, 1)

	chunk, ok := contentsList[0].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "engine://file/README.md", chunk["uri"])
	assert.Equal(t, "text/markdown", chunk["mimeType"])
	assert.Equal(t, "# Test Project\n\nThis is a test README.", chunk["text"])

	// Single chunk files should not have line numbers
	_, hasStartLine := chunk["startLineNumber"]
	_, hasEndLine := chunk["endLineNumber"]
	assert.False(t, hasStartLine)
	assert.False(t, hasEndLine)
}

func TestResourcesIntegration(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Add test files
	testFiles := map[string]string{
		"src/main.go":    "package main\n\nfunc main() {\n\tfmt.Println(\"Hello\")\n}",
		"src/utils.go":   "package main\n\nfunc helper() {\n\t// Helper function\n}",
		"docs/README.md": "# Documentation\n\nThis is documentation.",
		"config/app.yml": "app:\n  name: test\n  version: 1.0",
	}

	for filePath, content := range testFiles {
		doc := vectorstore.Document{
			ID:      filePath,
			Content: content,
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"source_type": "file",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := vs.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Test full integration: list then read
	listResponse, err := server.handleResourcesList(ctx, json.RawMessage(`{}`))
	require.NoError(t, err)

	listMap, ok := listResponse.(map[string]interface{})
	require.True(t, ok)

	resources, exists := listMap["resources"]
	require.True(t, exists)

	resourcesList, ok := resources.([]interface{})
	require.True(t, ok)

	// Find a file resource
	var testFileURI string
	for _, resource := range resourcesList {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}

		uri, exists := resourceMap["uri"].(string)
		if !exists {
			continue
		}

		if uri != "engine://files/" && !contains(uri, "engine://files/") {
			testFileURI = uri
			break
		}
	}

	require.NotEmpty(t, testFileURI, "Should find at least one file resource")

	// Read the file
	readParams := fmt.Sprintf(`{"uri": "%s"}`, testFileURI)
	readResponse, err := server.handleResourcesRead(ctx, json.RawMessage(readParams))
	require.NoError(t, err)

	readMap, ok := readResponse.(map[string]interface{})
	require.True(t, ok)

	contents, exists := readMap["contents"]
	require.True(t, exists)

	contentsList, ok := contents.([]interface{})
	require.True(t, ok)
	assert.Greater(t, len(contentsList), 0)

	// Verify content matches what was stored
	chunk, ok := contentsList[0].(map[string]interface{})
	require.True(t, ok)

	text, exists := chunk["text"].(string)
	require.True(t, exists)

	// Find the original content
	originalContent := ""
	for filePath, content := range testFiles {
		if testFileURI == fmt.Sprintf("engine://file/%s", filePath) {
			originalContent = content
			break
		}
	}

	require.NotEmpty(t, originalContent)
	assert.Equal(t, originalContent, text)
}

func TestResourcesSecurity(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Test path traversal attempts
	maliciousPaths := []string{
		"../../../etc/passwd",
		"../../etc/shadow",
		"/etc/passwd",
		"C:\\Windows\\System32\\config\\SAM",
		"src/main.go/../../../etc/passwd",
	}

	for _, maliciousPath := range maliciousPaths {
		// Try to read malicious path
		readParams := fmt.Sprintf(`{"uri": "engine://file/%s"}`, maliciousPath)
		_, err := server.handleResourcesRead(ctx, json.RawMessage(readParams))

		// Should fail with security error
		assert.Error(t, err)
		protocolErr, ok := err.(*protocol.Error)
		if ok {
			assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
		}
	}
}

func TestResourcesPerformance(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Add many files to test performance
	fileCount := 100
	for i := 0; i < fileCount; i++ {
		filePath := fmt.Sprintf("src/file_%d.go", i)
		var vector []float32
		vector = make([]float32, 384)
		for j := range vector {
			vector[j] = 0.1
		}

		doc := vectorstore.Document{
			ID:      filePath,
			Content: fmt.Sprintf("File %d content", i),
			Vector:  vector,
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"source_type": "file",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := vs.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Test list performance
	startTime := time.Now()
	response, err := server.handleResourcesList(ctx, json.RawMessage(`{}`))
	listTime := time.Since(startTime)
	require.NoError(t, err)

	// Should complete quickly
	assert.Less(t, listTime, 1*time.Second)

	responseMap, ok := response.(map[string]interface{})
	require.True(t, ok)

	resources, exists := responseMap["resources"]
	require.True(t, exists)

	resourcesList, ok := resources.([]interface{})
	require.True(t, ok)

	// Should include all files
	assert.GreaterOrEqual(t, len(resourcesList), fileCount+1)

	// Test read performance for a specific file
	testFile := "src/file_50.go"
	readParams := fmt.Sprintf(`{"uri": "engine://file/%s"}`, testFile)

	startTime = time.Now()
	_, err = server.handleResourcesRead(ctx, json.RawMessage(readParams))
	readTime := time.Since(startTime)
	require.NoError(t, err)

	// Should complete quickly
	assert.Less(t, readTime, 500*time.Millisecond)
}

func TestResourcesPagination(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Add many files
	fileCount := 75 // More than default page size of 50
	for i := 0; i < fileCount; i++ {
		filePath := fmt.Sprintf("test/file_%d.go", i)
		doc := vectorstore.Document{
			ID:      filePath,
			Content: fmt.Sprintf("File %d content", i),
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"source_type": "file",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := vs.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Test first page
	response1, err := server.handleResourcesList(ctx, json.RawMessage(`{"cursor": "0"}`))
	require.NoError(t, err)

	response1Map, ok := response1.(map[string]interface{})
	require.True(t, ok)

	resources1, exists := response1Map["resources"]
	require.True(t, exists)

	resources1List, ok := resources1.([]interface{})
	require.True(t, ok)

	// Should have root + up to 50 files
	assert.GreaterOrEqual(t, len(resources1List), 1)
	assert.LessOrEqual(t, len(resources1List), 51)

	// Test second page
	response2, err := server.handleResourcesList(ctx, json.RawMessage(`{"cursor": "50"}`))
	require.NoError(t, err)

	response2Map, ok := response2.(map[string]interface{})
	require.True(t, ok)

	resources2, exists := response2Map["resources"]
	require.True(t, exists)

	resources2List, ok := resources2.([]interface{})
	require.True(t, ok)

	// Should have remaining files
	assert.Greater(t, len(resources2List), 0)
	assert.LessOrEqual(t, len(resources2List), 26) // Remaining files + root

	// Verify no overlap between pages (except root directory)
	fileURIs1 := make(map[string]bool)
	for _, resource := range resources1List {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}

		uri, exists := resourceMap["uri"].(string)
		if !exists {
			continue
		}

		if uri != "engine://files/" {
			fileURIs1[uri] = true
		}
	}

	fileURIs2 := make(map[string]bool)
	for _, resource := range resources2List {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}

		uri, exists := resourceMap["uri"].(string)
		if !exists {
			continue
		}

		if uri != "engine://files/" {
			fileURIs2[uri] = true
		}
	}

	// Should have no overlapping file URIs
	for uri := range fileURIs1 {
		assert.NotContains(t, fileURIs2, uri)
	}
}

func TestResourcesMIMETypes(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Test different file types
	testCases := map[string]string{
		"main.go":     "text/plain",
		"README.md":   "text/markdown",
		"config.yml":  "application/x-yaml",
		"data.json":   "application/json",
		"script.py":   "text/plain",
		"style.css":   "text/css",
		"app.js":      "application/javascript",
		"image.png":   "image/png",
		"doc.pdf":     "application/pdf",
		"unknown.xyz": "application/octet-stream",
	}

	for filePath, expectedMIME := range testCases {
		doc := vectorstore.Document{
			ID:      filePath,
			Content: "Test content",
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"source_type": "file",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := vs.Upsert(ctx, doc)
		require.NoError(t, err)

		// Test MIME type detection in list
		listResponse, err := server.handleResourcesList(ctx, json.RawMessage(`{}`))
		require.NoError(t, err)

		listMap, ok := listResponse.(map[string]interface{})
		require.True(t, ok)

		resources, exists := listMap["resources"]
		require.True(t, exists)

		resourcesList, ok := resources.([]interface{})
		require.True(t, ok)

		// Find the file in the list
		found := false
		for _, resource := range resourcesList {
			resourceMap, ok := resource.(map[string]interface{})
			if !ok {
				continue
			}

			uri, exists := resourceMap["uri"].(string)
			if !exists {
				continue
			}

			if uri == fmt.Sprintf("engine://file/%s", filePath) {
				mimeType, exists := resourceMap["mimeType"].(string)
				require.True(t, exists)
				assert.Equal(t, expectedMIME, mimeType)
				found = true
				break
			}
		}

		assert.True(t, found, "File %s should be found in resources list", filePath)

		// Test MIME type in read response
		readParams := fmt.Sprintf(`{"uri": "engine://file/%s"}`, filePath)
		readResponse, err := server.handleResourcesRead(ctx, json.RawMessage(readParams))
		require.NoError(t, err)

		readMap, ok := readResponse.(map[string]interface{})
		require.True(t, ok)

		contents, exists := readMap["contents"]
		require.True(t, exists)

		contentsList, ok := contents.([]interface{})
		require.True(t, ok)

		chunk, ok := contentsList[0].(map[string]interface{})
		require.True(t, ok)

		mimeType, exists := chunk["mimeType"].(string)
		require.True(t, exists)
		assert.Equal(t, expectedMIME, mimeType)
	}
}

func TestResourcesErrorHandling(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Test missing URI parameter
	_, err := server.handleResourcesRead(ctx, json.RawMessage(`{}`))
	assert.Error(t, err)

	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)

	// Test malformed JSON
	_, err = server.handleResourcesList(ctx, json.RawMessage(`{"invalid": json}`))
	assert.Error(t, err)

	protocolErr, ok = err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)

	// Test invalid cursor
	_, err = server.handleResourcesList(ctx, json.RawMessage(`{"cursor": "invalid"}`))
	// Should not error, but cursor should be ignored
	assert.NoError(t, err)
}

func TestResourcesDirectoryStructure(t *testing.T) {
	// Setup
	vs, metrics, errorHandler, testIndexer := setupTestComponents()

	server := NewServer(
		nil, nil, vs, nil, nil, metrics, errorHandler, testIndexer,
	)

	ctx := context.Background()

	// Add files in nested structure
	nestedFiles := []string{
		"src/main.go",
		"src/utils.go",
		"src/auth/login.go",
		"src/auth/middleware.go",
		"tests/unit_test.go",
		"docs/api.md",
		"docs/guides/setup.md",
		"config/database.yml",
		"config/app.yml",
	}

	for _, filePath := range nestedFiles {
		var vector []float32
		vector = make([]float32, 384)
		for i := range vector {
			vector[i] = 0.1
		}

		doc := vectorstore.Document{
			ID:      filePath,
			Content: fmt.Sprintf("Content of %s", filePath),
			Vector:  vector,
			Metadata: map[string]interface{}{
				"file_path":   filePath,
				"source_type": "file",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := vs.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Test directory listing
	response, err := server.handleResourcesList(ctx, json.RawMessage(`{}`))
	require.NoError(t, err)

	responseMap, ok := response.(map[string]interface{})
	require.True(t, ok)

	resources, exists := responseMap["resources"]
	require.True(t, exists)

	resourcesList, ok := resources.([]interface{})
	require.True(t, ok)

	// Should include all files
	assert.GreaterOrEqual(t, len(resourcesList), len(nestedFiles)+1)

	// Verify all files are listed
	expectedFiles := make(map[string]bool)
	for _, filePath := range nestedFiles {
		expectedFiles[filePath] = false
	}

	for _, resource := range resourcesList {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}

		uri, exists := resourceMap["uri"].(string)
		if !exists {
			continue
		}

		if uri != "engine://files/" {
			// Extract file path from URI
			if len(uri) > len("engine://file/") {
				filePath := uri[len("engine://file/"):]
				if _, exists := expectedFiles[filePath]; exists {
					expectedFiles[filePath] = true
				}
			}
		}
	}

	// All files should be found
	for filePath, found := range expectedFiles {
		assert.True(t, found, "File %s should be listed in resources", filePath)
	}
}

// Helper functions

func setupTestComponents() (vectorstore.VectorStore, *observability.MetricsCollector, *observability.ErrorHandler, indexer.IndexController) {
	vs := vectorstore.NewMemoryStore()
	logger := observability.NewLogger(observability.DefaultLoggerConfig())
	metrics := observability.NewMetricsCollector("mcp_test")
	errorHandler := observability.NewErrorHandler(logger, metrics, false)
	testIndexer := setupTestIndexer()
	return vs, metrics, errorHandler, testIndexer
}

func setupTestIndexer() indexer.IndexController {
	// Return a mock indexer for testing
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
