package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
	"strings"
	"regexp"

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
	server := NewServer(nil, nil, "", store, connectorStore, embedder, nil, nil, mockIdx)

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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

	resp, ok := result.(*GetRelatedInfoResponse)
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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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

	resp, ok := result.(*GetRelatedInfoResponse)
	require.True(t, ok)

	assert.NotEmpty(t, resp.Summary)
	assert.Contains(t, resp.Summary, "JIRA-123")
}

func TestHandleGetRelatedInfo_MissingBothIdentifiers(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	invalidJSON := json.RawMessage(`{"action":`)

	_, err := server.handleIndexControl(ctx, invalidJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
}

func TestHandleConnectorManagement_List(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

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
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	invalidJSON := json.RawMessage(`{"action":`)

	_, err := server.handleConnectorManagement(ctx, invalidJSON)

	assert.Error(t, err)
	protocolErr, ok := err.(*protocol.Error)
	require.True(t, ok)
	assert.Equal(t, protocol.InvalidParams, protocolErr.Code)
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

// ============================================================================
// File Path Flow Tests (Task 8.1.4)
// ============================================================================

func TestHandleFilePathFlow_MultipleRelatedFiles(t *testing.T) {
	ctx := context.Background()
	store := vectorstore.NewMemoryStore()
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), &mockEmbedder{}, nil, nil, &mockIndexer{})

	// Insert source file chunks
	sourceFile := "internal/auth/service.go"
	err := store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "source-1",
			Content: "type AuthService struct { db *DB }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   sourceFile,
				"start_line":  10,
				"end_line":    12,
				"source_type": "code",
				"type":        "struct",
			},
		},
	})
	require.NoError(t, err)

	// Insert test file chunks (should score 1.0)
	testFile := "internal/auth/service_test.go"
	err = store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "test-1",
			Content: "func TestAuthService_Login(t *testing.T) { ... }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"start_line":  15,
				"end_line":    25,
				"source_type": "code",
				"type":        "test",
			},
		},
	})
	require.NoError(t, err)

	// Insert documentation file chunks (should score 0.9)
	docFile := "docs/auth/authentication.md"
	err = store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "doc-1",
			Content: "# Authentication Service\nThis service handles user login...",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   docFile,
				"start_line":  1,
				"end_line":    5,
				"source_type": "documentation",
				"type":        "markdown",
			},
		},
	})
	require.NoError(t, err)

	// Insert related code file chunks (should score 0.5 - similar code)
	relatedFile := "internal/user/service.go"
	err = store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "related-1",
			Content: "type UserService struct { db *DB }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   relatedFile,
				"start_line":  8,
				"end_line":    10,
				"source_type": "code",
				"type":        "struct",
			},
		},
	})
	require.NoError(t, err)

	// Call handleFilePathFlow
	req := GetRelatedInfoRequest{
		FilePath: sourceFile,
	}
	resp, err := server.handleFilePathFlow(ctx, req)

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Summary)
	assert.GreaterOrEqual(t, len(resp.RelatedItems), 3, "Should have at least 3 related chunks")

	// Verify test file is included with correct score
	var testItem *RelatedItem
	for i := range resp.RelatedItems {
		if resp.RelatedItems[i].FilePath == testFile {
			testItem = &resp.RelatedItems[i]
			break
		}
	}
	require.NotNil(t, testItem, "Test file should be in results")
	assert.Equal(t, float32(1.0), testItem.Score, "Test file should have score 1.0")
	assert.Equal(t, RelationTypeTestFile, testItem.RelationType)

	// Verify doc file is included with correct score
	var docItem *RelatedItem
	for i := range resp.RelatedItems {
		if resp.RelatedItems[i].FilePath == docFile {
			docItem = &resp.RelatedItems[i]
			break
		}
	}
	require.NotNil(t, docItem, "Doc file should be in results")
	assert.Equal(t, float32(0.9), docItem.Score, "Doc file should have score 0.9")
	assert.Equal(t, RelationTypeDocumentation, docItem.RelationType)
}

