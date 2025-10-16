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
				Logging: LoggingConfig{
					Level:  "debug",
					Format: "text",
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
				Logging: LoggingConfig{
					Level:  "warn",
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
			cfg: &Config{
				Server: ServerConfig{Port: 0},
			},
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
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}
