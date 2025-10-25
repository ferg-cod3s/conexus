// Package workflow provides multi-agent workflow coordination and execution.
//
// The workflow engine supports:
// - Sequential and parallel agent execution
// - Conditional branching
// - Dynamic workflow construction
// - Result aggregation
package workflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// ExecutionMode defines how a workflow should be executed
type ExecutionMode string

const (
	// SequentialMode executes steps one after another
	SequentialMode ExecutionMode = "sequential"

	// ParallelMode executes all steps concurrently
	ParallelMode ExecutionMode = "parallel"

	// ConditionalMode executes steps based on conditions
	ConditionalMode ExecutionMode = "conditional"
)

// Engine orchestrates workflow execution
type Engine struct {
	executor  Executor
	validator *Validator
}

// NewEngine creates a new workflow engine
func NewEngine(executor Executor) *Engine {
	return &Engine{
		executor:  executor,
		validator: NewValidator(),
	}
}

// Execute runs a workflow and returns the aggregated result
func (e *Engine) Execute(ctx context.Context, wf *Workflow) (*ExecutionResult, error) {
	// Validate workflow before execution
	if err := e.validator.Validate(wf); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	switch wf.Mode {
	case SequentialMode:
		return e.executeSequential(ctx, wf)
	case ParallelMode:
		return e.executeParallel(ctx, wf)
	case ConditionalMode:
		return e.executeConditional(ctx, wf)
	default:
		return nil, fmt.Errorf("unknown execution mode: %s", wf.Mode)
	}
}

// executeSequential runs workflow steps one after another
func (e *Engine) executeSequential(ctx context.Context, wf *Workflow) (*ExecutionResult, error) {
	result := &ExecutionResult{
		WorkflowID:  wf.ID,
		StepResults: make([]*StepResult, 0, len(wf.Steps)),
	}

	// Use index-based loop to allow dynamic step addition during escalation
	for i := 0; i < len(wf.Steps); i++ {
		step := wf.Steps[i]
		// Check context cancellation
		select {
		case <-ctx.Done():
			result.Status = StatusCancelled
			return result, ctx.Err()
		default:
		}

		// Execute step
		stepResult, err := e.executor.ExecuteStep(ctx, step, result)
		if err != nil {
			result.Status = StatusFailed
			result.Error = fmt.Sprintf("step %d failed: %v", i, err)
			return result, err
		}

		result.StepResults = append(result.StepResults, stepResult)

		// Handle escalation
		if stepResult.Status == StepStatusEscalated && stepResult.EscalationTarget != "" {
			newStep := &Step{
				ID:    fmt.Sprintf("%s-escalated-%d", step.ID, len(result.StepResults)),
				Agent: stepResult.EscalationTarget,
				Input: stepResult.EscalationReason,
			}
			wf.Steps = append(wf.Steps, newStep)
		}
	}

	result.Status = StatusCompleted
	return result, nil
}

// executeParallel runs all workflow steps concurrently
func (e *Engine) executeParallel(ctx context.Context, wf *Workflow) (*ExecutionResult, error) {
	result := &ExecutionResult{
		WorkflowID:  wf.ID,
		StepResults: make([]*StepResult, len(wf.Steps)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, step := range wf.Steps {
		wg.Add(1)
		go func(idx int, s *Step) {
			defer wg.Done()

			// Create a local copy of the current result for this goroutine
			localResult := &ExecutionResult{
				WorkflowID:  result.WorkflowID,
				StepResults: make([]*StepResult, 0),
			}

			// Execute step with local result to avoid races
			stepResult, err := e.executor.ExecuteStep(ctx, s, localResult)

			mu.Lock()
			result.StepResults[idx] = stepResult
			if err != nil && firstErr == nil {
				firstErr = err
			}
			mu.Unlock()
		}(i, step)
	}

	wg.Wait()

	if firstErr != nil {
		result.Status = StatusFailed
		result.Error = firstErr.Error()
		return result, firstErr
	}

	result.Status = StatusCompleted
	return result, nil
}

// executeConditional runs workflow steps based on conditions
func (e *Engine) executeConditional(ctx context.Context, wf *Workflow) (*ExecutionResult, error) {
	result := &ExecutionResult{
		WorkflowID:  wf.ID,
		StepResults: make([]*StepResult, 0),
	}

	// Use index-based loop to allow dynamic step addition during escalation
	for i := 0; i < len(wf.Steps); i++ {
		step := wf.Steps[i]
		// Check context cancellation
		select {
		case <-ctx.Done():
			result.Status = StatusCancelled
			return result, ctx.Err()
		default:
		}

		// Evaluate condition
		if step.Condition != nil {
			shouldExecute := step.Condition.Evaluate(result)
			if !shouldExecute {
				// Skip this step
				result.StepResults = append(result.StepResults, &StepResult{
					StepID: step.ID,
					Status: StepStatusSkipped,
				})
				continue
			}
		}

		// Execute step
		stepResult, err := e.executor.ExecuteStep(ctx, step, result)
		if err != nil {
			result.Status = StatusFailed
			result.Error = fmt.Sprintf("step %d failed: %v", i, err)
			return result, err
		}

		result.StepResults = append(result.StepResults, stepResult)
	}

	result.Status = StatusCompleted
	return result, nil
}

// ExecutionStatus represents the status of a workflow execution
type ExecutionStatus string

const (
	// StatusPending indicates workflow is waiting to start
	StatusPending ExecutionStatus = "pending"

	// StatusRunning indicates workflow is currently executing
	StatusRunning ExecutionStatus = "running"

	// StatusCompleted indicates workflow finished successfully
	StatusCompleted ExecutionStatus = "completed"

	// StatusFailed indicates workflow encountered an error
	StatusFailed ExecutionStatus = "failed"

	// StatusCancelled indicates workflow was cancelled
	StatusCancelled ExecutionStatus = "cancelled"
)

// ExecutionResult contains the results of a workflow execution
type ExecutionResult struct {
	// Workflow ID
	WorkflowID string

	// Execution status
	Status ExecutionStatus

	// Results from each step
	StepResults []*StepResult

	// Error message if status is Failed
	Error string

	// Aggregated output (optional)
	AggregatedOutput *schema.AgentOutputV1
}

// StepStatus represents the status of a workflow step
type StepStatus string

const (
	// StepStatusPending indicates step is waiting to execute
	StepStatusPending StepStatus = "pending"

	// StepStatusRunning indicates step is currently executing
	StepStatusRunning StepStatus = "running"

	// StepStatusCompleted indicates step finished successfully
	StepStatusCompleted StepStatus = "completed"

	// StepStatusFailed indicates step encountered an error
	StepStatusFailed StepStatus = "failed"

	// StepStatusSkipped indicates step was skipped (conditional workflows)
	StepStatusSkipped StepStatus = "skipped"

	// StepStatusEscalated indicates step requested escalation
	StepStatusEscalated StepStatus = "escalated"
)

// StepResult contains the result of a single workflow step
type StepResult struct {
	// Step ID
	StepID string

	// Agent that executed this step
	Agent string

	// Execution status
	Status StepStatus

	// Agent output
	Output *schema.AgentOutputV1

	// Error message if status is Failed
	Error string

	// Escalation target if status is Escalated
	EscalationTarget string

	// Escalation reason if status is Escalated
	EscalationReason string
}
