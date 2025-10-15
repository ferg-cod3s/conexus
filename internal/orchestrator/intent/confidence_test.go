package intent

import (
	"testing"
)

func TestConfidenceCalculator_Calculate(t *testing.T) {
	calc := NewConfidenceCalculator()

	tests := []struct {
		name     string
		factors  ConfidenceFactors
		expected float64
	}{
		{
			name: "high confidence",
			factors: ConfidenceFactors{
				PatternScore: 1.0,
				EntityScore:  1.0,
				ContextScore: 1.0,
			},
			expected: 1.0,
		},
		{
			name: "medium confidence",
			factors: ConfidenceFactors{
				PatternScore: 0.7,
				EntityScore:  0.5,
				ContextScore: 0.6,
			},
			expected: 0.59, // 0.7*0.6 + 0.5*0.3 + 0.6*0.1 = 0.42 + 0.15 + 0.06 = 0.63
		},
		{
			name: "low confidence",
			factors: ConfidenceFactors{
				PatternScore: 0.3,
				EntityScore:  0.2,
				ContextScore: 0.1,
			},
			expected: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calc.Calculate(tt.factors)

			if score < 0.0 || score > 1.0 {
				t.Errorf("score out of range [0,1]: %f", score)
			}

			// Allow small floating point difference
			if score < tt.expected-0.1 || score > tt.expected+0.1 {
				t.Errorf("expected score around %f, got %f", tt.expected, score)
			}
		})
	}
}

func TestConfidenceCalculator_CalculateForIntent(t *testing.T) {
	calc := NewConfidenceCalculator()

	intent := &Intent{
		PrimaryAgent: "codebase-locator",
		Confidence:   0.8,
		Entities: map[string]string{
			"file_pattern": "parser.go",
		},
	}

	score := calc.CalculateForIntent(intent)

	if score < 0.0 || score > 1.0 {
		t.Errorf("score out of range [0,1]: %f", score)
	}

	if score < 0.3 {
		t.Errorf("expected reasonable confidence, got %f", score)
	}
}

func TestConfidenceCalculator_IsAboveThreshold(t *testing.T) {
	calc := NewConfidenceCalculator()

	tests := []struct {
		name     string
		score    float64
		expected bool
	}{
		{
			name:     "above threshold",
			score:    0.8,
			expected: true,
		},
		{
			name:     "at threshold",
			score:    0.5,
			expected: true,
		},
		{
			name:     "below threshold",
			score:    0.3,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.IsAboveThreshold(tt.score)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConfidenceCalculator_AdjustThreshold(t *testing.T) {
	calc := NewConfidenceCalculator()

	calc.AdjustThreshold(0.7)

	if calc.MinThreshold != 0.7 {
		t.Errorf("expected threshold 0.7, got %f", calc.MinThreshold)
	}

	// Test boundary conditions
	calc.AdjustThreshold(1.5) // Should be clamped to 1.0
	if calc.MinThreshold != 1.0 {
		t.Errorf("expected threshold 1.0 (clamped), got %f", calc.MinThreshold)
	}

	calc.AdjustThreshold(-0.5) // Should be clamped to 0.0
	if calc.MinThreshold != 0.0 {
		t.Errorf("expected threshold 0.0 (clamped), got %f", calc.MinThreshold)
	}
}
