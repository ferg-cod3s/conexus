package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// integrityChecker provides tamper-evident logging through HMAC
type integrityChecker struct {
	key []byte
}

// newIntegrityChecker creates a new integrity checker with the given key
func newIntegrityChecker(key string) (*integrityChecker, error) {
	if key == "" {
		return nil, fmt.Errorf("integrity key cannot be empty")
	}

	return &integrityChecker{
		key: []byte(key),
	}, nil
}

// generateHash creates an HMAC-SHA256 hash of the audit event
func (ic *integrityChecker) generateHash(event AuditEvent) string {
	// Create a normalized JSON representation for consistent hashing
	eventJSON, err := json.Marshal(event)
	if err != nil {
		// Fallback to string representation if JSON marshaling fails
		eventJSON = []byte(fmt.Sprintf("%+v", event))
	}

	// Create HMAC
	h := hmac.New(sha256.New, ic.key)
	h.Write(eventJSON)
	hash := h.Sum(nil)

	return hex.EncodeToString(hash)
}

// verifyHash verifies the integrity of an audit event
func (ic *integrityChecker) verifyHash(event AuditEvent, expectedHash string) bool {
	computedHash := ic.generateHash(event)
	return hmac.Equal([]byte(computedHash), []byte(expectedHash))
}
