package schema

import (
	"encoding/json"
	"testing"
)

func TestAgentOutputV1_JSONMarshaling(t *testing.T) {
	output := AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "test-component",
		ScopeDescription: "Test component for validation",
		Overview:         "This is a test component that demonstrates schema marshaling",
		EntryPoints: []EntryPoint{
			{
				File:   "/test/main.go",
				Lines:  "10-20",
				Symbol: "main",
				Role:   "handler",
			},
		},
		RawEvidence: []Evidence{
			{
				Claim: "Main function exists",
				File:  "/test/main.go",
				Lines: "10-20",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded AgentOutputV1
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify key fields
	if decoded.Version != output.Version {
		t.Errorf("version mismatch: got %s, want %s", decoded.Version, output.Version)
	}

	if decoded.ComponentName != output.ComponentName {
		t.Errorf("component name mismatch: got %s, want %s", decoded.ComponentName, output.ComponentName)
	}

	if len(decoded.EntryPoints) != 1 {
		t.Errorf("entry points count mismatch: got %d, want 1", len(decoded.EntryPoints))
	}

	if len(decoded.RawEvidence) != 1 {
		t.Errorf("evidence count mismatch: got %d, want 1", len(decoded.RawEvidence))
	}
}

func TestAgentRequest_JSONMarshaling(t *testing.T) {
	req := AgentRequest{
		RequestID: "test-123",
		AgentID:   "codebase-locator",
		Task: AgentTask{
			TargetAgent:        "codebase-locator",
			Files:              []string{"/test/file.go"},
			AllowedDirectories: []string{"/test"},
			SpecificRequest:    "Find main function",
		},
		Permissions: Permissions{
			AllowedDirectories: []string{"/test"},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   30,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AgentRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.RequestID != req.RequestID {
		t.Errorf("request ID mismatch")
	}

	if decoded.Permissions.ReadOnly != req.Permissions.ReadOnly {
		t.Errorf("read-only permission mismatch")
	}
}

func TestAgentResponse_Escalation(t *testing.T) {
	resp := AgentResponse{
		RequestID: "test-456",
		AgentID:   "codebase-analyzer",
		Status:    StatusEscalationRequired,
		Escalation: &Escalation{
			Required:     true,
			TargetAgent:  "codebase-locator",
			Reason:       "Symbol not found in scope",
			RequiredInfo: "Location of processPayment function",
		},
	}

	if resp.Status != StatusEscalationRequired {
		t.Errorf("expected escalation status")
	}

	if resp.Escalation == nil {
		t.Fatalf("escalation should not be nil")
	}

	if !resp.Escalation.Required {
		t.Errorf("escalation should be required")
	}
}
