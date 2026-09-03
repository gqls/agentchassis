// FILE: platform/orchestration/actions/recommended_type_reconciliation_test.go
//
// bugs_open/428 — the strategy-to-plan reconciliation.
//
// THE NEGATIVES ARE THE POINT OF THIS FILE, and they are why
// reconcileRecommendedPageTypes returns a result at all. Two of its rules are
// rules about NOT acting — a gate hold files no second work item, and a deferral
// to a producer that is genuinely running files none either — and a test that
// proves those by simply omitting a mock expectation proves nothing: an
// unexpected BeginTx makes sqlmock return an error, the file logs it and carries
// on by design, and the test goes green whether the rule held or the code never
// reached it. So every negative below asserts on the returned struct (the
// omission was SEEN, classified, and deliberately not filed), not on the absence
// of a database call.
//
// MUTATION PROTOCOL — RUN 2026-09-03, results as measured, not as predicted.
// Each arm kills a different set:
//
//   - `actionable()` returns true unconditionally → kills exactly
//     TestHeldByGateIsRecordedButFilesNoSecondItem and
//     TestDeferralToLiveProducerFilesNoItem. Every other test stays green, so
//     the two silences are pinned separately from everything else.
//   - `producerLiveness.Live()` returns true unconditionally → kills
//     TestPlannerOmittedWithDormantProducerFilesGap (its warning becomes an info
//     and its item disappears) AND both liveness tests, which assert that an
//     unseen or unreadable producer is not live. ⚠ I predicted one failure and
//     measured three; the prediction is corrected here rather than the tests
//     loosened, because the extra two are the honest ones. What matters is the
//     other direction: TestDeferralToLiveProducerFilesNoItem stays GREEN under
//     this mutation, which is what makes that test and the dormant-producer test
//     a discriminating pair rather than two assertions about one branch.
//   - the `proposed`-snapshot arm never fires (`case false:`, so everything not
//     held by a gate reads as a planner omission) → kills exactly
//     TestPlannedThenDeletedIsDistinguishedFromNeverPlanned. This is the arm
//     encoding the gamedesign.uk measurement, so it is the one whose mutation
//     must not be silent.
//     ⚠ THE FIRST FORM OF THIS MUTATION PROVED NOTHING AND LOOKED LIKE A KILL:
//     deleting the `case` outright left omissionDroppedInValidation unreferenced
//     and the package stopped compiling, so `go test` reported FAIL for the
//     whole package — a build error wearing a passing mutation's clothes. A
//     mutant that does not build has not been tested. Check for
//     "[build failed]" before reading a red result as a kill.
//   - the section-index family check is removed → kills exactly
//     TestSectionIndexSiblingIsASubstitutionNotAnOmission, so the substitution
//     rule cannot be quietly widened into a general "some other page is present"
//     escape.
//
// TestAllRecommendedTypesPresentStillWritesTheAuditRow is the control that stops
// the whole file passing for the wrong reason: it pins that a CLEAN run is
// distinguishable from a run that never happened, which is the failure mode the
// file under test exists to remove.
package actions

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// --- harness ---------------------------------------------------------------

func recoParams(t *testing.T, siteID uuid.UUID, recommended []map[string]interface{}) ActionParams {
	t.Helper()
	specs := map[string]interface{}{}
	if recommended != nil {
		entries := make([]interface{}, 0, len(recommended))
		for _, r := range recommended {
			entries = append(entries, r)
		}
		specs["strategy"] = map[string]interface{}{"recommended_page_types": entries}
	}
	return ActionParams{
		CollectedData: map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"site_specs":  map[string]interface{}{"specs": specs},
		},
		Logger: zap.NewNop(),
	}
}

func recoTypes(types ...string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(types))
	for _, t := range types {
		out = append(out, map[string]interface{}{"page_type": t, "reasoning": "because the strategy says so: " + t})
	}
	return out
}

func recoPages(pairs ...string) []planPageView {
	views := make([]planPageView, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		views = append(views, planPageView{Name: pairs[i], Role: pairs[i+1]})
	}
	return views
}

