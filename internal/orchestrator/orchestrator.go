package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/process"
	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/internal/validation/evidence"
	"github.com/ferg-cod3s/conexus/internal/profiling"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// AgentFactory creates agent instances
type AgentFactory func(executor *tool.Executor) Agent

// Agent represents an executable agent
type Agent interface {
	Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error)
}

// Orchestrator coordinates agent execution and workflow management
type Orchestrator struct {
	processManager    *process.Manager
	toolExecutor      *tool.Executor
	agentRegistry     map[string]AgentFactory
	router            *Router
	evidenceValidator *evidence.Validator
	qualityGates      *QualityGateConfig
	enableProfiling   bool
}

// OrchestratorConfig contains configuration for the orchestrator
type OrchestratorConfig struct {
	ProcessManager    *process.Manager
	ToolExecutor      *tool.Executor
	EvidenceValidator *evidence.Validator
	QualityGates      *QualityGateConfig
	EnableProfiling   bool
}

// New creates a new Orchestrator with basic dependencies
func New(pm *process.Manager, te *tool.Executor) *Orchestrator {
	return NewWithConfig(OrchestratorConfig{
		ProcessManager:    pm,
		ToolExecutor:      te,
		EvidenceValidator: evidence.NewValidator(true), // strict mode by default
		QualityGates:      DefaultQualityGates(),
		EnableProfiling:   true,
	})
}

// NewWithConfig creates a new Orchestrator with custom configuration
func NewWithConfig(config OrchestratorConfig) *Orchestrator {
	return &Orchestrator{
		processManager:    config.ProcessManager,
		toolExecutor:      config.ToolExecutor,
		agentRegistry:     make(map[string]AgentFactory),
		router:            NewRouter(),
		evidenceValidator: config.EvidenceValidator,
		qualityGates:      config.QualityGates,
		enableProfiling:   config.EnableProfiling,
	}
}

// RegisterAgent adds an agent factory to the registry
func (o *Orchestrator) RegisterAgent(agentID string, factory AgentFactory) {
	o.agentRegistry[agentID] = factory
}

