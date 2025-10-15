// Package intent provides natural language intent parsing for user requests.
//
// The intent parser analyzes user requests to determine:
// - Which agents should handle the request
// - What parameters to extract
// - Confidence scores for routing decisions
package intent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Intent represents a parsed user intent
type Intent struct {
	// Primary agent to handle this request
	PrimaryAgent string

	// Additional agents that may be needed
	SecondaryAgents []string

	// Extracted entities (file paths, patterns, etc.)
	Entities map[string]string

	// Confidence score (0.0 to 1.0)
	Confidence float64

	// Original user request
	OriginalRequest string

	// Matched patterns that led to this intent
	MatchedPatterns []string
}

// Parser parses natural language requests into structured intents
type Parser struct {
	patterns []Pattern
}

// NewParser creates a new intent parser with default patterns
func NewParser() *Parser {
	return &Parser{
		patterns: DefaultPatterns(),
	}
}

// Parse analyzes a user request and returns the parsed intent
func (p *Parser) Parse(ctx context.Context, request string) (*Intent, error) {
	if request == "" {
		return nil, fmt.Errorf("empty request")
	}

	// Normalize the request
	normalized := strings.ToLower(strings.TrimSpace(request))

	intent := &Intent{
		OriginalRequest: request,
		Entities:        make(map[string]string),
		MatchedPatterns: []string{},
	}

	// Match against patterns
	var bestMatch *PatternMatch
	var bestScore float64 = 0.0

	for _, pattern := range p.patterns {
		match := pattern.Match(normalized)
		if match != nil && match.Score > bestScore {
			bestMatch = match
			bestScore = match.Score
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no matching pattern found for request: %s", request)
	}

	// Populate intent from best match
	intent.PrimaryAgent = bestMatch.Agent
	intent.SecondaryAgents = bestMatch.SecondaryAgents
	intent.Confidence = bestMatch.Score
	intent.MatchedPatterns = append(intent.MatchedPatterns, bestMatch.PatternName)

	// Extract entities
	intent.Entities = p.extractEntities(request, normalized)

	return intent, nil
}

// extractEntities extracts named entities from the request
func (p *Parser) extractEntities(original, normalized string) map[string]string {
	entities := make(map[string]string)

	// Extract file patterns
	filePattern := regexp.MustCompile(`([a-zA-Z0-9_/-]+\.(go|ts|js|py|java|rb|rs|c|cpp|h))`)
	if matches := filePattern.FindStringSubmatch(original); len(matches) > 0 {
		entities["file_pattern"] = matches[0]
	}

	// Extract glob patterns
	globPattern := regexp.MustCompile(`\*+\.[a-zA-Z0-9]+|\*+/[a-zA-Z0-9_/-]+`)
	if matches := globPattern.FindStringSubmatch(original); len(matches) > 0 {
		entities["glob_pattern"] = matches[0]
	}

	// Extract function/symbol names
	symbolPattern := regexp.MustCompile(`\b([A-Z][a-zA-Z0-9_]*)\b`)
	if matches := symbolPattern.FindStringSubmatch(original); len(matches) > 1 {
		entities["symbol"] = matches[1]
	}

	// Extract directory paths
	dirPattern := regexp.MustCompile(`(internal|pkg|cmd|src|lib)/[a-zA-Z0-9_/-]+`)
	if matches := dirPattern.FindStringSubmatch(original); len(matches) > 0 {
		entities["directory"] = matches[0]
	}

	// Extract quoted strings
	quotedPattern := regexp.MustCompile(`["']([^"']+)["']`)
	if matches := quotedPattern.FindStringSubmatch(original); len(matches) > 1 {
		entities["quoted_text"] = matches[1]
	}

	return entities
}

// AddPattern adds a custom pattern to the parser
func (p *Parser) AddPattern(pattern Pattern) {
	p.patterns = append(p.patterns, pattern)
}

// SetPatterns replaces all patterns with the provided set
func (p *Parser) SetPatterns(patterns []Pattern) {
	p.patterns = patterns
}

// GetPatterns returns all registered patterns
func (p *Parser) GetPatterns() []Pattern {
	return p.patterns
}
