# Operations Guide

## Overview

This operations guide provides comprehensive instructions for deploying, configuring, monitoring, and maintaining the Agentic Context Engine (Conexus) in production environments. Conexus is a RAG-based context system that requires careful operational management to ensure reliability, performance, and security.

## Prerequisites

### System Requirements

#### Minimum Production Requirements

- **CPU**: 4-core CPU (Intel/AMD x64)
- **Memory**: 16GB RAM
- **Storage**: 100GB SSD storage
- **Network**: 1Gbps network interface

#### Recommended Production Requirements

- **CPU**: 8-core CPU with AVX2 support
- **Memory**: 32GB RAM
- **Storage**: 500GB NVMe SSD
- **Network**: 10Gbps network interface

### Software Dependencies

#### Runtime Dependencies

```bash
# Required packages
sudo apt-get update
sudo apt-get install -y \
    postgresql-15 \
    redis-server \
    docker.io \
    docker-compose \
    nginx \
    certbot \
    ufw \
    fail2ban \
    logrotate \
    prometheus \
    grafana \
    jaeger
```

#### Development Tools

```bash
# Development dependencies
sudo apt-get install -y \
    git \
    curl \
    wget \
    htop \
    iotop \
    ncdu \
    tree \
    jq \
    yq
```

## Installation Procedures

### Automated Installation

#### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/conexus.git
cd conexus

# Copy environment configuration
cp .env.example .env.production

# Edit configuration
vim .env.production

# Deploy with Docker Compose
docker-compose -f docker-compose.production.yml up -d

# Run health checks
./scripts/health-check.sh
```

#### Configuration Variables

```bash
# Database Configuration
POSTGRES_HOST=conexus-postgres
POSTGRES_PORT=5432
POSTGRES_DB=ace_production
POSTGRES_USER=ace_user
POSTGRES_PASSWORD=<secure-password>

# Vector Database Configuration
QDRANT_HOST=conexus-qdrant
QDRANT_PORT=6333
QDRANT_API_KEY=<secure-api-key>

# Redis Configuration
REDIS_HOST=conexus-redis
REDIS_PORT=6379
REDIS_PASSWORD=<secure-password>

# Application Configuration
Conexus_HOST=0.0.0.0
Conexus_PORT=8080
Conexus_LOG_LEVEL=info
Conexus_ENVIRONMENT=production

# Security Configuration
Conexus_JWT_SECRET=<secure-jwt-secret>
Conexus_ENCRYPTION_KEY=<secure-encryption-key>

# Monitoring Configuration
OTEL_EXPORTER_OTLP_ENDPOINT=http://conexus-jaeger:14268/api/traces
METRICS_PORT=9090
```

### Manual Installation

#### Database Setup

```sql
-- Create production database
CREATE DATABASE ace_production;
CREATE USER ace_user WITH PASSWORD '<secure-password>';
GRANT ALL PRIVILEGES ON DATABASE ace_production TO ace_user;

-- Enable required extensions
\c ace_production;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "vector";
```

#### Vector Database Setup

```bash
# Start Qdrant
docker run -d \
  --name conexus-qdrant \
  -p 6333:6333 \
  -v ace_qdrant_storage:/qdrant/storage \
  qdrant/qdrant

# Configure collections
curl -X PUT "http://localhost:6333/collections/code_context" \
  -H "Content-Type: application/json" \
  -d '{
    "vectors": {
      "size": 768,
      "distance": "Cosine"
    }
  }'
```

#### Application Setup

```bash
# Build the application
go mod download
go build -o conexus-server ./cmd/server

