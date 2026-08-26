// FILE: platform/orchestration/actions/rerender_page_sections_scan_completeness_test.go
//
// loadStoredSections must not return a THINNED slice. bugs_open/410, instance 3.
//
// THE DEFECT. Its rows.Scan error branch logs a Warn and continues, so a scan
// failure returned fewer sections — or none — with no error. Because
// save_page_sections then replaces the page's rows WHOLESALE (`DELETE FROM
// page_components WHERE page_id = $1`), a section missing from that slice is not
// merely unrendered: it is DELETED. The page ships with a hole, under a fresh
// deploy stamp, with the work item reported complete. That is why this reader is
// strict where scanBlogArticles (rebuild_blog_listing_action.go) is graded — a
// listing that loses one post degrades a nav surface; this loses the page.
//
// THE REPRODUCTION THIS TEST EXISTS FOR, and it is real rather than
// constructed. Commit bd811fa93 ("035 P1: loadStoredSections reads the
// composition pair") added two columns to that SELECT. SEVEN tests went red,
// every one of them reporting a variant of "expected exactly one section, got
// 0" — and not one said "scan mismatch". A genuine, correct change to a genuine
// query presented as AN EMPTY PAGE rather than an error, and the covering tests
// encoded the symptom while losing the cause. The seven, by NAME because they
// have moved twice (supplied and verified by the news_editorial lane, who owns
// them):
//
//	TestRerenderPageSections_SuccessEntryCarriesTheStoredSlotName
//	TestRerenderPageSections_FailsWhenComponentUnresolvedByNameOrID
//	TestRerenderPageSections_ResolvesToolByComponentIDWithoutEscalating
//	TestRerenderPageSections_ComponentIDWinsOverNameWhenBothResolve
//	TestRerenderPageSections_InvalidTemplateByID_IsFatalAndNamed
//	TestRerenderPageSections_EmptyTemplateCarriesWithoutFailing
//	TestRerenderPageSections_StructuralCarryMakesANotReadySectionRerender
//
// ⚠ AND THE SEAM IS INVISIBLE TO THEM, NOT MERELY UNNAMED BY THEM. Those seven
// assert section COUNT and CONTENT, never rows-in-equals-rows-out — so they
// would have PASSED on a column change that was wrong in a way that still
// scanned. Worse, they call mock.ExpectationsWereMet(), so sqlmock's own
// completeness assertion ran and PASSED: the query was issued exactly as
// expected, and only the scan silently produced nothing. Even the mocking
// framework's completeness check cannot see this failure. That is what a check
// which cannot come out the other way looks like, and it is why the guard had to
// be a count rather than a better error message.
//
// MUTATION CHECK for whoever changes loadStoredSections: delete the
// datahelpers.ScanShortfall call and restore `return out, rows.Err()`. Then
// TestLoadStoredSections_RefusesAPartialScan and
// TestLoadStoredSections_TotalScanLossIsAnErrorNotAnEmptyPage must BOTH fail. If
// either still passes, the guard is not what is producing the refusal and the
// coverage hole is back.
//
// HOW THE ROWS ARE POISONED, stated because it is the one fragile part: a NULL
// in the `position` column. storedSection.position is a plain `int`, and
// database/sql cannot scan NULL into it. Every OTHER projected column is
// COALESCEd in the SQL or scans into a NULL-safe []byte, which is exactly why
// this guard cannot fire on live data today (measured 2026-08-26: page_components
// has `position integer NOT NULL` and `id uuid NOT NULL`; content_data is NULL on
// 54 live rows and scans to nil harmlessly). If a future change makes position
// nullable or COALESCEs it, THIS TEST GOES GREEN FOR THE WRONG REASON — pick
// another non-nullable destination to poison rather than deleting the test.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// storedSectionCols mirrors loadStoredSections' projection, in order.
func storedSectionCols() []string {
	return []string{"id", "parent_instance_id", "component_id", "slot_name",
		"content_data", "rendered_html", "position", "component_version_id"}
}

