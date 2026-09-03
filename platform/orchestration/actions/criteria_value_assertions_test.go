// FILE: platform/orchestration/actions/criteria_value_assertions_test.go
//
// bugs_open/449 — a fence that asserts no NUMBER must not produce a verdict
// that reads like the arithmetic was checked.
//
// THE BAR THE BUG SETS, and it governs every test here: *"A fix that only adds
// checks which pass is indistinguishable from no fix."* So each test names the
// mutation that must turn it red, and every "nothing happened" assertion is
// paired with a DEMAND CONTROL in the same test — a neighbouring case that must
// still fire. A zero from a blind check and a zero from a clean system are the
// same bytes, and this package has shipped that confusion before.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// The shape `tool-generator` actually writes: the four mandatory liveness
// checks plus one interaction that fills an input and asserts only that an
// element APPEARS. 115 of its 186 current fences were this on 2026-09-03.
const generatedLivenessFence = `{"profiles":["desktop","mobile"],"container":"#calc",
 "checks":[
   {"id":"boots","type":"selector_exists","selector":"#calc"},
   {"id":"console","type":"no_console_errors"},
   {"id":"status","type":"page_status_ok"},
   {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
   {"id":"calc","type":"interaction",
    "steps":[{"action":"fill","selector":"#amount","value":"250000"},
             {"action":"click","selector":"#go"}],
    "expect":{"selector":"#monthlyPayment"}}]}`

// ── 1. the grader ──────────────────────────────────────────────────────────

func TestSummariseCriteriaValueAssertions_GradesTheCorpusShapes(t *testing.T) {
	cases := []struct {
		name          string
		fence         string
		wantParsed    bool
		wantGrade     string
		wantDrives    bool
		wantBlindDriv bool
	}{
		{
			// THE DEFECT ITSELF. Fills an input, asserts only that an element
			// exists afterwards. Must grade `none` AND be flagged as driving.
			name:  "generated liveness fence that fills and asserts nothing",
			fence: generatedLivenessFence, wantParsed: true,
			wantGrade: criteriaAssertsNone, wantDrives: true, wantBlindDriv: true,
		},
		{
			// DEMAND CONTROL for the case above: the same document with one
			// text_matches added must stop being flagged. Without this arm a
			// rule that flags EVERY fence would pass the test above.
			name: "the same fence with a text_matches is a pattern assertion",
			fence: strings.Replace(generatedLivenessFence,
				`"expect":{"selector":"#monthlyPayment"}`,
				`"expect":{"selector":"#monthlyPayment","text_matches":"£[\\d,]+\\.\\d\\d"}`, 1),
			wantParsed: true, wantGrade: criteriaAssertsPattern,
			wantDrives: true, wantBlindDriv: false,
		},
		{
			name: "computed_values with expectations is an exact assertion",
			fence: `{"checks":[{"id":"sums","type":"computed_values",
			  "steps":[{"action":"fill","selector":"#amount","value":"250000"}],
			  "expect_values":{"#monthlyPayment":"£303.44"}}]}`,
			wantParsed: true, wantGrade: criteriaAssertsExact,
			wantDrives: true, wantBlindDriv: false,
		},
		{
			// MUTATION THAT MUST GO RED: credit `computed_values` by TYPE
			// instead of by having expectations. The runner refuses an empty
			// expect_values outright ("it would assert nothing and pass on any
			// page"), so a fence carrying one asserts nothing and can only ever
			// fail — crediting it would hand a blind fence the top grade on the
			// strength of a type name.
			name: "computed_values with an EMPTY expect_values is credited with nothing",
			fence: `{"checks":[{"id":"sums","type":"computed_values",
			  "steps":[{"action":"fill","selector":"#amount","value":"1"}],
			  "expect_values":{}}]}`,
			wantParsed: true, wantGrade: criteriaAssertsNone,
			wantDrives: true, wantBlindDriv: true,
		},
		{
			// A tool you only click is not an input-taker, so it is outside the
			// door's trigger even though it asserts nothing. This is what keeps
			// the rule off toggles, tabs and reveals.
			name: "clicks only — asserts nothing, but does not DRIVE",
			fence: `{"checks":[{"id":"open","type":"interaction",
			  "steps":[{"action":"click","selector":"#more"}],
			  "expect":{"selector":"#panel"}}]}`,
			wantParsed: true, wantGrade: criteriaAssertsNone,
			wantDrives: false, wantBlindDriv: false,
		},
		{
			// `select` supplies a value just as `fill` does. Collecting only
			// fills silently dropped a real input in the mcalc lane's own
			// verifier (its `driven()` docstring records it), so it is pinned.
			name: "select counts as driving, not just fill",
			fence: `{"checks":[{"id":"pick","type":"interaction",
			  "steps":[{"action":"select","selector":"#buyerType","value":"ftb"}],
			  "expect":{"selector":"#tax"}}]}`,
			wantParsed: true, wantGrade: criteriaAssertsNone,
			wantDrives: true, wantBlindDriv: true,
		},
		{
			// FAIL-OPEN. An unreadable fence must NOT report "asserts nothing":
			// that is a claim about a document nobody could read, and Tier 2
			// already reports this case separately as criteria_unparseable.
			name:       "an unparseable fence is UNKNOWN, never a finding",
			fence:      `{"checks":[{"id":"x","type":"interaction"},]}`,
			wantParsed: false, wantGrade: criteriaAssertsNone,
			wantDrives: false, wantBlindDriv: false,
		},
		{
			name: "no fence at all", fence: "  ",
			wantParsed: false, wantGrade: criteriaAssertsNone,
			wantDrives: false, wantBlindDriv: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := summariseCriteriaValueAssertions(c.fence)
			if got.Parsed != c.wantParsed {
				t.Fatalf("Parsed = %v, want %v", got.Parsed, c.wantParsed)
			}
			if g := got.Grade(); g != c.wantGrade {
				t.Errorf("Grade() = %q, want %q (exact=%d pattern=%d)",
					g, c.wantGrade, got.Exact, got.Pattern)
			}
			if got.DrivesInputs != c.wantDrives {
				t.Errorf("DrivesInputs = %v, want %v", got.DrivesInputs, c.wantDrives)
			}
			if got.DrivesButAssertsNothing() != c.wantBlindDriv {
				t.Errorf("DrivesButAssertsNothing() = %v, want %v",
					got.DrivesButAssertsNothing(), c.wantBlindDriv)
			}
		})
	}
}

