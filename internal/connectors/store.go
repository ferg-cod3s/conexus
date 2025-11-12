// Package connectors provides a SQLite-backed implementation of the ConnectorStore interface.
package connectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver" // Pure Go SQLite driver (WASM-based)
)

// Connector represents a connector configuration
type Connector struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Status    string                 `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// ConnectorStore defines the interface for connector persistence
type ConnectorStore interface {
	Add(ctx context.Context, connector *Connector) error
	Update(ctx context.Context, id string, connector *Connector) error
	Remove(ctx context.Context, id string) error
	List(ctx context.Context) ([]*Connector, error)
	Get(ctx context.Context, id string) (*Connector, error)
	Close() error
}

// Store is a SQLite-backed connector store
type Store struct {
	db *sql.DB
}

// NewStore creates a new SQLite connector store.
// The path can be ":memory:" for in-memory database or a file path for persistence.
func NewStore(path string) (*Store, error) {
	// Create directory for database file if it doesn't exist
	if path != ":memory:" {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// For :memory: databases, limit to 1 connection to ensure all goroutines
	// share the same database. Without this, the connection pool creates separate
	// in-memory databases per connection, causing "no such table" errors.
	db.SetMaxOpenConns(1)

	store := &Store{db: db}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		// #nosec G104 - Best-effort cleanup in error path, primary error (schema init) already captured
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return store, nil
}

// initSchema creates the required tables and indexes.
func (s *Store) initSchema() error {
	schema := `
	-- Connectors table
	CREATE TABLE IF NOT EXISTS connectors (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		config TEXT NOT NULL,  -- JSON-encoded config
		status TEXT NOT NULL DEFAULT 'active',
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	-- Index for faster lookups by type
	CREATE INDEX IF NOT EXISTS idx_connectors_type ON connectors(type);

	-- Index for faster lookups by status
	CREATE INDEX IF NOT EXISTS idx_connectors_status ON connectors(status);
	`

	_, err := s.db.Exec(schema)
	return err
}

// validateConnector performs validation on connector data
func validateConnector(connector *Connector) error {
	if connector.ID == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}
	if connector.Name == "" {
		return fmt.Errorf("connector name cannot be empty")
	}
	if connector.Type == "" {
		return fmt.Errorf("connector type cannot be empty")
	}

	// Validate connector type
	validTypes := map[string]bool{
		"filesystem": true,
		"github":     true,
		"git":        true,
		"database":   true,
		"api":        true,
		"s3":         true,
		"http":       true,
	}

	if !validTypes[connector.Type] {
		return fmt.Errorf("invalid connector type: %s", connector.Type)
	}

	// Validate status
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
		"error":    true,
	}

	if !validStatuses[connector.Status] {
		return fmt.Errorf("invalid connector status: %s", connector.Status)
	}

	return nil
}

// Add inserts a new connector with validation
func (s *Store) Add(ctx context.Context, connector *Connector) error {
	if connector == nil {
		return fmt.Errorf("connector cannot be nil")
	}

	if err := validateConnector(connector); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if connector already exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM connectors WHERE id = ?)", connector.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check existence: %w", err)
	}
	if exists {
		return fmt.Errorf("connector with ID %s already exists", connector.ID)
	}

	// Serialize config as JSON
	configJSON, err := json.Marshal(connector.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	now := time.Now().Unix()
	createdAt := now
	if !connector.CreatedAt.IsZero() {
		createdAt = connector.CreatedAt.Unix()
	}
	updatedAt := now
	if !connector.UpdatedAt.IsZero() {
		updatedAt = connector.UpdatedAt.Unix()
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO connectors (id, name, type, config, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		connector.ID, connector.Name, connector.Type, configJSON, connector.Status, createdAt, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert connector: %w", err)
	}

	return nil
}

// Update modifies an existing connector
func (s *Store) Update(ctx context.Context, id string, connector *Connector) error {
	if id == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}
	if connector == nil {
		return fmt.Errorf("connector cannot be nil")
	}

	// Set the ID from the parameter
	connector.ID = id

	if err := validateConnector(connector); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if connector exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM connectors WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("connector with ID %s not found", id)
	}

	// Serialize config as JSON
	configJSON, err := json.Marshal(connector.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	now := time.Now().Unix()
	updatedAt := now
	if !connector.UpdatedAt.IsZero() {
		updatedAt = connector.UpdatedAt.Unix()
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE connectors SET name = ?, type = ?, config = ?, status = ?, updated_at = ? WHERE id = ?`,
		connector.Name, connector.Type, configJSON, connector.Status, updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("update connector: %w", err)
	}

	return nil
}

// Remove deletes a connector by ID
func (s *Store) Remove(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}

	result, err := s.db.ExecContext(ctx, "DELETE FROM connectors WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete connector: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("connector %s not found", id)
	}

	return nil
}

// List returns all connectors
func (s *Store) List(ctx context.Context) ([]*Connector, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, config, status, created_at, updated_at
		 FROM connectors ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query connectors: %w", err)
	}
	defer rows.Close()

	var connectors []*Connector
	for rows.Next() {
		var connector Connector
		var configJSON []byte
		var createdAt, updatedAt int64

		err := rows.Scan(&connector.ID, &connector.Name, &connector.Type, &configJSON, &connector.Status, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan connector: %w", err)
		}

		// Deserialize config
		if err := json.Unmarshal(configJSON, &connector.Config); err != nil {
			return nil, fmt.Errorf("unmarshal config for connector %s: %w", connector.ID, err)
		}

		connector.CreatedAt = time.Unix(createdAt, 0)
		connector.UpdatedAt = time.Unix(updatedAt, 0)

		connectors = append(connectors, &connector)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate connectors: %w", err)
	}

	return connectors, nil
}

// Get retrieves a connector by ID
func (s *Store) Get(ctx context.Context, id string) (*Connector, error) {
	if id == "" {
		return nil, fmt.Errorf("connector ID cannot be empty")
	}

	var connector Connector
	var configJSON []byte
	var createdAt, updatedAt int64

	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, config, status, created_at, updated_at
		 FROM connectors WHERE id = ?`,
		id,
	).Scan(&connector.ID, &connector.Name, &connector.Type, &configJSON, &connector.Status, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("connector %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query connector: %w", err)
	}

	// Deserialize config
	if err := json.Unmarshal(configJSON, &connector.Config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	connector.CreatedAt = time.Unix(createdAt, 0)
	connector.UpdatedAt = time.Unix(updatedAt, 0)

	return &connector, nil
}

// Close releases database resources
func (s *Store) Close() error {
	return s.db.Close()
}
