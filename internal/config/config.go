// Package config provides configuration management for Conexus.
// It supports loading configuration from environment variables, files (YAML/JSON),
// and defaults, with a clear precedence order: env > file > defaults.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/security"
	"github.com/ferg-cod3s/conexus/internal/validation"
	"gopkg.in/yaml.v3"
)

// Config represents the complete Conexus configuration.
type Config struct {
	Server        ServerConfig        `json:"server" yaml:"server"`
	Database      DatabaseConfig      `json:"database" yaml:"database"`
	Indexer       IndexerConfig       `json:"indexer" yaml:"indexer"`
	Logging       LoggingConfig       `json:"logging" yaml:"logging"`
	Observability ObservabilityConfig `json:"observability" yaml:"observability"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
}

// DatabaseConfig holds database configuration.
type DatabaseConfig struct {
	Path string `json:"path" yaml:"path"`
}

// IndexerConfig holds indexer configuration.
type IndexerConfig struct {
	RootPath     string `json:"root_path" yaml:"root_path"`
	ChunkSize    int    `json:"chunk_size" yaml:"chunk_size"`
	ChunkOverlap int    `json:"chunk_overlap" yaml:"chunk_overlap"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

// ObservabilityConfig holds observability configuration.
type ObservabilityConfig struct {
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`
	Tracing TracingConfig `json:"tracing" yaml:"tracing"`
	Sentry  SentryConfig  `json:"sentry" yaml:"sentry"`
}

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Port    int    `json:"port" yaml:"port"`
	Path    string `json:"path" yaml:"path"`
}

// TracingConfig holds tracing configuration.
type TracingConfig struct {
	Enabled    bool    `json:"enabled" yaml:"enabled"`
	Endpoint   string  `json:"endpoint" yaml:"endpoint"`
	SampleRate float64 `json:"sample_rate" yaml:"sample_rate"`
}

// SentryConfig holds Sentry error monitoring configuration.
type SentryConfig struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	DSN         string  `json:"dsn" yaml:"dsn"`
	Environment string  `json:"environment" yaml:"environment"`
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate"`
	Release     string  `json:"release" yaml:"release"`
}

// Default values
const (
	DefaultHost             = "0.0.0.0"
	DefaultPort             = 8080
	DefaultDBPath           = "./data/conexus.db"
	DefaultRootPath         = "."
	DefaultChunkSize        = 512
	DefaultChunkOverlap     = 50
	DefaultLogLevel         = "info"
	DefaultLogFormat        = "json"
	DefaultMetricsEnabled   = false
	DefaultMetricsPort      = 9091
	DefaultMetricsPath      = "/metrics"
	DefaultTracingEnabled   = false
	DefaultTracingEndpoint  = "http://localhost:4318"
	DefaultSampleRate       = 0.1
	DefaultSentryEnabled    = false
	DefaultSentryDSN        = ""
	DefaultSentryEnv        = "development"
	DefaultSentrySampleRate = 1.0
	DefaultSentryRelease    = "0.1.0-alpha"
)

// Valid values for validation
var (
	ValidLogLevels  = []string{"debug", "info", "warn", "error"}
	ValidLogFormats = []string{"json", "text"}
)

// Load loads configuration from environment variables and optional config file.
// Precedence: env vars > config file > defaults.
func Load(ctx context.Context) (*Config, error) {
	// Start with defaults
	cfg := defaults()

	// Load from config file if specified
	if configFile := os.Getenv("CONEXUS_CONFIG_FILE"); configFile != "" {
		// Validate config file path to prevent path traversal
		if _, err := validation.ValidateConfigPath(configFile); err != nil {
			return nil, fmt.Errorf("config file path validation failed: %w", err)
		}

		fileCfg, err := loadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
		cfg = merge(cfg, fileCfg)
	}

	// Override with environment variables
	cfg = loadEnv(cfg)

	// Validate final configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// defaults returns a Config with all default values.
func defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Host: DefaultHost,
			Port: DefaultPort,
		},
		Database: DatabaseConfig{
			Path: DefaultDBPath,
		},
		Indexer: IndexerConfig{
			RootPath:     DefaultRootPath,
			ChunkSize:    DefaultChunkSize,
			ChunkOverlap: DefaultChunkOverlap,
		},
		Logging: LoggingConfig{
			Level:  DefaultLogLevel,
			Format: DefaultLogFormat,
		},
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled: DefaultMetricsEnabled,
				Port:    DefaultMetricsPort,
				Path:    DefaultMetricsPath,
			},
			Tracing: TracingConfig{
				Enabled:    DefaultTracingEnabled,
				Endpoint:   DefaultTracingEndpoint,
				SampleRate: DefaultSampleRate,
			},
			Sentry: SentryConfig{
				Enabled:     DefaultSentryEnabled,
				DSN:         DefaultSentryDSN,
				Environment: DefaultSentryEnv,
				SampleRate:  DefaultSentrySampleRate,
				Release:     DefaultSentryRelease,
			},
		},
	}
}

