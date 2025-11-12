// Package mcp implements the Model Context Protocol server for Conexus.
package mcp

import (
	"encoding/json"
	"time"
)

// Tool names exposed by the MCP server
const (
	ToolContextSearch              = "context.search"
	ToolContextGetRelatedInfo      = "context.get_related_info"
	ToolContextIndexControl        = "context.index_control"
	ToolContextConnectorManagement = "context.connector_management"
	ToolContextExplain             = "context.explain"
	ToolContextGrep                = "context.grep"
	ToolGitHubSyncStatus           = "github.sync_status"
	ToolGitHubSyncTrigger          = "github.sync_trigger"
	ToolGitHubSearchIssues         = "github.search_issues"
	ToolGitHubGetIssue             = "github.get_issue"
	ToolGitHubGetPR                = "github.get_pr"
	ToolGitHubListRepos            = "github.list_repos"
	// Slack connector tools
	ToolSlackSearch       = "slack.search"
	ToolSlackListChannels = "slack.list_channels"
	ToolSlackGetThread    = "slack.get_thread"
	// Jira connector tools
	ToolJiraSearch       = "jira.search"
	ToolJiraGetIssue     = "jira.get_issue"
	ToolJiraListProjects = "jira.list_projects"
	// Discord connector tools
	ToolDiscordSearch       = "discord.search"
	ToolDiscordListChannels = "discord.list_channels"
	ToolDiscordGetThread    = "discord.get_thread"
)

// Resource URI scheme
const (
	ResourceScheme = "engine"
	ResourceFiles  = "files"
)

// SearchRequest represents the input for context.search tool
type SearchRequest struct {
	Query       string         `json:"query"`
	WorkContext *WorkContext   `json:"work_context,omitempty"`
	TopK        int            `json:"top_k,omitempty"`
	Offset      int            `json:"offset,omitempty"` // For pagination
	Filters     *SearchFilters `json:"filters,omitempty"`
}

// WorkContext provides information about the user's current working context
type WorkContext struct {
	ActiveFile    string   `json:"active_file,omitempty"`
	GitBranch     string   `json:"git_branch,omitempty"`
	OpenTicketIDs []string `json:"open_ticket_ids,omitempty"`
}

// SearchFilters defines filtering options for search
type SearchFilters struct {
	SourceTypes []string            `json:"source_types,omitempty"`
	DateRange   *DateRange          `json:"date_range,omitempty"`
	WorkContext *WorkContextFilters `json:"work_context,omitempty"`
}

// WorkContextFilters defines filters based on work context
type WorkContextFilters struct {
	ActiveFile     string   `json:"active_file,omitempty"`
	GitBranch      string   `json:"git_branch,omitempty"`
	OpenTicketIDs  []string `json:"open_ticket_ids,omitempty"`
	CurrentStoryID string   `json:"current_story_id,omitempty"`
	BoostActive    bool     `json:"boost_active,omitempty"` // Boost results related to active file/tickets
}

// DateRange specifies a time range filter
type DateRange struct {
	From string `json:"from,omitempty"` // ISO 8601 date-time
	To   string `json:"to,omitempty"`   // ISO 8601 date-time
}

