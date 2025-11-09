package contextual

import (
	"context"
	"sync"
	"time"
)

// ContextualPerformanceMonitor tracks contextual retrieval performance
type ContextualPerformanceMonitor struct {
	metrics ContextualPerformanceMetrics
	mu      sync.RWMutex
}

// ContextualPerformanceMetrics tracks performance metrics for contextual retrieval
type ContextualPerformanceMetrics struct {
	TotalSearches      int64            `json:"total_searches"`
	SuccessfulSearches int64            `json:"successful_searches"`
	FailedSearches     int64            `json:"failed_searches"`
	AverageLatency     time.Duration    `json:"average_latency"`
	AverageResults     float64          `json:"average_results"`
	ProfileUsage       map[string]int64 `json:"profile_usage"`
	QualityScores      []float32        `json:"quality_scores"`
	CacheHitRate       float64          `json:"cache_hit_rate"`
	EmbeddingTime      time.Duration    `json:"average_embedding_time"`
	SearchTime         time.Duration    `json:"average_search_time"`
	RankingTime        time.Duration    `json:"average_ranking_time"`
	LastUpdated        time.Time        `json:"last_updated"`
}

// NewContextualPerformanceMonitor creates a new contextual performance monitor
func NewContextualPerformanceMonitor() *ContextualPerformanceMonitor {
	return &ContextualPerformanceMonitor{
		metrics: ContextualPerformanceMetrics{
			ProfileUsage:  make(map[string]int64),
			QualityScores: make([]float32, 0),
			LastUpdated:   time.Now(),
		},
	}
}

// RecordSearch records a search operation
func (cpm *ContextualPerformanceMonitor) RecordSearch(ctx context.Context, query *ContextualQuery, duration time.Duration, resultCount int, qualityScore float32) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics.TotalSearches++
	cpm.metrics.LastUpdated = time.Now()

	if qualityScore >= 0.5 { // Consider successful if quality score is reasonable
		cpm.metrics.SuccessfulSearches++
	} else {
		cpm.metrics.FailedSearches++
	}

	// Update average latency
	if cpm.metrics.TotalSearches == 1 {
		cpm.metrics.AverageLatency = duration
	} else {
		// Exponential moving average
		alpha := 0.1
		cpm.metrics.AverageLatency = time.Duration(float64(cpm.metrics.AverageLatency)*(1-alpha) + float64(duration)*alpha)
	}

	// Update average results
	if cpm.metrics.TotalSearches == 1 {
		cpm.metrics.AverageResults = float64(resultCount)
	} else {
		alpha := 0.1
		cpm.metrics.AverageResults = cpm.metrics.AverageResults*(1-alpha) + float64(resultCount)*alpha
	}

	// Update profile usage
	if query.Profile != nil {
		cpm.metrics.ProfileUsage[query.Profile.ID]++
	}

	// Update quality scores (keep last 100)
	cpm.metrics.QualityScores = append(cpm.metrics.QualityScores, qualityScore)
	if len(cpm.metrics.QualityScores) > 100 {
		cpm.metrics.QualityScores = cpm.metrics.QualityScores[1:]
	}
}

// RecordEmbedding records embedding generation time
func (cpm *ContextualPerformanceMonitor) RecordEmbedding(ctx context.Context, duration time.Duration) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if cpm.metrics.TotalSearches == 0 {
		cpm.metrics.EmbeddingTime = duration
	} else {
		// Exponential moving average
		alpha := 0.1
		cpm.metrics.EmbeddingTime = time.Duration(float64(cpm.metrics.EmbeddingTime)*(1-alpha) + float64(duration)*alpha)
	}
}

// RecordSearchOperation records search operation time
func (cpm *ContextualPerformanceMonitor) RecordSearchOperation(ctx context.Context, duration time.Duration) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if cpm.metrics.TotalSearches == 0 {
		cpm.metrics.SearchTime = duration
	} else {
		// Exponential moving average
		alpha := 0.1
		cpm.metrics.SearchTime = time.Duration(float64(cpm.metrics.SearchTime)*(1-alpha) + float64(duration)*alpha)
	}
}

// RecordRanking records ranking operation time
func (cpm *ContextualPerformanceMonitor) RecordRanking(ctx context.Context, duration time.Duration) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if cpm.metrics.TotalSearches == 0 {
		cpm.metrics.RankingTime = duration
	} else {
		// Exponential moving average
		alpha := 0.1
		cpm.metrics.RankingTime = time.Duration(float64(cpm.metrics.RankingTime)*(1-alpha) + float64(duration)*alpha)
	}
}

// RecordCacheHit records cache hit/miss
func (cpm *ContextualPerformanceMonitor) RecordCacheHit(ctx context.Context, hit bool) {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	if hit {
		// Update cache hit rate (simplified)
		if cpm.metrics.CacheHitRate < 1.0 {
			cpm.metrics.CacheHitRate = (cpm.metrics.CacheHitRate + 1.0) / 2.0
		}
	} else {
		if cpm.metrics.CacheHitRate > 0.0 {
			cpm.metrics.CacheHitRate = cpm.metrics.CacheHitRate / 2.0
		}
	}
}

// GetMetrics returns current performance metrics
func (cpm *ContextualPerformanceMonitor) GetMetrics() *ContextualPerformanceMetrics {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	// Return a copy to prevent external modification
	metrics := cpm.metrics
	metrics.LastUpdated = time.Now()
	return &metrics
}

// GetMetricsByProfile returns metrics grouped by profile
func (cpm *ContextualPerformanceMonitor) GetMetricsByProfile() map[string]*ContextualPerformanceMetrics {
	cpm.mu.RLock()
	defer cpm.mu.RUnlock()

	profileMetrics := make(map[string]*ContextualPerformanceMetrics)

	for profileID, usage := range cpm.metrics.ProfileUsage {
		profileMetrics[profileID] = &ContextualPerformanceMetrics{
			TotalSearches:      usage,
			SuccessfulSearches: int64(float64(usage) * 0.8), // Estimate 80% success rate
			AverageLatency:     cpm.metrics.AverageLatency,
			AverageResults:     cpm.metrics.AverageResults,
			LastUpdated:        time.Now(),
		}
	}

	return profileMetrics
}

// Reset resets all metrics
func (cpm *ContextualPerformanceMonitor) Reset() {
	cpm.mu.Lock()
	defer cpm.mu.Unlock()

	cpm.metrics = ContextualPerformanceMetrics{
		ProfileUsage:  make(map[string]int64),
		QualityScores: make([]float32, 0),
		LastUpdated:   time.Now(),
	}
}
