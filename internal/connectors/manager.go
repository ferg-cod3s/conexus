package connectors

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors/discord"
	"github.com/ferg-cod3s/conexus/internal/connectors/github"
	"github.com/ferg-cod3s/conexus/internal/connectors/jira"
	"github.com/ferg-cod3s/conexus/internal/connectors/slack"
)

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// getStringArrayFromMap safely extracts a string array value from a map
func getStringArrayFromMap(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	// Try direct string array
	if val, ok := m[key].([]string); ok {
		return val
	}
	return []string{}
}

// getIntFromMap safely extracts an int value from a map
func getIntFromMap(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

// getDurationFromMap safely extracts a duration string and converts it
func getDurationFromMap(m map[string]interface{}, key string, defaultVal time.Duration) time.Duration {
	if val, ok := m[key].(string); ok {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

// ConnectorManager manages different types of connectors
type ConnectorManager struct {
	store      ConnectorStore
	connectors map[string]interface{} // connector ID -> connector instance
}

// NewConnectorManager creates a new connector manager
func NewConnectorManager(store ConnectorStore) *ConnectorManager {
	return &ConnectorManager{
		store:      store,
		connectors: make(map[string]interface{}),
	}
}

// GetConnector gets or creates a connector instance
func (cm *ConnectorManager) GetConnector(ctx context.Context, id string) (interface{}, error) {
	// Check if we already have the connector instance
	if conn, exists := cm.connectors[id]; exists {
		return conn, nil
	}

	// Get connector config from store
	connector, err := cm.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector %s: %w", id, err)
	}

	// Create connector instance based on type
	var instance interface{}
	switch connector.Type {
	case "github":
		config := &github.Config{
			Token:      getStringFromMap(connector.Config, "token"),
			Repository: getStringFromMap(connector.Config, "repository"),
		}

		instance, err = github.NewConnector(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub connector: %w", err)
		}

	case "slack":
		config := &slack.Config{
			Token:        getStringFromMap(connector.Config, "token"),
			Channels:     getStringArrayFromMap(connector.Config, "channels"),
			SyncInterval: getDurationFromMap(connector.Config, "sync_interval", 5*time.Minute),
			MaxMessages:  getIntFromMap(connector.Config, "max_messages", 1000),
		}

		instance, err = slack.NewConnector(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Slack connector: %w", err)
		}

	case "jira":
		config := &jira.Config{
			BaseURL:      getStringFromMap(connector.Config, "base_url"),
			Username:     getStringFromMap(connector.Config, "username"),
			APIToken:     getStringFromMap(connector.Config, "api_token"),
			Projects:     getStringArrayFromMap(connector.Config, "projects"),
			SyncInterval: getDurationFromMap(connector.Config, "sync_interval", 5*time.Minute),
			MaxIssues:    getIntFromMap(connector.Config, "max_issues", 1000),
		}

		instance, err = jira.NewConnector(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Jira connector: %w", err)
		}

	case "discord":
		config := &discord.Config{
			Token:        getStringFromMap(connector.Config, "token"),
			GuildID:      getStringFromMap(connector.Config, "guild_id"),
			Channels:     getStringArrayFromMap(connector.Config, "channels"),
			SyncInterval: getDurationFromMap(connector.Config, "sync_interval", 5*time.Minute),
			MaxMessages:  getIntFromMap(connector.Config, "max_messages", 1000),
		}

		instance, err = discord.NewConnector(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Discord connector: %w", err)
		}

	case "filesystem":
		// For filesystem, we might not need a special instance
		instance = nil

	default:
		return nil, fmt.Errorf("unsupported connector type: %s", connector.Type)
	}

	// Cache the instance
	cm.connectors[id] = instance
	return instance, nil
}

// SyncGitHubIssues syncs issues from a GitHub connector
func (cm *ConnectorManager) SyncGitHubIssues(ctx context.Context, connectorID string) ([]github.Issue, error) {
	os.Stderr.WriteString(fmt.Sprintf("DEBUG: ConnectorManager.SyncGitHubIssues called for %s\n", connectorID))
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("DEBUG: Failed to get connector: %v\n", err))
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		os.Stderr.WriteString(fmt.Sprintf("DEBUG: Connector %s is not a GitHub connector\n", connectorID))
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	os.Stderr.WriteString("DEBUG: Calling githubConn.SyncIssues\n")
	return githubConn.SyncIssues(ctx)
}

// SyncGitHubPullRequests syncs pull requests from a GitHub connector
func (cm *ConnectorManager) SyncGitHubPullRequests(ctx context.Context, connectorID string) ([]github.PullRequest, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.SyncPullRequests(ctx)
}

// SearchGitHubIssues searches for issues in a GitHub connector
func (cm *ConnectorManager) SearchGitHubIssues(ctx context.Context, connectorID, query, state string) ([]github.Issue, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.SearchIssues(ctx, query, state)
}

// GetGitHubIssue retrieves a specific issue from a GitHub connector
func (cm *ConnectorManager) GetGitHubIssue(ctx context.Context, connectorID string, issueNumber int) (*github.Issue, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.GetIssue(ctx, issueNumber)
}

// GetGitHubIssueComments retrieves comments for a GitHub issue
func (cm *ConnectorManager) GetGitHubIssueComments(ctx context.Context, connectorID string, issueNumber int) ([]github.Comment, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.GetIssueComments(ctx, issueNumber)
}

// GetGitHubPullRequest retrieves a specific pull request from a GitHub connector
func (cm *ConnectorManager) GetGitHubPullRequest(ctx context.Context, connectorID string, prNumber int) (*github.PullRequest, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.GetPullRequest(ctx, prNumber)
}

// GetGitHubPRComments retrieves comments for a GitHub pull request
func (cm *ConnectorManager) GetGitHubPRComments(ctx context.Context, connectorID string, prNumber int) ([]github.Comment, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.GetPRComments(ctx, prNumber)
}

// ListGitHubRepositories lists all repositories from a GitHub connector
func (cm *ConnectorManager) ListGitHubRepositories(ctx context.Context, connectorID string) ([]github.Repository, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.ListRepositories(ctx)
}

// SyncSlackMessages syncs messages from a Slack connector
func (cm *ConnectorManager) SyncSlackMessages(ctx context.Context, connectorID string) ([]slack.Message, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	slackConn, ok := conn.(*slack.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Slack connector", connectorID)
	}

	return slackConn.SyncMessages(ctx)
}

// SearchSlackMessages searches for messages in a Slack connector
func (cm *ConnectorManager) SearchSlackMessages(ctx context.Context, connectorID, query string) ([]slack.Message, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	slackConn, ok := conn.(*slack.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Slack connector", connectorID)
	}

	return slackConn.SearchMessages(ctx, query)
}

// ListSlackChannels lists all channels from a Slack connector
func (cm *ConnectorManager) ListSlackChannels(ctx context.Context, connectorID string) ([]slack.Channel, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	slackConn, ok := conn.(*slack.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Slack connector", connectorID)
	}

	return slackConn.ListChannels(ctx)
}

// GetSlackThread retrieves a thread from a Slack connector
func (cm *ConnectorManager) GetSlackThread(ctx context.Context, connectorID, channelID, threadTS string) (*slack.Thread, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	slackConn, ok := conn.(*slack.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Slack connector", connectorID)
	}

	return slackConn.GetThread(ctx, channelID, threadTS)
}

// SyncJiraIssues syncs issues from a Jira connector
func (cm *ConnectorManager) SyncJiraIssues(ctx context.Context, connectorID string) ([]jira.Issue, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	jiraConn, ok := conn.(*jira.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Jira connector", connectorID)
	}

	return jiraConn.SyncIssues(ctx)
}

// SearchJiraIssues searches for issues in a Jira connector using JQL
func (cm *ConnectorManager) SearchJiraIssues(ctx context.Context, connectorID, jql string) ([]jira.Issue, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	jiraConn, ok := conn.(*jira.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Jira connector", connectorID)
	}

	return jiraConn.SearchIssues(ctx, jql)
}

// GetJiraIssue retrieves a single issue from a Jira connector
func (cm *ConnectorManager) GetJiraIssue(ctx context.Context, connectorID, issueKey string) (*jira.Issue, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	jiraConn, ok := conn.(*jira.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Jira connector", connectorID)
	}

	return jiraConn.GetIssue(ctx, issueKey)
}

// GetJiraIssueComments retrieves comments for a Jira issue
func (cm *ConnectorManager) GetJiraIssueComments(ctx context.Context, connectorID, issueKey string) ([]jira.Comment, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	jiraConn, ok := conn.(*jira.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Jira connector", connectorID)
	}

	return jiraConn.GetIssueComments(ctx, issueKey)
}

// ListJiraProjects lists all projects from a Jira connector
func (cm *ConnectorManager) ListJiraProjects(ctx context.Context, connectorID string) ([]jira.Project, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	jiraConn, ok := conn.(*jira.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Jira connector", connectorID)
	}

	return jiraConn.ListProjects(ctx)
}

// SyncDiscordMessages syncs messages from a Discord connector
func (cm *ConnectorManager) SyncDiscordMessages(ctx context.Context, connectorID string) ([]discord.Message, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	discordConn, ok := conn.(*discord.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Discord connector", connectorID)
	}

	return discordConn.SyncMessages(ctx)
}

// SearchDiscordMessages searches for messages in a Discord connector
func (cm *ConnectorManager) SearchDiscordMessages(ctx context.Context, connectorID, channelID, query string) ([]discord.Message, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	discordConn, ok := conn.(*discord.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Discord connector", connectorID)
	}

	return discordConn.SearchMessages(ctx, channelID, query)
}

// ListDiscordChannels lists all channels from a Discord connector
func (cm *ConnectorManager) ListDiscordChannels(ctx context.Context, connectorID string) ([]discord.Channel, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	discordConn, ok := conn.(*discord.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Discord connector", connectorID)
	}

	return discordConn.ListChannels(ctx)
}

// GetDiscordThread retrieves thread messages from a Discord connector
func (cm *ConnectorManager) GetDiscordThread(ctx context.Context, connectorID, threadID string) ([]discord.Message, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	discordConn, ok := conn.(*discord.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Discord connector", connectorID)
	}

	return discordConn.GetThreadMessages(ctx, threadID)
}

// GetDiscordGuild retrieves guild information from a Discord connector
func (cm *ConnectorManager) GetDiscordGuild(ctx context.Context, connectorID string) (*discord.Guild, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	discordConn, ok := conn.(*discord.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a Discord connector", connectorID)
	}

	return discordConn.GetGuild(ctx)
}
