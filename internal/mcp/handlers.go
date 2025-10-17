package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
	"strings"

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

	// Route to appropriate flow
	if req.FilePath != "" {
		return s.handleFilePathFlow(ctx, req)
	}
	return s.handleTicketIDFlow(ctx, req)
}

// handleFilePathFlow implements the file path-based relationship discovery
func (s *Server) handleFilePathFlow(ctx context.Context, req GetRelatedInfoRequest) (*GetRelatedInfoResponse, error) {
	detector := NewRelationshipDetector(req.FilePath)

	// Step 1: Get chunks for the source file (future optimization: use for symbol extraction)
	_, err := s.vectorStore.GetFileChunks(ctx, req.FilePath)
	if err != nil {
		// File not found is acceptable - we can still find related files
		// Log but don't fail
	}

	// Step 2: Get all indexed files for relationship detection
	allFiles, err := s.vectorStore.ListIndexedFiles(ctx)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to list indexed files: %v", err),
		}
	}

	// Step 3: Find related files and score them
	type relatedFileScore struct {
		filePath     string
		relationType string
		score        float32
		chunks       []vectorstore.Document
	}

	relatedFiles := make(map[string]*relatedFileScore)
	
	for _, candidateFile := range allFiles {
		// Skip the source file itself
		if candidateFile == req.FilePath {
			continue
		}

		// Detect relationship type
		// Note: We pass empty chunkType and nil metadata here since we're checking file-level
		// relationships. Chunk-level relationships are detected later.
		relationType := detector.DetectRelationType(candidateFile, "", nil)
		
		if relationType != "" {
			// Get chunks for this related file
			chunks, err := s.vectorStore.GetFileChunks(ctx, candidateFile)
			if err != nil {
				// Log error but continue with other files
				continue
			}

			// Calculate base score from relationship type
			score := s.getRelationshipScore(relationType)
			
			relatedFiles[candidateFile] = &relatedFileScore{
				filePath:     candidateFile,
				relationType: relationType,
				score:        score,
				chunks:       chunks,
			}
		}
	}

	// Step 4: Build RelatedItems from related file chunks
	relatedItems := make([]RelatedItem, 0)
	
	for _, rf := range relatedFiles {
		for _, chunk := range rf.chunks {
			// Extract metadata
			sourceType, _ := chunk.Metadata["source_type"].(string)
			startLine, _ := s.extractLineNumber(chunk.Metadata, "start_line")
			endLine, _ := s.extractLineNumber(chunk.Metadata, "end_line")
			chunkType, _ := chunk.Metadata["type"].(string)

			// Refine relationship type at chunk level if needed
			chunkRelationType := detector.DetectRelationType(rf.filePath, chunkType, chunk.Metadata)
			if chunkRelationType == "" {
				chunkRelationType = rf.relationType
			}

			// Adjust score based on chunk-level relationship
			chunkScore := rf.score
			if chunkRelationType != rf.relationType {
				chunkScore = s.getRelationshipScore(chunkRelationType)
			}

			relatedItems = append(relatedItems, RelatedItem{
				ID:           chunk.ID,
				Content:      chunk.Content,
				Score:        chunkScore,
				SourceType:   sourceType,
				FilePath:     rf.filePath,
				RelationType: chunkRelationType,
				StartLine:    startLine,
				EndLine:      endLine,
				Metadata:     chunk.Metadata,
			})
		}
	}

	// Step 5: Sort by score (descending) and relationship priority
	sort.Slice(relatedItems, func(i, j int) bool {
		// Primary sort by score
		if relatedItems[i].Score != relatedItems[j].Score {
			return relatedItems[i].Score > relatedItems[j].Score
		}
		// Secondary sort by relationship type priority
		return s.getRelationshipPriority(relatedItems[i].RelationType) < 
		       s.getRelationshipPriority(relatedItems[j].RelationType)
	})

	// Step 6: Limit results to top N (default 50)
	limit := 50
	if len(relatedItems) > limit {
		relatedItems = relatedItems[:limit]
	}

	// Step 7: Build response with summaries
	var relatedPRs, relatedIssues []string
	var discussions []DiscussionSummary
	fileCount := len(relatedFiles)

	for _, item := range relatedItems {
		switch item.SourceType {
		case "github_pr":
			if prNum, ok := item.Metadata["pr_number"].(string); ok {
				relatedPRs = append(relatedPRs, prNum)
			}
		case "github_issue", "jira":
			if issueID, ok := item.Metadata["issue_id"].(string); ok {
				relatedIssues = append(relatedIssues, issueID)
			}
		case "slack":
			channel, _ := item.Metadata["channel"].(string)
			timestamp, _ := item.Metadata["timestamp"].(string)
			discussions = append(discussions, DiscussionSummary{
				Channel:   channel,
				Timestamp: timestamp,
				Summary:   item.Content[:min(200, len(item.Content))],
			})
		}
	}

	summary := fmt.Sprintf("Found %d related files with %d chunks for %s (%d PRs, %d issues, %d discussions)",
		fileCount, len(relatedItems), req.FilePath, len(relatedPRs), len(relatedIssues), len(discussions))

	return &GetRelatedInfoResponse{
		Summary:       summary,
		RelatedPRs:    relatedPRs,
		RelatedIssues: relatedIssues,
		Discussions:   discussions,
		RelatedItems:  relatedItems,
	}, nil
}


