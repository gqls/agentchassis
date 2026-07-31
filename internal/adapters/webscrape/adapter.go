// internal/adapters/webscrape/adapter.go
package webscrape

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/adapters/shared/throttle"
	"github.com/gqls/agentchassis/internal/adapters/webscrape/providers"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/fetchguard"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

const (
	requestTopic  = "system.adapter.webscrape.requests"
	consumerGroup = "webscrape-adapter-group"
)

// RequestPayload for web scraping
type RequestPayload struct {
	RequestID       string                 `json:"request_id"`
	Action          string                 `json:"action"`
	URL             string                 `json:"url"`
	Config          map[string]interface{} `json:"config,omitempty"`
	ClientID        string                 `json:"client_id,omitempty"`
	UploadResults   bool                   `json:"upload_results,omitempty"`
	ReplyToTopic    string                 `json:"reply_to_topic"`
	CorrelationID   string                 `json:"correlation_id"`
	OrchestrationID string                 `json:"orchestration_id"`
}

// ResponsePayload with scraping results
type ResponsePayload struct {
	RequestID       string      `json:"request_id"`
	CorrelationID   string      `json:"correlation_id"`
	OrchestrationID string      `json:"orchestration_id"`
	Timestamp       string      `json:"timestamp"`
	Result          interface{} `json:"result"`
}

// Adapter handles web scraping requests with s3 storage
type Adapter struct {
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *zap.Logger
	consumer        *kafka.Consumer
	producer        kafka.Producer
	providers       map[string]providers.ScrapingProvider
	storageClient   storage.Client
	httpClient      *http.Client // fixed, trusted hosts only (the scraping provider's own API)
	imageHTTPClient *http.Client // bugs_open/159: fetches URLs taken from SCRAPED PAGE CONTENT — attacker-influenced by construction, so this one is fetchguard-wrapped and httpClient is deliberately not
	config          *config.ServiceConfig
	healthServer    *http.Server
	throttle        *throttle.Throttle
}

// NewAdapter creates a new web scraping adapter
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	// Create a cancelable context
	adapterCtx, cancel := context.WithCancel(ctx)

	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestTopic, consumerGroup, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		cancel()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Initialize storage client (S3)
	storageConfig := cfg.Infrastructure.ObjectStorage
	storageConfig.Provider = cfg.Infrastructure.ObjectStorage.Provider
	storageConfig.Endpoint = cfg.Infrastructure.ObjectStorage.Endpoint
	storageConfig.Bucket = cfg.Infrastructure.ObjectStorage.Bucket
	storageConfig.AccessKeyEnvVar = cfg.Infrastructure.ObjectStorage.AccessKeyEnvVar
	storageConfig.SecretKeyEnvVar = cfg.Infrastructure.ObjectStorage.SecretKeyEnvVar

	zap.Strings("environment Environ", os.Environ())

	if storageConfig.Endpoint == "" {
		storageConfig.Endpoint = os.Getenv("S3_ENDPOINT")
	}
	if storageConfig.Bucket == "" {
		storageConfig.Bucket = os.Getenv("IMAGE_BUCKET")
	}

	// Initialize storage client for S3/Backblaze
	storageClient, err := storage.NewS3Client(ctx, storageConfig, *logger)
	if err != nil {
		consumer.Close()
		producer.Close()
		cancel()
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// downloadImage fetches whatever image URL the scraped page's own content
	// named (bugs_open/159) — that page belongs to a domain this platform did
	// not choose, so the URL is attacker-influenced by construction. Guarded
	// separately from httpClient above, which only ever talks to the fixed,
	// trusted scraping-provider API.
	imageHTTPClient := fetchguard.NewClient(fetchguard.DefaultConfig())
	imageHTTPClient.Timeout = 120 * time.Second

	requestThrottle := throttle.New(logger)

	// Initialize providers
	scrapingProviders := make(map[string]providers.ScrapingProvider)

	// Add Firecrawl provider
	firecrawl := providers.NewFirecrawlScrapingProvider(httpClient, storageClient, logger)
	if firecrawl.IsAvailable() {
		scrapingProviders["firecrawl"] = firecrawl
		logger.Info("Firecrawl provider available with S3 support")
	}

	// Could add other providers here (Playwright, Puppeteer, etc.)

	if len(scrapingProviders) == 0 {
		consumer.Close()
		producer.Close()
		cancel()
		return nil, fmt.Errorf("no scraping providers available - check API keys")
	}

	return &Adapter{
		ctx:             adapterCtx,
		cancel:          cancel,
		logger:          logger,
		consumer:        consumer,
		producer:        producer,
		providers:       scrapingProviders,
		storageClient:   storageClient,
		httpClient:      httpClient,
		imageHTTPClient: imageHTTPClient,
		throttle:        requestThrottle,
		config:          cfg,
	}, nil
}

