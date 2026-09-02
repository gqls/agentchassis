// FILE: platform/orchestration/actions/listing_item_sources_test.go
//
// bugs_open/444 — the listing-page item-source gate. The healthy control
// matters as much as the broken cases: vetcomparison's shape (bare
// directory-listing + exporter config) must be KEPT, a recommendation-only
// news page must be KEPT (the trigger seeds sources from the spec), and a
// section-index whose children are in the plan but unbuilt (tool pages,
// bugs_open/311's ordering) must be KEPT. The glossary shape (content page,
// generic-text-block) is pinned as OUT of the gate's sight — that half of 444
// stays with the planner prompt + the copy lane's title-promise design, and
// this test is the honest statement of that boundary.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// --- resolver-level tests -------------------------------------------------

func TestResolveListing_NewsIndex_NoSourcesNoRecommendation_Held(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("FROM content_sources").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "coalesce"}).AddRow(false, false))

	page := planPageView{Name: "news-index", Role: "news-index", URL: "/news/index.html",
		Sections: []string{"hero", "news-listing"}}
	res := ResolveListingItemSource(context.Background(), db, siteID, page, []planPageView{page}, zap.NewNop())

	if !res.ListingFamily {
		t.Fatal("news-index must be listing-family")
	}
	if res.Producible {
		t.Fatalf("expected held: no sources, no recommendation (evidence: %s)", res.Evidence)
	}
	if res.ProducerNeeded != "news_source_enablement" {
		t.Fatalf("producer slug = %q, want news_source_enablement", res.ProducerNeeded)
	}
}

func TestResolveListing_NewsIndex_RecommendationOnly_Kept(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// No source rows yet, but the classification spec recommends the feed —
	// the 6-hourly trigger seeds sources from the spec (NEWS-001), so the
	// page is producible.
	mock.ExpectQuery("FROM content_sources").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "coalesce"}).AddRow(false, true))

	page := planPageView{Name: "news-index", Role: "news-index", Sections: []string{"hero", "news-listing"}}
	res := ResolveListingItemSource(context.Background(), db, uuid.New(), page, []planPageView{page}, zap.NewNop())
	if !res.Producible {
		t.Fatalf("recommendation alone must keep the page (evidence: %s)", res.Evidence)
	}
}

func TestResolveListing_BareDirectoryListing_NoConfig_Held(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The exporter-config lookup returns no rows — exactly the state in which
	// resolveBusinessDirectory ERRORS at render time (bugs_open/206/444).
	mock.ExpectQuery("directory-json-exporter").
		WillReturnRows(sqlmock.NewRows([]string{"vertical", "business_type_ilike"}))

	page := planPageView{Name: "channels-directory-index", Role: "entity-directory",
		URL: "/channels-directory/index.html", Sections: []string{"hero", "directory-listing"}}
	res := ResolveListingItemSource(context.Background(), db, uuid.New(), page, []planPageView{page}, zap.NewNop())
	if res.Producible {
		t.Fatalf("expected held: no exporter config (evidence: %s)", res.Evidence)
	}
	if res.ProducerNeeded != "business_directory_config" {
		t.Fatalf("producer slug = %q, want business_directory_config", res.ProducerNeeded)
	}
}

func TestResolveListing_BareDirectoryListing_ConfigPresent_Kept(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// The vetcomparison control: config row present, page must be KEPT.
	mock.ExpectQuery("directory-json-exporter").
		WillReturnRows(sqlmock.NewRows([]string{"vertical", "business_type_ilike"}).
			AddRow("veterinary", ""))

	page := planPageView{Name: "directory-index", Role: "entity-directory",
		Sections: []string{"hero", "directory-listing"}}
	res := ResolveListingItemSource(context.Background(), db, uuid.New(), page, []planPageView{page}, zap.NewNop())
	if !res.Producible {
		t.Fatalf("vetcomparison control must be kept (evidence: %s)", res.Evidence)
	}
}

func TestResolveListing_PerKindDirectory_OptInDecides(t *testing.T) {
	for _, tc := range []struct {
		name        string
		recommended bool
		wantKept    bool
	}{
		{"opted-in", true, true},
		{"not-opted-in", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			mock.ExpectQuery("content_features").
				WillReturnRows(sqlmock.NewRows([]string{"recommended"}).AddRow(tc.recommended))

			page := planPageView{Name: "mortgage-lenders", Role: "mortgage-lenders",
				Sections: []string{"hero", "mortgage-lender-directory-listing"}}
			res := ResolveListingItemSource(context.Background(), db, uuid.New(), page, []planPageView{page}, zap.NewNop())
			if res.Producible != tc.wantKept {
				t.Fatalf("producible = %v, want %v (evidence: %s)", res.Producible, tc.wantKept, res.Evidence)
			}
			if !tc.wantKept && res.ProducerNeeded != "directory_kind_opt_in:mortgage_lender_directory" {
				t.Fatalf("producer slug = %q", res.ProducerNeeded)
			}
		})
	}
}

