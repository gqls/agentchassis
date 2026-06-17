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
		metaArrayLen := -1
		if arr, isArr := metaData.([]interface{}); isArr {
			metaArrayLen = len(arr)
		}
		params.Logger.Info("SavePageSectionsAction: metadata field check",
			zap.String("field", metaField),
			zap.Bool("metadata_present", metaData != nil),
			zap.String("metadata_type", fmt.Sprintf("%T", metaData)),
			zap.Int("metadata_array_len", metaArrayLen))
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

	}

	// Enrich component IDs and section names from the page's planned sections array.
	// Runs for BOTH metadata and HTML paths — metadata path often lacks component_id,
	// and HTML path may have generic section names.
	if params.DB != nil && len(sections) > 0 {
		enrichSectionsWithPlannedNames(ctx, params.DB, pageID, sections, params.Logger)
		enrichSectionsWithComponentIDs(ctx, params.DB, sections, params.Logger)
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

	// DIAGNOSTIC: record what actually reached the save path — per-section HTML
	// lengths and the stripped-text total the regression guard will compute.
	// Logged unconditionally so these numbers are visible on a passing save too,
	// not only when the guard blocks.
	{
		diagStripper := regexp.MustCompile(`<[^>]*>`)
		diagTotal := 0
		diagPerSection := make([]string, 0, len(sections))
		for _, s := range sections {
			diagStripped := strings.TrimSpace(diagStripper.ReplaceAllString(s.HTML, ""))
			diagTotal += len(diagStripped)
			diagPerSection = append(diagPerSection,
				fmt.Sprintf("%s:html=%d,stripped=%d", s.ComponentName, len(s.HTML), len(diagStripped)))
		}
		params.Logger.Info("SavePageSectionsAction: sections reaching save",
			zap.String("page_name", pageName),
			zap.Int("section_count", len(sections)),
			zap.Int("stripped_text_total", diagTotal),
			zap.String("per_section", strings.Join(diagPerSection, " | ")),
		)
	}

	// --- Content regression guard ---
	// Refuse to overwrite content-rich pages with empty template shells.
	// This prevents LLM failures (credit exhaustion, timeouts, empty responses)
	// from wiping good content that was previously generated and deployed.
	{
		var existingTextLen int
		scanErr := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(
				LENGTH(REGEXP_REPLACE(rendered_html, '<[^>]*>', '', 'g'))
			), 0)
			FROM page_components
			WHERE page_id = $1 AND build_status = 'deployed'
		`, pageID).Scan(&existingTextLen)

		if scanErr == nil && existingTextLen > 200 {
			newTextLen := 0
			tagStripper := regexp.MustCompile(`<[^>]*>`)
			for _, s := range sections {
				stripped := tagStripper.ReplaceAllString(s.HTML, "")
				stripped = strings.TrimSpace(stripped)
				newTextLen += len(stripped)
			}

			if newTextLen < existingTextLen/4 {
				params.Logger.Warn("SavePageSectionsAction: CONTENT REGRESSION BLOCKED — new content has much less text than existing",
					zap.String("page_name", pageName),
					zap.Int("existing_text_chars", existingTextLen),
					zap.Int("new_text_chars", newTextLen),
					zap.Int("new_sections", len(sections)),
				)
				return nil, fmt.Errorf(
					"content regression blocked: new content has %d chars of text vs %d existing (page: %s). "+
						"This usually means the LLM returned empty content. Refusing to overwrite.",
					newTextLen, existingTextLen, pageName)
			}
		}
	}

	// Load page purpose for content_brief population
	var pagePurpose string
	_ = params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(page_spec->>'purpose', '') FROM pages WHERE id = $1
	`, pageID).Scan(&pagePurpose)

	// --- Snapshot existing content to history before overwrite ---
	_, snapshotErr := params.DB.ExecContext(ctx, `
		INSERT INTO page_component_history (component_id, page_id, site_id, content_data, source)
		SELECT pc.id, pc.page_id, p.site_id,
			   COALESCE(pc.content_data, jsonb_build_object(
				   'rendered_html', pc.rendered_html,
				   'slot_name', pc.slot_name,
				   'build_status', pc.build_status
			   )),
			   'save_page_sections_overwrite'
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.page_id = $1
		  AND pc.rendered_html IS NOT NULL
		  AND LENGTH(pc.rendered_html) > 0
	`, pageID)
	if snapshotErr != nil {
		params.Logger.Warn("SavePageSectionsAction: Failed to snapshot existing content to history",
			zap.Error(snapshotErr),
		)
		// Non-blocking — continue with the save even if history write fails
	} else {
		params.Logger.Info("SavePageSectionsAction: Snapshotted existing content to history",
			zap.String("page_name", pageName),
		)
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

		// Build content_brief from page purpose and section name
		var contentBriefJSON interface{} // nil = SQL NULL
		if pagePurpose != "" || section.ComponentName != "" {
			brief := map[string]string{
				"purpose":          pagePurpose,
				"tone_direction":   "",
				"section_guidance": section.ComponentName + " section",
			}
			if briefBytes, err := json.Marshal(brief); err == nil {
				contentBriefJSON = string(briefBytes)
			}
		}

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO page_components (page_id, position, rendered_html, slot_name, component_id, content_data, content_brief, build_status)
			VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, 'deployed')
		`, pageID, i+1, section.HTML, section.ComponentName, componentIDPtr, contentDataJSON, contentBriefJSON)

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

	skippedNotMap := 0
	skippedEmptyHTML := 0

	for i, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			skippedNotMap++
			continue
		}

		html, _ := m["rendered_html"].(string)
		if html == "" {
			skippedEmptyHTML++
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

	logger.Info("extractSectionsFromMetadata: parsed metadata array",
		zap.Int("items_in", len(items)),
		zap.Int("sections_out", len(sections)),
		zap.Int("skipped_not_map", skippedNotMap),
		zap.Int("skipped_empty_rendered_html", skippedEmptyHTML),
	)

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

	logger.Info("saveSectionsExtractFromHTML: input",
		zap.Int("html_length", len(html)),
		zap.Int("section_matches", len(matches)),
	)

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

	// Fallback: no <section> blocks found but the HTML has content.
	// Recreated tools/games (tool-recreation-handler) emit their body as
	// <div class="tool-page">…</div> with no <section> element, so the regex
	// above matches nothing. Returning empty here leaves the page with zero
	// page_components, which makes the rerender's getPageSections return empty
	// and skip the page — no git commit, no deployed file — while only logging
	// "no sections". Instead, store the whole fragment as a single section so it
	// deploys through the existing insert path.
	//
	// Guard: only do this for a content fragment, not a full document. If the
	// HTML carries <html>/<!doctype> it is an assembled page (header + footer +
	// chrome); wrapping that as one "section" would make the rerender double-wrap
	// site chrome. tool-recreation passes chrome-free inner HTML (its prompt
	// forbids <html>/<head>/<body>), so this fires exactly on the single-fragment
	// case it needs to and not on assembled pages.
	if len(sections) == 0 {
		trimmed := strings.TrimSpace(html)
		lower := strings.ToLower(trimmed)
		if trimmed != "" && !strings.Contains(lower, "<html") && !strings.Contains(lower, "<!doctype") {
			// Default name "section" so the existing enrichment path
			// (enrichSectionsWithPlannedNames / enrichSectionsWithComponentIDs)
			// can refine it from pages.sections / data-component, exactly as for
			// the <section> path. Tool HTML has no data-component, so it stays
			// "section" unless a planned name exists.
			componentName := "section"
			if componentMatch := dataComponentRe.FindStringSubmatch(trimmed); len(componentMatch) >= 2 {
				componentName = componentMatch[1]
			}
			sections = append(sections, SectionData{
				ComponentName: componentName,
				HTML:          trimmed,
				Position:      1,
			})
			logger.Info("saveSectionsExtractFromHTML: no <section> blocks found; stored whole fragment as one section",
				zap.String("component_name", componentName),
				zap.Int("html_length", len(trimmed)),
			)
		}
	}

	logger.Info("saveSectionsExtractFromHTML: Found sections",
		zap.Int("count", len(sections)),
	)

	return sections
}

// enrichSectionsWithComponentIDs looks up component_id from content_components
// for sections that have a component name but no component_id.
// Handles naming mismatches:
//
//	slot_name "social-proof" → function "social_proof" (hyphen vs underscore)
//	slot_name "case-studies-hero" → function "hero" with name matching
//	slot_name "differentiators-section" → function "differentiators" (suffix strip)
//	metadata ComponentName differs from data-component attr → prefer HTML attr
func enrichSectionsWithComponentIDs(ctx context.Context, db *sql.DB, sections []SectionData, logger *zap.Logger) {
	logger.Info("enrichSectionsWithComponentIDs: invoked",
		zap.Int("section_count", len(sections)),
		zap.Bool("db_nil", db == nil))

	dataComponentRe := regexp.MustCompile(`data-component="([^"]+)"`)

	for i := range sections {
		if sections[i].ComponentID != "" {
			continue // already has an ID
		}

		// Extract the data-component attribute from the rendered HTML first —
		// this is the authoritative name that matches content_components.function
		// and lets us recover when ComponentName is missing or generic ("section").
		htmlComponentName := ""
		if m := dataComponentRe.FindStringSubmatch(sections[i].HTML); len(m) >= 2 {
			htmlComponentName = m[1]
		}

		// If ComponentName is empty or the generic default "section", adopt the
		// HTML data-component value as the name. If neither is usable, skip.
		if sections[i].ComponentName == "" || sections[i].ComponentName == "section" {
			if htmlComponentName == "" || htmlComponentName == "section" {
				logger.Info("enrichSectionsWithComponentIDs: skipping — no usable name",
					zap.Int("position", i+1),
					zap.String("component_name", sections[i].ComponentName))
				continue
			}
			logger.Info("enrichSectionsWithComponentIDs: adopted data-component as name",
				zap.String("old_name", sections[i].ComponentName),
				zap.String("html_name", htmlComponentName),
				zap.Int("position", i+1))
			sections[i].ComponentName = htmlComponentName
		}

		slotName := sections[i].ComponentName

		// Build list of candidate names to try, in priority order.
		// If metadata name differs from the HTML data-component value, prefer
		// the HTML value — the metadata path may produce a different name
		// (e.g. "differentiators-section" from component_function while the
		// HTML has data-component="differentiators").
		candidates := []string{slotName}
		if htmlComponentName != "" && htmlComponentName != slotName {
			// Prefer the HTML data-component value — it matches what renders
			candidates = []string{htmlComponentName, slotName}
			// Also update the slot_name to match the HTML for consistency
			sections[i].ComponentName = htmlComponentName
			logger.Info("enrichSectionsWithComponentIDs: preferring data-component over metadata",
				zap.String("metadata_name", slotName),
				zap.String("html_name", htmlComponentName),
				zap.Int("position", i+1))
		}

		// Add suffix-stripped variants (differentiators-section → differentiators)
		for _, name := range []string{slotName, htmlComponentName} {
			if name == "" {
				continue
			}
			for _, suffix := range []string{"-section", "-container", "-wrapper", "-block"} {
				if strings.HasSuffix(name, suffix) {
					stripped := strings.TrimSuffix(name, suffix)
					candidates = append(candidates, stripped)
				}
			}
		}

		var componentID string
		var matchedBy string

		for _, candidate := range candidates {
			if candidate == "" {
				continue
			}

			// Try exact match
			err := db.QueryRowContext(ctx, `
				SELECT id::text FROM content_components 
				WHERE function = $1 AND is_active = true
				LIMIT 1
			`, candidate).Scan(&componentID)
			if err == nil {
				matchedBy = "exact:" + candidate
				break
			}
			if err != sql.ErrNoRows {
				logger.Warn("enrichSectionsWithComponentIDs: exact-match query error",
					zap.String("candidate", candidate),
					zap.Error(err))
			}

			// Try underscore variant (social-proof → social_proof)
			underscored := strings.ReplaceAll(candidate, "-", "_")
			if underscored != candidate {
				err = db.QueryRowContext(ctx, `
					SELECT id::text FROM content_components 
					WHERE function = $1 AND is_active = true
					LIMIT 1
				`, underscored).Scan(&componentID)
				if err == nil {
					matchedBy = "underscore:" + underscored
					break
				}
				if err != sql.ErrNoRows {
					logger.Warn("enrichSectionsWithComponentIDs: underscore-variant query error",
						zap.String("candidate", underscored),
						zap.Error(err))
				}
			}
		}

		// Try specialized hero variant (case-studies-hero → hero with name match)
		if componentID == "" && strings.HasSuffix(slotName, "-hero") {
			prefix := strings.TrimSuffix(slotName, "-hero")
			namePattern := "%" + strings.ReplaceAll(prefix, "-", "%") + "%"
			err := db.QueryRowContext(ctx, `
				SELECT id::text FROM content_components 
				WHERE function = 'hero' AND is_active = true
				  AND lower(name) LIKE lower($1)
				LIMIT 1
			`, namePattern).Scan(&componentID)
			if err == nil {
				matchedBy = "hero-variant:" + prefix
			} else if err != sql.ErrNoRows {
				logger.Warn("enrichSectionsWithComponentIDs: hero-variant query error",
					zap.String("pattern", namePattern),
					zap.Error(err))
			}
		}

		if componentID != "" {
			sections[i].ComponentID = componentID
			logger.Info("enrichSectionsWithComponentIDs: linked component",
				zap.String("slot_name", sections[i].ComponentName),
				zap.String("component_id", componentID),
				zap.String("matched_by", matchedBy),
				zap.Int("position", i+1))
		} else {
			logger.Info("enrichSectionsWithComponentIDs: no match found",
				zap.String("slot_name", slotName),
				zap.String("html_component", htmlComponentName),
				zap.Strings("candidates_tried", candidates),
				zap.Int("position", i+1))
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
		logger.Info("enrichSectionsWithPlannedNames: no sections array on page",
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
