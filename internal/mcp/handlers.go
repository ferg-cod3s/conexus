package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	_ "github.com/ferg-cod3s/conexus/internal/connectors/discord" // Imported for types used by connectorManager
	"github.com/ferg-cod3s/conexus/internal/connectors/github"
	_ "github.com/ferg-cod3s/conexus/internal/connectors/jira" // Imported for types used by connectorManager
	"github.com/ferg-cod3s/conexus/internal/connectors/slack"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/indexer"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
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
				// Add story context filtering
				if req.Filters.WorkContext.CurrentStoryID != "" {
					opts.Filters["story_ids"] = []string{req.Filters.WorkContext.CurrentStoryID}
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
		var searchErr error
		results, searchErr = s.vectorStore.SearchHybrid(ctx, req.Query, queryVec.Vector, opts)
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

	// Apply semantic reranking for better relevance
	if len(results) > 1 {
		results = s.applySemanticReranking(ctx, results, req.Query)
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

	// Build search query and filters based on provided identifiers
	var query string
	opts := vectorstore.SearchOptions{
		Limit:   20,
		Filters: make(map[string]interface{}),
	}

	if req.FilePath != "" {
		query = req.FilePath
		opts.Filters["file_path"] = req.FilePath
	} else {
		query = req.TicketID
		opts.Filters["ticket_id"] = req.TicketID
	}

	// Search for related documents
	queryVec, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to generate embedding: %v", err),
		}
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

		relatedItems = append(relatedItems, RelatedItem{
			ID:         r.Document.ID,
			Content:    r.Document.Content,
			Score:      r.Score,
			SourceType: sourceType,
			FilePath:   filePath,
			StartLine:  startLine,
			EndLine:    endLine,
			Metadata:   r.Document.Metadata,
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
		"index":         true,
		"sync_github":   true,
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
		if gitignore, err := loadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
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
		if gitignore, err := loadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
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
		if gitignore, err := loadGitignore(filepath.Join(rootPath, ".gitignore"), rootPath); err == nil {
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

	case "index":
		// Handle single document indexing
		if req.Content == nil {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "content is required for index action",
			}
		}

		// Validate content fields
		if req.Content.Path == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "content.path is required",
			}
		}
		if req.Content.Content == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "content.content is required",
			}
		}
		if req.Content.SourceType == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "content.source_type is required",
			}
		}

		// Create metadata for the document
		metadata := map[string]interface{}{
			"file_path":   req.Content.Path,
			"source_type": req.Content.SourceType,
			"indexed_at":  time.Now().Format(time.RFC3339),
		}

		// Add line range information if provided
		if req.Content.StartLine != nil {
			metadata["start_line"] = *req.Content.StartLine
		}
		if req.Content.EndLine != nil {
			metadata["end_line"] = *req.Content.EndLine
		}

		// Generate embedding for the content
		embedding, err := s.embedder.Embed(ctx, req.Content.Content)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to generate embedding: %v", err),
			}
		}

		// Create document record
		doc := vectorstore.Document{
			ID:       req.Content.Path,
			Content:  req.Content.Content,
			Vector:   embedding.Vector,
			Metadata: metadata,
		}

		// Store in vector store
		if err := s.vectorStore.Upsert(ctx, doc); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to store document: %v", err),
			}
		}

		return IndexControlResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Successfully indexed document: %s", req.Content.Path),
			Details: map[string]interface{}{
				"document_id":    req.Content.Path,
				"content_length": len(req.Content.Content),
			},
		}, nil

	case "sync_github":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required for sync_github action",
			}
		}

		// Sync GitHub issues and PRs using the connector manager
		issues, err := s.connectorManager.SyncGitHubIssues(ctx, req.ConnectorID)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to sync GitHub issues: %v", err),
			}
		}

		prs, err := s.connectorManager.SyncGitHubPullRequests(ctx, req.ConnectorID)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to sync GitHub PRs: %v", err),
			}
		}

		// Convert issues to documents and store them
		for _, issue := range issues {
			content := fmt.Sprintf("%s\n\n%s", issue.Title, issue.Description)
			if content == "" {
				content = issue.Title
			}

			doc := vectorstore.Document{
				ID:      fmt.Sprintf("github-issue-%d", issue.Number),
				Content: content,
				Metadata: map[string]interface{}{
					"source_type":  "github_issue",
					"issue_number": issue.Number,
					"title":        issue.Title,
					"state":        issue.State,
					"labels":       issue.Labels,
					"assignee":     issue.Assignee,
					"created_at":   issue.CreatedAt.Format(time.RFC3339),
					"updated_at":   issue.UpdatedAt.Format(time.RFC3339),
					"connector_id": req.ConnectorID,
				},
				CreatedAt: issue.CreatedAt,
				UpdatedAt: issue.UpdatedAt,
			}

			// Generate embedding
			embedding, err := s.embedder.Embed(ctx, doc.Content)
			if err != nil {
				continue // Skip if embedding fails
			}
			doc.Vector = embedding.Vector

			if err := s.vectorStore.Upsert(ctx, doc); err != nil {
				return nil, &protocol.Error{
					Code:    protocol.InternalError,
					Message: fmt.Sprintf("failed to store issue %d: %v", issue.Number, err),
				}
			}
		}

		// Convert PRs to documents and store them
		for _, pr := range prs {
			content := fmt.Sprintf("%s\n\n%s", pr.Title, pr.Description)
			if content == "" {
				content = pr.Title
			}

			doc := vectorstore.Document{
				ID:      fmt.Sprintf("github-pr-%d", pr.Number),
				Content: content,
				Metadata: map[string]interface{}{
					"source_type":   "github_pr",
					"pr_number":     pr.Number,
					"title":         pr.Title,
					"state":         pr.State,
					"labels":        pr.Labels,
					"assignee":      pr.Assignee,
					"created_at":    pr.CreatedAt.Format(time.RFC3339),
					"updated_at":    pr.UpdatedAt.Format(time.RFC3339),
					"linked_issues": pr.LinkedIssues,
					"connector_id":  req.ConnectorID,
					"merged":        pr.Merged,
					"head_branch":   pr.HeadBranch,
					"base_branch":   pr.BaseBranch,
				},
				CreatedAt: pr.CreatedAt,
				UpdatedAt: pr.UpdatedAt,
				PRNumbers: []string{fmt.Sprintf("%d", pr.Number)},
			}

			// Generate embedding
			embedding, err := s.embedder.Embed(ctx, doc.Content)
			if err != nil {
				continue // Skip if embedding fails
			}
			doc.Vector = embedding.Vector

			if err := s.vectorStore.Upsert(ctx, doc); err != nil {
				return nil, &protocol.Error{
					Code:    protocol.InternalError,
					Message: fmt.Sprintf("failed to store PR %d: %v", pr.Number, err),
				}
			}
		}

		return IndexControlResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Successfully synced %d issues and %d pull requests", len(issues), len(prs)),
			Details: map[string]interface{}{
				"issues_synced": len(issues),
				"prs_synced":    len(prs),
			},
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
			Message:    fmt.Sprintf("Retrieved %d connectors", len(connectors)),
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
			Connectors: []ConnectorInfo{{
				ID:     connector.ID,
				Type:   connector.Type,
				Name:   connector.Name,
				Status: connector.Status,
				Config: connector.Config,
			}},
		}, nil

	case "update":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		// Get existing connector first
		existing, err := s.connectorStore.Get(ctx, req.ConnectorID)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get connector: %v", err),
			}
		}

		// Update fields from request
		if configType, ok := req.ConnectorConfig["type"].(string); ok {
			existing.Type = configType
		}
		if configName, ok := req.ConnectorConfig["name"].(string); ok {
			existing.Name = configName
		}
		if configStatus, ok := req.ConnectorConfig["status"].(string); ok {
			existing.Status = configStatus
		}
		// Merge configs (request config overrides existing)
		for k, v := range req.ConnectorConfig {
			existing.Config[k] = v
		}

		if err := s.connectorStore.Update(ctx, req.ConnectorID, existing); err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to update connector: %v", err),
			}
		}

		return ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s updated successfully", req.ConnectorID),
			Connectors: []ConnectorInfo{{
				ID:     existing.ID,
				Type:   existing.Type,
				Name:   existing.Name,
				Status: existing.Status,
				Config: existing.Config,
			}},
		}, nil

	case "remove":
		if req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}

		// Get connector info before removal for response
		existing, err := s.connectorStore.Get(ctx, req.ConnectorID)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get connector: %v", err),
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
			Connectors: []ConnectorInfo{{
				ID:     existing.ID,
				Type:   existing.Type,
				Name:   existing.Name,
				Status: existing.Status,
				Config: existing.Config,
			}},
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

