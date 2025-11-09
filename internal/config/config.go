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
	Embedding     EmbeddingConfig     `json:"embedding" yaml:"embedding"`
	Logging       LoggingConfig       `json:"logging" yaml:"logging"`
	Observability ObservabilityConfig `json:"observability" yaml:"observability"`
	TLS           TLSConfig           `json:"tls" yaml:"tls"`
	Auth          AuthConfig          `json:"auth" yaml:"auth"`
	RateLimit     RateLimitConfig     `json:"rate_limit" yaml:"rate_limit"`
	Security      SecurityConfig      `json:"security" yaml:"security"`
	CORS          CORSConfig          `json:"cors" yaml:"cors"`
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

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	CertFile         string   `json:"cert_file" yaml:"cert_file"`
	KeyFile          string   `json:"key_file" yaml:"key_file"`
	AutoCert         bool     `json:"auto_cert" yaml:"auto_cert"`
	AutoCertDomains  []string `json:"auto_cert_domains" yaml:"auto_cert_domains"`
	AutoCertEmail    string   `json:"auto_cert_email" yaml:"auto_cert_email"`
	AutoCertCacheDir string   `json:"auto_cert_cache_dir" yaml:"auto_cert_cache_dir"`
	MinVersion       string   `json:"min_version" yaml:"min_version"`
	CipherSuites     []string `json:"cipher_suites" yaml:"cipher_suites"`
	CurvePreferences []string `json:"curve_preferences" yaml:"curve_preferences"`
	HTTPRedirectPort int      `json:"http_redirect_port" yaml:"http_redirect_port"`
}

// EmbeddingConfig holds embedding provider configuration.
type EmbeddingConfig struct {
	Provider   string                 `json:"provider" yaml:"provider"`
	Model      string                 `json:"model" yaml:"model"`
	Dimensions int                    `json:"dimensions" yaml:"dimensions"`
	Config     map[string]interface{} `json:"config" yaml:"config"`
}

// AuthConfig holds JWT authentication configuration.
type AuthConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	PrivateKey  string `json:"private_key" yaml:"private_key"`
	PublicKey   string `json:"public_key" yaml:"public_key"`
	Issuer      string `json:"issuer" yaml:"issuer"`
	Audience    string `json:"audience" yaml:"audience"`
	TokenExpiry int    `json:"token_expiry" yaml:"token_expiry"` // minutes
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled         bool        `json:"enabled" yaml:"enabled"`
	Algorithm       string      `json:"algorithm" yaml:"algorithm"` // token_bucket, sliding_window
	Redis           RedisConfig `json:"redis" yaml:"redis"`
	Default         LimitConfig `json:"default" yaml:"default"`
	Health          LimitConfig `json:"health" yaml:"health"`
	Webhook         LimitConfig `json:"webhook" yaml:"webhook"`
	Auth            LimitConfig `json:"auth" yaml:"auth"`
	BurstMultiplier float64     `json:"burst_multiplier" yaml:"burst_multiplier"`
	CleanupInterval string      `json:"cleanup_interval" yaml:"cleanup_interval"` // duration string
	SkipPaths       []string    `json:"skip_paths" yaml:"skip_paths"`
	SkipIPs         []string    `json:"skip_ips" yaml:"skip_ips"`
	TrustedProxies  []string    `json:"trusted_proxies" yaml:"trusted_proxies"`
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Addr      string `json:"addr" yaml:"addr"`
	Password  string `json:"password" yaml:"password"`
	DB        int    `json:"db" yaml:"db"`
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
}

// LimitConfig holds rate limit configuration for a specific endpoint type.
type LimitConfig struct {
	Requests          int    `json:"requests" yaml:"requests"`                       // requests per window
	Window            string `json:"window" yaml:"window"`                           // duration string (e.g., "1m", "1h")
	RequestsPerSecond int    `json:"requests_per_second" yaml:"requests_per_second"` // deprecated: use Requests instead
	Burst             int    `json:"burst" yaml:"burst"`                             // deprecated: use Window instead
}

// SecurityConfig holds security middleware configuration.
type SecurityConfig struct {
	Enabled         bool       `json:"enabled" yaml:"enabled"`
	CSP             CSPConfig  `json:"csp" yaml:"csp"`
	HSTS            HSTSConfig `json:"hsts" yaml:"hsts"`
	FrameOptions    string     `json:"frame_options" yaml:"frame_options"`
	ContentType     bool       `json:"content_type" yaml:"content_type"`
	BrowserXSSBlock bool       `json:"browser_xss_block" yaml:"browser_xss_block"`
	ReferrerPolicy  string     `json:"referrer_policy" yaml:"referrer_policy"`
}

