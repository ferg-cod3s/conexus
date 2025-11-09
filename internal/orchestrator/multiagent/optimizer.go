package multiagent

import (
	"context"
	"sync"
	"time"
)

// CoordinationOptimizer optimizes multi-agent coordination performance
type CoordinationOptimizer struct {
	agentPerformanceCache map[string]*AgentPerformanceCache
	coordinationPatterns  map[string]*CoordinationPattern
	performanceMetrics    *CoordinationPerformanceMetrics
	mu                    sync.RWMutex
	optimizationInterval  time.Duration
	lastOptimization      time.Time
}

// AgentPerformanceCache caches agent performance data
type AgentPerformanceCache struct {
	AgentID        string        `json:"agent_id"`
	AverageLatency time.Duration `json:"average_latency"`
	SuccessRate    float64       `json:"success_rate"`
	CurrentLoad    int           `json:"current_load"`
	MaxLoad        int           `json:"max_load"`
	LastUpdated    time.Time     `json:"last_updated"`
	CacheHits      int64         `json:"cache_hits"`
	CacheMisses    int64         `json:"cache_misses"`
}

// CoordinationPattern represents learned coordination patterns
type CoordinationPattern struct {
	PatternID      string                 `json:"pattern_id"`
	Description    string                 `json:"description"`
	AgentSequence  []string               `json:"agent_sequence"`
	SuccessRate    float64                `json:"success_rate"`
	AverageLatency time.Duration          `json:"average_latency"`
	UsageCount     int                    `json:"usage_count"`
	ContextFactors map[string]interface{} `json:"context_factors"`
	LastUsed       time.Time              `json:"last_used"`
}

// CoordinationPerformanceMetrics tracks coordination performance
type CoordinationPerformanceMetrics struct {
	TotalTasks         int64              `json:"total_tasks"`
	SuccessfulTasks    int64              `json:"successful_tasks"`
	FailedTasks        int64              `json:"failed_tasks"`
	AverageTaskLatency time.Duration      `json:"average_task_latency"`
	AverageAgentCount  float64            `json:"average_agent_count"`
	ConflictRate       float64            `json:"conflict_rate"`
	ResolutionRate     float64            `json:"resolution_rate"`
	CacheHitRate       float64            `json:"cache_hit_rate"`
	OptimizationGain   float64            `json:"optimization_gain"`
	AgentUtilization   map[string]float64 `json:"agent_utilization"`
	PatternEfficiency  map[string]float64 `json:"pattern_efficiency"`
	LastUpdated        time.Time          `json:"last_updated"`
}

// NewCoordinationOptimizer creates a new coordination optimizer
func NewCoordinationOptimizer() *CoordinationOptimizer {
	return &CoordinationOptimizer{
		agentPerformanceCache: make(map[string]*AgentPerformanceCache),
		coordinationPatterns:  make(map[string]*CoordinationPattern),
		performanceMetrics: &CoordinationPerformanceMetrics{
			AgentUtilization:  make(map[string]float64),
			PatternEfficiency: make(map[string]float64),
			LastUpdated:       time.Now(),
		},
		optimizationInterval: 5 * time.Minute,
		lastOptimization:     time.Now(),
	}
}

// OptimizeTask optimizes a multi-agent task for performance
func (co *CoordinationOptimizer) OptimizeTask(ctx context.Context, task *MultiAgentTask) (*OptimizedTask, error) {
	co.mu.RLock()
	defer co.mu.RUnlock()

	// Find optimal coordination pattern
	pattern := co.findOptimalPattern(ctx, task)
	if pattern != nil {
		return co.applyPatternOptimization(ctx, task, pattern)
	}

	// Fall back to agent-based optimization
	return co.optimizeAgentSelection(ctx, task)
}