// applySemanticReranking improves result relevance using cross-attention
func (s *Server) applySemanticReranking(ctx context.Context, results []vectorstore.SearchResult, query string) []vectorstore.SearchResult {
	if len(results) <= 1 {
		return results
	}

	// Generate query embedding for reranking
	queryVec, err := s.embedder.Embed(ctx, query)
	if err != nil {
		// If embedding fails, return original results
		return results
	}

	// Calculate semantic similarity scores
	for i := range results {
		docVec := results[i].Document.Vector
		if len(docVec) > 0 && len(queryVec.Vector) > 0 {
			// Simple cosine similarity for reranking
			similarity := s.calculateCosineSimilarity(queryVec.Vector, docVec)
			// Combine original score with semantic similarity (weighted average)
			results[i].Score = results[i].Score*0.7 + float32(similarity)*0.3
		}
	}

	// Re-sort by reranked scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// calculateCosineSimilarity computes cosine similarity between two vectors
func (s *Server) calculateCosineSimilarity(a, b embedding.Vector) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// loadGitignore loads .gitignore patterns if available.
func loadGitignore(gitignorePath, rootPath string) ([]string, error) {
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil, nil
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return nil, fmt.Errorf("read .gitignore: %w", err)
	}

	var patterns []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			// Convert to absolute path for proper matching
			if !filepath.IsAbs(line) {
				patterns = append(patterns, line)
			}
		}
	}

	return patterns, nil
}

// extractStoryIDsFromIssue extracts story IDs from GitHub issue content
func extractStoryIDsFromIssue(issue github.Issue) []string {
	var storyIDs []string

	// Extract from issue title and description
	content := fmt.Sprintf("%s\n%s", issue.Title, issue.Description)

	// Pattern for issue references: #123, PROJ-456, JIRA-999
	issuePattern := regexp.MustCompile(`(?:#|PROJ-|JIRA-)(\d+)`)
	if matches := issuePattern.FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				storyIDs = append(storyIDs, match[1])
			}
		}
	}

	// Extract from labels (e.g., "story: PROJ-123")
	for _, label := range issue.Labels {
		labelPattern := regexp.MustCompile(`(?:story|feature|bug):?\s*([A-Z]+-\d+)`)
		if matches := labelPattern.FindAllStringSubmatch(label, -1); matches != nil {
			for _, match := range matches {
				if len(match) > 1 {
					storyIDs = append(storyIDs, match[1])
				}
			}
		}
	}

	return storyIDs
}

