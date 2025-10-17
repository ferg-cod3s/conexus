// Package integration provides integration testing framework for multi-agent workflows.
package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ferg-cod3s/conexus/internal/security"
	"github.com/ferg-cod3s/conexus/internal/validation/evidence"
	schemaval "github.com/ferg-cod3s/conexus/internal/validation/schema"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// ValidationReport contains comprehensive validation results
type ValidationReport struct {
	SchemaValid    bool
	EvidenceValid  bool
	SchemaResult   *schemaval.ValidationResult
	EvidenceResult *evidence.ValidationResult
	Errors         []string
	Warnings       []string
}

// PerformanceMetrics contains performance measurements
type PerformanceMetrics struct {
	Duration       time.Duration
	StartTime      time.Time
	EndTime        time.Time
	MemoryUsedMB   float64
	AgentCount     int
	StepCount      int
	OutputSizeKB   float64
}

// TestCodebase represents a test fixture codebase
type TestCodebase struct {
	Name        string
	Path        string
	Files       map[string]string // filename -> content
	Description string
}

// VerifyEvidence validates evidence backing for an agent output
func VerifyEvidence(output *schema.AgentOutputV1, strictMode bool) (*ValidationReport, error) {
	if output == nil {
		return nil, fmt.Errorf("output is nil")
	}

	report := &ValidationReport{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Validate schema
	schemaValidator := schemaval.NewValidator(false) // Don't allow partial
	schemaResult, err := schemaValidator.Validate(output)
	if err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}
	report.SchemaResult = schemaResult
	report.SchemaValid = schemaResult.Valid

	// Collect schema errors
	for _, e := range schemaResult.Errors {
		report.Errors = append(report.Errors, fmt.Sprintf("Schema Error [%s]: %s", e.Field, e.Message))
	}
	for _, w := range schemaResult.Warnings {
		report.Warnings = append(report.Warnings, fmt.Sprintf("Schema Warning [%s]: %s", w.Field, w.Message))
	}

	// Validate evidence backing
	evidenceValidator := evidence.NewValidator(strictMode)
	evidenceResult, err := evidenceValidator.Validate(output)
	if err != nil {
		return nil, fmt.Errorf("evidence validation failed: %w", err)
	}
	report.EvidenceResult = evidenceResult
	report.EvidenceValid = evidenceResult.Valid

	// Collect evidence errors
	for _, claim := range evidenceResult.UnbackedClaims {
		report.Errors = append(report.Errors, 
			fmt.Sprintf("Unbacked Claim [%s]: %s", claim.Section, claim.Description))
	}
	for _, inv := range evidenceResult.InvalidEvidence {
		report.Errors = append(report.Errors, 
			fmt.Sprintf("Invalid Evidence [%s:%s]: %s", inv.File, inv.Lines, inv.Reason))
	}

	return report, nil
}

// VerifySchema validates schema compliance for an agent output
func VerifySchema(output *schema.AgentOutputV1) error {
	if output == nil {
		return fmt.Errorf("output is nil")
	}

	validator := schemaval.NewValidator(false)
	result, err := validator.Validate(output)
	if err != nil {
		return err
	}

	if !result.Valid {
		return fmt.Errorf("schema validation failed: %d errors", len(result.Errors))
	}

	return nil
}

// MeasurePerformance measures the performance of a function
func MeasurePerformance(fn func()) *PerformanceMetrics {
	metrics := &PerformanceMetrics{
		StartTime: time.Now(),
	}

	// Execute function
	fn()

	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	return metrics
}

