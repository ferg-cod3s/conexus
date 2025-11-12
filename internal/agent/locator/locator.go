package locator

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// LocatorAgent finds files and symbols in a codebase
type LocatorAgent struct {
	executor *tool.Executor
}

// New creates a new LocatorAgent
func New(executor *tool.Executor) *LocatorAgent {
	return &LocatorAgent{
		executor: executor,
	}
}

// Execute processes an agent request to locate files or symbols
func (a *LocatorAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	response := schema.AgentResponse{
		RequestID: req.RequestID,
		AgentID:   "codebase-locator",
		Status:    schema.StatusComplete,
		Timestamp: req.Timestamp,
	}

	// Determine search type from request
	searchType := determineSearchType(req.Task.SpecificRequest)

	var output *schema.AgentOutputV1
	var err error

	switch searchType {
	case "file":
		output, err = a.findFiles(ctx, req)
	case "symbol":
		output, err = a.findSymbols(ctx, req)
	default:
		// Default to file search
		output, err = a.findFiles(ctx, req)
	}

	if err != nil {
		response.Status = schema.StatusError
		response.Error = &schema.AgentError{
			Code:        "LOCATOR_ERROR",
			Message:     err.Error(),
			Recoverable: true,
		}
		return response, err
	}

	response.Output = output
	return response, nil
}

// findFiles locates files matching patterns
func (a *LocatorAgent) findFiles(ctx context.Context, req schema.AgentRequest) (*schema.AgentOutputV1, error) {
	// Determine which directories to search: Task.AllowedDirectories takes precedence
	searchDirs := req.Task.AllowedDirectories
	if len(searchDirs) == 0 {
		// Fall back to permissions if task does not specify
		searchDirs = req.Permissions.AllowedDirectories
	}

	output := &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "File Discovery",
		ScopeDescription: fmt.Sprintf("File search in directories: %v", searchDirs),
		Overview:         "Locates files matching specified patterns using glob and list tools",
	}

	var allFiles []string
	var evidence []schema.Evidence

	// Search in each allowed directory
	for _, dir := range searchDirs {
		// Use glob tool to find matching files
		pattern := extractPattern(req.Task.SpecificRequest)
		if pattern == "" {
			pattern = "**/*.go" // Default pattern
		}

		params := tool.ToolParams{
			Path:    dir,
			Pattern: pattern,
		}

		result, err := a.executor.Execute(ctx, "glob", params, req.Permissions)
		if err != nil {
			continue // Skip directories with errors
		}

		if files, ok := result.Output.([]string); ok {
			allFiles = append(allFiles, files...)

			// Add evidence for each file found
			for _, file := range files {
				evidence = append(evidence, schema.Evidence{
					Claim: fmt.Sprintf("File found: %s", filepath.Base(file)),
					File:  file,
					Lines: "1-1", // File existence evidence
				})
			}
		}
	}

	// Populate entry points for found files
	for _, file := range allFiles {
		output.EntryPoints = append(output.EntryPoints, schema.EntryPoint{
			File:   file,
			Lines:  "1-1",
			Symbol: filepath.Base(file),
			Role:   "file",
		})
	}

	output.RawEvidence = evidence
	output.Limitations = []string{
		"Pattern matching uses glob syntax only",
		"No content-based filtering applied",
	}

	return output, nil
}

// findSymbols locates symbols (functions, types, etc.) in code
func (a *LocatorAgent) findSymbols(ctx context.Context, req schema.AgentRequest) (*schema.AgentOutputV1, error) {
	output := &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "Symbol Discovery",
		ScopeDescription: fmt.Sprintf("Symbol search in files: %v", req.Task.Files),
		Overview:         "Locates function and type declarations using grep-based text search",
	}

	var evidence []schema.Evidence
	symbolPattern := extractSymbolName(req.Task.SpecificRequest)

	// Search for symbols in specified files
	for _, file := range req.Task.Files {
		// Use grep to find function declarations
		params := tool.ToolParams{
			Path:    file,
			Pattern: fmt.Sprintf(`func.*%s`, symbolPattern),
		}

		result, err := a.executor.Execute(ctx, "grep", params, req.Permissions)
		if err != nil {
			continue
		}

		// Parse grep results (placeholder - grep tool needs implementation)
		if matches, ok := result.Output.([]string); ok {
			for range matches {
				output.EntryPoints = append(output.EntryPoints, schema.EntryPoint{
					File:   file,
					Lines:  "0-0", // Line number would come from grep output
					Symbol: symbolPattern,
					Role:   "function",
				})

				evidence = append(evidence, schema.Evidence{
					Claim: fmt.Sprintf("Symbol found: %s", symbolPattern),
					File:  file,
					Lines: "0-0",
				})
			}
		}
	}

	output.RawEvidence = evidence
	output.Limitations = []string{
		"Grep-based search only - not AST-aware",
		"May miss symbols with complex declarations",
		"Line numbers not yet available from grep tool",
	}
	output.OpenQuestions = []string{
		"Should we implement AST-based symbol search?",
		"What symbol types should be prioritized (functions, types, interfaces)?",
	}

	return output, nil
}

// Helper functions

func determineSearchType(request string) string {
	lower := strings.ToLower(request)
	if strings.Contains(lower, "function") || strings.Contains(lower, "symbol") || strings.Contains(lower, "type") {
		return "symbol"
	}
	return "file"
}

func extractPattern(request string) string {
	// First, try to parse as JSON
	if strings.Contains(request, "{") {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(request), &parsed); err == nil {
			if pattern, ok := parsed["pattern"].(string); ok && pattern != "" {
				return pattern
			}
		}
	}

	// Fall back to plain text pattern extraction
	// Simple pattern extraction from request
	// Example: "find all .go files" -> "*.go"
	lower := strings.ToLower(request)

	if strings.Contains(lower, "*.") {
		// Extract pattern like *.go
		parts := strings.Fields(request)
		for _, part := range parts {
			if strings.HasPrefix(part, "*.") {
				return part
			}
		}
	}

	if strings.Contains(lower, ".go") {
		return "*.go"
	}
	if strings.Contains(lower, ".ts") {
		return "*.ts"
	}

	return ""
}

func extractSymbolName(request string) string {
	// Extract symbol name from request
	// Example: "find function Add" -> "Add"
	words := strings.Fields(request)
	for i, word := range words {
		if (word == "function" || word == "type" || word == "struct") && i+1 < len(words) {
			return words[i+1]
		}
	}

	// Default: return last word if nothing found
	if len(words) > 0 {
		return words[len(words)-1]
	}
	return ""
}
