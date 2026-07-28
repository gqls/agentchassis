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

// WebscrapeInputSpec declares what scrape_web actually reads.
//
// Declaring ConfigKeys opts this action into unknown-config-key detection
// (bugs_open/101): a step carrying a key that is not listed here is reported by
// the workflow validator instead of being silently ignored. That is the whole
// point of the bug — four keys sat in two live definitions describing a crawl
// that never happened, and nothing could tell them apart from working config.
//
// StrictConfig stays FALSE for now, deliberately. Turning it on makes an unknown
// key a hard validation failure, and two live definitions
// (vet-practice-verifier, domain-research-classifier) would have to be corrected
// first — one of which has no identified owner. Warn now, and let the coverage
// report drive the correction; flipping this before the definitions are clean
// would break running agents to make a point about their config.
var WebscrapeInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{},
	ConfigKeys: []string{
		// URL resolution
		"url_field",
		"fallback_url_field",
		"url",
		// Dispatch
		"upload_results",
		"scrape_config",
		"add_protocol",
		// Crawl shaping — honoured via scrape_config on the adapter's crawl path
		"max_pages",
		"follow_links",
		"extract_mode",
	},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("scrape_web", WebscrapeInputSpec)
}

// extractModeFormats maps the step-level extract_mode onto the adapter's formats
// list. Anything unrecognised is left alone rather than guessed at — silently
// coercing an unknown mode to a default is how the caller's instruction gets lost,
// which is the defect being fixed here.
var extractModeFormats = map[string][]interface{}{
	"text":     {"markdown"},
	"markdown": {"markdown"},
	"html":     {"html"},
	"raw_html": {"rawHtml"},
}

// buildScrapeConfig merges the step's explicit scrape_config with the settings
// that scrape_web advertises at the top level of its config.
//
// It returns a COPY. The step config map is shared state — mutating it would
// leak these derived keys into later steps and into any loop iteration reusing
// the same definition.
//
// An explicit scrape_config value always wins over a derived one: the specific
// dialect beats the convenience alias, and a caller who wrote scrape_config.limit
// by hand should not have it overwritten by max_pages.
func buildScrapeConfig(config map[string]interface{}, action string, logger *zap.Logger) map[string]interface{} {
	out := map[string]interface{}{}
	if explicit, ok := config["scrape_config"].(map[string]interface{}); ok {
		for k, v := range explicit {
			out[k] = v
		}
	}

	// extract_mode → formats. Also cuts the fetched payload substantially, which
	// is load-bearing beyond honesty: bugs_open/062 traced Kafka max-message-size
	// failures to the 4-format default.
	if mode, ok := config["extract_mode"].(string); ok && mode != "" {
		formats, known := extractModeFormats[strings.ToLower(strings.TrimSpace(mode))]
		switch {
		case !known:
			logger.Warn("Unrecognised extract_mode; leaving formats at the provider default",
				zap.String("extract_mode", mode),
				zap.Strings("supported", []string{"text", "markdown", "html", "raw_html"}),
			)
		case out["formats"] != nil:
			logger.Info("extract_mode ignored — scrape_config.formats is set explicitly and wins",
				zap.String("extract_mode", mode),
			)
		default:
			out["formats"] = formats
		}
	}

	// max_pages → limit, follow_links → include_paths. Both are read by the
	// adapter's CRAWL path only (providers/firecrawl.go buildCrawlPayload reads
	// "limit" and "include_paths"); Firecrawl's /scrape endpoint fetches exactly
	// one page and has nowhere to put them.
	//
	// So they are passed through AND, when the step is a single-page scrape, the
	// mismatch is named. Previously this combination was the bug: a step said
	// "fetch 3 pages following these links", performed one fetch, and reported
	// success. It now either happens or says why not.
	_, hasMaxPages := config["max_pages"]
	_, hasFollowLinks := config["follow_links"]

	if mp, ok := toInt(config["max_pages"]); ok && mp > 0 {
		if _, explicit := out["limit"]; !explicit {
			out["limit"] = mp
		}
	}
	if links, ok := config["follow_links"].([]interface{}); ok && len(links) > 0 {
		if _, explicit := out["include_paths"]; !explicit {
			out["include_paths"] = links
		}
	}

	if (hasMaxPages || hasFollowLinks) && !isCrawlAction(action) {
		logger.Warn("max_pages/follow_links are set on a single-page scrape and cannot take effect",
			zap.String("action", action),
			zap.Any("max_pages", config["max_pages"]),
			zap.Any("follow_links", config["follow_links"]),
			zap.String("consequence", "exactly one page is fetched, whatever these say"),
			zap.String("fix", `set the step's config action to "crawl" to follow links, or remove these keys`),
			zap.String("ref", "bugs_open/101"),
		)
	}

	return out
}

// isCrawlAction reports whether the adapter will treat this as a multi-page crawl.
// The adapter switches on the action name (internal/adapters/webscrape/adapter.go).
func isCrawlAction(action string) bool {
	return strings.Contains(strings.ToLower(action), "crawl")
}

// toInt accepts the numeric shapes JSON config arrives in. A step config value
// decoded from JSONB is a float64; one set in Go is an int.
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case float32:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	}
	return 0, false
}

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

	// fallback_url_field: the documented "no primary URL → use this instead" path
	// (bugs_open/101). It was configured on vet-practice-verifier and read by
	// nothing, so the fallback never fired; the config promised a safety net that
	// did not exist.
	if url == "" {
		if fallbackField, ok := config["fallback_url_field"].(string); ok && fallbackField != "" {
			url = datahelpers.ExtractNestedFieldString(params.CollectedData, fallbackField)
			if url != "" {
				params.Logger.Info("Primary url_field was empty; resolved URL from fallback_url_field",
					zap.String("fallback_url_field", fallbackField),
					zap.String("url", url),
				)
			}
		}
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

	// add_protocol: a fifth inert key, found by the audit this bug's fix added
	// (scripts/audit-config-keys.sh) rather than by reading the definitions —
	// bugs_open/101 listed four. domain-research-classifier's scrape_site step
	// sets it true and nothing read it, because the only Go code with this
	// intent reads "add_protocol_if_missing" and belongs to a DIFFERENT action.
	// A bare domain reaching the adapter is a failed fetch, so this was not
	// cosmetic.
	//
	// Fires only when explicitly requested and the URL has no scheme, so it
	// cannot rewrite a URL for any caller who has not asked.
	if addProtocol, ok := config["add_protocol"].(bool); ok && addProtocol {
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			original := url
			url = "https://" + url
			params.Logger.Info("add_protocol: URL had no scheme; defaulted to https",
				zap.String("original", original),
				zap.String("url", url),
			)
		}
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

	// Add any additional config from the step, plus the settings that used to be
	// declared on the step and read by nothing (bugs_open/101).
	if scrapeConfig := buildScrapeConfig(config, action, params.Logger); len(scrapeConfig) > 0 {
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
