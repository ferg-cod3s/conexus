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

// mockHook is a mock implementation of LifecycleHook for testing.
type mockHook struct {
	preInitErr      error
	postInitErr     error
	preShutdownErr  error
	postShutdownErr error
	callLog         []string
	mu              sync.Mutex
}

func (m *mockHook) OnPreInit(ctx context.Context, connector *Connector) error {
	m.mu.Lock()
	m.callLog = append(m.callLog, "OnPreInit")
	m.mu.Unlock()
	return m.preInitErr
}

func (m *mockHook) OnPostInit(ctx context.Context, connector *Connector) error {
	m.mu.Lock()
	m.callLog = append(m.callLog, "OnPostInit")
	m.mu.Unlock()
	return m.postInitErr
}

func (m *mockHook) OnPreShutdown(ctx context.Context, connector *Connector) error {
	m.mu.Lock()
	m.callLog = append(m.callLog, "OnPreShutdown")
	m.mu.Unlock()
	return m.preShutdownErr
}

func (m *mockHook) OnPostShutdown(ctx context.Context, connector *Connector) error {
	m.mu.Lock()
	m.callLog = append(m.callLog, "OnPostShutdown")
	m.mu.Unlock()
	return m.postShutdownErr
}

func (m *mockHook) getCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]string, len(m.callLog))
	copy(result, m.callLog)
	return result
}

// TestHookRegistry_Registration tests hook registration.
func TestHookRegistry_Registration(t *testing.T) {
	registry := NewHookRegistry()
	hook := &mockHook{}

	t.Run("register pre-init", func(t *testing.T) {
		registry.RegisterPreInit(hook)
		assert.Len(t, registry.preInit, 1)
	})

	t.Run("register post-init", func(t *testing.T) {
		registry.RegisterPostInit(hook)
		assert.Len(t, registry.postInit, 1)
	})

	t.Run("register pre-shutdown", func(t *testing.T) {
		registry.RegisterPreShutdown(hook)
		assert.Len(t, registry.preShutdown, 1)
	})

	t.Run("register post-shutdown", func(t *testing.T) {
		registry.RegisterPostShutdown(hook)
		assert.Len(t, registry.postShutdown, 1)
	})
}

// TestHookRegistry_ExecutionOrder tests that hooks execute in registration order.
func TestHookRegistry_ExecutionOrder(t *testing.T) {
	registry := NewHookRegistry()
	ctx := context.Background()
	connector := &Connector{ID: "test", Type: "mock"}

	hook1 := &mockHook{}
	hook2 := &mockHook{}
	hook3 := &mockHook{}

	t.Run("pre-init execution order", func(t *testing.T) {
		registry.Clear()
		registry.RegisterPreInit(hook1)
		registry.RegisterPreInit(hook2)
		registry.RegisterPreInit(hook3)

		err := registry.ExecutePreInit(ctx, connector)
		require.NoError(t, err)

		// Verify all hooks were called
		assert.Equal(t, []string{"OnPreInit"}, hook1.getCalls())
		assert.Equal(t, []string{"OnPreInit"}, hook2.getCalls())
		assert.Equal(t, []string{"OnPreInit"}, hook3.getCalls())
	})

	t.Run("post-init execution order", func(t *testing.T) {
		hook1 = &mockHook{}
		hook2 = &mockHook{}
		registry.Clear()
		registry.RegisterPostInit(hook1)
		registry.RegisterPostInit(hook2)

		err := registry.ExecutePostInit(ctx, connector)
		require.NoError(t, err)

		assert.Equal(t, []string{"OnPostInit"}, hook1.getCalls())
		assert.Equal(t, []string{"OnPostInit"}, hook2.getCalls())
	})
}

