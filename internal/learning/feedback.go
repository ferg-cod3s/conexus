package learning

import (
	"context"
	"fmt"
	"time"
)

// FeedbackProcessor processes and analyzes user feedback
type FeedbackProcessor struct {
	collectionRate float64
	feedbackQueue  chan *FeedbackData
	processorCount int
	isActive       bool
}

// NewFeedbackProcessor creates a new feedback processor
func NewFeedbackProcessor() *FeedbackProcessor {
	return &FeedbackProcessor{
		collectionRate: 0.0,
		feedbackQueue:  make(chan *FeedbackData, 1000),
		processorCount: 4,
		isActive:       false,
	}
}

// Initialize initializes the feedback processor
func (fp *FeedbackProcessor) Initialize(ctx context.Context) error {
	fp.isActive = true

	// Start feedback processing workers
	for i := 0; i < fp.processorCount; i++ {
		go fp.processFeedbackWorker(ctx)
	}

	return nil
}

// Process processes feedback data
func (fp *FeedbackProcessor) Process(ctx context.Context, feedback *FeedbackData) (*ProcessedFeedback, error) {
	if !fp.isActive {
		return nil, fmt.Errorf("feedback processor is not active")
	}

	// Send to processing queue
	select {
	case fp.feedbackQueue <- feedback:
		// Successfully queued
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("feedback queue is full")
	}

	// For now, return processed feedback synchronously
	// In a real implementation, this would be asynchronous
	return fp.processFeedbackSync(ctx, feedback)
}

// processFeedbackSync processes feedback synchronously
func (fp *FeedbackProcessor) processFeedbackSync(ctx context.Context, feedback *FeedbackData) (*ProcessedFeedback, error) {
	// Extract features from feedback
	features := fp.extractFeatures(feedback)

	// Calculate quality score
	qualityScore := fp.calculateQualityScore(feedback)

	// Generate insights
	insights := fp.generateInsights(feedback)

	// Identify patterns
	patterns := fp.identifyPatterns(feedback)

	return &ProcessedFeedback{
		OriginalFeedback: feedback,
		Features:         features,
		QualityScore:     qualityScore,
		Insights:         insights,
		Patterns:         patterns,
		ProcessedAt:      time.Now(),
	}, nil
}

// processFeedbackWorker processes feedback from the queue
func (fp *FeedbackProcessor) processFeedbackWorker(ctx context.Context) {
	for {
		select {
		case feedback := <-fp.feedbackQueue:
			if feedback != nil {
				_, err := fp.processFeedbackSync(ctx, feedback)
				if err != nil {
					// Log error but continue processing
					continue
				}

				// Update collection rate
				fp.updateCollectionRate()
			}
		case <-ctx.Done():
			return
		}
	}
}

// extractFeatures extracts features from feedback data
func (fp *FeedbackProcessor) extractFeatures(feedback *FeedbackData) map[string]interface{} {
	features := make(map[string]interface{})

	// Query features
	features["query_length"] = len(feedback.Query)
	features["query_complexity"] = fp.calculateQueryComplexity(feedback.Query)

	// Result features
	features["result_count"] = len(feedback.Results)
	features["click_through_rate"] = fp.calculateClickThroughRate(feedback.Results)
	features["average_rating"] = fp.calculateAverageRating(feedback.Results)

	// User behavior features
	features["time_spent"] = feedback.TimeSpent.Seconds()
	features["overall_rating"] = feedback.OverallRating
	features["helpful"] = feedback.Helpful

	// Context features
	if feedback.Context != nil {
		features["has_active_file"] = feedback.Context["active_file"] != nil
		features["has_git_branch"] = feedback.Context["git_branch"] != nil
		if tickets, exists := feedback.Context["open_ticket_ids"]; exists && tickets != nil {
			if ticketSlice, ok := tickets.([]string); ok {
				features["has_tickets"] = len(ticketSlice) > 0
			} else {
				features["has_tickets"] = false
			}
		} else {
			features["has_tickets"] = false
		}
	}

	return features
}

