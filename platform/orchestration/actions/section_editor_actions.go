// FILE: platform/orchestration/actions/section_editor_actions.go
//
// Actions for the section-editor agent. Enables granular edits to individual
// page sections without re-running the full content generation pipeline.
//
// Two actions:
//   - load_edit_context:  Gathers the target section + component + page metadata
//   - apply_section_edit: Performs the edit, updates page_components, reassembles page HTML
//
// Edit types supported:
//   - content_edit:    Update content_data fields, re-render from template + DB context
//   - component_swap:  Change component template, re-render with existing content_data
//   - rendered_html_transform: Apply a NAMED deterministic transform to the
//     existing rendered_html, content_data untouched. OPT-IN (config key
//     allow_rendered_html_transform, default OFF — enabled per step by
//     migration 513, per the 2026-08-02 §2 ruling on new authority on shared
//     seams). Added 2026-08-20 for bugs_open/277 §5: a component whose
//     content_data cannot reproduce its own rendered_html (the worked case:
//     Ported Page, 100 of 115 instances hold NONE of their template's fields)
//     is unreachable by both edit types above BY CONSTRUCTION, so its
//     rendered_html-surface findings can only be repaired by editing the
//     finished HTML. One transform exists: code_span_to_code_tag
//     (datahelpers.ConvertLiteralCodeSpansInHTML — byte-splicing, skip-set
//     shared with the detector; its safety argument lives in that file's
//     header). The transform never renders and never touches content_data;
//     its persist is the HTML-only branch of updatePageComponentAfterEdit.
//
// Design principle: content_data is the source of truth. Every edit updates
// content_data first, then re-renders from template. This means edits survive
// future re-renders (nav updates, theme changes, etc.).
// EXCEPTION, stated rather than implied: rendered_html_transform edits the
// render OUTPUT of components whose content_data cannot reproduce it — there
// is no source of truth to edit, and nothing regenerates these components
// (that same property is why their pages are refused by every rerender
// route), so the edit is as durable as the component itself.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ===========================================================================
// INPUT SPECS
// ===========================================================================

var LoadEditContextInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"page_component_id", "page_name", "slot_name", "domain"},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"site_id_field": "site_id",
	},
}

var ApplySectionEditInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"edit_type"},
	Optional: []string{
		"field_updates",            // content_edit: JSON object of fields to merge
		"replacement_content_data", // content_edit or component_swap: full replacement content_data
		"new_component_function",   // component_swap
		"page_component_id",        // target (can also come from edit_context)
		"acknowledges_decision",    // RFC_015 citation gate: decision key(s) this edit knowingly works within
		"supersedes_decision",      // RFC_015 citation gate: decision key(s) this edit knowingly replaces
		"transform_name",           // rendered_html_transform: which named transform to apply (code_span_to_code_tag)
	},
	// strip_literal_markdown is a SETTING, not a data reference (bugs_open/184):
	// when true, the merged content map is passed through
	// datahelpers.StripLiteralMarkdownFromContentData before each branch's
	// render. Default OFF; enabled on section-editor's apply_edit step by
	// migration 474.
	// allow_rendered_html_transform gates the rendered_html_transform edit
	// type (bugs_open/277 §5): default OFF, enabled on section-editor's
	// apply_edit step by migration 513. Kept a config key rather than an
	// input so a caller cannot grant itself the authority the 2026-08-02 §2
	// ruling says must sit where a reviewer of the STEP can see it.
	// refuse_absent_required_fields is the bugs_open/342 refusal half:
	// default OFF, declines to persist an edit whose render left a
	// schema-required source:"llm" field empty (the live section keeps what
	// it had). A config key for the same reason as the two above — the
	// authority must sit on the STEP, not be grantable by the caller. Armed
	// by migration after the carrying binary rolls; see
	// mistyped_llm_fields_gate.go for the policy history.
	ConfigKeys: []string{"strip_literal_markdown", "allow_rendered_html_transform", "refuse_absent_required_fields"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{
		"edit_type_field": "edit_type",
		"content_data":    "replacement_content_data", // backward compat: old callers sending content_data
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_edit_context", LoadEditContextInputSpec)
	datahelpers.RegisterActionInputSpec("apply_section_edit", ApplySectionEditInputSpec)
}

// ===========================================================================
// ACTION: load_edit_context
// ===========================================================================
//
// Loads everything needed to perform or plan an edit on a page section.
//
// Target identification (one of):
//   - page_component_id: direct UUID of the page_components row
//   - page_name + slot_name: look up by page name and section slot
//
// Returns:
//   - page_component: {id, page_id, slot_name, position, rendered_html, content_data, component_id}
//   - component:      {id, function, name, html_template, input_schema} (from content_components)
//   - page:           {id, name, title, url, filename, domain, site_id, meta_desc}
//   - site_id, domain

func LoadEditContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("LoadEditContextAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadEditContextInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// --- Identify target page_component ---
	var pcRow pageComponentRow

	if pcIDStr := inputs.Get("page_component_id"); pcIDStr != "" {
		pcID, err := uuid.Parse(pcIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page_component_id: %w", err)
		}
		pcRow, err = loadPageComponentByID(ctx, params.DB, pcID)
		if err != nil {
			return nil, fmt.Errorf("page_component not found: %w", err)
		}
	} else {
		pageName := inputs.Get("page_name")
		slotName := inputs.Get("slot_name")
		if pageName == "" || slotName == "" {
			return nil, fmt.Errorf("need either page_component_id or both page_name + slot_name")
		}
		// Normalize slot_name per naming contract
		slotName = NormalizeComponentFunction(slotName)
		pcRow, err = loadPageComponentBySlot(ctx, params.DB, siteID, pageName, slotName)
		if err != nil {
			return nil, fmt.Errorf("page_component not found for %s/%s: %w", pageName, slotName, err)
		}
	}

	// --- Load page info ---
	pageInfo, err := getPageInfo(ctx, params.DB, pcRow.PageID)
	if err != nil {
		return nil, fmt.Errorf("failed to load page info: %w", err)
	}

	// --- Load component template (if linked) ---
	var componentData map[string]interface{}
	if pcRow.ComponentID != nil {
		componentData, err = loadComponentForEdit(ctx, params.DB, *pcRow.ComponentID)
		if err != nil {
			logger.Warn("LoadEditContextAction: Failed to load component template",
				zap.String("component_id", pcRow.ComponentID.String()),
				zap.Error(err),
			)
		}
	}

	// If no component_id, try lookup by slot_name
	if componentData == nil && pcRow.SlotName != "" {
		componentData, err = loadComponentByFunction(ctx, params.DB, pcRow.SlotName)
		if err != nil {
			logger.Debug("LoadEditContextAction: No component found for slot_name",
				zap.String("slot_name", pcRow.SlotName),
			)
		}
	}

	// --- Build result ---
	pcMap := map[string]interface{}{
		"id":            pcRow.ID.String(),
		"page_id":       pcRow.PageID.String(),
		"slot_name":     pcRow.SlotName,
		"position":      pcRow.Position,
		"rendered_html": pcRow.RenderedHTML,
		"content_data":  pcRow.ContentData,
		"build_status":  pcRow.BuildStatus,
	}
	if pcRow.ComponentID != nil {
		pcMap["component_id"] = pcRow.ComponentID.String()
	}

	pageMap := map[string]interface{}{
		"id":        pageInfo.ID.String(),
		"name":      pageInfo.Name,
		"title":     pageInfo.Title,
		"url":       pageInfo.URL,
		"filename":  pageInfo.Filename,
		"domain":    pageInfo.Domain,
		"site_id":   pageInfo.SiteID.String(),
		"meta_desc": pageInfo.MetaDesc,
	}

	logger.Info("LoadEditContextAction: Complete",
		zap.String("page_component_id", pcRow.ID.String()),
		zap.String("page_name", pageInfo.Name),
		zap.String("slot_name", pcRow.SlotName),
		zap.Bool("has_component_template", componentData != nil),
		zap.Bool("has_content_data", pcRow.ContentData != nil),
	)

	return map[string]interface{}{
		"success":        true,
		"page_component": pcMap,
		"component":      componentData, // may be nil
		"page":           pageMap,
		"site_id":        siteID.String(),
		"domain":         pageInfo.Domain,
	}, nil
}

// ===========================================================================
// ACTION: apply_section_edit
// ===========================================================================
//
// Performs the edit on a page_components row, then reassembles the full page.
//
// Reads edit_context from collected_data (output of load_edit_context).
//
// Edit types:
//   - content_edit:    Updates content_data (merge or replace), then re-renders
//                      the component template with full site context from DB.
//                      content_data is the source of truth — this ensures edits
//                      survive future re-renders.
//   - component_swap:  Changes the component template, re-renders with existing
//                      content_data + site context.
//
// After editing:
//   1. Update content_data in page_components
//   2. Re-render component template with updated content_data + site context
//   3. UPDATE page_components.rendered_html
//   4. Reassemble full page via assemblePage()
//   5. Return assembled HTML + metadata for git_commit

func ApplySectionEditAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("ApplySectionEditAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ApplySectionEditInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	editType := inputs.Get("edit_type")
	if editType == "" {
		return nil, fmt.Errorf("edit_type is required (content_edit or component_swap)")
	}

	// --- Load edit context from collected_data ---
	editCtx, ok := params.CollectedData["edit_context"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("edit_context not found in collected_data — run load_edit_context first")
	}

	pcData, _ := editCtx["page_component"].(map[string]interface{})
	if pcData == nil {
		return nil, fmt.Errorf("edit_context.page_component is missing")
	}

	pageData, _ := editCtx["page"].(map[string]interface{})
	if pageData == nil {
		return nil, fmt.Errorf("edit_context.page is missing")
	}

	pcIDStr, _ := pcData["id"].(string)
	pcID, err := uuid.Parse(pcIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page_component id: %w", err)
	}

	pageIDStr, _ := pageData["id"].(string)
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page id: %w", err)
	}

	siteIDStr, _ := editCtx["site_id"].(string)
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id in edit_context: %w", err)
	}

	slotName, _ := pcData["slot_name"].(string)
	pageName := getStringVal(pageData, "name")

	// ── Lock gate (bugs_open/058) ────────────────────────────────────────────
	// A human-locked component must not be overwritten by an automated edit
	// (this action is reachable from the tool-improver loop, not only from a
	// human request). Skip-result, not error: an error would fail/retry the
	// orchestration for a state only a human unlock can change. The check is
	// advisory for messaging; the race-free enforcement is the lock predicate
	// on the UPDATEs themselves (errComponentLocked below). A check failure is
	// non-fatal for the same reason.
	if lock, lockErr := CheckComponentLock(ctx, params.DB, pcID, logger); lockErr != nil {
		logger.Warn("ApplySectionEditAction: component lock check failed — relying on guarded UPDATE",
			zap.Error(lockErr))
	} else if lock.IsLocked {
		emitLockBlockedChangeItem(ctx, params.DB, siteID, &pageID, &pcID,
			pageName, slotName, lock.LockedBy, lock.LockType,
			"edit", "apply_section_edit", logger)
		logger.Warn("ApplySectionEditAction: refusing to edit human-locked component (bugs_open/058)",
			zap.String("page_component_id", pcIDStr),
			zap.String("slot_name", slotName),
			zap.String("locked_by", lock.LockedBy),
		)
		return map[string]interface{}{
			"success": true,
			"skipped": true,
			"locked":  true,
			"reason": fmt.Sprintf("component %s (%s) is locked by %q — unlock via the admin dashboard to edit it",
				pcIDStr, slotName, lock.LockedBy),
		}, nil
	}

	// ── Tombstone gate (bugs_open/360) ──────────────────────────────────────
	// build_status='removed' is the assembly-excluded tombstone: the row is
	// NOT on the page, so an edit here repairs nothing a visitor can see — and
	// every persist branch below promotes build_status to 'approved', which
	// UN-RETIRES the slot and the next rerender publishes the retired content
	// again (measured 2026-08-21: four literal_markdown transforms resurrected
	// four retired ported slots; the pages publicly served two stacked tools
	// for ~19 h). Skip-result, not error, mirroring the lock gate above; the
	// race-free enforcement is pageComponentNotRemovedSQL on the UPDATEs
	// themselves.
	if bs, _ := pcData["build_status"].(string); bs == "removed" {
		logger.Warn("ApplySectionEditAction: refusing to edit a removed (tombstoned) component (bugs_open/360)",
			zap.String("page_component_id", pcIDStr),
			zap.String("slot_name", slotName),
		)
		return map[string]interface{}{
			"success":    true,
			"skipped":    true,
			"tombstoned": true,
			"reason": fmt.Sprintf("component %s (%s) has build_status='removed' — it is retired from the page and an automated edit would resurrect it; if a finding names this slot, the finder is scanning content the page no longer serves",
				pcIDStr, slotName),
		}, nil
	}

	// ── Composition-parent gate (features_open/035 P1, direction 1) ─────────
	// A composition PARENT's template references {{.slots.*}}, which this
	// single-target path has no slots map for. Under missingkey=zero those
	// resolve to the empty string, so rendering a parent alone would persist a
	// row whose children have vanished — and report success while doing it. That
	// is the bugs_open/018 class arriving through an edit rather than a build.
	//
	// Skip-result, not error, matching the lock and tombstone gates above: only a
	// deliberate act (editing a child, or re-rendering the page) can change this
	// state, and an error would fail and retry an orchestration over it.
	//
	// Costs nothing live today: 0 of 2,249 page_components carry a
	// parent_instance_id (2026-08-31). It exists so the FIRST caller that tries is
	// told, rather than silently emptying a section. Membership comes from the one
	// shared query in component_hierarchy_walk.go — not a second spelling.
	if kids, kidsErr := hierarchyChildrenOf(ctx, params.DB, pcID); kidsErr != nil {
		// Fail OPEN, like the lock check: an unreadable membership query must not
		// block an edit on a page that probably has no composition at all.
		logger.Warn("ApplySectionEditAction: composition-parent check failed — proceeding",
			zap.String("page_component_id", pcIDStr), zap.Error(kidsErr))
	} else if len(kids) > 0 {
		logger.Warn("ApplySectionEditAction: refusing to render a composition parent alone (035 P1)",
			zap.String("page_component_id", pcIDStr),
			zap.String("slot_name", slotName),
			zap.Int("children", len(kids)),
		)
		return map[string]interface{}{
			"success":          true,
			"skipped":          true,
			"composite_parent": true,
			"reason": fmt.Sprintf("component %s (%s) is a composition parent with %d child row(s); rendering it alone would resolve its {{.slots.*}} to empty and drop them. Edit a child, or re-render the page.",
				pcIDStr, slotName, len(kids)),
		}, nil
	}

	// ── Decision citation gate (RFC_015) ────────────────────────────────────
	// If an active decision record covers this page/slot, the edit must NAME
	// it (acknowledges_decision or supersedes_decision) to proceed. Change is
	// allowed — regression by an item that did not know the decision existed
	// is not. Skip-result, not error, mirroring the lock gate above; a
	// coverage-check failure fails OPEN with a warning (same posture as the
	// lock check: the gate is advisory messaging, the decision trail is the
	// authority). Self-scoping: sites/slots with no covering decision rows
	// are untouched.
	if covered, covErr := CheckDecisionCoverage(ctx, params.DB, siteID, pageName, slotName); covErr != nil {
		logger.Warn("ApplySectionEditAction: decision coverage check failed — proceeding without gate",
			zap.Error(covErr))
	} else if len(covered) > 0 {
		citation := inputs.Get("acknowledges_decision") + "," + inputs.Get("supersedes_decision")
		if !CitationSatisfies(citation, covered) {
			logger.Warn("ApplySectionEditAction: refusing edit on decision-covered slot without citation (RFC_015)",
				zap.String("page", pageName),
				zap.String("slot", slotName),
				zap.String("covering_decisions", CoveredKeys(covered)),
			)
			return map[string]interface{}{
				"success":        true,
				"skipped":        true,
				"decision_gated": true,
				"decisions":      CoveredKeys(covered),
				"reason": fmt.Sprintf(
					"slot %q on page %q is covered by decision(s) %s — re-submit with acknowledges_decision (work within it) or supersedes_decision (replace it) naming the key; read the decision row in doc_notes first",
					slotName, pageName, CoveredKeys(covered)),
			}, nil
		}
		logger.Info("ApplySectionEditAction: decision citation accepted",
			zap.String("covering_decisions", CoveredKeys(covered)))
	}

	// --- Apply the edit ---
	var outcome sectionEditOutcome

	// bugs_open/184: when enabled (migration 474), literal markdown is stripped
	// from the merged content map INSIDE each branch, before its RenderTemplate
	// call — the :398/:407 pre-persist window below is too late for this
	// transform, because rendered_html is already built from the unstripped
	// values by then. One gated strip per branch, both feeding the single
	// persist switch (the "count writes, not branches" rule holds: the persist
	// sites are unchanged).
	stripMarkdown, _ := params.StepConfig.Config["strip_literal_markdown"].(bool)

	switch editType {
	case "content_edit":
		outcome, err = applyContentEdit(ctx, params.DB, siteID, pcData, editCtx, inputs, stripMarkdown, logger)
	case "component_swap":
		outcome, err = applyComponentSwap(ctx, params.DB, siteID, pcData, editCtx, inputs, stripMarkdown, logger)
	case "rendered_html_transform":
		outcome, err = applyRenderedHTMLTransform(pcData, inputs, params.StepConfig.Config, logger)
	default:
		return nil, fmt.Errorf("unknown edit_type: %s (expected: content_edit, component_swap, rendered_html_transform)", editType)
	}

	if err != nil {
		if errors.Is(err, errComponentLocked) {
			// Race window: the lock landed between the pre-check and the write.
			return map[string]interface{}{
				"success": true,
				"skipped": true,
				"locked":  true,
				"reason":  fmt.Sprintf("component %s (%s) is locked — unlock via the admin dashboard to edit it", pcIDStr, slotName),
			}, nil
		}
		return nil, fmt.Errorf("edit failed (%s): %w", editType, err)
	}

	// --- Refuse to persist a section whose required fields rendered empty ---
	// (bugs_open/342, the refusal half.) The branches above have already
	// published which schema-required source:"llm" fields the render left
	// EMPTY, and already filed the required_fields_missing item — so a refusal
	// here loses no record. What it prevents is the write: these are the two
	// routes that put rendered_html straight onto an already-live page with no
	// validate_content between, and until now they filed the note and shipped
	// the blank anyway. ONE gate at the ONE persist switch, same placement
	// rule as the link repair and the envelope refusal below, so a future edit
	// branch inherits it.
	//
	// OPT-IN, default OFF (owner ruling 2026-08-02 §2): an edit like this
	// SUCCEEDS today, so refusing is new authority and arrives as a step
	// config field. Armed on section-editor's apply_edit by migration, only
	// after a binary carrying this code has rolled.
	if refusePersistForAbsentRequired(params.StepConfig.Config, outcome.AbsentRequiredFields) {
		return nil, fmt.Errorf(
			"apply_section_edit (%s) on %q: refusing to persist — %d schema-required field(s) rendered empty (%s); "+
				"the live section is left unchanged and a required_fields_missing item has been filed (bugs_open/342)",
			editType, slotName, len(outcome.AbsentRequiredFields), strings.Join(outcome.AbsentRequiredFields, ", "))
	}

	domain, _ := pageData["domain"].(string)
	pageURL := getStringVal(pageData, "url")

	// --- Dead-internal-link repair (bugs_open/136) ---
	// ONE call, before ONE persist switch. This action is where an LLM rewrites
	// a single section and the result goes straight to
	// page_components.rendered_html — the same freedom to invent /pricing that
	// produced 079's evidence, on the path CLAUDE.md itself points sessions at
	// for targeted edits. Before this, the swap branch persisted inside
	// applyComponentSwap and the content_edit branch persisted here, so a repair
	// at either point would have left the other open — this bug's own shape, one
	// level in. The swap's UPDATE moved out to the switch below so that cannot
	// be true again.
	outcome.HTML = repairComponentHTMLBeforePersist(ctx, params, siteID,
		domain, pageName, pageURL, "apply_section_edit", outcome.HTML, logger)

	// --- Refuse or decode a stored LLM transport envelope (bugs_open/190) ---
	// Same reasoning as the repair above, and placed for the same reason: ONE call
	// before ONE persist switch. applyContentEdit seeds its map from the existing
	// row, so a field_updates merge onto an already-poisoned row carries type and
	// result forward untouched; replacement_content_data is agent-supplied and can
	// be a raw step output. Refusing here means neither branch can persist one.
	if normalized, changed, envErr := normalizeContentDataEnvelope(outcome.ContentData); envErr != nil {
		return nil, fmt.Errorf("apply_section_edit (%s) on %q: %w", editType, slotName, envErr)
	} else if changed {
		logger.Warn("ApplySectionEditAction: decoded an LLM transport envelope out of content_data (bugs_open/190)",
			zap.String("page_name", pageName),
			zap.String("slot_name", slotName))
		outcome.ContentData = normalized
	}

	// Classify before the persist (bugs_open/229, council round 1
	// bug_historian: loudness must not stop at the two rebuild paths — that is
	// the "one call site rigorous, sibling heuristic" pattern). Advisory; the
	// 357 trigger archives the outgoing bytes whichever way this goes.
	divergent, classifyErr := classifyPageComponentArtefacts(ctx, params.DB, pageID)
	if classifyErr != nil {
		logger.Warn("ApplySectionEditAction: divergence classification failed — edit proceeds, the 357 trigger still archives (bugs_open/229)",
			zap.Error(classifyErr))
	}

	// --- Persist the page_components row ---
	switch editType {
	case "content_edit":
		// Per-slot floors, SAME rules as a whole-page save (council round 1 on
		// b30ac52c: both floors were wired only into SavePageSectionsAction, so
		// this path — the one decomposition exists to enable — bypassed them
		// and a flattening here would have failed as silently as the bug they
		// fix). Only content_edit: a component_swap deliberately changes the
		// component, so its markup is SUPPOSED to change.
		existingHTML, _ := pcData["rendered_html"].(string)
		if floorErr := enforceSingleSlotFloors(ctx, params, siteID, pageID,
			pageName, slotName, existingHTML, outcome.HTML); floorErr != nil {
			return nil, floorErr
		}
		// Regulated-identity refusal (CGV-033). Wired HERE for the same reason
		// the floors above are: a guard wired only into SavePageSectionsAction
		// has a hole exactly the shape of this path. Regulated family ONLY, so
		// no other claim behaviour on this seam changes — see
		// section_editor_regulated_guard.go.
		if regErr := refuseRegulatedIdentityEdit(ctx, params, siteID,
			pageName, slotName, outcome.HTML, logger); regErr != nil {
			return nil, regErr
		}
		err = updatePageComponentAfterEdit(ctx, params.DB, pcID, outcome.HTML, outcome.ContentData)
	case "rendered_html_transform":
		// Same guard posture as content_edit — the floors and the regulated
		// check judge the OUTGOING html, and this branch writes that column
		// like any other. (A code-span conversion shrinks visible text by two
		// backticks per span, so the floors pass trivially on a correct
		// transform and refuse a wrong one — which is the point of wiring
		// them rather than reasoning they cannot fire.)
		existingHTML, _ := pcData["rendered_html"].(string)
		if floorErr := enforceSingleSlotFloors(ctx, params, siteID, pageID,
			pageName, slotName, existingHTML, outcome.HTML); floorErr != nil {
			return nil, floorErr
		}
		if regErr := refuseRegulatedIdentityEdit(ctx, params, siteID,
			pageName, slotName, outcome.HTML, logger); regErr != nil {
			return nil, regErr
		}
		// HTML-only persist: content_data is deliberately NOT written — on
		// this population it is provenance metadata, not source, and the
		// transform must not be able to touch it (nil takes
		// updatePageComponentAfterEdit's html-only UPDATE branch).
		err = updatePageComponentAfterEdit(ctx, params.DB, pcID, outcome.HTML, nil)
	case "component_swap":
		// Regulated-identity refusal (CGV-033) on the THIRD write branch. Round 3
		// of the council round on correlation aac38d5b objected that the guard was
		// wired into one branch and asked whether other paths write the same
		// column. Checked: this switch has three persisting branches and this was
		// the only one left unguarded — content_edit was wired by the original
		// change, rendered_html_transform by the lane that added it (af0f00bb5),
		// and this one by nobody. A swap writes rendered_html AND content_data,
		// so it can carry the claim just as the other two can.
		if regErr := refuseRegulatedIdentityEdit(ctx, params, siteID,
			pageName, slotName, outcome.HTML, logger); regErr != nil {
			return nil, regErr
		}
		err = updatePageComponentSwap(ctx, params.DB, pcID,
			outcome.ComponentID, outcome.SlotName, outcome.HTML, outcome.ContentData)
	}
	if errors.Is(err, errComponentLocked) {
		return map[string]interface{}{
			"success": true,
			"skipped": true,
			"locked":  true,
			"reason":  fmt.Sprintf("component %s (%s) is locked — unlock via the admin dashboard to edit it", pcIDStr, slotName),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to persist section edit (%s): %w", editType, err)
	}

	// Emit only for THIS component and only after the persist succeeded (the
	// after-RowsAffected rule; the locked path returned above, so a refused
	// write cannot reach this).
	if classifyErr == nil {
		var mine []pageComponentDivergence
		for _, d := range divergent {
			if d.ComponentID == pcID {
				mine = append(mine, d)
			}
		}
		emitPageDivergenceItems(ctx, params.DB, pageID, pageName, mine, "apply_section_edit", logger)
	}

	logger.Info("ApplySectionEditAction: Updated page_component",
		zap.String("page_component_id", pcIDStr),
		zap.String("edit_type", editType),
		zap.String("slot_name", slotName),
		zap.Int("new_html_length", len(outcome.HTML)),
	)

	// ── Recompose the ancestors this row is embedded in (035 P1, direction 2) ──
	//
	// A composed page serves the TOPMOST ancestor's bytes, not the child's. So a
	// child edit that stops at the child is invisible on the page it just changed,
	// and the page keeps serving the pre-edit text while this action reports
	// success — bugs_open/384's shape, one level in.
	//
	// PLACEMENT IS LOAD-BEARING, on both sides:
	//   AFTER the persist, because the ancestors must embed the NEW child bytes,
	//   and because the child's UPDATE has committed by here (this action holds no
	//   transaction) so an ordinary read sees it;
	//   BEFORE assemblePage below, because the reassembly reads page_components
	//   back — recomposing after it would ship the stale parent in the very HTML
	//   this action returns for deployment.
	//
	// It cannot fail the edit (see the file header of component_hierarchy_recompose.go).
	// An ancestor it could not refresh — unreadable, unrenderable, floor-breaching,
	// locked or tombstoned — comes back as a slot name and is published in the
	// result below, where it is a durable record rather than a log line that
	// rotates within minutes.
	//
	// COSTS ONE INDEXED SELECT PER EDIT AND NOTHING ELSE TODAY: hierarchyAncestorChain
	// reads parent_instance_id for the edited row and returns empty on a top-level
	// one, and `[MEASURED 2026-09-03]` 0 of 3,229 page_components carry a
	// parent_instance_id, so the loop body is unreachable on today's data. That is
	// the RFC_022 shape this feature ships in: opt-in, unsafe side OFF, inert until
	// a row opts in.
	staleAncestorSlots := recomposeAncestors(ctx, params, params.DB, pcID, siteID, logger)
	if len(staleAncestorSlots) > 0 {
		logger.Warn("ApplySectionEditAction: some ancestors could not be recomposed — the page may serve stale parent bytes (035 P1)",
			zap.String("page_component_id", pcIDStr),
			zap.Strings("stale_ancestor_slots", staleAncestorSlots))
	}

	// --- Reassemble full page ---
	pageInfo, err := getPageInfo(ctx, params.DB, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to load page info for reassembly: %w", err)
	}

	fullHTML, assembly, err := assemblePage(ctx, params.DB, pageInfo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to reassemble page: %w", err)
	}

	// bugs_open/095, same class one surface over: this path returned
	// success=true with html="" if reassembly produced nothing — i.e. it would
	// report that a section edit had been applied while handing the deployer an
	// empty page.
	//
	// It uses the SAME discriminator as the re-renderer rather than failing on
	// any empty reassembly. The first version of this fix failed
	// unconditionally, on the reasoning that a section of this page has just
	// been edited so its components necessarily exist. Two council seats
	// objected at medium severity that this was asserted rather than evidenced,
	// and that an error with no escape hatch is a sharper change than the
	// re-render side, which keeps a legitimate-skip branch. They were right:
	// where the two surfaces agree on the rule, the rule is the thing that has
	// been reviewed. And it costs nothing here — the edit target IS a component
	// row, so ComponentRows > 0 holds whenever the edit actually happened.
	if fullHTML == "" && assembly.assembledToNothingDespiteComponents() {
		return nil, fmt.Errorf(
			"page %q reassembled to nothing after the section edit — %s",
			pageInfo.Name, assembly.describe())
	}

	logger.Info("ApplySectionEditAction: Page reassembled",
		zap.String("page_name", pageInfo.Name),
		zap.String("filename", pageInfo.Filename),
		zap.Int("full_html_length", len(fullHTML)),
	)

	return map[string]interface{}{
		"success":           true,
		"html":              fullHTML,
		"domain":            domain,
		"filename":          pageInfo.Filename,
		"page_id":           pageIDStr,
		"page_name":         pageInfo.Name,
		"page_component_id": pcIDStr,
		"edit_type":         editType,
		"slot_name":         slotName,
		// Merge-vs-replace provenance (bugs_open/260 §9d): before these keys the
		// mode existed only in log lines that rotate, so matching collected_data
		// for the "field_updates"/"replacement_content_data" KEY NAMES returned
		// every run — the action's config echo carries both names regardless of
		// use. content_edit_mode is empty for component_swap.
		"content_edit_mode":   outcome.ContentEditMode,
		"updated_field_count": outcome.UpdatedFieldCount,
		// bugs_open/184: durable record of a content-mutating strip — pod logs
		// rotate in minutes; collected_data survives (council 060bcc0a r2).
		"stripped_markdown_fields": outcome.StrippedMarkdownFields,
		"total_field_count":        outcome.TotalFieldCount,
		// bugs_open/277 §5: durable record of a rendered_html transform —
		// same reasoning. Empty/zero for the other two edit types.
		"transform_name":      outcome.TransformName,
		"transform_converted": outcome.TransformConverted,
		// 035 P1 direction 2: the ancestors this edit could NOT refresh, so a
		// parent left embedding stale bytes is queryable in collected_data
		// instead of only in a pod log. Empty on every page that is not composed,
		// which today is all of them.
		"stale_ancestor_slots": staleAncestorSlots,
	}, nil
}

