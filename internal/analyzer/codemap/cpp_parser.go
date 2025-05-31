package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// CppParser implements the LanguageParser interface for C/C++ code
type CppParser struct{}

// ExtractImports extracts import statements from C/C++ file content
func (p *CppParser) ExtractImports(fileContent string) []string {
	var imports []string

	// Match #include statements
	includeRegex := regexp.MustCompile(`#include\s+["<]([^">]+)[">]`)
	includeMatches := includeRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range includeMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	return imports
}

// ExtractFunctions extracts functions from C/C++ file content
func (p *CppParser) ExtractFunctions(fileContent string) []Function {
	var functions []Function

	// Match function declarations with bodies
	// This regex is simplified and may not catch all edge cases
	funcRegex := regexp.MustCompile(`(?:(?:static|inline|extern|virtual|explicit|friend|constexpr)\s+)*(?:[a-zA-Z0-9_:]+\s+)+([a-zA-Z0-9_:~]+)\s*\([^)]*\)(?:\s*const)?(?:\s*noexcept)?(?:\s*override)?(?:\s*final)?(?:\s*=\s*[^{;]+)?(?:\s*throw\s*\([^)]*\))?\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
	funcMatches := funcRegex.FindAllStringSubmatch(fileContent, -1)

	for _, match := range funcMatches {
		if len(match) >= 3 {
			functionName := match[1]

			// Check if it's a class method (contains ::)
			if strings.Contains(functionName, "::") {
				functions = append(functions, Function{
					Name: functionName,
					Body: match[2],
				})
			} else {
				// It's a standalone function
				functions = append(functions, Function{
					Name: functionName,
					Body: match[2],
				})
			}
		}
	}

	// Extract class/struct declarations to help with method resolution
	classRegex := regexp.MustCompile(`(?:class|struct)\s+([a-zA-Z0-9_]+)(?:\s*:\s*(?:public|protected|private)\s+[a-zA-Z0-9_:]+(?:\s*,\s*(?:public|protected|private)\s+[a-zA-Z0-9_:]+)*)?\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
	classMatches := classRegex.FindAllStringSubmatch(fileContent, -1)

	for _, classMatch := range classMatches {
		if len(classMatch) >= 3 {
			className := classMatch[1]
			classBody := classMatch[2]

			// Extract methods within the class
			methodRegex := regexp.MustCompile(`(?:(?:static|inline|virtual|explicit|friend|constexpr)\s+)*(?:[a-zA-Z0-9_:]+\s+)+([a-zA-Z0-9_~]+)\s*\([^)]*\)(?:\s*const)?(?:\s*noexcept)?(?:\s*override)?(?:\s*final)?(?:\s*=\s*[^{;]+)?(?:\s*throw\s*\([^)]*\))?\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
			methodMatches := methodRegex.FindAllStringSubmatch(classBody, -1)

			for _, methodMatch := range methodMatches {
				if len(methodMatch) >= 3 {
					methodName := methodMatch[1]

					// Check if it's a constructor or destructor
					if methodName == className || methodName == "~"+className {
						if methodName == className {
							functions = append(functions, Function{
								Name: className + "::constructor",
								Body: methodMatch[2],
							})
						} else {
							functions = append(functions, Function{
								Name: className + "::destructor",
								Body: methodMatch[2],
							})
						}
					} else {
						functions = append(functions, Function{
							Name: className + "::" + methodName,
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
func (p *CppParser) ExtractFunctionCalls(functionBody string) []string {
	var calls []string
	seen := make(map[string]bool)

	// Match function calls
	callRegex := regexp.MustCompile(`([a-zA-Z0-9_:]+)\s*\(`)
	matches := callRegex.FindAllStringSubmatch(functionBody, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			funcName := match[1]

			// Skip C++ built-ins and common patterns
			if isCppBuiltin(funcName) {
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

// ResolveImportPath resolves a C/C++ include statement to a file path
func (p *CppParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle system includes (those with angle brackets)
	if strings.HasPrefix(importPath, "<") && strings.HasSuffix(importPath, ">") {
		importPath = strings.TrimPrefix(importPath, "<")
		importPath = strings.TrimSuffix(importPath, ">")

		// System includes might be in standard locations, but we'll check in the project
		// This is a simplified approach and may not work for all cases
		return filepath.Join(basePath, "include", importPath)
	}

	// Handle relative includes (those with quotes)
	dir := filepath.Dir(currentFilePath)

	// Try direct path
	fullPath := filepath.Join(dir, importPath)
	if _, err := filepath.Abs(fullPath); err == nil {
		return fullPath
	}

	// Try in include directory
	fullPath = filepath.Join(basePath, "include", importPath)
	if _, err := filepath.Abs(fullPath); err == nil {
		return fullPath
	}

	return filepath.Join(dir, importPath)
}

// GetFunctionLineNumber gets the line number where a function starts
func (p *CppParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	// Handle class methods
	if strings.Contains(functionName, "::") {
		parts := strings.Split(functionName, "::")
		if len(parts) == 2 {
			className := parts[0]
			methodName := parts[1]

			// Special handling for constructors and destructors
			isConstructor := methodName == "constructor"
			isDestructor := methodName == "destructor"

			// Find the class first
			classPattern := `(?:class|struct)\s+` + regexp.QuoteMeta(className)
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
				// Now find the method within the class or in the global scope (for implementation)
				var methodPattern string
				if isConstructor {
					methodPattern = regexp.QuoteMeta(className) + `\s*\(`
				} else if isDestructor {
					methodPattern = `~` + regexp.QuoteMeta(className) + `\s*\(`
				} else {
					methodPattern = `(?:[a-zA-Z0-9_:]+\s+)+` + regexp.QuoteMeta(methodName) + `\s*\(`
				}

				// Look for method declaration in class or global implementation
				scanner := bufio.NewScanner(strings.NewReader(fileContent))
				lineNumber := 1

				for scanner.Scan() {
					line := scanner.Text()
					if regexp.MustCompile(methodPattern).MatchString(line) {
						// For class implementation, check if it's the right class
						if strings.Contains(line, className+"::") || lineNumber > classLineNumber {
							return lineNumber
						}
					}
					lineNumber++
				}
			}
		}
	} else {
		// Regular function search
		pattern := `(?:[a-zA-Z0-9_:]+\s+)+` + regexp.QuoteMeta(functionName) + `\s*\(`
		scanner := bufio.NewScanner(strings.NewReader(fileContent))
		lineNumber := 1

		for scanner.Scan() {
			line := scanner.Text()
			if regexp.MustCompile(pattern).MatchString(line) {
				return lineNumber
			}
			lineNumber++
		}
	}

	return -1
}

// isCppBuiltin checks if a function name is a C/C++ built-in
func isCppBuiltin(name string) bool {
	builtins := map[string]bool{
		"printf":           true,
		"scanf":            true,
		"malloc":           true,
		"free":             true,
		"calloc":           true,
		"realloc":          true,
		"memcpy":           true,
		"memset":           true,
		"strlen":           true,
		"strcpy":           true,
		"strcmp":           true,
		"strcat":           true,
		"fopen":            true,
		"fclose":           true,
		"fread":            true,
		"fwrite":           true,
		"fprintf":          true,
		"fscanf":           true,
		"std::cout":        true,
		"std::cin":         true,
		"std::cerr":        true,
		"std::endl":        true,
		"std::string":      true,
		"std::vector":      true,
		"std::map":         true,
		"std::make_shared": true,
		"std::make_unique": true,
		"new":              true,
		"delete":           true,
		"sizeof":           true,
	}

	return builtins[name]
}
