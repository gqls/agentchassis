// FILE: platform/aiservice/anthropic.go
package aiservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// AnthropicClient implements the AIService interface for Anthropic's Claude
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(ctx context.Context, config map[string]interface{}) (*AnthropicClient, error) {
	// Validate config exists
	if config == nil {
		return nil, fmt.Errorf("ai_service config is nil - ensure 'ai_service' block exists in step config")
	}

	// Safely extract api_key_env_var with helpful error message
	apiKeyEnvVarRaw, exists := config["api_key_env_var"]
	if !exists || apiKeyEnvVarRaw == nil {
		return nil, fmt.Errorf("ai_service.api_key_env_var not configured - add 'api_key_env_var: \"ANTHROPIC_API_KEY\"' to ai_service config")
	}
	apiKeyEnvVar, ok := apiKeyEnvVarRaw.(string)
	if !ok || apiKeyEnvVar == "" {
		return nil, fmt.Errorf("ai_service.api_key_env_var must be a non-empty string, got: %T", apiKeyEnvVarRaw)
	}

	// Get API key from environment
	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable '%s' is not set or empty", apiKeyEnvVar)
	}

	// Safely extract model with default fallback
	model := "claude-sonnet-4-6" // sensible default
	// model := "claude-haiku-4-5-20251001"
	if modelRaw, exists := config["model"]; exists && modelRaw != nil {
		if modelStr, ok := modelRaw.(string); ok && modelStr != "" {
			model = modelStr
		}
	}

	// Resolve alias to actual model name
	model = ResolveModelAlias(model, nil) // or pass a logger if available

	/*if model == "" {
		return nil, fmt.Errorf("no AI model selected")
	}*/
	return &AnthropicClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

// GenerateText generates text using Claude
func (c *AnthropicClient) GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	// Build request
	requestBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 2048,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// Check if extended thinking is requested
	if options != nil {
		if budgetTokens, ok := options["budget_tokens"]; ok {
			switch bt := budgetTokens.(type) {
			case float64:
				if bt > 0 {
					requestBody["thinking"] = map[string]interface{}{
						"type":          "enabled",
						"budget_tokens": int(bt),
					}
				}
			case int:
				if bt > 0 {
					requestBody["thinking"] = map[string]interface{}{
						"type":          "enabled",
						"budget_tokens": bt,
					}
				}
			}
		}
	}

	// Temperature is intentionally NOT sent to Anthropic.
	// Claude Opus 4.7+ returns a 400 for any non-default temperature, and
	// extended thinking is incompatible with temperature on any model.
	// Temperature stays in the generic options contract for other providers
	// (e.g. ollama); the Anthropic client simply ignores it.
	if options != nil {
		// Override max_tokens from provided options
		if maxTokens, ok := options["max_tokens"]; ok {
			requestBody["max_tokens"] = maxTokens
		}
		// Record what was actually sent, for llm_call_log (mirrors the
		// __usage_* write-back below). No temperature is sent, so
		// __sent_temperature is deliberately left unset.
		switch mt := requestBody["max_tokens"].(type) {
		case int:
			options["__sent_max_tokens"] = mt
		case float64:
			options["__sent_max_tokens"] = int(mt)
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response — handle both standard and extended thinking formats
	// With thinking enabled, content array has {type:"thinking"} blocks
	// followed by {type:"text"} blocks. We want the text block.
	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Write usage tokens back to options so the caller can log them
	if options != nil {
		options["__usage_input_tokens"] = response.Usage.InputTokens
		options["__usage_output_tokens"] = response.Usage.OutputTokens
	}

	// Extract the text BEFORE deciding whether this was a truncation, so a cut
	// completion can carry its partial back to the caller. Returning "" here is
	// what made bugs_open/019 unrecoverable: the platform detected the cut and
	// then destroyed the only thing that could survive it.
	// Block TYPES, not just a count. A single non-text block is the tell that
	// separates a refusal from a thinking-only response, and the old terminal
	// error reported only "had 1 blocks" — a number that distinguishes neither
	// (bugs_open/008 §REAL-CASE).
	blockTypes := make([]string, 0, len(response.Content))
	for _, block := range response.Content {
		if block.Type == "" {
			blockTypes = append(blockTypes, "(untyped)")
			continue
		}
		blockTypes = append(blockTypes, block.Type)
	}

	partial := ""
	for _, block := range response.Content {
		if block.Type == "text" || block.Type == "" {
			partial = block.Text
			break
		}
		if partial == "" && block.Text != "" {
			partial = block.Text
		}
	}

	if response.StopReason == "max_tokens" {
		// Non-nil error, so every existing `if err != nil` caller behaves exactly
		// as before. Only a caller that asks via aiservice.IsTruncated sees the
		// partial — tolerating a cut stays an opt-in decision at the step.
		return partial, &TruncatedError{
			Partial:      partial,
			OutputTokens: response.Usage.OutputTokens,
			Reason:       "stop_reason=max_tokens",
			Provider:     "anthropic",
		}
	}

	// A refusal is a DECISION, not a failure to produce: the model understood the
	// request and declined it. Decoded here so it cannot fall through to the
	// terminal fallback below, where it read as a parse fault and sent the last
	// reader looking in the wrong layer entirely (bugs_open/008 item 5).
	if response.StopReason == "refusal" {
		return "", &RefusalError{
			Reason:   "stop_reason=refusal",
			Blocks:   blockTypes,
			Provider: "anthropic",
			Model:    c.model,
		}
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response (stop_reason=%q)", response.StopReason)
	}

	if partial != "" {
		return partial, nil
	}

	// Fallback: return first block with any text
	for _, block := range response.Content {
		if block.Text != "" {
			return block.Text, nil
		}
	}

	// Name the cause. "had 1 blocks" was a count that distinguished a refusal, a
	// thinking-only response and a genuine parse fault from each other not at all.
	return "", fmt.Errorf("no text content in response (stop_reason=%q, %d block(s): %s)",
		response.StopReason, len(response.Content), strings.Join(blockTypes, ", "))
}

// GenerateEmbedding generates embeddings (not implemented for Anthropic)
func (c *AnthropicClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("embedding generation not supported by Anthropic")
}

func (c *AnthropicClient) Model() string {
	return c.model
}

func (c *AnthropicClient) Provider() string {
	return "anthropic"
}
