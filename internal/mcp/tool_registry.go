package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ferg-cod3s/conexus/internal/connectors"
)

// ToolRegistry manages dynamic tool registration based on available connectors
type ToolRegistry struct {
	connectorStore connectors.ConnectorStore
	baseTools      []ToolDefinition
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(connectorStore connectors.ConnectorStore) *ToolRegistry {
	return &ToolRegistry{
		connectorStore: connectorStore,
		baseTools:      getBaseToolDefinitions(),
	}
}

// GetAvailableTools returns tools based on configured connectors
func (tr *ToolRegistry) GetAvailableTools(ctx context.Context) ([]ToolDefinition, error) {
	// Start with base tools (always available)
	tools := make([]ToolDefinition, len(tr.baseTools))
	copy(tools, tr.baseTools)

	// Get list of configured connectors
	connectorList, err := tr.connectorStore.List(ctx)
	if err != nil {
		return tools, fmt.Errorf("failed to list connectors: %w", err)
	}

	// Track which connector types are available
	connectorTypes := make(map[string]bool)
	for _, conn := range connectorList {
		connectorTypes[conn.Type] = true
	}

	// Add GitHub tools if any GitHub connectors exist
	if connectorTypes["github"] {
		tools = append(tools, getGitHubToolDefinitions()...)
	}

	// Add Slack tools if any Slack connectors exist
	if connectorTypes["slack"] {
		tools = append(tools, getSlackToolDefinitions()...)
	}

	// Add Jira tools if any Jira connectors exist
	if connectorTypes["jira"] {
		tools = append(tools, getJiraToolDefinitions()...)
	}

	// Add Discord tools if any Discord connectors exist
	if connectorTypes["discord"] {
		tools = append(tools, getDiscordToolDefinitions()...)
	}

	return tools, nil
}

// GetAvailableConnectors returns list of available connectors by type
func (tr *ToolRegistry) GetAvailableConnectors(ctx context.Context) (map[string][]string, error) {
	connectorList, err := tr.connectorStore.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}

	result := make(map[string][]string)
	for _, conn := range connectorList {
		result[conn.Type] = append(result[conn.Type], conn.ID)
	}

	return result, nil
}

