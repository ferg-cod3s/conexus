// Package profiling provides reporting utilities for performance data.
package profiling

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Reporter formats and exports profiling data
type Reporter struct {
	profiler  *Profiler
	collector *MetricsCollector
}

// NewReporter creates a new profiling reporter
func NewReporter(profiler *Profiler, collector *MetricsCollector) *Reporter {
	return &Reporter{
		profiler:  profiler,
		collector: collector,
	}
}

// WriteText writes a human-readable text report
func (r *Reporter) WriteText(w io.Writer) error {
	report := r.profiler.GetReport()

	fmt.Fprintf(w, "Performance Report\n")
	fmt.Fprintf(w, "=================\n\n")
	fmt.Fprintf(w, "Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Total Executions: %d\n", report.TotalExecutions)
	fmt.Fprintf(w, "Overall Avg Duration: %s\n", report.OverallAvgDuration)
	fmt.Fprintf(w, "Overall Avg Memory: %s\n", formatBytes(report.OverallAvgMemory))
	fmt.Fprintf(w, "Overall Success Rate: %.2f%%\n\n", report.OverallSuccessRate)

	// Agent metrics
	fmt.Fprintf(w, "Agent Metrics\n")
	fmt.Fprintf(w, "-------------\n")
	for agent, metrics := range report.AgentMetrics {
		fmt.Fprintf(w, "\nAgent: %s\n", agent)
		fmt.Fprintf(w, "  Total Executions: %d\n", metrics.TotalExecutions)
		fmt.Fprintf(w, "  Success: %d (%.1f%%)\n", metrics.SuccessCount, float64(metrics.SuccessCount)/float64(metrics.TotalExecutions)*100)
		fmt.Fprintf(w, "  Failures: %d\n", metrics.FailureCount)
		fmt.Fprintf(w, "  Avg Duration: %s\n", metrics.AvgDuration)
		fmt.Fprintf(w, "  Min Duration: %s\n", metrics.MinDuration)
		fmt.Fprintf(w, "  Max Duration: %s\n", metrics.MaxDuration)
		fmt.Fprintf(w, "  Avg Memory: %s\n", formatBytes(metrics.AvgMemory))

		if metrics.Percentiles != nil {
			fmt.Fprintf(w, "  Percentiles:\n")
			fmt.Fprintf(w, "    P50: %s\n", metrics.Percentiles.P50)
			fmt.Fprintf(w, "    P90: %s\n", metrics.Percentiles.P90)
			fmt.Fprintf(w, "    P95: %s\n", metrics.Percentiles.P95)
			fmt.Fprintf(w, "    P99: %s\n", metrics.Percentiles.P99)
		}
	}

	// Bottlenecks
	if len(report.Bottlenecks) > 0 {
		fmt.Fprintf(w, "\nBottlenecks\n")
		fmt.Fprintf(w, "-----------\n")
		for _, bottleneck := range report.Bottlenecks {
			fmt.Fprintf(w, "  [%s] %s: %s (%s)\n",
				strings.ToUpper(bottleneck.Severity),
				bottleneck.Agent,
				bottleneck.Type,
				bottleneck.AvgDuration)
		}
	}

	// System metrics
	if r.collector != nil {
		fmt.Fprintf(w, "\nSystem Metrics\n")
		fmt.Fprintf(w, "--------------\n")

		if avg := r.collector.GetAverageMetrics(); avg != nil {
			fmt.Fprintf(w, "  Avg Memory Alloc: %s\n", formatBytes(avg.AvgMemoryAlloc))
			fmt.Fprintf(w, "  Avg Memory Sys: %s\n", formatBytes(avg.AvgMemorySys))
			fmt.Fprintf(w, "  Avg Goroutines: %.1f\n", avg.AvgGoroutines)
			fmt.Fprintf(w, "  Sample Count: %d\n", avg.SampleCount)
		}

		trend := r.collector.GetMemoryTrend()
		fmt.Fprintf(w, "\n  Memory Trend: %s\n", trend.Direction)
		if trend.Direction != "stable" {
			fmt.Fprintf(w, "    Rate: %.2f MB/s\n", trend.Rate/1024/1024)
			fmt.Fprintf(w, "    Start: %s\n", formatBytes(trend.StartMemory))
			fmt.Fprintf(w, "    End: %s\n", formatBytes(trend.EndMemory))
			fmt.Fprintf(w, "    Duration: %s\n", trend.Duration)
		}
	}

	return nil
}

