package github

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v45/github"
)

func TestParseRepository(t *testing.T) {
	tests := []struct {
		input         string
		expectedOwner string
		expectedName  string
	}{
		{"owner/repo", "owner", "repo"},
		{"owner/sub/repo", "owner", "sub"}, // parseRepository only splits on first /
		{"repo", "", "repo"},
		{"", "", ""},
		{"owner/", "owner", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			owner, name := parseRepository(test.input)
			if owner != test.expectedOwner {
				t.Errorf("Expected owner %s, got %s", test.expectedOwner, owner)
			}
			if name != test.expectedName {
				t.Errorf("Expected name %s, got %s", test.expectedName, name)
			}
		})
	}
}

func TestExtractIssueReferences(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "no references",
			text:     "This is just a regular PR description",
			expected: []string{},
		},
		{
			name:     "hash references",
			text:     "Fixes #123 and #456",
			expected: []string{"123", "456", "123"}, // Fixes #123 matches both patterns, #456 matches only #(\d+)
		},
		{
			name:     "PROJ references",
			text:     "Related to PROJ-789 and PROJ-101",
			expected: []string{"789", "101"},
		},
		{
			name:     "JIRA references",
			text:     "JIRA-202 and JIRA-303",
			expected: []string{"202", "303"},
		},
		{
			name:     "Fixes and Closes",
			text:     "Fixes #123 and Closes #456",
			expected: []string{"123", "456", "123", "456"}, // Both patterns will match
		},
		{
			name:     "mixed references",
			text:     "Fixes #123, relates to PROJ-456, and JIRA-789",
			expected: []string{"123", "456", "789", "123"}, // #123 matches both patterns
		},
		{
			name:     "duplicate references",
			text:     "Fixes #123 and #123 again",
			expected: []string{"123", "123", "123"}, // Fixes #123 matches both patterns, #123 matches only #(\d+)
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := extractIssueReferences(test.text)
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d references, got %d: %v", len(test.expected), len(result), result)
			}
			for i, expected := range test.expected {
				if i < len(result) && result[i] != expected {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			}
		})
	}
}

func TestNewConnector(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Token:      "test-token",
				Repository: "owner/repo",
			},
			wantErr: false,
		},
		{
			name: "missing token",
			config: &Config{
				Repository: "owner/repo",
			},
			wantErr: true,
		},
		{
			name: "missing repository",
			config: &Config{
				Token: "test-token",
			},
			wantErr: true,
		},
		{
			name: "valid config with sync interval",
			config: &Config{
				Token:        "test-token",
				Repository:   "owner/repo",
				SyncInterval: 10 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "config with zero sync interval should get default",
			config: &Config{
				Token:        "test-token",
				Repository:   "owner/repo",
				SyncInterval: 0,
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connector, err := NewConnector(test.config)
			if (err != nil) != test.wantErr {
				t.Errorf("NewConnector() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if !test.wantErr {
				if connector == nil {
					t.Error("Expected connector to not be nil")
				}
				if connector.config.SyncInterval == 0 {
					t.Error("Expected default sync interval to be set")
				}
			}
		})
	}
}

func TestConfigStruct(t *testing.T) {
	config := &Config{
		Token:         "test-token",
		Repository:    "owner/repo",
		WebhookSecret: "secret",
		SyncInterval:  5 * time.Minute,
	}

	if config.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", config.Token)
	}
	if config.Repository != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got '%s'", config.Repository)
	}
	if config.WebhookSecret != "secret" {
		t.Errorf("Expected webhook secret 'secret', got '%s'", config.WebhookSecret)
	}
	if config.SyncInterval != 5*time.Minute {
		t.Errorf("Expected sync interval 5m, got %v", config.SyncInterval)
	}
}

