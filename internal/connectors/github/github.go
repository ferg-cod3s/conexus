package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type Connector struct {
	client        GitHubClientInterface
	config        *Config
	rateLimit     *RateLimitInfo
	rateLimitMu   sync.RWMutex
	status        *SyncStatus
	statusMu      sync.RWMutex
	webhookSecret []byte
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
	ID           int64      `json:"id"`
	Number       int        `json:"number"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	State        string     `json:"state"`
	Labels       []string   `json:"labels"`
	Assignee     string     `json:"assignee"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LinkedIssues []string   `json:"linked_issues"`
	Merged       bool       `json:"merged"`
	MergedAt     *time.Time `json:"merged_at"`
	HeadBranch   string     `json:"head_branch"`
	BaseBranch   string     `json:"base_branch"`
	ReviewCount  int        `json:"review_count"`
	Additions    int        `json:"additions"`
	Deletions    int        `json:"deletions"`
	ChangedFiles int        `json:"changed_files"`
}

type Discussion struct {
	ID          int64     `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AnswerCount int       `json:"answer_count"`
	Locked      bool      `json:"locked"`
}

type WebhookEvent struct {
	Type      string      `json:"type"`
	Action    string      `json:"action"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

type SyncStatus struct {
	LastSync         time.Time      `json:"last_sync"`
	TotalIssues      int            `json:"total_issues"`
	TotalPRs         int            `json:"total_prs"`
	TotalDiscussions int            `json:"total_discussions"`
	SyncInProgress   bool           `json:"sync_in_progress"`
	Error            string         `json:"error,omitempty"`
	RateLimit        *RateLimitInfo `json:"rate_limit,omitempty"`
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

	connector := &Connector{
		client:        client,
		config:        config,
		rateLimit:     &RateLimitInfo{},
		status:        &SyncStatus{},
		webhookSecret: []byte(config.WebhookSecret),
	}

	// Initialize rate limit info
	if err := connector.updateRateLimit(context.Background()); err != nil {
		log.Printf("Warning: Failed to get initial rate limit: %v", err)
	}

	return connector, nil
}

func (gc *Connector) SyncIssues(ctx context.Context) ([]Issue, error) {
	// Update sync status
	gc.statusMu.Lock()
	gc.status.SyncInProgress = true
	gc.statusMu.Unlock()

	// Check rate limit before starting
	if err := gc.WaitForRateLimit(ctx); err != nil {
		gc.updateSyncStatus(0, 0, 0, err)
		return nil, err
	}

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
			gc.updateSyncStatus(0, 0, 0, err)
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

		// Update rate limit info from response headers
		if resp != nil {
			gc.rateLimitMu.Lock()
			gc.rateLimit.Limit = resp.Rate.Limit
			gc.rateLimit.Remaining = resp.Rate.Remaining
			gc.rateLimit.Reset = resp.Rate.Reset.Time
			gc.rateLimitMu.Unlock()
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

func (gc *Connector) SyncPullRequests(ctx context.Context) ([]PullRequest, error) {
	// Check rate limit before starting
	if err := gc.WaitForRateLimit(ctx); err != nil {
		return nil, err
	}

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

			mergedAt := pr.MergedAt
			headBranch := ""
			baseBranch := ""
			if pr.Head != nil {
				headBranch = pr.Head.GetRef()
			}
			if pr.Base != nil {
				baseBranch = pr.Base.GetRef()
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
				Merged:       pr.GetMerged(),
				MergedAt:     mergedAt,
				HeadBranch:   headBranch,
				BaseBranch:   baseBranch,
				ReviewCount:  pr.GetReviewComments(),
				Additions:    pr.GetAdditions(),
				Deletions:    pr.GetDeletions(),
				ChangedFiles: pr.GetChangedFiles(),
			})
		}

		// Update rate limit info from response headers
		if resp != nil {
			gc.rateLimitMu.Lock()
			gc.rateLimit.Limit = resp.Rate.Limit
			gc.rateLimit.Remaining = resp.Rate.Remaining
			gc.rateLimit.Reset = resp.Rate.Reset.Time
			gc.rateLimitMu.Unlock()
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

// SyncDiscussions syncs GitHub discussions
func (gc *Connector) SyncDiscussions(ctx context.Context) ([]Discussion, error) {
	owner, repo := parseRepository(gc.config.Repository)

	// GitHub Discussions API requires GraphQL, but for now we'll use REST API for categories
	// Note: Full discussions sync would require GraphQL API integration
	opts := &github.ListOptions{PerPage: 100}

	var allDiscussions []Discussion

	// Since GitHub Discussions API is limited in REST, we'll return empty for now
	// In a real implementation, you'd use GraphQL:
	// query($owner: String!, $repo: String!) {
	//   repository(owner: $owner, name: $repo) {
	//     discussions(first: 100) {
	//       nodes { ... }
	//     }
	//   }
	// }

	_ = opts // Suppress unused warning
	_ = owner
	_ = repo

	return allDiscussions, nil
}

// GetType returns the connector type
func (gc *Connector) GetType() string {
	return "github"
}

// GetRateLimit returns current rate limit information
func (gc *Connector) GetRateLimit() *RateLimitInfo {
	gc.rateLimitMu.RLock()
	defer gc.rateLimitMu.RUnlock()

	if gc.rateLimit != nil {
		return &RateLimitInfo{
			Limit:     gc.rateLimit.Limit,
			Remaining: gc.rateLimit.Remaining,
			Reset:     gc.rateLimit.Reset,
		}
	}
	return &RateLimitInfo{}
}

// GetSyncStatus returns current sync status
func (gc *Connector) GetSyncStatus() *SyncStatus {
	gc.statusMu.RLock()
	defer gc.statusMu.RUnlock()

	return &SyncStatus{
		LastSync:         gc.status.LastSync,
		TotalIssues:      gc.status.TotalIssues,
		TotalPRs:         gc.status.TotalPRs,
		TotalDiscussions: gc.status.TotalDiscussions,
		SyncInProgress:   gc.status.SyncInProgress,
		Error:            gc.status.Error,
		RateLimit:        gc.GetRateLimit(),
	}
}

// VerifyWebhookSignature verifies GitHub webhook signature
func (gc *Connector) VerifyWebhookSignature(payload []byte, signature string) bool {
	if len(gc.webhookSecret) == 0 {
		return true // No secret configured, skip verification
	}

	expectedSignature := "sha256=" + gc.generateHMAC(payload)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// generateHMAC generates HMAC-SHA256 signature
func (gc *Connector) generateHMAC(payload []byte) string {
	h := hmac.New(sha256.New, gc.webhookSecret)
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// ParseWebhookEvent parses incoming webhook payload
func (gc *Connector) ParseWebhookEvent(payload []byte, eventType string) (*WebhookEvent, error) {
	var event interface{}

	switch eventType {
	case "issues":
		var issueEvent github.IssuesEvent
		if err := json.Unmarshal(payload, &issueEvent); err != nil {
			return nil, fmt.Errorf("failed to parse issues event: %w", err)
		}
		event = issueEvent

	case "pull_request":
		var prEvent github.PullRequestEvent
		if err := json.Unmarshal(payload, &prEvent); err != nil {
			return nil, fmt.Errorf("failed to parse pull request event: %w", err)
		}
		event = prEvent

	case "discussion":
		var discussionEvent github.DiscussionEvent
		if err := json.Unmarshal(payload, &discussionEvent); err != nil {
			return nil, fmt.Errorf("failed to parse discussion event: %w", err)
		}
		event = discussionEvent

	default:
		// For unknown event types, store as raw JSON
		var rawEvent map[string]interface{}
		if err := json.Unmarshal(payload, &rawEvent); err != nil {
			return nil, fmt.Errorf("failed to parse event: %w", err)
		}
		event = rawEvent
	}

	return &WebhookEvent{
		Type:      eventType,
		Action:    gc.extractAction(event),
		Payload:   event,
		Timestamp: time.Now(),
	}, nil
}

// extractAction extracts the action from webhook event
func (gc *Connector) extractAction(event interface{}) string {
	if eventMap, ok := event.(map[string]interface{}); ok {
		if action, exists := eventMap["action"]; exists {
			if actionStr, ok := action.(string); ok {
				return actionStr
			}
		}
	}
	return "unknown"
}

// updateRateLimit updates rate limit information from GitHub API
func (gc *Connector) updateRateLimit(ctx context.Context) error {
	rateLimits, _, err := gc.client.GetRateLimits(ctx)
	if err != nil {
		return fmt.Errorf("failed to get rate limits: %w", err)
	}

	gc.rateLimitMu.Lock()
	defer gc.rateLimitMu.Unlock()

	if rateLimits.Core != nil {
		gc.rateLimit = &RateLimitInfo{
			Limit:     rateLimits.Core.Limit,
			Remaining: rateLimits.Core.Remaining,
			Reset:     rateLimits.Core.Reset.Time,
		}
	} else {
		// Fallback to defaults
		gc.rateLimit = &RateLimitInfo{
			Limit:     5000, // Default for authenticated requests
			Remaining: 5000,
			Reset:     time.Now().Add(time.Hour),
		}
	}

	return nil
}

// updateSyncStatus updates the sync status
func (gc *Connector) updateSyncStatus(totalIssues, totalPRs, totalDiscussions int, err error) {
	gc.statusMu.Lock()
	defer gc.statusMu.Unlock()

	gc.status.LastSync = time.Now()
	gc.status.TotalIssues = totalIssues
	gc.status.TotalPRs = totalPRs
	gc.status.TotalDiscussions = totalDiscussions
	gc.status.SyncInProgress = false

	if err != nil {
		gc.status.Error = err.Error()
	} else {
		gc.status.Error = ""
	}
}

// WaitForRateLimit waits if rate limit is exceeded
func (gc *Connector) WaitForRateLimit(ctx context.Context) error {
	gc.rateLimitMu.RLock()
	rateLimit := gc.rateLimit
	gc.rateLimitMu.RUnlock()

	if rateLimit.Remaining > 10 { // Keep some buffer
		return nil
	}

	now := time.Now()
	if rateLimit.Reset.After(now) {
		waitTime := rateLimit.Reset.Sub(now)
		log.Printf("Rate limit exceeded, waiting %v for reset", waitTime)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			return nil
		}
	}

	return nil
}

// SearchIssues searches for issues using GitHub search syntax
func (gc *Connector) SearchIssues(ctx context.Context, query string, state string) ([]Issue, error) {
	owner, repo := parseRepository(gc.config.Repository)

	// Build search query with repository filter
	searchQuery := fmt.Sprintf("repo:%s/%s %s", owner, repo, query)
	if state != "" && state != "all" {
		searchQuery = fmt.Sprintf("%s state:%s", searchQuery, state)
	}

	opts := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	result, _, err := gc.client.SearchIssues(ctx, searchQuery, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	var issues []Issue
	for _, ghIssue := range result.Issues {
		// Skip pull requests (GitHub API returns PRs in issue search)
		if ghIssue.PullRequestLinks != nil {
			continue
		}

		createdAt := time.Time{}
		if ghIssue.CreatedAt != nil {
			createdAt = *ghIssue.CreatedAt
		}

		updatedAt := time.Time{}
		if ghIssue.UpdatedAt != nil {
			updatedAt = *ghIssue.UpdatedAt
		}

		assignee := ""
		if ghIssue.Assignee != nil {
			assignee = ghIssue.Assignee.GetLogin()
		}

		var labels []string
		for _, label := range ghIssue.Labels {
			labels = append(labels, label.GetName())
		}

		issue := Issue{
			ID:          ghIssue.GetID(),
			Number:      ghIssue.GetNumber(),
			Title:       ghIssue.GetTitle(),
			Description: ghIssue.GetBody(),
			State:       ghIssue.GetState(),
			Labels:      labels,
			Assignee:    assignee,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// GetIssue retrieves a specific issue by number
func (gc *Connector) GetIssue(ctx context.Context, issueNumber int) (*Issue, error) {
	owner, repo := parseRepository(gc.config.Repository)

	ghIssue, _, err := gc.client.GetIssue(ctx, owner, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	createdAt := time.Time{}
	if ghIssue.CreatedAt != nil {
		createdAt = *ghIssue.CreatedAt
	}

	updatedAt := time.Time{}
	if ghIssue.UpdatedAt != nil {
		updatedAt = *ghIssue.UpdatedAt
	}

	assignee := ""
	if ghIssue.Assignee != nil {
		assignee = ghIssue.Assignee.GetLogin()
	}

	var labels []string
	for _, label := range ghIssue.Labels {
		labels = append(labels, label.GetName())
	}

	issue := &Issue{
		ID:          ghIssue.GetID(),
		Number:      ghIssue.GetNumber(),
		Title:       ghIssue.GetTitle(),
		Description: ghIssue.GetBody(),
		State:       ghIssue.GetState(),
		Labels:      labels,
		Assignee:    assignee,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return issue, nil
}

// Comment represents a GitHub comment
type Comment struct {
	ID        int64     `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetIssueComments retrieves all comments for an issue
func (gc *Connector) GetIssueComments(ctx context.Context, issueNumber int) ([]Comment, error) {
	owner, repo := parseRepository(gc.config.Repository)

	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allComments []Comment
	for {
		comments, resp, err := gc.client.ListIssueComments(ctx, owner, repo, issueNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issue comments: %w", err)
		}

		for _, c := range comments {
			createdAt := time.Time{}
			if c.CreatedAt != nil {
				createdAt = *c.CreatedAt
			}

			updatedAt := time.Time{}
			if c.UpdatedAt != nil {
				updatedAt = *c.UpdatedAt
			}

			author := ""
			if c.User != nil {
				author = c.User.GetLogin()
			}

			comment := Comment{
				ID:        c.GetID(),
				Author:    author,
				Body:      c.GetBody(),
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}
			allComments = append(allComments, comment)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// GetPullRequest retrieves a specific pull request by number
func (gc *Connector) GetPullRequest(ctx context.Context, prNumber int) (*PullRequest, error) {
	owner, repo := parseRepository(gc.config.Repository)

	ghPR, _, err := gc.client.GetPullRequest(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}

	createdAt := time.Time{}
	if ghPR.CreatedAt != nil {
		createdAt = *ghPR.CreatedAt
	}

	updatedAt := time.Time{}
	if ghPR.UpdatedAt != nil {
		updatedAt = *ghPR.UpdatedAt
	}

	assignee := ""
	if ghPR.Assignee != nil {
		assignee = ghPR.Assignee.GetLogin()
	}

	headBranch := ""
	if ghPR.Head != nil && ghPR.Head.Ref != nil {
		headBranch = *ghPR.Head.Ref
	}

	baseBranch := ""
	if ghPR.Base != nil && ghPR.Base.Ref != nil {
		baseBranch = *ghPR.Base.Ref
	}

	var labels []string
	for _, label := range ghPR.Labels {
		labels = append(labels, label.GetName())
	}

	var linkedIssues []string
	if ghPR.Body != nil {
		linkedIssues = extractIssueReferences(*ghPR.Body)
	}

	pr := &PullRequest{
		ID:           ghPR.GetID(),
		Number:       ghPR.GetNumber(),
		Title:        ghPR.GetTitle(),
		Description:  ghPR.GetBody(),
		State:        ghPR.GetState(),
		Merged:       ghPR.GetMerged(),
		Labels:       labels,
		Assignee:     assignee,
		HeadBranch:   headBranch,
		BaseBranch:   baseBranch,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		LinkedIssues: linkedIssues,
	}

	if ghPR.MergedAt != nil {
		pr.MergedAt = ghPR.MergedAt
	}

	return pr, nil
}

// GetPRComments retrieves all comments for a pull request
func (gc *Connector) GetPRComments(ctx context.Context, prNumber int) ([]Comment, error) {
	owner, repo := parseRepository(gc.config.Repository)

	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allComments []Comment
	for {
		comments, resp, err := gc.client.ListIssueComments(ctx, owner, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list PR comments: %w", err)
		}

		for _, c := range comments {
			createdAt := time.Time{}
			if c.CreatedAt != nil {
				createdAt = *c.CreatedAt
			}

			updatedAt := time.Time{}
			if c.UpdatedAt != nil {
				updatedAt = *c.UpdatedAt
			}

			author := ""
			if c.User != nil {
				author = c.User.GetLogin()
			}

			comment := Comment{
				ID:        c.GetID(),
				Author:    author,
				Body:      c.GetBody(),
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}
			allComments = append(allComments, comment)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// Repository represents a GitHub repository
type Repository struct {
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	Private       bool      `json:"private"`
	DefaultBranch string    `json:"default_branch"`
	Language      string    `json:"language"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	OpenIssues    int       `json:"open_issues"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	URL           string    `json:"url"`
}

// ListRepositories lists all accessible repositories for the authenticated user
func (gc *Connector) ListRepositories(ctx context.Context) ([]Repository, error) {
	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allRepos []Repository
	for {
		repos, resp, err := gc.client.ListRepositories(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, r := range repos {
			createdAt := time.Time{}
			if r.CreatedAt != nil {
				createdAt = r.CreatedAt.Time
			}

			updatedAt := time.Time{}
			if r.UpdatedAt != nil {
				updatedAt = r.UpdatedAt.Time
			}

			stars := 0
			if r.StargazersCount != nil {
				stars = *r.StargazersCount
			}

			forks := 0
			if r.ForksCount != nil {
				forks = *r.ForksCount
			}

			openIssues := 0
			if r.OpenIssuesCount != nil {
				openIssues = *r.OpenIssuesCount
			}

			repo := Repository{
				Name:          r.GetName(),
				FullName:      r.GetFullName(),
				Description:   r.GetDescription(),
				Private:       r.GetPrivate(),
				DefaultBranch: r.GetDefaultBranch(),
				Language:      r.GetLanguage(),
				URL:           r.GetHTMLURL(),
				Stars:         stars,
				Forks:         forks,
				OpenIssues:    openIssues,
				CreatedAt:     createdAt,
				UpdatedAt:     updatedAt,
			}

			allRepos = append(allRepos, repo)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}
