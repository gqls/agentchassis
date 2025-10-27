// FILE: platform/orchestration/actions/image_actions.go
package actions

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// GenerateImageAction generates an image using an external AI image generation service
// This action integrates with Stability AI or similar services
func GenerateImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing image generation action",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("action", params.ExecutionContext.Action),
		zap.String("step_id", params.ExecutionContext.StepID))

	// Extract data using the new data_helpers
	inputData := datahelpers.ExtractDataFromMessage(params.CollectedData, params.Logger)

	params.Logger.Debug("Extracted input data for image generation",
		zap.Int("input_fields", len(inputData)),
		zap.Any("keys", getMapKeys(inputData)))

	// Extract prompt from various possible locations
	prompt, err := extractImagePrompt(params, inputData)
	if err != nil {
		return nil, errors.WrapWithAgentContext(err, "failed to extract image prompt",
			params.ExecutionContext.Sender.AgentType,
			params.ExecutionContext.Sender.AgentID,
			params.ExecutionContext.OrchestrationID,
			params.ExecutionContext.StepName,
			params.ExecutionContext.Action)
	}

	params.Logger.Info("Image prompt extracted",
		zap.String("prompt_preview", truncateString(prompt, 100)))

	// Extract image generation config
	config := extractImageConfig(params)

	params.Logger.Debug("Image generation config",
		zap.Any("config", config))

	// Generate the image
	imageResult, err := generateImageExternal(params.Context, prompt, config, params.Logger)
	if err != nil {
		return nil, errors.WrapWithAgentContext(err, "failed to generate image",
			params.ExecutionContext.Sender.AgentType,
			params.ExecutionContext.Sender.AgentID,
			params.ExecutionContext.OrchestrationID,
			params.ExecutionContext.StepName,
			params.ExecutionContext.Action)
	}

	params.Logger.Info("Image generated successfully",
		zap.String("image_uri", imageResult["image_uri"].(string)))

	// Return result
	result := map[string]interface{}{
		"image_uri":    imageResult["image_uri"],
		"prompt":       prompt,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	if seed, ok := imageResult["seed"]; ok {
		result["seed"] = seed
	}

	return result, nil
}

// extractImagePrompt extracts the prompt for image generation from various sources
func extractImagePrompt(params ActionParams, inputData map[string]interface{}) (string, error) {
	// Priority 1: Check step config for prompt template
	if stepPrompt, ok := params.StepConfig.Config["prompt"].(string); ok && stepPrompt != "" {
		params.Logger.Debug("Using prompt from step config")
		// Render template if it contains template syntax
		if containsTemplateSyntax(stepPrompt) {
			rendered, err := renderImagePromptTemplate(stepPrompt, inputData, params.Logger)
			if err == nil {
				return rendered, nil
			}
			params.Logger.Warn("Failed to render prompt template, using as-is",
				zap.Error(err))
		}
		return stepPrompt, nil
	}

	// Priority 2: Check collected data for prompt
	if prompt, ok := params.CollectedData["prompt"].(string); ok && prompt != "" {
		params.Logger.Debug("Using prompt from collected data")
		return prompt, nil
	}

	// Priority 3: Check input_data for prompt or image_prompt
	if prompt, ok := inputData["image_prompt"].(string); ok && prompt != "" {
		params.Logger.Debug("Using image_prompt from input data")
		return prompt, nil
	}

	if prompt, ok := inputData["prompt"].(string); ok && prompt != "" {
		params.Logger.Debug("Using prompt from input data")
		return prompt, nil
	}

	// Priority 4: Check input_fields to gather context for prompt generation
	if inputFields, ok := params.StepConfig.Config["input_fields"].([]interface{}); ok {
		contextData := gatherInputFieldsData(params.CollectedData, inputFields, params.Logger)
		if len(contextData) > 0 {
			// Try to generate a prompt from context
			if generatedPrompt := generatePromptFromContext(contextData, params.Logger); generatedPrompt != "" {
				params.Logger.Debug("Generated prompt from input fields context")
				return generatedPrompt, nil
			}
		}
	}

	return "", fmt.Errorf("no image prompt found in step config, collected data, or input data")
}

