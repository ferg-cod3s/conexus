package embedding

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTestProvider is a test provider implementation.
type mockTestProvider struct {
	name string
}

func (p *mockTestProvider) Name() string {
	return p.name
}

func (p *mockTestProvider) Create(config map[string]interface{}) (Embedder, error) {
	return NewMock(384), nil
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()

	require.NotNil(t, r)
	assert.Empty(t, r.List())
}

func TestRegistry_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		r := NewRegistry()
		provider := &mockTestProvider{name: "test"}

		err := r.Register(provider)

		require.NoError(t, err)
		assert.Contains(t, r.List(), "test")
	})

	t.Run("rejects nil provider", func(t *testing.T) {
		r := NewRegistry()

		err := r.Register(nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil provider")
	})

	t.Run("rejects empty name", func(t *testing.T) {
		r := NewRegistry()
		provider := &mockTestProvider{name: ""}

		err := r.Register(provider)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name cannot be empty")
	})

	t.Run("rejects duplicate registration", func(t *testing.T) {
		r := NewRegistry()
		provider1 := &mockTestProvider{name: "duplicate"}
		provider2 := &mockTestProvider{name: "duplicate"}

		err1 := r.Register(provider1)
		require.NoError(t, err1)

		err2 := r.Register(provider2)
		assert.Error(t, err2)
		assert.Contains(t, err2.Error(), "already registered")
	})

	t.Run("allows different providers", func(t *testing.T) {
		r := NewRegistry()
		provider1 := &mockTestProvider{name: "first"}
		provider2 := &mockTestProvider{name: "second"}

		err1 := r.Register(provider1)
		err2 := r.Register(provider2)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.Len(t, r.List(), 2)
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Run("retrieves registered provider", func(t *testing.T) {
		r := NewRegistry()
		provider := &mockTestProvider{name: "test"}
		err := r.Register(provider)
		require.NoError(t, err)

		retrieved, err := r.Get("test")

		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, "test", retrieved.Name())
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		r := NewRegistry()

		provider, err := r.Get("nonexistent")

		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns correct provider among multiple", func(t *testing.T) {
		r := NewRegistry()
		p1 := &mockTestProvider{name: "first"}
		p2 := &mockTestProvider{name: "second"}
		p3 := &mockTestProvider{name: "third"}

		require.NoError(t, r.Register(p1))
		require.NoError(t, r.Register(p2))
		require.NoError(t, r.Register(p3))

		retrieved, err := r.Get("second")

		require.NoError(t, err)
		assert.Equal(t, "second", retrieved.Name())
	})
}

func TestRegistry_List(t *testing.T) {
	t.Run("returns empty list for new registry", func(t *testing.T) {
		r := NewRegistry()

		list := r.List()

		assert.Empty(t, list)
	})

	t.Run("returns all registered providers", func(t *testing.T) {
		r := NewRegistry()
		require.NoError(t, r.Register(&mockTestProvider{name: "alpha"}))
		require.NoError(t, r.Register(&mockTestProvider{name: "beta"}))
		require.NoError(t, r.Register(&mockTestProvider{name: "gamma"}))

		list := r.List()

		assert.Len(t, list, 3)
		assert.Contains(t, list, "alpha")
		assert.Contains(t, list, "beta")
		assert.Contains(t, list, "gamma")
	})

	t.Run("returns sorted list", func(t *testing.T) {
		r := NewRegistry()
		require.NoError(t, r.Register(&mockTestProvider{name: "zebra"}))
		require.NoError(t, r.Register(&mockTestProvider{name: "alpha"}))
		require.NoError(t, r.Register(&mockTestProvider{name: "middle"}))

		list := r.List()

		assert.Equal(t, []string{"alpha", "middle", "zebra"}, list)
	})
}

func TestRegistry_MustRegister(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		r := NewRegistry()
		provider := &mockTestProvider{name: "test"}

		assert.NotPanics(t, func() {
			r.MustRegister(provider)
		})
	})

	t.Run("panics on error", func(t *testing.T) {
		r := NewRegistry()
		provider := &mockTestProvider{name: "duplicate"}
		require.NoError(t, r.Register(provider))

		assert.Panics(t, func() {
			r.MustRegister(provider) // Try to register duplicate
		})
	})
}

