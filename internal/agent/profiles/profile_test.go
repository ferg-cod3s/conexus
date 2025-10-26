package profiles

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllProfiles(t *testing.T) {
	profiles := GetAllProfiles()

	// Should have 6 predefined profiles
	assert.Len(t, profiles, 6, "should have 6 predefined profiles")

	// Check that all expected profiles are present
	profileIDs := make(map[string]bool)
	for _, profile := range profiles {
		profileIDs[profile.ID] = true
	}

	expectedIDs := []string{
		"code_analysis", "documentation", "debugging",
		"architecture", "security", "general",
	}

	for _, id := range expectedIDs {
		assert.True(t, profileIDs[id], "should have profile with ID: %s", id)
	}
}

func TestGetProfileByID(t *testing.T) {
	tests := []struct {
		id          string
		expectedID  string
		description string
	}{
		{
			id:          "code_analysis",
			expectedID:  "code_analysis",
			description: "should return code analysis profile",
		},
		{
			id:          "documentation",
			expectedID:  "documentation",
			description: "should return documentation profile",
		},
		{
			id:          "debugging",
			expectedID:  "debugging",
			description: "should return debugging profile",
		},
		{
			id:          "unknown",
			expectedID:  "general",
			description: "should return general profile for unknown ID",
		},
		{
			id:          "",
			expectedID:  "general",
			description: "should return general profile for empty ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			profile := GetProfileByID(tt.id)
			require.NotNil(t, profile, "profile should not be nil")
			assert.Equal(t, tt.expectedID, profile.ID, "should return correct profile ID")
		})
	}
}

func TestCodeAnalysisProfile(t *testing.T) {
	profile := CodeAnalysisProfile

	// Test basic properties
	assert.Equal(t, "code_analysis", profile.ID)
	assert.Equal(t, "Code Analysis Agent", profile.Name)
	assert.NotEmpty(t, profile.Description)

	// Test context window
	assert.Equal(t, 4000, profile.ContextWindow.MinTokens)
	assert.Equal(t, 12000, profile.ContextWindow.MaxTokens)
	assert.Equal(t, 8000, profile.ContextWindow.OptimalTokens)
	assert.Equal(t, 0.1, profile.ContextWindow.OverlapRatio)
	assert.False(t, profile.ContextWindow.Compression)

	// Test chunking strategy
	assert.Equal(t, "semantic", profile.ChunkingStrategy.Strategy)
	assert.Equal(t, 300, profile.ChunkingStrategy.ChunkSize)
	assert.Equal(t, 30, profile.ChunkingStrategy.Overlap)
	assert.NotEmpty(t, profile.ChunkingStrategy.LanguageRules)
	assert.NotEmpty(t, profile.ChunkingStrategy.ContentType)

	// Test weights
	assert.Equal(t, 1.0, profile.Weights.Code)
	assert.Equal(t, 0.6, profile.Weights.Documentation)
	assert.Equal(t, 0.3, profile.Weights.Discussions)

	// Test capabilities
	assert.Contains(t, profile.Capabilities, "code_analysis")
	assert.Contains(t, profile.Capabilities, "syntax_understanding")

	// Test optimization hints
	assert.Equal(t, "moderate", profile.OptimizationHints.CacheStrategy)
	assert.True(t, profile.OptimizationHints.PrefetchRelated)
	assert.Equal(t, 2, profile.OptimizationHints.ParallelQueries)
	assert.Equal(t, 5000, profile.OptimizationHints.TimeoutMs)
}

func TestDocumentationProfile(t *testing.T) {
	profile := DocumentationProfile

	// Test context window - should be larger for documentation
	assert.Equal(t, 8000, profile.ContextWindow.MinTokens)
	assert.Equal(t, 32000, profile.ContextWindow.MaxTokens)
	assert.Equal(t, 16000, profile.ContextWindow.OptimalTokens)
	assert.Equal(t, 0.15, profile.ContextWindow.OverlapRatio)
	assert.True(t, profile.ContextWindow.Compression)

	// Test chunking strategy
	assert.Equal(t, "hierarchical", profile.ChunkingStrategy.Strategy)
	assert.Equal(t, 600, profile.ChunkingStrategy.ChunkSize)
	assert.Equal(t, 90, profile.ChunkingStrategy.Overlap)

	// Test weights - documentation should be highest
	assert.Equal(t, 0.5, profile.Weights.Code)
	assert.Equal(t, 1.0, profile.Weights.Documentation)
	assert.Equal(t, 0.6, profile.Weights.Discussions)

	// Test capabilities
	assert.Contains(t, profile.Capabilities, "documentation_analysis")
	assert.Contains(t, profile.Capabilities, "explanation_generation")
}