# Create systemd service
sudo tee /etc/systemd/system/conexus.service > /dev/null <<EOF
[Unit]
Description=Conexus Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=conexus
WorkingDirectory=/opt/conexus
ExecStart=/opt/conexus/conexus-server
Restart=always
RestartSec=5
EnvironmentFile=/opt/conexus/.env.production

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/conexus/logs /opt/conexus/data

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable conexus
sudo systemctl start conexus
```

## Configuration Management

### Environment-Based Configuration

#### Production Environment Template

```bash
# /opt/conexus/.env.production
Conexus_ENVIRONMENT=production
Conexus_LOG_LEVEL=warn
Conexus_LOG_FORMAT=json

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=ace_production
POSTGRES_USER=ace_user
POSTGRES_PASSWORD=<secure-password>
POSTGRES_SSLMODE=require
POSTGRES_MAX_CONNECTIONS=100

# Vector Database
QDRANT_HOST=localhost
QDRANT_PORT=6333
QDRANT_API_KEY=<secure-api-key>
QDRANT_TIMEOUT=30s

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=<secure-password>
REDIS_DB=0
REDIS_MAX_CONNECTIONS=50

# Application
Conexus_HOST=0.0.0.0
Conexus_PORT=8080
Conexus_READ_TIMEOUT=30s
Conexus_WRITE_TIMEOUT=30s
Conexus_MAX_REQUEST_SIZE=10MB

# Security
Conexus_JWT_SECRET=<secure-jwt-secret>
Conexus_ENCRYPTION_KEY=<secure-encryption-key>
Conexus_CORS_ALLOWED_ORIGINS=https://your-domain.com
Conexus_RATE_LIMIT_REQUESTS=1000
Conexus_RATE_LIMIT_WINDOW=1m

# Monitoring
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318/v1/traces
METRICS_PORT=9090
HEALTH_CHECK_PATH=/health

# Feature Flags
Conexus_ENABLE_DEBUG_ENDPOINTS=false
Conexus_ENABLE_PPROF=false
Conexus_ENABLE_METRICS=true
```

### Configuration Validation

#### Automated Validation Script

```bash
#!/bin/bash
# /opt/conexus/scripts/validate-config.sh

set -e

echo "Validating Conexus configuration..."

# Check required environment variables
REQUIRED_VARS=(
    "POSTGRES_HOST"
    "POSTGRES_PASSWORD"
    "QDRANT_API_KEY"
    "Conexus_JWT_SECRET"
    "Conexus_ENCRYPTION_KEY"
)

for var in "${REQUIRED_VARS[@]}"; do
    if [[ -z "${!var}" ]]; then
        echo "ERROR: Required environment variable $var is not set"
        exit 1
    fi
done

# Validate database connection
echo "Testing database connection..."
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "SELECT 1;" > /dev/null

if [[ $? -eq 0 ]]; then
    echo "✓ Database connection successful"
else
    echo "✗ Database connection failed"
    exit 1
fi

# Validate Qdrant connection
echo "Testing Qdrant connection..."
curl -s "http://$QDRANT_HOST:$QDRANT_PORT/health" > /dev/null

if [[ $? -eq 0 ]]; then
    echo "✓ Qdrant connection successful"
else
    echo "✗ Qdrant connection failed"
    exit 1
fi

echo "Configuration validation completed successfully!"
```

### Configuration Deployment

#### Zero-Downtime Configuration Updates

```bash
#!/bin/bash
# /opt/conexus/scripts/update-config.sh

set -e

NEW_CONFIG="$1"
if [[ -z "$NEW_CONFIG" ]]; then
    echo "Usage: $0 <new-config-file>"
    exit 1
fi

echo "Deploying new configuration..."

# Validate new configuration
/opt/conexus/scripts/validate-config.sh "$NEW_CONFIG"

# Backup current configuration
cp /opt/conexus/.env.production /opt/conexus/.env.production.backup.$(date +%Y%m%d_%H%M%S)

# Deploy new configuration
cp "$NEW_CONFIG" /opt/conexus/.env.production

# Reload application configuration
sudo systemctl reload conexus

# Verify application health
sleep 10
curl -f http://localhost:8080/health

