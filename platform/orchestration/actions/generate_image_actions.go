// internal/backend/agent-chassis/internal/actions/image_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ImageGenerationResult represents the result of an image generation request
type ImageGenerationResult struct {
	Success              bool   `json:"success"`
	ImageURI             string `json:"image_uri,omitempty"`
	ErrorMessage         string `json:"error_message,omitempty"`
	RequestID            string `json:"request_id"`
	ChildOrchestrationID string `json:"child_orchestration,omitempty"`
	ChildResponsesTopic  string `json:"child_responses_topic,omitempty"`
	TargetAgentType      string `json:"target_agent_type"`
	TopicSentTo          string `json:"topic_sent_to"`
	StableIdentity       string `json:"stable_identity"`
	AwaitResponse        bool   `json:"await_response"`
}

// ============================================================================
// Phase 2H — per-kind generation defaults
// ============================================================================

// imageDefaults captures per-kind generation parameters that callers may
// override via inputData. Zero values mean "no opinion; adapter applies
// its own fallback".
type imageDefaults struct {
	NegativePrompt string
	CfgScale       float64
	Steps          int
	Width          int
	Height         int
}

// kindDefaults keyed by the `kind` enum from site_plan_imagery.
// Tuned conservatively — biggest expected win is logo getting any
// negative prompt at all. Tune from observation.
var kindDefaults = map[string]imageDefaults{
	"logo": {
		NegativePrompt: "people, faces, text, signature, watermark, low quality, blurry, photorealistic, complex background",
		CfgScale:       8,
		Steps:          35,
		Width:          1024,
		Height:         1024,
	},
	"hero": {
		NegativePrompt: "text, watermark, signature, low quality, blurry, distorted",
		CfgScale:       7,
		Steps:          30,
		Width:          1024,
		Height:         1024,
	},
	"illustration": {
		NegativePrompt: "photorealistic, watermark, text, signature, low quality",
		CfgScale:       7,
		Steps:          30,
		Width:          1024,
		Height:         1024,
	},
	"icon": {
		NegativePrompt: "background, shadows, photorealistic, text, complex, busy",
		CfgScale:       7,
		Steps:          25,
		Width:          512,
		Height:         512,
	},
	"infographic": {
		NegativePrompt: "decorative, blurry, photorealistic, low quality",
		CfgScale:       8,
		Steps:          35,
		Width:          1024,
		Height:         1024,
	},
}

// resolveKind picks the effective kind from (in order of precedence):
//  1. inputData["kind"]         — caller-supplied via step input_mapping
//  2. inputData["default_kind"] — step-level fallback set by phase_2h.4
//     for legacy callers (call_logo_gen → "logo", call_hero_gen → "hero",
//     call_variant_gen → "hero")
//
// Returns "" when neither is set; imageDefaults[""] is the zero value
// meaning "no per-kind opinion".
func resolveKind(inputData map[string]interface{}, agentConfig map[string]interface{}) string {
	if k, ok := inputData["kind"].(string); ok && k != "" {
		return k
	}
	if k, ok := inputData["default_kind"].(string); ok && k != "" {
		return k
	}
	return ""
}

