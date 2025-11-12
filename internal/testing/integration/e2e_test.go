// Package integration_test provides end-to-end integration tests for multi-agent workflows
package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ferg-cod3s/conexus/internal/orchestrator/workflow"
	"github.com/ferg-cod3s/conexus/internal/testing/integration"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// MockAnalyzerAgent simulates the codebase-analyzer agent for testing
// Now enhanced to handle multiple fixture types
type MockAnalyzerAgent struct{}

// Execute implements workflow.Agent interface
func (m *MockAnalyzerAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	// Extract file path from the specific request (which is a JSON string)
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(req.Task.SpecificRequest), &input); err != nil {
		return schema.AgentResponse{
			Status: schema.StatusError,
			Error: &schema.AgentError{
				Code:    "INVALID_INPUT",
				Message: "Failed to parse input: " + err.Error(),
			},
		}, err
	}

	filePath, _ := input["file_path"].(string)

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return schema.AgentResponse{
			Status: schema.StatusError,
			Error: &schema.AgentError{
				Code:    "FILE_READ_ERROR",
				Message: "Failed to read file: " + err.Error(),
			},
		}, err
	}

	// Determine which fixture this is and generate appropriate output
	fileName := filepath.Base(filePath)
	var output *schema.AgentOutputV1

	switch fileName {
	case "simple_function.go":
		output = generateSimpleFunctionOutput(filePath)
	case "multiple_functions.go":
		output = generateMultipleFunctionsOutput(filePath)
	case "struct_methods.go":
		output = generateStructMethodsOutput(filePath)
	case "error_handling.go":
		output = generateErrorHandlingOutput(filePath)
	case "side_effects.go":
		output = generateSideEffectsOutput(filePath)
	default:
		return schema.AgentResponse{
			Status: schema.StatusError,
			Error: &schema.AgentError{
				Code:    "UNKNOWN_FIXTURE",
				Message: fmt.Sprintf("Unknown fixture: %s", fileName),
			},
		}, fmt.Errorf("unknown fixture: %s", fileName)
	}

	// Ensure content is used (for realistic behavior)
	_ = content

	return schema.AgentResponse{
		RequestID: req.RequestID,
		AgentID:   req.AgentID,
		Status:    schema.StatusComplete,
		Output:    output,
		Duration:  100 * time.Millisecond,
		Timestamp: time.Now(),
	}, nil
}

// generateSimpleFunctionOutput creates output for simple_function.go
func generateSimpleFunctionOutput(filePath string) *schema.AgentOutputV1 {
	return &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "simple_function",
		ScopeDescription: "Analysis of simple mathematical and greeting functions",
		Overview:         "This module contains basic utility functions for arithmetic operations and greetings. The Add and Multiply functions provide pure mathematical operations, while Greet performs formatted output to stdout.",

		EntryPoints: []schema.EntryPoint{
			{File: filePath, Lines: "6-8", Symbol: "Add", Role: "utility"},
			{File: filePath, Lines: "11-13", Symbol: "Greet", Role: "utility"},
			{File: filePath, Lines: "16-19", Symbol: "Multiply", Role: "utility"},
		},

		CallGraph: []schema.CallGraphEdge{
			{From: filePath + ":Add", To: "builtin:+", ViaLine: 7},
			{From: filePath + ":Greet", To: "fmt.Printf", ViaLine: 12},
			{From: filePath + ":Multiply", To: "builtin:*", ViaLine: 17},
		},

		DataFlow: schema.DataFlow{
			Inputs: []schema.DataPoint{
				{Source: filePath + ":6", Name: "a, b", Type: "int", Description: "Integer operands for addition"},
				{Source: filePath + ":11", Name: "name", Type: "string", Description: "Name for greeting"},
				{Source: filePath + ":16", Name: "x, y", Type: "int", Description: "Integer operands for multiplication"},
			},
			Transformations: []schema.Transformation{
				{File: filePath, Lines: "7", Operation: "arithmetic", Description: "Integer addition"},
				{File: filePath, Lines: "17", Operation: "arithmetic", Description: "Integer multiplication"},
			},
			Outputs: []schema.DataPoint{
				{Source: filePath + ":7", Name: "result", Type: "int", Description: "Sum of a and b"},
				{Source: filePath + ":18", Name: "result", Type: "int", Description: "Product of x and y"},
			},
		},

		StateManagement: []schema.StateOperation{},

		SideEffects: []schema.SideEffect{
			{File: filePath, Line: 12, Type: "log", Description: "Greet function writes to stdout via fmt.Printf"},
		},

		ErrorHandling: []schema.ErrorHandler{},
		Configuration: []schema.ConfigInfluence{},

		Patterns: []schema.Pattern{
			{Name: "Pure Function: Add", File: filePath, Lines: "6-8", Description: "Mathematical function with no side effects"},
			{Name: "Pure Function: Multiply", File: filePath, Lines: "16-19", Description: "Mathematical functions with no side effects"},
		},

		Concurrency:          []schema.ConcurrencyMechanism{},
		ExternalDependencies: []schema.ExternalDependency{{File: filePath, Line: 3, Module: "fmt", Purpose: "Formatted I/O operations"}},
		Limitations:          []string{"No error handling for edge cases"},
		OpenQuestions:        []string{},

		RawEvidence: []schema.Evidence{
			{Claim: "Complete source code of simple_function.go", File: filePath, Lines: "1-20"},
			{Claim: "Add function performs integer addition", File: filePath, Lines: "6-8"},
			{Claim: "Greet function prints greeting using fmt.Printf", File: filePath, Lines: "11-13"},
			{Claim: "Multiply function performs integer multiplication", File: filePath, Lines: "16-19"},
			{Claim: "Add function signature and parameters", File: filePath, Lines: "6"},
			{Claim: "Add function return statement with addition operation", File: filePath, Lines: "7"},
			{Claim: "Greet function signature and parameter", File: filePath, Lines: "11"},
			{Claim: "Printf call in Greet function", File: filePath, Lines: "12"},
			{Claim: "Multiply function signature and parameters", File: filePath, Lines: "16"},
			{Claim: "Multiplication operation in Multiply function", File: filePath, Lines: "17"},
			{Claim: "Multiply function return statement", File: filePath, Lines: "18"},
			{Claim: "Pure Function: Add", File: filePath, Lines: "6-8"},
			{Claim: "Pure Function: Multiply", File: filePath, Lines: "16-19"},
		},
	}
}

