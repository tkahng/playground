package googleai

import (
	"context"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AiClient struct {
	client *genai.Client
}

func (c *AiClient) Client() *genai.Client {
	return c.client
}
func NewAiClient(ctx context.Context, apiKey string) *genai.Client {
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	return client
}
