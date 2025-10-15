package profiling

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestNewProfiler(t *testing.T) {
	p := NewProfiler(true)
	if p == nil {
		t.Fatal("expected profiler, got nil")
	}

	if !p.enabled {
		t.Error("expected profiler to be enabled")
	}

	if p.executions == nil {
		t.Error("expected executions map to be initialized")
	}

	if p.aggregates == nil {
		t.Error("expected aggregates map to be initialized")
	}
}

func TestProfiler_StartEnd(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	ec := p.StartExecution(ctx, "test-agent", "test request")
	if ec == nil {
		t.Fatal("expected execution context")
	}

	if !ec.enabled {
		t.Error("expected execution context to be enabled")
	}

	// Create output
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// End execution
	ec.End(output, nil)

	// Verify execution was recorded
	p.mu.RLock()
	if len(p.executions) != 1 {
		t.Errorf("expected 1 execution, got %d", len(p.executions))
	}
	p.mu.RUnlock()

	// Verify metrics were recorded
	metrics, ok := p.GetAgentMetrics("test-agent")
	if !ok {
		t.Fatal("expected metrics for test-agent")
	}

	if metrics.TotalExecutions != 1 {
		t.Errorf("expected 1 execution, got %d", metrics.TotalExecutions)
	}

	if metrics.SuccessCount != 1 {
		t.Errorf("expected 1 success, got %d", metrics.SuccessCount)
	}
}

func TestProfiler_Disabled(t *testing.T) {
	p := NewProfiler(false)
	ctx := context.Background()

	ec := p.StartExecution(ctx, "test-agent", "test request")
	if ec.enabled {
		t.Error("expected execution context to be disabled")
	}

	ec.End(nil, nil)

	// No executions should be recorded
	p.mu.RLock()
	count := len(p.executions)
	p.mu.RUnlock()

	if count != 0 {
		t.Errorf("expected 0 executions when disabled, got %d", count)
	}
}

func TestExecutionContext_Phases(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	ec := p.StartExecution(ctx, "test-agent", "test request")

	// Start phases
	ec.StartPhase("phase1")
	time.Sleep(10 * time.Millisecond)
	ec.EndPhase()

	ec.StartPhase("phase2")
	time.Sleep(10 * time.Millisecond)
	ec.EndPhase()

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	ec.End(output, nil)

	// Verify phases were recorded
	profile, ok := p.GetExecution(ec.executionID)
	if !ok {
		t.Fatal("expected execution profile")
	}

	if len(profile.Phases) != 2 {
		t.Errorf("expected 2 phases, got %d", len(profile.Phases))
	}

	if profile.Phases[0].Name != "phase1" {
		t.Errorf("expected phase1, got %s", profile.Phases[0].Name)
	}

	if profile.Phases[1].Name != "phase2" {
		t.Errorf("expected phase2, got %s", profile.Phases[1].Name)
	}
}

func TestProfiler_ErrorHandling(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	ec := p.StartExecution(ctx, "test-agent", "test request")

	testErr := errors.New("test error")
	ec.End(nil, testErr)

	// Verify error was recorded
	profile, ok := p.GetExecution(ec.executionID)
	if !ok {
		t.Fatal("expected execution profile")
	}

	if profile.Success {
		t.Error("expected execution to be marked as failed")
	}

	if profile.Error != testErr {
		t.Errorf("expected error %v, got %v", testErr, profile.Error)
	}

	// Verify metrics reflect failure
	metrics, ok := p.GetAgentMetrics("test-agent")
	if !ok {
		t.Fatal("expected metrics")
	}

	if metrics.FailureCount != 1 {
		t.Errorf("expected 1 failure, got %d", metrics.FailureCount)
	}
}

func TestProfiler_MultipleExecutions(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Execute multiple times with varying durations
	for i := 0; i < 5; i++ {
		ec := p.StartExecution(ctx, "test-agent", "request")
		time.Sleep(time.Duration(i*5) * time.Millisecond)
		ec.End(output, nil)
	}

	// Verify metrics
	metrics, ok := p.GetAgentMetrics("test-agent")
	if !ok {
		t.Fatal("expected metrics")
	}

	if metrics.TotalExecutions != 5 {
		t.Errorf("expected 5 executions, got %d", metrics.TotalExecutions)
	}

	if metrics.MinDuration == 0 {
		t.Error("expected non-zero min duration")
	}

	if metrics.MaxDuration == 0 {
		t.Error("expected non-zero max duration")
	}

	if metrics.MaxDuration <= metrics.MinDuration {
		t.Error("expected max duration > min duration")
	}
}