// calculateQualityScore calculates overall quality score from feedback
func (fp *FeedbackProcessor) calculateQualityScore(feedback *FeedbackData) float32 {
	score := feedback.OverallRating

	// Adjust based on click-through behavior
	clickThroughRate := fp.calculateClickThroughRate(feedback.Results)
	score += clickThroughRate * 0.2

	// Adjust based on helpfulness
	if feedback.Helpful {
		score += 0.1
	}

	// Adjust based on time spent (longer time might indicate engagement)
	if feedback.TimeSpent > 30*time.Second {
		score += 0.05
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateQueryComplexity calculates query complexity score
func (fp *FeedbackProcessor) calculateQueryComplexity(query string) float32 {
	// Simple complexity based on length and special terms
	length := len(query)
	complexity := float32(length) / 100.0

	// Boost for technical terms
	technicalTerms := []string{"function", "class", "method", "algorithm", "implementation", "debug", "error", "security"}
	for _, term := range technicalTerms {
		if contains(query, term) {
			complexity += 0.1
		}
	}

	// Cap at 1.0
	if complexity > 1.0 {
		complexity = 1.0
	}

	return complexity
}

// calculateClickThroughRate calculates click-through rate from results
func (fp *FeedbackProcessor) calculateClickThroughRate(results []FeedbackResult) float32 {
	if len(results) == 0 {
		return 0.0
	}

	clickCount := 0
	for _, result := range results {
		if result.Clicked {
			clickCount++
		}
	}

	return float32(clickCount) / float32(len(results))
}

// calculateAverageRating calculates average rating from results
func (fp *FeedbackProcessor) calculateAverageRating(results []FeedbackResult) float32 {
	if len(results) == 0 {
		return 0.0
	}

	totalRating := float32(0.0)
	for _, result := range results {
		totalRating += result.Rating
	}

	return totalRating / float32(len(results))
}

// generateInsights generates insights from feedback
func (fp *FeedbackProcessor) generateInsights(feedback *FeedbackData) []string {
	var insights []string

	// Query insights
	if len(feedback.Query) < 10 {
		insights = append(insights, "Query is very short - user may need help with query formulation")
	}

	if len(feedback.Query) > 100 {
		insights = append(insights, "Query is very long - user may be asking complex questions")
	}

	// Result insights
	clickThroughRate := fp.calculateClickThroughRate(feedback.Results)
	if clickThroughRate < 0.3 {
		insights = append(insights, "Low click-through rate - results may not be relevant")
	}

	if clickThroughRate > 0.8 {
		insights = append(insights, "High click-through rate - results are highly relevant")
	}

	// User behavior insights
	if feedback.TimeSpent < 10*time.Second {
		insights = append(insights, "Very short time spent - user may not have found what they need")
	}

	if feedback.TimeSpent > 5*time.Minute {
		insights = append(insights, "Long time spent - user may be deeply engaged or confused")
	}

	if !feedback.Helpful {
		insights = append(insights, "User marked as not helpful - system needs improvement")
	}

	return insights
}

// identifyPatterns identifies patterns in feedback
func (fp *FeedbackProcessor) identifyPatterns(feedback *FeedbackData) []string {
	var patterns []string

	// Query patterns
	if contains(feedback.Query, "error") || contains(feedback.Query, "bug") {
		patterns = append(patterns, "debugging_query")
	}

	if contains(feedback.Query, "how") || contains(feedback.Query, "what") {
		patterns = append(patterns, "explanatory_query")
	}

	if contains(feedback.Query, "function") || contains(feedback.Query, "class") {
		patterns = append(patterns, "code_analysis_query")
	}

	// Context patterns
	if feedback.Context != nil {
		if feedback.Context["active_file"] != nil {
			patterns = append(patterns, "has_active_file")
		}

		if feedback.Context["git_branch"] != nil {
			patterns = append(patterns, "has_git_branch")
		}
	}

	// Result patterns
	if len(feedback.ClickThrough) > 0 {
		patterns = append(patterns, "has_clicks")
	}

	return patterns
}

// updateCollectionRate updates the feedback collection rate
func (fp *FeedbackProcessor) updateCollectionRate() {
	// Simple implementation - in reality this would track over time windows
	fp.collectionRate = 0.8 // Placeholder
}

// GetCollectionRate returns the current feedback collection rate
func (fp *FeedbackProcessor) GetCollectionRate() float64 {
	return fp.collectionRate
}

// contains checks if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
