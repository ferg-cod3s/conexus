package intent

import (
	"regexp"
	"strings"
)

// Pattern represents an intent matching pattern
type Pattern interface {
	// Match attempts to match the normalized request
	// Returns nil if no match, or a PatternMatch with score
	Match(normalized string) *PatternMatch

	// Name returns the pattern name
	Name() string
}

// PatternMatch represents a successful pattern match
type PatternMatch struct {
	// Name of the pattern that matched
	PatternName string

	// Primary agent to handle this request
	Agent string

	// Secondary agents that may be needed
	SecondaryAgents []string

	// Match confidence score (0.0 to 1.0)
	Score float64
}

// KeywordPattern matches based on keyword presence
type KeywordPattern struct {
	name            string
	keywords        []string
	agent           string
	secondaryAgents []string
	baseScore       float64
}

// NewKeywordPattern creates a new keyword-based pattern
func NewKeywordPattern(name string, keywords []string, agent string, secondaryAgents []string, baseScore float64) *KeywordPattern {
	return &KeywordPattern{
		name:            name,
		keywords:        keywords,
		agent:           agent,
		secondaryAgents: secondaryAgents,
		baseScore:       baseScore,
	}
}

// Match implements Pattern.Match
func (p *KeywordPattern) Match(normalized string) *PatternMatch {
	matchCount := 0
	for _, keyword := range p.keywords {
		if strings.Contains(normalized, keyword) {
			matchCount++
		}
	}

	if matchCount == 0 {
		return nil
	}

	// Score based on percentage of keywords matched
	score := p.baseScore * (float64(matchCount) / float64(len(p.keywords)))

	return &PatternMatch{
		PatternName:     p.name,
		Agent:           p.agent,
		SecondaryAgents: p.secondaryAgents,
		Score:           score,
	}
}

// Name implements Pattern.Name
func (p *KeywordPattern) Name() string {
	return p.name
}

// RegexPattern matches based on regular expressions
type RegexPattern struct {
	name            string
	regex           *regexp.Regexp
	agent           string
	secondaryAgents []string
	score           float64
}

// NewRegexPattern creates a new regex-based pattern
func NewRegexPattern(name string, pattern string, agent string, secondaryAgents []string, score float64) *RegexPattern {
	return &RegexPattern{
		name:            name,
		regex:           regexp.MustCompile(pattern),
		agent:           agent,
		secondaryAgents: secondaryAgents,
		score:           score,
	}
}

// Match implements Pattern.Match
func (p *RegexPattern) Match(normalized string) *PatternMatch {
	if !p.regex.MatchString(normalized) {
		return nil
	}

	return &PatternMatch{
		PatternName:     p.name,
		Agent:           p.agent,
		SecondaryAgents: p.secondaryAgents,
		Score:           p.score,
	}
}

// Name implements Pattern.Name
func (p *RegexPattern) Name() string {
	return p.name
}

// DefaultPatterns returns the default set of intent patterns
func DefaultPatterns() []Pattern {
	return []Pattern{
		// File location patterns
		NewKeywordPattern(
			"find_files",
			[]string{"find", "locate", "search", "where", "files"},
			"codebase-locator",
			nil,
			0.9,
		),

		// Symbol search patterns
		NewKeywordPattern(
			"find_symbols",
			[]string{"function", "class", "method", "symbol", "definition"},
			"codebase-locator",
			nil,
			0.85,
		),

		// Code analysis patterns
		NewKeywordPattern(
			"analyze_code",
			[]string{"analyze", "how", "works", "explain", "understand"},
			"codebase-analyzer",
			nil,
			0.9,
		),

		// Data flow patterns
		NewKeywordPattern(
			"trace_flow",
			[]string{"flow", "trace", "calls", "pathway", "sequence"},
			"codebase-analyzer",
			[]string{"codebase-locator"},
			0.85,
		),

		// Pattern detection
		NewKeywordPattern(
			"find_patterns",
			[]string{"pattern", "similar", "like", "examples"},
			"codebase-pattern-finder",
			[]string{"codebase-locator"},
			0.8,
		),

		// Multi-file analysis
		NewKeywordPattern(
			"analyze_multiple",
			[]string{"all", "every", "across", "throughout"},
			"codebase-analyzer",
			[]string{"codebase-locator"},
			0.75,
		),

		// Error analysis
		NewRegexPattern(
			"analyze_errors",
			`(error|exception|failure|bug).*?(handling|handler|catch)`,
			"codebase-analyzer",
			nil,
			0.9,
		),

		// Dependency analysis
		NewRegexPattern(
			"analyze_dependencies",
			`(import|require|dependency|dependencies|uses|using)`,
			"codebase-analyzer",
			[]string{"codebase-locator"},
			0.85,
		),

		// Implementation search
		NewRegexPattern(
			"find_implementation",
			`(implement|implementation|where.*?(defined|declared))`,
			"codebase-locator",
			[]string{"codebase-analyzer"},
			0.9,
		),

		// Call graph analysis
		NewRegexPattern(
			"analyze_calls",
			`(what.*?calls|who.*?calls|called.*?by|calling)`,
			"codebase-analyzer",
			nil,
			0.95,
		),
	}
}
