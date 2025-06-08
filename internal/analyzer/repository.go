package analyzer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/google/uuid"
	"github.com/o0olele/opendeepwiki-go/internal/analyzer/codemap"
	"github.com/o0olele/opendeepwiki-go/internal/config"
	"github.com/o0olele/opendeepwiki-go/internal/database/models"
	"github.com/o0olele/opendeepwiki-go/internal/llm/chat"
	"go.uber.org/zap"
)

// Repository represents a local code repository.
type Repository struct {
	Path               string // path to the local repository
	GitURL             string // git URL of the repository
	Name               string // name of the repository
	Description        string // description of the repository
	Branch             string // branch of the repository
	Readme             string // README content of the repository
	Overview           string // overview of the repository
	StructedCatalogue  string // structured catalogue of the repository
	StructedCodePath   string
	StructedVectorPath string
	Language           string // language of the repository

	// internal fields
	repo        *git.Repository
	codeIndexer *codemap.CodeMapService
	fileScanner *FileScanner
	catalogs    []PathInfo
	provider    chat.Provider
}

func NewRepositoryFromModel(repo *models.Repository) (*Repository, error) {
	r := &Repository{
		Path:               repo.Path,
		GitURL:             repo.GitURL,
		Name:               repo.Name,
		Description:        repo.Description,
		Branch:             repo.Branch,
		Readme:             repo.Readme,
		Overview:           repo.Overview,
		StructedCatalogue:  repo.StructedCatalogue,
		StructedCodePath:   repo.StructedCodePath,
		StructedVectorPath: repo.StructedVectorPath,
		Language:           repo.Language,
	}
	gitRepo, err := git.PlainOpen(repo.Path)
	if err != nil {
		return nil, fmt.Errorf("open repository failed: %w", err)
	}
	r.repo = gitRepo
	r.fileScanner = NewFileScanner(&AnalyzeOptions{
		EnableSmartFilter: true,
		ExcludedFiles:     DefaultExcludedFiles,
		MaxFileSize:       1024 * 1024, // 1MB
		MaxTokens:         8192,
		Language:          repo.Language,
	})

	// get the repository catalog
	catalogs, err := r.fileScanner.GetCatalogue(r.Path)
	if err != nil {
		zap.L().Error("get repository catalog failed", zap.Error(err))
		return nil, err
	}
	r.catalogs = catalogs

	llmConfig := config.GetLLMConfig()
	// create the llm provider
	provider, err := chat.NewProvider(&chat.ProviderConfig{
		Type:        chat.ProviderType(llmConfig.ProviderType),
		APIKey:      llmConfig.APIKey,
		Model:       llmConfig.Model,
		MaxTokens:   llmConfig.MaxTokens,
		Temperature: llmConfig.Temperature,
		BaseURL:     llmConfig.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create llm provider failed: %w", err)
	}
	r.provider = provider

	return r, nil
}

func (r *Repository) getStructedCodePath(raw string) string {
	if len(raw) == 0 {
		return ""
	}
	repoConfig := config.GetRepositoryConfig()
	return path.Join(repoConfig.Code, raw)
}

func (r *Repository) getStructedVectorPath(raw string) string {
	if len(raw) == 0 {
		return ""
	}
	repoConfig := config.GetRepositoryConfig()
	return path.Join(repoConfig.Vector, raw)
}

func (r *Repository) IndexCode() (bool, error) {
	var err error

	// indexing the repository
	r.codeIndexer, err = codemap.NewCodeMapService(r.Path)
	if err != nil {
		zap.L().Error("create code map service failed", zap.Error(err))
		return false, err
	}

	err = r.codeIndexer.LoadFromFile(r.getStructedCodePath(r.StructedCodePath), r.getStructedVectorPath(r.StructedVectorPath))
	if err != nil {
		zap.L().Warn("load code map service failed, start indexing ", zap.Error(err))
	}

	var needSave bool
	needSave, err = r.codeIndexer.IndexRepository(r.Path, r.GitURL)
	if err != nil {
		zap.L().Error("index repository failed", zap.Error(err))
		return false, err
	}
	// save the code map service to file
	if needSave {

		tmp := uuid.New().String()
		r.StructedCodePath = tmp + ".code"
		r.StructedVectorPath = tmp + ".vector"

		err = r.codeIndexer.SaveToFile(r.getStructedCodePath(r.StructedCodePath), r.getStructedVectorPath(r.StructedVectorPath))
		if err != nil {
			zap.L().Error("save code map service failed", zap.Error(err))
			return false, err
		}

	}

	return true, nil
}

func (r *Repository) CreateDocuments() (*WikiDocument, error) {

	documentResults, err := r.generateThinkCatalogue(r.provider)
	if err != nil {
		zap.L().Warn("generate documents failed", zap.Error(err))
		return nil, err
	}

	var doc = &WikiDocument{}
	documentResults.Generate(r.provider, r, doc)

	return doc, nil
}

// ParseReadme read the README file content
func (r *Repository) ParseReadme() (string, error) {
	// 尝试不同的 README 文件名
	readmeFiles := []string{
		filepath.Join(r.Path, "README.md"),
		filepath.Join(r.Path, "README.txt"),
		filepath.Join(r.Path, "README"),
		filepath.Join(r.Path, "Readme.md"),
		filepath.Join(r.Path, "readme.md"),
	}

	for _, file := range readmeFiles {
		if _, err := os.Stat(file); err == nil {
			content, err := ReadFile(file)
			if err == nil {
				r.Readme = content
				return content, nil
			}
		}
	}

	return "", nil
}

// ParseDescription parse the repository description from .git/description or README data.
func (r *Repository) ParseDescription() (string, error) {
	// try to read the .git/description file
	descPath := filepath.Join(r.Path, ".git", "description")
	if _, err := os.Stat(descPath); err == nil {
		data, err := os.ReadFile(descPath)
		if err == nil {
			desc := strings.TrimSpace(string(data))
			// 忽略默认描述
			if desc != "Unnamed repository; edit this file 'description' to name the repository." {
				return desc, nil
			}
		}
	}

	// if not found, try to read the README file
	readmePaths := []string{
		filepath.Join(r.Path, "README.md"),
		filepath.Join(r.Path, "README.txt"),
		filepath.Join(r.Path, "README"),
		filepath.Join(r.Path, "Readme.md"),
		filepath.Join(r.Path, "readme.md"),
	}

	for _, path := range readmePaths {
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err == nil {
				lines := strings.Split(string(data), "\n")
				if len(lines) > 0 {
					// use the first line as description
					r.Description = strings.TrimSpace(lines[0])
					// if the description starts with "# ", remove it
					r.Description = strings.TrimLeft(r.Description, "# ")
					return r.Description, nil
				}
			}
		}
	}

	return "", fmt.Errorf("cannot find repository description")
}
