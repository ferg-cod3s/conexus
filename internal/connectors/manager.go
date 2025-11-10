// Package connectors provides lifecycle management for connectors.
package connectors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager manages the lifecycle of connectors.
type Manager struct {
	store        ConnectorStore
	hooks        *HookRegistry
	connectors   map[string]*Connector
	mu           sync.RWMutex
	shutdownOnce sync.Once
}

// NewManager creates a new connector manager.
func NewManager(store ConnectorStore) *Manager {
	return &Manager{
		store:      store,
		hooks:      NewHookRegistry(),
		connectors: make(map[string]*Connector),
	}
}

// RegisterHook registers a lifecycle hook.
func (m *Manager) RegisterHook(hook LifecycleHook) {
	m.hooks.RegisterPreInit(hook)
	m.hooks.RegisterPostInit(hook)
	m.hooks.RegisterPreShutdown(hook)
	m.hooks.RegisterPostShutdown(hook)
}

// Initialize initializes a connector with lifecycle hooks.
// Returns error if pre-init or post-init hooks fail.
func (m *Manager) Initialize(ctx context.Context, connector *Connector) error {
	if connector == nil {
		return fmt.Errorf("connector cannot be nil")
	}

	// Execute pre-init hooks
	if err := m.hooks.ExecutePreInit(ctx, connector); err != nil {
		return fmt.Errorf("pre-init failed: %w", err)
	}

	// Add to store
	if err := m.store.Add(ctx, connector); err != nil {
		return fmt.Errorf("store add failed: %w", err)
	}

	// Execute post-init hooks
	if err := m.hooks.ExecutePostInit(ctx, connector); err != nil {
		// Rollback: remove from store if post-init fails
		if removeErr := m.store.Remove(ctx, connector.ID); removeErr != nil {
			log.Printf("WARNING: Failed to rollback connector %s during post-init cleanup: %v", connector.ID, removeErr)
		}
		return fmt.Errorf("post-init failed: %w", err)
	}

	// Track in memory
	m.mu.Lock()
	m.connectors[connector.ID] = connector
	m.mu.Unlock()

	return nil
}

// Shutdown gracefully shuts down a connector with hooks.
// Uses provided context for timeout control.
func (m *Manager) Shutdown(ctx context.Context, connectorID string) error {
	if connectorID == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}

	// Get connector
	m.mu.RLock()
	connector, exists := m.connectors[connectorID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connector %s not found", connectorID)
	}

	// Execute pre-shutdown hooks
	if err := m.hooks.ExecutePreShutdown(ctx, connector); err != nil {
		return fmt.Errorf("pre-shutdown failed: %w", err)
	}

	// Remove from store
	if err := m.store.Remove(ctx, connectorID); err != nil {
		return fmt.Errorf("store remove failed: %w", err)
	}

	// Execute post-shutdown hooks (best-effort, collect errors)
	var shutdownErr error
	if err := m.hooks.ExecutePostShutdown(ctx, connector); err != nil {
		shutdownErr = fmt.Errorf("post-shutdown failed: %w", err)
	}

	// Remove from memory
	m.mu.Lock()
	delete(m.connectors, connectorID)
	m.mu.Unlock()

	return shutdownErr
}

// ShutdownAll gracefully shuts down all connectors with timeout.
// Default timeout is 30 seconds, can be overridden via context.
func (m *Manager) ShutdownAll(ctx context.Context) error {
	var shutdownErr error
	m.shutdownOnce.Do(func() {
		m.mu.RLock()
		connectorIDs := make([]string, 0, len(m.connectors))
		for id := range m.connectors {
			connectorIDs = append(connectorIDs, id)
		}
		m.mu.RUnlock()

		// Shutdown connectors in parallel with timeout
		var wg sync.WaitGroup
		errChan := make(chan error, len(connectorIDs))

		for _, id := range connectorIDs {
			wg.Add(1)
			go func(connectorID string) {
				defer wg.Done()
				if err := m.Shutdown(ctx, connectorID); err != nil {
					errChan <- fmt.Errorf("connector %s: %w", connectorID, err)
				}
			}(id)
		}

		// Wait for all shutdowns to complete
		wg.Wait()
		close(errChan)

		// Collect errors
		var errs []error
		for err := range errChan {
			errs = append(errs, err)
		}

		if len(errs) > 0 {
			shutdownErr = fmt.Errorf("shutdown errors: %v", errs)
		}
	})

	return shutdownErr
}

// Get retrieves a connector by ID from memory or store.
func (m *Manager) Get(ctx context.Context, id string) (*Connector, error) {
	// Try memory first
	m.mu.RLock()
	connector, exists := m.connectors[id]
	m.mu.RUnlock()

	if exists {
		return connector, nil
	}

	// Fall back to store
	return m.store.Get(ctx, id)
}

// List returns all active connectors from memory.
// If no connectors in memory, falls back to store.
func (m *Manager) List(ctx context.Context) ([]*Connector, error) {
	m.mu.RLock()
	count := len(m.connectors)
	m.mu.RUnlock()

	// If we have connectors in memory, return them
	if count > 0 {
		m.mu.RLock()
		defer m.mu.RUnlock()
		connectors := make([]*Connector, 0, len(m.connectors))
		for _, c := range m.connectors {
			connectors = append(connectors, c)
		}
		return connectors, nil
	}

	// Fall back to store
	return m.store.List(ctx)
}

// Update updates a connector's configuration with validation.
func (m *Manager) Update(ctx context.Context, id string, connector *Connector) error {
	if id == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}
	if connector == nil {
		return fmt.Errorf("connector cannot be nil")
	}

	// Validate with pre-init hooks
	if err := m.hooks.ExecutePreInit(ctx, connector); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Update in store
	if err := m.store.Update(ctx, id, connector); err != nil {
		return fmt.Errorf("store update failed: %w", err)
	}

	// Update in memory
	m.mu.Lock()
	if _, exists := m.connectors[id]; exists {
		m.connectors[id] = connector
	}
	m.mu.Unlock()

	return nil
}

// Close closes the manager and releases resources.
func (m *Manager) Close(ctx context.Context) error {
	// Shutdown all connectors
	if err := m.ShutdownAll(ctx); err != nil {
		return err
	}

	// Close store
	return m.store.Close()
}

// DrainTimeout is the default timeout for graceful shutdown.
const DrainTimeout = 30 * time.Second

// GracefulShutdown performs a graceful shutdown with default timeout.
func (m *Manager) GracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), DrainTimeout)
	defer cancel()

	return m.Close(ctx)
}

// ListActive returns all active connectors currently in memory.
// Unlike List(), this does not fall back to the store.
func (m *Manager) ListActive() []*Connector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	connectors := make([]*Connector, 0, len(m.connectors))
	for _, c := range m.connectors {
		connectors = append(connectors, c)
	}
	return connectors
}