// getBaseToolDefinitions returns core tools that are always available
func getBaseToolDefinitions() []ToolDefinition {
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
								"items": {"type": "string", "enum": ["file", "slack", "github", "jira", "discord"]}
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
									"current_story_id": {"type": "string"},
									"boost_active": {"type": "boolean"}
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
			Description: "Retrieves contextually related information based on the current file or ticket.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"file_path": {"type": "string", "description": "Path to the file to get related info for"},
					"ticket_id": {"type": "string", "description": "Ticket ID to get related info for"},
					"query": {"type": "string", "description": "Optional additional context query"}
				}
			}`),
		},
		{
			Name:        ToolContextIndexControl,
			Description: "Controls indexing operations (start, stop, status).",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"action": {"type": "string", "enum": ["start", "stop", "status", "reindex"]},
					"source_type": {"type": "string", "enum": ["file", "slack", "github", "jira", "discord"]}
				},
				"required": ["action"]
			}`),
		},
		{
			Name:        ToolContextConnectorManagement,
			Description: "Manages connectors (list, add, remove, configure).",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"action": {"type": "string", "enum": ["list", "add", "remove", "configure", "test"]},
					"connector_id": {"type": "string"},
					"connector_type": {"type": "string", "enum": ["github", "slack", "jira", "discord"]},
					"config": {"type": "object"}
				},
				"required": ["action"]
			}`),
		},
		{
			Name:        ToolContextExplain,
			Description: "Explains code, patterns, or architecture using context from the codebase.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": {"type": "string", "description": "What to explain"},
					"file_path": {"type": "string", "description": "Optional file path for context"},
					"line_start": {"type": "integer", "description": "Starting line number"},
					"line_end": {"type": "integer", "description": "Ending line number"}
				},
				"required": ["query"]
			}`),
		},
		{
			Name:        ToolContextGrep,
			Description: "Performs code search with regex patterns across the codebase.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"pattern": {"type": "string", "description": "Regex pattern to search for"},
					"file_pattern": {"type": "string", "description": "File glob pattern to filter files"},
					"case_sensitive": {"type": "boolean", "default": false},
					"max_results": {"type": "integer", "default": 100}
				},
				"required": ["pattern"]
			}`),
		},
	}
}

// getGitHubToolDefinitions returns GitHub-specific tool definitions
func getGitHubToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolGitHubSyncStatus,
			Description: "Gets the sync status of a GitHub connector.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "GitHub connector ID"}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolGitHubSyncTrigger,
			Description: "Triggers a sync operation for a GitHub connector.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "GitHub connector ID"},
					"sync_type": {"type": "string", "enum": ["issues", "prs", "discussions", "all"], "default": "all"}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolGitHubSearchIssues,
			Description: "Searches for issues in a GitHub repository using GitHub search syntax.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "GitHub connector ID"},
					"query": {"type": "string", "description": "GitHub search query (e.g., 'is:open label:bug')"},
					"state": {"type": "string", "enum": ["open", "closed", "all"], "default": "open"}
				},
				"required": ["connector_id", "query"]
			}`),
		},
		{
			Name:        ToolGitHubGetIssue,
			Description: "Retrieves a specific GitHub issue with its comments.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "GitHub connector ID"},
					"issue_number": {"type": "integer", "description": "Issue number"}
				},
				"required": ["connector_id", "issue_number"]
			}`),
		},
		{
			Name:        ToolGitHubGetPR,
			Description: "Retrieves a specific GitHub pull request with its comments.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "GitHub connector ID"},
					"pr_number": {"type": "integer", "description": "Pull request number"}
				},
				"required": ["connector_id", "pr_number"]
			}`),
		},
		{
			Name:        ToolGitHubListRepos,
			Description: "Lists all repositories accessible to the GitHub connector.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "GitHub connector ID"}
				},
				"required": ["connector_id"]
			}`),
		},
	}
}

// getSlackToolDefinitions returns Slack-specific tool definitions
func getSlackToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolSlackSearch,
			Description: "Searches for messages in Slack channels.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Slack connector ID"},
					"query": {"type": "string", "description": "Search query"}
				},
				"required": ["connector_id", "query"]
			}`),
		},
		{
			Name:        ToolSlackListChannels,
			Description: "Lists all channels in the Slack workspace.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Slack connector ID"}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolSlackGetThread,
			Description: "Retrieves a specific Slack thread with all replies.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Slack connector ID"},
					"channel_id": {"type": "string", "description": "Channel ID"},
					"thread_ts": {"type": "string", "description": "Thread timestamp"}
				},
				"required": ["connector_id", "channel_id", "thread_ts"]
			}`),
		},
	}
}

// getJiraToolDefinitions returns Jira-specific tool definitions
func getJiraToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolJiraSearch,
			Description: "Searches for Jira issues using JQL (Jira Query Language).",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Jira connector ID"},
					"jql": {"type": "string", "description": "JQL query (e.g., 'project = PROJ AND status = Open')"}
				},
				"required": ["connector_id", "jql"]
			}`),
		},
		{
			Name:        ToolJiraGetIssue,
			Description: "Retrieves a specific Jira issue with its comments.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Jira connector ID"},
					"issue_key": {"type": "string", "description": "Issue key (e.g., 'PROJ-123')"}
				},
				"required": ["connector_id", "issue_key"]
			}`),
		},
		{
			Name:        ToolJiraListProjects,
			Description: "Lists all projects in the Jira instance.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Jira connector ID"}
				},
				"required": ["connector_id"]
			}`),
		},
	}
}

// getDiscordToolDefinitions returns Discord-specific tool definitions
func getDiscordToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolDiscordSearch,
			Description: "Searches for messages in a Discord channel.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Discord connector ID"},
					"channel_id": {"type": "string", "description": "Channel ID"},
					"query": {"type": "string", "description": "Search query"}
				},
				"required": ["connector_id", "channel_id", "query"]
			}`),
		},
		{
			Name:        ToolDiscordListChannels,
			Description: "Lists all channels in the Discord server.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Discord connector ID"}
				},
				"required": ["connector_id"]
			}`),
		},
		{
			Name:        ToolDiscordGetThread,
			Description: "Retrieves messages from a specific Discord thread.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"connector_id": {"type": "string", "description": "Discord connector ID"},
					"thread_id": {"type": "string", "description": "Thread ID"}
				},
				"required": ["connector_id", "thread_id"]
			}`),
		},
	}
}
