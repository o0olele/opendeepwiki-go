package analyzer

import (
	"time"
)

// PathInfo 表示文件或目录的路径信息
type PathInfo struct {
	Path string // 文件或目录的完整路径
	Name string // 文件或目录的名称
	Type string // 类型：File 或 Directory
}

// CatalogueItem 表示目录结构中的一个项目
type CatalogueItem struct {
	Title          string          `json:"title"`           // 标题
	Name           string          `json:"name"`            // 名称
	Prompt         string          `json:"prompt"`          // 提示信息
	DependentFiles []string        `json:"dependent_files"` // 依赖文件
	Children       []CatalogueItem `json:"children"`        // 子项目
}

// DocumentAnalysis 表示文档分析结果
type DocumentAnalysis struct {
	Overview  string          `json:"overview"`  // 项目概述
	Structure string          `json:"structure"` // 目录结构
	Catalogue []CatalogueItem `json:"catalogue"` // 目录项目
}

// CommitRecord 表示提交记录
type CommitRecord struct {
	Title       string    `json:"title"`       // 标题
	Description string    `json:"description"` // 描述
	Date        time.Time `json:"date"`        // 日期
	Author      string    `json:"author"`      // 作者
}

// WikiDocument 表示生成的 Wiki 文档
type WikiDocument struct {
	ID          string    `json:"id"`          // 文档 ID
	Title       string    `json:"title"`       // 标题
	Content     string    `json:"content"`     // 内容
	Description string    `json:"description"` // 描述
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间
}

// AnalyzeOptions 表示 Wiki 生成选项
type AnalyzeOptions struct {
	EnableSmartFilter bool     `json:"enable_smart_filter"` // 是否启用智能过滤
	ExcludedFiles     []string `json:"excluded_files"`      // 排除的文件
	MaxFileSize       int64    `json:"max_file_size"`       // 最大文件大小（字节）
	MaxTokens         int      `json:"max_tokens"`          // 最大令牌数
}

type DocumentResultCalalogue struct {
	Items []DocumentResultCalalogueItem `json:"items"`
}

type DocumentResultCalalogueItem struct {
	Name           string                        `json:"name"`
	Title          string                        `json:"title"`
	Prompt         string                        `json:"prompt"`
	DependentFiles []string                      `json:"dependent_file"`
	Children       []DocumentResultCalalogueItem `json:"children"`
}
