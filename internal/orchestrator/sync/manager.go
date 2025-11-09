package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/connectors/github"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// SyncManager manages synchronization of external data sources
type SyncManager struct {
	connectorStore connectors.ConnectorStore
	connectorMgr   *connectors.ConnectorManager
	embedder       embedding.Embedder
	vectorStore    vectorstore.VectorStore
	errorHandler   *observability.ErrorHandler

	// Sync state
	syncJobs   map[string]*SyncJob
	syncJobsMu sync.RWMutex
	isRunning  bool
	runningMu  sync.RWMutex
	stopChan   chan struct{}
}

// SyncJob represents an active sync job
type SyncJob struct {
	ID             string      `json:"id"`
	ConnectorID    string      `json:"connector_id"`
	Type           string      `json:"type"`   // "github", "jira", etc.
	Status         string      `json:"status"` // "running", "completed", "failed"
	StartedAt      time.Time   `json:"started_at"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty"`
	Progress       float64     `json:"progress"` // 0.0 to 1.0
	TotalItems     int         `json:"total_items"`
	ProcessedItems int         `json:"processed_items"`
	Error          string      `json:"error,omitempty"`
	Result         *SyncResult `json:"result,omitempty"`
}

// SyncResult contains the results of a sync operation
type SyncResult struct {
	IssuesSynced      int      `json:"issues_synced"`
	PRsSynced         int      `json:"prs_synced"`
	DiscussionsSynced int      `json:"discussions_synced"`
	Errors            []string `json:"errors,omitempty"`
}

// SyncStatus represents the overall sync status
type SyncStatus struct {
	IsRunning       bool       `json:"is_running"`
	ActiveJobs      []*SyncJob `json:"active_jobs"`
	CompletedJobs   []*SyncJob `json:"completed_jobs"`
	LastSyncTime    *time.Time `json:"last_sync_time,omitempty"`
	TotalSyncs      int        `json:"total_syncs"`
	SuccessfulSyncs int        `json:"successful_syncs"`
	FailedSyncs     int        `json:"failed_syncs"`
}

// NewSyncManager creates a new sync manager
func NewSyncManager(
	connectorStore connectors.ConnectorStore,
	connectorMgr *connectors.ConnectorManager,
	embedder embedding.Embedder,
	vectorStore vectorstore.VectorStore,
	errorHandler *observability.ErrorHandler,
) *SyncManager {
	return &SyncManager{
		connectorStore: connectorStore,
		connectorMgr:   connectorMgr,
		embedder:       embedder,
		vectorStore:    vectorStore,
		errorHandler:   errorHandler,
		syncJobs:       make(map[string]*SyncJob),
		stopChan:       make(chan struct{}),
	}
}

// Start starts the sync manager background processes
func (sm *SyncManager) Start(ctx context.Context) error {
	sm.runningMu.Lock()
	defer sm.runningMu.Unlock()

	if sm.isRunning {
		return fmt.Errorf("sync manager is already running")
	}

	sm.isRunning = true

	// Start background sync scheduler
	go sm.runSyncScheduler(ctx)

	log.Println("Sync manager started")
	return nil
}

// Stop stops the sync manager
func (sm *SyncManager) Stop(ctx context.Context) error {
	sm.runningMu.Lock()
	defer sm.runningMu.Unlock()

	if !sm.isRunning {
		return nil
	}

	sm.isRunning = false
	close(sm.stopChan)

	log.Println("Sync manager stopped")
	return nil
}

// TriggerSync triggers a manual sync for a specific connector
func (sm *SyncManager) TriggerSync(ctx context.Context, connectorID string) (*SyncJob, error) {
	// Get connector configuration
	connector, err := sm.connectorStore.Get(ctx, connectorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector: %w", err)
	}

	// Create sync job
	jobID := fmt.Sprintf("%s-%d", connectorID, time.Now().Unix())
	job := &SyncJob{
		ID:             jobID,
		ConnectorID:    connectorID,
		Type:           connector.Type,
		Status:         "running",
		StartedAt:      time.Now(),
		Progress:       0.0,
		TotalItems:     0,
		ProcessedItems: 0,
	}

	// Store job
	sm.syncJobsMu.Lock()
	sm.syncJobs[jobID] = job
	sm.syncJobsMu.Unlock()

	// Start sync in background
	go sm.runSyncJob(ctx, job)

	return job, nil
}

