package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/ferg-cod3s/conexus/internal/protocol"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// handleContextSearch implements the context.search tool
func (s *Server) handleContextSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var req SearchRequest
	if err := json.Unmarshal(args, &req); err != nil {
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
	
	startTime := time.Now()
	
	// Generate query embedding
	queryVec, err := s.embedder.Embed(ctx, req.Query)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("failed to generate query embedding: %v", err),
		}
	}
	
	// Prepare search options
	opts := vectorstore.SearchOptions{
		Limit: topK,
		Filters: make(map[string]interface{}),
	}
	
	// Apply filters if provided
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
	}
	
	// Perform hybrid search (combines vector + BM25)
	results, err := s.vectorStore.SearchHybrid(ctx, req.Query, queryVec.Vector, opts)
	if err != nil {
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: fmt.Sprintf("search failed: %v", err),
		}
	}
	
	queryTime := time.Since(startTime).Milliseconds()
	
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
		QueryTime:  float64(queryTime),
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
	
	// Build search query based on provided identifiers
	var query string
	if req.FilePath != "" {
		query = fmt.Sprintf("file:%s", req.FilePath)
	} else {
		query = fmt.Sprintf("ticket:%s", req.TicketID)
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
	
	// Group results by type
	var relatedPRs, relatedIssues []string
	var discussions []DiscussionSummary
	
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
	
	// Generate summary
	summary := fmt.Sprintf("Found %d related items", len(results))
	if req.FilePath != "" {
		summary = fmt.Sprintf("Related information for %s: %d PRs, %d issues, %d discussions",
			req.FilePath, len(relatedPRs), len(relatedIssues), len(discussions))
	} else {
		summary = fmt.Sprintf("Related information for ticket %s: %d PRs, %d issues, %d discussions",
			req.TicketID, len(relatedPRs), len(relatedIssues), len(discussions))
	}
	
	return GetRelatedInfoResponse{
		Summary:       summary,
		RelatedPRs:    relatedPRs,
		RelatedIssues: relatedIssues,
		Discussions:   discussions,
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
	}
	
	if !validActions[req.Action] {
		return nil, &protocol.Error{
			Code:    protocol.InvalidParams,
			Message: fmt.Sprintf("invalid action: %s", req.Action),
		}
	}
	
	// For now, handle status action (others will be implemented with indexer integration)
	switch req.Action {
	case "status":
		count, err := s.vectorStore.Count(ctx)
		if err != nil {
			return nil, &protocol.Error{
				Code:    protocol.InternalError,
				Message: fmt.Sprintf("failed to get status: %v", err),
			}
		}
		
		details := map[string]string{
			"documents_indexed": fmt.Sprintf("%d", count),
			"status":            "active",
		}
		
		return IndexControlResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Index contains %d documents", count),
			Details: details,
		}, nil
		
	default:
		// Placeholder for other actions
		return IndexControlResponse{
			Status:  "pending",
			Message: fmt.Sprintf("Action '%s' queued for execution", req.Action),
		}, nil
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
	
	// For now, return placeholder responses (will be implemented with connector system)
	switch req.Action {
	case "list":
		return ConnectorManagementResponse{
			Connectors: []ConnectorInfo{
				{
					ID:     "local-files",
					Type:   "filesystem",
					Name:   "Local Files",
					Status: "active",
					Config: map[string]interface{}{
						"path": ".",
					},
				},
			},
			Status:  "ok",
			Message: "Retrieved connector list",
		}, nil
		
	case "add", "update", "remove":
		if req.Action != "list" && req.ConnectorID == "" {
			return nil, &protocol.Error{
				Code:    protocol.InvalidParams,
				Message: "connector_id is required",
			}
		}
		
		return ConnectorManagementResponse{
			Status:  "ok",
			Message: fmt.Sprintf("Connector %s: %s", req.Action, req.ConnectorID),
		}, nil
		
	default:
		return nil, &protocol.Error{
			Code:    protocol.InternalError,
			Message: "unexpected error",
		}
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