// SearchResponse represents the output of context.search tool
type SearchResponse struct {
	Results    []SearchResultItem `json:"results"`
	TotalCount int                `json:"total_count"`
	QueryTime  float64            `json:"query_time_ms"`
	Offset     int                `json:"offset,omitempty"`
	Limit      int                `json:"limit,omitempty"`
	HasMore    bool               `json:"has_more,omitempty"`
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	Score      float32                `json:"score"`
	SourceType string                 `json:"source_type"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// GetRelatedInfoRequest represents the input for context.get_related_info tool
type GetRelatedInfoRequest struct {
	FilePath string `json:"file_path,omitempty"`
	TicketID string `json:"ticket_id,omitempty"`
}

// RelatedItem represents a single related item with relevance score
type RelatedItem struct {
	ID         string                 `json:"id"`
	Content    string                 `json:"content"`
	Score      float32                `json:"score"`
	SourceType string                 `json:"source_type"`
	FilePath   string                 `json:"file_path,omitempty"`
	StartLine  int                    `json:"start_line,omitempty"`
	EndLine    int                    `json:"end_line,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// GetRelatedInfoResponse represents the output of context.get_related_info tool
type GetRelatedInfoResponse struct {
	Summary       string              `json:"summary"`
	RelatedItems  []RelatedItem       `json:"related_items"`
	RelatedPRs    []string            `json:"related_prs,omitempty"`
	RelatedIssues []string            `json:"related_issues,omitempty"`
	Discussions   []DiscussionSummary `json:"discussions,omitempty"`
}

// DiscussionSummary provides a summary of a Slack discussion
type DiscussionSummary struct {
	Channel   string `json:"channel"`
	Timestamp string `json:"timestamp"`
	Summary   string `json:"summary"`
}

// IndexContent represents content to be indexed
type IndexContent struct {
	Path       string `json:"path"`                 // File path
	Content    string `json:"content"`              // File content
	SourceType string `json:"source_type"`          // Type of source (file, ticket, etc.)
	StartLine  *int   `json:"start_line,omitempty"` // Optional start line
	EndLine    *int   `json:"end_line,omitempty"`   // Optional end line
}

// IndexControlRequest represents the input for context.index_control tool
type IndexControlRequest struct {
	Action      string        `json:"action"`                 // "start", "stop", "status", "force_reindex", "reindex_paths", "index", "sync_github"
	Connectors  []string      `json:"connectors,omitempty"`   // Connectors to use for indexing
	ConnectorID string        `json:"connector_id,omitempty"` // Specific connector ID (for sync_github action)
	Paths       []string      `json:"paths,omitempty"`        // Specific paths/files to reindex (for reindex_paths action)
	Content     *IndexContent `json:"content,omitempty"`      // Content to index (for index action)
}

// IndexControlResponse represents the output of context.index_control tool
type IndexControlResponse struct {
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	IndexStatus *IndexStatus           `json:"index_status,omitempty"`
}

// IndexStatus represents the current status of indexing operations
type IndexStatus struct {
	IsIndexing     bool          `json:"is_indexing"`
	Phase          string        `json:"phase"`
	Progress       float64       `json:"progress"`
	FilesProcessed int           `json:"files_processed"`
	TotalFiles     int           `json:"total_files"`
	ChunksCreated  int           `json:"chunks_created"`
	StartTime      string        `json:"start_time,omitempty"`
	EstimatedEnd   string        `json:"estimated_end,omitempty"`
	LastError      string        `json:"last_error,omitempty"`
	Metrics        *IndexMetrics `json:"metrics,omitempty"`
}

// IndexMetrics provides statistics about indexing operations
type IndexMetrics struct {
	TotalFiles      int     `json:"total_files"`
	IndexedFiles    int     `json:"indexed_files"`
	SkippedFiles    int     `json:"skipped_files"`
	TotalChunks     int     `json:"total_chunks"`
	Duration        float64 `json:"duration_seconds"`
	BytesProcessed  int64   `json:"bytes_processed"`
	StateSize       int64   `json:"state_size_bytes"`
	IncrementalSave float64 `json:"incremental_save_seconds"`
}

// ConnectorManagementRequest represents the input for context.connector_management tool
type ConnectorManagementRequest struct {
	Action          string                 `json:"action"` // "list", "add", "update", "remove"
	ConnectorID     string                 `json:"connector_id,omitempty"`
	ConnectorConfig map[string]interface{} `json:"connector_config,omitempty"`
}

// ConnectorManagementResponse represents the output of context.connector_management tool
type ConnectorManagementResponse struct {
	Connectors []ConnectorInfo `json:"connectors"`
	Status     string          `json:"status,omitempty"`
	Message    string          `json:"message,omitempty"`
}

// ConnectorInfo provides information about a connector
type ConnectorInfo struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Status string                 `json:"status"`
	Config map[string]interface{} `json:"config"`
}

// ExplainRequest represents the input for context.explain tool
type ExplainRequest struct {
	Target  string `json:"target"`            // The code, function name, or concept to explain
	Context string `json:"context,omitempty"` // Additional context about what aspect to focus on
	Depth   string `json:"depth,omitempty"`   // "brief", "detailed", "comprehensive"
}