// LoadTestFixture loads a test codebase from the fixtures directory
func LoadTestFixture(name string) (*TestCodebase, error) {
	// Find the tests/fixtures directory
	fixturesDir := filepath.Join("tests", "fixtures")
	
	// Check if fixtures directory exists
	if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("fixtures directory not found: %s", fixturesDir)
	}

	// Load all .go files in the fixture
	fixture := &TestCodebase{
		Name:  name,
		Path:  fixturesDir,
		Files: make(map[string]string),
	}

	// Read fixture files
	files, err := os.ReadDir(fixturesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixtures: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		// Only load .go files
		if filepath.Ext(filename) != ".go" {
			continue
		}

		// Construct full path and validate to prevent directory traversal (G304)
		fullPath := filepath.Join(fixturesDir, filename)
		if _, err := security.ValidatePathWithinBase(fullPath, fixturesDir); err != nil {
			return nil, fmt.Errorf("invalid fixture path %s: %w", filename, err)
		}

		// #nosec G304 -- Path validated at line 166 with ValidatePathWithinBase
		// Read file content
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read fixture file %s: %w", filename, err)
		}

		fixture.Files[filename] = string(content)
	}

	if len(fixture.Files) == 0 {
		return nil, fmt.Errorf("no fixture files found in %s", fixturesDir)
	}

	return fixture, nil
}

// CreateTempCodebase creates a temporary directory with test code
func CreateTempCodebase(files map[string]string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "conexus-test-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	for filename, content := range files {
		fullPath := filepath.Join(tmpDir, filename)
		
		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
			// #nosec G104 - Best-effort cleanup in error path, primary error already captured
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			// #nosec G104 - Best-effort cleanup in error path, primary error already captured
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to write file %s: %w", filename, err)
		}
	}

	return tmpDir, nil
}

// CleanupTempCodebase removes a temporary codebase directory
func CleanupTempCodebase(path string) error {
	return os.RemoveAll(path)
}

// ParseAgentOutput parses AGENT_OUTPUT_V1 from JSON bytes
func ParseAgentOutput(data []byte) (*schema.AgentOutputV1, error) {
	var output schema.AgentOutputV1
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("failed to parse agent output: %w", err)
	}
	return &output, nil
}

// ValidateAgentOutput performs full validation on an agent output
func ValidateAgentOutput(output *schema.AgentOutputV1, strictMode bool) error {
	report, err := VerifyEvidence(output, strictMode)
	if err != nil {
		return err
	}

	if !report.SchemaValid {
		return fmt.Errorf("schema validation failed with %d errors", len(report.SchemaResult.Errors))
	}

	if !report.EvidenceValid {
		return fmt.Errorf("evidence validation failed: %d unbacked claims, %.1f%% coverage",
			len(report.EvidenceResult.UnbackedClaims),
			report.EvidenceResult.CoveragePercentage)
	}

	return nil
}

// CreateSimpleTestFile creates a simple Go test file
func CreateSimpleTestFile(name string) string {
	return fmt.Sprintf(`package testcode

// %s is a simple test function
func %s(x, y int) int {
	return x + y
}
`, name, name)
}

// AssertValidOutput is a helper assertion for valid AGENT_OUTPUT_V1
func AssertValidOutput(output *schema.AgentOutputV1) error {
	if output == nil {
		return fmt.Errorf("output is nil")
	}

	if output.Version != "AGENT_OUTPUT_V1" {
		return fmt.Errorf("invalid version: expected AGENT_OUTPUT_V1, got %s", output.Version)
	}

	if output.ComponentName == "" {
		return fmt.Errorf("component_name is empty")
	}

	if len(output.RawEvidence) == 0 {
		return fmt.Errorf("no evidence provided")
	}

	return nil
}

// CountTotalClaims counts all claims in an agent output
func CountTotalClaims(output *schema.AgentOutputV1) int {
	total := 0
	total += len(output.EntryPoints)
	total += len(output.CallGraph)
	total += len(output.DataFlow.Inputs)
	total += len(output.DataFlow.Transformations)
	total += len(output.DataFlow.Outputs)
	total += len(output.StateManagement)
	total += len(output.SideEffects)
	total += len(output.ErrorHandling)
	total += len(output.Configuration)
	total += len(output.Patterns)
	total += len(output.Concurrency)
	return total
}

// CalculateOutputSize calculates the size of an agent output in KB
func CalculateOutputSize(output *schema.AgentOutputV1) (float64, error) {
	data, err := json.Marshal(output)
	if err != nil {
		return 0, err
	}
	return float64(len(data)) / 1024.0, nil
}
