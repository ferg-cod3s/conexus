package orchestrator

import (
	"context"
	"time"

	"github.com/ferg-cod3s/conexus/internal/profiling"
)

// WorkflowProfiler wraps profiling functionality for orchestrator
type WorkflowProfiler struct {
	profiler       *profiling.Profiler
	workflowID     string
	startTime      time.Time
	execContexts   map[string]*profiling.ExecutionContext
}

// NewWorkflowProfiler creates a new workflow profiler
func NewWorkflowProfiler(workflowID string, enabled bool) *WorkflowProfiler {
	return &WorkflowProfiler{
		profiler:     profiling.NewProfiler(enabled),
		workflowID:   workflowID,
		startTime:    time.Now(),
		execContexts: make(map[string]*profiling.ExecutionContext),
	}
}

// StartAgentExecution begins profiling an agent execution
func (wp *WorkflowProfiler) StartAgentExecution(
	ctx context.Context,
	agentID string,
	request string,
) *profiling.ExecutionContext {
	execCtx := wp.profiler.StartExecution(ctx, agentID, request)
	wp.execContexts[agentID] = execCtx
	return execCtx
}

// GenerateReport creates a profiling report for the workflow
func (wp *WorkflowProfiler) GenerateReport() *ProfilingReport {
	totalDuration := time.Since(wp.startTime)
	
	report := &ProfilingReport{
		WorkflowID:      wp.workflowID,
		Timestamp:       time.Now(),
		TotalDuration:   totalDuration,
		AgentExecutions: make([]AgentExecutionMetrics, 0),
	}

	// Get all aggregate metrics from profiler
	aggregates := wp.profiler.GetAllMetrics()
	
	var totalMemory uint64
	var peakMemory uint64
	var profiledDuration time.Duration

	for agentID, agg := range aggregates {
		if agg.TotalExecutions == 0 {
			continue
		}

		metrics := AgentExecutionMetrics{
			Agent:           agentID,
			RequestID:       agentID, // Use agent ID as request ID for aggregates
			Duration:        agg.AvgDuration,
			MemoryAllocated: agg.AvgMemory,
			MemoryFreed:     0, // Not tracked separately
			Success:         agg.SuccessCount > 0,
		}

		report.AgentExecutions = append(report.AgentExecutions, metrics)

		// Track memory usage
		totalMemory += agg.TotalMemory
		if agg.AvgMemory > peakMemory {
			peakMemory = agg.AvgMemory
		}
		
		// Sum total profiled duration
		profiledDuration += agg.TotalDuration
	}

	report.TotalMemoryUsed = totalMemory
	report.PeakMemoryUsage = peakMemory

	// Calculate profiling overhead
	if totalDuration > 0 {
		overhead := float64(totalDuration-profiledDuration) / float64(totalDuration) * 100
		report.ProfilingOverhead = overhead
	}

	return report
}

// GetProfiler returns the underlying profiler
func (wp *WorkflowProfiler) GetProfiler() *profiling.Profiler {
	return wp.profiler
}
