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

	// Look up page_id (and the page's own url, best-effort origin metadata for
	// the link-repair record written before the insert below)
	pageID, pageURL, err := saveSectionsLookupPageID(ctx, params.DB, siteID, pageName)
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

	// --- Page-ownership guard (guard rail 1, experience loop) ---
	// A rebuild_policy='owned' page belongs to a tool/widget or is a
	// runtime-fill shell; this action's DELETE-and-reinsert of page_components
	// is exactly the TL-001 clobber. Refuse loudly rather than fall through to
	// the heuristic guards below: targeted edits go through apply_section_edit,
	// tool rebuilds through the tool pipeline. (Column ships in migration 164;
	// a scan error means an older schema, in which case the guard stands down.)
	{
		var rebuildPolicy string
		if scanErr := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(rebuild_policy, 'generic') FROM pages WHERE id = $1
		`, pageID).Scan(&rebuildPolicy); scanErr == nil && rebuildPolicy == "owned" {
			params.Logger.Warn("SavePageSectionsAction: OWNED PAGE — generic section save refused",
				zap.String("page_name", pageName),
				zap.String("page_id", pageID.String()),
			)
			return nil, fmt.Errorf(
				"page %s is rebuild_policy=owned (tool/widget-owned): a generic section save "+
					"would clobber it. Use apply_section_edit for targeted edits or the tool "+
					"pipeline for rebuilds. Refusing to overwrite.",
				pageName)
		}
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

	// --- Preserve interactive tool sections (Layer 2) ---
	// Interactive tools (games/simulators) exist ONLY as rendered_html in
	// page_components — their bespoke <canvas>/JS markup is not in the page
	// spec and is not LLM-regeneratable. A full rebuild plans sections from
	// the spec, which omits the tool, so the section set reaching this save
	// carries no interactive markup and the DELETE+INSERT below would drop it.
	// Carry any existing deployed interactive section the new set does not
	// reproduce forward, so a rebuild keeps the tool while still updating the
	// other sections. This is the only place with the rendered markup to
	// preserve; the planning path deals in section-name skeletons and cannot
	// reconstruct the tool.
	{
		rows, qErr := params.DB.QueryContext(ctx, `
			SELECT slot_name, rendered_html, content_data
			FROM page_components
			WHERE page_id = $1 AND build_status = 'deployed'
			  AND (rendered_html ILIKE '%<canvas%'
			    OR rendered_html ILIKE '%game-container%'
			    OR rendered_html ILIKE '%tool-page%')
		`, pageID)
		if qErr != nil {
			params.Logger.Warn("SavePageSectionsAction: interactive-section preload failed (Layer 2)",
				zap.Error(qErr))
		} else {
			type preservedSection struct {
				slot        string
				html        string
				contentData map[string]interface{}
			}
			var preserved []preservedSection
			for rows.Next() {
				var slot, html string
				var cdJSON []byte
				if scanErr := rows.Scan(&slot, &html, &cdJSON); scanErr != nil {
					params.Logger.Warn("SavePageSectionsAction: interactive-section scan failed (Layer 2)",
						zap.Error(scanErr))
					continue
				}
				var cd map[string]interface{}
				if len(cdJSON) > 0 {
					_ = json.Unmarshal(cdJSON, &cd)
				}
				preserved = append(preserved, preservedSection{slot: slot, html: html, contentData: cd})
			}
			rows.Close()

			for _, p := range preserved {
				matchedIdx := -1
				for i := range sections {
					if sections[i].ComponentName == p.slot {
						matchedIdx = i
						break
					}
				}
				switch {
				case matchedIdx >= 0 && sectionHTMLIsInteractive(sections[matchedIdx].HTML):
					// Rebuild reproduced an interactive section for this slot — keep it.
				case matchedIdx >= 0:
					// Same slot, non-interactive rebuild content (e.g. the hero
					// regenerated as plain text). Keep the existing interactive
					// markup in place rather than overwriting the tool with prose.
					params.Logger.Warn("SavePageSectionsAction: preserving interactive tool over non-interactive rebuild (Layer 2)",
						zap.String("page_name", pageName),
						zap.String("slot_name", p.slot))
					sections[matchedIdx].HTML = p.html
					if p.contentData != nil {
						sections[matchedIdx].ContentData = p.contentData
					}
				default:
					// Slot dropped entirely — re-append the tool so it survives.
					params.Logger.Warn("SavePageSectionsAction: re-appending dropped interactive tool (Layer 2)",
						zap.String("page_name", pageName),
						zap.String("slot_name", p.slot))
					sections = append(sections, SectionData{
						ComponentName: p.slot,
						HTML:          p.html,
						ContentData:   p.contentData,
						Position:      len(sections) + 1,
					})
				}
			}
		}
	}

	// --- Repair dead internal links before anything is persisted (bugs_open/079) ---
	// Persistence is the enforcement point: the build gate's repair lives in
	// `clean_html`, which the structured metadata path above never reads, so a
	// gate-side repair is dead config on the primary build plan and absent
	// entirely from page-rerender's structured save. Running here means no
	// save_page_sections INVOCATION can persist an unrepaired section, whatever
	// its workflow config says — and that is the whole of the claim. It is NOT
	// fleet-wide coverage of page_components: ten Go call sites write
	// rendered_html, and three of them persist LLM-authored prose with no repair
	// at all (ApplySectionEditAction, create_report_page, rebuild_blog_listing —
	// bugs_open/136). An earlier draft of this comment said "no build path",
	// which is the over-broad wording the council caught; it is corrected here
	// because the next reader of this file would otherwise inherit it.
	// Placed AFTER the interactive-tool preservation block — so the stored markup
	// that block carries forward is repaired too — and BEFORE the guards below,
	// so they measure the bytes that will actually be written. Unlinking keeps
	// the anchor text, so the guards' stripped-text totals are unchanged by it.
	// See save_sections_link_repair.go.
	repairSectionsBeforePersist(ctx, params, siteID,
		datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
		pageName, pageURL, sections, params.Logger)

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

	// --- Interactivity regression guard (Layer 1) ---
	// The text guard above misses the loss of an interactive tool: the tool is
	// mostly markup + JS, so tag-stripping leaves little text and a plain-
	// content replacement can have MORE text — passing the text check while
	// silently dropping a <canvas>/game-container/tool-page section. If Layer 2
	// above could not preserve it (e.g. the preload query failed), refuse the
	// overwrite rather than destroy the tool: a blocked save leaves the existing
	// good page in place and surfaces a loud error to investigate.
	{
		var existingInteractive bool
		scanErr := params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(bool_or(
				rendered_html ILIKE '%<canvas%'
			 OR rendered_html ILIKE '%game-container%'
			 OR rendered_html ILIKE '%tool-page%'
			), false)
			FROM page_components
			WHERE page_id = $1 AND build_status = 'deployed'
		`, pageID).Scan(&existingInteractive)

		if scanErr == nil && existingInteractive {
			newInteractive := false
			for _, s := range sections {
				if sectionHTMLIsInteractive(s.HTML) {
					newInteractive = true
					break
				}
			}
			if !newInteractive {
				params.Logger.Warn("SavePageSectionsAction: INTERACTIVITY REGRESSION BLOCKED — existing page has an interactive tool, new content has none",
					zap.String("page_name", pageName),
					zap.Int("new_sections", len(sections)),
				)
				return nil, fmt.Errorf(
					"interactivity regression blocked: page %s currently has an interactive tool "+
						"(<canvas>/game-container/tool-page) but the rebuilt content has none. "+
						"A full rebuild planned from the page spec does not include the tool. Refusing to overwrite.",
					pageName)
			}
		}
	}

	// Load page purpose for content_brief population
	var pagePurpose string
	_ = params.DB.QueryRowContext(ctx, `
		SELECT COALESCE(page_spec->>'purpose', '') FROM pages WHERE id = $1
	`, pageID).Scan(&pagePurpose)

	// Optional: stamp the work item driving this save into history for
	// attribution. The overwrite previously wrote NULL source_item_id, which
	// forced forensic-by-timing when tracing a destructive rebuild. Config-
	// driven path into collected_data; nil leaves SQL NULL (unchanged
	// behaviour) until the workflow sets work_item_id_field.
	var sourceItemID interface{} // nil = SQL NULL
	if f, ok := config["work_item_id_field"].(string); ok && f != "" {
		if v := datahelpers.ExtractNestedFieldString(params.CollectedData, f); v != "" {
			if parsed, parseErr := uuid.Parse(v); parseErr == nil {
				sourceItemID = parsed
			}
		}
	}

	// --- Snapshot existing content to history before overwrite ---
	_, snapshotErr := params.DB.ExecContext(ctx, `
		INSERT INTO page_component_history (component_id, page_id, site_id, content_data, source, source_item_id)
		SELECT pc.id, pc.page_id, p.site_id,
			   COALESCE(pc.content_data, jsonb_build_object(
				   'rendered_html', pc.rendered_html,
				   'slot_name', pc.slot_name,
				   'build_status', pc.build_status
			   )),
			   'save_page_sections_overwrite', $2::uuid
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.page_id = $1
		  AND pc.rendered_html IS NOT NULL
		  AND LENGTH(pc.rendered_html) > 0
	`, pageID, sourceItemID)
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

	// --- Locked-section preservation (bugs_open/058) ---
	// Human-locked rows must survive the rebuild with copy AND row identity
	// intact (admin lock/unlock addresses rows by id; "unchanged" means
	// updated_at too). So: preload the actively-locked rows, delete only the
	// agent-writable ones, and skip the incoming section that would have
	// replaced a locked slot — the locked copy stands, only its position moves
	// to follow the new composition. Expiry-aware predicate is the single
	// shared one (lock_helpers.go).
	lockedRows := loadActiveLockedRows(ctx, params.DB, pageID, params.Logger)

	// Clear existing components for this page — except actively-locked rows
	_, err = params.DB.ExecContext(ctx, `
		DELETE FROM page_components WHERE page_id = $1 AND `+pageComponentAgentWritableSQL(""),
		pageID)
	if err != nil {
		params.Logger.Warn("SavePageSectionsAction: Failed to clear existing components",
			zap.Error(err),
		)
		// Continue anyway
	}

	// Insert each section
	savedCount := 0
	var skippedStubs []string
	var lockedSlotsPreserved []string
	for i, section := range sections {
		// ── Locked-slot guard (bugs_open/058) ────────────────────────────────
		// The new composition produced fresh copy for a slot a human has
		// locked: the locked row (kept out of the DELETE above) stands, the
		// fresh copy is discarded, and the block is surfaced as a work item —
		// a silent skip would trade one silent failure for another. Only the
		// row's position moves, so ordering follows the new composition.
		if lr := matchLockedRow(lockedRows, section.ComponentName); lr != nil {
			lr.consumed = true
			if _, posErr := params.DB.ExecContext(ctx, `
				UPDATE page_components SET position = $2 WHERE id = $1
			`, lr.id, i+1); posErr != nil {
				params.Logger.Warn("SavePageSectionsAction: failed to reposition locked section",
					zap.String("slot_name", lr.slot), zap.Error(posErr))
			}
			params.Logger.Warn("SavePageSectionsAction: preserving human-locked section over rebuilt copy (bugs_open/058)",
				zap.String("page_name", pageName),
				zap.String("slot_name", lr.slot),
				zap.String("locked_by", lr.lockedBy),
			)
			lockedSlotsPreserved = append(lockedSlotsPreserved, lr.slot)
			lrID := lr.id
			emitLockBlockedChangeItem(ctx, params.DB, siteID, &pageID, &lrID,
				pageName, lr.slot, lr.lockedBy, lr.lockType,
				"overwrite", "save_page_sections", params.Logger)
			continue
		}
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

		// ── Unresolvable-section guard (bugs_open/039) ───────────────────────
		// A section name that resolves to no component falls back to
		// generic-text-block; with no content it renders a hollow ~208-byte
		// <section class="section section--generic"> — an empty heading and an
		// empty body. Persisting that as a component_id=NULL,
		// build_status='deployed' row is the defect: the page ships a hollow
		// section and the build reports success (7 such stubs were live across
		// 3 sites when this guard was written). This is the single INSERT every
		// page-composition path flows through, so refuse here regardless of
		// which upstream path produced the stub. Skip the row and raise a
		// needs_new_component item (deduped per section_type per site, routed to
		// the component-creator) so the gap is legible instead of rotting as an
		// unconsumed empty_section finding. The discriminator is deliberately
		// narrow — an empty generic stub AND no component link — so it never
		// touches a resolved component, a generic block that DID receive content
		// (those carry visible text), or a non-generic orphan.
		if componentIDPtr == nil && isEmptyGenericStub(section.HTML) {
			params.Logger.Error("SavePageSectionsAction: refusing to persist an empty generic stub for an unresolved section",
				zap.String("page_name", pageName),
				zap.String("slot_name", section.ComponentName),
				zap.Int("position", i+1),
			)
			if cErr := CreateNeedsNewComponentItem(ctx, params.DB, siteID.String(),
				section.ComponentName, pageName,
				fmt.Sprintf("Section %q on page %q resolves to no component and rendered an empty stub; a component template is needed (bugs_open/039).", section.ComponentName, pageName),
				"", "", params.Logger); cErr != nil {
				params.Logger.Warn("SavePageSectionsAction: failed to raise needs_new_component for stub section",
					zap.String("slot_name", section.ComponentName),
					zap.Error(cErr),
				)
			}
			skippedStubs = append(skippedStubs, section.ComponentName)
			continue
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

	// Locked rows the new composition no longer includes: the lock holds them
	// on the page (bugs_open/058 — automation may not remove human-locked
	// copy). Reposition after the new set in old-position order so the
	// assembler's bare ORDER BY position stays deterministic, and surface the
	// blocked removal.
	nextPos := len(sections) + 1
	for _, lr := range lockedRows {
		if lr.consumed {
			continue
		}
		if _, posErr := params.DB.ExecContext(ctx, `
			UPDATE page_components SET position = $2 WHERE id = $1
		`, lr.id, nextPos); posErr != nil {
			params.Logger.Warn("SavePageSectionsAction: failed to reposition retained locked section",
				zap.String("slot_name", lr.slot), zap.Error(posErr))
		}
		nextPos++
		params.Logger.Warn("SavePageSectionsAction: retaining human-locked section the new composition dropped (bugs_open/058)",
			zap.String("page_name", pageName),
			zap.String("slot_name", lr.slot),
			zap.String("locked_by", lr.lockedBy),
		)
		lockedSlotsPreserved = append(lockedSlotsPreserved, lr.slot)
		lrID := lr.id
		emitLockBlockedChangeItem(ctx, params.DB, siteID, &pageID, &lrID,
			pageName, lr.slot, lr.lockedBy, lr.lockType,
			"remove", "save_page_sections", params.Logger)
	}

	params.Logger.Info("SavePageSectionsAction: Complete",
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()),
		zap.Int("sections_found", len(sections)),
		zap.Int("sections_saved", savedCount),
		zap.Int("skipped_stub_sections", len(skippedStubs)),
		zap.Int("locked_sections_preserved", len(lockedSlotsPreserved)),
	)

	return map[string]interface{}{
		"success":                   true,
		"page_id":                   pageID.String(),
		"page_name":                 pageName,
		"sections_found":            len(sections),
		"sections_saved":            savedCount,
		"skipped_stub_sections":     skippedStubs,
		"locked_sections_preserved": lockedSlotsPreserved,
	}, nil
}

