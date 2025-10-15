package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/validation/evidence"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// ValidationReport aggregates validation results across workflow
type ValidationReport struct {
	WorkflowID       string
	Timestamp        time.Time
	TotalResponses   int
	ValidResponses   int
	InvalidResponses int
	EvidenceCoverage float64
	UnbackedClaims   int
	InvalidEvidence  int
	Details          []AgentValidationResult
}

// AgentValidationResult contains validation result for single agent
type AgentValidationResult struct {
	AgentID          string
	RequestID        string
	Valid            bool
	EvidenceCoverage float64
	UnbackedClaims   []evidence.UnbackedClaim
	InvalidEvidence  []evidence.InvalidEvidence
}

// ProfilingReport aggregates performance metrics across workflow
type ProfilingReport struct {
	WorkflowID        string
	Timestamp         time.Time
	TotalDuration     time.Duration
	AgentExecutions   []AgentExecutionMetrics
	PeakMemoryUsage   uint64
	TotalMemoryUsed   uint64
	ProfilingOverhead float64
}

// AgentExecutionMetrics contains performance metrics for single agent
type AgentExecutionMetrics struct {
	Agent           string
	RequestID       string
	Duration        time.Duration
	MemoryAllocated uint64
	MemoryFreed     uint64
	Success         bool
}

// WorkflowReport combines validation and profiling reports
type WorkflowReport struct {
	WorkflowID      string
	Timestamp       time.Time
	Validation      *ValidationReport
	Profiling       *ProfilingReport
	QualityGates    *QualityGateResult
	OverallStatus   string
}

// ReportFormat specifies the output format for reports
type ReportFormat string

const (
	FormatJSON     ReportFormat = "json"
	FormatText     ReportFormat = "text"
	FormatMarkdown ReportFormat = "markdown"
)

// GenerateReport creates a comprehensive workflow report
func GenerateReport(
	workflowID string,
	validation *ValidationReport,
	profiling *ProfilingReport,
	qualityGates *QualityGateResult,
) *WorkflowReport {
	status := "success"
	if qualityGates != nil && !qualityGates.Passed {
		status = "failed"
	}

	return &WorkflowReport{
		WorkflowID:    workflowID,
		Timestamp:     time.Now(),
		Validation:    validation,
		Profiling:     profiling,
		QualityGates:  qualityGates,
		OverallStatus: status,
	}
}

// Export exports the report in the specified format
func (r *WorkflowReport) Export(format ReportFormat) (string, error) {
	switch format {
	case FormatJSON:
		return r.exportJSON()
	case FormatText:
		return r.exportText()
	case FormatMarkdown:
		return r.exportMarkdown()
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// exportJSON exports report as JSON
func (r *WorkflowReport) exportJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal error: %w", err)
	}
	return string(data), nil
}

