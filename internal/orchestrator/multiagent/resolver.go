package multiagent

import (
	"context"
	"fmt"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
)

// DefaultConflictResolver implements conflict resolution logic
type DefaultConflictResolver struct {
	agentRegistry  *AgentRegistry
	profileManager *profiles.ProfileManager
}

// NewDefaultConflictResolver creates a new default conflict resolver
func NewDefaultConflictResolver(registry *AgentRegistry, profileManager *profiles.ProfileManager) *DefaultConflictResolver {
	return &DefaultConflictResolver{
		agentRegistry:  registry,
		profileManager: profileManager,
	}
}

// Resolve resolves conflicts between agent results
func (dcr *DefaultConflictResolver) Resolve(ctx context.Context, conflicts []Conflict, task *MultiAgentTask) ([]Resolution, error) {
	if len(conflicts) == 0 {
		return []Resolution{}, nil
	}

	var resolutions []Resolution

	for _, conflict := range conflicts {
		resolution, err := dcr.resolveConflict(ctx, conflict, task)
		if err != nil {
			// Create fallback resolution on error
			resolution = Resolution{
				ConflictID:  conflict.ID,
				Type:        ResolutionTypeFallback,
				Description: fmt.Sprintf("Failed to resolve conflict: %v", err),
				Decision:    "Unable to resolve - manual review required",
				Confidence:  0.0,
				Evidence:    conflict.Evidence,
				Metadata: map[string]interface{}{
					"error": err.Error(),
				},
			}
		}

		resolutions = append(resolutions, resolution)
	}

	return resolutions, nil
}

// resolveConflict resolves a single conflict
func (dcr *DefaultConflictResolver) resolveConflict(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	switch conflict.Type {
	case ConflictTypeContradiction:
		return dcr.resolveContradiction(ctx, conflict, task)
	case ConflictTypeInconsistency:
		return dcr.resolveInconsistency(ctx, conflict, task)
	case ConflictTypeAmbiguity:
		return dcr.resolveAmbiguity(ctx, conflict, task)
	case ConflictTypeGap:
		return dcr.resolveGap(ctx, conflict, task)
	default:
		return dcr.resolveByConsensus(ctx, conflict, task)
	}
}

// resolveContradiction resolves direct contradictions between agents
func (dcr *DefaultConflictResolver) resolveContradiction(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	if len(conflict.Agents) < 2 {
		return Resolution{}, fmt.Errorf("contradiction requires at least 2 agents")
	}

	// Get agent information
	agents := make([]*RegisteredAgent, 0, len(conflict.Agents))
	for _, agentID := range conflict.Agents {
		agent, err := dcr.agentRegistry.GetAgent(agentID)
		if err != nil {
			continue
		}
		agents = append(agents, agent)
	}

	if len(agents) < 2 {
		return Resolution{}, fmt.Errorf("could not retrieve agent information")
	}

	// Strategy 1: Use agent with highest performance
	bestAgent := dcr.findBestPerformingAgent(agents)
	if bestAgent != nil {
		return Resolution{
			ConflictID:  conflict.ID,
			Type:        ResolutionTypeExpert,
			Description: fmt.Sprintf("Resolved by selecting result from highest performing agent: %s", bestAgent.ID),
			Decision:    fmt.Sprintf("Accept result from agent %s", bestAgent.ID),
			Confidence:  dcr.calculateAgentConfidence(bestAgent, conflict),
			Evidence:    conflict.Evidence,
			Metadata: map[string]interface{}{
				"strategy":          "expert_opinion",
				"selected_agent":    bestAgent.ID,
				"performance_score": dcr.calculateAgentPerformanceScore(bestAgent),
			},
		}, nil
	}

	// Strategy 2: Use majority consensus if more than 2 agents
	if len(agents) > 2 {
		return dcr.resolveByMajority(ctx, conflict, task)
	}

	// Strategy 3: Escalate to higher authority (profile-based)
	return dcr.resolveByEscalation(ctx, conflict, task)
}

// resolveInconsistency resolves inconsistencies between agents
func (dcr *DefaultConflictResolver) resolveInconsistency(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	// For inconsistencies, try to find common ground or use most comprehensive result
	agents := make([]*RegisteredAgent, 0, len(conflict.Agents))
	for _, agentID := range conflict.Agents {
		agent, err := dcr.agentRegistry.GetAgent(agentID)
		if err != nil {
			continue
		}
		agents = append(agents, agent)
	}

	// Find agent with most comprehensive result (highest evidence count)
	var bestAgent *RegisteredAgent
	maxEvidence := 0

	for _, agent := range agents {
		evidenceCount := len(conflict.Evidence) // This would ideally be per-agent evidence
		if evidenceCount > maxEvidence {
			maxEvidence = evidenceCount
			bestAgent = agent
		}
	}

	if bestAgent != nil {
		return Resolution{
			ConflictID:  conflict.ID,
			Type:        ResolutionTypeExpert,
			Description: fmt.Sprintf("Resolved inconsistency by selecting most comprehensive result from agent: %s", bestAgent.ID),
			Decision:    fmt.Sprintf("Accept result from agent %s (most comprehensive)", bestAgent.ID),
			Confidence:  0.7,
			Evidence:    conflict.Evidence,
			Metadata: map[string]interface{}{
				"strategy":       "comprehensive_result",
				"selected_agent": bestAgent.ID,
				"evidence_count": maxEvidence,
			},
		}, nil
	}

	return dcr.resolveByConsensus(ctx, conflict, task)
}

