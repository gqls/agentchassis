// FILE: platform/orchestration/actions/triage_detected_items_site_state_test.go
//
// bugs_open/150 — the improvement loop reports "site is clean" after promoting
// findings.
//
// triage_detected_items is a step in THREE live agents, its promotion is
// unconditional over the site (site_id + status, no type filter), and the
// improvement loop runs its own copy AFTER both children. So the copy the
// parent branches on legitimately reports promoted: 0 for findings a child
// promoted seconds earlier — measured once at 67 findings, orchestration
// 30692439, which then terminated on "No issues found — site is clean" and
// skipped its own closing rerender and dispatch.
//
// The fix does not redefine has_items (a fleet-wide convention with three other
// live consumers). It adds the site-scoped answer beside it. These tests pin
// the distinction, both directions of the branch, and the failure mode of the
// count itself — the last one because a silent zero here is indistinguishable
// from the bug.
package actions

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// triageParams builds the action's inputs with site_id in collected data,
// which is where a live workflow's ensure_site_record step leaves it.
func triageParams(siteID uuid.UUID) ActionParams {
	return ActionParams{
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "triage_findings"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    map[string]interface{}{"site_id": siteID.String()},
		Logger:           zap.NewNop(),
	}
}

// runTriage wires the sqlmock DB in and calls the action.
func runTriage(t *testing.T, promoted int64, countRows *sqlmock.Rows, countErr error) map[string]interface{} {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(siteID, "build").
		WillReturnResult(sqlmock.NewResult(0, promoted))

	// The predicate is the contract: dispatchable statuses AND the target
	// pipeline. Expressed in the expectation so a change to either clause
	// fails this test rather than silently changing what "clean" means.
	countExpect := mock.ExpectQuery(`SELECT count\(\*\)[\s\S]*status IN \('triaged','approved'\)[\s\S]*pipeline = \$2`).
		WithArgs(siteID, "build")
	if countErr != nil {
		countExpect.WillReturnError(countErr)
	} else {
		countExpect.WillReturnRows(countRows)
	}

	params := triageParams(siteID)
	params.DB = db

	out, err := TriageDetectedItemsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}

	result, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected a map result, got %T", out)
	}
	return result
}

func countOf(n int64) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(n)
}

// TestTriageReportsSiteWorkWhenAnotherCopyAlreadyPromoted is bugs_open/150
// itself: this call promoted nothing because a child took the rows, and the
// site is nonetheless full of work. has_items must stay honest about the call;
// site_dispatchable must be true.
func TestTriageReportsSiteWorkWhenAnotherCopyAlreadyPromoted(t *testing.T) {
	result := runTriage(t, 0, countOf(67), nil)

	if result["has_items"] != false {
		t.Errorf("has_items is call-scoped and this call promoted nothing; got %v", result["has_items"])
	}
	if result["site_dispatchable"] != true {
		t.Errorf("67 items are waiting on this site — site_dispatchable must be true, got %v", result["site_dispatchable"])
	}
	if result["site_dispatchable_count"] != int64(67) {
		t.Errorf("site_dispatchable_count = %v, want 67", result["site_dispatchable_count"])
	}
}

// TestTriagePromotedAndDispatchableAgreeOnTheOrdinaryPath — the case that
// always worked. It must keep working, and both signals should agree.
func TestTriagePromotedAndDispatchableAgreeOnTheOrdinaryPath(t *testing.T) {
	result := runTriage(t, 3, countOf(3), nil)

	if result["has_items"] != true {
		t.Errorf("has_items = %v, want true after promoting 3", result["has_items"])
	}
	if result["site_dispatchable"] != true {
		t.Errorf("site_dispatchable = %v, want true", result["site_dispatchable"])
	}
}

// TestTriageStaysQuietOnAGenuinelyCleanSite — the negative control. A gate that
// fires on everything is as useless as one that fires on nothing, and this fix
// exists to make a branch fire MORE often, so the quiet case needs pinning.
func TestTriageStaysQuietOnAGenuinelyCleanSite(t *testing.T) {
	result := runTriage(t, 0, countOf(0), nil)

	if result["has_items"] != false {
		t.Errorf("has_items = %v, want false", result["has_items"])
	}
	if result["site_dispatchable"] != false {
		t.Errorf("nothing promoted and nothing waiting is the one state that IS clean; got site_dispatchable=%v", result["site_dispatchable"])
	}
	if result["site_dispatchable_count"] != int64(0) {
		t.Errorf("site_dispatchable_count = %v, want 0", result["site_dispatchable_count"])
	}
}

// TestTriageFailsTowardNotCleanWhenTheCountErrors — an unanswerable count must
// not read as "no work". The loud branch is the safe one: a needless rerender
// costs a render, a false clean costs the findings.
func TestTriageFailsTowardNotCleanWhenTheCountErrors(t *testing.T) {
	result := runTriage(t, 0, nil, errors.New("connection reset"))

	if result["site_dispatchable"] != true {
		t.Errorf("a failed count must report site_dispatchable=true, got %v", result["site_dispatchable"])
	}
	if result["site_dispatchable_error"] == nil {
		t.Error("the reason must be carried in the result, or the caller cannot tell a fail-safe from a finding")
	}
	if result["site_dispatchable_count"] != int64(-1) {
		t.Errorf("site_dispatchable_count = %v, want the -1 sentinel so a count is never confused with a guess", result["site_dispatchable_count"])
	}
	// The promotion itself still succeeded and must still be reported.
	if result["has_items"] != false || result["promoted"] != int64(0) {
		t.Errorf("the promotion result must survive a count failure: %+v", result)
	}
}

