// Package search provides hybrid search and reranking capabilities.
package search

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/schema"
)

// FederationCacheEntry represents a cached federation search result
type FederationCacheEntry struct {
	// Unique cache key
	Key string

	// Cached search response
	Response *schema.SearchResponse

	// Metadata about the cache entry
	Metadata FederationCacheMetadata

	// Timestamp when entry was created
	CreatedAt time.Time

	// Timestamp when entry was last accessed
	LastAccessed time.Time

	// Number of times this entry has been accessed
	AccessCount int

	// Connector configuration fingerprint for invalidation
	ConnectorFingerprint string
}

// FederationCacheMetadata contains metadata about a cached federation result
type FederationCacheMetadata struct {
	// Query text
	Query string

	// Active connector IDs involved in this search
	ConnectorIDs []string

	// Filters used
	Filters map[string]interface{}

	// Number of sources that contributed results
	SourceCount int

	// Total results before merging
	ResultsBeforeMerge int

	// Total results after merging/deduplication
	ResultsAfterMerge int

	// Deduplication ratio
	DeduplicationRatio float64

	// Tags for categorization
	Tags []string
}

// FederationCacheConfig configures federation cache behavior
type FederationCacheConfig struct {
	// Maximum number of entries to cache
	MaxEntries int

	// Time-to-live for cache entries
	TTL time.Duration

	// Enable content-based invalidation
	ContentBasedInvalidation bool

	// Invalidate cache when connector configs change
	InvalidateOnConnectorChange bool

	// Enable compression for large result sets
	EnableCompression bool
}

// FederationCache provides intelligent caching for federation search results
type FederationCache struct {
	mu           sync.RWMutex
	entries      map[string]*FederationCacheEntry
	config       *FederationCacheConfig
	stats        CacheStats
	invalidators map[string]func() error // Invalidation hooks by key
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	mu              sync.RWMutex
	Hits            int64
	Misses          int64
	Evictions       int64
	Invalidations   int64
	CurrentSize     int
	MaxSizeReached  int
	TotalCacheSaved int64 // Total time saved by cache hits (ms)
}

// NewFederationCache creates a new federation-aware cache with the given configuration
func NewFederationCache(config *FederationCacheConfig) *FederationCache {
	if config == nil {
		config = DefaultFederationCacheConfig()
	}

	return &FederationCache{
		entries:      make(map[string]*FederationCacheEntry),
		config:       config,
		invalidators: make(map[string]func() error),
		stats: CacheStats{
			Hits:           0,
			Misses:         0,
			Evictions:      0,
			Invalidations:  0,
			CurrentSize:    0,
			MaxSizeReached: 0,
			TotalCacheSaved: 0,
		},
	}
}

// DefaultFederationCacheConfig returns default federation cache configuration
func DefaultFederationCacheConfig() *FederationCacheConfig {
	return &FederationCacheConfig{
		MaxEntries:                  500,
		TTL:                         5 * time.Minute,
		ContentBasedInvalidation:    true,
		InvalidateOnConnectorChange: true,
		EnableCompression:           true,
	}
}

// Get retrieves a cached federation search result
// Returns (response, found, expired, invalidated)
func (fc *FederationCache) Get(key string) (*schema.SearchResponse, bool, bool, bool) {
	fc.mu.RLock()
	entry, ok := fc.entries[key]
	fc.mu.RUnlock()

	if !ok {
		fc.updateStats(func(s *CacheStats) {
			s.Misses++
		})
		return nil, false, false, false
	}

	// Check if entry has expired
	if time.Since(entry.CreatedAt) > fc.config.TTL {
		fc.updateStats(func(s *CacheStats) {
			s.Misses++
			s.Invalidations++
		})
		return nil, false, true, false
	}

	// Check if invalidation hook exists and entry is invalidated
	if hook, exists := fc.invalidators[key]; exists {
		if err := hook(); err == nil {
			// Entry is still valid according to invalidation hook
		} else {
			fc.updateStats(func(s *CacheStats) {
				s.Misses++
				s.Invalidations++
			})
			return nil, false, false, true
		}
	}

	// Update access statistics
	fc.mu.Lock()
	entry.LastAccessed = time.Now()
	entry.AccessCount++
	fc.mu.Unlock()

	fc.updateStats(func(s *CacheStats) {
		s.Hits++
	})

	return entry.Response, true, false, false
}

