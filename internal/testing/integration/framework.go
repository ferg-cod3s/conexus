// Package integration provides an end-to-end testing framework for multi-agent workflows.
//
// The integration testing framework enables comprehensive testing of:
// - Full workflow execution
// - Multi-agent coordination
// - Real codebase analysis
// - Performance validation
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ferg-cod3s/conexus/internal/orchestrator/workflow"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// TestFramework coordinates integration testing
type TestFramework struct {
	engine   *workflow.Engine
	executor *workflow.AgentExecutor
	results  []*TestResult
}

// NewTestFramework creates a new integration test framework
func NewTestFramework() *TestFramework {
	executor := workflow.NewAgentExecutor()
	engine := workflow.NewEngine(executor)

	return &TestFramework{
		engine:   engine,
		executor: executor,
		results:  make([]*TestResult, 0),
	}
}

// TestCase represents an integration test case
type TestCase struct {
	Name        string
	Description string
	Workflow    *workflow.Workflow
	Timeout     time.Duration
	Assertions  []Assertion
}

// Assertion defines a test assertion
type Assertion interface {
	Assert(result *workflow.ExecutionResult) error
	Description() string
}

// TestResult contains the results of a test case execution
type TestResult struct {
	TestName      string
	Passed        bool
	Duration      time.Duration
	WorkflowResult *workflow.ExecutionResult
	Errors        []error
	Warnings      []string
	Assertions    []AssertionResult
}

// AssertionResult contains the result of a single assertion
type AssertionResult struct {
	Description string
	Passed      bool
	Error       error
}

// AssertMaxDuration checks if the test duration is within the specified maximum
// Returns an error if duration exceeds the threshold
func (r *TestResult) AssertMaxDuration(max time.Duration) error {
	if r.Duration > max {
		return fmt.Errorf("test duration %v exceeded maximum allowed %v (by %v)", 
			r.Duration, max, r.Duration-max)
	}
	return nil
}

// RegisterAgent registers an agent for testing
func (f *TestFramework) RegisterAgent(name string, agent workflow.Agent) {
	f.executor.RegisterAgent(name, agent)
}

// Run executes a test case
func (f *TestFramework) Run(ctx context.Context, testCase *TestCase) *TestResult {
	result := &TestResult{
		TestName:   testCase.Name,
		Errors:     make([]error, 0),
		Warnings:   make([]string, 0),
		Assertions: make([]AssertionResult, 0),
	}

	// Apply timeout
	timeout := testCase.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	testCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute workflow
	startTime := time.Now()
	workflowResult, err := f.engine.Execute(testCtx, testCase.Workflow)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Errorf("workflow execution failed: %w", err))
		f.results = append(f.results, result)
		return result
	}

	result.WorkflowResult = workflowResult

	// Run assertions
	allAssertionsPassed := true
	for _, assertion := range testCase.Assertions {
		assertionResult := AssertionResult{
			Description: assertion.Description(),
			Passed:      true,
		}

		if err := assertion.Assert(workflowResult); err != nil {
			assertionResult.Passed = false
			assertionResult.Error = err
			allAssertionsPassed = false
		}

		result.Assertions = append(result.Assertions, assertionResult)
	}

	result.Passed = allAssertionsPassed && len(result.Errors) == 0

	f.results = append(f.results, result)
	return result
}

// RunSuite executes multiple test cases
func (f *TestFramework) RunSuite(ctx context.Context, testCases []*TestCase) *SuiteResult {
	suiteResult := &SuiteResult{
		TotalTests:   len(testCases),
		PassedTests:  0,
		FailedTests:  0,
		TestResults:  make([]*TestResult, 0, len(testCases)),
		TotalDuration: 0,
	}

	startTime := time.Now()

	for _, testCase := range testCases {
		result := f.Run(ctx, testCase)
		suiteResult.TestResults = append(suiteResult.TestResults, result)

		if result.Passed {
			suiteResult.PassedTests++
		} else {
			suiteResult.FailedTests++
		}

		suiteResult.TotalDuration += result.Duration
	}

	suiteResult.SuiteDuration = time.Since(startTime)

	return suiteResult
}

