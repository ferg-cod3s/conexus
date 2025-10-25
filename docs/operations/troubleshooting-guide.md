# Conexus Troubleshooting Guide

## Overview

This guide helps diagnose and resolve common issues with Conexus deployments. Issues are organized by category with symptoms, causes, and solutions.

## Quick Health Check

Run this script to check system health:

```bash
#!/bin/bash
echo "=== Conexus Health Check ==="
echo "Date: $(date)"
echo

# Check if container is running
echo "1. Container Status:"
docker-compose ps conexus
echo

# Check health endpoint
echo "2. Health Check:"
curl -s http://localhost:8080/health || echo "❌ Health check failed"
echo

# Check logs for errors
echo "3. Recent Errors:"
docker-compose logs --tail=10 conexus | grep -i error || echo "✅ No recent errors"
echo

# Check resource usage
echo "4. Resource Usage:"
docker stats --no-stream conexus 2>/dev/null || echo "⚠️  Unable to get stats"
echo

# Check database size
echo "5. Database Size:"
ls -lh data/conexus.db 2>/dev/null || echo "⚠️  Database not found"
echo

echo "=== End Health Check ==="
```

## Performance Issues

### High Latency (>100ms p95)

**Symptoms**:
- Slow response times
- MCP requests taking >100ms
- User complaints about sluggishness

**Causes & Solutions**:

1. **Large Database**
   ```
   Symptoms: High I/O wait, slow queries
   Check: ls -lh data/conexus.db
   Solution: VACUUM database
   docker-compose exec conexus sqlite3 /data/conexus.db "VACUUM;"
   ```

2. **Insufficient CPU**
   ```
   Symptoms: High CPU usage, slow embedding generation
   Check: docker stats conexus
   Solution: Increase CPU allocation or scale horizontally
   ```

3. **Memory Pressure**
   ```
   Symptoms: Frequent GC, OOM errors
   Check: docker logs conexus | grep -i "gc\|oom"
   Solution: Increase RAM or reduce chunk_size in config
   ```

### High Error Rates

**Symptoms**:
- >1% error rate in metrics
- Failed MCP requests
- 5xx HTTP status codes

**Common Causes**:

1. **Database Corruption**
   ```
   Symptoms: "database disk image is malformed"
   Logs: sqlite3 errors
   Solution:
   docker-compose down
   cp data/conexus.db data/conexus.db.backup
   docker-compose up -d  # Creates new DB
   # Re-index codebase if needed
   ```

2. **Resource Exhaustion**
   ```
   Symptoms: Connection refused, timeouts
   Check: docker stats conexus
   Solution: Scale resources or implement rate limiting
   ```

## Startup Issues

### Container Won't Start

**Symptoms**:
- `docker-compose up` fails
- Container exits immediately
- "CrashLoopBackOff" in Kubernetes

**Diagnosis**:
```bash
# Check container logs
docker-compose logs conexus

# Check exit code
docker-compose ps conexus

# Validate configuration
docker-compose config
```

**Common Issues**:

1. **Port Already in Use**
   ```
   Error: bind: address already in use
   Solution: Change port in docker-compose.yml or free port
   netstat -tulpn | grep :8080
   ```

2. **Permission Issues**
   ```
   Error: permission denied
   Solution: Fix volume permissions
   sudo chown -R 1000:1000 data/
   ```

3. **Configuration Errors**
   ```
   Error: invalid config
   Check: Validate YAML syntax
   docker-compose config > /dev/null
   ```

### Database Initialization Fails

**Symptoms**:
- Container starts but health check fails
- "unable to open database" errors

**Solutions**:
```bash
# Check disk space
df -h

# Fix permissions
sudo chown -R 1000:1000 data/

# Check filesystem
docker run --rm -v $(pwd)/data:/data alpine fsck -f /data/conexus.db
```

## MCP Tool Issues

### Tools Return "not_implemented"

**Symptoms**:
- MCP tools return placeholder responses
- Search and get_related_info don't work

**Causes**:
- HTTP handler bug (fixed in v0.1.0-alpha)
- Missing tool implementations

**Verification**:
```bash
# Test working tools
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context_index_control","arguments":{"action":"status"}}}'

# Should return: {"status":"ok","message":"Index contains X documents",...}
```

### Search Returns Empty Results

**Symptoms**:
- Search queries return no results
- "Index contains 0 documents"

**Causes**:
- Codebase not indexed
- Index corrupted

