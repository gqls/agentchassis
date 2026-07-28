// FILE: platform/orchestration/actions/feed_fetch_async_actions.go
//
// Async feed source actions for scrape and news_search source types.
//
// These follow the same wrapper pattern as FirecrawlScrapeAction (which wraps
// WebscrapeAction by setting config fields then delegating). The difference is
// that these read source_config from collected_data in Go rather than relying
// on workflow config path threading.
//
// Why wrappers instead of direct workflow config?
//
// The existing actions (WebscrapeAction, WebSearchAction) read URLs and queries
// from config fields like url_field, query_field, query. These expect dot-paths
// that resolve via ExtractNestedField. But source_config structures vary by type
// (arrays, nested objects), and the path resolver doesn't support array indexing.
// Pushing config extraction into the workflow creates fragile path expressions.
//
// These actions read source_config directly in Go, extract what they need,
// and set up the config that the underlying action expects. The workflow step
// config stays clean — no path threading.
//
// Dev guide rule #10: "Keep workflows simple — put parsing/creation logic in Go actions."
// Dev guide rule #6: "Verify the action actually reads the config fields you're setting."

package actions

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var FetchScrapeInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"source_config"},
	Optional: []string{"source_id"},
}

var FetchNewsSearchInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"source_config"},
	Optional: []string{"source_id"},
}

func init() {
	datahelpers.RegisterActionInputSpec("fetch_scrape", FetchScrapeInputSpec)
	datahelpers.RegisterActionInputSpec("fetch_news_search", FetchNewsSearchInputSpec)
}

// FetchScrapeAction reads the URL from source_config and delegates to WebscrapeAction.
//
// Source config shape:
//
//	{"url": "https://example.com/news", "scrape_config": {"only_main_content": true}, "max_items": 10}
//
// This action:
//  1. Extracts url from source_config (supports both string and first element of array)
//  2. Sets config["url"] so WebscrapeAction finds it directly (no path resolution needed)
//  3. Copies scrape_config if present
//  4. Delegates to WebscrapeAction which handles Kafka produce + await
//
// The adapter response arrives asynchronously and is stored at this step's output_field.
// The normalize_scrape step then transforms it into feed items.
func FetchScrapeAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "fetch_scrape"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Read source_config from collected_data
	sourceConfig := findSourceConfig(params.CollectedData, logger)
	if sourceConfig == nil {
		return nil, fmt.Errorf("source_config not found in collected_data or input_data")
	}

	// Extract URL — handle both "url" (string) and "urls" (array, take first)
	url := extractURLFromConfig(sourceConfig, logger)
	if url == "" {
		return nil, fmt.Errorf("no URL found in source_config — expected 'url' (string) or 'urls' (array)")
	}

	// Set up config for WebscrapeAction
	if params.StepConfig.Config == nil {
		params.StepConfig.Config = make(map[string]interface{})
	}

	// Set the URL directly — WebscrapeAction checks config["url"] as its second priority
	params.StepConfig.Config["url"] = url
	params.StepConfig.Config["action"] = "scrape"

	// Copy scrape_config if present
	if sc, ok := sourceConfig["scrape_config"].(map[string]interface{}); ok {
		params.StepConfig.Config["scrape_config"] = sc
	}

	logger.Info("FetchScrapeAction: delegating to WebscrapeAction",
		zap.String("url", url),
		zap.String("source_id", datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.source_id")),
	)

	// Delegate to WebscrapeAction — handles Kafka produce, await, response routing
	return WebscrapeAction(ctx, params)
}

