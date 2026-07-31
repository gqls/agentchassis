// FILE: platform/orchestration/actions/diagnose_plan_refusal_test.go
//
// bugs_open/099 candidate 2: a STRUCTURAL plan-validation refusal is recoverable
// when the step names a repair_step, and terminal otherwise.
//
// The first test is the load-bearing one. Three live agents share this action
// (council-gate → persist_submission, feature-designer → persist_plan,
// fix-proposer → persist_plan) and two of them are NOT opted in, so "unchanged
// when repair_step is unset" is a contract, not a detail.
package actions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// invalidStagedPlan is goodStagedPlan with ONE defect: the same file in two edits
// of one stage — exactly the rule bugs_open/099 was filed about. Built by mutating
// the shared good fixture so the plan is invalid for that reason and no other; a
// hand-rolled "invalid" plan can be invalid for reasons the test never intended.
func invalidStagedPlan(t *testing.T) []byte {
	t.Helper()
	p := goodStagedPlan()
	dup := p.Stages[0].Edits[0]
	dup.Symbol = "someOtherSymbol"
	dup.Sketch = "and adjust its caller in the same file"
	p.Stages[0].Edits = append(p.Stages[0].Edits, dup)

	// Assert the fixture really trips the rule, and only it. Without this the
	// end-to-end tests below could be passing because the plan was VALID.
	problems := validateStagedPlan(p, false, testStagedCaps)
	if len(problems) != 1 || !strings.Contains(problems[0], "more than one edit of this stage") {
		t.Fatalf("fixture must fail on exactly the duplicate-file rule, got: %v", problems)
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func validStagedPlanJSON(t *testing.T) []byte {
	t.Helper()
	b, err := json.Marshal(goodStagedPlan())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func refusalParams(t *testing.T, cfg map[string]interface{}, orchID string) (ActionParams, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return ActionParams{
		ExecutionContext: &types.ExecutionContext{OrchestrationID: orchID},
		StepConfig:       models.Step{Config: cfg},
		DB:               db,
		Logger:           zap.NewNop(),
		AgentType:        "feature-designer",
	}, mock, func() { db.Close() }
}

// TestPlanRefusal_TerminalWithoutRepairStep is the regression guard for the two
// consumers that are not opted in: no repair_step ⇒ the same error as before, and
// the DB is never touched (no artefact, no count).
func TestPlanRefusal_TerminalWithoutRepairStep(t *testing.T) {
	params, mock, closeDB := refusalParams(t, map[string]interface{}{}, "orch-1")
	defer closeDB()

	res, err := planValidationRefusal(context.Background(), params, zap.NewNop(),
		"corr-1", "staged plan", invalidStagedPlan(t), []string{"stage 1 edit 2: appears in more than one edit of this stage"})

	if err == nil {
		t.Fatal("want an error with no repair_step configured, got nil")
	}
	if res != nil {
		t.Fatalf("want a nil result on the terminal path, got %v", res)
	}
	if !strings.Contains(err.Error(), "more than one edit") {
		t.Errorf("the error must carry the problems verbatim, got %q", err)
	}
	// No ExpectExec/ExpectQuery were registered, so any DB call is a failure.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the terminal path must not touch the DB: %v", err)
	}
}

// TestPlanRefusal_RoutesToRepairOnFirstFailure is the fix: the design survives.
func TestPlanRefusal_RoutesToRepairOnFirstFailure(t *testing.T) {
	plan := invalidStagedPlan(t)
	params, mock, closeDB := refusalParams(t, map[string]interface{}{
		"repair_step":         "repair_plan",
		"max_repair_attempts": 1,
	}, "orch-1")
	defer closeDB()

	mock.ExpectExec("INSERT INTO diagnosis_artifacts").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM diagnosis_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// The operator-facing record. Added answering the council's bug_historian seat:
	// recoverable is not the same as visible, and 099's original complaint was that
	// the loss was SILENT.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := planValidationRefusal(context.Background(), params, zap.NewNop(),
		"corr-1", "staged plan", plan, []string{"stage 1 edit 2: appears in more than one edit of this stage"})
	if err != nil {
		t.Fatalf("want a recoverable refusal, got error: %v", err)
	}
	out, ok := res.(map[string]interface{})
	if !ok {
		t.Fatalf("want a result map, got %T", res)
	}
	if out["should_repair_plan"] != true {
		t.Error("should_repair_plan must be true so the router can branch")
	}
	if out["plan_valid"] != false || out["persisted"] != false {
		t.Error("a refused plan is neither valid nor persisted")
	}
	if out["repair_attempt"] != 1 {
		t.Errorf("want attempt 1, got %v", out["repair_attempt"])
	}
	// The rejected plan must be recoverable by the repair prompt...
	if got, _ := out["rejected_plan_json"].(string); got != string(plan) {
		t.Error("rejected_plan_json must carry the plan verbatim")
	}
	// ...but must NOT occupy plan_json, which repropose renders as "your previous
	// plan" and which downstream reads as persisted.
	if _, present := out["plan_json"]; present {
		t.Error("a refusal must never set plan_json — repropose renders it as a persisted plan")
	}
	if txt, _ := out["validation_problems_text"].(string); !strings.Contains(txt, "more than one edit") {
		t.Errorf("the problems must reach the repair prompt, got %q", txt)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPlanRefusal_ExhaustsAndGoesTerminal bounds the loop.
func TestPlanRefusal_ExhaustsAndGoesTerminal(t *testing.T) {
	params, mock, closeDB := refusalParams(t, map[string]interface{}{
		"repair_step":         "repair_plan",
		"max_repair_attempts": 1,
	}, "orch-1")
	defer closeDB()

	mock.ExpectExec("INSERT INTO diagnosis_artifacts").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Second refusal of this run: the count now exceeds the cap of 1.
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM diagnosis_artifacts").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	// Written on the TERMINAL outcome too, at severity error — so the row's absence
	// means "no refusal happened", never "it happened and was repaired quietly".
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := planValidationRefusal(context.Background(), params, zap.NewNop(),
		"corr-1", "staged plan", invalidStagedPlan(t), []string{"stage 1 edit 2: duplicate file"})
	if err == nil {
		t.Fatal("want a terminal error once the cap is passed")
	}
	if res != nil {
		t.Fatalf("want a nil result when exhausted, got %v", res)
	}
	if !strings.Contains(err.Error(), "repair attempt") {
		t.Errorf("the error should say the repair budget was spent, got %q", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestPlanRefusal_BookkeepingFailureStillRefuses: a failure in the machinery that
// RECORDS the refusal must not swallow the refusal. Both halves fail closed.
func TestPlanRefusal_BookkeepingFailureStillRefuses(t *testing.T) {
	cases := []struct {
		name  string
		setup func(sqlmock.Sqlmock)
	}{
		{"insert fails", func(m sqlmock.Sqlmock) {
			m.ExpectExec("INSERT INTO diagnosis_artifacts").WillReturnError(errors.New("disk full"))
		}},
		{"count fails", func(m sqlmock.Sqlmock) {
			m.ExpectExec("INSERT INTO diagnosis_artifacts").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectQuery("SELECT count\\(\\*\\) FROM diagnosis_artifacts").WillReturnError(errors.New("gone"))
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params, mock, closeDB := refusalParams(t, map[string]interface{}{
				"repair_step":         "repair_plan",
				"max_repair_attempts": 1,
			}, "orch-1")
			defer closeDB()
			tc.setup(mock)

			res, err := planValidationRefusal(context.Background(), params, zap.NewNop(),
				"corr-1", "plan", invalidStagedPlan(t), []string{"duplicate file"})
			if err == nil {
				t.Fatal("a bookkeeping failure must fail closed, not grant a repair round")
			}
			if res != nil {
				t.Fatalf("want a nil result, got %v", res)
			}
			if !strings.Contains(err.Error(), "duplicate file") {
				t.Errorf("the ORIGINAL validation problem must survive, got %q", err)
			}
		})
	}
}

// TestPlanRefusal_NoOrchestrationIDIsTerminal: without a run id the count cannot
// be scoped to this run, and an unscoped count cannot bound the loop — a NULL
// orchestration_id never satisfies `= $2`, so every attempt would read as the
// first and the repair loop would never end.
func TestPlanRefusal_NoOrchestrationIDIsTerminal(t *testing.T) {
	params, mock, closeDB := refusalParams(t, map[string]interface{}{
		"repair_step":         "repair_plan",
		"max_repair_attempts": 1,
	}, "")
	defer closeDB()

	res, err := planValidationRefusal(context.Background(), params, zap.NewNop(),
		"corr-1", "plan", invalidStagedPlan(t), []string{"duplicate file"})
	if err == nil {
		t.Fatal("want a terminal refusal with no orchestration id")
	}
	if res != nil {
		t.Fatalf("want a nil result, got %v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("must not touch the DB when the loop cannot be bounded: %v", err)
	}
}

// ── end to end, through the real action ──────────────────────────────────────
//
// The helper tests above hand `problems` in, so they cannot prove the action
// actually REACHES the refusal path. These drive DiagnosePersistFixPlanAction.

func runPersistPlan(t *testing.T, planJSON []byte, cfg map[string]interface{}, setup func(sqlmock.Sqlmock)) (map[string]interface{}, error) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	if setup != nil {
		setup(mock)
	}
	// Same shape as the live step config: the value is a PATH into collected data,
	// not the correlation itself.
	cfg["fix_correlation_id"] = "input_data.fix_correlation_id"
	res, err := DiagnosePersistFixPlanAction(context.Background(), ActionParams{
		ExecutionContext: &types.ExecutionContext{OrchestrationID: "orch-e2e"},
		StepConfig:       models.Step{Config: cfg},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"fix_correlation_id": "corr-e2e"},
			"proposal":   map[string]interface{}{"result": string(planJSON)},
		},
		DB:        db,
		Logger:    zap.NewNop(),
		AgentType: "feature-designer",
	})
	out, _ := res.(map[string]interface{})
	if mErr := mock.ExpectationsWereMet(); mErr != nil {
		t.Errorf("unmet expectations: %v", mErr)
	}
	return out, err
}

// TestPersistPlan_E2E_ValidPlanPersistsAndFlagsValid: the happy path still
// persists, and now carries the positive flag the router branches on.
func TestPersistPlan_E2E_ValidPlanPersistsAndFlagsValid(t *testing.T) {
	out, err := runPersistPlan(t, validStagedPlanJSON(t),
		map[string]interface{}{"repair_step": "repair_plan"},
		func(m sqlmock.Sqlmock) {
			m.ExpectExec("INSERT INTO diagnosis_artifacts").WillReturnResult(sqlmock.NewResult(1, 1))
		})
	if err != nil {
		t.Fatalf("a valid staged plan must persist, got: %v", err)
	}
	if out["plan_valid"] != true || out["persisted"] != true {
		t.Errorf("want plan_valid+persisted true, got %v / %v", out["plan_valid"], out["persisted"])
	}
	if out["plan_format"] != "staged-v1" {
		t.Errorf("want the staged discriminator, got %v", out["plan_format"])
	}
	if _, ok := out["plan_json"].(string); !ok {
		t.Error("the council reviewers need plan_json on the success path")
	}
}

// TestPersistPlan_E2E_InvalidPlanRoutesToRepair is the whole point of 099
// candidate 2: the design survives a validator rule instead of being discarded.
func TestPersistPlan_E2E_InvalidPlanRoutesToRepair(t *testing.T) {
	out, err := runPersistPlan(t, invalidStagedPlan(t),
		map[string]interface{}{"repair_step": "repair_plan", "max_repair_attempts": 1},
		func(m sqlmock.Sqlmock) {
			m.ExpectExec("INSERT INTO diagnosis_artifacts").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectQuery("SELECT count\\(\\*\\) FROM diagnosis_artifacts").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			m.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))
		})
	if err != nil {
		t.Fatalf("want a recoverable refusal, got error: %v", err)
	}
	if out["should_repair_plan"] != true {
		t.Errorf("want should_repair_plan true, got %v", out["should_repair_plan"])
	}
	if txt, _ := out["validation_problems_text"].(string); !strings.Contains(txt, "more than one edit") {
		t.Errorf("the real validator's problem must reach the repair prompt, got %q", txt)
	}
}

// TestPersistPlan_E2E_InvalidPlanUnchangedWithoutRepairStep: the same invalid plan,
// no repair_step — the pre-2026-07-30 outcome, which is what council-gate and
// fix-proposer still get.
func TestPersistPlan_E2E_InvalidPlanUnchangedWithoutRepairStep(t *testing.T) {
	out, err := runPersistPlan(t, invalidStagedPlan(t), map[string]interface{}{}, nil)
	if err == nil {
		t.Fatal("want the step to fail, as it did before")
	}
	if out != nil {
		t.Fatalf("want no result, got %v", out)
	}
	if !strings.Contains(err.Error(), "staged plan failed validation") {
		t.Errorf("want the original error text, got %q", err)
	}
}

// TestPersistPlan_E2E_TruncatedJSONStaysTerminal: a CUT completion is a max_tokens
// fault, not a repairable plan, so it must not enter the repair loop even with
// repair_step set (bugs_open/012 and 138's truncation family).
func TestPersistPlan_E2E_TruncatedJSONStaysTerminal(t *testing.T) {
	truncated := []byte(`{"plan_format":"staged-v1","summary":"cut off here`)
	out, err := runPersistPlan(t, truncated,
		map[string]interface{}{"repair_step": "repair_plan", "max_repair_attempts": 1}, nil)
	if err == nil {
		t.Fatal("truncated JSON must stay terminal")
	}
	if out != nil {
		t.Fatalf("want no result, got %v", out)
	}
	if !strings.Contains(err.Error(), "truncated") {
		t.Errorf("the error should name truncation, got %q", err)
	}
}

// TestPlanRefusal_ZeroAttemptCapIsTerminal: repair_step named but a cap of 0 means
// no repair rounds, so the refusal is terminal and nothing is written.
func TestPlanRefusal_ZeroAttemptCapIsTerminal(t *testing.T) {
	params, mock, closeDB := refusalParams(t, map[string]interface{}{
		"repair_step":         "repair_plan",
		"max_repair_attempts": 0,
	}, "orch-1")
	defer closeDB()

	_, err := planValidationRefusal(context.Background(), params, zap.NewNop(),
		"corr-1", "plan", invalidStagedPlan(t), []string{"duplicate file"})
	if err == nil {
		t.Fatal("want a terminal refusal when the cap is 0")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("must not touch the DB when no repair round is allowed: %v", err)
	}
}

// TestCountRunArtifacts_MetadataFilterIsOptional covers the one piece of branching
// logic in the extracted counter: with metaKey empty the query must carry NO
// metadata predicate (the council-decide round counter), and with it set the
// predicate must be bound as parameters (the reframe counter and the repair
// counter). A filter silently dropped would over-count and hand out extra rounds.
func TestCountRunArtifacts_MetadataFilterIsOptional(t *testing.T) {
	t.Run("no filter", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT count\\(\\*\\) FROM diagnosis_artifacts").
			WithArgs("corr", "orch", "council_report").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		n, err := countRunArtifacts(context.Background(), db, "corr", "orch", "council_report", "", "")
		if err != nil || n != 3 {
			t.Fatalf("want 3, nil; got %d, %v", n, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("three args only when unfiltered: %v", err)
		}
	})

	t.Run("with filter", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("metadata->>").
			WithArgs("corr", "orch", "iteration_note", "note_kind", planRefusalNoteKind).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		n, err := countRunArtifacts(context.Background(), db, "corr", "orch",
			"iteration_note", "note_kind", planRefusalNoteKind)
		if err != nil || n != 1 {
			t.Fatalf("want 1, nil; got %d, %v", n, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("filter must be bound as params: %v", err)
		}
	})

	t.Run("error propagates rather than returning a usable zero", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		defer db.Close()
		mock.ExpectQuery("SELECT count").WillReturnError(errors.New("gone"))
		if _, err := countRunArtifacts(context.Background(), db, "c", "o", "k", "", ""); err == nil {
			t.Fatal("a count failure must be an error — every caller here fails closed on it")
		}
	})
}

// TestPersistPlan_E2E_SchemaMismatchIsRepairable: a plan whose JSON is WELL FORMED
// but the wrong shape is a model formatting error, not a truncation, so it is
// recoverable. Widened 2026-07-31 on live evidence: the first real repair round
// emitted `"file": [...]` where the schema wants a string, and under the previous
// code that killed the run outright — after the repair budget had already been spent.
//
// The pair with TruncatedJSONStaysTerminal is the point: same action, same config,
// and the two must go opposite ways.
func TestPersistPlan_E2E_SchemaMismatchIsRepairable(t *testing.T) {
	// Exactly the live shape: edits[].file as an array instead of a string.
	wrongShape := []byte(`{"plan_format":"staged-v1","summary":"s","grounded_in":["g"],
	  "stages":[{"id":"s1","title":"t","goal":"g","edits":[
	    {"file":["a/b.go","a/c.go"],"operation":"modify","rationale":"r","sketch":"k"}]}]}`)
	if !json.Valid(wrongShape) {
		t.Fatal("fixture must be VALID json — otherwise it tests the truncation path")
	}

	out, err := runPersistPlan(t, wrongShape,
		map[string]interface{}{"repair_step": "repair_plan", "max_repair_attempts": 1},
		func(m sqlmock.Sqlmock) {
			m.ExpectExec("INSERT INTO diagnosis_artifacts").WillReturnResult(sqlmock.NewResult(1, 1))
			m.ExpectQuery("SELECT count\\(\\*\\) FROM diagnosis_artifacts").
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			m.ExpectExec("INSERT INTO agent_error_log").WillReturnResult(sqlmock.NewResult(1, 1))
		})
	if err != nil {
		t.Fatalf("a schema mismatch must be recoverable, got error: %v", err)
	}
	if out["should_repair_plan"] != true {
		t.Errorf("want routing to repair, got %v", out["should_repair_plan"])
	}
	txt, _ := out["validation_problems_text"].(string)
	if !strings.Contains(txt, "does not match the staged-plan schema") {
		t.Errorf("the repair prompt must be told what the shape error was, got %q", txt)
	}
	// The model needs the field name to fix it.
	if !strings.Contains(txt, "file") {
		t.Errorf("the unmarshal error names the offending field; it must survive into the prompt, got %q", txt)
	}
}

// TestPersistPlan_SchemaMismatchTerminalWithoutRepairStep: the widening must not
// change anything for the two consumers that are not opted in.
func TestPersistPlan_SchemaMismatchTerminalWithoutRepairStep(t *testing.T) {
	wrongShape := []byte(`{"plan_format":"staged-v1","stages":[{"id":"s1","edits":[{"file":["a"]}]}]}`)
	out, err := runPersistPlan(t, wrongShape, map[string]interface{}{}, nil)
	if err == nil {
		t.Fatal("want the step to fail with no repair_step")
	}
	if out != nil {
		t.Fatalf("want no result, got %v", out)
	}
	if !strings.Contains(err.Error(), "does not match the staged-plan schema") {
		t.Errorf("want the schema error, got %q", err)
	}
}
