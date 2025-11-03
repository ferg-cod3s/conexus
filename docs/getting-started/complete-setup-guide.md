# Complete Conexus Setup Guide

**Conexus - MCP Server for Context-Aware AI Assistants**

This comprehensive guide covers everything you need to set up Conexus in your development environment, from basic installation to advanced configuration. Whether you're a solo developer or part of a team, this guide will get you up and running in under 10 minutes.

## Table of Contents

- [Quick Start](#quick-start)
- [Installation Options](#installation-options)
- [MCP Client Configuration](#mcp-client-configuration)
- [Project Integration](#project-integration)
- [Advanced Configuration](#advanced-configuration)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

---

## Quick Start

Get Conexus running in your project in 5 minutes:

### Prerequisites

- **Node.js 18+** or **Bun** (for npm/bunx installation)
- Git
- Your preferred MCP-compatible client (Claude Desktop, Claude Code, Cursor, OpenCode, VS Code, etc.)

### 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Build binaries for your platform
./scripts/build-binaries.sh

# Verify the build
ls -la bin/
```

### 2. Configure MCP Client

**For Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "/path/to/project/.conexus/db.sqlite",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  }
}
```

**For Cursor** (`.cursor/mcp.json` in your project):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite"
      }
    }
  }
}
```

**For OpenCode** (`.opencode/opencode.jsonc` in your project):

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["bunx", "-y", "@agentic-conexus/mcp"],
      "environment": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_LOG_LEVEL": "info"
      },
      "enabled": true
    }
  }
}
```

**For Claude Code** (`~/.claude/mcp.json`):

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  }
}
```

**For VS Code** (with MCP extension, `.vscode/settings.json`):

```json
{
  "mcp.server.conexus": {
    "command": "npx",
    "args": ["-y", "@agentic-conexus/mcp"],
    "env": {
      "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite",
      "CONEXUS_ROOT_PATH": "${workspaceFolder}"
    }
  }
}
```

### 3. Test the Integration

Restart your MCP client and try:

```
"Search for authentication functions in this codebase"
"Find all database models"
"Show me the API endpoints"
```

---

## Installation Options

Conexus supports multiple installation methods depending on your needs:

### Option 1: Local Binary (Recommended)

**Best for:** Individual developers, offline usage, maximum performance

```bash
# Clone and build
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus
./scripts/build-binaries.sh

# Use the binary directly
./bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)
```

**Pre-built binaries available for:**
- macOS (Intel & Apple Silicon)
- Linux (amd64 & arm64)
- Windows (amd64)

### Option 2: Docker

**Best for:** Teams, consistent environments, easy deployment

```bash
# Quick start with Docker
docker run -d -p 8080:8080 --name conexus conexus:latest

# Or use Docker Compose
docker compose up -d

# Test the service
curl http://localhost:8080/health
```

### Option 3: From Source

**Best for:** Development, customization, latest features

```bash
# Requires Go 1.24+
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Build from source
go build -o conexus ./cmd/conexus

# Run tests
go test ./...

# Start the server
./conexus
```

### Option 4: Automated Setup Script

**Best for:** Quick project setup, automatic configuration

```bash
# Run the universal setup script
curl -fsSL https://raw.githubusercontent.com/ferg-cod3s/conexus/main/setup-conexus.sh | bash

# Or download and run locally
wget https://raw.githubusercontent.com/ferg-cod3s/conexus/main/setup-conexus.sh
chmod +x setup-conexus.sh
./setup-conexus.sh
```

The script automatically:
- Detects your project type (Node.js, Python, Go, etc.)
- Creates appropriate MCP configuration
- Sets up environment variables
- Configures project-specific agents

---

## MCP Client Configuration

Conexus integrates with any MCP-compatible client. Here's how to configure popular clients:

### Claude Desktop

**Configuration file:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "/path/to/project/.conexus/db.sqlite",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  }
}
```

**Restart Claude Desktop** after configuration changes.

### Cursor

