package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// TypeScriptParser implements the LanguageParser interface for TypeScript code
type TypeScriptParser struct {
	jsParser JavaScriptParser
}

// ExtractImports extracts import statements from TypeScript file content
func (p *TypeScriptParser) ExtractImports(fileContent string) []string {
	// TypeScript imports are similar to JavaScript ES6 imports
	imports := p.jsParser.ExtractImports(fileContent)

	// Also match type imports
	typeImportRegex := regexp.MustCompile(`import\s+type\s+(?:{[^}]*}|[^{;]*)\s+from\s+['"]([^'"]+)['"]`)
	typeImportMatches := typeImportRegex.FindAllStringSubmatch(fileContent, -1)

	for _, match := range typeImportMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	return imports
}

// ExtractFunctions extracts functions from TypeScript file content
func (p *TypeScriptParser) ExtractFunctions(fileContent string) []Function {
	// Start with JavaScript function extraction
	functions := p.jsParser.ExtractFunctions(fileContent)

	// Match TypeScript-specific function declarations with type annotations
	tsFuncRegex := regexp.MustCompile(`function\s+([a-zA-Z0-9_$]+)\s*<[^>]*>\s*\([^)]*\)(?:\s*:\s*[^{]+)?\s*{((?:[^{}]|{[^{}]*})*)}`)
	tsFuncMatches := tsFuncRegex.FindAllStringSubmatch(fileContent, -1)

	for _, match := range tsFuncMatches {
		if len(match) >= 3 {
			functions = append(functions, Function{
				Name: match[1],
				Body: match[2],
			})
		}
	}

	// Match arrow functions with type annotations
	tsArrowFuncRegex := regexp.MustCompile(`(?:const|let|var)\s+([a-zA-Z0-9_$]+)(?:\s*:\s*[^=]+)?\s*=\s*(?:\([^)]*\)|[^=]*)(?:\s*:\s*[^=]+)?\s*=>\s*{((?:[^{}]|{[^{}]*})*)}`)
	tsArrowFuncMatches := tsArrowFuncRegex.FindAllStringSubmatch(fileContent, -1)

	for _, match := range tsArrowFuncMatches {
		if len(match) >= 3 {
			functions = append(functions, Function{
				Name: match[1],
				Body: match[2],
			})
		}
	}

	// Match interface methods (for documentation purposes)
	interfaceRegex := regexp.MustCompile(`interface\s+([a-zA-Z0-9_$]+)(?:\s+extends\s+[a-zA-Z0-9_$.]+)?\s*{((?:[^{}]|{[^{}]*})*)}`)
	interfaceMatches := interfaceRegex.FindAllStringSubmatch(fileContent, -1)

	for _, interfaceMatch := range interfaceMatches {
		if len(interfaceMatch) >= 3 {
			interfaceName := interfaceMatch[1]
			interfaceBody := interfaceMatch[2]

			// Match method signatures in interfaces
			methodSigRegex := regexp.MustCompile(`([a-zA-Z0-9_$]+)\s*(?:<[^>]*>)?\s*\([^)]*\)\s*:\s*[^;]+;`)
			methodSigMatches := methodSigRegex.FindAllStringSubmatch(interfaceBody, -1)

			for _, methodMatch := range methodSigMatches {
				if len(methodMatch) >= 2 {
					methodName := methodMatch[1]
					functions = append(functions, Function{
						Name: interfaceName + "." + methodName + " (interface)",
						Body: "", // Interface methods don't have bodies
					})
				}
			}
		}
	}

	// Match type aliases and types (for documentation purposes)
	typeRegex := regexp.MustCompile(`type\s+([a-zA-Z0-9_$]+)(?:<[^>]*>)?\s*=\s*([^;]+);`)
	typeMatches := typeRegex.FindAllStringSubmatch(fileContent, -1)

	for _, typeMatch := range typeMatches {
		if len(typeMatch) >= 3 {
			typeName := typeMatch[1]
			typeBody := typeMatch[2]

			functions = append(functions, Function{
				Name: "type:" + typeName,
				Body: typeBody,
			})
		}
	}

	return functions
}

// ExtractFunctionCalls extracts function calls from a function body
func (p *TypeScriptParser) ExtractFunctionCalls(functionBody string) []string {
	// TypeScript function calls are similar to JavaScript
	return p.jsParser.ExtractFunctionCalls(functionBody)
}

