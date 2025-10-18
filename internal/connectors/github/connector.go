// Package github provides a GitHub connector for Conexus federation search.
package github

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/schema"
)

// Connector implements the SearchableConnector interface for GitHub.
type Connector struct {
	id     string
	client Client
	mu     sync.RWMutex
	closed bool
}

// NewConnector creates a new GitHub connector with the given ID and HTTP client.
func NewConnector(id string, client Client) *Connector {
	return &Connector{
		id:     id,
		client: client,
		closed: false,
	}
}

// GetID returns the connector's unique identifier.
func (c *Connector) GetID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.id
}

// GetType returns the connector type.
func (c *Connector) GetType() string {
	return "github"
}

// Search performs a search across GitHub issues and pull requests.
// It searches both issues and PRs in parallel and merges the results.
func (c *Connector) Search(ctx context.Context, req *schema.SearchRequest) ([]schema.SearchResultItem, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, fmt.Errorf("connector is closed")
	}
	c.mu.RUnlock()

	if req == nil {
		return nil, fmt.Errorf("search request cannot be nil")
	}

	if req.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Default pagination
	if req.TopK == 0 {
		req.TopK = 30
	}
	if req.TopK > 100 {
		req.TopK = 100 // GitHub API limit
	}

	// Search issues and PRs in parallel
	var wg sync.WaitGroup
	issuesChan := make(chan []schema.SearchResultItem, 1)
	prsChan := make(chan []schema.SearchResultItem, 1)
	errsChan := make(chan error, 2)

	wg.Add(2)

	// Search issues
	go func() {
		defer wg.Done()
		items, err := c.searchIssues(ctx, req)
		if err != nil {
			errsChan <- err
			issuesChan <- nil
			return
		}
		issuesChan <- items
	}()

	// Search pull requests
	go func() {
		defer wg.Done()
		items, err := c.searchPullRequests(ctx, req)
		if err != nil {
			errsChan <- err
			prsChan <- nil
			return
		}
		prsChan <- items
	}()

	wg.Wait()
	close(errsChan)
	close(issuesChan)
	close(prsChan)

	// Drain errors channel (non-blocking best-effort)
	// Errors are ignored to allow graceful degradation
	for err := range errsChan {
		_ = err // Explicitly mark as ignored for best-effort search
	}

	// Merge results from both searches
	issues := <-issuesChan
	prs := <-prsChan

	results := make([]schema.SearchResultItem, 0, len(issues)+len(prs))
	results = append(results, issues...)
	results = append(results, prs...)

	// Sort by score (descending) and limit to TopK
	sortByScore(results)
	if len(results) > req.TopK {
		results = results[:req.TopK]
	}

	return results, nil
}

// searchIssues searches GitHub issues matching the query.
func (c *Connector) searchIssues(ctx context.Context, req *schema.SearchRequest) ([]schema.SearchResultItem, error) {
	issues, err := c.client.SearchIssues(ctx, req.Query, req.TopK)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	results := make([]schema.SearchResultItem, 0, len(issues))
	for _, issue := range issues {
		item := schema.SearchResultItem{
			ID:         fmt.Sprintf("%d", issue.ID),
			Content:    fmt.Sprintf("%s\n%s", issue.Title, issue.Body),
			Score:      issue.Score,
			SourceType: "github_issue",
			Metadata: map[string]interface{}{
				"url":        issue.URL,
				"number":     issue.Number,
				"state":      issue.State,
				"created_at": issue.CreatedAt.Format(time.RFC3339),
				"updated_at": issue.UpdatedAt.Format(time.RFC3339),
				"author":     issue.Author,
				"labels":     issue.Labels,
			},
		}
		results = append(results, item)
	}

	return results, nil
}

// searchPullRequests searches GitHub pull requests matching the query.
func (c *Connector) searchPullRequests(ctx context.Context, req *schema.SearchRequest) ([]schema.SearchResultItem, error) {
	prs, err := c.client.SearchPullRequests(ctx, req.Query, req.TopK)
	if err != nil {
		return nil, fmt.Errorf("failed to search pull requests: %w", err)
	}

	results := make([]schema.SearchResultItem, 0, len(prs))
	for _, pr := range prs {
		item := schema.SearchResultItem{
			ID:         fmt.Sprintf("%d", pr.ID),
			Content:    fmt.Sprintf("%s\n%s", pr.Title, pr.Body),
			Score:      pr.Score,
			SourceType: "github_pr",
			Metadata: map[string]interface{}{
				"url":        pr.URL,
				"number":     pr.Number,
				"state":      pr.State,
				"created_at": pr.CreatedAt.Format(time.RFC3339),
				"updated_at": pr.UpdatedAt.Format(time.RFC3339),
				"author":     pr.Author,
				"labels":     pr.Labels,
			},
		}
		results = append(results, item)
	}

	return results, nil
}

// Health performs a health check on the GitHub connector by making a simple API call.
func (c *Connector) Health(ctx context.Context) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("connector is closed")
	}
	client := c.client
	c.mu.RUnlock()

	// Create a short context for health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Simple API call to verify connectivity
	return client.CheckAuth(healthCtx)
}

// Close closes the connector and releases any resources.
func (c *Connector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return nil
}

// sortByScore sorts results by score in descending order.
func sortByScore(items []schema.SearchResultItem) {
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].Score > items[i].Score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
