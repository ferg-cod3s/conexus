// Package connectors provides lifecycle management for connectors.
package connectors

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LifecycleHook defines the interface for connector lifecycle hooks.
// Hooks are called during initialization and shutdown to perform validation,
// health checks, and cleanup operations.
type LifecycleHook interface {
	// OnPreInit is called before a connector is initialized.
	// Use this to validate configuration, check prerequisites, etc.
	OnPreInit(ctx context.Context, connector *Connector) error

	// OnPostInit is called after a connector is initialized.
	// Use this for health checks, connectivity verification, etc.
	OnPostInit(ctx context.Context, connector *Connector) error

	// OnPreShutdown is called before a connector is shut down.
	// Use this to drain in-flight requests, notify clients, etc.
	OnPreShutdown(ctx context.Context, connector *Connector) error

	// OnPostShutdown is called after a connector is shut down.
	// Use this for cleanup, closing connections, releasing resources, etc.
	OnPostShutdown(ctx context.Context, connector *Connector) error
}

// HookRegistry manages lifecycle hooks for connectors.
// Hooks are executed in the order they are registered.
type HookRegistry struct {
	mu           sync.RWMutex
	preInit      []LifecycleHook
	postInit     []LifecycleHook
	preShutdown  []LifecycleHook
	postShutdown []LifecycleHook
}

// NewHookRegistry creates a new hook registry.
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		preInit:      make([]LifecycleHook, 0),
		postInit:     make([]LifecycleHook, 0),
		preShutdown:  make([]LifecycleHook, 0),
		postShutdown: make([]LifecycleHook, 0),
	}
}

// RegisterPreInit adds a hook to be called before initialization.
func (r *HookRegistry) RegisterPreInit(hook LifecycleHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preInit = append(r.preInit, hook)
}

// RegisterPostInit adds a hook to be called after initialization.
func (r *HookRegistry) RegisterPostInit(hook LifecycleHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.postInit = append(r.postInit, hook)
}

// RegisterPreShutdown adds a hook to be called before shutdown.
func (r *HookRegistry) RegisterPreShutdown(hook LifecycleHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preShutdown = append(r.preShutdown, hook)
}

// RegisterPostShutdown adds a hook to be called after shutdown.
func (r *HookRegistry) RegisterPostShutdown(hook LifecycleHook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.postShutdown = append(r.postShutdown, hook)
}

// ExecutePreInit runs all pre-init hooks in order.
func (r *HookRegistry) ExecutePreInit(ctx context.Context, connector *Connector) error {
	r.mu.RLock()
	hooks := r.preInit
	r.mu.RUnlock()

	for i, hook := range hooks {
		if err := hook.OnPreInit(ctx, connector); err != nil {
			return fmt.Errorf("pre-init hook %d failed: %w", i, err)
		}
	}
	return nil
}

// ExecutePostInit runs all post-init hooks in order.
func (r *HookRegistry) ExecutePostInit(ctx context.Context, connector *Connector) error {
	r.mu.RLock()
	hooks := r.postInit
	r.mu.RUnlock()

	for i, hook := range hooks {
		if err := hook.OnPostInit(ctx, connector); err != nil {
			return fmt.Errorf("post-init hook %d failed: %w", i, err)
		}
	}
	return nil
}

// ExecutePreShutdown runs all pre-shutdown hooks in order.
func (r *HookRegistry) ExecutePreShutdown(ctx context.Context, connector *Connector) error {
	r.mu.RLock()
	hooks := r.preShutdown
	r.mu.RUnlock()

	for i, hook := range hooks {
		if err := hook.OnPreShutdown(ctx, connector); err != nil {
			return fmt.Errorf("pre-shutdown hook %d failed: %w", i, err)
		}
	}
	return nil
}

// ExecutePostShutdown runs all post-shutdown hooks in order.
// Unlike other hooks, this collects all errors and returns a combined error.
func (r *HookRegistry) ExecutePostShutdown(ctx context.Context, connector *Connector) error {
	r.mu.RLock()
	hooks := r.postShutdown
	r.mu.RUnlock()

	var errs []error
	for i, hook := range hooks {
		if err := hook.OnPostShutdown(ctx, connector); err != nil {
			errs = append(errs, fmt.Errorf("post-shutdown hook %d: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("post-shutdown errors: %v", errs)
	}
	return nil
}

// Clear removes all registered hooks. Useful for testing.
func (r *HookRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.preInit = r.preInit[:0]
	r.postInit = r.postInit[:0]
	r.preShutdown = r.preShutdown[:0]
	r.postShutdown = r.postShutdown[:0]
}

// HealthCheckHook performs health checks on connectors.
type HealthCheckHook struct {
	Timeout time.Duration
}

// NewHealthCheckHook creates a health check hook with the given timeout.
func NewHealthCheckHook(timeout time.Duration) *HealthCheckHook {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HealthCheckHook{Timeout: timeout}
}

// OnPreInit validates connector configuration.
func (h *HealthCheckHook) OnPreInit(ctx context.Context, connector *Connector) error {
	if connector == nil {
		return fmt.Errorf("connector is nil")
	}
	if connector.ID == "" {
		return fmt.Errorf("connector ID is required")
	}
	if connector.Type == "" {
		return fmt.Errorf("connector type is required")
	}
	return nil
}

// OnPostInit performs health check with timeout.
func (h *HealthCheckHook) OnPostInit(ctx context.Context, connector *Connector) error {
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()

	// Simulate health check - in a real implementation, this would test connectivity
	select {
	case <-ctx.Done():
		return fmt.Errorf("health check timeout after %v", h.Timeout)
	case <-time.After(10 * time.Millisecond):
		// Health check passed
		return nil
	}
}

// OnPreShutdown is a no-op for health checks.
func (h *HealthCheckHook) OnPreShutdown(ctx context.Context, connector *Connector) error {
	return nil
}

// OnPostShutdown is a no-op for health checks.
func (h *HealthCheckHook) OnPostShutdown(ctx context.Context, connector *Connector) error {
	return nil
}

// ValidationHook validates connector configuration during initialization.
type ValidationHook struct {
	RequiredConfigKeys []string
}

// NewValidationHook creates a validation hook with required config keys.
func NewValidationHook(requiredKeys ...string) *ValidationHook {
	return &ValidationHook{RequiredConfigKeys: requiredKeys}
}

// OnPreInit validates required configuration keys.
func (v *ValidationHook) OnPreInit(ctx context.Context, connector *Connector) error {
	if connector == nil {
		return fmt.Errorf("connector is nil")
	}
	if connector.Config == nil {
		return fmt.Errorf("connector config is nil")
	}

	for _, key := range v.RequiredConfigKeys {
		if _, exists := connector.Config[key]; !exists {
			return fmt.Errorf("required config key missing: %s", key)
		}
	}
	return nil
}

// OnPostInit is a no-op for validation.
func (v *ValidationHook) OnPostInit(ctx context.Context, connector *Connector) error {
	return nil
}

// OnPreShutdown is a no-op for validation.
func (v *ValidationHook) OnPreShutdown(ctx context.Context, connector *Connector) error {
	return nil
}

// OnPostShutdown is a no-op for validation.
func (v *ValidationHook) OnPostShutdown(ctx context.Context, connector *Connector) error {
	return nil
}
