// FILE: platform/orchestration/actions/complete_work_item_unreadable_wiring_test.go
//
// The WIRING test for gate 1b's unreadable-payload refusal (bugs_open/302).
//
// WHY IT IS A SEPARATE FILE AND WHY IT EXISTS AT ALL. The logic test beside it
// (complete_work_item_no_change_test.go) proves handlerReportedNoChange returns the
// right verdict. It cannot prove that verifyBeforeComplete ASKS — and register entry
// WII-017 records exactly that gap in the original gate ("no unit test asserts that
// verifyBeforeComplete calls the gate"). A pure function returning the correct answer
// into a caller that ignores it is the shape of a guard that reads as protection and
// is not one, which is this estate's most repeated finding about its own tests.
//
// So these tests go through verifyBeforeComplete, with a real *sql.DB (sqlmock) for
// the item-row SELECT, and assert the CALLER's two outputs: mayComplete, and the
// payload that lands at result._verification.
//
// THE CONTROL THAT MATTERS MOST is TestVerifyBeforeComplete_ReadableZeroStillBlocksAsNoChange
// plus TestVerifyBeforeComplete_UnregisteredTypeCompletes: a gate that refused
// everything, or a caller that blocked on any non-nil payload, would satisfy the
// refusal test on its own. The three together discriminate.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// itemRow stubs the one query verifyBeforeComplete makes before the gates run.
// spec is '{}' and page_id NULL: this file is about gate 1b, which reads neither.
func itemRow(mock sqlmock.Sqlmock, itemType string) {
	mock.ExpectQuery("SELECT item_type").
		WillReturnRows(sqlmock.NewRows([]string{"item_type", "spec", "site_id", "page_id"}).
			AddRow(itemType, []byte(`{}`), uuid.New(), nil))
}

func TestVerifyBeforeComplete_UnreadableRefusesBlocks(t *testing.T) {
	// The live spawn-record payload — the shape that reversed this gate's own
	// refusal on 5 items between 08-14 and 08-17.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	itemRow(mock, "dark_section_audit")

	payload, mayComplete, abstained := verifyBeforeComplete(
		context.Background(), db, uuid.New(), spawnRecordPayload(), zap.NewNop())

	if mayComplete {
		t.Fatal("mayComplete = true on an unreadable payload for a type declaring unreadableRefuses — " +
			"the completion the whole change exists to stop")
	}
	if abstained != nil {
		t.Fatalf("abstained = %+v; a REFUSAL must not also be recorded as an abstention — the blocked row is the record", abstained)
	}
	if got := payload["status"]; got != "handler_result_unreadable" {
		t.Fatalf("_verification.status = %v, want handler_result_unreadable (a distinct status is what "+
			"blockedCompletionReason keys the operator's message on)", got)
	}
	if got := payload["item_type"]; got != "dark_section_audit" {
		t.Fatalf("_verification.item_type = %v, want dark_section_audit", got)
	}
	// The message an operator actually reads must carry the licence, or a blocked
	// item arrives with no way to check the claim that blocked it.
	msg, reason := blockedCompletionReason(payload)
	if reason != "handler_result_unreadable" {
		t.Fatalf("reason code = %q, want handler_result_unreadable", reason)
	}
	if !strings.Contains(msg, "unreadable to the no-change gate") || !strings.Contains(msg, "agent_id agent_type role topics") {
		t.Fatalf("operator message does not say what happened or what was in the payload: %q", msg)
	}
}

// TestVerifyBeforeComplete_ReadableZeroStillBlocksAsNoChange is the discrimination
// control: the two blocking arms must stay distinct through the caller. If a future
// edit collapsed them, the operator would be handed a finding no gate made.
func TestVerifyBeforeComplete_ReadableZeroStillBlocksAsNoChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	itemRow(mock, "dark_section_audit")

	payload, mayComplete, _ := verifyBeforeComplete(
		context.Background(), db, uuid.New(), fixerEnvelope(float64(0), float64(0)), zap.NewNop())

	if mayComplete {
		t.Fatal("mayComplete = true on a readable all-zero payload — gate 1b's original job")
	}
	if got := payload["status"]; got != "handler_reported_no_change" {
		t.Fatalf("_verification.status = %v, want handler_reported_no_change — a readable zero is a "+
			"different claim from an unreadable payload and must keep its own status", got)
	}
}

// TestVerifyBeforeComplete_UnregisteredTypeCompletes is the containment control, and
// the one that would catch the worst possible regression: a type that never opted in
// must be untouched by any of this. If it ever fails, every item type in the fleet is
// affected, not just the roster.
func TestVerifyBeforeComplete_UnregisteredTypeCompletes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	// An item_type with neither a roster entry nor a registered verifier, given the
	// payload that is refused above.
	itemRow(mock, "needs_content_planning")

	payload, mayComplete, abstained := verifyBeforeComplete(
		context.Background(), db, uuid.New(), spawnRecordPayload(), zap.NewNop())

	if !mayComplete {
		t.Fatal("mayComplete = false for an item_type that never opted in — the unsafe-default-OFF ruling breached")
	}
	if payload != nil {
		t.Fatalf("_verification written for an un-opted-in type: %+v", payload)
	}
	if abstained != nil {
		t.Fatalf("abstention recorded for an un-opted-in type: %+v", abstained)
	}
}

// TestVerifyBeforeComplete_AbstainDeclarationStillCompletes proves the per-type
// declaration is honoured through the caller in BOTH directions, not just the new
// one — same unreadable payload, a type declaring abstain, opposite outcome, and the
// abstention handed back for recording.
func TestVerifyBeforeComplete_AbstainDeclarationStillCompletes(t *testing.T) {
	withRosterEntry(t, "test_wiring_abstains", noChangeRule{
		Why:          "synthetic: fixture for the abstain declaration through the caller",
		CounterPaths: []string{"response.fix_result.total_fixed"},
		OnUnreadable: unreadableAbstains,
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	itemRow(mock, "test_wiring_abstains")

	payload, mayComplete, abstained := verifyBeforeComplete(
		context.Background(), db, uuid.New(), spawnRecordPayload(), zap.NewNop())

	if !mayComplete {
		t.Fatal("mayComplete = false for a type declaring unreadableAbstains — the declaration is not being read")
	}
	if abstained == nil {
		t.Fatal("no abstention returned — an unreadable payload that completes MUST be recorded, or the gate " +
			"goes blind with no queryable trace (recordUnknownNoChangeShape's whole reason for existing)")
	}
	if abstained.ItemType != "test_wiring_abstains" {
		t.Fatalf("abstention carries item_type %q, want test_wiring_abstains", abstained.ItemType)
	}
	if payload != nil {
		t.Fatalf("_verification written on an abstention: %+v", payload)
	}
}