func TestProfiler_Percentiles(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Execute enough times to calculate percentiles
	for i := 0; i < 100; i++ {
		ec := p.StartExecution(ctx, "test-agent", "request")
		time.Sleep(time.Duration(i) * time.Microsecond)
		ec.End(output, nil)
	}

	metrics, ok := p.GetAgentMetrics("test-agent")
	if !ok {
		t.Fatal("expected metrics")
	}

	if metrics.Percentiles == nil {
		t.Fatal("expected percentiles to be calculated")
	}

	if metrics.Percentiles.P50 == 0 {
		t.Error("expected non-zero P50")
	}

	if metrics.Percentiles.P90 == 0 {
		t.Error("expected non-zero P90")
	}

	if metrics.Percentiles.P95 == 0 {
		t.Error("expected non-zero P95")
	}

	if metrics.Percentiles.P99 == 0 {
		t.Error("expected non-zero P99")
	}

	// Verify percentiles are in ascending order
	if metrics.Percentiles.P50 > metrics.Percentiles.P90 {
		t.Error("expected P50 <= P90")
	}

	if metrics.Percentiles.P90 > metrics.Percentiles.P95 {
		t.Error("expected P90 <= P95")
	}

	if metrics.Percentiles.P95 > metrics.Percentiles.P99 {
		t.Error("expected P95 <= P99")
	}
}

func TestProfiler_Bottlenecks(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Create slow agent
	for i := 0; i < 5; i++ {
		ec := p.StartExecution(ctx, "slow-agent", "request")
		time.Sleep(200 * time.Millisecond)
		ec.End(output, nil)
	}

	// Create fast agent
	for i := 0; i < 5; i++ {
		ec := p.StartExecution(ctx, "fast-agent", "request")
		ec.End(output, nil)
	}

	// Get bottlenecks with 100ms threshold
	bottlenecks := p.GetBottlenecks(100 * time.Millisecond)

	if len(bottlenecks) == 0 {
		t.Fatal("expected bottlenecks to be identified")
	}

	// Verify slow agent is identified
	foundSlowAgent := false
	for _, b := range bottlenecks {
		if b.Agent == "slow-agent" {
			foundSlowAgent = true
			if b.Severity == "" {
				t.Error("expected severity to be set")
			}
		}
	}

	if !foundSlowAgent {
		t.Error("expected slow-agent to be identified as bottleneck")
	}
}

func TestProfiler_Report(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Execute some operations
	for i := 0; i < 3; i++ {
		ec := p.StartExecution(ctx, "agent-1", "request")
		ec.End(output, nil)
	}

	for i := 0; i < 2; i++ {
		ec := p.StartExecution(ctx, "agent-2", "request")
		ec.End(output, nil)
	}

	// Generate report
	report := p.GetReport()

	if report.TotalExecutions != 5 {
		t.Errorf("expected 5 total executions, got %d", report.TotalExecutions)
	}

	if len(report.AgentMetrics) != 2 {
		t.Errorf("expected 2 agents, got %d", len(report.AgentMetrics))
	}

	if report.OverallSuccessRate != 100.0 {
		t.Errorf("expected 100%% success rate, got %.2f%%", report.OverallSuccessRate)
	}
}

func TestProfiler_Clear(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Add some executions
	ec := p.StartExecution(ctx, "test-agent", "request")
	ec.End(output, nil)

	// Verify data exists
	if len(p.executions) == 0 {
		t.Fatal("expected executions before clear")
	}

	// Clear
	p.Clear()

	// Verify data is cleared
	if len(p.executions) != 0 {
		t.Errorf("expected 0 executions after clear, got %d", len(p.executions))
	}

	if len(p.aggregates) != 0 {
		t.Errorf("expected 0 aggregates after clear, got %d", len(p.aggregates))
	}
}

func TestProfiler_EnableDisable(t *testing.T) {
	p := NewProfiler(true)

	if !p.IsEnabled() {
		t.Error("expected profiler to be enabled")
	}

	p.Disable()

	if p.IsEnabled() {
		t.Error("expected profiler to be disabled")
	}

	p.Enable()

	if !p.IsEnabled() {
		t.Error("expected profiler to be enabled after re-enable")
	}
}

func TestCalculateSeverity(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		threshold time.Duration
		want      string
	}{
		{"low", 1100 * time.Millisecond, 1000 * time.Millisecond, "low"},
		{"medium", 1600 * time.Millisecond, 1000 * time.Millisecond, "medium"},
		{"high", 2500 * time.Millisecond, 1000 * time.Millisecond, "high"},
		{"critical", 4000 * time.Millisecond, 1000 * time.Millisecond, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateSeverity(tt.duration, tt.threshold)
			if got != tt.want {
				t.Errorf("calculateSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfiler_GetAllMetrics(t *testing.T) {
	p := NewProfiler(true)
	ctx := context.Background()

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Add executions for multiple agents
	agents := []string{"agent-1", "agent-2", "agent-3"}
	for _, agent := range agents {
		ec := p.StartExecution(ctx, agent, "request")
		ec.End(output, nil)
	}

	// Get all metrics
	allMetrics := p.GetAllMetrics()

	if len(allMetrics) != len(agents) {
		t.Errorf("expected %d agents, got %d", len(agents), len(allMetrics))
	}

	for _, agent := range agents {
		if _, ok := allMetrics[agent]; !ok {
			t.Errorf("expected metrics for %s", agent)
		}
	}
}