// AssertsNoValue must be false for a fence nobody could read — "I could not
// tell" and "it asserts nothing" are different states, and collapsing them
// would let a parse bug present as a corpus finding.
func TestSummariseCriteriaValueAssertions_UnreadableIsNotAFinding(t *testing.T) {
	unreadable := summariseCriteriaValueAssertions(`{"checks":[,]}`)
	if unreadable.AssertsNoValue() {
		t.Fatal("an unparseable fence must not report AssertsNoValue — Tier 2 owns criteria_unparseable")
	}
	// DEMAND CONTROL: the readable blind fence in the same test, so a
	// permanently-false AssertsNoValue cannot pass this.
	if !summariseCriteriaValueAssertions(generatedLivenessFence).AssertsNoValue() {
		t.Fatal("a readable fence asserting nothing MUST report AssertsNoValue — else the check above is vacuous")
	}
}

// The phrase is what a human actually reads on a verdict, so it is pinned:
// it must name the limit, not merely omit the claim.
func TestCriteriaAssertionPhrase_SaysWhatTheVerdictDoesNotCover(t *testing.T) {
	blind := criteriaAssertionPhrase(summariseCriteriaValueAssertions(generatedLivenessFence))
	if !strings.Contains(blind, "LIVENESS ONLY") || !strings.Contains(blind, "bugs_open/449") {
		t.Fatalf("a value-less pass must be labelled and traceable, got %q", blind)
	}
	if !strings.Contains(blind, "NOTHING") {
		t.Fatalf("the phrase must state what was NOT checked, got %q", blind)
	}
	exact := criteriaAssertionPhrase(summariseCriteriaValueAssertions(
		`{"checks":[{"id":"sums","type":"computed_values","expect_values":{"#p":"£1.00"}}]}`))
	if strings.Contains(exact, "LIVENESS ONLY") {
		t.Fatalf("a fence that compares an exact value must NOT be labelled liveness-only, got %q", exact)
	}
	if !strings.Contains(exact, "sums") {
		t.Fatalf("the phrase must cite the asserting check rather than assert a count, got %q", exact)
	}
}

// ── 2. the verdict (P1) ────────────────────────────────────────────────────

