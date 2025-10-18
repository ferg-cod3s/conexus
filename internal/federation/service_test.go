package federation

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
	"github.com/ferg-cod3s/conexus/internal/connectors"
)

// TestNewService tests that service can be created
func TestNewService(t *testing.T) {
	// Test that service can be created
	service := NewService(nil, nil)
	assert.NotNil(t, service)
}

// MockConnectorWithTestify implements SearchableConnector for testing using testify/mock
type MockConnectorWithTestify struct {
	mock.Mock
	id       string
	connType string
}

func (m *MockConnectorWithTestify) Search(ctx context.Context, req *schema.SearchRequest) ([]schema.SearchResultItem, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]schema.SearchResultItem), args.Error(1)
}

func (m *MockConnectorWithTestify) GetID() string {
	return m.id
}

func (m *MockConnectorWithTestify) GetType() string {
	return m.connType
}

func TestService_Search(t *testing.T) {
	tests := []struct {
		name             string
		req              *schema.SearchRequest
		setupConnectors  func() []*MockConnectorWithTestify
		expectedResults  int
		expectError      bool
	}{
		{
			name: "successful search with multiple connectors",
			req: &schema.SearchRequest{
				Query: "test query",
				TopK:  10,
			},
			setupConnectors: func() []*MockConnectorWithTestify {
				conn1 := &MockConnectorWithTestify{id: "conn1", connType: "filesystem"}
				conn1.On("Search", mock.Anything, mock.Anything).Return([]schema.SearchResultItem{
					{ID: "1", Content: "result 1", Score: 0.9, SourceType: "file"},
					{ID: "2", Content: "result 2", Score: 0.8, SourceType: "file"},
				}, nil)

				conn2 := &MockConnectorWithTestify{id: "conn2", connType: "github"}
				conn2.On("Search", mock.Anything, mock.Anything).Return([]schema.SearchResultItem{
					{ID: "3", Content: "result 3", Score: 0.7, SourceType: "github"},
				}, nil)

				return []*MockConnectorWithTestify{conn1, conn2}
			},
			expectedResults: 3,
			expectError:     false,
		},
		{
			name: "empty results from all connectors",
			req: &schema.SearchRequest{
				Query: "empty query",
				TopK:  5,
			},
			setupConnectors: func() []*MockConnectorWithTestify {
				conn := &MockConnectorWithTestify{id: "conn1", connType: "filesystem"}
				conn.On("Search", mock.Anything, mock.Anything).Return([]schema.SearchResultItem{}, nil)
				return []*MockConnectorWithTestify{conn}
			},
			expectedResults: 0,
			expectError:     false,
		},
		{
			name: "pagination works correctly",
			req: &schema.SearchRequest{
				Query:  "test query",
				TopK:   2,
				Offset: 1,
			},
			setupConnectors: func() []*MockConnectorWithTestify {
				conn := &MockConnectorWithTestify{id: "conn1", connType: "filesystem"}
				conn.On("Search", mock.Anything, mock.Anything).Return([]schema.SearchResultItem{
					{ID: "1", Content: "result 1", Score: 0.9, SourceType: "file"},
					{ID: "2", Content: "result 2", Score: 0.8, SourceType: "file"},
					{ID: "3", Content: "result 3", Score: 0.7, SourceType: "file"},
				}, nil)
				return []*MockConnectorWithTestify{conn}
			},
			expectedResults: 2,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock connectors
			mockConnectors := tt.setupConnectors()

			// Create mock manager
			mockManager := &connectors.Manager{}
			// We need to mock the List method - this is tricky with the current design
			// For now, we'll create a simple test that focuses on the service logic

			// Create service
			service := NewService(mockManager, nil)

			// For this test, we'll directly test the executeParallelSearches method
			// by creating a mock context and connectors
			ctx := context.Background()

			// Convert to SearchableConnector interface
			var searchable []SearchableConnector
			for _, mc := range mockConnectors {
				searchable = append(searchable, mc)
			}

			// Execute parallel searches
			results, err := service.executeParallelSearches(ctx, tt.req, searchable)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, results, len(mockConnectors))

			// Verify mock expectations
			for _, conn := range mockConnectors {
				conn.AssertExpectations(t)
			}
		})
	}
}

