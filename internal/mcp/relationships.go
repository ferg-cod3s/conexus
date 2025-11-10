// Package mcp implements relationship detection logic for related files and code.
package mcp

import (
	"path/filepath"
	"strings"
)

// RelationType constants for different kinds of file relationships
const (
	RelationTypeSymbolRef     = "symbol_ref"     // Files with shared symbols/functions
	RelationTypeTestFile      = "test_file"      // Test files for implementation files
	RelationTypeImport        = "import"         // Files that import/are imported by target
	RelationTypeDocumentation = "documentation"  // Documentation files
	RelationTypeCommitHistory = "commit_history" // Files from related git commits
	RelationTypeSimilarCode   = "similar_code"   // Similar code patterns
	RelationTypeUnknown       = ""               // Unknown relationship
)

// RelationshipDetector provides methods to detect relationships between files
type RelationshipDetector struct {
	targetPath string
}

// NewRelationshipDetector creates a new detector for the given target file
func NewRelationshipDetector(targetPath string) *RelationshipDetector {
	return &RelationshipDetector{
		targetPath: targetPath,
	}
}

// DetectRelationType determines the relationship type between target and related file
func (rd *RelationshipDetector) DetectRelationType(relatedPath string, relatedType string, relatedMetadata map[string]interface{}) string {
	// Priority 1: Test file detection
	if rd.isTestRelationship(relatedPath) {
		return RelationTypeTestFile
	}

	// Priority 2: Documentation
	if rd.isDocumentation(relatedPath, relatedType) {
		return RelationTypeDocumentation
	}

	// Priority 3: Symbol references (from chunk type)
	if rd.hasSymbolReference(relatedType, relatedMetadata) {
		return RelationTypeSymbolRef
	}

	// Priority 4: Import relationships (implied from same language/package)
	if rd.isPotentialImport(relatedPath, relatedType, relatedMetadata) {
		return RelationTypeImport
	}

	// Priority 5: Similar code (fallback for code files)
	if rd.isSimilarCode(relatedPath, relatedType) {
		return RelationTypeSimilarCode
	}

	return RelationTypeUnknown
}

// isTestRelationship checks if the related file is a test file for the target
func (rd *RelationshipDetector) isTestRelationship(relatedPath string) bool {
	if rd.targetPath == "" || relatedPath == "" {
		return false
	}

	targetExt := strings.ToLower(filepath.Ext(rd.targetPath))
	relatedExt := strings.ToLower(filepath.Ext(relatedPath))
	targetBase := strings.TrimSuffix(filepath.Base(rd.targetPath), filepath.Ext(rd.targetPath))
	relatedBase := strings.TrimSuffix(filepath.Base(relatedPath), filepath.Ext(relatedPath))

	// Go test files: foo.go <-> foo_test.go (case-insensitive)
	if targetExt == ".go" {
		targetBaseLower := strings.ToLower(targetBase)
		relatedBaseLower := strings.ToLower(relatedBase)
		if strings.HasSuffix(relatedBaseLower, "_test") && targetBaseLower == strings.TrimSuffix(relatedBaseLower, "_test") {
			return true
		}
		if strings.HasSuffix(targetBaseLower, "_test") && relatedBaseLower == strings.TrimSuffix(targetBaseLower, "_test") {
			return true
		}
	}

	// Java/Kotlin test files: Foo.java <-> FooTest.java or TestFoo.java (case-insensitive)
	if targetExt == ".java" || targetExt == ".kt" || relatedExt == ".java" || relatedExt == ".kt" {
		targetBaseLower := strings.ToLower(targetBase)
		relatedBaseLower := strings.ToLower(relatedBase)
		if strings.HasSuffix(relatedBaseLower, "test") && targetBaseLower == strings.TrimSuffix(relatedBaseLower, "test") {
			return true
		}
		if strings.HasPrefix(relatedBaseLower, "test") && targetBaseLower == strings.TrimPrefix(relatedBaseLower, "test") {
			return true
		}
		if strings.HasSuffix(targetBaseLower, "test") && relatedBaseLower == strings.TrimSuffix(targetBaseLower, "test") {
			return true
		}
		if strings.HasPrefix(targetBaseLower, "test") && relatedBaseLower == strings.TrimPrefix(targetBaseLower, "test") {
			return true
		}
	}

	// Python test files: foo.py <-> test_foo.py or foo_test.py (case-insensitive)
	if targetExt == ".py" || relatedExt == ".py" {
		targetBaseLower := strings.ToLower(targetBase)
		relatedBaseLower := strings.ToLower(relatedBase)
		if strings.HasPrefix(relatedBaseLower, "test_") && targetBaseLower == strings.TrimPrefix(relatedBaseLower, "test_") {
			return true
		}
		if strings.HasSuffix(relatedBaseLower, "_test") && targetBaseLower == strings.TrimSuffix(relatedBaseLower, "_test") {
			return true
		}
		if strings.HasPrefix(targetBaseLower, "test_") && relatedBaseLower == strings.TrimPrefix(targetBaseLower, "test_") {
			return true
		}
		if strings.HasSuffix(targetBaseLower, "_test") && relatedBaseLower == strings.TrimSuffix(targetBaseLower, "_test") {
			return true
		}
	}

	// JavaScript/TypeScript test files: foo.js <-> foo.test.js, foo.spec.js
	if isJSOrTS(targetExt) || isJSOrTS(relatedExt) {
		// Check the full base name (without extension) for test patterns
		targetFullBase := filepath.Base(rd.targetPath)
		relatedFullBase := filepath.Base(relatedPath)
		
		// Remove extensions
		targetFullBase = strings.TrimSuffix(targetFullBase, filepath.Ext(targetFullBase))
		relatedFullBase = strings.TrimSuffix(relatedFullBase, filepath.Ext(relatedFullBase))
		
		// Now check if one has .test. or .spec. and the other doesn't
		targetHasTest := strings.Contains(targetFullBase, ".test") || strings.Contains(targetFullBase, ".spec")
		relatedHasTest := strings.Contains(relatedFullBase, ".test") || strings.Contains(relatedFullBase, ".spec")
		
		// One should have test marker, one should not
		if targetHasTest != relatedHasTest {
			// Clean both and compare
			targetClean := cleanJSFileName(targetFullBase)
			relatedClean := cleanJSFileName(relatedFullBase)
			if strings.ToLower(targetClean) == strings.ToLower(relatedClean) {
				return true
			}
		}
	}

	// Rust test files: often in tests/ subdirectory with same basename
	if (targetExt == ".rs" || relatedExt == ".rs") && 
	   (strings.Contains(relatedPath, "/tests/") || 
	    strings.HasPrefix(relatedPath, "tests/")) {
		// Also verify the base filenames match
		targetBaseLower := strings.ToLower(targetBase)
		relatedBaseLower := strings.ToLower(relatedBase)
		if targetBaseLower == relatedBaseLower {
			return true
		}
	}

	return false
}