// GetSyncStatus returns the current sync status
func (sm *SyncManager) GetSyncStatus(ctx context.Context) (*SyncStatus, error) {
	sm.syncJobsMu.RLock()
	defer sm.syncJobsMu.RUnlock()

	var activeJobs, completedJobs []*SyncJob
	var lastSyncTime *time.Time
	totalSyncs := 0
	successfulSyncs := 0
	failedSyncs := 0

	for _, job := range sm.syncJobs {
		if job.Status == "running" {
			activeJobs = append(activeJobs, job)
		} else {
			completedJobs = append(completedJobs, job)

			totalSyncs++
			if job.Status == "completed" {
				successfulSyncs++
			} else if job.Status == "failed" {
				failedSyncs++
			}

			if job.CompletedAt != nil && (lastSyncTime == nil || job.CompletedAt.After(*lastSyncTime)) {
				lastSyncTime = job.CompletedAt
			}
		}
	}

	sm.runningMu.RLock()
	isRunning := sm.isRunning
	sm.runningMu.RUnlock()

	return &SyncStatus{
		IsRunning:       isRunning,
		ActiveJobs:      activeJobs,
		CompletedJobs:   completedJobs,
		LastSyncTime:    lastSyncTime,
		TotalSyncs:      totalSyncs,
		SuccessfulSyncs: successfulSyncs,
		FailedSyncs:     failedSyncs,
	}, nil
}

// GetSyncJob returns a specific sync job
func (sm *SyncManager) GetSyncJob(ctx context.Context, jobID string) (*SyncJob, error) {
	sm.syncJobsMu.RLock()
	defer sm.syncJobsMu.RUnlock()

	job, exists := sm.syncJobs[jobID]
	if !exists {
		return nil, fmt.Errorf("sync job %s not found", jobID)
	}

	return job, nil
}

// runSyncScheduler runs the background sync scheduler
func (sm *SyncManager) runSyncScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sm.stopChan:
			return
		case <-ticker.C:
			sm.checkAndRunScheduledSyncs(ctx)
		}
	}
}

// checkAndRunScheduledSyncs checks for connectors that need syncing
func (sm *SyncManager) checkAndRunScheduledSyncs(ctx context.Context) {
	// Get all active connectors
	connectors, err := sm.connectorStore.List(ctx)
	if err != nil {
		log.Printf("Failed to list connectors: %v", err)
		return
	}

	for _, connector := range connectors {
		if connector.Status != "active" {
			continue
		}

		// Check if connector needs syncing based on its configuration
		if sm.shouldSyncConnector(connector) {
			// Trigger sync
			_, err := sm.TriggerSync(ctx, connector.ID)
			if err != nil {
				log.Printf("Failed to trigger sync for connector %s: %v", connector.ID, err)
			}
		}
	}
}

// shouldSyncConnector determines if a connector needs syncing
func (sm *SyncManager) shouldSyncConnector(connector *connectors.Connector) bool {
	// Check if there's already an active sync for this connector
	sm.syncJobsMu.RLock()
	for _, job := range sm.syncJobs {
		if job.ConnectorID == connector.ID && job.Status == "running" {
			sm.syncJobsMu.RUnlock()
			return false
		}
	}
	sm.syncJobsMu.RUnlock()

	// Check sync interval from connector config
	syncInterval := 5 * time.Minute // Default
	if intervalStr, ok := connector.Config["sync_interval"].(string); ok {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			syncInterval = interval
		}
	}

	// Check last sync time for this connector
	sm.syncJobsMu.RLock()
	var lastSync *time.Time
	for _, job := range sm.syncJobs {
		if job.ConnectorID == connector.ID && job.Status == "completed" && job.CompletedAt != nil {
			if lastSync == nil || job.CompletedAt.After(*lastSync) {
				lastSync = job.CompletedAt
			}
		}
	}
	sm.syncJobsMu.RUnlock()

	// If no previous sync, should sync
	if lastSync == nil {
		return true
	}

	// If enough time has passed, should sync
	return time.Since(*lastSync) > syncInterval
}

