package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Per-slot shrink guard (bugs_open/178).
//
// The page-total content regression guard in SavePageSectionsAction refuses a
// save only when the WHOLE page's text falls below 25% of what is deployed. A
// regeneration that rewrites one prose section from scratch and drops most of
// it passes that guard twice over: the loss is diluted by healthy sibling
// slots, and 25% is a wipe threshold, not a shrink threshold. Measured
// 2026-08-02: a "add a tool reference" item cut one slot's prose by 57% (page
// total −24%) and reported complete; the same shape had already cost
// vetcomparison 70% of its FAQ and fundamentallyai a 32KB tool source on
// earlier days. Every structural check passed because the HTML was well-formed
// and the siblings were intact.
//
// This guard compares SLOT AGAINST SAME-NAMED SLOT and refuses the save when a
// prose-sized slot shrinks past the floor. Slots that are new, dropped, or small
// stay out of scope on purpose:
//   - a slot absent from the incoming set is a DROP — the completeness floor's
//     domain (save_sections_prune_floor.go), not a shrink;
//   - a slot below the governing minimum of existing text is parameter/
//     hero-sized, where large relative shrinks are routinely legitimate
//     normalisation (measured: gamesdesign's hero rewrite, −70%, was an
//     improvement);
//   - growth is always allowed.
//
// The floor is config-tunable per step (sectionShrinkFloorKey, default
// defaultSectionShrinkFloor, clamped to [0,0.95]; 0 disables), so an item that
// legitimately intends a large cut can declare that intent in config instead
// of being narrated around. Refusal follows this action's existing semantics:
// the whole save is refused and nothing is written.
//
// ── THE AXIS: VISIBLE TEXT, corrected 2026-08-17 (bugs_open/293) ─────────────
// Until this change both sides were measured with `<[^>]*>`-stripped length,
// which strips TAGS but not what is INSIDE <style> and <script> — so CSS
// declarations and JavaScript source counted as "text". The failure mode that
// makes invisible is the one this guard exists to stop: replace an article with
// a wrapper stylesheet and the count GOES UP.
//
// Measured on 1,079 exactly-paired whole-page rebuild writes (the join, its
// disconfirming control and its positive control are in
// docs024_key_docs_latest/bugfix_293_whole_page_shrink_axis/NOTES; the harness is
// shrink_axis_calibration_test.go and it is re-runnable):
//
//   - the retired axis judged 1,060 sections and ALLOWED a total prose wipe on
//     724 of them (68%) — constructed on the real sections, since no rebuild in
//     the archived window happened to hollow a page, so "it would have caught X"
//     was not available on this path and the case is made on mechanism. A second,
//     independent instrument agrees: 728 of 1,062 sections have over half of
//     their tag-stripped "text" inside style/script, 85–89% on some;
//   - the visible axis refuses 0 of the 1,079 writes that really happened. Across
//     the wider weak-join population it would refuse exactly ONE write the guard
//     would actually have judged — robot-hands.com/about/differentiators,
//     3,724→1,554 visible chars, read by hand and a legitimate tightening
//     rewrite. That is ~4 a month fleet-wide at any minimum, on an 8-day window,
//     against 724 wipes the retired axis waves through. sectionShrinkFloorKey is
//     the escape hatch for it and is live-immediate.
//
// THE MINIMUM IS A PARAMETER NOW, AND ITS VALUE CHANGED WITH THE AXIS.
// minShrinkGuardChars (500) was calibrated against CSS-inflated lengths; the same
// 500 applied to visible text puts 587 of 1,079 slots out of scope. Swept over
// both populations the guard-judged refusal count stays at exactly ONE at every
// step from 500 down to 50, so the minimum is not what protects against false
// refusals here — see minShrinkGuardVisibleChars.

