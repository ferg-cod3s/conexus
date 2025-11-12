package analyzer

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/tool"
	"github.com/ferg-cod3s/conexus/pkg/schema"
)

// AnalyzerAgent analyzes code to understand control flow, data flow, and patterns
type AnalyzerAgent struct {
	executor *tool.Executor
}

// New creates a new AnalyzerAgent
func New(executor *tool.Executor) *AnalyzerAgent {
	return &AnalyzerAgent{
		executor: executor,
	}
}

// Execute processes an agent request to analyze code
func (a *AnalyzerAgent) Execute(ctx context.Context, req schema.AgentRequest) (schema.AgentResponse, error) {
	response := schema.AgentResponse{
		RequestID: req.RequestID,
		AgentID:   "codebase-analyzer",
		Status:    schema.StatusComplete,
		Timestamp: req.Timestamp,
	}

	// Analyze the specified files
	output, err := a.analyzeFiles(ctx, req)
	if err != nil {
		response.Status = schema.StatusError
		response.Error = &schema.AgentError{
			Code:        "ANALYZER_ERROR",
			Message:     err.Error(),
			Recoverable: true,
		}
		return response, err
	}

	response.Output = output
	return response, nil
}

// analyzeFiles performs comprehensive analysis on specified files
func (a *AnalyzerAgent) analyzeFiles(ctx context.Context, req schema.AgentRequest) (*schema.AgentOutputV1, error) {
	if len(req.Task.Files) == 0 {
		return nil, fmt.Errorf("no files specified for analysis")
	}

	output := &schema.AgentOutputV1{
		Version:          "AGENT_OUTPUT_V1",
		ComponentName:    determineComponentName(req.Task.Files),
		ScopeDescription: fmt.Sprintf("Analysis of %d file(s): %v", len(req.Task.Files), fileNames(req.Task.Files)),
		Overview:         "",
	}

	var allEvidence []schema.Evidence
	var fileErrors []string

	// Analyze each file
	for _, file := range req.Task.Files {
		// Validate file exists
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				fileErrors = append(fileErrors, fmt.Sprintf("%s: file not found", file))
			} else {
				fileErrors = append(fileErrors, fmt.Sprintf("%s: %v", file, err))
			}
			continue
		}

		// Read file content
		params := tool.ToolParams{Path: file}
		result, err := a.executor.Execute(ctx, "read", params, req.Permissions)
		if err != nil {
			fileErrors = append(fileErrors, fmt.Sprintf("%s: %v", file, err))
			continue
		}

		content, ok := result.Output.(string)
		if !ok {
			fileErrors = append(fileErrors, fmt.Sprintf("%s: failed to read content", file))
			continue
		}

		// Parse and analyze
		fileAnalysis := a.analyzeFileContent(file, content)

		// Merge results
		output.EntryPoints = append(output.EntryPoints, fileAnalysis.EntryPoints...)
		output.CallGraph = append(output.CallGraph, fileAnalysis.CallGraph...)
		output.DataFlow.Inputs = append(output.DataFlow.Inputs, fileAnalysis.DataFlow.Inputs...)
		output.DataFlow.Transformations = append(output.DataFlow.Transformations, fileAnalysis.DataFlow.Transformations...)
		output.DataFlow.Outputs = append(output.DataFlow.Outputs, fileAnalysis.DataFlow.Outputs...)
		output.StateManagement = append(output.StateManagement, fileAnalysis.StateManagement...)
		output.SideEffects = append(output.SideEffects, fileAnalysis.SideEffects...)
		output.ErrorHandling = append(output.ErrorHandling, fileAnalysis.ErrorHandling...)
		output.Patterns = append(output.Patterns, fileAnalysis.Patterns...)
		output.Concurrency = append(output.Concurrency, fileAnalysis.Concurrency...)
		allEvidence = append(allEvidence, fileAnalysis.RawEvidence...)
	}

	// If all files had errors, return an error
	if len(fileErrors) > 0 && len(allEvidence) == 0 {
		return nil, fmt.Errorf("failed to analyze files: %s", strings.Join(fileErrors, "; "))
	}

	// Generate overview
	output.Overview = generateOverview(output)
	output.RawEvidence = allEvidence
	output.Limitations = []string{
		"Analysis based on text parsing and regex patterns, not AST",
		"May miss complex control flow patterns",
		"Call graph construction is heuristic-based",
		"Line number precision limited by regex matching",
	}
	output.OpenQuestions = []string{
		"Should we implement full AST parsing for more accurate analysis?",
		"Are there additional patterns we should detect?",
	}

	return output, nil
}

