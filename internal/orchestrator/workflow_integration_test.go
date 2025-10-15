package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/process"
	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/internal/validation/evidence"
	"github.com/ferg-cod3s/conexus/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAgent implements Agent interface for testing
type mockAgent struct {
	id      string
	handler func(context.Context, schema.AgentRequest) (schema.AgentResponse, error)
}

func (m *mockAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	return m.handler(ctx, req)
}

// setupTestOrchestrator creates an orchestrator with test configuration
func setupTestOrchestrator(t *testing.T, config OrchestratorConfig) *Orchestrator {
	t.Helper()

	// Fill defaults for missing dependencies
	if config.ProcessManager == nil {
		config.ProcessManager = process.NewManager()
	}
	if config.ToolExecutor == nil {
		config.ToolExecutor = tool.NewExecutor()
	}
	if config.EvidenceValidator == nil {
		config.EvidenceValidator = evidence.NewValidator(false)
	}
	if config.QualityGates == nil {
		config.QualityGates = RelaxedQualityGates()
	}
	if !config.EnableProfiling {
		config.EnableProfiling = true
	}

	return NewWithConfig(config)
}

// createValidOutput creates a valid AgentOutputV1 with proper evidence
func createValidOutput() *schema.AgentOutputV1 {
	return &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "test-component",
		ScopeDescription: "Test scope",
		Overview:         "Test overview",
		EntryPoints: []schema.EntryPoint{
			{
				File:   "test.go",
				Lines:  "1-10",
				Symbol: "main",
				Role:   "entry",
			},
		},
		RawEvidence: []schema.Evidence{
			{
				Claim: "Entry point exists",
				File:  "test.go",
				Lines: "1-10",
			},
		},
	}
}

// createInvalidOutput creates an AgentOutputV1 with missing evidence
func createInvalidOutput() *schema.AgentOutputV1 {
	return &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "test-component",
		ScopeDescription: "Test scope",
		Overview:         "Test overview with unbacked claims",
		EntryPoints: []schema.EntryPoint{
			{
				File:   "missing.go",
				Lines:  "1-10",
				Symbol: "nonexistent",
				Role:   "handler",
			},
		},
		// No RawEvidence - this creates unbacked claims
	}
}

// TestWorkflowIntegration_BasicExecution tests basic workflow execution with profiling
func TestWorkflowIntegration_BasicExecution(t *testing.T) {
	ctx := context.Background()
	orch := setupTestOrchestrator(t, OrchestratorConfig{})
	
	// Register test agent
	agentID := "test-agent"
	orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
		return &mockAgent{
			id: agentID,
			handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   agentID,
					Status:    schema.StatusComplete,
					Output:    createValidOutput(),
					Duration:  100 * time.Millisecond,
					Timestamp: time.Now(),
				}, nil
			},
		}
	})
	
	// Create workflow
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{
				AgentID: agentID,
				Request: "test objective",
			},
		},
	}
	
	// Execute workflow
	result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
		ReadOnly:    true,
		MaxFileSize: 1024 * 1024,
	})
	
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Responses, 1)
	assert.Equal(t, schema.StatusComplete, result.Responses[0].Status)
	
	// Verify profiling report was generated
	assert.NotNil(t, result.ProfilingReport, "Profiling report should be generated")
	if result.ProfilingReport != nil {
		assert.Greater(t, result.ProfilingReport.TotalDuration, time.Duration(0))
		assert.Len(t, result.ProfilingReport.AgentExecutions, 1)
	}
}

// TestWorkflowIntegration_ValidationSuccess tests workflow with valid evidence
func TestWorkflowIntegration_ValidationSuccess(t *testing.T) {
	ctx := context.Background()
	orch := setupTestOrchestrator(t, OrchestratorConfig{})
	
	agentID := "valid-agent"
	orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
		return &mockAgent{
			id: agentID,
			handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   agentID,
					Status:    schema.StatusComplete,
					Output:    createValidOutput(),
					Timestamp: time.Now(),
				}, nil
			},
		}
	})
	
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{
				AgentID: agentID,
				Request: "test objective",
			},
		},
	}
	
	result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
		ReadOnly:    true,
		MaxFileSize: 1024 * 1024,
	})
	
	require.NoError(t, err)
	assert.True(t, result.Success)
	
	// Verify validation report
	assert.NotNil(t, result.ValidationReport)
	assert.Equal(t, 1, result.ValidationReport.ValidResponses)
	assert.Equal(t, 0, result.ValidationReport.UnbackedClaims)
	
	// Verify quality gate passed
	assert.NotNil(t, result.QualityGateResult)
	assert.True(t, result.QualityGateResult.Passed)
	assert.True(t, result.QualityGateResult.ValidationPassed)
}

