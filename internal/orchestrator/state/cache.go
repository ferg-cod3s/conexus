package state

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Cache provides result caching for agent outputs
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	config  *CacheConfig
}

// CacheEntry represents a cached result
type CacheEntry struct {
	// Cache key
	Key string

	// Cached agent output
	Output *schema.AgentOutputV1

	// Metadata about the cached result
	Metadata CacheMetadata

	// Timestamp when entry was created
	CreatedAt time.Time

	// Timestamp when entry was last accessed
	LastAccessed time.Time

	// Number of times this entry has been accessed
	AccessCount int
}

// CacheMetadata contains metadata about a cached result
type CacheMetadata struct {
	// Agent that produced this result
	Agent string

	// Original request
	Request string

	// Content hash for invalidation
	ContentHash string

	// Tags for categorization
	Tags []string
}

// CacheConfig configures cache behavior
type CacheConfig struct {
	// Maximum number of entries to cache
	MaxEntries int

	// Time-to-live for cache entries
	TTL time.Duration

	// Enable content-based invalidation
	ContentBasedInvalidation bool
}

// NewCache creates a new cache with the given configuration
func NewCache(config *CacheConfig) *Cache {
	if config == nil {
		config = DefaultCacheConfig()
	}

	return &Cache{
		entries: make(map[string]*CacheEntry),
		config:  config,
	}
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxEntries:               1000,
		TTL:                      1 * time.Hour,
		ContentBasedInvalidation: true,
	}
}

// Get retrieves a cached result
func (c *Cache) Get(key string) (*schema.AgentOutputV1, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.CreatedAt) > c.config.TTL {
		return nil, false
	}

	// Update access statistics
	c.mu.RUnlock()
	c.mu.Lock()
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	c.mu.Unlock()
	c.mu.RLock()

	return entry.Output, true
}

// Set stores a result in the cache
func (c *Cache) Set(key string, output *schema.AgentOutputV1, metadata CacheMetadata) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.entries) >= c.config.MaxEntries {
		c.evictLRU()
	}

	entry := &CacheEntry{
		Key:          key,
		Output:       output,
		Metadata:     metadata,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		AccessCount:  0,
	}

	c.entries[key] = entry

	return nil
}

// GenerateKey generates a cache key from request parameters
func (c *Cache) GenerateKey(agent, request string, permissions schema.Permissions) string {
	hash := sha256.New()
	hash.Write([]byte(agent))
	hash.Write([]byte(request))
	hash.Write([]byte(fmt.Sprintf("%v", permissions)))

	return hex.EncodeToString(hash.Sum(nil))
}

// Invalidate removes a specific cache entry
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// InvalidateByAgent removes all cache entries for a specific agent
func (c *Cache) InvalidateByAgent(agent string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key, entry := range c.entries {
		if entry.Metadata.Agent == agent {
			delete(c.entries, key)
			count++
		}
	}

	return count
}

// InvalidateByContentHash removes entries with a specific content hash
func (c *Cache) InvalidateByContentHash(contentHash string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key, entry := range c.entries {
		if entry.Metadata.ContentHash == contentHash {
			delete(c.entries, key)
			count++
		}
	}

	return count
}

// InvalidateByTag removes all cache entries with a specific tag
func (c *Cache) InvalidateByTag(tag string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for key, entry := range c.entries {
		for _, t := range entry.Metadata.Tags {
			if t == tag {
				delete(c.entries, key)
				count++
				break
			}
		}
	}

	return count
}

// Clear removes all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// evictLRU removes the least recently used entry
func (c *Cache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.LastAccessed.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccessed
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// CleanupExpired removes expired cache entries
func (c *Cache) CleanupExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	cutoff := time.Now().Add(-c.config.TTL)

	for key, entry := range c.entries {
		if entry.CreatedAt.Before(cutoff) {
			delete(c.entries, key)
			count++
		}
	}

	return count
}

// GetStats returns cache statistics
func (c *Cache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		TotalEntries: len(c.entries),
		MaxEntries:   c.config.MaxEntries,
		TTL:          c.config.TTL,
	}

	totalAccess := 0
	for _, entry := range c.entries {
		totalAccess += entry.AccessCount
	}

	if len(c.entries) > 0 {
		stats.AverageAccessCount = float64(totalAccess) / float64(len(c.entries))
	}

	return stats
}

// CacheStats contains cache statistics
type CacheStats struct {
	TotalEntries       int
	MaxEntries         int
	TTL                time.Duration
	AverageAccessCount float64
}
