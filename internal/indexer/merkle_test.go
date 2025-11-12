package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMerkleTree(t *testing.T) {
	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	assert.NotNil(t, mt)
}

func TestMerkleTree_Hash(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		files   map[string]string
		wantErr bool
	}{
		{
			name: "single file",
			files: map[string]string{
				"file.txt": "content",
			},
		},
		{
			name: "multiple files",
			files: map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
				"file3.txt": "content3",
			},
		},
		{
			name: "nested directories",
			files: map[string]string{
				"dir1/file1.txt":      "content1",
				"dir1/dir2/file2.txt": "content2",
				"dir3/file3.txt":      "content3",
			},
		},
		{
			name: "empty directory",
			files: map[string]string{
				"dir1/.gitkeep": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestFiles(t, tt.files)

			walker := NewFileWalker(0)
			mt := NewMerkleTree(walker)

			state, err := mt.Hash(ctx, tmpDir, nil)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, state)

			// Verify it's valid JSON
			var treeState treeState
			err = json.Unmarshal(state, &treeState)
			require.NoError(t, err)
			assert.NotNil(t, treeState.Root)
			assert.NotEmpty(t, treeState.Root.Hash)
		})
	}
}

func TestMerkleTree_Hash_Deterministic(t *testing.T) {
	ctx := context.Background()
	tmpDir := createTestFiles(t, map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	})

	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	// Hash multiple times
	hashes := make([]string, 3)
	for i := 0; i < 3; i++ {
		state, err := mt.Hash(ctx, tmpDir, nil)
		require.NoError(t, err)

		var ts treeState
		require.NoError(t, json.Unmarshal(state, &ts))
		hashes[i] = ts.Root.Hash
	}

	// All hashes should be identical
	assert.Equal(t, hashes[0], hashes[1])
	assert.Equal(t, hashes[1], hashes[2])
}

func TestMerkleTree_Hash_SameContent(t *testing.T) {
	ctx := context.Background()

	// Create two directories with identical content
	tmpDir1 := createTestFiles(t, map[string]string{
		"file.txt": "test content",
	})
	tmpDir2 := createTestFiles(t, map[string]string{
		"file.txt": "test content",
	})

	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	state1, err := mt.Hash(ctx, tmpDir1, nil)
	require.NoError(t, err)

	state2, err := mt.Hash(ctx, tmpDir2, nil)
	require.NoError(t, err)

	// Extract root hashes
	var ts1, ts2 treeState
	require.NoError(t, json.Unmarshal(state1, &ts1))
	require.NoError(t, json.Unmarshal(state2, &ts2))

	assert.Equal(t, ts1.Root.Hash, ts2.Root.Hash)
}

func TestMerkleTree_Hash_DifferentContent(t *testing.T) {
	ctx := context.Background()

	tmpDir1 := createTestFiles(t, map[string]string{
		"file.txt": "content1",
	})
	tmpDir2 := createTestFiles(t, map[string]string{
		"file.txt": "content2",
	})

	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	state1, err := mt.Hash(ctx, tmpDir1, nil)
	require.NoError(t, err)

	state2, err := mt.Hash(ctx, tmpDir2, nil)
	require.NoError(t, err)

	var ts1, ts2 treeState
	require.NoError(t, json.Unmarshal(state1, &ts1))
	require.NoError(t, json.Unmarshal(state2, &ts2))

	assert.NotEqual(t, ts1.Root.Hash, ts2.Root.Hash)
}

func TestMerkleTree_Hash_WithIgnorePatterns(t *testing.T) {
	ctx := context.Background()

	tmpDir := createTestFiles(t, map[string]string{
		"include.txt":         "included",
		"node_modules/pkg.js": "excluded",
		".git/config":         "excluded",
	})

	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	ignorePatterns := []string{"node_modules/", ".git/"}
	state, err := mt.Hash(ctx, tmpDir, ignorePatterns)
	require.NoError(t, err)

	var ts treeState
	require.NoError(t, json.Unmarshal(state, &ts))

	// Verify ignored paths are not in tree
	assert.False(t, hasPath(ts.Root, "node_modules"))
	assert.False(t, hasPath(ts.Root, ".git"))
	assert.True(t, hasPath(ts.Root, "include.txt"))
}