// ===========================================================================
// buildRenderContextFromDB
// ===========================================================================
//
// Builds a full RenderContext from database state alone — no collected_data
// pipeline needed. This is the key function that enables section editing
// without re-running the entire content generation pipeline.
//
// Loads:
//   - Site data (company name, domain, email, phone, logo)
//   - Style collection (colors)
//   - CSS theme (theme_css)
//   - Navigation items (header + footer)
//   - Page metadata (title, description, current page)
//   - Content data (from page_components.content_data → RenderContext.ContentData)

func buildRenderContextFromDB(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageInfo *PageInfo,
	contentData map[string]interface{},
	logger *zap.Logger,
) (*RenderContext, error) {

	// 1. Load site data
	siteData, err := loadSiteDataFull(ctx, db, siteID, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load site data: %w", err)
	}

	// 2. Load style collection for colors
	var primaryColor, secondaryColor, accentColor string
	coll, err := GetStyleCollectionForSite(ctx, db, siteID, logger)
	if err != nil {
		logger.Warn("buildRenderContextFromDB: No style collection found, using defaults",
			zap.Error(err))
	} else if coll != nil && coll.ColorPalette != nil {
		primaryColor = coll.ColorPalette["primary"]
		secondaryColor = coll.ColorPalette["secondary"]
		accentColor = coll.ColorPalette["accent"]
	}

	// 3. Load CSS theme
	var themeCSS string
	if coll != nil && coll.CSSThemeID != nil {
		theme, err := getThemeByID(ctx, db, *coll.CSSThemeID, logger)
		if err != nil {
			logger.Warn("buildRenderContextFromDB: Failed to load CSS theme",
				zap.Error(err))
		} else if theme != nil {
			themeCSS = theme.CSSContent
		}
	}

	// 4. Load navigation
	navItems := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary}, NavFetchableOnly, 0, logger)
	footerNavItems := GetNavItems(ctx, db, siteID, []string{NavGroupPrimary, NavGroupUtility, NavGroupLegal}, NavFetchableOnly, 0, logger)

	// 5. Derive page-specific fields
	currentPage := ""
	if pageInfo != nil {
		currentPage = strings.TrimSuffix(pageInfo.Filename, ".html")
	}
	year := fmt.Sprintf("%d", time.Now().Year())

	// 6. Build the RenderContext
	renderCtx := &RenderContext{
		Domain:         siteData.Domain,
		SiteID:         siteID,
		CompanyName:    siteData.CompanyName,
		LogoText:       siteData.LogoText,
		LogoURL:        siteData.LogoURL,
		Tagline:        siteData.Tagline,
		Email:          siteData.Email,
		Phone:          siteData.Phone,
		PrimaryColor:   primaryColor,
		SecondaryColor: secondaryColor,
		AccentColor:    accentColor,
		ThemeCSS:       themeCSS,
		NavItems:       setActiveNavItems(navItems, currentPage),
		FooterNavItems: footerNavItems,
		CurrentPage:    currentPage,
		Year:           year,
	}

	if pageInfo != nil {
		renderCtx.Title = pageInfo.Title
		renderCtx.Description = pageInfo.MetaDesc
	}

	// bugs_open/420: NO fallback email from the domain. This used to synthesise
	// "info@<domain>" whenever the column was empty — a fabricated address that
	// nobody owns and no mailbox answers, published as though the business had
	// chosen it. Now that the owner's ruling is "publish nothing unless asked"
	// (2026-08-31), the empty column is the CORRECT and common state, and
	// synthesising here would make "the site publishes no contact" quietly
	// false. Absence renders as absence: the contact blocks are gated on a
	// non-empty email (bugs_open/111).

	// 7. Set ContentData — this is what templates use for section-specific content.
	//    contextToInterfaceMap() merges ContentData into the top-level template data,
	//    so templates access these as {{.headline}}, {{.features}}, etc.
	renderCtx.ContentData = make(map[string]interface{})

	// Site-level fields that templates might reference
	renderCtx.ContentData["company_name"] = siteData.CompanyName
	renderCtx.ContentData["brand_name"] = siteData.CompanyName
	renderCtx.ContentData["tagline"] = siteData.Tagline
	renderCtx.ContentData["domain"] = siteData.Domain
	renderCtx.ContentData["email"] = siteData.Email
	renderCtx.ContentData["contact_email"] = siteData.Email
	renderCtx.ContentData["phone"] = siteData.Phone
	renderCtx.ContentData["logo_text"] = siteData.LogoText
	renderCtx.ContentData["logo_url"] = siteData.LogoURL
	renderCtx.ContentData["year"] = year
	renderCtx.ContentData["copyright"] = fmt.Sprintf("© %s %s", year, siteData.CompanyName)

	// Navigation in the formats templates expect
	categories := make([]map[string]interface{}, len(navItems))
	for i, item := range navItems {
		categories[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label,
		}
	}
	renderCtx.ContentData["categories"] = categories
	renderCtx.ContentData["nav_items"] = categories

	footerLinks := make([]map[string]interface{}, len(footerNavItems))
	for i, item := range footerNavItems {
		footerLinks[i] = map[string]interface{}{
			"name":  item.Label,
			"slug":  strings.TrimSuffix(strings.TrimPrefix(item.URL, "/"), ".html"),
			"url":   item.URL,
			"label": item.Label,
		}
	}
	renderCtx.ContentData["footer_nav_items"] = footerLinks
	renderCtx.ContentData["quick_links"] = footerLinks

	// NO CTA defaults, deliberately (bugs_open/203's class, found here
	// 2026-08-22). A cta_url seeded ahead of the merge below manufactures the
	// very condition the templates' {{if ... cta_url}} guards test for, so a
	// section whose content carries no CTA shipped a fabricated
	// "Get Started" -> /contact.html anchor — a 404 on any site without that
	// page, and unfixable downstream because no data-side cleanup can express
	// "no button". Correct-or-absent is the ruled pattern: see
	// component_library.go's contextToInterfaceMap (LNK-005), which sets
	// cta_url from the context and invents nothing.

	// Now merge the actual section content_data on top — these take priority
	// over site-level defaults. This is where headline, features[],
	// testimonials[], body text etc. come from.
	for key, value := range contentData {
		renderCtx.ContentData[key] = value
	}

	logger.Info("buildRenderContextFromDB: Context built",
		zap.String("domain", siteData.Domain),
		zap.String("company", siteData.CompanyName),
		zap.Bool("has_colors", primaryColor != ""),
		zap.Bool("has_theme_css", themeCSS != ""),
		zap.Int("nav_items", len(navItems)),
		zap.Int("content_data_fields", len(contentData)),
	)

	return renderCtx, nil
}

