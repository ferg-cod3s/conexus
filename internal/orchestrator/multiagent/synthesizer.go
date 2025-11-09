package multiagent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
)

// DefaultResultSynthesizer implements result synthesis logic
type DefaultResultSynthesizer struct {
	conflictDetector ConflictDetector
	evidenceWeighter EvidenceWeighter
}

// ConflictDetector detects conflicts between agent results
type ConflictDetector interface {
	DetectConflicts(ctx context.Context, results []*AgentResult, task *MultiAgentTask) ([]Conflict, error)
}

// EvidenceWeighter weights evidence from different agents
type EvidenceWeighter interface {
	WeightEvidence(ctx context.Context, evidence []Evidence, agent *RegisteredAgent) (float64, error)
}

// NewDefaultResultSynthesizer creates a new default result synthesizer
func NewDefaultResultSynthesizer(detector ConflictDetector, weighter EvidenceWeighter) *DefaultResultSynthesizer {
	return &DefaultResultSynthesizer{
		conflictDetector: detector,
		evidenceWeighter: weighter,
	}
}

// Synthesize combines results from multiple agents
func (drs *DefaultResultSynthesizer) Synthesize(ctx context.Context, results []*AgentResult, task *MultiAgentTask) (*SynthesizedResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to synthesize")
	}

	// Check if all results are successful
	allSuccessful := true
	var successfulResults []*AgentResult
	for _, result := range results {
		if result.Success {
			successfulResults = append(successfulResults, result)
		} else {
			allSuccessful = false
		}
	}

	// If no successful results, return failure
	if len(successfulResults) == 0 {
		return &SynthesizedResult{
			TaskID:       task.ID,
			Success:      false,
			Summary:      "All agents failed to complete the task",
			Details:      make(map[string]interface{}),
			Confidence:   0.0,
			AgentResults: results,
			Conflicts:    []Conflict{},
			Resolutions:  []Resolution{},
			Metadata: map[string]interface{}{
				"failure_reason": "all_agents_failed",
			},
			Duration: time.Duration(0),
		}, nil
	}

	// Detect conflicts
	conflicts, err := drs.conflictDetector.DetectConflicts(ctx, successfulResults, task)
	if err != nil {
		return nil, fmt.Errorf("failed to detect conflicts: %w", err)
	}

	// Generate summary based on results
	summary := drs.generateSummary(successfulResults, task)

	// Calculate overall confidence
	confidence := drs.calculateOverallConfidence(successfulResults, task.Profile)

	// Create synthesized result
	synthesized := &SynthesizedResult{
		TaskID:       task.ID,
		Success:      allSuccessful,
		Summary:      summary,
		Details:      drs.extractDetails(successfulResults),
		Confidence:   confidence,
		AgentResults: results,
		Conflicts:    conflicts,
		Resolutions:  []Resolution{}, // Will be filled by conflict resolver
		Metadata: map[string]interface{}{
			"agent_count":       len(results),
			"successful_agents": len(successfulResults),
			"failed_agents":     len(results) - len(successfulResults),
			"conflict_count":    len(conflicts),
			"synthesis_method":  "weighted_consensus",
		},
		Duration: time.Duration(0), // Will be set by orchestrator
	}

	return synthesized, nil
}

// generateSummary creates a summary from multiple agent results
func (drs *DefaultResultSynthesizer) generateSummary(results []*AgentResult, task *MultiAgentTask) string {
	if len(results) == 0 {
		return "No results available"
	}

	if len(results) == 1 {
		// Single result - use it directly
		result := results[0]
		if result.Output != nil {
			if str, ok := result.Output.(string); ok {
				return str
			}
			return fmt.Sprintf("%v", result.Output)
		}
		return "Task completed successfully"
	}

	// Multiple results - create consensus summary
	var summaries []string
	totalConfidence := 0.0

	for _, result := range results {
		if result.Output != nil {
			if str, ok := result.Output.(string); ok {
				summaries = append(summaries, str)
			} else {
				summaries = append(summaries, fmt.Sprintf("%v", result.Output))
			}
		}
		totalConfidence += result.Confidence
	}

	// Create weighted summary
	averageConfidence := totalConfidence / float64(len(results))

	if averageConfidence > 0.8 {
		return fmt.Sprintf("High confidence consensus: %s", strings.Join(summaries, "; "))
	} else if averageConfidence > 0.6 {
		return fmt.Sprintf("Moderate confidence consensus: %s", strings.Join(summaries, "; "))
	} else {
		return fmt.Sprintf("Low confidence synthesis: %s", strings.Join(summaries, "; "))
	}
}

