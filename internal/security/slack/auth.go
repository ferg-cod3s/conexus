package slack

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

// Slack OAuth 2.0 endpoints
// Reference: https://api.slack.com/authentication/oauth-v2
const (
	slackAuthURL   = "https://slack.com/oauth/v2/authorize"
	slackTokenURL  = "https://slack.com/api/oauth.v2.access"
	slackRevokeURL = "https://slack.com/api/auth.revoke"
)

// AuthManager manages Slack OAuth authentication
type AuthManager struct {
	config       *AuthConfig
	oauth2Config *oauth2.Config
	tokenStore   TokenStore
	mu           sync.RWMutex
}

// AuthConfig contains Slack OAuth configuration
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

	// Team/workspace ID (optional, for additional validation)
	TeamID string `json:"team_id,omitempty"`

	// Auth type: "oauth" or "bot_token"
	AuthType string `json:"auth_type"`
}

// TokenInfo contains information about a Slack token
type TokenInfo struct {
	TeamID    string    `json:"team_id"`
	TeamName  string    `json:"team_name"`
	UserID    string    `json:"user_id"`
	Scopes    []string  `json:"scopes"`
	BotID     string    `json:"bot_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	TokenType string    `json:"token_type"` // "bot" or "user"
}

// SecureToken represents a secure Slack token
type SecureToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	TeamID       string    `json:"team_id"`
	TeamName     string    `json:"team_name"`
	BotUserID    string    `json:"bot_user_id,omitempty"`
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

	// Check if token is expired (for future when Slack supports expiring tokens)
	if !token.ExpiresAt.IsZero() && time.Now().After(token.ExpiresAt) {
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

// NewAuthManager creates a new Slack authentication manager
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
			"channels:history",
			"channels:read",
			"groups:history",
			"groups:read",
			"im:history",
			"mpim:history",
			"search:read",
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
				AuthURL:  slackAuthURL,
				TokenURL: slackTokenURL,
			},
		}
	}

	return am, nil
}

// GetAuthURL generates the OAuth authorization URL
// State should be a cryptographically random string to prevent CSRF
func (am *AuthManager) GetAuthURL(state string, userScopes []string) (string, error) {
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

	// Slack OAuth v2 requires user_scope parameter for user tokens
	// and scope parameter for bot tokens
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("user_scope", strings.Join(userScopes, ",")),
	}

	// Add team_id if specified (to pre-select workspace)
	if am.config.TeamID != "" {
		opts = append(opts, oauth2.SetAuthURLParam("team", am.config.TeamID))
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

	// Slack's OAuth response includes extra fields in token.Extra()
	// Extract team_id, team_name, bot_user_id, etc.
	secureToken := &SecureToken{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresAt:   token.Expiry,
	}

	// Parse additional Slack-specific fields from response
	if extra := token.Extra("team"); extra != nil {
		if team, ok := extra.(map[string]interface{}); ok {
			if id, ok := team["id"].(string); ok {
				secureToken.TeamID = id
			}
			if name, ok := team["name"].(string); ok {
				secureToken.TeamName = name
			}
		}
	}

	if botUserID := token.Extra("bot_user_id"); botUserID != nil {
		if id, ok := botUserID.(string); ok {
			secureToken.BotUserID = id
		}
	}

	// Store the token
	key := fmt.Sprintf("slack:%s", secureToken.TeamID)
	if err := am.tokenStore.Store(ctx, key, secureToken); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return secureToken, nil
}

// GetToken returns a valid access token
// For bot_token auth, returns the configured token
// For OAuth, retrieves from store and refreshes if needed
func (am *AuthManager) GetToken(ctx context.Context) (string, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.config.AuthType == "bot_token" {
		return am.config.BotToken, nil
	}

	// OAuth mode - retrieve from store
	if am.config.TeamID == "" {
		return "", fmt.Errorf("team_id required for OAuth token retrieval")
	}

	key := fmt.Sprintf("slack:%s", am.config.TeamID)
	token, err := am.tokenStore.Retrieve(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve token: %w", err)
	}

	// Check if token needs refresh (Slack tokens currently don't expire)
	// But keeping this logic for future compatibility
	if !token.ExpiresAt.IsZero() && time.Now().After(token.ExpiresAt.Add(-5*time.Minute)) {
		// Token expired or about to expire - refresh not currently supported by Slack
		return "", fmt.Errorf("token expired")
	}

	return token.AccessToken, nil
}

// GetTokenInfo retrieves information about the current token
func (am *AuthManager) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
	token, err := am.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	// Call Slack auth.test API to get token info
	// https://api.slack.com/methods/auth.test
	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call auth.test: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		OK     bool   `json:"ok"`
		Error  string `json:"error,omitempty"`
		URL    string `json:"url"`
		Team   string `json:"team"`
		User   string `json:"user"`
		TeamID string `json:"team_id"`
		UserID string `json:"user_id"`
		BotID  string `json:"bot_id,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.OK {
		return nil, fmt.Errorf("auth.test failed: %s", result.Error)
	}

	return &TokenInfo{
		TeamID:   result.TeamID,
		TeamName: result.Team,
		UserID:   result.UserID,
		BotID:    result.BotID,
		TokenType: func() string {
			if result.BotID != "" {
				return "bot"
			}
			return "user"
		}(),
	}, nil
}

// ValidateToken validates if the current token is valid
func (am *AuthManager) ValidateToken(ctx context.Context) error {
	_, err := am.GetTokenInfo(ctx)
	return err
}

// RevokeToken revokes the current token
func (am *AuthManager) RevokeToken(ctx context.Context) error {
	token, err := am.GetToken(ctx)
	if err != nil {
		return err
	}

	// Call Slack auth.revoke API
	// https://api.slack.com/methods/auth.revoke
	data := url.Values{}
	data.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, "POST", slackRevokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("revoke failed: %s", result.Error)
	}

	// Delete from store
	if am.config.TeamID != "" {
		key := fmt.Sprintf("slack:%s", am.config.TeamID)
		_ = am.tokenStore.Delete(ctx, key)
	}

	return nil
}

// GetHTTPClient returns an HTTP client with authentication configured
func (am *AuthManager) GetHTTPClient(ctx context.Context) (*http.Client, error) {
	token, err := am.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
		TokenType:   "Bearer",
	})

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
