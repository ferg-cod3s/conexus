package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/security"
)

// handleContextSearch implements the context.search tool
func (s *Server) handleContextSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req SearchRequest
	startTime := time.Now()

	if err := json.Unmarshal(args, &req); err != nil {
		errorCtx := observability.ExtractErrorContext(ctx, "context.search")
		errorCtx.ErrorType = "invalid_params"
		errorCtx.ErrorCode = protocol.InvalidParams
		errorCtx.Params = args
		errorCtx.Duration = time.Since(startTime)

		if s.errorHandler != nil {
			s.errorHandler.HandleError(ctx, err, errorCtx)
		}

		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid search request: %v", err),
		}
	}

	// Validate required fields
	if req.Query == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "query is required",
		}
	}

	// Set defaults
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

	// Check cache first (if available)
	var results []vectorstore.SearchResult
	var queryTime float64
	var cacheHit bool

	if s.searchCache != nil {
		filters := make(map[string]interface{})
		if req.Filters != nil {
			if len(req.Filters.SourceTypes) > 0 {
				filters["source_types"] = req.Filters.SourceTypes
			}
			if req.Filters.DateRange != nil {
				filters["date_range"] = map[string]string{
					"from": req.Filters.DateRange.From,
					"to":   req.Filters.DateRange.To,
				}
			}
			if req.Filters.WorkContext != nil {
				filters["work_context"] = req.Filters.WorkContext
			}
		}

		if cached, found := s.searchCache.Get(req.Query, filters); found {
			results = cached.Results
			queryTime = cached.QueryTime
			cacheHit = true

			// Record cache hit
			if s.metrics != nil {
				s.metrics.RecordSearchCacheHit()
			}
		}
	}

	// Perform search if not cached
	if !cacheHit {
		// Generate query embedding
		queryVec, err := s.embedder.Embed(ctx, req.Query)
		if err != nil {
			errorCtx := observability.ExtractErrorContext(ctx, "context.search")
			errorCtx.ErrorType = "embedding_error"
			errorCtx.ErrorCode = protocol.InternalError
			errorCtx.Params = args

			if s.errorHandler != nil {
				s.errorHandler.HandleError(ctx, err, errorCtx)
			}

			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to generate query embedding: %v", err),
			}
		}

		// Prepare search options
		opts := vectorstore.SearchOptions{
			Limit:   topK,
			Offset:  offset,
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

		// Perform hybrid search (combines vector + BM25)
		results, searchErr := s.vectorStore.SearchHybrid(ctx, req.Query, queryVec.Vector, opts)
		if searchErr != nil {
			errorCtx := observability.ExtractErrorContext(ctx, "context.search")
			errorCtx.ErrorType = "search_error"
			errorCtx.ErrorCode = protocol.InternalError
			errorCtx.Params = args

			if s.errorHandler != nil {
				s.errorHandler.HandleError(ctx, searchErr, errorCtx)
			}

			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("search failed: %v", searchErr),
			}
		}

		queryTime = float64(time.Since(startTime).Milliseconds())

		// Cache results
		if s.searchCache != nil {
			filters := make(map[string]interface{})
			if req.Filters != nil {
				if len(req.Filters.SourceTypes) > 0 {
					filters["source_types"] = req.Filters.SourceTypes
				}
				if req.Filters.DateRange != nil {
					filters["date_range"] = req.Filters.DateRange
				}
				if req.Filters.WorkContext != nil {
					filters["work_context"] = req.Filters.WorkContext
				}
			}
			s.searchCache.Set(req.Query, filters, results, queryTime)
		}

		// Record cache miss
		if s.metrics != nil && !cacheHit {
			s.metrics.RecordSearchCacheMiss()
		}
	}

	// Apply work context boosting if requested
	if req.Filters != nil && req.Filters.WorkContext != nil && req.Filters.WorkContext.BoostActive {
		results = s.applyWorkContextBoosting(results, req.Filters.WorkContext)
	}

	// Get total count for pagination
	totalCount, countErr := s.vectorStore.Count(ctx)
	if countErr != nil {
		countErrorCtx := observability.ExtractErrorContext(ctx, "context.search")
		countErrorCtx.ErrorType = "count_error"
		countErrorCtx.ErrorCode = protocol.InternalError
		countErrorCtx.Duration = time.Since(startTime)

		if s.errorHandler != nil {
			s.errorHandler.GracefulDegradation(ctx, "vector_store_count", countErr)
		}

		// Log error but don't fail the request
		totalCount = int64(len(results))
	}

	// Log successful search operation
	if s.errorHandler != nil {
		successCtx := observability.ExtractErrorContext(ctx, "context.search")
		successCtx.ErrorType = "success"
		successCtx.Params = args
		successCtx.Duration = time.Since(startTime)

		// Log success (no error to report)
		s.errorHandler.HandleError(ctx, nil, successCtx)
	}

	// Convert results to response format
	searchResults := make([]SearchResultItem, 0, len(results))
	for _, r := range results {
		// Extract source type from metadata
		sourceType := "file" // default
		if st, ok := r.Document.Metadata["source_type"].(string); ok {
			sourceType = st
		}

		searchResults = append(searchResults, SearchResultItem{
			ID:         r.Document.ID,
			Content:    r.Document.Content,
			Score:      r.Score,
			SourceType: sourceType,
			Metadata:   r.Document.Metadata,
		})
	}

	return SearchResponse{
		Results:    searchResults,
		TotalCount: len(searchResults),
		QueryTime:  queryTime,
		Offset:     offset,
		Limit:      topK,
		HasMore:    int64(offset+len(results)) < totalCount,
	}, nil
}