// calculateOverallConfidence calculates the overall confidence of the synthesis
func (drs *DefaultResultSynthesizer) calculateOverallConfidence(results []*AgentResult, profile *profiles.AgentProfile) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Weight confidence by agent performance and profile relevance
	totalWeightedConfidence := 0.0
	totalWeight := 0.0

	for _, result := range results {
		// Base weight from confidence
		weight := result.Confidence

		// Adjust weight based on profile relevance
		if profile != nil {
			profileWeight := drs.getProfileRelevanceWeight(result.AgentID, profile)
			weight *= profileWeight
		}

		// Adjust weight based on evidence strength
		evidenceWeight := drs.calculateEvidenceWeight(result.Evidence)
		weight *= evidenceWeight

		totalWeightedConfidence += result.Confidence * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	confidence := totalWeightedConfidence / totalWeight

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getProfileRelevanceWeight calculates how relevant an agent is for the given profile
func (drs *DefaultResultSynthesizer) getProfileRelevanceWeight(agentID string, profile *profiles.AgentProfile) float64 {
	// This would ideally check the agent's registered profile against the task profile
	// For now, use a simple heuristic based on profile ID matching

	// If agent profile matches task profile exactly, high relevance
	if strings.Contains(agentID, profile.ID) {
		return 1.2
	}

	// If agent profile is related, medium relevance
	relatedProfiles := map[string][]string{
		"code_analysis": {"debugging", "security"},
		"documentation": {"architecture"},
		"debugging":     {"code_analysis", "security"},
		"architecture":  {"documentation", "code_analysis"},
		"security":      {"code_analysis", "debugging"},
	}

	if related, exists := relatedProfiles[profile.ID]; exists {
		for _, relatedProfile := range related {
			if strings.Contains(agentID, relatedProfile) {
				return 1.0
			}
		}
	}

	// Default relevance
	return 0.8
}

// calculateEvidenceWeight calculates the weight of evidence
func (drs *DefaultResultSynthesizer) calculateEvidenceWeight(evidence []Evidence) float64 {
	if len(evidence) == 0 {
		return 0.5 // Low weight for no evidence
	}

	totalWeight := 0.0
	for _, ev := range evidence {
		// Weight based on evidence type and confidence
		typeWeight := drs.getEvidenceTypeWeight(ev.Type)
		totalWeight += typeWeight * ev.Confidence
	}

	averageWeight := totalWeight / float64(len(evidence))

	// Normalize to 0.5-1.5 range
	if averageWeight < 0.5 {
		averageWeight = 0.5
	} else if averageWeight > 1.5 {
		averageWeight = 1.5
	}

	return averageWeight
}

// getEvidenceTypeWeight returns weight for different evidence types
func (drs *DefaultResultSynthesizer) getEvidenceTypeWeight(evidenceType string) float64 {
	weights := map[string]float64{
		"analysis":      1.0,
		"code_example":  1.2,
		"test_result":   1.1,
		"documentation": 0.9,
		"discussion":    0.8,
		"configuration": 0.7,
		"performance":   1.1,
		"security_scan": 1.3,
		"error_log":     1.0,
		"stack_trace":   1.1,
		"user_feedback": 0.6,
	}

	if weight, exists := weights[evidenceType]; exists {
		return weight
	}

	return 0.8 // Default weight
}

// extractDetails extracts detailed information from results
func (drs *DefaultResultSynthesizer) extractDetails(results []*AgentResult) map[string]interface{} {
	details := make(map[string]interface{})

	// Group results by agent type/capability
	byCapability := make(map[string][]*AgentResult)
	byAgent := make(map[string]*AgentResult)

	for _, result := range results {
		byAgent[result.AgentID] = result

		// Extract capability from metadata or use agent ID
		capability := result.AgentID
		if cap, exists := result.Metadata["capability"]; exists {
			if capStr, ok := cap.(string); ok {
				capability = capStr
			}
		}

		byCapability[capability] = append(byCapability[capability], result)
	}

	details["results_by_agent"] = byAgent
	details["results_by_capability"] = byCapability
	details["total_results"] = len(results)

	// Extract key findings
	var keyFindings []string
	for _, result := range results {
		if result.Output != nil {
			if str, ok := result.Output.(string); ok {
				// Extract first sentence or key phrase
				sentences := strings.Split(str, ".")
				if len(sentences) > 0 && sentences[0] != "" {
					keyFindings = append(keyFindings, strings.TrimSpace(sentences[0]))
				}
			}
		}
	}

	details["key_findings"] = keyFindings

	return details
}

// DefaultConflictDetector implements conflict detection
type DefaultConflictDetector struct{}

// NewDefaultConflictDetector creates a new default conflict detector
func NewDefaultConflictDetector() *DefaultConflictDetector {
	return &DefaultConflictDetector{}
}

// DetectConflicts detects conflicts between agent results
func (dcd *DefaultConflictDetector) DetectConflicts(ctx context.Context, results []*AgentResult, task *MultiAgentTask) ([]Conflict, error) {
	var conflicts []Conflict

	if len(results) < 2 {
		return conflicts, nil // No conflicts possible with single result
	}

	// Compare each pair of results
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			result1 := results[i]
			result2 := results[j]

			conflict := dcd.detectPairConflict(result1, result2, task)
			if conflict != nil {
				conflicts = append(conflicts, *conflict)
			}
		}
	}

	return conflicts, nil
}

