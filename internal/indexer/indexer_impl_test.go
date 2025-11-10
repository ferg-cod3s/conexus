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

func TestNewIndexer(t *testing.T) {
	idx := NewIndexer("/tmp/test-state.json")

	assert.NotNil(t, idx)
	assert.NotNil(t, idx.walker)
	assert.NotNil(t, idx.merkleTree)
	assert.Equal(t, "/tmp/test-state.json", idx.statePath)
	assert.Len(t, idx.chunkers, 1) // Should have the code chunker
}

func TestIndexFullScan(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := map[string]string{
		"main.go":       "package main\n\nfunc main() {}\n",
		"helper.go":     "package main\n\nfunc helper() {}\n",
		"README.md":     "# Test Project\n",
		"config.json":   `{"key": "value"}`,
		"nested/sub.go": "package nested\n",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	// Create indexer
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	idx := NewIndexer(statePath)

	// Index the directory
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{},
		MaxFileSize:    1024 * 1024,
	}

	chunks, err := idx.Index(ctx, opts)
	require.NoError(t, err)

	// Verify chunks
	assert.Len(t, chunks, 5, "should create 5 chunks for 5 files")

	// Verify chunk contents
	foundFiles := make(map[string]bool)
	for _, chunk := range chunks {
		foundFiles[chunk.FilePath] = true
		assert.NotEmpty(t, chunk.ID)
		assert.NotEmpty(t, chunk.Content)
		assert.NotEmpty(t, chunk.Hash)
		assert.NotEmpty(t, chunk.Language)
		assert.Greater(t, chunk.EndLine, 0)
	}

	assert.True(t, foundFiles["main.go"])
	assert.True(t, foundFiles["helper.go"])
	assert.True(t, foundFiles["README.md"])
	assert.True(t, foundFiles["config.json"])
	assert.True(t, foundFiles[filepath.Join("nested", "sub.go")])
}

func TestIndexWithIgnorePatterns(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := []string{
		"main.go",
		"main_test.go",
		"vendor/lib.go",
		".git/config",
	}

	for _, file := range files {
		fullPath := filepath.Join(tmpDir, file)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte("content"), 0644))
	}

	// Index with ignore patterns
	idx := NewIndexer(filepath.Join(tmpDir, "state.json"))
	ctx := context.Background()
	opts := IndexOptions{
		RootPath: tmpDir,
		IgnorePatterns: []string{
			"*_test.go",
			"vendor/",
			".git/",
		},
		MaxFileSize: 1024 * 1024,
	}

	chunks, err := idx.Index(ctx, opts)
	require.NoError(t, err)

	// Should only index main.go
	assert.Len(t, chunks, 1)
	assert.Equal(t, "main.go", chunks[0].FilePath)
}

func TestIndexIncremental_FirstRun(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package test"), 0644))

	// First incremental run (no previous state)
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	idx := NewIndexer(statePath)
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
	}

	chunks, newState, err := idx.IndexIncremental(ctx, opts, nil)
	require.NoError(t, err)

	// Should perform full index
	assert.Len(t, chunks, 1)
	assert.NotNil(t, newState)
	assert.Greater(t, len(newState), 0)
}

func TestIndexIncremental_NoChanges(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package test"), 0644))

	// First run
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	idx := NewIndexer(statePath)
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
	}

	_, state1, err := idx.IndexIncremental(ctx, opts, nil)
	require.NoError(t, err)

	// Second run with no changes
	chunks, state2, err := idx.IndexIncremental(ctx, opts, state1)
	require.NoError(t, err)

	// Should return empty chunks
	assert.Empty(t, chunks, "no changes should result in empty chunks")
	assert.NotNil(t, state2)
}

