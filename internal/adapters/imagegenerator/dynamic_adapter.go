// internal/adapters/imagegenerator/dynamic_adapter.go
// Dynamic Image Generator Adapter with proper topic management and S3 integration
package imagegenerator

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"github.com/gqls/agentchassis/platform/resilience"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// DynamicImageAdapter handles image generation with dynamic topic routing
type DynamicImageAdapter struct {
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *zap.Logger
	config        *config.ServiceConfig
	consumer      *kafka.Consumer
	producer      kafka.Producer
	storageClient storage.Client
	httpClient    *resilience.HTTPClientWithBreaker
	externalAPI   string
	apiKey        string
	adapterID     string // Unique ID for this adapter instance
	podName       string
}

// ImageRequestData is the inner data payload of an image generation request.
// Phase 2H added: NegativePrompt, CfgScale, Steps, Seed. All optional; legacy
// callers sending only {prompt, style, width, height} produce identical JSON.
type ImageRequestData struct {
	Prompt         string  `json:"prompt"`
	Style          string  `json:"style,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	CfgScale       float64 `json:"cfg_scale,omitempty"`
	Steps          int     `json:"steps,omitempty"`
	Seed           int     `json:"seed,omitempty"`
}

// ImageRequest represents an incoming image generation request
type ImageRequest struct {
	Headers types.RequestHeaders `json:"headers"`
	Body    struct {
		Action       string                 `json:"action"`
		Data         ImageRequestData       `json:"data"`
		ReplyToTopic string                 `json:"reply_to_topic"`
		Metadata     map[string]interface{} `json:"metadata,omitempty"`
	} `json:"body"`
}

// ImageResponse represents the response sent back
type ImageResponse struct {
	Headers types.ResponseHeaders `json:"headers"`
	Body    types.ResponseBody    `json:"body"`
}

// NewDynamicImageAdapter creates a new dynamic image generator adapter
func NewDynamicImageAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*DynamicImageAdapter, error) {
	adapterCtx, cancel := context.WithCancel(ctx)

	// Generate adapter ID
	adapterID := uuid.NewString()
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = fmt.Sprintf("image-adapter-%s", adapterID[:8])
	}

	// debug
	envVars := os.Environ()
	logger.Info("New Dynamic Image Adapter - environment variables, try to un-hardcode vars",
		zap.Strings("DEBUGaa: try to un-hardcode env vars", envVars),
	)

	// Subscribe to the main image generator request topic
	// All adapters in the consumer group listen to this topic
	mainTopic := "system.adapter.image-generator.requests"

	// Initialize Kafka consumer with consumer group
	// Each adapter instance joins the same consumer group for load balancing
	consumerGroup := "image-generator-adapter-group"
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, mainTopic, consumerGroup, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	/*	// Get S3 credentials from environment variables
		accessKey := os.Getenv(cfg.Infrastructure.ObjectStorage.AccessKeyEnvVar)
		if accessKey == "" && cfg.Infrastructure.ObjectStorage.AccessKeyEnvVar != "" {
			accessKey = os.Getenv("S3_ACCESS_KEY") // fallback
		}

		secretKey := os.Getenv(cfg.Infrastructure.ObjectStorage.SecretKeyEnvVar)
		if secretKey == "" && cfg.Infrastructure.ObjectStorage.SecretKeyEnvVar != "" {
			secretKey = os.Getenv("S3_SECRET_KEY") // fallback
		}*/

	// Initialize storage client (S3)
	storageConfig := cfg.Infrastructure.ObjectStorage
	storageConfig.Provider = cfg.Infrastructure.ObjectStorage.Provider
	storageConfig.Endpoint = cfg.Infrastructure.ObjectStorage.Endpoint
	storageConfig.Bucket = cfg.Infrastructure.ObjectStorage.Bucket
	storageConfig.AccessKeyEnvVar = cfg.Infrastructure.ObjectStorage.AccessKeyEnvVar
	storageConfig.SecretKeyEnvVar = cfg.Infrastructure.ObjectStorage.SecretKeyEnvVar

	zap.Strings("environment Environ", os.Environ())

	externalAPI := os.Getenv("STABILITY_API_ENDPOINT")
	apiKey := os.Getenv("STABILITY_API_KEY")

	if storageConfig.Endpoint == "" {
		storageConfig.Endpoint = os.Getenv("S3_ENDPOINT")
	}
	if storageConfig.Bucket == "" {
		storageConfig.Bucket = os.Getenv("IMAGE_BUCKET")
	}

	storageClient, err := storage.NewS3Client(ctx, storageConfig, *logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	// Create base HTTP client.
	// Timeout covers the whole request lifecycle: connect + send + server-side generation
	// + response body read. SDXL at 1024x1024 with 30 steps commonly takes 30-60s
	// server-side, plus ~1.7MB base64 response to transfer. 30s was too tight and caused
	// "context deadline exceeded while reading body" failures.
	baseHTTPClient := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Create circuit breaker config
	cbConfig := resilience.CircuitBreakerConfig{
		Name:                "image-generator-api",
		MaxRequests:         3,
		Interval:            60 * time.Second,
		Timeout:             60 * time.Second,
		ConsecutiveFailures: 5,
		FailureRatio:        0.6,
	}

	// Wrap with circuit breaker
	httpClient := resilience.NewHTTPClientWithBreaker(baseHTTPClient, cbConfig, logger)

	adapter := &DynamicImageAdapter{
		ctx:           adapterCtx,
		cancel:        cancel,
		logger:        logger,
		config:        cfg,
		consumer:      consumer,
		producer:      producer,
		storageClient: storageClient,
		httpClient:    httpClient,
		externalAPI:   externalAPI,
		apiKey:        apiKey,
		adapterID:     adapterID,
		podName:       podName,
	}

	logger.Info("Dynamic image adapter initialized",
		zap.String("adapter_id", adapterID),
		zap.String("pod_name", podName),
		zap.String("consumer_group", consumerGroup),
		zap.Any("storageClient", storageClient),
		zap.String("externalAPI", externalAPI),
		zap.String("apiKey", apiKey),
		zap.String("adapterID", adapterID),
	)

	return adapter, nil
}

// Run starts the adapter's main processing loop
func (a *DynamicImageAdapter) Run() error {

	a.logger.Info("Image adapter listening for requests",
		zap.String("topic (set above in NewDynamicImageAdapter)", "system.adapter.image-generator.requests"),
		zap.String("adapter_id", a.adapterID),
	)

	// Main processing loop
	for {
		select {
		case <-a.ctx.Done():
			return a.ctx.Err()
		default:
			// Fetch and process messages
			msg, err := a.consumer.FetchMessage(a.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}

			// Process message asynchronously
			go a.handleMessage(msg)
		}
	}
}

// handleMessage processes a single image generation request
func (a *DynamicImageAdapter) handleMessage(msg kafka.Message) {
	startTime := time.Now()

	// Parse the request
	var request ImageRequest
	if err := json.Unmarshal(msg.Value, &request); err != nil {
		a.logger.Error("Failed to unmarshal request", zap.Error(err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	logger := a.logger.With(
		zap.String("correlation_id", request.Headers.CorrelationID),
		zap.String("orchestration_id", request.Headers.OrchestrationID),
		zap.String("step_name", request.Headers.StepName),
		zap.String("request_id", request.Headers.RequestID),
	)

	a.logger.Info("Processing image generation request",
		zap.String("prompt", request.Body.Data.Prompt),
		zap.String("reply_to_topic", request.Body.ReplyToTopic),
		zap.Any("the request body", request.Body),
		zap.Any("the request headers", request.Headers),
	)

	// Determine where to send the response
	responseTopic := request.Body.ReplyToTopic
	if responseTopic == "" {
		// Fallback to parent responses topic from headers
		responseTopic = request.Headers.ParentResponsesTopic
	}
	if responseTopic == "" {
		logger.Error("No response topic specified")
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Generate image using external API
	// Phase 2H: pass the whole data struct so negative_prompt, cfg_scale,
	// steps, and seed flow through to Stability.
	imageData, err := a.generateImage(request.Body.Data)
	if err != nil {
		logger.Error("Failed to generate image", zap.Error(err))
		a.sendErrorResponse(responseTopic, &request, err, logger)
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	a.logger.Info("image about to be uploaded in handleMessage dynamic adapter") //zap.Binary("DEBUGaa: image binary", imageData),

	// Upload to S3 and get presigned URL
	imageURI, presignedURL, err := a.uploadImage(imageData, request.Headers.ClientID, logger)
	if err != nil {
		logger.Error("Failed to upload image", zap.Error(err))
		a.sendErrorResponse(responseTopic, &request, err, logger)
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Send success response with both S3 URI and presigned URL
	a.sendSuccessResponse(responseTopic, &request, imageURI, presignedURL, time.Since(startTime), logger)

	// Commit the message
	a.consumer.CommitMessages(context.Background(), msg)

	logger.Info("Image generation complete",
		zap.String("image_uri", imageURI),
		zap.String("presigned_url", presignedURL),
		zap.Duration("duration", time.Since(startTime)),
	)
}

// generateImage calls the external image generation API.
// Phase 2H: takes the full data struct so negative_prompt, cfg_scale,
// steps, and seed flow through to Stability rather than being hardcoded.
func (a *DynamicImageAdapter) generateImage(data ImageRequestData) ([]byte, error) {
	// Resolve dimensions with fallbacks
	width := data.Width
	if width == 0 {
		width = 1024
	}
	height := data.Height
	if height == 0 {
		height = 1024
	}

	// Stability v1 SDXL: negative prompts are entries in text_prompts
	// with negative weight, not a separate field.
	textPrompts := []map[string]interface{}{
		{"text": data.Prompt, "weight": 1.0},
	}
	if data.NegativePrompt != "" {
		textPrompts = append(textPrompts, map[string]interface{}{
			"text":   data.NegativePrompt,
			"weight": -1.0,
		})
	}

	// CFG scale and steps: caller wins; otherwise the prior Stability
	// defaults (7, 30) preserved exactly.
	cfgScale := data.CfgScale
	if cfgScale <= 0 {
		cfgScale = 7
	}
	steps := data.Steps
	if steps <= 0 {
		steps = 30
	}

	// Build request body
	requestBody := map[string]interface{}{
		"text_prompts":         textPrompts,
		"cfg_scale":            cfgScale,
		"clip_guidance_preset": "FAST_BLUE",
		"height":               height,
		"width":                width,
		"samples":              1,
		"steps":                steps,
	}

	// Seed: 0 = unspecified (Stability generates random). Only include
	// when caller set a value, otherwise Stability uses its default.
	if data.Seed > 0 {
		requestBody["seed"] = data.Seed
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	a.logger.Info("generateImage: stability request",
		zap.Int("width", width),
		zap.Int("height", height),
		zap.Float64("cfg_scale", cfgScale),
		zap.Int("steps", steps),
		zap.Int("seed", data.Seed),
		zap.Bool("has_negative_prompt", data.NegativePrompt != ""),
		zap.Int("prompt_len", len(data.Prompt)),
		zap.String("a.externalAPI", a.externalAPI),
	)

	// Create HTTP request
	req, err := http.NewRequestWithContext(a.ctx, "POST", a.externalAPI, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	req.Header.Set("Accept", "application/json")

	a.logger.Info("in generate Image about to execute request dynamic adapter",
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Int("body_length_bytes", len(jsonBody)),
	)
	// Execute request through circuit breaker
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		Artifacts []struct {
			Base64 string `json:"base64"`
		} `json:"artifacts"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	a.logger.Info("in generate Image after sending, here is the api response. dynamic adapter",
		zap.Any("apiResponse", "too big for the moment dynamic_adapter"),
	)

	if len(apiResponse.Artifacts) == 0 {
		return nil, fmt.Errorf("no images in response")
	}

	// Decode base64 image
	return base64.StdEncoding.DecodeString(apiResponse.Artifacts[0].Base64)
}

