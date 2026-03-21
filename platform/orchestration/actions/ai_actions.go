// FILE: platform/orchestration/actions/ai_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
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

	/*	// Normalize the collected data using the helper
		normalizedData := datahelpers.NormalizeCollectedData(
			params.CollectedData,
			params.ExecutionContext,
			params.ExecutionContext.RequestsTopic,
			params.Logger,
		)

		// Update params.CollectedData with normalized version
		params.CollectedData = normalizedData*/

	// removed NormalizeCollectedData was destroying accumulated state (loop vars, render_context, etc.)
	// The params.CollectedData already has the correct state from the coordinator.
	// Only ensure essential topic fields are present if missing:
	if params.ExecutionContext.RequestsTopic != "" {
		if _, exists := params.CollectedData["__my_requests_topic__"]; !exists {
			params.CollectedData["__my_requests_topic__"] = params.ExecutionContext.RequestsTopic
		}
	}
	if params.ExecutionContext.ReplyToTopic != "" {
		if _, exists := params.CollectedData["__parent_responses_topic__"]; !exists {
			params.CollectedData["__parent_responses_topic__"] = params.ExecutionContext.ReplyToTopic
		}
	}

	params.Logger.Debug("ExecuteLLMPromptAction: preserving accumulated CollectedData",
		zap.Strings("keys", datahelpers.GetMapKeys(params.CollectedData)),
		zap.Bool("has_current_section", params.CollectedData["current_section"] != nil),
	)

	// Get the agent's configuration
	var agentConfig map[string]interface{}
	var ok bool

	// First try CollectedData (for orchestrated agents)
	agentConfig, ok = params.CollectedData["agent_config"].(map[string]interface{})
	params.Logger.Info("Checking agent_config in CollectedData",
		zap.Bool("found", ok),
		zap.Bool("is_nil", agentConfig == nil),
		zap.Any("agentConfig in first try", agentConfig),
		zap.Any("Config in first try", agentConfig),
		zap.String("step name is", params.ExecutionContext.StepName),
		//zap.Any("DEBUGaa: collected Data at this stage is:", params.CollectedData),
	)

	currentStep := params.ExecutionContext.StepName

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

	config, ok := params.CollectedData["config"].(map[string]interface{})
	if !ok {
		params.Logger.Warn("Failed to load normal config (overridden config)",
			zap.String("agent_type", params.AgentType),
		)
	}

	params.Logger.Info("Checking normal just config in CollectedData",
		zap.Any("config", config),
	)

	// Extract AI service configuration
	// First try to get ai_service from top-level agent_config
	var aiServiceConfig map[string]interface{}

	// Check if agent_config has ai_service at top level
	if agentConfig != nil {
		if aiService, ok := agentConfig["ai_service"].(map[string]interface{}); ok && aiService != nil {
			aiServiceConfig = aiService
			params.Logger.Info("Found ai_service at top level of agent_config",
				zap.Any("ai_service", aiService))
		}
	}

	// If not found at top level, check in step config
	if aiServiceConfig == nil {
		params.Logger.Info("ai_service not at top level, checking step config")

		// Look in the workflow steps for the current step's config
		if workflow, ok := agentConfig["workflow"].(map[string]interface{}); ok {
			if steps, ok := workflow["steps"].(map[string]interface{}); ok {
				if currentStepConfig, ok := steps[currentStep].(map[string]interface{}); ok {
					if stepConfig, ok := currentStepConfig["config"].(map[string]interface{}); ok {
						if aiService, ok := stepConfig["ai_service"].(map[string]interface{}); ok {
							aiServiceConfig = aiService
							params.Logger.Info("Found ai_service in step config",
								zap.String("step", currentStep),
								zap.Any("ai_service", aiService))
						}
					}
				}
			}
		}
	}

	// Also check StepConfig if provided
	if aiServiceConfig == nil && params.StepConfig.Config != nil {
		if aiService, ok := params.StepConfig.Config["ai_service"].(map[string]interface{}); ok {
			aiServiceConfig = aiService
			params.Logger.Info("Found ai_service in StepConfig",
				zap.Any("ai_service", aiService))
		}
	}

	if aiServiceConfig == nil || len(aiServiceConfig) == 0 {
		params.Logger.Error("ai_service configuration not found after checking all locations",
			zap.String("checked_locations", "agent_config top-level, workflow.steps.config, StepConfig"))
		return nil, fmt.Errorf("ai_service configuration not found")
	}

	// Extract prompt template
	// THREE-TIER PRIORITY:
	// Get prompt using three-tier priority system
	promptTemplate, promptSource := getPromptWithPriority(params, agentConfig)

	params.Logger.Info("Selected prompt for execution",
		zap.String("source", promptSource),
		zap.String("agent_type", params.AgentType),
		zap.String("prompt_preview", datahelpers.TruncateString(promptTemplate, 350)))

	// Create AI client based on provider
	aiClient, err := createAIClient(ctx, aiServiceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	params.Logger.Info("in ExecuteLLMPromptAction data from which were trying to extract templatedata",
		zap.Any("DEBUGaa: params sent to extractDataForAIAgent hoping for the correct cleandata", params),
	)

	extractedData := extractDataForAiAgent(params)
	templateData := extractedData.(map[string]interface{})

	validateTemplateData(templateData, params.StepConfig.Config, params.Logger)

	params.Logger.Info("in ExecuteLLMPromptAction Template Data",
		// zap.Any("template_data DEBUGaa", templateData),
		zap.Any("DEBUGaa template_data ai_actions this is what I want to pass - should be good ", templateData), // is good
		zap.Any("agent_config", agentConfig),
		//zap.Any("DEBUGaa: params.CollectedData[input_data] when extracting data in ai actions", params.CollectedData["input_data"]), // neither data or template is not in here
		zap.Any("promptTemplate", promptTemplate), // good
	)

	// Render the prompt template
	renderedPrompt, err := datahelpers.RenderPromptTemplate(promptTemplate, templateData, *params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt template: %w", err)
	}

	params.Logger.Info("Rendered prompt template",
		zap.String("template_preview", datahelpers.TruncateString(promptTemplate, 300)),
		zap.String("rendered_preview - renderedPrompt", datahelpers.TruncateString(renderedPrompt, 400)))

	// Append output format instructions based on output_type
	// Check both step config and ai_service config for output_type
	renderedPrompt = appendOutputInstructions(renderedPrompt, aiServiceConfig, params.StepConfig.Config, params.Logger)

	// Prepare AI service options
	options := make(map[string]interface{})
	if model, ok := aiServiceConfig["model"].(string); ok {
		options["model"] = model
	}
	if temp, ok := agentConfig["temperature"].(float64); ok {
		options["temperature"] = temp
	}

	var maxTokens float64
	if maxTokens, ok = agentConfig["max_tokens"].(float64); ok {
		options["max_tokens"] = int(maxTokens)
	} else if maxTokens, ok = aiServiceConfig["max_tokens"].(float64); ok {
		options["max_tokens"] = int(maxTokens)
	}

	// Pass through budget_tokens for extended thinking
	// Config: "ai_service": {"budget_tokens": 10000}
	if budgetTokens, ok := aiServiceConfig["budget_tokens"].(float64); ok && budgetTokens > 0 {
		options["budget_tokens"] = int(budgetTokens)
	}

	// Resolve model alias to actual API model name
	var modelAlias string // preserve original alias for logging
	var resolvedModel string
	if model, ok := options["model"].(string); ok {
		modelAlias = model
		resolvedModel = aiservice.ResolveModelAlias(model, params.Logger)
		options["model"] = resolvedModel
	}

	// Track provider for logging
	provider, _ := aiServiceConfig["provider"].(string)

	// Start timing the LLM call
	llmCallStart := time.Now()

	// Call the AI service
	result, err := aiClient.GenerateText(ctx, renderedPrompt, options)
	if err != nil {
		llmLatencyMs := int(time.Since(llmCallStart).Milliseconds())
		params.Logger.Info("AI call failed once",
			zap.Error(err),
		)

		// Log the failed call
		LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
			AgentType:       params.AgentType,
			AgentID:         params.Headers["agent_id"],
			StepName:        params.ExecutionContext.StepName,
			OrchestrationID: params.ExecutionContext.OrchestrationID,
			CorrelationID:   params.ExecutionContext.CorrelationID,
			Model:           modelAlias,
			ModelResolved:   resolvedModel,
			Provider:        provider,
			PromptTemplate:  promptTemplate,
			PromptRendered:  renderedPrompt,
			LatencyMs:       llmLatencyMs,
			Success:         false,
			ErrorMessage:    err.Error(),
		})

		errStr := err.Error()

		// Check for model-related errors
		if strings.Contains(errStr, "model") || strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
			modelUsed := fmt.Sprintf("%v", options["model"])

			params.Logger.Error("Model error - possibly invalid model name",
				zap.String("model_used", modelUsed),
				zap.Strings("available_aliases", aiservice.GetAvailableAliases()),
				zap.Error(err),
			)

			return nil, fmt.Errorf("model '%s' not found. Use aliases like: claude-sonnet-4-6, claude-opus-4-6, claude-haiku-4-5. Error: %w",
				modelUsed, err)
		}

		if strings.Contains(errStr, "529") || // overloaded
			strings.Contains(errStr, "503") || // service unavailable
			strings.Contains(errStr, "502") || // bad gateway
			strings.Contains(errStr, "500") { // internal server error

			retryDelays := []time.Duration{5 * time.Second, 15 * time.Second, 30 * time.Second, 60 * time.Second}
			var lastErr error
			for attempt, delay := range retryDelays {
				params.Logger.Warn("LLM call failed with transient error, retrying",
					zap.Int("attempt", attempt+1),
					zap.Int("max_attempts", len(retryDelays)),
					zap.Duration("backoff", delay),
					zap.String("error_type", errStr[:min(len(errStr), 80)]),
				)
				time.Sleep(delay)
				result, lastErr = aiClient.GenerateText(ctx, renderedPrompt, options)
				if lastErr == nil {
					params.Logger.Info("LLM retry succeeded",
						zap.Int("attempt", attempt+1))
					break
				}
			}
			if lastErr != nil {
				return nil, fmt.Errorf("AI call failed after %d retries. Last error: %w", len(retryDelays), lastErr)
			}
			err = nil // Clear the original error since retry succeeded
		}

		// Catch-all: any error not handled above (model check or transient retry)
		// is a hard failure. This catches 401 (auth), 402 (credit exhausted),
		// 408 (timeout), 429 (rate limited), and any other unexpected status codes.
		// Without this, the error is silently dropped and the pipeline continues
		// with an empty result — which can overwrite good page content with empty shells.
		if err != nil {
			params.Logger.Error("AI call failed with unhandled error — failing workflow",
				zap.Error(err),
			)
			return nil, fmt.Errorf("AI call failed with unhandled error: %w", err)
		}
	}

	params.Logger.Info("LLM response received",
		zap.String("result_preview", datahelpers.TruncateString(result, 200)))

	// Log the successful LLM call
	llmLatencyMs := int(time.Since(llmCallStart).Milliseconds())
	inputTokens := 0
	outputTokens := 0
	if it, ok := options["__usage_input_tokens"].(int); ok {
		inputTokens = it
	}
	if ot, ok := options["__usage_output_tokens"].(int); ok {
		outputTokens = ot
	}
	LogLLMCall(params.DB, params.Logger, LLMCallLogParams{
		AgentType:       params.AgentType,
		AgentID:         params.Headers["agent_id"],
		StepName:        params.ExecutionContext.StepName,
		OrchestrationID: params.ExecutionContext.OrchestrationID,
		CorrelationID:   params.ExecutionContext.CorrelationID,
		Model:           modelAlias,
		ModelResolved:   resolvedModel,
		Provider:        provider,
		PromptTemplate:  promptTemplate,
		PromptRendered:  renderedPrompt,
		ResponseText:    result,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		LatencyMs:       llmLatencyMs,
		Success:         true,
	})

	// Strip markdown code blocks from response before processing
	cleanedResult := stripMarkdownFromResponse(result)

	// Try to parse as JSON, if it fails return as plain text
	var parsedResult interface{}
	if err := json.Unmarshal([]byte(cleanedResult), &parsedResult); err != nil {
		// Not valid JSON, return as plain text (use cleaned result)
		return map[string]interface{}{
			"result": cleanedResult,
			"type":   "text",
		}, nil
	}

	// Valid JSON, return parsed result
	return map[string]interface{}{
		"result": parsedResult,
		"type":   "json",
	}, nil
}

