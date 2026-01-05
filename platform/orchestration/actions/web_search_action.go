// FILE: platform/orchestration/actions/web_search_action.go
// WebSearchAction sends search queries to the web search adapter
//
// ==============================================================================
// REGISTRATION REQUIRED - Add to TWO places:
// ==============================================================================
//
// 1. LocalActions map (platform/orchestration/actions_list/local_actions.go):
//    Add after the web scraping entries:
//
//        // Web search
//        "web_search": true,
//
// 2. GlobalActionRegistry (registry.go):
//    Add after the web scraping entries:
//
//        // Web search
//        "web_search": WebSearchAction,
//
// ==============================================================================

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

const (
	webSearchAdapterTopic = "system.adapter.web.search"
)

// WebSearchResult represents the result of initiating a web search
type WebSearchResult struct {
	Success       bool                   `json:"success"`
	RequestID     string                 `json:"request_id"`
	TopicSentTo   string                 `json:"topic_sent_to"`
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// WebSearchAction sends a search query to the web search adapter
// Config options:
//   - query: the search query string (can also come from "topic" field for compatibility)
//   - query_field: field path to extract query from collected data (default: "query")
//   - num_results: number of results to return (default: 10)
//   - search_type: "web", "news", or "images" (default: "web")
//   - provider: specific provider to use (optional)
func WebSearchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing WebSearchAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	config := params.StepConfig.Config

	// Extract search query from various sources
	query := extractSearchQuery(params, config)
	if query == "" {
		return nil, fmt.Errorf("search query not found - check 'query', 'topic', or 'query_field' config")
	}

	// Extract optional parameters
	numResults := 10
	if nr, ok := config["num_results"].(float64); ok {
		numResults = int(nr)
	}

	searchType := "web"
	if st, ok := config["search_type"].(string); ok {
		searchType = st
	}

	provider := ""
	if p, ok := config["provider"].(string); ok {
		provider = p
	}

	// Get client ID
	clientID := params.ExecutionContext.ClientID
	if clientID == "" {
		clientID = "default"
	}

	// Generate new request ID
	newRequestID := uuid.NewString()

	// Get response topic
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		myResponsesTopic = os.Getenv("RESPONSES_TOPIC")
	}
	if myResponsesTopic == "" {
		myResponsesTopic = fmt.Sprintf("system.agent.%s.responses", params.ExecutionContext.Sender.AgentType)
	}

	params.Logger.Info("Using responses topic for web search",
		zap.String("responses_topic", myResponsesTopic))

	// Build adapter request payload
	// The adapter expects: { "action": "search", "data": { "query": "...", ... } }
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",

			// Sender information
			"sender_agent_type":    params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":      params.ExecutionContext.OrchestrationID,
			"sender_pod_name":      params.ExecutionContext.Sender.PodName,
			"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,
			"sender_role":          params.ExecutionContext.Sender.Role,

			// Response routing - adapter needs to read this!
			"responses_topic":        myResponsesTopic,
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,

			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
		"body": map[string]interface{}{
			"action": "search",
			"data": map[string]interface{}{
				"query":       query,
				"num_results": numResults,
				"search_type": searchType,
				"provider":    provider,
			},
			// Include reply routing in body as well
			"reply_to_topic": myResponsesTopic,
		},
	}

	// Convert headers for Kafka
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
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	key := []byte(params.ExecutionContext.CorrelationID)

	params.Logger.Info("Sending request to web search adapter",
		zap.String("topic", webSearchAdapterTopic),
		zap.String("request_id", newRequestID),
		zap.String("reply_to_topic", myResponsesTopic),
		zap.String("query", query),
		zap.Int("num_results", numResults),
		zap.String("search_type", searchType),
	)

	// Send to adapter
	if err := params.Producer.ProduceWithValidation(
		ctx,
		webSearchAdapterTopic,
		headers,
		key,
		messageBytes,
	); err != nil {
		return nil, fmt.Errorf("failed to send to web search adapter: %w", err)
	}

	// Return result indicating we're waiting for response
	return &WebSearchResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   webSearchAdapterTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"query":           query,
			"num_results":     numResults,
			"search_type":     searchType,
			"provider":        provider,
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
			"responses_topic": myResponsesTopic,
		},
	}, nil
}

