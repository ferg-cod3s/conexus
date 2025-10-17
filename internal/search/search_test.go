// Package search provides tests for search functionality.
package search

import (
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
)

func TestSearchCache_Basic(t *testing.T) {
	cache := NewSearchCache(10, time.Minute)

	// Test cache miss
	query := "test query"
	filters := map[string]interface{}{"source_type": "file"}
	_, found := cache.Get(query, filters)
	assert.False(t, found)

	// Test cache set and get
	results := []vectorstore.SearchResult{
		{Score: 0.9},
		{Score: 0.8},
	}
	cache.Set(query, filters, results, 0.1)

	cached, found := cache.Get(query, filters)
	assert.True(t, found)
	assert.Equal(t, query, cached.Query)
	assert.Equal(t, results, cached.Results)
	assert.Equal(t, 0.1, cached.QueryTime)
}

func TestSearchCache_Size(t *testing.T) {
	cache := NewSearchCache(2, time.Minute)

	// Add entries
	cache.Set("query1", nil, nil, 0.1)
	cache.Set("query2", nil, nil, 0.1)
	cache.Set("query3", nil, nil, 0.1) // Should evict oldest

	assert.Equal(t, 2, cache.Size())

	// Oldest should be gone
	_, found := cache.Get("query1", nil)
	assert.False(t, found)

	// Newer ones should be there
	_, found = cache.Get("query2", nil)
	assert.True(t, found)
	_, found = cache.Get("query3", nil)
	assert.True(t, found)
}

func TestSearchCache_TTL(t *testing.T) {
	cache := NewSearchCache(10, time.Millisecond*10)

	// Add entry
	cache.Set("query", nil, nil, 0.1)

	// Should be available immediately
	_, found := cache.Get("query", nil)
	assert.True(t, found)

	// Wait for TTL to expire
	time.Sleep(time.Millisecond * 15)

	// Should be gone
	_, found = cache.Get("query", nil)
	assert.False(t, found)
}

func TestSearchCache_Clear(t *testing.T) {
	cache := NewSearchCache(10, time.Minute)

	cache.Set("query1", nil, nil, 0.1)
	cache.Set("query2", nil, nil, 0.1)

	assert.Equal(t, 2, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}
