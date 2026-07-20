// FILE: platform/orchestration/actions/discovery_checks/check_news_feed_test.go
//
// bugs_open/015: a mistyped page_type orphans a page from every gate that keys
// on it. relojistas.com's Spanish news listing was emitted by the planner as
// page_type='section-index' instead of 'news-index', so render_news_section
// produced no archive for it, page-build-handler had nothing to build, and the
// header nav link to it was a live 404 on a production site.
//
// MissingNewsPageCheck could only see "no news-index page exists" and would
// have told the planner to create one — shipping an English /news.html beside
// the existing Spanish /noticias rather than fixing the mistyped page.
package discovery_checks

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestFindStrandedNavPagesIsStructuralNotNameBased(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	// The regex pins the deployed-exclusion into the test: without it the
	// predicate returned six working, deployed tool/blog pages on
	// ai-agent-orchestration.com whose sections are legitimately empty.
	mock.ExpectQuery(`build_status.*<>.*'deployed'`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "page_type", "build_status"}).
			// The real relojistas row. Note the name is Spanish — a vocabulary
			// match on "news" would never have found it.
			AddRow("11111111-1111-1111-1111-111111111111", "noticias-index", "section-index", "planned").
			AddRow("22222222-2222-2222-2222-222222222222", "glosario-index", "section-index", "planned"))

	got, err := findStrandedNavPages(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d candidates, want 2", len(got))
	}
	if got[0].Name != "noticias-index" || got[0].PageType != "section-index" {
		t.Errorf("first candidate = %+v, want the mistyped Spanish listing", got[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The cap must be reported, not silently applied — a truncated list that reads
// as complete is how a caller concludes "that is all of them".
func TestFindStrandedNavPagesCapsAndReports(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "name", "page_type", "build_status"})
	for i := 0; i < strandedNavPageCap+5; i++ {
		rows.AddRow(uuid.New().String(), "page", "section-index", "planned")
	}
	// The regex pins the deployed-exclusion into the test: without it the
	// predicate returned six working, deployed tool/blog pages on
	// ai-agent-orchestration.com whose sections are legitimately empty.
	mock.ExpectQuery(`build_status.*<>.*'deployed'`).WithArgs(siteID).WillReturnRows(rows)

	got, err := findStrandedNavPages(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != strandedNavPageCap {
		t.Errorf("got %d, want the cap %d", len(got), strandedNavPageCap)
	}
}

func TestDescribeCandidatesNamesTypeAndBuildStatus(t *testing.T) {
	got := describeCandidates([]strandedNavPage{
		{Name: "noticias-index", PageType: "section-index", BuildStatus: "planned"},
		{Name: "orphan", PageType: "", BuildStatus: ""},
	})

	// The current type is the thing that is WRONG and the build status is what
	// shows the page never built — the planner needs both to judge.
	for _, want := range []string{"noticias-index", "page_type=section-index", "build_status=planned"} {
		if !strings.Contains(got, want) {
			t.Errorf("description %q must contain %q", got, want)
		}
	}
	// Empty values must not render as bare "page_type=", which reads like a
	// formatting bug rather than missing data.
	if !strings.Contains(got, "(no page_type)") || !strings.Contains(got, "(no build_status)") {
		t.Errorf("empty fields must be named explicitly, got %q", got)
	}
}
