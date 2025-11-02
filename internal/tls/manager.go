// Package tls provides TLS certificate management and secure HTTPS configuration.
package tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/config"
	"github.com/ferg-cod3s/conexus/internal/observability"
	"golang.org/x/crypto/acme/autocert"
)

// Manager handles TLS certificate management and server configuration.
type Manager struct {
	config    *config.TLSConfig
	logger    *observability.Logger
	certMgr   *autocert.Manager
	tlsConfig *tls.Config
}

// NewManager creates a new TLS manager with the given configuration.
func NewManager(cfg *config.TLSConfig, logger *observability.Logger) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("TLS config cannot be nil")
	}

	mgr := &Manager{
		config: cfg,
		logger: logger,
	}

	// Configure TLS settings
	tlsConfig, err := mgr.buildTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS config: %w", err)
	}
	mgr.tlsConfig = tlsConfig

	// Setup auto-cert if enabled
	if cfg.AutoCert {
		certMgr := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(cfg.AutoCertCacheDir),
			HostPolicy: autocert.HostWhitelist(cfg.AutoCertDomains...),
			Email:      cfg.AutoCertEmail,
		}
		mgr.certMgr = certMgr

		// Use auto-cert for TLS config
		mgr.tlsConfig.GetCertificate = certMgr.GetCertificate
	}

	return mgr, nil
}

// buildTLSConfig creates a secure TLS configuration.
func (m *Manager) buildTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:       m.parseTLSVersion(m.config.MinVersion),
		CipherSuites:     m.parseCipherSuites(m.config.CipherSuites),
		CurvePreferences: m.parseCurvePreferences(m.config.CurvePreferences),
	}

	// Set secure defaults if not specified
	if len(tlsConfig.CipherSuites) == 0 {
		tlsConfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		}
	}

	if len(tlsConfig.CurvePreferences) == 0 {
		tlsConfig.CurvePreferences = []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		}
	}

	return tlsConfig, nil
}

// parseTLSVersion converts string version to tls.Version constant.
func (m *Manager) parseTLSVersion(version string) uint16 {
	switch version {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		// Default to TLS 1.2 for security
		return tls.VersionTLS12
	}
}

// parseCipherSuites converts string cipher suite names to uint16 constants.
func (m *Manager) parseCipherSuites(suites []string) []uint16 {
	if len(suites) == 0 {
		return nil
	}

	var result []uint16
	for _, suite := range suites {
		suite = strings.TrimSpace(strings.ToUpper(suite))
		switch suite {
		case "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":
			result = append(result, tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384)
		case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		case "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":
			result = append(result, tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305)
		case "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":
			result = append(result, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305)
		case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
			result = append(result, tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256)
		case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
		default:
			m.logger.Warn("Unknown cipher suite, skipping", "suite", suite)
		}
	}
	return result
}

// parseCurvePreferences converts string curve names to tls.CurveID constants.
func (m *Manager) parseCurvePreferences(curves []string) []tls.CurveID {
	if len(curves) == 0 {
		return nil
	}

	var result []tls.CurveID
	for _, curve := range curves {
		curve = strings.TrimSpace(strings.ToUpper(curve))
		switch curve {
		case "X25519":
			result = append(result, tls.X25519)
		case "P256":
			result = append(result, tls.CurveP256)
		case "P384":
			result = append(result, tls.CurveP384)
		case "P521":
			result = append(result, tls.CurveP521)
		default:
			m.logger.Warn("Unknown curve preference, skipping", "curve", curve)
		}
	}
	return result
}

// GetTLSConfig returns the TLS configuration for the server.
func (m *Manager) GetTLSConfig() *tls.Config {
	return m.tlsConfig
}

// CreateHTTPRedirectServer creates an HTTP server that redirects to HTTPS.
func (m *Manager) CreateHTTPRedirectServer(httpsPort int) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf(":%d", m.config.HTTPRedirectPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Build HTTPS URL
			httpsURL := fmt.Sprintf("https://%s", r.Host)
			if httpsPort != 443 {
				httpsURL = fmt.Sprintf("https://%s:%d", r.Host, httpsPort)
			}
			httpsURL += r.RequestURI

			// Redirect to HTTPS
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
		}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

// ValidateCertificates validates that certificates can be loaded.
func (m *Manager) ValidateCertificates() error {
	if m.config.AutoCert {
		// For auto-cert, validation happens at runtime
		m.logger.Info("Using auto-cert - certificates will be obtained at runtime")
		return nil
	}

	if m.config.CertFile == "" || m.config.KeyFile == "" {
		return fmt.Errorf("cert_file and key_file must be specified when auto_cert is disabled")
	}

	// Try to load the certificate
	_, err := tls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %w", err)
	}

	m.logger.Info("Certificate validation successful",
		"cert_file", m.config.CertFile,
		"key_file", m.config.KeyFile)

	return nil
}

// StartHTTPRedirect starts the HTTP to HTTPS redirect server.
func (m *Manager) StartHTTPRedirect(ctx context.Context, httpsPort int) error {
	if !m.config.Enabled {
		return nil
	}

	redirectServer := m.CreateHTTPRedirectServer(httpsPort)

	m.logger.Info("Starting HTTP redirect server",
		"http_port", m.config.HTTPRedirectPort,
		"https_port", httpsPort)

	go func() {
		if err := redirectServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error("HTTP redirect server failed", "error", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redirectServer.Shutdown(shutdownCtx); err != nil {
			m.logger.Error("HTTP redirect server shutdown error", "error", err)
		}
	}()

	return nil
}
