package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore/sqlite"
)

// BenchmarkFileWalking benchmarks the file system traversal performance.
func BenchmarkFileWalking(b *testing.B) {
	testCases := []struct {
		name      string
		numFiles  int
		fileSize  int // bytes per file
		numDirs   int
	}{
		{"1K_Files_Small", 1000, 1024, 10},        // 1KB files
		{"1K_Files_Medium", 1000, 10240, 10},      // 10KB files
		{"10K_Files_Small", 10000, 1024, 100},     // 1KB files
		{"10K_Files_Medium", 10000, 10240, 100},   // 10KB files
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create test directory with files
			tmpDir := b.TempDir()
			setupTestFiles(b, tmpDir, tc.numFiles, tc.fileSize, tc.numDirs)

			walker := NewFileWalker(0) // No size limit
			ctx := context.Background()

			// Measure
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				fileCount := 0
				err := walker.Walk(ctx, tmpDir, DefaultIgnorePatterns(), func(path string, info os.FileInfo) error {
					fileCount++
					return nil
				})
				
				if err != nil {
					b.Fatalf("Walk failed: %v", err)
				}
				
				if fileCount != tc.numFiles {
					b.Fatalf("Expected %d files, got %d", tc.numFiles, fileCount)
				}
			}

			// Report custom metrics
			filesPerSec := float64(tc.numFiles) / b.Elapsed().Seconds() * float64(b.N)
			b.ReportMetric(filesPerSec, "files/sec")
		})
	}
}

// BenchmarkChunking benchmarks the text chunking performance.
func BenchmarkChunking(b *testing.B) {
	testCases := []struct {
		name      string
		numFiles  int
		fileSize  int
	}{
		{"100_Files_1KB", 100, 1024},
		{"100_Files_10KB", 100, 10240},
		{"1K_Files_1KB", 1000, 1024},
		{"1K_Files_10KB", 1000, 10240},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create test files
			tmpDir := b.TempDir()
			setupTestFiles(b, tmpDir, tc.numFiles, tc.fileSize, 10)

			// Collect file paths
			var files []string
			filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && filepath.Ext(path) == ".go" {
					files = append(files, path)
				}
				return nil
			})

			// Measure
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				totalChunks := 0
				for _, file := range files {
					content, err := os.ReadFile(file)
					if err != nil {
						b.Fatalf("Failed to read file: %v", err)
					}
					
					// Simulate chunking (split by 1024 bytes)
					chunkSize := 1024
					for offset := 0; offset < len(content); offset += chunkSize {
						end := offset + chunkSize
						if end > len(content) {
							end = len(content)
						}
						_ = content[offset:end] // Process chunk
						totalChunks++
					}
				}
				
				if totalChunks == 0 {
					b.Fatal("No chunks created")
				}
			}

			// Report custom metrics
			filesPerSec := float64(tc.numFiles) / b.Elapsed().Seconds() * float64(b.N)
			b.ReportMetric(filesPerSec, "files/sec")
		})
	}
}

// BenchmarkMerkleTreeHashing benchmarks the merkle tree hash computation.
func BenchmarkMerkleTreeHashing(b *testing.B) {
	testCases := []struct {
		name     string
		numFiles int
	}{
		{"1K_Files", 1000},
		{"5K_Files", 5000},
		{"10K_Files", 10000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create test directory
			tmpDir := b.TempDir()
			setupTestFiles(b, tmpDir, tc.numFiles, 1024, 50)

			ctx := context.Background()
			walker := NewFileWalker(0)
			merkle := NewMerkleTree(walker)

			// Measure
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := merkle.Hash(ctx, tmpDir, DefaultIgnorePatterns())
				if err != nil {
					b.Fatalf("Hash failed: %v", err)
				}
			}

			// Report custom metrics
			filesPerSec := float64(tc.numFiles) / b.Elapsed().Seconds() * float64(b.N)
			b.ReportMetric(filesPerSec, "files/sec")
		})
	}
}

// BenchmarkMerkleTreeDiff benchmarks the change detection performance.
func BenchmarkMerkleTreeDiff(b *testing.B) {
	testCases := []struct {
		name        string
		numFiles    int
		changeRate  float64 // percentage of files changed
	}{
		{"1K_Files_1pct", 1000, 0.01},
		{"1K_Files_5pct", 1000, 0.05},
		{"1K_Files_10pct", 1000, 0.10},
		{"10K_Files_1pct", 10000, 0.01},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create test directory and initial tree
			tmpDir := b.TempDir()
			setupTestFiles(b, tmpDir, tc.numFiles, 1024, 50)

			ctx := context.Background()
			walker := NewFileWalker(0)
			merkle := NewMerkleTree(walker)
			
			oldState, err := merkle.Hash(ctx, tmpDir, DefaultIgnorePatterns())
			if err != nil {
				b.Fatalf("Hash failed: %v", err)
			}

			// Modify some files
			numChanged := int(float64(tc.numFiles) * tc.changeRate)
			modifyFiles(b, tmpDir, numChanged)

			// Measure
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				newState, err := merkle.Hash(ctx, tmpDir, DefaultIgnorePatterns())
				if err != nil {
					b.Fatalf("Hash failed: %v", err)
				}

				changed, err := merkle.Diff(ctx, oldState, newState)
				if err != nil {
					b.Fatalf("Diff failed: %v", err)
				}
				
				if len(changed) == 0 {
					b.Fatal("No changes detected")
				}
			}

			// Report custom metrics
			b.ReportMetric(float64(numChanged), "files_changed")
		})
	}
}