// CSPConfig holds Content Security Policy configuration.
type CSPConfig struct {
	Enabled      bool     `json:"enabled" yaml:"enabled"`
	DefaultSrc   []string `json:"default_src" yaml:"default_src"`
	ScriptSrc    []string `json:"script_src" yaml:"script_src"`
	StyleSrc     []string `json:"style_src" yaml:"style_src"`
	ImgSrc       []string `json:"img_src" yaml:"img_src"`
	ConnectSrc   []string `json:"connect_src" yaml:"connect_src"`
	FontSrc      []string `json:"font_src" yaml:"font_src"`
	ObjectSrc    []string `json:"object_src" yaml:"object_src"`
	MediaSrc     []string `json:"media_src" yaml:"media_src"`
	FrameSrc     []string `json:"frame_src" yaml:"frame_src"`
	ReportURI    string   `json:"report_uri" yaml:"report_uri"`
	ReportOnly   bool     `json:"report_only" yaml:"report_only"`
	UpgradeInsec bool     `json:"upgrade_insecure" yaml:"upgrade_insecure"`
}

// HSTSConfig holds HTTP Strict Transport Security configuration.
type HSTSConfig struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	MaxAge            int  `json:"max_age" yaml:"max_age"`
	IncludeSubdomains bool `json:"include_subdomains" yaml:"include_subdomains"`
	Preload           bool `json:"preload" yaml:"preload"`
}