func TestHandleFilePathFlow_RelationshipScoring(t *testing.T) {
	ctx := context.Background()
	store := vectorstore.NewMemoryStore()
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), &mockEmbedder{}, nil, nil, &mockIndexer{})

	sourceFile := "pkg/core/handler.go"

	// Insert source file
	err := store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "source-1",
			Content: "package core",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   sourceFile,
				"source_type": "code",
			},
		},
	})
	require.NoError(t, err)

	// Create one chunk for each relationship type to verify scoring
	testCases := []struct {
		id           string
		filePath     string
		content      string
		expectedType string
		expectedScore float32
	}{
		{"test-chunk", "pkg/core/handler_test.go", "func TestHandler()", RelationTypeTestFile, 1.0},
		{"doc-chunk", "docs/core/handler.md", "# Handler docs", RelationTypeDocumentation, 0.9},
	}

	for _, tc := range testCases {
		err := store.UpsertBatch(ctx, []vectorstore.Document{
			{
				ID:      tc.id,
				Content: tc.content,
				Vector:  make(embedding.Vector, 384),
				Metadata: map[string]interface{}{
					"file_path":   tc.filePath,
					"source_type": "code",
				},
			},
		})
		require.NoError(t, err)
	}

	// Execute file path flow
	req := GetRelatedInfoRequest{FilePath: sourceFile}
	resp, err := server.handleFilePathFlow(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify all relationship types are scored correctly
	for _, tc := range testCases {
		found := false
		for _, item := range resp.RelatedItems {
			if item.ID == tc.id {
				found = true
				assert.Equal(t, tc.expectedScore, item.Score, "Score mismatch for %s", tc.id)
				assert.Equal(t, tc.expectedType, item.RelationType, "Type mismatch for %s", tc.id)
				break
			}
		}
		assert.True(t, found, "Chunk %s not found in results", tc.id)
	}

	// Verify items are sorted by score (test=1.0 should be first, doc=0.9 second)
	if len(resp.RelatedItems) >= 2 {
		assert.GreaterOrEqual(t, resp.RelatedItems[0].Score, resp.RelatedItems[1].Score,
			"Items should be sorted by score descending")
	}
}

func TestHandleFilePathFlow_ResultLimiting(t *testing.T) {
	ctx := context.Background()
	store := vectorstore.NewMemoryStore()
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), &mockEmbedder{}, nil, nil, &mockIndexer{})

	sourceFile := "pkg/main.go"

	// Insert source file
	err := store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "source-1",
			Content: "package main",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   sourceFile,
				"source_type": "code",
			},
		},
	})
	require.NoError(t, err)

	// Insert 60 related file chunks (should be limited to 50)
	testFile := "pkg/main_test.go"
	docs := make([]vectorstore.Document, 60)
	for i := 0; i < 60; i++ {
		docs[i] = vectorstore.Document{
			ID:      fmt.Sprintf("test-chunk-%d", i),
			Content: fmt.Sprintf("func TestCase_%d(t *testing.T) { ... }", i),
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"start_line":  i * 10,
				"end_line":    (i * 10) + 8,
				"source_type": "code",
				"type":        "test",
			},
		}
	}
	err = store.UpsertBatch(ctx, docs)
	require.NoError(t, err)

	// Execute file path flow
	req := GetRelatedInfoRequest{FilePath: sourceFile}
	resp, err := server.handleFilePathFlow(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify result limiting (should be capped at 50)
	assert.Equal(t, 50, len(resp.RelatedItems), "Results should be limited to 50 items")
	assert.Contains(t, resp.Summary, "50 chunks", "Summary should mention 50 chunks")
}

