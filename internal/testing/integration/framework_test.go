package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ferg-cod3s/conexus/internal/orchestrator/workflow"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestTestResult_AssertMaxDuration(t *testing.T) {
	tests := []struct {
		name          string
		duration      time.Duration
		maxDuration   time.Duration
		expectError   bool
		errorContains string
	}{
		{
			name:        "duration under threshold",
			duration:    100 * time.Millisecond,
			maxDuration: 200 * time.Millisecond,
			expectError: false,
		},
		{
			name:        "duration exactly at threshold",
			duration:    200 * time.Millisecond,
			maxDuration: 200 * time.Millisecond,
			expectError: false,
		},
		{
			name:          "duration over threshold",
			duration:      300 * time.Millisecond,
			maxDuration:   200 * time.Millisecond,
			expectError:   true,
			errorContains: "exceeded maximum allowed",
		},
		{
			name:          "duration significantly over threshold",
			duration:      5 * time.Second,
			maxDuration:   1 * time.Second,
			expectError:   true,
			errorContains: "by 4s",
		},
		{
			name:        "zero duration with zero threshold",
			duration:    0,
			maxDuration: 0,
			expectError: false,
		},
		{
			name:          "zero threshold with positive duration",
			duration:      100 * time.Millisecond,
			maxDuration:   0,
			expectError:   true,
			errorContains: "exceeded maximum allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			result := &TestResult{
				TestName: "test-case",
				Duration: tt.duration,
			}

			// Act
			err := result.AssertMaxDuration(tt.maxDuration)

			// Assert
			if tt.expectError {
				require.Error(t, err, "expected error but got none")
				assert.Contains(t, err.Error(), tt.errorContains,
					"error message should contain expected text")

				// Verify error message contains actual duration
				assert.Contains(t, err.Error(), tt.duration.String(),
					"error message should contain actual duration")

				// Verify error message contains max duration
				assert.Contains(t, err.Error(), tt.maxDuration.String(),
					"error message should contain max duration")
			} else {
				assert.NoError(t, err, "expected no error but got: %v", err)
			}
		})
	}
}

func TestTestResult_AssertMaxDuration_ErrorMessage(t *testing.T) {
	// Arrange
	result := &TestResult{
		TestName: "slow-test",
		Duration: 5 * time.Second,
	}
	maxDuration := 2 * time.Second

	// Act
	err := result.AssertMaxDuration(maxDuration)

	// Assert
	require.Error(t, err)

	// Verify error message format
	expectedSubstrings := []string{
		"5s",       // actual duration
		"2s",       // max allowed
		"3s",       // exceeded by
		"exceeded", // action word
		"maximum",  // threshold descriptor
	}

	for _, substring := range expectedSubstrings {
		assert.Contains(t, err.Error(), substring,
			"error message should contain '%s'", substring)
	}
}

func TestTestResult_AssertMaxDuration_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name        string
		duration    time.Duration
		maxDuration time.Duration
		scenario    string
	}{
		{
			name:        "quick unit test",
			duration:    50 * time.Millisecond,
			maxDuration: 100 * time.Millisecond,
			scenario:    "fast test under budget",
		},
		{
			name:        "slow integration test",
			duration:    15 * time.Second,
			maxDuration: 10 * time.Second,
			scenario:    "integration test timeout",
		},
		{
			name:        "acceptable e2e test",
			duration:    25 * time.Second,
			maxDuration: 30 * time.Second,
			scenario:    "e2e test within limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &TestResult{
				TestName: tt.name,
				Duration: tt.duration,
			}

			err := result.AssertMaxDuration(tt.maxDuration)

			if tt.duration > tt.maxDuration {
				assert.Error(t, err, "scenario: %s", tt.scenario)
			} else {
				assert.NoError(t, err, "scenario: %s", tt.scenario)
			}
		})
	}
}

// Mock Agent for multi-step workflow tests
type mockMultiStepAgent struct {
	name           string
	shouldFail     bool
	shouldEscalate bool
	escalateTo     string
	delay          time.Duration
}

