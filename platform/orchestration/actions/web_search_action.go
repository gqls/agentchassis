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

	// Priority 2: "topic" field (for compatibility with research-agent workflow)
	if t, ok := config["topic"].(string); ok && t != "" {
		params.Logger.Debug("Using topic as query from config", zap.String("query", t))
		return t
	}

	// Priority 2.5: Extract from collected data using query_from path
	if queryFrom, ok := config["query_from"].(string); ok && queryFrom != "" {
		if value := datahelpers.ExtractNestedField(params.CollectedData, queryFrom); value != nil {
			if queryStr, ok := value.(string); ok && queryStr != "" {
				// Sanity check - don't use LLM error messages as queries
				if !strings.Contains(strings.ToLower(queryStr), "cannot") &&
					!strings.Contains(strings.ToLower(queryStr), "no topic") &&
					len(queryStr) < 200 {
					params.Logger.Debug("Using query from query_from path",
						zap.String("path", queryFrom),
						zap.String("query", queryStr))
					return queryStr
				}
				params.Logger.Warn("query_from resolved to invalid query (likely LLM error)",
					zap.String("path", queryFrom),
					zap.String("value_preview", queryStr[:min(100, len(queryStr))]))
			}
		}
	}
	
	// Priority 3: Extract from collected data using query_field path
	queryField := "query"
	if qf, ok := config["query_field"].(string); ok {
		queryField = qf
	}

	// Try to find in collected data
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if q, ok := inputData[queryField].(string); ok && q != "" {
			params.Logger.Debug("Using query from input_data", zap.String("field", queryField), zap.String("query", q))
			return q
		}
		// Also try "topic" in input_data
		if t, ok := inputData["topic"].(string); ok && t != "" {
			params.Logger.Debug("Using topic from input_data as query", zap.String("query", t))
			return t
		}
		// Try "search_topic"
		if st, ok := inputData["search_topic"].(string); ok && st != "" {
			params.Logger.Debug("Using search_topic from input_data as query", zap.String("query", st))
			return st
		}
	}

	// Priority 4: Try direct collected data keys
	if q, ok := params.CollectedData[queryField].(string); ok && q != "" {
		return q
	}
	if t, ok := params.CollectedData["topic"].(string); ok && t != "" {
		return t
	}
	if st, ok := params.CollectedData["search_topic"].(string); ok && st != "" {
		return st
	}

	// Priority 5: Try to build query from section context (for research-agent)
	if currentSection, ok := params.CollectedData["current_section"].(map[string]interface{}); ok {
		// Build research query from section function/name and domain
		sectionName := ""
		if fn, ok := currentSection["function"].(string); ok {
			sectionName = fn
		} else if name, ok := currentSection["name"].(string); ok {
			sectionName = name
		}

		domain := ""
		if d, ok := params.CollectedData["domain"].(string); ok {
			domain = d
		} else if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if d, ok := inputData["domain"].(string); ok {
				domain = d
			}
		}

		if sectionName != "" {
			query := sectionName
			if domain != "" {
				query = fmt.Sprintf("%s %s", sectionName, domain)
			}
			params.Logger.Debug("Built query from section context", zap.String("query", query))
			return query
		}
	}

	return ""
}
