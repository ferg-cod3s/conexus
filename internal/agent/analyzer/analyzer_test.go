package analyzer

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestAnalyzerAgent_BasicAnalysis(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	fixturePath, _ := filepath.Abs("../../../../conexus/tests/fixtures")

	req := schema.AgentRequest{
		RequestID: "test-001",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:              []string{filepath.Join(fixturePath, "simple_function.go")},
			AllowedDirectories: []string{fixturePath},
			SpecificRequest:    "analyze this file",
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{fixturePath},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	response, err := agent.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if response.Status != schema.StatusComplete {
		t.Errorf("Expected status Complete, got %v", response.Status)
	}

	if response.Output == nil {
		t.Fatal("Expected output, got nil")
	}

	output := response.Output

	// Validate AGENT_OUTPUT_V1 format
	if output.Version != "AGENT_OUTPUT_V1" {
		t.Errorf("Expected version AGENT_OUTPUT_V1, got %s", output.Version)
	}

	// Should find entry points
	if len(output.EntryPoints) == 0 {
		t.Error("Expected entry points, got none")
	}

	// Should have evidence
	if len(output.RawEvidence) == 0 {
		t.Error("Expected evidence, got none")
	}

	t.Logf("Found %d entry points", len(output.EntryPoints))
	t.Logf("Evidence count: %d", len(output.RawEvidence))
}

func TestAnalyzerAgent_CallGraph(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	fixturePath, _ := filepath.Abs("../../../../conexus/tests/fixtures")

	req := schema.AgentRequest{
		RequestID: "test-002",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:              []string{filepath.Join(fixturePath, "multiple_functions.go")},
			AllowedDirectories: []string{fixturePath},
			SpecificRequest:    "analyze call graph",
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{fixturePath},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	response, err := agent.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := response.Output

	// Should detect function calls
	if len(output.CallGraph) == 0 {
		t.Error("Expected call graph edges, got none")
	}

	t.Logf("Call graph edges: %d", len(output.CallGraph))

	for _, edge := range output.CallGraph {
		t.Logf("Call: %s -> %s (line %d)", edge.From, edge.To, edge.ViaLine)
	}
}

func TestAnalyzerAgent_StructMethods(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	fixturePath, _ := filepath.Abs("../../../../conexus/tests/fixtures")

	req := schema.AgentRequest{
		RequestID: "test-003",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:              []string{filepath.Join(fixturePath, "struct_methods.go")},
			AllowedDirectories: []string{fixturePath},
			SpecificRequest:    "analyze struct and methods",
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{fixturePath},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	response, err := agent.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := response.Output

	// Should find methods
	methodCount := 0
	for _, ep := range output.EntryPoints {
		if ep.Role == "method" {
			methodCount++
		}
	}

	if methodCount == 0 {
		t.Error("Expected to find methods, found none")
	}

	// Should detect state operations
	if len(output.StateManagement) == 0 {
		t.Log("Note: No state operations detected (may be expected)")
	}

	t.Logf("Found %d methods", methodCount)
	t.Logf("State operations: %d", len(output.StateManagement))
}

func TestAnalyzerAgent_ErrorHandling(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	fixturePath, _ := filepath.Abs("../../../../conexus/tests/fixtures")

	req := schema.AgentRequest{
		RequestID: "test-004",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:              []string{filepath.Join(fixturePath, "error_handling.go")},
			AllowedDirectories: []string{fixturePath},
			SpecificRequest:    "analyze error handling",
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{fixturePath},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	response, err := agent.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := response.Output

	// Should detect error handling patterns
	if len(output.ErrorHandling) == 0 {
		t.Error("Expected error handling patterns, got none")
	}

	t.Logf("Error handling patterns: %d", len(output.ErrorHandling))

	for _, eh := range output.ErrorHandling {
		t.Logf("Error handler: %s (type: %s)", eh.Condition, eh.Type)
	}
}

func TestAnalyzerAgent_SideEffects(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	fixturePath, _ := filepath.Abs("../../../../conexus/tests/fixtures")

	req := schema.AgentRequest{
		RequestID: "test-005",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:              []string{filepath.Join(fixturePath, "side_effects.go")},
			AllowedDirectories: []string{fixturePath},
			SpecificRequest:    "analyze side effects",
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{fixturePath},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	response, err := agent.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := response.Output

	// Should detect side effects (logging, I/O)
	if len(output.SideEffects) == 0 {
		t.Error("Expected side effects, got none")
	}

	t.Logf("Side effects detected: %d", len(output.SideEffects))

	// Count by type
	typeCounts := make(map[string]int)
	for _, se := range output.SideEffects {
		typeCounts[se.Type]++
	}

	for effectType, count := range typeCounts {
		t.Logf("  %s: %d", effectType, count)
	}
}

func TestAnalyzerAgent_EvidenceBacking(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	fixturePath, _ := filepath.Abs("../../../../conexus/tests/fixtures")

	req := schema.AgentRequest{
		RequestID: "test-006",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:              []string{filepath.Join(fixturePath, "simple_function.go")},
			AllowedDirectories: []string{fixturePath},
			SpecificRequest:    "analyze with evidence",
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{fixturePath},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	response, err := agent.Execute(ctx, req)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	output := response.Output

	// Every entry point should have corresponding evidence
	if len(output.EntryPoints) > 0 {
		evidenceForEntryPoints := 0
		for _, evidence := range output.RawEvidence {
			if contains(evidence.Claim, "Entry point") {
				evidenceForEntryPoints++
			}
		}

		if evidenceForEntryPoints != len(output.EntryPoints) {
			t.Errorf("Evidence mismatch: %d entry points, %d evidence entries",
				len(output.EntryPoints), evidenceForEntryPoints)
		}
	}

	t.Logf("Evidence validation passed: %d evidence entries", len(output.RawEvidence))
}

func TestFindEntryPoints(t *testing.T) {
	content := `package test

func Add(a, b int) int {
	return a + b
}

func privateFunc() {
	// not exported
}

func (c *Calculator) Multiply(x, y int) int {
	return x * y
}
`
	lines := strings.Split(content, "\n")
	entryPoints := findEntryPoints("/test.go", lines)

	if len(entryPoints) != 2 {
		t.Errorf("Expected 2 entry points, got %d", len(entryPoints))
	}

	// Should find Add and Multiply (both exported)
	foundAdd := false
	foundMultiply := false
	for _, ep := range entryPoints {
		if ep.Symbol == "Add" {
			foundAdd = true
		}
		if ep.Symbol == "Multiply" {
			foundMultiply = true
		}
	}

	if !foundAdd {
		t.Error("Did not find Add function")
	}
	if !foundMultiply {
		t.Error("Did not find Multiply method")
	}
}

func TestIsExported(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"Add", true},
		{"Calculate", true},
		{"privateFunc", false},
		{"_internal", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExported(tt.name)
			if result != tt.expected {
				t.Errorf("isExported(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
