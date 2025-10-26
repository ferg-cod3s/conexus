package learning

import (
	"context"
	"sync"
	"time"
)

// PerformanceTracker tracks learning system performance
type PerformanceTracker struct {
	metrics         map[string]*UserPerformanceMetrics
	feedbackHistory []*ProcessedFeedback
	mu              sync.RWMutex
	lastUpdate      time.Time
}

// UserPerformanceMetrics tracks performance for a specific user
type UserPerformanceMetrics struct {
	UserID              string        `json:"user_id"`
	TotalSessions       int64         `json:"total_sessions"`
	SuccessfulSessions  int64         `json:"successful_sessions"`
	AverageSessionTime  time.Duration `json:"average_session_time"`
	AverageSatisfaction float32       `json:"average_satisfaction"`
	PreferredFeatures   []string      `json:"preferred_features"`
	LastActivity        time.Time     `json:"last_activity"`
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		metrics:         make(map[string]*UserPerformanceMetrics),
		feedbackHistory: make([]*ProcessedFeedback, 0),
		lastUpdate:      time.Now(),
	}
}

// Initialize initializes the performance tracker
func (pt *PerformanceTracker) Initialize(ctx context.Context) error {
	// Load existing metrics if available
	pt.loadExistingMetrics(ctx)
	return nil
}

// RecordFeedback records feedback and updates metrics
func (pt *PerformanceTracker) RecordFeedback(ctx context.Context, feedback *ProcessedFeedback) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	userID := feedback.OriginalFeedback.UserID
	if userID == "" {
		return
	}

	// Get or create user metrics
	metrics, exists := pt.metrics[userID]
	if !exists {
		metrics = &UserPerformanceMetrics{
			UserID:            userID,
			PreferredFeatures: make([]string, 0),
			LastActivity:      time.Now(),
		}
		pt.metrics[userID] = metrics
	}

	// Update metrics
	metrics.TotalSessions++
	metrics.LastActivity = time.Now()

	if feedback.QualityScore > 0.6 {
		metrics.SuccessfulSessions++
	}

	// Update average session time
	if metrics.TotalSessions == 1 {
		metrics.AverageSessionTime = feedback.OriginalFeedback.TimeSpent
	} else {
		// Exponential moving average
		alpha := 0.1
		metrics.AverageSessionTime = time.Duration(float64(metrics.AverageSessionTime)*(1-alpha) + float64(feedback.OriginalFeedback.TimeSpent)*alpha)
	}

	// Update satisfaction
	if metrics.TotalSessions == 1 {
		metrics.AverageSatisfaction = feedback.OriginalFeedback.OverallRating
	} else {
		alpha := float32(0.1)
		metrics.AverageSatisfaction = metrics.AverageSatisfaction*(1-alpha) + feedback.OriginalFeedback.OverallRating*alpha
	}

	// Update preferred features
	pt.updatePreferredFeatures(metrics, feedback)

	// Add to feedback history
	pt.feedbackHistory = append(pt.feedbackHistory, feedback)
	if len(pt.feedbackHistory) > 1000 {
		pt.feedbackHistory = pt.feedbackHistory[1:]
	}

	pt.lastUpdate = time.Now()
}

// GetUserSatisfaction returns average user satisfaction
func (pt *PerformanceTracker) GetUserSatisfaction() float64 {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	if len(pt.metrics) == 0 {
		return 0.0
	}

	totalSatisfaction := 0.0
	for _, metrics := range pt.metrics {
		totalSatisfaction += float64(metrics.AverageSatisfaction)
	}

	return totalSatisfaction / float64(len(pt.metrics))
}

// GetActiveUserCount returns the number of active users (recent activity)
func (pt *PerformanceTracker) GetActiveUserCount() int {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	cutoff := time.Now().Add(-7 * 24 * time.Hour) // Last 7 days
	activeCount := 0

	for _, metrics := range pt.metrics {
		if metrics.LastActivity.After(cutoff) {
			activeCount++
		}
	}

	return activeCount
}

// UpdateBaselines updates performance baselines
func (pt *PerformanceTracker) UpdateBaselines(ctx context.Context) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// Update baselines based on recent performance
	// In a real implementation, this would adjust thresholds and expectations
	pt.lastUpdate = time.Now()
}

// GetMetrics returns performance metrics
func (pt *PerformanceTracker) GetMetrics() map[string]interface{} {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	metrics := map[string]interface{}{
		"total_users":          len(pt.metrics),
		"total_feedback":       len(pt.feedbackHistory),
		"average_satisfaction": pt.GetUserSatisfaction(),
		"active_users":         pt.GetActiveUserCount(),
		"last_update":          pt.lastUpdate,
	}

	// User-specific metrics
	userMetrics := make(map[string]interface{})
	for userID, userMetric := range pt.metrics {
		userMetrics[userID] = map[string]interface{}{
			"total_sessions":       userMetric.TotalSessions,
			"successful_sessions":  userMetric.SuccessfulSessions,
			"average_session_time": userMetric.AverageSessionTime,
			"average_satisfaction": userMetric.AverageSatisfaction,
			"preferred_features":   userMetric.PreferredFeatures,
			"last_activity":        userMetric.LastActivity,
		}
	}
	metrics["user_metrics"] = userMetrics

	return metrics
}

// updatePreferredFeatures updates preferred features for a user
func (pt *PerformanceTracker) updatePreferredFeatures(metrics *UserPerformanceMetrics, feedback *ProcessedFeedback) {
	// Update based on feedback patterns
	for _, pattern := range feedback.Patterns {
		found := false
		for _, feature := range metrics.PreferredFeatures {
			if feature == pattern {
				found = true
				break
			}
		}
		if !found {
			metrics.PreferredFeatures = append(metrics.PreferredFeatures, pattern)
		}
	}

	// Limit to top 10 features
	if len(metrics.PreferredFeatures) > 10 {
		metrics.PreferredFeatures = metrics.PreferredFeatures[:10]
	}
}

// loadExistingMetrics loads existing performance metrics
func (pt *PerformanceTracker) loadExistingMetrics(ctx context.Context) {
	// Placeholder implementation
	// In a real system, this would load from database
}
