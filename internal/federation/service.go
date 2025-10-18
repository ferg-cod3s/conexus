// Package federation implements multi-source result federation for search queries.
package federation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
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
	detector         *Detector
	timeout          time.Duration
}
func NewService(manager *connectors.Manager, vectorStore vectorstore.VectorStore) *Service {
	return &Service{
		connectorManager: manager,
		vectorStore:      vectorStore,
		merger:           NewMerger(),
		detector:         NewDetector(),
		timeout:          10 * time.Second, // Default timeout
	}
}

// Search performs a federated search across all active searchable connectors
// Search performs a federated search across all active searchable connectors
func (s *Service) Search(ctx context.Context, req *schema.SearchRequest, embedder embedding.Embedder) (*schema.SearchResponse, error) {
	startTime := time.Now()

	// Discover active searchable connectors
	searchableConnectors, err := s.discoverSearchableConnectors(ctx, embedder)
	if err != nil {
		return nil, fmt.Errorf("failed to discover connectors: %w", err)
	}

	// If no connectors found, fall back to filesystem
	if len(searchableConnectors) == 0 {
		filesystemConn := NewFilesystemConnector(s.vectorStore, embedder)
		searchableConnectors = []SearchableConnector{filesystemConn}
	}

	// Execute parallel searches
	connectorResults, err := s.executeParallelSearches(ctx, req, searchableConnectors)
	if err != nil {
		return nil, fmt.Errorf("parallel search execution failed: %w", err)
	}

	// Merge results from all connectors
	mergedResults := s.merger.Merge(connectorResults)

	// Apply pagination
	topK := req.TopK
	if topK <= 0 {
		topK = 20
	}
	if topK > 100 {
		topK = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// Apply offset and limit
	var paginatedResults []schema.SearchResultItem
	hasMore := false
	if offset < len(mergedResults) {
		end := offset + topK
		// HasMore is true if there are results beyond this page
		if end < len(mergedResults) {
			hasMore = true
		}
		if end > len(mergedResults) {
			end = len(mergedResults)
		}
		paginatedResults = mergedResults[offset:end]
	}

	queryTime := float64(time.Since(startTime).Milliseconds())

	return &schema.SearchResponse{
		Results:    paginatedResults,
		TotalCount: len(mergedResults),
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

// ConnectorResult holds search results from a single connector
type ConnectorResult struct {
	ConnectorID   string
	ConnectorType string
	Results       []schema.SearchResultItem
}

// executeParallelSearches runs searches across multiple connectors concurrently
func (s *Service) executeParallelSearches(ctx context.Context, req *schema.SearchRequest, connectors []SearchableConnector) ([]ConnectorResult, error) {
	// Create context with timeout
	searchCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Channel for results
	resultChan := make(chan ConnectorResult, len(connectors))
	errorChan := make(chan error, len(connectors))

	// Launch goroutines for each connector
	var wg sync.WaitGroup
	for _, conn := range connectors {
		wg.Add(1)
		go func(connector SearchableConnector) {
			defer wg.Done()
			results, err := connector.Search(searchCtx, req)
			if err != nil {
				errorChan <- fmt.Errorf("connector %s (%s): %w", connector.GetID(), connector.GetType(), err)
				return
			}
			resultChan <- ConnectorResult{
				ConnectorID:   connector.GetID(),
				ConnectorType: connector.GetType(),
				Results:       results,
			}
		}(conn)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()
	// Collect results
	var allResults []ConnectorResult
	var errors []error

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				allResults = append(allResults, result)
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else {
				errors = append(errors, err)
			}
		case <-searchCtx.Done():
			return nil, fmt.Errorf("search timeout after %v", s.timeout)
		}

		if resultChan == nil && errorChan == nil {
			break
		}
	}

	// Log errors but don't fail the entire search
	if len(errors) > 0 {
		// In a real implementation, this would be logged
		// For now, we'll just continue with successful results
	}

	return allResults, nil

}

// FilesystemConnector implements SearchableConnector using the vector store
type FilesystemConnector struct {
	vectorStore vectorstore.VectorStore
	embedder    embedding.Embedder
}

// NewFilesystemConnector creates a new filesystem connector
func NewFilesystemConnector(vectorStore vectorstore.VectorStore, embedder embedding.Embedder) *FilesystemConnector {
	return &FilesystemConnector{
		vectorStore: vectorStore,
		embedder:    embedder,
	}
}

// Search performs a search using the vector store
func (f *FilesystemConnector) Search(ctx context.Context, req *schema.SearchRequest) ([]schema.SearchResultItem, error) {
	// Generate query embedding
	embedding, err := f.embedder.Embed(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Prepare search options
	topK := req.TopK
	if topK <= 0 {
		topK = 20
	}
	if topK > 100 {
		topK = 100
	}

	opts := vectorstore.SearchOptions{
		Limit:   req.Offset + topK + 1, // Request enough for pagination + 1 to detect more
		Offset:  0, // Connectors return all results, pagination at federation level
		Filters: make(map[string]interface{}),
	}

	// Apply filters
	if req.Filters != nil {
		if len(req.Filters.SourceTypes) > 0 {
			opts.Filters["source_types"] = req.Filters.SourceTypes
		}
		if req.Filters.DateRange != nil {
			opts.Filters["date_range"] = map[string]string{
				"from": req.Filters.DateRange.From,
				"to":   req.Filters.DateRange.To,
			}
		}
		// Apply work context filters
		if req.Filters.WorkContext != nil {
			if req.Filters.WorkContext.ActiveFile != "" {
				opts.Filters["related_files"] = req.Filters.WorkContext.ActiveFile
			}
			if req.Filters.WorkContext.GitBranch != "" {
				opts.Filters["git_branch"] = req.Filters.WorkContext.GitBranch
			}
			if len(req.Filters.WorkContext.OpenTicketIDs) > 0 {
				opts.Filters["ticket_ids"] = req.Filters.WorkContext.OpenTicketIDs
			}
		}
	}

	// Apply work context from request (overrides filter)
	if req.WorkContext != nil {
		if req.WorkContext.ActiveFile != "" {
			opts.Filters["boost_file"] = req.WorkContext.ActiveFile
		}
		if req.WorkContext.GitBranch != "" {
			opts.Filters["git_branch"] = req.WorkContext.GitBranch
		}
		if len(req.WorkContext.OpenTicketIDs) > 0 {
			opts.Filters["boost_tickets"] = req.WorkContext.OpenTicketIDs
		}
	}

	// Perform hybrid search
	results, err := f.vectorStore.SearchHybrid(ctx, req.Query, embedding.Vector, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results to SearchResultItem format
	searchResults := make([]schema.SearchResultItem, 0, len(results))
	for _, r := range results {
		// Extract source type from metadata
		sourceType := "file" // default
		if st, ok := r.Document.Metadata["source_type"].(string); ok {
			sourceType = st
		}

		searchResults = append(searchResults, schema.SearchResultItem{
			ID:         r.Document.ID,
			Content:    r.Document.Content,
			Score:      r.Score,
			SourceType: sourceType,
			Metadata:   r.Document.Metadata,
		})
	}

	return searchResults, nil
}

// GetID returns the connector ID
func (f *FilesystemConnector) GetID() string {
	return "filesystem"
}

// GetType returns the connector type
func (f *FilesystemConnector) GetType() string {
	return "filesystem"
}
