package embedding

import (
	"context"
	"errors"
	"fmt"

	"github.com/o0olele/opendeepwiki-go/internal/config"
)

// Service handles embedding operations
type Service struct {
	cfg    *config.Config
	client EmbeddingClient
}

// EmbeddingClient defines the interface for embedding clients
type EmbeddingClient interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	BatchEmbedding(ctx context.Context, texts []string) ([][]float32, error)
}

// NewService creates a new embedding service
func NewService(cfg *config.Config, client EmbeddingClient) *Service {
	return &Service{
		cfg:    cfg,
		client: client,
	}
}

// GenerateEmbedding generates an embedding for a text
func (s *Service) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if s.client == nil {
		return nil, errors.New("embedding client not initialized")
	}

	return s.client.GenerateEmbedding(ctx, text)
}

// BatchEmbedding generates embeddings for multiple texts
func (s *Service) BatchEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	if s.client == nil {
		return nil, errors.New("embedding client not initialized")
	}

	// Check if the batch size is within limits
	batchSize := s.cfg.Embedding.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	// If the number of texts is less than or equal to the batch size, process all at once
	if len(texts) <= batchSize {
		return s.client.BatchEmbedding(ctx, texts)
	}

	// Otherwise, process in batches
	var allEmbeddings [][]float32
	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := s.client.BatchEmbedding(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("error processing batch %d-%d: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// CosineSimilarity calculates the cosine similarity between two embeddings
func CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("embeddings have different dimensions")
	}

	var dotProduct, magnitudeA, magnitudeB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		magnitudeA += a[i] * a[i]
		magnitudeB += b[i] * b[i]
	}

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0, nil
	}

	return dotProduct / (float32(magnitudeA) * float32(magnitudeB)), nil
}