// TestWorkflowIntegration_ValidationFailure tests workflow with invalid evidence
func TestWorkflowIntegration_ValidationFailure(t *testing.T) {
	ctx := context.Background()
	orch := setupTestOrchestrator(t, OrchestratorConfig{})
	
	agentID := "invalid-agent"
	orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
		return &mockAgent{
			id: agentID,
			handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   agentID,
					Status:    schema.StatusComplete,
					Output:    createInvalidOutput(),
					Timestamp: time.Now(),
				}, nil
			},
		}
	})
	
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{
				AgentID: agentID,
				Request: "test objective",
			},
		},
	}
	
	result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
		ReadOnly:    true,
		MaxFileSize: 1024 * 1024,
	})
	
	require.NoError(t, err)
	assert.True(t, result.Success) // Still succeeds with relaxed gates
	
	// Verify validation report shows invalid response
	assert.NotNil(t, result.ValidationReport)
	assert.Equal(t, 0, result.ValidationReport.ValidResponses)
	assert.Equal(t, 1, result.ValidationReport.InvalidResponses)
	assert.Greater(t, result.ValidationReport.UnbackedClaims, 0)
}

// TestWorkflowIntegration_QualityGatesBlocking tests strict quality gates
func TestWorkflowIntegration_QualityGatesBlocking(t *testing.T) {
	ctx := context.Background()
	
	config := OrchestratorConfig{
		EnableProfiling:     true,
		QualityGates:        StrictQualityGates(),
	}
	orch := setupTestOrchestrator(t, config)
	
	agentID := "invalid-agent"
	orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
		return &mockAgent{
			id: agentID,
			handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   agentID,
					Status:    schema.StatusComplete,
					Output:    createInvalidOutput(),
					Timestamp: time.Now(),
				}, nil
			},
		}
	})
	
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{
				AgentID: agentID,
				Request: "test objective",
			},
		},
	}
	
	result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
		ReadOnly:    true,
		MaxFileSize: 1024 * 1024,
	})
	
	// Under strict gates with invalid output, validation failure blocks and returns error
	require.Error(t, err)
	assert.False(t, result.Success, "Should fail with strict quality gates")
	// Validation report should be present even on early return
	assert.NotNil(t, result.ValidationReport)
	// QualityGateResult may be nil due to early return before gate evaluation
}

// TestWorkflowIntegration_ReportGeneration tests report export in different formats
func TestWorkflowIntegration_ReportGeneration(t *testing.T) {
	ctx := context.Background()
	orch := setupTestOrchestrator(t, OrchestratorConfig{})
	
	agentID := "test-agent"
	orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
		return &mockAgent{
			id: agentID,
			handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   agentID,
					Status:    schema.StatusComplete,
					Output:    createValidOutput(),
					Timestamp: time.Now(),
				}, nil
			},
		}
	})
	
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{
				AgentID: agentID,
				Request: "test objective",
			},
		},
	}
	
	result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
		ReadOnly:    true,
		MaxFileSize: 1024 * 1024,
	})
	
	require.NoError(t, err)
	assert.NotNil(t, result.WorkflowReport)
	
	// Test JSON export
	jsonReport, err := result.WorkflowReport.Export(FormatJSON)
	require.NoError(t, err)
	assert.Contains(t, jsonReport, "WorkflowID")
	assert.Contains(t, jsonReport, "OverallStatus")
	
	// Test text export
	textReport, err := result.WorkflowReport.Export(FormatText)
	require.NoError(t, err)
	assert.Contains(t, textReport, "WORKFLOW REPORT")
	assert.Contains(t, textReport, "=== VALIDATION ===")
	
	// Test markdown export
	mdReport, err := result.WorkflowReport.Export(FormatMarkdown)
	require.NoError(t, err)
	assert.Contains(t, mdReport, "# Workflow Report")
	assert.Contains(t, mdReport, "## Validation Results")
}

