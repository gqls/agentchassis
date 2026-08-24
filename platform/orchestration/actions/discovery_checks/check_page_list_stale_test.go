// FILE: platform/orchestration/actions/discovery_checks/check_page_list_stale_test.go
//
// bugs_open/384, the sweep half. Pinned: a stored entry whose image differs
// from the fresh resolve files ONE page_rerender under the SHARED key (so it
// collapses onto the event emitter's item); matching images file nothing; a
// source that fails to resolve files nothing AND counts as unknown (never as
// current); and no Resolved entry is ever produced — the key is shared, so a
// retraction here would over-claim (see the file header).

package discovery_checks

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

const staleTestSchema = `{"fields":{"articles":{"type":"array","source":"query.blog_posts","required":true}}}`

func expectConsumers(mock sqlmock.Sqlmock, siteID, pageID uuid.UUID) {
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "url", "domain", "cc_name", "input_schema"}).
			AddRow(pageID, "index", "/index.html", "example.com", "content-listing", staleTestSchema))
}

func expectStored(mock sqlmock.Sqlmock, pageID uuid.UUID, contentData string) {
	mock.ExpectQuery("SELECT cc.name, COALESCE\\(pc.content_data::text").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"name", "content_data"}).AddRow("content-listing", contentData))
}

func freshRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"name", "title", "url", "meta_description", "nav_label", "card_key", "hero_key", "hero_purpose"})
}

func TestPageListStale_FilesOneReasonedItemUnderTheSharedKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID, pageID := uuid.New(), uuid.New()

	expectConsumers(mock, siteID, pageID)
	expectStored(mock, pageID, `{"heading":"x","articles":[{"url":"/a.html","image":""},{"url":"/b.html","image":"`+storage.DeployedWebPath("card_b", "card")+`"},{"url":"/gone.html","image":""}]}`)
	// The fresh resolve: /a now has a card, /b unchanged, /gone no longer listed.
	mock.ExpectQuery("FROM pages p").
		WithArgs(siteID, "blog-post", pageListStaleResolveLimit).
		WillReturnRows(freshRows().
			AddRow("a", "A", "/a.html", "", "A", "card_a", "", "").
			AddRow("b", "B", "/b.html", "", "B", "card_b", "", ""))

	res, err := (&PageListStaleCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Pipeline: "content",
		AgentType: "completeness-discovery-agent", BatchID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d: %+v", len(res.WorkItems), res.WorkItems)
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "page_rerender" || wi.HandlerAgent != "page-rerender" || wi.Status != "detected" {
		t.Errorf("routing = %s/%s/%s, want page_rerender/page-rerender/detected", wi.ItemType, wi.HandlerAgent, wi.Status)
	}
	if want := PageRerenderItemKey("index", siteID, "section_data_resolved"); wi.ItemKey != want {
		t.Errorf("key = %q, want the SHARED key %q so the sweep collapses onto the event emitter's item", wi.ItemKey, want)
	}
	if wi.PageID == nil || *wi.PageID != pageID {
		t.Errorf("page_id column not set (LANDMINES: a page_rerender item needs page_id in the spec AND the column)")
	}
	for _, needle := range []string{`"reason":"section_data_resolved"`, `"page_name":"index"`, `"page_id":"` + pageID.String() + `"`, `"url":"/a.html"`, `"current_image":"` + storage.DeployedWebPath("card_a", "card") + `"`} {
		if !strings.Contains(wi.SpecJSON, needle) {
			t.Errorf("spec %s is missing %s", wi.SpecJSON, needle)
		}
	}
	if strings.Contains(wi.SpecJSON, `"/b.html"`) || strings.Contains(wi.SpecJSON, `"/gone.html"`) {
		t.Errorf("only the MISMATCHED, still-listed entry belongs in the spec, got %s", wi.SpecJSON)
	}
	if len(res.Resolved) != 0 {
		t.Errorf("the sweep must never retract on the shared key, got %+v", res.Resolved)
	}
	// The stale counter must be asserted NON-ZERO somewhere, or a counter that
	// is never incremented is indistinguishable from one that is (352 lane).
	if s := sweepSummary(t, res); s["stale"] != 1 || s["current"] != 0 || s["unknown"] != 0 {
		t.Errorf("summary = %v, want stale 1 / current 0 / unknown 0", s)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestPageListStale_MatchingImagesFileNothingAndRetractNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID, pageID := uuid.New(), uuid.New()

	expectConsumers(mock, siteID, pageID)
	expectStored(mock, pageID, `{"articles":[{"url":"/a.html","image":"`+storage.DeployedWebPath("card_a", "card")+`"}]}`)
	mock.ExpectQuery("FROM pages p").
		WillReturnRows(freshRows().AddRow("a", "A", "/a.html", "", "A", "card_a", "", ""))

	res, err := (&PageListStaleCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, AgentType: "t", BatchID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("current page must file nothing and retract nothing, got %+v", res)
	}
	if s := sweepSummary(t, res); s["current"] != 1 || s["stale"] != 0 || s["unknown"] != 0 {
		t.Fatalf("matching images must count as CURRENT: %v", s)
	}
}

// Kills: treating a failed or empty resolve as "matches". The stored image is
// empty and the source cannot answer — that is UNKNOWN, and filing a re-render
// on it would be a remedy for a defect nobody observed.
func TestPageListStale_UnresolvableSourceIsUnknownNotCurrentAndNotStale(t *testing.T) {
	for _, mode := range []string{"error", "empty"} {
		t.Run(mode, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()
			siteID, pageID := uuid.New(), uuid.New()
			expectConsumers(mock, siteID, pageID)
			expectStored(mock, pageID, `{"articles":[{"url":"/a.html","image":""}]}`)
			q := mock.ExpectQuery("FROM pages p")
			if mode == "error" {
				q.WillReturnError(errors.New("boom"))
			} else {
				q.WillReturnRows(freshRows())
			}

			res, err := (&PageListStaleCheck{}).Run(DiscoveryCheckContext{
				Ctx: context.Background(), DB: db, SiteID: siteID, AgentType: "t", BatchID: uuid.New(), Logger: zap.NewNop(),
			})
			if err != nil {
				t.Fatalf("an unresolvable source must not fail the whole check: %v", err)
			}
			if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
				t.Fatalf("unknown must file nothing and retract nothing, got %+v", res)
			}
			s := sweepSummary(t, res)
			if s["unknown"] != 1 || s["current"] != 0 || s["stale"] != 0 {
				t.Fatalf("an unresolvable source must count the page as UNKNOWN (not current): %v", s)
			}
		})
	}
}

