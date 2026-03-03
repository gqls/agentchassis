// SavePageSectionsAction saves rendered HTML sections to page_components table
// This ensures rerender can reassemble pages from stored sections
//
// Called after deploy_page in pageflow-builder's build_pages_loop
//
// PATCH NOTES (2026-02-17):
// - Primary path: uses structured sections_metadata from CompilePageSectionsAction.
//   Each entry has rendered_html (with inline <style>), component_id, component_function.
//   No HTML parsing needed — data flows through from RenderComponentAction.
// - Fallback path: regex parsing of assembled HTML (for adopted sites or older pipelines).
//   Now also captures <style>/<script> blocks that follow </section>.
// - INSERT now sets component_id when available.
// - Fallback path looks up component_id from content_components.function matching
//   the data-component attribute.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
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
//   - sections_metadata_field: path to structured sections array (e.g. "page_content.response.sections_metadata")
//   - input_fields: alternative - array of field names to extract
//
// If sections_metadata_field is set and data exists, uses structured path (no parsing).
// Otherwise falls back to HTML parsing.
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

	// --- Resolve page name and site_id (needed for both paths) ---

	var pageName, siteIDStr string

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
	if pageName == "" || siteIDStr == "" {
		inputFields := []string{"page_content", "site_record", "current_page"}
		if fields, ok := config["input_fields"].([]interface{}); ok {
			inputFields = make([]string, len(fields))
			for i, f := range fields {
				inputFields[i], _ = f.(string)
			}
		}

		extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)

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

	// --- Try structured metadata path first ---

	var sections []SectionData

	if metaField, ok := config["sections_metadata_field"].(string); ok && metaField != "" {
		metaData := datahelpers.ExtractNestedField(params.CollectedData, metaField)
		if metaData != nil {
			sections = extractSectionsFromMetadata(metaData, params.Logger)
			if len(sections) > 0 {
				params.Logger.Info("SavePageSectionsAction: Using structured metadata path",
					zap.String("metadata_field", metaField),
					zap.Int("sections", len(sections)),
				)
			}
		}
	}

	// --- Fallback to HTML parsing ---

	if len(sections) == 0 {
		htmlField := "assembled_page.html"
		if f, ok := config["html_field"].(string); ok && f != "" {
			htmlField = f
		}
		html := datahelpers.ExtractNestedFieldString(params.CollectedData, htmlField)

		// Fallback extraction for html
		if html == "" {
			inputFields := []string{"page_content", "site_record", "current_page"}
			if fields, ok := config["input_fields"].([]interface{}); ok {
				inputFields = make([]string, len(fields))
				for i, f := range fields {
					inputFields[i], _ = f.(string)
				}
			}
			extracted := datahelpers.ExtractFields(params.CollectedData, inputFields, params.Logger)
			if pageContent, ok := extracted["page_content"].(map[string]interface{}); ok {
				if response, ok := pageContent["response"].(map[string]interface{}); ok {
					html, _ = response["page_html"].(string)
				}
			}
		}

		if html == "" {
			params.Logger.Warn("SavePageSectionsAction: No HTML and no metadata found, skipping",
				zap.String("html_field", htmlField),
			)
			return map[string]interface{}{
				"success":        true,
				"sections_saved": 0,
				"skipped":        true,
				"reason":         "no HTML content and no sections metadata",
			}, nil
		}

		sections = saveSectionsExtractFromHTML(html, params.Logger)
		params.Logger.Info("SavePageSectionsAction: Using HTML parsing fallback",
			zap.Int("sections", len(sections)),
		)

		// For HTML-parsed sections, try to look up component_id from content_components
		if params.DB != nil {
			enrichSectionsWithComponentIDs(ctx, params.DB, sections, params.Logger)
		}
	}

	// Enrich generic/empty section names from the page's planned sections array.
	// Runs for both metadata and HTML paths — either can produce generic names.
	if params.DB != nil && len(sections) > 0 {
		enrichSectionsWithPlannedNames(ctx, params.DB, pageID, sections, params.Logger)
	}

	if len(sections) == 0 {
		params.Logger.Info("SavePageSectionsAction: No sections found",
			zap.String("page_name", pageName),
		)
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"page_id":        pageID.String(),
			"reason":         "no sections found",
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
		// Dark section contract validation (warning only, non-blocking)
		// Auto-detects dark sections from CSS patterns in the HTML.
		if missing := ValidateDarkSectionContract(section.HTML, false, params.Logger); len(missing) > 0 {
			params.Logger.Warn("SavePageSectionsAction: Dark section missing --section-* variables",
				zap.String("slot_name", section.ComponentName),
				zap.Int("position", i+1),
				zap.Strings("missing_vars", missing),
			)
		}

		var componentIDPtr *uuid.UUID
		if section.ComponentID != "" {
			if parsed, err := uuid.Parse(section.ComponentID); err == nil {
				componentIDPtr = &parsed
			}
		}

		// Marshal content_data to JSON if present
		var contentDataJSON interface{} // nil = SQL NULL
		if section.ContentData != nil && len(section.ContentData) > 0 {
			if jsonBytes, err := json.Marshal(section.ContentData); err == nil {
				contentDataJSON = string(jsonBytes)
			} else {
				params.Logger.Warn("SavePageSectionsAction: Failed to marshal content_data",
					zap.Int("position", i+1),
					zap.Error(err),
				)
			}
		}

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO page_components (page_id, position, rendered_html, slot_name, component_id, content_data, build_status)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, 'deployed')
		`, pageID, i+1, section.HTML, section.ComponentName, componentIDPtr, contentDataJSON)

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
	ComponentID   string
	HTML          string
	Position      int
	ContentData   map[string]interface{} // structured content for re-rendering (source of truth)
}

// extractSectionsFromMetadata builds SectionData from the structured array
// produced by CompilePageSectionsAction's sections_metadata output.
func extractSectionsFromMetadata(metaData interface{}, logger *zap.Logger) []SectionData {
	var sections []SectionData

	items, ok := metaData.([]interface{})
	if !ok {
		logger.Warn("SavePageSectionsAction: sections_metadata is not an array",
			zap.String("type", fmt.Sprintf("%T", metaData)),
		)
		return nil
	}

	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		html, _ := m["rendered_html"].(string)
		if html == "" {
			continue
		}

		// component_function is the slot_name (e.g. "hero", "call-to-action")
		componentName := "section"
		if fn, ok := m["component_function"].(string); ok && fn != "" {
			componentName = fn
		} else if name, ok := m["component_name"].(string); ok && name != "" {
			componentName = name
		}

		// Enforce naming contract: slot_name must be kebab-case
		componentName = NormalizeComponentFunction(componentName)

		componentID := ""
		if id, ok := m["component_id"].(string); ok && id != "" {
			componentID = id
		} else if id, ok := m["component_id"]; ok && id != nil {
			componentID = fmt.Sprintf("%v", id)
		}

		// Extract content_data if present (from RenderComponentAction via CompilePageSectionsAction)
		var contentData map[string]interface{}
		if cd, ok := m["content_data"].(map[string]interface{}); ok {
			contentData = cd
		}

		sections = append(sections, SectionData{
			ComponentName: componentName,
			ComponentID:   componentID,
			HTML:          strings.TrimSpace(html),
			Position:      i + 1,
			ContentData:   contentData,
		})
	}

	return sections
}

// saveSectionsExtractFromHTML finds all <section> blocks with their trailing <style>/<script>
// CHANGED: regex now captures <style> and <script> blocks that follow </section>,
// since component templates place inline CSS after the closing </section> tag.
func saveSectionsExtractFromHTML(html string, logger *zap.Logger) []SectionData {
	var sections []SectionData

	// Match <section ...>...</section> followed by optional <style>...</style> and/or <script>...</script>
	// The (?:\s*<style>[\s\S]*?</style>)* captures zero or more style blocks after </section>
	// The (?:\s*<script>[\s\S]*?</script>)* captures zero or more script blocks after </section>
	sectionRe := regexp.MustCompile(
		`(?is)(<section[^>]*>.*?</section>)` +
			`((?:\s*<style[^>]*>[\s\S]*?</style>)*)` +
			`((?:\s*<script[^>]*>[\s\S]*?</script>)*)`,
	)
	dataComponentRe := regexp.MustCompile(`data-component="([^"]+)"`)

	matches := sectionRe.FindAllStringSubmatch(html, -1)

	for i, match := range matches {
		if len(match) < 2 {
			continue
		}

		sectionHTML := match[1]
		styleBlocks := ""
		scriptBlocks := ""
		if len(match) >= 3 {
			styleBlocks = match[2]
		}
		if len(match) >= 4 {
			scriptBlocks = match[3]
		}

		// Combine section + style + script into one stored unit
		fullHTML := sectionHTML
		if strings.TrimSpace(styleBlocks) != "" {
			fullHTML += "\n" + strings.TrimSpace(styleBlocks)
		}
		if strings.TrimSpace(scriptBlocks) != "" {
			fullHTML += "\n" + strings.TrimSpace(scriptBlocks)
		}

		// Extract component name from data-component attribute
		componentName := "section"
		if componentMatch := dataComponentRe.FindStringSubmatch(sectionHTML); len(componentMatch) >= 2 {
			componentName = componentMatch[1]
		}

		sections = append(sections, SectionData{
			ComponentName: componentName,
			HTML:          strings.TrimSpace(fullHTML),
			Position:      i + 1,
		})
	}

	logger.Debug("saveSectionsExtractFromHTML: Found sections",
		zap.Int("count", len(sections)),
	)

	return sections
}

