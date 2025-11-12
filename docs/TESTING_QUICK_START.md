# Quick Start: Testing Connectors

This guide will help you quickly set up and test the Conexus connectors.

## TL;DR - Fastest Path to Testing

### Option 1: Environment Variables (Recommended)

```bash
# Set your test credentials
export CONEXUS_TEST_GITHUB_TOKEN="ghp_xxxxxxxxxxxxx"
export CONEXUS_TEST_GITHUB_REPO="owner/repo"

export CONEXUS_TEST_SLACK_TOKEN="xoxb-xxxxxxxxxxxxx"

export CONEXUS_TEST_JIRA_URL="https://your-domain.atlassian.net"
export CONEXUS_TEST_JIRA_EMAIL="your@email.com"
export CONEXUS_TEST_JIRA_TOKEN="xxxxxxxxxxxxx"

export CONEXUS_TEST_DISCORD_TOKEN="YOUR_DISCORD_BOT_TOKEN"
export CONEXUS_TEST_DISCORD_GUILD="123456789012345678"

# Run integration tests
go test -v -tags=integration ./internal/connectors/...
```

### Option 2: Config File

```bash
# Create config file
mkdir -p config
cat > config/test_connectors.json <<EOF
{
  "github": {
    "token": "ghp_xxxxxxxxxxxxx",
    "repository": "owner/repo"
  },
  "slack": {
    "token": "xoxb-xxxxxxxxxxxxx",
    "channels": ["C01234567"]
  },
  "jira": {
    "base_url": "https://your-domain.atlassian.net",
    "email": "your@email.com",
    "api_token": "xxxxxxxxxxxxx",
    "projects": ["PROJ"]
  },
  "discord": {
    "token": "YOUR_DISCORD_BOT_TOKEN",
    "guild_id": "123456789012345678",
    "channels": ["123456789"]
  }
}
EOF

# IMPORTANT: Protect the file
chmod 600 config/test_connectors.json

# IMPORTANT: Never commit this file!
echo "config/test_connectors.json" >> .gitignore

# Run tests
CONEXUS_TEST_CONFIG=config/test_connectors.json go test -v -tags=integration ./internal/connectors/...
```

---

## Getting Test Credentials

### GitHub (Easiest)

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Select scopes: `repo`, `read:org`, `read:discussion`
4. Copy token → starts with `ghp_`

**Token Format:** `ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

### Slack (Requires App Creation)

1. Go to https://api.slack.com/apps
2. Click "Create New App" → "From scratch"
3. Give it a name and select your workspace
4. Go to "OAuth & Permissions"
5. Add Bot Token Scopes:
   - `channels:history`
   - `channels:read`
   - `groups:history` (for private channels)
   - `groups:read`
   - `search:read` (for search)
6. Install app to workspace
7. Copy "Bot User OAuth Token" → starts with `xoxb-`

**Token Format:** `xoxb-xxxxxxxxxxxxx-xxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx`

**Get Channel IDs:**
- Right-click channel → Copy link
- URL format: `https://your-workspace.slack.com/archives/C01234567`
- Channel ID is the last part: `C01234567`

### Jira (Cloud)

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a name (e.g., "Conexus Test")
4. Copy token
5. You'll also need:
   - Your Jira URL: `https://your-domain.atlassian.net`
   - Your email address

**Token Format:** `ATATT3xFfGF0xxxxxxxxxxxxxxxxxxxxx` (starts with `ATATT`)

### Discord (Requires Bot Creation)

1. Go to https://discord.com/developers/applications
2. Click "New Application"
3. Give it a name
4. Go to "Bot" → "Add Bot"
5. Under "Privileged Gateway Intents", enable:
   - MESSAGE CONTENT INTENT
   - SERVER MEMBERS INTENT (optional)
6. Copy bot token → long string with dots

**Token Format:** `XXXX...XXXX.YYYYYY.ZZZZZZ...ZZZZZZ` (long string with dots)

**Invite Bot to Server:**
1. Go to "OAuth2" → "URL Generator"
2. Select scopes: `bot`, `applications.commands`
3. Select permissions: `Read Messages`, `Read Message History`, `View Channels`
4. Copy generated URL and open in browser
5. Select your test server and authorize