// stripMarkdownFromResponse removes markdown code fences from LLM responses
// Handles ```json, ```html, ``` and similar patterns
func stripMarkdownFromResponse(s string) string {
	s = strings.TrimSpace(s)

	// Handle ```json, ```html, ```text, etc. at start
	if strings.HasPrefix(s, "```") {
		// Find end of first line (the language identifier line)
		newlineIdx := strings.Index(s, "\n")
		if newlineIdx > 0 {
			s = s[newlineIdx+1:] // Skip past ```json\n
		} else {
			s = strings.TrimPrefix(s, "```")
		}
		s = strings.TrimSpace(s)
	}

	// Remove trailing ```
	if strings.HasSuffix(s, "```") {
		// Find the last occurrence and remove it
		lastFence := strings.LastIndex(s, "```")
		if lastFence > 0 {
			s = s[:lastFence]
		} else {
			s = strings.TrimSuffix(s, "```")
		}
		s = strings.TrimSpace(s)
	}

	return s
}

// validateTemplateData checks if template data has fields referenced in the template
// Logs warnings for missing fields to help debug <no value> issues
func validateTemplateData(templateData map[string]interface{}, stepConfig map[string]interface{}, logger *zap.Logger) {
	// Get input_fields from config to know what we expected to find
	inputFields := []string{}
	if fields, ok := stepConfig["input_fields"].([]interface{}); ok {
		for _, f := range fields {
			if field, ok := f.(string); ok {
				inputFields = append(inputFields, field)
			}
		}
	}

	logger.Info("Template data validation",
		zap.Strings("expected_fields", inputFields),
		zap.Strings("available_fields", getTemplateDataKeys(templateData)),
	)

	// Check each expected field
	missingFields := []string{}
	for _, field := range inputFields {
		// Handle dot notation (e.g., "reviewed_brief.company_name")
		parts := strings.Split(field, ".")
		rootField := parts[0]

		if _, ok := templateData[rootField]; !ok {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		logger.Error("TEMPLATE DATA VALIDATION FAILED - Missing fields will render as <no value>",
			zap.Strings("missing_fields", missingFields),
			zap.Strings("available_fields", getTemplateDataKeys(templateData)),
		)
	}

	// Also check specific commonly-needed fields
	commonFields := []string{"reviewed_brief", "site_record", "input_data"}
	for _, field := range commonFields {
		if val, ok := templateData[field]; ok {
			if valMap, isMap := val.(map[string]interface{}); isMap {
				logger.Info("Template field contents",
					zap.String("field", field),
					zap.Strings("keys", getTemplateDataKeys(valMap)),
				)
			}
		}
	}
}

func getTemplateDataKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// appendOutputInstructions adds format-specific instructions based on output_type
// Checks both step config and ai_service config for output_type
func appendOutputInstructions(prompt string, aiConfig map[string]interface{}, stepConfig map[string]interface{}, logger *zap.Logger) string {
	// Check step config first (where it's typically defined in workflow)
	outputType := getOutputType(stepConfig)

	// Fallback to ai_service config
	if outputType == "" {
		outputType = getOutputType(aiConfig)
	}

	if outputType == "" {
		// No specific output type, use default clean output instructions
		return prompt + getDefaultOutputInstructions()
	}

	instructions := getOutputInstructions(outputType)
	if instructions == "" {
		return prompt
	}

	logger.Info("Appending output format instructions",
		zap.String("output_type", outputType))

	return prompt + instructions
}

// getOutputType extracts output_type from config
func getOutputType(config map[string]interface{}) string {
	if outputType, ok := config["output_type"].(string); ok {
		return outputType
	}
	return ""
}

// getOutputInstructions returns format-specific instructions
func getOutputInstructions(outputType string) string {
	switch outputType {
	case "json":
		return getJSONOutputInstructions()
	case "html":
		return getHTMLOutputInstructions()
	case "text":
		return getTextOutputInstructions()
	case "markdown":
		return getMarkdownOutputInstructions()
	default:
		return getDefaultOutputInstructions()
	}
}

// getJSONOutputInstructions returns JSON formatting instructions
func getJSONOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - JSON:
- Output ONLY the raw JSON object or array
- Do NOT wrap in markdown code fences (no ` + "```" + ` or ` + "```json" + `)
- Do NOT add explanatory text before or after the JSON
- Start your response with { or [ and end with } or ]
- Ensure valid JSON syntax (proper quotes, commas, brackets)

Example CORRECT:
{"site_type": "brochure", "recommended_builder": "multipage-website-builder"}

Example INCORRECT:
` + "```json\n{\"site_type\": \"brochure\"}\n```\n\nNote: This classification..."
}

// getHTMLOutputInstructions returns HTML formatting instructions
func getHTMLOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - HTML:
- Output ONLY the raw HTML code
- Do NOT wrap in markdown code fences (no ` + "```" + ` or ` + "```html" + `)
- Do NOT add explanatory text before or after the HTML
- Start with <!DOCTYPE html> or the appropriate opening tag
- Include complete, valid HTML structure

Example CORRECT:
<!DOCTYPE html>
<html lang="en">
<head>...</head>
<body>...</body>
</html>

Example INCORRECT:
` + "```html\n<!DOCTYPE html>...\n```\n\nHere's the HTML for your site..."
}

// getTextOutputInstructions returns text formatting instructions
func getTextOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - TEXT:
- Output ONLY the actual text content
- Do NOT wrap in markdown code fences
- Do NOT add meta-commentary like "Here's the content..." or "I've created..."
- Start directly with the content itself

Example CORRECT:
Welcome to our website. We provide excellent services...

Example INCORRECT:
Here's the content you requested:
` + "```\nWelcome to our website...\n```"
}

