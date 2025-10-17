// Package profiling provides metrics collection utilities.
package profiling

import (
	"math"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector collects system-level performance metrics
type MetricsCollector struct {
	mu              sync.RWMutex
	snapshots       []SystemSnapshot
	interval        time.Duration
	stopChan        chan struct{}
	running         bool
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(interval time.Duration) *MetricsCollector {
	return &MetricsCollector{
		snapshots: make([]SystemSnapshot, 0),
		interval:  interval,
		stopChan:  make(chan struct{}),
	}
}

// SystemSnapshot contains system metrics at a point in time
type SystemSnapshot struct {
	Timestamp       time.Time
	MemoryAlloc     uint64
	MemoryTotalAlloc uint64
	MemorySys       uint64
	NumGC           uint32
	NumGoroutine    int
	CPUCores        int
}

// Start begins collecting metrics at regular intervals
func (mc *MetricsCollector) Start() {
	mc.mu.Lock()
	if mc.running {
		mc.mu.Unlock()
		return
	}
	mc.running = true
	mc.mu.Unlock()

	go mc.collect()
}

// Stop stops collecting metrics
func (mc *MetricsCollector) Stop() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.running {
		return
	}

	close(mc.stopChan)
	mc.running = false
}

// collect runs the collection loop
func (mc *MetricsCollector) collect() {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.captureSnapshot()
		case <-mc.stopChan:
			return
		}
	}
}

// captureSnapshot captures current system metrics
func (mc *MetricsCollector) captureSnapshot() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	snapshot := SystemSnapshot{
		Timestamp:        time.Now(),
		MemoryAlloc:      m.Alloc,
		MemoryTotalAlloc: m.TotalAlloc,
		MemorySys:        m.Sys,
		NumGC:            m.NumGC,
		NumGoroutine:     runtime.NumGoroutine(),
		CPUCores:         runtime.NumCPU(),
	}

	mc.mu.Lock()
	mc.snapshots = append(mc.snapshots, snapshot)
	mc.mu.Unlock()
}

// GetSnapshots returns all collected snapshots
func (mc *MetricsCollector) GetSnapshots() []SystemSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make([]SystemSnapshot, len(mc.snapshots))
	copy(result, mc.snapshots)
	return result
}

// GetLatestSnapshot returns the most recent snapshot
func (mc *MetricsCollector) GetLatestSnapshot() *SystemSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(mc.snapshots) == 0 {
		return nil
	}

	return &mc.snapshots[len(mc.snapshots)-1]
}

// GetAverageMetrics calculates average metrics over all snapshots
func (mc *MetricsCollector) GetAverageMetrics() *AverageMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(mc.snapshots) == 0 {
		return nil
	}

	var totalMemAlloc uint64
	var totalMemSys uint64
	var totalGoroutines int

	for _, snapshot := range mc.snapshots {
		totalMemAlloc += snapshot.MemoryAlloc
		totalMemSys += snapshot.MemorySys
		totalGoroutines += snapshot.NumGoroutine
	}

	count := uint64(len(mc.snapshots))

	// Safe conversion: count is bounded by slice length which is much less than MaxInt
	sampleCount := len(mc.snapshots)
	if sampleCount > math.MaxInt32 {
		// Defensive: if we somehow have > MaxInt32 samples, cap it
		sampleCount = math.MaxInt32
	}

	return &AverageMetrics{
		AvgMemoryAlloc:   totalMemAlloc / count,
		AvgMemorySys:     totalMemSys / count,
		AvgGoroutines:    float64(totalGoroutines) / float64(count),
		SampleCount:      sampleCount,
	}
}

// AverageMetrics contains averaged system metrics
type AverageMetrics struct {
	AvgMemoryAlloc uint64
	AvgMemorySys   uint64
	AvgGoroutines  float64
	SampleCount    int
}

// Clear clears all snapshots
func (mc *MetricsCollector) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.snapshots = make([]SystemSnapshot, 0)
}

// IsRunning returns whether the collector is running
func (mc *MetricsCollector) IsRunning() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.running
}

// GetMemoryTrend analyzes memory usage trend
func (mc *MetricsCollector) GetMemoryTrend() MemoryTrend {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(mc.snapshots) < 2 {
		return MemoryTrend{
			Direction: "stable",
			Rate:      0,
		}
	}

	first := mc.snapshots[0]
	last := mc.snapshots[len(mc.snapshots)-1]

	// Safe conversion: check if values are within int64 range before subtraction
	// Memory values are typically well below MaxInt64 (2^63-1 ≈ 9.2 exabytes)
	var memoryDelta int64
	if last.MemoryAlloc > math.MaxInt64 || first.MemoryAlloc > math.MaxInt64 {
		// Extremely unlikely in practice, but handle defensively
		// Use float64 for calculation if values exceed int64 range
		memoryDelta = 0 // Fall back to stable trend
	} else {
		// Safe: both values confirmed to fit in int64
		memoryDelta = int64(last.MemoryAlloc) - int64(first.MemoryAlloc)
	}

	timeDelta := last.Timestamp.Sub(first.Timestamp).Seconds()

	rate := float64(memoryDelta) / timeDelta // bytes per second

	direction := "stable"
	if rate > 1000000 { // > 1MB/s
		direction = "increasing"
	} else if rate < -1000000 {
		direction = "decreasing"
	}

	return MemoryTrend{
		Direction:    direction,
		Rate:         rate,
		StartMemory:  first.MemoryAlloc,
		EndMemory:    last.MemoryAlloc,
		Duration:     last.Timestamp.Sub(first.Timestamp),
	}
}

// MemoryTrend represents memory usage trend
type MemoryTrend struct {
	Direction   string
	Rate        float64
	StartMemory uint64
	EndMemory   uint64
	Duration    time.Duration
}