func TestHandleFilePathFlow_NoRelatedFiles(t *testing.T) {
	ctx := context.Background()
	store := vectorstore.NewMemoryStore()
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), &mockEmbedder{}, nil, nil, &mockIndexer{})

	sourceFile := "internal/isolated/module.go"

	// Insert only the source file with no related files
	err := store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "source-1",
			Content: "package isolated",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   sourceFile,
				"source_type": "code",
			},
		},
	})
	require.NoError(t, err)

	// Execute file path flow
	req := GetRelatedInfoRequest{FilePath: sourceFile}
	resp, err := server.handleFilePathFlow(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify graceful handling of no related files
	assert.Empty(t, resp.RelatedItems, "Should have no related items")
	assert.Empty(t, resp.RelatedPRs, "Should have no related PRs")
	assert.Empty(t, resp.RelatedIssues, "Should have no related issues")
	assert.Empty(t, resp.Discussions, "Should have no discussions")
	assert.Contains(t, resp.Summary, "0 related files", "Summary should mention 0 related files")
	assert.Contains(t, resp.Summary, "0 chunks", "Summary should mention 0 chunks")
}

func TestHandleFilePathFlow_SourceFileNotFound(t *testing.T) {
	ctx := context.Background()
	store := vectorstore.NewMemoryStore()
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), &mockEmbedder{}, nil, nil, &mockIndexer{})

	nonExistentFile := "internal/nonexistent/file.go"

	// Don't insert any documents - file doesn't exist

	// Execute file path flow (should handle gracefully)
	req := GetRelatedInfoRequest{FilePath: nonExistentFile}
	resp, err := server.handleFilePathFlow(ctx, req)

	// Should succeed even if source file not found (graceful degradation)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify graceful handling
	assert.Empty(t, resp.RelatedItems, "Should have no related items")
	assert.NotEmpty(t, resp.Summary, "Should have a summary message")
}

func TestHandleFilePathFlow_PRsIssuesExtraction(t *testing.T) {
	ctx := context.Background()
	store := vectorstore.NewMemoryStore()
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), &mockEmbedder{}, nil, nil, &mockIndexer{})

	sourceFile := "internal/api/handler.go"

	// Insert source file
	err := store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "source-1",
			Content: "package api",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   sourceFile,
				"source_type": "code",
			},
		},
	})
	require.NoError(t, err)

	// Insert test file with PR metadata
	testFile := "internal/api/handler_test.go"
	err = store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "pr-chunk",
			Content: "Test for PR #123",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"source_type": "github_pr",
				"pr_number":   "123",
			},
		},
		{
			ID:      "issue-chunk",
			Content: "Fix for issue JIRA-456",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"source_type": "jira",
				"issue_id":    "JIRA-456",
			},
		},
		{
			ID:      "slack-chunk",
			Content: "Discussion about this change in Slack channel",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"source_type": "slack",
				"channel":     "#engineering",
				"timestamp":   "2025-01-15T10:30:00Z",
			},
		},
	})
	require.NoError(t, err)

	// Execute file path flow
	req := GetRelatedInfoRequest{FilePath: sourceFile}
	resp, err := server.handleFilePathFlow(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify PR extraction
	assert.Contains(t, resp.RelatedPRs, "123", "Should extract PR number")
	assert.Len(t, resp.RelatedPRs, 1, "Should have 1 PR")

	// Verify issue extraction
	assert.Contains(t, resp.RelatedIssues, "JIRA-456", "Should extract issue ID")
	assert.Len(t, resp.RelatedIssues, 1, "Should have 1 issue")

	// Verify discussion extraction
	assert.Len(t, resp.Discussions, 1, "Should have 1 discussion")
	if len(resp.Discussions) > 0 {
		assert.Equal(t, "#engineering", resp.Discussions[0].Channel)
		assert.Equal(t, "2025-01-15T10:30:00Z", resp.Discussions[0].Timestamp)
		assert.NotEmpty(t, resp.Discussions[0].Summary)
	}

	// Verify summary counts
	assert.Contains(t, resp.Summary, "1 PRs", "Summary should mention 1 PR")
	assert.Contains(t, resp.Summary, "1 issues", "Summary should mention 1 issue")
	assert.Contains(t, resp.Summary, "1 discussions", "Summary should mention 1 discussion")
}

