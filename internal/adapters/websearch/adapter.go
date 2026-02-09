// FILE: internal/adapters/websearch/adapter.go
package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/adapters/shared/throttle"
	"github.com/gqls/agentchassis/internal/adapters/websearch/providers"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

const (
	requestTopic     = "system.adapter.web.search"
	responsesTopic   = "system.agent.websearch.responses"
	consumerGroup    = "web-search-adapter-group"
	requestTimeout   = 120 * time.Second
	idleTimeout      = 90 * time.Second
	handshakeTimeout = 10 * time.Second

	// Retry configuration
	maxRetriesPerProvider = 4
	initialRetryBackoff   = 5 * time.Second
)

// MessageEnvelope wraps the body/headers structure sent by agents
type MessageEnvelope struct {
	Body    *RequestPayload        `json:"body,omitempty"`
	Headers map[string]interface{} `json:"headers,omitempty"`
}

// RequestPayload for web search
type RequestPayload struct {
	Action string `json:"action"`
	Data   struct {
		Query      string `json:"query"`
		NumResults int    `json:"num_results,omitempty"`
		SearchType string `json:"search_type,omitempty"` // web, news, images
		Provider   string `json:"provider,omitempty"`    // specific provider to use
	} `json:"data"`
	ReplyToTopic string `json:"reply_to_topic,omitempty"` // Sometimes included in body
}

// ResponsePayload with search results
type ResponsePayload struct {
	Query     string                   `json:"query"`
	Results   []providers.SearchResult `json:"results"`
	Total     int                      `json:"total"`
	Provider  string                   `json:"provider"`            // which provider was used
	Fallbacks []string                 `json:"fallbacks,omitempty"` // if any fallbacks were attempted
}

// Adapter handles web search requests
type Adapter struct {
	ctx             context.Context
	logger          *zap.Logger
	consumer        *kafka.Consumer
	producer        kafka.Producer
	providers       []providers.SearchProvider
	primaryProvider string
	httpClient      *http.Client
	throttle        *throttle.Throttle
}

// NewAdapter creates a new web search adapter
func NewAdapter(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Adapter, error) {
	consumer, err := kafka.NewConsumer(cfg.Infrastructure.KafkaBrokers, requestTopic, consumerGroup, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	producer, err := kafka.NewProducer(cfg.Infrastructure.KafkaBrokers, logger)
	if err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	httpClient := &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     idleTimeout,
			TLSHandshakeTimeout: handshakeTimeout,
		},
	}

	requestThrottle := throttle.New(logger)

	// Initialize providers
	searchProviders := []providers.SearchProvider{
		providers.NewFirecrawlProvider(httpClient, logger),
		providers.NewScrapingBeeProvider(httpClient, logger),
		providers.NewDuckDuckGoProvider(httpClient, logger),
	}

	// Filter available providers
	availableProviders := []providers.SearchProvider{}
	for _, p := range searchProviders {
		if p.IsAvailable() {
			logger.Info("Search provider available", zap.String("provider", p.Name()))
			availableProviders = append(availableProviders, p)
		} else {
			logger.Warn("Search provider not available (missing API key?)",
				zap.String("provider", p.Name()))
		}
	}

	if len(availableProviders) == 0 {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("no search providers available - check API keys")
	}

	primaryProvider := os.Getenv("PRIMARY_SEARCH_PROVIDER")
	if primaryProvider == "" {
		primaryProvider = availableProviders[0].Name()
	}

	return &Adapter{
		ctx:             ctx,
		logger:          logger,
		consumer:        consumer,
		producer:        producer,
		providers:       availableProviders,
		primaryProvider: primaryProvider,
		httpClient:      httpClient,
		throttle:        requestThrottle,
	}, nil
}