// WriteJSON writes a JSON-formatted report
func (r *Reporter) WriteJSON(w io.Writer) error {
	report := r.profiler.GetReport()

	// Create enhanced report with system metrics
	enhancedReport := struct {
		*PerformanceReport
		SystemMetrics *SystemMetricsReport `json:"system_metrics,omitempty"`
	}{
		PerformanceReport: report,
	}

	if r.collector != nil {
		enhancedReport.SystemMetrics = &SystemMetricsReport{
			Average: r.collector.GetAverageMetrics(),
			Trend:   r.collector.GetMemoryTrend(),
		}
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(enhancedReport)
}

// SystemMetricsReport contains system-level metrics
type SystemMetricsReport struct {
	Average *AverageMetrics `json:"average"`
	Trend   MemoryTrend     `json:"trend"`
}

// WriteCSV writes agent metrics in CSV format
func (r *Reporter) WriteCSV(w io.Writer) error {
	report := r.profiler.GetReport()

	// Header
	fmt.Fprintf(w, "Agent,Total,Success,Failures,AvgDuration,MinDuration,MaxDuration,AvgMemory,P50,P90,P95,P99\n")

	// Data
	for agent, metrics := range report.AgentMetrics {
		fmt.Fprintf(w, "%s,%d,%d,%d,%s,%s,%s,%d",
			agent,
			metrics.TotalExecutions,
			metrics.SuccessCount,
			metrics.FailureCount,
			metrics.AvgDuration,
			metrics.MinDuration,
			metrics.MaxDuration,
			metrics.AvgMemory)

		if metrics.Percentiles != nil {
			fmt.Fprintf(w, ",%s,%s,%s,%s\n",
				metrics.Percentiles.P50,
				metrics.Percentiles.P90,
				metrics.Percentiles.P95,
				metrics.Percentiles.P99)
		} else {
			fmt.Fprintf(w, ",,,\n")
		}
	}

	return nil
}

// WriteExecutionDetails writes detailed execution profiles
func (r *Reporter) WriteExecutionDetails(w io.Writer, agent string) error {
	metrics, ok := r.profiler.GetAgentMetrics(agent)
	if !ok {
		return fmt.Errorf("no metrics found for agent: %s", agent)
	}

	fmt.Fprintf(w, "Execution Details: %s\n", agent)
	fmt.Fprintf(w, "==================%s\n\n", strings.Repeat("=", len(agent)))

	// Get all executions for this agent
	r.profiler.mu.RLock()
	defer r.profiler.mu.RUnlock()

	for _, exec := range r.profiler.executions {
		if exec.Agent != agent {
			continue
		}

		fmt.Fprintf(w, "Execution ID: %s\n", exec.ID)
		fmt.Fprintf(w, "  Request: %s\n", exec.Request)
		fmt.Fprintf(w, "  Start: %s\n", exec.StartTime.Format(time.RFC3339))
		fmt.Fprintf(w, "  Duration: %s\n", exec.Duration)
		fmt.Fprintf(w, "  Memory: %s\n", formatBytes(exec.MemoryAllocated))
		fmt.Fprintf(w, "  Goroutines: %d\n", exec.GoroutineCount)
		fmt.Fprintf(w, "  Success: %v\n", exec.Success)

		if exec.Error != nil {
			fmt.Fprintf(w, "  Error: %s\n", exec.Error)
		}

		if len(exec.Phases) > 0 {
			fmt.Fprintf(w, "  Phases:\n")
			for _, phase := range exec.Phases {
				fmt.Fprintf(w, "    - %s: %s\n", phase.Name, phase.Duration)
			}
		}

		fmt.Fprintf(w, "\n")
	}

	// Summary
	fmt.Fprintf(w, "Summary\n")
	fmt.Fprintf(w, "-------\n")
	fmt.Fprintf(w, "Total Executions: %d\n", metrics.TotalExecutions)
	fmt.Fprintf(w, "Success Rate: %.2f%%\n", float64(metrics.SuccessCount)/float64(metrics.TotalExecutions)*100)

	return nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// WriteSummary writes a brief summary of performance
func (r *Reporter) WriteSummary(w io.Writer) error {
	report := r.profiler.GetReport()

	fmt.Fprintf(w, "=== Performance Summary ===\n")
	fmt.Fprintf(w, "Executions: %d | Avg Duration: %s | Success Rate: %.1f%%\n",
		report.TotalExecutions,
		report.OverallAvgDuration,
		report.OverallSuccessRate)

	if len(report.Bottlenecks) > 0 {
		fmt.Fprintf(w, "⚠️  %d bottleneck(s) detected\n", len(report.Bottlenecks))
	}

	return nil
}