// parseAspectRatio takes a "W:H" string (e.g. "16:9", "1:1") plus
// anchor dimensions, returns (width, height, ok). Snaps to multiples
// of 64 per SDXL's requirement; keeps the longer side at the anchor.
func parseAspectRatio(ar string, anchorW, anchorH int) (int, int, bool) {
	parts := strings.Split(ar, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	w, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || w <= 0 || h <= 0 {
		return 0, 0, false
	}
	anchor := anchorW
	if anchorH > anchor {
		anchor = anchorH
	}
	var width, height int
	if w >= h {
		width = anchor
		height = anchor * h / w
	} else {
		height = anchor
		width = anchor * w / h
	}
	width = (width / 64) * 64
	height = (height / 64) * 64
	if width < 256 || height < 256 {
		return 0, 0, false
	}
	return width, height, true
}

// GenerateImageAction handles image generation requests with dynamic topics
// It should send to the adapter and wait for response
func GenerateImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing GenerateImageAction action",
		zap.String("agent_type", params.AgentType),
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		zap.String("action", params.ExecutionContext.Action),
		zap.Bool("has_db", params.DB != nil),
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("correlation_id", params.ExecutionContext.CorrelationID),
		zap.Any("DEBUGaa: full params in GenerateImageAction", params),
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

	// removed NormalizeCollectedData was destroying accumulated state
	// Only ensure essential topic fields if missing:
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

		agentDef, err := loadAgentDefinitionForImageAction(ctx, params.DB, params.AgentType)
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

	// Extract input data
	inputData := datahelpers.GetInputData(params.CollectedData, params.Logger)

	// Optional parameters — existing
	style, _ := inputData["style"].(string)
	width, _ := inputData["width"].(int)
	height, _ := inputData["height"].(int)

	// Phase 2H — kind resolution and per-kind defaults
	kind := resolveKind(inputData, agentConfig)
	defaults := kindDefaults[kind] // zero value if kind unknown

	// Width / height: caller wins, then per-kind default, then 1024 fallback
	if width == 0 {
		width = defaults.Width
	}
	if width == 0 {
		width = 1024
	}
	if height == 0 {
		height = defaults.Height
	}
	if height == 0 {
		height = 1024
	}

	// Phase 2H — negative_prompt: caller wins; otherwise per-kind default
	negativePrompt := defaults.NegativePrompt
	if np, ok := inputData["negative_prompt"].(string); ok && np != "" {
		negativePrompt = np
	}

	// Phase 2H — fold constraints from imagery spec into negative_prompt.
	// constraints arrives as map[string]interface{} from input_data.spec.constraints.
	// Known keys: no_text, no_logos, no_people, no_humans. Unknown keys ignored.
	if c, ok := inputData["constraints"].(map[string]interface{}); ok {
		var addNegatives []string
		if v, ok := c["no_text"].(bool); ok && v {
			addNegatives = append(addNegatives, "text", "letters", "writing")
		}
		if v, ok := c["no_logos"].(bool); ok && v {
			addNegatives = append(addNegatives, "logo", "brand mark")
		}
		if v, ok := c["no_people"].(bool); ok && v {
			addNegatives = append(addNegatives, "people", "faces", "humans")
		} else if v, ok := c["no_humans"].(bool); ok && v {
			addNegatives = append(addNegatives, "people", "faces", "humans")
		}
		if len(addNegatives) > 0 {
			extra := strings.Join(addNegatives, ", ")
			if negativePrompt != "" {
				negativePrompt += ", " + extra
			} else {
				negativePrompt = extra
			}
		}
	}

	// Phase 2H — style_hints.aspect_ratio overrides width/height if supplied.
	// Format: "16:9", "1:1", "4:3", etc.
	if h, ok := inputData["style_hints"].(map[string]interface{}); ok {
		if ar, ok := h["aspect_ratio"].(string); ok && ar != "" {
			if w2, h2, ok := parseAspectRatio(ar, width, height); ok {
				width, height = w2, h2
			} else {
				params.Logger.Warn("generate_image: aspect_ratio not understood; keeping default",
					zap.String("aspect_ratio", ar))
			}
		}
	}

	// Phase 2H — cfg_scale and steps. Caller wins; per-kind default otherwise.
	cfgScale := defaults.CfgScale
	if cs, ok := inputData["cfg_scale"].(float64); ok && cs > 0 {
		cfgScale = cs
	}
	steps := defaults.Steps
	if s, ok := inputData["steps"].(int); ok && s > 0 {
		steps = s
	}

	// Phase 2H — seed. 0 = random/unspecified.
	seed, _ := inputData["seed"].(int)

	// Phase 2H — reference_image_uri: accepted but not forwarded in v1.
	// Phase 3.1 will switch the adapter to /image-to-image when set.
	if refURI, ok := inputData["reference_image_uri"].(string); ok && refURI != "" {
		params.Logger.Info("generate_image: reference_image_uri supplied; not wired in v1 (Phase 3.1)",
			zap.String("reference_image_uri", refURI),
			zap.String("kind", kind))
	}

	// Extract prompt template
	// Get prompt using three-tier priority system
	promptTemplate, promptSource := getImagePromptWithPriority(params, agentConfig)

	// Phase 0.1: prepend design_intent.imagery_direction (read from site_specs)
	// when site_id is available via input_mapping. Direction is enrichment, not
	// a requirement — if site_id is missing or no design_intent exists, the
	// prompt is used as-is. Parents pass site_id through input_mapping; see
	// migration phase_0_combined_migration.sql.
	if siteID, _ := inputData["site_id"].(string); siteID != "" {
		direction := getImageryDirectionForSite(ctx, params.DB, siteID, params.Logger)
		if direction != "" {
			composed := composeImagePromptWithDirection(promptTemplate, direction)
			params.Logger.Info("Prepended imagery_direction from design_intent",
				zap.String("site_id", siteID),
				zap.String("direction_preview", datahelpers.TruncateString(direction, 100)),
				zap.String("composed_preview", datahelpers.TruncateString(composed, 250)))
			promptTemplate = composed
			promptSource = promptSource + "+imagery_direction"
		}
	}

	params.Logger.Info("Selected prompt for execution",
		zap.String("source", promptSource),
		zap.String("agent_type", params.AgentType),
		zap.String("kind", kind),
		zap.String("prompt_preview", datahelpers.TruncateString(promptTemplate, 350)))

	// When running within the image-generator agent, we send to the adapter topic
	// The adapter will respond back to us (the image-generator agent)
	imageGeneratorAdapterTopic := "system.adapter.image-generator.requests"

	// Get our own responses topic - where the adapter should respond to
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		// Fallback to environment variable if not in context
		myResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	if myResponsesTopic == "" {
		// Build it from the context
		myResponsesTopic = fmt.Sprintf("job.%s-%s-%s-%s.responses",
			extractShortCorrelation(params.ExecutionContext.CorrelationID),
			extractShortOrchestration(params.ExecutionContext.OrchestrationID),
			"image-generator",
			params.ExecutionContext.StepName,
		)
	}

	params.Logger.Info("Image-generator agent sending to adapter",
		zap.String("adapter_topic", imageGeneratorAdapterTopic),
		zap.String("my_responses_topic", myResponsesTopic),
		zap.String("DEBUGaa: DEBUGbb: which one: prompt template", promptTemplate),
		zap.String("prompt source", promptSource),
	)

	// Phase 2H — extended imageData. Base fields preserved; new ones omitted
	// when empty/zero so adapter sees identical JSON for legacy callers that
	// don't pass them.
	imageData := map[string]interface{}{
		"prompt": promptTemplate,
		"style":  style,
		"width":  width,
		"height": height,
	}
	if negativePrompt != "" {
		imageData["negative_prompt"] = negativePrompt
	}
	if cfgScale > 0 {
		imageData["cfg_scale"] = cfgScale
	}
	if steps > 0 {
		imageData["steps"] = steps
	}
	if seed > 0 {
		imageData["seed"] = seed
	}

	newRequestID := uuid.NewString()

	// The adapter expects certain fields in a specific format
	// Build the image generation request for the adapter
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               params.ExecutionContext.ClientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",

			// flat sender fields (not nested!)
			"sender_agent_type":    params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":      params.ExecutionContext.OrchestrationID,
			"sender_pod_name":      params.ExecutionContext.Sender.PodName,
			"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,
			"sender_role":          params.ExecutionContext.Sender.Role,

			// Response routing
			"responses_topic":        myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
		},
		"body": map[string]interface{}{
			"action": "generate",
			"data":   imageData,
			// Tell adapter where to reply
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
			"metadata": map[string]interface{}{
				"requesting_agent_id":   params.ExecutionContext.OrchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
			},
		},
	}

	// produce map[string]string headers for validation
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string)

	for k, v := range rawHeaders {
		if str, ok := v.(string); ok {
			headers[k] = str
		} else {
			headers[k] = fmt.Sprintf("%v", v) // fallback stringify
		}
	}

	// Convert to JSON
	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal adapter request: %w", err)
	}

	// Send to adapter
	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("Sending request to image adapter",
		zap.String("topic", imageGeneratorAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("reply_to_topic", myResponsesTopic),
		zap.Any("request", adapterRequest),
	)

	// Send the message
	if err := params.Producer.ProduceWithValidation(
		ctx,
		imageGeneratorAdapterTopic,
		headers,
		key,
		messageBytes,
	); err != nil {
		return nil, fmt.Errorf("failed to send to adapter: %w", err)
	}

	// Store tracking info in collected data
	if params.CollectedData != nil {
		params.CollectedData[params.ExecutionContext.StepName] = map[string]interface{}{
			"request_id":      newRequestID,
			"adapter_topic":   imageGeneratorAdapterTopic,
			"responses_topic": myResponsesTopic,
			"prompt":          promptTemplate,
			"status":          "awaiting_adapter_response",
			"width":           width,
			"height":          height,
		}
	}

	// The image-generator agent will receive the response from the adapter
	// on its responses topic and can then complete the workflow
	result := ImageGenerationResult{
		Success:             true,
		RequestID:           newRequestID,
		ChildResponsesTopic: myResponsesTopic,
		TargetAgentType:     "adapter",
		TopicSentTo:         imageGeneratorAdapterTopic,
		AwaitResponse:       true, // Agent needs to wait for adapter response
	}

	params.Logger.Info("Image generation request sent to adapter, awaiting response",
		zap.String("request_id", newRequestID),
		zap.String("expecting_response_on", myResponsesTopic),
	)

	return result, nil
}

