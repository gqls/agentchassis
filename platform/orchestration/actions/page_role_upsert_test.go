// FILE: platform/orchestration/actions/page_role_upsert_test.go
//
// bugs_open/175 — four `pages` upserts named `page_type` in the INSERT and not in
// the `DO UPDATE SET`, so a name collision silently turned a CREATE into a PARTIAL
// update: this arm's content under the existing row's role.
//
// THE TESTS COME IN PAIRS ON PURPOSE, and this file inherits that discipline from
// bugs_closed/081's test file, which had to be corrected once for exactly the
// mistake the pairs prevent. A guard verified only on its firing branch is
// satisfied by DELETING the guard — a helper that refused everything would pass a
// refusal-only test. So every refusal case has a control that must still take the
// ordinary path, and the controls assert the write actually happens.
//
// `mock.ExpectationsWereMet()` is NOT "no database call happened": it reports
// expectations REGISTERED AND NOT CONSUMED, and never sees an EXTRA call
// (LANDMINES.md). What makes the "nothing was mutated" assertions real here is
// that UpsertPageForRole checks and PROPAGATES every statement's error — an
// unexpected statement inside the transaction errors, the error reaches the
// caller, and the t.Fatalf fires. That property is load-bearing; if you add a
// statement to the helper, check its error.
//
// The strongest assertions in this file read the SQL the helper actually built,
// captured through the query matcher, because the defect under test is a COLUMN
// LIST — "did page_type appear in the SET clause?" is the whole bug, and no
// expectation on statement ORDER can see it.
package actions

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// newCapturingMock returns a mock that records every statement it is asked to
// match, so a test can assert on the generated SQL itself — which is where this
// bug lives, in a column list no ordering expectation can see.
func newCapturingMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func() []string) {
	t.Helper()
	var seen []string
	matcher := sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		seen = append(seen, actualSQL)
		return sqlmock.QueryMatcherRegexp.Match(expectedSQL, actualSQL)
	})
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock, func() []string { return seen }
}

// toolPageWrite is the request the tool deployer makes, reused across the tests
// so a change in the contract shows up once rather than five times.
func toolPageWrite(name string) PageRoleUpsert {
	return PageRoleUpsert{
		SiteID:   uuid.New(),
		Domain:   "robot-hands.com",
		Name:     name,
		PageType: "tool",
		Source:   "tool-deployer",
		Columns: []PageColumn{
			Col("url", "/tools/"+name+".html"),
			Col("title", "Gripper Selection"),
			Col("nav_order", 7),
			Col("in_header", true),
			Col("in_footer", false),
			Col("meta_description", "Pick a gripper."),
			JSONCol("sections", `["hero-tool","tool-cta"]`),
			Col("build_status", "planned"),
			Col("status", "active"),
		},
		Refresh: []string{"url", "title", "sections"},
	}
}

