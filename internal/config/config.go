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
	"time"

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
	Auth          AuthConfig          `json:"auth" yaml:"auth"`
	Security      SecurityConfig      `json:"security" yaml:"security"`
	CORS          CORSConfig          `json:"cors" yaml:"cors"`
	TLS           TLSConfig           `json:"tls" yaml:"tls"`
	RateLimit     RateLimitConfig     `json:"rate_limit" yaml:"rate_limit"`
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

// EmbeddingConfig holds embedding provider configuration.
type EmbeddingConfig struct {
	Provider   string                 `json:"provider" yaml:"provider"`
	Model      string                 `json:"model" yaml:"model"`
	Dimensions int                    `json:"dimensions" yaml:"dimensions"`
	Config     map[string]interface{} `json:"config" yaml:"config"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

// AuthConfig holds JWT authentication configuration.
type AuthConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Issuer      string `json:"issuer" yaml:"issuer"`
	Audience    string `json:"audience" yaml:"audience"`
	PublicKey   string `json:"public_key" yaml:"public_key"`
	PrivateKey  string `json:"private_key" yaml:"private_key"`
	TokenExpiry int    `json:"token_expiry" yaml:"token_expiry"` // in minutes
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

// SecurityConfig holds security headers configuration.
type SecurityConfig struct {
	CSP                 CSPConfig  `json:"csp" yaml:"csp"`
	HSTS                HSTSConfig `json:"hsts" yaml:"hsts"`
	XFrameOptions       string     `json:"x_frame_options" yaml:"x_frame_options"`
	XContentTypeOptions string     `json:"x_content_type_options" yaml:"x_content_type_options"`
	ReferrerPolicy      string     `json:"referrer_policy" yaml:"referrer_policy"`
	PermissionsPolicy   string     `json:"permissions_policy" yaml:"permissions_policy"`
}

// CSPConfig holds Content Security Policy configuration.
type CSPConfig struct {
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Default []string `json:"default" yaml:"default"`
	Script  []string `json:"script" yaml:"script"`
	Style   []string `json:"style" yaml:"style"`
	Image   []string `json:"image" yaml:"image"`
	Font    []string `json:"font" yaml:"font"`
	Connect []string `json:"connect" yaml:"connect"`
	Media   []string `json:"media" yaml:"media"`
	Object  []string `json:"object" yaml:"object"`
	Frame   []string `json:"frame" yaml:"frame"`
	Report  string   `json:"report" yaml:"report"`
}

// HSTSConfig holds HTTP Strict Transport Security configuration.
type HSTSConfig struct {
	Enabled           bool `json:"enabled" yaml:"enabled"`
	MaxAge            int  `json:"max_age" yaml:"max_age"`
	IncludeSubdomains bool `json:"include_subdomains" yaml:"include_subdomains"`
	Preload           bool `json:"preload" yaml:"preload"`
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// TLSConfig holds TLS/HTTPS configuration.
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

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled         bool                 `json:"enabled" yaml:"enabled"`
	Algorithm       string               `json:"algorithm" yaml:"algorithm"`
	Redis           RateLimitRedisConfig `json:"redis" yaml:"redis"`
	Default         RateLimitRuleConfig  `json:"default" yaml:"default"`
	Health          RateLimitRuleConfig  `json:"health" yaml:"health"`
	Webhook         RateLimitRuleConfig  `json:"webhook" yaml:"webhook"`
	Auth            RateLimitRuleConfig  `json:"auth" yaml:"auth"`
	BurstMultiplier float64              `json:"burst_multiplier" yaml:"burst_multiplier"`
	CleanupInterval time.Duration        `json:"cleanup_interval" yaml:"cleanup_interval"`
	SkipPaths       []string             `json:"skip_paths" yaml:"skip_paths"`
	SkipIPs         []string             `json:"skip_ips" yaml:"skip_ips"`
	TrustedProxies  []string             `json:"trusted_proxies" yaml:"trusted_proxies"`
}

// RateLimitRedisConfig holds Redis configuration for rate limiting.
type RateLimitRedisConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Addr      string `json:"addr" yaml:"addr"`
	Password  string `json:"password" yaml:"password"`
	DB        int    `json:"db" yaml:"db"`
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
}

// RateLimitRuleConfig holds rate limit configuration for a specific endpoint type.
type RateLimitRuleConfig struct {
	Requests int           `json:"requests" yaml:"requests"`
	Window   time.Duration `json:"window" yaml:"window"`
}

// Default values
const (
	DefaultHost                = "0.0.0.0"
	DefaultPort                = 0 // Default to stdio mode for MCP compatibility
	DefaultDBPath              = "./data/conexus.db"
	DefaultRootPath            = "."
	DefaultChunkSize           = 512
	DefaultChunkOverlap        = 50
	DefaultEmbeddingProvider   = "mock"
	DefaultEmbeddingModel      = "mock-768"
	DefaultEmbeddingDimensions = 768
	DefaultLogLevel            = "info"
	DefaultLogFormat           = "json"
	DefaultAuthEnabled         = false
	DefaultAuthIssuer          = "conexus"
	DefaultAuthAudience        = "conexus-api"
	DefaultAuthTokenExpiry     = 60 // 1 hour in minutes
	DefaultSecurityCSPEnabled  = true
	DefaultSecurityHSTSEnabled = true
	DefaultSecurityHSTSMaxAge  = 31536000 // 1 year
	DefaultCORSEnabled         = false
	DefaultCORSMaxAge          = 86400 // 24 hours
	DefaultTLSEnabled          = false
	DefaultTLSCertFile         = ""
	DefaultTLSKeyFile          = ""
	DefaultTLSAutoCert         = false
	DefaultTLSAutoCertEmail    = ""
	DefaultTLSAutoCertCacheDir = "./data/tls-cache"
	DefaultTLSMinVersion       = "1.2"
	DefaultTLSHTTPRedirectPort = 80
	DefaultMetricsEnabled      = false
	DefaultMetricsPort         = 9091
	DefaultMetricsPath         = "/metrics"
	DefaultTracingEnabled      = false
	DefaultTracingEndpoint     = "http://localhost:4318"
	DefaultSampleRate          = 0.1
	DefaultSentryEnabled       = false
	DefaultSentryDSN           = ""
	DefaultSentryEnv           = "development"
	DefaultSentrySampleRate    = 1.0
	DefaultSentryRelease       = "0.1.2-alpha"
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
		validatedPath, err := validation.ValidateConfigPath(configFile)
		if err != nil {
			return nil, fmt.Errorf("config file path validation failed: %w", err)
		}

		fileCfg, err := loadFile(validatedPath)
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
		Embedding: EmbeddingConfig{
			Provider:   DefaultEmbeddingProvider,
			Model:      DefaultEmbeddingModel,
			Dimensions: DefaultEmbeddingDimensions,
			Config:     make(map[string]interface{}),
		},
		Logging: LoggingConfig{
			Level:  DefaultLogLevel,
			Format: DefaultLogFormat,
		},
		Auth: AuthConfig{
			Enabled:     DefaultAuthEnabled,
			Issuer:      DefaultAuthIssuer,
			Audience:    DefaultAuthAudience,
			TokenExpiry: DefaultAuthTokenExpiry,
		},
		Security: SecurityConfig{
			CSP: CSPConfig{
				Enabled: DefaultSecurityCSPEnabled,
				Default: []string{"'none'"},
				Script:  []string{"'self'"},
				Style:   []string{"'self'"},
				Image:   []string{"'self'"},
				Font:    []string{"'self'"},
				Connect: []string{"'self'"},
				Media:   []string{"'none'"},
				Object:  []string{"'none'"},
				Frame:   []string{"'none'"},
			},
			HSTS: HSTSConfig{
				Enabled:           DefaultSecurityHSTSEnabled,
				MaxAge:            DefaultSecurityHSTSMaxAge,
				IncludeSubdomains: true,
				Preload:           false,
			},
			XFrameOptions:       "DENY",
			XContentTypeOptions: "nosniff",
			ReferrerPolicy:      "strict-origin-when-cross-origin",
			PermissionsPolicy:   "camera=(), microphone=(), geolocation=(), payment=()",
		},
		CORS: CORSConfig{
			Enabled:          DefaultCORSEnabled,
			AllowedOrigins:   []string{},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{},
			AllowCredentials: false,
			MaxAge:           DefaultCORSMaxAge,
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
	// Clean path to prevent basic traversal attacks
	safePath := filepath.Clean(path)

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

	// Embedding config
	if provider := os.Getenv("CONEXUS_EMBEDDING_PROVIDER"); provider != "" {
		cfg.Embedding.Provider = provider
	}
	if model := os.Getenv("CONEXUS_EMBEDDING_MODEL"); model != "" {
		cfg.Embedding.Model = model
	}
	if dimensions := os.Getenv("CONEXUS_EMBEDDING_DIMENSIONS"); dimensions != "" {
		if dim, err := strconv.Atoi(dimensions); err == nil {
			cfg.Embedding.Dimensions = dim
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

	// Auth config
	if authEnabled := os.Getenv("CONEXUS_AUTH_ENABLED"); authEnabled != "" {
		if enabled, err := strconv.ParseBool(authEnabled); err == nil {
			cfg.Auth.Enabled = enabled
		}
	}
	if authIssuer := os.Getenv("CONEXUS_AUTH_ISSUER"); authIssuer != "" {
		cfg.Auth.Issuer = authIssuer
	}
	if authAudience := os.Getenv("CONEXUS_AUTH_AUDIENCE"); authAudience != "" {
		cfg.Auth.Audience = authAudience
	}
	if authPublicKey := os.Getenv("CONEXUS_AUTH_PUBLIC_KEY"); authPublicKey != "" {
		cfg.Auth.PublicKey = authPublicKey
	}
	if authPrivateKey := os.Getenv("CONEXUS_AUTH_PRIVATE_KEY"); authPrivateKey != "" {
		cfg.Auth.PrivateKey = authPrivateKey
	}
	if authTokenExpiry := os.Getenv("CONEXUS_AUTH_TOKEN_EXPIRY"); authTokenExpiry != "" {
		if expiry, err := strconv.Atoi(authTokenExpiry); err == nil {
			cfg.Auth.TokenExpiry = expiry
		}
	}

	// Security config
	if securityCSPEnabled := os.Getenv("CONEXUS_SECURITY_CSP_ENABLED"); securityCSPEnabled != "" {
		if enabled, err := strconv.ParseBool(securityCSPEnabled); err == nil {
			cfg.Security.CSP.Enabled = enabled
		}
	}
	if securityHSTSEnabled := os.Getenv("CONEXUS_SECURITY_HSTS_ENABLED"); securityHSTSEnabled != "" {
		if enabled, err := strconv.ParseBool(securityHSTSEnabled); err == nil {
			cfg.Security.HSTS.Enabled = enabled
		}
	}
	if securityHSTSMaxAge := os.Getenv("CONEXUS_SECURITY_HSTS_MAX_AGE"); securityHSTSMaxAge != "" {
		if maxAge, err := strconv.Atoi(securityHSTSMaxAge); err == nil {
			cfg.Security.HSTS.MaxAge = maxAge
		}
	}
	if securityHSTSIncludeSubdomains := os.Getenv("CONEXUS_SECURITY_HSTS_INCLUDE_SUBDOMAINS"); securityHSTSIncludeSubdomains != "" {
		if include, err := strconv.ParseBool(securityHSTSIncludeSubdomains); err == nil {
			cfg.Security.HSTS.IncludeSubdomains = include
		}
	}
	if securityHSTSPreload := os.Getenv("CONEXUS_SECURITY_HSTS_PRELOAD"); securityHSTSPreload != "" {
		if preload, err := strconv.ParseBool(securityHSTSPreload); err == nil {
			cfg.Security.HSTS.Preload = preload
		}
	}
	if securityXFrameOptions := os.Getenv("CONEXUS_SECURITY_X_FRAME_OPTIONS"); securityXFrameOptions != "" {
		cfg.Security.XFrameOptions = securityXFrameOptions
	}
	if securityXContentTypeOptions := os.Getenv("CONEXUS_SECURITY_X_CONTENT_TYPE_OPTIONS"); securityXContentTypeOptions != "" {
		cfg.Security.XContentTypeOptions = securityXContentTypeOptions
	}
	if securityReferrerPolicy := os.Getenv("CONEXUS_SECURITY_REFERRER_POLICY"); securityReferrerPolicy != "" {
		cfg.Security.ReferrerPolicy = securityReferrerPolicy
	}
	if securityPermissionsPolicy := os.Getenv("CONEXUS_SECURITY_PERMISSIONS_POLICY"); securityPermissionsPolicy != "" {
		cfg.Security.PermissionsPolicy = securityPermissionsPolicy
	}

	// CORS config
	if corsEnabled := os.Getenv("CONEXUS_CORS_ENABLED"); corsEnabled != "" {
		if enabled, err := strconv.ParseBool(corsEnabled); err == nil {
			cfg.CORS.Enabled = enabled
		}
	}
	if corsAllowedOrigins := os.Getenv("CONEXUS_CORS_ALLOWED_ORIGINS"); corsAllowedOrigins != "" {
		// Parse comma-separated list
		if origins := strings.Split(corsAllowedOrigins, ","); len(origins) > 0 {
			cfg.CORS.AllowedOrigins = make([]string, 0, len(origins))
			for _, origin := range origins {
				if trimmed := strings.TrimSpace(origin); trimmed != "" {
					cfg.CORS.AllowedOrigins = append(cfg.CORS.AllowedOrigins, trimmed)
				}
			}
		}
	}
	if corsAllowedMethods := os.Getenv("CONEXUS_CORS_ALLOWED_METHODS"); corsAllowedMethods != "" {
		// Parse comma-separated list
		if methods := strings.Split(corsAllowedMethods, ","); len(methods) > 0 {
			cfg.CORS.AllowedMethods = make([]string, 0, len(methods))
			for _, method := range methods {
				if trimmed := strings.TrimSpace(method); trimmed != "" {
					cfg.CORS.AllowedMethods = append(cfg.CORS.AllowedMethods, trimmed)
				}
			}
		}
	}
	if corsAllowedHeaders := os.Getenv("CONEXUS_CORS_ALLOWED_HEADERS"); corsAllowedHeaders != "" {
		// Parse comma-separated list
		if headers := strings.Split(corsAllowedHeaders, ","); len(headers) > 0 {
			cfg.CORS.AllowedHeaders = make([]string, 0, len(headers))
			for _, header := range headers {
				if trimmed := strings.TrimSpace(header); trimmed != "" {
					cfg.CORS.AllowedHeaders = append(cfg.CORS.AllowedHeaders, trimmed)
				}
			}
		}
	}
	if corsExposedHeaders := os.Getenv("CONEXUS_CORS_EXPOSED_HEADERS"); corsExposedHeaders != "" {
		// Parse comma-separated list
		if headers := strings.Split(corsExposedHeaders, ","); len(headers) > 0 {
			cfg.CORS.ExposedHeaders = make([]string, 0, len(headers))
			for _, header := range headers {
				if trimmed := strings.TrimSpace(header); trimmed != "" {
					cfg.CORS.ExposedHeaders = append(cfg.CORS.ExposedHeaders, trimmed)
				}
			}
		}
	}
	if corsAllowCredentials := os.Getenv("CONEXUS_CORS_ALLOW_CREDENTIALS"); corsAllowCredentials != "" {
		if allow, err := strconv.ParseBool(corsAllowCredentials); err == nil {
			cfg.CORS.AllowCredentials = allow
		}
	}
	if corsMaxAge := os.Getenv("CONEXUS_CORS_MAX_AGE"); corsMaxAge != "" {
		if maxAge, err := strconv.Atoi(corsMaxAge); err == nil {
			cfg.CORS.MaxAge = maxAge
		}
	}

	// TLS config
	if tlsEnabled := os.Getenv("CONEXUS_TLS_ENABLED"); tlsEnabled != "" {
		if enabled, err := strconv.ParseBool(tlsEnabled); err == nil {
			cfg.TLS.Enabled = enabled
		}
	}
	if tlsCertFile := os.Getenv("CONEXUS_TLS_CERT_FILE"); tlsCertFile != "" {
		cfg.TLS.CertFile = tlsCertFile
	}
	if tlsKeyFile := os.Getenv("CONEXUS_TLS_KEY_FILE"); tlsKeyFile != "" {
		cfg.TLS.KeyFile = tlsKeyFile
	}
	if tlsAutoCert := os.Getenv("CONEXUS_TLS_AUTO_CERT"); tlsAutoCert != "" {
		if auto, err := strconv.ParseBool(tlsAutoCert); err == nil {
			cfg.TLS.AutoCert = auto
		}
	}
	if tlsAutoCertDomains := os.Getenv("CONEXUS_TLS_AUTO_CERT_DOMAINS"); tlsAutoCertDomains != "" {
		// Parse comma-separated list
		if domains := strings.Split(tlsAutoCertDomains, ","); len(domains) > 0 {
			cfg.TLS.AutoCertDomains = make([]string, 0, len(domains))
			for _, domain := range domains {
				if trimmed := strings.TrimSpace(domain); trimmed != "" {
					cfg.TLS.AutoCertDomains = append(cfg.TLS.AutoCertDomains, trimmed)
				}
			}
		}
	}
	if tlsAutoCertEmail := os.Getenv("CONEXUS_TLS_AUTO_CERT_EMAIL"); tlsAutoCertEmail != "" {
		cfg.TLS.AutoCertEmail = tlsAutoCertEmail
	}
	if tlsAutoCertCacheDir := os.Getenv("CONEXUS_TLS_AUTO_CERT_CACHE_DIR"); tlsAutoCertCacheDir != "" {
		cfg.TLS.AutoCertCacheDir = tlsAutoCertCacheDir
	}
	if tlsMinVersion := os.Getenv("CONEXUS_TLS_MIN_VERSION"); tlsMinVersion != "" {
		cfg.TLS.MinVersion = tlsMinVersion
	}
	if tlsCipherSuites := os.Getenv("CONEXUS_TLS_CIPHER_SUITES"); tlsCipherSuites != "" {
		// Parse comma-separated list
		if suites := strings.Split(tlsCipherSuites, ","); len(suites) > 0 {
			cfg.TLS.CipherSuites = make([]string, 0, len(suites))
			for _, suite := range suites {
				if trimmed := strings.TrimSpace(suite); trimmed != "" {
					cfg.TLS.CipherSuites = append(cfg.TLS.CipherSuites, trimmed)
				}
			}
		}
	}
	if tlsCurvePreferences := os.Getenv("CONEXUS_TLS_CURVE_PREFERENCES"); tlsCurvePreferences != "" {
		// Parse comma-separated list
		if curves := strings.Split(tlsCurvePreferences, ","); len(curves) > 0 {
			cfg.TLS.CurvePreferences = make([]string, 0, len(curves))
			for _, curve := range curves {
				if trimmed := strings.TrimSpace(curve); trimmed != "" {
					cfg.TLS.CurvePreferences = append(cfg.TLS.CurvePreferences, trimmed)
				}
			}
		}
	}
	if tlsHTTPRedirectPort := os.Getenv("CONEXUS_TLS_HTTP_REDIRECT_PORT"); tlsHTTPRedirectPort != "" {
		if port, err := strconv.Atoi(tlsHTTPRedirectPort); err == nil {
			cfg.TLS.HTTPRedirectPort = port
		}
	}

	// RateLimit config
	if rateLimitEnabled := os.Getenv("CONEXUS_RATE_LIMIT_ENABLED"); rateLimitEnabled != "" {
		if enabled, err := strconv.ParseBool(rateLimitEnabled); err == nil {
			cfg.RateLimit.Enabled = enabled
		}
	}
	if rateLimitAlgorithm := os.Getenv("CONEXUS_RATE_LIMIT_ALGORITHM"); rateLimitAlgorithm != "" {
		cfg.RateLimit.Algorithm = rateLimitAlgorithm
	}
	if rateLimitRedisEnabled := os.Getenv("CONEXUS_RATE_LIMIT_REDIS_ENABLED"); rateLimitRedisEnabled != "" {
		if enabled, err := strconv.ParseBool(rateLimitRedisEnabled); err == nil {
			cfg.RateLimit.Redis.Enabled = enabled
		}
	}
	if rateLimitRedisAddr := os.Getenv("CONEXUS_RATE_LIMIT_REDIS_ADDR"); rateLimitRedisAddr != "" {
		cfg.RateLimit.Redis.Addr = rateLimitRedisAddr
	}
	if rateLimitRedisPassword := os.Getenv("CONEXUS_RATE_LIMIT_REDIS_PASSWORD"); rateLimitRedisPassword != "" {
		cfg.RateLimit.Redis.Password = rateLimitRedisPassword
	}
	if rateLimitRedisDB := os.Getenv("CONEXUS_RATE_LIMIT_REDIS_DB"); rateLimitRedisDB != "" {
		if db, err := strconv.Atoi(rateLimitRedisDB); err == nil {
			cfg.RateLimit.Redis.DB = db
		}
	}
	if rateLimitRedisKeyPrefix := os.Getenv("CONEXUS_RATE_LIMIT_REDIS_KEY_PREFIX"); rateLimitRedisKeyPrefix != "" {
		cfg.RateLimit.Redis.KeyPrefix = rateLimitRedisKeyPrefix
	}
	if rateLimitDefaultRequests := os.Getenv("CONEXUS_RATE_LIMIT_DEFAULT_REQUESTS"); rateLimitDefaultRequests != "" {
		if requests, err := strconv.Atoi(rateLimitDefaultRequests); err == nil {
			cfg.RateLimit.Default.Requests = requests
		}
	}
	if rateLimitDefaultWindow := os.Getenv("CONEXUS_RATE_LIMIT_DEFAULT_WINDOW"); rateLimitDefaultWindow != "" {
		if window, err := time.ParseDuration(rateLimitDefaultWindow); err == nil {
			cfg.RateLimit.Default.Window = window
		}
	}
	if rateLimitHealthRequests := os.Getenv("CONEXUS_RATE_LIMIT_HEALTH_REQUESTS"); rateLimitHealthRequests != "" {
		if requests, err := strconv.Atoi(rateLimitHealthRequests); err == nil {
			cfg.RateLimit.Health.Requests = requests
		}
	}
	if rateLimitHealthWindow := os.Getenv("CONEXUS_RATE_LIMIT_HEALTH_WINDOW"); rateLimitHealthWindow != "" {
		if window, err := time.ParseDuration(rateLimitHealthWindow); err == nil {
			cfg.RateLimit.Health.Window = window
		}
	}
	if rateLimitWebhookRequests := os.Getenv("CONEXUS_RATE_LIMIT_WEBHOOK_REQUESTS"); rateLimitWebhookRequests != "" {
		if requests, err := strconv.Atoi(rateLimitWebhookRequests); err == nil {
			cfg.RateLimit.Webhook.Requests = requests
		}
	}
	if rateLimitWebhookWindow := os.Getenv("CONEXUS_RATE_LIMIT_WEBHOOK_WINDOW"); rateLimitWebhookWindow != "" {
		if window, err := time.ParseDuration(rateLimitWebhookWindow); err == nil {
			cfg.RateLimit.Webhook.Window = window
		}
	}
	if rateLimitAuthRequests := os.Getenv("CONEXUS_RATE_LIMIT_AUTH_REQUESTS"); rateLimitAuthRequests != "" {
		if requests, err := strconv.Atoi(rateLimitAuthRequests); err == nil {
			cfg.RateLimit.Auth.Requests = requests
		}
	}
	if rateLimitAuthWindow := os.Getenv("CONEXUS_RATE_LIMIT_AUTH_WINDOW"); rateLimitAuthWindow != "" {
		if window, err := time.ParseDuration(rateLimitAuthWindow); err == nil {
			cfg.RateLimit.Auth.Window = window
		}
	}
	if rateLimitBurstMultiplier := os.Getenv("CONEXUS_RATE_LIMIT_BURST_MULTIPLIER"); rateLimitBurstMultiplier != "" {
		if multiplier, err := strconv.ParseFloat(rateLimitBurstMultiplier, 64); err == nil {
			cfg.RateLimit.BurstMultiplier = multiplier
		}
	}
	if rateLimitCleanupInterval := os.Getenv("CONEXUS_RATE_LIMIT_CLEANUP_INTERVAL"); rateLimitCleanupInterval != "" {
		if interval, err := time.ParseDuration(rateLimitCleanupInterval); err == nil {
			cfg.RateLimit.CleanupInterval = interval
		}
	}
	if rateLimitSkipPaths := os.Getenv("CONEXUS_RATE_LIMIT_SKIP_PATHS"); rateLimitSkipPaths != "" {
		// Parse comma-separated list
		if paths := strings.Split(rateLimitSkipPaths, ","); len(paths) > 0 {
			cfg.RateLimit.SkipPaths = make([]string, 0, len(paths))
			for _, path := range paths {
				if trimmed := strings.TrimSpace(path); trimmed != "" {
					cfg.RateLimit.SkipPaths = append(cfg.RateLimit.SkipPaths, trimmed)
				}
			}
		}
	}
	if rateLimitSkipIPs := os.Getenv("CONEXUS_RATE_LIMIT_SKIP_IPS"); rateLimitSkipIPs != "" {
		// Parse comma-separated list
		if ips := strings.Split(rateLimitSkipIPs, ","); len(ips) > 0 {
			cfg.RateLimit.SkipIPs = make([]string, 0, len(ips))
			for _, ip := range ips {
				if trimmed := strings.TrimSpace(ip); trimmed != "" {
					cfg.RateLimit.SkipIPs = append(cfg.RateLimit.SkipIPs, trimmed)
				}
			}
		}
	}
	if rateLimitTrustedProxies := os.Getenv("CONEXUS_RATE_LIMIT_TRUSTED_PROXIES"); rateLimitTrustedProxies != "" {
		// Parse comma-separated list
		if proxies := strings.Split(rateLimitTrustedProxies, ","); len(proxies) > 0 {
			cfg.RateLimit.TrustedProxies = make([]string, 0, len(proxies))
			for _, proxy := range proxies {
				if trimmed := strings.TrimSpace(proxy); trimmed != "" {
					cfg.RateLimit.TrustedProxies = append(cfg.RateLimit.TrustedProxies, trimmed)
				}
			}
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
	if override.Embedding.Config != nil {
		result.Embedding.Config = override.Embedding.Config
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

	// Auth
	if override.Auth.Enabled != DefaultAuthEnabled {
		result.Auth.Enabled = override.Auth.Enabled
	}
	if override.Auth.Issuer != "" {
		result.Auth.Issuer = override.Auth.Issuer
	}
	if override.Auth.Audience != "" {
		result.Auth.Audience = override.Auth.Audience
	}
	if override.Auth.PublicKey != "" {
		result.Auth.PublicKey = override.Auth.PublicKey
	}
	if override.Auth.PrivateKey != "" {
		result.Auth.PrivateKey = override.Auth.PrivateKey
	}
	if override.Auth.TokenExpiry != 0 {
		result.Auth.TokenExpiry = override.Auth.TokenExpiry
	}

	// Security
	if override.Security.CSP.Enabled != DefaultSecurityCSPEnabled {
		result.Security.CSP.Enabled = override.Security.CSP.Enabled
	}
	if len(override.Security.CSP.Default) > 0 {
		result.Security.CSP.Default = override.Security.CSP.Default
	}
	if len(override.Security.CSP.Script) > 0 {
		result.Security.CSP.Script = override.Security.CSP.Script
	}
	if len(override.Security.CSP.Style) > 0 {
		result.Security.CSP.Style = override.Security.CSP.Style
	}
	if len(override.Security.CSP.Image) > 0 {
		result.Security.CSP.Image = override.Security.CSP.Image
	}
	if len(override.Security.CSP.Font) > 0 {
		result.Security.CSP.Font = override.Security.CSP.Font
	}
	if len(override.Security.CSP.Connect) > 0 {
		result.Security.CSP.Connect = override.Security.CSP.Connect
	}
	if len(override.Security.CSP.Media) > 0 {
		result.Security.CSP.Media = override.Security.CSP.Media
	}
	if len(override.Security.CSP.Object) > 0 {
		result.Security.CSP.Object = override.Security.CSP.Object
	}
	if len(override.Security.CSP.Frame) > 0 {
		result.Security.CSP.Frame = override.Security.CSP.Frame
	}
	if override.Security.CSP.Report != "" {
		result.Security.CSP.Report = override.Security.CSP.Report
	}

	if override.Security.HSTS.Enabled != DefaultSecurityHSTSEnabled {
		result.Security.HSTS.Enabled = override.Security.HSTS.Enabled
	}
	if override.Security.HSTS.MaxAge != 0 {
		result.Security.HSTS.MaxAge = override.Security.HSTS.MaxAge
	}
	if override.Security.HSTS.IncludeSubdomains {
		result.Security.HSTS.IncludeSubdomains = override.Security.HSTS.IncludeSubdomains
	}
	if override.Security.HSTS.Preload {
		result.Security.HSTS.Preload = override.Security.HSTS.Preload
	}

	if override.Security.XFrameOptions != "" {
		result.Security.XFrameOptions = override.Security.XFrameOptions
	}
	if override.Security.XContentTypeOptions != "" {
		result.Security.XContentTypeOptions = override.Security.XContentTypeOptions
	}
	if override.Security.ReferrerPolicy != "" {
		result.Security.ReferrerPolicy = override.Security.ReferrerPolicy
	}
	if override.Security.PermissionsPolicy != "" {
		result.Security.PermissionsPolicy = override.Security.PermissionsPolicy
	}

	// CORS
	if override.CORS.Enabled != DefaultCORSEnabled {
		result.CORS.Enabled = override.CORS.Enabled
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
	if override.CORS.AllowCredentials {
		result.CORS.AllowCredentials = override.CORS.AllowCredentials
	}
	if override.CORS.MaxAge != 0 {
		result.CORS.MaxAge = override.CORS.MaxAge
	}

	// TLS
	if override.TLS.Enabled != DefaultTLSEnabled {
		result.TLS.Enabled = override.TLS.Enabled
	}
	if override.TLS.CertFile != "" {
		result.TLS.CertFile = override.TLS.CertFile
	}
	if override.TLS.KeyFile != "" {
		result.TLS.KeyFile = override.TLS.KeyFile
	}
	if override.TLS.AutoCert != DefaultTLSAutoCert {
		result.TLS.AutoCert = override.TLS.AutoCert
	}
	if len(override.TLS.AutoCertDomains) > 0 {
		result.TLS.AutoCertDomains = override.TLS.AutoCertDomains
	}
	if override.TLS.AutoCertEmail != "" {
		result.TLS.AutoCertEmail = override.TLS.AutoCertEmail
	}
	if override.TLS.AutoCertCacheDir != "" {
		result.TLS.AutoCertCacheDir = override.TLS.AutoCertCacheDir
	}
	if override.TLS.MinVersion != "" {
		result.TLS.MinVersion = override.TLS.MinVersion
	}
	if len(override.TLS.CipherSuites) > 0 {
		result.TLS.CipherSuites = override.TLS.CipherSuites
	}
	if len(override.TLS.CurvePreferences) > 0 {
		result.TLS.CurvePreferences = override.TLS.CurvePreferences
	}
	if override.TLS.HTTPRedirectPort != 0 {
		result.TLS.HTTPRedirectPort = override.TLS.HTTPRedirectPort
	}

	// RateLimit
	if override.RateLimit.Enabled {
		result.RateLimit.Enabled = override.RateLimit.Enabled
	}
	if override.RateLimit.Algorithm != "" {
		result.RateLimit.Algorithm = override.RateLimit.Algorithm
	}
	if override.RateLimit.Redis.Enabled {
		result.RateLimit.Redis.Enabled = override.RateLimit.Redis.Enabled
	}
	if override.RateLimit.Redis.Addr != "" {
		result.RateLimit.Redis.Addr = override.RateLimit.Redis.Addr
	}
	if override.RateLimit.Redis.Password != "" {
		result.RateLimit.Redis.Password = override.RateLimit.Redis.Password
	}
	if override.RateLimit.Redis.DB != 0 {
		result.RateLimit.Redis.DB = override.RateLimit.Redis.DB
	}
	if override.RateLimit.Redis.KeyPrefix != "" {
		result.RateLimit.Redis.KeyPrefix = override.RateLimit.Redis.KeyPrefix
	}
	if override.RateLimit.Default.Requests != 0 {
		result.RateLimit.Default.Requests = override.RateLimit.Default.Requests
	}
	if override.RateLimit.Default.Window != 0 {
		result.RateLimit.Default.Window = override.RateLimit.Default.Window
	}
	if override.RateLimit.Health.Requests != 0 {
		result.RateLimit.Health.Requests = override.RateLimit.Health.Requests
	}
	if override.RateLimit.Health.Window != 0 {
		result.RateLimit.Health.Window = override.RateLimit.Health.Window
	}
	if override.RateLimit.Webhook.Requests != 0 {
		result.RateLimit.Webhook.Requests = override.RateLimit.Webhook.Requests
	}
	if override.RateLimit.Webhook.Window != 0 {
		result.RateLimit.Webhook.Window = override.RateLimit.Webhook.Window
	}
	if override.RateLimit.Auth.Requests != 0 {
		result.RateLimit.Auth.Requests = override.RateLimit.Auth.Requests
	}
	if override.RateLimit.Auth.Window != 0 {
		result.RateLimit.Auth.Window = override.RateLimit.Auth.Window
	}
	if override.RateLimit.BurstMultiplier != 0 {
		result.RateLimit.BurstMultiplier = override.RateLimit.BurstMultiplier
	}
	if override.RateLimit.CleanupInterval != 0 {
		result.RateLimit.CleanupInterval = override.RateLimit.CleanupInterval
	}
	if len(override.RateLimit.SkipPaths) > 0 {
		result.RateLimit.SkipPaths = override.RateLimit.SkipPaths
	}
	if len(override.RateLimit.SkipIPs) > 0 {
		result.RateLimit.SkipIPs = override.RateLimit.SkipIPs
	}
	if len(override.RateLimit.TrustedProxies) > 0 {
		result.RateLimit.TrustedProxies = override.RateLimit.TrustedProxies
	}

	return &result
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	// Validate server config - port 0 is allowed for stdio mode
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 0-65535, 0 for stdio mode)", c.Server.Port)
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

	// Validate auth config
	if c.Auth.Enabled {
		if c.Auth.Issuer == "" {
			return fmt.Errorf("auth issuer cannot be empty when auth enabled")
		}
		if c.Auth.Audience == "" {
			return fmt.Errorf("auth audience cannot be empty when auth enabled")
		}
		if c.Auth.PublicKey == "" {
			return fmt.Errorf("auth public key cannot be empty when auth enabled")
		}
		if c.Auth.PrivateKey == "" {
			return fmt.Errorf("auth private key cannot be empty when auth enabled")
		}
		if c.Auth.TokenExpiry <= 0 {
			return fmt.Errorf("auth token expiry must be positive: %d", c.Auth.TokenExpiry)
		}
	}

	// Validate TLS config
	if c.TLS.Enabled {
		// If not using auto-cert, cert and key files are required
		if !c.TLS.AutoCert {
			if c.TLS.CertFile == "" {
				return fmt.Errorf("TLS cert file cannot be empty when TLS enabled and auto-cert disabled")
			}
			if c.TLS.KeyFile == "" {
				return fmt.Errorf("TLS key file cannot be empty when TLS enabled and auto-cert disabled")
			}
		} else {
			// Auto-cert requires domains and email
			if len(c.TLS.AutoCertDomains) == 0 {
				return fmt.Errorf("auto-cert domains cannot be empty when auto-cert enabled")
			}
			if c.TLS.AutoCertEmail == "" {
				return fmt.Errorf("auto-cert email cannot be empty when auto-cert enabled")
			}
		}

		// Validate TLS version
		validTLSVersions := []string{"1.0", "1.1", "1.2", "1.3"}
		if c.TLS.MinVersion != "" && !contains(validTLSVersions, c.TLS.MinVersion) {
			return fmt.Errorf("invalid TLS min version: %s (valid: %v)", c.TLS.MinVersion, validTLSVersions)
		}

		// Validate HTTP redirect port (only if TLS is enabled)
		if c.TLS.HTTPRedirectPort != 0 && (c.TLS.HTTPRedirectPort < 1 || c.TLS.HTTPRedirectPort > 65535) {
			return fmt.Errorf("invalid HTTP redirect port: %d (must be 1-65535 or 0 to disable)", c.TLS.HTTPRedirectPort)
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
			Config:     make(map[string]interface{}),
		},
		Logging: LoggingConfig{
			Level:  DefaultLogLevel,
			Format: DefaultLogFormat,
		},
		Auth: AuthConfig{
			Enabled:     DefaultAuthEnabled,
			Issuer:      DefaultAuthIssuer,
			Audience:    DefaultAuthAudience,
			TokenExpiry: DefaultAuthTokenExpiry,
		},
		Security: SecurityConfig{
			CSP: CSPConfig{
				Enabled: DefaultSecurityCSPEnabled,
				Default: []string{"'none'"},
				Script:  []string{"'self'"},
				Style:   []string{"'self'"},
				Image:   []string{"'self'"},
				Font:    []string{"'self'"},
				Connect: []string{"'self'"},
				Media:   []string{"'none'"},
				Object:  []string{"'none'"},
				Frame:   []string{"'none'"},
			},
			HSTS: HSTSConfig{
				Enabled:           DefaultSecurityHSTSEnabled,
				MaxAge:            DefaultSecurityHSTSMaxAge,
				IncludeSubdomains: true,
				Preload:           false,
			},
			XFrameOptions:       "DENY",
			XContentTypeOptions: "nosniff",
			ReferrerPolicy:      "strict-origin-when-cross-origin",
			PermissionsPolicy:   "camera=(), microphone=(), geolocation=(), payment=()",
		},
		CORS: CORSConfig{
			Enabled:          DefaultCORSEnabled,
			AllowedOrigins:   []string{},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{},
			AllowCredentials: false,
			MaxAge:           DefaultCORSMaxAge,
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
