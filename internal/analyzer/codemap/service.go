package codemap

import (
	"fmt"
	"os"
	"path/filepath"
)

// CodeMapService provides a high-level interface for code mapping functionality
type CodeMapService struct {
	indexer  *CodeIndexer
	analyzer *DependencyAnalyzer
}

// NewCodeMapService creates a new code map service
func NewCodeMapService(apiKey, model, baseUrl string, dimensions int, basePath string) (*CodeMapService, error) {
	// Create storage
	storage := NewInMemoryEmbeddingStorage()

	// Create embedder
	embedder, err := NewOpenAIEmbedder(apiKey, model, baseUrl, dimensions, storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	// Create indexer
	indexer := NewCodeIndexer(embedder, basePath)

	// Create analyzer
	analyzer := NewDependencyAnalyzer(basePath)

	return &CodeMapService{
		indexer:  indexer,
		analyzer: analyzer,
	}, nil
}

// IndexRepository indexes all code files in a repository
func (s *CodeMapService) IndexRepository(repoPath, warehouseID string) error {
	// Initialize the analyzer
	if err := s.analyzer.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize analyzer: %w", err)
	}

	// Walk the repository and index all code files
	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-code files
		if !isSupportedExtension(filepath.Ext(path)) {
			return nil
		}

		// Index the file
		if err := s.indexer.IndexCodeFile(path, warehouseID); err != nil {
			return fmt.Errorf("failed to index file %s: %w", path, err)
		}

		return nil
	})
}

// SearchCode searches for code matching the query
func (s *CodeMapService) SearchCode(query, warehouseID string, limit int) ([]SearchResult, error) {
	return s.indexer.SearchCode(query, warehouseID, limit, 0.3) // 0.3 is the minimum relevance threshold
}

// AnalyzeFileDependencies analyzes the dependencies of a file
func (s *CodeMapService) AnalyzeFileDependencies(filePath string) (*DependencyTree, error) {
	return s.analyzer.AnalyzeFileDependencyTree(filePath)
}

// AnalyzeFunctionDependencies analyzes the dependencies of a function
func (s *CodeMapService) AnalyzeFunctionDependencies(filePath, functionName string) (*DependencyTree, error) {
	return s.analyzer.AnalyzeFunctionDependencyTree(filePath, functionName)
}

// GetSupportedLanguages returns a list of supported languages
func (s *CodeMapService) GetSupportedLanguages() []string {
	return []string{
		"go",
		"javascript",
		"typescript",
		"python",
		"java",
		"c",
		"cpp",
		"csharp",
		"ruby",
		"php",
		"swift",
		"kotlin",
		"rust",
		"scala",
	}
}
