package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/process"
	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// ============================================================================
// Benchmark Helpers
// ============================================================================

// createBenchmarkOrchestrator creates a fully configured orchestrator for benchmarking
func createBenchmarkOrchestrator() *Orchestrator {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	// Register mock agents with realistic behavior
	orch.RegisterAgent("codebase-locator", createFastMockAgent("codebase-locator", 10*time.Millisecond))
	orch.RegisterAgent("codebase-analyzer", createFastMockAgent("codebase-analyzer", 50*time.Millisecond))
	orch.RegisterAgent("system-architect", createFastMockAgent("system-architect", 100*time.Millisecond))

	return orch
}

// createFastMockAgent creates a mock agent with configurable latency
func createFastMockAgent(name string, delay time.Duration) AgentFactory {
	return func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				// Simulate processing time
				time.Sleep(delay)

				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   name,
					Status:    schema.StatusComplete,
					Output: &schema.AgentOutputV1{
						Version:          "AGENT_OUTPUT_V1",
						ComponentName:    "BenchmarkComponent",
						ScopeDescription: "Benchmark test scope",
						Overview:         "Benchmark test overview",
						EntryPoints: []schema.EntryPoint{
							{
								File:   "/test/main.go",
								Lines:  "10-20",
								Symbol: "main",
								Role:   "entry",
							},
						},
					},
					Timestamp: time.Now(),
				}, nil
			},
		}
	}
}

// createBenchmarkWorkflow creates a workflow with N steps
func createBenchmarkWorkflow(id string, stepCount int) (*Workflow, schema.Permissions) {
	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
	}

	steps := []WorkflowStep{}
	for i := 0; i < stepCount; i++ {
		steps = append(steps, WorkflowStep{
			AgentID: "codebase-locator",
			Request: fmt.Sprintf("Step %d input", i),
			Files:   []string{},
		})
	}

	wf := &Workflow{
		Steps: steps,
	}

	return wf, perms
}

// ============================================================================
// Benchmark: Request Routing
// Target: <100ms per routing decision
// ============================================================================

func BenchmarkRequestRouting_Simple(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	requests := []string{
		"find all .go files",
		"locate function main",
		"search for struct definitions",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		_, _ = orch.HandleRequest(ctx, req, perms)
	}
}

func BenchmarkRequestRouting_Complex(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	requests := []string{
		"find all database migration files and analyze their structure",
		"locate error handling patterns across the codebase",
		"search for unused imports and suggest cleanup",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		_, _ = orch.HandleRequest(ctx, req, perms)
	}
}

func BenchmarkRequestRouting_NoMatch(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	requests := []string{
		"what is the meaning of life",
		"tell me a joke",
		"random unrelated query",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		_, _ = orch.HandleRequest(ctx, req, perms)
	}
}

// ============================================================================
// Benchmark: Agent Invocation
// Tests single agent execution overhead
// ============================================================================

func BenchmarkAgentInvocation_Fast(b *testing.B) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	// Register fast agent (10ms)
	orch.RegisterAgent("test-agent", createFastMockAgent("test-agent", 10*time.Millisecond))

	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.HandleRequest(ctx, "find files", perms)
	}
}

func BenchmarkAgentInvocation_Slow(b *testing.B) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	// Register slower agent (100ms)
	orch.RegisterAgent("test-agent", createFastMockAgent("test-agent", 100*time.Millisecond))

	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.HandleRequest(ctx, "find files", perms)
	}
}

// ============================================================================
// Benchmark: Workflow Execution
// Tests multi-step workflow performance at different scales
// ============================================================================

func BenchmarkWorkflowExecution_2Steps(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	wf, perms := createBenchmarkWorkflow("2-step", 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.ExecuteWorkflow(ctx, wf, perms)
	}
}

func BenchmarkWorkflowExecution_5Steps(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	wf, perms := createBenchmarkWorkflow("5-step", 5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.ExecuteWorkflow(ctx, wf, perms)
	}
}

func BenchmarkWorkflowExecution_10Steps(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	wf, perms := createBenchmarkWorkflow("10-step", 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.ExecuteWorkflow(ctx, wf, perms)
	}
}