// detectPairConflict detects conflicts between two results
func (dcd *DefaultConflictDetector) detectPairConflict(result1, result2 *AgentResult, task *MultiAgentTask) *Conflict {
	// Extract text content for comparison
	text1 := dcd.extractText(result1.Output)
	text2 := dcd.extractText(result2.Output)

	if text1 == "" || text2 == "" {
		return nil
	}

	// Check for contradictions
	contradictionScore := dcd.calculateContradictionScore(text1, text2)
	if contradictionScore > 0.7 {
		return &Conflict{
			ID:          generateConflictID(result1.AgentID, result2.AgentID),
			Type:        ConflictTypeContradiction,
			Description: fmt.Sprintf("Contradiction detected between %s and %s", result1.AgentID, result2.AgentID),
			Agents:      []string{result1.AgentID, result2.AgentID},
			Severity:    dcd.calculateSeverity(contradictionScore),
			Evidence:    dcd.combineEvidence(result1.Evidence, result2.Evidence),
			Metadata: map[string]interface{}{
				"contradiction_score": contradictionScore,
				"confidence_diff":     abs(result1.Confidence - result2.Confidence),
			},
		}
	}

	// Check for inconsistencies
	inconsistencyScore := dcd.calculateInconsistencyScore(text1, text2)
	if inconsistencyScore > 0.6 {
		return &Conflict{
			ID:          generateConflictID(result1.AgentID, result2.AgentID),
			Type:        ConflictTypeInconsistency,
			Description: fmt.Sprintf("Inconsistency detected between %s and %s", result1.AgentID, result2.AgentID),
			Agents:      []string{result1.AgentID, result2.AgentID},
			Severity:    dcd.calculateSeverity(inconsistencyScore),
			Evidence:    dcd.combineEvidence(result1.Evidence, result2.Evidence),
			Metadata: map[string]interface{}{
				"inconsistency_score": inconsistencyScore,
			},
		}
	}

	return nil
}

// extractText extracts text content from result output
func (dcd *DefaultConflictDetector) extractText(output interface{}) string {
	if output == nil {
		return ""
	}

	if str, ok := output.(string); ok {
		return str
	}

	return fmt.Sprintf("%v", output)
}

// calculateContradictionScore calculates how contradictory two texts are
func (dcd *DefaultConflictDetector) calculateContradictionScore(text1, text2 string) float64 {
	lower1 := strings.ToLower(text1)
	lower2 := strings.ToLower(text2)

	// Check for direct contradictions
	contradictionPairs := [][2]string{
		{"yes", "no"}, {"true", "false"}, {"correct", "incorrect"},
		{"works", "doesn't work"}, {"valid", "invalid"}, {"secure", "insecure"},
		{"fast", "slow"}, {"good", "bad"}, {"safe", "unsafe"},
	}

	for _, pair := range contradictionPairs {
		if (strings.Contains(lower1, pair[0]) && strings.Contains(lower2, pair[1])) ||
			(strings.Contains(lower1, pair[1]) && strings.Contains(lower2, pair[0])) {
			return 1.0 // Direct contradiction
		}
	}

	// Check for opposing concepts
	opposingConcepts := map[string][]string{
		"yes":    {"no", "not", "never", "cannot"},
		"true":   {"false", "incorrect", "wrong"},
		"works":  {"fails", "broken", "error"},
		"secure": {"vulnerable", "insecure", "unsafe"},
		"fast":   {"slow", "performance", "bottleneck"},
	}

	score := 0.0
	for concept, opposites := range opposingConcepts {
		if strings.Contains(lower1, concept) {
			for _, opposite := range opposites {
				if strings.Contains(lower2, opposite) {
					score += 0.3
				}
			}
		}
		if strings.Contains(lower2, concept) {
			for _, opposite := range opposites {
				if strings.Contains(lower1, opposite) {
					score += 0.3
				}
			}
		}
	}

	return min(score, 1.0)
}

