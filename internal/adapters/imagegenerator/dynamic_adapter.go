// internal/adapters/imagegenerator/dynamic_adapter.go
// Dynamic Image Generator Adapter with proper topic management and S3 integration
package imagegenerator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/adapters/imagegenerator/banana"
	"github.com/gqls/agentchassis/internal/adapters/imagegenerator/provider"
	"github.com/gqls/agentchassis/internal/adapters/imagegenerator/stability"
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

	// Phase 2I — image generation providers. Image work is routed via
	// these (kind=icon → bananaProvider, others → stabilityProvider);
	// see generateImage. httpClient/externalAPI/apiKey remain on the
	// struct because StartHealthServer still reports circuit-breaker
	// state from httpClient — they are no longer used for image
	// generation calls.
	// TODO(provider-circuit-breaker): the resilience wrapper isn't
	// threaded into stability.api.Client, so image generation HTTP
	// calls have no circuit breaker now. Plumb it through by adding a
	// Doer parameter to api.NewClient.
	stabilityProvider *stability.Provider
	bananaProvider    *banana.Provider
}

// ImageRequestData is the inner data payload of an image generation request.
// Phase 2H added: NegativePrompt, CfgScale, Steps, Seed.
// Phase 2I added: Kind, AspectRatio, ReferenceImageURIs (provider abstraction —
// the adapter routes by Kind and lets the chosen provider snap dimensions and
// honour reference images as it can). All fields optional; legacy callers
// sending only {prompt, style, width, height} produce identical JSON.
type ImageRequestData struct {
	Prompt         string  `json:"prompt"`
	Style          string  `json:"style,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	CfgScale       float64 `json:"cfg_scale,omitempty"`
	Steps          int     `json:"steps,omitempty"`
	Seed           int     `json:"seed,omitempty"`

	// Kind identifies the image purpose ("icon", "hero", "logo",
	// "illustration", "infographic"). Drives provider routing
	// (icon → banana, others → stability) and per-kind defaults
	// when AspectRatio/Width/Height are absent.
	Kind string `json:"kind,omitempty"`

	// AspectRatio is a semantic ratio label ("1:1", "16:9", "9:16",
	// "4:3", "3:4", "3:2", "2:3", "21:9", "5:4", etc). Providers
	// translate to provider-specific dimensions (e.g. Stability snaps
	// to SDXL v1.0 whitelist; Banana passes through). Takes precedence
	// over Width/Height when set; both forms coexist for backward compat.
	AspectRatio string `json:"aspect_ratio,omitempty"`

	// ReferenceImageURIs are optional style/content anchors. S3 URIs
	// (s3://...) or HTTPS URLs. Banana fetches and sends as multimodal
	// input; Stability v1 REST API ignores (no IP-Adapter on the public
	// endpoint). Plumbed end-to-end so the field is forward-compatible
	// even before discovery emits values.
	ReferenceImageURIs []string `json:"reference_image_uris,omitempty"`
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

	// Phase 2I — construct image-generation providers.
	// STABILITY_API_KEY / STABILITY_API_ENDPOINT are existing env vars.
	// BANANA_API_KEY / BANANA_DEFAULT_MODEL are new; the BANANA_API_KEY
	// must exist in personae-default-secrets. BANANA_DEFAULT_MODEL has
	// a code default if env is unset.
	stabilityProvider, err := stability.New(stability.Config{
		APIKey:  apiKey,
		BaseURL: deriveStabilityBaseURL(externalAPI),
	}, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to construct stability provider: %w", err)
	}

	bananaProvider, err := banana.New(banana.Config{
		APIKey:       os.Getenv("BANANA_API_KEY"),
		BaseURL:      os.Getenv("BANANA_BASE_URL"),
		DefaultModel: os.Getenv("BANANA_DEFAULT_MODEL"),
	}, &referenceFetcher{
		storageClient: storageClient,
		bucket:        storageConfig.Bucket,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		logger:        logger,
	}, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to construct banana provider: %w", err)
	}

	adapter := &DynamicImageAdapter{
		ctx:               adapterCtx,
		cancel:            cancel,
		logger:            logger,
		config:            cfg,
		consumer:          consumer,
		producer:          producer,
		storageClient:     storageClient,
		httpClient:        httpClient,
		externalAPI:       externalAPI,
		apiKey:            apiKey,
		adapterID:         adapterID,
		podName:           podName,
		stabilityProvider: stabilityProvider,
		bananaProvider:    bananaProvider,
	}

	logger.Info("Dynamic image adapter initialized",
		zap.String("adapter_id", adapterID),
		zap.String("pod_name", podName),
		zap.String("consumer_group", consumerGroup),
		zap.Any("storageClient", storageClient),
		zap.String("externalAPI", externalAPI),
		zap.String("apiKey", apiKey),
		zap.String("adapterID", adapterID),
		zap.String("banana_default_model", os.Getenv("BANANA_DEFAULT_MODEL")),
		zap.Bool("banana_key_present", os.Getenv("BANANA_API_KEY") != ""),
	)

	return adapter, nil
}

// deriveStabilityBaseURL extracts the Stability API base URL from the
// legacy STABILITY_API_ENDPOINT env var (which is a full path including
// engine + operation, e.g. ".../v1/generation/{engine}/text-to-image").
// The stability provider's api.Client wants just the host root
// ("https://api.stability.ai") so it can append /v1/generation/{engine}/
// internally per request. If the env var is empty or malformed, we
// fall back to api.DefaultBaseURL (set inside the provider).
func deriveStabilityBaseURL(legacyEndpoint string) string {
	if legacyEndpoint == "" {
		return ""
	}
	// Trim any path suffix; keep scheme + host. Best-effort — if the
	// env var doesn't fit the expected shape, return empty and let the
	// provider's default kick in.
	if i := strings.Index(legacyEndpoint, "/v1/"); i > 0 {
		return legacyEndpoint[:i]
	}
	return ""
}

// referenceFetcher is the project's provider.ReferenceFetcher
// implementation. Handles two URI schemes:
//
//   - s3://<bucket>/<key>  — dispatches to storageClient.Download.
//     The storageClient is bound to exactly one bucket (IMAGE_BUCKET),
//     so only refs in that same bucket can be fetched this way. Refs in
//     other buckets (e.g. site-assets bucket) error with a clear message.
//   - https:// / http:// — plain HTTP GET. Useful for presigned URLs
//     and external CDN-hosted references.
//
// MIME type for s3:// is inferred from the key extension (.png, .jpg,
// .jpeg, .webp). storage.Client.Download doesn't expose the real
// Content-Type from the S3 response; if exact MIME matters for a
// reference, prefer the https:// scheme (response carries Content-Type).
type referenceFetcher struct {
	storageClient storage.Client
	bucket        string // the bucket storageClient is bound to
	httpClient    *http.Client
	logger        *zap.Logger
}

// Compile-time check.
var _ provider.ReferenceFetcher = (*referenceFetcher)(nil)

func (f *referenceFetcher) Fetch(ctx context.Context, uri string) ([]byte, string, error) {
	switch {
	case storage.IsS3URI(uri):
		return f.fetchS3(ctx, uri)
	case strings.HasPrefix(uri, "http://"), strings.HasPrefix(uri, "https://"):
		return f.fetchHTTP(ctx, uri)
	default:
		return nil, "", fmt.Errorf("referenceFetcher: unsupported URI scheme: %s", uri)
	}
}

func (f *referenceFetcher) fetchS3(ctx context.Context, uri string) ([]byte, string, error) {
	bucket, key, err := storage.ParseS3URI(uri)
	if err != nil {
		return nil, "", fmt.Errorf("referenceFetcher: parse s3 URI: %w", err)
	}
	if bucket != f.bucket {
		return nil, "", fmt.Errorf("referenceFetcher: s3 URI bucket %q does not match configured bucket %q (cross-bucket refs not supported)",
			bucket, f.bucket)
	}
	reader, err := f.storageClient.Download(ctx, key)
	if err != nil {
		return nil, "", fmt.Errorf("referenceFetcher: download s3://%s/%s: %w", bucket, key, err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("referenceFetcher: read s3 body: %w", err)
	}
	mime := mimeFromKey(key)
	f.logger.Debug("Reference image fetched (s3)",
		zap.String("uri", uri),
		zap.String("mime", mime),
		zap.Int("bytes", len(data)),
	)
	return data, mime, nil
}

func (f *referenceFetcher) fetchHTTP(ctx context.Context, uri string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, "", fmt.Errorf("referenceFetcher: build request: %w", err)
	}
	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("referenceFetcher: http fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("referenceFetcher: http %s: status %d", uri, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("referenceFetcher: read http body: %w", err)
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = mimeFromKey(uri)
	}
	f.logger.Debug("Reference image fetched (http)",
		zap.String("uri", uri),
		zap.String("mime", mime),
		zap.Int("bytes", len(body)),
	)
	return body, mime, nil
}

// mimeFromKey infers an image MIME type from a path's file extension.
// Defaults to image/png — Gemini accepts PNG/JPEG/WEBP and PNG is the
// safest fallback (Stability outputs PNG too, so generated refs match).
func mimeFromKey(key string) string {
	lower := strings.ToLower(key)
	switch {
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	default:
		return "image/png"
	}
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

	// Generate image via the routed provider.
	// Phase 2I: returns origin ("provider/model_id") for asset
	// provenance — passed into sendSuccessResponse so the workflow's
	// store_*_asset step can record it instead of hardcoded "sdxl".
	imageData, _, origin, err := a.generateImage(request.Body.Data)
	if err != nil {
		logger.Error("Failed to generate image", zap.Error(err))
		a.sendErrorResponse(responseTopic, &request, err, logger)
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	a.logger.Info("image about to be uploaded in handleMessage dynamic adapter",
		zap.String("origin", origin))

	// Upload to S3 and get presigned URL
	imageURI, presignedURL, err := a.uploadImage(imageData, request.Headers.ClientID, logger)
	if err != nil {
		logger.Error("Failed to upload image", zap.Error(err))
		a.sendErrorResponse(responseTopic, &request, err, logger)
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Send success response with both S3 URI and presigned URL
	a.sendSuccessResponse(responseTopic, &request, imageURI, presignedURL, origin, time.Since(startTime), logger)

	// Commit the message
	a.consumer.CommitMessages(context.Background(), msg)

	logger.Info("Image generation complete",
		zap.String("image_uri", imageURI),
		zap.String("presigned_url", presignedURL),
		zap.Duration("duration", time.Since(startTime)),
	)
}

// generateImage routes the request to the appropriate provider based on
// data.Kind. Returns the raw image bytes, the MIME type, and a
// provenance string ("provider/model_id") for asset-row origin tracking.
//
// Phase 2I — replaces the previous inline Stability HTTP call. SDXL
// dimension whitelist, kind defaults, negative-prompt defaults and
// Banana's multimodal reference handling now live inside the provider
// implementations (internal/adapters/imagegenerator/stability,
// internal/adapters/imagegenerator/banana). This function is just a
// router; nothing here knows about Stability or Gemini wire formats.
func (a *DynamicImageAdapter) generateImage(data ImageRequestData) ([]byte, string, string, error) {
	req := provider.Request{
		Kind:               data.Kind,
		Prompt:             data.Prompt,
		NegativePrompt:     data.NegativePrompt,
		AspectRatio:        data.AspectRatio,
		ReferenceImageURIs: data.ReferenceImageURIs,
	}

	// Routing: icons → banana (flat illustration quality), everything
	// else → stability (proven on hero/logo/illustration kinds).
	// Empty kind also goes to stability for backward compat with legacy
	// callers that don't set the field.
	var p provider.Provider
	switch data.Kind {
	case "icon":
		p = a.bananaProvider
	default:
		p = a.stabilityProvider
	}

	a.logger.Info("generateImage: dispatching to provider",
		zap.String("kind", data.Kind),
		zap.String("provider", p.Name()),
		zap.String("aspect_ratio", data.AspectRatio),
		zap.Int("prompt_len", len(data.Prompt)),
		zap.Bool("has_negative_prompt", data.NegativePrompt != ""),
		zap.Int("reference_count", len(data.ReferenceImageURIs)),
	)

	result, err := p.GenerateImage(a.ctx, req)
	if err != nil {
		// Surface provider sentinels in the error so the action layer
		// can errors.Is against provider.ErrSafetyBlocked / ErrInvalidRequest.
		return nil, "", "", fmt.Errorf("provider %s: %w", p.Name(), err)
	}

	origin := result.ProviderName + "/" + result.ModelID
	a.logger.Info("generateImage: provider succeeded",
		zap.String("provider", result.ProviderName),
		zap.String("model_id", result.ModelID),
		zap.String("mime_type", result.MimeType),
		zap.Int("bytes", len(result.ImageBytes)),
	)
	return result.ImageBytes, result.MimeType, origin, nil
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
	origin string,
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
			// Phase 2I — origin = "provider/model_id" (e.g.
			// "banana/gemini-3-pro-image-preview"). The workflow's
			// store_*_asset step can read this from
			// generate.response.origin_model via output_mapping
			// instead of the hardcoded "sdxl" it uses today.
			"origin_model": origin,
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
