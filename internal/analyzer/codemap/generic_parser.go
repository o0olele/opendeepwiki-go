package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// GenericParser implements a basic parser for any language
type GenericParser struct{}

// ExtractImports attempts to extract import statements from file content
func (p *GenericParser) ExtractImports(fileContent string) []string {
	var imports []string

	// Try to match common import patterns across languages
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`import\s+["']([^"']+)["']`),
		regexp.MustCompile(`import\s+([a-zA-Z0-9_.]+)`),
		regexp.MustCompile(`#include\s+["<]([^">]+)[">]`),
		regexp.MustCompile(`require\s+["']([^"']+)["']`),
		regexp.MustCompile(`using\s+([a-zA-Z0-9_.]+)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(fileContent, -1)
		for _, match := range matches {
			if len(match) >= 2 && match[1] != "" {
				imports = append(imports, match[1])
			}
		}
	}

	return imports
}

// ExtractFunctions attempts to extract functions from file content
func (p *GenericParser) ExtractFunctions(fileContent string) []Function {
	var functions []Function

	// Try to match common function patterns across languages
	patterns := []*regexp.Regexp{
		// Function with name and body
		regexp.MustCompile(`(?:function|func|def|sub|void|public|private|protected|internal|static)\s+([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s*(?:->|:)\s*[a-zA-Z0-9_<>[\]]+)?\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`),
		// Method with name and body
		regexp.MustCompile(`(?:public|private|protected|internal|static|final|async|override|virtual|abstract)\s+(?:[a-zA-Z0-9_<>[\]]+\s+)?([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s*(?:->|:)\s*[a-zA-Z0-9_<>[\]]+)?\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(fileContent, -1)
		for _, match := range matches {
			if len(match) >= 3 && match[1] != "" {
				functions = append(functions, Function{
					Name: match[1],
					Body: match[2],
				})
			}
		}
	}

	return functions
}

// ExtractFunctionCalls attempts to extract function calls from a function body
func (p *GenericParser) ExtractFunctionCalls(functionBody string) []string {
	var calls []string
	seen := make(map[string]bool)

	// Try to match common function call patterns
	pattern := regexp.MustCompile(`([a-zA-Z0-9_]+)\s*\(`)

	matches := pattern.FindAllStringSubmatch(functionBody, -1)
	for _, match := range matches {
		if len(match) >= 2 && match[1] != "" {
			funcName := match[1]
			if !seen[funcName] {
				calls = append(calls, funcName)
				seen[funcName] = true
			}
		}
	}

	return calls
}

// ResolveImportPath attempts to resolve an import statement to a file path
func (p *GenericParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle relative imports
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		dir := filepath.Dir(currentFilePath)
		return filepath.Join(dir, importPath)
	}

	// For other imports, try to find them in the base path
	// This is a simplified approach and may not work for all languages
	return filepath.Join(basePath, importPath)
}

// GetFunctionLineNumber attempts to find the line number where a function starts
func (p *GenericParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	// Try to match common function declaration patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:function|func|def|sub|void|public|private|protected|internal|static)\s+` + regexp.QuoteMeta(functionName) + `\s*\(`),
		regexp.MustCompile(`(?:public|private|protected|internal|static|final|async|override|virtual|abstract)\s+(?:[a-zA-Z0-9_<>[\]]+\s+)?` + regexp.QuoteMeta(functionName) + `\s*\(`),
	}

	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				return lineNumber
			}
		}
		lineNumber++
	}

	return -1
}
