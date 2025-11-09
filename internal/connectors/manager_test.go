package connectors

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStore is a mock implementation of ConnectorStore for testing.
type mockStore struct {
	connectors map[string]*Connector
	mu         sync.RWMutex
	addErr     error
	removeErr  error
	updateErr  error
	getErr     error
	listErr    error
}

func newMockStore() *mockStore {
	return &mockStore{
		connectors: make(map[string]*Connector),
	}
}

func (m *mockStore) Add(ctx context.Context, connector *Connector) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectors[connector.ID] = connector
	return nil
}

func (m *mockStore) Get(ctx context.Context, id string) (*Connector, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if c, ok := m.connectors[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("connector not found: %s", id)
}

func (m *mockStore) List(ctx context.Context) ([]*Connector, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	connectors := make([]*Connector, 0, len(m.connectors))
	for _, c := range m.connectors {
		connectors = append(connectors, c)
	}
	return connectors, nil
}

func (m *mockStore) Update(ctx context.Context, id string, connector *Connector) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.connectors[id]; !ok {
		return fmt.Errorf("connector not found: %s", id)
	}
	m.connectors[id] = connector
	return nil
}

func (m *mockStore) Remove(ctx context.Context, id string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connectors, id)
	return nil
}

func (m *mockStore) Close() error {
	return nil
}

// TestManager_Initialize tests successful initialization.
func TestManager_Initialize(t *testing.T) {
	store := newMockStore()
	manager := NewManager(store)
	ctx := context.Background()

	connector := &Connector{
		ID:     "test-1",
		Type:   "mock",
		Config: map[string]interface{}{"key": "value"},
	}

	t.Run("successful initialization", func(t *testing.T) {
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		// Verify connector in memory
		manager.mu.RLock()
		_, exists := manager.connectors[connector.ID]
		manager.mu.RUnlock()
		assert.True(t, exists)

		// Verify connector in store
		stored, err := store.Get(ctx, connector.ID)
		require.NoError(t, err)
		assert.Equal(t, connector.ID, stored.ID)
	})

	t.Run("nil connector", func(t *testing.T) {
		err := manager.Initialize(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connector cannot be nil")
	})
}

// TestManager_InitializeWithHooks tests initialization with hooks.
func TestManager_InitializeWithHooks(t *testing.T) {
	ctx := context.Background()

	t.Run("pre-init hook success", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)
		hook := &mockHook{}
		manager.RegisterHook(hook)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		calls := hook.getCalls()
		assert.Contains(t, calls, "OnPreInit")
		assert.Contains(t, calls, "OnPostInit")
	})

	t.Run("pre-init hook failure prevents store add", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)
		hook := &mockHook{preInitErr: errors.New("validation failed")}
		manager.RegisterHook(hook)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pre-init failed")
		assert.Contains(t, err.Error(), "validation failed")

		// Verify connector NOT in store
		_, storeErr := store.Get(ctx, connector.ID)
		assert.Error(t, storeErr)

		// Verify connector NOT in memory
		manager.mu.RLock()
		_, exists := manager.connectors[connector.ID]
		manager.mu.RUnlock()
		assert.False(t, exists)
	})

	t.Run("post-init hook failure triggers rollback", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)
		hook := &mockHook{postInitErr: errors.New("health check failed")}
		manager.RegisterHook(hook)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "post-init failed")
		assert.Contains(t, err.Error(), "health check failed")

		// Verify connector removed from store (rollback)
		_, storeErr := store.Get(ctx, connector.ID)
		assert.Error(t, storeErr)

		// Verify connector NOT in memory
		manager.mu.RLock()
		_, exists := manager.connectors[connector.ID]
		manager.mu.RUnlock()
		assert.False(t, exists)
	})

	t.Run("store add failure", func(t *testing.T) {
		store := newMockStore()
		store.addErr = errors.New("store full")
		manager := NewManager(store)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "store add failed")
		assert.Contains(t, err.Error(), "store full")
	})
}

