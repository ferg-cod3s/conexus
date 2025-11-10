package mcp

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitTicketInfo contains information about files related to a ticket ID
type GitTicketInfo struct {
	TicketID       string
	Branches       []string
	Commits        []CommitInfo
	ModifiedFiles  []string
	PRDescriptions []string
}

// CommitInfo contains information about a single commit
type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Date    string
	Files   []string
}

// findTicketInGit searches git history for a ticket ID and returns related files
func (s *Server) findTicketInGit(ctx context.Context, ticketID string, repoPath string) (*GitTicketInfo, error) {
	// Validate ticket ID format (alphanumeric, dash, underscore only)
	if !isValidTicketID(ticketID) {
		return nil, fmt.Errorf("invalid ticket ID format: %s", ticketID)
	}

	// Open the git repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	info := &GitTicketInfo{
		TicketID:      ticketID,
		Branches:      []string{},
		Commits:       []CommitInfo{},
		ModifiedFiles: []string{},
	}

	// Create case-insensitive pattern for ticket ID
	// Matches: TICKET-123, ticket-123, feature/TICKET-123, etc.
	ticketPattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(ticketID) + `\b`)

	// 1. Search branches for ticket ID
	branches, err := repo.Branches()
	if err == nil {
		_ = branches.ForEach(func(ref *plumbing.Reference) error {
			branchName := ref.Name().Short()
			if ticketPattern.MatchString(branchName) {
				info.Branches = append(info.Branches, branchName)
			}
			return nil
		})
	}

	// 2. Search commit messages for ticket ID
	commitIter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	// Limit to last 1000 commits for performance
	commitCount := 0
	maxCommits := 1000
	fileMap := make(map[string]bool) // Track unique files

	err = commitIter.ForEach(func(c *object.Commit) error {
		if commitCount >= maxCommits {
			return fmt.Errorf("reached max commits") // Stop iteration
		}
		commitCount++

		// Check if commit message or author contains ticket ID
		if ticketPattern.MatchString(c.Message) {
			commitInfo := CommitInfo{
				Hash:    c.Hash.String(),
				Message: strings.TrimSpace(c.Message),
				Author:  c.Author.Name,
				Date:    c.Author.When.Format("2006-01-02 15:04:05"),
				Files:   []string{},
			}

			// Get files modified in this commit
			if c.NumParents() > 0 {
				parent, err := c.Parent(0)
				if err == nil {
					changes, err := c.Patch(parent)
					if err == nil {
						for _, fileStat := range changes.Stats() {
							filePath := fileStat.Name
							commitInfo.Files = append(commitInfo.Files, filePath)
							fileMap[filePath] = true
						}
					}
				}
			} else {
				// First commit - get all files in tree
				tree, err := c.Tree()
				if err == nil {
					_ = tree.Files().ForEach(func(f *object.File) error {
						commitInfo.Files = append(commitInfo.Files, f.Name)
						fileMap[f.Name] = true
						return nil
					})
				}
			}

			info.Commits = append(info.Commits, commitInfo)
		}

		return nil
	})

	// Convert file map to slice
	for file := range fileMap {
		info.ModifiedFiles = append(info.ModifiedFiles, file)
	}

	// Return error only if it's not the "reached max commits" signal
	if err != nil && !strings.Contains(err.Error(), "reached max commits") {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return info, nil
}

// isValidTicketID validates that a ticket ID contains only safe characters
func isValidTicketID(ticketID string) bool {
	if ticketID == "" || len(ticketID) > 100 {
		return false
	}
	// Allow alphanumeric, dash, underscore, and dot
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	return validPattern.MatchString(ticketID)
}

// getRepoRoot finds the git repository root from the current working directory
func getRepoRoot(startPath string) (string, error) {
	// Clean the path
	path := filepath.Clean(startPath)

	// Try to open as git repo
	_, err := git.PlainOpen(path)
	if err == nil {
		return path, nil
	}

	// Walk up the directory tree
	for {
		parent := filepath.Dir(path)
		if parent == path {
			// Reached root
			return "", fmt.Errorf("not a git repository (or any parent up to root)")
		}

		_, err := git.PlainOpen(parent)
		if err == nil {
			return parent, nil
		}

		path = parent
	}
}
