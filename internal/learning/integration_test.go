package learning

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
	"github.com/ferg-cod3s/conexus/internal/orchestrator/multiagent"
	"github.com/ferg-cod3s/conexus/internal/search/contextual"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLearningSystemIntegration tests the complete learning system integration
func TestLearningSystemIntegration(t *testing.T) {
	// Create all components
	profileManager := profiles.NewProfileManager(&MockClassifier{})
	multiAgentSystem := &multiagent.MultiAgentOrchestrator{}          // Mock for testing
	contextualFramework := &contextual.ContextualRetrievalFramework{} // Mock for testing

	feedbackProcessor := NewFeedbackProcessor()
	rankingModel := NewAdaptiveRankingModel()
	preferenceLearner := NewUserPreferenceLearner()
	performanceTracker := NewPerformanceTracker()
	modelUpdater := NewModelUpdater()
	analyticsEngine := NewAnalyticsEngine()

	config := LearningConfig{
		ProfileManager:      profileManager,
		MultiAgentSystem:    multiAgentSystem,
		ContextualFramework: contextualFramework,
		FeedbackProcessor:   feedbackProcessor,
		RankingModel:        rankingModel,
		PreferenceLearner:   preferenceLearner,
		PerformanceTracker:  performanceTracker,
		ModelUpdater:        modelUpdater,
		AnalyticsEngine:     analyticsEngine,
	}

	system := NewLearningSystem(config)

	// Start the learning system
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := system.Start(ctx)
	require.NoError(t, err)

	// Simulate user interactions and feedback
	userID := "test-user-123"

	// Process multiple feedback instances
	feedbacks := []*FeedbackData{
		{
			SessionID:     "session-1",
			UserID:        userID,
			Query:         "debug authentication error",
			OverallRating: 0.8,
			Helpful:       true,
			TimeSpent:     2 * time.Minute,
			Results: []FeedbackResult{
				{Rank: 1, Clicked: true, Rating: 0.9, Useful: true},
				{Rank: 2, Clicked: false, Rating: 0.6, Useful: false},
			},
			Context: map[string]interface{}{
				"active_file": "auth.go",
				"git_branch":  "bugfix/auth",
			},
			Timestamp: time.Now(),
		},
		{
			SessionID:     "session-2",
			UserID:        userID,
			Query:         "security implementation review",
			OverallRating: 0.9,
			Helpful:       true,
			TimeSpent:     3 * time.Minute,
			Results: []FeedbackResult{
				{Rank: 1, Clicked: true, Rating: 0.95, Useful: true},
				{Rank: 2, Clicked: true, Rating: 0.85, Useful: true},
			},
			Context: map[string]interface{}{
				"active_file": "security.go",
				"git_branch":  "security",
			},
			Timestamp: time.Now(),
		},
		{
			SessionID:     "session-3",
			UserID:        userID,
			Query:         "documentation for API",
			OverallRating: 0.7,
			Helpful:       true,
			TimeSpent:     1 * time.Minute,
			Results: []FeedbackResult{
				{Rank: 1, Clicked: true, Rating: 0.8, Useful: true},
			},
			Context: map[string]interface{}{
				"active_file": "README.md",
			},
			Timestamp: time.Now(),
		},
	}

	// Process all feedback
	for _, feedback := range feedbacks {
		err := system.ProcessFeedback(ctx, feedback)
		require.NoError(t, err)

		// Small delay to simulate real-time processing
		time.Sleep(10 * time.Millisecond)
	}

	// Get personalized recommendations
	recommendations, err := system.GetRecommendations(ctx, userID, "debug security issue", map[string]interface{}{
		"active_file": "security.go",
		"profile_id":  "security",
	})
	require.NoError(t, err)

	assert.NotNil(t, recommendations)
	assert.Equal(t, userID, recommendations.UserID)
	assert.NotEmpty(t, recommendations.RankingSuggestions)
	assert.NotEmpty(t, recommendations.ProfileSuggestions)

	// Get system metrics
	metrics := system.GetMetrics()
	assert.Greater(t, metrics.FeedbackCollectionRate, float64(0))
	assert.Greater(t, metrics.ModelAccuracy, float64(0))
	assert.Greater(t, metrics.UserSatisfaction, float64(0))

	// Verify learning adaptation
	// The system should have learned from the feedback
	assert.Greater(t, metrics.PerformanceImprovement, float64(0))

	// Stop the learning system
	err = system.Stop(ctx)
	require.NoError(t, err)
}

