package audit

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ferg-cod3s/conexus/internal/observability"
	"github.com/ferg-cod3s/conexus/internal/security"
)

// fileOutput implements file-based audit logging with rotation
type fileOutput struct {
	config OutputConfig
	file   *os.File
	writer *bufio.Writer
	size   int64
	logger *observability.Logger
	mu     sync.Mutex
}

func newFileOutput(config OutputConfig) (*fileOutput, error) {
	// Validate and clean the file path to prevent path traversal
	cleanPath, err := security.ValidatePath(config.FilePath, "")
	if err != nil {
		return nil, fmt.Errorf("invalid audit log file path: %w", err)
	}
	config.FilePath = cleanPath

	// Set defaults
	if config.MaxSize == 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB
	}
	if config.MaxBackups == 0 {
		config.MaxBackups = 10
	}
	if config.MaxAge == 0 {
		config.MaxAge = 30 // 30 days
	}

	output := &fileOutput{
		config: config,
		logger: observability.NewLogger(observability.LoggerConfig{
			Level:     "info",
			Format:    "text",
			Output:    os.Stderr,
			AddSource: false,
		}),
	}

	// Open initial file
	if err := output.openFile(); err != nil {
		return nil, err
	}

	return output, nil
}

func (fo *fileOutput) Write(event AuditEvent) error {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	data, err := fo.formatEvent(event)
	if err != nil {
		return err
	}

	// Check if rotation is needed
	if fo.size+int64(len(data)) > fo.config.MaxSize {
		if err := fo.rotate(); err != nil {
			return err
		}
	}

	// Write data
	n, err := fo.writer.Write(data)
	if err != nil {
		return err
	}

	fo.size += int64(n)

	// Flush immediately for audit logs
	return fo.writer.Flush()
}

func (fo *fileOutput) Close() error {
	fo.mu.Lock()
	defer fo.mu.Unlock()

	if fo.writer != nil {
		fo.writer.Flush()
	}
	if fo.file != nil {
		return fo.file.Close()
	}
	return nil
}

func (fo *fileOutput) openFile() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(fo.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// #nosec G304 - FilePath validated in newFileOutput constructor at line 31
	file, err := os.OpenFile(fo.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	fo.file = file
	fo.writer = bufio.NewWriter(file)

	// Get current file size
	if stat, err := file.Stat(); err == nil {
		fo.size = stat.Size()
	}

	return nil
}

func (fo *fileOutput) rotate() error {
	// Close current file
	if fo.writer != nil {
		fo.writer.Flush()
	}
	if fo.file != nil {
		fo.file.Close()
	}

	// Rotate existing files
	if err := fo.rotateFiles(); err != nil {
		return err
	}

	// Open new file
	return fo.openFile()
}

func (fo *fileOutput) rotateFiles() error {
	// Remove oldest file if we have too many backups
	maxBackup := fmt.Sprintf("%s.%d", fo.config.FilePath, fo.config.MaxBackups)
	if fo.config.Compress && fo.config.MaxBackups > 0 {
		maxBackup += ".gz"
	}
	os.Remove(maxBackup)

	// Rotate existing backups
	for i := fo.config.MaxBackups - 1; i > 0; i-- {
		src := fmt.Sprintf("%s.%d", fo.config.FilePath, i)
		dst := fmt.Sprintf("%s.%d", fo.config.FilePath, i+1)

		if fo.config.Compress {
			src += ".gz"
			dst += ".gz"
		}

		os.Rename(src, dst)
	}

	// Move current file to .1
	current := fo.config.FilePath
	backup := fmt.Sprintf("%s.1", fo.config.FilePath)

	if fo.config.Compress {
		// Compress the current file
		if err := fo.compressFile(current, backup+".gz"); err != nil {
			return err
		}
		os.Remove(current)
	} else {
		os.Rename(current, backup)
	}

	// Clean up old files based on age
	fo.cleanupOldFiles()

	return nil
}

func (fo *fileOutput) compressFile(src, dst string) error {
	// Validate paths to prevent traversal attacks
	safeSrc, err := security.ValidatePath(src, "")
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	safeDst, err := security.ValidatePath(dst, "")
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// #nosec G304 - Paths validated above with security.ValidatePath
	srcFile, err := os.Open(safeSrc)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// #nosec G304 - Paths validated above with security.ValidatePath
	dstFile, err := os.Create(safeDst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	gzipWriter := gzip.NewWriter(dstFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, srcFile)
	return err
}

func (fo *fileOutput) cleanupOldFiles() error {
	pattern := fo.config.FilePath + ".*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -fo.config.MaxAge)

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(match)
		}
	}

	return nil
}