const (
	// sectionShrinkFloorKey is the step-config key holding the minimum
	// fraction of a slot's existing stripped text that a replacement must
	// retain. 0 disables the guard for that step.
	sectionShrinkFloorKey = "section_shrink_floor"

	// defaultSectionShrinkFloor: a replacement keeping less than half of a
	// prose slot's text is refused unless the step opts out.
	defaultSectionShrinkFloor = 0.5

	// minShrinkGuardChars: the TAG-STRIPPED "prose-sized" threshold. No longer
	// governs either floor (see minShrinkGuardVisibleChars) and is NOT dead: its
	// remaining consumer is load_current_section_content_action.go:262, which uses
	// it with shrinkGuardTagStripper to decide whether an unclaimed stored slot is
	// a plausible body-text match for an unpaired section. That is a PAIRING
	// HEURISTIC, not a content floor — nothing is refused on it — so it keeps the
	// old axis and the old number deliberately: re-tuning it would change which
	// slots get paired, with no calibration for that decision. Do not "unify" the
	// two constants; they answer different questions.
	minShrinkGuardChars = 500

	// minShrinkGuardVisibleChars: slots whose existing VISIBLE text is below this
	// are out of scope for both text floors — param blobs, headlines and captions
	// shrink legitimately and a ratio on 40 characters is noise.
	//
	// 200, and why not lower (bugs_open/293, 2026-08-17). Scope rises 492 → 959 of
	// 1,079 rebuild pairs going from 500 to 200, and 1,046 at 120, with the
	// guard-judged refusal count pinned at ONE the whole way — so on the rebuild
	// population 120 is free. 200 is chosen because it is the deepest step the
	// SECTION EDITOR's own population also covers (263 overwrite pairs: scope
	// 153 → 204 at 200, the same 4 refusals, all real hollowings). Both call sites
	// share this constant, and moving the editor onto a minimum with no
	// editor-side evidence would be the exact mistake 293 was filed to avoid, one
	// parameter over. The 200 → 120 step buys 87 slots; it is not worth the
	// evidence gap. Lowering later is this constant plus a re-run of the harness.
	minShrinkGuardVisibleChars = 200

	// pageTotalTextFloorKey / defaultPageTotalTextFloor: the PAGE-TOTAL floor's
	// escape hatch. It had none before 2026-08-17 — the rule was inlined in
	// SavePageSectionsAction with hardcoded thresholds — so a page that
	// legitimately shed most of its prose could only be saved by rolling a binary.
	pageTotalTextFloorKey    = "page_total_text_floor"
	defaultPageTotalTextFloor = 0.25

	// minPageTotalTextChars: below this a page has too little text for a ratio to
	// mean anything. Carried over unchanged from the inline rule (200).
	minPageTotalTextChars = 200
)

// shrinkGuardTagStripper is the RETIRED axis. Kept because
// load_current_section_content_action.go still pairs with it (see
// minShrinkGuardChars above) and because shrink_axis_calibration_test.go pins the
// harness's comparator against strippedIncomingBySlot. Neither floor measures
// with it any more — if you are reaching for it to guard content, you want
// visibleTextLength.
var shrinkGuardTagStripper = regexp.MustCompile(`<[^>]*>`)

// slotShrink describes one same-named slot whose replacement fell below the
// floor. Existing/Incoming are stripped-text lengths, not HTML lengths.
type slotShrink struct {
	Slot     string
	Existing int
	Incoming int
}

func (s slotShrink) ratio() float64 {
	if s.Existing == 0 {
		return 1
	}
	return float64(s.Incoming) / float64(s.Existing)
}