// generateMultipleFunctionsOutput creates output for multiple_functions.go
func generateMultipleFunctionsOutput(filePath string) *schema.AgentOutputV1 {
	return &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "multiple_functions",
		ScopeDescription: "Analysis of multi-step calculation pipeline with function composition",
		Overview:         "This module implements a calculation pipeline with multiple stages: Calculate orchestrates the flow, Process validates input, Transform applies business logic, and Finalize produces output. Helper provides utility string formatting.",

		EntryPoints: []schema.EntryPoint{
			{File: filePath, Lines: "6-10", Symbol: "Calculate", Role: "orchestrator"},
			{File: filePath, Lines: "13-18", Symbol: "Process", Role: "validator"},
			{File: filePath, Lines: "21-23", Symbol: "Transform", Role: "transformer"},
			{File: filePath, Lines: "26-30", Symbol: "Finalize", Role: "finalizer"},
			{File: filePath, Lines: "33-35", Symbol: "Helper", Role: "utility"},
		},

		CallGraph: []schema.CallGraphEdge{
			{From: filePath + ":Calculate", To: filePath + ":Process", ViaLine: 7},
			{From: filePath + ":Calculate", To: filePath + ":Transform", ViaLine: 8},
			{From: filePath + ":Calculate", To: filePath + ":Finalize", ViaLine: 9},
			{From: filePath + ":Finalize", To: "fmt.Printf", ViaLine: 28},
			{From: filePath + ":Helper", To: "fmt.Sprintf", ViaLine: 34},
		},

		DataFlow: schema.DataFlow{
			Inputs: []schema.DataPoint{
				{Source: filePath + ":6", Name: "input", Type: "int", Description: "Initial calculation input"},
				{Source: filePath + ":13", Name: "value", Type: "int", Description: "Value to process"},
				{Source: filePath + ":21", Name: "value", Type: "int", Description: "Value to transform"},
				{Source: filePath + ":26", Name: "value", Type: "int", Description: "Value to finalize"},
				{Source: filePath + ":33", Name: "msg", Type: "string", Description: "Message for helper"},
			},
			Transformations: []schema.Transformation{
				{File: filePath, Lines: "14-17", Operation: "validation", Description: "Validate input (negative check) and multiply by 2"},
				{File: filePath, Lines: "22", Operation: "arithmetic", Description: "Add 10 to value"},
				{File: filePath, Lines: "27", Operation: "arithmetic", Description: "Multiply by 3"},
				{File: filePath, Lines: "34", Operation: "formatting", Description: "Format string with prefix"},
			},
			Outputs: []schema.DataPoint{
				{Source: filePath + ":9", Name: "result", Type: "int", Description: "Final calculated result"},
				{Source: filePath + ":17", Name: "processed", Type: "int", Description: "Processed value"},
				{Source: filePath + ":22", Name: "transformed", Type: "int", Description: "Transformed value"},
				{Source: filePath + ":29", Name: "finalized", Type: "int", Description: "Finalized result"},
				{Source: filePath + ":34", Name: "formatted", Type: "string", Description: "Formatted helper message"},
			},
		},

		StateManagement: []schema.StateOperation{},

		SideEffects: []schema.SideEffect{
			{File: filePath, Line: 28, Type: "log", Description: "Finalize prints result to stdout"},
		},

		ErrorHandling: []schema.ErrorHandler{
			{File: filePath, Lines: "14-16", Type: "guard", Condition: "value < 0", Effect: "return 0"},
		},

		Configuration: []schema.ConfigInfluence{},

		Patterns: []schema.Pattern{
			{Name: "Pipeline Pattern", File: filePath, Lines: "6-10", Description: "Calculate composes multiple functions in sequence"},
			{Name: "Guard Clause", File: filePath, Lines: "14-16", Description: "Early return for invalid input"},
		},

		Concurrency:          []schema.ConcurrencyMechanism{},
		ExternalDependencies: []schema.ExternalDependency{{File: filePath, Line: 3, Module: "fmt", Purpose: "Formatted I/O operations"}},
		Limitations:          []string{"No error propagation from pipeline stages"},
		OpenQuestions:        []string{},

		RawEvidence: []schema.Evidence{
			{Claim: "Complete source code of multiple_functions.go", File: filePath, Lines: "1-36"},
			{Claim: "Calculate orchestrates multi-step pipeline", File: filePath, Lines: "6-10"},
			{Claim: "Process validates and doubles input", File: filePath, Lines: "13-18"},
			{Claim: "Transform adds 10 to value", File: filePath, Lines: "21-23"},
			{Claim: "Finalize multiplies by 3 and prints", File: filePath, Lines: "26-30"},
			{Claim: "Helper formats string with prefix", File: filePath, Lines: "33-35"},
			{Claim: "Pipeline Pattern in Calculate", File: filePath, Lines: "6-10"},
			{Claim: "Guard Clause in Process", File: filePath, Lines: "14-16"},
			{Claim: "Call to Process from Calculate", File: filePath, Lines: "7"},
			{Claim: "Call to Transform from Calculate", File: filePath, Lines: "8"},
			{Claim: "Call to Finalize from Calculate", File: filePath, Lines: "9"},
			{Claim: "Printf in Finalize", File: filePath, Lines: "28"},
			// CallGraph - missing call to Helper
			{Claim: "Call to Helper from Finalize", File: filePath, Lines: "34"},

			// DataFlow.Inputs - function parameters (file:line format)
			{Claim: "Input parameter x in Calculate", File: filePath, Lines: "6"},
			{Claim: "Input parameter n in Process", File: filePath, Lines: "13"},
			{Claim: "Input parameter val in Transform", File: filePath, Lines: "21"},
			{Claim: "Input parameter v in Finalize", File: filePath, Lines: "26"},
			{Claim: "Input parameter msg in Helper", File: filePath, Lines: "33"},

			// DataFlow.Transformations - computation steps
			{Claim: "Process validation and transformation", File: filePath, Lines: "14-17"},
			{Claim: "Transform arithmetic operation", File: filePath, Lines: "22"},
			{Claim: "Finalize arithmetic operation", File: filePath, Lines: "27"},
			{Claim: "Helper string formatting", File: filePath, Lines: "34"},

			// DataFlow.Outputs - return values (file:line format)
			{Claim: "Calculate return value", File: filePath, Lines: "9"},
			{Claim: "Process return value", File: filePath, Lines: "17"},
			{Claim: "Transform return value", File: filePath, Lines: "22"},
			{Claim: "Finalize return value", File: filePath, Lines: "29"},
			{Claim: "Helper return value", File: filePath, Lines: "34"},
		},
	}
}

// generateStructMethodsOutput creates output for struct_methods.go
func generateStructMethodsOutput(filePath string) *schema.AgentOutputV1 {
	return &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "struct_methods",
		ScopeDescription: "Analysis of stateful Calculator type with methods",
		Overview:         "This module implements a Calculator struct with stateful operations. It maintains Total and Count fields, providing methods for Add, Subtract, Average, Reset, and Display. NewCalculator is a constructor function.",

		EntryPoints: []schema.EntryPoint{
			{File: filePath, Lines: "12-17", Symbol: "NewCalculator", Role: "constructor"},
			{File: filePath, Lines: "20-23", Symbol: "Calculator.Add", Role: "mutator"},
			{File: filePath, Lines: "26-29", Symbol: "Calculator.Subtract", Role: "mutator"},
			{File: filePath, Lines: "32-37", Symbol: "Calculator.Average", Role: "accessor"},
			{File: filePath, Lines: "40-43", Symbol: "Calculator.Reset", Role: "mutator"},
			{File: filePath, Lines: "46-49", Symbol: "Calculator.Display", Role: "accessor"},
		},

		CallGraph: []schema.CallGraphEdge{
			{From: filePath + ":NewCalculator", To: "builtin:new", ViaLine: 13},
			{From: filePath + ":Calculator.Add", To: "builtin:+=", ViaLine: 21},
			{From: filePath + ":Calculator.Add", To: "builtin:++", ViaLine: 22},
			{From: filePath + ":Calculator.Subtract", To: "builtin:-=", ViaLine: 27},
			{From: filePath + ":Calculator.Subtract", To: "builtin:++", ViaLine: 28},
			{From: filePath + ":Calculator.Average", To: "builtin:/", ViaLine: 36},
			{From: filePath + ":Calculator.Display", To: "fmt.Printf", ViaLine: 47},
			{From: filePath + ":Calculator.Display", To: filePath + ":Calculator.Average", ViaLine: 48},
		},

		DataFlow: schema.DataFlow{
			Inputs: []schema.DataPoint{
				{Source: filePath + ":20", Name: "value", Type: "int", Description: "Value to add"},
				{Source: filePath + ":26", Name: "value", Type: "int", Description: "Value to subtract"},
			},
			Transformations: []schema.Transformation{
				{File: filePath, Lines: "21-22", Operation: "state-mutation", Description: "Add value to Total and increment Count"},
				{File: filePath, Lines: "27-28", Operation: "state-mutation", Description: "Subtract value from Total and increment Count"},
				{File: filePath, Lines: "36", Operation: "arithmetic", Description: "Calculate average from Total and Count"},
				{File: filePath, Lines: "41-42", Operation: "state-mutation", Description: "Reset Total and Count to 0"},
			},
			Outputs: []schema.DataPoint{
				{Source: filePath + ":13", Name: "*Calculator", Type: "*Calculator", Description: "New Calculator instance"},
				{Source: filePath + ":36", Name: "average", Type: "float64", Description: "Calculated average"},
			},
		},

		StateManagement: []schema.StateOperation{
			{File: filePath, Lines: "7-8", Kind: "struct-field", Description: "Calculator maintains Total (int) and Count (int) state"},
			{File: filePath, Lines: "21", Kind: "mutation", Description: "Add mutates Total field"},
			{File: filePath, Lines: "22", Kind: "mutation", Description: "Add mutates Count field"},
			{File: filePath, Lines: "27", Kind: "mutation", Description: "Subtract mutates Total field"},
			{File: filePath, Lines: "28", Kind: "mutation", Description: "Subtract mutates Count field"},
			{File: filePath, Lines: "41-42", Kind: "mutation", Description: "Reset clears all state"},
		},

		SideEffects: []schema.SideEffect{
			{File: filePath, Line: 47, Type: "log", Description: "Display prints state to stdout"},
		},

		ErrorHandling: []schema.ErrorHandler{
			{File: filePath, Lines: "33-35", Type: "guard", Condition: "c.count == 0", Effect: "return 0.0"},
		},

		Configuration: []schema.ConfigInfluence{},

		Patterns: []schema.Pattern{
			{Name: "Constructor Pattern", File: filePath, Lines: "12-17", Description: "NewCalculator initializes Calculator with zero values"},
			{Name: "Stateful Object", File: filePath, Lines: "6-9", Description: "Calculator maintains mutable state across method calls"},
			{Name: "Guard Clause", File: filePath, Lines: "33-35", Description: "Average protects against division by zero"},
		},

		Concurrency:          []schema.ConcurrencyMechanism{},
		ExternalDependencies: []schema.ExternalDependency{{File: filePath, Line: 3, Module: "fmt", Purpose: "Formatted I/O operations"}},
		Limitations:          []string{"Not thread-safe", "No overflow protection"},
		OpenQuestions:        []string{},

		RawEvidence: []schema.Evidence{
			{Claim: "Complete source code of struct_methods.go", File: filePath, Lines: "1-50"},
			{Claim: "Calculator struct definition", File: filePath, Lines: "6-9"},
			{Claim: "NewCalculator constructor", File: filePath, Lines: "12-17"},
			{Claim: "Add method mutates state", File: filePath, Lines: "20-23"},
			{Claim: "Subtract method mutates state", File: filePath, Lines: "26-29"},
			{Claim: "Average method with guard clause", File: filePath, Lines: "32-37"},
			{Claim: "Reset method clears state", File: filePath, Lines: "40-43"},
			{Claim: "Display method with Printf", File: filePath, Lines: "46-49"},
			{Claim: "Constructor Pattern", File: filePath, Lines: "12-17"},
			{Claim: "Stateful Object Pattern", File: filePath, Lines: "6-9"},
			{Claim: "Guard Clause in Average", File: filePath, Lines: "33-35"},
			{Claim: "Total field mutation in Add", File: filePath, Lines: "21"},
			{Claim: "Count field mutation in Add", File: filePath, Lines: "22"},
			// CallGraph evidence for method calls
			{Claim: "NewCalculator allocates struct", File: filePath, Lines: "13"},
			{Claim: "Add operator in Add method", File: filePath, Lines: "21"},
			{Claim: "Increment operator in Add method", File: filePath, Lines: "22"},
			{Claim: "Subtract operator in Subtract method", File: filePath, Lines: "27"},
			{Claim: "Increment operator in Subtract method", File: filePath, Lines: "28"},
			{Claim: "Division in Average method", File: filePath, Lines: "36"},
			{Claim: "Printf in Display method", File: filePath, Lines: "47"},
			{Claim: "Average call from Display", File: filePath, Lines: "48"},

			// DataFlow.Inputs
			{Claim: "Add value parameter", File: filePath, Lines: "20"},
			{Claim: "Subtract value parameter", File: filePath, Lines: "26"},

			// DataFlow.Outputs
			{Claim: "NewCalculator return value", File: filePath, Lines: "13"},
			{Claim: "Average return value", File: filePath, Lines: "36"},

			// SideEffects
			{Claim: "Display Printf side effect", File: filePath, Lines: "47"},

			// ErrorHandling
			{Claim: "Average guard clause", File: filePath, Lines: "33-35"},

			// StateManagement evidence
			{Claim: "Calculator struct fields", File: filePath, Lines: "7-8"},
			{Claim: "Total mutation in Add", File: filePath, Lines: "21"},
			{Claim: "Count mutation in Add", File: filePath, Lines: "22"},
			{Claim: "Total mutation in Subtract", File: filePath, Lines: "27"},
			{Claim: "Count mutation in Subtract", File: filePath, Lines: "28"},
			{Claim: "Reset state mutation", File: filePath, Lines: "41-42"},

			// DataFlow.Transformations evidence
			{Claim: "Add state mutation transformation", File: filePath, Lines: "21-22"},
			{Claim: "Subtract state mutation transformation", File: filePath, Lines: "27-28"},
			{Claim: "Average arithmetic transformation", File: filePath, Lines: "36"},
			{Claim: "Reset state mutation transformation", File: filePath, Lines: "41-42"},
		},
	}
}

