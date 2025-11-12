// Package evidence provides validation for evidence-backed claims in agent outputs.
//
// The evidence validator enforces the requirement that 100% of claims in
// AGENT_OUTPUT_V1 outputs must be backed by verifiable file:line evidence.
package evidence

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// Validator validates evidence backing for agent outputs
type Validator struct {
	strictMode bool
}

// NewValidator creates a new evidence validator
func NewValidator(strictMode bool) *Validator {
	return &Validator{
		strictMode: strictMode,
	}
}

// ValidationResult contains the results of evidence validation
type ValidationResult struct {
	Valid              bool
	TotalClaims        int
	BackedClaims       int
	UnbackedClaims     []UnbackedClaim
	InvalidEvidence    []InvalidEvidence
	CoveragePercentage float64
}

// UnbackedClaim represents a claim without evidence
type UnbackedClaim struct {
	Section     string
	Description string
	Index       int
}

// InvalidEvidence represents evidence that failed validation
type InvalidEvidence struct {
	File   string
	Lines  string
	Reason string
}

// Validate checks if an agent output has 100% evidence backing
func (v *Validator) Validate(output *schema.AgentOutputV1) (*ValidationResult, error) {
	if output == nil {
		return nil, fmt.Errorf("output is nil")
	}

	result := &ValidationResult{
		Valid:           true,
		UnbackedClaims:  make([]UnbackedClaim, 0),
		InvalidEvidence: make([]InvalidEvidence, 0),
	}

	// Build evidence index
	evidenceIndex := v.buildEvidenceIndex(output.RawEvidence)

	// Validate all sections
	v.validateEntryPoints(output.EntryPoints, evidenceIndex, result)
	v.validateCallGraph(output.CallGraph, evidenceIndex, result)
	v.validateDataFlow(output.DataFlow, evidenceIndex, result)
	v.validateStateManagement(output.StateManagement, evidenceIndex, result)
	v.validateSideEffects(output.SideEffects, evidenceIndex, result)
	v.validateErrorHandling(output.ErrorHandling, evidenceIndex, result)
	v.validateConfiguration(output.Configuration, evidenceIndex, result)
	v.validatePatterns(output.Patterns, evidenceIndex, result)
	v.validateConcurrency(output.Concurrency, evidenceIndex, result)

	// Validate evidence references exist
	if v.strictMode {
		v.validateEvidenceFiles(output.RawEvidence, result)
	}

	// Calculate coverage
	if result.TotalClaims > 0 {
		result.CoveragePercentage = float64(result.BackedClaims) / float64(result.TotalClaims) * 100
	} else {
		result.CoveragePercentage = 100.0 // No claims = 100% coverage
	}

	// Determine validity
	if len(result.UnbackedClaims) > 0 || len(result.InvalidEvidence) > 0 {
		result.Valid = false
	}

	return result, nil
}

// buildEvidenceIndex creates a map of file:lines to evidence entries
func (v *Validator) buildEvidenceIndex(evidence []schema.Evidence) map[string]bool {
	index := make(map[string]bool)
	for _, e := range evidence {
		key := fmt.Sprintf("%s:%s", e.File, e.Lines)
		index[key] = true
	}
	return index
}

// validateEntryPoints checks entry point evidence
func (v *Validator) validateEntryPoints(entryPoints []schema.EntryPoint, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, ep := range entryPoints {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%s", ep.File, ep.Lines)
		if evidenceIndex[key] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "entry_points",
				Description: fmt.Sprintf("Entry point %s at %s:%s", ep.Symbol, ep.File, ep.Lines),
				Index:       i,
			})
		}
	}
}

