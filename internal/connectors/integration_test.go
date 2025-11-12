//go:build integration
// +build integration

package connectors_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/connectors/discord"
	"github.com/ferg-cod3s/conexus/internal/connectors/github"
	"github.com/ferg-cod3s/conexus/internal/connectors/jira"
	"github.com/ferg-cod3s/conexus/internal/connectors/slack"
)

// TestConfig represents the test configuration file
type TestConfig struct {
	GitHub  *GitHubTestConfig  `json:"github,omitempty"`
	Slack   *SlackTestConfig   `json:"slack,omitempty"`
	Jira    *JiraTestConfig    `json:"jira,omitempty"`
	Discord *DiscordTestConfig `json:"discord,omitempty"`
}

type GitHubTestConfig struct {
	Token      string `json:"token"`
	Repository string `json:"repository"`
}

type SlackTestConfig struct {
	Token    string   `json:"token"`
	Channels []string `json:"channels"`
}

type JiraTestConfig struct {
	BaseURL  string   `json:"base_url"`
	Email    string   `json:"email"`
	APIToken string   `json:"api_token"`
	Projects []string `json:"projects"`
}

type DiscordTestConfig struct {
	Token    string   `json:"token"`
	GuildID  string   `json:"guild_id"`
	Channels []string `json:"channels"`
}

// loadTestConfig loads test configuration from file or environment variables
func loadTestConfig(t *testing.T) *TestConfig {
	config := &TestConfig{}

	// Try to load from config file first
	configPath := os.Getenv("CONEXUS_TEST_CONFIG")
	if configPath == "" {
		configPath = "../../config/test_connectors.json"
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := json.Unmarshal(data, config); err != nil {
			t.Logf("Warning: failed to parse config file: %v", err)
		}
	}

	// Override with environment variables if set
	if token := os.Getenv("CONEXUS_TEST_GITHUB_TOKEN"); token != "" {
		if config.GitHub == nil {
			config.GitHub = &GitHubTestConfig{}
		}
		config.GitHub.Token = token
	}
	if repo := os.Getenv("CONEXUS_TEST_GITHUB_REPO"); repo != "" {
		if config.GitHub == nil {
			config.GitHub = &GitHubTestConfig{}
		}
		config.GitHub.Repository = repo
	}

	if token := os.Getenv("CONEXUS_TEST_SLACK_TOKEN"); token != "" {
		if config.Slack == nil {
			config.Slack = &SlackTestConfig{}
		}
		config.Slack.Token = token
	}

	if token := os.Getenv("CONEXUS_TEST_JIRA_TOKEN"); token != "" {
		if config.Jira == nil {
			config.Jira = &JiraTestConfig{}
		}
		config.Jira.APIToken = token
	}
	if email := os.Getenv("CONEXUS_TEST_JIRA_EMAIL"); email != "" {
		if config.Jira == nil {
			config.Jira = &JiraTestConfig{}
		}
		config.Jira.Email = email
	}
	if url := os.Getenv("CONEXUS_TEST_JIRA_URL"); url != "" {
		if config.Jira == nil {
			config.Jira = &JiraTestConfig{}
		}
		config.Jira.BaseURL = url
	}

	if token := os.Getenv("CONEXUS_TEST_DISCORD_TOKEN"); token != "" {
		if config.Discord == nil {
			config.Discord = &DiscordTestConfig{}
		}
		config.Discord.Token = token
	}
	if guildID := os.Getenv("CONEXUS_TEST_DISCORD_GUILD"); guildID != "" {
		if config.Discord == nil {
			config.Discord = &DiscordTestConfig{}
		}
		config.Discord.GuildID = guildID
	}

	return config
}

