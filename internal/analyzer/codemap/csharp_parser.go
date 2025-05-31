package codemap

import (
	"bufio"
	"path/filepath"
	"regexp"
	"strings"
)

// CSharpParser implements the LanguageParser interface for C# code
type CSharpParser struct{}

// ExtractImports extracts import statements from C# file content
func (p *CSharpParser) ExtractImports(fileContent string) []string {
	var imports []string

	// Match using statements
	usingRegex := regexp.MustCompile(`using\s+([a-zA-Z0-9_.]+)(?:\s*=\s*[a-zA-Z0-9_.]+)?;`)
	usingMatches := usingRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range usingMatches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// Match using static statements
	usingStaticRegex := regexp.MustCompile(`using\s+static\s+([a-zA-Z0-9_.]+);`)
	usingStaticMatches := usingStaticRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range usingStaticMatches {
		if len(match) >= 2 {
			imports = append(imports, "static:"+match[1])
		}
	}

	return imports
}

// ExtractFunctions extracts functions from C# file content
func (p *CSharpParser) ExtractFunctions(fileContent string) []Function {
	var functions []Function

	// First, extract namespace and class information
	namespaceRegex := regexp.MustCompile(`namespace\s+([a-zA-Z0-9_.]+)\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
	namespaceMatches := namespaceRegex.FindAllStringSubmatch(fileContent, -1)

	for _, namespaceMatch := range namespaceMatches {
		if len(namespaceMatch) >= 3 {
			namespace := namespaceMatch[1]
			namespaceBody := namespaceMatch[2]

			// Extract classes within the namespace
			classRegex := regexp.MustCompile(`(?:public|internal|private|protected|static|sealed|abstract)?\s*(?:class|struct|interface|record)\s+([a-zA-Z0-9_]+)(?:<[^>]*>)?(?:\s*:\s*[a-zA-Z0-9_<>.,\s]+)?\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
			classMatches := classRegex.FindAllStringSubmatch(namespaceBody, -1)

			for _, classMatch := range classMatches {
				if len(classMatch) >= 3 {
					className := classMatch[1]
					classBody := classMatch[2]

					// Extract methods within the class
					methodRegex := regexp.MustCompile(`(?:public|internal|private|protected|static|virtual|override|abstract|sealed|async|partial|extern)?\s+(?:async\s+)?(?:[a-zA-Z0-9_<>[\],.]+\s+)+([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s*where\s+[^{]+)?\s*(?:=>.*?;|{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\})`)
					methodMatches := methodRegex.FindAllStringSubmatch(classBody, -1)

					for _, methodMatch := range methodMatches {
						if len(methodMatch) >= 3 {
							methodName := methodMatch[1]

							// Skip constructors (same name as class)
							if methodName != className {
								functions = append(functions, Function{
									Name: namespace + "." + className + "." + methodName,
									Body: methodMatch[2],
								})
							} else {
								// It's a constructor
								functions = append(functions, Function{
									Name: namespace + "." + className + ".constructor",
									Body: methodMatch[2],
								})
							}
						}
					}

					// Extract constructors explicitly
					constructorRegex := regexp.MustCompile(`(?:public|internal|private|protected)\s+` + regexp.QuoteMeta(className) + `\s*\([^)]*\)(?:\s*:\s*(?:base|this)\s*\([^)]*\))?\s*{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
					constructorMatches := constructorRegex.FindAllStringSubmatch(classBody, -1)

					for _, constructorMatch := range constructorMatches {
						if len(constructorMatch) >= 2 {
							functions = append(functions, Function{
								Name: namespace + "." + className + ".constructor",
								Body: constructorMatch[1],
							})
						}
					}

					// Extract properties with non-trivial getters or setters
					propertyRegex := regexp.MustCompile(`(?:public|internal|private|protected|static|virtual|override|abstract|sealed)?\s+(?:[a-zA-Z0-9_<>[\],.]+\s+)+([a-zA-Z0-9_]+)\s*\{\s*(?:get\s*(?:=>.*?;|{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}))?\s*(?:set\s*(?:=>.*?;|{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}))?\s*\}`)
					propertyMatches := propertyRegex.FindAllStringSubmatch(classBody, -1)

					for _, propertyMatch := range propertyMatches {
						if len(propertyMatch) >= 3 {
							propertyName := propertyMatch[1]
							getterBody := propertyMatch[2]
							setterBody := ""
							if len(propertyMatch) >= 4 {
								setterBody = propertyMatch[3]
							}

							// Only add properties with non-trivial getters or setters
							if getterBody != "" || setterBody != "" {
								functions = append(functions, Function{
									Name: namespace + "." + className + "." + propertyName + ".get",
									Body: getterBody,
								})

								if setterBody != "" {
									functions = append(functions, Function{
										Name: namespace + "." + className + "." + propertyName + ".set",
										Body: setterBody,
									})
								}
							}
						}
					}
				}
			}

			// Extract top-level functions within the namespace
			funcRegex := regexp.MustCompile(`(?:public|internal|private|protected|static|partial|async)?\s+(?:async\s+)?(?:[a-zA-Z0-9_<>[\],.]+\s+)+([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s*where\s+[^{]+)?\s*{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
			funcMatches := funcRegex.FindAllStringSubmatch(namespaceBody, -1)

			for _, funcMatch := range funcMatches {
				if len(funcMatch) >= 3 {
					funcName := funcMatch[1]
					functions = append(functions, Function{
						Name: namespace + "." + funcName,
						Body: funcMatch[2],
					})
				}
			}
		}
	}

	// Also extract top-level classes outside of namespaces
	topClassRegex := regexp.MustCompile(`(?:public|internal|private|protected|static|sealed|abstract)?\s*(?:class|struct|interface|record)\s+([a-zA-Z0-9_]+)(?:<[^>]*>)?(?:\s*:\s*[a-zA-Z0-9_<>.,\s]+)?\s*\{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\}`)
	topClassMatches := topClassRegex.FindAllStringSubmatch(fileContent, -1)

	for _, classMatch := range topClassMatches {
		if len(classMatch) >= 3 {
			className := classMatch[1]
			classBody := classMatch[2]

			// Extract methods within the class
			methodRegex := regexp.MustCompile(`(?:public|internal|private|protected|static|virtual|override|abstract|sealed|async|partial|extern)?\s+(?:async\s+)?(?:[a-zA-Z0-9_<>[\],.]+\s+)+([a-zA-Z0-9_]+)\s*\([^)]*\)(?:\s*where\s+[^{]+)?\s*(?:=>.*?;|{((?:[^{}]|(?:\{(?:[^{}]|(?:\{(?:[^{}]|(?:\{[^{}]*\})|)*\})|)*\}))*)\})`)
			methodMatches := methodRegex.FindAllStringSubmatch(classBody, -1)

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
		}
	}

	return functions
}

// ExtractFunctionCalls extracts function calls from a function body
func (p *CSharpParser) ExtractFunctionCalls(functionBody string) []string {
	var calls []string
	seen := make(map[string]bool)

	// Match method calls
	callRegex := regexp.MustCompile(`([a-zA-Z0-9_]+(?:\.[a-zA-Z0-9_]+)*)\s*\(`)
	matches := callRegex.FindAllStringSubmatch(functionBody, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			funcName := match[1]

			// Skip C# built-ins and common patterns
			if isCSharpBuiltin(funcName) {
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

// ResolveImportPath resolves a C# using statement to a file path
func (p *CSharpParser) ResolveImportPath(importPath string, currentFilePath string, basePath string) string {
	// Handle static imports
	importPath = strings.TrimPrefix(importPath, "static:")

	// Convert namespace to directory structure
	importPath = strings.ReplaceAll(importPath, ".", string(filepath.Separator))

	// Try to find the file in common project structures

	// Try in the same directory
	dir := filepath.Dir(currentFilePath)
	fullPath := filepath.Join(dir, importPath+".cs")
	if _, err := filepath.Abs(fullPath); err == nil {
		return fullPath
	}

	// Try in project root
	fullPath = filepath.Join(basePath, importPath+".cs")
	if _, err := filepath.Abs(fullPath); err == nil {
		return fullPath
	}

	// Try in common source directories
	for _, sourceDir := range []string{"src", "source", "Sources", "App_Code"} {
		fullPath = filepath.Join(basePath, sourceDir, importPath+".cs")
		if _, err := filepath.Abs(fullPath); err == nil {
			return fullPath
		}
	}

	return filepath.Join(basePath, importPath+".cs")
}

// GetFunctionLineNumber gets the line number where a function starts
func (p *CSharpParser) GetFunctionLineNumber(fileContent string, functionName string) int {
	// Parse the function name
	parts := strings.Split(functionName, ".")
	if len(parts) >= 3 {
		// Format is namespace.class.method or class.method
		namespace := ""
		className := ""
		methodName := ""

		if len(parts) >= 3 {
			namespace = parts[0]
			className = parts[1]
			methodName = parts[2]
		} else {
			className = parts[0]
			methodName = parts[1]
		}

		// Special handling for constructors and properties
		isConstructor := methodName == "constructor"
		isProperty := len(parts) >= 4 && (parts[3] == "get" || parts[3] == "set")
		propertyAccessor := ""
		if isProperty {
			propertyAccessor = parts[3]
			methodName = parts[2]
		}

		// Find the namespace first (if applicable)
		var namespaceLineNumber int
		namespaceFound := false

		if namespace != "" {
			namespacePattern := `namespace\s+` + regexp.QuoteMeta(namespace)
			namespaceScanner := bufio.NewScanner(strings.NewReader(fileContent))
			namespaceLineNumber = 1

			for namespaceScanner.Scan() {
				line := namespaceScanner.Text()
				if regexp.MustCompile(namespacePattern).MatchString(line) {
					namespaceFound = true
					break
				}
				namespaceLineNumber++
			}

			if !namespaceFound {
				return -1
			}
		}

		// Find the class
		classPattern := `(?:class|struct|interface|record)\s+` + regexp.QuoteMeta(className)
		classScanner := bufio.NewScanner(strings.NewReader(fileContent))
		classLineNumber := 1

		if namespaceFound {
			// Skip to namespace line
			for i := 1; i < namespaceLineNumber; i++ {
				classScanner.Scan()
				classLineNumber++
			}
		}

		classFound := false
		for classScanner.Scan() {
			line := classScanner.Text()
			if regexp.MustCompile(classPattern).MatchString(line) {
				classFound = true
				break
			}
			classLineNumber++
		}

		if !classFound {
			return -1
		}

		// Now find the method or property
		var methodPattern string
		if isConstructor {
			methodPattern = `(?:public|internal|private|protected)\s+` + regexp.QuoteMeta(className) + `\s*\(`
		} else if isProperty {
			methodPattern = `(?:public|internal|private|protected|static|virtual|override|abstract|sealed)?\s+(?:[a-zA-Z0-9_<>[\],.]+\s+)+` + regexp.QuoteMeta(methodName) + `\s*\{`
		} else {
			methodPattern = `(?:public|internal|private|protected|static|virtual|override|abstract|sealed|async|partial|extern)?\s+(?:async\s+)?(?:[a-zA-Z0-9_<>[\],.]+\s+)+` + regexp.QuoteMeta(methodName) + `\s*\(`
		}

		methodScanner := bufio.NewScanner(strings.NewReader(fileContent))
		methodLineNumber := 1

		// Skip to class line
		for i := 1; i < classLineNumber; i++ {
			methodScanner.Scan()
			methodLineNumber++
		}

		for methodScanner.Scan() {
			line := methodScanner.Text()
			if regexp.MustCompile(methodPattern).MatchString(line) {
				if isProperty {
					// For properties, find the specific accessor
					accessorPattern := propertyAccessor + `\s*(?:=>|{)`
					inProperty := true
					braceCount := 0

					// Skip to property body
					for methodScanner.Scan() {
						methodLineNumber++
						line = methodScanner.Text()

						// Count braces to track property scope
						for _, char := range line {
							if char == '{' {
								braceCount++
							} else if char == '}' {
								braceCount--
								if braceCount < 0 {
									inProperty = false
									break
								}
							}
						}

						if !inProperty {
							break
						}

						if regexp.MustCompile(accessorPattern).MatchString(line) {
							return methodLineNumber
						}
					}
				} else {
					return methodLineNumber
				}
			}
			methodLineNumber++
		}
	}

	return -1
}

// isCSharpBuiltin checks if a function name is a C# built-in
func isCSharpBuiltin(name string) bool {
	builtins := map[string]bool{
		"Console.WriteLine":         true,
		"Console.Write":             true,
		"Console.ReadLine":          true,
		"Console.Read":              true,
		"string.Format":             true,
		"int.Parse":                 true,
		"double.Parse":              true,
		"bool.Parse":                true,
		"Convert.ToInt32":           true,
		"Convert.ToString":          true,
		"Convert.ToDouble":          true,
		"Convert.ToBoolean":         true,
		"Math.Abs":                  true,
		"Math.Min":                  true,
		"Math.Max":                  true,
		"Math.Sqrt":                 true,
		"Math.Round":                true,
		"string.IsNullOrEmpty":      true,
		"string.IsNullOrWhiteSpace": true,
		"Enumerable.Where":          true,
		"Enumerable.Select":         true,
		"Enumerable.OrderBy":        true,
		"Enumerable.GroupBy":        true,
		"Enumerable.ToList":         true,
		"Enumerable.ToArray":        true,
		"List.Add":                  true,
		"List.Remove":               true,
		"List.Clear":                true,
		"Dictionary.Add":            true,
		"Dictionary.Remove":         true,
		"Dictionary.ContainsKey":    true,
		"Task.Run":                  true,
		"Task.Delay":                true,
		"Task.WhenAll":              true,
		"Task.WhenAny":              true,
		"Task.FromResult":           true,
		"Equals":                    true,
		"ToString":                  true,
		"GetHashCode":               true,
		"GetType":                   true,
	}

	return builtins[name]
}