func BenchmarkWorkflowExecution_20Steps(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	wf, perms := createBenchmarkWorkflow("20-step", 20)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.ExecuteWorkflow(ctx, wf, perms)
	}
}

// ============================================================================
// Benchmark: Quality Gate Validation
// Tests validation overhead at different strictness levels
// ============================================================================

func BenchmarkQualityGates_Default(b *testing.B) {
	gates := DefaultQualityGates()
	
	// Create mock validation and profiling reports
	validationReport := &ValidationReport{
		WorkflowID:       "test-workflow",
		Timestamp:        time.Now(),
		TotalResponses:   1,
		ValidResponses:   1,
		InvalidResponses: 0,
		EvidenceCoverage: 1.0,
	}
	profilingReport := &ProfilingReport{
		WorkflowID:    "test-workflow",
		TotalDuration: 100 * time.Millisecond,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gates.CheckQualityGates(validationReport, profilingReport)
	}
}

func BenchmarkQualityGates_Strict(b *testing.B) {
	gates := StrictQualityGates()
	
	validationReport := &ValidationReport{
		WorkflowID:       "test-workflow",
		Timestamp:        time.Now(),
		TotalResponses:   1,
		ValidResponses:   1,
		InvalidResponses: 0,
		EvidenceCoverage: 1.0,
	}
	profilingReport := &ProfilingReport{
		WorkflowID:    "test-workflow",
		TotalDuration: 100 * time.Millisecond,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gates.CheckQualityGates(validationReport, profilingReport)
	}
}

func BenchmarkQualityGates_Relaxed(b *testing.B) {
	gates := RelaxedQualityGates()
	
	validationReport := &ValidationReport{
		WorkflowID:       "test-workflow",
		Timestamp:        time.Now(),
		TotalResponses:   1,
		ValidResponses:   1,
		InvalidResponses: 0,
		EvidenceCoverage: 1.0,
	}
	profilingReport := &ProfilingReport{
		WorkflowID:    "test-workflow",
		TotalDuration: 100 * time.Millisecond,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gates.CheckQualityGates(validationReport, profilingReport)
	}
}

// ============================================================================
// Benchmark: Concurrent Request Handling
// Target: Support 100+ concurrent users
// ============================================================================

func BenchmarkConcurrentRequests_10(b *testing.B) {
	benchmarkConcurrentRequests(b, 10)
}

func BenchmarkConcurrentRequests_50(b *testing.B) {
	benchmarkConcurrentRequests(b, 50)
}

func BenchmarkConcurrentRequests_100(b *testing.B) {
	benchmarkConcurrentRequests(b, 100)
}

func BenchmarkConcurrentRequests_200(b *testing.B) {
	benchmarkConcurrentRequests(b, 200)
}

func benchmarkConcurrentRequests(b *testing.B, concurrency int) {
	orch := createBenchmarkOrchestrator()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	requests := []string{
		"find all .go files",
		"analyze error handling",
		"locate main function",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		i := 0
		for pb.Next() {
			req := requests[i%len(requests)]
			_, _ = orch.HandleRequest(ctx, req, perms)
			i++
		}
	})
}

// ============================================================================
// Benchmark: Profiling Overhead
// Measures impact of profiling on execution time
// ============================================================================

func BenchmarkProfilingOverhead_WithProfiling(b *testing.B) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	
	// Create orchestrator with profiling enabled
	config := OrchestratorConfig{
		ProcessManager:  pm,
		ToolExecutor:    te,
		EnableProfiling: true,
		QualityGates:    DefaultQualityGates(),
	}
	orch := NewWithConfig(config)
	orch.RegisterAgent("test-agent", createFastMockAgent("test-agent", 10*time.Millisecond))

	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.HandleRequest(ctx, "find files", perms)
	}
}

func BenchmarkProfilingOverhead_WithoutProfiling(b *testing.B) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	
	// Create orchestrator with profiling disabled
	config := OrchestratorConfig{
		ProcessManager:  pm,
		ToolExecutor:    te,
		EnableProfiling: false,
		QualityGates:    DefaultQualityGates(),
	}
	orch := NewWithConfig(config)
	orch.RegisterAgent("test-agent", createFastMockAgent("test-agent", 10*time.Millisecond))

	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.HandleRequest(ctx, "find files", perms)
	}
}

