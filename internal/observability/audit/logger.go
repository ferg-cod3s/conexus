package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/observability"
)

// OutputType represents the type of audit log output
type OutputType string

// Audit log output types
const (
	OutputTypeFile     OutputType = "file"
	OutputTypeSyslog   OutputType = "syslog"
	OutputTypeExternal OutputType = "external"
	OutputTypeStdout   OutputType = "stdout"
	OutputTypeStderr   OutputType = "stderr"
)

// OutputConfig holds configuration for a single audit log output
type OutputConfig struct {
	Type OutputType `json:"type" yaml:"type"`

	// File output configuration
	FilePath   string `json:"file_path,omitempty" yaml:"file_path"`
	MaxSize    int64  `json:"max_size,omitempty" yaml:"max_size"`       // Max size in bytes (default: 100MB)
	MaxBackups int    `json:"max_backups,omitempty" yaml:"max_backups"` // Max backup files (default: 10)
	MaxAge     int    `json:"max_age,omitempty" yaml:"max_age"`         // Max age in days (default: 30)
	Compress   bool   `json:"compress,omitempty" yaml:"compress"`       // Compress rotated files

	// Syslog output configuration
	SyslogNetwork string `json:"syslog_network,omitempty" yaml:"syslog_network"` // "tcp", "udp", "unix"
	SyslogAddr    string `json:"syslog_addr,omitempty" yaml:"syslog_addr"`       // address for syslog server
	SyslogTag     string `json:"syslog_tag,omitempty" yaml:"syslog_tag"`         // syslog tag (default: "conexus-audit")

	// External output configuration (HTTP/Syslog forwarder)
	ExternalURL      string            `json:"external_url,omitempty" yaml:"external_url"`
	ExternalHeaders  map[string]string `json:"external_headers,omitempty" yaml:"external_headers"`
	ExternalTimeout  time.Duration     `json:"external_timeout,omitempty" yaml:"external_timeout"`
	ExternalInsecure bool              `json:"external_insecure,omitempty" yaml:"external_insecure"` // Skip TLS verification

	// Common configuration
	Format string `json:"format,omitempty" yaml:"format"` // "json" or "text" (default: "json")
	Level  string `json:"level,omitempty" yaml:"level"`   // minimum log level (default: "info")
}

// Config holds the complete audit logging configuration
type Config struct {
	Enabled bool           `json:"enabled" yaml:"enabled"`
	Outputs []OutputConfig `json:"outputs" yaml:"outputs"`

	// Integrity protection
	EnableIntegrity bool   `json:"enable_integrity" yaml:"enable_integrity"`
	IntegrityKey    string `json:"integrity_key,omitempty" yaml:"integrity_key"` // HMAC key for integrity

	// Performance tuning
	BufferSize    int           `json:"buffer_size,omitempty" yaml:"buffer_size"`       // Buffer size for async logging (default: 4096)
	FlushInterval time.Duration `json:"flush_interval,omitempty" yaml:"flush_interval"` // Flush interval (default: 1s)

	// Compliance settings
	GDPRCompliant    bool     `json:"gdpr_compliant" yaml:"gdpr_compliant"`       // Enable GDPR compliance features
	RetentionPeriod  int      `json:"retention_period" yaml:"retention_period"`   // Days to retain logs (default: 2555 for ~7 years)
	DataMinimization bool     `json:"data_minimization" yaml:"data_minimization"` // Minimize PII in logs
	SensitiveFields  []string `json:"sensitive_fields" yaml:"sensitive_fields"`   // Fields to mask/hash

	// Service information
	ServiceName    string `json:"service_name,omitempty" yaml:"service_name"`
	ServiceVersion string `json:"service_version,omitempty" yaml:"service_version"`
	Environment    string `json:"environment,omitempty" yaml:"environment"`
}

// Logger provides comprehensive audit logging functionality
type Logger struct {
	config    Config
	outputs   []outputWriter
	buffer    chan AuditEvent
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *observability.Logger
	mu        sync.RWMutex
	integrity *integrityChecker
}

// outputWriter interface for different audit log outputs
type outputWriter interface {
	Write(event AuditEvent) error
	Close() error
}

// NewLogger creates a new audit logger with the given configuration
func NewLogger(config Config, logger *observability.Logger) (*Logger, error) {
	if !config.Enabled {
		return &Logger{config: config}, nil
	}

	// Set defaults
	if config.BufferSize == 0 {
		config.BufferSize = 4096
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = time.Second
	}
	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 2555 // ~7 years for SOC 2 compliance
	}

	ctx, cancel := context.WithCancel(context.Background())

	auditLogger := &Logger{
		config: config,
		buffer: make(chan AuditEvent, config.BufferSize),
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}

	// Initialize integrity checker if enabled
	if config.EnableIntegrity {
		var err error
		auditLogger.integrity, err = newIntegrityChecker(config.IntegrityKey)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize integrity checker: %w", err)
		}
	}

	// Initialize outputs
	for i, outputConfig := range config.Outputs {
		writer, err := auditLogger.createOutputWriter(outputConfig)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create output writer %d: %w", i, err)
		}
		auditLogger.outputs = append(auditLogger.outputs, writer)
	}

	// Start the logging goroutine
	auditLogger.wg.Add(1)
	go auditLogger.processEvents()

	return auditLogger, nil
}