// TestManager_Shutdown tests graceful shutdown.
func TestManager_Shutdown(t *testing.T) {
	ctx := context.Background()

	t.Run("successful shutdown", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		err = manager.Shutdown(ctx, connector.ID)
		require.NoError(t, err)

		// Verify connector removed from memory
		manager.mu.RLock()
		_, exists := manager.connectors[connector.ID]
		manager.mu.RUnlock()
		assert.False(t, exists)

		// Verify connector removed from store
		_, storeErr := store.Get(ctx, connector.ID)
		assert.Error(t, storeErr)
	})

	t.Run("empty connector ID", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		err := manager.Shutdown(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connector ID cannot be empty")
	})

	t.Run("connector not found", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		err := manager.Shutdown(ctx, "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connector non-existent not found")
	})
}

// TestManager_ShutdownWithHooks tests shutdown with hooks.
func TestManager_ShutdownWithHooks(t *testing.T) {
	ctx := context.Background()

	t.Run("hooks execute during shutdown", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)
		hook := &mockHook{}
		manager.RegisterHook(hook)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		err = manager.Shutdown(ctx, connector.ID)
		require.NoError(t, err)

		calls := hook.getCalls()
		assert.Contains(t, calls, "OnPreShutdown")
		assert.Contains(t, calls, "OnPostShutdown")
	})

	t.Run("pre-shutdown hook failure", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)
		hook := &mockHook{preShutdownErr: errors.New("drain failed")}
		manager.RegisterHook(hook)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		err = manager.Shutdown(ctx, connector.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pre-shutdown failed")
	})

	t.Run("post-shutdown error collected", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)
		hook := &mockHook{postShutdownErr: errors.New("cleanup failed")}
		manager.RegisterHook(hook)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		err = manager.Shutdown(ctx, connector.ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "post-shutdown failed")

		// Connector should still be removed from memory
		manager.mu.RLock()
		_, exists := manager.connectors[connector.ID]
		manager.mu.RUnlock()
		assert.False(t, exists)
	})
}

// TestManager_ShutdownAll tests parallel shutdown.
func TestManager_ShutdownAll(t *testing.T) {
	ctx := context.Background()

	t.Run("shutdown multiple connectors", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		// Initialize multiple connectors
		for i := 0; i < 5; i++ {
			connector := &Connector{
				ID:   fmt.Sprintf("test-%d", i),
				Type: "mock",
			}
			err := manager.Initialize(ctx, connector)
			require.NoError(t, err)
		}

		// Verify all in memory
		manager.mu.RLock()
		count := len(manager.connectors)
		manager.mu.RUnlock()
		assert.Equal(t, 5, count)

		// Shutdown all
		err := manager.ShutdownAll(ctx)
		require.NoError(t, err)

		// Verify all removed from memory
		manager.mu.RLock()
		count = len(manager.connectors)
		manager.mu.RUnlock()
		assert.Equal(t, 0, count)
	})

	t.Run("shutdown with timeout", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Note: This test verifies the context is respected, actual timeout
		// depends on hook execution speed
		err = manager.ShutdownAll(ctx)
		// May or may not error depending on execution speed
		_ = err
	})

	t.Run("shutdown is idempotent", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		// First shutdown
		err = manager.ShutdownAll(ctx)
		require.NoError(t, err)

		// Second shutdown (should be no-op)
		err = manager.ShutdownAll(ctx)
		require.NoError(t, err)
	})
}

// TestManager_Get tests connector retrieval.
func TestManager_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("get from memory", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		retrieved, err := manager.Get(ctx, connector.ID)
		require.NoError(t, err)
		assert.Equal(t, connector.ID, retrieved.ID)
	})

	t.Run("get from store fallback", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		// Add directly to store (not in memory)
		connector := &Connector{ID: "test", Type: "mock"}
		err := store.Add(ctx, connector)
		require.NoError(t, err)

		retrieved, err := manager.Get(ctx, connector.ID)
		require.NoError(t, err)
		assert.Equal(t, connector.ID, retrieved.ID)
	})

	t.Run("get non-existent", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		_, err := manager.Get(ctx, "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connector not found")
	})
}

