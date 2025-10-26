package contextual

import (
	"context"
	"math"
	"strings"
)

// QualityAssessor assesses the quality of search results
type QualityAssessor struct {
	thresholds QualityThresholds
}

// QualityThresholds defines quality assessment thresholds
type QualityThresholds struct {
	MinRelevanceScore  float32
	MinDiversityScore  float32
	MinCoverageScore   float32
	MinConfidenceScore float32
}

// NewQualityAssessor creates a new quality assessor
func NewQualityAssessor() *QualityAssessor {
	return &QualityAssessor{
		thresholds: QualityThresholds{
			MinRelevanceScore:  0.6,
			MinDiversityScore:  0.5,
			MinCoverageScore:   0.7,
			MinConfidenceScore: 0.5,
		},
	}
}

// Assess assesses the quality of contextual search results
func (qa *QualityAssessor) Assess(ctx context.Context, results []ContextualResult, query *ContextualQuery) (*QualityMetrics, error) {
	if len(results) == 0 {
		return &QualityMetrics{
			OverallScore:    0.0,
			RelevanceScore:  0.0,
			DiversityScore:  0.0,
			CoverageScore:   0.0,
			ConfidenceScore: 0.0,
			Issues:          []string{"no_results"},
			Recommendations: []string{"check_query_terms", "expand_search"},
		}, nil
	}

	// Calculate relevance score
	relevanceScore := qa.calculateRelevanceScore(results)

	// Calculate diversity score
	diversityScore := qa.calculateDiversityScore(results)

	// Calculate coverage score
	coverageScore := qa.calculateCoverageScore(results, query)

	// Calculate confidence score
	confidenceScore := qa.calculateConfidenceScore(results)

	// Calculate overall score
	overallScore := (relevanceScore + diversityScore + coverageScore + confidenceScore) / 4.0

	// Generate issues and recommendations
	issues := qa.identifyIssues(results, relevanceScore, diversityScore, coverageScore, confidenceScore)
	recommendations := qa.generateRecommendations(results, query, issues)

	return &QualityMetrics{
		OverallScore:    overallScore,
		RelevanceScore:  relevanceScore,
		DiversityScore:  diversityScore,
		CoverageScore:   coverageScore,
		ConfidenceScore: confidenceScore,
		Issues:          issues,
		Recommendations: recommendations,
		Metadata: map[string]interface{}{
			"result_count": len(results),
		},
	}, nil
}

// calculateRelevanceScore calculates how relevant the results are
func (qa *QualityAssessor) calculateRelevanceScore(results []ContextualResult) float32 {
	if len(results) == 0 {
		return 0.0
	}

	totalScore := float32(0.0)
	for _, result := range results {
		// Use contextual score as primary relevance indicator
		score := result.ContextualScore

		// Boost for high evidence count
		evidenceCount := len(result.Evidence)
		if evidenceCount > 3 {
			score += 0.1
		} else if evidenceCount > 1 {
			score += 0.05
		}

		// Boost for high confidence evidence
		avgEvidenceConfidence := float32(0.0)
		if evidenceCount > 0 {
			totalEvidenceScore := float32(0.0)
			for _, evidence := range result.Evidence {
				totalEvidenceScore += evidence.Score
			}
			avgEvidenceConfidence = totalEvidenceScore / float32(evidenceCount)
			score += avgEvidenceConfidence * 0.1
		}

		totalScore += score
	}

	averageScore := totalScore / float32(len(results))

	// Normalize to 0-1 range
	if averageScore > 1.0 {
		averageScore = 1.0
	}

	return averageScore
}

// calculateDiversityScore calculates how diverse the results are
func (qa *QualityAssessor) calculateDiversityScore(results []ContextualResult) float32 {
	if len(results) <= 1 {
		return 1.0 // Perfect diversity with single result
	}

	// Calculate diversity based on source types and content types
	sourceTypes := make(map[string]int)
	contentTypes := make(map[string]int)

	for _, result := range results {
		if sourceType, ok := result.Document.Metadata["source_type"].(string); ok {
			sourceTypes[sourceType]++
		}

		// Check for content type in metadata
		if contentType, ok := result.Document.Metadata["content_type"].(string); ok {
			contentTypes[contentType]++
		}
	}

	// Calculate source type diversity
	sourceDiversity := float32(len(sourceTypes)) / float32(len(results))
	if sourceDiversity > 1.0 {
		sourceDiversity = 1.0
	}

	// Calculate content type diversity
	contentDiversity := float32(len(contentTypes)) / float32(len(results))
	if contentDiversity > 1.0 {
		contentDiversity = 1.0
	}

	// Combine diversity scores
	diversityScore := (sourceDiversity + contentDiversity) / 2.0

	return diversityScore
}

