# Connector Configuration Guide

This guide explains how to configure connectors dynamically so that only the tools you need are available.

---

## Overview

Conexus now supports **dynamic tool registration** based on which connectors you have configured. This means:

✅ **Only see tools for connectors you use**
✅ **Configure multiple instances** of the same connector type
✅ **Project-specific configurations**
✅ **No unused tools cluttering your interface**

---

## How It Works

### Dynamic Tool Discovery

When the MCP server starts, it:
1. **Checks which connectors are configured** in your connector store
2. **Registers only the relevant tools** for those connectors
3. **Updates available tools dynamically** as you add/remove connectors

**Example:**

```
Configured Connectors:
- github-main (type: github)
- slack-engineering (type: slack)

Available Tools:
✓ context.search
✓ context.get_related_info
✓ context.index_control
✓ context.connector_management
✓ context.explain
✓ context.grep
✓ github.search_issues         ← GitHub tools
✓ github.get_issue
✓ github.get_pr
✓ github.list_repos
✓ github.sync_status
✓ github.sync_trigger
✓ slack.search                 ← Slack tools
✓ slack.list_channels
✓ slack.get_thread

NOT Available (no connectors configured):
✗ jira.* tools
✗ discord.* tools
```

---

## Configuration Methods

### Method 1: Configuration File (Recommended)

Create `config/connectors.json`:

```json
{
  "connectors": [
    {
      "id": "github-conexus",
      "name": "Conexus Repository",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "ferg-cod3s/conexus",
        "sync_interval": "5m"
      }
    },
    {
      "id": "slack-engineering",
      "name": "Engineering Slack",
      "type": "slack",
      "config": {
        "token": "${SLACK_TOKEN}",
        "channels": ["C01234567", "C98765432"],
        "sync_interval": "2m"
      }
    }
  ]
}
```

**Features:**
- Environment variable substitution (`${VAR_NAME}`)
- Multiple connectors of the same type
- Human-readable connector names
- Per-connector configuration

### Method 2: Environment Variables

```bash
# GitHub connector
export CONEXUS_CONNECTOR_GITHUB_MAIN_TOKEN="ghp_xxxxx"
export CONEXUS_CONNECTOR_GITHUB_MAIN_REPO="owner/repo"

# Slack connector
export CONEXUS_CONNECTOR_SLACK_ENG_TOKEN="xoxb-xxxxx"
export CONEXUS_CONNECTOR_SLACK_ENG_CHANNELS="C01234567,C98765432"

# Jira connector (optional)
export CONEXUS_CONNECTOR_JIRA_MAIN_URL="https://company.atlassian.net"
export CONEXUS_CONNECTOR_JIRA_MAIN_EMAIL="user@company.com"
export CONEXUS_CONNECTOR_JIRA_MAIN_TOKEN="xxxxx"
```

### Method 3: Runtime API

Add connectors via the MCP tool:

```json
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "add",
    "connector_type": "github",
    "connector_id": "github-new-project",
    "config": {
      "token": "ghp_xxxxx",
      "repository": "company/new-project"
    }
  }
}
```

---

## Multiple Connector Instances

You can configure **multiple instances** of the same connector type:

```json
{
  "connectors": [
    {
      "id": "github-conexus",
      "name": "Conexus (Main Project)",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "ferg-cod3s/conexus"
      }
    },
    {
      "id": "github-client-app",
      "name": "Client Application",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "company/client-app"
      }
    },
    {
      "id": "github-docs",
      "name": "Documentation Site",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "company/docs"
      }
    }
  ]
}
```

**Usage:**

```json
// Search issues in the main project
{
  "tool": "github.search_issues",
  "arguments": {
    "connector_id": "github-conexus",
    "query": "is:open label:bug"
  }
}

// Search issues in client app
{
  "tool": "github.search_issues",
  "arguments": {
    "connector_id": "github-client-app",
    "query": "is:open label:bug"
  }
}
```

---

## Project-Specific Configurations

### Use Case: Different Projects, Different Tools

**Project A:** Uses GitHub + Slack
**Project B:** Uses GitHub + Jira + Discord

**Solution:** Use different configuration files per project.

#### Project A Configuration

`config/project-a.json`:
```json
{
  "connectors": [
    {
      "id": "github-project-a",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "company/project-a"
      }
    },
    {
      "id": "slack-project-a",
      "type": "slack",
      "config": {
        "token": "${SLACK_TOKEN_A}",
        "channels": ["C111111"]
      }
    }
  ]
}
```

**Available Tools:** GitHub + Slack tools only

#### Project B Configuration

`config/project-b.json`:
```json
{
  "connectors": [
    {
      "id": "github-project-b",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "company/project-b"
      }
    },
    {
      "id": "jira-project-b",
      "type": "jira",
      "config": {
        "base_url": "https://company.atlassian.net",
        "email": "${JIRA_EMAIL}",
        "api_token": "${JIRA_TOKEN}",
        "projects": ["PROJB"]
      }
    },
    {
      "id": "discord-project-b",
      "type": "discord",
      "config": {
        "token": "${DISCORD_TOKEN}",
        "guild_id": "123456789",
        "channels": ["999999"]
      }
    }
  ]
}
```

