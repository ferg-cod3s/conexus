# Story Context Implementation Guide

This guide provides step-by-step instructions to enhance Conexus for better story/feature context availability, transforming it from a code indexer into a comprehensive context engine.

## Quick Start: Immediate Wins (First 2 Days)

### 1. Enhanced Metadata Schema

**File**: `internal/vectorstore/store.go`

Add these fields to the `Document` struct:

```go
type Document struct {
    ID        string                 `json:"id"`
    Content   string                 `json:"content"`
    Vector    []float32              `json:"vector"`
    Metadata  map[string]interface{} `json:"metadata"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
    
    // New fields for story context
    StoryIDs     []string `json:"story_ids,omitempty"`
    TicketIDs    []string `json:"ticket_ids,omitempty"`
    PRNumbers    []string `json:"pr_numbers,omitempty"`
    DiscussionID string    `json:"discussion_id,omitempty"`
    BranchName   string    `json:"branch_name,omitempty"`
}
```

### 2. Basic Story Reference Extraction

**File**: `internal/enrichment/story_extractor.go` (new)

```go
package enrichment

import (
    "regexp"
    "strings"
)

type StoryExtractor struct {
    issuePattern   *regexp.Regexp
    prPattern      *regexp.Regexp
    branchPattern  *regexp.Regexp
}

func NewStoryExtractor() *StoryExtractor {
    return &StoryExtractor{
        issuePattern:  regexp.MustCompile(`(?:#|PROJ-|JIRA-)(\d+)`),
        prPattern:     regexp.MustCompile(`(?:#|pull/)(\d+)`),
        branchPattern: regexp.MustCompile(`(?:feature|bugfix|hotfix)\/([A-Z]+-\d+)`),
    }
}

func (se *StoryExtractor) ExtractStoryReferences(content string) map[string][]string {
    references := make(map[string][]string)
    
    // Extract issue references
    if matches := se.issuePattern.FindAllStringSubmatch(content, -1); matches != nil {
        for _, match := range matches {
            if len(match) > 1 {
                references["issues"] = append(references["issues"], match[1])
            }
        }
    }
    
    // Extract PR references
    if matches := se.prPattern.FindAllStringSubmatch(content, -1); matches != nil {
        for _, match := range matches {
            if len(match) > 1 {
                references["prs"] = append(references["prs"], match[1])
            }
        }
    }
    
    // Extract branch references
    if matches := se.branchPattern.FindAllStringSubmatch(content, -1); matches != nil {
        for _, match := range matches {
            if len(match) > 1 {
                references["branches"] = append(references["branches"], match[1])
            }
        }
    }
    
    return references
}
```

### 3. Update Indexing Pipeline

**File**: `internal/indexer/chunker.go`

Add story extraction to the chunking process:

```go
// Add to existing chunker struct
type Chunker struct {
    // ... existing fields ...
    storyExtractor *enrichment.StoryExtractor
}

// Update NewChunker function
func NewChunker(config ChunkerConfig) *Chunker {
    return &Chunker{
        // ... existing initialization ...
        storyExtractor: enrichment.NewStoryExtractor(),
    }
}

// Add to chunk processing
func (c *Chunker) ProcessFile(filePath string, content []byte) ([]Chunk, error) {
    chunks, err := c.chunkContent(filePath, content)
    if err != nil {
        return nil, err
    }
    
    // Extract story references from content
    storyRefs := c.storyExtractor.ExtractStoryReferences(string(content))
    
    // Add story references to chunk metadata
    for i := range chunks {
        if len(storyRefs["issues"]) > 0 {
            chunks[i].Metadata["story_ids"] = storyRefs["issues"]
        }
        if len(storyRefs["prs"]) > 0 {
            chunks[i].Metadata["pr_numbers"] = storyRefs["prs"]
        }
        if len(storyRefs["branches"]) > 0 {
            chunks[i].Metadata["branch_names"] = storyRefs["branches"]
        }
    }
    
    return chunks, nil
}
```

### 4. Enhanced Search with Story Context

**File**: `internal/mcp/handlers.go`

Update the search handler to use story context:

```go
// Update handleContextSearch function
func (s *Server) handleContextSearch(ctx context.Context, args json.RawMessage) (interface{}, error) {
    // ... existing validation code ...
    
    // Prepare search options
    opts := vectorstore.SearchOptions{
        Limit:   topK,
        Offset:  offset,
        Filters: make(map[string]interface{}),
    }
    
    // Add story context filtering if provided
    if req.Filters != nil && req.Filters.WorkContext != nil {
        if len(req.Filters.WorkContext.OpenTicketIDs) > 0 {
            opts.Filters["story_ids"] = req.Filters.WorkContext.OpenTicketIDs
        }
    }
    
    // ... rest of search logic ...
}
```

## Phase 1: GitHub Connector (Week 1)

### 1. Create GitHub Connector

**File**: `internal/connectors/github/github.go` (new)

```go
package github

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/go-github/v45/github"
    "golang.org/x/oauth2"
)

