// FILE: platform/orchestration/actions/criteria_facts_declaration_gate_test.go
//
// bugs_open/288 Phase 1 — a `facts` declaration that cannot be read must be
// LOUD, at the door where it is written and again at the sweep that reads it.
//
// The two defects these tests pin were both live on 2026-08-24 and both were
// recorded in the bug file, the lane PLAN and the council submission as ALREADY
// HANDLED:
//
//   A. "validator rule P11 refuses a malformed declaration where it is written."
//      P11 lives in ValidateExperienceCriteria, whose only production caller is
//      write_experience_pattern_action.go — the EXPERIENCE-PATTERN register.
//      Tool PLANs are written by WriteDocPlanAction, which validated nothing.
//   B. parseCriteriaFacts returned (nil, nil) on a whole-fence JSON error — no
//      ids AND no issues — contradicting its own file header, and disarming the
//      zero-rows warning two rungs down, which is gated on issues being
//      non-empty. One trailing comma switched the mechanism off with no signal.
//
// Each test names the mutation that must turn it red. Per this lane's own
// three-mutation lesson (WRONG_CALLS, 2026-08-16): an induced red is evidence
// only when the mutation IS the defect, so the assertions are on ISSUES being
// reported, never merely on "no ids came back" — the pre-fix code also returned
// no ids, and a test satisfied by that is satisfied by the bug.

package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// ── 1a: the parser stops failing open on a fence that DID say something ─────

func TestParseCriteriaFacts_UnreadableFenceMentioningFactsIsReported(t *testing.T) {
	// The real shape: someone adds a check and leaves a trailing comma. The
	// declaration is right there in the text and the whole fence will not parse.
	const trailingComma = `{"profiles":["desktop"],"facts":["sdlt-ftb-relief-cap"],"checks":[],}`

	ids, issues := parseCriteriaFacts(trailingComma)

	// MUTATION THAT MUST GO RED: delete the factsKeyMentioned branch in
	// parseCriteriaFacts. Asserting on issues, not on ids, is the whole point —
	// the defective version returned zero ids too.
	if len(issues) == 0 {
		t.Fatalf("a fence that declares facts and does not parse must be REPORTED, got ids=%v issues=%v", ids, issues)
	}
	if !strings.Contains(issues[0], "IGNORED") {
		t.Fatalf("the issue must say the declaration was ignored, so a reader knows the fence is inert: %q", issues[0])
	}
	if len(ids) != 0 {
		t.Fatalf("an unreadable fence must yield no ids: %v", ids)
	}
}

// The narrowing is load-bearing in the other direction: the fail-open contract
// for a fence that never mentioned facts must survive, or every tool PLAN with
// a hand-written fence starts erroring somewhere it never used to.
func TestParseCriteriaFacts_UnreadableFenceWithoutFactsStillFailsOpen(t *testing.T) {
	for _, criteria := range []string{`{not json`, `{"profiles":["desktop"],}`, ``} {
		ids, issues := parseCriteriaFacts(criteria)
		if len(ids) != 0 || len(issues) != 0 {
			t.Fatalf("criteria %q must fail open silently, got ids=%v issues=%v", criteria, ids, issues)
		}
	}
}

// factsKeyMentioned is deliberately the same predicate as the fan-out's SQL
// prefilter (`dp.body LIKE '%"facts"%'`). If the two drift, the write gate
// refuses what the sweep ignores, or the sweep acts on what the gate never saw.
func TestFactsKeyMentioned_MatchesTheSQLPrefilter(t *testing.T) {
	cases := map[string]bool{
		`{"facts":["a"]}`:                                   true,
		`{"facts": []}`:                                     true,
		`{"profiles":["desktop"]}`:                          false,
		`{"checks":[{"facts_note":"no"}]}`:                  false, // not the "facts" key
		`{"checks":[{"id":"mentions \"facts\" in prose"}]}`: false, // guardian objection: a VALUE quoting the word must not trip it
		`{"facts" : ["a"]}`:                                 true,  // whitespace before the colon is still key position
		``:                                                  false,
	}
	for criteria, want := range cases {
		if got := factsKeyMentioned(criteria); got != want {
			t.Fatalf("factsKeyMentioned(%q) = %v, want %v", criteria, got, want)
		}
	}
}

// ── 1c: the write door refuses a declaration it cannot read ─────────────────

func writeDocPlanParams(t *testing.T, subjectType, subjectKey, body string) (ActionParams, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		Headers:          map[string]string{"agent_type": "tool-generator"},
		CollectedData: map[string]interface{}{
			"doc_plan_body": body,
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"subject_type": subjectType,
			"subject_key":  subjectKey,
		}},
	}, mock
}