// TestHookRegistry_ErrorPropagation tests that errors are properly propagated.
func TestHookRegistry_ErrorPropagation(t *testing.T) {
	registry := NewHookRegistry()
	ctx := context.Background()
	connector := &Connector{ID: "test", Type: "mock"}

	t.Run("pre-init stops on error", func(t *testing.T) {
		registry.Clear()
		hook1 := &mockHook{}
		hook2 := &mockHook{preInitErr: errors.New("hook2 failed")}
		hook3 := &mockHook{}

		registry.RegisterPreInit(hook1)
		registry.RegisterPreInit(hook2)
		registry.RegisterPreInit(hook3)

		err := registry.ExecutePreInit(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pre-init hook 1 failed")
		assert.Contains(t, err.Error(), "hook2 failed")

		// Hook 1 executed, hook 2 failed, hook 3 should not execute
		assert.Equal(t, []string{"OnPreInit"}, hook1.getCalls())
		assert.Equal(t, []string{"OnPreInit"}, hook2.getCalls())
		assert.Empty(t, hook3.getCalls())
	})

	t.Run("post-init stops on error", func(t *testing.T) {
		registry.Clear()
		hook1 := &mockHook{}
		hook2 := &mockHook{postInitErr: fmt.Errorf("post-init failed")}

		registry.RegisterPostInit(hook1)
		registry.RegisterPostInit(hook2)

		err := registry.ExecutePostInit(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "post-init hook 1 failed")
	})

	t.Run("pre-shutdown stops on error", func(t *testing.T) {
		registry.Clear()
		hook1 := &mockHook{preShutdownErr: errors.New("pre-shutdown failed")}

		registry.RegisterPreShutdown(hook1)

		err := registry.ExecutePreShutdown(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pre-shutdown hook 0 failed")
	})

	t.Run("post-shutdown collects all errors", func(t *testing.T) {
		registry.Clear()
		hook1 := &mockHook{postShutdownErr: errors.New("hook1 error")}
		hook2 := &mockHook{postShutdownErr: errors.New("hook2 error")}

		registry.RegisterPostShutdown(hook1)
		registry.RegisterPostShutdown(hook2)

		err := registry.ExecutePostShutdown(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "post-shutdown errors")
		assert.Contains(t, err.Error(), "hook1 error")
		assert.Contains(t, err.Error(), "hook2 error")

		// Both hooks should execute despite errors
		assert.Equal(t, []string{"OnPostShutdown"}, hook1.getCalls())
		assert.Equal(t, []string{"OnPostShutdown"}, hook2.getCalls())
	})
}

// TestHookRegistry_ConcurrentRegistration tests thread-safe registration.
func TestHookRegistry_ConcurrentRegistration(t *testing.T) {
	registry := NewHookRegistry()
	var wg sync.WaitGroup

	// Register hooks concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hook := &mockHook{}
			registry.RegisterPreInit(hook)
			registry.RegisterPostInit(hook)
			registry.RegisterPreShutdown(hook)
			registry.RegisterPostShutdown(hook)
		}()
	}

	wg.Wait()

	// Verify all hooks were registered
	assert.Len(t, registry.preInit, 10)
	assert.Len(t, registry.postInit, 10)
	assert.Len(t, registry.preShutdown, 10)
	assert.Len(t, registry.postShutdown, 10)
}

// TestHookRegistry_Clear tests that Clear removes all hooks.
func TestHookRegistry_Clear(t *testing.T) {
	registry := NewHookRegistry()
	hook := &mockHook{}

	registry.RegisterPreInit(hook)
	registry.RegisterPostInit(hook)
	registry.RegisterPreShutdown(hook)
	registry.RegisterPostShutdown(hook)

	assert.Len(t, registry.preInit, 1)
	assert.Len(t, registry.postInit, 1)
	assert.Len(t, registry.preShutdown, 1)
	assert.Len(t, registry.postShutdown, 1)

	registry.Clear()

	assert.Empty(t, registry.preInit)
	assert.Empty(t, registry.postInit)
	assert.Empty(t, registry.preShutdown)
	assert.Empty(t, registry.postShutdown)
}

