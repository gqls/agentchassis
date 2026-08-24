// FILE: platform/orchestration/actions/page_list_reresolve_test.go
//
// bugs_open/384. Pins, in order: the item each consumer gets (type, handler,
// reason, key — the key is the dedup contract with the page_list_stale sweep,
// asserted through the exported helper rather than restated); the branches
// that decide NOT to file (no consumers, lookup failure, provenance not
// recorded, derive raised) — each of which is silent by construction, so each
// is driven and its disposition asserted; and the two call sites, pinned by a
// comment-stripped source ratchet because derive_card_asset cannot be driven
// end-to-end without a real S3 client (it type-asserts *storage.S3Client).
//
// The negative assertions ("nothing is written") are carried by sqlmock's
// refusal of unexpected calls PLUS the disposition string: the helpers swallow
// errors by design, so an unexpected query would surface as lookup_failed, not
// as the skip disposition the test demands.

package actions

import (
	"context"
	"database/sql/driver"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"go.uber.org/zap"
)

// specArgWith matches a string argument that carries every needle.
type specArgWith []string

func (a specArgWith) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	for _, n := range a {
		if !strings.Contains(s, n) {
			return false
		}
	}
	return true
}

func pageListConsumerRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "url", "domain", "cc_name", "input_schema"})
}

const blogPostsSchema = `{"fields":{"articles":{"type":"array","source":"query.blog_posts","required":true}}}`

// expectPageListReresolve scripts one consumer page and the page_rerender
// INSERT it must produce. Shared with page_section_satisfiability_test.go's
// flag_page_image_rebuild tests, which now run this leg inside their tx.
func expectPageListReresolve(mock sqlmock.Sqlmock, source string, siteID uuid.UUID) uuid.UUID {
	pageID := uuid.New()
	mock.ExpectQuery("FROM page_components pc").
		WillReturnRows(pageListConsumerRows().AddRow(pageID, "index", "/index.html", "example.com", "content-listing", blogPostsSchema))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			siteID,
			source, // $2 source = created_by
			"low",
			sqlmock.AnyArg(), // summary
			pageID,
			specArgWith{`"reason":"section_data_resolved"`, `"page_name":"index"`},
			discovery_checks.PageRerenderItemKey("index", siteID, "section_data_resolved"),
			sqlmock.AnyArg(), // batch
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	return pageID
}

