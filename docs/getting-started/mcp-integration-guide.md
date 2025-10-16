# MCP Integration Guide

**Conexus Model Context Protocol (MCP) Server**  
**Version:** 1.0  
**Last Updated:** October 15, 2025

## 🎯 Overview

This guide will help you integrate Conexus with AI assistants (like Claude Code, Cursor, and other MCP-compatible tools) in under 5 minutes. By the end, you'll be able to search your codebase using natural language queries through your AI assistant.

### What is MCP?

The [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) is an open standard that enables AI assistants to interact with external tools and data sources. Conexus implements an MCP server that exposes its context engine capabilities through a standardized JSON-RPC 2.0 interface.

### Why Use Conexus with MCP?

- **🔍 Intelligent Search**: Ask natural language questions about your codebase
- **🎯 Context-Aware**: Search results adapt to your current working context
- **🚀 Hybrid Search**: Combines vector similarity and BM25 for better results
- **📊 Multi-Source**: Search across files, documentation, Slack, GitHub, and more
- **⚡ Fast**: Optimized indexing and search with sub-second response times

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Claude Code Integration](#claude-code-integration)
- [Available Tools](#available-tools)
- [Common Integration Patterns](#common-integration-patterns)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)
- [Next Steps](#next-steps)

---

## Prerequisites

### Required Software

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.23.4+ | Conexus runtime |
| **Git** | 2.30+ | Clone repository |
| **Claude Desktop** | Latest | MCP client (recommended) |

### Installation Commands

**macOS (using Homebrew):**
```bash
brew install go git
# Download Claude Desktop from https://claude.ai/download
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y golang-1.23 git
# Download Claude Desktop from https://claude.ai/download
```

**Windows (using WSL2):**
```bash
# Install WSL2: https://learn.microsoft.com/en-us/windows/wsl/install
wsl --install
# Then follow Ubuntu instructions above
```

**Verify Installations:**
```bash
go version          # Should show go1.23.4 or higher
git --version       # Should show 2.30+
```

---

## 🚀 Quick Start

### Step 1: Install Conexus

```bash
# Clone the repository
git clone https://github.com/ferg-cod3s/conexus.git
cd conexus

# Install dependencies
go mod download

# Build Conexus
go build ./cmd/conexus

# Verify installation
./conexus --version
```

**Expected Output:**
```
Conexus v0.0.5 (Agentic Context Engine)
Go Version: 1.23.4
```

### Step 2: Start MCP Server

```bash
# Start the MCP server (runs on stdio by default)
./conexus mcp serve

# Or with custom configuration
./conexus mcp serve --config config.yml
```

**Note**: The MCP server uses stdio transport by default, which is the standard for MCP integrations. It will wait for JSON-RPC 2.0 messages on stdin and respond on stdout.

### Step 3: Configure Your MCP Client

#### For Claude Desktop (Recommended)

1. **Open Claude Desktop Configuration**
   ```bash
   # macOS
   open ~/Library/Application\ Support/Claude/claude_desktop_config.json
   
   # Linux
   nano ~/.config/Claude/claude_desktop_config.json
   
   # Windows (WSL)
   notepad.exe "$(wslpath -w ~/.config/Claude/claude_desktop_config.json)"
   ```

2. **Add Conexus MCP Server**
   ```json
   {
     "mcpServers": {
       "conexus": {
         "command": "/path/to/conexus/conexus",
         "args": ["mcp", "serve"],
         "env": {
           "CONEXUS_CONFIG": "/path/to/conexus/config.yml"
         }
       }
     }
   }
   ```

3. **Update the Path**
   - Replace `/path/to/conexus/conexus` with the absolute path to your Conexus binary
   - Replace `/path/to/conexus/config.yml` with your config file location (optional)

4. **Restart Claude Desktop**

### Step 4: Test the Integration

1. **Open Claude Desktop**
2. **Start a new conversation**
3. **Try a search query:**

```
@conexus search for authentication middleware
```

**Expected Response:**
```
I found 15 results related to authentication middleware:

1. internal/auth/middleware.go (score: 0.92)
   Contains AuthMiddleware function that validates JWT tokens...

2. docs/architecture/security.md (score: 0.85)
   Documentation on authentication patterns...

3. tests/integration/auth_test.go (score: 0.78)
   Integration tests for authentication...

[Additional results...]
```

**🎉 Success!** You've successfully integrated Conexus with Claude Code!

---

## 🤖 Claude Code Integration

### Configuration Deep Dive

#### Basic Configuration

```json
{
  "mcpServers": {
    "conexus": {
      "command": "/usr/local/bin/conexus",
      "args": ["mcp", "serve"]
    }
  }
}
```

#### Advanced Configuration with Environment Variables

```json
{
  "mcpServers": {
    "conexus": {
      "command": "/usr/local/bin/conexus",
      "args": ["mcp", "serve", "--log-level", "info"],
      "env": {
        "CONEXUS_CONFIG": "/home/user/.conexus/config.yml",
        "CONEXUS_DATA_DIR": "/home/user/.conexus/data",
        "CONEXUS_LOG_LEVEL": "info"
      }
    }
  }
}
```

#### Multiple Conexus Instances (Monorepo Support)

```json
{
  "mcpServers": {
    "conexus-frontend": {
      "command": "/usr/local/bin/conexus",
      "args": ["mcp", "serve"],
      "env": {
        "CONEXUS_CONFIG": "/projects/myapp/frontend/conexus.yml"
      }
    },
    "conexus-backend": {
      "command": "/usr/local/bin/conexus",
      "args": ["mcp", "serve"],
      "env": {
        "CONEXUS_CONFIG": "/projects/myapp/backend/conexus.yml"
      }
    }
  }
}
```

### Conexus Configuration File

Create a `config.yml` file to customize Conexus behavior:

```yaml
# config.yml
version: "1.0"

# Data storage location
data_dir: "/home/user/.conexus/data"

# Embedding provider
embedding:
  provider: "openai"  # or "local", "cohere", "anthropic"
  model: "text-embedding-ada-002"
  api_key: "${OPENAI_API_KEY}"  # Use environment variable

# Vector store configuration
vectorstore:
  type: "sqlite"  # or "memory", "postgres"
  path: "${data_dir}/vectors.db"

# Search settings
search:
  default_top_k: 20
  max_top_k: 100
  min_score: 0.3
  hybrid_weight: 0.7  # 70% vector, 30% BM25

# Indexer settings
indexer:
  auto_index: true
  watch_files: true
  ignore_patterns:
    - "node_modules"
    - ".git"
    - "vendor"
    - "*.test.go"

# Logging
log_level: "info"  # debug, info, warn, error
log_format: "json"  # json or text
```

### Tool Discovery

Once configured, Claude Desktop automatically discovers available tools:

```
Available Conexus Tools:
✅ context.search - Search codebase with natural language
✅ context.get_related_info - Get related code/docs for a specific item
⏳ context.index_control - Manage indexing (status only)
⏳ context.connector_management - Manage data sources (list only)
```

**Legend:**
- ✅ **Fully Implemented** - Production ready
- ⏳ **Partial Implementation** - Limited functionality
- ❌ **Not Implemented** - Placeholder

---

## 🛠️ Available Tools

### 1. `context.search`

**Purpose:** Search your codebase, documentation, and other indexed content using natural language queries.

**Use When:**
- You need to find relevant files or code snippets
- You want to understand how a feature is implemented
- You're looking for examples or patterns
- You need to locate documentation

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | ✅ Yes | - | Natural language search query |
| `work_context` | object | ❌ No | - | Your current working context |
| `work_context.active_file` | string | ❌ No | - | Currently open file path |
| `work_context.git_branch` | string | ❌ No | - | Current git branch |
| `work_context.open_ticket_ids` | array | ❌ No | - | Related ticket/issue IDs |
| `top_k` | integer | ❌ No | 20 | Max results to return (1-100) |
| `filters` | object | ❌ No | - | Search filters |
| `filters.source_types` | array | ❌ No | - | Filter by source type |
| `filters.date_range` | object | ❌ No | - | Filter by date range |

**Example Usage in Claude:**

```
Simple Query:
"Search for authentication middleware implementations"

Context-Aware Query:
"I'm working on internal/auth/handler.go - find related authentication code"

Filtered Query:
"Search for database migrations in the last 30 days, files only"

Complex Query:
"Find examples of how we handle JWT token validation across the codebase"
```

**Response Format:**

```json
{
  "results": [
    {
      "id": "doc_12345",
      "content": "package auth\n\nfunc AuthMiddleware() gin.HandlerFunc {...}",
      "score": 0.92,
      "source_type": "file",
      "metadata": {
        "file_path": "internal/auth/middleware.go",
        "language": "go",
        "last_modified": "2024-01-15T10:30:00Z",
        "line_start": 45,
        "line_end": 78
      }
    }
  ],
  "total_count": 15,
  "query_time_ms": 45.2
}
```

**Tips:**
- Be specific in your queries for better results
- Provide work context when available (Claude does this automatically)
- Use `top_k` to control result quantity vs quality
- Filter by `source_types` to focus on specific content types

---

### 2. `context.get_related_info`

**Purpose:** Get contextually related information for a specific code item (file, function, class, etc.).

**Use When:**
- You want to understand what depends on a piece of code
- You need to find related documentation
- You're exploring code relationships
- You want to see usage examples

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `item_id` | string | ✅ Yes | - | ID of the item to get related info for |
| `relation_types` | array | ❌ No | all | Types of relations to include |
| `max_depth` | integer | ❌ No | 2 | Max depth for transitive relations |
| `include_metadata` | boolean | ❌ No | true | Include full metadata in results |

**Supported Relation Types:**
- `depends_on` - Direct dependencies
- `depended_by` - Things that depend on this item
- `calls` - Functions/methods this calls
- `called_by` - What calls this function/method
- `imports` - Imports this module
- `imported_by` - What imports this module
- `related_docs` - Related documentation
- `related_discussions` - Related Slack/GitHub discussions

**Example Usage in Claude:**

```
Basic:
"Get related info for internal/auth/middleware.go"

Specific Relations:
"Show me what depends on the AuthMiddleware function"

Deep Exploration:
"Find all code and docs related to our JWT implementation, depth 3"
```

**Response Format:**

```json
{
  "item": {
    "id": "file_12345",
    "type": "file",
    "path": "internal/auth/middleware.go"
  },
  "relations": [
    {
      "type": "called_by",
      "target": {
        "id": "func_67890",
        "type": "function",
        "name": "SetupRoutes",
        "file": "internal/server/routes.go"
      },
      "strength": 0.95
    },
    {
      "type": "related_docs",
      "target": {
        "id": "doc_54321",
        "type": "markdown",
        "title": "Authentication Architecture",
        "file": "docs/architecture/auth.md"
      },
      "strength": 0.88
    }
  ],
  "total_relations": 12
}
```

---

### 3. `context.index_control` ⏳

**Status:** Partially Implemented (status queries only)

**Purpose:** Manage the indexing process for your codebase.

**Available Commands:**
- `status` - Get current indexing status ✅
- `start` - Start/trigger indexing ⏳ (planned)
- `stop` - Stop active indexing ⏳ (planned)
- `rebuild` - Rebuild entire index ⏳ (planned)

**Example Usage in Claude:**

```
"Check the indexing status"
"What's the current state of the index?"
```

---

### 4. `context.connector_management` ⏳

**Status:** Partially Implemented (list connectors only)

**Purpose:** Manage data source connectors (GitHub, Slack, Jira, etc.).

**Available Commands:**
- `list` - List configured connectors ✅
- `add` - Add new connector ⏳ (planned)
- `remove` - Remove connector ⏳ (planned)
- `sync` - Trigger connector sync ⏳ (planned)

**Example Usage in Claude:**

```
"List available data connectors"
"What connectors are configured?"
```

---

## 💡 Common Integration Patterns

### Pattern 1: Code Understanding Workflow

**Scenario:** You're new to a codebase and need to understand how authentication works.

```
Step 1: Broad Search
You: "Search for authentication implementation"

Step 2: Focused Investigation
You: "Get related info for internal/auth/middleware.go"

Step 3: Deep Dive
You: "Search for JWT token validation, show me tests and docs"

Step 4: Understand Dependencies
You: "What depends on the AuthMiddleware function?"
```

### Pattern 2: Bug Investigation Workflow

**Scenario:** You're debugging a production issue with user sessions.

```
Step 1: Find Relevant Code
You: "Search for session management and timeout handling"

Step 2: Context-Aware Search
You (with internal/session/manager.go open):
"I'm looking at session/manager.go - find related timeout code"

Step 3: Check Recent Changes
You: "Search for session-related changes in the last 7 days"

Step 4: Find Tests
You: "Show me tests for session timeout behavior"
```

### Pattern 3: Feature Development Workflow

**Scenario:** You're implementing OAuth2 integration.

```
Step 1: Find Examples
You: "Search for existing OAuth implementations in our codebase"

Step 2: Check Patterns
You: "How do we handle external authentication providers?"

Step 3: Find Dependencies
You: "What authentication libraries are we using?"

Step 4: Check Documentation
You: "Find docs about our authentication architecture"
```

### Pattern 4: Code Review Workflow

**Scenario:** You're reviewing a pull request that changes the API layer.

```
Step 1: Understand Context
You: "Get related info for internal/api/handlers.go"

Step 2: Check Dependencies
You: "What depends on the handlers in this file?"

Step 3: Find Tests
You: "Search for API handler tests"

Step 4: Check Documentation
You: "Find API documentation for these endpoints"
```

---

## 🔧 Troubleshooting

### Connection Issues

**Problem:** Claude Code can't connect to Conexus

**Solutions:**

1. **Verify Conexus is in PATH:**
   ```bash
   which conexus
   # Should output: /usr/local/bin/conexus (or your install location)
   ```

2. **Check Configuration Path:**
   ```bash
   # Verify the path in claude_desktop_config.json is absolute
   # ❌ Wrong: "command": "conexus"
   # ✅ Correct: "command": "/usr/local/bin/conexus"
   ```

3. **Test Conexus Manually:**
   ```bash
   ./conexus mcp serve
   # Should start without errors
   ```

4. **Check Claude Logs:**
   ```bash
   # macOS
   tail -f ~/Library/Logs/Claude/mcp*.log
   
   # Linux
   tail -f ~/.local/share/Claude/logs/mcp*.log
   ```

---

### Tool Not Found

**Problem:** Claude says "Tool 'context.search' not found"

**Solutions:**

1. **Verify Tool Registration:**
   ```bash
   # Test MCP server locally
   echo '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' | ./conexus mcp serve
   ```

2. **Check Server Logs:**
   ```bash
   # Enable debug logging
   export CONEXUS_LOG_LEVEL=debug
   ./conexus mcp serve
   ```

3. **Restart Claude Desktop:**
   - Completely quit Claude Desktop (Cmd+Q on macOS)
   - Clear cache (optional): `rm -rf ~/Library/Caches/Claude`
   - Restart Claude Desktop

---

### Search Returns No Results

**Problem:** Searches return empty results

**Solutions:**

1. **Check Index Status:**
   ```
   Ask Claude: "Check the indexing status"
   ```

2. **Verify Data Directory:**
   ```bash
   # Check if index exists
   ls -lh ~/.conexus/data/
   # Should show conexus.db and other index files
   ```

3. **Trigger Indexing:**
   ```bash
   # Manual indexing (if auto-index is disabled)
   ./conexus index --path /path/to/your/codebase
   ```

4. **Check Configuration:**
   ```yaml
   # Verify ignore patterns aren't too broad
   indexer:
     ignore_patterns:
       - "node_modules"  # Good
       - "*"             # ❌ Bad - ignores everything!
   ```

5. **Test Search Directly:**
   ```bash
   # Use conexus CLI to test search
   ./conexus search "authentication middleware"
   ```

---

### Performance Issues

**Problem:** Searches are slow or Claude times out

**Solutions:**

1. **Check Index Size:**
   ```bash
   # Large indexes can be slow
   du -sh ~/.conexus/data/
   ```

2. **Optimize Configuration:**
   ```yaml
   search:
     default_top_k: 10  # Reduce from 20
     max_top_k: 50      # Reduce from 100
     min_score: 0.5     # Increase from 0.3 (fewer results)
   ```

3. **Tune Hybrid Search:**
   ```yaml
   search:
     hybrid_weight: 0.8  # More vector (faster), less BM25
   ```

4. **Reduce Indexing Scope:**
   ```yaml
   indexer:
     ignore_patterns:
       - "vendor"
       - "node_modules"
       - "*.min.js"
       - "*.map"
       - "dist"
       - "build"
   ```

5. **Check System Resources:**
   ```bash
   # Monitor during search
   top -p $(pgrep conexus)
   ```

---

### Configuration Not Loading

**Problem:** Environment variables or config file not being used

**Solutions:**

1. **Verify Environment Variable Format:**
   ```json
   {
     "mcpServers": {
       "conexus": {
         "env": {
           "CONEXUS_CONFIG": "/absolute/path/config.yml",
           "OPENAI_API_KEY": "${OPENAI_API_KEY}"
         }
       }
     }
   }
   ```

2. **Check File Permissions:**
   ```bash
   ls -l config.yml
   # Should be readable: -rw-r--r--
   chmod 644 config.yml
   ```

3. **Validate YAML Syntax:**
   ```bash
   # Use a YAML validator
   python -c "import yaml; yaml.safe_load(open('config.yml'))"
   ```

4. **Test Configuration:**
   ```bash
   # Run with explicit config
   CONEXUS_CONFIG=/path/to/config.yml ./conexus mcp serve --log-level debug
   ```

---

## ⚙️ Advanced Configuration

### Custom Embedding Providers

#### OpenAI (Default)
```yaml
embedding:
  provider: "openai"
  model: "text-embedding-ada-002"
  api_key: "${OPENAI_API_KEY}"
  dimensions: 1536
```

#### Anthropic
```yaml
embedding:
  provider: "anthropic"
  model: "claude-3-sonnet-20240229"
  api_key: "${ANTHROPIC_API_KEY}"
```

#### Local Models (Ollama)
```yaml
embedding:
  provider: "local"
  model: "nomic-embed-text"
  endpoint: "http://localhost:11434"
  dimensions: 768
```

#### Cohere
```yaml
embedding:
  provider: "cohere"
  model: "embed-english-v3.0"
  api_key: "${COHERE_API_KEY}"
  dimensions: 1024
```

### Vector Store Backends

#### SQLite (Default, Recommended for Single User)
```yaml
vectorstore:
  type: "sqlite"
  path: "${data_dir}/vectors.db"
  options:
    cache_size: 10000
    mmap_size: 268435456  # 256MB
```

#### PostgreSQL (Recommended for Teams)
```yaml
vectorstore:
  type: "postgres"
  connection_string: "postgresql://user:pass@localhost:5432/conexus?sslmode=disable"
  options:
    max_connections: 10
    max_idle: 5
```

#### In-Memory (Development Only)
```yaml
vectorstore:
  type: "memory"
  # Fast but data is lost on restart
```

### Search Optimization

#### Latency-Optimized
```yaml
search:
  default_top_k: 10       # Fewer results
  max_top_k: 50
  min_score: 0.6          # Higher threshold
  hybrid_weight: 0.9      # Mostly vector (faster)
  timeout_ms: 1000        # 1 second max
```

#### Quality-Optimized
```yaml
search:
  default_top_k: 50       # More results
  max_top_k: 100
  min_score: 0.2          # Lower threshold
  hybrid_weight: 0.5      # Balanced vector/BM25
  timeout_ms: 5000        # 5 seconds max
```

### Indexing Strategies

#### Aggressive Indexing (Small Codebases)
```yaml
indexer:
  auto_index: true
  watch_files: true       # Real-time updates
  scan_interval: "30s"    # Scan every 30s
  batch_size: 1000
  ignore_patterns:
    - ".git"
```

#### Conservative Indexing (Large Codebases)
```yaml
indexer:
  auto_index: false       # Manual control
  watch_files: false      # Disable file watching
  scan_interval: "1h"     # Scan hourly
  batch_size: 100
  ignore_patterns:
    - ".git"
    - "node_modules"
    - "vendor"
    - "dist"
    - "build"
    - "*.test.*"
    - "test_*"
    - "*_test.*"
```

### Security Configuration

#### API Key Rotation
```yaml
embedding:
  api_key: "${OPENAI_API_KEY}"  # Use env vars
  api_key_rotation: true
  api_key_rotation_interval: "24h"
```

#### Access Control
```yaml
security:
  enable_auth: true
  auth_token: "${CONEXUS_AUTH_TOKEN}"
  allowed_origins:
    - "https://claude.ai"
    - "http://localhost:*"
```

#### Rate Limiting
```yaml
rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst: 20
```

---

## 📚 Next Steps

### Learn More

- **[Full Tool Reference](../internal/mcp/README.md)** - Complete API documentation
- **[Configuration Reference](../config.example.yml)** - All configuration options
- **[Architecture Docs](../docs/architecture/)** - Deep dive into internals
- **[Developer Onboarding](./developer-onboarding.md)** - Contributing to Conexus

### Extend Conexus

1. **Add Custom Connectors:**
   - Integrate with your team's data sources
   - See: [Connector Development Guide](../docs/architecture/integration.md)

2. **Custom Embedding Models:**
   - Use domain-specific embeddings
   - See: [Embedding Provider Guide](../internal/embedding/README.md)

3. **Advanced Workflows:**
   - Build multi-step agent workflows
   - See: [Orchestration Guide](../internal/orchestrator/README.md)

### Join the Community

- **GitHub**: [github.com/ferg-cod3s/conexus](https://github.com/ferg-cod3s/conexus)
- **Issues**: Report bugs and request features
- **Discussions**: Ask questions and share ideas
- **Contributing**: See [CONTRIBUTING.md](../contributing/contributing-guide.md)

---

## 📖 Additional Resources

### MCP Specification
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)
- [MCP SDK Documentation](https://github.com/anthropics/mcp-sdk)

### Conexus Documentation
- [Technical Architecture](../Technical-Architecture.md)
- [API Reference](../api-reference.md)
- [Security & Compliance](../Security-Compliance.md)
- [Operations Guide](../operations/operations-guide.md)

### Example Configurations

#### Minimal Setup (Quick Start)
```yaml
version: "1.0"
data_dir: "~/.conexus/data"
embedding:
  provider: "openai"
  api_key: "${OPENAI_API_KEY}"
```

#### Full-Featured Setup (Production)
```yaml
version: "1.0"
data_dir: "/var/lib/conexus"

embedding:
  provider: "openai"
  model: "text-embedding-ada-002"
  api_key: "${OPENAI_API_KEY}"
  dimensions: 1536

vectorstore:
  type: "postgres"
  connection_string: "${DATABASE_URL}"
  options:
    max_connections: 20
    max_idle: 10

search:
  default_top_k: 20
  max_top_k: 100
  min_score: 0.3
  hybrid_weight: 0.7
  timeout_ms: 3000

indexer:
  auto_index: true
  watch_files: true
  scan_interval: "5m"
  batch_size: 500
  ignore_patterns:
    - ".git"
    - "node_modules"
    - "vendor"
    - "*.min.js"
    - "dist"
    - "build"

log_level: "info"
log_format: "json"

security:
  enable_auth: true
  auth_token: "${CONEXUS_AUTH_TOKEN}"

rate_limiting:
  enabled: true
  requests_per_minute: 100
  burst: 20
```

---

## ✅ Success Checklist

Before you finish, make sure you can:

- [ ] Start Conexus MCP server without errors
- [ ] See Conexus tools in Claude Desktop
- [ ] Perform a successful search query
- [ ] Get related info for a code file
- [ ] Understand search results and scores
- [ ] Access Conexus logs for debugging
- [ ] Configure custom search parameters
- [ ] Use work context for better results

**Congratulations!** 🎉 You've successfully integrated Conexus with your AI assistant. You can now leverage natural language search to explore and understand your codebase more effectively.

---

**Questions or Issues?**
- Open an issue: [github.com/ferg-cod3s/conexus/issues](https://github.com/ferg-cod3s/conexus/issues)
- Check logs: `~/.conexus/logs/` or `~/Library/Logs/Claude/mcp*.log`
- Enable debug mode: `export CONEXUS_LOG_LEVEL=debug`