func (fo *fileOutput) formatEvent(event AuditEvent) ([]byte, error) {
	var data []byte
	var err error

	if fo.config.Format == "json" {
		data, err = json.Marshal(event)
		if err != nil {
			return nil, err
		}
		data = append(data, '\n')
	} else {
		// Text format
		data = []byte(fmt.Sprintf("[%s] %s %s %s %s\n",
			event.Timestamp.Format(time.RFC3339),
			event.EventType,
			event.Category,
			event.Outcome,
			event.UserID,
		))
	}

	return data, nil
}

// syslogOutput implements syslog-based audit logging
type syslogOutput struct {
	config OutputConfig
	conn   net.Conn
	logger *observability.Logger
	mu     sync.Mutex
}

func newSyslogOutput(config OutputConfig) (*syslogOutput, error) {
	if config.SyslogAddr == "" {
		return nil, fmt.Errorf("syslog address is required")
	}
	if config.SyslogNetwork == "" {
		config.SyslogNetwork = "udp"
	}
	if config.SyslogTag == "" {
		config.SyslogTag = "conexus-audit"
	}

	output := &syslogOutput{
		config: config,
		logger: observability.NewLogger(observability.LoggerConfig{
			Level:     "info",
			Format:    "text",
			Output:    os.Stderr,
			AddSource: false,
		}),
	}

	// Connect to syslog server
	conn, err := net.Dial(config.SyslogNetwork, config.SyslogAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to syslog: %w", err)
	}

	output.conn = conn
	return output, nil
}

func (so *syslogOutput) Write(event AuditEvent) error {
	so.mu.Lock()
	defer so.mu.Unlock()

	data, err := so.formatEvent(event)
	if err != nil {
		return err
	}

	_, err = so.conn.Write(data)
	return err
}

func (so *syslogOutput) Close() error {
	so.mu.Lock()
	defer so.mu.Unlock()

	if so.conn != nil {
		return so.conn.Close()
	}
	return nil
}

func (so *syslogOutput) formatEvent(event AuditEvent) ([]byte, error) {
	// RFC 5424 syslog format
	timestamp := event.Timestamp.Format(time.RFC3339)
	hostname, _ := os.Hostname()

	var msg string
	if so.config.Format == "json" {
		jsonData, err := json.Marshal(event)
		if err != nil {
			return nil, err
		}
		msg = string(jsonData)
	} else {
		msg = fmt.Sprintf("%s %s %s %s", event.EventType, event.Category, event.Outcome, event.UserID)
	}

	// <priority>version timestamp hostname app-name procid msgid [structured-data] msg
	syslogMsg := fmt.Sprintf("<134>1 %s %s %s - - - %s\n",
		timestamp, hostname, so.config.SyslogTag, msg)

	return []byte(syslogMsg), nil
}

// externalOutput implements external audit logging (HTTP forwarder)
type externalOutput struct {
	config OutputConfig
	logger *observability.Logger
}

func newExternalOutput(config OutputConfig) (*externalOutput, error) {
	if config.ExternalURL == "" {
		return nil, fmt.Errorf("external URL is required")
	}

	return &externalOutput{
		config: config,
		logger: observability.NewLogger(observability.LoggerConfig{
			Level:     "info",
			Format:    "text",
			Output:    os.Stderr,
			AddSource: false,
		}),
	}, nil
}

func (eo *externalOutput) Write(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// For now, just log that we would send to external system
	// In a real implementation, this would make HTTP requests
	eo.logger.Info("External audit event",
		"url", eo.config.ExternalURL,
		"event_type", event.EventType,
		"size", len(data))

	return nil
}

func (eo *externalOutput) Close() error {
	return nil
}

// stdOutput implements stdout/stderr audit logging
type stdOutput struct {
	config OutputConfig
	writer io.Writer
	logger *observability.Logger
}

func newStdOutput(config OutputConfig, writer io.Writer) (*stdOutput, error) {
	return &stdOutput{
		config: config,
		writer: writer,
		logger: observability.NewLogger(observability.LoggerConfig{
			Level:     "info",
			Format:    "text",
			Output:    os.Stderr,
			AddSource: false,
		}),
	}, nil
}

func (so *stdOutput) Write(event AuditEvent) error {
	data, err := so.formatEvent(event)
	if err != nil {
		return err
	}

	_, err = so.writer.Write(data)
	return err
}

func (so *stdOutput) Close() error {
	return nil
}

func (so *stdOutput) formatEvent(event AuditEvent) ([]byte, error) {
	var data []byte
	var err error

	if so.config.Format == "json" {
		data, err = json.Marshal(event)
		if err != nil {
			return nil, err
		}
		data = append(data, '\n')
	} else {
		// Text format
		data = []byte(fmt.Sprintf("[%s] %s %s %s %s\n",
			event.Timestamp.Format(time.RFC3339),
			event.EventType,
			event.Category,
			event.Outcome,
			event.UserID,
		))
	}

	return data, nil
}
