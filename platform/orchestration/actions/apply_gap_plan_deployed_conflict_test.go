// FILE: platform/orchestration/actions/apply_gap_plan_deployed_conflict_test.go
//
// bugs_open/081 — a DEPLOYED but mistyped page had no repair path, and the
// `new_page` arm made it worse rather than leaving it alone.
//
// The arm's INSERT carried `ON CONFLICT (site_id, name) DO UPDATE SET title,
// sections`. `page_type` was in the INSERT and NOT in the DO UPDATE, so on a
// name collision the row became a hybrid nobody asked for: the plan's content
// under the existing row's role. Two live consequences, both measured —
//   - the live page's title and sections were overwritten by a plan written for
//     a role it does not hold;
//   - the mistype survived, so the check that raised the gap fired again on the
//     next sweep. ai-agent-orchestration.com looped this from 2026-05-01.
//
// THE TESTS COME IN PAIRS ON PURPOSE. A guard verified only on its firing branch
// is satisfied by deleting the guard: refusing everything would pass a
// refusal-only test. So each refusal case has a control that must still take the
// old path, and the controls assert the `UPDATE pages` refresh actually happens.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Every test below drives applyNewPage down the ON CONFLICT branch by returning
// no rows from the INSERT — which is exactly what `DO NOTHING ... RETURNING id`
// does when the name is taken — then answers the follow-up read with the row the
// arm collided with.
//
// The growth-budget probes that run first are deliberately not expected, for the
// reason recorded in apply_gap_plan_new_page_test.go: CheckPageGrowthBudget
// swallows its own Scan errors, so an unmatched SELECT leaves the counts at zero
// and the budget allows the page.

// TestApplyNewPage_DeployedTypeConflictIsRefused is the firing branch: the plan
// wants `news-index`, a DEPLOYED page already holds that name as `content`.
// Nothing about the page may change, and the originating item must be BLOCKED
// rather than completed — 'complete' on an item whose defect is untouched is the
// false green this bug spent three months producing.
func TestApplyNewPage_DeployedTypeConflictIsRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	existingID := uuid.New()
	originalItem := uuid.New()

	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status"}).
			AddRow(existingID, "content", "deployed"))
	// The refusal files a review item and blocks the originator. It must NOT
	// file a needs_content_page item, and must NOT update `pages`.
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE site_work_items").
		WithArgs(originalItem, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := applyNewPage(context.Background(), db,
		newPagePlan("news", "news-index"), uuid.New(), "ai-agent-orchestration.com",
		&originalItem, zap.NewNop())
	if err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}

	m, ok := res.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T, want map", res)
	}
	if applied, _ := m["applied"].(bool); applied {
		t.Error("applied = true on a refused conflict — the caller cannot tell the plan did nothing")
	}
	if got := m["reason"]; got != "deployed_page_type_conflict" {
		t.Errorf("reason = %v, want deployed_page_type_conflict", got)
	}
	if created, _ := m["page_created"].(bool); created {
		t.Error("page_created = true, but the page already existed and was not touched")
	}

	// The load-bearing assertion: no UPDATE pages was issued. sqlmock fails an
	// unexpected call, so an added mutation shows up here.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestApplyNewPage_DeployedSameTypeStillRefreshes is the first control. The page
// is deployed, but it already holds the role the plan asked for — so the plan is
// for this page's actual job and refreshing its content is coherent. This is the
// case that fails if the guard is widened to "never touch a deployed page".
func TestApplyNewPage_DeployedSameTypeStillRefreshes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	existingID := uuid.New()

	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status"}).
			AddRow(existingID, "news-index", "deployed"))
	mock.ExpectExec("UPDATE pages").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := applyNewPage(context.Background(), db,
		newPagePlan("news", "news-index"), uuid.New(), "webdesign.co.uk", nil, zap.NewNop())
	if err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}

	m := res.(map[string]interface{})
	if applied, _ := m["applied"].(bool); !applied {
		t.Error("applied = false — a same-type refresh is not the defect and must still work")
	}
	if created, _ := m["page_created"].(bool); created {
		t.Error("page_created = true on a conflict — the row already existed")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestApplyNewPage_UndeployedTypeConflictStillRefreshes is the second control,
// and it is the one the fleet measurement chose. On 2026-07-31 every mistyped
// page in production was `deployed` — 0 planned, 0 needs_rebuild — so the guard
// is scoped to deployed rows and a never-shipped page keeps the old behaviour.
// If this test starts failing, the guard has been widened past its evidence.
func TestApplyNewPage_UndeployedTypeConflictStillRefreshes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	existingID := uuid.New()

	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT id, COALESCE").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_type", "build_status"}).
			AddRow(existingID, "content", "planned"))
	mock.ExpectExec("UPDATE pages").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := applyNewPage(context.Background(), db,
		newPagePlan("news", "news-index"), uuid.New(), "example.com", nil, zap.NewNop())
	if err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}

	m := res.(map[string]interface{})
	if applied, _ := m["applied"].(bool); !applied {
		t.Error("applied = false on an undeployed conflict — the guard has widened past the measurement")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestApplyNewPage_CleanCreateReportsCreated pins the ordinary path: no
// conflict, a row comes back, and the result says so. dartsonline.com took
// exactly this branch on 2026-07-29 and its missing_news_page item completed —
// which is the evidence that the defect lives in the conflict branch alone.
func TestApplyNewPage_CleanCreateReportsCreated(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := applyNewPage(context.Background(), db,
		newPagePlan("news", "news-index"), uuid.New(), "dartsonline.com", nil, zap.NewNop())
	if err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}

	m := res.(map[string]interface{})
	if created, _ := m["page_created"].(bool); !created {
		t.Error("page_created = false on a clean insert")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
