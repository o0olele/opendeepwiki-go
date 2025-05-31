package config

import (
	"embed"

	"go.uber.org/zap"
)

//go:embed templates/prompts/*.txt
var templateFS embed.FS

var OverviewPrompt string
var GenerateDocsPrompt string
var HistoryPrompt string
var GenerateCatalogPrompt string
var FirstPrompt string
var DeepFirstPrompt string
var ChatPrompt string
var AnalyzeCatalogPrompt string
var AnalyzeNewCatalogPrompt string
var SimplifyDirsPrompt string
var GenerateReadmePrompt string

func LoadTemplates() {

	OverviewPrompt = readFile("templates/prompts/overviews.txt")
	GenerateDocsPrompt = readFile("templates/prompts/generate_docs.txt")
	HistoryPrompt = readFile("templates/prompts/history.txt")
	GenerateCatalogPrompt = readFile("templates/prompts/generate_catalog.txt")
	FirstPrompt = readFile("templates/prompts/first.txt")
	DeepFirstPrompt = readFile("templates/prompts/deepfirst.txt")
	ChatPrompt = readFile("templates/prompts/chat.txt")
	AnalyzeCatalogPrompt = readFile("templates/prompts/analyze_catalog.txt")
	AnalyzeNewCatalogPrompt = readFile("templates/prompts/analyze_newcatalog.txt")
	SimplifyDirsPrompt = readFile("templates/prompts/simplify_dirs.txt")
	GenerateReadmePrompt = readFile("templates/prompts/generate_readme.txt")

	zap.L().Info("Loaded templates", zap.String("overview_prompt", OverviewPrompt))
}

func readFile(path string) string {

	content, err := templateFS.ReadFile(path)
	if err != nil {
		zap.L().Error("Failed to read file: %v", zap.Error(err))
		return ""
	}
	return string(content)
}