// validateCallGraph checks call graph evidence
func (v *Validator) validateCallGraph(callGraph []schema.CallGraphEdge, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, edge := range callGraph {
		result.TotalClaims++

		// Extract file and line from "file.go:funcA" format
		fromParts := strings.Split(edge.From, ":")
		if len(fromParts) >= 1 {
			if v.lineHasEvidence(fromParts[0], edge.ViaLine, evidenceIndex) {
				result.BackedClaims++
			} else {
				result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
					Section:     "call_graph",
					Description: fmt.Sprintf("Call from %s to %s at line %d", edge.From, edge.To, edge.ViaLine),
					Index:       i,
				})
			}
		}
	}
}

// validateDataFlow checks data flow evidence
func (v *Validator) validateDataFlow(dataFlow schema.DataFlow, evidenceIndex map[string]bool, result *ValidationResult) {
	// Validate inputs
	for i, input := range dataFlow.Inputs {
		result.TotalClaims++
		if v.hasEvidence(input.Source, evidenceIndex) {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "data_flow.inputs",
				Description: fmt.Sprintf("Input %s at %s", input.Name, input.Source),
				Index:       i,
			})
		}
	}

	// Validate transformations
	for i, transform := range dataFlow.Transformations {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%s", transform.File, transform.Lines)
		if evidenceIndex[key] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "data_flow.transformations",
				Description: fmt.Sprintf("Transformation %s at %s:%s", transform.Operation, transform.File, transform.Lines),
				Index:       i,
			})
		}
	}

	// Validate outputs
	for i, output := range dataFlow.Outputs {
		result.TotalClaims++
		if v.hasEvidence(output.Source, evidenceIndex) {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "data_flow.outputs",
				Description: fmt.Sprintf("Output %s at %s", output.Name, output.Source),
				Index:       i,
			})
		}
	}
}

// validateStateManagement checks state operation evidence
func (v *Validator) validateStateManagement(stateOps []schema.StateOperation, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, op := range stateOps {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%s", op.File, op.Lines)
		if evidenceIndex[key] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "state_management",
				Description: fmt.Sprintf("State operation %s on %s at %s:%s", op.Operation, op.Entity, op.File, op.Lines),
				Index:       i,
			})
		}
	}
}

// validateSideEffects checks side effect evidence
func (v *Validator) validateSideEffects(sideEffects []schema.SideEffect, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, effect := range sideEffects {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%d", effect.File, effect.Line)
		keyRange := fmt.Sprintf("%s:%d-%d", effect.File, effect.Line, effect.Line)
		if evidenceIndex[key] || evidenceIndex[keyRange] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "side_effects",
				Description: fmt.Sprintf("Side effect %s at %s:%d", effect.Type, effect.File, effect.Line),
				Index:       i,
			})
		}
	}
}

// validateErrorHandling checks error handler evidence
func (v *Validator) validateErrorHandling(errorHandlers []schema.ErrorHandler, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, handler := range errorHandlers {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%s", handler.File, handler.Lines)
		if evidenceIndex[key] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "error_handling",
				Description: fmt.Sprintf("Error handler %s at %s:%s", handler.Type, handler.File, handler.Lines),
				Index:       i,
			})
		}
	}
}

// validateConfiguration checks configuration influence evidence
func (v *Validator) validateConfiguration(configs []schema.ConfigInfluence, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, config := range configs {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%d", config.File, config.Line)
		keyRange := fmt.Sprintf("%s:%d-%d", config.File, config.Line, config.Line)
		if evidenceIndex[key] || evidenceIndex[keyRange] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "configuration",
				Description: fmt.Sprintf("Config %s at %s:%d", config.Name, config.File, config.Line),
				Index:       i,
			})
		}
	}
}

// validatePatterns checks pattern evidence
func (v *Validator) validatePatterns(patterns []schema.Pattern, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, pattern := range patterns {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%s", pattern.File, pattern.Lines)
		if evidenceIndex[key] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "patterns",
				Description: fmt.Sprintf("Pattern %s at %s:%s", pattern.Name, pattern.File, pattern.Lines),
				Index:       i,
			})
		}
	}
}

