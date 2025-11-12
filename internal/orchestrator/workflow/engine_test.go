package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// mockAgent implements the Agent interface for testing
type mockAgent struct {
	name           string
	shouldFail     bool
	shouldEscalate bool
}

func (m *mockAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	if m.shouldFail {
		return schema.AgentResponse{}, fmt.Errorf("mock agent failed")
	}

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: m.name,
		Overview:      "Mock agent output",
	}

	resp := schema.AgentResponse{
		Output: output,
	}

	if m.shouldEscalate {
		resp.Escalation = &schema.Escalation{
			Required:    true,
			TargetAgent: "escalated-agent",
			Reason:      "test escalation",
		}
	}

	return resp, nil
}

func TestEngine_ExecuteSequential(t *testing.T) {
	executor := NewAgentExecutor()
	executor.RegisterAgent("agent1", &mockAgent{name: "agent1"})
	executor.RegisterAgent("agent2", &mockAgent{name: "agent2"})

	engine := NewEngine(executor)

	workflow := &Workflow{
		ID:   "test-workflow",
		Mode: SequentialMode,
		Steps: []*Step{
			{
				ID:    "step1",
				Agent: "agent1",
				Input: "test input 1",
			},
			{
				ID:    "step2",
				Agent: "agent2",
				Input: "test input 2",
			},
		},
	}

	result, err := engine.Execute(context.Background(), workflow)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, result.Status)
	}

	if len(result.StepResults) != 2 {
		t.Errorf("expected 2 step results, got %d", len(result.StepResults))
	}

	for i, stepResult := range result.StepResults {
		if stepResult.Status != StepStatusCompleted {
			t.Errorf("step %d: expected status %s, got %s", i, StepStatusCompleted, stepResult.Status)
		}
	}
}

func TestEngine_ExecuteParallel(t *testing.T) {
	executor := NewAgentExecutor()
	executor.RegisterAgent("agent1", &mockAgent{name: "agent1"})
	executor.RegisterAgent("agent2", &mockAgent{name: "agent2"})

	engine := NewEngine(executor)

	workflow := &Workflow{
		ID:   "test-workflow",
		Mode: ParallelMode,
		Steps: []*Step{
			{
				ID:    "step1",
				Agent: "agent1",
				Input: "test input 1",
			},
			{
				ID:    "step2",
				Agent: "agent2",
				Input: "test input 2",
			},
		},
	}

	result, err := engine.Execute(context.Background(), workflow)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, result.Status)
	}

	if len(result.StepResults) != 2 {
		t.Errorf("expected 2 step results, got %d", len(result.StepResults))
	}
}

func TestEngine_ExecuteConditional(t *testing.T) {
	executor := NewAgentExecutor()
	executor.RegisterAgent("agent1", &mockAgent{name: "agent1"})
	executor.RegisterAgent("agent2", &mockAgent{name: "agent2"})

	engine := NewEngine(executor)

	workflow := &Workflow{
		ID:   "test-workflow",
		Mode: ConditionalMode,
		Steps: []*Step{
			{
				ID:    "step1",
				Agent: "agent1",
				Input: "test input 1",
			},
			{
				ID:    "step2",
				Agent: "agent2",
				Input: "test input 2",
				Condition: &PreviousStepSuccessCondition{
					StepID: "step1",
				},
			},
		},
	}

	result, err := engine.Execute(context.Background(), workflow)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, result.Status)
	}

	if len(result.StepResults) != 2 {
		t.Errorf("expected 2 step results, got %d", len(result.StepResults))
	}

	// Both steps should execute since step1 succeeds
	for i, stepResult := range result.StepResults {
		if stepResult.Status != StepStatusCompleted {
			t.Errorf("step %d: expected status %s, got %s", i, StepStatusCompleted, stepResult.Status)
		}
	}
}

