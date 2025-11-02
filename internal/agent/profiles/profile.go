package profiles

import (
	"time"
)

// AgentProfile defines the context optimization profile for different agent types
type AgentProfile struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	ContextWindow     ContextWindow     `json:"context_window"`
	ChunkingStrategy  ChunkingStrategy  `json:"chunking_strategy"`
	PriorityFeatures  []string          `json:"priority_features"`
	Weights           ProfileWeights    `json:"weights"`
	Capabilities      []string          `json:"capabilities"`
	OptimizationHints OptimizationHints `json:"optimization_hints"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// ContextWindow defines optimal context window parameters for an agent type
type ContextWindow struct {
	MinTokens     int     `json:"min_tokens"`
	MaxTokens     int     `json:"max_tokens"`
	OptimalTokens int     `json:"optimal_tokens"`
	OverlapRatio  float64 `json:"overlap_ratio"`
	Compression   bool    `json:"compression"`
}

// ChunkingStrategy defines how content should be chunked for this agent type
type ChunkingStrategy struct {
	Strategy      string            `json:"strategy"` // "semantic", "fixed", "hierarchical", "hybrid"
	ChunkSize     int               `json:"chunk_size"`
	Overlap       int               `json:"overlap"`
	LanguageRules map[string]string `json:"language_rules"` // language-specific rules
	ContentType   map[string]string `json:"content_type"`   // content-type-specific rules
}

// ProfileWeights defines weighting factors for different content types
type ProfileWeights struct {
	Code          float64 `json:"code"`
	Documentation float64 `json:"documentation"`
	Discussions   float64 `json:"discussions"`
	Config        float64 `json:"config"`
	Tests         float64 `json:"tests"`
	Comments      float64 `json:"comments"`
	Metadata      float64 `json:"metadata"`
}

// OptimizationHints provides performance optimization hints for the agent
type OptimizationHints struct {
	CacheStrategy   string    `json:"cache_strategy"`   // "aggressive", "moderate", "minimal"
	PrefetchRelated bool      `json:"prefetch_related"` // whether to prefetch related content
	ParallelQueries int       `json:"parallel_queries"` // max parallel queries
	TimeoutMs       int       `json:"timeout_ms"`       // query timeout in milliseconds
	RetryAttempts   int       `json:"retry_attempts"`   // number of retry attempts
	FallbackProfile string    `json:"fallback_profile"` // fallback profile if classification fails
	LastUpdated     time.Time `json:"last_updated"`
}

// Predefined agent profiles for common agent types
var (
	// CodeAnalysisProfile optimized for code analysis and understanding
	CodeAnalysisProfile = &AgentProfile{
		ID:          "code_analysis",
		Name:        "Code Analysis Agent",
		Description: "Optimized for code analysis, function understanding, and syntax-aware context",
		ContextWindow: ContextWindow{
			MinTokens:     4000,
			MaxTokens:     12000,
			OptimalTokens: 8000,
			OverlapRatio:  0.1,
			Compression:   false,
		},
		ChunkingStrategy: ChunkingStrategy{
			Strategy:  "semantic",
			ChunkSize: 300,
			Overlap:   30,
			LanguageRules: map[string]string{
				"go":      "function_boundary",
				"js":      "function_boundary",
				"ts":      "function_boundary",
				"py":      "function_boundary",
				"java":    "class_method_boundary",
				"cpp":     "function_boundary",
				"rust":    "function_boundary",
				"default": "semantic_boundary",
			},
			ContentType: map[string]string{
				"code":          "function_semantic",
				"documentation": "section_boundary",
				"config":        "key_value_pairs",
				"test":          "test_case_boundary",
				"discussion":    "message_thread",
			},
		},
		PriorityFeatures: []string{
			"function_signatures",
			"type_definitions",
			"imports_dependencies",
			"error_handling",
			"test_coverage",
		},
		Weights: ProfileWeights{
			Code:          1.0,
			Documentation: 0.6,
			Discussions:   0.3,
			Config:        0.4,
			Tests:         0.8,
			Comments:      0.5,
			Metadata:      0.2,
		},
		Capabilities: []string{
			"code_analysis",
			"syntax_understanding",
			"dependency_tracking",
			"pattern_recognition",
		},
		OptimizationHints: OptimizationHints{
			CacheStrategy:   "moderate",
			PrefetchRelated: true,
			ParallelQueries: 2,
			TimeoutMs:       5000,
			RetryAttempts:   2,
			FallbackProfile: "general",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// DocumentationProfile optimized for documentation and explanatory content
	DocumentationProfile = &AgentProfile{
		ID:          "documentation",
		Name:        "Documentation Agent",
		Description: "Optimized for comprehensive documentation and explanatory context",
		ContextWindow: ContextWindow{
			MinTokens:     8000,
			MaxTokens:     32000,
			OptimalTokens: 16000,
			OverlapRatio:  0.15,
			Compression:   true,
		},
		ChunkingStrategy: ChunkingStrategy{
			Strategy:  "hierarchical",
			ChunkSize: 600,
			Overlap:   90,
			LanguageRules: map[string]string{
				"md":      "section_hierarchy",
				"rst":     "section_hierarchy",
				"txt":     "paragraph_boundary",
				"default": "semantic_boundary",
			},
			ContentType: map[string]string{
				"documentation": "section_hierarchy",
				"code":          "code_block_boundary",
				"config":        "key_value_pairs",
				"discussion":    "message_thread",
			},
		},
		PriorityFeatures: []string{
			"section_structure",
			"code_examples",
			"api_references",
			"tutorials",
			"explanations",
		},
		Weights: ProfileWeights{
			Code:          0.5,
			Documentation: 1.0,
			Discussions:   0.6,
			Config:        0.3,
			Tests:         0.4,
			Comments:      0.7,
			Metadata:      0.4,
		},
		Capabilities: []string{
			"documentation_analysis",
			"content_organization",
			"explanation_generation",
			"tutorial_creation",
		},
		OptimizationHints: OptimizationHints{
			CacheStrategy:   "aggressive",
			PrefetchRelated: true,
			ParallelQueries: 3,
			TimeoutMs:       8000,
			RetryAttempts:   3,
			FallbackProfile: "general",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// DebuggingProfile optimized for debugging and error analysis
	DebuggingProfile = &AgentProfile{
		ID:          "debugging",
		Name:        "Debugging Agent",
		Description: "Optimized for precise, error-focused context and debugging tasks",
		ContextWindow: ContextWindow{
			MinTokens:     2000,
			MaxTokens:     8000,
			OptimalTokens: 4000,
			OverlapRatio:  0.05,
			Compression:   false,
		},
		ChunkingStrategy: ChunkingStrategy{
			Strategy:  "semantic",
			ChunkSize: 200,
			Overlap:   10,
			LanguageRules: map[string]string{
				"go":      "error_handling_boundary",
				"js":      "error_handling_boundary",
				"ts":      "error_handling_boundary",
				"py":      "exception_boundary",
				"java":    "exception_boundary",
				"default": "semantic_boundary",
			},
			ContentType: map[string]string{
				"code":          "error_context_boundary",
				"logs":          "log_entry_boundary",
				"stack_traces":  "stack_trace_boundary",
				"documentation": "troubleshooting_section",
			},
		},
		PriorityFeatures: []string{
			"error_messages",
			"stack_traces",
			"exception_handling",
			"log_entries",
			"debug_symbols",
		},
		Weights: ProfileWeights{
			Code:          0.9,
			Documentation: 0.4,
			Discussions:   0.7, // bug reports, debugging discussions
			Config:        0.5,
			Tests:         0.8,
			Comments:      0.6,
			Metadata:      0.3,
		},
		Capabilities: []string{
			"error_analysis",
			"debug_analysis",
			"stack_trace_analysis",
			"log_analysis",
			"bug_detection",
		},
		OptimizationHints: OptimizationHints{
			CacheStrategy:   "minimal", // need fresh data for debugging
			PrefetchRelated: false,
			ParallelQueries: 1,
			TimeoutMs:       3000,
			RetryAttempts:   1,
			FallbackProfile: "code_analysis",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ArchitectureProfile optimized for system architecture and design analysis
	ArchitectureProfile = &AgentProfile{
		ID:          "architecture",
		Name:        "Architecture Agent",
		Description: "Optimized for holistic, system-wide context and architectural analysis",
		ContextWindow: ContextWindow{
			MinTokens:     12000,
			MaxTokens:     48000,
			OptimalTokens: 24000,
			OverlapRatio:  0.2,
			Compression:   true,
		},
		ChunkingStrategy: ChunkingStrategy{
			Strategy:  "hybrid",
			ChunkSize: 800,
			Overlap:   160,
			LanguageRules: map[string]string{
				"default": "module_boundary",
			},
			ContentType: map[string]string{
				"code":          "module_architecture_boundary",
				"documentation": "architecture_section",
				"config":        "system_config_boundary",
				"diagrams":      "diagram_boundary",
			},
		},
		PriorityFeatures: []string{
			"system_design",
			"module_relationships",
			"data_flow",
			"architecture_patterns",
			"design_decisions",
		},
		Weights: ProfileWeights{
			Code:          0.7,
			Documentation: 0.9,
			Discussions:   0.8,
			Config:        0.8,
			Tests:         0.5,
			Comments:      0.6,
			Metadata:      0.7,
		},
		Capabilities: []string{
			"architecture_analysis",
			"system_design",
			"pattern_recognition",
			"dependency_analysis",
		},
		OptimizationHints: OptimizationHints{
			CacheStrategy:   "aggressive",
			PrefetchRelated: true,
			ParallelQueries: 4,
			TimeoutMs:       10000,
			RetryAttempts:   3,
			FallbackProfile: "documentation",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// SecurityProfile optimized for security analysis and vulnerability detection
	SecurityProfile = &AgentProfile{
		ID:          "security",
		Name:        "Security Agent",
		Description: "Optimized for vulnerability-focused context and security analysis",
		ContextWindow: ContextWindow{
			MinTokens:     6000,
			MaxTokens:     20000,
			OptimalTokens: 12000,
			OverlapRatio:  0.1,
			Compression:   false, // need exact security-relevant code
		},
		ChunkingStrategy: ChunkingStrategy{
			Strategy:  "semantic",
			ChunkSize: 400,
			Overlap:   40,
			LanguageRules: map[string]string{
				"default": "security_boundary",
			},
			ContentType: map[string]string{
				"code":                  "security_function_boundary",
				"config":                "security_config_boundary",
				"documentation":         "security_section",
				"vulnerability_reports": "vulnerability_boundary",
			},
		},
		PriorityFeatures: []string{
			"security_functions",
			"authentication_code",
			"authorization_logic",
			"input_validation",
			"vulnerability_patterns",
		},
		Weights: ProfileWeights{
			Code:          0.9,
			Documentation: 0.5,
			Discussions:   0.8, // security discussions, vulnerability reports
			Config:        0.9, // security configurations
			Tests:         0.7,
			Comments:      0.4,
			Metadata:      0.6,
		},
		Capabilities: []string{
			"security_analysis",
			"vulnerability_detection",
			"security_audit",
			"compliance_checking",
		},
		OptimizationHints: OptimizationHints{
			CacheStrategy:   "moderate",
			PrefetchRelated: true,
			ParallelQueries: 2,
			TimeoutMs:       6000,
			RetryAttempts:   2,
			FallbackProfile: "code_analysis",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// GeneralProfile as fallback for general-purpose queries
	GeneralProfile = &AgentProfile{
		ID:          "general",
		Name:        "General Agent",
		Description: "General-purpose profile for balanced context retrieval",
		ContextWindow: ContextWindow{
			MinTokens:     4000,
			MaxTokens:     16000,
			OptimalTokens: 8000,
			OverlapRatio:  0.1,
			Compression:   true,
		},
		ChunkingStrategy: ChunkingStrategy{
			Strategy:  "hybrid",
			ChunkSize: 400,
			Overlap:   40,
			LanguageRules: map[string]string{
				"default": "semantic_boundary",
			},
			ContentType: map[string]string{
				"code":          "semantic_boundary",
				"documentation": "section_boundary",
				"config":        "key_value_pairs",
				"discussion":    "message_boundary",
			},
		},
		PriorityFeatures: []string{
			"general_content",
			"mixed_types",
			"balanced_context",
		},
		Weights: ProfileWeights{
			Code:          0.6,
			Documentation: 0.7,
			Discussions:   0.5,
			Config:        0.5,
			Tests:         0.6,
			Comments:      0.5,
			Metadata:      0.4,
		},
		Capabilities: []string{
			"general_analysis",
			"mixed_content_handling",
			"balanced_retrieval",
		},
		OptimizationHints: OptimizationHints{
			CacheStrategy:   "moderate",
			PrefetchRelated: false,
			ParallelQueries: 2,
			TimeoutMs:       5000,
			RetryAttempts:   2,
			FallbackProfile: "general",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
)

// GetAllProfiles returns all predefined agent profiles
func GetAllProfiles() []*AgentProfile {
	return []*AgentProfile{
		CodeAnalysisProfile,
		DocumentationProfile,
		DebuggingProfile,
		ArchitectureProfile,
		SecurityProfile,
		GeneralProfile,
	}
}

// GetProfileByID returns a profile by its ID
func GetProfileByID(id string) *AgentProfile {
	profiles := GetAllProfiles()
	for _, profile := range profiles {
		if profile.ID == id {
			return profile
		}
	}
	return GeneralProfile // fallback to general profile
}
