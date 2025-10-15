package escalation

import (
	"context"
	"testing"
	"time"
)

func TestHandler_Handle(t *testing.T) {
	policy := NewPolicy()
	handler := NewHandler(policy)

	tests := []struct {
		name             string
		request          *Request
		expectedApproval bool
		expectedTarget   string
	}{
		{
			name: "valid escalation",
			request: &Request{
				SourceAgent:     "codebase-locator",
				Reason:          "need pattern analysis",
				SuggestedTarget: "codebase-pattern-finder",
				OriginalTask:    "find similar patterns",
			},
			expectedApproval: true,
			expectedTarget:   "codebase-pattern-finder",
		},
		{
			name: "escalation with auto-selected target",
			request: &Request{
				SourceAgent:  "codebase-locator",
				Reason:       "need to analyze code structure",
				OriginalTask: "analyze component",
			},
			expectedApproval: true,
			expectedTarget:   "codebase-analyzer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.Handle(context.Background(), tt.request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp.Approved != tt.expectedApproval {
				t.Errorf("expected approval %v, got %v", tt.expectedApproval, resp.Approved)
			}

			if tt.expectedApproval && resp.TargetAgent != tt.expectedTarget {
				t.Errorf("expected target %s, got %s", tt.expectedTarget, resp.TargetAgent)
			}
		})
	}
}

func TestHandler_HandleInvalidRequest(t *testing.T) {
	policy := NewPolicy()
	handler := NewHandler(policy)

	tests := []struct {
		name    string
		request *Request
	}{
		{
			name:    "nil request",
			request: nil,
		},
		{
			name: "empty source agent",
			request: &Request{
				SourceAgent: "",
				Reason:      "test",
			},
		},
		{
			name: "empty reason",
			request: &Request{
				SourceAgent: "agent1",
				Reason:      "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.Handle(context.Background(), tt.request)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestHandler_HandleDisallowedEscalation(t *testing.T) {
	policy := NewPolicy()
	handler := NewHandler(policy)

	req := &Request{
		SourceAgent:     "codebase-locator",
		Reason:          "test",
		SuggestedTarget: "unknown-agent", // Not in policy
		OriginalTask:    "test task",
	}

	resp, err := handler.Handle(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp.Approved {
		t.Error("expected escalation to be denied")
	}
}

func TestHandler_HandleEscalationLoop(t *testing.T) {
	policy := NewPolicy()
	handler := NewHandler(policy)

	// Create first escalation: agent1 -> agent2
	req1 := &Request{
		SourceAgent:     "codebase-locator",
		Reason:          "test",
		SuggestedTarget: "codebase-analyzer",
		OriginalTask:    "test task",
	}

	resp1, _ := handler.Handle(context.Background(), req1)
	if !resp1.Approved {
		t.Error("first escalation should be approved")
		return
	}

	// Create second escalation: agent2 -> agent1 (would create loop)
	req2 := &Request{
		SourceAgent:     "codebase-analyzer",
		Reason:          "test",
		SuggestedTarget: "codebase-locator",
		OriginalTask:    "test task",
	}

	resp2, _ := handler.Handle(context.Background(), req2)
	if resp2.Approved {
		t.Error("second escalation should be denied (loop detected)")
	}
}

func TestPolicy_AllowEscalation(t *testing.T) {
	policy := NewPolicy()

	tests := []struct {
		name     string
		source   string
		target   string
		expected bool
	}{
		{
			name:     "allowed escalation",
			source:   "codebase-locator",
			target:   "codebase-analyzer",
			expected: true,
		},
		{
			name:     "disallowed escalation",
			source:   "codebase-locator",
			target:   "unknown-agent",
			expected: false,
		},
		{
			name:     "self escalation",
			source:   "codebase-locator",
			target:   "codebase-locator",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := policy.AllowEscalation(tt.source, tt.target)

			if allowed != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, allowed)
			}
		})
	}
}

func TestPolicy_DetermineTarget(t *testing.T) {
	policy := NewPolicy()

	tests := []struct {
		name     string
		source   string
		reason   string
		expected string
	}{
		{
			name:     "pattern-based target",
			source:   "codebase-locator",
			reason:   "need to find similar patterns",
			expected: "codebase-pattern-finder",
		},
		{
			name:     "analyze-based target",
			source:   "codebase-locator",
			reason:   "need to analyze code structure",
			expected: "codebase-analyzer",
		},
		{
			name:     "default target",
			source:   "codebase-locator",
			reason:   "random reason",
			expected: "codebase-analyzer", // First in list
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := policy.DetermineTarget(tt.source, tt.reason)

			if target != tt.expected {
				t.Errorf("expected target %s, got %s", tt.expected, target)
			}
		})
	}
}

func TestPolicy_GetFallbacks(t *testing.T) {
	policy := NewPolicy()

	fallbacks := policy.GetFallbacks("codebase-locator")

	if len(fallbacks) == 0 {
		t.Error("expected fallbacks for codebase-locator")
	}
}

func TestHistory_RecordAndRetrieve(t *testing.T) {
	history := NewHistory()

	req := &Request{
		SourceAgent:  "agent1",
		Reason:       "test",
		OriginalTask: "test task",
		Timestamp:    time.Now(),
	}

	resp := &Response{
		Approved:    true,
		TargetAgent: "agent2",
		Reason:      "approved",
	}

	history.RecordAttempt(req)
	history.RecordDecision(req, resp)

	if history.Count() != 1 {
		t.Errorf("expected 1 entry, got %d", history.Count())
	}

	entries := history.GetEscalationsForAgent("agent1")
	if len(entries) != 1 {
		t.Errorf("expected 1 entry for agent1, got %d", len(entries))
	}
}

func TestHistory_GetSuccessRate(t *testing.T) {
	history := NewHistory()

	// Add successful escalation
	req1 := &Request{
		SourceAgent:  "agent1",
		Reason:       "test",
		OriginalTask: "task1",
	}
	resp1 := &Response{
		Approved:    true,
		TargetAgent: "agent2",
	}
	history.RecordAttempt(req1)
	history.RecordDecision(req1, resp1)

	// Add failed escalation
	req2 := &Request{
		SourceAgent:  "agent1",
		Reason:       "test",
		OriginalTask: "task2",
	}
	resp2 := &Response{
		Approved: false,
	}
	history.RecordAttempt(req2)
	history.RecordDecision(req2, resp2)

	rate := history.GetSuccessRate("agent1", 1*time.Hour)

	expected := 0.5 // 1 success out of 2
	if rate != expected {
		t.Errorf("expected success rate %f, got %f", expected, rate)
	}
}

func TestHistory_Clear(t *testing.T) {
	history := NewHistory()

	req := &Request{
		SourceAgent:  "agent1",
		Reason:       "test",
		OriginalTask: "task",
	}
	resp := &Response{
		Approved: true,
	}
	history.RecordAttempt(req)
	history.RecordDecision(req, resp)

	if history.Count() != 1 {
		t.Error("expected 1 entry before clear")
	}

	history.Clear()

	if history.Count() != 0 {
		t.Error("expected 0 entries after clear")
	}
}
