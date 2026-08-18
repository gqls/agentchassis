package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// THE WIRING OF THE WHOLE-PAGE TEXT FLOORS, not their arithmetic (bugs_open/293).
//
// Every test in this file drives an enforce* function against a mocked database,
// and each one is paired with a named MUTATION that must make it fail. That
// pairing is the whole point: the 285 lane shipped four tests that all passed
// with the axis REVERTED at the call site, because they exercised the helper and
// the arithmetic and never asked whether anyone called them. The mutations were
// performed on 2026-08-17 before commit, and the result of each is recorded in
// NOTES_whole_page_shrink_axis.md — a mutation you did not run is a claim.
//
// The mocks match the guards' real queries in order; each guard reads
// page_components once. sqlmock's ordered expectations are what make "this guard
// ran" observable at all.

func shrinkGuardTestParams(db *sql.DB) ActionParams {
	return ActionParams{
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
	}
}

// MUT-293-A: in enforceSectionShrinkFloor, measure the existing and incoming
// sides with the retired axis (strippedIncomingBySlot / shrinkGuardTagStripper)
// instead of visibleTextLength. This test must then FAIL — on this pair the
// retired axis reads 262% retained and ALLOWS the write that emptied a live
// article body.
func TestEnforceSectionShrinkFloor_RefusesTheStylesheetForProseSwap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	article, poison := portedArticleFixture(2750), poisonFixture(7000)

	// Fixture guard: without this the test could pass for the wrong reason — a
	// pair that shrinks on BOTH axes proves nothing about which one is wired.
	if tagStrippedLengthForCalibration(poison) <= tagStrippedLengthForCalibration(article) {
		t.Fatalf("fixture no longer reproduces the trap: the retired axis must GROW here (%d → %d)",
			tagStrippedLengthForCalibration(article), tagStrippedLengthForCalibration(poison))
	}

	mock.ExpectQuery("SELECT slot_name, rendered_html").
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}).
			AddRow("ported-page", article))

	err = enforceSectionShrinkFloor(context.Background(), shrinkGuardTestParams(db),
		uuid.Nil, uuid.Nil, "learn-ai-builders-content-first",
		[]SectionData{{ComponentName: "ported-page", HTML: poison, Position: 1}})

	if err == nil {
		t.Fatal("the whole-page shrink floor ALLOWED a stylesheet replacing a live article body " +
			"(bugs_closed/285's write). The retired tag-stripped axis reads 262% retained on this pair; " +
			"only visible text refuses it.")
	}
	if !strings.Contains(err.Error(), "VISIBLE text") {
		t.Errorf("the refusal must name what was measured, or the operator tunes the wrong thing: %v", err)
	}
}

// The other direction, and it is not symmetry for its own sake: the retired axis
// refuses THIS write (38% kept), so a guard enforcing both axes with an OR would
// block the only thing that repairs the damage. That is why 293 is a correction
// of the axis rather than an added floor.
//
// MUT-293-A also breaks this test, in the opposite direction — which is what
// makes the pair diagnostic rather than merely red.
func TestEnforceSectionShrinkFloor_AllowsTheRepairThatPutsTheArticleBack(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	article, poison := portedArticleFixture(2750), poisonFixture(7000)

	mock.ExpectQuery("SELECT slot_name, rendered_html").
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}).
			AddRow("ported-page", poison))

	if err := enforceSectionShrinkFloor(context.Background(), shrinkGuardTestParams(db),
		uuid.Nil, uuid.Nil, "learn-ai-builders-content-first",
		[]SectionData{{ComponentName: "ported-page", HTML: article, Position: 1}}); err != nil {
		t.Fatalf("the floor REFUSED the repair that restored the article (seed 431's write): %v", err)
	}
}

