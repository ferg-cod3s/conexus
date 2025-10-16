package embedding

import (
	"fmt"
	"sort"
	"sync"
)

// registry is the default global provider registry.
var registry = NewRegistry()

// Register adds a provider to the global registry.
func Register(provider Provider) error {
	return registry.Register(provider)
}

// Get retrieves a provider from the global registry.
func Get(name string) (Provider, error) {
	return registry.Get(name)
}

// List returns all provider names from the global registry.
func List() []string {
	return registry.List()
}

// Registry is a thread-safe provider registry implementation.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry.
// Returns an error if a provider with the same name already exists.
func (r *Registry) Register(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("cannot register nil provider")
	}
	
	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}
	
	r.providers[name] = provider
	return nil
}

// Get retrieves a provider by name.
// Returns an error if the provider is not found.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	
	return provider, nil
}

// List returns all registered provider names in sorted order.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	
	sort.Strings(names)
	return names
}

// MustRegister registers a provider and panics on error.
// Useful for init() functions.
func (r *Registry) MustRegister(provider Provider) {
	if err := r.Register(provider); err != nil {
		panic(err)
	}
}

// Unregister removes a provider from the registry.
// Useful for testing.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, name)
}

// Clear removes all providers from the registry.
// Useful for testing.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = make(map[string]Provider)
}

func init() {
	// Register the mock provider by default
	if err := Register(&MockProvider{}); err != nil {
		panic(fmt.Sprintf("failed to register mock provider: %v", err))
	}
}