**Available Tools:** GitHub + Jira + Discord tools

#### Start Server with Project Config

```bash
# Project A
conexus server --config config/project-a.json

# Project B
conexus server --config config/project-b.json
```

---

## Connector Types

### GitHub

**Tools Provided:**
- `github.sync_status` - Check sync status
- `github.sync_trigger` - Trigger sync
- `github.search_issues` - Search issues
- `github.get_issue` - Get specific issue
- `github.get_pr` - Get pull request
- `github.list_repos` - List repositories

**Configuration:**
```json
{
  "type": "github",
  "config": {
    "token": "ghp_xxxxx",              // Required
    "repository": "owner/repo",         // Required
    "webhook_secret": "secret",         // Optional
    "sync_interval": "5m"               // Optional, default 5m
  }
}
```

**OAuth Configuration:**
```json
{
  "type": "github",
  "config": {
    "auth_type": "oauth",
    "client_id": "xxx",
    "client_secret": "xxx",
    "redirect_uri": "http://localhost:8080/oauth/github/callback",
    "repository": "owner/repo"
  }
}
```

### Slack

**Tools Provided:**
- `slack.search` - Search messages
- `slack.list_channels` - List channels
- `slack.get_thread` - Get thread messages

**Configuration:**
```json
{
  "type": "slack",
  "config": {
    "token": "xoxb-xxxxx",             // Required (bot token)
    "channels": ["C01234", "C56789"],  // Required
    "sync_interval": "2m",             // Optional, default 5m
    "max_messages": 1000               // Optional, default 1000
  }
}
```

**OAuth Configuration:**
```json
{
  "type": "slack",
  "config": {
    "auth_type": "oauth",
    "client_id": "xxx",
    "client_secret": "xxx",
    "redirect_uri": "http://localhost:8080/oauth/slack/callback",
    "team_id": "T01234567"
  }
}
```

### Jira

**Tools Provided:**
- `jira.search` - Search with JQL
- `jira.get_issue` - Get specific issue
- `jira.list_projects` - List projects

**Configuration:**
```json
{
  "type": "jira",
  "config": {
    "base_url": "https://company.atlassian.net",  // Required
    "email": "user@company.com",                   // Required
    "api_token": "xxxxx",                          // Required
    "projects": ["PROJ1", "PROJ2"],                // Required
    "sync_interval": "10m"                         // Optional, default 5m
  }
}
```

**OAuth Configuration:**
```json
{
  "type": "jira",
  "config": {
    "auth_type": "oauth",
    "client_id": "xxx",
    "client_secret": "xxx",
    "redirect_uri": "http://localhost:8080/oauth/jira/callback",
    "base_url": "https://company.atlassian.net",
    "cloud_id": "xxx"
  }
}
```

### Discord

**Tools Provided:**
- `discord.search` - Search messages in channel
- `discord.list_channels` - List channels
- `discord.get_thread` - Get thread messages

**Configuration:**
```json
{
  "type": "discord",
  "config": {
    "token": "xxxxx.xxxxx.xxxxx",      // Required (bot token)
    "guild_id": "123456789",            // Required
    "channels": ["999999", "888888"],   // Required
    "sync_interval": "2m"               // Optional, default 5m
  }
}
```

**OAuth Configuration:**
```json
{
  "type": "discord",
  "config": {
    "auth_type": "oauth",
    "client_id": "xxx",
    "client_secret": "xxx",
    "redirect_uri": "http://localhost:8080/oauth/discord/callback",
    "guild_id": "123456789"
  }
}
```

---

## Dynamic Configuration Examples

### Example 1: Minimal Setup (GitHub Only)

```json
{
  "connectors": [
    {
      "id": "github-main",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "ferg-cod3s/conexus"
      }
    }
  ]
}
```

**Result:**
- Base tools (context.search, etc.)
- GitHub tools only
- No Slack, Jira, or Discord tools

### Example 2: Full Stack (All Connectors)

```json
{
  "connectors": [
    {
      "id": "github-main",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "company/project"
      }
    },
    {
      "id": "slack-eng",
      "type": "slack",
      "config": {
        "token": "${SLACK_TOKEN}",
        "channels": ["C01234"]
      }
    },
    {
      "id": "jira-main",
      "type": "jira",
      "config": {
        "base_url": "https://company.atlassian.net",
        "email": "${JIRA_EMAIL}",
        "api_token": "${JIRA_TOKEN}",
        "projects": ["PROJ"]
      }
    },
    {
      "id": "discord-community",
      "type": "discord",
      "config": {
        "token": "${DISCORD_TOKEN}",
        "guild_id": "123456789",
        "channels": ["999999"]
      }
    }
  ]
}
```

**Result:** All connector tools available

### Example 3: Multi-Workspace Setup

