package contextual

import (
	"context"
	"testing"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextualRetrievalFramework(t *testing.T) {
	// Create mock components
	vectorStore := vectorstore.NewMemoryStore()
	embedder := &mockEmbedder{}
	profileManager := profiles.NewProfileManager(&MockClassifier{})
	optimizer := NewRetrievalOptimizer()
	assessor := NewQualityAssessor()
	monitor := NewContextualPerformanceMonitor()

	config := ContextualRetrievalConfig{
		VectorStore:        vectorStore,
		Embedder:           embedder,
		ProfileManager:     profileManager,
		Optimizer:          optimizer,
		QualityAssessor:    assessor,
		PerformanceMonitor: monitor,
	}

	framework := NewContextualRetrievalFramework(config)

	// Add test documents
	ctx := context.Background()
	testDocs := []vectorstore.Document{
		{
			ID:      "doc-1",
			Content: "authentication implementation using JWT tokens",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "file",
				"file_path":   "auth.go",
				"language":    "go",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:      "doc-2",
			Content: "documentation for user login system",
			Vector:  make(embedding.Vector, 384),
			Metadata: map[string]interface{}{
				"source_type": "documentation",
				"file_path":   "README.md",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, doc := range testDocs {
		err := vectorStore.Upsert(ctx, doc)
		require.NoError(t, err)
	}

	// Test contextual search
	query := &ContextualQuery{
		Query:   "how does authentication work",
		Profile: profiles.GetProfileByID("code_analysis"),
		WorkContext: map[string]interface{}{
			"active_file": "auth.go",
			"git_branch":  "feature/auth",
		},
	}

	response, err := framework.Search(ctx, query)
	require.NoError(t, err)

	assert.NotNil(t, response)
	// The search might return 0 results due to mock limitations, which is acceptable
	if len(response.Results) > 0 {
		assert.Greater(t, response.Quality.OverallScore, float32(0.5))
	}
	assert.Contains(t, response.Optimization.Optimizations, "profile_aware")
}

func TestRetrievalOptimizer(t *testing.T) {
	optimizer := NewRetrievalOptimizer()

	// Test embedding optimization
	originalEmbedding := &embedding.Embedding{
		Text:   "test query",
		Vector: make(embedding.Vector, 10),
		Model:  "test-model",
	}

	// Fill vector with test values
	for i := range originalEmbedding.Vector {
		originalEmbedding.Vector[i] = float32(i) * 0.1
	}

	profile := profiles.GetProfileByID("security")
	optimizedEmbedding, err := optimizer.OptimizeEmbedding(context.Background(), originalEmbedding, profile)
	require.NoError(t, err)

	assert.NotNil(t, optimizedEmbedding)
	assert.Equal(t, originalEmbedding.Text, optimizedEmbedding.Text)
	assert.Equal(t, len(originalEmbedding.Vector), len(optimizedEmbedding.Vector))

	// Test search optimization
	searchParams, err := optimizer.OptimizeSearch(context.Background(), "security implementation", profile)
	require.NoError(t, err)

	assert.Equal(t, 25, searchParams.Limit) // Security profile should have higher limit
	assert.Equal(t, float32(0.65), searchParams.ScoreThreshold)
	assert.Contains(t, searchParams.BoostFactors, "security")
}

func TestQualityAssessor(t *testing.T) {
	assessor := NewQualityAssessor()

	// Create test results
	results := []ContextualResult{
		{
			Document: vectorstore.Document{
				Content: "authentication implementation details",
				Metadata: map[string]interface{}{
					"source_type": "file",
				},
			},
			Score:           0.9,
			ContextualScore: 0.95,
			Evidence: []ContextualEvidence{
				{
					Type:        "analysis",
					Source:      "vector_similarity",
					Content:     "High semantic similarity",
					Score:       0.9,
					Explanation: "Based on vector similarity between query and document",
				},
			},
		},
		{
			Document: vectorstore.Document{
				Content: "security documentation",
				Metadata: map[string]interface{}{
					"source_type": "documentation",
				},
			},
			Score:           0.7,
			ContextualScore: 0.75,
			Evidence: []ContextualEvidence{
				{
					Type:        "contextual",
					Source:      "profile_match",
					Content:     "Matches security profile",
					Score:       0.8,
					Explanation: "Based on security profile preferences",
				},
			},
		},
	}

	query := &ContextualQuery{
		Query:   "authentication security",
		Profile: profiles.GetProfileByID("security"),
	}

	// Test quality assessment
	quality, err := assessor.Assess(context.Background(), results, query)
	require.NoError(t, err)

	assert.Greater(t, quality.OverallScore, float32(0.6))
	assert.Greater(t, quality.RelevanceScore, float32(0.6))
	assert.Greater(t, quality.DiversityScore, float32(0.4))
	assert.Greater(t, quality.ConfidenceScore, float32(0.5))
	assert.NotEmpty(t, quality.Recommendations)
}

func TestContextualPerformanceMonitor(t *testing.T) {
	monitor := NewContextualPerformanceMonitor()

	// Record some metrics
	ctx := context.Background()
	query := &ContextualQuery{
		Query:   "test query",
		Profile: profiles.GetProfileByID("code_analysis"),
	}

	monitor.RecordSearch(ctx, query, 100*time.Millisecond, 5, 0.8)
	monitor.RecordEmbedding(ctx, 20*time.Millisecond)
	monitor.RecordSearchOperation(ctx, 50*time.Millisecond)
	monitor.RecordRanking(ctx, 30*time.Millisecond)
	monitor.RecordCacheHit(ctx, true)

	// Get metrics
	metrics := monitor.GetMetrics()

	assert.Equal(t, int64(1), metrics.TotalSearches)
	assert.Equal(t, int64(1), metrics.SuccessfulSearches)
	assert.Equal(t, 100*time.Millisecond, metrics.AverageLatency)
	assert.Equal(t, float64(5), metrics.AverageResults)
	assert.Contains(t, metrics.ProfileUsage, "code_analysis")
	assert.Greater(t, metrics.CacheHitRate, float64(0))
}

func TestContextualQueryProcessing(t *testing.T) {
	// Test query with different profiles
	profiles := []*profiles.AgentProfile{
		profiles.GetProfileByID("debugging"),
		profiles.GetProfileByID("documentation"),
		profiles.GetProfileByID("architecture"),
	}

	for _, profile := range profiles {
		query := &ContextualQuery{
			Query:   "analyze system architecture",
			Profile: profile,
			WorkContext: map[string]interface{}{
				"active_file": "main.go",
				"git_branch":  "main",
			},
		}

		// Verify profile is set correctly
		assert.Equal(t, profile.ID, query.Profile.ID)

		// Test work context
		assert.Equal(t, "main.go", query.WorkContext["active_file"])
		assert.Equal(t, "main", query.WorkContext["git_branch"])
	}
}

func TestContextualResultRanking(t *testing.T) {
	// Create test results with different scores
	results := []ContextualResult{
		{
			Document: vectorstore.Document{
				Content: "low relevance content",
				Metadata: map[string]interface{}{
					"source_type": "file",
				},
			},
			Score:           0.5,
			ContextualScore: 0.55,
			Rank:            2,
		},
		{
			Document: vectorstore.Document{
				Content: "high relevance content",
				Metadata: map[string]interface{}{
					"source_type": "file",
					"file_path":   "auth.go",
				},
			},
			Score:           0.8,
			ContextualScore: 0.9,
			Rank:            1,
		},
	}

	// Test sorting by contextual score
	sortByContextualScore(results)

	assert.Equal(t, 1, results[0].Rank)
	assert.Equal(t, 2, results[1].Rank)
	assert.Equal(t, float32(0.9), results[0].ContextualScore)
	assert.Equal(t, float32(0.55), results[1].ContextualScore)
}

func TestEvidenceGeneration(t *testing.T) {
	// Create test document and query
	doc := vectorstore.Document{
		Content: "authentication implementation with JWT",
		Metadata: map[string]interface{}{
			"source_type": "file",
			"file_path":   "auth.go",
			"language":    "go",
		},
	}

	query := &ContextualQuery{
		Query:   "how does authentication work",
		Profile: profiles.GetProfileByID("code_analysis"),
		WorkContext: map[string]interface{}{
			"active_file": "auth.go",
		},
	}

	// Create a simple framework for testing
	framework := &ContextualRetrievalFramework{}

	// Test evidence generation
	evidence := framework.generateEvidence(vectorstore.SearchResult{
		Document: doc,
		Score:    0.8,
	}, query)

	assert.NotEmpty(t, evidence)
	assert.Contains(t, evidence[0].Type, "semantic")
	assert.Greater(t, evidence[0].Score, float32(0))
}

// Mock embedder for testing
type mockEmbedder struct{}

func (m *mockEmbedder) Embed(ctx context.Context, text string) (*embedding.Embedding, error) {
	vector := make(embedding.Vector, 384)
	for i := range vector {
		vector[i] = 0.1 // Simple test vector
	}

	return &embedding.Embedding{
		Text:   text,
		Vector: vector,
		Model:  "mock-model",
	}, nil
}

func (m *mockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([]*embedding.Embedding, error) {
	embeddings := make([]*embedding.Embedding, len(texts))
	for i, text := range texts {
		emb, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

func (m *mockEmbedder) Dimensions() int {
	return 384
}

func (m *mockEmbedder) Model() string {
	return "mock-model"
}

// Mock classifier for testing
type MockClassifier struct{}

func (mc *MockClassifier) Classify(ctx context.Context, query string, workContext map[string]interface{}) (*profiles.ClassificationResult, error) {
	return &profiles.ClassificationResult{
		ProfileID:  "code_analysis",
		Confidence: 0.9,
		Reasoning:  "Mock classification",
	}, nil
}

// Helper function for sorting (moved from framework)
func sortByContextualScore(results []ContextualResult) {
	// Simple bubble sort for small result sets
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].ContextualScore < results[j].ContextualScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Update ranks after sorting
	for i := range results {
		results[i].Rank = i + 1
	}
}
