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
	"github.com/gqls/agentchassis/internal/adapters/websearch/providers"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

const (
	requestTopic   = "system.adapter.web.search"
	responsesTopic = "system.agent.websearch.responses"
	consumerGroup  = "web-search-adapter-group"
)

// RequestPayload for web search
type RequestPayload struct {
	Action string `json:"action"`
	Data   struct {
		Query      string `json:"query"`
		NumResults int    `json:"num_results,omitempty"`
		SearchType string `json:"search_type,omitempty"` // web, news, images
		Provider   string `json:"provider,omitempty"`    // specific provider to use
	} `json:"data"`
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
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Initialize providers
	searchProviders := []providers.SearchProvider{
		providers.NewFirecrawlProvider(httpClient, logger),
		providers.NewScrapingBeeProvider(httpClient, logger),
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
			go a.handleMessage(msg)
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

	var req RequestPayload
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.Error("Failed to unmarshal request", zap.Error(err))
		a.sendErrorResponse(headers, "Invalid request format: "+err.Error())
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Validate request
	if strings.TrimSpace(req.Data.Query) == "" {
		l.Error("Empty search query")
		a.sendErrorResponse(headers, "Search query cannot be empty")
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

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

// performSearchWithFallback tries providers with fallback logic
func (a *Adapter) performSearchWithFallback(query string, numResults int, preferredProvider string) ([]providers.SearchResult, string, []string, error) {
	var fallbacks []string

	// If specific provider requested, try it first
	if preferredProvider != "" {
		for _, p := range a.providers {
			if p.Name() == preferredProvider {
				results, err := p.Search(a.ctx, query, numResults)
				if err == nil {
					return results, p.Name(), fallbacks, nil
				}
				a.logger.Warn("Preferred provider failed",
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
				results, err := p.Search(a.ctx, query, numResults)
				if err == nil {
					return results, p.Name(), fallbacks, nil
				}
				a.logger.Warn("Primary provider failed",
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

		results, err := p.Search(a.ctx, query, numResults)
		if err == nil {
			a.logger.Info("Search successful with fallback provider",
				zap.String("provider", p.Name()))
			return results, p.Name(), fallbacks, nil
		}

		a.logger.Error("Provider failed",
			zap.String("provider", p.Name()),
			zap.Error(err))
		fallbacks = append(fallbacks, p.Name())
	}

	return nil, "", fallbacks, fmt.Errorf("all %d providers failed", len(a.providers))
}

// sendResponse sends a successful response to the caller's topic
func (a *Adapter) sendResponse(headers map[string]string, payload ResponsePayload) {
	responseBytes, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"data":    payload,
	})

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

	responseHeaders := map[string]string{
		"correlation_id":            headers["correlation_id"],
		"causation_id":              headers["request_id"],
		"request_id":                uuid.NewString(),
		"message_type":              "response",
		"in_response_to_request_id": headers["request_id"],
		"in_response_to_step_name":  headers["step_name"],
		"in_response_to_step_id":    headers["step_id"],
		"orchestration_id":          headers["orchestration_id"],
		"parent_orchestration_id":   headers["parent_orchestration_id"],
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

// sendErrorResponse sends an error response to the caller's topic
func (a *Adapter) sendErrorResponse(headers map[string]string, errorMsg string) {
	payload := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}
	responseBytes, _ := json.Marshal(payload)

	// Determine where to send the response (same logic as sendResponse)
	responseTopic := responsesTopic // fallback to default
	if rt := headers["reply_to_topic"]; rt != "" {
		responseTopic = rt
	} else if rt := headers["responses_topic"]; rt != "" {
		responseTopic = rt
	} else if rt := headers["parent_responses_topic"]; rt != "" {
		responseTopic = rt
	}

	a.logger.Info("Sending search error response",
		zap.String("to_topic", responseTopic),
		zap.String("correlation_id", headers["correlation_id"]),
		zap.String("error", errorMsg),
	)

	responseHeaders := map[string]string{
		"correlation_id":            headers["correlation_id"],
		"causation_id":              headers["request_id"],
		"request_id":                uuid.NewString(),
		"message_type":              "response",
		"in_response_to_request_id": headers["request_id"],
		"in_response_to_step_name":  headers["step_name"],
		"in_response_to_step_id":    headers["step_id"],
		"orchestration_id":          headers["orchestration_id"],
		"parent_orchestration_id":   headers["parent_orchestration_id"],
		"status":                    "error",
		"is_complete":               "true",
		"is_error":                  "true",
	}

	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to produce error response",
			zap.Error(err),
			zap.String("topic", responseTopic))
	}
}
