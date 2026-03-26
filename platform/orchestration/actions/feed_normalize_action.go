// FILE: platform/orchestration/actions/feed_normalize_action.go
// NormalizeToFeedItemsAction transforms outputs from web_search or firecrawl_scrape
// into the normalised {title, url, summary, external_id, published_at} format
// expected by WriteFeedItemsAction.
//
// This is the bridge between the async source types (news_search, scrape)
// and the shared write_feed_items action.

package actions

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var NormalizeToFeedItemsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"source_format", "results_field", "max_items"},
	Defaults: map[string]interface{}{
		"source_format": "search",
		"max_items":     20,
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("normalize_to_feed_items", NormalizeToFeedItemsInputSpec)
}

// NormalizeToFeedItemsAction reads results from collected_data (from a preceding
// web_search or firecrawl_scrape step) and transforms them into the normalised
// array format for write_feed_items.
//
// Config:
//   - source_format: "search" (default) or "scrape"
//   - results_field: dot-path to the results in collected_data
//     For search: defaults to "search_results.results"
//     For scrape: defaults to "scrape_results"
//   - max_items: cap on items to normalise (default 20)
//
// Output:
//   - items: normalised array of {title, url, summary, external_id, published_at}
//   - item_count: number of items produced

func NormalizeToFeedItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "normalize_to_feed_items"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	sourceFormat := "search"
	if sf, ok := config["source_format"].(string); ok {
		sourceFormat = sf
	}

	maxItems := 20
	if mi, ok := config["max_items"].(float64); ok {
		maxItems = int(mi)
	}

	// Determine where to find results
	resultsField := ""
	if rf, ok := config["results_field"].(string); ok {
		resultsField = rf
	}

	var items []map[string]interface{}

	switch sourceFormat {
	case "search":
		items = normalizeSearchResults(params.CollectedData, resultsField, maxItems, logger)
	case "scrape":
		items = normalizeScrapeResults(params.CollectedData, resultsField, maxItems, logger)
	default:
		return nil, fmt.Errorf("unsupported source_format: %s (expected 'search' or 'scrape')", sourceFormat)
	}

	logger.Info("NormalizeToFeedItemsAction: normalised items",
		zap.String("source_format", sourceFormat),
		zap.Int("item_count", len(items)),
	)

	return map[string]interface{}{
		"items":      items,
		"item_count": len(items),
	}, nil
}

// normalizeSearchResults transforms web search adapter output into feed items.
// Web search results have: {url, title, snippet, published_at, source}
func normalizeSearchResults(collectedData map[string]interface{}, field string, maxItems int, logger *zap.Logger) []map[string]interface{} {
	if field == "" {
		field = "search_results"
	}

	// The web search adapter returns: {success, results: [...], query, total, provider}
	// Stored at collected_data[field]
	results := datahelpers.FindResultsArray(collectedData, field, logger)
	if results == nil {
		logger.Warn("normalizeSearchResults: no results found",
			zap.String("field", field),
			zap.Strings("keys", datahelpers.GetMapKeys(collectedData)),
		)
		return nil
	}

	var items []map[string]interface{}
	for i, r := range results {
		if i >= maxItems {
			break
		}
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		url, _ := result["url"].(string)
		title, _ := result["title"].(string)
		snippet, _ := result["snippet"].(string)
		publishedAt, _ := result["published_at"].(string)

		if title == "" && url == "" {
			continue
		}

		externalID := url
		if externalID == "" {
			h := sha256.Sum256([]byte(title))
			externalID = fmt.Sprintf("search-%x", h[:8])
		}

		items = append(items, map[string]interface{}{
			"title":        strings.TrimSpace(title),
			"url":          strings.TrimSpace(url),
			"summary":      strings.TrimSpace(snippet),
			"external_id":  externalID,
			"published_at": publishedAt,
		})
	}
	return items
}

// normalizeScrapeResults transforms firecrawl/webscrape results into feed items.
// Scrape results vary but typically have: {url, title, content/markdown_content, success}
// For news scraping, we extract article links from the scraped page.
func normalizeScrapeResults(collectedData map[string]interface{}, field string, maxItems int, logger *zap.Logger) []map[string]interface{} {
	if field == "" {
		field = "scrape_results"
	}

	scrapeData := datahelpers.ExtractNestedField(collectedData, field)
	if scrapeData == nil {
		logger.Warn("normalizeScrapeResults: no scrape data found",
			zap.String("field", field),
		)
		return nil
	}

	// Handle single scrape result (map) or batch results (array in results key)
	var scrapeResults []interface{}

	switch v := scrapeData.(type) {
	case map[string]interface{}:
		// Single result or wrapper with results array
		if results, ok := v["results"].([]interface{}); ok {
			scrapeResults = results
		} else {
			scrapeResults = []interface{}{v}
		}
	case []interface{}:
		scrapeResults = v
	}

	var items []map[string]interface{}
	for i, r := range scrapeResults {
		if i >= maxItems {
			break
		}
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		success, _ := result["success"].(bool)
		if !success {
			continue
		}

		url, _ := result["url"].(string)
		title, _ := result["title"].(string)

		// Get content - prefer markdown, then clean, then raw HTML
		content := ""
		if md, ok := result["markdown_content"].(string); ok && md != "" {
			content = md
		} else if clean, ok := result["clean_content"].(string); ok && clean != "" {
			content = clean
		} else if html, ok := result["content"].(string); ok && html != "" {
			content = stripHTML(html)
		}

		if title == "" && url == "" {
			continue
		}

		// For scrape results, the content IS the article — use first ~500 chars as summary
		summary := content
		if len(summary) > 500 {
			summary = summary[:500] + "..."
		}

		externalID := url
		if externalID == "" {
			contentForHash := content
			if len(contentForHash) > 100 {
				contentForHash = contentForHash[:100]
			}
			h := sha256.Sum256([]byte(title + contentForHash))
			externalID = fmt.Sprintf("scrape-%x", h[:8])
		}

		items = append(items, map[string]interface{}{
			"title":        strings.TrimSpace(title),
			"url":          strings.TrimSpace(url),
			"summary":      strings.TrimSpace(summary),
			"external_id":  externalID,
			"published_at": "", // scrape results rarely have published dates
			"full_content": content,
		})
	}
	return items
}
