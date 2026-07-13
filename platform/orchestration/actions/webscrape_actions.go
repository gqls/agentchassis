// internal/backend/agent-chassis/platform/orchestration/actions/webscrape_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// WebscrapeResult represents the result of a webscraping operation
type WebscrapeResult struct {
	Success       bool                   `json:"success"`
	RequestID     string                 `json:"request_id"`
	TopicSentTo   string                 `json:"topic_sent_to"`
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// WebscrapeAction sends a scraping request with optional S3 upload
func WebscrapeAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing WebscrapeAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.Any("DEBUGaa: where is the responses topic in the execcontext?", params.ExecutionContext),
	)

	// Extract configuration
	config := params.StepConfig.Config

	// Determine action type (scrape, crawl, extract)
	action := "scrape"
	if a, ok := config["action"].(string); ok {
		action = a
	}

	// Check if we should upload results to S3
	uploadResults := false
	if upload, ok := config["upload_results"].(bool); ok {
		uploadResults = upload
	}

	// Get client ID for organized storage
	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "default"
	}

	// Extract URL from input data
	// Support url_field config for flexible URL location
	url := ""

	// First check if url_field is specified in config
	if urlField, ok := config["url_field"].(string); ok && urlField != "" {
		// Use datahelpers to extract from any path
		url = datahelpers.ExtractNestedFieldString(params.CollectedData, urlField)
	}

	// Fallback to direct url in config
	if url == "" {
		if directURL, ok := config["url"].(string); ok {
			url = directURL
		}
	}

	// Fallback to original behavior: input_data.target_url
	if url == "" {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if targetURL, ok := inputData["target_url"].(string); ok {
				url = targetURL
			}
			// Also check input_data.url (common alternative)
			if url == "" {
				if u, ok := inputData["url"].(string); ok {
					url = u
				}
			}
		}
	}

	if url == "" {
		return nil, fmt.Errorf("URL not found - check 'url_field', 'url' config, or input_data.target_url/url")
	}

	// The above allows workflows to use:
	//
	// Option 1: Default behavior (input_data.target_url)
	// {
	//     "action": "firecrawl_scrape",
	//     "config": {}
	// }
	//
	// Option 2: Custom field path
	// {
	//     "action": "firecrawl_scrape",
	//     "config": {
	//         "url_field": "input_data.url"
	//     }
	// }
	//
	// Option 3: Direct URL
	// {
	//     "action": "firecrawl_scrape",
	//     "config": {
	//         "url": "https://example.com"
	//     }
	// }

	// Generate new request ID for this operation
	newRequestID := uuid.NewString()

	// Topic configuration
	webscrapeAdapterTopic := "system.adapter.webscrape.requests"

	// Get response topic (where we want the adapter to reply)
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		// First try RESPONSES_TOPIC from environment (for job agents)
		myResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	if myResponsesTopic == "" {
		// Last resort: use generic agent type topic
		myResponsesTopic = fmt.Sprintf("system.agent.%s.responses", params.ExecutionContext.Sender.AgentType)
	}

	params.Logger.Info("Using responses topic",
		zap.String("responses_topic", myResponsesTopic),
		zap.String("from", "execution_context"))

	// Build the scraping data payload
	scrapeData := map[string]interface{}{
		"url":            url,
		"action":         action,
		"upload_results": uploadResults,
		"client_id":      clientID,
	}

	// Add any additional config from the step
	if scrapeConfig, ok := config["scrape_config"].(map[string]interface{}); ok {
		scrapeData["config"] = scrapeConfig
	}

	// Build comprehensive adapter request matching framework pattern
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			// Core message identification
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",

			// Sender information (flat fields, not nested!)
			"sender_agent_type":    params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":      params.ExecutionContext.OrchestrationID,
			"sender_pod_name":      params.ExecutionContext.Sender.PodName,
			"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,
			"sender_role":          params.ExecutionContext.Sender.Role,

			// Response routing
			"responses_topic":        myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,

			// Additional metadata
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"action":     action,
			"target_url": url,
		},
		"body": map[string]interface{}{
			"action": action,
			"data":   scrapeData,

			// Response routing in body as well (some adapters check here)
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,

			// Additional metadata for the adapter
			"metadata": map[string]interface{}{
				"requesting_agent_id":   params.ExecutionContext.OrchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
				"upload_enabled":        uploadResults,
			},

			// Include original request context
			"request_context": map[string]interface{}{
				"correlation_id":   params.ExecutionContext.CorrelationID,
				"orchestration_id": params.ExecutionContext.OrchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	// Convert headers to map[string]string for validation
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string)

	for k, v := range rawHeaders {
		if str, ok := v.(string); ok {
			headers[k] = str
		} else {
			headers[k] = fmt.Sprintf("%v", v) // fallback stringify for non-string values
		}
	}

	// Convert entire request to JSON
	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal adapter request: %w", err)
	}

	// Use correlation ID as message key for Kafka partitioning
	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("still in WebscrapeAction Sending request to webscrape adapter",
		zap.String("topic", webscrapeAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("reply_to_topic", myResponsesTopic),
		zap.String("action", action),
		zap.String("url", url),
		zap.Bool("upload_results", uploadResults),
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("client_id", clientID),
		zap.Int("message_size", len(messageBytes)),
		zap.Any("DEBUGaa: headers", headers),
		zap.Any("DEBUGaa: Adapter request", adapterRequest),
	)

	// Log the full request for debugging (be careful in production)
	params.Logger.Debug("Webscrape adapter request details",
		zap.Any("request", adapterRequest))

	// Send the message with validation
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
	return &WebscrapeResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   webscrapeAdapterTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"url":             url,
			"action":          action,
			"upload_results":  uploadResults,
			"client_id":       clientID,
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
			"responses_topic": myResponsesTopic,
		},
	}, nil
}

