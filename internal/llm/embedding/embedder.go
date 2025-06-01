package embedding

import (
	"context"
)

// Embedder defines the interface for generating embeddings
type Embedder interface {
	// Embed generates embeddings for the given texts
	Embed(ctx context.Context, texts []string) ([][]float32, error)

	// BatchEmbed generates embeddings for the given texts in batches
	BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error)

	// GetDimensions returns the dimensions of the embeddings
	GetDimensions() int

	// GetModel returns the model used for embeddings
	GetModel() string
}

// Factory creates embedders based on provider
type Factory struct {
	Provider   string
	APIKey     string
	Model      string
	BaseUrl    string
	Dimensions int
}

// NewFactory creates a new embedder factory
func NewFactory(provider, apiKey, model, baseUrl string, dimensions int) *Factory {
	return &Factory{
		Provider:   provider,
		APIKey:     apiKey,
		Model:      model,
		Dimensions: dimensions,
		BaseUrl:    baseUrl,
	}
}

// Create creates a new embedder based on the provider
func (f *Factory) Create() (Embedder, error) {
	switch f.Provider {
	case "openai":
		return NewOpenAIEmbedder(f.APIKey, f.Model, f.BaseUrl, f.Dimensions)
	case "llamacpp":
		return NewLlamaCppEmbedder(f.APIKey, f.Model, f.BaseUrl, f.Dimensions)
	default:
		// Default to OpenAI
		return NewOpenAIEmbedder(f.APIKey, f.Model, f.BaseUrl, f.Dimensions)
	}
}
