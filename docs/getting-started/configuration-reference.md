# Configuration Reference

This appendix provides a complete reference for all Conexus configuration options, environment variables, and configuration file formats.

## Configuration File Format

Conexus supports YAML configuration files for complex setups. The configuration file is optional but recommended for production deployments.

### File Location

- **Default**: `config.yml` in the working directory
- **Custom**: Set via `CONEXUS_CONFIG` environment variable
- **Search order**:
  1. `CONEXUS_CONFIG` environment variable
  2. `config.yml` in current directory
  3. Built-in defaults

### Complete Configuration Schema

```yaml
# Conexus Configuration File
# Version: 0.1.0-alpha

# Project metadata
project:
  name: "my-project"                    # Project name
  description: "Project description"    # Project description
  version: "1.0.0"                      # Project version
  environment: "development"            # Environment (development/staging/production)

# Codebase settings
codebase:
  root: "."                             # Root directory (relative to config file)
  include_patterns:                     # File patterns to include
    - "**/*.go"
    - "**/*.js"
    - "**/*.py"
    - "**/*.rs"
    - "**/*.md"
  exclude_patterns:                     # File patterns to exclude
    - "**/node_modules/**"
    - "**/vendor/**"
    - "**/.git/**"
    - "**/dist/**"
    - "**/build/**"
    - "**/*.log"
  ignore_patterns:                      # Additional ignore patterns
    - "**/testdata/**"
    - "**/mocks/**"

# Indexing configuration
indexing:
  auto_reindex: true                    # Enable automatic reindexing
  reindex_interval: "30m"              # Reindex interval (duration)
  chunk_size: 500                      # Text chunk size
  workers: 2                           # Number of indexing workers
  memory_limit: "256MB"                # Memory limit for indexing
  max_file_size: "10MB"                # Maximum file size to index
  follow_symlinks: false               # Follow symbolic links
  encoding: "utf-8"                    # File encoding

# Search configuration
search:
  max_results: 50                      # Maximum search results
  similarity_threshold: 0.7            # Similarity threshold (0.0-1.0)
  cache_enabled: true                  # Enable search caching
  cache_ttl: "1h"                      # Cache time-to-live
  rerank_enabled: true                 # Enable result reranking
  rerank_model: "cross-encoder"        # Reranking model
  max_concurrent: 5                    # Maximum concurrent searches
  timeout: "30s"                       # Search timeout

# Embedding configuration
embedding:
  provider: "openai"                   # Provider: mock/openai/anthropic
  model: "text-embedding-3-small"      # Model name
  dimensions: 1536                     # Vector dimensions
  api_key: "${OPENAI_API_KEY}"         # API key (environment variable)
  api_base_url: "https://api.openai.com/v1"  # API base URL
  batch_size: 100                      # Batch size for processing
  rate_limit: 1000                     # Requests per minute
  timeout: "30s"                       # Request timeout
  retry_attempts: 3                    # Number of retry attempts
  retry_delay: "1s"                    # Delay between retries

# Vector store configuration
vectorstore:
  type: "sqlite"                       # Store type: sqlite/memory
  path: ".conexus/vectors.db"          # Database path (SQLite only)
  memory_limit: "512MB"                # Memory limit
  cache_size: 1000                     # Cache size
  optimize_on_startup: true            # Optimize on startup
  wal_mode: true                       # Write-ahead logging (SQLite)
  synchronous: "normal"                # Synchronization mode (SQLite)

# Security configuration
security:
  rate_limiting:
    enabled: true                      # Enable rate limiting
    algorithm: "sliding_window"        # Algorithm: sliding_window/token_bucket
    default_requests: 100              # Default requests per window
    default_window: "1m"               # Default time window
    redis:
      enabled: false                   # Enable Redis for distributed limiting
      addr: "localhost:6379"           # Redis address
      password: ""                     # Redis password
      db: 0                            # Redis database
      key_prefix: "conexus:ratelimit"  # Redis key prefix
    endpoints:                         # Endpoint-specific limits
      "/api/search": 1000
      "/api/index": 10

  tls:
    enabled: false                     # Enable TLS/HTTPS
    cert_file: ""                      # TLS certificate file
    key_file: ""                       # TLS private key file
    auto_cert: false                   # Enable automatic Let's Encrypt
    auto_cert_domains: []              # Domains for auto-cert
    auto_cert_email: ""                # Email for Let's Encrypt
    auto_cert_staging: true            # Use staging server

  cors:
    enabled: true                      # Enable CORS
    allowed_origins: ["*"]             # Allowed origins
    allowed_methods: ["GET", "POST"]   # Allowed methods
    allowed_headers: ["*"]             # Allowed headers
    allow_credentials: false           # Allow credentials
    max_age: "24h"                     # Preflight cache duration

  auth:
    enabled: false                     # Enable authentication
    provider: "jwt"                    # Auth provider: jwt/api_key
    jwt:
      secret: ""                       # JWT secret key
      algorithm: "HS256"               # JWT algorithm
      token_ttl: "24h"                 # Token time-to-live
      refresh_token_ttl: "168h"        # Refresh token TTL
    api_keys: []                       # API keys (for api_key provider)

# Observability configuration
observability:
  log_level: "info"                    # Log level: debug/info/warn/error
  log_format: "text"                   # Log format: text/json
  log_file: ""                         # Log file path (empty = stderr)
  log_max_size: "100MB"                # Max log file size
  log_max_age: "30d"                   # Max log file age
  log_max_backups: 5                   # Max log file backups

  metrics:
    enabled: false                     # Enable Prometheus metrics
    port: 9090                         # Metrics server port
    path: "/metrics"                   # Metrics endpoint path
    namespace: "conexus"               # Metrics namespace
    subsystem: "api"                   # Metrics subsystem

  tracing:
    enabled: false                     # Enable distributed tracing
    provider: "jaeger"                 # Tracing provider
    jaeger:
      endpoint: "http://localhost:14268/api/traces"  # Jaeger endpoint
      service_name: "conexus"          # Service name
      tags: {}                         # Additional tags
    otel:
      endpoint: ""                     # OTEL endpoint
      headers: {}                      # OTEL headers

  sentry:
    enabled: false                     # Enable Sentry error tracking
    dsn: ""                            # Sentry DSN
    environment: "development"         # Environment
    sample_rate: 1.0                   # Sample rate (0.0-1.0)
    traces_sample_rate: 1.0            # Traces sample rate
    release: ""                        # Release version
    debug: false                       # Debug mode

# MCP (Model Context Protocol) configuration
mcp:
  server_port: 9090                    # MCP server port
  allowed_origins: ["*"]               # Allowed CORS origins
  max_request_size: "10MB"             # Maximum request size
  request_timeout: "30s"               # Request timeout
  shutdown_timeout: "10s"              # Shutdown timeout
  health_check:
    enabled: true                      # Enable health checks
    path: "/health"                    # Health check path
    interval: "30s"                    # Health check interval

# Connectors configuration
connectors:
  - type: "github"                     # Connector type
    name: "Company GitHub"             # Connector name
    enabled: true                      # Enable connector
    config:                            # Connector-specific config
      token: "${GITHUB_TOKEN}"         # GitHub personal access token
      org: "mycompany"                 # Organization name
      repos: ["api", "web", "mobile"]  # Repository list
      include_prs: true                # Include pull requests
      include_issues: true             # Include issues
      include_discussions: false       # Include discussions
      since_days: 30                   # Look back days

  - type: "slack"                      # Slack connector
    name: "Engineering Slack"
    enabled: true
    config:
      token: "${SLACK_TOKEN}"          # Slack bot token
      channels: ["#dev", "#backend"]   # Channel list
      include_threads: true            # Include thread replies
      since_days: 7                    # Look back days

  - type: "jira"                       # Jira connector
    name: "Company Jira"
    enabled: true
    config:
      url: "https://company.atlassian.net"  # Jira instance URL
      username: "${JIRA_USERNAME}"     # Jira username
      api_token: "${JIRA_API_TOKEN}"   # Jira API token
      projects: ["PROJ", "DEV"]        # Project keys
      issue_types: ["Bug", "Task"]     # Issue types to include
      since_days: 14                   # Look back days

# Development configuration
development:
  hot_reload: false                    # Enable hot reload
  debug_mode: false                    # Enable debug mode
  profiling:
    enabled: false                     # Enable profiling
    port: 6060                         # Profiling port
  testing:
    mock_external: true                # Mock external services in tests
    parallel: true                     # Run tests in parallel
    verbose: false                     # Verbose test output
```