func TestIndexIncremental_WithChanges(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create initial files
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	require.NoError(t, os.WriteFile(file1, []byte("package test1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("package test2"), 0644))

	// First run
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	idx := NewIndexer(statePath)
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
	}

	_, state1, err := idx.IndexIncremental(ctx, opts, nil)
	require.NoError(t, err)

	// Modify one file
	time.Sleep(10 * time.Millisecond) // Ensure timestamp changes
	require.NoError(t, os.WriteFile(file1, []byte("package test1\n// modified"), 0644))

	// Second run with changes
	chunks, state2, err := idx.IndexIncremental(ctx, opts, state1)
	require.NoError(t, err)

	// Should only index the modified file
	assert.Len(t, chunks, 1, "should only reindex changed file")
	assert.Equal(t, "file1.go", chunks[0].FilePath)
	assert.NotEqual(t, state1, state2)
}

func TestIndexIncremental_DeletedFile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create initial files
	file1 := filepath.Join(tmpDir, "file1.go")
	file2 := filepath.Join(tmpDir, "file2.go")
	require.NoError(t, os.WriteFile(file1, []byte("package test1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("package test2"), 0644))

	// First run
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	idx := NewIndexer(statePath)
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
	}

	_, state1, err := idx.IndexIncremental(ctx, opts, nil)
	require.NoError(t, err)

	// Delete one file
	require.NoError(t, os.Remove(file1))

	// Second run after deletion
	chunks, _, err := idx.IndexIncremental(ctx, opts, state1)
	require.NoError(t, err)

	// Should not include deleted file
	for _, chunk := range chunks {
		assert.NotEqual(t, "file1.go", chunk.FilePath)
	}
}

func TestSaveAndLoadState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	idx := NewIndexer(statePath)
	ctx := context.Background()

	// Save state
	testState := []byte(`{"test": "data"}`)
	err = idx.SaveState(ctx, testState)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(statePath)
	require.NoError(t, err)

	// Load state
	loadedState, err := idx.LoadState(ctx)
	require.NoError(t, err)
	assert.Equal(t, testState, loadedState)
}

func TestLoadState_NotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, "nonexistent.json")
	idx := NewIndexer(statePath)
	ctx := context.Background()

	// Load non-existent state
	state, err := idx.LoadState(ctx)
	require.NoError(t, err)
	assert.Nil(t, state)
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "go"},
		{"script.js", "javascript"},
		{"component.jsx", "javascript"},
		{"module.ts", "typescript"},
		{"app.tsx", "typescript"},
		{"script.py", "python"},
		{"lib.rs", "rust"},
		{"Main.java", "java"},
		{"program.cpp", "cpp"},
		{"code.c", "c"},
		{"README.md", "markdown"},
		{"notes.txt", "text"},
		{"config.yaml", "yaml"},
		{"data.json", "json"},
		{"unknown.xyz", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectLanguage(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"single line", 1},
		{"line1\nline2", 2},
		{"line1\nline2\nline3", 3},
		{"line1\nline2\n", 3},
		{"\n\n\n", 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := countLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStateManager_SaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, "state.json")
	sm := NewStateManager(statePath)
	ctx := context.Background()

	// Save
	testData := []byte(`{"version": "1.0"}`)
	err = sm.Save(ctx, testData)
	require.NoError(t, err)

	// Load
	loaded, err := sm.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, testData, loaded)

	// Exists
	assert.True(t, sm.Exists())
}

func TestStateManager_AtomicWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, "state.json")
	sm := NewStateManager(statePath)
	ctx := context.Background()

	// Initial save
	require.NoError(t, sm.Save(ctx, []byte("v1")))

	// Second save should atomically replace
	require.NoError(t, sm.Save(ctx, []byte("v2")))

	// Verify no .tmp file remains
	tmpPath := statePath + ".tmp"
	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err))

	// Verify final content
	loaded, err := sm.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("v2"), loaded)
}

func TestStateManager_Clear(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "state-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	statePath := filepath.Join(tmpDir, "state.json")
	sm := NewStateManager(statePath)
	ctx := context.Background()

	// Save and verify
	require.NoError(t, sm.Save(ctx, []byte("data")))
	assert.True(t, sm.Exists())

	// Clear
	require.NoError(t, sm.Clear(ctx))
	assert.False(t, sm.Exists())

	// Clear non-existent should not error
	require.NoError(t, sm.Clear(ctx))
}

func TestIndexStats_Duration(t *testing.T) {
	stats := IndexStats{
		StartTime: time.Now().Add(-5 * time.Second),
		EndTime:   time.Now(),
	}

	duration := stats.Duration()
	assert.GreaterOrEqual(t, duration, 4*time.Second)
	assert.LessOrEqual(t, duration, 6*time.Second)
}

func TestIndexStats_ToMetrics(t *testing.T) {
	stats := IndexStats{
		StartTime:      time.Now().Add(-1 * time.Second),
		EndTime:        time.Now(),
		TotalFiles:     100,
		IndexedFiles:   50,
		SkippedFiles:   50,
		TotalChunks:    200,
		BytesProcessed: 1024 * 1024,
	}

	metrics := stats.ToMetrics()
	assert.Equal(t, 100, metrics.TotalFiles)
	assert.Equal(t, 50, metrics.IndexedFiles)
	assert.Equal(t, 50, metrics.SkippedFiles)
	assert.Equal(t, 200, metrics.TotalChunks)
	assert.Equal(t, int64(1024*1024), metrics.BytesProcessed)
	assert.Greater(t, metrics.Duration, time.Duration(0))
}