// ExplainResponse represents the output of context.explain tool
type ExplainResponse struct {
	Explanation string                 `json:"explanation"`
	Examples    []CodeExample          `json:"examples,omitempty"`
	Related     []RelatedItem          `json:"related,omitempty"`
	Complexity  string                 `json:"complexity"` // "simple", "moderate", "complex"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CodeExample provides a code example with explanation
type CodeExample struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Language    string `json:"language"`
}

// GrepRequest represents the input for context.grep tool
type GrepRequest struct {
	Pattern         string `json:"pattern"`                    // The pattern to search for (supports regex)
	Path            string `json:"path,omitempty"`             // Base directory to search in
	Include         string `json:"include,omitempty"`          // File pattern to include
	CaseInsensitive bool   `json:"case_insensitive,omitempty"` // Case insensitive search
	Context         int    `json:"context,omitempty"`          // Lines of context around matches
}

// GrepResult represents a single grep match
type GrepResult struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
	Match   string `json:"match"`
}

// GrepResponse represents the output of context.grep tool
type GrepResponse struct {
	Results    []GrepResult `json:"results"`
	TotalCount int          `json:"total_count"`
	SearchTime float64      `json:"search_time_ms"`
}

// GitHubSyncStatusRequest represents input for github.sync_status tool
type GitHubSyncStatusRequest struct {
	ConnectorID string `json:"connector_id,omitempty"` // Optional, if empty returns status for all connectors
}

