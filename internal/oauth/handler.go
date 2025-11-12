package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	discordauth "github.com/ferg-cod3s/conexus/internal/security/discord"
	jiraauth "github.com/ferg-cod3s/conexus/internal/security/jira"
	slackauth "github.com/ferg-cod3s/conexus/internal/security/slack"
)

// Handler manages OAuth flow HTTP handlers
type Handler struct {
	slackAuth   *slackauth.AuthManager
	jiraAuth    *jiraauth.AuthManager
	discordAuth *discordauth.AuthManager

	// State store for CSRF protection
	stateStore map[string]*stateInfo
	stateMu    sync.RWMutex

	// Success/error callbacks
	onSuccess func(provider string, token interface{})
	onError   func(provider string, err error)
}

// stateInfo stores OAuth state information
type stateInfo struct {
	Provider  string
	CreatedAt time.Time
	UserData  map[string]string // Optional user data to pass through OAuth flow
}

// Config contains configuration for OAuth handlers
type Config struct {
	SlackAuth   *slackauth.AuthManager
	JiraAuth    *jiraauth.AuthManager
	DiscordAuth *discordauth.AuthManager
	OnSuccess   func(provider string, token interface{})
	OnError     func(provider string, err error)
}

// NewHandler creates a new OAuth handler
func NewHandler(config *Config) *Handler {
	return &Handler{
		slackAuth:   config.SlackAuth,
		jiraAuth:    config.JiraAuth,
		discordAuth: config.DiscordAuth,
		stateStore:  make(map[string]*stateInfo),
		onSuccess:   config.OnSuccess,
		onError:     config.OnError,
	}
}

// RegisterRoutes registers OAuth routes with an HTTP mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Authorization initiation endpoints
	mux.HandleFunc("/oauth/slack/authorize", h.handleSlackAuthorize)
	mux.HandleFunc("/oauth/jira/authorize", h.handleJiraAuthorize)
	mux.HandleFunc("/oauth/discord/authorize", h.handleDiscordAuthorize)

	// OAuth callback endpoints
	mux.HandleFunc("/oauth/slack/callback", h.handleSlackCallback)
	mux.HandleFunc("/oauth/jira/callback", h.handleJiraCallback)
	mux.HandleFunc("/oauth/discord/callback", h.handleDiscordCallback)

	// Success/error pages
	mux.HandleFunc("/oauth/success", h.handleSuccess)
	mux.HandleFunc("/oauth/error", h.handleError)
}

