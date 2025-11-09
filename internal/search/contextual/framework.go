package contextual

import (
	"context"
	"fmt"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
	"github.com/ferg-cod3s/conexus/internal/embedding"
	"github.com/ferg-cod3s/conexus/internal/vectorstore"
)

// ContextualRetrievalFramework provides optimization-aware contextual retrieval
type ContextualRetrievalFramework struct {
	vectorStore        vectorstore.VectorStore
	embedder           embedding.Embedder
	profileManager     *profiles.ProfileManager
	optimizer          *RetrievalOptimizer
	qualityAssessor    *QualityAssessor
	performanceMonitor *ContextualPerformanceMonitor
}

// ContextualRetrievalConfig configures the contextual retrieval framework
type ContextualRetrievalConfig struct {
	VectorStore        vectorstore.VectorStore
	Embedder           embedding.Embedder
	ProfileManager     *profiles.ProfileManager
	Optimizer          *RetrievalOptimizer
	QualityAssessor    *QualityAssessor
	PerformanceMonitor *ContextualPerformanceMonitor
}

// ContextualQuery represents a query with contextual information
type ContextualQuery struct {
	Query        string                 `json:"query"`
	Profile      *profiles.AgentProfile `json:"profile"`
	WorkContext  map[string]interface{} `json:"work_context"`
	Requirements map[string]interface{} `json:"requirements"`
	Constraints  []string               `json:"constraints"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ContextualResult represents a search result with contextual relevance
type ContextualResult struct {
	Document        vectorstore.Document   `json:"document"`
	Score           float32                `json:"score"`
	Rank            int                    `json:"rank"`
	ContextualScore float32                `json:"contextual_score"`
	Relevance       RelevanceScore         `json:"relevance"`
	Evidence        []ContextualEvidence   `json:"evidence"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// RelevanceScore represents different types of relevance
type RelevanceScore struct {
	Semantic   float32 `json:"semantic"`   // Semantic similarity
	Contextual float32 `json:"contextual"` // Context-based relevance
	Temporal   float32 `json:"temporal"`   // Time-based relevance
	Structural float32 `json:"structural"` // Code structure relevance
	Behavioral float32 `json:"behavioral"` // User behavior relevance
}

// ContextualEvidence provides evidence for why a result is relevant
type ContextualEvidence struct {
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Content     string                 `json:"content"`
	Score       float32                `json:"score"`
	Explanation string                 `json:"explanation"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ContextualSearchResponse represents the response from contextual search
type ContextualSearchResponse struct {
	Results        []ContextualResult     `json:"results"`
	TotalCount     int                    `json:"total_count"`
	QueryTime      float64                `json:"query_time_ms"`
	ContextualTime float64                `json:"contextual_time_ms"`
	Optimization   *OptimizationInfo      `json:"optimization"`
	Quality        *QualityMetrics        `json:"quality"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// OptimizationInfo provides information about optimizations applied
type OptimizationInfo struct {
	EmbeddingOptimized bool                   `json:"embedding_optimized"`
	SearchOptimized    bool                   `json:"search_optimized"`
	RankingOptimized   bool                   `json:"ranking_optimized"`
	CacheHit           bool                   `json:"cache_hit"`
	CacheKey           string                 `json:"cache_key"`
	Optimizations      []string               `json:"optimizations"`
	PerformanceGain    float64                `json:"performance_gain_ms"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// QualityMetrics provides quality assessment of results
type QualityMetrics struct {
	OverallScore    float32                `json:"overall_score"`
	RelevanceScore  float32                `json:"relevance_score"`
	DiversityScore  float32                `json:"diversity_score"`
	CoverageScore   float32                `json:"coverage_score"`
	ConfidenceScore float32                `json:"confidence_score"`
	Issues          []string               `json:"issues"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewContextualRetrievalFramework creates a new contextual retrieval framework
func NewContextualRetrievalFramework(config ContextualRetrievalConfig) *ContextualRetrievalFramework {
	return &ContextualRetrievalFramework{
		vectorStore:        config.VectorStore,
		embedder:           config.Embedder,
		profileManager:     config.ProfileManager,
		optimizer:          config.Optimizer,
		qualityAssessor:    config.QualityAssessor,
		performanceMonitor: config.PerformanceMonitor,
	}
}

// Search performs contextual search with optimization and quality assessment
func (crf *ContextualRetrievalFramework) Search(ctx context.Context, query *ContextualQuery) (*ContextualSearchResponse, error) {
	startTime := time.Now()

	// Select appropriate profile if not provided
	if query.Profile == nil {
		profile, _, err := crf.profileManager.SelectProfile(ctx, query.Query, query.WorkContext)
		if err != nil {
			return nil, fmt.Errorf("failed to select profile: %w", err)
		}
		query.Profile = profile
	}

	// Generate optimized embedding
	embedding, embeddingTime, err := crf.generateOptimizedEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Perform contextual search
	searchResults, searchTime, err := crf.performContextualSearch(ctx, query, embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search: %w", err)
	}

	// Apply contextual ranking
	rankedResults, rankingTime, err := crf.applyContextualRanking(ctx, searchResults, query)
	if err != nil {
		return nil, fmt.Errorf("failed to apply ranking: %w", err)
	}

	// Assess quality
	quality, qualityTime, err := crf.assessQuality(ctx, rankedResults, query)
	if err != nil {
		return nil, fmt.Errorf("failed to assess quality: %w", err)
	}

	// Create optimization info
	optimization := &OptimizationInfo{
		EmbeddingOptimized: true,
		SearchOptimized:    true,
		RankingOptimized:   true,
		Optimizations:      []string{"profile_aware", "context_boosted", "quality_filtered"},
		Metadata: map[string]interface{}{
			"profile_id": query.Profile.ID,
		},
	}

	totalTime := time.Since(startTime)

	// Record performance metrics
	if crf.performanceMonitor != nil {
		crf.performanceMonitor.RecordSearch(ctx, query, totalTime, len(rankedResults), quality.OverallScore)
	}

	return &ContextualSearchResponse{
		Results:        rankedResults,
		TotalCount:     len(rankedResults),
		QueryTime:      searchTime.Seconds() * 1000,
		ContextualTime: (embeddingTime + rankingTime + qualityTime).Seconds() * 1000,
		Optimization:   optimization,
		Quality:        quality,
		Metadata: map[string]interface{}{
			"profile_id":        query.Profile.ID,
			"total_time_ms":     totalTime.Seconds() * 1000,
			"embedding_time_ms": embeddingTime.Seconds() * 1000,
			"search_time_ms":    searchTime.Seconds() * 1000,
			"ranking_time_ms":   rankingTime.Seconds() * 1000,
			"quality_time_ms":   qualityTime.Seconds() * 1000,
		},
	}, nil
}

// generateOptimizedEmbedding generates an embedding optimized for the query profile
func (crf *ContextualRetrievalFramework) generateOptimizedEmbedding(ctx context.Context, query *ContextualQuery) (*embedding.Embedding, time.Duration, error) {
	startTime := time.Now()

	// Use the embedder to generate embedding
	emb, err := crf.embedder.Embed(ctx, query.Query)
	if err != nil {
		return nil, 0, err
	}

	// Apply profile-specific optimizations
	if crf.optimizer != nil {
		optimizedEmb, err := crf.optimizer.OptimizeEmbedding(ctx, emb, query.Profile)
		if err == nil {
			emb = optimizedEmb
		}
	}

	embeddingTime := time.Since(startTime)
	return emb, embeddingTime, nil
}

// performContextualSearch performs search with contextual parameters
func (crf *ContextualRetrievalFramework) performContextualSearch(ctx context.Context, query *ContextualQuery, embedding *embedding.Embedding) ([]vectorstore.SearchResult, time.Duration, error) {
	startTime := time.Now()

	// Convert embedding to vector
	vector := embedding.Vector

	// Create search filters based on profile and context
	filters := crf.buildSearchFilters(query)

	// Perform hybrid search (vector + BM25)
	results, err := crf.vectorStore.SearchHybrid(ctx, query.Query, vector, vectorstore.SearchOptions{
		Limit:     crf.getOptimalLimit(query.Profile),
		Filters:   filters,
		Threshold: crf.getScoreThreshold(query.Profile),
	})
	if err != nil {
		return nil, 0, err
	}

	searchTime := time.Since(startTime)
	return results, searchTime, nil
}

// applyContextualRanking applies contextual ranking to search results
func (crf *ContextualRetrievalFramework) applyContextualRanking(ctx context.Context, results []vectorstore.SearchResult, query *ContextualQuery) ([]ContextualResult, time.Duration, error) {
	startTime := time.Now()

	var contextualResults []ContextualResult

	for i, result := range results {
		// Calculate contextual relevance
		contextualScore := crf.calculateContextualScore(result, query)

		// Calculate relevance breakdown
		relevance := crf.calculateRelevanceScore(result, query, contextualScore)

		// Generate evidence
		evidence := crf.generateEvidence(result, query)

		contextualResult := ContextualResult{
			Document:        result.Document,
			Score:           result.Score,
			Rank:            i + 1,
			ContextualScore: contextualScore,
			Relevance:       relevance,
			Evidence:        evidence,
			Metadata: map[string]interface{}{
				"original_score": result.Score,
				"boost_applied":  contextualScore != result.Score,
			},
		}

		contextualResults = append(contextualResults, contextualResult)
	}

	// Sort by contextual score
	crf.sortByContextualScore(contextualResults)

	rankingTime := time.Since(startTime)
	return contextualResults, rankingTime, nil
}

// assessQuality assesses the quality of search results
func (crf *ContextualRetrievalFramework) assessQuality(ctx context.Context, results []ContextualResult, query *ContextualQuery) (*QualityMetrics, time.Duration, error) {
	startTime := time.Now()

	if crf.qualityAssessor == nil {
		// Return default metrics if no assessor
		quality := &QualityMetrics{
			OverallScore:    0.8,
			RelevanceScore:  0.8,
			DiversityScore:  0.7,
			CoverageScore:   0.8,
			ConfidenceScore: 0.8,
			Issues:          []string{},
			Recommendations: []string{},
		}
		return quality, time.Since(startTime), nil
	}

	quality, err := crf.qualityAssessor.Assess(ctx, results, query)
	if err != nil {
		return nil, 0, err
	}

	qualityTime := time.Since(startTime)
	return quality, qualityTime, nil
}

// buildSearchFilters builds search filters based on profile and context
func (crf *ContextualRetrievalFramework) buildSearchFilters(query *ContextualQuery) map[string]interface{} {
	filters := make(map[string]interface{})

	// Add profile-based filters
	if query.Profile != nil {
		// Filter by content types relevant to the profile
		contentTypes := crf.getRelevantContentTypes(query.Profile)
		if len(contentTypes) > 0 {
			filters["source_types"] = contentTypes
		}

		// Add time-based filters if specified
		if query.WorkContext != nil {
			if dateRange, exists := query.WorkContext["date_range"]; exists {
				filters["date_range"] = dateRange
			}
		}
	}

	// Add work context filters
	if query.WorkContext != nil {
		if activeFile, exists := query.WorkContext["active_file"]; exists {
			filters["active_file"] = activeFile
		}

		if gitBranch, exists := query.WorkContext["git_branch"]; exists {
			filters["git_branch"] = gitBranch
		}

		if openTickets, exists := query.WorkContext["open_ticket_ids"]; exists {
			filters["open_ticket_ids"] = openTickets
		}
	}

	return filters
}

// getOptimalLimit returns the optimal result limit based on profile
func (crf *ContextualRetrievalFramework) getOptimalLimit(profile *profiles.AgentProfile) int {
	if profile == nil {
		return 20
	}

	// Base limit on context window size
	baseLimit := 20
	if profile.ContextWindow.OptimalTokens > 16000 {
		baseLimit = 30
	} else if profile.ContextWindow.OptimalTokens > 8000 {
		baseLimit = 25
	} else if profile.ContextWindow.OptimalTokens < 4000 {
		baseLimit = 15
	}

	// Adjust based on optimization hints
	if profile.OptimizationHints.ParallelQueries > 1 {
		baseLimit = int(float64(baseLimit) * 1.2)
	}

	return min(baseLimit, 100) // Cap at 100
}

// getScoreThreshold returns the score threshold based on profile
func (crf *ContextualRetrievalFramework) getScoreThreshold(profile *profiles.AgentProfile) float32 {
	if profile == nil {
		return 0.5
	}

	// Higher precision profiles need higher thresholds
	switch profile.ID {
	case "debugging":
		return 0.7 // High precision for debugging
	case "security":
		return 0.65 // High precision for security
	case "code_analysis":
		return 0.6 // Medium-high precision
	case "architecture":
		return 0.5 // Lower threshold for broader context
	case "documentation":
		return 0.45 // Lower threshold for comprehensive docs
	default:
		return 0.5
	}
}

// getRelevantContentTypes returns content types relevant to the profile
func (crf *ContextualRetrievalFramework) getRelevantContentTypes(profile *profiles.AgentProfile) []string {
	if profile == nil {
		return []string{"file", "github", "slack"}
	}

	// Map profile to relevant content types based on weights
	var contentTypes []string

	// Always include high-weight content types
	if profile.Weights.Code > 0.5 {
		contentTypes = append(contentTypes, "file")
	}
	if profile.Weights.Documentation > 0.5 {
		contentTypes = append(contentTypes, "documentation")
	}
	if profile.Weights.Discussions > 0.5 {
		contentTypes = append(contentTypes, "slack", "github_issue", "github_pr")
	}
	if profile.Weights.Config > 0.5 {
		contentTypes = append(contentTypes, "config")
	}
	if profile.Weights.Tests > 0.5 {
		contentTypes = append(contentTypes, "test")
	}

	// Default fallback
	if len(contentTypes) == 0 {
		contentTypes = []string{"file", "github", "slack"}
	}

	return contentTypes
}

// calculateContextualScore calculates contextual relevance score
func (crf *ContextualRetrievalFramework) calculateContextualScore(result vectorstore.SearchResult, query *ContextualQuery) float32 {
	baseScore := result.Score

	// Apply work context boosting
	boost := crf.calculateWorkContextBoost(result, query)
	baseScore += boost

	// Apply profile-based adjustments
	if query.Profile != nil {
		profileBoost := crf.calculateProfileBoost(result, query.Profile)
		baseScore += profileBoost
	}

	// Apply temporal relevance
	temporalBoost := crf.calculateTemporalBoost(result, query)
	baseScore += temporalBoost

	// Cap at 1.0
	if baseScore > 1.0 {
		baseScore = 1.0
	}

	return baseScore
}

// calculateWorkContextBoost calculates boost based on work context
func (crf *ContextualRetrievalFramework) calculateWorkContextBoost(result vectorstore.SearchResult, query *ContextualQuery) float32 {
	boost := float32(0.0)

	if query.WorkContext == nil {
		return boost
	}

	// Boost for active file matches
	if activeFile, exists := query.WorkContext["active_file"]; exists {
		if filePath, ok := result.Document.Metadata["file_path"].(string); ok {
			if filePath == activeFile {
				boost += 0.2
			}
		}
	}

	// Boost for git branch matches
	if gitBranch, exists := query.WorkContext["git_branch"]; exists {
		if branch, ok := result.Document.Metadata["git_branch"].(string); ok {
			if branch == gitBranch {
				boost += 0.1
			}
		}
	}

	// Boost for related tickets
	if openTickets, exists := query.WorkContext["open_ticket_ids"]; exists {
		if ticketIDs, ok := openTickets.([]string); ok {
			for _, ticketID := range ticketIDs {
				if result.Document.Metadata["ticket_id"] == ticketID {
					boost += 0.15
					break
				}
			}
		}
	}

	return boost
}

// calculateProfileBoost calculates boost based on profile preferences
func (crf *ContextualRetrievalFramework) calculateProfileBoost(result vectorstore.SearchResult, profile *profiles.AgentProfile) float32 {
	boost := float32(0.0)

	// Get source type from metadata
	sourceType, hasSourceType := result.Document.Metadata["source_type"].(string)
	if !hasSourceType {
		return boost
	}

	// Apply weights based on profile preferences
	switch sourceType {
	case "file":
		boost += float32(profile.Weights.Code) * 0.1
	case "documentation":
		boost += float32(profile.Weights.Documentation) * 0.1
	case "github_issue", "github_pr", "slack":
		boost += float32(profile.Weights.Discussions) * 0.1
	case "config":
		boost += float32(profile.Weights.Config) * 0.1
	case "test":
		boost += float32(profile.Weights.Tests) * 0.1
	}

	return boost
}

// calculateTemporalBoost calculates temporal relevance boost
func (crf *ContextualRetrievalFramework) calculateTemporalBoost(result vectorstore.SearchResult, query *ContextualQuery) float32 {
	// Simple implementation - boost recent content
	if result.Document.UpdatedAt.IsZero() {
		return 0.0
	}

	hoursSinceUpdate := time.Since(result.Document.UpdatedAt).Hours()

	// Boost recent content (within last 24 hours)
	if hoursSinceUpdate < 24 {
		return 0.05
	}

	// Small boost for content within last week
	if hoursSinceUpdate < 168 {
		return 0.02
	}

	return 0.0
}

// calculateRelevanceScore calculates detailed relevance breakdown
func (crf *ContextualRetrievalFramework) calculateRelevanceScore(result vectorstore.SearchResult, query *ContextualQuery, contextualScore float32) RelevanceScore {
	return RelevanceScore{
		Semantic:   result.Score,
		Contextual: contextualScore - result.Score,
		Temporal:   crf.calculateTemporalBoost(result, query),
		Structural: crf.calculateStructuralRelevance(result, query),
		Behavioral: crf.calculateBehavioralRelevance(result, query),
	}
}

// calculateStructuralRelevance calculates structural relevance (code structure)
func (crf *ContextualRetrievalFramework) calculateStructuralRelevance(result vectorstore.SearchResult, query *ContextualQuery) float32 {
	// Placeholder implementation
	// In a real system, this would analyze code structure relevance
	return 0.1
}

// calculateBehavioralRelevance calculates behavioral relevance (user patterns)
func (crf *ContextualRetrievalFramework) calculateBehavioralRelevance(result vectorstore.SearchResult, query *ContextualQuery) float32 {
	// Placeholder implementation
	// In a real system, this would analyze user behavior patterns
	return 0.1
}

// generateEvidence generates evidence for result relevance
func (crf *ContextualRetrievalFramework) generateEvidence(result vectorstore.SearchResult, query *ContextualQuery) []ContextualEvidence {
	var evidence []ContextualEvidence

	// Add semantic evidence
	evidence = append(evidence, ContextualEvidence{
		Type:        "semantic",
		Source:      "vector_similarity",
		Content:     fmt.Sprintf("Semantic similarity score: %.3f", result.Score),
		Score:       result.Score,
		Explanation: "Based on vector similarity between query and document",
	})

	// Add contextual evidence
	contextualBoost := crf.calculateWorkContextBoost(result, query)
	if contextualBoost > 0 {
		evidence = append(evidence, ContextualEvidence{
			Type:        "contextual",
			Source:      "work_context",
			Content:     fmt.Sprintf("Contextual boost: %.3f", contextualBoost),
			Score:       contextualBoost,
			Explanation: "Based on work context relevance (active file, branch, tickets)",
		})
	}

	// Add profile evidence
	if query.Profile != nil {
		profileBoost := crf.calculateProfileBoost(result, query.Profile)
		if profileBoost > 0 {
			evidence = append(evidence, ContextualEvidence{
				Type:        "profile",
				Source:      query.Profile.ID,
				Content:     fmt.Sprintf("Profile boost: %.3f", profileBoost),
				Score:       profileBoost,
				Explanation: fmt.Sprintf("Based on %s profile preferences", query.Profile.Name),
			})
		}
	}

	return evidence
}

// sortByContextualScore sorts results by contextual score
func (crf *ContextualRetrievalFramework) sortByContextualScore(results []ContextualResult) {
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

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
