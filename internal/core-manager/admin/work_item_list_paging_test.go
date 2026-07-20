// FILE: internal/core-manager/admin/work_item_list_paging_test.go
//
// Covers HandleListWorkItems' paging and counting contract.
//
// The defect these tests exist to prevent (bugs_open/033): the handler returned
// a hardcoded 50 rows with no total, and the dashboard filtered that window
// client-side and derived its status counts from it. With 687 open build items,
// the newest 50 contained no needs_human_review rows at all — so a 208-item
// backlog rendered as an empty list AND reported its own count as 0, for four
// months. A cap on a READ path corrupts the absence of evidence, which is what
// people reason from when deciding whether a problem exists.
//
// Two invariants follow, and both are easy to regress by "tidying" the query
// assembly:
//
//  1. The count queries must NOT carry the status or item_type predicates. A
//     count scoped by the filter it populates collapses its own dropdown to the
//     option already selected, with no way back to the others. (I wrote that bug
//     while fixing this one; hence the test.)
//  2. The page query MUST carry them, and must always be accompanied by a total,
//     so a truncated window can never again be mistaken for the whole set.

package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func newListMock(t *testing.T) (*SiteAdminHandlers, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	return NewSiteAdminHandlers(db, zap.NewNop()), mock, db
}

// doList runs the handler against a request URL and returns the decoded body.
func doList(t *testing.T, h *SiteAdminHandlers, url string) map[string]interface{} {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, url, nil)

	h.HandleListWorkItems(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v (body %s)", err, w.Body.String())
	}
	return body
}

// statusCountRows / typeCountRows / itemRows build the three result sets the
// handler reads, in the order it reads them.
func statusCountRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"status", "count"}).
		AddRow("needs_human_review", 208).
		AddRow("complete", 4647).
		AddRow("failed", 155)
}

func typeCountRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"item_type", "count"}).
		AddRow("unresolved_cta", 66).
		AddRow("voice_tells", 25)
}

func emptyItemRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "site_id", "domain", "item_type", "status",
		"severity", "summary", "spec",
		"handler_agent", "attempt_count", "max_attempts",
		"error", "created_at", "completed_at",
	})
}

// The load-bearing one: with BOTH a status and an item_type filter applied, the
// two count queries must still be scoped only by pipeline — no wi.status and no
// wi.item_type predicate — while the page query carries both.
func TestListWorkItems_CountsIgnoreStatusAndTypeFilters(t *testing.T) {
	h, mock, db := newListMock(t)
	defer db.Close()

	// Status counts: pipeline only. WithArgs asserts exactly one arg, which is
	// what proves status/item_type were not appended.
	mock.ExpectQuery("SELECT wi.status, count").
		WithArgs("build").
		WillReturnRows(statusCountRows())

	// Type counts: likewise pipeline only.
	mock.ExpectQuery("SELECT wi.item_type, count").
		WithArgs("build").
		WillReturnRows(typeCountRows())

	// Total and page: both filters applied.
	mock.ExpectQuery("SELECT count").
		WithArgs("build", "voice_tells", "needs_human_review").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(25))

	mock.ExpectQuery("SELECT wi.id, wi.site_id").
		WithArgs("build", "voice_tells", "needs_human_review", 200, 0).
		WillReturnRows(emptyItemRows())

	body := doList(t, h, "/work-items?pipeline=build&status=needs_human_review&item_type=voice_tells")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("query expectations: %v", err)
	}

	counts, ok := body["status_counts"].(map[string]interface{})
	if !ok {
		t.Fatalf("status_counts missing: %v", body)
	}
	// The whole point: the count survives a status filter that would otherwise
	// have scoped it to itself, and reports the table's 208 rather than the
	// page's 0.
	if got := counts["needs_human_review"]; got != float64(208) {
		t.Errorf("needs_human_review count = %v, want 208", got)
	}
	types, ok := body["type_counts"].(map[string]interface{})
	if !ok || len(types) != 2 {
		t.Errorf("type_counts should list every type in scope, got %v", body["type_counts"])
	}
}

// A truncated response must say so, so no consumer can read a page as a queue.
func TestListWorkItems_ReportsTruncationAndTotal(t *testing.T) {
	h, mock, db := newListMock(t)
	defer db.Close()

	mock.ExpectQuery("SELECT wi.status, count").WillReturnRows(statusCountRows())
	mock.ExpectQuery("SELECT wi.item_type, count").WillReturnRows(typeCountRows())
	mock.ExpectQuery("SELECT count").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(687))

	// One row returned against a total of 687.
	rows := emptyItemRows().AddRow(
		"11111111-1111-1111-1111-111111111111",
		"22222222-2222-2222-2222-222222222222",
		"example.com", "voice_tells", "needs_human_review",
		"medium", "a summary", []byte(`{}`),
		nil, 0, 3, nil, nil, nil,
	)
	mock.ExpectQuery("SELECT wi.id, wi.site_id").WillReturnRows(rows)

	body := doList(t, h, "/work-items?pipeline=build")

	if body["total"] != float64(687) {
		t.Errorf("total = %v, want 687", body["total"])
	}
	if body["truncated"] != true {
		t.Errorf("truncated = %v, want true (1 of 687 returned)", body["truncated"])
	}
	if body["count"] != float64(1) {
		t.Errorf("count = %v, want 1 (the window size)", body["count"])
	}
}

// pipeline=all must drop the predicate entirely — the dashboard hardcoded
// pipeline=build, which made the content-pipeline items unreachable by any route.
func TestListWorkItems_PipelineAllDropsThePredicate(t *testing.T) {
	h, mock, db := newListMock(t)
	defer db.Close()

	// No args at all: the pipeline predicate must be absent, not set to "all".
	mock.ExpectQuery("SELECT wi.status, count").
		WithArgs().
		WillReturnRows(statusCountRows())
	mock.ExpectQuery("SELECT wi.item_type, count").
		WithArgs().
		WillReturnRows(typeCountRows())
	mock.ExpectQuery("SELECT count").
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(302))
	mock.ExpectQuery("SELECT wi.id, wi.site_id").
		WithArgs(200, 0).
		WillReturnRows(emptyItemRows())

	doList(t, h, "/work-items?pipeline=all")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("query expectations: %v", err)
	}
}

// The limit is caller-settable but bounded; an absent or junk value falls back to
// the default rather than to zero, which would return an empty page that looks
// exactly like an empty queue.
func TestParseBoundedQueryInt(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  int
	}{
		{"absent falls back to default", "/x", defaultWorkItemPageSize},
		{"junk falls back to default", "/x?limit=abc", defaultWorkItemPageSize},
		{"empty falls back to default", "/x?limit=", defaultWorkItemPageSize},
		{"honoured within bounds", "/x?limit=25", 25},
		{"clamped to ceiling", "/x?limit=99999", maxWorkItemPageSize},
		{"clamped to floor, never zero", "/x?limit=0", 1},
		{"negative clamped to floor", "/x?limit=-5", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, tc.query, nil)
			got := parseBoundedQueryInt(c, "limit", defaultWorkItemPageSize, 1, maxWorkItemPageSize)
			if got != tc.want {
				t.Errorf("limit = %d, want %d", got, tc.want)
			}
		})
	}
}
