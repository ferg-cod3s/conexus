package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient("test-token")

	assert.NotNil(t, client)
	assert.Equal(t, "test-token", client.token)
	assert.NotNil(t, client.httpClient)
}

func TestHTTPClient_SearchIssues_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/search/issues")
		assert.Contains(t, r.URL.RawQuery, "type%3Aissue")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "59")
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"total_count": 1,
			"incomplete_results": false,
			"items": [
				{
					"id": 123,
					"number": 456,
					"title": "Test Issue",
					"body": "Test body",
					"html_url": "https://github.com/test/repo/issues/456",
					"state": "open",
					"user": {"login": "testuser", "id": 1, "html_url": "https://github.com/testuser"},
					"labels": [{"id": 1, "name": "bug", "color": "ff0000"}],
					"comments": 5,
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-02T00:00:00Z"
				}
			]
		}`)
	}))
	defer server.Close()

	client := NewHTTPClient("test-token")
	client.baseURL = server.URL

	issues, err := client.SearchIssues(context.Background(), "golang", 10)

	require.NoError(t, err)
	require.Len(t, issues, 1)

	issue := issues[0]
	assert.Equal(t, int64(123), issue.ID)
	assert.Equal(t, 456, issue.Number)
	assert.Equal(t, "Test Issue", issue.Title)
	assert.Equal(t, "Test body", issue.Body)
	assert.Equal(t, "https://github.com/test/repo/issues/456", issue.URL)
	assert.Equal(t, "open", issue.State)
	assert.Equal(t, "testuser", issue.Author)
	assert.Equal(t, []string{"bug"}, issue.Labels)
}

func TestHTTPClient_SearchIssues_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"message": "Bad credentials"}`)
	}))
	defer server.Close()

	client := NewHTTPClient("invalid-token")
	client.baseURL = server.URL

	_, err := client.SearchIssues(context.Background(), "golang", 10)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub API error")
}

func TestHTTPClient_SearchPullRequests_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.RawQuery, "type%3Apr")

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "59")
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"total_count": 1,
			"incomplete_results": false,
			"items": [
				{
					"id": 789,
					"number": 123,
					"title": "Test PR",
					"body": "PR body",
					"html_url": "https://github.com/test/repo/pull/123",
					"state": "open",
					"user": {"login": "prauthor", "id": 2, "html_url": "https://github.com/prauthor"},
					"labels": [{"id": 2, "name": "feature", "color": "00ff00"}],
					"comments": 3,
					"created_at": "2024-01-01T00:00:00Z",
					"updated_at": "2024-01-03T00:00:00Z"
				}
			]
		}`)
	}))
	defer server.Close()

	client := NewHTTPClient("test-token")
	client.baseURL = server.URL

	prs, err := client.SearchPullRequests(context.Background(), "golang", 10)

	require.NoError(t, err)
	require.Len(t, prs, 1)

	pr := prs[0]
	assert.Equal(t, int64(789), pr.ID)
	assert.Equal(t, 123, pr.Number)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, "PR body", pr.Body)
	assert.Equal(t, "https://github.com/test/repo/pull/123", pr.URL)
	assert.Equal(t, "open", pr.State)
	assert.Equal(t, "prauthor", pr.Author)
	assert.Equal(t, []string{"feature"}, pr.Labels)
}

func TestHTTPClient_CheckAuth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/user", r.URL.Path)
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"login": "testuser", "id": 1}`)
	}))
	defer server.Close()

	client := NewHTTPClient("test-token")
	client.baseURL = server.URL

	err := client.CheckAuth(context.Background())

	require.NoError(t, err)
}

func TestHTTPClient_CheckAuth_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewHTTPClient("invalid-token")
	client.baseURL = server.URL

	err := client.CheckAuth(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestHTTPClient_SetAuthHeaders(t *testing.T) {
	client := NewHTTPClient("test-token")

	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	client.setAuthHeaders(req)

	assert.Equal(t, "token test-token", req.Header.Get("Authorization"))
}

func TestHTTPClient_SetAuthHeaders_NoToken(t *testing.T) {
	client := NewHTTPClient("")

	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	client.setAuthHeaders(req)

	assert.Equal(t, "", req.Header.Get("Authorization"))
}

func TestHTTPClient_UpdateRateLimit(t *testing.T) {
	client := NewHTTPClient("test-token")

	// Use ResponseRecorder to properly initialize HTTP headers
	recorder := httptest.NewRecorder()
	recorder.Header().Set("X-RateLimit-Remaining", "45")
	recorder.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
	resp := recorder.Result()

	client.updateRateLimit(resp)

	assert.Equal(t, 45, client.rateLimiter.Remaining())
}

func TestCalculateIssueScore(t *testing.T) {
	tests := []struct {
		name     string
		item     GitHubSearchItem
		minScore float32
	}{
		{
			name: "no comments, old update",
			item: GitHubSearchItem{
				Comments:  0,
				UpdatedAt: time.Now().AddDate(-1, 0, 0),
			},
			minScore: 1.0,
		},
		{
			name: "with comments, recent update",
			item: GitHubSearchItem{
				Comments:  10,
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			minScore: 1.5,
		},
		{
			name: "max score scenario",
			item: GitHubSearchItem{
				Comments:  50,
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			minScore: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateIssueScore(tt.item)
			assert.Greater(t, score, float32(0.9))
		})
	}
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow server
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient("test-token")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.SearchIssues(ctx, "golang", 10)

	assert.Error(t, err)
}

func TestHTTPClient_RateLimitHeaders(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		remaining := 60 - callCount
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"total_count": 0,
			"incomplete_results": false,
			"items": []
		}`)
	}))
	defer server.Close()

	client := NewHTTPClient("test-token")
	client.baseURL = server.URL

	// First search
	_, _ = client.SearchIssues(context.Background(), "golang", 10)
	remaining1 := client.rateLimiter.Remaining()

	// Second search
	_, _ = client.SearchIssues(context.Background(), "golang", 10)
	remaining2 := client.rateLimiter.Remaining()

	// Rate limit should decrease
	assert.Greater(t, remaining1, remaining2)
}
