package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/o0olele/opendeepwiki-go/internal/analyzer/codemap"
	"github.com/o0olele/opendeepwiki-go/internal/config"
	"github.com/o0olele/opendeepwiki-go/internal/llm/chat"
	"github.com/o0olele/opendeepwiki-go/internal/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"go.uber.org/zap"
)

// Repository represents a local code repository.
type Repository struct {
	Path        string // path to the local repository
	GitURL      string // git URL of the repository
	Name        string // name of the repository
	Description string // description of the repository
	Branch      string // branch of the repository
	Readme      string // README content of the repository
	repo        *git.Repository
	codeIndexer *codemap.CodeMapService
}

// NewRepository creates a new Repository instance.
func NewRepository(repoDir string, gitURL string) (*Repository, error) {
	// extract the repository name from the git URL
	name, err := utils.ExtractRepoName(gitURL)
	if err != nil {
		return nil, fmt.Errorf("extract repository name failed: %w", err)
	}
	repoPath := path.Join(repoDir, name)
	// check if the path exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository path does not exist: %s", repoPath)
	}

	// open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repository failed: %w", err)
	}

	zap.L().Info("repository name", zap.String("name", name))

	return &Repository{
		Path:   repoPath,
		GitURL: gitURL,
		Name:   name,
		repo:   repo,
	}, nil
}

// Parse parses the repository and extracts information.
func (r *Repository) Parse(options *AnalyzeOptions) error {
	zap.L().Info("parsing repository", zap.String("path", r.Path))

	var scanner = NewFileScanner(options)
	// get the repository catalog
	catalogs, err := scanner.GetCatalogue(r.Path)
	if err != nil {
		zap.L().Error("get repository catalog failed", zap.Error(err))
		return err
	}

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

	// get the repository readme content
	r.Readme, err = r.parseReadme()
	if err != nil || len(r.Readme) == 0 {
		// generate README content if not found or failed to parse
		r.Readme, err = r.generateReadme(provider, r.Path, catalogs)
	}

	// get the repository catalog string
	catalog, err := scanner.GetSimplifyCatalogueString(provider, r.Path, catalogs, r.Readme)
	if err != nil {
		zap.L().Error("get repository catalog failed", zap.Error(err))
		return err
	}

	// generate the repository overview
	overview, err := r.generateOverview(provider, catalog)
	if err != nil {
		zap.L().Error("generate repository overview failed", zap.Error(err))
		return err
	}

	zap.L().Info("repository overview", zap.String("overview", overview))

	var embeddingConfig = config.GetEmbeddingConfig()
	// indexing the repository
	r.codeIndexer, err = codemap.NewCodeMapService(embeddingConfig.APIKey, embeddingConfig.Model, embeddingConfig.BaseURL, 3, r.Path)
	if err != nil {
		zap.L().Error("create code map service failed", zap.Error(err))
		return err
	}
	r.codeIndexer.IndexRepository(r.Path, r.GitURL)

	documentResults, err := r.generateThinkCatalogue(provider, catalog)
	if err != nil {
		zap.L().Error("generate documents failed", zap.Error(err))
		return err
	}

	var m = make(map[string]*WikiDocument)
	for _, doc := range documentResults.Items {
		wiki, err := r.generateCatalogueItem(provider, &doc, catalog)
		if err != nil {
			continue
		}
		m[wiki.ID] = wiki
	}

	s, _ := json.Marshal(m)
	os.WriteFile("./wiki_"+r.Name, s, os.ModePerm)

	return nil
}

// parseReadme read the README file content
func (r *Repository) parseReadme() (string, error) {
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
				return content, nil
			}
		}
	}

	return "", nil
}

// parseDescription parse the repository description from .git/description or README data.
func (r *Repository) parseDescription() error {
	// try to read the .git/description file
	descPath := filepath.Join(r.Path, ".git", "description")
	if _, err := os.Stat(descPath); err == nil {
		data, err := os.ReadFile(descPath)
		if err == nil {
			desc := strings.TrimSpace(string(data))
			// 忽略默认描述
			if desc != "Unnamed repository; edit this file 'description' to name the repository." {
				r.Description = desc
				return nil
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
					return nil
				}
			}
		}
	}

	return fmt.Errorf("cannot find repository description")
}

// getFileStats 获取仓库中的文件统计信息
func (r *Repository) getFileStats() (map[string]int, error) {
	fileStats := make(map[string]int)

	// 获取最新的提交
	ref, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("获取HEAD引用失败: %w", err)
	}

	commit, err := r.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("获取提交对象失败: %w", err)
	}

	// 获取文件树
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("获取文件树失败: %w", err)
	}

	// 遍历文件树
	err = tree.Files().ForEach(func(f *object.File) error {
		// 获取文件扩展名
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == "" {
			ext = "(no extension)"
		}
		fileStats[ext]++
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历文件失败: %w", err)
	}

	return fileStats, nil
}

// printFileStats 打印文件统计信息
func (r *Repository) printFileStats(fileStats map[string]int) {
	log.Printf("仓库 %s 的文件类型统计:", r.Name)
	for ext, count := range fileStats {
		log.Printf("  %s: %d 个文件", ext, count)
	}
}

