package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/o0olele/opendeepwiki-go/internal/config"
	"github.com/o0olele/opendeepwiki-go/internal/llm/chat"
	"github.com/o0olele/opendeepwiki-go/internal/utils"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"go.uber.org/zap"
)

func (r *Repository) GenerateReadme() (string, error) {
	var prompt = prompts.PromptTemplate{
		Template: config.GenerateReadmePrompt,
		PartialVariables: map[string]any{
			"catalogue":      CatalogueToString(r.Path, r.catalogs),
			"branch":         r.Branch,
			"git_repository": r.GitURL,
			"language":       r.Language,
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

	response, err := r.provider.GetModel().GenerateContent(context.Background(), message,
		llms.WithMaxTokens(8192),
	)
	if err != nil {
		zap.L().Error("cannot get model response", zap.Error(err))
		return "", err
	}

	r.Readme = utils.ExtractTagContent(response.Choices[0].Content, "readme")
	return r.Readme, nil
}

func (r *Repository) GenerateStructedCatalogue() (string, error) {
	return r.fileScanner.GetSimplifyCatalogueString(r.provider, r.Path, r.catalogs, r.Readme)
}

func (r *Repository) GenerateOverview() (string, error) {
	zap.L().Info("generating repository overview", zap.String("repository", r.StructedCatalogue))
	var prompt = prompts.PromptTemplate{
		Template: config.OverviewPrompt,
		PartialVariables: map[string]any{
			"catalogue":      r.StructedCatalogue,
			"branch":         r.Branch,
			"git_repository": r.GitURL,
			"readme":         r.Readme,
			"language":       r.Language,
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

	response, err := r.provider.GetModel().GenerateContent(context.Background(), message,
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

func (r *Repository) generateCatalogue(provider chat.Provider, think string) (*DocumentResultCalalogue, error) {
	var prompt = prompts.PromptTemplate{
		Template: config.AnalyzeCatalogPrompt,
		PartialVariables: map[string]any{
			"think":           think,
			"code_files":      r.StructedCatalogue,
			"repository_name": r.Name,
			"language":        r.Language,
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
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				fmt.Print(string(chunk))
				return nil
			}),
		)
		if err != nil {
			zap.L().Warn("cannot get model response", zap.Error(err))
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

func (r *Repository) generateThinkCatalogue(provider chat.Provider) (*DocumentResultCalalogue, error) {

	var prompt = prompts.PromptTemplate{
		Template: config.GenerateCatalogPrompt,
		PartialVariables: map[string]any{
			"code_files":         r.StructedCatalogue,
			"git_repository_url": r.GitURL,
			"repository_name":    r.Name,
			"language":           r.Language,
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
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		zap.L().Error("cannot get model response", zap.Error(err))
		return nil, err
	}
	choice := response.Choices[0]

	return r.generateCatalogue(provider, choice.Content)
}

func (r *Repository) generateCatalogueItem(provider chat.Provider, catalogItem *DocumentResultCalalogueItem) (*WikiDocument, error) {
	var prompt = prompts.PromptTemplate{
		Template: config.GenerateDocsPrompt,
		PartialVariables: map[string]any{
			"prompt":         catalogItem.Prompt,
			"title":          catalogItem.Title,
			"git_repository": r.GitURL,
			"branch":         r.Branch,
			"catalogue":      r.StructedCatalogue,
			"language":       r.Language,
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
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				fmt.Print(string(chunk))
				return nil
			}),
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
		Content: utils.ExtractTagContent(str.String(), "docs"),
		Title:   catalogItem.Title,
	}

	return result, nil

}
