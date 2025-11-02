package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/security/auth"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// Middleware returns an HTTP middleware function that validates JWT tokens
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for certain endpoints
		if am.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		tokenString, err := am.extractToken(r)
		if err != nil {
			am.unauthorized(w, "Invalid or missing authorization token")
			return
		}

		// Validate token
		claims, err := am.jwtManager.ValidateToken(r.Context(), tokenString)
		if err != nil {
			am.unauthorized(w, "Invalid token")
			return
		}

		// Add claims to request context
		ctx := am.addClaimsToContext(r.Context(), claims)
		r = r.WithContext(ctx)

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

// shouldSkipAuth determines if authentication should be skipped for the given path
func (am *AuthMiddleware) shouldSkipAuth(path string) bool {
	// Skip authentication for health check and webhook endpoints
	skipPaths := []string{
		"/health",
		"/webhooks/github",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

// extractToken extracts the JWT token from the Authorization header
func (am *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie
	}

	// Check if it's a Bearer token
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", http.ErrNoCookie
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", http.ErrNoCookie
	}

	return token, nil
}

// unauthorized sends an unauthorized response
func (am *AuthMiddleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// addClaimsToContext adds JWT claims to the request context
func (am *AuthMiddleware) addClaimsToContext(ctx context.Context, claims *auth.TokenClaims) context.Context {
	// Add user information to context
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "username", claims.Username)
	ctx = context.WithValue(ctx, "roles", claims.Roles)
	ctx = context.WithValue(ctx, "token_id", claims.ID)

	return ctx
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetUsername extracts the username from the request context
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value("username").(string)
	return username, ok
}

// GetRoles extracts the user roles from the request context
func GetRoles(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value("roles").([]string)
	return roles, ok
}

// GetTokenID extracts the token ID from the request context
func GetTokenID(ctx context.Context) (string, bool) {
	tokenID, ok := ctx.Value("token_id").(string)
	return tokenID, ok
}

// RequireRole is a helper function to check if the user has a required role
func RequireRole(ctx context.Context, requiredRole string) bool {
	roles, ok := GetRoles(ctx)
	if !ok {
		return false
	}

	for _, role := range roles {
		if role == requiredRole {
			return true
		}
	}

	return false
}

// RequireAnyRole is a helper function to check if the user has any of the required roles
func RequireAnyRole(ctx context.Context, requiredRoles []string) bool {
	roles, ok := GetRoles(ctx)
	if !ok {
		return false
	}

	for _, userRole := range roles {
		for _, requiredRole := range requiredRoles {
			if userRole == requiredRole {
				return true
			}
		}
	}

	return false
}