func (r *Repository) generateReadme(provider chat.Provider, repoPath string, catalogs []PathInfo) (string, error) {
	var prompt = prompts.PromptTemplate{
		Template: config.GenerateReadmePrompt,
		PartialVariables: map[string]any{
			"catalogue":      CatalogueToString(repoPath, catalogs),
			"branch":         r.Branch,
			"git_repository": r.GitURL,
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
		llms.WithMaxTokens(8192),
	)
	if err != nil {
		zap.L().Error("cannot get model response", zap.Error(err))
		return "", err
	}

	return utils.ExtractTagContent(response.Choices[0].Content, "readme"), nil
}

func (r *Repository) generateOverview(provider chat.Provider, catalogs string) (string, error) {
	zap.L().Info("generating repository overview", zap.String("repository", catalogs))
	var prompt = prompts.PromptTemplate{
		Template: config.OverviewPrompt,
		PartialVariables: map[string]any{
			"catalogue":      catalogs,
			"branch":         r.Branch,
			"git_repository": r.GitURL,
			"readme":         r.Readme,
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
		llms.WithMaxTokens(8192),
		llms.WithTools(llmTools),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		zap.L().Error("cannot get model response", zap.Error(err))
		return "", err
	}
	overview := utils.ExtractTagContent(response.Choices[0].Content, "blog")
	overview = utils.ReplaceTagContent(overview, "project_analysis", "")
	return overview, nil
}

func (r *Repository) generateCatalogue(provider chat.Provider, think string, catalogs string) (*DocumentResultCalalogue, error) {
	var prompt = prompts.PromptTemplate{
		Template: config.AnalyzeCatalogPrompt,
		PartialVariables: map[string]any{
			"think":           think,
			"code_files":      catalogs,
			"repository_name": r.Name,
		},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	content, err := prompt.Format(nil)
	if err != nil {
		zap.L().Error("cannot format prompt", zap.Error(err))
		return nil, err
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, content),
	}

	var str strings.Builder
	for i := 0; i < 16; i++ {
		response, err := provider.GetModel().GenerateContent(context.Background(), messages,
			llms.WithTools(llmTools),
		)
		if err != nil {
			zap.L().Error("cannot get model response", zap.Error(err))
			continue
		}
		choice := response.Choices[0]
		if len(choice.ToolCalls) > 0 {
			messages = r.updateMessageHistory(messages, choice)
			messages = r.executeToolCalls(provider.GetModel(), messages, choice)
		} else {
			str.WriteString(choice.Content)
			break
		}
	}

	extract := utils.ExtractTagContent(str.String(), "documentation_structure")
	extract = utils.ExtractJSON(extract)

	var result = new(DocumentResultCalalogue)
	if err := json.Unmarshal([]byte(extract), result); err != nil {
		zap.L().Error("cannot unmarshal extract", zap.Error(err))
		return nil, err
	}

	return result, nil
}

func (r *Repository) generateThinkCatalogue(provider chat.Provider, catalog string) (*DocumentResultCalalogue, error) {

	var prompt = prompts.PromptTemplate{
		Template: config.GenerateCatalogPrompt,
		PartialVariables: map[string]any{
			"code_files":         catalog,
			"git_repository_url": r.GitURL,
			"repository_name":    r.Name,
		},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	content, err := prompt.Format(nil)
	if err != nil {
		zap.L().Error("cannot format prompt", zap.Error(err))
		return nil, err
	}

	message := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, content),
	}

	response, err := provider.GetModel().GenerateContent(context.Background(), message,
		llms.WithMaxTokens(8192),
	)
	if err != nil {
		zap.L().Error("cannot get model response", zap.Error(err))
		return nil, err
	}
	choice := response.Choices[0]

	return r.generateCatalogue(provider, choice.Content, catalog)
}

func (r *Repository) generateCatalogueItem(provider chat.Provider, catalogItem *DocumentResultCalalogueItem, catalog string) (*WikiDocument, error) {
	var prompt = prompts.PromptTemplate{
		Template: config.GenerateDocsPrompt,
		PartialVariables: map[string]any{
			"prompt":         catalogItem.Prompt,
			"title":          catalogItem.Title,
			"git_repository": r.GitURL,
			"branch":         r.Branch,
			"catalogue":      catalog,
		},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	content, err := prompt.Format(nil)
	if err != nil {
		zap.L().Error("cannot format prompt", zap.Error(err))
		return nil, err
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, content),
	}

	var str strings.Builder
	for i := 0; i < 16; i++ {
		response, err := provider.GetModel().GenerateContent(context.Background(), messages,
			llms.WithTools(llmTools),
		)
		if err != nil {
			zap.L().Error("cannot get model response", zap.Error(err))
			continue
		}
		choice := response.Choices[0]
		if len(choice.ToolCalls) > 0 {
			messages = r.updateMessageHistory(messages, choice)
			messages = r.executeToolCalls(provider.GetModel(), messages, choice)
		} else {
			str.WriteString(choice.Content)
			break
		}
	}

	var result = &WikiDocument{
		ID:      catalogItem.Name,
		Content: utils.ExtractTagContent(str.String(), "docs"),
		Title:   catalogItem.Title,
	}

	return result, nil

}
