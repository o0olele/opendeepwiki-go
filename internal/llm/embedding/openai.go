package embedding

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAIEmbedder implements the Embedder interface using OpenAI's embedding API
type OpenAIEmbedder struct {
	llm        *openai.LLM
	model      string
	dimensions int
}

// NewOpenAIEmbedder creates a new OpenAIEmbedder
func NewOpenAIEmbedder(apiKey, model, baseUrl string, dimensions int) (*OpenAIEmbedder, error) {
	if apiKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	opts := []openai.Option{
		openai.WithToken(apiKey),
	}
	if len(model) > 0 {
		opts = append(opts, openai.WithModel(model))
	}
	if len(baseUrl) > 0 {
		opts = append(opts, openai.WithBaseURL(baseUrl))
	}
	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}

	return &OpenAIEmbedder{
		llm:        llm,
		dimensions: dimensions,
	}, nil
}

// Embed generates embeddings for the given texts
func (e *OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("no texts provided for embedding")
	}

	response, err := e.llm.CreateEmbedding(
		ctx,
		texts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	if len(response) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(response))
	}

	embeddings := make([][]float32, len(response))
	for i, item := range response {
		embeddings[i] = item
	}

	return embeddings, nil
}

// BatchEmbed generates embeddings for the given texts in batches
func (e *OpenAIEmbedder) BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("no texts provided for embedding")
	}

	if batchSize <= 0 {
		batchSize = 100
	}

	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := e.Embed(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to embed batch %d-%d: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// GetDimensions returns the dimensions of the embeddings
func (e *OpenAIEmbedder) GetDimensions() int {
	return e.dimensions
}

// GetModel returns the model used for embeddings
func (e *OpenAIEmbedder) GetModel() string {
	return e.model
}
