package actions

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Per-slot COMPONENT floor (bugs_open/253_..._framework_rewrite_of_a_prose_block).
// Sibling of the text shrink guard in save_sections_shrink_guard.go, and it exists
// because that guard is blind in a direction nobody had measured.
//
// THE CASE. loanandmortgagecalculator.co.uk's homepage was decomposed into one
// `prose-0` component holding the hand-built body byte-for-byte. Four hours later
// the generic pipeline rewrote it. It kept the words and it kept every link — 14
// calculator links before and after, and MORE internal links overall (28 → 34) —
// while removing the site's entire visual vocabulary: `class="card"` 18 → 0,
// `tool-grid` 3 → 0, `btn-primary` 15 → 0, `highlight-box` 1 → 0, `hero` 1 → 0.
// The shopfront became a flat run of headings. Nothing refused it and nothing
// reported it; it was found by an offline byte-diff a human happened to run.
//
// WHY THE TEXT FLOOR CANNOT SEE IT. That guard measures STRIPPED TEXT. An earlier
// attempt on the same page WAS refused by it (3776→1334 chars, 35% kept, floor
// 50%) — it works. The one that landed kept 84% of the text and 2% of the markup.
// Prose volume and layout structure are independent quantities, and only one of
// them was guarded.
//
// CALIBRATION, against the real before/after (2026-08-12), which is the only
// reason the numbers below are not invented:
//
//	prose-0 `class="` occurrences   ratio    verdict wanted
//	  before the rewrite      43      —
//	  the FLATTENED save       1     0.02    must REFUSE
//	  a later GOOD rewrite    31     0.72    must ALLOW
//
// The two cases are 35× apart, so the floor is not finely balanced: 0.5 mirrors
// the text floor and sits in open space between them. All three of 0.25/0.34/0.50
// separate the pair correctly; 0.5 is chosen for consistency with its sibling,
// not because the data demanded that exact value.
//
// SCOPE. Fleet distribution of class occurrences per unlocked slot (1,422 rows,
// measured 2026-08-12): median 5, p90 35, max 424. A floor applied to every slot
// would fire on ordinary variation in a 5-class slot, so slots below
// minComponentGuardClasses are out of scope — that puts ~31% of slots in scope,
// which is the structurally rich cohort where flattening is a real loss.
//
// Like its sibling: growth is always allowed; a slot absent from the incoming set
// is a DROP (the completeness floor's domain, not this one); locked rows are
// excluded because the save discards the incoming copy for them anyway; the floor
// is config-tunable per step and 0 disables it; and a refusal writes NOTHING and
// emits a queue item, because the 178 failures were invisible precisely because
// every surface reported success.

const (
	// sectionComponentFloorKey is the step-config key holding the minimum
	// fraction of a slot's existing class-attribute occurrences that a
	// replacement must retain. 0 disables the guard for that step.
	sectionComponentFloorKey = "section_component_floor"

	// defaultSectionComponentFloor mirrors defaultSectionShrinkFloor. See the
	// calibration table above: the observed bad case is 0.02 and the observed
	// good case 0.72, so this is not a knife-edge.
	defaultSectionComponentFloor = 0.5

	// minComponentGuardClasses: slots carrying fewer class attributes than this
	// are out of scope. Median across the fleet is 5, so a lower bound would
	// make ordinary rewrites of plain-prose slots refusable.
	minComponentGuardClasses = 10
)

// componentClassCounter counts class ATTRIBUTES, not class tokens. Counting
// tokens would let a rewrite that replaces one multi-token element with several
// single-token ones look like growth; counting attributes tracks "how many
// elements carry styling", which is the quantity that collapsed in the case
// above. It deliberately ignores WHICH classes: the site vocabulary is per-site
// and unknown here, and a rewrite that swaps one valid layout for another is not
// what this guard is for.
var componentClassCounter = regexp.MustCompile(`(?i)\sclass\s*=\s*["']`)

func countComponentClasses(html string) int {
	return len(componentClassCounter.FindAllStringIndex(html, -1))
}

// slotComponentLoss describes one same-named slot whose replacement fell below
// the component floor. Existing/Incoming are class-attribute counts.
type slotComponentLoss struct {
	Slot     string
	Existing int
	Incoming int
}

func (s slotComponentLoss) ratio() float64 {
	if s.Existing == 0 {
		return 1
	}
	return float64(s.Incoming) / float64(s.Existing)
}

// evaluateComponentLoss is the pure decision, split from the SQL so the rule is
// testable without a database — same split as evaluateSectionShrink.
func evaluateComponentLoss(floor float64, existing map[string]int, incoming map[string]int) []slotComponentLoss {
	if floor <= 0 {
		return nil
	}
	if floor > 0.95 {
		floor = 0.95
	}
	var violations []slotComponentLoss
	for slot, existingCount := range existing {
		if existingCount < minComponentGuardClasses {
			continue
		}
		incomingCount, present := incoming[slot]
		if !present {
			// A drop, not a flattening — the completeness floor's case.
			continue
		}
		if float64(incomingCount) < float64(existingCount)*floor {
			violations = append(violations, slotComponentLoss{
				Slot: slot, Existing: existingCount, Incoming: incomingCount,
			})
		}
	}
	return violations
}

