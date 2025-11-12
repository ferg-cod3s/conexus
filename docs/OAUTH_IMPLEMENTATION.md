# OAuth 2.0 Implementation Guide

This guide explains how to use the OAuth 2.0 implementation for Slack, Jira, and Discord connectors.

---

## Overview

Conexus now supports **OAuth 2.0 authentication** for all major connectors:

- ✅ **Slack**: OAuth 2.0 with bot and user tokens
- ✅ **Jira**: OAuth 2.0 (3LO) for Jira Cloud with token refresh
- ✅ **Discord**: OAuth 2.0 with token refresh
- ✅ **GitHub**: Full OAuth support (already implemented)

**Benefits of OAuth:**
- ✅ User-initiated authorization (better UX)
- ✅ Automatic token refresh (no manual rotation)
- ✅ Granular permission scopes
- ✅ Token revocation support
- ✅ Secure token storage

---

## Quick Start

### 1. Create OAuth Apps

#### Slack OAuth App
1. Go to https://api.slack.com/apps
2. Click "Create New App" → "From scratch"
3. Configure:
   - **OAuth & Permissions** → Redirect URLs: `http://localhost:8080/oauth/slack/callback`
   - **Bot Token Scopes**: `channels:history`, `channels:read`, `groups:history`, `search:read`
4. Install to workspace
5. Note: **Client ID** and **Client Secret**

#### Jira OAuth App
1. Go to https://developer.atlassian.com/console/myapps/
2. Create OAuth 2.0 (3LO) app
3. Configure:
   - **Authorization URL**: `http://localhost:8080/oauth/jira/callback`
   - **Scopes**: `read:jira-work`, `read:jira-user`, `offline_access`
4. Note: **Client ID** and **Client Secret**

#### Discord OAuth App
1. Go to https://discord.com/developers/applications
2. Create new application
3. Go to **OAuth2**:
   - Add redirect: `http://localhost:8080/oauth/discord/callback`
   - Scopes: `identify`, `guilds`, `guilds.members.read`
4. Note: **Client ID** and **Client Secret**

---

### 2. Configure OAuth

Create configuration file `config/oauth.json`:

```json
{
  "slack": {
    "client_id": "YOUR_SLACK_CLIENT_ID",
    "client_secret": "YOUR_SLACK_CLIENT_SECRET",
    "redirect_uri": "http://localhost:8080/oauth/slack/callback",
    "scopes": [
      "channels:history",
      "channels:read",
      "groups:history",
      "search:read"
    ]
  },
  "jira": {
    "client_id": "YOUR_JIRA_CLIENT_ID",
    "client_secret": "YOUR_JIRA_CLIENT_SECRET",
    "redirect_uri": "http://localhost:8080/oauth/jira/callback",
    "base_url": "https://your-domain.atlassian.net",
    "scopes": [
      "read:jira-work",
      "read:jira-user",
      "offline_access"
    ]
  },
  "discord": {
    "client_id": "YOUR_DISCORD_CLIENT_ID",
    "client_secret": "YOUR_DISCORD_CLIENT_SECRET",
    "redirect_uri": "http://localhost:8080/oauth/discord/callback",
    "scopes": [
      "identify",
      "guilds",
      "guilds.members.read"
    ]
  }
}
```

**IMPORTANT:** Never commit this file! Add to `.gitignore`:
```bash
echo "config/oauth.json" >> .gitignore
```

---

### 3. Start OAuth Server

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/ferg-cod3s/conexus/internal/oauth"
    slackauth "github.com/ferg-cod3s/conexus/internal/security/slack"
    jiraauth "github.com/ferg-cod3s/conexus/internal/security/jira"
    discordauth "github.com/ferg-cod3s/conexus/internal/security/discord"
)

