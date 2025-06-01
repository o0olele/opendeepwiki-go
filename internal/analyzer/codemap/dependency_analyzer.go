package codemap

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DependencyAnalyzer analyzes code dependencies in a repository
type DependencyAnalyzer struct {
	FileDependencies     map[string]map[string]bool // Map of file to its dependencies
	FunctionDependencies map[string]map[string]bool // Map of function to its dependencies
	FileToFunctions      map[string][]*FunctionInfo // Map of file to its functions
	FunctionToFile       map[string]string          // Map of function full name to its file
	BasePath             string                     // Base path of the repository
	mutex                sync.RWMutex               // Mutex for concurrent access
	initialized          bool                       // Whether the analyzer has been initialized
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(basePath string) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		FileDependencies:     make(map[string]map[string]bool),
		FunctionDependencies: make(map[string]map[string]bool),
		FileToFunctions:      make(map[string][]*FunctionInfo),
		FunctionToFile:       make(map[string]string),
		BasePath:             basePath,
		initialized:          false,
	}
}

// Initialize initializes the dependency analyzer
func (a *DependencyAnalyzer) Initialize() error {
	if a.initialized {
		return nil
	}

	files, err := a.getAllSourceFiles(a.BasePath)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(files))

	// Process files concurrently
	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			parser := GetParserForFile(filePath)
			if parser == nil {
				return
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				errors <- err
				return
			}

			if err := a.processFile(filePath, string(content), parser); err != nil {
				errors <- err
				return
			}
		}(file)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	a.initialized = true
	return nil
}

func (a *DependencyAnalyzer) LoadFromFile(filePath string) error {
	err := readFromFile(filePath, a, "")
	if err != nil {
		return err
	}

	a.initialized = true
	return nil
}

func (a *DependencyAnalyzer) SaveToFile(filePath string) error {
	return persistToFile(filePath, a, false, "")
}

// processFile processes a single file
func (a *DependencyAnalyzer) processFile(filePath, fileContent string, parser LanguageParser) error {
	// Extract imports
	imports := parser.ExtractImports(fileContent)
	resolvedImports := a.resolveImportPaths(imports, filePath, a.BasePath)

	// Store file dependencies
	a.mutex.Lock()
	a.FileDependencies[filePath] = resolvedImports
	a.mutex.Unlock()

	// Extract functions
	functions := parser.ExtractFunctions(fileContent)
	var functionInfoList []*FunctionInfo

	for _, function := range functions {
		functionInfo := &FunctionInfo{
			Name:       function.Name,
			FullName:   filePath + ":" + function.Name,
			Body:       function.Body,
			FilePath:   filePath,
			LineNumber: parser.GetFunctionLineNumber(fileContent, function.Name),
			Calls:      parser.ExtractFunctionCalls(function.Body),
		}

		functionInfoList = append(functionInfoList, functionInfo)

		a.mutex.Lock()
		a.FunctionToFile[functionInfo.FullName] = filePath
		a.mutex.Unlock()
	}

	a.mutex.Lock()
	a.FileToFunctions[filePath] = functionInfoList
	a.mutex.Unlock()

	return nil
}

// resolveImportPaths resolves import paths to file paths
func (a *DependencyAnalyzer) resolveImportPaths(imports []string, currentFile, basePath string) map[string]bool {
	result := make(map[string]bool)

	for _, importPath := range imports {
		parser := GetParserForFile(currentFile)
		if parser == nil {
			continue
		}

		resolvedPath := parser.ResolveImportPath(importPath, currentFile, basePath)
		if resolvedPath != "" {
			result[resolvedPath] = true
		}
	}

	return result
}

