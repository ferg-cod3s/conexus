package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// Helper function to create a test store
func newTestStore(t *testing.T) *Store {
	store, err := NewStore(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})
	return store
}

func TestNewStore(t *testing.T) {
	t.Run("in-memory database", func(t *testing.T) {
		store, err := NewStore(":memory:")
		require.NoError(t, err)
		defer store.Close()

		assert.NotNil(t, store.db)
	})

	t.Run("file-based database", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test.db"
		store, err := NewStore(tmpFile)
		require.NoError(t, err)
		defer store.Close()

		assert.NotNil(t, store.db)
	})
}

func TestStore_Upsert(t *testing.T) {
	ctx := context.Background()

	t.Run("insert new document", func(t *testing.T) {
		store := newTestStore(t)

		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "This is a test document",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"language": "go",
				"file":     "test.go",
			},
		}

		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Verify document was inserted
		retrieved, err := store.Get(ctx, "doc1")
		require.NoError(t, err)
		assert.Equal(t, doc.ID, retrieved.ID)
		assert.Equal(t, doc.Content, retrieved.Content)
		assert.Equal(t, doc.Vector, retrieved.Vector)
		assert.Equal(t, "go", retrieved.Metadata["language"])
		assert.False(t, retrieved.CreatedAt.IsZero())
		assert.False(t, retrieved.UpdatedAt.IsZero())
	})

	t.Run("update existing document", func(t *testing.T) {
		store := newTestStore(t)

		// Insert initial document
		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "Original content",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Wait a bit to ensure different timestamp
		time.Sleep(1 * time.Second)

		// Update document
		doc.Content = "Updated content"
		doc.Vector = embedding.Vector{0.4, 0.5, 0.6}
		err = store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Verify update
		retrieved, err := store.Get(ctx, "doc1")
		require.NoError(t, err)
		assert.Equal(t, "Updated content", retrieved.Content)
		assert.Equal(t, embedding.Vector{0.4, 0.5, 0.6}, retrieved.Vector)
		assert.True(t, retrieved.UpdatedAt.After(retrieved.CreatedAt))
	})

	t.Run("empty ID error", func(t *testing.T) {
		store := newTestStore(t)

		doc := vectorstore.Document{
			Content: "Test",
			Vector:  embedding.Vector{0.1, 0.2},
		}

		err := store.Upsert(ctx, doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID cannot be empty")
	})

	t.Run("empty vector error", func(t *testing.T) {
		store := newTestStore(t)

		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "Test",
		}

		err := store.Upsert(ctx, doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vector cannot be empty")
	})

	t.Run("nil metadata handled", func(t *testing.T) {
		store := newTestStore(t)

		doc := vectorstore.Document{
			ID:       "doc1",
			Content:  "Test",
			Vector:   embedding.Vector{0.1, 0.2},
			Metadata: nil,
		}

		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "doc1")
		require.NoError(t, err)
		assert.Nil(t, retrieved.Metadata)
	})

	t.Run("context cancellation", func(t *testing.T) {
		store := newTestStore(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "Test",
			Vector:  embedding.Vector{0.1, 0.2},
		}

		err := store.Upsert(ctx, doc)
		assert.Error(t, err)
	})
}

