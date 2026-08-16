// FILE: platform/orchestration/actions/refresh_evidence_fact_drift_test.go
//
// Piece 3's tests. The induced reds are the point: each of the four routing
// decisions has a test that FAILS if the guard is removed, because a fan-out
// that cannot be shown to fire is the "armed but inert" shape this estate has
// been bitten by twice (016b §9).
//
// The fixtures are the real bugs_closed/225 shape, synthetic ids: a citation
// fact for the first-time-buyer relief cap, declared by a stamp-duty tool.

package actions

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var fdSiteID = uuid.MustParse("62b5978e-0000-4000-8000-000000000001")

const fdSiteIDStr = "62b5978e-0000-4000-8000-000000000001"

// sdltReliefCapFact is mortgagecalculator's real fact shape (GOV.UK citation,
// writer_line with {value}, unit GBP), value as float64 the way json.Unmarshal
// produces it.
func sdltReliefCapFact(value float64) map[string]interface{} {
	return map[string]interface{}{
		"id":    "sdlt-ftb-relief-cap",
		"kind":  "metric",
		"unit":  "GBP",
		"value": value,
		"claim": "First-time buyer relief cannot be claimed at all if the price is over £500,000",
		"source": map[string]interface{}{
			"citation": map[string]interface{}{
				"url":   "https://www.gov.uk/stamp-duty-land-tax/residential-property-rates",
				"quote": "If the price is over £500,000, you cannot claim the relief.",
			},
		},
		"verified_at": "2026-08-15",
		"writer_line": "Above £{value} first-time buyer relief disappears entirely",
	}
}

func fdTool(subjectKey string, fork bool, noAutoFix bool) factDriftTool {
	t := factDriftTool{
		SubjectKey:    subjectKey,
		PageID:        "3d7d0d72-0000-4000-8000-000000000001",
		PageName:      "tool-" + subjectKey,
		PageURL:       "https://example.test/tools/" + subjectKey + "/index.html",
		DeclaredFacts: []string{"sdlt-ftb-relief-cap"},
		Criteria:      `{"profiles":["desktop"],"facts":["sdlt-ftb-relief-cap"],"checks":[]}`,
		NoAutoFix:     noAutoFix,
	}
	if fork {
		t.ForkComponentID = "392e979d-0000-4000-8000-000000000001"
	}
	if noAutoFix {
		t.NoAutoFixReason = "arithmetic re-derived from the register; a human decides what may change"
	}
	return t
}

func f64(v float64) *float64 { return &v }

// ── Piece 2: the declaration parser ────────────────────────────────────────

func TestParseCriteriaFacts(t *testing.T) {
	cases := []struct {
		name       string
		criteria   string
		wantIDs    []string
		wantIssues int
	}{
		{"absent key", `{"profiles":["desktop"],"checks":[]}`, nil, 0},
		{"empty criteria fails open", ``, nil, 0},
		{"unparseable fails open", `{not json`, nil, 0},
		{"null facts", `{"facts":null}`, nil, 0},
		{"well formed", `{"facts":["a","b"]}`, []string{"a", "b"}, 0},
		{"trims", `{"facts":["  a  "]}`, []string{"a"}, 0},
		{"dedups and reports", `{"facts":["a","a"]}`, []string{"a"}, 1},
		{"non-string entry reported", `{"facts":["a",7]}`, []string{"a"}, 1},
		{"empty entry reported", `{"facts":["a",""]}`, []string{"a"}, 1},
		{"object form refused", `{"facts":{"id":"a"}}`, nil, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ids, issues := parseCriteriaFacts(c.criteria)
			if len(ids) != len(c.wantIDs) {
				t.Fatalf("ids = %v, want %v", ids, c.wantIDs)
			}
			for i := range ids {
				if ids[i] != c.wantIDs[i] {
					t.Fatalf("ids = %v, want %v", ids, c.wantIDs)
				}
			}
			if len(issues) != c.wantIssues {
				t.Fatalf("issues = %v, want %d", issues, c.wantIssues)
			}
		})
	}
}

// ── Piece 3: routing ───────────────────────────────────────────────────────

