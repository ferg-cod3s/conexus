package profiling

import (
	"testing"
	"time"
)

func TestNewMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector(100 * time.Millisecond)
	if mc == nil {
		t.Fatal("expected collector, got nil")
	}

	if mc.interval != 100*time.Millisecond {
		t.Errorf("expected interval 100ms, got %v", mc.interval)
	}

	if mc.snapshots == nil {
		t.Error("expected snapshots to be initialized")
	}
}

func TestMetricsCollector_StartStop(t *testing.T) {
	mc := NewMetricsCollector(50 * time.Millisecond)

	if mc.IsRunning() {
		t.Error("expected collector to not be running initially")
	}

	mc.Start()

	if !mc.IsRunning() {
		t.Error("expected collector to be running after start")
	}

	// Let it collect a few snapshots
	time.Sleep(200 * time.Millisecond)

	mc.Stop()

	if mc.IsRunning() {
		t.Error("expected collector to not be running after stop")
	}

	// Verify snapshots were collected
	snapshots := mc.GetSnapshots()
	if len(snapshots) == 0 {
		t.Error("expected at least one snapshot")
	}
}

func TestMetricsCollector_CaptureSnapshot(t *testing.T) {
	mc := NewMetricsCollector(100 * time.Millisecond)

	mc.captureSnapshot()

	snapshots := mc.GetSnapshots()
	if len(snapshots) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(snapshots))
	}

	snapshot := snapshots[0]

	if snapshot.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}

	if snapshot.MemoryAlloc == 0 {
		t.Error("expected non-zero memory alloc")
	}

	if snapshot.NumGoroutine == 0 {
		t.Error("expected non-zero goroutine count")
	}

	if snapshot.CPUCores == 0 {
		t.Error("expected non-zero CPU cores")
	}
}

func TestMetricsCollector_GetLatestSnapshot(t *testing.T) {
	mc := NewMetricsCollector(100 * time.Millisecond)

	// Initially no snapshots
	latest := mc.GetLatestSnapshot()
	if latest != nil {
		t.Error("expected nil for no snapshots")
	}

	// Capture snapshots
	mc.captureSnapshot()
	time.Sleep(10 * time.Millisecond)
	mc.captureSnapshot()

	latest = mc.GetLatestSnapshot()
	if latest == nil {
		t.Fatal("expected latest snapshot")
	}

	// Verify it's the most recent
	snapshots := mc.GetSnapshots()
	if latest.Timestamp != snapshots[len(snapshots)-1].Timestamp {
		t.Error("expected latest to match last snapshot")
	}
}

func TestMetricsCollector_GetAverageMetrics(t *testing.T) {
	mc := NewMetricsCollector(100 * time.Millisecond)

	// No snapshots
	avg := mc.GetAverageMetrics()
	if avg != nil {
		t.Error("expected nil for no snapshots")
	}

	// Capture multiple snapshots
	for i := 0; i < 5; i++ {
		mc.captureSnapshot()
		time.Sleep(10 * time.Millisecond)
	}

	avg = mc.GetAverageMetrics()
	if avg == nil {
		t.Fatal("expected average metrics")
	}

	if avg.SampleCount != 5 {
		t.Errorf("expected 5 samples, got %d", avg.SampleCount)
	}

	if avg.AvgMemoryAlloc == 0 {
		t.Error("expected non-zero avg memory")
	}

	if avg.AvgGoroutines == 0 {
		t.Error("expected non-zero avg goroutines")
	}
}

func TestMetricsCollector_GetMemoryTrend(t *testing.T) {
	mc := NewMetricsCollector(100 * time.Millisecond)

	// Less than 2 snapshots
	trend := mc.GetMemoryTrend()
	if trend.Direction != "stable" {
		t.Error("expected stable for insufficient data")
	}

	// Capture snapshots
	for i := 0; i < 10; i++ {
		mc.captureSnapshot()
		time.Sleep(20 * time.Millisecond)
	}

	trend = mc.GetMemoryTrend()

	if trend.Direction == "" {
		t.Error("expected direction to be set")
	}

	if trend.Duration == 0 {
		t.Error("expected non-zero duration")
	}

	validDirections := map[string]bool{
		"stable":     true,
		"increasing": true,
		"decreasing": true,
	}

	if !validDirections[trend.Direction] {
		t.Errorf("unexpected direction: %s", trend.Direction)
	}
}

func TestMetricsCollector_Clear(t *testing.T) {
	mc := NewMetricsCollector(100 * time.Millisecond)

	// Capture snapshots
	for i := 0; i < 5; i++ {
		mc.captureSnapshot()
	}

	snapshots := mc.GetSnapshots()
	if len(snapshots) == 0 {
		t.Fatal("expected snapshots before clear")
	}

	// Clear
	mc.Clear()

	snapshots = mc.GetSnapshots()
	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots after clear, got %d", len(snapshots))
	}
}

func TestMetricsCollector_MultipleStartStop(t *testing.T) {
	mc := NewMetricsCollector(50 * time.Millisecond)

	// Start multiple times
	mc.Start()
	mc.Start() // Should not create duplicate collection loop

	if !mc.IsRunning() {
		t.Error("expected collector to be running")
	}

	mc.Stop()

	if mc.IsRunning() {
		t.Error("expected collector to be stopped")
	}

	// Stop again should be safe
	mc.Stop()

	if mc.IsRunning() {
		t.Error("expected collector to still be stopped")
	}
}

func TestMemoryTrend_IncreasingDetection(t *testing.T) {
	mc := NewMetricsCollector(10 * time.Millisecond)

	// Capture baseline
	mc.captureSnapshot()

	// Allocate memory to create increasing trend
	data := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		data[i] = make([]byte, 100000) // 100KB each
		if i%10 == 0 {
			mc.captureSnapshot()
			time.Sleep(10 * time.Millisecond)
		}
	}

	trend := mc.GetMemoryTrend()

	// Should detect some change (may be increasing, stable depends on GC)
	if trend.EndMemory < trend.StartMemory {
		t.Error("expected end memory >= start memory after allocations")
	}

	// Keep data alive to prevent GC
	_ = data
}

func TestSystemSnapshot_Fields(t *testing.T) {
	mc := NewMetricsCollector(100 * time.Millisecond)
	mc.captureSnapshot()

	snapshot := mc.GetLatestSnapshot()
	if snapshot == nil {
		t.Fatal("expected snapshot")
	}

	// Verify all fields are populated
	if snapshot.Timestamp.IsZero() {
		t.Error("timestamp should be set")
	}

	if snapshot.MemoryAlloc == 0 {
		t.Error("memory alloc should be > 0")
	}

	if snapshot.MemoryTotalAlloc == 0 {
		t.Error("memory total alloc should be > 0")
	}

	if snapshot.MemorySys == 0 {
		t.Error("memory sys should be > 0")
	}

	if snapshot.NumGoroutine == 0 {
		t.Error("num goroutine should be > 0")
	}

	if snapshot.CPUCores == 0 {
		t.Error("CPU cores should be > 0")
	}
}
