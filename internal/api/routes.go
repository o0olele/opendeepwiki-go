package api

import (
	"github.com/gin-gonic/gin"
	"github.com/o0olele/opendeepwiki-go/internal/api/document"
	"github.com/o0olele/opendeepwiki-go/internal/api/repository"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(router *gin.Engine) {
	// API group
	apiGroup := router.Group("/api")

	// Register warehouse routes
	repository.RegisterRoutes(apiGroup)

	// Register other routes here
	document.RegisterRoutes(apiGroup)
}
