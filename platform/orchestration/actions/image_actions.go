// internal/backend/agent-chassis/internal/actions/image_actions.go
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

	// Build the image generation request
	imageRequest := buildImageGenerationRequest(
		params.ExecutionContext,
		prompt,
		style,
		width,
		height,
		stableIdentity,
		responsesTopic,       // The image generator should respond to this topic
		parentResponsesTopic, // But ultimately the response goes to parent
		logger,
	)

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

	// Store request tracking info in collected data
	if params.CollectedData != nil {
		params.CollectedData[params.ExecutionContext.StepName] = map[string]interface{}{
			"request_id":      requestID,
			"stable_identity": stableIdentity,
			"requests_topic":  requestsTopic,
			"responses_topic": responsesTopic,
			"prompt":          prompt,
			"status":          "awaiting_response",
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

// CallImageGeneratorAction is an alternative action for calling an existing image generator agent
func CallImageGeneratorAction(ctx context.Context, params ActionParams) (interface{}, error) {
	if params.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	logger := params.Logger.With(
		zap.String("action", "call_image_generator"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Extract target agent info from collected data
	inputData := datahelpers.GetInputData(params.CollectedData, logger)

	// Check if we have a specific image generator agent to call
	imageGeneratorInfo, ok := params.CollectedData["image_generator_agent"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no image generator agent info found")
	}

	agentID, _ := imageGeneratorInfo["agent_id"].(string)
	requestsTopic, _ := imageGeneratorInfo["requests_topic"].(string)
	responsesTopic, _ := imageGeneratorInfo["responses_topic"].(string)

	if agentID == "" || requestsTopic == "" {
		return nil, fmt.Errorf("invalid image generator agent info")
	}

	// Build request message for the image generator
	childCtx := params.ExecutionContext.CreateChildContext("image-generator")
	childCtx.Action = "generate"

	requestMessage := datahelpers.BuildRequestMessage(
		childCtx,
		"image-generator",
		"generate",
		inputData,
		nil,
		logger,
	)

	// Send to the image generator's request topic
	requestID := uuid.NewString()
	if err := sendAgentImageRequest(
		params.Context,
		params.Producer,
		requestsTopic,
		requestID,
		requestMessage,
		logger,
	); err != nil {
		return nil, fmt.Errorf("failed to call image generator: %w", err)
	}

	// Return result
	result := map[string]interface{}{
		"success":               true,
		"request_id":            requestID,
		"agent_called":          agentID,
		"agent_type":            "image-generator",
		"child_responses_topic": responsesTopic,
		"await_response":        true,
		"action_sent":           "generate",
	}

	return result, nil
}

// buildImageGenerationRequest creates the request message for image generation
func buildImageGenerationRequest(
	execCtx *types.ExecutionContext,
	prompt string,
	style string,
	width int,
	height int,
	stableIdentity string,
	responsesTopic string,
	parentResponsesTopic string,
	logger *zap.Logger,
) *types.RequestMessage {

	// Create a child context for the image generation
	childCtx := execCtx.CreateChildContext("image-generator")
	childCtx.Action = "generate"
	childCtx.RequestsTopic = fmt.Sprintf("job.%s.requests", stableIdentity)
	childCtx.ResponsesTopic = responsesTopic
	childCtx.ReplyToTopic = parentResponsesTopic // Image generator should ultimately reply here

	// Build headers
	headers := childCtx.ToRequestHeaders()

	// Build body with image generation parameters
	body := map[string]interface{}{
		"action": "generate",
		"data": map[string]interface{}{
			"prompt": prompt,
			"style":  style,
			"width":  width,
			"height": height,
		},
		// Add specific image generation values
		"stable_identity": stableIdentity,
		"image_request":   true,
		"reply_to_topic":  parentResponsesTopic, // Explicitly tell where to reply
		"metadata": map[string]interface{}{
			"stable_identity":       stableIdentity,
			"parent_orchestration":  execCtx.ParentOrchestrationID,
			"requesting_agent_type": execCtx.Sender.AgentType,
			"requesting_step":       execCtx.StepName,
		},
	}

	logger.Debug("Built image generation request",
		zap.String("stable_identity", stableIdentity),
		zap.String("prompt", prompt),
		zap.String("reply_to_topic", parentResponsesTopic),
	)

	return &types.RequestMessage{
		Headers: headers,
		Body:    body,
	}
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

// sendImageGenerationRequest sends the request to the image generator
func sendImageGenerationRequest(
	ctx context.Context,
	producer kafka.Producer,
	topic string,
	requestID string,
	message *types.RequestMessage,
	logger *zap.Logger,
) error {

	// Serialize the message
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Convert headers to Kafka headers
	headers := message.Headers.ToMap()
	headers["request_id"] = requestID

	// Send the message
	key := []byte(message.Headers.CorrelationID)

	logger.Debug("Sending image generation request",
		zap.String("topic", topic),
		zap.String("request_id", requestID),
		zap.Int("message_size", len(messageBytes)),
	)

	return producer.ProduceWithValidation(ctx, topic, headers, key, messageBytes)
}

// sendAgentRequest sends a request to a specific agent
func sendAgentImageRequest(
	ctx context.Context,
	producer kafka.Producer,
	topic string,
	requestID string,
	message *types.RequestMessage,
	logger *zap.Logger,
) error {
	return sendImageGenerationRequest(ctx, producer, topic, requestID, message, logger)
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

	// Extract image URI from response
	imageURI, _ := stepData["image_uri"].(string)
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
			inputData["image_generation_complete"] = true
		}
	}

	logger.Info("Image generation complete",
		zap.String("image_uri", imageURI),
	)

	return map[string]interface{}{
		"success":   true,
		"image_uri": imageURI,
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
