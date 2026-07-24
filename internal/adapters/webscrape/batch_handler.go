// FILE: internal/adapters/webscrape/batch_handler.go
package webscrape

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

		// Extract metadata
		pageTitle := title
		if t, ok := scrapeResult["title"].(string); ok && t != "" {
			pageTitle = t
		}

		pageDescription := ""
		if d, ok := scrapeResult["description"].(string); ok {
			pageDescription = d
		}

		// Build the result LEAN by default (bugs_open/062): the response
		// travels as ONE Kafka message for the whole batch, and the broker
		// refuses anything over max.message.bytes (~1 MiB) — a refusal this
		// handler used to swallow, starving the caller through its full
		// retry budget. So: one content field, not four. `content` carries
		// markdown (html fallback); `raw_html`/`html_content` ship ONLY
		// when the request opts in via include_raw_html, and no current
		// caller does.
		pageResult := map[string]interface{}{
			"index":       index,
			"url":         url,
			"title":       pageTitle,
			"description": pageDescription,
			"success":     true,
		}

		content := ""
		if md, ok := scrapeResult["markdown_content"].(string); ok && md != "" {
			content = md
			pageResult["markdown"] = md
		} else if html, ok := scrapeResult["html_content"].(string); ok && html != "" {
			content = html
		}
		pageResult["content"] = content

		if includeRaw, _ := scrapeConfig["include_raw_html"].(bool); includeRaw {
			if rawHTML, ok := scrapeResult["raw_html"].(string); ok && rawHTML != "" {
				pageResult["raw_html"] = rawHTML
			}
			if html, ok := scrapeResult["html_content"].(string); ok && html != "" {
				pageResult["html_content"] = html
			}
		}
		if meta, ok := scrapeResult["metadata"].(map[string]interface{}); ok {
			pageResult["metadata"] = meta
		}

		truncateBatchResult(pageResult, batchResultContentCap)

		results = append(results, pageResult)
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

// batchResultContentCap bounds one result's content within the batch reply.
// The whole batch travels as ONE Kafka message against a ~1 MiB broker cap
// (bugs_open/062); 150 KiB × a 3–5 URL batch leaves comfortable envelope
// headroom. The cut is VISIBLE — `truncated: true` on the result — because
// an invisible cut is the damage (the 012 truncation family).
const batchResultContentCap = 150 * 1024

// oversizeStripContentCap is the drastic per-result cap for the one retry
// after the broker refuses a reply as too large.
const oversizeStripContentCap = 8 * 1024

// truncateBatchResult caps a result's content-bearing fields in place,
// marking the result truncated if anything was actually cut.
func truncateBatchResult(result map[string]interface{}, capBytes int) {
	truncated := false
	for _, key := range []string{"content", "markdown", "raw_html", "html_content"} {
		if s, ok := result[key].(string); ok && len(s) > capBytes {
			result[key] = s[:capBytes]
			truncated = true
		}
	}
	if truncated {
		result["truncated"] = true
	}
}

// stripBatchResultsForRetry applies the drastic cap to every result, for the
// single resend after an oversize refusal. Raw HTML fields are dropped
// outright — if the reply did not fit WITH them, they are what did not fit.
func stripBatchResultsForRetry(results []map[string]interface{}) {
	for _, r := range results {
		delete(r, "raw_html")
		delete(r, "html_content")
		truncateBatchResult(r, oversizeStripContentCap)
	}
}

// isKafkaMessageTooLarge reports whether a produce error is the broker's
// message-size refusal — the one produce failure that is deterministic
// (resending the same bytes can never succeed), so it gets a degrade-and-
// retry rather than being surfaced as-is.
func isKafkaMessageTooLarge(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Message Size Too Large")
}

// sendBatchSuccessResponse sends a successful batch scrape response
// Follows the exact pattern from sendSuccessResponse in adapter.go
//
// bugs_open/062: a reply the broker refuses as too large is degraded
// (raw HTML dropped, content cut to a stub, truncated markers on) and
// resent ONCE; if even the stub reply is refused, an error response goes
// out instead. A response that cannot be delivered must become a
// deliverable error, never silence — the caller is listening on the reply
// topic, not reading this pod's logs, and the silent drop starved callers
// through 4 × 180s of retries on a failure that is deterministic.
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
		zap.Int("response_bytes", len(responseBytes)),
		zap.Int("result_count", len(result["results"].([]map[string]interface{}))))

	err = a.producer.ProduceWithValidation(
		a.ctx,
		replyTopic,
		headers,
		[]byte(correlationID),
		responseBytes,
	)
	if err == nil {
		return
	}
	if !isKafkaMessageTooLarge(err) {
		// Transient produce failures (broker unreachable etc.) keep the old
		// behaviour: the coordinator's retry resends the request and the
		// next attempt may deliver.
		a.logger.Error("Failed to produce batch response",
			zap.Error(err),
			zap.String("topic", replyTopic))
		return
	}

	// Deterministic refusal: degrade hard and resend once.
	results, _ := result["results"].([]map[string]interface{})
	stripBatchResultsForRetry(results)
	responseBody["results"] = results
	strippedBytes, merr := json.Marshal(response)
	if merr == nil {
		a.logger.Warn("Batch response exceeded broker max message size — resending with stripped content",
			zap.String("request_id", requestID),
			zap.Int("original_bytes", len(responseBytes)),
			zap.Int("stripped_bytes", len(strippedBytes)))
		if perr := a.producer.ProduceWithValidation(
			a.ctx, replyTopic, headers, []byte(correlationID), strippedBytes,
		); perr == nil {
			return
		} else {
			err = perr
		}
	}

	// Even the stub would not fit (or re-marshal failed): the caller gets a
	// real error now instead of a timeout in 12 minutes.
	a.logger.Error("Batch response undeliverable even after stripping — sending error response",
		zap.Error(err),
		zap.String("request_id", requestID),
		zap.String("topic", replyTopic))
	a.sendBatchErrorResponse(requestID, correlationID, orchestrationID, replyTopic,
		clientID, stepName, stepID,
		fmt.Sprintf("batch scrape succeeded but the response could not be delivered: %v", err))
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
