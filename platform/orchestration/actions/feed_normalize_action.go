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
//
// The firecrawl adapter returns: {response: {data: {markdown_content, title, url, links, ...}}, response_status: "complete"}
// For news listing pages, the links array contains article URLs we can use as individual feed items.
// For single article pages, we treat the whole page as one item.
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

	scrapeMap, ok := scrapeData.(map[string]interface{})
	if !ok {
		logger.Warn("normalizeScrapeResults: scrape data is not a map")
		return nil
	}

	// Navigate into firecrawl adapter response wrapper: response.data
	data := scrapeMap
	if response, ok := scrapeMap["response"].(map[string]interface{}); ok {
		if responseData, ok := response["data"].(map[string]interface{}); ok {
			data = responseData
		}
	}

	// Check we have content (response_status=complete or data has content)
	if responseStatus, ok := scrapeMap["response_status"].(string); ok && responseStatus != "complete" {
		logger.Warn("normalizeScrapeResults: scrape response not complete",
			zap.String("response_status", responseStatus))
		return nil
	}

	pageURL, _ := data["url"].(string)
	pageTitle, _ := data["title"].(string)

	// Get content - prefer markdown, then HTML
	content := ""
	if md, ok := data["markdown_content"].(string); ok && md != "" {
		content = md
	} else if html, ok := data["html_content"].(string); ok && html != "" {
		content = stripHTML(html)
	}

	// Strategy 1: Extract links from the scraped page as individual feed items.
	// News listing pages contain article links — each becomes a feed item.
	var items []map[string]interface{}

	if links, ok := data["links"].([]interface{}); ok && len(links) > 0 {
		logger.Info("normalizeScrapeResults: extracting links from page",
			zap.String("page_url", pageURL),
			zap.Int("link_count", len(links)),
		)

		for i, linkRaw := range links {
			if i >= maxItems {
				break
			}

			// Links can be strings or maps with {href, text} or {url, title}
			var linkURL, linkTitle string
			switch l := linkRaw.(type) {
			case string:
				linkURL = l
			case map[string]interface{}:
				linkURL, _ = l["href"].(string)
				if linkURL == "" {
					linkURL, _ = l["url"].(string)
				}
				linkTitle, _ = l["text"].(string)
				if linkTitle == "" {
					linkTitle, _ = l["title"].(string)
				}
			}

			if linkURL == "" {
				continue
			}

			// Skip non-article links (navigation, social, etc.)
			if isNavigationLink(linkURL, pageURL) {
				continue
			}

			externalID := linkURL
			if linkTitle == "" {
				linkTitle = linkURL
			}

			items = append(items, map[string]interface{}{
				"title":        strings.TrimSpace(linkTitle),
				"url":          strings.TrimSpace(linkURL),
				"summary":      "", // we only have the link, not the article content
				"external_id":  externalID,
				"published_at": "",
			})
		}
	}

	// Strategy 2: If no usable links found, treat the whole page as one item
	if len(items) == 0 && (pageTitle != "" || pageURL != "") {
		logger.Info("normalizeScrapeResults: no links extracted, using page as single item",
			zap.String("page_url", pageURL),
			zap.String("page_title", pageTitle),
		)

		// STRIP BEFORE THE CUT (bugs_open/332, 2026-09-03). `content` is
		// firecrawl's raw markdown_content, so cutting it at 500 bytes can sever
		// a link mid-URL and leave "[text](https://…" — a half-pattern that the
		// display strip could not match until this bug, and that nothing
		// downstream can repair once written.
		//
		// INERT TODAY, deliberately shipped anyway: source_type='scrape' is 472
		// rows in 30 days with EMPTY summaries [MEASURED 2026-09-03], because
		// Strategy 1 above sets summary:"" and this Strategy-2 branch rarely
		// fires. Inert-today is the cheapest moment to close a manufacturing
		// site, and this is the only guard Strategy 2 has.
		//
		// The sibling cut at websearch/providers/firecrawl.go — the one that
		// produced the 288 severed links this bug is about — is NOT touched here.
		// It was routed to the news_feed_ingestion lane, which owns that file and
		// fixed it themselves (6f0a246de): editing an ingest record is
		// irreversible, unreachable by DISABLE_NEWS_MARKDOWN_STRIP, and feeds
		// cmd/reasoningset's training corpus.
		summary := content
		if cleaned, _ := datahelpers.StripFeedDisplayMarkdown(summary, !datahelpers.HTMLMarkupRe.MatchString(summary)); cleaned != "" {
			summary = cleaned
		}
		// WHY NOT queryresolve.FeedDisplaySummary, which this lane just built for
		// the other three readers (council 803f0d81, reuse_agent, medium — the
		// seat was right to force the question):
		//
		//   1. That projection is gated on DISABLE_NEWS_MARKDOWN_STRIP, a DISPLAY
		//      kill switch. Calling it here would put a display lever in charge of
		//      what gets WRITTEN to content_feed_items — flip the switch and the
		//      stored record silently changes shape. That is precisely the
		//      irreversibility this lane refused at firecrawl.go.
		//   2. This is an INGEST path. The projection's contract is "prepare text
		//      for a visitor"; this one's is "do not manufacture a shape the
		//      readers cannot handle". Different guarantees, deliberately.
		//
		// And TruncateString is NOT a third truncation primitive: data_helpers.go
		// defines it AS SafeCut plus the ellipsis, so this call site and
		// feedSummaryCut bottom out in the same rune-safe cut. It is used rather
		// than SafeCut directly because the ellipsis is wanted here.
		summary = datahelpers.TruncateString(summary, 500)

		externalID := pageURL
		if externalID == "" {
			contentForHash := content
			if len(contentForHash) > 100 {
				contentForHash = contentForHash[:100]
			}
			h := sha256.Sum256([]byte(pageTitle + contentForHash))
			externalID = fmt.Sprintf("scrape-%x", h[:8])
		}

		items = append(items, map[string]interface{}{
			"title":        strings.TrimSpace(pageTitle),
			"url":          strings.TrimSpace(pageURL),
			"summary":      strings.TrimSpace(summary),
			"external_id":  externalID,
			"published_at": "",
			"full_content": content,
		})
	}

	logger.Info("normalizeScrapeResults: normalised items",
		zap.Int("item_count", len(items)),
		zap.String("page_url", pageURL),
	)

	return items
}

// isNavigationLink filters out common non-article links from scraped pages.
// Keeps links that look like article/content pages, skips navigation, social, etc.
func isNavigationLink(linkURL, pageURL string) bool {
	lower := strings.ToLower(linkURL)

	// Skip anchors, javascript, mailto
	if strings.HasPrefix(lower, "#") || strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "mailto:") {
		return true
	}

	// Skip common non-article paths
	skipPatterns := []string{
		"/privacy", "/terms", "/cookie", "/contact", "/about",
		"/login", "/register", "/subscribe", "/account",
		"/rss", "/feed", "/sitemap",
		"facebook.com", "twitter.com", "linkedin.com", "instagram.com",
		"youtube.com", "t.co/", "x.com/",
	}
	for _, pattern := range skipPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}

	return false
}