**Get Guild ID:**
- Right-click your server icon → Copy Server ID
- (Enable Developer Mode in Discord Settings if you don't see this)

---

## Running Tests

### Unit Tests (No Credentials Needed)

```bash
# Run all unit tests
go test ./internal/connectors/...

# Run specific connector tests
go test ./internal/connectors/github/...
go test ./internal/connectors/slack/...
go test ./internal/connectors/jira/...
go test ./internal/connectors/discord/...
```

### Integration Tests (Requires Credentials)

```bash
# Run all integration tests
go test -v -tags=integration ./internal/connectors/...

# Run specific connector integration tests
go test -v -tags=integration ./internal/connectors/ -run TestGitHubConnector_Integration
go test -v -tags=integration ./internal/connectors/ -run TestSlackConnector_Integration
go test -v -tags=integration ./internal/connectors/ -run TestJiraConnector_Integration
go test -v -tags=integration ./internal/connectors/ -run TestDiscordConnector_Integration
```

### Manual MCP Testing

```bash
# Start Conexus MCP server
conexus server

# In another terminal, test with MCP client
# (or use Claude Desktop configured to use Conexus)
```

---

## What Gets Tested

### GitHub Connector Tests
- ✅ Connector initialization
- ✅ BaseConnector interface compliance
- ✅ GetType() returns "github"
- ✅ GetCapabilities() shows search support
- ✅ GetRateLimit() returns rate limit info
- ✅ GetSyncStatus() returns sync status
- ✅ SearchIssues() finds open issues
- ✅ ListRepositories() lists accessible repos
- ✅ GetIssue() retrieves specific issue
- ✅ GetIssueComments() fetches issue comments

### Slack Connector Tests
- ✅ Connector initialization
- ✅ BaseConnector interface compliance
- ✅ GetType() returns "slack"
- ✅ ListChannels() lists workspace channels
- ✅ SearchMessages() searches for messages (if scope available)

### Jira Connector Tests
- ✅ Connector initialization
- ✅ BaseConnector interface compliance
- ✅ GetType() returns "jira"
- ✅ ListProjects() lists accessible projects
- ✅ SearchIssues() executes JQL queries
- ✅ GetIssue() retrieves specific issue
- ✅ GetIssueComments() fetches issue comments

### Discord Connector Tests
- ✅ Connector initialization
- ✅ BaseConnector interface compliance
- ✅ GetType() returns "discord"
- ✅ ListChannels() lists server channels
- ✅ SearchMessages() searches in channel
- ✅ GetGuild() retrieves server info

---

## Troubleshooting

### Tests Skip with "credentials not configured"

This is normal! Tests automatically skip if credentials aren't set:

```
=== RUN   TestGitHubConnector_Integration
--- SKIP: TestGitHubConnector_Integration (0.00s)
    integration_test.go:50: GitHub test credentials not configured
```

**Solution:** Set the appropriate environment variables or config file.

### "401 Unauthorized" or "403 Forbidden"

Your token is invalid, expired, or lacks required scopes.

**Solutions:**
- Regenerate token with correct scopes
- Check token hasn't expired
- Verify token is correctly copied (no extra spaces)

### "404 Not Found" for GitHub

Repository doesn't exist or token doesn't have access.

**Solutions:**
- Check repository name format: `owner/repo`
- Verify token has `repo` scope
- Ensure you have access to the repository

### "channel_not_found" for Slack

Channel ID is incorrect or bot isn't in the channel.

**Solutions:**
- Verify channel ID format: `C01234567`
- Invite bot to the channel
- Check bot has correct permissions

### "invalid_auth" for Jira

Email/token combination is incorrect.

**Solutions:**
- Verify Jira URL is correct (Cloud vs Server)
- Check API token hasn't been revoked
- Ensure email matches Jira account

### "401 Unauthorized" for Discord

Bot token is invalid or bot isn't in the server.

**Solutions:**
- Regenerate bot token
- Re-invite bot to server with correct permissions
- Enable required Gateway Intents

---

## Security Best Practices

✅ **DO:**
- Use test accounts/workspaces
- Store tokens in environment variables
- Use config files with `chmod 600` permissions
- Add config files to `.gitignore`
- Rotate tokens regularly
- Use minimal required scopes/permissions

❌ **DON'T:**
- Commit tokens to git
- Share tokens in chat/email
- Use production accounts for testing
- Grant excessive permissions
- Leave tokens in terminal history

---

## Next Steps

After testing locally:

1. **Set up CI/CD testing** - Use GitHub Actions secrets
2. **Implement OAuth** - Replace static tokens with OAuth flow
3. **Add monitoring** - Track connector health and rate limits
4. **Production deployment** - Use secure secret management

For detailed OAuth implementation guide, see [CONNECTOR_TESTING_AUTH.md](./CONNECTOR_TESTING_AUTH.md).