// INDUCED RED for the whole mechanism: the fact moved and a tool says it
// encodes it. Something must be owed. A fan-out that stays silent here is the
// bug this file exists to prevent.
func TestClassifyFactDrift_ValueMovedForkFilesImproveTool(t *testing.T) {
	em, ok := classifyFactDrift(nil, sdltReliefCapFact(550000), fdTool("stamp-duty", true, false), f64(500000), fdSiteIDStr)
	if !ok {
		t.Fatal("a moved value with a declaring tool must be owed something")
	}
	if em.Kind != "value_drift" {
		t.Fatalf("kind = %q, want value_drift", em.Kind)
	}
	if em.Route != "improve_tool" {
		t.Fatalf("a FORK with no no_auto_fix must route to improve_tool, got %q (reason %q)", em.Route, em.Reason)
	}
	if em.ItemKey != "fact_drift:sdlt-ftb-relief-cap:stamp-duty:"+fdSiteIDStr {
		t.Fatalf("item key = %q", em.ItemKey)
	}
	if em.OldValue == nil || *em.OldValue != 500000 || em.NewValue == nil || *em.NewValue != 550000 {
		t.Fatalf("old/new values not carried: %v → %v", em.OldValue, em.NewValue)
	}
	if em.ComponentID == "" {
		t.Error("a fork emission must carry the tool component id (seed 426 pre-flight needs page_id + component_id)")
	}
}

// INDUCED RED for the no_auto_fix guard (TL-040). Remove `case tool.NoAutoFix`
// from classifyFactDrift and this fails: an automated rewriter would be pointed
// at a calculator whose fence says a human decides.
func TestClassifyFactDrift_NoAutoFixRoutesToHuman(t *testing.T) {
	em, ok := classifyFactDrift(nil, sdltReliefCapFact(550000), fdTool("stamp-duty", true, true), f64(500000), fdSiteIDStr)
	if !ok {
		t.Fatal("expected an emission")
	}
	if em.Route != "fact_drift_review" || em.Reason != "no_auto_fix" {
		t.Fatalf("a no_auto_fix fence must route to a human: route %q reason %q", em.Route, em.Reason)
	}
	if em.ItemKey != "fact_drift_review:sdlt-ftb-relief-cap:stamp-duty:"+fdSiteIDStr {
		t.Fatalf("review item key = %q — a review item must NOT share the improve_tool key, or refreshOnConflict would rewrite one as the other", em.ItemKey)
	}
}

// INDUCED RED for the fork guard (bugs_open/281, TL-042). Remove
// `case !tool.isFork()` and this fails: tool-improver would be aimed at the
// SHARED wrapper component of a ported/decomposed page, which is how one
// finding rewrote 115 pages fleet-wide on 2026-08-05 and again on 08-14.
func TestClassifyFactDrift_NonForkRoutesToHuman(t *testing.T) {
	em, ok := classifyFactDrift(nil, sdltReliefCapFact(550000), fdTool("stamp-duty", false, false), f64(500000), fdSiteIDStr)
	if !ok {
		t.Fatal("expected an emission")
	}
	if em.Route != "fact_drift_review" || em.Reason != "not_a_fork" {
		t.Fatalf("a non-fork must route to a human: route %q reason %q", em.Route, em.Reason)
	}
	if em.ComponentID != "" {
		t.Error("a non-fork emission must NOT carry a component id — there is no fork component to point a fixer at")
	}
}

// INDUCED RED for the evidence/value split (plan §2 Piece 3). Route
// evidence_drift to improve_tool and this fails: a lost GOV.UK citation is not
// evidence the number moved, and an automated rewrite on that evidence is
// bugs_open/126 with arithmetic as the target.
func TestClassifyFactDrift_CitationLostIsAlwaysHuman(t *testing.T) {
	entry := &evidenceFactRefresh{FactID: "sdlt-ftb-relief-cap", Outcome: "drifted", Detail: "quote no longer found at source"}
	em, ok := classifyFactDrift(entry, sdltReliefCapFact(500000), fdTool("stamp-duty", true, false), f64(500000), fdSiteIDStr)
	if !ok {
		t.Fatal("expected an emission")
	}
	if em.Kind != "evidence_drift" {
		t.Fatalf("kind = %q, want evidence_drift", em.Kind)
	}
	if em.Route != "fact_drift_review" || em.Reason != "evidence_drift" {
		t.Fatalf("evidence drift must ALWAYS be human-routed, even for an auto-fixable fork: route %q", em.Route)
	}
}

