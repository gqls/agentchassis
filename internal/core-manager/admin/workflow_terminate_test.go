// FILE: internal/core-manager/admin/workflow_terminate_test.go
//
// Covers HandleResumeWorkflow's terminate arm and getWorkflowState's row
// selection — the two findings of the 2026-08-25 post-roll review
// (web_admin_console PLAN §7, F1 + F2).
//
// The defects these tests exist to prevent:
//
//  F1: correlation_id is NOT unique (sub-orchestrations share their parent's;
//      up to 19 rows measured under one). An unscoped terminate UPDATE
//      relabelled every sibling — including COMPLETED sub-orchestrations — as
//      FAILED. The UPDATE must carry a non-terminal status filter, and "every
//      row already terminal" must answer 409 ALREADY_TERMINAL, not 500.
//
//  F2: getWorkflowState selected one of those rows with no ORDER BY — an
//      ARBITRARY sibling, different between two opens of the same detail view.
//      The query must pin root-first, oldest-first, LIMIT 1.
//
// The regexes here are the contract: sqlmock matches expected SQL as a regular
// expression against the query the handler actually issues, so deleting the
// status filter or the ORDER BY makes these tests fail on a non-matching query.

package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const terminateTestCorrelation = "4de13ffb-94e7-430a-aecb-d1dc156bd6d4"

// The full column set getWorkflowState scans, in order.
func workflowStateRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"correlation_id", "orchestration_id", "client_id", "status",
		"current_step", "awaited_steps", "collected_data",
		"initial_request_data", "final_result", "error",
		"created_at", "updated_at",
	}).AddRow(
		terminateTestCorrelation, "7d9a1f00-0000-0000-0000-0000000000aa",
		"demo_client", "EXECUTING_STEP", "render", []byte(`[]`), []byte(`{}`),
		[]byte(`{}`), nil, nil, time.Now(), time.Now(),
	)
}

// getWorkflowState must select deterministically: root first, oldest first,
// exactly one row. Deleting the ORDER BY or the LIMIT breaks this match.
const workflowStateQueryRE = `SELECT correlation_id, orchestration_id[\s\S]*FROM orchestration_states[\s\S]*WHERE correlation_id = \$1[\s\S]*ORDER BY \(parent_orchestration_id IS NOT NULL\), created_at[\s\S]*LIMIT 1`

// The terminate UPDATE must never touch a terminal row — terminal by the Go
// constants OR by orchestration_status_vocabulary.is_terminal (belt and
// braces: CANCELLED is terminal by convention and exists only in the
// vocabulary; an emptied vocabulary leaves the literal set standing).
// Deleting either half of the filter breaks this match.
const terminateUpdateRE = `UPDATE orchestration_states[\s\S]*WHERE correlation_id = \$1[\s\S]*AND status NOT IN \('COMPLETED', 'FAILED'\)[\s\S]*AND NOT EXISTS \([\s\S]*FROM orchestration_status_vocabulary v[\s\S]*v\.is_terminal[\s\S]*\)`

func doResume(t *testing.T, h *SystemHandlers, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "correlation_id", Value: terminateTestCorrelation}}
	c.Request = httptest.NewRequest(http.MethodPost,
		"/admin/workflows/"+terminateTestCorrelation+"/resume", strings.NewReader(body))
	h.HandleResumeWorkflow(c)
	return w
}

func TestTerminateScopesToNonTerminalRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSystemHandlers(db, nil, nil, zap.NewNop())

	mock.ExpectQuery(workflowStateQueryRE).WillReturnRows(workflowStateRows())
	mock.ExpectExec(terminateUpdateRE).
		WillReturnResult(sqlmock.NewResult(0, 1))

	w := doResume(t, h, `{"action": "terminate"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTerminateAllTerminalIs409NotError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSystemHandlers(db, nil, nil, zap.NewNop())

	mock.ExpectQuery(workflowStateQueryRE).WillReturnRows(workflowStateRows())
	// The correlation exists but every row under it is already terminal:
	// the scoped UPDATE matches nothing.
	mock.ExpectExec(terminateUpdateRE).
		WillReturnResult(sqlmock.NewResult(0, 0))

	w := doResume(t, h, `{"action": "terminate"}`)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body = %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["code"] != "ALREADY_TERMINAL" {
		t.Errorf("code = %v, want ALREADY_TERMINAL", body["code"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestTerminateUnknownCorrelationIs404(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	h := NewSystemHandlers(db, nil, nil, zap.NewNop())

	mock.ExpectQuery(workflowStateQueryRE).
		WillReturnRows(sqlmock.NewRows([]string{"correlation_id"}))

	w := doResume(t, h, `{"action": "terminate"}`)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body = %s", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
