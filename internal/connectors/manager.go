package connectors

import (
	"context"
	"fmt"

	"github.com/ferg-cod3s/conexus/internal/connectors/github"
)

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// ConnectorManager manages different types of connectors
type ConnectorManager struct {
	store      ConnectorStore
	connectors map[string]interface{} // connector ID -> connector instance
}

// NewConnectorManager creates a new connector manager
func NewConnectorManager(store ConnectorStore) *ConnectorManager {
	return &ConnectorManager{
		store:      store,
		connectors: make(map[string]interface{}),
	}
}

// GetConnector gets or creates a connector instance
func (cm *ConnectorManager) GetConnector(ctx context.Context, id string) (interface{}, error) {
	// Check if we already have the connector instance
	if conn, exists := cm.connectors[id]; exists {
		return conn, nil
	}

	// Get connector config from store
	connector, err := cm.store.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector %s: %w", id, err)
	}

	// Create connector instance based on type
	var instance interface{}
	switch connector.Type {
	case "github":
		config := &github.Config{
			Token:      getStringFromMap(connector.Config, "token"),
			Repository: getStringFromMap(connector.Config, "repository"),
		}

		instance, err = github.NewConnector(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub connector: %w", err)
		}

	case "filesystem":
		// For filesystem, we might not need a special instance
		instance = nil

	default:
		return nil, fmt.Errorf("unsupported connector type: %s", connector.Type)
	}

	// Cache the instance
	cm.connectors[id] = instance
	return instance, nil
}

// SyncGitHubIssues syncs issues from a GitHub connector
func (cm *ConnectorManager) SyncGitHubIssues(ctx context.Context, connectorID string) ([]github.Issue, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.SyncIssues(ctx)
}

// SyncGitHubPullRequests syncs pull requests from a GitHub connector
func (cm *ConnectorManager) SyncGitHubPullRequests(ctx context.Context, connectorID string) ([]github.PullRequest, error) {
	conn, err := cm.GetConnector(ctx, connectorID)
	if err != nil {
		return nil, err
	}

	githubConn, ok := conn.(*github.Connector)
	if !ok {
		return nil, fmt.Errorf("connector %s is not a GitHub connector", connectorID)
	}

	return githubConn.SyncPullRequests(ctx)
}
