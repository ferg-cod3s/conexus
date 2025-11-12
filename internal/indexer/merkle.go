package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ferg-cod3s/conexus/internal/security"
)

// merkleTree is the concrete implementation of the MerkleTree interface.
// It uses SHA256 hashing to create a hierarchical hash tree matching the filesystem layout.
type merkleTree struct {
	walker Walker
}

// NewMerkleTree creates a new MerkleTree implementation.
func NewMerkleTree(walker Walker) MerkleTree {
	return &merkleTree{
		walker: walker,
	}
}

// treeNode represents a node in the internal Merkle tree structure.
type treeNode struct {
	Path     string               `json:"path"`
	Hash     string               `json:"hash"`
	IsFile   bool                 `json:"isFile"`
	Size     int64                `json:"size"`
	Children map[string]*treeNode `json:"children,omitempty"`
}

// treeState represents the serializable state of a Merkle tree.
type treeState struct {
	Root *treeNode `json:"root"`
}

// Hash computes a Merkle tree hash for the given directory.
// Returns a compact JSON representation of the tree state for later comparison.
func (mt *merkleTree) Hash(ctx context.Context, root string, ignorePatterns []string) ([]byte, error) {
	if mt.walker == nil {
		return nil, fmt.Errorf("walker cannot be nil")
	}

	// Build tree structure
	tree := &treeNode{
		Path:     "",
		IsFile:   false,
		Children: make(map[string]*treeNode),
	}

	// Walk the filesystem
	err := mt.walker.Walk(ctx, root, ignorePatterns, func(path string, info fs.FileInfo) error {
		// Get relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		// Normalize path separators
		relPath = filepath.ToSlash(relPath)

		// Add node to tree
		if info.IsDir() {
			mt.addDirectory(tree, relPath)
		} else {
			// Compute file hash
			hash, err := computeFileHash(path, root)
			if err != nil {
				return fmt.Errorf("failed to hash file %s: %w", path, err)
			}
			mt.addFile(tree, relPath, hash, info.Size())
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	// Compute directory hashes bottom-up
	mt.computeDirectoryHashes(tree)

	// Serialize tree state
	state := treeState{Root: tree}
	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tree state: %w", err)
	}

	return data, nil
}

// Diff compares two tree states and returns paths that changed.
func (mt *merkleTree) Diff(ctx context.Context, oldState, newState []byte) ([]string, error) {
	if len(oldState) == 0 || len(newState) == 0 {
		return nil, fmt.Errorf("both states must be non-empty")
	}

	// Deserialize both states
	var oldTree, newTree treeState
	if err := json.Unmarshal(oldState, &oldTree); err != nil {
		return nil, fmt.Errorf("failed to deserialize old state: %w", err)
	}
	if err := json.Unmarshal(newState, &newTree); err != nil {
		return nil, fmt.Errorf("failed to deserialize new state: %w", err)
	}

	// Compare trees and collect changed paths
	changes := make([]string, 0)
	mt.diffNodes(oldTree.Root, newTree.Root, &changes)

	return changes, nil
}

// addDirectory adds a directory node to the tree, creating parent directories as needed.
func (mt *merkleTree) addDirectory(root *treeNode, path string) {
	parts := strings.Split(path, "/")
	current := root

	for _, part := range parts {
		if _, exists := current.Children[part]; !exists {
			current.Children[part] = &treeNode{
				Path:     filepath.Join(current.Path, part),
				IsFile:   false,
				Children: make(map[string]*treeNode),
			}
		}
		current = current.Children[part]
	}
}

// addFile adds a file node to the tree, creating parent directories as needed.
func (mt *merkleTree) addFile(root *treeNode, path string, hash string, size int64) {
	parts := strings.Split(path, "/")
	current := root

	// Create parent directories
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, exists := current.Children[part]; !exists {
			current.Children[part] = &treeNode{
				Path:     filepath.Join(current.Path, part),
				IsFile:   false,
				Children: make(map[string]*treeNode),
			}
		}
		current = current.Children[part]
	}

	// Add file node
	fileName := parts[len(parts)-1]
	current.Children[fileName] = &treeNode{
		Path:     path,
		Hash:     hash,
		IsFile:   true,
		Size:     size,
		Children: nil,
	}
}

// computeDirectoryHashes computes hashes for directory nodes based on their children.
func (mt *merkleTree) computeDirectoryHashes(node *treeNode) string {
	if node.IsFile {
		return node.Hash
	}

	// Get sorted list of child names for deterministic hashing
	childNames := make([]string, 0, len(node.Children))
	for name := range node.Children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)

	// Compute hash from children
	h := sha256.New()
	for _, name := range childNames {
		child := node.Children[name]
		childHash := mt.computeDirectoryHashes(child)
		fmt.Fprintf(h, "%s:%s\n", name, childHash)
	}

	node.Hash = hex.EncodeToString(h.Sum(nil))
	return node.Hash
}

// diffNodes recursively compares two nodes and accumulates changed paths.
func (mt *merkleTree) diffNodes(oldNode, newNode *treeNode, changes *[]string) {
	// Handle nil nodes
	if oldNode == nil && newNode == nil {
		return
	}

	if oldNode == nil {
		// New node added
		mt.collectAllPaths(newNode, changes)
		return
	}

	if newNode == nil {
		// Node deleted
		mt.collectAllPaths(oldNode, changes)
		return
	}

	// Both nodes exist - check if changed
	if oldNode.IsFile && newNode.IsFile {
		// File comparison
		if oldNode.Hash != newNode.Hash {
			*changes = append(*changes, newNode.Path)
		}
		return
	}

	if oldNode.IsFile != newNode.IsFile {
		// Type changed (file <-> directory)
		*changes = append(*changes, oldNode.Path)
		return
	}

	// Both are directories - compare children
	if oldNode.Hash == newNode.Hash {
		// Hashes match - no changes in this subtree
		return
	}

	// Get all child names
	allChildren := make(map[string]bool)
	for name := range oldNode.Children {
		allChildren[name] = true
	}
	for name := range newNode.Children {
		allChildren[name] = true
	}

	// Compare children recursively
	for name := range allChildren {
		oldChild := oldNode.Children[name]
		newChild := newNode.Children[name]
		mt.diffNodes(oldChild, newChild, changes)
	}
}

// collectAllPaths collects all file paths in a subtree.
func (mt *merkleTree) collectAllPaths(node *treeNode, paths *[]string) {
	if node == nil {
		return
	}

	if node.IsFile {
		*paths = append(*paths, node.Path)
		return
	}

	// Recursively collect from children
	for _, child := range node.Children {
		mt.collectAllPaths(child, paths)
	}
}

// computeFileHash computes the SHA256 hash of a file's contents.
func computeFileHash(path string, basePath string) (string, error) {
	// G304: Path validation to prevent path traversal
	if _, err := security.ValidatePathWithinBase(path, basePath); err != nil {
		if errors.Is(err, security.ErrPathTraversal) {
			return "", fmt.Errorf("security: path traversal detected for %s: %w", path, err)
		}
		return "", fmt.Errorf("security: invalid path %s: %w", path, err)
	}

	// #nosec G304 - Path validated at line 279 with ValidatePathWithinBase
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