type Connector struct {
    client *github.Client
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

func NewConnector(config *Config) (*Connector, error) {
    if config.Token == "" {
        return nil, fmt.Errorf("GitHub token is required")
    }
    
    ts := oauth2.StaticTokenSource(
        &oauth2.Token{AccessToken: config.Token},
    )
    tc := oauth2.NewClient(context.Background(), ts)
    client := github.NewClient(tc)
    
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
        issues, resp, err := gc.client.Issues.ListByRepo(ctx, owner, repo, opts)
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
                
                allIssues = append(allIssues, Issue{
                    ID:          issue.GetID(),
                    Number:      issue.GetNumber(),
                    Title:       issue.GetTitle(),
                    Description: issue.GetBody(),
                    State:       issue.GetState(),
                    Labels:      labels,
                    Assignee:    assignee,
                    CreatedAt:   issue.GetCreatedAt().Time,
                    UpdatedAt:   issue.GetUpdatedAt().Time,
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

func parseRepository(repo string) (owner, name string) {
    parts := strings.Split(repo, "/")
    if len(parts) >= 2 {
        return parts[0], parts[1]
    }
    return "", repo
}
```

### 2. Register GitHub Connector

**File**: `internal/connectors/store.go`

Add "github" to valid connector types:

```go
// Update validateConnector function
func validateConnector(connector *Connector) error {
    validTypes := map[string]bool{
        "filesystem": true,
        "github":     true, // Add this
        "git":        true,
        "database":   true,
        "api":        true,
        "s3":         true,
        "http":       true,
    }
    
    if !validTypes[connector.Type] {
        return fmt.Errorf("invalid connector type: %s", connector.Type)
    }
    
    // ... rest of validation ...
}
```

### 3. Add GitHub Sync to Indexer

**File**: `internal/indexer/controller.go`

Add GitHub sync capability:

```go
// Add to IndexController struct
type IndexController struct {
    // ... existing fields ...
    githubConnector *github.Connector
}

// Add GitHub sync method
func (ic *IndexController) SyncGitHubIssues(ctx context.Context) error {
    if ic.githubConnector == nil {
        return fmt.Errorf("GitHub connector not configured")
    }
    
    issues, err := ic.githubConnector.SyncIssues(ctx)
    if err != nil {
        return fmt.Errorf("failed to sync GitHub issues: %w", err)
    }
    
    // Convert issues to documents and store
    for _, issue := range issues {
        doc := vectorstore.Document{
            ID:      fmt.Sprintf("github-issue-%d", issue.Number),
            Content: fmt.Sprintf("%s\n\n%s", issue.Title, issue.Description),
            Metadata: map[string]interface{}{
                "source_type": "github_issue",
                "issue_number": issue.Number,
                "title":        issue.Title,
                "state":        issue.State,
                "labels":       issue.Labels,
                "assignee":     issue.Assignee,
                "created_at":   issue.CreatedAt,
                "updated_at":   issue.UpdatedAt,
                "url":          fmt.Sprintf("https://github.com/%s/issues/%d", ic.githubConnector.config.Repository, issue.Number),
            },
        }
        
        if err := ic.vectorStore.Upsert(ctx, doc); err != nil {
            return fmt.Errorf("failed to store issue %d: %w", issue.Number, err)
        }
    }
    
    return nil
}
```

## Phase 2: Work Context Tracking (Week 2)

### 1. Work Context Service

**File**: `internal/context/work_context.go` (new)

```go
package context

import (
    "context"
    "fmt"
    "sync"
    "time"
)

type WorkContext struct {
    SessionID     string            `json:"session_id"`
    ActiveFile    string            `json:"active_file,omitempty"`
    GitBranch     string            `json:"git_branch,omitempty"`
    OpenTicketIDs []string          `json:"open_ticket_ids,omitempty"`
    CurrentStory  string            `json:"current_story,omitempty"`
    LastActivity  time.Time         `json:"last_activity"`
    TeamContext   map[string]string `json:"team_context,omitempty"`
}

type Service struct {
    contexts map[string]*WorkContext
    mutex    sync.RWMutex
}

func NewService() *Service {
    return &Service{
        contexts: make(map[string]*WorkContext),
    }
}

func (s *Service) UpdateContext(ctx context.Context, sessionID string, update *WorkContext) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    current, exists := s.contexts[sessionID]
    if !exists {
        current = &WorkContext{
            SessionID: sessionID,
        }
    }
    
    // Apply updates
    if update.ActiveFile != "" {
        current.ActiveFile = update.ActiveFile
    }
    if update.GitBranch != "" {
        current.GitBranch = update.GitBranch
    }
    if len(update.OpenTicketIDs) > 0 {
        current.OpenTicketIDs = update.OpenTicketIDs
    }
    if update.CurrentStory != "" {
        current.CurrentStory = update.CurrentStory
    }
    if update.TeamContext != nil {
        if current.TeamContext == nil {
            current.TeamContext = make(map[string]string)
        }
        for k, v := range update.TeamContext {
            current.TeamContext[k] = v
        }
    }
    
    current.LastActivity = time.Now()
    s.contexts[sessionID] = current
    
    return nil
}

func (s *Service) GetContext(ctx context.Context, sessionID string) (*WorkContext, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    context, exists := s.contexts[sessionID]
    if !exists {
        return nil, fmt.Errorf("work context not found for session: %s", sessionID)
    }
    
    return context, nil
}

func (s *Service) InferStoryFromBranch(branch string) string {
    // Pattern: feature/PROJ-123-description
    re := regexp.MustCompile(`(?:feature|bugfix|hotfix)\/([A-Z]+-\d+)`)
    if matches := re.FindStringSubmatch(branch); len(matches) > 1 {
        return matches[1]
    }
    
    // Pattern: 123-feature-description
    re = regexp.MustCompile(`^(\d+)-`)
    if matches := re.FindStringSubmatch(branch); len(matches) > 1 {
        return matches[1]
    }
    
    return ""
}
```

### 2. Add Work Context MCP Tool

**File**: `internal/mcp/schema.go`

Add new tool definition:

```go
// Add to toolDefinitions
{
    Name:        "work_context_update",
    Description: "Update current work context for better search results",
    InputSchema: json.RawMessage(`{
        "type": "object",
        "properties": {
            "session_id": {
                "type": "string",
                "description": "Session identifier"
            },
            "active_file": {
                "type": "string",
                "description": "Currently active file"
            },
            "git_branch": {
                "type": "string",
                "description": "Current git branch"
            },
            "current_story": {
                "type": "string",
                "description": "Currently working on story"
            }
        },
        "required": ["session_id"]
    }`),
},
```

### 3. Add Work Context Handler

**File**: `internal/mcp/handlers.go`

Add the handler:

```go
// Add to Server struct
type Server struct {
    // ... existing fields ...
    workContextService *context.Service
}

// Add handler function
func (s *Server) handleWorkContextUpdate(ctx context.Context, args json.RawMessage) (interface{}, error) {
    var req WorkContextUpdateRequest
    if err := json.Unmarshal(args, &req); err != nil {
        return nil, &protocol.Error{
            Code:    protocol.InvalidParams,
            Message: fmt.Sprintf("invalid request: %v", err),
        }
    }
    
    if req.SessionID == "" {
        return nil, &protocol.Error{
            Code:    protocol.InvalidParams,
            Message: "session_id is required",
        }
    }
    
    workContext := &context.WorkContext{
        SessionID:    req.SessionID,
        ActiveFile:   req.ActiveFile,
        GitBranch:    req.GitBranch,
        CurrentStory: req.CurrentStory,
    }
    
    // Infer story from branch if not provided
    if workContext.CurrentStory == "" && workContext.GitBranch != "" {
        inferredStory := s.workContextService.InferStoryFromBranch(workContext.GitBranch)
        if inferredStory != "" {
            workContext.CurrentStory = inferredStory
        }
    }
    
    if err := s.workContextService.UpdateContext(ctx, req.SessionID, workContext); err != nil {
        return nil, &protocol.Error{
            Code:    protocol.InternalError,
            Message: fmt.Sprintf("failed to update work context: %v", err),
        }
    }
    
    return map[string]interface{}{
        "status":  "ok",
        "message": "Work context updated successfully",
        "inferred_story": workContext.CurrentStory,
    }, nil
}
```

## Usage Examples

### 1. Basic Story-Aware Search

```bash
# Update work context
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "work_context_update",
      "arguments": {
        "session_id": "dev-session-123",
        "git_branch": "feature/PROJ-456-user-authentication",
        "active_file": "src/auth/authenticator.go"
      }
    }
  }'

# Search with story context
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "context_search",
      "arguments": {
        "query": "authentication logic",
        "work_context": {
          "session_id": "dev-session-123"
        }
      }
    }
  }'
