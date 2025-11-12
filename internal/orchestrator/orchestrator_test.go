package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/process"
	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// MockAgent implements the Agent interface for testing
type MockAgent struct {
	ExecuteFunc func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error)
}

func (m *MockAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	return m.ExecuteFunc(ctx, req)
}

func TestOrchestrator_RegisterAgent(t *testing.T) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	factory := func(executor *tool.Executor) Agent {
		return &MockAgent{}
	}

	orch.RegisterAgent("test-agent", factory)

	if _, exists := orch.agentRegistry["test-agent"]; !exists {
		t.Error("Agent was not registered")
	}
}

func TestOrchestrator_HandleRequest_Routing(t *testing.T) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	// Register mock agents
	mockLocator := func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "codebase-locator",
					Status:    schema.StatusComplete,
					Output: &schema.AgentOutputV1{
						Version:          "AGENT_OUTPUT_V1",
						ComponentName:    "Test",
						ScopeDescription: "Test scope",
						Overview:         "Test overview",
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}

	mockAnalyzer := func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "codebase-analyzer",
					Status:    schema.StatusComplete,
					Output: &schema.AgentOutputV1{
						Version:          "AGENT_OUTPUT_V1",
						ComponentName:    "Test",
						ScopeDescription: "Test scope",
						Overview:         "Test overview",
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}

	orch.RegisterAgent("codebase-locator", mockLocator)
	orch.RegisterAgent("codebase-analyzer", mockAnalyzer)

	tests := []struct {
		request       string
		expectedAgent string
	}{
		{"find all .go files", "codebase-locator"},
		{"locate function Add", "codebase-locator"},
		{"analyze this code", "codebase-analyzer"},
		{"how does this work", "codebase-analyzer"},
		{"search for files", "codebase-locator"},
	}

	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			ctx := context.Background()
			perms := schema.Permissions{
				AllowedDirectories: []string{"/tmp"},
				ReadOnly:           true,
			}

			result, err := orch.HandleRequest(ctx, tt.request, perms)
			if err != nil {
				t.Fatalf("HandleRequest failed: %v", err)
			}

			if !result.Success {
				t.Errorf("Expected success, got failure: %s", result.Error)
			}

			if len(result.Responses) != 1 {
				t.Errorf("Expected 1 response, got %d", len(result.Responses))
			}

			if len(result.Responses) > 0 && result.Responses[0].AgentID != tt.expectedAgent {
				t.Errorf("Expected agent %s, got %s", tt.expectedAgent, result.Responses[0].AgentID)
			}
		})
	}
}

func TestOrchestrator_WorkflowExecution(t *testing.T) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	executionOrder := []string{}

	mockAgent1 := func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				executionOrder = append(executionOrder, "agent1")
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "agent1",
					Status:    schema.StatusComplete,
					Output: &schema.AgentOutputV1{
						Version:          "AGENT_OUTPUT_V1",
						ComponentName:    "Test1",
						ScopeDescription: "Test",
						Overview:         "Test",
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}

	mockAgent2 := func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				executionOrder = append(executionOrder, "agent2")
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "agent2",
					Status:    schema.StatusComplete,
					Output: &schema.AgentOutputV1{
						Version:          "AGENT_OUTPUT_V1",
						ComponentName:    "Test2",
						ScopeDescription: "Test",
						Overview:         "Test",
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}

	orch.RegisterAgent("agent1", mockAgent1)
	orch.RegisterAgent("agent2", mockAgent2)

	workflow := &Workflow{
		Steps: []WorkflowStep{
			{AgentID: "agent1", Request: "step 1"},
			{AgentID: "agent2", Request: "step 2"},
		},
	}

	ctx := context.Background()
	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
	}

	result, err := orch.ExecuteWorkflow(ctx, workflow, perms)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got failure: %s", result.Error)
	}

	if len(result.Responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(result.Responses))
	}

	// Verify execution order
	if len(executionOrder) != 2 || executionOrder[0] != "agent1" || executionOrder[1] != "agent2" {
		t.Errorf("Expected execution order [agent1, agent2], got %v", executionOrder)
	}
}

func TestOrchestrator_ErrorHandling(t *testing.T) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	mockErrorAgent := func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "error-agent",
					Status:    schema.StatusError,
					Error: &schema.AgentError{
						Code:        "TEST_ERROR",
						Message:     "Simulated error",
						Recoverable: false,
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}

	orch.RegisterAgent("error-agent", mockErrorAgent)

	workflow := &Workflow{
		Steps: []WorkflowStep{
			{AgentID: "error-agent", Request: "test error"},
		},
	}

	ctx := context.Background()
	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
	}

	result, err := orch.ExecuteWorkflow(ctx, workflow, perms)

	// Should not return error from ExecuteWorkflow
	if err != nil {
		t.Logf("Workflow error: %v", err)
	}

	// But result should indicate failure
	if result.Success {
		t.Error("Expected failure, got success")
	}

	if result.Error == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestOrchestrator_Escalation(t *testing.T) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	executionCount := 0

	mockEscalatingAgent := func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				executionCount++
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "escalating-agent",
					Status:    schema.StatusEscalationRequired,
					Escalation: &schema.Escalation{
						Required:     true,
						TargetAgent:  "target-agent",
						Reason:       "Need more analysis",
						RequiredInfo: "Additional context needed",
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}

	mockTargetAgent := func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				executionCount++
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "target-agent",
					Status:    schema.StatusComplete,
					Output: &schema.AgentOutputV1{
						Version:          "AGENT_OUTPUT_V1",
						ComponentName:    "Escalated",
						ScopeDescription: "Escalated analysis",
						Overview:         "Completed after escalation",
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}

	orch.RegisterAgent("escalating-agent", mockEscalatingAgent)
	orch.RegisterAgent("target-agent", mockTargetAgent)

	workflow := &Workflow{
		Steps: []WorkflowStep{
			{AgentID: "escalating-agent", Request: "initial request"},
		},
	}

	ctx := context.Background()
	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
	}

	result, err := orch.ExecuteWorkflow(ctx, workflow, perms)
	if err != nil {
		t.Fatalf("ExecuteWorkflow failed: %v", err)
	}

	// Should have executed both agents
	if executionCount != 2 {
		t.Errorf("Expected 2 agent executions, got %d", executionCount)
	}

	if len(result.Responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(result.Responses))
	}
}

func TestRouter_Route(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		request       string
		expectedAgent string
	}{
		{"find all files", "codebase-locator"},
		{"locate function", "codebase-locator"},
		{"analyze this code", "codebase-analyzer"},
		{"how does this work", "codebase-analyzer"},
		{"understand the flow", "codebase-analyzer"},
		{"where is the main function", "codebase-locator"},
		{"random request", "codebase-locator"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			selection, err := router.Route(tt.request)
			if err != nil {
				t.Fatalf("Route failed: %v", err)
			}

			if selection.PrimaryAgent != tt.expectedAgent {
				t.Errorf("Expected agent %s, got %s", tt.expectedAgent, selection.PrimaryAgent)
			}
		})
	}
}

func TestRouter_AddRule(t *testing.T) {
	router := NewRouter()

	initialRuleCount := len(router.rules)

	router.AddRule(RoutingRule{
		Keywords: []string{"test"},
		AgentID:  "test-agent",
		Priority: 5,
	})

	if len(router.rules) != initialRuleCount+1 {
		t.Error("Rule was not added")
	}
}