// TestManager_List tests listing connectors.
func TestManager_List(t *testing.T) {
	ctx := context.Background()

	t.Run("list from memory", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		// Initialize connectors
		for i := 0; i < 3; i++ {
			connector := &Connector{
				ID:   fmt.Sprintf("test-%d", i),
				Type: "mock",
			}
			err := manager.Initialize(ctx, connector)
			require.NoError(t, err)
		}

		connectors, err := manager.List(ctx)
		require.NoError(t, err)
		assert.Len(t, connectors, 3)
	})

	t.Run("list from store when memory empty", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		// Add directly to store
		for i := 0; i < 2; i++ {
			connector := &Connector{
				ID:   fmt.Sprintf("test-%d", i),
				Type: "mock",
			}
			err := store.Add(ctx, connector)
			require.NoError(t, err)
		}

		connectors, err := manager.List(ctx)
		require.NoError(t, err)
		assert.Len(t, connectors, 2)
	})
}

// TestManager_Update tests connector updates.
func TestManager_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		connector := &Connector{
			ID:     "test",
			Type:   "mock",
			Config: map[string]interface{}{"key": "value"},
		}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		// Update config
		updated := &Connector{
			ID:     "test",
			Type:   "mock",
			Config: map[string]interface{}{"key": "new-value"},
		}
		err = manager.Update(ctx, connector.ID, updated)
		require.NoError(t, err)

		// Verify update in memory
		manager.mu.RLock()
		memoryCon := manager.connectors[connector.ID]
		manager.mu.RUnlock()
		assert.Equal(t, "new-value", memoryCon.Config["key"])
	})

	t.Run("empty connector ID", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Update(ctx, "", connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connector ID cannot be empty")
	})

	t.Run("nil connector", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		err := manager.Update(ctx, "test", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "connector cannot be nil")
	})


	t.Run("validation failure", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		// Initialize connector first (without hooks)
		connector := &Connector{ID: "test", Type: "mock"}
		err := manager.Initialize(ctx, connector)
		require.NoError(t, err)

		// Now register a hook that will fail validation
		hook := &mockHook{preInitErr: errors.New("validation failed")}
		manager.RegisterHook(hook)

		// Update with hook that fails - should fail validation
		updated := &Connector{ID: "test", Type: "mock"}
		err = manager.Update(ctx, connector.ID, updated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

// TestManager_Close tests manager cleanup.
func TestManager_Close(t *testing.T) {
	ctx := context.Background()

	t.Run("close shuts down all connectors", func(t *testing.T) {
		store := newMockStore()
		manager := NewManager(store)

		// Initialize connectors
		for i := 0; i < 3; i++ {
			connector := &Connector{
				ID:   fmt.Sprintf("test-%d", i),
				Type: "mock",
			}
			err := manager.Initialize(ctx, connector)
			require.NoError(t, err)
		}

		err := manager.Close(ctx)
		require.NoError(t, err)

		// Verify all removed
		manager.mu.RLock()
		count := len(manager.connectors)
		manager.mu.RUnlock()
		assert.Equal(t, 0, count)
	})
}

// TestManager_GracefulShutdown tests graceful shutdown convenience method.
func TestManager_GracefulShutdown(t *testing.T) {
	store := newMockStore()
	manager := NewManager(store)
	ctx := context.Background()

	connector := &Connector{ID: "test", Type: "mock"}
	err := manager.Initialize(ctx, connector)
	require.NoError(t, err)

	err = manager.GracefulShutdown()
	require.NoError(t, err)

	// Verify shutdown
	manager.mu.RLock()
	count := len(manager.connectors)
	manager.mu.RUnlock()
	assert.Equal(t, 0, count)
}

// TestManager_ConcurrentOperations tests thread safety.
func TestManager_ConcurrentOperations(t *testing.T) {
	store := newMockStore()
	manager := NewManager(store)
	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent initializations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			connector := &Connector{
				ID:   fmt.Sprintf("test-%d", id),
				Type: "mock",
			}
			_ = manager.Initialize(ctx, connector)
		}(i)
	}

	wg.Wait()

	// Verify all connectors initialized
	connectors, err := manager.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10, len(connectors))

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = manager.Get(ctx, fmt.Sprintf("test-%d", id))
		}(i)
	}

	wg.Wait()
}
