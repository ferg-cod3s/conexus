package learning

import (
	"context"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
)

// UserPreferenceLearner learns user preferences from feedback
type UserPreferenceLearner struct {
	userPreferences map[string]*UserPreferences
	feedbackHistory []*ProcessedFeedback
	mu              sync.RWMutex
	lastUpdate      time.Time
}

// NewUserPreferenceLearner creates a new user preference learner
func NewUserPreferenceLearner() *UserPreferenceLearner {
	return &UserPreferenceLearner{
		userPreferences: make(map[string]*UserPreferences),
		feedbackHistory: make([]*ProcessedFeedback, 0),
		lastUpdate:      time.Now(),
	}
}

// Initialize initializes the preference learner
func (upl *UserPreferenceLearner) Initialize(ctx context.Context) error {
	// Load existing preferences if available
	upl.loadExistingPreferences(ctx)
	return nil
}

// Update updates preferences based on feedback
func (upl *UserPreferenceLearner) Update(ctx context.Context, feedback *ProcessedFeedback) error {
	upl.mu.Lock()
	defer upl.mu.Unlock()

	userID := feedback.OriginalFeedback.UserID
	if userID == "" {
		return nil // Skip if no user ID
	}

	// Get or create user preferences
	preferences, exists := upl.userPreferences[userID]
	if !exists {
		preferences = &UserPreferences{
			UserID:            userID,
			PreferredProfiles: make([]profiles.AgentProfile, 0),
			FeatureWeights:    make(map[string]float64),
			ContextWeights:    make(map[string]float64),
			QualityThreshold:  0.6,
			LastUpdated:       time.Now(),
		}
		upl.userPreferences[userID] = preferences
	}

	// Update preferences based on feedback
	upl.updateFromFeedback(preferences, feedback)

	// Add to feedback history
	upl.feedbackHistory = append(upl.feedbackHistory, feedback)
	if len(upl.feedbackHistory) > 1000 {
		upl.feedbackHistory = upl.feedbackHistory[1:]
	}

	upl.lastUpdate = time.Now()
	return nil
}

// GetPreferences gets preferences for a user
func (upl *UserPreferenceLearner) GetPreferences(ctx context.Context, userID string) (*UserPreferences, error) {
	upl.mu.RLock()
	defer upl.mu.RUnlock()

	preferences, exists := upl.userPreferences[userID]
	if !exists {
		// Return default preferences
		return &UserPreferences{
			UserID:            userID,
			PreferredProfiles: []profiles.AgentProfile{*profiles.GetProfileByID("general")},
			FeatureWeights:    make(map[string]float64),
			ContextWeights:    make(map[string]float64),
			QualityThreshold:  0.6,
			LastUpdated:       time.Now(),
		}, nil
	}

	return preferences, nil
}

// UpdateFromFeedback updates preferences from recent feedback
func (upl *UserPreferenceLearner) UpdateFromFeedback(ctx context.Context) {
	upl.mu.Lock()
	defer upl.mu.Unlock()

	// Get recent feedback (last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	var recentFeedback []*ProcessedFeedback

	for _, feedback := range upl.feedbackHistory {
		if feedback.ProcessedAt.After(cutoff) {
			recentFeedback = append(recentFeedback, feedback)
		}
	}

	// Update preferences for all users with recent feedback
	for _, feedback := range recentFeedback {
		userID := feedback.OriginalFeedback.UserID
		if userID == "" {
			continue
		}

		preferences, exists := upl.userPreferences[userID]
		if !exists {
			preferences = &UserPreferences{
				UserID:            userID,
				PreferredProfiles: make([]profiles.AgentProfile, 0),
				FeatureWeights:    make(map[string]float64),
				ContextWeights:    make(map[string]float64),
				QualityThreshold:  0.6,
				LastUpdated:       time.Now(),
			}
			upl.userPreferences[userID] = preferences
		}

		upl.updateFromFeedback(preferences, feedback)
	}

	upl.lastUpdate = time.Now()
}

// Save saves user preferences
func (upl *UserPreferenceLearner) Save(ctx context.Context) error {
	// Placeholder implementation
	// In a real system, this would save to persistent storage
	return nil
}

