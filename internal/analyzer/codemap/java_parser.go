package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// JavaParser implements the LanguageParser interface for Java code
type JavaParser struct{}

// ExtractImports extracts import statements from Java file content
func (p *JavaParser) ExtractImports(fileContent string) []string {
	var imports []string

	// Match import statements
	importRegex := regexp.MustCompile(`import\s+([a-zA-Z0-9_.]+(?:\.[a-zA-Z0-9_*]+)*)\s*;`)
	importMatches := importRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range importMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Match static imports
	staticImportRegex := regexp.MustCompile(`import\s+static\s+([a-zA-Z0-9_.]+(?:\.[a-zA-Z0-9_*]+)*)\s*;`)
	staticImportMatches := staticImportRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range staticImportMatches {
		if len(match) >= 2 {
			imports = append(imports, "static:"+match[1])
		}
	}

	return imports
}

// ExtractFunctions extracts functions from Java file content
func (p *JavaParser) ExtractFunctions(fileContent string) []Function {
	var functions []Function

	// First, extract class names
	classRegex := regexp.MustCompile(`(?:public|protected|private|static|final|abstract)?\s*(?:class|interface|enum)\s+([a-zA-Z0-9_]+)(?:\s+extends\s+[a-zA-Z0-9_<>.]+)?(?:\s+implements\s+[a-zA-Z0-9_<>.,\s]+)?\s*\{`)
	classMatches := classRegex.FindAllStringSubmatch(fileContent, -1)

	for _, classMatch := range classMatches {
		if len(classMatch) >= 2 {
			className := classMatch[1]

			// Extract methods within the class
			// This is a simplified approach and may not work for all cases
			methodRegex := regexp.MustCompile(`(?:public|protected|private|static|final|synchronized|native|abstract)?\s+(?:[a-zA-Z0-9_<>[\],.]+\s+)+([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s+throws\s+[a-zA-Z0-9_<>.,\s]+)?\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
			methodMatches := methodRegex.FindAllStringSubmatch(fileContent, -1)

			for _, methodMatch := range methodMatches {
				if len(methodMatch) >= 3 {
					methodName := methodMatch[1]

					// Skip constructors (same name as class)
					if methodName != className {
						functions = append(functions, Function{
							Name: className + "." + methodName,
							Body: methodMatch[2],
						})
					} else {
						// It's a constructor
						functions = append(functions, Function{
							Name: className + ".constructor",
							Body: methodMatch[2],
						})
					}
				}
			}

			// Extract constructors explicitly
			constructorRegex := regexp.MustCompile(`(?:public|protected|private)\s+` + regexp.QuoteMeta(className) + `\s*\([^)]*\)(?:\s+throws\s+[a-zA-Z0-9_<>.,\s]+)?\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
			constructorMatches := constructorRegex.FindAllStringSubmatch(fileContent, -1)

			for _, constructorMatch := range constructorMatches {
				if len(constructorMatch) >= 2 {
					functions = append(functions, Function{
						Name: className + ".constructor",
						Body: constructorMatch[1],
					})
				}
			}
		}
	}

	return functions
}

// ExtractFunctionCalls extracts function calls from a function body
func (p *JavaParser) ExtractFunctionCalls(functionBody string) []string {
	var calls []string
	seen := make(map[string]bool)

	// Match method calls
	callRegex := regexp.MustCompile(`([a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)*)\s*\(`)
	matches := callRegex.FindAllStringSubmatch(functionBody, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			funcName := match[1]

			// Skip Java built-ins and common patterns
			if isJavaBuiltin(funcName) {
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

// ResolveImportPath resolves a Java import statement to a file path
func (p *JavaParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle static imports
	importPath = strings.TrimPrefix(importPath, "static:")

	// Handle wildcard imports
	importPath = strings.TrimSuffix(importPath, ".*")

	// Convert package path to directory path
	importPath = strings.ReplaceAll(importPath, ".", string(filepath.Separator))

	// Try to find the file
	javaFile := importPath + ".java"
	fullPath := filepath.Join(basePath, "src", javaFile)

	// Try in src/main/java if it's a Maven/Gradle project
	if _, err := filepath.Abs(fullPath); err != nil {
		fullPath = filepath.Join(basePath, "src", "main", "java", javaFile)
	}

	return fullPath
}

// GetFunctionLineNumber gets the line number where a function starts
func (p *JavaParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	// Handle class methods
	if strings.Contains(functionName, ".") {
		parts := strings.Split(functionName, ".")
		if len(parts) == 2 {
			className := parts[0]
			methodName := parts[1]

			// Special handling for constructors
			isConstructor := methodName == "constructor"

			// Find the class first
			classPattern := `(?:public|protected|private|static|final|abstract)?\s*(?:class|interface|enum)\s+` + regexp.QuoteMeta(className)
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
				var methodPattern string
				if isConstructor {
					methodPattern = `(?:public|protected|private)\s+` + regexp.QuoteMeta(className) + `\s*\(`
				} else {
					methodPattern = `(?:public|protected|private|static|final|synchronized|native|abstract)?\s+(?:[a-zA-Z0-9_<>[\],.]+\s+)+` + regexp.QuoteMeta(methodName) + `\s*\(`
				}

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

	return -1
}

// isJavaBuiltin checks if a function name is a Java built-in
func isJavaBuiltin(name string) bool {
	builtins := map[string]bool{
		"System.out.println":   true,
		"System.out.print":     true,
		"System.err.println":   true,
		"System.err.print":     true,
		"String.format":        true,
		"Integer.parseInt":     true,
		"Double.parseDouble":   true,
		"Boolean.parseBoolean": true,
		"Math.abs":             true,
		"Math.min":             true,
		"Math.max":             true,
		"Math.sqrt":            true,
		"Math.random":          true,
		"Arrays.toString":      true,
		"Arrays.asList":        true,
		"Collections.sort":     true,
		"equals":               true,
		"toString":             true,
		"hashCode":             true,
		"clone":                true,
		"compareTo":            true,
		"super":                true,
		"this":                 true,
	}

	return builtins[name]
}
