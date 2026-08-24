// FILE: platform/orchestration/actions/refresh_evidence_fact_probe_test.go
//
// bugs_open/288 Phase 3a. The decisive test in this file is
// TestFactProbe_TheWholePageCheckWouldHavePassedBug225 — every other assertion
// is scaffolding around it.
//
// Fixtures are the real shape, measured off the live mortgagecalculator
// stamp-duty page on 2026-08-24: the band table is raw digits in code
// (`{ upTo: 500000, rate: 0.00 }`), the human-readable figures are comma-form
// inside JS comments and strings, and the page's prose carries the comma form
// too because the register's own writer_line put it there.

package actions

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var errSurfaceUnavailable = errors.New("surface read failed")

// The page as it really is: correct copy, correct band table.
const probePageCorrect = `
<section><h1>Stamp duty</h1>
  <p>First-time buyers pay nothing up to &pound;300,000, and relief disappears entirely above &pound;500,000.</p>
</section>
<script>
  const FTB_BANDS = [ { upTo: 300000, rate: 0.00 }, { upTo: 500000, rate: 0.05 } ];
  const FTB_RELIEF_CEILING = 500000;   // "relief disappears above £500,000"
</script>`

// THE MOTIVATING BUG. Identical copy — the register keeps it current — while the
// code still carries the cap that expired on 2025-03-31. bugs_closed/225 ran in
// exactly this state for sixteen months and every check passed it.
//
// ⚠ THE data-relief-cap ATTRIBUTE IS THE WHOLE TEST, and it is here because the
// first version of this fixture DID NOT DISCRIMINATE. Without it, mutating the
// probe to read the whole page instead of the script text left the headline test
// GREEN: the prose writes the comma form (£500,000) and the code surface is
// searched for RAW literals, so the raw search failed on the whole page too and
// the outcome was markup-only either way. The test asserted the right answer for
// the wrong reason.
//
// Component markup really does carry raw values — data attributes, JSON-LD,
// hidden inputs — so a page whose MARKUP says 500000 while its CODE says 625000
// is the realistic case, and it is the only one where "read the script" and
// "read the page" give different verdicts. Read the page and this stale
// calculator is certified present_in_script; read the script and it is correctly
// markup-only.
const probePageBug225 = `
<section><h1>Stamp duty</h1>
  <p>First-time buyers pay nothing up to &pound;300,000, and relief disappears entirely above &pound;500,000.</p>
  <div class="sdlt-widget" data-relief-cap="500000" data-ftb-band="300000"></div>
</section>
<script>
  const FTB_BANDS = [ { upTo: 300000, rate: 0.00 }, { upTo: 625000, rate: 0.05 } ];
  const FTB_RELIEF_CEILING = 625000;
</script>`

// ── The one that matters ───────────────────────────────────────────────────

// A whole-page probe finds £500,000 in the PROSE and reports the tool fine. The
// script-only probe does not. If this test ever goes green while the probe reads
// the whole page, the mechanism is blind to the bug it was built for.
//
// MUTATION THAT MUST GO RED: make probeFactValueOnSurface match against pageHTML
// instead of extractScriptText(pageHTML).
func TestFactProbe_TheWholePageCheckWouldHavePassedBug225(t *testing.T) {
	// PREMISE, asserted rather than assumed: the naive check really would pass.
	// The page carries the current figure OUTSIDE its script (in the copy as
	// £500,000 and in a data attribute as 500000) while the code carries the
	// expired one. If this stops being true the test below proves nothing, and
	// silently — which is the failure this whole file is about.
	if !valueOccursGuarded(probePageBug225, "500,000") {
		t.Fatal("premise broken: the stale page's PROSE must carry the current figure")
	}
	if !valueOccursGuarded(probePageBug225, "500000") {
		t.Fatal("premise broken: the stale page's MARKUP must carry the raw current figure, or reading the whole page is indistinguishable from reading the script")
	}
	if valueOccursGuarded(extractScriptText(probePageBug225), "500000") {
		t.Fatal("premise broken: the stale page's SCRIPT must NOT carry the current figure")
	}

	got := probeFactValueOnSurface(probePageBug225, 500000, true)
	if got.Outcome != factProbeMarkupOnly {
		t.Fatalf("the stale tool must be reported as markup-only, got %q — %s", got.Outcome, got.Detail)
	}
	if !strings.Contains(got.Detail, "225") {
		t.Errorf("the detail should tell the reader what shape this is: %s", got.Detail)
	}
}

