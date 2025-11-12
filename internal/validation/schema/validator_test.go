package schema

import (
	"testing"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator(true)
	if v == nil {
		t.Fatal("expected validator, got nil")
	}
	if !v.allowPartial {
		t.Error("expected allowPartial to be true")
	}

	v2 := NewValidator(false)
	if v2.allowPartial {
		t.Error("expected allowPartial to be false")
	}
}

func TestValidate_NilOutput(t *testing.T) {
	v := NewValidator(false)
	_, err := v.Validate(nil)
	if err == nil {
		t.Error("expected error for nil output")
	}
}

func TestValidate_MinimalValid(t *testing.T) {
	v := NewValidator(true) // Allow partial

	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test-component",
	}

	result, err := v.Validate(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("expected valid result")
		for _, e := range result.Errors {
			t.Logf("Error: %s - %s", e.Field, e.Message)
		}
	}

	if len(result.Errors) > 0 {
		t.Errorf("expected no errors, got %d", len(result.Errors))
	}
}

func TestValidateVersion(t *testing.T) {
	v := NewValidator(false)

	tests := []struct {
		name      string
		version   string
		wantError bool
	}{
		{"valid", "AGENT_OUTPUT_V1", false},
		{"empty", "", true},
		{"wrong version", "AGENT_OUTPUT_V2", true},
		{"lowercase", "agent_output_v1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &schema.AgentOutputV1{
				Version:       tt.version,
				ComponentName: "test",
			}

			result := &ValidationResult{
				Errors: make([]ValidationError, 0),
			}

			v.validateVersion(output, result)

			hasError := len(result.Errors) > 0
			if hasError != tt.wantError {
				t.Errorf("validateVersion(%q) hasError = %v, want %v", tt.version, hasError, tt.wantError)
			}
		})
	}
}

func TestValidateComponentName(t *testing.T) {
	v := NewValidator(false)

	tests := []struct {
		name      string
		component string
		wantError bool
	}{
		{"valid", "my-component", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &schema.AgentOutputV1{
				Version:       "AGENT_OUTPUT_V1",
				ComponentName: tt.component,
			}

			result := &ValidationResult{
				Errors: make([]ValidationError, 0),
			}

			v.validateComponentName(output, result)

			hasError := len(result.Errors) > 0
			if hasError != tt.wantError {
				t.Errorf("validateComponentName(%q) hasError = %v, want %v", tt.component, hasError, tt.wantError)
			}
		})
	}
}

func TestValidateRawEvidence(t *testing.T) {
	v := NewValidator(false)

	tests := []struct {
		name         string
		evidence     []schema.Evidence
		wantErrors   int
		wantWarnings int
	}{
		{
			name:         "no evidence",
			evidence:     []schema.Evidence{},
			wantErrors:   0,
			wantWarnings: 1,
		},
		{
			name: "valid evidence",
			evidence: []schema.Evidence{
				{File: "/home/user/file.go", Lines: "10"},
			},
			wantErrors:   0,
			wantWarnings: 0,
		},
		{
			name: "missing file",
			evidence: []schema.Evidence{
				{File: "", Lines: "10"},
			},
			wantErrors:   1,
			wantWarnings: 0,
		},
		{
			name: "missing lines",
			evidence: []schema.Evidence{
				{File: "/home/user/file.go", Lines: ""},
			},
			wantErrors:   1,
			wantWarnings: 0,
		},
		{
			name: "relative path",
			evidence: []schema.Evidence{
				{File: "relative/file.go", Lines: "10"},
			},
			wantErrors:   1,
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &schema.AgentOutputV1{
				Version:       "AGENT_OUTPUT_V1",
				ComponentName: "test",
				RawEvidence:   tt.evidence,
			}

			result := &ValidationResult{
				Errors:   make([]ValidationError, 0),
				Warnings: make([]ValidationWarning, 0),
			}

			v.validateRawEvidence(output, result)

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("got %d errors, want %d", len(result.Errors), tt.wantErrors)
			}

			if len(result.Warnings) != tt.wantWarnings {
				t.Errorf("got %d warnings, want %d", len(result.Warnings), tt.wantWarnings)
			}
		})
	}
}

