// Package search provides tests for search functionality.
package search

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
)

func TestSearchCache_Basic(t *testing.T) {
	cache := NewSearchCache(10, time.Minute)

	// Test cache miss
	query := "test query"
	filters := map[string]interface{}{"source_type": "file"}
	_, found := cache.Get(query, filters)
	assert.False(t, found)

	// Test cache set and get
	results := []vectorstore.SearchResult{
		{Score: 0.9},
		{Score: 0.8},
	}
	cache.Set(query, filters, results, 0.1)

	cached, found := cache.Get(query, filters)
	assert.True(t, found)
	assert.Equal(t, query, cached.Query)
	assert.Equal(t, results, cached.Results)
	assert.Equal(t, 0.1, cached.QueryTime)
}

func TestSearchCache_Size(t *testing.T) {
	cache := NewSearchCache(2, time.Minute)

	// Add entries
	cache.Set("query1", nil, nil, 0.1)
	cache.Set("query2", nil, nil, 0.1)
	cache.Set("query3", nil, nil, 0.1) // Should evict oldest

	assert.Equal(t, 2, cache.Size())

	// Oldest should be gone
	_, found := cache.Get("query1", nil)
	assert.False(t, found)

	// Newer ones should be there
	_, found = cache.Get("query2", nil)
	assert.True(t, found)
	_, found = cache.Get("query3", nil)
	assert.True(t, found)
}

func TestSearchCache_TTL(t *testing.T) {
	cache := NewSearchCache(10, time.Millisecond*10)

	// Add entry
	cache.Set("query", nil, nil, 0.1)

	// Should be available immediately
	_, found := cache.Get("query", nil)
	assert.True(t, found)

	// Wait for TTL to expire
	time.Sleep(time.Millisecond * 15)

	// Should be gone
	_, found = cache.Get("query", nil)
	assert.False(t, found)
}

func TestSearchCache_Clear(t *testing.T) {
	cache := NewSearchCache(10, time.Minute)

	cache.Set("query1", nil, nil, 0.1)
	cache.Set("query2", nil, nil, 0.1)

	assert.Equal(t, 2, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())
}

// Mock implementations for testing
type mockVectorStore struct {
	searchBM25Func   func(ctx context.Context, query string, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error)
	searchVectorFunc func(ctx context.Context, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error)
}

func (m *mockVectorStore) SearchBM25(ctx context.Context, query string, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	if m.searchBM25Func != nil {
		return m.searchBM25Func(ctx, query, opts)
	}
	return nil, nil
}

func (m *mockVectorStore) SearchVector(ctx context.Context, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	if m.searchVectorFunc != nil {
		return m.searchVectorFunc(ctx, vector, opts)
	}
	return nil, nil
}

func (m *mockVectorStore) SearchHybrid(ctx context.Context, query string, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
	return nil, nil
}

func (m *mockVectorStore) Upsert(ctx context.Context, doc vectorstore.Document) error {
	return nil
}

func (m *mockVectorStore) UpsertBatch(ctx context.Context, docs []vectorstore.Document) error {
	return nil
}

func (m *mockVectorStore) Get(ctx context.Context, id string) (*vectorstore.Document, error) {
	return nil, nil
}