if [[ $? -eq 0 ]]; then
    echo "✓ Configuration updated successfully"
else
    echo "✗ Configuration update failed, rolling back..."
    cp /opt/conexus/.env.production.backup.$(date +%Y%m%d_%H%M%S) /opt/conexus/.env.production
    sudo systemctl reload conexus
    exit 1
fi
```

## Monitoring Setup

### OpenTelemetry Integration

#### Collector Configuration

```yaml
# /opt/conexus/otel-collector-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
  hostmetrics:
    collection_interval: 60s
    scrapers:
      cpu:
        metrics:
          system.cpu.utilization:
            enabled: true
      memory:
        metrics:
          system.memory.utilization:
            enabled: true
      disk:
        metrics:
          system.disk.operation_time:
            enabled: true

processors:
  batch:
    send_batch_size: 1024
    timeout: 1s
  memory_limiter:
    limit_mib: 512
    spike_limit_mib: 128

exporters:
  otlp:
    endpoint: "http://jaeger:14268/api/traces"
    insecure: true
  prometheus:
    endpoint: "0.0.0.0:9090"

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [otlp]
    metrics:
      receivers: [hostmetrics]
      processors: [memory_limiter, batch]
      exporters: [prometheus]
```

#### Application Instrumentation

```go
// Initialize OpenTelemetry
func initOpenTelemetry() (*oteltrace.TracerProvider, error) {
    ctx := context.Background()

    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String("conexus-server"),
        semconv.ServiceVersionKey.String("1.0.0"),
        attribute.String("environment", getEnv("Conexus_ENVIRONMENT", "development")),
    )

    conn, err := grpc.DialContext(ctx, "localhost:4317",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
    }

    traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
    if err != nil {
        return nil, fmt.Errorf("failed to create trace exporter: %w", err)
    }

    tracerProvider := oteltrace.NewTracerProvider(
        oteltrace.WithBatcher(traceExporter),
        oteltrace.WithResource(res),
    )

    otel.SetTracerProvider(tracerProvider)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tracerProvider, nil
}
```

### Metrics Collection

#### Prometheus Metrics

```go
// Custom metrics registration
func registerMetrics() {
    // Request metrics
    requestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ace_requests_total",
            Help: "Total number of requests processed",
        },
        []string{"method", "endpoint", "status"},
    )

    // Latency metrics
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ace_request_duration_seconds",
            Help:    "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // Context retrieval metrics
    contextRetrievalDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ace_context_retrieval_duration_seconds",
            Help:    "Context retrieval duration in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
        },
        []string{"retrieval_type"},
    )
}
```

#### Grafana Dashboards

Key dashboard panels:

1. **Request Volume**: Requests per second, error rates
2. **Latency Distribution**: P50, P95, P99 latencies
3. **Context Retrieval Performance**: Retrieval times, cache hit rates
4. **System Resources**: CPU, memory, disk usage
5. **Database Performance**: Connection pool, query performance
6. **Vector Database Metrics**: Index performance, storage usage

### Logging Configuration

#### Structured Logging

```go
// Production logging configuration
logger := slog.New(
    slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
        ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
            if a.Key == slog.TimeKey {
                return slog.Attr{
                    Key:   "timestamp",
                    Value: slog.StringValue(time.Now().UTC().Format(time.RFC3339)),
                }
            }
            return a
        },
    }),
)
```

#### Log Aggregation

```bash
# Fluent Bit configuration for log shipping
[INPUT]
    Name              tail
    Path              /opt/conexus/logs/*.log
    Parser            json
    Tag               conexus.*
    Refresh_Interval  5

[OUTPUT]
    Name  loki
    Match *
    Host  loki.monitoring.svc.cluster.local
    Port  3100
    Labels service=conexus,environment=production
```

## Backup and Recovery

### Backup Strategy

#### Automated Backup Script

```bash
#!/bin/bash
# /opt/conexus/scripts/backup.sh

set -e

BACKUP_DIR="/opt/conexus/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="ace_backup_$TIMESTAMP"

echo "Starting Conexus backup: $BACKUP_NAME"

# Create backup directory
mkdir -p "$BACKUP_DIR/$BACKUP_NAME"

# Backup PostgreSQL database
echo "Backing up PostgreSQL database..."
PGPASSWORD="$POSTGRES_PASSWORD" pg_dump \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -F c \
    -f "$BACKUP_DIR/$BACKUP_NAME/postgres.dump"

# Backup Qdrant data
echo "Backing up Qdrant data..."
docker run --rm \
    -v ace_qdrant_storage:/source \
    -v "$BACKUP_DIR/$BACKUP_NAME":/backup \
    alpine tar czf /backup/qdrant.tar.gz -C /source ./

# Backup configuration files
echo "Backing up configuration files..."
cp /opt/conexus/.env.production "$BACKUP_DIR/$BACKUP_NAME/config.env"
cp /opt/conexus/otel-collector-config.yaml "$BACKUP_DIR/$BACKUP_NAME/otel-config.yaml"

# Create backup metadata
cat > "$BACKUP_DIR/$BACKUP_NAME/metadata.json" << EOF
{
    "timestamp": "$TIMESTAMP",
    "version": "$(git rev-parse HEAD)",
    "components": ["postgres", "qdrant", "config"],
    "size": "$(du -sb "$BACKUP_DIR/$BACKUP_NAME" | cut -f1)"
}
EOF

# Compress backup
echo "Compressing backup..."
cd "$BACKUP_DIR"
tar czf "$BACKUP_NAME.tar.gz" "$BACKUP_NAME"
rm -rf "$BACKUP_NAME"

# Cleanup old backups (keep last 7 days)
find "$BACKUP_DIR" -name "*.tar.gz" -type f -mtime +7 -delete

echo "Backup completed: $BACKUP_DIR/$BACKUP_NAME.tar.gz"
```

### Recovery Procedures

#### Full System Recovery

```bash
#!/bin/bash
# /opt/conexus/scripts/restore.sh

set -e

BACKUP_FILE="$1"
if [[ -z "$BACKUP_FILE" ]]; then
    echo "Usage: $0 <backup-file.tar.gz>"
    exit 1
fi

echo "Starting Conexus restoration from: $BACKUP_FILE"

# Stop services
sudo systemctl stop conexus

# Extract backup
BACKUP_DIR="/tmp/ace_restore_$(date +%s)"
mkdir -p "$BACKUP_DIR"
tar xzf "$BACKUP_FILE" -C "$BACKUP_DIR"

# Restore PostgreSQL database
echo "Restoring PostgreSQL database..."
PGPASSWORD="$POSTGRES_PASSWORD" pg_restore \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "$BACKUP_DIR/*/postgres.dump"

