// FILE: platform/orchestration/actions/prepare_extraction_context.go
// Action: prepare_extraction_context
//
// Preprocesses raw search results and scraped content into clean text
// that can be directly injected into an LLM prompt template.
//
// Config:
//   - search_field: path to search results in collected_data (default: "search_results")
//   - scrape_field: path to scraped data in collected_data (default: "scraped_data")
//   - max_content_length: max chars of website content to include (default: 8000)
//   - max_snippets: max number of search result snippets (default: 10)
//
// Input paths (what the adapter responses look like in collected_data):
//   - search_results.response.results[]  -> {title, url, snippet, source}
//   - scraped_data.response.data.markdown_content -> string
//   - scraped_data.response.data.title -> string
//   - scraped_data.response.data.url -> string
//   - scraped_data.response.data.metadata -> map with description, keywords, etc.
//
// Output (stored at output_field, e.g. "extraction_context"):
//   - website_content: formatted string of scraped website content
//   - search_summary: formatted string of search results
//   - website_title: page title from scrape
//   - website_url: URL that was scraped
//   - source_count: how many search results were available
//   - has_website_content: bool
//   - has_search_results: bool

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func PrepareExtractionContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PrepareExtractionContextAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Config values
	searchField := "search_results"
	if sf, ok := config["search_field"].(string); ok {
		searchField = sf
	}

	scrapeField := "scraped_data"
	if sf, ok := config["scrape_field"].(string); ok {
		scrapeField = sf
	}

	maxContentLength := 8000
	if mc, ok := config["max_content_length"].(float64); ok {
		maxContentLength = int(mc)
	}

	maxSnippets := 10
	if ms, ok := config["max_snippets"].(float64); ok {
		maxSnippets = int(ms)
	}

	// ---- Format search results ----
	searchSummary := "No search results available."
	sourceCount := 0
	hasSearchResults := false

	// Path: search_results.response.results
	resultsPath := searchField + ".response.results"
	if rawResults := datahelpers.ExtractNestedField(params.CollectedData, resultsPath); rawResults != nil {
		if resultsArr, ok := rawResults.([]interface{}); ok && len(resultsArr) > 0 {
			var sb strings.Builder
			for i, item := range resultsArr {
				if i >= maxSnippets {
					break
				}
				result, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				title, _ := result["title"].(string)
				url, _ := result["url"].(string)
				snippet, _ := result["snippet"].(string)

				if title == "" && url == "" {
					continue
				}

				sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, title))
				if url != "" {
					sb.WriteString(fmt.Sprintf("    URL: %s\n", url))
				}
				if snippet != "" {
					sb.WriteString(fmt.Sprintf("    %s\n", snippet))
				}
				sb.WriteString("\n")
				sourceCount++
			}
			if sourceCount > 0 {
				searchSummary = sb.String()
				hasSearchResults = true
			}
		}
	}

	// Also try flat response (in case adapter returns at search_results.results directly)
	if !hasSearchResults {
		flatPath := searchField + ".results"
		if rawResults := datahelpers.ExtractNestedField(params.CollectedData, flatPath); rawResults != nil {
			if resultsArr, ok := rawResults.([]interface{}); ok && len(resultsArr) > 0 {
				var sb strings.Builder
				for i, item := range resultsArr {
					if i >= maxSnippets {
						break
					}
					result, ok := item.(map[string]interface{})
					if !ok {
						continue
					}
					title, _ := result["title"].(string)
					url, _ := result["url"].(string)
					snippet, _ := result["snippet"].(string)

					if title == "" && url == "" {
						continue
					}

					sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, title))
					if url != "" {
						sb.WriteString(fmt.Sprintf("    URL: %s\n", url))
					}
					if snippet != "" {
						sb.WriteString(fmt.Sprintf("    %s\n", snippet))
					}
					sb.WriteString("\n")
					sourceCount++
				}
				if sourceCount > 0 {
					searchSummary = sb.String()
					hasSearchResults = true
				}
			}
		}
	}

	params.Logger.Info("PrepareExtractionContextAction: Search results processed",
		zap.Int("source_count", sourceCount),
		zap.Bool("has_search_results", hasSearchResults))

	// ---- Format scraped website content ----
	websiteContent := "No website content available."
	websiteTitle := ""
	websiteURL := ""
	hasWebsiteContent := false

	// Try markdown_content first, then html_content
	scrapeDataPath := scrapeField + ".response.data"
	if scrapeData := datahelpers.ExtractNestedField(params.CollectedData, scrapeDataPath); scrapeData != nil {
		if dataMap, ok := scrapeData.(map[string]interface{}); ok {
			// Get title and URL
			websiteTitle, _ = dataMap["title"].(string)
			websiteTitle = strings.TrimSpace(websiteTitle)
			websiteURL, _ = dataMap["url"].(string)

			// Get content - prefer markdown
			content := ""
			if md, ok := dataMap["markdown_content"].(string); ok && md != "" {
				content = md
			} else if html, ok := dataMap["html_content"].(string); ok && html != "" {
				content = html
			}

			if content != "" {
				// Truncate if needed
				if len(content) > maxContentLength {
					content = content[:maxContentLength] + "\n\n[Content truncated at " + fmt.Sprintf("%d", maxContentLength) + " characters...]"
				}
				websiteContent = content
				hasWebsiteContent = true
			}

			// Also extract useful metadata
			if metadata, ok := dataMap["metadata"].(map[string]interface{}); ok {
				// Append description from metadata if available and not already in content
				if desc, ok := metadata["description"].(string); ok && desc != "" {
					websiteContent = "Site description: " + desc + "\n\n" + websiteContent
				}
			}
		}
	}

	params.Logger.Info("PrepareExtractionContextAction: Website content processed",
		zap.String("website_title", websiteTitle),
		zap.String("website_url", websiteURL),
		zap.Bool("has_website_content", hasWebsiteContent),
		zap.Int("content_length", len(websiteContent)))

	return map[string]interface{}{
		"website_content":     websiteContent,
		"search_summary":      searchSummary,
		"website_title":       websiteTitle,
		"website_url":         websiteURL,
		"source_count":        sourceCount,
		"has_website_content": hasWebsiteContent,
		"has_search_results":  hasSearchResults,
	}, nil
}