// enrichSectionsWithComponentIDs looks up component_id from content_components
// for HTML-parsed sections that have a data-component attribute but no component_id.
// Handles naming mismatches:
//
//	slot_name "social-proof" → function "social_proof" (hyphen vs underscore)
//	slot_name "case-studies-hero" → function "hero" with name matching
func enrichSectionsWithComponentIDs(ctx context.Context, db *sql.DB, sections []SectionData, logger *zap.Logger) {
	for i := range sections {
		if sections[i].ComponentID != "" {
			continue // already has an ID
		}
		if sections[i].ComponentName == "" || sections[i].ComponentName == "section" {
			continue // no function name to look up
		}

		slotName := sections[i].ComponentName
		var componentID string

		// Try exact match first
		err := db.QueryRowContext(ctx, `
			SELECT id::text FROM content_components 
			WHERE function = $1 AND is_active = true
			LIMIT 1
		`, slotName).Scan(&componentID)

		// Try underscore variant (social-proof → social_proof)
		if err != nil {
			underscored := strings.ReplaceAll(slotName, "-", "_")
			if underscored != slotName {
				err = db.QueryRowContext(ctx, `
					SELECT id::text FROM content_components 
					WHERE function = $1 AND is_active = true
					LIMIT 1
				`, underscored).Scan(&componentID)
			}
		}

		// Try specialized hero variant (case-studies-hero → hero with name match)
		if err != nil && strings.HasSuffix(slotName, "-hero") {
			prefix := strings.TrimSuffix(slotName, "-hero")
			namePattern := "%" + strings.ReplaceAll(prefix, "-", "%") + "%"
			err = db.QueryRowContext(ctx, `
				SELECT id::text FROM content_components 
				WHERE function = 'hero' AND is_active = true
				  AND lower(name) LIKE lower($1)
				LIMIT 1
			`, namePattern).Scan(&componentID)
		}

		if err == nil && componentID != "" {
			sections[i].ComponentID = componentID
			logger.Debug("enrichSectionsWithComponentIDs: Found component",
				zap.String("slot_name", slotName),
				zap.String("component_id", componentID),
			)
		}
	}
}

