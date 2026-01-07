// FILE: platform/orchestration/actions/research_actions.go
// Actions for research workflow: PrepareUrlsAction, FormatResearchContentAction
package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// PrepareUrlsAction extracts top URLs from search results for batch scraping
// Also collects snippets from all results for additional context
//
// Config options:
//   - results_field: path to search results (default: "search_results")
//   - max_scrapes: number of URLs to scrape (default: 3)
//   - max_snippets: number of snippets to collect (default: 5)
//   - exclude_domains: domains to skip (default: social media sites)
//   - prefer_domains: domains to prioritize (default: authoritative sources)
//
// Input: search_results from web_search action
//
//	Web search adapter returns flat body: {success, results, query, total, provider}
//	This is stored at collected_data["search_results"] (the output_field)
//	Results array path: search_results.results
//	Each result: {url, title, snippet, published_at, source}
//
// Output (stored at output_field, e.g., "prepared_urls"):
//   - urls_to_scrape: array of {url, title, index} - for BatchWebscrapeAction
//   - scrape_count: number of URLs to scrape
//   - snippet_context: formatted string of snippets for LLM
//   - total_results: total search results received
func PrepareUrlsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PrepareUrlsAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get config values
	maxScrapes := 3
	if ms, ok := config["max_scrapes"].(float64); ok {
		maxScrapes = int(ms)
	}

	maxSnippets := 5
	if ms, ok := config["max_snippets"].(float64); ok {
		maxSnippets = int(ms)
	}

	excludeDomains := []string{
		"pinterest.com", "facebook.com", "twitter.com", "x.com",
		"instagram.com", "youtube.com", "reddit.com", "tiktok.com",
		"linkedin.com",
	}
	if ed, ok := config["exclude_domains"].([]interface{}); ok {
		excludeDomains = toStringSlice(ed)
	}

	preferDomains := []string{
		".gov", ".edu", ".org",
		"reuters.com", "bbc.com", "forbes.com", "hbr.org",
		"mckinsey.com", "harvard.edu", "mit.edu",
	}
	if pd, ok := config["prefer_domains"].([]interface{}); ok {
		preferDomains = toStringSlice(pd)
	}

	// Find results array - try multiple paths
	// Web search adapter returns: { results: [...], success: true, ... }
	// Stored at: collected_data["search_results"]
	// So array is at: search_results.results
	results := datahelpers.FindResultsArray(params.CollectedData, "search_results", params.Logger)

	if results == nil || len(results) == 0 {
		params.Logger.Warn("PrepareUrlsAction: No search results found",
			zap.Strings("collected_data_keys", getMapKeys(params.CollectedData)))
		return map[string]interface{}{
			"urls_to_scrape":  []interface{}{},
			"scrape_count":    0,
			"snippet_context": "No search results were found.",
			"total_results":   0,
		}, nil
	}

	params.Logger.Info("PrepareUrlsAction: Found results",
		zap.Int("count", len(results)))

	// Process results - separate into preferred and regular
	var preferredResults []map[string]interface{}
	var regularResults []map[string]interface{}
	var snippets []string

	for _, r := range results {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		url, _ := result["url"].(string)
		title, _ := result["title"].(string)
		snippet, _ := result["snippet"].(string)

		// Collect snippet regardless of URL (snippets are useful even without URL)
		if snippet != "" && len(snippets) < maxSnippets {
			if title == "" {
				title = "Untitled"
			}
			snippetText := truncateString(snippet, 250)
			if url != "" {
				snippets = append(snippets, fmt.Sprintf("- **%s** (%s): %s", title, url, snippetText))
			} else {
				snippets = append(snippets, fmt.Sprintf("- **%s**: %s", title, snippetText))
			}
		}

		// Skip if no URL for scraping
		if url == "" {
			continue
		}

		// Check exclusions
		lowerURL := strings.ToLower(url)
		excluded := false
		for _, domain := range excludeDomains {
			if strings.Contains(lowerURL, strings.ToLower(domain)) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check if preferred domain
		isPreferred := false
		for _, domain := range preferDomains {
			if strings.Contains(lowerURL, strings.ToLower(domain)) {
				isPreferred = true
				break
			}
		}

		if isPreferred {
			preferredResults = append(preferredResults, result)
		} else {
			regularResults = append(regularResults, result)
		}
	}

	// Combine: preferred first, then regular
	allValid := append(preferredResults, regularResults...)

	// Build URLs to scrape (up to maxScrapes)
	var urlsToScrape []map[string]interface{}
	for i, result := range allValid {
		if i >= maxScrapes {
			break
		}
		url, _ := result["url"].(string)
		title, _ := result["title"].(string)
		if title == "" {
			title = fmt.Sprintf("Source %d", i)
		}
		urlsToScrape = append(urlsToScrape, map[string]interface{}{
			"url":   url,
			"title": title,
			"index": i,
		})
	}

	// Build snippet context
	snippetContext := "No snippets available."
	if len(snippets) > 0 {
		snippetContext = strings.Join(snippets, "\n")
	}

	params.Logger.Info("PrepareUrlsAction: Complete",
		zap.Int("total_results", len(results)),
		zap.Int("preferred", len(preferredResults)),
		zap.Int("regular", len(regularResults)),
		zap.Int("urls_to_scrape", len(urlsToScrape)),
		zap.Int("snippets", len(snippets)))

	return map[string]interface{}{
		"urls_to_scrape":  urlsToScrape,
		"scrape_count":    len(urlsToScrape),
		"snippet_context": snippetContext,
		"total_results":   len(results),
		"preferred_count": len(preferredResults),
	}, nil
}