func TestRequestPageListReresolve_FilesOneReasonedItemPerConsumer(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	siteID, batchID := uuid.New(), uuid.New()
	home, guides := uuid.New(), uuid.New()
	mock.ExpectQuery("FROM page_components pc").
		WithArgs(siteID).
		WillReturnRows(pageListConsumerRows().
			AddRow(home, "index", "/index.html", "example.com", "content-listing", blogPostsSchema).
			AddRow(guides, "guides-index", "/guides/index.html", "example.com", "content-listing", blogPostsSchema))

	// The canonical page_rerender shape: (site, source, severity, summary,
	// page_id, spec, key, batch). item_type and handler are literals inside
	// insertPageRerenderItem's statement, which the regex below names.
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "derive_card_asset", "low", specArgWith{"index", "card_landed:barrel-shapes"}, home,
			specArgWith{`"reason":"section_data_resolved"`, `"page_id":"` + home.String() + `"`, `"cause":"card_landed:barrel-shapes"`, `"consumes":["query.blog_posts"]`, `"domain":"example.com"`},
			discovery_checks.PageRerenderItemKey("index", siteID, "section_data_resolved"), batchID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Second consumer already has an open reasoned item: dedup-suppressed.
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "derive_card_asset", "low", sqlmock.AnyArg(), guides, sqlmock.AnyArg(),
			discovery_checks.PageRerenderItemKey("guides-index", siteID, "section_data_resolved"), batchID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	got := requestPageListReresolve(context.Background(), db, siteID, "derive_card_asset", "card_landed:barrel-shapes", batchID, zap.NewNop())
	if got.Disposition != "queued" || got.Consumers != 2 || got.Queued != 1 || got.Deduped != 1 || got.Failed != 0 {
		t.Fatalf("got %+v, want queued/2 consumers/1 queued/1 deduped", got)
	}
	if len(got.Pages) != 1 || got.Pages[0] != "index" {
		t.Errorf("Pages = %v, want [index] (only the page that actually got a new row)", got.Pages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestRequestPageListReresolve_NoConsumersFilesNothing(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(pageListConsumerRows())

	got := requestPageListReresolve(context.Background(), db, uuid.New(), "derive_card_asset", "card_landed:x", uuid.New(), zap.NewNop())
	if got.Disposition != "no_consumers" || got.Queued != 0 {
		t.Fatalf("got %+v, want no_consumers", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestRequestPageListReresolve_LookupFailureIsNonFatal(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.ExpectQuery("FROM page_components pc").WillReturnError(errors.New("boom"))

	got := requestPageListReresolve(context.Background(), db, uuid.New(), "derive_card_asset", "card_landed:x", uuid.New(), zap.NewNop())
	if got.Disposition != "lookup_failed" {
		t.Fatalf("got %+v, want lookup_failed — and no panic, no error to the caller", got)
	}
}

func TestRequestPageListReresolve_OneInsertFailureDoesNotStopTheRest(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	siteID := uuid.New()
	a, b := uuid.New(), uuid.New()
	mock.ExpectQuery("FROM page_components pc").
		WillReturnRows(pageListConsumerRows().
			AddRow(a, "a", "/a.html", "example.com", "tool-list", `{"fields":{"items":{"type":"array","source":"query.pages_where_type:tool"}}}`).
			AddRow(b, "b", "/b.html", "example.com", "tool-list", `{"fields":{"items":{"type":"array","source":"query.pages_where_type:tool"}}}`))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "image-build-handler", "low", sqlmock.AnyArg(), a, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("boom"))
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(siteID, "image-build-handler", "low", sqlmock.AnyArg(), b, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	got := requestPageListReresolve(context.Background(), db, siteID, "image-build-handler", "page_image_landed:x", uuid.New(), zap.NewNop())
	if got.Disposition != "partial" || got.Queued != 1 || got.Failed != 1 {
		t.Fatalf("got %+v, want partial/1 queued/1 failed", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

// Kills: dropping the provenanceRecorded guard. A lock-suppressed upsert changed
// no row, so the join answers exactly as before — telling the listings would be
// a re-render for nothing. No query is scripted; the disposition discriminates
// "skipped on purpose" from "tried and failed".
func TestReresolvePageListsAfterCard_SkipsWhenProvenanceNotRecorded(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	got := reresolvePageListsAfterCard(context.Background(), db, uuid.New(), "barrel-shapes", false, zap.NewNop())
	if got["page_list_reresolve"] != "skipped_provenance_not_recorded" {
		t.Fatalf("got %v, want skipped_provenance_not_recorded", got["page_list_reresolve"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestReresolvePageListsAfterCard_TellsTheListingsWhenRecorded(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	siteID := uuid.New()
	expectPageListReresolve(mock, "derive_card_asset", siteID)

	got := reresolvePageListsAfterCard(context.Background(), db, siteID, "barrel-shapes", true, zap.NewNop())
	if got["page_list_reresolve"] != "queued" || got["page_list_reresolve_queued"] != 1 {
		t.Fatalf("got %v, want queued/1", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

// Kills: inverting or dropping the cardEmit gate. A raised derive means the
// card is coming and will supersede the hero in the projection; the listing is
// told by derive_card_asset when it lands, not twice.
func TestReresolvePageListsAfterPageImage_DefersWhenADeriveWasRaised(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	if got := reresolvePageListsAfterPageImage(context.Background(), tx, uuid.New(), "board-setup", "raised", uuid.New(), zap.NewNop()); got != "deferred_to_card_derive" {
		t.Fatalf("got %q, want deferred_to_card_derive", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestReresolvePageListsAfterPageImage_TellsTheListingsWhenNoDeriveIsComing(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()
	siteID := uuid.New()
	mock.ExpectBegin()
	expectPageListReresolve(mock, "image-build-handler", siteID)
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	for _, emit := range []string{"skipped_no_content_hero", "skipped_card_exists"} {
		// Only the first call has a scripted INSERT; the second finds no
		// expectation and reports lookup_failed — which is fine here, the
		// point is that neither of these dispositions defers.
		got := reresolvePageListsAfterPageImage(context.Background(), tx, siteID, "board-setup", emit, uuid.New(), zap.NewNop())
		if got == "deferred_to_card_derive" {
			t.Fatalf("cardEmit=%q must not defer — no card is coming for this page", emit)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

// The key is ONE spelling, and it cannot collide with the needs_page
// producers' `page_rerender:<page>` keys that share the site-wide
// (site_id, item_key) namespace of idx_swi_dedup (item_type is not a column).
func TestPageRerenderItemKeyHasOneSpellingAndItsOwnNamespace(t *testing.T) {
	siteID := uuid.New()
	for _, reason := range []string{"section_data_resolved", "image_landed", ""} {
		got, want := pageRerenderItemKey("index", siteID, reason), discovery_checks.PageRerenderItemKey("index", siteID, reason)
		if got != want {
			t.Fatalf("pageRerenderItemKey(%q) = %q, want the exported %q", reason, got, want)
		}
		if strings.HasPrefix(got, "page_rerender:") {
			t.Fatalf("key %q shares the needs_page producers' `page_rerender:` shape — the dedup index is per (site_id, item_key) across ALL item types", got)
		}
	}
	if got := pageRerenderItemKey("index", siteID, ""); !strings.HasSuffix(got, "_assemble") {
		t.Fatalf("an empty reason must key as assemble, got %q", got)
	}
	if got := pageRerenderItemKey("index", siteID, "section_data_resolved"); got != "page_rerender_index_"+siteID.String()+"_section_data_resolved" {
		t.Fatalf("key shape changed: %q", got)
	}
}

// The two call sites, pinned. derive_card_asset cannot be driven to its upsert
// under sqlmock (it type-asserts *storage.S3Client), so the wiring is a
// comment-stripped source ratchet anchored on the line each call must follow:
// the card caller after the provenance decision, the image caller after the
// card-derive emit whose disposition it reads. Comments are stripped first so a
// prose mention of the helper cannot satisfy it (the
// a-source-scanning-test-makes-comments-load-bearing trap).
func TestPageListReresolveCallSitesAreWired(t *testing.T) {
	for _, site := range []struct{ file, anchor, call string }{
		{"derive_card_asset_action.go", "provenanceRecorded := true", "reresolvePageListsAfterCard("},
		{"flag_page_image_rebuild_action.go", "cardEmit := emitContentCardDerive(", "reresolvePageListsAfterPageImage("},
	} {
		raw, err := os.ReadFile(site.file)
		if err != nil {
			t.Fatalf("read %s: %v", site.file, err)
		}
		src := stripGoComments(string(raw))
		at := strings.Index(src, site.anchor)
		if at < 0 {
			t.Fatalf("%s: anchor %q not found — update the anchor here rather than deleting the row", site.file, site.anchor)
		}
		if !strings.Contains(src[at:], site.call) {
			t.Fatalf("%s: %s is not called after %q — the listings would never be told (bugs_open/384)", site.file, site.call, site.anchor)
		}
	}
}