// handleGetRelatedInfo implements the context.get_related_info tool
func (s *Server) handleGetRelatedInfo(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GetRelatedInfoRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate that at least one identifier is provided
	if req.FilePath == "" && req.TicketID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "either file_path or ticket_id must be provided",
		}
	}

	// Validate file path for security if provided
	if req.FilePath != "" {
		safePath, err := security.ValidatePath(req.FilePath, s.rootPath)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: fmt.Sprintf("invalid file path: %v", err),
			}
		}
		req.FilePath = safePath
	}

	// Build search query based on provided identifiers
	var query string
	if req.FilePath != "" {
		query = fmt.Sprintf("file:%s", req.FilePath)
	} else {
		query = fmt.Sprintf("ticket:%s", req.TicketID)
	}

	// Create relationship detector if we have a file path
	var detector *RelationshipDetector
	if req.FilePath != "" {
		detector = NewRelationshipDetector(req.FilePath)
	}

	// Search for related documents
	queryVec, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to generate embedding: %v", err),
		}
	}

	opts := vectorstore.SearchOptions{
		Limit: 20,
	}

	results, err := s.vectorStore.SearchHybrid(ctx, query, queryVec.Vector, opts)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("search failed: %v", err),
		}
	}


	// Group results by type and build RelatedItems
	var relatedPRs, relatedIssues []string
	var discussions []DiscussionSummary
	relatedItems := make([]RelatedItem, 0, len(results))

	for _, r := range results {
		sourceType, _ := r.Document.Metadata["source_type"].(string)

		switch sourceType {
		case "github_pr":
			if prNum, ok := r.Document.Metadata["pr_number"].(string); ok {
				relatedPRs = append(relatedPRs, prNum)
			}
		case "github_issue", "jira":
			if issueID, ok := r.Document.Metadata["issue_id"].(string); ok {
				relatedIssues = append(relatedIssues, issueID)
			}
		case "slack":
			channel, _ := r.Document.Metadata["channel"].(string)
			timestamp, _ := r.Document.Metadata["timestamp"].(string)
			discussions = append(discussions, DiscussionSummary{
				Channel:   channel,
				Timestamp: timestamp,
				Summary:   r.Document.Content[:min(200, len(r.Document.Content))],
			})
		}

		// Build RelatedItem for all results
		filePath, _ := r.Document.Metadata["file_path"].(string)
		startLine, _ := r.Document.Metadata["start_line"].(int)
		endLine, _ := r.Document.Metadata["end_line"].(int)

		// Handle different numeric types for line numbers
		if startLine == 0 {
			if sl, ok := r.Document.Metadata["start_line"].(float64); ok {
				startLine = int(sl)
			}
		}
		if endLine == 0 {
			if el, ok := r.Document.Metadata["end_line"].(float64); ok {
				endLine = int(el)
			}
		}

		// Detect relationship type if we have a detector
		var relationType string
		if detector != nil {
			chunkType, _ := r.Document.Metadata["type"].(string)
			relationType = detector.DetectRelationType(filePath, chunkType, r.Document.Metadata)
		}

		relatedItems = append(relatedItems, RelatedItem{
			ID:           r.Document.ID,
			Content:      r.Document.Content,
			Score:        r.Score,
			SourceType:   sourceType,
			FilePath:     filePath,
			RelationType: relationType,
			StartLine:    startLine,
			EndLine:      endLine,
			Metadata:     r.Document.Metadata,
		})
	}

	// Generate summary
	summary := fmt.Sprintf("Found %d related items", len(results))
	if req.FilePath != "" {
		summary = fmt.Sprintf("Related information for %s: %d items (%d PRs, %d issues, %d discussions)",
			req.FilePath, len(relatedItems), len(relatedPRs), len(relatedIssues), len(discussions))
	} else {
		summary = fmt.Sprintf("Related information for ticket %s: %d items (%d PRs, %d issues, %d discussions)",
			req.TicketID, len(relatedItems), len(relatedPRs), len(relatedIssues), len(discussions))
	}

	return GetRelatedInfoResponse{
		Summary:       summary,
		RelatedPRs:    relatedPRs,
		RelatedIssues: relatedIssues,
		Discussions:   discussions,
		RelatedItems:  relatedItems,
	}, nil
}

