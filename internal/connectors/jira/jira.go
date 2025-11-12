package jira

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

type Connector struct {
	client      JiraClientInterface
	config      *Config
	rateLimit   *RateLimitInfo
	rateLimitMu sync.RWMutex
	status      *SyncStatus
	statusMu    sync.RWMutex
}

type Config struct {
	BaseURL      string        `json:"base_url"`       // Jira instance URL
	Username     string        `json:"username"`       // Email for Jira Cloud, username for Jira Server
	APIToken     string        `json:"api_token"`      // API token or password
	Projects     []string      `json:"projects"`       // Project keys to index
	SyncInterval time.Duration `json:"sync_interval"`  // How often to sync
	MaxIssues    int           `json:"max_issues"`     // Max issues per project
}

type Issue struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	IssueType   string    `json:"issue_type"`
	Assignee    string    `json:"assignee"`
	Reporter    string    `json:"reporter"`
	Labels      []string  `json:"labels"`
	Components  []string  `json:"components"`
	FixVersions []string  `json:"fix_versions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	Project     string    `json:"project"`
}

type Comment struct {
	ID        string    `json:"id"`
	IssueKey  string    `json:"issue_key"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Project struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Lead        string `json:"lead"`
	Type        string `json:"type"`
}

type RateLimitInfo struct {
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

type SyncStatus struct {
	LastSync       time.Time      `json:"last_sync"`
	TotalIssues    int            `json:"total_issues"`
	TotalProjects  int            `json:"total_projects"`
	SyncInProgress bool           `json:"sync_in_progress"`
	Error          string         `json:"error,omitempty"`
	RateLimit      *RateLimitInfo `json:"rate_limit,omitempty"`
}

func NewConnector(config *Config) (*Connector, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("Jira base URL is required")
	}

	if config.Username == "" {
		return nil, fmt.Errorf("Jira username is required")
	}

	if config.APIToken == "" {
		return nil, fmt.Errorf("Jira API token is required")
	}

	if len(config.Projects) == 0 {
		return nil, fmt.Errorf("at least one Jira project is required")
	}

	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute // Default sync interval
	}

	if config.MaxIssues == 0 {
		config.MaxIssues = 1000 // Default max issues per project
	}

	tp := jira.BasicAuthTransport{
		Username: config.Username,
		Password: config.APIToken,
	}

	jiraClient, err := jira.NewClient(tp.Client(), config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira client: %w", err)
	}

	client := NewRealJiraClient(jiraClient)

	connector := &Connector{
		client:    client,
		config:    config,
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	return connector, nil
}

// SyncIssues syncs issues from configured projects
func (jc *Connector) SyncIssues(ctx context.Context) ([]Issue, error) {
	// Update sync status
	jc.statusMu.Lock()
	jc.status.SyncInProgress = true
	jc.statusMu.Unlock()

	var allIssues []Issue
	totalProjects := len(jc.config.Projects)

	for _, projectKey := range jc.config.Projects {
		issues, err := jc.getProjectIssues(ctx, projectKey)
		if err != nil {
			log.Printf("Warning: Failed to sync project %s: %v", projectKey, err)
			jc.updateSyncStatus(0, 0, err)
			continue
		}

		allIssues = append(allIssues, issues...)
	}

	jc.updateSyncStatus(len(allIssues), totalProjects, nil)
	return allIssues, nil
}