// calculateCoverageScore calculates how well the results cover the query
func (qa *QualityAssessor) calculateCoverageScore(results []ContextualResult, query *ContextualQuery) float32 {
	if query == nil || query.Query == "" {
		return 0.5
	}

	// Simple coverage based on query term coverage
	queryTerms := strings.Fields(strings.ToLower(query.Query))
	if len(queryTerms) == 0 {
		return 0.5
	}

	totalCoverage := 0
	for _, term := range queryTerms {
		termFound := false
		for _, result := range results {
			content := strings.ToLower(result.Document.Content)
			if strings.Contains(content, term) {
				termFound = true
				break
			}
		}
		if termFound {
			totalCoverage++
		}
	}

	coverageScore := float32(totalCoverage) / float32(len(queryTerms))

	// Boost for results with comprehensive content
	if len(results) > 0 {
		avgContentLength := 0
		for _, result := range results {
			avgContentLength += len(result.Document.Content)
		}
		avgContentLength /= len(results)

		// Boost for longer, more comprehensive results
		if avgContentLength > 500 {
			coverageScore += 0.1
		} else if avgContentLength > 200 {
			coverageScore += 0.05
		}
	}

	// Cap at 1.0
	if coverageScore > 1.0 {
		coverageScore = 1.0
	}

	return coverageScore
}

// calculateConfidenceScore calculates overall confidence in results
func (qa *QualityAssessor) calculateConfidenceScore(results []ContextualResult) float32 {
	if len(results) == 0 {
		return 0.0
	}

	// Average confidence across all results
	totalConfidence := float32(0.0)
	for _, result := range results {
		// Use the higher of contextual score and evidence confidence
		confidence := result.ContextualScore

		// Average evidence confidence
		if len(result.Evidence) > 0 {
			evidenceConfidence := float32(0.0)
			for _, evidence := range result.Evidence {
				evidenceConfidence += evidence.Score
			}
			evidenceConfidence /= float32(len(result.Evidence))

			if evidenceConfidence > confidence {
				confidence = evidenceConfidence
			}
		}

		totalConfidence += confidence
	}

	averageConfidence := totalConfidence / float32(len(results))

	// Apply penalty for inconsistent scores
	scoreVariance := qa.calculateScoreVariance(results)
	variancePenalty := math.Min(float64(scoreVariance*2.0), 0.2)
	averageConfidence -= float32(variancePenalty)

	// Ensure non-negative
	if averageConfidence < 0.0 {
		averageConfidence = 0.0
	}

	return averageConfidence
}

// calculateScoreVariance calculates variance in result scores
func (qa *QualityAssessor) calculateScoreVariance(results []ContextualResult) float32 {
	if len(results) <= 1 {
		return 0.0
	}

	// Calculate mean
	mean := float32(0.0)
	for _, result := range results {
		mean += result.ContextualScore
	}
	mean /= float32(len(results))

	// Calculate variance
	variance := float32(0.0)
	for _, result := range results {
		diff := result.ContextualScore - mean
		variance += diff * diff
	}
	variance /= float32(len(results))

	return variance
}

// identifyIssues identifies quality issues in results
func (qa *QualityAssessor) identifyIssues(results []ContextualResult, relevanceScore, diversityScore, coverageScore, confidenceScore float32) []string {
	var issues []string

	if relevanceScore < qa.thresholds.MinRelevanceScore {
		issues = append(issues, "low_relevance")
	}

	if diversityScore < qa.thresholds.MinDiversityScore {
		issues = append(issues, "low_diversity")
	}

	if coverageScore < qa.thresholds.MinCoverageScore {
		issues = append(issues, "low_coverage")
	}

	if confidenceScore < qa.thresholds.MinConfidenceScore {
		issues = append(issues, "low_confidence")
	}

	// Check for result quality issues
	if len(results) < 3 {
		issues = append(issues, "insufficient_results")
	}

	// Check for score consistency
	if qa.calculateScoreVariance(results) > 0.3 {
		issues = append(issues, "inconsistent_scores")
	}

	return issues
}

// generateRecommendations generates recommendations for improving results
func (qa *QualityAssessor) generateRecommendations(results []ContextualResult, query *ContextualQuery, issues []string) []string {
	var recommendations []string

	for _, issue := range issues {
		switch issue {
		case "low_relevance":
			recommendations = append(recommendations, "try different search terms")
			recommendations = append(recommendations, "check for typos in query")
		case "low_diversity":
			recommendations = append(recommendations, "expand search to include related topics")
			recommendations = append(recommendations, "consider broader search terms")
		case "low_coverage":
			recommendations = append(recommendations, "increase result limit")
			recommendations = append(recommendations, "remove restrictive filters")
		case "low_confidence":
			recommendations = append(recommendations, "verify query specificity")
			recommendations = append(recommendations, "consider more specific search terms")
		case "insufficient_results":
			recommendations = append(recommendations, "increase search scope")
			recommendations = append(recommendations, "lower score threshold")
		case "inconsistent_scores":
			recommendations = append(recommendations, "review query for ambiguity")
			recommendations = append(recommendations, "consider multiple interpretations")
		}
	}

	// Add general recommendations
	if len(results) > 0 {
		recommendations = append(recommendations, "review top results for relevance")
	}

	return recommendations
}
