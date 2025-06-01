package models

import (
	"gorm.io/gorm"
)

// Document 文档模型
type Document struct {
	gorm.Model
	RepoId      uint   `json:"repo_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `gorm:"type:text" json:"content"`
	ParentId    uint   `json:"parent_id"`
	Index       int    `json:"index"`
}
