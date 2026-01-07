// FILE: internal/adapters/webscrape/batch_handler.go
package webscrape

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

// Batch scraping handler for webscrape adapter
//
// This file adds batch_scrape action support to the existing webscrape adapter.
// Add to handleMessage switch statement in adapter.go:
//
//   case "batch_scrape":
//       a.handleBatchScrape(ctx, headers, body, replyToTopic, l)
//       a.consumer.CommitMessages(context.Background(), msg)
//       return

// handleBatchScrape processes a batch scrape request
// Scrapes multiple URLs and returns combined results
func (a *Adapter) handleBatchScrape(
	ctx context.Context,
	headers map[string]interface{},
	body map[string]interface{},
	replyToTopic string,
	logger *zap.Logger,
) {
	startTime := time.Now()

	// Extract header fields for response
	requestID, _ := headers["request_id"].(string)
	correlationID, _ := headers["correlation_id"].(string)
	orchestrationID, _ := headers["orchestration_id"].(string)
	clientID, _ := headers["client_id"].(string)
	stepName, _ := headers["step_name"].(string)
	stepID, _ := headers["step_id"].(string)

	// Extract data from body
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		a.sendBatchErrorResponse(requestID, correlationID, orchestrationID, replyToTopic,
			clientID, stepName, stepID, "Missing data in request body")
		return
	}

	// Get URLs array
	urlsData, ok := data["urls"].([]interface{})
	if !ok {
		a.sendBatchErrorResponse(requestID, correlationID, orchestrationID, replyToTopic,
			clientID, stepName, stepID, "Missing or invalid 'urls' array in data")
		return
	}

	if len(urlsData) == 0 {
		// Empty array is valid - return empty results
		a.sendBatchSuccessResponse(requestID, correlationID, orchestrationID, replyToTopic,
			clientID, stepName, stepID, map[string]interface{}{
				"results":       []interface{}{},
				"success_count": 0,
				"error_count":   0,
				"total_count":   0,
			})
		return
	}

	// Get scrape config
	scrapeConfig, _ := data["config"].(map[string]interface{})
	if scrapeConfig == nil {
		scrapeConfig = make(map[string]interface{})
	}

	// Default config
	if _, exists := scrapeConfig["only_main_content"]; !exists {
		scrapeConfig["only_main_content"] = true
	}
	if _, exists := scrapeConfig["capture_screenshot"]; !exists {
		scrapeConfig["capture_screenshot"] = false
	}

	logger.Info("Processing batch scrape request",
		zap.String("request_id", requestID),
		zap.Int("url_count", len(urlsData)))

	// Get provider (default to firecrawl)
	providerName := "firecrawl"
	if p, ok := scrapeConfig["provider"].(string); ok {
		providerName = p
	}

	provider, ok := a.providers[providerName]
	if !ok {
		a.sendBatchErrorResponse(requestID, correlationID, orchestrationID, replyToTopic,
			clientID, stepName, stepID, fmt.Sprintf("Provider %s not available", providerName))
		return
	}

	// Process each URL
	results := make([]map[string]interface{}, 0, len(urlsData))
	successCount := 0
	errorCount := 0

	for i, urlItem := range urlsData {
		var url, title string
		var index int = i

		// Handle different URL formats
		switch v := urlItem.(type) {
		case string:
			url = v
			title = fmt.Sprintf("Source %d", i)
		case map[string]interface{}:
			url, _ = v["url"].(string)
			title, _ = v["title"].(string)
			if idx, ok := v["index"].(float64); ok {
				index = int(idx)
			}
			if title == "" {
				title = fmt.Sprintf("Source %d", index)
			}
		default:
			logger.Warn("Invalid URL item format",
				zap.Int("index", i),
				zap.String("type", fmt.Sprintf("%T", urlItem)))
			continue
		}

		if url == "" {
			logger.Warn("Empty URL, skipping",
				zap.Int("index", index))
			results = append(results, map[string]interface{}{
				"index":   index,
				"url":     "",
				"title":   title,
				"success": false,
				"error":   "Empty URL",
			})
			errorCount++
			continue
		}

		logger.Info("Scraping URL",
			zap.Int("index", index),
			zap.String("url", url),
			zap.String("title", title))

		// Call provider's Scrape method
		scrapeResult, err := provider.Scrape(ctx, url, scrapeConfig)

		if err != nil {
			logger.Warn("Failed to scrape URL",
				zap.Int("index", index),
				zap.String("url", url),
				zap.Error(err))

			results = append(results, map[string]interface{}{
				"index":   index,
				"url":     url,
				"title":   title,
				"success": false,
				"error":   err.Error(),
			})
			errorCount++
			continue
		}

		// Extract content from scrape result
		content := ""
		if markdown, ok := scrapeResult["markdown_content"].(string); ok && markdown != "" {
			content = markdown
		} else if cleanContent, ok := scrapeResult["clean_content"].(string); ok && cleanContent != "" {
			content = cleanContent
		} else if html, ok := scrapeResult["html_content"].(string); ok && html != "" {
			// Use HTML as fallback - could strip tags here if needed
			content = html
		}

		// Extract metadata
		pageTitle := title
		if t, ok := scrapeResult["title"].(string); ok && t != "" {
			pageTitle = t
		}

		pageDescription := ""
		if d, ok := scrapeResult["description"].(string); ok {
			pageDescription = d
		}

		results = append(results, map[string]interface{}{
			"index":       index,
			"url":         url,
			"title":       pageTitle,
			"description": pageDescription,
			"content":     content,
			"success":     true,
		})
		successCount++
	}

	logger.Info("Batch scrape completed",
		zap.Int("total", len(urlsData)),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount),
		zap.Duration("duration", time.Since(startTime)))

	// Send success response with results
	a.sendBatchSuccessResponse(requestID, correlationID, orchestrationID, replyToTopic,
		clientID, stepName, stepID, map[string]interface{}{
			"results":       results,
			"success_count": successCount,
			"error_count":   errorCount,
			"total_count":   len(urlsData),
		})
}

