
## 🐳 Docker Deployment

### Quick Start with Docker

```bash
# Pull and run the latest image (when available)
docker pull conexus:latest
docker run -d -p 8080:8080 --name conexus conexus:latest

# Or build locally
docker build -t conexus:latest .
docker run -d -p 8080:8080 --name conexus conexus:latest

# Test the service
curl http://localhost:8080/health
```

### Docker Compose (Recommended)

**Production deployment:**

```bash
# Start the service
docker compose up -d

# View logs
docker compose logs -f

# Stop the service
docker compose down

# Rebuild after code changes
docker compose up -d --build
```

**Development deployment:**

```bash
# Use development configuration with debug logging
docker compose -f docker-compose.yml -f docker-compose.dev.yml up

# View debug logs
docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f
```

### Configuration

**Environment Variables:**

```bash
# Server configuration
CONEXUS_HOST=0.0.0.0              # Server bind address
CONEXUS_PORT=8080                  # Server port

# Database configuration
CONEXUS_DB_PATH=/data/conexus.db   # SQLite database path

# Codebase configuration
CONEXUS_ROOT_PATH=/data/codebase   # Path to codebase to index

# Logging configuration
CONEXUS_LOG_LEVEL=info             # Log level (debug|info|warn|error)
CONEXUS_LOG_FORMAT=json            # Log format (json|text)

# Embedding configuration (optional)
CONEXUS_EMBEDDING_PROVIDER=openai  # Embedding provider (mock|openai)
CONEXUS_EMBEDDING_MODEL=text-embedding-3-small
OPENAI_API_KEY=sk-...              # OpenAI API key
```

**Volume Mounts:**

```yaml
volumes:
  # Persistent database storage
  - ./data:/data
  
  # Optional: Mount your codebase for indexing
  - /path/to/your/code:/data/codebase:ro
  
  # Optional: Mount config file
  - ./config.yml:/app/config.yml:ro
```

### Docker Image Details

**Multi-stage build:**
- **Builder**: `golang:1.24-alpine` (CGO enabled for SQLite)
- **Runtime**: `alpine:3.19` (minimal base, ca-certificates + sqlite-libs)

**Image specifications:**
- **Size**: ~19.5MB (optimized with multi-stage build)
- **User**: Non-root `conexus:1000`
- **Port**: 8080 (HTTP + MCP over JSON-RPC 2.0)
- **Health Check**: `GET /health` every 30s

**Security features:**
- Non-root execution (UID 1000)
- Static binary (no dynamic linking)
- Minimal attack surface (Alpine base)
- Read-only config option
- Health check monitoring

### MCP Server Endpoints

Once running, the service exposes:

**HTTP Endpoints:**
```bash
# Health check
curl http://localhost:8080/health
# Response: {"status":"healthy","version":"0.1.0-alpha"}

# Service info
curl http://localhost:8080/
# Response: Service info with MCP endpoint

# MCP JSON-RPC endpoint
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
```

**MCP Tools:**
1. `context.search` - Comprehensive search with filters
2. `context.get_related_info` - File/ticket context retrieval
3. `context.index_control` - Indexing operations
4. `context.connector_management` - Data source management

### Production Deployment

**With Docker Compose:**

```yaml
# docker-compose.prod.yml
services:
  conexus:
    image: conexus:latest
    restart: always
    environment:
      - CONEXUS_LOG_LEVEL=info
      - CONEXUS_LOG_FORMAT=json
    volumes:
      - conexus-data:/data
      - /mnt/codebase:/data/codebase:ro
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

volumes:
  conexus-data:
    driver: local
```

**Deploy:**
```bash
docker compose -f docker-compose.prod.yml up -d
```

### Monitoring

**Check health:**
```bash
# Container status
docker compose ps

# Health check status
docker inspect conexus | jq '.[0].State.Health'

# View logs
docker compose logs -f

# Check metrics
curl http://localhost:8080/health
```

**Troubleshooting:**
```bash
# View container logs
docker compose logs --tail=100

# Execute commands in container
docker compose exec conexus sh

# Check database
docker compose exec conexus ls -la /data/

# Restart service
docker compose restart
```

### Building from Source

```bash
# Build Docker image
docker build -t conexus:custom .

# Build with specific Go version
docker build --build-arg GO_VERSION=1.24 -t conexus:custom .

# Build and tag
docker build -t conexus:v0.1.0 -t conexus:latest .

# Push to registry (configure your registry)
docker tag conexus:latest registry.example.com/conexus:latest
docker push registry.example.com/conexus:latest
```

### Docker Best Practices

1. **Use Docker Compose** for orchestration
2. **Mount volumes** for data persistence
3. **Configure environment variables** for secrets
4. **Enable health checks** for monitoring
5. **Use named volumes** in production
6. **Check logs regularly** with `docker compose logs`
7. **Backup database** in `/data` directory regularly
8. **Limit resources** with Docker resource constraints if needed

---

