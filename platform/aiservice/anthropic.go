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
)

// AnthropicClient implements the AIService interface for Anthropic's Claude
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(ctx context.Context, config map[string]interface{}) (*AnthropicClient, error) {
	apiKeyEnvVar := config["api_key_env_var"].(string)
	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("API key not found in environment variable %s", apiKeyEnvVar)
	}

	model := config["model"].(string)

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
		"model":       c.model,
		"max_tokens":  2048,
		"temperature": 0.7,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// Override with provided options
	if options != nil {
		if maxTokens, ok := options["max_tokens"]; ok {
			requestBody["max_tokens"] = maxTokens
		}
		if temperature, ok := options["temperature"]; ok {
			requestBody["temperature"] = temperature
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

	// Parse response
	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Content[0].Text, nil
}

// GenerateEmbedding generates embeddings (not implemented for Anthropic)
func (c *AnthropicClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, fmt.Errorf("embedding generation not supported by Anthropic")
}