func TestWriteDocPlan_RefusesAToolPlanWhoseFactsDeclarationCannotBeRead(t *testing.T) {
	body := "## Acceptance criteria\n\n```criteria\n" +
		`{"profiles":["desktop"],"facts":["sdlt-ftb-relief-cap"],"checks":[],}` +
		"\n```\n"
	params, mock := writeDocPlanParams(t, "tool", "stamp-duty", body)

	// No DB expectations registered AT ALL. That is the assertion: the refusal
	// must happen before the transaction, so a rejected PLAN never supersedes
	// the current row. If the gate is removed the action reaches BeginTx and
	// fails on an unexpected call — a different error, caught below.
	_, err := WriteDocPlanAction(context.Background(), params)
	if err == nil {
		t.Fatal("a tool PLAN whose facts declaration cannot be read must be REFUSED")
	}
	if !strings.Contains(err.Error(), "cannot be read") {
		t.Fatalf("refusal must name the reason, and must not be an incidental DB error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no query should have run: %v", err)
	}
}

// A duplicate id is the other malformed shape criteriaFactsFromValue reports,
// and it is the one a human is most likely to produce by hand. It must be
// refused too — otherwise the gate only catches syntax, not semantics.
func TestWriteDocPlan_RefusesADuplicateFactID(t *testing.T) {
	body := "```criteria\n" + `{"facts":["sdlt-ftb-relief-cap","sdlt-ftb-relief-cap"],"checks":[]}` + "\n```"
	params, _ := writeDocPlanParams(t, "tool", "stamp-duty", body)

	if _, err := WriteDocPlanAction(context.Background(), params); err == nil {
		t.Fatal("a fence declaring the same fact id twice must be refused at the write door")
	}
}

// THE NO-OP PROOF. Everything that is not a tool PLAN with a facts declaration
// must reach the database exactly as it did before. A gate that quietly changed
// the behaviour of the other 131 current tool PLANs would be a far worse defect
// than the one it closes.
func TestWriteDocPlan_UntouchedPathsStillWrite(t *testing.T) {
	cases := []struct {
		name        string
		subjectType string
		body        string
	}{
		{"tool PLAN with no criteria fence at all", "tool", "## Plan\n\nprose only\n"},
		{"tool PLAN whose fence never mentions facts", "tool",
			"```criteria\n" + `{"profiles":["desktop"],"checks":[{"id":"c1"}]}` + "\n```"},
		{"tool PLAN with a well-formed declaration", "tool",
			"```criteria\n" + `{"facts":["sdlt-ftb-relief-cap"],"checks":[]}` + "\n```"},
		{"a pipeline PLAN carrying a broken facts fence is NOT this gate's business", "pipeline",
			"```criteria\n" + `{"facts":["x"],}` + "\n```"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			params, mock := writeDocPlanParams(t, c.subjectType, "stamp-duty", c.body)
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE doc_plans").
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectQuery("INSERT INTO doc_plans").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.NewString()))
			mock.ExpectCommit()

			if _, err := WriteDocPlanAction(context.Background(), params); err != nil {
				t.Fatalf("must still write: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("the write did not happen as before: %v", err)
			}
		})
	}
}

// ── 1b: the finding gets a durable surface, not a log line ─────────────────

func factDriftIndexWithProblems() *factDriftIndex {
	return &factDriftIndex{
		byFact: map[string][]factDriftTool{},
		issues: []string{"stamp-duty: facts[1] is empty"},
		issuesBySubject: map[string][]string{
			"stamp-duty": {"facts[1] is empty"},
		},
		unresolved: []string{"annuity:sdlt-ghost-id"},
		unresolvedBySubject: map[string][]string{
			"annuity": {"sdlt-ghost-id"},
		},
	}
}

func TestNoteBrokenFactDeclarations_FilesOneNotePerSubject(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Subjects are visited in sorted order so the expectations are stable:
	// "annuity" (unresolved id) then "stamp-duty" (malformed).
	site := uuid.New()
	for _, subject := range []string{"annuity", "stamp-duty"} {
		// (subject, site): ONE fleet-global PLAN resolves on many sites, and the
		// unresolved-ids half of the finding is per-site.
		mock.ExpectQuery("SELECT EXISTS").WithArgs(subject, site).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectExec("INSERT INTO doc_notes").
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	noteBrokenFactDeclarations(context.Background(), db, site,
		factDriftIndexWithProblems(), false, zap.NewNop())

	// MUTATION THAT MUST GO RED: delete the noteBrokenFactDeclarations call from
	// planSiteFactDrift, or make this function return early. The unmet INSERT
	// expectation is the failure — asserting the effect, never the absence of a
	// call (LANDMINES: a test asserting a query is NOT issued passes vacuously).
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a note per affected subject must be written: %v", err)
	}
}

func TestNoteBrokenFactDeclarations_CooldownSuppressesARepeat(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Both subjects already noted inside the 30-day window: no INSERT may run.
	// A daily sweep over a fence that stays broken must not mint thirty notes.
	site := uuid.New()
	for _, subject := range []string{"annuity", "stamp-duty"} {
		mock.ExpectQuery("SELECT EXISTS").WithArgs(subject, site).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	}

	noteBrokenFactDeclarations(context.Background(), db, site,
		factDriftIndexWithProblems(), false, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("cooldown must suppress the insert: %v", err)
	}
}

