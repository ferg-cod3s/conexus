// Package mcp implements the Model Context Protocol server for Conexus.
package mcp

import "encoding/json"

// Tool names exposed by the MCP server
const (
	ToolContextSearch            = "context.search"
	ToolContextGetRelatedInfo    = "context.get_related_info"
	ToolContextIndexControl      = "context.index_control"
	ToolContextConnectorManagement = "context.connector_management"
)

// Resource URI scheme
const (
	ResourceScheme = "engine"
	ResourceFiles  = "files"
)

// SearchRequest represents the input for context.search tool
type SearchRequest struct {
	Query       string                 `json:"query"`
	WorkContext *WorkContext           `json:"work_context,omitempty"`
	TopK        int                    `json:"top_k,omitempty"`
	Filters     *SearchFilters         `json:"filters,omitempty"`
}

// WorkContext provides information about the user's current working context
type WorkContext struct {
	ActiveFile    string   `json:"active_file,omitempty"`
	GitBranch     string   `json:"git_branch,omitempty"`
	OpenTicketIDs []string `json:"open_ticket_ids,omitempty"`
}

// SearchFilters defines filtering options for search
type SearchFilters struct {
	SourceTypes []string   `json:"source_types,omitempty"`
	DateRange   *DateRange `json:"date_range,omitempty"`
}

// DateRange specifies a time range filter
type DateRange struct {
	From string `json:"from,omitempty"` // ISO 8601 date-time
	To   string `json:"to,omitempty"`   // ISO 8601 date-time
}

// SearchResponse represents the output of context.search tool
type SearchResponse struct {
	Results      []SearchResultItem `json:"results"`
	TotalCount   int                `json:"total_count"`
	QueryTime    float64            `json:"query_time_ms"`
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

// GetRelatedInfoResponse represents the output of context.get_related_info tool
type GetRelatedInfoResponse struct {
	Summary      string              `json:"summary"`
	RelatedPRs   []string            `json:"related_prs,omitempty"`
	RelatedIssues []string           `json:"related_issues,omitempty"`
	Discussions  []DiscussionSummary `json:"discussions,omitempty"`
}

// DiscussionSummary provides a summary of a Slack discussion
type DiscussionSummary struct {
	Channel   string `json:"channel"`
	Timestamp string `json:"timestamp"`
	Summary   string `json:"summary"`
}

// IndexControlRequest represents the input for context.index_control tool
type IndexControlRequest struct {
	Action     string   `json:"action"` // "start", "stop", "status", "force_reindex"
	Connectors []string `json:"connectors,omitempty"`
}

// IndexControlResponse represents the output of context.index_control tool
type IndexControlResponse struct {
	Status      string            `json:"status"`
	Message     string            `json:"message"`
	Details     map[string]string `json:"details,omitempty"`
}

// ConnectorManagementRequest represents the input for context.connector_management tool
type ConnectorManagementRequest struct {
	Action          string                 `json:"action"` // "list", "add", "update", "remove"
	ConnectorID     string                 `json:"connector_id,omitempty"`
	ConnectorConfig map[string]interface{} `json:"connector_config,omitempty"`
}

// ConnectorManagementResponse represents the output of context.connector_management tool
type ConnectorManagementResponse struct {
	Connectors []ConnectorInfo `json:"connectors,omitempty"`
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
						"enum": ["start", "stop", "status", "force_reindex"]
					},
					"connectors": {
						"type": "array",
						"items": {"type": "string"}
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
	}
}
