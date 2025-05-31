package dao

import (
	"github.com/o0olele/opendeepwiki-go/internal/database"
	"github.com/o0olele/opendeepwiki-go/internal/database/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RepositoryTaskDAO Repository data access object.
type RepositoryDAO struct {
	db *gorm.DB
}

// NewRepositoryDAO Create a new repository data access object.
func NewRepositoryDAO() *RepositoryDAO {
	return &RepositoryDAO{
		db: database.GetDB(),
	}
}

// CreateRepository Create a new repository record.
func (dao *RepositoryDAO) CreateRepository(gitURL string, name string, description string) (*models.Repository, error) {
	repo := &models.Repository{
		GitURL:      gitURL,
		Name:        name,
		Description: description,
	}
	result := dao.db.Create(repo)
	if result.Error != nil {
		zap.L().Error("Failed to create repository: %v", zap.Error(result.Error))
		return nil, result.Error
	}

	return repo, nil
}

// GetRepositoryByID Get a repository by ID.
func (dao *RepositoryDAO) GetRepositoryByID(id uint) (*models.Repository, error) {
	var repo = new(models.Repository)
	result := dao.db.First(repo, id)
	if result.Error != nil {
		zap.L().Error("Failed to get repository by ID: %v", zap.Error(result.Error))
		return nil, result.Error
	}
	return repo, nil
}

// GetRepositoryByGitURL Get a repository by Git URL.
func (dao *RepositoryDAO) GetRepositoryByGitURL(gitURL string) (*models.Repository, error) {
	var repo = new(models.Repository)
	result := dao.db.Where("git_url = ?", gitURL).First(repo)
	if result.Error != nil {
		zap.L().Error("Failed to get repository by gitURL: %v", zap.Error(result.Error))
		return nil, result.Error
	}
	return repo, nil
}

// ListRepositories List all repositories.
func (dao *RepositoryDAO) ListRepositories(limit, offset int) ([]models.Repository, error) {
	var repos []models.Repository
	result := dao.db.Limit(limit).Offset(offset).Find(&repos)
	if result.Error != nil {
		zap.L().Error("Failed to list repositories: %v", zap.Error(result.Error))
		return nil, result.Error
	}
	return repos, nil
}

// UpdateRepository Update a repository record.
func (dao *RepositoryDAO) UpdateRepository(repo *models.Repository) error {
	result := dao.db.Save(repo)
	return result.Error
}

// DeleteRepository Delete a repository record.
func (dao *RepositoryDAO) DeleteRepository(id uint) error {
	result := dao.db.Delete(&models.Repository{}, id)
	return result.Error
}
