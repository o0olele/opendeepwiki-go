package analyzer

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/o0olele/opendeepwiki-go/internal/config"
	"github.com/o0olele/opendeepwiki-go/internal/llm/chat"
	"github.com/o0olele/opendeepwiki-go/internal/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"go.uber.org/zap"
)

// DefaultExcludedFiles 默认排除的文件和目录
var DefaultExcludedFiles = []string{
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
	"*.exe",
	"*.dll",
	"*.so",
	"*.dylib",
	"*.test",
	"*.out",
	"*.log",
}

// FileScanner Scans files and directories in a repository
type FileScanner struct {
	options *AnalyzeOptions
}

// NewFileScanner Creates a new FileScanner with the given options.
func NewFileScanner(options *AnalyzeOptions) *FileScanner {
	if options == nil {
		options = &AnalyzeOptions{
			EnableSmartFilter: true,
			ExcludedFiles:     DefaultExcludedFiles,
			MaxFileSize:       1024 * 1024, // 1MB
			MaxTokens:         8192,
			Language:          "english",
		}
	}
	return &FileScanner{options: options}
}

// GetIgnorePatterns get the ignore patterns from .gitignore file
func (fs *FileScanner) GetIgnorePatterns(repoPath string) []string {
	var ignorePatterns []string
	var gitignorePath = filepath.Join(repoPath, ".gitignore")
	// add default patterns
	ignorePatterns = append(ignorePatterns, fs.options.ExcludedFiles...)
	// read .gitignore file
	if _, err := os.Stat(gitignorePath); err != nil {
		return ignorePatterns // no.gitignore file, return default patterns
	}

	file, err := os.Open(gitignorePath)
	if err != nil {
		zap.L().Error("cannot open .gitignore file", zap.Error(err))
		return ignorePatterns
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ignorePatterns = append(ignorePatterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		zap.L().Error("cannot read.gitignore file", zap.Error(err))
	}

	return ignorePatterns
}

// ScanDirectory Scans the directory and returns a list of files and directories.
func (fs *FileScanner) ScanDirectory(repoPath string, ignorePatterns []string) ([]PathInfo, error) {
	var pathInfos []PathInfo

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// get relative path
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}

		// ignore root directory
		if relPath == "." {
			return nil
		}

		// to Unix style path
		relPath = filepath.ToSlash(relPath)

		// check if the file or directory should be ignored
		if fs.shouldIgnore(relPath, ignorePatterns, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// check if the file size exceeds the limit
		if !info.IsDir() && info.Size() > fs.options.MaxFileSize {
			return nil
		}

		// add to pathInfos
		fileType := "File"
		if info.IsDir() {
			fileType = "Directory"
		}

		pathInfos = append(pathInfos, PathInfo{
			Path: path,
			Name: info.Name(),
			Type: fileType,
		})

		return nil
	})

	return pathInfos, err
}

// shouldIgnore checks if the file or directory should be ignored.
func (fs *FileScanner) shouldIgnore(path string, patterns []string, isDir bool) bool {
	// if the file or directory is hidden, ignore it
	if strings.HasPrefix(filepath.Base(path), ".") {
		return true
	}

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		// handle directory patterns
		isDirPattern := strings.HasSuffix(pattern, "/")
		if isDirPattern {
			pattern = strings.TrimSuffix(pattern, "/")
			if !isDir {
				continue
			}
		}

		// handle file patterns
		if strings.Contains(pattern, "*") {
			regexPattern := "^" + regexp.QuoteMeta(pattern)
			regexPattern = strings.Replace(regexPattern, "\\*", ".*", -1)
			regexPattern += "$"

			matched, err := regexp.MatchString(regexPattern, path)
			if err == nil && matched {
				return true
			}

			// check if the directory matches the pattern
			parts := strings.Split(path, "/")
			for _, part := range parts {
				matched, err := regexp.MatchString(regexPattern, part)
				if err == nil && matched {
					return true
				}
			}
		} else if path == pattern || filepath.Base(path) == pattern {
			// check if the directory matches the pattern
			return true
		}
	}

	return false
}

// GetCatalogue get the catalogue of the repository.
func (fs *FileScanner) GetCatalogue(repoPath string) ([]PathInfo, error) {
	var ignorePatterns = fs.GetIgnorePatterns(repoPath)
	// get the files and directories
	pathInfos, err := fs.ScanDirectory(repoPath, ignorePatterns)
	if err != nil {
		zap.L().Error("cannot scan directory", zap.Error(err))
		return nil, err
	}
	return pathInfos, nil
}

func (fs *FileScanner) GetSimplifyCatalogueString(provider chat.Provider, repoPath string, catalogs []PathInfo, readme string) (string, error) {

	if len(catalogs) < 800 || !fs.options.EnableSmartFilter {
		return CatalogueToString(repoPath, catalogs), nil
	}

	var prompt = prompts.PromptTemplate{
		Template: config.SimplifyDirsPrompt,
		PartialVariables: map[string]any{
			"code_files": CatalogueToString(repoPath, catalogs),
			"readme":     readme,
			"language":   fs.options.Language,
		},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	content, err := prompt.Format(nil)
	if err != nil {
		zap.L().Error("cannot format prompt", zap.Error(err))
		return "", err
	}

	message := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, content),
	}

	response, err := provider.GetModel().GenerateContent(context.Background(), message,
		llms.WithMaxTokens(16384),
	)
	if err != nil {
		zap.L().Error("cannot get model response", zap.Error(err))
		return "", err
	}

	filesContent := utils.ExtractTagContent(response.Choices[0].Content, "response_file")
	filesContent = utils.ExtractJSON(filesContent)
	return filesContent, nil
}

// GetCatalogueString get the catalogue of the repository in string format.
func (fs *FileScanner) GetCatalogueString(repoPath string) (string, error) {
	pathInfos, err := fs.GetCatalogue(repoPath)
	if err != nil {
		return "", err
	}

	return CatalogueToString(repoPath, pathInfos), nil
}

func CatalogueToString(repoPath string, pathInfos []PathInfo) string {
	// build the catalogue string
	var sb strings.Builder
	for _, info := range pathInfos {
		// get relative path
		relPath, err := filepath.Rel(repoPath, info.Path)
		if err != nil {
			continue
		}

		// change to Unix style path
		relPath = filepath.ToSlash(relPath)

		sb.WriteString(relPath)
		sb.WriteString("\n")
	}
	return sb.String()
}

// ReadFile reads the content of a file.
func ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
