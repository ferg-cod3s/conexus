# Connector Testing & Authentication Guide

## Overview

This guide covers authentication mechanisms and testing strategies for all Conexus connectors.

---

## Authentication Methods by Connector

### 1. GitHub Connector

**Current Implementation:** ✅ Full OAuth & App Support

**Location:** `internal/security/github/auth.go`

**Supported Auth Methods:**

#### A. Personal Access Token (PAT) - Simplest
```json
{
  "type": "github",
  "config": {
    "auth_type": "token",
    "token": "ghp_xxxxxxxxxxxxx",
    "repository": "owner/repo"
  }
}
```

**Setup:**
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token (classic)
3. Required scopes: `repo`, `read:org`, `read:discussion`

#### B. GitHub App - Recommended for Organizations
```json
{
  "type": "github",
  "config": {
    "auth_type": "app",
    "app_id": "123456",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...",
    "installation_id": "12345678",
    "repository": "owner/repo"
  }
}
```

**Setup:**
1. Create GitHub App: Settings → Developer settings → GitHub Apps
2. Generate private key
3. Install app to repository
4. Get installation ID from webhook or API

#### C. OAuth 2.0 - For User Authorization
```json
{
  "type": "github",
  "config": {
    "auth_type": "oauth",
    "client_id": "Iv1.xxxxxxxxxxxx",
    "client_secret": "xxxxxxxxxxxxx",
    "repository": "owner/repo"
  }
}
```

**Features:**
- ✅ Token validation
- ✅ Token rotation
- ✅ Webhook signature verification
- ✅ Secure token storage

---

### 2. Slack Connector

**Current Implementation:** 🔶 Bot Token Only

**Location:** `internal/connectors/slack/slack.go`

**Current Auth Method:**

#### Bot Token (xoxb-*)
```json
{
  "type": "slack",
  "config": {
    "token": "xoxb-xxxxxxxxxxxxx",
    "channels": ["C01234567", "C98765432"]
  }
}
```

**Setup:**
1. Create Slack App: https://api.slack.com/apps
2. Add Bot Token Scopes:
   - `channels:history` - Read public channel messages
   - `channels:read` - List public channels
   - `groups:history` - Read private channel messages
   - `groups:read` - List private channels
   - `im:history` - Read direct messages
   - `search:read` - Search messages
3. Install app to workspace
4. Copy Bot User OAuth Token

**OAuth 2.0 Flow (NOT YET IMPLEMENTED):**
```json
{
  "type": "slack",
  "config": {
    "auth_type": "oauth",
    "client_id": "xxx.yyy",
    "client_secret": "xxxxxxxxxxxxx",
    "redirect_uri": "https://your-app.com/slack/callback"
  }
}
```

**To Implement:**
- [ ] OAuth authorization flow
- [ ] Token refresh handling
- [ ] User token vs bot token handling

---

### 3. Jira Connector

**Current Implementation:** 🔶 API Token Only

**Location:** `internal/connectors/jira/jira.go`

**Current Auth Method:**

#### API Token (Cloud)
```json
{
  "type": "jira",
  "config": {
    "base_url": "https://your-domain.atlassian.net",
    "email": "user@example.com",
    "api_token": "xxxxxxxxxxxxx",
    "projects": ["PROJ", "TEST"]
  }
}
```

**Setup (Cloud):**
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Create API token
3. Use email + API token for Basic Auth

**OAuth 2.0 (3LO) - NOT YET IMPLEMENTED:**
```json
{
  "type": "jira",
  "config": {
    "auth_type": "oauth",
    "base_url": "https://your-domain.atlassian.net",
    "client_id": "xxxxxxxxxxxxx",
    "client_secret": "xxxxxxxxxxxxx",
    "redirect_uri": "https://your-app.com/jira/callback"
  }
}
```

**To Implement:**
- [ ] OAuth 2.0 (3LO) for Cloud
- [ ] OAuth 1.0a for Server/Data Center
- [ ] Token refresh
- [ ] Multi-tenant support

---

### 4. Discord Connector

**Current Implementation:** 🔶 Bot Token Only

**Location:** `internal/connectors/discord/discord.go`

**Current Auth Method:**

#### Bot Token
```json
{
  "type": "discord",
  "config": {
    "token": "xxxxxxxxxxxxx.xxxxxx.xxxxxxxxxxxxx",
    "guild_id": "123456789012345678",
    "channels": ["123456789", "987654321"]
  }
}
```

**Setup:**
1. Go to https://discord.com/developers/applications
2. Create New Application
3. Go to Bot → Add Bot
4. Copy Bot Token
5. Enable required Privileged Gateway Intents:
   - MESSAGE CONTENT INTENT
   - SERVER MEMBERS INTENT (optional)
6. Generate invite URL (OAuth2 → URL Generator):
   - Scopes: `bot`, `applications.commands`
   - Permissions: `Read Messages`, `Read Message History`, `View Channels`

