package services

import (
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/o0olele/opendeepwiki-go/internal/analyzer"
	"github.com/o0olele/opendeepwiki-go/internal/database/dao"
	"github.com/o0olele/opendeepwiki-go/internal/database/models"
	"github.com/o0olele/opendeepwiki-go/internal/utils"
	"go.uber.org/zap"
)

// TaskProcessParams represents the parameters for processing a task.
type TaskProcessParams struct {
	taskDao *dao.RepositoryTaskDAO
	repoDir string
}

// Task represents a task to be processed.
type Task struct {
	ID        uint      `json:"id"` // task id
	GitURL    string    `json:"git_url"`
	CreatedAt time.Time `json:"created_at"`
	Status    int       `json:"status"`
}

// NewTaskFromModel creates a new Task from a models.RepositoryTask.
func NewTaskFromModel(task *models.RepositoryTask) *Task {
	return &Task{
		ID:        task.ID,
		GitURL:    task.GitURL,
		CreatedAt: task.CreatedAt,
		Status:    int(task.Status),
	}
}

// Process processes the task.
func (t *Task) Process(params *TaskProcessParams) {

	for {
		var breakFlag = false
		switch t.Status {
		case models.RepositoryTaskStatusPending:
			// first clone the repository to the local machine.
			err := t.Clone(params)
			if err != nil {
				break
			}
		case models.RepositoryTaskStatusCloned:
			// then process the repository.
			err := t.Analyze(params)
			if err != nil {
				break
			}
		default:
			breakFlag = true
		}
		if breakFlag ||
			t.Status == models.RepositoryTaskStatusCompleted ||
			t.Status == models.RepositoryTaskStatusFailed {
			break
		}
	}

}

// UpdateStatus updates the task status.
func (t *Task) UpdateStatus(params *TaskProcessParams, status int) error {
	dao := params.taskDao

	err := dao.UpdateRepositoryTaskStatus(t.ID, status)
	if err != nil {
		return err
	}

	t.Status = status
	return nil
}

// Clone clones the repository to the local machine.
func (t *Task) Clone(params *TaskProcessParams) error {

	// first check if the repository already exists.
	repoName, err := utils.ExtractRepoName(t.GitURL)
	if err != nil {
		zap.L().Error("Failed to extract repo name from git url: %v", zap.Error(err))
		return err
	}

	var repoPath = filepath.Join(params.repoDir, repoName)

	if _, err = os.Stat(repoPath); err == nil {
		_, err = git.PlainOpen(repoPath)
		if err == nil {
			// repo already exists, update status to cloned.
			return t.UpdateStatus(params, models.RepositoryTaskStatusCloned)
		} else {
			// repo exists but is not a valid git repo, remove it and clone again.
			zap.L().Info("Invalid git repository for task %s, removing and re-cloning", zap.Uint("task_id", t.ID), zap.String("git_url", t.GitURL))
			os.RemoveAll(repoPath)
		}
	}

	// clone the repository.
	_, err = git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:      t.GitURL,
		Progress: os.Stdout,
	})
	if err != nil {
		zap.L().Error("Failed to clone repository: %v", zap.Error(err))
		t.UpdateStatus(params, models.RepositoryTaskStatusFailed)
		return err
	}

	zap.L().Info("Cloned repository", zap.String("git_url", t.GitURL))
	return t.UpdateStatus(params, models.RepositoryTaskStatusCloned)
}

// Analyze analyzes the repository.
func (t *Task) Analyze(params *TaskProcessParams) error {

	r, err := analyzer.NewRepository(params.repoDir, t.GitURL)
	if err != nil {
		zap.L().Error("Failed to create repository: %v", zap.Error(err))
		t.UpdateStatus(params, models.RepositoryTaskStatusFailed)
		return err
	}

	err = r.Parse(nil)
	if err != nil {
		zap.L().Error("Failed to parse repository: %v", zap.Error(err))
		t.UpdateStatus(params, models.RepositoryTaskStatusFailed)
		return err
	}

	return nil
}