// lockedPageRow is an actively-locked page_components row preserved through a
// rebuild (bugs_open/058).
type lockedPageRow struct {
	id       uuid.UUID
	slot     string
	position int
	lockedBy string
	lockType string
	consumed bool // matched (and thereby blocked) an incoming section this save
}

// loadActiveLockedRows returns the page's rows automation may not overwrite,
// in position order. Best-effort: on query failure it returns nil and the save
// proceeds exactly as before this guard existed (the DELETE predicate still
// protects the rows themselves; only slot-matching is lost, which would leave
// a duplicate slot rather than destroy locked copy).
func loadActiveLockedRows(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) []*lockedPageRow {
	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(slot_name, ''), position, COALESCE(locked_by, ''), COALESCE(lock_type, '')
		FROM page_components
		WHERE page_id = $1 AND NOT `+pageComponentAgentWritableSQL("")+`
		ORDER BY position ASC
	`, pageID)
	if err != nil {
		logger.Warn("SavePageSectionsAction: locked-row preload failed (bugs_open/058)", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var locked []*lockedPageRow
	for rows.Next() {
		lr := &lockedPageRow{}
		if scanErr := rows.Scan(&lr.id, &lr.slot, &lr.position, &lr.lockedBy, &lr.lockType); scanErr != nil {
			logger.Warn("SavePageSectionsAction: locked-row scan failed", zap.Error(scanErr))
			continue
		}
		locked = append(locked, lr)
	}
	return locked
}

// matchLockedRow finds the first unconsumed locked row whose slot matches the
// incoming section name — exact first, then kebab-normalised (the 041 naming
// landmine: the library stores kebab-case but older rows/plans may carry
// snake_case or CamelCase variants of the same slot). Each locked row matches
// at most one incoming section, so a page with duplicate slot names cannot
// have one lock swallow several sections.
func matchLockedRow(lockedRows []*lockedPageRow, sectionName string) *lockedPageRow {
	if sectionName == "" {
		return nil
	}
	for _, lr := range lockedRows {
		if !lr.consumed && lr.slot == sectionName {
			return lr
		}
	}
	norm := NormalizeComponentFunction(sectionName)
	if norm == "" {
		return nil
	}
	for _, lr := range lockedRows {
		if !lr.consumed && lr.slot != "" && NormalizeComponentFunction(lr.slot) == norm {
			return lr
		}
	}
	return nil
}

// isEmptyGenericStub reports whether rendered HTML is the hollow section that
// generic-text-block emits when it stands in for a section name that resolves
// to no component and receives no content: the `section--generic` marker with
// an empty heading and empty body (~208 bytes). It is intentionally strict —
// the class marker must be present AND stripping tags must leave no visible
// text — so a generic block that DID receive content is never matched, only the
// unresolvable-section stub. See bugs_open/039.
func isEmptyGenericStub(html string) bool {
	if !strings.Contains(html, "section--generic") {
		return false
	}
	return strings.TrimSpace(stripHTMLTags(html)) == ""
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

// sectionHTMLIsInteractive reports whether a section's rendered HTML carries an
// interactive tool/game. Same signal used by the blast-radius sweep, the
// Layer 2 carry-forward, and the Layer 1 interactivity guard.
func sectionHTMLIsInteractive(html string) bool {
	h := strings.ToLower(html)
	return strings.Contains(h, "<canvas") ||
		strings.Contains(h, "game-container") ||
		strings.Contains(h, "tool-page")
}

// saveSectionsLookupPageID finds page UUID by site_id and page name, and
// returns the page's stored url alongside it. The url is best-effort origin
// metadata for the link-repair record (bugs_open/079) — pages.url is nullable,
// so an absent one degrades to "" rather than failing the lookup.
func saveSectionsLookupPageID(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string) (uuid.UUID, string, error) {
	var pageID uuid.UUID
	var pageURL sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT id, url FROM pages WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&pageID, &pageURL)
	return pageID, pageURL.String, err
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
