package jira

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// Jira OAuth 2.0 (3LO) endpoints for Jira Cloud
// Reference: https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/
const (
	jiraAuthURL  = "https://auth.atlassian.com/authorize"
	jiraTokenURL = "https://auth.atlassian.com/oauth/token"
	jiraAPIURL   = "https://api.atlassian.com"
)

// AuthManager manages Jira OAuth authentication
type AuthManager struct {
	config       *AuthConfig
	oauth2Config *oauth2.Config
	tokenStore   TokenStore
	mu           sync.RWMutex
}

// AuthConfig contains Jira OAuth configuration
type AuthConfig struct {
	// OAuth App credentials
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`

	// API token (for backward compatibility with non-OAuth)
	// Format: email + api_token for Basic Auth
	Email    string `json:"email,omitempty"`
	APIToken string `json:"api_token,omitempty"`

	// Base URL (for Jira Cloud: https://your-domain.atlassian.net)
	BaseURL string `json:"base_url"`

	// Cloud ID (obtained after OAuth, identifies the Jira site)
	CloudID string `json:"cloud_id,omitempty"`

	// OAuth scopes to request
	Scopes []string `json:"scopes,omitempty"`

	// Auth type: "oauth" or "api_token"
	AuthType string `json:"auth_type"`
}

// TokenInfo contains information about a Jira token
type TokenInfo struct {
	CloudID   string    `json:"cloud_id"`
	SiteName  string    `json:"site_name"`
	SiteURL   string    `json:"site_url"`
	Scopes    []string  `json:"scopes"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id,omitempty"`
}