func TestFactProbe_CorrectToolIsPresentInScript(t *testing.T) {
	got := probeFactValueOnSurface(probePageCorrect, 500000, true)
	if got.Outcome != factProbePresentInScript {
		t.Fatalf("a tool whose code carries the figure must read present_in_script, got %q — %s", got.Outcome, got.Detail)
	}
	if got.Form != "500000" {
		t.Errorf("the matched form should be recorded for the human reading the item, got %q", got.Form)
	}
}

// Absent from code AND copy. Consistent with a stale tool and with four benign
// causes, so the detail must say so rather than assert a defect.
func TestFactProbe_AbsentEverywhereSaysWhatItCannotConclude(t *testing.T) {
	got := probeFactValueOnSurface(`<p>nothing</p><script>var x = 1;</script>`, 500000, true)
	if got.Outcome != factProbeAbsent {
		t.Fatalf("want absent, got %q", got.Outcome)
	}
	for _, want := range []string{"DERIVES", "benign", "human decides"} {
		if !strings.Contains(got.Detail, want) {
			t.Errorf("absence must not read as a proven defect; detail missing %q: %s", want, got.Detail)
		}
	}
}

// ── The distinctiveness floor, and why it is where it is ───────────────────

// MEASURED 2026-08-24 over the script text of all 161 tool pages that have any,
// using INVENTED values so every match is a false positive by construction:
// 32.75% at one digit, 3.79% at two, 0.06% at three, 0.03% at four, 0.00% at
// five or more. A percentage rate (5, 2, 10, 12 — 110 of 294 current facts are
// under 100) must therefore be refused, not guessed at.
func TestFactProbe_ShortValuesAreRefusedNotGuessed(t *testing.T) {
	page := `<script>const RATES = [2, 5, 10, 12]; var pad = 12;</script>`
	for _, v := range []float64{2, 5, 10, 12, 999} {
		got := probeFactValueOnSurface(page, v, true)
		if got.Outcome != factProbeNotProbed {
			t.Fatalf("value %v is below the measured floor and must be refused, got %q", v, got.Outcome)
		}
		if !strings.Contains(got.Detail, "artifact_check") {
			t.Errorf("a refusal must name the escape hatch that DOES work for a small figure: %s", got.Detail)
		}
	}
	// And the floor is not so high that it refuses the case it exists for.
	if got := probeFactValueOnSurface(probePageCorrect, 500000, true); got.Outcome == factProbeNotProbed {
		t.Fatal("the floor must not refuse bug 225's own fact")
	}
}

func TestFactProbe_NoValueAndNoSurfaceAreDistinctFromAbsent(t *testing.T) {
	if got := probeFactValueOnSurface(probePageCorrect, 0, false); got.Outcome != factProbeNotProbed {
		t.Errorf("a fact with no numeric value cannot be probed, got %q", got.Outcome)
	}
	if got := probeFactValueOnSurface("", 500000, true); got.Outcome != factProbeNoSurface {
		t.Errorf("no stored HTML must be no_surface, never absent — nothing was read, so nothing is claimed, got %q", got.Outcome)
	}
}

// ── The boundary rule, which the first version got wrong ───────────────────

