package github

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type Connector struct {
	client GitHubClientInterface
	config *Config
}

type Config struct {
	Token         string        `json:"token"`
	Repository    string        `json:"repository"`
	WebhookSecret string        `json:"webhook_secret"`
	SyncInterval  time.Duration `json:"sync_interval"`
}

type Issue struct {
	ID          int64     `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Labels      []string  `json:"labels"`
	Assignee    string    `json:"assignee"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PullRequest struct {
	ID           int64     `json:"id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	State        string    `json:"state"`
	Labels       []string  `json:"labels"`
	Assignee     string    `json:"assignee"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LinkedIssues []string  `json:"linked_issues"`
}

func NewConnector(config *Config) (*Connector, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}

	if config.Repository == "" {
		return nil, fmt.Errorf("GitHub repository is required")
	}

	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute // Default sync interval
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	githubClient := github.NewClient(tc)
	client := NewRealGitHubClient(githubClient)

	return &Connector{
		client: client,
		config: config,
	}, nil
}

func (gc *Connector) SyncIssues(ctx context.Context) ([]Issue, error) {
	owner, repo := parseRepository(gc.config.Repository)

	opts := &github.IssueListByRepoOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allIssues []Issue

	for {
		issues, resp, err := gc.client.ListIssuesByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}

		for _, issue := range issues {
			if issue.PullRequestLinks == nil { // Skip PRs
				var labels []string
				for _, label := range issue.Labels {
					labels = append(labels, label.GetName())
				}

				assignee := ""
				if issue.Assignee != nil {
					assignee = issue.Assignee.GetLogin()
				}

				createdAt := time.Time{}
				if issue.CreatedAt != nil {
					createdAt = *issue.CreatedAt
				}

				updatedAt := time.Time{}
				if issue.UpdatedAt != nil {
					updatedAt = *issue.UpdatedAt
				}

				allIssues = append(allIssues, Issue{
					ID:          issue.GetID(),
					Number:      issue.GetNumber(),
					Title:       issue.GetTitle(),
					Description: issue.GetBody(),
					State:       issue.GetState(),
					Labels:      labels,
					Assignee:    assignee,
					CreatedAt:   createdAt,
					UpdatedAt:   updatedAt,
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

func (gc *Connector) SyncPullRequests(ctx context.Context) ([]PullRequest, error) {
	owner, repo := parseRepository(gc.config.Repository)

	opts := &github.PullRequestListOptions{
		State:       "all",
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allPRs []PullRequest

	for {
		prs, resp, err := gc.client.ListPullRequests(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch pull requests: %w", err)
		}

		for _, pr := range prs {
			var labels []string
			for _, label := range pr.Labels {
				labels = append(labels, label.GetName())
			}

			assignee := ""
			if pr.Assignee != nil {
				assignee = pr.Assignee.GetLogin()
			}

			// Extract linked issues from PR body
			var linkedIssues []string
			if pr.Body != nil {
				linkedIssues = extractIssueReferences(*pr.Body)
			}

			createdAt := time.Time{}
			if pr.CreatedAt != nil {
				createdAt = *pr.CreatedAt
			}

			updatedAt := time.Time{}
			if pr.UpdatedAt != nil {
				updatedAt = *pr.UpdatedAt
			}

			allPRs = append(allPRs, PullRequest{
				ID:           pr.GetID(),
				Number:       pr.GetNumber(),
				Title:        pr.GetTitle(),
				Description:  pr.GetBody(),
				State:        pr.GetState(),
				Labels:       labels,
				Assignee:     assignee,
				CreatedAt:    createdAt,
				UpdatedAt:    updatedAt,
				LinkedIssues: linkedIssues,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

func parseRepository(repo string) (owner, name string) {
	parts := strings.Split(repo, "/")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return "", repo
}

func extractIssueReferences(text string) []string {
	// Extract issue references like #123, PROJ-456, etc.
	var issues []string
	patterns := []string{
		`#(\d+)`,        // #123
		`PROJ-(\d+)`,    // PROJ-123
		`JIRA-(\d+)`,    // JIRA-123
		`Fixes #(\d+)`,  // Fixes #123
		`Closes #(\d+)`, // Closes #123
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				issues = append(issues, match[1])
			}
		}
	}

	return issues
}