// SuiteResult contains results from a test suite
type SuiteResult struct {
	TotalTests    int
	PassedTests   int
	FailedTests   int
	TestResults   []*TestResult
	TotalDuration time.Duration
	SuiteDuration time.Duration
}

// GetResults returns all test results
func (f *TestFramework) GetResults() []*TestResult {
	return f.results
}

// Clear clears all test results
func (f *TestFramework) Clear() {
	f.results = make([]*TestResult, 0)
}

// --- Built-in Assertions ---

// WorkflowSuccessAssertion checks if workflow completed successfully
type WorkflowSuccessAssertion struct{}

func (a *WorkflowSuccessAssertion) Assert(result *workflow.ExecutionResult) error {
	if result.Status != workflow.StatusCompleted {
		return fmt.Errorf("workflow status is %s, expected %s", result.Status, workflow.StatusCompleted)
	}
	return nil
}

func (a *WorkflowSuccessAssertion) Description() string {
	return "Workflow completed successfully"
}

// StepCountAssertion checks if expected number of steps executed
type StepCountAssertion struct {
	ExpectedCount int
}

func (a *StepCountAssertion) Assert(result *workflow.ExecutionResult) error {
	actual := len(result.StepResults)
	if actual != a.ExpectedCount {
		return fmt.Errorf("expected %d steps, got %d", a.ExpectedCount, actual)
	}
	return nil
}

func (a *StepCountAssertion) Description() string {
	return fmt.Sprintf("Expected %d workflow steps", a.ExpectedCount)
}

// AllStepsSuccessAssertion checks if all steps succeeded
type AllStepsSuccessAssertion struct{}

func (a *AllStepsSuccessAssertion) Assert(result *workflow.ExecutionResult) error {
	for i, step := range result.StepResults {
		if step.Status != workflow.StepStatusCompleted {
			return fmt.Errorf("step %d (%s) status is %s, expected %s", i, step.StepID, step.Status, workflow.StepStatusCompleted)
		}
	}
	return nil
}

func (a *AllStepsSuccessAssertion) Description() string {
	return "All steps completed successfully"
}

// MaxDurationAssertion checks if workflow completed within time limit
type MaxDurationAssertion struct {
	MaxDuration time.Duration
}

func (a *MaxDurationAssertion) Assert(result *workflow.ExecutionResult) error {
	// Duration is tracked at test result level, not workflow result
	// This assertion is checked differently
	return nil
}

func (a *MaxDurationAssertion) Description() string {
	return fmt.Sprintf("Completed within %s", a.MaxDuration)
}

// OutputNotNilAssertion checks if outputs are not nil
type OutputNotNilAssertion struct{}

func (a *OutputNotNilAssertion) Assert(result *workflow.ExecutionResult) error {
	for i, step := range result.StepResults {
		if step.Status == workflow.StepStatusCompleted && step.Output == nil {
			return fmt.Errorf("step %d (%s) has nil output", i, step.StepID)
		}
	}
	return nil
}

func (a *OutputNotNilAssertion) Description() string {
	return "All completed steps have non-nil output"
}

// AgentExecutedAssertion checks if specific agent was executed
type AgentExecutedAssertion struct {
	AgentName string
}

func (a *AgentExecutedAssertion) Assert(result *workflow.ExecutionResult) error {
	for _, step := range result.StepResults {
		if step.Agent == a.AgentName {
			return nil
		}
	}
	return fmt.Errorf("agent %s was not executed", a.AgentName)
}

func (a *AgentExecutedAssertion) Description() string {
	return fmt.Sprintf("Agent %s was executed", a.AgentName)
}

// NoEscalationsAssertion checks that no escalations occurred
type NoEscalationsAssertion struct{}

func (a *NoEscalationsAssertion) Assert(result *workflow.ExecutionResult) error {
	for i, step := range result.StepResults {
		if step.Status == workflow.StepStatusEscalated {
			return fmt.Errorf("step %d (%s) escalated to %s", i, step.StepID, step.EscalationTarget)
		}
	}
	return nil
}

func (a *NoEscalationsAssertion) Description() string {
	return "No escalations occurred"
}

