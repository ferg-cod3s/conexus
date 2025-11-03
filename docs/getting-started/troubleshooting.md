# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with Conexus setup and operation.

## Quick Diagnosis

### 1. Check Conexus Status

```bash
# Test if Conexus is accessible
curl -s http://localhost:8080/health | jq .

# Expected response:
{
  "status": "healthy",
  "version": "0.1.0-alpha",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### 2. Check MCP Tool Availability

```bash
# Test MCP tools list
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | \
  bunx -y @agentic-conexus/mcp
```

### 3. Verify Database

```bash
# Check if database exists
ls -la .conexus/db.sqlite

# Check database size (should be > 0 for indexed projects)
du -h .conexus/db.sqlite
```

### 4. Check Logs

```bash
# Enable debug logging
export CONEXUS_LOG_LEVEL=debug

# Run with verbose output
bunx -y @agentic-conexus/mcp 2>&1 | head -20
```

## Installation Issues

### "bunx: command not found"

**Problem:** Bun is not installed.

**Solution:**
```bash
# Install Bun
curl -fsSL https://bun.sh/install | bash

# Or use npx instead
npx -y @agentic-conexus/mcp
```

### "Cannot find module '@agentic-conexus/mcp'"

**Problem:** Package not available or network issues.

**Solutions:**
```bash
# Clear npm cache
npm cache clean --force

# Try with different registry
npm install --registry https://registry.npmjs.org/

# Use npx with --ignore-existing
npx --ignore-existing @agentic-conexus/mcp
```

### "Permission denied" on binary

**Problem:** Executable permissions missing.

**Solution:**
```bash
# Make binary executable
chmod +x bin/conexus-*

# Or reinstall
./scripts/build-binaries.sh
```

## MCP Configuration Issues

### "MCP server failed to start"

**Common causes:**
1. Incorrect command path
2. Missing environment variables
3. Port conflicts
4. Database permission issues

**Diagnosis:**
```bash
# Test command directly
bunx -y @agentic-conexus/mcp --help

# Check environment variables
echo $CONEXUS_DB_PATH
echo $CONEXUS_LOG_LEVEL

# Test database access
ls -la $CONEXUS_DB_PATH
```

### Claude Desktop Issues

**"Connection refused":**
```json
// Check configuration path
// macOS
cat ~/Library/Application\ Support/Claude/claude_desktop_config.json

// Windows
type %APPDATA%\Claude\claude_desktop_config.json

// Linux
cat ~/.config/Claude/claude_desktop_config.json
```

**Restart required:**
```bash
# macOS
killall "Claude"
open -a "Claude"

# Windows
taskkill /f /im claude.exe
start claude.exe

# Linux
pkill -f claude
claude &
```

### Cursor Issues

**Configuration not loading:**
```bash
# Check project .cursor directory
ls -la .cursor/

# Validate JSON syntax
cat .cursor/mcp.json | jq .
```

**Hot reload not working:**
```bash
# Restart Cursor
# Or reload window: Ctrl+Shift+P → "Developer: Reload Window"
```

### Claude Code Issues

**Configuration not found:**
```bash
# Check global config location
ls -la ~/.claude/

# Validate JSON syntax
cat ~/.claude/mcp.json | jq .
```

**MCP tools not available:**
```bash
# Test MCP connection in Claude Code
/mcp conexus tools/list

# Check if server is running
ps aux | grep conexus
```

### OpenCode Issues

**Agent not found:**
```jsonc
// Check agent configuration
{
  "agent": {
    "typescript-pro": {
      "tools": {
        "conexus": true
      }
    }
  }
}
```

**Tool access denied:**
```bash
# Check agent permissions in config
grep -A 5 '"agent"' .opencode/opencode.jsonc
```

## Search and Indexing Issues

### "No search results"

**Possible causes:**
1. Codebase not indexed
2. Incorrect search query
3. Database corruption
4. Configuration issues

**Diagnosis:**
```bash
# Check index status
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | \
  bunx -y @agentic-conexus/mcp

# Force reindex
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"force_reindex"}}}' | \
  bunx -y @agentic-conexus/mcp

# Test search
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"function","top_k":5}}}' | \
  bunx -y @agentic-conexus/mcp