**Configuration file:** `.cursor/mcp.json` (project-specific)

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite"
      }
    }
  }
}
```

Cursor supports hot-reload - no restart required.

### OpenCode

**Configuration file:** `.opencode/opencode.jsonc` (project-specific)

```jsonc
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "conexus": {
      "type": "local",
      "command": ["bunx", "-y", "@agentic-conexus/mcp"],
      "environment": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_LOG_LEVEL": "info"
      },
      "enabled": true
    }
  },
  "agent": {
    "typescript-pro": {
      "tools": {
        "conexus": true
      }
    }
  }
}
```

### Claude Code

**Configuration file:** `~/.claude/mcp.json`

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  }
}
```

**Usage:** Start Claude Code with `claude` command. Use `/mcp conexus tools/list` to test.

### VS Code with MCP Extension

**Configuration file:** `.vscode/settings.json`

```json
{
  "mcp.server.conexus": {
    "command": "npx",
    "args": ["-y", "@agentic-conexus/mcp"],
    "env": {
      "CONEXUS_DB_PATH": "${workspaceFolder}/.conexus/db.sqlite",
      "CONEXUS_ROOT_PATH": "${workspaceFolder}"
    }
  }
}
```

### Other MCP Clients

For other MCP-compatible clients, use the stdio transport:

```json
{
  "mcpServers": {
    "conexus": {
      "command": "bunx",
      "args": ["-y", "@agentic-conexus/mcp"],
      "env": {
        "CONEXUS_DB_PATH": "./.conexus/db.sqlite"
      }
    }
  }
}
```

---

## Project Integration

Integrate Conexus with your existing development workflow:

### Node.js/JavaScript Projects

```bash
# Install as dev dependency (optional)
npm install --save-dev @agentic-conexus/mcp

# Add to package.json scripts
{
  "scripts": {
    "conexus": "conexus"
  }
}
```

**Recommended agents:**
- `typescript-pro`
- `javascript-pro`
- `frontend-developer`
- `api-builder-enhanced`

### Python Projects

```bash
# Add to requirements-dev.txt
conexus-mcp>=0.1.0

# Or pyproject.toml
[tool.poetry.dev-dependencies]
conexus-mcp = "^0.1.0"
```

**Recommended agents:**
- `python-pro`
- `api-builder-enhanced`
- `data-scientist`

### Go Projects

```bash
# Add to go.mod (if using as library)
require github.com/ferg-cod3s/conexus v0.1.0
```

**Recommended agents:**
- `go-expert`
- `golang-developer`
- `api-builder-enhanced`

### Team Integration

For team environments:

```bash
# Create shared configuration
mkdir .conexus
cat > .conexus/config.yml << EOF
project:
  name: "my-team-project"
  description: "Team codebase"

codebase:
  root: "."
  include_patterns:
    - "**/*.go"
    - "**/*.js"
    - "**/*.py"
  exclude_patterns:
    - "**/node_modules/**"
    - "**/vendor/**"
    - "**/.git/**"

indexing:
  auto_reindex: true
  reindex_interval: "1h"
EOF

# Share with team
git add .conexus/
git commit -m "Add Conexus team configuration"
```

### CI/CD Integration

Add Conexus to your CI pipeline:

```yaml
# .github/workflows/conexus.yml
name: Conexus Index
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  index:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Conexus
        run: |
          curl -fsSL https://raw.githubusercontent.com/ferg-cod3s/conexus/main/setup-conexus.sh | bash
      - name: Index codebase
        run: |
          ./conexus index --root . --output .conexus/index.json
      - name: Upload index
        uses: actions/upload-artifact@v3
        with:
          name: conexus-index
          path: .conexus/index.json
```

---