func TestDebuggingProfile(t *testing.T) {
	profile := DebuggingProfile

	// Test context window - should be smaller and focused
	assert.Equal(t, 2000, profile.ContextWindow.MinTokens)
	assert.Equal(t, 8000, profile.ContextWindow.MaxTokens)
	assert.Equal(t, 4000, profile.ContextWindow.OptimalTokens)
	assert.Equal(t, 0.05, profile.ContextWindow.OverlapRatio)
	assert.False(t, profile.ContextWindow.Compression)

	// Test chunking strategy
	assert.Equal(t, "semantic", profile.ChunkingStrategy.Strategy)
	assert.Equal(t, 200, profile.ChunkingStrategy.ChunkSize)
	assert.Equal(t, 10, profile.ChunkingStrategy.Overlap)

	// Test weights - debugging should prioritize code and tests
	assert.Equal(t, 0.9, profile.Weights.Code)
	assert.Equal(t, 0.4, profile.Weights.Documentation)
	assert.Equal(t, 0.8, profile.Weights.Tests)

	// Test capabilities
	assert.Contains(t, profile.Capabilities, "error_analysis")
	assert.Contains(t, profile.Capabilities, "debug_analysis")

	// Test optimization hints - minimal caching for debugging
	assert.Equal(t, "minimal", profile.OptimizationHints.CacheStrategy)
	assert.False(t, profile.OptimizationHints.PrefetchRelated)
	assert.Equal(t, 1, profile.OptimizationHints.ParallelQueries)
	assert.Equal(t, 3000, profile.OptimizationHints.TimeoutMs)
}

func TestArchitectureProfile(t *testing.T) {
	profile := ArchitectureProfile

	// Test context window - should be largest for system-wide analysis
	assert.Equal(t, 12000, profile.ContextWindow.MinTokens)
	assert.Equal(t, 48000, profile.ContextWindow.MaxTokens)
	assert.Equal(t, 24000, profile.ContextWindow.OptimalTokens)
	assert.Equal(t, 0.2, profile.ContextWindow.OverlapRatio)
	assert.True(t, profile.ContextWindow.Compression)

	// Test chunking strategy
	assert.Equal(t, "hybrid", profile.ChunkingStrategy.Strategy)
	assert.Equal(t, 800, profile.ChunkingStrategy.ChunkSize)
	assert.Equal(t, 160, profile.ChunkingStrategy.Overlap)

	// Test weights - balanced but documentation-focused
	assert.Equal(t, 0.7, profile.Weights.Code)
	assert.Equal(t, 0.9, profile.Weights.Documentation)
	assert.Equal(t, 0.8, profile.Weights.Discussions)

	// Test capabilities
	assert.Contains(t, profile.Capabilities, "architecture_analysis")
	assert.Contains(t, profile.Capabilities, "system_design")

	// Test optimization hints - aggressive caching for architecture
	assert.Equal(t, "aggressive", profile.OptimizationHints.CacheStrategy)
	assert.True(t, profile.OptimizationHints.PrefetchRelated)
	assert.Equal(t, 4, profile.OptimizationHints.ParallelQueries)
	assert.Equal(t, 10000, profile.OptimizationHints.TimeoutMs)
}

func TestSecurityProfile(t *testing.T) {
	profile := SecurityProfile

	// Test context window - medium-sized, no compression
	assert.Equal(t, 6000, profile.ContextWindow.MinTokens)
	assert.Equal(t, 20000, profile.ContextWindow.MaxTokens)
	assert.Equal(t, 12000, profile.ContextWindow.OptimalTokens)
	assert.Equal(t, 0.1, profile.ContextWindow.OverlapRatio)
	assert.False(t, profile.ContextWindow.Compression)

	// Test chunking strategy
	assert.Equal(t, "semantic", profile.ChunkingStrategy.Strategy)
	assert.Equal(t, 400, profile.ChunkingStrategy.ChunkSize)
	assert.Equal(t, 40, profile.ChunkingStrategy.Overlap)

	// Test weights - security should prioritize code and config
	assert.Equal(t, 0.9, profile.Weights.Code)
	assert.Equal(t, 0.5, profile.Weights.Documentation)
	assert.Equal(t, 0.9, profile.Weights.Config)

	// Test capabilities
	assert.Contains(t, profile.Capabilities, "security_analysis")
	assert.Contains(t, profile.Capabilities, "vulnerability_detection")

	// Test optimization hints
	assert.Equal(t, "moderate", profile.OptimizationHints.CacheStrategy)
	assert.True(t, profile.OptimizationHints.PrefetchRelated)
	assert.Equal(t, 2, profile.OptimizationHints.ParallelQueries)
	assert.Equal(t, 6000, profile.OptimizationHints.TimeoutMs)
}

func TestGeneralProfile(t *testing.T) {
	profile := GeneralProfile

	// Test context window - balanced general purpose
	assert.Equal(t, 4000, profile.ContextWindow.MinTokens)
	assert.Equal(t, 16000, profile.ContextWindow.MaxTokens)
	assert.Equal(t, 8000, profile.ContextWindow.OptimalTokens)
	assert.Equal(t, 0.1, profile.ContextWindow.OverlapRatio)
	assert.True(t, profile.ContextWindow.Compression)

	// Test chunking strategy
	assert.Equal(t, "hybrid", profile.ChunkingStrategy.Strategy)
	assert.Equal(t, 400, profile.ChunkingStrategy.ChunkSize)
	assert.Equal(t, 40, profile.ChunkingStrategy.Overlap)

	// Test weights - balanced across all types
	assert.Equal(t, 0.6, profile.Weights.Code)
	assert.Equal(t, 0.7, profile.Weights.Documentation)
	assert.Equal(t, 0.5, profile.Weights.Discussions)

	// Test capabilities
	assert.Contains(t, profile.Capabilities, "general_analysis")
	assert.Contains(t, profile.Capabilities, "balanced_retrieval")

	// Test optimization hints
	assert.Equal(t, "moderate", profile.OptimizationHints.CacheStrategy)
	assert.False(t, profile.OptimizationHints.PrefetchRelated)
	assert.Equal(t, 2, profile.OptimizationHints.ParallelQueries)
	assert.Equal(t, 5000, profile.OptimizationHints.TimeoutMs)
}