// INDUCED RED for CLM-008: a fetch error must file nothing. Make the error case
// fall through to a route and this fails — every 403 day would file items and
// train people to ignore the queue.
func TestClassifyFactDrift_FetchErrorFilesNothing(t *testing.T) {
	entry := &evidenceFactRefresh{FactID: "sdlt-ftb-relief-cap", Outcome: "error", Detail: "403 from source"}
	em, ok := classifyFactDrift(entry, sdltReliefCapFact(500000), fdTool("stamp-duty", true, false), f64(500000), fdSiteIDStr)
	if !ok {
		t.Fatal("a fetch error should still be REPORTED (kind skipped_unknown), just not filed")
	}
	if em.Route != "none" || em.ItemKey != "" || em.Outcome != "none" {
		t.Fatalf("a fetch error must file nothing: route %q key %q outcome %q", em.Route, em.ItemKey, em.Outcome)
	}
}

// The quiet case: nothing moved, nothing drifted, nothing owed.
func TestClassifyFactDrift_SteadyStateIsSilent(t *testing.T) {
	entry := &evidenceFactRefresh{FactID: "sdlt-ftb-relief-cap", Outcome: "fresh"}
	if _, ok := classifyFactDrift(entry, sdltReliefCapFact(500000), fdTool("stamp-duty", true, false), f64(500000), fdSiteIDStr); ok {
		t.Fatal("an unchanged fact on a fresh citation must file nothing")
	}
}

// INDUCED RED for the gap the council found (editquality, high, 2026-08-16), and
// the most important test in this file. The FIRST version of this code returned
// silently when there was no baseline, reasoning that "is the current number
// right" is Piece 4's question. That reasoning was correct and the behaviour was
// still wrong: a tool that is stale on the day it opts in, against a register
// fact that has not moved since, has no baseline and no drift — so the mechanism
// built for bugs_closed/225 would have been silent on bugs_closed/225's own
// shape. A first declaration must ask a human to reconcile the pair once.
func TestClassifyFactDrift_FirstDeclarationAsksForReconciliation(t *testing.T) {
	em, ok := classifyFactDrift(nil, sdltReliefCapFact(500000), fdTool("stamp-duty", true, false), nil, fdSiteIDStr)
	if !ok {
		t.Fatal("a first-time declaration must not be silent — that is exactly the 225 shape (correct register, stale code, no subsequent change)")
	}
	if em.Kind != "unreconciled_declaration" || em.Reason != "never_reconciled" {
		t.Fatalf("kind/reason = %q/%q, want unreconciled_declaration/never_reconciled", em.Kind, em.Reason)
	}
	if em.Route != "fact_drift_review" {
		t.Fatalf("a reconciliation request is a human's, never an auto-fix: route %q", em.Route)
	}
	if em.NewValue == nil || *em.NewValue != 500000 {
		t.Fatal("the item must carry the register's current value — it becomes the baseline that makes this self-quieting")
	}
	if em.OldValue != nil {
		t.Error("there is no old value to report on a first declaration; claiming one would invent a move")
	}
}

// The other half of the same design: having asked once, it must not ask again.
// The item filed above carries new_value, which becomes the lastItem baseline —
// so the next pass sees baseline == current and says nothing. Without this, the
// fix for the objection above would file a review item every single day for
// every declared fact on the fleet.
func TestClassifyFactDrift_ReconciliationIsSelfQuieting(t *testing.T) {
	if _, ok := classifyFactDrift(nil, sdltReliefCapFact(500000), fdTool("stamp-duty", true, false), f64(500000), fdSiteIDStr); ok {
		t.Fatal("once reconciled at this value, the pair must go quiet — a detector that repeats daily is noise, not a finding")
	}
}

// A fact with no numeric value at all (a capability or entity fact someone
// declared) has nothing to reconcile and must stay silent — the guard is
// `hasVal`, not just the baseline.
func TestClassifyFactDrift_ValuelessFactIsSilent(t *testing.T) {
	fact := sdltReliefCapFact(0)
	delete(fact, "value")
	if _, ok := classifyFactDrift(nil, fact, fdTool("stamp-duty", true, false), nil, fdSiteIDStr); ok {
		t.Fatal("a fact with no numeric value has nothing to reconcile")
	}
}

