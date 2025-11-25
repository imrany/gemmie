package genai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	geminiwrapper "github.com/imrany/wrapper/pkg/gemini"
	openaiwrapper "github.com/imrany/wrapper/pkg/openai"
)

type GENAISERVICE struct {
	Model  string
	APIKey string
}

type GenAiResponse struct {
	Prompt   string
	Response string
}

func (g *GENAISERVICE) GenerateAIResponse(ctx context.Context, prompt string) (*GenAiResponse, error) {
	if prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	// Extract provider from model name (case-insensitive)
	modelParts := strings.Split(g.Model, "-")
	if len(modelParts) == 0 {
		slog.Error("Invalid model format", "model", g.Model)
		return nil, fmt.Errorf("invalid model format: %s", g.Model)
	}
	provider := strings.ToLower(modelParts[0])

	switch provider {
	case "gemini":
		config := geminiwrapper.GeminiClientConfig{
			APIKey: g.APIKey,
			Model:  g.Model,
		}
		result, err := config.GenerateGeminiContent(ctx, prompt)
		if err != nil {
			slog.Error("Gemini generation failed", "error", err)
			// Check if it's a context error
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			// Return the actual error to the client for debugging
			return nil, err
		}
		return &GenAiResponse{
			Prompt:   prompt,
			Response: result,
		}, nil

	case "gpt", "o1":
		config := openaiwrapper.OpenAIClient{
			APIKey: g.APIKey,
			Model:  g.Model,
		}
		result, err := config.GenerateOpenAIContent(ctx, prompt)
		if err != nil {
			slog.Error("OpenAI generation failed", "error", err)
			// Check if it's a context error
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			// Return the actual error to the client for debugging
			return nil, err
		}
		return &GenAiResponse{
			Prompt:   prompt,
			Response: result,
		}, nil

	default:
		slog.Warn("Unsupported model", "model", g.Model)
		return nil, fmt.Errorf("unsupported model: %s", g.Model)
	}
}
