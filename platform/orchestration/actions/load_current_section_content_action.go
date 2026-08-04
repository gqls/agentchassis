// FILE: platform/orchestration/actions/load_current_section_content_action.go
//
// bugs_open/178. content_rewrite items never set spec.mode, so
// load_existing_content's adoption-only gate no-ops for them
// (load_existing_content_action.go:64-69 — mode != "recreate" → no-op), and
// even when mode IS "recreate" that step sources research_results (the
// original adoption-crawl snapshot), never a page's CURRENT page_components.
// There is therefore no channel that hands a page's own live stored prose
// back to its writer for editing — page-content-writer gets the item's
// guidance text and nothing to work from, so it fabricates a full
// replacement section that satisfies the instruction's shape while dropping
// most of what was there (measured: 4439 → 1806 chars on one page, one
// paragraph in three preserved).
//
// This is the channel. It is a THIRD value on the existing spec.mode field
// (2026-08-02 owner ruling: new authority on a shared seam ships as an
// opt-in field with the unsafe default OFF) — "edit_live" — alongside the
// existing "recreate". Every item that leaves mode unset or sets it to
// anything else gets the section_plan it was handed, byte-for-byte
// unchanged: this step is a structural no-op for every caller that does not
// name it, which is what makes it safe to insert into the one shared
// page-build-handler workflow every content build already runs.
//
// Read-only, and deliberately keyed the same way the WRITE side is: slot_name
// is what save_page_sections_action.go stores component_function under, and
// plan_sections' sectionPlanItem.Name is that same component_function/
// section_type — see plan_sections_action.go's loadComponentSchemas comment
// ("callers look it up by the REQUESTED name ... rerender's slot_name") and
// aliasNormalisedSectionKeys. Matching on that name is what lets each
// section in the loop find its OWN prior content rather than the whole
// page's, which matters because content_rewrite can touch every ready
// section on a page in one run (bugs_open/178: three slots each rewritten
// from the same guidance).
//
// Registration:
//   "load_current_section_content": {
//       Handler:     LoadCurrentSectionContentAction,
//       Category:    "site",
//       Description: "When spec.mode=edit_live, attach each ready section's current rendered_html so the writer edits instead of regenerating",
//       IsLocal:     true,
//   }
//
// Workflow position: after plan_sections/check_has_ready_sections, before
// spawn_content_writer — section_plan must already exist. Reuses the SAME
// output_field ("section_plan") plan_sections itself uses, deliberately: it
// means call_content_writer's existing input_mapping needs no change at all,
// and a reader diffing the workflow sees one step transparently refining the
// value the previous step produced, not two competing sources of truth.
//
// THE RETURN VALUE IS THE PLAN — never a wrapper around it (bugs_open/192).
// Reusing an upstream key as output_field is a REPLACEMENT, not an annotation:
// coordinator.go's storeActionResult stores an action's return value wholesale
// under output_field, so whatever this function returns simply BECOMES
// collected_data.section_plan. The first version of this file returned
// {"section_plan": plan, "applied": …, "reason"|"matched": …} on every path,
// including all eight it called "pass-through". That silently demoted the real
// plan one level on EVERY page build in EVERY mode, and the wrong result looked
// exactly like the right one because all the data was still there — just at
// section_plan.section_plan. Downstream, page-content-writer's select_sections
// (input_data.section_plan.sections_ready) and resolve_links
// ("sections?": the same path) both resolved nothing, and every page build in
// the fleet failed at process_sections_loop with an error naming the missing
// key rather than the failed extraction.
//
// So: a step whose output_field names its own input must be SHAPE-PRESERVING.
// Bookkeeping goes to the log, and — only on the path that actually did
// something — to one namespaced key INSIDE the plan (edit_live_meta), which is
// additive: every consumer reads named fields off the plan, none iterates its
// keys (grep-verified, 2026-08-04).

package actions

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var LoadCurrentSectionContentInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"page_id", "mode"},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_current_section_content", LoadCurrentSectionContentInputSpec)
}

// editLiveMode is the spec.mode value that opts a content_rewrite item into
// this channel. Any other value (including unset, and "recreate") leaves
// this action a pass-through.
const editLiveMode = "edit_live"

func LoadCurrentSectionContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "load_current_section_content"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// section_plan is read directly rather than through the ActionInputSpec:
	// it is always the fixed collected-data key the immediately preceding
	// plan_sections step wrote, never a caller-configurable path, so there is
	// nothing for a config key to name.
	sectionPlanRaw := datahelpers.ExtractNestedField(params.CollectedData, "section_plan")

	// output_field is "section_plan", so the return value REPLACES the plan
	// (storeActionResult). A pass-through therefore returns the plan value
	// ITSELF, byte-identical — which is what this file's header has always
	// promised and what, before bugs_open/192, it did not do. The reason is
	// logged rather than returned: there is no second output channel on a step,
	// and inventing one by wrapping is precisely the defect.
	passthrough := func(reason string) (interface{}, error) {
		logger.Info("LoadCurrentSectionContent: pass-through, section_plan returned unchanged",
			zap.String("reason", reason))
		return sectionPlanRaw, nil
	}

	if params.DB == nil {
		return passthrough("no_db")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		LoadCurrentSectionContentInputSpec, logger,
	)
	if err != nil {
		logger.Info("LoadCurrentSectionContent: input extraction failed, section_plan passed through unchanged",
			zap.Error(err))
		return passthrough("no_inputs")
	}

	if inputs.Get("mode") != editLiveMode {
		return passthrough("not_edit_live")
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		logger.Info("LoadCurrentSectionContent: invalid site_id, section_plan passed through unchanged",
			zap.String("site_id", siteIDStr))
		return passthrough("invalid_site_id")
	}

	pageIDStr := inputs.Get("page_id")
	pageID, err := uuid.Parse(pageIDStr)
	if err != nil {
		logger.Info("LoadCurrentSectionContent: invalid page_id, section_plan passed through unchanged",
			zap.String("page_id", pageIDStr))
		return passthrough("invalid_page_id")
	}

	if sectionPlanRaw == nil {
		return passthrough("no_section_plan")
	}

	// Normalise to a generic map by round-tripping through JSON. In-process
	// this value may still be the Go-typed []sectionPlanItem plan_sections
	// produced; after a restart it decodes from persisted collected_data as
	// []interface{} of map[string]interface{}. The round trip handles both
	// shapes the same way instead of type-switching on which one this run
	// has, and is safe here because this value is about to cross an agent
	// boundary to page-content-writer via call_agent, which JSON-marshals it
	// regardless.
	var planMap map[string]interface{}
	buf, err := json.Marshal(sectionPlanRaw)
	if err != nil {
		logger.Warn("LoadCurrentSectionContent: section_plan not marshalable, passed through unchanged", zap.Error(err))
		return passthrough("unmarshalable_section_plan")
	}
	if err := json.Unmarshal(buf, &planMap); err != nil {
		logger.Warn("LoadCurrentSectionContent: section_plan not a JSON object, passed through unchanged", zap.Error(err))
		return passthrough("malformed_section_plan")
	}

	readySections, _ := planMap["sections_ready"].([]interface{})
	if len(readySections) == 0 {
		return passthrough("no_ready_sections")
	}

	// Read-only. Joined through pages so a page_id that (by whatever bug)
	// belongs to a different site cannot leak that site's content into this
	// one's build.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT pc.slot_name, COALESCE(pc.rendered_html, '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE pc.page_id = $1 AND p.site_id = $2
	`, pageID, siteID)
	if err != nil {
		logger.Warn("LoadCurrentSectionContent: page_components query failed, section_plan passed through unchanged",
			zap.Error(err))
		return passthrough("query_failed")
	}
	defer rows.Close()

	existingBySlot := make(map[string]string)
	for rows.Next() {
		var slotName, html string
		if scanErr := rows.Scan(&slotName, &html); scanErr != nil {
			continue
		}
		if slotName != "" && html != "" {
			existingBySlot[slotName] = html
		}
	}
	if err := rows.Err(); err != nil {
		logger.Warn("LoadCurrentSectionContent: page_components rows error", zap.Error(err))
	}

	matched := 0
	claimedSlots := make(map[string]bool, len(readySections))
	var unmatchedSections []map[string]interface{}
	for _, raw := range readySections {
		section, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := section["name"].(string)
		if html, found := existingBySlot[name]; found {
			section["existing_content_html"] = html
			claimedSlots[name] = true
			matched++
			continue
		}
		unmatchedSections = append(unmatchedSections, section)
	}

	// Fallback for bugs_open/178's second, distinct gap: the exact-name join
	// above only finds content when THIS build's resolved section name agrees
	// with what a PRIOR build stored it under. It does not always — a page's
	// current plan can name a position differently from the component that
	// build-time selection actually attached last time (measured 2026-08-04:
	// 3/127 pages with a current plan already carry a stored slot absent from
	// that plan's own name list). When that happens the join legitimately
	// finds nothing, and without this fallback the writer fabricates a fresh
	// section exactly as it did before this action existed — the original bug,
	// reached by a route the exact match does not cover.
	//
	// Only acted on when it is UNAMBIGUOUS: exactly one ready section missed
	// its exact match, and exactly one still-unclaimed page_components row on
	// this page is prose-sized (reuses the shrink guard's own definition of
	// "prose-sized", minShrinkGuardChars stripped chars — a hero/param-sized
	// slot is not a plausible body-text match). Two or more of either side and
	// there is no principled way to pair them, so both stay unmatched — never
	// worse than this step's behaviour before the fallback existed.
	fallbackMatched := 0
	if len(unmatchedSections) == 1 {
		var candidateSlot string
		candidateCount := 0
		for slot, html := range existingBySlot {
			if claimedSlots[slot] {
				continue
			}
			stripped := strings.TrimSpace(shrinkGuardTagStripper.ReplaceAllString(html, ""))
			if len(stripped) < minShrinkGuardChars {
				continue
			}
			candidateSlot = slot
			candidateCount++
		}
		if candidateCount == 1 {
			unmatchedSections[0]["existing_content_html"] = existingBySlot[candidateSlot]
			fallbackMatched = 1
			logger.Info("LoadCurrentSectionContent: attached via single-unmatched-prose-slot fallback",
				zap.Any("section_name", unmatchedSections[0]["name"]),
				zap.String("stored_slot", candidateSlot))
		}
	}

	planMap["sections_ready"] = readySections

	logger.Info("LoadCurrentSectionContent: attached current content for edit mode",
		zap.String("site_id", siteIDStr),
		zap.String("page_id", pageIDStr),
		zap.Int("sections_ready", len(readySections)),
		zap.Int("matched", matched),
		zap.Int("fallback_matched", fallbackMatched))

	// The one DB-visible record that this channel fired, namespaced INSIDE the
	// plan rather than wrapped around it, and written only on the path that
	// actually attached content. Additive: consumers read named fields off the
	// plan, none iterates its keys. fallback_matched is kept separate from
	// matched rather than folded in — bugs_open/178's own open item is that the
	// component-identity-drift RATE is unmeasured fleet-wide; a distinct
	// counter is that measurement, for free, on every future edit_live build.
	planMap["edit_live_meta"] = map[string]interface{}{
		"applied":          true,
		"matched":          matched,
		"fallback_matched": fallbackMatched,
	}

	return planMap, nil
}