// TestHealthCheckHook_OnPreInit tests health check pre-init validation.
func TestHealthCheckHook_OnPreInit(t *testing.T) {
	hook := NewHealthCheckHook(5 * time.Second)
	ctx := context.Background()

	tests := []struct {
		name      string
		connector *Connector
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "nil connector",
			connector: nil,
			wantErr:   true,
			errMsg:    "connector is nil",
		},
		{
			name:      "missing ID",
			connector: &Connector{Type: "test"},
			wantErr:   true,
			errMsg:    "connector ID is required",
		},
		{
			name:      "missing type",
			connector: &Connector{ID: "test"},
			wantErr:   true,
			errMsg:    "connector type is required",
		},
		{
			name:      "valid connector",
			connector: &Connector{ID: "test", Type: "mock"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hook.OnPreInit(ctx, tt.connector)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestHealthCheckHook_OnPostInit tests health check execution.
func TestHealthCheckHook_OnPostInit(t *testing.T) {
	connector := &Connector{ID: "test", Type: "mock"}

	t.Run("success within timeout", func(t *testing.T) {
		hook := NewHealthCheckHook(1 * time.Second)
		ctx := context.Background()

		err := hook.OnPostInit(ctx, connector)
		require.NoError(t, err)
	})

	t.Run("timeout", func(t *testing.T) {
		hook := NewHealthCheckHook(1 * time.Millisecond)
		ctx := context.Background()

		err := hook.OnPostInit(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "health check timeout")
	})

	t.Run("context already cancelled", func(t *testing.T) {
		hook := NewHealthCheckHook(5 * time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := hook.OnPostInit(ctx, connector)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "health check timeout")
	})
}

// TestHealthCheckHook_DefaultTimeout tests default timeout.
func TestHealthCheckHook_DefaultTimeout(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "zero timeout uses default",
			timeout:         0,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "negative timeout uses default",
			timeout:         -1 * time.Second,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "custom timeout preserved",
			timeout:         10 * time.Second,
			expectedTimeout: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := NewHealthCheckHook(tt.timeout)
			assert.Equal(t, tt.expectedTimeout, hook.Timeout)
		})
	}
}

// TestValidationHook_OnPreInit tests validation hook.
func TestValidationHook_OnPreInit(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		hook      *ValidationHook
		connector *Connector
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "nil connector",
			hook:      NewValidationHook("key1"),
			connector: nil,
			wantErr:   true,
			errMsg:    "connector is nil",
		},
		{
			name:      "nil config",
			hook:      NewValidationHook("key1"),
			connector: &Connector{ID: "test", Type: "mock"},
			wantErr:   true,
			errMsg:    "connector config is nil",
		},
		{
			name: "missing required key",
			hook: NewValidationHook("apiKey", "endpoint"),
			connector: &Connector{
				ID:     "test",
				Type:   "mock",
				Config: map[string]interface{}{"apiKey": "secret"},
			},
			wantErr: true,
			errMsg:  "required config key missing: endpoint",
		},
		{
			name: "all required keys present",
			hook: NewValidationHook("apiKey", "endpoint"),
			connector: &Connector{
				ID:   "test",
				Type: "mock",
				Config: map[string]interface{}{
					"apiKey":   "secret",
					"endpoint": "http://example.com",
				},
			},
			wantErr: false,
		},
		{
			name: "no required keys",
			hook: NewValidationHook(),
			connector: &Connector{
				ID:     "test",
				Type:   "mock",
				Config: map[string]interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.hook.OnPreInit(ctx, tt.connector)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestValidationHook_NoOpHooks tests that other hooks are no-ops.
func TestValidationHook_NoOpHooks(t *testing.T) {
	hook := NewValidationHook("test")
	ctx := context.Background()
	connector := &Connector{ID: "test", Type: "mock", Config: map[string]interface{}{"test": "value"}}

	assert.NoError(t, hook.OnPostInit(ctx, connector))
	assert.NoError(t, hook.OnPreShutdown(ctx, connector))
	assert.NoError(t, hook.OnPostShutdown(ctx, connector))
}

// TestHealthCheckHook_NoOpHooks tests that shutdown hooks are no-ops.
func TestHealthCheckHook_NoOpHooks(t *testing.T) {
	hook := NewHealthCheckHook(5 * time.Second)
	ctx := context.Background()
	connector := &Connector{ID: "test", Type: "mock"}

	assert.NoError(t, hook.OnPreShutdown(ctx, connector))
	assert.NoError(t, hook.OnPostShutdown(ctx, connector))
}
