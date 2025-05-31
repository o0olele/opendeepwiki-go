package chat

import (
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// OllamaProvider Ollama 提供商实现
type OllamaProvider struct {
	model  llms.Model
	config *ProviderConfig
}

// NewOllamaProvider 创建一个新的 Ollama 提供商
func NewOllamaProvider(config *ProviderConfig) (Provider, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("未提供 OpenAI API 密钥")
		}
	}

	// 创建 OpenAI 选项
	opts := []ollama.Option{
		ollama.WithModel(config.Model),
		ollama.WithServerURL(config.BaseURL),
	}

	m, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 OpenAI LLM 失败: %w", err)
	}

	return &GoogleProvider{
		model:  m,
		config: config,
	}, nil
}

func (p *OllamaProvider) HandleResponse(response llms.ContentResponse) {

}

func (p *OllamaProvider) GetModel() llms.Model {
	return p.model
}