func TestValidateEntryPoints(t *testing.T) {
	v := NewValidator(false)

	entryPoints := []schema.EntryPoint{
		{File: "/home/user/file.go", Lines: "10", Symbol: "main"},
		{File: "", Lines: "20", Symbol: "helper"},               // Missing file
		{File: "/home/user/file.go", Lines: "", Symbol: "util"}, // Missing lines
		{File: "/home/user/file.go", Lines: "30", Symbol: ""},   // Missing symbol
		{File: "relative.go", Lines: "40", Symbol: "func"},      // Relative path
	}

	result := &ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	v.validateEntryPoints(entryPoints, result)

	expectedErrors := 5 // 4 validation errors + 1 relative path
	if len(result.Errors) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(result.Errors))
		for _, e := range result.Errors {
			t.Logf("Error: %s - %s", e.Field, e.Message)
		}
	}
}

func TestValidateCallGraph(t *testing.T) {
	v := NewValidator(false)

	callGraph := []schema.CallGraphEdge{
		{From: "main", To: "helper", ViaLine: 10},
		{From: "", To: "helper", ViaLine: 20},   // Missing from
		{From: "main", To: "", ViaLine: 30},     // Missing to
		{From: "main", To: "util", ViaLine: 0},  // Invalid line
		{From: "main", To: "func", ViaLine: -5}, // Negative line
	}

	result := &ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	v.validateCallGraph(callGraph, result)

	expectedErrors := 4 // Missing from, missing to, 2 invalid lines
	if len(result.Errors) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(result.Errors))
		for _, e := range result.Errors {
			t.Logf("Error: %s - %s", e.Field, e.Message)
		}
	}
}

func TestValidateDataFlow(t *testing.T) {
	v := NewValidator(false)

	dataFlow := schema.DataFlow{
		Inputs: []schema.DataPoint{
			{Source: "/home/user/file.go:10", Name: "input1"},
			{Source: "", Name: "input2"},                // Missing source
			{Source: "/home/user/file.go:20", Name: ""}, // Missing name
		},
		Transformations: []schema.Transformation{
			{File: "/home/user/file.go", Lines: "30-40", Operation: "transform"},
			{File: "", Lines: "50", Operation: "op"},                 // Missing file
			{File: "/home/user/file.go", Lines: "", Operation: "op"}, // Missing lines
		},
		Outputs: []schema.DataPoint{
			{Source: "/home/user/file.go:60", Name: "output1"},
			{Source: "", Name: "output2"}, // Missing source
		},
	}

	result := &ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	v.validateDataFlow(dataFlow, result)

	if len(result.Errors) == 0 {
		t.Error("expected errors for invalid data flow")
	}
}

func TestValidateSideEffects(t *testing.T) {
	v := NewValidator(false)

	sideEffects := []schema.SideEffect{
		{File: "/home/user/file.go", Line: 10, Type: "io"},
		{File: "", Line: 20, Type: "io"},                  // Missing file
		{File: "/home/user/file.go", Line: 0, Type: "io"}, // Invalid line
		{File: "relative.go", Line: 30, Type: "io"},       // Relative path
	}

	result := &ValidationResult{
		Errors: make([]ValidationError, 0),
	}

	v.validateSideEffects(sideEffects, result)

	expectedErrors := 3 // Missing file, invalid line, relative path
	if len(result.Errors) != expectedErrors {
		t.Errorf("expected %d errors, got %d", expectedErrors, len(result.Errors))
	}
}

