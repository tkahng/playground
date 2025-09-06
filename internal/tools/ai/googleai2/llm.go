package googleai2

// nolint:exhaustruct

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tkahng/playground/internal/conf"
	"google.golang.org/genai"
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
	config = &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		SystemInstruction: genai.NewContentFromText(`# System Prompt For ChatGPT

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
	`, genai.RoleModel),
		ResponseSchema: &genai.Schema{
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
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  config.GoogleGeminiApiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &AiService{client: client}
}

type TaskProjectGenerator interface {
	GenerateProjectPlan(ctx context.Context, projectInput string) (*AiTaskResponse, error)
}

func (c *AiService) GenerateProjectPlan(ctx context.Context, projectInput string) (*AiTaskResponse, error) {
	resp, err := c.client.Models.GenerateContent(ctx,
		"gemini-2.5-flash", // or a different supported model
		genai.Text(projectInput),
		config,
	)
	if err != nil {
		return nil, err
	}

	var taskResp TaskResponse
	if err := json.Unmarshal([]byte(resp.Text()), &taskResp); err != nil {
		return nil, err
	}

	back := AiUsage{
		PromptTokens: int64(resp.UsageMetadata.TotalTokenCount), // adapt based on response API
		TotalTokens:  int64(resp.UsageMetadata.TotalTokenCount),
	}

	return &AiTaskResponse{
		TaskResponse: taskResp,
		Usage:        back,
	}, nil
}
