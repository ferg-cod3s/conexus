// Package schema defines shared types for search operations across the application.
package schema

// SearchRequest represents a search request for federation
type SearchRequest struct {
	Query       string                 `json:"query"`
	TopK        int                    `json:"top_k,omitempty"`
	Offset      int                    `json:"offset,omitempty"`
	Filters     *SearchFilters         `json:"filters,omitempty"`
	WorkContext *WorkContextFilters    `json:"work_context,omitempty"`
}

// WorkContext provides information about the user's current working context
type WorkContext struct {
	ActiveFile    string   `json:"active_file,omitempty"`
	GitBranch     string   `json:"git_branch,omitempty"`
	OpenTicketIDs []string `json:"open_ticket_ids,omitempty"`
}

// SearchFilters represents filters for search
type SearchFilters struct {
	SourceTypes   []string             `json:"source_types,omitempty"`
	DateRange     *DateRange           `json:"date_range,omitempty"`
	WorkContext   *WorkContextFilters  `json:"work_context,omitempty"`
}

// DateRange represents a date range filter
type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// WorkContextFilters represents work context filters
type WorkContextFilters struct {
	ActiveFile     string   `json:"active_file,omitempty"`
	GitBranch      string   `json:"git_branch,omitempty"`
	OpenTicketIDs  []string `json:"open_ticket_ids,omitempty"`
	BoostActive    bool     `json:"boost_active,omitempty"`
}

// SearchResponse represents a search response from federation
type SearchResponse struct {
	Results    []SearchResultItem `json:"results"`
	TotalCount int                 `json:"total_count"`
	QueryTime  float64             `json:"query_time"`
	Offset     int                 `json:"offset"`
	Limit      int                 `json:"limit"`
	HasMore    bool                `json:"has_more"`
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	Score      float32                `json:"score"`
	SourceType string                 `json:"source_type"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