# Restore Qdrant data
echo "Restoring Qdrant data..."
# Stop Qdrant
sudo docker stop conexus-qdrant
# Restore data
sudo rm -rf /opt/conexus/qdrant/storage/*
sudo tar xzf "$BACKUP_DIR/*/qdrant.tar.gz" -C /opt/conexus/qdrant/storage/
# Start Qdrant
sudo docker start conexus-qdrant

# Restore configuration
echo "Restoring configuration..."
cp "$BACKUP_DIR/*/config.env" /opt/conexus/.env.production

# Start services
sudo systemctl start conexus

# Verify restoration
sleep 10
curl -f http://localhost:8080/health

if [[ $? -eq 0 ]]; then
    echo "✓ Restoration completed successfully"
else
    echo "✗ Restoration failed"
    exit 1
fi

# Cleanup
rm -rf "$BACKUP_DIR"
```

#### Point-in-Time Recovery

```bash
# Restore to specific timestamp
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "SELECT pg_wal_replay_resume();"

# Restore specific table
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "DROP TABLE IF EXISTS corrupted_table;"

PGPASSWORD="$POSTGRES_PASSWORD" pg_restore \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -t corrupted_table \
    /path/to/backup.dump
```

## Incident Response

### Incident Response Plan

#### Severity Levels

1. **SEV-1 (Critical)**: Complete system outage, data loss, security breach
2. **SEV-2 (High)**: Significant performance degradation, partial outage
3. **SEV-3 (Medium)**: Minor issues affecting user experience
4. **SEV-4 (Low)**: Non-critical issues, feature requests

#### Response Procedures

##### SEV-1 Incident Response

```bash
#!/bin/bash
# /opt/conexus/scripts/incident-response-sev1.sh

