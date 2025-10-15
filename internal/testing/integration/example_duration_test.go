package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssertMaxDuration_WithMockWorkflow demonstrates usage with simulated results
func TestAssertMaxDuration_WithMockWorkflow(t *testing.T) {
	// Simulate a completed workflow with 150ms duration
	result := &TestResult{
		TestName: "Mock Fast Test",
		Passed:   true,
		Duration: 150 * time.Millisecond,
	}

	// Test case 1: Should pass with generous budget
	err := result.AssertMaxDuration(200 * time.Millisecond)
	assert.NoError(t, err, "should pass with 200ms budget")

	// Test case 2: Should fail with tight budget
	err = result.AssertMaxDuration(100 * time.Millisecond)
	require.Error(t, err, "should fail with 100ms budget")
	assert.Contains(t, err.Error(), "150ms", "error should show actual duration")
	assert.Contains(t, err.Error(), "100ms", "error should show max allowed")
	assert.Contains(t, err.Error(), "50ms", "error should show exceeded amount")
}

// TestAssertMaxDuration_PerformanceRegression demonstrates catching regressions
func TestAssertMaxDuration_PerformanceRegression(t *testing.T) {
	tests := []struct {
		name        string
		duration    time.Duration
		maxAllowed  time.Duration
		shouldPass  bool
		description string
	}{
		{
			name:        "baseline-performance",
			duration:    100 * time.Millisecond,
			maxAllowed:  150 * time.Millisecond,
			shouldPass:  true,
			description: "Normal operation within bounds",
		},
		{
			name:        "performance-regression",
			duration:    250 * time.Millisecond,
			maxAllowed:  150 * time.Millisecond,
			shouldPass:  false,
			description: "Detected 100ms regression",
		},
		{
			name:        "performance-improvement",
			duration:    50 * time.Millisecond,
			maxAllowed:  150 * time.Millisecond,
			shouldPass:  true,
			description: "Improved by 50ms",
		},
		{
			name:        "slow-integration-test",
			duration:    15 * time.Second,
			maxAllowed:  30 * time.Second,
			shouldPass:  true,
			description: "Long-running integration test within limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &TestResult{
				TestName: tt.name,
				Duration: tt.duration,
				Passed:   true,
			}

			err := result.AssertMaxDuration(tt.maxAllowed)

			if tt.shouldPass {
				assert.NoError(t, err, tt.description)
			} else {
				require.Error(t, err, tt.description)
				t.Logf("Performance regression detected: %v", err)
			}
		})
	}
}

// TestAssertMaxDuration_UsagePattern shows typical usage pattern
func TestAssertMaxDuration_UsagePattern(t *testing.T) {
	// This demonstrates how developers would typically use AssertMaxDuration
	// after running a test workflow
	
	// Step 1: Run a test (simulated here)
	result := &TestResult{
		TestName: "API Response Time Test",
		Duration: 85 * time.Millisecond,
		Passed:   true,
	}
	
	// Step 2: Verify functional correctness
	require.True(t, result.Passed, "test must pass functionally")
	
	// Step 3: Verify performance requirement (SLA: 100ms)
	err := result.AssertMaxDuration(100 * time.Millisecond)
	assert.NoError(t, err, "API response should meet 100ms SLA")
	
	// If the test passed both functional and performance checks,
	// we have high confidence in the implementation
	t.Logf("✓ Test passed in %v (under 100ms budget)", result.Duration)
}
