package chat

import (
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAIProvider OpenAI 提供商实现
type OpenAIProvider struct {
	model  llms.Model
	config *ProviderConfig
}

// NewOpenAIProvider 创建一个新的 OpenAI 提供商
func NewOpenAIProvider(config *ProviderConfig) (Provider, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("未提供 OpenAI API 密钥")
		}
	}

	// 创建 OpenAI 选项
	opts := []openai.Option{
		openai.WithToken(apiKey),
		openai.WithModel(config.Model),
	}

	// 如果提供了自定义 URL，则使用它
	if config.BaseURL != "" {
		fmt.Println(config.BaseURL)
		opts = append(opts, openai.WithBaseURL(config.BaseURL))
	}

	m, err := openai.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("创建 OpenAI LLM 失败: %w", err)
	}

	return &OpenAIProvider{
		model:  m,
		config: config,
	}, nil
}

func (p *OpenAIProvider) HandleResponse(response llms.ContentResponse) {

}

func (p *OpenAIProvider) GetModel() llms.Model {
	return p.model
}
