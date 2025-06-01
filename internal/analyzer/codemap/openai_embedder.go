package codemap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/o0olele/opendeepwiki-go/internal/config"
	"github.com/o0olele/opendeepwiki-go/internal/llm/embedding"
	"github.com/tmc/langchaingo/textsplitter"
)

// DocEmbedder implements the Embedder interface using OpenAI embeddings
type DocEmbedder struct {
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

// NewEmbedder creates a new OpenAI embedder
func NewEmbedder(storage EmbeddingStorage) (*DocEmbedder, error) {
	var embeddingConfig = config.GetEmbeddingConfig()
	var factory = embedding.NewFactory(
		embeddingConfig.ProviderType,
		embeddingConfig.APIKey,
		embeddingConfig.Model,
		embeddingConfig.BaseURL,
		3)
	embedder, err := factory.Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create embedder: %w", err)
	}

	return &DocEmbedder{
		embedder:   embedder,
		storage:    storage,
		dimensions: 3,
	}, nil
}

// IndexDocument indexes a document with the given content and metadata
func (e *DocEmbedder) IndexDocument(id string, content string, metadata map[string]string) error {

	splitter := textsplitter.NewTokenSplitter(
		textsplitter.WithChunkSize(4096),
		textsplitter.WithChunkOverlap(128))

	chunks, err := splitter.SplitText(content)
	if err != nil {
		return fmt.Errorf("failed to split text: %w", err)
	}

	for idx, chunk := range chunks {
		// Generate embedding for the content
		embeddings, err := e.embedder.Embed(context.Background(), []string{chunk})
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		if len(embeddings) == 0 {
			return fmt.Errorf("no embedding generated")
		}
		// Store the embedding
		if err := e.storage.StoreEmbedding(fmt.Sprintf("%s_%d", id, idx), embeddings[0], chunk, metadata); err != nil {
			return fmt.Errorf("failed to store embedding: %w", err)
		}
	}

	return nil
}

// Search searches for documents matching the query
func (e *DocEmbedder) Search(query string, filter map[string]string, limit int, minRelevance float64) ([]SearchResult, error) {
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
