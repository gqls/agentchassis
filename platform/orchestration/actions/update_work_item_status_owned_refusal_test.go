package actions

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// Owner decision 1, 2026-08-18: an owned-page REFUSAL must record something
// other than `failed`. bugs_open/301 §3, bugs_open/083.
//
// WHY THIS MATTERS ENOUGH TO TEST. The detected-item-promoter's floor counts
// complete+verified against failed, per (item_type, handler_agent), and holds
// the pair below 25%. It names no other status — read from the live pre_query
// 2026-08-18. So `failed` on an ownership refusal is a vote in a competence
// measure the refusal says nothing about, and a pair crossing the floor stops
// being dispatched at all, including on the generic pages where it was working
// (phantom_internal_link: 69% generic, 0/14 owned, 47% overall).
//
// The three assertions that matter are a set, not a list. The downgrade alone
// proves nothing: a rule that fires on everything would pass it. The default-OFF
// case and the genuine-failure case are what show the field DISCRIMINATES —
// without them this test passes on an implementation that marks every failure
// wont_fix and blinds the floor completely, which is worse than the bug.

// ownedRefusalRun is what one drive of the action wrote.
type ownedRefusalRun struct {
	status    string
	resultJSN string
	errCol    string
}

// captureTextArg is captureArg's sibling for an argument that reaches the driver
// as []byte rather than string — the result payload is json.Marshal output, so
// captureArg silently records "" for it and every assertion on the JSONB then
// passes vacuously against an empty string.
type captureTextArg struct{ got *string }

func (m captureTextArg) Match(v driver.Value) bool {
	switch t := v.(type) {
	case string:
		*m.got = t
	case []byte:
		*m.got = string(t)
	}
	return true
}

// runStatusUpdate drives UpdateWorkItemStatusAction against sqlmock and returns
// the status, result JSONB and error column it wrote.
func runStatusUpdate(t *testing.T, config map[string]interface{}, collected map[string]interface{}) ownedRefusalRun {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	itemID := uuid.New()
	if collected == nil {
		collected = map[string]interface{}{}
	}
	if _, ok := collected["input_data"]; !ok {
		collected["input_data"] = map[string]interface{}{"work_item_id": itemID.String()}
	}

	var got ownedRefusalRun
	mock.ExpectExec(`UPDATE site_work_items`).
		WithArgs(itemID,
			captureArg{got: &got.status},
			captureTextArg{got: &got.resultJSN},
			captureArg{got: &got.errCol}).
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: config},
	}

	if _, err := UpdateWorkItemStatusAction(context.Background(), params); err != nil {
		t.Fatalf("UpdateWorkItemStatusAction: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
	return got
}

// ownedRefusalError is the live shape of the refusal as it reaches the routed
// step: SavePageSectionsAction's message, copied verbatim into
// __step_error.message by routeToErrorStep. Built from the constant rather than
// hard-coded so that renaming the marker breaks this test rather than silently
// switching the downgrade off in production.
func ownedRefusalError(t *testing.T) map[string]interface{} {
	t.Helper()
	msg := ownedPageSkipReasonPrefix + ": page tool-archetype-taster-quiz is rebuild_policy=owned " +
		"(tool/widget-owned): a generic section save would clobber it. Use apply_section_edit for " +
		"targeted edits or the tool pipeline for rebuilds. Refusing to overwrite."
	return map[string]interface{}{
		"__step_error": map[string]interface{}{
			"failed_step": "save_sections",
			"message":     msg,
		},
	}
}

// genuineSaveFailure is a real save failure on the SAME step — the negative
// control. The shrink guard refusing a near-wipe is a handler that tried and
// could not, and the floor must keep counting it.
func genuineSaveFailure() map[string]interface{} {
	return map[string]interface{}{
		"__step_error": map[string]interface{}{
			"failed_step": "save_sections",
			"message": "step save_sections failed: failed to execute action save_page_sections: " +
				"content regression guard: new sections carry 412 chars against 3,180 stored (13%)",
		},
	}
}

func TestUpdateWorkItemStatus_OwnedPageRefusalIsNotAFailure(t *testing.T) {
	withRefusal := map[string]interface{}{
		"status":                    "failed",
		"owned_page_refusal_status": "wont_fix",
	}
	// page-build-handler's mark_item_failed as it stands today, unchanged.
	asConfiguredToday := map[string]interface{}{"status": "failed"}

	t.Run("opted in, ownership refusal — records wont_fix", func(t *testing.T) {
		got := runStatusUpdate(t, withRefusal, ownedRefusalError(t))
		if got.status != "wont_fix" {
			t.Errorf("status: got %q, want %q — the refusal is still voting in the promoter's floor", got.status, "wont_fix")
		}
		// The reason must survive the substitution: triage reads the row, not
		// the log, and a wont_fix with no error column is unreadable.
		if !strings.Contains(got.errCol, ownedPageSkipReasonPrefix) {
			t.Errorf("error column lost the refusal reason: %q", got.errCol)
		}
		// Auditable on the row, or a census cannot tell this apart from a
		// human's wont_fix.
		if !strings.Contains(got.resultJSN, `"owned_page_refusal":true`) {
			t.Errorf("result JSONB carries no owned_page_refusal stamp: %s", got.resultJSN)
		}
		if !strings.Contains(got.resultJSN, `"owned_page_refusal_replaced_status":"failed"`) {
			t.Errorf("result JSONB does not record what the status would have been: %s", got.resultJSN)
		}
	})

	t.Run("field absent — an ownership refusal still records failed (default OFF)", func(t *testing.T) {
		got := runStatusUpdate(t, asConfiguredToday, ownedRefusalError(t))
		if got.status != "failed" {
			t.Errorf("status: got %q, want %q — the unsafe default must be OFF for every caller that has not opted in", got.status, "failed")
		}
		if strings.Contains(got.resultJSN, "owned_page_refusal") {
			t.Errorf("stamped a refusal on a caller that never opted in: %s", got.resultJSN)
		}
	})

	t.Run("opted in, genuine save failure — still records failed", func(t *testing.T) {
		got := runStatusUpdate(t, withRefusal, genuineSaveFailure())
		if got.status != "failed" {
			t.Errorf("status: got %q, want %q — a handler that tried and could not must keep counting, or the floor goes blind", got.status, "failed")
		}
		if strings.Contains(got.resultJSN, "owned_page_refusal") {
			t.Errorf("stamped a refusal on a real failure: %s", got.resultJSN)
		}
	})

	t.Run("opted in, no routed error at all — status is whatever was configured", func(t *testing.T) {
		// The success path through the same action (mark_complete). There is no
		// __step_error, so there is nothing to match and nothing to downgrade.
		got := runStatusUpdate(t, map[string]interface{}{
			"status":                    "complete",
			"owned_page_refusal_status": "wont_fix",
		}, map[string]interface{}{})
		if got.status != "complete" {
			t.Errorf("status: got %q, want %q", got.status, "complete")
		}
	})
}

// A misconfigured refusal status must fail loudly at the step rather than write
// a status the vocabulary does not contain. It shares validStatuses with
// `status` on purpose — two vocabularies for one column is the drift class.
func TestUpdateWorkItemStatus_RefusalStatusIsValidated(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"work_item_id": uuid.New().String()},
		},
		StepConfig: models.Step{Config: map[string]interface{}{
			"status":                    "failed",
			"owned_page_refusal_status": "wontfix", // the plausible typo
		}},
	}

	if _, err := UpdateWorkItemStatusAction(context.Background(), params); err == nil {
		t.Fatal("expected an error for an invalid owned_page_refusal_status, got nil")
	}
	// And nothing was written.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