// GitHubSyncStatusResponse represents output of github.sync_status tool
type GitHubSyncStatusResponse struct {
	Status      string                 `json:"status"` // "ok", "error"
	Message     string                 `json:"message"`
	SyncStatus  *SyncStatus            `json:"sync_status,omitempty"`
	ConnectorID string                 `json:"connector_id,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// GitHubSyncTriggerRequest represents input for github.sync_trigger tool
type GitHubSyncTriggerRequest struct {
	ConnectorID string `json:"connector_id"`    // Required
	Force       bool   `json:"force,omitempty"` // Force sync even if recently synced
}

// GitHubSyncTriggerResponse represents output of github.sync_trigger tool
type GitHubSyncTriggerResponse struct {
	Status      string                 `json:"status"` // "ok", "error"
	Message     string                 `json:"message"`
	JobID       string                 `json:"job_id,omitempty"`
	ConnectorID string                 `json:"connector_id"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// SyncStatus represents the sync status for a connector
type SyncStatus struct {
	IsRunning           bool           `json:"is_running"`
	ActiveJobs          []*SyncJob     `json:"active_jobs"`
	CompletedJobs       []*SyncJob     `json:"completed_jobs"`
	LastSyncTime        *time.Time     `json:"last_sync_time,omitempty"`
	TotalSyncs          int            `json:"total_syncs"`
	SuccessfulSyncs     int            `json:"successful_syncs"`
	FailedSyncs         int            `json:"failed_syncs"`
	CurrentSyncProgress float64        `json:"current_sync_progress,omitempty"` // 0.0 to 1.0
	RateLimit           *RateLimitInfo `json:"rate_limit,omitempty"`
}

// SyncJob represents a sync job
type SyncJob struct {
	ID             string     `json:"id"`
	ConnectorID    string     `json:"connector_id"`
	Type           string     `json:"type"`
	Status         string     `json:"status"` // "running", "completed", "failed"
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Progress       float64    `json:"progress"` // 0.0 to 1.0
	TotalItems     int        `json:"total_items"`
	ProcessedItems int        `json:"processed_items"`
	Error          string     `json:"error,omitempty"`
}

// RateLimitInfo represents GitHub API rate limit information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

// GitHubSearchIssuesRequest represents input for github.search_issues tool
type GitHubSearchIssuesRequest struct {
	ConnectorID string `json:"connector_id"`    // Required
	Query       string `json:"query"`           // Search query (supports GitHub search syntax)
	State       string `json:"state,omitempty"` // Filter by state: "open", "closed", "all"
}

// GitHubSearchIssuesResponse represents output of github.search_issues tool
type GitHubSearchIssuesResponse struct {
	Status  string                 `json:"status"` // "ok", "error"
	Message string                 `json:"message"`
	Issues  []GitHubIssue          `json:"issues,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// GitHubIssue represents a GitHub issue
type GitHubIssue struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Labels      []string  `json:"labels,omitempty"`
	Assignee    string    `json:"assignee,omitempty"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    time.Time `json:"closed_at,omitempty"`
	Repository  string    `json:"repository"`
	URL         string    `json:"url"`
}

// GitHubGetIssueRequest represents input for github.get_issue tool
type GitHubGetIssueRequest struct {
	ConnectorID string `json:"connector_id"`         // Required
	IssueNumber int    `json:"issue_number"`         // Required
	Repository  string `json:"repository,omitempty"` // Optional, uses default repo if not specified
}

// GitHubGetIssueResponse represents output of github.get_issue tool
type GitHubGetIssueResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Issue    *GitHubIssue           `json:"issue,omitempty"`
	Comments []GitHubComment        `json:"comments,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// GitHubComment represents a GitHub issue/PR comment
type GitHubComment struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GitHubGetPRRequest represents input for github.get_pr tool
type GitHubGetPRRequest struct {
	ConnectorID string `json:"connector_id"`         // Required
	PRNumber    int    `json:"pr_number"`            // Required
	Repository  string `json:"repository,omitempty"` // Optional, uses default repo if not specified
}

// GitHubGetPRResponse represents output of github.get_pr tool
type GitHubGetPRResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	PR       *GitHubPullRequest     `json:"pr,omitempty"`
	Comments []GitHubComment        `json:"comments,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// GitHubPullRequest represents a GitHub pull request
type GitHubPullRequest struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	State        string    `json:"state"`
	Labels       []string  `json:"labels,omitempty"`
	Assignee     string    `json:"assignee,omitempty"`
	Author       string    `json:"author"`
	HeadBranch   string    `json:"head_branch"`
	BaseBranch   string    `json:"base_branch"`
	Merged       bool      `json:"merged"`
	LinkedIssues []string  `json:"linked_issues,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MergedAt     time.Time `json:"merged_at,omitempty"`
	ClosedAt     time.Time `json:"closed_at,omitempty"`
	Repository   string    `json:"repository"`
	URL          string    `json:"url"`
}

// GitHubListReposRequest represents input for github.list_repos tool
type GitHubListReposRequest struct {
	ConnectorID string `json:"connector_id"` // Required
}

// GitHubListReposResponse represents output of github.list_repos tool
type GitHubListReposResponse struct {
	Status       string                 `json:"status"` // "ok", "error"
	Message      string                 `json:"message"`
	Repositories []GitHubRepository     `json:"repositories,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description,omitempty"`
	Private       bool      `json:"private"`
	DefaultBranch string    `json:"default_branch"`
	Language      string    `json:"language,omitempty"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	OpenIssues    int       `json:"open_issues"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	URL           string    `json:"url"`
}

// SlackSearchRequest represents input for slack.search tool
type SlackSearchRequest struct {
	ConnectorID string `json:"connector_id"` // Required
	Query       string `json:"query"`        // Search query
}

// SlackSearchResponse represents output of slack.search tool
type SlackSearchResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Messages []SlackMessage         `json:"messages,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// SlackMessage represents a Slack message
type SlackMessage struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	UserID    string    `json:"user_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	ThreadTS  string    `json:"thread_ts,omitempty"`
	IsBot     bool      `json:"is_bot"`
}

// SlackListChannelsRequest represents input for slack.list_channels tool
type SlackListChannelsRequest struct {
	ConnectorID string `json:"connector_id"` // Required
}

// SlackListChannelsResponse represents output of slack.list_channels tool
type SlackListChannelsResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Channels []SlackChannel         `json:"channels,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// SlackChannel represents a Slack channel
type SlackChannel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsPrivate   bool   `json:"is_private"`
	MemberCount int    `json:"member_count"`
	Topic       string `json:"topic,omitempty"`
	Purpose     string `json:"purpose,omitempty"`
}

// SlackGetThreadRequest represents input for slack.get_thread tool
type SlackGetThreadRequest struct {
	ConnectorID string `json:"connector_id"` // Required
	ChannelID   string `json:"channel_id"`   // Required
	ThreadTS    string `json:"thread_ts"`    // Required - thread timestamp
}