func TestStore_UpsertBatch(t *testing.T) {
	ctx := context.Background()

	t.Run("insert multiple documents", func(t *testing.T) {
		store := newTestStore(t)

		docs := []vectorstore.Document{
			{
				ID:      "doc1",
				Content: "First document",
				Vector:  embedding.Vector{0.1, 0.2, 0.3},
				Metadata: map[string]interface{}{
					"language": "go",
				},
			},
			{
				ID:      "doc2",
				Content: "Second document",
				Vector:  embedding.Vector{0.4, 0.5, 0.6},
				Metadata: map[string]interface{}{
					"language": "python",
				},
			},
			{
				ID:      "doc3",
				Content: "Third document",
				Vector:  embedding.Vector{0.7, 0.8, 0.9},
				Metadata: map[string]interface{}{
					"language": "go",
				},
			},
		}

		err := store.UpsertBatch(ctx, docs)
		require.NoError(t, err)

		// Verify all documents were inserted
		count, err := store.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		// Verify each document
		for _, doc := range docs {
			retrieved, err := store.Get(ctx, doc.ID)
			require.NoError(t, err)
			assert.Equal(t, doc.Content, retrieved.Content)
		}
	})

	t.Run("empty batch", func(t *testing.T) {
		store := newTestStore(t)

		err := store.UpsertBatch(ctx, []vectorstore.Document{})
		assert.NoError(t, err)
	})

	t.Run("batch with invalid document", func(t *testing.T) {
		store := newTestStore(t)

		docs := []vectorstore.Document{
			{
				ID:      "doc1",
				Content: "Valid document",
				Vector:  embedding.Vector{0.1, 0.2},
			},
			{
				ID:      "", // Invalid: empty ID
				Content: "Invalid document",
				Vector:  embedding.Vector{0.3, 0.4},
			},
		}

		err := store.UpsertBatch(ctx, docs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID cannot be empty")

		// First document should not be inserted (transaction rollback)
		_, err = store.Get(ctx, "doc1")
		assert.Error(t, err)
	})

	t.Run("batch update existing documents", func(t *testing.T) {
		store := newTestStore(t)

		// Insert initial documents
		docs := []vectorstore.Document{
			{ID: "doc1", Content: "Original 1", Vector: embedding.Vector{0.1, 0.2}},
			{ID: "doc2", Content: "Original 2", Vector: embedding.Vector{0.3, 0.4}},
		}
		err := store.UpsertBatch(ctx, docs)
		require.NoError(t, err)

		// Update documents
		docs[0].Content = "Updated 1"
		docs[1].Content = "Updated 2"
		err = store.UpsertBatch(ctx, docs)
		require.NoError(t, err)

		// Verify updates
		retrieved1, err := store.Get(ctx, "doc1")
		require.NoError(t, err)
		assert.Equal(t, "Updated 1", retrieved1.Content)

		retrieved2, err := store.Get(ctx, "doc2")
		require.NoError(t, err)
		assert.Equal(t, "Updated 2", retrieved2.Content)
	})
}

func TestStore_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("delete existing document", func(t *testing.T) {
		store := newTestStore(t)

		// Insert document
		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "Test",
			Vector:  embedding.Vector{0.1, 0.2},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Delete document
		err = store.Delete(ctx, "doc1")
		require.NoError(t, err)

		// Verify deletion
		_, err = store.Get(ctx, "doc1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		count, err := store.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete non-existent document", func(t *testing.T) {
		store := newTestStore(t)

		err := store.Delete(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("fts5 trigger removes entry", func(t *testing.T) {
		store := newTestStore(t)

		// Insert document (will trigger FTS5 insert)
		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "Test content for full-text search",
			Vector:  embedding.Vector{0.1, 0.2},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Delete document
		err = store.Delete(ctx, "doc1")
		require.NoError(t, err)

		// Verify FTS5 entry was removed
		var count int
		err = store.db.QueryRow("SELECT COUNT(*) FROM documents_fts WHERE id = ?", "doc1").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestStore_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing document", func(t *testing.T) {
		store := newTestStore(t)

		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "Test content",
			Vector:  embedding.Vector{0.1, 0.2, 0.3},
			Metadata: map[string]interface{}{
				"language": "go",
				"score":    42.5,
			},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		retrieved, err := store.Get(ctx, "doc1")
		require.NoError(t, err)
		assert.Equal(t, doc.ID, retrieved.ID)
		assert.Equal(t, doc.Content, retrieved.Content)
		assert.Equal(t, doc.Vector, retrieved.Vector)
		assert.Equal(t, "go", retrieved.Metadata["language"])
		assert.Equal(t, 42.5, retrieved.Metadata["score"])
	})

	t.Run("get non-existent document", func(t *testing.T) {
		store := newTestStore(t)

		_, err := store.Get(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestStore_Count(t *testing.T) {
	ctx := context.Background()

	t.Run("empty store", func(t *testing.T) {
		store := newTestStore(t)

		count, err := store.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("with documents", func(t *testing.T) {
		store := newTestStore(t)

		// Insert documents
		for i := 0; i < 5; i++ {
			doc := vectorstore.Document{
				ID:      string(rune('a' + i)),
				Content: "Test",
				Vector:  embedding.Vector{float32(i), float32(i)},
			}
			err := store.Upsert(ctx, doc)
			require.NoError(t, err)
		}

		count, err := store.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})
}

func TestStore_Stats(t *testing.T) {
	ctx := context.Background()

	t.Run("empty store", func(t *testing.T) {
		store := newTestStore(t)

		stats, err := store.Stats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), stats.TotalDocuments)
		assert.Equal(t, int64(0), stats.TotalChunks)
		assert.Empty(t, stats.Languages)
		assert.True(t, stats.LastIndexedAt.IsZero())
	})

	t.Run("with documents", func(t *testing.T) {
		store := newTestStore(t)

		// Insert documents with different languages
		docs := []vectorstore.Document{
			{
				ID:      "doc1",
				Content: "Go code",
				Vector:  embedding.Vector{0.1, 0.2},
				Metadata: map[string]interface{}{
					"language": "go",
				},
			},
			{
				ID:      "doc2",
				Content: "Python code",
				Vector:  embedding.Vector{0.3, 0.4},
				Metadata: map[string]interface{}{
					"language": "python",
				},
			},
			{
				ID:      "doc3",
				Content: "More Go code",
				Vector:  embedding.Vector{0.5, 0.6},
				Metadata: map[string]interface{}{
					"language": "go",
				},
			},
		}

		for _, doc := range docs {
			err := store.Upsert(ctx, doc)
			require.NoError(t, err)
		}

		stats, err := store.Stats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), stats.TotalDocuments)
		assert.Equal(t, int64(3), stats.TotalChunks)
		assert.Equal(t, int64(2), stats.Languages["go"])
		assert.Equal(t, int64(1), stats.Languages["python"])
		assert.False(t, stats.LastIndexedAt.IsZero())
	})

	t.Run("documents without language metadata", func(t *testing.T) {
		store := newTestStore(t)

		doc := vectorstore.Document{
			ID:       "doc1",
			Content:  "Test",
			Vector:   embedding.Vector{0.1, 0.2},
			Metadata: map[string]interface{}{}, // No language
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		stats, err := store.Stats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), stats.TotalDocuments)
		assert.Empty(t, stats.Languages)
	})
}

func TestStore_FTS5Integration(t *testing.T) {
	ctx := context.Background()

	t.Run("fts5 insert trigger", func(t *testing.T) {
		store := newTestStore(t)

		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "The quick brown fox jumps over the lazy dog",
			Vector:  embedding.Vector{0.1, 0.2},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Verify FTS5 entry exists
		var count int
		err = store.db.QueryRow("SELECT COUNT(*) FROM documents_fts WHERE id = ?", "doc1").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("fts5 update trigger", func(t *testing.T) {
		store := newTestStore(t)

		// Insert document
		doc := vectorstore.Document{
			ID:      "doc1",
			Content: "Original content",
			Vector:  embedding.Vector{0.1, 0.2},
		}
		err := store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Update document
		doc.Content = "Updated content"
		err = store.Upsert(ctx, doc)
		require.NoError(t, err)

		// Verify FTS5 was updated
		var content string
		err = store.db.QueryRow("SELECT content FROM documents_fts WHERE id = ?", "doc1").Scan(&content)
		require.NoError(t, err)
		assert.Equal(t, "Updated content", content)
	})
}

func TestStore_SearchPlaceholders(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// These are placeholders - will be implemented in Task 6.4.4

	t.Run("SearchHybrid not implemented", func(t *testing.T) {
		_, err := store.SearchHybrid(ctx, "test query", embedding.Vector{0.1, 0.2}, vectorstore.SearchOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})
}

func TestStore_Close(t *testing.T) {
	store, err := NewStore(":memory:")
	require.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)

	// Verify database is closed (operations should fail)
	_, err = store.Count(context.Background())
	assert.Error(t, err)
}

func TestStore_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	// Test concurrent writes
	t.Run("concurrent upserts", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				doc := vectorstore.Document{
					ID:      fmt.Sprintf("doc%d", id),
					Content: "Concurrent test",
					Vector:  embedding.Vector{float32(id), float32(id)},
				}
				err := store.Upsert(ctx, doc)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify all documents were inserted
		count, err := store.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(numGoroutines), count)
	})
}
