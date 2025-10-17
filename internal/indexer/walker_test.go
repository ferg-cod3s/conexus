package indexer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWalker_Walk(t *testing.T) {
	// Create a temporary test directory
	tmpDir := t.TempDir()

	// Create test file structure
	files := map[string]string{
		"main.go":                  "package main",
		"README.md":                "# Project",
		"internal/app/app.go":      "package app",
		"internal/app/app_test.go": "package app",
		"vendor/lib/lib.go":        "package lib",
		"node_modules/pkg/pkg.js":  "module.exports = {}",
		".git/config":              "[core]",
		"build/output.bin":         "binary",
		"data.log":                 "logs",
		"large.txt":                strings.Repeat("x", 2000),
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	tests := []struct {
		name            string
		ignorePatterns  []string
		maxFileSize     int64
		expectedFiles   []string
		unexpectedFiles []string
	}{
		{
			name:           "no patterns - walk all files",
			ignorePatterns: nil,
			maxFileSize:    0,
			expectedFiles: []string{
				"main.go",
				"README.md",
				"internal/app/app.go",
				"internal/app/app_test.go",
				"vendor/lib/lib.go",
				"node_modules/pkg/pkg.js",
				".git/config",
				"build/output.bin",
				"data.log",
				"large.txt",
			},
		},
		{
			name:           "ignore directories",
			ignorePatterns: []string{".git/", "vendor/", "node_modules/"},
			maxFileSize:    0,
			expectedFiles: []string{
				"main.go",
				"README.md",
				"internal/app/app.go",
				"internal/app/app_test.go",
				"build/output.bin",
				"data.log",
				"large.txt",
			},
			unexpectedFiles: []string{
				".git/config",
				"vendor/lib/lib.go",
				"node_modules/pkg/pkg.js",
			},
		},
		{
			name:           "ignore file patterns",
			ignorePatterns: []string{"*.log", "*.bin"},
			maxFileSize:    0,
			expectedFiles: []string{
				"main.go",
				"README.md",
				"internal/app/app.go",
				"internal/app/app_test.go",
				"large.txt",
			},
			unexpectedFiles: []string{
				"data.log",
				"build/output.bin",
			},
		},
		{
			name:           "default ignore patterns",
			ignorePatterns: DefaultIgnorePatterns(),
			maxFileSize:    0,
			expectedFiles: []string{
				"main.go",
				"README.md",
				"internal/app/app.go",
				"internal/app/app_test.go",
				"data.log",
				"large.txt",
			},
			unexpectedFiles: []string{
				".git/config",
				"vendor/lib/lib.go",
				"node_modules/pkg/pkg.js",
				"build/output.bin",
			},
		},
		{
			name:           "max file size filter",
			ignorePatterns: nil,
			maxFileSize:    1000, // 1KB
			expectedFiles: []string{
				"main.go",
				"README.md",
				"internal/app/app.go",
				"internal/app/app_test.go",
				"vendor/lib/lib.go",
				"node_modules/pkg/pkg.js",
				".git/config",
				"build/output.bin",
				"data.log",
			},
			unexpectedFiles: []string{
				"large.txt", // 2000 bytes > 1000
			},
		},
		{
			name:           "anchored patterns",
			ignorePatterns: []string{"/build/"},
			maxFileSize:    0,
			expectedFiles: []string{
				"main.go",
				"README.md",
				"internal/app/app.go",
				"data.log",
			},
			unexpectedFiles: []string{
				"build/output.bin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walker := NewFileWalker(tt.maxFileSize)
			var foundFiles []string

			err := walker.Walk(context.Background(), tmpDir, tt.ignorePatterns, func(path string, info fs.FileInfo) error {
				relPath, err := filepath.Rel(tmpDir, path)
				require.NoError(t, err)
				foundFiles = append(foundFiles, filepath.ToSlash(relPath))
				return nil
			})

			require.NoError(t, err)

			// Check expected files are present
			for _, expected := range tt.expectedFiles {
				assert.Contains(t, foundFiles, expected, "expected file not found")
			}

			// Check unexpected files are absent
			for _, unexpected := range tt.unexpectedFiles {
				assert.NotContains(t, foundFiles, unexpected, "unexpected file found")
			}
		})
	}
}

func TestFileWalker_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files
	for i := 0; i < 100; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		require.NoError(t, os.WriteFile(path, []byte("data"), 0644))
	}

	walker := NewFileWalker(0)
	ctx, cancel := context.WithCancel(context.Background())

	fileCount := 0
	err := walker.Walk(ctx, tmpDir, nil, func(path string, info fs.FileInfo) error {
		fileCount++
		if fileCount == 10 {
			cancel() // Cancel after 10 files
		}
		return nil
	})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 10, fileCount)
}

