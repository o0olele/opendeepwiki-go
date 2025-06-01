package document

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/o0olele/opendeepwiki-go/internal/database/dao"
)

// DocumentHandler Document handler.
type DocumentHandler struct {
	docDao  *dao.DocumentDao
	repoDao *dao.RepositoryDAO
}

// NewDocumentHandler Create a new document handler.
func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{
		docDao:  dao.NewDocumentDao(),
		repoDao: dao.NewRepositoryDAO(),
	}
}

func (h *DocumentHandler) GetOverview(c *gin.Context) {
	repoIdStr := c.Param("id")
	if repoIdStr == "" {
		c.JSON(400, gin.H{"error": "repository id is required"})
		return
	}

	repoId, err := strconv.ParseUint(repoIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid repository id"})
		return
	}

	repo, err := h.repoDao.GetRepositoryByID(uint(repoId)) // TODO: use repoNam
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	docs, err := h.docDao.GetDocumentByRepoId(repo.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, NewOverview(repo.ID, repo.Overview, docs))
	return
}

func (h *DocumentHandler) GetDetail(c *gin.Context) {
	repoIdStr := c.Query("id")
	if repoIdStr == "" {
		c.JSON(400, gin.H{"error": "repository id is required"})
		return
	}
	repoId, err := strconv.ParseUint(repoIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid repository id"})
	}
	indexStr := c.Query("index")
	if indexStr == "" {
		c.JSON(400, gin.H{"error": "index is required"})
		return
	}
	index, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid index"})
	}

	doc, err := h.docDao.GetDocumentByRepoIdAndIndex(uint(repoId), int(index))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, &Document{
		ID:      doc.ID,
		Content: doc.Content,
		Title:   doc.Title,
	})
	return
}

// RegisterRoutes Register repository routes.
func RegisterRoutes(router *gin.RouterGroup) {
	handler := NewDocumentHandler()

	group := router.Group("/doc")
	group.GET("/:id", handler.GetOverview)
	group.GET("/detail", handler.GetDetail)
}