// handleIndexControl implements the context.index_control tool
func (s *Server) handleIndexControl(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req IndexControlRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate action
	validActions := map[string]bool{
		"start":         true,
		"stop":          true,
		"status":        true,
		"force_reindex": true,
		"reindex_paths": true,
	}

	if !validActions[req.Action] {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid action: %s", req.Action),
		}
	}

	// Check if indexer is available
	if s.indexer == nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: "index controller not available",
		}
	}

	switch req.Action {
	case "status":
		// Get document count
		count, err := s.vectorStore.Count(ctx)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get document count: %v", err),
			}
		}

		// Get indexer status
		idxStatus := s.indexer.GetStatus()

		// Convert to response format
		var startTime, estimatedEnd string
		if !idxStatus.StartTime.IsZero() {
			startTime = idxStatus.StartTime.Format(time.RFC3339)
		}
		if !idxStatus.EstimatedEnd.IsZero() {
			estimatedEnd = idxStatus.EstimatedEnd.Format(time.RFC3339)
		}

		var metrics *IndexMetrics
		if idxStatus.Metrics.TotalFiles > 0 {
			metrics = &IndexMetrics{
				TotalFiles:      idxStatus.Metrics.TotalFiles,
				IndexedFiles:    idxStatus.Metrics.IndexedFiles,
				SkippedFiles:    idxStatus.Metrics.SkippedFiles,
				TotalChunks:     idxStatus.Metrics.TotalChunks,
				Duration:        idxStatus.Metrics.Duration.Seconds(),
				BytesProcessed:  idxStatus.Metrics.BytesProcessed,
				StateSize:       idxStatus.Metrics.StateSize,
				IncrementalSave: idxStatus.Metrics.IncrementalSave.Seconds(),
			}
		}

		details := map[string]interface{}{
			"documents_indexed": count,
			"indexer_available": true,
		}

		return IndexControlResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Index contains %d documents", count),
			Details: details,
			IndexStatus: &IndexStatus{
				IsIndexing:     idxStatus.IsIndexing,
				Phase:          idxStatus.Phase,
				Progress:       idxStatus.Progress,
				FilesProcessed: idxStatus.FilesProcessed,
				TotalFiles:     idxStatus.TotalFiles,
				ChunksCreated:  idxStatus.ChunksCreated,
				StartTime:      startTime,
				EstimatedEnd:   estimatedEnd,
				LastError:      idxStatus.LastError,
				Metrics:        metrics,
			},
		}, nil

	case "start":
		// Get current working directory
		rootPath, err := os.Getwd()
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get working directory: %v", err),
			}
		}

		// Load ignore patterns
		ignorePatterns := []string{".git"}
		if gitignore, err := indexer.LoadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
			ignorePatterns = append(ignorePatterns, gitignore...)
		}

		opts := indexer.IndexOptions{
			RootPath:       rootPath,
			IgnorePatterns: ignorePatterns,
			MaxFileSize:    1024 * 1024, // 1MB
			IncludeGitInfo: true,
			Embedder:       s.embedder,
			VectorStore:    s.vectorStore,
		}

		if err := s.indexer.Start(ctx, opts); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to start indexing: %v", err),
			}
		}

		return IndexControlResponse{
			Status:  "ok",
			Message: "Background indexing started",
		}, nil

	case "stop":
		if err := s.indexer.Stop(ctx); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to stop indexing: %v", err),
			}
		}

		return IndexControlResponse{
			Status:  "ok",
			Message: "Indexing stopped",
		}, nil

	case "force_reindex":
		// Get current working directory
		rootPath, err := os.Getwd()
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get working directory: %v", err),
			}
		}

		// Load ignore patterns
		ignorePatterns := []string{".git"}
		if gitignore, err := indexer.LoadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
			ignorePatterns = append(ignorePatterns, gitignore...)
		}

		opts := indexer.IndexOptions{
			RootPath:       rootPath,
			IgnorePatterns: ignorePatterns,
			MaxFileSize:    1024 * 1024, // 1MB
			IncludeGitInfo: true,
			Embedder:       s.embedder,
			VectorStore:    s.vectorStore,
		}

		if err := s.indexer.ForceReindex(ctx, opts); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to start force reindex: %v", err),
			}
		}

		return IndexControlResponse{
			Status:  "ok",
			Message: "Force reindex started",
		}, nil

	case "reindex_paths":
		if len(req.Paths) == 0 {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "paths are required for reindex_paths action",
			}
		}

		// Get current working directory
		rootPath, err := os.Getwd()
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get working directory: %v", err),
			}
		}

		// Load ignore patterns
		ignorePatterns := []string{".git"}
		if gitignore, err := indexer.LoadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
			ignorePatterns = append(ignorePatterns, gitignore...)
		}

		opts := indexer.IndexOptions{
			RootPath:       rootPath,
			IgnorePatterns: ignorePatterns,
			MaxFileSize:    1024 * 1024, // 1MB
			IncludeGitInfo: true,
			Embedder:       s.embedder,
			VectorStore:    s.vectorStore,
		}

		if err := s.indexer.ReindexPaths(ctx, opts, req.Paths); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to start selective reindex: %v", err),
			}
		}

		return IndexControlResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Reindexing %d paths", len(req.Paths)),
		}, nil

	default:
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("unimplemented action: %s", req.Action),
		}
	}
}