func TestFileWalker_CallbackError(t *testing.T) {
	tmpDir := t.TempDir()

	files := []string{"a.txt", "b.txt", "c.txt"}
	for _, f := range files {
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, f), []byte("data"), 0644))
	}

	walker := NewFileWalker(0)
	expectedErr := fmt.Errorf("callback error")

	err := walker.Walk(context.Background(), tmpDir, nil, func(path string, info fs.FileInfo) error {
		if strings.HasSuffix(path, "b.txt") {
			return expectedErr
		}
		return nil
	})

	assert.ErrorIs(t, err, expectedErr)
}

func TestFileWalker_InvalidRoot(t *testing.T) {
	walker := NewFileWalker(0)
	err := walker.Walk(context.Background(), "/nonexistent/path", nil, func(path string, info fs.FileInfo) error {
		return nil
	})

	assert.Error(t, err)
}

func TestPatternMatcher_Match(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		isDir    bool
		ignored  bool
	}{
		// Basic patterns
		{
			name:     "exact file match",
			patterns: []string{"test.txt"},
			path:     "test.txt",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "exact file no match",
			patterns: []string{"test.txt"},
			path:     "other.txt",
			isDir:    false,
			ignored:  false,
		},
		{
			name:     "wildcard extension",
			patterns: []string{"*.log"},
			path:     "app.log",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "wildcard extension nested",
			patterns: []string{"*.log"},
			path:     "logs/app.log",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "wildcard extension no match",
			patterns: []string{"*.log"},
			path:     "app.txt",
			isDir:    false,
			ignored:  false,
		},

		// Directory patterns
		{
			name:     "directory pattern - match dir",
			patterns: []string{"node_modules/"},
			path:     "node_modules",
			isDir:    true,
			ignored:  true,
		},
		{
			name:     "directory pattern - file inside dir",
			patterns: []string{"node_modules/"},
			path:     "node_modules/pkg/index.js",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "directory pattern - file not inside",
			patterns: []string{"node_modules/"},
			path:     "src/index.js",
			isDir:    false,
			ignored:  false,
		},

		// Anchored patterns
		{
			name:     "anchored pattern - match from root",
			patterns: []string{"/build/"},
			path:     "build",
			isDir:    true,
			ignored:  true,
		},
		{
			name:     "anchored pattern - no match nested",
			patterns: []string{"/build/"},
			path:     "src/build",
			isDir:    true,
			ignored:  false,
		},

		// Negation patterns
		{
			name:     "negation - ignore all except",
			patterns: []string{"*.txt", "!important.txt"},
			path:     "test.txt",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "negation - exception",
			patterns: []string{"*.txt", "!important.txt"},
			path:     "important.txt",
			isDir:    false,
			ignored:  false,
		},

		// Path matching
		{
			name:     "path pattern - match full path",
			patterns: []string{"src/test/*.go"},
			path:     "src/test/main.go",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "path pattern - no match different path",
			patterns: []string{"src/test/*.go"},
			path:     "src/main.go",
			isDir:    false,
			ignored:  false,
		},

		// Comments and empty lines
		{
			name:     "ignore comments",
			patterns: []string{"# This is a comment", "*.log"},
			path:     "app.log",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "ignore empty lines",
			patterns: []string{"", "*.log", ""},
			path:     "app.log",
			isDir:    false,
			ignored:  true,
		},

		// Complex scenarios
		{
			name:     "multiple patterns - first match",
			patterns: []string{"*.log", "*.txt"},
			path:     "app.log",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "multiple patterns - second match",
			patterns: []string{"*.log", "*.txt"},
			path:     "data.txt",
			isDir:    false,
			ignored:  true,
		},
		{
			name:     "dir-only pattern on file",
			patterns: []string{"test/"},
			path:     "test",
			isDir:    false,
			ignored:  false,
		},
		{
			name:     "dir-only pattern on dir",
			patterns: []string{"test/"},
			path:     "test",
			isDir:    true,
			ignored:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := newPatternMatcher(tt.patterns)
			result := matcher.match(tt.path, tt.isDir)
			assert.Equal(t, tt.ignored, result, "unexpected match result")
		})
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	patterns := DefaultIgnorePatterns()

	// Verify it includes common patterns
	assert.Contains(t, patterns, ".git/")
	assert.Contains(t, patterns, "node_modules/")
	assert.Contains(t, patterns, "vendor/")
	assert.Contains(t, patterns, "*.pyc")
	assert.Contains(t, patterns, ".DS_Store")

	// Verify all are non-empty
	for _, p := range patterns {
		assert.NotEmpty(t, p)
	}
}

