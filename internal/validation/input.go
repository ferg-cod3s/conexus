// Package validation provides security-focused input validation utilities.
package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// ErrInvalidPath indicates an invalid or unsafe path.
	ErrInvalidPath = fmt.Errorf("invalid or unsafe path")

	// ErrPathTraversal indicates a path traversal attempt.
	ErrPathTraversal = fmt.Errorf("path traversal attempt detected")

	// ErrAbsolutePathRequired indicates an absolute path was required but not provided.
	ErrAbsolutePathRequired = fmt.Errorf("absolute path required")
)

// PathValidator provides secure path validation within a root directory scope.
// It uses os.Root (Go 1.24+) to ensure paths stay within bounds and safely handle symlinks.
//
// Security properties:
//   - Prevents path traversal outside root directory
//   - Safely follows symlinks (rejects absolute symlinks)
//   - Thread-safe for concurrent use
//
// Usage:
//
//	validator, err := NewPathValidator("/safe/root")
//	if err != nil {
//	    return err
//	}
//	defer validator.Close()
//
//	safePath, err := validator.ValidatePath(userInput)
//	if err != nil {
//	    return fmt.Errorf("invalid path: %w", err)
//	}
type PathValidator struct {
	root     *os.Root
	rootPath string
}

// NewPathValidator creates a new PathValidator scoped to the given root directory.
// The root must be an absolute path to an existing directory.
func NewPathValidator(rootPath string) (*PathValidator, error) {
	if !filepath.IsAbs(rootPath) {
		return nil, fmt.Errorf("%w: root must be absolute path", ErrInvalidPath)
	}

	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("root path does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory")
	}

	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open root directory: %w", err)
	}

	return &PathValidator{root: root, rootPath: rootPath}, nil
}

// ValidatePath checks if a path is safe and within the validator's root scope.
// It returns the cleaned, validated path or an error.
// os.Root automatically ensures the path stays within bounds and handles symlinks safely.
func (v *PathValidator) ValidatePath(path string) (string, error) {
	// Basic checks first
	if path == "" {
		return "", fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Check for dangerous patterns BEFORE cleaning
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("%w: contains parent directory reference", ErrPathTraversal)
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Use os.Root to verify path is within bounds
	// This automatically handles symlinks and prevents escapes
	_, err := v.root.Stat(cleaned)
	if err != nil {
		// Path doesn't exist - try to validate that at least it's within root bounds
		// by checking the parent exists (if there is one)
		parent := filepath.Dir(cleaned)
		if parent != "." && parent != cleaned {
			_, parentErr := v.root.Stat(parent)
			if parentErr != nil {
				// Parent doesn't exist either - this is likely invalid
				// However, for paths like "nonexistent" in root, parent is "."
				// which os.Root.Stat will fail on. Let's be more lenient.
				// We'll allow the path if it's a simple relative path with no subdirectories
				if !strings.Contains(cleaned, string(filepath.Separator)) {
					// Single component path like "file.txt" - allow it
					return cleaned, nil
				}
				return "", fmt.Errorf("%w: parent directory not accessible: %v", ErrInvalidPath, parentErr)
			}
		}
		// Parent exists or path is single-component, so this is valid for creation
	}

	// os.Root.Stat succeeding means the path is safe and within root
	return cleaned, nil
}

// SanitizePath performs basic path sanitization without root scoping.
// Use ValidatePath for full security checks.
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Check for dangerous patterns BEFORE cleaning
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("%w: contains parent directory reference", ErrPathTraversal)
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Double-check after cleaning (belt and suspenders)
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("%w: cleaned path still contains ..", ErrPathTraversal)
	}

	return cleaned, nil
}

// IsPathSafe performs lightweight checks on a path without filesystem access.
// It checks for common unsafe patterns but doesn't verify the path exists.
func IsPathSafe(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Check for null bytes
	if strings.ContainsRune(path, '\x00') {
		return fmt.Errorf("%w: contains null byte", ErrInvalidPath)
	}

	// Check for parent directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("%w: contains parent directory reference", ErrPathTraversal)
	}

	// Check cleaned path
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("%w: cleaned path contains ..", ErrPathTraversal)
	}

	return nil
}

// ValidateConfigPath validates a configuration file path.
// Config files must be absolute paths to prevent ambiguity.
func ValidateConfigPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: empty config path", ErrInvalidPath)
	}

	// Config paths must be absolute
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("%w: config path must be absolute", ErrAbsolutePathRequired)
	}

	// Basic safety checks
	if err := IsPathSafe(path); err != nil {
		return "", err
	}

	cleaned := filepath.Clean(path)
	return cleaned, nil
}

// MustValidatePath is like ValidatePath but panics on error.
// Use only in initialization code where invalid paths are programmer errors.
func (v *PathValidator) MustValidatePath(path string) string {
	validated, err := v.ValidatePath(path)
	if err != nil {
		panic(fmt.Errorf("path validation failed: %w", err))
	}
	return validated
}

// Close releases resources associated with the PathValidator.
// After Close, the validator should not be used.
func (v *PathValidator) Close() error {
	if v.root != nil {
		return v.root.Close()
	}
	return nil
}

// ValidateAgentID validates that an agent identifier contains only safe characters.
// Agent IDs should only contain alphanumeric characters, hyphens, and underscores.
// This prevents command injection when constructing agent binary paths.
func ValidateAgentID(agentID string) error {
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	// Check length (reasonable limit)
	if len(agentID) > 128 {
		return fmt.Errorf("agent ID too long (max 128 characters)")
	}

	// Check for only allowed characters: alphanumeric, hyphen, underscore
	for _, r := range agentID {
		if !((r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '-' ||
			r == '_') {
			return fmt.Errorf("agent ID contains invalid character: %q", r)
		}
	}

	// Prevent starting with hyphen (could be interpreted as flag)
	if agentID[0] == '-' {
		return fmt.Errorf("agent ID cannot start with hyphen")
	}

	return nil
}
