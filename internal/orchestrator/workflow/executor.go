package workflow

import (
	"context"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Executor defines the interface for executing workflow steps
type Executor interface {
	// ExecuteStep executes a single workflow step
	ExecuteStep(ctx context.Context, step *Step, currentResult *ExecutionResult) (*StepResult, error)
}

// AgentExecutor executes workflow steps using registered agents
type AgentExecutor struct {
	agents     map[string]Agent
	maxRetries int
}

// Agent defines the interface for agents that can be used in workflows
type Agent interface {
	// Execute runs the agent with the given request
	Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error)
}

// NewAgentExecutor creates a new agent-based executor
func NewAgentExecutor() *AgentExecutor {
	return &AgentExecutor{
		agents:     make(map[string]Agent),
		maxRetries: 3, // Default to 3 retry attempts
	}
}

// RegisterAgent registers an agent for use in workflows
func (e *AgentExecutor) RegisterAgent(name string, agent Agent) {
	e.agents[name] = agent
}

// ExecuteStep implements Executor.ExecuteStep
func (e *AgentExecutor) ExecuteStep(ctx context.Context, step *Step, currentResult *ExecutionResult) (*StepResult, error) {
	agent, ok := e.agents[step.Agent]
	if !ok {
		return &StepResult{
			StepID: step.ID,
			Agent:  step.Agent,
			Status: StepStatusFailed,
			Error:  "agent not found: " + step.Agent,
		}, nil
	}

	// Build agent request
	req := schema.AgentRequest{
		Task: schema.AgentTask{
			SpecificRequest: step.Input,
		},
		Permissions: step.Permissions,
		Context: schema.ConversationContext{
			UserRequest:        step.Input,
			PreviousAgents:     e.getPreviousAgents(currentResult),
			AccumulatedContext: e.buildContext(currentResult),
		},
	}

	// Execute agent with retry logic
	var resp schema.AgentResponse
	var err error

	for attempt := 1; attempt <= e.maxRetries; attempt++ {
		resp, err = agent.Execute(ctx, req)
		
		// Success - no error
		if err == nil {
			break
		}
		
		// Check if error is recoverable
		recoverable := false
		if resp.Error != nil && resp.Error.Recoverable {
			recoverable = true
		}
		
		// If not recoverable or max attempts reached, fail
		if !recoverable || attempt >= e.maxRetries {
			return &StepResult{
				StepID: step.ID,
				Agent:  step.Agent,
				Status: StepStatusFailed,
				Error:  err.Error(),
			}, err
		}
		
		// Exponential backoff before retry
		backoff := time.Duration(10*attempt*attempt) * time.Millisecond
		
		select {
		case <-ctx.Done():
			return &StepResult{
				StepID: step.ID,
				Agent:  step.Agent,
				Status: StepStatusFailed,
				Error:  "context cancelled during retry",
			}, ctx.Err()
		case <-time.After(backoff):
			// Continue to next retry attempt
		}
	}
	
	// Check for final error after all retries
	if err != nil {
		return &StepResult{
			StepID: step.ID,
			Agent:  step.Agent,
			Status: StepStatusFailed,
			Error:  err.Error(),
		}, err
	}

	// Process response
	result := &StepResult{
		StepID: step.ID,
		Agent:  step.Agent,
		Output: resp.Output,
	}

	// Check for escalation
	if resp.Escalation != nil && resp.Escalation.Required {
		result.Status = StepStatusEscalated
		result.EscalationTarget = resp.Escalation.TargetAgent
		result.EscalationReason = resp.Escalation.Reason
	} else {
		result.Status = StepStatusCompleted
	}

	return result, nil
}

// getPreviousAgents extracts the list of agents from previous step results
func (e *AgentExecutor) getPreviousAgents(currentResult *ExecutionResult) []string {
	if currentResult == nil || len(currentResult.StepResults) == 0 {
		return []string{}
	}

	agents := make([]string, 0, len(currentResult.StepResults))
	for _, stepResult := range currentResult.StepResults {
		if stepResult != nil && stepResult.Status == StepStatusCompleted && stepResult.Agent != "" {
			agents = append(agents, stepResult.Agent)
		}
	}

	return agents
}

// buildContext builds the context from previous step results
func (e *AgentExecutor) buildContext(currentResult *ExecutionResult) map[string]interface{} {
	context := make(map[string]interface{})

	if currentResult == nil || len(currentResult.StepResults) == 0 {
		return context
	}

	// Add results from previous steps
	previousResults := make([]map[string]interface{}, 0, len(currentResult.StepResults))
	for _, stepResult := range currentResult.StepResults {
		if stepResult != nil && stepResult.Status == StepStatusCompleted && stepResult.Output != nil {
			previousResults = append(previousResults, map[string]interface{}{
				"agent":  stepResult.Agent,
				"output": stepResult.Output,
			})
		}
	}

	context["previous_results"] = previousResults
	return context
}
