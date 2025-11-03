# Advanced Configuration Guide

This guide covers advanced Conexus configuration options, performance tuning, security settings, and enterprise deployment scenarios.

## Environment Variables Reference

### Core Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_DB_PATH` | `~/.conexus/db.sqlite` | SQLite database location |
| `CONEXUS_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `CONEXUS_LOG_FORMAT` | `text` | Log format: `text`, `json` |
| `CONEXUS_PORT` | `0` | HTTP server port (0 = stdio only) |
| `CONEXUS_ROOT_PATH` | Current directory | Project root directory |
| `CONEXUS_CONFIG` | `config.yml` | Configuration file path |

### Embedding Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_EMBEDDING_PROVIDER` | `mock` | Provider: `mock` (only for MVP) |
| `CONEXUS_EMBEDDING_MODEL` | `mock-384` | Model name (mock-384 only for MVP) |
| `CONEXUS_EMBEDDING_DIMENSIONS` | `384` | Vector dimensions (384 only for MVP) |
| `CONEXUS_EMBEDDING_API_KEY` | - | API key for cloud providers |
| `OPENAI_API_KEY` | - | OpenAI API key |
| `ANTHROPIC_API_KEY` | - | Anthropic API key |

### Indexing Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_INDEXING_CHUNK_SIZE` | `500` | Text chunk size for indexing |
| `CONEXUS_INDEXING_WORKERS` | `2` | Number of indexing workers |
| `CONEXUS_INDEXING_MEMORY_LIMIT` | `256MB` | Memory limit for indexing |
| `CONEXUS_INDEXING_AUTO_REINDEX` | `true` | Enable automatic reindexing |
| `CONEXUS_INDEXING_REINDEX_INTERVAL` | `1h` | Reindexing interval |

### Search Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_SEARCH_MAX_RESULTS` | `50` | Maximum search results |
| `CONEXUS_SEARCH_SIMILARITY_THRESHOLD` | `0.7` | Similarity threshold (0.0-1.0) |
| `CONEXUS_SEARCH_CACHE_ENABLED` | `true` | Enable search caching |
| `CONEXUS_SEARCH_CACHE_TTL` | `1h` | Cache time-to-live |

### Vector Store Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_VECTORSTORE_TYPE` | `sqlite` | Store type: `sqlite`, `memory` |
| `CONEXUS_VECTORSTORE_MEMORY_LIMIT` | `512MB` | Memory limit for vector store |
| `CONEXUS_VECTORSTORE_CACHE_SIZE` | `1000` | Vector cache size |

### Security Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_RATE_LIMIT_ENABLED` | `false` | Enable rate limiting |
| `CONEXUS_RATE_LIMIT_ALGORITHM` | `sliding_window` | Algorithm: `sliding_window`, `token_bucket` |
| `CONEXUS_RATE_LIMIT_DEFAULT_REQUESTS` | `100` | Default requests per window |
| `CONEXUS_RATE_LIMIT_DEFAULT_WINDOW` | `1m` | Default time window |
| `CONEXUS_RATE_LIMIT_REDIS_ENABLED` | `false` | Enable Redis for distributed rate limiting |
| `CONEXUS_RATE_LIMIT_REDIS_ADDR` | `localhost:6379` | Redis address |
| `CONEXUS_RATE_LIMIT_REDIS_PASSWORD` | - | Redis password |

### TLS/HTTPS Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_TLS_ENABLED` | `false` | Enable TLS/HTTPS |
| `CONEXUS_TLS_CERT_FILE` | - | TLS certificate file path |
| `CONEXUS_TLS_KEY_FILE` | - | TLS private key file path |
| `CONEXUS_TLS_AUTO_CERT` | `false` | Enable automatic Let's Encrypt certificates |
| `CONEXUS_TLS_AUTO_CERT_DOMAINS` | - | Comma-separated domain list |
| `CONEXUS_TLS_AUTO_CERT_EMAIL` | - | Email for Let's Encrypt |

