// FILE: platform/orchestration/actions/rerender_page_sections_removed_test.go
//
// A deliberately removed section must not come back through the light rerender
// path. loadStoredSections re-renders from stored content_data and
// save_page_sections then replaces the page's rows wholesale, so a
// build_status='removed' row that reaches the render is resurrected AS
// 'deployed' — with every step reporting success.
//
// This is a live incident, not a hypothetical: on 2026-08-10 the tool-list
// section returned to idea.uk/index (removed at the owner's instruction on
// 08-05) via an unrelated section_data_resolved rerender. The current plan
// (site_plan_sections for the is_current plan) and pages.sections both still
// excluded it; page_components was the only store that disagreed, because
// removal marks the status and empties rendered_html but leaves content_data —
// exactly what this path renders from.
//
// WHY THE QUERY TEXT IS THE ASSERTION HERE: the filtering is done by the
// DATABASE, so the predicate in the SQL *is* the behaviour. sqlmock matches
// ExpectQuery's pattern against the statement the action actually sends, so
// deleting the predicate from the query makes this test fail — which is the
// property a guard test needs. Asserting on returned rows instead would prove
// nothing: the mock returns whatever rows the test hands it, no matter what the
// WHERE clause says.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// TestRerenderPageSections_ExcludesRemovedSections pins the predicate that stops
// the resurrection.
func TestRerenderPageSections_ExcludesRemovedSections(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	mock.ExpectQuery("SELECT p.id, s.domain").
		WithArgs(siteID, pageID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain", "url", "name"}).
			AddRow(pageID, "example.com", "/index.html", "index"))

	// The stored-section read must carry the removed-row exclusion in SOME form —
	// any of !=, <> or IS DISTINCT FROM satisfies this test; which one is the
	// separate test below. An empty row set short-circuits the rest of the
	// action, which is all this test needs.
	//
	// The pattern must name the PREDICATE. An earlier version of this test
	// expected only "FROM page_components", which matches the query whether or
	// not the exclusion is there — it passed with the predicate deleted, so it
	// asserted nothing. Caught by mutating the source rather than by reading it.
	mock.ExpectQuery(`build_status\s*(!=|<>|IS DISTINCT FROM)\s*'removed'`).
		WillReturnRows(sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}))

	p := resolveParams(db, map[string]interface{}{
		"input_data": map[string]interface{}{
			"page_id": pageID.String(),
			"site_id": siteID.String(),
			"domain":  "example.com",
			"spec":    map[string]interface{}{"reason": "section_data_resolved"},
		},
	}, map[string]interface{}{
		"target_site_id": "input_data.site_id",
		"reason":         "input_data.spec.reason",
	})

	if _, err = RerenderPageSectionsAction(context.Background(), p); err != nil {
		t.Fatalf("action failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the stored-section read has no removed-row exclusion — a deliberately "+
			"removed section will be re-rendered and re-saved as deployed: %v", err)
	}
}

// TestRerenderPageSections_RemovedFilterIsNullSafe fails if the predicate is
// written as `!= 'removed'` (or `<> 'removed'`) rather than IS DISTINCT FROM.
// Separate from the test above so the failure message says WHICH property broke:
// the exclusion existing and the exclusion being NULL-safe are different bugs
// with different consequences.
func TestRerenderPageSections_RemovedFilterIsNullSafe(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	mock.ExpectQuery("SELECT p.id, s.domain").
		WithArgs(siteID, pageID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "domain", "url", "name"}).
			AddRow(pageID, "example.com", "/index.html", "index"))
	mock.ExpectQuery(`build_status IS DISTINCT FROM 'removed'`).
		WillReturnRows(sqlmock.NewRows([]string{"component_id", "slot_name", "content_data", "rendered_html", "position"}))

	p := resolveParams(db, map[string]interface{}{
		"input_data": map[string]interface{}{
			"page_id": pageID.String(),
			"site_id": siteID.String(),
			"domain":  "example.com",
			"spec":    map[string]interface{}{"reason": "section_data_resolved"},
		},
	}, map[string]interface{}{
		"target_site_id": "input_data.site_id",
		"reason":         "input_data.spec.reason",
	})

	if _, err = RerenderPageSectionsAction(context.Background(), p); err != nil {
		t.Fatalf("action failed: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the removed-row exclusion is not the NULL-safe form: %v", err)
	}
}