## Advanced Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONEXUS_DB_PATH` | SQLite database location | `~/.conexus/db.sqlite` |
| `CONEXUS_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |
| `CONEXUS_PORT` | HTTP server port (0 = stdio only) | `0` |
| `CONEXUS_ROOT_PATH` | Project root directory | Current directory |
| `CONEXUS_EMBEDDING_PROVIDER` | Embedding provider | `mock` |
| `CONEXUS_EMBEDDING_MODEL` | Embedding model | `mock-768` |
| `ANTHROPIC_API_KEY` | Anthropic API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |

### Embedding Providers

**Mock (Default):**
```bash
export CONEXUS_EMBEDDING_PROVIDER=mock
export CONEXUS_EMBEDDING_MODEL=mock-768
```

**Anthropic:**
```bash
export CONEXUS_EMBEDDING_PROVIDER=anthropic
export CONEXUS_EMBEDDING_MODEL=claude-3-haiku-20240307
export ANTHROPIC_API_KEY=your-api-key
```

**OpenAI:**
```bash
export CONEXUS_EMBEDDING_PROVIDER=openai
export CONEXUS_EMBEDDING_MODEL=text-embedding-3-small
export OPENAI_API_KEY=your-api-key
```

### Security Configuration

**Rate Limiting:**
```bash
export CONEXUS_RATE_LIMIT_ENABLED=true
export CONEXUS_RATE_LIMIT_DEFAULT_REQUESTS=100
export CONEXUS_RATE_LIMIT_DEFAULT_WINDOW=1m
```

**TLS/HTTPS:**
```bash
export CONEXUS_TLS_ENABLED=true
export CONEXUS_TLS_CERT_FILE=/path/to/cert.pem
export CONEXUS_TLS_KEY_FILE=/path/to/key.pem
```

### Performance Tuning

**For large codebases:**
```bash
export CONEXUS_INDEXING_CHUNK_SIZE=500
export CONEXUS_SEARCH_MAX_RESULTS=100
export CONEXUS_VECTORSTORE_CACHE_SIZE=1000
```

**Memory optimization:**
```bash
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=512MB
export CONEXUS_INDEXING_MEMORY_LIMIT=256MB
```

### Multi-Project Setup

For monorepos or multiple projects:

```bash
# Project-specific databases
export CONEXUS_DB_PATH=./project-a/.conexus/db.sqlite  # For project A
export CONEXUS_DB_PATH=./project-b/.conexus/db.sqlite  # For project B

# Shared configuration
export CONEXUS_CONFIG=./shared/conexus.yml
```

---

## Troubleshooting

### Common Issues

#### "Command not found" or "Binary not executable"

**Solution:**
```bash
# Check if binary exists
ls -la bin/conexus-*

# Make executable
chmod +x bin/conexus-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)

# Verify platform
uname -s  # Should show Darwin, Linux, etc.
uname -m  # Should show x86_64, arm64, etc.
```

#### "Connection refused" or "MCP server not responding"

**Solution:**
```bash
# Check if Conexus is running
ps aux | grep conexus

# Test stdio mode directly
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus

# Check logs (if available)
tail -f /tmp/conexus.log
```

#### "No search results" or "Empty responses"

**Solution:**
```bash
# Check database exists and has content
ls -la .conexus/db.sqlite

# Force reindex
./conexus index --force

# Check configuration
cat .conexus/config.yml
```

#### "Permission denied" or "Access issues"

**Solution:**
```bash
# Check file permissions
ls -la bin/conexus-*

# Fix permissions
chmod 755 bin/conexus-*

# Check directory permissions
ls -ld .conexus/
chmod 755 .conexus/
```

#### "Out of memory" errors

**Solution:**
```bash
# Reduce memory usage
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=256MB
export CONEXUS_INDEXING_MEMORY_LIMIT=128MB

