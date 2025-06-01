package codemap

import (
	"fmt"
	"os"
	"path/filepath"
)

// CodeMapService provides a high-level interface for code mapping functionality
type CodeMapService struct {
	indexer   *CodeIndexer
	analyzer  *DependencyAnalyzer
	embedding *InMemoryEmbeddingStorage
}

// NewCodeMapService creates a new code map service
func NewCodeMapService(basePath string) (*CodeMapService, error) {

	// Create storage
	var storage = NewInMemoryEmbeddingStorage()

	// Create embedder
	embedder, err := NewEmbedder(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}
	// Create analyzer
	analyzer := NewDependencyAnalyzer(basePath)

	// Create indexer
	indexer := NewCodeIndexer(embedder, basePath, analyzer)

	return &CodeMapService{
		indexer:   indexer,
		analyzer:  analyzer,
		embedding: storage,
	}, nil
}

func (s *CodeMapService) LoadFromFile(codePath, vectorPath string) error {
	if len(codePath) > 0 {
		err := s.indexer.LoadFromFile(codePath)
		if err != nil {
			return fmt.Errorf("failed to load analyzer: %w", err)
		}
	}

	if len(vectorPath) > 0 {
		err := s.embedding.LoadFromFile(vectorPath)
		if err != nil {
			return fmt.Errorf("failed to load embedding: %w", err)
		}
	}

	return nil
}

func (s *CodeMapService) SaveToFile(codePath, vectorPath string) error {
	if len(codePath) > 0 {
		err := s.indexer.SaveToFile(codePath)
		if err != nil {
			return fmt.Errorf("failed to save analyzer: %w", err)
		}
	}

	if len(vectorPath) > 0 {
		err := s.embedding.SaveToFile(vectorPath)
		if err != nil {
			return fmt.Errorf("failed to save embedding: %w", err)
		}
	}
	return nil
}

// IndexRepository indexes all code files in a repository
func (s *CodeMapService) IndexRepository(repoPath, warehouseID string) (bool, error) {
	// Initialize the analyzer
	if err := s.analyzer.Initialize(); err != nil {
		return true, fmt.Errorf("failed to initialize analyzer: %w", err)
	}

	if s.indexer.inited {
		return false, nil
	}

	// Walk the repository and index all code files
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
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
	if err != nil {
		return true, fmt.Errorf("failed to walk repository: %w", err)
	}

	return true, nil
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