// uploadImage uploads the generated image to S3 and returns both URI and presigned URL
func (a *DynamicImageAdapter) uploadImage(imageData []byte, clientID string, logger *zap.Logger) (string, string, error) {

	// Generate unique filename
	timestamp := time.Now().Format("20060102-150405")
	imageID := uuid.NewString()
	fileName := fmt.Sprintf("images/%s/%s/%s.png", clientID, timestamp[:8], imageID)

	logger.Info("uploadImage about to upload image",
		zap.String("filename", fileName),
		zap.String("imageID uuid", imageID),
	)

	// Upload to S3
	imageURI, err := a.storageClient.Upload(
		a.ctx,
		fileName,
		"image/png",
		bytes.NewReader(imageData),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	logger.Info("Image uploaded to S3",
		zap.String("uri", imageURI),
		zap.String("key", fileName),
		zap.Int("size", len(imageData)),
	)

	// Generate presigned URL (valid for 7 days = 10080 minutes)
	presignedURL, err := a.storageClient.GetPresignedURL(a.ctx, fileName, 10080)
	if err != nil {
		logger.Warn("Failed to generate presigned URL, continuing with S3 URI only",
			zap.Error(err),
			zap.String("uri", imageURI),
		)
		// Return S3 URI even if presigned URL generation fails
		return imageURI, "", nil
	}

	logger.Info("Generated presigned URL for image",
		zap.String("uri", imageURI),
		zap.String("presigned_url", presignedURL),
		zap.String("expires_in", "7 days"),
	)

	return imageURI, presignedURL, nil
}

// sendSuccessResponse sends a successful response to the parent
func (a *DynamicImageAdapter) sendSuccessResponse(
	topic string,
	request *ImageRequest,
	imageURI string,
	presignedURL string,
	duration time.Duration,
	logger *zap.Logger,
) {
	// Build response headers
	responseHeaders := types.ResponseHeaders{
		Sender: types.AgentIdentity{
			AgentType:    "image-generator",
			AgentID:      a.adapterID,
			PodName:      a.podName,
			AgentVersion: "1.0",
		},
		InResponseToRequestID: request.Headers.RequestID,
		ReplyToRequestID:      request.Headers.ReplyToRequestID,
		InResponseToStepID:    request.Headers.StepID,
		InResponseToStepName:  request.Headers.StepName,
		OrchestrationID:       request.Headers.OrchestrationID,
		OrchestrationName:     request.Headers.OrchestrationName,
		MyOrchestrationID:     a.adapterID,
		MyOrchestrationName:   "orchestration name in dynamic adapter",
		TopicSentTo:           topic,
		CorrelationID:         request.Headers.CorrelationID,
		ClientID:              request.Headers.ClientID,
		MessageType:           "response",
		Status:                "complete",
		IsComplete:            true,
		TimeSent:              time.Now(),
		TimeSpent:             duration,
	}

	// Build response body with both S3 URI and presigned URL
	responseBody := types.ResponseBody{
		Success: true,
		Body: map[string]interface{}{
			"image_uri":       imageURI,     // S3 URI for storage/reference
			"image_url":       presignedURL, // Presigned URL for web use (7 days)
			"prompt":          request.Body.Data.Prompt,
			"generated_at":    time.Now().Format(time.RFC3339),
			"generation_time": duration.Seconds(),
			"adapter_id":      a.adapterID,
		},
		Error: nil,
	}

	response := &ImageResponse{
		Headers: responseHeaders,
		Body:    responseBody,
	}

	a.logger.Info("In sending success response here are response details",
		zap.Any("response", response),
	)

	// Send the response
	a.sendResponse(topic, response, logger)
}

// sendErrorResponse sends an error response
func (a *DynamicImageAdapter) sendErrorResponse(
	topic string,
	request *ImageRequest,
	err error,
	logger *zap.Logger,
) {
	responseHeaders := types.ResponseHeaders{
		Sender: types.AgentIdentity{
			AgentType:    "image-generator",
			AgentID:      a.adapterID,
			PodName:      a.podName,
			AgentVersion: "1.0",
		},
		InResponseToRequestID: request.Headers.RequestID,
		OrchestrationID:       request.Headers.OrchestrationID,
		MyOrchestrationID:     a.adapterID,
		TopicSentTo:           topic,
		CorrelationID:         request.Headers.CorrelationID,
		ClientID:              request.Headers.ClientID,
		MessageType:           "response",
		Status:                "error",
		IsComplete:            true,
		TimeSent:              time.Now(),
	}

	errorInfo := &types.ErrorInfo{
		Code:        "IMAGE_GENERATION_ERROR",
		Message:     err.Error(),
		Details:     map[string]interface{}{"adapter_id": a.adapterID},
		Timestamp:   time.Now(),
		Recoverable: strings.Contains(err.Error(), "circuit breaker") || strings.Contains(err.Error(), "timeout"),
	}

	responseBody := types.ResponseBody{
		Success: false,
		Body:    nil,
		Error:   errorInfo,
	}

	response := &ImageResponse{
		Headers: responseHeaders,
		Body:    responseBody,
	}

	a.logger.Info("In sending error response here are error response details",
		zap.Any("response", response),
	)

	a.sendResponse(topic, response, logger)
}

// sendResponse sends a response message to the specified topic
func (a *DynamicImageAdapter) sendResponse(topic string, response *ImageResponse, logger *zap.Logger) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal response", zap.Error(err))
		return
	}

	headers := response.Headers.ToMap()
	key := []byte(response.Headers.CorrelationID)

	if err := a.producer.ProduceWithValidation(a.ctx, topic, headers, key, responseBytes); err != nil {
		logger.Error("Failed to send response",
			zap.String("topic", topic),
			zap.Error(err),
		)
	} else {
		logger.Info("Response sent successfully",
			zap.String("topic", topic),
			zap.String("correlation_id", response.Headers.CorrelationID),
			zap.Bool("success", response.Body.Success),
		)
	}
}

// Shutdown gracefully shuts down the adapter
func (a *DynamicImageAdapter) Shutdown() {
	a.logger.Info("Shutting down image adapter", zap.String("adapter_id", a.adapterID))
	a.cancel()

	// Close connections
	if a.consumer != nil {
		a.consumer.Close()
	}
	if a.producer != nil {
		a.producer.Close()
	}
}

// StartHealthServer starts a simple HTTP health check server
func (a *DynamicImageAdapter) StartHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":          "healthy",
			"adapter":         "image-generator",
			"adapter_id":      a.adapterID,
			"circuit_breaker": a.httpClient.State(),
			"circuit_counts":  a.httpClient.Counts(),
		}

		w.Header().Set("Content-Type", "application/json")
		if a.httpClient.Breaker.IsOpen() {
			w.WriteHeader(http.StatusServiceUnavailable)
			status["status"] = "degraded"
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(status)
	})

	go func() {
		a.logger.Info("Starting health server", zap.String("port", port))
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			a.logger.Error("Health server failed", zap.Error(err))
		}
	}()
}