// ===========================================================================
// EDIT IMPLEMENTATIONS
// ===========================================================================

// applyContentEdit updates content_data and re-renders the component template
// with full site context from DB.
//
// Two modes for specifying the update:
//   - field_updates: merge specific fields into existing content_data
//     (e.g. change headline, update a phone number)
//   - content_data:  replace entire content_data with new object
//     (e.g. rewrite a whole case study)
//
// After updating content_data, loads the component template and builds a
// full RenderContext from DB state, then calls RenderTemplate.
// sectionEditOutcome is what an edit PRODUCED, before anything is written.
//
// The two edit types used to differ in where they persisted — content_edit
// returned its HTML for the caller to write, component_swap wrote its own row
// and returned the HTML as well. That gave the action two persist sites, so any
// guard placed before one of them left the other open (bugs_open/136). Both now
// return this, and ApplySectionEditAction repairs once and persists once.
//
// ComponentID and SlotName are set by component_swap only; content_edit leaves
// them zero because it changes neither.
//
// ContentEditMode, UpdatedFieldCount and TotalFieldCount are set by content_edit
// only; component_swap leaves them zero. ContentEditMode names the input key the
// edit resolved ("field_updates" or "replacement_content_data") — before this,
// the merge-vs-replace decision existed only as Info log lines that rotate, so
// which mode a run took was unanswerable from the DB (bugs_open/260 §9d; only a
// replacement can retype a field the agent did not name, so the split matters
// when auditing type damage). It is returned to the caller's result map, which
// lands in collected_data.
type sectionEditOutcome struct {
	HTML              string
	ContentData       map[string]interface{}
	ComponentID       uuid.UUID
	SlotName          string
	ContentEditMode   string
	UpdatedFieldCount int
	TotalFieldCount   int
	// Field paths StripLiteralMarkdown changed (bugs_open/184). Surfaced on
	// the action result — pod logs rotate in minutes; collected_data is the
	// durable record of a content-mutating transform (council 060bcc0a r2).
	StrippedMarkdownFields []string
	// rendered_html_transform provenance (bugs_open/277 §5): which named
	// transform ran and how many sites it converted. Same durability
	// reasoning as StrippedMarkdownFields — both surfaced on the result.
	TransformName      string
	TransformConverted int
	// Schema-required source:"llm" fields the render left EMPTY, copied from
	// RenderContext.AbsentRequiredFields by the branches that render
	// (bugs_open/342). Carried on the outcome so the refusal decision can sit
	// at the ONE persist switch in ApplySectionEditAction — the same reason
	// the link repair and the envelope refusal live there and not in the
	// branches: a future edit branch inherits the gate instead of
	// re-remembering it. Empty on the rendered_html_transform branch, which
	// re-renders nothing.
	AbsentRequiredFields []string
}