// TestWorkflowIntegration_MultiStepWorkflow tests multi-agent coordination
func TestWorkflowIntegration_MultiStepWorkflow(t *testing.T) {
	ctx := context.Background()
	orch := setupTestOrchestrator(t, OrchestratorConfig{})
	
	// Register multiple agents
	for i := 1; i <= 3; i++ {
		agentID := "agent-" + string(rune('0'+i))
		orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
			return &mockAgent{
				id: agentID,
				handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
					return schema.AgentResponse{
						RequestID: req.RequestID,
						AgentID:   agentID,
						Status:    schema.StatusComplete,
						Output:    createValidOutput(),
						Duration:  50 * time.Millisecond,
						Timestamp: time.Now(),
					}, nil
				},
			}
		})
	}
	
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{AgentID: "agent-1", Request: "step 1"},
			{AgentID: "agent-2", Request: "step 2"},
			{AgentID: "agent-3", Request: "step 3"},
		},
	}
	
	result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
		ReadOnly:    true,
		MaxFileSize: 1024 * 1024,
	})
	
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Responses, 3)
	
	// Verify profiling tracked all executions
	assert.NotNil(t, result.ProfilingReport)
	assert.Len(t, result.ProfilingReport.AgentExecutions, 3)
	
	// Verify validation checked all responses
	assert.NotNil(t, result.ValidationReport)
	assert.Equal(t, 3, result.ValidationReport.TotalResponses)
	assert.Equal(t, 3, result.ValidationReport.ValidResponses)
}

// TestWorkflowIntegration_ErrorHandling tests error propagation
func TestWorkflowIntegration_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	orch := setupTestOrchestrator(t, OrchestratorConfig{})
	
	agentID := "error-agent"
	orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
		return &mockAgent{
			id: agentID,
			handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   agentID,
					Status:    schema.StatusError,
					Error: &schema.AgentError{
						Code:    "TEST_ERROR",
						Message: "Simulated error",
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	})
	
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{
				AgentID: agentID,
				Request: "test objective",
			},
		},
	}
	
	result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
		ReadOnly:    true,
		MaxFileSize: 1024 * 1024,
	})
	
	require.NoError(t, err) // No execution error
	assert.False(t, result.Success) // But workflow failed
	assert.Len(t, result.Responses, 1)
	assert.Equal(t, schema.StatusError, result.Responses[0].Status)
	assert.NotNil(t, result.Responses[0].Error)
}

// TestWorkflowIntegration_QualityGateEvaluation tests different quality gate configurations
func TestWorkflowIntegration_QualityGateEvaluation(t *testing.T) {
	tests := []struct {
		name         string
		config       *QualityGateConfig
		output       *schema.AgentOutputV1
		expectPass   bool
		expectViolations bool
	}{
		{
			name:         "Valid with default gates",
			config:       DefaultQualityGates(),
			output:       createValidOutput(),
			expectPass:   true,
			expectViolations: false,
		},
		{
			name:         "Invalid with relaxed gates",
			config:       RelaxedQualityGates(),
			output:       createInvalidOutput(),
			expectPass:   true, // Relaxed gates allow some violations
			expectViolations: true,
		},
		{
			name:         "Invalid with strict gates",
			config:       StrictQualityGates(),
			output:       createInvalidOutput(),
			expectPass:   false,
			expectViolations: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			
			config := OrchestratorConfig{
				EnableProfiling:     true,
				QualityGates:        tt.config,
			}
			orch := setupTestOrchestrator(t, config)
			
			agentID := "test-agent"
			output := tt.output
			orch.RegisterAgent(agentID, func(exec *tool.Executor) Agent {
				return &mockAgent{
					id: agentID,
					handler: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
						return schema.AgentResponse{
							RequestID: req.RequestID,
							AgentID:   agentID,
							Status:    schema.StatusComplete,
							Output:    output,
							Timestamp: time.Now(),
						}, nil
					},
				}
			})
			
			workflow := &Workflow{
				Steps: []WorkflowStep{
					{
						AgentID: agentID,
						Request: "test objective",
					},
				},
			}
			
			result, err := orch.ExecuteWorkflow(ctx, workflow, schema.Permissions{
				ReadOnly:    true,
				MaxFileSize: 1024 * 1024,
			})
			
			// If gates are strict and output is invalid, the orchestrator may return early
			// with an error on validation failure before computing QualityGateResult.
			if !tt.expectPass && tt.config.BlockOnValidationFailure {
				require.Error(t, err)
				assert.False(t, result.Success)
				assert.NotNil(t, result.ValidationReport)
				return
			}

			// Otherwise, expect normal completion with a populated QualityGateResult
			require.NoError(t, err)
			assert.NotNil(t, result.QualityGateResult)
			assert.Equal(t, tt.expectPass, result.QualityGateResult.Passed, 
				"Quality gate pass status mismatch")
			
			if tt.expectViolations {
				assert.NotEmpty(t, result.QualityGateResult.Violations,
					"Expected violations but found none")
			}
		})
	}
}