## Environment Variables Reference

### Core Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_CONFIG` | string | `config.yml` | Configuration file path |
| `CONEXUS_LOG_LEVEL` | string | `info` | Log level |
| `CONEXUS_LOG_FORMAT` | string | `text` | Log format |
| `CONEXUS_LOG_FILE` | string | - | Log file path |

### Database and Storage

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_DB_PATH` | string | `~/.conexus/db.sqlite` | SQLite database path |
| `CONEXUS_VECTORSTORE_TYPE` | string | `sqlite` | Vector store type |
| `CONEXUS_VECTORSTORE_MEMORY_LIMIT` | string | `512MB` | Vector store memory limit |

### Server Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_PORT` | int | `0` | HTTP server port (0 = stdio only) |
| `CONEXUS_HOST` | string | `localhost` | Server bind address |
| `CONEXUS_ROOT_PATH` | string | Current directory | Project root path |
| `CONEXUS_MAX_REQUEST_SIZE` | string | `10MB` | Maximum request size |

### Embedding Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_EMBEDDING_PROVIDER` | string | `mock` | Embedding provider (mock only for MVP) |
| `CONEXUS_EMBEDDING_MODEL` | string | `mock-384` | Embedding model (mock-384 only for MVP) |
| `CONEXUS_EMBEDDING_DIMENSIONS` | int | `384` | Vector dimensions (384 only for MVP) |
| `OPENAI_API_KEY` | string | - | OpenAI API key (post-MVP) |
| `ANTHROPIC_API_KEY` | string | - | Anthropic API key (post-MVP) |