// sendBatchSuccessResponse sends a successful batch scrape response
// Follows the exact pattern from sendSuccessResponse in adapter.go
func (a *Adapter) sendBatchSuccessResponse(
	requestID, correlationID, orchestrationID, replyTopic,
	clientID, stepName, stepID string,
	result map[string]interface{},
) {
	if replyTopic == "" {
		a.logger.Warn("No reply topic specified for batch response",
			zap.String("request_id", requestID))
		return
	}

	// Response body - flat structure matching web_search adapter pattern
	// The entire body becomes collected_data[output_field]
	responseBody := map[string]interface{}{
		"success":       true,
		"results":       result["results"],
		"success_count": result["success_count"],
		"error_count":   result["error_count"],
		"total_count":   result["total_count"],
	}

	// Full response with headers and body
	response := map[string]interface{}{
		"headers": map[string]interface{}{
			// Required fields for validation
			"correlation_id":   correlationID,
			"orchestration_id": orchestrationID,
			"message_type":     "response",
			"status":           "complete",

			// Response tracking
			"in_response_to_request_id": requestID,
			"in_response_to_step_name":  stepName,
			"in_response_to_step_id":    stepID,
			"request_id":                requestID,
			"causation_id":              requestID,

			// Client ID - must echo back for validation
			"client_id": clientID,

			// Sender info
			"sender_agent_type": "webscrape-adapter",
			"sender": map[string]interface{}{
				"agent_type": "webscrape-adapter",
				"agent_id":   "webscrape-adapter-001",
				"pod_name":   os.Getenv("HOSTNAME"),
			},

			// Metadata
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"success":     true,
			"is_complete": "true",
		},
		"body": responseBody,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		a.logger.Error("Failed to marshal batch response",
			zap.Error(err),
			zap.String("request_id", requestID))
		return
	}

	// Create headers map for Kafka - must include all required fields
	headers := map[string]string{
		"correlation_id":            correlationID,
		"orchestration_id":          orchestrationID,
		"message_type":              "response",
		"status":                    "complete",
		"in_response_to_request_id": requestID,
		"in_response_to_step_name":  stepName,
		"in_response_to_step_id":    stepID,
		"request_id":                requestID,
		"causation_id":              requestID,
		"client_id":                 clientID,
		"sender_agent_type":         "webscrape-adapter",
		"is_complete":               "true",
	}

	a.logger.Info("Sending batch scrape success response",
		zap.String("request_id", requestID),
		zap.String("reply_topic", replyTopic),
		zap.Int("result_count", len(result["results"].([]map[string]interface{}))))

	if err := a.producer.ProduceWithValidation(
		a.ctx,
		replyTopic,
		headers,
		[]byte(correlationID),
		responseBytes,
	); err != nil {
		a.logger.Error("Failed to produce batch response",
			zap.Error(err),
			zap.String("topic", replyTopic))
	}
}

// sendBatchErrorResponse sends an error response for batch scrape
// Follows the exact pattern from sendErrorResponse in adapter.go
func (a *Adapter) sendBatchErrorResponse(
	requestID, correlationID, orchestrationID, replyTopic,
	clientID, stepName, stepID, errorMsg string,
) {
	if replyTopic == "" {
		a.logger.Warn("No reply topic specified for batch error response",
			zap.String("request_id", requestID))
		return
	}

	// Error response body
	responseBody := map[string]interface{}{
		"success":       false,
		"results":       []interface{}{},
		"success_count": 0,
		"error_count":   0,
		"total_count":   0,
		"error": map[string]interface{}{
			"message":     errorMsg,
			"code":        "BATCH_SCRAPE_ERROR",
			"recoverable": true,
		},
	}

	response := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":   correlationID,
			"orchestration_id": orchestrationID,
			"message_type":     "response",
			"status":           "error_recoverable",

			"in_response_to_request_id": requestID,
			"in_response_to_step_name":  stepName,
			"in_response_to_step_id":    stepID,
			"request_id":                requestID,
			"causation_id":              requestID,

			"client_id": clientID,

			"sender_agent_type": "webscrape-adapter",
			"sender": map[string]interface{}{
				"agent_type": "webscrape-adapter",
				"agent_id":   "webscrape-adapter-001",
				"pod_name":   os.Getenv("HOSTNAME"),
			},

			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"success":   false,
			"is_error":  "true",
		},
		"body": responseBody,
	}

	responseBytes, _ := json.Marshal(response)

	headers := map[string]string{
		"correlation_id":            correlationID,
		"orchestration_id":          orchestrationID,
		"message_type":              "response",
		"status":                    "error_recoverable",
		"in_response_to_request_id": requestID,
		"in_response_to_step_name":  stepName,
		"in_response_to_step_id":    stepID,
		"request_id":                requestID,
		"causation_id":              requestID,
		"client_id":                 clientID,
		"sender_agent_type":         "webscrape-adapter",
		"is_error":                  "true",
	}

	a.logger.Error("Sending batch scrape error response",
		zap.String("request_id", requestID),
		zap.String("error", errorMsg))

	if err := a.producer.ProduceWithValidation(
		a.ctx,
		replyTopic,
		headers,
		[]byte(correlationID),
		responseBytes,
	); err != nil {
		a.logger.Error("Failed to produce batch error response",
			zap.Error(err),
			zap.String("topic", replyTopic))
	}
}