func TestLoadStoredSections_RefusesAPartialScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	// Two rows offered by the cursor. The second carries NULL position, which
	// cannot scan into int — so the loop keeps one and drops one.
	rows := sqlmock.NewRows(storedSectionCols()).
		AddRow(uuid.New().String(), "", uuid.New().String(), "hero",
			[]byte(`{"a":1}`), "<section>hero</section>", 0, "").
		AddRow(uuid.New().String(), "", uuid.New().String(), "features",
			[]byte(`{"b":2}`), "<section>features</section>", nil, "")

	mock.ExpectQuery("FROM page_components").WithArgs(pageID).WillReturnRows(rows)

	out, err := loadStoredSections(context.Background(), db, pageID, zap.NewNop())
	if err == nil {
		t.Fatalf("loadStoredSections silently dropped a row and returned %d section(s) with no error — "+
			"save_page_sections would then DELETE the missing section's row and the page would ship "+
			"with a hole, looking freshly built (bugs_open/410 instance 3)", len(out))
	}

	// The error must say what was lost, not merely that something went wrong.
	msg := err.Error()
	for _, want := range []string{"kept 1 of 2", "rerender_page_sections"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal must name the shortfall so it is diagnosable from the log alone; "+
				"want %q in: %s", want, msg)
		}
	}
}

func TestLoadStoredSections_TotalScanLossIsAnErrorNotAnEmptyPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	// Both rows poisoned: the projection/destination divergence case. This is
	// what bd811fa93 produced, and before the guard it surfaced as "expected
	// exactly one section, got 0" — a page with no sections, and no error.
	rows := sqlmock.NewRows(storedSectionCols()).
		AddRow(uuid.New().String(), "", uuid.New().String(), "hero",
			[]byte(`{}`), "", nil, "").
		AddRow(uuid.New().String(), "", uuid.New().String(), "features",
			[]byte(`{}`), "", nil, "")

	mock.ExpectQuery("FROM page_components").WithArgs(pageID).WillReturnRows(rows)

	out, err := loadStoredSections(context.Background(), db, pageID, zap.NewNop())
	if err == nil {
		t.Fatalf("every offered row failed to scan and loadStoredSections returned (%d sections, nil error) — "+
			"an empty page reported as a successful load. This is the exact shape of the bd811fa93 "+
			"reproduction: the SELECT list and the Scan destinations had diverged, and the seven "+
			"covering tests reported the symptom (\"got 0\") while the cause said nothing", len(out))
	}
	if len(out) != 0 {
		t.Errorf("expected no sections back alongside the refusal, got %d", len(out))
	}
}

func TestLoadStoredSections_HealthyRowsAllKept(t *testing.T) {
	// The guard must NOT fire on correct input. This is not a formality: the
	// bug file pins "a guard that fires constantly on correct input gets
	// loosened within a week, and a loosened guard is a dead one" as the reason
	// the intuitive version of this check (comparing against a second COUNT
	// query over page_components) must not be used — it would fire on every page
	// carrying a build_status='removed' tombstone, a large healthy population.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	rows := sqlmock.NewRows(storedSectionCols()).
		AddRow(uuid.New().String(), "", uuid.New().String(), "hero",
			[]byte(`{"a":1}`), "<section>hero</section>", 0, "").
		AddRow(uuid.New().String(), "", uuid.New().String(), "features",
			[]byte(`{"b":2}`), "<section>features</section>", 1, "")

	mock.ExpectQuery("FROM page_components").WithArgs(pageID).WillReturnRows(rows)

	out, err := loadStoredSections(context.Background(), db, pageID, zap.NewNop())
	if err != nil {
		t.Fatalf("the guard fired on two healthy rows — a guard that refuses correct input gets "+
			"loosened and then removed: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected both healthy sections, got %d", len(out))
	}
}

func TestLoadStoredSections_EmptyPageIsNotALoss(t *testing.T) {
	// A page with no components yields nothing and loses nothing. The guard must
	// stay silent, or every genuinely empty page becomes a failed rerender.
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery("FROM page_components").WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows(storedSectionCols()))

	out, err := loadStoredSections(context.Background(), db, pageID, zap.NewNop())
	if err != nil {
		t.Fatalf("an empty result set is an empty page, not a scan loss: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no sections, got %d", len(out))
	}
}
