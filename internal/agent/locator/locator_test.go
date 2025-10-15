package locator

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestLocatorAgent_FileDiscovery(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	req := schema.AgentRequest{
		RequestID: "test-001",
		AgentID:   "codebase-locator",
		Task: schema.AgentTask{
			SpecificRequest:    "find all *.go files",
			AllowedDirectories: []string{"/home/f3rg/src/github/conexus/tests/fixtures"},
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{"/home/f3rg/src/github/conexus/tests/fixtures"},
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
	if output.Version != "AGENT_OUTPUT_V1" {
		t.Errorf("Expected version AGENT_OUTPUT_V1, got %s", output.Version)
	}

	if len(output.EntryPoints) == 0 {
		t.Error("Expected entry points, got none")
	}

	if len(output.RawEvidence) == 0 {
		t.Error("Expected evidence, got none")
	}

	t.Logf("Found %d files", len(output.EntryPoints))
	t.Logf("Evidence count: %d", len(output.RawEvidence))
}

func TestLocatorAgent_SymbolSearch(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	req := schema.AgentRequest{
		RequestID: "test-002",
		AgentID:   "codebase-locator",
		Task: schema.AgentTask{
			SpecificRequest:    "find function Add",
			Files:              []string{"/home/f3rg/src/github/conexus/tests/fixtures/simple_function.go"},
			AllowedDirectories: []string{"/home/f3rg/src/github/conexus/tests/fixtures"},
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{"/home/f3rg/src/github/conexus/tests/fixtures"},
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
	if output.Version != "AGENT_OUTPUT_V1" {
		t.Errorf("Expected version AGENT_OUTPUT_V1, got %s", output.Version)
	}

	t.Logf("Symbol search completed with %d results", len(output.EntryPoints))
}

func TestLocatorAgent_OutputValidation(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	req := schema.AgentRequest{
		RequestID: "test-003",
		AgentID:   "codebase-locator",
		Task: schema.AgentTask{
			SpecificRequest:    "find all files",
			AllowedDirectories: []string{"/home/f3rg/src/github/conexus/tests/fixtures"},
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{"/home/f3rg/src/github/conexus/tests/fixtures"},
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

	// Validate required fields
	if output.Version == "" {
		t.Error("Version is required")
	}
	if output.ComponentName == "" {
		t.Error("ComponentName is required")
	}
	if output.ScopeDescription == "" {
		t.Error("ScopeDescription is required")
	}
	if output.Overview == "" {
		t.Error("Overview is required")
	}

	// Validate evidence backing
	if len(output.EntryPoints) > 0 && len(output.RawEvidence) == 0 {
		t.Error("Entry points exist but no evidence provided")
	}

	t.Log("Output validation passed")
}

func TestLocatorAgent_PermissionBoundaries(t *testing.T) {
	executor := tool.NewExecutor()
	agent := New(executor)

	// Try to access directory outside permissions
	req := schema.AgentRequest{
		RequestID: "test-004",
		AgentID:   "codebase-locator",
		Task: schema.AgentTask{
			SpecificRequest:    "find all files",
			AllowedDirectories: []string{"/tmp/restricted"},
		},
		Permissions: schema.Permissions{
			AllowedDirectories: []string{"/home/f3rg/src/github/conexus/tests/fixtures"},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
		Timestamp: time.Now(),
	}

	ctx := context.Background()
	response, err := agent.Execute(ctx, req)

	// Should complete but with no results (or error)
	if err != nil {
		t.Logf("Expected permission error or empty results: %v", err)
	}

	if response.Output != nil && len(response.Output.EntryPoints) > 0 {
		t.Error("Should not have found files outside allowed directories")
	}

	t.Log("Permission boundary test passed")
}

func TestDetermineSearchType(t *testing.T) {
	tests := []struct {
		request  string
		expected string
	}{
		{"find all .go files", "file"},
		{"locate function Add", "symbol"},
		{"find type Calculator", "symbol"},
		{"search for files matching *.ts", "file"},
		{"where is the main function", "symbol"},
	}

	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			result := determineSearchType(tt.request)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractPattern(t *testing.T) {
	tests := []struct {
		request  string
		expected string
	}{
		{"find all *.go files", "*.go"},
		{"locate *.ts files", "*.ts"},
		{"find .go files", "*.go"},
		{"search for files", ""},
	}

	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			result := extractPattern(tt.request)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestExtractSymbolName(t *testing.T) {
	tests := []struct {
		request  string
		expected string
	}{
		{"find function Add", "Add"},
		{"locate type Calculator", "Calculator"},
		{"where is struct User", "User"},
		{"search", "search"},
	}

	for _, tt := range tests {
		t.Run(tt.request, func(t *testing.T) {
			result := extractSymbolName(tt.request)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
