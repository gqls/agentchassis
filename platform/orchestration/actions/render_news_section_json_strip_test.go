// bugs_open/332: the JSON producer is the surface the bug file never named, and
// the one carrying the live defect. /data/latest-news.json and
// /data/news-archive.json are public, and a published news-listing script
// fetches the archive and assigns it straight into innerHTML — so on five sites
// the JSON is what a JS-enabled visitor actually reads, and it was completely
// unstripped.
//
// Measured at the artefact 2026-09-03, before this fix: boxingonline's archive
// JSON served 7 ATX headings, 4 complete markdown links, 5 truncated links, a
// list marker, an image and a bold marker — while the server-rendered HTML of
// the SAME query served ZERO headings. That pair is the whole diagnosis, and
// these fixtures are it in miniature.

package actions

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func newsRowCols() []string {
	return []string{"title", "summary", "url", "source_published_at", "source_name", "topics_json"}
}

func TestLoadNewsItemsStripsLiteralMarkdown(t *testing.T) {
	t.Setenv("DISABLE_NEWS_MARKDOWN_STRIP", "")
	logger := zap.NewNop()
	ctx := context.Background()
	siteID := uuid.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pub := time.Date(2026, 9, 1, 8, 0, 0, 0, time.UTC)
	// QueryNewsItems runs the item select, then a best-effort content_sources
	// lookup for the tool cap; the second is allowed to fail.
	mock.ExpectQuery("FROM content_feed_items").
		WillReturnRows(sqlmock.NewRows(newsRowCols()).
			// the heading case — the pattern that proved the JSON path was raw
			AddRow("# The housing market's fragile place", "## Executive takeaways\nSome prose.",
				"https://example.com/1", pub, "Axios", "[]").
			// the live severed-link case
			AddRow("Fight results", "- Tennis (W)\n- [NLL (Lacrosse)](https://www.espn.com/boxin...",
				"https://example.com/2", pub, "ESPN", "[]"))
	mock.ExpectQuery("FROM content_sources").WillReturnRows(sqlmock.NewRows([]string{"query"}))

	items, err := loadNewsItems(ctx, db, siteID, 720, 20, false, logger)
	if err != nil {
		t.Fatalf("loadNewsItems: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	for i, it := range items {
		for _, marker := range []string{"](http", "**", "!["} {
			if strings.Contains(it.Title, marker) || strings.Contains(it.Summary, marker) {
				t.Errorf("item %d: marker %q survived — title %q summary %q", i, marker, it.Title, it.Summary)
			}
		}
		// The heading marker is line-anchored, so check it as a prefix per line.
		for _, line := range strings.Split(it.Title+"\n"+it.Summary, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "# ") || strings.HasPrefix(strings.TrimSpace(line), "## ") {
				t.Errorf("item %d: an ATX heading survived: %q", i, line)
			}
		}
	}

	// Visible text must survive — "no markers" alone would be satisfied by a
	// projection that emptied every field.
	if !strings.Contains(items[0].Title, "housing market") {
		t.Errorf("title lost its visible text: %q", items[0].Title)
	}
	if !strings.Contains(items[1].Summary, "NLL (Lacrosse)") {
		t.Errorf("summary lost the link text it should keep: %q", items[1].Summary)
	}
	// And the truncation marker survives the severed link.
	if !strings.HasSuffix(items[1].Summary, "...") {
		t.Errorf("truncation marker lost: %q", items[1].Summary)
	}

	// The JSON must NOT be HTML-escaped: it is data, and its consumer escapes at
	// render. Escaping here would double-escape for any correct consumer.
	if strings.Contains(items[0].Title+items[0].Summary, "&amp;") ||
		strings.Contains(items[0].Title+items[0].Summary, "&#39;") {
		t.Errorf("the JSON projection HTML-escaped its output: %+v", items[0])
	}
}
