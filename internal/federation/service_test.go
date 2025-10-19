package federation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockVectorStore is a mock implementation of VectorStore
type mockVectorStore struct {
	results []vectorstore.SearchResult
	err     error
}

func (m *mockVectorStore) Upsert(ctx context.Context, doc vectorstore.Document) error {
	return m.err
}

func (m *mockVectorStore) UpsertBatch(ctx context.Context, docs []vectorstore.Document) error {
	return m.err
}

func (m *mockVectorStore) Delete(ctx context.Context, id string) error {
	return m.err
}

func (m *mockVectorStore) Get(ctx context.Context, id string) (*vectorstore.Document, error) {
	return nil, m.err
}

func (m *mockVectorStore) SearchVector(ctx context.Context, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	return m.results, m.err
}

func (m *mockVectorStore) SearchBM25(ctx context.Context, query string, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	limit := opts.Limit
	if limit > 0 && len(m.results) > limit {
		return m.results[:limit], nil
	}
	return m.results, nil
}

func (m *mockVectorStore) SearchHybrid(ctx context.Context, query string, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	return m.results, m.err
}

func (m *mockVectorStore) Count(ctx context.Context) (int64, error) {
	return int64(len(m.results)), m.err
}

func (m *mockVectorStore) ListIndexedFiles(ctx context.Context) ([]string, error) {
	return []string{}, m.err
}

func (m *mockVectorStore) GetFileChunks(ctx context.Context, filePath string) ([]vectorstore.Document, error) {
	return []vectorstore.Document{}, m.err
}

func (m *mockVectorStore) Close() error {
	return nil
}

// mockConnectorStore is a mock implementation of ConnectorStore
type mockConnectorStore struct {
	connectors map[string]*connectors.Connector
	err        error
}

func (m *mockConnectorStore) Add(ctx context.Context, connector *connectors.Connector) error {
	if m.err != nil {
		return m.err
	}
	m.connectors[connector.ID] = connector
	return nil
}

func (m *mockConnectorStore) Get(ctx context.Context, id string) (*connectors.Connector, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.connectors[id]
	if !ok {
		return nil, fmt.Errorf("connector not found")
	}
	return c, nil
}

func (m *mockConnectorStore) List(ctx context.Context) ([]*connectors.Connector, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*connectors.Connector, 0, len(m.connectors))
	for _, c := range m.connectors {
		result = append(result, c)
	}
	return result, nil
}

func (m *mockConnectorStore) Update(ctx context.Context, id string, connector *connectors.Connector) error {
	if m.err != nil {
		return m.err
	}
	m.connectors[id] = connector
	return nil
}

func (m *mockConnectorStore) Remove(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.connectors, id)
	return nil
}

func (m *mockConnectorStore) Close() error {
	return nil
}

// TestNewService tests service initialization
func TestNewService(t *testing.T) {
	vs := &mockVectorStore{}
	mgr := &connectors.Manager{}
	svc := NewService(mgr, vs, 5*time.Second)

	require.NotNil(t, svc)
	assert.Equal(t, 5*time.Second, svc.timeout)
}

// TestNewService_DefaultTimeout tests service with default timeout
func TestNewService_DefaultTimeout(t *testing.T) {
	vs := &mockVectorStore{}
	mgr := &connectors.Manager{}
	svc := NewService(mgr, vs, 0)

	require.NotNil(t, svc)
	assert.Equal(t, 30*time.Second, svc.timeout)
}

// TestQueryMultipleSources_EmptyConnectors tests query with no active connectors
func TestQueryMultipleSources_EmptyConnectors(t *testing.T) {
	vs := &mockVectorStore{}
	mgr := connectors.NewManager(&mockConnectorStore{connectors: make(map[string]*connectors.Connector)})
	svc := NewService(mgr, vs, 5*time.Second)

	result, err := svc.QueryMultipleSources(context.Background(), "test query")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result.Items))
	assert.Equal(t, 0, len(result.SourceCounts))
}

// TestQueryMultipleSources_EmptyQuery tests query with empty string
func TestQueryMultipleSources_EmptyQuery(t *testing.T) {
	vs := &mockVectorStore{}
	mgr := connectors.NewManager(&mockConnectorStore{connectors: make(map[string]*connectors.Connector)})
	svc := NewService(mgr, vs, 5*time.Second)

	_, err := svc.QueryMultipleSources(context.Background(), "")

	assert.Error(t, err)
}

// TestQueryMultipleSources_SingleConnector tests query with one connector
func TestQueryMultipleSources_SingleConnector(t *testing.T) {
	vs := &mockVectorStore{
		results: []vectorstore.SearchResult{
			{
				Document: vectorstore.Document{
					ID:      "doc1",
					Content: "test content",
					Metadata: map[string]interface{}{
						"file_path": "/test/file.txt",
					},
				},
				Score:  0.95,
				Method: "bm25",
			},
		},
	}
	
	store := &mockConnectorStore{connectors: make(map[string]*connectors.Connector)}
	store.connectors["conn1"] = &connectors.Connector{
		ID:   "conn1",
		Type: "local-files",
		Name: "Test Connector",
	}
	mgr := connectors.NewManager(store)
	require.NoError(t, mgr.Initialize(context.Background(), store.connectors["conn1"]))
	svc := NewService(mgr, vs, 5*time.Second)

	result, err := svc.QueryMultipleSources(context.Background(), "test query")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result.Items), 0)
}