INCIDENT_ID="$1"
if [[ -z "$INCIDENT_ID" ]]; then
    echo "Usage: $0 <incident-id>"
    exit 1
fi

echo "Initiating SEV-1 incident response for: $INCIDENT_ID"

# Notify on-call team
/opt/conexus/scripts/notify-team.sh "$INCIDENT_ID" "SEV-1"

# Capture system state
/opt/conexus/scripts/capture-system-state.sh "$INCIDENT_ID"

# Isolate affected components
sudo systemctl stop conexus
sudo docker pause conexus-qdrant

# Activate backup systems
sudo systemctl start conexus-backup

# Begin investigation
echo "Incident timeline:"
echo "$(date): SEV-1 incident declared"
echo "$(date): System state captured"
echo "$(date): Backup systems activated"
```

##### System State Capture

```bash
#!/bin/bash
# /opt/conexus/scripts/capture-system-state.sh

INCIDENT_ID="$1"
CAPTURE_DIR="/opt/conexus/incidents/$INCIDENT_ID"

mkdir -p "$CAPTURE_DIR"

echo "Capturing system state for incident: $INCIDENT_ID"

# Capture process information
top -b -n1 > "$CAPTURE_DIR/top.txt"
ps auxf > "$CAPTURE_DIR/ps.txt"

# Capture system logs
journalctl -u conexus --since "1 hour ago" > "$CAPTURE_DIR/conexus.log"
journalctl -u postgresql --since "1 hour ago" > "$CAPTURE_DIR/postgres.log"

# Capture application metrics
curl http://localhost:9090/metrics > "$CAPTURE_DIR/metrics.txt"

# Capture database status
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "SELECT * FROM pg_stat_activity;" > "$CAPTURE_DIR/pg_stat_activity.txt"

echo "System state captured in: $CAPTURE_DIR"
```

### Post-Incident Review

#### Automated Report Generation

```bash
#!/bin/bash
# /opt/conexus/scripts/generate-incident-report.sh

INCIDENT_ID="$1"
if [[ -z "$INCIDENT_ID" ]]; then
    echo "Usage: $0 <incident-id>"
    exit 1
fi

INCIDENT_DIR="/opt/conexus/incidents/$INCIDENT_ID"
REPORT_FILE="/opt/conexus/reports/incident_$INCIDENT_ID.md"

echo "Generating incident report for: $INCIDENT_ID"

cat > "$REPORT_FILE" << EOF
# Incident Report: $INCIDENT_ID

## Timeline
$(cat "$INCIDENT_DIR/timeline.txt")

## Root Cause Analysis
$(cat "$INCIDENT_DIR/root_cause.txt")

## Impact Assessment
$(cat "$INCIDENT_DIR/impact.txt")

## Resolution Steps
$(cat "$INCIDENT_DIR/resolution.txt")

## Preventive Measures
$(cat "$INCIDENT_DIR/prevention.txt")

## Lessons Learned
$(cat "$INCIDENT_DIR/lessons.txt")

## Action Items
$(cat "$INCIDENT_DIR/actions.txt")

---
*Report generated: $(date)*
*Incident duration: $(cat "$INCIDENT_DIR/duration.txt")*
EOF

