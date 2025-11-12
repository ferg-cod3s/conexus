package orchestrator

import (
	"fmt"
	"time"
)

// QualityGateConfig defines quality thresholds for workflow execution
type QualityGateConfig struct {
	// Validation thresholds
	RequireEvidenceBacking bool    // Block on validation failures
	MinEvidenceCoverage    float64 // Minimum evidence coverage percentage (0-100)
	AllowUnbackedClaims    int     // Maximum unbacked claims allowed

	// Performance thresholds
	MaxExecutionTime          time.Duration // Maximum total workflow time
	MaxAgentExecutionTime     time.Duration // Maximum single agent execution time
	MaxMemoryUsage            uint64        // Maximum memory usage in bytes
	PerformanceOverheadTarget float64       // Maximum profiling overhead percentage

	// Error handling
	BlockOnValidationFailure  bool // Stop workflow on validation failure
	BlockOnPerformanceFailure bool // Stop workflow on performance threshold breach
}

// DefaultQualityGates returns sensible default quality gate configuration
func DefaultQualityGates() *QualityGateConfig {
	return &QualityGateConfig{
		RequireEvidenceBacking:    true,
		MinEvidenceCoverage:       100.0,
		AllowUnbackedClaims:       0,
		MaxExecutionTime:          5 * time.Minute,
		MaxAgentExecutionTime:     1 * time.Minute,
		MaxMemoryUsage:            500 * 1024 * 1024, // 500MB
		PerformanceOverheadTarget: 10.0,              // <10% overhead
		BlockOnValidationFailure:  true,
		BlockOnPerformanceFailure: false,
	}
}

// RelaxedQualityGates returns more permissive thresholds for development
func RelaxedQualityGates() *QualityGateConfig {
	return &QualityGateConfig{
		RequireEvidenceBacking:    false,
		MinEvidenceCoverage:       80.0,
		AllowUnbackedClaims:       5,
		MaxExecutionTime:          10 * time.Minute,
		MaxAgentExecutionTime:     2 * time.Minute,
		MaxMemoryUsage:            1024 * 1024 * 1024, // 1GB
		PerformanceOverheadTarget: 20.0,
		BlockOnValidationFailure:  false,
		BlockOnPerformanceFailure: false,
	}
}

// StrictQualityGates returns strict thresholds for production
func StrictQualityGates() *QualityGateConfig {
	return &QualityGateConfig{
		RequireEvidenceBacking:    true,
		MinEvidenceCoverage:       100.0,
		AllowUnbackedClaims:       0,
		MaxExecutionTime:          2 * time.Minute,
		MaxAgentExecutionTime:     30 * time.Second,
		MaxMemoryUsage:            256 * 1024 * 1024, // 256MB
		PerformanceOverheadTarget: 5.0,
		BlockOnValidationFailure:  true,
		BlockOnPerformanceFailure: true,
	}
}

// QualityGateResult contains the results of quality gate checks
type QualityGateResult struct {
	Passed            bool
	ValidationPassed  bool
	PerformancePassed bool
	Violations        []Violation
}

// Violation represents a quality gate threshold breach
type Violation struct {
	Type        ViolationType
	Severity    Severity
	Description string
	Actual      interface{}
	Expected    interface{}
}

// ViolationType categorizes quality gate violations
type ViolationType string

const (
	ViolationEvidenceCoverage  ViolationType = "evidence_coverage"
	ViolationUnbackedClaims    ViolationType = "unbacked_claims"
	ViolationInvalidEvidence   ViolationType = "invalid_evidence"
	ViolationExecutionTime     ViolationType = "execution_time"
	ViolationAgentTime         ViolationType = "agent_execution_time"
	ViolationMemoryUsage       ViolationType = "memory_usage"
	ViolationProfilingOverhead ViolationType = "profiling_overhead"
)