func (m *mockVectorStore) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockVectorStore) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func (m *mockVectorStore) ListIndexedFiles(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (m *mockVectorStore) GetFileChunks(ctx context.Context, filePath string) ([]vectorstore.Document, error) {
	return nil, nil
}

func (m *mockVectorStore) Close() error {
	return nil
}

type mockEmbedder struct {
	embedFunc func(ctx context.Context, text string) (*embedding.Embedding, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) (*embedding.Embedding, error) {
	if m.embedFunc != nil {
		return m.embedFunc(ctx, text)
	}
	return &embedding.Embedding{
		Text:   text,
		Vector: make(embedding.Vector, 384),
	}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([]*embedding.Embedding, error) {
	result := make([]*embedding.Embedding, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		result[i] = emb
	}
	return result, nil
}

func (m *mockEmbedder) Dimensions() int {
	return 384
}

func (m *mockEmbedder) Model() string {
	return "mock-model"
}

type mockFusion struct {
	fuseFunc func(ctx context.Context, sparseResults, denseResults []vectorstore.SearchResult, mode HybridMode) ([]Result, error)
}

func (m *mockFusion) Fuse(ctx context.Context, sparseResults, denseResults []vectorstore.SearchResult, mode HybridMode) ([]Result, error) {
	if m.fuseFunc != nil {
		return m.fuseFunc(ctx, sparseResults, denseResults, mode)
	}

	// Simple fusion: combine both result sets
	var results []Result
	for _, r := range sparseResults {
		results = append(results, Result{
			Document:    r.Document,
			Score:       r.Score,
			SparseScore: r.Score,
		})
	}
	for _, r := range denseResults {
		results = append(results, Result{
			Document:   r.Document,
			Score:      r.Score,
			DenseScore: r.Score,
		})
	}
	return results, nil
}

type mockReranker struct {
	rerankFunc func(ctx context.Context, query string, results []Result) ([]Result, error)
}

func (m *mockReranker) Rerank(ctx context.Context, query string, results []Result) ([]Result, error) {
	if m.rerankFunc != nil {
		return m.rerankFunc(ctx, query, results)
	}
	// Simple reranking: just return results as-is
	return results, nil
}

func TestNewPipeline(t *testing.T) {
	store := &mockVectorStore{}
	embedder := &mockEmbedder{}
	fusion := &mockFusion{}
	reranker := &mockReranker{}

	pipeline := NewPipeline(store, embedder, fusion, reranker)

	assert.NotNil(t, pipeline)
	assert.Equal(t, store, pipeline.Store)
	assert.Equal(t, embedder, pipeline.Embedder)
	assert.Equal(t, fusion, pipeline.Fusion)
	assert.Equal(t, reranker, pipeline.Reranker)
}

func TestNewPipeline_WithoutReranker(t *testing.T) {
	store := &mockVectorStore{}
	embedder := &mockEmbedder{}
	fusion := &mockFusion{}

	pipeline := NewPipeline(store, embedder, fusion, nil)

	assert.NotNil(t, pipeline)
	assert.Equal(t, store, pipeline.Store)
	assert.Equal(t, embedder, pipeline.Embedder)
	assert.Equal(t, fusion, pipeline.Fusion)
	assert.Nil(t, pipeline.Reranker)
}

func TestPipeline_Search_DenseMode(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	store := &mockVectorStore{
		searchVectorFunc: func(ctx context.Context, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
			return []vectorstore.SearchResult{
				{
					Document: vectorstore.Document{ID: "doc1", Content: "test content"},
					Score:    0.9,
				},
			}, nil
		},
	}

	embedder := &mockEmbedder{}
	fusion := &mockFusion{}

	pipeline := NewPipeline(store, embedder, fusion, nil)

	// Execute search
	query := Query{
		Text:       "test query",
		Limit:      10,
		Threshold:  0.5,
		HybridMode: HybridModeDense,
	}

	results, err := pipeline.Search(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)
	assert.Equal(t, float32(0.9), results[0].Score)
	assert.Equal(t, float32(0.9), results[0].DenseScore)
}

func TestPipeline_Search_SparseMode(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	store := &mockVectorStore{
		searchBM25Func: func(ctx context.Context, query string, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
			return []vectorstore.SearchResult{
				{
					Document: vectorstore.Document{ID: "doc2", Content: "test content"},
					Score:    0.8,
				},
			}, nil
		},
	}

	embedder := &mockEmbedder{}
	fusion := &mockFusion{}

	pipeline := NewPipeline(store, embedder, fusion, nil)

	// Execute search
	query := Query{
		Text:       "test query",
		Limit:      10,
		Threshold:  0.5,
		HybridMode: HybridModeSparse,
	}

	results, err := pipeline.Search(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc2", results[0].Document.ID)
	assert.Equal(t, float32(0.8), results[0].Score)
	assert.Equal(t, float32(0.8), results[0].SparseScore)
}

func TestPipeline_Search_HybridMode(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	store := &mockVectorStore{
		searchBM25Func: func(ctx context.Context, query string, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
			return []vectorstore.SearchResult{
				{
					Document: vectorstore.Document{ID: "doc1", Content: "sparse result"},
					Score:    0.8,
				},
			}, nil
		},
		searchVectorFunc: func(ctx context.Context, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
			return []vectorstore.SearchResult{
				{
					Document: vectorstore.Document{ID: "doc2", Content: "dense result"},
					Score:    0.9,
				},
			}, nil
		},
	}

	embedder := &mockEmbedder{}
	fusion := &mockFusion{}

	pipeline := NewPipeline(store, embedder, fusion, nil)

	// Execute search
	query := Query{
		Text:       "test query",
		Limit:      10,
		Threshold:  0.5,
		HybridMode: HybridModeRRF,
	}

	results, err := pipeline.Search(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, results, 2) // Should get both sparse and dense results
}

func TestPipeline_Search_WithReranker(t *testing.T) {
	ctx := context.Background()

	// Setup mocks
	store := &mockVectorStore{
		searchVectorFunc: func(ctx context.Context, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
			return []vectorstore.SearchResult{
				{
					Document: vectorstore.Document{ID: "doc1", Content: "test content"},
					Score:    0.9,
				},
			}, nil
		},
	}

	embedder := &mockEmbedder{}
	fusion := &mockFusion{}
	reranker := &mockReranker{
		rerankFunc: func(ctx context.Context, query string, results []Result) ([]Result, error) {
			// Modify scores to simulate reranking
			for i := range results {
				results[i].Score *= 0.5
				results[i].RerankedFrom = i
			}
			return results, nil
		},
	}

	pipeline := NewPipeline(store, embedder, fusion, reranker)

	// Execute search
	query := Query{
		Text:       "test query",
		Limit:      10,
		Threshold:  0.5,
		HybridMode: HybridModeDense,
	}

	results, err := pipeline.Search(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "doc1", results[0].Document.ID)
	assert.Equal(t, float32(0.45), results[0].Score) // 0.9 * 0.5
	assert.Equal(t, 0, results[0].RerankedFrom)      // Original rank was 0
}

func TestPipeline_Search_EmbeddingError(t *testing.T) {
	ctx := context.Background()

	// Setup mock embedder that returns error
	embedder := &mockEmbedder{
		embedFunc: func(ctx context.Context, text string) (*embedding.Embedding, error) {
			return nil, assert.AnError
		},
	}

	store := &mockVectorStore{}
	fusion := &mockFusion{}

	pipeline := NewPipeline(store, embedder, fusion, nil)

	// Execute search
	query := Query{
		Text:       "test query",
		Limit:      10,
		HybridMode: HybridModeDense,
	}

	_, err := pipeline.Search(ctx, query)

	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestPipeline_Search_LimitTrimming(t *testing.T) {
	ctx := context.Background()

	// Setup mocks that return more results than limit
	store := &mockVectorStore{
		searchVectorFunc: func(ctx context.Context, vector embedding.Vector, opts vectorstore.SearchOptions) ([]vectorstore.SearchResult, error) {
			// Return 5 results (pipeline will request limit*2 = 4)
			return []vectorstore.SearchResult{
				{Document: vectorstore.Document{ID: "doc1"}, Score: 0.9},
				{Document: vectorstore.Document{ID: "doc2"}, Score: 0.8},
				{Document: vectorstore.Document{ID: "doc3"}, Score: 0.7},
				{Document: vectorstore.Document{ID: "doc4"}, Score: 0.6},
				{Document: vectorstore.Document{ID: "doc5"}, Score: 0.5},
			}, nil
		},
	}

	embedder := &mockEmbedder{}
	fusion := &mockFusion{}

	pipeline := NewPipeline(store, embedder, fusion, nil)

	// Execute search with limit of 2
	query := Query{
		Text:       "test query",
		Limit:      2,
		HybridMode: HybridModeDense,
	}

	results, err := pipeline.Search(ctx, query)

	assert.NoError(t, err)
	assert.Len(t, results, 2) // Should be trimmed to limit
}