// syncGitHubData syncs GitHub issues and PRs using the connector configuration
func (s *Server) syncGitHubData(ctx context.Context, connector *connectors.Connector) ([]github.Issue, []github.PullRequest, error) {
	// Extract GitHub configuration
	config := connector.Config
	repoURL, _ := config["repo_url"].(string)

	if repoURL == "" {
		return nil, nil, fmt.Errorf("repo_url not configured for connector %s", connector.ID)
	}

	// Parse repository information from URL
	// Expected format: https://github.com/owner/repo
	parts := strings.TrimPrefix(repoURL, "https://github.com/")
	parts = strings.TrimSuffix(parts, "/")
	ownerRepo := strings.Split(parts, "/")
	if len(ownerRepo) != 2 {
		return nil, nil, fmt.Errorf("invalid GitHub repository URL: %s", repoURL)
	}

	// TODO: Implement actual GitHub API integration
	// This would require:
	// 1. GitHub API client setup
	// 2. Authentication using token from config
	// 3. Fetching issues and PRs from the repository
	// 4. Converting API responses to internal types

	// For now, return empty results as this would require external API integration
	return []github.Issue{}, []github.PullRequest{}, nil
}

// handleContextExplain implements the context.explain tool
func (s *Server) handleContextExplain(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req ExplainRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.Target == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "target is required",
		}
	}

	// Set defaults
	if req.Depth == "" {
		req.Depth = "detailed"
	}

	// Search for relevant code and documentation
	queryVec, err := s.embedder.Embed(ctx, req.Target)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to generate embedding: %v", err),
		}
	}

	// Search with broader context for explanations
	opts := vectorstore.SearchOptions{
		Limit:   15, // Get more results for comprehensive explanations
		Filters: make(map[string]interface{}),
	}

	results, err := s.vectorStore.SearchHybrid(ctx, req.Target, queryVec.Vector, opts)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("search failed: %v", err),
		}
	}

	// Generate explanation based on results
	explanation := s.generateExplanation(req.Target, req.Context, req.Depth, results)

	// Find related examples
	var examples []CodeExample
	var related []RelatedItem

	for _, result := range results[:min(5, len(results))] {
		// Add as related item
		filePath, _ := result.Document.Metadata["file_path"].(string)
		startLine, _ := result.Document.Metadata["start_line"].(float64)
		endLine, _ := result.Document.Metadata["end_line"].(float64)

		related = append(related, RelatedItem{
			ID:         result.Document.ID,
			Content:    result.Document.Content,
			Score:      result.Score,
			SourceType: getStringFromMetadata(result.Document.Metadata, "source_type"),
			FilePath:   filePath,
			StartLine:  int(startLine),
			EndLine:    int(endLine),
			Metadata:   result.Document.Metadata,
		})

		// Extract code examples from function/struct definitions
		if chunkType, ok := result.Document.Metadata["chunk_type"].(string); ok {
			if chunkType == "function" || chunkType == "struct" {
				examples = append(examples, CodeExample{
					Code:        result.Document.Content,
					Description: fmt.Sprintf("Example %s from %s", chunkType, filePath),
					Language:    getStringFromMetadata(result.Document.Metadata, "language"),
				})
			}
		}
	}

	// Determine complexity
	complexity := s.assessComplexity(results)

	return ExplainResponse{
		Explanation: explanation,
		Examples:    examples,
		Related:     related,
		Complexity:  complexity,
		Metadata: map[string]interface{}{
			"search_results":     len(results),
			"explanation_length": len(explanation),
		},
	}, nil
}

// handleContextGrep implements the context.grep tool
func (s *Server) handleContextGrep(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GrepRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.Pattern == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "pattern is required",
		}
	}

	// Set defaults
	if req.Context == 0 {
		req.Context = 3
	}
	if req.Path == "" {
		req.Path = "."
	}

	startTime := time.Now()

	// Use ripgrep for fast pattern matching
	var results []GrepResult
	var err error

	if req.Include != "" {
		// Search in specific file types
		results, err = s.grepInFiles(ctx, req.Pattern, req.Path, req.Include, req.CaseInsensitive, req.Context)
	} else {
		// Search in all files
		results, err = s.grepInFiles(ctx, req.Pattern, req.Path, "*", req.CaseInsensitive, req.Context)
	}

	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("grep failed: %v", err),
		}
	}

	searchTime := float64(time.Since(startTime).Nanoseconds()) / 1e6 // Convert to milliseconds

	return GrepResponse{
		Results:    results,
		TotalCount: len(results),
		SearchTime: searchTime,
	}, nil
}