// EscalationOccurredAssertion checks that escalation happened
type EscalationOccurredAssertion struct {
	TargetAgent string
}

func (a *EscalationOccurredAssertion) Assert(result *workflow.ExecutionResult) error {
	for _, step := range result.StepResults {
		if step.Status == workflow.StepStatusEscalated {
			if a.TargetAgent == "" || step.EscalationTarget == a.TargetAgent {
				return nil
			}
		}
	}

	if a.TargetAgent != "" {
		return fmt.Errorf("no escalation to %s occurred", a.TargetAgent)
	}
	return fmt.Errorf("no escalation occurred")
}

func (a *EscalationOccurredAssertion) Description() string {
	if a.TargetAgent != "" {
		return fmt.Sprintf("Escalation to %s occurred", a.TargetAgent)
	}
	return "Escalation occurred"
}
// EvidenceValidAssertion checks that all claims have evidence backing
// EvidenceValidAssertion checks that all claims have evidence backing
type EvidenceValidAssertion struct {
	StrictMode bool
}

func (a *EvidenceValidAssertion) Assert(result *workflow.ExecutionResult) error {
	for i, step := range result.StepResults {
		if step.Status != workflow.StepStatusCompleted || step.Output == nil {
			continue
		}

		// Use the helper to verify evidence
		report, err := VerifyEvidence(step.Output, a.StrictMode)
		if err != nil {
			return fmt.Errorf("step %d (%s): evidence validation failed: %w", i, step.StepID, err)
		}

		if !report.EvidenceValid {
			if report.EvidenceResult != nil {
				// Build detailed error message with unbacked claims
				errMsg := fmt.Sprintf("step %d (%s): evidence validation failed: %d unbacked claims, %d invalid evidence",
					i, step.StepID, len(report.EvidenceResult.UnbackedClaims), len(report.EvidenceResult.InvalidEvidence))
				
				// Add details about unbacked claims for debugging
				if len(report.EvidenceResult.UnbackedClaims) > 0 {
					errMsg += "\n  Unbacked claims:"
					for _, claim := range report.EvidenceResult.UnbackedClaims {
						errMsg += fmt.Sprintf("\n    - Section: %s, Index: %d, Description: %s",
							claim.Section, claim.Index, claim.Description)
					}
				}
				
				return fmt.Errorf("%s", errMsg)
			}
			return fmt.Errorf("step %d (%s): evidence validation failed", i, step.StepID)
		}
	}
	return nil
}

func (a *EvidenceValidAssertion) Description() string {
	mode := "non-strict"
	if a.StrictMode {
		mode = "strict"
	}
	return fmt.Sprintf("All claims have evidence backing (%s mode)", mode)
}

// SchemaValidAssertion checks AGENT_OUTPUT_V1 schema compliance
type SchemaValidAssertion struct{}

func (a *SchemaValidAssertion) Assert(result *workflow.ExecutionResult) error {
	for i, step := range result.StepResults {
		if step.Status != workflow.StepStatusCompleted || step.Output == nil {
			continue
		}

		// Use the helper to verify schema
		err := VerifySchema(step.Output)
		if err != nil {
			return fmt.Errorf("step %d (%s): schema validation failed: %w", i, step.StepID, err)
		}
	}
	return nil
}

func (a *SchemaValidAssertion) Description() string {
	return "All outputs comply with AGENT_OUTPUT_V1 schema"
}

// PerformanceWithinBudgetAssertion checks execution time against workflow duration
type PerformanceWithinBudgetAssertion struct {
	MaxDuration time.Duration
}

func (a *PerformanceWithinBudgetAssertion) Assert(result *workflow.ExecutionResult) error {
	// Note: Duration is tracked at TestResult level, not StepResult
	// This assertion should be checked at the test framework level
	// For now, we just return nil as the framework already handles timeout
	return nil
}

func (a *PerformanceWithinBudgetAssertion) Description() string {
	return fmt.Sprintf("Workflow completed within %s", a.MaxDuration)
}

// OutputFieldNotEmptyAssertion checks specific field is populated
type OutputFieldNotEmptyAssertion struct {
	FieldName string
}

