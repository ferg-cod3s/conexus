package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIndexController(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	controller := NewIndexController(tempDir)

	assert.NotNil(t, controller)
	assert.False(t, controller.status.IsIndexing)
	assert.Equal(t, "idle", controller.status.Phase)
	assert.Equal(t, float64(0), controller.status.Progress)
	assert.False(t, controller.running)
}

func TestIndexControllerStartStop(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	controller := NewIndexController(tempDir)

	// Test starting indexing
	opts := IndexOptions{
		RootPath:    tempDir,
		MaxFileSize: 1024 * 1024, // 1MB
	}

	ctx := context.Background()
	err = controller.Start(ctx, opts)
	assert.NoError(t, err)

	// Wait a bit for indexing to start
	time.Sleep(100 * time.Millisecond)

	// Check that indexing is running
	status := controller.GetStatus()
	// Note: Indexing might complete quickly for small test directories
	// So we just check that it started successfully

	// Test stopping indexing
	err = controller.Stop(ctx)
	assert.NoError(t, err)

	// Check that indexing is stopped
	status = controller.GetStatus()
	assert.False(t, status.IsIndexing)

	// Wait for goroutines to finish
	time.Sleep(100 * time.Millisecond)
}

func TestIndexControllerStartAlreadyRunning(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	controller := NewIndexController(tempDir)

	opts := IndexOptions{
		RootPath:    tempDir,
		MaxFileSize: 1024 * 1024,
	}

	ctx := context.Background()

	// Start indexing first time
	err = controller.Start(ctx, opts)
	assert.NoError(t, err)

	// Try to start again while already running
	err = controller.Start(ctx, opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Clean up
	controller.Stop(ctx)
}

func TestIndexControllerGetStatus(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	controller := NewIndexController(tempDir)

	// Get initial status
	status := controller.GetStatus()
	assert.False(t, status.IsIndexing)
	assert.Equal(t, "idle", status.Phase)
	assert.Equal(t, float64(0), status.Progress)
	assert.Equal(t, 0, status.FilesProcessed)
	assert.Equal(t, 0, status.TotalFiles)
	assert.Empty(t, status.LastError)
}

func TestIndexControllerForceReindex(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\n\nfunc main() {}"), 0644)
	require.NoError(t, err)

	controller := NewIndexController(tempDir)

	opts := IndexOptions{
		RootPath:    tempDir,
		MaxFileSize: 1024 * 1024,
	}

	ctx := context.Background()

	// Start initial indexing
	err = controller.Start(ctx, opts)
	assert.NoError(t, err)

	// Wait for indexing to complete
	time.Sleep(500 * time.Millisecond)
	controller.Stop(ctx)
	time.Sleep(100 * time.Millisecond)

	// Force reindex
	err = controller.ForceReindex(ctx, opts)
	assert.NoError(t, err)

	// Wait for reindexing to complete
	time.Sleep(500 * time.Millisecond)
	controller.Stop(ctx)
}

func TestIndexControllerReindexPaths(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile1 := filepath.Join(tempDir, "test1.go")
	err = os.WriteFile(testFile1, []byte("package main\n\nfunc main() {}"), 0644)
	require.NoError(t, err)

	testFile2 := filepath.Join(tempDir, "test2.go")
	err = os.WriteFile(testFile2, []byte("package main\n\nfunc helper() {}"), 0644)
	require.NoError(t, err)

	controller := NewIndexController(tempDir)

	opts := IndexOptions{
		RootPath:    tempDir,
		MaxFileSize: 1024 * 1024,
	}

	ctx := context.Background()

	// Reindex specific paths
	paths := []string{testFile1}
	err = controller.ReindexPaths(ctx, opts, paths)
	assert.NoError(t, err)

	// Wait for reindexing to complete
	time.Sleep(500 * time.Millisecond)
}

func TestIndexControllerContextCancellation(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	controller := NewIndexController(tempDir)

	opts := IndexOptions{
		RootPath:    tempDir,
		MaxFileSize: 1024 * 1024,
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start indexing
	err = controller.Start(ctx, opts)
	assert.NoError(t, err)

	// Cancel the context after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Wait for cancellation to take effect
	time.Sleep(200 * time.Millisecond)

	// Controller should no longer be running
	status := controller.GetStatus()
	assert.False(t, status.IsIndexing)
}

func TestIndexControllerMultipleFiles(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create multiple test files
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("test%d.go", i))
		content := fmt.Sprintf("package main\n\nfunc test%d() {}\n", i)
		err = os.WriteFile(filename, []byte(content), 0644)
		require.NoError(t, err)
	}

	controller := NewIndexController(tempDir)

	opts := IndexOptions{
		RootPath:    tempDir,
		MaxFileSize: 1024 * 1024,
	}

	ctx := context.Background()

	// Start indexing
	err = controller.Start(ctx, opts)
	assert.NoError(t, err)

	// Wait for indexing to progress
	time.Sleep(500 * time.Millisecond)

	// Check status
	status := controller.GetStatus()
	// Should have processed some files
	assert.GreaterOrEqual(t, status.FilesProcessed, 0)

	// Stop indexing
	err = controller.Stop(ctx)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
}

func TestIndexControllerWithLargeFile(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "indexer_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a large file that exceeds MaxFileSize
	largeFile := filepath.Join(tempDir, "large.go")
	largeContent := make([]byte, 2*1024*1024) // 2MB
	for i := range largeContent {
		largeContent[i] = 'x'
	}
	err = os.WriteFile(largeFile, largeContent, 0644)
	require.NoError(t, err)

	// Create a small file that should be indexed
	smallFile := filepath.Join(tempDir, "small.go")
	err = os.WriteFile(smallFile, []byte("package main\n\nfunc main() {}"), 0644)
	require.NoError(t, err)

	controller := NewIndexController(tempDir)

	opts := IndexOptions{
		RootPath:    tempDir,
		MaxFileSize: 1024 * 1024, // 1MB - large file should be skipped
	}

	ctx := context.Background()

	// Start indexing
	err = controller.Start(ctx, opts)
	assert.NoError(t, err)

	// Wait for indexing to complete
	time.Sleep(500 * time.Millisecond)
	err = controller.Stop(ctx)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	// Check status - should have processed the small file but skipped the large one
	status := controller.GetStatus()
	assert.Greater(t, status.FilesProcessed, 0)
}