// Run starts the adapter's main loop
func (a *Adapter) Run() error {
	a.logger.Info("Web search adapter running",
		zap.Int("providers_count", len(a.providers)),
		zap.String("primary_provider", a.primaryProvider))

	for {
		select {
		case <-a.ctx.Done():
			a.logger.Info("Shutting down web search adapter")
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
				time.Sleep(time.Second) // Brief pause on error
				continue
			}
			// go a.handleMessage(msg)
			a.handleMessage(msg) // Sequential, not concurrent
			a.throttle.Wait()    // Delay before next request
		}
	}
}

// handleMessage processes a search request
func (a *Adapter) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("request_id", headers["request_id"]),
	)

	l.Info("Processing search request")

	// Debug: log the raw message to help diagnose issues
	l.Info("Raw message received",
		zap.String("raw_value", string(msg.Value)),
		zap.Int("value_length", len(msg.Value)),
	)

	// Extract the request payload - handle both envelope and direct formats
	req, err := a.extractRequestPayload(msg.Value, l)
	if err != nil {
		l.Error("Failed to extract request payload", zap.Error(err))
		a.sendErrorResponse(headers, "Invalid request format: "+err.Error())
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Debug: log the extracted query
	l.Info("Extracted request payload",
		zap.String("query", req.Data.Query),
		zap.Int("num_results", req.Data.NumResults),
		zap.String("search_type", req.Data.SearchType),
		zap.String("provider", req.Data.Provider),
	)

	// Validate request
	if strings.TrimSpace(req.Data.Query) == "" {
		l.Error("Empty search query",
			zap.String("raw_query", req.Data.Query),
			zap.String("action", req.Action),
		)
		a.sendErrorResponse(headers, "Search query cannot be empty")
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	l.Info("Executing search",
		zap.String("query", req.Data.Query),
		zap.Int("num_results", req.Data.NumResults),
	)

	// Perform search with fallback
	results, provider, fallbacks, err := a.performSearchWithFallback(
		req.Data.Query,
		req.Data.NumResults,
		req.Data.Provider,
	)

	if err != nil {
		l.Error("All search providers failed",
			zap.Error(err),
			zap.Strings("attempted_providers", fallbacks))
		a.sendErrorResponse(headers, "Search failed: "+err.Error())
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Send response
	response := ResponsePayload{
		Query:     req.Data.Query,
		Results:   results,
		Total:     len(results),
		Provider:  provider,
		Fallbacks: fallbacks,
	}

	a.sendResponse(headers, response)
	a.consumer.CommitMessages(context.Background(), msg)

	l.Info("Search request processed successfully",
		zap.String("provider_used", provider),
		zap.Int("results_count", len(results)))
}

// extractRequestPayload handles both envelope format (with body wrapper) and direct format
func (a *Adapter) extractRequestPayload(data []byte, l *zap.Logger) (*RequestPayload, error) {
	// First, try to parse as a generic map to inspect structure
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse message as JSON: %w", err)
	}

	l.Info("Parsed message structure",
		zap.Strings("top_level_keys", getMapKeys(raw)),
	)

	// Check if message has a "body" wrapper (envelope format from agents)
	if body, hasBody := raw["body"]; hasBody {
		l.Info("Message has body wrapper, extracting payload from body")

		bodyMap, ok := body.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("body field is not a map, got %T", body)
		}

		// Re-marshal the body and unmarshal into RequestPayload
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return nil, fmt.Errorf("failed to re-marshal body: %w", err)
		}

		var req RequestPayload
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			return nil, fmt.Errorf("failed to parse body as RequestPayload: %w", err)
		}

		l.Info("Successfully extracted from body wrapper",
			zap.String("query", req.Data.Query),
		)

		return &req, nil
	}

	// No body wrapper - try direct format
	l.Info("No body wrapper, trying direct format")

	var req RequestPayload
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse as direct RequestPayload: %w", err)
	}

	// If still no query, try to extract from nested data paths
	if req.Data.Query == "" {
		l.Info("Direct parse yielded empty query, trying fallback extraction")
		req.Data.Query = a.extractQueryFallback(raw, l)
	}

	return &req, nil
}