func TestLoadGitignore(t *testing.T) {
	t.Run("valid gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")

		content := `# Comment line
*.log
node_modules/

# Another comment
/build/
vendor/
`
		require.NoError(t, os.WriteFile(gitignorePath, []byte(content), 0644))

		patterns, err := LoadGitignore(gitignorePath, tmpDir)
		require.NoError(t, err)

		expected := []string{"*.log", "node_modules/", "/build/", "vendor/"}
		assert.Equal(t, expected, patterns)
	})

	t.Run("file not exists", func(t *testing.T) {
		patterns, err := LoadGitignore("/nonexistent/.gitignore", "/")
		assert.NoError(t, err)
		assert.Nil(t, patterns)
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		require.NoError(t, os.WriteFile(gitignorePath, []byte(""), 0644))

		patterns, err := LoadGitignore(gitignorePath, tmpDir)
		require.NoError(t, err)
		assert.Empty(t, patterns)
	})

	t.Run("only comments", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		content := `# Comment 1
# Comment 2
# Comment 3
`
		require.NoError(t, os.WriteFile(gitignorePath, []byte(content), 0644))

		patterns, err := LoadGitignore(gitignorePath, tmpDir)
		require.NoError(t, err)
		assert.Empty(t, patterns)
	})
}

func TestFileWalker_RealWorldScenario(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a realistic Go project structure
	structure := map[string]string{
		"go.mod":                           "module example.com/project",
		"main.go":                          "package main",
		"README.md":                        "# Project",
		"cmd/server/main.go":               "package main",
		"internal/app/app.go":              "package app",
		"internal/app/app_test.go":         "package app",
		"pkg/utils/utils.go":               "package utils",
		"vendor/github.com/lib/lib.go":     "package lib",
		".git/HEAD":                        "ref: refs/heads/main",
		".github/workflows/ci.yml":         "name: CI",
		"node_modules/package/index.js":    "module.exports = {}",
		"build/bin/server":                 "binary",
		"testdata/sample.txt":              "test data",
		".env":                             "SECRET=value",
		"coverage.out":                     "coverage data",
	}

	for path, content := range structure {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	// Use realistic ignore patterns
	ignorePatterns := append(DefaultIgnorePatterns(),
		".env",
		"coverage.out",
		"testdata/",
	)

	walker := NewFileWalker(0)
	var foundFiles []string

	err := walker.Walk(context.Background(), tmpDir, ignorePatterns, func(path string, info fs.FileInfo) error {
		relPath, _ := filepath.Rel(tmpDir, path)
		foundFiles = append(foundFiles, filepath.ToSlash(relPath))
		return nil
	})

	require.NoError(t, err)

	// Should include source files
	assert.Contains(t, foundFiles, "main.go")
	assert.Contains(t, foundFiles, "cmd/server/main.go")
	assert.Contains(t, foundFiles, "internal/app/app.go")
	assert.Contains(t, foundFiles, "pkg/utils/utils.go")
	assert.Contains(t, foundFiles, "README.md")
	assert.Contains(t, foundFiles, ".github/workflows/ci.yml")

	// Should exclude ignored paths
	assert.NotContains(t, foundFiles, ".git/HEAD")
	assert.NotContains(t, foundFiles, "vendor/github.com/lib/lib.go")
	assert.NotContains(t, foundFiles, "node_modules/package/index.js")
	assert.NotContains(t, foundFiles, "build/bin/server")
	assert.NotContains(t, foundFiles, ".env")
	assert.NotContains(t, foundFiles, "coverage.out")
	assert.NotContains(t, foundFiles, "testdata/sample.txt")
}

func BenchmarkFileWalker_Walk(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a moderate-sized file tree
	for i := 0; i < 100; i++ {
		dir := filepath.Join(tmpDir, fmt.Sprintf("dir%d", i))
		require.NoError(b, os.MkdirAll(dir, 0755))
		for j := 0; j < 10; j++ {
			path := filepath.Join(dir, fmt.Sprintf("file%d.go", j))
			require.NoError(b, os.WriteFile(path, []byte("package main"), 0644))
		}
	}

	walker := NewFileWalker(0)
	patterns := DefaultIgnorePatterns()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		_ = walker.Walk(context.Background(), tmpDir, patterns, func(path string, info fs.FileInfo) error {
			count++
			return nil
		})
	}
}

func BenchmarkPatternMatcher_Match(b *testing.B) {
	patterns := []string{
		".git/",
		"node_modules/",
		"vendor/",
		"*.log",
		"*.pyc",
		"build/",
		"/dist/",
	}

	matcher := newPatternMatcher(patterns)
	testPaths := []string{
		"src/main.go",
		"internal/app/app.go",
		"vendor/lib/lib.go",
		"node_modules/pkg/index.js",
		"logs/app.log",
		"build/output.bin",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			matcher.match(path, false)
		}
	}
}