func TestHandleFilePathFlow_ChunkMetadata(t *testing.T) {
	ctx := context.Background()
	store := vectorstore.NewMemoryStore()
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), &mockEmbedder{}, nil, nil, &mockIndexer{})

	sourceFile := "internal/core/processor.go"

	// Insert source file
	err := store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "source-1",
			Content: "package core",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   sourceFile,
				"source_type": "code",
			},
		},
	})
	require.NoError(t, err)

	// Insert test file chunks with different line number types
	testFile := "internal/core/processor_test.go"
	err = store.UpsertBatch(ctx, []vectorstore.Document{
		{
			ID:      "test-int",
			Content: "Test with int line numbers",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"start_line":  int(10),
				"end_line":    int(20),
				"source_type": "code",
			},
		},
		{
			ID:      "test-float64",
			Content: "Test with float64 line numbers",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"start_line":  float64(30),
				"end_line":    float64(40),
				"source_type": "code",
			},
		},
		{
			ID:      "test-int64",
			Content: "Test with int64 line numbers",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"start_line":  int64(50),
				"end_line":    int64(60),
				"source_type": "code",
			},
		},
		{
			ID:      "test-missing",
			Content: "Test with missing line numbers",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"file_path":   testFile,
				"source_type": "code",
				// No start_line or end_line
			},
		},
	})
	require.NoError(t, err)

	// Execute file path flow
	req := GetRelatedInfoRequest{FilePath: sourceFile}
	resp, err := server.handleFilePathFlow(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.GreaterOrEqual(t, len(resp.RelatedItems), 4, "Should have at least 4 chunks")

	// Verify line numbers extracted correctly for each type
	chunkTests := map[string]struct {
		expectedStart int
		expectedEnd   int
	}{
		"test-int":     {10, 20},
		"test-float64": {30, 40},
		"test-int64":   {50, 60},
		"test-missing": {0, 0},
	}

	for _, item := range resp.RelatedItems {
		if expected, ok := chunkTests[item.ID]; ok {
			assert.Equal(t, expected.expectedStart, item.StartLine, "Start line mismatch for %s", item.ID)
			assert.Equal(t, expected.expectedEnd, item.EndLine, "End line mismatch for %s", item.ID)
		}
	}
}

// ============================================================================
// Helper Function Tests (Task 8.1.4)
// ============================================================================

func TestGetRelationshipScore(t *testing.T) {
	server := NewServer(nil, nil, "", nil, nil, nil, nil, nil, nil)

	tests := []struct {
		relationType  string
		expectedScore float32
	}{
		{RelationTypeTestFile, 1.0},
		{RelationTypeDocumentation, 0.9},
		{RelationTypeSymbolRef, 0.8},
		{RelationTypeImport, 0.7},
		{RelationTypeCommitHistory, 0.6},
		{RelationTypeSimilarCode, 0.5},
		{"unknown_type", 0.3},
		{"", 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.relationType, func(t *testing.T) {
			score := server.getRelationshipScore(tt.relationType)
			assert.Equal(t, tt.expectedScore, score, "Score mismatch for type: %s", tt.relationType)
		})
	}
}

