package analyzer

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"
	"go.uber.org/zap"
)

// FileParser parses files and extracts information
type FileParser struct {
	supportedExtensions map[string]bool // supported file extensions
	ignoreDirs          []string        // ignore directories
}

// NewFileParser 创建一个新的文件解析器
func NewFileParser() *FileParser {
	// 定义支持的文件扩展名
	supportedExts := map[string]bool{
		".go":    true,
		".py":    true,
		".js":    true,
		".ts":    true,
		".jsx":   true,
		".tsx":   true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".hpp":   true,
		".cs":    true,
		".rb":    true,
		".php":   true,
		".swift": true,
		".kt":    true,
		".rs":    true,
		".scala": true,
		".sh":    true,
		".bat":   true,
		".ps1":   true,
		".html":  true,
		".css":   true,
		".md":    true,
		".json":  true,
		".yaml":  true,
		".yml":   true,
		".xml":   true,
		".sql":   true,
	}

	// 定义要忽略的目录
	ignoreDirs := []string{
		"node_modules",
		"vendor",
		"dist",
		"build",
		".git",
		".github",
		".idea",
		".vscode",
		"__pycache__",
		".pytest_cache",
		"coverage",
	}

	return &FileParser{
		supportedExtensions: supportedExts,
		ignoreDirs:          ignoreDirs,
	}
}

// ParseFile parses a file and extracts information, including line count and file data.
func (fp *FileParser) ParseFile(file *object.File) (*FileInfo, error) {
	// check if the file extension is supported, if not, return an error message.
	ext := strings.ToLower(filepath.Ext(file.Name))
	if !fp.supportedExtensions[ext] {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	// check if the file is in the ignore directory, if yes, return an error message.
	for _, dir := range fp.ignoreDirs {
		if strings.Contains(file.Name, "/"+dir+"/") || strings.HasPrefix(file.Name, dir+"/") {
			return nil, fmt.Errorf("file in ignores: %s", file.Name)
		}
	}

	// get the file contents
	reader, err := file.Reader()
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read file content failed: %w", err)
	}

	// create a FileInfo struct to store the file information
	fileInfo := &FileInfo{
		Name:      file.Name,
		Path:      file.Name,
		Ext:       ext,
		Size:      file.Size,
		Content:   string(content),
		LineCount: countLines(string(content)),
	}

	zap.L().Info("Parsed file", zap.String("name", fileInfo.Name), zap.Int("lineCount", fileInfo.LineCount))
	return fileInfo, nil
}

// FileInfo 文件信息
type FileInfo struct {
	Name      string // 文件名
	Path      string // 文件路径
	Ext       string // 文件扩展名
	Size      int64  // 文件大小（字节）
	Content   string // 文件内容
	LineCount int    // 行数
}

// countLines 计算文本的行数
func countLines(text string) int {
	return len(strings.Split(text, "\n"))
}
