package workflow

import (
	"fmt"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Workflow represents a multi-agent workflow
type Workflow struct {
	// Unique workflow identifier
	ID string

	// Workflow description
	Description string

	// Execution mode (sequential, parallel, conditional)
	Mode ExecutionMode

	// Workflow steps
	Steps []*Step

	// Workflow metadata
	Metadata map[string]interface{}
}

// Step represents a single step in a workflow
type Step struct {
	// Unique step identifier
	ID string

	// Agent to execute this step
	Agent string

	// Input for this step
	Input string

	// Permissions for this step
	Permissions schema.Permissions

	// Optional condition for conditional workflows
	Condition Condition

	// Dependencies on other steps (for parallel workflows)
	Dependencies []string

	// Step metadata
	Metadata map[string]interface{}
}

// Condition defines when a step should execute
type Condition interface {
	// Evaluate returns true if the step should execute
	Evaluate(result *ExecutionResult) bool
}

// PreviousStepSuccessCondition checks if a previous step succeeded
type PreviousStepSuccessCondition struct {
	StepID string
}

// Evaluate implements Condition.Evaluate
func (c *PreviousStepSuccessCondition) Evaluate(result *ExecutionResult) bool {
	if result == nil {
		return false
	}

	for _, stepResult := range result.StepResults {
		if stepResult.StepID == c.StepID {
			return stepResult.Status == StepStatusCompleted
		}
	}

	return false
}

// OutputContainsCondition checks if a previous step's output contains specific data
type OutputContainsCondition struct {
	StepID string
	Field  string
	Value  string
}

// Evaluate implements Condition.Evaluate
func (c *OutputContainsCondition) Evaluate(result *ExecutionResult) bool {
	if result == nil {
		return false
	}

	for _, stepResult := range result.StepResults {
		if stepResult.StepID == c.StepID && stepResult.Output != nil {
			// Simple field check (can be enhanced)
			switch c.Field {
			case "component_name":
				return stepResult.Output.ComponentName == c.Value
			case "scope_description":
				return stepResult.Output.ScopeDescription == c.Value
			}
		}
	}

	return false
}

// Builder provides a fluent interface for constructing workflows
type Builder struct {
	workflow *Workflow
}

// NewBuilder creates a new workflow builder
func NewBuilder(id string) *Builder {
	return &Builder{
		workflow: &Workflow{
			ID:       id,
			Steps:    make([]*Step, 0),
			Mode:     SequentialMode,
			Metadata: make(map[string]interface{}),
		},
	}
}

// WithDescription sets the workflow description
func (b *Builder) WithDescription(desc string) *Builder {
	b.workflow.Description = desc
	return b
}

// WithMode sets the execution mode
func (b *Builder) WithMode(mode ExecutionMode) *Builder {
	b.workflow.Mode = mode
	return b
}

// AddStep adds a step to the workflow
func (b *Builder) AddStep(step *Step) *Builder {
	b.workflow.Steps = append(b.workflow.Steps, step)
	return b
}

// AddSequentialStep adds a step that runs sequentially
func (b *Builder) AddSequentialStep(id, agent, input string, permissions schema.Permissions) *Builder {
	step := &Step{
		ID:          id,
		Agent:       agent,
		Input:       input,
		Permissions: permissions,
		Metadata:    make(map[string]interface{}),
	}
	b.workflow.Steps = append(b.workflow.Steps, step)
	return b
}

// AddConditionalStep adds a step with a condition
func (b *Builder) AddConditionalStep(id, agent, input string, permissions schema.Permissions, condition Condition) *Builder {
	step := &Step{
		ID:          id,
		Agent:       agent,
		Input:       input,
		Permissions: permissions,
		Condition:   condition,
		Metadata:    make(map[string]interface{}),
	}
	b.workflow.Steps = append(b.workflow.Steps, step)
	return b
}

// AddParallelStep adds a step that can run in parallel
func (b *Builder) AddParallelStep(id, agent, input string, permissions schema.Permissions, dependencies []string) *Builder {
	step := &Step{
		ID:           id,
		Agent:        agent,
		Input:        input,
		Permissions:  permissions,
		Dependencies: dependencies,
		Metadata:     make(map[string]interface{}),
	}
	b.workflow.Steps = append(b.workflow.Steps, step)
	return b
}

// WithMetadata sets workflow metadata
func (b *Builder) WithMetadata(key string, value interface{}) *Builder {
	b.workflow.Metadata[key] = value
	return b
}

// Build returns the constructed workflow
func (b *Builder) Build() (*Workflow, error) {
	if b.workflow.ID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	if len(b.workflow.Steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	return b.workflow, nil
}