// evaluateSectionShrink is the pure decision: given text lengths by slot on both
// sides, return the slots that violate the floor. Split from the SQL so the rule
// is testable without a database.
//
// IT STAYS LENGTH-BASED, AND THE AXIS STAYS THE CALLER'S TO SUPPLY — deliberately,
// against the obvious alternative of having it take HTML and measure. Two reasons.
// The calibration harness's whole value is that it can run BOTH axes through this
// exact decision on real pairs (shrink_axis_calibration_test.go); a decision that
// measures for itself cannot be calibrated, only trusted. And what stops a fifth
// caller inventing a fifth axis is not this signature — it is
// shrink_axis_coverage_test.go, which fails the build when a caller of this
// function does not measure with visibleTextLength. A type cannot express
// "measured the right way"; a test can check it.
//
// minExistingChars is a parameter for the same reason it exists at all: the number
// depends on the axis, and both live callers pass minShrinkGuardVisibleChars.
func evaluateSectionShrink(floor float64, minExistingChars int, existing map[string]int, incoming map[string]int) []slotShrink {
	if floor <= 0 {
		return nil
	}
	if floor > 0.95 {
		floor = 0.95
	}
	var violations []slotShrink
	for slot, existingLen := range existing {
		if existingLen < minExistingChars {
			continue
		}
		incomingLen, present := incoming[slot]
		if !present {
			// A drop, not a shrink — the completeness floor's case.
			continue
		}
		if float64(incomingLen) < float64(existingLen)*floor {
			violations = append(violations, slotShrink{Slot: slot, Existing: existingLen, Incoming: incomingLen})
		}
	}
	return violations
}

// strippedIncomingBySlot maps the incoming sections to RETIRED-axis lengths.
// No longer used by either floor; retained as the calibration harness's
// comparator, which pins its own tag-stripped measure against this function so it
// cannot drift into measuring a third axis nobody ships.
func strippedIncomingBySlot(sections []SectionData) map[string]int {
	return incomingBySlot(sections, func(html string) int {
		return len(strings.TrimSpace(shrinkGuardTagStripper.ReplaceAllString(html, "")))
	})
}

// incomingBySlot folds the incoming sections into per-slot totals under whatever
// measure it is given — ONE aggregation, three measures.
//
// Council 823679dc, reuse seat: three functions in this package had grown the same
// shape (tag-stripped, visible text, class attributes), each with its own copy of
// the keying rule, which is how the last-write-wins defect came to exist in two of
// them independently. The keying question ("what does a repeated slot name mean?")
// is answered once, here, and the measure is the parameter — which is also the
// separation the calibration harness needs.
//
// ComponentName is what the insert loop writes as slot_name; += because a slot name
// legitimately repeats and the insert loop writes every instance (see
// visibleIncomingBySlot's note).
func incomingBySlot(sections []SectionData, measure func(string) int) map[string]int {
	m := make(map[string]int, len(sections))
	for _, s := range sections {
		m[s.ComponentName] += measure(s.HTML)
	}
	return m
}

// visibleIncomingBySlot maps the incoming sections to VISIBLE-text lengths keyed
// by slot name, SUMMING same-named sections rather than letting the last one win.
//
// Summing, not overwriting (bugs_open/293). A slot name legitimately repeats on a
// page — 14 pages do it, `generic-text-block` 2–3× with differing content, and
// LANDMINES.md records that 11 of 17 such groups are legitimate, so making the
// name unique is not available. The insert loop writes EVERY instance it is given
// (position = i+1 for each), so last-write-wins undercounts what the save will
// actually write, and which instance won depended on slice order. Summing makes
// the unit of judgement the slot-name group's total: identical arithmetic for the
// ~97% of groups with one instance, deterministic for the rest, and it needs no
// assumption that position survives a rebuild. One stated consequence: instances
// individually under the minimum can aggregate over it (three 80-char blocks =
// 240), which brings the group into scope — correct, since the group does carry
// prose.
func visibleIncomingBySlot(sections []SectionData) map[string]int {
	return incomingBySlot(sections, visibleTextLength)
}

