// FILE: platform/aiservice/interface.go
package aiservice

import "context"

// AIService defines the interface for AI providers
type AIService interface {
	GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// TextGenerationOptions contains common options for text generation
type TextGenerationOptions struct {
	Temperature float64
	MaxTokens   int
	Model       string
}