// TestQueryMultipleSources_Timeout tests query timeout handling
func TestQueryMultipleSources_Timeout(t *testing.T) {
	vs := &mockVectorStore{}
	store := &mockConnectorStore{connectors: make(map[string]*connectors.Connector)}
	store.connectors["conn1"] = &connectors.Connector{
		ID:   "conn1",
		Type: "external-api",
		Name: "External Connector",
	}
	mgr := connectors.NewManager(store)
	svc := NewService(mgr, vs, 1*time.Millisecond) // Very short timeout

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := svc.QueryMultipleSources(ctx, "test query")

	// Should either succeed with empty results or handle timeout gracefully
	require.NoError(t, err)
	require.NotNil(t, result)
}

// TestQueryMultipleSources_MultipleConnectors tests query with multiple connectors
func TestQueryMultipleSources_MultipleConnectors(t *testing.T) {
	vs := &mockVectorStore{
		results: []vectorstore.SearchResult{
			{
				Document: vectorstore.Document{
					ID:      "doc1",
					Content: "first result",
					Metadata: map[string]interface{}{
						"file_path": "/file1.txt",
					},
				},
				Score:  0.9,
				Method: "bm25",
			},
			{
				Document: vectorstore.Document{
					ID:      "doc2",
					Content: "second result",
					Metadata: map[string]interface{}{
						"file_path": "/file2.txt",
					},
				},
				Score:  0.85,
				Method: "bm25",
			},
		},
	}

	store := &mockConnectorStore{connectors: make(map[string]*connectors.Connector)}
	store.connectors["conn1"] = &connectors.Connector{
		ID:   "conn1",
		Type: "local-files",
		Name: "Local Files",
	}
	store.connectors["conn2"] = &connectors.Connector{
		ID:   "conn2",
		Type: "local-files",
		Name: "Local Files 2",
	}
	mgr := connectors.NewManager(store)
	require.NoError(t, mgr.Initialize(context.Background(), store.connectors["conn1"]))
	require.NoError(t, mgr.Initialize(context.Background(), store.connectors["conn2"]))
	svc := NewService(mgr, vs, 5*time.Second)

	result, err := svc.QueryMultipleSources(context.Background(), "test query")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.TotalDuration, time.Duration(0))
}

// TestQueryMultipleSources_WithErrors tests handling of connector errors
func TestQueryMultipleSources_WithErrors(t *testing.T) {
	vs := &mockVectorStore{
		err: fmt.Errorf("vectorstore error"),
	}

	store := &mockConnectorStore{connectors: make(map[string]*connectors.Connector)}
	store.connectors["conn1"] = &connectors.Connector{
		ID:   "conn1",
		Type: "local-files",
		Name: "Local Files",
	}
	mgr := connectors.NewManager(store)
	require.NoError(t, mgr.Initialize(context.Background(), store.connectors["conn1"]))
	svc := NewService(mgr, vs, 5*time.Second)

	result, err := svc.QueryMultipleSources(context.Background(), "test query")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, len(result.Errors), 0)
}

// TestQueryMultipleSources_Deduplication tests result deduplication
func TestQueryMultipleSources_Deduplication(t *testing.T) {
	// Create duplicate results across connectors
	vs := &mockVectorStore{
		results: []vectorstore.SearchResult{
			{
				Document: vectorstore.Document{
					ID:      "same-id",
					Content: "duplicate content",
					Metadata: map[string]interface{}{
						"file_path": "/dup.txt",
					},
				},
				Score:  0.9,
				Method: "bm25",
			},
		},
	}

	store := &mockConnectorStore{connectors: make(map[string]*connectors.Connector)}
	store.connectors["conn1"] = &connectors.Connector{
		ID:   "conn1",
		Type: "local-files",
		Name: "Local Files 1",
	}
	store.connectors["conn2"] = &connectors.Connector{
		ID:   "conn2",
		Type: "local-files",
		Name: "Local Files 2",
	}
	mgr := connectors.NewManager(store)
	svc := NewService(mgr, vs, 5*time.Second)

	result, err := svc.QueryMultipleSources(context.Background(), "test query")

	require.NoError(t, err)
	assert.NotNil(t, result.DeduplicationStats)
}