// SecureToken represents a secure Jira token
type SecureToken struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	Scope        string     `json:"scope"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CloudID      string     `json:"cloud_id"`
	Resources    []Resource `json:"resources,omitempty"`
}

// Resource represents an accessible Jira resource (site)
type Resource struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Scopes    []string `json:"scopes"`
	AvatarURL string   `json:"avatarUrl,omitempty"`
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

// NewAuthManager creates a new Jira authentication manager
func NewAuthManager(config *AuthConfig, tokenStore TokenStore) (*AuthManager, error) {
	if config == nil {
		return nil, fmt.Errorf("auth config is required")
	}

	if config.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}

	// Default to api_token if auth_type not specified
	if config.AuthType == "" {
		if config.Email != "" && config.APIToken != "" {
			config.AuthType = "api_token"
		} else if config.ClientID != "" {
			config.AuthType = "oauth"
		} else {
			return nil, fmt.Errorf("either api_token or oauth credentials required")
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
	} else if config.AuthType == "api_token" {
		if config.Email == "" || config.APIToken == "" {
			return nil, fmt.Errorf("email and api_token required for API token auth")
		}
	}

	// Set default scopes if not provided
	if len(config.Scopes) == 0 {
		config.Scopes = []string{
			"read:jira-work", // Read Jira issues, projects, etc.
			"read:jira-user", // Read user information
			"offline_access", // Get refresh token
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
				AuthURL:  jiraAuthURL,
				TokenURL: jiraTokenURL,
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

	// Jira OAuth requires specific parameters
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("audience", "api.atlassian.com"),
		oauth2.SetAuthURLParam("prompt", "consent"),
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

	// Get accessible resources (Jira sites)
	resources, err := am.getAccessibleResources(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessible resources: %w", err)
	}

	secureToken := &SecureToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
		Resources:    resources,
	}

	// If we have resources, use the first one as default cloud ID
	if len(resources) > 0 {
		secureToken.CloudID = resources[0].ID
	}

	// Store the token
	key := fmt.Sprintf("jira:%s", secureToken.CloudID)
	if err := am.tokenStore.Store(ctx, key, secureToken); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return secureToken, nil
}

// getAccessibleResources gets the list of Jira sites the token can access
func (am *AuthManager) getAccessibleResources(ctx context.Context, accessToken string) ([]Resource, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", jiraAPIURL+"/oauth/token/accessible-resources", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get resources failed with status %d: %s", resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resources []Resource
	if err := json.Unmarshal(body, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse resources: %w", err)
	}

	return resources, nil
}

// RefreshToken refreshes an expired token
func (am *AuthManager) RefreshToken(ctx context.Context) (*SecureToken, error) {
	if am.config.AuthType != "oauth" {
		return nil, fmt.Errorf("OAuth not enabled")
	}

	if am.config.CloudID == "" {
		return nil, fmt.Errorf("cloud_id required for token refresh")
	}

	key := fmt.Sprintf("jira:%s", am.config.CloudID)
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
		CloudID:      oldToken.CloudID,
		Resources:    oldToken.Resources,
	}

	// Store the refreshed token
	if err := am.tokenStore.Store(ctx, key, secureToken); err != nil {
		return nil, fmt.Errorf("failed to store refreshed token: %w", err)
	}

	return secureToken, nil
}

// GetToken returns a valid access token
// For api_token auth, returns the Basic Auth header value
// For OAuth, retrieves from store and refreshes if needed
func (am *AuthManager) GetToken(ctx context.Context) (string, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.config.AuthType == "api_token" {
		// Return Basic Auth token (base64 encoded email:api_token)
		auth := base64.StdEncoding.EncodeToString([]byte(am.config.Email + ":" + am.config.APIToken))
		return "Basic " + auth, nil
	}

	// OAuth mode - retrieve from store
	if am.config.CloudID == "" {
		return "", fmt.Errorf("cloud_id required for OAuth token retrieval")
	}

	key := fmt.Sprintf("jira:%s", am.config.CloudID)
	token, err := am.tokenStore.Retrieve(ctx, key)
	if err != nil {
		// Token not found or expired - try to refresh
		am.mu.RUnlock()
		refreshed, refreshErr := am.RefreshToken(ctx)
		am.mu.RLock()
		if refreshErr != nil {
			return "", fmt.Errorf("token expired and refresh failed: %w", refreshErr)
		}
		return "Bearer " + refreshed.AccessToken, nil
	}

	// Check if token needs refresh (refresh 5 minutes before expiry)
	if time.Now().After(token.ExpiresAt.Add(-5 * time.Minute)) {
		am.mu.RUnlock()
		refreshed, err := am.RefreshToken(ctx)
		am.mu.RLock()
		if err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		return "Bearer " + refreshed.AccessToken, nil
	}

	return "Bearer " + token.AccessToken, nil
}

// GetTokenInfo retrieves information about the current token
func (am *AuthManager) GetTokenInfo(ctx context.Context) (*TokenInfo, error) {
	if am.config.AuthType == "api_token" {
		// API token doesn't have detailed info, return basic info
		return &TokenInfo{
			SiteName: am.config.BaseURL,
			SiteURL:  am.config.BaseURL,
		}, nil
	}

	// OAuth mode - get from stored token
	if am.config.CloudID == "" {
		return nil, fmt.Errorf("cloud_id required for OAuth")
	}

	key := fmt.Sprintf("jira:%s", am.config.CloudID)
	token, err := am.tokenStore.Retrieve(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}

	// Find the resource matching our cloud ID
	var resource *Resource
	for _, r := range token.Resources {
		if r.ID == am.config.CloudID {
			resource = &r
			break
		}
	}

	if resource == nil {
		return nil, fmt.Errorf("cloud_id not found in accessible resources")
	}

	return &TokenInfo{
		CloudID:   token.CloudID,
		SiteName:  resource.Name,
		SiteURL:   resource.URL,
		Scopes:    resource.Scopes,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

// ValidateToken validates if the current token is valid
func (am *AuthManager) ValidateToken(ctx context.Context) error {
	token, err := am.GetToken(ctx)
	if err != nil {
		return err
	}

	// Make a simple API call to validate the token
	// Use /rest/api/3/myself to check authentication
	url := am.config.BaseURL + "/rest/api/3/myself"
	if am.config.AuthType == "oauth" && am.config.CloudID != "" {
		// For OAuth, construct URL using cloud ID
		key := fmt.Sprintf("jira:%s", am.config.CloudID)
		storedToken, err := am.tokenStore.Retrieve(ctx, key)
		if err != nil {
			return err
		}
		// Find resource URL
		for _, r := range storedToken.Resources {
			if r.ID == am.config.CloudID {
				url = r.URL + "/rest/api/3/myself"
				break
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
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

// GetHTTPClient returns an HTTP client with authentication configured
func (am *AuthManager) GetHTTPClient(ctx context.Context) (*http.Client, error) {
	token, err := am.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	// Create client with custom transport that adds auth header
	transport := &authTransport{
		authHeader: token,
		base:       http.DefaultTransport,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

// authTransport adds authentication header to requests
type authTransport struct {
	authHeader string
	base       http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", t.authHeader)
	return t.base.RoundTrip(req)
}

// generateRandomState generates a cryptographically random state string
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
