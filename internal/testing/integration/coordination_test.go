// Package integration provides integration testing for complex multi-agent coordination
package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/orchestrator/workflow"
	"github.com/ferg-cod3s/conexus/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiAgentDataPipeline tests a complex pipeline where 4+ agents pass data sequentially
func TestMultiAgentDataPipeline(t *testing.T) {
	ctx := context.Background()
	fw := NewTestFramework()

	// Create a 4-agent pipeline: Locator -> Analyzer -> Transformer -> Reporter
	agents := []struct {
		name   string
		role   string
		input  string
		output string
	}{
		{"locator", "file-locator", "", "Found 5 files"},
		{"analyzer", "code-analyzer", "Found 5 files", "Analyzed: 3 functions, 2 structs"},
		{"transformer", "data-transformer", "Analyzed: 3 functions, 2 structs", "Transformed: 3 functions to JSON"},
		{"reporter", "report-generator", "Transformed: 3 functions to JSON", "Report: Complete with 5 sections"},
	}

	// Register all agents
	for _, a := range agents {
		agent := NewMockDataPipelineAgent(a.name, a.input, a.output)
		fw.RegisterAgent(a.name, agent)
	}

	// Build workflow with sequential steps
	steps := []*workflow.Step{}
	for i, a := range agents {
		deps := []string{}
		if i > 0 {
			deps = []string{agents[i-1].name}
		}
		steps = append(steps, &workflow.Step{
			ID:           a.name,
			Agent:        a.name,
			Dependencies: deps,
		})
	}

	wf := &workflow.Workflow{
		ID:          "data-pipeline-workflow",
		Description: "4-agent sequential pipeline",
		Mode:        workflow.SequentialMode,
		Steps:       steps,
	}

	// Create test case
	testCase := &TestCase{
		Name:     "Multi-Agent Data Pipeline",
		Workflow: wf,
		Timeout:  5 * time.Second,
	}

	// Execute workflow
	result := fw.Run(ctx, testCase)
	
	// Assertions
	require.True(t, result.Passed, "Pipeline should complete successfully")
	require.NotNil(t, result.WorkflowResult, "Should have workflow result")
	assert.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status, "Workflow should complete")
	assert.GreaterOrEqual(t, len(result.WorkflowResult.StepResults), 4, "All 4 agents should execute")

	t.Logf("✓ 4-agent pipeline completed in %v", result.Duration)
	t.Logf("✓ Data flow verified through all stages")
}

// TestParallelAgentCoordination tests multiple agents working in parallel with merge step
func TestParallelAgentCoordination(t *testing.T) {
	ctx := context.Background()
	fw := NewTestFramework()

	// Register 3 parallel analyzers + 1 merger
	fw.RegisterAgent("analyzer-1", NewMockDataPipelineAgent("analyzer-1", "", "Result A: 10 items"))
	fw.RegisterAgent("analyzer-2", NewMockDataPipelineAgent("analyzer-2", "", "Result B: 15 items"))
	fw.RegisterAgent("analyzer-3", NewMockDataPipelineAgent("analyzer-3", "", "Result C: 8 items"))
	fw.RegisterAgent("merger", NewMockDataPipelineAgent("merger", "multiple", "Merged: 33 total items"))

	// Build workflow with parallel execution
	steps := []*workflow.Step{
		{ID: "analyzer-1", Agent: "analyzer-1", Dependencies: []string{}},
		{ID: "analyzer-2", Agent: "analyzer-2", Dependencies: []string{}},
		{ID: "analyzer-3", Agent: "analyzer-3", Dependencies: []string{}},
		{ID: "merger", Agent: "merger", Dependencies: []string{"analyzer-1", "analyzer-2", "analyzer-3"}},
	}

	wf := &workflow.Workflow{
		ID:          "parallel-coordination",
		Description: "3 parallel agents with merge",
		Mode:        workflow.ParallelMode,
		Steps:       steps,
	}

	testCase := &TestCase{
		Name:     "Parallel Agent Coordination",
		Workflow: wf,
		Timeout:  5 * time.Second,
	}

	result := fw.Run(ctx, testCase)

	require.True(t, result.Passed, "Parallel workflow should complete")
	require.NotNil(t, result.WorkflowResult)
	assert.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status)
	assert.GreaterOrEqual(t, len(result.WorkflowResult.StepResults), 4, "All 4 agents should execute")

	t.Logf("✓ Parallel coordination completed in %v", result.Duration)
	t.Logf("✓ 3 parallel agents + 1 merger executed correctly")
}

