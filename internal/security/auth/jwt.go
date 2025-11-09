package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	audience   string
	expiry     time.Duration
}

// TokenClaims represents the JWT claims
type TokenClaims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager with RSA keys
func NewJWTManager(privateKeyPEM, publicKeyPEM, issuer, audience string, expiryMinutes int) (*JWTManager, error) {
	if privateKeyPEM == "" {
		return nil, fmt.Errorf("private key cannot be empty")
	}
	if publicKeyPEM == "" {
		return nil, fmt.Errorf("public key cannot be empty")
	}
	if issuer == "" {
		return nil, fmt.Errorf("issuer cannot be empty")
	}
	if audience == "" {
		return nil, fmt.Errorf("audience cannot be empty")
	}
	if expiryMinutes <= 0 {
		return nil, fmt.Errorf("expiry must be positive")
	}

	// Parse private key
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Parse public key
	publicKey, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		audience:   audience,
		expiry:     time.Duration(expiryMinutes) * time.Minute,
	}, nil
}

// GenerateToken generates a new JWT token for a user
func (jm *JWTManager) GenerateToken(ctx context.Context, userID, username string, roles []string) (string, error) {
	now := time.Now()

	claims := TokenClaims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Audience:  jwt.ClaimStrings{jm.audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        generateTokenID(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(jm.privateKey)
}

// ValidateToken validates a JWT token and returns the claims
func (jm *JWTManager) ValidateToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Additional validation
	if err := jm.validateClaims(claims); err != nil {
		return nil, fmt.Errorf("token claims validation failed: %w", err)
	}

	return claims, nil
}

// RefreshToken generates a new token with updated expiry for the same user
func (jm *JWTManager) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := jm.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Generate new token with same claims but updated timestamps
	return jm.GenerateToken(ctx, claims.UserID, claims.Username, claims.Roles)
}

// validateClaims performs additional validation on token claims
func (jm *JWTManager) validateClaims(claims *TokenClaims) error {
	// Validate issuer
	if claims.Issuer != jm.issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", jm.issuer, claims.Issuer)
	}

	// Validate audience
	if !containsAudience(claims.Audience, jm.audience) {
		return fmt.Errorf("invalid audience: expected %s", jm.audience)
	}

	// Validate expiry (this is also done by the JWT library, but explicit check)
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("token has expired")
	}

	// Validate not before
	if claims.NotBefore != nil && claims.NotBefore.After(time.Now()) {
		return fmt.Errorf("token not yet valid")
	}

	return nil
}

// parsePrivateKey parses an RSA private key from PEM format
func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "RSA PRIVATE KEY" && block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}

	var privateKey *rsa.PrivateKey
	var err error

	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	}

	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// parsePublicKey parses an RSA public key from PEM format
func parsePublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "RSA PUBLIC KEY" && block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}

	var publicKey *rsa.PublicKey

	if block.Type == "RSA PUBLIC KEY" {
		key, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		publicKey = key
	} else {
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		publicKey, ok = key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA public key")
		}
	}

	return publicKey, nil
}

// generateTokenID generates a random token ID
func generateTokenID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", bytes)
}

// GetExpiry returns the token expiry duration
func (jm *JWTManager) GetExpiry() time.Duration {
	return jm.expiry
}

// GetIssuer returns the token issuer
func (jm *JWTManager) GetIssuer() string {
	return jm.issuer
}

// GetAudience returns the token audience
func (jm *JWTManager) GetAudience() string {
	return jm.audience
}

// containsAudience checks if the audience list contains the expected audience
func containsAudience(audiences []string, expected string) bool {
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}
