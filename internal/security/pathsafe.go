// Package security provides security utilities for Conexus.
package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var (
	// ErrPathTraversal indicates a path traversal attempt was detected
	ErrPathTraversal = errors.New("path traversal detected")
	// ErrInvalidPath indicates an invalid or unsafe path
	ErrInvalidPath = errors.New("invalid path")
)

// ValidatePath sanitizes and validates a file path to prevent path traversal attacks.
// It cleans the path, checks for traversal attempts, and optionally validates against a base directory.
//
// Parameters:
//   - path: The file path to validate
//   - basePath: Optional base directory to restrict access to (empty string to skip)
//
// Returns the cleaned, validated path or an error if validation fails.
func ValidatePath(path string, basePath string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Clean the path to resolve . and .. and remove redundant separators
	cleaned := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("%w: path contains '..'", ErrPathTraversal)
	}

	// If no base path specified, return cleaned path
	if basePath == "" {
		return cleaned, nil
	}

	// Clean base path
	cleanedBase := filepath.Clean(basePath)

	// Ensure path is within base directory
	// Convert both to absolute paths for comparison
	absPath := cleaned
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(cleanedBase, cleaned)
	}

	// Use filepath.Rel to check if path is within base
	rel, err := filepath.Rel(cleanedBase, absPath)
	if err != nil {
		return "", fmt.Errorf("%w: cannot compute relative path: %v", ErrInvalidPath, err)
	}

	// If relative path starts with .., it's outside base directory
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%w: path outside base directory", ErrPathTraversal)
	}

	return absPath, nil
}

// ValidatePathWithinBase validates that a path is within a base directory.
// This is a convenience wrapper around ValidatePath that requires a base path.
func ValidatePathWithinBase(path string, basePath string) (string, error) {
	if basePath == "" {
		return "", fmt.Errorf("%w: base path required", ErrInvalidPath)
	}
	return ValidatePath(path, basePath)
}

// SafeJoin safely joins path elements and validates the result.
// It's equivalent to filepath.Join but with traversal protection.
func SafeJoin(basePath string, elements ...string) (string, error) {
	if basePath == "" {
		return "", fmt.Errorf("%w: base path required", ErrInvalidPath)
	}

	// Join all elements
	joined := filepath.Join(append([]string{basePath}, elements...)...)

	// Validate the result
	return ValidatePath(joined, basePath)
}