// generateExplanation creates a detailed explanation from search results
func (s *Server) generateExplanation(target, context, depth string, results []vectorstore.SearchResult) string {
	if len(results) == 0 {
		return fmt.Sprintf("No information found for '%s'. Try broadening your search or check if the code has been indexed.", target)
	}

	var explanation strings.Builder
	explanation.WriteString(fmt.Sprintf("## Explanation of: %s\n\n", target))

	// Add context if provided
	if context != "" {
		explanation.WriteString(fmt.Sprintf("**Context:** %s\n\n", context))
	}

	// Group results by type for better organization
	var functions, structs, files []vectorstore.SearchResult
	for _, result := range results {
		chunkType := getStringFromMetadata(result.Document.Metadata, "chunk_type")
		switch chunkType {
		case "function":
			functions = append(functions, result)
		case "struct":
			structs = append(structs, result)
		default:
			files = append(files, result)
		}
	}

	// Explain functions
	if len(functions) > 0 {
		explanation.WriteString("### Functions:\n")
		for _, fn := range functions[:min(3, len(functions))] {
			funcName := getStringFromMetadata(fn.Document.Metadata, "function_name")
			if funcName == "" {
				funcName = "unnamed function"
			}
			explanation.WriteString(fmt.Sprintf("- **%s**: %s\n",
				funcName, s.summarizeContent(fn.Document.Content, 100)))
		}
		explanation.WriteString("\n")
	}

	// Explain structs/types
	if len(structs) > 0 {
		explanation.WriteString("### Data Structures:\n")
		for _, st := range structs[:min(3, len(structs))] {
			structName := getStringFromMetadata(st.Document.Metadata, "type_name")
			if structName == "" {
				structName = "unnamed struct"
			}
			explanation.WriteString(fmt.Sprintf("- **%s**: %s\n",
				structName, s.summarizeContent(st.Document.Content, 100)))
		}
		explanation.WriteString("\n")
	}

	// Add implementation details based on depth
	if depth == "comprehensive" || depth == "detailed" {
		explanation.WriteString("### Implementation Details:\n")
		for _, result := range results[:min(5, len(results))] {
			filePath := getStringFromMetadata(result.Document.Metadata, "file_path")
			if filePath != "" {
				explanation.WriteString(fmt.Sprintf("- **%s**: Located in %s\n",
					getStringFromMetadata(result.Document.Metadata, "chunk_type"),
					filePath))
			}
		}
		explanation.WriteString("\n")
	}

	// Add usage guidance
	if depth == "comprehensive" {
		explanation.WriteString("### Usage Guidance:\n")
		explanation.WriteString("Consider the context and related functions when using this code. ")
		explanation.WriteString("Check for error handling patterns and ensure proper resource cleanup.\n\n")
	}

	return explanation.String()
}

// grepInFiles performs pattern matching across files
func (s *Server) grepInFiles(ctx context.Context, pattern, basePath, filePattern string, caseInsensitive bool, contextLines int) ([]GrepResult, error) {
	// Get list of files to search
	files, err := s.getFilesToSearch(basePath, filePattern)
	if err != nil {
		return nil, err
	}

	var results []GrepResult

	// For each file, search for the pattern
	for _, file := range files {
		matches, err := s.grepInFile(file, pattern, caseInsensitive, contextLines)
		if err != nil {
			continue // Skip files with errors
		}
		results = append(results, matches...)
	}

	return results, nil
}

// getFilesToSearch returns list of files matching the pattern
func (s *Server) getFilesToSearch(basePath, filePattern string) ([]string, error) {
	var files []string

	// If filePattern is just "*" or empty, search all files
	if filePattern == "*" || filePattern == "" {
		err := filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !d.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return files, nil
	}

	// Use glob for specific patterns
	pattern := filepath.Join(basePath, "**", filePattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Filter out directories and validate paths
	for _, match := range matches {
		if info, err := os.Stat(match); err == nil && !info.IsDir() {
			files = append(files, match)
		}
	}

	return files, nil
}

// grepInFile searches for pattern in a single file
func (s *Server) grepInFile(filePath, pattern string, caseInsensitive bool, contextLines int) ([]GrepResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var results []GrepResult

	// Compile regex pattern - escape special regex characters for literal matches
	regexPattern := regexp.QuoteMeta(pattern)
	if caseInsensitive {
		regexPattern = "(?i)" + regexPattern
	}

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	// Search each line
	for i, line := range lines {
		if re.MatchString(line) {
			// Extract the matching part (first match)
			match := re.FindString(line)
			if match == "" {
				continue
			}

			// Get context lines
			start := i - contextLines
			if start < 0 {
				start = 0
			}
			end := i + contextLines + 1
			if end > len(lines) {
				end = len(lines)
			}

			contextContent := strings.Join(lines[start:end], "\n")

			results = append(results, GrepResult{
				File:    filePath,
				Line:    i + 1,
				Content: contextContent,
				Match:   match,
			})
		}
	}

	return results, nil
}

// assessComplexity determines the complexity level of the code
func (s *Server) assessComplexity(results []vectorstore.SearchResult) string {
	if len(results) == 0 {
		return "unknown"
	}

	// Simple heuristic based on number of results and content length
	totalContentLength := 0
	maxScore := float32(0)

	for _, result := range results {
		totalContentLength += len(result.Document.Content)
		if result.Score > maxScore {
			maxScore = result.Score
		}
	}

	avgContentLength := totalContentLength / len(results)

	if len(results) <= 2 && avgContentLength < 200 && maxScore > 0.8 {
		return "simple"
	} else if len(results) <= 5 && avgContentLength < 500 {
		return "moderate"
	} else {
		return "complex"
	}
}

// summarizeContent creates a brief summary of content
func (s *Server) summarizeContent(content string, maxLength int) string {
	lines := strings.Split(content, "\n")
	var summaryLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "/*") && !strings.HasPrefix(line, "*") {
			summaryLines = append(summaryLines, line)
			if len(summaryLines) >= 2 {
				break
			}
		}
	}

	summary := strings.Join(summaryLines, " ")
	if len(summary) > maxLength {
		summary = summary[:maxLength-3] + "..."
	}

	return summary
}

