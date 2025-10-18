package github

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClient is a mock implementation of Client for testing
type MockClient struct {
	SearchIssuesFunc        func(ctx context.Context, query string, limit int) ([]Issue, error)
	SearchPullRequestsFunc  func(ctx context.Context, query string, limit int) ([]PullRequest, error)
	CheckAuthFunc           func(ctx context.Context) error
}

func (m *MockClient) SearchIssues(ctx context.Context, query string, limit int) ([]Issue, error) {
	if m.SearchIssuesFunc != nil {
		return m.SearchIssuesFunc(ctx, query, limit)
	}
	return nil, nil
}

func (m *MockClient) SearchPullRequests(ctx context.Context, query string, limit int) ([]PullRequest, error) {
	if m.SearchPullRequestsFunc != nil {
		return m.SearchPullRequestsFunc(ctx, query, limit)
	}
	return nil, nil
}

func (m *MockClient) CheckAuth(ctx context.Context) error {
	if m.CheckAuthFunc != nil {
		return m.CheckAuthFunc(ctx)
	}
	return nil
}

func TestNewConnector(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("github-1", client)

	require.NotNil(t, connector)
	assert.Equal(t, "github-1", connector.GetID())
	assert.Equal(t, "github", connector.GetType())
}

func TestConnector_GetID(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("test-connector", client)

	assert.Equal(t, "test-connector", connector.GetID())
}

func TestConnector_GetType(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("github-1", client)

	assert.Equal(t, "github", connector.GetType())
}