// generateErrorHandlingOutput creates output for error_handling.go
func generateErrorHandlingOutput(filePath string) *schema.AgentOutputV1 {
	return &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "error_handling",
		ScopeDescription: "Analysis of error handling patterns and error propagation",
		Overview:         "This module demonstrates comprehensive error handling patterns in Go: sentinel errors, error wrapping, validation, retry logic, and error propagation. It includes Divide, Fetch, ProcessWithRetry, and ValidateAndProcess functions with helper utilities.",

		EntryPoints: []schema.EntryPoint{
			{File: filePath, Lines: "15-20", Symbol: "Divide", Role: "validator"},
			{File: filePath, Lines: "23-31", Symbol: "Fetch", Role: "retriever"},
			{File: filePath, Lines: "34-48", Symbol: "ProcessWithRetry", Role: "processor"},
			{File: filePath, Lines: "58-74", Symbol: "ValidateAndProcess", Role: "processor"},
			{File: filePath, Lines: "50-55", Symbol: "attemptProcess", Role: "helper"},
			{File: filePath, Lines: "76-78", Symbol: "validate", Role: "helper"},
			{File: filePath, Lines: "80-82", Symbol: "process", Role: "helper"},
		},

		CallGraph: []schema.CallGraphEdge{
			{From: filePath + ":ProcessWithRetry", To: filePath + ":attemptProcess", ViaLine: 39},
			{From: filePath + ":ProcessWithRetry", To: filePath + ":attemptProcess", ViaLine: 42},
			{From: filePath + ":ProcessWithRetry", To: "fmt.Errorf", ViaLine: 44},
			{From: filePath + ":ValidateAndProcess", To: filePath + ":validate", ViaLine: 63},
			{From: filePath + ":ValidateAndProcess", To: filePath + ":process", ViaLine: 68},
			{From: filePath + ":ValidateAndProcess", To: "fmt.Errorf", ViaLine: 70},
		},

		DataFlow: schema.DataFlow{
			Inputs: []schema.DataPoint{
				{Source: filePath + ":15", Name: "a, b", Type: "int", Description: "Division operands"},
				{Source: filePath + ":23", Name: "id", Type: "int", Description: "Record ID to fetch"},
				{Source: filePath + ":34", Name: "data", Type: "string", Description: "Data to process with retry"},
				{Source: filePath + ":58", Name: "input", Type: "string", Description: "Input to validate and process"},
			},
			Transformations: []schema.Transformation{
				{File: filePath, Lines: "19", Operation: "arithmetic", Description: "Integer division"},
				{File: filePath, Lines: "30", Operation: "formatting", Description: "Format data string with ID"},
				{File: filePath, Lines: "81", Operation: "formatting", Description: "Prepend 'processed:' to input"},
			},
			Outputs: []schema.DataPoint{
				{Source: filePath + ":19", Name: "result", Type: "int", Description: "Division result or error"},
				{Source: filePath + ":30", Name: "data", Type: "string", Description: "Fetched data or error"},
				{Source: filePath + ":47", Name: "error", Type: "error", Description: "Process result or retry error"},
				{Source: filePath + ":73", Name: "result", Type: "string", Description: "Processed result or validation error"},
			},
		},

		StateManagement: []schema.StateOperation{},

		SideEffects: []schema.SideEffect{},

		ErrorHandling: []schema.ErrorHandler{
			{File: filePath, Lines: "9-11", Type: "guard", Condition: "errors.Is(err, ErrInvalidInput)", Effect: "propagate"},
			{File: filePath, Lines: "16-18", Type: "guard", Condition: "b == 0", Effect: "propagate"},
			{File: filePath, Lines: "24-26", Type: "guard", Condition: "id < 0", Effect: "propagate"},
			{File: filePath, Lines: "27-29", Type: "guard", Condition: "id > 100", Effect: "propagate"},
			{File: filePath, Lines: "35-37", Type: "guard", Condition: "data == \"\"", Effect: "propagate"},
			{File: filePath, Lines: "39-46", Type: "retry", Condition: "maxRetries = 2", Effect: "retry"},
			{File: filePath, Lines: "44", Type: "catch", Condition: "err != nil", Effect: "propagate"},
			{File: filePath, Lines: "51-53", Type: "guard", Condition: "len(data) < 3", Effect: "propagate"},
			{File: filePath, Lines: "60-62", Type: "guard", Condition: "input == \"\"", Effect: "propagate"},
			{File: filePath, Lines: "63-66", Type: "guard", Condition: "validate(input) != nil", Effect: "propagate"},
			{File: filePath, Lines: "68-71", Type: "catch", Condition: "err != nil", Effect: "propagate"},
		},

		Configuration: []schema.ConfigInfluence{},

		Patterns: []schema.Pattern{
			{Name: "Sentinel Errors", File: filePath, Lines: "9-11", Description: "Define package-level error constants"},
			{Name: "Error Wrapping", File: filePath, Lines: "44", Description: "Use fmt.Errorf with %w to wrap errors"},
			{Name: "Retry Logic", File: filePath, Lines: "39-46", Description: "Attempt operation twice before returning error"},
			{Name: "Validation Chain", File: filePath, Lines: "58-74", Description: "Multiple validation steps with early returns"},
		},

		Concurrency: []schema.ConcurrencyMechanism{},
		ExternalDependencies: []schema.ExternalDependency{
			{File: filePath, Line: 4, Module: "errors", Purpose: "Standard error creation and handling"},
			{File: filePath, Line: 5, Module: "fmt", Purpose: "Error formatting and wrapping"},
		},
		Limitations:   []string{"Fixed retry count (no exponential backoff)", "No timeout handling in retry"},
		OpenQuestions: []string{},

		RawEvidence: []schema.Evidence{
			{Claim: "Complete source code of error_handling.go", File: filePath, Lines: "1-83"},
			{Claim: "Sentinel errors definition", File: filePath, Lines: "9-11"},
			{Claim: "Divide function with validation", File: filePath, Lines: "15-20"},
			{Claim: "Fetch function with multiple error conditions", File: filePath, Lines: "23-31"},
			{Claim: "ProcessWithRetry with retry logic", File: filePath, Lines: "34-48"},
			{Claim: "ValidateAndProcess with validation chain", File: filePath, Lines: "58-74"},
			{Claim: "attemptProcess helper", File: filePath, Lines: "50-55"},
			{Claim: "validate helper", File: filePath, Lines: "76-78"},
			{Claim: "process helper", File: filePath, Lines: "80-82"},
			{Claim: "Error wrapping pattern", File: filePath, Lines: "44"},
			{Claim: "Retry Logic pattern", File: filePath, Lines: "39-46"},
			{Claim: "Division by zero check", File: filePath, Lines: "16-18"},
			{Claim: "Negative ID validation", File: filePath, Lines: "24-26"},
			{Claim: "ID range validation", File: filePath, Lines: "27-29"},
			// CallGraph evidence
			{Claim: "ProcessWithRetry calls attemptProcess (first attempt)", File: filePath, Lines: "39"},
			{Claim: "ProcessWithRetry calls attemptProcess (retry)", File: filePath, Lines: "42"},
			{Claim: "ProcessWithRetry wraps error with fmt.Errorf", File: filePath, Lines: "44"},
			{Claim: "ValidateAndProcess calls validate helper", File: filePath, Lines: "63"},
			{Claim: "ValidateAndProcess calls process helper", File: filePath, Lines: "68"},
			{Claim: "ValidateAndProcess wraps error with fmt.Errorf", File: filePath, Lines: "70"},
			// DataFlow.Inputs evidence
			{Claim: "Divide input parameters a, b", File: filePath, Lines: "15"},
			{Claim: "Fetch input parameter id", File: filePath, Lines: "23"},
			{Claim: "ProcessWithRetry input parameter data", File: filePath, Lines: "34"},
			{Claim: "ValidateAndProcess input parameter input", File: filePath, Lines: "58"},
			// DataFlow.Transformations evidence
			{Claim: "Division operation a / b", File: filePath, Lines: "19"},
			{Claim: "Format string with ID", File: filePath, Lines: "30"},
			{Claim: "Process transformation prepends processed: ", File: filePath, Lines: "81"},
			// DataFlow.Outputs evidence
			{Claim: "Divide returns division result", File: filePath, Lines: "19"},
			{Claim: "Fetch returns formatted data string", File: filePath, Lines: "30"},
			{Claim: "ProcessWithRetry returns error or nil", File: filePath, Lines: "47"},
			{Claim: "ValidateAndProcess returns processed result", File: filePath, Lines: "73"},
			// ErrorHandling detailed evidence
			{Claim: "Empty data validation in ProcessWithRetry", File: filePath, Lines: "35-37"},
			{Claim: "Empty input validation in ValidateAndProcess", File: filePath, Lines: "59-62"},
			{Claim: "Error handler guard for empty input", File: filePath, Lines: "60-62"},
			{Claim: "Validation check result", File: filePath, Lines: "63-66"},
			{Claim: "Process error handling", File: filePath, Lines: "68-71"},
			{Claim: "attemptProcess data length validation", File: filePath, Lines: "51-53"},
		},
	}
}