// handleGitHubSyncStatus implements the github.sync_status tool
func (s *Server) handleGitHubSyncStatus(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GitHubSyncStatusRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Get all connectors or specific connector
	var connectorList []*connectors.Connector
	var err error

	if req.ConnectorID != "" {
		connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get connector: %v", err),
			}
		}
		connectorList = []*connectors.Connector{connector}
	} else {
		connectorList, err = s.connectorStore.List(ctx)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to list connectors: %v", err),
			}
		}
	}

	// Collect sync status for all GitHub connectors
	var githubConnectors []map[string]interface{}
	for _, connector := range connectorList {
		if connector.Type == "github" {
			// Get sync status from connector
			if conn, err := s.connectorManager.GetConnector(ctx, connector.ID); err == nil {
				if githubConn, ok := conn.(*github.Connector); ok {
					syncStatus := githubConn.GetSyncStatus()
					rateLimit := githubConn.GetRateLimit()

					githubConnectors = append(githubConnectors, map[string]interface{}{
						"connector_id":   connector.ID,
						"connector_name": connector.Name,
						"sync_status":    syncStatus,
						"rate_limit":     rateLimit,
					})
				}
			}
		}
	}

	if len(githubConnectors) == 0 {
		return GitHubSyncStatusResponse{
			Status:  "ok",
			Message: "No GitHub connectors found",
			Details: map[string]interface{}{
				"github_connectors": githubConnectors,
			},
		}, nil
	}

	return GitHubSyncStatusResponse{
		Status:  "ok",
		Message: fmt.Sprintf("Retrieved status for %d GitHub connector(s)", len(githubConnectors)),
		Details: map[string]interface{}{
			"github_connectors": githubConnectors,
		},
	}, nil
}

// handleGitHubSyncTrigger implements the github.sync_trigger tool
func (s *Server) handleGitHubSyncTrigger(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GitHubSyncTriggerRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}

	// Check if connector exists and is GitHub type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "github" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a GitHub connector", req.ConnectorID),
		}
	}

	// Generate job ID for tracking
	jobID := fmt.Sprintf("github-sync-%s-%d", req.ConnectorID, time.Now().Unix())

	// Start sync in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		// Sync issues
		issues, err := s.connectorManager.SyncGitHubIssues(ctx, req.ConnectorID)
		if err != nil {
			s.errorHandler.HandleError(ctx, err, observability.ExtractErrorContext(ctx, "github_sync_issues"))
			return
		}

		// Sync pull requests
		prs, err := s.connectorManager.SyncGitHubPullRequests(ctx, req.ConnectorID)
		if err != nil {
			s.errorHandler.HandleError(ctx, err, observability.ExtractErrorContext(ctx, "github_sync_prs"))
			return
		}

		// Store issues in vector store
		for _, issue := range issues {
			content := fmt.Sprintf("%s\n\n%s", issue.Title, issue.Description)
			if content == "" {
				content = issue.Title
			}

			doc := vectorstore.Document{
				ID:      fmt.Sprintf("github-issue-%d", issue.Number),
				Content: content,
				Metadata: map[string]interface{}{
					"source_type":  "github_issue",
					"issue_number": issue.Number,
					"title":        issue.Title,
					"state":        issue.State,
					"labels":       issue.Labels,
					"assignee":     issue.Assignee,
					"created_at":   issue.CreatedAt.Format(time.RFC3339),
					"updated_at":   issue.UpdatedAt.Format(time.RFC3339),
					"connector_id": req.ConnectorID,
				},
				CreatedAt: issue.CreatedAt,
				UpdatedAt: issue.UpdatedAt,
			}

			// Generate embedding
			embedding, err := s.embedder.Embed(ctx, doc.Content)
			if err != nil {
				continue // Skip if embedding fails
			}
			doc.Vector = embedding.Vector

			if err := s.vectorStore.Upsert(ctx, doc); err != nil {
				s.errorHandler.HandleError(ctx, err, observability.ExtractErrorContext(ctx, "github_store_issue"))
			}
		}

		// Store PRs in vector store
		for _, pr := range prs {
			content := fmt.Sprintf("%s\n\n%s", pr.Title, pr.Description)
			if content == "" {
				content = pr.Title
			}

			doc := vectorstore.Document{
				ID:      fmt.Sprintf("github-pr-%d", pr.Number),
				Content: content,
				Metadata: map[string]interface{}{
					"source_type":   "github_pr",
					"pr_number":     pr.Number,
					"title":         pr.Title,
					"state":         pr.State,
					"labels":        pr.Labels,
					"assignee":      pr.Assignee,
					"created_at":    pr.CreatedAt.Format(time.RFC3339),
					"updated_at":    pr.UpdatedAt.Format(time.RFC3339),
					"linked_issues": pr.LinkedIssues,
					"connector_id":  req.ConnectorID,
					"merged":        pr.Merged,
					"head_branch":   pr.HeadBranch,
					"base_branch":   pr.BaseBranch,
				},
				CreatedAt: pr.CreatedAt,
				UpdatedAt: pr.UpdatedAt,
				PRNumbers: []string{fmt.Sprintf("%d", pr.Number)},
			}

			// Generate embedding
			embedding, err := s.embedder.Embed(ctx, doc.Content)
			if err != nil {
				continue // Skip if embedding fails
			}
			doc.Vector = embedding.Vector

			if err := s.vectorStore.Upsert(ctx, doc); err != nil {
				s.errorHandler.HandleError(ctx, err, observability.ExtractErrorContext(ctx, "github_store_pr"))
			}
		}
	}()

	return GitHubSyncTriggerResponse{
		Status:      "ok",
		Message:     fmt.Sprintf("GitHub sync started for connector %s", req.ConnectorID),
		JobID:       jobID,
		ConnectorID: req.ConnectorID,
		Details: map[string]interface{}{
			"force": req.Force,
		},
	}, nil
}