// Run starts the adapter's main loop
func (a *Adapter) Run() error {
	a.logger.Info("Web scraping adapter running",
		zap.Int("providers_count", len(a.providers)))

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Shutting down web scraping adapter")
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
				time.Sleep(time.Second)
				continue
			}
			// go a.handleMessage(msg)
			a.handleMessage(msg) // Sequential, not concurrent
			a.throttle.Wait()    // Delay before next request
		}
	}
}

// handleMessage processes a scraping request
func (a *Adapter) handleMessage(msg kafka.Message) {
	startTime := time.Now()

	// Parse the comprehensive message format
	var fullMessage map[string]interface{}
	if err := json.Unmarshal(msg.Value, &fullMessage); err != nil {
		a.logger.Error("Failed to unmarshal message", zap.Error(err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Extract headers and body
	headers, headersOk := fullMessage["headers"].(map[string]interface{})
	body, bodyOk := fullMessage["body"].(map[string]interface{})

	if !headersOk || !bodyOk {
		a.logger.Error("Invalid message format - missing headers or body")
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Extract key fields from headers
	requestID, _ := headers["request_id"].(string)
	correlationID, _ := headers["correlation_id"].(string)
	orchestrationID, _ := headers["orchestration_id"].(string)
	clientID, _ := headers["client_id"].(string)
	stepName, _ := headers["step_name"].(string)

	// Extract action and data from body
	action, _ := body["action"].(string)
	data, _ := body["data"].(map[string]interface{})
	replyToTopic, _ := body["reply_to_topic"].(string)

	// Extract URL and config from data
	url, _ := data["url"].(string)
	uploadResults, _ := data["upload_results"].(bool)
	scrapeConfig, _ := data["config"].(map[string]interface{})

	// Build RequestPayload struct for upload function
	req := RequestPayload{
		RequestID:       requestID,
		Action:          action,
		URL:             url,
		Config:          scrapeConfig,
		ClientID:        clientID,
		UploadResults:   uploadResults,
		ReplyToTopic:    replyToTopic,
		CorrelationID:   correlationID,
		OrchestrationID: orchestrationID,
	}

	l := a.logger.With(
		zap.String("request_id", requestID),
		zap.String("correlation_id", correlationID),
		zap.String("orchestration_id", orchestrationID),
		zap.String("action", action),
		zap.String("url", url),
		zap.String("client_id", clientID),
	)

	l.Info("Processing webscrape request")

	// batch_scrape carries its URLs in data["urls"] and legitimately has NO
	// top-level url, so it must branch out BEFORE the single-url validation.
	// Until 2026-07-20 this guard ran first, which rejected every batch_scrape
	// ever sent as "Empty URL in request" — batch_webscrape (research-agent's
	// scrape step, the evidence-researcher) could not work at all.
	if action == "batch_scrape" {
		a.handleBatchScrape(a.ctx, headers, body, replyToTopic, l)
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Validate required fields (single-url actions only)
	if url == "" {
		l.Error("Empty URL in request")
		a.sendErrorResponse(requestID, correlationID, orchestrationID, replyToTopic, clientID, stepName, "URL cannot be empty")
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Get provider (default to firecrawl)
	providerName := "firecrawl"
	if p, ok := scrapeConfig["provider"].(string); ok {
		providerName = p
	}

	provider, ok := a.providers[providerName]
	if !ok {
		l.Error("Provider not found", zap.String("provider", providerName))
		a.sendErrorResponse(requestID, correlationID, orchestrationID, replyToTopic, clientID, stepName,
			fmt.Sprintf("Provider %s not available", providerName))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Execute the scraping action
	var result interface{}
	var err error

	switch action {
	case "scrape":
		result, err = provider.Scrape(a.ctx, url, scrapeConfig)
	case "crawl":
		result, err = provider.Crawl(a.ctx, url, scrapeConfig)
	case "map":
		result, err = provider.Map(a.ctx, url, scrapeConfig)
	case "extract":
		schema, ok := scrapeConfig["schema"].(map[string]interface{})
		if !ok {
			err = fmt.Errorf("schema not provided for extract action")
		} else {
			result, err = provider.ExtractStructured(a.ctx, url, schema, scrapeConfig)
		}
	// batch_scrape is handled above, before the single-url validation.
	default:
		err = fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		l.Error("Action failed", zap.Error(err))
		a.sendErrorResponse(requestID, correlationID, orchestrationID, replyToTopic, clientID, stepName, err.Error())
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	l.Info("In handleMessage in adapter.go about to upload results",
		zap.Any("DEBUGaa: result", result))

	// Upload results to S3 if requested
	if uploadResults && a.storageClient != nil {
		uploadedResult, err := a.uploadScrapingResults(result, req, l)
		l.Info("In handleMessage result of upload to S3",
			zap.Any("uploadedResult", uploadedResult),
		)
		if err != nil {
			l.Error("Failed to upload results", zap.Error(err))
		} else {
			// Merge upload info with result
			if resultMap, ok := result.(map[string]interface{}); ok {
				resultMap["storage"] = uploadedResult
				result = resultMap
			}
		}
	}

	l.Info("In handleMessage in adapter.go just uploaded results - about to send success response")

	// Cap content-bearing fields for transport. The marker each cut carries can
	// only claim a stored copy when the upload above actually recorded a URI for
	// THAT field — see truncation.go. bugs_open/133: this used to claim
	// "full version in S3" unconditionally, including on the four live steps
	// that never upload anything.
	if resultMap, ok := result.(map[string]interface{}); ok {
		truncateResultForTransport(resultMap, storageInfoOf(resultMap), l)
		result = resultMap
	}

	// Send success response
	a.sendSuccessResponse(requestID, correlationID, orchestrationID, replyToTopic, clientID, stepName, result)
	a.consumer.CommitMessages(context.Background(), msg)

	l.Info("Request processed successfully",
		zap.Duration("duration", time.Since(startTime)))
}

// Updated response methods to handle new parameters
func (a *Adapter) sendSuccessResponse(requestID, correlationID, orchestrationID, replyTopic, clientID, stepName string, result interface{}) {
	if replyTopic == "" {
		a.logger.Warn("No reply topic specified", zap.String("request_id", requestID))
		return
	}

	response := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":            correlationID,
			"orchestration_id":          orchestrationID,
			"in_response_to_request_id": requestID,
			"status":                    "complete",
			"request_id":                requestID,
			"message_type":              "response",
			"timestamp":                 time.Now().UTC().Format(time.RFC3339),
			"success":                   true,
			"client_id":                 clientID,
			"sender_agent_type":         "webscrape-adapter",
			"in_response_to_step_name":  stepName,
			"sender": map[string]interface{}{
				"agent_type": "webscrape-adapter",
				"agent_id":   "webscrape-adapter-001",
				"pod_name":   os.Getenv("HOSTNAME"),
			},
		},
		"body": map[string]interface{}{
			"success": true,
			"body": map[string]interface{}{
				"data": result,
			},
			"error": nil,
			"metadata": map[string]interface{}{
				"processed_at": time.Now().UTC().Format(time.RFC3339),
				"adapter":      "webscrape",
			},
		},
	}

	responseBytes, _ := json.Marshal(response)

	// Create headers map for Kafka
	headers := make(map[string]string)
	headers["correlation_id"] = correlationID
	headers["in_response_to_request_id"] = requestID
	headers["request_id"] = requestID
	headers["orchestration_id"] = orchestrationID
	headers["message_type"] = "response"
	headers["status"] = "complete"
	headers["client_id"] = clientID
	headers["sender_agent_type"] = "webscrape-adapter"
	headers["in_response_to_step_name"] = stepName

	a.logger.Info("Sending success response",
		zap.String("request_id", requestID),
		zap.String("reply_topic", replyTopic),
		zap.String("status", "complete"))

	// bugs_open/133 defect B: this used to log the produce failure and return,
	// so a reply the broker refused as too large was never delivered, never
	// degraded and never reported — the caller waits on the reply topic, not on
	// this pod's logs, and starved through the coordinator's whole retry budget.
	// The policy is shared with the batch path (platform/kafka.DeliverReply);
	// only the degraded form and the error envelope are ours.
	outcome, derr := kafka.DeliverReply(
		a.ctx,
		a.producer,
		a.logger,
		replyTopic,
		headers,
		[]byte(correlationID),
		responseBytes,
		func() ([]byte, error) {
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("scrape result is not a map, nothing to strip")
			}
			// Mutates the map the envelope already holds by reference, so
			// re-marshalling the same envelope yields the degraded reply.
			stripResultForRetry(resultMap, storageInfoOf(resultMap))
			return json.Marshal(response)
		},
	)

	if outcome == kafka.FailedUndeliverable {
		// A response that cannot be delivered must become a deliverable error,
		// never silence (016b §9, from bugs_closed/062).
		a.logger.Error("Scrape response undeliverable — sending error response instead",
			zap.Error(derr),
			zap.String("request_id", requestID),
			zap.String("topic", replyTopic))
		a.sendErrorResponse(requestID, correlationID, orchestrationID, replyTopic, clientID, stepName,
			fmt.Sprintf("scrape succeeded but the response could not be delivered: %v", derr))
	}
}

func (a *Adapter) sendErrorResponse(requestID, correlationID, orchestrationID, replyTopic, clientID, stepName, errorMsg string) {
	if replyTopic == "" {
		a.logger.Warn("No reply topic specified for error response", zap.String("request_id", requestID))
		return
	}

	response := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":            correlationID,
			"orchestration_id":          orchestrationID,
			"in_response_to_request_id": requestID,
			"request_id":                requestID,
			"status":                    "error_recoverable",
			"message_type":              "response",
			"timestamp":                 time.Now().UTC().Format(time.RFC3339),
			"success":                   false,
			"client_id":                 clientID,
			"sender_agent_type":         "webscrape-adapter",
			"in_response_to_step_name":  stepName,
			"sender": map[string]interface{}{
				"agent_type": "webscrape-adapter",
				"agent_id":   "webscrape-adapter-001",
				"pod_name":   os.Getenv("HOSTNAME"),
			},
		},
		"body": map[string]interface{}{
			"success": false,
			"body": map[string]interface{}{
				"data": nil,
			},
			"error": map[string]interface{}{
				"message":     errorMsg,
				"code":        "WEBSCRAPE_ERROR",
				"recoverable": true,
			},
			"metadata": map[string]interface{}{
				"processed_at": time.Now().UTC().Format(time.RFC3339),
				"adapter":      "webscrape",
			},
		},
	}

	responseBytes, _ := json.Marshal(response)

	// Create headers map for Kafka
	headers := make(map[string]string)
	headers["correlation_id"] = correlationID
	headers["in_response_to_request_id"] = requestID
	headers["request_id"] = requestID
	headers["orchestration_id"] = orchestrationID
	headers["message_type"] = "response"
	headers["status"] = "error_recoverable"
	headers["success"] = "false"
	headers["client_id"] = clientID
	headers["sender_agent_type"] = "webscrape-adapter"
	headers["in_response_to_step_name"] = stepName

	a.logger.Error("Sending error response",
		zap.String("request_id", requestID),
		zap.String("error", errorMsg))

	if err := a.producer.ProduceWithValidation(
		a.ctx,
		replyTopic,
		headers,
		[]byte(correlationID),
		responseBytes,
	); err != nil {
		a.logger.Error("Failed to produce error response",
			zap.Error(err),
			zap.String("topic", replyTopic))
	}
}

// uploadScrapingResults uploads various scraping results to S3
func (a *Adapter) uploadScrapingResults(result interface{}, req RequestPayload, logger *zap.Logger) (map[string]interface{}, error) {
	a.logger.Error("In uploadScrapingResults",
		zap.Any("DEBUGaa: result", result),
		zap.Any("DEBUGaa: request payload", req),
	)

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("result is not a map")
	}

	uploadInfo := make(map[string]interface{})
	timestamp := time.Now().Format("20060102-150405")
	scrapeID := uuid.NewString()

	// Create base path for this scrape session
	basePath := fmt.Sprintf("webscrape/%s/%s/%s", req.ClientID, timestamp[:8], scrapeID)

	// Upload HTML content if present
	if htmlContent, ok := resultMap["html_content"].(string); ok && htmlContent != "" {
		htmlURI, htmlURL, err := a.uploadContent(
			[]byte(htmlContent),
			fmt.Sprintf("%s/content.html", basePath),
			"text/html",
			logger,
		)
		if err != nil {
			logger.Warn("Failed to upload HTML content", zap.Error(err))
		} else {
			uploadInfo["html_uri"] = htmlURI
			if htmlURL != "" {
				uploadInfo["html_url"] = htmlURL
			}
		}
	}

	// Upload Markdown content if present
	if markdownContent, ok := resultMap["markdown_content"].(string); ok && markdownContent != "" {
		mdURI, mdURL, err := a.uploadContent(
			[]byte(markdownContent),
			fmt.Sprintf("%s/content.md", basePath),
			"text/markdown",
			logger,
		)
		if err != nil {
			logger.Warn("Failed to upload Markdown content", zap.Error(err))
		} else {
			uploadInfo["markdown_uri"] = mdURI
			if mdURL != "" {
				uploadInfo["markdown_url"] = mdURL
			}
		}
	}

	// Upload screenshot - handle both base64 and URL formats
	var screenshotData []byte
	var screenshotSource string

	// Try base64 first
	if screenshot, ok := resultMap["screenshot_base64"].(string); ok && screenshot != "" {
		data, err := base64.StdEncoding.DecodeString(screenshot)
		if err == nil {
			screenshotData = data
			screenshotSource = "base64"
			logger.Debug("Screenshot from base64", zap.Int("size", len(screenshotData)))
		} else {
			logger.Warn("Failed to decode base64 screenshot", zap.Error(err))
		}
	}

	// Try URL if no base64 data
	if len(screenshotData) == 0 {
		if screenshotURL, ok := resultMap["screenshot_url"].(string); ok && screenshotURL != "" {
			logger.Info("Downloading screenshot from URL", zap.String("url", screenshotURL))

			// Download screenshot from URL (e.g., Google Cloud Storage)
			data, contentType, err := a.downloadImage(a.ctx, screenshotURL)
			if err != nil {
				logger.Warn("Failed to download screenshot from URL",
					zap.String("url", screenshotURL),
					zap.Error(err))
			} else {
				screenshotData = data
				screenshotSource = "url"
				logger.Info("Screenshot downloaded from URL",
					zap.Int("size", len(screenshotData)),
					zap.String("content_type", contentType))
			}
		}
	}

	// Upload screenshot to S3 if we have data
	if len(screenshotData) > 0 {
		screenshotURI, screenshotURL, err := a.uploadContent(
			screenshotData,
			fmt.Sprintf("%s/screenshot.png", basePath),
			"image/png",
			logger,
		)
		if err != nil {
			logger.Warn("Failed to upload screenshot", zap.Error(err))
		} else {
			uploadInfo["screenshot_uri"] = screenshotURI
			if screenshotURL != "" {
				uploadInfo["screenshot_url"] = screenshotURL
			}
			uploadInfo["screenshot_source"] = screenshotSource
			logger.Info("Screenshot uploaded to S3",
				zap.String("uri", screenshotURI),
				zap.String("source", screenshotSource),
				zap.Int("size", len(screenshotData)))
		}
	}

	// Upload site images if present
	if images, ok := resultMap["images"].([]interface{}); ok && len(images) > 0 {
		logger.Info("Found images to upload", zap.Int("count", len(images)))

		imageUploadInfo := []map[string]interface{}{}
		successCount := 0
		failCount := 0

		for i, img := range images {
			// Limit to reasonable number of images
			if i >= 50 {
				logger.Warn("Reached image upload limit", zap.Int("limit", 50))
				break
			}

			var imageURL string
			var imageAlt string

			// Images can be strings (URLs) or objects with url/alt
			switch imgData := img.(type) {
			case string:
				imageURL = imgData
			case map[string]interface{}:
				if url, ok := imgData["url"].(string); ok {
					imageURL = url
				}
				if alt, ok := imgData["alt"].(string); ok {
					imageAlt = alt
				}
			default:
				logger.Warn("Unexpected image format", zap.Any("image", img))
				continue
			}

			if imageURL == "" {
				continue
			}

			// Skip data URLs (too large, already embedded)
			if len(imageURL) > 10 && imageURL[:5] == "data:" {
				logger.Debug("Skipping data URL image", zap.Int("index", i))
				continue
			}

			// Download the image
			imageData, contentType, err := a.downloadImage(a.ctx, imageURL)
			if err != nil {
				logger.Warn("Failed to download image",
					zap.Int("index", i),
					zap.String("url", imageURL),
					zap.Error(err))
				failCount++
				continue
			}

			// Skip very small images (likely tracking pixels)
			if len(imageData) < 1024 {
				logger.Debug("Skipping small image",
					zap.Int("index", i),
					zap.Int("size", len(imageData)))
				continue
			}

			// Upload to S3
			ext := getImageExtension(contentType)
			imageKey := fmt.Sprintf("%s/images/image_%03d%s", basePath, i, ext)

			imageURI, imagePresignedURL, err := a.uploadContent(
				imageData,
				imageKey,
				contentType,
				logger,
			)
			if err != nil {
				logger.Warn("Failed to upload image",
					zap.Int("index", i),
					zap.String("url", imageURL),
					zap.Error(err))
				failCount++
				continue
			}

			// Store image info
			imgInfo := map[string]interface{}{
				"index":        i,
				"original_url": imageURL,
				"s3_uri":       imageURI,
				"content_type": contentType,
				"size_bytes":   len(imageData),
			}

			if imagePresignedURL != "" {
				imgInfo["s3_url"] = imagePresignedURL
			}

			if imageAlt != "" {
				imgInfo["alt"] = imageAlt
			}

			imageUploadInfo = append(imageUploadInfo, imgInfo)
			successCount++

			logger.Debug("Image uploaded successfully",
				zap.Int("index", i),
				zap.String("uri", imageURI),
				zap.Int("size", len(imageData)))
		}

		if len(imageUploadInfo) > 0 {
			uploadInfo["images"] = imageUploadInfo
			uploadInfo["images_uploaded_count"] = successCount
			uploadInfo["images_failed_count"] = failCount

			logger.Info("Images upload complete",
				zap.Int("success", successCount),
				zap.Int("failed", failCount),
				zap.Int("total", len(images)))
		}
	}

	// Upload raw HTML if present
	if rawHTML, ok := resultMap["raw_html"].(string); ok && rawHTML != "" {
		rawURI, rawURL, err := a.uploadContent(
			[]byte(rawHTML),
			fmt.Sprintf("%s/raw.html", basePath),
			"text/html",
			logger,
		)
		if err != nil {
			logger.Warn("Failed to upload raw HTML", zap.Error(err))
		} else {
			uploadInfo["raw_html_uri"] = rawURI
			if rawURL != "" {
				uploadInfo["raw_html_url"] = rawURL
			}
		}
	}

	// Upload metadata as JSON
	metadata := map[string]interface{}{
		"url":            req.URL,
		"action":         req.Action,
		"client_id":      req.ClientID,
		"request_id":     req.RequestID,
		"correlation_id": req.CorrelationID,
		"scraped_at":     time.Now().UTC().Format(time.RFC3339),
		"scrape_id":      scrapeID,
	}

	// Add any additional metadata from result
	if resultMeta, ok := resultMap["metadata"].(map[string]interface{}); ok {
		for k, v := range resultMeta {
			metadata[k] = v
		}
	}

	metadataJSON, _ := json.MarshalIndent(metadata, "", "  ")
	metaURI, metaURL, err := a.uploadContent(
		metadataJSON,
		fmt.Sprintf("%s/metadata.json", basePath),
		"application/json",
		logger,
	)
	if err != nil {
		logger.Warn("Failed to upload metadata", zap.Error(err))
	} else {
		uploadInfo["metadata_uri"] = metaURI
		if metaURL != "" {
			uploadInfo["metadata_url"] = metaURL
		}
	}

	// For crawl actions with multiple pages
	if pages, ok := resultMap["pages"].([]interface{}); ok {
		pageURIs := []map[string]string{}
		for i, page := range pages {
			if pageMap, ok := page.(map[string]interface{}); ok {
				pageInfo := make(map[string]string)

				// Upload each page's content
				if pageHTML, ok := pageMap["html"].(string); ok {
					pageURI, pageURL, err := a.uploadContent(
						[]byte(pageHTML),
						fmt.Sprintf("%s/pages/page_%d.html", basePath, i),
						"text/html",
						logger,
					)
					if err == nil {
						pageInfo["html_uri"] = pageURI
						if pageURL != "" {
							pageInfo["html_url"] = pageURL
						}
					}
				}

				if pageMarkdown, ok := pageMap["markdown"].(string); ok {
					pageURI, pageURL, err := a.uploadContent(
						[]byte(pageMarkdown),
						fmt.Sprintf("%s/pages/page_%d.md", basePath, i),
						"text/markdown",
						logger,
					)
					if err == nil {
						pageInfo["markdown_uri"] = pageURI
						if pageURL != "" {
							pageInfo["markdown_url"] = pageURL
						}
					}
				}

				if len(pageInfo) > 0 {
					pageURIs = append(pageURIs, pageInfo)
				}
			}
		}
		if len(pageURIs) > 0 {
			uploadInfo["pages"] = pageURIs
		}
	}

	uploadInfo["base_path"] = basePath
	uploadInfo["scrape_id"] = scrapeID
	uploadInfo["uploaded_at"] = time.Now().UTC().Format(time.RFC3339)

	logger.Info("Scraping results uploaded to S3",
		zap.String("base_path", basePath),
		zap.String("scrape_id", scrapeID),
		zap.Int("files_uploaded", len(uploadInfo)-3), // Subtract metadata fields
		// zap.Any("DEBUGaa: files_uploaded uploadedInfo", uploadInfo),
	)

	return uploadInfo, nil
}

// uploadContent uploads content to S3 and returns URI and presigned URL
func (a *Adapter) uploadContent(data []byte, key string, contentType string, logger *zap.Logger) (string, string, error) {
	logger.Debug("Uploading content to S3",
		zap.String("key", key),
		zap.String("content_type", contentType),
		zap.Int("size", len(data)),
	)

	// Upload to S3
	uri, err := a.storageClient.Upload(
		a.ctx,
		key,
		contentType,
		bytes.NewReader(data),
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	logger.Info("Content uploaded to S3",
		zap.String("uri", uri),
		zap.String("key", key),
		zap.Int("size", len(data)),
	)

	// Generate presigned URL (valid for 7 days)
	presignedURL, err := a.storageClient.GetPresignedURL(a.ctx, key, 10080)
	if err != nil {
		logger.Warn("Failed to generate presigned URL",
			zap.Error(err),
			zap.String("uri", uri),
		)
		return uri, "", nil
	}

	logger.Debug("Generated presigned URL",
		zap.String("uri", uri),
		zap.String("presigned_url", presignedURL),
		zap.String("expires_in", "7 days"),
	)

	return uri, presignedURL, nil
}

// downloadImage downloads an image from a URL discovered in SCRAPED PAGE
// CONTENT — not a URL this platform chose, so it is attacker-influenced by
// construction (bugs_open/159). Fetched via a.imageHTTPClient, which is
// fetchguard-wrapped: every dial (including a redirect target) is refused if
// it resolves to a private/loopback/link-local address, closing the SSRF
// path a scraped page could otherwise use to reach the pod's own cloud
// metadata endpoint or another service on the cluster network.
func (a *Adapter) downloadImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.imageHTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("image download failed with status: %d", resp.StatusCode)
	}

	// Capped read: a hostile or misconfigured origin that streams forever
	// must not be read into memory without bound. A response that hits the
	// cap is a FAILURE, not a partial image — storing a cut-off image would
	// look like a complete one to everything downstream, exactly the
	// truncation-looks-complete shape this platform has been burned by
	// before (bugs_open/012).
	data, truncated, err := fetchguard.LimitedRead(resp, fetchguard.DefaultConfig().MaxResponseBytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}
	if truncated {
		return nil, "", fmt.Errorf("image at %s exceeded the %d-byte cap — refusing a truncated image", imageURL, fetchguard.DefaultConfig().MaxResponseBytes)
	}

	// Get content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from data
		contentType = http.DetectContentType(data)
	}

	return data, contentType, nil
}

