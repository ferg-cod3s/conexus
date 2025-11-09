package federation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// QueryResult represents results from a single source
type QueryResult struct {
	Source       string
	Items        []interface{}
	Error        error
	Duration     time.Duration
	ItemCount    int
	DeduplicateID string
}

// FederatedResult represents merged results from multiple sources
type FederatedResult struct {
	Items              []interface{}
	SourceCounts       map[string]int
	DeduplicationStats DeduplicationStats
	CrosSourceLinks    map[string][]string // entity ID -> list of IDs in other sources
	TotalDuration      time.Duration
	Errors             []error
	SourceAttributions map[string]map[string]interface{} // item ID -> source metadata
}

// DeduplicationStats tracks deduplication metrics
type DeduplicationStats struct {
	TotalResults    int
	DuplicatesFound int
	UniqueResults   int
	MergedResults   int
}

// Service provides multi-source query capabilities
type Service struct {
	manager     *connectors.Manager
	vectorstore vectorstore.VectorStore
	timeout     time.Duration
	mu          sync.RWMutex
	metrics     *MetricsCollector
	logger      *Logger
}

// NewService creates a new federation service
func NewService(manager *connectors.Manager, vs vectorstore.VectorStore, timeout time.Duration) *Service {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	return &Service{
		manager:     manager,
		vectorstore: vs,
		timeout:     timeout,
		metrics:     NewMetricsCollector(),
		logger:      NewLogger(context.Background()),
	}
}

// GetMetrics returns current metrics snapshot
func (s *Service) GetMetrics() MetricsSnapshot {
	return s.metrics.GetMetrics()
}

// ResetMetrics resets all collected metrics
func (s *Service) ResetMetrics() {
	s.metrics.Reset()
}

// QueryMultipleSources executes a query across all active connectors
func (s *Service) QueryMultipleSources(ctx context.Context, query string) (*FederatedResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Generate query ID for tracing
	queryID := fmt.Sprintf("query_%d", time.Now().UnixNano())

	// Record query start
	s.metrics.RecordQueryStart()
	s.logger.LogQueryStart(queryID, query, 0)

	start := time.Now()

	// Get all active connectors
	conns := s.manager.ListActive()
	s.logger.LogQueryStart(queryID, query, len(conns))

	if len(conns) == 0 {
		result := &FederatedResult{
			Items:              []interface{}{},
			SourceCounts:       make(map[string]int),
			SourceAttributions: make(map[string]map[string]interface{}),
		}
		totalDuration := time.Since(start)
		result.TotalDuration = totalDuration

		s.metrics.RecordQueryEnd(totalDuration.Nanoseconds(), []string{}, true)
		s.logger.LogQueryEnd(queryID, totalDuration, 0, 0, 0, true)

		return result, nil
	}

	// Execute queries in parallel
	resultsChan := make(chan *QueryResult, len(conns))
	wg := sync.WaitGroup{}

	for _, conn := range conns {
		wg.Add(1)
		go func(c *connectors.Connector) {
			defer wg.Done()
			result := s.queryConnector(ctx, queryID, c, query)
			resultsChan <- result
		}(conn)
	}

	wg.Wait()
	close(resultsChan)

	// Collect all results
	var results []*QueryResult
	var errorCount int
	var sourcesQueried []string

	for result := range resultsChan {
		results = append(results, result)
		sourcesQueried = append(sourcesQueried, result.Source)
		if result.Error != nil {
			errorCount++
		}
	}

	totalDuration := time.Since(start)

	// Merge and deduplicate results
	mergeStart := time.Now()
	fedResult := s.mergeResults(results)
	mergeDuration := time.Since(mergeStart)
	fedResult.TotalDuration = totalDuration

	s.metrics.RecordMergeOperation(mergeDuration.Nanoseconds())
	s.logger.LogMergeEnd(mergeDuration, fedResult.DeduplicationStats)

	// Detect cross-source relationships
	relStart := time.Now()
	fedResult.CrosSourceLinks = s.detectRelationships(results, fedResult.Items)
	relDuration := time.Since(relStart)

	s.metrics.RecordRelationshipDetection(len(fedResult.CrosSourceLinks))
	s.logger.LogRelationshipDetection(len(fedResult.CrosSourceLinks), relDuration)

	// Record query completion
	success := errorCount == 0
	s.metrics.RecordQueryEnd(totalDuration.Nanoseconds(), sourcesQueried, success)
	s.logger.LogQueryEnd(queryID, totalDuration, len(fedResult.Items), errorCount,
		len(fedResult.CrosSourceLinks), success)

	return fedResult, nil
}