// A dry run writes NOTHING. planSiteFactDrift runs before the action's own
// dry-run return, so this function is the only place that can honour it — and
// the lane's induced-proof recipe depends on a dry run being side-effect free.
func TestNoteBrokenFactDeclarations_DryRunWritesNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	noteBrokenFactDeclarations(context.Background(), db, uuid.New(),
		factDriftIndexWithProblems(), true, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a dry run must not touch the database: %v", err)
	}
}

// The clean case: a site whose declarations all parse and all resolve must be
// silent. This is the population — 131 of 132 current tool PLANs declare
// nothing, and the one that declares is well formed.
func TestNoteBrokenFactDeclarations_CleanSiteIsSilent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	clean := &factDriftIndex{
		byFact:              map[string][]factDriftTool{"sdlt-ftb-relief-cap": {fdTool("stamp-duty", false, true)}},
		issuesBySubject:     map[string][]string{},
		unresolvedBySubject: map[string][]string{},
	}
	noteBrokenFactDeclarations(context.Background(), db, uuid.New(), clean, false, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a clean site must write nothing: %v", err)
	}
}

// ── 1b, WIRED: the surface is reached from the sweep, not just callable ─────
//
// ⚠ THIS TEST EXISTS BECAUSE THE FIRST THREE DID NOT CATCH THE DEFECT.
// Mutation 3 (delete the noteBrokenFactDeclarations call from planSiteFactDrift)
// left every note test above GREEN, because they call the writer directly and
// so cannot see whether anything calls it. That is this lane's own recorded
// lesson repeating in the same file — a mutation that passes usually means the
// test bypassed the path (WRONG_CALLS 2026-08-16, three mutations to prove one
// fix). The assertion here is on the INSERT actually being issued while driving
// the real entry point.
func TestPlanSiteFactDrift_MalformedDeclarationReachesTheDurableSurface(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// A fence that declares the same id twice: it parses, so ids survive and the
	// fan-out still runs — and it carries one issue that must not vanish.
	body := "```criteria\n" +
		`{"facts":["sdlt-ftb-relief-cap","sdlt-ftb-relief-cap"],"no_auto_fix":true,"checks":[]}` +
		"\n```"
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}).
			AddRow("3d7d0d72-0000-4000-8000-000000000001", "tool-stamp-duty", "", "complete", "stamp-duty", body, ""))
	// The durable surface, in the order planSiteFactDrift reaches it: cooldown
	// probe, then the note. sqlmock is ordered and strict, so dropping the call
	// leaves both unmet and this test red.
	mock.ExpectQuery("SELECT EXISTS").WithArgs("stamp-duty", fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec("INSERT INTO doc_notes").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(factDriftLastItemQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subject_key", "new_value"}))

	res := &siteRefreshResult{SiteID: fdSiteIDStr, Domain: "example.test"}
	eb := map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(500000)}}
	// dryRun=false: this is the real pass, the one that may write.
	plan := planSiteFactDrift(context.Background(), db, fdSiteID, eb, res, false, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a malformed declaration must reach the durable surface from the sweep itself: %v", err)
	}
	// And the pre-existing behaviour is untouched: the well-formed half of the
	// declaration still fans out.
	if len(plan.Emissions) != 1 {
		t.Fatalf("the readable half of the declaration must still fan out, got %d emissions", len(plan.Emissions))
	}
}

// The no-op twin, wired: a site whose declarations are all clean must issue NO
// note query at all — not a cooldown probe, not an insert. Without this, a bug
// that fired the cooldown probe on every clean site every day would be invisible.
func TestPlanSiteFactDrift_CleanDeclarationTouchesNoNoteQueries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	body := "```criteria\n" + `{"facts":["sdlt-ftb-relief-cap"],"no_auto_fix":true,"checks":[]}` + "\n```"
	mock.ExpectQuery(regexp.QuoteMeta(factDriftIndexQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "build_status", "subject_key", "body", "fork_component_id"}).
			AddRow("3d7d0d72-0000-4000-8000-000000000001", "tool-stamp-duty", "", "complete", "stamp-duty", body, ""))
	mock.ExpectQuery(regexp.QuoteMeta(factDriftLastItemQuery)).WithArgs(fdSiteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subject_key", "new_value"}))

	res := &siteRefreshResult{SiteID: fdSiteIDStr, Domain: "example.test"}
	eb := map[string]interface{}{"facts": []interface{}{sdltReliefCapFact(500000)}}
	if plan := planSiteFactDrift(context.Background(), db, fdSiteID, eb, res, false, zap.NewNop()); len(plan.Emissions) != 1 {
		t.Fatalf("the clean declaration must still fan out, got %d", len(plan.Emissions))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a clean site must issue no note queries: %v", err)
	}
}