// saveSectionsLookupPageID finds page UUID by site_id and page name
func saveSectionsLookupPageID(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string) (uuid.UUID, error) {
	var pageID uuid.UUID
	err := db.QueryRowContext(ctx, `
		SELECT id FROM pages WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&pageID)
	return pageID, err
}

// Problem: When saveSectionsExtractFromHTML can't find data-component attributes,
// ComponentName defaults to "section" (generic). This becomes the slot_name in
// page_components, making individual sections unaddressable by section-editor.
//
// Fix: After HTML extraction, enrich section names from pages.sections JSON array.
// pages.sections stores the planned section names in position order (1-indexed).
// If a section has a generic/empty ComponentName AND its position maps to the
// sections array, use the planned name instead.
//
// Call site: In SavePageSectionsAction, right after enrichSectionsWithComponentIDs:
//
//     enrichSectionsWithComponentIDs(ctx, params.DB, sections, params.Logger)
// +   enrichSectionsWithPlannedNames(ctx, params.DB, pageID, sections, params.Logger)
//

// enrichSectionsWithPlannedNames fills in generic/empty ComponentName values from
// the page's planned sections array (pages.sections column).
func enrichSectionsWithPlannedNames(ctx context.Context, db *sql.DB, pageID uuid.UUID, sections []SectionData, logger *zap.Logger) {
	// Count how many sections need enrichment
	needsEnrichment := 0
	for _, s := range sections {
		if s.ComponentName == "" || s.ComponentName == "section" {
			needsEnrichment++
		}
	}
	if needsEnrichment == 0 {
		return
	}

	// Load planned section names from pages.sections
	var sectionsJSON []byte
	err := db.QueryRowContext(ctx, `SELECT sections FROM pages WHERE id = $1`, pageID).Scan(&sectionsJSON)
	if err != nil || len(sectionsJSON) == 0 {
		logger.Debug("enrichSectionsWithPlannedNames: no sections array on page",
			zap.String("page_id", pageID.String()),
			zap.Error(err),
		)
		return
	}

	var planned []string
	if err := json.Unmarshal(sectionsJSON, &planned); err != nil {
		logger.Warn("enrichSectionsWithPlannedNames: failed to parse sections JSON",
			zap.Error(err))
		return
	}

	enriched := 0
	for i := range sections {
		if sections[i].ComponentName != "" && sections[i].ComponentName != "section" {
			continue // already has a meaningful name
		}
		// Position is 1-indexed, planned array is 0-indexed
		idx := sections[i].Position - 1
		if idx < 0 || idx >= len(planned) {
			continue
		}
		plannedName := NormalizeComponentFunction(planned[idx])
		if plannedName != "" {
			logger.Info("enrichSectionsWithPlannedNames: using planned section name",
				zap.Int("position", sections[i].Position),
				zap.String("old_name", sections[i].ComponentName),
				zap.String("planned_name", plannedName),
			)
			sections[i].ComponentName = plannedName
			enriched++
		}
	}

	if enriched > 0 {
		logger.Info("enrichSectionsWithPlannedNames: enriched sections",
			zap.Int("enriched", enriched),
			zap.Int("total", len(sections)),
		)
	}
}