// analyzeFileContent performs detailed analysis of file content
func (a *AnalyzerAgent) analyzeFileContent(filepath string, content string) *schema.AgentOutputV1 {
	result := &schema.AgentOutputV1{
		DataFlow: schema.DataFlow{
			Inputs:          []schema.DataPoint{},
			Transformations: []schema.Transformation{},
			Outputs:         []schema.DataPoint{},
		},
	}

	lines := strings.Split(content, "\n")

	// Find entry points (exported functions)
	entryPoints := findEntryPoints(filepath, lines)
	result.EntryPoints = entryPoints
	for _, ep := range entryPoints {
		result.RawEvidence = append(result.RawEvidence, schema.Evidence{
			Claim: fmt.Sprintf("Entry point: %s", ep.Symbol),
			File:  filepath,
			Lines: ep.Lines,
		})
	}

	// Build call graph
	callGraph := findFunctionCalls(filepath, lines, entryPoints)
	result.CallGraph = callGraph
	for _, edge := range callGraph {
		result.RawEvidence = append(result.RawEvidence, schema.Evidence{
			Claim: fmt.Sprintf("Function call: %s → %s", edge.From, edge.To),
			File:  filepath,
			Lines: fmt.Sprintf("%d-%d", edge.ViaLine, edge.ViaLine),
		})
	}

	// Detect data flow
	dataFlow := analyzeDataFlow(filepath, lines)
	result.DataFlow = dataFlow

	// Find state operations
	stateOps := findStateOperations(filepath, lines)
	result.StateManagement = stateOps
	for _, op := range stateOps {
		result.RawEvidence = append(result.RawEvidence, schema.Evidence{
			Claim: fmt.Sprintf("State operation: %s %s", op.Operation, op.Entity),
			File:  filepath,
			Lines: op.Lines,
		})
	}

	// Detect side effects
	sideEffects := findSideEffects(filepath, lines)
	result.SideEffects = sideEffects
	for _, se := range sideEffects {
		result.RawEvidence = append(result.RawEvidence, schema.Evidence{
			Claim: fmt.Sprintf("Side effect: %s", se.Type),
			File:  filepath,
			Lines: fmt.Sprintf("%d-%d", se.Line, se.Line),
		})
	}

	// Analyze error handling
	errorHandlers := findErrorHandling(filepath, lines)
	result.ErrorHandling = errorHandlers
	for _, eh := range errorHandlers {
		result.RawEvidence = append(result.RawEvidence, schema.Evidence{
			Claim: fmt.Sprintf("Error handling: %s", eh.Type),
			File:  filepath,
			Lines: eh.Lines,
		})
	}

	// Detect patterns
	patterns := detectPatterns(filepath, lines)
	result.Patterns = patterns

	// Find concurrency mechanisms
	concurrency := findConcurrency(filepath, lines)
	result.Concurrency = concurrency
	for _, c := range concurrency {
		result.RawEvidence = append(result.RawEvidence, schema.Evidence{
			Claim: fmt.Sprintf("Concurrency: %s", c.Mechanism),
			File:  filepath,
			Lines: c.Lines,
		})
	}

	return result
}

// findEntryPoints identifies exported functions and methods
func findEntryPoints(filepath string, lines []string) []schema.EntryPoint {
	var entryPoints []schema.EntryPoint
	funcRegex := regexp.MustCompile(`^func\s+(\w+)\(`)
	methodRegex := regexp.MustCompile(`^func\s+\(.*\)\s+(\w+)\(`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Check for function declarations
		if matches := funcRegex.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			// Exported if starts with capital letter
			if len(name) > 0 && isExported(name) {
				entryPoints = append(entryPoints, schema.EntryPoint{
					File:   filepath,
					Lines:  fmt.Sprintf("%d-%d", i+1, i+1),
					Symbol: name,
					Role:   "function",
				})
			}
		}

		// Check for method declarations
		if matches := methodRegex.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			if len(name) > 0 && isExported(name) {
				entryPoints = append(entryPoints, schema.EntryPoint{
					File:   filepath,
					Lines:  fmt.Sprintf("%d-%d", i+1, i+1),
					Symbol: name,
					Role:   "method",
				})
			}
		}
	}

	return entryPoints
}

