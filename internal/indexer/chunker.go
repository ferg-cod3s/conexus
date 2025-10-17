package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// CodeChunker implements semantic code chunking for various programming languages.
type CodeChunker struct {
	maxChunkSize int // Maximum characters per chunk
	overlapSize  int // Characters to overlap between chunks
}

// NewCodeChunker creates a new code chunker with configurable sizes.
func NewCodeChunker(maxChunkSize, overlapSize int) *CodeChunker {
	if maxChunkSize <= 0 {
		maxChunkSize = 2000 // Default
	}
	if overlapSize < 0 {
		overlapSize = 200 // Default
	}
	return &CodeChunker{
		maxChunkSize: maxChunkSize,
		overlapSize:  overlapSize,
	}
}

// Supports returns true if this chunker handles the given file extension.
func (c *CodeChunker) Supports(fileExtension string) bool {
	supported := map[string]bool{
		".go":    true,
		".py":    true,
		".js":    true,
		".jsx":   true,
		".ts":    true,
		".tsx":   true,
		".java":  true,
		".cpp":   true,
		".cc":    true,
		".cxx":   true,
		".c++":   true,
		".c":     true,
		".rs":    true,
		".rb":    true,
		".php":   true,
		".cs":    true,
		".scala": true,
		".kt":    true,
		".swift": true,
	}
	return supported[strings.ToLower(fileExtension)]
}

// Chunk splits code content into semantic chunks based on language-specific constructs.
func (c *CodeChunker) Chunk(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		return c.chunkGoCode(ctx, content, filePath)
	case ".py":
		return c.chunkPythonCode(ctx, content, filePath)
	case ".js", ".jsx", ".ts", ".tsx":
		return c.chunkJavaScriptCode(ctx, content, filePath)
	case ".java":
		return c.chunkJavaCode(ctx, content, filePath)
	case ".cpp", ".cc", ".cxx", ".c++", ".c":
		return c.chunkCCode(ctx, content, filePath)
	case ".rs":
		return c.chunkRustCode(ctx, content, filePath)
	default:
		// Fallback to generic code chunking
		return c.chunkGenericCode(ctx, content, filePath)
	}
}

// chunkGoCode implements semantic chunking for Go code using AST parsing.
func (c *CodeChunker) chunkGoCode(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		// If parsing fails, fall back to generic chunking
		return c.chunkGenericCode(ctx, content, filePath)
	}

	var chunks []Chunk

	// Extract function declarations
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			startPos := fset.Position(fn.Pos())
			endPos := fset.Position(fn.End())

			// Extract function content
			lines := strings.Split(content, "\n")
			fnContent := strings.Join(lines[startPos.Line-1:endPos.Line], "\n")

			chunk := Chunk{
				ID:        generateChunkID(filePath, "function", fn.Name.Name, startPos.Line),
				Content:   fnContent,
				FilePath:  filePath,
				Language:  "go",
				Type:      ChunkTypeFunction,
				StartLine: startPos.Line,
				EndLine:   endPos.Line - 1,
				Metadata: map[string]string{
					"function_name": fn.Name.Name,
					"receiver":      c.getReceiverName(fn),
				},
				Hash:      generateContentHash(fnContent),
				IndexedAt: time.Now(),
			}
			chunks = append(chunks, chunk)
		}
	}

	// Extract struct/type declarations
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						startPos := fset.Position(typeSpec.Pos())
						endPos := fset.Position(typeSpec.End())

						lines := strings.Split(content, "\n")
						structContent := strings.Join(lines[startPos.Line-1:endPos.Line-1], "\n")

						chunk := Chunk{
							ID:        generateChunkID(filePath, "struct", typeSpec.Name.Name, startPos.Line),
							Content:   structContent,
							FilePath:  filePath,
							Language:  "go",
							Type:      ChunkTypeStruct,
							StartLine: startPos.Line,
							EndLine:   endPos.Line - 1,
							Metadata: map[string]string{
								"struct_name": typeSpec.Name.Name,
							},
							Hash:      generateContentHash(structContent),
							IndexedAt: time.Now(),
						}
						chunks = append(chunks, chunk)
					}
				}
			}
		}
	}

	// If no semantic chunks found, fall back to generic chunking
	if len(chunks) == 0 {
		return c.chunkGenericCode(ctx, content, filePath)
	}

	return chunks, nil
}

