// Package schema provides JSON schema validation for AGENT_OUTPUT_V1 structures.
//
// The schema validator ensures that agent outputs conform to the
// AGENT_OUTPUT_V1 specification with correct structure and field types.
package schema

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Validator validates AGENT_OUTPUT_V1 schema compliance
type Validator struct {
	allowPartial bool
}

// NewValidator creates a new schema validator
func NewValidator(allowPartial bool) *Validator {
	return &Validator{
		allowPartial: allowPartial,
	}
}

// ValidationResult contains schema validation results
type ValidationResult struct {
	Valid         bool
	Errors        []ValidationError
	Warnings      []ValidationWarning
	MissingFields []string
	InvalidFields []InvalidField
}

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// ValidationWarning represents a schema validation warning
type ValidationWarning struct {
	Field   string
	Message string
}

// InvalidField represents a field with invalid type or value
type InvalidField struct {
	Field        string
	ExpectedType string
	ActualType   string
	Value        interface{}
}

// Validate checks if output conforms to AGENT_OUTPUT_V1 schema
func (v *Validator) Validate(output *schema.AgentOutputV1) (*ValidationResult, error) {
	if output == nil {
		return nil, fmt.Errorf("output is nil")
	}

	result := &ValidationResult{
		Valid:         true,
		Errors:        make([]ValidationError, 0),
		Warnings:      make([]ValidationWarning, 0),
		MissingFields: make([]string, 0),
		InvalidFields: make([]InvalidField, 0),
	}

	// Validate required fields
	v.validateVersion(output, result)
	v.validateComponentName(output, result)
	v.validateScopeDescription(output, result)
	v.validateOverview(output, result)
	v.validateRawEvidence(output, result)

	// Validate optional but important fields
	v.validateEntryPoints(output.EntryPoints, result)
	v.validateCallGraph(output.CallGraph, result)
	v.validateDataFlow(output.DataFlow, result)
	v.validateStateManagement(output.StateManagement, result)
	v.validateSideEffects(output.SideEffects, result)
	v.validateErrorHandling(output.ErrorHandling, result)
	v.validateConfiguration(output.Configuration, result)
	v.validatePatterns(output.Patterns, result)
	v.validateConcurrency(output.Concurrency, result)
	v.validateExternalDependencies(output.ExternalDependencies, result)

	// Determine overall validity
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	if !v.allowPartial && len(result.MissingFields) > 0 {
		result.Valid = false
	}

	return result, nil
}

// validateVersion checks version field
func (v *Validator) validateVersion(output *schema.AgentOutputV1, result *ValidationResult) {
	if output.Version == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "version",
			Message: "version is required",
		})
	} else if output.Version != "AGENT_OUTPUT_V1" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("version must be 'AGENT_OUTPUT_V1', got '%s'", output.Version),
			Value:   output.Version,
		})
	}
}

// validateComponentName checks component_name field
func (v *Validator) validateComponentName(output *schema.AgentOutputV1, result *ValidationResult) {
	if output.ComponentName == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "component_name",
			Message: "component_name is required",
		})
	}
}

// validateScopeDescription checks scope_description field
func (v *Validator) validateScopeDescription(output *schema.AgentOutputV1, result *ValidationResult) {
	if output.ScopeDescription == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "scope_description",
			Message: "scope_description is empty",
		})
	}
}

// validateOverview checks overview field
func (v *Validator) validateOverview(output *schema.AgentOutputV1, result *ValidationResult) {
	if output.Overview == "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "overview",
			Message: "overview is empty",
		})
	} else if len(output.Overview) < 20 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "overview",
			Message: "overview is very short (should be 2-4 sentences)",
		})
	}
}

// validateRawEvidence checks raw_evidence field
func (v *Validator) validateRawEvidence(output *schema.AgentOutputV1, result *ValidationResult) {
	if len(output.RawEvidence) == 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "raw_evidence",
			Message: "no evidence provided (100% evidence backing required)",
		})
	}

	for i, evidence := range output.RawEvidence {
		if evidence.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("raw_evidence[%d].file", i),
				Message: "evidence file is required",
			})
		}

		if evidence.Lines == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("raw_evidence[%d].lines", i),
				Message: "evidence lines are required",
			})
		}

		if evidence.File != "" && !v.isAbsolutePath(evidence.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("raw_evidence[%d].file", i),
				Message: "evidence file must be absolute path",
				Value:   evidence.File,
			})
		}
	}
}

// validateEntryPoints checks entry_points structure
func (v *Validator) validateEntryPoints(entryPoints []schema.EntryPoint, result *ValidationResult) {
	for i, ep := range entryPoints {
		if ep.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("entry_points[%d].file", i),
				Message: "file is required",
			})
		}

		if ep.Lines == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("entry_points[%d].lines", i),
				Message: "lines are required",
			})
		}

		if ep.Symbol == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("entry_points[%d].symbol", i),
				Message: "symbol is required",
			})
		}

		if !v.isAbsolutePath(ep.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("entry_points[%d].file", i),
				Message: "file must be absolute path",
				Value:   ep.File,
			})
		}
	}
}

// validateCallGraph checks call_graph structure
func (v *Validator) validateCallGraph(callGraph []schema.CallGraphEdge, result *ValidationResult) {
	for i, edge := range callGraph {
		if edge.From == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("call_graph[%d].from", i),
				Message: "from is required",
			})
		}

		if edge.To == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("call_graph[%d].to", i),
				Message: "to is required",
			})
		}

		if edge.ViaLine <= 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("call_graph[%d].via_line", i),
				Message: "via_line must be positive",
				Value:   edge.ViaLine,
			})
		}
	}
}

