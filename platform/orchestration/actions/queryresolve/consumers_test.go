// FILE: platform/orchestration/actions/queryresolve/consumers_test.go
//
// bugs_open/384. Three things are pinned: the SQL carries the owned-page
// exclusion (asserted on the statement, not on the result — a result-only test
// passes with the clause deleted); the Go filter keeps exactly the fields whose
// source reads page images and drops the rest; and both schema dialects are
// read through datahelpers.SchemaContentFields.

package queryresolve

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func consumerRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "url", "domain", "cc_name", "input_schema"})
}

func TestPageListConsumerPages_KeepsOnlyPageImageSources(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	home, news, legacy := uuid.New(), uuid.New(), uuid.New()

	rows := consumerRows().
		// home page: an article listing AND a tool strip → one page, two fields
		AddRow(home, "index", "/index.html", "example.com", "content-listing",
			`{"fields":{"heading":{"type":"string","source":"llm"},"articles":{"type":"array","source":"query.blog_posts","required":true}}}`).
		AddRow(home, "index", "/index.html", "example.com", "tool-list",
			`{"fields":{"items":{"type":"array","source":"query.pages_where_type:tool"}}}`).
		// news index: a query source that does NOT read page images → dropped
		AddRow(news, "news-index", "/news/index.html", "example.com", "news-listing",
			`{"fields":{"items":{"type":"array","source":"query.news_archive"}}}`).
		// legacy JSON-Schema dialect, projected by SchemaContentFields → kept
		AddRow(legacy, "guides-index", "/guides/index.html", "example.com", "guide-list_pre_037",
			`{"properties":{"items":{"type":"array","source":"query.pages_where_type:guide"}},"required":["items"]}`).
		// a schema that is not JSON at all → logged and skipped, not fatal
		AddRow(uuid.New(), "broken", "/broken.html", "example.com", "odd", `not json`)

	// The owned-page exclusion AND the renders-image predicate are asserted on
	// the STATEMENT. Deleting either clause from pageListConsumerSQL fails
	// here; a result-shaped assertion would not notice.
	mock.ExpectQuery(regexp.QuoteMeta("COALESCE(p.rebuild_policy, 'generic') <> 'owned'") + `[\s\S]*` + regexp.QuoteMeta(`cc.html_template ~ '\.image\y'`)).
		WithArgs(siteID).
		WillReturnRows(rows)

	got, err := PageListConsumerPages(context.Background(), db, siteID, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 consumer pages (index, guides-index), got %d: %+v", len(got), got)
	}
	if got[0].ID != home || got[0].Name != "index" || got[0].Domain != "example.com" {
		t.Errorf("first page = %+v, want index/%s", got[0], home)
	}
	if len(got[0].Fields) != 2 ||
		got[0].Fields[0] != (ConsumerField{Component: "content-listing", Field: "articles", Source: "query.blog_posts"}) ||
		got[0].Fields[1] != (ConsumerField{Component: "tool-list", Field: "items", Source: "query.pages_where_type:tool"}) {
		t.Errorf("index fields = %+v, want content-listing/articles then tool-list/items", got[0].Fields)
	}
	if s := got[0].Sources(); len(s) != 2 || s[0] != "query.blog_posts" || s[1] != "query.pages_where_type:tool" {
		t.Errorf("index Sources() = %v", s)
	}
	if got[1].ID != legacy || len(got[1].Fields) != 1 || got[1].Fields[0].Source != "query.pages_where_type:guide" {
		t.Errorf("legacy-dialect page = %+v, want guides-index with one pages_where_type:guide field", got[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet: %v", err)
	}
}

func TestPageListConsumerPages_NoConsumersIsEmptyNotError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM page_components pc").WillReturnRows(consumerRows())

	got, err := PageListConsumerPages(context.Background(), db, uuid.New(), zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want no consumers, got %+v", got)
	}
}

func TestPageListConsumerPages_QueryFailureIsAnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM page_components pc").WillReturnError(errors.New("boom"))

	if _, err := PageListConsumerPages(context.Background(), db, uuid.New(), zap.NewNop()); err == nil {
		t.Fatal("a failed lookup must be an error — an empty answer here would silently tell no consumer")
	}
	if _, err := PageListConsumerPages(context.Background(), db, uuid.Nil, zap.NewNop()); err == nil {
		t.Fatal("empty site_id must be refused, not queried")
	}
}
