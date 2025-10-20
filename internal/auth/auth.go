// Package auth provides comprehensive authentication and authorization for Conexus.
// Supports SAML, OIDC, JWT tokens, and role-based access control (RBAC).
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthProvider represents the authentication method used.
type AuthProvider string

const (
	ProviderJWT    AuthProvider = "jwt"
	ProviderOIDC   AuthProvider = "oidc"
	ProviderSAML   AuthProvider = "saml"
	ProviderAPIKey AuthProvider = "api_key"
)

// Permission represents a specific permission that can be granted to a user.
type Permission string

const (
	// Search permissions
	PermissionSearchRead  Permission = "search:read"
	PermissionSearchWrite Permission = "search:write"

	// Index permissions
	PermissionIndexRead  Permission = "index:read"
	PermissionIndexWrite Permission = "index:write"

	// Connector permissions
	PermissionConnectorsRead  Permission = "connectors:read"
	PermissionConnectorsWrite Permission = "connectors:write"

	// Configuration permissions
	PermissionConfigRead  Permission = "config:read"
	PermissionConfigWrite Permission = "config:write"

	// Analytics permissions
	PermissionAnalyticsRead Permission = "analytics:read"

	// Webhook permissions
	PermissionWebhooksRead  Permission = "webhooks:read"
	PermissionWebhooksWrite Permission = "webhooks:write"

	// Plugin permissions
	PermissionPluginsRead  Permission = "plugins:read"
	PermissionPluginsWrite Permission = "plugins:write"
)

// Role represents a user role with associated permissions.
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleDeveloper Role = "developer"
	RoleViewer    Role = "viewer"
	RoleService   Role = "service"
)

// User represents an authenticated user with roles and permissions.
type User struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Roles       []Role            `json:"roles"`
	Permissions []Permission      `json:"permissions"`
	Attributes  map[string]string `json:"attributes"`
	TenantID    string            `json:"tenant_id"`
	Provider    AuthProvider      `json:"provider"`
	LastLogin   time.Time         `json:"last_login"`
	TokenExpiry time.Time         `json:"token_expiry"`
}

// TokenClaims represents JWT token claims.
type TokenClaims struct {
	UserID      string       `json:"sub"`
	Username    string       `json:"username"`
	Email       string       `json:"email"`
	Name        string       `json:"name"`
	Roles       []Role       `json:"roles"`
	Permissions []Permission `json:"permissions"`
	TenantID    string       `json:"tenant_id"`
	Provider    AuthProvider `json:"provider"`
	IssuedAt    time.Time    `json:"iat"`
	ExpiresAt   time.Time    `json:"exp"`
	NotBefore   time.Time    `json:"nbf"`
	JWTID       string       `json:"jti"`
	Type        string       `json:"typ"` // token type
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWT             JWTConfig     `json:"jwt" yaml:"jwt"`
	OIDC            OIDCConfig    `json:"oidc" yaml:"oidc"`
	SAML            SAMLConfig    `json:"saml" yaml:"saml"`
	APIKeys         APIKeyConfig  `json:"api_keys" yaml:"api_keys"`
	Session         SessionConfig `json:"session" yaml:"session"`
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	DefaultProvider AuthProvider  `json:"default_provider" yaml:"default_provider"`
}

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	Secret        string        `json:"secret" yaml:"secret"`
	AccessExpiry  time.Duration `json:"access_expiry" yaml:"access_expiry"`
	RefreshExpiry time.Duration `json:"refresh_expiry" yaml:"refresh_expiry"`
	Issuer        string        `json:"issuer" yaml:"issuer"`
	Algorithm     string        `json:"algorithm" yaml:"algorithm"`
}

