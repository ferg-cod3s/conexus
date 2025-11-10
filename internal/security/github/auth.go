package github

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// AuthManager manages GitHub authentication
type AuthManager struct {
	config *AuthConfig
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	// Personal Access Token authentication
	Token string `json:"token,omitempty"`

	// GitHub App authentication
	AppID          string `json:"app_id,omitempty"`
	PrivateKey     string `json:"private_key,omitempty"`
	InstallationID string `json:"installation_id,omitempty"`

	// Webhook authentication
	WebhookSecret string `json:"webhook_secret,omitempty"`

	// OAuth App authentication
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`

	// Authentication type: "token", "app", "oauth"
	AuthType string `json:"auth_type"`
}

// TokenType represents the type of GitHub token
type TokenType string

const (
	// nosemgrep: go-hardcoded-credentials
	TokenTypePersonal TokenType = "personal" // Type constant, not a credential
	// nosemgrep: go-hardcoded-credentials
	TokenTypeApp TokenType = "app" // Type constant, not a credential
	// nosemgrep: go-hardcoded-credentials
	TokenTypeOAuth TokenType = "oauth" // Type constant, not a credential
	// nosemgrep: go-hardcoded-credentials
	TokenTypeWebhook TokenType = "webhook" // Type constant, not a credential
)

// TokenInfo contains information about a token
type TokenInfo struct {
	Type        TokenType  `json:"type"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Scopes      []string   `json:"scopes,omitempty"`
	Permissions []string   `json:"permissions,omitempty"`
	Repository  string     `json:"repository,omitempty"`
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *AuthConfig) (*AuthManager, error) {
	if config == nil {
		return nil, fmt.Errorf("auth config is required")
	}

	if config.AuthType == "" {
		config.AuthType = "token" // Default to personal access token
	}

	// Validate configuration based on auth type
	switch config.AuthType {
	case "token":
		if config.Token == "" {
			return nil, fmt.Errorf("token is required for personal access token authentication")
		}
	case "app":
		if config.AppID == "" || config.PrivateKey == "" {
			return nil, fmt.Errorf("app_id and private_key are required for GitHub App authentication")
		}
	case "oauth":
		if config.ClientID == "" || config.ClientSecret == "" {
			return nil, fmt.Errorf("client_id and client_secret are required for OAuth authentication")
		}
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", config.AuthType)
	}

	return &AuthManager{
		config: config,
	}, nil
}

// GetHTTPClient returns an HTTP client with authentication configured
func (am *AuthManager) GetHTTPClient(ctx context.Context) (*http.Client, error) {
	switch am.config.AuthType {
	case "token":
		return am.getPersonalTokenClient(ctx)
	case "app":
		return am.getAppClient(ctx)
	case "oauth":
		return am.getOAuthClient(ctx)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", am.config.AuthType)
	}
}

// GetTokenInfo returns information about the current token
func (am *AuthManager) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
	switch am.config.AuthType {
	case "token":
		return am.getPersonalTokenInfo(ctx)
	case "app":
		return am.getAppTokenInfo(ctx)
	case "oauth":
		return am.getOAuthTokenInfo(ctx)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", am.config.AuthType)
	}
}

// ValidateToken validates if the current token is valid
func (am *AuthManager) ValidateToken(ctx context.Context) error {
	client, err := am.GetHTTPClient(ctx)
	if err != nil {
		return err
	}

	// Make a simple API call to validate the token
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("token is invalid or expired")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	return nil
}

// RotateToken rotates the token (for apps and OAuth)
func (am *AuthManager) RotateToken(ctx context.Context) error {
	switch am.config.AuthType {
	case "app":
		return am.rotateAppToken(ctx)
	case "oauth":
		return am.rotateOAuthToken(ctx)
	default:
		return fmt.Errorf("token rotation not supported for auth type: %s", am.config.AuthType)
	}
}

// getPersonalTokenClient creates an HTTP client with personal access token
func (am *AuthManager) getPersonalTokenClient(ctx context.Context) (*http.Client, error) {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: am.config.Token,
		TokenType:   "Bearer",
	})

	return oauth2.NewClient(ctx, tokenSource), nil
}

// getAppClient creates an HTTP client with GitHub App token
func (am *AuthManager) getAppClient(ctx context.Context) (*http.Client, error) {
	// Generate JWT token for the app
	jwtToken, err := am.generateAppJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to generate app JWT: %w", err)
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: jwtToken,
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(10 * time.Minute), // JWT expires in 10 minutes
	})

	return oauth2.NewClient(ctx, tokenSource), nil
}

// getOAuthClient creates an HTTP client with OAuth token
func (am *AuthManager) getOAuthClient(ctx context.Context) (*http.Client, error) {
	// For OAuth, you would typically exchange the authorization code for an access token
	// This is a simplified implementation
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: am.config.ClientSecret, // This would be the actual access token
		TokenType:   "Bearer",
	})

	return oauth2.NewClient(ctx, tokenSource), nil
}

// generateAppJWT generates a JWT token for GitHub App authentication
func (am *AuthManager) generateAppJWT() (string, error) {
	// Parse private key
	block, _ := pem.Decode([]byte(am.config.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse private key PEM")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("private key is not RSA")
		}
	}

	// Create JWT token
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
		Issuer:    am.config.AppID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// getPersonalTokenInfo returns information about personal access token
func (am *AuthManager) getPersonalTokenInfo(ctx context.Context) (*TokenInfo, error) {
	// Personal access tokens don't have an API to get detailed info
	// We can make a basic validation call
	err := am.ValidateToken(ctx)
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Type: TokenTypePersonal,
	}, nil
}

// getAppTokenInfo returns information about app token
func (am *AuthManager) getAppTokenInfo(ctx context.Context) (*TokenInfo, error) {
	// For GitHub Apps, we would need to get the installation access token
	// This is a simplified implementation
	return &TokenInfo{
		Type:        TokenTypeApp,
		ExpiresAt:   nil,        // Would be set from actual token
		Permissions: []string{}, // Would be populated from app permissions
	}, nil
}

// getOAuthTokenInfo returns information about OAuth token
func (am *AuthManager) getOAuthTokenInfo(ctx context.Context) (*TokenInfo, error) {
	// For OAuth, we would introspect the token
	return &TokenInfo{
		Type:      TokenTypeOAuth,
		ExpiresAt: nil,        // Would be set from actual token
		Scopes:    []string{}, // Would be populated from token scopes
	}, nil
}

// rotateAppToken rotates the GitHub App installation token
func (am *AuthManager) rotateAppToken(ctx context.Context) error {
	// For GitHub Apps, this would create a new installation access token
	// This is a simplified implementation
	return fmt.Errorf("app token rotation not implemented")
}

// rotateOAuthToken rotates the OAuth token
func (am *AuthManager) rotateOAuthToken(ctx context.Context) error {
	// For OAuth, this would refresh the access token using the refresh token
	// This is a simplified implementation
	return fmt.Errorf("OAuth token rotation not implemented")
}

// VerifyWebhookSignature verifies a GitHub webhook signature
func (am *AuthManager) VerifyWebhookSignature(payload []byte, signature string) bool {
	if am.config.WebhookSecret == "" {
		return true // No secret configured
	}

	if signature == "" {
		return false
	}

	// Expected format: sha256=<hex>
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	expectedSignature := signature[7:] // Remove "sha256=" prefix

	// Generate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(am.config.WebhookSecret))
	h.Write(payload)
	actualSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(actualSignature))
}

// GenerateWebhookSecret generates a secure webhook secret
func GenerateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SecureToken represents a secure token storage
type SecureToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
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
