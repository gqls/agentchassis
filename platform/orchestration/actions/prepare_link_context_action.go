// FILE: platform/orchestration/actions/prepare_link_context_action.go
// PrepareLinkContextAction prepares available pages context for content generation
// Used by page-content-writer to constrain internal links

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// PrepareLinkContextAction extracts available pages and builds link constraint text
// Config:
//   - pages_field: path to pages array (default: "db_sync.pages")
//   - max_links_per_section: int (default: 3)
//   - enabled: bool (default: true)
//
// Output:
//   - available_pages: []PageInfo
//   - link_constraint_text: string (for prompt inclusion)
//   - page_count: int
func PrepareLinkContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PrepareLinkContextAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Check if enabled
	enabled := true
	if v, ok := config["enabled"].(bool); ok {
		enabled = v
	}
	if !enabled {
		return map[string]interface{}{
			"enabled":              false,
			"available_pages":      []interface{}{},
			"link_constraint_text": "",
			"page_count":           0,
		}, nil
	}

	// Get max links per section
	maxLinks := 3
	if v, ok := config["max_links_per_section"].(float64); ok {
		maxLinks = int(v)
	}

	// Extract pages from collected data
	pages := extractPagesForLinking(params.CollectedData, config, params.Logger)

	// Build constraint text for prompt inclusion
	constraintText := buildLinkConstraintText(pages, maxLinks)

	params.Logger.Info("PrepareLinkContextAction: Complete",
		zap.Int("pages_found", len(pages)),
		zap.Int("max_links", maxLinks),
	)

	return map[string]interface{}{
		"enabled":              true,
		"available_pages":      pages,
		"link_constraint_text": constraintText,
		"page_count":           len(pages),
	}, nil
}

// extractPagesForLinking gets pages from collected data
func extractPagesForLinking(collectedData map[string]interface{}, config map[string]interface{}, logger *zap.Logger) []PageInfo {
	var pages []PageInfo

	// Try configured field first
	pagesField := "db_sync.pages"
	if f, ok := config["pages_field"].(string); ok && f != "" {
		pagesField = f
	}

	// Extract pages array
	pagesData := datahelpers.ExtractNestedField(collectedData, pagesField)
	if pagesData == nil {
		// Try alternate locations
		alternates := []string{
			"site_plan.pages",
			"pages_to_build.pages",
			"render_context.available_pages",
		}
		for _, alt := range alternates {
			pagesData = datahelpers.ExtractNestedField(collectedData, alt)
			if pagesData != nil {
				logger.Debug("Found pages at alternate location", zap.String("field", alt))
				break
			}
		}
	}

	if pagesData == nil {
		logger.Warn("PrepareLinkContextAction: No pages found")
		return pages
	}

	// Convert to PageInfo slice
	pagesArray, ok := pagesData.([]interface{})
	if !ok {
		logger.Warn("PrepareLinkContextAction: pages is not an array")
		return pages
	}

	for _, p := range pagesArray {
		pageMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		page := PageInfo{
			URL:         datahelpers.GetStringField(pageMap, "url", ""),
			Title:       datahelpers.GetStringField(pageMap, "title", ""),
			Name:        datahelpers.GetStringField(pageMap, "name", ""),
			Description: datahelpers.GetStringField(pageMap, "description", ""),
		}

		// Build URL from name if missing
		if page.URL == "" && page.Name != "" {
			if page.Name == "index" || page.Name == "home" {
				page.URL = "/index.html"
			} else {
				page.URL = "/" + page.Name + ".html"
			}
		}

		// Use name as title fallback
		if page.Title == "" && page.Name != "" {
			page.Title = strings.Title(strings.ReplaceAll(page.Name, "-", " "))
		}

		if page.URL != "" {
			pages = append(pages, page)
		}
	}

	return pages
}

// buildLinkConstraintText creates the constraint text for prompt inclusion
func buildLinkConstraintText(pages []PageInfo, maxLinks int) string {
	if len(pages) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("## Internal Links\n\n")
	sb.WriteString("When creating internal links, ONLY link to these pages:\n\n")

	for _, page := range pages {
		sb.WriteString(fmt.Sprintf("- %s", page.URL))
		if page.Title != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", page.Title))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\nDo NOT create links to pages not in this list. ")
	sb.WriteString("If mentioning a topic without a corresponding page, write about it without making it a hyperlink.\n")

	if maxLinks > 0 {
		sb.WriteString(fmt.Sprintf("\nUse at most %d internal links per section.", maxLinks))
	}

	return sb.String()
}