// ResolveImportPath resolves a TypeScript import statement to a file path
func (p *TypeScriptParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := filepath.Dir(currentFilePath)

		// Try with TypeScript extensions first
		for _, ext := range []string{"", ".ts", ".tsx", ".d.ts", ".js", ".jsx"} {
			fullPath := filepath.Join(dir, importPath+ext)
			if _, err := filepath.Abs(fullPath); err == nil {
				return fullPath
			}
		}

		// Try as directory with index file
		for _, indexFile := range []string{"index.ts", "index.tsx", "index.js", "index.jsx"} {
			fullPath := filepath.Join(dir, importPath, indexFile)
			if _, err := filepath.Abs(fullPath); err == nil {
				return fullPath
			}
		}
	}

	// Handle node_modules or other absolute imports
	// This is a simplified approach and may not work for all cases
	return filepath.Join(basePath, "node_modules", importPath)
}

// GetFunctionLineNumber gets the line number where a function starts
func (p *TypeScriptParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	// For TypeScript-specific patterns
	tsPatterns := []string{
		`function\s+` + regexp.QuoteMeta(functionName) + `\s*<[^>]*>\s*\(`,
		`(?:const|let|var)\s+` + regexp.QuoteMeta(functionName) + `(?:\s*:\s*[^=]+)?\s*=\s*(?:function|\([^)]*\)(?:\s*:\s*[^=]+)?\s*=>)`,
	}

	// Handle type aliases
	if strings.HasPrefix(functionName, "type:") {
		typeName := strings.TrimPrefix(functionName, "type:")
		typePattern := `type\s+` + regexp.QuoteMeta(typeName) + `(?:<[^>]*>)?\s*=`

		scanner := bufio.NewScanner(strings.NewReader(fileContent))
		lineNumber := 1

		for scanner.Scan() {
			line := scanner.Text()
			if regexp.MustCompile(typePattern).MatchString(line) {
				return lineNumber
			}
			lineNumber++
		}
	}

	// Handle interface methods
	if strings.Contains(functionName, " (interface)") {
		parts := strings.Split(strings.TrimSuffix(functionName, " (interface)"), ".")
		if len(parts) == 2 {
			interfaceName := parts[0]
			methodName := parts[1]

			// Find the interface first
			interfacePattern := `interface\s+` + regexp.QuoteMeta(interfaceName) + `(?:\s+extends\s+[a-zA-Z0-9_$.]+)?\s*{`
			interfaceScanner := bufio.NewScanner(strings.NewReader(fileContent))
			interfaceLineNumber := 1
			interfaceFound := false
			interfaceEndLine := 0

			for interfaceScanner.Scan() {
				line := interfaceScanner.Text()
				if regexp.MustCompile(interfacePattern).MatchString(line) {
					interfaceFound = true
					break
				}
				interfaceLineNumber++
			}

			if interfaceFound {
				// Find the interface end
				braceCount := 1
				interfaceBodyScanner := bufio.NewScanner(strings.NewReader(fileContent))
				for i := 0; i < interfaceLineNumber; i++ {
					interfaceBodyScanner.Scan()
				}

				currentLine := interfaceLineNumber
				for interfaceBodyScanner.Scan() {
					currentLine++
					line := interfaceBodyScanner.Text()

					for _, char := range line {
						if char == '{' {
							braceCount++
						} else if char == '}' {
							braceCount--
							if braceCount == 0 {
								interfaceEndLine = currentLine
								break
							}
						}
					}

					if interfaceEndLine > 0 {
						break
					}
				}

				// Now find the method within the interface
				methodPattern := regexp.QuoteMeta(methodName) + `\s*(?:<[^>]*>)?\s*\(`
				methodScanner := bufio.NewScanner(strings.NewReader(fileContent))
				methodLineNumber := 1

				for i := 0; i < interfaceLineNumber; i++ {
					methodScanner.Scan()
					methodLineNumber++
				}

				for methodScanner.Scan() {
					if methodLineNumber > interfaceEndLine {
						break
					}

					line := methodScanner.Text()
					if regexp.MustCompile(methodPattern).MatchString(line) {
						return methodLineNumber
					}

					methodLineNumber++
				}
			}
		}
	}

	// Check TypeScript-specific patterns
	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()

		for _, pattern := range tsPatterns {
			if regexp.MustCompile(pattern).MatchString(line) {
				return lineNumber
			}
		}

		lineNumber++
	}

	// Fall back to JavaScript patterns
	return p.jsParser.GetFunctionLineNumber(fileContent, functionName)
}
