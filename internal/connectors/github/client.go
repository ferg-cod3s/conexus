package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// GitHubAPIBaseURL is the GitHub REST API base URL
	GitHubAPIBaseURL = "https://api.github.com"

	// GitHubSearchIssuesEndpoint is the endpoint for searching issues
	GitHubSearchIssuesEndpoint = "/search/issues"

	// GitHubRateLimitHeader is the rate limit header
	GitHubRateLimitHeader = "X-RateLimit-Remaining"

	// GitHubRateLimitResetHeader is the rate limit reset header
	GitHubRateLimitResetHeader = "X-RateLimit-Reset"
)

// Issue represents a GitHub issue
type Issue struct {
	ID        int64
	Number    int
	Title     string
	Body      string
	URL       string
	State     string
	Author    string
	Labels    []string
	Score     float32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID        int64
	Number    int
	Title     string
	Body      string
	URL       string
	State     string
	Author    string
	Labels    []string
	Score     float32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Client defines the interface for GitHub API operations
type Client interface {
	// SearchIssues searches for GitHub issues
	SearchIssues(ctx context.Context, query string, limit int) ([]Issue, error)

	// SearchPullRequests searches for GitHub pull requests
	SearchPullRequests(ctx context.Context, query string, limit int) ([]PullRequest, error)

	// CheckAuth verifies authentication and connectivity
	CheckAuth(ctx context.Context) error
}

// HTTPClient implements the Client interface using GitHub's REST API
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
	rateLimiter *RateLimiter
}

// NewHTTPClient creates a new HTTP-based GitHub client
func NewHTTPClient(token string) *HTTPClient {
	return &HTTPClient{
		baseURL: GitHubAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token:       token,
		rateLimiter: NewRateLimiter(),
	}
}

// SearchIssues searches for GitHub issues matching the query
func (c *HTTPClient) SearchIssues(ctx context.Context, query string, limit int) ([]Issue, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build search query for issues only
	issueQuery := fmt.Sprintf("type:issue %s", query)
	url := fmt.Sprintf("%s%s?q=%s&per_page=%d&sort=updated&order=desc",
		c.baseURL, GitHubSearchIssuesEndpoint, url.QueryEscape(issueQuery), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter
	c.updateRateLimit(resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var searchResp GitHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert GitHub issues to our Issue format
	issues := make([]Issue, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		labels := make([]string, 0, len(item.Labels))
		for _, label := range item.Labels {
			labels = append(labels, label.Name)
		}

		// Calculate relevance score (simple heuristic)
		score := calculateIssueScore(item)

		issue := Issue{
			ID:        item.ID,
			Number:    item.Number,
			Title:     item.Title,
			Body:      item.Body,
			URL:       item.HTMLURL,
			State:     item.State,
			Author:    item.User.Login,
			Labels:    labels,
			Score:     score,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		issues = append(issues, issue)
	}

	return issues, nil
}

// SearchPullRequests searches for GitHub pull requests matching the query
func (c *HTTPClient) SearchPullRequests(ctx context.Context, query string, limit int) ([]PullRequest, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Build search query for PRs only
	prQuery := fmt.Sprintf("type:pr %s", query)
	url := fmt.Sprintf("%s%s?q=%s&per_page=%d&sort=updated&order=desc",
		c.baseURL, GitHubSearchIssuesEndpoint, url.QueryEscape(prQuery), limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limiter
	c.updateRateLimit(resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var searchResp GitHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert GitHub PRs to our PullRequest format
	prs := make([]PullRequest, 0, len(searchResp.Items))
	for _, item := range searchResp.Items {
		labels := make([]string, 0, len(item.Labels))
		for _, label := range item.Labels {
			labels = append(labels, label.Name)
		}

		// Calculate relevance score
		score := calculateIssueScore(item) // Same scoring for PRs

		pr := PullRequest{
			ID:        item.ID,
			Number:    item.Number,
			Title:     item.Title,
			Body:      item.Body,
			URL:       item.HTMLURL,
			State:     item.State,
			Author:    item.User.Login,
			Labels:    labels,
			Score:     score,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

// CheckAuth verifies authentication by making a simple API call
func (c *HTTPClient) CheckAuth(ctx context.Context) error {
	url := fmt.Sprintf("%s/user", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setAuthHeaders(req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	return nil
}

// setAuthHeaders adds authentication headers to the request
func (c *HTTPClient) setAuthHeaders(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", c.token))
	}
}

// updateRateLimit updates the rate limiter based on response headers
func (c *HTTPClient) updateRateLimit(resp *http.Response) {
	remaining := resp.Header.Get(GitHubRateLimitHeader)
	resetStr := resp.Header.Get(GitHubRateLimitResetHeader)

	if remaining != "" && resetStr != "" {
		c.rateLimiter.Update(remaining, resetStr)
	}
}

// calculateIssueScore calculates a relevance score for an issue/PR
// Simple heuristic: more recent updates and comments = higher score
func calculateIssueScore(item GitHubSearchItem) float32 {
	score := float32(1.0)

	// Boost score for items with comments (indicates activity/relevance)
	if item.Comments > 0 {
		score += float32(item.Comments) * 0.1
	}

	// Boost score for items updated recently
	daysSinceUpdate := time.Since(item.UpdatedAt).Hours() / 24
	if daysSinceUpdate < 30 {
		score += float32(30-daysSinceUpdate) * 0.01
	}

	return score
}

// GitHubSearchResponse represents the GitHub Search API response
type GitHubSearchResponse struct {
	TotalCount        int                 `json:"total_count"`
	IncompleteResults bool                `json:"incomplete_results"`
	Items             []GitHubSearchItem  `json:"items"`
}

// GitHubSearchItem represents a single item in GitHub search results
type GitHubSearchItem struct {
	ID        int64           `json:"id"`
	Number    int             `json:"number"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	HTMLURL   string          `json:"html_url"`
	State     string          `json:"state"`
	User      GitHubUser      `json:"user"`
	Labels    []GitHubLabel   `json:"labels"`
	Comments  int             `json:"comments"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login   string `json:"login"`
	ID      int64  `json:"id"`
	HTMLURL string `json:"html_url"`
}

// GitHubLabel represents a GitHub label
type GitHubLabel struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}
