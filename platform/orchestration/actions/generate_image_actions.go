// internal/backend/agent-chassis/internal/actions/image_actions.go
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

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

	// Check if we have a SagaCoordinator - not used
	/*if params.SagaCoordinator == nil {
		params.Logger.Error("SagaCoordinator is nil - cannot register awaited request")
		return nil, fmt.Errorf("SagaCoordinator is required for async image generation")
	}*/

	// Normalize the collected data using the helper
	/*normalizedData := datahelpers.NormalizeCollectedData(
		params.CollectedData,
		params.ExecutionContext,
		params.ExecutionContext.RequestsTopic,
		params.Logger,
	)*/

	// Update params.CollectedData with normalized version
	//params.CollectedData = normalizedData

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

	// Optional parameters
	style, _ := inputData["style"].(string)
	width, _ := inputData["width"].(int)
	height, _ := inputData["height"].(int)

	// Set defaults if not provided
	if width == 0 {
		width = 1024
	}
	if height == 0 {
		height = 1024
	}

	// Extract prompt template
	// Get prompt using three-tier priority system
	promptTemplate, promptSource := getImagePromptWithPriority(params, agentConfig)

	params.Logger.Info("Selected prompt for execution",
		zap.String("source", promptSource),
		zap.String("agent_type", params.AgentType),
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

	// Build the image generation request for the adapter
	imageData := map[string]interface{}{
		"prompt": promptTemplate,
		"style":  style,
		"width":  width,
		"height": height,
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

// sendImageGenerationRequest is no longer needed as we use producer directly
//
/*// internal/backend/agent-chassis/internal/actions/image_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
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

// GenerateImageAction handles image generation requests with dynamic topics
func GenerateImageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	params.Logger.Info("in GenerateImageAction",
		zap.Any("DEBUGaa: input action params", params),
	)

	logger := params.Logger.With(
		zap.String("action", "generate_image"),
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("correlation_id", params.ExecutionContext.CorrelationID),
	)

	// Extract input data
	inputData := datahelpers.GetInputData(params.CollectedData, logger)

	// Get the prompt from CollectedData top level
	prompt, ok := params.CollectedData["prompt"].(string)
	if !ok || prompt == "" {
		// or from input data
		prompt, ok = inputData["prompt"].(string)
	}

	if !ok || prompt == "" {
		return nil, fmt.Errorf("prompt is required for image generation")
	}

	// Optional parameters
	style, _ := inputData["style"].(string)
	width, _ := inputData["width"].(int)
	height, _ := inputData["height"].(int)

	// Set defaults if not provided
	if width == 0 {
		width = 1024
	}
	if height == 0 {
		height = 1024
	}

	// Create stable identity for this image generation request
	// This ensures unique topics for each image generation
	stableIdentity := fmt.Sprintf("%s-%s-image-generator-%s",
		extractShortCorrelation(params.ExecutionContext.CorrelationID),
		extractShortOrchestration(params.ExecutionContext.OrchestrationID),
		params.ExecutionContext.StepName,
	)

	// Create dynamic topics for this specific image generation
	requestsTopic := fmt.Sprintf("job.%s.requests", stableIdentity)
	responsesTopic := fmt.Sprintf("job.%s.responses", stableIdentity)

	// Get parent responses topic - where we should get the response
	parentResponsesTopic := params.ExecutionContext.ResponsesTopic
	if parentResponsesTopic == "" {
		parentResponsesTopic = params.ExecutionContext.ReplyToTopic
	}

	logger.Info("Creating image generation topics",
		zap.String("stable_identity", stableIdentity),
		zap.String("requests_topic", requestsTopic),
		zap.String("responses_topic", responsesTopic),
		zap.String("parent_responses_topic", parentResponsesTopic),
	)

	// Create topics using the topic manager
	if err := createImageGenerationTopics(params.Context, requestsTopic, responsesTopic, params.Logger); err != nil {
		logger.Error("Failed to create image generation topics", zap.Error(err))
		return nil, fmt.Errorf("failed to create topics: %w", err)
	}

	// Build the image generation request using the new data helpers
	// This ensures all required fields are properly populated
	imageData := map[string]interface{}{
		"prompt": prompt,
		"style":  style,
		"width":  width,
		"height": height,
	}

	// Use BuildRequestMessage from data_helpers to construct a properly formatted message
	imageRequest := datahelpers.BuildRequestMessage(
		params.ExecutionContext,
		"image-generator", // target agent type
		"generate",        // action
		imageData,         // data
		map[string]interface{}{ // config
			"stable_identity":        stableIdentity,
			"parent_responses_topic": parentResponsesTopic,
		},
		logger,
	)

	// Override the response topics to use our custom ones
	imageRequest.Headers.RequestsTopic = requestsTopic
	imageRequest.Headers.ResponsesTopic = responsesTopic
	imageRequest.Headers.ParentResponsesTopic = parentResponsesTopic
	imageRequest.Headers.Sender = params.ExecutionContext.Sender
	imageRequest.Headers.StepID = params.ExecutionContext.StepID
	imageRequest.Headers.StepName = params.ExecutionContext.StepName

	// Add image generation specific metadata to body
	if body, ok := imageRequest.Body.(map[string]interface{}); ok {
		body["reply_to_topic"] = parentResponsesTopic
		body["parent_responses_topic"] = parentResponsesTopic
		body["stable_identity"] = stableIdentity
		body["image_request"] = true
		body["metadata"] = map[string]interface{}{
			"stable_identity":       stableIdentity,
			"parent_orchestration":  params.ExecutionContext.ParentOrchestrationID,
			"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
			"requesting_step":       params.ExecutionContext.StepName,
		}
	}

	// Determine which image generator topic to use
	// We use a stable topic that multiple image generator containers listen to
	imageGeneratorTopic := "system.adapter.image-generator.requests"

	// Send the request
	requestID := uuid.NewString()
	if err := sendImageGenerationRequest(
		params.Context,
		params.Producer,
		imageGeneratorTopic,
		requestID,
		imageRequest,
		logger,
	); err != nil {
		logger.Error("Failed to send image generation request", zap.Error(err))
		return nil, err
	}

	logger.Info("Image generation request sent successfully",
		zap.String("request_id", requestID),
		zap.String("topic", imageGeneratorTopic),
		zap.String("stable_identity", stableIdentity),
		zap.Any("DEBUGaa: image request sent", imageRequest),
	)

	// Store request tracking info in collected data
	if params.CollectedData != nil {
		params.CollectedData[params.ExecutionContext.StepName] = map[string]interface{}{
			"request_id":            requestID,
			"stable_identity":       stableIdentity,
			"requests_topic":        requestsTopic,
			"responses_topic":       responsesTopic,
			"parent_topic":          parentResponsesTopic,
			"image_generator_topic": imageGeneratorTopic,
			"prompt":                prompt,
			"status":                "awaiting_response",
		}
	}

	// Return result with await flag
	result := ImageGenerationResult{
		Success:             true,
		RequestID:           requestID,
		ChildResponsesTopic: responsesTopic,
		TargetAgentType:     "image-generator",
		TopicSentTo:         imageGeneratorTopic,
		StableIdentity:      stableIdentity,
		AwaitResponse:       true, // We need to wait for the image generation
	}

	logger.Info("Image generation request sent",
		zap.String("request_id", requestID),
		zap.String("topic", imageGeneratorTopic),
		zap.Bool("await_response", true),
	)

	return result, nil
}

// sendImageGenerationRequest sends the request message with proper validation
func sendImageGenerationRequest(
	ctx context.Context,
	producer kafka.Producer,
	topic string,
	requestID string,
	message *types.RequestMessage,
	logger *zap.Logger,
) error {
	// Ensure the message has the request_id in headers
	if message.Headers.RequestID == "" {
		message.Headers.RequestID = requestID
	}

	// Ensure all required validation fields are present
	// The validator checks these specific fields
	if message.Headers.SenderAgentType == "" && message.Headers.Sender.AgentType != "" {
		message.Headers.SenderAgentType = message.Headers.Sender.AgentType
	}
	if message.Headers.InResponseToStepName == "" && message.Headers.StepName != "" {
		message.Headers.InResponseToStepName = message.Headers.StepName
	}

	logger.Info("Sending image generation request",
		zap.String("topic", topic),
		zap.String("request_id", requestID),
		zap.String("correlation_id", message.Headers.CorrelationID),
		zap.String("sender_agent_type", message.Headers.SenderAgentType),
		zap.String("to_agent_type", message.Headers.ToAgentType),
		zap.Any("DEBUGaa: message", message),
	)

	// Convert headers to map
	headers := message.Headers.ToMap()

	logger.Info("Sending image generation request - headers",
		zap.Any("headers", headers),
	)

	// Convert message to bytes
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send with validation
	key := []byte(message.Headers.CorrelationID)
	return producer.ProduceWithValidation(ctx, topic, headers, key, messageBytes)
}

// ProcessImageResponse processes responses from image generation
func ProcessImageResponse(params *ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "process_image_response"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Get the response from collected data
	stepData, ok := datahelpers.GetStepData(params.CollectedData, params.ExecutionContext.StepName, logger)
	if !ok {
		return nil, fmt.Errorf("no response data found for step %s", params.ExecutionContext.StepName)
	}

	// Extract image URI and URL from response
	imageURI, _ := stepData["image_uri"].(string)
	imageURL, _ := stepData["image_url"].(string)
	if imageURI == "" {
		// Check for error
		if errorMsg, ok := stepData["error"].(string); ok {
			return nil, fmt.Errorf("image generation failed: %s", errorMsg)
		}
		return nil, fmt.Errorf("no image URI in response")
	}

	// Store in collected data for use by other steps
	if params.CollectedData != nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			inputData["generated_image_uri"] = imageURI
			inputData["generated_image_url"] = imageURL
			inputData["image_generation_complete"] = true
		}
	}

	logger.Info("Image generation complete",
		zap.String("image_uri", imageURI),
		zap.String("image_url", imageURL),
	)

	return map[string]interface{}{
		"success":   true,
		"image_uri": imageURI,  // S3 URI for storage reference
		"image_url": imageURL,  // Presigned URL for web use (valid for 7 days)
	}, nil
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// createImageGenerationTopics creates the dynamic topics for image generation
func createImageGenerationTopics(ctx context.Context, requestsTopic, responsesTopic string, logger *zap.Logger) error {
	// Get Kafka brokers from environment or config
	brokers := strings.Split(getEnvOrDefault("KAFKA_BROKERS", "kafka:9092"), ",")

	// Create topic manager
	topicManager := kafka.NewTopicManager(brokers, logger)

	// Define topic configurations
	topics := []kafka.TopicDefinition{
		{
			Name:              requestsTopic,
			Partitions:        2,
			ReplicationFactor: 1,
		},
		{
			Name:              responsesTopic,
			Partitions:        2,
			ReplicationFactor: 1,
		},
	}

	// Create topics
	for _, topic := range topics {
		if err := topicManager.CreateTopic(ctx, topic); err != nil {
			// Check if topic already exists (not an error)
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create topic %s: %w", topic.Name, err)
			}
			logger.Debug("Topic already exists", zap.String("topic", topic.Name))
		} else {
			logger.Info("Created image generation topic",
				zap.String("topic", topic.Name),
				zap.Int("partitions", topic.Partitions),
			)
		}

		// Wait for topic to be ready
		if err := topicManager.WaitForTopic(ctx, topic.Name, logger); err != nil {
			logger.Warn("Topic may not be fully ready",
				zap.String("topic", topic.Name),
				zap.Error(err),
			)
		}
	}

	return nil
}
*/

// getPromptWithPriority implements three-tier priority for prompt selection
func getImagePromptWithPriority(params ActionParams, agentConfig map[string]interface{}) (prompt string, source string) {
	logger := params.Logger

	logger.Info("in getPromptWithPriority")

	// PRIORITY 1: Check incoming message for prompt (from parent's call_agent)
	// Check in StepConfig.Config first (this is where call_agent passes it)
	if configPrompt, ok := params.StepConfig.Config["prompt"].(string); ok && configPrompt != "" {
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