// chunkPythonCode implements semantic chunking for Python code.
func (c *CodeChunker) chunkPythonCode(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []Chunk

	// Python function/class detection using regex
	fnRegex := regexp.MustCompile(`^\s*def\s+(\w+)\s*\(`)
	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)

	currentChunk := ""
	currentType := ChunkTypeUnknown
	currentStartLine := 1
	currentName := ""
	braceCount := 0

	for i, line := range lines {
		lineNum := i + 1

		// Count braces to track function/class boundaries
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		// Check for function definition
		if fnMatch := fnRegex.FindStringSubmatch(line); fnMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeFunction
			currentStartLine = lineNum
			currentName = fnMatch[1]
			if currentName == "" {
				currentName = fnMatch[2] // arrow function or const function
			}
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if classMatch := classRegex.FindStringSubmatch(line); classMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeClass
			currentStartLine = lineNum
			currentName = classMatch[1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if currentChunk != "" {
			currentChunk += line + "\n"

			// End chunk when braces balance out
			if braceCount <= 0 && strings.TrimSpace(line) != "" {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum, currentName))
				currentChunk = ""
				currentType = ChunkTypeUnknown
				currentName = ""
			}
		}
	}

	// Save final chunk
	if currentChunk != "" {
		chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, len(lines), currentName))
	}

	if len(chunks) == 0 {
		return c.chunkGenericCode(ctx, content, filePath)
	}

	return chunks, nil
}

// chunkJavaScriptCode implements semantic chunking for JavaScript/TypeScript code.
func (c *CodeChunker) chunkJavaScriptCode(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []Chunk

	// JavaScript function/class detection
	fnRegex := regexp.MustCompile(`^\s*(?:function\s+(\w+)|(?:const|let|var)\s+(\w+)\s*=\s*(?:\([^)]*\)\s*=>|function))`)
	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)

	currentChunk := ""
	currentType := ChunkTypeUnknown
	currentStartLine := 1
	currentName := ""
	braceCount := 0

	for i, line := range lines {
		lineNum := i + 1

		// Count braces to track function/class boundaries
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		// Check for function definition
		if fnMatch := fnRegex.FindStringSubmatch(line); fnMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeFunction
			currentStartLine = lineNum
			currentName = fnMatch[1]
			if currentName == "" {
				currentName = fnMatch[2] // arrow function or const function
			}
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if classMatch := classRegex.FindStringSubmatch(line); classMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeClass
			currentStartLine = lineNum
			currentName = classMatch[1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if currentChunk != "" {
			currentChunk += line + "\n"

			// End chunk when braces balance out
			if braceCount <= 0 && strings.TrimSpace(line) != "" {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum, currentName))
				currentChunk = ""
				currentType = ChunkTypeUnknown
				currentName = ""
			}
		}
	}

	// Save final chunk
	if currentChunk != "" {
		chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, len(lines), currentName))
	}

	if len(chunks) == 0 {
		return c.chunkGenericCode(ctx, content, filePath)
	}

	return chunks, nil
}

// chunkJavaCode implements semantic chunking for Java code.
func (c *CodeChunker) chunkJavaCode(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []Chunk

	// Java method/class detection
	methodRegex := regexp.MustCompile(`^\s*(?:public|private|protected)?\s*(?:static)?\s*(?:\w+\s+)+\s*(\w+)\s*\(`)
	classRegex := regexp.MustCompile(`^\s*(?:public|private|protected)?\s*class\s+(\w+)`)

	currentChunk := ""
	currentType := ChunkTypeUnknown
	currentStartLine := 1
	currentName := ""
	braceCount := 0

	for i, line := range lines {
		lineNum := i + 1

		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		if methodMatch := methodRegex.FindStringSubmatch(line); methodMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "java", currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeFunction
			currentStartLine = lineNum
			currentName = methodMatch[len(methodMatch)-1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if classMatch := classRegex.FindStringSubmatch(line); classMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "java", currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeClass
			currentStartLine = lineNum
			currentName = classMatch[1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if currentChunk != "" {
			currentChunk += line + "\n"

			if braceCount <= 0 && strings.TrimSpace(line) != "" {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "java", currentType, currentStartLine, lineNum, currentName))
				currentChunk = ""
				currentType = ChunkTypeUnknown
				currentName = ""
			}
		}
	}

	// Save final chunk
	if currentChunk != "" {
		chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "java", currentType, currentStartLine, len(lines), currentName))
	}

	if len(chunks) == 0 {
		return c.chunkGenericCode(ctx, content, filePath)
	}

	return chunks, nil
}

// chunkCCode implements semantic chunking for C/C++ code.
func (c *CodeChunker) chunkCCode(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []Chunk

	// C/C++ function detection
	fnRegex := regexp.MustCompile(`^\s*(?:\w+\s+)+\s*\**\s*(\w+)\s*\(`)

	currentChunk := ""
	currentType := ChunkTypeFunction
	currentStartLine := 1
	currentName := ""
	braceCount := 0

	for i, line := range lines {
		lineNum := i + 1

		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		if fnMatch := fnRegex.FindStringSubmatch(line); fnMatch != nil && !strings.Contains(line, ";") {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentStartLine = lineNum
			currentName = fnMatch[len(fnMatch)-1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if currentChunk != "" {
			currentChunk += line + "\n"

			if braceCount <= 0 && strings.TrimSpace(line) != "" && strings.HasSuffix(strings.TrimSpace(line), "}") {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, lineNum, currentName))
				currentChunk = ""
				currentName = ""
			}
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, detectLanguage(filePath), currentType, currentStartLine, len(lines), currentName))
	}

	if len(chunks) == 0 {
		return c.chunkGenericCode(ctx, content, filePath)
	}

	return chunks, nil
}

