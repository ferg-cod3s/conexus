package github

import (
	"context"

	"github.com/google/go-github/v45/github"
)

// GitHubClientInterface defines the interface for GitHub client operations
type GitHubClientInterface interface {
	ListIssuesByRepo(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
	ListPullRequests(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
}

// RealGitHubClient wraps the actual GitHub client
type RealGitHubClient struct {
	client *github.Client
}

func NewRealGitHubClient(client *github.Client) *RealGitHubClient {
	return &RealGitHubClient{client: client}
}

func (r *RealGitHubClient) ListIssuesByRepo(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error) {
	return r.client.Issues.ListByRepo(ctx, owner, repo, opts)
}

func (r *RealGitHubClient) ListPullRequests(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	return r.client.PullRequests.List(ctx, owner, repo, opts)
}

// MockGitHubClient implements GitHubClientInterface for testing
type MockGitHubClient struct {
	ListIssuesByRepoFunc func(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
	ListPullRequestsFunc func(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
}

func (m *MockGitHubClient) ListIssuesByRepo(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error) {
	if m.ListIssuesByRepoFunc != nil {
		return m.ListIssuesByRepoFunc(ctx, owner, repo, opts)
	}
	return nil, nil, nil
}

func (m *MockGitHubClient) ListPullRequests(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	if m.ListPullRequestsFunc != nil {
		return m.ListPullRequestsFunc(ctx, owner, repo, opts)
	}
	return nil, nil, nil
}
