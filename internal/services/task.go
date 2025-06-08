package services

import (
	"encoding/json"
	"fmt"
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
	repoDao *dao.RepositoryDAO
	repoDir string
}

// Task represents a task to be processed.
type Task struct {
	ID        uint      `json:"id"` // task id
	GitURL    string    `json:"git_url"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	Status    int       `json:"status"`
}

// NewTaskFromModel creates a new Task from a models.RepositoryTask.
func NewTaskFromModel(task *models.RepositoryTask) *Task {
	return &Task{
		ID:        task.ID,
		GitURL:    task.GitURL,
		Language:  task.Language,
		CreatedAt: task.CreatedAt,
		Status:    int(task.Status),
	}
}

// Process processes the task.
func (t *Task) Process(params *TaskProcessParams) {

	for {
		var breakFlag = false
		switch t.Status {
		case models.RepositoryStatusPending:
			// first clone the repository to the local machine.
			err := t.Clone(params)
			if err != nil {
				break
			}
		case models.RepositoryStatusCloned:
			// then process the repository.
			err := t.Analyze(params)
			if err != nil {
				break
			}
		default:
			breakFlag = true
		}
		if breakFlag ||
			t.Status == models.RepositoryStatusCompleted ||
			t.Status == models.RepositoryStatusFailed {
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
			return t.UpdateStatus(params, models.RepositoryStatusCloned)
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
		t.UpdateStatus(params, models.RepositoryStatusFailed)
		return err
	}

	zap.L().Info("Cloned repository", zap.String("git_url", t.GitURL))
	err = t.UpdateStatus(params, models.RepositoryStatusCloned)
	if err != nil {
		return err
	}

	params.repoDao.CreateRepository(t.GitURL, repoName, repoPath, models.RepositoryStatusCloned, t.Language)
	return nil
}

// Analyze analyzes the repository.
func (t *Task) Analyze(params *TaskProcessParams) error {

	var r *analyzer.Repository
	// first check if the repository already exists.
	repoModal, err := params.repoDao.GetRepositoryByGitURL(t.GitURL)
	if err == nil {
		// repo already exists, update status to analyzed.
		zap.L().Info("Repository already exists", zap.Uint("repository_id", repoModal.ID))
		r, err = analyzer.NewRepositoryFromModel(repoModal)
	} else {
		zap.L().Info("Repository not found", zap.String("git_url", t.GitURL))
		return fmt.Errorf("repository not found")
	}

	if err != nil {
		zap.L().Error("Failed to create repository: %v", zap.Error(err))
		t.UpdateStatus(params, models.RepositoryStatusFailed)
		return err
	}

	// get the repository description
	if len(r.Description) == 0 {
		r.Description, err = r.ParseDescription()
		if err != nil {
			zap.L().Error("parse repository description failed", zap.Error(err))
			return err
		}
		params.repoDao.UpdateRepositoryDescription(repoModal.ID, r.Description)
	}

	// get the repository readme content
	if len(r.Readme) == 0 {
		r.Readme, err = r.ParseReadme()
		if err != nil || len(r.Readme) == 0 {
			// generate README content if not found or failed to parse
			r.Readme, err = r.GenerateReadme()
		}
		params.repoDao.UpdateRepositoryReadme(repoModal.ID, r.Readme)
	}

	// get the repository catalog string
	if len(r.StructedCatalogue) == 0 {
		r.StructedCatalogue, err = r.GenerateStructedCatalogue()
		if err != nil {
			zap.L().Error("get repository catalog failed", zap.Error(err))
			return err
		}
		params.repoDao.UpdateRepositoryCatalogue(repoModal.ID, r.StructedCatalogue)
	}

	// generate the repository overview
	if len(r.Overview) == 0 {
		r.Overview, err = r.GenerateOverview()
		if err != nil {
			zap.L().Error("generate repository overview failed", zap.Error(err))
			return err
		}
		params.repoDao.UpdateRepositoryOverview(repoModal.ID, r.Overview)
	}

	savePath, err := r.IndexCode()
	if err != nil {
		zap.L().Error("Failed to index repository: %v", zap.Error(err))
		t.UpdateStatus(params, models.RepositoryStatusFailed)
		return err
	}

	if savePath {
		params.repoDao.UpdateRepositoryCodePath(repoModal.ID, r.StructedCodePath)
		params.repoDao.UpdateRepositoryVectorPath(repoModal.ID, r.StructedVectorPath)
	}

	doc, err := r.CreateDocuments()
	if err != nil {
		zap.L().Error("Failed to create documents: %v", zap.Error(err))
		t.UpdateStatus(params, models.RepositoryStatusFailed)

		s, _ := json.Marshal(doc)
		os.WriteFile(r.Name+".json", s, 0644)

		t.dumpDocuments(repoModal.ID, doc)
		return err
	}

	t.dumpDocuments(repoModal.ID, doc)
	t.UpdateStatus(params, models.RepositoryStatusCompleted)
	return nil
}

func (t *Task) dumpDocuments(repoId uint, doc *analyzer.WikiDocument) error {
	docModels := []*models.Document{}
	docModels = wikiDocumentToModel(doc, docModels, repoId, 0)

	docDao := dao.NewDocumentDao()
	docDao.CreateDocuments(docModels)
	return nil
}

func wikiDocumentToModel(doc *analyzer.WikiDocument, list []*models.Document, repoId uint, depth int) []*models.Document {

	tmp := &models.Document{
		Title:    doc.Title,
		Content:  doc.Content,
		RepoId:   repoId,
		Index:    len(list) + 1,
		ParentId: uint(doc.ParentId),
	}
	list = append(list, tmp)

	for _, child := range doc.Children {
		child.ParentId = tmp.Index
		list = wikiDocumentToModel(child, list, repoId, depth+1)
	}

	return list
}
