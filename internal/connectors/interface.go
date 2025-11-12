package connectors

import (
	"context"
	"time"
)

// BaseConnector is the minimal interface that all connectors must implement.
// It provides basic metadata and rate limiting information.
type BaseConnector interface {
	// GetType returns the connector type (github, slack, jira, discord, etc.)
	GetType() string

	// GetRateLimit returns current rate limit information
	GetRateLimit() *RateLimitInfo

	// GetSyncStatus returns current synchronization status
	GetSyncStatus() *SyncStatus
}

// Syncable connectors can perform bulk data synchronization.
// This is optional - not all connectors need bulk sync capabilities.
type Syncable interface {
	BaseConnector
	// Note: Sync methods are connector-specific (SyncMessages, SyncIssues, etc.)
	// They cannot be abstracted due to different return types
}

// Searchable connectors support search operations with queries.
// This is optional - not all connectors need search capabilities.
type Searchable interface {
	BaseConnector
	// Note: Search methods are connector-specific (SearchMessages, SearchIssues, etc.)
	// They cannot be abstracted due to different parameters and return types
}

// ResourceLister connectors can list available resources like channels, projects, or repositories.
// This is optional - not all connectors need resource listing.
type ResourceLister interface {
	BaseConnector
	// Note: List methods are connector-specific (ListChannels, ListProjects, ListRepositories, etc.)
	// They cannot be abstracted due to different return types
}

// RateLimitInfo provides information about API rate limits
type RateLimitInfo struct {
	// Remaining is the number of requests remaining in the current rate limit window
	Remaining int `json:"remaining"`

	// Reset is the time when the rate limit will reset
	Reset time.Time `json:"reset"`

	// Limit is the maximum number of requests allowed in the rate limit window (optional)
	Limit int `json:"limit,omitempty"`
}

// SyncStatus provides information about the synchronization state
type SyncStatus struct {
	// LastSync is the time of the last successful sync
	LastSync time.Time `json:"last_sync"`

	// SyncInProgress indicates if a sync is currently running
	SyncInProgress bool `json:"sync_in_progress"`

	// Error contains the last error message if sync failed
	Error string `json:"error,omitempty"`

	// RateLimit contains the current rate limit information
	RateLimit *RateLimitInfo `json:"rate_limit,omitempty"`

	// Connector-specific stats (e.g., TotalMessages, TotalIssues, etc.)
	// These vary by connector type so we keep them in the concrete types
}

// ConnectorCapabilities describes what a connector can do.
// This is useful for runtime capability detection.
type ConnectorCapabilities struct {
	// Type is the connector type
	Type string

	// SupportsSync indicates if the connector supports bulk synchronization
	SupportsSync bool

	// SupportsSearch indicates if the connector supports search operations
	SupportsSearch bool

	// SupportsResourceListing indicates if the connector can list resources
	SupportsResourceListing bool

	// HasRateLimit indicates if the connector tracks rate limits
	HasRateLimit bool
}

// GetCapabilities returns the capabilities of a connector.
// This is a helper function that checks interface implementations.
func GetCapabilities(conn interface{}) ConnectorCapabilities {
	caps := ConnectorCapabilities{}

	// Check base connector
	if bc, ok := conn.(BaseConnector); ok {
		caps.Type = bc.GetType()
		caps.HasRateLimit = true
	}

	// Check optional interfaces
	if _, ok := conn.(Syncable); ok {
		caps.SupportsSync = true
	}

	if _, ok := conn.(Searchable); ok {
		caps.SupportsSearch = true
	}

	if _, ok := conn.(ResourceLister); ok {
		caps.SupportsResourceListing = true
	}

	return caps
}

// ValidateConnector checks if a connector implements the BaseConnector interface.
// Returns error if the connector doesn't meet minimum requirements.
func ValidateConnector(conn interface{}) error {
	if _, ok := conn.(BaseConnector); !ok {
		return ErrInvalidConnector
	}
	return nil
}