// extractSearchQuery extracts the search query from various sources
func extractSearchQuery(params ActionParams, config map[string]interface{}) string {
	// Priority 1: Direct "query" in config
	if q, ok := config["query"].(string); ok && q != "" {
		params.Logger.Debug("Using query from config", zap.String("query", q))
		return q
	}

	// Priority 2: "topic" field (for compatibility)
	if t, ok := config["topic"].(string); ok && t != "" {
		params.Logger.Debug("Using topic as query from config", zap.String("query", t))
		return t
	}

	// Priority 3: Extract from collected data using query_from path
	// This handles "query_from": "search_query.result" from research-agent workflow
	if queryFrom, ok := config["query_from"].(string); ok && queryFrom != "" {
		params.Logger.Info("DEBUGaa: extractSearchQuery query_from path",
			zap.String("query_from", queryFrom),
			zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
			zap.Any("search_query_value", params.CollectedData["search_query"]),
		)
		if value := extractWithFallbacks(params.CollectedData, queryFrom, params.Logger); value != nil {
			if queryStr, ok := value.(string); ok && queryStr != "" {
				// Sanity check - don't use LLM error messages as queries
				lowerQuery := strings.ToLower(queryStr)
				if !strings.Contains(lowerQuery, "cannot generate") &&
					!strings.Contains(lowerQuery, "no topic") &&
					!strings.Contains(lowerQuery, "please supply") &&
					!strings.Contains(lowerQuery, "please provide") &&
					len(queryStr) < 200 {
					params.Logger.Debug("Using query from query_from path",
						zap.String("path", queryFrom),
						zap.String("query", queryStr))
					return queryStr
				}
				params.Logger.Warn("query_from resolved to invalid query (likely LLM error message)",
					zap.String("path", queryFrom),
					zap.String("value_preview", queryStr[:min(100, len(queryStr))]))
			}
		}
	}

	// Priority 4: Extract from collected data using query_field path
	queryField := "query"
	if qf, ok := config["query_field"].(string); ok {
		queryField = qf
	}

	// Try various locations
	if value := extractWithFallbacks(params.CollectedData, queryField, params.Logger); value != nil {
		if q, ok := value.(string); ok && q != "" {
			return q
		}
	}

	// Priority 5: Try standard field names
	standardFields := []string{"topic", "search_topic", "search_query"}
	for _, field := range standardFields {
		if value := extractWithFallbacks(params.CollectedData, field, params.Logger); value != nil {
			if q, ok := value.(string); ok && q != "" {
				params.Logger.Debug("Using query from standard field", zap.String("field", field))
				return q
			}
		}
	}

	// Priority 6: Build query from section context (fallback for research-agent)
	query := buildQueryFromSectionContext(params, params.Logger)
	if query != "" {
		return query
	}

	return ""
}

// extractWithFallbacks tries multiple path variations to find data
func extractWithFallbacks(data map[string]interface{}, path string, logger *zap.Logger) interface{} {
	logger.Info("DEBUGaa: extractWithFallbacks starting",
		zap.String("path", path),
		zap.Any("data_keys", datahelpers.GetMapKeys(data)),
	)

	// Try direct path first
	if value := datahelpers.ExtractNestedField(data, path); value != nil {
		logger.Info("DEBUGaa: extractWithFallbacks found via direct path",
			zap.String("path", path),
			zap.Any("value", value),
		)
		return value
	}
	logger.Debug("DEBUGaa: extractWithFallbacks direct path returned nil",
		zap.String("path", path),
	)

	// If path doesn't start with input_data, try with prefix
	if !strings.HasPrefix(path, "input_data.") {
		prefixedPath := "input_data." + path
		if value := datahelpers.ExtractNestedField(data, prefixedPath); value != nil {
			logger.Info("Found via input_data prefix",
				zap.String("original", path),
				zap.String("actual", prefixedPath))
			return value
		}
	}

	// Try __raw_message__.body.input_data path
	if !strings.HasPrefix(path, "__raw_message__") {
		rawPath := "__raw_message__.body.input_data." + path
		if value := datahelpers.ExtractNestedField(data, rawPath); value != nil {
			logger.Info("Found via __raw_message__ prefix", zap.String("original", path))
			return value
		}
	}

	return nil
}

// buildQueryFromSectionContext builds a search query from current_section data
// This is a fallback when no explicit query is provided
func buildQueryFromSectionContext(params ActionParams, logger *zap.Logger) string {
	var currentSection map[string]interface{}

	// Try root level first
	if cs, ok := params.CollectedData["current_section"].(map[string]interface{}); ok {
		currentSection = cs
	}
	// Try input_data.current_section
	if currentSection == nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if cs, ok := inputData["current_section"].(map[string]interface{}); ok {
				currentSection = cs
			}
		}
	}
	// Try __raw_message__.body.input_data.current_section
	if currentSection == nil {
		if rawMsg, ok := params.CollectedData["__raw_message__"].(map[string]interface{}); ok {
			if body, ok := rawMsg["body"].(map[string]interface{}); ok {
				if inputData, ok := body["input_data"].(map[string]interface{}); ok {
					if cs, ok := inputData["current_section"].(map[string]interface{}); ok {
						currentSection = cs
					}
				}
			}
		}
	}

	if currentSection == nil {
		return ""
	}

	// Get section name/function
	sectionName := ""
	for _, key := range []string{"function", "name", "topic", "research_query"} {
		if val, ok := currentSection[key].(string); ok && val != "" {
			sectionName = val
			break
		}
	}

	if sectionName == "" {
		return ""
	}

	// Get company name from reviewed_brief
	companyName := ""
	if brief := extractWithFallbacks(params.CollectedData, "reviewed_brief", logger); brief != nil {
		if briefMap, ok := brief.(map[string]interface{}); ok {
			if cn, ok := briefMap["company_name"].(string); ok {
				companyName = cn
			}
		}
	}

	// Get domain
	domain := ""
	if siteRecord := extractWithFallbacks(params.CollectedData, "site_record", logger); siteRecord != nil {
		if srMap, ok := siteRecord.(map[string]interface{}); ok {
			if d, ok := srMap["domain"].(string); ok {
				domain = d
			}
		}
	}

	// Build query - combine section name with company/domain context
	query := sectionName
	if companyName != "" {
		query = fmt.Sprintf("%s %s", sectionName, companyName)
	} else if domain != "" {
		query = fmt.Sprintf("%s %s", sectionName, domain)
	}

	logger.Info("Built query from section context",
		zap.String("section", sectionName),
		zap.String("company", companyName),
		zap.String("domain", domain),
		zap.String("query", query))

	return query
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
