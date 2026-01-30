// FILE: platform/orchestration/actions/link_constraints.go
// Adds available pages context to LLM prompts to prevent link hallucination
//
// Usage: Call InjectLinkConstraints before executing LLM prompt
// It adds a section listing available pages and instructing the LLM
// to only link to those pages.
//
// NOTE: This ONLY constrains links. Content/link suggestions are handled
// by a separate analysis process (link-analyzer or maintenance workflow).

package actions

import (
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// LinkConstraintsConfig configures link constraint behavior
type LinkConstraintsConfig struct {
	Enabled                    bool `json:"enabled"`
	MaxInternalLinksPerSection int  `json:"max_internal_links_per_section"`
}

// PageForLinking represents a page available for internal links
type PageForLinking struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Name        string `json:"name"`
}

// InjectLinkConstraints adds available pages context to a prompt
// Returns the modified prompt with link constraints prepended
func InjectLinkConstraints(prompt string, collectedData map[string]interface{}, config LinkConstraintsConfig, logger *zap.Logger) string {
	if !config.Enabled {
		return prompt
	}

	// Extract available pages from collected data
	pages := extractAvailablePages(collectedData, logger)
	if len(pages) == 0 {
		logger.Debug("InjectLinkConstraints: No available pages found")
		return prompt
	}

	// Build the constraint section
	var sb strings.Builder
	sb.WriteString("\n## INTERNAL LINKS\n\n")
	sb.WriteString("When creating internal links, ONLY link to pages from this list:\n\n")

	for _, page := range pages {
		if page.URL != "" {
			sb.WriteString(fmt.Sprintf("- %s", page.URL))
			if page.Title != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", page.Title))
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	sb.WriteString("Do NOT create links to pages not in this list. ")
	sb.WriteString("If mentioning a topic without a corresponding page, write about it without making it a hyperlink.\n")

	if config.MaxInternalLinksPerSection > 0 {
		sb.WriteString(fmt.Sprintf("\nUse at most %d internal links per section.\n", config.MaxInternalLinksPerSection))
	}

	sb.WriteString("\n---\n\n")

	// Prepend constraints to prompt
	return sb.String() + prompt
}

// extractAvailablePages gets pages from various possible locations in collected data
func extractAvailablePages(collectedData map[string]interface{}, logger *zap.Logger) []PageForLinking {
	var pages []PageForLinking

	// Try db_sync.pages first (most common)
	if dbSync, ok := collectedData["db_sync"].(map[string]interface{}); ok {
		if pagesData, ok := dbSync["pages"].([]interface{}); ok {
			pages = convertToPageList(pagesData, logger)
			if len(pages) > 0 {
				return pages
			}
		}
	}

	// Try render_context.available_pages
	if renderCtx, ok := collectedData["render_context"].(map[string]interface{}); ok {
		if pagesData, ok := renderCtx["available_pages"].([]interface{}); ok {
			pages = convertToPageList(pagesData, logger)
			if len(pages) > 0 {
				return pages
			}
		}
	}

	// Try site_plan.pages
	if sitePlan, ok := collectedData["site_plan"].(map[string]interface{}); ok {
		if pagesData, ok := sitePlan["pages"].([]interface{}); ok {
			pages = convertToPageList(pagesData, logger)
			if len(pages) > 0 {
				return pages
			}
		}
	}

	// Try pages_to_build.pages
	if pagesToBuild, ok := collectedData["pages_to_build"].(map[string]interface{}); ok {
		if pagesData, ok := pagesToBuild["pages"].([]interface{}); ok {
			pages = convertToPageList(pagesData, logger)
		}
	}

	return pages
}

// convertToPageList converts interface slice to PageForLinking slice
func convertToPageList(pagesData []interface{}, logger *zap.Logger) []PageForLinking {
	var pages []PageForLinking

	for _, p := range pagesData {
		pageMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		page := PageForLinking{
			URL:         datahelpers.GetStringField(pageMap, "url", ""),
			Title:       datahelpers.GetStringField(pageMap, "title", ""),
			Description: datahelpers.GetStringField(pageMap, "description", ""),
			Name:        datahelpers.GetStringField(pageMap, "name", ""),
		}

		// Build URL from name if not present
		if page.URL == "" && page.Name != "" {
			if page.Name == "index" || page.Name == "home" {
				page.URL = "/index.html"
			} else {
				page.URL = "/" + page.Name + ".html"
			}
		}

		// Use name as title fallback
		if page.Title == "" && page.Name != "" {
			page.Title = strings.Title(page.Name)
		}

		if page.URL != "" {
			pages = append(pages, page)
		}
	}

	logger.Debug("convertToPageList: Extracted pages",
		zap.Int("count", len(pages)),
	)

	return pages
}
