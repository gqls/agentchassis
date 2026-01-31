// FILE: platform/orchestration/actions/save_page_sections_action.go
// SavePageSectionsAction saves rendered HTML sections to page_components table
// This ensures rerender can reassemble pages from stored sections
//
// Called after deploy_page in pageflow-builder's build_pages_loop

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// SavePageSectionsAction extracts sections from rendered page HTML and saves to page_components
// Config:
//   - html_field: path to HTML content (default: "assembled_page.html")
//   - page_name_field: path to page name (default: "current_page.name")
//   - site_id_field: path to site_id (default: "site_record.site_id")
//   - input_fields: alternative - array of field names to extract
//
// This action parses the HTML for <section> elements and stores each one in page_components.rendered_html
func SavePageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SavePageSectionsAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.DB == nil {
		params.Logger.Warn("SavePageSectionsAction: No database connection, skipping")
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"skipped":        true,
			"reason":         "no database connection",
		}, nil
	}

	config := params.StepConfig.Config

	// Extract fields - try direct paths first, then input_fields pattern
	var html, pageName, siteIDStr string

	// Try direct field paths (preferred for workflow config clarity)
	htmlField := "assembled_page.html"
	if f, ok := config["html_field"].(string); ok && f != "" {
		htmlField = f
	}
	html = datahelpers.ExtractNestedFieldString(params.CollectedData, htmlField)

	pageNameField := "current_page.name"
	if f, ok := config["page_name_field"].(string); ok && f != "" {
		pageNameField = f
	}
	pageName = datahelpers.ExtractNestedFieldString(params.CollectedData, pageNameField)

	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)

	// Fallback to input_fields pattern if direct paths didn't work
	if html == "" || pageName == "" || siteIDStr == "" {
		inputFields := []string{"page_content", "site_record", "current_page"}
		if fields, ok := config["input_fields"].([]interface{}); ok {
			inputFields = make([]string, len(fields))
			for i, f := range fields {
				inputFields[i], _ = f.(string)
			}
		}

		extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

		if html == "" {
			if pageContent, ok := extracted["page_content"].(map[string]interface{}); ok {
				if response, ok := pageContent["response"].(map[string]interface{}); ok {
					html, _ = response["page_html"].(string)
				}
			}
		}

		if pageName == "" {
			if currentPage, ok := extracted["current_page"].(map[string]interface{}); ok {
				pageName, _ = currentPage["name"].(string)
			}
		}

		if siteIDStr == "" {
			if siteRecord, ok := extracted["site_record"].(map[string]interface{}); ok {
				siteIDStr, _ = siteRecord["site_id"].(string)
			}
		}
	}

	// Validate we have what we need
	if html == "" {
		params.Logger.Warn("SavePageSectionsAction: No HTML found, skipping",
			zap.String("html_field", htmlField),
		)
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"skipped":        true,
			"reason":         "no HTML content",
		}, nil
	}

	if pageName == "" {
		params.Logger.Warn("SavePageSectionsAction: No page name found, skipping")
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"skipped":        true,
			"reason":         "no page name",
		}, nil
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		params.Logger.Warn("SavePageSectionsAction: Invalid site_id, skipping",
			zap.String("site_id_str", siteIDStr),
			zap.Error(err),
		)
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"skipped":        true,
			"reason":         "invalid site_id",
		}, nil
	}

	// Look up page_id
	pageID, err := saveSectionsLookupPageID(ctx, params.DB, siteID, pageName)
	if err != nil {
		params.Logger.Warn("SavePageSectionsAction: Page not found, skipping",
			zap.String("page_name", pageName),
			zap.Error(err),
		)
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"skipped":        true,
			"reason":         fmt.Sprintf("page not found: %s", pageName),
		}, nil
	}

	// Extract sections from HTML
	sections := saveSectionsExtractFromHTML(html, params.Logger)

	if len(sections) == 0 {
		params.Logger.Info("SavePageSectionsAction: No sections found in HTML",
			zap.String("page_name", pageName),
			zap.Int("html_len", len(html)),
		)
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"page_id":        pageID.String(),
			"reason":         "no sections found in HTML",
		}, nil
	}

	// Clear existing components for this page
	_, err = params.DB.ExecContext(ctx, `
		DELETE FROM page_components WHERE page_id = $1
	`, pageID)
	if err != nil {
		params.Logger.Warn("SavePageSectionsAction: Failed to clear existing components",
			zap.Error(err),
		)
		// Continue anyway
	}

	// Insert each section
	savedCount := 0
	for i, section := range sections {
		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO page_components (page_id, position, rendered_html, slot_name, build_status)
			VALUES ($1, $2, $3, $4, 'deployed')
		`, pageID, i+1, section.HTML, section.ComponentName)

		if err != nil {
			params.Logger.Warn("SavePageSectionsAction: Failed to insert section",
				zap.Int("position", i+1),
				zap.String("component", section.ComponentName),
				zap.Error(err),
			)
			continue
		}
		savedCount++
	}

	params.Logger.Info("SavePageSectionsAction: Complete",
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()),
		zap.Int("sections_found", len(sections)),
		zap.Int("sections_saved", savedCount),
	)

	return map[string]interface{}{
		"success":        true,
		"page_id":        pageID.String(),
		"page_name":      pageName,
		"sections_found": len(sections),
		"sections_saved": savedCount,
	}, nil
}

// SectionData holds extracted section data
type SectionData struct {
	ComponentName string
	HTML          string
	Position      int
}

// saveSectionsExtractFromHTML finds all <section> blocks
func saveSectionsExtractFromHTML(html string, logger *zap.Logger) []SectionData {
	var sections []SectionData

	// Match <section ...>...</section> blocks
	sectionRe := regexp.MustCompile(`(?is)<section([^>]*)>(.*?)</section>`)
	dataComponentRe := regexp.MustCompile(`data-component="([^"]+)"`)

	matches := sectionRe.FindAllStringSubmatch(html, -1)

	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		attrs := match[1]
		fullSection := match[0]

		// Extract component name from data-component attribute
		componentName := "section"
		if componentMatch := dataComponentRe.FindStringSubmatch(attrs); len(componentMatch) >= 2 {
			componentName = componentMatch[1]
		}

		sections = append(sections, SectionData{
			ComponentName: componentName,
			HTML:          strings.TrimSpace(fullSection),
			Position:      i + 1,
		})
	}

	logger.Debug("saveSectionsExtractFromHTML: Found sections",
		zap.Int("count", len(sections)),
	)

	return sections
}

// saveSectionsLookupPageID finds page UUID by site_id and page name
func saveSectionsLookupPageID(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string) (uuid.UUID, error) {
	var pageID uuid.UUID
	err := db.QueryRowContext(ctx, `
		SELECT id FROM pages WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&pageID)
	return pageID, err
}