// TestImprovementLoopConditionReadsTheSiteScopedField pins the exact literal
// that migration 281 writes into improvement-loop.check_has_findings against
// the exact shape this action returns. Config lives in the database and code
// lives in the image; nothing else makes them fail together when they drift.
func TestImprovementLoopConditionReadsTheSiteScopedField(t *testing.T) {
	const liveCondition = "triage_result.site_dispatchable == true"

	// The bugs_open/150 shape: this copy promoted nothing, the site has work.
	collected := map[string]interface{}{
		"triage_result": map[string]interface{}{
			"promoted":                int64(0),
			"has_items":               false,
			"site_dispatchable":       true,
			"site_dispatchable_count": int64(67),
		},
	}

	met, err := evaluateStringCondition(liveCondition, collected, zap.NewNop())
	if err != nil {
		t.Fatalf("condition failed to evaluate: %v", err)
	}
	if !met {
		t.Errorf("%q must take the then_step (insert_rerender_item) on a site with 67 items waiting", liveCondition)
	}

	// And the old condition on the same data is the bug, kept here as the
	// contrast that explains why the migration exists.
	oldMet, err := evaluateStringCondition("triage_result.has_items == true", collected, zap.NewNop())
	if err != nil {
		t.Fatalf("old condition failed to evaluate: %v", err)
	}
	if oldMet {
		t.Error("the old condition is supposed to be false here — that is bugs_open/150")
	}

	// Genuinely clean: the new condition must NOT fire.
	clean := map[string]interface{}{
		"triage_result": map[string]interface{}{
			"promoted":                int64(0),
			"has_items":               false,
			"site_dispatchable":       false,
			"site_dispatchable_count": int64(0),
		},
	}
	cleanMet, err := evaluateStringCondition(liveCondition, clean, zap.NewNop())
	if err != nil {
		t.Fatalf("condition failed to evaluate: %v", err)
	}
	if cleanMet {
		t.Error("a site with nothing waiting must still reach complete_clean")
	}
}

// TestConditionOnAPreUpgradeBinaryDoesNotSilentlyInvert is the ordering guard.
// Migration 281 must not be applied before the image ships: on a binary with no
// site_dispatchable key the field resolves to nil and the loop takes the clean
// branch every time. Pinning it here makes the ordering rule a property of the
// code rather than a sentence in a SQL header.
func TestConditionOnAPreUpgradeBinaryDoesNotSilentlyInvert(t *testing.T) {
	preUpgrade := map[string]interface{}{
		"triage_result": map[string]interface{}{
			"promoted":  int64(0),
			"has_items": false,
		},
	}

	met, err := evaluateStringCondition("triage_result.site_dispatchable == true", preUpgrade, zap.NewNop())
	if err != nil {
		t.Fatalf("condition failed to evaluate: %v", err)
	}
	if met {
		t.Fatal("unexpected: an absent field evaluated true")
	}
	// It is false, i.e. the loop would report clean on a site with work —
	// which is exactly today's bug, and exactly why the migration carries an
	// ORDER IS LOAD-BEARING banner and a pod-grep gate.
}

// TestDispatchableStatusesStayInLockstepWithTheDispatcher answers the council's
// `guardian` seat (low): workItemDispatchableStatuses is a THIRD literal of a
// status list that already exists in claim_work_item_action.go and
// load_work_item_actions.go, and "the plan admits these must stay in lockstep
// but there is no shared constant enforcing it".
//
// Making it a genuinely shared constant would mean editing the fleet's dispatch
// query — the highest-blast-radius SQL in the platform — as a side effect of a
// bug fix in a different file, which is the trade this test refuses. Instead it
// is the alarm: if either sibling's status set changes, this fails and names the
// file, in the same spirit as the workItemTerminalStatuses / idx_swi_dedup note
// those constants already carry.
func TestDispatchableStatusesStayInLockstepWithTheDispatcher(t *testing.T) {
	// What the shared constant produces, normalised the way SQL would be read.
	want := normaliseSQLStatusList(sqlInList(workItemDispatchableStatuses))

	// `status IN ( … )` — whitespace-insensitive, so a reformat does not fail it.
	inClause := regexp.MustCompile(`(?i)status\s+IN\s*\(([^)]*)\)`)

	for _, sibling := range []string{
		"claim_work_item_action.go", // the atomic claim's guard
		"load_work_item_actions.go", // the dispatcher's selection query
	} {
		src, err := os.ReadFile(sibling)
		if err != nil {
			t.Fatalf("cannot read %s: %v", sibling, err)
		}

		var found bool
		for _, m := range inClause.FindAllStringSubmatch(string(src), -1) {
			if normaliseSQLStatusList(m[1]) == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s no longer contains a `status IN (%s)` clause.\n"+
				"Either that file's dispatchable set changed and workItemDispatchableStatuses "+
				"must follow it, or this constant changed and that file must. They are one "+
				"contract: a drift here does not error at runtime, it silently changes what "+
				"\"this site has work waiting\" means (bugs_open/150).", sibling, want)
		}
	}
}

// normaliseSQLStatusList reduces "'triaged', 'approved'" and "'triaged','approved'"
// to one comparable form. Order is preserved deliberately — a reordering is a
// change worth a human look, and the lists are two entries long.
func normaliseSQLStatusList(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(s, ",", " , ")), " ")
}