// TestLearningSystemStress tests the learning system under load
func TestLearningSystemStress(t *testing.T) {
	// Create learning system
	profileManager := profiles.NewProfileManager(&MockClassifier{})
	feedbackProcessor := NewFeedbackProcessor()
	rankingModel := NewAdaptiveRankingModel()
	preferenceLearner := NewUserPreferenceLearner()
	performanceTracker := NewPerformanceTracker()
	modelUpdater := NewModelUpdater()
	analyticsEngine := NewAnalyticsEngine()

	config := LearningConfig{
		ProfileManager:     profileManager,
		FeedbackProcessor:  feedbackProcessor,
		RankingModel:       rankingModel,
		PreferenceLearner:  preferenceLearner,
		PerformanceTracker: performanceTracker,
		ModelUpdater:       modelUpdater,
		AnalyticsEngine:    analyticsEngine,
	}

	system := NewLearningSystem(config)

	// Start system
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := system.Start(ctx)
	require.NoError(t, err)

	// Simulate high load with many users and feedback
	users := 10
	feedbackPerUser := 5

	startTime := time.Now()

	for userID := 0; userID < users; userID++ {
		for feedbackID := 0; feedbackID < feedbackPerUser; feedbackID++ {
			feedback := &FeedbackData{
				SessionID:     fmt.Sprintf("stress-session-%d-%d", userID, feedbackID),
				UserID:        fmt.Sprintf("stress-user-%d", userID),
				Query:         fmt.Sprintf("query %d from user %d", feedbackID, userID),
				OverallRating: 0.7 + float32(feedbackID)*0.05, // Varying ratings
				Helpful:       feedbackID%2 == 0,
				TimeSpent:     time.Duration(30+feedbackID*10) * time.Second,
				Results: []FeedbackResult{
					{Rank: 1, Clicked: true, Rating: 0.8, Useful: true},
				},
				Context: map[string]interface{}{
					"active_file": fmt.Sprintf("file-%d.go", userID),
				},
				Timestamp: time.Now(),
			}

			err := system.ProcessFeedback(ctx, feedback)
			require.NoError(t, err)
		}
	}

	processingTime := time.Since(startTime)

	// Verify system handled the load
	metrics := system.GetMetrics()
	assert.Greater(t, metrics.FeedbackCollectionRate, float64(0.5)) // Should maintain good collection rate
	assert.Greater(t, metrics.ActiveUsers, 5)                       // Should track multiple users

	// Performance should be reasonable
	assert.Less(t, processingTime, 10*time.Second) // Should process quickly

	// Stop system
	err = system.Stop(ctx)
	require.NoError(t, err)
}