// enforceSectionShrinkFloor loads the existing per-slot stripped lengths and
// refuses the save when any same-named prose slot shrinks past the floor. A
// refusal writes nothing and emits a refusal work item so the event is visible
// in the queue rather than only in a pod log (the 178 failures were invisible
// precisely because every surface reported success).
func enforceSectionShrinkFloor(ctx context.Context, params ActionParams, siteID, pageID uuid.UUID,
	pageName string, sections []SectionData) error {

	floor, _ := pruneFloorFromConfig(params.StepConfig.Config, sectionShrinkFloorKey, defaultSectionShrinkFloor)
	if floor <= 0 {
		params.Logger.Info("save_page_sections: section shrink guard disabled by config",
			zap.String("page_name", pageName))
		return nil
	}

	// Locked rows are excluded from the existing side: for a locked slot the
	// save DISCARDS the incoming copy and the locked row stands (bugs_open/058),
	// so this save cannot shrink it — and comparing the discarded incoming
	// against the locked existing is the false-refusal trap the completeness
	// floor already documented for its own cohort.
	//
	// The measurement moved OUT of SQL, for the reason the sibling component floor
	// already gives two functions away (save_sections_component_floor.go:153):
	// "an undercount on the EXISTING side silently shrinks the floor". Visible text
	// needs a real parse — REGEXP_REPLACE cannot exclude what is inside <style> —
	// and that floor already fetches these exact rows and columns on every save of
	// this action, so this is a second copy of a fetch the path already pays for,
	// not a new class of cost (~10 KB × ~8 slots on the connection that is about to
	// DELETE and re-INSERT the same rows).
	rows, err := params.DB.QueryContext(ctx, `
		SELECT slot_name, rendered_html
		FROM page_components
		WHERE page_id = $1
		  AND rendered_html IS NOT NULL
		  AND `+pageComponentAgentWritableSQL("")+`
	`, pageID)
	if err != nil {
		// Fail CLOSED, matching the completeness floor: a guard that cannot
		// measure must not wave the write through on the strength of that.
		reason := fmt.Sprintf("save_page_sections: REFUSED for page %q — the section shrink guard could not measure existing slots (%v), so nothing was deleted or written", pageName, err)
		params.Logger.Error(reason)
		_ = emitPruneRefusalWorkItem(ctx, params.DB, savePageSectionsRefusal(siteID, pageID, pageName, reason,
			fmt.Sprintf("Page save refused: the section shrink guard could not measure %q's existing sections", pageName),
			shrinkMeasurementErrorFix), params.Logger)
		return fmt.Errorf("%s", reason)
	}
	defer rows.Close()

	existing := make(map[string]int)
	for rows.Next() {
		var slot, html string
		if scanErr := rows.Scan(&slot, &html); scanErr != nil {
			continue
		}
		// += for the same reason as visibleIncomingBySlot: slot names repeat, and
		// last-row-wins made this comparison depend on DB row order.
		existing[slot] += visibleTextLength(html)
	}

	violations := evaluateSectionShrink(floor, minShrinkGuardVisibleChars, existing, visibleIncomingBySlot(sections))
	if len(violations) == 0 {
		return nil
	}

	parts := make([]string, 0, len(violations))
	for _, v := range violations {
		parts = append(parts, fmt.Sprintf("%s %d→%d chars of VISIBLE text, stylesheet and script content excluded (%.0f%% kept, floor %.0f%%)",
			v.Slot, v.Existing, v.Incoming, v.ratio()*100, floor*100))
	}
	reason := fmt.Sprintf(
		"save_page_sections: SECTION SHRINK REFUSED for page %q — %s. A same-named prose slot may not lose more than %.0f%% of the text a READER sees in one save; if this shrink is intended, set %s in the step config. Nothing was written (bugs_open/178, axis corrected by bugs_open/293).",
		pageName, strings.Join(parts, "; "), (1-floor)*100, sectionShrinkFloorKey)

	params.Logger.Warn("SavePageSectionsAction: SECTION SHRINK BLOCKED",
		zap.String("page_name", pageName),
		zap.Float64("floor", floor),
		zap.Int("violations", len(violations)))
	_ = emitPruneRefusalWorkItem(ctx, params.DB, savePageSectionsRefusal(siteID, pageID, pageName, reason,
		fmt.Sprintf("Page save refused: a prose section of %q shrank past the floor", pageName),
		shrinkRefusalFix), params.Logger)
	return fmt.Errorf("%s", reason)
}