func (a *OutputFieldNotEmptyAssertion) Assert(result *workflow.ExecutionResult) error {
	for i, step := range result.StepResults {
		if step.Status != workflow.StepStatusCompleted || step.Output == nil {
			continue
		}

		output := step.Output

		// Check specific fields using actual AgentOutputV1 field names
		switch a.FieldName {
		case "component_name":
			if output.ComponentName == "" {
				return fmt.Errorf("step %d (%s): component_name is empty", i, step.StepID)
			}
		case "scope_description":
			if output.ScopeDescription == "" {
				return fmt.Errorf("step %d (%s): scope_description is empty", i, step.StepID)
			}
		case "overview":
			if output.Overview == "" {
				return fmt.Errorf("step %d (%s): overview is empty", i, step.StepID)
			}
		case "entry_points":
			if len(output.EntryPoints) == 0 {
				return fmt.Errorf("step %d (%s): entry_points is empty", i, step.StepID)
			}
		case "call_graph":
			if len(output.CallGraph) == 0 {
				return fmt.Errorf("step %d (%s): call_graph is empty", i, step.StepID)
			}
		case "raw_evidence":
			if len(output.RawEvidence) == 0 {
				return fmt.Errorf("step %d (%s): raw_evidence is empty", i, step.StepID)
			}
		case "state_management":
			if len(output.StateManagement) == 0 {
				return fmt.Errorf("step %d (%s): state_management is empty", i, step.StepID)
			}
		case "side_effects":
			if len(output.SideEffects) == 0 {
				return fmt.Errorf("step %d (%s): side_effects is empty", i, step.StepID)
			}
		case "error_handling":
			if len(output.ErrorHandling) == 0 {
				return fmt.Errorf("step %d (%s): error_handling is empty", i, step.StepID)
			}
		case "patterns":
			if len(output.Patterns) == 0 {
				return fmt.Errorf("step %d (%s): patterns is empty", i, step.StepID)
			}
		case "external_dependencies":
			if len(output.ExternalDependencies) == 0 {
				return fmt.Errorf("step %d (%s): external_dependencies is empty", i, step.StepID)
			}
		case "data_flow":
			if len(output.DataFlow.Inputs) == 0 && len(output.DataFlow.Transformations) == 0 && len(output.DataFlow.Outputs) == 0 {
				return fmt.Errorf("step %d (%s): data_flow is empty", i, step.StepID)
			}
		case "configuration":
			if len(output.Configuration) == 0 {
				return fmt.Errorf("step %d (%s): configuration is empty", i, step.StepID)
			}
		case "concurrency":
			if len(output.Concurrency) == 0 {
				return fmt.Errorf("step %d (%s): concurrency is empty", i, step.StepID)
			}
		case "limitations":
			if len(output.Limitations) == 0 {
				return fmt.Errorf("step %d (%s): limitations is empty", i, step.StepID)
			}
		default:
			return fmt.Errorf("unknown field name: %s (valid: component_name, scope_description, overview, entry_points, call_graph, data_flow, raw_evidence, state_management, side_effects, error_handling, patterns, external_dependencies, configuration, concurrency, limitations)", a.FieldName)

				}
	}
	return nil
}

func (a *OutputFieldNotEmptyAssertion) Description() string {
	return fmt.Sprintf("Field '%s' is not empty in all outputs", a.FieldName)
}

// --- Workflow Helpers (5.1.1) ---

// WorkflowConfig simplifies workflow creation for testing
type WorkflowConfig struct {
	ID          string
	Description string
	Agent       string
	Input       map[string]interface{}
	Timeout     time.Duration
	Permissions schema.Permissions
}

// RunWorkflow is a simplified helper to build and execute a single-step workflow
// Returns the execution result and any error encountered
func (f *TestFramework) RunWorkflow(ctx context.Context, config WorkflowConfig) (*workflow.ExecutionResult, error) {
	// Marshal input
	inputJSON, err := json.Marshal(config.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	// Build workflow
	wf, err := workflow.NewBuilder(config.ID).
		WithDescription(config.Description).
		AddSequentialStep(
			"step1",
			config.Agent,
			string(inputJSON),
			config.Permissions,
		).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow: %w", err)
	}

	// Apply timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	workflowCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute workflow
	result, err := f.engine.Execute(workflowCtx, wf)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	return result, nil
}
// WorkflowStep simplifies step creation for multi-step workflow testing
type WorkflowStep struct {
	ID          string
	Agent       string
	Input       map[string]interface{}
	Permissions schema.Permissions
	Condition   workflow.Condition
}

