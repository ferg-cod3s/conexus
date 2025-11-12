package jira

import (
	"context"
	"testing"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
)

// MockJiraClient implements JiraClientInterface for testing
type MockJiraClient struct {
	SearchIssuesFunc func(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error)
	GetIssueFunc     func(issueKey string, options *jira.GetQueryOptions) (*jira.Issue, *jira.Response, error)
	ListProjectsFunc func() (*jira.ProjectList, *jira.Response, error)
}

func (m *MockJiraClient) SearchIssues(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
	if m.SearchIssuesFunc != nil {
		return m.SearchIssuesFunc(jql, options)
	}
	return []jira.Issue{}, nil, nil
}

func (m *MockJiraClient) GetIssue(issueKey string, options *jira.GetQueryOptions) (*jira.Issue, *jira.Response, error) {
	if m.GetIssueFunc != nil {
		return m.GetIssueFunc(issueKey, options)
	}
	return &jira.Issue{}, nil, nil
}

func (m *MockJiraClient) ListProjects() (*jira.ProjectList, *jira.Response, error) {
	if m.ListProjectsFunc != nil {
		return m.ListProjectsFunc()
	}
	projectList := jira.ProjectList{}
	return &projectList, nil, nil
}

func TestNewConnector(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				BaseURL:  "https://example.atlassian.net",
				Username: "user@example.com",
				APIToken: "test-token",
				Projects: []string{"PROJ"},
			},
			expectErr: false,
		},
		{
			name: "Missing base URL",
			config: &Config{
				Username: "user@example.com",
				APIToken: "test-token",
				Projects: []string{"PROJ"},
			},
			expectErr: true,
		},
		{
			name: "Missing username",
			config: &Config{
				BaseURL:  "https://example.atlassian.net",
				APIToken: "test-token",
				Projects: []string{"PROJ"},
			},
			expectErr: true,
		},
		{
			name: "Missing API token",
			config: &Config{
				BaseURL:  "https://example.atlassian.net",
				Username: "user@example.com",
				Projects: []string{"PROJ"},
			},
			expectErr: true,
		},
		{
			name: "Missing projects",
			config: &Config{
				BaseURL:  "https://example.atlassian.net",
				Username: "user@example.com",
				APIToken: "test-token",
			},
			expectErr: true,
		},
		{
			name: "With custom settings",
			config: &Config{
				BaseURL:      "https://example.atlassian.net",
				Username:     "user@example.com",
				APIToken:     "test-token",
				Projects:     []string{"PROJ1", "PROJ2"},
				SyncInterval: 10 * time.Minute,
				MaxIssues:    500,
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			connector, err := NewConnector(tc.config)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, connector)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, connector)
				assert.NotNil(t, connector.client)
				assert.NotNil(t, connector.config)
			}
		})
	}
}