// UpdateAgentPerformance updates cached agent performance
func (co *CoordinationOptimizer) UpdateAgentPerformance(ctx context.Context, agentID string, latency time.Duration, success bool, load int) {
	co.mu.Lock()
	defer co.mu.Unlock()

	cache, exists := co.agentPerformanceCache[agentID]
	if !exists {
		cache = &AgentPerformanceCache{
			AgentID:     agentID,
			MaxLoad:     10, // Default max load
			LastUpdated: time.Now(),
		}
		co.agentPerformanceCache[agentID] = cache
	}

	// Update performance metrics
	cache.LastUpdated = time.Now()
	cache.CurrentLoad = load

	// Update average latency
	if cache.CacheHits+cache.CacheMisses == 0 {
		cache.AverageLatency = latency
	} else {
		// Exponential moving average
		alpha := 0.1
		cache.AverageLatency = time.Duration(float64(cache.AverageLatency)*(1-alpha) + float64(latency)*alpha)
	}

	// Update success rate
	if success {
		cache.SuccessRate = (cache.SuccessRate + 1.0) / 2.0
	} else {
		cache.SuccessRate = cache.SuccessRate / 2.0
	}

	// Update cache statistics
	if latency < 100*time.Millisecond {
		cache.CacheHits++
	} else {
		cache.CacheMisses++
	}
}

// RecordTaskCompletion records task completion metrics
func (co *CoordinationOptimizer) RecordTaskCompletion(ctx context.Context, task *MultiAgentTask, duration time.Duration, success bool, agentCount int, conflicts int, resolutions int) {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.performanceMetrics.TotalTasks++
	co.performanceMetrics.LastUpdated = time.Now()

	if success {
		co.performanceMetrics.SuccessfulTasks++
	} else {
		co.performanceMetrics.FailedTasks++
	}

	// Update average task latency
	if co.performanceMetrics.TotalTasks == 1 {
		co.performanceMetrics.AverageTaskLatency = duration
	} else {
		alpha := 0.1
		co.performanceMetrics.AverageTaskLatency = time.Duration(float64(co.performanceMetrics.AverageTaskLatency)*(1-alpha) + float64(duration)*alpha)
	}

	// Update average agent count
	if co.performanceMetrics.TotalTasks == 1 {
		co.performanceMetrics.AverageAgentCount = float64(agentCount)
	} else {
		alpha := 0.1
		co.performanceMetrics.AverageAgentCount = co.performanceMetrics.AverageAgentCount*(1-alpha) + float64(agentCount)*alpha
	}

	// Update conflict and resolution rates
	if conflicts > 0 {
		co.performanceMetrics.ConflictRate = (co.performanceMetrics.ConflictRate + 1.0) / 2.0
		if resolutions > 0 {
			co.performanceMetrics.ResolutionRate = (co.performanceMetrics.ResolutionRate + 1.0) / 2.0
		}
	}

	// Update optimization gain (estimated improvement)
	co.performanceMetrics.OptimizationGain = co.calculateOptimizationGain()
}

// GetOptimalAgentSequence suggests optimal agent execution sequence
func (co *CoordinationOptimizer) GetOptimalAgentSequence(ctx context.Context, requiredCapabilities []string, context map[string]interface{}) ([]string, error) {
	co.mu.RLock()
	defer co.mu.RUnlock()

	// Score agents for each capability
	agentScores := make(map[string]float64)

	for _, capability := range requiredCapabilities {
		agents := co.getAgentsByCapability(capability)
		for _, agent := range agents {
			score := co.calculateAgentScore(agent, context)
			if score > agentScores[agent.ID] {
				agentScores[agent.ID] = score
			}
		}
	}

	// Sort agents by score
	type agentScore struct {
		id    string
		score float64
	}

	var sortedAgents []agentScore
	for id, score := range agentScores {
		sortedAgents = append(sortedAgents, agentScore{id, score})
	}

	// Simple sort by score (descending)
	for i := 0; i < len(sortedAgents); i++ {
		for j := i + 1; j < len(sortedAgents); j++ {
			if sortedAgents[i].score < sortedAgents[j].score {
				sortedAgents[i], sortedAgents[j] = sortedAgents[j], sortedAgents[i]
			}
		}
	}

	// Extract agent IDs
	var sequence []string
	for _, agent := range sortedAgents {
		sequence = append(sequence, agent.id)
	}

	return sequence, nil
}

