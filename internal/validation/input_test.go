package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPathValidator(t *testing.T) {
	tests := []struct {
		name      string
		rootPath  string
		setupFunc func() string
		wantErr   bool
		errType   error
	}{
		{
			name: "valid absolute directory",
			setupFunc: func() string {
				dir := t.TempDir()
				return dir
			},
			wantErr: false,
		},
		{
			name:     "relative path rejected",
			rootPath: "relative/path",
			wantErr:  true,
			errType:  ErrInvalidPath,
		},
		{
			name:     "non-existent path",
			rootPath: "/nonexistent/path/should/not/exist",
			wantErr:  true,
		},
		{
			name: "file instead of directory",
			setupFunc: func() string {
				dir := t.TempDir()
				file := filepath.Join(dir, "file.txt")
				_ = os.WriteFile(file, []byte("test"), 0600)
				return file
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootPath := tt.rootPath
			if tt.setupFunc != nil {
				rootPath = tt.setupFunc()
			}

			validator, err := NewPathValidator(rootPath)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, validator)
			defer validator.Close()
		})
	}
}

func TestPathValidator_ValidatePath(t *testing.T) {
	// Setup test directory structure
	rootDir := t.TempDir()

	// Create subdirectories
	subDir := filepath.Join(rootDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0750))

	// Create a file
	testFile := filepath.Join(subDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0600))

	// Create symlink within root
	symlinkInRoot := filepath.Join(rootDir, "link_to_subdir")
	require.NoError(t, os.Symlink(subDir, symlinkInRoot))

	validator, err := NewPathValidator(rootDir)
	require.NoError(t, err)
	defer validator.Close()

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType error
	}{
		{
			name:    "valid relative path to existing file",
			path:    "subdir/test.txt",
			wantErr: false,
		},
		{
			name:    "valid relative path to existing directory",
			path:    "subdir",
			wantErr: false,
		},
		{
			name:    "valid path with . components",
			path:    "./subdir/./test.txt",
			wantErr: false,
		},
		{
			name:    "path traversal with ..",
			path:    "subdir/../../etc/passwd",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "path with parent reference",
			path:    "../outside",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "valid non-existent path in existing directory",
			path:    "subdir/newfile.txt",
			wantErr: false,
		},
		{
			name:    "symlink within root (allowed)",
			path:    "link_to_subdir",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidatePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result)

			// Validated path should not contain ..
			assert.NotContains(t, result, "..")
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
		errType error
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "simple relative path",
			path:    "foo/bar",
			want:    "foo/bar",
			wantErr: false,
		},
		{
			name:    "path with dot components",
			path:    "./foo/./bar",
			want:    "foo/bar",
			wantErr: false,
		},
		{
			name:    "path with parent directory",
			path:    "foo/../bar",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "absolute path",
			path:    "/foo/bar",
			want:    "/foo/bar",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "../../etc/passwd",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "path with redundant separators",
			path:    "foo//bar///baz",
			want:    "foo/bar/baz",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsPathSafe(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType error
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "safe relative path",
			path:    "foo/bar/baz",
			wantErr: false,
		},
		{
			name:    "safe absolute path",
			path:    "/foo/bar/baz",
			wantErr: false,
		},
		{
			name:    "path with null byte",
			path:    "foo\x00bar",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "path traversal",
			path:    "../etc/passwd",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "path with dot components only",
			path:    "./foo/./bar",
			wantErr: false,
		},
		{
			name:    "complex traversal attempt",
			path:    "foo/../../bar",
			wantErr: true,
			errType: ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsPathSafe(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestValidateConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType error
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "relative path rejected",
			path:    "config.yml",
			wantErr: true,
			errType: ErrAbsolutePathRequired,
		},
		{
			name:    "valid absolute path",
			path:    "/etc/conexus/config.yml",
			wantErr: false,
		},
		{
			name:    "absolute path with traversal",
			path:    "/etc/../../../etc/passwd",
			wantErr: true,
			errType: ErrPathTraversal,
		},
		{
			name:    "absolute path with null byte",
			path:    "/etc/config\x00.yml",
			wantErr: true,
			errType: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateConfigPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.True(t, filepath.IsAbs(result))
			assert.NotContains(t, result, "..")
		})
	}
}

func TestPathValidator_MustValidatePath(t *testing.T) {
	rootDir := t.TempDir()
	validator, err := NewPathValidator(rootDir)
	require.NoError(t, err)
	defer validator.Close()

	t.Run("valid path does not panic", func(t *testing.T) {
		// Create the valid directory structure
		validDir := filepath.Join(rootDir, "valid")
		err := os.Mkdir(validDir, 0755)
		require.NoError(t, err)

		assert.NotPanics(t, func() {
			result := validator.MustValidatePath("valid/path")
			assert.NotEmpty(t, result)
		})
	})

	t.Run("invalid path panics", func(t *testing.T) {
		assert.Panics(t, func() {
			validator.MustValidatePath("../../../etc/passwd")
		})
	})
}

func TestPathValidator_Close(t *testing.T) {
	rootDir := t.TempDir()
	validator, err := NewPathValidator(rootDir)
	require.NoError(t, err)

	err = validator.Close()
	assert.NoError(t, err)

	// Calling Close again should not panic
	err = validator.Close()
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkSanitizePath(b *testing.B) {
	path := "foo/bar/baz/qux/file.txt"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SanitizePath(path)
	}
}

func BenchmarkIsPathSafe(b *testing.B) {
	path := "foo/bar/baz/qux/file.txt"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsPathSafe(path)
	}
}

func BenchmarkPathValidator_ValidatePath(b *testing.B) {
	rootDir := b.TempDir()
	validator, err := NewPathValidator(rootDir)
	require.NoError(b, err)
	defer validator.Close()

	// Create test file
	subdir := filepath.Join(rootDir, "subdir")
	require.NoError(b, os.Mkdir(subdir, 0750))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.ValidatePath("subdir/file.txt")
	}
}

func TestValidateAgentID(t *testing.T) {
	tests := []struct {
		name    string
		agentID string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple ID",
			agentID: "test-agent",
			wantErr: false,
		},
		{
			name:    "valid with underscores",
			agentID: "test_agent_123",
			wantErr: false,
		},
		{
			name:    "valid alphanumeric",
			agentID: "agent123",
			wantErr: false,
		},
		{
			name:    "valid mixed case",
			agentID: "TestAgent-v1",
			wantErr: false,
		},
		{
			name:    "empty ID",
			agentID: "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "path traversal attempt",
			agentID: "../etc/passwd",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "command injection attempt with semicolon",
			agentID: "agent; rm -rf /",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "command injection with pipe",
			agentID: "agent|cat /etc/passwd",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "starts with hyphen",
			agentID: "-agent",
			wantErr: true,
			errMsg:  "cannot start with hyphen",
		},
		{
			name:    "contains slash",
			agentID: "agent/subagent",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "contains backslash",
			agentID: "agent\\subagent",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "contains space",
			agentID: "test agent",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "contains dollar sign",
			agentID: "$agent",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "contains backtick",
			agentID: "`agent`",
			wantErr: true,
			errMsg:  "invalid character",
		},
		{
			name:    "too long",
			agentID: string(make([]byte, 129)),
			wantErr: true,
			errMsg:  "too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentID(tt.agentID)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