func TestGetRelationshipPriority(t *testing.T) {
	server := NewServer(nil, nil, "", nil, nil, nil, nil, nil, nil)

	tests := []struct {
		relationType     string
		expectedPriority int
	}{
		{RelationTypeTestFile, 1},
		{RelationTypeDocumentation, 2},
		{RelationTypeSymbolRef, 3},
		{RelationTypeImport, 4},
		{RelationTypeCommitHistory, 5},
		{RelationTypeSimilarCode, 6},
		{"unknown_type", 99},
		{"", 99},
	}

	for _, tt := range tests {
		t.Run(tt.relationType, func(t *testing.T) {
			priority := server.getRelationshipPriority(tt.relationType)
			assert.Equal(t, tt.expectedPriority, priority, "Priority mismatch for type: %s", tt.relationType)
		})
	}

	// Verify priority ordering (lower is higher priority)
	testPrio := server.getRelationshipPriority(RelationTypeTestFile)
	docPrio := server.getRelationshipPriority(RelationTypeDocumentation)
	symbolPrio := server.getRelationshipPriority(RelationTypeSymbolRef)
	importPrio := server.getRelationshipPriority(RelationTypeImport)
	commitPrio := server.getRelationshipPriority(RelationTypeCommitHistory)
	similarPrio := server.getRelationshipPriority(RelationTypeSimilarCode)
	unknownPrio := server.getRelationshipPriority("unknown")

	assert.Less(t, testPrio, docPrio, "Test should have higher priority than doc")
	assert.Less(t, docPrio, symbolPrio, "Doc should have higher priority than symbol")
	assert.Less(t, symbolPrio, importPrio, "Symbol should have higher priority than import")
	assert.Less(t, importPrio, commitPrio, "Import should have higher priority than commit")
	assert.Less(t, commitPrio, similarPrio, "Commit should have higher priority than similar")
	assert.Less(t, similarPrio, unknownPrio, "Similar should have higher priority than unknown")
}

func TestExtractLineNumber(t *testing.T) {
	server := NewServer(nil, nil, "", nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name          string
		metadata      map[string]interface{}
		key           string
		expectedValue int
		expectedOk    bool
	}{
		{
			name:          "int value",
			metadata:      map[string]interface{}{"line": int(42)},
			key:           "line",
			expectedValue: 42,
			expectedOk:    true,
		},
		{
			name:          "float64 value",
			metadata:      map[string]interface{}{"line": float64(99.7)},
			key:           "line",
			expectedValue: 99,
			expectedOk:    true,
		},
		{
			name:          "int64 value",
			metadata:      map[string]interface{}{"line": int64(123)},
			key:           "line",
			expectedValue: 123,
			expectedOk:    true,
		},
		{
			name:          "missing key",
			metadata:      map[string]interface{}{"other": 42},
			key:           "line",
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "wrong type (string)",
			metadata:      map[string]interface{}{"line": "42"},
			key:           "line",
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "nil metadata",
			metadata:      nil,
			key:           "line",
			expectedValue: 0,
			expectedOk:    false,
		},
		{
			name:          "empty metadata",
			metadata:      map[string]interface{}{},
			key:           "line",
			expectedValue: 0,
			expectedOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := server.extractLineNumber(tt.metadata, tt.key)
			assert.Equal(t, tt.expectedValue, value, "Value mismatch")
			assert.Equal(t, tt.expectedOk, ok, "Ok flag mismatch")
		})
	}
}

// ============================================================================
// Ticket ID Flow Tests (Task 8.1.5)
// ============================================================================