// validateConcurrency checks concurrency mechanism evidence
func (v *Validator) validateConcurrency(concurrency []schema.ConcurrencyMechanism, evidenceIndex map[string]bool, result *ValidationResult) {
	for i, mech := range concurrency {
		result.TotalClaims++
		key := fmt.Sprintf("%s:%s", mech.File, mech.Lines)
		if evidenceIndex[key] {
			result.BackedClaims++
		} else {
			result.UnbackedClaims = append(result.UnbackedClaims, UnbackedClaim{
				Section:     "concurrency",
				Description: fmt.Sprintf("Concurrency %s at %s:%s", mech.Mechanism, mech.File, mech.Lines),
				Index:       i,
			})
		}
	}
}

// hasEvidence checks if a source reference has evidence
func (v *Validator) hasEvidence(source string, evidenceIndex map[string]bool) bool {
	// Source format: "file.go:line" or "file.go:line-line"
	return evidenceIndex[source]
}

// validateEvidenceFiles checks if evidence files actually exist and lines are valid
func (v *Validator) validateEvidenceFiles(evidence []schema.Evidence, result *ValidationResult) {
	for _, e := range evidence {
		// Check if file exists
		if _, err := os.Stat(e.File); os.IsNotExist(err) {
			result.InvalidEvidence = append(result.InvalidEvidence, InvalidEvidence{
				File:   e.File,
				Lines:  e.Lines,
				Reason: "file does not exist",
			})
			continue
		}

		// Validate line range format
		if err := v.validateLineRange(e.Lines); err != nil {
			result.InvalidEvidence = append(result.InvalidEvidence, InvalidEvidence{
				File:   e.File,
				Lines:  e.Lines,
				Reason: err.Error(),
			})
		}
	}
}

// validateLineRange validates the line range format
func (v *Validator) validateLineRange(lines string) error {
	if lines == "" {
		return fmt.Errorf("empty line range")
	}

	// Check for single line or range
	if strings.Contains(lines, "-") {
		parts := strings.Split(lines, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid line range format: %s", lines)
		}

		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid start line: %s", parts[0])
		}

		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("invalid end line: %s", parts[1])
		}

		if start <= 0 || end <= 0 {
			return fmt.Errorf("line numbers must be positive")
		}

		if start > end {
			return fmt.Errorf("start line cannot be greater than end line")
		}
	} else {
		line, err := strconv.Atoi(lines)
		if err != nil {
			return fmt.Errorf("invalid line number: %s", lines)
		}

		if line <= 0 {
			return fmt.Errorf("line number must be positive")
		}
	}

	return nil
}

// lineHasEvidence checks if a specific line has evidence, including checking if the line
// falls within any range evidence entries.
func (v *Validator) lineHasEvidence(file string, line int, evidenceIndex map[string]bool) bool {
	// Check for exact line match
	exactKey := fmt.Sprintf("%s:%d", file, line)
	if evidenceIndex[exactKey] {
		return true
	}

	// Check for single-line range match (e.g., file.go:20-20 for line 20)
	singleRangeKey := fmt.Sprintf("%s:%d-%d", file, line, line)
	if evidenceIndex[singleRangeKey] {
		return true
	}

	// Check if line falls within any range evidence
	// We need to iterate through the evidence index to find range entries
	for key := range evidenceIndex {
		// Parse the key format: "file:lines"
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			continue
		}

		// Check if this evidence is for the same file
		if parts[0] != file {
			continue
		}

		// Check if this is a range (contains "-")
		linesPart := parts[1]
		if !strings.Contains(linesPart, "-") {
			continue
		}

		// Parse the range
		rangeParts := strings.Split(linesPart, "-")
		if len(rangeParts) != 2 {
			continue
		}

		start, err1 := strconv.Atoi(rangeParts[0])
		end, err2 := strconv.Atoi(rangeParts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		// Check if our line falls within this range
		if line >= start && line <= end {
			return true
		}
	}

	return false
}