### Observability Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CONEXUS_METRICS_ENABLED` | `false` | Enable Prometheus metrics |
| `CONEXUS_METRICS_PORT` | `9090` | Metrics server port |
| `CONEXUS_TRACING_ENABLED` | `false` | Enable distributed tracing |
| `CONEXUS_SENTRY_ENABLED` | `false` | Enable Sentry error tracking |
| `CONEXUS_SENTRY_DSN` | - | Sentry DSN |
| `CONEXUS_SENTRY_ENVIRONMENT` | `development` | Sentry environment |
| `CONEXUS_SENTRY_SAMPLE_RATE` | `1.0` | Sentry sample rate |

## Configuration File Format

Conexus supports YAML configuration files for complex setups:

```yaml
# config.yml
project:
  name: "my-enterprise-app"
  description: "Enterprise web application"
  version: "1.0.0"

codebase:
  root: "."
  include_patterns:
    - "**/*.go"
    - "**/*.js"
    - "**/*.py"
    - "**/*.rs"
    - "**/*.md"
  exclude_patterns:
    - "**/node_modules/**"
    - "**/vendor/**"
    - "**/target/**"
    - "**/.git/**"
    - "**/dist/**"
    - "**/build/**"
    - "**/*.log"

indexing:
  auto_reindex: true
  reindex_interval: "30m"
  chunk_size: 500
  workers: 4
  memory_limit: "1GB"
  ignore_patterns:
    - "**/testdata/**"
    - "**/mocks/**"

search:
  max_results: 100
  similarity_threshold: 0.8
  cache_enabled: true
  cache_ttl: "2h"
  rerank_enabled: true
  rerank_model: "cross-encoder"

embedding:
  provider: "openai"
  model: "text-embedding-3-small"
  dimensions: 1536
  api_key: "${OPENAI_API_KEY}"
  batch_size: 100
  rate_limit: 1000

vectorstore:
  type: "sqlite"
  path: ".conexus/vectors.db"
  memory_limit: "2GB"
  cache_size: 5000
  optimize_on_startup: true

security:
  rate_limiting:
    enabled: true
    algorithm: "sliding_window"
    default_requests: 1000
    default_window: "1m"
    redis:
      enabled: true
      addr: "redis:6379"
      password: "${REDIS_PASSWORD}"

  tls:
    enabled: true
    auto_cert: true
    domains: ["api.mycompany.com", "conexus.mycompany.com"]
    email: "admin@mycompany.com"

observability:
  log_level: "info"
  log_format: "json"
  metrics:
    enabled: true
    port: 9090
    path: "/metrics"
  tracing:
    enabled: true
    jaeger_endpoint: "http://jaeger:14268/api/traces"
  sentry:
    enabled: true
    dsn: "${SENTRY_DSN}"
    environment: "production"
    sample_rate: 0.1
    traces_sample_rate: 0.1

connectors:
  - type: "github"
    name: "Company GitHub"
    config:
      token: "${GITHUB_TOKEN}"
      org: "mycompany"
      repos: ["api", "web", "mobile"]
  - type: "slack"
    name: "Engineering Slack"
    config:
      token: "${SLACK_TOKEN}"
      channels: ["#dev", "#backend", "#frontend"]

mcp:
  server_port: 9090
  allowed_origins: ["https://cursor.sh", "https://claude.ai"]
  auth:
    enabled: true
    jwt_secret: "${JWT_SECRET}"
    token_ttl: "24h"
```

## Performance Tuning

### For Large Codebases (>100K files)

```yaml
indexing:
  chunk_size: 1000  # Larger chunks for better context
  workers: 8        # More workers for parallel processing
  memory_limit: "4GB"
  reindex_interval: "6h"

search:
  max_results: 200
  cache_enabled: true
  cache_ttl: "4h"

vectorstore:
  memory_limit: "8GB"
  cache_size: 10000
```

### For High-Traffic Applications

```yaml
search:
  max_results: 50
  cache_enabled: true
  cache_ttl: "30m"

vectorstore:
  cache_size: 5000

security:
  rate_limiting:
    enabled: true
    default_requests: 10000
    default_window: "1m"
    redis:
      enabled: true
```

### Memory Optimization