**OAuth 2.0 - NOT YET IMPLEMENTED:**
```json
{
  "type": "discord",
  "config": {
    "auth_type": "oauth",
    "client_id": "123456789012345678",
    "client_secret": "xxxxxxxxxxxxx",
    "redirect_uri": "https://your-app.com/discord/callback"
  }
}
```

**To Implement:**
- [ ] OAuth 2.0 user authorization
- [ ] Token refresh
- [ ] Guild-specific permissions

---

## Testing Strategies

### Unit Testing

**Mock Client Approach (Already Implemented):**

Each connector has a mock client interface:
- `internal/connectors/github/client_interface.go` → `MockGitHubClient`
- `internal/connectors/slack/client_interface.go` → `MockSlackClient`
- `internal/connectors/jira/client_interface.go` → `MockJiraClient`
- `internal/connectors/discord/client_interface.go` → `MockDiscordClient`

**Example Test:**
```go
func TestGitHubConnector_SearchIssues(t *testing.T) {
    mockClient := github.NewMockGitHubClient()
    // Configure mock responses

    connector := &github.Connector{
        client: mockClient,
        config: &github.Config{...},
    }

    issues, err := connector.SearchIssues(ctx, "is:open", "open")
    // Assert results
}
```

**Benefits:**
- ✅ Fast (no network calls)
- ✅ No external dependencies
- ✅ Deterministic results
- ✅ Can test error conditions

---

### Integration Testing

**Approach 1: Test Credentials in Environment Variables**

```bash
# Set up test credentials
export CONEXUS_TEST_GITHUB_TOKEN="ghp_xxxxxxxxxxxxx"
export CONEXUS_TEST_SLACK_TOKEN="xoxb-xxxxxxxxxxxxx"
export CONEXUS_TEST_JIRA_TOKEN="xxxxxxxxxxxxx"
export CONEXUS_TEST_JIRA_EMAIL="test@example.com"
export CONEXUS_TEST_DISCORD_TOKEN="xxxxxxxxxxxxx"

# Run integration tests
go test -v -tags=integration ./internal/connectors/...
```

**Test File Structure:**
```go
// +build integration

package github_test

import (
    "os"
    "testing"
)

func TestGitHubConnector_Integration(t *testing.T) {
    token := os.Getenv("CONEXUS_TEST_GITHUB_TOKEN")
    if token == "" {
        t.Skip("CONEXUS_TEST_GITHUB_TOKEN not set")
    }

    config := &github.Config{
        Token: token,
        Repository: "ferg-cod3s/conexus",
    }

    connector, err := github.NewConnector(config)
    // Run real API tests
}
```

**Approach 2: Test Configuration File**

Create `config/test_connectors.json`:
```json
{
  "github": {
    "token": "ghp_xxxxxxxxxxxxx",
    "repository": "ferg-cod3s/conexus"
  },
  "slack": {
    "token": "xoxb-xxxxxxxxxxxxx",
    "channels": ["C01234567"]
  },
  "jira": {
    "base_url": "https://test.atlassian.net",
    "email": "test@example.com",
    "api_token": "xxxxxxxxxxxxx",
    "projects": ["TEST"]
  },
  "discord": {
    "token": "xxxxxxxxxxxxx",
    "guild_id": "123456789012345678",
    "channels": ["123456789"]
  }
}
```

**Never commit this file!** Add to `.gitignore`:
```gitignore
config/test_connectors.json
config/*_credentials.json
*.env
```

**Approach 3: Docker Test Environment**

Create isolated test instances:
- Jira: Use Jira Cloud trial or local Jira instance
- Slack: Create test workspace
- Discord: Create test server
- GitHub: Use test organization/repository

---

### End-to-End Testing via MCP

**Test MCP Tools:**
```bash
# Start Conexus MCP server
conexus server --config config/test_config.json

# Test with MCP client or Claude Desktop
# Send MCP tool calls:
```

**Example MCP Test Cases:**

```json
// 1. List GitHub repositories
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "github.list_repos",
    "arguments": {
      "connector_id": "github-test"
    }
  }
}

// 2. Search Slack messages
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "slack.search",
    "arguments": {
      "connector_id": "slack-test",
      "query": "conexus"
    }
  }
}

// 3. Search Jira issues
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "jira.search",
    "arguments": {
      "connector_id": "jira-test",
      "jql": "project = TEST AND status = Open"
    }
  }
}

// 4. Search Discord messages
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "discord.search",
    "arguments": {
      "connector_id": "discord-test",
      "channel_id": "123456789",
      "query": "test"
    }
  }
}
```

---

## Recommended Testing Approach

### Phase 1: Unit Tests (Current State)
✅ Use existing mock clients
✅ Test core logic without external dependencies
✅ Fast, reliable, no setup required

### Phase 2: Manual Integration Testing
1. Create test accounts/workspaces for each service
2. Generate test credentials (tokens/API keys)
3. Store in local config file (NOT in git)
4. Manually test each connector via MCP

