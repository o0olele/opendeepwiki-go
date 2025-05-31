package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// GoParser implements the LanguageParser interface for Go code
type GoParser struct{}

// ExtractImports extracts import statements from Go file content
func (p *GoParser) ExtractImports(fileContent string) []string {
	var imports []string

	// Match both single imports and import blocks
	singleImportRegex := regexp.MustCompile(`import\s+"([^"]+)"`)
	blockImportRegex := regexp.MustCompile(`import\s+\(\s*((?:.|\n)*?)\s*\)`)

	// Extract single imports
	matches := singleImportRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Extract block imports
	blockMatches := blockImportRegex.FindAllStringSubmatch(fileContent, -1)
	for _, blockMatch := range blockMatches {
		if len(blockMatch) >= 2 {
			importBlock := blockMatch[1]
			scanner := bufio.NewScanner(strings.NewReader(importBlock))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "//") {
					continue
				}

				// Handle named imports like "fmt" as f
				parts := strings.Fields(line)
				var importPath string
				if len(parts) == 1 {
					importPath = strings.Trim(parts[0], "\"")
				} else if len(parts) >= 2 {
					importPath = strings.Trim(parts[len(parts)-1], "\"")
				}

				if importPath != "" {
					imports = append(imports, importPath)
				}
			}
		}
	}

	return imports
}

// ExtractFunctions extracts functions from Go file content
func (p *GoParser) ExtractFunctions(fileContent string) []Function {
	var functions []Function

	// Match function declarations
	// This regex is simplified and may not catch all edge cases
	funcRegex := regexp.MustCompile(`func\s+(?:\([^)]+\)\s+)?([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s+[^{]+)?\s*{((?:[^{}]|{[^{}]*})*)}`)

	matches := funcRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			functions = append(functions, Function{
				Name: match[1],
				Body: match[2],
			})
		}
	}

	return functions
}

// ExtractFunctionCalls extracts function calls from a function body
func (p *GoParser) ExtractFunctionCalls(functionBody string) []string {
	var calls []string

	// Match function calls
	// This regex is simplified and may not catch all edge cases
	callRegex := regexp.MustCompile(`([a-zA-Z0-9_\.]+)\s*\(`)

	matches := callRegex.FindAllStringSubmatch(functionBody, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			funcName := match[1]
			// Filter out common Go built-ins
			if !isGoBuiltin(funcName) {
				calls = append(calls, funcName)
			}
		}
	}

	return calls
}

// ResolveImportPath resolves a Go import statement to a file path
func (p *GoParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle standard library imports
	if !strings.Contains(importPath, ".") && !strings.Contains(importPath, "/") {
		return "" // Standard library imports don't have a local file path
	}

	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := filepath.Dir(currentFilePath)
		return filepath.Join(dir, importPath)
	}

	// Handle absolute imports (within the module)
	// This is simplified and assumes the import is within the same module
	return filepath.Join(basePath, importPath)
}

// GetFunctionLineNumber gets the line number where a function starts
func (p *GoParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	pattern := regexp.MustCompile(`func\s+(?:\([^)]+\)\s+)?` + regexp.QuoteMeta(functionName) + `\s*\(`)
	scanner := bufio.NewScanner(strings.NewReader(fileContent))

	lineNumber := 1
	for scanner.Scan() {
		line := scanner.Text()
		if pattern.MatchString(line) {
			return lineNumber
		}
		lineNumber++
	}

	return -1
}

// isGoBuiltin checks if a function name is a Go built-in
func isGoBuiltin(name string) bool {
	builtins := map[string]bool{
		"append":  true,
		"cap":     true,
		"close":   true,
		"complex": true,
		"copy":    true,
		"delete":  true,
		"imag":    true,
		"len":     true,
		"make":    true,
		"new":     true,
		"panic":   true,
		"print":   true,
		"println": true,
		"real":    true,
		"recover": true,
	}

	return builtins[name]
}
