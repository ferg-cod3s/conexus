package escalation

import (
	"sync"
	"time"
)

// History tracks escalation attempts and decisions
type History struct {
	mu      sync.RWMutex
	entries []HistoryEntry
}

// HistoryEntry represents a single escalation event
type HistoryEntry struct {
	// Request that triggered the escalation
	Request *Request

	// Response from the handler
	Response *Response

	// Timestamp of the decision
	Timestamp time.Time
}

// NewHistory creates a new escalation history tracker
func NewHistory() *History {
	return &History{
		entries: make([]HistoryEntry, 0),
	}
}

// RecordAttempt records an escalation attempt
func (h *History) RecordAttempt(req *Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	req.Timestamp = time.Now()
}

// RecordDecision records an escalation decision
func (h *History) RecordDecision(req *Request, resp *Response) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = append(h.entries, HistoryEntry{
		Request:   req,
		Response:  resp,
		Timestamp: time.Now(),
	})
}

// HasEscalationLoop checks if escalating would create a loop
func (h *History) HasEscalationLoop(source, target string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Check recent history for patterns
	recentWindow := time.Now().Add(-5 * time.Minute)

	var chain []string
	chain = append(chain, source)

	// Build the escalation chain from recent history
	for i := len(h.entries) - 1; i >= 0; i-- {
		entry := h.entries[i]

		if entry.Timestamp.Before(recentWindow) {
			break
		}

		if entry.Response.Approved {
			// Check if this creates a loop
			if entry.Request.SourceAgent == target && entry.Response.TargetAgent == source {
				return true
			}

			// Build chain
			if entry.Request.SourceAgent == chain[len(chain)-1] {
				chain = append(chain, entry.Response.TargetAgent)
			}
		}
	}

	// Check if target is already in the chain
	for _, agent := range chain {
		if agent == target {
			return true
		}
	}

	return false
}

// GetRecentEscalations returns escalations within the specified time window
func (h *History) GetRecentEscalations(window time.Duration) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	cutoff := time.Now().Add(-window)
	recent := make([]HistoryEntry, 0)

	for _, entry := range h.entries {
		if entry.Timestamp.After(cutoff) {
			recent = append(recent, entry)
		}
	}

	return recent
}

// GetEscalationsForAgent returns all escalations involving the specified agent
func (h *History) GetEscalationsForAgent(agent string) []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make([]HistoryEntry, 0)

	for _, entry := range h.entries {
		if entry.Request.SourceAgent == agent ||
			(entry.Response.Approved && entry.Response.TargetAgent == agent) {
			results = append(results, entry)
		}
	}

	return results
}

// GetSuccessRate calculates the success rate for escalations
func (h *History) GetSuccessRate(agent string, window time.Duration) float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	cutoff := time.Now().Add(-window)

	total := 0
	approved := 0

	for _, entry := range h.entries {
		if entry.Timestamp.Before(cutoff) {
			continue
		}

		if entry.Request.SourceAgent == agent {
			total++
			if entry.Response.Approved {
				approved++
			}
		}
	}

	if total == 0 {
		return 0.0
	}

	return float64(approved) / float64(total)
}

// Clear clears the escalation history
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = make([]HistoryEntry, 0)
}

// Count returns the total number of escalations
func (h *History) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.entries)
}