func TestSyncIssues(t *testing.T) {
	mockClient := &MockJiraClient{
		SearchIssuesFunc: func(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
			return []jira.Issue{
				{
					ID:  "10001",
					Key: "PROJ-1",
					Fields: &jira.IssueFields{
						Summary:     "Test issue 1",
						Description: "Description 1",
						Status: &jira.Status{
							Name: "In Progress",
						},
						Priority: &jira.Priority{
							Name: "High",
						},
						Type: jira.IssueType{
							Name: "Bug",
						},
						Project: jira.Project{
							Key: "PROJ",
						},
						Created: jira.Time(time.Now()),
						Updated: jira.Time(time.Now()),
					},
				},
				{
					ID:  "10002",
					Key: "PROJ-2",
					Fields: &jira.IssueFields{
						Summary:     "Test issue 2",
						Description: "Description 2",
						Status: &jira.Status{
							Name: "Done",
						},
						Priority: &jira.Priority{
							Name: "Medium",
						},
						Type: jira.IssueType{
							Name: "Story",
						},
						Project: jira.Project{
							Key: "PROJ",
						},
						Created: jira.Time(time.Now()),
						Updated: jira.Time(time.Now()),
					},
				},
			}, nil, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			BaseURL:   "https://example.atlassian.net",
			Username:  "user@example.com",
			APIToken:  "test-token",
			Projects:  []string{"PROJ"},
			MaxIssues: 1000,
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	issues, err := connector.SyncIssues(ctx)

	assert.NoError(t, err)
	assert.Len(t, issues, 2)
	assert.Equal(t, "PROJ-1", issues[0].Key)
	assert.Equal(t, "Test issue 1", issues[0].Summary)
	assert.Equal(t, "Bug", issues[0].IssueType)
	assert.Equal(t, "High", issues[0].Priority)
}

func TestSearchIssues(t *testing.T) {
	mockClient := &MockJiraClient{
		SearchIssuesFunc: func(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
			return []jira.Issue{
				{
					ID:  "10001",
					Key: "PROJ-1",
					Fields: &jira.IssueFields{
						Summary:     "Found issue",
						Description: "Matching description",
						Status: &jira.Status{
							Name: "Open",
						},
						Project: jira.Project{
							Key: "PROJ",
						},
						Created: jira.Time(time.Now()),
						Updated: jira.Time(time.Now()),
					},
				},
			}, nil, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			BaseURL:  "https://example.atlassian.net",
			Username: "user@example.com",
			APIToken: "test-token",
			Projects: []string{"PROJ"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	issues, err := connector.SearchIssues(ctx, "text ~ \"test\"")

	assert.NoError(t, err)
	assert.Len(t, issues, 1)
	assert.Equal(t, "PROJ-1", issues[0].Key)
	assert.Equal(t, "Found issue", issues[0].Summary)
}

func TestGetIssue(t *testing.T) {
	mockClient := &MockJiraClient{
		GetIssueFunc: func(issueKey string, options *jira.GetQueryOptions) (*jira.Issue, *jira.Response, error) {
			return &jira.Issue{
				ID:  "10001",
				Key: issueKey,
				Fields: &jira.IssueFields{
					Summary:     "Single issue",
					Description: "Issue description",
					Status: &jira.Status{
						Name: "In Review",
					},
					Priority: &jira.Priority{
						Name: "Critical",
					},
					Type: jira.IssueType{
						Name: "Bug",
					},
					Assignee: &jira.User{
						DisplayName: "John Doe",
					},
					Reporter: &jira.User{
						DisplayName: "Jane Smith",
					},
					Project: jira.Project{
						Key: "PROJ",
					},
					Created: jira.Time(time.Now()),
					Updated: jira.Time(time.Now()),
				},
			}, nil, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			BaseURL:  "https://example.atlassian.net",
			Username: "user@example.com",
			APIToken: "test-token",
			Projects: []string{"PROJ"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	issue, err := connector.GetIssue(ctx, "PROJ-1")

	assert.NoError(t, err)
	assert.NotNil(t, issue)
	assert.Equal(t, "PROJ-1", issue.Key)
	assert.Equal(t, "Single issue", issue.Summary)
	assert.Equal(t, "Critical", issue.Priority)
	assert.Equal(t, "John Doe", issue.Assignee)
	assert.Equal(t, "Jane Smith", issue.Reporter)
}

func TestGetIssueComments(t *testing.T) {
	mockClient := &MockJiraClient{
		GetIssueFunc: func(issueKey string, options *jira.GetQueryOptions) (*jira.Issue, *jira.Response, error) {
			return &jira.Issue{
				ID:  "10001",
				Key: issueKey,
				Fields: &jira.IssueFields{
					Summary: "Issue with comments",
					Comments: &jira.Comments{
						Comments: []*jira.Comment{
							{
								ID:   "100",
								Body: "First comment",
								Author: jira.User{
									DisplayName: "John Doe",
								},
								Created: time.Now().Format(time.RFC3339),
								Updated: time.Now().Format(time.RFC3339),
							},
							{
								ID:   "101",
								Body: "Second comment",
								Author: jira.User{
									DisplayName: "Jane Smith",
								},
								Created: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
								Updated: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
							},
						},
					},
					Created: jira.Time(time.Now()),
					Updated: jira.Time(time.Now()),
				},
			}, nil, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			BaseURL:  "https://example.atlassian.net",
			Username: "user@example.com",
			APIToken: "test-token",
			Projects: []string{"PROJ"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	comments, err := connector.GetIssueComments(ctx, "PROJ-1")

	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, "100", comments[0].ID)
	assert.Equal(t, "First comment", comments[0].Body)
	assert.Equal(t, "John Doe", comments[0].Author)
	assert.Equal(t, "101", comments[1].ID)
	assert.Equal(t, "Second comment", comments[1].Body)
	assert.Equal(t, "Jane Smith", comments[1].Author)
}

func TestListProjects(t *testing.T) {
	mockClient := &MockJiraClient{
		ListProjectsFunc: func() (*jira.ProjectList, *jira.Response, error) {
			projectList := jira.ProjectList{
				{
					ID:             "10000",
					Key:            "PROJ1",
					Name:           "Project 1",
					ProjectTypeKey: "software",
				},
				{
					ID:             "10001",
					Key:            "PROJ2",
					Name:           "Project 2",
					ProjectTypeKey: "business",
				},
			}
			return &projectList, nil, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			BaseURL:  "https://example.atlassian.net",
			Username: "user@example.com",
			APIToken: "test-token",
			Projects: []string{"PROJ1"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	projects, err := connector.ListProjects(ctx)

	assert.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Equal(t, "PROJ1", projects[0].Key)
	assert.Equal(t, "Project 1", projects[0].Name)
	assert.Equal(t, "software", projects[0].Type)
	assert.Equal(t, "PROJ2", projects[1].Key)
	assert.Equal(t, "business", projects[1].Type)
}

func TestGetSyncStatus(t *testing.T) {
	connector := &Connector{
		config: &Config{
			BaseURL:  "https://example.atlassian.net",
			Username: "user@example.com",
			APIToken: "test-token",
			Projects: []string{"PROJ"},
		},
		rateLimit: &RateLimitInfo{
			Remaining: 100,
			Reset:     time.Now().Add(1 * time.Hour),
		},
		status: &SyncStatus{
			LastSync:       time.Now(),
			TotalIssues:    150,
			TotalProjects:  2,
			SyncInProgress: false,
		},
	}

	status := connector.GetSyncStatus()

	assert.NotNil(t, status)
	assert.Equal(t, 150, status.TotalIssues)
	assert.Equal(t, 2, status.TotalProjects)
	assert.False(t, status.SyncInProgress)
	assert.NotNil(t, status.RateLimit)
	assert.Equal(t, 100, status.RateLimit.Remaining)
}
