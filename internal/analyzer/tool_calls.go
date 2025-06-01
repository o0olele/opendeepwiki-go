package analyzer

import (
	"encoding/json"
	"os"
	"path"

	"github.com/tmc/langchaingo/jsonschema"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"
)

var llmTools []llms.Tool

func init() {

	llmTools = append(llmTools, llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "readFiles",
			Description: "Read the specified file content",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"filePaths": {
						Type: jsonschema.Array,
						Items: &jsonschema.Definition{
							Type:        jsonschema.String,
							Description: "File Path",
						},
						Description: "The file paths to read",
					},
				},
				Required: []string{"filePaths"},
			},
		},
	})

	llmTools = append(llmTools, llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "seachCode",
			Description: "help you search the code in the repository",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"query": {
						Type:        jsonschema.String,
						Description: "The query to search for, usually a function name or a class name",
					},
					"minRelevance": {
						Type:        jsonschema.Number,
						Description: "The minimum relevance score for the search results",
					},
				},
				Required: []string{"query"},
			},
		},
	})
}

func (r *Repository) readFiles(paths []string) string {
	var (
		total int
		dic   = make(map[string]string)
	)

	for _, p := range paths {
		item := path.Join(r.Path, p)
		if info, err := os.Stat(item); err == nil {
			total += int(info.Size())
			if total > 81920 {
				return "target files is too large."
			}

			content, err := ReadFile(item)
			if err == nil {
				dic[p] = content
			}
		}
	}
	s, _ := json.Marshal(dic)
	return string(s)
}

func (r *Repository) searchCode(query string, minRelevance float64) string {

	results, err := r.codeIndexer.SearchCode(query, r.Description, 3)
	if err != nil {
		zap.L().Error("search code failed", zap.Error(err))
		return "search code failed"
	}

	s, _ := json.Marshal(results)
	return string(s)
}

func (r *Repository) updateMessageHistory(messageHistory []llms.MessageContent, choice *llms.ContentChoice) []llms.MessageContent {
	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, choice.Content)
	for _, tc := range choice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	return append(messageHistory, assistantResponse)
}

func (r *Repository) executeToolCalls(llm llms.Model, messageHistory []llms.MessageContent, choice *llms.ContentChoice) []llms.MessageContent {

	for idx, toolCall := range choice.ToolCalls {
		if len(toolCall.ID) <= 0 {
			continue
		}
		for idy := idx + 1; idy < len(choice.ToolCalls); idy++ {
			tmpCall := choice.ToolCalls[idy]
			if len(tmpCall.ID) > 0 {
				break
			}
			toolCall.ID += tmpCall.ID
			if toolCall.FunctionCall != nil {
				toolCall.FunctionCall.Name += tmpCall.FunctionCall.Name
				toolCall.FunctionCall.Arguments += tmpCall.FunctionCall.Arguments
			}
		}
	}

	for _, toolCall := range choice.ToolCalls {
		switch toolCall.FunctionCall.Name {
		case "readFiles":
			var args struct {
				FilePaths []string `json:"filePaths"`
			}
			if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
				zap.L().Warn("tool call arguments unmarshal failed: %v", zap.Error(err))
				continue
			}

			response := llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: toolCall.ID,
						Name:       toolCall.FunctionCall.Name,
						Content:    r.readFiles(args.FilePaths),
					},
				},
			}
			messageHistory = append(messageHistory, response)
		case "seachCode":
			var args struct {
				Query        string  `json:"query"`
				MinRelevance float64 `json:"minRelevance"`
			}
			if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
				zap.L().Warn("tool call arguments unmarshal failed: %v", zap.Error(err))
				continue
			}

			response := llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: toolCall.ID,
						Name:       toolCall.FunctionCall.Name,
						Content:    r.searchCode(args.Query, args.MinRelevance),
					},
				},
			}
			messageHistory = append(messageHistory, response)
		}
	}
	return messageHistory
}
