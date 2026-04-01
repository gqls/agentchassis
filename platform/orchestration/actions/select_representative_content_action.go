// FILE: platform/orchestration/actions/select_representative_content_action.go
//
// SelectRepresentativeContentAction picks 2-3 pages from the crawl that best
// represent the site's writing style. These feed into the content direction
// analysis — a separate LLM call that extracts detailed writing guidelines.
//
// Selection strategy:
//   - Always includes the homepage (usually has brand voice)
//   - Picks the longest prose-heavy pages (guides, articles, about pages)
//   - Skips tool/game/calculator pages (mostly UI markup, not prose)
//   - Caps total output to avoid blowing context limits
//
// Registration:
//   "select_representative_content": {
//       Handler:     SelectRepresentativeContentAction,
//       Category:    "site",
//       Description: "Select best pages from crawl for writing style analysis",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var SelectRepresentativeContentInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{"max_pages", "max_total_chars"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("select_representative_content", SelectRepresentativeContentInputSpec)
}

func SelectRepresentativeContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "select_representative_content"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	maxPages := 3
	if mp, ok := config["max_pages"].(float64); ok {
		maxPages = int(mp)
	}

	maxTotalChars := 15000
	if mc, ok := config["max_total_chars"].(float64); ok {
		maxTotalChars = int(mc)
	}

	// Find crawl pages — same paths as buildCrawlPageIndex
	paths := []string{
		"crawl_result.pages",
		"crawl_result.response.data.pages",
		"crawl_result.body.data.pages",
		"crawl_result.data.pages",
	}

	var pages []interface{}
	for _, path := range paths {
		raw := datahelpers.ExtractNestedField(params.CollectedData, path)
		if raw != nil {
			if arr, ok := raw.([]interface{}); ok && len(arr) > 0 {
				pages = arr
				break
			}
		}
	}

	if len(pages) == 0 {
		logger.Warn("SelectRepresentativeContent: no crawl pages found")
		return map[string]interface{}{
			"has_content":   false,
			"page_count":    0,
			"selected_text": "",
		}, nil
	}

	// Score each page for prose quality
	type scoredPage struct {
		URL       string
		Title     string
		Markdown  string
		ProseLen  int
		IsHome    bool
		IsListing bool // index/listing pages — less useful for style
	}

	var scored []scoredPage

	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}

		markdown, _ := page["markdown"].(string)
		if markdown == "" {
			continue
		}

		// Skip error pages
		if strings.Contains(markdown, "NoSuchKey") {
			continue
		}

		metadata, _ := page["metadata"].(map[string]interface{})
		pageURL, _ := metadata["url"].(string)
		pageTitle, _ := metadata["title"].(string)
		statusCode, _ := metadata["statusCode"].(float64)

		if statusCode != 0 && statusCode != 200 {
			continue
		}

		// Estimate prose quality — count characters that aren't in tables,
		// code blocks, or short UI lines (buttons, labels, navigation)
		proseLen := estimateProseLength(markdown)

		isHome := strings.TrimRight(pageURL, "/") == "" ||
			strings.HasSuffix(pageURL, ".com") ||
			strings.HasSuffix(pageURL, ".co.uk") ||
			strings.HasSuffix(pageURL, ".uk") ||
			pageURL == "/" ||
			strings.HasSuffix(strings.TrimRight(pageURL, "/"), pageTitle)

		// Detect if URL looks like root domain
		for _, suffix := range []string{".com/", ".co.uk/", ".uk/", ".org/", ".net/"} {
			if strings.Contains(pageURL, suffix) {
				afterDomain := pageURL[strings.Index(pageURL, suffix)+len(suffix):]
				if afterDomain == "" || afterDomain == "index.html" {
					isHome = true
				}
			}
		}

		isListing := strings.Contains(pageURL, "/index.html") && !isHome

		scored = append(scored, scoredPage{
			URL:       pageURL,
			Title:     pageTitle,
			Markdown:  markdown,
			ProseLen:  proseLen,
			IsHome:    isHome,
			IsListing: isListing,
		})
	}

	if len(scored) == 0 {
		return map[string]interface{}{
			"has_content":   false,
			"page_count":    0,
			"selected_text": "",
		}, nil
	}

	// Sort: homepage first, then by prose length descending, listings last
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].IsHome != scored[j].IsHome {
			return scored[i].IsHome
		}
		if scored[i].IsListing != scored[j].IsListing {
			return !scored[i].IsListing // non-listings first
		}
		return scored[i].ProseLen > scored[j].ProseLen
	})

	// Select top N pages within char budget
	var selected []scoredPage
	totalChars := 0

	for _, p := range scored {
		if len(selected) >= maxPages {
			break
		}

		contentLen := len(p.Markdown)
		if totalChars+contentLen > maxTotalChars && len(selected) > 0 {
			// Try to fit a truncated version if we have room
			remaining := maxTotalChars - totalChars
			if remaining > 2000 {
				p.Markdown = p.Markdown[:remaining] + "\n... [truncated for analysis]"
				contentLen = remaining
			} else {
				continue
			}
		}

		selected = append(selected, p)
		totalChars += contentLen
	}

	// Build output text
	var parts []string
	for i, p := range selected {
		label := "CONTENT PAGE"
		if p.IsHome {
			label = "HOMEPAGE"
		}
		parts = append(parts, fmt.Sprintf(
			"=== %s %d: %s ===\nURL: %s\n\n%s",
			label, i+1, p.Title, p.URL, p.Markdown,
		))
	}

	selectedText := strings.Join(parts, "\n\n---\n\n")

	logger.Info("SelectRepresentativeContent: selected pages",
		zap.Int("total_crawled", len(scored)),
		zap.Int("selected", len(selected)),
		zap.Int("total_chars", totalChars))

	return map[string]interface{}{
		"has_content":   true,
		"page_count":    len(selected),
		"total_chars":   totalChars,
		"selected_text": selectedText,
		"selected_pages": func() []map[string]interface{} {
			out := make([]map[string]interface{}, len(selected))
			for i, p := range selected {
				out[i] = map[string]interface{}{
					"url":       p.URL,
					"title":     p.Title,
					"prose_len": p.ProseLen,
					"is_home":   p.IsHome,
				}
			}
			return out
		}(),
	}, nil
}

// estimateProseLength counts characters that are likely prose rather than
// UI markup, tables, or short navigational elements. Longer lines are more
// likely to be real sentences.
func estimateProseLength(markdown string) int {
	proseChars := 0
	lines := strings.Split(markdown, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			continue
		}

		// Skip table rows
		if strings.HasPrefix(trimmed, "|") {
			continue
		}

		// Skip image references
		if strings.HasPrefix(trimmed, "![") {
			continue
		}

		// Skip very short lines (nav items, button labels, badges)
		if len(trimmed) < 30 {
			continue
		}

		// Skip lines that are mostly links
		linkChars := 0
		for _, part := range strings.Split(trimmed, "](") {
			if strings.Contains(part, "http") {
				linkChars += len(part)
			}
		}
		if linkChars > len(trimmed)/2 {
			continue
		}

		// This line is probably prose
		proseChars += len(trimmed)
	}

	return proseChars
}