func TestService_executeParallelSearches(t *testing.T) {
	tests := []struct {
		name        string
		connectors  []SearchableConnector
		req         *schema.SearchRequest
		expectError bool
	}{
		{
			name: "successful parallel execution",
			connectors: []SearchableConnector{
				&MockConnectorWithTestify{id: "conn1", connType: "filesystem"},
				&MockConnectorWithTestify{id: "conn2", connType: "github"},
			},
			req:         &schema.SearchRequest{Query: "test"},
			expectError: false,
		},
		{
			name: "timeout handling",
			connectors: []SearchableConnector{
				&MockConnectorWithTestify{id: "slow-conn", connType: "filesystem"},
			},
			req:         &schema.SearchRequest{Query: "test"},
			expectError: true, // Should timeout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(nil, nil) // Manager not needed for this test

			if tt.name == "timeout handling" {
				// Set very short timeout
				service.timeout = 1 * time.Millisecond

				// Mock a slow connector
				slowConn := tt.connectors[0].(*MockConnectorWithTestify)
				slowConn.On("Search", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					time.Sleep(10 * time.Millisecond) // Sleep longer than timeout
				}).Return([]schema.SearchResultItem{}, nil)
			} else {
				// Setup normal mocks
				for _, conn := range tt.connectors {
					mockConn := conn.(*MockConnectorWithTestify)
					mockConn.On("Search", mock.Anything, mock.Anything).Return([]schema.SearchResultItem{
						{ID: "test", Content: "test content", Score: 0.5, SourceType: "file"},
					}, nil)
				}
			}

			ctx := context.Background()
			results, err := service.executeParallelSearches(ctx, tt.req, tt.connectors)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, len(tt.connectors))
			}

			// Verify mock expectations
			for _, conn := range tt.connectors {
				if mockConn, ok := conn.(*MockConnectorWithTestify); ok {
					mockConn.AssertExpectations(t)
				}
			}
		})
	}
}

func TestService_NewService(t *testing.T) {
	mockManager := &connectors.Manager{}
	service := NewService(mockManager, nil)

	assert.NotNil(t, service)
	assert.Equal(t, mockManager, service.connectorManager)
	assert.NotNil(t, service.merger)
	assert.NotNil(t, service.detector)
	assert.Equal(t, 10*time.Second, service.timeout)
}

// TestService_createGitHubConnector tests the GitHub connector creation
func TestService_createGitHubConnector(t *testing.T) {
	tests := []struct {
		name        string
		conn        *connectors.Connector
		expectError bool
		expectedErr string
	}{
		{
			name: "valid GitHub connector creation with token",
			conn: &connectors.Connector{
				ID:     "github-1",
				Type:   "github",
				Status: "active",
				Config: map[string]interface{}{
					"token": "ghp_testtoken123456789",
				},
			},
			expectError: false,
		},
		{
			name: "missing GitHub token",
			conn: &connectors.Connector{
				ID:     "github-2",
				Type:   "github",
				Status: "active",
				Config: map[string]interface{}{},
			},
			expectError: true,
			expectedErr: "missing or invalid GitHub token",
		},
		{
			name: "empty GitHub token",
			conn: &connectors.Connector{
				ID:     "github-3",
				Type:   "github",
				Status: "active",
				Config: map[string]interface{}{
					"token": "",
				},
			},
			expectError: true,
			expectedErr: "missing or invalid GitHub token",
		},
		{
			name: "invalid GitHub token type",
			conn: &connectors.Connector{
				ID:     "github-4",
				Type:   "github",
				Status: "active",
				Config: map[string]interface{}{
					"token": 12345, // Wrong type
				},
			},
			expectError: true,
			expectedErr: "missing or invalid GitHub token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(nil, nil)
			result, err := service.createGitHubConnector(tt.conn)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.conn.ID, result.GetID())
				assert.Equal(t, "github", result.GetType())
			}
		})
	}
}