// runJudgePassPath drives JudgeAcceptanceResultsAction down its ALL-PASSED arm
// with the given fence, and returns the acceptance-run note body plus the
// action result.
func runJudgePassPath(t *testing.T, criteria string) (string, map[string]interface{}) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	var note string
	mock.ExpectQuery("doc_notes").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			captureArg{got: &note}, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("note-1"))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		Headers:          map[string]string{"agent_type": "tool-acceptance-agent"},
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data":  map[string]interface{}{"spec": map[string]interface{}{"function": "tool-simple"}},
			"site_record": map[string]interface{}{"site_id": uuid.NewString()},
			"doc_context": map[string]interface{}{"criteria_json": criteria},
			"browser_run": map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{"check_id": "boots", "profile": "desktop", "pass": true},
					map[string]interface{}{"check_id": "calc", "profile": "desktop", "pass": true},
				},
			},
		},
		StepConfig: models.Step{Config: map[string]interface{}{}},
	}

	res, err := JudgeAcceptanceResultsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("judge failed: %v", err)
	}
	out, _ := res.(map[string]interface{})
	return note, out
}

// THE CORE CASE. A PASS on a fence that fills inputs and checks no number must
// say so, in the note a human reads and in the result a machine reads.
//
// MUTATION THAT MUST GO RED: delete the `Scope of this verdict:` line from the
// pass note, or stop setting verdict_scope. Note the assertion is on the
// QUALIFICATION being present — not on the pass being absent — because the
// pre-fix code passed too, and a test satisfied by "all_passed is true" is
// satisfied by the bug.
func TestJudgePass_OnAValuelessFenceReportsLivenessOnly(t *testing.T) {
	note, out := runJudgePassPath(t, generatedLivenessFence)

	if !strings.Contains(note, "Scope of this verdict:") {
		t.Fatalf("a PASS note must state its own scope:\n%s", note)
	}
	if !strings.Contains(note, "LIVENESS ONLY") {
		t.Fatalf("a value-less PASS must be labelled liveness-only:\n%s", note)
	}
	if got := out["assertion_grade"]; got != criteriaAssertsNone {
		t.Errorf("assertion_grade = %v, want %q", got, criteriaAssertsNone)
	}
	if got := out["verdict_scope"]; got != "liveness_only" {
		t.Errorf("verdict_scope = %v, want liveness_only", got)
	}
	if got, _ := out["value_assertions"].(int); got != 0 {
		t.Errorf("value_assertions = %v, want 0", got)
	}
	// The pass is still a pass. This fix changes what the record CLAIMS, never
	// what it decides — a verdict that started failing would be a different and
	// much larger change than the one under review.
	if passed, _ := out["all_passed"].(bool); !passed {
		t.Error("the verdict itself must be unchanged: a passing tool still passes")
	}
}

// THE DEMAND CONTROL, and it is the load-bearing half of the pair. If the
// counter were blind — always zero, always liveness_only — the test above would
// pass and mean nothing. A fence that DOES compare a number must come back
// non-zero and must NOT be labelled.
func TestJudgePass_OnAComputedValuesFenceIsNotLabelled(t *testing.T) {
	note, out := runJudgePassPath(t, `{"profiles":["desktop"],"checks":[
	  {"id":"boots","type":"selector_exists","selector":"#calc"},
	  {"id":"sums","type":"computed_values",
	   "steps":[{"action":"fill","selector":"#amount","value":"250000"}],
	   "expect_values":{"#monthlyPayment":"£303.44","#totalInterest":"£41,238.40"}}]}`)

	if strings.Contains(note, "LIVENESS ONLY") {
		t.Fatalf("a fence comparing exact values must NOT be labelled liveness-only:\n%s", note)
	}
	if got := out["assertion_grade"]; got != criteriaAssertsExact {
		t.Errorf("assertion_grade = %v, want %q", got, criteriaAssertsExact)
	}
	if _, present := out["verdict_scope"]; present {
		t.Error("verdict_scope must be ABSENT when the fence asserts values — a consumer branches on its presence")
	}
	if got, _ := out["exact_value_assertions"].(int); got != 1 {
		t.Errorf("exact_value_assertions = %v, want 1", got)
	}
}

// A pattern assertion is real evidence and WEAKER evidence. Reporting it as
// either `none` or `exact` would be a lie in one direction or the other, and
// both errors are live in the corpus (bugs_open/449 §2 corrected itself on
// exactly this point).
func TestJudgePass_PatternAssertionIsGradedBetweenTheTwo(t *testing.T) {
	fence := strings.Replace(generatedLivenessFence,
		`"expect":{"selector":"#monthlyPayment"}`,
		`"expect":{"selector":"#monthlyPayment","text_matches":"£[\\d,]+\\.\\d\\d"}`, 1)
	note, out := runJudgePassPath(t, fence)

	if got := out["assertion_grade"]; got != criteriaAssertsPattern {
		t.Fatalf("assertion_grade = %v, want %q", got, criteriaAssertsPattern)
	}
	if _, present := out["verdict_scope"]; present {
		t.Error("a pattern assertion is an assertion — verdict_scope must be absent")
	}
	if !strings.Contains(note, "weaker") {
		t.Errorf("the note must say a hand-authored pattern is weaker than an arithmetic check:\n%s", note)
	}
}

