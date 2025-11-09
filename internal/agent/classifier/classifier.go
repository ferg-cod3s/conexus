package classifier

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ferg-cod3s/conexus/internal/agent/profiles"
)

// QueryClassifier implements profile classification logic
type QueryClassifier struct {
	patterns    map[string]*ClassificationPattern
	mlModel     *MLModel // Future: ML-based classification
	confidence  float64  // Minimum confidence threshold
	lastUpdated time.Time
}

// ClassificationPattern defines patterns for profile classification
type ClassificationPattern struct {
	ProfileID     string            `json:"profile_id"`
	Keywords      []string          `json:"keywords"`
	RegexPatterns []string          `json:"regex_patterns"`
	ContextRules  map[string]string `json:"context_rules"`
	Weight        float64           `json:"weight"`
}

// MLModel represents a machine learning model for classification (placeholder)
type MLModel struct {
	// Future implementation for ML-based classification
	Trained bool
	Version string
}

// ClassificationRequest represents a classification request
type ClassificationRequest struct {
	Query       string                 `json:"query"`
	WorkContext map[string]interface{} `json:"work_context"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ClassificationScore represents a score for a specific profile
type ClassificationScore struct {
	ProfileID   string   `json:"profile_id"`
	Score       float64  `json:"score"`
	Confidence  float64  `json:"confidence"`
	Reasoning   string   `json:"reasoning"`
	MatchedKeys []string `json:"matched_keys"`
}

// NewQueryClassifier creates a new query classifier
func NewQueryClassifier() *QueryClassifier {
	classifier := &QueryClassifier{
		patterns:    make(map[string]*ClassificationPattern),
		confidence:  0.7, // 70% minimum confidence
		lastUpdated: time.Now(),
	}

	// Initialize classification patterns
	classifier.initializePatterns()

	return classifier
}

// initializePatterns sets up classification patterns for different agent types
func (qc *QueryClassifier) initializePatterns() {
	// Code Analysis patterns
	qc.patterns["code_analysis"] = &ClassificationPattern{
		ProfileID: "code_analysis",
		Keywords: []string{
			"function", "class", "method", "variable", "algorithm", "implementation",
			"code", "programming", "syntax", "logic", "structure", "pattern",
			"refactor", "optimize", "debug", "test", "interface", "api",
		},
		RegexPatterns: []string{
			`(?i)how.*implement`,
			`(?i)what.*function`,
			`(?i)explain.*code`,
			`(?i)analyze.*function`,
			`(?i)refactor.*code`,
			`(?i)optimize.*algorithm`,
			`(?i)debug.*function`,
			`(?i)write.*function`,
			`(?i)create.*class`,
		},
		ContextRules: map[string]string{
			"active_file":  `\.(go|js|ts|py|java|cpp|rs)$`,
			"git_branch":   "feature/.*|bugfix/.*|hotfix/.*",
			"open_tickets": "BUG|TASK|STORY",
		},
		Weight: 1.0,
	}

	// Documentation patterns
	qc.patterns["documentation"] = &ClassificationPattern{
		ProfileID: "documentation",
		Keywords: []string{
			"documentation", "readme", "guide", "tutorial", "explain", "overview",
			"introduction", "getting started", "how to", "usage", "example",
			"manual", "reference", "api docs", "docs", "walkthrough",
		},
		RegexPatterns: []string{
			`(?i)explain.*concept`,
			`(?i)how.*work`,
			`(?i)what.*purpose`,
			`(?i)documentation.*for`,
			`(?i)guide.*using`,
			`(?i)tutorial.*on`,
			`(?i)readme.*content`,
			`(?i)api.*reference`,
		},
		ContextRules: map[string]string{
			"active_file":  `\.(md|rst|txt|doc)$`,
			"git_branch":   "docs/.*|documentation/.*",
			"open_tickets": "DOC|README|GUIDE",
		},
		Weight: 0.9,
	}

	// Debugging patterns
	qc.patterns["debugging"] = &ClassificationPattern{
		ProfileID: "debugging",
		Keywords: []string{
			"error", "bug", "issue", "problem", "crash", "exception", "fail",
			"debug", "troubleshoot", "fix", "broken", "not working", "stack trace",
			"panic", "assert", "validation", "incorrect", "unexpected",
		},
		RegexPatterns: []string{
			`(?i)error.*message`,
			`(?i)stack.*trace`,
			`(?i)debug.*issue`,
			`(?i)fix.*bug`,
			`(?i)troubleshoot`,
			`(?i)why.*fail`,
			`(?i)exception.*thrown`,
			`(?i)panic.*occurred`,
			`(?i)assert.*failed`,
		},
		ContextRules: map[string]string{
			"active_file":  `\.(go|js|ts|py|java|cpp|rs)$`,
			"git_branch":   "bugfix/.*|hotfix/.*|debug/.*",
			"open_tickets": "BUG|ERROR|ISSUE|CRASH",
		},
		Weight: 1.1, // Higher weight for debugging
	}

	// Architecture patterns
	qc.patterns["architecture"] = &ClassificationPattern{
		ProfileID: "architecture",
		Keywords: []string{
			"architecture", "design", "system", "structure", "pattern", "framework",
			"microservice", "module", "component", "integration", "scalability",
			"diagram", "flow", "pipeline", "workflow", "organization",
		},
		RegexPatterns: []string{
			`(?i)system.*design`,
			`(?i)architecture.*pattern`,
			`(?i)how.*organize`,
			`(?i)module.*structure`,
			`(?i)component.*design`,
			`(?i)integration.*pattern`,
			`(?i)scalability.*approach`,
			`(?i)data.*flow`,
			`(?i)service.*architecture`,
		},
		ContextRules: map[string]string{
			"active_file":  `\.(go|js|ts|py|java|cpp|rs)$`,
			"git_branch":   "arch/.*|design/.*|refactor/.*",
			"open_tickets": "ARCH|DESIGN|REFACTOR",
		},
		Weight: 0.95,
	}

	// Security patterns
	qc.patterns["security"] = &ClassificationPattern{
		ProfileID: "security",
		Keywords: []string{
			"security", "authentication", "authorization", "vulnerability", "encryption",
			"password", "token", "jwt", "oauth", "ssl", "tls", "https",
			"input validation", "sanitization", "xss", "csrf", "injection",
		},
		RegexPatterns: []string{
			`(?i)security.*implementation`,
			`(?i)authentication.*flow`,
			`(?i)authorization.*check`,
			`(?i)vulnerability.*scan`,
			`(?i)password.*policy`,
			`(?i)token.*validation`,
			`(?i)input.*sanitization`,
			`(?i)xss.*prevention`,
			`(?i)csrf.*protection`,
			`(?i)sql.*injection`,
		},
		ContextRules: map[string]string{
			"active_file":  `\.(go|js|ts|py|java|cpp|rs)$`,
			"git_branch":   "security/.*|auth/.*|fix/.*",
			"open_tickets": "SECURITY|AUTH|VULN|CVE",
		},
		Weight: 1.05, // Slightly higher weight for security
	}
}

// Classify determines the best profile for a given query and context
func (qc *QueryClassifier) Classify(ctx context.Context, query string, workContext map[string]interface{}) (*profiles.ClassificationResult, error) {
	if query == "" {
		return &profiles.ClassificationResult{
			ProfileID:    "general",
			Confidence:   0.5,
			Reasoning:    "Empty query, using general profile",
			Alternatives: []profiles.AlternativeProfile{},
		}, nil
	}

	// Calculate scores for each profile
	scores := qc.calculateScores(query, workContext)

	// Find the best match
	bestScore := qc.findBestScore(scores)

	// Generate alternatives
	alternatives := qc.generateAlternatives(scores, bestScore.ProfileID)

	// Check confidence threshold
	if bestScore.Confidence < qc.confidence {
		return &profiles.ClassificationResult{
			ProfileID:    "general",
			Confidence:   bestScore.Confidence,
			Reasoning:    fmt.Sprintf("Low confidence (%.2f) for %s, using general profile. Best match: %s", bestScore.Confidence, bestScore.ProfileID, bestScore.Reasoning),
			Alternatives: alternatives,
		}, nil
	}

	return &profiles.ClassificationResult{
		ProfileID:    bestScore.ProfileID,
		Confidence:   bestScore.Confidence,
		Reasoning:    bestScore.Reasoning,
		Alternatives: alternatives,
	}, nil
}

// calculateScores calculates classification scores for all profiles
func (qc *QueryClassifier) calculateScores(query string, workContext map[string]interface{}) map[string]*ClassificationScore {
	scores := make(map[string]*ClassificationScore)

	lowerQuery := strings.ToLower(query)

	for profileID, pattern := range qc.patterns {
		score := &ClassificationScore{
			ProfileID: profileID,
		}

		// Keyword matching
		keywordScore := qc.calculateKeywordScore(lowerQuery, pattern.Keywords)

		// Regex pattern matching
		regexScore := qc.calculateRegexScore(lowerQuery, pattern.RegexPatterns)

		// Context matching
		contextScore := qc.calculateContextScore(workContext, pattern.ContextRules)

		// Combine scores with weights
		totalScore := (keywordScore * 0.4) + (regexScore * 0.4) + (contextScore * 0.2)
		finalScore := totalScore * pattern.Weight

		score.Score = finalScore
		score.Confidence = qc.normalizeScore(finalScore)
		score.MatchedKeys = qc.getMatchedKeywords(lowerQuery, pattern.Keywords)
		score.Reasoning = qc.generateReasoning(keywordScore, regexScore, contextScore, score.MatchedKeys)

		scores[profileID] = score
	}

	return scores
}

// calculateKeywordScore calculates score based on keyword matching
func (qc *QueryClassifier) calculateKeywordScore(query string, keywords []string) float64 {
	if len(keywords) == 0 {
		return 0.0
	}

	matchedCount := 0
	queryWords := strings.Fields(query)

	for _, keyword := range keywords {
		lowerKeyword := strings.ToLower(keyword)
		for _, word := range queryWords {
			if strings.Contains(word, lowerKeyword) || strings.Contains(lowerKeyword, word) {
				matchedCount++
				break
			}
		}
	}

	return float64(matchedCount) / float64(len(keywords))
}

// calculateRegexScore calculates score based on regex pattern matching
func (qc *QueryClassifier) calculateRegexScore(query string, patterns []string) float64 {
	if len(patterns) == 0 {
		return 0.0
	}

	matchedCount := 0
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue // Skip invalid patterns
		}
		if re.MatchString(query) {
			matchedCount++
		}
	}

	return float64(matchedCount) / float64(len(patterns))
}

// calculateContextScore calculates score based on work context matching
func (qc *QueryClassifier) calculateContextScore(workContext map[string]interface{}, rules map[string]string) float64 {
	if len(rules) == 0 || len(workContext) == 0 {
		return 0.0
	}

	matchedRules := 0
	totalRules := 0

	for contextKey, rule := range rules {
		totalRules++
		if value, exists := workContext[contextKey]; exists {
			if strValue, ok := value.(string); ok {
				re, err := regexp.Compile(rule)
				if err != nil {
					continue
				}
				if re.MatchString(strValue) {
					matchedRules++
				}
			}
		}
	}

	if totalRules == 0 {
		return 0.0
	}

	return float64(matchedRules) / float64(totalRules)
}

// normalizeScore normalizes score to 0-1 range
func (qc *QueryClassifier) normalizeScore(score float64) float64 {
	if score > 1.0 {
		return 1.0
	}
	if score < 0.0 {
		return 0.0
	}
	return score
}

// getMatchedKeywords returns keywords that matched in the query
func (qc *QueryClassifier) getMatchedKeywords(query string, keywords []string) []string {
	var matched []string
	queryWords := strings.Fields(query)

	for _, keyword := range keywords {
		lowerKeyword := strings.ToLower(keyword)
		for _, word := range queryWords {
			if strings.Contains(word, lowerKeyword) || strings.Contains(lowerKeyword, word) {
				matched = append(matched, keyword)
				break
			}
		}
	}

	return matched
}

// generateReasoning creates explanation for classification score
func (qc *QueryClassifier) generateReasoning(keywordScore, regexScore, contextScore float64, matchedKeys []string) string {
	var parts []string

	if len(matchedKeys) > 0 {
		parts = append(parts, fmt.Sprintf("matched keywords: %s", strings.Join(matchedKeys, ", ")))
	}

	if regexScore > 0 {
		parts = append(parts, fmt.Sprintf("pattern matches: %.2f", regexScore))
	}

	if contextScore > 0 {
		parts = append(parts, fmt.Sprintf("context matches: %.2f", contextScore))
	}

	if len(parts) == 0 {
		return "no specific patterns matched"
	}

	return strings.Join(parts, "; ")
}

// findBestScore finds the highest scoring profile
func (qc *QueryClassifier) findBestScore(scores map[string]*ClassificationScore) *ClassificationScore {
	var best *ClassificationScore
	for _, score := range scores {
		if best == nil || score.Score > best.Score {
			best = score
		}
	}
	return best
}

// generateAlternatives creates alternative profile suggestions
func (qc *QueryClassifier) generateAlternatives(scores map[string]*ClassificationScore, bestProfileID string) []profiles.AlternativeProfile {
	var alternatives []profiles.AlternativeProfile

	// Sort profiles by score
	type profileScore struct {
		id    string
		score *ClassificationScore
	}

	var sortedProfiles []profileScore
	for id, score := range scores {
		if id != bestProfileID {
			sortedProfiles = append(sortedProfiles, profileScore{id, score})
		}
	}

	// Simple sort by score (descending)
	for i := 0; i < len(sortedProfiles); i++ {
		for j := i + 1; j < len(sortedProfiles); j++ {
			if sortedProfiles[i].score.Score < sortedProfiles[j].score.Score {
				sortedProfiles[i], sortedProfiles[j] = sortedProfiles[j], sortedProfiles[i]
			}
		}
	}

	// Take top 3 alternatives
	maxAlternatives := 3
	if len(sortedProfiles) < maxAlternatives {
		maxAlternatives = len(sortedProfiles)
	}

	for i := 0; i < maxAlternatives; i++ {
		alt := sortedProfiles[i]
		alternatives = append(alternatives, profiles.AlternativeProfile{
			ProfileID:  alt.id,
			Confidence: alt.score.Confidence,
			Reason:     alt.score.Reasoning,
		})
	}

	return alternatives
}

// UpdateConfidence updates the minimum confidence threshold
func (qc *QueryClassifier) UpdateConfidence(confidence float64) {
	if confidence >= 0.0 && confidence <= 1.0 {
		qc.confidence = confidence
		qc.lastUpdated = time.Now()
	}
}

// GetPatterns returns all classification patterns
func (qc *QueryClassifier) GetPatterns() map[string]*ClassificationPattern {
	// Return a copy to prevent external modification
	patterns := make(map[string]*ClassificationPattern)
	for id, pattern := range qc.patterns {
		patterns[id] = pattern
	}
	return patterns
}

// AddPattern adds a new classification pattern
func (qc *QueryClassifier) AddPattern(profileID string, pattern *ClassificationPattern) error {
	if profileID == "" {
		return fmt.Errorf("profile ID cannot be empty")
	}
	if pattern == nil {
		return fmt.Errorf("pattern cannot be nil")
	}

	qc.patterns[profileID] = pattern
	qc.lastUpdated = time.Now()
	return nil
}

// RemovePattern removes a classification pattern
func (qc *QueryClassifier) RemovePattern(profileID string) {
	delete(qc.patterns, profileID)
	qc.lastUpdated = time.Now()
}