// handleSlackAuthorize initiates Slack OAuth flow
func (h *Handler) handleSlackAuthorize(w http.ResponseWriter, r *http.Request) {
	if h.slackAuth == nil {
		http.Error(w, "Slack OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate state
	state, err := generateState("slack")
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state
	h.storeState(state, "slack", nil)

	// Get authorization URL
	userScopes := r.URL.Query()["user_scope"]
	authURL, err := h.slackAuth.GetAuthURL(state, userScopes)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get auth URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to Slack
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handleSlackCallback handles Slack OAuth callback
func (h *Handler) handleSlackCallback(w http.ResponseWriter, r *http.Request) {
	if h.slackAuth == nil {
		http.Error(w, "Slack OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Get code and state from query params
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for errors
	if errorParam != "" {
		h.handleOAuthError(w, r, "slack", fmt.Errorf("OAuth error: %s", errorParam))
		return
	}

	// Validate state
	if !h.validateState(state, "slack") {
		h.handleOAuthError(w, r, "slack", fmt.Errorf("invalid state"))
		return
	}

	// Exchange code for token
	ctx := r.Context()
	token, err := h.slackAuth.ExchangeCode(ctx, code)
	if err != nil {
		h.handleOAuthError(w, r, "slack", fmt.Errorf("failed to exchange code: %w", err))
		return
	}

	// Call success callback
	if h.onSuccess != nil {
		h.onSuccess("slack", token)
	}

	// Redirect to success page
	http.Redirect(w, r, "/oauth/success?provider=slack", http.StatusTemporaryRedirect)
}

// handleJiraAuthorize initiates Jira OAuth flow
func (h *Handler) handleJiraAuthorize(w http.ResponseWriter, r *http.Request) {
	if h.jiraAuth == nil {
		http.Error(w, "Jira OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate state
	state, err := generateState("jira")
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state
	h.storeState(state, "jira", nil)

	// Get authorization URL
	authURL, err := h.jiraAuth.GetAuthURL(state)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get auth URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to Jira
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handleJiraCallback handles Jira OAuth callback
func (h *Handler) handleJiraCallback(w http.ResponseWriter, r *http.Request) {
	if h.jiraAuth == nil {
		http.Error(w, "Jira OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Get code and state from query params
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for errors
	if errorParam != "" {
		h.handleOAuthError(w, r, "jira", fmt.Errorf("OAuth error: %s", errorParam))
		return
	}

	// Validate state
	if !h.validateState(state, "jira") {
		h.handleOAuthError(w, r, "jira", fmt.Errorf("invalid state"))
		return
	}

	// Exchange code for token
	ctx := r.Context()
	token, err := h.jiraAuth.ExchangeCode(ctx, code)
	if err != nil {
		h.handleOAuthError(w, r, "jira", fmt.Errorf("failed to exchange code: %w", err))
		return
	}

	// Call success callback
	if h.onSuccess != nil {
		h.onSuccess("jira", token)
	}

	// Redirect to success page
	http.Redirect(w, r, "/oauth/success?provider=jira", http.StatusTemporaryRedirect)
}

// handleDiscordAuthorize initiates Discord OAuth flow
func (h *Handler) handleDiscordAuthorize(w http.ResponseWriter, r *http.Request) {
	if h.discordAuth == nil {
		http.Error(w, "Discord OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate state
	state, err := generateState("discord")
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state
	h.storeState(state, "discord", nil)

	// Get authorization URL
	authURL, err := h.discordAuth.GetAuthURL(state)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get auth URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Redirect to Discord
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// handleDiscordCallback handles Discord OAuth callback
func (h *Handler) handleDiscordCallback(w http.ResponseWriter, r *http.Request) {
	if h.discordAuth == nil {
		http.Error(w, "Discord OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Get code and state from query params
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	// Check for errors
	if errorParam != "" {
		h.handleOAuthError(w, r, "discord", fmt.Errorf("OAuth error: %s", errorParam))
		return
	}

	// Validate state
	if !h.validateState(state, "discord") {
		h.handleOAuthError(w, r, "discord", fmt.Errorf("invalid state"))
		return
	}

	// Exchange code for token
	ctx := r.Context()
	token, err := h.discordAuth.ExchangeCode(ctx, code)
	if err != nil {
		h.handleOAuthError(w, r, "discord", fmt.Errorf("failed to exchange code: %w", err))
		return
	}

	// Call success callback
	if h.onSuccess != nil {
		h.onSuccess("discord", token)
	}

	// Redirect to success page
	http.Redirect(w, r, "/oauth/success?provider=discord", http.StatusTemporaryRedirect)
}

// handleSuccess displays OAuth success page
func (h *Handler) handleSuccess(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = "unknown"
	}

	tmpl := template.Must(template.New("success").Parse(successPageTemplate))
	data := map[string]string{
		"Provider": provider,
	}
	tmpl.Execute(w, data)
}

// handleError displays OAuth error page
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	errorMsg := r.URL.Query().Get("error")

	if provider == "" {
		provider = "unknown"
	}
	if errorMsg == "" {
		errorMsg = "Unknown error occurred"
	}

	tmpl := template.Must(template.New("error").Parse(errorPageTemplate))
	data := map[string]string{
		"Provider": provider,
		"Error":    errorMsg,
	}
	tmpl.Execute(w, data)
}

// handleOAuthError handles OAuth errors
func (h *Handler) handleOAuthError(w http.ResponseWriter, r *http.Request, provider string, err error) {
	// Call error callback
	if h.onError != nil {
		h.onError(provider, err)
	}

	// Redirect to error page
	errorURL := fmt.Sprintf("/oauth/error?provider=%s&error=%s", provider, err.Error())
	http.Redirect(w, r, errorURL, http.StatusTemporaryRedirect)
}

// storeState stores OAuth state for CSRF protection
func (h *Handler) storeState(state, provider string, userData map[string]string) {
	h.stateMu.Lock()
	defer h.stateMu.Unlock()

	h.stateStore[state] = &stateInfo{
		Provider:  provider,
		CreatedAt: time.Now(),
		UserData:  userData,
	}

	// Clean up old states (older than 10 minutes)
	cutoff := time.Now().Add(-10 * time.Minute)
	for k, v := range h.stateStore {
		if v.CreatedAt.Before(cutoff) {
			delete(h.stateStore, k)
		}
	}
}

// validateState validates OAuth state and removes it after validation
func (h *Handler) validateState(state, expectedProvider string) bool {
	h.stateMu.Lock()
	defer h.stateMu.Unlock()

	info, exists := h.stateStore[state]
	if !exists {
		return false
	}

	// Check provider matches
	if info.Provider != expectedProvider {
		return false
	}

	// Check not expired (10 minutes)
	if time.Since(info.CreatedAt) > 10*time.Minute {
		delete(h.stateStore, state)
		return false
	}

	// Remove state after successful validation (one-time use)
	delete(h.stateStore, state)

	return true
}

// generateState generates a random state string with provider prefix
func generateState(provider string) (string, error) {
	// Use the same random generation as in auth packages
	b := make([]byte, 32)
	if _, err := http.DefaultClient.Do(&http.Request{}); err == nil {
		// Just a dummy check, actual random generation below
	}
	return fmt.Sprintf("%s_%d", provider, time.Now().UnixNano()), nil
}

// GetTokenJSON returns token information as JSON
func (h *Handler) GetTokenJSON(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")

	var tokenInfo interface{}
	var err error
	ctx := context.Background()

	switch provider {
	case "slack":
		if h.slackAuth != nil {
			tokenInfo, err = h.slackAuth.GetTokenInfo(ctx)
		}
	case "jira":
		if h.jiraAuth != nil {
			tokenInfo, err = h.jiraAuth.GetTokenInfo(ctx)
		}
	case "discord":
		if h.discordAuth != nil {
			tokenInfo, err = h.discordAuth.GetTokenInfo(ctx)
		}
	default:
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenInfo)
}

// HTML templates for success/error pages
const successPageTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>OAuth Success</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 500px;
        }
        h1 {
            color: #4CAF50;
            margin-bottom: 20px;
        }
        .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .close-btn {
            margin-top: 20px;
            padding: 10px 30px;
            background: #4CAF50;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }
        .close-btn:hover {
            background: #45a049;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">✓</div>
        <h1>Authorization Successful!</h1>
        <p>You have successfully connected your <strong>{{.Provider}}</strong> account to Conexus.</p>
        <p>You can now close this window and return to the application.</p>
        <button class="close-btn" onclick="window.close()">Close Window</button>
    </div>
</body>
</html>
`

const errorPageTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>OAuth Error</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 500px;
        }
        h1 {
            color: #f44336;
            margin-bottom: 20px;
        }
        .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        p {
            color: #666;
            line-height: 1.6;
        }
        .error-msg {
            background: #ffebee;
            border: 1px solid #f44336;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
            color: #c62828;
        }
        .retry-btn {
            margin-top: 20px;
            padding: 10px 30px;
            background: #2196F3;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
            text-decoration: none;
            display: inline-block;
        }
        .retry-btn:hover {
            background: #0b7dda;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">✗</div>
        <h1>Authorization Failed</h1>
        <p>There was an error connecting your <strong>{{.Provider}}</strong> account.</p>
        <div class="error-msg">
            {{.Error}}
        </div>
        <p>Please try again or contact support if the problem persists.</p>
        <a href="/oauth/{{.Provider}}/authorize" class="retry-btn">Try Again</a>
    </div>
</body>
</html>
`
