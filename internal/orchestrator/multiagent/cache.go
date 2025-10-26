package multiagent

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AdvancedCache provides advanced caching for multi-agent coordination
type AdvancedCache struct {
	agentResults      map[string]*CachedAgentResult
	coordinationPlans map[string]*CachedCoordinationPlan
	performanceData   map[string]*CachedPerformanceData
	mu                sync.RWMutex
	maxSize           int
	ttl               time.Duration
	hitCount          int64
	missCount         int64
	lastCleanup       time.Time
}

// CachedAgentResult represents a cached agent result
type CachedAgentResult struct {
	AgentID     string                 `json:"agent_id"`
	Query       string                 `json:"query"`
	Result      *AgentResult           `json:"result"`
	Context     map[string]interface{} `json:"context"`
	ProfileID   string                 `json:"profile_id"`
	CachedAt    time.Time              `json:"cached_at"`
	AccessCount int                    `json:"access_count"`
	LastAccess  time.Time              `json:"last_access"`
	TTL         time.Duration          `json:"ttl"`
}

// CachedCoordinationPlan represents a cached coordination plan
type CachedCoordinationPlan struct {
	PlanID        string                 `json:"plan_id"`
	TaskQuery     string                 `json:"task_query"`
	AgentSequence []string               `json:"agent_sequence"`
	Performance   time.Duration          `json:"performance"`
	SuccessRate   float64                `json:"success_rate"`
	Context       map[string]interface{} `json:"context"`
	CachedAt      time.Time              `json:"cached_at"`
	AccessCount   int                    `json:"access_count"`
	LastAccess    time.Time              `json:"last_access"`
	TTL           time.Duration          `json:"ttl"`
}

// CachedPerformanceData represents cached performance data
type CachedPerformanceData struct {
	AgentID        string        `json:"agent_id"`
	Capability     string        `json:"capability"`
	AverageLatency time.Duration `json:"average_latency"`
	SuccessRate    float64       `json:"success_rate"`
	Load           int           `json:"load"`
	CachedAt       time.Time     `json:"cached_at"`
	TTL            time.Duration `json:"ttl"`
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	HitRate       float64            `json:"hit_rate"`
	TotalHits     int64              `json:"total_hits"`
	TotalMisses   int64              `json:"total_misses"`
	CacheSize     int                `json:"cache_size"`
	HitRateByType map[string]float64 `json:"hit_rate_by_type"`
	LastCleanup   time.Time          `json:"last_cleanup"`
}

// NewAdvancedCache creates a new advanced cache
func NewAdvancedCache(maxSize int, ttl time.Duration) *AdvancedCache {
	return &AdvancedCache{
		agentResults:      make(map[string]*CachedAgentResult),
		coordinationPlans: make(map[string]*CachedCoordinationPlan),
		performanceData:   make(map[string]*CachedPerformanceData),
		maxSize:           maxSize,
		ttl:               ttl,
		lastCleanup:       time.Now(),
	}
}

// GetAgentResult gets a cached agent result
func (ac *AdvancedCache) GetAgentResult(ctx context.Context, agentID, query string, context map[string]interface{}) (*AgentResult, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	key := ac.generateAgentResultKey(agentID, query, context)

	cached, exists := ac.agentResults[key]
	if !exists {
		ac.missCount++
		return nil, false
	}

	// Check if expired
	if time.Since(cached.CachedAt) > cached.TTL {
		ac.missCount++
		return nil, false
	}

	// Update access statistics
	cached.AccessCount++
	cached.LastAccess = time.Now()

	ac.hitCount++
	return cached.Result, true
}

// SetAgentResult sets a cached agent result
func (ac *AdvancedCache) SetAgentResult(ctx context.Context, agentID, query string, context map[string]interface{}, result *AgentResult, profileID string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	key := ac.generateAgentResultKey(agentID, query, context)

	cached := &CachedAgentResult{
		AgentID:     agentID,
		Query:       query,
		Result:      result,
		Context:     context,
		ProfileID:   profileID,
		CachedAt:    time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		TTL:         ac.ttl,
	}

	ac.agentResults[key] = cached

	// Cleanup if cache is too large
	ac.cleanupIfNeeded()
}