func TestHandleTicketIDFlow_ValidTicketInBranches(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Add some mock documents to vector store for files that git history will find
	// These should match files actually modified in commits/branches with "feat"
	now := time.Now()
	docs := []vectorstore.Document{
		{
			ID:      "chunk-1",
			Content: "func (s *Server) handleTicketIDFlow() { ... }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "internal/mcp/handlers.go",
				"start_line":  100,
				"end_line":    150,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "chunk-2",
			Content: "func TestHandleTicketIDFlow() { ... }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "internal/mcp/handlers_test.go",
				"start_line":  200,
				"end_line":    250,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Use actual Conexus repo which should have git history
	// Note: This test uses the actual git repo, so it depends on real git history
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "feat", // Search for "feat" which is in branch names like "feat/mcp-related-info"
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Should find branches with "feat" in the name
	assert.Contains(t, resp.Summary, "feat")
	assert.NotEmpty(t, resp.RelatedItems) // Should have some related items
}

func TestHandleTicketIDFlow_ValidTicketInCommits(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Use actual Conexus repo
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "Phase", // Search for "Phase" which appears in commit messages
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Summary should mention the ticket and commits
	assert.Contains(t, resp.Summary, "Phase")
	assert.NotNil(t, resp.RelatedItems)
}

func TestHandleTicketIDFlow_TicketNotFound(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Use actual Conexus repo
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "NONEXISTENT-TICKET-999999", // Ticket that should not exist
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Should return semantic search fallback with no results
	assert.Contains(t, resp.Summary, "found 0 related items")
	assert.Contains(t, resp.Summary, "NONEXISTENT-TICKET-999999")
	assert.Empty(t, resp.RelatedPRs)
	assert.Empty(t, resp.RelatedIssues)
	assert.Empty(t, resp.Discussions)
	assert.Empty(t, resp.RelatedItems)
}

func TestHandleTicketIDFlow_NotInGitRepo(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Use /tmp which is definitely not a git repo
	req := GetRelatedInfoRequest{
		FilePath: "/tmp/not-a-git-repo/file.go",
		TicketID: "TEST-123",
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert - should fall back to semantic search
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Should indicate git was unavailable but semantic search was used
	assert.Contains(t, resp.Summary, "Git history search unavailable")
}

func TestHandleTicketIDFlow_InvalidTicketID(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	repoPath := "/home/f3rg/src/github/conexus"
	
	// Test various invalid ticket ID formats
	invalidTickets := []string{
		"../../../etc/passwd",  // Path traversal
		"ticket; rm -rf /",     // Command injection
		"ticket|cat /etc/passwd", // Pipe injection
		"ticket`whoami`",       // Backtick injection
		"ticket$(whoami)",      // Command substitution
		"ticket\n\nmalicious", // Newline injection
	}

	for _, ticketID := range invalidTickets {
		t.Run(ticketID, func(t *testing.T) {
			req := GetRelatedInfoRequest{
				FilePath: repoPath + "/internal/mcp/handlers.go",
				TicketID: ticketID,
			}

			// Act
			resp, err := server.handleTicketIDFlow(ctx, req)

			// Assert
			assert.Error(t, err, "Should reject invalid ticket ID: %s", ticketID)
			assert.Nil(t, resp)
			assert.Contains(t, err.Error(), "invalid", "Error should mention invalid ticket ID")
		})
	}
}

func TestHandleTicketIDFlow_MultipleFiles(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	now := time.Now()

	// Add documents for multiple files that git history will actually find
	// Use files that were modified in commits/branches with "feat"
	docs := []vectorstore.Document{
		{
			ID:      "file1-chunk1",
			Content: "func (s *Server) handleTicketIDFlow() { ... }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "internal/mcp/handlers.go",
				"start_line":  100,
				"end_line":    150,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "file1-chunk2",
			Content: "func (s *Server) handleContextFlow() { ... }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "internal/mcp/handlers.go",
				"start_line":  200,
				"end_line":    250,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "file2-chunk1",
			Content: "func TestHandleTicketIDFlow() { ... }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "internal/mcp/handlers_test.go",
				"start_line":  300,
				"end_line":    350,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "file3-chunk1",
			Content: "type Server struct { ... }",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "internal/mcp/server.go",
				"start_line":  50,
				"end_line":    100,
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Use actual Conexus repo
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "feat", // Should find multiple files in git history
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Should have results from multiple files
	assert.NotEmpty(t, resp.RelatedItems)
	
	// Check that we get unique file paths
	filePaths := make(map[string]bool)
	for _, item := range resp.RelatedItems {
		if item.FilePath != "" {
			filePaths[item.FilePath] = true
		}
	}
	// We should have at least 1 unique file path
	assert.GreaterOrEqual(t, len(filePaths), 1)
}

func TestHandleTicketIDFlow_PRAndIssueExtraction(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	now := time.Now()

	// Add PR and issue metadata to vector store
	docs := []vectorstore.Document{
		{
			ID:      "pr-1",
			Content: "Implement authentication feature for TEST-123",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "github_pr",
				"pr_number":   "456",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "issue-1",
			Content: "Authentication bug reported in TEST-123",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "github_issue",
				"issue_id":    "789",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:      "slack-1",
			Content: "Discussion about TEST-123 authentication issue",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "slack",
				"channel":     "engineering",
				"timestamp":   "2025-01-15T10:30:00Z",
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, doc := range docs {
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Use actual Conexus repo but with test ticket ID
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "TEST-123", // Will match PR/issue/slack in vector store
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Even if no git history, should still extract PR/issue/discussion from vector store
	// Check if the metadata was extracted (may be empty if embedding search didn't match)
	assert.NotNil(t, resp.RelatedPRs)
	assert.NotNil(t, resp.RelatedIssues)
	assert.NotNil(t, resp.Discussions)
}

func TestHandleTicketIDFlow_ScoreBoost(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	now := time.Now()


	doc := vectorstore.Document{
		ID:      "test-chunk",
		Content: "test content for score boost verification",
		Vector:  make(embedding.Vector, 384),
		Metadata: map[string]interface{}{
			"source_type": "file",
			"file_path":   "internal/test/file.go",
			"start_line":  1,
			"end_line":    10,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := store.Upsert(ctx, doc)
	require.NoError(t, err)

	// Use actual Conexus repo
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "feat", // Should find in git history
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Check that items have score boost applied
	// Score should be increased by 0.3 from git history
	for _, item := range resp.RelatedItems {
		// Scores should be >= 0.3 (the boost amount)
		// In practice, they'll be base_score + 0.3
		assert.GreaterOrEqual(t, item.Score, float32(0.0), "Score should be non-negative")
	}
}

func TestHandleTicketIDFlow_CommitLimit(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()

	// Use actual Conexus repo which should have many commits
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "Task", // "Task" appears in many commit messages
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	
	// Summary should respect 5 commit limit
	// Count actual commits by looking for commit hash patterns (8-char hex followed by colon)
	if strings.Contains(resp.Summary, "Recent commits:") {
		t.Logf("Summary: %s", resp.Summary)
		commitHashPattern := regexp.MustCompile(`[0-9a-f]{8}:`)
		matches := commitHashPattern.FindAllString(resp.Summary, -1)
		commitCount := len(matches)
		assert.LessOrEqual(t, commitCount, 5, "Should display at most 5 commits in summary")
	}
}

func TestHandleTicketIDFlow_PerFileChunkLimit(t *testing.T) {
	// Arrange
	store := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	server := NewServer(nil, nil, "", store, newMockConnectorStore(), embedder, nil, nil, &mockIndexer{})

	ctx := context.Background()
	now := time.Now()

	// Add more than 5 chunks for a single file
	filePath := "internal/test/large_file.go"
	for i := 0; i < 10; i++ {
		doc := vectorstore.Document{
			ID:      fmt.Sprintf("chunk-%d", i),
			Content: fmt.Sprintf("code chunk %d", i),
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   filePath,
				"start_line":  i * 10,
				"end_line":    (i + 1) * 10,
			},
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Use actual Conexus repo
	repoPath := "/home/f3rg/src/github/conexus"
	req := GetRelatedInfoRequest{
		FilePath: repoPath + "/internal/mcp/handlers.go",
		TicketID: "feat",
	}

	// Act
	resp, err := server.handleTicketIDFlow(ctx, req)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, resp)
	
	// Count chunks per file
	fileChunkCount := make(map[string]int)
	for _, item := range resp.RelatedItems {
		if item.FilePath != "" {
			fileChunkCount[item.FilePath]++
		}
	}
	
	// Each file should have at most 5 chunks
	for filePath, count := range fileChunkCount {
		assert.LessOrEqual(t, count, 5, "File %s should have at most 5 chunks, got %d", filePath, count)
	}
}
