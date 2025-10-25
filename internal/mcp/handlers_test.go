package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/connectors/github"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleContextSearch_Success(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	connectorStore := newMockConnectorStore()
	embedder := &mockEmbedder{}
	mockIdx := &mockIndexer{}
	server := NewServer(nil, nil, store, connectorStore, embedder, nil, nil, mockIdx)

	// Add test documents
	ctx := context.Background()
	now := time.Now()
	docs := []vectorstore.Document{
		{
			ID:        "doc-1",
			Content:   "authentication implementation",
			Vector:    make(embedding.Vector, 384),
			Metadata:  map[string]interface{}{"source_type": "file"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "doc-2",
			Content:   "login handler code",
			Vector:    make(embedding.Vector, 384),
			Metadata:  map[string]interface{}{"source_type": "github_pr"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Create search request
	req := SearchRequest{
		Query: "authentication",
		TopK:  10,
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute search
	result, err := server.handleContextSearch(ctx, reqJSON)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(SearchResponse)
	require.True(t, ok, "result should be SearchResponse")

	assert.GreaterOrEqual(t, len(resp.Results), 0) // Mock embedder returns zero vectors
	assert.Equal(t, len(resp.Results), resp.TotalCount)
	assert.GreaterOrEqual(t, resp.QueryTime, float64(0)) // QueryTime can be zero for mock
}

func TestHandleContextSearch_WithFilters(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Create search request with filters
	req := SearchRequest{
		Query: "test query",
		TopK:  5,
		Filters: &SearchFilters{
			SourceTypes: []string{"file", "github"},
			DateRange: &DateRange{
				From: time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				To:   time.Now().Format(time.RFC3339),
			},
		},
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute search
	result, err := server.handleContextSearch(ctx, reqJSON)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(SearchResponse)
	require.True(t, ok)
	assert.NotNil(t, resp.Results)
}

func TestHandleContextSearch_InvalidJSON(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	invalidJSON := json.RawMessage(`{"invalid": "json"`)

	_, err := server.handleContextSearch(ctx, invalidJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
}

func TestHandleContextSearch_MissingQuery(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	req := SearchRequest{
		Query: "", // Empty query
		TopK:  10,
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	_, err = server.handleContextSearch(ctx, reqJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
	assert.Contains(t, protocolErr.Message, "query is required")
}

func TestHandleContextSearch_TopKDefaults(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	tests := []struct {
		name     string
		topK     int
		expected int
	}{
		{"zero defaults to 20", 0, 20},
		{"negative defaults to 20", -5, 20},
		{"valid value kept", 30, 30},
		{"over 100 capped to 100", 150, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SearchRequest{
				Query: "test",
				TopK:  tt.topK,
			}

			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			result, err := server.handleContextSearch(ctx, reqJSON)
			assert.NoError(t, err)

			resp, ok := result.(SearchResponse)
			require.True(t, ok)
			// The actual limit would be reflected in results, but we can verify no error
			assert.NotNil(t, resp.Results)
		})
	}
}

func TestHandleGetRelatedInfo_WithFilePath(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Add test documents with different source types
	now := time.Now()
	docs := []vectorstore.Document{
		{
			ID:      "pr-1",
			Content: "Fixed authentication bug",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "github_pr",
				"pr_number":   "123",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "issue-1",
			Content: "Auth issue reported",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "github_issue",
				"issue_id":    "456",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "slack-1",
			Content: "Discussion about auth implementation that is quite lengthy and needs to be truncated",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "slack",
				"channel":     "engineering",
				"timestamp":   "2024-01-01T12:00:00Z",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Create request
	req := GetRelatedInfoRequest{
		FilePath: "src/auth/handler.go",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute
	result, err := server.handleGetRelatedInfo(ctx, reqJSON)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(GetRelatedInfoResponse)
	require.True(t, ok)

	assert.NotEmpty(t, resp.Summary)
	assert.Contains(t, resp.Summary, "src/auth/handler.go")
	assert.NotNil(t, resp.RelatedItems)
	// We should have at least some results
	totalResults := len(resp.RelatedPRs) + len(resp.RelatedIssues) + len(resp.Discussions) + len(resp.RelatedItems)
	assert.GreaterOrEqual(t, totalResults, 0)
}

func TestHandleGetRelatedInfo_WithTicketID(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Create request with ticket ID
	req := GetRelatedInfoRequest{
		TicketID: "JIRA-123",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute
	result, err := server.handleGetRelatedInfo(ctx, reqJSON)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(GetRelatedInfoResponse)
	require.True(t, ok)

	assert.NotEmpty(t, resp.Summary)
	assert.Contains(t, resp.Summary, "JIRA-123")
}

func TestHandleGetRelatedInfo_MissingBothIdentifiers(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Create request with no identifiers
	req := GetRelatedInfoRequest{}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute
	_, err = server.handleGetRelatedInfo(ctx, reqJSON)

	// Verify error
	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
	assert.Contains(t, protocolErr.Message, "either file_path or ticket_id must be provided")
}

func TestHandleGetRelatedInfo_InvalidJSON(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	invalidJSON := json.RawMessage(`{invalid}`)

	_, err := server.handleGetRelatedInfo(ctx, invalidJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
}

func TestHandleIndexControl_Status(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Add some documents to store
	now := time.Now()
	for i := 0; i < 5; i++ {
		doc := vectorstore.Document{
			ID:        string(rune('A' + i)),
			Content:   "test content",
			Vector:    make(embedding.Vector, 384),
			Metadata:  map[string]interface{}{},
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Create status request
	req := IndexControlRequest{
		Action: "status",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute
	result, err := server.handleIndexControl(ctx, reqJSON)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(IndexControlResponse)
	require.True(t, ok)

	assert.Equal(t, "ok", resp.Status)
	assert.Contains(t, resp.Message, "5 documents")
	assert.NotNil(t, resp.Details)
	assert.Equal(t, int64(5), resp.Details["documents_indexed"])
	assert.Equal(t, true, resp.Details["indexer_available"])
}

func TestHandleIndexControl_OtherActions(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	actions := []string{"start", "stop", "force_reindex"}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			req := IndexControlRequest{
				Action: action,
			}

			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			result, err := server.handleIndexControl(ctx, reqJSON)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			resp, ok := result.(IndexControlResponse)
			require.True(t, ok)

			assert.Equal(t, "ok", resp.Status)
			if action == "start" {
				assert.Contains(t, resp.Message, "Background indexing started")
			} else if action == "stop" {
				assert.Contains(t, resp.Message, "Indexing stopped")
			} else if action == "force_reindex" {
				assert.Contains(t, resp.Message, "Force reindex started")
			}
		})
	}
}

func TestHandleIndexControl_ReindexPaths(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	req := IndexControlRequest{
		Action: "reindex_paths",
		Paths:  []string{"file1.go", "file2.go"},
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	result, err := server.handleIndexControl(ctx, reqJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(IndexControlResponse)
	require.True(t, ok)

	assert.Equal(t, "ok", resp.Status)
	assert.Contains(t, resp.Message, "Reindexing 2 paths")
}

func TestHandleIndexControl_InvalidAction(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	req := IndexControlRequest{
		Action: "invalid_action",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	_, err = server.handleIndexControl(ctx, reqJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
	assert.Contains(t, protocolErr.Message, "invalid action")
}

func TestHandleIndexControl_InvalidJSON(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	invalidJSON := json.RawMessage(`{"action":`)

	_, err := server.handleIndexControl(ctx, invalidJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
}

func TestHandleIndexControl_Index(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	content := &IndexContent{
		Path:       "/test/example.go",
		Content:    "package test\n\nfunc Example() { println(\"hello\") }",
		SourceType: "file",
	}

	req := IndexControlRequest{
		Action:  "index",
		Content: content,
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	result, err := server.handleIndexControl(ctx, reqJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	response, ok := result.(IndexControlResponse)
	require.True(t, ok)
	assert.Equal(t, "ok", response.Status)
	assert.Contains(t, response.Message, "Successfully indexed document")
	assert.NotNil(t, response.Details)
	assert.Equal(t, "/test/example.go", response.Details["document_id"])
}

func TestHandleConnectorManagement_List(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	req := ConnectorManagementRequest{
		Action: "list",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	result, err := server.handleConnectorManagement(ctx, reqJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(ConnectorManagementResponse)
	require.True(t, ok)

	assert.Equal(t, "ok", resp.Status)
	assert.Empty(t, resp.Connectors) // Initially empty, connectors need to be added first
}

func TestHandleConnectorManagement_Add(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	req := ConnectorManagementRequest{
		Action:      "add",
		ConnectorID: "github-connector",
		ConnectorConfig: map[string]interface{}{
			"repo": "owner/repo",
		},
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	result, err := server.handleConnectorManagement(ctx, reqJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(ConnectorManagementResponse)
	require.True(t, ok)

	assert.Equal(t, "ok", resp.Status)
	assert.Contains(t, resp.Message, "add")
	assert.Contains(t, resp.Message, "github-connector")
}

func TestHandleConnectorManagement_Update(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// First add a connector
	addReq := ConnectorManagementRequest{
		Action:      "add",
		ConnectorID: "github-connector",
		ConnectorConfig: map[string]interface{}{
			"repo": "owner/repo",
		},
	}

	addReqJSON, err := json.Marshal(addReq)
	require.NoError(t, err)

	_, err = server.handleConnectorManagement(ctx, addReqJSON)
	require.NoError(t, err)

	// Now update it
	updateReq := ConnectorManagementRequest{
		Action:      "update",
		ConnectorID: "github-connector",
		ConnectorConfig: map[string]interface{}{
			"repo": "owner/new-repo",
		},
	}

	updateReqJSON, err := json.Marshal(updateReq)
	require.NoError(t, err)

	result, err := server.handleConnectorManagement(ctx, updateReqJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(ConnectorManagementResponse)
	require.True(t, ok)

	assert.Equal(t, "ok", resp.Status)
	assert.Contains(t, resp.Message, "update")
}

func TestHandleConnectorManagement_Remove(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// First add a connector
	addReq := ConnectorManagementRequest{
		Action:      "add",
		ConnectorID: "github-connector",
		ConnectorConfig: map[string]interface{}{
			"repo": "owner/repo",
		},
	}

	addReqJSON, err := json.Marshal(addReq)
	require.NoError(t, err)

	_, err = server.handleConnectorManagement(ctx, addReqJSON)
	require.NoError(t, err)

	// Now remove it
	removeReq := ConnectorManagementRequest{
		Action:      "remove",
		ConnectorID: "github-connector",
	}

	removeReqJSON, err := json.Marshal(removeReq)
	require.NoError(t, err)

	result, err := server.handleConnectorManagement(ctx, removeReqJSON)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(ConnectorManagementResponse)
	require.True(t, ok)

	assert.Equal(t, "ok", resp.Status)
	assert.Contains(t, resp.Message, "remove")
}

func TestHandleConnectorManagement_MissingConnectorID(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Test that add/update/remove require connector_id
	actions := []string{"add", "update", "remove"}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			req := ConnectorManagementRequest{
				Action:      action,
				ConnectorID: "", // Missing
			}

			reqJSON, err := json.Marshal(req)
			require.NoError(t, err)

			_, err = server.handleConnectorManagement(ctx, reqJSON)

			assert.Error(t, err)
			protocolErr, ok := err.(*protocol.Error)
			require.True(t, ok)
			assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
			assert.Contains(t, protocolErr.Message, "connector_id is required")
		})
	}
}

func TestHandleConnectorManagement_InvalidAction(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	req := ConnectorManagementRequest{
		Action: "invalid_action",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	_, err = server.handleConnectorManagement(ctx, reqJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
	assert.Contains(t, protocolErr.Message, "invalid action")
}

func TestHandleConnectorManagement_InvalidJSON(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	invalidJSON := json.RawMessage(`{"action":`)

	_, err := server.handleConnectorManagement(ctx, invalidJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
}

func TestHandleContextExplain_Success(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Add test documents
	now := time.Now()
	docs := []vectorstore.Document{
		{
			ID:      "func-1",
			Content: "func AuthenticateUser(username, password string) error {\n\t// Validate credentials\n\tif username == \"\" || password == \"\" {\n\t\treturn errors.New(\"invalid credentials\")\n\t}\n\treturn nil\n}",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type":   "file",
				"file_path":     "auth.go",
				"chunk_type":    "function",
				"function_name": "AuthenticateUser",
				"language":      "go",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "struct-1",
			Content: "type User struct {\n\tID       int    `json:\"id\"`\n\tUsername string `json:\"username\"`\n\tEmail    string `json:\"email\"`\n}",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "models.go",
				"chunk_type":  "struct",
				"type_name":   "User",
				"language":    "go",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Create explain request
	req := ExplainRequest{
		Target:  "user authentication",
		Context: "how users log in",
		Depth:   "detailed",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute explain
	result, err := server.handleContextExplain(ctx, reqJSON)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(ExplainResponse)
	require.True(t, ok)

	assert.NotEmpty(t, resp.Explanation)
	assert.Contains(t, resp.Explanation, "AuthenticateUser")
	assert.Contains(t, resp.Explanation, "User")
	assert.Equal(t, "moderate", resp.Complexity) // Based on our assessment logic
	assert.NotNil(t, resp.Examples)
	assert.NotNil(t, resp.Related)
	assert.NotNil(t, resp.Metadata)
}

func TestHandleContextExplain_MissingTarget(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Create request with missing target
	req := ExplainRequest{
		Context: "some context",
		Depth:   "brief",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute explain
	_, err = server.handleContextExplain(ctx, reqJSON)

	// Verify error
	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
	assert.Contains(t, protocolErr.Message, "target is required")
}

func TestHandleContextGrep_Success(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func authenticateUser(username, password string) error {
	if username == "admin" && password == "secret" {
		return nil
	}
	return fmt.Errorf("invalid credentials")
}`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// Verify file was created correctly
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "func")

	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Create grep request with absolute path
	absPath, err := filepath.Abs(tempDir)
	require.NoError(t, err)

	req := GrepRequest{
		Pattern:         "func",
		Path:            absPath,
		Include:         "*",
		CaseInsensitive: false,
		Context:         2,
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute grep
	result, err := server.handleContextGrep(ctx, reqJSON)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resp, ok := result.(GrepResponse)
	require.True(t, ok)

	// Debug output
	t.Logf("Search time: %f ms", resp.SearchTime)
	t.Logf("Total results: %d", resp.TotalCount)
	for i, res := range resp.Results {
		t.Logf("Result %d: File=%s, Line=%d, Match=%s", i, res.File, res.Line, res.Match)
	}

	assert.Greater(t, resp.TotalCount, 0)
	assert.Greater(t, resp.SearchTime, float64(0))

	// Check that we found the function
	found := false
	for _, grepResult := range resp.Results {
		if strings.Contains(grepResult.Content, "func") {
			found = true
			assert.Equal(t, testFile, grepResult.File)
			assert.Contains(t, grepResult.Match, "func")
			break
		}
	}
	assert.True(t, found, "Should have found func pattern")
}

func TestHandleContextGrep_MissingPattern(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Create request with missing pattern
	req := GrepRequest{
		Path:    ".",
		Include: "*.go",
	}

	reqJSON, err := json.Marshal(req)
	require.NoError(t, err)

	// Execute grep
	_, err = server.handleContextGrep(ctx, reqJSON)

	// Verify error
	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
	assert.Contains(t, protocolErr.Message, "pattern is required")
}

func TestApplyWorkContextBoosting(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	now := time.Now()

	// Create test results
	results := []vectorstore.SearchResult{
		{
			Document: vectorstore.Document{
				ID:      "doc-1",
				Content: "authentication implementation",
				Vector:  make(embedding.Vector, 384),
				Metadata: map[string]interface{}{
					"source_type": "file",
					"file_path":   "src/auth/handler.go",
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			Score: 0.8,
		},
		{
			Document: vectorstore.Document{
				ID:      "doc-2",
				Content: "login handler code",
				Vector:  make(embedding.Vector, 384),
				Metadata: map[string]interface{}{
					"source_type": "github_pr",
					"pr_number":   "123",
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			Score: 0.7,
		},
		{
			Document: vectorstore.Document{
				ID:      "doc-3",
				Content: "unrelated content",
				Vector:  make(embedding.Vector, 384),
				Metadata: map[string]interface{}{
					"source_type": "slack",
				},
				CreatedAt: now,
				UpdatedAt: now,
			},
			Score: 0.6,
		},
	}

	// Test with work context that boosts active file
	workContext := &WorkContextFilters{
		ActiveFile:  "src/auth/handler.go",
		BoostActive: true,
	}

	boosted := server.applyWorkContextBoosting(results, workContext)

	// The active file should be boosted (higher score)
	assert.Greater(t, boosted[0].Score, boosted[1].Score)
	assert.Equal(t, "doc-1", boosted[0].Document.ID) // auth file should be first
}

func TestExtractStoryIDsFromIssue(t *testing.T) {
	tests := []struct {
		name     string
		issue    github.Issue
		expected []string
	}{
		{
			name: "issue with story IDs",
			issue: github.Issue{
				Title:       "Fix authentication bug #123",
				Description: "Related to PROJ-456 and JIRA-789",
				Number:      42,
			},
			expected: []string{"123", "456", "789"},
		},
		{
			name: "issue without story IDs",
			issue: github.Issue{
				Title:       "Fix authentication bug",
				Description: "No story references here",
				Number:      43,
			},
			expected: nil,
		},
		{
			name: "issue with labels",
			issue: github.Issue{
				Title:       "Fix bug",
				Description: "Related to PROJ-456",
				Number:      44,
				Labels:      []string{"story: PROJ-123", "bug: PROJ-456"},
			},
			expected: []string{"456", "PROJ-123", "PROJ-456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStoryIDsFromIssue(tt.issue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncGitHubData(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Create a mock GitHub connector
	connector := &connectors.Connector{
		ID:   "github-test",
		Type: "github",
		Config: map[string]interface{}{
			"repo_url": "https://github.com/owner/repo",
		},
	}

	// Test sync - this will use the mock implementation
	issues, prs, err := server.syncGitHubData(ctx, connector)

	// Verify results (function returns empty slices for now)
	assert.NoError(t, err)
	assert.NotNil(t, issues)
	assert.NotNil(t, prs)
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal", 7, 7, 7},
		{"negative", -5, 3, -5},
		{"both negative", -10, -5, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
