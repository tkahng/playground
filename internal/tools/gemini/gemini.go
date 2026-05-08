package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tkahng/playground/internal/conf"
)

const endpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"

const systemPrompt = `# System Prompt For ChatGPT

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
	`

type Task struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProjectPlan struct {
	Project Project `json:"project"`
	Tasks   []Task  `json:"tasks"`
}

type Usage struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}

type ProjectPlanResult struct {
	ProjectPlan
	Usage Usage `json:"usage"`
}

// request/response types for the Gemini REST API

type part struct {
	Text string `json:"text"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type schema struct {
	Type       string             `json:"type"`
	Properties map[string]*schema `json:"properties,omitempty"`
	Items      *schema            `json:"items,omitempty"`
	Required   []string           `json:"required,omitempty"`
}

type generationConfig struct {
	ResponseMIMEType string  `json:"responseMimeType"`
	ResponseSchema   *schema `json:"responseSchema"`
}

type generateRequest struct {
	SystemInstruction *content         `json:"systemInstruction,omitempty"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type candidate struct {
	Content content `json:"content"`
}

type usageMetadata struct {
	PromptTokenCount     int64 `json:"promptTokenCount"`
	CandidatesTokenCount int64 `json:"candidatesTokenCount"`
	TotalTokenCount      int64 `json:"totalTokenCount"`
}

type generateResponse struct {
	Candidates    []candidate   `json:"candidates"`
	UsageMetadata usageMetadata `json:"usageMetadata"`
}

var projectPlanSchema = &schema{
	Type: "object",
	Properties: map[string]*schema{
		"project": {
			Type: "object",
			Properties: map[string]*schema{
				"name":        {Type: "string"},
				"description": {Type: "string"},
			},
			Required: []string{"name"},
		},
		"tasks": {
			Type: "array",
			Items: &schema{
				Type: "object",
				Properties: map[string]*schema{
					"name":        {Type: "string"},
					"description": {Type: "string"},
				},
				Required: []string{"name"},
			},
		},
	},
	Required: []string{"tasks", "project"},
}

type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient(cfg conf.AiConfig) *Client {
	return &Client{
		apiKey: cfg.GoogleGeminiApiKey,
		http:   &http.Client{},
	}
}

func (c *Client) GenerateProjectPlan(ctx context.Context, input string) (*ProjectPlanResult, error) {
	req := generateRequest{
		SystemInstruction: &content{
			Role:  "model",
			Parts: []part{{Text: systemPrompt}},
		},
		Contents: []content{
			{Parts: []part{{Text: input}}},
		},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   projectPlanSchema,
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?key="+c.apiKey, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("gemini API error %d: %v", resp.StatusCode, errBody)
	}

	var gemResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gemResp); err != nil {
		return nil, err
	}

	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned no content")
	}

	var plan ProjectPlan
	if err := json.Unmarshal([]byte(gemResp.Candidates[0].Content.Parts[0].Text), &plan); err != nil {
		return nil, err
	}

	return &ProjectPlanResult{
		ProjectPlan: plan,
		Usage: Usage{
			PromptTokens:     gemResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: gemResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gemResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}
