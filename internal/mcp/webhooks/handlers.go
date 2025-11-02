package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/connectors"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// WebhookHandler handles GitHub webhook events
type WebhookHandler struct {
	connectorStore connectors.ConnectorStore
	embedder       embedding.Embedder
	vectorStore    vectorstore.VectorStore
	errorHandler   *observability.ErrorHandler
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	connectorStore connectors.ConnectorStore,
	embedder embedding.Embedder,
	vectorStore vectorstore.VectorStore,
	errorHandler *observability.ErrorHandler,
) *WebhookHandler {
	return &WebhookHandler{
		connectorStore: connectorStore,
		embedder:       embedder,
		vectorStore:    vectorStore,
		errorHandler:   errorHandler,
	}
}

// HandleGitHubWebhook handles incoming GitHub webhook events
func (wh *WebhookHandler) HandleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Extract event type from header
	eventType := r.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "Missing X-GitHub-Event header", http.StatusBadRequest)
		return
	}

	// Extract signature for verification
	signature := r.Header.Get("X-Hub-Signature-256")

	// Read request body
	defer r.Body.Close()
	var payload []byte
	if r.Body != nil {
		payload = make([]byte, r.ContentLength)
		_, err := r.Body.Read(payload)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
	}

	// Get connector ID from URL path or query parameter
	connectorID := extractConnectorID(r)
	if connectorID == "" {
		http.Error(w, "Missing connector ID", http.StatusBadRequest)
		return
	}

	// Get connector configuration
	connector, err := wh.connectorStore.Get(ctx, connectorID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get connector: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify webhook signature if secret is configured
	if connector.Config["webhook_secret"] != nil {
		secret := connector.Config["webhook_secret"].(string)
		if !verifyWebhookSignature(payload, signature, secret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Process the webhook event
	err = wh.processWebhookEvent(ctx, connector, eventType, payload)
	if err != nil {
		wh.errorHandler.HandleError(ctx, err, observability.ExtractErrorContext(ctx, "webhook"))
		http.Error(w, fmt.Sprintf("Failed to process webhook: %v", err), http.StatusInternalServerError)
		return
	}

	// Log successful webhook processing
	if wh.errorHandler != nil {
		successCtx := observability.ExtractErrorContext(ctx, "webhook")
		successCtx.ErrorType = "success"
		successCtx.Duration = time.Since(startTime)
		wh.errorHandler.HandleError(ctx, nil, successCtx)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook processed successfully"))
}

// processWebhookEvent processes a webhook event and updates the vector store
func (wh *WebhookHandler) processWebhookEvent(ctx context.Context, connector *connectors.Connector, eventType string, payload []byte) error {
	switch eventType {
	case "issues", "issue_comment":
		return wh.handleIssueEvent(ctx, connector, payload)
	case "pull_request", "pull_request_review":
		return wh.handlePullRequestEvent(ctx, connector, payload)
	case "discussion", "discussion_comment":
		return wh.handleDiscussionEvent(ctx, connector, payload)
	case "push":
		return wh.handlePushEvent(ctx, connector, payload)
	default:
		log.Printf("Unhandled webhook event type: %s", eventType)
		return nil
	}
}

// handleIssueEvent handles issue-related webhook events
func (wh *WebhookHandler) handleIssueEvent(ctx context.Context, connector *connectors.Connector, payload []byte) error {
	var issueEvent map[string]interface{}
	if err := json.Unmarshal(payload, &issueEvent); err != nil {
		return fmt.Errorf("failed to parse issue event: %w", err)
	}

	issue, ok := issueEvent["issue"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid issue event format")
	}

	// Extract issue data
	number, _ := issue["number"].(float64)
	title, _ := issue["title"].(string)
	body, _ := issue["body"].(string)
	state, _ := issue["state"].(string)
	user, _ := issue["user"].(map[string]interface{})
	author, _ := user["login"].(string)

	// Extract labels
	var labels []string
	if labelsData, ok := issue["labels"].([]interface{}); ok {
		for _, label := range labelsData {
			if labelMap, ok := label.(map[string]interface{}); ok {
				if name, ok := labelMap["name"].(string); ok {
					labels = append(labels, name)
				}
			}
		}
	}

	// Create document content
	content := fmt.Sprintf("%s\n\n%s", title, body)
	if content == "" {
		content = title
	}

	// Generate embedding
	embedding, err := wh.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create document
	doc := vectorstore.Document{
		ID:      fmt.Sprintf("github-issue-%d", int(number)),
		Content: content,
		Vector:  embedding.Vector,
		Metadata: map[string]interface{}{
			"source_type":  "github_issue",
			"issue_number": int(number),
			"title":        title,
			"state":        state,
			"labels":       labels,
			"author":       author,
			"connector_id": connector.ID,
			"updated_at":   time.Now().Format(time.RFC3339),
		},
		UpdatedAt: time.Now(),
	}

	// Store in vector store
	return wh.vectorStore.Upsert(ctx, doc)
}

// handlePullRequestEvent handles pull request-related webhook events
func (wh *WebhookHandler) handlePullRequestEvent(ctx context.Context, connector *connectors.Connector, payload []byte) error {
	var prEvent map[string]interface{}
	if err := json.Unmarshal(payload, &prEvent); err != nil {
		return fmt.Errorf("failed to parse PR event: %w", err)
	}

	pr, ok := prEvent["pull_request"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid PR event format")
	}

	// Extract PR data
	number, _ := pr["number"].(float64)
	title, _ := pr["title"].(string)
	body, _ := pr["body"].(string)
	state, _ := pr["state"].(string)
	merged, _ := pr["merged"].(bool)
	user, _ := pr["user"].(map[string]interface{})
	author, _ := user["login"].(string)

	// Extract branch information
	head, _ := pr["head"].(map[string]interface{})
	headBranch, _ := head["ref"].(string)
	base, _ := pr["base"].(map[string]interface{})
	baseBranch, _ := base["ref"].(string)

	// Extract labels
	var labels []string
	if labelsData, ok := pr["labels"].([]interface{}); ok {
		for _, label := range labelsData {
			if labelMap, ok := label.(map[string]interface{}); ok {
				if name, ok := labelMap["name"].(string); ok {
					labels = append(labels, name)
				}
			}
		}
	}

	// Create document content
	content := fmt.Sprintf("%s\n\n%s", title, body)
	if content == "" {
		content = title
	}

	// Generate embedding
	embedding, err := wh.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create document
	doc := vectorstore.Document{
		ID:      fmt.Sprintf("github-pr-%d", int(number)),
		Content: content,
		Vector:  embedding.Vector,
		Metadata: map[string]interface{}{
			"source_type":  "github_pr",
			"pr_number":    int(number),
			"title":        title,
			"state":        state,
			"merged":       merged,
			"labels":       labels,
			"author":       author,
			"head_branch":  headBranch,
			"base_branch":  baseBranch,
			"connector_id": connector.ID,
			"updated_at":   time.Now().Format(time.RFC3339),
		},
		PRNumbers: []string{fmt.Sprintf("%d", int(number))},
		UpdatedAt: time.Now(),
	}

	// Store in vector store
	return wh.vectorStore.Upsert(ctx, doc)
}

// handleDiscussionEvent handles discussion-related webhook events
func (wh *WebhookHandler) handleDiscussionEvent(ctx context.Context, connector *connectors.Connector, payload []byte) error {
	var discussionEvent map[string]interface{}
	if err := json.Unmarshal(payload, &discussionEvent); err != nil {
		return fmt.Errorf("failed to parse discussion event: %w", err)
	}

	discussion, ok := discussionEvent["discussion"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid discussion event format")
	}

	// Extract discussion data
	number, _ := discussion["number"].(float64)
	title, _ := discussion["title"].(string)
	body, _ := discussion["body"].(string)
	user, _ := discussion["user"].(map[string]interface{})
	author, _ := user["login"].(string)

	// Create document content
	content := fmt.Sprintf("%s\n\n%s", title, body)
	if content == "" {
		content = title
	}

	// Generate embedding
	embedding, err := wh.embedder.Embed(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Create document
	doc := vectorstore.Document{
		ID:      fmt.Sprintf("github-discussion-%d", int(number)),
		Content: content,
		Vector:  embedding.Vector,
		Metadata: map[string]interface{}{
			"source_type":       "github_discussion",
			"discussion_number": int(number),
			"title":             title,
			"author":            author,
			"connector_id":      connector.ID,
			"updated_at":        time.Now().Format(time.RFC3339),
		},
		UpdatedAt: time.Now(),
	}

	// Store in vector store
	return wh.vectorStore.Upsert(ctx, doc)
}

// handlePushEvent handles push events (for code changes)
func (wh *WebhookHandler) handlePushEvent(ctx context.Context, connector *connectors.Connector, payload []byte) error {
	// Push events are handled by the file system watcher/indexer
	// We just log them for now
	log.Printf("Push event received for connector %s", connector.ID)
	return nil
}

// extractConnectorID extracts connector ID from request
func extractConnectorID(r *http.Request) string {
	// Try to get from URL path first
	path := strings.TrimPrefix(r.URL.Path, "/webhooks/github/")
	if path != "" && path != r.URL.Path {
		return path
	}

	// Try query parameter
	return r.URL.Query().Get("connector_id")
}

// verifyWebhookSignature verifies GitHub webhook signature
func verifyWebhookSignature(payload []byte, signature, secret string) bool {
	if secret == "" {
		return true // No secret configured
	}

	if signature == "" {
		return false
	}

	// Expected format: sha256=<hex>
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedSignature := signature[7:] // Remove "sha256=" prefix

	// Generate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	actualSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(actualSignature))
}
