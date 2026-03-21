// FILE: platform/aiservice/ollama.go
//
// Ollama provider for locally-hosted models.
// Implements the same AIService interface as AnthropicClient.
//
// Agent definitions use it like:
//   "ai_service": {
//       "provider": "ollama",
//       "model": "nomic-embed-text",
//       "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
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
	"time"
)

type OllamaClient struct {
	apiURL     string
	model      string
	httpClient *http.Client
}

func NewOllamaClient(ctx context.Context, config map[string]interface{}) (*OllamaClient, error) {
	if config == nil {
		return nil, fmt.Errorf("ai_service config is nil for ollama provider")
	}

	apiURL := "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
	if urlRaw, ok := config["api_url"].(string); ok && urlRaw != "" {
		apiURL = urlRaw
	} else if envURL := os.Getenv("OLLAMA_API_URL"); envURL != "" {
		apiURL = envURL
	}
	apiURL = strings.TrimRight(apiURL, "/")

	model, ok := config["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("model not specified in ollama ai_service config")
	}

	return &OllamaClient{
		apiURL: apiURL,
		model:  model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

func (c *OllamaClient) GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	requestBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.7,
		},
	}

	if options != nil {
		optionsBlock := requestBody["options"].(map[string]interface{})
		if temperature, ok := options["temperature"].(float64); ok {
			optionsBlock["temperature"] = temperature
		}
		if maxTokens, ok := options["max_tokens"].(float64); ok {
			optionsBlock["num_predict"] = int(maxTokens)
		}

		// System prompt support
		if systemPrompt, ok := options["system"].(string); ok && systemPrompt != "" {
			msgs := requestBody["messages"].([]map[string]string)
			requestBody["messages"] = append([]map[string]string{
				{"role": "system", "content": systemPrompt},
			}, msgs...)
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ollama request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/api/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read ollama response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Model           string `json:"model"`
		PromptEvalCount int    `json:"prompt_eval_count"`
		EvalCount       int    `json:"eval_count"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse ollama response: %w", err)
	}

	// Write usage back to options for logging
	if options != nil {
		options["__usage_input_tokens"] = response.PromptEvalCount
		options["__usage_output_tokens"] = response.EvalCount
		if response.Model != "" {
			options["__model_used"] = response.Model
		}
	}

	return response.Message.Content, nil
}

func (c *OllamaClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	requestBody := map[string]interface{}{
		"model":  c.model,
		"prompt": text,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/api/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embedding API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding from ollama model %s", c.model)
	}

	return response.Embedding, nil
}

func (c *OllamaClient) Model() string {
	return c.model
}

func (c *OllamaClient) Provider() string {
	return "ollama"
}
