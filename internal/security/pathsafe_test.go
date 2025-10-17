package security

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		basePath  string
		wantErr   bool
		errType   error
		expectAbs bool // expect absolute path in result
	}{
		{
			name:     "empty path",
			path:     "",
			basePath: "",
			wantErr:  true,
			errType:  ErrInvalidPath,
		},
		{
			name:     "simple path no base",
			path:     "file.txt",
			basePath: "",
			wantErr:  false,
		},
		{
			name:     "path with dots cleaned",
			path:     "./dir/../file.txt",
			basePath: "",
			wantErr:  false,
		},
		{
			name:     "traversal attempt with ..",
			path:     "../etc/passwd",
			basePath: "",
			wantErr:  true,
			errType:  ErrPathTraversal,
		},
		{
			name:     "traversal in middle of path",
			path:     "safe/../../etc/passwd",
			basePath: "",
			wantErr:  true,
			errType:  ErrPathTraversal,
		},
		{
			name:      "valid path within base",
			path:      "subdir/file.txt",
			basePath:  "/tmp/safe",
			wantErr:   false,
			expectAbs: true,
		},
		{
			name:     "path escaping base",
			path:     "../../../etc/passwd",
			basePath: "/tmp/safe",
			wantErr:  true,
			errType:  ErrPathTraversal,
		},
		{
			name:      "absolute path within base",
			path:      "/tmp/safe/subdir/file.txt",
			basePath:  "/tmp/safe",
			wantErr:   false,
			expectAbs: true,
		},
		{
			name:     "absolute path outside base",
			path:     "/etc/passwd",
			basePath: "/tmp/safe",
			wantErr:  true,
			errType:  ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePath(tt.path, tt.basePath)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result)

			if tt.expectAbs {
				assert.True(t, filepath.IsAbs(result), "expected absolute path")
			}

			// Should never contain .. after validation
			assert.NotContains(t, result, "..")
		})
	}
}

func TestValidatePathWithinBase(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		basePath string
		wantErr  bool
		errType  error
	}{
		{
			name:     "missing base path",
			path:     "file.txt",
			basePath: "",
			wantErr:  true,
			errType:  ErrInvalidPath,
		},
		{
			name:     "valid path in base",
			path:     "file.txt",
			basePath: "/tmp/safe",
			wantErr:  false,
		},
		{
			name:     "traversal attempt",
			path:     "../etc/passwd",
			basePath: "/tmp/safe",
			wantErr:  true,
			errType:  ErrPathTraversal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePathWithinBase(tt.path, tt.basePath)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestSafeJoin(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		elements []string
		wantErr  bool
		errType  error
	}{
		{
			name:     "empty base",
			basePath: "",
			elements: []string{"file.txt"},
			wantErr:  true,
			errType:  ErrInvalidPath,
		},
		{
			name:     "simple join",
			basePath: "/tmp/safe",
			elements: []string{"subdir", "file.txt"},
			wantErr:  false,
		},
		{
			name:     "join with traversal attempt",
			basePath: "/tmp/safe",
			elements: []string{"..", "..", "etc", "passwd"},
			wantErr:  true,
			errType:  ErrPathTraversal,
		},
		{
			name:     "join with cleaned dots",
			basePath: "/tmp/safe",
			elements: []string{".", "subdir", ".", "file.txt"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeJoin(tt.basePath, tt.elements...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.True(t, filepath.IsAbs(result))
			assert.NotContains(t, result, "..")
		})
	}
}

// Benchmark path validation
func BenchmarkValidatePath(b *testing.B) {
	path := "subdir/file.txt"
	basePath := "/tmp/safe"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidatePath(path, basePath)
	}
}

func BenchmarkSafeJoin(b *testing.B) {
	basePath := "/tmp/safe"
	elements := []string{"subdir", "nested", "file.txt"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SafeJoin(basePath, elements...)
	}
}