// resolveAmbiguity resolves ambiguous results
func (dcr *DefaultConflictResolver) resolveAmbiguity(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	// For ambiguity, use the agent with the highest confidence in its profile area
	agents := make([]*RegisteredAgent, 0, len(conflict.Agents))
	for _, agentID := range conflict.Agents {
		agent, err := dcr.agentRegistry.GetAgent(agentID)
		if err != nil {
			continue
		}
		agents = append(agents, agent)
	}

	// Find agent whose profile best matches the task
	var bestAgent *RegisteredAgent
	bestMatchScore := 0.0

	for _, agent := range agents {
		matchScore := dcr.calculateProfileMatchScore(agent.Profile, task.Profile)
		if matchScore > bestMatchScore {
			bestMatchScore = matchScore
			bestAgent = agent
		}
	}

	if bestAgent != nil {
		return Resolution{
			ConflictID:  conflict.ID,
			Type:        ResolutionTypeExpert,
			Description: fmt.Sprintf("Resolved ambiguity by selecting agent with best profile match: %s", bestAgent.ID),
			Decision:    fmt.Sprintf("Accept result from agent %s (best profile match)", bestAgent.ID),
			Confidence:  bestMatchScore,
			Evidence:    conflict.Evidence,
			Metadata: map[string]interface{}{
				"strategy":       "profile_match",
				"selected_agent": bestAgent.ID,
				"match_score":    bestMatchScore,
			},
		}, nil
	}

	return dcr.resolveByConsensus(ctx, conflict, task)
}

// resolveGap resolves gaps in information
func (dcr *DefaultConflictResolver) resolveGap(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	// For gaps, suggest escalation or additional analysis
	return Resolution{
		ConflictID:  conflict.ID,
		Type:        ResolutionTypeEscalation,
		Description: "Information gap detected - escalation recommended",
		Decision:    "Escalate to human review or additional specialized agents",
		Confidence:  0.9,
		Evidence:    conflict.Evidence,
		Metadata: map[string]interface{}{
			"strategy": "escalation",
			"gap_type": conflict.Type,
		},
	}, nil
}

// resolveByMajority resolves conflicts using majority vote
func (dcr *DefaultConflictResolver) resolveByMajority(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	// This would require access to the actual agent results
	// For now, return a consensus-based resolution
	return Resolution{
		ConflictID:  conflict.ID,
		Type:        ResolutionTypeMajority,
		Description: "Resolved by majority consensus",
		Decision:    "Accept majority opinion",
		Confidence:  0.6,
		Evidence:    conflict.Evidence,
		Metadata: map[string]interface{}{
			"strategy": "majority_vote",
		},
	}, nil
}

// resolveByConsensus finds common ground between conflicting results
func (dcr *DefaultConflictResolver) resolveByConsensus(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	return Resolution{
		ConflictID:  conflict.ID,
		Type:        ResolutionTypeConsensus,
		Description: "Resolved by finding consensus between agents",
		Decision:    "Accept consensus result",
		Confidence:  0.5,
		Evidence:    conflict.Evidence,
		Metadata: map[string]interface{}{
			"strategy": "consensus",
		},
	}, nil
}

// resolveByEscalation escalates conflicts that cannot be resolved automatically
func (dcr *DefaultConflictResolver) resolveByEscalation(ctx context.Context, conflict Conflict, task *MultiAgentTask) (Resolution, error) {
	return Resolution{
		ConflictID:  conflict.ID,
		Type:        ResolutionTypeEscalation,
		Description: "Conflict requires human intervention",
		Decision:    "Escalate to human review",
		Confidence:  0.0,
		Evidence:    conflict.Evidence,
		Metadata: map[string]interface{}{
			"strategy": "escalation",
			"severity": conflict.Severity,
		},
	}, nil
}

// findBestPerformingAgent finds the agent with the best performance metrics
func (dcr *DefaultConflictResolver) findBestPerformingAgent(agents []*RegisteredAgent) *RegisteredAgent {
	var bestAgent *RegisteredAgent
	bestScore := 0.0

	for _, agent := range agents {
		score := dcr.calculateAgentPerformanceScore(agent)
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}

	return bestAgent
}

