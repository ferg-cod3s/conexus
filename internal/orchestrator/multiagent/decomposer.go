package multiagent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
)

// DefaultTaskDecomposer implements task decomposition logic
type DefaultTaskDecomposer struct {
	agentRegistry  *AgentRegistry
	profileManager *profiles.ProfileManager
}

// NewDefaultTaskDecomposer creates a new default task decomposer
func NewDefaultTaskDecomposer(registry *AgentRegistry, profileManager *profiles.ProfileManager) *DefaultTaskDecomposer {
	return &DefaultTaskDecomposer{
		agentRegistry:  registry,
		profileManager: profileManager,
	}
}

// Decompose breaks down a complex task into subtasks
func (dtd *DefaultTaskDecomposer) Decompose(ctx context.Context, task *MultiAgentTask) ([]*SubTask, error) {
	if task.Query == "" {
		return nil, fmt.Errorf("task query cannot be empty")
	}

	// Analyze the query to determine decomposition strategy
	strategy := dtd.analyzeQuery(task.Query, task.Profile)

	switch strategy {
	case DecompositionStrategySimple:
		return dtd.decomposeSimple(ctx, task)
	case DecompositionStrategySequential:
		return dtd.decomposeSequential(ctx, task)
	case DecompositionStrategyParallel:
		return dtd.decomposeParallel(ctx, task)
	case DecompositionStrategyHierarchical:
		return dtd.decomposeHierarchical(ctx, task)
	default:
		return dtd.decomposeSimple(ctx, task)
	}
}

// DecompositionStrategy represents different decomposition approaches
type DecompositionStrategy string

const (
	DecompositionStrategySimple       DecompositionStrategy = "simple"
	DecompositionStrategySequential   DecompositionStrategy = "sequential"
	DecompositionStrategyParallel     DecompositionStrategy = "parallel"
	DecompositionStrategyHierarchical DecompositionStrategy = "hierarchical"
)

// analyzeQuery determines the best decomposition strategy
func (dtd *DefaultTaskDecomposer) analyzeQuery(query string, profile *profiles.AgentProfile) DecompositionStrategy {
	lowerQuery := strings.ToLower(query)

	// Check for sequential patterns
	sequentialPatterns := []string{
		"first", "then", "next", "after", "before", "step by step",
		"analyze and then", "find and then", "search and then",
	}

	for _, pattern := range sequentialPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return DecompositionStrategySequential
		}
	}

	// Check for parallel patterns
	parallelPatterns := []string{
		"and", "also", "both", "multiple", "different aspects",
		"compare", "analyze from different angles", "various perspectives",
	}

	parallelCount := 0
	for _, pattern := range parallelPatterns {
		if strings.Contains(lowerQuery, pattern) {
			parallelCount++
		}
	}

	if parallelCount >= 2 {
		return DecompositionStrategyParallel
	}

	// Check for hierarchical patterns
	hierarchicalPatterns := []string{
		"comprehensive", "complete", "full", "detailed", "thorough",
		"system", "architecture", "overview", "analysis",
	}

	for _, pattern := range hierarchicalPatterns {
		if strings.Contains(lowerQuery, pattern) {
			return DecompositionStrategyHierarchical
		}
	}

	// Default to simple
	return DecompositionStrategySimple
}

// decomposeSimple creates a single subtask for simple queries
func (dtd *DefaultTaskDecomposer) decomposeSimple(ctx context.Context, task *MultiAgentTask) ([]*SubTask, error) {
	// Find the best agent for the task
	agent, err := dtd.findBestAgentForTask(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to find suitable agent: %w", err)
	}

	subtask := &SubTask{
		ID:           generateSubTaskID(task.ID),
		TaskID:       task.ID,
		AgentID:      agent.ID,
		Capability:   dtd.getPrimaryCapability(task.Query, agent),
		Query:        task.Query,
		Context:      task.Context,
		Requirements: task.Requirements,
		Priority:     task.Priority,
		Timeout:      dtd.calculateTimeout(task.Priority, task.Profile),
		Metadata:     task.Metadata,
	}

	return []*SubTask{subtask}, nil
}

// decomposeSequential creates subtasks that must be executed in sequence
func (dtd *DefaultTaskDecomposer) decomposeSequential(ctx context.Context, task *MultiAgentTask) ([]*SubTask, error) {
	// Parse sequential steps from query
	steps := dtd.parseSequentialSteps(task.Query)
	if len(steps) == 0 {
		return dtd.decomposeSimple(ctx, task)
	}

	var subtasks []*SubTask

	for i, step := range steps {
		// Find agent for this step
		agent, err := dtd.findBestAgentForQuery(ctx, step)
		if err != nil {
			continue // Skip if no agent found
		}

		subtask := &SubTask{
			ID:           generateSubTaskID(task.ID, i),
			TaskID:       task.ID,
			AgentID:      agent.ID,
			Capability:   dtd.getPrimaryCapability(step, agent),
			Query:        step,
			Context:      task.Context,
			Requirements: task.Requirements,
			Priority:     task.Priority,
			Dependencies: dtd.getDependencies(i, len(steps)),
			Timeout:      dtd.calculateTimeout(task.Priority, task.Profile),
			Metadata: map[string]interface{}{
				"step":        i + 1,
				"total_steps": len(steps),
				"sequential":  true,
			},
		}

		subtasks = append(subtasks, subtask)
	}

	if len(subtasks) == 0 {
		return dtd.decomposeSimple(ctx, task)
	}

	return subtasks, nil
}

