package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestExecutor_ReadTool(t *testing.T) {
	// Create temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	exec := NewExecutor()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{tmpDir},
		ReadOnly:           true,
		MaxFileSize:        1024 * 1024,
	}

	// Test basic read
	result, err := exec.Execute(ctx, "read", ToolParams{
		Path: testFile,
	}, perms)

	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if !result.Success {
		t.Errorf("read should succeed")
	}

	// Test with offset and limit
	result, err = exec.Execute(ctx, "read", ToolParams{
		Path:   testFile,
		Offset: 1,
		Limit:  2,
	}, perms)

	if err != nil {
		t.Fatalf("read with offset/limit failed: %v", err)
	}

	if !result.Success {
		t.Errorf("read with offset/limit should succeed")
	}
}

func TestExecutor_PermissionValidation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	exec := NewExecutor()
	ctx := context.Background()

	// Test with disallowed directory
	perms := schema.Permissions{
		AllowedDirectories: []string{"/other/path"},
		ReadOnly:           true,
	}

	_, err := exec.Execute(ctx, "read", ToolParams{
		Path: testFile,
	}, perms)

	if err == nil {
		t.Errorf("expected permission error")
	}
}

func TestExecutor_PathValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file with double dots in filename
	testFile := filepath.Join(tmpDir, "file..txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	exec := NewExecutor()
	ctx := context.Background()
	perms := schema.Permissions{
		AllowedDirectories: []string{tmpDir},
		ReadOnly:           true,
		MaxFileSize:        1024,
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with double dots in filename",
			path:    testFile,
			wantErr: false, // should be allowed after our fix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := exec.Execute(ctx, "read", ToolParams{
				Path: tt.path,
			}, perms)

			if tt.wantErr && err == nil {
				t.Errorf("expected error for path %q", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for path %q: %v", tt.path, err)
			}
		})
	}
}

func TestExecutor_GlobTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte("test"), 0644)

	exec := NewExecutor()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{tmpDir},
		ReadOnly:           true,
	}

	result, err := exec.Execute(ctx, "glob", ToolParams{
		Path:    tmpDir,
		Pattern: "*.go",
	}, perms)

	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}

	if !result.Success {
		t.Errorf("glob should succeed")
	}

	matches, ok := result.Output.([]string)
	if !ok {
		t.Fatalf("expected []string output")
	}

	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestExecutor_ListTool(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	exec := NewExecutor()
	ctx := context.Background()

	perms := schema.Permissions{
		AllowedDirectories: []string{tmpDir},
		ReadOnly:           true,
	}

	result, err := exec.Execute(ctx, "list", ToolParams{
		Path: tmpDir,
	}, perms)

	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if !result.Success {
		t.Errorf("list should succeed")
	}

	files, ok := result.Output.([]string)
	if !ok {
		t.Fatalf("expected []string output")
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}

func TestIsPathAllowed(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		allowedDirs []string
		want        bool
	}{
		{
			name:        "path within allowed directory",
			path:        "/home/user/project/file.go",
			allowedDirs: []string{"/home/user/project"},
			want:        true,
		},
		{
			name:        "path outside allowed directory",
			path:        "/etc/passwd",
			allowedDirs: []string{"/home/user/project"},
			want:        false,
		},
		{
			name:        "path with multiple allowed dirs",
			path:        "/opt/app/file.go",
			allowedDirs: []string{"/home/user", "/opt/app"},
			want:        true,
		},
		{
			name:        "empty allowed dirs",
			path:        "/any/path",
			allowedDirs: []string{},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathAllowed(tt.path, tt.allowedDirs)
			if got != tt.want {
				t.Errorf("isPathAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
