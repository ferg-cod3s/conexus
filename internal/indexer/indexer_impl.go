// Package indexer provides concrete implementation of the Indexer interface.
package indexer

import (
	"context"
	"errors"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/security"
	"github.com/ferg-cod3s/conexus/internal/validation"
)

// DefaultIndexer implements the Indexer interface with incremental indexing support.
type DefaultIndexer struct {
	walker     Walker
	merkleTree MerkleTree
	chunkers   []Chunker
	statePath  string // Where to persist merkle tree state
}

// NewIndexer creates a new indexer with default components.
func NewIndexer(statePath string) *DefaultIndexer {
	return &DefaultIndexer{
		walker:     NewFileWalker(1024 * 1024), // 1MB max file size default
		merkleTree: NewMerkleTree(NewFileWalker(0)),
		chunkers:   []Chunker{}, // Will be populated as chunkers are implemented
		statePath:  statePath,
	}
}

// Index performs a full index of the file system.
func (idx *DefaultIndexer) Index(ctx context.Context, opts IndexOptions) ([]Chunk, error) {
	var chunks []Chunk
	
	// Walk the file system
	err := idx.walker.Walk(ctx, opts.RootPath, opts.IgnorePatterns, func(path string, info os.FileInfo) error {
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Skip files exceeding max size
		if opts.MaxFileSize > 0 && info.Size() > opts.MaxFileSize {
			return nil
		}
		
		// G304: Validate path before reading file
		if _, err := security.ValidatePathWithinBase(path, opts.RootPath); err != nil {
			if errors.Is(err, security.ErrPathTraversal) {
				return fmt.Errorf("security: path traversal detected for %s: %w", path, err)
			}
			return fmt.Errorf("path validation failed for %s: %w", path, err)
		}
		// Read file content
		// #nosec G304 - Path validated at line 56 with ValidatePathWithinBase
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}
		
		// Get relative path from root
		relPath, err := filepath.Rel(opts.RootPath, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}
		
		// Validate path for security
		if err := validation.IsPathSafe(relPath); err != nil {
			return fmt.Errorf("path validation failed for %s: %w", relPath, err)
		}
		
		// Find appropriate chunker
		chunker := idx.findChunker(path)
		if chunker == nil {
			// No chunker available, create a single chunk for the whole file
			chunk := idx.createSingleChunk(string(content), relPath, info)
			chunks = append(chunks, chunk)
			return nil
		}
		
		// Chunk the file
		fileChunks, err := chunker.Chunk(ctx, string(content), relPath)
		if err != nil {
			// If chunking fails, fall back to single chunk
			chunk := idx.createSingleChunk(string(content), relPath, info)
			chunks = append(chunks, chunk)
			return nil
		}
		
		chunks = append(chunks, fileChunks...)
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("walk file system: %w", err)
	}
	
	// If embedder and vectorstore provided, generate and store vectors
	if opts.Embedder != nil && opts.VectorStore != nil {
		if err := idx.storeVectors(ctx, chunks, opts); err != nil {
			return nil, fmt.Errorf("store vectors: %w", err)
		}
	}
	
	return chunks, nil
}

