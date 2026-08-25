package actions

// bugs_open/345 (skip-empty guard, at bugs_open/395's request): recordRetryFeedback
// REPLACES site_work_items.retry_feedback wholesale, and the reader keys on a
// non-blank message. So a blank message must be a NO-OP, never a write — a blank
// write clobbers a specific producer's feedback with {"message":""}, which the
// reader then drops, sending the retry blind. These tests pin both directions.
//
// The blank case is asserted NON-VACUOUSLY (the 016b/LANDMINES trap: a test that
// merely asserts "no query issued" passes even when the function swallows the
// mock's error). Here a real ExpectExec is SET and ExpectationsWereMet is required
// to FAIL for the blank input: if the guard is removed, the UPDATE fires, the
// expectation is met, ExpectationsWereMet returns nil, and the test fails. Proven:
// deleting the `TrimSpace(message)==""` guard flips both assertions.

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const rfWorkItem = "11111111-1111-1111-1111-111111111111"

func TestRecordRetryFeedback_BlankMessageIsNeverWritten(t *testing.T) {
	for _, blank := range []string{"", "   ", "\t\n"} {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock: %v", err)
		}
		// Set the expectation the write WOULD satisfy. If the guard is gone, the
		// UPDATE fires and fulfils this — which we then detect as a FAILURE.
		mock.ExpectExec("UPDATE site_work_items").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		recordRetryFeedback(context.Background(), db, zap.NewNop(),
			rfWorkItem, "component_validation_rejected", blank, "orch", "store_component")

		if err := mock.ExpectationsWereMet(); err == nil {
			t.Fatalf("blank message %q issued an UPDATE — the guard did not skip; a blank write clobbers real feedback to blind", blank)
		}
		db.Close()
	}
}

func TestRecordRetryFeedback_RealMessageIsWritten(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// $1 id, $2 code, $3 message, $4 orch, $5 step — pin code and message, the
	// two the reader consumes.
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(rfWorkItem, "component_validation_orphan_schema_field",
			`schema field "x" has no template variable`, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	recordRetryFeedback(context.Background(), db, zap.NewNop(),
		rfWorkItem, "component_validation_orphan_schema_field",
		`schema field "x" has no template variable`, "orch", "store_component")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("a real message did not issue the expected UPDATE: %v", err)
	}
}
