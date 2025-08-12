// FILE: platform/orchestration/actions/ai_actions.go
package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/gqls/agentchassis/platform/aiservice"
	"go.uber.org/zap"
)

// ExecuteLLMPromptAction executes an LLM prompt using the agent's configured template
// 1. Get agent's own config (loaded by MessageProcessor and available in params.CollectedData).
// 2. Extract `prompt_template` and `ai_service` config from the agent's config.
// 3. Dynamically initialize the correct AI client (e.g., NewAnthropicClient, NewOpenAIClient) based on the "provider".
// 4. Create a data map for the template using the `params.CollectedData` from previous steps.
// 5. Use Go's `text/template` package to render the final prompt string, filling in the placeholders.
// 6. Call `aiClient.GenerateText(ctx, renderedPrompt, nil)`.
// 7. Return the raw text response from the LLM.
func ExecuteLLMPromptAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing LLM prompt action")

	// Get the agent's configuration (loaded by MessageProcessor)
	agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("agent configuration not found in collected data")
	}

	// Extract AI service configuration
	aiServiceConfig, ok := agentConfig["ai_service"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ai_service configuration not found in agent config")
	}

	// Extract prompt template
	promptTemplate, ok := agentConfig["prompt_template"].(string)
	if !ok {
		return nil, fmt.Errorf("prompt_template not found in agent config")
	}

	// Create AI client based on provider
	aiClient, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	// Prepare template data - merge all collected data and input data
	templateData := make(map[string]interface{})

	// Add all collected data from previous workflow steps
	for key, value := range params.CollectedData {
		templateData[key] = value
	}

	// Parse the input data if it exists
	if params.InputData != nil {
		var inputPayload map[string]interface{}
		if err := json.Unmarshal(params.InputData, &inputPayload); err == nil {
			// Add the data field as "input" for template access
			if data, ok := inputPayload["data"].(map[string]interface{}); ok {
				templateData["input"] = data
			}
		}
	}

	// Render the prompt template
	renderedPrompt, err := renderTemplate(promptTemplate, templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt template: %w", err)
	}

	params.Logger.Info("Rendered prompt template",
		zap.String("template_preview", truncateString(promptTemplate, 100)),
		zap.String("rendered_preview", truncateString(renderedPrompt, 200)))

	// Prepare AI service options
	options := make(map[string]interface{})
	if model, ok := aiServiceConfig["model"].(string); ok {
		options["model"] = model
	}
	// Check for temperature in agent config
	if temp, ok := agentConfig["temperature"].(float64); ok {
		options["temperature"] = temp
	}
	if maxTokens, ok := agentConfig["max_tokens"].(float64); ok {
		options["max_tokens"] = int(maxTokens)
	}

	// Call the AI service
	result, err := aiClient.GenerateText(ctx, renderedPrompt, options)
	if err != nil {
		return nil, fmt.Errorf("AI service call failed: %w", err)
	}

	params.Logger.Info("LLM response received",
		zap.String("result_preview", truncateString(result, 200)))

	// Try to parse as JSON, if it fails return as plain text
	var parsedResult interface{}
	if err := json.Unmarshal([]byte(result), &parsedResult); err != nil {
		// Not valid JSON, return as plain text
		return map[string]interface{}{
			"result": result,
			"type":   "text",
		}, nil
	}

	// Valid JSON, return parsed result
	return map[string]interface{}{
		"result": parsedResult,
		"type":   "json",
	}, nil
}

// Helper functions remain the same...
func createAIClient(ctx context.Context, aiServiceConfig map[string]interface{}) (aiservice.AIService, error) {
	provider, ok := aiServiceConfig["provider"].(string)
	if !ok {
		return nil, fmt.Errorf("provider not specified in ai_service config")
	}

	switch provider {
	case "anthropic":
		return aiservice.NewAnthropicClient(ctx, aiServiceConfig)
	case "openai":
		// Add OpenAI support if needed
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

func renderTemplate(templateStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("agent_prompt").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}