func TestEngine_ExecuteWithFailure(t *testing.T) {
	executor := NewAgentExecutor()
	executor.RegisterAgent("failing-agent", &mockAgent{name: "failing-agent", shouldFail: true})

	engine := NewEngine(executor)

	workflow := &Workflow{
		ID:   "test-workflow",
		Mode: SequentialMode,
		Steps: []*Step{
			{
				ID:    "step1",
				Agent: "failing-agent",
				Input: "test input",
			},
		},
	}

	result, err := engine.Execute(context.Background(), workflow)
	if err == nil {
		t.Error("expected error, got nil")
		return
	}

	if result.Status != StatusFailed {
		t.Errorf("expected status %s, got %s", StatusFailed, result.Status)
	}
}

func TestEngine_ExecuteWithEscalation(t *testing.T) {
	executor := NewAgentExecutor()
	executor.RegisterAgent("escalating-agent", &mockAgent{name: "escalating-agent", shouldEscalate: true})
	executor.RegisterAgent("escalated-agent", &mockAgent{name: "escalated-agent"})

	engine := NewEngine(executor)

	workflow := &Workflow{
		ID:   "test-workflow",
		Mode: SequentialMode,
		Steps: []*Step{
			{
				ID:    "step1",
				Agent: "escalating-agent",
				Input: "test input",
			},
		},
	}

	result, err := engine.Execute(context.Background(), workflow)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if result.Status != StatusCompleted {
		t.Errorf("expected status %s, got %s", StatusCompleted, result.Status)
	}

	// Should have original step + escalated step
	if len(result.StepResults) < 2 {
		t.Errorf("expected at least 2 step results (including escalation), got %d", len(result.StepResults))
	}

	// First step should be escalated
	if result.StepResults[0].Status != StepStatusEscalated {
		t.Errorf("first step should be escalated, got status %s", result.StepResults[0].Status)
	}
}

func TestValidator_Validate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name        string
		workflow    *Workflow
		shouldError bool
	}{
		{
			name: "valid workflow",
			workflow: &Workflow{
				ID:   "test",
				Mode: SequentialMode,
				Steps: []*Step{
					{ID: "step1", Agent: "agent1", Input: "input1"},
				},
			},
			shouldError: false,
		},
		{
			name:        "nil workflow",
			workflow:    nil,
			shouldError: true,
		},
		{
			name: "empty ID",
			workflow: &Workflow{
				ID:   "",
				Mode: SequentialMode,
				Steps: []*Step{
					{ID: "step1", Agent: "agent1", Input: "input1"},
				},
			},
			shouldError: true,
		},
		{
			name: "no steps",
			workflow: &Workflow{
				ID:    "test",
				Mode:  SequentialMode,
				Steps: []*Step{},
			},
			shouldError: true,
		},
		{
			name: "duplicate step IDs",
			workflow: &Workflow{
				ID:   "test",
				Mode: SequentialMode,
				Steps: []*Step{
					{ID: "step1", Agent: "agent1", Input: "input1"},
					{ID: "step1", Agent: "agent2", Input: "input2"},
				},
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.workflow)

			if tt.shouldError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	builder := NewBuilder("test-workflow")

	workflow, err := builder.
		WithDescription("Test workflow").
		WithMode(SequentialMode).
		AddSequentialStep("step1", "agent1", "input1", schema.Permissions{}).
		AddSequentialStep("step2", "agent2", "input2", schema.Permissions{}).
		Build()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if workflow.ID != "test-workflow" {
		t.Errorf("expected ID 'test-workflow', got %s", workflow.ID)
	}

	if workflow.Description != "Test workflow" {
		t.Errorf("expected description 'Test workflow', got %s", workflow.Description)
	}

	if len(workflow.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(workflow.Steps))
	}
}

func TestBuilder_BuildWithoutSteps(t *testing.T) {
	builder := NewBuilder("test-workflow")

	_, err := builder.Build()

	if err == nil {
		t.Error("expected error for workflow without steps")
	}
}
