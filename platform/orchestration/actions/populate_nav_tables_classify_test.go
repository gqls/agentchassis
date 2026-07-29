package actions

import (
	"testing"

	"go.uber.org/zap"
)

// A typed section index under its own section prefix is the section's PARENT
// and must survive the child-prefix skip. news-index was missing from
// isSectionIndexType, so a news page at the canonical /news/index.html could
// never enter primary nav, while the identical page at the non-canonical
// /news.html (the shape bugs_open/080 exists to eliminate) navigated fine.
// Observed live on webdesign.co.uk, nav-updater run 2026-07-29 07:52:09Z:
// "classifyPagesForNav: skipping child page" name=news url=/news/index.html.
// bugs_open/141.
func TestClassifyPagesForNavNewsIndexAtCanonicalURL(t *testing.T) {
	logger := zap.NewNop()

	pages := []pageNavInfo{
		{Name: "tools-index", Title: "Tools", URL: "/tools/index.html", PageType: "section-index", InHeader: true, NavOrder: 10},
		{Name: "news", Title: "News", URL: "/news/index.html", PageType: "news-index", InHeader: true, NavOrder: 40},
		// Children that must STAY skipped: the exception is keyed on page_type,
		// not on living under a prefix.
		{Name: "tool-touch-target", Title: "Touch Target", URL: "/tools/touch-target/index.html", PageType: "tool", InHeader: true, NavOrder: 100},
		{Name: "some-post", Title: "A Post", URL: "/news/some-post.html", PageType: "blog-post", InHeader: true, NavOrder: 100},
	}

	primary, legal, utility := classifyPagesForNav(pages, logger)

	names := map[string]bool{}
	for _, p := range primary {
		names[p.Name] = true
	}
	if !names["news"] {
		t.Fatalf("news-index page at /news/index.html missing from primary nav; primary=%v", names)
	}
	if !names["tools-index"] {
		t.Fatalf("section-index page regressed out of primary nav; primary=%v", names)
	}
	if names["tool-touch-target"] {
		t.Fatalf("tool child page leaked into primary nav; primary=%v", names)
	}
	if names["some-post"] {
		t.Fatalf("blog-post under /news/ leaked into primary nav; primary=%v", names)
	}
	if len(legal) != 0 {
		t.Fatalf("expected no legal pages, got %d", len(legal))
	}
	_ = utility
}

// The pre-fix working shape must keep working: a news-index page at a flat
// /news.html URL (gaswholesalers, robot-hands) never matched the child-prefix
// skip and must still classify into primary.
func TestClassifyPagesForNavNewsIndexAtFlatURL(t *testing.T) {
	logger := zap.NewNop()

	pages := []pageNavInfo{
		{Name: "news", Title: "News", URL: "/news.html", PageType: "news-index", InHeader: true, NavOrder: 40},
	}

	primary, _, _ := classifyPagesForNav(pages, logger)
	if len(primary) != 1 || primary[0].Name != "news" {
		t.Fatalf("news-index at /news.html must stay in primary nav, got %+v", primary)
	}
}