// getMarkdownOutputInstructions returns markdown formatting instructions
func getMarkdownOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT - MARKDOWN:
- Output ONLY the markdown content
- Do NOT wrap in code fences
- Do NOT add explanatory text before or after
- Use proper markdown syntax for headings, lists, links, etc.

Example CORRECT:
# Welcome

This is **bold** and this is *italic*.

Example INCORRECT:
` + "```markdown\n# Welcome\n```\n\nI've created markdown content for you..."
}

// getDefaultOutputInstructions returns general clean output instructions
func getDefaultOutputInstructions() string {
	return `

CRITICAL OUTPUT FORMAT:
- Output ONLY the requested content
- Do NOT wrap in code fences or markdown formatting
- Do NOT add preambles like "Here's what you asked for..."
- Do NOT add post-ambles like "Hope this helps!"
- Start directly with the actual content`
}

// helper function to load agent definition
func loadAgentDefinitionForAction(ctx context.Context, db interface{}, agentType string) (*AgentDefinition, error) {

	fmt.Printf("DEBUG: loadAgentDefinitionForAction called with agentType=%s, db type=%T\n", agentType, db)

	query := `
		SELECT id, type, display_name, description, category,
		       image_repository, image_tag, 
		       resources, default_config, capabilities, topics,
		       health_config, env_vars, is_active
		FROM agent_definitions
		WHERE type = $1 AND is_active = true
		ORDER BY version DESC
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
	rawProvider, exists := aiServiceConfig["provider"]
	if !exists || rawProvider == nil {
		return nil, fmt.Errorf("provider not specified in ai_service config - api_key_env_var included in config.ai_service?")
	}

	provider, ok := rawProvider.(string)
	if !ok || provider == "" {
		return nil, fmt.Errorf("provider must be a non-empty string in ai_service config")
	}
	switch provider {
	case "anthropic":
		return aiservice.NewAnthropicClient(ctx, aiServiceConfig)
	case "ollama":
		return aiservice.NewOllamaClient(ctx, aiServiceConfig)
	case "openai":
		return nil, fmt.Errorf("OpenAI provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s. Supported: anthropic, ollama", provider)
	}
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

// extractDataForAiAgent merges data from multiple sources specified in the step's 'input_fields' config.
func extractDataForAiAgent(params ActionParams) interface{} {
	params.Logger.Info("Extracting data for AI agent using UNIFIED EXTRACTOR",
		zap.Any("available_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Determine which fields to extract
	var inputFields []string
	if fields, ok := params.StepConfig.Config["input_fields"].([]interface{}); ok {
		for _, fieldInterface := range fields {
			if field, ok := fieldInterface.(string); ok {
				inputFields = append(inputFields, field)
			}
		}
	} else {
		params.Logger.Warn("No 'input_fields' found in config, defaulting to ['input_data']")
		inputFields = []string{"input_data"}
	}

	params.Logger.Info("Processing input_fields", zap.Strings("fields", inputFields))

	// USE THE UNIFIED EXTRACTOR
	templateData := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

	params.Logger.Info("Template data extracted",
		zap.Int("field_count", len(templateData)),
		zap.Strings("keys", datahelpers.GetMapKeys(templateData)),
	)

	return templateData
}

// getPromptWithPriority implements three-tier priority for prompt selection
func getPromptWithPriority(params ActionParams, agentConfig map[string]interface{}) (prompt string, source string) {
	logger := params.Logger

	logger.Info("in getPromptWithPriority")

	// PRIORITY 1: Check incoming message for prompt (from parent's call_agent)
	// Check in StepConfig.Config first (this is where call_agent passes it)
	if configPrompt, ok := params.StepConfig.Config["prompt"].(string); ok && configPrompt != "" {
		// Check if prompt contains template syntax and needs interpolation
		if strings.Contains(configPrompt, "{{") && strings.Contains(configPrompt, "}}") {
			logger.Info("Prompt contains template syntax, interpolating against CollectedData",
				zap.String("raw_prompt", configPrompt))

			interpolated, err := datahelpers.RenderPromptTemplate(configPrompt, params.CollectedData, *logger)
			if err != nil {
				logger.Warn("Failed to interpolate prompt template, using raw prompt",
					zap.Error(err),
					zap.String("raw_prompt", configPrompt))
				// Fall through to use raw prompt
			} else if interpolated != "" && interpolated != configPrompt {
				logger.Info("Prompt interpolated successfully",
					zap.String("interpolated_preview", datahelpers.TruncateString(interpolated, 200)))
				configPrompt = interpolated
			}
		}

		logger.Info("Using prompt from step config (Priority 1 - from parent)",
			zap.String("prompt_preview", datahelpers.TruncateString(configPrompt, 100)))
		return configPrompt, "parent_message"
	}

	// Also check in CollectedData["prompt"] as a fallback
	if collectedPrompt, ok := params.CollectedData["prompt"].(string); ok && collectedPrompt != "" {
		logger.Info("Using prompt from collected data (Priority 1 - from parent)",
			zap.String("prompt_preview", datahelpers.TruncateString(collectedPrompt, 100)))
		return collectedPrompt, "parent_message"
	}

	// PRIORITY 2: Check agent's own default_config.prompt_template
	// This comes from the agent_definitions table for this specific agent type
	if agentPrompt, ok := agentConfig["prompt_template"].(string); ok && agentPrompt != "" {
		logger.Info("Using prompt from agent config (Priority 2 - agent's default)",
			zap.String("agent_type", params.AgentType),
			zap.String("prompt_preview", datahelpers.TruncateString(agentPrompt, 100)))
		return agentPrompt, "agent_default"
	}

	// PRIORITY 3: Check workflow step config (fallback)
	// This is the hardcoded fallback in the workflow definition
	if stepConfig, ok := params.StepConfig.Config["prompt_template"].(string); ok && stepConfig != "" {
		logger.Info("Using prompt from workflow step config (Priority 3 - fallback)",
			zap.String("prompt_preview", datahelpers.TruncateString(stepConfig, 100)))
		return stepConfig, "workflow_fallback"
	}

	// Generic fallback if nothing found
	logger.Warn("No prompt found in any tier, using generic fallback")
	return "Generate content based on the provided context.", "generic_fallback"
}
