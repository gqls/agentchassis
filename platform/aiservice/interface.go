// FILE: platform/aiservice/interface.go
package aiservice

import "context"

// AIService defines the interface for AI providers
type AIService interface {
	GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	Provider() string
	Model() string
}

// ImageInput is one image handed to a vision-capable provider. Data is raw
// bytes (the provider encodes as its wire format needs); MediaType is the MIME
// type, e.g. "image/png".
type ImageInput struct {
	MediaType string
	Data      []byte
}

// VisionCapable is the OPTIONAL capability interface for providers that accept
// images alongside a prompt. Deliberately separate from AIService: widening the
// core interface would break every existing implementer and test fake for a
// capability most call sites never use. Callers type-assert:
//
//	v, ok := svc.(VisionCapable)
//
// and treat !ok as "this provider has no eyes" — a configuration error to
// surface, never to paper over. anthropic and gemini implement it (the
// design-critic model trial is a config switch between them); ollama does not.
type VisionCapable interface {
	GenerateWithImages(ctx context.Context, prompt string, images []ImageInput, options map[string]interface{}) (string, error)
}

// TextGenerationOptions contains common options for text generation
type TextGenerationOptions struct {
	Temperature float64
	MaxTokens   int
	Model       string
}
