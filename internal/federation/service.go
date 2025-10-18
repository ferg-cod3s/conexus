// Package federation implements multi-source result federation for search queries.
package federation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/connectors/github"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/schema"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/embedding"
)

// SearchableConnector defines the interface for connectors that support search operations
type SearchableConnector interface {
	// Search performs a search operation and returns results
	Search(ctx context.Context, req *schema.SearchRequest) ([]schema.SearchResultItem, error)
	// GetID returns the connector's unique identifier
	GetID() string
	// GetType returns the connector type
	GetType() string
}

// Service coordinates search operations across multiple connectors
type Service struct {
	connectorManager *connectors.Manager
	vectorStore      vectorstore.VectorStore
	merger           *Merger
}

// NewService creates a new federation service
func NewService(connectorManager *connectors.Manager, vectorStore vectorstore.VectorStore) *Service {
	return &Service{
		connectorManager: connectorManager,
		vectorStore:      vectorStore,
		merger:           NewMerger(),
	}
}

// Search executes a federated search across all available connectors
func (s *Service) Search(ctx context.Context, req *schema.SearchRequest, embedder embedding.Embedder) (*schema.SearchResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("search request cannot be nil")
	}

	if req.Query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	startTime := time.Now()

	// Default TopK
	if req.TopK == 0 {
		req.TopK = 10
	}

	// Discover active searchable connectors
	searchableConnectors, err := s.discoverSearchableConnectors(ctx, embedder)
	if err != nil {
		return nil, fmt.Errorf("discover connectors: %w", err)
	}

	if len(searchableConnectors) == 0 {
		return &schema.SearchResponse{
			Query:      req.Query,
			Results:    []schema.SearchResultItem{},
			QueryTime:  0,
			Offset:     0,
			Limit:      int32(req.TopK),
			HasMore:    false,
		}, nil
	}

	// Execute searches in parallel across connectors
	connectorResults, err := s.executeParallelSearches(ctx, req, searchableConnectors)
	if err != nil {
		return nil, fmt.Errorf("execute parallel searches: %w", err)
	}

	// Merge and rank results
	mergedResults := s.merger.Merge(connectorResults)

	// Apply pagination
	offset := int(req.Offset)
	topK := int(req.TopK)
	hasMore := false

	if offset > len(mergedResults) {
		offset = len(mergedResults)
		mergedResults = []schema.SearchResultItem{}
	} else {
		if offset+topK < len(mergedResults) {
			hasMore = true
		}
		mergedResults = mergedResults[offset : offset+topK]
	}

	queryTime := time.Since(startTime).Milliseconds()

	// Record metrics
	if logger := observability.FromContext(ctx); logger != nil {
		logger.WithField("query_time_ms", queryTime).
			WithField("result_count", len(mergedResults)).
			WithField("connector_count", len(searchableConnectors)).
			Info("Federated search completed")
	}

	return &schema.SearchResponse{
		Query:      req.Query,
		Results:    mergedResults,
		QueryTime:  queryTime,
		Offset:     offset,
		Limit:      topK,
		HasMore:    hasMore,
	}, nil
}

// discoverSearchableConnectors finds all active connectors that support search operations
func (s *Service) discoverSearchableConnectors(ctx context.Context, embedder embedding.Embedder) ([]SearchableConnector, error) {
	// Get all active connectors from manager
	connectors, err := s.connectorManager.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}

	var searchableConnectors []SearchableConnector
	for _, conn := range connectors {
		// Only include active connectors
		if conn.Status != "active" {
			continue
		}

		// Create searchable connector based on type
		var searchableConn SearchableConnector
		switch conn.Type {
		case "filesystem":
			searchableConn = NewFilesystemConnector(s.vectorStore, embedder)
		case "github":
			// Create GitHub connector from configuration
			githubConn, err := createGitHubConnector(conn)
			if err != nil {
				// Log error and continue with other connectors
				if logger := observability.FromContext(ctx); logger != nil {
					logger.WithError(err).
						WithField("connector_id", conn.ID).
						Warn("Failed to create GitHub connector")
				}
				continue
			}
			searchableConn = githubConn
		default:
			// Skip unsupported connector types
			continue
		}

		if searchableConn != nil {
			searchableConnectors = append(searchableConnectors, searchableConn)
		}
	}

	return searchableConnectors, nil
}

// createGitHubConnector creates a GitHub connector from a Connector configuration
func createGitHubConnector(conn *connectors.Connector) (SearchableConnector, error) {
	if conn == nil {
		return nil, fmt.Errorf("connector cannot be nil")
	}

	// Extract GitHub token from config
	token, ok := conn.Config["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("GitHub connector requires 'token' in config")
	}

	// Create GitHub client with token
	client := github.NewClient(token)

	// Create GitHub connector
	gitHubConnector := github.NewConnector(conn.ID, client)

	return gitHubConnector, nil
}

// ConnectorResult holds search results from a single connector
type ConnectorResult struct {
	ConnectorID   string
	ConnectorType string
	Results       []schema.SearchResultItem
}

// executeParallelSearches runs searches across multiple connectors concurrently
func (s *Service) executeParallelSearches(ctx context.Context, req *schema.SearchRequest, connectors []SearchableConnector) ([]ConnectorResult, error) {
	// Create context with timeout
	searchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resultsChan := make(chan ConnectorResult, len(connectors))
	errChan := make(chan error, len(connectors))
	var wg sync.WaitGroup

	// Launch search goroutines for each connector
	for _, conn := range connectors {
		wg.Add(1)
		go func(searchableConn SearchableConnector) {
			defer wg.Done()

			results, err := searchableConn.Search(searchCtx, req)
			if err != nil {
				errChan <- fmt.Errorf("search failed for connector %s: %w", searchableConn.GetID(), err)
				return
			}

			resultsChan <- ConnectorResult{
				ConnectorID:   searchableConn.GetID(),
				ConnectorType: searchableConn.GetType(),
				Results:       results,
			}
		}(conn)
	}

	// Wait for all searches to complete or context to cancel
	wg.Wait()
	close(resultsChan)
	close(errChan)

	// Collect results
	var connectorResults []ConnectorResult
	for result := range resultsChan {
		connectorResults = append(connectorResults, result)
	}

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// If all searches failed, return error
	if len(connectorResults) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("all connector searches failed: %v", errs)
	}

	return connectorResults, nil
}