// FormatResearchContentAction formats scraped pages and snippets for LLM synthesis
//
// Config options:
//   - scrape_field: path to scrape results (default: "scrape_results")
//   - snippets_field: path to snippet context (default: "prepared_urls.snippet_context")
//   - max_content_per_source: max chars per source (default: 6000)
//
// Input:
//
//   - scrape_results: from batch_webscrape action
//     Adapter returns flat body: {success, results, success_count, error_count, total_count}
//     This is stored at collected_data["scrape_results"] (the output_field)
//     Results array path: scrape_results.results
//     Each result: {index, url, title, content, success, error?}
//
//   - prepared_urls: from prepare_urls action
//     Stored at collected_data["prepared_urls"]
//     Snippet context path: prepared_urls.snippet_context
//
// Output (stored at output_field, e.g., "research_content"):
//   - research_text: formatted string for LLM prompt (with ### [N] Source headers)
//   - sources: array of {index, url, title} for citation reference
//   - source_count: number of successful sources
//   - content_quality: "good" (3+) | "limited" (1-2) | "snippets_only" | "none"
func FormatResearchContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("FormatResearchContentAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	maxContentPerSource := 6000
	if mc, ok := config["max_content_per_source"].(float64); ok {
		maxContentPerSource = int(mc)
	}

	// Get scrape results
	// batch_webscrape returns: { results: [...], success_count: N, ... }
	// Stored at: collected_data["scrape_results"]
	scrapeField := "scrape_results"
	if sf, ok := config["scrape_field"].(string); ok {
		scrapeField = sf
	}

	var scrapedPages []interface{}
	if scrapeData := datahelpers.ExtractNestedField(params.CollectedData, scrapeField); scrapeData != nil {
		if scrapeMap, ok := scrapeData.(map[string]interface{}); ok {
			// Results are in scrape_results.results
			if results, ok := scrapeMap["results"].([]interface{}); ok {
				scrapedPages = results
			}
		}
	}

	// Get snippet context from prepare_urls step
	// prepare_urls returns: { snippet_context: "...", ... }
	// Stored at: collected_data["prepared_urls"]
	snippetsField := "prepared_urls.snippet_context"
	if sf, ok := config["snippets_field"].(string); ok {
		snippetsField = sf
	}

	snippetContext := ""
	if sc := datahelpers.ExtractNestedField(params.CollectedData, snippetsField); sc != nil {
		snippetContext, _ = sc.(string)
	}

	// Build research text
	var researchText strings.Builder
	var sources []map[string]interface{}
	successCount := 0

	// Add scraped pages
	if len(scrapedPages) > 0 {
		researchText.WriteString("## Research Sources\n\n")

		for _, page := range scrapedPages {
			pageMap, ok := page.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract fields - handle both float64 and int for index
			var index int
			if idx, ok := pageMap["index"].(float64); ok {
				index = int(idx)
			} else if idx, ok := pageMap["index"].(int); ok {
				index = idx
			}

			url, _ := pageMap["url"].(string)
			title, _ := pageMap["title"].(string)
			content, _ := pageMap["content"].(string)
			success, _ := pageMap["success"].(bool)

			if title == "" {
				title = fmt.Sprintf("Source %d", index)
			}

			if !success {
				errMsg, _ := pageMap["error"].(string)
				researchText.WriteString(fmt.Sprintf("### [%d] %s\n", index, title))
				researchText.WriteString(fmt.Sprintf("URL: %s\n", url))
				researchText.WriteString(fmt.Sprintf("*Failed to retrieve content: %s*\n\n", errMsg))
				continue
			}

			// Truncate content if needed
			if len(content) > maxContentPerSource {
				content = content[:maxContentPerSource] + "\n\n[Content truncated...]"
			}

			researchText.WriteString(fmt.Sprintf("### [%d] %s\n", index, title))
			researchText.WriteString(fmt.Sprintf("URL: %s\n\n", url))
			researchText.WriteString(content)
			researchText.WriteString("\n\n---\n\n")

			sources = append(sources, map[string]interface{}{
				"index": index,
				"url":   url,
				"title": title,
			})
			successCount++
		}
	}

	// Add snippet context for additional sources
	if snippetContext != "" && snippetContext != "No snippets available." {
		researchText.WriteString("## Additional Context (from search snippets)\n\n")
		researchText.WriteString(snippetContext)
		researchText.WriteString("\n")
	}

	finalText := researchText.String()
	if finalText == "" {
		finalText = "No research content could be retrieved."
	}

	// Determine content quality
	contentQuality := "none"
	if successCount >= 3 {
		contentQuality = "good"
	} else if successCount >= 1 {
		contentQuality = "limited"
	} else if snippetContext != "" && snippetContext != "No snippets available." {
		contentQuality = "snippets_only"
	}

	params.Logger.Info("FormatResearchContentAction: Complete",
		zap.Int("scraped_pages", len(scrapedPages)),
		zap.Int("success_count", successCount),
		zap.Int("sources", len(sources)),
		zap.String("quality", contentQuality),
		zap.Int("text_length", len(finalText)))

	return map[string]interface{}{
		"research_text":   finalText,
		"sources":         sources,
		"source_count":    len(sources),
		"content_quality": contentQuality,
	}, nil
}

// Helper functions

func toStringSlice(items []interface{}) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