// IndexIncremental only indexes files that have changed since the last run.
func (idx *DefaultIndexer) IndexIncremental(ctx context.Context, opts IndexOptions, previousState []byte) ([]Chunk, []byte, error) {
	// 1. Load or create state directory
	if err := idx.ensureStateDir(); err != nil {
		return nil, nil, fmt.Errorf("ensure state dir: %w", err)
	}
	
	// 2. Hash current state
	currentState, err := idx.merkleTree.Hash(ctx, opts.RootPath, opts.IgnorePatterns)
	if err != nil {
		return nil, nil, fmt.Errorf("hash current state: %w", err)
	}
	
	// 3. If no previous state, do full index
	if previousState == nil || len(previousState) == 0 {
		chunks, err := idx.Index(ctx, opts)
		if err != nil {
			return nil, nil, fmt.Errorf("full index: %w", err)
		}
		return chunks, currentState, nil
	}
	
	// 4. Diff to find changed files
	changedPaths, err := idx.merkleTree.Diff(ctx, previousState, currentState)
	if err != nil {
		return nil, nil, fmt.Errorf("diff states: %w", err)
	}
	
	// 5. If no changes, return empty
	if len(changedPaths) == 0 {
		return []Chunk{}, currentState, nil
	}
	
	// 6. Index only changed files
	var chunks []Chunk
	deletedPaths := make(map[string]bool)
	
	for _, relPath := range changedPaths {
		// Validate path for security
		if err := validation.IsPathSafe(relPath); err != nil {
			return nil, nil, fmt.Errorf("path validation failed for %s: %w", relPath, err)
		}
		
		fullPath := filepath.Join(opts.RootPath, relPath)
		
		// Check if file still exists (might have been deleted)
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted, track it
				deletedPaths[relPath] = true
				continue
			}
			return nil, nil, fmt.Errorf("stat file %s: %w", fullPath, err)
		}
		
		// Skip directories
		if info.IsDir() {
			continue
		}
		
		// Skip files exceeding max size
		if opts.MaxFileSize > 0 && info.Size() > opts.MaxFileSize {
			continue
		}
		
		// G304: Validate path before reading file
		if _, err := security.ValidatePathWithinBase(fullPath, opts.RootPath); err != nil {
			if errors.Is(err, security.ErrPathTraversal) {
				return nil, nil, fmt.Errorf("security: path traversal detected for %s: %w", fullPath, err)
			}
			return nil, nil, fmt.Errorf("path validation failed for %s: %w", fullPath, err)
		}
		// Read file content
		// #nosec G304 - Path validated at line 183 with ValidatePathWithinBase
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, nil, fmt.Errorf("read file %s: %w", fullPath, err)
		}
		
		// Find appropriate chunker
		chunker := idx.findChunker(fullPath)
		if chunker == nil {
			// No chunker available, create a single chunk for the whole file
			chunk := idx.createSingleChunk(string(content), relPath, info)
			chunks = append(chunks, chunk)
			continue
		}
		
		// Chunk the file
		fileChunks, err := chunker.Chunk(ctx, string(content), relPath)
		if err != nil {
			// If chunking fails, fall back to single chunk
			chunk := idx.createSingleChunk(string(content), relPath, info)
			chunks = append(chunks, chunk)
			continue
		}
		
		chunks = append(chunks, fileChunks...)
	}
	
	// Handle vector store updates for incremental indexing
	if opts.VectorStore != nil {
		// Delete vectors for removed files
		if err := idx.deleteVectorsForPaths(ctx, deletedPaths, opts.VectorStore); err != nil {
			return nil, nil, fmt.Errorf("delete vectors: %w", err)
		}
		
		// Delete old vectors for changed files (will be replaced)
		changedFilePaths := make(map[string]bool)
		for _, chunk := range chunks {
			changedFilePaths[chunk.FilePath] = true
		}
		if err := idx.deleteVectorsForPaths(ctx, changedFilePaths, opts.VectorStore); err != nil {
			return nil, nil, fmt.Errorf("delete old vectors: %w", err)
		}
		
		// Store new vectors if embedder available
		if opts.Embedder != nil {
			if err := idx.storeVectors(ctx, chunks, opts); err != nil {
				return nil, nil, fmt.Errorf("store vectors: %w", err)
			}
		}
	}
	
	return chunks, currentState, nil
}

// storeVectors generates embeddings and stores them in the vector store.
func (idx *DefaultIndexer) storeVectors(ctx context.Context, chunks []Chunk, opts IndexOptions) error {
	if len(chunks) == 0 {
		return nil
	}
	
	var docs []vectorstore.Document
	for _, chunk := range chunks {
		// Generate embedding
		embedding, err := opts.Embedder.Embed(ctx, chunk.Content)
		if err != nil {
			return fmt.Errorf("embed chunk %s: %w", chunk.ID, err)
		}
		
		// Convert to document
		doc := chunkToDocument(chunk, embedding.Vector)
		docs = append(docs, doc)
	}
	
	// Batch upsert
	if err := opts.VectorStore.UpsertBatch(ctx, docs); err != nil {
		return fmt.Errorf("upsert batch: %w", err)
	}
	
	return nil
}