// exportText exports report as plain text
func (r *WorkflowReport) exportText() (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("WORKFLOW REPORT: %s\n", r.WorkflowID))
	sb.WriteString(fmt.Sprintf("Status: %s\n", r.OverallStatus))
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", r.Timestamp.Format(time.RFC3339)))
	sb.WriteString("\n")

	// Validation section
	if r.Validation != nil {
		sb.WriteString("=== VALIDATION ===\n")
		sb.WriteString(fmt.Sprintf("Total Responses: %d\n", r.Validation.TotalResponses))
		sb.WriteString(fmt.Sprintf("Valid: %d\n", r.Validation.ValidResponses))
		sb.WriteString(fmt.Sprintf("Invalid: %d\n", r.Validation.InvalidResponses))
		sb.WriteString(fmt.Sprintf("Evidence Coverage: %.1f%%\n", r.Validation.EvidenceCoverage))
		sb.WriteString(fmt.Sprintf("Unbacked Claims: %d\n", r.Validation.UnbackedClaims))
		sb.WriteString(fmt.Sprintf("Invalid Evidence: %d\n", r.Validation.InvalidEvidence))
		sb.WriteString("\n")
	}

	// Profiling section
	if r.Profiling != nil {
		sb.WriteString("=== PERFORMANCE ===\n")
		sb.WriteString(fmt.Sprintf("Total Duration: %v\n", r.Profiling.TotalDuration))
		sb.WriteString(fmt.Sprintf("Peak Memory: %d bytes (%.2f MB)\n", 
			r.Profiling.PeakMemoryUsage, 
			float64(r.Profiling.PeakMemoryUsage)/(1024*1024)))
		sb.WriteString(fmt.Sprintf("Profiling Overhead: %.2f%%\n", r.Profiling.ProfilingOverhead))
		sb.WriteString(fmt.Sprintf("Agent Executions: %d\n", len(r.Profiling.AgentExecutions)))
		sb.WriteString("\n")
	}

	// Quality gates section
	if r.QualityGates != nil {
		sb.WriteString("=== QUALITY GATES ===\n")
		sb.WriteString(fmt.Sprintf("Passed: %v\n", r.QualityGates.Passed))
		sb.WriteString(fmt.Sprintf("Validation: %v\n", r.QualityGates.ValidationPassed))
		sb.WriteString(fmt.Sprintf("Performance: %v\n", r.QualityGates.PerformancePassed))
		
		if len(r.QualityGates.Violations) > 0 {
			sb.WriteString(fmt.Sprintf("Violations: %d\n", len(r.QualityGates.Violations)))
			for _, v := range r.QualityGates.Violations {
				sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", 
					v.Severity, v.Type, v.Description))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// exportMarkdown exports report as Markdown
func (r *WorkflowReport) exportMarkdown() (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Workflow Report: %s\n\n", r.WorkflowID))
	sb.WriteString(fmt.Sprintf("**Status:** %s  \n", r.OverallStatus))
	sb.WriteString(fmt.Sprintf("**Timestamp:** %s  \n\n", r.Timestamp.Format(time.RFC3339)))

	// Validation section
	if r.Validation != nil {
		sb.WriteString("## Validation Results\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Responses | %d |\n", r.Validation.TotalResponses))
		sb.WriteString(fmt.Sprintf("| Valid | %d |\n", r.Validation.ValidResponses))
		sb.WriteString(fmt.Sprintf("| Invalid | %d |\n", r.Validation.InvalidResponses))
		sb.WriteString(fmt.Sprintf("| Evidence Coverage | %.1f%% |\n", r.Validation.EvidenceCoverage))
		sb.WriteString(fmt.Sprintf("| Unbacked Claims | %d |\n", r.Validation.UnbackedClaims))
		sb.WriteString(fmt.Sprintf("| Invalid Evidence | %d |\n\n", r.Validation.InvalidEvidence))
	}

	// Profiling section
	if r.Profiling != nil {
		sb.WriteString("## Performance Metrics\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Total Duration | %v |\n", r.Profiling.TotalDuration))
		sb.WriteString(fmt.Sprintf("| Peak Memory | %.2f MB |\n", 
			float64(r.Profiling.PeakMemoryUsage)/(1024*1024)))
		sb.WriteString(fmt.Sprintf("| Profiling Overhead | %.2f%% |\n", r.Profiling.ProfilingOverhead))
		sb.WriteString(fmt.Sprintf("| Agent Executions | %d |\n\n", len(r.Profiling.AgentExecutions)))
	}

	// Quality gates section
	if r.QualityGates != nil {
		sb.WriteString("## Quality Gates\n\n")
		
		passIcon := "✅"
		if !r.QualityGates.Passed {
			passIcon = "❌"
		}
		
		sb.WriteString(fmt.Sprintf("%s **Overall:** %v  \n", passIcon, r.QualityGates.Passed))
		
		valIcon := "✅"
		if !r.QualityGates.ValidationPassed {
			valIcon = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s **Validation:** %v  \n", valIcon, r.QualityGates.ValidationPassed))
		
		perfIcon := "✅"
		if !r.QualityGates.PerformancePassed {
			perfIcon = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s **Performance:** %v  \n\n", perfIcon, r.QualityGates.PerformancePassed))
		
		if len(r.QualityGates.Violations) > 0 {
			sb.WriteString("### Violations\n\n")
			for _, v := range r.QualityGates.Violations {
				icon := "⚠️"
				if v.Severity == SeverityCritical {
					icon = "🚨"
				} else if v.Severity == SeverityHigh {
					icon = "⛔"
				}
				sb.WriteString(fmt.Sprintf("%s **[%s] %s:** %s\n", 
					icon, v.Severity, v.Type, v.Description))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// CreateValidationReport aggregates validation results
func CreateValidationReport(
	workflowID string,
	responses []schema.AgentResponse,
	validator *evidence.Validator,
) (*ValidationReport, error) {
	report := &ValidationReport{
		WorkflowID:     workflowID,
		Timestamp:      time.Now(),
		TotalResponses: len(responses),
		Details:        make([]AgentValidationResult, 0),
	}

	totalCoverage := 0.0
	validCount := 0

	for _, resp := range responses {
		// Skip non-AGENT_OUTPUT_V1 responses
		if resp.Output == nil {
			continue
		}

		// Validate the output
		validationResult, err := validator.Validate(resp.Output)
		if err != nil {
			return nil, fmt.Errorf("validation error for %s: %w", resp.AgentID, err)
		}

		// Track metrics
		if validationResult.Valid {
			validCount++
		}
		totalCoverage += validationResult.CoveragePercentage
		report.UnbackedClaims += len(validationResult.UnbackedClaims)
		report.InvalidEvidence += len(validationResult.InvalidEvidence)

		// Add details
		report.Details = append(report.Details, AgentValidationResult{
			AgentID:          resp.AgentID,
			RequestID:        resp.RequestID,
			Valid:            validationResult.Valid,
			EvidenceCoverage: validationResult.CoveragePercentage,
			UnbackedClaims:   validationResult.UnbackedClaims,
			InvalidEvidence:  validationResult.InvalidEvidence,
		})
	}

	report.ValidResponses = validCount
	report.InvalidResponses = report.TotalResponses - validCount
	
	if report.TotalResponses > 0 {
		report.EvidenceCoverage = totalCoverage / float64(report.TotalResponses)
	}

	return report, nil
}
// CreateValidationReportFromResults creates a validation report from pre-validated results
func CreateValidationReportFromResults(
	workflowID string,
	results []AgentValidationResult,
) *ValidationReport {
	report := &ValidationReport{
		WorkflowID:     workflowID,
		Timestamp:      time.Now(),
		TotalResponses: len(results),
		Details:        results,
	}

	totalCoverage := 0.0
	validCount := 0

	for _, result := range results {
		if result.Valid {
			validCount++
		}
		totalCoverage += result.EvidenceCoverage
		report.UnbackedClaims += len(result.UnbackedClaims)
		report.InvalidEvidence += len(result.InvalidEvidence)
	}

	report.ValidResponses = validCount
	report.InvalidResponses = report.TotalResponses - validCount
	
	if report.TotalResponses > 0 {
		report.EvidenceCoverage = totalCoverage / float64(report.TotalResponses)
	}

	return report
}
