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
// This guard compares SLOT AGAINST SAME-NAMED SLOT, stripped of tags, and
// refuses the save when a prose-sized slot shrinks past the floor. Slots that
// are new, dropped, or small stay out of scope on purpose:
//   - a slot absent from the incoming set is a DROP — the completeness floor's
//     domain (save_sections_prune_floor.go), not a shrink;
//   - a slot below minShrinkGuardChars of existing stripped text is parameter/
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

const (
	// sectionShrinkFloorKey is the step-config key holding the minimum
	// fraction of a slot's existing stripped text that a replacement must
	// retain. 0 disables the guard for that step.
	sectionShrinkFloorKey = "section_shrink_floor"

	// defaultSectionShrinkFloor: a replacement keeping less than half of a
	// prose slot's text is refused unless the step opts out.
	defaultSectionShrinkFloor = 0.5

	// minShrinkGuardChars: slots whose existing stripped text is below this
	// are out of scope — param blobs and heroes shrink legitimately.
	minShrinkGuardChars = 500
)

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

// evaluateSectionShrink is the pure decision: given existing stripped-text
// lengths by slot and the incoming sections, return the slots that violate the
// floor. Split from the SQL so the rule is testable without a database.
func evaluateSectionShrink(floor float64, existing map[string]int, incoming map[string]int) []slotShrink {
	if floor <= 0 {
		return nil
	}
	if floor > 0.95 {
		floor = 0.95
	}
	var violations []slotShrink
	for slot, existingLen := range existing {
		if existingLen < minShrinkGuardChars {
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

// strippedIncomingBySlot maps the incoming sections to stripped-text lengths
// keyed by slot name, merging duplicates (last write wins matches the insert
// loop's behaviour of writing every section it is given).
func strippedIncomingBySlot(sections []SectionData) map[string]int {
	m := make(map[string]int, len(sections))
	for _, s := range sections {
		stripped := strings.TrimSpace(shrinkGuardTagStripper.ReplaceAllString(s.HTML, ""))
		// ComponentName is what the insert loop writes as slot_name.
		m[s.ComponentName] = len(stripped)
	}
	return m
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
	rows, err := params.DB.QueryContext(ctx, `
		SELECT slot_name,
		       LENGTH(TRIM(REGEXP_REPLACE(rendered_html, '<[^>]*>', '', 'g')))
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
			shrinkRefusalFix), params.Logger)
		return fmt.Errorf("%s", reason)
	}
	defer rows.Close()

	existing := make(map[string]int)
	for rows.Next() {
		var slot string
		var strippedLen int
		if scanErr := rows.Scan(&slot, &strippedLen); scanErr != nil {
			continue
		}
		existing[slot] = strippedLen
	}

	violations := evaluateSectionShrink(floor, existing, strippedIncomingBySlot(sections))
	if len(violations) == 0 {
		return nil
	}

	parts := make([]string, 0, len(violations))
	for _, v := range violations {
		parts = append(parts, fmt.Sprintf("%s %d→%d chars (%.0f%% kept, floor %.0f%%)",
			v.Slot, v.Existing, v.Incoming, v.ratio()*100, floor*100))
	}
	reason := fmt.Sprintf(
		"save_page_sections: SECTION SHRINK REFUSED for page %q — %s. A same-named prose slot may not lose more than %.0f%% of its text in one save; if this shrink is intended, set %s in the step config. Nothing was written (bugs_open/178).",
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
	"slot and the exact before/after sizes."