// deleteVectorsForPaths removes vectors for the given file paths.
func (idx *DefaultIndexer) deleteVectorsForPaths(ctx context.Context, paths map[string]bool, store vectorstore.VectorStore) error {
	for path := range paths {
		// Query by file path metadata
		filter := map[string]interface{}{"file_path": path}
		// Use empty vector for query (we only care about metadata filter)
		opts := vectorstore.SearchOptions{
			Limit: 10000,
			Filters: filter,
		}
		results, err := store.SearchVector(ctx, nil, opts)
		if err != nil {
			// Non-fatal: log but continue
			continue
		}
		
		// Delete all chunks for this file
		for _, result := range results {
			if err := store.Delete(ctx, result.Document.ID); err != nil {
				// Non-fatal: log but continue
				continue
			}
		}
	}
	
	return nil
}

// chunkToDocument converts a Chunk to a vectorstore.Document.
func chunkToDocument(chunk Chunk, vector embedding.Vector) vectorstore.Document {
	return vectorstore.Document{
		ID:      chunk.ID,
		Content: chunk.Content,
		Vector:  vector,
		Metadata: map[string]interface{}{
			"file_path":  chunk.FilePath,
			"language":   chunk.Language,
			"type":       string(chunk.Type),
			"start_line": chunk.StartLine,
			"end_line":   chunk.EndLine,
			"hash":       chunk.Hash,
		},
		CreatedAt: chunk.IndexedAt,
		UpdatedAt: chunk.IndexedAt,
	}
}

// SaveState persists the Merkle tree state to disk.
func (idx *DefaultIndexer) SaveState(ctx context.Context, state []byte) error {
	if err := idx.ensureStateDir(); err != nil {
		return fmt.Errorf("ensure state dir: %w", err)
	}
	
	if err := os.WriteFile(idx.statePath, state, 0600); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	
	return nil
}

// LoadState reads the persisted Merkle tree state from disk.
func (idx *DefaultIndexer) LoadState(ctx context.Context) ([]byte, error) {
	data, err := os.ReadFile(idx.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// No previous state exists, return nil
			return nil, nil
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}
	
	return data, nil
}

// GetMetrics returns statistics about the last indexing operation.
func (idx *DefaultIndexer) GetMetrics() IndexMetrics {
	// TODO: Implement metrics collection during indexing
	return IndexMetrics{}
}

// IndexMetrics provides statistics about indexing operations.
type IndexMetrics struct {
	TotalFiles      int           // Total files scanned
	IndexedFiles    int           // Files actually indexed (changed)
	SkippedFiles    int           // Files skipped (unchanged)
	TotalChunks     int           // Total chunks created
	Duration        time.Duration // Time taken
	BytesProcessed  int64         // Total bytes processed
	StateSize       int64         // Size of merkle state in bytes
	IncrementalSave time.Duration // Time saved by incremental approach
}

// Helper: findChunker selects the appropriate chunker for a file.
func (idx *DefaultIndexer) findChunker(path string) Chunker {
	ext := filepath.Ext(path)
	for _, chunker := range idx.chunkers {
		if chunker.Supports(ext) {
			return chunker
		}
	}
	return nil
}

// Helper: createSingleChunk creates a single chunk for an entire file.
func (idx *DefaultIndexer) createSingleChunk(content, relPath string, info os.FileInfo) Chunk {
	hash := sha256.Sum256([]byte(content))
	
	return Chunk{
		ID:        hex.EncodeToString(hash[:]),
		Content:   content,
		FilePath:  relPath,
		Language:  detectLanguage(relPath),
		Type:      ChunkTypeUnknown,
		StartLine: 1,
		EndLine:   countLines(content),
		Metadata:  map[string]string{},
		Hash:      hex.EncodeToString(hash[:]),
		IndexedAt: time.Now(),
	}
}