```

### 2. GitHub Connector Setup

```yaml
# config.yml
connectors:
  github:
    enabled: true
    token: "${GITHUB_TOKEN}"
    repository: "your-org/your-repo"
    sync_interval: "5m"
```

```bash
# Add GitHub connector
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "context_connector_management",
      "arguments": {
        "action": "add",
        "connector_id": "github-main",
        "connector_config": {
          "type": "github",
          "token": "your-github-token",
          "repository": "your-org/your-repo"
        }
      }
    }
  }'
```

## Testing

### 1. Unit Tests

**File**: `internal/enrichment/story_extractor_test.go` (new)

```go
package enrichment

import (
    "testing"
)

func TestStoryExtractor_ExtractStoryReferences(t *testing.T) {
    extractor := NewStoryExtractor()
    
    tests := []struct {
        content string
        expected map[string][]string
    }{
        {
            content: "Fixes #123 and relates to PROJ-456",
            expected: map[string][]string{
                "issues": {"123", "456"},
            },
        },
        {
            content: "See pull/789 for details",
            expected: map[string][]string{
                "prs": {"789"},
            },
        },
        {
            content: "Working on feature/JIRA-999-user-login",
            expected: map[string][]string{
                "branches": {"JIRA-999"},
            },
        },
    }
    
    for _, test := range tests {
        result := extractor.ExtractStoryReferences(test.content)
        assert.Equal(t, test.expected, result)
    }
}
```

### 2. Integration Tests

**File**: `internal/testing/integration/story_context_test.go` (new)

```go
package integration