### Indexing Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_INDEXING_CHUNK_SIZE` | int | `500` | Text chunk size |
| `CONEXUS_INDEXING_WORKERS` | int | `2` | Number of workers |
| `CONEXUS_INDEXING_MEMORY_LIMIT` | string | `256MB` | Memory limit |
| `CONEXUS_INDEXING_AUTO_REINDEX` | bool | `true` | Auto reindex |
| `CONEXUS_INDEXING_REINDEX_INTERVAL` | string | `1h` | Reindex interval |

### Search Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_SEARCH_MAX_RESULTS` | int | `50` | Max results |
| `CONEXUS_SEARCH_SIMILARITY_THRESHOLD` | float | `0.7` | Similarity threshold |
| `CONEXUS_SEARCH_CACHE_ENABLED` | bool | `true` | Enable caching |
| `CONEXUS_SEARCH_CACHE_TTL` | string | `1h` | Cache TTL |

### Security Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_RATE_LIMIT_ENABLED` | bool | `false` | Enable rate limiting |
| `CONEXUS_RATE_LIMIT_ALGORITHM` | string | `sliding_window` | Rate limit algorithm |
| `CONEXUS_RATE_LIMIT_DEFAULT_REQUESTS` | int | `100` | Default requests |
| `CONEXUS_RATE_LIMIT_DEFAULT_WINDOW` | string | `1m` | Default window |
| `CONEXUS_RATE_LIMIT_REDIS_ENABLED` | bool | `false` | Enable Redis |
| `CONEXUS_RATE_LIMIT_REDIS_ADDR` | string | `localhost:6379` | Redis address |

### TLS Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_TLS_ENABLED` | bool | `false` | Enable TLS |
| `CONEXUS_TLS_CERT_FILE` | string | - | Certificate file |
| `CONEXUS_TLS_KEY_FILE` | string | - | Private key file |
| `CONEXUS_TLS_AUTO_CERT` | bool | `false` | Auto Let's Encrypt |
| `CONEXUS_TLS_AUTO_CERT_DOMAINS` | string | - | Domains (comma-separated) |
| `CONEXUS_TLS_AUTO_CERT_EMAIL` | string | - | Email for Let's Encrypt |

### Observability Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONEXUS_METRICS_ENABLED` | bool | `false` | Enable metrics |
| `CONEXUS_METRICS_PORT` | int | `9090` | Metrics port |
| `CONEXUS_TRACING_ENABLED` | bool | `false` | Enable tracing |
| `CONEXUS_SENTRY_ENABLED` | bool | `false` | Enable Sentry |
| `CONEXUS_SENTRY_DSN` | string | - | Sentry DSN |

## Configuration Validation

Conexus validates configuration on startup. Invalid configurations will prevent startup with detailed error messages.

### Validation Rules

- **File paths**: Must be readable/writable as appropriate
- **URLs**: Must be valid HTTP/HTTPS URLs
- **Durations**: Must be valid Go duration strings (e.g., "30s", "5m", "1h")
- **Sizes**: Must be valid size strings (e.g., "10MB", "512KB")
- **Percentages**: Must be between 0.0 and 1.0
- **Ports**: Must be between 1 and 65535