// getAllSourceFiles gets all source files in a directory
func (a *DependencyAnalyzer) getAllSourceFiles(path string) ([]string, error) {
	var files []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip vendor and node_modules directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == "node_modules") {
			return filepath.SkipDir
		}

		// Only include files with recognized extensions
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if isSupportedExtension(ext) {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

// isSupportedExtension checks if a file extension is supported
func isSupportedExtension(ext string) bool {
	supportedExts := map[string]bool{
		".go":   true,
		".js":   true,
		".jsx":  true,
		".ts":   true,
		".tsx":  true,
		".py":   true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
		".cs":   true,
		".rb":   true,
		".php":  true,
	}

	return supportedExts[ext]
}

// AnalyzeFileDependencyTree analyzes the dependency tree for a file
func (a *DependencyAnalyzer) AnalyzeFileDependencyTree(filePath string) (*DependencyTree, error) {
	if !a.initialized {
		if err := a.Initialize(); err != nil {
			return nil, err
		}
	}

	normalizedPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	return a.buildFileDependencyTree(normalizedPath, visited, 0), nil
}

// AnalyzeFunctionDependencyTree analyzes the dependency tree for a function
func (a *DependencyAnalyzer) AnalyzeFunctionDependencyTree(filePath, functionName string) (*DependencyTree, error) {
	if !a.initialized {
		if err := a.Initialize(); err != nil {
			return nil, err
		}
	}

	normalizedPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	return a.buildFunctionDependencyTree(normalizedPath, functionName, visited, 0), nil
}

// buildFileDependencyTree builds a dependency tree for a file
func (a *DependencyAnalyzer) buildFileDependencyTree(filePath string, visited map[string]bool, level int) *DependencyTree {
	const maxDepth = 10

	if level > maxDepth || visited[filePath] {
		return &DependencyTree{
			NodeType:  FileNodeType,
			Name:      filepath.Base(filePath),
			FullPath:  filePath,
			IsCyclic:  visited[filePath],
			Children:  []*DependencyTree{},
			Functions: []*DependencyFunction{},
		}
	}

	visited[filePath] = true

	tree := &DependencyTree{
		NodeType:  FileNodeType,
		Name:      filepath.Base(filePath),
		FullPath:  filePath,
		IsCyclic:  false,
		Children:  []*DependencyTree{},
		Functions: []*DependencyFunction{},
	}

	// Add child file dependencies
	a.mutex.RLock()
	dependencies := a.FileDependencies[filePath]
	a.mutex.RUnlock()

	for dependency := range dependencies {
		childVisited := make(map[string]bool)
		for k, v := range visited {
			childVisited[k] = v
		}

		child := a.buildFileDependencyTree(dependency, childVisited, level+1)
		tree.Children = append(tree.Children, child)
	}

	// Add functions in the file
	a.mutex.RLock()
	functions := a.FileToFunctions[filePath]
	a.mutex.RUnlock()

	for _, function := range functions {
		tree.Functions = append(tree.Functions, &DependencyFunction{
			Name:       function.Name,
			LineNumber: function.LineNumber,
		})
	}

	return tree
}

// buildFunctionDependencyTree builds a dependency tree for a function
func (a *DependencyAnalyzer) buildFunctionDependencyTree(filePath, functionName string, visited map[string]bool, level int) *DependencyTree {
	const maxDepth = 20

	// fullFunctionID is used as a unique identifier for the function in the visited map
	// to detect cycles in the dependency graph
	fullFunctionID := filePath + ":" + functionName

	if level > maxDepth || visited[fullFunctionID] {
		return &DependencyTree{
			NodeType:  FunctionNodeType,
			Name:      functionName,
			FullPath:  fullFunctionID,
			IsCyclic:  visited[fullFunctionID],
			Children:  []*DependencyTree{},
			Functions: []*DependencyFunction{},
		}
	}

	visited[fullFunctionID] = true

	tree := &DependencyTree{
		NodeType:  FunctionNodeType,
		Name:      functionName,
		FullPath:  fullFunctionID,
		IsCyclic:  false,
		Children:  []*DependencyTree{},
		Functions: []*DependencyFunction{},
	}

	// Find the current function info
	a.mutex.RLock()
	functions := a.FileToFunctions[filePath]
	a.mutex.RUnlock()

	for _, function := range functions {
		if function.Name == functionName {
			tree.LineNumber = function.LineNumber

			// Add function calls as children
			for _, call := range function.Calls {
				// Try to resolve the function call
				callInfo := a.resolveFunctionCall(call, filePath)
				if callInfo != nil {
					childVisited := make(map[string]bool)
					for k, v := range visited {
						childVisited[k] = v
					}

					child := a.buildFunctionDependencyTree(callInfo.FilePath, callInfo.Name, childVisited, level+1)
					tree.Children = append(tree.Children, child)
				}
			}

			break
		}
	}

	return tree
}

// resolveFunctionCall resolves a function call to a function info
func (a *DependencyAnalyzer) resolveFunctionCall(functionCall, currentFile string) *FunctionInfo {
	// Check if the function is in the current file
	a.mutex.RLock()
	functions := a.FileToFunctions[currentFile]
	a.mutex.RUnlock()

	for _, function := range functions {
		if function.Name == functionCall {
			return function
		}
	}

	// Check if the function is in a different file
	// This is a simplified approach and may not work for all cases
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	for _, fileFunctions := range a.FileToFunctions {
		for _, function := range fileFunctions {
			if function.Name == functionCall {
				return function
			}
		}
	}

	return nil
}
