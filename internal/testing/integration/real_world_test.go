// Package integration_test provides real-world integration testing using actual Conexus source code
package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/analyzer"
	"github.com/ferg-cod3s/conexus/internal/agent/locator"
	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getProjectRoot returns the absolute path to the project root
func getProjectRoot(t *testing.T) string {
	// From internal/testing/integration, go up 3 levels
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	require.NoError(t, err, "Should get project root")
	return root
}

// TestRealCodebaseAnalysis tests analyzing actual Conexus analyzer.go file
func TestRealCodebaseAnalysis(t *testing.T) {
	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	// Get path to actual analyzer.go (absolute)
	analyzerPath := filepath.Join(projectRoot, "internal", "agent", "analyzer", "analyzer.go")
	
	// Verify file exists
	_, err := os.Stat(analyzerPath)
	require.NoError(t, err, "analyzer.go should exist")

	// Create tool executor
	executor := tool.NewExecutor()

	// Create real analyzer agent (v2 API)
	analyzerAgent := analyzer.New(executor)

	// Create permissions allowing file reads (use absolute path for allowed dir)
	perms := schema.Permissions{
		AllowedDirectories: []string{projectRoot}, // Allow entire project
		ReadOnly:           true,
	}

	// Analyze the file (v2 API)
	req := schema.AgentRequest{
		RequestID: "test-real-analysis-001",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:           []string{analyzerPath},
			SpecificRequest: "Analyze the analyzer.go implementation",
		},
		Permissions: perms,
		Timestamp:   time.Now(),
	}

	resp, err := analyzerAgent.Execute(ctx, req)
	require.NoError(t, err, "Analysis should complete")
	assert.Equal(t, schema.StatusComplete, resp.Status, "Analysis should succeed")

	// Verify output structure (v2: resp.Output is *AgentOutputV1, not []AgentOutputV1)
	require.NotNil(t, resp.Output, "Should have output")
	output := resp.Output
	
	assert.Equal(t, "AGENT_OUTPUT_V1", output.Version)
	assert.NotEmpty(t, output.ComponentName, "Should identify component")
	assert.NotEmpty(t, output.RawEvidence, "Should provide evidence")
	
	// Verify analysis found key structures
	assert.NotEmpty(t, output.EntryPoints, "Should identify entry points")
	
	t.Logf("✓ Analyzed real Conexus source file: %s", analyzerPath)
	t.Logf("✓ Found %d entry points", len(output.EntryPoints))
	t.Logf("✓ Found %d evidence items", len(output.RawEvidence))
}

// TestLocatorAnalyzerIntegration tests real locator -> analyzer workflow
func TestLocatorAnalyzerIntegration(t *testing.T) {
	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	// Create tool executor
	executor := tool.NewExecutor()

	// Create permissions (allow entire project)
	perms := schema.Permissions{
		AllowedDirectories: []string{projectRoot},
		ReadOnly:           true,
	}

	// Phase 1: Use locator to find Go files
	locatorAgent := locator.New(executor)

	analyzerDir := filepath.Join(projectRoot, "internal", "agent", "analyzer")
	locReq := schema.AgentRequest{
		RequestID: "test-locator-001",
		AgentID:   "codebase-locator",
		Task: schema.AgentTask{
			Files:              []string{},
			AllowedDirectories: []string{analyzerDir},
			SpecificRequest:    `{"pattern": "*.go", "exclude": ["*_test.go"], "root_path": "` + analyzerDir + `"}`,
		},
		Permissions: perms,
		Timestamp:   time.Now(),
	}

	locResp, err := locatorAgent.Execute(ctx, locReq)
	require.NoError(t, err, "Locator should complete")
	require.Equal(t, schema.StatusComplete, locResp.Status)
	require.NotNil(t, locResp.Output, "Locator should find files")

	locOutput := locResp.Output
	assert.NotEmpty(t, locOutput.RawEvidence, "Should have file evidence")

	// Phase 2: Analyze located files
	analyzerAgent := analyzer.New(executor)

	// Find analyzer.go in evidence
	targetFile := ""
	for _, ev := range locOutput.RawEvidence {
		if strings.HasSuffix(ev.File, "analyzer.go") {
			targetFile = ev.File
			break
		}
	}
	require.NotEmpty(t, targetFile, "Should find analyzer.go")

	anaReq := schema.AgentRequest{
		RequestID: "test-analyzer-001",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:           []string{targetFile},
			SpecificRequest: "Analyze located file",
		},
		Permissions: perms,
		Timestamp:   time.Now(),
	}

	anaResp, err := analyzerAgent.Execute(ctx, anaReq)
	require.NoError(t, err, "Analyzer should complete")
	require.Equal(t, schema.StatusComplete, anaResp.Status)

	t.Logf("✓ Locator found files in %s", analyzerDir)
	t.Logf("✓ Analyzer processed: %s", targetFile)
	t.Logf("✓ Two-phase workflow completed successfully")
}