func TestMerkleTree_Hash_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("nil walker", func(t *testing.T) {
		mt := NewMerkleTree(nil)
		_, err := mt.Hash(ctx, "/tmp", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "walker cannot be nil")
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)
		_, err := mt.Hash(ctx, "/nonexistent/path/12345", nil)
		assert.Error(t, err)
	})
}

func TestMerkleTree_Diff(t *testing.T) {
	ctx := context.Background()

	t.Run("no changes", func(t *testing.T) {
		tmpDir := createTestFiles(t, map[string]string{
			"file.txt": "content",
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		state1, err := mt.Hash(ctx, tmpDir, nil)
		require.NoError(t, err)

		state2, err := mt.Hash(ctx, tmpDir, nil)
		require.NoError(t, err)

		changes, err := mt.Diff(ctx, state1, state2)
		require.NoError(t, err)
		assert.Empty(t, changes)
	})

	t.Run("file added", func(t *testing.T) {
		tmpDir1 := createTestFiles(t, map[string]string{
			"file1.txt": "content1",
		})
		tmpDir2 := createTestFiles(t, map[string]string{
			"file1.txt": "content1",
			"file2.txt": "content2", // New file
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		state1, err := mt.Hash(ctx, tmpDir1, nil)
		require.NoError(t, err)

		state2, err := mt.Hash(ctx, tmpDir2, nil)
		require.NoError(t, err)

		changes, err := mt.Diff(ctx, state1, state2)
		require.NoError(t, err)
		assert.Contains(t, changes, "file2.txt")
	})

	t.Run("file deleted", func(t *testing.T) {
		tmpDir1 := createTestFiles(t, map[string]string{
			"file1.txt": "content1",
			"file2.txt": "content2",
		})
		tmpDir2 := createTestFiles(t, map[string]string{
			"file1.txt": "content1",
			// file2.txt deleted
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		state1, err := mt.Hash(ctx, tmpDir1, nil)
		require.NoError(t, err)

		state2, err := mt.Hash(ctx, tmpDir2, nil)
		require.NoError(t, err)

		changes, err := mt.Diff(ctx, state1, state2)
		require.NoError(t, err)
		assert.Contains(t, changes, "file2.txt")
	})

	t.Run("file modified", func(t *testing.T) {
		tmpDir1 := createTestFiles(t, map[string]string{
			"file.txt": "old content",
		})
		tmpDir2 := createTestFiles(t, map[string]string{
			"file.txt": "new content", // Modified
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		state1, err := mt.Hash(ctx, tmpDir1, nil)
		require.NoError(t, err)

		state2, err := mt.Hash(ctx, tmpDir2, nil)
		require.NoError(t, err)

		changes, err := mt.Diff(ctx, state1, state2)
		require.NoError(t, err)
		assert.Contains(t, changes, "file.txt")
	})

	t.Run("multiple changes", func(t *testing.T) {
		tmpDir1 := createTestFiles(t, map[string]string{
			"file1.txt": "content1",
			"file2.txt": "old content",
		})
		tmpDir2 := createTestFiles(t, map[string]string{
			"file2.txt": "new content", // Modified
			"file3.txt": "content3",    // Added
			// file1.txt deleted
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		state1, err := mt.Hash(ctx, tmpDir1, nil)
		require.NoError(t, err)

		state2, err := mt.Hash(ctx, tmpDir2, nil)
		require.NoError(t, err)

		changes, err := mt.Diff(ctx, state1, state2)
		require.NoError(t, err)

		assert.Len(t, changes, 3)
		assert.Contains(t, changes, "file1.txt") // Deleted
		assert.Contains(t, changes, "file2.txt") // Modified
		assert.Contains(t, changes, "file3.txt") // Added
	})

	t.Run("nested directory changes", func(t *testing.T) {
		tmpDir1 := createTestFiles(t, map[string]string{
			"dir1/file1.txt": "old content",
		})
		tmpDir2 := createTestFiles(t, map[string]string{
			"dir1/file1.txt":      "new content", // Modified
			"dir1/dir2/file2.txt": "content2",    // Added
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		state1, err := mt.Hash(ctx, tmpDir1, nil)
		require.NoError(t, err)

		state2, err := mt.Hash(ctx, tmpDir2, nil)
		require.NoError(t, err)

		changes, err := mt.Diff(ctx, state1, state2)
		require.NoError(t, err)

		assert.Contains(t, changes, "dir1/file1.txt")
		assert.Contains(t, changes, "dir1/dir2/file2.txt")
	})
}

func TestMerkleTree_Diff_Errors(t *testing.T) {
	ctx := context.Background()
	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	t.Run("empty old state", func(t *testing.T) {
		_, err := mt.Diff(ctx, []byte{}, []byte("{}"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be non-empty")
	})

	t.Run("empty new state", func(t *testing.T) {
		_, err := mt.Diff(ctx, []byte("{}"), []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be non-empty")
	})

	t.Run("invalid old state JSON", func(t *testing.T) {
		_, err := mt.Diff(ctx, []byte("invalid"), []byte("{}"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to deserialize old state")
	})

	t.Run("invalid new state JSON", func(t *testing.T) {
		_, err := mt.Diff(ctx, []byte("{}"), []byte("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to deserialize new state")
	})
}

func TestComputeFileHash(t *testing.T) {
	t.Run("hash file content", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.txt")
		content := "test content"
		require.NoError(t, os.WriteFile(tmpFile, []byte(content), 0644))

		hash, err := computeFileHash(tmpFile, tmpDir)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		// Verify it's a valid SHA256 hash
		assert.Len(t, hash, 64)

		// Compute expected hash
		h := sha256.New()
		h.Write([]byte(content))
		expectedHash := hex.EncodeToString(h.Sum(nil))

		assert.Equal(t, expectedHash, hash)
	})

	t.Run("same content produces same hash", func(t *testing.T) {
		content := "test content"

		tmpDir1 := t.TempDir()
		tmpFile1 := filepath.Join(tmpDir1, "file1.txt")
		require.NoError(t, os.WriteFile(tmpFile1, []byte(content), 0644))

		tmpDir2 := t.TempDir()
		tmpFile2 := filepath.Join(tmpDir2, "file2.txt")
		require.NoError(t, os.WriteFile(tmpFile2, []byte(content), 0644))

		hash1, err := computeFileHash(tmpFile1, tmpDir1)
		require.NoError(t, err)

		hash2, err := computeFileHash(tmpFile2, tmpDir2)
		require.NoError(t, err)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("different content produces different hash", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpFile1 := filepath.Join(tmpDir1, "file1.txt")
		require.NoError(t, os.WriteFile(tmpFile1, []byte("content1"), 0644))

		tmpDir2 := t.TempDir()
		tmpFile2 := filepath.Join(tmpDir2, "file2.txt")
		require.NoError(t, os.WriteFile(tmpFile2, []byte("content2"), 0644))

		hash1, err := computeFileHash(tmpFile1, tmpDir1)
		require.NoError(t, err)

		hash2, err := computeFileHash(tmpFile2, tmpDir2)
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := computeFileHash("/nonexistent/file.txt", "/")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open file")
	})
}

func TestMerkleTree_DirectoryHashing(t *testing.T) {
	ctx := context.Background()

	t.Run("directory hash changes when child added", func(t *testing.T) {
		tmpDir1 := createTestFiles(t, map[string]string{
			"dir/file1.txt": "content",
		})
		tmpDir2 := createTestFiles(t, map[string]string{
			"dir/file1.txt": "content",
			"dir/file2.txt": "content",
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		state1, err := mt.Hash(ctx, tmpDir1, nil)
		require.NoError(t, err)

		state2, err := mt.Hash(ctx, tmpDir2, nil)
		require.NoError(t, err)

		var ts1, ts2 treeState
		require.NoError(t, json.Unmarshal(state1, &ts1))
		require.NoError(t, json.Unmarshal(state2, &ts2))

		// Root hashes should differ
		assert.NotEqual(t, ts1.Root.Hash, ts2.Root.Hash)
	})

	t.Run("directory hash is deterministic", func(t *testing.T) {
		tmpDir := createTestFiles(t, map[string]string{
			"dir/b.txt": "b",
			"dir/a.txt": "a",
			"dir/c.txt": "c",
		})

		walker := NewFileWalker(0)
		mt := NewMerkleTree(walker)

		// Hash multiple times
		hashes := make([]string, 3)
		for i := 0; i < 3; i++ {
			state, err := mt.Hash(ctx, tmpDir, nil)
			require.NoError(t, err)

			var ts treeState
			require.NoError(t, json.Unmarshal(state, &ts))
			hashes[i] = ts.Root.Hash
		}

		assert.Equal(t, hashes[0], hashes[1])
		assert.Equal(t, hashes[1], hashes[2])
	})
}

// Helper functions

func createTestFiles(t *testing.T, files map[string]string) string {
	tmpDir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}
	return tmpDir
}

func hasPath(root *treeNode, path string) bool {
	if root == nil {
		return false
	}

	if path == "" {
		return true
	}

	parts := filepath.SplitList(path)
	if len(parts) <= 1 {
		// Manual split
		parts = []string{}
		current := ""
		for _, r := range path {
			if r == '/' || r == filepath.Separator {
				if current != "" {
					parts = append(parts, current)
					current = ""
				}
			} else {
				current += string(r)
			}
		}
		if current != "" {
			parts = append(parts, current)
		}
	}

	current := root
	for _, part := range parts {
		if current.Children == nil {
			return false
		}
		child, exists := current.Children[part]
		if !exists {
			return false
		}
		current = child
	}

	return true
}

// Benchmarks

func BenchmarkMerkleTree_Hash(b *testing.B) {
	ctx := context.Background()
	tmpDir := b.TempDir()

	// Create a reasonable-sized directory structure
	for i := 0; i < 10; i++ {
		dir := filepath.Join(tmpDir, fmt.Sprintf("dir%d", i))
		require.NoError(b, os.MkdirAll(dir, 0755))
		for j := 0; j < 10; j++ {
			file := filepath.Join(dir, fmt.Sprintf("file%d.txt", j))
			require.NoError(b, os.WriteFile(file, []byte("content"), 0644))
		}
	}

	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mt.Hash(ctx, tmpDir, nil)
	}
}

func BenchmarkMerkleTree_Diff(b *testing.B) {
	ctx := context.Background()
	tmpDir1 := b.TempDir()
	tmpDir2 := b.TempDir()

	// Create directory structures
	for i := 0; i < 10; i++ {
		dir1 := filepath.Join(tmpDir1, fmt.Sprintf("dir%d", i))
		dir2 := filepath.Join(tmpDir2, fmt.Sprintf("dir%d", i))
		require.NoError(b, os.MkdirAll(dir1, 0755))
		require.NoError(b, os.MkdirAll(dir2, 0755))
		for j := 0; j < 10; j++ {
			file1 := filepath.Join(dir1, fmt.Sprintf("file%d.txt", j))
			file2 := filepath.Join(dir2, fmt.Sprintf("file%d.txt", j))
			require.NoError(b, os.WriteFile(file1, []byte("content"), 0644))
			require.NoError(b, os.WriteFile(file2, []byte("content"), 0644))
		}
	}

	// Modify a few files
	require.NoError(b, os.WriteFile(filepath.Join(tmpDir2, "dir5/file5.txt"), []byte("modified"), 0644))

	walker := NewFileWalker(0)
	mt := NewMerkleTree(walker)

	state1, err := mt.Hash(ctx, tmpDir1, nil)
	require.NoError(b, err)

	state2, err := mt.Hash(ctx, tmpDir2, nil)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mt.Diff(ctx, state1, state2)
	}
}

func BenchmarkComputeFileHash(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := make([]byte, 1024*1024) // 1MB file
	for i := range content {
		content[i] = byte(i % 256)
	}
	require.NoError(b, os.WriteFile(tmpFile, content, 0644))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = computeFileHash(tmpFile, tmpDir)
	}
}
