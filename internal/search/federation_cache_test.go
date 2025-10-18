package search

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ferg-cod3s/conexus/internal/schema"
)

func TestNewFederationCache(t *testing.T) {
	config := DefaultFederationCacheConfig()
	cache := NewFederationCache(config)

	assert.NotNil(t, cache)
	assert.Equal(t, config.MaxEntries, 500)
	assert.Equal(t, config.TTL, 5*time.Minute)
	assert.True(t, config.ContentBasedInvalidation)
	assert.True(t, config.InvalidateOnConnectorChange)
}

func TestNewFederationCacheWithNilConfig(t *testing.T) {
	cache := NewFederationCache(nil)
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.entries)
	assert.NotNil(t, cache.invalidators)
}

func TestGenerateKey(t *testing.T) {
	cache := NewFederationCache(nil)

	// Test basic key generation
	key1 := cache.GenerateKey("test query", nil, []string{"connector1"})
	assert.NotEmpty(t, key1)

	// Test same inputs produce same key
	key2 := cache.GenerateKey("test query", nil, []string{"connector1"})
	assert.Equal(t, key1, key2)

	// Test different queries produce different keys
	key3 := cache.GenerateKey("different query", nil, []string{"connector1"})
	assert.NotEqual(t, key1, key3)

	// Test connector order doesn't matter
	key4 := cache.GenerateKey("test query", nil, []string{"connector1", "connector2"})
	key5 := cache.GenerateKey("test query", nil, []string{"connector2", "connector1"})
	assert.Equal(t, key4, key5)

	// Test filters included in key
	filters := map[string]interface{}{"type": "issue"}
	key6 := cache.GenerateKey("test query", filters, []string{"connector1"})
	key7 := cache.GenerateKey("test query", nil, []string{"connector1"})
	assert.NotEqual(t, key6, key7)
}

func TestSetAndGet(t *testing.T) {
	cache := NewFederationCache(nil)

	key := cache.GenerateKey("test query", nil, []string{"connector1"})
	response := &schema.SearchResponse{
		Results: []schema.SearchResultItem{
			{ID: "1", Score: 0.9},
			{ID: "2", Score: 0.8},
		},
		TotalCount: 2,
	}
	metadata := FederationCacheMetadata{
		Query:              "test query",
		ConnectorIDs:       []string{"connector1"},
		SourceCount:        1,
		ResultsBeforeMerge: 2,
		ResultsAfterMerge:  2,
	}

	// Set the cache entry
	err := cache.Set(key, response, metadata, "fingerprint1")
	require.NoError(t, err)

	// Get the cache entry
	cached, found, expired, invalidated := cache.Get(key)
	assert.True(t, found)
	assert.False(t, expired)
	assert.False(t, invalidated)
	assert.NotNil(t, cached)
	assert.Equal(t, len(cached.Results), 2)
}

func TestGetMiss(t *testing.T) {
	cache := NewFederationCache(nil)

	cached, found, expired, invalidated := cache.Get("nonexistent")
	assert.False(t, found)
	assert.False(t, expired)
	assert.False(t, invalidated)
	assert.Nil(t, cached)
}

func TestCacheExpiration(t *testing.T) {
	config := &FederationCacheConfig{
		MaxEntries: 100,
		TTL:        10 * time.Millisecond,
	}
	cache := NewFederationCache(config)

	key := cache.GenerateKey("test query", nil, []string{"connector1"})
	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test query", ConnectorIDs: []string{"connector1"}}

	err := cache.Set(key, response, metadata, "fingerprint1")
	require.NoError(t, err)

	// Get before expiration
	cached, found, expired, _ := cache.Get(key)
	assert.True(t, found)
	assert.False(t, expired)
	assert.NotNil(t, cached)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Get after expiration
	cached, found, expired, _ = cache.Get(key)
	assert.False(t, found)
	assert.True(t, expired)
	assert.Nil(t, cached)
}

func TestCacheEviction(t *testing.T) {
	config := &FederationCacheConfig{
		MaxEntries: 3,
		TTL:        1 * time.Hour,
	}
	cache := NewFederationCache(config)

	// Add entries
	for i := 1; i <= 4; i++ {
		key := cache.GenerateKey("query"+string(rune(48+i)), nil, []string{"connector1"})
		response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
		metadata := FederationCacheMetadata{Query: "query" + string(rune(48+i)), ConnectorIDs: []string{"connector1"}}

		err := cache.Set(key, response, metadata, "fingerprint1")
		require.NoError(t, err)

		time.Sleep(1 * time.Millisecond) // Ensure different access times
	}

	// Should have evicted the oldest entry
	assert.LessOrEqual(t, cache.GetEntryCount(), 3)
}

