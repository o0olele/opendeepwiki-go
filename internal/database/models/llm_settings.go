package models

import (
	"gorm.io/gorm"
)

// LLMSettings 语言模型设置
type LLMSettings struct {
	gorm.Model
	ProviderType string  `json:"provider_type"` // openai, google, deepseek, ollama, llamacpp, vllm
	APIKey       string  `json:"api_key"`
	ModelLLM     string  `json:"model_llm"`
	BaseURL      string  `json:"base_url"`
	MaxTokens    int     `json:"max_tokens"`
	Temperature  float64 `json:"temperature"`
	IsDefault    bool    `json:"is_default"` // 是否为默认设置
}
