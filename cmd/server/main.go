package main

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/o0olele/opendeepwiki-go/internal/api"
	"github.com/o0olele/opendeepwiki-go/internal/config"
	"github.com/o0olele/opendeepwiki-go/internal/database"
	"github.com/o0olele/opendeepwiki-go/internal/services"
	"go.uber.org/zap"
)

func main() {

	logger := zap.Must(zap.NewProduction())
	zap.ReplaceGlobals(logger)
	// Load configuration from YAML file
	cfg := config.LoadConfig()

	// Initialize database
	if err := database.InitDB(cfg.Database.Path); err != nil {
		zap.L().Error("Failed to initialize database", zap.Error(err))
	}

	llmSettings, err := database.GetLLMSettings()
	if err != nil {
		zap.L().Error("Failed to get LLM settings", zap.Error(err))
	}

	cfg.LLM.ProviderType = llmSettings.ProviderType
	cfg.LLM.APIKey = llmSettings.APIKey
	cfg.LLM.Model = llmSettings.ModelLLM
	cfg.LLM.MaxTokens = llmSettings.MaxTokens
	cfg.LLM.Temperature = llmSettings.Temperature

	// Initialize task queue with config
	taskQueue := services.NewTaskQueue(cfg.Repository.Dir)
	if taskQueue == nil {
		zap.L().Error("Failed to initialize task queue")
		return
	}

	// Start task processor in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		taskQueue.ProcessTasks()
	}()

	// Setup Gin router
	router := gin.Default()

	// Register API routes
	api.RegisterRoutes(router)

	// Start server
	zap.L().Info("Starting server", zap.String("address", cfg.Server.Address)) // Use zap for logging in main.go instead of in api.go or htt
	if err := router.Run(cfg.Server.Address); err != nil {
		zap.L().Error("Failed to start server", zap.Error(err))
	}

	// Wait for task processor to finish (this won't happen in normal operation)
	wg.Wait()
}
