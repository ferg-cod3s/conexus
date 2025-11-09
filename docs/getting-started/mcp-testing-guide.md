# MCP Testing Guide

## Overview

This guide covers testing and validating Conexus as an MCP (Model Context Protocol) server. It includes strategies for stdio transport testing, MCP Inspector usage, client-specific testing, and deployment recommendations.

---

## Table of Contents

- [Testing Strategies](#testing-strategies)
- [stdio Transport Testing](#stdio-transport-testing)
- [MCP Inspector Setup](#mcp-inspector-setup)
- [Client-Specific Testing](#client-specific-testing)
- [Deployment Strategies](#deployment-strategies)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Testing Strategies

### Testing Pyramid for MCP Servers

```
    /\
   /  \      E2E MCP Client Tests (Claude Desktop, Cursor)
  /____\
 /      \    MCP Inspector Tests (stdio protocol validation)
/________\
/          \  Unit Tests (handler logic, JSON-RPC 2.0)
/____________\
```

### Test Levels

| Level | What to Test | Tools | Frequency |
|-------|--------------|-------|-----------|
| **Unit** | Handler logic, error handling | Go tests | Every commit |
| **Integration** | stdio protocol, JSON-RPC 2.0 | MCP Inspector | Pre-release |
| **E2E** | Real client interaction | Claude Desktop, Cursor | Release validation |

---

## stdio Transport Testing

### Understanding stdio Transport

MCP uses **JSON-RPC 2.0 over stdin/stdout**. This means:
- Requests are sent via **stdin** (standard input)
- Responses are received via **stdout** (standard output)
- No HTTP, no sockets, no network calls
- Line-delimited JSON (one request/response per line)

### Manual stdio Testing

#### 1. Basic Protocol Test

```bash
# Test tools/list endpoint
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus

# Expected response:
# {"jsonrpc":"2.0","id":1,"result":{"tools":[...]}}
```

#### 2. Test All MCP Tools

```bash
# Test context.search
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"authentication","top_k":5}}}' | ./conexus

# Test context.get_related_info
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.get_related_info","arguments":{"file_path":"internal/mcp/server.go"}}}' | ./conexus

# Test context.index_control
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | ./conexus

# Test context.connector_management
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"context.connector_management","arguments":{"action":"list"}}}' | ./conexus

# Test context.explain (new in v0.2.1)
echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"context.explain","arguments":{"target":"auth middleware","context":"How does JWT validation work?"}}}' | ./conexus

# Test context.grep (new in v0.2.1)
echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"context.grep","arguments":{"pattern":"func.*Auth"}}}' | ./conexus
```

#### 3. Error Handling Test

```bash
# Test invalid JSON
echo 'invalid json' | ./conexus

# Test missing required field
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{}}' | ./conexus

# Test unknown method
echo '{"jsonrpc":"2.0","id":1,"method":"unknown/method","params":{}}' | ./conexus
```

### Automated stdio Test Script

Create `test-stdio.sh`:

```bash
#!/bin/bash
set -e

CONEXUS_BINARY="${1:-./conexus}"

if [ ! -f "$CONEXUS_BINARY" ]; then
  echo "❌ Conexus binary not found: $CONEXUS_BINARY"
  exit 1
fi

echo "🧪 Testing Conexus MCP Server (stdio)"
echo "======================================="
echo ""

# Test 1: tools/list
echo "Test 1: tools/list"
RESPONSE=$(echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | "$CONEXUS_BINARY")
if echo "$RESPONSE" | grep -q '"result"'; then
  echo "✅ PASS: tools/list"
else
  echo "❌ FAIL: tools/list"
  echo "$RESPONSE"
  exit 1
fi
echo ""

# Test 2: context.search
echo "Test 2: context.search"
REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"context.search","arguments":{"query":"test","top_k":5}}}'
RESPONSE=$(echo "$REQUEST" | "$CONEXUS_BINARY")
if echo "$RESPONSE" | grep -q '"result"'; then
  echo "✅ PASS: context.search"
else
  echo "❌ FAIL: context.search"
  echo "$RESPONSE"
  exit 1
fi
echo ""

# Test 3: context.index_control
echo "Test 3: context.index_control"
REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}'
RESPONSE=$(echo "$REQUEST" | "$CONEXUS_BINARY")
if echo "$RESPONSE" | grep -q '"result"'; then
  echo "✅ PASS: context.index_control"
else
  echo "❌ FAIL: context.index_control"
  echo "$RESPONSE"
  exit 1
fi
echo ""

# Test 4: Error handling
echo "Test 4: Error handling (invalid JSON)"
RESPONSE=$(echo 'invalid json' | "$CONEXUS_BINARY" 2>&1 || true)
if echo "$RESPONSE" | grep -q '"error"'; then
  echo "✅ PASS: Error handling"
else
  echo "⚠️  WARNING: Error handling might need improvement"
fi
echo ""

echo "🎉 All stdio tests passed!"
```

Run it:

```bash
chmod +x test-stdio.sh
./test-stdio.sh ./conexus
```

---

## MCP Inspector Setup

### What is MCP Inspector?

[MCP Inspector](https://modelcontextprotocol.io/docs/tools/inspector) is an official MCP debugging tool that:
- Validates JSON-RPC 2.0 protocol compliance
- Tests stdio transport
- Inspects tool schemas
- Validates request/response formats

### Installation

```bash
# Install via npm
npm install -g @modelcontextprotocol/inspector

# Or via npx (no installation)
npx @modelcontextprotocol/inspector
```

### Using MCP Inspector with Conexus

#### 1. Start Inspector in Server Mode

```bash
# Inspect Conexus MCP server
mcp-inspector ./conexus

# With environment variables
CONEXUS_CONFIG=./config.yml mcp-inspector ./conexus

# With arguments
mcp-inspector ./conexus --config ./config.yml
```

#### 2. Inspector Interface

Once started, you'll see:

```
MCP Inspector
Server: ./conexus
Status: ✅ Connected

Available Tools:
  1. context.search
  2. context.get_related_info
  3. context.index_control
  4. context.connector_management
  5. context.explain
  6. context.grep

> Type 'help' for commands
```

#### 3. Inspector Commands

```bash
# List all tools
> tools

# Get tool schema
> describe context.search

# Call a tool
> call context.search {"query": "authentication", "top_k": 5}

# View last request/response
> last

# Check protocol compliance
> validate

# Exit inspector
> exit
```

#### 4. Automated Inspector Testing

Create `test-inspector.js`:

```javascript
#!/usr/bin/env node
import { spawn } from 'child_process';
import { readFileSync } from 'fs';

const CONEXUS_PATH = process.argv[2] || './conexus';

async function testWithInspector() {
  console.log('🔍 Testing with MCP Inspector\n');

  const inspector = spawn('mcp-inspector', [CONEXUS_PATH], {
    stdio: ['pipe', 'pipe', 'inherit']
  });

  // Test tools/list
  inspector.stdin.write('tools\n');

  // Test context.search
  inspector.stdin.write('call context.search {"query": "test", "top_k": 5}\n');

  // Test validation
  inspector.stdin.write('validate\n');

  // Exit
  setTimeout(() => {
    inspector.stdin.write('exit\n');
  }, 5000);

  inspector.stdout.on('data', (data) => {
    console.log(data.toString());
  });
}

testWithInspector();
```

---

## Client-Specific Testing

### Claude Desktop

#### Setup

1. Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "conexus": {
      "command": "/absolute/path/to/conexus",
      "args": [],
      "env": {
        "CONEXUS_CONFIG": "/absolute/path/to/config.yml",
        "CONEXUS_LOG_LEVEL": "debug"
      }
    }
  }
}
```

2. Restart Claude Desktop

#### Testing

1. **Verify Connection**:
   - Look for hammer icon (🔨) in Claude Desktop
   - Click to see available tools
   - Should show all 6 Conexus tools

2. **Test Each Tool**:

```
User: "Use context.search to find authentication code"
→ Claude should invoke context.search tool

User: "Get related info for internal/mcp/server.go"
→ Claude should invoke context.get_related_info

User: "Check index status"
→ Claude should invoke context.index_control

User: "Explain how the auth middleware works"
→ Claude should invoke context.explain

User: "Search for all function definitions matching 'Auth'"
→ Claude should invoke context.grep
```

3. **Check Logs**:

```bash
# macOS
tail -f ~/Library/Logs/Claude/mcp-conexus.log

# Windows
type %APPDATA%\Claude\Logs\mcp-conexus.log

# Linux
tail -f ~/.local/share/Claude/logs/mcp-conexus.log
```

#### Expected Log Output

```
[INFO] Conexus MCP server starting
[INFO] Loaded configuration from /path/to/config.yml
[INFO] Vector store initialized: 1234 documents
[INFO] MCP server ready (stdio mode)
[DEBUG] Received request: tools/list
[DEBUG] Sending response: 6 tools
[DEBUG] Received request: tools/call (context.search)
[DEBUG] Query: "authentication" (top_k: 5)
[DEBUG] Search completed: 5 results in 23ms
```

---

### Cursor

#### Setup

Create `.cursor/mcp.json` in your project:

```json
{
  "mcpServers": {
    "conexus": {
      "command": "/absolute/path/to/conexus",
      "args": [],
      "env": {
        "CONEXUS_CONFIG": "/absolute/path/to/config.yml"
      }
    }
  }
}
```

#### Testing

1. **Verify in Cursor Settings**:
   - Settings → MCP → Servers
   - Should show "conexus" with status "Connected"

2. **Test via AI Chat**:

```
User: "@conexus search for authentication code"
→ Cursor should invoke context.search
```

3. **Check Cursor Logs**:

```bash
# macOS
tail -f ~/Library/Application\ Support/Cursor/logs/mcp.log

# Windows
type %APPDATA%\Cursor\logs\mcp.log

# Linux
tail -f ~/.config/Cursor/logs/mcp.log
```

---

### OpenCode

#### Setup

Create `opencode-config.jsonc`:

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "model": "opencode/code-supernova",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["/absolute/path/to/conexus"],
      "environment": {
        "CONEXUS_CONFIG": "./config.yml"
      },
      "enabled": true
    }
  }
}
```

#### Testing

1. **Verify Connection**:
   - OpenCode will auto-load MCP servers on startup
   - Check status: `/mcp list`

2. **Test Tools**:

```bash
# List tools
/mcp tools conexus

# Call context.search
/mcp call conexus context.search '{"query":"auth","top_k":5}'

# Check index status
/mcp call conexus context.index_control '{"action":"status"}'
```

---

### Claude Code (CLI)

#### Setup

Create `~/.claude/mcp.json`:

```json
{
  "conexus": {
    "command": "/absolute/path/to/conexus",
    "env": {
      "CONEXUS_CONFIG": "/path/to/config.yml"
    }
  }
}
```

#### Testing

```bash
# List MCP servers
claude-code mcp list

# List tools
claude-code mcp tools conexus

# Call context.search
claude-code mcp call conexus context.search '{"query":"authentication","top_k":5}'

# Interactive mode
claude-code
> /mcp conexus tools/call context.search {"query":"auth"}
```

---

## Deployment Strategies

### stdio vs HTTP/TLS: Decision Matrix

| Criteria | stdio (Standard) | HTTP/TLS (Custom) |
|----------|------------------|-------------------|
| **MCP Compliance** | ✅ Native MCP protocol | ⚠️ Custom transport |
| **Client Support** | ✅ All MCP clients | ❌ Requires custom client |
| **Setup Complexity** | ✅ Simple (no certs) | ⚠️ TLS cert management |
| **Local Development** | ✅ Perfect fit | ⚠️ Overkill |
| **Network Access** | ❌ Local only | ✅ Remote access |
| **Security** | ✅ Process isolation | ✅ TLS encryption |
| **Performance** | ✅ No network overhead | ⚠️ Network latency |
| **Use Case** | Local MCP clients | API/enterprise |

### Recommended Strategies

#### Strategy 1: stdio-First (Recommended)

**Best for**: Standard MCP usage, local development, CI/CD

```yaml
# config.yml
server:
  mode: stdio  # Default MCP mode
  
embedding:
  provider: anthropic
  api_key: ${ANTHROPIC_API_KEY}

vectorstore:
  type: sqlite
  path: ./data/conexus.db
```

**Deploy**:

```bash
# Build
go build -o conexus ./cmd/conexus

# Run (stdio mode is default)
./conexus

# With config
CONEXUS_CONFIG=./config.yml ./conexus
```

**Advantages**:
- ✅ Works with all MCP clients out of the box
- ✅ No certificate management
- ✅ Simple deployment
- ✅ Process-level security

**Use for**:
- Claude Desktop integration
- Cursor integration
- OpenCode integration
- Local development
- CI/CD testing

---

#### Strategy 2: HTTP/TLS (Optional)

**Best for**: Remote access, API usage, enterprise deployments

```yaml
# config.tls.yml
server:
  mode: http
  port: 8080
  tls:
    enabled: true
    cert_file: ./certs/server.crt
    key_file: ./certs/server.key
    client_auth: require

authentication:
  enabled: true
  api_keys:
    - key: ${CONEXUS_API_KEY}
      name: "production"
```

**Deploy**:

```bash
# Generate certificates
./scripts/generate-dev-certs.sh

# Run with TLS
CONEXUS_CONFIG=./config.tls.yml ./conexus

# Test HTTPS endpoint
curl https://localhost:8080/mcp \
  -X POST \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
```

**Advantages**:
- ✅ Remote access from anywhere
- ✅ TLS encryption
- ✅ API key authentication
- ✅ Rate limiting possible

**Use for**:
- Remote team access
- API integrations
- Enterprise deployments
- Future SaaS offering

---

#### Strategy 3: Hybrid (Advanced)

**Best for**: Development teams needing both local and remote access

Run both modes simultaneously:

```bash
# Terminal 1: stdio mode for local MCP clients
CONEXUS_CONFIG=./config.yml ./conexus

# Terminal 2: HTTP mode for API access
CONEXUS_CONFIG=./config.tls.yml CONEXUS_PORT=8080 ./conexus
```

Or use Docker Compose:

```yaml
# docker-compose.hybrid.yml
version: '3.8'

services:
  conexus-stdio:
    build: .
    command: ./conexus
    volumes:
      - ./config.yml:/app/config.yml
      - ./data:/app/data

  conexus-api:
    build: .
    command: ./conexus
    ports:
      - "8080:8080"
    volumes:
      - ./config.tls.yml:/app/config.yml
      - ./data:/app/data
      - ./certs:/app/certs
```

---

## Troubleshooting

### Issue: Claude Desktop Not Connecting

**Symptoms**:
- No hammer icon (🔨) in Claude Desktop
- Tools not appearing
- Connection errors in logs

**Debug Steps**:

1. **Verify binary path**:
```bash
# Test binary manually
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | /path/to/conexus
```

2. **Check configuration**:
```bash
# Verify config file exists
cat ~/Library/Application\ Support/Claude/claude_desktop_config.json

# Ensure absolute paths (not relative)
```

3. **Check permissions**:
```bash
# Make binary executable
chmod +x /path/to/conexus

# Verify ownership
ls -la /path/to/conexus
```

4. **Review logs**:
```bash
# macOS
tail -f ~/Library/Logs/Claude/mcp-conexus.log

# Look for startup errors
grep ERROR ~/Library/Logs/Claude/mcp-conexus.log
```

**Common Fixes**:
- Use **absolute paths** (not `~/` or `./`)
- Ensure binary is **executable**
- Verify environment variables are set
- Restart Claude Desktop after config changes

---

### Issue: stdio Communication Failure

**Symptoms**:
- No response from conexus
- Broken pipe errors
- Incomplete JSON responses

**Debug Steps**:

1. **Test JSON-RPC manually**:
```bash
# Valid request
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus

# Check for newline handling
printf '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}\n' | ./conexus
```

2. **Check buffer flushing**:
```go
// Ensure stdout is flushed after each response
fmt.Fprintf(os.Stdout, "%s\n", responseJSON)
os.Stdout.Sync() // Explicit flush
```

3. **Verify line-delimited JSON**:
```bash
# Each request/response must be on a single line
# NO pretty-printing in stdio mode
```

---

### Issue: Tool Not Found

**Symptoms**:
- `Method not found` error
- Tools not listed in `tools/list`
- Client can't invoke tool

**Debug Steps**:

1. **Verify tool registration**:
```bash
# Check tools/list response
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus | jq '.result.tools[] | .name'

# Expected output:
# context.search
# context.get_related_info
# context.index_control
# context.connector_management
# context.explain
# context.grep
```

2. **Check tool naming**:
```go
// Ensure dot notation (not underscore)
// ✅ CORRECT: "context.search"
// ❌ WRONG:   "context_search"
```

3. **Verify schema**:
```bash
# Get tool schema
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus | jq '.result.tools[] | select(.name=="context.search")'
```

---

### Issue: Slow Search Performance

**Symptoms**:
- Search takes >1 second
- Timeout errors
- High memory usage

**Debug Steps**:

1. **Check index size**:
```bash
# Query index status
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"context.index_control","arguments":{"action":"status"}}}' | ./conexus

# Check database size
ls -lh ./data/conexus.db
```

2. **Profile performance**:
```bash
# Enable profiling
CONEXUS_PROFILE=true ./conexus

# View CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# View memory profile
go tool pprof http://localhost:6060/debug/pprof/heap
```

3. **Optimize queries**:
```json
{
  "query": "specific search term",
  "top_k": 5,  // Reduce for faster results
  "filters": {
    "source_types": ["file"]  // Filter source types
  }
}
```

**Fixes**:
- Reduce `top_k` for faster queries
- Add source type filters
- Use more specific queries
- Consider indexing fewer files
- See [Performance Optimization](../operations/performance-tuning.md)

---

## Best Practices

### 1. Always Use Absolute Paths in Client Configs

❌ **Bad**:
```json
{
  "command": "~/conexus/bin/conexus",
  "env": {
    "CONEXUS_CONFIG": "./config.yml"
  }
}
```

✅ **Good**:
```json
{
  "command": "/Users/username/conexus/bin/conexus",
  "env": {
    "CONEXUS_CONFIG": "/Users/username/conexus/config.yml"
  }
}
```

---

### 2. Test stdio Protocol Before Client Integration

```bash
# Always test manually first
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus

# Verify JSON response is valid
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus | jq .

# Test all tools
./test-stdio.sh
```

---

### 3. Enable Debug Logging for Troubleshooting

```json
{
  "command": "/path/to/conexus",
  "env": {
    "CONEXUS_LOG_LEVEL": "debug",
    "CONEXUS_CONFIG": "/path/to/config.yml"
  }
}
```

---

### 4. Validate JSON-RPC Compliance

```bash
# Use MCP Inspector for validation
mcp-inspector ./conexus

# Run validator
> validate
✅ All tools compliant with JSON-RPC 2.0
✅ All schemas valid
✅ Error handling correct
```

---

### 5. Use Separate Configs for Different Modes

```bash
# stdio mode (local MCP clients)
config.yml

# HTTP/TLS mode (API access)
config.tls.yml

# Development mode (verbose logging)
config.dev.yml

# Production mode (optimized)
config.prod.yml
```

---

### 6. Monitor MCP Server Health

```bash
# Check process is running
ps aux | grep conexus

# Monitor logs in real-time
tail -f ~/Library/Logs/Claude/mcp-conexus.log

# Check for errors
grep ERROR ~/Library/Logs/Claude/mcp-conexus.log

# Monitor metrics (if HTTP mode)
curl http://localhost:9090/metrics
```

---

### 7. Version Your MCP Configuration

```bash
# Track configuration in git
git add claude_desktop_config.json
git commit -m "Update Conexus MCP config for v0.2.1"

# Use environment-specific configs
claude_desktop_config.dev.json
claude_desktop_config.prod.json
```

---

## Testing Checklist

### Pre-Release Testing

- [ ] All unit tests pass (`go test ./...`)
- [ ] stdio protocol test passes (`./test-stdio.sh`)
- [ ] MCP Inspector validation passes
- [ ] Claude Desktop integration tested
- [ ] Cursor integration tested (if applicable)
- [ ] OpenCode integration tested (if applicable)
- [ ] All 6 tools tested manually
- [ ] Error handling tested
- [ ] Performance benchmarks meet targets
- [ ] Security scan passes (`gosec ./...`)
- [ ] Documentation updated

### Client Integration Checklist

- [ ] Configuration file created with absolute paths
- [ ] Environment variables set correctly
- [ ] Binary is executable and in correct location
- [ ] Client restarted after config changes
- [ ] Tools appear in client UI
- [ ] Test query executed successfully
- [ ] Logs reviewed for errors
- [ ] Performance acceptable (<1s response time)

---

## References

- **MCP Specification**: https://modelcontextprotocol.io/docs/spec
- **MCP Inspector**: https://modelcontextprotocol.io/docs/tools/inspector
- **JSON-RPC 2.0**: https://www.jsonrpc.org/specification
- **Conexus MCP Integration Guide**: [mcp-integration-guide.md](./mcp-integration-guide.md)
- **Conexus API Specification**: [../API-Specification.md](../API-Specification.md)

---

## Support

- **GitHub Issues**: [Report a bug](https://github.com/ferg-cod3s/conexus/issues)
- **Discussions**: [Ask a question](https://github.com/ferg-cod3s/conexus/discussions)
- **Documentation**: [docs/README.md](../README.md)

---

**Last Updated**: 2025-10-26  
**Version**: v0.2.1-alpha  
**Status**: ✅ Complete
