// FILE: platform/orchestration/actions/firecrawl_map_action.go
//
// FirecrawlMapAction wraps WebscrapeAction for URL discovery via firecrawl /map.
// Returns a list of URLs found on the site without fetching page content.
// Used as Phase 1 of paginated crawling for large sites.
//
// Workflow config:
//   "discover_urls": {
//       "action": "firecrawl_map",
//       "config": {
//           "url_field": "input_data.url",
//           "limit": 200
//       },
//       "output_field": "discovered_urls",
//       "next_step": "prepare_batches"
//   }
//
// Output: {"links": ["https://...", ...], "total": N, "mapped_url": "..."}
//
// Registration:
//   "firecrawl_map": {
//       Handler:     FirecrawlMapAction,
//       Category:    "webscrape",
//       Description: "Discover site URLs via firecrawl /map endpoint (no content fetched)",
//       IsLocal:     false,
//   },

package actions

import (
	"context"

	"go.uber.org/zap"
)

func FirecrawlMapAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("in FirecrawlMapAction",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Set the action to "map" in config
	if params.StepConfig.Config == nil {
		params.StepConfig.Config = make(map[string]interface{})
	}
	params.StepConfig.Config["action"] = "map"

	// Pass limit through to scrape_config if specified at step level
	if limit, ok := params.StepConfig.Config["limit"].(float64); ok {
		scrapeConfig, _ := params.StepConfig.Config["scrape_config"].(map[string]interface{})
		if scrapeConfig == nil {
			scrapeConfig = make(map[string]interface{})
		}
		scrapeConfig["limit"] = limit
		params.StepConfig.Config["scrape_config"] = scrapeConfig
	}

	// Call the main WebscrapeAction
	return WebscrapeAction(ctx, params)
}