// generateSideEffectsOutput creates output for side_effects.go
func generateSideEffectsOutput(filePath string) *schema.AgentOutputV1 {
	return &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    "side_effects",
		ScopeDescription: "Analysis of functions with various side effects (I/O, logging, external calls)",
		Overview:         "This module demonstrates multiple types of side effects: logging (log.Printf), file I/O (os.WriteFile/ReadFile), console output (fmt.Println), metrics recording, and simulated HTTP calls. Functions include LogOperation, WriteToFile, ReadFromFile, ProcessWithSideEffects, and NotifyUser.",

		EntryPoints: []schema.EntryPoint{
			{File: filePath, Lines: "10-15", Symbol: "LogOperation", Role: "processor"},
			{File: filePath, Lines: "18-27", Symbol: "WriteToFile", Role: "writer"},
			{File: filePath, Lines: "30-39", Symbol: "ReadFromFile", Role: "reader"},
			{File: filePath, Lines: "42-56", Symbol: "ProcessWithSideEffects", Role: "processor"},
			{File: filePath, Lines: "67-72", Symbol: "NotifyUser", Role: "notifier"},
			{File: filePath, Lines: "58-60", Symbol: "recordMetric", Role: "helper"},
			{File: filePath, Lines: "62-64", Symbol: "transform", Role: "helper"},
		},

		CallGraph: []schema.CallGraphEdge{
			{From: filePath + ":LogOperation", To: "log.Printf", ViaLine: 11},
			{From: filePath + ":LogOperation", To: "log.Printf", ViaLine: 13},
			{From: filePath + ":WriteToFile", To: "log.Printf", ViaLine: 19},
			{From: filePath + ":WriteToFile", To: "os.WriteFile", ViaLine: 20},
			{From: filePath + ":WriteToFile", To: "log.Printf", ViaLine: 22},
			{From: filePath + ":WriteToFile", To: "log.Printf", ViaLine: 25},
			{From: filePath + ":ReadFromFile", To: "log.Printf", ViaLine: 31},
			{From: filePath + ":ReadFromFile", To: "os.ReadFile", ViaLine: 32},
			{From: filePath + ":ReadFromFile", To: "log.Printf", ViaLine: 34},
			{From: filePath + ":ReadFromFile", To: "log.Printf", ViaLine: 37},
			{From: filePath + ":ProcessWithSideEffects", To: "fmt.Println", ViaLine: 44},
			{From: filePath + ":ProcessWithSideEffects", To: filePath + ":recordMetric", ViaLine: 47},
			{From: filePath + ":ProcessWithSideEffects", To: filePath + ":transform", ViaLine: 50},
			{From: filePath + ":ProcessWithSideEffects", To: "fmt.Println", ViaLine: 53},
			{From: filePath + ":NotifyUser", To: "log.Printf", ViaLine: 68},
			{From: filePath + ":NotifyUser", To: "fmt.Printf", ViaLine: 70},
			{From: filePath + ":recordMetric", To: "fmt.Printf", ViaLine: 59},
		},

		DataFlow: schema.DataFlow{
			Inputs: []schema.DataPoint{
				{Source: filePath + ":10", Name: "name, value", Type: "string, int", Description: "Operation name and value"},
				{Source: filePath + ":18", Name: "filename, data", Type: "string, []byte", Description: "File path and content to write"},
				{Source: filePath + ":30", Name: "filename", Type: "string", Description: "File path to read"},
				{Source: filePath + ":42", Name: "input", Type: "string", Description: "Input to process"},
				{Source: filePath + ":67", Name: "userID, message", Type: "int, string", Description: "User ID and notification message"},
			},
			Transformations: []schema.Transformation{
				{File: filePath, Lines: "12", Operation: "arithmetic", Description: "Multiply value by 2"},
				{File: filePath, Lines: "20", Operation: "file-write", Description: "Write bytes to file"},
				{File: filePath, Lines: "32", Operation: "file-read", Description: "Read bytes from file"},
				{File: filePath, Lines: "63", Operation: "formatting", Description: "Prepend 'transformed:' to string"},
			},
			Outputs: []schema.DataPoint{
				{Source: filePath + ":14", Name: "result", Type: "int", Description: "Doubled value"},
				{Source: filePath + ":20", Name: "error", Type: "error", Description: "Write error or nil"},
				{Source: filePath + ":32", Name: "data, error", Type: "[]byte, error", Description: "Read bytes and error"},
				{Source: filePath + ":55", Name: "result", Type: "string", Description: "Transformed string"},
				{Source: filePath + ":70", Name: "error", Type: "error", Description: "Notification error or nil"},
			},
		},

		StateManagement: []schema.StateOperation{},

		SideEffects: []schema.SideEffect{
			{File: filePath, Line: 11, Type: "log", Description: "log.Printf: Starting operation"},
			{File: filePath, Line: 13, Type: "log", Description: "log.Printf: Operation completed"},
			{File: filePath, Line: 19, Type: "log", Description: "log.Printf: Writing bytes to file"},
			{File: filePath, Line: 20, Type: "file-write", Description: "os.WriteFile: Write data to filesystem"},
			{File: filePath, Line: 22, Type: "log", Description: "log.Printf: Error writing to file"},
			{File: filePath, Line: 25, Type: "log", Description: "log.Printf: Successfully wrote to file"},
			{File: filePath, Line: 31, Type: "log", Description: "log.Printf: Reading from file"},
			{File: filePath, Line: 32, Type: "file-read", Description: "os.ReadFile: Read data from filesystem"},
			{File: filePath, Line: 34, Type: "log", Description: "log.Printf: Error reading file"},
			{File: filePath, Line: 37, Type: "log", Description: "log.Printf: Successfully read from file"},
			{File: filePath, Line: 44, Type: "console", Description: "fmt.Println: Processing input"},
			{File: filePath, Line: 47, Type: "metric", Description: "recordMetric: Increment process_count"},
			{File: filePath, Line: 53, Type: "console", Description: "fmt.Println: Processing complete"},
			{File: filePath, Line: 59, Type: "console", Description: "fmt.Printf: Record metric"},
			{File: filePath, Line: 68, Type: "log", Description: "log.Printf: Sending notification"},
			{File: filePath, Line: 70, Type: "http", Description: "Simulated HTTP POST to notification endpoint"},
		},

		ErrorHandling: []schema.ErrorHandler{
			{File: filePath, Lines: "21-24", Type: "catch", Condition: "err != nil", Effect: "fallback"},
			{File: filePath, Lines: "33-36", Type: "catch", Condition: "err != nil", Effect: "fallback"},
		},

		Configuration: []schema.ConfigInfluence{},

		Patterns: []schema.Pattern{
			{Name: "Logging Pattern", File: filePath, Lines: "10-15", Description: "Log operation start and completion"},
			{Name: "File I/O with Logging", File: filePath, Lines: "18-27", Description: "Log file operations and errors"},
			{Name: "Multiple Side Effects", File: filePath, Lines: "42-56", Description: "Function combines logging, metrics, and state transformation"},
			{Name: "Simulated External Call", File: filePath, Lines: "67-72", Description: "Simulate HTTP notification"},
		},

		Concurrency: []schema.ConcurrencyMechanism{},
		ExternalDependencies: []schema.ExternalDependency{
			{File: filePath, Line: 4, Module: "fmt", Purpose: "Formatted I/O operations"},
			{File: filePath, Line: 5, Module: "log", Purpose: "Logging operations"},
			{File: filePath, Line: 6, Module: "os", Purpose: "File system operations"},
		},
		Limitations:   []string{"No buffered I/O", "Synchronous file operations", "No HTTP retry logic"},
		OpenQuestions: []string{},

		RawEvidence: []schema.Evidence{
			// General evidence
			{Claim: "Complete source code of side_effects.go", File: filePath, Lines: "1-73"},

			// Entry points
			{Claim: "LogOperation with logging side effects", File: filePath, Lines: "10-15"},
			{Claim: "WriteToFile with file I/O", File: filePath, Lines: "18-27"},
			{Claim: "ReadFromFile with file I/O", File: filePath, Lines: "30-39"},
			{Claim: "ProcessWithSideEffects with multiple effects", File: filePath, Lines: "42-56"},
			{Claim: "NotifyUser with simulated HTTP", File: filePath, Lines: "67-72"},
			{Claim: "recordMetric helper", File: filePath, Lines: "58-60"},
			{Claim: "transform helper", File: filePath, Lines: "62-64"},

			// Data flow - Inputs (5 items)
			{Claim: "Input name, value at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:10", File: filePath, Lines: "10"},
			{Claim: "Input filename, data at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:18", File: filePath, Lines: "18"},
			{Claim: "Input filename at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:30", File: filePath, Lines: "30"},
			{Claim: "Input input at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:42", File: filePath, Lines: "42"},
			{Claim: "Input userID, message at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:67", File: filePath, Lines: "67"},

			// Data flow - Transformations (4 items)
			{Claim: "Transformation arithmetic at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:12", File: filePath, Lines: "12"},
			{Claim: "Transformation file-write at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:20", File: filePath, Lines: "20"},
			{Claim: "Transformation file-read at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:32", File: filePath, Lines: "32"},
			{Claim: "Transformation formatting at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:63", File: filePath, Lines: "63"},

			// Data flow - Outputs (5 items)
			{Claim: "Output result at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:14", File: filePath, Lines: "14"},
			{Claim: "Output error at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:20", File: filePath, Lines: "20"},
			{Claim: "Output data, error at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:32", File: filePath, Lines: "32"},
			{Claim: "Output result at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:55", File: filePath, Lines: "55"},
			{Claim: "Output error at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:70", File: filePath, Lines: "70"},

			// Side effects (16 items - matching lines 11,13,19,20,22,25,31,32,34,37,44,47,53,59,68,70)
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:11", File: filePath, Lines: "11"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:13", File: filePath, Lines: "13"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:19", File: filePath, Lines: "19"},
			{Claim: "Side effect file-write at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:20", File: filePath, Lines: "20"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:22", File: filePath, Lines: "22"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:25", File: filePath, Lines: "25"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:31", File: filePath, Lines: "31"},
			{Claim: "Side effect file-read at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:32", File: filePath, Lines: "32"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:34", File: filePath, Lines: "34"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:37", File: filePath, Lines: "37"},
			{Claim: "Side effect console at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:44", File: filePath, Lines: "44"},
			{Claim: "Side effect metric at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:47", File: filePath, Lines: "47"},
			{Claim: "Side effect console at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:53", File: filePath, Lines: "53"},
			{Claim: "Side effect console at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:59", File: filePath, Lines: "59"},
			{Claim: "Side effect log at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:68", File: filePath, Lines: "68"},
			{Claim: "Side effect http at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:70", File: filePath, Lines: "70"},

			// Error handling (2 items)
			{Claim: "Error handler catch at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:21-24", File: filePath, Lines: "21-24"},
			{Claim: "Error handler catch at /home/f3rg/src/github/conexus/tests/fixtures/side_effects.go:33-36", File: filePath, Lines: "33-36"},

			// Specific call sites
			{Claim: "log.Printf at start of LogOperation", File: filePath, Lines: "11"},
			{Claim: "log.Printf at end of LogOperation", File: filePath, Lines: "13"},
			{Claim: "os.WriteFile call", File: filePath, Lines: "20"},
			{Claim: "os.ReadFile call", File: filePath, Lines: "32"},
			{Claim: "fmt.Println for processing", File: filePath, Lines: "44"},
			{Claim: "recordMetric call", File: filePath, Lines: "47"},
			{Claim: "Simulated HTTP POST", File: filePath, Lines: "70"},
		},
	}
}