// GetPerformanceMetrics returns current performance metrics
func (co *CoordinationOptimizer) GetPerformanceMetrics() *CoordinationPerformanceMetrics {
	co.mu.RLock()
	defer co.mu.RUnlock()

	// Return a copy
	metrics := *co.performanceMetrics
	metrics.LastUpdated = time.Now()
	return &metrics
}

// findOptimalPattern finds the optimal coordination pattern for a task
func (co *CoordinationOptimizer) findOptimalPattern(ctx context.Context, task *MultiAgentTask) *CoordinationPattern {
	// Simple pattern matching based on query and profile
	patternKey := co.generatePatternKey(task)

	if pattern, exists := co.coordinationPatterns[patternKey]; exists {
		// Check if pattern is still effective
		if pattern.SuccessRate > 0.7 && time.Since(pattern.LastUsed) < 24*time.Hour {
			return pattern
		}
	}

	return nil
}

// applyPatternOptimization applies optimization from a coordination pattern
func (co *CoordinationOptimizer) applyPatternOptimization(ctx context.Context, task *MultiAgentTask, pattern *CoordinationPattern) (*OptimizedTask, error) {
	optimized := &OptimizedTask{
		OriginalTask:       task,
		OptimizedSequence:  pattern.AgentSequence,
		EstimatedLatency:   pattern.AverageLatency,
		SuccessProbability: pattern.SuccessRate,
		Optimizations: []string{
			"pattern_based_optimization",
			"predefined_sequence",
		},
		Metadata: map[string]interface{}{
			"pattern_id":  pattern.PatternID,
			"usage_count": pattern.UsageCount,
		},
	}

	return optimized, nil
}

// optimizeAgentSelection optimizes agent selection for a task
func (co *CoordinationOptimizer) optimizeAgentSelection(ctx context.Context, task *MultiAgentTask) (*OptimizedTask, error) {
	// Extract required capabilities from query
	capabilities := co.extractCapabilitiesFromQuery(task.Query)

	// Get optimal agent sequence
	sequence, err := co.GetOptimalAgentSequence(ctx, capabilities, task.Context)
	if err != nil {
		return nil, err
	}

	// Estimate performance
	estimatedLatency := co.estimateTaskLatency(sequence)
	successProbability := co.estimateSuccessProbability(sequence, task)

	optimized := &OptimizedTask{
		OriginalTask:       task,
		OptimizedSequence:  sequence,
		EstimatedLatency:   estimatedLatency,
		SuccessProbability: successProbability,
		Optimizations: []string{
			"agent_selection_optimization",
			"performance_based_ordering",
		},
		Metadata: map[string]interface{}{
			"capabilities": capabilities,
			"agent_count":  len(sequence),
		},
	}

	return optimized, nil
}

// calculateAgentScore calculates a performance score for an agent
func (co *CoordinationOptimizer) calculateAgentScore(agent *RegisteredAgent, context map[string]interface{}) float64 {
	score := 0.0

	// Base score from performance cache
	if cache, exists := co.agentPerformanceCache[agent.ID]; exists {
		score += cache.SuccessRate * 0.6
		score += (1.0 - float64(cache.CurrentLoad)/float64(cache.MaxLoad)) * 0.4
	} else {
		score = 0.5 // Default score
	}

	// Adjust based on context relevance
	if context != nil {
		contextScore := co.calculateContextRelevance(agent, context)
		score += contextScore * 0.2
	}

	return score
}

// calculateContextRelevance calculates how relevant an agent is for the context
func (co *CoordinationOptimizer) calculateContextRelevance(agent *RegisteredAgent, context map[string]interface{}) float64 {
	relevance := 0.0

	// Check if agent's profile matches context requirements
	if context["profile_id"] != nil {
		if profileID, ok := context["profile_id"].(string); ok {
			if agent.Profile.ID == profileID {
				relevance += 0.5
			}
		}
	}

	// Check work context relevance
	if _, exists := context["active_file"]; exists {
		for _, capability := range agent.Capabilities {
			if capability.Category == "analysis" {
				relevance += 0.3
				break
			}
		}
	}

	return relevance
}

