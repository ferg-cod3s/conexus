# Conexus Docker Integration Tests

Comprehensive integration tests for Conexus Docker deployment, validating Docker Compose orchestration, MCP protocol compliance, data persistence, and configuration management.

## Overview

These tests validate the complete Docker deployment stack:

- **Docker Image Build**: Validates image builds successfully and meets size targets
- **Container Lifecycle**: Tests startup, health checks, restarts, and graceful shutdown
- **MCP Protocol**: Validates all 4 MCP tools via JSON-RPC 2.0
- **Configuration**: Tests environment variable loading and config file handling
- **Data Persistence**: Validates database survives container restarts
- **Production & Dev Modes**: Tests both deployment configurations

## Test Categories

### 1. Docker Image Build Tests
- ✅ Image builds successfully
- ✅ Image size is optimal (< 50MB target)
- ✅ Multi-stage build creates minimal runtime image

### 2. Container Lifecycle Tests
- ✅ Container starts successfully (production mode)
- ✅ Health check endpoint responds
- ✅ Container restarts cleanly
- ✅ Graceful shutdown (no panic/fatal errors)

### 3. MCP Protocol Tests
- ✅ `tools/list` endpoint returns valid JSON-RPC 2.0
- ✅ At least 4 tools are available
- ✅ `codebase/locate` tool works end-to-end
- ✅ `codebase/analyze` tool works end-to-end
- ✅ `codebase/search` tool works (implicit via locate)
- ✅ `codebase/index` tool works (implicit via analyze)

### 4. Configuration Tests
- ✅ Environment variables are loaded correctly
- ✅ Server listens on configured port (8080)
- ✅ Log level configuration works
- ✅ Database path configuration works

### 5. Data Persistence Tests
- ✅ Database file is created in mounted volume
- ✅ Database survives container restart
- ✅ Data integrity maintained across restarts
- ✅ Volume mounts work correctly

### 6. Development Mode Tests
- ✅ Dev compose override applies correctly
- ✅ Debug logging enabled in dev mode
- ✅ Local codebase mount works
- ✅ No auto-restart in dev mode

## Prerequisites

### Required Tools
- Docker (>= 20.10)
- Docker Compose (>= 2.0)
- curl
- jq (optional, for enhanced JSON validation)

### Check Prerequisites
```bash
# Check Docker
docker --version

# Check Docker Compose
docker compose version

# Check curl
curl --version

# Check jq (optional but recommended)
jq --version
```

## Running Tests

### Quick Start
```bash
# Run all tests
./tests/integration/docker/docker_test.sh

# Run with automatic cleanup
./tests/integration/docker/docker_test.sh --cleanup
```

### From Project Root
```bash
cd /path/to/conexus

# Run tests
./tests/integration/docker/docker_test.sh

# Run with cleanup
./tests/integration/docker/docker_test.sh --cleanup
```

### Manual Cleanup
```bash
# Stop containers
docker compose down -v

# Remove data directory
rm -rf ./data
```

## Test Execution Flow

```
1. Pre-flight Checks
   └─ Verify Docker, Docker Compose, curl installed

2. Environment Cleanup
   └─ Stop containers, remove volumes, clean data directory

3. Image Build Test
   └─ docker compose build
   └─ Verify image size < 50MB

4. Production Mode Tests
   ├─ docker compose up -d
   ├─ Wait for health check (max 60s)
   ├─ Test MCP tools/list endpoint
   ├─ Test codebase/locate tool
   ├─ Test codebase/analyze tool
   └─ Verify configuration loading

5. Data Persistence Tests
   ├─ Create data via indexing
   ├─ Record database size
   ├─ Restart container
   └─ Verify database integrity

6. Graceful Shutdown Test
   └─ docker compose stop --timeout 10
   └─ Verify no panic/fatal errors

7. Development Mode Tests
   ├─ docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d
   ├─ Verify debug logging
   └─ Verify codebase mount

8. Results Summary
   └─ Display pass/fail counts
```

## Expected Output

### Successful Test Run
```
==========================================
  Conexus Docker Integration Tests
==========================================

[INFO] Running pre-flight checks...
[PASS] Pre-flight checks passed

[INFO] Ensuring clean state...
[PASS] Cleanup complete

[INFO] Test 1: Docker image builds successfully
[PASS] Docker image built successfully
[INFO] Image size: 19.5MB
[PASS] Image size is optimal (< 50MB)

[INFO] Test 2: Container starts successfully (production mode)
[PASS] Container started
[INFO] Waiting for service at http://localhost:8080/health (max 60s)...
[PASS] Service ready after 4s
[PASS] Health check passed

[INFO] Test 3: MCP protocol: tools/list endpoint
[PASS] tools/list endpoint responded
[PASS] Response is valid JSON-RPC 2.0
[PASS] Response contains at least 4 tools
[INFO] Available tools:
  - codebase/locate
  - codebase/analyze
  - codebase/search
  - codebase/index

[INFO] Test 4: MCP protocol: Individual tool invocations
[PASS] codebase/locate tool works
[PASS] codebase/analyze tool works

[INFO] Test 5: Environment variable configuration
[PASS] Server started with configuration
[PASS] Server listening on configured port 8080

[INFO] Test 6: Data persistence across restarts
[INFO] Creating data via indexing...
[PASS] Database file created
[INFO] Database size before restart: 32768 bytes
[INFO] Restarting container...
[PASS] Container restarted successfully
[INFO] Database size after restart: 32768 bytes
[PASS] Database persisted across restart

[INFO] Test 7: Graceful shutdown
[INFO] Stopping container...
[PASS] Container stopped gracefully
[PASS] No errors during shutdown

[INFO] Test 8: Development mode configuration
[INFO] Starting in development mode...
[PASS] Dev container started
[PASS] Dev container health check passed
[PASS] Debug logging enabled in dev mode

==========================================
  Test Results Summary
==========================================
Tests Run:    8
Tests Passed: 18
Tests Failed: 0
==========================================

[PASS] All tests passed! ✅
```