func TestMarshalUnmarshalState(t *testing.T) {
	state := []byte(`{"nodes": [{"path": "test"}]}`)
	info := StateInfo{
		Version:   "1.0",
		RootPath:  "/test/path",
		Timestamp: time.Now().Round(time.Second),
		FileCount: 42,
		StateHash: "abc123",
	}

	// Marshal
	marshaled, err := MarshalState(state, info)
	require.NoError(t, err)
	assert.NotEmpty(t, marshaled)

	// Unmarshal
	unmarshaledState, unmarshaledInfo, err := UnmarshalState(marshaled)
	require.NoError(t, err)
	assert.Equal(t, state, unmarshaledState)
	assert.Equal(t, info.Version, unmarshaledInfo.Version)
	assert.Equal(t, info.RootPath, unmarshaledInfo.RootPath)
	assert.Equal(t, info.FileCount, unmarshaledInfo.FileCount)
	assert.Equal(t, info.StateHash, unmarshaledInfo.StateHash)
	assert.True(t, info.Timestamp.Equal(unmarshaledInfo.Timestamp))
}

func TestIndexWithMaxFileSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create small and large files
	smallFile := filepath.Join(tmpDir, "small.txt")
	largeFile := filepath.Join(tmpDir, "large.txt")

	require.NoError(t, os.WriteFile(smallFile, []byte("small"), 0644))
	require.NoError(t, os.WriteFile(largeFile, make([]byte, 2048), 0644))

	// Index with max file size of 1KB
	idx := NewIndexer(filepath.Join(tmpDir, "state.json"))
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{},
		MaxFileSize:    1024, // 1KB
	}

	chunks, err := idx.Index(ctx, opts)
	require.NoError(t, err)

	// Should only index small file
	assert.Len(t, chunks, 1)
	assert.Equal(t, "small.txt", chunks[0].FilePath)
}

func TestCreateSingleChunk(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "indexer-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	content := "package test\n\nfunc test() {}\n"
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	info, err := os.Stat(testFile)
	require.NoError(t, err)

	idx := NewIndexer(filepath.Join(tmpDir, "state.json"))
	chunk := idx.createSingleChunk(content, "test.go", info)

	assert.NotEmpty(t, chunk.ID)
	assert.Equal(t, content, chunk.Content)
	assert.Equal(t, "test.go", chunk.FilePath)
	assert.Equal(t, "go", chunk.Language)
	assert.Equal(t, ChunkTypeUnknown, chunk.Type)
	assert.Equal(t, 1, chunk.StartLine)
	assert.Equal(t, 4, chunk.EndLine)
	assert.NotEmpty(t, chunk.Hash)
}

// Benchmark tests
func BenchmarkIndexFullScan(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	// Create 100 test files
	for i := 0; i < 100; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i))
		content := "package test\n\nfunc test() {}\n"
		require.NoError(b, os.WriteFile(path, []byte(content), 0644))
	}

	idx := NewIndexer(filepath.Join(tmpDir, "state.json"))
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{},
		MaxFileSize:    1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := idx.Index(ctx, opts)
		require.NoError(b, err)
	}
}

func BenchmarkIndexIncremental_NoChanges(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tmpDir)

	// Create test files
	for i := 0; i < 50; i++ {
		path := filepath.Join(tmpDir, fmt.Sprintf("file%d.go", i))
		require.NoError(b, os.WriteFile(path, []byte("content"), 0644))
	}

	idx := NewIndexer(filepath.Join(tmpDir, "state.json"))
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{},
		MaxFileSize:    1024 * 1024,
	}

	// Initial state
	_, state, err := idx.IndexIncremental(ctx, opts, nil)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := idx.IndexIncremental(ctx, opts, state)
		require.NoError(b, err)
	}
}

func TestReindexPaths_PathValidation(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("package main"), 0644))

	ctx := context.Background()

	opts := IndexOptions{
		RootPath:    tmpDir,
		MaxFileSize: 1024,
	}

	tests := []struct {
		name    string
		paths   []string
		wantErr bool
	}{
		{
			name:    "valid file path",
			paths:   []string{testFile},
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			paths:   []string{"../../../etc/passwd"},
			wantErr: false, // Should skip invalid paths, not error
		},
		{
			name:    "non-existent path",
			paths:   []string{filepath.Join(tmpDir, "nonexistent.go")},
			wantErr: false, // Should skip non-existent paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh indexer for each test case to avoid "indexing already running" race condition
			testStatePath := filepath.Join(tmpDir, fmt.Sprintf("state_%s.json", tt.name))
			idx := NewIndexer(testStatePath)

			err := idx.ReindexPaths(ctx, opts, tt.paths)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
