package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// TestIndex_WithVectorStore verifies that a full index stores vectors correctly.
func TestIndex_WithVectorStore(t *testing.T) {
	// Setup temporary directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte(`package main

func main() {
	println("Hello, World!")
}
`), 0644))

	// Create indexer with vector store
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	indexer := NewIndexer(statePath)

	embedder := embedding.NewMock(384)
	store := vectorstore.NewMemoryStore()

	// Index with vector store
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
		Embedder:       embedder,
		VectorStore:    store,
	}

	chunks, err := indexer.Index(ctx, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks)

	// Verify vectors were stored
	count, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(len(chunks)), count, "all chunks should be stored in vector store")

	// Verify we can retrieve documents
	for _, chunk := range chunks {
		doc, err := store.Get(ctx, chunk.ID)
		require.NoError(t, err, "should be able to retrieve stored document")
		assert.Equal(t, chunk.Content, doc.Content)
		assert.Equal(t, chunk.FilePath, doc.Metadata["file_path"])
		assert.NotEmpty(t, doc.Vector, "document should have a vector")
	}
}

// TestIndex_WithoutVectorStore verifies that indexing works without vector store.
func TestIndex_WithoutVectorStore(t *testing.T) {
	// Setup temporary directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte(`package main

func main() {
	println("Hello, World!")
}
`), 0644))

	// Create indexer without vector store
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	indexer := NewIndexer(statePath)

	// Index without vector store (nil embedder and store)
	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
		Embedder:       nil,
		VectorStore:    nil,
	}

	chunks, err := indexer.Index(ctx, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks, "indexing should work without vector store")
}

// TestIndexIncremental_VectorStoreUpdates verifies that changed files update vectors.
func TestIndexIncremental_VectorStoreUpdates(t *testing.T) {
	// Setup temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	// Initial content
	initialContent := `package main

func main() {
	println("Version 1")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(initialContent), 0644))

	// Create indexer with vector store
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	indexer := NewIndexer(statePath)

	embedder := embedding.NewMock(384)
	store := vectorstore.NewMemoryStore()

	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
		Embedder:       embedder,
		VectorStore:    store,
	}

	// First index
	chunks1, err := indexer.Index(ctx, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks1)

	// Compute initial state
	state1, err := indexer.merkleTree.Hash(ctx, tmpDir, opts.IgnorePatterns)
	require.NoError(t, err)

	count1, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(len(chunks1)), count1)

	// Store original chunk ID
	originalChunkID := chunks1[0].ID

	// Modify file
	modifiedContent := `package main

func main() {
	println("Version 2 - Updated!")
}
`
	require.NoError(t, os.WriteFile(testFile, []byte(modifiedContent), 0644))

	// Incremental index
	chunks2, _, err := indexer.IndexIncremental(ctx, opts, state1)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks2, "should return updated chunks")

	// Vector store should still have same number of documents
	count2, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, count1, count2, "count should remain the same after update")

	// New chunk should have different content
	newChunk := chunks2[0]
	doc, err := store.Get(ctx, newChunk.ID)
	require.NoError(t, err)
	assert.Contains(t, doc.Content, "Version 2 - Updated!")

	// Old chunk should not be retrievable (if ID changed due to content hash)
	if originalChunkID != newChunk.ID {
		_, err = store.Get(ctx, originalChunkID)
		assert.Error(t, err, "old chunk should be deleted")
	}
}

// TestIndexIncremental_VectorStoreDeletions verifies that deleted files remove vectors.
func TestIndexIncremental_VectorStoreDeletions(t *testing.T) {
	// Setup temporary directory with two files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "keep.go")
	file2 := filepath.Join(tmpDir, "delete.go")

	require.NoError(t, os.WriteFile(file1, []byte(`package main
func keep() {}
`), 0644))
	require.NoError(t, os.WriteFile(file2, []byte(`package main
func delete() {}
`), 0644))

	// Create indexer with vector store
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	indexer := NewIndexer(statePath)

	embedder := embedding.NewMock(384)
	store := vectorstore.NewMemoryStore()

	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
		Embedder:       embedder,
		VectorStore:    store,
	}

	// First index
	chunks1, err := indexer.Index(ctx, opts)
	require.NoError(t, err)
	assert.Len(t, chunks1, 2, "should have 2 chunks initially")

	// Compute initial state
	state1, err := indexer.merkleTree.Hash(ctx, tmpDir, opts.IgnorePatterns)
	require.NoError(t, err)

	count1, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count1)

	// Find chunk IDs from both files
	var file1ChunkID, file2ChunkID string
	for _, chunk := range chunks1 {
		switch filepath.Base(chunk.FilePath) {
		case "delete.go":
			file2ChunkID = chunk.ID
		case "keep.go":
			file1ChunkID = chunk.ID
		}
	}
	require.NotEmpty(t, file1ChunkID, "should find chunk from keep.go")
	require.NotEmpty(t, file2ChunkID, "should find chunk from delete.go")

	// Delete file2
	require.NoError(t, os.Remove(file2))

	// Incremental index
	// Note: chunks2 will be empty because only deletions occurred, no modifications
	chunks2, _, err := indexer.IndexIncremental(ctx, opts, state1)
	require.NoError(t, err)

	// Should only have 1 chunk now (from file1)
	count2, err := store.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count2, "should have 1 chunk after deletion")

	// Deleted file's chunk should not be retrievable
	_, err = store.Get(ctx, file2ChunkID)
	assert.Error(t, err, "deleted file's chunk should not exist")

	// Remaining file (keep.go) should still be in vector store
	// We check the store directly, not chunks2, since chunks2 only contains modified files
	_, err = store.Get(ctx, file1ChunkID)
	require.NoError(t, err, "keep.go chunk should still exist in vector store")

	// chunks2 should be empty since no files were modified, only deleted
	assert.Empty(t, chunks2, "no chunks should be returned when only files are deleted")
}

// TestIndex_OnlyEmbedderProvided verifies that both embedder AND store are required.
func TestIndex_OnlyEmbedderProvided(t *testing.T) {
	// Setup temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte(`package main
func test() {}
`), 0644))

	// Create indexer with embedder but no store
	statePath := filepath.Join(tmpDir, ".conexus", "state.json")
	indexer := NewIndexer(statePath)

	embedder := embedding.NewMock(384)

	ctx := context.Background()
	opts := IndexOptions{
		RootPath:       tmpDir,
		IgnorePatterns: []string{".conexus/"},
		MaxFileSize:    1024 * 1024,
		Embedder:       embedder,
		VectorStore:    nil, // No store!
	}

	// Index should succeed but not store vectors (embedder check short-circuits)
	chunks, err := indexer.Index(ctx, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks, "indexing should succeed without vector store")
}
