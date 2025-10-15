package escalation

import (
	"strings"
)

// Policy defines escalation rules and allowed transitions
type Policy struct {
	// Allowed escalation paths (source -> targets)
	paths map[string][]string

	// Fallback agents for each agent type
	fallbacks map[string][]string

	// Maximum escalation depth
	maxDepth int
}

// NewPolicy creates a new escalation policy with default rules
func NewPolicy() *Policy {
	return &Policy{
		paths:     DefaultEscalationPaths(),
		fallbacks: DefaultFallbacks(),
		maxDepth:  3,
	}
}

// AllowEscalation checks if escalation from source to target is allowed
func (p *Policy) AllowEscalation(source, target string) bool {
	if source == target {
		return false // No self-escalation
	}

	targets, ok := p.paths[source]
	if !ok {
		return false
	}

	for _, allowed := range targets {
		if allowed == target {
			return true
		}
	}

	return false
}

// DetermineTarget determines the best target agent for an escalation
func (p *Policy) DetermineTarget(source, reason string) string {
	targets, ok := p.paths[source]
	if !ok || len(targets) == 0 {
		return ""
	}

	// Use reason-based heuristics to select best target
	reasonLower := strings.ToLower(reason)

	// Pattern matching for common escalation scenarios
	if strings.Contains(reasonLower, "pattern") || strings.Contains(reasonLower, "similar") {
		if p.isTargetAvailable(targets, "codebase-pattern-finder") {
			return "codebase-pattern-finder"
		}
	}

	if strings.Contains(reasonLower, "analyze") || strings.Contains(reasonLower, "understand") {
		if p.isTargetAvailable(targets, "codebase-analyzer") {
			return "codebase-analyzer"
		}
	}

	if strings.Contains(reasonLower, "find") || strings.Contains(reasonLower, "locate") {
		if p.isTargetAvailable(targets, "codebase-locator") {
			return "codebase-locator"
		}
	}

	// Default to first available target
	return targets[0]
}

// isTargetAvailable checks if a target is in the available list
func (p *Policy) isTargetAvailable(targets []string, target string) bool {
	for _, t := range targets {
		if t == target {
			return true
		}
	}
	return false
}

// GetFallbacks returns fallback agents for a given agent
func (p *Policy) GetFallbacks(agent string) []string {
	fallbacks, ok := p.fallbacks[agent]
	if !ok {
		return []string{}
	}
	return fallbacks
}

// AddPath adds an escalation path
func (p *Policy) AddPath(source string, targets []string) {
	p.paths[source] = targets
}

// AddFallback adds a fallback agent
func (p *Policy) AddFallback(agent string, fallbacks []string) {
	p.fallbacks[agent] = fallbacks
}

// SetMaxDepth sets the maximum escalation depth
func (p *Policy) SetMaxDepth(depth int) {
	if depth > 0 {
		p.maxDepth = depth
	}
}

// GetMaxDepth returns the maximum escalation depth
func (p *Policy) GetMaxDepth() int {
	return p.maxDepth
}

// DefaultEscalationPaths returns the default escalation paths
func DefaultEscalationPaths() map[string][]string {
	return map[string][]string{
		// Locator can escalate to analyzer or pattern-finder
		"codebase-locator": {
			"codebase-analyzer",
			"codebase-pattern-finder",
		},

		// Analyzer can escalate to pattern-finder or locator
		"codebase-analyzer": {
			"codebase-pattern-finder",
			"codebase-locator",
		},

		// Pattern-finder can escalate to analyzer
		"codebase-pattern-finder": {
			"codebase-analyzer",
		},

		// Orchestrator can escalate to any specialized agent
		"orchestrator": {
			"codebase-locator",
			"codebase-analyzer",
			"codebase-pattern-finder",
		},
	}
}

// DefaultFallbacks returns the default fallback agents
func DefaultFallbacks() map[string][]string {
	return map[string][]string{
		"codebase-locator": {
			"codebase-analyzer", // If locator fails, try analyzer
		},
		"codebase-analyzer": {
			"codebase-locator", // If analyzer fails, try locator
		},
		"codebase-pattern-finder": {
			"codebase-analyzer", // If pattern-finder fails, try analyzer
		},
	}
}