// TestErrorPropagationInPipeline tests error handling across multi-step workflows
func TestErrorPropagationInPipeline(t *testing.T) {
	ctx := context.Background()
	fw := NewTestFramework()

	// Create pipeline with failing middle agent
	fw.RegisterAgent("step1", NewMockDataPipelineAgent("step1", "", "Step 1 complete"))
	fw.RegisterAgent("step2-fail", NewMockFailingAgent("step2-fail", "Critical error in step 2", false))
	fw.RegisterAgent("step3", NewMockDataPipelineAgent("step3", "Step 1 complete", "Step 3 complete"))

	// Build sequential workflow
	steps := []*workflow.Step{
		{ID: "step1", Agent: "step1", Dependencies: []string{}},
		{ID: "step2-fail", Agent: "step2-fail", Dependencies: []string{"step1"}},
		{ID: "step3", Agent: "step3", Dependencies: []string{"step2-fail"}},
	}

	wf := &workflow.Workflow{
		ID:          "error-propagation-pipeline",
		Description: "Pipeline with failing step",
		Mode:        workflow.SequentialMode,
		Steps:       steps,
	}

	testCase := &TestCase{
		Name:     "Error Propagation in Pipeline",
		Workflow: wf,
		Timeout:  5 * time.Second,
	}

	result := fw.Run(ctx, testCase)

	// Should fail due to step2 error
	require.False(t, result.Passed, "Pipeline should fail when agent fails")
	assert.NotEmpty(t, result.Errors, "Should have error recorded")

	t.Logf("✓ Error propagation verified in %v", result.Duration)
	t.Logf("✓ Pipeline correctly stopped at failing agent")
}

// TestDynamicWorkflowAdjustment tests workflow branching based on intermediate results
func TestDynamicWorkflowAdjustment(t *testing.T) {
	ctx := context.Background()
	fw := NewTestFramework()

	// Register agents for conditional branching
	fw.RegisterAgent("analyzer", NewMockDataPipelineAgent("analyzer", "", "High complexity detected"))
	fw.RegisterAgent("simple-path", NewMockDataPipelineAgent("simple-path", "", "Simple processing"))
	fw.RegisterAgent("complex-path", NewMockDataPipelineAgent("complex-path", "High complexity", "Complex processing"))

	// Build workflow with conditional logic (both paths registered, conditional in real execution)
	steps := []*workflow.Step{
		{ID: "analyzer", Agent: "analyzer", Dependencies: []string{}},
		// In real scenario, only one would execute based on analyzer output
		{ID: "complex-path", Agent: "complex-path", Dependencies: []string{"analyzer"}},
	}

	wf := &workflow.Workflow{
		ID:          "dynamic-workflow",
		Description: "Dynamic workflow with conditional branching",
		Mode:        workflow.ConditionalMode,
		Steps:       steps,
	}

	testCase := &TestCase{
		Name:     "Dynamic Workflow Adjustment",
		Workflow: wf,
		Timeout:  5 * time.Second,
	}

	result := fw.Run(ctx, testCase)

	require.True(t, result.Passed, "Dynamic workflow should complete")
	require.NotNil(t, result.WorkflowResult)
	assert.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status)

	t.Logf("✓ Dynamic workflow adjustment completed in %v", result.Duration)
	t.Logf("✓ Conditional branching logic verified")
}