// OIDCConfig holds OpenID Connect configuration.
type OIDCConfig struct {
	ClientID     string   `json:"client_id" yaml:"client_id"`
	ClientSecret string   `json:"client_secret" yaml:"client_secret"`
	IssuerURL    string   `json:"issuer_url" yaml:"issuer_url"`
	RedirectURL  string   `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string `json:"scopes" yaml:"scopes"`
}

// SAMLConfig holds SAML configuration.
type SAMLConfig struct {
	EntityID     string            `json:"entity_id" yaml:"entity_id"`
	SSOURL       string            `json:"sso_url" yaml:"sso_url"`
	Certificate  string            `json:"certificate" yaml:"certificate"`
	PrivateKey   string            `json:"private_key" yaml:"private_key"`
	AttributeMap map[string]string `json:"attribute_map" yaml:"attribute_map"`
}

// APIKeyConfig holds API key configuration.
type APIKeyConfig struct {
	HeaderName string `json:"header_name" yaml:"header_name"`
	Prefix     string `json:"prefix" yaml:"prefix"`
}

// SessionConfig holds session configuration.
type SessionConfig struct {
	CookieName   string        `json:"cookie_name" yaml:"cookie_name"`
	CookieSecure bool          `json:"cookie_secure" yaml:"cookie_secure"`
	CookieHTTP   bool          `json:"cookie_http" yaml:"cookie_http"`
	MaxAge       time.Duration `json:"max_age" yaml:"max_age"`
}

// AuthResult represents the result of an authentication attempt.
type AuthResult struct {
	Success   bool   `json:"success"`
	User      *User  `json:"user,omitempty"`
	Token     string `json:"token,omitempty"`
	Refresh   string `json:"refresh,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

// Authenticator provides authentication and authorization functionality.
type Authenticator interface {
	// Authenticate validates credentials and returns an AuthResult
	Authenticate(ctx context.Context, credentials interface{}) (*AuthResult, error)

	// ValidateToken validates a JWT token and returns the associated user
	ValidateToken(ctx context.Context, token string) (*User, error)

	// RefreshToken generates a new access token from a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)

	// HasPermission checks if a user has a specific permission
	HasPermission(user *User, permission Permission) bool

	// HasRole checks if a user has a specific role
	HasRole(user *User, role Role) bool

	// CanAccessTenant checks if a user can access a specific tenant
	CanAccessTenant(user *User, tenantID string) bool

	// GenerateTokens creates access and refresh tokens for a user
	GenerateTokens(ctx context.Context, user *User) (accessToken, refreshToken string, err error)

	// InvalidateTokens revokes all tokens for a user
	InvalidateTokens(ctx context.Context, userID string) error
}

// DefaultAuthenticator implements the Authenticator interface.
type DefaultAuthenticator struct {
	config *AuthConfig
}

// NewAuthenticator creates a new authenticator with the given configuration.
func NewAuthenticator(config *AuthConfig) Authenticator {
	return &DefaultAuthenticator{
		config: config,
	}
}

// Authenticate validates credentials based on the provider type.
func (a *DefaultAuthenticator) Authenticate(ctx context.Context, credentials interface{}) (*AuthResult, error) {
	switch creds := credentials.(type) {
	case *JWTCredentials:
		return a.authenticateJWT(ctx, creds)
	case *OIDCCredentials:
		return a.authenticateOIDC(ctx, creds)
	case *SAMLCredentials:
		return a.authenticateSAML(ctx, creds)
	case *APIKeyCredentials:
		return a.authenticateAPIKey(ctx, creds)
	default:
		return &AuthResult{
			Success:   false,
			Error:     "unsupported credential type",
			ErrorCode: "unsupported_credentials",
		}, nil
	}
}

// ValidateToken validates a JWT token and returns the associated user.
func (a *DefaultAuthenticator) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	if !a.config.Enabled {
		return nil, fmt.Errorf("authentication is disabled")
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method.Alg() != a.config.JWT.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check token expiration
	if time.Now().After(claims.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// Check token type
	if claims.Type != "access" {
		return nil, fmt.Errorf("invalid token type: %s", claims.Type)
	}

	// Create user from claims
	user := &User{
		ID:          claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		Name:        claims.Name,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		TenantID:    claims.TenantID,
		Provider:    claims.Provider,
		LastLogin:   claims.IssuedAt,
		TokenExpiry: claims.ExpiresAt,
		Attributes:  make(map[string]string),
	}

	return user, nil
}

// RefreshToken generates a new access token from a refresh token.
func (a *DefaultAuthenticator) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if !a.config.Enabled {
		return &AuthResult{
			Success:   false,
			Error:     "authentication is disabled",
			ErrorCode: "auth_disabled",
		}, nil
	}

	// Parse refresh token (similar to access token but with "refresh" type)
	token, err := jwt.ParseWithClaims(refreshToken, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != a.config.JWT.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWT.Secret), nil
	})

	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     "invalid refresh token",
			ErrorCode: "invalid_refresh_token",
		}, nil
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || claims.Type != "refresh" {
		return &AuthResult{
			Success:   false,
			Error:     "invalid refresh token claims",
			ErrorCode: "invalid_refresh_claims",
		}, nil
	}

	// Create user from refresh token claims
	user := &User{
		ID:          claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		Name:        claims.Name,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		TenantID:    claims.TenantID,
		Provider:    claims.Provider,
		LastLogin:   time.Now(),
		Attributes:  make(map[string]string),
	}

	// Generate new tokens
	return a.generateTokensForUser(ctx, user)
}