// handleTicketIDFlow implements ticket ID-based relationship discovery
func (s *Server) handleTicketIDFlow(ctx context.Context, req GetRelatedInfoRequest) (*GetRelatedInfoResponse, error) {
	// Step 1: Get repository root
	repoRoot, err := getRepoRoot(req.FilePath)
	if err != nil {
		// If not in a git repo, fall back to error
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("not in a git repository: %v", err),
		}
	}

	// Step 2: Find ticket in git history
	gitInfo, err := s.findTicketInGit(ctx, req.TicketID, repoRoot)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to search git history: %v", err),
		}
	}

	// Step 3: Check if we found any matches
	if len(gitInfo.Branches) == 0 && len(gitInfo.Commits) == 0 {
		return &GetRelatedInfoResponse{
			Summary:       fmt.Sprintf("No git history found for ticket %s", req.TicketID),
			RelatedPRs:    []string{},
			RelatedIssues: []string{},
			Discussions:   []DiscussionSummary{},
			RelatedItems:  []RelatedItem{},
		}, nil
	}

	// Step 4: Query vector store for each modified file to get context
	relatedItems := make([]RelatedItem, 0)
	filesSeen := make(map[string]bool)
	
	for _, filePath := range gitInfo.ModifiedFiles {
		if filesSeen[filePath] {
			continue
		}
		filesSeen[filePath] = true

		// Query vector store for this file
		queryVec, err := s.embedder.Embed(ctx, fmt.Sprintf("file:%s", filePath))
		if err != nil {
			continue // Skip files we can't embed
		}

		opts := vectorstore.SearchOptions{
			Limit: 5, // Limit per file to avoid overwhelming results
			Filters: map[string]interface{}{
				"file_path": filePath,
			},
		}

		results, err := s.vectorStore.SearchHybrid(ctx, filePath, queryVec.Vector, opts)
		if err != nil {
			continue // Skip files with search errors
		}

		// Add chunks for this file
		for _, r := range results {
			sourceType, _ := r.Document.Metadata["source_type"].(string)
			startLine, _ := s.extractLineNumber(r.Document.Metadata, "start_line")
			endLine, _ := s.extractLineNumber(r.Document.Metadata, "end_line")

			relatedItems = append(relatedItems, RelatedItem{
				ID:         r.Document.ID,
				Content:    r.Document.Content,
				Score:      r.Score + 0.3, // Boost score since from git history
				SourceType: sourceType,
				FilePath:   filePath,
				StartLine:  startLine,
				EndLine:    endLine,
				Metadata:   r.Document.Metadata,
			})
		}
	}

	// Step 5: Search for PR descriptions and issue metadata in vector store
	var relatedPRs, relatedIssues []string
	var discussions []DiscussionSummary

	// Search for ticket ID in PR/issue metadata
	queryVec, err := s.embedder.Embed(ctx, fmt.Sprintf("ticket:%s", req.TicketID))
	if err == nil {
		opts := vectorstore.SearchOptions{
			Limit: 20,
		}

		results, err := s.vectorStore.SearchHybrid(ctx, req.TicketID, queryVec.Vector, opts)
		if err == nil {
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
			}
		}
	}

	// Step 6: Build summary
	summary := fmt.Sprintf(
		"Ticket %s: found in %d branches, %d commits, %d modified files. Related: %d PRs, %d issues, %d discussions",
		req.TicketID,
		len(gitInfo.Branches),
		len(gitInfo.Commits),
		len(gitInfo.ModifiedFiles),
		len(relatedPRs),
		len(relatedIssues),
		len(discussions),
	)

	// Add git commit information to summary
	if len(gitInfo.Commits) > 0 {
		summary += fmt.Sprintf("\n\nRecent commits:\n")
		for i, commit := range gitInfo.Commits {
			if i >= 5 { // Limit to 5 most recent
				break
			}
			summary += fmt.Sprintf("- %s: %s (%s)\n", 
				commit.Hash[:8], 
				commit.Message[:min(80, len(commit.Message))],
				commit.Author,
			)
		}
	}

	if len(gitInfo.Branches) > 0 {
		summary += fmt.Sprintf("\nBranches: %s", strings.Join(gitInfo.Branches, ", "))
	}

	return &GetRelatedInfoResponse{
		Summary:       summary,
		RelatedPRs:    relatedPRs,
		RelatedIssues: relatedIssues,
		Discussions:   discussions,
		RelatedItems:  relatedItems,
	}, nil
}

// getRelationshipScore returns a score for a relationship type
func (s *Server) getRelationshipScore(relationType string) float32 {
	switch relationType {
	case RelationTypeTestFile:
		return 1.0
	case RelationTypeDocumentation:
		return 0.9
	case RelationTypeSymbolRef:
		return 0.8
	case RelationTypeImport:
		return 0.7
	case RelationTypeCommitHistory:
		return 0.6
	case RelationTypeSimilarCode:
		return 0.5
	default:
		return 0.3
	}
}

// getRelationshipPriority returns priority order for sorting (lower is higher priority)
func (s *Server) getRelationshipPriority(relationType string) int {
	switch relationType {
	case RelationTypeTestFile:
		return 1
	case RelationTypeDocumentation:
		return 2
	case RelationTypeSymbolRef:
		return 3
	case RelationTypeImport:
		return 4
	case RelationTypeCommitHistory:
		return 5
	case RelationTypeSimilarCode:
		return 6
	default:
		return 99
	}
}

// extractLineNumber safely extracts line numbers from metadata
func (s *Server) extractLineNumber(metadata map[string]interface{}, key string) (int, bool) {
	if val, ok := metadata[key]; ok {
		switch v := val.(type) {
		case int:
			return v, true
		case float64:
			return int(v), true
		case int64:
			return int(v), true
		}
	}
	return 0, false
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