// TestComplexWorkflowWithRealCode tests a manual multi-step workflow on real code
// This test demonstrates locator->analyzer coordination without framework
func TestComplexWorkflowWithRealCode(t *testing.T) {
	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	// Create tool executor
	executor := tool.NewExecutor()

	// Create agents
	locatorAgent := locator.New(executor)
	analyzerAgent := analyzer.New(executor)

	// Create permissions (allow entire project)
	perms := schema.Permissions{
		AllowedDirectories: []string{projectRoot},
		ReadOnly:           true,
	}

	// Step 1: Locate all Go files in agent directory
	agentDir := filepath.Join(projectRoot, "internal", "agent")
	locReq := schema.AgentRequest{
		RequestID: "workflow-step-1",
		AgentID:   "codebase-locator",
		Task: schema.AgentTask{
			AllowedDirectories: []string{agentDir},
			SpecificRequest:    `{"pattern": "*.go", "exclude": ["*_test.go"], "root_path": "` + agentDir + `"}`,
		},
		Permissions: perms,
		Timestamp:   time.Now(),
	}

	locResp, err := locatorAgent.Execute(ctx, locReq)
	require.NoError(t, err, "Step 1: Locator should complete")
	require.Equal(t, schema.StatusComplete, locResp.Status)
	require.NotNil(t, locResp.Output)
	
	locatedFiles := locResp.Output.RawEvidence
	require.NotEmpty(t, locatedFiles, "Step 1: Should find files")

	// Step 2: Analyze first 2 located files
	analyzedCount := 0
	maxToAnalyze := 2

	for _, ev := range locatedFiles {
		if analyzedCount >= maxToAnalyze {
			break
		}
		
		if !strings.HasSuffix(ev.File, ".go") {
			continue
		}

		anaReq := schema.AgentRequest{
			RequestID: "workflow-step-2-" + string(rune('A'+analyzedCount)),
			AgentID:   "codebase-analyzer",
			Task: schema.AgentTask{
				Files:           []string{ev.File},
				SpecificRequest: "Analyze source file",
			},
			Permissions: perms,
			Timestamp:   time.Now(),
		}

		anaResp, err := analyzerAgent.Execute(ctx, anaReq)
		require.NoError(t, err, "Step 2: Analyzer should complete for %s", ev.File)
		require.Equal(t, schema.StatusComplete, anaResp.Status)
		require.NotNil(t, anaResp.Output)
		
		analyzedCount++
		t.Logf("  → Analyzed: %s (%d entry points)", 
			anaResp.Output.ComponentName, len(anaResp.Output.EntryPoints))
	}

	assert.Equal(t, maxToAnalyze, analyzedCount, "Should analyze expected number of files")

	t.Logf("✓ Complex workflow completed: Located %d files, analyzed %d", 
		len(locatedFiles), analyzedCount)
}