```bash
# Environment variables for memory-constrained environments
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=256MB
export CONEXUS_INDEXING_MEMORY_LIMIT=128MB
export CONEXUS_SEARCH_CACHE_TTL=15m
export CONEXUS_INDEXING_WORKERS=1
```

### CPU Optimization

```yaml
# config.yml
indexing:
  workers: 1  # Reduce for CPU-constrained environments
  chunk_size: 250  # Smaller chunks

search:
  max_results: 25  # Limit results
```

## Security Configuration

### Rate Limiting

**Local Rate Limiting:**
```yaml
security:
  rate_limiting:
    enabled: true
    algorithm: "sliding_window"
    default_requests: 100
    default_window: "1m"
    endpoints:
      "/api/search": 1000
      "/api/index": 10
```

**Distributed Rate Limiting with Redis:**
```yaml
security:
  rate_limiting:
    enabled: true
    redis:
      enabled: true
      addr: "redis-cluster:6379"
      password: "${REDIS_PASSWORD}"
      db: 1
```

### TLS/HTTPS Setup

**Manual Certificates:**
```bash
# Generate self-signed for development
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Configure environment
export CONEXUS_TLS_ENABLED=true
export CONEXUS_TLS_CERT_FILE=/path/to/cert.pem
export CONEXUS_TLS_KEY_FILE=/path/to/key.pem
```

**Let's Encrypt Automatic Certificates:**
```yaml
security:
  tls:
    enabled: true
    auto_cert: true
    domains: ["conexus.yourcompany.com"]
    email: "admin@yourcompany.com"
```

### Authentication and Authorization

**JWT Authentication:**
```yaml
mcp:
  auth:
    enabled: true
    jwt_secret: "your-256-bit-secret"
    token_ttl: "24h"
    refresh_token_ttl: "168h"
```

**API Key Authentication:**
```yaml
mcp:
  auth:
    enabled: true
    api_keys:
      - key: "sk-1234567890abcdef"
        name: "cursor-integration"
        permissions: ["read", "search"]
      - key: "sk-abcdef1234567890"
        name: "claude-integration"
        permissions: ["read", "search", "index"]
```

## Enterprise Deployment

### Docker Compose Production Setup

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  conexus:
    image: conexus:latest
    restart: always
    environment:
      - CONEXUS_LOG_LEVEL=info
      - CONEXUS_LOG_FORMAT=json
      - CONEXUS_DB_PATH=/data/conexus.db
      - CONEXUS_CONFIG=/config/config.yml
    volumes:
      - conexus-data:/data
      - ./config:/config:ro
      - /mnt/codebase:/codebase:ro
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp

  redis:
    image: redis:7-alpine
    restart: always
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 5s
      retries: 3

volumes:
  conexus-data:
  redis-data:
```

### Kubernetes Deployment

```yaml
# conexus-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: conexus
spec:
  replicas: 3
  selector:
    matchLabels:
      app: conexus
  template:
    metadata:
      labels:
        app: conexus
    spec:
      containers:
      - name: conexus
        image: conexus:latest
        ports:
        - containerPort: 8080
        env:
        - name: CONEXUS_LOG_LEVEL
          value: "info"
        - name: CONEXUS_RATE_LIMIT_REDIS_ENABLED
          value: "true"
        - name: CONEXUS_RATE_LIMIT_REDIS_ADDR
          value: "redis-service:6379"
        volumeMounts:
        - name: config
          mountPath: /config
          readOnly: true
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: conexus-config
      - name: data
        persistentVolumeClaim:
          claimName: conexus-data

---
apiVersion: v1
kind: Service
metadata:
  name: conexus-service
spec:
  selector:
    app: conexus
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: conexus-config
data:
  config.yml: |
    project:
      name: "enterprise-conexus"
    security:
      rate_limiting:
        enabled: true
        redis:
          enabled: true
```

### Monitoring and Observability

**Prometheus Metrics:**
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'conexus'
    static_configs:
      - targets: ['conexus:8080']
    metrics_path: '/metrics'
```