// SlackGetThreadResponse represents output of slack.get_thread tool
type SlackGetThreadResponse struct {
	Status  string                 `json:"status"` // "ok", "error"
	Message string                 `json:"message"`
	Thread  *SlackThread           `json:"thread,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SlackThread represents a Slack thread with messages
type SlackThread struct {
	ChannelID    string         `json:"channel_id"`
	ThreadTS     string         `json:"thread_ts"`
	MessageCount int            `json:"message_count"`
	Messages     []SlackMessage `json:"messages"`
}

// JiraSearchRequest represents input for jira.search tool
type JiraSearchRequest struct {
	ConnectorID string `json:"connector_id"` // Required
	JQL         string `json:"jql"`          // JQL query string
}

// JiraSearchResponse represents output of jira.search tool
type JiraSearchResponse struct {
	Status  string                 `json:"status"` // "ok", "error"
	Message string                 `json:"message"`
	Issues  []JiraIssue            `json:"issues,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// JiraIssue represents a Jira issue
type JiraIssue struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	IssueType   string    `json:"issue_type"`
	Assignee    string    `json:"assignee"`
	Reporter    string    `json:"reporter"`
	Labels      []string  `json:"labels,omitempty"`
	Components  []string  `json:"components,omitempty"`
	FixVersions []string  `json:"fix_versions,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ResolvedAt  time.Time `json:"resolved_at,omitempty"`
	Project     string    `json:"project"`
}

// JiraGetIssueRequest represents input for jira.get_issue tool
type JiraGetIssueRequest struct {
	ConnectorID string `json:"connector_id"` // Required
	IssueKey    string `json:"issue_key"`    // Required - issue key (e.g., PROJ-123)
}

// JiraGetIssueResponse represents output of jira.get_issue tool
type JiraGetIssueResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Issue    *JiraIssue             `json:"issue,omitempty"`
	Comments []JiraComment          `json:"comments,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// JiraComment represents a Jira issue comment
type JiraComment struct {
	ID        string    `json:"id"`
	IssueKey  string    `json:"issue_key"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JiraListProjectsRequest represents input for jira.list_projects tool
type JiraListProjectsRequest struct {
	ConnectorID string `json:"connector_id"` // Required
}

// JiraListProjectsResponse represents output of jira.list_projects tool
type JiraListProjectsResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Projects []JiraProject          `json:"projects,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// JiraProject represents a Jira project
type JiraProject struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Lead        string `json:"lead,omitempty"`
	Type        string `json:"type"`
}

// DiscordSearchRequest represents input for discord.search tool
type DiscordSearchRequest struct {
	ConnectorID string `json:"connector_id"` // Required
	ChannelID   string `json:"channel_id"`   // Required - channel to search in
	Query       string `json:"query"`        // Search query
}

// DiscordSearchResponse represents output of discord.search tool
type DiscordSearchResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Messages []DiscordMessage       `json:"messages,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// DiscordMessage represents a Discord message
type DiscordMessage struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	GuildID   string    `json:"guild_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsBot     bool      `json:"is_bot"`
}

// DiscordListChannelsRequest represents input for discord.list_channels tool
type DiscordListChannelsRequest struct {
	ConnectorID string `json:"connector_id"` // Required
}