// handleConnectorManagement implements the context.connector_management tool
func (s *Server) handleConnectorManagement(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req ConnectorManagementRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate action
	validActions := map[string]bool{
		"list":   true,
		"add":    true,
		"update": true,
		"remove": true,
	}

	if !validActions[req.Action] {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid action: %s", req.Action),
		}
	}

	switch req.Action {
	case "list":
		connectors, err := s.connectorStore.List(ctx)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to list connectors: %v", err),
			}
		}

		connectorInfos := make([]ConnectorInfo, len(connectors))
		for i, conn := range connectors {
			connectorInfos[i] = ConnectorInfo{
				ID:     conn.ID,
				Type:   conn.Type,
				Name:   conn.Name,
				Status: conn.Status,
				Config: conn.Config,
			}
		}

		return ConnectorManagementResponse{
			Connectors: connectorInfos,
			Status:     "ok",
			Message:    "Retrieved connector list",
		}, nil

	case "add":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		connector := &connectors.Connector{
			ID:     req.ConnectorID,
			Name:   req.ConnectorID, // Default name to ID, can be updated later
			Type:   "filesystem",    // Default type, should be specified in config
			Config: req.ConnectorConfig,
			Status: "active",
		}

		// Extract type and name from config if provided
		if configType, ok := req.ConnectorConfig["type"].(string); ok {
			connector.Type = configType
		}
		if configName, ok := req.ConnectorConfig["name"].(string); ok {
			connector.Name = configName
		}

		if err := s.connectorStore.Add(ctx, connector); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to add connector: %v", err),
			}
		}

		return ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s added successfully", req.ConnectorID),
		}, nil

	case "update":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		connector := &connectors.Connector{
			Type:   "filesystem", // Default type
			Config: req.ConnectorConfig,
			Status: "active",
		}

		// Extract type and name from config if provided
		if configType, ok := req.ConnectorConfig["type"].(string); ok {
			connector.Type = configType
		}
		if configName, ok := req.ConnectorConfig["name"].(string); ok {
			connector.Name = configName
		}

		if err := s.connectorStore.Update(ctx, req.ConnectorID, connector); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to update connector: %v", err),
			}
		}

		return ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s updated successfully", req.ConnectorID),
		}, nil

	case "remove":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		if err := s.connectorStore.Remove(ctx, req.ConnectorID); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to remove connector: %v", err),
			}
		}

		return ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s removed successfully", req.ConnectorID),
		}, nil

	default:
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: "unexpected error",
		}
	}
}

// applyWorkContextBoosting boosts results related to active work context
func (s *Server) applyWorkContextBoosting(results []vectorstore.SearchResult, workContext *WorkContextFilters) []vectorstore.SearchResult {
	if workContext == nil {
		return results
	}

	// Create boosted results
	boosted := make([]vectorstore.SearchResult, len(results))
	copy(boosted, results)

	boostFactor := float32(1.2) // 20% boost for relevant results

	for i := range boosted {
		score := boosted[i].Score

		// Boost if result is related to active file
		if workContext.ActiveFile != "" {
			if filePath, ok := boosted[i].Document.Metadata["file_path"].(string); ok {
				if filePath == workContext.ActiveFile {
					score *= boostFactor
				}
			}
		}

		// Boost if result is related to open tickets
		if len(workContext.OpenTicketIDs) > 0 {
			if ticketID, ok := boosted[i].Document.Metadata["ticket_id"].(string); ok {
				for _, openTicket := range workContext.OpenTicketIDs {
					if ticketID == openTicket {
						score *= boostFactor
						break
					}
				}
			}
		}

		// Boost if result matches git branch
		if workContext.GitBranch != "" {
			if branch, ok := boosted[i].Document.Metadata["git_branch"].(string); ok {
				if branch == workContext.GitBranch {
					score *= boostFactor
				}
			}
		}

		boosted[i].Score = score
	}

	// Re-sort by boosted scores
	sort.Slice(boosted, func(i, j int) bool {
		return boosted[i].Score > boosted[j].Score
	})

	return boosted
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
