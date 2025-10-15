// Package escalation provides the escalation protocol for multi-agent coordination.
//
// The escalation protocol enables:
// - Agents to delegate tasks to more specialized agents
// - Automatic fallback strategies
// - Escalation history tracking
// - Policy-based escalation decisions
package escalation

import (
	"context"
	"fmt"
	"time"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Handler manages escalation requests from agents
type Handler struct {
	policy  *Policy
	history *History
}

// NewHandler creates a new escalation handler
func NewHandler(policy *Policy) *Handler {
	return &Handler{
		policy:  policy,
		history: NewHistory(),
	}
}

// Request represents an escalation request from an agent
type Request struct {
	// Source agent making the escalation
	SourceAgent string

	// Reason for escalation
	Reason string

	// Suggested target agent (optional)
	SuggestedTarget string

	// Original task that couldn't be completed
	OriginalTask string

	// Permissions for the escalated task
	Permissions schema.Permissions

	// Context from the source agent
	Context map[string]interface{}

	// Timestamp of the escalation request
	Timestamp time.Time
}

// Response represents the handler's response to an escalation
type Response struct {
	// Whether escalation was approved
	Approved bool

	// Target agent to handle the escalated task
	TargetAgent string

	// Modified task description (if needed)
	Task string

	// Reason for approval or rejection
	Reason string

	// Fallback agents if target agent fails
	Fallbacks []string
}

// Handle processes an escalation request
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("escalation request is nil")
	}

	if req.SourceAgent == "" {
		return nil, fmt.Errorf("source agent is required")
	}

	if req.Reason == "" {
		return nil, fmt.Errorf("escalation reason is required")
	}

	// Record the escalation attempt
	h.history.RecordAttempt(req)

	// Determine target agent
	targetAgent := req.SuggestedTarget
	if targetAgent == "" {
		// Use policy to determine target
		targetAgent = h.policy.DetermineTarget(req.SourceAgent, req.Reason)
	}

	// Check if escalation is allowed
	if !h.policy.AllowEscalation(req.SourceAgent, targetAgent) {
		resp := &Response{
			Approved:    false,
			TargetAgent: "",
			Reason:      "escalation not allowed by policy",
		}
		h.history.RecordDecision(req, resp)
		return resp, nil
	}

	// Check for escalation loops
	if h.history.HasEscalationLoop(req.SourceAgent, targetAgent) {
		resp := &Response{
			Approved:    false,
			TargetAgent: "",
			Reason:      "escalation loop detected",
		}
		h.history.RecordDecision(req, resp)
		return resp, nil
	}

	// Approve escalation
	resp := &Response{
		Approved:    true,
		TargetAgent: targetAgent,
		Task:        req.OriginalTask,
		Reason:      "escalation approved",
		Fallbacks:   h.policy.GetFallbacks(targetAgent),
	}

	h.history.RecordDecision(req, resp)
	return resp, nil
}

// GetHistory returns the escalation history
func (h *Handler) GetHistory() *History {
	return h.history
}

// GetPolicy returns the escalation policy
func (h *Handler) GetPolicy() *Policy {
	return h.policy
}

// SetPolicy updates the escalation policy
func (h *Handler) SetPolicy(policy *Policy) {
	h.policy = policy
}