// ============================================================================
// Benchmark: Memory Allocations
// Tests memory efficiency under various loads
// ============================================================================

func BenchmarkMemoryAllocation_SimpleRequest(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.HandleRequest(ctx, "find files", perms)
	}
}

func BenchmarkMemoryAllocation_ComplexWorkflow(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()
	wf, perms := createBenchmarkWorkflow("complex", 10)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.ExecuteWorkflow(ctx, wf, perms)
	}
}

// ============================================================================
// Benchmark: Agent Registry Operations
// Tests registry lookup performance
// ============================================================================

func BenchmarkAgentRegistry_Lookup(b *testing.B) {
	orch := createBenchmarkOrchestrator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.agentRegistry["codebase-locator"]
	}
}

func BenchmarkAgentRegistry_Register(b *testing.B) {
	pm := process.NewManager()
	te := tool.NewExecutor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		orch := New(pm, te)
		b.StartTimer()

		orch.RegisterAgent("test-agent", createFastMockAgent("test-agent", 1*time.Millisecond))
	}
}

// ============================================================================
// Benchmark: Router Performance
// Tests routing decision performance
// ============================================================================

func BenchmarkRouter_RouteDecision(b *testing.B) {
	router := &Router{}

	requests := []string{
		"find all .go files",
		"analyze this code",
		"locate function main",
		"how does this work",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := requests[i%len(requests)]
		_, _ = router.Route(req)
	}
}

// ============================================================================
// Benchmark: State Management
// Tests workflow state tracking overhead
// ============================================================================

func BenchmarkStateManagement_SmallWorkflow(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()

	workflows := make([]*Workflow, 100)
	permsSlice := make([]schema.Permissions, 100)
	for i := range workflows {
		workflows[i], permsSlice[i] = createBenchmarkWorkflow(fmt.Sprintf("wf-%d", i), 3)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(workflows)
		_, _ = orch.ExecuteWorkflow(ctx, workflows[idx], permsSlice[idx])
	}
}

func BenchmarkStateManagement_LargeWorkflow(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	ctx := context.Background()

	workflows := make([]*Workflow, 100)
	permsSlice := make([]schema.Permissions, 100)
	for i := range workflows {
		workflows[i], permsSlice[i] = createBenchmarkWorkflow(fmt.Sprintf("wf-%d", i), 20)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % len(workflows)
		_, _ = orch.ExecuteWorkflow(ctx, workflows[idx], permsSlice[idx])
	}
}

// ============================================================================
// Benchmark: Error Handling Path
// Tests error path performance
// ============================================================================

func BenchmarkErrorHandling_AgentFailure(b *testing.B) {
	pm := process.NewManager()
	te := tool.NewExecutor()
	orch := New(pm, te)

	// Register failing agent
	orch.RegisterAgent("failing-agent", func(executor *tool.Executor) Agent {
		return &MockAgent{
			ExecuteFunc: func(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
				return schema.AgentResponse{
					RequestID: req.RequestID,
					AgentID:   "failing-agent",
					Status:    schema.StatusError,
					Error: &schema.AgentError{
						Message: "simulated failure",
					},
					Timestamp: time.Now(),
				}, fmt.Errorf("simulated failure")
			},
		}
	})

	ctx := context.Background()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orch.HandleRequest(ctx, "find files", perms)
	}
}

// ============================================================================
// Benchmark: Throughput Under Load
// Tests system throughput with sustained load
// ============================================================================

func BenchmarkThroughput_Sustained(b *testing.B) {
	orch := createBenchmarkOrchestrator()
	perms := schema.Permissions{AllowedDirectories: []string{"/tmp"}}

	var wg sync.WaitGroup
	results := make(chan time.Duration, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			ctx := context.Background()
			start := time.Now()
			_, _ = orch.HandleRequest(ctx, "find files", perms)
			results <- time.Since(start)
		}(i)
	}

	wg.Wait()
	close(results)

	// Calculate statistics
	var total time.Duration
	count := 0
	for d := range results {
		total += d
		count++
	}

	if count > 0 {
		avgLatency := total / time.Duration(count)
		b.ReportMetric(float64(avgLatency.Milliseconds()), "avg_latency_ms")
	}
}
