package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock embedder for testing
type mockEmbedder struct {
	embedFunc  func(ctx context.Context, text string) (*embedding.Embedding, error)
	dimensions int
	model      string
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) (*embedding.Embedding, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, text)
	}
	dims := m.dimensions
	if dims == 0 {
		dims = 384
	}
	return &embedding.Embedding{
		Text:   text,
		Vector: make(embedding.Vector, dims),
		Model:  m.model,
	}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([]*embedding.Embedding, error) {
	result := make([]*embedding.Embedding, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		result[i] = emb
	}
	return result, nil
}

// Mock indexer for testing
type mockIndexer struct {
	status indexer.IndexStatus
}

func (m *mockIndexer) Start(ctx context.Context, opts indexer.IndexOptions) error {
	m.status.IsIndexing = true
	m.status.Phase = "running"
	return nil
}

func (m *mockIndexer) Stop(ctx context.Context) error {
	m.status.IsIndexing = false
	m.status.Phase = "stopped"
	return nil
}

func (m *mockIndexer) ForceReindex(ctx context.Context, opts indexer.IndexOptions) error {
	m.status.IsIndexing = true
	m.status.Phase = "force_reindex"
	return nil
}

func (m *mockIndexer) ReindexPaths(ctx context.Context, opts indexer.IndexOptions, paths []string) error {
	m.status.IsIndexing = true
	m.status.Phase = "reindex_paths"
	return nil
}

func (m *mockIndexer) GetStatus() indexer.IndexStatus {
	return m.status
}

func (m *mockIndexer) HealthCheck(ctx context.Context) error {
	return nil
}

func (m *mockEmbedder) Dimensions() int {
	if m.dimensions == 0 {
		return 384
	}
	return m.dimensions
}

func (m *mockEmbedder) Model() string {
	if m.model == "" {
		return "mock"
	}
	return m.model
}

// Mock connector store for testing
type mockConnectorStore struct {
	connectors map[string]*connectors.Connector
}

func newMockConnectorStore() *mockConnectorStore {
	return &mockConnectorStore{
		connectors: make(map[string]*connectors.Connector),
	}
}

func (m *mockConnectorStore) Add(ctx context.Context, connector *connectors.Connector) error {
	if _, exists := m.connectors[connector.ID]; exists {
		return fmt.Errorf("connector with ID %s already exists", connector.ID)
	}
	m.connectors[connector.ID] = connector
	return nil
}

func (m *mockConnectorStore) Update(ctx context.Context, id string, connector *connectors.Connector) error {
	if _, exists := m.connectors[id]; !exists {
		return fmt.Errorf("connector with ID %s not found", id)
	}
	connector.ID = id
	m.connectors[id] = connector
	return nil
}

func (m *mockConnectorStore) Remove(ctx context.Context, id string) error {
	if _, exists := m.connectors[id]; !exists {
		return fmt.Errorf("connector with ID %s not found", id)
	}
	delete(m.connectors, id)
	return nil
}

func (m *mockConnectorStore) List(ctx context.Context) ([]*connectors.Connector, error) {
	result := make([]*connectors.Connector, 0, len(m.connectors))
	for _, conn := range m.connectors {
		result = append(result, conn)
	}
	return result, nil
}

func (m *mockConnectorStore) Get(ctx context.Context, id string) (*connectors.Connector, error) {
	connector, exists := m.connectors[id]
	if !exists {
		return nil, fmt.Errorf("connector with ID %s not found", id)
	}
	return connector, nil
}

func (m *mockConnectorStore) Close() error {
	return nil
}

func TestNewServer(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	assert.NotNil(t, server)
	assert.NotNil(t, server.vectorStore)
	assert.NotNil(t, server.embedder)
	assert.NotNil(t, server.indexer)
	assert.NotNil(t, server.jsonrpcSrv)
}

func TestServer_Close(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)
	err := server.Close()
	assert.NoError(t, err)
}