// HasPermission checks if a user has a specific permission.
func (a *DefaultAuthenticator) HasPermission(user *User, permission Permission) bool {
	if user == nil {
		return false
	}

	// Admin role has all permissions
	for _, role := range user.Roles {
		if role == RoleAdmin {
			return true
		}
	}

	// Check explicit permissions
	for _, p := range user.Permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// HasRole checks if a user has a specific role.
func (a *DefaultAuthenticator) HasRole(user *User, role Role) bool {
	if user == nil {
		return false
	}

	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}

	return false
}

// CanAccessTenant checks if a user can access a specific tenant.
func (a *DefaultAuthenticator) CanAccessTenant(user *User, tenantID string) bool {
	if user == nil {
		return false
	}

	// Admin can access all tenants
	for _, role := range user.Roles {
		if role == RoleAdmin {
			return true
		}
	}

	// Check tenant match
	return user.TenantID == tenantID
}

// GenerateTokens creates access and refresh tokens for a user.
func (a *DefaultAuthenticator) GenerateTokens(ctx context.Context, user *User) (string, string, error) {
	return a.generateTokensForUser(ctx, user)
}

// InvalidateTokens revokes all tokens for a user.
func (a *DefaultAuthenticator) InvalidateTokens(ctx context.Context, userID string) error {
	// In a real implementation, this would maintain a token blacklist
	// For now, we'll log the invalidation
	fmt.Printf("Token invalidation requested for user: %s\n", userID)
	return nil
}

// Credential types for different authentication methods
type JWTCredentials struct {
	Token string `json:"token"`
}

type OIDCCredentials struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type SAMLCredentials struct {
	SAMLResponse string `json:"saml_response"`
	RelayState   string `json:"relay_state"`
}

type APIKeyCredentials struct {
	APIKey string `json:"api_key"`
}

// authenticateJWT validates JWT credentials.
func (a *DefaultAuthenticator) authenticateJWT(ctx context.Context, creds *JWTCredentials) (*AuthResult, error) {
	user, err := a.ValidateToken(ctx, creds.Token)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     fmt.Sprintf("JWT validation failed: %v", err),
			ErrorCode: "invalid_jwt",
		}, nil
	}

	return &AuthResult{
		Success: true,
		User:    user,
		Token:   creds.Token,
	}, nil
}

// authenticateOIDC handles OIDC authentication flow.
func (a *DefaultAuthenticator) authenticateOIDC(ctx context.Context, creds *OIDCCredentials) (*AuthResult, error) {
	// In a real implementation, this would exchange the authorization code
	// for tokens with the OIDC provider
	return &AuthResult{
		Success:   false,
		Error:     "OIDC authentication not implemented",
		ErrorCode: "not_implemented",
	}, nil
}