func TestConnector_Search_EmptyQuery(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("github-1", client)

	req := &schema.SearchRequest{Query: ""}
	_, err := connector.Search(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search query cannot be empty")
}

func TestConnector_Search_NilRequest(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("github-1", client)

	_, err := connector.Search(context.Background(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "search request cannot be nil")
}

func TestConnector_Search_DefaultLimit(t *testing.T) {
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return []Issue{
				{
					ID:     1,
					Number: 1,
					Title:  "Test Issue",
					Body:   "Test body",
					Score:  1.0,
				},
			}, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return nil, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test"}

	results, err := connector.Search(context.Background(), req)

	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestConnector_Search_MaxLimit(t *testing.T) {
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			// Verify limit is capped at 100
			assert.Equal(t, 100, limit)
			return nil, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			assert.Equal(t, 100, limit)
			return nil, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test", TopK: 200}

	_, err := connector.Search(context.Background(), req)

	require.NoError(t, err)
}

func TestConnector_Search_MergesResults(t *testing.T) {
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return []Issue{
				{
					ID:     1,
					Number: 1,
					Title:  "Issue 1",
					Body:   "Body 1",
					Score:  2.0,
				},
			}, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return []PullRequest{
				{
					ID:     2,
					Number: 2,
					Title:  "PR 1",
					Body:   "PR Body 1",
					Score:  1.0,
				},
			}, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test"}

	results, err := connector.Search(context.Background(), req)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "github_issue", results[0].SourceType)
	assert.Equal(t, "github_pr", results[1].SourceType)
}

func TestConnector_Search_LimitResults(t *testing.T) {
	issues := make([]Issue, 50)
	for i := 0; i < 50; i++ {
		issues[i] = Issue{
			ID:     int64(i),
			Number: i,
			Title:  fmt.Sprintf("Issue %d", i),
			Score:  float32(i) * 0.1,
		}
	}

	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return issues, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return nil, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test", TopK: 10}

	results, err := connector.Search(context.Background(), req)

	require.NoError(t, err)
	assert.Len(t, results, 10)
}

func TestConnector_Search_SortsByScore(t *testing.T) {
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return []Issue{
				{
					ID:     1,
					Number: 1,
					Title:  "Low score",
					Score:  1.0,
				},
				{
					ID:     3,
					Number: 3,
					Title:  "High score",
					Score:  3.0,
				},
				{
					ID:     2,
					Number: 2,
					Title:  "Medium score",
					Score:  2.0,
				},
			}, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return nil, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test"}

	results, err := connector.Search(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, float32(3.0), results[0].Score)
	assert.Equal(t, float32(2.0), results[1].Score)
	assert.Equal(t, float32(1.0), results[2].Score)
}

func TestConnector_Search_Closed(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("github-1", client)
	connector.Close()

	req := &schema.SearchRequest{Query: "test"}
	_, err := connector.Search(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connector is closed")
}

func TestConnector_Health_Success(t *testing.T) {
	client := &MockClient{
		CheckAuthFunc: func(ctx context.Context) error {
			return nil
		},
	}

	connector := NewConnector("github-1", client)
	err := connector.Health(context.Background())

	require.NoError(t, err)
}

func TestConnector_Health_Failure(t *testing.T) {
	client := &MockClient{
		CheckAuthFunc: func(ctx context.Context) error {
			return fmt.Errorf("auth failed")
		},
	}

	connector := NewConnector("github-1", client)
	err := connector.Health(context.Background())

	assert.Error(t, err)
}

func TestConnector_Health_Closed(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("github-1", client)
	connector.Close()

	err := connector.Health(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connector is closed")
}

func TestConnector_Close(t *testing.T) {
	client := &MockClient{}
	connector := NewConnector("github-1", client)

	err := connector.Close()
	require.NoError(t, err)

	// Second close should not error
	err = connector.Close()
	require.NoError(t, err)
}

func TestConnector_IssuesToSearchResultItems(t *testing.T) {
	now := time.Now()
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return []Issue{
				{
					ID:        1,
					Number:    10,
					Title:     "Test Issue",
					Body:      "Test Body",
					URL:       "https://github.com/test/repo/issues/10",
					State:     "open",
					Author:    "testuser",
					Labels:    []string{"bug", "urgent"},
					Score:     2.5,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return nil, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test"}

	results, err := connector.Search(context.Background(), req)

	require.NoError(t, err)
	require.Len(t, results, 1)

	item := results[0]
	assert.Equal(t, "1", item.ID)
	assert.Equal(t, "github_issue", item.SourceType)
	assert.Equal(t, float32(2.5), item.Score)
	assert.Contains(t, item.Content, "Test Issue")
	assert.Contains(t, item.Content, "Test Body")

	// Check metadata
	assert.Equal(t, "https://github.com/test/repo/issues/10", item.Metadata["url"])
	assert.Equal(t, 10, item.Metadata["number"])
	assert.Equal(t, "open", item.Metadata["state"])
	assert.Equal(t, "testuser", item.Metadata["author"])
	assert.Equal(t, []string{"bug", "urgent"}, item.Metadata["labels"])
}

func TestConnector_PRsToSearchResultItems(t *testing.T) {
	now := time.Now()
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return nil, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return []PullRequest{
				{
					ID:        2,
					Number:    20,
					Title:     "Test PR",
					Body:      "PR Body",
					URL:       "https://github.com/test/repo/pull/20",
					State:     "open",
					Author:    "prauthor",
					Labels:    []string{"review"},
					Score:     1.8,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test"}

	results, err := connector.Search(context.Background(), req)

	require.NoError(t, err)
	require.Len(t, results, 1)

	item := results[0]
	assert.Equal(t, "2", item.ID)
	assert.Equal(t, "github_pr", item.SourceType)
	assert.Equal(t, float32(1.8), item.Score)
	assert.Contains(t, item.Content, "Test PR")
	assert.Contains(t, item.Content, "PR Body")

	// Check metadata
	assert.Equal(t, "https://github.com/test/repo/pull/20", item.Metadata["url"])
	assert.Equal(t, 20, item.Metadata["number"])
	assert.Equal(t, "open", item.Metadata["state"])
	assert.Equal(t, "prauthor", item.Metadata["author"])
}

func TestConnector_IssueSearchError(t *testing.T) {
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return nil, fmt.Errorf("search error")
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return nil, nil
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test"}

	_, err := connector.Search(context.Background(), req)

	// Should not fail when one search succeeds
	require.NoError(t, err)
}

func TestConnector_PRSearchError(t *testing.T) {
	client := &MockClient{
		SearchIssuesFunc: func(ctx context.Context, query string, limit int) ([]Issue, error) {
			return nil, nil
		},
		SearchPullRequestsFunc: func(ctx context.Context, query string, limit int) ([]PullRequest, error) {
			return nil, fmt.Errorf("search error")
		},
	}

	connector := NewConnector("github-1", client)
	req := &schema.SearchRequest{Query: "test"}

	_, err := connector.Search(context.Background(), req)

	// Should not fail when one search succeeds
	require.NoError(t, err)
}

func TestSortByScore(t *testing.T) {
	tests := []struct {
		name     string
		items    []schema.SearchResultItem
		expected []float32
	}{
		{
			name: "already sorted",
			items: []schema.SearchResultItem{
				{ID: "1", Score: 3.0},
				{ID: "2", Score: 2.0},
				{ID: "3", Score: 1.0},
			},
			expected: []float32{3.0, 2.0, 1.0},
		},
		{
			name: "reverse sorted",
			items: []schema.SearchResultItem{
				{ID: "1", Score: 1.0},
				{ID: "2", Score: 2.0},
				{ID: "3", Score: 3.0},
			},
			expected: []float32{3.0, 2.0, 1.0},
		},
		{
			name: "random order",
			items: []schema.SearchResultItem{
				{ID: "1", Score: 2.0},
				{ID: "2", Score: 1.0},
				{ID: "3", Score: 3.0},
			},
			expected: []float32{3.0, 2.0, 1.0},
		},
		{
			name:     "empty",
			items:    []schema.SearchResultItem{},
			expected: []float32{},
		},
		{
			name: "single item",
			items: []schema.SearchResultItem{
				{ID: "1", Score: 1.0},
			},
			expected: []float32{1.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortByScore(tt.items)
			for i, item := range tt.items {
				assert.Equal(t, tt.expected[i], item.Score)
			}
		})
	}
}