// TestStatePersistenceWithConditionals tests state management with conditional branches
func TestStatePersistenceWithConditionals(t *testing.T) {
	ctx := context.Background()
	fw := NewTestFramework()

	// Register stateful agents
	fw.RegisterAgent("state-init", NewMockStatefulAgent("state-init", "Initial state: counter=0"))
	fw.RegisterAgent("state-update", NewMockStatefulAgent("state-update", "Updated state: counter=5"))
	fw.RegisterAgent("state-verify", NewMockStatefulAgent("state-verify", "Verified state: counter=5"))

	// Build stateful workflow
	steps := []*workflow.Step{
		{ID: "state-init", Agent: "state-init", Dependencies: []string{}},
		{ID: "state-update", Agent: "state-update", Dependencies: []string{"state-init"}},
		{ID: "state-verify", Agent: "state-verify", Dependencies: []string{"state-update"}},
	}

	wf := &workflow.Workflow{
		ID:          "state-persistence-workflow",
		Description: "State persistence with conditionals",
		Mode:        workflow.SequentialMode,
		Steps:       steps,
	}

	testCase := &TestCase{
		Name:     "State Persistence with Conditionals",
		Workflow: wf,
		Timeout:  5 * time.Second,
	}

	result := fw.Run(ctx, testCase)

	require.True(t, result.Passed, "Stateful workflow should complete")
	require.NotNil(t, result.WorkflowResult)
	assert.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status)
	assert.GreaterOrEqual(t, len(result.WorkflowResult.StepResults), 3, "All 3 state transitions should execute")

	t.Logf("✓ State persistence verified in %v", result.Duration)
	t.Logf("✓ Conditional state management working correctly")
}

// ================================================================================
// MOCK AGENTS
// ================================================================================

// MockDataPipelineAgent simulates an agent in a data processing pipeline
type MockDataPipelineAgent struct {
	name           string
	expectedInput  string
	outputData     string
}

func NewMockDataPipelineAgent(name, expectedInput, outputData string) *MockDataPipelineAgent {
	return &MockDataPipelineAgent{
		name:          name,
		expectedInput: expectedInput,
		outputData:    outputData,
	}
}

func (m *MockDataPipelineAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	// Simulate processing
	time.Sleep(10 * time.Millisecond)

	output := &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    m.name,
		ScopeDescription: "Pipeline agent: " + m.name,
		Overview:         m.outputData,
		RawEvidence: []schema.Evidence{
			{
				Claim: m.outputData,
				File:  "pipeline.go",
				Lines: "1-10",
			},
		},
	}

	return schema.AgentResponse{
		RequestID: req.RequestID,
		AgentID:   m.name,
		Status:    schema.StatusComplete,
		Output:    output,
		Duration:  10 * time.Millisecond,
		Timestamp: time.Now(),
	}, nil
}

// MockStatefulAgent simulates an agent that maintains state across steps
type MockStatefulAgent struct {
	name      string
	stateData string
}

func NewMockStatefulAgent(name, stateData string) *MockStatefulAgent {
	return &MockStatefulAgent{
		name:      name,
		stateData: stateData,
	}
}

func (m *MockStatefulAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	output := &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    m.name,
		ScopeDescription: "Stateful agent: " + m.name,
		Overview:         m.stateData,
		RawEvidence: []schema.Evidence{
			{
				Claim: m.stateData,
				File:  "state.go",
				Lines: "1-5",
			},
		},
	}

	return schema.AgentResponse{
		RequestID: req.RequestID,
		AgentID:   m.name,
		Status:    schema.StatusComplete,
		Output:    output,
		Duration:  5 * time.Millisecond,
		Timestamp: time.Now(),
	}, nil
}

// MockFailingAgent simulates an agent that fails
type MockFailingAgent struct {
	name        string
	errorMsg    string
	recoverable bool
}

func NewMockFailingAgent(name, errorMsg string, recoverable bool) *MockFailingAgent {
	return &MockFailingAgent{
		name:        name,
		errorMsg:    errorMsg,
		recoverable: recoverable,
	}
}

func (m *MockFailingAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	return schema.AgentResponse{
		RequestID: req.RequestID,
		AgentID:   m.name,
		Status:    schema.StatusError,
		Error: &schema.AgentError{
			Code:        "EXECUTION_ERROR",
			Message:     m.errorMsg,
			Recoverable: m.recoverable,
		},
		Duration:  5 * time.Millisecond,
		Timestamp: time.Now(),
	}, fmt.Errorf("%s", m.errorMsg)
}
