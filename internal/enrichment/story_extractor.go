package enrichment

import (
	"regexp"
)

type StoryExtractor struct {
	issuePattern  *regexp.Regexp
	prPattern     *regexp.Regexp
	branchPattern *regexp.Regexp
}

func NewStoryExtractor() *StoryExtractor {
	return &StoryExtractor{
		issuePattern:  regexp.MustCompile(`(?:#|PROJ-|JIRA-)(\d+)`),
		prPattern:     regexp.MustCompile(`(?:#|pull/)(\d+)`),
		branchPattern: regexp.MustCompile(`(?:feature|bugfix|hotfix)\/([A-Z]+-\d+)`),
	}
}

func (se *StoryExtractor) ExtractStoryReferences(content string) map[string][]string {
	references := make(map[string][]string)

	// Extract issue references
	if matches := se.issuePattern.FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				references["issues"] = append(references["issues"], match[1])
			}
		}
	}

	// Extract PR references
	if matches := se.prPattern.FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				references["prs"] = append(references["prs"], match[1])
			}
		}
	}

	// Extract branch references
	if matches := se.branchPattern.FindAllStringSubmatch(content, -1); matches != nil {
		for _, match := range matches {
			if len(match) > 1 {
				references["branches"] = append(references["branches"], match[1])
			}
		}
	}

	return references
}