// applyRenderedHTMLTransform is the rendered_html_transform branch: apply one
// NAMED deterministic transform to the component's EXISTING rendered_html.
// It never renders a template and never touches content_data (ContentData is
// left nil so the caller's persist takes the html-only branch). Gated by the
// allow_rendered_html_transform config key — see the input spec comment.
//
// Every refusal is an ERROR, not a skip: a config-off, unknown-transform or
// matched-nothing outcome routes the item into the attempt machinery and then
// to a human, which is the correct destination for a repair that could not be
// performed — a skip would read as success to the check_edit_skipped branch.
func applyRenderedHTMLTransform(
	pcData map[string]interface{},
	inputs *datahelpers.ActionInputs,
	stepConfig map[string]interface{},
	logger *zap.Logger,
) (sectionEditOutcome, error) {

	if allowed, _ := stepConfig["allow_rendered_html_transform"].(bool); !allowed {
		return sectionEditOutcome{}, fmt.Errorf(
			"edit_type rendered_html_transform is not enabled on this step " +
				"(config key allow_rendered_html_transform, default OFF — migration 513 enables it " +
				"on section-editor's apply_edit only); live section left unchanged")
	}

	transformName := inputs.Get("transform_name")
	existingHTML, _ := pcData["rendered_html"].(string)
	if strings.TrimSpace(existingHTML) == "" {
		return sectionEditOutcome{}, fmt.Errorf(
			"rendered_html_transform %q: component has no rendered_html — this edit type edits finished HTML only", transformName)
	}

	// THIS SWITCH IS THE TRANSFORM REGISTRY, and it accumulates INVISIBLY to
	// the RFC_022 optional-key counter (council corr b72a4029 r1, architecture
	// seat): transform_name is one key however many transforms hide behind it.
	// So the rule is stated here, where the case would be added: EACH NEW
	// NAMED TRANSFORM IS ITS OWN COUNCIL ROUND with its own safety argument —
	// this one's lives in rendered_html_code_spans.go's header — and gets its
	// own line in register CQ-028. A transform is deterministic, never
	// LLM-fed, or it does not belong in this edit type at all.
	switch transformName {
	case "code_span_to_code_tag":
		out, converted, err := datahelpers.ConvertLiteralCodeSpansInHTML(existingHTML)
		if err != nil {
			return sectionEditOutcome{}, fmt.Errorf("rendered_html_transform code_span_to_code_tag refused: %w", err)
		}
		if converted == 0 {
			// The finding exists (something routed this item here) but the
			// conversion pattern cannot reach it — a span crossing inline
			// elements, an entity-encoded backtick, or a finding already
			// repaired. Deploying identical bytes and reporting an edit would
			// be the exact "complete is not proof" shape; refuse instead.
			return sectionEditOutcome{}, fmt.Errorf(
				"rendered_html_transform code_span_to_code_tag matched nothing in %d bytes — "+
					"the finding is not reachable by this transform (conversion is narrower than detection, by design); "+
					"live section left unchanged", len(existingHTML))
		}
		logger.Info("applyRenderedHTMLTransform: converted code spans",
			zap.Int("converted", converted),
			zap.Int("html_bytes_in", len(existingHTML)),
			zap.Int("html_bytes_out", len(out)))
		return sectionEditOutcome{
			HTML:               out,
			ContentData:        nil, // deliberate: html-only persist
			ContentEditMode:    "rendered_html_transform",
			TransformName:      transformName,
			TransformConverted: converted,
		}, nil
	default:
		return sectionEditOutcome{}, fmt.Errorf(
			"unknown transform_name %q (known: code_span_to_code_tag); live section left unchanged", transformName)
	}
}