// HandleRequest processes a user request and routes it to appropriate agents
func (o *Orchestrator) HandleRequest(ctx context.Context, userRequest string, permissions schema.Permissions) (*Result, error) {
	startTime := time.Now()

	// Route the request to determine which agent(s) to invoke
	selection, err := o.router.Route(userRequest)
	if err != nil {
		return nil, fmt.Errorf("routing error: %w", err)
	}

	// Create the workflow
	workflow := &Workflow{
		Steps: []WorkflowStep{
			{
				AgentID: selection.PrimaryAgent,
				Request: userRequest,
			},
		},
	}

	// Execute the workflow
	result, err := o.ExecuteWorkflow(ctx, workflow, permissions)
	if err != nil {
		return nil, fmt.Errorf("workflow execution error: %w", err)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// ExecuteWorkflow executes a series of agent invocations with validation and profiling
func (o *Orchestrator) ExecuteWorkflow(ctx context.Context, workflow *Workflow, permissions schema.Permissions) (*Result, error) {
	workflowID := generateWorkflowID()
	startTime := time.Now()

	result := &Result{
		Success:   true,
		Responses: []schema.AgentResponse{},
	}

	// Initialize profiling if enabled
	var profiler *WorkflowProfiler
	if o.enableProfiling {
		profiler = NewWorkflowProfiler(workflowID, true)
	}

	// Initialize validation tracking
	validationResults := make([]AgentValidationResult, 0)

	accumulatedContext := make(map[string]interface{})
	previousAgents := []string{}

	// Execute each step sequentially
	for i := 0; i < len(workflow.Steps); i++ {
		step := workflow.Steps[i]

		// Start profiling for this agent
		var execCtx *profiling.ExecutionContext
		if profiler != nil {
			execCtx = profiler.StartAgentExecution(ctx, step.AgentID, step.Request)
		}

		// Execute the agent
		agentResponse, err := o.invokeAgent(ctx, step, permissions, accumulatedContext, previousAgents)
		
		// Finalize profiling (always) so aggregates are updated
		if execCtx != nil {
			execCtx.End(agentResponse.Output, err)
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			return result, err
		}

		result.Responses = append(result.Responses, agentResponse)
		previousAgents = append(previousAgents, agentResponse.AgentID)

		// Validate the response if it contains AGENT_OUTPUT_V1
		if o.evidenceValidator != nil && agentResponse.Output != nil {
			validationResult := o.validateAgentResponse(agentResponse)
			validationResults = append(validationResults, validationResult)

			// Block on validation failure if configured
			if o.qualityGates.BlockOnValidationFailure && !validationResult.Valid {
				result.Success = false
				result.Error = fmt.Sprintf("Agent %s validation failed: %d unbacked claims, %d invalid evidence",
					agentResponse.AgentID,
					len(validationResult.UnbackedClaims),
					len(validationResult.InvalidEvidence))
				
				// Generate reports before returning
				o.generateReports(result, workflowID, startTime, validationResults, profiler)
				return result, fmt.Errorf("%s", result.Error)
			}
		}

		// Handle errors
		if agentResponse.Status == schema.StatusError {
			result.Success = false
			if agentResponse.Error != nil {
				result.Error = agentResponse.Error.Message
			}
			break
		}

		// Check for escalation
		if agentResponse.Escalation != nil && agentResponse.Escalation.Required {
			// Add escalated agent to workflow
			nextStep := WorkflowStep{
				AgentID: agentResponse.Escalation.TargetAgent,
				Request: agentResponse.Escalation.RequiredInfo,
			}
			workflow.Steps = append(workflow.Steps, nextStep)
		}
	}

	// Generate comprehensive reports
	o.generateReports(result, workflowID, startTime, validationResults, profiler)

	// Check quality gates
	if result.ValidationReport != nil || result.ProfilingReport != nil {
		qualityGateResult := o.qualityGates.CheckQualityGates(
			result.ValidationReport,
			result.ProfilingReport,
		)
		result.QualityGateResult = qualityGateResult

		// Block on quality gate failure if configured
		if !qualityGateResult.Passed {
			if o.qualityGates.BlockOnValidationFailure && !qualityGateResult.ValidationPassed {
				result.Success = false
				result.Error = "Quality gate validation check failed"
				return result, fmt.Errorf("%s", result.Error)
			}
			if o.qualityGates.BlockOnPerformanceFailure && !qualityGateResult.PerformancePassed {
				result.Success = false
				result.Error = "Quality gate performance check failed"
				return result, fmt.Errorf("%s", result.Error)
			}
		}
	}

	return result, nil
}

// validateAgentResponse validates an agent response using the evidence validator
func (o *Orchestrator) validateAgentResponse(response schema.AgentResponse) AgentValidationResult {
	result := AgentValidationResult{
		AgentID:   response.AgentID,
		RequestID: response.RequestID,
		Valid:     true,
	}

	if response.Output == nil {
		return result
	}

	// Validate evidence backing
	validationResult, err := o.evidenceValidator.Validate(response.Output)
	if err != nil {
		result.Valid = false
		return result
	}

	result.Valid = validationResult.Valid
	result.EvidenceCoverage = validationResult.CoveragePercentage
	result.UnbackedClaims = validationResult.UnbackedClaims
	result.InvalidEvidence = validationResult.InvalidEvidence

	return result
}

// generateReports creates validation and profiling reports
func (o *Orchestrator) generateReports(
	result *Result,
	workflowID string,
	startTime time.Time,
	validationResults []AgentValidationResult,
	profiler *WorkflowProfiler,
) {
	// Generate validation report
	if len(validationResults) > 0 {
		result.ValidationReport = CreateValidationReportFromResults(workflowID, validationResults)
	}

	// Generate profiling report
	if profiler != nil {
		result.ProfilingReport = profiler.GenerateReport()
	}

	// Generate combined workflow report
	if result.ValidationReport != nil || result.ProfilingReport != nil {
		result.WorkflowReport = GenerateReport(
			workflowID,
			result.ValidationReport,
			result.ProfilingReport,
			result.QualityGateResult,
		)
	}
}

// invokeAgent invokes a single agent
func (o *Orchestrator) invokeAgent(
	ctx context.Context,
	step WorkflowStep,
	permissions schema.Permissions,
	accumulatedContext map[string]interface{},
	previousAgents []string,
) (schema.AgentResponse, error) {
	// Get agent factory
	factory, exists := o.agentRegistry[step.AgentID]
	if !exists {
		return schema.AgentResponse{}, fmt.Errorf("agent not found: %s", step.AgentID)
	}

	// Create agent instance
	agent := factory(o.toolExecutor)

	// Build agent request
	agentReq := schema.AgentRequest{
		RequestID: generateRequestID(),
		AgentID:   step.AgentID,
		Task: schema.AgentTask{
			TargetAgent:        step.AgentID,
			Files:              step.Files,
			AllowedDirectories: permissions.AllowedDirectories,
			SpecificRequest:    step.Request,
		},
		Context: schema.ConversationContext{
			UserRequest:        step.Request,
			PreviousAgents:     previousAgents,
			AccumulatedContext: accumulatedContext,
		},
		Permissions: permissions,
		Timestamp:   time.Now(),
	}

	// Execute agent
	response, err := agent.Execute(ctx, agentReq)
	if err != nil {
		return schema.AgentResponse{
			RequestID: agentReq.RequestID,
			AgentID:   step.AgentID,
			Status:    schema.StatusError,
			Error: &schema.AgentError{
				Code:        "AGENT_EXECUTION_ERROR",
				Message:     err.Error(),
				Recoverable: false,
			},
			Timestamp: time.Now(),
		}, err
	}

	return response, nil
}

// Router handles request routing logic
type Router struct {
	rules []RoutingRule
}

// RoutingRule maps patterns to agents
type RoutingRule struct {
	Keywords []string
	AgentID  string
	Priority int
}

// NewRouter creates a new Router with default rules
func NewRouter() *Router {
	router := &Router{
		rules: []RoutingRule{},
	}

	// Add default routing rules
	router.AddRule(RoutingRule{
		Keywords: []string{"find", "locate", "search", "files", "where"},
		AgentID:  "codebase-locator",
		Priority: 10,
	})

	router.AddRule(RoutingRule{
		Keywords: []string{"analyze", "how", "works", "flow", "calls", "understand"},
		AgentID:  "codebase-analyzer",
		Priority: 10,
	})

	return router
}

// AddRule adds a routing rule
func (r *Router) AddRule(rule RoutingRule) {
	r.rules = append(r.rules, rule)
}

// Route determines which agent should handle the request
func (r *Router) Route(userRequest string) (AgentSelection, error) {
	lower := strings.ToLower(userRequest)

	bestMatch := AgentSelection{}
	bestScore := 0

	for _, rule := range r.rules {
		score := 0
		for _, keyword := range rule.Keywords {
			if strings.Contains(lower, keyword) {
				score += rule.Priority
			}
		}

		if score > bestScore {
			bestScore = score
			bestMatch = AgentSelection{
				PrimaryAgent: rule.AgentID,
				Parameters:   extractParameters(userRequest),
			}
		}
	}

	// Default to locator if no match
	if bestMatch.PrimaryAgent == "" {
		bestMatch.PrimaryAgent = "codebase-locator"
	}

	return bestMatch, nil
}

// AgentSelection represents the result of routing
type AgentSelection struct {
	PrimaryAgent   string
	FallbackAgents []string
	Parameters     map[string]interface{}
}

// Workflow represents a sequence of agent invocations
type Workflow struct {
	Steps []WorkflowStep
}

// WorkflowStep represents a single agent invocation
type WorkflowStep struct {
	AgentID string
	Request string
	Files   []string
}

// Result contains the outcome of orchestration
type Result struct {
	Success            bool
	Responses          []schema.AgentResponse
	Error              string
	Duration           time.Duration
	ValidationReport   *ValidationReport
	ProfilingReport    *ProfilingReport
	QualityGateResult  *QualityGateResult
	WorkflowReport     *WorkflowReport
}

// Helper functions

func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

func generateWorkflowID() string {
	return fmt.Sprintf("workflow-%d", time.Now().UnixNano())
}

func extractParameters(request string) map[string]interface{} {
	params := make(map[string]interface{})

	// Extract file patterns
	if strings.Contains(request, "*.") {
		words := strings.Fields(request)
		for _, word := range words {
			if strings.HasPrefix(word, "*.") {
				params["pattern"] = word
				break
			}
		}
	}

	return params
}
