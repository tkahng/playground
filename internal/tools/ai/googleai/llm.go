package googleai

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/generative-ai-go/genai"
	"github.com/tkahng/authgo/internal/conf"
	"google.golang.org/api/option"
)

var (
	Schema = &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"project": {
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"name": {
						Type: genai.TypeString,
					},
					"description": {
						Type: genai.TypeString,
					},
				},
				Required: []string{"name"},
			},
			"tasks": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"name": {
							Type: genai.TypeString,
						},
						"description": {
							Type: genai.TypeString,
						},
					},
					Required: []string{"name"},
				},
			},
		},
		Required: []string{"tasks", "project"},
	}
	SystemPrompt = &genai.Content{
		Parts: []genai.Part{
			genai.Text(`# System Prompt For ChatGPT

## Objective and Scope

**Objective**

Assistant is a Project Planning Expert. Your role is to assist users by mainly providing following aspects step-by-step.

## Guidelines and Instructions

**General Instructions**:

- Gather information about the project scope and the steps involved.
- Break down the project into manageable tasks, listing them in a logical order.
- Include preparation steps, construction steps, and finishing touches.
- Provide estimated timeframes for each task if possible.
- For each task, provide a name, step number starting from zero, task details, and estimated time to complete.
- Provide a project name and description based on the user's request.
- Respond according to structured output.
	`),
		},
	}
)

type Task struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type TaskResponse struct {
	Project Project `json:"project"`
	Tasks   []Task  `json:"tasks"`
}

type AiTaskResponse struct {
	TaskResponse
	Usage AiUsage `json:"usage"`
}

type AiUsage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type AiService struct {
	client *genai.Client
}

func (c *AiService) Client() *genai.Client {
	return c.client
}
func NewAiService(ctx context.Context, config conf.AiConfig) *AiService {
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.GoogleGeminiApiKey))
	if err != nil {
		log.Fatal(err)
	}
	return &AiService{client: client}
}

type TaskProjectGenerator interface {
	GenerateProjectPlan(ctx context.Context, projectInput string) (*AiTaskResponse, error)
}

func (c *AiService) GenerateProjectPlan(ctx context.Context, projectInput string) (*AiTaskResponse, error) {
	model := c.client.GenerativeModel("gemini-1.5-pro-latest")
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = Schema
	model.SystemInstruction = SystemPrompt
	resp, err := model.GenerateContent(ctx, genai.Text(projectInput))
	if err != nil {
		return nil, err
	}
	var taskResponse TaskResponse
	if err := json.Unmarshal([]byte(resp.Candidates[0].Content.Parts[0].(genai.Text)), &taskResponse); err != nil {
		return nil, err
	}
	usage := AiUsage{
		PromptTokens:     int64(resp.UsageMetadata.PromptTokenCount),
		CompletionTokens: int64(resp.UsageMetadata.CandidatesTokenCount),
		TotalTokens:      int64(resp.UsageMetadata.TotalTokenCount),
	}
	return &AiTaskResponse{
		TaskResponse: taskResponse,
		Usage:        usage,
	}, nil
}