func applyContentEdit(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pcData map[string]interface{},
	editCtx map[string]interface{},
	inputs *datahelpers.ActionInputs,
	stripMarkdown bool,
	logger *zap.Logger,
) (sectionEditOutcome, error) {

	// --- Build updated content_data ---
	existingContentData := make(map[string]interface{})
	if cd, ok := pcData["content_data"].(map[string]interface{}); ok {
		for k, v := range cd {
			existingContentData[k] = v
		}
	}

	// Which input key this edit resolved, and how many fields it touched —
	// recorded on the outcome so the mode survives into collected_data rather
	// than living only in the Info lines below (bugs_open/260 §9d).
	editMode := ""
	updatedFields := 0

	// Check field_updates first (merge mode — more common, more specific)
	// This is checked BEFORE replacement_content_data because ExtractActionInputs
	// nested lookup can pick up site_record.content_data (the site plan) as a
	// false match for any field named "content_data".
	if fieldUpdates := inputs.GetRaw("field_updates"); fieldUpdates != nil {
		// Merge mode — update specific fields
		var updates map[string]interface{}
		switch v := fieldUpdates.(type) {
		case map[string]interface{}:
			updates = v
		case string:
			if err := json.Unmarshal([]byte(v), &updates); err != nil {
				return sectionEditOutcome{}, fmt.Errorf("field_updates must be valid JSON object: %w", err)
			}
		default:
			return sectionEditOutcome{}, fmt.Errorf("field_updates must be a JSON object, got %T", fieldUpdates)
		}
		for k, v := range updates {
			existingContentData[k] = v
			logger.Debug("applyContentEdit: Merged field", zap.String("field", k))
		}
		editMode = "field_updates"
		updatedFields = len(updates)
		logger.Info("applyContentEdit: Merged field_updates into content_data",
			zap.Int("updated_fields", len(updates)),
			zap.Int("total_fields", len(existingContentData)))
	} else if fullReplace := inputs.GetRaw("replacement_content_data"); fullReplace != nil {
		// Full replacement mode
		editMode = "replacement_content_data"
		switch v := fullReplace.(type) {
		case map[string]interface{}:
			existingContentData = v
			logger.Info("applyContentEdit: Full content_data replacement",
				zap.Int("field_count", len(v)))
		case string:
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err != nil {
				return sectionEditOutcome{}, fmt.Errorf("replacement_content_data must be valid JSON object: %w", err)
			}
			existingContentData = parsed
			logger.Info("applyContentEdit: Full content_data replacement (from JSON string)",
				zap.Int("field_count", len(parsed)))
		}
		updatedFields = len(existingContentData)
	} else {
		return sectionEditOutcome{}, fmt.Errorf("content_edit requires either 'field_updates' or 'replacement_content_data' parameter")
	}

	// --- Get component template ---
	componentData, _ := editCtx["component"].(map[string]interface{})
	if componentData == nil {
		return sectionEditOutcome{}, fmt.Errorf("no component template available — cannot re-render (component_id may be NULL)")
	}
	htmlTemplate, _ := componentData["html_template"].(string)
	if htmlTemplate == "" {
		return sectionEditOutcome{}, fmt.Errorf("component template is empty — cannot re-render")
	}

	// bugs_open/184: strip literal markdown from the MERGED map (LLM
	// field_updates included) before the render below, so rendered_html and the
	// persisted ContentData are both built from clean values. Gated, default OFF.
	var strippedMarkdownFields []string
	if stripMarkdown {
		if changed := datahelpers.StripLiteralMarkdownFromContentData(existingContentData); len(changed) > 0 {
			strippedMarkdownFields = changed
			logger.Info("applyContentEdit: stripped literal markdown",
				zap.Strings("fields", changed))
		}
	}

	// --- Build render context from DB ---
	pageData, _ := editCtx["page"].(map[string]interface{})
	var pageInfoForRender *PageInfo
	if pageData != nil {
		pageID, _ := uuid.Parse(getStringVal(pageData, "id"))
		pageInfoForRender = &PageInfo{
			ID:       pageID,
			Name:     getStringVal(pageData, "name"),
			Title:    getStringVal(pageData, "title"),
			Filename: getStringVal(pageData, "filename"),
			MetaDesc: getStringVal(pageData, "meta_desc"),
			Domain:   getStringVal(pageData, "domain"),
			SiteID:   siteID,
		}
	}

	renderCtx, err := buildRenderContextFromDB(ctx, db, siteID, pageInfoForRender, existingContentData, logger)
	if err != nil {
		return sectionEditOutcome{}, fmt.Errorf("failed to build render context: %w", err)
	}

	// --- Render ---
	// One section — but this path holds the STORED ROW, so it knows its page and
	// its 1-based position and can count its same-function predecessors exactly
	// (component_instance_occurrence.go, bugs_closed/283 / RFC_032 step 3).
	// Before this it bound occurrence 0 unconditionally, which put every instance
	// on a multi-instance page back on identical element ids after any edit.
	// If the count cannot be taken it still binds 0 — never worse than before,
	// and it cannot fail the edit.
	DeriveAndBindInstanceToken(ctx, db, renderCtx,
		getStringVal(componentData, "function"), PlacementFromStoredRow(pcData), logger)
	// bugs_open/342 — loadComponentForEdit already selects input_schema, so this
	// path has the contract in hand. It matters MORE here than anywhere: this
	// route writes rendered_html straight to an already-live page with no
	// validate_content between it and the reader, so a required field that
	// renders empty ships blank to a page that is currently serving.
	if schema, ok := componentData["input_schema"].(map[string]interface{}); ok {
		renderCtx.InputSchema = schema
	}
	rendered, _, _, err := RenderTemplate(htmlTemplate, renderCtx, logger)
	// ESCALATE, not just detect (council bb7f5d0e round 5 — editquality AND
	// bug_historian, both HIGH, both making the same point against my own
	// words). The submission called these two routes "the two with the most
	// exposure" because they write rendered_html straight to an already-live
	// page with no validate_content between, and then wired only the log here.
	// If they are the riskiest, they are the ones that most need the queue
	// entry; detection-only on the highest-exposure path is the "one call site
	// gets the rigorous fix, the sibling stays heuristic" shape 016b §9 names.
	//
	// The page context is what makes the item ROUTABLE — see the emitter's own
	// comment. Without it the router classifies `malformed` and the ladder parks
	// the item, which is what happened to the first one this path ever filed.
	emitRequiredFieldsMissing(ctx, db, siteID,
		editPageContext(pageInfoForRender, pcData),
		nil, getStringVal(componentData, "function"),
		fmt.Sprintf("Section edit on %s", getStringVal(componentData, "function")),
		"page_component", "section_editor", renderCtx.AbsentRequiredFields,
		map[string]interface{}{"route": "content_edit", "live_page": true}, logger)
	if err != nil {
		// bugs_open/260. THIS PATH HAS NO GATE DOWNSTREAM: the caller writes
		// rendered_html straight to an already-live page, with no
		// validate_content between — so the seam's error IS the gate, and
		// refusing here is what leaves the live section untouched.
		//
		// The `rendered == ""` check below is kept but was never able to catch
		// this: the deleted regex fallback returned MANGLED html, never empty,
		// so this route could ship a section with its {{if}}/{{range}}
		// directives intact to a page already serving traffic.
		return sectionEditOutcome{}, fmt.Errorf("template rendering failed, live section left unchanged: %w", err)
	}
	if rendered == "" {
		return sectionEditOutcome{}, fmt.Errorf("template rendering produced empty output")
	}

	logger.Info("applyContentEdit: Re-rendered component from template",
		zap.Int("output_length", len(rendered)),
		zap.Int("content_data_fields", len(existingContentData)),
	)

	return sectionEditOutcome{
		HTML:                   rendered,
		ContentData:            existingContentData,
		ContentEditMode:        editMode,
		UpdatedFieldCount:      updatedFields,
		TotalFieldCount:        len(existingContentData),
		StrippedMarkdownFields: strippedMarkdownFields,
		AbsentRequiredFields:   renderCtx.AbsentRequiredFields,
	}, nil
}