// GetCoordinationPlan gets a cached coordination plan
func (ac *AdvancedCache) GetCoordinationPlan(ctx context.Context, query string, context map[string]interface{}) (*CachedCoordinationPlan, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	key := ac.generateCoordinationPlanKey(query, context)

	cached, exists := ac.coordinationPlans[key]
	if !exists {
		ac.missCount++
		return nil, false
	}

	// Check if expired
	if time.Since(cached.CachedAt) > cached.TTL {
		ac.missCount++
		return nil, false
	}

	// Update access statistics
	cached.AccessCount++
	cached.LastAccess = time.Now()

	ac.hitCount++
	return cached, true
}

// SetCoordinationPlan sets a cached coordination plan
func (ac *AdvancedCache) SetCoordinationPlan(ctx context.Context, query string, context map[string]interface{}, plan *CachedCoordinationPlan) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	key := ac.generateCoordinationPlanKey(query, context)

	plan.CachedAt = time.Now()
	plan.LastAccess = time.Now()
	plan.TTL = ac.ttl

	ac.coordinationPlans[key] = plan

	// Cleanup if cache is too large
	ac.cleanupIfNeeded()
}

// GetPerformanceData gets cached performance data
func (ac *AdvancedCache) GetPerformanceData(ctx context.Context, agentID, capability string) (*CachedPerformanceData, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	key := ac.generatePerformanceDataKey(agentID, capability)

	cached, exists := ac.performanceData[key]
	if !exists {
		ac.missCount++
		return nil, false
	}

	// Check if expired
	if time.Since(cached.CachedAt) > cached.TTL {
		ac.missCount++
		return nil, false
	}

	ac.hitCount++
	return cached, true
}

// SetPerformanceData sets cached performance data
func (ac *AdvancedCache) SetPerformanceData(ctx context.Context, agentID, capability string, data *CachedPerformanceData) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	key := ac.generatePerformanceDataKey(agentID, capability)

	data.CachedAt = time.Now()
	data.TTL = ac.ttl

	ac.performanceData[key] = data

	// Cleanup if cache is too large
	ac.cleanupIfNeeded()
}

// GetStats returns cache statistics
func (ac *AdvancedCache) GetStats() *CacheStats {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	totalRequests := ac.hitCount + ac.missCount
	hitRate := float64(0)
	if totalRequests > 0 {
		hitRate = float64(ac.hitCount) / float64(totalRequests)
	}

	// Calculate hit rate by type
	hitRateByType := map[string]float64{
		"agent_results":      ac.calculateTypeHitRate(len(ac.agentResults)),
		"coordination_plans": ac.calculateTypeHitRate(len(ac.coordinationPlans)),
		"performance_data":   ac.calculateTypeHitRate(len(ac.performanceData)),
	}

	return &CacheStats{
		HitRate:       hitRate,
		TotalHits:     ac.hitCount,
		TotalMisses:   ac.missCount,
		CacheSize:     len(ac.agentResults) + len(ac.coordinationPlans) + len(ac.performanceData),
		HitRateByType: hitRateByType,
		LastCleanup:   ac.lastCleanup,
	}
}

// InvalidateAgent invalidates all cache entries for an agent
func (ac *AdvancedCache) InvalidateAgent(ctx context.Context, agentID string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Remove agent results
	for key, cached := range ac.agentResults {
		if cached.AgentID == agentID {
			delete(ac.agentResults, key)
		}
	}

	// Remove performance data
	for key, cached := range ac.performanceData {
		if cached.AgentID == agentID {
			delete(ac.performanceData, key)
		}
	}

	// Remove coordination plans that include this agent
	for key, plan := range ac.coordinationPlans {
		for _, agent := range plan.AgentSequence {
			if agent == agentID {
				delete(ac.coordinationPlans, key)
				break
			}
		}
	}
}

// InvalidateProfile invalidates cache entries for a profile
func (ac *AdvancedCache) InvalidateProfile(ctx context.Context, profileID string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	// Remove agent results for this profile
	for key, cached := range ac.agentResults {
		if cached.ProfileID == profileID {
			delete(ac.agentResults, key)
		}
	}
}

// Clear clears all cache entries
func (ac *AdvancedCache) Clear(ctx context.Context) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.agentResults = make(map[string]*CachedAgentResult)
	ac.coordinationPlans = make(map[string]*CachedCoordinationPlan)
	ac.performanceData = make(map[string]*CachedPerformanceData)
	ac.hitCount = 0
	ac.missCount = 0
	ac.lastCleanup = time.Now()
}