### Phase 3: Automated Integration Tests
1. Set up CI/CD with encrypted secrets
2. Use service-specific test environments
3. Run integration tests on PR/merge

### Phase 4: OAuth Implementation
1. Implement OAuth flows for each connector
2. Add web-based authorization flow
3. Implement token refresh
4. Add secure token storage (encrypted database)

---

## OAuth Implementation Roadmap

### Priority 1: GitHub (Already Implemented) ✅
- [x] Personal Access Token
- [x] GitHub App
- [x] OAuth 2.0
- [x] Token validation
- [x] Token rotation

### Priority 2: Slack OAuth
```go
// internal/security/slack/auth.go
type SlackAuth struct {
    ClientID     string
    ClientSecret string
    RedirectURI  string
}

func (sa *SlackAuth) GetAuthURL(state string) string
func (sa *SlackAuth) ExchangeCode(code string) (*oauth2.Token, error)
func (sa *SlackAuth) RefreshToken(refreshToken string) (*oauth2.Token, error)
```

**OAuth Scopes:**
- `channels:history`, `channels:read`
- `groups:history`, `groups:read`
- `im:history`, `search:read`

### Priority 3: Jira OAuth
```go
// internal/security/jira/auth.go
type JiraAuth struct {
    CloudID      string
    ClientID     string
    ClientSecret string
    RedirectURI  string
}

// OAuth 2.0 (3LO) for Cloud
func (ja *JiraAuth) GetAuthURL(state string) string
func (ja *JiraAuth) ExchangeCode(code string) (*oauth2.Token, error)
```

**OAuth Scopes:**
- `read:jira-work`
- `read:jira-user`
- `offline_access` (for refresh tokens)

### Priority 4: Discord OAuth
```go
// internal/security/discord/auth.go
type DiscordAuth struct {
    ClientID     string
    ClientSecret string
    RedirectURI  string
}

func (da *DiscordAuth) GetAuthURL(state string, scopes []string) string
func (da *DiscordAuth) ExchangeCode(code string) (*oauth2.Token, error)
```

**OAuth Scopes:**
- `identify`
- `guilds`
- `guilds.members.read`
- `messages.read`

---

## Security Considerations

### Token Storage
- **Never commit tokens to git**
- Use environment variables or encrypted config
- Consider using system keychain (macOS/Windows) or secret manager
- Implement token encryption at rest

### Token Rotation
- Implement automatic token refresh for OAuth
- Handle token expiration gracefully
- Log token rotation events

### Rate Limiting
- All connectors already track rate limits via `GetRateLimit()`
- Implement backoff when approaching limits
- Log rate limit warnings

### Webhook Security
- GitHub webhook signature verification already implemented
- Add similar verification for Slack, Discord events
- Use secure webhook secrets

---

## Next Steps

1. **Create Test Credentials:**
   - Set up test accounts for each service
   - Generate API tokens/bot tokens
   - Document in team password manager (NOT in git)

2. **Manual Testing:**
   - Test each connector with real credentials
   - Verify all MCP tools work end-to-end
   - Document any issues or limitations

3. **Automated Tests:**
   - Add integration test suite
   - Set up CI/CD with secrets
   - Add test coverage reporting

4. **OAuth Implementation:**
   - Implement OAuth for Slack
   - Implement OAuth for Jira
   - Implement OAuth for Discord
   - Add web UI for authorization flow

5. **Production Readiness:**
   - Implement secure token storage
   - Add token encryption
   - Add audit logging
   - Add monitoring and alerting

---

## Quick Start for Local Testing

1. **Create local config:**
```bash
mkdir -p ~/.conexus
cat > ~/.conexus/test_config.json <<EOF
{
  "connectors": {
    "github-test": {
      "type": "github",
      "auth_type": "token",
      "token": "YOUR_GITHUB_TOKEN",
      "repository": "owner/repo"
    },
    "slack-test": {
      "type": "slack",
      "token": "YOUR_SLACK_BOT_TOKEN",
      "channels": ["CHANNEL_ID"]
    }
  }
}
EOF
chmod 600 ~/.conexus/test_config.json
```

2. **Test connectors:**
```bash
# Use environment variables (safer)
export GITHUB_TOKEN="ghp_xxxxx"
export SLACK_TOKEN="xoxb-xxxxx"

# Or point to config file
conexus test-connectors --config ~/.conexus/test_config.json
```

3. **Manual MCP testing:**
```bash
# Start server
conexus server --config ~/.conexus/test_config.json

# In another terminal, use mcp-client or Claude Desktop
# to test tool calls
```

---

## Resources

- [GitHub Authentication](https://docs.github.com/en/authentication)
- [Slack API Authentication](https://api.slack.com/authentication)
- [Jira Cloud OAuth 2.0](https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/)
- [Discord OAuth2](https://discord.com/developers/docs/topics/oauth2)
