// Package sqlite provides a SQLite-backed implementation of the VectorStore interface.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver

	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// Store is a SQLite-backed vector store with FTS5 support for BM25 search.
type Store struct {
	db *sql.DB
}

// NewStore creates a new SQLite vector store.
// The path can be ":memory:" for in-memory database or a file path for persistence.
func NewStore(path string) (*Store, error) {
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
	-- Main documents table
	CREATE TABLE IF NOT EXISTS documents (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		vector TEXT NOT NULL,  -- JSON-encoded float array
		metadata TEXT,         -- JSON-encoded metadata
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);

	-- FTS5 virtual table for full-text search
	CREATE VIRTUAL TABLE IF NOT EXISTS documents_fts USING fts5(
		id UNINDEXED,
		content,
		tokenize='porter unicode61'
	);

	-- Triggers to keep FTS5 in sync
	CREATE TRIGGER IF NOT EXISTS documents_ai AFTER INSERT ON documents BEGIN
		INSERT INTO documents_fts(id, content) VALUES (new.id, new.content);
	END;

	CREATE TRIGGER IF NOT EXISTS documents_ad AFTER DELETE ON documents BEGIN
		DELETE FROM documents_fts WHERE id = old.id;
	END;

	CREATE TRIGGER IF NOT EXISTS documents_au AFTER UPDATE ON documents BEGIN
		UPDATE documents_fts SET content = new.content WHERE id = old.id;
	END;

	-- Index for metadata filtering (will add JSON support later)
	CREATE INDEX IF NOT EXISTS idx_documents_updated_at ON documents(updated_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Upsert inserts or updates a document with its vector.
func (s *Store) Upsert(ctx context.Context, doc vectorstore.Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document ID cannot be empty")
	}
	if len(doc.Vector) == 0 {
		return fmt.Errorf("document vector cannot be empty")
	}

	// Serialize vector as JSON
	vectorJSON, err := json.Marshal(doc.Vector)
	if err != nil {
		return fmt.Errorf("marshal vector: %w", err)
	}

	// Serialize metadata as JSON
	var metadataJSON []byte
	if doc.Metadata != nil {
		metadataJSON, err = json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	}

	// Check if document exists
	var exists bool
	err = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM documents WHERE id = ?)", doc.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check existence: %w", err)
	}

	now := time.Now().Unix()

	if exists {
		// Update existing document
		_, err = s.db.ExecContext(ctx,
			`UPDATE documents SET content = ?, vector = ?, metadata = ?, updated_at = ? WHERE id = ?`,
			doc.Content, vectorJSON, metadataJSON, now, doc.ID,
		)
		if err != nil {
			return fmt.Errorf("update document: %w", err)
		}
	} else {
		// Insert new document
		createdAt := now
		if !doc.CreatedAt.IsZero() {
			createdAt = doc.CreatedAt.Unix()
		}
		updatedAt := now
		if !doc.UpdatedAt.IsZero() {
			updatedAt = doc.UpdatedAt.Unix()
		}

		_, err = s.db.ExecContext(ctx,
			`INSERT INTO documents (id, content, vector, metadata, created_at, updated_at) 
			 VALUES (?, ?, ?, ?, ?, ?)`,
			doc.ID, doc.Content, vectorJSON, metadataJSON, createdAt, updatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert document: %w", err)
		}
	}

	return nil
}

