package intent

import (
	"context"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name            string
		request         string
		expectedAgent   string
		expectedConfidence float64
		shouldError     bool
	}{
		{
			name:            "find files request",
			request:         "find all Go files in the internal directory",
			expectedAgent:   "codebase-locator",
			expectedConfidence: 0.3,
			shouldError:     false,
		},
		{
			name:            "analyze code request",
			request:         "analyze how the parser works",
			expectedAgent:   "codebase-analyzer",
			expectedConfidence: 0.5,
			shouldError:     false,
		},
		{
			name:            "find patterns request",
			request:         "pattern similar examples like this",
			expectedAgent:   "codebase-pattern-finder",
			expectedConfidence: 0.3,
			shouldError:     false,
		},
		{
			name:            "empty request",
			request:         "",
			expectedAgent:   "",
			expectedConfidence: 0.0,
			shouldError:     true,
		},
		{
			name:            "no matching pattern",
			request:         "xyz123 random words",
			expectedAgent:   "",
			expectedConfidence: 0.0,
			shouldError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intent, err := parser.Parse(context.Background(), tt.request)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if intent.PrimaryAgent != tt.expectedAgent {
				t.Errorf("expected agent %s, got %s", tt.expectedAgent, intent.PrimaryAgent)
			}

			if intent.Confidence < tt.expectedConfidence {
				t.Errorf("expected confidence >= %f, got %f", tt.expectedConfidence, intent.Confidence)
			}

			if intent.OriginalRequest != tt.request {
				t.Errorf("expected original request %s, got %s", tt.request, intent.OriginalRequest)
			}
		})
	}
}

func TestParser_ExtractEntities(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name           string
		request        string
		expectedEntity string
		entityKey      string
	}{
		{
			name:           "file pattern extraction",
			request:        "find parser.go in the codebase",
			expectedEntity: "parser.go",
			entityKey:      "file_pattern",
		},
		{
			name:           "directory extraction",
			request:        "search in internal/orchestrator directory",
			expectedEntity: "internal/orchestrator",
			entityKey:      "directory",
		},
		{
			name:           "quoted text extraction",
			request:        `find "NewParser" function`,
			expectedEntity: "NewParser",
			entityKey:      "quoted_text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entities := parser.extractEntities(tt.request, tt.request)

			if entity, ok := entities[tt.entityKey]; ok {
				if entity != tt.expectedEntity {
					t.Errorf("expected %s entity %s, got %s", tt.entityKey, tt.expectedEntity, entity)
				}
			} else {
				t.Errorf("expected entity key %s not found", tt.entityKey)
			}
		})
	}
}

func TestParser_AddPattern(t *testing.T) {
	parser := NewParser()
	initialCount := len(parser.GetPatterns())

	customPattern := NewKeywordPattern(
		"custom_test",
		[]string{"test"},
		"test-agent",
		nil,
		0.9,
	)

	parser.AddPattern(customPattern)

	if len(parser.GetPatterns()) != initialCount+1 {
		t.Errorf("expected %d patterns, got %d", initialCount+1, len(parser.GetPatterns()))
	}
}

func TestParser_SetPatterns(t *testing.T) {
	parser := NewParser()

	customPatterns := []Pattern{
		NewKeywordPattern("test1", []string{"test"}, "agent1", nil, 0.9),
		NewKeywordPattern("test2", []string{"test"}, "agent2", nil, 0.8),
	}

	parser.SetPatterns(customPatterns)

	if len(parser.GetPatterns()) != 2 {
		t.Errorf("expected 2 patterns, got %d", len(parser.GetPatterns()))
	}
}

func TestKeywordPattern_Match(t *testing.T) {
	pattern := NewKeywordPattern(
		"test_pattern",
		[]string{"find", "files"},
		"test-agent",
		nil,
		0.9,
	)

	tests := []struct {
		name        string
		input       string
		shouldMatch bool
		minScore    float64
	}{
		{
			name:        "full match",
			input:       "find files in directory",
			shouldMatch: true,
			minScore:    0.9,
		},
		{
			name:        "partial match",
			input:       "find something",
			shouldMatch: true,
			minScore:    0.45,
		},
		{
			name:        "no match",
			input:       "xyz random words",
			shouldMatch: false,
			minScore:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := pattern.Match(tt.input)

			if tt.shouldMatch {
				if match == nil {
					t.Errorf("expected match, got nil")
					return
				}

				if match.Score < tt.minScore {
					t.Errorf("expected score >= %f, got %f", tt.minScore, match.Score)
				}

				if match.Agent != "test-agent" {
					t.Errorf("expected agent %s, got %s", "test-agent", match.Agent)
				}
			} else {
				if match != nil {
					t.Errorf("expected no match, got %v", match)
				}
			}
		})
	}
}

func TestRegexPattern_Match(t *testing.T) {
	pattern := NewRegexPattern(
		"test_pattern",
		`error.*?handling`,
		"test-agent",
		nil,
		0.9,
	)

	tests := []struct {
		name        string
		input       string
		shouldMatch bool
	}{
		{
			name:        "matches pattern",
			input:       "error handling code",
			shouldMatch: true,
		},
		{
			name:        "no match",
			input:       "random text",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := pattern.Match(tt.input)

			if tt.shouldMatch {
				if match == nil {
					t.Errorf("expected match, got nil")
				}
			} else {
				if match != nil {
					t.Errorf("expected no match, got %v", match)
				}
			}
		})
	}
}