// isDocumentation checks if the related file is documentation
func (rd *RelationshipDetector) isDocumentation(relatedPath string, chunkType string) bool {
	relatedExt := strings.ToLower(filepath.Ext(relatedPath))
	relatedDir := filepath.Dir(relatedPath)

	// Check for documentation file extensions
	if relatedExt == ".md" || relatedExt == ".rst" || relatedExt == ".txt" ||
		relatedExt == ".adoc" || relatedExt == ".asciidoc" {
		return true
	}

	// Check for documentation directories
	if strings.Contains(strings.ToLower(relatedDir), "docs") ||
		strings.Contains(strings.ToLower(relatedDir), "documentation") ||
		strings.Contains(strings.ToLower(relatedDir), "wiki") {
		return true
	}

	// Check for README files
	if strings.HasPrefix(strings.ToUpper(filepath.Base(relatedPath)), "README") {
		return true
	}

	return false
}

// hasSymbolReference checks if the chunk contains function/class/struct definitions
func (rd *RelationshipDetector) hasSymbolReference(chunkType string, metadata map[string]interface{}) bool {
	// Check chunk type for semantic symbols
	if chunkType == "function" || chunkType == "class" || chunkType == "struct" ||
		chunkType == "interface" || chunkType == "method" {
		return true
	}

	// Check metadata for symbol information
	if _, ok := metadata["symbol_name"]; ok {
		return true
	}

	return false
}

// isPotentialImport checks if files might have import relationships
func (rd *RelationshipDetector) isPotentialImport(relatedPath string, chunkType string, metadata map[string]interface{}) bool {
	if rd.targetPath == "" || relatedPath == "" {
		return false
	}

	targetExt := strings.ToLower(filepath.Ext(rd.targetPath))
	relatedExt := strings.ToLower(filepath.Ext(relatedPath))

	// Different file types can't import each other
	if targetExt != relatedExt {
		return false
	}

	// Same directory often means import relationships
	targetDir := filepath.Dir(rd.targetPath)
	relatedDir := filepath.Dir(relatedPath)

	// Same package/module directory
	if targetDir == relatedDir {
		return true
	}

	// Go: same package structure
	if targetExt == ".go" {
		if strings.HasPrefix(relatedDir, targetDir) || strings.HasPrefix(targetDir, relatedDir) {
			return true
		}
	}

	// Python: same module hierarchy
	if targetExt == ".py" {
		if strings.HasPrefix(relatedDir, targetDir) || strings.HasPrefix(targetDir, relatedDir) {
			return true
		}
	}

	// JavaScript/TypeScript: same module structure
	if isJSOrTS(targetExt) {
		if strings.HasPrefix(relatedDir, targetDir) || strings.HasPrefix(targetDir, relatedDir) {
			return true
		}
	}

	return false
}

// isSimilarCode checks if this is similar code (fallback category)
func (rd *RelationshipDetector) isSimilarCode(relatedPath string, chunkType string) bool {
	if rd.targetPath == "" || relatedPath == "" {
		return false
	}

	targetExt := strings.ToLower(filepath.Ext(rd.targetPath))
	relatedExt := strings.ToLower(filepath.Ext(relatedPath))

	// Same language is a weak signal for similar code
	if targetExt == relatedExt && isCodeFile(relatedExt) {
		return true
	}

	return false
}

// Helper functions

func isJSOrTS(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" ||
		ext == ".mjs" || ext == ".cjs"
}

func cleanJSFileName(basename string) string {
	// Remove .test, .spec, .min, etc. from basename
	basename = strings.ReplaceAll(basename, ".test", "")
	basename = strings.ReplaceAll(basename, ".spec", "")
	basename = strings.ReplaceAll(basename, ".min", "")
	return basename
}

func isCodeFile(ext string) bool {
	codeExts := map[string]bool{
		".go": true, ".java": true, ".py": true, ".js": true, ".ts": true,
		".jsx": true, ".tsx": true, ".c": true, ".cpp": true, ".h": true,
		".hpp": true, ".rs": true, ".rb": true, ".php": true, ".cs": true,
		".swift": true, ".kt": true, ".scala": true, ".clj": true,
	}
	return codeExts[strings.ToLower(ext)]
}
