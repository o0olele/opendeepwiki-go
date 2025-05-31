package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LLMConfig 语言模型配置
type LLMConfig struct {
	ProviderType string  `yaml:"provider_type"` // openai, google, deepseek, ollama, llamacpp, vllm
	APIKey       string  `yaml:"api_key"`
	Model        string  `yaml:"model"`
	BaseURL      string  `yaml:"base_url"`
	MaxTokens    int     `yaml:"max_tokens"`
	Temperature  float64 `yaml:"temperature"`
}

type EmbeddingConfig struct {
	ProviderType string `yaml:"provider_type"` // openai, google, deepseek, ollama, llamacpp, vllm
	APIKey       string `yaml:"api_key"`
	Model        string `yaml:"model"`
	BatchSize    int    `yaml:"batch_size"`
	BaseURL      string `yaml:"base_url"`
}

// Config holds all configuration for the application
type Config struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
	Repository struct {
		Dir string `yaml:"dir"`
	} `yaml:"repository"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	LLM       LLMConfig       `yaml:"llm"`
	Embedding EmbeddingConfig `yaml:"embedding"`
}

var cfg Config

// LoadConfig loads configuration from a YAML file
func LoadConfig() *Config {
	LoadTemplates()
	// Default configuration file path
	configPath := "config.yaml"

	// Check if config file path is provided via environment variable
	if envPath := os.Getenv("CONFIG_FILE"); envPath != "" {
		configPath = envPath
	}

	// Create a new config with default values
	config := &cfg
	config.Server.Address = ":8080"
	config.Repository.Dir = "./repos"
	config.Database.Path = "./data/opendeepwiki.db"

	// 默认 LLM 配置
	config.LLM.ProviderType = "openai"
	config.LLM.Model = "gpt-4"
	config.LLM.MaxTokens = 8192
	config.LLM.Temperature = 0.5

	// Try to read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		// If the file doesn't exist, create it with default values
		if os.IsNotExist(err) {
			fmt.Printf("Configuration file not found at %s, creating with default values\n", configPath)
			createDefaultConfig(configPath, config)
		} else {
			fmt.Printf("Warning: Could not read config file: %v\n", err)
		}
	} else {
		// Parse the YAML file
		if err := yaml.Unmarshal(data, config); err != nil {
			fmt.Printf("Warning: Could not parse config file: %v\n", err)
		}
	}

	// Create repository directory if it doesn't exist
	if _, err := os.Stat(config.Repository.Dir); os.IsNotExist(err) {
		err := os.MkdirAll(config.Repository.Dir, 0755)
		if err != nil {
			panic("Failed to create repository directory: " + err.Error())
		}
	}

	fmt.Println(config.LLM.Model, config.LLM.BaseURL, config.LLM.APIKey)

	return config
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string, config *Config) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Warning: Could not create config directory: %v\n", err)
			return
		}
	}

	// Marshal the default config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Warning: Could not marshal default config: %v\n", err)
		return
	}

	// Add header comment
	content := "# OpenDeepWiki Configuration\n\n" + string(data)

	// Write the file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Printf("Warning: Could not write default config file: %v\n", err)
	}
}

func GetLLMConfig() *LLMConfig {
	return &cfg.LLM
}

func GetEmbeddingConfig() *EmbeddingConfig {
	return &cfg.Embedding
}
