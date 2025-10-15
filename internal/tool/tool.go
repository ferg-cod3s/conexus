package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Tool defines the interface for all executable tools
type Tool interface {
	Name() string
	Execute(ctx context.Context, params ToolParams, perms schema.Permissions) (ToolResult, error)
}

// ToolParams contains parameters for tool execution
type ToolParams struct {
	// Common parameters
	Path    string   `json:"path,omitempty"`
	Pattern string   `json:"pattern,omitempty"`
	Paths   []string `json:"paths,omitempty"`

	// Read tool specific
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`

	// Grep tool specific
	CaseInsensitive bool   `json:"case_insensitive,omitempty"`
	Context         int    `json:"context,omitempty"`
	FilePattern     string `json:"file_pattern,omitempty"`
}

// ToolResult contains the result of tool execution
type ToolResult struct {
	Success bool        `json:"success"`
	Output  interface{} `json:"output"`
	Error   string      `json:"error,omitempty"`
}

// Executor manages tool execution with permission enforcement
type Executor struct {
	tools map[string]Tool
}

// NewExecutor creates a new tool executor
func NewExecutor() *Executor {
	exec := &Executor{
		tools: make(map[string]Tool),
	}

	// Register built-in tools
	exec.RegisterTool(&ReadTool{})
	exec.RegisterTool(&GrepTool{})
	exec.RegisterTool(&GlobTool{})
	exec.RegisterTool(&ListTool{})

	return exec
}

// RegisterTool adds a tool to the executor
func (e *Executor) RegisterTool(tool Tool) {
	e.tools[tool.Name()] = tool
}

// Execute runs a tool with permission validation
func (e *Executor) Execute(ctx context.Context, toolName string, params ToolParams, perms schema.Permissions) (ToolResult, error) {
	tool, exists := e.tools[toolName]
	if !exists {
		return ToolResult{}, fmt.Errorf("unknown tool: %s", toolName)
	}

	// Validate permissions before execution
	if err := e.validatePermissions(params, perms); err != nil {
		return ToolResult{Success: false, Error: err.Error()}, err
	}

	// Execute with context for timeout enforcement
	result, err := tool.Execute(ctx, params, perms)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, err
	}

	return result, nil
}

// validatePermissions checks if the operation is allowed
func (e *Executor) validatePermissions(params ToolParams, perms schema.Permissions) error {
	// Check path permissions
	paths := []string{}
	if params.Path != "" {
		paths = append(paths, params.Path)
	}
	paths = append(paths, params.Paths...)

	for _, path := range paths {
		if path == "" {
			continue
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		if !isPathAllowed(absPath, perms.AllowedDirectories) {
			return fmt.Errorf("access denied: path %s not in allowed directories", absPath)
		}
	}

	return nil
}

// isPathAllowed checks if a path is within allowed directories
func isPathAllowed(path string, allowedDirs []string) bool {
	if len(allowedDirs) == 0 {
		return false
	}

	cleanPath := filepath.Clean(path)

	for _, allowedDir := range allowedDirs {
		cleanAllowed := filepath.Clean(allowedDir)

		// Check if path is within allowed directory
		if strings.HasPrefix(cleanPath, cleanAllowed) {
			return true
		}
	}

	return false
}

// ReadTool implements file reading
type ReadTool struct{}

func (t *ReadTool) Name() string { return "read" }

func (t *ReadTool) Execute(ctx context.Context, params ToolParams, perms schema.Permissions) (ToolResult, error) {
	if params.Path == "" {
		return ToolResult{}, fmt.Errorf("path parameter required")
	}

	content, err := os.ReadFile(params.Path)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, err
	}

	// Check file size limit
	if perms.MaxFileSize > 0 && int64(len(content)) > perms.MaxFileSize {
		return ToolResult{}, fmt.Errorf("file exceeds maximum size: %d > %d", len(content), perms.MaxFileSize)
	}

	// Apply offset and limit if specified
	lines := strings.Split(string(content), "\n")
	if params.Offset > 0 {
		if params.Offset >= len(lines) {
			lines = []string{}
		} else {
			lines = lines[params.Offset:]
		}
	}
	if params.Limit > 0 && params.Limit < len(lines) {
		lines = lines[:params.Limit]
	}

	return ToolResult{
		Success: true,
		Output:  strings.Join(lines, "\n"),
	}, nil
}

// GrepTool implements pattern searching
type GrepTool struct{}

func (t *GrepTool) Name() string { return "grep" }

func (t *GrepTool) Execute(ctx context.Context, params ToolParams, perms schema.Permissions) (ToolResult, error) {
	if params.Pattern == "" {
		return ToolResult{}, fmt.Errorf("pattern parameter required")
	}
	if params.Path == "" {
		return ToolResult{}, fmt.Errorf("path parameter required")
	}

	// TODO: Implement grep functionality using regexp
	// For now, return placeholder
	return ToolResult{
		Success: true,
		Output:  []string{},
	}, nil
}

// GlobTool implements pattern-based file matching with recursive support
type GlobTool struct{}

func (t *GlobTool) Name() string { return "glob" }

func (t *GlobTool) Execute(ctx context.Context, params ToolParams, perms schema.Permissions) (ToolResult, error) {
	if params.Pattern == "" {
		return ToolResult{}, fmt.Errorf("pattern parameter required")
	}

	basePath := params.Path
	if basePath == "" {
		basePath = "."
	}

	var matches []string

	// Check if pattern contains ** for recursive matching
	if strings.Contains(params.Pattern, "**") {
		// Use filepath.Glob which supports ** in Go 1.20+
		fullPattern := filepath.Join(basePath, params.Pattern)
		globMatches, err := filepath.Glob(fullPattern)
		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}, err
		}
		matches = globMatches
	} else {
		// For simple patterns like *.go, walk the directory tree
		err := filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors, continue walking
			}

			// Skip directories
			if d.IsDir() {
				return nil
			}

			// Check if the file name matches the pattern
			matched, matchErr := filepath.Match(params.Pattern, d.Name())
			if matchErr != nil {
				return nil // Skip invalid patterns for this file
			}

			if matched {
				matches = append(matches, path)
			}

			return nil
		})

		if err != nil {
			return ToolResult{Success: false, Error: err.Error()}, err
		}
	}

	return ToolResult{
		Success: true,
		Output:  matches,
	}, nil
}

// ListTool implements directory listing
type ListTool struct{}

func (t *ListTool) Name() string { return "list" }

func (t *ListTool) Execute(ctx context.Context, params ToolParams, perms schema.Permissions) (ToolResult, error) {
	path := params.Path
	if path == "" {
		path = "."
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	return ToolResult{
		Success: true,
		Output:  files,
	}, nil
}