// FetchNewsSearchAction reads the query from source_config and delegates to WebSearchAction.
//
// Source config shape:
//
//	{"query": "UK wholesale gas prices news", "num_results": 5, "time_range": "week"}
//
// or (multiple queries — takes first, each additional query should be a separate source):
//
//	{"queries": ["boxing news", "boxing fights"], "num_results": 10}
//
// This action:
//  1. Extracts query from source_config (supports both string and first element of array)
//  2. Sets config["query"] so WebSearchAction finds it at priority 1 (no path resolution)
//  3. Forces search_type to "news"
//  4. Delegates to WebSearchAction which handles Kafka produce + await
func FetchNewsSearchAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "fetch_news_search"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Read source_config from collected_data
	sourceConfig := findSourceConfig(params.CollectedData, logger)
	if sourceConfig == nil {
		return nil, fmt.Errorf("source_config not found in collected_data or input_data")
	}

	// Extract query — handle both "query" (string) and "queries" (array, take first)
	query := extractQueryFromConfig(sourceConfig, logger)
	if query == "" {
		return nil, fmt.Errorf("no query found in source_config — expected 'query' (string) or 'queries' (array)")
	}

	// Extract optional config
	numResults := 10
	if nr, ok := sourceConfig["num_results"].(float64); ok {
		numResults = int(nr)
	}

	provider := ""
	if p, ok := sourceConfig["provider"].(string); ok {
		provider = p
	}

	// Set up config for WebSearchAction
	if params.StepConfig.Config == nil {
		params.StepConfig.Config = make(map[string]interface{})
	}

	// Set the query directly — WebSearchAction checks config["query"] as its first priority
	params.StepConfig.Config["query"] = query
	params.StepConfig.Config["search_type"] = "news"
	params.StepConfig.Config["num_results"] = float64(numResults)
	if provider != "" {
		params.StepConfig.Config["provider"] = provider
	}
	// Optional recency window ("day", "week", "month", "year") — passed
	// through to the provider's date filter; absent means the news vertical's
	// own recency ranking applies.
	if tr, ok := sourceConfig["time_range"].(string); ok && tr != "" {
		params.StepConfig.Config["time_range"] = tr
	}

	logger.Info("FetchNewsSearchAction: delegating to WebSearchAction",
		zap.String("query", query),
		zap.Int("num_results", numResults),
		zap.String("source_id", datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.source_id")),
	)

	// Delegate to WebSearchAction — handles Kafka produce, await, response routing
	return WebSearchAction(ctx, params)
}

// ---------------------------------------------------------------------------
// Private helpers — shared between FetchScrape and FetchNewsSearch
// ---------------------------------------------------------------------------

// findSourceConfig locates source_config in collected_data.
// Checks input_data.source_config first (the standard path from call_agent input_mapping),
// then falls back to top-level source_config.
func findSourceConfig(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Primary: input_data.source_config (set by call_agent's input_mapping)
	if inputData, ok := collectedData["input_data"].(map[string]interface{}); ok {
		if sc, ok := inputData["source_config"].(map[string]interface{}); ok {
			return sc
		}
	}

	// Fallback: top-level source_config
	if sc, ok := collectedData["source_config"].(map[string]interface{}); ok {
		return sc
	}

	// Last resort: use datahelpers path resolution
	if raw := datahelpers.ExtractNestedField(collectedData, "input_data.source_config"); raw != nil {
		if sc, ok := raw.(map[string]interface{}); ok {
			return sc
		}
	}

	logger.Warn("findSourceConfig: source_config not found",
		zap.Strings("collected_data_keys", datahelpers.GetMapKeys(collectedData)),
	)
	return nil
}

// extractURLFromConfig gets a URL from source_config.
// Handles: {"url": "..."} and {"urls": ["...", "..."]} (takes first).
func extractURLFromConfig(config map[string]interface{}, logger *zap.Logger) string {
	// Priority 1: flat "url" string
	if url, ok := config["url"].(string); ok && url != "" {
		return url
	}

	// Priority 2: "urls" array (take first)
	if urls, ok := config["urls"].([]interface{}); ok && len(urls) > 0 {
		if url, ok := urls[0].(string); ok && url != "" {
			if len(urls) > 1 {
				logger.Info("extractURLFromConfig: source has multiple URLs, using first only",
					zap.Int("total_urls", len(urls)),
					zap.String("using", url),
				)
			}
			return url
		}
	}

	// Priority 3: "feed_url" (in case someone uses RSS-style config)
	if url, ok := config["feed_url"].(string); ok && url != "" {
		return url
	}

	return ""
}

// extractQueryFromConfig gets a search query from source_config.
// Handles: {"query": "..."} and {"queries": ["...", "..."]} (takes first).
func extractQueryFromConfig(config map[string]interface{}, logger *zap.Logger) string {
	// Priority 1: flat "query" string
	if query, ok := config["query"].(string); ok && query != "" {
		return query
	}

	// Priority 2: "queries" array (take first)
	if queries, ok := config["queries"].([]interface{}); ok && len(queries) > 0 {
		if query, ok := queries[0].(string); ok && query != "" {
			if len(queries) > 1 {
				logger.Info("extractQueryFromConfig: source has multiple queries, using first only",
					zap.Int("total_queries", len(queries)),
					zap.String("using", query),
				)
			}
			return query
		}
	}

	// Priority 3: "topic" (compat with web_search)
	if topic, ok := config["topic"].(string); ok && topic != "" {
		return topic
	}

	return ""
}