**Solutions**:
```bash
# Check index status
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context_index_control","arguments":{"action":"status"}}}'

# If 0 documents, re-index
# Mount codebase and trigger indexing
echo "Re-indexing required - mount codebase to /data/codebase"
```

## Observability Issues

### Metrics Not Available

**Symptoms**:
- /metrics endpoint returns 404
- Prometheus can't scrape metrics

**Causes**:
- Metrics disabled in configuration
- Wrong port

**Solutions**:
```yaml
# config.yml
observability:
  metrics:
    enabled: true
    port: 9091
```

### Traces Not Appearing

**Symptoms**:
- Jaeger UI shows no traces
- Tracing enabled but no data

**Causes**:
- Jaeger not running
- Wrong endpoint configuration

**Check**:
```bash
# Verify Jaeger
curl http://localhost:16686/api/services

# Check configuration
docker-compose logs jaeger
```

## Networking Issues

### Cannot Connect to Conexus

**Symptoms**:
- Connection refused
- Timeout errors

**Diagnosis**:
```bash
# Check if port is open
netstat -tulpn | grep 8080

# Test connectivity
telnet localhost 8080

# Check firewall
sudo ufw status
```

### Load Balancer Issues

**Symptoms**:
- 502/503 errors behind load balancer
- Uneven load distribution

**Check**:
```bash
# Health check configuration
curl -H "Host: your-domain.com" http://localhost:8080/health

# Load balancer logs
# Check nginx/haproxy logs for upstream errors
```

## Database Issues

### SQLite Performance Degradation

**Symptoms**:
- Queries getting slower over time
- High I/O wait

**Solutions**:
```sql
-- Connect to database
docker-compose exec conexus sqlite3 /data/conexus.db

-- Optimize database
VACUUM;
REINDEX;
ANALYZE;

-- Check fragmentation
PRAGMA integrity_check;
```

### Database Corruption

**Symptoms**:
- "database disk image is malformed"
- Random crashes

**Recovery**:
```bash
# Stop container
docker-compose down

# Backup corrupted DB
cp data/conexus.db data/conexus.db.corrupt

# Remove corrupted DB
rm data/conexus.db

# Restart (creates new DB)
docker-compose up -d

# Restore from backup if available
# Re-index codebase
```

## Resource Issues

### Memory Leaks

**Symptoms**:
- Gradually increasing memory usage
- OOM kills

**Monitoring**:
```bash
# Monitor memory usage
docker stats conexus

# Check for leaks in logs
docker-compose logs conexus | grep -i "alloc\|gc"
```

### Disk Space Issues

**Symptoms**:
- "no space left on device"
- Database operations fail

**Check**:
```bash
# Disk usage
df -h

# Database size
ls -lh data/

# Clean up old logs
docker-compose logs conexus | head -1000 > conexus.log
docker-compose logs -f conexus > /dev/null &
```

## Emergency Procedures

### Complete System Reset

**Use only as last resort**:
```bash
# Stop everything
docker-compose down -v

# Remove all data
sudo rm -rf data/

# Clean rebuild
docker-compose build --no-cache
docker-compose up -d
```

### Emergency Logging

**When normal logs aren't available**:
```bash
# Enable debug logging
export CONEXUS_LOG_LEVEL=debug
docker-compose up -d

# Tail logs
docker-compose logs -f conexus
```

## Support Information

### Gathering Diagnostics

**Run this script before contacting support**:
```bash
#!/bin/bash
echo "=== Conexus Diagnostics ==="
echo "Timestamp: $(date)"
echo "Version: $(curl -s http://localhost:8080/health | jq -r .version)"
echo

echo "=== System Info ==="
uname -a
docker --version
docker-compose --version
echo

echo "=== Container Status ==="
docker-compose ps
echo

echo "=== Resource Usage ==="
docker stats --no-stream
echo

echo "=== Recent Logs ==="
docker-compose logs --tail=50
echo

echo "=== Configuration ==="
cat config.yml 2>/dev/null || echo "No config.yml found"
echo

echo "=== Database Info ==="
ls -la data/conexus.db 2>/dev/null || echo "Database not found"
echo

echo "=== End Diagnostics ==="
```

### Contact Information

- **GitHub Issues**: https://github.com/your-org/conexus/issues
- **Emergency**: security@your-org.com
- **Documentation**: https://docs.conexus.dev

---

**Last Updated**: October 16, 2025
**Version**: 0.1.0-alpha