// TestSimpleFunctionAnalysis tests the analyzer agent on simple_function.go
func TestSimpleFunctionAnalysis(t *testing.T) {
	runFixtureTest(t, "simple_function.go", "Simple Function Analysis", []integration.Assertion{
		&integration.WorkflowSuccessAssertion{},
		&integration.StepCountAssertion{ExpectedCount: 1},
		&integration.AllStepsSuccessAssertion{},
		&integration.MaxDurationAssertion{MaxDuration: 5 * time.Second},
		&integration.OutputNotNilAssertion{},
		&integration.SchemaValidAssertion{},
		&integration.EvidenceValidAssertion{StrictMode: true},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "entry_points"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "raw_evidence"},
	})
}

// TestMultipleFunctionsAnalysis tests the analyzer on a more complex fixture
func TestMultipleFunctionsAnalysis(t *testing.T) {
	runFixtureTest(t, "multiple_functions.go", "Multiple Functions Analysis", []integration.Assertion{
		&integration.WorkflowSuccessAssertion{},
		&integration.AllStepsSuccessAssertion{},
		&integration.MaxDurationAssertion{MaxDuration: 5 * time.Second},
		&integration.OutputNotNilAssertion{},
		&integration.SchemaValidAssertion{},
		&integration.EvidenceValidAssertion{StrictMode: true},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "entry_points"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "call_graph"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "patterns"},
	})
}