func TestProfileTimestamps(t *testing.T) {
	profile := CodeAnalysisProfile

	// Test that timestamps are set
	assert.False(t, profile.CreatedAt.IsZero(), "created_at should be set")
	assert.False(t, profile.UpdatedAt.IsZero(), "updated_at should be set")
	assert.True(t, profile.UpdatedAt.After(profile.CreatedAt) || profile.UpdatedAt.Equal(profile.CreatedAt),
		"updated_at should be after or equal to created_at")
}

func TestProfileLanguageRules(t *testing.T) {
	profile := CodeAnalysisProfile

	// Test language-specific rules
	assert.Equal(t, "function_boundary", profile.ChunkingStrategy.LanguageRules["go"])
	assert.Equal(t, "function_boundary", profile.ChunkingStrategy.LanguageRules["js"])
	assert.Equal(t, "function_boundary", profile.ChunkingStrategy.LanguageRules["ts"])
	assert.Equal(t, "function_boundary", profile.ChunkingStrategy.LanguageRules["py"])
	assert.Equal(t, "class_method_boundary", profile.ChunkingStrategy.LanguageRules["java"])
	assert.Equal(t, "semantic_boundary", profile.ChunkingStrategy.LanguageRules["default"])
}

func TestProfileContentTypeRules(t *testing.T) {
	profile := CodeAnalysisProfile

	// Test content-type-specific rules
	assert.Equal(t, "function_semantic", profile.ChunkingStrategy.ContentType["code"])
	assert.Equal(t, "section_boundary", profile.ChunkingStrategy.ContentType["documentation"])
	assert.Equal(t, "key_value_pairs", profile.ChunkingStrategy.ContentType["config"])
	assert.Equal(t, "test_case_boundary", profile.ChunkingStrategy.ContentType["test"])
	assert.Equal(t, "message_thread", profile.ChunkingStrategy.ContentType["discussion"])
}

func TestProfileFallbackProfiles(t *testing.T) {
	// Test that all profiles have fallback profiles
	profiles := GetAllProfiles()
	for _, profile := range profiles {
		assert.NotEmpty(t, profile.OptimizationHints.FallbackProfile,
			"profile %s should have fallback profile", profile.ID)

		// Verify fallback profile exists
		fallback := GetProfileByID(profile.OptimizationHints.FallbackProfile)
		assert.NotNil(t, fallback,
			"fallback profile %s should exist for profile %s",
			profile.OptimizationHints.FallbackProfile, profile.ID)
	}
}

func TestProfilePriorityFeatures(t *testing.T) {
	profiles := GetAllProfiles()
	for _, profile := range profiles {
		assert.NotEmpty(t, profile.PriorityFeatures,
			"profile %s should have priority features", profile.ID)

		// Check that features are reasonable
		for _, feature := range profile.PriorityFeatures {
			assert.NotEmpty(t, feature,
				"priority feature should not be empty for profile %s", profile.ID)
		}
	}
}

func TestProfileOptimizationHints(t *testing.T) {
	profiles := GetAllProfiles()
	for _, profile := range profiles {
		// Test cache strategy is valid
		validCacheStrategies := []string{"aggressive", "moderate", "minimal"}
		assert.Contains(t, validCacheStrategies, profile.OptimizationHints.CacheStrategy,
			"invalid cache strategy for profile %s", profile.ID)

		// Test timeout is reasonable
		assert.Greater(t, profile.OptimizationHints.TimeoutMs, 0,
			"timeout should be positive for profile %s", profile.ID)
		assert.Less(t, profile.OptimizationHints.TimeoutMs, 30000,
			"timeout should be reasonable for profile %s", profile.ID)

		// Test retry attempts is reasonable
		assert.GreaterOrEqual(t, profile.OptimizationHints.RetryAttempts, 0,
			"retry attempts should be non-negative for profile %s", profile.ID)
		assert.LessOrEqual(t, profile.OptimizationHints.RetryAttempts, 5,
			"retry attempts should be reasonable for profile %s", profile.ID)

		// Test parallel queries is reasonable
		assert.GreaterOrEqual(t, profile.OptimizationHints.ParallelQueries, 1,
			"parallel queries should be at least 1 for profile %s", profile.ID)
		assert.LessOrEqual(t, profile.OptimizationHints.ParallelQueries, 10,
			"parallel queries should be reasonable for profile %s", profile.ID)
	}
}