import (
    "testing"
    // ... other imports
)

func TestStoryContextSearch(t *testing.T) {
    // Setup test with GitHub connector
    store := vectorstore.NewMemoryStore()
    connector := &mockGitHubConnector{
        issues: []github.Issue{
            {
                Number: github.Int(123),
                Title:  github.String("Add user authentication"),
                Body:   github.String("Implement OAuth2 authentication"),
            },
        },
    }
    
    // Index issues
    indexer := indexer.NewIndexer(store, connector)
    err := indexer.SyncGitHubIssues(context.Background())
    require.NoError(t, err)
    
    // Test search with story context
    results, err := store.SearchHybrid(context.Background(), "authentication", nil, vectorstore.SearchOptions{
        Filters: map[string]interface{}{
            "story_ids": []string{"123"},
        },
    })
    require.NoError(t, err)
    assert.Greater(t, len(results), 0)
}
```

## Deployment

### 1. Environment Variables

```bash
export GITHUB_TOKEN="your-github-token"
export GITHUB_WEBHOOK_SECRET="your-webhook-secret"
export WORK_CONTEXT_SESSION_TIMEOUT="2h"
export STORY_SYNC_INTERVAL="5m"
```

### 2. Docker Compose

```yaml
version: '3.8'
services:
  conexus:
    build: .
    environment:
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - GITHUB_WEBHOOK_SECRET=${GITHUB_WEBHOOK_SECRET}
    ports:
      - "8080:8080"
    volumes:
      - ./config.yml:/app/config.yml
      - ./data:/app/data
```

## Monitoring

### 1. Metrics

Add these metrics to track story context effectiveness:

```go
// In internal/observability/metrics.go
var (
    StoryContextHitRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "conexus_story_context_hit_rate",
            Help: "Rate of successful story context matches",
        },
        []string{"session_id"},
    )
    
    GitHubSyncLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "conexus_github_sync_duration_seconds",
            Help: "Time taken to sync GitHub issues",
        },
        []string{"repository"},
    )
)
```

### 2. Health Checks

```go
// Add to health check endpoint
func (s *Server) checkStoryContextHealth() map[string]interface{} {
    health := map[string]interface{}{
        "status": "healthy",
        "details": map[string]interface{}{
            "github_connector": s.githubConnector != nil,
            "work_context_sessions": len(s.workContextService.contexts),
            "last_sync": s.lastGitHubSync,
        },
    }
    
    return health
}
```

## Next Steps

1. **Week 1**: Implement basic story reference extraction and GitHub connector
2. **Week 2**: Add work context tracking and enhanced search
3. **Week 3**: Implement Slack connector for discussion context
4. **Week 4**: Add relationship graph and advanced context inference
5. **Week 5**: Implement real-time webhook updates
6. **Week 6**: Performance optimization and comprehensive testing

This implementation guide provides the foundation for transforming Conexus into a story-aware context engine that significantly improves developer productivity by automatically maintaining relevant context for features and stories.