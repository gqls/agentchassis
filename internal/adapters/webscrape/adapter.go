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
	"github.com/gqls/agentchassis/internal/adapters/webscrape/providers"
	"github.com/gqls/agentchassis/platform/config"
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
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *zap.Logger
	consumer      *kafka.Consumer
	producer      kafka.Producer
	providers     map[string]providers.ScrapingProvider
	storageClient storage.Client
	httpClient    *http.Client
	config        *config.ServiceConfig
	healthServer  *http.Server
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
		ctx:           adapterCtx,
		cancel:        cancel,
		logger:        logger,
		consumer:      consumer,
		producer:      producer,
		providers:     scrapingProviders,
		storageClient: storageClient,
		httpClient:    httpClient,
		config:        cfg,
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
			go a.handleMessage(msg)
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

	// Validate required fields
	if url == "" {
		l.Error("Empty URL in request")
		a.sendErrorResponse(requestID, correlationID, orchestrationID, replyToTopic, "URL cannot be empty")
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
		a.sendErrorResponse(requestID, correlationID, orchestrationID, replyToTopic,
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
	case "extract":
		schema, ok := scrapeConfig["schema"].(map[string]interface{})
		if !ok {
			err = fmt.Errorf("schema not provided for extract action")
		} else {
			result, err = provider.ExtractStructured(a.ctx, url, schema, scrapeConfig)
		}
	default:
		err = fmt.Errorf("unknown action: %s", action)
	}

	if err != nil {
		l.Error("Action failed", zap.Error(err))
		a.sendErrorResponse(requestID, correlationID, orchestrationID, replyToTopic, err.Error())
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Upload results to S3 if requested
	if uploadResults && a.storageClient != nil {
		uploadedResult, err := a.uploadScrapingResults(result, req, l)
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

	// Send success response
	a.sendSuccessResponse(requestID, correlationID, orchestrationID, replyToTopic, result)
	a.consumer.CommitMessages(context.Background(), msg)

	l.Info("Request processed successfully",
		zap.Duration("duration", time.Since(startTime)))
}

// Updated response methods to handle new parameters
func (a *Adapter) sendSuccessResponse(requestID, correlationID, orchestrationID, replyTopic string, result interface{}) {
	if replyTopic == "" {
		a.logger.Warn("No reply topic specified", zap.String("request_id", requestID))
		return
	}

	response := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":   correlationID,
			"orchestration_id": orchestrationID,
			"request_id":       requestID,
			"message_type":     "response",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
			"success":          true,
		},
		"body": map[string]interface{}{
			"success": true,
			"data":    result,
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
	headers["request_id"] = requestID
	headers["orchestration_id"] = orchestrationID
	headers["message_type"] = "response"

	if err := a.producer.ProduceWithValidation(
		a.ctx,
		replyTopic,
		headers,
		[]byte(correlationID),
		responseBytes,
	); err != nil {
		a.logger.Error("Failed to produce response",
			zap.Error(err),
			zap.String("topic", replyTopic))
	}
}

func (a *Adapter) sendErrorResponse(requestID, correlationID, orchestrationID, replyTopic, errorMsg string) {
	if replyTopic == "" {
		a.logger.Warn("No reply topic specified for error response", zap.String("request_id", requestID))
		return
	}

	response := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":   correlationID,
			"orchestration_id": orchestrationID,
			"request_id":       requestID,
			"message_type":     "response",
			"timestamp":        time.Now().UTC().Format(time.RFC3339),
			"success":          false,
		},
		"body": map[string]interface{}{
			"success": false,
			"error":   errorMsg,
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
	headers["request_id"] = requestID
	headers["orchestration_id"] = orchestrationID
	headers["message_type"] = "response"
	headers["success"] = "false"

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

	// Upload screenshot if it's base64 encoded
	if screenshot, ok := resultMap["screenshot_base64"].(string); ok && screenshot != "" {
		// Decode base64 screenshot
		screenshotData, err := base64.StdEncoding.DecodeString(screenshot)
		if err == nil {
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
			}
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
