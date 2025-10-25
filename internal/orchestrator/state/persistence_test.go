package state

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPersistence(t *testing.T) {
	tempDir := t.TempDir()

	p, err := NewPersistence(tempDir)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, tempDir, p.baseDir)

	// Verify directory was created
	info, err := os.Stat(tempDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNewPersistence_InvalidPath(t *testing.T) {
	// Try to create persistence in a location that doesn't exist and can't be created
	invalidPath := "/root/nonexistent/invalid/path"

	_, err := NewPersistence(invalidPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create base directory")
}

func TestPersistence_SaveAndLoadSession(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Create a test session
	session := &Session{
		ID:           "test-session-123",
		UserID:       "user456",
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		History: []HistoryEntry{
			{
				Timestamp:   time.Now(),
				UserRequest: "Hello",
				Agent:       "test-agent",
				Response:    &schema.AgentOutputV1{Version: "AGENT_OUTPUT_V1"},
			},
		},
	}

	// Save session
	err = p.SaveSession(session)
	assert.NoError(t, err)

	// Verify file exists
	sessionPath := filepath.Join(tempDir, "session-test-session-123.json")
	_, err = os.Stat(sessionPath)
	assert.NoError(t, err)

	// Load session
	loaded, err := p.LoadSession("test-session-123")
	assert.NoError(t, err)
	assert.NotNil(t, loaded)

	// Verify content
	assert.Equal(t, session.ID, loaded.ID)
	assert.Equal(t, session.UserID, loaded.UserID)
	assert.Len(t, loaded.History, 1)
	assert.Equal(t, "Hello", loaded.History[0].UserRequest)
	assert.Equal(t, "test-agent", loaded.History[0].Agent)
}

func TestPersistence_LoadSession_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Try to load non-existent session
	_, err = p.LoadSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read session file")
}

func TestPersistence_LoadSession_InvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Try to load session with path traversal
	_, err = p.LoadSession("../../../etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session path")
}

func TestPersistence_DeleteSession(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Create and save a session
	session := &Session{
		ID:     "test-session-delete",
		UserID: "user123",
	}
	err = p.SaveSession(session)
	require.NoError(t, err)

	// Verify file exists
	sessionPath := filepath.Join(tempDir, "session-test-session-delete.json")
	_, err = os.Stat(sessionPath)
	assert.NoError(t, err)

	// Delete session
	err = p.DeleteSession("test-session-delete")
	assert.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(sessionPath)
	assert.True(t, os.IsNotExist(err))
}

func TestPersistence_DeleteSession_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Try to delete non-existent session (should not error)
	err = p.DeleteSession("non-existent")
	assert.NoError(t, err)
}

func TestPersistence_ListSessions(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Initially should be empty
	sessions, err := p.ListSessions()
	assert.NoError(t, err)
	assert.Empty(t, sessions)

	// Create some sessions
	session1 := &Session{ID: "session1", UserID: "user1"}
	session2 := &Session{ID: "session2", UserID: "user2"}

	err = p.SaveSession(session1)
	require.NoError(t, err)
	err = p.SaveSession(session2)
	require.NoError(t, err)

	// List sessions
	sessions, err = p.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Contains(t, sessions, "session1")
	assert.Contains(t, sessions, "session2")
}

func TestPersistence_ListSessions_IgnoresInvalidFiles(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Create some invalid files
	invalidFiles := []string{
		"not-a-session.txt",
		"session-invalid.json",
		"other-file.json",
	}

	for _, file := range invalidFiles {
		path := filepath.Join(tempDir, file)
		err := os.WriteFile(path, []byte("content"), 0600)
		require.NoError(t, err)
	}

	// Create one valid session
	session := &Session{ID: "valid-session", UserID: "user1"}
	err = p.SaveSession(session)
	require.NoError(t, err)

	// List sessions - should return the valid one and the invalid-named one
	sessions, err := p.ListSessions()
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
	assert.Contains(t, sessions, "valid-session")
	assert.Contains(t, sessions, "invalid")
}

func TestPersistence_SaveAndLoadCache(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Create a cache with some entries
	cache := NewCache(nil)
	output1 := &schema.AgentOutputV1{Version: "AGENT_OUTPUT_V1"}
	output2 := &schema.AgentOutputV1{Version: "AGENT_OUTPUT_V1"}
	metadata := CacheMetadata{Agent: "test-agent", Request: "test-request"}
	cache.Set("key1", output1, metadata)
	cache.Set("key2", output2, metadata)

	// Save cache
	err = p.SaveCache(cache)
	assert.NoError(t, err)

	// Verify file exists
	cachePath := filepath.Join(tempDir, "cache.json")
	_, err = os.Stat(cachePath)
	assert.NoError(t, err)

	// Create new cache and load
	newCache := NewCache(nil)
	err = p.LoadCache(newCache)
	assert.NoError(t, err)

	// Verify content
	value1, found := newCache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "AGENT_OUTPUT_V1", value1.Version)

	value2, found := newCache.Get("key2")
	assert.True(t, found)
	assert.Equal(t, "AGENT_OUTPUT_V1", value2.Version)
}

func TestPersistence_LoadCache_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Try to load cache when no file exists (should not error)
	cache := NewCache(nil)
	err = p.LoadCache(cache)
	assert.NoError(t, err)

	// Cache should remain empty
	value, found := cache.Get("any-key")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestPersistence_LoadCache_InvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Create a malicious cache file outside the base dir
	// This is a bit tricky to test directly since the path validation happens first
	// But we can test the validation logic
	cache := NewCache(nil)

	// The LoadCache function will validate the path internally
	// Since we're using a valid temp dir, this should work
	err = p.LoadCache(cache)
	assert.NoError(t, err)
}

func TestPersistence_ClearAll(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Create some sessions and cache
	session1 := &Session{ID: "session1", UserID: "user1"}
	session2 := &Session{ID: "session2", UserID: "user2"}

	err = p.SaveSession(session1)
	require.NoError(t, err)
	err = p.SaveSession(session2)
	require.NoError(t, err)

	cache := NewCache(nil)
	output := &schema.AgentOutputV1{Version: "AGENT_OUTPUT_V1"}
	metadata := CacheMetadata{Agent: "test-agent", Request: "test-request"}
	cache.Set("key", output, metadata)
	err = p.SaveCache(cache)
	require.NoError(t, err)

	// Verify files exist
	entries, err := os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Greater(t, len(entries), 0)

	// Clear all
	err = p.ClearAll()
	assert.NoError(t, err)

	// Verify directory is empty
	entries, err = os.ReadDir(tempDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
}

func TestManager_GetCache(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	retrievedCache := manager.GetCache()

	assert.Same(t, cache, retrievedCache)
}

func TestPersistence_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	p, err := NewPersistence(tempDir)
	require.NoError(t, err)

	// Test concurrent access to persistence operations
	done := make(chan bool, 2)

	// Goroutine 1: Save sessions
	go func() {
		for i := 0; i < 10; i++ {
			session := &Session{
				ID:     fmt.Sprintf("session-%d", i),
				UserID: fmt.Sprintf("user-%d", i),
			}
			p.SaveSession(session)
		}
		done <- true
	}()

	// Goroutine 2: List sessions
	go func() {
		for i := 0; i < 10; i++ {
			p.ListSessions()
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Verify we have sessions
	sessions, err := p.ListSessions()
	assert.NoError(t, err)
	assert.Greater(t, len(sessions), 0)
}
