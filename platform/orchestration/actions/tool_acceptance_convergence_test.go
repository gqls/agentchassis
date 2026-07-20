package actions

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// Covers bugs_open/010 candidate (b): nothing bounded the fix loop. Each Tier-4
// fail raised a FRESH improve_tool item, so the fixer's per-item
// max_fix_attempts never engaged across cycles, and an insufficient fix was
// re-attempted — identically — on the 7-day acceptance cadence, forever.
//
// The guard counts terminal improve_tool cycles that failed at the criteria
// failing NOW, since the tool last passed Tier 4, and escalates to
// needs_human_review rather than raising an identical N+1th attempt.

const benchSite = "e33263f4-74f8-494f-b191-546845dbbddf"

// judgeRun is what a driven judge call produced: the SQL it issued (so a test
// can assert WHICH branch fired, from the statement itself rather than from
// prose), the work-item arguments, the doc_note body, and the action result.
type judgeRun struct {
	sql     []string
	spec    string
	itemKey string
	summary string
	note    string
	out     map[string]interface{}
}

// itemInsert returns the site_work_items INSERT the judge issued, or "".
func (r judgeRun) itemInsert() string {
	for _, s := range r.sql {
		if strings.Contains(s, "INSERT INTO site_work_items") {
			return s
		}
	}
	return ""
}

func (r judgeRun) raisedItemType() string {
	ins := r.itemInsert()
	switch {
	case strings.Contains(ins, "'acceptance_stuck'"):
		return "acceptance_stuck"
	case strings.Contains(ins, "'improve_tool'"):
		return "improve_tool"
	}
	return ""
}

type captureArg struct{ got *string }

func (m captureArg) Match(v driver.Value) bool {
	if s, ok := v.(string); ok {
		*m.got = s
	}
	return true
}

// runJudgeFailPath drives JudgeAcceptanceResultsAction down its failure branch
// against a mock DB. priorAttempts is what the convergence count returns;
// countErr (if set) makes that query fail, exercising the fail-open path.
func runJudgeFailPath(t *testing.T, config map[string]interface{}, priorAttempts int, countErr error) judgeRun {
	t.Helper()

	run := judgeRun{}

	// Match every statement permissively, recording it: the assertions are about
	// the SQL the action chose to issue, not about matching a canned string.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(
		sqlmock.QueryMatcherFunc(func(_, actual string) error {
			run.sql = append(run.sql, actual)
			return nil
		})))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1. the convergence count — the query under test
	countQ := mock.ExpectQuery("count")
	if countErr != nil {
		countQ.WillReturnError(countErr)
	} else {
		countQ.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(priorAttempts))
	}
	// 2. the acceptance-fail doc_note
	mock.ExpectQuery("doc_notes").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			captureArg{got: &run.note}, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("note-1"))
	// 3. the component lookup
	mock.ExpectQuery("content_components").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_id"}).AddRow("comp-1", "page-1"))
	// 4. whichever work item the guard chose
	mock.ExpectExec("site_work_items").
		WithArgs(
			sqlmock.AnyArg(),              // $1 site_id
			captureArg{got: &run.summary}, // $2 summary
			captureArg{got: &run.spec},    // $3 spec
			captureArg{got: &run.itemKey}, // $4 item_key
			sqlmock.AnyArg(),              // $5 batch_id
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	stepConfig := map[string]interface{}{}
	for k, v := range config {
		stepConfig[k] = v
	}

	collected := map[string]interface{}{
		"input_data":  map[string]interface{}{"spec": map[string]interface{}{"function": "tool-loot-table-balancer"}},
		"site_record": map[string]interface{}{"site_id": benchSite},
		"browser_run": map[string]interface{}{
			"results": []interface{}{
				map[string]interface{}{
					"check_id": "mobile-fit", "profile": "mobile", "passed": false,
					"detail": "page overflows at 390px, widest element fieldset (419px)",
				},
			},
		},
	}

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		Headers:          map[string]string{"agent_type": "tool-acceptance-agent"},
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: stepConfig},
	}

	res, err := JudgeAcceptanceResultsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("judge failed: %v", err)
	}
	run.out, _ = res.(map[string]interface{})
	return run
}

// The core case: two cycles have already failed at mobile-fit, so the third
// verdict must NOT raise a third identical improve_tool item.
func TestConvergenceGuard_EscalatesAfterTwoFailedCycles(t *testing.T) {
	run := runJudgeFailPath(t, nil, 2, nil)

	if got := run.raisedItemType(); got != "acceptance_stuck" {
		t.Fatalf("expected an escalation item, got %q\nSQL: %s", got, run.itemInsert())
	}
	if !strings.Contains(run.itemInsert(), "'needs_human_review'") {
		t.Errorf("escalation must be raised AT needs_human_review:\n%s", run.itemInsert())
	}
	if !strings.HasPrefix(run.itemKey, "acceptance_stuck:tool-loot-table-balancer:") {
		t.Errorf("escalation must carry its own dedup key, got %q", run.itemKey)
	}
	if escalated, _ := run.out["escalated"].(bool); !escalated {
		t.Errorf("result must report escalated=true, got %v", run.out["escalated"])
	}
	if created, _ := run.out["improve_tool_created"].(bool); created {
		t.Errorf("no improve_tool item may be raised when escalating")
	}
	if spent, _ := run.out["fix_cycles_spent"].(int); spent != 2 {
		t.Errorf("fix_cycles_spent = %v, want 2", run.out["fix_cycles_spent"])
	}
	// An escalation with no stated reason is another item nobody can action:
	// it must name the criterion and the count that justified stopping.
	if !strings.Contains(run.spec, "mobile-fit") || !strings.Contains(run.spec, "why_escalated") {
		t.Errorf("escalation spec lacks its reason: %s", run.spec)
	}
	// The note is the loop's own record; it must not still claim a fix was queued.
	if strings.Contains(run.note, "Fix: improve_tool item created") {
		t.Errorf("acceptance-fail note still claims an improve_tool item was created:\n%s", run.note)
	}
	if !strings.Contains(run.note, "escalated to human review") {
		t.Errorf("acceptance-fail note does not record the escalation:\n%s", run.note)
	}
}