// expectFindings registers exactly n durable finding writes.
func expectFindings(mock sqlmock.Sqlmock, n int) {
	for i := 0; i < n; i++ {
		mock.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(0, 1))
	}
}

// expectLiveness registers one agent_run_stats probe.
// hasRow=false is the "no row since tracking began" shape.
func expectLiveness(mock sqlmock.Sqlmock, hasRow bool, lastRan time.Time) {
	cols := []string{"run_count", "last_ran_at", "tracking_since"}
	trackingSince := time.Now().Add(-40 * 24 * time.Hour)
	if hasRow {
		mock.ExpectQuery("FROM agent_run_stats").
			WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(7), lastRan, trackingSince))
		return
	}
	mock.ExpectQuery("FROM agent_run_stats").
		WillReturnRows(sqlmock.NewRows(cols).AddRow(nil, nil, trackingSince))
}

// expectGapItem registers the parked capability_gap write through the SHARED
// writer, pinning the three arguments that carry this row's whole meaning:
// $4 item_type, $12 status, $14 item_key. A loose "INSERT INTO site_work_items"
// matcher would pass for a dispatchable row of any type under any key, which is
// precisely the mistake this shape exists to prevent (bugs_closed/078/291).
func expectGapItem(mock sqlmock.Sqlmock, pageType string) {
	args := make([]driver.Value, 16)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[3] = "capability_gap"
	args[11] = "deferred"
	args[13] = "recommended_type_gap:" + pageType
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

// --- the three classes -----------------------------------------------------

// The gamedesign.uk case, 2026-09-03: the planner DID plan blog-post pages and
// this action deleted them. Before this file, that produced exactly the same
// evidence as a planner that never proposed them — no finding, no work item, no
// error, and a healthy-looking page count.
func TestPlannedThenDeletedIsDistinguishedFromNeverPlanned(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	params := recoParams(t, siteID, recoTypes("index", "blog-post"))
	params.DB = sqlDB

	// blog-post has a registered producer, so liveness is consulted; it has not
	// run since tracking began, which is blog-content-planner's real position.
	expectLiveness(mock, false, time.Time{})
	expectFindings(mock, 2) // the audit row + one omission
	expectGapItem(mock, "blog-post")

	proposed := recoPages("index", "index", "article-one", "blog-post", "article-two", "blog-post")
	preGate := recoPages("index", "index") // Pass C already ate them
	final := recoPages("index", "index")

	res := reconcileRecommendedPageTypes(context.Background(), params,
		map[string]interface{}{"strategy_notes": ""}, proposed, preGate, final)

	if res.Skipped != "" {
		t.Fatalf("reconciliation skipped (%s) — it must run", res.Skipped)
	}
	if len(res.Omissions) != 1 {
		t.Fatalf("omissions = %d, want 1 (blog-post)", len(res.Omissions))
	}
	om := res.Omissions[0]
	if om.PageType != "blog-post" {
		t.Fatalf("omitted type = %q, want blog-post", om.PageType)
	}
	if om.Class != omissionDroppedInValidation {
		t.Fatalf("class = %q, want %q — the planner PROPOSED this type; classifying it as a planner omission blames the wrong stage",
			om.Class, omissionDroppedInValidation)
	}
	if om.builderNeeded() != "plan_page_identity" {
		t.Fatalf("builder_needed = %q, want plan_page_identity — a page the planner planned and validation deleted does not need a builder",
			om.builderNeeded())
	}
	if got := om.code(); got != findingRecommendedTypeDropped {
		t.Fatalf("code = %q, want %q", got, findingRecommendedTypeDropped)
	}
	if len(om.ProposedPages) != 2 {
		t.Fatalf("proposed_pages = %v, want the two article pages that carried the type", om.ProposedPages)
	}
	if res.GapsFiled != 1 {
		t.Fatalf("gaps filed = %d, want 1", res.GapsFiled)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected the liveness probe, both findings and the parked gap: %v", err)
	}
}

// A type the planner never proposed, deferred to a producer that is not running,
// is bugs_open/428's own residual: 687's "name your reason" obligation met with a
// reason that is false as applied.
func TestPlannerOmittedWithDormantProducerFilesGap(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	params := recoParams(t, siteID, recoTypes("index", "blog-post"))
	params.DB = sqlDB

	// Registered producer, has a row, but last ran far outside the window.
	expectLiveness(mock, true, time.Now().Add(-120*24*time.Hour))
	expectFindings(mock, 2)
	expectGapItem(mock, "blog-post")

	plan := map[string]interface{}{
		"strategy_notes": "The blog-post type is satisfied by the blog infrastructure; individual posts are not planned as static pages here.",
		"page_type_decisions": []interface{}{
			map[string]interface{}{
				"page_type":   "blog-post",
				"decision":    "deferred",
				"reason":      "satisfied by the blog infrastructure",
				"deferred_to": "blog-content-planner",
			},
		},
	}
	pages := recoPages("index", "index")

	res := reconcileRecommendedPageTypes(context.Background(), params, plan, pages, pages, pages)

	if len(res.Omissions) != 1 {
		t.Fatalf("omissions = %d, want 1", len(res.Omissions))
	}
	om := res.Omissions[0]
	if om.Class != omissionPlannerOmitted {
		t.Fatalf("class = %q, want %q", om.Class, omissionPlannerOmitted)
	}
	if om.ClaimedTo != "blog-content-planner" {
		t.Fatalf("claimed producer = %q — the structured decision names it and the check must read it, since that claim is the thing under test", om.ClaimedTo)
	}
	if om.Liveness.Verdict != "dormant" {
		t.Fatalf("liveness verdict = %q, want dormant", om.Liveness.Verdict)
	}
	if om.code() != findingRecommendedTypePlannerOmitted || om.severity() != "warning" {
		t.Fatalf("code/severity = %s/%s, want %s/warning", om.code(), om.severity(), findingRecommendedTypePlannerOmitted)
	}
	if res.GapsFiled != 1 {
		t.Fatalf("gaps filed = %d, want 1", res.GapsFiled)
	}
	if om.ReasonSource != "page_type_decisions" {
		t.Fatalf("reason source = %q, want page_type_decisions", om.ReasonSource)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

// THE DISCRIMINATING CONTROL. Same shape as the test above in every respect
// except that the named producer is running — and a deferral to a running
// producer is a sound decision, not a defect. Without this, the check above
// would fire on every deferral and be worth nothing.
func TestDeferralToLiveProducerFilesNoItem(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	params := recoParams(t, siteID, recoTypes("index", "blog-post"))
	params.DB = sqlDB

	expectLiveness(mock, true, time.Now().Add(-2*time.Hour))
	expectFindings(mock, 2)
	// Deliberately NO gap-item expectation: the assertion that nothing was filed
	// is made on the result, below, not on this absence.

	plan := map[string]interface{}{
		"page_type_decisions": []interface{}{
			map[string]interface{}{"page_type": "blog-post", "decision": "deferred", "deferred_to": "blog-content-planner"},
		},
	}
	pages := recoPages("index", "index")

	res := reconcileRecommendedPageTypes(context.Background(), params, plan, pages, pages, pages)

	if len(res.Omissions) != 1 {
		t.Fatalf("omissions = %d, want 1 — the type IS absent and must still be seen and recorded", len(res.Omissions))
	}
	om := res.Omissions[0]
	if !om.Liveness.Live() {
		t.Fatalf("liveness verdict = %q, want live", om.Liveness.Verdict)
	}
	if om.actionable() {
		t.Fatal("a deferral to a RUNNING producer must not be actionable — the pages can still arrive by that route")
	}
	if om.code() != findingRecommendedTypeDeferredLive || om.severity() != "info" {
		t.Fatalf("code/severity = %s/%s, want %s/info", om.code(), om.severity(), findingRecommendedTypeDeferredLive)
	}
	if res.GapsFiled != 0 {
		t.Fatalf("gaps filed = %d, want 0", res.GapsFiled)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the liveness probe and both findings must still be written: %v", err)
	}
}

// A gate that held the page has already filed a capability_gap naming the
// enablement. Filing a second row for the same fact is the duplicate-producer
// shape this estate keeps paying for, so this one records and stops.
func TestHeldByGateIsRecordedButFilesNoSecondItem(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	params := recoParams(t, siteID, recoTypes("index", "entity-directory"))
	params.DB = sqlDB
	expectFindings(mock, 2)

	proposed := recoPages("index", "index", "brands", "entity-directory")
	preGate := recoPages("index", "index", "brands", "entity-directory") // survived to the gate
	final := recoPages("index", "index")                                 // the gate held it

	res := reconcileRecommendedPageTypes(context.Background(), params,
		map[string]interface{}{}, proposed, preGate, final)

	if len(res.Omissions) != 1 {
		t.Fatalf("omissions = %d, want 1", len(res.Omissions))
	}
	om := res.Omissions[0]
	if om.Class != omissionHeldByGate {
		t.Fatalf("class = %q, want %q — a page present before the gates and absent after was held, not dropped", om.Class, omissionHeldByGate)
	}
	if om.actionable() {
		t.Fatal("a gate hold must not file a second work item — the gate filed one")
	}
	if om.severity() != "info" {
		t.Fatalf("severity = %q, want info", om.severity())
	}
	if res.GapsFiled != 0 {
		t.Fatalf("gaps filed = %d, want 0", res.GapsFiled)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

// --- the substitution rule and the clean control ---------------------------

// gamedesign.uk planned a section-index "articles-index" where its strategy
// recommended blog-index, and said so. Those are one family by the estate's ONE
// definition of it, so this is a substitution and not an omission — a check that
// reported it would be crying wolf on the very site it was built for.
func TestSectionIndexSiblingIsASubstitutionNotAnOmission(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	params := recoParams(t, siteID, recoTypes("index", "blog-index"))
	params.DB = sqlDB
	expectFindings(mock, 1) // the audit row only

	pages := recoPages("index", "index", "articles-index", "section-index")
	res := reconcileRecommendedPageTypes(context.Background(), params, map[string]interface{}{}, pages, pages, pages)

	if len(res.Omissions) != 0 {
		t.Fatalf("omissions = %v, want none — a section-index serves the blog-index recommendation", res.Omissions)
	}
	if len(res.Substitutions) != 1 || res.Substitutions[0] != "blog-index" {
		t.Fatalf("substitutions = %v, want [blog-index] recorded rather than silently ignored", res.Substitutions)
	}
	if res.GapsFiled != 0 {
		t.Fatalf("gaps filed = %d, want 0", res.GapsFiled)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

// THE CONTROL FOR THE WHOLE FILE: a clean run must leave a mark. Without the
// always-written audit row, "every recommended type is present" and "the
// reconciliation never ran" are the same evidence — which is the exact failure
// this file exists to remove, reproduced one level up.
func TestAllRecommendedTypesPresentStillWritesTheAuditRow(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	params := recoParams(t, siteID, recoTypes("index", "content"))
	params.DB = sqlDB
	expectFindings(mock, 1)

	pages := recoPages("index", "index", "about", "content")
	res := reconcileRecommendedPageTypes(context.Background(), params, map[string]interface{}{}, pages, pages, pages)

	if res.Skipped != "" {
		t.Fatalf("skipped %q — a clean site must still reconcile", res.Skipped)
	}
	if len(res.Omissions) != 0 {
		t.Fatalf("omissions = %v, want none", res.Omissions)
	}
	if res.FindingsAttempted != 1 || res.FindingsRecorded != 1 {
		t.Fatalf("findings attempted/recorded = %d/%d, want 1/1 — the audit row is what makes a clean run visible",
			res.FindingsAttempted, res.FindingsRecorded)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

// --- fail-open arms --------------------------------------------------------

func TestReconciliationFailsOpen(t *testing.T) {
	siteID := uuid.New()

	t.Run("nil DB", func(t *testing.T) {
		params := recoParams(t, siteID, recoTypes("blog-post"))
		res := reconcileRecommendedPageTypes(context.Background(), params, map[string]interface{}{}, nil, nil, nil)
		if res.Skipped != "no_db" {
			t.Fatalf("skipped = %q, want no_db", res.Skipped)
		}
	})

	t.Run("kill switch", func(t *testing.T) {
		t.Setenv(recommendedTypeReconciliationKillSwitch, "1")
		sqlDB, _, _ := sqlmock.New()
		defer sqlDB.Close()
		params := recoParams(t, siteID, recoTypes("blog-post"))
		params.DB = sqlDB
		res := reconcileRecommendedPageTypes(context.Background(), params, map[string]interface{}{}, nil, nil, nil)
		if res.Skipped != "kill_switch" {
			t.Fatalf("skipped = %q, want kill_switch — the lever must work with no build", res.Skipped)
		}
	})

	t.Run("site with no strategy writes nothing", func(t *testing.T) {
		sqlDB, mock, _ := sqlmock.New()
		defer sqlDB.Close()
		params := recoParams(t, siteID, nil)
		params.DB = sqlDB
		res := reconcileRecommendedPageTypes(context.Background(), params, map[string]interface{}{}, nil, nil, nil)
		if res.Skipped != "no_recommendations" {
			t.Fatalf("skipped = %q, want no_recommendations", res.Skipped)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("a site with no strategy must touch the database not at all: %v", err)
		}
	})
}

// --- liveness honesty ------------------------------------------------------

// agent_run_stats is FORWARD-ONLY: it began counting on 2026-08-02 and
// blog-content-planner's real last run was 2026-04-24, so "no row" cannot
// support "never ran". The verdict is named for what the instrument can actually
// see, and the tracking window travels with it so no reader can quietly upgrade
// the claim.
func TestLivenessNeverSinceTrackingCarriesItsWindow(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	since := time.Now().Add(-32 * 24 * time.Hour)
	mock.ExpectQuery("FROM agent_run_stats").
		WillReturnRows(sqlmock.NewRows([]string{"run_count", "last_ran_at", "tracking_since"}).
			AddRow(nil, nil, since))

	l := readProducerLiveness(context.Background(), sqlDB, "blog-content-planner", zap.NewNop())

	if l.Verdict != "never_since_tracking" {
		t.Fatalf("verdict = %q, want never_since_tracking (NOT 'never ran' — the table cannot say that)", l.Verdict)
	}
	if l.Live() {
		t.Fatal("an unseen producer must never count as live")
	}
	if l.TrackingSince.IsZero() {
		t.Fatal("the tracking window must travel with the verdict, or a forward-only zero reads as an all-history absence")
	}

	om := recommendedTypeOmission{PageType: "blog-post", Class: omissionPlannerOmitted, Liveness: l}
	if _, ok := om.context()["liveness_tracking_since"]; !ok {
		t.Fatal("the durable row must carry liveness_tracking_since")
	}
}

func TestLivenessUnreadableIsNotLive(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.ExpectQuery("FROM agent_run_stats").WillReturnError(context.DeadlineExceeded)

	l := readProducerLiveness(context.Background(), sqlDB, "blog-content-planner", zap.NewNop())
	if l.Verdict != "unreadable" || l.Live() {
		t.Fatalf("verdict = %q, live = %v — an unreadable instrument must count AGAINST the deferral, not for it",
			l.Verdict, l.Live())
	}
}

// --- input tolerance and the record's shape --------------------------------

func TestRecommendedTypesFromToleratesBothShapes(t *testing.T) {
	collected := map[string]interface{}{
		"site_specs": map[string]interface{}{"specs": map[string]interface{}{"strategy": map[string]interface{}{
			"recommended_page_types": []interface{}{
				map[string]interface{}{"page_type": "entity_page", "reasoning": "fighter profiles"},
				"blog-post",
				map[string]interface{}{"page_type": " Entity_Page "}, // duplicate after canonicalisation
				map[string]interface{}{"reasoning": "no type at all"},
			},
		}}},
	}
	got := recommendedTypesFrom(collected)
	if len(got) != 2 {
		t.Fatalf("got %v, want exactly entity-page and blog-post (underscore canonicalised, duplicate and typeless dropped)", got)
	}
	if got[0].PageType != "entity-page" || got[1].PageType != "blog-post" {
		t.Fatalf("got %v, want [entity-page blog-post]", got)
	}
	if got[0].Reasoning != "fighter profiles" {
		t.Fatalf("reasoning lost: %q", got[0].Reasoning)
	}
}

// The parked row must be un-releasable BY CONSTRUCTION, not by intention:
// HandleReleaseRecordVerdict's WHERE clause requires a non-empty
// spec.routed_handler AND spec.routed_status, and there is no automatic repair
// for "your strategy asked for a page type you have none of". A row that carried
// a route would offer a button that dispatches something nobody chose.
func TestParkedGapCarriesRecordModeAndNoRoute(t *testing.T) {
	om := recommendedTypeOmission{
		PageType:  "blog-post",
		Class:     omissionPlannerOmitted,
		Reasoning: strings.Repeat("x", 900),
		Liveness:  producerLiveness{Verdict: "never_since_tracking"},
	}
	spec := map[string]interface{}{
		"gap_kind":       "recommended_page_type_absent",
		"page_type":      om.PageType,
		"omission_class": string(om.Class),
		"filing_mode":    "record",
	}
	for k, v := range om.context() {
		if _, taken := spec[k]; !taken {
			spec[k] = v
		}
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back map[string]interface{}
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back["filing_mode"] != "record" {
		t.Fatalf("filing_mode = %v, want record (RFC_056) so no promoter can dispatch it", back["filing_mode"])
	}
	if _, ok := back["routed_handler"]; ok {
		t.Fatal("routed_handler must be ABSENT — a release button that dispatches an unchosen repair is the bugs_closed/238 shape")
	}
	if _, ok := back["routed_status"]; ok {
		t.Fatal("routed_status must be ABSENT for the same reason")
	}
	if r, _ := back["strategy_reasoning"].(string); len([]rune(r)) > 420 {
		t.Fatalf("strategy_reasoning not bounded (%d runes) — the whole of it already lives in site_specs", len([]rune(r)))
	}
}

// THE STATED BLIND SPOT, pinned so it stays stated. This check is TYPE-level: a
// recommended type whose pages were partly dropped is still PRESENT, so it is
// not an omission and must not be flagged — the preserve/union machinery
// reshapes page sets on every re-plan and a per-page rule would be noise. But it
// must not vanish either: this is the population the check cannot see, and it is
// why bugs_open/463's defect hid from 2026-05-21 (an established site restores
// its children via Pass A and looks healthy). Counted in the audit row so a
// fleet census can find it.
func TestTypePresentButFewerPagesIsCountedNotFlagged(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()
	params := recoParams(t, siteID, recoTypes("index", "blog-post"))
	params.DB = sqlDB
	expectFindings(mock, 1) // audit row only — no omission

	proposed := recoPages("index", "index",
		"post-a", "blog-post", "post-b", "blog-post", "post-c", "blog-post")
	final := recoPages("index", "index", "post-a", "blog-post") // two children lost, type survives

	res := reconcileRecommendedPageTypes(context.Background(), params, map[string]interface{}{}, proposed, final, final)

	if len(res.Omissions) != 0 {
		t.Fatalf("omissions = %v, want none — the type IS in the plan", res.Omissions)
	}
	if res.GapsFiled != 0 {
		t.Fatalf("gaps filed = %d, want 0", res.GapsFiled)
	}
	got, ok := res.PresentButFewer["blog-post"].(map[string]interface{})
	if !ok {
		t.Fatalf("present_but_fewer = %v, want blog-post recorded — an admitted blind spot that is not counted is an unmeasurable one", res.PresentButFewer)
	}
	if got["proposed"] != 3 || got["final"] != 1 {
		t.Fatalf("recorded %v, want proposed 3 / final 1", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}
