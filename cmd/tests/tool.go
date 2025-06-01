package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tmc/langchaingo/llms/openai"
)

type embeddingPayload struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponsePayload []struct {
	Object    string      `json:"object"`
	Embedding [][]float32 `json:"embedding"`
	Index     int         `json:"index"`
}

func Raw(url, model, token string, texts []string) (*embeddingResponsePayload, error) {
	payload := embeddingPayload{
		Model: model,
		Input: texts,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url+"/embeddings", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}

	r, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer r.Body.Close()

	bb, _ := io.ReadAll(r.Body)
	fmt.Println(string(bb))

	// if r.StatusCode != http.StatusOK {
	// 	msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)

	// 	// No need to check the error here: if it fails, we'll just return the
	// 	// status code.
	// 	var errResp errorMessage
	// 	if err := json.NewDecoder(r.Body).Decode(&errResp); err != nil {
	// 		return nil, errors.New(msg) // nolint:goerr113
	// 	}

	// 	return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message) // nolint:goerr113
	// }

	var response embeddingResponsePayload

	json.Unmarshal(bb, &response)

	println(response[0].Embedding[0][0])
	return nil, nil
}

func main() {

	s := &strings.Builder{}
	for i := 0; i < 1024*3; i++ {
		s.WriteString(strconv.Itoa(i))
	}
	Raw("http://192.168.97.93:8080", "text-embedding-3-large", "sk-1234567890", []string{s.String()})

	opts := []openai.Option{
		openai.WithModel("gpt-3.5-turbo-0125"),
		openai.WithEmbeddingModel("text-embedding-3-large"),
		openai.WithBaseURL("http://192.168.97.93:8080"),
		openai.WithToken("sk-1234567890"),
	}
	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	embedings, err := llm.CreateEmbedding(ctx, []string{"ola", "mundo"})
	if err != nil {
		log.Fatal(err)
	}

	println(embedings)
}
