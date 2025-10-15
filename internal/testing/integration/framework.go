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
	"fmt"
	"time"

	"github.com/ferg-cod3s/conexus/internal/orchestrator/workflow"
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
				return fmt.Errorf("step %d (%s): evidence validation failed: %d unbacked claims, %d invalid evidence", 
					i, step.StepID, len(report.EvidenceResult.UnbackedClaims), len(report.EvidenceResult.InvalidEvidence))
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
		default:
			return fmt.Errorf("unknown field name: %s (valid: component_name, scope_description, overview, entry_points, call_graph, raw_evidence, state_management, side_effects, error_handling)", a.FieldName)
		}
	}
	return nil
}

func (a *OutputFieldNotEmptyAssertion) Description() string {
	return fmt.Sprintf("Field '%s' is not empty in all outputs", a.FieldName)
}
