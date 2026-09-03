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

	// --- Ownership skip from assemble_page (bugs_open/208) ---
	//
	// The owned-page guard upstream refused to assemble this page, so there is
	// nothing to save — and this must be a SKIP, not the hard refusal below. None of
	// the three build loops sets continue_on_error (loop_error_handler.go:70-90
	// requires it), so an error here fails the whole workflow and strands every page
	// after this one. Before this check, one owned page in a set turned into "the
	// operator's entire rebuild dies".
	//
	// This does NOT weaken the ownership refusal below: it fires only on the guard's
	// own skip, so an owned page arriving WITH content — an older image, or a future
	// path that bypasses assemble_page — still meets that loud refusal.
	//
	// KEYED TO THE OWNERSHIP MARKER, not to any assembly skip, and the distinction
	// is load-bearing rather than fussy. An ordinary content-failure skip can arrive
	// with `sections_metadata` populated but no assembled HTML, and in that case the
	// metadata path below legitimately writes the sections — content_data being the
	// only thing a later re-render can regenerate from. Exiting early on every skip
	// would silently stop those writes on the fleet's highest-traffic save path,
	// which is a change this bug did not ask for.
	if skipped, skipReason := upstreamAssemblySkipped(params.CollectedData); skipped &&
		strings.Contains(skipReason, ownedPageSkipReasonPrefix) {
		params.Logger.Info("SavePageSectionsAction: owned-page guard skipped the assembly, nothing to save",
			zap.String("skip_reason", skipReason),
		)
		return map[string]interface{}{
			"success":        true,
			"sections_saved": 0,
			"skipped":        true,
			"reason":         skipReason,
		}, nil
	}

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
	//
	// UNIFIED 2026-08-06 (bugs_open/208, council `reuse_agent` seat): the predicate
	// now goes through pageIsOwnedForGuard, the single place ownership policy is
	// read. Same SQL, same fail-open-on-scan-error behaviour as the inline version
	// this replaced — the point is that there is one predicate rather than two
	// drifting copies of "is this page owned".
	{
		if refused, class, _ := pageRefusesGenericBuild(ctx, params.DB, pageID, params.Logger); refused {
			params.Logger.Warn("SavePageSectionsAction: PAGE REFUSES GENERIC BUILD — section save refused",
				zap.String("page_name", pageName),
				zap.String("page_id", pageID.String()),
				zap.String("refusal_class", class),
			)

			// bugs_open/295: refusing is right; refusing SILENTLY is not.
			//
			// The two sibling guards on this same predicate both record the refusal —
			// censusExcludedOwnedPages (selection) and AssemblePageAction (assemble) —
			// and this one did not. That gap is invisible on the normal build route,
			// where assemble_page refuses first and files the row. It is NOT invisible
			// on page-build-handler's route, which has no assemble_page step at all: a
			// content_rewrite work item reaches this action directly, dies `failed`, and
			// the only statement of why lives in the orchestration's __step_error, which
			// ages out at ~24h retention. Measured 2026-08-17: 0 owned_page_review rows
			// have ever carried refused_by='save_page_sections', against 172 of 704
			// pages estate-wide being rebuild_policy='owned'.
			//
			// UPDATED 2026-08-19: that zero was the BEFORE reading and it now reads as
			// a live claim, which it is not. This emit has filed 64 rows (2026-08-17 →
			// 08-18) — the mechanism works and is filing. Flagged by the
			// copy_quality_two_stage lane while reading this file; the figure is dated
			// rather than wrong, but a dated zero in a comment is indistinguishable
			// from a broken emit to anyone who does not check the date. Re-measure:
			//   SELECT count(*) FROM site_work_items WHERE item_type='owned_page_review'
			//     AND spec->>'refused_by'='save_page_sections';
			//
			// The work item still fails — the save genuinely did not happen, and saying
			// otherwise would be worse. What changes is that the refusal now leaves the
			// same deduped, human-routable row its siblings leave, whose spec.fix already
			// names apply_section_edit as the route that DOES work on an owned page.
			// Errors inside the emit are swallowed by design: reporting must never be
			// what breaks a guard.
			//
			// bugs_open/450 added the second refusal class. The owned wording is
			// preserved BYTE FOR BYTE — it is the text a live operator query and a
			// pinned test both read — and the tool_pending case gets its own, because
			// telling someone their page is "rebuild_policy=owned" when it is
			// 'generic' would send them to look at a column that says the opposite.
			reason := fmt.Sprintf(
				"%s: page %s is rebuild_policy=owned (tool/widget-owned); a generic "+
					"section save would clobber it. Use apply_section_edit for targeted "+
					"edits or the tool pipeline for rebuilds.",
				ownedPageSkipReasonPrefix, pageName)
			if class == refusalToolPending {
				reason = fmt.Sprintf(
					"%s: page %s is page_type=tool with no tool component; a generic "+
						"section save would publish prose about a tool that is not there. "+
						"The tool pipeline builds it (add_tool → tool-deployer), after which "+
						"this refusal lifts by itself.",
					ownedPageSkipReasonPrefix, pageName)
			}
			emitOwnedPageReviewItem(ctx, params.DB, siteID, pageName, "save_page_sections",
				reason, class, params.Logger)

			// The error LEADS with ownedPageSkipReasonPrefix (owner decision 1,
			// 2026-08-18). routeToErrorStep copies this message verbatim into
			// collected_data.__step_error.message, which is the only channel that
			// survives the action → coordinator → error_step boundary — an action
			// cannot name its own error step, and a typed error does not cross it.
			// So the marker is what lets the routed step tell THIS refusal apart
			// from a genuine save failure (a shrink guard, a banned claim), which
			// is what update_work_item_status' owned_page_refusal_status reads.
			// Without it every save error looks alike at the only place the
			// item's terminal status is chosen.
			if class == refusalToolPending {
				return nil, fmt.Errorf(
					"%s: page %s is page_type=tool with no tool component: a generic section save "+
						"would publish prose about a tool that is not there. The tool pipeline "+
						"builds the component (add_tool → tool-deployer) and this refusal then "+
						"lifts by itself. Refusing to overwrite.",
					ownedPageSkipReasonPrefix, pageName)
			}
			return nil, fmt.Errorf(
				"%s: page %s is rebuild_policy=owned (tool/widget-owned): a generic section save "+
					"would clobber it. Use apply_section_edit for targeted edits or the tool "+
					"pipeline for rebuilds. Refusing to overwrite.",
				ownedPageSkipReasonPrefix, pageName)
		}
	}

	// --- Try structured metadata path first ---
	//
	// WHICH path is no longer purely the caller's to remember (bugs_open/194).
	// With the key unset the action consults the same default the validate gate
	// uses, so a caller that says nothing gets the right answer instead of a
	// silent NULL; a caller that names a field still wins outright, which is what
	// keeps page-rerender's own path (rerender_sections.sections_metadata)
	// untouched. See save_sections_metadata_source.go for the three declared
	// states and why this is a single default rather than a probe.
	var sections []SectionData

	metaField, metaFieldOrigin := resolveSectionsMetadataField(config)
	sectionsSource := sectionsSourceHTMLParse

	// AMBIGUOUS CALLER (code-review F3): an explicit html_field says "my content
	// is an HTML blob", but saying nothing about metadata resolves to the
	// IMPLICIT default path, which is tried FIRST. If another step's reply
	// happens to sit there in collected_data, this save silently prefers it over
	// the field the caller actually named.
	//
	// Not reachable on any live caller, measured 2026-08-05: all SIX are
	// explicit — page-build-handler, pageflow-builder, page-rebuild,
	// page-rerender and site-work-orchestrator name sections_metadata_field;
	// tool-recreation-handler declares expects_no_sections_metadata. This fires
	// for a FUTURE caller that sets html_field and neither key, which is the one
	// combination the three declared states do not cover.
	//
	// > **CORRECTED 2026-08-05, same day.** First measured as "all three
	// > callers" from a TOP-LEVEL jsonb_each over {workflow,steps}. That finds 3
	// > of 6: the step is nested inside a loop sub_workflow for the rest, and
	// > LANDMINES already warned this exact query under-reports here. The
	// > conclusion was right and the evidence was not. Use the nested walk:
	// > `FROM agent_definitions ad, LATERAL jsonb_path_query(ad.default_config,
	// > '$.**.steps') AS steps, LATERAL jsonb_each(steps) AS s(key,value)
	// > WHERE s.value->>'action'='save_page_sections'`.
	//
	// A warning, not a refusal or a resolution change: altering which path wins
	// is a semantics change on the fleet's highest-traffic save path, and it
	// would be made on a prediction rather than a measurement — the same reason
	// writeContentDataRegressionLog records instead of refusing.
	if metaFieldOrigin == metadataOriginDefault {
		if hf, ok := config["html_field"].(string); ok && hf != "" {
			params.Logger.Warn("SavePageSectionsAction: caller names html_field but declares nothing about sections_metadata — "+
				"the implicit default path is consulted FIRST and may pre-empt it (code-review F3). "+
				"Set sections_metadata_field, or declare expects_no_sections_metadata if this caller has none",
				zap.String("html_field", hf),
				zap.String("implicit_metadata_field", metaField))
		}
	}

	if metaField != "" {
		metaData := datahelpers.ExtractNestedField(params.CollectedData, metaField)
		metaArrayLen := -1
		if arr, isArr := metaData.([]interface{}); isArr {
			metaArrayLen = len(arr)
		}
		params.Logger.Info("SavePageSectionsAction: metadata field check",
			zap.String("field", metaField),
			zap.String("field_origin", metaFieldOrigin),
			zap.Bool("metadata_present", metaData != nil),
			zap.String("metadata_type", fmt.Sprintf("%T", metaData)),
			zap.Int("metadata_array_len", metaArrayLen))
		if metaData != nil {
			sections = extractSectionsFromMetadata(metaData, params.Logger)
			if len(sections) > 0 {
				sectionsSource = sectionsSourceMetadata
				params.Logger.Info("SavePageSectionsAction: Using structured metadata path",
					zap.String("metadata_field", metaField),
					zap.String("metadata_field_origin", metaFieldOrigin),
					zap.Int("sections", len(sections)),
				)
			}
		}
	}

	// --- Declared expectation: absence is a refusal, for callers that opt in ---
	//
	// OFF by default and seeded on nobody in the change that introduced it
	// (RFC_010, 2026-08-02: new authority on a shared seam ships as an opt-in
	// field with the unsafe default OFF). A caller that knows its writer always
	// returns sections can set refuse_save_without_sections_metadata and get a loud failure
	// instead of a page saved without its regeneration source. Refusing HERE, before
	// the history snapshot and the DELETE, means a refused save writes nothing at
	// all — the same placement rule the completeness floor states below.
	if len(sections) == 0 && configBoolOrDefault(config, refuseSaveWithoutMetadataKey, false) {
		params.Logger.Error("SavePageSectionsAction: required sections_metadata absent — refusing the save",
			zap.String("page_name", pageName),
			zap.String("metadata_field", metaField),
			zap.String("metadata_field_origin", metaFieldOrigin))
		return nil, fmt.Errorf(
			"this step declares %s but no usable sections_metadata arrived at %q (%s): saving from parsed "+
				"HTML would persist the page without the content_data the rerender path regenerates from. "+
				"Refusing (bugs_open/194)",
			refuseSaveWithoutMetadataKey, metaField, metaFieldOrigin)
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
		// The planned-name pass is UNCHANGED and still runs: the slot name is the
		// page's positional identity, it keeps the row in step with pages.sections,
		// and it is the key Layer 2 matches on. Only the COMPONENT binding below
		// learns to decline (RFC_046 / bugs_open/357 — the damage is on the
		// component axis, the landmine is on the slot axis).
		enrichSectionsWithPlannedNames(ctx, params.DB, pageID, sections, params.Logger)
		enrichSectionsWithComponentIDs(ctx, params.DB, sections, params.Logger,
			configBoolOrDefault(config, adoptFragmentsKey, false))
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
	// lengths and the text totals the floors below will compute.
	// Logged unconditionally so these numbers are visible on a passing save too,
	// not only when the guard blocks.
	//
	// It reports BOTH axes on purpose (bugs_open/293). This line used to advertise
	// itself as "the stripped-text total the regression guard will compute", and
	// tag-stripped length stopped being that the moment the floors moved to visible
	// text — a diagnostic that names the wrong quantity is worse than none, because
	// whoever debugs a refusal reads it as the guard's own arithmetic. Keeping the
	// retired axis alongside is what makes a refusal legible: the signature of this
	// bug's failure mode is html and stripped GROWING while visible COLLAPSES.
	{
		diagStripper := regexp.MustCompile(`<[^>]*>`)
		diagTotal := 0
		diagVisibleTotal := 0
		diagPerSection := make([]string, 0, len(sections))
		for _, s := range sections {
			diagStripped := strings.TrimSpace(diagStripper.ReplaceAllString(s.HTML, ""))
			diagVisible := visibleTextLength(s.HTML)
			diagTotal += len(diagStripped)
			diagVisibleTotal += diagVisible
			diagPerSection = append(diagPerSection,
				fmt.Sprintf("%s:html=%d,stripped=%d,visible=%d", s.ComponentName, len(s.HTML), len(diagStripped), diagVisible))
		}
		params.Logger.Info("SavePageSectionsAction: sections reaching save",
			zap.String("page_name", pageName),
			zap.Int("section_count", len(sections)),
			zap.Int("stripped_text_total", diagTotal),
			zap.Int("visible_text_total", diagVisibleTotal),
			zap.String("per_section", strings.Join(diagPerSection, " | ")),
		)
	}

	// --- Collapse byte-identical duplicate sections (bugs_open/156) ---
	// FIRST of this action's section-set operations, and deliberately AFTER the
	// diagnostic above so that log keeps recording the TRUE arrival count.
	//
	// Placement is not tidiness. Every guard below compares the incoming set
	// against existing rows or against a floor; none compares it against itself,
	// so a list carrying each planned section twice passes all of them AND makes
	// four of them measure a number that is not true:
	//   - the content-regression guard's newTextLen is doubled, so a page truly
	//     cut to 13% of its deployed text reads as 26% and clears the 25% floor;
	//   - the completeness floor's numerator is doubled, so a save that saw 2 of
	//     6 planned sections but emitted them twice scores 67% and clears 0.5;
	//   - the claims record and the content_data record both double-count;
	//   - and the locked-slot path in the insert loop MANUFACTURES a duplicate of
	//     human-locked copy — the first copy consumes the lock and is discarded,
	//     the second falls through and is INSERTed beside the locked row.
	// Collapsing here is what makes all of those measure the truth.
	//
	// Nothing below can re-introduce a duplicate under this key: the Layer 2
	// carry-forward appends only stored rows absent from the set by EVERY arm of
	// matchPreservedSectionIdx — identity as well as name, since bugs_open/385
	// proved slot-name-only absence is not absence (a plan-named section and a
	// positionally-named stored row are the same tool under two names, and the
	// "rescue" of one beside the other is how a locked calculator duplicated).
	// Note SectionData.Position is left stale by the collapse —
	// nothing after the enrichers reads it (the insert loop numbers from i+1), but
	// do not add a reader without renumbering here.
	// See save_sections_dedup.go for the identity rule and why a unique index on
	// (page_id, slot_name) is the wrong fix.
	sections, duplicatesCollapsed := dedupSectionsBeforePersist(ctx, params, siteID,
		datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
		pageName, pageURL, sectionsSource, metaField, metaFieldOrigin, sections)

	// The adoption flag, read once. It governs BOTH halves of the phase-2 change
	// (RFC_046 / bugs_open/357): binding an unidentified fragment to a component
	// that provably produces it, and — here — letting carried bytes keep the
	// identity they came with. Neither half survives without the other, so they
	// share one key rather than having one each. Default OFF: with the flag unset
	// this whole file behaves exactly as it did before.
	carryStoredIdentity := configBoolOrDefault(config, adoptFragmentsKey, false)

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
	//
	// "Does not reproduce" is judged by matchPreservedSectionIdx — identity
	// first, name arms after — NOT by slot name alone: on the build arm the
	// incoming names are the plan's function names, and judging a positionally
	// named stored tool "dropped" by string comparison is how a rebuild
	// appended a byte-identical, component-less copy of a locked calculator
	// beside the row the lock guard had just correctly preserved
	// (bugs_open/385 §5c; predicted by bugs_closed/189's "STILL OPEN" note).
	{
		rows, qErr := params.DB.QueryContext(ctx, `
			SELECT pc.slot_name, pc.rendered_html, pc.content_data,
			       COALESCE(pc.component_version_id::text, ''),
			       COALESCE(pc.component_id::text, ''),
			       COALESCE(cc.function, '')
			FROM page_components pc
			LEFT JOIN content_components cc ON cc.id = pc.component_id
			WHERE pc.page_id = $1 AND pc.build_status = 'deployed'
			  AND `+interactiveHTMLSQL("pc.rendered_html")+`
		`, pageID)
		if qErr != nil {
			params.Logger.Warn("SavePageSectionsAction: interactive-section preload failed (Layer 2)",
				zap.Error(qErr))
		} else {
			var preserved []preservedSection
			for rows.Next() {
				var slot, html string
				var cdJSON []byte
				var storedVersionID, storedComponentID, storedFunction string
				if scanErr := rows.Scan(&slot, &html, &cdJSON, &storedVersionID, &storedComponentID, &storedFunction); scanErr != nil {
					params.Logger.Warn("SavePageSectionsAction: interactive-section scan failed (Layer 2)",
						zap.Error(scanErr))
					continue
				}
				var cd map[string]interface{}
				if len(cdJSON) > 0 {
					_ = json.Unmarshal(cdJSON, &cd)
				}
				preserved = append(preserved, preservedSection{
					slot: slot, html: html, contentData: cd,
					componentVersionID: storedVersionID,
					componentID:        storedComponentID,
					componentFunction:  storedFunction,
				})
			}
			rows.Close()

			claimed := make(map[int]bool, len(preserved))
			for _, p := range preserved {
				matchedIdx := matchPreservedSectionIdx(sections, p, claimed)
				if matchedIdx >= 0 {
					claimed[matchedIdx] = true
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
					adoptCarriedProvenance(&sections[matchedIdx], p.componentVersionID)
					// The stored bytes keep their own identity, instead of
					// inheriting the identity of the section they displaced
					// (RFC_046 / bugs_open/357). Without this, adoption does not
					// survive: the incoming section carries the PLAN's component,
					// so the very next rebuild re-mints `hero` over an adopted row
					// and the population renews itself. Opt-in with the adoption
					// itself — neither half is useful alone, and the flag's default
					// is OFF, so this is byte-identical to today until armed.
					if id := carriedIdentity(carryStoredIdentity, p.componentID, p.componentFunction); id != "" {
						sections[matchedIdx].ComponentID = id
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
						// Same rule as the splice arm (adoptCarriedProvenance): the stored
						// bytes keep the stamp they earned, and there is no incoming
						// digest here to clear because this section is BUILT from the
						// stored row. Today it cannot reach the database — the INSERT
						// resolves a version only when the section also has a
						// component_id, and a re-appended section has none — but the two
						// carry arms must state the same thing, or the next edit reopens
						// the gap on whichever one nobody was looking at.
						ComponentVersionID: p.componentVersionID,
						ComponentID:        carriedIdentity(carryStoredIdentity, p.componentID, p.componentFunction),
						Position:           len(sections) + 1,
					})
				}
			}
		}
	}

	// --- Refuse or decode a stored LLM transport envelope (bugs_open/190) ---
	// Placed HERE deliberately: after the Layer 2 carry-forward above, which is one
	// of the two paths that recycles a stored envelope back into the section set,
	// and so after the last point at which content_data can enter it — and before
	// the history snapshot and the DELETE below, so a refused save writes nothing
	// at all. See content_data_envelope_guard.go for the discriminator, the
	// decode-vs-refuse policy and why this refusal takes no opt-in field.
	if err := sanitizeSectionsContentData(ctx, params, siteID, pageName, sections); err != nil {
		return nil, err
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

	// --- CTA label/destination agreement (bugs_open/399) ---
	// Records, never refuses and never rewrites — see cta_label_audit.go for the
	// measurements that ruled both of those out.
	//
	// PLACED HERE, at the ONE seam both writers of page_components pass through,
	// and that placement is the whole reason it is in this file rather than at
	// the render seam where the fresh label is first visible.
	// RenderComponentAction looks like the natural home and covers only HALF the
	// population: RerenderPageSectionsAction does not go through it at all
	// (rerender_page_sections_action.go:662 calls RenderTemplate directly), so a
	// gate there would be blind to the repair loop — which is the loop actually
	// minting the churn (182 misdirected_cta item_keys have been filed more than
	// once, [MEASURED 2026-08-26]). Both paths end here:
	//     page-build-handler → call_content_writer → save_sections
	//     page-rerender      → rerender_sections   → save_sections
	// verified against the live agent_definitions rows on 2026-08-26.
	//
	// ⚠ TWO OF THREE, NOT ALL THREE. ApplySectionEditAction writes
	// page_components.content_data directly (section_editor_actions.go) and does
	// NOT come through here, so it is outside this pass. Live, not dormant: 144
	// section_edit items, newest 2026-08-26, of which 3 name a CTA field
	// [MEASURED 2026-08-26]. Stated rather than widened — see cta_label_audit.go.
	//
	// AFTER the link repair on purpose, for that pass's own stated reason: it
	// rewrites content_data urls, so judging before it would judge values that
	// are about to change and report a contradiction this save then fixed.
	auditCTALabelAgreement(ctx, params, siteID,
		datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
		pageName, pageURL, sections, params.Logger)

	// --- Claims floor (bugs_open/149 C1) ---
	// Placed immediately after the link repair so it scans the bytes that will
	// actually be written, and BEFORE the regression guards so a page refused for
	// asserting a falsehood is refused for that reason rather than for whatever
	// the guards notice next.
	//
	// Six live agents persist sections through here; only two of them run
	// validate_page_content, so for the other four this is the only claims check
	// that exists. A banned claim refuses the save; an unregistered number is
	// recorded and allowed. See save_sections_claims_guard.go for the severity
	// rule, the measured blast radius and the scope boundary.
	if err := claimsGuardBeforePersist(ctx, params, siteID, pageID,
		datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
		pageName, pageURL, sections, params.Logger); err != nil {
		return nil, err
	}

	// --- Content regression guard ---
	// Refuse to overwrite content-rich pages with empty template shells.
	// This prevents LLM failures (credit exhaustion, timeouts, empty responses)
	// from wiping good content that was previously generated and deployed.
	// Extracted to save_sections_shrink_guard.go (bugs_open/293) so it has a home,
	// a test, and the config escape hatch it never had — and so its axis is the same
	// VISIBLE-text measure its two siblings below use. As an inline block measuring
	// tag-stripped length in SQL it would allow a whole-page prose wipe on 337 of
	// 366 pages: a stylesheet anywhere on the page props up the total that every
	// slot's loss is diluted into. Same guarantee, same thresholds, honest measure.
	if regressionErr := enforcePageTotalTextFloor(ctx, params, siteID, pageID, pageName, sections); regressionErr != nil {
		return nil, regressionErr
	}

	// --- Per-slot shrink guard (bugs_open/178) ---
	// The page-total guard above only refuses a near-wipe (<25% of the page's
	// text) — a regeneration that guts ONE prose slot passes it because the
	// loss is diluted by intact siblings. This one compares slot against
	// same-named slot and refuses a prose slot shrinking past its floor.
	// See save_sections_shrink_guard.go for scope and the config escape hatch.
	if shrinkErr := enforceSectionShrinkFloor(ctx, params, siteID, pageID, pageName, sections); shrinkErr != nil {
		return nil, shrinkErr
	}

	// --- Per-slot COMPONENT floor (bugs_open/253, framework_rewrite_...) ---
	// The shrink guard above measures TEXT, and is blind to a rewrite that keeps
	// the words and strips the layout. Measured on the LMC homepage: the save it
	// let through kept 84% of the text and 2% of the class attributes, turning a
	// styled calculator directory into a flat run of headings. Text volume and
	// layout structure are independent quantities; this guards the second.
	// See save_sections_component_floor.go for the calibration and escape hatch.
	if flatErr := enforceSectionComponentFloor(ctx, params, siteID, pageID, pageName, sections); flatErr != nil {
		return nil, flatErr
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
			SELECT COALESCE(bool_or(`+interactiveHTMLSQL("rendered_html")+`), false)
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

	// --- Locked-section preservation (bugs_open/058) ---
	// Human-locked rows must survive the rebuild with copy AND row identity
	// intact (admin lock/unlock addresses rows by id; "unchanged" means
	// updated_at too). So: preload the actively-locked rows, delete only the
	// agent-writable ones, and skip the incoming section that would have
	// replaced a locked slot — the locked copy stands, only its position moves
	// to follow the new composition. Expiry-aware predicate is the single
	// shared one (lock_helpers.go).
	//
	// Loaded HERE, ahead of the completeness floor below, because the floor's
	// numerator is what this save will actually insert and a lock discards the
	// incoming section that matches it.
	lockedRows := loadActiveLockedRows(ctx, params.DB, pageID, params.Logger)

	// --- Decision citation gate, the rebuild door (RFC_015 §5b) ---
	// Owner ruling 2026-08-10, option 1 ("gentler version"): a slot covered by a
	// decision this save did not CITE keeps its stored row instead of being
	// overwritten, and the blocked overwrite is filed. The rebuild is NOT failed.
	// Loaded here for the same reason as the locked rows above — ahead of the
	// completeness floor, whose numerator is what this save will actually insert,
	// and a protected slot discards the incoming section that matches it.
	// See save_sections_decision_gate.go for why there is no bypass field.
	decisionCitation := readDecisionCitation(params.CollectedData, config)
	decisionProtected := loadDecisionProtectedRows(ctx, params.DB, siteID, pageID,
		pageName, decisionCitation, lockedRows, params.Logger)

	// --- Completeness floor (bugs_open/165 site A; rule from bugs_closed/135) ---
	// LAST of this action's refusal guards, because it is the only one that needs
	// the FINAL section set: the interactive-tool preservation above re-appends
	// sections a rebuild dropped, and a floor measured before that would refuse a
	// save the preservation had already repaired. Ahead of the history snapshot
	// below so that a refused save writes nothing whatever — not even a history
	// row recording content it never went on to replace.
	//
	// The other guards ask whether this save is ENTITLED to delete these rows, or
	// whether the replacement looks impoverished in characters. This one asks the
	// question none of them do: did this run see enough of the page to be
	// replacing it at all? See save_sections_prune_floor.go.
	floorDetail, floorErr := enforcePageSectionFloor(ctx, params, siteID, pageID, pageName, sections, lockedRows)
	if floorErr != nil {
		return nil, floorErr
	}

	// --- content_data regression record (bugs_open/194) ---
	// NOT a guard: it records, it never refuses. Placed after the floor so it sees
	// the FINAL section set (the interactive carry-forward above can restore a
	// section's content_data, and a report taken before it would be reporting a
	// loss that had already been repaired), and before the snapshot so the rows it
	// counts are still the ones this save is about to replace.
	//
	// The condition it exists to catch is the whole of bugs_open/194: a page that
	// held structured content is saved with none, succeeds, serves perfectly, and
	// has silently lost the only source a re-render can rebuild it from.
	// The DB read is skipped where its answer provably cannot change the outcome
	// (code-review F12). shouldReportContentDataLoss discards it in exactly two
	// cases, both decidable in memory: a declared-absent caller never had
	// structured content to lose, and incoming sections that already carry
	// content_data mean this save is not the loss it looks for. Passing 0 in
	// those cases is equivalent, not merely close — the predicate returns false
	// on both before it ever reads the count.
	//
	// Worth doing here specifically because F9 widened this query (the
	// build_status filter is gone, so it now considers every agent-writable row
	// on the page) and this is the fleet's highest-traffic save path.
	incomingWithContentData := countSectionsWithContentData(sections)
	contentDataRowsBefore := 0
	if metaFieldOrigin != metadataOriginDeclaredAbsent && incomingWithContentData == 0 {
		contentDataRowsBefore = countExistingRowsWithContentData(ctx, params, pageID)
	}
	if shouldReportContentDataLoss(metaFieldOrigin, contentDataRowsBefore, sections) {
		writeContentDataRegressionLog(ctx, params, siteID,
			datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain"),
			pageName, metaField, metaFieldOrigin, contentDataRowsBefore, sections, params.Logger)
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

	// Classify what the DELETE below is about to destroy, BEFORE destroying it
	// (bugs_open/229): rows whose bytes no longer match their render stamp hold
	// content only the artefact has. Advisory — the save proceeds regardless;
	// recovery is the 357 trigger's, which archives every deleted artefact
	// atomically with the DELETE. The predicate inside classify matches the
	// DELETE's exactly, so locked rows that survive are not counted.
	divergent, classifyErr := classifyPageComponentArtefacts(ctx, params.DB, pageID)
	if classifyErr != nil {
		params.Logger.Warn("SavePageSectionsAction: divergence classification failed — save proceeds, the 357 trigger still archives (bugs_open/229)",
			zap.Error(classifyErr),
		)
	}

	// Clear existing components for this page — except actively-locked rows and
	// decision-protected ones. lockedRows and decisionProtected were both loaded
	// above (see the preservation notes there), and the completeness floor has
	// already established that this run saw enough of the page to be replacing it.
	//
	// The decision exclusion is a SEPARATE clause, not a change to
	// pageComponentAgentWritableSQL: that fragment is the fleet's single source of
	// truth for "may automation write this row?" and is shared with writers and a
	// discovery check that must all agree about LOCKS. Decision protection is a
	// per-save question (it depends on this envelope's citation), so it cannot
	// live in a static predicate — and folding it in would silently change what
	// every other caller of that fragment means.
	delRes, err := params.DB.ExecContext(ctx, `
		DELETE FROM page_components WHERE page_id = $1 AND `+pageComponentAgentWritableSQL("")+`
		  AND NOT (id = ANY($2::uuid[]))`,
		pageID, decisionProtectedIDArrayLiteral(decisionProtected))
	if err != nil {
		params.Logger.Warn("SavePageSectionsAction: Failed to clear existing components",
			zap.Error(err),
		)
		// Continue anyway
	}

	// Emit only after the DELETE actually removed rows (mirror of the chrome
	// emit-after-RowsAffected rule): a failed or empty DELETE destroyed
	// nothing. When the pre-classify failed, fall back to the verdicts the 357
	// trigger just wrote to the ledger — same judgement, DB-side.
	if err == nil {
		if n, raErr := delRes.RowsAffected(); raErr == nil && n > 0 {
			if classifyErr != nil {
				if fromLedger, lbErr := readBackPageDivergenceFromLedger(ctx, params.DB, pageID); lbErr == nil {
					divergent = fromLedger
				} else {
					params.Logger.Warn("SavePageSectionsAction: ledger read-back also failed — divergence signal lost for this save, archive rows remain (bugs_open/229)",
						zap.Error(lbErr),
					)
					divergent = nil
				}
			}
			emitPageDivergenceItems(ctx, params.DB, pageID, pageName, divergent, "save_page_sections", params.Logger)
		}
	}

	// Insert each section
	savedCount := 0
	// Counted as rows are actually written, not from the incoming list
	// (code-review F11): the loop below skips locked slots and unresolvable
	// stubs, so a figure taken beforehand reports content_data this save
	// discarded. The field exists to say what the page now HOLDS.
	savedWithContentData := 0
	var skippedStubs []string
	var lockedSlotsPreserved []string
	var decisionSlotsPreserved []string
	for i, section := range sections {
		// ── Locked-slot guard (bugs_open/058) ────────────────────────────────
		// The new composition produced fresh copy for a slot a human has
		// locked: the locked row (kept out of the DELETE above) stands, the
		// fresh copy is discarded, and the block is surfaced as a work item —
		// a silent skip would trade one silent failure for another. Only the
		// row's position moves, so ordering follows the new composition.
		if lr := matchLockedRow(lockedRows, section.ComponentName, section.ComponentID); lr != nil {
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

		// ── Decision citation gate, same shape as the lock guard above ────────
		// The slot is covered by a decision this save did not name: the stored
		// row stands, the fresh copy is discarded, only the position moves, and
		// the blocked overwrite is filed. The rebuild continues — that is the
		// "gentler version" of the owner ruling of 2026-08-10 (RFC_015 §5b).
		if dr := matchDecisionProtectedRow(decisionProtected, section.ComponentName, section.ComponentID); dr != nil {
			dr.consumed = true
			if _, posErr := params.DB.ExecContext(ctx, `
				UPDATE page_components SET position = $2 WHERE id = $1
			`, dr.id, i+1); posErr != nil {
				params.Logger.Warn("SavePageSectionsAction: failed to reposition decision-protected section",
					zap.String("slot_name", dr.slot), zap.Error(posErr))
			}
			params.Logger.Warn("SavePageSectionsAction: preserving decision-protected section over rebuilt copy (RFC_015)",
				zap.String("page_name", pageName),
				zap.String("slot_name", dr.slot),
				zap.Strings("decisions", dr.decisions),
			)
			decisionSlotsPreserved = append(decisionSlotsPreserved, dr.slot)
			drID := dr.id
			emitDecisionBlockedChangeItem(ctx, params.DB, siteID, &pageID, &drID,
				pageName, dr.slot, dr.decisions,
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
		if sectionIsUnresolvableStub(section) {
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

		// rendered_html_digest = md5($3) in the SAME statement as the bytes
		// (bugs_open/229; the IMP-052 same-statement principle): the stamp
		// means "reproducible from content_data", and only the render/save
		// path may write it.
		// Provenance stamp (RFC_046, ruled 2026-08-22): WHICH version of the
		// component produced these bytes. nil = unknown, and unknown is written as
		// NULL — resolveComponentVersionID refuses far more often than it answers,
		// deliberately. Nothing reads this column yet (0 readers, measured
		// 2026-08-22), so this cannot change what any page serves.
		var componentVersionIDPtr interface{}
		if componentIDPtr != nil {
			if vid, ok := resolveComponentVersionID(ctx, params.DB, *componentIDPtr,
				componentVersionSource{
					carriedVersionID: section.ComponentVersionID,
					renderedSHA:      section.RenderedTemplateSHA,
				}, params.Logger); ok {
				componentVersionIDPtr = vid
			}
		}

		_, err := params.DB.ExecContext(ctx, `
			INSERT INTO page_components (page_id, position, rendered_html, rendered_html_digest, slot_name, component_id, content_data, content_brief, build_status, component_version_id)
			VALUES ($1, $2, $3, md5($3), $4, $5, $6::jsonb, $7::jsonb, 'deployed', $8)
		`, pageID, i+1, section.HTML, section.ComponentName, componentIDPtr, contentDataJSON, contentBriefJSON, componentVersionIDPtr)

		if err != nil {
			params.Logger.Warn("SavePageSectionsAction: Failed to insert section",
				zap.Int("position", i+1),
				zap.String("component", section.ComponentName),
				zap.Error(err),
			)
			continue
		}
		savedCount++
		if len(section.ContentData) > 0 {
			savedWithContentData++
		}
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

	// Decision-protected rows the new composition no longer includes. Same
	// treatment as the locked leftovers above, and it is the case that matters
	// most in practice: a rebuild DROPPING a protected slot is how the 2026-08-10
	// regression would have been reported had the gate existed, and equally how a
	// deliberately REMOVED section keeps its removal — preserving the stored row
	// preserves whatever state it holds, including build_status='removed'.
	for _, dr := range decisionProtected {
		if dr.consumed {
			continue
		}
		if _, posErr := params.DB.ExecContext(ctx, `
			UPDATE page_components SET position = $2 WHERE id = $1
		`, dr.id, nextPos); posErr != nil {
			params.Logger.Warn("SavePageSectionsAction: failed to reposition retained decision-protected section",
				zap.String("slot_name", dr.slot), zap.Error(posErr))
		}
		nextPos++
		params.Logger.Warn("SavePageSectionsAction: retaining decision-protected section the new composition dropped (RFC_015)",
			zap.String("page_name", pageName),
			zap.String("slot_name", dr.slot),
			zap.Strings("decisions", dr.decisions),
		)
		decisionSlotsPreserved = append(decisionSlotsPreserved, dr.slot)
		drID := dr.id
		emitDecisionBlockedChangeItem(ctx, params.DB, siteID, &pageID, &drID,
			pageName, dr.slot, dr.decisions,
			"remove", "save_page_sections", params.Logger)
	}

	params.Logger.Info("SavePageSectionsAction: Complete",
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()),
		zap.Int("sections_found", len(sections)),
		zap.Int("sections_saved", savedCount),
		zap.Int("skipped_stub_sections", len(skippedStubs)),
		zap.Int("locked_sections_preserved", len(lockedSlotsPreserved)),
		zap.Int("decision_protected_sections_preserved", len(decisionSlotsPreserved)),
	)

	result := map[string]interface{}{
		"success":                   true,
		"page_id":                   pageID.String(),
		"page_name":                 pageName,
		"sections_found":            len(sections),
		"sections_saved":            savedCount,
		"skipped_stub_sections":     skippedStubs,
		"locked_sections_preserved": lockedSlotsPreserved,
		// RFC_015 §5b: slots this rebuild would have overwritten but a decision
		// protects and the envelope did not cite. Surfaced on every save, not
		// only a gated one, so an acceptance run can assert the branch rather
		// than infer it from an absence — the same reasoning as sections_source
		// below, and the reason the 08-09 citation-gate proof was ambiguous.
		"decision_protected_sections_preserved": decisionSlotsPreserved,
		// Which representation was persisted, and how that was decided
		// (bugs_open/194). Reported on every save, not only a losing one, so an
		// acceptance run can assert the BRANCH rather than infer it from a
		// non-NULL column — content_data can also arrive via the interactive
		// carry-forward, which would make a NULL check a false pass.
		"sections_source":            sectionsSource,
		"metadata_field_origin":      metaFieldOrigin,
		"sections_with_content_data": savedWithContentData,
		// The incoming figure kept beside it: when these two differ, the save
		// discarded structured content (a locked slot or an unresolvable stub),
		// which is exactly what the single pre-loop count used to hide (F11).
		"incoming_sections_with_content_data": incomingWithContentData,
		// Reported on every save, not only a firing one (bugs_open/156): a zero
		// here is the assertion that this save carried no byte-identical
		// duplicates, which is a different statement from the key being absent.
		"duplicate_sections_collapsed": duplicatesCollapsed,
	}
	// The floor's numbers are reported on a PASSING save too, not only when it
	// refuses. "sections_saved: 2" without the denominator is the alarm presented
	// as output — 135 candidate (3), and the reason orchestration_states could not
	// answer "was that rebuild thin?" after the fact.
	for k, v := range floorDetail {
		result[k] = v
	}
	return result, nil
}

// lockedPageRow is an actively-locked page_components row preserved through a
// rebuild (bugs_open/058).
type lockedPageRow struct {
	id       uuid.UUID
	slot     string
	position int
	lockedBy string
	lockType string
	// componentID is the content_components row this locked section renders.
	// Carried so matchLockedRow can pair on IDENTITY as well as slot name —
	// the sibling guard (matchDecisionProtectedRow) always could, and the
	// asymmetry is a real defect on any page whose slot names are positional
	// rather than component functions. See matchLockedRow.
	componentID string
	consumed    bool // matched (and thereby blocked) an incoming section this save
}

// loadActiveLockedRows returns the page's rows automation may not overwrite,
// in position order. Best-effort: on query failure it returns nil and the save
// proceeds exactly as before this guard existed (the DELETE predicate still
// protects the rows themselves; only slot-matching is lost, which would leave
// a duplicate slot rather than destroy locked copy).
//
// activeLockedRowsSQL is hoisted so a lockstep test can pin that this loader
// and the list-side loader (datahelpers.LockedPageSlotsSQL, bugs_open/285)
// negate the SAME predicate string — the two answer "may automation rewrite
// this row?" for the ROW and for the LIST, and must never drift on the lock
// test itself (they legitimately differ on membership: this one carries every
// locked row so the DELETE cannot destroy it; the list one omits
// build_status='removed', which is not on the page).
var activeLockedRowsSQL = `
		SELECT id, COALESCE(slot_name, ''), position, COALESCE(locked_by, ''), COALESCE(lock_type, ''),
		       COALESCE(component_id::text, '')
		FROM page_components
		WHERE page_id = $1 AND NOT ` + pageComponentAgentWritableSQL("") + `
		ORDER BY position ASC
	`

func loadActiveLockedRows(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) []*lockedPageRow {
	rows, err := db.QueryContext(ctx, activeLockedRowsSQL, pageID)
	if err != nil {
		logger.Warn("SavePageSectionsAction: locked-row preload failed (bugs_open/058)", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var locked []*lockedPageRow
	for rows.Next() {
		lr := &lockedPageRow{}
		if scanErr := rows.Scan(&lr.id, &lr.slot, &lr.position, &lr.lockedBy, &lr.lockType, &lr.componentID); scanErr != nil {
			logger.Warn("SavePageSectionsAction: locked-row scan failed", zap.Error(scanErr))
			continue
		}
		locked = append(locked, lr)
	}
	return locked
}

// matchLockedRow finds the first unconsumed locked row matching the incoming
// section — by component IDENTITY first, then slot name exact, then slot name
// kebab-normalised (the 041 naming landmine: the library stores kebab-case but
// older rows/plans may carry snake_case or CamelCase variants of the same
// slot). Each locked row matches at most one incoming section, so a page with
// duplicate slot names cannot have one lock swallow several sections.
//
// The identity arm mirrors matchDecisionProtectedRow, which has always had it
// (save_sections_decision_gate.go: "Identity beats naming"), and its absence
// here was a real defect rather than a deliberate difference. Slot-name-only
// matching silently assumes slot names ARE component functions. On a page
// decomposed into POSITIONAL slots — prose-0, tool-2, the shape adoption and
// the decompose lane produce — a replan that names the very component a locked
// row renders does not pair with it. The locked row then falls through to the
// unconsumed tail pass below and is repositioned after the whole new set,
// while the loop separately inserts a freshly composed section for the SAME
// component: the page ends up with the automation's copy in place and the
// human-locked original exiled to the foot. That is two defects at once —
// a duplicate, and a lock that "held" while losing its position entirely.
//
// Guarded on non-empty exactly as the sibling is: sections often arrive before
// enrichSectionsWithComponentIDs has resolved an id, and an empty-string match
// would pair every unresolved section with the first idless locked row.
//
// The arms themselves live in datahelpers/slot_pairing.go — ONE relation
// shared with MergeLockedPageSlots and matchPreservedSectionIdx (council
// ece638fb's reuse gate: hand-mirrored copies of these arms drifted, and the
// drift is bugs_open/385). The function/name arm is a structural NO-OP here,
// deliberately: loadActiveLockedRows does not join content_components, so the
// SlotIdentity views carry empty function/name and this extraction changes
// nothing about which rows pair — the suite in
// save_sections_locked_identity_test.go and save_sections_positional_tool_slot_test.go
// is the equivalence proof. Widening this matcher to the function arm would
// mean joining cc in the loader, which is a behaviour change to make on its
// own evidence, not as a refactor side effect.
func matchLockedRow(lockedRows []*lockedPageRow, sectionName, sectionComponentID string) *lockedPageRow {
	stored := make([]datahelpers.SlotIdentity, len(lockedRows))
	for i, lr := range lockedRows {
		stored[i] = datahelpers.SlotIdentity{Slot: lr.slot, ComponentID: lr.componentID}
	}
	idx := datahelpers.PairIncomingToStored(sectionName, sectionComponentID, stored,
		func(i int) bool { return lockedRows[i].consumed })
	if idx < 0 {
		return nil
	}
	return lockedRows[idx]
}

// preservedSection is one stored interactive row the Layer 2 carry-forward
// preloaded (build_status='deployed' AND interactive) so a rebuild cannot
// destroy a tool the composition failed to reproduce.
type preservedSection struct {
	slot        string
	html        string
	contentData map[string]interface{}
	// The stamp that describes THESE bytes (RFC_046). A carry does not
	// change the bytes, so it must not change their provenance — and it
	// must not let the incoming section's provenance describe them.
	componentVersionID string
	// The stored row's own component. Carried with the bytes only when
	// the adoption flag is on: without it, a rebuild re-imposes the
	// PLAN's identity on bytes the plan did not produce, which is what
	// re-mints bugs_open/357's population on every rebuild.
	componentID string
	// The stored component's FUNCTION, so the carry can be narrowed to
	// exactly what adoption created rather than re-typing every carried
	// section on the page — and so the matcher below can pair a stored
	// positional slot with the plan entry that names the same component.
	componentFunction string
}

// matchPreservedSectionIdx pairs one preloaded interactive row against the
// incoming set: "is this stored tool already represented in what this save is
// about to write?" It is the THIRD of the save path's three judgements of
// that question, and until bugs_open/385 it was the only one still answering
// by exact slot-name string — the slot-names-are-not-component-functions
// defect matchLockedRow's identity arm and MergeLockedPageSlots' arm 3 were
// built to close (182/189/204, LOCK-008). On the build arm the incoming names
// are the PLAN's function names, so a positionally-named stored tool
// (`tool-2`) read as "dropped entirely" while the very same component sat in
// the set as `tool-loan-vs-savings`, and the re-append arm below duplicated a
// locked calculator (bugs_open/385 §5c, 2026-08-23).
//
// The arms live in datahelpers/slot_pairing.go — ONE relation shared with
// matchLockedRow and MergeLockedPageSlots (council ece638fb's reuse gate:
// three hand-mirrored copies of these arms is how the third one drifted and
// minted 385's orphan). Order: component IDENTITY first
// (enrichSectionsWithComponentIDs has already run by the time Layer 2
// executes, so a plan-named section carries its resolved id), then slot name
// exact (the pre-385 behaviour, kept — the rerender arm's name space), then
// kebab-normalised slot name, then the stored component's function/name
// against the incoming name — the arm that still decides the 385 shape when
// the incoming side failed to enrich. Every arm guarded on non-empty, so an
// empty id or name never pairs.
//
// `claimed` gives each stored row at most one incoming section — the
// consumption rule both siblings already have. Two stored instances of one
// component against a composition naming it once must pair one-to-one, so the
// second instance is re-appended (preserved) rather than silently judged
// "already present".
func matchPreservedSectionIdx(sections []SectionData, p preservedSection, claimed map[int]bool) int {
	incoming := make([]datahelpers.IncomingSection, len(sections))
	for i := range sections {
		incoming[i] = datahelpers.IncomingSection{
			Name:        sections[i].ComponentName,
			ComponentID: sections[i].ComponentID,
		}
	}
	return datahelpers.PairStoredToIncoming(
		datahelpers.SlotIdentity{Slot: p.slot, ComponentID: p.componentID,
			ComponentFunction: p.componentFunction},
		incoming,
		func(i int) bool { return claimed[i] },
	)
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

// sectionIsUnresolvableStub is the bugs_open/039 skip predicate, extracted so
// that the insert loop and the completeness floor (save_sections_prune_floor.go)
// decide by the SAME rule. They must agree: the floor counts what this save will
// write, and a section the loop is about to discard is not written. Two copies of
// this condition sixty lines apart is the drift class the council reviews for —
// a floor that counted discarded stubs as confirmed content would be permissive
// in the one direction it must not be.
//
// Mirrors the loop's original test exactly: no resolvable component_id AND the
// hollow generic stub.
func sectionIsUnresolvableStub(s SectionData) bool {
	if s.ComponentID != "" {
		if _, err := uuid.Parse(s.ComponentID); err == nil {
			return false
		}
	}
	return isEmptyGenericStub(s.HTML)
}

// SectionData holds extracted section data
type SectionData struct {
	ComponentName string
	ComponentID   string
	HTML          string
	Position      int
	ContentData   map[string]interface{} // structured content for re-rendering (source of truth)

	// ── Provenance (RFC_046, ruled 2026-08-22) ──────────────────────────────
	// RenderedTemplateSHA is set when these bytes came out of RenderTemplate:
	// the digest of the template text that produced them. ComponentVersionID is
	// set when the bytes were CARRIED from a stored row that already had a stamp —
	// a carry does not change the bytes, so it must not change their provenance.
	//
	// BOTH EMPTY MEANS UNKNOWN, and unknown must be persisted as NULL. Nothing in
	// this file may infer a version, for the same reason nothing should have been
	// inferring a component: the row that claims to know something it guessed is
	// the defect (bugs_open/357).
	RenderedTemplateSHA string
	ComponentVersionID  string

	// FallbackAdopted marks a section that exists only because the page's HTML
	// carried no <section> at all and the whole fragment was stored as one
	// (saveSectionsExtractFromHTML's documented fallback). It is set ONLY when the
	// fragment also declares no data-component — i.e. when nothing about the bytes
	// says what they are.
	//
	// This is the one place where identity is INVENTED rather than carried, and
	// marking it is what lets the enrichment stop inventing (RFC_046,
	// bugs_open/357). It is not persisted: it describes how the section got here,
	// which the row itself has no business claiming.
	FallbackAdopted bool
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

		// ── Slot identity before component identity (bugs_open/189) ──────────
		// The slot name says WHICH section this is on the page; the component
		// function says WHAT renders it. They are different facts, and the
		// function-first rule below collapsed them into one — so the first time
		// a positionally-named slot ("prose-0", "tool-2") became resolvable, the
		// save silently RENAMED it to the component's own identity, and
		// matchLockedRow (which matches the incoming name against the locked
		// rows' stored slot_name) then missed the very row it exists to protect.
		// A human-locked section was duplicated on the page rather than
		// preserved.
		//
		// So a producer that HOLDS the page_components row's own slot identity
		// now says so in stored_slot_name, and it is taken VERBATIM: this is a
		// row identity being matched back to the row that issued it, and
		// normalising a legacy spelling would un-match it.
		// NormalizeComponentFunction stays on the DERIVED-name path below, where
		// bugs_closed/041 put it — a name we invent from a component must obey
		// the kebab-case contract; a name the page already carries must not be
		// rewritten by us at all.
		//
		// Absence is today's behaviour byte for byte: the tool-recreation path
		// regenerates single-tool HTML with no structured slot identity and must
		// keep working unchanged, as must any orchestration expanded before the
		// producers gained the field.
		componentName := ""
		if slot, ok := m["stored_slot_name"].(string); ok && slot != "" {
			componentName = slot
		} else {
			// component_function is the slot_name (e.g. "hero", "call-to-action")
			componentName = "section"
			if fn, ok := m["component_function"].(string); ok && fn != "" {
				componentName = fn
			} else if name, ok := m["component_name"].(string); ok && name != "" {
				componentName = name
			}

			// Enforce naming contract: slot_name must be kebab-case
			componentName = NormalizeComponentFunction(componentName)
		}

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

		// Provenance travels with the bytes (RFC_046). A fresh render reports the
		// template digest; a carried section reports the version its stored row
		// already carried. Absent from the map = unknown = persisted as NULL.
		renderedTemplateSHA, _ := m["rendered_template_sha"].(string)
		componentVersionID, _ := m["component_version_id"].(string)

		sections = append(sections, SectionData{
			ComponentName:       componentName,
			ComponentID:         componentID,
			HTML:                strings.TrimSpace(html),
			Position:            i + 1,
			ContentData:         contentData,
			RenderedTemplateSHA: renderedTemplateSHA,
			ComponentVersionID:  componentVersionID,
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
				// Nothing about these bytes says what they are: no <section>, and
				// (when the name is still the sentinel) no data-component either.
				// Everything downstream that names this section is guessing, and
				// this flag is what lets the component binding decline to
				// (bugs_open/357).
				FallbackAdopted: componentName == "section",
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
func enrichSectionsWithComponentIDs(ctx context.Context, db *sql.DB, sections []SectionData, logger *zap.Logger, adoptFragments bool) {
	logger.Info("enrichSectionsWithComponentIDs: invoked",
		zap.Int("section_count", len(sections)),
		zap.Bool("db_nil", db == nil),
		zap.Bool("adopt_fragments", adoptFragments))

	dataComponentRe := regexp.MustCompile(`data-component="([^"]+)"`)

	for i := range sections {
		if sections[i].ComponentID != "" {
			continue // already has an ID
		}

		// RFC_046 / bugs_open/357 — the one place identity is INVENTED.
		//
		// This section exists only because the page had no <section> at all, and
		// its bytes declare no component. By this point the planned-name pass has
		// given it a slot name from POSITION in the plan — correctly, that is what
		// slot names are — and the resolution below would then read that positional
		// name as a statement about what the bytes ARE, binding a whole interactive
		// tool to the shared `hero` because hero was planned first.
		//
		// So: bind it to a component that provably reproduces these bytes, or bind
		// it to nothing. Never to the name.
		if adoptFragments && sections[i].FallbackAdopted {
			if dataComponentRe.FindStringSubmatch(sections[i].HTML) == nil {
				if adoptFragmentSection(ctx, db, &sections[i], logger) {
					continue
				}
				logger.Info("enrichSectionsWithComponentIDs: fragment not adoptable — leaving it unidentified "+
					"rather than binding it to its positional name (bugs_open/357)",
					zap.String("slot_name", sections[i].ComponentName),
					zap.Int("position", i+1))
				continue
			}
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
// ── The interactivity predicate: ONE definition, two languages ──────────
//
// This predicate decides whether a section holds a hand-built interactive tool
// that a rebuild must not destroy. It was previously spelled out five times —
// four SQL OR-chains plus one Go function — which is the drift shape this
// platform has been bitten by before (idx_swi_dedup vs workItemTerminalStatuses).
// Both languages now derive from the slices below, and a test asserts they cover
// the same markers.
//
// WIDENED 2026-07-30. The original three markers (<canvas, game-container,
// tool-page) recognise games and marked tool pages. They do NOT recognise a
// calculator: loancalculator.co.uk's twelve tools are <input> fields plus an
// inline <script> using getElementById, containing none of the three. A rebuild
// of those pages would have passed both guards and silently dropped every
// calculator — the same silent-loss class the guards exist to prevent.
//
// Deliberately NOT widened to "<input> anywhere". Over-matching has a real cost
// in the other direction: a section wrongly judged interactive is carried
// forward verbatim and can never be improved by the writer or the improvement
// loops, which is the opposite of what an evolving site needs. A plain contact
// form or a newsletter box must stay editable. So a control alone is not enough
// — it must be a control the page's own script drives.
var (
	// Structural markers strong enough on their own: bespoke rendering surfaces
	// and explicit tool markup.
	interactiveStructuralMarkers = []string{"<canvas", "game-container", "tool-page", "data-tool"}

	// Interactive controls — only meaningful together with a script (below).
	interactiveControlMarkers = []string{"<input", "<select", "<textarea", "oninput=", "onchange=", "onclick="}
)

// sectionHTMLIsInteractive reports whether the section holds an interactive tool.
// Structural marker alone, OR a script driving a control.
func sectionHTMLIsInteractive(html string) bool {
	h := strings.ToLower(html)
	for _, m := range interactiveStructuralMarkers {
		if strings.Contains(h, m) {
			return true
		}
	}
	if !strings.Contains(h, "<script") {
		return false
	}
	for _, m := range interactiveControlMarkers {
		if strings.Contains(h, m) {
			return true
		}
	}
	return false
}

// interactiveHTMLSQL renders the same predicate as a SQL boolean over `col`,
// so the queries and sectionHTMLIsInteractive cannot disagree. ILIKE gives the
// case-insensitivity that strings.ToLower gives the Go side.
func interactiveHTMLSQL(col string) string {
	lit := func(m string) string { return "'%" + strings.ReplaceAll(m, "'", "''") + "%'" }
	var structural []string
	for _, m := range interactiveStructuralMarkers {
		structural = append(structural, col+" ILIKE "+lit(m))
	}
	var controls []string
	for _, m := range interactiveControlMarkers {
		controls = append(controls, col+" ILIKE "+lit(m))
	}
	return "((" + strings.Join(structural, " OR ") + ") OR (" +
		col + " ILIKE '%<script%' AND (" + strings.Join(controls, " OR ") + ")))"
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
