package codemap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/o0olele/opendeepwiki-go/internal/llm/embedding"
)

// OpenAIEmbedder implements the Embedder interface using OpenAI embeddings
type OpenAIEmbedder struct {
	embedder   embedding.Embedder
	storage    EmbeddingStorage
	dimensions int
}

// EmbeddingStorage defines the interface for storing and retrieving embeddings
type EmbeddingStorage interface {
	// StoreEmbedding stores an embedding with metadata
	StoreEmbedding(id string, embedding []float32, content string, metadata map[string]string) error

	// SearchEmbeddings searches for embeddings similar to the query embedding
	SearchEmbeddings(queryEmbedding []float32, filter map[string]string, limit int, minRelevance float64) ([]EmbeddingRecord, error)
}

// EmbeddingRecord represents a stored embedding with metadata
type EmbeddingRecord struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Metadata  map[string]string `json:"metadata"`
	Embedding []float32         `json:"embedding"`
	Score     float64           `json:"score"`
}

// NewOpenAIEmbedder creates a new OpenAI embedder
func NewOpenAIEmbedder(apiKey, model, baseUrl string, dimensions int, storage EmbeddingStorage) (*OpenAIEmbedder, error) {
	factory := embedding.NewFactory("openai", apiKey, model, baseUrl, dimensions)
	embedder, err := factory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &OpenAIEmbedder{
		embedder:   embedder,
		storage:    storage,
		dimensions: dimensions,
	}, nil
}

// IndexDocument indexes a document with the given content and metadata
func (e *OpenAIEmbedder) IndexDocument(id string, content string, metadata map[string]string) error {
	// Generate embedding for the content
	embeddings, err := e.embedder.Embed(context.Background(), []string{content})
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return fmt.Errorf("no embedding generated")
	}

	// Store the embedding
	if err := e.storage.StoreEmbedding(id, embeddings[0], content, metadata); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	return nil
}

// Search searches for documents matching the query
func (e *OpenAIEmbedder) Search(query string, filter map[string]string, limit int, minRelevance float64) ([]SearchResult, error) {
	// Generate embedding for the query
	embeddings, err := e.embedder.Embed(context.Background(), []string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no query embedding generated")
	}

	// Search for similar embeddings
	records, err := e.storage.SearchEmbeddings(embeddings[0], filter, limit, minRelevance)
	if err != nil {
		return nil, fmt.Errorf("failed to search embeddings: %w", err)
	}

	// Convert to search results
	results := make([]SearchResult, 0, len(records))
	for _, record := range records {
		// Try to parse dependencies from metadata
		var dependencyTree *DependencyTree
		if depJSON, ok := record.Metadata["dependencies"]; ok {
			if err := json.Unmarshal([]byte(depJSON), &dependencyTree); err == nil {
				// Successfully parsed dependency tree
			}
		}

		result := SearchResult{
			ID:          record.ID,
			Code:        record.Content,
			Description: buildDescription(record.Metadata),
			Relevance:   record.Score,
			References:  dependencyTree,
		}

		results = append(results, result)
	}

	return results, nil
}

// buildDescription builds a description from metadata
func buildDescription(metadata map[string]string) string {
	fileName := metadata["file_name"]
	language := metadata["code_language"]

	description := fmt.Sprintf("Code from %s (language: %s)", fileName, language)
	return description
}
