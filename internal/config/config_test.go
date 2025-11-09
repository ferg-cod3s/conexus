package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	assert.Equal(t, DefaultHost, cfg.Server.Host)
	assert.Equal(t, DefaultPort, cfg.Server.Port)
	assert.Equal(t, DefaultDBPath, cfg.Database.Path)
	assert.Equal(t, DefaultRootPath, cfg.Indexer.RootPath)
	assert.Equal(t, DefaultChunkSize, cfg.Indexer.ChunkSize)
	assert.Equal(t, DefaultChunkOverlap, cfg.Indexer.ChunkOverlap)
	assert.Equal(t, DefaultEmbeddingProvider, cfg.Embedding.Provider)
	assert.Equal(t, DefaultEmbeddingModel, cfg.Embedding.Model)
	assert.Equal(t, DefaultEmbeddingDimensions, cfg.Embedding.Dimensions)
	assert.Equal(t, DefaultLogLevel, cfg.Logging.Level)
	assert.Equal(t, DefaultLogFormat, cfg.Logging.Format)
}

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name: "all env vars",
			envVars: map[string]string{
				"CONEXUS_HOST":          "127.0.0.1",
				"CONEXUS_PORT":          "9090",
				"CONEXUS_DB_PATH":       "/custom/db.sqlite",
				"CONEXUS_ROOT_PATH":     "/custom/root",
				"CONEXUS_CHUNK_SIZE":    "1024",
				"CONEXUS_CHUNK_OVERLAP": "100",
				"CONEXUS_LOG_LEVEL":     "debug",
				"CONEXUS_LOG_FORMAT":    "text",
			},
			expected: &Config{
				Server: ServerConfig{
					Host: "127.0.0.1",
					Port: 9090,
				},
				Database: DatabaseConfig{
					Path: "/custom/db.sqlite",
				},
				Indexer: IndexerConfig{
					RootPath:     "/custom/root",
					ChunkSize:    1024,
					ChunkOverlap: 100,
				},
				Embedding: EmbeddingConfig{
					Provider:   DefaultEmbeddingProvider,
					Model:      DefaultEmbeddingModel,
					Dimensions: DefaultEmbeddingDimensions,
					Config:     nil,
				},
				Logging: LoggingConfig{
					Level:  "debug",
					Format: "text",
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
					Enabled: DefaultSecurityEnabled,
					CSP: CSPConfig{
						Enabled:    false,
						DefaultSrc: nil,
						ScriptSrc:  nil,
						StyleSrc:   nil,
						ImgSrc:     nil,
						FontSrc:    nil,
						ConnectSrc: nil,
						MediaSrc:   nil,
						ObjectSrc:  nil,
						FrameSrc:   nil,
					},
					HSTS: HSTSConfig{
						Enabled:           false,
						MaxAge:            DefaultHSTSMaxAge,
						IncludeSubdomains: false,
						Preload:           false,
					},
					FrameOptions:    DefaultFrameOptions,
					ContentType:     DefaultContentType,
					BrowserXSSBlock: DefaultBrowserXSSBlock,
					ReferrerPolicy:  DefaultReferrerPolicy,
				},
				CORS: CORSConfig{
					Enabled:          DefaultCORSEnabled,
					AllowedOrigins:   nil,
					AllowedMethods:   nil,
					AllowedHeaders:   nil,
					ExposedHeaders:   nil,
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
			},
		},
		{
			name: "partial env vars",
			envVars: map[string]string{
				"CONEXUS_PORT":      "3000",
				"CONEXUS_LOG_LEVEL": "warn",
			},
			expected: &Config{
				Server: ServerConfig{
					Host: DefaultHost,
					Port: 3000,
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
					Level:  "warn",
					Format: DefaultLogFormat,
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
			},
		},
		{
			name:    "no env vars (defaults)",
			envVars: map[string]string{},
			expected: &Config{
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
					Config:     nil,
				},
				Logging: LoggingConfig{
					Level:  DefaultLogLevel,
					Format: DefaultLogFormat,
				},
				TLS: TLSConfig{
					Enabled:          DefaultTLSEnabled,
					MinVersion:       DefaultTLSMinVersion,
					HTTPRedirectPort: DefaultTLSHTTPRedirectPort,
					AutoCertCacheDir: DefaultAutoCertCacheDir,
				},
				Auth: AuthConfig{
					Enabled:     DefaultAuthEnabled,
					Issuer:      "",
					Audience:    "",
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
					Enabled: DefaultSecurityEnabled,
					CSP: CSPConfig{
						Enabled:    false,
						DefaultSrc: nil,
						ScriptSrc:  nil,
						StyleSrc:   nil,
						ImgSrc:     nil,
						FontSrc:    nil,
						ConnectSrc: nil,
						MediaSrc:   nil,
						ObjectSrc:  nil,
						FrameSrc:   nil,
					},
					HSTS: HSTSConfig{
						Enabled:           false,
						MaxAge:            DefaultHSTSMaxAge,
						IncludeSubdomains: false,
						Preload:           false,
					},
					FrameOptions:    DefaultFrameOptions,
					ContentType:     DefaultContentType,
					BrowserXSSBlock: DefaultBrowserXSSBlock,
					ReferrerPolicy:  DefaultReferrerPolicy,
				},
				CORS: CORSConfig{
					Enabled:          DefaultCORSEnabled,
					AllowedOrigins:   nil,
					AllowedMethods:   nil,
					AllowedHeaders:   nil,
					ExposedHeaders:   nil,
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
			},
		},
		{
			name: "invalid int values ignored",
			envVars: map[string]string{
				"CONEXUS_PORT":          "invalid",
				"CONEXUS_CHUNK_SIZE":    "not-a-number",
				"CONEXUS_CHUNK_OVERLAP": "also-invalid",
			},
			expected: &Config{
				Server: ServerConfig{
					Host: DefaultHost,
					Port: DefaultPort, // unchanged
				},
				Database: DatabaseConfig{
					Path: DefaultDBPath,
				},
				Indexer: IndexerConfig{
					RootPath:     DefaultRootPath,
					ChunkSize:    DefaultChunkSize,    // unchanged
					ChunkOverlap: DefaultChunkOverlap, // unchanged
				},
				Embedding: EmbeddingConfig{
					Provider:   DefaultEmbeddingProvider,
					Model:      DefaultEmbeddingModel,
					Dimensions: DefaultEmbeddingDimensions,
					Config:     nil,
				},
				Logging: LoggingConfig{
					Level:  DefaultLogLevel,
					Format: DefaultLogFormat,
				},
				TLS: TLSConfig{
					Enabled:          DefaultTLSEnabled,
					MinVersion:       DefaultTLSMinVersion,
					HTTPRedirectPort: DefaultTLSHTTPRedirectPort,
					AutoCertCacheDir: DefaultAutoCertCacheDir,
				},
				Auth: AuthConfig{
					Enabled:     DefaultAuthEnabled,
					Issuer:      "",
					Audience:    "",
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
					Enabled: DefaultSecurityEnabled,
					CSP: CSPConfig{
						Enabled:    false,
						DefaultSrc: nil,
						ScriptSrc:  nil,
						StyleSrc:   nil,
						ImgSrc:     nil,
						FontSrc:    nil,
						ConnectSrc: nil,
						MediaSrc:   nil,
						ObjectSrc:  nil,
						FrameSrc:   nil,
					},
					HSTS: HSTSConfig{
						Enabled:           false,
						MaxAge:            DefaultHSTSMaxAge,
						IncludeSubdomains: false,
						Preload:           false,
					},
					FrameOptions:    DefaultFrameOptions,
					ContentType:     DefaultContentType,
					BrowserXSSBlock: DefaultBrowserXSSBlock,
					ReferrerPolicy:  DefaultReferrerPolicy,
				},
				CORS: CORSConfig{
					Enabled:          DefaultCORSEnabled,
					AllowedOrigins:   nil,
					AllowedMethods:   nil,
					AllowedHeaders:   nil,
					ExposedHeaders:   nil,
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv(t)

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			t.Cleanup(func() { clearEnv(t) })

			cfg := defaults()
			result := loadEnv(cfg)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		ext         string
		expected    *Config
		expectError bool
	}{
		{
			name: "valid yaml",
			content: `
server:
  host: "127.0.0.1"
  port: 9090
database:
  path: "/custom/db.sqlite"
indexer:
  root_path: "/custom/root"
  chunk_size: 1024
  chunk_overlap: 100
logging:
  level: "debug"
  format: "text"
`,
			ext: ".yaml",
			expected: &Config{
				Server: ServerConfig{
					Host: "127.0.0.1",
					Port: 9090,
				},
				Database: DatabaseConfig{
					Path: "/custom/db.sqlite",
				},
				Indexer: IndexerConfig{
					RootPath:     "/custom/root",
					ChunkSize:    1024,
					ChunkOverlap: 100,
				},
				Embedding: EmbeddingConfig{},
				Logging: LoggingConfig{
					Level:  "debug",
					Format: "text",
				},
			},
		},
		{
			name: "valid json",
			content: `{
  "server": {
    "host": "127.0.0.1",
    "port": 9090
  },
  "database": {
    "path": "/custom/db.sqlite"
  },
  "indexer": {
    "root_path": "/custom/root",
    "chunk_size": 1024,
    "chunk_overlap": 100
  },
  "logging": {
    "level": "debug",
    "format": "text"
  }
}`,
			ext: ".json",
			expected: &Config{
				Server: ServerConfig{
					Host: "127.0.0.1",
					Port: 9090,
				},
				Database: DatabaseConfig{
					Path: "/custom/db.sqlite",
				},
				Indexer: IndexerConfig{
					RootPath:     "/custom/root",
					ChunkSize:    1024,
					ChunkOverlap: 100,
				},
				Embedding: EmbeddingConfig{},
				Logging: LoggingConfig{
					Level:  "debug",
					Format: "text",
				},
			},
		},
		{
			name: "partial yaml",
			content: `
server:
  port: 3000
logging:
  level: "warn"
`,
			ext: ".yaml",
			expected: &Config{
				Server: ServerConfig{
					Port: 3000,
				},
				Database:  DatabaseConfig{},
				Indexer:   IndexerConfig{},
				Embedding: EmbeddingConfig{},
				Logging: LoggingConfig{
					Level: "warn",
				},
			},
		},
		{
			name:        "invalid yaml",
			content:     "invalid: yaml: content: [",
			ext:         ".yaml",
			expectError: true,
		},
		{
			name:        "invalid json",
			content:     "{invalid json",
			ext:         ".json",
			expectError: true,
		},
		{
			name:        "unsupported extension",
			content:     "some content",
			ext:         ".txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config"+tt.ext)
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load file
			result, err := loadFile(tmpFile)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := loadFile("/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read file")
}

func TestMerge(t *testing.T) {
	base := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Path: "./data/db.sqlite",
		},
		Indexer: IndexerConfig{
			RootPath:     ".",
			ChunkSize:    512,
			ChunkOverlap: 50,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	override := &Config{
		Server: ServerConfig{
			Port: 9090, // override
		},
		Logging: LoggingConfig{
			Level: "debug", // override
		},
	}

	result := merge(base, override)

	// Overridden values
	assert.Equal(t, 9090, result.Server.Port)
	assert.Equal(t, "debug", result.Logging.Level)

	// Preserved values
	assert.Equal(t, "0.0.0.0", result.Server.Host)
	assert.Equal(t, "./data/db.sqlite", result.Database.Path)
	assert.Equal(t, ".", result.Indexer.RootPath)
	assert.Equal(t, 512, result.Indexer.ChunkSize)
	assert.Equal(t, 50, result.Indexer.ChunkOverlap)
	assert.Equal(t, "json", result.Logging.Format)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			cfg:         defaults(),
			expectError: false,
		},
		{
			name: "invalid port - too low",
			cfg: func() *Config {
				cfg := defaults()
				cfg.Server.Port = -1
				return cfg
			}(),
			expectError: true,
			errorMsg:    "invalid port",
		},
		{
			name: "invalid port - too high",
			cfg: &Config{
				Server: ServerConfig{Port: 99999},
			},
			expectError: true,
			errorMsg:    "invalid port",
		},
		{
			name: "empty database path",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: ""},
			},
			expectError: true,
			errorMsg:    "database path cannot be empty",
		},
		{
			name: "empty root path",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer:  IndexerConfig{RootPath: ""},
			},
			expectError: true,
			errorMsg:    "root path cannot be empty",
		},
		{
			name: "invalid chunk size",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:  ".",
					ChunkSize: 0,
				},
			},
			expectError: true,
			errorMsg:    "chunk size must be positive",
		},
		{
			name: "negative chunk overlap",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: -1,
				},
			},
			expectError: true,
			errorMsg:    "chunk overlap cannot be negative",
		},
		{
			name: "chunk overlap >= chunk size",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 512,
				},
			},
			expectError: true,
			errorMsg:    "chunk overlap",
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: "json",
				},
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name: "invalid log format",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "invalid log format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("defaults only", func(t *testing.T) {
		clearEnv(t)
		t.Cleanup(func() { clearEnv(t) })

		cfg, err := Load(context.Background())
		require.NoError(t, err)

		expected := defaults()
		assert.Equal(t, expected, cfg)
	})

	t.Run("with config file", func(t *testing.T) {
		clearEnv(t)
		t.Cleanup(func() { clearEnv(t) })

		// Create temp config file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		content := `
server:
  port: 9090
logging:
  level: "debug"
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		os.Setenv("CONEXUS_CONFIG_FILE", configFile)

		cfg, err := Load(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 9090, cfg.Server.Port)
		assert.Equal(t, "debug", cfg.Logging.Level)
		// Defaults should still be present
		assert.Equal(t, DefaultHost, cfg.Server.Host)
		assert.Equal(t, DefaultDBPath, cfg.Database.Path)
	})

	t.Run("env overrides file", func(t *testing.T) {
		clearEnv(t)
		t.Cleanup(func() { clearEnv(t) })

		// Create temp config file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		content := `
server:
  port: 9090
logging:
  level: "debug"
`
		err := os.WriteFile(configFile, []byte(content), 0644)
		require.NoError(t, err)

		os.Setenv("CONEXUS_CONFIG_FILE", configFile)
		os.Setenv("CONEXUS_PORT", "3000")          // override file
		os.Setenv("CONEXUS_LOG_LEVEL", "error")    // override file
		os.Setenv("CONEXUS_HOST", "192.168.1.100") // not in file

		cfg, err := Load(context.Background())
		require.NoError(t, err)

		// Env vars should win
		assert.Equal(t, 3000, cfg.Server.Port)
		assert.Equal(t, "error", cfg.Logging.Level)
		assert.Equal(t, "192.168.1.100", cfg.Server.Host)
	})

	t.Run("invalid config file", func(t *testing.T) {
		clearEnv(t)
		t.Cleanup(func() { clearEnv(t) })

		os.Setenv("CONEXUS_CONFIG_FILE", "/nonexistent/config.yaml")

		_, err := Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load config file")
	})

	t.Run("validation error", func(t *testing.T) {
		clearEnv(t)
		t.Cleanup(func() { clearEnv(t) })

		os.Setenv("CONEXUS_PORT", "99999") // invalid port

		_, err := Load(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validate config")
	})
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.True(t, contains(slice, "c"))
	assert.False(t, contains(slice, "d"))
	assert.False(t, contains(slice, ""))
	assert.False(t, contains([]string{}, "a"))
}

func TestDefault(t *testing.T) {
	cfg := Default()

	// Test that it returns the same as defaults()
	expectedDefaults := defaults()
	assert.Equal(t, expectedDefaults, cfg)

	// Test specific values
	assert.Equal(t, DefaultHost, cfg.Server.Host)
	assert.Equal(t, DefaultPort, cfg.Server.Port)
	assert.Equal(t, DefaultDBPath, cfg.Database.Path)
	assert.Equal(t, DefaultRootPath, cfg.Indexer.RootPath)
	assert.Equal(t, DefaultChunkSize, cfg.Indexer.ChunkSize)
	assert.Equal(t, DefaultChunkOverlap, cfg.Indexer.ChunkOverlap)
	assert.Equal(t, DefaultEmbeddingProvider, cfg.Embedding.Provider)
	assert.Equal(t, DefaultEmbeddingModel, cfg.Embedding.Model)
	assert.Equal(t, DefaultEmbeddingDimensions, cfg.Embedding.Dimensions)
	assert.Equal(t, DefaultLogLevel, cfg.Logging.Level)
	assert.Equal(t, DefaultLogFormat, cfg.Logging.Format)
}

func TestLoadEnv_Observability(t *testing.T) {
	// Test observability environment variables
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name: "metrics enabled",
			envVars: map[string]string{
				"CONEXUS_METRICS_ENABLED": "true",
				"CONEXUS_METRICS_PORT":    "9090",
				"CONEXUS_METRICS_PATH":    "/custom/metrics",
			},
			expected: &Config{
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{
						Enabled: true,
						Port:    9090,
						Path:    "/custom/metrics",
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
			},
		},
		{
			name: "tracing enabled",
			envVars: map[string]string{
				"CONEXUS_TRACING_ENABLED":     "true",
				"CONEXUS_TRACING_ENDPOINT":    "http://custom:4318",
				"CONEXUS_TRACING_SAMPLE_RATE": "0.5",
			},
			expected: &Config{
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{
						Enabled: DefaultMetricsEnabled,
						Port:    DefaultMetricsPort,
						Path:    DefaultMetricsPath,
					},
					Tracing: TracingConfig{
						Enabled:    true,
						Endpoint:   "http://custom:4318",
						SampleRate: 0.5,
					},
					Sentry: SentryConfig{
						Enabled:     DefaultSentryEnabled,
						DSN:         DefaultSentryDSN,
						Environment: DefaultSentryEnv,
						SampleRate:  DefaultSentrySampleRate,
						Release:     DefaultSentryRelease,
					},
				},
			},
		},
		{
			name: "sentry enabled",
			envVars: map[string]string{
				"CONEXUS_SENTRY_ENABLED":     "true",
				"CONEXUS_SENTRY_DSN":         "https://test@sentry.io/123",
				"CONEXUS_SENTRY_ENVIRONMENT": "production",
				"CONEXUS_SENTRY_SAMPLE_RATE": "0.8",
				"CONEXUS_SENTRY_RELEASE":     "v1.0.0",
			},
			expected: &Config{
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
						Enabled:     true,
						DSN:         "https://test@sentry.io/123",
						Environment: "production",
						SampleRate:  0.8,
						Release:     "v1.0.0",
					},
				},
			},
		},
		{
			name: "invalid boolean values ignored",
			envVars: map[string]string{
				"CONEXUS_METRICS_ENABLED": "invalid",
				"CONEXUS_TRACING_ENABLED": "not-a-bool",
				"CONEXUS_SENTRY_ENABLED":  "maybe",
			},
			expected: &Config{
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{
						Enabled: DefaultMetricsEnabled, // unchanged
						Port:    DefaultMetricsPort,
						Path:    DefaultMetricsPath,
					},
					Tracing: TracingConfig{
						Enabled:    DefaultTracingEnabled, // unchanged
						Endpoint:   DefaultTracingEndpoint,
						SampleRate: DefaultSampleRate,
					},
					Sentry: SentryConfig{
						Enabled:     DefaultSentryEnabled, // unchanged
						DSN:         DefaultSentryDSN,
						Environment: DefaultSentryEnv,
						SampleRate:  DefaultSentrySampleRate,
						Release:     DefaultSentryRelease,
					},
				},
			},
		},
		{
			name: "invalid float values ignored",
			envVars: map[string]string{
				"CONEXUS_TRACING_SAMPLE_RATE": "not-a-float",
				"CONEXUS_SENTRY_SAMPLE_RATE":  "invalid",
			},
			expected: &Config{
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{
						Enabled: DefaultMetricsEnabled,
						Port:    DefaultMetricsPort,
						Path:    DefaultMetricsPath,
					},
					Tracing: TracingConfig{
						Enabled:    DefaultTracingEnabled,
						Endpoint:   DefaultTracingEndpoint,
						SampleRate: DefaultSampleRate, // unchanged
					},
					Sentry: SentryConfig{
						Enabled:     DefaultSentryEnabled,
						DSN:         DefaultSentryDSN,
						Environment: DefaultSentryEnv,
						SampleRate:  DefaultSentrySampleRate, // unchanged
						Release:     DefaultSentryRelease,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv(t)

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			t.Cleanup(func() { clearEnv(t) })

			cfg := defaults()
			result := loadEnv(cfg)

			assert.Equal(t, tt.expected.Observability, result.Observability)
		})
	}
}

func TestMerge_Observability(t *testing.T) {
	base := &Config{
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled: false,
				Port:    9090,
				Path:    "/metrics",
			},
			Tracing: TracingConfig{
				Enabled:    false,
				Endpoint:   "http://localhost:4318",
				SampleRate: 0.1,
			},
			Sentry: SentryConfig{
				Enabled:     false,
				DSN:         "",
				Environment: "development",
				SampleRate:  1.0,
				Release:     "v0.1.0",
			},
		},
	}

	override := &Config{
		Observability: ObservabilityConfig{
			Metrics: MetricsConfig{
				Enabled: true,      // override
				Port:    8080,      // override
				Path:    "/custom", // override
			},
			Tracing: TracingConfig{
				Enabled:    true,                 // override
				Endpoint:   "http://custom:4318", // override
				SampleRate: 0.5,                  // override
			},
			Sentry: SentryConfig{
				Enabled:     true,                         // override
				DSN:         "https://test@sentry.io/123", // override
				Environment: "production",                 // override
				SampleRate:  0.8,                          // override
				Release:     "v1.0.0",                     // override
			},
		},
	}

	result := merge(base, override)

	// All observability values should be overridden
	assert.True(t, result.Observability.Metrics.Enabled)
	assert.Equal(t, 8080, result.Observability.Metrics.Port)
	assert.Equal(t, "/custom", result.Observability.Metrics.Path)

	assert.True(t, result.Observability.Tracing.Enabled)
	assert.Equal(t, "http://custom:4318", result.Observability.Tracing.Endpoint)
	assert.Equal(t, 0.5, result.Observability.Tracing.SampleRate)

	assert.True(t, result.Observability.Sentry.Enabled)
	assert.Equal(t, "https://test@sentry.io/123", result.Observability.Sentry.DSN)
	assert.Equal(t, "production", result.Observability.Sentry.Environment)
	assert.Equal(t, 0.8, result.Observability.Sentry.SampleRate)
	assert.Equal(t, "v1.0.0", result.Observability.Sentry.Release)
}

func TestValidate_Observability(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid observability disabled",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Embedding: EmbeddingConfig{
					Provider:   "mock",
					Model:      "mock-768",
					Dimensions: 768,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{Enabled: false},
					Tracing: TracingConfig{Enabled: false},
					Sentry:  SentryConfig{Enabled: false},
				},
			},
			expectError: false,
		},
		{
			name: "valid metrics enabled",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Embedding: EmbeddingConfig{
					Provider:   "mock",
					Model:      "mock-768",
					Dimensions: 768,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{
						Enabled: true,
						Port:    9090,
						Path:    "/metrics",
					},
					Tracing: TracingConfig{Enabled: false},
					Sentry:  SentryConfig{Enabled: false},
				},
			},
			expectError: false,
		},
		{
			name: "invalid metrics port",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{
						Enabled: true,
						Port:    0, // invalid
						Path:    "/metrics",
					},
					Tracing: TracingConfig{Enabled: false},
					Sentry:  SentryConfig{Enabled: false},
				},
			},
			expectError: true,
			errorMsg:    "invalid metrics port",
		},
		{
			name: "empty metrics path when enabled",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{
						Enabled: true,
						Port:    9090,
						Path:    "", // invalid
					},
					Tracing: TracingConfig{Enabled: false},
					Sentry:  SentryConfig{Enabled: false},
				},
			},
			expectError: true,
			errorMsg:    "metrics path cannot be empty",
		},
		{
			name: "valid tracing enabled",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Embedding: EmbeddingConfig{
					Provider:   "mock",
					Model:      "mock-768",
					Dimensions: 768,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{Enabled: false},
					Tracing: TracingConfig{
						Enabled:    true,
						Endpoint:   "http://localhost:4318",
						SampleRate: 0.1,
					},
					Sentry: SentryConfig{Enabled: false},
				},
			},
			expectError: false,
		},
		{
			name: "empty tracing endpoint when enabled",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{Enabled: false},
					Tracing: TracingConfig{
						Enabled:  true,
						Endpoint: "", // invalid
					},
					Sentry: SentryConfig{Enabled: false},
				},
			},
			expectError: true,
			errorMsg:    "tracing endpoint cannot be empty",
		},
		{
			name: "invalid tracing sample rate",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{Enabled: false},
					Tracing: TracingConfig{
						Enabled:    true,
						Endpoint:   "http://localhost:4318",
						SampleRate: 1.5, // invalid
					},
					Sentry: SentryConfig{Enabled: false},
				},
			},
			expectError: true,
			errorMsg:    "tracing sample rate must be between 0 and 1",
		},
		{
			name: "valid sentry enabled",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Embedding: EmbeddingConfig{
					Provider:   "mock",
					Model:      "mock-768",
					Dimensions: 768,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{Enabled: false},
					Tracing: TracingConfig{Enabled: false},
					Sentry: SentryConfig{
						Enabled:     true,
						DSN:         "https://test@sentry.io/123",
						Environment: "production",
						SampleRate:  0.8,
						Release:     "v1.0.0",
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty sentry DSN when enabled",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{Enabled: false},
					Tracing: TracingConfig{Enabled: false},
					Sentry: SentryConfig{
						Enabled: true,
						DSN:     "", // invalid
					},
				},
			},
			expectError: true,
			errorMsg:    "sentry DSN cannot be empty",
		},
		{
			name: "invalid sentry sample rate",
			cfg: &Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Path: "/db"},
				Indexer: IndexerConfig{
					RootPath:     ".",
					ChunkSize:    512,
					ChunkOverlap: 50,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Observability: ObservabilityConfig{
					Metrics: MetricsConfig{Enabled: false},
					Tracing: TracingConfig{Enabled: false},
					Sentry: SentryConfig{
						Enabled:    true,
						DSN:        "https://test@sentry.io/123",
						SampleRate: 1.5, // invalid
					},
				},
			},
			expectError: true,
			errorMsg:    "sentry sample rate must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper to clear all CONEXUS_* env vars
func clearEnv(t *testing.T) {
	vars := []string{
		"CONEXUS_HOST",
		"CONEXUS_PORT",
		"CONEXUS_DB_PATH",
		"CONEXUS_ROOT_PATH",
		"CONEXUS_CHUNK_SIZE",
		"CONEXUS_CHUNK_OVERLAP",
		"CONEXUS_LOG_LEVEL",
		"CONEXUS_LOG_FORMAT",
		"CONEXUS_CONFIG_FILE",
		"CONEXUS_METRICS_ENABLED",
		"CONEXUS_METRICS_PORT",
		"CONEXUS_METRICS_PATH",
		"CONEXUS_TRACING_ENABLED",
		"CONEXUS_TRACING_ENDPOINT",
		"CONEXUS_TRACING_SAMPLE_RATE",
		"CONEXUS_SENTRY_ENABLED",
		"CONEXUS_SENTRY_DSN",
		"CONEXUS_SENTRY_ENVIRONMENT",
		"CONEXUS_SENTRY_SAMPLE_RATE",
		"CONEXUS_SENTRY_RELEASE",
		"CONEXUS_SECURITY_CSP_ENABLED",
		"CONEXUS_SECURITY_HSTS_ENABLED",
		"CONEXUS_SECURITY_HSTS_MAX_AGE",
		"CONEXUS_SECURITY_HSTS_INCLUDE_SUBDOMAINS",
		"CONEXUS_SECURITY_HSTS_PRELOAD",
		"CONEXUS_SECURITY_X_FRAME_OPTIONS",
		"CONEXUS_SECURITY_X_CONTENT_TYPE_OPTIONS",
		"CONEXUS_SECURITY_REFERRER_POLICY",
		"CONEXUS_SECURITY_PERMISSIONS_POLICY",
		"CONEXUS_CORS_ENABLED",
		"CONEXUS_CORS_ALLOWED_ORIGINS",
		"CONEXUS_CORS_ALLOWED_METHODS",
		"CONEXUS_CORS_ALLOWED_HEADERS",
		"CONEXUS_CORS_EXPOSED_HEADERS",
		"CONEXUS_CORS_ALLOW_CREDENTIALS",
		"CONEXUS_CORS_MAX_AGE",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}
