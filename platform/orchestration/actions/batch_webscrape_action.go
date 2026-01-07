// FILE: platform/orchestration/actions/batch_webscrape_action.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// BatchWebscrapeAction sends batch scrape requests to the webscrape adapter
//
// ==============================================================================
// REGISTRATION REQUIRED - Add to TWO places:
// ==============================================================================
//
// 1. LocalActions map (platform/orchestration/actions_list/local_actions.go):
//        "batch_webscrape": true,
//
// 2. GlobalActionRegistry (registry.go):
//        "batch_webscrape": BatchWebscrapeAction,
//
// ==============================================================================

const (
	webscrapeAdapterTopic = "system.adapter.webscrape.requests"
)

// BatchWebscrapeResult represents the result of initiating a batch scrape
type BatchWebscrapeResult struct {
	Success       bool                   `json:"success"`
	RequestID     string                 `json:"request_id"`
	TopicSentTo   string                 `json:"topic_sent_to"`
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// BatchWebscrapeAction sends a batch scrape request to the webscrape adapter
//
// Config options:
//   - urls_field: path to URLs array in collected_data (default: "prepared_urls.urls_to_scrape")
//   - scrape_config: config passed to scraping provider
//   - only_main_content: bool (default: true)
//   - capture_screenshot: bool (default: false)
//   - wait_for: int milliseconds to wait for JS
//
// Input: URLs from prepare_urls action
//   - prepared_urls.urls_to_scrape: array of {url, title, index}
//
// Output (stored at output_field, e.g., "scrape_results"):
//   - Adapter returns: {results: [...], success_count, error_count, total_count}
//   - Each result: {index, url, title, content, success, error?}
func BatchWebscrapeAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BatchWebscrapeAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get URLs to scrape from collected data
	// Default path: prepared_urls.urls_to_scrape (from PrepareUrlsAction)
	urlsField := "prepared_urls.urls_to_scrape"
	if uf, ok := config["urls_field"].(string); ok && uf != "" {
		urlsField = uf
	}

	urlsData := datahelpers.ExtractNestedField(params.CollectedData, urlsField)
	if urlsData == nil {
		params.Logger.Warn("BatchWebscrapeAction: No URLs to scrape",
			zap.String("urls_field", urlsField),
			zap.Strings("collected_data_keys", getMapKeys(params.CollectedData)))

		// Return empty result - not an error, just nothing to scrape
		return map[string]interface{}{
			"results":       []interface{}{},
			"success_count": 0,
			"error_count":   0,
			"total_count":   0,
		}, nil
	}

	// Convert to array
	var urls []interface{}
	switch v := urlsData.(type) {
	case []interface{}:
		urls = v
	case []map[string]interface{}:
		for _, u := range v {
			urls = append(urls, u)
		}
	default:
		params.Logger.Error("BatchWebscrapeAction: urls_to_scrape is not an array",
			zap.String("type", fmt.Sprintf("%T", urlsData)))
		return nil, fmt.Errorf("urls_to_scrape at '%s' is not an array (got %T)", urlsField, urlsData)
	}

	if len(urls) == 0 {
		params.Logger.Info("BatchWebscrapeAction: Empty URLs array")
		return map[string]interface{}{
			"results":       []interface{}{},
			"success_count": 0,
			"error_count":   0,
			"total_count":   0,
		}, nil
	}

	params.Logger.Info("BatchWebscrapeAction: Found URLs to scrape",
		zap.Int("count", len(urls)))

	// Get client ID
	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "default"
	}

	// Generate new request ID for this batch operation
	newRequestID := uuid.NewString()

	// Get response topic (where we want the adapter to reply)
	// Follow the established pattern from WebscrapeAction
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		myResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	if myResponsesTopic == "" {
		myResponsesTopic = fmt.Sprintf("system.agent.%s.responses", params.ExecutionContext.Sender.AgentType)
	}

	params.Logger.Info("BatchWebscrapeAction: Using responses topic",
		zap.String("responses_topic", myResponsesTopic))

	// Build scrape config
	scrapeConfig := map[string]interface{}{
		"only_main_content":  true,
		"capture_screenshot": false,
	}
	if sc, ok := config["scrape_config"].(map[string]interface{}); ok {
		for k, v := range sc {
			scrapeConfig[k] = v
		}
	}

	// Build adapter request - following established patterns from WebscrapeAction
	// Headers must include all required fields for validation
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			// Core message identification (required for validation)
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",

			// Sender information (flat fields, matching WebscrapeAction pattern)
			"sender_agent_type":    params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":      params.ExecutionContext.OrchestrationID,
			"sender_pod_name":      params.ExecutionContext.Sender.PodName,
			"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,
			"sender_role":          params.ExecutionContext.Sender.Role,

			// Response routing
			"responses_topic":        myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,

			// Additional metadata
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"action":    "batch_scrape",
		},
		"body": map[string]interface{}{
			"action": "batch_scrape",
			"data": map[string]interface{}{
				"urls":   urls,
				"config": scrapeConfig,
			},

			// Response routing in body as well (some adapters check here)
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,

			// Additional metadata for the adapter
			"metadata": map[string]interface{}{
				"requesting_agent_id":   params.ExecutionContext.OrchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
				"url_count":             len(urls),
			},

			// Include original request context
			"request_context": map[string]interface{}{
				"correlation_id":   params.ExecutionContext.CorrelationID,
				"orchestration_id": params.ExecutionContext.OrchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	// Convert headers to map[string]string for Kafka/validation
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string)
	for k, v := range rawHeaders {
		if str, ok := v.(string); ok {
			headers[k] = str
		} else {
			headers[k] = fmt.Sprintf("%v", v)
		}
	}

	// Marshal request
	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch scrape request: %w", err)
	}

	// Use correlation ID as message key for Kafka partitioning
	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("BatchWebscrapeAction: Sending request to adapter",
		zap.String("topic", webscrapeAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("reply_to_topic", myResponsesTopic),
		zap.Int("url_count", len(urls)),
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("client_id", clientID),
	)

	// Send to adapter using ProduceWithValidation (includes header validation)
	if err := params.Producer.ProduceWithValidation(
		ctx,
		webscrapeAdapterTopic,
		headers,
		key,
		messageBytes,
	); err != nil {
		return nil, fmt.Errorf("failed to send to webscrape adapter: %w", err)
	}

	// Return result indicating we're waiting for response
	return &BatchWebscrapeResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   webscrapeAdapterTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"url_count":       len(urls),
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
			"responses_topic": myResponsesTopic,
			"client_id":       clientID,
		},
	}, nil
}