// TestStructMethodsAnalysis tests the analyzer on struct methods
func TestStructMethodsAnalysis(t *testing.T) {
	runFixtureTest(t, "struct_methods.go", "Struct Methods Analysis", []integration.Assertion{
		&integration.WorkflowSuccessAssertion{},
		&integration.AllStepsSuccessAssertion{},
		&integration.MaxDurationAssertion{MaxDuration: 5 * time.Second},
		&integration.OutputNotNilAssertion{},
		&integration.SchemaValidAssertion{},
		&integration.EvidenceValidAssertion{StrictMode: true},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "entry_points"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "state_management"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "patterns"},
	})
}

// TestErrorHandlingAnalysis tests the analyzer on error handling patterns
func TestErrorHandlingAnalysis(t *testing.T) {
	runFixtureTest(t, "error_handling.go", "Error Handling Analysis", []integration.Assertion{
		&integration.WorkflowSuccessAssertion{},
		&integration.AllStepsSuccessAssertion{},
		&integration.MaxDurationAssertion{MaxDuration: 5 * time.Second},
		&integration.OutputNotNilAssertion{},
		&integration.SchemaValidAssertion{},
		&integration.EvidenceValidAssertion{StrictMode: true},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "entry_points"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "error_handling"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "patterns"},
	})
}

// TestSideEffectsAnalysis tests the analyzer on functions with side effects
func TestSideEffectsAnalysis(t *testing.T) {
	runFixtureTest(t, "side_effects.go", "Side Effects Analysis", []integration.Assertion{
		&integration.WorkflowSuccessAssertion{},
		&integration.AllStepsSuccessAssertion{},
		&integration.MaxDurationAssertion{MaxDuration: 5 * time.Second},
		&integration.OutputNotNilAssertion{},
		&integration.SchemaValidAssertion{},
		&integration.EvidenceValidAssertion{StrictMode: true},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "entry_points"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "side_effects"},
		&integration.OutputFieldNotEmptyAssertion{FieldName: "external_dependencies"},
	})
}

// runFixtureTest is a helper to reduce test code duplication
func runFixtureTest(t *testing.T, fixtureName, testName string, assertions []integration.Assertion) {
	t.Helper()

	// Setup test framework
	framework := integration.NewTestFramework()
	framework.RegisterAgent("codebase-analyzer", &MockAnalyzerAgent{})

	// Get fixture path
	fixtureDir, err := filepath.Abs(filepath.Join("..", "..", "..", "tests", "fixtures"))
	require.NoError(t, err, "Should resolve fixtures directory")

	fixturePath := filepath.Join(fixtureDir, fixtureName)
	fixturePath, err = filepath.Abs(fixturePath)
	require.NoError(t, err, "Should resolve fixture path")

	// Verify fixture exists
	_, err = os.Stat(fixturePath)
	require.NoError(t, err, "Test fixture should exist: %s", fixturePath)

	// Create input
	inputData := map[string]interface{}{
		"file_path": fixturePath,
		"scope":     "function-level",
	}
	inputJSON, err := json.Marshal(inputData)
	require.NoError(t, err, "Should marshal input")

	// Create workflow
	wf, err := workflow.NewBuilder("analyze-"+strings.TrimSuffix(fixtureName, ".go")).
		WithDescription("Analyze "+fixtureName+" fixture").
		AddSequentialStep(
			"analyze",
			"codebase-analyzer",
			string(inputJSON),
			schema.Permissions{
				AllowedDirectories: []string{fixtureDir},
				ReadOnly:           true,
				MaxFileSize:        1024 * 1024,
				MaxExecutionTime:   10,
			},
		).
		Build()
	require.NoError(t, err, "Should build workflow")

	// Define test case
	testCase := &integration.TestCase{
		Name:        testName,
		Description: "Test analyzer agent on " + fixtureName,
		Workflow:    wf,
		Timeout:     10 * time.Second,
		Assertions:  assertions,
	}

	// Execute test
	result := framework.Run(context.Background(), testCase)

	// Debug output
	t.Logf("Test result - Passed: %v", result.Passed)
	if len(result.Errors) > 0 {
		t.Logf("Test errors: %v", result.Errors)
	}

	// Show assertion results
	for i, assertion := range result.Assertions {
		t.Logf("Assertion %d: %s - Passed: %v", i+1, assertion.Description, assertion.Passed)
		if !assertion.Passed {
			t.Logf("  Error: %v", assertion.Error)
		}
	}

	// Assertions
	require.True(t, result.Passed, "Test should pass")
	require.Empty(t, result.Errors, "Should have no errors")
	require.NotNil(t, result.WorkflowResult, "Should have workflow result")

	// Additional verification
	require.NotEmpty(t, result.WorkflowResult.StepResults, "Should have step results")
	stepResult := result.WorkflowResult.StepResults[0]
	require.NotNil(t, stepResult.Output, "Step output should not be nil")

	output := stepResult.Output
	require.NotEmpty(t, output.ComponentName, "Component name should be set")
	require.NotEmpty(t, output.EntryPoints, "Should have entry points")
	require.NotEmpty(t, output.RawEvidence, "Should have raw evidence")

	t.Logf("✓ Test completed in %v", result.Duration)
	t.Logf("✓ Found %d entry points", len(output.EntryPoints))
	t.Logf("✓ Found %d evidence items", len(output.RawEvidence))
}

// TestRunWorkflowHelper demonstrates the simplified workflow execution helper (5.1.1)
func TestRunWorkflowHelper(t *testing.T) {
	// Setup
	framework := integration.NewTestFramework()
	mockAnalyzer := &MockAnalyzerAgent{}
	framework.RegisterAgent("codebase-analyzer", mockAnalyzer)

	// Get fixture path
	wd, err := os.Getwd()
	require.NoError(t, err)
	projectRoot := filepath.Join(wd, "../../..")
	fixtureDir := filepath.Join(projectRoot, "tests/fixtures")
	fixturePath := filepath.Join(fixtureDir, "simple_function.go")

	// NEW WAY (simple - 8 lines):
	result, err := framework.RunWorkflow(context.Background(), integration.WorkflowConfig{
		ID:          "analyze-simple",
		Description: "Analyze simple_function.go fixture",
		Agent:       "codebase-analyzer",
		Input: map[string]interface{}{
			"file_path": fixturePath,
			"scope":     "function-level",
		},
		Timeout: 10 * time.Second,
		Permissions: schema.Permissions{
			AllowedDirectories: []string{fixtureDir},
			ReadOnly:           true,
			MaxFileSize:        1024 * 1024,
			MaxExecutionTime:   10,
		},
	})

	// Verify
	require.NoError(t, err, "RunWorkflow should succeed")
	require.NotNil(t, result, "Result should not be nil")
	require.Equal(t, workflow.StatusCompleted, result.Status)
	require.Len(t, result.StepResults, 1)

	step := result.StepResults[0]
	require.Equal(t, workflow.StepStatusCompleted, step.Status)
	require.NotNil(t, step.Output)
	require.Equal(t, "simple_function", step.Output.ComponentName)

	t.Logf("✓ RunWorkflow helper successfully executed workflow")
	t.Logf("✓ Simplified API reduced code by ~60%% (8 lines vs 20+ lines)")
}

// Fix for mock agents - replace lines 868-end

// ============================================================================
// Phase 5.1.3: Advanced Workflow Tests - Mock Agents
// ============================================================================

// MockParallelAgent simulates an agent that executes quickly for parallel tests
type MockParallelAgent struct {
	Name       string
	Duration   time.Duration
	ExecutedAt time.Time
}

func (m *MockParallelAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	m.ExecutedAt = time.Now()
	time.Sleep(m.Duration)

	return schema.AgentResponse{
		Status: schema.StatusComplete,
		Output: &schema.AgentOutputV1{
			Version:          "AGENT_OUTPUT_V1",
			ComponentName:    fmt.Sprintf("parallel-component-%s", m.Name),
			ScopeDescription: fmt.Sprintf("Parallel execution test by %s", m.Name),
			Overview:         fmt.Sprintf("This is a mock parallel agent for testing concurrent workflow execution. Agent %s completes work in %v.", m.Name, m.Duration),
			EntryPoints:      []schema.EntryPoint{},
			CallGraph:        []schema.CallGraphEdge{},
			DataFlow:         schema.DataFlow{},
			RawEvidence:      []schema.Evidence{},
		},
	}, nil
}

// MockEscalatingAgent simulates an agent that escalates to a supervisor
type MockEscalatingAgent struct {
	Name string
}