// generateAgentResultKey generates a cache key for agent results
func (ac *AdvancedCache) generateAgentResultKey(agentID, query string, context map[string]interface{}) string {
	// Create a hash of the query and relevant context
	data := struct {
		AgentID string                 `json:"agent_id"`
		Query   string                 `json:"query"`
		Context map[string]interface{} `json:"context"`
	}{
		AgentID: agentID,
		Query:   query,
		Context: context,
	}

	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("agent_result:%x", hash)
}

// generateCoordinationPlanKey generates a cache key for coordination plans
func (ac *AdvancedCache) generateCoordinationPlanKey(query string, context map[string]interface{}) string {
	// Create a hash of the query and relevant context
	data := struct {
		Query   string                 `json:"query"`
		Context map[string]interface{} `json:"context"`
	}{
		Query:   query,
		Context: context,
	}

	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("coordination_plan:%x", hash)
}

// generatePerformanceDataKey generates a cache key for performance data
func (ac *AdvancedCache) generatePerformanceDataKey(agentID, capability string) string {
	data := struct {
		AgentID    string `json:"agent_id"`
		Capability string `json:"capability"`
	}{
		AgentID:    agentID,
		Capability: capability,
	}

	jsonData, _ := json.Marshal(data)
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("performance_data:%x", hash)
}

// cleanupIfNeeded performs cache cleanup if needed
func (ac *AdvancedCache) cleanupIfNeeded() {
	totalSize := len(ac.agentResults) + len(ac.coordinationPlans) + len(ac.performanceData)

	if totalSize > ac.maxSize {
		ac.performCleanup()
	}
}

// performCleanup removes expired and least recently used entries
func (ac *AdvancedCache) performCleanup() {
	cutoff := time.Now().Add(-ac.ttl)

	// Remove expired entries
	for key, cached := range ac.agentResults {
		if cached.CachedAt.Before(cutoff) || cached.AccessCount == 0 {
			delete(ac.agentResults, key)
		}
	}

	for key, cached := range ac.coordinationPlans {
		if cached.CachedAt.Before(cutoff) || cached.AccessCount == 0 {
			delete(ac.coordinationPlans, key)
		}
	}

	for key, cached := range ac.performanceData {
		if cached.CachedAt.Before(cutoff) {
			delete(ac.performanceData, key)
		}
	}

	// If still too large, remove LRU entries
	ac.removeLRUEntries()

	ac.lastCleanup = time.Now()
}

// removeLRUEntries removes least recently used entries
func (ac *AdvancedCache) removeLRUEntries() {
	// Remove oldest agent results
	type cacheEntry struct {
		key        string
		lastAccess time.Time
	}

	var agentEntries []cacheEntry
	for key, cached := range ac.agentResults {
		agentEntries = append(agentEntries, cacheEntry{key, cached.LastAccess})
	}

	// Sort by last access (oldest first)
	for i := 0; i < len(agentEntries); i++ {
		for j := i + 1; j < len(agentEntries); j++ {
			if agentEntries[i].lastAccess.After(agentEntries[j].lastAccess) {
				agentEntries[i], agentEntries[j] = agentEntries[j], agentEntries[i]
			}
		}
	}

	// Remove oldest 20%
	removeCount := len(agentEntries) / 5
	for i := 0; i < removeCount && i < len(agentEntries); i++ {
		delete(ac.agentResults, agentEntries[i].key)
	}

	// Similar cleanup for coordination plans
	var planEntries []cacheEntry
	for key, cached := range ac.coordinationPlans {
		planEntries = append(planEntries, cacheEntry{key, cached.LastAccess})
	}

	// Sort and remove
	for i := 0; i < len(planEntries); i++ {
		for j := i + 1; j < len(planEntries); j++ {
			if planEntries[i].lastAccess.After(planEntries[j].lastAccess) {
				planEntries[i], planEntries[j] = planEntries[j], planEntries[i]
			}
		}
	}

	removeCount = len(planEntries) / 5
	for i := 0; i < removeCount && i < len(planEntries); i++ {
		delete(ac.coordinationPlans, planEntries[i].key)
	}
}

// calculateTypeHitRate calculates hit rate for a specific cache type
func (ac *AdvancedCache) calculateTypeHitRate(entryCount int) float64 {
	if entryCount == 0 {
		return 0.0
	}

	// Simplified calculation - in reality this would track per-type hits/misses
	return float64(ac.hitCount) / float64(ac.hitCount+ac.missCount)
}
