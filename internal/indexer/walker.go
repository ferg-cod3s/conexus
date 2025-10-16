package indexer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/security"
	"github.com/ferg-cod3s/conexus/internal/validation"
)

// FileWalker implements Walker with .gitignore-style pattern matching.
type FileWalker struct {
	maxFileSize int64 // Skip files larger than this (0 = no limit)
}

// NewFileWalker creates a new FileWalker with optional size limits.
func NewFileWalker(maxFileSize int64) *FileWalker {
	return &FileWalker{
		maxFileSize: maxFileSize,
	}
}

// Walk traverses the directory tree starting at root, respecting ignore patterns.
// Calls fn for each regular file that passes filters.
func (w *FileWalker) Walk(ctx context.Context, root string, ignorePatterns []string, fn func(path string, info fs.FileInfo) error) error {
	// Normalize root path
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Compile ignore patterns
	matcher := newPatternMatcher(ignorePatterns)

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Handle walk errors
		if err != nil {
			return err
		}

		// Get relative path for pattern matching
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Normalize relative path (use forward slashes)
		relPath = filepath.ToSlash(relPath)

		// Validate path to prevent traversal attacks
		if err := validation.IsPathSafe(relPath); err != nil {
			return fmt.Errorf("path validation failed for %s: %w", relPath, err)
		}

		// Check if path should be ignored
		if matcher.match(relPath, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories (we only call fn for files)
		if d.IsDir() {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", path, err)
		}

		// Skip if file exceeds max size
		if w.maxFileSize > 0 && info.Size() > w.maxFileSize {
			return nil
		}

		// Call the callback with the file
		return fn(path, info)
	})
}

// patternMatcher handles .gitignore-style pattern matching.
type patternMatcher struct {
	patterns []pattern
}

type pattern struct {
	raw       string
	negate    bool   // Pattern starts with !
	dirOnly   bool   // Pattern ends with /
	anchored  bool   // Pattern starts with /
	glob      string // Pattern for matching
}

// newPatternMatcher creates a matcher from ignore patterns.
func newPatternMatcher(patterns []string) *patternMatcher {
	m := &patternMatcher{
		patterns: make([]pattern, 0, len(patterns)),
	}

	for _, p := range patterns {
		if p == "" || strings.HasPrefix(p, "#") {
			continue // Skip empty lines and comments
		}

		pat := pattern{raw: p}

		// Check for negation
		if strings.HasPrefix(p, "!") {
			pat.negate = true
			p = p[1:]
		}

		// Check for directory-only
		if strings.HasSuffix(p, "/") {
			pat.dirOnly = true
			p = strings.TrimSuffix(p, "/")
		}

		// Check for anchored pattern
		if strings.HasPrefix(p, "/") {
			pat.anchored = true
			p = strings.TrimPrefix(p, "/")
		}

		pat.glob = p
		m.patterns = append(m.patterns, pat)
	}

	return m
}

// match checks if the path matches any ignore pattern.
// Returns true if the path should be ignored.
func (m *patternMatcher) match(relPath string, isDir bool) bool {
	// Track the current ignore state (last matching pattern wins)
	ignored := false

	for _, pat := range m.patterns {
		// For directory-only patterns (e.g., "node_modules/"):
		// - Match the directory itself
		// - Match all files/dirs inside that directory
		if pat.dirOnly {
			// Check if this is the directory itself
			if relPath == pat.glob && isDir {
				ignored = !pat.negate
				continue
			}
			// Check if this is inside the directory (file or subdir)
			if strings.HasPrefix(relPath, pat.glob+"/") {
				ignored = !pat.negate
				continue
			}
			// Also check for non-anchored directory patterns
			// e.g., "node_modules/" should match "a/b/node_modules/c.js"
			if !pat.anchored {
				parts := strings.Split(relPath, "/")
				for i := 0; i < len(parts); i++ {
					if parts[i] == pat.glob {
						// Found the directory in the path
						// If this is the dir itself or something inside it, match
						if i == len(parts)-1 && isDir {
							ignored = !pat.negate
							break
						}
						if i < len(parts)-1 {
							// Something inside the directory
							ignored = !pat.negate
							break
						}
					}
				}
			}
			continue
		}

		matches := m.matchPattern(pat, relPath, isDir)
		if matches {
			ignored = !pat.negate
		}
	}

	return ignored
}

// matchPattern checks if a single pattern matches the path.
func (m *patternMatcher) matchPattern(pat pattern, relPath string, isDir bool) bool {
	// Handle anchored patterns (match from root)
	if pat.anchored {
		matched, _ := filepath.Match(pat.glob, relPath)
		if matched {
			return true
		}
		// Also try matching with directory prefix
		if isDir {
			matched, _ = filepath.Match(pat.glob, relPath+"/")
			return matched
		}
		return false
	}

	// For non-anchored patterns, match against any path segment
	// e.g., "*.log" matches "a/b/c.log"
	matched, _ := filepath.Match(pat.glob, filepath.Base(relPath))
	if matched {
		return true
	}

	// Try matching the full path for patterns with path separators
	if strings.Contains(pat.glob, "/") {
		matched, _ := filepath.Match(pat.glob, relPath)
		if matched {
			return true
		}
	}

	// Try matching any suffix of the path
	// e.g., "foo/bar" matches "a/b/foo/bar/baz"
	parts := strings.Split(relPath, "/")
	for i := 0; i < len(parts); i++ {
		suffix := strings.Join(parts[i:], "/")
		matched, _ := filepath.Match(pat.glob, suffix)
		if matched {
			return true
		}
	}

	return false
}

// DefaultIgnorePatterns returns common patterns to ignore in codebases.
func DefaultIgnorePatterns() []string {
	return []string{
		".git/",
		".svn/",
		".hg/",
		"node_modules/",
		"vendor/",
		"target/",
		"build/",
		"dist/",
		"*.pyc",
		"*.pyo",
		"*.class",
		"*.o",
		"*.so",
		"*.dylib",
		"*.dll",
		"*.exe",
		".DS_Store",
		"Thumbs.db",
	}
}

// LoadGitignore reads a .gitignore file and returns its patterns.
func LoadGitignore(path string, basePath string) ([]string, error) {
	// G304: Validate path to prevent directory traversal
	if _, err := security.ValidatePathWithinBase(path, basePath); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// #nosec G304 - Path validated at line 271 with ValidatePathWithinBase
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read .gitignore: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	patterns := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	return patterns, nil
}