// BenchmarkIncrementalIndexing benchmarks incremental indexing performance.
func BenchmarkIncrementalIndexing(b *testing.B) {
	testCases := []struct {
		name       string
		numFiles   int
		changeRate float64
	}{
		{"1K_Files_1pct", 1000, 0.01},
		{"1K_Files_5pct", 1000, 0.05},
		{"1K_Files_10pct", 1000, 0.10},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create test directory and do initial index
			tmpDir := b.TempDir()
			setupTestFiles(b, tmpDir, tc.numFiles, 1024, 10)

			stateDir := filepath.Join(tmpDir, ".conexus-state")
			indexer := NewIndexer(stateDir)
			
			ctx := context.Background()
			opts := IndexOptions{
				RootPath:       tmpDir,
				IgnorePatterns: DefaultIgnorePatterns(),
				MaxFileSize:    1024 * 1024,
			}

			// Do initial full index
			_, err := indexer.Index(ctx, opts)
			if err != nil {
				b.Fatalf("Initial index failed: %v", err)
			}

			// Modify files before benchmark
			numChanged := int(float64(tc.numFiles) * tc.changeRate)
			modifyFiles(b, tmpDir, numChanged)

			// Measure incremental indexing
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				chunks, err := indexer.Index(ctx, opts)
				if err != nil {
					b.Fatalf("Incremental index failed: %v", err)
				}
				
				// Should only process changed files
				if len(chunks) == 0 {
					b.Fatal("No chunks returned from incremental index")
				}
			}

			// Report custom metrics
			msPerFile := b.Elapsed().Milliseconds() / int64(numChanged) / int64(b.N)
			b.ReportMetric(float64(msPerFile), "ms/file")
		})
	}
}

// BenchmarkFullIndexWithEmbeddings benchmarks end-to-end indexing with embeddings.
func BenchmarkFullIndexWithEmbeddings(b *testing.B) {
	testCases := []struct {
		name     string
		numFiles int
	}{
		{"100_Files", 100},
		{"1K_Files", 1000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create test directory and dependencies
			tmpDir := b.TempDir()
			setupTestFiles(b, tmpDir, tc.numFiles, 1024, 10)

			// Create mock embedder
			mockEmbedder := &mockEmbedder{dimension: 384}

			// Create in-memory vector store
			dbPath := filepath.Join(b.TempDir(), "test.db")
			store, err := sqlite.NewStore(dbPath)
			if err != nil {
				b.Fatalf("Failed to create store: %v", err)
			}
			defer store.Close()

			indexer := NewIndexer(filepath.Join(tmpDir, ".conexus-state"))
			ctx := context.Background()
			opts := IndexOptions{
				RootPath:       tmpDir,
				IgnorePatterns: DefaultIgnorePatterns(),
				MaxFileSize:    1024 * 1024,
				Embedder:       mockEmbedder,
				VectorStore:    store,
			}

			// Measure
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				chunks, err := indexer.Index(ctx, opts)
				if err != nil {
					b.Fatalf("Index failed: %v", err)
				}
				
				if len(chunks) == 0 {
					b.Fatal("No chunks created")
				}
			}

			// Report custom metrics
			filesPerSec := float64(tc.numFiles) / b.Elapsed().Seconds() * float64(b.N)
			b.ReportMetric(filesPerSec, "files/sec")
		})
	}
}

