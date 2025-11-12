package github

import (
	"context"

	"github.com/google/go-github/v45/github"
)

// GitHubClientInterface defines the interface for GitHub client operations
type GitHubClientInterface interface {
	ListIssuesByRepo(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
	ListPullRequests(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
	GetRateLimits(ctx context.Context) (*github.RateLimits, *github.Response, error)
	SearchIssues(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error)
	GetIssue(ctx context.Context, owner string, repo string, number int) (*github.Issue, *github.Response, error)
	GetPullRequest(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error)
	ListIssueComments(ctx context.Context, owner string, repo string, number int, opts *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error)
	ListRepositories(ctx context.Context, user string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error)
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

func (r *RealGitHubClient) GetRateLimits(ctx context.Context) (*github.RateLimits, *github.Response, error) {
	return r.client.RateLimits(ctx)
}

func (r *RealGitHubClient) SearchIssues(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	return r.client.Search.Issues(ctx, query, opts)
}

func (r *RealGitHubClient) GetIssue(ctx context.Context, owner string, repo string, number int) (*github.Issue, *github.Response, error) {
	return r.client.Issues.Get(ctx, owner, repo, number)
}

func (r *RealGitHubClient) GetPullRequest(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error) {
	return r.client.PullRequests.Get(ctx, owner, repo, number)
}

func (r *RealGitHubClient) ListIssueComments(ctx context.Context, owner string, repo string, number int, opts *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error) {
	return r.client.Issues.ListComments(ctx, owner, repo, number, opts)
}

func (r *RealGitHubClient) ListRepositories(ctx context.Context, user string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error) {
	return r.client.Repositories.List(ctx, user, opts)
}

// MockGitHubClient implements GitHubClientInterface for testing
type MockGitHubClient struct {
	ListIssuesByRepoFunc  func(ctx context.Context, owner string, repo string, opts *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
	ListPullRequestsFunc  func(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
	GetRateLimitsFunc     func(ctx context.Context) (*github.RateLimits, *github.Response, error)
	SearchIssuesFunc      func(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error)
	GetIssueFunc          func(ctx context.Context, owner string, repo string, number int) (*github.Issue, *github.Response, error)
	GetPullRequestFunc    func(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error)
	ListIssueCommentsFunc func(ctx context.Context, owner string, repo string, number int, opts *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error)
	ListRepositoriesFunc  func(ctx context.Context, user string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error)
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

func (m *MockGitHubClient) GetRateLimits(ctx context.Context) (*github.RateLimits, *github.Response, error) {
	if m.GetRateLimitsFunc != nil {
		return m.GetRateLimitsFunc(ctx)
	}
	return &github.RateLimits{}, &github.Response{}, nil
}

func (m *MockGitHubClient) SearchIssues(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	if m.SearchIssuesFunc != nil {
		return m.SearchIssuesFunc(ctx, query, opts)
	}
	return nil, nil, nil
}

func (m *MockGitHubClient) GetIssue(ctx context.Context, owner string, repo string, number int) (*github.Issue, *github.Response, error) {
	if m.GetIssueFunc != nil {
		return m.GetIssueFunc(ctx, owner, repo, number)
	}
	return nil, nil, nil
}

func (m *MockGitHubClient) GetPullRequest(ctx context.Context, owner string, repo string, number int) (*github.PullRequest, *github.Response, error) {
	if m.GetPullRequestFunc != nil {
		return m.GetPullRequestFunc(ctx, owner, repo, number)
	}
	return nil, nil, nil
}

func (m *MockGitHubClient) ListIssueComments(ctx context.Context, owner string, repo string, number int, opts *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error) {
	if m.ListIssueCommentsFunc != nil {
		return m.ListIssueCommentsFunc(ctx, owner, repo, number, opts)
	}
	return nil, nil, nil
}

func (m *MockGitHubClient) ListRepositories(ctx context.Context, user string, opts *github.RepositoryListOptions) ([]*github.Repository, *github.Response, error) {
	if m.ListRepositoriesFunc != nil {
		return m.ListRepositoriesFunc(ctx, user, opts)
	}
	return nil, nil, nil
}
