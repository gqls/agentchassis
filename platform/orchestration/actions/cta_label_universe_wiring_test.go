// FILE: platform/orchestration/actions/cta_label_universe_wiring_test.go
//
// bugs_open/308 Phase B — THE WIRING, not the helpers.
//
// Every other test in this lane exercises a pure function with a candidate list
// handed to it, so all of them stay green if the two writers go on building
// that list the OLD way. That is the exact shape Phase A was bitten by (its
// seed calls were deletable with the whole tree green until they were moved
// inside the writers), and it is what this file exists to stop: the widening is
// a change to WHICH LIST the writers are handed, so the assertion has to be
// made where the list is built.
//
// The first test fails if a future edit re-points either writer at a hub/tool-
// only supply — which is precisely what `candidatesFromHubs` did before it was
// deleted, and precisely what "tidying" this back would look like.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// ctaUniverseRows is one site's worth of pages as CTALabelUniverseSQL returns
// them: a tool, a hub, and the contact page that the old supply filtered out.
func ctaUniverseRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "title", "nav_label", "url", "page_type", "eligible_as_cta_target"}).
		AddRow("1", "tool-breakeven", "Break-Even Volume Calculator", "", "/tools/breakeven.html", "tool", true).
		AddRow("2", "services", "Services", "", "/services.html", "section-index", true).
		AddRow("3", "contact", "Contact our supply team", "", "/contact.html", "content", true)
}

// TestRerenderCTAStateOffersTheContactPage pins the REPAIR path's supply — the
// half bug 308 is named for. The check files a cta_links_stale rerender naming
// /contact.html; if this state's candidate list cannot contain one, the repair
// runs, reports success and changes nothing.
func TestRerenderCTAStateOffersTheContactPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The two positional loaders and the validity set still run; only their
	// shape matters here, so they return nothing.
	empty := sqlmock.NewRows([]string{"name", "title", "url", "nav_order", "eligible_as_cta_target"})
	mock.ExpectQuery("page_type = 'section-index'").WillReturnRows(empty)
	mock.ExpectQuery(`page_type IN \('tool', 'game'\)`).WillReturnRows(
		sqlmock.NewRows([]string{"name", "title", "url", "nav_order", "eligible_as_cta_target"}))
	// THE ASSERTION: the universe query is issued at all. Matched on the
	// build-state predicate, which is the part of it no other query on pages
	// carries — a re-point at the hub loaders fails here, not on a value check.
	mock.ExpectQuery(`build_status, ''\) = 'planned'`).WillReturnRows(ctaUniverseRows())
	mock.ExpectQuery("SELECT url FROM pages").WillReturnRows(
		sqlmock.NewRows([]string{"url"}).AddRow("/contact.html"))

	params := ActionParams{DB: db, Logger: zap.NewNop()}
	state := loadRerenderCTAState(context.Background(), params, uuid.New(), "home", "/index.html", zap.NewNop())

	match, ok, ambiguous := datahelpers.BestLabelMatch("Contact our supply team", state.candidates)
	if !ok || ambiguous || match.URL != "/contact.html" {
		t.Errorf("the repair path cannot reach the contact page: match=%+v ok=%v ambiguous=%v "+
			"(candidates=%d) — this is bug 308 itself", match, ok, ambiguous, len(state.candidates))
	}
}

// TestCTALabelUniverseExcludesAPageNoWriterCouldLinkTo pins the second axis of
// the same defect, found 2026-08-23: a page that is planned and never deployed
// fails the writers' validPages gate, so naming one is a repair no writer can
// perform. 43 of 764 live pages are in that state and 10 live findings named
// one. The candidate universe therefore carries the build-state predicate, and
// this test fails if it is dropped.
func TestCTALabelUniverseExcludesAPageNoWriterCouldLinkTo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// sqlmock matches the query as a regexp: a loader that dropped the
	// predicate would not match this expectation, and the call would error.
	mock.ExpectQuery(`NOT \(p.deployed_at IS NULL AND COALESCE\(p.build_status, ''\) = 'planned'\)`).
		WillReturnRows(ctaUniverseRows())

	got, err := datahelpers.LoadCTALabelUniverse(context.Background(), db, uuid.New())
	if err != nil {
		t.Fatalf("LoadCTALabelUniverse: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 candidates, got %d", len(got))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the universe query was not issued as written: %v", err)
	}
}

// TestResolveInternalLinksBuildPathOffersTheContactPage pins the BUILD path's
// supply. It exists because the sibling test above pins only the repair path,
// and the two writers build their candidate list in separate functions — the
// state bugs_open/248 was created by (the recompute was fixed and the rebuild
// was not, so an authored /contact.html died on the next full regeneration).
//
// Asserted through the action's own OUTPUT rather than by inspecting a list:
// the section's resolved_data must carry the contact url, which is only
// reachable if the label match was handed a universe containing it.
func TestResolveInternalLinksBuildPathOffersTheContactPage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("page_type = 'section-index'").WillReturnRows(
		sqlmock.NewRows([]string{"name", "title", "url", "nav_order", "eligible_as_cta_target"}))
	mock.ExpectQuery(`page_type IN \('tool', 'game'\)`).WillReturnRows(
		sqlmock.NewRows([]string{"name", "title", "url", "nav_order", "eligible_as_cta_target"}).
			AddRow("tool-breakeven", "Break-Even Volume Calculator", "/tools/breakeven.html", 10, true))
	mock.ExpectQuery("SELECT url FROM pages").WillReturnRows(
		sqlmock.NewRows([]string{"url"}).AddRow("/contact.html").AddRow("/tools/breakeven.html"))
	mock.ExpectQuery(`build_status, ''\) = 'planned'`).WillReturnRows(ctaUniverseRows())
	// the page's currently-published labels
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(
		sqlmock.NewRows([]string{"slot_name", "content_data"}).
			AddRow("hero", []byte(`{"cta_label":"Contact our supply team"}`)))

	section := map[string]interface{}{
		"name": "hero",
		"component": map[string]interface{}{
			"function":     "hero",
			"input_schema": `{"fields":{"cta_url":{"type":"string"},"cta_label":{"type":"string"}}}`,
		},
	}
	params := ActionParams{
		DB: db, Logger: zap.NewNop(),
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "resolve_internal_links"},
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id":  siteID.String(),
			"sections": "sections_ready",
		}},
		CollectedData: map[string]interface{}{
			"site_id":        siteID.String(),
			"page_name":      "home",
			"page_type":      "content",
			"sections_ready": []interface{}{section},
		},
	}

	out, err := ResolveInternalLinksAction(context.Background(), params)
	if err != nil {
		t.Fatalf("ResolveInternalLinksAction: %v", err)
	}
	res, _ := out.(map[string]interface{})
	ready, _ := res["sections_ready"].([]interface{})
	if len(ready) != 1 {
		t.Fatalf("expected one section back, got %v", res)
	}
	got, _ := ready[0].(map[string]interface{})["resolved_data"].(map[string]interface{})
	if got["cta_url"] != "/contact.html" {
		t.Errorf("build path could not reach the contact page: resolved_data=%v — "+
			"a label naming the contact page took the positional pick instead", got)
	}
	// …and the write is STAMPED, or the next pass reads it as authored and
	// freezes it (bugs_open/308 Phase A's whole purpose).
	if !datahelpers.CTAMintedCovers(got, "cta_url", "/contact.html") {
		t.Errorf("minted contact url carries no provenance record: %v", got)
	}
}