// BenchmarkConcurrentIndexing benchmarks parallel directory indexing.
func BenchmarkConcurrentIndexing(b *testing.B) {
	testCases := []struct {
		name        string
		numDirs     int
		filesPerDir int
	}{
		{"10_Dirs_100_Files", 10, 100},
		{"50_Dirs_100_Files", 50, 100},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create multiple directories
			tmpDir := b.TempDir()
			totalFiles := tc.numDirs * tc.filesPerDir
			setupTestFiles(b, tmpDir, totalFiles, 1024, tc.numDirs)

			indexer := NewIndexer(filepath.Join(tmpDir, ".conexus-state"))
			ctx := context.Background()
			opts := IndexOptions{
				RootPath:       tmpDir,
				IgnorePatterns: DefaultIgnorePatterns(),
				MaxFileSize:    1024 * 1024,
			}

			// Measure
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				chunks, err := indexer.Index(ctx, opts)
				if err != nil {
					b.Fatalf("Index failed: %v", err)
				}
				
				if len(chunks) == 0 {
					b.Fatal("No chunks created")
				}
			}

			// Report custom metrics
			filesPerSec := float64(totalFiles) / b.Elapsed().Seconds() * float64(b.N)
			b.ReportMetric(filesPerSec, "files/sec")
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory consumption during indexing.
func BenchmarkMemoryUsage(b *testing.B) {
	testCases := []struct {
		name     string
		numFiles int
	}{
		{"1K_Files", 1000},
		{"5K_Files", 5000},
		{"10K_Files", 10000},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Setup: Create test directory
			tmpDir := b.TempDir()
			setupTestFiles(b, tmpDir, tc.numFiles, 1024, 50)

			indexer := NewIndexer(filepath.Join(tmpDir, ".conexus-state"))
			ctx := context.Background()
			opts := IndexOptions{
				RootPath:       tmpDir,
				IgnorePatterns: DefaultIgnorePatterns(),
				MaxFileSize:    1024 * 1024,
			}

			// Measure
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := indexer.Index(ctx, opts)
				if err != nil {
					b.Fatalf("Index failed: %v", err)
				}
			}
		})
	}
}

// Helper: setupTestFiles creates a directory structure with test files.
func setupTestFiles(tb testing.TB, rootDir string, numFiles, fileSize, numDirs int) {
	tb.Helper()

	filesPerDir := numFiles / numDirs
	if filesPerDir == 0 {
		filesPerDir = 1
	}

	for dir := 0; dir < numDirs; dir++ {
		dirPath := filepath.Join(rootDir, fmt.Sprintf("dir%d", dir))
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			tb.Fatalf("Failed to create directory: %v", err)
		}

		for file := 0; file < filesPerDir; file++ {
			filePath := filepath.Join(dirPath, fmt.Sprintf("file%d.go", file))
			content := generateTestContent(fileSize, dir*filesPerDir+file)
			
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				tb.Fatalf("Failed to write file: %v", err)
			}
		}
	}
}

// Helper: generateTestContent creates realistic Go code content.
func generateTestContent(size, seed int) string {
	template := `package test

import (
	"context"
	"fmt"
	"time"
)

// Function%d is an auto-generated test function.
func Function%d(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty input")
	}
	
	result := fmt.Sprintf("processed: %%s (seed: %%d, time: %%v)", input, %d, time.Now())
	return result, nil
}

// Helper%d provides utility functionality
func Helper%d(data []byte) string {
	return fmt.Sprintf("helper-%%d: %%d bytes", %d, len(data))
}
`
	content := fmt.Sprintf(template, seed, seed, seed, seed, seed, seed)
	
	// Pad to desired size
	for len(content) < size {
		content += fmt.Sprintf("\n// Padding line to reach size: %d\n", len(content))
	}
	
	return content[:size]
}

// Helper: modifyFiles modifies random files in a directory.
func modifyFiles(tb testing.TB, rootDir string, numFiles int) {
	tb.Helper()

	// Walk directory and collect file paths
	var files []string
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".go" {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		tb.Fatal("No files to modify")
	}

	// Modify first numFiles files
	for i := 0; i < numFiles && i < len(files); i++ {
		content, err := os.ReadFile(files[i])
		if err != nil {
			tb.Fatalf("Failed to read file: %v", err)
		}

		// Append a comment to modify the file
		modified := string(content) + fmt.Sprintf("\n// Modified at: %v\n", time.Now())
		
		if err := os.WriteFile(files[i], []byte(modified), 0644); err != nil {
			tb.Fatalf("Failed to write file: %v", err)
		}
	}
}

// mockEmbedder implements embedding.Embedder for testing.
type mockEmbedder struct {
	dimension int
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) (*embedding.Embedding, error) {
	// Generate deterministic embedding from text hash
	hash := sha256.Sum256([]byte(text))
	vector := make(embedding.Vector, m.dimension)
	
	for i := 0; i < m.dimension; i++ {
		vector[i] = float32(hash[i%32]) / 255.0
	}
	
	return &embedding.Embedding{
		Text:   text,
		Vector: vector,
		Model:  "mock-model",
	}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([]*embedding.Embedding, error) {
	results := make([]*embedding.Embedding, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		results[i] = emb
	}
	return results, nil
}

func (m *mockEmbedder) Dimensions() int {
	return m.dimension
}

func (m *mockEmbedder) Model() string {
	return "mock-model"
}

// Ensure mockEmbedder implements embedding.Embedder
var _ embedding.Embedder = (*mockEmbedder)(nil)