// findFunctionCalls constructs a call graph
func findFunctionCalls(filepath string, lines []string, entryPoints []schema.EntryPoint) []schema.CallGraphEdge {
	var callGraph []schema.CallGraphEdge
	callRegex := regexp.MustCompile(`\b([A-Z]\w+)\(`)

	currentFunc := ""
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Track current function context
		if strings.HasPrefix(line, "func ") {
			for _, ep := range entryPoints {
				if strings.Contains(line, ep.Symbol+"(") {
					currentFunc = ep.Symbol
					break
				}
			}
		}

		// Find function calls
		if currentFunc != "" {
			matches := callRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				calledFunc := match[1]
				if calledFunc != currentFunc { // Avoid self-references
					callGraph = append(callGraph, schema.CallGraphEdge{
						From:    fmt.Sprintf("%s:%s", filepath, currentFunc),
						To:      fmt.Sprintf("%s:%s", filepath, calledFunc),
						ViaLine: i + 1,
					})
				}
			}
		}
	}

	return callGraph
}

// analyzeDataFlow tracks data inputs, transformations, and outputs
func analyzeDataFlow(filepath string, lines []string) schema.DataFlow {
	dataFlow := schema.DataFlow{
		Inputs:          []schema.DataPoint{},
		Transformations: []schema.Transformation{},
		Outputs:         []schema.DataPoint{},
	}

	// Find input parameters
	paramRegex := regexp.MustCompile(`func\s+\w+\((.*?)\)`)
	for i, line := range lines {
		if matches := paramRegex.FindStringSubmatch(line); matches != nil {
			if params := matches[1]; params != "" {
				dataFlow.Inputs = append(dataFlow.Inputs, schema.DataPoint{
					Source:      fmt.Sprintf("%s:%d", filepath, i+1),
					Name:        "parameters",
					Type:        params,
					Description: "Function input parameters",
				})
			}
		}
	}

	// Find transformations (assignments, operations)
	assignRegex := regexp.MustCompile(`^\s*(\w+)\s*:?=`)
	for i, line := range lines {
		if matches := assignRegex.FindStringSubmatch(line); matches != nil {
			varName := matches[1]
			dataFlow.Transformations = append(dataFlow.Transformations, schema.Transformation{
				File:        filepath,
				Lines:       fmt.Sprintf("%d-%d", i+1, i+1),
				Operation:   "assign",
				Description: fmt.Sprintf("Variable assignment: %s", varName),
			})
		}
	}

	// Find return statements (outputs)
	returnRegex := regexp.MustCompile(`^\s*return\s+(.+)`)
	for i, line := range lines {
		if matches := returnRegex.FindStringSubmatch(line); matches != nil {
			returnVal := matches[1]
			dataFlow.Outputs = append(dataFlow.Outputs, schema.DataPoint{
				Source:      fmt.Sprintf("%s:%d", filepath, i+1),
				Name:        "return",
				Type:        "unknown",
				Description: fmt.Sprintf("Return value: %s", returnVal),
			})
		}
	}

	return dataFlow
}

// findStateOperations detects state management patterns
func findStateOperations(filepath string, lines []string) []schema.StateOperation {
	var stateOps []schema.StateOperation

	// Look for struct field assignments
	fieldRegex := regexp.MustCompile(`^\s*(\w+)\.(\w+)\s*=`)
	for i, line := range lines {
		if matches := fieldRegex.FindStringSubmatch(line); matches != nil {
			receiver := matches[1]
			field := matches[2]
			stateOps = append(stateOps, schema.StateOperation{
				File:        filepath,
				Lines:       fmt.Sprintf("%d-%d", i+1, i+1),
				Kind:        "memory",
				Operation:   "write",
				Entity:      fmt.Sprintf("%s.%s", receiver, field),
				Description: fmt.Sprintf("Struct field assignment: %s.%s", receiver, field),
			})
		}
	}

	return stateOps
}

// findSideEffects detects I/O, logging, and external interactions
func findSideEffects(filepath string, lines []string) []schema.SideEffect {
	var sideEffects []schema.SideEffect

	patterns := map[string]string{
		`fmt\.Print`:    "log",
		`log\.Print`:    "log",
		`os\.WriteFile`: "fs",
		`os\.ReadFile`:  "fs",
		`http\.`:        "http",
	}

	for i, line := range lines {
		for pattern, effectType := range patterns {
			if matched, _ := regexp.MatchString(pattern, line); matched {
				sideEffects = append(sideEffects, schema.SideEffect{
					File:        filepath,
					Line:        i + 1,
					Type:        effectType,
					Description: fmt.Sprintf("%s operation detected", effectType),
				})
			}
		}
	}

	return sideEffects
}