// DiscordListChannelsResponse represents output of discord.list_channels tool
type DiscordListChannelsResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Channels []DiscordChannel       `json:"channels,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// DiscordChannel represents a Discord channel
type DiscordChannel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Topic    string `json:"topic,omitempty"`
	Position int    `json:"position"`
}

// DiscordGetThreadRequest represents input for discord.get_thread tool
type DiscordGetThreadRequest struct {
	ConnectorID string `json:"connector_id"` // Required
	ThreadID    string `json:"thread_id"`    // Required - thread ID
}

// DiscordGetThreadResponse represents output of discord.get_thread tool
type DiscordGetThreadResponse struct {
	Status   string                 `json:"status"` // "ok", "error"
	Message  string                 `json:"message"`
	Messages []DiscordMessage       `json:"messages,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// ToolDefinition represents an MCP tool definition
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ResourceDefinition represents an MCP resource
type ResourceDefinition struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// GetToolDefinitions returns all tool definitions for the MCP server
func GetToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolContextSearch,
			Description: "Performs a comprehensive search using the user's query and current working context to find the most relevant code, discussions, and documents.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {
						"type": "string",
						"description": "The user's natural language query."
					},
					"work_context": {
						"type": "object",
						"properties": {
							"active_file": {"type": "string"},
							"git_branch": {"type": "string"},
							"open_ticket_ids": {"type": "array", "items": {"type": "string"}}
						}
					},
					"top_k": {
						"type": "integer",
						"default": 20,
						"maximum": 100
					},
					"offset": {
						"type": "integer",
						"default": 0,
						"minimum": 0
					},
					"filters": {
						"type": "object",
						"properties": {
							"source_types": {
								"type": "array",
								"items": {"type": "string", "enum": ["file", "slack", "github", "jira"]}
							},
							"date_range": {
								"type": "object",
								"properties": {
									"from": {"type": "string", "format": "date-time"},
									"to": {"type": "string", "format": "date-time"}
								}
							},
							"work_context": {
								"type": "object",
								"properties": {
									"active_file": {"type": "string"},
									"git_branch": {"type": "string"},
									"open_ticket_ids": {"type": "array", "items": {"type": "string"}},
									"boost_active": {"type": "boolean", "default": true}
								}
							}
						}
					}
				},
				"required": ["query"]
			}`),
		},
		{
			Name:        ToolContextGetRelatedInfo,
			Description: "Finds information directly related to the user's active file or ticket. Use this when the user asks a vague question like 'what's the history of this file?'",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"file_path": {
						"type": "string",
						"description": "Path to the file to get related info for"
					},
					"ticket_id": {
						"type": "string",
						"description": "Ticket ID to get related info for"
					}
				}
			}`),
		},
		{
			Name:        ToolContextIndexControl,
			Description: "Control indexing operations",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"action": {
						"type": "string",
						"enum": ["start", "stop", "status", "force_reindex", "reindex_paths"]
					},
					"connectors": {
						"type": "array",
						"items": {"type": "string"}
					},
					"paths": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Specific paths/files to reindex (required for reindex_paths action)"
					}
				},
				"required": ["action"]
			}`),
		},
		{
			Name:        ToolContextConnectorManagement,
			Description: "Manage data source connectors",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"action": {
						"type": "string",
						"enum": ["list", "add", "update", "remove"]
					},
					"connector_id": {"type": "string"},
					"connector_config": {"type": "object"}
				},
				"required": ["action"]
			}`),
		},
		{
			Name:        ToolContextExplain,
			Description: "Provides detailed explanations of code, functions, or concepts found in the codebase. Use this when users need deep understanding of implementation details.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"target": {
						"type": "string",
						"description": "The code, function name, or concept to explain"
					},
					"context": {
						"type": "string",
						"description": "Additional context about what aspect to focus on"
					},
					"depth": {
						"type": "string",
						"enum": ["brief", "detailed", "comprehensive"],
						"default": "detailed"
					}
				},
				"required": ["target"]
			}`),
		},
		{
			Name:        ToolContextGrep,
			Description: "Performs fast, exact pattern matching across codebase using ripgrep. Use this for finding specific strings, function calls, or code patterns.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"pattern": {
						"type": "string",
						"description": "The pattern to search for (supports regex)"
					},
					"path": {
						"type": "string",
						"description": "Base directory to search in (defaults to current directory)"
					},
					"include": {
						"type": "string",
						"description": "File pattern to include (e.g., *.go, *.js)"
					},
					"case_insensitive": {
						"type": "boolean",
						"default": false
					},
					"context": {
						"type": "integer",
						"default": 3,
						"description": "Lines of context to show around matches"
					}
				},
				"required": ["pattern"]
			}`),
		},
		{
			Name:        ToolGitHubSyncStatus,
			Description: "Get GitHub synchronization status for connectors",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "GitHub connector ID (optional, returns status for all if not provided)"
					}
				}
			}`),
		},
		{
			Name:        ToolGitHubSyncTrigger,
			Description: "Trigger manual GitHub synchronization for a specific connector",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "GitHub connector ID to sync"
					},
					"force": {
						"type": "boolean",
						"default": false,
						"description": "Force sync even if recently synced"
					}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolGitHubSearchIssues,
			Description: "Search for GitHub issues using query syntax (supports filters like label:bug, is:open, etc.)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "GitHub connector ID"
					},
					"query": {
						"type": "string",
						"description": "Search query (supports GitHub search syntax)"
					},
					"state": {
						"type": "string",
						"enum": ["open", "closed", "all"],
						"default": "open",
						"description": "Filter by issue state"
					}
				},
				"required": ["connector_id", "query"]
			}`),
		},
		{
			Name:        ToolGitHubGetIssue,
			Description: "Retrieve a specific GitHub issue by number, including all comments",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "GitHub connector ID"
					},
					"issue_number": {
						"type": "integer",
						"description": "Issue number"
					},
					"repository": {
						"type": "string",
						"description": "Repository name (optional, uses default if not specified)"
					}
				},
				"required": ["connector_id", "issue_number"]
			}`),
		},
		{
			Name:        ToolGitHubGetPR,
			Description: "Retrieve a specific GitHub pull request by number, including comments and linked issues",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "GitHub connector ID"
					},
					"pr_number": {
						"type": "integer",
						"description": "Pull request number"
					},
					"repository": {
						"type": "string",
						"description": "Repository name (optional, uses default if not specified)"
					}
				},
				"required": ["connector_id", "pr_number"]
			}`),
		},
		{
			Name:        ToolGitHubListRepos,
			Description: "List all accessible GitHub repositories for the connector",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "GitHub connector ID"
					}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolSlackSearch,
			Description: "Search for messages across Slack channels using a query string",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Slack connector ID to search in"
					},
					"query": {
						"type": "string",
						"description": "Search query to find messages"
					}
				},
				"required": ["connector_id", "query"]
			}`),
		},
		{
			Name:        ToolSlackListChannels,
			Description: "List all accessible Slack channels in the workspace",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Slack connector ID"
					}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolSlackGetThread,
			Description: "Retrieve all messages from a specific Slack thread",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Slack connector ID"
					},
					"channel_id": {
						"type": "string",
						"description": "Slack channel ID containing the thread"
					},
					"thread_ts": {
						"type": "string",
						"description": "Thread timestamp identifier"
					}
				},
				"required": ["connector_id", "channel_id", "thread_ts"]
			}`),
		},
		{
			Name:        ToolJiraSearch,
			Description: "Search for Jira issues using JQL (Jira Query Language)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Jira connector ID to search in"
					},
					"jql": {
						"type": "string",
						"description": "JQL query string (e.g., 'project = PROJ AND status = Open')"
					}
				},
				"required": ["connector_id", "jql"]
			}`),
		},
		{
			Name:        ToolJiraGetIssue,
			Description: "Retrieve a specific Jira issue by key, including comments",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Jira connector ID"
					},
					"issue_key": {
						"type": "string",
						"description": "Jira issue key (e.g., PROJ-123)"
					}
				},
				"required": ["connector_id", "issue_key"]
			}`),
		},
		{
			Name:        ToolJiraListProjects,
			Description: "List all accessible Jira projects",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Jira connector ID"
					}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolDiscordSearch,
			Description: "Search for messages in a Discord channel",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Discord connector ID"
					},
					"channel_id": {
						"type": "string",
						"description": "Discord channel ID to search in"
					},
					"query": {
						"type": "string",
						"description": "Search query to find messages"
					}
				},
				"required": ["connector_id", "channel_id", "query"]
			}`),
		},
		{
			Name:        ToolDiscordListChannels,
			Description: "List all accessible Discord channels in the guild/server",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Discord connector ID"
					}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolDiscordGetThread,
			Description: "Retrieve all messages from a specific Discord thread",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {
						"type": "string",
						"description": "Discord connector ID"
					},
					"thread_id": {
						"type": "string",
						"description": "Discord thread ID"
					}
				},
				"required": ["connector_id", "thread_id"]
			}`),
		},
	}
}