```json
{
  "connectors": [
    // Multiple Slack workspaces
    {
      "id": "slack-engineering",
      "name": "Engineering Team",
      "type": "slack",
      "config": {
        "token": "${SLACK_TOKEN_ENG}",
        "channels": ["C01234", "C56789"]
      }
    },
    {
      "id": "slack-product",
      "name": "Product Team",
      "type": "slack",
      "config": {
        "token": "${SLACK_TOKEN_PROD}",
        "channels": ["C11111", "C22222"]
      }
    },

    // Multiple GitHub repos
    {
      "id": "github-backend",
      "name": "Backend Service",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "company/backend"
      }
    },
    {
      "id": "github-frontend",
      "name": "Frontend App",
      "type": "github",
      "config": {
        "token": "${GITHUB_TOKEN}",
        "repository": "company/frontend"
      }
    }
  ]
}
```

---

## Managing Connectors at Runtime

### List Configured Connectors

```json
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "list"
  }
}
```

**Response:**
```json
{
  "connectors": [
    {
      "id": "github-main",
      "name": "Main Repository",
      "type": "github",
      "status": "active"
    },
    {
      "id": "slack-eng",
      "name": "Engineering Slack",
      "type": "slack",
      "status": "active"
    }
  ]
}
```

### Add Connector

```json
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "add",
    "connector_type": "jira",
    "connector_id": "jira-new",
    "config": {
      "base_url": "https://company.atlassian.net",
      "email": "user@company.com",
      "api_token": "xxxxx",
      "projects": ["NEW"]
    }
  }
}
```

### Remove Connector

```json
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "remove",
    "connector_id": "slack-old"
  }
}
```

### Test Connector

```json
{
  "tool": "context.connector_management",
  "arguments": {
    "action": "test",
    "connector_id": "github-main"
  }
}
```

---

## Best Practices

### 1. Use Environment Variables for Secrets

❌ **Bad:**
```json
{
  "config": {
    "token": "xoxb-actual-token-here"
  }
}
```

✅ **Good:**
```json
{
  "config": {
    "token": "${SLACK_TOKEN}"
  }
}
```

### 2. Descriptive Connector IDs

❌ **Bad:** `github-1`, `slack-2`

✅ **Good:** `github-backend`, `slack-engineering`

### 3. Project-Specific Configs

Keep different projects separated:

```
config/
├── project-a.json
├── project-b.json
└── personal.json
```

### 4. Minimal Connector Set

Only configure connectors you actively use. Less clutter = better UX.

### 5. Document Your Setup

Add a `CONNECTORS.md` to your project:

```markdown
# Connector Setup

## Required Connectors
- GitHub (repo: company/project)
- Slack (channels: #engineering, #general)

## Environment Variables
- GITHUB_TOKEN
- SLACK_TOKEN
```

---

## Troubleshooting

### Tools Not Showing Up

**Problem:** Added a connector but tools aren't appearing

**Solution:**
1. Check connector is properly configured: `context.connector_management` → `list`
2. Restart MCP server to reload tools
3. Verify connector status is "active"

### Too Many Tools

**Problem:** Seeing tools for connectors you don't use

**Solution:** Remove unused connectors via `context.connector_management` → `remove`

### Wrong Connector ID

**Problem:** "Connector not found" error

**Solution:** List connectors to get correct IDs:
```json
{"tool": "context.connector_management", "arguments": {"action": "list"}}
```

### Authentication Errors

**Problem:** Tools fail with auth errors

**Solution:**
1. Test connector: `context.connector_management` → `test`
2. Check token hasn't expired
3. Verify environment variables are set
4. For OAuth: Re-authorize if token expired

---

## Migration Guide

### From Static to Dynamic Configuration

**Old Way (hardcoded):**
All tools always available, connector_id hardcoded in tool calls.

**New Way (dynamic):**
Only configured connector tools available, connector_id explicitly specified.

**Migration Steps:**

1. **Identify your connectors**
   ```bash
   # What connectors do you use?
   # - GitHub for repo X?
   # - Slack for workspace Y?
   # - Jira for project Z?
   ```

2. **Create configuration file**
   ```json
   {
     "connectors": [
       {"id": "...", "type": "github", "config": {...}},
       {"id": "...", "type": "slack", "config": {...}}
     ]
   }
   ```

3. **Test configuration**
   ```bash
   conexus server --config config/connectors.json --dry-run
   ```

4. **Start using dynamic tools**
   - Tools automatically filtered
   - Only see what you configured

---

## Next Steps

1. **Create your config file** - Start with one connector
2. **Test locally** - Verify tools appear correctly
3. **Add more connectors** - As needed
4. **Set up OAuth** - For production use
5. **Document your setup** - For your team

For more information, see:
- [TESTING_QUICK_START.md](./TESTING_QUICK_START.md)
- [OAUTH_IMPLEMENTATION.md](./OAUTH_IMPLEMENTATION.md)
- [CONNECTOR_TESTING_AUTH.md](./CONNECTOR_TESTING_AUTH.md)
