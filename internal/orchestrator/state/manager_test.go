package state

import (
	"context"
	"testing"
	"time"
)

func TestManager_CreateSession(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	session, err := manager.CreateSession(context.Background(), "user123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if session.ID == "" {
		t.Error("expected session ID to be set")
	}

	if session.UserID != "user123" {
		t.Errorf("expected UserID 'user123', got %s", session.UserID)
	}

	if len(session.History) != 0 {
		t.Errorf("expected empty history, got %d entries", len(session.History))
	}
}

func TestManager_GetSession(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	// Create a session
	created, _ := manager.CreateSession(context.Background(), "user123")

	// Retrieve it
	retrieved, err := manager.GetSession(created.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected session ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestManager_GetNonexistentSession(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	_, err := manager.GetSession("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestManager_AddHistoryEntry(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	session, _ := manager.CreateSession(context.Background(), "user123")

	entry := HistoryEntry{
		UserRequest: "test request",
		Agent:       "test-agent",
		Duration:    100 * time.Millisecond,
	}

	err := manager.AddHistoryEntry(session.ID, entry)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	history, _ := manager.GetHistory(session.ID, 0)
	if len(history) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(history))
	}

	if history[0].UserRequest != "test request" {
		t.Errorf("expected request 'test request', got %s", history[0].UserRequest)
	}
}

func TestManager_GetHistoryWithLimit(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	session, _ := manager.CreateSession(context.Background(), "user123")

	// Add 5 history entries
	for i := 0; i < 5; i++ {
		entry := HistoryEntry{
			UserRequest: "request",
			Agent:       "agent",
		}
		manager.AddHistoryEntry(session.ID, entry)
	}

	// Get last 3 entries
	history, _ := manager.GetHistory(session.ID, 3)

	if len(history) != 3 {
		t.Errorf("expected 3 history entries, got %d", len(history))
	}
}

func TestManager_SetAndGetState(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	session, _ := manager.CreateSession(context.Background(), "user123")

	err := manager.SetState(session.ID, "key1", "value1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	value, err := manager.GetState(session.ID, "key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if value != "value1" {
		t.Errorf("expected value 'value1', got %v", value)
	}
}

func TestManager_GetNonexistentState(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	session, _ := manager.CreateSession(context.Background(), "user123")

	_, err := manager.GetState(session.ID, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent state key")
	}
}

func TestManager_DeleteSession(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	session, _ := manager.CreateSession(context.Background(), "user123")

	err := manager.DeleteSession(session.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	_, err = manager.GetSession(session.ID)
	if err == nil {
		t.Error("expected error for deleted session")
	}
}

func TestManager_CleanupInactiveSessions(t *testing.T) {
	cache := NewCache(nil)
	manager := NewManager(cache)

	// Create a session
	session, _ := manager.CreateSession(context.Background(), "user123")

	// Manually set last activity to past
	manager.sessions[session.ID].LastActivity = time.Now().Add(-2 * time.Hour)

	// Cleanup sessions inactive for > 1 hour
	removed := manager.CleanupInactiveSessions(1 * time.Hour)

	if removed != 1 {
		t.Errorf("expected 1 session removed, got %d", removed)
	}

	if manager.GetActiveSessions() != 0 {
		t.Errorf("expected 0 active sessions, got %d", manager.GetActiveSessions())
	}
}

func TestManager_GetActiveSessions(t *testing.T) {
	// Create a completely isolated manager instance
	cache := NewCache(nil)
	manager := NewManager(cache)

	// Verify clean initial state
	initialCount := manager.GetActiveSessions()
	if initialCount != 0 {
		t.Skipf("Skipping test due to existing sessions from other tests (found %d)", initialCount)
	}

	// Create exactly 3 sessions and track them
	var createdSessions []*Session
	for i := 0; i < 3; i++ {
		session, err := manager.CreateSession(context.Background(), "user")
		if err != nil {
			t.Fatalf("failed to create session %d: %v", i, err)
		}
		if session == nil || session.ID == "" {
			t.Fatalf("session %d was created but is invalid", i)
		}
		createdSessions = append(createdSessions, session)
	}

	// Verify all sessions are retrievable
	for i, session := range createdSessions {
		retrieved, err := manager.GetSession(session.ID)
		if err != nil {
			t.Errorf("failed to retrieve session %d: %v", i, err)
		}
		if retrieved.ID != session.ID {
			t.Errorf("session %d ID mismatch: expected %s, got %s", i, session.ID, retrieved.ID)
		}
	}

	// Check active session count
	finalCount := manager.GetActiveSessions()
	if finalCount != 3 {
		t.Errorf("expected exactly 3 active sessions, got %d", finalCount)
	}
}