// updateFromFeedback updates preferences based on a single feedback
func (upl *UserPreferenceLearner) updateFromFeedback(preferences *UserPreferences, feedback *ProcessedFeedback) {
	// Update quality threshold
	qualityScore := feedback.QualityScore
	preferences.QualityThreshold = (preferences.QualityThreshold + qualityScore) / 2.0

	// Update feature weights based on feedback patterns
	for _, pattern := range feedback.Patterns {
		if weight, exists := preferences.FeatureWeights[pattern]; exists {
			preferences.FeatureWeights[pattern] = (weight + 0.1) / 1.1
		} else {
			preferences.FeatureWeights[pattern] = 0.5
		}
	}

	// Update context weights based on context usage
	if feedback.OriginalFeedback.Context != nil {
		if feedback.OriginalFeedback.Context["active_file"] != nil {
			if weight, exists := preferences.ContextWeights["active_file"]; exists {
				preferences.ContextWeights["active_file"] = (weight + 0.1) / 1.1
			} else {
				preferences.ContextWeights["active_file"] = 0.5
			}
		}

		if feedback.OriginalFeedback.Context["git_branch"] != nil {
			if weight, exists := preferences.ContextWeights["git_branch"]; exists {
				preferences.ContextWeights["git_branch"] = (weight + 0.1) / 1.1
			} else {
				preferences.ContextWeights["git_branch"] = 0.5
			}
		}
	}

	// Update preferred profiles based on feedback quality
	if qualityScore > 0.7 {
		// High quality feedback - reinforce current profile preferences
		// In a real implementation, this would analyze which profile was used
	} else if qualityScore < 0.4 {
		// Low quality feedback - consider alternative profiles
		// In a real implementation, this would suggest profile changes
	}

	preferences.LastUpdated = time.Now()
}

// loadExistingPreferences loads existing user preferences
func (upl *UserPreferenceLearner) loadExistingPreferences(ctx context.Context) {
	// Placeholder implementation
	// In a real system, this would load from database
}

// GetAllPreferences returns all user preferences
func (upl *UserPreferenceLearner) GetAllPreferences() map[string]*UserPreferences {
	upl.mu.RLock()
	defer upl.mu.RUnlock()

	// Return copies to prevent external modification
	preferences := make(map[string]*UserPreferences)
	for userID, pref := range upl.userPreferences {
		preferences[userID] = pref
	}

	return preferences
}

// GetPreferenceStats returns preference statistics
func (upl *UserPreferenceLearner) GetPreferenceStats() map[string]interface{} {
	upl.mu.RLock()
	defer upl.mu.RUnlock()

	stats := map[string]interface{}{
		"total_users":     len(upl.userPreferences),
		"feedback_count":  len(upl.feedbackHistory),
		"average_quality": 0.0,
		"common_patterns": make(map[string]int),
		"common_contexts": make(map[string]int),
		"last_update":     upl.lastUpdate,
	}

	// Calculate average quality
	totalQuality := float32(0.0)
	feedbackCount := 0
	for _, feedback := range upl.feedbackHistory {
		totalQuality += feedback.QualityScore
		feedbackCount++
	}
	if feedbackCount > 0 {
		stats["average_quality"] = float64(totalQuality / float32(feedbackCount))
	}

	// Count common patterns
	patternCounts := make(map[string]int)
	for _, feedback := range upl.feedbackHistory {
		for _, pattern := range feedback.Patterns {
			patternCounts[pattern]++
		}
	}
	stats["common_patterns"] = patternCounts

	// Count common contexts
	contextCounts := make(map[string]int)
	for _, feedback := range upl.feedbackHistory {
		if feedback.OriginalFeedback.Context != nil {
			if feedback.OriginalFeedback.Context["active_file"] != nil {
				contextCounts["active_file"]++
			}
			if feedback.OriginalFeedback.Context["git_branch"] != nil {
				contextCounts["git_branch"]++
			}
			if feedback.OriginalFeedback.Context["open_ticket_ids"] != nil {
				contextCounts["open_tickets"]++
			}
		}
	}
	stats["common_contexts"] = contextCounts

	return stats
}