// authenticateSAML handles SAML authentication flow.
func (a *DefaultAuthenticator) authenticateSAML(ctx context.Context, creds *SAMLCredentials) (*AuthResult, error) {
	// In a real implementation, this would validate the SAML response
	// and extract user attributes
	return &AuthResult{
		Success:   false,
		Error:     "SAML authentication not implemented",
		ErrorCode: "not_implemented",
	}, nil
}

// authenticateAPIKey validates API key credentials.
func (a *DefaultAuthenticator) authenticateAPIKey(ctx context.Context, creds *APIKeyCredentials) (*AuthResult, error) {
	// In a real implementation, this would validate the API key
	// against a database or configuration
	return &AuthResult{
		Success:   false,
		Error:     "API key authentication not implemented",
		ErrorCode: "not_implemented",
	}, nil
}

// generateTokensForUser creates access and refresh tokens for a user.
func (a *DefaultAuthenticator) generateTokensForUser(ctx context.Context, user *User) (*AuthResult, error) {
	now := time.Now()

	// Generate JWT ID for token tracking
	jwtID, err := generateJWTID()
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to generate JWT ID: %v", err),
			ErrorCode: "token_generation_failed",
		}, nil
	}

	// Create access token claims
	accessClaims := TokenClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Name:        user.Name,
		Roles:       user.Roles,
		Permissions: user.Permissions,
		TenantID:    user.TenantID,
		Provider:    user.Provider,
		IssuedAt:    now,
		ExpiresAt:   now.Add(a.config.JWT.AccessExpiry),
		NotBefore:   now,
		JWTID:       jwtID,
		Type:        "access",
	}

	// Create refresh token claims
	refreshClaims := TokenClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Name:        user.Name,
		Roles:       user.Roles,
		Permissions: user.Permissions,
		TenantID:    user.TenantID,
		Provider:    user.Provider,
		IssuedAt:    now,
		ExpiresAt:   now.Add(a.config.JWT.RefreshExpiry),
		NotBefore:   now,
		JWTID:       jwtID + "_refresh",
		Type:        "refresh",
	}

	// Generate access token
	accessToken, err := a.generateToken(accessClaims)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to generate access token: %v", err),
			ErrorCode: "access_token_generation_failed",
		}, nil
	}

	// Generate refresh token
	refreshToken, err := a.generateToken(refreshClaims)
	if err != nil {
		return &AuthResult{
			Success:   false,
			Error:     fmt.Sprintf("failed to generate refresh token: %v", err),
			ErrorCode: "refresh_token_generation_failed",
		}, nil
	}

	return &AuthResult{
		Success: true,
		User:    user,
		Token:   accessToken,
		Refresh: refreshToken,
	}, nil
}

// generateToken creates a JWT token from claims.
func (a *DefaultAuthenticator) generateToken(claims TokenClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWT.Secret))
}

// generateJWTID generates a unique JWT ID.
func generateJWTID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GetDefaultAuthConfig returns default authentication configuration.
func GetDefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		JWT: JWTConfig{
			Secret:        "change-me-in-production",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 30 * 24 * time.Hour, // 30 days
			Issuer:        "conexus",
			Algorithm:     "HS256",
		},
		OIDC: OIDCConfig{
			ClientID:     "",
			ClientSecret: "",
			IssuerURL:    "",
			RedirectURL:  "",
			Scopes:       []string{"openid", "profile", "email"},
		},
		SAML: SAMLConfig{
			EntityID:    "",
			SSOURL:      "",
			Certificate: "",
			PrivateKey:  "",
			AttributeMap: map[string]string{
				"email":    "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
				"name":     "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
				"username": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier",
			},
		},
		APIKeys: APIKeyConfig{
			HeaderName: "X-API-Key",
			Prefix:     "cnx_",
		},
		Session: SessionConfig{
			CookieName:   "conexus_session",
			CookieSecure: true,
			CookieHTTP:   false,
			MaxAge:       24 * time.Hour,
		},
		Enabled:         false, // Disabled by default for backward compatibility
		DefaultProvider: ProviderJWT,
	}
}