// helper function to load agent definition
func loadAgentDefinitionForImageAction(ctx context.Context, db interface{}, agentType string) (*AgentDefinition, error) {

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

// ProcessImageResponse processes the response from the image adapter
// This is called when the adapter sends back the result
func ProcessImageResponse(params *ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "process_image_response"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Get the response from collected data
	// The response from the adapter should be stored here
	stepData, ok := datahelpers.GetStepData(params.CollectedData, params.ExecutionContext.StepName, logger)
	if !ok {
		// Check if response came in a different way
		if responseData, ok := params.CollectedData["adapter_response"].(map[string]interface{}); ok {
			stepData = responseData
		} else {
			return nil, fmt.Errorf("no response data found from adapter")
		}
	}

	// Extract image  URI  and  URL  from response
	imageURI, _ := stepData["image_uri"].(string)
	imageURL, _ := stepData["image_url"].(string)
	if imageURI == "" {
		// Check in body.data structure
		if body, ok := stepData["body"].(map[string]interface{}); ok {
			if data, ok := body["data"].(map[string]interface{}); ok {
				imageURI, _ = data["image_uri"].(string)
				imageURL, _ = data["image_url"].(string)
			}
		}
	}

	if imageURI == "" {
		// Check for error
		if errorMsg, ok := stepData["error"].(string); ok {
			return nil, fmt.Errorf("image generation failed: %s", errorMsg)
		}
		return nil, fmt.Errorf("no image URI in adapter response")
	}

	// Store result for parent agent with both S3 URI and presigned URL
	result := map[string]interface{}{
		"success":   true,
		"image_uri": imageURI, // S3 URI for storage reference
		"image_url": imageURL, // Presigned URL for web use (valid for 7 days)
	}

	// Update collected data so parent can access the result
	if params.CollectedData != nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			inputData["generated_image_uri"] = imageURI
			inputData["generated_image_url"] = imageURL
			inputData["image_generation_complete"] = true
		}
		params.CollectedData["image_result"] = result
	}

	logger.Info("Image generation complete",
		zap.String("image_uri", imageURI),
		zap.String("image_url", imageURL),
	)

	return result, nil
}

