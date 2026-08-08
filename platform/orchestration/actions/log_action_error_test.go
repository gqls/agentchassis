// FILE: platform/orchestration/actions/log_action_error_test.go
//
// The merge's contract, pinned on the axis NOTHING in this package pinned
// before: agent_type and step_name.
//
// This file exists because of a measurement, not a hunch. Four council seats
// (bug_historian, guardian, architecture, and the RFC_012 lane's own risks
// block) independently reported that LogActionEntry filled a FORGOTTEN
// provenance silently from the running step — and that "no test in the package
// would catch it", because every sqlmock expectation here pins the error_code,
// the action and the message, and passes sqlmock.AnyArg() for agent_type. A
// broken merge was green on the whole suite. That was the defect, and a suite
// that stays green is therefore not evidence: each test below is written to FAIL
// against the pre-split merge, and was run against a mutated build to prove it.
//
// Argument positions, because they are easy to miscount and every test here
// depends on them (agenterrors.Write's 13-argument INSERT):
//
//	1 site_id   2 domain    3 work_item_id  4 orchestration_id
//	5 agent_type  6 agent_id  7 pod_name    8 step_name  9 action
//	10 error_message  11 error_code  12 severity  13 context
package actions

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// runningStepParams is a params set whose running identity is UNAMBIGUOUS and
// unlike any provenance a test sets explicitly — so a test can never pass by
// coincidence when the merge picks the wrong source.
func runningStepParams(db *sql.DB) ActionParams {
	return ActionParams{
		DB:          db,
		Logger:      zap.NewNop(),
		AgentType:   "params-fallback-agent",
		CurrentStep: "params_fallback_step",
		ExecutionContext: &types.ExecutionContext{
			OrchestrationID: "orch-1234",
			StepName:        "running_step",
			Sender: types.AgentIdentity{
				AgentType: "running-agent",
				AgentID:   "agent-id-1",
				PodName:   "pod-1",
			},
		},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"work_item_id": "item-1234"},
		},
	}
}

// jsonHas matches a marshalled context payload containing every given key.
// Asserting on the whole JSON string would pin key ORDER, which encoding/json
// does not guarantee for a map.
type jsonHas []string

func (j jsonHas) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		return false
	}
	for _, k := range j {
		if _, present := payload[k]; !present {
			return false
		}
	}
	return true
}

// jsonLacks is jsonHas's negative: no listed key may be present. Used to prove
// a clean row is not annotated with an unattributed row's bookkeeping.
type jsonLacks []string

func (j jsonLacks) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		// "null" from a nil map unmarshals to a nil map, which lacks everything.
		return strings.TrimSpace(s) == "null"
	}
	for _, k := range j {
		if _, present := payload[k]; present {
			return false
		}
	}
	return true
}

const insertRE = `INSERT INTO agent_error_log`

// ---------------------------------------------------------------------------
// The reported defect: a forgotten provenance must not be credited to whatever
// step happened to be running.
// ---------------------------------------------------------------------------

func TestLogActionEntry_ForgottenProvenanceIsNotCreditedToTheRunningStep(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// FAILS on the pre-split merge, which sent "running-agent"/"running_step".
	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"", "", "item-1234", "orch-1234",
			"unattributed", // 5: NOT "running-agent"
			"agent-id-1", "pod-1",
			"", // 8: step_name left empty — a row that names no agent must not imply one
			"some_action", "msg", "SOME_CODE", "warning",
			jsonHas{"provenance_missing", "provenance_running_agent_type", "provenance_running_step_name", "remedy"},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ok := LogActionEntry(context.Background(), runningStepParams(db), agenterrors.Entry{
		Action:       "some_action",
		ErrorMessage: "msg",
		ErrorCode:    "SOME_CODE",
		Severity:     "warning",
		// AgentType deliberately absent — this is the forgotten-provenance case.
	}, zap.NewNop())

	if !ok {
		t.Fatal("the row must still land: agent_error_log is the only sink that survives an awaited step, so refusing the write destroys a finding")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a forgotten provenance was not marked unattributed: %v", err)
	}
}