// TestUpsertPageForRole_CleanCreate pins the ordinary path: the name is free, a
// row comes back, and the result says CREATED — the distinction the raw upsert
// could not express, because `RETURNING id` yielded an id either way.
func TestUpsertPageForRole_CleanCreate(t *testing.T) {
	db, mock, captured := newCapturingMock(t)

	newID := uuid.New()
	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(newID))

	res, err := UpsertPageForRole(context.Background(), db, toolPageWrite("tool-gripper-selection"), zap.NewNop())
	if err != nil {
		t.Fatalf("UpsertPageForRole: %v", err)
	}
	if res.Outcome != PageRoleCreated {
		t.Errorf("outcome = %q, want %q", res.Outcome, PageRoleCreated)
	}
	if res.PageID != newID {
		t.Errorf("page id = %s, want %s", res.PageID, newID)
	}

	// The statement must be DO NOTHING. A DO UPDATE here is the bug returning:
	// the conflict has to reach Go, or the database silently decides it.
	insert := captured()[0]
	if !strings.Contains(insert, "DO NOTHING") {
		t.Errorf("INSERT is not ON CONFLICT DO NOTHING — the collision never reaches the branch logic:\n%s", insert)
	}
	if strings.Contains(insert, "DO UPDATE") {
		t.Errorf("INSERT still carries a DO UPDATE — that is the shape 175 is about:\n%s", insert)
	}
	if !strings.Contains(insert, "page_type") {
		t.Errorf("INSERT does not write page_type at all:\n%s", insert)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpsertPageForRole_SameRoleRefreshesDeclaredColumnsOnly is the first
// control. The page already holds this arm's role, so refreshing is coherent and
// must still happen — this is the test that fails if the guard is widened to
// "never touch an existing page".
//
// It also pins the narrower half of the contract: the refresh writes the caller's
// DECLARED subset and nothing else. The companion-guide arms declare `title`
// alone so a written guide does not lose its sections to a tool redeploy, and a
// helper that "helpfully" wrote every column would destroy exactly that.
func TestUpsertPageForRole_SameRoleRefreshesDeclaredColumnsOnly(t *testing.T) {
	db, mock, captured := newCapturingMock(t)

	existing := uuid.New()
	mock.ExpectQuery("INSERT INTO pages").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status", "status", "ever"}).
			AddRow(existing, "tool", "deployed", "active", true))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	res, err := UpsertPageForRole(context.Background(), db, toolPageWrite("tool-gripper-selection"), zap.NewNop())
	if err != nil {
		t.Fatalf("UpsertPageForRole: %v", err)
	}
	if res.Outcome != PageRoleRefreshed {
		t.Fatalf("outcome = %q, want %q — a same-role refresh is not the defect and must still work", res.Outcome, PageRoleRefreshed)
	}

	update := lastMatching(captured(), "UPDATE pages")
	if update == "" {
		t.Fatal("no UPDATE pages was issued — the refresh did not happen")
	}
	// The refresh must NOT re-type the page. Only the adopt branch may.
	if strings.Contains(update, "page_type") {
		t.Errorf("the same-role refresh writes page_type; only ADOPT may re-type a row:\n%s", update)
	}
	for _, want := range []string{"url =", "title =", "sections ="} {
		if !strings.Contains(update, want) {
			t.Errorf("refresh omits %q:\n%s", want, update)
		}
	}
	// Declared-subset discipline: nav_order was written on CREATE but is not in
	// Refresh, so it must not appear here.
	if strings.Contains(update, "nav_order") {
		t.Errorf("refresh writes nav_order, which the caller did not declare:\n%s", update)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpsertPageForRole_LivePageOfAnotherRoleIsRefused is the firing branch, and
// the live surface it describes is real: robot-hands.com carries a DEPLOYED page
// named `gripper-selection-guide` typed `content` (measured 2026-08-02). Under
// the old upsert a tool deploy would have overwritten it and left it typed
// `content`. Nothing may be mutated, and a human decision must be filed.
func TestUpsertPageForRole_LivePageOfAnotherRoleIsRefused(t *testing.T) {
	db, mock, captured := newCapturingMock(t)

	existing := uuid.New()
	mock.ExpectQuery("INSERT INTO pages").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status", "status", "ever"}).
			AddRow(existing, "content", "deployed", "active", true))
	// insertWorkItem's two-strike probe, then the filing itself. No UPDATE pages
	// may appear anywhere between them.
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	res, err := UpsertPageForRole(context.Background(), db, toolPageWrite("gripper-selection-guide"), zap.NewNop())
	if err != nil {
		t.Fatalf("UpsertPageForRole: %v", err)
	}
	if !res.Refused() {
		t.Fatalf("outcome = %q, want %q", res.Outcome, PageRoleRefused)
	}
	if res.PageID != existing {
		t.Errorf("page id = %s, want the existing row %s", res.PageID, existing)
	}
	if res.ExistingType != "content" {
		t.Errorf("existing type = %q, want content — the caller cannot explain the refusal without it", res.ExistingType)
	}
	if !res.ItemFiled {
		t.Error("no work item filed — a refusal that files nothing is a silent drop")
	}
	if !strings.Contains(res.Reason, "gripper-selection-guide") {
		t.Errorf("reason does not name the page: %q", res.Reason)
	}

	// The mutation assertion, made on the SQL rather than on the mock's
	// bookkeeping: no statement touched `pages` beyond the initial INSERT and the
	// locking SELECT.
	for _, q := range captured() {
		if strings.Contains(q, "UPDATE pages") {
			t.Errorf("the refusal path mutated pages:\n%s", q)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpsertPageForRole_HasShippedComesFromTheSharedPredicate stops this helper
// re-deriving "has this page been served" in Go.
//
// > **It replaced a test that pinned the WRONG rule, and the replacement is the
// > interesting part.** The first version asserted that `build_status =
// > 'needs_rebuild'` counts as live ON ITS OWN. Measured live 2026-08-02: 11 rows
// > are `needs_rebuild` with NO `deployed_at`, no `last_built_at` and mostly zero
// > components — never built, never served, three of them tool pages created that
// > same day. The old rule REFUSED all eleven.
// > `datahelpers.NeverDeployedPagePredicate` already existed, with two other
// > consumers and a test forbidding exactly the clause that was added here,
// > because singling out that status had produced a 34-page false-positive class
// > for the nav lane.
//
// So the SQL must carry the shared predicate rather than a Go restatement of it.
// Inline a hand-rolled version again and this goes red.
func TestUpsertPageForRole_HasShippedComesFromTheSharedPredicate(t *testing.T) {
	db, mock, captured := newCapturingMock(t)

	mock.ExpectQuery("INSERT INTO pages").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status", "status", "shipped"}).
			AddRow(uuid.New(), "content", "needs_rebuild", "active", true))
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	res, err := UpsertPageForRole(context.Background(), db, toolPageWrite("selection-guide"), zap.NewNop())
	if err != nil {
		t.Fatalf("UpsertPageForRole: %v", err)
	}
	if !res.Refused() {
		t.Fatalf("outcome = %q — this row HAS shipped, so it must be refused", res.Outcome)
	}

	sel := lastMatching(captured(), "FOR UPDATE")
	if !strings.Contains(sel, datahelpers.NeverDeployedPagePredicate) {
		t.Errorf("the locking SELECT does not use datahelpers.NeverDeployedPagePredicate — a SECOND definition of \"has this page shipped\" has been introduced:\n%s", sel)
	}
	if strings.Contains(sel, "needs_rebuild") {
		t.Errorf("the SELECT singles out needs_rebuild, which is the 34-page false-positive class (datahelpers/links_deployment_test.go):\n%s", sel)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpsertPageForRole_NeedsRebuildButNeverBuiltIsAdopted is the control the
// deleted test would have FAILED, and it is the one with live rows behind it: 11
// fleet pages are `needs_rebuild` with no `deployed_at`. Nothing has ever served
// them, so a constant-role arm colliding with one must ADOPT — not refuse, not
// file a human decision, and not hard-error the tool deploy.
func TestUpsertPageForRole_NeedsRebuildButNeverBuiltIsAdopted(t *testing.T) {
	db, mock, captured := newCapturingMock(t)

	mock.ExpectQuery("INSERT INTO pages").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	// shipped=false is what the shared predicate returns for this row.
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status", "status", "shipped"}).
			AddRow(uuid.New(), "content", "needs_rebuild", "active", false))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	res, err := UpsertPageForRole(context.Background(), db, toolPageWrite("tool-price-cap-checker"), zap.NewNop())
	if err != nil {
		t.Fatalf("UpsertPageForRole: %v", err)
	}
	if res.Outcome != PageRoleAdopted {
		t.Fatalf("outcome = %q, want %q — this row has never been built, so refusing it files a human decision nobody needs", res.Outcome, PageRoleAdopted)
	}
	if update := lastMatching(captured(), "UPDATE pages"); !strings.Contains(update, "page_type") {
		t.Errorf("ADOPT did not write page_type:\n%s", update)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpsertPageForRole_DeployedAtAloneCountsAsShipped covers the row whose
// build_status says nothing useful — 'planned', or empty — but which carries a
// deployed_at. It has been served; the status string is not the authority. This
// is the half of the shared predicate that `deployed_at IS NULL` carries.
func TestUpsertPageForRole_DeployedAtAloneCountsAsShipped(t *testing.T) {
	db, mock, _ := newCapturingMock(t)

	mock.ExpectQuery("INSERT INTO pages").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status", "status", "shipped"}).
			AddRow(uuid.New(), "content", "planned", "active", true))
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	res, err := UpsertPageForRole(context.Background(), db, toolPageWrite("selection-guide"), zap.NewNop())
	if err != nil {
		t.Fatalf("UpsertPageForRole: %v", err)
	}
	if !res.Refused() {
		t.Fatalf("outcome = %q — deployed_at is set, so the page has been served", res.Outcome)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpsertPageForRole_UnshippedPageOfAnotherRoleIsAdopted is the second
// control, and it is the branch that actually repairs 175's defect rather than
// merely refusing to make it worse.
//
// The row has never been served (planned, deployed_at NULL). The arm's role is a
// constant it owns, so it takes the row over COMPLETELY — page_type included.
// Leaving the type wrong is the defect; adopting half the columns would mint a
// fresh hybrid, which is the same defect wearing different clothes.
func TestUpsertPageForRole_UnshippedPageOfAnotherRoleIsAdopted(t *testing.T) {
	db, mock, captured := newCapturingMock(t)

	existing := uuid.New()
	mock.ExpectQuery("INSERT INTO pages").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status", "status", "ever"}).
			AddRow(existing, "content", "planned", "active", false))
	mock.ExpectExec("UPDATE pages").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	res, err := UpsertPageForRole(context.Background(), db, toolPageWrite("tool-gripper-selection"), zap.NewNop())
	if err != nil {
		t.Fatalf("UpsertPageForRole: %v", err)
	}
	if res.Outcome != PageRoleAdopted {
		t.Fatalf("outcome = %q, want %q", res.Outcome, PageRoleAdopted)
	}

	update := lastMatching(captured(), "UPDATE pages")
	if !strings.Contains(update, "page_type") {
		t.Errorf("ADOPT did not write page_type — that omission IS bugs_open/175:\n%s", update)
	}
	// Whole takeover: the columns the refresh subset leaves out must be here.
	for _, want := range []string{"url =", "title =", "sections =", "nav_order =", "meta_description =", "status ="} {
		if !strings.Contains(update, want) {
			t.Errorf("ADOPT omits %q, leaving a half-claimed page:\n%s", want, update)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpsertPageForRole_RejectsAContractBreach guards the seam itself. A caller
// that passes its own page_type column, or an empty role, would put the class of
// bug back — one through the column list, the other by making every collision
// look like a role change and sending working pages to the refusal queue.
func TestUpsertPageForRole_RejectsAContractBreach(t *testing.T) {
	cases := []struct {
		name string
		req  PageRoleUpsert
		want string
	}{
		{
			name: "page_type passed as a column",
			req: PageRoleUpsert{SiteID: uuid.New(), Name: "x", PageType: "tool",
				Columns: []PageColumn{Col("page_type", "content")}},
			want: "owned by the helper",
		},
		{
			name: "no role",
			req:  PageRoleUpsert{SiteID: uuid.New(), Name: "x"},
			want: "page_type is required",
		},
		{
			name: "refresh names a column that is never written",
			req: PageRoleUpsert{SiteID: uuid.New(), Name: "x", PageType: "tool",
				Columns: []PageColumn{Col("title", "t")}, Refresh: []string{"sections"}},
			want: "not one of the written columns",
		},
		{
			name: "column name is not a bare identifier",
			req: PageRoleUpsert{SiteID: uuid.New(), Name: "x", PageType: "tool",
				Columns: []PageColumn{Col("title = 'x', page_type", "t")}},
			want: "not a bare column identifier",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, _ := newCapturingMock(t)

			_, err := UpsertPageForRole(context.Background(), db, tc.req, zap.NewNop())
			if err == nil {
				t.Fatal("accepted a request that breaks the contract")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not explain the breach (want it to mention %q)", err, tc.want)
			}
			// No statement may have been issued: the validation runs first.
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func lastMatching(queries []string, substr string) string {
	for i := len(queries) - 1; i >= 0; i-- {
		if strings.Contains(queries[i], substr) {
			return queries[i]
		}
	}
	return ""
}
