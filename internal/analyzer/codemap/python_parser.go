package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// PythonParser implements the LanguageParser interface for Python code
type PythonParser struct{}

// ExtractImports extracts import statements from Python file content
func (p *PythonParser) ExtractImports(fileContent string) []string {
	var imports []string

	// Match standard imports
	importRegex := regexp.MustCompile(`import\s+([a-zA-Z0-9_.]+)`)
	importMatches := importRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range importMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Match from imports
	fromImportRegex := regexp.MustCompile(`from\s+([a-zA-Z0-9_.]+)\s+import`)
	fromImportMatches := fromImportRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range fromImportMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Match relative imports
	relativeImportRegex := regexp.MustCompile(`from\s+(\.\S+)\s+import`)
	relativeImportMatches := relativeImportRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range relativeImportMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	return imports
}

// ExtractFunctions extracts functions from Python file content
func (p *PythonParser) ExtractFunctions(fileContent string) []Function {
	var functions []Function

	// Match function declarations
	funcRegex := regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s*->\s*[^:]+)?:(?:\s*(?:"""(?:.|\n)*?"""|'''(?:.|\n)*?'''))?(?:\s*(?:#[^\n]*\n))*\s*((?:.|\n)*?)(?:(?:\ndef\s+)|(?:\nclass\s+)|$)`)
	funcMatches := funcRegex.FindAllStringSubmatch(fileContent, -1)

	for _, match := range funcMatches {
		if len(match) >= 3 {
			functions = append(functions, Function{
				Name: match[1],
				Body: strings.TrimSpace(match[2]),
			})
		}
	}

	// Match class methods
	classRegex := regexp.MustCompile(`class\s+([a-zA-Z0-9_]+)(?:\([^)]*\))?:(?:\s*(?:"""(?:.|\n)*?"""|'''(?:.|\n)*?'''))?(?:\s*(?:#[^\n]*\n))*\s*((?:.|\n)*?)(?:(?:\nclass\s+)|$)`)
	classMatches := classRegex.FindAllStringSubmatch(fileContent, -1)

	for _, classMatch := range classMatches {
		if len(classMatch) >= 3 {
			className := classMatch[1]
			classBody := classMatch[2]

			methodRegex := regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s*->\s*[^:]+)?:(?:\s*(?:"""(?:.|\n)*?"""|'''(?:.|\n)*?'''))?(?:\s*(?:#[^\n]*\n))*\s*((?:.|\n)*?)(?:(?:\n\s*def\s+)|(?:\n\s*class\s+)|$)`)
			methodMatches := methodRegex.FindAllStringSubmatch(classBody, -1)

			for _, methodMatch := range methodMatches {
				if len(methodMatch) >= 3 {
					methodName := methodMatch[1]
					functions = append(functions, Function{
						Name: className + "." + methodName,
						Body: strings.TrimSpace(methodMatch[2]),
					})
				}
			}
		}
	}

	return functions
}

// ExtractFunctionCalls extracts function calls from a function body
func (p *PythonParser) ExtractFunctionCalls(functionBody string) []string {
	var calls []string
	seen := make(map[string]bool)

	// Match function calls
	callRegex := regexp.MustCompile(`([a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)*)\s*\(`)
	matches := callRegex.FindAllStringSubmatch(functionBody, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			funcName := match[1]

			// Skip Python built-ins
			if isPythonBuiltin(funcName) {
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

// ResolveImportPath resolves a Python import statement to a file path
func (p *PythonParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle relative imports
	if strings.HasPrefix(importPath, ".") {
		dir := filepath.Dir(currentFilePath)

		// Count the number of dots to determine how many directories to go up
		dotCount := 0
		for _, char := range importPath {
			if char == '.' {
				dotCount++
			} else {
				break
			}
		}

		// Go up directories based on dot count
		for i := 1; i < dotCount; i++ {
			dir = filepath.Dir(dir)
		}

		// Remove the dots from the import path
		importPath = importPath[dotCount:]
		if importPath != "" && importPath[0] == '.' {
			importPath = importPath[1:]
		}

		// Replace dots with path separators
		importPath = strings.ReplaceAll(importPath, ".", string(filepath.Separator))

		// Try with .py extension
		fullPath := filepath.Join(dir, importPath+".py")
		if _, err := filepath.Abs(fullPath); err == nil {
			return fullPath
		}

		// Try as directory with __init__.py
		fullPath = filepath.Join(dir, importPath, "__init__.py")
		if _, err := filepath.Abs(fullPath); err == nil {
			return fullPath
		}

		return filepath.Join(dir, importPath)
	}

	// Handle absolute imports
	// Replace dots with path separators
	importPath = strings.ReplaceAll(importPath, ".", string(filepath.Separator))

	// Try with .py extension
	fullPath := filepath.Join(basePath, importPath+".py")
	if _, err := filepath.Abs(fullPath); err == nil {
		return fullPath
	}

	// Try as directory with __init__.py
	fullPath = filepath.Join(basePath, importPath, "__init__.py")
	if _, err := filepath.Abs(fullPath); err == nil {
		return fullPath
	}

	return filepath.Join(basePath, importPath)
}

// GetFunctionLineNumber gets the line number where a function starts
func (p *PythonParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	// Handle class methods
	if strings.Contains(functionName, ".") {
		parts := strings.Split(functionName, ".")
		if len(parts) == 2 {
			className := parts[0]
			methodName := parts[1]

			// Find the class first
			classPattern := `class\s+` + regexp.QuoteMeta(className) + `(?:\([^)]*\))?:`
			classScanner := bufio.NewScanner(strings.NewReader(fileContent))
			classLineNumber := 1
			classFound := false

			for classScanner.Scan() {
				line := classScanner.Text()
				if regexp.MustCompile(classPattern).MatchString(line) {
					classFound = true
					break
				}
				classLineNumber++
			}

			if classFound {
				// Now find the method within the class
				methodPattern := `def\s+` + regexp.QuoteMeta(methodName) + `\s*\(`
				methodScanner := bufio.NewScanner(strings.NewReader(fileContent))
				methodLineNumber := 1

				for methodScanner.Scan() {
					line := methodScanner.Text()
					if methodLineNumber > classLineNumber && regexp.MustCompile(methodPattern).MatchString(line) {
						return methodLineNumber
					}
					methodLineNumber++
				}
			}
		}
	}

	// Regular function search
	pattern := `def\s+` + regexp.QuoteMeta(functionName) + `\s*\(`
	scanner := bufio.NewScanner(strings.NewReader(fileContent))
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()
		if regexp.MustCompile(pattern).MatchString(line) {
			return lineNumber
		}
		lineNumber++
	}

	return -1
}

// isPythonBuiltin checks if a function name is a Python built-in
func isPythonBuiltin(name string) bool {
	builtins := map[string]bool{
		"print":     true,
		"len":       true,
		"range":     true,
		"enumerate": true,
		"zip":       true,
		"map":       true,
		"filter":    true,
		"sorted":    true,
		"reversed":  true,
		"list":      true,
		"dict":      true,
		"set":       true,
		"tuple":     true,
		"str":       true,
		"int":       true,
		"float":     true,
		"bool":      true,
		"sum":       true,
		"min":       true,
		"max":       true,
		"abs":       true,
		"all":       true,
		"any":       true,
		"open":      true,
		"input":     true,
		"super":     true,
	}

	return builtins[name]
}