// findErrorHandling identifies error handling patterns
func findErrorHandling(filepath string, lines []string) []schema.ErrorHandler {
	var handlers []schema.ErrorHandler

	// Find error returns
	errorReturnRegex := regexp.MustCompile(`return.*,\s*err`)
	for i, line := range lines {
		if errorReturnRegex.MatchString(line) {
			handlers = append(handlers, schema.ErrorHandler{
				File:      filepath,
				Lines:     fmt.Sprintf("%d-%d", i+1, i+1),
				Type:      "throw",
				Condition: "error occurred",
				Effect:    "propagate",
			})
		}
	}

	// Find error checks
	errorCheckRegex := regexp.MustCompile(`if\s+err\s*!=\s*nil`)
	for i, line := range lines {
		if errorCheckRegex.MatchString(line) {
			handlers = append(handlers, schema.ErrorHandler{
				File:      filepath,
				Lines:     fmt.Sprintf("%d-%d", i+1, i+1),
				Type:      "guard",
				Condition: "err != nil",
				Effect:    "handle or propagate",
			})
		}
	}

	return handlers
}

// detectPatterns identifies common design patterns
func detectPatterns(filepath string, lines []string) []schema.Pattern {
	var patterns []schema.Pattern

	// Detect factory pattern
	if hasPattern(lines, `func\s+New\w+\(`) {
		patterns = append(patterns, schema.Pattern{
			Name:        "Factory",
			File:        filepath,
			Lines:       "1-N",
			Description: "Constructor function following New* naming convention",
		})
	}

	// Detect method receiver pattern
	if hasPattern(lines, `func\s+\(\w+\s+\*?\w+\)`) {
		patterns = append(patterns, schema.Pattern{
			Name:        "Method Receiver",
			File:        filepath,
			Lines:       "1-N",
			Description: "Methods defined on struct receivers",
		})
	}

	return patterns
}

// findConcurrency detects concurrency mechanisms
func findConcurrency(filepath string, lines []string) []schema.ConcurrencyMechanism {
	var mechanisms []schema.ConcurrencyMechanism

	patterns := map[string]string{
		`go\s+\w+\(`:      "goroutine",
		`make\(chan\s`:    "channel",
		`sync\.Mutex`:     "mutex",
		`sync\.RWMutex`:   "mutex",
		`sync\.WaitGroup`: "waitgroup",
	}

	for i, line := range lines {
		for pattern, mechanism := range patterns {
			if matched, _ := regexp.MatchString(pattern, line); matched {
				mechanisms = append(mechanisms, schema.ConcurrencyMechanism{
					File:        filepath,
					Lines:       fmt.Sprintf("%d-%d", i+1, i+1),
					Mechanism:   mechanism,
					Description: fmt.Sprintf("%s usage detected", mechanism),
				})
			}
		}
	}

	return mechanisms
}

// Helper functions

func determineComponentName(files []string) string {
	if len(files) == 1 {
		return files[0]
	}
	return fmt.Sprintf("Multi-file analysis (%d files)", len(files))
}

func fileNames(files []string) []string {
	names := make([]string, len(files))
	for i, f := range files {
		parts := strings.Split(f, "/")
		names[i] = parts[len(parts)-1]
	}
	return names
}

func generateOverview(output *schema.AgentOutputV1) string {
	return fmt.Sprintf(
		"Analyzed code contains %d entry point(s), %d function call(s), "+
			"%d state operation(s), %d side effect(s), and %d error handler(s). "+
			"Analysis covers data flow, control flow, and common patterns.",
		len(output.EntryPoints),
		len(output.CallGraph),
		len(output.StateManagement),
		len(output.SideEffects),
		len(output.ErrorHandling),
	)
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	firstRune := []rune(name)[0]
	return firstRune >= 'A' && firstRune <= 'Z'
}

func hasPattern(lines []string, pattern string) bool {
	regex := regexp.MustCompile(pattern)
	for _, line := range lines {
		if regex.MatchString(line) {
			return true
		}
	}
	return false
}