// UpsertBatch efficiently inserts or updates multiple documents in a transaction.
func (s *Store) UpsertBatch(ctx context.Context, docs []vectorstore.Document) error {
	if len(docs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, doc := range docs {
		if err := s.upsertInTx(ctx, tx, doc); err != nil {
			return fmt.Errorf("upsert document %s: %w", doc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// upsertInTx performs upsert within a transaction.
func (s *Store) upsertInTx(ctx context.Context, tx *sql.Tx, doc vectorstore.Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document ID cannot be empty")
	}
	if len(doc.Vector) == 0 {
		return fmt.Errorf("document vector cannot be empty")
	}

	vectorJSON, err := json.Marshal(doc.Vector)
	if err != nil {
		return fmt.Errorf("marshal vector: %w", err)
	}

	var metadataJSON []byte
	if doc.Metadata != nil {
		metadataJSON, err = json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	}

	now := time.Now().Unix()

	// Use INSERT OR REPLACE for efficiency
	createdAt := now
	if !doc.CreatedAt.IsZero() {
		createdAt = doc.CreatedAt.Unix()
	}
	updatedAt := now
	if !doc.UpdatedAt.IsZero() {
		updatedAt = doc.UpdatedAt.Unix()
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO documents (id, content, vector, metadata, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		 content = excluded.content,
		 vector = excluded.vector,
		 metadata = excluded.metadata,
		 updated_at = excluded.updated_at`,
		doc.ID, doc.Content, vectorJSON, metadataJSON, createdAt, updatedAt,
	)

	return err
}

// Delete removes a document by ID.
func (s *Store) Delete(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM documents WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("document %s not found", id)
	}

	return nil
}

// Get retrieves a document by ID.
func (s *Store) Get(ctx context.Context, id string) (*vectorstore.Document, error) {
	var doc vectorstore.Document
	var vectorJSON, metadataJSON []byte
	var createdAt, updatedAt int64

	err := s.db.QueryRowContext(ctx,
		`SELECT id, content, vector, metadata, created_at, updated_at 
		 FROM documents WHERE id = ?`,
		id,
	).Scan(&doc.ID, &doc.Content, &vectorJSON, &metadataJSON, &createdAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("document %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("query document: %w", err)
	}

	// Deserialize vector
	if err := json.Unmarshal(vectorJSON, &doc.Vector); err != nil {
		return nil, fmt.Errorf("unmarshal vector: %w", err)
	}

	// Deserialize metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	doc.CreatedAt = time.Unix(createdAt, 0)
	doc.UpdatedAt = time.Unix(updatedAt, 0)

	return &doc, nil
}

// Count returns the total number of documents.
func (s *Store) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return count, nil
}

// ListIndexedFiles returns a list of all unique file paths that have been indexed.
func (s *Store) ListIndexedFiles(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT json_extract(metadata, '$.file_path') as file_path
		FROM documents
		WHERE metadata IS NOT NULL AND json_extract(metadata, '$.file_path') IS NOT NULL
		ORDER BY file_path
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query indexed files: %w", err)
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var filePath string
		if err := rows.Scan(&filePath); err != nil {
			return nil, fmt.Errorf("scan file path: %w", err)
		}
		files = append(files, filePath)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return files, nil
}

// GetFileChunks returns all chunks for a specific file path, sorted by start_line.
func (s *Store) GetFileChunks(ctx context.Context, filePath string) ([]vectorstore.Document, error) {
	query := `
		SELECT id, content, vector, metadata, created_at, updated_at
		FROM documents
		WHERE metadata IS NOT NULL AND json_extract(metadata, '$.file_path') = ?
		ORDER BY json_extract(metadata, '$.start_line')
	`

	rows, err := s.db.QueryContext(ctx, query, filePath)
	if err != nil {
		return nil, fmt.Errorf("query file chunks: %w", err)
	}
	defer rows.Close()

	var docs []vectorstore.Document
	for rows.Next() {
		var doc vectorstore.Document
		var vectorJSON, metadataJSON []byte
		var createdAt, updatedAt int64

		err := rows.Scan(&doc.ID, &doc.Content, &vectorJSON, &metadataJSON, &createdAt, &updatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}

		if err := deserializeDocument(&doc, vectorJSON, metadataJSON, createdAt, updatedAt); err != nil {
			return nil, fmt.Errorf("deserialize document %s: %w", doc.ID, err)
		}

		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}

	return docs, nil
}

// Close releases database resources.
func (s *Store) Close() error {
	return s.db.Close()
}

// Stats returns index statistics.
func (s *Store) Stats(ctx context.Context) (*vectorstore.IndexStats, error) {
	stats := &vectorstore.IndexStats{
		Languages: make(map[string]int64),
	}

	// Get total documents
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents").Scan(&stats.TotalDocuments)
	if err != nil {
		return nil, fmt.Errorf("count documents: %w", err)
	}
	stats.TotalChunks = stats.TotalDocuments

	// Get last indexed timestamp
	var lastUpdated sql.NullInt64
	err = s.db.QueryRowContext(ctx, "SELECT MAX(updated_at) FROM documents").Scan(&lastUpdated)
	if err != nil {
		return nil, fmt.Errorf("get last updated: %w", err)
	}
	if lastUpdated.Valid {
		stats.LastIndexedAt = time.Unix(lastUpdated.Int64, 0)
	}

	// Count documents by language (from metadata)
	rows, err := s.db.QueryContext(ctx, "SELECT metadata FROM documents WHERE metadata IS NOT NULL")
	if err != nil {
		return nil, fmt.Errorf("query metadata: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var metadataJSON []byte
		if err := rows.Scan(&metadataJSON); err != nil {
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			continue
		}

		if lang, ok := metadata["language"].(string); ok {
			stats.Languages[lang]++
		}
	}

	// Get database file size (approximate)
	// Note: For in-memory databases, this will be 0
	err = s.db.QueryRowContext(ctx, "SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()").Scan(&stats.IndexSize)
	if err != nil {
		// Ignore error for in-memory databases
		stats.IndexSize = 0
	}

	return stats, nil
}

// deserializeDocument unmarshals vector and metadata JSON into a document.
// This is shared by Get() and SearchBM25() to avoid code duplication.
func deserializeDocument(doc *vectorstore.Document, vectorJSON, metadataJSON []byte, createdAt, updatedAt int64) error {
	// Deserialize vector
	if err := json.Unmarshal(vectorJSON, &doc.Vector); err != nil {
		return fmt.Errorf("unmarshal vector: %w", err)
	}

	// Deserialize metadata if present
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	// Convert timestamps
	doc.CreatedAt = time.Unix(createdAt, 0)
	doc.UpdatedAt = time.Unix(updatedAt, 0)

	return nil
}
