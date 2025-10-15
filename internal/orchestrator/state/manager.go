// Package state provides conversation state management and caching.
//
// The state manager handles:
// - Conversation history tracking
// - Result caching
// - Session management
// - State persistence
package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Manager manages conversation state and caching
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	cache    *Cache
}

// NewManager creates a new state manager
func NewManager(cache *Cache) *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
		cache:    cache,
	}
}

// Session represents a conversation session
type Session struct {
	// Unique session identifier
	ID string

	// User identifier (optional)
	UserID string

	// Conversation history
	History []HistoryEntry

	// Session metadata
	Metadata map[string]interface{}

	// Session creation time
	CreatedAt time.Time

	// Last activity time
	LastActivity time.Time

	// Session state
	State map[string]interface{}
}

// HistoryEntry represents a single interaction in the conversation
type HistoryEntry struct {
	// Timestamp of the interaction
	Timestamp time.Time

	// User request
	UserRequest string

	// Agent that handled the request
	Agent string

	// Agent response
	Response *schema.AgentOutputV1

	// Escalations that occurred
	Escalations []EscalationRecord

	// Execution time
	Duration time.Duration
}

// EscalationRecord tracks an escalation event
type EscalationRecord struct {
	SourceAgent string
	TargetAgent string
	Reason      string
	Timestamp   time.Time
}

// CreateSession creates a new session
func (m *Manager) CreateSession(ctx context.Context, userID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := generateSessionID()

	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		History:      make([]HistoryEntry, 0),
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		State:        make(map[string]interface{}),
	}

	m.sessions[sessionID] = session

	return session, nil
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// AddHistoryEntry adds an entry to the session history
func (m *Manager) AddHistoryEntry(sessionID string, entry HistoryEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	entry.Timestamp = time.Now()
	session.History = append(session.History, entry)
	session.LastActivity = time.Now()

	return nil
}

// GetHistory returns the conversation history for a session
func (m *Manager) GetHistory(sessionID string, limit int) ([]HistoryEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	if limit <= 0 || limit > len(session.History) {
		return session.History, nil
	}

	// Return the most recent entries
	start := len(session.History) - limit
	return session.History[start:], nil
}

// SetState sets a state value for a session
func (m *Manager) SetState(sessionID string, key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.State[key] = value
	session.LastActivity = time.Now()

	return nil
}

// GetState retrieves a state value from a session
func (m *Manager) GetState(sessionID string, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	value, ok := session.State[key]
	if !ok {
		return nil, fmt.Errorf("state key not found: %s", key)
	}

	return value, nil
}

// DeleteSession removes a session
func (m *Manager) DeleteSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[sessionID]; !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(m.sessions, sessionID)

	return nil
}

// CleanupInactiveSessions removes sessions that haven't been active
func (m *Manager) CleanupInactiveSessions(maxInactivity time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxInactivity)
	removed := 0

	for sessionID, session := range m.sessions {
		if session.LastActivity.Before(cutoff) {
			delete(m.sessions, sessionID)
			removed++
		}
	}

	return removed
}

// GetCache returns the cache instance
func (m *Manager) GetCache() *Cache {
	return m.cache
}

// GetActiveSessions returns the number of active sessions
func (m *Manager) GetActiveSessions() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.sessions)
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
