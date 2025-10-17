// Package sqlite provides benchmarks for SQLite-backed vector store operations.
package sqlite

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime"
	"testing"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// BenchmarkVectorSearch benchmarks vector search performance across different dataset sizes.
func BenchmarkVectorSearch(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Docs_%d", size), func(b *testing.B) {
			store, cleanup := setupBenchmarkStore(b, size)
			defer cleanup()
			
			// Generate query vector
			queryVector := generateVector(384)
			ctx := context.Background()
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				opts := vectorstore.SearchOptions{Limit: 10}
				_, err := store.SearchVector(ctx, queryVector, opts)
				if err != nil {
					b.Fatalf("search failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkBM25Search benchmarks BM25 full-text search performance.
func BenchmarkBM25Search(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	queries := []string{
		"function implementation",
		"error handling context",
		"database query optimization",
	}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Docs_%d", size), func(b *testing.B) {
			store, cleanup := setupBenchmarkStore(b, size)
			defer cleanup()
			
			ctx := context.Background()
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				query := queries[i%len(queries)]
				opts := vectorstore.SearchOptions{Limit: 10}
				_, err := store.SearchBM25(ctx, query, opts)
				if err != nil {
					b.Fatalf("search failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkHybridSearch benchmarks hybrid search combining vector and BM25.
func BenchmarkHybridSearch(b *testing.B) {
	sizes := []int{100, 1000, 10000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Docs_%d", size), func(b *testing.B) {
			store, cleanup := setupBenchmarkStore(b, size)
			defer cleanup()
			
			queryVector := generateVector(384)
			query := "function implementation error handling"
			ctx := context.Background()
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				opts := vectorstore.SearchOptions{Limit: 10}
				_, err := store.SearchHybrid(ctx, query, queryVector, opts)
				if err != nil {
					b.Fatalf("search failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkInsert benchmarks document insertion performance.
func BenchmarkInsert(b *testing.B) {
	store, err := NewStore(":memory:")
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()
	
	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		doc := vectorstore.Document{
			ID:      fmt.Sprintf("doc_%d", i),
			Content: fmt.Sprintf("This is test document %d with some content for benchmarking.", i),
			Vector:  generateVector(384),
			Metadata: map[string]interface{}{
				"path": fmt.Sprintf("/test/file_%d.go", i),
				"type": "code",
			},
		}
		
		if err := store.Upsert(ctx, doc); err != nil {
			b.Fatalf("insert failed: %v", err)
		}
	}
}

// BenchmarkBatchInsert benchmarks batch insertion performance.
func BenchmarkBatchInsert(b *testing.B) {
	batchSizes := []int{10, 100, 1000}
	
	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			store, err := NewStore(":memory:")
			if err != nil {
				b.Fatalf("failed to create store: %v", err)
			}
			defer store.Close()
			
			ctx := context.Background()
			
			b.ResetTimer()
			b.ReportAllocs()
			
			for i := 0; i < b.N; i++ {
				docs := make([]vectorstore.Document, batchSize)
				for j := 0; j < batchSize; j++ {
					docs[j] = vectorstore.Document{
						ID:      fmt.Sprintf("doc_%d_%d", i, j),
						Content: fmt.Sprintf("Batch document %d-%d content", i, j),
						Vector:  generateVector(384),
						Metadata: map[string]interface{}{
							"batch": i,
							"index": j,
						},
					}
				}
				
				if err := store.UpsertBatch(ctx, docs); err != nil {
					b.Fatalf("batch insert failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkUpdate benchmarks document update performance.
func BenchmarkUpdate(b *testing.B) {
	store, cleanup := setupBenchmarkStore(b, 1000)
	defer cleanup()
	
	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		docID := fmt.Sprintf("doc_%d", i%1000)
		doc := vectorstore.Document{
			ID:      docID,
			Content: fmt.Sprintf("Updated content %d", i),
			Vector:  generateVector(384),
			Metadata: map[string]interface{}{
				"updated": true,
				"version": i,
			},
		}
		
		if err := store.Upsert(ctx, doc); err != nil {
			b.Fatalf("update failed: %v", err)
		}
	}
}

// BenchmarkDelete benchmarks document deletion performance.
func BenchmarkDelete(b *testing.B) {
	// Setup with more docs than iterations
	store, cleanup := setupBenchmarkStore(b, b.N+1000)
	defer cleanup()
	
	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		docID := fmt.Sprintf("doc_%d", i)
		if err := store.Delete(ctx, docID); err != nil {
			b.Fatalf("delete failed: %v", err)
		}
	}
}

// BenchmarkConcurrentSearch benchmarks concurrent search operations.
func BenchmarkConcurrentSearch(b *testing.B) {
	store, cleanup := setupBenchmarkStore(b, 10000)
	defer cleanup()
	
	ctx := context.Background()
	queryVector := generateVector(384)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			opts := vectorstore.SearchOptions{Limit: 10}
			_, err := store.SearchVector(ctx, queryVector, opts)
			if err != nil {
				b.Fatalf("search failed: %v", err)
			}
		}
	})
}

// BenchmarkMemoryUsage measures memory usage for different dataset sizes.
func BenchmarkMemoryUsage(b *testing.B) {
	sizes := []int{1000, 5000, 10000, 50000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Docs_%d", size), func(b *testing.B) {
			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)
			
			_, cleanup := setupBenchmarkStore(b, size)
			defer cleanup()
			
			runtime.GC()
			runtime.ReadMemStats(&m2)
			
			allocMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
			b.ReportMetric(allocMB, "MB/docs")
			b.ReportMetric(allocMB/float64(size)*1000, "KB/1k-docs")
		})
	}
}

// setupBenchmarkStore creates a store with the specified number of documents.
func setupBenchmarkStore(b *testing.B, numDocs int) (*Store, func()) {
	b.Helper()
	
	// Create temp directory for benchmark database
	tmpDir, err := os.MkdirTemp("", "benchmark_*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	
	dbPath := fmt.Sprintf("%s/benchmark.db", tmpDir)
	store, err := NewStore(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		b.Fatalf("failed to create store: %v", err)
	}
	
	// Populate with test data
	ctx := context.Background()
	docs := make([]vectorstore.Document, numDocs)
	
	for i := 0; i < numDocs; i++ {
		docs[i] = vectorstore.Document{
			ID:      fmt.Sprintf("doc_%d", i),
			Content: generateContent(i),
			Vector:  generateVector(384),
			Metadata: map[string]interface{}{
				"path":  fmt.Sprintf("/src/pkg%d/file%d.go", i/100, i%100),
				"type":  "code",
				"lines": rand.Intn(500) + 10,
			},
		}
	}
	
	// Batch insert for faster setup
	batchSize := 100
	for i := 0; i < len(docs); i += batchSize {
		end := i + batchSize
		if end > len(docs) {
			end = len(docs)
		}
		if err := store.UpsertBatch(ctx, docs[i:end]); err != nil {
			store.Close()
			os.RemoveAll(tmpDir)
			b.Fatalf("failed to populate store: %v", err)
		}
	}
	
	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}
	
	return store, cleanup
}

// generateVector creates a random embedding vector of the specified dimension.
func generateVector(dim int) embedding.Vector {
	vec := make(embedding.Vector, dim)
	for i := range vec {
		vec[i] = rand.Float32()*2 - 1 // Range [-1, 1]
	}
	
	// Normalize
	var sum float32
	for _, v := range vec {
		sum += v * v
	}
	norm := float32(1.0 / math.Sqrt(float64(sum)))
	for i := range vec {
		vec[i] *= norm
	}
	
	return vec
}

// generateContent creates realistic test content for benchmarking.
func generateContent(index int) string {
	templates := []string{
		"func Process%d(ctx context.Context, input string) (string, error) {\n\treturn fmt.Sprintf(\"processed: %%s\", input), nil\n}",
		"type Handler%d struct {\n\tdb *sql.DB\n\tcache *Cache\n}\n\nfunc (h *Handler%d) Handle(req Request) Response {\n\treturn Response{Status: 200}\n}",
		"// Package%d provides utilities for data processing\npackage pkg%d\n\nimport (\n\t\"context\"\n\t\"fmt\"\n)",
		"const (\n\tMaxRetries%d = 3\n\tTimeout%d = 30 * time.Second\n\tBufferSize%d = 1024\n)",
	}
	
	template := templates[index%len(templates)]
	return fmt.Sprintf(template, index, index, index)
}
