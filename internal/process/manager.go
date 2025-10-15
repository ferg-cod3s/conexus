package process

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Manager handles agent process lifecycle
type Manager struct {
	processes map[string]*AgentProcess
	mu        sync.RWMutex
}

// AgentProcess represents a running agent process
type AgentProcess struct {
	ID          string
	AgentID     string
	Cmd         *exec.Cmd
	Stdin       io.WriteCloser
	Stdout      io.ReadCloser
	Stderr      io.ReadCloser
	StartTime   time.Time
	Permissions schema.Permissions
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewManager creates a new process manager
func NewManager() *Manager {
	return &Manager{
		processes: make(map[string]*AgentProcess),
	}
}

// Spawn creates and starts a new agent process
func (m *Manager) Spawn(ctx context.Context, agentID string, perms schema.Permissions) (*AgentProcess, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create context with timeout if specified
	processCtx := ctx
	var cancel context.CancelFunc
	if perms.MaxExecutionTime > 0 {
		processCtx, cancel = context.WithTimeout(ctx, time.Duration(perms.MaxExecutionTime)*time.Second)
	} else {
		processCtx, cancel = context.WithCancel(ctx)
	}

	// TODO: Determine agent binary path based on agentID
	// For now, we'll use a placeholder
	agentBinary := fmt.Sprintf("./agents/%s", agentID)

	cmd := exec.CommandContext(processCtx, agentBinary)

	// Set up pipes for communication
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start agent process: %w", err)
	}

	processID := fmt.Sprintf("%s-%d", agentID, time.Now().UnixNano())
	process := &AgentProcess{
		ID:          processID,
		AgentID:     agentID,
		Cmd:         cmd,
		Stdin:       stdin,
		Stdout:      stdout,
		Stderr:      stderr,
		StartTime:   time.Now(),
		Permissions: perms,
		ctx:         processCtx,
		cancel:      cancel,
	}

	m.processes[processID] = process

	return process, nil
}

// Kill terminates an agent process
func (m *Manager) Kill(processID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	process, exists := m.processes[processID]
	if !exists {
		return fmt.Errorf("process not found: %s", processID)
	}

	// Cancel the context to trigger cleanup
	process.cancel()

	// Kill the process if it's still running
	if process.Cmd.Process != nil {
		if err := process.Cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	// Clean up
	delete(m.processes, processID)

	return nil
}

// Wait blocks until a process completes
func (m *Manager) Wait(processID string) error {
	m.mu.RLock()
	process, exists := m.processes[processID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("process not found: %s", processID)
	}

	return process.Cmd.Wait()
}

// GetProcess retrieves a process by ID
func (m *Manager) GetProcess(processID string) (*AgentProcess, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	process, exists := m.processes[processID]
	if !exists {
		return nil, fmt.Errorf("process not found: %s", processID)
	}

	return process, nil
}

// ListProcesses returns all running processes
func (m *Manager) ListProcesses() []*AgentProcess {
	m.mu.RLock()
	defer m.mu.RUnlock()

	processes := make([]*AgentProcess, 0, len(m.processes))
	for _, p := range m.processes {
		processes = append(processes, p)
	}

	return processes
}

// Cleanup terminates all running processes
func (m *Manager) Cleanup() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error
	for id := range m.processes {
		if err := m.Kill(id); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}