// calculateInconsistencyScore calculates inconsistency between texts
func (dcd *DefaultConflictDetector) calculateInconsistencyScore(text1, text2 string) float64 {
	// Simple implementation - could be enhanced with NLP
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	// Find common words
	commonWords := 0
	wordMap := make(map[string]bool)
	for _, word := range words2 {
		wordMap[word] = true
	}

	for _, word := range words1 {
		if wordMap[word] {
			commonWords++
		}
	}

	// Calculate Jaccard similarity
	totalUniqueWords := len(words1) + len(words2) - commonWords
	if totalUniqueWords == 0 {
		return 0.0
	}

	similarity := float64(commonWords) / float64(totalUniqueWords)

	// Inconsistency is inverse of similarity (but not exactly)
	inconsistency := 1.0 - similarity

	// Boost inconsistency if there are contradictory elements
	contradictionScore := dcd.calculateContradictionScore(text1, text2)
	inconsistency = (inconsistency + contradictionScore) / 2.0

	return min(inconsistency, 1.0)
}

// calculateSeverity calculates conflict severity based on score
func (dcd *DefaultConflictDetector) calculateSeverity(score float64) ConflictSeverity {
	if score > 0.8 {
		return SeverityCritical
	} else if score > 0.6 {
		return SeverityHigh
	} else if score > 0.4 {
		return SeverityMedium
	}
	return SeverityLow
}

// combineEvidence combines evidence from multiple agents
func (dcd *DefaultConflictDetector) combineEvidence(evidence1, evidence2 []Evidence) []Evidence {
	combined := make([]Evidence, 0, len(evidence1)+len(evidence2))
	combined = append(combined, evidence1...)
	combined = append(combined, evidence2...)
	return combined
}

// DefaultEvidenceWeighter implements evidence weighting
type DefaultEvidenceWeighter struct{}

// NewDefaultEvidenceWeighter creates a new default evidence weighter
func NewDefaultEvidenceWeighter() *DefaultEvidenceWeighter {
	return &DefaultEvidenceWeighter{}
}

// WeightEvidence weights evidence based on agent performance and evidence quality
func (dew *DefaultEvidenceWeighter) WeightEvidence(ctx context.Context, evidence []Evidence, agent *RegisteredAgent) (float64, error) {
	if len(evidence) == 0 {
		return 0.0, nil
	}

	totalWeight := 0.0

	for _, ev := range evidence {
		// Base weight from evidence confidence
		weight := ev.Confidence

		// Adjust based on evidence type
		typeWeight := dew.getEvidenceTypeWeight(ev.Type)
		weight *= typeWeight

		// Adjust based on agent performance (if available)
		if agent != nil {
			agentWeight := dew.getAgentEvidenceWeight(agent)
			weight *= agentWeight
		}

		totalWeight += weight
	}

	averageWeight := totalWeight / float64(len(evidence))

	// Normalize to 0-1 range
	if averageWeight > 1.0 {
		averageWeight = 1.0
	}

	return averageWeight, nil
}

// getEvidenceTypeWeight returns weight for evidence types
func (dew *DefaultEvidenceWeighter) getEvidenceTypeWeight(evidenceType string) float64 {
	weights := map[string]float64{
		"analysis":      1.0,
		"code_example":  1.2,
		"test_result":   1.1,
		"documentation": 0.9,
		"discussion":    0.8,
		"configuration": 0.7,
		"performance":   1.1,
		"security_scan": 1.3,
		"error_log":     1.0,
		"stack_trace":   1.1,
		"user_feedback": 0.6,
	}

	if weight, exists := weights[evidenceType]; exists {
		return weight
	}

	return 0.8
}

// getAgentEvidenceWeight returns weight based on agent performance
func (dew *DefaultEvidenceWeighter) getAgentEvidenceWeight(agent *RegisteredAgent) float64 {
	// Find the best performing capability
	bestPerformance := 0.0

	for _, capability := range agent.Capabilities {
		if capability.Performance != nil {
			performance := capability.Performance.SuccessRate*0.6 +
				(1.0-capability.Performance.ErrorRate)*0.4

			if performance > bestPerformance {
				bestPerformance = performance
			}
		}
	}

	if bestPerformance == 0.0 {
		return 0.8 // Default weight
	}

	return bestPerformance
}

// Helper functions

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func generateConflictID(agent1, agent2 string) string {
	return fmt.Sprintf("conflict-%s-%s-%d", agent1, agent2, time.Now().UnixNano())
}