func (m *MockEscalatingAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	return schema.AgentResponse{
		Status: schema.StatusEscalationRequired,
		Output: &schema.AgentOutputV1{
			Version:          "AGENT_OUTPUT_V1",
			ComponentName:    "escalating-component",
			ScopeDescription: "Task requiring escalation",
			Overview:         "This task has exceeded complexity threshold and requires supervisor intervention.",
			EntryPoints:      []schema.EntryPoint{},
			CallGraph:        []schema.CallGraphEdge{},
			DataFlow:         schema.DataFlow{},
			RawEvidence:      []schema.Evidence{},
		},
		Escalation: &schema.Escalation{
			TargetAgent:  "supervisor-agent",
			Required:     true,
			Reason:       "complexity_threshold_exceeded",
			RequiredInfo: "architectural_context,performance_requirements",
		},
	}, nil
}

// MockSupervisorAgent simulates an agent that handles escalated work
type MockSupervisorAgent struct{}

func (m *MockSupervisorAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	return schema.AgentResponse{
		Status: schema.StatusComplete,
		Output: &schema.AgentOutputV1{
			Version:          "AGENT_OUTPUT_V1",
			ComponentName:    "supervisor-handled-component",
			ScopeDescription: "Escalated task resolution",
			Overview:         "This supervisor agent has successfully handled the escalated task with enhanced capabilities.",
			EntryPoints:      []schema.EntryPoint{},
			CallGraph:        []schema.CallGraphEdge{},
			DataFlow:         schema.DataFlow{},
			RawEvidence:      []schema.Evidence{},
		},
	}, nil
}

// MockRetryAgent simulates an agent that fails N times before succeeding
type MockRetryAgent struct {
	MaxFailures    int
	CurrentAttempt int
}

func (m *MockRetryAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	m.CurrentAttempt++

	if m.CurrentAttempt <= m.MaxFailures {
		return schema.AgentResponse{
			Status: schema.StatusError,
			Error: &schema.AgentError{
				Code:        "TEMPORARY_FAILURE",
				Message:     fmt.Sprintf("Attempt %d failed", m.CurrentAttempt),
				Recoverable: true,
			},
		}, fmt.Errorf("temporary failure on attempt %d", m.CurrentAttempt)
	}

	return schema.AgentResponse{
		Status: schema.StatusComplete,
		Output: &schema.AgentOutputV1{
			Version:          "AGENT_OUTPUT_V1",
			ComponentName:    "retry-component",
			ScopeDescription: fmt.Sprintf("Successful retry on attempt %d", m.CurrentAttempt),
			Overview:         fmt.Sprintf("This mock agent demonstrates retry behavior. Failed %d times, succeeded on attempt %d.", m.MaxFailures, m.CurrentAttempt),
			EntryPoints:      []schema.EntryPoint{},
			CallGraph:        []schema.CallGraphEdge{},
			DataFlow:         schema.DataFlow{},
			RawEvidence:      []schema.Evidence{},
		},
	}, nil
}

// MockSlowAgent simulates a slow agent for timeout tests
type MockSlowAgent struct {
	SleepDuration time.Duration
}

func (m *MockSlowAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	select {
	case <-time.After(m.SleepDuration):
		return schema.AgentResponse{
			Status: schema.StatusComplete,
			Output: &schema.AgentOutputV1{
				Version:          "AGENT_OUTPUT_V1",
				ComponentName:    "slow-component",
				ScopeDescription: "Long-running operation",
				Overview:         fmt.Sprintf("This slow operation completed after %v.", m.SleepDuration),
				EntryPoints:      []schema.EntryPoint{},
				CallGraph:        []schema.CallGraphEdge{},
				DataFlow:         schema.DataFlow{},
				RawEvidence:      []schema.Evidence{},
			},
		}, nil
	case <-ctx.Done():
		return schema.AgentResponse{
			Status: schema.StatusError,
			Error: &schema.AgentError{
				Code:    "TIMEOUT",
				Message: "Context cancelled",
			},
		}, ctx.Err()
	}
}

// MockConditionalAgent simulates an agent for conditional branching tests
type MockConditionalAgent struct {
	Name string
}

func (m *MockConditionalAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	return schema.AgentResponse{
		Status: schema.StatusComplete,
		Output: &schema.AgentOutputV1{
			Version:          "AGENT_OUTPUT_V1",
			ComponentName:    fmt.Sprintf("conditional-component-%s", m.Name),
			ScopeDescription: fmt.Sprintf("Conditional execution by %s", m.Name),
			Overview:         fmt.Sprintf("This mock agent executes conditionally based on workflow branching logic. Agent: %s", m.Name),
			EntryPoints:      []schema.EntryPoint{},
			CallGraph:        []schema.CallGraphEdge{},
			DataFlow:         schema.DataFlow{},
			RawEvidence:      []schema.Evidence{},
		},
	}, nil
}

// MockStatePassingAgent simulates an agent that uses data from previous steps
type MockStatePassingAgent struct {
	Name string
}

func (m *MockStatePassingAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	// Parse input to get data from previous step
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(req.Task.SpecificRequest), &input); err != nil {
		return schema.AgentResponse{
			Status: schema.StatusError,
			Error: &schema.AgentError{
				Code:    "INVALID_INPUT",
				Message: "Failed to parse input: " + err.Error(),
			},
		}, err
	}

	// Get component name from previous step (if exists)
	previousComponent := "none"
	if prevComp, ok := input["previous_component"].(string); ok {
		previousComponent = prevComp
	}

	return schema.AgentResponse{
		Status: schema.StatusComplete,
		Output: &schema.AgentOutputV1{
			Version:          "AGENT_OUTPUT_V1",
			ComponentName:    fmt.Sprintf("state-component-%s", m.Name),
			ScopeDescription: fmt.Sprintf("State passing test by %s", m.Name),
			Overview:         fmt.Sprintf("This agent demonstrates state persistence across workflow steps. Previous component: %s", previousComponent),
			EntryPoints:      []schema.EntryPoint{},
			CallGraph:        []schema.CallGraphEdge{},
			DataFlow:         schema.DataFlow{},
			RawEvidence:      []schema.Evidence{},
		},
	}, nil
}

// TestParallelWorkflowExecution tests concurrent execution of multiple agents
func TestParallelWorkflowExecution(t *testing.T) {
	// Setup
	framework := integration.NewTestFramework()

	agent1 := &MockParallelAgent{Name: "agent1", Duration: 100 * time.Millisecond}
	agent2 := &MockParallelAgent{Name: "agent2", Duration: 100 * time.Millisecond}
	agent3 := &MockParallelAgent{Name: "agent3", Duration: 100 * time.Millisecond}

	framework.RegisterAgent("agent1", agent1)
	framework.RegisterAgent("agent2", agent2)
	framework.RegisterAgent("agent3", agent3)

	ctx := context.Background()

	// Execute
	start := time.Now()
	result, err := framework.RunMultiStepWorkflow(ctx, integration.MultiStepWorkflowConfig{
		ID: "parallel-test",
		Steps: []integration.WorkflowStep{
			{ID: "step1", Agent: "agent1", Input: map[string]interface{}{}},
			{ID: "step2", Agent: "agent2", Input: map[string]interface{}{}},
			{ID: "step3", Agent: "agent3", Input: map[string]interface{}{}},
		},
		Mode:    workflow.ParallelMode,
		Timeout: 5 * time.Second,
	})
	duration := time.Since(start)

	// Verify
	require.NoError(t, err, "Parallel workflow should succeed")
	require.NotNil(t, result, "Result should not be nil")
	require.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status)
	require.Len(t, result.WorkflowResult.StepResults, 3)

	// Verify all steps completed
	for i, step := range result.WorkflowResult.StepResults {
		require.Equal(t, workflow.StepStatusCompleted, step.Status, "Step %d should complete", i)
		require.NotNil(t, step.Output, "Step %d should have output", i)
	}

	// Verify parallel execution (should be ~100ms, not ~300ms for sequential)
	require.Less(t, duration, 250*time.Millisecond, "Parallel execution should be faster than sequential")

	t.Logf("✓ Parallel workflow completed in %v (expected < 250ms)", duration)
	t.Logf("✓ All 3 agents executed successfully")
}

