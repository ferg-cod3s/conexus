package schema

import "time"

// AGENT_OUTPUT_V1 defines the standardized output format for all Conexus agents.
// Every agent must produce outputs conforming to this schema with 100% evidence backing.
type AgentOutputV1 struct {
	Version            string                `json:"version"`             // Must be "AGENT_OUTPUT_V1"
	ComponentName      string                `json:"component_name"`      // User-supplied or inferred label
	ScopeDescription   string                `json:"scope_description"`   // Concise scope definition
	Overview           string                `json:"overview"`            // 2-4 sentence HOW summary
	EntryPoints        []EntryPoint          `json:"entry_points"`        // Entry point identification
	CallGraph          []CallGraphEdge       `json:"call_graph"`          // Execution flow graph
	DataFlow           DataFlow              `json:"data_flow"`           // Data transformation pipeline
	StateManagement    []StateOperation      `json:"state_management"`    // Persistence operations
	SideEffects        []SideEffect          `json:"side_effects"`        // External interactions
	ErrorHandling      []ErrorHandler        `json:"error_handling"`      // Error handling paths
	Configuration      []ConfigInfluence     `json:"configuration"`       // Configuration influence
	Patterns           []Pattern             `json:"patterns"`            // Design patterns (descriptive only)
	Concurrency        []ConcurrencyMechanism `json:"concurrency"`        // Concurrency mechanisms
	ExternalDependencies []ExternalDependency `json:"external_dependencies"` // External dependencies
	Limitations        []string              `json:"limitations"`         // Transparency requirements
	OpenQuestions      []string              `json:"open_questions"`      // Areas needing clarification
	RawEvidence        []Evidence            `json:"raw_evidence"`        // Evidence traceability (MANDATORY)
}

// EntryPoint represents a function or symbol that serves as an entry into the component
type EntryPoint struct {
	File   string `json:"file"`   // Absolute path
	Lines  string `json:"lines"`  // Line range (e.g., "24-31")
	Symbol string `json:"symbol"` // Function or export name
	Role   string `json:"role"`   // handler|service|utility|etc
}

// CallGraphEdge represents a function call relationship
type CallGraphEdge struct {
	From    string `json:"from"`     // Source location (file.go:funcA)
	To      string `json:"to"`       // Target location (other.go:funcB)
	ViaLine int    `json:"via_line"` // Line number of call
}

// DataFlow describes how data transforms through the system
type DataFlow struct {
	Inputs          []DataPoint        `json:"inputs"`
	Transformations []Transformation   `json:"transformations"`
	Outputs         []DataPoint        `json:"outputs"`
}

// DataPoint represents a data input or output
type DataPoint struct {
	Source      string `json:"source"`      // file.go:line
	Name        string `json:"name"`        // Variable name
	Type        string `json:"type"`        // Inferred/simple type
	Description string `json:"description"` // Purpose
}

// Transformation represents a data transformation operation
type Transformation struct {
	File        string `json:"file"`
	Lines       string `json:"lines"`
	Operation   string `json:"operation"`    // parse|validate|map|filter|aggregate|serialize
	Description string `json:"description"`  // What changes
	BeforeShape string `json:"before_shape,omitempty"` // Optional
	AfterShape  string `json:"after_shape,omitempty"`  // Optional
}

// StateOperation represents a persistence operation
type StateOperation struct {
	File        string `json:"file"`
	Lines       string `json:"lines"`
	Kind        string `json:"kind"`        // db|cache|memory|fs
	Operation   string `json:"operation"`   // read|write|update|delete
	Entity      string `json:"entity"`      // table|collection|key
	Description string `json:"description"`
}

// SideEffect represents an external interaction
type SideEffect struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Type        string `json:"type"`        // log|metric|emit|publish|http|fs
	Description string `json:"description"`
}

