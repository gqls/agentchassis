package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The SINGLE-ROW form of both per-slot floors, for writers that update ONE
// page_components row directly instead of composing a whole page.
//
// WHY THIS EXISTS (council round 1 on b30ac52c, gating objection from
// bug_historian, severity high). Both floors — the text shrink floor
// (bugs_open/178) and the component floor (bugs_open/253) — were wired only into
// SavePageSectionsAction. Audited 2026-08-13: NINE Go call sites write
// page_components.rendered_html and exactly one was guarded. The seat's words:
// "if any of those paths bypass SavePageSectionsAction, they bypass BOTH the
// pre-existing text shrink floor and this new component floor, and a flattening
// save through one of them will fail exactly as silently as the bug this plan
// fixes."
//
// It was right, and worse than stated. The unguarded writer that matters is
// ApplySectionEditAction: it is LIVE (the `section-editor` agent definition), it
// UPDATEs rendered_html directly, and it is PRECISELY the per-component edit path
// decomposition exists to enable — HANDOFF_2026-08-10c §3's stated benefit is
// "after decomposition you can rewrite one prose block without touching the
// calculator", and that is this action. So the guards covered the door the
// observed incident came through and missed the one the design steers edits
// toward.
//
// ONE FUNCTION, NOT A COPY, on purpose. The defect being fixed is literally "one
// call site of a shared judgement gets the rigorous fix; the sibling stays
// heuristic" (016b §9). Fixing it by pasting the floor logic into a second call
// site would reproduce the defect with an extra copy to drift. Both decisions
// stay in their own files as pure functions (evaluateSectionShrink,
// evaluateComponentLoss); this composes them for the single-row case, so a THIRD
// writer adopts both floors in one call.
//
// SCOPE — which writers should call this, decided per writer rather than wiring
// all nine reflexively (the audit is in bugs_open/253):
//
//	CALLS IT   ApplySectionEditAction, editType "content_edit" — replaces an
//	           existing slot's HTML with rewritten content. The exact shape.
//	DOES NOT   ApplySectionEditAction, editType "component_swap" — deliberately
//	           changes component_id AND slot_name AND html: the markup is SUPPOSED
//	           to change because the component is a different one. A floor here
//	           would refuse the operation for doing its job.
//	N/A        create_tool_component, deploy_tool — INSERT only, ZERO
//	           `UPDATE ... SET rendered_html` sites (measured 2026-08-13). Not
//	           rewriters at all, so out of scope by CONSTRUCTION, not by
//	           exemption — they never appear before the coverage test.
//	EXEMPT     adopt_verbatim — writes the ORIGINAL adopted document, i.e. it
//	           creates the prior that everything else is measured against.
//	EXEMPT     create_report_page — ⚠ CORRECTED: an earlier draft of this comment
//	           filed it as create-only. It is NOT: it looks up its report row by
//	           (page_id, slot_name) and OVERWRITES it. The coverage test caught
//	           that, the manual audit had not. Exempt for a different reason —
//	           the row is a machine-rendered dossier this action owns and
//	           regenerates wholesale, never a decomposed prose block.
//	EXEMPT     rebuild_blog_listing — machine-generated listing, same reasoning.
//	EXEMPT     fix_forced_text_colours, fix_harcoded_colours — colour rewrites
//	           inside <style> blocks/declarations. `[MEASURED 2026-08-14]`:
//	           colour_fixer_class_preservation_test.go drives both rendered-row
//	           transforms (processComponentCSS, checks.ReplaceHardcodedColors)
//	           over class-carrying fixtures and asserts the floor's own census
//	           (countComponentClasses) unchanged AND the class-attribute list
//	           identical — with controls that each transform actually fired and
//	           that the census detects a removed class. The regex reading that
//	           stood here (style-block scope, never element attributes) is now
//	           the explanation, not the evidence.
//
// TWO KINDS OF EVIDENCE HERE, AND THEY ARE NOT EQUAL (council round 3, advisory
// objection from editquality — my own submission claimed all nine dispositions
// were "measured rather than asserted" while two were marked [UNMEASURED], which
// overstated exactly the evidence the reader most needs to weigh):
//
//	MEASURED   create_tool_component, deploy_tool — counted: 1 INSERT each, ZERO
//	           `UPDATE ... SET rendered_html`. A count that could have come out
//	           otherwise.
//	MEASURED   create_report_page — the coverage TEST found it, contradicting the
//	           manual audit. Disconfirmation, which is the strongest kind here.
//	REASONED   adopt_verbatim, rebuild_blog_listing — read from source. Sound,
//	           and still a different thing.
//	MEASURED   the two colour fixers — converted 2026-08-14 by exactly the
//	           experiment this note used to prescribe: transforms over a
//	           class-carrying fixture, census unchanged, with transform-fired
//	           and instrument-detects controls
//	           (colour_fixer_class_preservation_test.go).
//
// Every one of the nine writers now has a disposition. The exemptions live in
// page_component_writer_coverage_test.go's exemptWriters, where a reason is
// mandatory and a stale entry fails a second test.
func enforceSingleSlotFloors(ctx context.Context, params ActionParams, siteID, pageID uuid.UUID,
	pageName, slot, existingHTML, incomingHTML string) error {

	// Same config keys as the whole-page path, so a step that has already opted
	// out of a floor is not surprised by it reappearing on a different writer.
	componentFloor, _ := pruneFloorFromConfig(params.StepConfig.Config, sectionComponentFloorKey, defaultSectionComponentFloor)
	textFloor, _ := pruneFloorFromConfig(params.StepConfig.Config, sectionShrinkFloorKey, defaultSectionShrinkFloor)

	// An empty existing row is a first write, not a shrink — nothing to compare.
	if strings.TrimSpace(existingHTML) == "" {
		return nil
	}

	// CALLER CENSUS (council 3279156b, guardian seat: "if any other write path
	// calls this function it inherits the loosened axis with zero calibration
	// coverage"). Measured 2026-08-17 over platform/ internal/ pkg/ cmd/, non-test:
	// enforceSingleSlotFloors has exactly ONE caller, section_editor_actions.go:436
	// (the content_edit branch) — the path the 117 archived pairs came from. The
	// shared pure decision evaluateSectionShrink has two callers, this one and
	// save_sections_shrink_guard.go:165, and that one still passes tag-stripped
	// lengths: the axis lives in the CALLER, so the whole-page path is unaffected
	// by construction, not by convention. A second caller of THIS function inherits
	// the visible axis with no calibration of its own — re-run the harness first.
	//
	// The text axis is VISIBLE text — style/script/comment blocks removed with
	// their content — not the tag-stripped length the whole-page path uses.
	// bugs_closed/285: on the write that emptied a live article body the
	// tag-stripped axis read 262% retained (a wrapper stylesheet replaced the
	// prose) while visible text fell 2,143 → 16 chars. Calibrated against all
	// 117 archived overwrite pairs, including what the change LOOSENS — the
	// table, the disagreement in both directions, and the scope limit are in
	// section_visible_text.go's header. RE-RUN THAT CALIBRATION before changing it.
	existingText := map[string]int{slot: visibleTextLength(existingHTML)}
	incomingText := map[string]int{slot: visibleTextLength(incomingHTML)}
	existingComp := map[string]int{slot: countComponentClasses(existingHTML)}
	incomingComp := map[string]int{slot: countComponentClasses(incomingHTML)}

	var reasons []string
	// evaluateSectionShrink applies its own minShrinkGuardChars (500) to the
	// EXISTING side, so a short caption is out of scope without a second rule
	// here. An earlier version of this change added its own 120-char minimum;
	// MUT-2 proved it dead code (removing it broke nothing, because 500
	// dominates) and it was deleted rather than kept as decoration.
	for _, v := range evaluateSectionShrink(textFloor, existingText, incomingText) {
		reasons = append(reasons, fmt.Sprintf("%s %d→%d chars of VISIBLE text, stylesheet and script content excluded (%.0f%% kept, floor %.0f%%)",
			v.Slot, v.Existing, v.Incoming, v.ratio()*100, textFloor*100))
	}
	for _, v := range evaluateComponentLoss(componentFloor, existingComp, incomingComp) {
		reasons = append(reasons, fmt.Sprintf("%s %d→%d class attributes (%.0f%% kept, floor %.0f%%)",
			v.Slot, v.Existing, v.Incoming, v.ratio()*100, componentFloor*100))
	}
	if len(reasons) == 0 {
		return nil
	}

	reason := fmt.Sprintf(
		"apply_section_edit: SLOT FLOOR REFUSED for page %q slot %q — %s. Nothing was written and the existing component still stands. A single-component edit is held to the same floors as a whole-page save (bugs_open/178 text, bugs_open/253 components); set %s or %s on this step to change them (0 disables).",
		pageName, slot, strings.Join(reasons, "; "), sectionShrinkFloorKey, sectionComponentFloorKey)

	params.Logger.Warn("ApplySectionEditAction: SLOT FLOOR BLOCKED",
		zap.String("page_name", pageName),
		zap.String("slot_name", slot),
		zap.Int("violations", len(reasons)))
	_ = emitPruneRefusalWorkItem(ctx, params.DB, savePageSectionsRefusal(siteID, pageID, pageName, reason,
		fmt.Sprintf("Section edit refused: %q would lose too much of its text or layout", slot),
		singleSlotFloorFix), params.Logger)
	return fmt.Errorf("%s", reason)
}

// singleSlotFloorFix is this path's OWN aftermath sentence. The whole-page
// guards' remedies name save_page_sections and its step config, which is the
// wrong step to go and edit when the refusal came from the section editor.
const singleSlotFloorFix = "A single-component edit would have removed more of a section's text or its layout " +
	"components (cards, grids, buttons) than the floors allow, so NOTHING was written and the existing " +
	"component still stands. This is the section editor, not a page rebuild — check the edit instruction " +
	"that was given: an instruction to 'simplify' or 'rewrite' a slot that carries layout markup will " +
	"produce clean prose and lose the structure, because the writer is not told what the markup means. " +
	"Giving it the component vocabulary in content_direction is what fixed the motivating page " +
	"(bugs_open/253); lowering section_shrink_floor / section_component_floor on this step is the " +
	"deliberate escape hatch when the simplification is genuinely intended."