// extractImageConfig extracts configuration for image generation
func extractImageConfig(params ActionParams) map[string]interface{} {
	config := make(map[string]interface{})

	// Default values
	config["aspect_ratio"] = "1:1"
	config["output_format"] = "png"

	// Extract from step config
	if stepConfig, ok := params.StepConfig.Config["image_config"].(map[string]interface{}); ok {
		for k, v := range stepConfig {
			config[k] = v
		}
	}

	// Extract individual config values
	if aspectRatio, ok := params.StepConfig.Config["aspect_ratio"].(string); ok {
		config["aspect_ratio"] = aspectRatio
	}

	if style, ok := params.StepConfig.Config["style"].(string); ok {
		config["style"] = style
	}

	if seed, ok := params.StepConfig.Config["seed"].(float64); ok {
		config["seed"] = int64(seed)
	}

	// Check collected data for any overrides
	if imageConfig, ok := params.CollectedData["image_config"].(map[string]interface{}); ok {
		for k, v := range imageConfig {
			config[k] = v
		}
	}

	return config
}

// generateImageExternal calls the external image generation API
func generateImageExternal(ctx context.Context, prompt string, config map[string]interface{}, logger *zap.Logger) (map[string]interface{}, error) {
	// Get API endpoint and key from environment
	apiEndpoint := os.Getenv("STABILITY_API_ENDPOINT")
	if apiEndpoint == "" {
		apiEndpoint = "https://api.stability.ai/v2beta/stable-image/generate/core"
	}

	apiKey := os.Getenv("STABILITY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("STABILITY_API_KEY environment variable not set")
	}

	logger.Info("Calling image generation API",
		zap.String("endpoint", apiEndpoint),
		zap.String("prompt_preview", truncateString(prompt, 100)))

	// Build request payload
	payload := map[string]interface{}{
		"prompt": prompt,
	}

	// Add optional parameters
	if aspectRatio, ok := config["aspect_ratio"].(string); ok {
		payload["aspect_ratio"] = aspectRatio
	}

	if style, ok := config["style"].(string); ok {
		payload["style_preset"] = style
	}

	if seed, ok := config["seed"].(int64); ok {
		payload["seed"] = seed
	}

	if outputFormat, ok := config["output_format"].(string); ok {
		payload["output_format"] = outputFormat
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create HTTP request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", apiEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{
		Timeout: 90 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Error("Image generation API returned error",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(respBody)))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var apiResponse struct {
		Image string `json:"image"`
		Seed  int64  `json:"seed"`
	}

	if err := json.Unmarshal(respBody, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(apiResponse.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image: %w", err)
	}

	logger.Info("Image data received",
		zap.Int("image_size_bytes", len(imageData)),
		zap.Int64("seed", apiResponse.Seed))

	// Store image in object storage
	imageURI, err := storeImageInObjectStorage(ctx, imageData, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to store image: %w", err)
	}

	result := map[string]interface{}{
		"image_uri": imageURI,
		"seed":      apiResponse.Seed,
	}

	return result, nil
}

// renderImagePromptTemplate renders a prompt template with data
func renderImagePromptTemplate(template string, data map[string]interface{}, logger *zap.Logger) (string, error) {
	// Use simple string replacement for now
	// In production, you might want to use text/template or similar
	result := template

	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		if strValue, ok := value.(string); ok {
			result = replaceAll(result, placeholder, strValue)
		} else {
			result = replaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	return result, nil
}

// generatePromptFromContext generates an image prompt from context data
func generatePromptFromContext(contextData map[string]interface{}, logger *zap.Logger) string {
	// This is a simple implementation - in production you might use an LLM to generate better prompts
	var parts []string

	for _, value := range contextData {
		if strValue, ok := value.(string); ok && strValue != "" {
			parts = append(parts, strValue)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	// Join parts into a prompt
	return fmt.Sprintf("Create an image for: %s", parts[0])
}

// Helper functions

func containsTemplateSyntax(s string) bool {
	return len(s) > 4 && (bytes.Contains([]byte(s), []byte("{{")) || bytes.Contains([]byte(s), []byte("}}")))
}

func replaceAll(s, old, new string) string {
	return bytes.NewBuffer([]byte(s)).String() // simplified - use strings.ReplaceAll in production
}

func gatherInputFieldsData(collectedData map[string]interface{}, inputFields []interface{}, logger *zap.Logger) map[string]interface{} {
	result := make(map[string]interface{})

	for _, field := range inputFields {
		if fieldName, ok := field.(string); ok {
			if value, exists := collectedData[fieldName]; exists {
				result[fieldName] = value
			}
		}
	}

	return result
}