// loadFile loads configuration from a YAML or JSON file.
func loadFile(path string) (*Config, error) {
	// Validate and sanitize path to prevent traversal attacks (G304)
	safePath, err := security.ValidatePath(path, "")
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	// #nosec G304 - Path validated at line 165 with ValidatePath
	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	cfg := &Config{}
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return cfg, nil
}

// loadEnv loads configuration from environment variables.
// Only overrides non-zero values from the provided config.
func loadEnv(cfg *Config) *Config {
	// Server config
	if host := os.Getenv("CONEXUS_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("CONEXUS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if stdio := os.Getenv("CONEXUS_STDIO"); stdio != "" {
		if enabled, err := strconv.ParseBool(stdio); err == nil && enabled {
			cfg.Server.Port = 0
		}
	}

	// Database config
	if dbPath := os.Getenv("CONEXUS_DB_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}

	// Indexer config
	if rootPath := os.Getenv("CONEXUS_ROOT_PATH"); rootPath != "" {
		cfg.Indexer.RootPath = rootPath
	}
	if chunkSize := os.Getenv("CONEXUS_CHUNK_SIZE"); chunkSize != "" {
		if cs, err := strconv.Atoi(chunkSize); err == nil {
			cfg.Indexer.ChunkSize = cs
		}
	}
	if chunkOverlap := os.Getenv("CONEXUS_CHUNK_OVERLAP"); chunkOverlap != "" {
		if co, err := strconv.Atoi(chunkOverlap); err == nil {
			cfg.Indexer.ChunkOverlap = co
		}
	}

	// Logging config
	if logLevel := os.Getenv("CONEXUS_LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if logFormat := os.Getenv("CONEXUS_LOG_FORMAT"); logFormat != "" {
		cfg.Logging.Format = logFormat
	}

	// Metrics config
	if metricsEnabled := os.Getenv("CONEXUS_METRICS_ENABLED"); metricsEnabled != "" {
		if enabled, err := strconv.ParseBool(metricsEnabled); err == nil {
			cfg.Observability.Metrics.Enabled = enabled
		}
	}
	if metricsPort := os.Getenv("CONEXUS_METRICS_PORT"); metricsPort != "" {
		if port, err := strconv.Atoi(metricsPort); err == nil {
			cfg.Observability.Metrics.Port = port
		}
	}
	if metricsPath := os.Getenv("CONEXUS_METRICS_PATH"); metricsPath != "" {
		cfg.Observability.Metrics.Path = metricsPath
	}

	// Tracing config
	if tracingEnabled := os.Getenv("CONEXUS_TRACING_ENABLED"); tracingEnabled != "" {
		if enabled, err := strconv.ParseBool(tracingEnabled); err == nil {
			cfg.Observability.Tracing.Enabled = enabled
		}
	}
	if tracingEndpoint := os.Getenv("CONEXUS_TRACING_ENDPOINT"); tracingEndpoint != "" {
		cfg.Observability.Tracing.Endpoint = tracingEndpoint
	}
	if sampleRate := os.Getenv("CONEXUS_TRACING_SAMPLE_RATE"); sampleRate != "" {
		if rate, err := strconv.ParseFloat(sampleRate, 64); err == nil {
			cfg.Observability.Tracing.SampleRate = rate
		}
	}

	// Sentry config
	if sentryEnabled := os.Getenv("CONEXUS_SENTRY_ENABLED"); sentryEnabled != "" {
		if enabled, err := strconv.ParseBool(sentryEnabled); err == nil {
			cfg.Observability.Sentry.Enabled = enabled
		}
	}
	if sentryDSN := os.Getenv("CONEXUS_SENTRY_DSN"); sentryDSN != "" {
		cfg.Observability.Sentry.DSN = sentryDSN
	}
	if sentryEnv := os.Getenv("CONEXUS_SENTRY_ENVIRONMENT"); sentryEnv != "" {
		cfg.Observability.Sentry.Environment = sentryEnv
	}
	if sentrySampleRate := os.Getenv("CONEXUS_SENTRY_SAMPLE_RATE"); sentrySampleRate != "" {
		if rate, err := strconv.ParseFloat(sentrySampleRate, 64); err == nil {
			cfg.Observability.Sentry.SampleRate = rate
		}
	}
	if sentryRelease := os.Getenv("CONEXUS_SENTRY_RELEASE"); sentryRelease != "" {
		cfg.Observability.Sentry.Release = sentryRelease
	}

	return cfg
}

// merge merges two configs, preferring values from 'override' when non-zero.
func merge(base, override *Config) *Config {
	result := *base

	// Server
	if override.Server.Host != "" {
		result.Server.Host = override.Server.Host
	}
	if override.Server.Port != 0 {
		result.Server.Port = override.Server.Port
	}

	// Database
	if override.Database.Path != "" {
		result.Database.Path = override.Database.Path
	}

	// Indexer
	if override.Indexer.RootPath != "" {
		result.Indexer.RootPath = override.Indexer.RootPath
	}
	if override.Indexer.ChunkSize != 0 {
		result.Indexer.ChunkSize = override.Indexer.ChunkSize
	}
	if override.Indexer.ChunkOverlap != 0 {
		result.Indexer.ChunkOverlap = override.Indexer.ChunkOverlap
	}

	// Logging
	if override.Logging.Level != "" {
		result.Logging.Level = override.Logging.Level
	}
	if override.Logging.Format != "" {
		result.Logging.Format = override.Logging.Format
	}

	// Observability - Metrics
	// For boolean flags, we need to check if they differ from defaults
	if override.Observability.Metrics.Enabled != DefaultMetricsEnabled {
		result.Observability.Metrics.Enabled = override.Observability.Metrics.Enabled
	}
	if override.Observability.Metrics.Port != 0 {
		result.Observability.Metrics.Port = override.Observability.Metrics.Port
	}
	if override.Observability.Metrics.Path != "" {
		result.Observability.Metrics.Path = override.Observability.Metrics.Path
	}

	// Observability - Tracing
	if override.Observability.Tracing.Enabled != DefaultTracingEnabled {
		result.Observability.Tracing.Enabled = override.Observability.Tracing.Enabled
	}
	if override.Observability.Tracing.Endpoint != "" {
		result.Observability.Tracing.Endpoint = override.Observability.Tracing.Endpoint
	}
	if override.Observability.Tracing.SampleRate != 0 {
		result.Observability.Tracing.SampleRate = override.Observability.Tracing.SampleRate
	}

	// Observability - Sentry
	if override.Observability.Sentry.Enabled != DefaultSentryEnabled {
		result.Observability.Sentry.Enabled = override.Observability.Sentry.Enabled
	}
	if override.Observability.Sentry.DSN != "" {
		result.Observability.Sentry.DSN = override.Observability.Sentry.DSN
	}
	if override.Observability.Sentry.Environment != "" {
		result.Observability.Sentry.Environment = override.Observability.Sentry.Environment
	}
	if override.Observability.Sentry.SampleRate != 0 {
		result.Observability.Sentry.SampleRate = override.Observability.Sentry.SampleRate
	}
	if override.Observability.Sentry.Release != "" {
		result.Observability.Sentry.Release = override.Observability.Sentry.Release
	}

	return &result
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Server.Port)
	}

	// Validate database config
	if c.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	// Validate indexer config
	if c.Indexer.RootPath == "" {
		return fmt.Errorf("indexer root path cannot be empty")
	}
	if c.Indexer.ChunkSize < 1 {
		return fmt.Errorf("chunk size must be positive: %d", c.Indexer.ChunkSize)
	}
	if c.Indexer.ChunkOverlap < 0 {
		return fmt.Errorf("chunk overlap cannot be negative: %d", c.Indexer.ChunkOverlap)
	}
	if c.Indexer.ChunkOverlap >= c.Indexer.ChunkSize {
		return fmt.Errorf("chunk overlap (%d) must be less than chunk size (%d)",
			c.Indexer.ChunkOverlap, c.Indexer.ChunkSize)
	}

	// Validate logging config
	if !contains(ValidLogLevels, c.Logging.Level) {
		return fmt.Errorf("invalid log level: %s (valid: %v)", c.Logging.Level, ValidLogLevels)
	}
	if !contains(ValidLogFormats, c.Logging.Format) {
		return fmt.Errorf("invalid log format: %s (valid: %v)", c.Logging.Format, ValidLogFormats)
	}

	// Validate metrics config
	if c.Observability.Metrics.Enabled {
		if c.Observability.Metrics.Port < 1 || c.Observability.Metrics.Port > 65535 {
			return fmt.Errorf("invalid metrics port: %d (must be 1-65535)", c.Observability.Metrics.Port)
		}
		if c.Observability.Metrics.Path == "" {
			return fmt.Errorf("metrics path cannot be empty when metrics enabled")
		}
	}

	// Validate tracing config
	if c.Observability.Tracing.Enabled {
		if c.Observability.Tracing.Endpoint == "" {
			return fmt.Errorf("tracing endpoint cannot be empty when tracing enabled")
		}
		if c.Observability.Tracing.SampleRate < 0 || c.Observability.Tracing.SampleRate > 1 {
			return fmt.Errorf("tracing sample rate must be between 0 and 1: %f", c.Observability.Tracing.SampleRate)
		}
	}

	// Validate sentry config
	if c.Observability.Sentry.Enabled {
		if c.Observability.Sentry.DSN == "" {
			return fmt.Errorf("sentry DSN cannot be empty when sentry enabled")
		}
		if c.Observability.Sentry.SampleRate < 0 || c.Observability.Sentry.SampleRate > 1 {
			return fmt.Errorf("sentry sample rate must be between 0 and 1: %f", c.Observability.Sentry.SampleRate)
		}
	}

	return nil
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Default returns a default configuration for testing and documentation.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host: DefaultHost,
			Port: DefaultPort,
		},
		Database: DatabaseConfig{
			Path: DefaultDBPath,
		},
		Indexer: IndexerConfig{
			RootPath:     DefaultRootPath,
			ChunkSize:    DefaultChunkSize,
			ChunkOverlap: DefaultChunkOverlap,
		},
		Logging: LoggingConfig{
			Level:  DefaultLogLevel,
			Format: DefaultLogFormat,
		},
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled: DefaultMetricsEnabled,
				Port:    DefaultMetricsPort,
				Path:    DefaultMetricsPath,
			},
			Tracing: TracingConfig{
				Enabled:    DefaultTracingEnabled,
				Endpoint:   DefaultTracingEndpoint,
				SampleRate: DefaultSampleRate,
			},
			Sentry: SentryConfig{
				Enabled:     DefaultSentryEnabled,
				DSN:         DefaultSentryDSN,
				Environment: DefaultSentryEnv,
				SampleRate:  DefaultSentrySampleRate,
				Release:     DefaultSentryRelease,
			},
		},
	}
}