// Log records an audit event
func (l *Logger) Log(event AuditEvent) {
	if !l.config.Enabled {
		return
	}

	// Apply data minimization if enabled
	if l.config.DataMinimization {
		event = l.minimizeData(event)
	}

	// Add service information
	event.ServiceName = l.config.ServiceName
	event.ServiceVersion = l.config.ServiceVersion
	event.Environment = l.config.Environment

	// Add host information
	if event.Host == "" {
		if hostname, err := os.Hostname(); err == nil {
			event.Host = hostname
		}
	}

	// Apply integrity protection if enabled
	if l.integrity != nil {
		hash := l.integrity.generateHash(event)
		if event.Details == nil {
			event.Details = make(map[string]interface{})
		}
		if details, ok := event.Details.(map[string]interface{}); ok {
			details["integrity_hash"] = hash
			event.Details = details
		}
	}

	// Send to processing goroutine (non-blocking)
	select {
	case l.buffer <- event:
	default:
		// Buffer is full, log warning but don't block
		l.logger.Warn("Audit log buffer full, dropping event",
			"event_type", event.EventType,
			"category", event.Category)
	}
}

// LogAuthSuccess logs a successful authentication event
func (l *Logger) LogAuthSuccess(ctx context.Context, userID, username, method string, ipAddr string) {
	event := NewAuditEventBuilder(EventTypeAuthSuccess, CategoryAuthentication).
		WithOutcome(OutcomeSuccess).
		WithUser(userID, username, "").
		WithRequest(ipAddr, "").
		WithResource("authentication", "", method).
		WithSystem(l.config.ServiceName, l.config.ServiceVersion, l.config.Environment, "").
		Build()

	l.Log(event)
}

// LogAuthFailure logs a failed authentication event
func (l *Logger) LogAuthFailure(ctx context.Context, method string, ipAddr string, reason string) {
	event := NewAuditEventBuilder(EventTypeAuthFailure, CategoryAuthentication).
		WithOutcome(OutcomeFailure).
		WithRequest(ipAddr, "").
		WithResource("authentication", "", method).
		WithError(reason, "auth_failed").
		WithSystem(l.config.ServiceName, l.config.ServiceVersion, l.config.Environment, "").
		Build()

	l.Log(event)
}

// LogToolExecution logs an MCP tool execution event
func (l *Logger) LogToolExecution(ctx context.Context, toolName string, success bool, duration time.Duration, userID string) {
	outcome := OutcomeSuccess
	if !success {
		outcome = OutcomeFailure
	}

	event := NewAuditEventBuilder(EventTypeToolExecution, CategoryAccess).
		WithOutcome(outcome).
		WithUser(userID, "", "").
		WithResource("tool", toolName, "execute").
		WithDuration(duration).
		WithSystem(l.config.ServiceName, l.config.ServiceVersion, l.config.Environment, "").
		Build()

	l.Log(event)
}

// LogRateLimitHit logs a rate limit violation
func (l *Logger) LogRateLimitHit(ctx context.Context, ipAddr, path, method string, limit int64) {
	event := NewAuditEventBuilder(EventTypeRateLimitHit, CategorySecurity).
		WithOutcome(OutcomeFailure).
		WithRequest(ipAddr, "").
		WithOperation(method, path, nil).
		WithDetails(map[string]interface{}{
			"limit": limit,
		}).
		WithSystem(l.config.ServiceName, l.config.ServiceVersion, l.config.Environment, "").
		Build()

	l.Log(event)
}

// LogConfigChange logs a configuration change event
func (l *Logger) LogConfigChange(ctx context.Context, userID, action, resource string, oldValue, newValue interface{}) {
	event := NewAuditEventBuilder(EventTypeConfigChange, CategoryConfiguration).
		WithOutcome(OutcomeSuccess).
		WithUser(userID, "", "").
		WithResource("configuration", resource, action).
		WithDetails(map[string]interface{}{
			"old_value": l.maskSensitiveData(oldValue),
			"new_value": l.maskSensitiveData(newValue),
		}).
		WithSystem(l.config.ServiceName, l.config.ServiceVersion, l.config.Environment, "").
		Build()

	l.Log(event)
}

// Close gracefully shuts down the audit logger
func (l *Logger) Close() error {
	if !l.config.Enabled {
		return nil
	}

	// Signal shutdown
	l.cancel()

	// Wait for processing to complete
	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		l.logger.Warn("Audit logger shutdown timed out")
	}

	// Close all outputs
	var errs []error
	for _, output := range l.outputs {
		if err := output.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing audit outputs: %v", errs)
	}

	return nil
}