// ── Baseline precedence ────────────────────────────────────────────────────

// INDUCED RED for the re-fire guard: the last item the tool was told about wins
// over the previous register row. Drop the lastItem lookup and a fact whose
// item was already filed re-fires every single day.
func TestFactDriftBaselines_LastItemBeatsPreviousRow(t *testing.T) {
	b := factDriftBaselines{
		lastItem:    map[string]float64{baselineKey("sdlt-ftb-relief-cap", "stamp-duty"): 550000},
		previousRow: map[string]float64{"sdlt-ftb-relief-cap": 500000},
	}
	got := b.baselineFor("sdlt-ftb-relief-cap", "stamp-duty")
	if got == nil || *got != 550000 {
		t.Fatalf("baseline = %v, want the value the tool was last TOLD (550000)", got)
	}
	// A different tool has been told nothing, so it falls back to the register.
	other := b.baselineFor("sdlt-ftb-relief-cap", "other-tool")
	if other == nil || *other != 500000 {
		t.Fatalf("fallback baseline = %v, want the previous register row (500000)", other)
	}
	if b.baselineFor("unknown-fact", "stamp-duty") != nil {
		t.Fatal("an unknown fact must have no baseline")
	}
}

// ── The site-level planner, over the DB ────────────────────────────────────

func TestLoadFactDriftIndex_ResolvesDeclarationsAndReportsUnknownIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	body := "# PLAN\n\n```criteria\n{\"facts\":[\"sdlt-ftb-relief-cap\",\"no-such-fact\"],\"no_auto_fix\":true,\"checks\":[]}\n```\n"
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}).
			AddRow("3d7d0d72-0000-4000-8000-000000000001", "tool-stamp-duty", "https://x/tools/stamp-duty/", "complete", "stamp-duty", body, ""))

	idx, err := loadFactDriftIndex(context.Background(), db, fdSiteID, map[string]bool{"sdlt-ftb-relief-cap": true})
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if idx == nil {
		t.Fatal("a declaring PLAN must produce an index")
	}
	if len(idx.byFact["sdlt-ftb-relief-cap"]) != 1 {
		t.Fatalf("declared fact not indexed: %+v", idx.byFact)
	}
	tool := idx.byFact["sdlt-ftb-relief-cap"][0]
	if !tool.NoAutoFix {
		t.Error("the fence's no_auto_fix must be carried onto the tool")
	}
	if tool.isFork() {
		t.Error("a page with no tool-level component must not read as a fork")
	}
	if len(idx.unresolved) != 1 || idx.unresolved[0] != "stamp-duty:no-such-fact" {
		t.Fatalf("an id the register does not carry must be reported inert, got %v", idx.unresolved)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// The NO-OP path: a site whose PLANs declare nothing costs exactly one SELECT,
// writes nothing, and leaves the result JSON byte-identical to before Piece 3.
func TestPlanSiteFactDrift_NoDeclarationsIsOneQueryAndNoResultChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}))

	res := &siteRefreshResult{SiteID: fdSiteIDStr, Domain: "example.test", WriterBlock: "unchanged"}
	before, _ := json.Marshal(res)

	eb := map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(500000)}}
	plan := planSiteFactDrift(context.Background(), db, fdSiteID, uuid.New(), eb, res, false, zap.NewNop())

	if len(plan.Emissions) != 0 {
		t.Fatalf("a site with no declarations must emit nothing, got %d", len(plan.Emissions))
	}
	res.FactDrift = plan.Emissions
	after, _ := json.Marshal(res)
	if string(before) != string(after) {
		t.Fatalf("no-op result JSON changed:\n before %s\n after  %s", before, after)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// A dry run PLANS the fan-out and marks it as such, so the induced live proof
// can be read before anything is written.
func TestPlanSiteFactDrift_DryRunPlansButMarksDryRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	body := "```criteria\n{\"facts\":[\"sdlt-ftb-relief-cap\"],\"no_auto_fix\":true,\"checks\":[]}\n```"
	specRow := uuid.New()
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}).
			AddRow("3d7d0d72-0000-4000-8000-000000000001", "tool-stamp-duty", "", "complete", "stamp-duty", body, "392e979d-0000-4000-8000-000000000001"))
	mock.ExpectQuery(regexp.QuoteMeta(factDriftLastItemQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subject_key", "new_value"}))
	prev, _ := json.Marshal(map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(500000)}})
	mock.ExpectQuery(regexp.QuoteMeta(factDriftPreviousRowQuery)).WithArgs(fdSiteID, specRow).
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow(prev))

	res := &siteRefreshResult{SiteID: fdSiteIDStr, Domain: "example.test"}
	eb := map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(550000)}}
	plan := planSiteFactDrift(context.Background(), db, fdSiteID, specRow, eb, res, true, zap.NewNop())

	if len(plan.Emissions) != 1 {
		t.Fatalf("expected one emission, got %d", len(plan.Emissions))
	}
	em := plan.Emissions[0]
	if em.Outcome != "dry_run" {
		t.Fatalf("a dry run must mark its emissions dry_run, got %q", em.Outcome)
	}
	if em.SubjectKey != "stamp-duty" || em.Kind != "value_drift" || em.Route != "fact_drift_review" {
		t.Fatalf("dry run did not name the tool correctly: %+v", em)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// The stamped EncodedByTools is what makes the EXISTING per-site
// stale_evidence item name the tools, without changing when it is raised.
func TestPlanFactDriftFanOut_StampsEncodedByToolsOnDriftedEntries(t *testing.T) {
	res := &siteRefreshResult{
		SiteID: fdSiteIDStr,
		Facts:  []evidenceFactRefresh{{FactID: "sdlt-ftb-relief-cap", Outcome: "drifted", Detail: "citation lost"}},
	}
	idx := &factDriftIndex{byFact: map[string][]factDriftTool{
		"sdlt-ftb-relief-cap": {fdTool("stamp-duty", true, true)},
	}}
	facts := map[string]map[string]interface{}{"sdlt-ftb-relief-cap": sdltReliefCapFact(500000)}
	_ = planFactDriftFanOut(res, facts, idx, factDriftBaselines{}, fdSiteIDStr)

	if len(res.Facts[0].EncodedByTools) != 1 || res.Facts[0].EncodedByTools[0] != "stamp-duty" {
		t.Fatalf("a drifted fact must name the tools that encode it, got %v", res.Facts[0].EncodedByTools)
	}
}

// The issue text is the payload tool-improver actually reads
// ({{.input_data.issue}}), so its contents are asserted, not assumed.
func TestFactDriftIssueText_CarriesEverythingTheFixerNeeds(t *testing.T) {
	fact := sdltReliefCapFact(550000)
	em, _ := classifyFactDrift(nil, fact, fdTool("stamp-duty", true, false), f64(500000), fdSiteIDStr)
	txt := factDriftIssueText(fact, em, fdTool("stamp-duty", true, false))
	for _, want := range []string{"sdlt-ftb-relief-cap", "500,000", "550,000", "GBP", "gov.uk", "Above £550,000"} {
		if !strings.Contains(txt, want) {
			t.Errorf("issue text missing %q:\n%s", want, txt)
		}
	}
}

// The council's compliance seat caught a risk inversion in the first version:
// every human-routed finding sat at medium/35, so a CONFIRMED value move on a
// no_auto_fix tax calculator — the exact case this mechanism exists for — ranked
// BELOW an auto-fixable drift at high/30. Severity now tracks the finding, not
// the route.
func TestFactDriftSeverity_AConfirmedMoveOutranksAnEvidenceQuestion(t *testing.T) {
	moved, _ := classifyFactDrift(nil, sdltReliefCapFact(550000), fdTool("stamp-duty", true, true), f64(500000), fdSiteIDStr)
	if moved.Kind != "value_drift" || moved.Route != "fact_drift_review" {
		t.Fatalf("fixture wrong: %+v", moved)
	}
	firstTime, _ := classifyFactDrift(nil, sdltReliefCapFact(500000), fdTool("stamp-duty", true, true), nil, fdSiteIDStr)
	if firstTime.Kind != "unreconciled_declaration" {
		t.Fatalf("fixture wrong: %+v", firstTime)
	}
	// The banding itself lives in writeFactDriftItems; assert the discriminator
	// it keys on is present and distinct, which is what makes the banding possible.
	if moved.Kind == firstTime.Kind {
		t.Fatal("a moved number and an unreconciled declaration must be distinguishable by Kind, or they cannot be banded differently")
	}
}