```

### "Indexing taking too long"

**Performance issues:**
```bash
# Check system resources
top -p $(pgrep conexus) 2>/dev/null || echo "Conexus not running"

# Reduce indexing scope
export CONEXUS_INDEXING_CHUNK_SIZE=250
export CONEXUS_INDEXING_WORKERS=1

# Exclude large directories
echo "**/node_modules/**" >> .conexus/.ignore
echo "**/vendor/**" >> .conexus/.ignore
```

### "Database locked" errors

**Concurrent access issues:**
```bash
# Check for multiple instances
pgrep -f conexus

# Kill duplicate processes
pkill -f "@agentic-conexus/mcp"

# Restart with single instance
bunx -y @agentic-conexus/mcp
```

## Performance Issues

### Slow Search Response

**Optimization steps:**
```bash
# Enable caching
export CONEXUS_SEARCH_CACHE_ENABLED=true
export CONEXUS_SEARCH_CACHE_TTL=30m

# Reduce result count
export CONEXUS_SEARCH_MAX_RESULTS=25

# Check vector store performance
export CONEXUS_VECTORSTORE_CACHE_SIZE=2000
```

### High Memory Usage

**Memory optimization:**
```bash
# Limit memory usage
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=512MB
export CONEXUS_INDEXING_MEMORY_LIMIT=256MB

# Use smaller chunks
export CONEXUS_INDEXING_CHUNK_SIZE=250

# Monitor memory
watch -n 5 'ps aux --sort=-%mem | head -5'
```

### CPU Usage Issues

**CPU optimization:**
```bash
# Reduce workers
export CONEXUS_INDEXING_WORKERS=1

# Limit parallel operations
export CONEXUS_SEARCH_MAX_CONCURRENT=2

# Monitor CPU
top -p $(pgrep conexus) 2>/dev/null || echo "Conexus not running"
```

## Embedding Provider Issues

### OpenAI API Issues

**"Invalid API key":**
```bash
# Check API key
echo $OPENAI_API_KEY | head -c 10

# Test API key
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
     https://api.openai.com/v1/models
```

**Rate limiting:**
```bash
# Reduce batch size
export CONEXUS_EMBEDDING_BATCH_SIZE=10

# Add delays
export CONEXUS_EMBEDDING_RATE_LIMIT=1000
```

### Anthropic API Issues

**"Invalid API key":**
```bash
# Check API key format
echo $ANTHROPIC_API_KEY | grep -E "^sk-ant-"

# Test API
curl -H "x-api-key: $ANTHROPIC_API_KEY" \
     -H "Content-Type: application/json" \
     https://api.anthropic.com/v1/messages \
     -d '{"model":"claude-3-haiku-20240307","max_tokens":10,"messages":[{"role":"user","content":"test"}]}'
```

### Mock Provider Issues

**Fallback not working:**
```bash
# Force mock provider
export CONEXUS_EMBEDDING_PROVIDER=mock
export CONEXUS_EMBEDDING_MODEL=mock-768

# Clear any cached embeddings
rm -rf .conexus/embeddings/
```

## Network and Connectivity Issues

### "Connection timeout"

**Network issues:**
```bash
# Test internet connectivity
curl -I https://api.openai.com

# Check DNS resolution
nslookup api.openai.com

# Test with different timeout
export CONEXUS_HTTP_TIMEOUT=30s
```

### Proxy Issues

**Behind corporate proxy:**
```bash
# Configure proxy
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080

# Test proxy
curl -I --proxy $HTTP_PROXY https://api.openai.com
```

## File System Issues

### "Permission denied" on database

**File system permissions:**
```bash
# Check directory permissions
ls -ld .conexus/

# Fix permissions
chmod 755 .conexus/
chmod 644 .conexus/db.sqlite

# Check disk space
df -h .
```

### "Read-only file system"

**Disk space or permissions:**
```bash
# Check disk usage
du -sh .conexus/

# Check available space
df -h $(pwd)

# Move to different location
export CONEXUS_DB_PATH=/tmp/conexus.db
```

## Platform-Specific Issues

### macOS Issues

**"Cannot execute binary":**
```bash
# Check Gatekeeper
spctl --assess --verbose bin/conexus-darwin-arm64