// FirecrawlScrapeAction wraps WebscrapeAction for single page scraping
func FirecrawlScrapeAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("in FirecrawlScrapeAction",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Set the action to "scrape" in config
	if params.StepConfig.Config == nil {
		params.StepConfig.Config = make(map[string]interface{})
	}
	params.StepConfig.Config["action"] = "scrape"

	// Call the main WebscrapeAction
	return WebscrapeAction(ctx, params)
}

// FirecrawlCrawlAction wraps WebscrapeAction for multi-page crawling
func FirecrawlCrawlAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("in FirecrawlCrawlAction",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Set the action to "crawl" in config
	if params.StepConfig.Config == nil {
		params.StepConfig.Config = make(map[string]interface{})
	}
	params.StepConfig.Config["action"] = "crawl"

	// Call the main WebscrapeAction
	return WebscrapeAction(ctx, params)
}

// FirecrawlExtractAction wraps WebscrapeAction for structured extraction
func FirecrawlExtractAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("in FirecrawlExtractAction",
		zap.String("step_name", params.ExecutionContext.StepName))

	// Set the action to "extract" in config
	if params.StepConfig.Config == nil {
		params.StepConfig.Config = make(map[string]interface{})
	}
	params.StepConfig.Config["action"] = "extract"

	// Call the main WebscrapeAction
	return WebscrapeAction(ctx, params)
}

// ValidateURLAction validates and normalizes URLs before scraping
func ValidateURLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing ValidateURLAction",
		zap.String("step_name", params.ExecutionContext.StepName))

	config := params.StepConfig.Config
	urlField := "target_url"
	if uf, ok := config["url_field"].(string); ok {
		urlField = uf
	}

	// Get URL from input data
	var url string
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if u, ok := inputData[urlField].(string); ok {
			url = u
		}
	}

	if url == "" {
		return nil, fmt.Errorf("URL field '%s' not found in input_data", urlField)
	}

	// Add protocol if missing
	if addProtocol, ok := config["add_protocol_if_missing"].(bool); ok && addProtocol {
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			url = "https://" + url
		}
	}

	// Update the input data with normalized URL
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		inputData[urlField] = url
		params.CollectedData["input_data"] = inputData
	}

	params.Logger.Debug("URL validated - returning",
		zap.String("original_field", urlField),
		zap.String("normalized_url", url))

	return map[string]interface{}{
		"validated_url": url,
		"url_field":     urlField,
		"success":       true,
	}, nil
}
