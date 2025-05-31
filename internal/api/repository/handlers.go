package repository

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/o0olele/opendeepwiki-go/internal/database/dao"
)

// RepositoryHandler Warehouse handler.
type RepositoryHandler struct {
	taskDao *dao.RepositoryTaskDAO
}

// NewRepositoryHandler Create a new warehouse handler.
func NewRepositoryHandler() *RepositoryHandler {
	return &RepositoryHandler{
		taskDao: dao.NewRepositoryTaskDAO(),
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

// RegisterRoutes Register repository routes.
func RegisterRoutes(router *gin.RouterGroup) {
	handler := NewRepositoryHandler()

	group := router.Group("/repo")
	group.POST("/create", handler.CreateRepository)
}

// isValidGitURL Verify that the Git URL format is correct.
func isValidGitURL(url string) bool {
	// here you can add more validation logic if needed
	return len(url) > 8 && (url[:8] == "https://" ||
		url[:7] == "http://" ||
		url[:6] == "git://" ||
		url[:4] == "ssh:")
}