// editPageContext assembles the page identity a required_fields_missing item
// needs to be ROUTABLE, from the two things both edit branches already hold.
// One helper rather than two literals because the two branches filing the same
// item type with different context is precisely how the editor route came to
// file unroutable items in the first place (bugs_open/342, 2026-08-23).
//
// A nil PageInfo yields the zero pageContext, which the emitter reads as "no
// page" and files for human review instead of handing it to a page-resolving
// router. That is the correct behaviour, not a degraded one: an item nobody can
// route is better parked visibly than failed three times.
func editPageContext(page *PageInfo, pcData map[string]interface{}) pageContext {
	if page == nil {
		return pageContext{}
	}
	pc := pageContext{name: page.Name, slot: getStringVal(pcData, "slot_name")}
	if page.ID != uuid.Nil {
		id := page.ID
		pc.id = &id
	}
	return pc
}

// applyComponentSwap changes the component template for this section.
// Looks up the new component, then re-renders with existing content_data
// using full site context from DB.
//
// It does NOT write the row. It used to (bugs_open/136): the swap persisted
// itself here while content_edit persisted in the caller, which gave the action
// two persist sites and made any single pre-persist guard bypassable by the
// other branch. The caller now performs both writes, after one repair pass.
func applyComponentSwap(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pcData map[string]interface{},
	editCtx map[string]interface{},
	inputs *datahelpers.ActionInputs,
	stripMarkdown bool,
	logger *zap.Logger,
) (sectionEditOutcome, error) {

	newFunction := inputs.Get("new_component_function")
	if newFunction == "" {
		return sectionEditOutcome{}, fmt.Errorf("component_swap requires 'new_component_function' parameter")
	}

	// Normalize per naming contract
	newFunction = NormalizeComponentFunction(newFunction)

	// Look up the new component
	comp, err := GetComponentWithFallback(ctx, db, newFunction, logger)
	if err != nil {
		return sectionEditOutcome{}, fmt.Errorf("component %q not found: %w", newFunction, err)
	}

	if comp.HTMLTemplate == "" {
		return sectionEditOutcome{}, fmt.Errorf("component %q has no HTML template", newFunction)
	}

	// Determine content_data to render with.
	// Priority: replacement_content_data from caller > existing content_data from page_component
	contentData := make(map[string]interface{})

	if rcd := inputs.GetMap("replacement_content_data"); rcd != nil {
		// Caller provided replacement content (component_swap with new data)
		for k, v := range rcd {
			contentData[k] = v
		}
		logger.Info("applyComponentSwap: Using replacement_content_data from caller",
			zap.Int("field_count", len(rcd)))
	} else if cd, ok := pcData["content_data"].(map[string]interface{}); ok {
		// No replacement — use existing content_data from page_component
		for k, v := range cd {
			contentData[k] = v
		}
		logger.Info("applyComponentSwap: Using existing content_data from page_component",
			zap.Int("field_count", len(cd)))
	}

	// bugs_open/184: same gated strip as applyContentEdit, before the render.
	var strippedMarkdownFields []string
	if stripMarkdown {
		if changed := datahelpers.StripLiteralMarkdownFromContentData(contentData); len(changed) > 0 {
			strippedMarkdownFields = changed
			logger.Info("applyComponentSwap: stripped literal markdown",
				zap.Strings("fields", changed))
		}
	}

	// Build render context from DB with existing content_data
	pageData, _ := editCtx["page"].(map[string]interface{})
	var pageInfoForRender *PageInfo
	if pageData != nil {
		pageID, _ := uuid.Parse(getStringVal(pageData, "id"))
		pageInfoForRender = &PageInfo{
			ID:       pageID,
			Name:     getStringVal(pageData, "name"),
			Title:    getStringVal(pageData, "title"),
			Filename: getStringVal(pageData, "filename"),
			MetaDesc: getStringVal(pageData, "meta_desc"),
			Domain:   getStringVal(pageData, "domain"),
			SiteID:   siteID,
		}
	}

	renderCtx, err := buildRenderContextFromDB(ctx, db, siteID, pageInfoForRender, contentData, logger)
	if err != nil {
		return sectionEditOutcome{}, fmt.Errorf("failed to build render context for swap: %w", err)
	}

	// Same stored-row case as applyContentEdit above. The function is the NEW
	// component's: counting ITS same-function predecessors is what the page will
	// look like after the swap, which is the page this render belongs to.
	DeriveAndBindInstanceToken(ctx, db, renderCtx, comp.Function,
		PlacementFromStoredRow(pcData), logger)
	renderCtx.InputSchema = comp.InputSchema // bugs_open/342 — same live-page exposure as applyContentEdit
	rendered, _, _, err := RenderTemplate(comp.HTMLTemplate, renderCtx, logger)
	emitRequiredFieldsMissing(ctx, db, siteID,
		editPageContext(pageInfoForRender, pcData),
		nil, comp.Function,
		fmt.Sprintf("Section swap on %s", comp.Function),
		"page_component", "section_editor", renderCtx.AbsentRequiredFields,
		map[string]interface{}{"route": "component_swap", "live_page": true}, logger)
	if err != nil {
		// Same ungated live-page route as applyContentEdit above (bugs_open/260):
		// refuse the swap rather than write a section that did not render.
		return sectionEditOutcome{}, fmt.Errorf("template rendering failed after swap, live section left unchanged: %w", err)
	}
	if rendered == "" {
		return sectionEditOutcome{}, fmt.Errorf("template rendering produced empty output after swap")
	}

	logger.Info("applyComponentSwap: Swapped and re-rendered component",
		zap.String("old_slot", getStringVal(pcData, "slot_name")),
		zap.String("new_function", comp.Function),
		zap.String("new_component_id", comp.ID),
		zap.Int("output_length", len(rendered)),
	)

	// component_id, slot_name, rendered_html and content_data are written by the
	// caller (updatePageComponentSwap), after the link repair.
	compID, _ := uuid.Parse(comp.ID)

	return sectionEditOutcome{
		HTML:                   rendered,
		ContentData:            contentData,
		ComponentID:            compID,
		SlotName:               comp.Function,
		StrippedMarkdownFields: strippedMarkdownFields,
		AbsentRequiredFields:   renderCtx.AbsentRequiredFields,
	}, nil
}

// ===========================================================================
// DB HELPERS
// ===========================================================================

type pageComponentRow struct {
	ID           uuid.UUID
	PageID       uuid.UUID
	ComponentID  *uuid.UUID
	Position     int
	SlotName     string
	RenderedHTML string
	ContentData  map[string]interface{}
	BuildStatus  string
}

