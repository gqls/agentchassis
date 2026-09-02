// FILE: platform/orchestration/actions/chrome_failure_report_utf8_test.go
//
// bugs_open/423's latent sibling, and the reason it had to be fixed in the same
// pass as half 1: emitChromeRenderFailedItem builds its summary by interpolating
// an ARBITRARY error string, then capped it with a raw byte slice
// (`summary[:247]`). Once the store-failure and UTF-8-refusal branches started
// using this surface, that slice sat on the path that REPORTS a chrome failure —
// so a long error whose 248th byte fell inside a multi-byte character would mint
// invalid UTF-8 and fail its own INSERT. The surface would have died of exactly
// the disease it exists to report, and (because every failure here is logged
// rather than returned) it would have done so silently.
//
// This test DRIVES THE REAL EMITTER rather than calling insertWorkItem with a
// prepared summary — the vacuous-test trap cta_override_rejected_item_test.go
// documents. sqlmock rejects an argument the matcher refuses, so an emitter
// reverted to the byte slice leaves the expectation unmet and this fails.
package actions

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// validUTF8Arg matches any string argument that is valid UTF-8 — i.e. any
// argument Postgres would actually accept. A non-string, or a string carrying a
// severed rune, fails to match and the expectation goes unmet.
type validUTF8Arg struct{}

func (validUTF8Arg) Match(v driver.Value) bool {
	s, ok := v.(string)
	return ok && utf8.ValidString(s)
}

// MUTATION KILLED: `summary = datahelpers.SafeCut(summary, 247) + "..."`
// reverted to `summary = summary[:247] + "..."`.
//
// The input is built so the cap lands INSIDE a multi-byte character: the
// emitter's own prefix ("Chrome footer <phase>: ") plus padding places an
// em-dash across bytes 245-247, so a byte slice at 247 keeps two thirds of it.
func TestChromeFailureSummaryIsRuneSafe(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	const slot, phase = "footer", "rendered but was not stored"
	prefix := fmt.Sprintf("Chrome %s %s: ", slot, phase)
	if len(prefix) >= 245 {
		t.Fatalf("test premise broken: the emitter prefix is already %d bytes", len(prefix))
	}
	// Pad so the em-dash occupies bytes 245, 246 and 247 of the summary, then
	// overflow the 250-byte threshold so the cap actually fires.
	errText := strings.Repeat("a", 245-len(prefix)) + "—" + strings.Repeat("b", 60)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			validUTF8Arg{}, // $6 — the summary Postgres has to accept
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	emitChromeRenderFailedItem(context.Background(), db, uuid.New(), uuid.New(),
		slot, phase, fmt.Errorf("%s", errText), true, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the chrome-failure report could not be written — the reporting "+
			"path minted invalid UTF-8 while truncating its own summary "+
			"(bugs_open/423): %v", err)
	}
}
