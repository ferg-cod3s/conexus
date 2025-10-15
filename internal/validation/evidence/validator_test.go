package evidence

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator(true)
	if v == nil {
		t.Fatal("expected validator, got nil")
	}
	if !v.strictMode {
		t.Error("expected strictMode to be true")
	}

	v2 := NewValidator(false)
	if v2.strictMode {
		t.Error("expected strictMode to be false")
	}
}

func TestValidate_NilOutput(t *testing.T) {
	v := NewValidator(false)
	_, err := v.Validate(nil)
	if err == nil {
		t.Error("expected error for nil output")
	}
}

func TestValidate_FullCoverage(t *testing.T) {
	v := NewValidator(false)

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test-component",
		RawEvidence: []schema.Evidence{
			{File: "/test/file.go", Lines: "10"},
			{File: "/test/file.go", Lines: "20-30"},
		},
		EntryPoints: []schema.EntryPoint{
			{File: "/test/file.go", Lines: "10", Symbol: "main"},
		},
		CallGraph: []schema.CallGraphEdge{
			{From: "/test/file.go:main", To: "/test/file.go:helper", ViaLine: 20},
		},
	}

	result, err := v.Validate(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("expected valid result")
	}

	if result.CoveragePercentage != 100.0 {
		t.Errorf("expected 100%% coverage, got %.2f%%", result.CoveragePercentage)
	}

	if len(result.UnbackedClaims) > 0 {
		t.Errorf("expected no unbacked claims, got %d", len(result.UnbackedClaims))
	}
}

func TestValidate_PartialCoverage(t *testing.T) {
	v := NewValidator(false)

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test-component",
		RawEvidence: []schema.Evidence{
			{File: "/test/file.go", Lines: "10"},
		},
		EntryPoints: []schema.EntryPoint{
			{File: "/test/file.go", Lines: "10", Symbol: "main"},
			{File: "/test/file.go", Lines: "20", Symbol: "helper"}, // No evidence
		},
	}

	result, err := v.Validate(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid result")
	}

	if result.CoveragePercentage != 50.0 {
		t.Errorf("expected 50%% coverage, got %.2f%%", result.CoveragePercentage)
	}

	if len(result.UnbackedClaims) != 1 {
		t.Errorf("expected 1 unbacked claim, got %d", len(result.UnbackedClaims))
	}
}

func TestValidate_NoClaims(t *testing.T) {
	v := NewValidator(false)

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test-component",
	}

	result, err := v.Validate(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("expected valid result for no claims")
	}

	if result.CoveragePercentage != 100.0 {
		t.Errorf("expected 100%% coverage for no claims, got %.2f%%", result.CoveragePercentage)
	}
}

func TestValidateEntryPoints(t *testing.T) {
	v := NewValidator(false)

	evidenceIndex := map[string]bool{
		"/test/file.go:10":    true,
		"/test/file.go:20-30": true,
	}

	entryPoints := []schema.EntryPoint{
		{File: "/test/file.go", Lines: "10", Symbol: "main"},      // Backed
		{File: "/test/file.go", Lines: "40", Symbol: "unbacked"},  // Not backed
	}

	result := &ValidationResult{
		UnbackedClaims: make([]UnbackedClaim, 0),
	}

	v.validateEntryPoints(entryPoints, evidenceIndex, result)

	if result.TotalClaims != 2 {
		t.Errorf("expected 2 total claims, got %d", result.TotalClaims)
	}

	if result.BackedClaims != 1 {
		t.Errorf("expected 1 backed claim, got %d", result.BackedClaims)
	}

	if len(result.UnbackedClaims) != 1 {
		t.Errorf("expected 1 unbacked claim, got %d", len(result.UnbackedClaims))
	}
}

func TestValidateCallGraph(t *testing.T) {
	v := NewValidator(false)

	evidenceIndex := map[string]bool{
		"/test/file.go:20": true,
	}

	callGraph := []schema.CallGraphEdge{
		{From: "/test/file.go:main", To: "/test/file.go:helper", ViaLine: 20},
		{From: "/test/file.go:helper", To: "/test/file.go:util", ViaLine: 30}, // Not backed
	}

	result := &ValidationResult{
		UnbackedClaims: make([]UnbackedClaim, 0),
	}

	v.validateCallGraph(callGraph, evidenceIndex, result)

	if result.BackedClaims != 1 {
		t.Errorf("expected 1 backed claim, got %d", result.BackedClaims)
	}

	if len(result.UnbackedClaims) != 1 {
		t.Errorf("expected 1 unbacked claim, got %d", len(result.UnbackedClaims))
	}
}