// shrinkRefusalFix is the shrink guard's own aftermath sentence. The first
// induced refusal surfaced in the queue wearing the completeness floor's
// summary ("returned too few sections") — false for a shrink, where every
// section came back and one was too small — so this guard now states its own
// case instead of borrowing its sibling's.
const shrinkRefusalFix = "A save would have shrunk a prose section past the configured floor, so NOTHING was " +
	"written and the existing page still stands (bugs_open/178). Decide: if the shrink is intended, " +
	"lower section_shrink_floor on the save_page_sections step (0 disables the guard); otherwise find " +
	"why the writer regenerated the whole section instead of editing it — the reason field names the " +
	"slot and the exact before/after sizes. The sizes are VISIBLE text — what a reader sees, with " +
	"stylesheet and script content excluded — so a slot whose HTML grew can still be refused, and " +
	"that is the case worth looking at first (bugs_open/293)."

// shrinkMeasurementErrorFix is the fail-closed path's OWN sentence — the 98aa9103
// council round approved the split with three seats objecting to this path
// borrowing the violation sentence: nothing shrank here, the guard could not
// MEASURE, and 'lower the floor' is not a remedy for a query that failed.
const shrinkMeasurementErrorFix = "The shrink guard could not measure the page's existing sections, so it failed " +
	"closed and NOTHING was written; the existing page still stands. This is usually a transient " +
	"database error — the save retries with the queue. If it recurs on this page, inspect its " +
	"page_components rows rather than tuning any floor; section_shrink_floor=0 bypasses the guard " +
	"entirely and is the deliberate escape hatch, not a fix."