// runSyncJob executes a sync job
func (sm *SyncManager) runSyncJob(ctx context.Context, job *SyncJob) {
	defer func() {
		// Update job status on completion
		now := time.Now()
		job.CompletedAt = &now
		if job.Status != "failed" {
			job.Status = "completed"
		}
		job.Progress = 1.0
	}()

	log.Printf("Starting sync job %s for connector %s", job.ID, job.ConnectorID)

	// Get connector instance
	connector, err := sm.connectorMgr.GetConnector(ctx, job.ConnectorID)
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("Failed to get connector: %v", err)
		return
	}

	// Perform sync based on connector type
	var result *SyncResult
	switch job.Type {
	case "github":
		result, err = sm.syncGitHub(ctx, connector)
	default:
		err = fmt.Errorf("unsupported connector type: %s", job.Type)
	}

	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		log.Printf("Sync job %s failed: %v", job.ID, err)
	} else {
		job.Result = result
		log.Printf("Sync job %s completed successfully", job.ID)
	}
}

// syncGitHub syncs GitHub data
func (sm *SyncManager) syncGitHub(ctx context.Context, connector interface{}) (*SyncResult, error) {
	githubConn, ok := connector.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector is not a GitHub connector")
	}

	result := &SyncResult{}
	var errors []string

	// Sync issues
	issues, err := githubConn.SyncIssues(ctx)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to sync issues: %v", err))
	} else {
		result.IssuesSynced = len(issues)

		// Store issues in vector store
		for _, issue := range issues {
			doc := vectorstore.Document{
				ID:      fmt.Sprintf("github-issue-%d", issue.Number),
				Content: fmt.Sprintf("%s\n\n%s", issue.Title, issue.Description),
				Metadata: map[string]interface{}{
					"source_type":  "github_issue",
					"issue_number": issue.Number,
					"title":        issue.Title,
					"state":        issue.State,
					"labels":       issue.Labels,
					"assignee":     issue.Assignee,
					"created_at":   issue.CreatedAt,
					"updated_at":   issue.UpdatedAt,
				},
				CreatedAt: issue.CreatedAt,
				UpdatedAt: issue.UpdatedAt,
			}

			// Generate embedding
			embedding, err := sm.embedder.Embed(ctx, doc.Content)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to embed issue %d: %v", issue.Number, err))
				continue
			}
			doc.Vector = embedding.Vector

			// Store in vector store
			if err := sm.vectorStore.Upsert(ctx, doc); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to store issue %d: %v", issue.Number, err))
			}
		}
	}

	// Sync pull requests
	prs, err := githubConn.SyncPullRequests(ctx)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to sync PRs: %v", err))
	} else {
		result.PRsSynced = len(prs)

		// Store PRs in vector store
		for _, pr := range prs {
			doc := vectorstore.Document{
				ID:      fmt.Sprintf("github-pr-%d", pr.Number),
				Content: fmt.Sprintf("%s\n\n%s", pr.Title, pr.Description),
				Metadata: map[string]interface{}{
					"source_type":   "github_pr",
					"pr_number":     pr.Number,
					"title":         pr.Title,
					"state":         pr.State,
					"labels":        pr.Labels,
					"assignee":      pr.Assignee,
					"created_at":    pr.CreatedAt,
					"updated_at":    pr.UpdatedAt,
					"linked_issues": pr.LinkedIssues,
					"merged":        pr.Merged,
					"head_branch":   pr.HeadBranch,
					"base_branch":   pr.BaseBranch,
				},
				CreatedAt: pr.CreatedAt,
				UpdatedAt: pr.UpdatedAt,
				PRNumbers: []string{fmt.Sprintf("%d", pr.Number)},
			}

			// Generate embedding
			embedding, err := sm.embedder.Embed(ctx, doc.Content)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to embed PR %d: %v", pr.Number, err))
				continue
			}
			doc.Vector = embedding.Vector

			// Store in vector store
			if err := sm.vectorStore.Upsert(ctx, doc); err != nil {
				errors = append(errors, fmt.Sprintf("Failed to store PR %d: %v", pr.Number, err))
			}
		}
	}

	// Sync discussions (placeholder for now)
	discussions, err := githubConn.SyncDiscussions(ctx)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to sync discussions: %v", err))
	} else {
		result.DiscussionsSynced = len(discussions)
	}

	result.Errors = errors
	return result, nil
}
