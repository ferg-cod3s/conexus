# Connector Setup Guide

Complete guide to setting up and configuring Conexus connectors.

---

## 📖 Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Supported Connectors](#supported-connectors)
4. [Configuration](#configuration)
5. [Authentication](#authentication)
6. [Testing](#testing)
7. [Advanced Usage](#advanced-usage)
8. [Troubleshooting](#troubleshooting)

---

## Overview

Conexus supports **dynamic connector registration** - you only get tools for the connectors you configure. This means:

✅ **Clean interface** - No unused tools
✅ **Multiple instances** - Run multiple GitHub repos, Slack workspaces, etc.
✅ **Project-specific** - Different configs for different projects
✅ **OAuth ready** - Modern authentication for all platforms

### Supported Platforms

| Platform | Status | Auth Methods | Tools Available |
|----------|--------|--------------|-----------------|
| **GitHub** | ✅ Ready | PAT, OAuth, GitHub App | 6 tools |
| **Slack** | ✅ Ready | Bot Token, OAuth | 3 tools |
| **Jira** | ✅ Ready | API Token, OAuth 2.0 | 3 tools |
| **Discord** | ✅ Ready | Bot Token, OAuth | 3 tools |

---

## Quick Start

### 1. Choose Your Connectors

**Question:** What platforms does your project use?

- ✅ GitHub for code/issues
- ✅ Slack for team communication
- ✅ Jira for project management
- ✅ Discord for community

### 2. Create Configuration

Copy the example:
```bash
cp config/connectors.example.json config/connectors.json
chmod 600 config/connectors.json
```

Enable only what you need:
```json
{
  "connectors": [
    {
      "id": "github-main",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "your-org/your-repo"
      },
      "enabled": true  ← Set to true
    },
    {
      "id": "slack-team",
      "type": "slack",
      "config": {
        "token": "${SLACK_TOKEN}",
        "channels": ["C01234567"]
      },
      "enabled": false  ← Disabled = no slack.* tools
    }
  ]
}
```

### 3. Get Credentials

See [Authentication](#authentication) section for platform-specific instructions.

### 4. Start Server

```bash
export GITHUB_TOKEN="ghp_xxxxx"
export SLACK_TOKEN="xoxb_xxxxx"

conexus server --config config/connectors.json
```

### 5. Verify

Check available tools:
```json
{
  "jsonrpc": "2.0",
  "method": "tools/list",
  "id": 1
}
```

You should only see tools for enabled connectors!

---

## Supported Connectors

### GitHub

**What it does:**
- Search issues and pull requests
- Get detailed issue/PR information with comments
- List repositories
- Sync repository data for indexing

**Available Tools:**
- `github.search_issues` - Search with GitHub syntax
- `github.get_issue` - Get specific issue + comments
- `github.get_pr` - Get pull request + comments
- `github.list_repos` - List accessible repositories
- `github.sync_status` - Check sync status
- `github.sync_trigger` - Trigger data sync

**Configuration:**
```json
{
  "id": "github-main",
  "type": "github",
  "config": {
    "token": "${GITHUB_TOKEN}",
    "repository": "owner/repo",
    "sync_interval": "5m"
  }
}
```

**Learn More:** [GitHub Setup Guide](./CONNECTOR_TESTING_AUTH.md#github)

---

### Slack

**What it does:**
- Search messages across channels
- List workspace channels
- Retrieve complete message threads

**Available Tools:**
- `slack.search` - Search messages
- `slack.list_channels` - List all channels
- `slack.get_thread` - Get thread with replies

**Configuration:**
```json
{
  "id": "slack-engineering",
  "type": "slack",
  "config": {
    "token": "${SLACK_TOKEN}",
    "channels": ["C01234567", "C98765432"],
    "sync_interval": "2m",
    "max_messages": 1000
  }
}
```

**Learn More:** [Slack Setup Guide](./CONNECTOR_TESTING_AUTH.md#slack)

---

### Jira

**What it does:**
- Search issues with JQL
- Get detailed issue information with comments
- List projects

**Available Tools:**
- `jira.search` - Search with JQL
- `jira.get_issue` - Get specific issue + comments
- `jira.list_projects` - List accessible projects

**Configuration:**
```json
{
  "id": "jira-main",
  "type": "jira",
  "config": {
    "base_url": "https://your-domain.atlassian.net",
    "email": "${JIRA_EMAIL}",
    "api_token": "${JIRA_TOKEN}",
    "projects": ["PROJ", "TEST"],
    "sync_interval": "10m"
  }
}
```

**Learn More:** [Jira Setup Guide](./CONNECTOR_TESTING_AUTH.md#jira)

---

### Discord

**What it does:**
- Search messages in channels
- List server channels
- Retrieve thread messages

**Available Tools:**
- `discord.search` - Search in specific channel
- `discord.list_channels` - List all channels in server
- `discord.get_thread` - Get messages from thread

**Configuration:**
```json
{
  "id": "discord-community",
  "type": "discord",
  "config": {
    "token": "${DISCORD_TOKEN}",
    "guild_id": "123456789012345678",
    "channels": ["999999999", "888888888"],
    "sync_interval": "2m"
  }
}
```

**Learn More:** [Discord Setup Guide](./CONNECTOR_TESTING_AUTH.md#discord)

---

## Configuration

### Configuration File Structure

```json
{
  "connectors": [
    {
      "id": "unique-identifier",        // Used in tool calls
      "name": "Human Readable Name",    // For display
      "type": "github|slack|jira|discord",
      "config": {
        // Platform-specific configuration
      },
      "enabled": true                   // Enable/disable without removing
    }
  ]
}
```

### Environment Variables

Use `${VAR_NAME}` syntax for secrets:

```json
{
  "config": {
    "token": "${GITHUB_TOKEN}",
    "api_key": "${JIRA_API_KEY}"
  }
}
```

Then set environment variables:
```bash
export GITHUB_TOKEN="ghp_xxxxx"
export JIRA_API_KEY="xxxxx"
```

### Multiple Instances

Configure multiple connectors of the same type:

```json
{
  "connectors": [
    {
      "id": "github-backend",
      "type": "github",
      "config": {"repository": "company/backend", "token": "${GITHUB_TOKEN}"}
    },
    {
      "id": "github-frontend",
      "type": "github",
      "config": {"repository": "company/frontend", "token": "${GITHUB_TOKEN}"}
    }
  ]
}
```

Use different `connector_id` in tool calls:
```json
// Backend
{"tool": "github.search_issues", "arguments": {"connector_id": "github-backend"}}

// Frontend
{"tool": "github.search_issues", "arguments": {"connector_id": "github-frontend"}}
```

**Learn More:** [Connector Configuration Guide](./CONNECTOR_CONFIGURATION.md)

---

## Authentication

### Quick Setup (Tokens)

**Fastest way to get started:**

| Platform | Get Token | Time | Scopes Needed |
|----------|-----------|------|---------------|
| **GitHub** | [tokens](https://github.com/settings/tokens) | 1 min | `repo`, `read:org` |
| **Slack** | [apps](https://api.slack.com/apps) | 5 min | `channels:history`, `search:read` |
| **Jira** | [tokens](https://id.atlassian.com/manage-profile/security/api-tokens) | 2 min | N/A (uses email + token) |
| **Discord** | [apps](https://discord.com/developers/applications) | 5 min | Bot with MESSAGE CONTENT INTENT |

### OAuth Setup (Production)

For production deployments, use OAuth:

**Benefits:**
- ✅ User-initiated authorization
- ✅ Automatic token refresh
- ✅ Better security
- ✅ Granular permissions

**Setup:**
1. Create OAuth app on each platform
2. Configure redirect URIs
3. Store client ID/secret securely
4. Implement OAuth flow

**Learn More:** [OAuth Implementation Guide](./OAUTH_IMPLEMENTATION.md)

---

## Testing

### Test Individual Connectors

```bash
# Set credentials
export CONEXUS_TEST_GITHUB_TOKEN="ghp_xxxxx"
export CONEXUS_TEST_GITHUB_REPO="owner/repo"

# Run integration tests
go test -v -tags=integration ./internal/connectors/ -run TestGitHubConnector_Integration
```

### Test All Connectors

```bash
# Set all credentials
export CONEXUS_TEST_GITHUB_TOKEN="ghp_xxxxx"
export CONEXUS_TEST_SLACK_TOKEN="xoxb_xxxxx"
export CONEXUS_TEST_JIRA_TOKEN="xxxxx"
export CONEXUS_TEST_DISCORD_TOKEN="xxxxx"

# Run all integration tests
go test -v -tags=integration ./internal/connectors/...
```

### Test MCP Tools

```bash
# Start server
conexus server --config config/connectors.json

# In another terminal, use MCP client or Claude Desktop
# to test tool calls
```

**Learn More:** [Testing Quick Start](./TESTING_QUICK_START.md)

---

## Advanced Usage

### Project-Specific Configurations

**Use Case:** Different projects need different connectors.

**Solution:** Multiple config files.

```
config/
├── project-a.json  (GitHub + Slack)
├── project-b.json  (GitHub + Jira + Discord)
└── personal.json   (GitHub only)
```

Start with specific config:
```bash
conexus server --config config/project-a.json
```

### Dynamic Connector Management

Add/remove connectors at runtime:

```json
// List connectors
{
  "tool": "context.connector_management",
  "arguments": {"action": "list"}
}

// Add connector
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "add",
    "connector_type": "github",
    "connector_id": "github-new",
    "config": {...}
  }
}

// Remove connector
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "remove",
    "connector_id": "github-old"
  }
}

// Test connector
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "test",
    "connector_id": "github-main"
  }
}
```

### Rate Limiting

All connectors track rate limits:

```json
// Check rate limit
{
  "tool": "github.sync_status",
  "arguments": {"connector_id": "github-main"}
}

// Response includes rate limit info
{
  "rate_limit": {
    "remaining": 4500,
    "reset": "2025-11-12T13:00:00Z",
    "limit": 5000
  }
}
```

### Sync Intervals

Configure how often connectors sync data:

```json
{
  "config": {
    "sync_interval": "5m"  // 5 minutes
  }
}
```

**Recommendations:**
- GitHub: `5m` (rate limit: 5000/hour)
- Slack: `2m` (more real-time needs)
- Jira: `10m` (slower changing data)
- Discord: `2m` (real-time chat)

---

## Troubleshooting

### Problem: Tools Not Appearing

**Symptom:** Expected tools not in `tools/list`

**Solutions:**
1. Check connector is enabled: `"enabled": true`
2. Verify configuration is valid
3. Restart server to reload config
4. Check logs for connector errors

```bash
# Check connector status
{
  "tool": "context.connector_management",
  "arguments": {"action": "list"}
}
```

### Problem: Authentication Errors

**Symptom:** "401 Unauthorized" or "403 Forbidden"

**Solutions:**
1. Verify token is valid and not expired
2. Check token has required scopes/permissions
3. Test credentials directly with platform API
4. Regenerate token if needed

```bash
# Test connector
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "test",
    "connector_id": "github-main"
  }
}
```

### Problem: Rate Limiting

**Symptom:** "429 Too Many Requests"

**Solutions:**
1. Check rate limit status via sync_status
2. Increase sync_interval
3. Use authenticated requests (higher limits)
4. Wait for rate limit reset

### Problem: Connector Not Found

**Symptom:** "connector not found" in tool calls

**Solutions:**
1. List connectors to get correct IDs
2. Verify connector_id matches config
3. Check connector is enabled

```json
// Get correct IDs
{
  "tool": "context.connector_management",
  "arguments": {"action": "list"}
}
```

### Problem: Missing Permissions

**Symptom:** Some features don't work (e.g., can't search)

**Solutions:**
1. **GitHub**: Add `repo`, `read:org`, `read:discussion` scopes
2. **Slack**: Add `search:read`, `channels:history` scopes
3. **Jira**: Verify account has access to projects
4. **Discord**: Enable MESSAGE CONTENT INTENT

---

## Documentation Index

### Getting Started
- 📘 [Testing Quick Start](./TESTING_QUICK_START.md) - Get testing in 5 minutes
- 📙 [OAuth Implementation](./OAUTH_IMPLEMENTATION.md) - Production OAuth setup
- 📗 [Connector Configuration](./CONNECTOR_CONFIGURATION.md) - Advanced configuration

### Reference
- 📕 [Connector Testing & Auth](./CONNECTOR_TESTING_AUTH.md) - Complete auth guide
- 📔 [API Reference](./api-reference.md) - All tools documented
- 📓 [MCP Integration Guide](./getting-started/mcp-integration-guide.md) - MCP client setup

### Development
- 🔧 [AGENTS.md](../AGENTS.md) - Development guidelines
- 🔧 [CLAUDE.md](../CLAUDE.md) - AI assistant guidelines
- 🔧 [Contributing Guide](./contributing/contributing-guide.md) - How to contribute

---

## Quick Reference Card

### Configuration Template
```json
{
  "connectors": [
    {
      "id": "github-main",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "owner/repo"
      },
      "enabled": true
    }
  ]
}
```

### Common Commands
```bash
# Start server
conexus server --config config/connectors.json

# Test connector
go test -v -tags=integration ./internal/connectors/ -run TestGitHub

# List tools
# (via MCP client: tools/list)
```

### Environment Variables
```bash
# Tokens
export GITHUB_TOKEN="ghp_xxxxx"
export SLACK_TOKEN="xoxb_xxxxx"
export JIRA_TOKEN="xxxxx"
export DISCORD_TOKEN="xxxxx"

# Config
export CONEXUS_CONFIG="config/connectors.json"
export CONEXUS_LOG_LEVEL="debug"
```

---

## Support

### Common Issues
- Authentication → See [CONNECTOR_TESTING_AUTH.md](./CONNECTOR_TESTING_AUTH.md)
- OAuth Setup → See [OAUTH_IMPLEMENTATION.md](./OAUTH_IMPLEMENTATION.md)
- Configuration → See [CONNECTOR_CONFIGURATION.md](./CONNECTOR_CONFIGURATION.md)
- Testing → See [TESTING_QUICK_START.md](./TESTING_QUICK_START.md)

### Getting Help
- 📖 Check documentation above
- 🐛 [Report issues](https://github.com/ferg-cod3s/conexus/issues)
- 💬 [Discussions](https://github.com/ferg-cod3s/conexus/discussions)
- 📧 Contact maintainers

---

**Last Updated:** 2025-11-12
**Version:** 0.2.1-alpha

Ready to connect your tools! 🚀
