package chat

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

// GoogleProvider Google 提供商实现
type GoogleProvider struct {
	model  llms.Model
	config *ProviderConfig
}

// NewGoogleProvider 创建一个新的 Google 提供商
func NewGoogleProvider(config *ProviderConfig) (Provider, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("未提供 OpenAI API 密钥")
		}
	}

	// 创建 OpenAI 选项
	opts := []googleai.Option{
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel(config.Model),
	}

	m, err := googleai.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 OpenAI LLM 失败: %w", err)
	}

	return &GoogleProvider{
		model:  m,
		config: config,
	}, nil
}

func (p *GoogleProvider) HandleResponse(response llms.ContentResponse) {

}

func (p *GoogleProvider) GetModel() llms.Model {
	return p.model
}
