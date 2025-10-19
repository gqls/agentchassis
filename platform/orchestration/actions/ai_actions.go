// FILE: platform/orchestration/actions/ai_actions.go
package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Fake response
func ExecuteLLMPromptActionFAKE(ctx context.Context, params ActionParams) (interface{}, error) {
	// Fake parsed result to return instead of calling the AI
	parsedResult := map[string]interface{}{
		"summary": "This is a fake response for testing.",
		"status":  "success",
		"results": map[string]interface{}{
			"business_name": "Test Company",
			"domain":        "test.com",
			"description":   "A test company",
			"analysis":      "THIS IS PLACEHOLDER ANALYSIS DATA.",
		},
	}

	return map[string]interface{}{
		"result": parsedResult,
		"type":   "json",
	}, nil

}

func ExecuteLLMPromptAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing LLM prompt action",
		zap.String("agent_type", params.AgentType),
		zap.Any("collected_data_keys", GetMapKeys(params.CollectedData)),
		zap.String("action", params.ExecutionContext.Action),
		zap.Bool("has_db", params.DB != nil),
		zap.Any("DEBUGaa: full params in ExecuteLLMPromptAction", params),
	)

	// initialise
	if params.ExecutionContext.Action == "initialize" {
		params.Logger.Info("handling initialization")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 2: Check for initialization flag in collected data
	if isInit, ok := params.CollectedData["is_initialization"].(bool); ok && isInit {
		params.Logger.Info("initialization detected via flag")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Method 3: Check the action from collected data (if passed through)
	if action, ok := params.CollectedData["action"].(string); ok && action == "initialize" {
		params.Logger.Info("initialization detected via action field")
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Normalize the collected data using the helper
	normalizedData := datahelpers.NormalizeCollectedData(
		params.CollectedData,
		params.ExecutionContext,
		params.ExecutionContext.RequestsTopic,
		params.Logger,
	)

	// Update params.CollectedData with normalized version
	params.CollectedData = normalizedData

	// Get the agent's configuration
	var agentConfig map[string]interface{}
	var ok bool

	// First try CollectedData (for orchestrated agents)
	agentConfig, ok = params.CollectedData["agent_config"].(map[string]interface{})
	params.Logger.Info("Checking agent_config in CollectedData",
		zap.Bool("found", ok),
		zap.Bool("is_nil", agentConfig == nil))

	// If not found, load it directly from the database
	if !ok && params.AgentType != "" {
		params.Logger.Info("Agent config not in collected data, loading from database",
			zap.String("agent_type", params.AgentType))

		agentDef, err := loadAgentDefinitionForAction(ctx, params.DB, params.AgentType)
		if err != nil {
			params.Logger.Error("Failed to load agent definition",
				zap.String("agent_type", params.AgentType),
				zap.Error(err))
			return nil, fmt.Errorf("failed to load agent definition: %w", err)
		}

		agentConfig = agentDef.DefaultConfig
		params.CollectedData["agent_config"] = agentConfig
	}

	if agentConfig == nil {
		params.Logger.Error("Agent configuration is nil after all attempts")
		return nil, fmt.Errorf("agent configuration not found")
	}

	// Extract AI service configuration
	aiServiceConfig, ok := agentConfig["ai_service"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("ai_service configuration not found in agent config")
	}

	// Extract prompt template
	// THREE-TIER PRIORITY:
	// Get prompt using three-tier priority system
	promptTemplate, promptSource := getPromptWithPriority(params, agentConfig)

	params.Logger.Info("Selected prompt for execution",
		zap.String("source", promptSource),
		zap.String("agent_type", params.AgentType),
		zap.String("prompt_preview", truncateString(promptTemplate, 350)))

	// Create AI client based on provider
	aiClient, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	extractedData := extractDataForAiAgent(params)
	templateData := extractedData.(map[string]interface{})

	params.Logger.Info("in ExecuteLLMPromptAction Template Data",
		// zap.Any("template_data DEBUGaa", templateData),
		zap.Any("DEBUGaa template_data ai_actions this is what I want to pass - should be good ", templateData), // is good
		zap.Any("agent_config", agentConfig),
		//zap.Any("DEBUGaa: params.CollectedData[input_data] when extracting data in ai actions", params.CollectedData["input_data"]), // neither data or template is not in here
		zap.Any("promptTemplate", promptTemplate), // good
	)

	// Render the prompt template
	renderedPrompt, err := renderPromptTemplate(promptTemplate, templateData, *params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt template: %w", err)
	}

	params.Logger.Info("Rendered prompt template",
		zap.String("template_preview", truncateString(promptTemplate, 300)),
		zap.String("rendered_preview - renderedPrompt", truncateString(renderedPrompt, 400)))

	// Prepare AI service options
	options := make(map[string]interface{})
	if model, ok := aiServiceConfig["model"].(string); ok {
		options["model"] = model
	}
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

// Add this helper function to load agent definition
func loadAgentDefinitionForAction(ctx context.Context, db interface{}, agentType string) (*AgentDefinition, error) {

	fmt.Printf("DEBUG: loadAgentDefinitionForAction called with agentType=%s, db type=%T\n", agentType, db)

	query := `
		SELECT id, type, display_name, description, category,
		       image_repository, image_tag, 
		       resources, default_config, capabilities, topics,
		       health_config, env_vars, is_active
		FROM agent_definitions
		WHERE type = $1 AND is_active = true
		LIMIT 1
	`

	var def AgentDefinition
	var defaultConfigJSON json.RawMessage // Read as RawMessage first
	var capabilitiesJSON json.RawMessage

	// Handle both *sql.DB and *pgxpool.Pool
	switch d := db.(type) {
	case *sql.DB:
		// For *sql.DB, we need to handle the Command field differently
		err := d.QueryRowContext(ctx, query, agentType).Scan(
			&def.ID,
			&def.Type,
			&def.DisplayName,
			&def.Description,
			&def.Category,
			&def.ImageRepository,
			&def.ImageTag,
			&def.Resources,
			&def.DefaultConfig,
			&def.Capabilities,
			&def.Topics,
			&def.HealthConfig,
			&def.EnvVars,
			&def.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query agent definition: %w", err)
		}
		// Note: Command field is not loaded here as it's not needed for LLM prompt
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, agentType).Scan(
			&def.ID,
			&def.Type,
			&def.DisplayName,
			&def.Description,
			&def.Category,
			&def.ImageRepository,
			&def.ImageTag,
			&def.Resources,
			&def.DefaultConfig,
			&def.Capabilities,
			&def.Topics,
			&def.HealthConfig,
			&def.EnvVars,
			&def.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query agent definition: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	// Now unmarshal the JSON into the map
	if len(defaultConfigJSON) > 0 && string(defaultConfigJSON) != "null" {
		if err := json.Unmarshal(defaultConfigJSON, &def.DefaultConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal default_config: %w", err)
		}
	}

	// Unmarshal capabilities
	if len(capabilitiesJSON) > 0 {
		json.Unmarshal(capabilitiesJSON, &def.Capabilities)
	}

	// Validate that we have a config
	if def.DefaultConfig == nil {
		return nil, fmt.Errorf("agent %s has no default config", agentType)
	}

	return &def, nil
}

// Keep all existing helper functions...
func createAIClient(ctx context.Context, aiServiceConfig map[string]interface{}) (aiservice.AIService, error) {
	provider, ok := aiServiceConfig["provider"].(string)
	if !ok {
		return nil, fmt.Errorf("provider not specified in ai_service config")
	}

	switch provider {
	case "anthropic":
		return aiservice.NewAnthropicClient(ctx, aiServiceConfig)
	case "openai":
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

func renderPromptTemplate(templateStr string, data map[string]interface{}, logger zap.Logger) (string, error) {
	tmpl := template.New("agent_prompt")
	parsedTemplate, err := tmpl.Parse(templateStr)
	logger.Info("DEBUGaa: parsing template in renderTemplate",
		zap.String("template", templateStr),
		zap.Any("data", data),
		zap.Any("tmpl", tmpl),
		zap.Any("parsedTemplate", parsedTemplate),
	)
	if err != nil {
		return "", fmt.Errorf("failed to parse template in render template: %w", err)
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

func ConditionalRouteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	config := params.StepConfig.Config

	conditionField := config["condition_field"].(string)
	routes := config["routes"].(map[string]interface{})

	// Evaluate the condition
	// Check for condition value in input_data first
	var conditionValue interface{}

	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		// Look in input_data first
		if val, exists := inputData[conditionField]; exists {
			conditionValue = val
		}
	}

	// Fall back to collected data
	if conditionValue == nil {
		conditionValue = params.CollectedData[conditionField]
	}

	// If no explicit condition value, evaluate it
	if conditionValue == nil {
		conditionValue = evaluateCondition(params)
		params.CollectedData[conditionField] = conditionValue
	}

	// Determine next step
	nextStep, ok := routes[fmt.Sprintf("%v", conditionValue)]
	if !ok {
		// Use default route if available
		nextStep = routes["default"]
	}

	return map[string]interface{}{
		"next_step": nextStep,
		"condition": conditionValue,
	}, nil
}

func evaluateCondition(params ActionParams) interface{} {
	// Simple complexity evaluation
	inputSize := len(params.CollectedData)

	if inputSize < 3 {
		return "simple"
	} else if inputSize < 10 {
		return "moderate"
	}
	return "complex"
}

func EvaluateTaskAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Evaluating task complexity")

	complexity := "simple" // Default

	// Check input data size and structure
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		dataSize := len(inputData)

		// Check for indicators of complexity
		hasNestedData := false

		for _, v := range inputData {
			switch v.(type) {
			case map[string]interface{}, []interface{}:
				hasNestedData = true
			}
		}

		// Determine complexity
		if dataSize > 10 || hasNestedData {
			complexity = "complex"
		} else if dataSize > 5 {
			complexity = "moderate"
		}
	}

	params.Logger.Info("Task complexity evaluated",
		zap.String("complexity", complexity))

	return map[string]interface{}{
		"complexity": complexity,
	}, nil
}

func extractDataForAiAgent(params ActionParams) interface{} {
	// Get which input field to use - ditching this for now
	/*inputField := "input_data"
	if field, ok := params.StepConfig.Config["input_field"].(string); ok && field != "" {
		inputField = field
	}*/

	params.Logger.Info("Extracting data for ai agent",
		//zap.String("requested_field", inputField), // input_data
		zap.Any("available_keys", getMapKeys(params.CollectedData)),
		zap.Any("DEBUGaa: params.CollectedData for passing to the new agent in executellmprompt action extractDataForAiAgent", params.CollectedData),
	)

	// Use the GetInputData helper which handles all the extraction logic
	cleanData := datahelpers.GetInputData(params.CollectedData, params.Logger)

	// If we got a valid map, return it
	if len(cleanData) > 0 {
		params.Logger.Info("Extracted clean data for AI agent",
			zap.Int("field_count", len(cleanData)))
		return cleanData
	}

	// Fallback to empty map
	params.Logger.Warn("No data found for AI agent, using empty map")
	return make(map[string]interface{})

	// Define search paths from most to least specific
	/*searchPaths := [][]string{
		{"input_data", "body", "input_data", inputField}, // Deeply nested with body
		{"input_data", "input_data", inputField},         // Deeply nested
		{"input_data", "body", inputField},               // medium nested with body
		{"input_data", inputField},                       // Medium nested
		{inputField},                                     // Top level
	}

	// Try each path
	for _, path := range searchPaths {
		if data, ok := getNestedInputValue(params.CollectedData, path...); ok {
			params.Logger.Info("Found data via path",
				zap.Strings("path", path),
				zap.Any("DEBUGaa: data found via path", data),
			)
			return data
		}
	}*/

	// Special case: if looking for input_data, try nested version
	/*	if inputField == "input_data" {
			if data, ok := getNestedInputValue(params.CollectedData, "input_data", "input_data"); ok {
				params.Logger.Info("Using nested input_data")
				return data
			}
		}

		// Final fallback
		data := params.CollectedData["input_data"]
		if data == nil {
			params.Logger.Warn("No data found, using empty map")
			return make(map[string]interface{})
		}

		params.Logger.Info("Using top-level input_data as fallback")
		return data*/
}

// getPromptWithPriority implements three-tier priority for prompt selection
func getPromptWithPriority(params ActionParams, agentConfig map[string]interface{}) (prompt string, source string) {
	logger := params.Logger

	logger.Info("in getPromptWithPriority")

	// PRIORITY 1: Check incoming message for prompt (from parent's call_agent)
	// Check in StepConfig.Config first (this is where call_agent passes it)
	if configPrompt, ok := params.StepConfig.Config["prompt"].(string); ok && configPrompt != "" {
		logger.Info("Using prompt from step config (Priority 1 - from parent)",
			zap.String("prompt_preview", truncateString(configPrompt, 100)))
		return configPrompt, "parent_message"
	}

	// Also check in CollectedData["prompt"] as a fallback
	if collectedPrompt, ok := params.CollectedData["prompt"].(string); ok && collectedPrompt != "" {
		logger.Info("Using prompt from collected data (Priority 1 - from parent)",
			zap.String("prompt_preview", truncateString(collectedPrompt, 100)))
		return collectedPrompt, "parent_message"
	}

	// PRIORITY 2: Check agent's own default_config.prompt_template
	// This comes from the agent_definitions table for this specific agent type
	if agentPrompt, ok := agentConfig["prompt_template"].(string); ok && agentPrompt != "" {
		logger.Info("Using prompt from agent config (Priority 2 - agent's default)",
			zap.String("agent_type", params.AgentType),
			zap.String("prompt_preview", truncateString(agentPrompt, 100)))
		return agentPrompt, "agent_default"
	}

	// PRIORITY 3: Check workflow step config (fallback)
	// This is the hardcoded fallback in the workflow definition
	if stepConfig, ok := params.StepConfig.Config["prompt_template"].(string); ok && stepConfig != "" {
		logger.Info("Using prompt from workflow step config (Priority 3 - fallback)",
			zap.String("prompt_preview", truncateString(stepConfig, 100)))
		return stepConfig, "workflow_fallback"
	}

	// Generic fallback if nothing found
	logger.Warn("No prompt found in any tier, using generic fallback")
	return "Generate content based on the provided context.", "generic_fallback"
}