// MUT-293-B: restore minShrinkGuardChars (500) as the minimum passed by
// enforceSectionShrinkFloor. This test must then FAIL — a 300-visible-char prose
// block is out of scope at 500 and its whole content can be deleted unnoticed.
// 587 of 1,079 archived rebuild pairs sat in exactly this band.
//
// Nothing in the suite before this test could catch that mutation: every other
// fixture is comfortably over 500 visible chars.
func TestEnforceSectionShrinkFloor_MidSizedProseBlockIsInScope(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// ~300 visible chars of prose, plus a stylesheet big enough that the retired
	// axis would have called it a comfortable 2,000-character section.
	existing := `<section class="prose"><style>` +
		strings.Repeat(".prose p{margin-block:1.25rem;line-height:1.6;}", 40) +
		`</style><p>` + strings.Repeat("word ", 60) + `</p></section>`
	incoming := `<section class="prose"><style>` +
		strings.Repeat(".prose p{margin-block:1.25rem;line-height:1.6;}", 40) +
		`</style><p>` + strings.Repeat("word ", 10) + `</p></section>`

	ex := visibleTextLength(existing)
	if ex < minShrinkGuardVisibleChars || ex >= minShrinkGuardChars {
		t.Fatalf("fixture is %d visible chars: it must sit BETWEEN the shipped minimum (%d) and the "+
			"retired one (%d) or it cannot detect MUT-293-B", ex, minShrinkGuardVisibleChars, minShrinkGuardChars)
	}

	mock.ExpectQuery("SELECT slot_name, rendered_html").
		WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}).
			AddRow("generic-text-block", existing))

	if err := enforceSectionShrinkFloor(context.Background(), shrinkGuardTestParams(db),
		uuid.Nil, uuid.Nil, "about",
		[]SectionData{{ComponentName: "generic-text-block", HTML: incoming, Position: 1}}); err == nil {
		t.Fatalf("a %d→%d visible-char prose block lost 83%% of what a reader sees and the floor allowed it: "+
			"it is in scope only if the minimum is %d, not %d",
			ex, visibleTextLength(incoming), minShrinkGuardVisibleChars, minShrinkGuardChars)
	}
}

// MUT-293-C: replace `existing[slot] +=` and `m[s.ComponentName] +=` with `=`.
// This test must then FAIL for ONE of the two row orderings — which one depends
// on the mutation, and that is the finding: with last-wins the verdict is decided
// by the order the database happened to return rows in.
//
// Slot names legitimately repeat on a page (14 pages; LANDMINES.md records that
// 11 of 17 such groups are legitimate, `generic-text-block` 2–3× with differing
// content), so this is not a hypothetical shape.
func TestEnforceSectionShrinkFloor_RepeatedSlotNameIsJudgedAsAGroup(t *testing.T) {
	big := `<div class="block"><p>` + strings.Repeat("word ", 120) + `</p></div>`   // ~600 visible
	small := `<div class="block"><p>` + strings.Repeat("word ", 12) + `</p></div>`  // ~60 visible
	tiny := `<div class="block"><p>` + strings.Repeat("word ", 4) + `</p></div>`    // ~20 visible

	// Both orderings must reach the SAME verdict. Under last-write-wins they do
	// not: whichever instance is scanned last becomes the whole comparison.
	for _, order := range [][2]string{{big, small}, {small, big}} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		mock.ExpectQuery("SELECT slot_name, rendered_html").
			WillReturnRows(sqlmock.NewRows([]string{"slot_name", "rendered_html"}).
				AddRow("generic-text-block", order[0]).
				AddRow("generic-text-block", order[1]))

		// The group holds ~660 visible chars and the save offers ~40: a collapse
		// on any honest reading of the group.
		err = enforceSectionShrinkFloor(context.Background(), shrinkGuardTestParams(db),
			uuid.Nil, uuid.Nil, "technical-architecture",
			[]SectionData{
				{ComponentName: "generic-text-block", HTML: tiny, Position: 1},
				{ComponentName: "generic-text-block", HTML: tiny, Position: 2},
			})
		if err == nil {
			t.Fatalf("two instances of one slot name holding ~660 visible chars were replaced by ~40 and the "+
				"floor allowed it (row order %.30s…): the comparison must be over the slot-name GROUP, not "+
				"whichever instance was scanned last", order[0])
		}
		db.Close()
	}
}