// ── THE PAGE-TOTAL TEXT FLOOR ────────────────────────────────────────────────
//
// The oldest of this family's three text guards, and until 2026-08-17 the only one
// with no file, no test and no escape hatch: it was an anonymous block inlined in
// SavePageSectionsAction, comparing the whole page's tag-stripped text against a
// quarter of what is deployed. Its stated purpose is unchanged and it is a good
// one — "LLM failures (credit exhaustion, timeouts, empty responses) must not wipe
// good content that was previously generated and deployed".
//
// WHY IT MOVED HERE AND CHANGED AXIS (bugs_open/293). It made the same
// stylesheet-counts-as-text substitution as its per-slot sibling, and being a
// page-TOTAL made it worse: a stylesheet anywhere on the page props up the total
// that every slot's loss is diluted into. `[APPROXIMATE — measured over paired
// slots, not every deployed row, so treat the ratio and not the count as the
// finding]`: across 366 pages it would ALLOW a whole-page prose wipe on 337 (92%),
// where the visible axis refuses 363 of 363 — and it refuses ZERO of the real
// rebuild writes in the same window on either axis, so the correction costs
// nothing measurable. Leaving the blindest of the three uncorrected inside the fix
// for "one call site gets the rigorous fix, the sibling stays heuristic" (016b §9)
// would have reproduced the defect being fixed.
//
// The GUARANTEE is unchanged — refuse a near-wipe of the page — so this is a
// correction of a measure toward its own stated intent, not new authority on a
// shared seam. What IS new is pageTotalTextFloorKey, and it is opt-in with the
// unsafe side (today's behaviour, no operator override) as the default: zero live
// consumers name it at the time of writing.
//
// It keeps its own population deliberately: build_status='deployed' rows, NOT the
// agent-writable set the per-slot floor measures. The two answer different
// questions — "is the page about to lose what it is currently SERVING?" versus
// "may automation rewrite this row?" — and folding them together would silently
// change one of them.
//
// AND THE POPULATION IS LIVE, which is a measurement and not an inheritance
// (council 823679dc round 1, guardian seat, HIGH and gating — the right question,
// asked because the landmine on the SIBLING table records
// site_components.build_status as 'rendered' and NEVER 'deployed'; had this table
// matched, the extracted floor would have returned zero rows for every page,
// short-circuited on its own minimum every time, and never engaged, with a fresh
// sqlmock suite certifying it). `[MEASURED 2026-08-17]`
// SELECT build_status, count(*) FROM page_components GROUP BY 1 → deployed 1,575
// rows / 617 pages, approved 85, pending 19, removed 4; and 617 of 729 pages
// (84.6%) carry at least one deployed row with a non-null rendered_html. The
// analogy does not transfer. Re-run that GROUP BY before trusting this predicate
// again — an inherited predicate has whatever validity it always had, and
// extraction is the moment it becomes yours.
//
// CONSUMERS OF THE SHARED SAVE PATH, enumerated rather than asserted (same council
// round, guardian, medium: the minimum moving 500 → 200 is a behavioural change to
// a guard every content-mutating workflow inherits, so name them). Live, active,
// non-snapshot agent_definitions with a `save_page_sections` step: **page-build-handler,
// page-rerender, tool-recreation-handler** — and `apply_section_edit`, which shares the
// minimum through enforceSingleSlotFloors: **section-editor**. Four in total, and
// NONE of them pins `section_shrink_floor` in step config, so all four take the new
// default. That is the set to tell, per the owner ruling of 2026-07-29 §3.
func enforcePageTotalTextFloor(ctx context.Context, params ActionParams, siteID, pageID uuid.UUID,
	pageName string, sections []SectionData) error {

	floor, _ := pruneFloorFromConfig(params.StepConfig.Config, pageTotalTextFloorKey, defaultPageTotalTextFloor)
	if floor <= 0 {
		params.Logger.Info("save_page_sections: page-total text floor disabled by config",
			zap.String("page_name", pageName))
		return nil
	}
	// CLAMPED, like both sibling floors (evaluateSectionShrink, evaluateComponentLoss)
	// — and found MISSING here by using it, 2026-08-18. Inducing a refusal to prove
	// this floor fires at the artefact needed a floor above 1.0 (so that a payload of
	// the page's OWN sections, byte-for-byte, would be refused and a failure to refuse
	// would write identical content — safe in both branches). `page_total_text_floor:
	// 1.5` did exactly that, which is the proof AND the defect: a floor above 1 demands
	// that a save GROW, so a config typo of 1.5 or 2.5 refuses every save on that step
	// silently and for ever, and on this path a refusal fails the step and can strand a
	// whole build loop. Its siblings have clamped from the start, for the reason their
	// own test states: "an absurd floor is clamped to 0.95, not treated as 'refuse
	// everything'". The clamp costs the byte-for-byte induction (0.95 allows identical
	// content), so the repeatable recipe now reduces one slot's prose by ~6% instead —
	// see the RUNBOOK.
	if floor > 0.95 {
		floor = 0.95
	}

	rows, err := params.DB.QueryContext(ctx, `
		SELECT rendered_html
		FROM page_components
		WHERE page_id = $1
		  AND rendered_html IS NOT NULL
		  AND build_status = 'deployed'
	`, pageID)
	if err != nil {
		// FAILS CLOSED, like its two siblings — changed in round 2 of council
		// 823679dc, and the change is the objection rather than a deferral of it.
		//
		// It fails open in the inline rule this replaced (`scanErr == nil &&`), and
		// round 1 shipped that behaviour unchanged plus a breadcrumb work item, on
		// the reasoning that reconciling it was a behavioural change that should not
		// ride an axis correction. Two seats pushed back and both were right.
		// bug_historian: a breadcrumb "patches the SYMPTOM (visibility after the
		// fact) while leaving the mechanism (fail-open on a content-guard) live and
		// generic", and a record with no consumer is the estate's own
		// "recorded decision with no enforcement point is decorative". reuse_agent:
		// a work-item type whose whole point is that nothing acts on it does not
		// belong in the DISPATCH table at all.
		//
		// WHAT MADE IT CHEAP, which is what I had not measured when I deferred it:
		// this floor runs FIRST of the three (save_page_sections_action.go:593, then
		// :603, then :614), and both of the others query the SAME table for the SAME
		// page moments later and REFUSE on a query error. So the fail-open window was
		// never "an error means the page saves unguarded" — it was only "an error
		// affecting this one statement and not the next two", i.e. a transient blip
		// between statements. Fail-closed therefore changes almost nothing in
		// practice and makes the three consistent, which is worth more than the
		// asymmetry was.
		reason := fmt.Sprintf("save_page_sections: REFUSED for page %q — the page-total text floor could not measure the deployed sections (%v), so nothing was deleted or written", pageName, err)
		params.Logger.Error(reason)
		_ = emitPruneRefusalWorkItem(ctx, params.DB, savePageSectionsRefusal(siteID, pageID, pageName, reason,
			fmt.Sprintf("Page save refused: the page-total text floor could not measure %q's deployed sections", pageName),
			pageTotalMeasurementErrorFix), params.Logger)
		return fmt.Errorf("%s", reason)
	}
	defer rows.Close()

	existingTotal := 0
	deployedRows := 0
	for rows.Next() {
		var html string
		if scanErr := rows.Scan(&html); scanErr != nil {
			continue
		}
		deployedRows++
		existingTotal += visibleTextLength(html)
	}

	// Below the minimum a ratio is noise, and a page with nothing deployed has
	// nothing to protect — both were the inline rule's behaviour.
	if existingTotal <= minPageTotalTextChars {
		return nil
	}

	incomingTotal := 0
	for _, s := range sections {
		incomingTotal += visibleTextLength(s.HTML)
	}
	if float64(incomingTotal) >= float64(existingTotal)*floor {
		return nil
	}

	reason := fmt.Sprintf(
		"save_page_sections: PAGE CONTENT REGRESSION REFUSED for page %q — the incoming sections carry %d chars of VISIBLE text against %d deployed across %d sections (%.0f%% kept, floor %.0f%%), with stylesheet and script content excluded from both sides. This usually means the writer returned empty or near-empty content. Nothing was written; if the cut is intended, set %s on this step (0 disables). (bugs_open/293)",
		pageName, incomingTotal, existingTotal, deployedRows,
		float64(incomingTotal)/float64(existingTotal)*100, floor*100, pageTotalTextFloorKey)

	params.Logger.Warn("SavePageSectionsAction: PAGE CONTENT REGRESSION BLOCKED",
		zap.String("page_name", pageName),
		zap.Int("existing_visible_chars", existingTotal),
		zap.Int("incoming_visible_chars", incomingTotal),
		zap.Int("deployed_sections", deployedRows),
		zap.Float64("floor", floor))
	_ = emitPruneRefusalWorkItem(ctx, params.DB, savePageSectionsRefusal(siteID, pageID, pageName, reason,
		fmt.Sprintf("Page save refused: %q would lose most of the text it currently serves", pageName),
		pageTotalRefusalFix), params.Logger)
	return fmt.Errorf("%s", reason)
}