// The trailing-comma case is the one that failed on the real page: excluding
// every trailing comma made `{ upTo: 1500000, rate: 0.10 }` invisible, i.e. the
// probe could not see the actual band table it was written to read.
func TestFactProbe_BoundaryRuleMatchesRealCodeAndRejectsRealTraps(t *testing.T) {
	match := []struct{ text, lit, why string }{
		{`{ upTo: 1500000, rate: 0.10 }`, "1500000", "a trailing comma is a LIST SEPARATOR — the case the first rule broke on"},
		{`const CAP = 500000;`, "500000", "a statement terminator"},
		{`[125000]`, "125000", "a bracket"},
		{`x = 40000`, "40000", "end of input"},
		{`1_500_000;`, "1_500_000", "the underscore form is legal JS"},
	}
	for _, c := range match {
		if !valueOccursGuarded(c.text, c.lit) {
			t.Errorf("must match %q in %q — %s", c.lit, c.text, c.why)
		}
	}
	reject := []struct{ text, lit, why string }{
		{`21500000`, "1500000", "inside a longer number"},
		{`15000007`, "1500000", "inside a longer number, trailing"},
		{`1,500000`, "500000", "a comma BETWEEN digits is a thousands separator"},
		{`0.500000`, "500000", "a dot between digits is a decimal"},
		{`500000.5`, "500000", "a decimal follows"},
		{`500000,000`, "500000", "a thousands separator follows"},
		{`ID500000`, "500000", "part of an identifier"},
		{`500000px`, "500000", "part of an identifier"},
		{`v_500000`, "500000", "part of an identifier"},
	}
	for _, c := range reject {
		if valueOccursGuarded(c.text, c.lit) {
			t.Errorf("must NOT match %q in %q — %s", c.lit, c.text, c.why)
		}
	}
}

// The platform's own documented landmine, applied to this probe: grepping for
// 10000 must not match inside 100000. It is the reason bareNumericPattern exists
// and the reason this probe is guarded rather than a substring search.
func TestFactProbe_TheTenThousandInsideHundredThousandCanary(t *testing.T) {
	if got := probeFactValueOnSurface(`<script>var cap = 100000;</script>`, 10000, true); got.Outcome != factProbeAbsent {
		t.Fatalf("10000 must not be found inside 100000, got %q — this is the estate's own documented trap", got.Outcome)
	}
}

// ── Script extraction ──────────────────────────────────────────────────────

func TestExtractScriptText_TakesOnlyScriptBodies(t *testing.T) {
	got := extractScriptText(probePageCorrect)
	if !strings.Contains(got, "FTB_RELIEF_CEILING") {
		t.Fatal("script bodies must be extracted")
	}
	if strings.Contains(got, "First-time buyers pay nothing") {
		t.Fatal("PROSE must not be in the extracted script text — that is the whole safety property")
	}
	if strings.Contains(got, "<h1>") {
		t.Fatal("markup must not leak into the script text")
	}
}

func TestExtractScriptText_HandlesNoScriptAndJunk(t *testing.T) {
	if got := extractScriptText(`<p>no script here</p>`); strings.TrimSpace(got) != "" {
		t.Errorf("a page with no script has no script text, got %q", got)
	}
	if got := extractScriptText(``); got != "" {
		t.Errorf("empty in, empty out, got %q", got)
	}
	// Malformed markup must not panic or return the whole document.
	if got := extractScriptText(`<script>var a = 1;`); !strings.Contains(got, "var a = 1;") {
		t.Errorf("an unclosed script should still yield its body, got %q", got)
	}
}

// A script with a src and no body contributes nothing — and that is an honest
// no_surface-shaped answer rather than a false absence. 38 of the 178 tool pages
// on register-bearing sites pull an external src (measured 2026-08-24), so this
// is a real population, and it is named as a limit rather than hidden.
func TestFactProbe_ExternalScriptSrcIsNotReadAndDoesNotFakeAnAbsence(t *testing.T) {
	page := `<p>Relief ends above &pound;500,000.</p><script src="/tools/sdlt.js"></script>`
	got := probeFactValueOnSurface(page, 500000, true)
	if got.Outcome != factProbeMarkupOnly {
		t.Fatalf("with the code in an external file the figure is only in the markup, got %q", got.Outcome)
	}
}