// extractQueryFallback attempts to find the query in various nested locations
func (a *Adapter) extractQueryFallback(data map[string]interface{}, l *zap.Logger) string {
	// Try common paths where query might be located
	paths := [][]string{
		{"data", "query"},
		{"body", "data", "query"},
		{"input", "query"},
		{"query"},
		{"search_query"},
		{"topic"},
	}

	for _, path := range paths {
		if val := getNestedValue(data, path); val != "" {
			l.Info("Found query via fallback path",
				zap.Strings("path", path),
				zap.String("query", val),
			)
			return val
		}
	}

	return ""
}

// getNestedValue extracts a string value from a nested map path
func getNestedValue(data map[string]interface{}, path []string) string {
	current := interface{}(data)

	for _, key := range path {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current, ok = m[key]
		if !ok {
			return ""
		}
	}

	if str, ok := current.(string); ok {
		return str
	}
	return ""
}

// getMapKeys returns the keys of a map for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// searchProviderWithRetry wraps a provider search with retry logic
func (a *Adapter) searchProviderWithRetry(ctx context.Context, provider providers.SearchProvider, query string, numResults int) ([]providers.SearchResult, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetriesPerProvider; attempt++ {
		// Create per-attempt timeout context
		attemptCtx, cancel := context.WithTimeout(ctx, requestTimeout)

		results, err := provider.Search(attemptCtx, query, numResults)
		cancel()

		if err == nil {
			if attempt > 1 {
				a.logger.Info("Search succeeded after retry",
					zap.String("provider", provider.Name()),
					zap.Int("attempt", attempt))
			}
			return results, nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			a.logger.Info("Non-retryable error, not retrying",
				zap.String("provider", provider.Name()),
				zap.Error(err))
			return nil, err
		}

		if attempt < maxRetriesPerProvider {
			// Exponential backoff: 5s, 10s, 20s
			backoff := initialRetryBackoff * time.Duration(1<<uint(attempt-1))

			a.logger.Warn("Provider failed with retryable error, will retry",
				zap.String("provider", provider.Name()),
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", maxRetriesPerProvider),
				zap.Duration("backoff", backoff),
				zap.Error(err))

			time.Sleep(backoff)
		}
	}

	return nil, lastErr
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())

	// Retryable conditions
	return strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "try again") ||
		strings.Contains(errStr, "context deadline") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "temporarily unavailable") ||
		strings.Contains(errStr, "too many requests")
}

// performSearchWithFallback tries providers with fallback and retry logic
func (a *Adapter) performSearchWithFallback(query string, numResults int, preferredProvider string) ([]providers.SearchResult, string, []string, error) {
	var fallbacks []string

	// If specific provider requested, try it first
	if preferredProvider != "" {
		for _, p := range a.providers {
			if p.Name() == preferredProvider {
				results, err := a.searchProviderWithRetry(a.ctx, p, query, numResults)
				if err == nil {
					return results, p.Name(), fallbacks, nil
				}
				a.logger.Warn("Preferred provider failed after retries",
					zap.String("provider", p.Name()),
					zap.Error(err))
				fallbacks = append(fallbacks, p.Name())
				break
			}
		}
	}

	// Try primary provider if not already attempted
	if preferredProvider != a.primaryProvider {
		for _, p := range a.providers {
			if p.Name() == a.primaryProvider {
				results, err := a.searchProviderWithRetry(a.ctx, p, query, numResults)
				if err == nil {
					return results, p.Name(), fallbacks, nil
				}
				a.logger.Warn("Primary provider failed after retries",
					zap.String("provider", p.Name()),
					zap.Error(err))
				fallbacks = append(fallbacks, p.Name())
				break
			}
		}
	}

	// Try all other providers
	for _, p := range a.providers {
		// Skip if already tried
		alreadyTried := false
		for _, f := range fallbacks {
			if f == p.Name() {
				alreadyTried = true
				break
			}
		}
		if alreadyTried {
			continue
		}

		results, err := a.searchProviderWithRetry(a.ctx, p, query, numResults)
		if err == nil {
			a.logger.Info("Search successful with fallback provider",
				zap.String("provider", p.Name()))
			return results, p.Name(), fallbacks, nil
		}

		a.logger.Error("Provider failed after retries",
			zap.String("provider", p.Name()),
			zap.Error(err))
		fallbacks = append(fallbacks, p.Name())
	}

	return nil, "", fallbacks, fmt.Errorf("all %d providers failed after retries", len(a.providers))
}

