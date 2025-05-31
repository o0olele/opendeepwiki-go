package models

import (
	"time"
)

// Document 文档模型
type Document struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TaskID      string    `json:"task_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Content     string    `gorm:"type:text" json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CodeAnalysis 代码分析结果模型
type CodeAnalysis struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TaskID         string    `json:"task_id"`
	FileCount      int       `json:"file_count"`
	DirectoryCount int       `json:"directory_count"`
	FileTypes      string    `gorm:"type:text" json:"file_types"` // JSON 格式的文件类型统计
	CreatedAt      time.Time `json:"created_at"`
}
