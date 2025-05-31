package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// JavaScriptParser implements the LanguageParser interface for JavaScript code
type JavaScriptParser struct{}

// ExtractImports extracts import statements from JavaScript file content
func (p *JavaScriptParser) ExtractImports(fileContent string) []string {
	var imports []string

	// Match ES6 imports
	es6ImportRegex := regexp.MustCompile(`import\s+(?:{[^}]*}|[^{;]*)\s+from\s+['"]([^'"]+)['"]`)
	es6ImportMatches := es6ImportRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range es6ImportMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Match CommonJS require
	requireRegex := regexp.MustCompile(`(?:const|let|var)\s+(?:{[^}]*}|[^=]*)\s*=\s*require\s*\(['"]([^'"]+)['"]\)`)
	requireMatches := requireRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range requireMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Match direct require calls
	directRequireRegex := regexp.MustCompile(`require\s*\(['"]([^'"]+)['"]\)`)
	directRequireMatches := directRequireRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range directRequireMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Match dynamic imports
	dynamicImportRegex := regexp.MustCompile(`import\s*\(['"]([^'"]+)['"]\)`)
	dynamicImportMatches := dynamicImportRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range dynamicImportMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	return imports
}

// ExtractFunctions extracts functions from JavaScript file content
func (p *JavaScriptParser) ExtractFunctions(fileContent string) []Function {
	var functions []Function

	// Match named function declarations
	namedFuncRegex := regexp.MustCompile(`function\s+([a-zA-Z0-9_$]+)\s*\([^)]*\)\s*{((?:[^{}]|{[^{}]*})*)}`)
	namedFuncMatches := namedFuncRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range namedFuncMatches {
		if len(match) >= 3 {
			functions = append(functions, Function{
				Name: match[1],
				Body: match[2],
			})
		}
	}

	// Match arrow functions with explicit names (const/let/var assignments)
	arrowFuncRegex := regexp.MustCompile(`(?:const|let|var)\s+([a-zA-Z0-9_$]+)\s*=\s*(?:\([^)]*\)|[^=]*)\s*=>\s*{((?:[^{}]|{[^{}]*})*)}`)
	arrowFuncMatches := arrowFuncRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range arrowFuncMatches {
		if len(match) >= 3 {
			functions = append(functions, Function{
				Name: match[1],
				Body: match[2],
			})
		}
	}

	// Match class methods
	methodRegex := regexp.MustCompile(`(?:async\s+)?([a-zA-Z0-9_$]+)\s*\([^)]*\)\s*{((?:[^{}]|{[^{}]*})*)}`)
	classRegex := regexp.MustCompile(`class\s+([a-zA-Z0-9_$]+)(?:\s+extends\s+[a-zA-Z0-9_$.]+)?\s*{((?:[^{}]|{[^{}]*})*)}`)
	classMatches := classRegex.FindAllStringSubmatch(fileContent, -1)

	for _, classMatch := range classMatches {
		if len(classMatch) >= 3 {
			className := classMatch[1]
			classBody := classMatch[2]

			methodMatches := methodRegex.FindAllStringSubmatch(classBody, -1)
			for _, methodMatch := range methodMatches {
				if len(methodMatch) >= 3 {
					methodName := methodMatch[1]
					if methodName != "constructor" {
						functions = append(functions, Function{
							Name: className + "." + methodName,
							Body: methodMatch[2],
						})
					} else {
						functions = append(functions, Function{
							Name: className + ".constructor",
							Body: methodMatch[2],
						})
					}
				}
			}
		}
	}

	return functions
}

// ExtractFunctionCalls extracts function calls from a function body
func (p *JavaScriptParser) ExtractFunctionCalls(functionBody string) []string {
	var calls []string
	seen := make(map[string]bool)

	// Match function calls
	callRegex := regexp.MustCompile(`([a-zA-Z0-9_$]+(?:\.[a-zA-Z0-9_$]+)*)\s*\(`)
	matches := callRegex.FindAllStringSubmatch(functionBody, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			funcName := match[1]

			// Skip common JavaScript built-ins
			if isJSBuiltin(funcName) {
				continue
			}

			if !seen[funcName] {
				calls = append(calls, funcName)
				seen[funcName] = true
			}
		}
	}

	return calls
}