// Set stores a federation search result in the cache
func (fc *FederationCache) Set(key string, response *schema.SearchResponse, metadata FederationCacheMetadata, connectorFingerprint string) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Check if we need to evict entries
	if len(fc.entries) >= fc.config.MaxEntries {
		fc.evictLRU()
	}

	entry := &FederationCacheEntry{
		Key:                   key,
		Response:              response,
		Metadata:              metadata,
		CreatedAt:             time.Now(),
		LastAccessed:          time.Now(),
		AccessCount:           0,
		ConnectorFingerprint:  connectorFingerprint,
	}

	fc.entries[key] = entry

	fc.updateStats(func(s *CacheStats) {
		s.CurrentSize = len(fc.entries)
		if s.CurrentSize > s.MaxSizeReached {
			s.MaxSizeReached = s.CurrentSize
		}
	})

	return nil
}

// GenerateKey generates a cache key from search parameters
// Key includes query text, filters, and connector IDs for federation-specific caching
func (fc *FederationCache) GenerateKey(query string, filters map[string]interface{}, connectorIDs []string) string {
	hash := sha256.New()

	// Include query text
	hash.Write([]byte(query))

	// Include normalized filters
	if len(filters) > 0 {
		filterStr := fc.normalizeFilters(filters)
		hash.Write([]byte(filterStr))
	}

	// Include sorted connector IDs (order-independent)
	sortedConnectors := make([]string, len(connectorIDs))
	copy(sortedConnectors, connectorIDs)
	sort.Strings(sortedConnectors)
	hash.Write([]byte(strings.Join(sortedConnectors, ",")))

	return hex.EncodeToString(hash.Sum(nil))
}