func TestIsAbsolutePath(t *testing.T) {
	v := NewValidator(false)

	tests := []struct {
		path string
		want bool
	}{
		{"/home/user/file.go", true},
		{"/absolute/path", true},
		{"relative/path", false},
		{"./relative", false},
		{"../parent", false},
		{"", false},
		{"C:/Windows/path", true},
		{"c:/windows/path", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := v.isAbsolutePath(tt.path)
			if got != tt.want {
				t.Errorf("isAbsolutePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidate_AllowPartial(t *testing.T) {
	output := &schema.AgentOutputV1{
		Version:       "AGENT_OUTPUT_V1",
		ComponentName: "test",
		// Missing many optional fields
	}

	t.Run("allow partial", func(t *testing.T) {
		v := NewValidator(true)
		result, err := v.Validate(output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.Valid {
			t.Error("expected valid when allowPartial=true")
		}
	})

	t.Run("require complete", func(t *testing.T) {
		v := NewValidator(false)
		result, err := v.Validate(output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should still be valid if no errors, just warnings
		if len(result.Errors) > 0 {
			t.Error("unexpected errors")
		}
	})
}

func TestValidate_ComprehensiveOutput(t *testing.T) {
	v := NewValidator(false)

	output := &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "comprehensive-component",
		ScopeDescription: "Complete test output with all sections",
		Overview:         "This is a comprehensive overview with sufficient detail for validation purposes.",
		RawEvidence: []schema.Evidence{
			{File: "/home/user/project/main.go", Lines: "1-10"},
			{File: "/home/user/project/helper.go", Lines: "20"},
		},
		EntryPoints: []schema.EntryPoint{
			{File: "/home/user/project/main.go", Lines: "5", Symbol: "main"},
		},
		CallGraph: []schema.CallGraphEdge{
			{From: "/home/user/project/main.go:main", To: "/home/user/project/helper.go:help", ViaLine: 7},
		},
		DataFlow: schema.DataFlow{
			Inputs: []schema.DataPoint{
				{Source: "/home/user/project/main.go:3", Name: "args"},
			},
			Transformations: []schema.Transformation{
				{File: "/home/user/project/helper.go", Lines: "20-25", Operation: "process"},
			},
			Outputs: []schema.DataPoint{
				{Source: "/home/user/project/main.go:9", Name: "result"},
			},
		},
		StateManagement: []schema.StateOperation{
			{File: "/home/user/project/state.go", Lines: "10", Operation: "read", Entity: "config"},
		},
		SideEffects: []schema.SideEffect{
			{File: "/home/user/project/io.go", Line: 15, Type: "file-write"},
		},
		ErrorHandling: []schema.ErrorHandler{
			{File: "/home/user/project/main.go", Lines: "8-9", Type: "if-err"},
		},
		Configuration: []schema.ConfigInfluence{
			{File: "/home/user/project/config.go", Line: 5, Name: "PORT"},
		},
		Patterns: []schema.Pattern{
			{File: "/home/user/project/factory.go", Lines: "1-20", Name: "Factory"},
		},
		Concurrency: []schema.ConcurrencyMechanism{
			{File: "/home/user/project/worker.go", Lines: "30-40", Mechanism: "goroutine"},
		},
		ExternalDependencies: []schema.ExternalDependency{
			{File: "/home/user/project/import.go", Line: 2, Module: "fmt"},
		},
	}

	result, err := v.Validate(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("expected valid comprehensive output")
		t.Logf("Errors: %d", len(result.Errors))
		for _, e := range result.Errors {
			t.Logf("  - %s: %s", e.Field, e.Message)
		}
		t.Logf("Warnings: %d", len(result.Warnings))
		for _, w := range result.Warnings {
			t.Logf("  - %s: %s", w.Field, w.Message)
		}
	}

	if len(result.Errors) > 0 {
		t.Errorf("expected no errors, got %d", len(result.Errors))
	}
}

func TestGetFieldType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"nil", nil, "nil"},
		{"string", "test", "string"},
		{"int", 42, "int"},
		{"bool", true, "bool"},
		{"slice", []string{}, "[]string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFieldType(tt.value)
			if got != tt.want {
				t.Errorf("GetFieldType() = %v, want %v", got, tt.want)
			}
		})
	}
}
