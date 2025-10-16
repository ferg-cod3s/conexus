package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ferg-cod3s/conexus/internal/security"
	"github.com/ferg-cod3s/conexus/internal/validation"
)

// Persistence handles state persistence to disk
type Persistence struct {
	mu      sync.RWMutex
	baseDir string
}

// NewPersistence creates a new persistence handler
func NewPersistence(baseDir string) (*Persistence, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Persistence{
		baseDir: baseDir,
	}, nil
}

// SaveSession persists a session to disk
func (p *Persistence) SaveSession(session *Session) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sessionPath := filepath.Join(p.baseDir, fmt.Sprintf("session-%s.json", session.ID))

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(sessionPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session from disk
func (p *Persistence) LoadSession(sessionID string) (*Session, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	sessionPath := filepath.Join(p.baseDir, fmt.Sprintf("session-%s.json", sessionID))

	// G304: Validate path to prevent directory traversal
	if _, err := security.ValidatePathWithinBase(sessionPath, p.baseDir); err != nil {
		return nil, fmt.Errorf("invalid session path: %w", err)
	}

	// #nosec G304 - Path validated at line 59 with ValidatePathWithinBase
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// DeleteSession removes a persisted session
func (p *Persistence) DeleteSession(sessionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	sessionPath := filepath.Join(p.baseDir, fmt.Sprintf("session-%s.json", sessionID))

	if err := os.Remove(sessionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}

// ListSessions returns all persisted session IDs
func (p *Persistence) ListSessions() ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entries, err := os.ReadDir(p.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	sessionIDs := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Validate filename to prevent path traversal
		if err := validation.IsPathSafe(name); err != nil {
			continue // Skip invalid filenames
		}

		if len(name) > 13 && name[:8] == "session-" && name[len(name)-5:] == ".json" {
			sessionID := name[8 : len(name)-5]
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	return sessionIDs, nil
}

// SaveCache persists cache entries to disk
func (p *Persistence) SaveCache(cache *Cache) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	cachePath := filepath.Join(p.baseDir, "cache.json")

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	data, err := json.MarshalIndent(cache.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// LoadCache loads cache entries from disk
func (p *Persistence) LoadCache(cache *Cache) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	cachePath := filepath.Join(p.baseDir, "cache.json")

	// G304: Validate path to prevent directory traversal
	if _, err := security.ValidatePathWithinBase(cachePath, p.baseDir); err != nil {
		return fmt.Errorf("invalid cache path: %w", err)
	}

	// #nosec G304 - Path validated at line 151 with ValidatePathWithinBase
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No cache file yet
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	if err := json.Unmarshal(data, &cache.entries); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	return nil
}

// ClearAll removes all persisted data
func (p *Persistence) ClearAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entries, err := os.ReadDir(p.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Validate filename to prevent path traversal
		if err := validation.IsPathSafe(entry.Name()); err != nil {
			continue // Skip invalid filenames
		}

		path := filepath.Join(p.baseDir, entry.Name())
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", path, err)
		}
	}

	return nil
}