// CORSConfig holds Cross-Origin Resource Sharing configuration.
type CORSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
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

	// TLS defaults
	DefaultTLSEnabled          = false
	DefaultTLSMinVersion       = "1.2"
	DefaultTLSHTTPRedirectPort = 80
	DefaultAutoCertCacheDir    = "./data/autocert"

	// Embedding defaults
	DefaultEmbeddingProvider   = "mock"
	DefaultEmbeddingModel      = "mock-768"
	DefaultEmbeddingDimensions = 768

	// Auth defaults
	DefaultAuthEnabled = false
	DefaultTokenExpiry = 60 // minutes

	// Rate limit defaults
	DefaultRateLimitEnabled   = false
	DefaultRateLimitAlgorithm = "token_bucket"
	DefaultBurstMultiplier    = 2.0
	DefaultCleanupInterval    = "5m"
	DefaultRequestsPerSecond  = 10
	DefaultBurst              = 20

	// Security defaults
	DefaultSecurityEnabled = true
	DefaultFrameOptions    = "DENY"
	DefaultReferrerPolicy  = "strict-origin-when-cross-origin"
	DefaultHSTSMaxAge      = 31536000 // 1 year in seconds
	DefaultContentType     = true
	DefaultBrowserXSSBlock = true

	// CORS defaults
	DefaultCORSEnabled = false
	DefaultCORSMaxAge  = 86400 // 24 hours in seconds
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
		TLS: TLSConfig{
			Enabled:          DefaultTLSEnabled,
			MinVersion:       DefaultTLSMinVersion,
			HTTPRedirectPort: DefaultTLSHTTPRedirectPort,
			AutoCertCacheDir: DefaultAutoCertCacheDir,
		},
		Embedding: EmbeddingConfig{
			Provider:   DefaultEmbeddingProvider,
			Model:      DefaultEmbeddingModel,
			Dimensions: DefaultEmbeddingDimensions,
		},
		Auth: AuthConfig{
			Enabled:     DefaultAuthEnabled,
			TokenExpiry: DefaultTokenExpiry,
		},
		RateLimit: RateLimitConfig{
			Enabled:         DefaultRateLimitEnabled,
			Algorithm:       DefaultRateLimitAlgorithm,
			BurstMultiplier: DefaultBurstMultiplier,
			Redis: RedisConfig{
				Enabled: false,
			},
			Default: LimitConfig{
				RequestsPerSecond: DefaultRequestsPerSecond,
				Burst:             DefaultBurst,
			},
		},
		Security: SecurityConfig{
			Enabled:         DefaultSecurityEnabled,
			FrameOptions:    DefaultFrameOptions,
			ContentType:     DefaultContentType,
			BrowserXSSBlock: DefaultBrowserXSSBlock,
			ReferrerPolicy:  DefaultReferrerPolicy,
			CSP: CSPConfig{
				Enabled: false,
			},
			HSTS: HSTSConfig{
				MaxAge: DefaultHSTSMaxAge,
			},
		},
		CORS: CORSConfig{
			Enabled: DefaultCORSEnabled,
			MaxAge:  DefaultCORSMaxAge,
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

	// TLS config
	if tlsEnabled := os.Getenv("CONEXUS_TLS_ENABLED"); tlsEnabled != "" {
		if enabled, err := strconv.ParseBool(tlsEnabled); err == nil {
			cfg.TLS.Enabled = enabled
		}
	}
	if certFile := os.Getenv("CONEXUS_TLS_CERT_FILE"); certFile != "" {
		cfg.TLS.CertFile = certFile
	}
	if keyFile := os.Getenv("CONEXUS_TLS_KEY_FILE"); keyFile != "" {
		cfg.TLS.KeyFile = keyFile
	}
	if autoCert := os.Getenv("CONEXUS_TLS_AUTOCERT"); autoCert != "" {
		if enabled, err := strconv.ParseBool(autoCert); err == nil {
			cfg.TLS.AutoCert = enabled
		}
	}
	if autoCertDomains := os.Getenv("CONEXUS_TLS_AUTOCERT_DOMAINS"); autoCertDomains != "" {
		cfg.TLS.AutoCertDomains = strings.Split(autoCertDomains, ",")
	}
	if autoCertEmail := os.Getenv("CONEXUS_TLS_AUTOCERT_EMAIL"); autoCertEmail != "" {
		cfg.TLS.AutoCertEmail = autoCertEmail
	}
	if autoCertCacheDir := os.Getenv("CONEXUS_TLS_AUTOCERT_CACHE_DIR"); autoCertCacheDir != "" {
		cfg.TLS.AutoCertCacheDir = autoCertCacheDir
	}
	if minVersion := os.Getenv("CONEXUS_TLS_MIN_VERSION"); minVersion != "" {
		cfg.TLS.MinVersion = minVersion
	}
	if cipherSuites := os.Getenv("CONEXUS_TLS_CIPHER_SUITES"); cipherSuites != "" {
		cfg.TLS.CipherSuites = strings.Split(cipherSuites, ",")
	}
	if curvePreferences := os.Getenv("CONEXUS_TLS_CURVE_PREFERENCES"); curvePreferences != "" {
		cfg.TLS.CurvePreferences = strings.Split(curvePreferences, ",")
	}
	if httpRedirectPort := os.Getenv("CONEXUS_TLS_HTTP_REDIRECT_PORT"); httpRedirectPort != "" {
		if port, err := strconv.Atoi(httpRedirectPort); err == nil {
			cfg.TLS.HTTPRedirectPort = port
		}
	}

	// Embedding config
	if embeddingProvider := os.Getenv("CONEXUS_EMBEDDING_PROVIDER"); embeddingProvider != "" {
		cfg.Embedding.Provider = embeddingProvider
	}
	if embeddingModel := os.Getenv("CONEXUS_EMBEDDING_MODEL"); embeddingModel != "" {
		cfg.Embedding.Model = embeddingModel
	}
	if embeddingDimensions := os.Getenv("CONEXUS_EMBEDDING_DIMENSIONS"); embeddingDimensions != "" {
		if dims, err := strconv.Atoi(embeddingDimensions); err == nil {
			cfg.Embedding.Dimensions = dims
		}
	}

	// Auth config
	if authEnabled := os.Getenv("CONEXUS_AUTH_ENABLED"); authEnabled != "" {
		if enabled, err := strconv.ParseBool(authEnabled); err == nil {
			cfg.Auth.Enabled = enabled
		}
	}
	if tokenExpiry := os.Getenv("CONEXUS_AUTH_TOKEN_EXPIRY"); tokenExpiry != "" {
		if expiry, err := strconv.Atoi(tokenExpiry); err == nil {
			cfg.Auth.TokenExpiry = expiry
		}
	}

	// Rate limiting config
	if rateLimitEnabled := os.Getenv("CONEXUS_RATELIMIT_ENABLED"); rateLimitEnabled != "" {
		if enabled, err := strconv.ParseBool(rateLimitEnabled); err == nil {
			cfg.RateLimit.Enabled = enabled
		}
	}
	if rateLimitAlgorithm := os.Getenv("CONEXUS_RATELIMIT_ALGORITHM"); rateLimitAlgorithm != "" {
		cfg.RateLimit.Algorithm = rateLimitAlgorithm
	}
	if rateLimitRPS := os.Getenv("CONEXUS_RATELIMIT_REQUESTS_PER_SECOND"); rateLimitRPS != "" {
		if rps, err := strconv.Atoi(rateLimitRPS); err == nil {
			cfg.RateLimit.Default.RequestsPerSecond = rps
		}
	}
	if rateLimitBurst := os.Getenv("CONEXUS_RATELIMIT_BURST"); rateLimitBurst != "" {
		if burst, err := strconv.Atoi(rateLimitBurst); err == nil {
			cfg.RateLimit.Default.Burst = burst
		}
	}

	// Security config
	if securityEnabled := os.Getenv("CONEXUS_SECURITY_ENABLED"); securityEnabled != "" {
		if enabled, err := strconv.ParseBool(securityEnabled); err == nil {
			cfg.Security.Enabled = enabled
		}
	}
	if frameOptions := os.Getenv("CONEXUS_SECURITY_FRAME_OPTIONS"); frameOptions != "" {
		cfg.Security.FrameOptions = frameOptions
	}
	if cspEnabled := os.Getenv("CONEXUS_SECURITY_CSP_ENABLED"); cspEnabled != "" {
		if enabled, err := strconv.ParseBool(cspEnabled); err == nil {
			cfg.Security.CSP.Enabled = enabled
		}
	}
	if hstsMaxAge := os.Getenv("CONEXUS_SECURITY_HSTS_MAX_AGE"); hstsMaxAge != "" {
		if maxAge, err := strconv.Atoi(hstsMaxAge); err == nil {
			cfg.Security.HSTS.MaxAge = maxAge
		}
	}

	// CORS config
	if corsEnabled := os.Getenv("CONEXUS_CORS_ENABLED"); corsEnabled != "" {
		if enabled, err := strconv.ParseBool(corsEnabled); err == nil {
			cfg.CORS.Enabled = enabled
		}
	}
	if corsMaxAge := os.Getenv("CONEXUS_CORS_MAX_AGE"); corsMaxAge != "" {
		if maxAge, err := strconv.Atoi(corsMaxAge); err == nil {
			cfg.CORS.MaxAge = maxAge
		}
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

	// Merge TLS config
	if override.TLS.Enabled != DefaultTLSEnabled {
		result.TLS.Enabled = override.TLS.Enabled
	}
	if override.TLS.CertFile != "" {
		result.TLS.CertFile = override.TLS.CertFile
	}
	if override.TLS.KeyFile != "" {
		result.TLS.KeyFile = override.TLS.KeyFile
	}
	if override.TLS.AutoCert != DefaultTLSEnabled {
		result.TLS.AutoCert = override.TLS.AutoCert
	}
	if len(override.TLS.AutoCertDomains) > 0 {
		result.TLS.AutoCertDomains = override.TLS.AutoCertDomains
	}
	if override.TLS.AutoCertEmail != "" {
		result.TLS.AutoCertEmail = override.TLS.AutoCertEmail
	}
	if override.TLS.AutoCertCacheDir != DefaultAutoCertCacheDir {
		result.TLS.AutoCertCacheDir = override.TLS.AutoCertCacheDir
	}
	if override.TLS.MinVersion != DefaultTLSMinVersion {
		result.TLS.MinVersion = override.TLS.MinVersion
	}
	if len(override.TLS.CipherSuites) > 0 {
		result.TLS.CipherSuites = override.TLS.CipherSuites
	}
	if len(override.TLS.CurvePreferences) > 0 {
		result.TLS.CurvePreferences = override.TLS.CurvePreferences
	}
	if override.TLS.HTTPRedirectPort != DefaultTLSHTTPRedirectPort {
		result.TLS.HTTPRedirectPort = override.TLS.HTTPRedirectPort
	}

	// Embedding
	if override.Embedding.Provider != "" {
		result.Embedding.Provider = override.Embedding.Provider
	}
	if override.Embedding.Model != "" {
		result.Embedding.Model = override.Embedding.Model
	}
	if override.Embedding.Dimensions != 0 {
		result.Embedding.Dimensions = override.Embedding.Dimensions
	}

	// Auth
	if override.Auth.Enabled != DefaultAuthEnabled {
		result.Auth.Enabled = override.Auth.Enabled
	}
	if override.Auth.TokenExpiry != 0 {
		result.Auth.TokenExpiry = override.Auth.TokenExpiry
	}

	// RateLimit
	if override.RateLimit.Enabled != DefaultRateLimitEnabled {
		result.RateLimit.Enabled = override.RateLimit.Enabled
	}
	if override.RateLimit.Algorithm != "" {
		result.RateLimit.Algorithm = override.RateLimit.Algorithm
	}
	if override.RateLimit.BurstMultiplier != 0 {
		result.RateLimit.BurstMultiplier = override.RateLimit.BurstMultiplier
	}
	if override.RateLimit.Default.RequestsPerSecond != 0 {
		result.RateLimit.Default.RequestsPerSecond = override.RateLimit.Default.RequestsPerSecond
	}
	if override.RateLimit.Default.Burst != 0 {
		result.RateLimit.Default.Burst = override.RateLimit.Default.Burst
	}

	// Security
	if override.Security.Enabled != DefaultSecurityEnabled {
		result.Security.Enabled = override.Security.Enabled
	}
	if override.Security.FrameOptions != "" {
		result.Security.FrameOptions = override.Security.FrameOptions
	}
	if override.Security.ContentType != DefaultContentType {
		result.Security.ContentType = override.Security.ContentType
	}
	if override.Security.BrowserXSSBlock != DefaultBrowserXSSBlock {
		result.Security.BrowserXSSBlock = override.Security.BrowserXSSBlock
	}
	if override.Security.ReferrerPolicy != "" {
		result.Security.ReferrerPolicy = override.Security.ReferrerPolicy
	}
	if override.Security.HSTS.MaxAge != 0 {
		result.Security.HSTS.MaxAge = override.Security.HSTS.MaxAge
	}

	// CORS
	if override.CORS.Enabled != DefaultCORSEnabled {
		result.CORS.Enabled = override.CORS.Enabled
	}
	if override.CORS.MaxAge != 0 {
		result.CORS.MaxAge = override.CORS.MaxAge
	}
	if len(override.CORS.AllowedOrigins) > 0 {
		result.CORS.AllowedOrigins = override.CORS.AllowedOrigins
	}
	if len(override.CORS.AllowedMethods) > 0 {
		result.CORS.AllowedMethods = override.CORS.AllowedMethods
	}
	if len(override.CORS.AllowedHeaders) > 0 {
		result.CORS.AllowedHeaders = override.CORS.AllowedHeaders
	}
	if len(override.CORS.ExposedHeaders) > 0 {
		result.CORS.ExposedHeaders = override.CORS.ExposedHeaders
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

	// Validate TLS config
	if c.TLS.Enabled {
		// Manual TLS: requires cert and key files
		if !c.TLS.AutoCert {
			if c.TLS.CertFile == "" {
				return fmt.Errorf("TLS cert file cannot be empty when TLS is enabled")
			}
			if c.TLS.KeyFile == "" {
				return fmt.Errorf("TLS key file cannot be empty when TLS is enabled")
			}
		}

		// AutoCert: requires domains and email
		if c.TLS.AutoCert {
			if len(c.TLS.AutoCertDomains) == 0 {
				return fmt.Errorf("AutoCert domains cannot be empty when AutoCert is enabled")
			}
			if c.TLS.AutoCertEmail == "" {
				return fmt.Errorf("AutoCert email cannot be empty when AutoCert is enabled")
			}
		}

		// Validate MinVersion
		validVersions := []string{"1.0", "1.1", "1.2", "1.3"}
		if !contains(validVersions, c.TLS.MinVersion) {
			return fmt.Errorf("invalid TLS min version: %s (valid: %v)", c.TLS.MinVersion, validVersions)
		}

		// Validate HTTPRedirectPort
		if c.TLS.HTTPRedirectPort < 1 || c.TLS.HTTPRedirectPort > 65535 {
			return fmt.Errorf("invalid TLS HTTP redirect port: %d (must be 1-65535)", c.TLS.HTTPRedirectPort)
		}
	}

	// Validate embedding config
	validProviders := []string{"mock", "anthropic", "openai", "cohere"}
	if !contains(validProviders, c.Embedding.Provider) {
		return fmt.Errorf("invalid embedding provider: %s (valid: %v)", c.Embedding.Provider, validProviders)
	}
	if c.Embedding.Model == "" {
		return fmt.Errorf("embedding model cannot be empty")
	}
	if c.Embedding.Dimensions < 1 {
		return fmt.Errorf("embedding dimensions must be positive: %d", c.Embedding.Dimensions)
	}

	// Validate auth config
	if c.Auth.Enabled {
		if c.Auth.TokenExpiry < 1 {
			return fmt.Errorf("auth token expiry must be positive: %d", c.Auth.TokenExpiry)
		}
	}

	// Validate rate limit config
	if c.RateLimit.Enabled {
		validAlgorithms := []string{"token_bucket", "sliding_window", "fixed_window"}
		if !contains(validAlgorithms, c.RateLimit.Algorithm) {
			return fmt.Errorf("invalid rate limit algorithm: %s (valid: %v)", c.RateLimit.Algorithm, validAlgorithms)
		}
		if c.RateLimit.BurstMultiplier < 1.0 {
			return fmt.Errorf("rate limit burst multiplier must be >= 1.0: %f", c.RateLimit.BurstMultiplier)
		}
		// Validate default limits
		if c.RateLimit.Default.RequestsPerSecond < 0 {
			return fmt.Errorf("default rate limit requests per second cannot be negative: %d", c.RateLimit.Default.RequestsPerSecond)
		}
		if c.RateLimit.Default.Burst < 0 {
			return fmt.Errorf("default rate limit burst cannot be negative: %d", c.RateLimit.Default.Burst)
		}
	}

	// Validate security config
	if c.Security.Enabled {
		validFrameOptions := []string{"DENY", "SAMEORIGIN"}
		if !contains(validFrameOptions, c.Security.FrameOptions) {
			return fmt.Errorf("invalid security frame options: %s (valid: %v)", c.Security.FrameOptions, validFrameOptions)
		}
		validReferrerPolicies := []string{"no-referrer", "no-referrer-when-downgrade", "origin", "origin-when-cross-origin", "same-origin", "strict-origin", "strict-origin-when-cross-origin", "unsafe-url"}
		if !contains(validReferrerPolicies, c.Security.ReferrerPolicy) {
			return fmt.Errorf("invalid security referrer policy: %s (valid: %v)", c.Security.ReferrerPolicy, validReferrerPolicies)
		}
		if c.Security.HSTS.MaxAge < 0 {
			return fmt.Errorf("security HSTS max age cannot be negative: %d", c.Security.HSTS.MaxAge)
		}
	}

	// Validate CORS config
	if c.CORS.Enabled {
		if c.CORS.MaxAge < 0 {
			return fmt.Errorf("CORS max age cannot be negative: %d", c.CORS.MaxAge)
		}
		if len(c.CORS.AllowedOrigins) == 0 {
			return fmt.Errorf("CORS allowed origins cannot be empty when CORS is enabled")
		}
		if len(c.CORS.AllowedMethods) == 0 {
			return fmt.Errorf("CORS allowed methods cannot be empty when CORS is enabled")
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
		Embedding: EmbeddingConfig{
			Provider:   DefaultEmbeddingProvider,
			Model:      DefaultEmbeddingModel,
			Dimensions: DefaultEmbeddingDimensions,
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
		TLS: TLSConfig{
			Enabled:          DefaultTLSEnabled,
			MinVersion:       DefaultTLSMinVersion,
			HTTPRedirectPort: DefaultTLSHTTPRedirectPort,
			AutoCertCacheDir: DefaultAutoCertCacheDir,
		},
		Auth: AuthConfig{
			Enabled:     DefaultAuthEnabled,
			TokenExpiry: DefaultTokenExpiry,
		},
		RateLimit: RateLimitConfig{
			Enabled:   DefaultRateLimitEnabled,
			Algorithm: DefaultRateLimitAlgorithm,
			Default: LimitConfig{
				RequestsPerSecond: DefaultRequestsPerSecond,
				Burst:             DefaultBurst,
			},
			BurstMultiplier: DefaultBurstMultiplier,
		},
		Security: SecurityConfig{
			Enabled:         DefaultSecurityEnabled,
			FrameOptions:    DefaultFrameOptions,
			ContentType:     DefaultContentType,
			BrowserXSSBlock: DefaultBrowserXSSBlock,
			ReferrerPolicy:  DefaultReferrerPolicy,
			HSTS: HSTSConfig{
				MaxAge: DefaultHSTSMaxAge,
			},
		},
		CORS: CORSConfig{
			Enabled: DefaultCORSEnabled,
			MaxAge:  DefaultCORSMaxAge,
		},
	}
}
