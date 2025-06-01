package codemap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CodeIndexer indexes code for searching and analysis
type CodeIndexer struct {
	embedder Embedder
	analyzer *DependencyAnalyzer
	basePath string
	inited   bool
}

// Embedder defines the interface for generating embeddings
type Embedder interface {
	// IndexDocument indexes a document with the given content and metadata
	IndexDocument(id string, content string, metadata map[string]string) error

	// Search searches for documents matching the query
	Search(query string, filter map[string]string, limit int, minRelevance float64) ([]SearchResult, error)
}

// NewCodeIndexer creates a new code indexer
func NewCodeIndexer(embedder Embedder, basePath string, analyzer *DependencyAnalyzer) *CodeIndexer {
	return &CodeIndexer{
		embedder: embedder,
		analyzer: analyzer,
		basePath: basePath,
	}
}

func (i *CodeIndexer) LoadFromFile(codePath string) error {
	err := i.analyzer.LoadFromFile(codePath)
	if err != nil {
		return fmt.Errorf("failed to load analyzer: %w", err)
	}
	i.inited = true
	return nil
}

func (i *CodeIndexer) SaveToFile(codePath string) error {
	err := i.analyzer.SaveToFile(codePath)
	if err != nil {
		return fmt.Errorf("failed to save analyzer: %w", err)
	}
	return nil
}

// IndexCodeFile indexes a code file for searching
func (i *CodeIndexer) IndexCodeFile(filePath string, warehouseID string) error {
	if i.inited {
		return nil
	}
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Determine language
	language := determineLanguage(filePath)
	fileName := filepath.Base(filePath)

	// Analyze dependencies
	dependencyTree, err := i.analyzer.AnalyzeFileDependencyTree(filePath)
	if err != nil {
		return fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	// Serialize dependency tree to JSON
	dependencyJSON, err := json.Marshal(dependencyTree)
	if err != nil {
		return fmt.Errorf("failed to serialize dependency tree: %w", err)
	}

	// Create metadata
	metadata := map[string]string{
		"warehouse_id":  warehouseID,
		"file_name":     fileName,
		"file_path":     filePath,
		"code_language": language,
		"language":      language,
		"dependencies":  string(dependencyJSON),
	}

	// Index the document
	documentID := fmt.Sprintf("%s:%s", warehouseID, filePath)
	if err := i.embedder.IndexDocument(documentID, string(content), metadata); err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}

	return nil
}

// SearchCode searches for code matching the query
func (i *CodeIndexer) SearchCode(query string, warehouseID string, limit int, minRelevance float64) ([]SearchResult, error) {
	filter := map[string]string{
		"warehouse_id": warehouseID,
	}

	return i.embedder.Search(query, filter, limit, minRelevance)
}

// determineLanguage determines the programming language of a file based on its extension
func determineLanguage(filePath string) string {
	extension := strings.ToLower(filepath.Ext(filePath))

	switch extension {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".h", ".hpp":
		return "cpp_header"
	case ".cs":
		return "csharp"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".rs":
		return "rust"
	case ".scala":
		return "scala"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".md", ".markdown":
		return "markdown"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".xml":
		return "xml"
	case ".sql":
		return "sql"
	case ".sh", ".bash":
		return "shell"
	case ".ps1":
		return "powershell"
	default:
		return "unknown"
	}
}
