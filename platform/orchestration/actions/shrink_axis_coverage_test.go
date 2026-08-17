package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// AXIS COVERAGE — every caller of the shared shrink decision must measure with
// visibleTextLength, and the whole-page action must still call both text floors.
//
// WHY A COVERAGE TEST AND NOT A TYPE (bugs_open/293). The obvious structural fix
// is to make evaluateSectionShrink take HTML and measure for itself, so no caller
// CAN choose an axis. It was considered and rejected: the calibration harness's
// entire value is that it runs BOTH axes through the real decision on real pairs
// (shrink_axis_calibration_test.go), and a decision that measures for itself can
// only be trusted, never calibrated. That leaves the axis in the caller's hands,
// which is precisely the arrangement that let the whole-page path go on counting
// stylesheets as prose for a fortnight after the section editor stopped — so what
// holds it now is this test, mechanically, rather than the paragraph in
// single_slot_floors.go that used to call the arrangement "unaffected by
// construction".
//
// Modelled on page_component_writer_coverage_test.go, including its stated
// weakness: it reads SOURCE, so it proves WIRING EXISTS, not that the wiring
// EXECUTES. The behavioural half is in save_sections_shrink_guard_wiring_test.go,
// where each guard is driven against a mocked database and each assertion is
// paired with a named mutation.
//
// It also closes a hole found while writing it: the writer-coverage test
// STRUCTURALLY CANNOT SEE the whole-page save. Its predicate is
// `UPDATE page_components … SET … rendered_html =`, and the rebuild path is
// DELETE+INSERT — the only UPDATEs in save_page_sections_action.go touch
// `position`. So the fleet's highest-volume rendered_html writer (3,603 archived
// rebuild writes in 8 days, against 281 edit writes) was invisible to the fleet's
// writer-coverage test, and unwiring a floor there would have failed nothing.

// callsTheSharedDecision matches a call to the shared pure decision.
var callsTheSharedDecision = regexp.MustCompile(`evaluateSectionShrink\(`)

// measuresVisibleText matches the one sanctioned measure.
var measuresVisibleText = regexp.MustCompile(`visibleTextLength\(`)

// feedsTheRetiredAxis matches a tag-strip being applied to a value — the shape a
// fifth caller would arrive with if it copied the pre-2026-08-17 code. Comments
// cannot trip it: it requires the call, not the name (the landmine about
// source-scanning tests making comments load-bearing is real, and this file's own
// prose mentions the retired symbol repeatedly).
var feedsTheRetiredAxis = regexp.MustCompile(`shrinkGuardTagStripper\.ReplaceAllString\(`)

// retiredAxisConsumers — files allowed to apply the retired tag-strip, each with
// the reason. A file here is a DECISION. Both entries were audited 2026-08-17.
var retiredAxisConsumers = map[string]string{
	// The one remaining live consumer, and deliberately so: it uses the tag strip
	// with minShrinkGuardChars (500) to judge whether an unclaimed stored slot is
	// "prose-sized" enough to be a plausible body-text match for an unpaired
	// section. It REFUSES NOTHING — it is a pairing heuristic, not a content floor
	// — and re-tuning it would change which slots get paired, with no calibration
	// covering that decision.
	"load_current_section_content_action.go": "pairing heuristic, not a floor: decides whether a stored slot is prose-sized enough to pair, refuses nothing",

	// Defines the retired stripper and keeps strippedIncomingBySlot alive as the
	// calibration harness's comparator, which is how the harness proves it is
	// measuring the axis that shipped rather than a lookalike.
	"save_sections_shrink_guard.go": "declares the retired axis and the harness's comparator; neither floor measures with it",
}

func TestEveryShrinkDecisionCallerMeasuresVisibleText(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	callers := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Fatalf("reading %s: %v", f, readErr)
		}
		body := string(src)

		if callsTheSharedDecision.MatchString(body) {
			callers++
			if !measuresVisibleText.MatchString(body) {
				t.Errorf("%s calls evaluateSectionShrink but never calls visibleTextLength.\n"+
					"The axis lives in the CALLER, so a caller that measures some other way silently gets a "+
					"different guarantee — which is bugs_open/293 exactly: the whole-page path counted CSS and "+
					"JavaScript as prose for a fortnight after the section editor stopped. Measure with "+
					"visibleTextLength, or if this caller genuinely needs another axis, calibrate it first "+
					"(shrink_axis_calibration_test.go) and change this test deliberately.", f)
			}
		}

		if feedsTheRetiredAxis.MatchString(body) {
			if reason, ok := retiredAxisConsumers[filepath.Base(f)]; !ok {
				t.Errorf("%s applies the RETIRED tag-stripped axis (shrinkGuardTagStripper) and is not in "+
					"retiredAxisConsumers.\nThat measure counts the contents of <style> and <script> as text, so "+
					"anything guarding CONTENT with it is blind to prose deletion — measured 2026-08-17: it "+
					"allows a total prose wipe on 724 of the 1,060 sections it judges. Use visibleTextLength, or "+
					"add this file to retiredAxisConsumers WITH the reason it is not guarding content.", f)
			} else if strings.TrimSpace(reason) == "" {
				t.Errorf("%s is in retiredAxisConsumers with an empty reason — the reason is the point", f)
			}
		}
	}

	// VACUITY CONTROL, mirroring the writer-coverage test's sawAnyWriter: a glob
	// that matched nothing, or a regexp that stopped matching after a rename,
	// would report perfect coverage of an empty set.
	if callers < 2 {
		t.Fatalf("found %d callers of evaluateSectionShrink; expected at least 2 (the whole-page guard and the "+
			"section editor). A scan that finds nothing PASSES, so this check is what stops the test quietly "+
			"certifying an empty set.", callers)
	}
}

// The whole-page action must call BOTH text floors. This is the assertion the
// writer-coverage test cannot make, because that test only recognises an UPDATE.
//
// The page-total floor needs it more than its sibling: it fails OPEN on a
// measurement error (as the inline rule it replaced did), so a test whose mock
// simply does not expect its query sees the guard stand down and passes. "The
// suite is green" therefore does not establish that this floor runs at all —
// only this assertion and the wiring test's mocked drive do.
func TestSavePageSectionsWiresBothTextFloors(t *testing.T) {
	src, err := os.ReadFile("save_page_sections_action.go")
	if err != nil {
		t.Fatalf("reading save_page_sections_action.go: %v", err)
	}
	body := string(src)

	for _, want := range []string{
		"enforceSectionShrinkFloor(",
		"enforcePageTotalTextFloor(",
		"enforceSectionComponentFloor(",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("save_page_sections_action.go no longer calls %s — the highest-volume rendered_html writer "+
				"on the fleet (3,603 rebuild writes in 8 days) would then save with one fewer floor, and the "+
				"writer-coverage test cannot see it because this path is DELETE+INSERT, not UPDATE.", want)
		}
	}

	// And the diagnostic must keep reporting the axis the floors actually use.
	// It previously advertised itself as "the stripped-text total the regression
	// guard will compute", which stopped being true the moment the floors moved to
	// visible text — a diagnostic naming the wrong quantity is worse than none,
	// because whoever debugs a refusal reads it as the guard's own arithmetic.
	if !strings.Contains(body, "visible_text_total") {
		t.Error("the sections-reaching-save diagnostic no longer logs visible_text_total: a refusal's cause is " +
			"illegible without the axis the floors measure, and the signature of this bug's failure mode is " +
			"stripped GROWING while visible COLLAPSES — which needs both numbers in the same line")
	}
}