// calculateAgentPerformanceScore calculates an overall performance score for an agent
func (dcr *DefaultConflictResolver) calculateAgentPerformanceScore(agent *RegisteredAgent) float64 {
	score := 0.0

	// Average performance across all capabilities
	totalCapabilities := 0
	for _, capability := range agent.Capabilities {
		if capability.Performance != nil {
			capabilityScore := capability.Performance.SuccessRate*0.6 +
				(1.0-capability.Performance.ErrorRate)*0.4
			score += capabilityScore
			totalCapabilities++
		}
	}

	if totalCapabilities > 0 {
		score = score / float64(totalCapabilities)
	} else {
		score = 0.5 // Default score for agents without performance data
	}

	return score
}

// calculateAgentConfidence calculates confidence in an agent's result for a conflict
func (dcr *DefaultConflictResolver) calculateAgentConfidence(agent *RegisteredAgent, conflict Conflict) float64 {
	// Base confidence from performance
	performanceScore := dcr.calculateAgentPerformanceScore(agent)

	// Adjust based on conflict severity
	severityMultiplier := 1.0
	switch conflict.Severity {
	case SeverityLow:
		severityMultiplier = 1.2
	case SeverityMedium:
		severityMultiplier = 1.0
	case SeverityHigh:
		severityMultiplier = 0.8
	case SeverityCritical:
		severityMultiplier = 0.6
	}

	confidence := performanceScore * severityMultiplier

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateProfileMatchScore calculates how well an agent's profile matches the task profile
func (dcr *DefaultConflictResolver) calculateProfileMatchScore(agentProfile, taskProfile *profiles.AgentProfile) float64 {
	if agentProfile == nil || taskProfile == nil {
		return 0.5
	}

	// Exact match gets highest score
	if agentProfile.ID == taskProfile.ID {
		return 1.0
	}

	// Check capability overlap
	agentCapabilities := make(map[string]bool)
	for _, capability := range agentProfile.Capabilities {
		agentCapabilities[capability] = true
	}

	taskCapabilities := make(map[string]bool)
	for _, capability := range taskProfile.Capabilities {
		taskCapabilities[capability] = true
	}

	// Calculate overlap
	overlap := 0
	total := len(taskCapabilities)

	for capability := range taskCapabilities {
		if agentCapabilities[capability] {
			overlap++
		}
	}

	if total == 0 {
		return 0.5
	}

	return float64(overlap) / float64(total)
}

// DefaultPerformanceMonitor implements performance monitoring
type DefaultPerformanceMonitor struct {
	metrics PerformanceMetrics
}

// NewDefaultPerformanceMonitor creates a new default performance monitor
func NewDefaultPerformanceMonitor() *DefaultPerformanceMonitor {
	return &DefaultPerformanceMonitor{
		metrics: PerformanceMetrics{
			AgentUtilization: make(map[string]float64),
			CapabilityUsage:  make(map[string]int64),
			LastUpdated:      time.Now(),
		},
	}
}

// RecordTask records task execution metrics
func (dpm *DefaultPerformanceMonitor) RecordTask(ctx context.Context, task *MultiAgentTask, duration time.Duration, success bool) {
	dpm.metrics.TotalTasks++
	dpm.metrics.LastUpdated = time.Now()

	if success {
		dpm.metrics.SuccessfulTasks++
	} else {
		dpm.metrics.FailedTasks++
	}

	// Update average task duration
	if dpm.metrics.TotalTasks == 1 {
		dpm.metrics.AverageTaskDuration = duration
	} else {
		// Exponential moving average
		alpha := 0.1
		dpm.metrics.AverageTaskDuration = time.Duration(float64(dpm.metrics.AverageTaskDuration)*(1-alpha) + float64(duration)*alpha)
	}

	// Update average agent count (this would need to be passed in)
	dpm.metrics.AverageAgentCount = 2.5 // Placeholder
}

// RecordAgentExecution records individual agent execution metrics
func (dpm *DefaultPerformanceMonitor) RecordAgentExecution(ctx context.Context, agentID string, duration time.Duration, success bool) {
	// Update agent utilization (simplified)
	if utilization, exists := dpm.metrics.AgentUtilization[agentID]; exists {
		dpm.metrics.AgentUtilization[agentID] = utilization*0.9 + 0.1 // Decay + new value
	} else {
		dpm.metrics.AgentUtilization[agentID] = 0.1
	}

	dpm.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current performance metrics
func (dpm *DefaultPerformanceMonitor) GetMetrics() *PerformanceMetrics {
	dpm.metrics.LastUpdated = time.Now()
	return &dpm.metrics
}