// The sentinel is load-bearing in a way an empty string is not: agent_type is
// NOT NULL, and agenterrors.Write swallows a constraint violation as a warn —
// so an empty string does not produce a bad row, it produces NO row.
func TestLogActionEntry_UnattributedIsNeverTheEmptyString(t *testing.T) {
	if unattributedAgentType == "" {
		t.Fatal("unattributedAgentType is empty: agent_type is NOT NULL, and the best-effort writer would swallow the violation, losing the row entirely")
	}
	// It must also not collide with a value real traffic already uses, or a
	// detector keyed on it cannot separate a mistake from ordinary noise.
	// 'generic' has 559 live rows across 25 step_names (measured 2026-08-08).
	for _, live := range []string{"generic", "unknown"} {
		if unattributedAgentType == live {
			t.Errorf("unattributedAgentType = %q, which is a live agent_type value with real traffic — the detector could never separate a provenance mistake from noise", live)
		}
	}
}

// ---------------------------------------------------------------------------
// The declared opt-in: the 7 sites where the running step IS the provenance.
// ---------------------------------------------------------------------------

func TestLogActionEntryInheritingProvenance_DoesInheritTheRunningStep(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"", "", "item-1234", "orch-1234",
			"running-agent", // 5
			"agent-id-1", "pod-1",
			"running_step", // 8
			"some_action", "msg", "SOME_CODE", "warning",
			jsonLacks{"provenance_missing"}, // a declared row is not a mistake
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionEntryInheritingProvenance(context.Background(), runningStepParams(db), agenterrors.Entry{
		Action:       "some_action",
		ErrorMessage: "msg",
		ErrorCode:    "SOME_CODE",
		Severity:     "warning",
	}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the declared opt-in did not inherit the running step: %v", err)
	}
}