// getProjectIssues retrieves all issues for a specific project
func (jc *Connector) getProjectIssues(ctx context.Context, projectKey string) ([]Issue, error) {
	jql := fmt.Sprintf("project = %s ORDER BY updated DESC", projectKey)

	var allIssues []Issue
	startAt := 0
	maxResults := 50 // Jira API pagination size

	for {
		if len(allIssues) >= jc.config.MaxIssues {
			break
		}

		issues, _, err := jc.client.SearchIssues(jql, &jira.SearchOptions{
			StartAt:    startAt,
			MaxResults: maxResults,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to search issues: %w", err)
		}

		if len(issues) == 0 {
			break
		}

		for _, jiraIssue := range issues {
			issue := jc.convertIssue(&jiraIssue)
			allIssues = append(allIssues, issue)

			if len(allIssues) >= jc.config.MaxIssues {
				break
			}
		}

		if len(issues) < maxResults {
			break
		}

		startAt += maxResults
	}

	return allIssues, nil
}

// SearchIssues searches for issues using JQL
func (jc *Connector) SearchIssues(ctx context.Context, jql string) ([]Issue, error) {
	issues, _, err := jc.client.SearchIssues(jql, &jira.SearchOptions{
		MaxResults: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}

	var results []Issue
	for _, jiraIssue := range issues {
		results = append(results, jc.convertIssue(&jiraIssue))
	}

	return results, nil
}

// GetIssue retrieves a single issue by key
func (jc *Connector) GetIssue(ctx context.Context, issueKey string) (*Issue, error) {
	jiraIssue, _, err := jc.client.GetIssue(issueKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	issue := jc.convertIssue(jiraIssue)
	return &issue, nil
}

// GetIssueComments retrieves all comments for an issue
func (jc *Connector) GetIssueComments(ctx context.Context, issueKey string) ([]Comment, error) {
	jiraIssue, _, err := jc.client.GetIssue(issueKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	var comments []Comment
	if jiraIssue.Fields != nil && jiraIssue.Fields.Comments != nil {
		for _, c := range jiraIssue.Fields.Comments.Comments {
			author := ""
			if c.Author.DisplayName != "" {
				author = c.Author.DisplayName
			}

			createdAt, _ := time.Parse(time.RFC3339, c.Created)
			updatedAt, _ := time.Parse(time.RFC3339, c.Updated)

			comment := Comment{
				ID:        c.ID,
				IssueKey:  issueKey,
				Author:    author,
				Body:      c.Body,
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			}

			comments = append(comments, comment)
		}
	}

	return comments, nil
}

// ListProjects lists all accessible projects
func (jc *Connector) ListProjects(ctx context.Context) ([]Project, error) {
	jiraProjects, _, err := jc.client.ListProjects()
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var projects []Project
	for _, jp := range *jiraProjects {
		project := Project{
			ID:          jp.ID,
			Key:         jp.Key,
			Name:        jp.Name,
			Description: "", // Description not available in ProjectList
			Lead:        "", // Lead not available in ProjectList
			Type:        jp.ProjectTypeKey,
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// convertIssue converts a Jira API issue to our internal format
func (jc *Connector) convertIssue(jiraIssue *jira.Issue) Issue {
	fields := jiraIssue.Fields
	if fields == nil {
		fields = &jira.IssueFields{}
	}

	var labels []string
	if fields.Labels != nil {
		labels = fields.Labels
	}

	var components []string
	if fields.Components != nil {
		for _, c := range fields.Components {
			components = append(components, c.Name)
		}
	}

	var fixVersions []string
	if fields.FixVersions != nil {
		for _, v := range fields.FixVersions {
			fixVersions = append(fixVersions, v.Name)
		}
	}

	assignee := ""
	if fields.Assignee != nil {
		assignee = fields.Assignee.DisplayName
	}

	reporter := ""
	if fields.Reporter != nil {
		reporter = fields.Reporter.DisplayName
	}

	status := ""
	if fields.Status != nil {
		status = fields.Status.Name
	}

	priority := ""
	if fields.Priority != nil {
		priority = fields.Priority.Name
	}

	issueType := ""
	if fields.Type.Name != "" {
		issueType = fields.Type.Name
	}

	project := ""
	if fields.Project.Key != "" {
		project = fields.Project.Key
	}

	createdAt := time.Time(fields.Created)
	updatedAt := time.Time(fields.Updated)

	var resolvedAt *time.Time
	resolutionTime := time.Time(fields.Resolutiondate)
	if !resolutionTime.IsZero() {
		resolvedAt = &resolutionTime
	}

	return Issue{
		ID:          jiraIssue.ID,
		Key:         jiraIssue.Key,
		Summary:     fields.Summary,
		Description: fields.Description,
		Status:      status,
		Priority:    priority,
		IssueType:   issueType,
		Assignee:    assignee,
		Reporter:    reporter,
		Labels:      labels,
		Components:  components,
		FixVersions: fixVersions,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		ResolvedAt:  resolvedAt,
		Project:     project,
	}
}

// GetType returns the connector type
func (jc *Connector) GetType() string {
	return "jira"
}

// GetRateLimit returns current rate limit information
func (jc *Connector) GetRateLimit() *RateLimitInfo {
	jc.rateLimitMu.RLock()
	defer jc.rateLimitMu.RUnlock()

	if jc.rateLimit != nil {
		return &RateLimitInfo{
			Remaining: jc.rateLimit.Remaining,
			Reset:     jc.rateLimit.Reset,
		}
	}
	return &RateLimitInfo{}
}

// GetSyncStatus returns current sync status
func (jc *Connector) GetSyncStatus() *SyncStatus {
	jc.statusMu.RLock()
	defer jc.statusMu.RUnlock()

	return &SyncStatus{
		LastSync:       jc.status.LastSync,
		TotalIssues:    jc.status.TotalIssues,
		TotalProjects:  jc.status.TotalProjects,
		SyncInProgress: jc.status.SyncInProgress,
		Error:          jc.status.Error,
		RateLimit:      jc.GetRateLimit(),
	}
}

// updateSyncStatus updates the sync status
func (jc *Connector) updateSyncStatus(totalIssues, totalProjects int, err error) {
	jc.statusMu.Lock()
	defer jc.statusMu.Unlock()

	jc.status.LastSync = time.Now()
	jc.status.TotalIssues = totalIssues
	jc.status.TotalProjects = totalProjects
	jc.status.SyncInProgress = false

	if err != nil {
		jc.status.Error = err.Error()
	} else {
		jc.status.Error = ""
	}
}
