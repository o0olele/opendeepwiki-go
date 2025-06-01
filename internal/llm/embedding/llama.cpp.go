package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/tmc/langchaingo/llms/openai"
)

type LlamaCppEmbedder struct {
	client     *http.Client
	model      string
	dimensions int
	baseUrl    string
	apiKey     string
}

type embeddingPayload struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponsePayload []struct {
	Object    string      `json:"object"`
	Embedding [][]float32 `json:"embedding"`
	Index     int         `json:"index"`
}

type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func NewLlamaCppEmbedder(apiKey, model, baseUrl string, dimensions int) (*LlamaCppEmbedder, error) {

	return &LlamaCppEmbedder{
		client:     &http.Client{},
		model:      model,
		dimensions: dimensions,
		baseUrl:    baseUrl,
		apiKey:     apiKey,
	}, nil
}

func (e *LlamaCppEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	payload := embeddingPayload{
		Model: e.model,
		Input: texts,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, e.baseUrl+"/embeddings", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	r, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)

		// No need to check the error here: if it fails, we'll just return the
		// status code.
		var errResp errorMessage
		if err := json.NewDecoder(r.Body).Decode(&errResp); err != nil {
			return nil, errors.New(msg) // nolint:goerr113
		}

		return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message) // nolint:goerr113
	}

	var response embeddingResponsePayload

	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(response) == 0 {
		return nil, openai.ErrEmptyResponse
	}

	embeddings := make([][]float32, 0)
	for i := 0; i < len(response[0].Embedding); i++ {
		embeddings = append(embeddings, response[0].Embedding[i])
	}

	if len(embeddings) == 0 {
		return nil, openai.ErrEmptyResponse
	}
	if len(texts) != len(embeddings) {
		return embeddings, openai.ErrUnexpectedResponseLength
	}

	return embeddings, nil
}

// BatchEmbed generates embeddings for the given texts in batches
func (e *LlamaCppEmbedder) BatchEmbed(ctx context.Context, texts []string, batchSize int) ([][]float32, error) {
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
func (e *LlamaCppEmbedder) GetDimensions() int {
	return e.dimensions
}

// GetModel returns the model used for embeddings
func (e *LlamaCppEmbedder) GetModel() string {
	return e.model
}