// TestMultiFileAnalysis tests analyzing multiple real source files
func TestMultiFileAnalysis(t *testing.T) {
	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	// Absolute paths to files
	files := []string{
		filepath.Join(projectRoot, "internal", "agent", "analyzer", "analyzer.go"),
		filepath.Join(projectRoot, "internal", "agent", "locator", "locator.go"),
	}

	// Create tool executor
	executor := tool.NewExecutor()
	analyzerAgent := analyzer.New(executor)

	// Create permissions (allow entire project)
	perms := schema.Permissions{
		AllowedDirectories: []string{projectRoot},
		ReadOnly:           true,
	}

	results := make([]*schema.AgentOutputV1, 0, len(files))

	for i, file := range files {
		// Verify file exists
		_, err := os.Stat(file)
		if os.IsNotExist(err) {
			t.Logf("Skipping missing file: %s", file)
			continue
		}

		req := schema.AgentRequest{
			RequestID: "test-multi-file-" + string(rune('A'+i)),
			AgentID:   "codebase-analyzer",
			Task: schema.AgentTask{
				Files:           []string{file},
				SpecificRequest: "Analyze source file",
			},
			Permissions: perms,
			Timestamp:   time.Now(),
		}

		resp, err := analyzerAgent.Execute(ctx, req)
		require.NoError(t, err, "Analysis of %s should complete", file)
		require.Equal(t, schema.StatusComplete, resp.Status)
		
		if resp.Output != nil {
			results = append(results, resp.Output)
		}
	}

	require.NotEmpty(t, results, "Should have analyzed at least one file")

	// Verify all outputs are valid
	for i, output := range results {
		assert.Equal(t, "AGENT_OUTPUT_V1", output.Version)
		assert.NotEmpty(t, output.RawEvidence, "File %d should have evidence", i)
		t.Logf("✓ File %d: %s (%d evidence items)", 
			i+1, output.ComponentName, len(output.RawEvidence))
	}

	t.Logf("✓ Analyzed %d real source files", len(results))
}

// TestErrorHandlingWithRealInvalidInput tests error handling with invalid real-world input
func TestErrorHandlingWithRealInvalidInput(t *testing.T) {
	ctx := context.Background()
	projectRoot := getProjectRoot(t)

	// Create tool executor
	executor := tool.NewExecutor()
	analyzerAgent := analyzer.New(executor)

	// Create permissions (allow entire project)
	perms := schema.Permissions{
		AllowedDirectories: []string{projectRoot},
		ReadOnly:           true,
	}

	// Test 1: Non-existent file
	req1 := schema.AgentRequest{
		RequestID: "test-error-001",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:           []string{filepath.Join(projectRoot, "does_not_exist.go")},
			SpecificRequest: "Analyze non-existent file",
		},
		Permissions: perms,
		Timestamp:   time.Now(),
	}

	resp1, err1 := analyzerAgent.Execute(ctx, req1)
	// Should handle gracefully (either error or error status)
	if err1 != nil {
		t.Logf("✓ Non-existent file error handled: %v", err1)
	} else {
		assert.Equal(t, schema.StatusError, resp1.Status, "Should return error status")
		t.Logf("✓ Non-existent file handled with error response")
	}

	// Test 2: Empty file list
	req2 := schema.AgentRequest{
		RequestID: "test-error-002",
		AgentID:   "codebase-analyzer",
		Task: schema.AgentTask{
			Files:           []string{}, // Empty!
			SpecificRequest: "Analyze with no files",
		},
		Permissions: perms,
		Timestamp:   time.Now(),
	}

	resp2, err2 := analyzerAgent.Execute(ctx, req2)
	// Should handle gracefully
	if err2 != nil {
		t.Logf("✓ Empty files error handled: %v", err2)
		assert.Contains(t, err2.Error(), "no files", "Error should mention no files")
	} else {
		assert.Equal(t, schema.StatusError, resp2.Status, "Should return error status")
		t.Logf("✓ Empty files handled with error response")
	}

	t.Logf("✓ Error handling tested with invalid inputs")
}