// ErrorHandler represents error handling logic
type ErrorHandler struct {
	File      string `json:"file"`
	Lines     string `json:"lines"`
	Type      string `json:"type"`      // throw|catch|guard|retry
	Condition string `json:"condition"` // Expression or pattern
	Effect    string `json:"effect"`    // propagate|fallback|retry
}

// ConfigInfluence represents how configuration affects behavior
type ConfigInfluence struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Kind      string `json:"kind"`      // env|flag|configObject
	Name      string `json:"name"`      // CONFIG_NAME
	Influence string `json:"influence"` // Description of impact
}

// Pattern represents a design pattern usage
type Pattern struct {
	Name        string `json:"name"`        // Factory|Observer|etc
	File        string `json:"file"`
	Lines       string `json:"lines"`
	Description string `json:"description"` // Existing usage only
}

// ConcurrencyMechanism represents concurrent execution patterns
type ConcurrencyMechanism struct {
	File        string `json:"file"`
	Lines       string `json:"lines"`
	Mechanism   string `json:"mechanism"`   // goroutine|channel|mutex|waitgroup
	Description string `json:"description"`
}

// ExternalDependency represents external module usage
type ExternalDependency struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Module  string `json:"module"`  // Package or internal boundary
	Purpose string `json:"purpose"`
}

// Evidence provides file:line backing for claims (MANDATORY)
type Evidence struct {
	Claim string `json:"claim"` // The claim being evidenced
	File  string `json:"file"`  // Absolute path
	Lines string `json:"lines"` // Line range
}

// AgentRequest defines the input format for agent invocation
type AgentRequest struct {
	RequestID   string            `json:"request_id"`
	AgentID     string            `json:"agent_id"`
	Task        AgentTask         `json:"task"`
	Context     ConversationContext `json:"context"`
	Permissions Permissions       `json:"permissions"`
	Timestamp   time.Time         `json:"timestamp"`
}

// AgentTask defines the specific task for an agent
type AgentTask struct {
	TargetAgent       string   `json:"target_agent"`
	Files             []string `json:"files"`
	EntrySymbols      []string `json:"entry_symbols,omitempty"`
	AllowedDirectories []string `json:"allowed_directories"`
	SpecificRequest   string   `json:"specific_request"`
}

// ConversationContext tracks the conversation history
type ConversationContext struct {
	UserRequest        string                `json:"user_request"`
	PreviousAgents     []string              `json:"previous_agents"`
	AccumulatedContext map[string]interface{} `json:"accumulated_context"`
}

// Permissions defines what an agent is allowed to do
type Permissions struct {
	AllowedDirectories []string `json:"allowed_directories"`
	ReadOnly           bool     `json:"read_only"`
	MaxFileSize        int64    `json:"max_file_size"`      // bytes
	MaxExecutionTime   int      `json:"max_execution_time"` // seconds
}

// AgentResponse defines the output format from agent execution
type AgentResponse struct {
	RequestID  string          `json:"request_id"`
	AgentID    string          `json:"agent_id"`
	Status     ResponseStatus  `json:"status"`
	Output     *AgentOutputV1  `json:"output,omitempty"`
	Escalation *Escalation     `json:"escalation,omitempty"`
	Error      *AgentError     `json:"error,omitempty"`
	Duration   time.Duration   `json:"duration"`
	Timestamp  time.Time       `json:"timestamp"`
}

// ResponseStatus indicates the outcome of agent execution
type ResponseStatus string

const (
	StatusComplete           ResponseStatus = "complete"
	StatusPartial            ResponseStatus = "partial"
	StatusEscalationRequired ResponseStatus = "escalation_required"
	StatusError              ResponseStatus = "error"
)

// Escalation indicates the agent needs assistance from another agent
type Escalation struct {
	Required     bool   `json:"required"`
	TargetAgent  string `json:"target_agent"`
	Reason       string `json:"reason"`
	RequiredInfo string `json:"required_info"`
}

// AgentError represents an error during agent execution
type AgentError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Recoverable bool  `json:"recoverable"`
	Details    string `json:"details,omitempty"`
}