// Helper: ensureStateDir creates the state directory if it doesn't exist.
func (idx *DefaultIndexer) ensureStateDir() error {
	dir := filepath.Dir(idx.statePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	return nil
}

// Helper: detectLanguage attempts to detect the programming language from file extension.
func detectLanguage(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "cpp"
	case ".c":
		return "c"
	case ".md":
		return "markdown"
	case ".txt":
		return "text"
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".toml":
		return "toml"
	default:
		return "unknown"
	}
}

// Helper: countLines counts the number of lines in a string.
func countLines(s string) int {
	if len(s) == 0 {
		return 0
	}
	lines := 1
	for _, c := range s {
		if c == '\n' {
			lines++
		}
	}
	return lines
}

// StateManager provides persistence for indexer state.
type StateManager struct {
	statePath string
}

// NewStateManager creates a new state manager.
func NewStateManager(statePath string) *StateManager {
	return &StateManager{
		statePath: statePath,
	}
}

// Save persists state to disk with atomic write.
func (sm *StateManager) Save(ctx context.Context, state []byte) error {
	// Write to temp file first
	tempPath := sm.statePath + ".tmp"
	if err := os.WriteFile(tempPath, state, 0600); err != nil {
		return fmt.Errorf("write temp state: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempPath, sm.statePath); err != nil {
		// #nosec G104 - Best-effort cleanup of temp file, primary error (rename failure) already captured
		os.Remove(tempPath) // Clean up on failure
		return fmt.Errorf("rename state file: %w", err)
	}
	
	return nil
}

// Load reads state from disk.
func (sm *StateManager) Load(ctx context.Context) ([]byte, error) {
	data, err := os.ReadFile(sm.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No previous state
		}
		return nil, fmt.Errorf("read state: %w", err)
	}
	
	return data, nil
}

// Clear removes the persisted state.
func (sm *StateManager) Clear(ctx context.Context) error {
	if err := os.Remove(sm.statePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove state: %w", err)
	}
	return nil
}

// Exists checks if state file exists.
func (sm *StateManager) Exists() bool {
	_, err := os.Stat(sm.statePath)
	return err == nil
}

// IndexStats tracks statistics during indexing.
type IndexStats struct {
	StartTime      time.Time
	EndTime        time.Time
	TotalFiles     int
	IndexedFiles   int
	SkippedFiles   int
	TotalChunks    int
	BytesProcessed int64
	Errors         []error
}

// Duration returns the total indexing duration.
func (s *IndexStats) Duration() time.Duration {
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}

// ToMetrics converts IndexStats to IndexMetrics.
func (s *IndexStats) ToMetrics() IndexMetrics {
	return IndexMetrics{
		TotalFiles:     s.TotalFiles,
		IndexedFiles:   s.IndexedFiles,
		SkippedFiles:   s.SkippedFiles,
		TotalChunks:    s.TotalChunks,
		Duration:       s.Duration(),
		BytesProcessed: s.BytesProcessed,
	}
}

// StateInfo provides metadata about persisted state.
type StateInfo struct {
	Version   string    `json:"version"`
	RootPath  string    `json:"root_path"`
	Timestamp time.Time `json:"timestamp"`
	FileCount int       `json:"file_count"`
	StateHash string    `json:"state_hash"`
}

// MarshalState serializes state with metadata.
func MarshalState(state []byte, info StateInfo) ([]byte, error) {
	wrapper := struct {
		Info  StateInfo `json:"info"`
		State []byte    `json:"state"`
	}{
		Info:  info,
		State: state,
	}
	
	return json.Marshal(wrapper)
}

// UnmarshalState deserializes state and extracts metadata.
func UnmarshalState(data []byte) ([]byte, StateInfo, error) {
	var wrapper struct {
		Info  StateInfo `json:"info"`
		State []byte    `json:"state"`
	}
	
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, StateInfo{}, fmt.Errorf("unmarshal state: %w", err)
	}
	
	return wrapper.State, wrapper.Info, nil
}
