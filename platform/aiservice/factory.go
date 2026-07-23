// FILE: platform/aiservice/factory.go
package aiservice

import (
	"context"
	"fmt"
)

// NewClient dispatches to the constructor for config["provider"], so callers
// that read a provider from agent config (orchestration steps, standalone
// agents such as content-creator) share one place that knows which providers
// exist rather than re-implementing the switch at each call site.
func NewClient(ctx context.Context, config map[string]interface{}) (AIService, error) {
	rawProvider, exists := config["provider"]
	if !exists || rawProvider == nil {
		return nil, fmt.Errorf("provider not specified in ai_service config - api_key_env_var included in config.ai_service?")
	}

	provider, ok := rawProvider.(string)
	if !ok || provider == "" {
		return nil, fmt.Errorf("provider must be a non-empty string in ai_service config")
	}

	switch provider {
	case "anthropic":
		return NewAnthropicClient(ctx, config)
	case "ollama":
		return NewOllamaClient(ctx, config)
	case "gemini":
		return NewGeminiClient(ctx, config)
	case "openai":
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s. Supported: anthropic, ollama, gemini", provider)
	}
}