// TestLearningSystemAdaptation tests how the system adapts to user behavior
func TestLearningSystemAdaptation(t *testing.T) {
	profileManager := profiles.NewProfileManager(&MockClassifier{})
	feedbackProcessor := NewFeedbackProcessor()
	rankingModel := NewAdaptiveRankingModel()
	preferenceLearner := NewUserPreferenceLearner()
	performanceTracker := NewPerformanceTracker()
	modelUpdater := NewModelUpdater()
	analyticsEngine := NewAnalyticsEngine()

	config := LearningConfig{
		ProfileManager:     profileManager,
		FeedbackProcessor:  feedbackProcessor,
		RankingModel:       rankingModel,
		PreferenceLearner:  preferenceLearner,
		PerformanceTracker: performanceTracker,
		ModelUpdater:       modelUpdater,
		AnalyticsEngine:    analyticsEngine,
	}

	system := NewLearningSystem(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := system.Start(ctx)
	require.NoError(t, err)

	userID := "adaptation-user"

	// Initial feedback - user prefers debugging
	initialFeedback := []*FeedbackData{
		{
			UserID:        userID,
			Query:         "debug this function",
			OverallRating: 0.9,
			Helpful:       true,
			Context:       map[string]interface{}{"active_file": "debug.go"},
		},
		{
			UserID:        userID,
			Query:         "fix error in code",
			OverallRating: 0.8,
			Helpful:       true,
			Context:       map[string]interface{}{"active_file": "debug.go"},
		},
	}

	for _, feedback := range initialFeedback {
		err := system.ProcessFeedback(ctx, feedback)
		require.NoError(t, err)
	}

	// Get initial recommendations
	initialRecs, err := system.GetRecommendations(ctx, userID, "debug issue", map[string]interface{}{
		"active_file": "debug.go",
	})
	require.NoError(t, err)

	// More feedback - user starts asking about security
	securityFeedback := []*FeedbackData{
		{
			UserID:        userID,
			Query:         "security vulnerability check",
			OverallRating: 0.9,
			Helpful:       true,
			Context:       map[string]interface{}{"active_file": "security.go"},
		},
		{
			UserID:        userID,
			Query:         "authentication implementation",
			OverallRating: 0.8,
			Helpful:       true,
			Context:       map[string]interface{}{"active_file": "security.go"},
		},
	}

	for _, feedback := range securityFeedback {
		err := system.ProcessFeedback(ctx, feedback)
		require.NoError(t, err)
	}

	// Get updated recommendations
	updatedRecs, err := system.GetRecommendations(ctx, userID, "security issue", map[string]interface{}{
		"active_file": "security.go",
	})
	require.NoError(t, err)

	// System should adapt to new preferences
	assert.NotNil(t, initialRecs)
	assert.NotNil(t, updatedRecs)

	// Verify learning occurred
	metrics := system.GetMetrics()
	assert.Greater(t, metrics.ModelAccuracy, float64(0.6)) // Should improve with more data

	err = system.Stop(ctx)
	require.NoError(t, err)
}

// TestLearningSystemRecovery tests system recovery from errors
func TestLearningSystemRecovery(t *testing.T) {
	profileManager := profiles.NewProfileManager(&MockClassifier{})
	feedbackProcessor := NewFeedbackProcessor()
	rankingModel := NewAdaptiveRankingModel()
	preferenceLearner := NewUserPreferenceLearner()
	performanceTracker := NewPerformanceTracker()
	modelUpdater := NewModelUpdater()
	analyticsEngine := NewAnalyticsEngine()

	config := LearningConfig{
		ProfileManager:     profileManager,
		FeedbackProcessor:  feedbackProcessor,
		RankingModel:       rankingModel,
		PreferenceLearner:  preferenceLearner,
		PerformanceTracker: performanceTracker,
		ModelUpdater:       modelUpdater,
		AnalyticsEngine:    analyticsEngine,
	}

	system := NewLearningSystem(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start system
	err := system.Start(ctx)
	require.NoError(t, err)

	// Process some feedback
	feedback := &FeedbackData{
		UserID:        "recovery-user",
		Query:         "test query",
		OverallRating: 0.8,
		Helpful:       true,
	}

	err = system.ProcessFeedback(ctx, feedback)
	require.NoError(t, err)

	// Simulate system stress
	for i := 0; i < 100; i++ {
		feedback := &FeedbackData{
			UserID:        fmt.Sprintf("stress-user-%d", i),
			Query:         fmt.Sprintf("stress query %d", i),
			OverallRating: 0.7,
			Helpful:       true,
		}

		err := system.ProcessFeedback(ctx, feedback)
		require.NoError(t, err)
	}

	// System should still be functional
	metrics := system.GetMetrics()
	assert.Greater(t, metrics.FeedbackCollectionRate, float64(0))

	// Stop system
	err = system.Stop(ctx)
	require.NoError(t, err)
}

// TestLearningSystemConcurrency tests concurrent access to the learning system
func TestLearningSystemConcurrency(t *testing.T) {
	profileManager := profiles.NewProfileManager(&MockClassifier{})
	feedbackProcessor := NewFeedbackProcessor()
	rankingModel := NewAdaptiveRankingModel()
	preferenceLearner := NewUserPreferenceLearner()
	performanceTracker := NewPerformanceTracker()
	modelUpdater := NewModelUpdater()
	analyticsEngine := NewAnalyticsEngine()

	config := LearningConfig{
		ProfileManager:     profileManager,
		FeedbackProcessor:  feedbackProcessor,
		RankingModel:       rankingModel,
		PreferenceLearner:  preferenceLearner,
		PerformanceTracker: performanceTracker,
		ModelUpdater:       modelUpdater,
		AnalyticsEngine:    analyticsEngine,
	}

	system := NewLearningSystem(config)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	err := system.Start(ctx)
	require.NoError(t, err)

	// Test concurrent feedback processing
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(userID int) {
			defer func() { done <- true }()

			for j := 0; j < 5; j++ {
				feedback := &FeedbackData{
					UserID:        fmt.Sprintf("concurrent-user-%d", userID),
					Query:         fmt.Sprintf("concurrent query %d-%d", userID, j),
					OverallRating: 0.8,
					Helpful:       true,
				}

				err := system.ProcessFeedback(ctx, feedback)
				if err != nil {
					t.Errorf("Concurrent feedback processing failed: %v", err)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent test timed out")
		}
	}

	// Verify system is still functional
	metrics := system.GetMetrics()
	assert.Greater(t, metrics.FeedbackCollectionRate, float64(0))

	err = system.Stop(ctx)
	require.NoError(t, err)
}

// TestLearningSystemPersistence tests data persistence across restarts
func TestLearningSystemPersistence(t *testing.T) {
	// This would test saving and loading learning data
	// For now, just verify the system can be started and stopped multiple times

	profileManager := profiles.NewProfileManager(&MockClassifier{})
	feedbackProcessor := NewFeedbackProcessor()
	rankingModel := NewAdaptiveRankingModel()
	preferenceLearner := NewUserPreferenceLearner()
	performanceTracker := NewPerformanceTracker()
	modelUpdater := NewModelUpdater()
	analyticsEngine := NewAnalyticsEngine()

	config := LearningConfig{
		ProfileManager:     profileManager,
		FeedbackProcessor:  feedbackProcessor,
		RankingModel:       rankingModel,
		PreferenceLearner:  preferenceLearner,
		PerformanceTracker: performanceTracker,
		ModelUpdater:       modelUpdater,
		AnalyticsEngine:    analyticsEngine,
	}

	// Start and stop system multiple times
	for i := 0; i < 3; i++ {
		system := NewLearningSystem(config)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := system.Start(ctx)
		require.NoError(t, err)

		// Process some feedback
		feedback := &FeedbackData{
			UserID:        fmt.Sprintf("persist-user-%d", i),
			Query:         "persistence test",
			OverallRating: 0.8,
			Helpful:       true,
		}

		err = system.ProcessFeedback(ctx, feedback)
		require.NoError(t, err)

		err = system.Stop(ctx)
		require.NoError(t, err)
	}
}