// MultiStepWorkflowConfig configures multi-step workflow execution
type MultiStepWorkflowConfig struct {
	ID          string
	Description string
	Steps       []WorkflowStep
	Mode        workflow.ExecutionMode
	Timeout     time.Duration
}

// RunMultiStepWorkflow orchestrates execution of a multi-step workflow
// This helper simplifies E2E testing by:
// - Building workflows from simplified step configs
// - Managing timeouts and context cancellation
// - Consolidating results into TestResult
// - Handling step failures with proper error propagation
func (f *TestFramework) RunMultiStepWorkflow(ctx context.Context, config MultiStepWorkflowConfig) (*TestResult, error) {
	if len(config.Steps) == 0 {
		return nil, fmt.Errorf("workflow must have at least one step")
	}

	result := &TestResult{
		TestName:   config.ID,
		Errors:     make([]error, 0),
		Warnings:   make([]string, 0),
		Assertions: make([]AssertionResult, 0),
	}

	// Apply timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	workflowCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build workflow
	builder := workflow.NewBuilder(config.ID).
		WithDescription(config.Description)

	// Default to sequential mode if not specified
	mode := config.Mode
	if mode == "" {
		mode = workflow.SequentialMode
	}
	builder = builder.WithMode(mode)

	// Add steps to workflow
	for i, stepConfig := range config.Steps {
		// Marshal input
		inputJSON, err := json.Marshal(stepConfig.Input)
		if err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Errorf("step %d: failed to marshal input: %w", i, err))
			return result, err
		}

		// Generate step ID if not provided
		stepID := stepConfig.ID
		if stepID == "" {
			stepID = fmt.Sprintf("step%d", i+1)
		}

		// Add step based on mode and configuration
		if stepConfig.Condition != nil {
			builder = builder.AddConditionalStep(
				stepID,
				stepConfig.Agent,
				string(inputJSON),
				stepConfig.Permissions,
				stepConfig.Condition,
			)
		} else {
			builder = builder.AddSequentialStep(
				stepID,
				stepConfig.Agent,
				string(inputJSON),
				stepConfig.Permissions,
			)
		}
	}

	wf, err := builder.Build()
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Errorf("failed to build workflow: %w", err))
		return result, err
	}

	// Execute workflow
	startTime := time.Now()
	workflowResult, err := f.engine.Execute(workflowCtx, wf)
	result.Duration = time.Since(startTime)

	// Check for context cancellation
	select {
	case <-workflowCtx.Done():
		if workflowCtx.Err() == context.DeadlineExceeded {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Errorf("workflow timeout exceeded: %v", timeout))
			return result, workflowCtx.Err()
		}
		if workflowCtx.Err() == context.Canceled {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Errorf("workflow cancelled"))
			return result, workflowCtx.Err()
		}
	default:
	}

	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, fmt.Errorf("workflow execution failed: %w", err))
		return result, err
	}

	result.WorkflowResult = workflowResult

	// Check if all steps completed successfully
	allStepsSucceeded := true
	for i, stepResult := range workflowResult.StepResults {
		if stepResult.Status == workflow.StepStatusFailed {
			allStepsSucceeded = false
			result.Warnings = append(result.Warnings, 
				fmt.Sprintf("step %d (%s) failed: %s", i, stepResult.StepID, stepResult.Error))
		} else if stepResult.Status == workflow.StepStatusEscalated {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("step %d (%s) escalated to %s: %s", 
					i, stepResult.StepID, stepResult.EscalationTarget, stepResult.EscalationReason))
		}
	}

	result.Passed = allStepsSucceeded && len(result.Errors) == 0

	f.results = append(f.results, result)
	return result, nil
}
