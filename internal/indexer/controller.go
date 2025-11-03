// Package indexer provides background indexing controller implementation.
package indexer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// DefaultIndexController implements IndexController interface with background operations.
type DefaultIndexController struct {
	indexer   Indexer
	status    IndexStatus
	statusMu  sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	runningMu sync.RWMutex
}

// NewIndexController creates a new index controller.
func NewIndexController(statePath string) *DefaultIndexController {
	ctx, cancel := context.WithCancel(context.Background())
	return &DefaultIndexController{
		indexer: NewIndexer(statePath),
		status: IndexStatus{
			IsIndexing: false,
			Phase:      "idle",
			Progress:   0,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins background indexing with given options.
func (c *DefaultIndexController) Start(ctx context.Context, opts IndexOptions) error {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if c.running {
		return fmt.Errorf("indexing is already running")
	}

	c.running = true
	c.updateStatus(IndexStatus{
		IsIndexing:     true,
		Phase:          "starting",
		Progress:       0,
		FilesProcessed: 0,
		TotalFiles:     0,
		ChunksCreated:  0,
		StartTime:      time.Now(),
		EstimatedEnd:   time.Time{},
		LastError:      "",
	})

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			c.runningMu.Lock()
			c.running = false
			c.runningMu.Unlock()
		}()

		// Perform indexing in background with a separate context
		indexCtx := context.Background()
		chunks, err := c.indexer.Index(indexCtx, opts)
		if err != nil {
			c.updateStatus(IndexStatus{
				IsIndexing: false,
				Phase:      "error",
				Progress:   0,
				LastError:  err.Error(),
			})
			return
		}

		// Store chunks if vector store and embedder are provided
		if opts.VectorStore != nil && opts.Embedder != nil {
			c.updateStatus(IndexStatus{
				IsIndexing: true,
				Phase:      "embedding",
				Progress:   80,
				LastError:  "",
			})

			docs := make([]vectorstore.Document, 0, len(chunks))
			for _, chunk := range chunks {
				// Generate embedding
				vec, err := opts.Embedder.Embed(ctx, chunk.Content)
				if err != nil {
					// Log error but continue with other chunks
					continue
				}

				// Convert metadata
				metadata := make(map[string]interface{})
				for k, v := range chunk.Metadata {
					metadata[k] = v
				}
				metadata["file_path"] = chunk.FilePath
				metadata["language"] = chunk.Language
				metadata["chunk_type"] = string(chunk.Type)
				metadata["start_line"] = chunk.StartLine
				metadata["end_line"] = chunk.EndLine
				metadata["indexed_at"] = chunk.IndexedAt.Format(time.RFC3339)

				docs = append(docs, vectorstore.Document{
					ID:        chunk.ID,
					Content:   chunk.Content,
					Vector:    vec.Vector,
					Metadata:  metadata,
					CreatedAt: chunk.IndexedAt,
					UpdatedAt: chunk.IndexedAt,
				})
			}

			// Store in batches
			if len(docs) > 0 {
				batchSize := 100
				for i := 0; i < len(docs); i += batchSize {
					end := i + batchSize
					if end > len(docs) {
						end = len(docs)
					}
					batch := docs[i:end]
					if err := opts.VectorStore.UpsertBatch(ctx, batch); err != nil {
						c.updateStatus(IndexStatus{
							IsIndexing: false,
							Phase:      "error",
							Progress:   0,
							LastError:  fmt.Sprintf("failed to store documents: %v", err),
						})
						return
					}
				}
			}
		}

		c.updateStatus(IndexStatus{
			IsIndexing:     false,
			Phase:          "completed",
			Progress:       100,
			FilesProcessed: len(chunks),
			TotalFiles:     len(chunks),
			ChunksCreated:  len(chunks),
			StartTime:      time.Now(),
			EstimatedEnd:   time.Now(),
			LastError:      "",
		})
	}()

	return nil
}

// Stop gracefully stops background indexing.
func (c *DefaultIndexController) Stop(ctx context.Context) error {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if !c.running {
		return nil
	}

	// Cancel the context
	c.cancel()

	// Wait for background goroutine to finish
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.updateStatus(IndexStatus{
			IsIndexing: false,
			Phase:      "stopped",
			Progress:   0,
			LastError:  "",
		})
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for indexing to stop")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ForceReindex performs a complete reindex of the codebase.
func (c *DefaultIndexController) ForceReindex(ctx context.Context, opts IndexOptions) error {
	// Clear any existing state
	if err := c.clearState(); err != nil {
		return fmt.Errorf("failed to clear state: %w", err)
	}

	// Start fresh indexing
	return c.Start(ctx, opts)
}

// ReindexPaths reindexes only the specified paths.
func (c *DefaultIndexController) ReindexPaths(ctx context.Context, opts IndexOptions, paths []string) error {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if c.running {
		return fmt.Errorf("indexing is already running")
	}

	c.running = true
	c.updateStatus(IndexStatus{
		IsIndexing:     true,
		Phase:          "reindexing_paths",
		Progress:       0,
		FilesProcessed: 0,
		TotalFiles:     len(paths),
		ChunksCreated:  0,
		StartTime:      time.Now(),
		EstimatedEnd:   time.Now().Add(time.Duration(len(paths)) * time.Second),
		LastError:      "",
	})

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer func() {
			c.runningMu.Lock()
			c.running = false
			c.runningMu.Unlock()
		}()

		var totalChunks int
		for i, path := range paths {
			select {
			case <-ctx.Done():
				c.updateStatus(IndexStatus{
					IsIndexing: false,
					Phase:      "stopped",
					Progress:   float64(i) / float64(len(paths)) * 100,
					LastError:  ctx.Err().Error(),
				})
				return
			default:
			}

			// Index single path
			pathOpts := opts
			pathOpts.RootPath = path

			chunks, err := c.indexer.Index(ctx, pathOpts)
			if err != nil {
				c.updateStatus(IndexStatus{
					IsIndexing: false,
					Phase:      "error",
					Progress:   float64(i) / float64(len(paths)) * 100,
					LastError:  fmt.Sprintf("failed to index %s: %v", path, err),
				})
				return
			}

			// Store chunks if vector store and embedder are provided
			if opts.VectorStore != nil && opts.Embedder != nil {
				for _, chunk := range chunks {
					vec, err := opts.Embedder.Embed(ctx, chunk.Content)
					if err != nil {
						continue
					}

					metadata := make(map[string]interface{})
					for k, v := range chunk.Metadata {
						metadata[k] = v
					}
					metadata["file_path"] = chunk.FilePath
					metadata["language"] = chunk.Language
					metadata["chunk_type"] = string(chunk.Type)
					metadata["start_line"] = chunk.StartLine
					metadata["end_line"] = chunk.EndLine
					metadata["indexed_at"] = chunk.IndexedAt.Format(time.RFC3339)

					doc := vectorstore.Document{
						ID:        chunk.ID,
						Content:   chunk.Content,
						Vector:    vec.Vector,
						Metadata:  metadata,
						CreatedAt: chunk.IndexedAt,
						UpdatedAt: chunk.IndexedAt,
					}

					if err := opts.VectorStore.Upsert(ctx, doc); err != nil {
						// Log error but continue
						continue
					}
				}
			}

			totalChunks += len(chunks)
			progress := float64(i+1) / float64(len(paths)) * 100

			c.updateStatus(IndexStatus{
				IsIndexing:     true,
				Phase:          "reindexing_paths",
				Progress:       progress,
				FilesProcessed: i + 1,
				TotalFiles:     len(paths),
				ChunksCreated:  totalChunks,
				StartTime:      time.Now(),
				EstimatedEnd:   time.Now().Add(time.Duration(len(paths)-i-1) * time.Second),
				LastError:      "",
			})
		}

		c.updateStatus(IndexStatus{
			IsIndexing:     false,
			Phase:          "completed",
			Progress:       100,
			FilesProcessed: len(paths),
			TotalFiles:     len(paths),
			ChunksCreated:  totalChunks,
			StartTime:      time.Now(),
			EstimatedEnd:   time.Now(),
			LastError:      "",
		})
	}()

	return nil
}

// GetStatus returns current indexing status.
func (c *DefaultIndexController) GetStatus() IndexStatus {
	c.statusMu.RLock()
	defer c.statusMu.RUnlock()
	return c.status
}

// HealthCheck performs health validation of the index.
func (c *DefaultIndexController) HealthCheck(ctx context.Context) error {
	status := c.GetStatus()

	// Check if stuck in indexing state for too long
	if status.IsIndexing && !status.StartTime.IsZero() {
		if time.Since(status.StartTime) > 30*time.Minute {
			return fmt.Errorf("indexing appears stuck (running for %v)", time.Since(status.StartTime))
		}
	}

	// Check for repeated errors
	if status.LastError != "" && status.Phase == "error" {
		return fmt.Errorf("indexing in error state: %s", status.LastError)
	}

	return nil
}

// updateStatus safely updates the internal status.
func (c *DefaultIndexController) updateStatus(status IndexStatus) {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()
	c.status = status
}

// clearState removes any persisted indexing state.
func (c *DefaultIndexController) clearState() error {
	// This would clear any persisted state files
	// For now, just return nil as the indexer handles this internally
	return nil
}
