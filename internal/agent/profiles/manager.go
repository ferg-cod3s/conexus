package profiles

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProfileManager manages agent profiles and provides profile selection logic
type ProfileManager struct {
	profiles       map[string]*AgentProfile
	classifier     ProfileClassifier
	registry       *ProfileRegistry
	mu             sync.RWMutex
	defaultProfile *AgentProfile
}

// ProfileClassifier determines the appropriate profile for a given query
type ProfileClassifier interface {
	Classify(ctx context.Context, query string, workContext map[string]interface{}) (*ClassificationResult, error)
}

// ClassificationResult contains the result of profile classification
type ClassificationResult struct {
	ProfileID    string               `json:"profile_id"`
	Confidence   float64              `json:"confidence"`
	Reasoning    string               `json:"reasoning"`
	Alternatives []AlternativeProfile `json:"alternatives"`
}

// AlternativeProfile represents an alternative profile choice
type AlternativeProfile struct {
	ProfileID  string  `json:"profile_id"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// ProfileRegistry manages profile registration and discovery
type ProfileRegistry struct {
	profiles map[string]*AgentProfile
	mu       sync.RWMutex
}

// NewProfileManager creates a new profile manager
func NewProfileManager(classifier ProfileClassifier) *ProfileManager {
	pm := &ProfileManager{
		profiles:       make(map[string]*AgentProfile),
		classifier:     classifier,
		registry:       NewProfileRegistry(),
		defaultProfile: GeneralProfile,
	}

	// Register all predefined profiles
	for _, profile := range GetAllProfiles() {
		pm.RegisterProfile(profile)
	}

	return pm
}

// NewProfileRegistry creates a new profile registry
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		profiles: make(map[string]*AgentProfile),
	}
}

// RegisterProfile registers a new profile
func (pm *ProfileManager) RegisterProfile(profile *AgentProfile) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if profile.ID == "" {
		return fmt.Errorf("profile ID cannot be empty")
	}

	// Update timestamp
	profile.UpdatedAt = time.Now()

	pm.profiles[profile.ID] = profile
	pm.registry.Register(profile)

	return nil
}

// GetProfile retrieves a profile by ID
func (pm *ProfileManager) GetProfile(id string) (*AgentProfile, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profile, exists := pm.profiles[id]
	if !exists {
		return nil, fmt.Errorf("profile with ID '%s' not found", id)
	}

	return profile, nil
}

// GetAllProfiles returns all registered profiles
func (pm *ProfileManager) GetAllProfiles() []*AgentProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	profiles := make([]*AgentProfile, 0, len(pm.profiles))
	for _, profile := range pm.profiles {
		profiles = append(profiles, profile)
	}

	return profiles
}

// SelectProfile selects the best profile for a given query and context
func (pm *ProfileManager) SelectProfile(ctx context.Context, query string, workContext map[string]interface{}) (*AgentProfile, *ClassificationResult, error) {
	// Use classifier to determine the best profile
	result, err := pm.classifier.Classify(ctx, query, workContext)
	if err != nil {
		// Fallback to default profile on classification error
		return pm.defaultProfile, nil, fmt.Errorf("classification failed: %w, using default profile", err)
	}

	// Get the selected profile
	profile, err := pm.GetProfile(result.ProfileID)
	if err != nil {
		// Fallback to default profile if profile not found
		return pm.defaultProfile, result, fmt.Errorf("profile not found: %w, using default profile", err)
	}

	return profile, result, nil
}

// UpdateProfile updates an existing profile
func (pm *ProfileManager) UpdateProfile(profile *AgentProfile) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.profiles[profile.ID]; !exists {
		return fmt.Errorf("profile with ID '%s' does not exist", profile.ID)
	}

	profile.UpdatedAt = time.Now()
	pm.profiles[profile.ID] = profile
	pm.registry.Register(profile)

	return nil
}

// DeleteProfile removes a profile (except for predefined profiles)
func (pm *ProfileManager) DeleteProfile(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Prevent deletion of predefined profiles
	predefinedIDs := map[string]bool{
		"code_analysis": true,
		"documentation": true,
		"debugging":     true,
		"architecture":  true,
		"security":      true,
		"general":       true,
	}

	if predefinedIDs[id] {
		return fmt.Errorf("cannot delete predefined profile '%s'", id)
	}

	if _, exists := pm.profiles[id]; !exists {
		return fmt.Errorf("profile with ID '%s' not found", id)
	}

	delete(pm.profiles, id)
	pm.registry.Unregister(id)

	return nil
}

// GetProfilesByCapability returns profiles that have specific capabilities
func (pm *ProfileManager) GetProfilesByCapability(capability string) []*AgentProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var matchingProfiles []*AgentProfile
	for _, profile := range pm.profiles {
		for _, cap := range profile.Capabilities {
			if cap == capability {
				matchingProfiles = append(matchingProfiles, profile)
				break
			}
		}
	}

	return matchingProfiles
}

// ValidateProfile validates a profile configuration
func (pm *ProfileManager) ValidateProfile(profile *AgentProfile) error {
	if profile.ID == "" {
		return fmt.Errorf("profile ID is required")
	}

	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	if profile.ContextWindow.MinTokens <= 0 {
		return fmt.Errorf("context window min tokens must be positive")
	}

	if profile.ContextWindow.MaxTokens <= profile.ContextWindow.MinTokens {
		return fmt.Errorf("context window max tokens must be greater than min tokens")
	}

	if profile.ContextWindow.OptimalTokens < profile.ContextWindow.MinTokens ||
		profile.ContextWindow.OptimalTokens > profile.ContextWindow.MaxTokens {
		return fmt.Errorf("optimal tokens must be between min and max tokens")
	}

	if profile.ChunkingStrategy.ChunkSize <= 0 {
		return fmt.Errorf("chunk size must be positive")
	}

	if profile.ChunkingStrategy.Overlap < 0 {
		return fmt.Errorf("chunk overlap cannot be negative")
	}

	if profile.ChunkingStrategy.Overlap >= profile.ChunkingStrategy.ChunkSize {
		return fmt.Errorf("chunk overlap must be less than chunk size")
	}

	return nil
}

// ProfileRegistry methods

// Register registers a profile in the registry
func (pr *ProfileRegistry) Register(profile *AgentProfile) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.profiles[profile.ID] = profile
}

// Unregister removes a profile from the registry
func (pr *ProfileRegistry) Unregister(id string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	delete(pr.profiles, id)
}

// Get retrieves a profile from the registry
func (pr *ProfileRegistry) Get(id string) (*AgentProfile, bool) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	profile, exists := pr.profiles[id]
	return profile, exists
}

// List returns all profiles in the registry
func (pr *ProfileRegistry) List() []*AgentProfile {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	profiles := make([]*AgentProfile, 0, len(pr.profiles))
	for _, profile := range pr.profiles {
		profiles = append(profiles, profile)
	}

	return profiles
}

// GetStats returns registry statistics
func (pr *ProfileRegistry) GetStats() map[string]interface{} {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	stats := map[string]interface{}{
		"total_profiles":    len(pr.profiles),
		"capability_counts": make(map[string]int),
	}

	// Count capabilities across all profiles
	capabilityCounts := make(map[string]int)
	for _, profile := range pr.profiles {
		for _, capability := range profile.Capabilities {
			capabilityCounts[capability]++
		}
	}

	stats["capability_counts"] = capabilityCounts
	return stats
}