// decomposeParallel creates subtasks that can be executed in parallel
func (dtd *DefaultTaskDecomposer) decomposeParallel(ctx context.Context, task *MultiAgentTask) ([]*SubTask, error) {
	// Parse parallel aspects from query
	aspects := dtd.parseParallelAspects(task.Query)
	if len(aspects) <= 1 {
		return dtd.decomposeSimple(ctx, task)
	}

	var subtasks []*SubTask

	for i, aspect := range aspects {
		// Find agent for this aspect
		agent, err := dtd.findBestAgentForQuery(ctx, aspect)
		if err != nil {
			continue // Skip if no agent found
		}

		subtask := &SubTask{
			ID:           generateSubTaskID(task.ID, i),
			TaskID:       task.ID,
			AgentID:      agent.ID,
			Capability:   dtd.getPrimaryCapability(aspect, agent),
			Query:        aspect,
			Context:      task.Context,
			Requirements: task.Requirements,
			Priority:     task.Priority,
			Timeout:      dtd.calculateTimeout(task.Priority, task.Profile),
			Metadata: map[string]interface{}{
				"aspect":        i + 1,
				"total_aspects": len(aspects),
				"parallel":      true,
			},
		}

		subtasks = append(subtasks, subtask)
	}

	if len(subtasks) == 0 {
		return dtd.decomposeSimple(ctx, task)
	}

	return subtasks, nil
}

// decomposeHierarchical creates a hierarchical decomposition with multiple levels
func (dtd *DefaultTaskDecomposer) decomposeHierarchical(ctx context.Context, task *MultiAgentTask) ([]*SubTask, error) {
	// For hierarchical tasks, create multiple subtasks covering different aspects
	aspects := []string{
		"overview and structure",
		"implementation details",
		"dependencies and relationships",
		"performance and optimization",
		"security and validation",
	}

	var subtasks []*SubTask

	for i, aspect := range aspects {
		// Find agent for this aspect
		agent, err := dtd.findBestAgentForQuery(ctx, aspect)
		if err != nil {
			continue // Skip if no agent found
		}

		subtask := &SubTask{
			ID:           generateSubTaskID(task.ID, i),
			TaskID:       task.ID,
			AgentID:      agent.ID,
			Capability:   dtd.getPrimaryCapability(aspect, agent),
			Query:        fmt.Sprintf("%s: %s", task.Query, aspect),
			Context:      task.Context,
			Requirements: task.Requirements,
			Priority:     task.Priority,
			Timeout:      dtd.calculateTimeout(task.Priority, task.Profile),
			Metadata: map[string]interface{}{
				"aspect":       aspect,
				"hierarchical": true,
				"level":        1,
			},
		}

		subtasks = append(subtasks, subtask)
	}

	if len(subtasks) == 0 {
		return dtd.decomposeSimple(ctx, task)
	}

	return subtasks, nil
}

// findBestAgentForTask finds the best agent for a complete task
func (dtd *DefaultTaskDecomposer) findBestAgentForTask(ctx context.Context, task *MultiAgentTask) (*RegisteredAgent, error) {
	// Determine required capabilities based on query and profile
	capabilities := dtd.extractRequiredCapabilities(task.Query, task.Profile)

	for _, capability := range capabilities {
		agent, err := dtd.agentRegistry.FindBestAgent(ctx, capability, task.Requirements)
		if err == nil {
			return agent, nil
		}
	}

	return nil, fmt.Errorf("no suitable agent found for task")
}