// queryConnector queries a single connector with timeout
func (s *Service) queryConnector(ctx context.Context, queryID string, conn *connectors.Connector, query string) *QueryResult {
	result := &QueryResult{
		Source: conn.ID,
	}

	// Create context with timeout
	queryCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	start := time.Now()

	// For now, we'll simulate querying the connector
	// In a real implementation, this would call the connector's query method
	items, err := s.executeQuery(queryCtx, conn, query)

	duration := time.Since(start)
	result.Duration = duration
	result.Items = items
	result.Error = err
	result.ItemCount = len(items)

	// Log source query
	s.logger.LogSourceQuery(queryID, conn.ID, duration, len(items), err)

	// Record source query metrics
	success := err == nil
	s.metrics.RecordSourceQuery(conn.ID, duration.Nanoseconds(), success, len(items))

	if err != nil {
		s.metrics.RecordError("query_error")
	}

	return result
}

// executeQuery executes a query against a connector
// This is a placeholder - real implementation would use connector-specific APIs
func (s *Service) executeQuery(ctx context.Context, conn *connectors.Connector, query string) ([]interface{}, error) {
	// For local-files connector, we can query the vectorstore
	if conn.Type == "local-files" {
		// Query vectorstore for matching documents using BM25
		opts := vectorstore.SearchOptions{
			Limit: 10,
		}
		searchResults, err := s.vectorstore.SearchBM25(ctx, query, opts)
		if err != nil {
			return nil, fmt.Errorf("vectorstore query failed: %w", err)
		}

		items := make([]interface{}, len(searchResults))
		for i, sr := range searchResults {
			items[i] = map[string]interface{}{
				"id":        sr.Document.ID,
				"content":   sr.Document.Content,
				"file_path": sr.Document.Metadata["file_path"],
				"score":     sr.Score,
			}
		}
		return items, nil
	}

	// For other connector types, return empty for now
	// Real implementation would use connector-specific APIs
	return []interface{}{}, nil
}

// mergeResults merges results from multiple sources
func (s *Service) mergeResults(results []*QueryResult) *FederatedResult {
	fedResult := &FederatedResult{
		Items:              []interface{}{},
		SourceCounts:       make(map[string]int),
		SourceAttributions: make(map[string]map[string]interface{}),
		Errors:             []error{},
	}

	// Use merger to deduplicate and merge
	merger := NewMerger()
	for _, result := range results {
		if result.Error != nil {
			fedResult.Errors = append(fedResult.Errors, result.Error)
			continue
		}

		fedResult.SourceCounts[result.Source] = result.ItemCount
		merger.AddResults(result.Source, result.Items)
	}

	// Get merged and deduplicated results
	mergedItems, stats := merger.MergeAndDeduplicate()
	fedResult.Items = mergedItems
	fedResult.DeduplicationStats = stats
	fedResult.SourceAttributions = merger.GetSourceAttributions()

	// Record deduplication stats in metrics
	s.metrics.RecordDeduplicationStats(stats)

	return fedResult
}

// detectRelationships detects cross-source relationships
func (s *Service) detectRelationships(results []*QueryResult, mergedItems []interface{}) map[string][]string {
	detector := NewDetector()
	return detector.DetectRelationships(results, mergedItems)
}