// chunkRustCode implements semantic chunking for Rust code.
func (c *CodeChunker) chunkRustCode(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	lines := strings.Split(content, "\n")
	var chunks []Chunk

	// Rust function/struct/impl detection
	fnRegex := regexp.MustCompile(`^\s*fn\s+(\w+)\s*\(`)
	structRegex := regexp.MustCompile(`^\s*struct\s+(\w+)`)
	implRegex := regexp.MustCompile(`^\s*impl\s+(?:\w+::)?(\w+)`)

	currentChunk := ""
	currentType := ChunkTypeUnknown
	currentStartLine := 1
	currentName := ""
	braceCount := 0

	for i, line := range lines {
		lineNum := i + 1

		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		if fnMatch := fnRegex.FindStringSubmatch(line); fnMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "rust", currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeFunction
			currentStartLine = lineNum
			currentName = fnMatch[1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if structMatch := structRegex.FindStringSubmatch(line); structMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "rust", currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeStruct
			currentStartLine = lineNum
			currentName = structMatch[1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if implMatch := implRegex.FindStringSubmatch(line); implMatch != nil {
			if currentChunk != "" && braceCount <= 0 {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "rust", currentType, currentStartLine, lineNum-1, currentName))
			}

			currentChunk = line + "\n"
			currentType = ChunkTypeInterface
			currentStartLine = lineNum
			currentName = implMatch[1]
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")

		} else if currentChunk != "" {
			currentChunk += line + "\n"

			if braceCount <= 0 && strings.TrimSpace(line) != "" {
				chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "rust", currentType, currentStartLine, lineNum, currentName))
				currentChunk = ""
				currentType = ChunkTypeUnknown
				currentName = ""
			}
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, c.createCodeChunk(currentChunk, filePath, "rust", currentType, currentStartLine, len(lines), currentName))
	}

	if len(chunks) == 0 {
		return c.chunkGenericCode(ctx, content, filePath)
	}

	return chunks, nil
}

// chunkGenericCode implements fallback chunking for unsupported languages.
func (c *CodeChunker) chunkGenericCode(ctx context.Context, content string, filePath string) ([]Chunk, error) {
	if len(content) <= c.maxChunkSize {
		// Single chunk for small files
		return []Chunk{c.createCodeChunk(content, filePath, detectLanguage(filePath), ChunkTypeUnknown, 1, countLines(content), "")}, nil
	}

	var chunks []Chunk
	runes := []rune(content)
	totalLen := len(runes)

	for start := 0; start < totalLen; start += c.maxChunkSize - c.overlapSize {
		end := start + c.maxChunkSize
		if end > totalLen {
			end = totalLen
		}

		// Adjust end to avoid splitting words
		if end < totalLen {
			for end > start && !unicode.IsSpace(runes[end-1]) {
				end--
			}
		}

		chunkContent := string(runes[start:end])
		if strings.TrimSpace(chunkContent) == "" {
			continue
		}

		// Calculate line numbers
		startLine := 1
		endLine := countLines(chunkContent)
		if start > 0 {
			prefix := string(runes[:start])
			startLine = countLines(prefix) + 1
			endLine = startLine + countLines(chunkContent) - 1
		}

		chunk := c.createCodeChunk(chunkContent, filePath, detectLanguage(filePath), ChunkTypeUnknown, startLine, endLine, "")
		chunks = append(chunks, chunk)

		// Break if we've reached the end
		if end >= totalLen {
			break
		}
	}

	return chunks, nil
}

// createCodeChunk creates a chunk with the given parameters.
func (c *CodeChunker) createCodeChunk(content, filePath, language string, chunkType ChunkType, startLine, endLine int, name string) Chunk {
	metadata := make(map[string]string)
	if name != "" {
		switch chunkType {
		case ChunkTypeFunction:
			metadata["function_name"] = name
		case ChunkTypeClass, ChunkTypeStruct:
			metadata["type_name"] = name
		case ChunkTypeInterface:
			metadata["interface_name"] = name
		}
	}

	return Chunk{
		ID:        generateChunkID(filePath, string(chunkType), name, startLine),
		Content:   content,
		FilePath:  filePath,
		Language:  language,
		Type:      chunkType,
		StartLine: startLine,
		EndLine:   endLine,
		Metadata:  metadata,
		Hash:      generateContentHash(content),
		IndexedAt: time.Now(),
	}
}

// getReceiverName extracts the receiver name from a Go function declaration.
func (c *CodeChunker) getReceiverName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	recv := fn.Recv.List[0]
	if recv.Type != nil {
		if ident, ok := recv.Type.(*ast.Ident); ok {
			return ident.Name
		}
		if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				return ident.Name
			}
		}
	}
	return ""
}

// generateChunkID creates a unique identifier for a chunk.
func generateChunkID(filePath, chunkType, name string, line int) string {
	return fmt.Sprintf("%s:%s:%s:%d", filePath, chunkType, name, line)
}

// generateContentHash creates a hash of the content for deduplication.
func generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
