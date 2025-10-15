package profiling

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestNewReporter(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)

	r := NewReporter(p, mc)
	if r == nil {
		t.Fatal("expected reporter, got nil")
	}

	if r.profiler != p {
		t.Error("expected profiler to be set")
	}

	if r.collector != mc {
		t.Error("expected collector to be set")
	}
}

func TestReporter_WriteText(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)
	r := NewReporter(p, mc)

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Add some executions
	for i := 0; i < 3; i++ {
		ec := p.StartExecution(ctx, "test-agent", "request")
		ec.End(output, nil)
	}

	// Start collector and capture some snapshots
	mc.Start()
	time.Sleep(150 * time.Millisecond)
	mc.Stop()

	var buf bytes.Buffer
	err := r.WriteText(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output_text := buf.String()

	// Verify key sections are present
	if !strings.Contains(output_text, "Performance Report") {
		t.Error("expected 'Performance Report' header")
	}

	if !strings.Contains(output_text, "Total Executions") {
		t.Error("expected 'Total Executions' field")
	}

	if !strings.Contains(output_text, "Agent Metrics") {
		t.Error("expected 'Agent Metrics' section")
	}

	if !strings.Contains(output_text, "test-agent") {
		t.Error("expected agent name in output")
	}

	if !strings.Contains(output_text, "System Metrics") {
		t.Error("expected 'System Metrics' section")
	}
}

func TestReporter_WriteJSON(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)
	r := NewReporter(p, mc)

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Add executions
	ec := p.StartExecution(ctx, "test-agent", "request")
	ec.End(output, nil)

	// Capture system metrics
	mc.captureSnapshot()

	var buf bytes.Buffer
	err := r.WriteJSON(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify structure
	if _, ok := result["total_executions"]; !ok {
		t.Error("expected total_executions field")
	}

	if _, ok := result["agent_metrics"]; !ok {
		t.Error("expected agent_metrics field")
	}

	if _, ok := result["system_metrics"]; !ok {
		t.Error("expected system_metrics field")
	}
}

func TestReporter_WriteCSV(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)
	r := NewReporter(p, mc)

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Add executions for multiple agents
	for i := 0; i < 5; i++ {
		ec := p.StartExecution(ctx, "agent-1", "request")
		ec.End(output, nil)
	}

	for i := 0; i < 3; i++ {
		ec := p.StartExecution(ctx, "agent-2", "request")
		ec.End(output, nil)
	}

	var buf bytes.Buffer
	err := r.WriteCSV(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	csv := buf.String()
	lines := strings.Split(strings.TrimSpace(csv), "\n")

	// Verify header + 2 data rows
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (header + 2 data), got %d", len(lines))
	}

	// Verify header
	if !strings.Contains(lines[0], "Agent") {
		t.Error("expected Agent column in header")
	}

	if !strings.Contains(lines[0], "Total") {
		t.Error("expected Total column in header")
	}

	// Verify data contains agent names
	csvData := strings.Join(lines[1:], "\n")
	if !strings.Contains(csvData, "agent-1") {
		t.Error("expected agent-1 in CSV data")
	}

	if !strings.Contains(csvData, "agent-2") {
		t.Error("expected agent-2 in CSV data")
	}
}

func TestReporter_WriteExecutionDetails(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)
	r := NewReporter(p, mc)

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Add execution with phases
	ec := p.StartExecution(ctx, "test-agent", "detailed request")
	ec.StartPhase("phase1")
	time.Sleep(10 * time.Millisecond)
	ec.EndPhase()
	ec.End(output, nil)

	var buf bytes.Buffer
	err := r.WriteExecutionDetails(&buf, "test-agent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	details := buf.String()

	// Verify content
	if !strings.Contains(details, "Execution Details") {
		t.Error("expected 'Execution Details' header")
	}

	if !strings.Contains(details, "detailed request") {
		t.Error("expected request text in details")
	}

	if !strings.Contains(details, "phase1") {
		t.Error("expected phase information")
	}

	if !strings.Contains(details, "Summary") {
		t.Error("expected summary section")
	}
}