func TestIssueStruct(t *testing.T) {
	now := time.Now()
	issue := Issue{
		ID:          123,
		Number:      456,
		Title:       "Test Issue",
		Description: "Test description",
		State:       "open",
		Labels:      []string{"bug", "enhancement"},
		Assignee:    "testuser",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if issue.ID != 123 {
		t.Errorf("Expected ID 123, got %d", issue.ID)
	}
	if issue.Number != 456 {
		t.Errorf("Expected Number 456, got %d", issue.Number)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Expected Title 'Test Issue', got '%s'", issue.Title)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(issue.Labels))
	}
}

func TestPullRequestStruct(t *testing.T) {
	pr := &PullRequest{
		ID:          123,
		Number:      42,
		Title:       "Test PR",
		Description: "Test description",
		State:       "open",
		Assignee:    "testuser",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if pr.ID != 123 {
		t.Errorf("Expected ID 123, got %d", pr.ID)
	}
	if pr.Number != 42 {
		t.Errorf("Expected Number 42, got %d", pr.Number)
	}
	if pr.Title != "Test PR" {
		t.Errorf("Expected Title 'Test PR', got '%s'", pr.Title)
	}
}

func TestConnectorMethods(t *testing.T) {
	// Test basic connector methods that don't require external calls
	connector := &Connector{
		client: nil,
		config: &Config{
			Token:         "test-token",
			Repository:    "owner/repo",
			WebhookSecret: "secret",
			SyncInterval:  5 * time.Minute,
		},
	}

	// Test that config is accessible
	if connector.config.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", connector.config.Token)
	}
	if connector.config.Repository != "owner/repo" {
		t.Errorf("Expected repository 'owner/repo', got '%s'", connector.config.Repository)
	}
	if connector.config.SyncInterval != 5*time.Minute {
		t.Errorf("Expected sync interval 5m, got %v", connector.config.SyncInterval)
	}
}

func TestSyncIssues(t *testing.T) {
	// Create a mock client
	mockClient := &MockGitHubClient{}

	// Setup mock data
	now := time.Now()
	issueID := int64(123)
	issueNumber := 42
	title := "Test Issue"
	body := "Test description"
	state := "open"
	labelName := "bug"
	assigneeLogin := "testuser"

	mockClient.ListIssuesByRepoFunc = func(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error) {
		// Return mock issue
		issue := &github.Issue{
			ID:     &issueID,
			Number: &issueNumber,
			Title:  &title,
			Body:   &body,
			State:  &state,
			Labels: []*github.Label{
				{Name: &labelName},
			},
			Assignee: &github.User{
				Login: &assigneeLogin,
			},
			CreatedAt: &now,
			UpdatedAt: &now,
		}

		return []*github.Issue{issue}, &github.Response{NextPage: 0}, nil
	}

	// Create connector with mock client
	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:        "test-token",
			Repository:   "owner/repo",
			SyncInterval: 5 * time.Minute,
		},
	}

	// Test SyncIssues
	ctx := context.Background()
	issues, err := connector.SyncIssues(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}

	if len(issues) > 0 {
		issue := issues[0]
		if issue.ID != 123 {
			t.Errorf("Expected ID 123, got %d", issue.ID)
		}
		if issue.Number != 42 {
			t.Errorf("Expected Number 42, got %d", issue.Number)
		}
		if issue.Title != "Test Issue" {
			t.Errorf("Expected Title 'Test Issue', got '%s'", issue.Title)
		}
		if issue.State != "open" {
			t.Errorf("Expected State 'open', got '%s'", issue.State)
		}
		if len(issue.Labels) != 1 || issue.Labels[0] != "bug" {
			t.Errorf("Expected Labels ['bug'], got %v", issue.Labels)
		}
		if issue.Assignee != "testuser" {
			t.Errorf("Expected Assignee 'testuser', got '%s'", issue.Assignee)
		}
	}
}

func TestSyncPullRequests(t *testing.T) {
	// Create a mock client
	mockClient := &MockGitHubClient{}

	// Setup mock data
	now := time.Now()
	prID := int64(456)
	prNumber := 789
	title := "Test PR"
	body := "Fixes #123 and relates to PROJ-456"
	state := "open"
	labelName := "enhancement"
	assigneeLogin := "testuser"

	mockClient.ListPullRequestsFunc = func(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
		// Return mock PR
		pr := &github.PullRequest{
			ID:     &prID,
			Number: &prNumber,
			Title:  &title,
			Body:   &body,
			State:  &state,
			Labels: []*github.Label{
				{Name: &labelName},
			},
			Assignee: &github.User{
				Login: &assigneeLogin,
			},
			CreatedAt: &now,
			UpdatedAt: &now,
		}

		return []*github.PullRequest{pr}, &github.Response{NextPage: 0}, nil
	}

	// Create connector with mock client
	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:        "test-token",
			Repository:   "owner/repo",
			SyncInterval: 5 * time.Minute,
		},
	}

	// Test SyncPullRequests
	ctx := context.Background()
	prs, err := connector.SyncPullRequests(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(prs))
	}

	if len(prs) > 0 {
		pr := prs[0]
		if pr.ID != 456 {
			t.Errorf("Expected ID 456, got %d", pr.ID)
		}
		if pr.Number != 789 {
			t.Errorf("Expected Number 789, got %d", pr.Number)
		}
		if pr.Title != "Test PR" {
			t.Errorf("Expected Title 'Test PR', got '%s'", pr.Title)
		}
		if pr.State != "open" {
			t.Errorf("Expected State 'open', got '%s'", pr.State)
		}
		if len(pr.Labels) != 1 || pr.Labels[0] != "enhancement" {
			t.Errorf("Expected Labels ['enhancement'], got %v", pr.Labels)
		}
		if pr.Assignee != "testuser" {
			t.Errorf("Expected Assignee 'testuser', got '%s'", pr.Assignee)
		}
		// Check that linked issues were extracted
		if len(pr.LinkedIssues) != 3 {
			t.Errorf("Expected 3 linked issues, got %d", len(pr.LinkedIssues))
		}
	}
}

func TestSyncPullRequestsWithNoBody(t *testing.T) {
	// Create a mock client
	mockClient := &MockGitHubClient{}

	// Setup mock data with no body
	now := time.Now()
	prID := int64(456)
	prNumber := 789
	title := "Test PR"
	state := "open"

	mockClient.ListPullRequestsFunc = func(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
		// Return mock PR with no body
		pr := &github.PullRequest{
			ID:        &prID,
			Number:    &prNumber,
			Title:     &title,
			State:     &state,
			CreatedAt: &now,
			UpdatedAt: &now,
		}

		return []*github.PullRequest{pr}, &github.Response{NextPage: 0}, nil
	}

	// Create connector with mock client
	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:        "test-token",
			Repository:   "owner/repo",
			SyncInterval: 5 * time.Minute,
		},
	}

	// Test SyncPullRequests
	ctx := context.Background()
	prs, err := connector.SyncPullRequests(ctx)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(prs) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(prs))
	}

	if len(prs) > 0 {
		pr := prs[0]
		// Check that no linked issues were extracted when body is nil
		if len(pr.LinkedIssues) != 0 {
			t.Errorf("Expected 0 linked issues, got %d", len(pr.LinkedIssues))
		}
	}
}