// MUT-293-D: revert enforcePageTotalTextFloor to the retired axis. This test must
// then FAIL — every slot here is individually below the per-slot minimum, so the
// per-slot floor is silent by its own rule and the page-total floor is the only
// thing standing between a whole-page prose wipe and the bucket. The stylesheets
// keep the tag-stripped total up, which is exactly why the inline rule would have
// allowed a whole-page wipe on 337 of 366 pages.
//
// It also asserts the per-slot floor stays silent, so the test cannot pass by
// accident on its sibling's verdict.
func TestEnforcePageTotalTextFloor_CatchesTheWipeThePerSlotFloorCannotSee(t *testing.T) {
	css := strings.Repeat(".s{display:grid;gap:1rem;padding:2rem;border-radius:8px;}", 30)
	small := func(words int) string {
		return `<section class="s"><style>` + css + `</style><p>` +
			strings.Repeat("word ", words) + `</p></section>`
	}
	// DISTINCT slot names, deliberately. Four instances of ONE name would be
	// aggregated by the per-slot floor into a 480-visible-char group that clears
	// its minimum — which is the aggregation working, but it would mean the
	// per-slot floor, not the page-total one, produced the refusal. Distinct names
	// are what leave every slot individually under the minimum.
	slotNames := []string{"hero", "intro", "detail", "closing"}
	existingRows := []string{small(30), small(30), small(30), small(30)} // ~150 visible each
	incoming := []SectionData{}
	for i := range existingRows {
		incoming = append(incoming, SectionData{
			ComponentName: slotNames[i], HTML: small(1), Position: i + 1,
		})
	}

	// Preconditions, both load-bearing.
	for _, h := range existingRows {
		if v := visibleTextLength(h); v >= minShrinkGuardVisibleChars {
			t.Fatalf("fixture slot has %d visible chars — at or above the per-slot minimum %d, so this test "+
				"would not isolate the page-total floor", v, minShrinkGuardVisibleChars)
		}
	}
	// The trap condition, stated as the thing it actually is: on the RETIRED axis
	// this wipe must come out ALLOWED, or the test cannot detect MUT-293-D. The
	// earlier version of this assertion compared the existing total against twice
	// the incoming total, which is not the same quantity and would have passed on a
	// fixture the retired axis refused.
	retiredExisting, retiredIncoming := 0, 0
	for _, h := range existingRows {
		retiredExisting += tagStrippedLengthForCalibration(h)
	}
	for _, s := range incoming {
		retiredIncoming += tagStrippedLengthForCalibration(s.HTML)
	}
	if float64(retiredIncoming) < float64(retiredExisting)*defaultPageTotalTextFloor {
		t.Fatalf("fixture no longer reproduces the trap: the retired axis REFUSES this wipe (%d→%d tag-stripped, "+
			"floor %.0f%%), so passing proves nothing about which axis is wired",
			retiredExisting, retiredIncoming, defaultPageTotalTextFloor*100)
	}

	// The per-slot floor must stand down: every slot is under its minimum.
	dbA, mockA, _ := sqlmock.New()
	defer dbA.Close()
	rowsA := sqlmock.NewRows([]string{"slot_name", "rendered_html"})
	for i, h := range existingRows {
		rowsA.AddRow(slotNames[i], h)
	}
	mockA.ExpectQuery("SELECT slot_name, rendered_html").WillReturnRows(rowsA)
	if err := enforceSectionShrinkFloor(context.Background(), shrinkGuardTestParams(dbA),
		uuid.Nil, uuid.Nil, "services", incoming); err != nil {
		t.Fatalf("precondition: the per-slot floor must be SILENT here (every slot is individually under its "+
			"minimum) — got %v", err)
	}

	// The page-total floor must refuse.
	dbB, mockB, _ := sqlmock.New()
	defer dbB.Close()
	rowsB := sqlmock.NewRows([]string{"rendered_html"})
	for _, h := range existingRows {
		rowsB.AddRow(h)
	}
	mockB.ExpectQuery("SELECT rendered_html").WillReturnRows(rowsB)
	err := enforcePageTotalTextFloor(context.Background(), shrinkGuardTestParams(dbB),
		uuid.Nil, uuid.Nil, "services", incoming)
	if err == nil {
		t.Fatal("the page-total floor ALLOWED a whole-page prose wipe: four sections holding ~600 visible " +
			"chars between them replaced by ~20, with only their stylesheets left. On the retired axis the " +
			"stylesheets keep the total up, which is the 337-of-366 blindness bugs_open/293 measured.")
	}
	if !strings.Contains(err.Error(), "VISIBLE text") {
		t.Errorf("the refusal must name what was measured: %v", err)
	}
}

// The allow arm. A floor with no allow arm is a floor nobody can distinguish from
// "refuse everything", and this one guards a whole-page save — the most expensive
// thing on the path to refuse wrongly.
func TestEnforcePageTotalTextFloor_AllowsAnOrdinaryRebuild(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	page := func(words int) string {
		return `<section class="s"><style>.s{gap:1rem}</style><p>` +
			strings.Repeat("word ", words) + `</p></section>`
	}
	mock.ExpectQuery("SELECT rendered_html").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(page(200)))

	// A rewrite that tightens the page by a third: well inside the 0.25 floor.
	if err := enforcePageTotalTextFloor(context.Background(), shrinkGuardTestParams(db),
		uuid.Nil, uuid.Nil, "about",
		[]SectionData{{ComponentName: "block", HTML: page(130), Position: 1}}); err != nil {
		t.Fatalf("an ordinary tightening rebuild was refused: %v", err)
	}
}

