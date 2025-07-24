// FILE: internal/adapters/websearch/adapter.go
package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

const (
	requestTopic  = "system.adapter.web.search"
	responseTopic = "system.responses.websearch"
	consumerGroup = "web-search-adapter-group"
)

// RequestPayload for web search
type RequestPayload struct {
	Action string `json:"action"`
	Data   struct {
		Query      string `json:"query"`
		NumResults int    `json:"num_results,omitempty"`
		SearchType string `json:"search_type,omitempty"` // web, news, images
	} `json:"data"`
}

// ResponsePayload with search results
type ResponsePayload struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	PublishedAt string `json:"published_at,omitempty"`
}

// Adapter handles web search requests
type Adapter struct {
	ctx          context.Context
	logger       *zap.Logger
	consumer     *kafka.Consumer
	producer     kafka.Producer
	httpClient   *http.Client
	apiKey       string
	searchAPIURL string
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

	apiKey := os.Getenv("SCRAPING_BEE_KEY")
	if apiKey == "" {
		consumer.Close()
		producer.Close()
		return nil, fmt.Errorf("SCRAPING_BEE_KEY not set")
	}

	return &Adapter{
		ctx:          ctx,
		logger:       logger,
		consumer:     consumer,
		producer:     producer,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		apiKey:       apiKey,
		searchAPIURL: "https://scrapingbee.com/search",
	}, nil
}

// Run starts the adapter's main loop
func (a *Adapter) Run() error {
	a.logger.Info("Web search adapter running")

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

// handleMessage processes a search request
func (a *Adapter) handleMessage(msg kafka.Message) {
	headers := kafka.HeadersToMap(msg.Headers)
	l := a.logger.With(zap.String("correlation_id", headers["correlation_id"]))

	var req RequestPayload
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.Error("Failed to unmarshal request", zap.Error(err))
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Perform the search
	results, err := a.performSearch(req.Data.Query, req.Data.NumResults)
	if err != nil {
		l.Error("Search failed", zap.Error(err))
		a.sendErrorResponse(headers, "Search failed: "+err.Error())
		a.consumer.CommitMessages(context.Background(), msg)
		return
	}

	// Send response
	response := ResponsePayload{
		Query:   req.Data.Query,
		Results: results,
		Total:   len(results),
	}

	a.sendResponse(headers, response)
	a.consumer.CommitMessages(context.Background(), msg)
}

// performSearch executes the actual web search
func (a *Adapter) performSearch(query string, numResults int) ([]SearchResult, error) {
	if numResults == 0 {
		numResults = 10
	}

	// Build search URL
	params := url.Values{}
	params.Add("q", query)
	params.Add("api_key", a.apiKey)
	params.Add("num", fmt.Sprintf("%d", numResults))
	params.Add("engine", "google")

	searchURL := fmt.Sprintf("%s?%s", a.searchAPIURL, params.Encode())

	// Execute request
	resp, err := a.httpClient.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		OrganicResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			Date    string `json:"date,omitempty"`
		} `json:"organic_results"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	// Convert to our format
	results := make([]SearchResult, 0, len(apiResponse.OrganicResults))
	for _, r := range apiResponse.OrganicResults {
		results = append(results, SearchResult{
			Title:       r.Title,
			URL:         r.Link,
			Snippet:     r.Snippet,
			PublishedAt: r.Date,
		})
	}

	return results, nil
}

// sendResponse sends a successful response
func (a *Adapter) sendResponse(headers map[string]string, payload ResponsePayload) {
	responseBytes, _ := json.Marshal(payload)
	responseHeaders := map[string]string{
		"correlation_id": headers["correlation_id"],
		"causation_id":   headers["request_id"],
		"request_id":     uuid.NewString(),
	}

	if err := a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes); err != nil {
		a.logger.Error("Failed to produce response", zap.Error(err))
	}
}

// sendErrorResponse sends an error response
func (a *Adapter) sendErrorResponse(headers map[string]string, errorMsg string) {
	payload := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}
	responseBytes, _ := json.Marshal(payload)
	responseHeaders := map[string]string{
		"correlation_id": headers["correlation_id"],
		"causation_id":   headers["request_id"],
		"request_id":     uuid.NewString(),
	}

	a.producer.Produce(a.ctx, responseTopic, responseHeaders,
		[]byte(headers["correlation_id"]), responseBytes)
}