func TestServer_Handle_ToolsList(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	result, err := server.Handle("tools/list", nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	// The result should be map[string]interface{} with "tools" key
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")

	toolsInterface, ok := resultMap["tools"]
	require.True(t, ok, "result should have 'tools' key")

	// Convert to []ToolDefinition for verification
	resultJSON, err := json.Marshal(toolsInterface)
	require.NoError(t, err)

	var tools []ToolDefinition
	err = json.Unmarshal(resultJSON, &tools)
	require.NoError(t, err)

	// Verify we have 6 tools
	assert.Len(t, tools, 6)

	// Verify tool names
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}
	assert.True(t, toolNames[ToolContextSearch])
	assert.True(t, toolNames[ToolContextGetRelatedInfo])
	assert.True(t, toolNames[ToolContextIndexControl])
	assert.True(t, toolNames[ToolContextConnectorManagement])
	assert.True(t, toolNames[ToolContextExplain])
	assert.True(t, toolNames[ToolContextGrep])
}

func TestServer_Handle_ContextSearch(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	// Add some test documents
	ctx := context.Background()
	now := time.Now()
	doc := vectorstore.Document{
		ID:        "test-1",
		Content:   "test content for search",
		Vector:    make(embedding.Vector, 384),
		Metadata:  map[string]interface{}{"source_type": "file"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	req := SearchRequest{
		Query: "test query",
		TopK:  5,
	}

	// Marshal request to JSON
	reqJSON, err := json.Marshal(map[string]interface{}{
		"name":      ToolContextSearch,
		"arguments": req,
	})
	require.NoError(t, err)

	result, err := server.Handle("tools/call", json.RawMessage(reqJSON))

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestServer_Handle_UnknownMethod(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	_, err := server.Handle("unknown/method", nil)
	assert.Error(t, err)

	// Should be a protocol error
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.MethodNotFound, protocolErr.Code)
}

func TestServer_Serve(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	// Test Serve - it should return nil for now (placeholder implementation)
	err := server.Serve()
	assert.NoError(t, err)
}

func TestServer_HandleInitialize(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	ctx := context.Background()

	// Create initialize request
	req := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"clientInfo": map[string]interface{}{
			"name":    "test-client",
			"version": "1.0.0",
		},
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Test handleInitialize
	result, err := server.handleInitialize(ctx, json.RawMessage(reqJSON))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify response structure
	response, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2025-06-18", response["protocolVersion"])
	assert.Contains(t, response, "capabilities")
	assert.Contains(t, response, "serverInfo")
}

func TestServer_HandleResourcesList(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	ctx := context.Background()

	// Test handleResourcesList
	result, err := server.handleResourcesList(ctx, json.RawMessage("{}"))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify response structure
	response, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, response, "resources")
}

func TestServer_HandleResourcesRead(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	ctx := context.Background()

	// Add a test document to the vector store
	now := time.Now()
	doc := vectorstore.Document{
		ID:      "test-doc",
		Content: "test content",
		Vector:  make(embedding.Vector, 384),
		Metadata: map[string]interface{}{
			"source_type": "file",
			"file_path":   "test.txt",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Create read request with correct URI format
	req := map[string]interface{}{
		"uri": "engine://file/test.txt",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Test handleResourcesRead
	result, err := server.handleResourcesRead(ctx, json.RawMessage(reqJSON))

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify response structure
	response, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, response, "contents")
}

func TestServer_ValidateFilePath(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	tests := []struct {
		name        string
		filePath    string
		expectError bool
	}{
		{"relative file path", "path/to/file.txt", false},
		{"path traversal attempt", "../../../etc/passwd", true},
		{"absolute path", "/absolute/path/file.go", true},
		{"relative path", "relative/path/file.go", false},
		{"empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateFilePath(tt.filePath)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_GetMimeType(t *testing.T) {
	reader := strings.NewReader("")
	writer := &bytes.Buffer{}
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}

	server := NewServer(reader, writer, store, connectorStore, embedder, nil, nil, mockIdx)

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{"go file", "main.go", "text/x-go"},
		{"javascript file", "app.js", "application/javascript"},
		{"typescript file", "app.ts", "application/typescript"},
		{"json file", "config.json", "application/json"},
		{"text file", "readme.txt", "text/plain"},
		{"unknown file", "unknown.xyz", "text/plain"},
		{"no extension", "Makefile", "text/plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.getMimeType(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}