func TestResolveListing_SectionIndex_PlanChildren(t *testing.T) {
	// Pure plan-internal resolution: no DB expectations at all.
	index := planPageView{Name: "guides-index", Role: "section-index", URL: "/guides/index.html",
		Sections: []string{"hero", "generic-text-block"}}
	toolGuide := planPageView{Name: "guide-a", Role: "guide", URL: "/guides/a/index.html"}

	// With a child in the plan — kept, even though the child is unbuilt
	// (tool-deployer ordering, bugs_open/311: the pages are IN the plan).
	res := ResolveListingItemSource(context.Background(), nil, uuid.New(), index,
		[]planPageView{index, toolGuide}, zap.NewNop())
	if !res.Producible {
		t.Fatalf("section-index with a planned child must be kept (evidence: %s)", res.Evidence)
	}

	// Without children — held (designblog /inspiration/, /the-design-feed/).
	empty := planPageView{Name: "inspiration-index", Role: "section-index", URL: "/inspiration/index.html",
		Sections: []string{"hero", "generic-text-block", "call-to-action"}}
	res = ResolveListingItemSource(context.Background(), nil, uuid.New(), empty,
		[]planPageView{index, toolGuide, empty}, zap.NewNop())
	if res.Producible {
		t.Fatal("section-index with no children anywhere must be held")
	}
	if res.ProducerNeeded != "section_children:inspiration-index" {
		t.Fatalf("producer slug = %q", res.ProducerNeeded)
	}
}

func TestResolveListing_ContentGlossary_OutOfSight(t *testing.T) {
	// The honest boundary: a glossary typed `content` with generic sections is
	// NOT listing-family to this gate. If this test ever fails because the
	// family test widened, update the bug-444 scope statement too.
	page := planPageView{Name: "glossary", Role: "content", URL: "/glossary.html",
		Sections: []string{"hero", "generic-text-block"}}
	res := ResolveListingItemSource(context.Background(), nil, uuid.New(), page, []planPageView{page}, zap.NewNop())
	if res.ListingFamily {
		t.Fatal("content+generic-text-block must be out of the gate's sight (scope pin)")
	}
}

func TestResolveListing_UnknownListingComponent_FailsOpen(t *testing.T) {
	// category-listing has no resolver — the page must be KEPT (fail-open) so
	// an unmodelled shape can never silently shrink a site.
	page := planPageView{Name: "shop-index", Role: "landing",
		Sections: []string{"hero", "category-listing"}}
	res := ResolveListingItemSource(context.Background(), nil, uuid.New(), page, []planPageView{page}, zap.NewNop())
	if !res.ListingFamily || !res.Producible {
		t.Fatalf("unknown listing component must fail open (family=%v producible=%v)", res.ListingFamily, res.Producible)
	}
}

// --- gate-level tests -----------------------------------------------------

func gateParams(dbConn interface{}, siteID uuid.UUID) ActionParams {
	return ActionParams{
		CollectedData: map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
		},
		Logger: zap.NewNop(),
	}
}

func TestEnforceListingItemSources_DropsFilesGapAndKeeps(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	// news-index resolution: no sources, no recommendation → drop.
	mock.ExpectQuery("FROM content_sources").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "coalesce"}).AddRow(false, false))
	// The capability_gap receipt, via the SHARED writer (insertWorkItem):
	// tx open → anti-churn probe → INSERT → commit.
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	// The durable findings row (agent_error_log) is best-effort; allow it.
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := gateParams(db, siteID)
	params.DB = db

	pages := []interface{}{
		map[string]interface{}{"name": "index", "page_type": "landing",
			"sections": []interface{}{"hero"}},
		map[string]interface{}{"name": "news-index", "page_type": "news-index",
			"url": "/news/index.html", "sections": []interface{}{"hero", "news-listing"}},
	}
	kept := enforceListingItemSources(context.Background(), params, pages, nil)

	if len(kept) != 1 {
		t.Fatalf("kept %d pages, want 1 (the landing page)", len(kept))
	}
	if name := kept[0].(map[string]interface{})["name"]; name != "index" {
		t.Fatalf("kept page = %v, want index", name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected the news check + gap insert to run: %v", err)
	}
}

func TestEnforceListingItemSources_RealisedPageKeptWithReceipt(t *testing.T) {
	// A listing page that already EXISTS on the site is never dropped (the
	// bugs_open/001 preserve guard owns built pages) — but the capability_gap
	// receipt is still filed so the enablement debt is on record.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID := uuid.New()

	mock.ExpectQuery("FROM content_sources").
		WillReturnRows(sqlmock.NewRows([]string{"exists", "coalesce"}).AddRow(false, false))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(0, 999.0))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	params := gateParams(db, siteID)
	params.DB = db

	newsPage := map[string]interface{}{"name": "news-index", "page_type": "news-index",
		"url": "/news/index.html", "sections": []interface{}{"hero", "news-listing"}}
	kept := enforceListingItemSources(context.Background(), params,
		[]interface{}{newsPage},
		[]interface{}{map[string]interface{}{"name": "news-index", "url": "/news/index.html", "page_type": "news-index"}})

	if len(kept) != 1 {
		t.Fatalf("realised page must be kept, got %d pages", len(kept))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the receipt must still be filed for a kept realised page: %v", err)
	}
}

func TestEnforceListingItemSources_NilDBFailsOpen(t *testing.T) {
	params := gateParams(nil, uuid.New())
	pages := []interface{}{
		map[string]interface{}{"name": "news-index", "page_type": "news-index",
			"sections": []interface{}{"hero", "news-listing"}},
	}
	kept := enforceListingItemSources(context.Background(), params, pages, nil)
	if len(kept) != 1 {
		t.Fatal("nil DB must fail open (page kept)")
	}
}
