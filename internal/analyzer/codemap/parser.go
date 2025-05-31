package codemap

// LanguageParser defines the interface for parsing code in different languages
type LanguageParser interface {
	// ExtractImports extracts import statements from file content
	ExtractImports(fileContent string) []string

	// ExtractFunctions extracts functions from file content
	ExtractFunctions(fileContent string) []Function

	// ExtractFunctionCalls extracts function calls from a function body
	ExtractFunctionCalls(functionBody string) []string

	// ResolveImportPath resolves an import statement to a file path
	ResolveImportPath(importPath string, currentFilePath string, basePath string) string

	// GetFunctionLineNumber gets the line number where a function starts
	GetFunctionLineNumber(fileContent string, functionName string) int
}

// GetParserForFile returns the appropriate parser for a given file path
func GetParserForFile(filePath string) LanguageParser {
	extension := getFileExtension(filePath)

	switch extension {
	case ".go":
		return &GoParser{}
	case ".js", ".jsx":
		return &JavaScriptParser{}
	case ".ts", ".tsx":
		return &TypeScriptParser{}
	case ".py":
		return &PythonParser{}
	case ".java":
		return &JavaParser{}
	case ".c", ".cpp", ".h", ".hpp":
		return &CppParser{}
	case ".cs":
		return &CSharpParser{}
	default:
		// Return a generic parser that does minimal parsing
		return &GenericParser{}
	}
}

// getFileExtension extracts the file extension from a path
func getFileExtension(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			return filePath[i:]
		}
		if filePath[i] == '/' || filePath[i] == '\\' {
			break
		}
	}
	return ""
}
