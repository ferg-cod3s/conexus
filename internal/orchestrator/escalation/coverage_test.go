package escalation

import (
	"testing"
	"time"
)

// TestHandler_GetHistory tests the GetHistory method
func TestHandler_GetHistory(t *testing.T) {
	policy := NewPolicy()
	handler := NewHandler(policy)

	history := handler.GetHistory()
	if history == nil {
		t.Error("expected history to be non-nil")
	}

	// Verify it's the same instance
	history2 := handler.GetHistory()
	if history != history2 {
		t.Error("expected to get the same history instance")
	}
}

// TestHandler_GetPolicy tests the GetPolicy method
func TestHandler_GetPolicy(t *testing.T) {
	policy := NewPolicy()
	handler := NewHandler(policy)

	retrievedPolicy := handler.GetPolicy()
	if retrievedPolicy == nil {
		t.Error("expected policy to be non-nil")
	}

	// Verify it's the same instance
	retrievedPolicy2 := handler.GetPolicy()
	if retrievedPolicy != retrievedPolicy2 {
		t.Error("expected to get the same policy instance")
	}
}

// TestHandler_SetPolicy tests the SetPolicy method
func TestHandler_SetPolicy(t *testing.T) {
	policy := NewPolicy()
	handler := NewHandler(policy)

	// Create a new policy
	newPolicy := NewPolicy()
	newPolicy.SetMaxDepth(5)

	// Set the new policy
	handler.SetPolicy(newPolicy)

	// Verify the policy was updated
	retrievedPolicy := handler.GetPolicy()
	if retrievedPolicy.GetMaxDepth() != 5 {
		t.Errorf("expected max depth 5, got %d", retrievedPolicy.GetMaxDepth())
	}
}

// TestHistory_GetRecentEscalations tests the GetRecentEscalations method
func TestHistory_GetRecentEscalations(t *testing.T) {
	history := NewHistory()

	// Add an old entry first
	oldReq := &Request{
		SourceAgent:  "agent1",
		Reason:       "old test",
		OriginalTask: "old task",
	}
	oldResp := &Response{
		Approved:    true,
		TargetAgent: "agent2",
	}
	history.RecordDecision(oldReq, oldResp)

	// Wait a bit to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	// Add a recent entry
	recentReq := &Request{
		SourceAgent:  "agent1",
		Reason:       "recent test",
		OriginalTask: "recent task",
	}
	recentResp := &Response{
		Approved:    true,
		TargetAgent: "agent3",
	}
	history.RecordDecision(recentReq, recentResp)

	// Get recent escalations within a very short window (should only get the latest)
	recent := history.GetRecentEscalations(1 * time.Millisecond)
	if len(recent) != 1 {
		t.Errorf("expected 1 recent escalation, got %d", len(recent))
	}

	if len(recent) > 0 && recent[0].Request.Reason != "recent test" {
		t.Errorf("expected recent test, got %s", recent[0].Request.Reason)
	}

	// Get all escalations within a longer window
	all := history.GetRecentEscalations(1 * time.Second)
	if len(all) != 2 {
		t.Errorf("expected 2 escalations, got %d", len(all))
	}
}

// TestPolicy_AddPath tests the AddPath method
func TestPolicy_AddPath(t *testing.T) {
	policy := NewPolicy()

	// Add a new path
	newTargets := []string{"new-agent-1", "new-agent-2"}
	policy.AddPath("test-source", newTargets)

	// Verify the path was added
	if !policy.AllowEscalation("test-source", "new-agent-1") {
		t.Error("expected escalation to be allowed for new path")
	}

	if !policy.AllowEscalation("test-source", "new-agent-2") {
		t.Error("expected escalation to be allowed for new path")
	}
}

// TestPolicy_AddFallback tests the AddFallback method
func TestPolicy_AddFallback(t *testing.T) {
	policy := NewPolicy()

	// Add new fallbacks
	newFallbacks := []string{"fallback-1", "fallback-2"}
	policy.AddFallback("test-agent", newFallbacks)

	// Verify the fallbacks were added
	fallbacks := policy.GetFallbacks("test-agent")
	if len(fallbacks) != 2 {
		t.Errorf("expected 2 fallbacks, got %d", len(fallbacks))
	}

	if fallbacks[0] != "fallback-1" || fallbacks[1] != "fallback-2" {
		t.Errorf("expected fallback-1, fallback-2, got %v", fallbacks)
	}
}

// TestPolicy_SetMaxDepth tests the SetMaxDepth method
func TestPolicy_SetMaxDepth(t *testing.T) {
	policy := NewPolicy()

	// Test setting valid depth
	policy.SetMaxDepth(10)
	if policy.GetMaxDepth() != 10 {
		t.Errorf("expected max depth 10, got %d", policy.GetMaxDepth())
	}

	// Test setting invalid depth (should not change)
	originalDepth := policy.GetMaxDepth()
	policy.SetMaxDepth(-1)
	if policy.GetMaxDepth() != originalDepth {
		t.Error("expected max depth to remain unchanged when setting invalid value")
	}

	// Test setting zero (should not change)
	policy.SetMaxDepth(0)
	if policy.GetMaxDepth() != originalDepth {
		t.Error("expected max depth to remain unchanged when setting zero")
	}
}

// TestPolicy_GetMaxDepth tests the GetMaxDepth method
func TestPolicy_GetMaxDepth(t *testing.T) {
	policy := NewPolicy()

	// Test default depth
	depth := policy.GetMaxDepth()
	if depth != 3 {
		t.Errorf("expected default max depth 3, got %d", depth)
	}

	// Test after setting custom depth
	policy.SetMaxDepth(7)
	depth = policy.GetMaxDepth()
	if depth != 7 {
		t.Errorf("expected max depth 7, got %d", depth)
	}
}
