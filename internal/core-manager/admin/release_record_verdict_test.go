// FILE: internal/core-manager/admin/release_record_verdict_test.go
//
// bugs_open/428: HandleReleaseRecordVerdict is the human half of RFC_056's
// filing_mode='record' circuit breaker (write_audit_findings_action.go) — it
// must release ONE row a person has reviewed, and refuse everything else by
// construction, because the whole reason this endpoint exists is that an
// earlier standing promoter for this exact finding class destroyed live
// content (bugs_closed/238).

package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func newReleaseMock(t *testing.T) (*SiteAdminHandlers, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewSiteAdminHandlers(db, zap.NewNop()), mock, db
}

func doRelease(t *testing.T, h *SiteAdminHandlers, itemID uuid.UUID, body map[string]interface{}) (int, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	b, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(http.MethodPost, "/work-items/"+itemID.String()+"/release", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "item_id", Value: itemID.String()}}

	h.HandleReleaseRecordVerdict(c)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return w.Code, resp
}

// TestReleaseRecordVerdict_Success — the row matches every precondition;
// status/handler_agent transition to the values RFC_056 stashed in spec.
func TestReleaseRecordVerdict_Success(t *testing.T) {
	h, mock, _ := newReleaseMock(t)
	itemID := uuid.New()

	mock.ExpectQuery(`UPDATE site_work_items`).
		WithArgs(itemID, "aaa@designconsultancy.co.uk", "checked — this is a real gap, not an aspirational rewrite").
		WillReturnRows(sqlmock.NewRows([]string{"status", "handler_agent"}).
			AddRow("detected", "content-gap-planner"))

	code, resp := doRelease(t, h, itemID, map[string]interface{}{
		"released_by": "aaa@designconsultancy.co.uk",
		"notes":       "checked — this is a real gap, not an aspirational rewrite",
	})
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %+v", code, resp)
	}
	if resp["released"] != true || resp["status"] != "detected" || resp["handler_agent"] != "content-gap-planner" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestReleaseRecordVerdict_RequiresReleasedBy — this is a human decision on a
// per-row basis by design; refusing an anonymous release BEFORE touching the
// database is what makes that a real constraint rather than a convention.
func TestReleaseRecordVerdict_RequiresReleasedBy(t *testing.T) {
	h, _, db := newReleaseMock(t)
	db.Close() // any query at all is a failure for this test

	code, resp := doRelease(t, h, uuid.New(), map[string]interface{}{"notes": "no releaser named"})
	if code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %+v", code, resp)
	}
}

// TestReleaseRecordVerdict_NoMatchIsNotFound covers every shape of "this row
// must not be released": wrong status, wrong filing_mode, or the routed_*
// fields missing. All three collapse to 0 rows affected by the same WHERE
// clause, and all three must read as 404 — the caller cannot distinguish
// "already released" from "was never a record verdict" from this response,
// which is the safe default for an endpoint whose whole job is refusing
// almost everything it is pointed at.
func TestReleaseRecordVerdict_NoMatchIsNotFound(t *testing.T) {
	h, mock, _ := newReleaseMock(t)
	itemID := uuid.New()

	mock.ExpectQuery(`UPDATE site_work_items`).
		WithArgs(itemID, "reviewer", "").
		WillReturnError(sql.ErrNoRows)

	code, resp := doRelease(t, h, itemID, map[string]interface{}{"released_by": "reviewer"})
	if code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404, body = %+v", code, resp)
	}
}

// TestReleaseRecordVerdict_QueryCarriesEveryPrecondition — a mutation-style
// guard: the WHERE clause's job is to be the ONLY thing standing between this
// endpoint and rebuilding the promoter RFC_056 removed, so pin the presence
// of each predicate by regex rather than trusting the happy-path test alone
// to have exercised the real query (sqlmock's default matcher is substring,
// so a happy-path test passes even against a query missing a predicate, as
// long as the mocked row still comes back).
func TestReleaseRecordVerdict_QueryCarriesEveryPrecondition(t *testing.T) {
	h, mock, _ := newReleaseMock(t)
	itemID := uuid.New()

	for _, needle := range []string{
		`status = 'deferred'`,
		`spec->>'filing_mode' = 'record'`,
		`COALESCE\(spec->>'routed_handler', ''\) <> ''`,
		`COALESCE\(spec->>'routed_status', ''\) <> ''`,
		`status = spec->>'routed_status'`,
		`handler_agent = spec->>'routed_handler'`,
	} {
		t.Run(needle, func(t *testing.T) {
			h2, mock2, _ := newReleaseMock(t)
			_ = h
			_ = mock
			mock2.ExpectQuery(needle).
				WithArgs(itemID, "reviewer", "").
				WillReturnRows(sqlmock.NewRows([]string{"status", "handler_agent"}).AddRow("detected", "x"))
			code, resp := doRelease(t, h2, itemID, map[string]interface{}{"released_by": "reviewer"})
			if code != http.StatusOK {
				t.Fatalf("query does not contain required predicate %q — status=%d body=%+v", needle, code, resp)
			}
		})
	}
}