echo "Incident report generated: $REPORT_FILE"
```

## Security Operations

### Access Control

#### Service Account Management

```bash
# Create deployment user
sudo useradd -r -s /bin/false -d /opt/conexus conexus

# Set proper permissions
sudo chown -R conexus:conexus /opt/conexus
sudo chmod -R 750 /opt/conexus

# Configure sudo access for specific commands
sudo tee /etc/sudoers.d/conexus > /dev/null <<EOF
conexus ALL=(ALL) NOPASSWD: /bin/systemctl start conexus, /bin/systemctl stop conexus, /bin/systemctl restart conexus
conexus ALL=(ALL) NOPASSWD: /bin/systemctl status conexus
EOF
```

#### Network Security

```bash
# Configure firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow specific ports
sudo ufw allow from 10.0.0.0/8 to any port 22   # SSH from internal network
sudo ufw allow from 10.0.0.0/8 to any port 8080 # Conexus API from internal network
sudo ufw allow 443                          # HTTPS for external access

sudo ufw enable

# Configure fail2ban
sudo tee /etc/fail2ban/jail.local > /dev/null <<EOF
[conexus-api]
enabled = true
port = 8080
filter = conexus-api
logpath = /opt/conexus/logs/access.log
maxretry = 5
bantime = 3600

[conexus-ssh]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF
```

### Data Protection

#### Encryption at Rest

```bash
# Configure encrypted storage for backups
sudo apt-get install -y ecryptfs-utils

# Create encrypted backup directory
sudo mkdir -p /opt/conexus/encrypted_backups
sudo mount -t ecryptfs /opt/conexus/encrypted_backups /opt/conexus/encrypted_backups \
    -o key=passphrase:passphrase_passwd=<backup-encryption-key>,ecryptfs_cipher=aes,ecryptfs_key_bytes=32,ecryptfs_passthrough=no,ecryptfs_enable_filename_crypto=yes
```

#### TLS Configuration

```bash
# Generate self-signed certificate (development)
openssl req -x509 -newkey rsa:4096 -keyout conexus.key -out conexus.crt \
    -days 365 -nodes -subj "/CN=conexus.local"

# Configure nginx with TLS
sudo tee /etc/nginx/sites-available/conexus > /dev/null <<EOF
server {
    listen 80;
    server_name conexus.your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name conexus.your-domain.com;

    ssl_certificate /etc/ssl/certs/conexus.crt;
    ssl_certificate_key /etc/ssl/private/conexus.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /metrics {
        proxy_pass http://localhost:9090;
        allow 10.0.0.0/8;
        deny all;
    }
}
EOF
```

## Performance Optimization

### Resource Tuning

#### PostgreSQL Optimization

```ini
# /etc/postgresql/15/main/postgresql.conf
# Memory settings
shared_buffers = 1GB
effective_cache_size = 4GB
work_mem = 64MB
maintenance_work_mem = 512MB

# Connection settings
max_connections = 200
max_prepared_transactions = 200

# WAL settings
wal_buffers = 16MB
checkpoint_segments = 128

# Query optimization
random_page_cost = 1.5
effective_io_concurrency = 200
```

#### Redis Optimization

```conf
# /etc/redis/redis.conf
# Memory settings
maxmemory 2GB
maxmemory-policy allkeys-lru

# Performance settings
tcp-keepalive 300
timeout 300
tcp-keepalive 60

# Persistence settings
save 900 1
save 300 10
save 60 10000
```

### Scaling Considerations

#### Horizontal Scaling

```bash
# Load balancer configuration (nginx)
upstream ace_backend {
    least_conn;
    server conexus-01:8080 max_fails=3 fail_timeout=30s;
    server conexus-02:8080 max_fails=3 fail_timeout=30s;
    server conexus-03:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    location / {
        proxy_pass http://ace_backend;
        proxy_next_upstream error timeout invalid_header http_500 http_502 http_503 http_504;
    }
}
```

#### Database Scaling

```sql
-- Read replica setup
CREATE PUBLICATION ace_publication FOR ALL TABLES;

