package multiagent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
)

// AgentCapability represents what an agent can do
type AgentCapability struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"` // "analysis", "search", "debugging", "documentation", "security", "architecture"
	Skills      []string               `json:"skills"`
	Parameters  map[string]interface{} `json:"parameters"`
	ProfileID   string                 `json:"profile_id"` // Links to agent profile
	Performance *CapabilityPerformance `json:"performance"`
	Available   bool                   `json:"available"`
	LastSeen    time.Time              `json:"last_seen"`
}

// CapabilityPerformance tracks agent performance metrics
type CapabilityPerformance struct {
	SuccessRate    float64   `json:"success_rate"`
	AverageLatency float64   `json:"average_latency_ms"`
	Throughput     float64   `json:"throughput_qps"`
	ErrorRate      float64   `json:"error_rate"`
	LoadCapacity   int       `json:"load_capacity"`
	CurrentLoad    int       `json:"current_load"`
	LastUpdated    time.Time `json:"last_updated"`
}

// AgentRegistry manages agent discovery and capabilities
type AgentRegistry struct {
	agents              map[string]*RegisteredAgent
	capabilities        map[string][]*AgentCapability
	mu                  sync.RWMutex
	healthCheckInterval time.Duration
	lastHealthCheck     time.Time
}

// RegisteredAgent represents an agent in the registry
type RegisteredAgent struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Capabilities []*AgentCapability     `json:"capabilities"`
	Profile      *profiles.AgentProfile `json:"profile"`
	Endpoint     string                 `json:"endpoint"`
	Status       AgentStatus            `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
	RegisteredAt time.Time              `json:"registered_at"`
	LastSeen     time.Time              `json:"last_seen"`
}

// AgentStatus represents the current status of an agent
type AgentStatus string

const (
	AgentStatusAvailable   AgentStatus = "available"
	AgentStatusBusy        AgentStatus = "busy"
	AgentStatusUnavailable AgentStatus = "unavailable"
	AgentStatusError       AgentStatus = "error"
	AgentStatusMaintenance AgentStatus = "maintenance"
)

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents:              make(map[string]*RegisteredAgent),
		capabilities:        make(map[string][]*AgentCapability),
		healthCheckInterval: 30 * time.Second,
		lastHealthCheck:     time.Now(),
	}
}

// RegisterAgent registers a new agent with its capabilities
func (ar *AgentRegistry) RegisterAgent(ctx context.Context, agent *RegisteredAgent) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if agent.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	agent.RegisteredAt = time.Now()
	agent.LastSeen = time.Now()
	agent.Status = AgentStatusAvailable

	// Set default profile if not provided
	if agent.Profile == nil {
		agent.Profile = profiles.GetProfileByID("general")
	}

	ar.agents[agent.ID] = agent

	// Update capabilities index
	ar.updateCapabilitiesIndex(agent)

	return nil
}

// UnregisterAgent removes an agent from the registry
func (ar *AgentRegistry) UnregisterAgent(ctx context.Context, agentID string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	agent, exists := ar.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Remove from capabilities index
	ar.removeFromCapabilitiesIndex(agent)

	delete(ar.agents, agentID)
	return nil
}

// GetAgent retrieves an agent by ID
func (ar *AgentRegistry) GetAgent(agentID string) (*RegisteredAgent, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	agent, exists := ar.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return agent, nil
}

// GetAgentsByCapability returns agents that have a specific capability
func (ar *AgentRegistry) GetAgentsByCapability(capabilityID string) []*RegisteredAgent {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	var matchingAgents []*RegisteredAgent

	for _, agent := range ar.agents {
		if agent.Status != AgentStatusAvailable {
			continue
		}

		for _, capability := range agent.Capabilities {
			if capability.ID == capabilityID {
				matchingAgents = append(matchingAgents, agent)
				break
			}
		}
	}

	return matchingAgents
}

// GetAgentsByCategory returns agents in a specific category
func (ar *AgentRegistry) GetAgentsByCategory(category string) []*RegisteredAgent {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	var matchingAgents []*RegisteredAgent

	for _, agent := range ar.agents {
		if agent.Status != AgentStatusAvailable {
			continue
		}

		for _, capability := range agent.Capabilities {
			if capability.Category == category {
				matchingAgents = append(matchingAgents, agent)
				break
			}
		}
	}

	return matchingAgents
}

// GetAgentsByProfile returns agents using a specific profile
func (ar *AgentRegistry) GetAgentsByProfile(profileID string) []*RegisteredAgent {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	var matchingAgents []*RegisteredAgent

	for _, agent := range ar.agents {
		if agent.Status != AgentStatusAvailable {
			continue
		}

		if agent.Profile.ID == profileID {
			matchingAgents = append(matchingAgents, agent)
		}
	}

	return matchingAgents
}

// GetAllAgents returns all registered agents
func (ar *AgentRegistry) GetAllAgents() []*RegisteredAgent {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	agents := make([]*RegisteredAgent, 0, len(ar.agents))
	for _, agent := range ar.agents {
		agents = append(agents, agent)
	}

	return agents
}

// UpdateAgentStatus updates an agent's status
func (ar *AgentRegistry) UpdateAgentStatus(ctx context.Context, agentID string, status AgentStatus) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	agent, exists := ar.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.Status = status
	agent.LastSeen = time.Now()

	return nil
}

// UpdateAgentPerformance updates an agent's performance metrics
func (ar *AgentRegistry) UpdateAgentPerformance(ctx context.Context, agentID, capabilityID string, metrics *CapabilityPerformance) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	agent, exists := ar.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// Find and update the capability
	for _, capability := range agent.Capabilities {
		if capability.ID == capabilityID {
			capability.Performance = metrics
			capability.LastSeen = time.Now()
			break
		}
	}

	agent.LastSeen = time.Now()
	return nil
}

// GetCapabilityIndex returns agents indexed by capability
func (ar *AgentRegistry) GetCapabilityIndex() map[string][]*AgentCapability {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	index := make(map[string][]*AgentCapability)
	for capabilityID, capabilities := range ar.capabilities {
		// Return copies to prevent external modification
		capabilityCopies := make([]*AgentCapability, len(capabilities))
		copy(capabilityCopies, capabilities)
		index[capabilityID] = capabilityCopies
	}

	return index
}

// GetRegistryStats returns registry statistics
func (ar *AgentRegistry) GetRegistryStats() map[string]interface{} {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	stats := map[string]interface{}{
		"total_agents":       len(ar.agents),
		"available_agents":   0,
		"busy_agents":        0,
		"unavailable_agents": 0,
		"error_agents":       0,
		"categories":         make(map[string]int),
		"capabilities":       make(map[string]int),
		"profiles":           make(map[string]int),
		"last_health_check":  ar.lastHealthCheck,
	}

	// Count agents by status
	for _, agent := range ar.agents {
		switch agent.Status {
		case AgentStatusAvailable:
			stats["available_agents"] = stats["available_agents"].(int) + 1
		case AgentStatusBusy:
			stats["busy_agents"] = stats["busy_agents"].(int) + 1
		case AgentStatusUnavailable:
			stats["unavailable_agents"] = stats["unavailable_agents"].(int) + 1
		case AgentStatusError:
			stats["error_agents"] = stats["error_agents"].(int) + 1
		}

		// Count by category
		for _, capability := range agent.Capabilities {
			category := capability.Category
			if category != "" {
				stats["categories"].(map[string]int)[category]++
			}
		}

		// Count by profile
		profileID := agent.Profile.ID
		stats["profiles"].(map[string]int)[profileID]++
	}

	// Count capabilities
	for capabilityID, agents := range ar.capabilities {
		stats["capabilities"].(map[string]int)[capabilityID] = len(agents)
	}

	return stats
}

// HealthCheck performs health checks on all agents
func (ar *AgentRegistry) HealthCheck(ctx context.Context) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	ar.lastHealthCheck = time.Now()

	// In a real implementation, this would ping each agent
	// For now, we'll mark agents as available if they were seen recently
	cutoff := time.Now().Add(-2 * ar.healthCheckInterval)

	for _, agent := range ar.agents {
		if agent.LastSeen.Before(cutoff) {
			agent.Status = AgentStatusUnavailable
		} else {
			agent.Status = AgentStatusAvailable
		}
	}

	return nil
}

// FindBestAgent finds the best agent for a given capability based on performance
func (ar *AgentRegistry) FindBestAgent(ctx context.Context, capabilityID string, requirements map[string]interface{}) (*RegisteredAgent, error) {
	agents := ar.GetAgentsByCapability(capabilityID)
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents available for capability %s", capabilityID)
	}

	// Score agents based on performance and requirements
	var bestAgent *RegisteredAgent
	bestScore := 0.0

	for _, agent := range agents {
		score := ar.scoreAgent(agent, capabilityID, requirements)
		if score > bestScore {
			bestScore = score
			bestAgent = agent
		}
	}

	if bestAgent == nil {
		return nil, fmt.Errorf("no suitable agent found for capability %s", capabilityID)
	}

	return bestAgent, nil
}

// scoreAgent calculates a score for an agent based on performance and requirements
func (ar *AgentRegistry) scoreAgent(agent *RegisteredAgent, capabilityID string, requirements map[string]interface{}) float64 {
	score := 0.0

	// Find the capability
	var capability *AgentCapability
	for _, cap := range agent.Capabilities {
		if cap.ID == capabilityID {
			capability = cap
			break
		}
	}

	if capability == nil || capability.Performance == nil {
		return 0.0
	}

	perf := capability.Performance

	// Base score from success rate (0-100)
	score += perf.SuccessRate

	// Bonus for low latency (prefer faster agents)
	if perf.AverageLatency > 0 {
		latencyScore := 100.0 / (1.0 + perf.AverageLatency/1000.0) // Normalize to 0-100
		score += latencyScore * 0.2
	}

	// Bonus for high throughput
	if perf.Throughput > 0 {
		throughputScore := min(perf.Throughput/10.0, 20.0) // Cap at 20 points
		score += throughputScore
	}

	// Penalty for high error rate
	score -= perf.ErrorRate * 50.0

	// Penalty for high load
	if perf.LoadCapacity > 0 {
		loadRatio := float64(perf.CurrentLoad) / float64(perf.LoadCapacity)
		if loadRatio > 0.8 {
			score -= (loadRatio - 0.8) * 100.0 // Penalty for high load
		}
	}

	// Check requirements match
	for key, value := range requirements {
		if agent.Metadata[key] != value {
			score -= 10.0 // Penalty for not matching requirements
		}
	}

	return max(score, 0.0) // Ensure non-negative score
}

// updateCapabilitiesIndex updates the capabilities index when an agent is registered
func (ar *AgentRegistry) updateCapabilitiesIndex(agent *RegisteredAgent) {
	for _, capability := range agent.Capabilities {
		ar.capabilities[capability.ID] = append(ar.capabilities[capability.ID], capability)
	}
}

// removeFromCapabilitiesIndex removes an agent from the capabilities index
func (ar *AgentRegistry) removeFromCapabilitiesIndex(agent *RegisteredAgent) {
	for capabilityID, capabilities := range ar.capabilities {
		var filtered []*AgentCapability
		for _, cap := range capabilities {
			// Remove capabilities belonging to this agent
			if cap.ID != capabilityID { // This logic needs to be fixed
				filtered = append(filtered, cap)
			}
		}
		ar.capabilities[capabilityID] = filtered
	}
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
