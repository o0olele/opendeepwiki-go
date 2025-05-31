package database

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/o0olele/opendeepwiki-go/internal/config"
	"github.com/o0olele/opendeepwiki-go/internal/database/models"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// InitDB Initialize the database connection and perform migrations.
func InitDB(dbPath string) error {
	var err error
	once.Do(func() {
		zap.L().Info("initializing database: ", zap.String("dbPath", dbPath))

		// make sure the database directory exists
		dbDir := filepath.Dir(dbPath)
		if _, err = os.Stat(dbDir); os.IsNotExist(err) {
			if err = os.MkdirAll(dbDir, 0755); err != nil {
				zap.L().Error("Failed to create database directory: %v", zap.Error(err))
				return
			}
		}

		// configure GORM logger to log SQL queries in inf
		config := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}

		// open the database connection
		db, err = gorm.Open(sqlite.Open(dbPath), config)
		if err != nil {
			zap.L().Error("Failed to open database: %v", zap.Error(err))
			return
		}

		// migrate the database models
		err = migrateModels()
		if err != nil {
			zap.L().Error("Failed to migrate database models: %v", zap.Error(err))
			return
		}

		// init default LLM settings
		err = initDefaultLLMSettings()
		if err != nil {
			zap.L().Error("Failed to init default LLM settings: %v", zap.Error(err))
			return
		}

		zap.L().Info("Database initialized successfully")
	})
	return err
}

// migrateModels migrate the database models
func migrateModels() error {
	zap.L().Info("migrating database models...")
	return db.AutoMigrate(
		&models.Repository{},
		&models.RepositoryTask{},
		&models.Document{},
		&models.CodeAnalysis{},
		&models.DocumentCommitRecord{},
		&models.LLMSettings{},
	)
}

// initDefaultLLMSettings initialize default LLM settings
func initDefaultLLMSettings() error {
	var count int64
	db.Model(&models.LLMSettings{}).Count(&count)

	// if no settings exist, create default settings
	if count == 0 {
		llm := config.GetLLMConfig()
		defaultSettings := models.LLMSettings{
			ProviderType: llm.ProviderType,
			ModelLLM:     llm.Model,
			MaxTokens:    llm.MaxTokens,
			Temperature:  llm.Temperature,
			IsDefault:    true,
			APIKey:       llm.APIKey,
			BaseURL:      llm.BaseURL,
		}

		return db.Create(&defaultSettings).Error
	}

	return nil
}

// GetDB get the database connection
func GetDB() *gorm.DB {
	if db == nil {
		zap.L().Error("database not initialized")
	}
	return db
}

// GetLLMSettings 获取 LLM 设置
func GetLLMSettings() (*models.LLMSettings, error) {
	var settings models.LLMSettings

	// 首先尝试获取默认设置
	result := db.Where("is_default = ?", true).First(&settings)
	if result.Error == nil {
		return &settings, nil
	}

	// 如果没有默认设置，获取第一个设置
	result = db.First(&settings)
	if result.Error != nil {
		return nil, result.Error
	}

	return &settings, nil
}