// handleSlackSearch implements the slack.search tool
func (s *Server) handleSlackSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req SlackSearchRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.Query == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "query is required",
		}
	}

	// Check if connector exists and is Slack type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "slack" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Slack connector", req.ConnectorID),
		}
	}

	// Search messages
	messages, err := s.connectorManager.SearchSlackMessages(ctx, req.ConnectorID, req.Query)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to search Slack messages: %v", err),
		}
	}

	// Convert to MCP response format
	mcpMessages := make([]SlackMessage, 0, len(messages))
	for _, msg := range messages {
		// Parse timestamp to time.Time
		timestamp := msg.CreatedAt

		mcpMessages = append(mcpMessages, SlackMessage{
			ID:        msg.ID,
			ChannelID: msg.Channel,
			UserID:    msg.User,
			Author:    msg.User, // Use User as Author
			Content:   msg.Text,
			Timestamp: timestamp,
			ThreadTS:  msg.ThreadTS,
			IsBot:     false, // slack.Message doesn't track this in our struct
		})
	}

	return SlackSearchResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Found %d messages", len(mcpMessages)),
		Messages: mcpMessages,
	}, nil
}

// handleSlackListChannels implements the slack.list_channels tool
func (s *Server) handleSlackListChannels(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req SlackListChannelsRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}

	// Check if connector exists and is Slack type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "slack" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Slack connector", req.ConnectorID),
		}
	}

	// List channels
	channels, err := s.connectorManager.ListSlackChannels(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to list Slack channels: %v", err),
		}
	}

	// Convert to MCP response format
	mcpChannels := make([]SlackChannel, 0, len(channels))
	for _, ch := range channels {
		mcpChannels = append(mcpChannels, SlackChannel{
			ID:          ch.ID,
			Name:        ch.Name,
			IsPrivate:   ch.IsPrivate,
			MemberCount: ch.Members,
			Topic:       "", // Not available in slack.Channel
			Purpose:     "", // Not available in slack.Channel
		})
	}

	return SlackListChannelsResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Found %d channels", len(mcpChannels)),
		Channels: mcpChannels,
	}, nil
}

// handleSlackGetThread implements the slack.get_thread tool
func (s *Server) handleSlackGetThread(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req SlackGetThreadRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.ChannelID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "channel_id is required",
		}
	}
	if req.ThreadTS == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "thread_ts is required",
		}
	}

	// Check if connector exists and is Slack type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "slack" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Slack connector", req.ConnectorID),
		}
	}

	// Get thread
	thread, err := s.connectorManager.GetSlackThread(ctx, req.ConnectorID, req.ChannelID, req.ThreadTS)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get Slack thread: %v", err),
		}
	}

	// Convert to MCP response format
	// Thread has ParentMessage and Replies, we need to combine them
	allMessages := []slack.Message{thread.ParentMessage}
	allMessages = append(allMessages, thread.Replies...)

	mcpMessages := make([]SlackMessage, 0, len(allMessages))
	for _, msg := range allMessages {
		mcpMessages = append(mcpMessages, SlackMessage{
			ID:        msg.ID,
			ChannelID: msg.Channel,
			UserID:    msg.User,
			Author:    msg.User,
			Content:   msg.Text,
			Timestamp: msg.CreatedAt,
			ThreadTS:  msg.ThreadTS,
			IsBot:     false,
		})
	}

	mcpThread := &SlackThread{
		ChannelID:    req.ChannelID,
		ThreadTS:     req.ThreadTS,
		MessageCount: len(mcpMessages),
		Messages:     mcpMessages,
	}

	return SlackGetThreadResponse{
		Status:  "ok",
		Message: fmt.Sprintf("Found thread with %d messages", len(mcpMessages)),
		Thread:  mcpThread,
	}, nil
}

// handleJiraSearch implements the jira.search tool
func (s *Server) handleJiraSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req JiraSearchRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.JQL == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "jql is required",
		}
	}

	// Check if connector exists and is Jira type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "jira" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Jira connector", req.ConnectorID),
		}
	}

	// Search issues
	issues, err := s.connectorManager.SearchJiraIssues(ctx, req.ConnectorID, req.JQL)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to search Jira issues: %v", err),
		}
	}

	// Convert to MCP response format
	mcpIssues := make([]JiraIssue, 0, len(issues))
	for _, issue := range issues {
		mcpIssue := JiraIssue{
			ID:          issue.ID,
			Key:         issue.Key,
			Summary:     issue.Summary,
			Description: issue.Description,
			Status:      issue.Status,
			Priority:    issue.Priority,
			IssueType:   issue.IssueType,
			Assignee:    issue.Assignee,
			Reporter:    issue.Reporter,
			Labels:      issue.Labels,
			Components:  issue.Components,
			FixVersions: issue.FixVersions,
			CreatedAt:   issue.CreatedAt,
			UpdatedAt:   issue.UpdatedAt,
			Project:     issue.Project,
		}
		if issue.ResolvedAt != nil {
			mcpIssue.ResolvedAt = *issue.ResolvedAt
		}
		mcpIssues = append(mcpIssues, mcpIssue)
	}

	return JiraSearchResponse{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d issues", len(mcpIssues)),
		Issues:  mcpIssues,
	}, nil
}