// Helper functions
func extractShortCorrelation(correlationID string) string {
	if len(correlationID) >= 8 {
		return correlationID[:8]
	}
	return correlationID
}

func extractShortOrchestration(orchestrationID string) string {
	if len(orchestrationID) >= 8 {
		return orchestrationID[:8]
	}
	return orchestrationID
}

func createImageGenerationTopics(ctx context.Context, requestsTopic, responsesTopic string, logger *zap.Logger) error {
	// Topic creation logic if needed
	// This might be handled by the platform automatically
	logger.Info("Topics for image generation",
		zap.String("requests", requestsTopic),
		zap.String("responses", responsesTopic),
	)
	return nil
}

// getPromptWithPriority implements three-tier priority for prompt selection
func getImagePromptWithPriority(params ActionParams, agentConfig map[string]interface{}) (prompt string, source string) {
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

	// Check input_data.prompt (from parent's input_mapping resolution)
	// When parent uses input_mapping: {"prompt": "site_plan.response.image_prompts.logo"},
	// the resolved value lands in CollectedData["input_data"]["prompt"]
	inputData := datahelpers.GetInputData(params.CollectedData, logger)
	if inputPrompt, ok := inputData["prompt"].(string); ok && inputPrompt != "" {
		logger.Info("Using prompt from input_data.prompt (Priority 1 - from parent input_mapping)",
			zap.String("prompt_preview", datahelpers.TruncateString(inputPrompt, 100)))
		return inputPrompt, "input_data"
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

// Add to platform/orchestration/actions/generate_image_actions.go
// This action stores the image URI/URL after call_agent returns from image-generator

// StoreGeneratedImageAction stores the S3 URI from image generation response
// and sets the relative URL for templates
func StoreGeneratedImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "store_generated_image"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("Storing generated image reference",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))

	// Get configuration
	purpose := "hero_home"
	if cfg, ok := params.StepConfig.Config["purpose"].(string); ok && cfg != "" {
		purpose = cfg
	}

	// Extract site_record to get site_id
	siteRecord := datahelpers.ExtractNestedFieldMap(params.CollectedData, "site_record")
	if siteRecord == nil {
		logger.Error("No site_record found")
		return map[string]interface{}{
			"success": false,
			"error":   "site_record not found",
		}, nil
	}
	siteID, _ := siteRecord["site_id"].(string)
	if siteID == "" {
		siteID, _ = siteRecord["id"].(string)
	}

	// Extract image URI from the call_agent response
	var imageURI string

	// Check possible locations for the image result
	possiblePaths := []string{"hero_image_result", "image_result", "generate_hero_image"}

	for _, path := range possiblePaths {
		if result, ok := params.CollectedData[path].(map[string]interface{}); ok {
			if uri, ok := result["image_uri"].(string); ok && uri != "" {
				imageURI = uri
				break
			}
			if response, ok := result["response"].(map[string]interface{}); ok {
				if uri, ok := response["image_uri"].(string); ok && uri != "" {
					imageURI = uri
					break
				}
			}
		}
	}

	if imageURI == "" {
		logger.Warn("No image URI found in result")
		return map[string]interface{}{
			"success": false,
			"error":   "No image URI in response",
			"skipped": true,
		}, nil
	}

	// Relative URL for templates - assemble_site will create this file
	relativeURL := fmt.Sprintf("/assets/images/%s.jpg", purpose)

	logger.Info("Storing image",
		zap.String("site_id", siteID),
		zap.String("purpose", purpose),
		zap.String("image_uri", imageURI),
		zap.String("relative_url", relativeURL))

	// Store in database
	if params.DB != nil {
		// Store S3 URI for assemble_site to fetch
		uriField := purpose + "_uri"
		if err := updateSiteContentField(ctx, params.DB, siteID, uriField, imageURI); err != nil {
			logger.Error("Failed to store URI", zap.Error(err))
		}

		// Store relative URL for templates
		urlField := purpose + "_url"
		if err := updateSiteContentField(ctx, params.DB, siteID, urlField, relativeURL); err != nil {
			logger.Error("Failed to store URL", zap.Error(err))
			return map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Failed to update content_data: %v", err),
			}, nil
		}
	}

	// Update collected_data for downstream steps
	if params.CollectedData != nil {
		params.CollectedData[purpose+"_uri"] = imageURI
		params.CollectedData[purpose+"_url"] = relativeURL

		if siteRecord != nil {
			if contentData, ok := siteRecord["content_data"].(map[string]interface{}); ok {
				contentData[purpose+"_uri"] = imageURI
				contentData[purpose+"_url"] = relativeURL
			}
		}
	}

	logger.Info("Image reference stored",
		zap.String("purpose", purpose),
		zap.String("s3_uri", imageURI),
		zap.String("relative_url", relativeURL))

	return map[string]interface{}{
		"success":   true,
		"site_id":   siteID,
		"purpose":   purpose,
		"image_uri": imageURI,
		"image_url": relativeURL,
	}, nil
}