func main() {
    // Load OAuth configuration (from config file or env vars)
    config := loadOAuthConfig()

    // Create auth managers
    slackAuth, _ := slackauth.NewAuthManager(&slackauth.AuthConfig{
        ClientID:     config.Slack.ClientID,
        ClientSecret: config.Slack.ClientSecret,
        RedirectURI:  config.Slack.RedirectURI,
        Scopes:       config.Slack.Scopes,
        AuthType:     "oauth",
    }, nil)

    jiraAuth, _ := jiraauth.NewAuthManager(&jiraauth.AuthConfig{
        ClientID:     config.Jira.ClientID,
        ClientSecret: config.Jira.ClientSecret,
        RedirectURI:  config.Jira.RedirectURI,
        BaseURL:      config.Jira.BaseURL,
        Scopes:       config.Jira.Scopes,
        AuthType:     "oauth",
    }, nil)

    discordAuth, _ := discordauth.NewAuthManager(&discordauth.AuthConfig{
        ClientID:     config.Discord.ClientID,
        ClientSecret: config.Discord.ClientSecret,
        RedirectURI:  config.Discord.RedirectURI,
        Scopes:       config.Discord.Scopes,
        AuthType:     "oauth",
    }, nil)

    // Create OAuth handler
    oauthHandler := oauth.NewHandler(&oauth.Config{
        SlackAuth:   slackAuth,
        JiraAuth:    jiraAuth,
        DiscordAuth: discordAuth,
        OnSuccess: func(provider string, token interface{}) {
            fmt.Printf("✓ %s OAuth successful!\n", provider)
            fmt.Printf("Token: %+v\n", token)
        },
        OnError: func(provider string, err error) {
            fmt.Printf("✗ %s OAuth failed: %v\n", provider, err)
        },
    })

    // Register routes
    mux := http.NewServeMux()
    oauthHandler.RegisterRoutes(mux)

    // Start server
    fmt.Println("OAuth server running on http://localhost:8080")
    fmt.Println("\nTo authorize:")
    fmt.Println("  Slack:   http://localhost:8080/oauth/slack/authorize")
    fmt.Println("  Jira:    http://localhost:8080/oauth/jira/authorize")
    fmt.Println("  Discord: http://localhost:8080/oauth/discord/authorize")

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

---

### 4. Authorize Connectors

Open in your browser:

- **Slack**: http://localhost:8080/oauth/slack/authorize
- **Jira**: http://localhost:8080/oauth/jira/authorize
- **Discord**: http://localhost:8080/oauth/discord/authorize

**Flow:**
1. Browser redirects to provider's authorization page
2. User approves requested permissions
3. Provider redirects back to callback URL
4. Code exchanged for access token
5. Token stored securely
6. Success page displayed

---

## API Reference

### Slack OAuth

```go
import "github.com/ferg-cod3s/conexus/internal/security/slack"

// Create auth manager
authManager, err := slack.NewAuthManager(&slack.AuthConfig{
    ClientID:     "YOUR_CLIENT_ID",
    ClientSecret: "YOUR_CLIENT_SECRET",
    RedirectURI:  "http://localhost:8080/oauth/slack/callback",
    Scopes: []string{
        "channels:history",
        "channels:read",
        "search:read",
    },
    AuthType: "oauth",
}, tokenStore)

// Get authorization URL
authURL, err := authManager.GetAuthURL("random_state", []string{"search:read"})

// Exchange code for token
token, err := authManager.ExchangeCode(ctx, code)

// Get access token (auto-refreshes if needed)
accessToken, err := authManager.GetToken(ctx)

// Validate token
err = authManager.ValidateToken(ctx)

// Revoke token
err = authManager.RevokeToken(ctx)
```

**Scopes:**
- `channels:history` - Read public channel messages
- `channels:read` - List public channels
- `groups:history` - Read private channel messages
- `groups:read` - List private channels
- `search:read` - Search messages
- `im:history` - Read direct messages

**Token Expiry:** Slack tokens don't expire by default

---

### Jira OAuth

```go
import "github.com/ferg-cod3s/conexus/internal/security/jira"

// Create auth manager
authManager, err := jira.NewAuthManager(&jira.AuthConfig{
    ClientID:     "YOUR_CLIENT_ID",
    ClientSecret: "YOUR_CLIENT_SECRET",
    RedirectURI:  "http://localhost:8080/oauth/jira/callback",
    BaseURL:      "https://your-domain.atlassian.net",
    Scopes: []string{
        "read:jira-work",
        "read:jira-user",
        "offline_access",
    },
    AuthType: "oauth",
}, tokenStore)

// Get authorization URL
authURL, err := authManager.GetAuthURL("random_state")

// Exchange code for token
token, err := authManager.ExchangeCode(ctx, code)
// Returns token with CloudID and accessible resources

// Refresh token (automatically called when expired)
newToken, err := authManager.RefreshToken(ctx)

// Get access token (auto-refreshes)
accessToken, err := authManager.GetToken(ctx)

// Get token info
tokenInfo, err := authManager.GetTokenInfo(ctx)
```

**Scopes:**
- `read:jira-work` - Read issues, projects, boards
- `read:jira-user` - Read user information
- `write:jira-work` - Create/update issues
- `offline_access` - Get refresh token

**Token Expiry:** Jira tokens expire after 1 hour (auto-refresh)

---

### Discord OAuth

```go
import "github.com/ferg-cod3s/conexus/internal/security/discord"

// Create auth manager
authManager, err := discord.NewAuthManager(&discord.AuthConfig{
    ClientID:     "YOUR_CLIENT_ID",
    ClientSecret: "YOUR_CLIENT_SECRET",
    RedirectURI:  "http://localhost:8080/oauth/discord/callback",
    Scopes: []string{
        "identify",
        "guilds",
        "guilds.members.read",
    },
    AuthType: "oauth",
}, tokenStore)

// Get authorization URL
authURL, err := authManager.GetAuthURL("random_state")

// Exchange code for token
token, err := authManager.ExchangeCode(ctx, code)

// Refresh token
newToken, err := authManager.RefreshToken(ctx, userID)

// Get access token
accessToken, err := authManager.GetToken(ctx)

// Revoke token
err = authManager.RevokeToken(ctx, userID)
```

**Scopes:**
- `identify` - Read user info (username, avatar, etc.)
- `email` - Read user's email
- `guilds` - Read user's guilds
- `guilds.members.read` - Read guild members
- `messages.read` - Read message history (requires bot)

**Token Expiry:** Discord tokens expire after 7 days (auto-refresh)

---

## Token Storage

All OAuth implementations support pluggable token storage:

### In-Memory Storage (Default)

```go
// Automatically used if no store provided
authManager, _ := slack.NewAuthManager(config, nil)
```

**Pros:** Fast, no dependencies
**Cons:** Tokens lost on restart

### Custom Storage Implementation

Implement the `TokenStore` interface:

```go
type TokenStore interface {
    Store(ctx context.Context, key string, token *SecureToken) error
    Retrieve(ctx context.Context, key string) (*SecureToken, error)
    Delete(ctx context.Context, key string) error
}
```

**Example: Database Storage**

```go
type DBTokenStore struct {
    db *sql.DB
}

func (s *DBTokenStore) Store(ctx context.Context, key string, token *SecureToken) error {
    // Encrypt token before storing
    encrypted, _ := encrypt(token)

    _, err := s.db.ExecContext(ctx,
        "INSERT INTO tokens (key, data, expires_at) VALUES (?, ?, ?) ON CONFLICT (key) DO UPDATE SET data = ?, expires_at = ?",
        key, encrypted, token.ExpiresAt, encrypted, token.ExpiresAt)
    return err
}

func (s *DBTokenStore) Retrieve(ctx context.Context, key string) (*SecureToken, error) {
    var encrypted []byte
    err := s.db.QueryRowContext(ctx,
        "SELECT data FROM tokens WHERE key = ? AND expires_at > NOW()",
        key).Scan(&encrypted)
    if err != nil {
        return nil, err
    }

    // Decrypt token
    token, _ := decrypt(encrypted)
    return token, nil
}
```

---

## Security Best Practices

### 1. State Parameter (CSRF Protection)

Always use a random state parameter:

```go
import "crypto/rand"
import "encoding/base64"

func generateState() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

### 2. Secure Token Storage

- ✅ Encrypt tokens at rest
- ✅ Use secure key management (HashiCorp Vault, AWS KMS, etc.)
- ✅ Never log tokens
- ✅ Set appropriate database permissions

### 3. HTTPS in Production

Always use HTTPS for OAuth callbacks:

```go
redirect_uri: "https://your-domain.com/oauth/slack/callback"
```

### 4. Token Rotation

Tokens are automatically refreshed before expiry:

```go
// Jira tokens refresh 5 minutes before expiry
// Discord tokens refresh when expired
// Slack tokens don't expire (but can be revoked)
```

### 5. Scope Minimization

Only request necessary scopes:

```go
// Bad - too many permissions
scopes: ["admin", "write:everything"]

// Good - minimal permissions
scopes: ["read:jira-work", "read:jira-user"]
```

---

## Backward Compatibility

OAuth is **fully backward compatible** with token-based auth:

### Old Way (Still Works)

```go
slack.NewConnector(&slack.Config{
    Token: "xoxb-your-bot-token",
    Channels: []string{"C123456"},
})
```

### New Way (OAuth)

```go
authManager, _ := slack.NewAuthManager(&slack.AuthConfig{
    AuthType: "oauth",
    ClientID: "...",
    ClientSecret: "...",
}, nil)

// After OAuth flow completes
token, _ := authManager.GetToken(ctx)

// Use token with connector
slack.NewConnector(&slack.Config{
    Token: token,
    Channels: []string{"C123456"},
})
```

---

## Migration Guide

### From Bot Tokens to OAuth

**Step 1: Keep existing token config**
```go
// This still works!
config := &slack.Config{
    Token: "xoxb-old-token",
}
```

**Step 2: Add OAuth configuration**
```go
oauthConfig := &slack.AuthConfig{
    ClientID: "...",
    ClientSecret: "...",
    AuthType: "oauth",
}
```

**Step 3: Let users re-authorize**
- Show "Connect with Slack" button
- Users go through OAuth flow
- New OAuth token replaces old token
- Old token can be revoked

**Step 4: Update connector**
```go
// New token from OAuth
newToken, _ := authManager.GetToken(ctx)

// Update connector config
connector.UpdateToken(newToken)
```

---

## Troubleshooting

### "invalid_client" Error

**Cause:** Wrong Client ID or Secret

**Fix:** Verify credentials from OAuth app settings

### "redirect_uri_mismatch" Error

**Cause:** Redirect URI doesn't match registered URI

**Fix:** Ensure exact match (including http/https and trailing slash)

### "insufficient_scope" Error

**Cause:** Token doesn't have required permissions

**Fix:** Add missing scopes and re-authorize

### Token Refresh Fails

**Cause:** Refresh token expired or invalid

**Fix:** User must re-authorize through OAuth flow

### State Validation Failed

**Cause:** Possible CSRF attack or expired state

**Fix:** States expire after 10 minutes - try again

---

## Production Deployment

### Environment Variables

```bash
export SLACK_CLIENT_ID="..."
export SLACK_CLIENT_SECRET="..."
export JIRA_CLIENT_ID="..."
export JIRA_CLIENT_SECRET="..."
export DISCORD_CLIENT_ID="..."
export DISCORD_CLIENT_SECRET="..."

export OAUTH_REDIRECT_BASE="https://your-domain.com"
export TOKEN_ENCRYPTION_KEY="..." # 32-byte key for token encryption
```

### Docker Compose

```yaml
version: '3.8'
services:
  conexus:
    image: conexus:latest
    ports:
      - "8080:8080"
    environment:
      - SLACK_CLIENT_ID=${SLACK_CLIENT_ID}
      - SLACK_CLIENT_SECRET=${SLACK_CLIENT_SECRET}
      - JIRA_CLIENT_ID=${JIRA_CLIENT_ID}
      - JIRA_CLIENT_SECRET=${JIRA_CLIENT_SECRET}
      - DISCORD_CLIENT_ID=${DISCORD_CLIENT_ID}
      - DISCORD_CLIENT_SECRET=${DISCORD_CLIENT_SECRET}
      - OAUTH_REDIRECT_BASE=https://your-domain.com
    volumes:
      - ./data:/data
```

### Kubernetes Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oauth-credentials
type: Opaque
stringData:
  slack-client-id: "..."
  slack-client-secret: "..."
  jira-client-id: "..."
  jira-client-secret: "..."
  discord-client-id: "..."
  discord-client-secret: "..."
```

---

## Next Steps

1. **Set up OAuth apps** for your platforms
2. **Test locally** with the example server
3. **Implement custom token storage** for production
4. **Add OAuth UI** to your application
5. **Monitor token refresh** and handle errors
6. **Set up alerting** for auth failures

For more information, see:
- [TESTING_QUICK_START.md](./TESTING_QUICK_START.md)
- [CONNECTOR_TESTING_AUTH.md](./CONNECTOR_TESTING_AUTH.md)