// getAgentsByCapability gets agents with a specific capability (placeholder)
func (co *CoordinationOptimizer) getAgentsByCapability(capability string) []*RegisteredAgent {
	// This would integrate with the agent registry
	// For now, return empty slice
	return []*RegisteredAgent{}
}

// generatePatternKey generates a key for pattern matching
func (co *CoordinationOptimizer) generatePatternKey(task *MultiAgentTask) string {
	// Simple key based on profile and query characteristics
	key := task.Profile.ID

	if len(task.Query) > 50 {
		key += "_complex"
	} else {
		key += "_simple"
	}

	if task.Context != nil {
		if task.Context["active_file"] != nil {
			key += "_with_file"
		}
		if task.Context["git_branch"] != nil {
			key += "_with_branch"
		}
	}

	return key
}

// extractCapabilitiesFromQuery extracts required capabilities from query
func (co *CoordinationOptimizer) extractCapabilitiesFromQuery(query string) []string {
	// Simple capability extraction based on keywords
	var capabilities []string

	if contains(query, "debug") || contains(query, "error") || contains(query, "bug") {
		capabilities = append(capabilities, "debugging")
	}

	if contains(query, "security") || contains(query, "vulnerability") || contains(query, "auth") {
		capabilities = append(capabilities, "security")
	}

	if contains(query, "design") || contains(query, "architecture") || contains(query, "system") {
		capabilities = append(capabilities, "architecture")
	}

	if contains(query, "documentation") || contains(query, "explain") || contains(query, "how") {
		capabilities = append(capabilities, "documentation")
	}

	// Default to code analysis
	if len(capabilities) == 0 {
		capabilities = append(capabilities, "code_analysis")
	}

	return capabilities
}

// estimateTaskLatency estimates task completion latency
func (co *CoordinationOptimizer) estimateTaskLatency(sequence []string) time.Duration {
	// Simple estimation based on agent count and average latency
	baseLatency := 100 * time.Millisecond
	agentOverhead := 50 * time.Millisecond

	estimatedLatency := baseLatency + time.Duration(len(sequence))*agentOverhead

	// Apply optimization factor
	optimizationFactor := 0.8 // 20% improvement from optimization
	estimatedLatency = time.Duration(float64(estimatedLatency) * optimizationFactor)

	return estimatedLatency
}

// estimateSuccessProbability estimates success probability
func (co *CoordinationOptimizer) estimateSuccessProbability(sequence []string, task *MultiAgentTask) float64 {
	// Base probability
	probability := 0.8

	// Adjust based on agent count (more agents = more complexity = lower probability)
	agentPenalty := float64(len(sequence)-1) * 0.05
	probability -= agentPenalty

	// Adjust based on profile optimization
	if task.Profile != nil {
		switch task.Profile.ID {
		case "debugging":
			probability += 0.1 // Debugging tasks are often straightforward
		case "security":
			probability += 0.05 // Security tasks are important but complex
		case "architecture":
			probability -= 0.1 // Architecture tasks are complex
		}
	}

	// Ensure probability is in valid range
	if probability > 1.0 {
		probability = 1.0
	}
	if probability < 0.0 {
		probability = 0.0
	}

	return probability
}

// calculateOptimizationGain calculates the performance gain from optimizations
func (co *CoordinationOptimizer) calculateOptimizationGain() float64 {
	// Placeholder implementation
	// In a real system, this would compare optimized vs non-optimized performance
	return 0.25 // 25% improvement
}

// OptimizedTask represents an optimized multi-agent task
type OptimizedTask struct {
	OriginalTask       *MultiAgentTask        `json:"original_task"`
	OptimizedSequence  []string               `json:"optimized_sequence"`
	EstimatedLatency   time.Duration          `json:"estimated_latency"`
	SuccessProbability float64                `json:"success_probability"`
	Optimizations      []string               `json:"optimizations"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// contains checks if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