// Helper: Update a field in sites.content_data
func updateSiteContentField(ctx context.Context, db interface{}, siteID, field, value string) error {
	query := `
		UPDATE sites 
		SET content_data = jsonb_set(
			COALESCE(content_data, '{}'::jsonb),
			$2::text[],
			to_jsonb($3::text),
			true
		),
		updated_at = NOW()
		WHERE id = $1
	`
	jsonPath := fmt.Sprintf("{%s}", field)

	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, query, siteID, jsonPath, value)
		return err
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, query, siteID, jsonPath, value)
		return err
	}
	return fmt.Errorf("unsupported database type")
}

// ============================================================================
// Phase 0.1 — imagery_direction composition helpers
//
// Today this captures only the strategic-spec layer (site_specs.design_intent).
// Once Phase 1 of doc 030 ships (site_plan_directives + brief renderer), this
// helper should also compose plan-time directives where category=design AND
// subject LIKE 'imagery%'. Tracked in PLAN_imagery_loop_closure.md as a
// successor work item.
// ============================================================================

// maxImageryDirectionInPrompt is the soft cap for the style-direction half of
// a composed image prompt. Set to fit comfortably under SDXL's 75-77 CLIP
// token limit (~350-380 chars total) while leaving 100-150 chars for the
// subject prompt.
//
// Provider-specific notes (relevant when provider routing lands):
//   - SDXL / Stable Diffusion family: hard wall at 77 tokens; this cap is correct
//   - FLUX (T5 encoder): ~256 tokens; cap could be raised to ~600 chars
//   - Imagen / Banana Pro: ~1000+ char effective; cap could be raised significantly
//   - Anthropic vision (audit-only): no meaningful cap; cap should not apply
//
// Until provider routing lands, 200 chars is the safe default for the only
// generation backend (Stability hosted SDXL).
const maxImageryDirectionInPrompt = 200