// validateDataFlow checks data_flow structure
func (v *Validator) validateDataFlow(dataFlow schema.DataFlow, result *ValidationResult) {
	// Validate inputs
	for i, input := range dataFlow.Inputs {
		if input.Source == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("data_flow.inputs[%d].source", i),
				Message: "source is required",
			})
		}

		if input.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("data_flow.inputs[%d].name", i),
				Message: "name is required",
			})
		}
	}

	// Validate transformations
	for i, transform := range dataFlow.Transformations {
		if transform.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("data_flow.transformations[%d].file", i),
				Message: "file is required",
			})
		}

		if transform.Lines == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("data_flow.transformations[%d].lines", i),
				Message: "lines are required",
			})
		}

		if !v.isAbsolutePath(transform.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("data_flow.transformations[%d].file", i),
				Message: "file must be absolute path",
				Value:   transform.File,
			})
		}
	}

	// Validate outputs
	for i, output := range dataFlow.Outputs {
		if output.Source == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("data_flow.outputs[%d].source", i),
				Message: "source is required",
			})
		}

		if output.Name == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("data_flow.outputs[%d].name", i),
				Message: "name is required",
			})
		}
	}
}

// validateStateManagement checks state_management structure
func (v *Validator) validateStateManagement(stateOps []schema.StateOperation, result *ValidationResult) {
	for i, op := range stateOps {
		if op.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("state_management[%d].file", i),
				Message: "file is required",
			})
		}

		if op.Lines == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("state_management[%d].lines", i),
				Message: "lines are required",
			})
		}

		if !v.isAbsolutePath(op.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("state_management[%d].file", i),
				Message: "file must be absolute path",
				Value:   op.File,
			})
		}
	}
}

// validateSideEffects checks side_effects structure
func (v *Validator) validateSideEffects(sideEffects []schema.SideEffect, result *ValidationResult) {
	for i, effect := range sideEffects {
		if effect.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("side_effects[%d].file", i),
				Message: "file is required",
			})
		}

		if effect.Line <= 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("side_effects[%d].line", i),
				Message: "line must be positive",
				Value:   effect.Line,
			})
		}

		if effect.File != "" && !v.isAbsolutePath(effect.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("side_effects[%d].file", i),
				Message: "file must be absolute path",
				Value:   effect.File,
			})
		}
	}
}

// validateErrorHandling checks error_handling structure
func (v *Validator) validateErrorHandling(errorHandlers []schema.ErrorHandler, result *ValidationResult) {
	for i, handler := range errorHandlers {
		if handler.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("error_handling[%d].file", i),
				Message: "file is required",
			})
		}

		if handler.Lines == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("error_handling[%d].lines", i),
				Message: "lines are required",
			})
		}

		if !v.isAbsolutePath(handler.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("error_handling[%d].file", i),
				Message: "file must be absolute path",
				Value:   handler.File,
			})
		}
	}
}

// validateConfiguration checks configuration structure
func (v *Validator) validateConfiguration(configs []schema.ConfigInfluence, result *ValidationResult) {
	for i, config := range configs {
		if config.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("configuration[%d].file", i),
				Message: "file is required",
			})
		}

		if config.Line <= 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("configuration[%d].line", i),
				Message: "line must be positive",
				Value:   config.Line,
			})
		}

		if !v.isAbsolutePath(config.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("configuration[%d].file", i),
				Message: "file must be absolute path",
				Value:   config.File,
			})
		}
	}
}

// validatePatterns checks patterns structure
func (v *Validator) validatePatterns(patterns []schema.Pattern, result *ValidationResult) {
	for i, pattern := range patterns {
		if pattern.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("patterns[%d].file", i),
				Message: "file is required",
			})
		}

		if pattern.Lines == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("patterns[%d].lines", i),
				Message: "lines are required",
			})
		}

		if !v.isAbsolutePath(pattern.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("patterns[%d].file", i),
				Message: "file must be absolute path",
				Value:   pattern.File,
			})
		}
	}
}

// validateConcurrency checks concurrency structure
func (v *Validator) validateConcurrency(concurrency []schema.ConcurrencyMechanism, result *ValidationResult) {
	for i, mech := range concurrency {
		if mech.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("concurrency[%d].file", i),
				Message: "file is required",
			})
		}

		if mech.Lines == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("concurrency[%d].lines", i),
				Message: "lines are required",
			})
		}

		if !v.isAbsolutePath(mech.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("concurrency[%d].file", i),
				Message: "file must be absolute path",
				Value:   mech.File,
			})
		}
	}
}

// validateExternalDependencies checks external_dependencies structure
func (v *Validator) validateExternalDependencies(deps []schema.ExternalDependency, result *ValidationResult) {
	for i, dep := range deps {
		if dep.File == "" {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("external_dependencies[%d].file", i),
				Message: "file is required",
			})
		}

		if dep.Line <= 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("external_dependencies[%d].line", i),
				Message: "line must be positive",
				Value:   dep.Line,
			})
		}

		if !v.isAbsolutePath(dep.File) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("external_dependencies[%d].file", i),
				Message: "file must be absolute path",
				Value:   dep.File,
			})
		}
	}
}

// isAbsolutePath checks if a path is absolute
func (v *Validator) isAbsolutePath(path string) bool {
	return strings.HasPrefix(path, "/") || strings.HasPrefix(path, "C:") || strings.HasPrefix(path, "c:")
}

// GetFieldType returns the type of a field value
func GetFieldType(value interface{}) string {
	if value == nil {
		return "nil"
	}
	return reflect.TypeOf(value).String()
}
