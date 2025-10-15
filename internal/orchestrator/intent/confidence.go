package intent

import (
	"math"
)

// ConfidenceCalculator calculates confidence scores for intent matches
type ConfidenceCalculator struct {
	// Minimum confidence threshold (0.0 to 1.0)
	MinThreshold float64

	// Weights for different confidence factors
	PatternMatchWeight float64
	EntityMatchWeight  float64
	ContextWeight      float64
}

// NewConfidenceCalculator creates a new confidence calculator with default weights
func NewConfidenceCalculator() *ConfidenceCalculator {
	return &ConfidenceCalculator{
		MinThreshold:       0.5,
		PatternMatchWeight: 0.6,
		EntityMatchWeight:  0.3,
		ContextWeight:      0.1,
	}
}

// ConfidenceFactors represents factors contributing to confidence score
type ConfidenceFactors struct {
	// Pattern match score (0.0 to 1.0)
	PatternScore float64

	// Entity extraction success rate (0.0 to 1.0)
	EntityScore float64

	// Context relevance score (0.0 to 1.0)
	ContextScore float64
}

// Calculate computes the overall confidence score
func (c *ConfidenceCalculator) Calculate(factors ConfidenceFactors) float64 {
	score := (factors.PatternScore * c.PatternMatchWeight) +
		(factors.EntityScore * c.EntityMatchWeight) +
		(factors.ContextScore * c.ContextWeight)

	// Normalize to 0.0-1.0 range
	score = math.Min(1.0, math.Max(0.0, score))

	return score
}

// CalculateForIntent computes confidence score for a parsed intent
func (c *ConfidenceCalculator) CalculateForIntent(intent *Intent) float64 {
	factors := ConfidenceFactors{
		PatternScore: intent.Confidence,
		EntityScore:  c.calculateEntityScore(intent),
		ContextScore: 0.5, // Default context score (no conversation history yet)
	}

	return c.Calculate(factors)
}

// calculateEntityScore computes entity extraction success rate
func (c *ConfidenceCalculator) calculateEntityScore(intent *Intent) float64 {
	if len(intent.Entities) == 0 {
		return 0.0
	}

	// Expected entities based on agent type
	expectedEntities := c.getExpectedEntities(intent.PrimaryAgent)
	if len(expectedEntities) == 0 {
		return 1.0 // No specific entities required
	}

	matchCount := 0
	for _, expected := range expectedEntities {
		if _, ok := intent.Entities[expected]; ok {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(expectedEntities))
}

// getExpectedEntities returns expected entities for an agent type
func (c *ConfidenceCalculator) getExpectedEntities(agentType string) []string {
	switch agentType {
	case "codebase-locator":
		return []string{"file_pattern", "glob_pattern", "directory"}
	case "codebase-analyzer":
		return []string{"file_pattern", "symbol"}
	case "codebase-pattern-finder":
		return []string{"file_pattern", "symbol"}
	default:
		return []string{}
	}
}

// IsAboveThreshold checks if a confidence score meets the minimum threshold
func (c *ConfidenceCalculator) IsAboveThreshold(score float64) bool {
	return score >= c.MinThreshold
}

// AdjustThreshold dynamically adjusts the confidence threshold
func (c *ConfidenceCalculator) AdjustThreshold(newThreshold float64) {
	c.MinThreshold = math.Min(1.0, math.Max(0.0, newThreshold))
}