// TestWorkflowEscalationHandling tests escalation flow
func TestWorkflowEscalationHandling(t *testing.T) {
	// Setup
	framework := integration.NewTestFramework()

	escalatingAgent := &MockEscalatingAgent{Name: "junior-agent"}
	supervisorAgent := &MockSupervisorAgent{}

	framework.RegisterAgent("junior-agent", escalatingAgent)
	framework.RegisterAgent("supervisor-agent", supervisorAgent)

	ctx := context.Background()

	// Execute
	result, err := framework.RunMultiStepWorkflow(ctx, integration.MultiStepWorkflowConfig{
		ID: "escalation-test",
		Steps: []integration.WorkflowStep{
			{ID: "step1", Agent: "junior-agent", Input: map[string]interface{}{"task": "complex-task"}},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	})

	// Verify
	require.NoError(t, err, "Workflow should complete despite escalation")
	require.NotNil(t, result, "Result should not be nil")

	// Check that escalation was detected and handled
	require.Len(t, result.WorkflowResult.StepResults, 2, "Should have 2 steps: original + escalated")

	// First step should be escalated
	step1 := result.WorkflowResult.StepResults[0]
	require.Equal(t, workflow.StepStatusEscalated, step1.Status, "First step should be escalated")
	require.Equal(t, "supervisor-agent", step1.EscalationTarget, "Should escalate to supervisor")
	require.NotEmpty(t, step1.EscalationReason, "Should have escalation reason")

	// Second step should be the escalated work handled by supervisor
	step2 := result.WorkflowResult.StepResults[1]
	require.Equal(t, "supervisor-agent", step2.Agent, "Second step should be handled by supervisor")
	require.Equal(t, workflow.StepStatusCompleted, step2.Status, "Supervisor step should complete")

	t.Logf("✓ Escalation detected: %s", step1.EscalationReason)
	t.Logf("✓ Target agent: %s", step1.EscalationTarget)
	t.Logf("✓ Escalation handled by: %s", step2.Agent)
}

// TestStatePersistenceAcrossSteps tests data passing between sequential steps
func TestStatePersistenceAcrossSteps(t *testing.T) {
	// Setup
	framework := integration.NewTestFramework()

	agent1 := &MockStatePassingAgent{Name: "agent1"}
	agent2 := &MockStatePassingAgent{Name: "agent2"}
	agent3 := &MockStatePassingAgent{Name: "agent3"}

	framework.RegisterAgent("agent1", agent1)
	framework.RegisterAgent("agent2", agent2)
	framework.RegisterAgent("agent3", agent3)

	ctx := context.Background()

	// Execute - each step should receive output from previous step
	result, err := framework.RunMultiStepWorkflow(ctx, integration.MultiStepWorkflowConfig{
		ID: "state-persistence-test",
		Steps: []integration.WorkflowStep{
			{ID: "step1", Agent: "agent1", Input: map[string]interface{}{"initial": "data"}},
			{ID: "step2", Agent: "agent2", Input: map[string]interface{}{"previous_component": "state-component-agent1"}},
			{ID: "step3", Agent: "agent3", Input: map[string]interface{}{"previous_component": "state-component-agent2"}},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	})

	// Verify
	require.NoError(t, err, "Sequential workflow should succeed")
	require.NotNil(t, result, "Result should not be nil")
	require.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status)
	require.Len(t, result.WorkflowResult.StepResults, 3)

	// Verify state was passed through steps (using ComponentName and Overview, not AgentName/Summary)
	step1 := result.WorkflowResult.StepResults[0]
	require.Equal(t, "state-component-agent1", step1.Output.ComponentName)
	require.Contains(t, step1.Output.Overview, "none", "First step should have no previous component")

	step2 := result.WorkflowResult.StepResults[1]
	require.Equal(t, "state-component-agent2", step2.Output.ComponentName)
	require.Contains(t, step2.Output.Overview, "state-component-agent1", "Second step should reference first")

	step3 := result.WorkflowResult.StepResults[2]
	require.Equal(t, "state-component-agent3", step3.Output.ComponentName)
	require.Contains(t, step3.Output.Overview, "state-component-agent2", "Third step should reference second")

	t.Logf("✓ State passed through 3 sequential steps")
	t.Logf("✓ Step 1: %s", step1.Output.ComponentName)
	t.Logf("✓ Step 2: %s (references %s)", step2.Output.ComponentName, "agent1")
	t.Logf("✓ Step 3: %s (references %s)", step3.Output.ComponentName, "agent2")
}

// TestErrorRecoveryWithRetry tests retry mechanism for transient failures
func TestErrorRecoveryWithRetry(t *testing.T) {
	// Setup
	framework := integration.NewTestFramework()

	retryAgent := &MockRetryAgent{MaxFailures: 2} // Fail twice, then succeed

	framework.RegisterAgent("retry-agent", retryAgent)

	ctx := context.Background()

	// Execute
	result, err := framework.RunMultiStepWorkflow(ctx, integration.MultiStepWorkflowConfig{
		ID: "retry-test",
		Steps: []integration.WorkflowStep{
			{ID: "step1", Agent: "retry-agent", Input: map[string]interface{}{"task": "unstable-operation"}},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 5 * time.Second,
	})

	// Verify
	require.NoError(t, err, "Workflow should eventually succeed after retries")
	require.NotNil(t, result, "Result should not be nil")
	require.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status)
	require.Len(t, result.WorkflowResult.StepResults, 1)

	step := result.WorkflowResult.StepResults[0]
	require.Equal(t, workflow.StepStatusCompleted, step.Status, "Step should eventually complete")
	require.NotNil(t, step.Output, "Step should have output")
	require.Equal(t, "retry-component", step.Output.ComponentName)
	require.Contains(t, step.Output.Overview, "attempt 3", "Should succeed on third attempt")

	t.Logf("✓ Agent succeeded after 2 failures")
	t.Logf("✓ Output: %s", step.Output.Overview)
}

// TestConditionalWorkflowBranching tests conditional execution paths
func TestConditionalWorkflowBranching(t *testing.T) {
	// Setup
	framework := integration.NewTestFramework()

	agent1 := &MockConditionalAgent{Name: "agent1"}
	agent3 := &MockConditionalAgent{Name: "agent3"}

	framework.RegisterAgent("agent1", agent1)
	framework.RegisterAgent("agent3", agent3)

	ctx := context.Background()

	// Execute - step2 should only run if step1 succeeds
	result, err := framework.RunMultiStepWorkflow(ctx, integration.MultiStepWorkflowConfig{
		ID: "conditional-test",
		Steps: []integration.WorkflowStep{
			{ID: "step1", Agent: "agent1", Input: map[string]interface{}{}},
			{
				ID:        "step2",
				Agent:     "agent3",
				Input:     map[string]interface{}{"previous_step": "step1"},
				Condition: &workflow.PreviousStepSuccessCondition{StepID: "step1"},
			},
		},
		Mode:    workflow.ConditionalMode,
		Timeout: 5 * time.Second,
	})

	// Verify
	require.NoError(t, err, "Conditional workflow should succeed")
	require.NotNil(t, result, "Result should not be nil")
	require.Equal(t, workflow.StatusCompleted, result.WorkflowResult.Status)
	require.Len(t, result.WorkflowResult.StepResults, 2, "Both steps should execute (step1 succeeds)")

	step1 := result.WorkflowResult.StepResults[0]
	require.Equal(t, workflow.StepStatusCompleted, step1.Status)
	require.Equal(t, "conditional-component-agent1", step1.Output.ComponentName)

	step2 := result.WorkflowResult.StepResults[1]
	require.Equal(t, workflow.StepStatusCompleted, step2.Status)
	require.Equal(t, "conditional-component-agent3", step2.Output.ComponentName)

	t.Logf("✓ Conditional branching worked correctly")
	t.Logf("✓ Step 1 succeeded, Step 2 executed")
}

// TestWorkflowTimeoutHandling tests timeout enforcement
func TestWorkflowTimeoutHandling(t *testing.T) {
	// Setup
	framework := integration.NewTestFramework()

	slowAgent := &MockSlowAgent{SleepDuration: 10 * time.Second}

	framework.RegisterAgent("slow-agent", slowAgent)

	ctx := context.Background()

	// Execute with short timeout
	start := time.Now()
	_, err := framework.RunMultiStepWorkflow(ctx, integration.MultiStepWorkflowConfig{
		ID: "timeout-test",
		Steps: []integration.WorkflowStep{
			{ID: "step1", Agent: "slow-agent", Input: map[string]interface{}{}},
		},
		Mode:    workflow.SequentialMode,
		Timeout: 1 * time.Second, // Should timeout before 10s sleep completes
	})
	duration := time.Since(start)

	// Verify timeout occurred
	require.Error(t, err, "Workflow should timeout")
	require.True(t, strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline"), "Error should mention timeout or deadline")
	require.Less(t, duration, 2*time.Second, "Should timeout quickly, not wait for full 10s")

	// When timeout occurs, result may be nil - that's expected behavior
	// No need to check result.Status since error is returned

	t.Logf("✓ Timeout enforced after %v (expected ~1s)", duration)
}
