// FILE: platform/orchestration/actions/format_crawl_for_analysis_action.go
//
// FormatCrawlForAnalysisAction takes Firecrawl crawl output and formats it
// into readable text for LLM analysis. Unlike format_research_content which
// expects batch_webscrape format ({results[].content}), this handles the
// crawl format ({pages[].markdown, pages[].metadata}).
//
// Registration:
//   "format_crawl_for_analysis": {
//       Handler:     FormatCrawlForAnalysisAction,
//       Category:    "web",
//       Description: "Format Firecrawl crawl output into readable text for LLM analysis",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func FormatCrawlForAnalysisAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "format_crawl_for_analysis"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Config: crawl_field — path to crawl result in collected_data (default: "crawl_result")
	crawlField := "crawl_result"
	if cf, ok := config["crawl_field"].(string); ok && cf != "" {
		crawlField = cf
	}

	// Try multiple paths to find the pages array — the adapter response
	// nesting can vary depending on how BuildCollectedData unwraps it
	var pages []interface{}

	paths := []string{
		crawlField + ".pages",
		crawlField + ".body.data.pages",
		crawlField + ".data.pages",
		crawlField + ".body.pages",
	}

	for _, path := range paths {
		raw := datahelpers.ExtractNestedField(params.CollectedData, path)
		if raw != nil {
			if arr, ok := raw.([]interface{}); ok && len(arr) > 0 {
				pages = arr
				logger.Info("Found pages at path",
					zap.String("path", path),
					zap.Int("count", len(arr)))
				break
			}
		}
	}

	if len(pages) == 0 {
		logger.Warn("No pages found in crawl result — trying DB fallback",
			zap.String("crawl_field", crawlField),
			zap.Strings("tried_paths", paths))

		// DB fallback: paginated crawl stores pages in research_results
		// instead of collected_data. Read from there.
		if params.DB != nil {
			siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
			if siteIDStr != "" {
				rows, err := params.DB.QueryContext(ctx, `
					SELECT data FROM research_results
					WHERE site_id = $1 AND result_type = 'adoption_crawl_page'
					ORDER BY created_at
				`, siteIDStr)
				if err == nil {
					defer rows.Close()
					for rows.Next() {
						var dataJSON []byte
						if err := rows.Scan(&dataJSON); err != nil {
							continue
						}
						var pageData map[string]interface{}
						if err := json.Unmarshal(dataJSON, &pageData); err != nil {
							continue
						}
						// Reshape to match the format the rest of this function expects:
						// {markdown: "...", metadata: {url: "...", title: "..."}}
						page := map[string]interface{}{
							"markdown": pageData["markdown"],
						}
						metadata := map[string]interface{}{}
						if url, ok := pageData["url"].(string); ok {
							metadata["url"] = url
							metadata["sourceURL"] = url
						}
						if title, ok := pageData["title"].(string); ok {
							metadata["title"] = title
						}
						if meta, ok := pageData["metadata"].(map[string]interface{}); ok {
							// Merge stored metadata
							for k, v := range meta {
								metadata[k] = v
							}
						}
						metadata["statusCode"] = float64(200)
						page["metadata"] = metadata
						pages = append(pages, page)
					}
					if len(pages) > 0 {
						logger.Info("Loaded pages from DB fallback",
							zap.Int("count", len(pages)))
					}
				}
			}
		}

		// If still empty after DB fallback
		if len(pages) == 0 {
			return map[string]interface{}{
				"research_text":   "No crawl content could be retrieved.",
				"source_count":    0,
				"content_quality": "none",
				"sources":         []interface{}{},
			}, nil
		}
	}

	// Config: summary mode for LLM (light) vs full content
	// Default to summary — the LLM only needs structure, not full content.
	// Full content stays in crawl_result for Go-side extraction.
	summaryChars := 500
	if sc, ok := config["summary_chars_per_page"].(float64); ok {
		summaryChars = int(sc)
	}

	var textParts []string
	var sources []map[string]interface{}
	successCount := 0

	for i, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract markdown content
		markdown, _ := page["markdown"].(string)
		if markdown == "" {
			continue
		}

		// Skip 404/error pages
		if strings.Contains(markdown, "NoSuchKey") || strings.Contains(markdown, "\"error\":") {
			metadata, _ := page["metadata"].(map[string]interface{})
			statusCode, _ := metadata["statusCode"].(float64)
			if statusCode == 404 || statusCode == 0 {
				continue
			}
		}

		// Extract metadata
		metadata, _ := page["metadata"].(map[string]interface{})
		pageURL, _ := metadata["url"].(string)
		pageTitle, _ := metadata["title"].(string)
		statusCode, _ := metadata["statusCode"].(float64)

		// Skip non-200 pages
		if statusCode != 0 && statusCode != 200 {
			continue
		}

		if pageURL == "" {
			pageURL, _ = metadata["sourceURL"].(string)
		}

		// For LLM: just enough to classify the page structure.
		// Full content is available in crawl_result for Go-side extraction.
		summary := markdown
		if len(summary) > summaryChars {
			summary = summary[:summaryChars] + "\n... [truncated for classification]"
		}

		textParts = append(textParts, fmt.Sprintf(
			"=== PAGE %d: %s ===\nURL: %s\n\n%s",
			i+1, pageTitle, pageURL, summary,
		))

		sources = append(sources, map[string]interface{}{
			"index": i + 1,
			"url":   pageURL,
			"title": pageTitle,
		})
		successCount++
	}

	if successCount == 0 {
		return map[string]interface{}{
			"research_text":   "Crawl completed but no usable page content was found.",
			"source_count":    0,
			"content_quality": "none",
			"sources":         []interface{}{},
		}, nil
	}

	quality := "good"
	if successCount < 3 {
		quality = "limited"
	}

	researchText := strings.Join(textParts, "\n\n---\n\n")

	logger.Info("Crawl content formatted",
		zap.Int("pages_found", len(pages)),
		zap.Int("usable_pages", successCount),
		zap.String("quality", quality),
		zap.Int("text_length", len(researchText)))

	return map[string]interface{}{
		"research_text":   researchText,
		"source_count":    successCount,
		"content_quality": quality,
		"sources":         sources,
	}, nil
}