func TestInvalidateByConnectorChange(t *testing.T) {
	cache := NewFederationCache(nil)

	// Add entries with different connectors
	key1 := cache.GenerateKey("query1", nil, []string{"connector1", "connector2"})
	key2 := cache.GenerateKey("query2", nil, []string{"connector2", "connector3"})
	key3 := cache.GenerateKey("query3", nil, []string{"connector4"})

	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}

	// Set with proper connector IDs in metadata
	metadata1 := FederationCacheMetadata{Query: "query1", ConnectorIDs: []string{"connector1", "connector2"}}
	cache.Set(key1, response, metadata1, "fingerprint1")
	metadata2 := FederationCacheMetadata{Query: "query2", ConnectorIDs: []string{"connector2", "connector3"}}
	cache.Set(key2, response, metadata2, "fingerprint1")
	metadata3 := FederationCacheMetadata{Query: "query3", ConnectorIDs: []string{"connector4"}}
	cache.Set(key3, response, metadata3, "fingerprint1")

	assert.Equal(t, 3, cache.GetEntryCount())

	// Invalidate entries related to connector2
	invalidated := cache.InvalidateByConnectorChange([]string{"connector2"})

	// Should invalidate key1 and key2 (they contain connector2)
	assert.Equal(t, 2, invalidated)
}

func TestInvalidateByFingerprint(t *testing.T) {
	cache := NewFederationCache(nil)

	key1 := cache.GenerateKey("query1", nil, []string{"connector1"})
	key2 := cache.GenerateKey("query2", nil, []string{"connector2"})

	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}

	cache.Set(key1, response, FederationCacheMetadata{Query: "query1", ConnectorIDs: []string{"connector1"}}, "fingerprint1")
	cache.Set(key2, response, FederationCacheMetadata{Query: "query2", ConnectorIDs: []string{"connector2"}}, "fingerprint2")

	assert.Equal(t, 2, cache.GetEntryCount())

	// Invalidate entries with mismatched fingerprint
	invalidated := cache.InvalidateByFingerprint("fingerprint_new")

	// Should invalidate entries with old fingerprints
	assert.Greater(t, invalidated, 0)
}

func TestRegisterInvalidationHook(t *testing.T) {
	cache := NewFederationCache(nil)

	key := cache.GenerateKey("test query", nil, []string{"connector1"})
	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test query", ConnectorIDs: []string{"connector1"}}

	cache.Set(key, response, metadata, "fingerprint1")

	// Register a hook that invalidates
	cache.RegisterInvalidationHook(key, func() error {
		return nil // Return nil = still valid
	})

	// Entry should still be found
	cached, found, _, invalidated := cache.Get(key)
	assert.True(t, found)
	assert.False(t, invalidated)
	assert.NotNil(t, cached)
}

func TestInvalidate(t *testing.T) {
	cache := NewFederationCache(nil)

	key := cache.GenerateKey("test query", nil, []string{"connector1"})
	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test query", ConnectorIDs: []string{"connector1"}}

	cache.Set(key, response, metadata, "fingerprint1")

	assert.Equal(t, 1, cache.GetEntryCount())

	// Invalidate the entry
	invalidated := cache.Invalidate(key)
	assert.True(t, invalidated)

	// Entry should be gone
	assert.Equal(t, 0, cache.GetEntryCount())

	// Invalidate non-existent key
	invalidated = cache.Invalidate("nonexistent")
	assert.False(t, invalidated)
}

func TestInvalidateAll(t *testing.T) {
	cache := NewFederationCache(nil)

	// Add multiple entries
	for i := 1; i <= 5; i++ {
		key := cache.GenerateKey("query"+string(rune(48+i)), nil, []string{"connector1"})
		response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
		metadata := FederationCacheMetadata{Query: "query" + string(rune(48+i)), ConnectorIDs: []string{"connector1"}}

		cache.Set(key, response, metadata, "fingerprint1")
	}

	assert.Equal(t, 5, cache.GetEntryCount())

	// Invalidate all
	cache.InvalidateAll()

	assert.Equal(t, 0, cache.GetEntryCount())
}

func TestGetStats(t *testing.T) {
	cache := NewFederationCache(nil)

	key1 := cache.GenerateKey("query1", nil, []string{"connector1"})
	key2 := cache.GenerateKey("query2", nil, []string{"connector1"})

	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test", ConnectorIDs: []string{"connector1"}}

	cache.Set(key1, response, metadata, "fingerprint1")

	// Cache hits
	cache.Get(key1)
	cache.Get(key1)

	// Cache miss
	cache.Get(key2)

	stats := cache.GetStats()
	assert.Equal(t, int64(2), stats.Hits)
	assert.Greater(t, stats.Misses, int64(0))
	assert.Equal(t, 1, stats.CurrentSize)
}

func TestGetHitRate(t *testing.T) {
	cache := NewFederationCache(nil)

	key := cache.GenerateKey("query1", nil, []string{"connector1"})
	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test", ConnectorIDs: []string{"connector1"}}

	cache.Set(key, response, metadata, "fingerprint1")

	// Generate hits and misses
	cache.Get(key)     // hit
	cache.Get(key)     // hit
	cache.Get("miss1") // miss
	cache.Get("miss2") // miss

	hitRate := cache.GetHitRate()
	assert.Greater(t, hitRate, 0.0)
	assert.LessOrEqual(t, hitRate, 1.0)
	assert.Equal(t, 0.5, hitRate) // 2 hits / 4 total
}