func loadPageComponentByID(ctx context.Context, db *sql.DB, id uuid.UUID) (pageComponentRow, error) {
	var row pageComponentRow
	var componentID sql.NullString
	var contentDataJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT id, page_id, component_id, position, slot_name,
		       COALESCE(rendered_html, ''), COALESCE(content_data, '{}'::jsonb),
		       COALESCE(build_status, 'pending')
		FROM page_components
		WHERE id = $1
	`, id).Scan(
		&row.ID, &row.PageID, &componentID, &row.Position, &row.SlotName,
		&row.RenderedHTML, &contentDataJSON, &row.BuildStatus,
	)
	if err != nil {
		return row, err
	}

	if componentID.Valid {
		parsed, _ := uuid.Parse(componentID.String)
		row.ComponentID = &parsed
	}

	if len(contentDataJSON) > 0 {
		json.Unmarshal(contentDataJSON, &row.ContentData)
	}

	return row, nil
}

// Problem: page_components rows sometimes have empty slot_name (e.g. placeholder rows,
// builds where metadata path fell through to HTML parsing with no data-component attribute).
// The current query only matches pc.slot_name = $3, so these rows are invisible to
// section-editor.
//
// Fix: Add two fallback match paths, preferring direct match:
//   1. pc.slot_name = $3                     (direct match — best)
//   2. cc.function = $3                       (component knows its function)
//   3. pages.sections[position-1] = $3        (plan-based position match)
//
// When a fallback match succeeds, the function also backfills the empty slot_name
// so subsequent queries use the direct path.
//
// ALSO: Returns the matched slot_name (COALESCE) so the caller always sees it.

func loadPageComponentBySlot(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, slotName string) (pageComponentRow, error) {
	row, err := loadPageComponentBySlotRO(ctx, db, siteID, pageName, slotName)
	if err != nil {
		return row, err
	}

	// Backfill: if we matched via fallback, update the slot_name in DB
	// so future queries use the direct path. Non-blocking — log and continue.
	if row.SlotName == slotName {
		// Check if the DB value was actually empty (we COALESCEd it)
		var dbSlotName sql.NullString
		db.QueryRowContext(ctx, `SELECT slot_name FROM page_components WHERE id = $1`, row.ID).Scan(&dbSlotName)
		if !dbSlotName.Valid || dbSlotName.String == "" || dbSlotName.String != slotName {
			_, _ = db.ExecContext(ctx, `UPDATE page_components SET slot_name = $2 WHERE id = $1`, row.ID, slotName)
		}
	}

	return row, nil
}

// loadPageComponentBySlotRO is loadPageComponentBySlot's matching logic with NO
// write side-effects — same three-way match (slot_name, component function,
// plan position), no slot_name backfill.
//
// Split out for read-only callers that must not mutate what they are inspecting:
// revalidate_review_queue sweeps the parked review queue in a dry_run mode whose
// whole contract is "report, change nothing", and the backfill above would break
// that. Keeping ONE copy of the match logic is the point — a second hand-rolled
// slot lookup is precisely the drift this repo keeps paying for (bugs_closed/041,
// section lookup never normalising).
func loadPageComponentBySlotRO(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, slotName string) (pageComponentRow, error) {
	var row pageComponentRow
	var componentID sql.NullString
	var contentDataJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT pc.id, pc.page_id, pc.component_id, pc.position,
		       COALESCE(NULLIF(pc.slot_name, ''), $3) as slot_name,
		       COALESCE(pc.rendered_html, ''),
		       COALESCE(pc.content_data, '{}'::jsonb),
		       COALESCE(pc.build_status, 'pending')
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		LEFT JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1 AND p.name = $2
		  AND (
		    pc.slot_name = $3
		    OR cc.function = $3
		    OR (
		      (pc.slot_name IS NULL OR pc.slot_name = '')
		      AND p.sections IS NOT NULL
		      AND pc.position > 0
		      AND pc.position <= jsonb_array_length(p.sections)
		      AND trim(both '"' from (p.sections->(pc.position - 1))::text) = $3
		    )
		  )
		ORDER BY
		  CASE WHEN pc.slot_name = $3 THEN 0
		       WHEN cc.function = $3 THEN 1
		       ELSE 2
		  END
		LIMIT 1
	`, siteID, pageName, slotName).Scan(
		&row.ID, &row.PageID, &componentID, &row.Position, &row.SlotName,
		&row.RenderedHTML, &contentDataJSON, &row.BuildStatus,
	)
	if err != nil {
		return row, err
	}

	if componentID.Valid {
		parsed, _ := uuid.Parse(componentID.String)
		row.ComponentID = &parsed
	}

	if len(contentDataJSON) > 0 {
		json.Unmarshal(contentDataJSON, &row.ContentData)
	}

	return row, nil
}

func loadComponentForEdit(ctx context.Context, db *sql.DB, componentID uuid.UUID) (map[string]interface{}, error) {
	var id, function, name string
	var htmlTemplate string
	var inputSchemaJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT id::text, function, name, html_template, COALESCE(input_schema, '{}'::jsonb)
		FROM content_components
		WHERE id = $1
	`, componentID).Scan(&id, &function, &name, &htmlTemplate, &inputSchemaJSON)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":            id,
		"function":      function,
		"name":          name,
		"html_template": htmlTemplate,
	}

	if len(inputSchemaJSON) > 0 {
		var schema interface{}
		if json.Unmarshal(inputSchemaJSON, &schema) == nil {
			result["input_schema"] = schema
		}
	}

	return result, nil
}

func loadComponentByFunction(ctx context.Context, db *sql.DB, function string) (map[string]interface{}, error) {
	var id, funcVal, name string
	var htmlTemplate string
	var inputSchemaJSON []byte

	err := db.QueryRowContext(ctx, `
		SELECT id::text, function, name, html_template, COALESCE(input_schema, '{}'::jsonb)
		FROM content_components
		WHERE function = $1 AND (is_active = true OR is_active IS NULL)
		LIMIT 1
	`, function).Scan(&id, &funcVal, &name, &htmlTemplate, &inputSchemaJSON)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":            id,
		"function":      funcVal,
		"name":          name,
		"html_template": htmlTemplate,
	}

	if len(inputSchemaJSON) > 0 {
		var schema interface{}
		if json.Unmarshal(inputSchemaJSON, &schema) == nil {
			result["input_schema"] = schema
		}
	}

	return result, nil
}

// errComponentLocked is returned by the page_components UPDATE helpers when
// the row exists but carries an active human lock (bugs_open/058) — or, since
// bugs_open/360, when it is a removed tombstone the race window let through
// (the advisory tombstone gate in ApplySectionEditAction catches the ordinary
// case with its own skip-result; this sentinel is the race-free backstop).
// The predicates live in the UPDATE's WHERE clause, so the refusal is
// race-free; callers convert this to a skip-result rather than a failure.
var errComponentLocked = errors.New("page component is locked, removed, or missing — automated edit refused (bugs_open/058, bugs_open/360)")

// pageComponentNotRemovedSQL is the tombstone predicate (bugs_open/360).
// build_status='removed' is the documented assembly-excluded tombstone
// (rerender_single_page_action.go:843): the row is NOT on the page, and the
// helpers below all promote build_status to 'approved' — so without this
// predicate an automated edit UN-RETIRES the slot and the next rerender
// publishes the retired content again. Measured 2026-08-21: four
// literal_markdown transform edits resurrected four retired ported slots and
// the pages publicly served two stacked tools for ~19 h.
const pageComponentNotRemovedSQL = "COALESCE(build_status, 'pending') <> 'removed'"

func updatePageComponentAfterEdit(ctx context.Context, db *sql.DB, pcID uuid.UUID, html string, contentData map[string]interface{}) error {
	var contentDataJSON []byte
	var err error
	var res sql.Result

	if contentData != nil {
		contentDataJSON, err = json.Marshal(contentData)
		if err != nil {
			return fmt.Errorf("failed to marshal content_data: %w", err)
		}
	}

	// rendered_html_digest = md5($2) same-statement (bugs_open/229): the
	// section editor renders the edited content, so its output is
	// machine-made. The 357 trigger archives whatever these statements
	// replace.
	// stampedExecContext (bugs_open/355 A1): the 357 trigger archives whatever
	// these statements replace, and the stamp is what lets that archive row
	// name THIS writer instead of the connection's socket.
	if contentDataJSON != nil {
		res, err = stampedExecContext(ctx, db, contentWriterSectionEditorUpdate, `
			UPDATE page_components
			SET rendered_html = $2,
			    rendered_html_digest = md5($2),
			    content_data = $3::jsonb,
			    build_status = 'approved',
			    updated_at = NOW()
			WHERE id = $1 AND `+pageComponentNotRemovedSQL+` AND `+pageComponentAgentWritableSQL("")+`
		`, pcID, html, string(contentDataJSON))
	} else {
		res, err = stampedExecContext(ctx, db, contentWriterSectionEditorUpdate, `
			UPDATE page_components
			SET rendered_html = $2,
			    rendered_html_digest = md5($2),
			    build_status = 'approved',
			    updated_at = NOW()
			WHERE id = $1 AND `+pageComponentNotRemovedSQL+` AND `+pageComponentAgentWritableSQL("")+`
		`, pcID, html)
	}

	if err != nil {
		return err
	}
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		return errComponentLocked
	}
	return nil
}

func updatePageComponentSwap(ctx context.Context, db *sql.DB, pcID, componentID uuid.UUID, newSlotName, html string, contentData map[string]interface{}) error {
	contentDataJSON, err := json.Marshal(contentData)
	if err != nil {
		return fmt.Errorf("failed to marshal content_data: %w", err)
	}

	// stampedExecContext (bugs_open/355 A1): see updatePageComponent above.
	res, err := stampedExecContext(ctx, db, contentWriterSectionEditorSwap, `
		UPDATE page_components
		SET component_id = $2,
		    slot_name = $3,
		    rendered_html = $4,
		    content_data = $5::jsonb,
		    build_status = 'approved',
		    updated_at = NOW()
		WHERE id = $1 AND `+pageComponentNotRemovedSQL+` AND `+pageComponentAgentWritableSQL("")+`
	`, pcID, componentID, newSlotName, html, string(contentDataJSON))

	if err != nil {
		return err
	}
	if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
		return errComponentLocked
	}
	return nil
}

// ===========================================================================
// UTILITY HELPERS
// ===========================================================================

func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}

func getStringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