# Allow execution
xattr -rd com.apple.quarantine bin/
```

**Homebrew conflicts:**
```bash
# Check for conflicts
brew doctor

# Reinstall conflicting packages
brew reinstall sqlite
```

### Linux Issues

**"Shared object not found":**
```bash
# Install missing libraries
sudo apt-get install libsqlite3-dev

# Check library paths
ldd bin/conexus-linux-amd64
```

**SELinux issues:**
```bash
# Check SELinux status
sestatus

# Allow execution if needed
chcon -t bin_t bin/conexus-linux-amd64
```

### Windows Issues

**"Access denied":**
```bash
# Run as administrator
# Or check permissions
icacls bin\conexus-windows-amd64.exe

# Add execution permission
icacls bin\conexus-windows-amd64.exe /grant Everyone:RX
```

**Path issues:**
```bash
# Use absolute paths
set CONEXUS_DB_PATH=C:\path\to\project\.conexus\db.sqlite

# Check PATH
echo %PATH%
```

## Advanced Debugging

### Enable Debug Logging

```bash
# Maximum verbosity
export CONEXUS_LOG_LEVEL=debug
export CONEXUS_LOG_FORMAT=json

# Log to file
export CONEXUS_LOG_FILE=/tmp/conexus.log

# Run with logging
bunx -y @agentic-conexus/mcp 2>&1 | tee /tmp/conexus-debug.log
```

### Database Inspection

```bash
# Connect to SQLite database
sqlite3 .conexus/db.sqlite

# Check tables
.schema

# Count records
SELECT COUNT(*) FROM chunks;
SELECT COUNT(*) FROM documents;

# Check for corruption
PRAGMA integrity_check;

# Vacuum database
VACUUM;
```

### Performance Profiling

```bash
# Enable profiling
export CONEXUS_PROFILING_ENABLED=true
export CONEXUS_PROFILING_PORT=6060

# Access profiling data
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Network Debugging

```bash
# Enable HTTP debugging
export CONEXUS_HTTP_DEBUG=true

# Monitor network traffic
tcpdump -i any port 8080 -w conexus-traffic.pcap

# Check open connections
netstat -tlnp | grep :8080
```

## Recovery Procedures

### Database Recovery

```bash
# Backup current database
cp .conexus/db.sqlite .conexus/db.sqlite.backup

# Attempt repair
sqlite3 .conexus/db.sqlite ".recover" > recovered.sql
sqlite3 recovered.db < recovered.sql

# If repair fails, reindex
rm .conexus/db.sqlite
bunx -y @agentic-conexus/mcp &
sleep 5
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"force_reindex"}}}' | \
  bunx -y @agentic-conexus/mcp
```

### Configuration Reset

```bash
# Backup current config
cp .conexus/config.yml .conexus/config.yml.backup

# Reset to defaults
rm .conexus/config.yml

# Regenerate configuration
# Edit with your settings
```

### Clean Reinstall

```bash
# Stop all instances
pkill -f conexus
pkill -f "@aagentic-conexus"

# Remove all data
rm -rf .conexus/

# Clear caches
npm cache clean --force
rm -rf ~/.cache/conexus/

# Reinstall
npm install -g @agentic-conexus/mcp

# Restart MCP client
```

## Getting Help

### Community Support

1. **Check existing issues:**
   ```bash
   # Search GitHub issues
   open "https://github.com/ferg-cod3s/conexus/issues"
   ```

2. **Gather diagnostic information:**
   ```bash
   # System information
   uname -a
   node --version || bun --version
   npm --version

   # Conexus version
   bunx -y @agentic-conexus/mcp --version

   # Configuration
   cat .conexus/config.yml
   ```

3. **Create a bug report:**
   - Include system information
   - Attach configuration files
   - Include error logs
   - Describe reproduction steps

### Professional Support

For enterprise deployments, contact:
- **Email:** support@conexus.dev
- **Enterprise SLA:** Available for production deployments

---

**Still having issues?** Check the [GitHub Issues](https://github.com/ferg-cod3s/conexus/issues) or create a new issue with your diagnostic information.