// handleJiraGetIssue implements the jira.get_issue tool
func (s *Server) handleJiraGetIssue(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req JiraGetIssueRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.IssueKey == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "issue_key is required",
		}
	}

	// Check if connector exists and is Jira type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "jira" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Jira connector", req.ConnectorID),
		}
	}

	// Get issue
	issue, err := s.connectorManager.GetJiraIssue(ctx, req.ConnectorID, req.IssueKey)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get Jira issue: %v", err),
		}
	}

	// Get comments
	comments, err := s.connectorManager.GetJiraIssueComments(ctx, req.ConnectorID, req.IssueKey)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get Jira issue comments: %v", err),
		}
	}

	// Convert to MCP response format
	mcpIssue := &JiraIssue{
		ID:          issue.ID,
		Key:         issue.Key,
		Summary:     issue.Summary,
		Description: issue.Description,
		Status:      issue.Status,
		Priority:    issue.Priority,
		IssueType:   issue.IssueType,
		Assignee:    issue.Assignee,
		Reporter:    issue.Reporter,
		Labels:      issue.Labels,
		Components:  issue.Components,
		FixVersions: issue.FixVersions,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
		Project:     issue.Project,
	}
	if issue.ResolvedAt != nil {
		mcpIssue.ResolvedAt = *issue.ResolvedAt
	}

	mcpComments := make([]JiraComment, 0, len(comments))
	for _, comment := range comments {
		mcpComments = append(mcpComments, JiraComment{
			ID:        comment.ID,
			IssueKey:  comment.IssueKey,
			Author:    comment.Author,
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	return JiraGetIssueResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Retrieved issue %s with %d comments", req.IssueKey, len(mcpComments)),
		Issue:    mcpIssue,
		Comments: mcpComments,
	}, nil
}

// handleJiraListProjects implements the jira.list_projects tool
func (s *Server) handleJiraListProjects(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req JiraListProjectsRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}

	// Check if connector exists and is Jira type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "jira" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Jira connector", req.ConnectorID),
		}
	}

	// List projects
	projects, err := s.connectorManager.ListJiraProjects(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to list Jira projects: %v", err),
		}
	}

	// Convert to MCP response format
	mcpProjects := make([]JiraProject, 0, len(projects))
	for _, proj := range projects {
		mcpProjects = append(mcpProjects, JiraProject{
			ID:          proj.ID,
			Key:         proj.Key,
			Name:        proj.Name,
			Description: proj.Description,
			Lead:        proj.Lead,
			Type:        proj.Type,
		})
	}

	return JiraListProjectsResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Found %d projects", len(mcpProjects)),
		Projects: mcpProjects,
	}, nil
}

// handleDiscordSearch implements the discord.search tool
func (s *Server) handleDiscordSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req DiscordSearchRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.ChannelID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "channel_id is required",
		}
	}
	if req.Query == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "query is required",
		}
	}

	// Check if connector exists and is Discord type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "discord" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Discord connector", req.ConnectorID),
		}
	}

	// Search messages
	messages, err := s.connectorManager.SearchDiscordMessages(ctx, req.ConnectorID, req.ChannelID, req.Query)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to search Discord messages: %v", err),
		}
	}

	// Convert to MCP response format
	mcpMessages := make([]DiscordMessage, 0, len(messages))
	for _, msg := range messages {
		mcpMessages = append(mcpMessages, DiscordMessage{
			ID:        msg.ID,
			ChannelID: msg.ChannelID,
			GuildID:   msg.GuildID,
			Author:    msg.Author,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
			IsBot:     false, // discord.Message doesn't have IsBot field
		})
	}

	return DiscordSearchResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Found %d messages", len(mcpMessages)),
		Messages: mcpMessages,
	}, nil
}

// handleDiscordListChannels implements the discord.list_channels tool
func (s *Server) handleDiscordListChannels(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req DiscordListChannelsRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}

	// Check if connector exists and is Discord type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "discord" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Discord connector", req.ConnectorID),
		}
	}

	// List channels
	channels, err := s.connectorManager.ListDiscordChannels(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to list Discord channels: %v", err),
		}
	}

	// Convert to MCP response format
	mcpChannels := make([]DiscordChannel, 0, len(channels))
	for _, ch := range channels {
		mcpChannels = append(mcpChannels, DiscordChannel{
			ID:       ch.ID,
			Name:     ch.Name,
			Type:     ch.Type,
			Topic:    ch.Topic,
			Position: ch.Position,
		})
	}

	return DiscordListChannelsResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Found %d channels", len(mcpChannels)),
		Channels: mcpChannels,
	}, nil
}

// handleDiscordGetThread implements the discord.get_thread tool
func (s *Server) handleDiscordGetThread(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req DiscordGetThreadRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.ThreadID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "thread_id is required",
		}
	}

	// Check if connector exists and is Discord type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "discord" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a Discord connector", req.ConnectorID),
		}
	}

	// Get thread messages
	messages, err := s.connectorManager.GetDiscordThread(ctx, req.ConnectorID, req.ThreadID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get Discord thread: %v", err),
		}
	}

	// Convert to MCP response format
	mcpMessages := make([]DiscordMessage, 0, len(messages))
	for _, msg := range messages {
		mcpMessages = append(mcpMessages, DiscordMessage{
			ID:        msg.ID,
			ChannelID: msg.ChannelID,
			GuildID:   msg.GuildID,
			Author:    msg.Author,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
			IsBot:     false, // discord.Message doesn't have IsBot field
		})
	}

	return DiscordGetThreadResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Found thread with %d messages", len(mcpMessages)),
		Messages: mcpMessages,
	}, nil
}

// getStringFromMetadata safely extracts string from metadata
func getStringFromMetadata(metadata map[string]interface{}, key string) string {
	if value, ok := metadata[key].(string); ok {
		return value
	}
	return ""
}