func TestRegistry_Unregister(t *testing.T) {
	t.Run("removes registered provider", func(t *testing.T) {
		r := NewRegistry()
		provider := &mockTestProvider{name: "test"}
		require.NoError(t, r.Register(provider))

		r.Unregister("test")

		assert.Empty(t, r.List())
		_, err := r.Get("test")
		assert.Error(t, err)
	})

	t.Run("no-op for non-existent provider", func(t *testing.T) {
		r := NewRegistry()

		assert.NotPanics(t, func() {
			r.Unregister("nonexistent")
		})
	})

	t.Run("only removes specified provider", func(t *testing.T) {
		r := NewRegistry()
		require.NoError(t, r.Register(&mockTestProvider{name: "first"}))
		require.NoError(t, r.Register(&mockTestProvider{name: "second"}))

		r.Unregister("first")

		assert.Len(t, r.List(), 1)
		assert.Contains(t, r.List(), "second")
	})
}

func TestRegistry_Clear(t *testing.T) {
	t.Run("removes all providers", func(t *testing.T) {
		r := NewRegistry()
		require.NoError(t, r.Register(&mockTestProvider{name: "first"}))
		require.NoError(t, r.Register(&mockTestProvider{name: "second"}))
		require.NoError(t, r.Register(&mockTestProvider{name: "third"}))

		r.Clear()

		assert.Empty(t, r.List())
	})

	t.Run("no-op on empty registry", func(t *testing.T) {
		r := NewRegistry()

		assert.NotPanics(t, func() {
			r.Clear()
		})
		assert.Empty(t, r.List())
	})
}

func TestRegistry_Concurrency(t *testing.T) {
	t.Run("concurrent registration", func(t *testing.T) {
		r := NewRegistry()
		var wg sync.WaitGroup

		// Register 100 providers concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				provider := &mockTestProvider{name: string(rune('a' + n))}
				_ = r.Register(provider)
			}(i)
		}

		wg.Wait()

		// Should have successfully registered multiple providers
		assert.NotEmpty(t, r.List())
	})

	t.Run("concurrent read/write", func(t *testing.T) {
		r := NewRegistry()
		require.NoError(t, r.Register(&mockTestProvider{name: "test"}))

		var wg sync.WaitGroup

		// 50 readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = r.Get("test")
				_ = r.List()
			}()
		}

		// 50 writers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				provider := &mockTestProvider{name: string(rune('a' + n))}
				_ = r.Register(provider)
			}(i)
		}

		wg.Wait()

		// Should not panic and test provider should still exist
		provider, err := r.Get("test")
		require.NoError(t, err)
		assert.Equal(t, "test", provider.Name())
	})
}

func TestGlobalRegistry(t *testing.T) {
	t.Run("global Register/Get/List functions work", func(t *testing.T) {
		// Create isolated test provider (name won't conflict with "mock")
		provider := &mockTestProvider{name: "test-isolated"}
		err := Register(provider)

		// May error if already registered, that's ok
		if err == nil {
			defer registry.Unregister("test-isolated")

			retrieved, err := Get("test-isolated")
			require.NoError(t, err)
			assert.Equal(t, "test-isolated", retrieved.Name())

			list := List()
			assert.Contains(t, list, "test-isolated")
		}
	})

	t.Run("mock provider registered by default", func(t *testing.T) {
		// The init() function should have registered "mock"
		list := List()
		assert.Contains(t, list, "mock")

		provider, err := Get("mock")
		require.NoError(t, err)
		assert.Equal(t, "mock", provider.Name())
	})
}

func TestRegistry_CreateFromProvider(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register(&MockProvider{}))

	provider, err := r.Get("mock")
	require.NoError(t, err)

	t.Run("creates embedder with default config", func(t *testing.T) {
		embedder, err := provider.Create(map[string]interface{}{})

		require.NoError(t, err)
		require.NotNil(t, embedder)
		assert.Equal(t, 384, embedder.Dimensions())
	})

	t.Run("creates embedder with custom config", func(t *testing.T) {
		config := map[string]interface{}{
			"dimensions": 512,
		}

		embedder, err := provider.Create(config)

		require.NoError(t, err)
		assert.Equal(t, 512, embedder.Dimensions())
	})
}

// Benchmark tests
func BenchmarkRegistry_Register(b *testing.B) {
	r := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider := &mockTestProvider{name: string(rune('a' + i%26))}
		_ = r.Register(provider)
	}
}

func BenchmarkRegistry_Get(b *testing.B) {
	r := NewRegistry()
	require.NoError(b, r.Register(&mockTestProvider{name: "test"}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Get("test")
	}
}

func BenchmarkRegistry_List(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 10; i++ {
		require.NoError(b, r.Register(&mockTestProvider{name: string(rune('a' + i))}))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.List()
	}
}

func BenchmarkRegistry_ConcurrentGet(b *testing.B) {
	r := NewRegistry()
	require.NoError(b, r.Register(&mockTestProvider{name: "test"}))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = r.Get("test")
		}
	})
}