**Grafana Dashboard:**
```json
{
  "dashboard": {
    "title": "Conexus Metrics",
    "panels": [
      {
        "title": "Search Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(conexus_search_duration_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Index Size",
        "type": "stat",
        "targets": [
          {
            "expr": "conexus_index_documents_total",
            "legendFormat": "Total Documents"
          }
        ]
      }
    ]
  }
}
```

### Load Balancing

**NGINX Configuration:**
```nginx
# nginx.conf
upstream conexus_backend {
    server conexus-1:8080;
    server conexus-2:8080;
    server conexus-3:8080;
}

server {
    listen 80;
    server_name conexus.yourcompany.com;

    location / {
        proxy_pass http://conexus_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Rate limiting
        limit_req zone=api burst=10 nodelay;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
    }

    # Metrics endpoint (internal only)
    location /metrics {
        proxy_pass http://conexus_backend;
        allow 10.0.0.0/8;
        deny all;
    }
}

limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
```

## Backup and Recovery

### Database Backup

```bash
# Backup script
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup
sqlite3 .conexus/db.sqlite ".backup '${BACKUP_DIR}/conexus_${DATE}.db'"

# Compress
gzip "${BACKUP_DIR}/conexus_${DATE}.db"

# Clean old backups (keep last 30 days)
find "$BACKUP_DIR" -name "conexus_*.db.gz" -mtime +30 -delete
```

### Configuration Backup

```bash
# Backup configurations
tar -czf config_backup_$(date +%Y%m%d).tar.gz \
  .conexus/config.yml \
  .opencode/opencode.jsonc \
  .env
```

### Disaster Recovery

```bash
# Restore from backup
gunzip conexus_backup.db.gz
cp conexus_backup.db .conexus/db.sqlite

# Verify integrity
sqlite3 .conexus/db.sqlite "PRAGMA integrity_check;"

# Reindex if needed
./conexus index --force
```

## Migration Guides

### Migrating from Local to Cloud

**Step 1: Export Local Data**
```bash
# Export database
sqlite3 .conexus/db.sqlite ".dump" > conexus_dump.sql

# Export configuration
cp .conexus/config.yml config_backup.yml
```

**Step 2: Set Up Cloud Infrastructure**
```bash
# Create cloud database
aws rds create-db-instance \
  --db-instance-identifier conexus-prod \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --master-username conexus \
  --master-user-password "${DB_PASSWORD}" \
  --allocated-storage 20
```

**Step 3: Update Configuration**
```yaml
# config.yml
database:
  type: "postgres"
  url: "${DATABASE_URL}"

vectorstore:
  type: "qdrant"
  url: "${QDRANT_URL}"
  api_key: "${QDRANT_API_KEY}"
```

**Step 4: Migrate Data**
```bash
# Convert SQLite dump to PostgreSQL
pgloader conexus_dump.sql postgresql://conexus:${DB_PASSWORD}@localhost/conexus

# Migrate vectors (if using Qdrant)
./conexus migrate --from sqlite --to qdrant
```

### Upgrading Conexus Versions

**Minor Version Upgrade:**
```bash
# Stop service
docker compose down

# Update image
docker pull conexus:latest

# Start service
docker compose up -d

# Check logs
docker compose logs -f
```

**Major Version Upgrade:**
```bash
# Backup data
./backup.sh

# Review changelog
curl https://raw.githubusercontent.com/ferg-cod3s/conexus/main/CHANGELOG.md

# Update configuration if needed
vim .conexus/config.yml

# Upgrade
docker pull conexus:v2.0.0
docker compose up -d conexus

# Verify functionality
curl http://localhost:8080/health
./conexus search "test query"
```

### Migrating Between Providers

**From OpenAI to Anthropic:**
```yaml
# Before
embedding:
  provider: "openai"
  model: "text-embedding-3-small"
  api_key: "${OPENAI_API_KEY}"

# After
embedding:
  provider: "anthropic"
  model: "claude-3-haiku-20240307"
  api_key: "${ANTHROPIC_API_KEY}"
```

**Reindex Required:**
```bash
# Clear existing vectors
rm .conexus/vectors.db

# Reindex with new provider
./conexus index --force
```

This advanced configuration guide provides comprehensive setup options for production deployments, performance optimization, and enterprise integration scenarios.