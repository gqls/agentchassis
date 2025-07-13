// FILE: internal/adapters/imagegenerator/adapter.go
package imagegenerator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/ai-persona-system/platform/config"
	"github.com/gqls/ai-persona-system/platform/kafka"
	"github.com/gqls/ai-persona-system/platform/storage" // Our new storage library
	"go.uber.org/zap"
)

const (
	requestTopic  = "system.adapter.image.generate"
	responseTopic = "system.responses.image"
	consumerGroup = "image-generator-adapter-group"
)

// RequestPayload defines the expected data for an image generation request.
type RequestPayload struct {
	Action string `json:"action"` // e.g., "generate_from_prompt"
	Data   struct {
		Prompt      string  `json:"prompt"`
		AspectRatio string  `json:"aspect_ratio,omitempty"`
		Style       string  `json:"style,omitempty"`
		Seed        float64 `json:"seed,omitempty"`
	} `json:"data"`
}

// ResponsePayload defines the data sent back after successful generation.
type ResponsePayload struct {
	ImageURI string `json:"image_uri"`
	Prompt   string `json:"prompt"`
	Seed     int64  `json:"seed"`
}

// Adapter handles the translation between our internal system and an external API.
type Adapter struct {
	ctx           context.Context
	logger        *zap.Logger
	consumer      *kafka.Consumer
	producer      *kafka.Producer
	storageClient storage.Client // Interface for S3/MinIO
	httpClient    *http.Client
	externalAPI   string
	apiKey        string
}

// NewAdapter initializes all dependencies for the adapter.
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	// Initialize Kafka consumer and producer from platform library
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestTopic, consumerGroup, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	// Initialize Object Storage client from platform library
	storageClient, err := storage.NewS3Client(ctx, cfg.Infrastructure.ObjectStorage)
	if err != nil {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	// In a real config, you'd have a section for this adapter's specific settings
	externalAPIEndpoint := "https://api.stability.ai/v1/generation/stable-diffusion-v1-6/text-to-image"
	apiKey := os.Getenv("STABILITY_API_KEY") // Securely get API key from environment

	return &Adapter{
		ctx:           ctx,
		logger:        logger,
		consumer:      consumer,
		producer:      producer,
		storageClient: storageClient,
		httpClient:    &http.Client{Timeout: 90 * time.Second},
		externalAPI:   externalAPIEndpoint,
		apiKey:        apiKey,
	}, nil
}

// Run starts the consumer loop.
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

// handleMessage processes a single image generation request.
func (a *Adapter) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(zap.String("correlation_id", headers["correlation_id"]))

	var req RequestPayload
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.Error("Failed to unmarshal request payload", zap.Error(err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// 1. Call the external API
	imageData, err := a.callExternalImageAPI(req.Data.Prompt)
	if err != nil {
		l.Error("External image API call failed", zap.Error(err))
		// Here you might produce an error message back to a failure topic
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// 2. Upload the resulting image to Object Storage
	fileName := fmt.Sprintf("images/%s/%s.png", headers["client_id"], uuid.NewString())
	imageURI, err := a.storageClient.Upload(a.ctx, fileName, "image/png", bytes.NewReader(imageData))
	if err != nil {
		l.Error("Failed to upload image to object storage", zap.Error(err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}
	l.Info("Image successfully uploaded to storage", zap.String("uri", imageURI))

	// 3. Produce a standard response message with the URI
	responsePayload := ResponsePayload{
		ImageURI: imageURI,
		Prompt:   req.Data.Prompt,
	}
	responseBytes, _ := json.Marshal(responsePayload)

	// Prepare response headers
	responseHeaders := map[string]string{
		"correlation_id": headers["correlation_id"],
		"causation_id":   headers["request_id"],
		"request_id":     uuid.NewString(),
	}

	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders, msg.Key, responseBytes); err != nil {
		l.Error("Failed to produce response message", zap.Error(err))
	}

	// 4. Commit the original message
	a.consumer.CommitMessages(context.Background(), msg)
}

// callExternalImageAPI is a placeholder for the actual HTTP call.
func (a *Adapter) callExternalImageAPI(prompt string) ([]byte, error) {
	// ... This is where you would build the HTTP request for Stability AI,
	// add the `Authorization: Bearer ...` header, and handle the response.
	// For this example, we'll return a placeholder byte slice.
	a.logger.Info("Calling external image API", zap.String("prompt", prompt))
	return []byte("---simulated-png-image-data---"), nil
}