func TestGitHubConnector_Integration(t *testing.T) {
	config := loadTestConfig(t)
	if config.GitHub == nil || config.GitHub.Token == "" {
		t.Skip("GitHub test credentials not configured")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ghConfig := &github.Config{
		Token:      config.GitHub.Token,
		Repository: config.GitHub.Repository,
	}

	connector, err := github.NewConnector(ghConfig)
	if err != nil {
		t.Fatalf("Failed to create GitHub connector: %v", err)
	}

	// Verify connector implements BaseConnector interface
	if err := connectors.ValidateConnector(connector); err != nil {
		t.Errorf("Connector validation failed: %v", err)
	}

	// Test GetType
	connType := connector.GetType()
	if connType != "github" {
		t.Errorf("GetType() = %q, want %q", connType, "github")
	}

	// Test capabilities
	caps := connectors.GetCapabilities(connector)
	t.Logf("GitHub capabilities: %+v", caps)
	if !caps.SupportsSearch {
		t.Error("GitHub should support search")
	}

	// Test GetRateLimit
	rateLimit := connector.GetRateLimit()
	if rateLimit == nil {
		t.Error("GetRateLimit() returned nil")
	} else {
		t.Logf("GitHub rate limit: %+v", rateLimit)
	}

	// Test GetSyncStatus
	syncStatus := connector.GetSyncStatus()
	if syncStatus == nil {
		t.Error("GetSyncStatus() returned nil")
	} else {
		t.Logf("GitHub sync status: %+v", syncStatus)
	}

	// Test SearchIssues
	issues, err := connector.SearchIssues(ctx, "is:open", "open")
	if err != nil {
		t.Errorf("SearchIssues() error = %v", err)
	} else {
		t.Logf("Found %d open issues", len(issues))
	}

	// Test ListRepositories
	repos, err := connector.ListRepositories(ctx)
	if err != nil {
		t.Errorf("ListRepositories() error = %v", err)
	} else {
		t.Logf("Found %d repositories", len(repos))
	}

	// If we have issues, test GetIssue
	if len(issues) > 0 {
		issue, err := connector.GetIssue(ctx, issues[0].Number)
		if err != nil {
			t.Errorf("GetIssue(%d) error = %v", issues[0].Number, err)
		} else {
			t.Logf("Retrieved issue #%d: %s", issue.Number, issue.Title)
		}

		// Test GetIssueComments
		comments, err := connector.GetIssueComments(ctx, issues[0].Number)
		if err != nil {
			t.Errorf("GetIssueComments(%d) error = %v", issues[0].Number, err)
		} else {
			t.Logf("Issue #%d has %d comments", issues[0].Number, len(comments))
		}
	}
}

func TestSlackConnector_Integration(t *testing.T) {
	config := loadTestConfig(t)
	if config.Slack == nil || config.Slack.Token == "" {
		t.Skip("Slack test credentials not configured")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	slackConfig := &slack.Config{
		Token:    config.Slack.Token,
		Channels: config.Slack.Channels,
	}

	connector, err := slack.NewConnector(slackConfig)
	if err != nil {
		t.Fatalf("Failed to create Slack connector: %v", err)
	}

	// Verify connector implements BaseConnector interface
	if err := connectors.ValidateConnector(connector); err != nil {
		t.Errorf("Connector validation failed: %v", err)
	}

	// Test GetType
	connType := connector.GetType()
	if connType != "slack" {
		t.Errorf("GetType() = %q, want %q", connType, "slack")
	}

	// Test capabilities
	caps := connectors.GetCapabilities(connector)
	t.Logf("Slack capabilities: %+v", caps)

	// Test ListChannels
	channels, err := connector.ListChannels(ctx)
	if err != nil {
		t.Errorf("ListChannels() error = %v", err)
	} else {
		t.Logf("Found %d channels", len(channels))
		for _, ch := range channels[:min(5, len(channels))] {
			t.Logf("  - %s (%s)", ch.Name, ch.ID)
		}
	}

	// Test SearchMessages (if search is supported)
	messages, err := connector.SearchMessages(ctx, "test")
	if err != nil {
		t.Logf("SearchMessages() error = %v (may not have search scope)", err)
	} else {
		t.Logf("Found %d messages matching 'test'", len(messages))
	}

	// Test GetThread (if we have messages with threads)
	if len(channels) > 0 && len(config.Slack.Channels) > 0 {
		// This would require knowing a thread timestamp
		t.Log("GetThread test skipped (requires known thread_ts)")
	}
}

func TestJiraConnector_Integration(t *testing.T) {
	config := loadTestConfig(t)
	if config.Jira == nil || config.Jira.APIToken == "" {
		t.Skip("Jira test credentials not configured")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	jiraConfig := &jira.Config{
		BaseURL:  config.Jira.BaseURL,
		Email:    config.Jira.Email,
		APIToken: config.Jira.APIToken,
		Projects: config.Jira.Projects,
	}

	connector, err := jira.NewConnector(jiraConfig)
	if err != nil {
		t.Fatalf("Failed to create Jira connector: %v", err)
	}

	// Verify connector implements BaseConnector interface
	if err := connectors.ValidateConnector(connector); err != nil {
		t.Errorf("Connector validation failed: %v", err)
	}

	// Test GetType
	connType := connector.GetType()
	if connType != "jira" {
		t.Errorf("GetType() = %q, want %q", connType, "jira")
	}

	// Test capabilities
	caps := connectors.GetCapabilities(connector)
	t.Logf("Jira capabilities: %+v", caps)

	// Test ListProjects
	projects, err := connector.ListProjects(ctx)
	if err != nil {
		t.Errorf("ListProjects() error = %v", err)
	} else {
		t.Logf("Found %d projects", len(projects))
		for _, proj := range projects[:min(5, len(projects))] {
			t.Logf("  - %s (%s)", proj.Name, proj.Key)
		}
	}

	// Test SearchIssues
	if len(config.Jira.Projects) > 0 {
		jql := "project = " + config.Jira.Projects[0] + " ORDER BY created DESC"
		issues, err := connector.SearchIssues(ctx, jql)
		if err != nil {
			t.Errorf("SearchIssues() error = %v", err)
		} else {
			t.Logf("Found %d issues in project %s", len(issues), config.Jira.Projects[0])

			// Test GetIssue if we have issues
			if len(issues) > 0 {
				issue, err := connector.GetIssue(ctx, issues[0].Key)
				if err != nil {
					t.Errorf("GetIssue(%s) error = %v", issues[0].Key, err)
				} else {
					t.Logf("Retrieved issue %s: %s", issue.Key, issue.Summary)
				}

				// Test GetIssueComments
				comments, err := connector.GetIssueComments(ctx, issues[0].Key)
				if err != nil {
					t.Errorf("GetIssueComments(%s) error = %v", issues[0].Key, err)
				} else {
					t.Logf("Issue %s has %d comments", issues[0].Key, len(comments))
				}
			}
		}
	}
}

func TestDiscordConnector_Integration(t *testing.T) {
	config := loadTestConfig(t)
	if config.Discord == nil || config.Discord.Token == "" {
		t.Skip("Discord test credentials not configured")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	discordConfig := &discord.Config{
		Token:    config.Discord.Token,
		GuildID:  config.Discord.GuildID,
		Channels: config.Discord.Channels,
	}

	connector, err := discord.NewConnector(discordConfig)
	if err != nil {
		t.Fatalf("Failed to create Discord connector: %v", err)
	}

	// Verify connector implements BaseConnector interface
	if err := connectors.ValidateConnector(connector); err != nil {
		t.Errorf("Connector validation failed: %v", err)
	}

	// Test GetType
	connType := connector.GetType()
	if connType != "discord" {
		t.Errorf("GetType() = %q, want %q", connType, "discord")
	}

	// Test capabilities
	caps := connectors.GetCapabilities(connector)
	t.Logf("Discord capabilities: %+v", caps)

	// Test ListChannels
	channels, err := connector.ListChannels(ctx)
	if err != nil {
		t.Errorf("ListChannels() error = %v", err)
	} else {
		t.Logf("Found %d channels", len(channels))
		for _, ch := range channels[:min(5, len(channels))] {
			t.Logf("  - %s (%s)", ch.Name, ch.ID)
		}
	}

	// Test SearchMessages (requires channel ID)
	if len(config.Discord.Channels) > 0 {
		messages, err := connector.SearchMessages(ctx, config.Discord.Channels[0], "test")
		if err != nil {
			t.Logf("SearchMessages() error = %v", err)
		} else {
			t.Logf("Found %d messages matching 'test' in channel", len(messages))
		}
	}

	// Test GetGuild
	guild, err := connector.GetGuild(ctx)
	if err != nil {
		t.Errorf("GetGuild() error = %v", err)
	} else {
		t.Logf("Guild: %s (ID: %s, Members: %d)", guild.Name, guild.ID, guild.MemberCount)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