// findBestAgentForQuery finds the best agent for a specific query
func (dtd *DefaultTaskDecomposer) findBestAgentForQuery(ctx context.Context, query string) (*RegisteredAgent, error) {
	// Determine required capabilities based on query
	capabilities := dtd.extractRequiredCapabilities(query, nil)

	for _, capability := range capabilities {
		agents := dtd.agentRegistry.GetAgentsByCapability(capability)
		if len(agents) > 0 {
			// Return the first available agent
			for _, agent := range agents {
				if agent.Status == AgentStatusAvailable {
					return agent, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no suitable agent found for query: %s", query)
}

// extractRequiredCapabilities extracts required capabilities from query and profile
func (dtd *DefaultTaskDecomposer) extractRequiredCapabilities(query string, profile *profiles.AgentProfile) []string {
	var capabilities []string
	lowerQuery := strings.ToLower(query)

	// Map query patterns to capabilities
	capabilityPatterns := map[string][]string{
		"code_analysis": {
			"function", "class", "method", "implementation", "algorithm",
			"code", "programming", "syntax", "logic", "structure",
		},
		"documentation": {
			"documentation", "readme", "guide", "tutorial", "explain",
			"overview", "introduction", "manual", "reference",
		},
		"debugging": {
			"error", "bug", "issue", "problem", "debug", "fix",
			"troubleshoot", "exception", "crash", "fail",
		},
		"architecture": {
			"architecture", "design", "system", "structure", "pattern",
			"framework", "module", "component", "integration",
		},
		"security": {
			"security", "authentication", "authorization", "vulnerability",
			"encryption", "password", "token", "validation",
		},
	}

	for capability, patterns := range capabilityPatterns {
		for _, pattern := range patterns {
			if strings.Contains(lowerQuery, pattern) {
				capabilities = append(capabilities, capability)
				break
			}
		}
	}

	// If no capabilities found, use profile-based defaults
	if len(capabilities) == 0 && profile != nil {
		switch profile.ID {
		case "code_analysis":
			capabilities = append(capabilities, "code_analysis")
		case "documentation":
			capabilities = append(capabilities, "documentation")
		case "debugging":
			capabilities = append(capabilities, "debugging")
		case "architecture":
			capabilities = append(capabilities, "architecture")
		case "security":
			capabilities = append(capabilities, "security")
		default:
			capabilities = append(capabilities, "general_analysis")
		}
	}

	// Default fallback
	if len(capabilities) == 0 {
		capabilities = append(capabilities, "general_analysis")
	}

	return capabilities
}

// getPrimaryCapability gets the primary capability for an agent based on query
func (dtd *DefaultTaskDecomposer) getPrimaryCapability(query string, agent *RegisteredAgent) string {
	// Find the most relevant capability for the query
	lowerQuery := strings.ToLower(query)

	for _, capability := range agent.Capabilities {
		for _, skill := range capability.Skills {
			if strings.Contains(lowerQuery, strings.ToLower(skill)) {
				return capability.ID
			}
		}
	}

	// Return first capability as fallback
	if len(agent.Capabilities) > 0 {
		return agent.Capabilities[0].ID
	}

	return "general_analysis"
}

// parseSequentialSteps parses sequential steps from a query
func (dtd *DefaultTaskDecomposer) parseSequentialSteps(query string) []string {
	// Simple regex-based parsing for sequential patterns
	re := regexp.MustCompile(`(?i)(?:first|then|next|after|before|step \d+):\s*([^.]+)`)
	matches := re.FindAllStringSubmatch(query, -1)

	var steps []string
	for _, match := range matches {
		if len(match) > 1 {
			steps = append(steps, strings.TrimSpace(match[1]))
		}
	}

	return steps
}

// parseParallelAspects parses parallel aspects from a query
func (dtd *DefaultTaskDecomposer) parseParallelAspects(query string) []string {
	// Split on common parallel indicators
	indicators := []string{" and ", " also ", " both ", " as well as "}

	for _, indicator := range indicators {
		if strings.Contains(strings.ToLower(query), indicator) {
			parts := strings.Split(strings.ToLower(query), indicator)
			if len(parts) >= 2 {
				var aspects []string
				for _, part := range parts {
					aspects = append(aspects, strings.TrimSpace(part))
				}
				return aspects
			}
		}
	}

	return []string{query}
}

// getDependencies calculates dependencies for sequential tasks
func (dtd *DefaultTaskDecomposer) getDependencies(stepIndex, totalSteps int) []string {
	var dependencies []string

	// Each step depends on the previous step
	if stepIndex > 0 {
		dependencies = append(dependencies, generateSubTaskID("", stepIndex-1))
	}

	return dependencies
}

// calculateTimeout calculates timeout based on priority and profile
func (dtd *DefaultTaskDecomposer) calculateTimeout(priority TaskPriority, profile *profiles.AgentProfile) time.Duration {
	baseTimeout := 30 * time.Second

	// Adjust based on priority
	switch priority {
	case PriorityLow:
		baseTimeout = 60 * time.Second
	case PriorityMedium:
		baseTimeout = 30 * time.Second
	case PriorityHigh:
		baseTimeout = 15 * time.Second
	case PriorityCritical:
		baseTimeout = 10 * time.Second
	}

	// Adjust based on profile optimization hints
	if profile != nil && profile.OptimizationHints.TimeoutMs > 0 {
		profileTimeout := time.Duration(profile.OptimizationHints.TimeoutMs) * time.Millisecond
		if profileTimeout < baseTimeout {
			baseTimeout = profileTimeout
		}
	}

	return baseTimeout
}

// Helper functions

func generateSubTaskID(taskID string, index ...int) string {
	if taskID == "" {
		taskID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}

	if len(index) > 0 {
		return fmt.Sprintf("%s-subtask-%d", taskID, index[0])
	}

	return fmt.Sprintf("%s-subtask-%d", taskID, time.Now().UnixNano())
}
