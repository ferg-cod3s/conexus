# Conexus Validation Guide

**Version**: 0.0.5  
**Last Updated**: 2025-01-15

---

## Overview

Conexus implements a **two-layer validation system** to ensure all agent outputs meet quality standards:

1. **Evidence Validation** - Verifies that every claim has file/line references
2. **Schema Validation** - Ensures compliance with AGENT_OUTPUT_V1 format

This guide covers how to use, configure, and extend the validation systems.

---

## Table of Contents

- [Evidence Validation](#evidence-validation)
- [Schema Validation](#schema-validation)
- [Integration Patterns](#integration-patterns)
- [Configuration](#configuration)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

---

## Evidence Validation

### Overview

Evidence validation ensures that **100% of agent output items** have valid file/line references. This prevents hallucinations and maintains traceability.

### Core Concepts

**Evidence Item**: A reference to source code location
```go
type Item struct {
    FilePath        string `json:"file_path"`
    LineStart       int    `json:"line_start"`
    LineEnd         int    `json:"line_end"`
    EvidenceFilePath   string `json:"evidence_file_path"`
    EvidenceLineStart  int    `json:"evidence_line_start"`
    EvidenceLineEnd    int    `json:"evidence_line_end"`
}
```

**Validation Requirements**:
- ✅ File path must be non-empty
- ✅ Line numbers must be positive
- ✅ Line range must be valid (start ≤ end)
- ✅ Evidence path must match file path (for primary items)
- ✅ Evidence lines must overlap with item lines

### Basic Usage

```go
import (
    "context"
    "github.com/ferg-cod3s/conexus/internal/validation/evidence"
    "github.com/ferg-cod3s/conexus/pkg/schema"
)

// Create validator
validator := evidence.NewValidator()

// Validate agent output
output := &schema.AgentOutput{
    SchemaVersion: "AGENT_OUTPUT_V1",
    Items: []schema.Item{
        {
            Type: "function",
            Name: "HandleRequest",
            FilePath: "/internal/server/handler.go",
            LineStart: 42,
            LineEnd: 68,
            EvidenceFilePath: "/internal/server/handler.go",
            EvidenceLineStart: 42,
            EvidenceLineEnd: 68,
        },
    },
}

report, err := validator.Validate(context.Background(), output)
if err != nil {
    log.Fatalf("Validation error: %v", err)
}

if !report.IsValid {
    log.Printf("Validation failed: %d errors", len(report.Errors))
    for _, e := range report.Errors {
        log.Printf("  - %s", e)
    }
}
```

### Validation Report Structure

```go
type ValidationReport struct {
    IsValid         bool
    Errors          []string
    TotalItems      int
    ValidItems      int
    MissingEvidence int
    Summary         string
}
```

### Example Report

```
Evidence Validation Report
--------------------------
Status: FAILED
Total Items: 10
Valid Items: 8
Missing Evidence: 2

Errors:
  - Item "ProcessData" (file.go:42-68): missing evidence_file_path
  - Item "ValidateInput" (validator.go:100-120): line range invalid (start > end)

Coverage: 80.00%
```

---

## Schema Validation

### Overview

Schema validation ensures that all agent outputs conform to the **AGENT_OUTPUT_V1** specification, including required fields, types, and value ranges.

### Required Fields

```json
{
  "schema_version": "AGENT_OUTPUT_V1",        // REQUIRED: Must be exact string
  "task_description": "...",                  // REQUIRED: Non-empty string
  "result_summary": "...",                    // REQUIRED: Non-empty string
  "confidence_score": 0.95,                   // REQUIRED: 0.0 - 1.0
  "items": [...],                             // REQUIRED: Array (can be empty)
  "files_examined": [...],                    // REQUIRED: Array (can be empty)
  "metadata": {...}                           // REQUIRED: Object with specific fields
}
```

### Basic Usage

```go
import (
    "context"
    "github.com/ferg-cod3s/conexus/internal/validation/schema"
)

// Create validator
validator := schema.NewValidator()

// Validate output
output := &schema.AgentOutput{
    SchemaVersion: "AGENT_OUTPUT_V1",
    TaskDescription: "Locate HTTP handlers",
    ResultSummary: "Found 5 handlers",
    ConfidenceScore: 0.95,
    Items: []schema.Item{...},
    FilesExamined: []string{"/path/to/file.go"},
    Metadata: schema.Metadata{
        AgentName: "locator",
        ExecutionTimeMs: 45,
    },
}

report, err := validator.Validate(context.Background(), output)
if err != nil {
    log.Fatalf("Schema validation failed: %v", err)
}
```

### Validation Rules

#### Top-Level Fields

| Field | Type | Rules |
|-------|------|-------|
| `schema_version` | string | Must be "AGENT_OUTPUT_V1" |
| `task_description` | string | Non-empty, max 1000 chars |
| `result_summary` | string | Non-empty, max 5000 chars |
| `confidence_score` | float64 | 0.0 ≤ score ≤ 1.0 |
| `items` | array | Can be empty |
| `files_examined` | array | Can be empty |
| `metadata` | object | Must contain required fields |

#### Item Fields

| Field | Type | Rules |
|-------|------|-------|
| `type` | string | One of: function, file, class, variable, etc. |
| `name` | string | Non-empty |
| `file_path` | string | Valid file path |
| `line_start` | int | > 0 |
| `line_end` | int | ≥ line_start |
| `classification` | string | One of: primary, supporting, context |
| `explanation` | string | Non-empty |

#### Metadata Fields

| Field | Type | Rules |
|-------|------|-------|
| `agent_name` | string | Non-empty |
| `execution_time_ms` | int | ≥ 0 |
| `timestamp` | string | RFC3339 format (optional) |

---

## Integration Patterns

### Pattern 1: Inline Validation

Validate immediately after agent execution:

```go
func executeAgent(ctx context.Context, agent Agent) (*schema.AgentOutput, error) {
    // Execute agent
    output, err := agent.Execute(ctx)
    if err != nil {
        return nil, fmt.Errorf("agent execution failed: %w", err)
    }

    // Validate schema
    schemaValidator := schema.NewValidator()
    schemaReport, err := schemaValidator.Validate(ctx, output)
    if err != nil || !schemaReport.IsValid {
        return nil, fmt.Errorf("schema validation failed: %v", err)
    }

    // Validate evidence
    evidenceValidator := evidence.NewValidator()
    evidenceReport, err := evidenceValidator.Validate(ctx, output)
    if err != nil || !evidenceReport.IsValid {
        return nil, fmt.Errorf("evidence validation failed: %v", err)
    }

    return output, nil
}
```

### Pattern 2: Workflow Validation Gate

Add validation as a quality gate in workflows:

```go
func (w *Workflow) executeWithValidation(ctx context.Context) error {
    for _, step := range w.steps {
        output, err := step.Execute(ctx)
        if err != nil {
            return err
        }

        // Quality gate: validate before proceeding
        if err := w.validateOutput(ctx, output); err != nil {
            return fmt.Errorf("quality gate failed at step %s: %w", 
                step.Name, err)
        }

        w.results = append(w.results, output)
    }
    return nil
}
```

### Pattern 3: Batch Validation

Validate multiple outputs together:

```go
func validateBatch(ctx context.Context, outputs []*schema.AgentOutput) error {
    validator := evidence.NewValidator()
    
    var allErrors []string
    for i, output := range outputs {
        report, err := validator.Validate(ctx, output)
        if err != nil {
            return err
        }
        
        if !report.IsValid {
            allErrors = append(allErrors, 
                fmt.Sprintf("Output %d: %s", i, report.Summary))
        }
    }
    
    if len(allErrors) > 0 {
        return fmt.Errorf("batch validation failed:\n%s", 
            strings.Join(allErrors, "\n"))
    }
    
    return nil
}
```

### Pattern 4: Test Assertions

Use in integration tests:

```go
func TestAgentOutput(t *testing.T) {
    output := executeAgent(t)
    
    // Schema validation
    schemaValidator := schema.NewValidator()
    schemaReport, err := schemaValidator.Validate(context.Background(), output)
    require.NoError(t, err)
    require.True(t, schemaReport.IsValid, "Schema validation failed")
    
    // Evidence validation
    evidenceValidator := evidence.NewValidator()
    evidenceReport, err := evidenceValidator.Validate(context.Background(), output)
    require.NoError(t, err)
    assert.Equal(t, 100.0, evidenceReport.Coverage, "Expected 100% evidence")
}
```

---

## Configuration

### Environment Variables

```bash
# Enable/disable validation
export CONEXUS_VALIDATE_EVIDENCE=true
export CONEXUS_VALIDATE_SCHEMA=true

# Validation strictness
export CONEXUS_SCHEMA_MODE=strict  # strict|lenient
export CONEXUS_REQUIRE_FULL_EVIDENCE=true

# Error tolerance
export CONEXUS_MAX_VALIDATION_ERRORS=10

# Logging
export CONEXUS_VALIDATION_LOG_LEVEL=info  # debug|info|warn|error
```

### Programmatic Configuration

```go
type ValidationConfig struct {
    RequireFullEvidence  bool
    SchemaMode          string  // "strict" or "lenient"
    MaxErrors           int
    FailFast            bool
}

config := ValidationConfig{
    RequireFullEvidence: true,
    SchemaMode: "strict",
    MaxErrors: 10,
    FailFast: false,
}

validator := evidence.NewValidatorWithConfig(config)
```

---

## Error Handling

### Common Validation Errors

#### Missing Evidence

```
Error: Item "ProcessData" missing evidence_file_path
Solution: Ensure all items have evidence_file_path field populated
```

#### Invalid Line Range

```
Error: Line range invalid (start=100, end=50)
Solution: Ensure line_start <= line_end
```

#### Schema Version Mismatch

```
Error: Invalid schema_version "AGENT_OUTPUT_V2" (expected "AGENT_OUTPUT_V1")
Solution: Use exact string "AGENT_OUTPUT_V1"
```

#### Confidence Score Out of Range

```
Error: Confidence score 1.5 out of valid range [0.0, 1.0]
Solution: Ensure 0.0 <= confidence_score <= 1.0
```

### Error Recovery Strategies

#### Strategy 1: Retry with Corrections

```go
func executeWithRetry(ctx context.Context, agent Agent, maxRetries int) (*schema.AgentOutput, error) {
    for attempt := 0; attempt < maxRetries; attempt++ {
        output, err := agent.Execute(ctx)
        if err != nil {
            return nil, err
        }

        report, err := validator.Validate(ctx, output)
        if err != nil {
            return nil, err
        }

        if report.IsValid {
            return output, nil
        }

        // Log validation errors for debugging
        log.Printf("Attempt %d failed validation: %s", attempt+1, report.Summary)
        
        // Optional: Send feedback to agent for correction
        agent.ApplyFeedback(report.Errors)
    }

    return nil, fmt.Errorf("validation failed after %d attempts", maxRetries)
}
```

#### Strategy 2: Partial Acceptance

```go
func extractValidItems(output *schema.AgentOutput) (*schema.AgentOutput, error) {
    validator := evidence.NewValidator()
    
    validItems := make([]schema.Item, 0)
    for _, item := range output.Items {
        // Validate individual item
        singleItemOutput := &schema.AgentOutput{
            SchemaVersion: output.SchemaVersion,
            Items: []schema.Item{item},
        }
        
        report, _ := validator.Validate(context.Background(), singleItemOutput)
        if report.IsValid {
            validItems = append(validItems, item)
        }
    }
    
    output.Items = validItems
    return output, nil
}
```

---

## Best Practices

### 1. Validate Early

```go
// ✅ Good: Validate immediately after generation
output, err := agent.Execute(ctx)
if err != nil {
    return nil, err
}

if err := validate(ctx, output); err != nil {
    return nil, fmt.Errorf("validation failed: %w", err)
}
```

```go
// ❌ Bad: Delay validation until later
outputs = append(outputs, output)  // Store unvalidated
// ... many lines later ...
validate(ctx, outputs)  // Hard to trace errors
```

### 2. Use Structured Logging

```go
report, err := validator.Validate(ctx, output)
if !report.IsValid {
    log.WithFields(log.Fields{
        "agent": output.Metadata.AgentName,
        "total_items": report.TotalItems,
        "valid_items": report.ValidItems,
        "coverage": report.Coverage,
    }).Error("Validation failed")
}
```

### 3. Add Context to Errors

```go
if !report.IsValid {
    return fmt.Errorf("agent %s validation failed (step %d): %s",
        agentName, stepNumber, report.Summary)
}
```

### 4. Test Validation in CI/CD

```bash
# In CI pipeline
go test -run TestValidation ./...

# Fail build on validation errors
if ! go test -run TestEvidenceValidation ./...; then
    echo "Evidence validation tests failed"
    exit 1
fi
```

### 5. Monitor Validation Metrics

```go
type ValidationMetrics struct {
    TotalValidations int64
    FailedValidations int64
    AverageCoverage float64
}

func recordValidationResult(report *ValidationReport) {
    metrics.TotalValidations++
    if !report.IsValid {
        metrics.FailedValidations++
    }
    metrics.AverageCoverage = updateAverage(metrics.AverageCoverage, report.Coverage)
}
```

---

## Troubleshooting

### Issue: Validation Too Strict

**Symptom**: Valid outputs being rejected

**Solutions**:
1. Check schema version: `"AGENT_OUTPUT_V1"` (exact string)
2. Verify confidence score range: `0.0 <= score <= 1.0`
3. Enable lenient mode: `export CONEXUS_SCHEMA_MODE=lenient`

### Issue: Missing Evidence Errors

**Symptom**: "Item missing evidence_file_path"

**Solutions**:
1. Ensure all items populate evidence fields:
   ```go
   item.EvidenceFilePath = item.FilePath
   item.EvidenceLineStart = item.LineStart
   item.EvidenceLineEnd = item.LineEnd
   ```
2. Check file path format (absolute vs relative)
3. Verify line numbers are positive

### Issue: Performance Degradation

**Symptom**: Validation taking too long

**Solutions**:
1. Enable validation caching
2. Reduce validation frequency (validate batches)
3. Use async validation for non-critical paths
4. Profile validation code

### Issue: Inconsistent Validation Results

**Symptom**: Same output validates differently

**Solutions**:
1. Ensure deterministic line numbering
2. Check for race conditions in parallel validation
3. Verify file paths are normalized (absolute)
4. Use consistent validator instances

---

## Advanced Topics

### Custom Validators

Create domain-specific validators:

```go
type CustomValidator struct {
    baseValidator *evidence.Validator
}

func (v *CustomValidator) ValidateWithDomainRules(
    ctx context.Context,
    output *schema.AgentOutput,
) (*ValidationReport, error) {
    // First, run base validation
    report, err := v.baseValidator.Validate(ctx, output)
    if err != nil || !report.IsValid {
        return report, err
    }

    // Add custom rules
    if len(output.Items) < 3 {
        report.IsValid = false
        report.Errors = append(report.Errors, 
            "Expected at least 3 items")
    }

    return report, nil
}
```

### Validation Extensions

Extend validation with plugins:

```go
type ValidationPlugin interface {
    Name() string
    Validate(ctx context.Context, output *schema.AgentOutput) error
}

type PluggableValidator struct {
    baseValidator *evidence.Validator
    plugins       []ValidationPlugin
}

func (v *PluggableValidator) Validate(
    ctx context.Context,
    output *schema.AgentOutput,
) (*ValidationReport, error) {
    // Run base validation
    report, err := v.baseValidator.Validate(ctx, output)
    if err != nil {
        return nil, err
    }

    // Run plugins
    for _, plugin := range v.plugins {
        if err := plugin.Validate(ctx, output); err != nil {
            report.IsValid = false
            report.Errors = append(report.Errors,
                fmt.Sprintf("%s: %v", plugin.Name(), err))
        }
    }

    return report, nil
}
```

---

## API Reference

See **[API Reference](api-reference.md#validation-api)** for complete API documentation.

---

## See Also

- **[Profiling Guide](profiling-guide.md)** - Performance monitoring
- **[API Reference](api-reference.md)** - Complete API documentation
- **[Testing Strategy](../contributing/testing-strategy.md)** - Testing best practices

---

**Last Updated**: 2025-01-15  
**Maintainer**: Conexus Team