-- On replica
CREATE SUBSCRIPTION ace_subscription
    CONNECTION 'host=primary-host port=5432 user=replica_user dbname=ace_production'
    PUBLICATION ace_publication;
```

## Maintenance Procedures

### Regular Maintenance Tasks

#### Daily Tasks

```bash
#!/bin/bash
# /opt/conexus/scripts/daily-maintenance.sh

echo "Running daily maintenance tasks..."

# Clean old log files
find /opt/conexus/logs -name "*.log" -type f -mtime +7 -delete

# Update system packages
sudo apt-get update && sudo apt-get upgrade -y

# Check disk usage
df -h | grep -v tmpfs > /opt/conexus/logs/disk_usage.log

# Verify backups
ls -la /opt/conexus/backups/ | tail -5 > /opt/conexus/logs/backup_status.log

echo "Daily maintenance completed at $(date)"
```

#### Weekly Tasks

```bash
#!/bin/bash
# /opt/conexus/scripts/weekly-maintenance.sh

echo "Running weekly maintenance tasks..."

# Database maintenance
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "VACUUM ANALYZE;"

# Rebuild search indexes
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "REINDEX INDEX CONCURRENTLY idx_code_chunks_embedding;"

# Update virus definitions (if applicable)
sudo freshclam

# Generate maintenance report
/opt/conexus/scripts/generate-maintenance-report.sh

echo "Weekly maintenance completed at $(date)"
```

### Health Monitoring

#### Health Check Endpoints

```go
// Health check handler
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   version,
    }

    // Check database connectivity
    if err := checkDatabase(ctx); err != nil {
        health["status"] = "unhealthy"
        health["database"] = err.Error()
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    // Check Qdrant connectivity
    if err := checkQdrant(ctx); err != nil {
        health["status"] = "unhealthy"
        health["qdrant"] = err.Error()
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

#### Automated Health Monitoring

```bash
#!/bin/bash
# /opt/conexus/scripts/health-monitor.sh

HEALTH_URL="http://localhost:8080/health"
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

while true; do
    if ! curl -f -s "$HEALTH_URL" > /dev/null; then
        # Send alert
        curl -X POST -H 'Content-type: application/json' \
            --data '{"text":"Conexus health check failed"}' \
            "$SLACK_WEBHOOK"
    fi
    sleep 60
done
```

## Troubleshooting

### Common Issues

#### High Latency Issues

```bash
# Diagnose latency issues
curl -w "@-" -o /dev/null -s "http://localhost:8080/health" <<'EOF'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF
```

#### Memory Issues

```bash
# Check memory usage
free -h
cat /proc/meminfo | grep -E "(MemTotal|MemFree|Buffers|Cached)"

# Check for memory leaks
valgrind --tool=memcheck --leak-check=full ./conexus-server
```

#### Database Connection Issues

```bash
# Check PostgreSQL connections
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';"

# Check connection pool
ss -tuln | grep :5432
```

### Debug Mode

#### Enable Debug Logging

```bash
# Temporarily enable debug logging
export Conexus_LOG_LEVEL=debug
sudo systemctl reload conexus

# Capture debug information
curl -v http://localhost:8080/debug/pprof/goroutine > goroutines.txt
curl -v http://localhost:8080/debug/pprof/heap > heap.txt
```

## Conclusion

This operations guide provides comprehensive coverage for deploying and maintaining Conexus in production environments. Regular monitoring, automated backups, and proper incident response procedures ensure high availability and reliability.

For environment-specific configurations or additional operational requirements, consult with the infrastructure team and update this guide accordingly.

## Support

For operational issues or questions:

- **On-call**: operations@your-company.com
- **Documentation**: Update this guide for new procedures
- **Escalation**: Follow the incident response procedures above