// componentClassesIncomingBySlot maps incoming sections to class-attribute counts
// by slot, SUMMING same-named sections.
//
// It used to take the last write, "matching strippedIncomingBySlot and the insert
// loop" — and that comment was half right in a way that hid the defect
// (bugs_open/293). It did match its sibling, because both had the same bug; it did
// NOT match the insert loop, which writes EVERY instance it is given (position =
// i+1 for each), so a page with a repeated slot name had its floor decided by one
// arbitrary instance while all of them were written. Which instance won depended
// on slice order here and on DB row order on the existing side, so the verdict was
// not even stable across runs.
//
// A repeated slot name is NORMAL and cannot be made unique: 14 pages carry one, and
// LANDMINES.md records that 11 of 17 such groups are legitimate — `generic-text-block`
// used 2–3× on a page with differing content, `info-card-grid` ×2 — with only 6 ever
// being true duplication. So the unit of judgement is the slot-name GROUP's total
// class count: bit-identical for the ~97% of groups with one instance, deterministic
// for the rest, and needing no assumption that position survives a rebuild.
func componentClassesIncomingBySlot(sections []SectionData) map[string]int {
	return incomingBySlot(sections, countComponentClasses)
}

// enforceSectionComponentFloor refuses a save that would flatten a structurally
// rich slot's markup, on the same terms as the text shrink guard.
func enforceSectionComponentFloor(ctx context.Context, params ActionParams, siteID, pageID uuid.UUID,
	pageName string, sections []SectionData) error {

	floor, _ := pruneFloorFromConfig(params.StepConfig.Config, sectionComponentFloorKey, defaultSectionComponentFloor)
	if floor <= 0 {
		params.Logger.Info("save_page_sections: section component floor disabled by config",
			zap.String("page_name", pageName))
		return nil
	}

	// The count is done in Go, not SQL: the regex tolerates `class = "` and
	// single quotes, which a naive SQL REPLACE of the literal `class="` does
	// not — and an undercount on the EXISTING side silently shrinks the floor.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT slot_name, rendered_html
		FROM page_components
		WHERE page_id = $1
		  AND rendered_html IS NOT NULL
		  AND `+pageComponentAgentWritableSQL("")+`
	`, pageID)
	if err != nil {
		// Fail CLOSED, matching both sibling floors.
		reason := fmt.Sprintf("save_page_sections: REFUSED for page %q — the section component floor could not measure existing slots (%v), so nothing was deleted or written", pageName, err)
		params.Logger.Error(reason)
		_ = emitPruneRefusalWorkItem(ctx, params.DB, savePageSectionsRefusal(siteID, pageID, pageName, reason,
			fmt.Sprintf("Page save refused: the section component floor could not measure %q's existing sections", pageName),
			componentMeasurementErrorFix), params.Logger)
		return fmt.Errorf("%s", reason)
	}
	defer rows.Close()

	existing := make(map[string]int)
	for rows.Next() {
		var slot, html string
		if scanErr := rows.Scan(&slot, &html); scanErr != nil {
			continue
		}
		// += for the same reason as componentClassesIncomingBySlot above: slot
		// names repeat, and last-row-scanned-wins made this comparison depend on
		// the order the database happened to return rows in.
		existing[slot] += countComponentClasses(html)
	}

	violations := evaluateComponentLoss(floor, existing, componentClassesIncomingBySlot(sections))
	if len(violations) == 0 {
		return nil
	}

	parts := make([]string, 0, len(violations))
	for _, v := range violations {
		parts = append(parts, fmt.Sprintf("%s %d→%d class attributes (%.0f%% kept, floor %.0f%%)",
			v.Slot, v.Existing, v.Incoming, v.ratio()*100, floor*100))
	}
	reason := fmt.Sprintf(
		"save_page_sections: SECTION COMPONENT FLOOR REFUSED for page %q — %s. A same-named slot may not lose more than %.0f%% of the elements carrying layout classes in one save, even when its TEXT survives; if this flattening is intended, set %s in the step config. Nothing was written (bugs_open/253, framework_rewrite_of_a_prose_block).",
		pageName, strings.Join(parts, "; "), (1-floor)*100, sectionComponentFloorKey)

	params.Logger.Warn("SavePageSectionsAction: SECTION COMPONENT FLATTENING BLOCKED",
		zap.String("page_name", pageName),
		zap.Float64("floor", floor),
		zap.Int("violations", len(violations)))
	_ = emitPruneRefusalWorkItem(ctx, params.DB, savePageSectionsRefusal(siteID, pageID, pageName, reason,
		fmt.Sprintf("Page save refused: a section of %q would lose its layout components", pageName),
		componentRefusalFix), params.Logger)
	return fmt.Errorf("%s", reason)
}

// componentRefusalFix states this guard's own case. Its sibling's sentence would
// be actively misleading here: nothing shrank, the text may be entirely intact,
// and "the writer regenerated instead of editing" is only one of the causes.
const componentRefusalFix = "A save would have stripped the layout components (cards, grids, buttons) from a " +
	"section while keeping its words, so NOTHING was written and the existing page still stands " +
	"(bugs_open/253). The text shrink guard cannot see this: the case that prompted it kept 84% of " +
	"the text and 2% of the markup. Decide: if the page is deliberately being simplified, set " +
	"section_component_floor on the save_page_sections step (0 disables); otherwise the writer was " +
	"handed a section whose markup it did not know how to preserve — give it the component vocabulary " +
	"in content_direction rather than lowering the floor, which is what fixed the motivating page."

// componentMeasurementErrorFix is the fail-closed path's own sentence, following
// the 98aa9103 council round's finding that a measurement failure must not borrow
// a violation's remedy — nothing was flattened here, the guard could not measure.
const componentMeasurementErrorFix = "The component floor could not measure the page's existing sections, so it " +
	"failed closed and NOTHING was written; the existing page still stands. This is usually a transient " +
	"database error — the save retries with the queue. If it recurs on this page, inspect its " +
	"page_components rows rather than tuning any floor; section_component_floor=0 bypasses the guard " +
	"entirely and is the deliberate escape hatch, not a fix."