func (m *mockMultiStepAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	// Simulate processing delay
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return schema.AgentResponse{}, ctx.Err()
		}
	}

	if m.shouldFail {
		return schema.AgentResponse{}, fmt.Errorf("agent %s failed", m.name)
	}

	resp := schema.AgentResponse{
		Output: &schema.AgentOutputV1{
			ComponentName:    "test-component",
			ScopeDescription: fmt.Sprintf("processed by %s", m.name),
		},
	}

	if m.shouldEscalate {
		resp.Escalation = &schema.Escalation{
			Required:    true,
			TargetAgent: m.escalateTo,
			Reason:      fmt.Sprintf("escalated from %s", m.name),
		}
	}

	return resp, nil
}

func TestRunMultiStepWorkflow_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Register test agents
	framework := NewTestFramework()

	framework.RegisterAgent("locator", &mockMultiStepAgent{name: "locator"})
	framework.RegisterAgent("analyzer", &mockMultiStepAgent{name: "analyzer"})
	framework.RegisterAgent("implementer", &mockMultiStepAgent{name: "implementer"})

	config := MultiStepWorkflowConfig{
		ID:          "test-success",
		Description: "Three-step successful workflow",
		Steps: []WorkflowStep{
			{
				ID:    "locate",
				Agent: "locator",
				Input: map[string]interface{}{"query": "find files"},
			},
			{
				ID:    "analyze",
				Agent: "analyzer",
				Input: map[string]interface{}{"query": "analyze code"},
			},
			{
				ID:    "implement",
				Agent: "implementer",
				Input: map[string]interface{}{"query": "implement feature"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed, "workflow should pass")
	assert.Equal(t, "test-success", result.TestName)
	assert.Len(t, result.Errors, 0, "should have no errors")
	assert.Len(t, result.Warnings, 0, "should have no warnings")
	assert.NotNil(t, result.WorkflowResult)
	assert.Len(t, result.WorkflowResult.StepResults, 3, "should have 3 step results")

	// Verify all steps completed
	for i, stepResult := range result.WorkflowResult.StepResults {
		assert.Equal(t, workflow.StepStatusCompleted, stepResult.Status,
			"step %d should be completed", i)
		assert.NotNil(t, stepResult.Output, "step %d should have output", i)
	}
}

func TestRunMultiStepWorkflow_Timeout(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Register slow agent
	framework := NewTestFramework()

	framework.RegisterAgent("slow-agent", &mockMultiStepAgent{
		name:  "slow-agent",
		delay: 3 * time.Second,
	})

	config := MultiStepWorkflowConfig{
		ID:          "test-timeout",
		Description: "Workflow that exceeds timeout",
		Steps: []WorkflowStep{
			{
				Agent: "slow-agent",
				Input: map[string]interface{}{"query": "slow operation"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 500 * time.Millisecond, // Shorter than agent delay
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed, "workflow should not pass")
	assert.Contains(t, result.Errors[0].Error(), "timeout exceeded")
}

func TestRunMultiStepWorkflow_ContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())

	// Register agent that checks cancellation
	framework := NewTestFramework()

	framework.RegisterAgent("cancellable", &mockMultiStepAgent{
		name:  "cancellable",
		delay: 2 * time.Second,
	})

	config := MultiStepWorkflowConfig{
		ID:          "test-cancel",
		Description: "Workflow that gets cancelled",
		Steps: []WorkflowStep{
			{
				Agent: "cancellable",
				Input: map[string]interface{}{"query": "long operation"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	}

	// Cancel context after 200ms
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed, "workflow should not pass")
	assert.Contains(t, result.Errors[0].Error(), "cancelled")
}

func TestRunMultiStepWorkflow_StepFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()

	framework := NewTestFramework()

	framework.RegisterAgent("succeeds", &mockMultiStepAgent{name: "succeeds"})
	framework.RegisterAgent("fails", &mockMultiStepAgent{
		name:       "fails",
		shouldFail: true,
	})
	framework.RegisterAgent("never-runs", &mockMultiStepAgent{name: "never-runs"})

	config := MultiStepWorkflowConfig{
		ID:          "test-failure",
		Description: "Workflow with failing step",
		Steps: []WorkflowStep{
			{
				Agent: "succeeds",
				Input: map[string]interface{}{"query": "step 1"},
			},
			{
				Agent: "fails",
				Input: map[string]interface{}{"query": "step 2"},
			},
			{
				Agent: "never-runs",
				Input: map[string]interface{}{"query": "step 3"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed, "workflow should not pass")
	assert.Greater(t, len(result.Errors), 0, "should have errors")
	assert.Contains(t, result.Errors[0].Error(), "execution failed")
}

func TestRunMultiStepWorkflow_Escalation(t *testing.T) {
	// Arrange
	ctx := context.Background()

	framework := NewTestFramework()

	framework.RegisterAgent("basic-agent", &mockMultiStepAgent{
		name:           "basic-agent",
		shouldEscalate: true,
		escalateTo:     "expert-agent",
	})
	framework.RegisterAgent("expert-agent", &mockMultiStepAgent{name: "expert-agent"})

	config := MultiStepWorkflowConfig{
		ID:          "test-escalation",
		Description: "Workflow with escalation",
		Steps: []WorkflowStep{
			{
				Agent: "basic-agent",
				Input: map[string]interface{}{"query": "complex task"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, len(result.Warnings), 0, "should have escalation warnings")
	assert.Contains(t, result.Warnings[0], "escalated")
	assert.Contains(t, result.Warnings[0], "expert-agent")

	// Verify escalation happened
	assert.Greater(t, len(result.WorkflowResult.StepResults), 1,
		"should have original step + escalated step")
}

func TestRunMultiStepWorkflow_EmptySteps(t *testing.T) {
	// Arrange
	ctx := context.Background()
	framework := NewTestFramework()

	config := MultiStepWorkflowConfig{
		ID:          "test-empty",
		Description: "Workflow with no steps",
		Steps:       []WorkflowStep{},
		Mode:        workflow.SequentialMode,
		Timeout:     5 * time.Second,
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "at least one step")
}

func TestRunMultiStepWorkflow_InputMarshalError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	framework := NewTestFramework()

	framework.RegisterAgent("test-agent", &mockMultiStepAgent{name: "test-agent"})

	// Create input that cannot be marshaled
	invalidInput := map[string]interface{}{
		"channel": make(chan int), // channels cannot be marshaled to JSON
	}

	config := MultiStepWorkflowConfig{
		ID:          "test-marshal-error",
		Description: "Workflow with unmarshalable input",
		Steps: []WorkflowStep{
			{
				Agent: "test-agent",
				Input: invalidInput,
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Greater(t, len(result.Errors), 0)
	assert.Contains(t, result.Errors[0].Error(), "marshal")
}

func TestRunMultiStepWorkflow_ParallelMode(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Register multiple agents
	framework := NewTestFramework()

	framework.RegisterAgent("agent1", &mockMultiStepAgent{
		name:  "agent1",
		delay: 100 * time.Millisecond,
	})
	framework.RegisterAgent("agent2", &mockMultiStepAgent{
		name:  "agent2",
		delay: 100 * time.Millisecond,
	})
	framework.RegisterAgent("agent3", &mockMultiStepAgent{
		name:  "agent3",
		delay: 100 * time.Millisecond,
	})

	config := MultiStepWorkflowConfig{
		ID:          "test-parallel",
		Description: "Parallel workflow execution",
		Steps: []WorkflowStep{
			{Agent: "agent1", Input: map[string]interface{}{"query": "task1"}},
			{Agent: "agent2", Input: map[string]interface{}{"query": "task2"}},
			{Agent: "agent3", Input: map[string]interface{}{"query": "task3"}},
		},
		Mode:    workflow.ParallelMode,
		Timeout: 5 * time.Second,
	}

	// Act
	start := time.Now()
	result, err := framework.RunMultiStepWorkflow(ctx, config)
	duration := time.Since(start)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)

	// Parallel execution should be faster than sequential
	// (3 * 100ms sequentially = 300ms, parallel should be ~100-150ms)
	assert.Less(t, duration, 250*time.Millisecond,
		"parallel execution should be faster than sequential")

	assert.Len(t, result.WorkflowResult.StepResults, 3)
}

func TestRunMultiStepWorkflow_ConditionalMode(t *testing.T) {
	// Arrange
	ctx := context.Background()

	framework := NewTestFramework()

	framework.RegisterAgent("first-agent", &mockMultiStepAgent{name: "first-agent"})
	framework.RegisterAgent("conditional-agent", &mockMultiStepAgent{name: "conditional-agent"})

	config := MultiStepWorkflowConfig{
		ID:          "test-conditional",
		Description: "Conditional workflow execution",
		Steps: []WorkflowStep{
			{
				ID:    "first",
				Agent: "first-agent",
				Input: map[string]interface{}{"query": "first task"},
			},
			{
				ID:    "conditional",
				Agent: "conditional-agent",
				Input: map[string]interface{}{"query": "conditional task"},
				Condition: &workflow.PreviousStepSuccessCondition{
					StepID: "first",
				},
			},
		},
		Mode:    workflow.ConditionalMode,
		Timeout: 5 * time.Second,
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.NotNil(t, result.WorkflowResult)
}

func TestRunMultiStepWorkflow_DefaultTimeout(t *testing.T) {
	// Arrange
	ctx := context.Background()
	framework := NewTestFramework()

	framework.RegisterAgent("test-agent", &mockMultiStepAgent{name: "test-agent"})

	config := MultiStepWorkflowConfig{
		ID:          "test-default-timeout",
		Description: "Workflow with default timeout",
		Steps: []WorkflowStep{
			{
				Agent: "test-agent",
				Input: map[string]interface{}{"query": "task"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 0, // Should default to 30 seconds
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert - should succeed with default timeout
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
}

func TestRunMultiStepWorkflow_AutoGeneratedStepIDs(t *testing.T) {
	// Arrange
	ctx := context.Background()
	framework := NewTestFramework()

	framework.RegisterAgent("agent1", &mockMultiStepAgent{name: "agent1"})
	framework.RegisterAgent("agent2", &mockMultiStepAgent{name: "agent2"})

	config := MultiStepWorkflowConfig{
		ID:          "test-auto-ids",
		Description: "Steps without explicit IDs",
		Steps: []WorkflowStep{
			{
				// No ID provided - should auto-generate
				Agent: "agent1",
				Input: map[string]interface{}{"query": "task1"},
			},
			{
				// No ID provided - should auto-generate
				Agent: "agent2",
				Input: map[string]interface{}{"query": "task2"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	}

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.Len(t, result.WorkflowResult.StepResults, 2)

	// Verify step IDs were auto-generated
	assert.NotEmpty(t, result.WorkflowResult.StepResults[0].StepID)
	assert.NotEmpty(t, result.WorkflowResult.StepResults[1].StepID)
}

func TestRunMultiStepWorkflow_ResultStoredInFramework(t *testing.T) {
	// Arrange
	ctx := context.Background()
	framework := NewTestFramework()

	framework.RegisterAgent("test-agent", &mockMultiStepAgent{name: "test-agent"})

	config := MultiStepWorkflowConfig{
		ID:          "test-result-storage",
		Description: "Verify result is stored",
		Steps: []WorkflowStep{
			{
				Agent: "test-agent",
				Input: map[string]interface{}{"query": "task"},
			},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	}

	initialResultCount := len(framework.results)

	// Act
	result, err := framework.RunMultiStepWorkflow(ctx, config)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, framework.results, initialResultCount+1,
		"result should be stored in framework")
	assert.Equal(t, result, framework.results[len(framework.results)-1],
		"stored result should match returned result")
}
