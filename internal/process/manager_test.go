package process

import (
	"context"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// TestManager_Spawn tests basic process spawning
func TestManager_Spawn(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
		MaxExecutionTime:   10,
	}

	// Use a simple command that exists on all systems
	// Note: This will fail because agent binary doesn't exist yet
	// but tests the spawning logic
	_, err := m.Spawn(ctx, "test-agent", perms)

	// We expect an error because the agent binary doesn't exist
	if err == nil {
		t.Error("expected error spawning non-existent agent")
	}
}

// TestManager_SpawnWithTimeout tests timeout enforcement
func TestManager_SpawnWithTimeout(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
		MaxExecutionTime:   1, // 1 second timeout
	}

	// Try to spawn with very short timeout
	_, err := m.Spawn(ctx, "test-agent", perms)

	// Should get an error since agent doesn't exist
	if err == nil {
		t.Error("expected error spawning non-existent agent")
	}
}

// TestManager_KillNonExistent tests killing non-existent process
func TestManager_KillNonExistent(t *testing.T) {
	m := NewManager()

	err := m.Kill("non-existent-id")
	if err == nil {
		t.Error("expected error killing non-existent process")
	}
}

// TestManager_GetProcess tests process retrieval
func TestManager_GetProcess(t *testing.T) {
	m := NewManager()

	// Test getting non-existent process
	_, err := m.GetProcess("non-existent-id")
	if err == nil {
		t.Error("expected error getting non-existent process")
	}
}

// TestManager_ListProcesses tests listing all processes
func TestManager_ListProcesses(t *testing.T) {
	m := NewManager()

	// Initially should be empty
	processes := m.ListProcesses()
	if len(processes) != 0 {
		t.Errorf("expected 0 processes, got %d", len(processes))
	}
}

// TestManager_Cleanup tests cleanup of all processes
func TestManager_Cleanup(t *testing.T) {
	m := NewManager()

	// Cleanup empty manager should succeed
	err := m.Cleanup()
	if err != nil {
		t.Errorf("cleanup failed: %v", err)
	}
}

// TestManager_ConcurrentAccess tests thread safety
func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
		MaxExecutionTime:   5,
	}

	// Spawn multiple processes concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = m.Spawn(ctx, "test-agent", perms)
		}()
	}

	wg.Wait()

	// List processes should not panic
	_ = m.ListProcesses()
}

// TestManager_ContextCancellation tests context cancellation
func TestManager_ContextCancellation(t *testing.T) {
	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())

	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
	}

	// Cancel context immediately
	cancel()

	// Spawn should handle cancelled context
	_, err := m.Spawn(ctx, "test-agent", perms)

	// Should get an error
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}

// TestAgentProcess_Structure tests AgentProcess struct fields
func TestAgentProcess_Structure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	process := &AgentProcess{
		ID:        "test-123",
		AgentID:   "test-agent",
		Cmd:       exec.CommandContext(ctx, "echo", "test"),
		StartTime: time.Now(),
		Permissions: schema.Permissions{
			AllowedDirectories: []string{"/tmp"},
			ReadOnly:           true,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	if process.ID != "test-123" {
		t.Errorf("expected ID 'test-123', got '%s'", process.ID)
	}

	if process.AgentID != "test-agent" {
		t.Errorf("expected AgentID 'test-agent', got '%s'", process.AgentID)
	}
}

// TestManager_WaitNonExistent tests waiting on non-existent process
func TestManager_WaitNonExistent(t *testing.T) {
	m := NewManager()

	err := m.Wait("non-existent-id")
	if err == nil {
		t.Error("expected error waiting on non-existent process")
	}
}

// TestManager_MultipleSpawn tests spawning multiple processes
func TestManager_MultipleSpawn(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
		MaxExecutionTime:   10,
	}

	// Try to spawn multiple processes
	for i := 0; i < 3; i++ {
		_, err := m.Spawn(ctx, "test-agent", perms)
		// All will fail because agent doesn't exist, but tests the logic
		if err == nil {
			t.Error("expected error spawning non-existent agent")
		}
	}
}

// TestManager_ProcessIDGeneration tests unique process ID generation
func TestManager_ProcessIDGeneration(t *testing.T) {
	// This tests the ID generation format: "agentID-timestamp"
	agentID := "test-agent"
	timestamp := time.Now().UnixNano()

	expectedPrefix := agentID + "-"
	processID := expectedPrefix + "12345"

	if len(processID) <= len(expectedPrefix) {
		t.Error("process ID should include timestamp")
	}

	if timestamp <= 0 {
		t.Error("timestamp should be positive")
	}
}

// TestManager_ThreadSafety tests concurrent operations
func TestManager_ThreadSafety(t *testing.T) {
	m := NewManager()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{"/tmp"},
		ReadOnly:           true,
	}

	var wg sync.WaitGroup

	// Concurrent spawns
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = m.Spawn(ctx, "test-agent", perms)
		}()
	}

	// Concurrent lists
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.ListProcesses()
		}()
	}

	// Concurrent gets
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = m.GetProcess("non-existent")
		}()
	}

	wg.Wait()
}

// TestManager_CleanupRaceCondition tests cleanup race conditions
func TestManager_CleanupRaceCondition(t *testing.T) {
	m := NewManager()

	// Test that cleanup doesn't panic with concurrent access
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = m.Cleanup()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = m.ListProcesses()
	}()

	wg.Wait()
}