// getImageryDirectionForSite reads design_intent.imagery_direction from
// site_specs. Returns empty string if not found or on error — never blocks
// image generation.
//
// Handles both *sql.DB and *pgxpool.Pool to match the action's database
// abstraction (params.DB is interface{}).
func getImageryDirectionForSite(ctx context.Context, db interface{}, siteID string, logger *zap.Logger) string {
	if siteID == "" || db == nil {
		return ""
	}
	const query = `
		SELECT data->>'imagery_direction'
		FROM site_specs
		WHERE site_id = $1
		  AND aspect = 'design_intent'
		  AND is_current = true
		LIMIT 1
	`
	var direction sql.NullString
	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, siteID).Scan(&direction)
		if err != nil {
			if err != sql.ErrNoRows {
				logger.Warn("getImageryDirectionForSite: query failed",
					zap.String("site_id", siteID),
					zap.Error(err))
			}
			return ""
		}
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, siteID).Scan(&direction)
		if err != nil {
			// pgx returns its own no-rows sentinel; we treat any error as miss
			// without warning when the message indicates no rows — keeps logs
			// clean for sites without design_intent yet.
			if !strings.Contains(err.Error(), "no rows") {
				logger.Warn("getImageryDirectionForSite: query failed",
					zap.String("site_id", siteID),
					zap.Error(err))
			}
			return ""
		}
	default:
		logger.Warn("getImageryDirectionForSite: unsupported database type",
			zap.String("site_id", siteID))
		return ""
	}
	if !direction.Valid {
		return ""
	}
	return strings.TrimSpace(direction.String)
}

// composeImagePromptWithDirection prepends a design direction to the subject
// prompt. The direction is truncated to maxImageryDirectionInPrompt with a
// preference for ending at a sentence or comma boundary so the truncation
// reads naturally. The full subject prompt is preserved — never truncated —
// because losing the subject is unrecoverable while losing trailing style
// modifiers is graceful.
//
// Format is intentionally label-less. SDXL doesn't reason about "Style:" /
// "Subject:" meta-instructions and they cost ~17 tokens of the 77-token
// budget for no model benefit.
func composeImagePromptWithDirection(prompt, direction string) string {
	prompt = strings.TrimSpace(prompt)
	direction = strings.TrimSpace(direction)
	if prompt == "" || direction == "" {
		return prompt
	}
	if len(direction) > maxImageryDirectionInPrompt {
		cut := direction[:maxImageryDirectionInPrompt]
		// Prefer first sentence boundary within the cap; fall back to comma;
		// fall back to hard cut. The minKeep guard avoids degenerate cases
		// where a sentence boundary near the start would lose most of the
		// direction.
		minKeep := maxImageryDirectionInPrompt / 2
		if idx := strings.Index(cut, ". "); idx > minKeep {
			cut = cut[:idx+1]
		} else if idx := strings.LastIndex(cut, ", "); idx > minKeep {
			cut = cut[:idx]
		}
		direction = strings.TrimSpace(cut)
	}
	// Use period separator if direction doesn't already terminate with one.
	sep := ". "
	if endsWithSentenceBoundary(direction) {
		sep = " "
	}
	return direction + sep + prompt
}

// endsWithSentenceBoundary reports whether s ends with ., !, or ?.
func endsWithSentenceBoundary(s string) bool {
	if s == "" {
		return false
	}
	last := s[len(s)-1]
	return last == '.' || last == '!' || last == '?'
}