// LogActionError's whole contract is "file under the executing step", so it
// declares inheritance on its caller's behalf. If this regresses, the simple
// door starts writing unattributed rows fleet-wide.
func TestLogActionError_StillFilesUnderTheRunningStep(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"site-1", "example.com", "item-1234", "orch-1234",
			"running-agent", "agent-id-1", "pod-1", "running_step",
			"an_action", "boom", "A_CODE", "error",
			jsonLacks{"provenance_missing"},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionError(context.Background(), runningStepParams(db),
		"site-1", "example.com", "an_action", "A_CODE", "error", "boom", nil, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("LogActionError no longer files under the running step: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Invariants the split must NOT break — the no-op half of the check, which is
// the half that gets skipped (memory: check-the-no-op-case-not-only-the-damage-case).
// ---------------------------------------------------------------------------

// The 13 sites that name their own provenance: a field the caller set is used
// verbatim, and was already safe. Proving it still is, is the regression guard.
func TestLogActionEntry_ExplicitProvenanceIsUsedVerbatim(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"", "", "item-1234", "orch-1234",
			"component-creator", // 5: the caller's literal, not "running-agent"
			"agent-id-1", "pod-1",
			"store_component", // 8: likewise
			"store_generated_component", "msg", "A_CODE", "warning",
			jsonLacks{"provenance_missing"},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionEntry(context.Background(), runningStepParams(db), agenterrors.Entry{
		AgentType:    "component-creator",
		StepName:     "store_component",
		Action:       "store_generated_component",
		ErrorMessage: "msg",
		ErrorCode:    "A_CODE",
		Severity:     "warning",
	}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("an explicit provenance literal was not used verbatim: %v", err)
	}
}

// The strict door is strict on BOTH provenance columns, not just agent_type.
//
// The first version of this seam inherited step_name whenever agent_type was
// named, and this test asserted that. Two seats of the approving council round
// (bug_historian at medium, editquality) independently objected that a mixed row
// — agent from the caller, step from the runtime — is "a narrower recurrence of
// the exact defect being fixed", leaving the same failure shape live for any
// future caller of that shape. They were right, so the assertion is INVERTED
// here rather than deleted: an empty step_name asserts nothing, a borrowed one
// asserts something no caller claimed. The only two sites of that shape
// (prepare_link_context, diagnose_persist_fix_plan) now declare inheritance, so
// no live row's contents changed — see the two tests below this one.
func TestLogActionEntry_NamingOnlyAgentTypeDoesNotBorrowTheRunningStep(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"", "", "item-1234", "orch-1234",
			"unknown",    // 5: the site's own NOT NULL fallback, kept verbatim
			"agent-id-1", // 6
			"pod-1",      // 7
			"",           // 8: NOT "running_step" — the caller named no step, so the row claims none
			"an_action", "msg", "A_CODE", "warning",
			jsonLacks{"provenance_missing"}, // agent_type WAS named, so this is not an unattributed row
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionEntry(context.Background(), runningStepParams(db), agenterrors.Entry{
		AgentType:    "unknown",
		Action:       "an_action",
		ErrorMessage: "msg",
		ErrorCode:    "A_CODE",
		Severity:     "warning",
	}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("a caller that named only agent_type must not have step_name filled from the running step: %v", err)
	}
}

// The two sites the inversion above moved, pinned so the claim "no live row's
// contents changed" is a test rather than an assertion. Both name their own
// agent_type (each carries a NOT NULL fallback of its own) and want the running
// step's step_name — which is now declared instead of assumed.
func TestLogActionEntryInheritingProvenance_KeepsAnExplicitAgentTypeAndStillFillsStepName(t *testing.T) {
	for _, tc := range []struct{ name, agentType, action string }{
		{"prepare_link_context", "unknown", "prepare_link_context"},
		{"diagnose_persist_fix_plan", "diagnose-agent", "diagnose_persist_fix_plan"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectExec(regexp.QuoteMeta(insertRE)).
				WithArgs(
					"", "", "item-1234", "orch-1234",
					tc.agentType, // 5: the caller's, NOT "running-agent"
					"agent-id-1", "pod-1",
					"running_step", // 8: inherited, because inheritance is DECLARED
					tc.action, "msg", "A_CODE", "warning",
					jsonLacks{"provenance_missing"},
				).
				WillReturnResult(sqlmock.NewResult(1, 1))

			LogActionEntryInheritingProvenance(context.Background(), runningStepParams(db), agenterrors.Entry{
				AgentType:    tc.agentType,
				Action:       tc.action,
				ErrorMessage: "msg",
				ErrorCode:    "A_CODE",
				Severity:     "warning",
			}, zap.NewNop())

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("the declared door must keep an explicit agent_type AND still fill step_name: %v", err)
			}
		})
	}
}

// RFC_012 option B's whole win was giving the nine historically
// orchestration_id-less sites their run join for free. Refusing to inherit
// PROVENANCE must not cost the JOIN half — an unattributed row still has to be
// findable from its run, its pod and its work item, or the diagnostic is inert.
func TestLogActionEntry_JoinIdentityIsInheritedEvenWhenProvenanceIsRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"", "",
			"item-1234", // 3
			"orch-1234", // 4
			"unattributed",
			"agent-id-1", // 6
			"pod-1",      // 7
			"", "an_action", "msg", "A_CODE", "warning", sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionEntry(context.Background(), runningStepParams(db), agenterrors.Entry{
		Action:       "an_action",
		ErrorMessage: "msg",
		ErrorCode:    "A_CODE",
		Severity:     "warning",
	}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the join half must survive a refused provenance: %v", err)
	}
}

// An explicit join field beats the params one. reconcile_superseded_reviews and
// complete_work_item_verification both pass a work_item_id that is the PAIR's,
// not the running step's, and read those rows back by it.
func TestLogActionEntry_ExplicitJoinFieldsAreNotOverwritten(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"", "",
			"the-pairs-item", // 3: NOT "item-1234" from input_data
			"orch-1234", "running-agent", "agent-id-1", "pod-1", "running_step",
			"reconcile_superseded_reviews", "msg", "A_CODE", "warning", sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionEntryInheritingProvenance(context.Background(), runningStepParams(db), agenterrors.Entry{
		WorkItemID:   "the-pairs-item",
		Action:       "reconcile_superseded_reviews",
		ErrorMessage: "msg",
		ErrorCode:    "A_CODE",
		Severity:     "warning",
	}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("an explicit work_item_id was overwritten by the merge: %v", err)
	}
}

