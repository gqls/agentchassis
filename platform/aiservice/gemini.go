// FILE: platform/aiservice/gemini.go
//
// Gemini provider (Google's generativelanguage.googleapis.com API).
// Implements the same AIService interface as AnthropicClient/OllamaClient.
//
// Agent definitions use it like:
//   "ai_service": {
//       "provider": "gemini",
//       "model": "gemini-2.5-pro",
//       "api_key_env_var": "GEMINI_API_KEY"
//   }

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

const geminiAPIBase = "https://generativelanguage.googleapis.com/v1beta/models"

// GeminiClient implements the AIService interface for Google's Gemini
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(ctx context.Context, config map[string]interface{}) (*GeminiClient, error) {
	if config == nil {
		return nil, fmt.Errorf("ai_service config is nil - ensure 'ai_service' block exists in step config")
	}

	apiKeyEnvVarRaw, exists := config["api_key_env_var"]
	if !exists || apiKeyEnvVarRaw == nil {
		return nil, fmt.Errorf("ai_service.api_key_env_var not configured - add 'api_key_env_var: \"GEMINI_API_KEY\"' to ai_service config")
	}
	apiKeyEnvVar, ok := apiKeyEnvVarRaw.(string)
	if !ok || apiKeyEnvVar == "" {
		return nil, fmt.Errorf("ai_service.api_key_env_var must be a non-empty string, got: %T", apiKeyEnvVarRaw)
	}

	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable '%s' is not set or empty", apiKeyEnvVar)
	}

	model := "gemini-2.5-pro" // sensible default
	if modelRaw, exists := config["model"]; exists && modelRaw != nil {
		if modelStr, ok := modelRaw.(string); ok && modelStr != "" {
			model = modelStr
		}
	}

	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}, nil
}

// GenerateText generates text using Gemini
func (c *GeminiClient) GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	generationConfig := map[string]interface{}{}

	maxTokens := 2048
	if options != nil {
		switch mt := options["max_tokens"].(type) {
		case int:
			maxTokens = mt
		case float64:
			maxTokens = int(mt)
		}
		if temperature, ok := options["temperature"].(float64); ok {
			generationConfig["temperature"] = temperature
			options["__sent_temperature"] = temperature
		}
	}
	generationConfig["maxOutputTokens"] = maxTokens
	if options != nil {
		options["__sent_max_tokens"] = maxTokens
	}

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": []map[string]string{{"text": prompt}},
			},
		},
		"generationConfig": generationConfig,
	}

	if options != nil {
		if systemPrompt, ok := options["system"].(string); ok && systemPrompt != "" {
			requestBody["systemInstruction"] = map[string]interface{}{
				"parts": []map[string]string{{"text": systemPrompt}},
			}
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent", geminiAPIBase, c.model)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason"`
		} `json:"promptFeedback"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if options != nil {
		options["__usage_input_tokens"] = response.UsageMetadata.PromptTokenCount
		options["__usage_output_tokens"] = response.UsageMetadata.CandidatesTokenCount
	}

	// The prompt itself can be blocked before any candidate is produced —
	// there is no candidate to inspect at all in that case, unlike a
	// per-candidate finishReason below.
	if response.PromptFeedback.BlockReason != "" {
		return "", &RefusalError{
			Reason:   fmt.Sprintf("promptFeedback.blockReason=%s", response.PromptFeedback.BlockReason),
			Blocks:   nil,
			Provider: "gemini",
			Model:    c.model,
		}
	}

	if len(response.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := response.Candidates[0]
	var textBuilder strings.Builder
	for _, part := range candidate.Content.Parts {
		textBuilder.WriteString(part.Text)
	}
	partial := textBuilder.String()

	// Mirrors anthropic.go: extract the partial BEFORE deciding whether this
	// was a truncation or refusal, so a cut or declined completion still
	// carries back whatever text the model produced (bugs_open/019).
	if candidate.FinishReason == "MAX_TOKENS" {
		return partial, &TruncatedError{
			Partial:      partial,
			OutputTokens: response.UsageMetadata.CandidatesTokenCount,
			Reason:       "finishReason=MAX_TOKENS",
			Provider:     "gemini",
		}
	}

	switch candidate.FinishReason {
	case "SAFETY", "RECITATION", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII":
		return "", &RefusalError{
			Reason:   fmt.Sprintf("finishReason=%s", candidate.FinishReason),
			Blocks:   []string{candidate.FinishReason},
			Provider: "gemini",
			Model:    c.model,
		}
	}

	if partial == "" {
		return "", fmt.Errorf("no text content in response (finishReason=%q)", candidate.FinishReason)
	}

	return partial, nil
}

// GenerateEmbedding generates embeddings using Gemini's embedding model
func (c *GeminiClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	requestBody := map[string]interface{}{
		"model": "models/text-embedding-004",
		"content": map[string]interface{}{
			"parts": []map[string]string{{"text": text}},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	url := fmt.Sprintf("%s/text-embedding-004:embedContent", geminiAPIBase)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(response.Embedding.Values) == 0 {
		return nil, fmt.Errorf("empty embedding from gemini")
	}

	return response.Embedding.Values, nil
}

func (c *GeminiClient) Model() string {
	return c.model
}

func (c *GeminiClient) Provider() string {
	return "gemini"
}
