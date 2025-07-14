// FILE: internal/adapters/imagegenerator/adapter.go (updated with circuit breaker)
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
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/resilience"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

const (
	requestTopic  = "system.adapter.image.generate"
	responseTopic = "system.responses.image"
	consumerGroup = "image-generator-adapter-group"
)

// RequestPayload defines the expected data for an image generation request
type RequestPayload struct {
	Action string `json:"action"`
	Data   struct {
		Prompt      string  `json:"prompt"`
		AspectRatio string  `json:"aspect_ratio,omitempty"`
		Style       string  `json:"style,omitempty"`
		Seed        float64 `json:"seed,omitempty"`
	} `json:"data"`
}

// ResponsePayload defines the data sent back after successful generation
type ResponsePayload struct {
	ImageURI string `json:"image_uri"`
	Prompt   string `json:"prompt"`
	Seed     int64  `json:"seed"`
}

// Adapter handles the translation between our internal system and an external API
type Adapter struct {
	ctx           context.Context
	logger        *zap.Logger
	consumer      *kafka.Consumer
	producer      kafka.Producer
	storageClient storage.Client
	httpClient    *resilience.HTTPClientWithBreaker
	externalAPI   string
	apiKey        string
}

// NewAdapter initializes all dependencies for the adapter
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestTopic, consumerGroup, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Initialize Object Storage client
	storageClient, err := storage.NewS3Client(ctx, cfg.Infrastructure.ObjectStorage)
	if err != nil {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	// Setup HTTP client with circuit breaker
	baseClient := &http.Client{Timeout: 90 * time.Second}
	cbConfig := resilience.DefaultCircuitBreakerConfig("stability-ai")
	cbConfig.ConsecutiveFailures = 3
	cbConfig.FailureRatio = 0.5
	httpClient := resilience.NewHTTPClientWithBreaker(baseClient, cbConfig, logger)

	externalAPIEndpoint := "https://api.stability.ai/v1/generation/stable-diffusion-v1-6/text-to-image"
	apiKey := os.Getenv("STABILITY_API_KEY")

	return &Adapter{
		ctx:           ctx,
		logger:        logger,
		consumer:      consumer,
		producer:      producer,
		storageClient: storageClient,
		httpClient:    httpClient,
		externalAPI:   externalAPIEndpoint,
		apiKey:        apiKey,
	}, nil
}

// Run starts the consumer loop
func (a *Adapter) Run() error {
	for {
		select {
		case <-a.ctx.Done():
			a.consumer.Close()
			a.producer.Close()
			return nil
		default:
			msg, err := a.consumer.FetchMessage(a.ctx)
			if err != nil {
				if err == context.Canceled {
					continue
				}
				a.logger.Error("Failed to fetch message", zap.Error(err))
				continue
			}
			go a.handleMessage(msg)
		}
	}
}

// handleMessage processes a single image generation request
func (a *Adapter) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("request_id", headers["request_id"]),
	)

	var req RequestPayload
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.Error("Failed to unmarshal request payload", zap.Error(err))
		a.sendErrorResponse(headers, errors.ValidationError("payload", "invalid JSON"))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Call the external API with circuit breaker protection
	imageData, err := a.callExternalImageAPI(req.Data.Prompt)
	if err != nil {
		l.Error("External image API call failed", zap.Error(err))

		// Check if it's a circuit breaker error
		if resilience.IsCircuitBreakerError(err) {
			retryAfter := 30 * time.Second
			a.sendErrorResponse(headers, errors.New(errors.ErrExternalService, "Image service temporarily unavailable").
				AsRetryable(&retryAfter).
				Build())
		} else {
			a.sendErrorResponse(headers, errors.New(errors.ErrAIServiceError, "Failed to generate image").
				WithCause(err).
				Build())
		}
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Upload the resulting image to Object Storage
	fileName := fmt.Sprintf("images/%s/%s.png", headers["client_id"], uuid.NewString())
	imageURI, err := a.storageClient.Upload(a.ctx, fileName, "image/png", bytes.NewReader(imageData))
	if err != nil {
		l.Error("Failed to upload image to object storage", zap.Error(err))
		a.sendErrorResponse(headers, errors.InternalError("Failed to store image", err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}
	l.Info("Image successfully uploaded to storage", zap.String("uri", imageURI))

	// Produce a standard response message with the URI
	responsePayload := ResponsePayload{
		ImageURI: imageURI,
		Prompt:   req.Data.Prompt,
	}
	a.sendSuccessResponse(headers, responsePayload)

	// Commit the original message
	a.consumer.CommitMessages(context.Background(), msg)
}

// callExternalImageAPI calls the Stability AI API with proper error handling
func (a *Adapter) callExternalImageAPI(prompt string) ([]byte, error) {
	a.logger.Info("Calling external image API", zap.String("prompt", prompt))

	requestBody := map[string]interface{}{
		"text_prompts": []map[string]interface{}{
			{"text": prompt, "weight": 1},
		},
		"cfg_scale":            7,
		"clip_guidance_preset": "FAST_BLUE",
		"height":               512,
		"width":                512,
		"samples":              1,
		"steps":                30,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(a.ctx, "POST", a.externalAPI, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.apiKey))
	req.Header.Set("Accept", "application/json")

	// Execute through circuit breaker
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response to extract the image
	var apiResponse struct {
		Artifacts []struct {
			Base64 string `json:"base64"`
		} `json:"artifacts"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResponse.Artifacts) == 0 {
		return nil, fmt.Errorf("no images in response")
	}

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(apiResponse.Artifacts[0].Base64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return imageData, nil
}

// sendSuccessResponse sends a successful response
func (a *Adapter) sendSuccessResponse(headers map[string]string, payload ResponsePayload) {
	responseBytes, _ := json.Marshal(payload)
	responseHeaders := a.createResponseHeaders(headers)

	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to produce response message", zap.Error(err))
	}
}

// sendErrorResponse sends an error response
func (a *Adapter) sendErrorResponse(headers map[string]string, domainErr *errors.DomainError) {
	responseHeaders := a.createResponseHeaders(headers)
	domainErr.TraceID = headers["correlation_id"]

	errorBytes, _ := json.Marshal(domainErr)

	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), errorBytes); err != nil {
		a.logger.Error("Failed to produce error response", zap.Error(err))
	}
}

// createResponseHeaders creates response headers with proper causality tracking
func (a *Adapter) createResponseHeaders(originalHeaders map[string]string) map[string]string {
	return map[string]string{
		"correlation_id": originalHeaders["correlation_id"],
		"causation_id":   originalHeaders["request_id"],
		"request_id":     uuid.NewString(),
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
}

// StartHealthServer starts a simple HTTP server for health checks
func (a *Adapter) StartHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":          "healthy",
			"adapter":         "image-generator",
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