// getImageExtension returns file extension for content type
func getImageExtension(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/bmp":
		return ".bmp"
	case "image/tiff":
		return ".tiff"
	default:
		return ".jpg" // default
	}
}

// StartHealthServer starts the health check HTTP server
func (a *Adapter) StartHealthServer(port string) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Ready check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check if consumer is connected
		if a.consumer != nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready"}`))
		}
	})

	// Metrics endpoint (placeholder)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Add Prometheus metrics
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# HELP webscrape_adapter_info Webscrape adapter information\n"))
		w.Write([]byte("# TYPE webscrape_adapter_info gauge\n"))
		w.Write([]byte("webscrape_adapter_info{version=\"1.0.0\"} 1\n"))
	})

	a.healthServer = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := a.healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Health server error", zap.Error(err))
		}
	}()
}

// Shutdown gracefully shuts down the adapter
func (a *Adapter) Shutdown() {
	a.logger.Info("Shutting down webscrape adapter")

	// Cancel context to stop Run loop
	if a.cancel != nil {
		a.cancel()
	}

	// Shutdown health server
	if a.healthServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.healthServer.Shutdown(ctx); err != nil {
			a.logger.Error("Failed to shutdown health server", zap.Error(err))
		}
	}

	// Close Kafka connections
	if a.consumer != nil {
		a.consumer.Close()
	}
	if a.producer != nil {
		a.producer.Close()
	}

	// Close storage client
	/*	if a.storageClient != nil {
		a.storageClient.Close()
	}*/

	// Close HTTP client (if it has idle connections)
	if a.httpClient != nil {
		a.httpClient.CloseIdleConnections()
	}

	a.logger.Info("Webscrape adapter shutdown complete")
}
