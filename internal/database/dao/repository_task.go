package dao

import (
	"github.com/o0olele/opendeepwiki-go/internal/database"
	"github.com/o0olele/opendeepwiki-go/internal/database/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RepositoryTaskDAO Repository task data access object.
type RepositoryTaskDAO struct {
	db *gorm.DB
}

// NewRepositoryTaskDAO Create a new repository task data access object.
func NewRepositoryTaskDAO() *RepositoryTaskDAO {
	return &RepositoryTaskDAO{
		db: database.GetDB(),
	}
}

// CreateRepositoryTask Create a new repository task record.
func (dao *RepositoryTaskDAO) CreateRepositoryTask(gitURL string) (*models.RepositoryTask, error) {
	task := &models.RepositoryTask{
		GitURL: gitURL,
		Status: models.RepositoryStatusPending,
	}

	result := dao.db.Create(task)
	if result.Error != nil {
		zap.L().Error("Failed to create repository task: %v", zap.Error(result.Error))
		return nil, result.Error
	}

	return task, nil
}

// GetRepositoryTaskByID Get a repository task by ID.
func (dao *RepositoryTaskDAO) GetRepositoryTaskByID(id uint) (*models.RepositoryTask, error) {
	var task = new(models.RepositoryTask)
	result := dao.db.First(task, id)
	if result.Error != nil {
		zap.L().Error("Failed to get repository task by ID: %v", zap.Error(result.Error))
		return nil, result.Error
	}
	return task, nil
}

// GetRepositoryTaskByGitURL Get a repository task by Git URL.
func (dao *RepositoryTaskDAO) GetRepositoryTaskByGitURL(gitURL string) (*models.RepositoryTask, error) {
	var task = new(models.RepositoryTask)
	result := dao.db.Where("git_url =?", gitURL).First(task)
	if result.Error != nil {
		zap.L().Error("Failed to get repository task by gitURL: %v", zap.Error(result.Error))
		return nil, result.Error
	}
	return task, nil
}

// ListRepositoryTasksByStatus List repository tasks by status.
func (dao *RepositoryTaskDAO) ListRepositoryTasksByStatus(status int, limit, offset int) ([]*models.RepositoryTask, error) {
	var tasks []*models.RepositoryTask
	result := dao.db.Where("status =?", status).Limit(limit).Offset(offset).Find(&tasks)
	if result.Error != nil {
		zap.L().Error("Failed to list repository tasks by status: %v", zap.Error(result.Error))
		return nil, result.Error
	}
	return tasks, nil
}

// UpdateRepositoryTask Update a repository task.
func (dao *RepositoryTaskDAO) UpdateRepositoryTask(task *models.RepositoryTask) error {
	result := dao.db.Save(task)
	if result.Error != nil {
		zap.L().Error("Failed to update repository task: %v", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

// DeleteRepositoryTask Delete a repository task.
func (dao *RepositoryTaskDAO) DeleteRepositoryTask(id uint) error {
	result := dao.db.Delete(&models.RepositoryTask{}, id)
	if result.Error != nil {
		zap.L().Error("Failed to delete repository task: %v", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

func (dao *RepositoryTaskDAO) UpdateRepositoryTaskStatus(id uint, status int) error {
	result := dao.db.Model(&models.RepositoryTask{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		zap.L().Error("Failed to update repository task status: %v", zap.Error(result.Error))
		return result.Error
	}
	zap.L().Info("Update repository task status success", zap.Uint("task_id", id), zap.Int("status", status))
	return nil
}