### Example Validation Errors

```
Error: invalid configuration
- indexing.chunk_size: must be positive
- embedding.api_key: required when provider is 'openai'
- security.tls.cert_file: file does not exist
- search.similarity_threshold: must be between 0.0 and 1.0
```

## Configuration Precedence

Configuration values are resolved in this order (highest to lowest priority):

1. **Command-line flags** (if available)
2. **Environment variables**
3. **Configuration file**
4. **Built-in defaults**

### Example Precedence

```bash
# Configuration file sets log_level to "info"
# Environment variable overrides to "debug"
export CONEXUS_LOG_LEVEL=debug

# Result: log level is "debug"
```

## Configuration Hot Reload

Some configuration changes can be applied without restarting Conexus:

### Reloadable Settings

- Log level (`CONEXUS_LOG_LEVEL`)
- Search cache settings
- Rate limiting settings (without Redis)
- Some observability settings

### Non-Reloadable Settings

- Database paths
- Server ports
- TLS certificates
- Embedding provider configuration

### Triggering Reload

```bash
# Send SIGHUP signal (Unix-like systems)
kill -HUP $(pgrep conexus)

# Or restart the process
# Changes take effect on next request
```

## Configuration Examples

### Development Configuration

```yaml
project:
  name: "my-app-dev"
  environment: "development"

codebase:
  include_patterns:
    - "**/*.go"
    - "**/*.js"

indexing:
  auto_reindex: true
  reindex_interval: "10s"  # Fast reindexing for development

search:
  max_results: 100
  cache_enabled: false  # Disable caching for development

observability:
  log_level: "debug"
  metrics:
    enabled: true
    port: 9090

development:
  hot_reload: true
  debug_mode: true
```

### Production Configuration

```yaml
project:
  name: "my-app-prod"
  environment: "production"

security:
  rate_limiting:
    enabled: true
    default_requests: 1000
    default_window: "1m"
    redis:
      enabled: true
      addr: "redis:6379"

  tls:
    enabled: true
    auto_cert: true
    auto_cert_domains: ["api.mycompany.com"]
    auto_cert_email: "admin@mycompany.com"

observability:
  log_level: "info"
  log_format: "json"
  metrics:
    enabled: true
  tracing:
    enabled: true
  sentry:
    enabled: true
    dsn: "${SENTRY_DSN}"
    environment: "production"
```

### Enterprise Configuration

```yaml
project:
  name: "enterprise-app"
  environment: "production"

security:
  auth:
    enabled: true
    provider: "jwt"
    jwt:
      secret: "${JWT_SECRET}"

connectors:
  - type: "github"
    name: "Enterprise GitHub"
    config:
      token: "${GITHUB_TOKEN}"
      org: "mycompany"
  - type: "jira"
    name: "Enterprise Jira"
    config:
      url: "https://mycompany.atlassian.net"
      username: "${JIRA_USERNAME}"
      api_token: "${JIRA_API_TOKEN}"

observability:
  tracing:
    enabled: true
    provider: "jaeger"
    jaeger:
      endpoint: "http://jaeger-collector:14268/api/traces"
  sentry:
    enabled: true
    sample_rate: 0.1
```

## Configuration Migration

### Migrating from Environment Variables to Config File

1. **Create config file:**
   ```bash
   touch config.yml
   ```

2. **Move environment variables to file:**
   ```yaml
   # Before: export CONEXUS_LOG_LEVEL=debug
   # After:
   observability:
     log_level: "debug"
   ```

3. **Remove environment variables:**
   ```bash
   unset CONEXUS_LOG_LEVEL
   ```

4. **Test configuration:**
   ```bash
   conexus --validate-config
   ```

### Upgrading Configuration Versions

Configuration is backward compatible. New options use sensible defaults.

1. **Check current configuration:**
   ```bash
   conexus --show-config
   ```

2. **Update configuration file with new options**
3. **Validate and test**

## Best Practices

1. **Use configuration files** for complex setups
2. **Store secrets in environment variables**, not config files
3. **Validate configuration** before deployment
4. **Use different configurations** for different environments
5. **Document custom configurations** for team members
6. **Version control** configuration files (without secrets)
7. **Test configuration changes** in staging first
8. **Monitor configuration impact** on performance