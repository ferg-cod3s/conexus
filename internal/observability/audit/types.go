package audit

import (
	"time"
)

// EventType represents the type of security event being logged
type EventType string

// Security event types
const (
	// Authentication events
	EventTypeAuthSuccess         EventType = "auth.success"
	EventTypeAuthFailure         EventType = "auth.failure"
	// nosemgrep: go-hardcoded-credentials
	EventTypeAuthTokenValidation EventType = "auth.token_validation" // Type constant, not a credential

	// Authorization events
	EventTypeAuthzSuccess EventType = "authz.success"
	EventTypeAuthzFailure EventType = "authz.failure"

	// MCP tool execution events
	EventTypeToolExecution EventType = "tool.execution"
	EventTypeToolFailure   EventType = "tool.failure"

	// Configuration events
	EventTypeConfigChange EventType = "config.change"
	EventTypeConfigRead   EventType = "config.read"

	// Rate limiting events
	EventTypeRateLimitHit EventType = "rate_limit.hit"

	// Security-relevant errors
	EventTypeSecurityError EventType = "security.error"

	// Administrative events
	EventTypeAdminLogin  EventType = "admin.login"
	EventTypeAdminAction EventType = "admin.action"
	EventTypeAdminConfig EventType = "admin.config"

	// Data access events
	EventTypeDataAccess   EventType = "data.access"
	EventTypeDataExport   EventType = "data.export"
	EventTypeDataDeletion EventType = "data.deletion"
)

// EventCategory represents the category of security event
type EventCategory string

// Security event categories
const (
	CategoryAuthentication EventCategory = "authentication"
	CategoryAuthorization  EventCategory = "authorization"
	CategoryAccess         EventCategory = "access"
	CategoryConfiguration  EventCategory = "configuration"
	CategoryData           EventCategory = "data"
	CategoryAdministrative EventCategory = "administrative"
	CategorySecurity       EventCategory = "security"
)

// Outcome represents the outcome of a security event
type Outcome string

// Security event outcomes
const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
	OutcomeUnknown Outcome = "unknown"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	// Core event information
	Timestamp time.Time     `json:"timestamp"`
	EventType EventType     `json:"event_type"`
	Category  EventCategory `json:"category"`
	Outcome   Outcome       `json:"outcome"`

	// Identity information
	UserID    string `json:"user_id,omitempty"`
	Username  string `json:"username,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	TokenID   string `json:"token_id,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`

	// Resource information
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	Action       string `json:"action,omitempty"`

	// Request/Operation details
	Method     string        `json:"method,omitempty"`
	Path       string        `json:"path,omitempty"`
	Parameters interface{}   `json:"parameters,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`

	// Additional context
	ErrorMessage string      `json:"error_message,omitempty"`
	ErrorCode    string      `json:"error_code,omitempty"`
	Details      interface{} `json:"details,omitempty"`

	// Compliance metadata
	ComplianceFlags []string `json:"compliance_flags,omitempty"`
	DataSensitivity string   `json:"data_sensitivity,omitempty"`

	// System information
	ServiceName    string `json:"service_name,omitempty"`
	ServiceVersion string `json:"service_version,omitempty"`
	Environment    string `json:"environment,omitempty"`
	Host           string `json:"host,omitempty"`
}

// AuditEventBuilder helps construct audit events with a fluent interface
type AuditEventBuilder struct {
	event AuditEvent
}

// NewAuditEventBuilder creates a new audit event builder
func NewAuditEventBuilder(eventType EventType, category EventCategory) *AuditEventBuilder {
	return &AuditEventBuilder{
		event: AuditEvent{
			Timestamp: time.Now().UTC(),
			EventType: eventType,
			Category:  category,
			Outcome:   OutcomeUnknown,
		},
	}
}

// WithOutcome sets the event outcome
func (b *AuditEventBuilder) WithOutcome(outcome Outcome) *AuditEventBuilder {
	b.event.Outcome = outcome
	return b
}

// WithUser sets user identity information
func (b *AuditEventBuilder) WithUser(userID, username, email string) *AuditEventBuilder {
	b.event.UserID = userID
	b.event.Username = username
	b.event.UserEmail = email
	return b
}

// WithSession sets session information
func (b *AuditEventBuilder) WithSession(sessionID, tokenID string) *AuditEventBuilder {
	b.event.SessionID = sessionID
	b.event.TokenID = tokenID
	return b
}

// WithRequest sets request information
func (b *AuditEventBuilder) WithRequest(ip, userAgent string) *AuditEventBuilder {
	b.event.IPAddress = ip
	b.event.UserAgent = userAgent
	return b
}

// WithResource sets resource information
func (b *AuditEventBuilder) WithResource(resourceType, resourceID, action string) *AuditEventBuilder {
	b.event.ResourceType = resourceType
	b.event.ResourceID = resourceID
	b.event.Action = action
	return b
}

// WithOperation sets operation details
func (b *AuditEventBuilder) WithOperation(method, path string, parameters interface{}) *AuditEventBuilder {
	b.event.Method = method
	b.event.Path = path
	b.event.Parameters = parameters
	return b
}

// WithDuration sets operation duration
func (b *AuditEventBuilder) WithDuration(duration time.Duration) *AuditEventBuilder {
	b.event.Duration = duration
	return b
}

// WithError sets error information
func (b *AuditEventBuilder) WithError(message, code string) *AuditEventBuilder {
	b.event.ErrorMessage = message
	b.event.ErrorCode = code
	return b
}

// WithDetails sets additional context details
func (b *AuditEventBuilder) WithDetails(details interface{}) *AuditEventBuilder {
	b.event.Details = details
	return b
}

// WithCompliance sets compliance metadata
func (b *AuditEventBuilder) WithCompliance(flags []string, sensitivity string) *AuditEventBuilder {
	b.event.ComplianceFlags = flags
	b.event.DataSensitivity = sensitivity
	return b
}

// WithSystem sets system information
func (b *AuditEventBuilder) WithSystem(serviceName, serviceVersion, environment, host string) *AuditEventBuilder {
	b.event.ServiceName = serviceName
	b.event.ServiceVersion = serviceVersion
	b.event.Environment = environment
	b.event.Host = host
	return b
}

// Build constructs the final audit event
func (b *AuditEventBuilder) Build() AuditEvent {
	return b.event
}
