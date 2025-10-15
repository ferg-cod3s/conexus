# Profiling Guide

This guide covers performance profiling and metrics collection in Conexus, including setup, usage, metric interpretation, and optimization strategies.

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [Core Components](#core-components)
- [Basic Usage](#basic-usage)
- [Metrics & Interpretation](#metrics--interpretation)
- [Advanced Profiling](#advanced-profiling)
- [Reporting & Visualization](#reporting--visualization)
- [Optimization Strategies](#optimization-strategies)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Conexus provides a comprehensive profiling system for monitoring and analyzing multi-agent workflow performance. The profiling system tracks:

- **Execution Metrics**: Duration, memory usage, goroutines
- **System Metrics**: Memory trends, GC activity, CPU cores
- **Performance Bottlenecks**: Slow executions, high variance
- **Aggregate Statistics**: Averages, percentiles (P50, P90, P95, P99)

### Key Features

- ✅ **Zero-overhead when disabled** - No performance impact in production
- ✅ **Phase-level tracking** - Granular performance insights
- ✅ **Automatic bottleneck detection** - Identify slow components
- ✅ **Multiple report formats** - Text, JSON, CSV
- ✅ **Real-time monitoring** - Live system metrics collection

## Getting Started

### Installation

Profiling is built into Conexus. No additional dependencies required.

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/ferg-cod3s/conexus/internal/profiling"
    "github.com/ferg-cod3s/conexus/pkg/schema"
)

func main() {
    // Create profiler (enabled)
    profiler := profiling.NewProfiler(true)
    
    // Start execution tracking
    ctx := context.Background()
    ec := profiler.StartExecution(ctx, "codebase-locator", "Find main.go")
    
    // Simulate work
    ec.StartPhase("search")
    time.Sleep(50 * time.Millisecond)
    ec.EndPhase()
    
    ec.StartPhase("parse")
    time.Sleep(30 * time.Millisecond)
    ec.EndPhase()
    
    // Complete execution
    output := &schema.AgentOutputV1{
        Version:       "AGENT_OUTPUT_V1",
        ComponentName: "codebase-locator",
    }
    ec.End(output, nil)
    
    // Generate report
    report := profiler.GetReport()
    fmt.Printf("Avg Duration: %s\n", report.OverallAvgDuration)
    fmt.Printf("Success Rate: %.2f%%\n", report.OverallSuccessRate)
}
```

### Configuration

Profiling can be controlled via environment variables or programmatically:

```go
// Enable/disable at runtime
profiler.Enable()
profiler.Disable()

// Check status
if profiler.IsEnabled() {
    // Profiling is active
}

// Clear historical data
profiler.Clear()
```

**Environment Variables**:

```bash
# Enable profiling (default: disabled in production)
CONEXUS_PROFILING_ENABLED=true

# Metrics collection interval
CONEXUS_METRICS_INTERVAL=1s
```

## Core Components

### 1. Profiler

Tracks execution metrics for individual agent runs and aggregates statistics.

```go
type Profiler struct {
    executions     map[string]*ExecutionProfile  // Individual runs
    aggregates     map[string]*AggregateMetrics  // Per-agent stats
    memoryBaseline uint64                        // Baseline memory
    enabled        bool                           // Enable/disable flag
}
```

**Key Methods**:

- `StartExecution(ctx, agent, request)` - Begin tracking
- `GetAgentMetrics(agent)` - Retrieve aggregate stats
- `GetBottlenecks(threshold)` - Identify slow components
- `GetReport()` - Generate comprehensive report

### 2. MetricsCollector

Collects system-level performance metrics at regular intervals.

```go
type MetricsCollector struct {
    snapshots []SystemSnapshot  // Time-series data
    interval  time.Duration     // Collection frequency
    running   bool              // Active/inactive
}
```

**Key Methods**:

- `Start()` - Begin background collection
- `Stop()` - Halt collection
- `GetSnapshots()` - Retrieve all data points
- `GetMemoryTrend()` - Analyze memory patterns

### 3. Reporter

Formats profiling data into various output formats.

```go
type Reporter struct {
    profiler  *Profiler
    collector *MetricsCollector
}
```

**Key Methods**:

- `WriteText(w)` - Human-readable report
- `WriteJSON(w)` - Structured JSON output
- `WriteCSV(w)` - CSV for analysis tools
- `WriteSummary(w)` - Brief performance overview

## Basic Usage

### Tracking Agent Execution

```go
// 1. Start execution
ec := profiler.StartExecution(ctx, "codebase-analyzer", "Analyze src/")

// 2. Track phases (optional)
ec.StartPhase("file_discovery")
// ... work ...
ec.EndPhase()

ec.StartPhase("parsing")
// ... work ...
ec.EndPhase()

ec.StartPhase("analysis")
// ... work ...
ec.EndPhase()

// 3. Complete execution
output, err := doWork()
ec.End(output, err)
```

### System Metrics Collection

```go
// Create collector (1 second intervals)
collector := profiling.NewMetricsCollector(1 * time.Second)

// Start background collection
collector.Start()
defer collector.Stop()

// Run workload
performLongRunningTask()

// Retrieve metrics
snapshots := collector.GetSnapshots()
average := collector.GetAverageMetrics()
trend := collector.GetMemoryTrend()

fmt.Printf("Avg Memory: %d bytes\n", average.AvgMemoryAlloc)
fmt.Printf("Memory Trend: %s\n", trend.Direction)
```

## Metrics & Interpretation

### Execution Metrics

```go
type ExecutionProfile struct {
    ID              string          // Unique execution ID
    Agent           string          // Agent name
    Duration        time.Duration   // Total execution time
    MemoryAllocated uint64          // Memory used (bytes)
    GoroutineCount  int             // Active goroutines
    Success         bool            // Success/failure
    Phases          []PhaseProfile  // Phase-level breakdown
}
```

**Key Indicators**:

- **Duration**: Total execution time. Baseline: <100ms for simple agents.
- **MemoryAllocated**: Heap allocations. Watch for >10MB on repeated executions.
- **GoroutineCount**: Concurrent operations. Sudden spikes indicate leaks.
- **Phases**: Identify which phase consumes most time.

### Aggregate Metrics

```go
type AggregateMetrics struct {
    TotalExecutions int           // Total runs
    SuccessCount    int           // Successful runs
    AvgDuration     time.Duration // Average execution time
    MinDuration     time.Duration // Fastest execution
    MaxDuration     time.Duration // Slowest execution
    Percentiles     *Percentiles  // P50, P90, P95, P99
}
```

**Interpretation**:

- **Success Rate**: Should be >95% in production.
- **AvgDuration vs MinDuration**: Large gap suggests inconsistent performance.
- **P95/P99**: Critical for understanding tail latency.

**Example Analysis**:

```go
metrics, _ := profiler.GetAgentMetrics("codebase-locator")

// Check for performance issues
if metrics.AvgDuration > 500*time.Millisecond {
    fmt.Println("WARNING: Agent is slow on average")
}

if metrics.Percentiles.P99 > 2*metrics.AvgDuration {
    fmt.Println("WARNING: High variance detected (tail latency)")
}

if float64(metrics.SuccessCount)/float64(metrics.TotalExecutions) < 0.95 {
    fmt.Println("WARNING: High failure rate")
}
```

### System Metrics

```go
type SystemSnapshot struct {
    Timestamp        time.Time  // When collected
    MemoryAlloc      uint64     // Heap allocations
    MemoryTotalAlloc uint64     // Cumulative allocations
    MemorySys        uint64     // OS memory reserved
    NumGC            uint32     // GC runs
    NumGoroutine     int        // Active goroutines
}
```

**Key Indicators**:

- **MemoryAlloc**: Current heap usage. Watch for steady growth (memory leak).
- **NumGC**: Frequent GC (>100/sec) hurts performance.
- **NumGoroutine**: Should be stable. Growth indicates leaks.

### Bottleneck Detection

```go
// Detect bottlenecks (threshold: 100ms)
bottlenecks := profiler.GetBottlenecks(100 * time.Millisecond)

for _, bn := range bottlenecks {
    fmt.Printf("[%s] %s: %s (%s)\n",
        bn.Severity,    // critical, high, medium, low
        bn.Agent,       // Which agent
        bn.Type,        // slow_execution, high_variance
        bn.AvgDuration) // How slow
}
```

**Severity Levels**:

- **Critical**: >3x threshold - Immediate attention required
- **High**: 2-3x threshold - Should be optimized soon
- **Medium**: 1.5-2x threshold - Monitor closely
- **Low**: 1-1.5x threshold - Acceptable but improving

## Advanced Profiling

### Integration Pattern 1: Workflow Gate

Use profiling to enforce performance SLAs:

```go
func executeWithSLA(profiler *profiling.Profiler, agent string, threshold time.Duration) error {
    ec := profiler.StartExecution(ctx, agent, request)
    defer func() {
        metrics, _ := profiler.GetAgentMetrics(agent)
        if metrics.AvgDuration > threshold {
            log.Warnf("SLA violation: %s took %s (threshold: %s)",
                agent, metrics.AvgDuration, threshold)
        }
    }()
    
    // Execute agent
    output, err := agent.Run()
    ec.End(output, err)
    return err
}
```

### Integration Pattern 2: Adaptive Execution

Adjust behavior based on performance trends:

```go
func adaptiveExecution(profiler *profiling.Profiler, agent string) {
    metrics, exists := profiler.GetAgentMetrics(agent)
    
    // First run: use defaults
    if !exists {
        runWithTimeout(agent, 5*time.Second)
        return
    }
    
    // Adapt timeout based on P95
    timeout := metrics.Percentiles.P95 * 2
    runWithTimeout(agent, timeout)
}
```

### Integration Pattern 3: Comparative Analysis

Compare agent performance across versions:

```go
func compareAgents(p1, p2 *profiling.Profiler, agent string) {
    m1, _ := p1.GetAgentMetrics(agent)
    m2, _ := p2.GetAgentMetrics(agent)
    
    improvement := float64(m1.AvgDuration-m2.AvgDuration) / float64(m1.AvgDuration) * 100
    fmt.Printf("Performance change: %.1f%%\n", improvement)
}
```

### Integration Pattern 4: Continuous Monitoring

Long-running monitoring with alerts:

```go
func monitorContinuously(profiler *profiling.Profiler, collector *profiling.MetricsCollector) {
    collector.Start()
    defer collector.Stop()
    
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // Check for bottlenecks
        bottlenecks := profiler.GetBottlenecks(200 * time.Millisecond)
        for _, bn := range bottlenecks {
            if bn.Severity == "critical" {
                alertOps(bn)
            }
        }
        
        // Check memory trend
        trend := collector.GetMemoryTrend()
        if trend.Direction == "increasing" && trend.Rate > 5*1024*1024 {
            alertOps(fmt.Errorf("memory leak detected: %s", trend))
        }
    }
}
```

## Reporting & Visualization

### Text Reports

Human-readable console output:

```go
reporter := profiling.NewReporter(profiler, collector)

// Write to stdout
reporter.WriteText(os.Stdout)

// Output:
// Performance Report
// =================
// 
// Generated: 2025-10-15T10:30:00Z
// Total Executions: 150
// Overall Avg Duration: 45ms
// Overall Avg Memory: 2.5 MiB
// Overall Success Rate: 98.67%
// 
// Agent Metrics
// -------------
// 
// Agent: codebase-locator
//   Total Executions: 50
//   Success: 49 (98.0%)
//   Failures: 1
//   Avg Duration: 42ms
//   Min Duration: 28ms
//   Max Duration: 156ms
//   Avg Memory: 1.8 MiB
//   Percentiles:
//     P50: 40ms
//     P90: 65ms
//     P95: 88ms
//     P99: 145ms
```

### JSON Reports

Structured data for tools:

```go
reporter.WriteJSON(os.Stdout)

// Output:
// {
//   "generated_at": "2025-10-15T10:30:00Z",
//   "total_executions": 150,
//   "overall_avg_duration": 45000000,
//   "overall_success_rate": 98.67,
//   "agent_metrics": {
//     "codebase-locator": {
//       "total_executions": 50,
//       "success_count": 49,
//       "avg_duration": 42000000,
//       "percentiles": {
//         "p50": 40000000,
//         "p90": 65000000,
//         "p95": 88000000,
//         "p99": 145000000
//       }
//     }
//   },
//   "system_metrics": {
//     "average": {
//       "avg_memory_alloc": 2621440,
//       "avg_goroutines": 12.5
//     }
//   }
// }
```

### CSV Reports

For spreadsheet analysis:

```go
reporter.WriteCSV(os.Stdout)

// Output:
// Agent,Total,Success,Failures,AvgDuration,MinDuration,MaxDuration,AvgMemory,P50,P90,P95,P99
// codebase-locator,50,49,1,42ms,28ms,156ms,1887436,40ms,65ms,88ms,145ms
// codebase-analyzer,100,99,1,48ms,30ms,200ms,2985216,45ms,75ms,95ms,180ms
```

### Summary Reports

Quick overview:

```go
reporter.WriteSummary(os.Stdout)

// Output:
// === Performance Summary ===
// Executions: 150 | Avg Duration: 45ms | Success Rate: 98.7%
// ⚠️  2 bottleneck(s) detected
```

### Execution Details

Detailed per-execution breakdown:

```go
reporter.WriteExecutionDetails(os.Stdout, "codebase-locator")

// Output:
// Execution Details: codebase-locator
// ===================================
// 
// Execution ID: codebase-locator-1728990000
//   Request: Find main.go
//   Start: 2025-10-15T10:30:00Z
//   Duration: 42ms
//   Memory: 1.8 MiB
//   Goroutines: 8
//   Success: true
//   Phases:
//     - file_discovery: 15ms
//     - parsing: 12ms
//     - analysis: 10ms
```

## Optimization Strategies

### Strategy 1: Identify Slowest Phase

```go
ec := profiler.StartExecution(ctx, "agent", "request")

ec.StartPhase("phase1")
doPhase1()
ec.EndPhase()

ec.StartPhase("phase2")
doPhase2()
ec.EndPhase()

ec.End(output, nil)

// Analyze phases
profile, _ := profiler.GetExecution(ec.executionID)
for _, phase := range profile.Phases {
    fmt.Printf("%s: %s\n", phase.Name, phase.Duration)
}
// Focus optimization on longest phase
```

### Strategy 2: Memory Optimization

```go
// Before optimization
ec1 := profiler.StartExecution(ctx, "agent-v1", "request")
runOriginalVersion()
ec1.End(output, nil)

// After optimization
profiler.Clear()
ec2 := profiler.StartExecution(ctx, "agent-v2", "request")
runOptimizedVersion()
ec2.End(output, nil)

// Compare
m1, _ := profiler.GetAgentMetrics("agent-v1")
m2, _ := profiler.GetAgentMetrics("agent-v2")

memoryReduction := float64(m1.AvgMemory-m2.AvgMemory) / float64(m1.AvgMemory) * 100
fmt.Printf("Memory reduced by %.1f%%\n", memoryReduction)
```

### Strategy 3: Bottleneck Resolution

```go
// 1. Identify bottlenecks
bottlenecks := profiler.GetBottlenecks(100 * time.Millisecond)

// 2. Focus on critical issues first
for _, bn := range bottlenecks {
    if bn.Severity == "critical" {
        fmt.Printf("Optimize %s (%s): %s\n", bn.Agent, bn.Type, bn.AvgDuration)
        
        // 3. Profile detailed execution
        reporter.WriteExecutionDetails(os.Stdout, bn.Agent)
        
        // 4. Apply optimizations based on type
        switch bn.Type {
        case "slow_execution":
            // Cache results, parallelize work
        case "high_variance":
            // Identify outliers, add timeouts
        }
    }
}
```

### Strategy 4: Load Testing

```go
func loadTest(profiler *profiling.Profiler, agent string, n int) {
    for i := 0; i < n; i++ {
        ec := profiler.StartExecution(ctx, agent, fmt.Sprintf("request-%d", i))
        runAgent()
        ec.End(output, nil)
    }
    
    // Analyze results
    metrics, _ := profiler.GetAgentMetrics(agent)
    
    fmt.Printf("Load test (%d runs):\n", n)
    fmt.Printf("  Avg: %s\n", metrics.AvgDuration)
    fmt.Printf("  P95: %s\n", metrics.Percentiles.P95)
    fmt.Printf("  P99: %s\n", metrics.Percentiles.P99)
    fmt.Printf("  Max: %s\n", metrics.MaxDuration)
}
```

## Best Practices

### 1. Enable Profiling in Development

Always profile during development to catch issues early:

```go
// Development
profiler := profiling.NewProfiler(true)

// Production (controlled)
enabled := os.Getenv("ENABLE_PROFILING") == "true"
profiler := profiling.NewProfiler(enabled)
```

### 2. Use Phase Tracking

Break complex operations into phases for granular insights:

```go
ec := profiler.StartExecution(ctx, "complex-agent", "request")

ec.StartPhase("initialization")
// ... setup ...
ec.EndPhase()

ec.StartPhase("processing")
// ... main work ...
ec.EndPhase()

ec.StartPhase("cleanup")
// ... cleanup ...
ec.EndPhase()

ec.End(output, nil)
```

### 3. Set Performance SLAs

Define acceptable thresholds:

```go
const (
    AgentSLA_Locator  = 100 * time.Millisecond
    AgentSLA_Analyzer = 500 * time.Millisecond
    AgentSLA_Reporter = 200 * time.Millisecond
)

func checkSLA(profiler *profiling.Profiler, agent string, sla time.Duration) bool {
    metrics, _ := profiler.GetAgentMetrics(agent)
    return metrics.AvgDuration <= sla
}
```

### 4. Monitor Continuously in Production

Use lightweight monitoring:

```go
// Only check periodically
if time.Now().Unix()%300 == 0 { // Every 5 minutes
    report := profiler.GetReport()
    if report.OverallSuccessRate < 95.0 {
        log.Warn("Success rate dropped below 95%")
    }
}
```

### 5. Clear Data Regularly

Prevent unbounded memory growth:

```go
// Clear after generating report
report := profiler.GetReport()
saveReport(report)
profiler.Clear()
```

## Troubleshooting

### Issue 1: High Memory Usage

**Symptoms**: MemoryAlloc grows continuously, GC frequency increases.

**Diagnosis**:

```go
trend := collector.GetMemoryTrend()
if trend.Direction == "increasing" {
    fmt.Printf("Memory leak detected!\n")
    fmt.Printf("  Rate: %.2f MB/s\n", trend.Rate/1024/1024)
    fmt.Printf("  Start: %d bytes\n", trend.StartMemory)
    fmt.Printf("  End: %d bytes\n", trend.EndMemory)
}
```

**Solutions**:
- Clear profiler data regularly: `profiler.Clear()`
- Check for goroutine leaks: `runtime.NumGoroutine()`
- Use pprof for detailed analysis

### Issue 2: Slow Execution

**Symptoms**: AvgDuration exceeds SLA, P95/P99 are very high.

**Diagnosis**:

```go
bottlenecks := profiler.GetBottlenecks(100 * time.Millisecond)
for _, bn := range bottlenecks {
    fmt.Printf("%s is slow: %s\n", bn.Agent, bn.AvgDuration)
}

// Get detailed profile
profile, _ := profiler.GetExecution(executionID)
for _, phase := range profile.Phases {
    fmt.Printf("  %s: %s\n", phase.Name, phase.Duration)
}
```

**Solutions**:
- Identify slowest phase
- Add caching
- Parallelize independent operations
- Reduce allocations

### Issue 3: High Variance (Tail Latency)

**Symptoms**: P99 >> P50, inconsistent performance.

**Diagnosis**:

```go
metrics, _ := profiler.GetAgentMetrics("agent")
variance := float64(metrics.MaxDuration) / float64(metrics.MinDuration)

if variance > 5.0 {
    fmt.Printf("High variance detected: %s to %s (%.1fx)\n",
        metrics.MinDuration, metrics.MaxDuration, variance)
}
```

**Solutions**:
- Add timeouts to prevent outliers
- Identify external dependencies causing delays
- Use circuit breakers
- Pre-warm caches

### Issue 4: Profiler Overhead

**Symptoms**: Profiling itself impacts performance.

**Diagnosis**:

```go
// Run with profiling
p1 := profiling.NewProfiler(true)
start := time.Now()
runWorkload(p1)
withProfiling := time.Since(start)

// Run without profiling
p2 := profiling.NewProfiler(false)
start = time.Now()
runWorkload(p2)
withoutProfiling := time.Since(start)

overhead := float64(withProfiling-withoutProfiling) / float64(withoutProfiling) * 100
fmt.Printf("Profiling overhead: %.2f%%\n", overhead)
```

**Solutions**:
- Disable profiling in hot paths: `profiler.Disable()`
- Use sampling (profile 1/10 executions)
- Reduce phase granularity

## Summary

Conexus profiling provides comprehensive performance insights with:

- **Zero overhead when disabled** - Safe for production
- **Granular tracking** - Execution, phase, and system metrics
- **Automatic analysis** - Bottleneck detection, percentiles
- **Flexible reporting** - Text, JSON, CSV formats
- **Real-time monitoring** - Live system metrics

**Key Takeaways**:

1. Always profile during development
2. Use phase tracking for complex agents
3. Set and monitor performance SLAs
4. Clear profiler data regularly
5. Focus on P95/P99 for user experience

**Next Steps**:

- Review [Validation Guide](./validation-guide.md) for quality gates
- See [Architecture Guide](./architecture/integration.md) for integration patterns
- Check [API Reference](./api-reference.md) for detailed API docs