// The diagnostics go into a COPY. The caller's map is its own state, and in the
// findings path it is shared across every row in the batch — mutating it would
// leak this row's bookkeeping into the caller and into its siblings.
func TestLogActionEntry_DiagnosticsDoNotMutateTheCallersContextMap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectExec(regexp.QuoteMeta(insertRE)).WillReturnResult(sqlmock.NewResult(1, 1))

	callersContext := map[string]interface{}{"page_name": "home"}
	LogActionEntry(context.Background(), runningStepParams(db), agenterrors.Entry{
		Action:       "an_action",
		ErrorMessage: "msg",
		Context:      callersContext,
	}, zap.NewNop())

	if len(callersContext) != 1 {
		t.Fatalf("the caller's context map was mutated: %v", callersContext)
	}
	if _, leaked := callersContext["provenance_missing"]; leaked {
		t.Error("provenance_missing leaked into the caller's own map")
	}
}

// ---------------------------------------------------------------------------
// The findings doors — same contract, and one extra trap of their own.
// ---------------------------------------------------------------------------

// agenterrors.RecordFindings REPLACES the base Context with each Finding's own,
// so a diagnostic left on the base entry is silently dropped for every row.
// The annotation therefore has to be applied per finding.
func TestLogActionEntryFindings_ForgottenProvenanceMarksEveryRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	for _, code := range []string{"FIRST", "SECOND"} {
		mock.ExpectExec(regexp.QuoteMeta(insertRE)).
			WithArgs(
				"", "", "item-1234", "orch-1234",
				"unattributed", "agent-id-1", "pod-1", "",
				"an_action", sqlmock.AnyArg(), code, "warning",
				jsonHas{"provenance_missing", "remedy"},
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	attempted, recorded := LogActionEntryFindings(context.Background(), runningStepParams(db),
		agenterrors.Entry{Action: "an_action"},
		[]agenterrors.Finding{
			{ErrorCode: "FIRST", Severity: "warning", Message: "one"},
			{ErrorCode: "SECOND", Severity: "warning", Message: "two", Context: map[string]interface{}{"k": "v"}},
		}, zap.NewNop())

	if attempted != 2 || recorded != 2 {
		t.Fatalf("got (attempted=%d, recorded=%d), want (2, 2) — the rows must still land", attempted, recorded)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("every unattributed findings row must carry the diagnostic: %v", err)
	}
}

func TestLogActionFindings_StillFilesUnderTheRunningStep(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"site-1", "example.com", "item-1234", "orch-1234",
			"running-agent", "agent-id-1", "pod-1", "running_step",
			"an_action", "one", "FIRST", "info",
			jsonLacks{"provenance_missing"},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionFindings(context.Background(), runningStepParams(db), "site-1", "example.com", "an_action",
		[]agenterrors.Finding{{ErrorCode: "FIRST", Severity: "info", Message: "one"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("LogActionFindings no longer files under the running step: %v", err)
	}
}

// A findings site that names its own provenance keeps it, and the batch is not
// annotated. discovery_checks is the live caller of this door.
func TestLogActionEntryFindings_ExplicitProvenanceIsUsedVerbatim(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(insertRE)).
		WithArgs(
			"", "", "item-1234", "orch-1234",
			"discovery-agent", "agent-id-1", "pod-1", "run_discovery_checks",
			"run_discovery_checks", "one", "FIRST", "warning",
			jsonLacks{"provenance_missing"},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	LogActionEntryFindings(context.Background(), runningStepParams(db), agenterrors.Entry{
		AgentType: "discovery-agent",
		StepName:  "run_discovery_checks",
		Action:    "run_discovery_checks",
	}, []agenterrors.Finding{{ErrorCode: "FIRST", Severity: "warning", Message: "one"}}, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("an explicit findings provenance was not used verbatim: %v", err)
	}
}