// Severity indicates the importance of a violation
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// CheckQualityGates evaluates whether execution meets quality thresholds
func (qg *QualityGateConfig) CheckQualityGates(
	validationResult *ValidationReport,
	profilingResult *ProfilingReport,
) *QualityGateResult {
	result := &QualityGateResult{
		Passed:            true,
		ValidationPassed:  true,
		PerformancePassed: true,
		Violations:        []Violation{},
	}

	// Check validation gates
	if qg.RequireEvidenceBacking && validationResult != nil {
		result.ValidationPassed = qg.checkValidationGates(validationResult, result)
	}

	// Check performance gates
	if profilingResult != nil {
		result.PerformancePassed = qg.checkPerformanceGates(profilingResult, result)
	}

	// Overall pass/fail
	result.Passed = result.ValidationPassed && result.PerformancePassed

	return result
}

// checkValidationGates validates evidence backing requirements
func (qg *QualityGateConfig) checkValidationGates(
	vr *ValidationReport,
	result *QualityGateResult,
) bool {
	passed := true

	// Check evidence coverage
	if vr.EvidenceCoverage < qg.MinEvidenceCoverage {
		passed = false
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationEvidenceCoverage,
			Severity: SeverityCritical,
			Description: fmt.Sprintf("Evidence coverage below threshold: %.1f%% < %.1f%%",
				vr.EvidenceCoverage, qg.MinEvidenceCoverage),
			Actual:   vr.EvidenceCoverage,
			Expected: qg.MinEvidenceCoverage,
		})
	}

	// Check unbacked claims
	if vr.UnbackedClaims > qg.AllowUnbackedClaims {
		passed = false
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationUnbackedClaims,
			Severity: SeverityHigh,
			Description: fmt.Sprintf("Too many unbacked claims: %d > %d",
				vr.UnbackedClaims, qg.AllowUnbackedClaims),
			Actual:   vr.UnbackedClaims,
			Expected: qg.AllowUnbackedClaims,
		})
	}

	// Check invalid evidence
	if vr.InvalidEvidence > 0 {
		passed = false
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationInvalidEvidence,
			Severity: SeverityHigh,
			Description: fmt.Sprintf("Invalid evidence references found: %d",
				vr.InvalidEvidence),
			Actual:   vr.InvalidEvidence,
			Expected: 0,
		})
	}

	return passed
}

// checkPerformanceGates validates performance requirements
func (qg *QualityGateConfig) checkPerformanceGates(
	pr *ProfilingReport,
	result *QualityGateResult,
) bool {
	passed := true

	// Check total execution time
	if pr.TotalDuration > qg.MaxExecutionTime {
		passed = false
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationExecutionTime,
			Severity: SeverityMedium,
			Description: fmt.Sprintf("Workflow execution time exceeded: %v > %v",
				pr.TotalDuration, qg.MaxExecutionTime),
			Actual:   pr.TotalDuration,
			Expected: qg.MaxExecutionTime,
		})
	}

	// Check individual agent execution times
	for _, exec := range pr.AgentExecutions {
		if exec.Duration > qg.MaxAgentExecutionTime {
			passed = false
			result.Violations = append(result.Violations, Violation{
				Type:     ViolationAgentTime,
				Severity: SeverityLow,
				Description: fmt.Sprintf("Agent %s execution time exceeded: %v > %v",
					exec.Agent, exec.Duration, qg.MaxAgentExecutionTime),
				Actual:   exec.Duration,
				Expected: qg.MaxAgentExecutionTime,
			})
		}
	}

	// Check memory usage
	if pr.PeakMemoryUsage > qg.MaxMemoryUsage {
		passed = false
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationMemoryUsage,
			Severity: SeverityHigh,
			Description: fmt.Sprintf("Peak memory usage exceeded: %d > %d bytes",
				pr.PeakMemoryUsage, qg.MaxMemoryUsage),
			Actual:   pr.PeakMemoryUsage,
			Expected: qg.MaxMemoryUsage,
		})
	}

	// Check profiling overhead
	if pr.ProfilingOverhead > qg.PerformanceOverheadTarget {
		// Don't fail on overhead, just warn
		result.Violations = append(result.Violations, Violation{
			Type:     ViolationProfilingOverhead,
			Severity: SeverityLow,
			Description: fmt.Sprintf("Profiling overhead exceeded target: %.2f%% > %.2f%%",
				pr.ProfilingOverhead, qg.PerformanceOverheadTarget),
			Actual:   pr.ProfilingOverhead,
			Expected: qg.PerformanceOverheadTarget,
		})
	}

	return passed
}