// GenerateConnectorFingerprint creates a fingerprint of active connector configs
// This can be used to invalidate cache when connectors change
func (fc *FederationCache) GenerateConnectorFingerprint(connectorConfigs map[string]map[string]interface{}) string {
	hash := sha256.New()

	// Sort connector IDs for consistent ordering
	ids := make([]string, 0, len(connectorConfigs))
	for id := range connectorConfigs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	// Hash each connector's config in order
	for _, id := range ids {
		hash.Write([]byte(id))
		config := connectorConfigs[id]
		configStr := fc.normalizeFilters(config)
		hash.Write([]byte(configStr))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// InvalidateByConnectorChange invalidates all cache entries related to specific connectors
func (fc *FederationCache) InvalidateByConnectorChange(affectedConnectorIDs []string) int {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	invalidated := 0
	affectedMap := make(map[string]bool)
	for _, id := range affectedConnectorIDs {
		affectedMap[id] = true
	}

	keysToRemove := []string{}

	for key, entry := range fc.entries {
		// Check if any of this entry's connectors are in the affected list
		for _, connID := range entry.Metadata.ConnectorIDs {
			if affectedMap[connID] {
				keysToRemove = append(keysToRemove, key)
				invalidated++
				break
			}
		}
	}

	// Remove invalidated entries
	for _, key := range keysToRemove {
		delete(fc.entries, key)
		delete(fc.invalidators, key)
	}

	fc.updateStats(func(s *CacheStats) {
		s.Invalidations += int64(invalidated)
		s.CurrentSize = len(fc.entries)
	})

	return invalidated
}

// InvalidateByFingerprint invalidates entries with mismatched connector fingerprints
func (fc *FederationCache) InvalidateByFingerprint(newFingerprint string) int {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	invalidated := 0
	keysToRemove := []string{}

	for key, entry := range fc.entries {
		if entry.ConnectorFingerprint != newFingerprint && entry.ConnectorFingerprint != "" {
			keysToRemove = append(keysToRemove, key)
			invalidated++
		}
	}

	// Remove invalidated entries
	for _, key := range keysToRemove {
		delete(fc.entries, key)
		delete(fc.invalidators, key)
	}

	fc.updateStats(func(s *CacheStats) {
		s.Invalidations += int64(invalidated)
		s.CurrentSize = len(fc.entries)
	})

	return invalidated
}

// RegisterInvalidationHook registers a custom invalidation hook for a cache entry
// The hook returns nil if entry is still valid, or an error if invalidated
func (fc *FederationCache) RegisterInvalidationHook(key string, hook func() error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.invalidators[key] = hook
}

// Invalidate removes a specific cache entry
func (fc *FederationCache) Invalidate(key string) bool {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if _, exists := fc.entries[key]; exists {
		delete(fc.entries, key)
		delete(fc.invalidators, key)
		fc.updateStats(func(s *CacheStats) {
			s.Invalidations++
			s.CurrentSize = len(fc.entries)
		})
		return true
	}

	return false
}

// InvalidateAll clears the entire cache
func (fc *FederationCache) InvalidateAll() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	count := len(fc.entries)
	fc.entries = make(map[string]*FederationCacheEntry)
	fc.invalidators = make(map[string]func() error)

	fc.updateStats(func(s *CacheStats) {
		s.Invalidations += int64(count)
		s.CurrentSize = 0
	})
}

// GetStats returns current cache statistics
func (fc *FederationCache) GetStats() CacheStats {
	fc.stats.mu.RLock()
	defer fc.stats.mu.RUnlock()

	return CacheStats{
		Hits:            fc.stats.Hits,
		Misses:          fc.stats.Misses,
		Evictions:       fc.stats.Evictions,
		Invalidations:   fc.stats.Invalidations,
		CurrentSize:     fc.stats.CurrentSize,
		MaxSizeReached:  fc.stats.MaxSizeReached,
		TotalCacheSaved: fc.stats.TotalCacheSaved,
	}
}

// GetHitRate returns the cache hit rate (0-1)
func (fc *FederationCache) GetHitRate() float64 {
	fc.stats.mu.RLock()
	defer fc.stats.mu.RUnlock()

	total := fc.stats.Hits + fc.stats.Misses
	if total == 0 {
		return 0
	}

	return float64(fc.stats.Hits) / float64(total)
}

// ListEntries returns a list of all cache entries (for monitoring/debugging)
func (fc *FederationCache) ListEntries() []FederationCacheEntry {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	entries := make([]FederationCacheEntry, 0, len(fc.entries))
	for _, entry := range fc.entries {
		entries = append(entries, *entry)
	}

	// Sort by most recently accessed
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastAccessed.After(entries[j].LastAccessed)
	})

	return entries
}

// GetEntryCount returns the current number of cached entries
func (fc *FederationCache) GetEntryCount() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return len(fc.entries)
}

// evictLRU evicts the least recently used entry (must be called with lock held)
func (fc *FederationCache) evictLRU() {
	if len(fc.entries) == 0 {
		return
	}

	var lruKey string
	var lruTime time.Time

	// Find least recently used entry
	for key, entry := range fc.entries {
		if lruTime.IsZero() || entry.LastAccessed.Before(lruTime) {
			lruKey = key
			lruTime = entry.LastAccessed
		}
	}

	if lruKey != "" {
		delete(fc.entries, lruKey)
		delete(fc.invalidators, lruKey)
		fc.stats.mu.Lock()
		fc.stats.Evictions++
		fc.stats.CurrentSize = len(fc.entries)
		fc.stats.mu.Unlock()
	}
}

// normalizeFilters converts filters to a consistent string representation
func (fc *FederationCache) normalizeFilters(filters map[string]interface{}) string {
	if len(filters) == 0 {
		return ""
	}

	// Create sorted keys for consistent ordering
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build normalized string
	var result strings.Builder
	for i, k := range keys {
		if i > 0 {
			result.WriteString("|")
		}
		result.WriteString(k)
		result.WriteString("=")
		result.WriteString(fmt.Sprintf("%v", filters[k]))
	}

	return result.String()
}

// updateStats safely updates cache statistics
func (fc *FederationCache) updateStats(fn func(*CacheStats)) {
	fc.stats.mu.Lock()
	defer fc.stats.mu.Unlock()
	fn(&fc.stats)
}