// ── 3. the write door (P2) ─────────────────────────────────────────────────

// runDoorWithFence writes a tool PLAN carrying `fence` and reports whether the
// door emitted a fence_asserts_no_value note. `expectNote` drives the mock's
// expectations, so a note that fires when none was expected fails on an
// unexpected call rather than passing silently.
func runDoorWithFence(t *testing.T, fence string, expectNote bool) string {
	t.Helper()
	body := "## Acceptance criteria\n\n```criteria\n" + fence + "\n```\n"
	params, mock := writeDocPlanParams(t, "tool", "simple", body)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE doc_plans").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO doc_plans").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.NewString()))
	mock.ExpectCommit()

	var note string
	if expectNote {
		mock.ExpectQuery("doc_notes").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				captureArg{got: &note}, sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("note-1"))
	}

	if _, err := WriteDocPlanAction(context.Background(), params); err != nil {
		t.Fatalf("the PLAN must still be written — this rule records, it does not refuse: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("door behaved unexpectedly (expectNote=%v): %v", expectNote, err)
	}
	return note
}

// MUTATION THAT MUST GO RED: remove the door's 449 block. And note what is
// asserted — that the PLAN was still WRITTEN. Refusing here would strand the
// tool: a tool with no PLAN is inert at BOTH tiers, which is worse than a blind
// fence, so "records but does not refuse" is the behaviour under test.
func TestWriteDocPlan_RecordsAFenceThatDrivesInputsAndAssertsNothing(t *testing.T) {
	note := runDoorWithFence(t, generatedLivenessFence, true)

	if !strings.Contains(note, "fence_asserts_no_value") {
		t.Errorf("the note must carry its category so the sweep can find it:\n%s", note)
	}
	if !strings.Contains(note, "DRIVES its inputs") {
		t.Errorf("the note must say why this fence was singled out:\n%s", note)
	}
	if !strings.Contains(note, "bugs_open/449") {
		t.Errorf("the note must be traceable to the bug:\n%s", note)
	}
	if !strings.Contains(note, "not a refusal") {
		t.Errorf("the note must say the PLAN was accepted, or a reader will hunt for a rejected document:\n%s", note)
	}
}

// THE DEMAND CONTROLS. Each is a fence the rule must stay silent on, and
// together they are what stops "note everything" passing the test above.
func TestWriteDocPlan_StaysSilentWhereItShould(t *testing.T) {
	cases := []struct {
		name  string
		fence string
	}{
		{"a fence that asserts a pattern", strings.Replace(generatedLivenessFence,
			`"expect":{"selector":"#monthlyPayment"}`,
			`"expect":{"selector":"#monthlyPayment","text_matches":"£[\\d,]+"}`, 1)},
		{"a fence that asserts exact values",
			`{"checks":[{"id":"sums","type":"computed_values",
			  "steps":[{"action":"fill","selector":"#a","value":"1"}],
			  "expect_values":{"#p":"£1.00"}}]}`},
		{"a click-only tool never claimed to take input",
			`{"checks":[{"id":"open","type":"interaction",
			  "steps":[{"action":"click","selector":"#more"}],"expect":{"selector":"#panel"}}]}`},
		{"an unreadable fence is Tier 2's criteria_unparseable, not ours",
			`{"checks":[{"id":"x",},]}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			runDoorWithFence(t, c.fence, false)
		})
	}
}

// A non-tool PLAN is not this rule's business, however its fence looks — the
// same scoping the facts gate uses, so every other PLAN write is byte-identical
// to before.
func TestWriteDocPlan_NonToolSubjectIsUntouched(t *testing.T) {
	body := "```criteria\n" + generatedLivenessFence + "\n```"
	params, mock := writeDocPlanParams(t, "pipeline", "some-pipeline", body)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE doc_plans").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO doc_plans").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.NewString()))
	mock.ExpectCommit()

	if _, err := WriteDocPlanAction(context.Background(), params); err != nil {
		t.Fatalf("a pipeline PLAN must write exactly as before: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a pipeline PLAN must emit no note: %v", err)
	}
}