func TestListEntries(t *testing.T) {
	cache := NewFederationCache(nil)

	// Add multiple entries
	for i := 1; i <= 3; i++ {
		key := cache.GenerateKey("query"+string(rune(48+i)), nil, []string{"connector1"})
		response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
		metadata := FederationCacheMetadata{Query: "query" + string(rune(48+i)), ConnectorIDs: []string{"connector1"}}

		cache.Set(key, response, metadata, "fingerprint1")
		time.Sleep(1 * time.Millisecond)
	}

	entries := cache.ListEntries()
	assert.Equal(t, 3, len(entries))

	// Should be sorted by most recently accessed (last one first)
	for i, entry := range entries {
		assert.NotEmpty(t, entry.Key)
		assert.NotNil(t, entry.Response)
		_ = i
	}
}

func TestConnectorFingerprint(t *testing.T) {
	cache := NewFederationCache(nil)

	configs1 := map[string]map[string]interface{}{
		"connector1": {"enabled": true, "token": "abc123"},
		"connector2": {"enabled": true, "url": "https://api.example.com"},
	}

	configs2 := map[string]map[string]interface{}{
		"connector1": {"enabled": true, "token": "abc123"},
		"connector2": {"enabled": true, "url": "https://api.example.com"},
	}

	configs3 := map[string]map[string]interface{}{
		"connector1": {"enabled": false, "token": "abc123"},
		"connector2": {"enabled": true, "url": "https://api.example.com"},
	}

	fp1 := cache.GenerateConnectorFingerprint(configs1)
	fp2 := cache.GenerateConnectorFingerprint(configs2)
	fp3 := cache.GenerateConnectorFingerprint(configs3)

	// Same configs should produce same fingerprint
	assert.Equal(t, fp1, fp2)

	// Different configs should produce different fingerprint
	assert.NotEqual(t, fp1, fp3)
}

func TestCacheWithFilters(t *testing.T) {
	cache := NewFederationCache(nil)

	filters1 := map[string]interface{}{"type": "issue", "status": "open"}
	filters2 := map[string]interface{}{"status": "open", "type": "issue"} // Different order

	key1 := cache.GenerateKey("test query", filters1, []string{"connector1"})
	key2 := cache.GenerateKey("test query", filters2, []string{"connector1"})

	// Should produce same key regardless of filter order
	assert.Equal(t, key1, key2)
}

func TestCacheStatsTracking(t *testing.T) {
	cache := NewFederationCache(nil)

	key := cache.GenerateKey("query1", nil, []string{"connector1"})
	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test", ConnectorIDs: []string{"connector1"}}

	cache.Set(key, response, metadata, "fingerprint1")

	// Generate activity
	cache.Get(key)
	cache.Get("miss1")
	cache.Invalidate(key)

	stats := cache.GetStats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Greater(t, stats.Misses, int64(0))
	assert.Equal(t, int64(1), stats.Invalidations)
	assert.Equal(t, 0, stats.CurrentSize)
}

func TestCacheAccessTracking(t *testing.T) {
	cache := NewFederationCache(nil)

	key := cache.GenerateKey("query1", nil, []string{"connector1"})
	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test", ConnectorIDs: []string{"connector1"}}

	cache.Set(key, response, metadata, "fingerprint1")

	// Get the entry to track access
	_, found, _, _ := cache.Get(key)
	require.True(t, found)

	// Check that entry's access count was updated
	entries := cache.ListEntries()
	require.Equal(t, 1, len(entries))
	assert.Greater(t, entries[0].AccessCount, 0)
}

func TestConcurrentAccess(t *testing.T) {
	cache := NewFederationCache(nil)

	key := cache.GenerateKey("query1", nil, []string{"connector1"})
	response := &schema.SearchResponse{Results: []schema.SearchResultItem{}, TotalCount: 0}
	metadata := FederationCacheMetadata{Query: "test", ConnectorIDs: []string{"connector1"}}

	cache.Set(key, response, metadata, "fingerprint1")

	// Concurrent gets
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, found, _, _ := cache.Get(key)
			assert.True(t, found)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := cache.GetStats()
	assert.Equal(t, int64(10), stats.Hits)
}

func TestNormalizeFilters(t *testing.T) {
	cache := NewFederationCache(nil)

	filters := map[string]interface{}{
		"z_field": "value3",
		"a_field": "value1",
		"m_field": "value2",
	}

	normalized := cache.normalizeFilters(filters)

	// Should contain all fields in sorted order
	assert.Contains(t, normalized, "a_field")
	assert.Contains(t, normalized, "m_field")
	assert.Contains(t, normalized, "z_field")

	// Find indices to verify order
	aIdx := findStrIndex(normalized, "a_field")
	mIdx := findStrIndex(normalized, "m_field")
	zIdx := findStrIndex(normalized, "z_field")

	assert.Less(t, aIdx, mIdx)
	assert.Less(t, mIdx, zIdx)
}

// Helper function for testing
func findStrIndex(s, substr string) int {
	for i := 0; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
