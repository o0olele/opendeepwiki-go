package repository

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/o0olele/opendeepwiki-go/internal/database/dao"
)

// RepositoryHandler Warehouse handler.
type RepositoryHandler struct {
	taskDao *dao.RepositoryTaskDAO
	repoDao *dao.RepositoryDAO
}

// NewRepositoryHandler Create a new warehouse handler.
func NewRepositoryHandler() *RepositoryHandler {
	return &RepositoryHandler{
		taskDao: dao.NewRepositoryTaskDAO(),
		repoDao: dao.NewRepositoryDAO(),
	}
}

// CreateRepository Create a new repository task.
func (h *RepositoryHandler) CreateRepository(c *gin.Context) {
	// 解析请求
	var req struct {
		GitURL string `json:"git_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// 验证 Git URL
	if !isValidGitURL(req.GitURL) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Git URL format",
		})
		return
	}

	// 检查是否已存在相同的仓库任务
	if task, err := h.taskDao.GetRepositoryTaskByGitURL(req.GitURL); err == nil {
		// 已存在进行中的任务
		c.JSON(http.StatusOK, gin.H{
			"existing": true,
			"message":  "Repository already being processed",
			"task_id":  task.ID,
			"status":   task.StatusString(),
		})
		return
	}

	task, err := h.taskDao.CreateRepositoryTask(req.GitURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task: " + err.Error(),
		})
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Repository submitted for processing",
		"task_id": task.ID,
	})
}

func (h *RepositoryHandler) GetRepositoryList(c *gin.Context) {
	// TODO memory cache repository list

	// get repository list from database
	repositories, err := h.repoDao.ListRepositories(10, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get repository list: " + err.Error(),
		})
		return
	}

	var list []*Repository

	list = append(list, &Repository{
		RepoId:      1,
		URL:         "https//github.com/o0olele/opendeepwiki-go.git",
		Name:        "opendeepwiki-go",
		Description: "OpenDeepWiki Go",
	})
	list = append(list, &Repository{
		RepoId:      2,
		URL:         "https//github.com/gin-gonic/gin.git",
		Name:        "gin",
		Description: "gin web framework",
	})

	for _, repo := range repositories {
		list = append(list, &Repository{
			RepoId:      repo.ID,
			URL:         repo.GitURL,
			Name:        repo.Name,
			Description: repo.Description,
			Status:      repo.WebStatus(),
		})
	}

	c.JSON(http.StatusOK, list)
}

func (h *RepositoryHandler) GetRepositoryById(c *gin.Context) {
	var req struct {
		Id uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
	}

	repo, err := h.repoDao.GetRepositoryByID(req.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get repository: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &Repository{
		RepoId:      repo.ID,
		URL:         repo.GitURL,
		Name:        repo.Name,
		Description: repo.Description,
		Status:      repo.WebStatus(),
	})
}

// RegisterRoutes Register repository routes.
func RegisterRoutes(router *gin.RouterGroup) {
	handler := NewRepositoryHandler()

	group := router.Group("/repo")
	group.POST("/create", handler.CreateRepository)
	group.GET("/list", handler.GetRepositoryList)
	group.GET("/status", handler.GetRepositoryById)
}

// isValidGitURL Verify that the Git URL format is correct.
func isValidGitURL(url string) bool {
	// here you can add more validation logic if needed
	return len(url) > 8 && (url[:8] == "https://" ||
		url[:7] == "http://" ||
		url[:6] == "git://" ||
		url[:4] == "ssh:")
}