// sweepSummary finds the per-run summary finding — the only place the
// unknown/current split is observable from outside the check.
func sweepSummary(t *testing.T, res *CheckResult) map[string]interface{} {
	t.Helper()
	for _, f := range res.Findings {
		if f["summary"] == true {
			return f
		}
	}
	t.Fatal("no summary finding — unknown and current would be indistinguishable")
	return nil
}

func TestPageListStale_ConsumerLookupFailureIsAnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM page_components pc").WillReturnError(errors.New("boom"))
	if _, err := (&PageListStaleCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(), AgentType: "t", BatchID: uuid.New(), Logger: zap.NewNop(),
	}); err == nil {
		t.Fatal("a failed consumer lookup must surface as the check failing, not as a clean empty sweep")
	}
}

func TestPageListStaleIsRegisteredUnderItsConfigName(t *testing.T) {
	if got := Get("page_list_stale"); got == nil {
		t.Fatal("page_list_stale is not registered — the checks array in migration 603 would warn-and-skip it")
	}
}

// Round 2005a846, bug_historian: a component whose content_data is not a JSON
// object used to be skipped with a bare `continue`, which read as "nothing
// stored" — one layer down from the unknown/current collapse the summary
// finding exists to prevent. It must count the page UNKNOWN (and file nothing).
func TestPageListStale_UnparseableContentDataIsUnknownNotCurrent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	siteID, pageID := uuid.New(), uuid.New()
	expectConsumers(mock, siteID, pageID)
	// TWO components on the page: one unreadable, one current. Without the
	// unknown mark the current one would carry the page to "current" — that is
	// the discriminating shape (a single unreadable component is already caught
	// by compared == 0, which is why the first cut of this test could not kill
	// the mutation: a guard in series).
	mock.ExpectQuery("SELECT cc.name, COALESCE\\(pc.content_data::text").
		WithArgs(pageID).
		WillReturnRows(sqlmock.NewRows([]string{"name", "content_data"}).
			AddRow("content-listing", `not json at all`).
			AddRow("content-listing", `{"articles":[{"url":"/a.html","image":"`+storage.DeployedWebPath("card_a", "card")+`"}]}`))
	// The source resolves fine — the page is unknown because of ITS data, not the source's.
	mock.ExpectQuery("FROM pages p").
		WillReturnRows(freshRows().AddRow("a", "A", "/a.html", "", "A", "card_a", "", ""))

	res, err := (&PageListStaleCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, AgentType: "t", BatchID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("nothing can be filed on data that could not be read, got %+v", res.WorkItems)
	}
	if s := sweepSummary(t, res); s["unknown"] != 1 || s["current"] != 0 {
		t.Fatalf("unparseable content_data must count the page UNKNOWN, not current: %v", s)
	}
}