// pageTotalRefusalFix is this floor's own aftermath sentence. Its remedy is NOT
// the per-slot floor's: no single slot need have shrunk for this to fire, so
// "find why the writer regenerated that section" sends the reader to the wrong
// question.
const pageTotalRefusalFix = "The whole page's text fell past the page-total floor, so NOTHING was written and " +
	"the deployed page still stands. No single section need have shrunk for this to fire — the usual " +
	"cause is an LLM step that returned empty or truncated content for most sections at once, so check " +
	"the composition step's output and llm_call_log for that run before touching any floor. The sizes " +
	"are VISIBLE text, so a page whose HTML is larger than before can still be refused. " +
	"page_total_text_floor on the save_page_sections step is the escape hatch (0 disables) when the " +
	"page is genuinely meant to shed most of its prose."

// pageTotalMeasurementErrorFix — the fail-CLOSED path's own sentence. Nothing
// shrank here and the floor was never consulted, so "lower the floor" is the wrong
// advice; that distinction is why its sibling carries shrinkMeasurementErrorFix
// separately from shrinkRefusalFix.
const pageTotalMeasurementErrorFix = "The page-total text floor could not read this page's deployed sections, so it " +
	"failed CLOSED and NOTHING was written; the page still serves what it served before. Usually a transient " +
	"database error — the save retries with the queue, and its two sibling floors would have refused this save " +
	"anyway, since they query the same table for the same page immediately afterwards. If it recurs on one page, " +
	"read that page's page_components rows rather than tuning a floor: a page with no build_status='deployed' rows " +
	"is out of this floor's POPULATION by design and is not an error, and page_total_text_floor=0 is the deliberate " +
	"bypass rather than a fix (bugs_open/293)."
