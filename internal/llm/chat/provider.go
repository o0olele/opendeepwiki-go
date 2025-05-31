package chat

import (
	"github.com/tmc/langchaingo/llms"
)

// ProviderType 表示 LLM 提供商类型
type ProviderType string

const (
	ProviderOpenAI   ProviderType = "openai"
	ProviderGoogle   ProviderType = "google"
	ProviderDeepseek ProviderType = "deepseek"
	ProviderOllama   ProviderType = "ollama"
	ProviderLlamaCPP ProviderType = "llamacpp"
	ProviderVLLM     ProviderType = "vllm"
)

// ProviderConfig LLM 提供商配置
type ProviderConfig struct {
	Type        ProviderType `json:"type"`        // 提供商类型
	APIKey      string       `json:"api_key"`     // API 密钥
	Model       string       `json:"model"`       // 模型名称
	BaseURL     string       `json:"base_url"`    // 基础 URL（对于自定义端点）
	MaxTokens   int          `json:"max_tokens"`  // 最大令牌数
	Temperature float64      `json:"temperature"` // 温度
}

// Provider LLM 提供商接口
type Provider interface {
	// GenerateProjectOverview 生成项目概述
	HandleResponse(response llms.ContentResponse)
	GetModel() llms.Model
}

// NewProvider 创建一个新的 LLM 提供商
func NewProvider(config *ProviderConfig) (Provider, error) {
	switch config.Type {
	case ProviderOpenAI:
		return NewOpenAIProvider(config)
	case ProviderGoogle:
		return NewGoogleProvider(config)
	case ProviderOllama:
		return NewOllamaProvider(config)
	default:
		// 默认使用 OpenAI
		return NewOpenAIProvider(config)
	}
}