# Use smaller batch sizes
export CONEXUS_INDEXING_CHUNK_SIZE=250
```

### MCP Client-Specific Issues

#### Claude Desktop
- **Issue:** "MCP server failed to start"
  - **Solution:** Check absolute paths in config, restart Claude Desktop
- **Issue:** "Tool not available"
  - **Solution:** Verify JSON syntax, check logs in `~/Library/Logs/Claude/`

#### Cursor
- **Issue:** "Configuration not loaded"
  - **Solution:** Ensure `.cursor/mcp.json` is in project root
- **Issue:** "Hot reload not working"
  - **Solution:** Restart Cursor, check file permissions

#### OpenCode
- **Issue:** "Agent not found"
  - **Solution:** Verify agent names in `.opencode/opencode.jsonc`
- **Issue:** "Tool access denied"
  - **Solution:** Check agent tool permissions in config

### Performance Issues

#### Slow search responses
```bash
# Check system resources
top -p $(pgrep conexus)

# Optimize configuration
export CONEXUS_SEARCH_MAX_RESULTS=50
export CONEXUS_VECTORSTORE_CACHE_SIZE=500
```

#### High memory usage
```bash
# Monitor memory
ps aux --sort=-%mem | head

# Configure limits
export CONEXUS_VECTORSTORE_MEMORY_LIMIT=512MB
```

#### Indexing taking too long
```bash
# Check progress
./conexus index --status

# Optimize for large codebases
export CONEXUS_INDEXING_CHUNK_SIZE=100
export CONEXUS_INDEXING_WORKERS=2
```

### Getting Help

1. **Check the logs:**
   ```bash
   # Enable debug logging
   export CONEXUS_LOG_LEVEL=debug
   ./conexus 2>&1 | tee conexus.log
   ```

2. **Test with minimal configuration:**
   ```bash
   # Use defaults
   unset CONEXUS_*
   ./conexus
   ```

3. **Community support:**
   - [GitHub Issues](https://github.com/ferg-cod3s/conexus/issues)
   - [GitHub Discussions](https://github.com/ferg-cod3s/conexus/discussions)

---

## Next Steps

### Start Using Conexus

Now that Conexus is set up, try these queries in your MCP client:

**Code Understanding:**
- "Find all authentication functions"
- "Show me the database schema"
- "What are the main API endpoints?"

**Bug Investigation:**
- "Search for error handling patterns"
- "Find all logging statements"
- "Show me the test files"

**Feature Development:**
- "Locate the user management code"
- "Find similar implementations to this function"
- "Show me the configuration loading logic"

### Advanced Usage

1. **Explore MCP Tools:**
   - `context.search` - Semantic search
   - `context.get_related_info` - File relationships
   - `context.index_control` - Manage indexing
   - `context.connector_management` - Data sources

2. **Customize Your Setup:**
   - Add project-specific configurations
   - Integrate with your CI/CD pipeline
   - Set up team-shared databases

3. **Contribute Back:**
   - Report bugs and issues
   - Suggest improvements
   - Submit pull requests

## Appendices

### Configuration Reference

- **[Configuration Reference](configuration-reference.md)** - Complete configuration options, environment variables, and file formats
- **[Advanced Configuration](advanced-configuration.md)** - Enterprise deployment, security, and performance tuning
- **[Troubleshooting](troubleshooting.md)** - Common issues and diagnostic procedures

### API and Performance

- **[API Reference](../api-reference.md)** - Complete API documentation and schemas
- **[Performance Baseline](../PERFORMANCE_BASELINE.md)** - Benchmarks and performance metrics

### Integration Guides

- **[MCP Integration Guide](mcp-integration-guide.md)** - Detailed MCP setup and tools
- **[Developer Onboarding](developer-onboarding.md)** - Development workflow and best practices
- **[Project Integration](project-integration/)** - Framework-specific setup guides
  - [Node.js/JavaScript](project-integration/nodejs.md)
  - [Python](project-integration/python.md)
  - [Go](project-integration/go.md)
  - [Rust](project-integration/rust.md)

---

**Ready to supercharge your AI-assisted development?** 🚀

Conexus provides intelligent codebase context that transforms how AI assistants understand and work with your code. Start exploring your codebase with natural language queries and experience the difference!

For questions or issues, check the [troubleshooting section](#troubleshooting) or open an issue on GitHub.</content>
<parameter name="filePath">docs/getting-started/complete-setup-guide.md