func TestReporter_WriteExecutionDetails_AgentNotFound(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)
	r := NewReporter(p, mc)

	var buf bytes.Buffer
	err := r.WriteExecutionDetails(&buf, "nonexistent-agent")
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestReporter_WriteSummary(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)
	r := NewReporter(p, mc)

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Add execution
	ec := p.StartExecution(ctx, "test-agent", "request")
	ec.End(output, nil)

	var buf bytes.Buffer
	err := r.WriteSummary(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summary := buf.String()

	if !strings.Contains(summary, "Performance Summary") {
		t.Error("expected 'Performance Summary' header")
	}

	if !strings.Contains(summary, "Executions") {
		t.Error("expected execution count")
	}

	if !strings.Contains(summary, "Success Rate") {
		t.Error("expected success rate")
	}
}

func TestReporter_WriteSummary_WithBottlenecks(t *testing.T) {
	p := NewProfiler(true)
	mc := NewMetricsCollector(100 * time.Millisecond)
	r := NewReporter(p, mc)

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Create slow agent to trigger bottleneck
	for i := 0; i < 5; i++ {
		ec := p.StartExecution(ctx, "slow-agent", "request")
		time.Sleep(200 * time.Millisecond)
		ec.End(output, nil)
	}

	var buf bytes.Buffer
	err := r.WriteSummary(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	summary := buf.String()

	if !strings.Contains(summary, "bottleneck") {
		t.Error("expected bottleneck warning in summary")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1.50 KiB"},
		{"megabytes", 1572864, "1.50 MiB"},
		{"gigabytes", 1610612736, "1.50 GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %v, want %v", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestReporter_WithoutCollector(t *testing.T) {
	p := NewProfiler(true)
	r := NewReporter(p, nil) // No collector

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	ec := p.StartExecution(ctx, "test-agent", "request")
	ec.End(output, nil)

	var buf bytes.Buffer
	err := r.WriteText(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := buf.String()

	// Should still work without system metrics
	if !strings.Contains(text, "Performance Report") {
		t.Error("expected report header")
	}

	// System metrics section should be absent or empty
	if strings.Contains(text, "Avg Memory Alloc") {
		t.Error("unexpected system metrics without collector")
	}
}

func TestReporter_CompleteWorkflow(t *testing.T) {
	// Integration test: full profiling and reporting workflow
	p := NewProfiler(true)
	mc := NewMetricsCollector(50 * time.Millisecond)
	r := NewReporter(p, mc)

	ctx := context.Background()
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
	}

	// Start system metrics collection
	mc.Start()
	defer mc.Stop()

	// Execute multiple agents
	agents := []string{"locator", "analyzer", "pattern-finder"}
	for _, agent := range agents {
		for i := 0; i < 3; i++ {
			ec := p.StartExecution(ctx, agent, "test request")
			ec.StartPhase("setup")
			time.Sleep(5 * time.Millisecond)
			ec.EndPhase()

			ec.StartPhase("execute")
			time.Sleep(10 * time.Millisecond)
			ec.EndPhase()

			ec.End(output, nil)
		}
	}

	// Wait for system metrics
	time.Sleep(100 * time.Millisecond)

	// Generate all report formats
	t.Run("text report", func(t *testing.T) {
		var buf bytes.Buffer
		if err := r.WriteText(&buf); err != nil {
			t.Errorf("WriteText failed: %v", err)
		}
		if buf.Len() == 0 {
			t.Error("expected non-empty text report")
		}
	})

	t.Run("json report", func(t *testing.T) {
		var buf bytes.Buffer
		if err := r.WriteJSON(&buf); err != nil {
			t.Errorf("WriteJSON failed: %v", err)
		}
		if buf.Len() == 0 {
			t.Error("expected non-empty JSON report")
		}
	})

	t.Run("csv report", func(t *testing.T) {
		var buf bytes.Buffer
		if err := r.WriteCSV(&buf); err != nil {
			t.Errorf("WriteCSV failed: %v", err)
		}
		if buf.Len() == 0 {
			t.Error("expected non-empty CSV report")
		}
	})

	t.Run("summary", func(t *testing.T) {
		var buf bytes.Buffer
		if err := r.WriteSummary(&buf); err != nil {
			t.Errorf("WriteSummary failed: %v", err)
		}
		if buf.Len() == 0 {
			t.Error("expected non-empty summary")
		}
	})
}