// handleGitHubSearchIssues implements the github.search_issues tool
func (s *Server) handleGitHubSearchIssues(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GitHubSearchIssuesRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.Query == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "query is required",
		}
	}

	// Check if connector exists and is GitHub type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "github" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a GitHub connector", req.ConnectorID),
		}
	}

	// Search issues
	issues, err := s.connectorManager.SearchGitHubIssues(ctx, req.ConnectorID, req.Query, req.State)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to search GitHub issues: %v", err),
		}
	}

	// Convert to MCP response format
	mcpIssues := make([]GitHubIssue, 0, len(issues))
	for _, issue := range issues {
		mcpIssues = append(mcpIssues, GitHubIssue{
			ID:          issue.ID,
			Number:      issue.Number,
			Title:       issue.Title,
			Description: issue.Description,
			State:       issue.State,
			Labels:      issue.Labels,
			Assignee:    issue.Assignee,
			CreatedAt:   issue.CreatedAt,
			UpdatedAt:   issue.UpdatedAt,
		})
	}

	return GitHubSearchIssuesResponse{
		Status:  "ok",
		Message: fmt.Sprintf("Found %d issues", len(mcpIssues)),
		Issues:  mcpIssues,
	}, nil
}

// handleGitHubGetIssue implements the github.get_issue tool
func (s *Server) handleGitHubGetIssue(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GitHubGetIssueRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.IssueNumber <= 0 {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "issue_number is required and must be positive",
		}
	}

	// Check if connector exists and is GitHub type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "github" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a GitHub connector", req.ConnectorID),
		}
	}

	// Get issue
	issue, err := s.connectorManager.GetGitHubIssue(ctx, req.ConnectorID, req.IssueNumber)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get GitHub issue: %v", err),
		}
	}

	// Get comments
	comments, err := s.connectorManager.GetGitHubIssueComments(ctx, req.ConnectorID, req.IssueNumber)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get GitHub issue comments: %v", err),
		}
	}

	// Convert to MCP response format
	mcpIssue := &GitHubIssue{
		ID:          issue.ID,
		Number:      issue.Number,
		Title:       issue.Title,
		Description: issue.Description,
		State:       issue.State,
		Labels:      issue.Labels,
		Assignee:    issue.Assignee,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
	}

	mcpComments := make([]GitHubComment, 0, len(comments))
	for _, comment := range comments {
		mcpComments = append(mcpComments, GitHubComment{
			ID:        comment.ID,
			Author:    comment.Author,
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	return GitHubGetIssueResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Retrieved issue #%d with %d comments", req.IssueNumber, len(mcpComments)),
		Issue:    mcpIssue,
		Comments: mcpComments,
	}, nil
}

// handleGitHubGetPR implements the github.get_pr tool
func (s *Server) handleGitHubGetPR(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GitHubGetPRRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}
	if req.PRNumber <= 0 {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "pr_number is required and must be positive",
		}
	}

	// Check if connector exists and is GitHub type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "github" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a GitHub connector", req.ConnectorID),
		}
	}

	// Get pull request
	pr, err := s.connectorManager.GetGitHubPullRequest(ctx, req.ConnectorID, req.PRNumber)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get GitHub pull request: %v", err),
		}
	}

	// Get comments
	comments, err := s.connectorManager.GetGitHubPRComments(ctx, req.ConnectorID, req.PRNumber)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get GitHub PR comments: %v", err),
		}
	}

	// Convert to MCP response format
	mcpPR := &GitHubPullRequest{
		ID:           pr.ID,
		Number:       pr.Number,
		Title:        pr.Title,
		Description:  pr.Description,
		State:        pr.State,
		Labels:       pr.Labels,
		Assignee:     pr.Assignee,
		HeadBranch:   pr.HeadBranch,
		BaseBranch:   pr.BaseBranch,
		Merged:       pr.Merged,
		LinkedIssues: pr.LinkedIssues,
		CreatedAt:    pr.CreatedAt,
		UpdatedAt:    pr.UpdatedAt,
	}

	mcpComments := make([]GitHubComment, 0, len(comments))
	for _, comment := range comments {
		mcpComments = append(mcpComments, GitHubComment{
			ID:        comment.ID,
			Author:    comment.Author,
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		})
	}

	return GitHubGetPRResponse{
		Status:   "ok",
		Message:  fmt.Sprintf("Retrieved PR #%d with %d comments", req.PRNumber, len(mcpComments)),
		PR:       mcpPR,
		Comments: mcpComments,
	}, nil
}

// handleGitHubListRepos implements the github.list_repos tool
func (s *Server) handleGitHubListRepos(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req GitHubListReposRequest
	if err := json.Unmarshal(args, &req); err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid request: %v", err),
		}
	}

	// Validate required fields
	if req.ConnectorID == "" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: "connector_id is required",
		}
	}

	// Check if connector exists and is GitHub type
	connector, err := s.connectorStore.Get(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to get connector: %v", err),
		}
	}

	if connector.Type != "github" {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("connector %s is not a GitHub connector", req.ConnectorID),
		}
	}

	// List repositories
	repos, err := s.connectorManager.ListGitHubRepositories(ctx, req.ConnectorID)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to list GitHub repositories: %v", err),
		}
	}

	// Convert to MCP response format
	mcpRepos := make([]GitHubRepository, 0, len(repos))
	for _, repo := range repos {
		mcpRepos = append(mcpRepos, GitHubRepository{
			ID:          repo.ID,
			Name:        repo.Name,
			FullName:    repo.FullName,
			Description: repo.Description,
			URL:         repo.URL,
			Private:     repo.Private,
			Language:    repo.Language,
			Stars:       repo.Stars,
			Forks:       repo.Forks,
			CreatedAt:   repo.CreatedAt,
			UpdatedAt:   repo.UpdatedAt,
		})
	}

	return GitHubListReposResponse{
		Status:       "ok",
		Message:      fmt.Sprintf("Found %d repositories", len(mcpRepos)),
		Repositories: mcpRepos,
	}, nil
}
