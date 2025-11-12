package jira

import (
	jira "github.com/andygrunwald/go-jira"
)

// JiraClientInterface defines the interface for Jira API operations
// This allows for easier testing with mock clients
type JiraClientInterface interface {
	SearchIssues(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error)
	GetIssue(issueKey string, options *jira.GetQueryOptions) (*jira.Issue, *jira.Response, error)
	ListProjects() (*jira.ProjectList, *jira.Response, error)
}

// RealJiraClient wraps the real Jira client to implement JiraClientInterface
type RealJiraClient struct {
	client *jira.Client
}

// NewRealJiraClient creates a new real Jira client wrapper
func NewRealJiraClient(client *jira.Client) *RealJiraClient {
	return &RealJiraClient{client: client}
}

func (r *RealJiraClient) SearchIssues(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
	return r.client.Issue.Search(jql, options)
}

func (r *RealJiraClient) GetIssue(issueKey string, options *jira.GetQueryOptions) (*jira.Issue, *jira.Response, error) {
	return r.client.Issue.Get(issueKey, options)
}

func (r *RealJiraClient) ListProjects() (*jira.ProjectList, *jira.Response, error) {
	return r.client.Project.GetList()
}