// ResolveImportPath resolves a JavaScript import statement to a file path
func (p *JavaScriptParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := filepath.Dir(currentFilePath)

		// Try with extensions
		for _, ext := range []string{"", ".js", ".jsx", ".ts", ".tsx"} {
			fullPath := filepath.Join(dir, importPath+ext)
			if _, err := filepath.Abs(fullPath); err == nil {
				return fullPath
			}
		}

		// Try as directory with index file
		for _, indexFile := range []string{"index.js", "index.jsx", "index.ts", "index.tsx"} {
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
func (p *JavaScriptParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	patterns := []string{
		`function\s+` + regexp.QuoteMeta(functionName) + `\s*\(`,
		`(?:const|let|var)\s+` + regexp.QuoteMeta(functionName) + `\s*=\s*(?:function|\([^)]*\)\s*=>)`,
	}

	// Handle class methods
	if strings.Contains(functionName, ".") {
		parts := strings.Split(functionName, ".")
		if len(parts) == 2 {
			className := parts[0]
			methodName := parts[1]

			// Find the class first
			classPattern := `class\s+` + regexp.QuoteMeta(className) + `(?:\s+extends\s+[a-zA-Z0-9_$.]+)?\s*{`
			classScanner := bufio.NewScanner(strings.NewReader(fileContent))
			classLineNumber := 1
			classFound := false
			classEndLine := 0

			for classScanner.Scan() {
				line := classScanner.Text()
				if regexp.MustCompile(classPattern).MatchString(line) {
					classFound = true
					break
				}
				classLineNumber++
			}

			if classFound {
				// Find the class end
				braceCount := 1
				classBodyScanner := bufio.NewScanner(strings.NewReader(fileContent))
				for i := 0; i < classLineNumber; i++ {
					classBodyScanner.Scan()
				}

				currentLine := classLineNumber
				for classBodyScanner.Scan() {
					currentLine++
					line := classBodyScanner.Text()

					for _, char := range line {
						if char == '{' {
							braceCount++
						} else if char == '}' {
							braceCount--
							if braceCount == 0 {
								classEndLine = currentLine
								break
							}
						}
					}

					if classEndLine > 0 {
						break
					}
				}

				// Now find the method within the class
				methodPattern := `(?:async\s+)?` + regexp.QuoteMeta(methodName) + `\s*\(`
				methodScanner := bufio.NewScanner(strings.NewReader(fileContent))
				methodLineNumber := 1

				for i := 0; i < classLineNumber; i++ {
					methodScanner.Scan()
					methodLineNumber++
				}

				for methodScanner.Scan() {
					if methodLineNumber > classEndLine {
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

	// Regular function search
	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()

		for _, pattern := range patterns {
			if regexp.MustCompile(pattern).MatchString(line) {
				return lineNumber
			}
		}

		lineNumber++
	}

	return -1
}

// isJSBuiltin checks if a function name is a JavaScript built-in
func isJSBuiltin(name string) bool {
	builtins := map[string]bool{
		"console.log":        true,
		"parseInt":           true,
		"parseFloat":         true,
		"setTimeout":         true,
		"setInterval":        true,
		"clearTimeout":       true,
		"clearInterval":      true,
		"encodeURI":          true,
		"decodeURI":          true,
		"encodeURIComponent": true,
		"decodeURIComponent": true,
		"isNaN":              true,
		"isFinite":           true,
		"eval":               true,
		"alert":              true,
		"confirm":            true,
		"prompt":             true,
		"Math.abs":           true,
		"Math.ceil":          true,
		"Math.floor":         true,
		"Math.max":           true,
		"Math.min":           true,
		"Math.random":        true,
		"Math.round":         true,
		"JSON.parse":         true,
		"JSON.stringify":     true,
		"Object.keys":        true,
		"Object.values":      true,
		"Object.entries":     true,
		"Array.isArray":      true,
		"Date.now":           true,
	}

	return builtins[name]
}
