package state

import (
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache(nil)

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
		Overview:      "test output",
	}

	metadata := CacheMetadata{
		Agent:   "test-agent",
		Request: "test request",
	}

	err := cache.Set("key1", output, metadata)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	retrieved, found := cache.Get("key1")
	if !found {
		t.Error("expected to find cached entry")
		return
	}

	if retrieved.ComponentName != "test" {
		t.Errorf("expected component name 'test', got %s", retrieved.ComponentName)
	}
}

func TestCache_GetNonexistent(t *testing.T) {
	cache := NewCache(nil)

	_, found := cache.Get("nonexistent")
	if found {
		t.Error("expected not to find nonexistent entry")
	}
}

func TestCache_GetExpired(t *testing.T) {
	config := &CacheConfig{
		MaxEntries: 100,
		TTL:        1 * time.Millisecond,
	}
	cache := NewCache(config)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	metadata := CacheMetadata{
		Agent:   "test-agent",
		Request: "test",
	}

	cache.Set("key1", output, metadata)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	_, found := cache.Get("key1")
	if found {
		t.Error("expected not to find expired entry")
	}
}

func TestCache_GenerateKey(t *testing.T) {
	cache := NewCache(nil)

	key1 := cache.GenerateKey("agent1", "request1", schema.Permissions{})
	key2 := cache.GenerateKey("agent1", "request1", schema.Permissions{})
	key3 := cache.GenerateKey("agent1", "request2", schema.Permissions{})

	if key1 != key2 {
		t.Error("expected same key for same inputs")
	}

	if key1 == key3 {
		t.Error("expected different key for different inputs")
	}
}

func TestCache_Invalidate(t *testing.T) {
	cache := NewCache(nil)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	metadata := CacheMetadata{
		Agent:   "test-agent",
		Request: "test",
	}

	cache.Set("key1", output, metadata)

	cache.Invalidate("key1")

	_, found := cache.Get("key1")
	if found {
		t.Error("expected not to find invalidated entry")
	}
}

func TestCache_InvalidateByAgent(t *testing.T) {
	cache := NewCache(nil)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	cache.Set("key1", output, CacheMetadata{Agent: "agent1", Request: "test1"})
	cache.Set("key2", output, CacheMetadata{Agent: "agent1", Request: "test2"})
	cache.Set("key3", output, CacheMetadata{Agent: "agent2", Request: "test3"})

	count := cache.InvalidateByAgent("agent1")

	if count != 2 {
		t.Errorf("expected 2 entries invalidated, got %d", count)
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("expected key1 to be invalidated")
	}

	_, found = cache.Get("key3")
	if !found {
		t.Error("expected key3 to still exist")
	}
}

func TestCache_InvalidateByContentHash(t *testing.T) {
	cache := NewCache(nil)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	cache.Set("key1", output, CacheMetadata{ContentHash: "hash1"})
	cache.Set("key2", output, CacheMetadata{ContentHash: "hash1"})
	cache.Set("key3", output, CacheMetadata{ContentHash: "hash2"})

	count := cache.InvalidateByContentHash("hash1")

	if count != 2 {
		t.Errorf("expected 2 entries invalidated, got %d", count)
	}
}

func TestCache_InvalidateByTag(t *testing.T) {
	cache := NewCache(nil)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	cache.Set("key1", output, CacheMetadata{Tags: []string{"tag1", "tag2"}})
	cache.Set("key2", output, CacheMetadata{Tags: []string{"tag1"}})
	cache.Set("key3", output, CacheMetadata{Tags: []string{"tag3"}})

	count := cache.InvalidateByTag("tag1")

	if count != 2 {
		t.Errorf("expected 2 entries invalidated, got %d", count)
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache(nil)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	metadata := CacheMetadata{
		Agent:   "test-agent",
		Request: "test",
	}

	cache.Set("key1", output, metadata)
	cache.Set("key2", output, metadata)

	cache.Clear()

	stats := cache.GetStats()
	if stats.TotalEntries != 0 {
		t.Errorf("expected 0 entries after clear, got %d", stats.TotalEntries)
	}
}

func TestCache_EvictLRU(t *testing.T) {
	config := &CacheConfig{
		MaxEntries: 2,
		TTL:        1 * time.Hour,
	}
	cache := NewCache(config)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	metadata := CacheMetadata{
		Agent:   "test-agent",
		Request: "test",
	}

	// Fill cache to capacity
	cache.Set("key1", output, metadata)
	cache.Set("key2", output, metadata)

	// Access key1 to make it more recently used
	cache.Get("key1")

	// Add new entry, should evict key2 (LRU)
	cache.Set("key3", output, metadata)

	_, found1 := cache.Get("key1")
	if !found1 {
		t.Error("expected key1 to still exist")
	}

	_, found2 := cache.Get("key2")
	if found2 {
		t.Error("expected key2 to be evicted")
	}

	_, found3 := cache.Get("key3")
	if !found3 {
		t.Error("expected key3 to exist")
	}
}

func TestCache_CleanupExpired(t *testing.T) {
	config := &CacheConfig{
		MaxEntries: 100,
		TTL:        1 * time.Millisecond,
	}
	cache := NewCache(config)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	metadata := CacheMetadata{
		Agent:   "test-agent",
		Request: "test",
	}

	cache.Set("key1", output, metadata)
	cache.Set("key2", output, metadata)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	count := cache.CleanupExpired()

	if count != 2 {
		t.Errorf("expected 2 entries cleaned up, got %d", count)
	}

	stats := cache.GetStats()
	if stats.TotalEntries != 0 {
		t.Errorf("expected 0 entries after cleanup, got %d", stats.TotalEntries)
	}
}

func TestCache_GetStats(t *testing.T) {
	cache := NewCache(nil)

	output := &schema.AgentOutputV1{
		Version: "AGENT_OUTPUT_V1",
	}

	metadata := CacheMetadata{
		Agent:   "test-agent",
		Request: "test",
	}

	cache.Set("key1", output, metadata)
	cache.Set("key2", output, metadata)

	// Access entries to increment access count
	cache.Get("key1")
	cache.Get("key1")
	cache.Get("key2")

	stats := cache.GetStats()

	if stats.TotalEntries != 2 {
		t.Errorf("expected 2 entries, got %d", stats.TotalEntries)
	}

	// Average access count should be (2+1)/2 = 1.5
	if stats.AverageAccessCount < 1.0 || stats.AverageAccessCount > 2.0 {
		t.Errorf("expected average access count around 1.5, got %f", stats.AverageAccessCount)
	}
}