// TestQueryMultipleSources_CrossSourceRelationships tests relationship detection
func TestQueryMultipleSources_CrossSourceRelationships(t *testing.T) {
	vs := &mockVectorStore{
		results: []vectorstore.SearchResult{
			{
				Document: vectorstore.Document{
					ID:      "issue-123",
					Content: "GitHub issue",
					Metadata: map[string]interface{}{
						"file_path": "/issues/123",
					},
				},
				Score:  0.9,
				Method: "bm25",
			},
		},
	}

	store := &mockConnectorStore{connectors: make(map[string]*connectors.Connector)}
	store.connectors["github"] = &connectors.Connector{
		ID:   "github",
		Type: "local-files",
		Name: "GitHub Issues",
	}
	mgr := connectors.NewManager(store)
	require.NoError(t, mgr.Initialize(context.Background(), store.connectors["github"]))
	svc := NewService(mgr, vs, 5*time.Second)

	result, err := svc.QueryMultipleSources(context.Background(), "test query")

	require.NoError(t, err)
	assert.NotNil(t, result.CrosSourceLinks)
}

// TestExecuteQuery_LocalFiles tests local-files connector query execution
func TestExecuteQuery_LocalFiles(t *testing.T) {
	vs := &mockVectorStore{
		results: []vectorstore.SearchResult{
			{
				Document: vectorstore.Document{
					ID:      "doc1",
					Content: "test",
					Metadata: map[string]interface{}{
						"file_path": "/test.txt",
					},
				},
				Score:  0.95,
				Method: "bm25",
			},
		},
	}

	svc := &Service{vectorstore: vs}
	conn := &connectors.Connector{
		ID:   "local",
		Type: "local-files",
	}

	items, err := svc.executeQuery(context.Background(), conn, "test")

	require.NoError(t, err)
	assert.Greater(t, len(items), 0)
}

// TestExecuteQuery_NonLocalFiles tests non-local-files connector
func TestExecuteQuery_NonLocalFiles(t *testing.T) {
	vs := &mockVectorStore{}
	svc := &Service{vectorstore: vs}
	conn := &connectors.Connector{
		ID:   "external",
		Type: "external-api",
	}

	items, err := svc.executeQuery(context.Background(), conn, "test")

	require.NoError(t, err)
	assert.Equal(t, 0, len(items))
}

// TestExecuteQuery_VectorstoreError tests handling of vectorstore errors
func TestExecuteQuery_VectorstoreError(t *testing.T) {
	vs := &mockVectorStore{
		err: fmt.Errorf("connection failed"),
	}

	svc := &Service{vectorstore: vs}
	conn := &connectors.Connector{
		ID:   "local",
		Type: "local-files",
	}

	_, err := svc.executeQuery(context.Background(), conn, "test")

	assert.Error(t, err)
}

// TestMergeResults_Empty tests merging empty results
func TestMergeResults_Empty(t *testing.T) {
	svc := &Service{}
	result := svc.mergeResults([]*QueryResult{})

	require.NotNil(t, result)
	assert.Equal(t, 0, len(result.Items))
	assert.Equal(t, 0, len(result.SourceCounts))
}

// TestMergeResults_WithErrors tests merging results with errors
func TestMergeResults_WithErrors(t *testing.T) {
	svc := &Service{}
	results := []*QueryResult{
		{
			Source: "source1",
			Error:  fmt.Errorf("test error"),
		},
		{
			Source: "source2",
			Items:  []interface{}{map[string]interface{}{"id": "item1"}},
		},
	}

	result := svc.mergeResults(results)

	assert.Greater(t, len(result.Errors), 0)
	assert.Greater(t, len(result.Items), 0)
}

// TestDetectRelationships tests relationship detection
func TestDetectRelationships(t *testing.T) {
	svc := &Service{}
	results := []*QueryResult{
		{
			Source: "source1",
			Items:  []interface{}{map[string]interface{}{"id": "item1"}},
		},
		{
			Source: "source2",
			Items:  []interface{}{map[string]interface{}{"id": "item1"}},
		},
	}

	relationships := svc.detectRelationships(results, nil)

	require.NotNil(t, relationships)
}

// TestIntegration_CompleteFlow tests complete query flow
func TestIntegration_CompleteFlow(t *testing.T) {
	vs := &mockVectorStore{
		results: []vectorstore.SearchResult{
			{
				Document: vectorstore.Document{
					ID:      "result1",
					Content: "test content",
					Metadata: map[string]interface{}{
						"file_path": "/test.txt",
					},
				},
				Score:  0.95,
				Method: "bm25",
			},
		},
	}

	store := &mockConnectorStore{connectors: make(map[string]*connectors.Connector)}
	store.connectors["local"] = &connectors.Connector{
		ID:   "local",
		Type: "local-files",
		Name: "Local Files",
	}
	mgr := connectors.NewManager(store)
	require.NoError(t, mgr.Initialize(context.Background(), store.connectors["local"]))
	svc := NewService(mgr, vs, 5*time.Second)

	result, err := svc.QueryMultipleSources(context.Background(), "integration test")

	require.NoError(t, err)
	assert.NotNil(t, result.DeduplicationStats)
	assert.NotNil(t, result.CrosSourceLinks)
	assert.Greater(t, result.TotalDuration, time.Duration(0))
}