// Below the threshold the loop must be untouched — the guard bounds the loop,
// it does not replace it.
func TestConvergenceGuard_FirstAndSecondCyclesStillAutoFix(t *testing.T) {
	for _, prior := range []int{0, 1} {
		run := runJudgeFailPath(t, nil, prior, nil)
		if got := run.raisedItemType(); got != "improve_tool" {
			t.Fatalf("prior=%d: expected improve_tool, got %q", prior, got)
		}
		if created, _ := run.out["improve_tool_created"].(bool); !created {
			t.Errorf("prior=%d: improve_tool_created should be true", prior)
		}
		if escalated, _ := run.out["escalated"].(bool); escalated {
			t.Errorf("prior=%d: must not escalate below the threshold", prior)
		}
		if !strings.Contains(run.note, "Fix: improve_tool item created") {
			t.Errorf("prior=%d: note should record the queued fix:\n%s", prior, run.note)
		}
		// The fixer's pointed hints must survive on the normal path.
		if !strings.Contains(run.spec, "acceptance_test") {
			t.Errorf("prior=%d: improve_tool spec lost its criteria: %s", prior, run.spec)
		}
	}
}

// A counting error must cost nothing: the loop keeps working exactly as it did
// before the guard existed. Fail-closed would turn a transient DB error into a
// silently dropped fix.
func TestConvergenceGuard_FailsOpenOnCountError(t *testing.T) {
	run := runJudgeFailPath(t, nil, 0, errors.New("connection reset"))

	if got := run.raisedItemType(); got != "improve_tool" {
		t.Fatalf("count error must fail open to improve_tool, got %q", got)
	}
	if escalated, _ := run.out["escalated"].(bool); escalated {
		t.Errorf("a counting error must never escalate")
	}
	if spent, _ := run.out["fix_cycles_spent"].(int); spent != 0 {
		t.Errorf("unknown count must report 0, got %v", run.out["fix_cycles_spent"])
	}
}

// The threshold is step config, so a workstream can widen it without a rebuild.
func TestConvergenceGuard_ThresholdIsConfigurable(t *testing.T) {
	run := runJudgeFailPath(t, map[string]interface{}{"max_fix_cycles": 4}, 3, nil)
	if got := run.raisedItemType(); got != "improve_tool" {
		t.Fatalf("with max_fix_cycles=4, 3 prior attempts must still auto-fix, got %q", got)
	}

	run = runJudgeFailPath(t, map[string]interface{}{"max_fix_cycles": 4}, 4, nil)
	if got := run.raisedItemType(); got != "acceptance_stuck" {
		t.Fatalf("with max_fix_cycles=4, 4 prior attempts must escalate, got %q", got)
	}
}

// The count query IS the guard. Assert the clauses that make it measure
// non-convergence rather than history — each was chosen for a reason a later
// edit could quietly drop, and dropping any one of them silently changes when
// the loop stops. Verified against live data 2026-07-20 (the benchmark tool
// counts 4; a synthetic green verdict at 07-18 00:00 cuts it to 2; an unrelated
// criterion counts 0; the in-flight open item is excluded).
func TestConvergenceAttempts_QueryIsBounded(t *testing.T) {
	var gotSQL string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(
		sqlmock.QueryMatcherFunc(func(_, actual string) error {
			gotSQL = actual
			return nil
		})))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("count").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	n, err := convergenceAttempts(context.Background(), db, "tool-loot-table-balancer", benchSite,
		[]string{"mobile-fit"})
	if err != nil {
		t.Fatalf("convergenceAttempts: %v", err)
	}
	if n != 2 {
		t.Errorf("count = %d, want 2", n)
	}

	for _, clause := range []struct{ name, needle string }{
		{"terminal attempts only", "status IN ('complete', 'failed')"},
		{"reset on a green verdict", "acceptance-run"},
		{"overlap with the criteria failing now", "jsonb_array_elements_text"},
		{"guard against a non-array spec field", "jsonb_typeof"},
		{"scoped to this judge's own item_key", "w.item_key = $2"},
	} {
		if !strings.Contains(gotSQL, clause.needle) {
			t.Errorf("count query lost its %s clause (%q):\n%s", clause.name, clause.needle, gotSQL)
		}
	}
}

// An empty failing set cannot overlap anything; short-circuit rather than send
// Postgres an empty array literal. No expectations are set, so sqlmock fails
// the test if the function queries at all.
func TestConvergenceAttempts_NoFailingChecksCountsZero(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	n, err := convergenceAttempts(context.Background(), db, "t", benchSite, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("count = %d, want 0", n)
	}
}