// The escape hatch, which this floor did not have before it was extracted: as an
// inline block with hardcoded thresholds, the only way past it was a binary roll.
// A hatch nothing exercises is a hatch nobody can rely on in an incident.
func TestEnforcePageTotalTextFloor_ConfigZeroDisablesIt(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	params := shrinkGuardTestParams(db)
	params.StepConfig = models.Step{Config: map[string]interface{}{pageTotalTextFloorKey: 0}}

	// No query is expected: disabled means it must not even measure. An
	// unexpected query would surface as a refusal here only if the guard also
	// failed closed, so assert on the absence of the expectation instead.
	if err := enforcePageTotalTextFloor(context.Background(), params, uuid.Nil, uuid.Nil,
		"about", []SectionData{{ComponentName: "block", HTML: "<p>x</p>", Position: 1}}); err != nil {
		t.Fatalf("%s=0 must disable the floor entirely: %v", pageTotalTextFloorKey, err)
	}
}

// MUT-293-I: make enforcePageTotalTextFloor return nil on a query error again (the
// fail-open behaviour it shipped with in round 1, inherited from the inline rule).
// This test must then FAIL.
//
// Why it exists at all: while this floor failed open, a test whose mock simply did
// not expect its query saw the guard stand down and PASSED — so "the suite is
// green" said nothing about whether the floor ran. That is the shape bug_historian
// objected to in council 823679dc, and the objection was right twice over: it made
// the production behaviour indistinguishable from no floor, and it made the test
// suite unable to notice.
//
// It fails closed now because the change turned out to be nearly free: this floor
// runs FIRST of the three, and the two after it query the same table for the same
// page and refuse on a query error — so the fail-open window was only ever a
// blip affecting one statement and not the next two.
func TestEnforcePageTotalTextFloor_FailsClosedWhenItCannotMeasure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT rendered_html").WillReturnError(sql.ErrConnDone)

	err = enforcePageTotalTextFloor(context.Background(), shrinkGuardTestParams(db),
		uuid.Nil, uuid.Nil, "about",
		[]SectionData{{ComponentName: "block", HTML: "<p>anything</p>", Position: 1}})
	if err == nil {
		t.Fatal("the page-total text floor could not measure the page and ALLOWED the save anyway. A content " +
			"guard that stands down on its own measurement error is indistinguishable from no guard — and its " +
			"two siblings refuse on exactly this error, so allowing here is the odd one out, not the safe one.")
	}
	if !strings.Contains(err.Error(), "could not measure") {
		t.Errorf("the refusal must say the measurement failed, not imply content shrank — nothing shrank here "+
			"and 'lower the floor' is the wrong remedy for a query that errored. Got: %v", err)
	}
}

// MUT-293-J: delete the `if floor > 0.95 { floor = 0.95 }` clamp from
// enforcePageTotalTextFloor. This test must then FAIL.
//
// The clamp was found missing by USING it: proving this floor fires at the artefact
// (2026-08-18, v1.0.1309) needed a floor above 1.0, so that a payload of the page's
// own sections byte-for-byte would be refused — which made the induction safe in
// both branches and simultaneously exposed the gap. A floor above 1 demands that a
// save GROW, so a typo of `1.5` refuses every save on that step, silently and for
// ever; and on this path a refusal fails the step, which can strand a whole build
// loop. Both siblings clamp, and this one did not.
func TestEnforcePageTotalTextFloor_AbsurdFloorIsClampedNotRefuseEverything(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	page := `<section class="s"><p>` + strings.Repeat("word ", 200) + `</p></section>`
	mock.ExpectQuery("SELECT rendered_html").
		WillReturnRows(sqlmock.NewRows([]string{"rendered_html"}).AddRow(page))

	params := shrinkGuardTestParams(db)
	// A floor of 1.5 would demand 150% of the deployed text: unsatisfiable by any
	// save that is not a large expansion, so unclamped it refuses everything.
	params.StepConfig = models.Step{Config: map[string]interface{}{pageTotalTextFloorKey: 1.5}}

	// IDENTICAL content — 100% kept. Clamped to 0.95 this must be ALLOWED.
	if err := enforcePageTotalTextFloor(context.Background(), params, uuid.Nil, uuid.Nil,
		"about", []SectionData{{ComponentName: "block", HTML: page, Position: 1}}); err != nil {
		t.Fatalf("a floor of 1.5 must clamp to 0.95, so a save keeping 100%% of the page's visible text is "+
			"ALLOWED. Unclamped it demands the save GROW by half, which no ordinary rebuild does — a config "+
			"typo would then refuse every save on this step, and a refusal here fails the step and can strand "+
			"a build loop. Got: %v", err)
	}
}