## Test Configuration

### Endpoints Tested
- **Health**: `http://localhost:8080/health`
- **MCP**: `http://localhost:8080/mcp`

### Timeouts
- **Service Startup**: 60 seconds max
- **Graceful Shutdown**: 10 seconds
- **Health Check Interval**: 30 seconds

### Volume Mounts
- **Production**: `./data:/data` (persistent database)
- **Development**: `.:/data/codebase:ro` (live codebase mount)

## Troubleshooting

### Test Failures

#### "Service did not become ready within 60s"
**Cause**: Container failed to start or health check failing

**Solutions**:
```bash
# Check container logs
docker compose logs conexus

# Check container status
docker compose ps

# Verify port not in use
lsof -i :8080

# Try manual startup
docker compose up
```

#### "Docker image build failed"
**Cause**: Build dependencies missing or Dockerfile errors

**Solutions**:
```bash
# Check Docker daemon
docker info

# Try clean build
docker compose build --no-cache

# Check Dockerfile syntax
docker compose config
```

#### "Database file not found"
**Cause**: Volume mount issues or permission problems

**Solutions**:
```bash
# Check volume mounts
docker compose config | grep volumes

# Check permissions
ls -la ./data/

# Create data directory
mkdir -p ./data
chmod 755 ./data
```

#### "MCP endpoint failed"
**Cause**: Server not responding or port conflict

**Solutions**:
```bash
# Check server is listening
curl http://localhost:8080/health

# Check logs for errors
docker compose logs conexus | grep -i error

# Verify MCP handler initialized
docker compose logs conexus | grep -i mcp
```

### Common Issues

#### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
CONEXUS_PORT=8081 docker compose up
```

#### Permission Denied on ./data
```bash
# Fix permissions
sudo chown -R $(id -u):$(id -g) ./data
chmod -R 755 ./data
```

#### Container Fails to Stop
```bash
# Force stop
docker compose down --timeout 1

# Force remove
docker compose rm -fsv

# Check for orphaned containers
docker ps -a | grep conexus
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Docker Integration Tests

on: [push, pull_request]

jobs:
  docker-tests:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq
      
      - name: Run Docker integration tests
        run: |
          ./tests/integration/docker/docker_test.sh --cleanup
      
      - name: Upload logs on failure
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: docker-logs
          path: |
            ./data/
            docker-compose-logs.txt
```

### GitLab CI Example
```yaml
docker-integration:
  stage: test
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - apk add --no-cache curl jq bash
  script:
    - ./tests/integration/docker/docker_test.sh --cleanup
  artifacts:
    when: on_failure
    paths:
      - ./data/
    expire_in: 1 week
```

## Test Maintenance

### Adding New Tests

1. **Add test function to `docker_test.sh`**:
```bash
# ==========================================
# Test 9: Your New Test
# ==========================================
test_start "Your new test description"

# Your test logic here
if [ condition ]; then
  log_success "Test passed"
else
  log_error "Test failed"
fi
echo ""
```

2. **Update this README**:
- Add test to "Test Categories" section
- Document expected behavior
- Add troubleshooting tips if needed

### Modifying Tests

1. **Preserve test counter logic**:
```bash
((TESTS_RUN++))    # Increment on test_start()
((TESTS_PASSED++)) # Increment on log_success()
((TESTS_FAILED++)) # Increment on log_error()
```

2. **Maintain idempotency**:
- Tests should clean up after themselves
- Use `--cleanup` flag for automatic cleanup
- Don't assume previous test state

3. **Keep tests independent**:
- Each test should work in isolation
- Don't rely on side effects from other tests
- Use unique identifiers for test data

## Performance Benchmarks

### Expected Timings (Reference System: 4-core, 8GB RAM)
- **Image Build**: ~30-60s (first build), ~5-10s (cached)
- **Container Start**: ~3-5s
- **Health Check**: ~1-3s
- **MCP Tool Call**: ~100-500ms
- **Full Test Suite**: ~90-120s

### Resource Usage
- **CPU**: ~5-10% during idle, ~30-50% during indexing
- **Memory**: ~50-100MB base, ~200-300MB during indexing
- **Disk**: ~20MB image, ~1-5MB database (varies by codebase size)

## Related Documentation

- [Docker Deployment Guide](../../../README.md#docker-deployment)
- [MCP Protocol Specification](../../../docs/api-reference.md)
- [Configuration Reference](../../../docs/operations/operations-guide.md)
- [Development Guide](../../../docs/contributing/contributing-guide.md)

## Support

For issues with integration tests:

1. Check [Troubleshooting](#troubleshooting) section
2. Review container logs: `docker compose logs conexus`
3. Check GitHub Issues: https://github.com/ferg-cod3s/conexus/issues
4. Run with verbose output: `bash -x ./tests/integration/docker/docker_test.sh`

## License

MIT License - see [LICENSE](../../../LICENSE) for details
