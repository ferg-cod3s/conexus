package learning

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
)

// AdaptiveRankingModel provides adaptive ranking based on user feedback
type AdaptiveRankingModel struct {
	featureWeights     map[string]float64
	userPreferences    map[string]*UserPreferences
	feedbackHistory    []*ProcessedFeedback
	performanceMetrics map[string]*RankingPerformance
	mu                 sync.RWMutex
	accuracy           float64
	healthScore        float64
	lastUpdate         time.Time
}

// RankingPerformance tracks ranking model performance
type RankingPerformance struct {
	Accuracy      float64   `json:"accuracy"`
	Precision     float64   `json:"precision"`
	Recall        float64   `json:"recall"`
	FeedbackCount int       `json:"feedback_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// UserPreferences represents learned user preferences
type UserPreferences struct {
	UserID            string                  `json:"user_id"`
	PreferredProfiles []profiles.AgentProfile `json:"preferred_profiles"`
	FeatureWeights    map[string]float64      `json:"feature_weights"`
	ContextWeights    map[string]float64      `json:"context_weights"`
	QualityThreshold  float32                 `json:"quality_threshold"`
	LastUpdated       time.Time               `json:"last_updated"`
}

// GetProfileConfidence returns confidence in a profile for the user
func (up *UserPreferences) GetProfileConfidence(profileID string) float64 {
	for _, profile := range up.PreferredProfiles {
		if profile.ID == profileID {
			return 0.8 // High confidence for preferred profiles
		}
	}
	return 0.3 // Low confidence for non-preferred profiles
}

// NewAdaptiveRankingModel creates a new adaptive ranking model
func NewAdaptiveRankingModel() *AdaptiveRankingModel {
	return &AdaptiveRankingModel{
		featureWeights:     make(map[string]float64),
		userPreferences:    make(map[string]*UserPreferences),
		feedbackHistory:    make([]*ProcessedFeedback, 0),
		performanceMetrics: make(map[string]*RankingPerformance),
		accuracy:           0.7,
		healthScore:        1.0,
		lastUpdate:         time.Now(),
	}
}

// Initialize initializes the ranking model
func (arm *AdaptiveRankingModel) Initialize(ctx context.Context) error {
	// Initialize default feature weights
	arm.initializeDefaultWeights()

	// Load historical data if available
	arm.loadHistoricalData(ctx)

	return nil
}

// Update updates the model with new feedback
func (arm *AdaptiveRankingModel) Update(ctx context.Context, feedback *ProcessedFeedback) error {
	arm.mu.Lock()
	defer arm.mu.Unlock()

	// Add to feedback history
	arm.feedbackHistory = append(arm.feedbackHistory, feedback)

	// Keep only recent feedback (last 1000 entries)
	if len(arm.feedbackHistory) > 1000 {
		arm.feedbackHistory = arm.feedbackHistory[1:]
	}

	// Update feature weights based on feedback
	arm.updateFeatureWeights(feedback)

	// Update user preferences
	arm.updateUserPreferences(feedback)

	// Update performance metrics
	arm.updatePerformanceMetrics(feedback)

	arm.lastUpdate = time.Now()
	return nil
}

// GetSuggestions gets ranking suggestions for a query
func (arm *AdaptiveRankingModel) GetSuggestions(ctx context.Context, query string, context map[string]interface{}, preferences *UserPreferences) ([]RankingSuggestion, error) {
	arm.mu.RLock()
	defer arm.mu.RUnlock()

	var suggestions []RankingSuggestion

	// Analyze query features
	features := arm.extractQueryFeatures(query, context)

	// Generate suggestions based on features and preferences
	for feature, weight := range features {
		if weight > 0.1 {
			suggestion := RankingSuggestion{
				Factor:      feature,
				Weight:      float32(weight),
				Explanation: arm.getFeatureExplanation(feature),
			}
			suggestions = append(suggestions, suggestion)
		}
	}

	// Sort by weight
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Weight > suggestions[j].Weight
	})

	// Limit to top 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions, nil
}

// UpdateFromRecentFeedback updates the model from recent feedback
func (arm *AdaptiveRankingModel) UpdateFromRecentFeedback(ctx context.Context) {
	arm.mu.Lock()
	defer arm.mu.Unlock()

	// Get recent feedback (last 24 hours)
	cutoff := time.Now().Add(-24 * time.Hour)
	var recentFeedback []*ProcessedFeedback

	for _, feedback := range arm.feedbackHistory {
		if feedback.ProcessedAt.After(cutoff) {
			recentFeedback = append(recentFeedback, feedback)
		}
	}

	// Update model based on recent feedback
	for _, feedback := range recentFeedback {
		arm.updateFeatureWeights(feedback)
		arm.updateUserPreferences(feedback)
	}

	arm.lastUpdate = time.Now()
}

// GetAccuracy returns the current model accuracy
func (arm *AdaptiveRankingModel) GetAccuracy() float64 {
	arm.mu.RLock()
	defer arm.mu.RUnlock()
	return arm.accuracy
}

// GetHealthScore returns the current model health score
func (arm *AdaptiveRankingModel) GetHealthScore() float64 {
	arm.mu.RLock()
	defer arm.mu.RUnlock()
	return arm.healthScore
}

// Save saves the model state
func (arm *AdaptiveRankingModel) Save(ctx context.Context) error {
	// Placeholder implementation
	// In a real system, this would save to persistent storage
	return nil
}

// initializeDefaultWeights initializes default feature weights
func (arm *AdaptiveRankingModel) initializeDefaultWeights() {
	arm.featureWeights = map[string]float64{
		"semantic_similarity":  1.0,
		"contextual_relevance": 0.8,
		"temporal_freshness":   0.6,
		"structural_relevance": 0.7,
		"user_behavior":        0.5,
		"query_complexity":     0.4,
		"result_diversity":     0.3,
		"evidence_quality":     0.9,
		"profile_match":        0.8,
		"click_through":        0.6,
	}
}

// loadHistoricalData loads historical training data
func (arm *AdaptiveRankingModel) loadHistoricalData(ctx context.Context) {
	// Placeholder implementation
	// In a real system, this would load from database or files
}

// extractQueryFeatures extracts features from query and context
func (arm *AdaptiveRankingModel) extractQueryFeatures(query string, context map[string]interface{}) map[string]float64 {
	features := make(map[string]float64)

	// Query-based features
	features["query_length"] = float64(len(query)) / 100.0
	features["query_complexity"] = arm.calculateQueryComplexity(query)

	// Context-based features
	if context != nil {
		if context["active_file"] != nil {
			features["has_active_file"] = 1.0
		}
		if context["git_branch"] != nil {
			features["has_git_branch"] = 1.0
		}
		if context["open_ticket_ids"] != nil {
			features["has_tickets"] = 1.0
		}
	}

	// Profile-based features
	if profileID, exists := context["profile_id"]; exists {
		if profileIDStr, ok := profileID.(string); ok {
			features["profile_"+profileIDStr] = 1.0
		}
	}

	return features
}

// calculateQueryComplexity calculates query complexity
func (arm *AdaptiveRankingModel) calculateQueryComplexity(query string) float64 {
	// Simple complexity based on length and technical terms
	length := len(query)
	complexity := float64(length) / 100.0

	technicalTerms := []string{"function", "class", "method", "algorithm", "implementation", "debug", "error", "security"}
	for _, term := range technicalTerms {
		if contains(query, term) {
			complexity += 0.1
		}
	}

	if complexity > 1.0 {
		complexity = 1.0
	}

	return complexity
}

// updateFeatureWeights updates feature weights based on feedback
func (arm *AdaptiveRankingModel) updateFeatureWeights(feedback *ProcessedFeedback) {
	// Update weights based on feedback quality
	qualityMultiplier := float64(feedback.QualityScore)

	for feature, currentWeight := range arm.featureWeights {
		// Adjust weight based on feedback patterns
		if arm.isFeatureRelevant(feature, feedback) {
			// Increase weight for relevant features
			arm.featureWeights[feature] = currentWeight * (1.0 + 0.1*qualityMultiplier)
		} else {
			// Decrease weight for irrelevant features
			arm.featureWeights[feature] = currentWeight * (1.0 - 0.05*qualityMultiplier)
		}

		// Ensure weights stay in reasonable range
		if arm.featureWeights[feature] > 2.0 {
			arm.featureWeights[feature] = 2.0
		}
		if arm.featureWeights[feature] < 0.1 {
			arm.featureWeights[feature] = 0.1
		}
	}
}

// isFeatureRelevant checks if a feature is relevant based on feedback
func (arm *AdaptiveRankingModel) isFeatureRelevant(feature string, feedback *ProcessedFeedback) bool {
	// Check if feature appears in feedback patterns or insights
	for _, pattern := range feedback.Patterns {
		if contains(pattern, feature) {
			return true
		}
	}

	for _, insight := range feedback.Insights {
		if contains(insight, feature) {
			return true
		}
	}

	return false
}

// updateUserPreferences updates user preferences based on feedback
func (arm *AdaptiveRankingModel) updateUserPreferences(feedback *ProcessedFeedback) {
	userID := feedback.OriginalFeedback.UserID
	if userID == "" {
		return
	}

	preferences, exists := arm.userPreferences[userID]
	if !exists {
		preferences = &UserPreferences{
			UserID:            userID,
			PreferredProfiles: make([]profiles.AgentProfile, 0),
			FeatureWeights:    make(map[string]float64),
			ContextWeights:    make(map[string]float64),
			QualityThreshold:  0.6,
			LastUpdated:       time.Now(),
		}
		arm.userPreferences[userID] = preferences
	}

	// Update based on feedback
	if feedback.OriginalFeedback.OverallRating > 0.7 {
		// Positive feedback - reinforce current preferences
		preferences.QualityThreshold = (preferences.QualityThreshold + feedback.QualityScore) / 2.0
	} else {
		// Negative feedback - adjust preferences
		preferences.QualityThreshold = (preferences.QualityThreshold + feedback.QualityScore) / 2.0
		if preferences.QualityThreshold < 0.3 {
			preferences.QualityThreshold = 0.3
		}
	}

	preferences.LastUpdated = time.Now()
}

// updatePerformanceMetrics updates performance metrics
func (arm *AdaptiveRankingModel) updatePerformanceMetrics(feedback *ProcessedFeedback) {
	// Calculate accuracy based on feedback quality
	qualityScore := float64(feedback.QualityScore)

	// Update running accuracy
	alpha := 0.1 // Learning rate
	arm.accuracy = arm.accuracy*(1-alpha) + qualityScore*alpha

	// Update health score
	if qualityScore > 0.7 {
		arm.healthScore = math.Min(arm.healthScore*1.01, 1.0)
	} else {
		arm.healthScore = math.Max(arm.healthScore*0.99, 0.5)
	}
}

// getFeatureExplanation returns explanation for a feature
func (arm *AdaptiveRankingModel) getFeatureExplanation(feature string) string {
	explanations := map[string]string{
		"semantic_similarity":  "Based on semantic similarity between query and content",
		"contextual_relevance": "Based on work context relevance (active files, branches, tickets)",
		"temporal_freshness":   "Based on how recent the content is",
		"structural_relevance": "Based on code structure and organization",
		"user_behavior":        "Based on user interaction patterns",
		"query_complexity":     "Based on query complexity and technical terms",
		"result_diversity":     "Based on diversity of result sources",
		"evidence_quality":     "Based on quality and quantity of supporting evidence",
		"profile_match":        "Based on match with agent profile preferences",
		"click_through":        "Based on user click-through behavior",
	}

	if explanation, exists := explanations[feature]; exists {
		return explanation
	}

	return "Based on learned patterns from user feedback"
}
