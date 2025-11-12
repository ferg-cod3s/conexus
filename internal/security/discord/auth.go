package discord

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// Discord OAuth 2.0 endpoints
// Reference: https://discord.com/developers/docs/topics/oauth2
const (
	discordAuthURL   = "https://discord.com/api/oauth2/authorize"
	discordTokenURL  = "https://discord.com/api/oauth2/token"
	discordRevokeURL = "https://discord.com/api/oauth2/token/revoke"
	discordAPIURL    = "https://discord.com/api/v10"
)

// AuthManager manages Discord OAuth authentication
type AuthManager struct {
	config       *AuthConfig
	oauth2Config *oauth2.Config
	tokenStore   TokenStore
	mu           sync.RWMutex
}

// AuthConfig contains Discord OAuth configuration
type AuthConfig struct {
	// OAuth App credentials
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`

	// Bot token (for backward compatibility with non-OAuth)
	BotToken string `json:"bot_token,omitempty"`

	// User token (from OAuth flow)
	UserToken string `json:"user_token,omitempty"`

	// OAuth scopes to request
	Scopes []string `json:"scopes,omitempty"`

	// Guild ID (for validation/filtering)
	GuildID string `json:"guild_id,omitempty"`

	// Auth type: "oauth" or "bot_token"
	AuthType string `json:"auth_type"`
}

// TokenInfo contains information about a Discord token
type TokenInfo struct {
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	Avatar        string    `json:"avatar,omitempty"`
	Scopes        []string  `json:"scopes"`
	ExpiresAt     time.Time `json:"expires_at"`
	TokenType     string    `json:"token_type"` // "Bot" or "Bearer"
}

// SecureToken represents a secure Discord token
type SecureToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       string    `json:"user_id,omitempty"`
	Username     string    `json:"username,omitempty"`
}

// TokenStore interface for secure token storage
type TokenStore interface {
	Store(ctx context.Context, key string, token *SecureToken) error
	Retrieve(ctx context.Context, key string) (*SecureToken, error)
	Delete(ctx context.Context, key string) error
}

// InMemoryTokenStore is an in-memory implementation of TokenStore
type InMemoryTokenStore struct {
	tokens map[string]*SecureToken
	mu     sync.RWMutex
}

// NewInMemoryTokenStore creates a new in-memory token store
func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		tokens: make(map[string]*SecureToken),
	}
}

// Store stores a token securely
func (s *InMemoryTokenStore) Store(ctx context.Context, key string, token *SecureToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[key] = token
	return nil
}

// Retrieve retrieves a token
func (s *InMemoryTokenStore) Retrieve(ctx context.Context, key string) (*SecureToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[key]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return token, nil
}

// Delete deletes a token
func (s *InMemoryTokenStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, key)
	return nil
}

// NewAuthManager creates a new Discord authentication manager
func NewAuthManager(config *AuthConfig, tokenStore TokenStore) (*AuthManager, error) {
	if config == nil {
		return nil, fmt.Errorf("auth config is required")
	}

	// Default to bot_token if auth_type not specified
	if config.AuthType == "" {
		if config.BotToken != "" {
			config.AuthType = "bot_token"
		} else if config.ClientID != "" {
			config.AuthType = "oauth"
		} else {
			return nil, fmt.Errorf("either bot_token or oauth credentials required")
		}
	}

	// Validate configuration based on auth type
	if config.AuthType == "oauth" {
		if config.ClientID == "" || config.ClientSecret == "" {
			return nil, fmt.Errorf("client_id and client_secret required for OAuth")
		}
		if config.RedirectURI == "" {
			return nil, fmt.Errorf("redirect_uri required for OAuth")
		}
	} else if config.AuthType == "bot_token" {
		if config.BotToken == "" {
			return nil, fmt.Errorf("bot_token required for token auth")
		}
	}

	// Set default scopes if not provided
	if len(config.Scopes) == 0 {
		config.Scopes = []string{
			"identify",            // Get user info
			"guilds",              // Get user's guilds
			"guilds.members.read", // Read guild members
		}
	}

	// Use in-memory store if none provided
	if tokenStore == nil {
		tokenStore = NewInMemoryTokenStore()
	}

	am := &AuthManager{
		config:     config,
		tokenStore: tokenStore,
	}

	// Set up OAuth2 config if using OAuth
	if config.AuthType == "oauth" {
		am.oauth2Config = &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURI,
			Scopes:       config.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  discordAuthURL,
				TokenURL: discordTokenURL,
			},
		}
	}

	return am, nil
}

// GetAuthURL generates the OAuth authorization URL
// State should be a cryptographically random string to prevent CSRF
func (am *AuthManager) GetAuthURL(state string) (string, error) {
	if am.config.AuthType != "oauth" {
		return "", fmt.Errorf("OAuth not enabled")
	}

	if state == "" {
		// Generate random state if not provided
		var err error
		state, err = generateRandomState()
		if err != nil {
			return "", fmt.Errorf("failed to generate state: %w", err)
		}
	}

	// Discord OAuth supports additional parameters
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("response_type", "code"),
	}

	// Add guild_id for guild-specific authorization
	if am.config.GuildID != "" {
		opts = append(opts, oauth2.SetAuthURLParam("guild_id", am.config.GuildID))
	}

	url := am.oauth2Config.AuthCodeURL(state, opts...)
	return url, nil
}

// ExchangeCode exchanges an authorization code for access tokens
func (am *AuthManager) ExchangeCode(ctx context.Context, code string) (*SecureToken, error) {
	if am.config.AuthType != "oauth" {
		return nil, fmt.Errorf("OAuth not enabled")
	}

	// Exchange code for token
	token, err := am.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info to store with token
	userInfo, err := am.getUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	secureToken := &SecureToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		UserID:       userInfo.ID,
		Username:     userInfo.Username,
	}

	// Store the token
	key := fmt.Sprintf("discord:%s", secureToken.UserID)
	if err := am.tokenStore.Store(ctx, key, secureToken); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return secureToken, nil
}

// DiscordUser represents a Discord user
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Bot           bool   `json:"bot,omitempty"`
	Verified      bool   `json:"verified,omitempty"`
	Email         string `json:"email,omitempty"`
}

// getUserInfo gets user information from Discord API
func (am *AuthManager) getUserInfo(ctx context.Context, accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", discordAPIURL+"/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user info failed with status %d: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var user DiscordUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &user, nil
}

// RefreshToken refreshes an expired token
func (am *AuthManager) RefreshToken(ctx context.Context, userID string) (*SecureToken, error) {
	if am.config.AuthType != "oauth" {
		return nil, fmt.Errorf("OAuth not enabled")
	}

	key := fmt.Sprintf("discord:%s", userID)
	oldToken, err := am.tokenStore.Retrieve(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	if oldToken.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Create token source with refresh capability
	tokenSource := am.oauth2Config.TokenSource(ctx, &oauth2.Token{
		AccessToken:  oldToken.AccessToken,
		RefreshToken: oldToken.RefreshToken,
		TokenType:    oldToken.TokenType,
		Expiry:       oldToken.ExpiresAt,
	})

	// Get new token (will automatically refresh if expired)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	secureToken := &SecureToken{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		TokenType:    newToken.TokenType,
		ExpiresAt:    newToken.Expiry,
		UserID:       oldToken.UserID,
		Username:     oldToken.Username,
	}

	// Store the refreshed token
	if err := am.tokenStore.Store(ctx, key, secureToken); err != nil {
		return nil, fmt.Errorf("failed to store refreshed token: %w", err)
	}

	return secureToken, nil
}

// GetToken returns a valid access token
// For bot_token auth, returns the bot token
// For OAuth, retrieves from store and refreshes if needed
func (am *AuthManager) GetToken(ctx context.Context) (string, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.config.AuthType == "bot_token" {
		return am.config.BotToken, nil
	}

	// OAuth mode - retrieve from store
	// Note: For OAuth, we need a user ID. This should be set in config after authentication.
	if am.config.UserToken != "" {
		// If user token is set directly, use it
		return am.config.UserToken, nil
	}

	return "", fmt.Errorf("OAuth user token not available, please authenticate first")
}

// GetTokenInfo retrieves information about the current token
func (am *AuthManager) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
	token, err := am.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	// Get current user info
	req, err := http.NewRequestWithContext(ctx, "GET", discordAPIURL+"/users/@me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header based on token type
	if am.config.AuthType == "bot_token" {
		req.Header.Set("Authorization", "Bot "+token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get user info failed with status %d: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var user DiscordUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	tokenInfo := &TokenInfo{
		UserID:        user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		Avatar:        user.Avatar,
		Scopes:        am.config.Scopes,
		TokenType:     am.config.AuthType,
	}

	return tokenInfo, nil
}

// ValidateToken validates if the current token is valid
func (am *AuthManager) ValidateToken(ctx context.Context) error {
	_, err := am.GetTokenInfo(ctx)
	return err
}

// RevokeToken revokes the current token
func (am *AuthManager) RevokeToken(ctx context.Context, userID string) error {
	if am.config.AuthType != "oauth" {
		return fmt.Errorf("token revocation only supported for OAuth")
	}

	key := fmt.Sprintf("discord:%s", userID)
	token, err := am.tokenStore.Retrieve(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to retrieve token: %w", err)
	}

	// Revoke the token via Discord API
	data := url.Values{}
	data.Set("token", token.AccessToken)
	data.Set("token_type_hint", "access_token")

	req, err := http.NewRequestWithContext(ctx, "POST", discordRevokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use Basic Auth with client credentials
	req.SetBasicAuth(am.config.ClientID, am.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("revoke failed with status %d: %s", resp.StatusCode, body)
	}

	// Delete from store
	_ = am.tokenStore.Delete(ctx, key)

	return nil
}

// GetHTTPClient returns an HTTP client with authentication configured
func (am *AuthManager) GetHTTPClient(ctx context.Context) (*http.Client, error) {
	token, err := am.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	// Create appropriate token source based on auth type
	var tokenSource oauth2.TokenSource
	if am.config.AuthType == "bot_token" {
		tokenSource = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token,
			TokenType:   "Bot",
		})
	} else {
		tokenSource = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token,
			TokenType:   "Bearer",
		})
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

// generateRandomState generates a cryptographically random state string
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