// ── WIRED: the annotation reaches the emission from the real planner ───────
//
// Written FIRST this time. Twice already in this change a guard's unit tests all
// passed while the call to it was deleted (WRONG_CALLS 2026-08-24), so the probe
// gets its call-site test before it gets the benefit of the doubt.
//
// MUTATION THAT MUST GO RED: delete the annotateFactDriftEvidence call from
// planSiteFactDrift.
func TestPlanSiteFactDrift_AnnotatesEmissionsWithByteEvidence(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	body := "```criteria\n" + `{"facts":["sdlt-ftb-relief-cap"],"no_auto_fix":true,"checks":[]}` + "\n```"
	pageID := "3d7d0d72-0000-4000-8000-000000000001"
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}).
			AddRow(pageID, "tool-stamp-duty", "", "complete", "stamp-duty", body, ""))
	mock.ExpectQuery(regexp.QuoteMeta(factDriftLastItemQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subject_key", "new_value"}))
	// THE PROBE'S OWN READ. bugs_closed/225's page: current copy, stale code.
	mock.ExpectQuery(regexp.QuoteMeta(pageSurfaceQuery)).WithArgs(uuid.MustParse(pageID)).
		WillReturnRows(sqlmock.NewRows([]string{"surface"}).AddRow(probePageBug225))

	res := &siteRefreshResult{SiteID: fdSiteIDStr, Domain: "example.test"}
	eb := map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(500000)}}
	plan := planSiteFactDrift(context.Background(), db, fdSiteID, eb, res, true, zap.NewNop())

	if len(plan.Emissions) != 1 {
		t.Fatalf("expected one emission, got %d", len(plan.Emissions))
	}
	if plan.Emissions[0].Evidence != factProbeMarkupOnly {
		t.Fatalf("the emission must carry what the probe saw, got %q", plan.Emissions[0].Evidence)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the planner did not reach the probe's surface read: %v", err)
	}
}

// ANNOTATION ONLY — the probe must not move a route, a kind or a reason. If this
// ever fails, Phase 3a has quietly become Phase 3b without the council round
// that is supposed to authorise it on measured numbers.
func TestPlanSiteFactDrift_ProbeChangesNoRoutingDecision(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	body := "```criteria\n" + `{"facts":["sdlt-ftb-relief-cap"],"no_auto_fix":true,"checks":[]}` + "\n```"
	pageID := "3d7d0d72-0000-4000-8000-000000000001"
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}).
			AddRow(pageID, "tool-stamp-duty", "", "complete", "stamp-duty", body, ""))
	mock.ExpectQuery(regexp.QuoteMeta(factDriftLastItemQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subject_key", "new_value"}))
	mock.ExpectQuery(regexp.QuoteMeta(pageSurfaceQuery)).WithArgs(uuid.MustParse(pageID)).
		WillReturnRows(sqlmock.NewRows([]string{"surface"}).AddRow(probePageBug225))

	res := &siteRefreshResult{SiteID: fdSiteIDStr, Domain: "example.test"}
	eb := map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(500000)}}
	em := planSiteFactDrift(context.Background(), db, fdSiteID, eb, res, true, zap.NewNop()).Emissions[0]

	// Exactly what an unreconciled first declaration produced before Phase 3a.
	if em.Kind != "unreconciled_declaration" || em.Route != "fact_drift_review" || em.Reason != "never_reconciled" {
		t.Fatalf("the probe must not change routing: kind=%q route=%q reason=%q", em.Kind, em.Route, em.Reason)
	}
}

// A page the probe cannot read must not lose the finding. The measurement may
// never suppress the thing being measured.
func TestPlanSiteFactDrift_UnreadableSurfaceStillEmits(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	body := "```criteria\n" + `{"facts":["sdlt-ftb-relief-cap"],"no_auto_fix":true,"checks":[]}` + "\n```"
	pageID := "3d7d0d72-0000-4000-8000-000000000001"
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}).
			AddRow(pageID, "tool-stamp-duty", "", "complete", "stamp-duty", body, ""))
	mock.ExpectQuery(regexp.QuoteMeta(factDriftLastItemQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subject_key", "new_value"}))
	mock.ExpectQuery(regexp.QuoteMeta(pageSurfaceQuery)).WithArgs(uuid.MustParse(pageID)).
		WillReturnError(errSurfaceUnavailable)

	res := &siteRefreshResult{SiteID: fdSiteIDStr, Domain: "example.test"}
	eb := map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(500000)}}
	plan := planSiteFactDrift(context.Background(), db, fdSiteID, eb, res, true, zap.NewNop())

	if len(plan.Emissions) != 1 {
		t.Fatalf("a failed probe must not lose the finding, got %d emissions", len(plan.Emissions))
	}
	if plan.Emissions[0].Evidence != factProbeNoSurface {
		t.Fatalf("an unreadable page is no_surface, never absent, got %q", plan.Emissions[0].Evidence)
	}
}
