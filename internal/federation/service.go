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
	detector         *Detector
	metrics          *observability.FederationMetrics
	timeout          time.Duration
}

// NewService creates a new federation service without metrics
func NewService(manager *connectors.Manager, vectorStore vectorstore.VectorStore) *Service {
	return NewServiceWithMetrics(manager, vectorStore, nil)
}

// NewServiceWithMetrics creates a new federation service with optional metrics
func NewServiceWithMetrics(manager *connectors.Manager, vectorStore vectorstore.VectorStore, metrics *observability.FederationMetrics) *Service {
	return &Service{
		connectorManager: manager,
		vectorStore:      vectorStore,
		merger:           NewMerger(),
		detector:         NewDetector(),
		metrics:          metrics,
		timeout:          10 * time.Second, // Default timeout
	}
}

// Search performs a federated search across all active searchable connectors
func (s *Service) Search(ctx context.Context, req *schema.SearchRequest, embedder embedding.Embedder) (*schema.SearchResponse, error) {
	startTime := time.Now()
	status := "success"
	defer func() {
		if s.metrics != nil {
			duration := time.Since(startTime)
			s.metrics.RecordFederationSearch(status, duration, 0)
		}
	}()

	// Discover active searchable connectors
	searchableConnectors, err := s.discoverSearchableConnectors(ctx, embedder)
	if err != nil {
		status = "error"
		return nil, fmt.Errorf("failed to discover connectors: %w", err)
	}

	if s.metrics != nil {
		s.metrics.UpdateActiveConnectors(len(searchableConnectors))
	}

	// If no connectors found, fall back to filesystem
	if len(searchableConnectors) == 0 {
		filesystemConn := NewFilesystemConnector(s.vectorStore, embedder)
		searchableConnectors = []SearchableConnector{filesystemConn}
	}

	// Execute parallel searches
	connectorResults, err := s.executeParallelSearches(ctx, req, searchableConnectors)
	if err != nil {
		status = "error"
		return nil, fmt.Errorf("parallel search execution failed: %w", err)
	}

	// Record merged results before merge
	totalBeforeMerge := 0
	for _, cr := range connectorResults {
		totalBeforeMerge += len(cr.Results)
	}
	if s.metrics != nil {
		s.metrics.RecordMergedResults("before_merge", totalBeforeMerge)
	}

	// Merge results from all connectors
	mergeStart := time.Now()
	mergedResults := s.merger.Merge(connectorResults)
	if s.metrics != nil {
		s.metrics.RecordMergeDuration(time.Since(mergeStart))
		s.metrics.RecordMergedResults("after_merge", len(mergedResults))

		// Calculate deduplication ratio
		if totalBeforeMerge > 0 {
			dedupRatio := float64(totalBeforeMerge-len(mergedResults)) / float64(totalBeforeMerge)
			s.metrics.RecordDeduplicationRatio(dedupRatio)
		}
	}

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

	// Record pagination and final results
	if s.metrics != nil {
		s.metrics.RecordPaginationOperation(fmt.Sprintf("%d", topK))
		s.metrics.RecordMergedResults("after_pagination", len(paginatedResults))
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
		case "github":
			var err error
			searchableConn, err = s.createGitHubConnector(conn)
			if err != nil {
				// Failed to create GitHub connector, skip it
				continue
			}
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

// createGitHubConnector creates a GitHub connector from a connector configuration
func (s *Service) createGitHubConnector(conn *connectors.Connector) (SearchableConnector, error) {
	// Extract GitHub token from config
	token, ok := conn.Config["token"].(string)
	if !ok || token == "" {
		return nil, fmt.Errorf("missing or invalid GitHub token in connector config")
	}

	// Create GitHub HTTP client
	client := github.NewHTTPClient(token)

	// Create and return GitHub connector
	gitHubConnector := github.NewConnector(conn.ID, client)
	if gitHubConnector == nil {
		return nil, fmt.Errorf("failed to instantiate GitHub connector")
	}

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
	searchCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Channel for results
	resultChan := make(chan ConnectorResult, len(connectors))
	errorChan := make(chan error, len(connectors))

	// Launch goroutines for each connector
	var wg sync.WaitGroup
	startTimes := make(map[string]time.Time)
	mu := sync.Mutex{}

	for _, conn := range connectors {
		wg.Add(1)
		go func(connector SearchableConnector) {
			defer wg.Done()

			connectorStartTime := time.Now()
			mu.Lock()
			startTimes[connector.GetID()] = connectorStartTime
			mu.Unlock()

			results, err := connector.Search(searchCtx, req)

			connectorDuration := time.Since(connectorStartTime)
			connectorStatus := "success"
			if err != nil {
				connectorStatus = "error"
				if s.metrics != nil {
					s.metrics.RecordConnectorError(connector.GetID(), connector.GetType(), "search_error")
				}
				errorChan <- fmt.Errorf("connector %s (%s): %w", connector.GetID(), connector.GetType(), err)
				return
			}

			if s.metrics != nil {
				s.metrics.RecordConnectorSearch(connector.GetID(), connector.GetType(), connectorStatus, connectorDuration, len(results))
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
			// Record timeouts
			if s.metrics != nil {
				for _, conn := range connectors {
					s.metrics.RecordConnectorTimeout(conn.GetID(), conn.GetType())
				}
			}
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

	// Calculate and record parallel execution efficiency
	if s.metrics != nil && len(allResults) > 0 {
		totalSequentialTime := float64(0)
		maxParallelTime := float64(0)

		for _, result := range allResults {
			if startTime, ok := startTimes[result.ConnectorID]; ok {
				duration := float64(time.Since(startTime).Milliseconds())
				totalSequentialTime += duration
				if duration > maxParallelTime {
					maxParallelTime = duration
				}
			}
		}

		if maxParallelTime > 0 && totalSequentialTime > 0 {
			efficiency := (totalSequentialTime / float64(len(connectors))) / maxParallelTime
			if efficiency > 1.0 {
				efficiency = 1.0
			}
			s.metrics.UpdateParallelExecutionEfficiency(efficiency)
		}
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
		Offset:  0,                      // Connectors return all results, pagination at federation level
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