// processEvents processes audit events from the buffer
func (l *Logger) processEvents() {
	defer l.wg.Done()

	ticker := time.NewTicker(l.config.FlushInterval)
	defer ticker.Stop()

	var batch []AuditEvent

	for {
		select {
		case event := <-l.buffer:
			batch = append(batch, event)

			// Flush if batch is getting large
			if len(batch) >= l.config.BufferSize/4 {
				l.flushBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				l.flushBatch(batch)
				batch = batch[:0]
			}

		case <-l.ctx.Done():
			// Final flush
			if len(batch) > 0 {
				l.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch writes a batch of events to all outputs
func (l *Logger) flushBatch(events []AuditEvent) {
	for _, event := range events {
		for _, output := range l.outputs {
			if err := output.Write(event); err != nil {
				l.logger.Error("Failed to write audit event",
					"error", err,
					"event_type", event.EventType,
					"output_type", fmt.Sprintf("%T", output))
			}
		}
	}
}

// createOutputWriter creates an output writer based on configuration
func (l *Logger) createOutputWriter(config OutputConfig) (outputWriter, error) {
	switch config.Type {
	case OutputTypeFile:
		return newFileOutput(config)
	case OutputTypeSyslog:
		return newSyslogOutput(config)
	case OutputTypeExternal:
		return newExternalOutput(config)
	case OutputTypeStdout:
		return newStdOutput(config, os.Stdout)
	case OutputTypeStderr:
		return newStdOutput(config, os.Stderr)
	default:
		return nil, fmt.Errorf("unsupported output type: %s", config.Type)
	}
}

// minimizeData applies data minimization to reduce PII in audit logs
func (l *Logger) minimizeData(event AuditEvent) AuditEvent {
	// Hash sensitive fields instead of storing them in plain text
	if event.UserEmail != "" {
		event.UserEmail = l.hashValue(event.UserEmail)
	}
	if event.IPAddress != "" {
		event.IPAddress = l.maskIPAddress(event.IPAddress)
	}
	if event.SessionID != "" {
		event.SessionID = l.hashValue(event.SessionID)
	}

	// Remove or mask sensitive details
	if details, ok := event.Details.(map[string]interface{}); ok {
		for _, field := range l.config.SensitiveFields {
			if value, exists := details[field]; exists {
				details[field] = l.maskSensitiveData(value)
			}
		}
		event.Details = details
	}

	return event
}

// maskSensitiveData masks sensitive data based on its type
func (l *Logger) maskSensitiveData(value interface{}) interface{} {
	if value == nil {
		return value
	}

	switch v := value.(type) {
	case string:
		// Mask strings that look like secrets
		if l.isSensitiveString(v) {
			return l.hashValue(v)
		}
		return v
	case map[string]interface{}:
		// Recursively mask sensitive data in maps
		masked := make(map[string]interface{})
		for k, val := range v {
			masked[k] = l.maskSensitiveData(val)
		}
		return masked
	default:
		return value
	}
}

// isSensitiveString checks if a string contains sensitive information
func (l *Logger) isSensitiveString(s string) bool {
	// Check for common patterns
	sensitivePatterns := []string{
		"password", "secret", "token", "key", "auth",
		"bearer", "authorization", "cookie", "session",
	}

	sLower := strings.ToLower(s)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(sLower, pattern) {
			return true
		}
	}

	return false
}

// hashValue creates a SHA-256 hash of a value for privacy
func (l *Logger) hashValue(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

// maskIPAddress masks the last octet of an IPv4 address or compresses IPv6
func (l *Logger) maskIPAddress(ip string) string {
	if net.ParseIP(ip) == nil {
		return ip // Not a valid IP
	}

	if strings.Contains(ip, ":") {
		// IPv6 - compress but keep network portion
		return ip[:len(ip)-4] + ":xxxx"
	}

	// IPv4 - mask last octet
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		parts[3] = "xxx"
		return strings.Join(parts, ".")
	}

	return ip
}

// DefaultConfig returns a secure default audit configuration
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Outputs: []OutputConfig{
			{
				Type:       OutputTypeFile,
				FilePath:   "/var/log/conexus/audit.log",
				MaxSize:    100 * 1024 * 1024, // 100MB
				MaxBackups: 10,
				MaxAge:     30,
				Compress:   true,
				Format:     "json",
			},
		},
		EnableIntegrity:  true,
		BufferSize:       4096,
		FlushInterval:    time.Second,
		GDPRCompliant:    true,
		RetentionPeriod:  2555, // ~7 years
		DataMinimization: true,
		SensitiveFields:  []string{"password", "secret", "token", "key", "auth"},
		ServiceName:      "conexus",
		ServiceVersion:   "0.1.2-alpha",
		Environment:      "production",
	}
}