// Problem: sendResponse sends a flat JSON payload:
//   {"success": true, "results": [...], "query": "...", "total": 5}
//
// But the chassis deserializes into types.ResponseMessage which expects:
//   {"headers": {...}, "body": {"success": true, "body": <data>, "error": null}}
//
// The webscrape adapter already uses this envelope format (sendSuccessResponse at line 3656).
// The web search adapter needs to match it.
// sendResponse sends a successful response to the caller's topic
func (a *Adapter) sendResponse(headers map[string]string, payload ResponsePayload) {
	// Build the actual result data
	resultData := map[string]interface{}{
		"results":   payload.Results,
		"query":     payload.Query,
		"total":     payload.Total,
		"provider":  payload.Provider,
		"fallbacks": payload.Fallbacks,
	}

	// Wrap in the envelope format that the chassis expects (types.ResponseMessage)
	response := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":            headers["correlation_id"],
			"orchestration_id":          headers["orchestration_id"],
			"in_response_to_request_id": headers["request_id"],
			"in_response_to_step_name":  headers["step_name"],
			"in_response_to_step_id":    headers["step_id"],
			"status":                    "complete",
			"request_id":                headers["request_id"],
			"message_type":              "response",
			"timestamp":                 time.Now().UTC().Format(time.RFC3339),
			"success":                   true,
			"is_complete":               true,
			"client_id":                 headers["client_id"],
			"sender_agent_type":         "web-search-adapter",
			"parent_orchestration_id":   headers["parent_orchestration_id"],
			"sender": map[string]interface{}{
				"agent_type": "web-search-adapter",
				"agent_id":   "web-search-adapter-001",
				"pod_name":   os.Getenv("HOSTNAME"),
			},
		},
		"body": map[string]interface{}{
			"success": true,
			"body":    resultData,
			"error":   nil,
			"metadata": map[string]interface{}{
				"processed_at": time.Now().UTC().Format(time.RFC3339),
				"adapter":      "web-search",
			},
		},
	}

	responseBytes, _ := json.Marshal(response)

	// Determine where to send the response
	// Priority: reply_to_topic > responses_topic > parent_responses_topic > default
	responseTopic := responsesTopic // fallback to default
	if rt := headers["reply_to_topic"]; rt != "" {
		responseTopic = rt
	} else if rt := headers["responses_topic"]; rt != "" {
		responseTopic = rt
	} else if rt := headers["parent_responses_topic"]; rt != "" {
		responseTopic = rt
	}

	a.logger.Info("Sending search response",
		zap.String("to_topic", responseTopic),
		zap.String("correlation_id", headers["correlation_id"]),
		zap.Int("results_count", payload.Total),
	)

	// NOTE: Keep Kafka headers the same as before - they are used for routing
	responseHeaders := map[string]string{
		"correlation_id":            headers["correlation_id"],
		"causation_id":              headers["request_id"],
		"request_id":                uuid.NewString(),
		"client_id":                 headers["client_id"],
		"message_type":              "response",
		"in_response_to_request_id": headers["request_id"],
		"in_response_to_step_name":  headers["step_name"],
		"in_response_to_step_id":    headers["step_id"],
		"orchestration_id":          headers["orchestration_id"],
		"parent_orchestration_id":   headers["parent_orchestration_id"],
		"from_agent_type":           headers["from_agent_type"],
		"sender_agent_type":         headers["sender_agent_type"],
		"status":                    "complete",
		"is_complete":               "true",
	}

	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to produce response",
			zap.Error(err),
			zap.String("topic", responseTopic))
	}
}