func TestValidateDataFlow(t *testing.T) {
	v := NewValidator(false)

	evidenceIndex := map[string]bool{
		"/test/file.go:10":    true,
		"/test/file.go:20-30": true,
	}

	dataFlow := schema.DataFlow{
		Inputs: []schema.DataPoint{
			{Source: "/test/file.go:10", Name: "input1"},
			{Source: "/test/file.go:99", Name: "unbacked"}, // Not backed
		},
		Transformations: []schema.Transformation{
			{File: "/test/file.go", Lines: "20-30", Operation: "transform"},
		},
		Outputs: []schema.DataPoint{
			{Source: "/test/file.go:10", Name: "output1"},
		},
	}

	result := &ValidationResult{
		UnbackedClaims: make([]UnbackedClaim, 0),
	}

	v.validateDataFlow(dataFlow, evidenceIndex, result)

	if result.TotalClaims != 4 {
		t.Errorf("expected 4 total claims, got %d", result.TotalClaims)
	}

	if result.BackedClaims != 3 {
		t.Errorf("expected 3 backed claims, got %d", result.BackedClaims)
	}

	if len(result.UnbackedClaims) != 1 {
		t.Errorf("expected 1 unbacked claim, got %d", len(result.UnbackedClaims))
	}
}

func TestValidateEvidenceFiles_StrictMode(t *testing.T) {
	// Create temp file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator(true) // Strict mode

	evidence := []schema.Evidence{
		{File: tmpFile, Lines: "1"},                           // Valid
		{File: "/nonexistent/file.go", Lines: "1"},            // File doesn't exist
		{File: tmpFile, Lines: "invalid"},                     // Invalid line format
		{File: tmpFile, Lines: "10-5"},                        // Invalid range
	}

	result := &ValidationResult{
		InvalidEvidence: make([]InvalidEvidence, 0),
	}

	v.validateEvidenceFiles(evidence, result)

	if len(result.InvalidEvidence) != 3 {
		t.Errorf("expected 3 invalid evidence entries, got %d", len(result.InvalidEvidence))
		for _, inv := range result.InvalidEvidence {
			t.Logf("Invalid: %s:%s - %s", inv.File, inv.Lines, inv.Reason)
		}
	}
}

func TestValidateLineRange(t *testing.T) {
	v := NewValidator(false)

	tests := []struct {
		name    string
		lines   string
		wantErr bool
	}{
		{"valid single", "10", false},
		{"valid range", "10-20", false},
		{"empty", "", true},
		{"invalid format", "10-20-30", true},
		{"invalid start", "abc-20", true},
		{"invalid end", "10-xyz", true},
		{"negative", "-5", true},
		{"zero", "0", true},
		{"reverse range", "20-10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.validateLineRange(tt.lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLineRange(%q) error = %v, wantErr %v", tt.lines, err, tt.wantErr)
			}
		})
	}
}

func TestBuildEvidenceIndex(t *testing.T) {
	v := NewValidator(false)

	evidence := []schema.Evidence{
		{File: "/test/file.go", Lines: "10"},
		{File: "/test/file.go", Lines: "20-30"},
		{File: "/test/other.go", Lines: "5"},
	}

	index := v.buildEvidenceIndex(evidence)

	expected := map[string]bool{
		"/test/file.go:10":    true,
		"/test/file.go:20-30": true,
		"/test/other.go:5":    true,
	}

	for key, val := range expected {
		if index[key] != val {
			t.Errorf("expected index[%q] = %v, got %v", key, val, index[key])
		}
	}

	if len(index) != len(expected) {
		t.Errorf("expected index length %d, got %d", len(expected), len(index))
	}
}

func TestValidate_AllSections(t *testing.T) {
	v := NewValidator(false)

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "comprehensive-test",
		RawEvidence: []schema.Evidence{
			{File: "/test/file.go", Lines: "1"},
			{File: "/test/file.go", Lines: "10"},
			{File: "/test/file.go", Lines: "20"},
			{File: "/test/file.go", Lines: "30"},
			{File: "/test/file.go", Lines: "40"},
			{File: "/test/file.go", Lines: "50-60"},
		},
		EntryPoints: []schema.EntryPoint{
			{File: "/test/file.go", Lines: "1", Symbol: "main"},
		},
		CallGraph: []schema.CallGraphEdge{
			{From: "/test/file.go:main", To: "/test/file.go:helper", ViaLine: 10},
		},
		DataFlow: schema.DataFlow{
			Inputs: []schema.DataPoint{
				{Source: "/test/file.go:20", Name: "input"},
			},
		},
		StateManagement: []schema.StateOperation{
			{File: "/test/file.go", Lines: "30", Operation: "read"},
		},
		SideEffects: []schema.SideEffect{
			{File: "/test/file.go", Line: 40, Type: "io"},
		},
		ErrorHandling: []schema.ErrorHandler{
			{File: "/test/file.go", Lines: "50-60", Type: "try-catch"},
		},
	}

	result, err := v.Validate(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("expected valid result")
		t.Logf("Unbacked claims: %d", len(result.UnbackedClaims))
		for _, claim := range result.UnbackedClaims {
			t.Logf("  - %s: %s", claim.Section, claim.Description)
		}
	}

	if result.CoveragePercentage != 100.0 {
		t.Errorf("expected 100%% coverage, got %.2f%%", result.CoveragePercentage)
	}
}
