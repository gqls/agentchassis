package actions

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
)

// readSourceFile reads a file in this package for the source-scanning assertions
// below. Source-scanning is a blunt instrument and the estate has a landmine
// about it (it makes comments load-bearing), so it is used here for exactly one
// purpose: pinning SQL clauses whose REMOVAL is silent and consequential.
func readSourceFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("cannot read %s: %v", name, err)
	}
	return string(b)
}

// TestSitemapActionIsRegistered guards the omission the council has caught twice
// on this tree: an action that exists but is not in the registry cannot be
// invoked, and nothing fails until a workflow names it.
func TestSitemapActionIsRegistered(t *testing.T) {
	entry, ok := GlobalActionRegistry["render_sitemap"]
	if !ok {
		t.Fatal("render_sitemap is not in GlobalActionRegistry — it cannot be invoked")
	}
	if entry.Handler == nil {
		t.Fatal("render_sitemap has a nil handler")
	}
	// The inverted gate is the design decision most likely to be "tidied" by
	// someone matching it to render_rss_feed. Pin the reason in the description
	// so a reader of the registry alone sees it.
	if !strings.Contains(entry.Description, "opt OUT") {
		t.Errorf("registry description should record that the gate is opt-OUT, got: %q", entry.Description)
	}
}

// TestAbsoluteURLHandlesEveryStoredShape — pages.url is stored inconsistently
// across the estate (leading slash, no slash, already absolute), and a sitemap
// full of malformed locs is worse than no sitemap because a crawler acts on it.
func TestAbsoluteURLHandlesEveryStoredShape(t *testing.T) {
	cases := []struct{ domain, path, want string }{
		{"example.co.uk", "/about.html", "https://example.co.uk/about.html"},
		{"example.co.uk", "about.html", "https://example.co.uk/about.html"},
		{"example.co.uk", "/", "https://example.co.uk/"},
		{"example.co.uk", "  /spaced.html  ", "https://example.co.uk/spaced.html"},
		{"example.co.uk", "https://other.com/x", "https://other.com/x"},
		{"example.co.uk", "http://other.com/x", "http://other.com/x"},
	}
	for _, c := range cases {
		if got := absoluteURL(c.domain, c.path); got != c.want {
			t.Errorf("absoluteURL(%q, %q) = %q, want %q", c.domain, c.path, got, c.want)
		}
	}
}

// TestSitemapXMLIsWellFormedAndCorrectlyNamespaced round-trips the marshalled
// document. A sitemap with the wrong namespace is silently ignored by crawlers —
// it parses, it serves, and it does nothing, which is the failure mode this
// estate keeps naming.
func TestSitemapXMLIsWellFormedAndCorrectlyNamespaced(t *testing.T) {
	set := sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []sitemapURL{
			{Loc: "https://example.co.uk/", LastMod: "2026-08-20"},
			{Loc: "https://example.co.uk/about.html"}, // no lastmod — must be omitted
		},
	}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	doc := xml.Header + string(body) + "\n"

	if !strings.Contains(doc, `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`) {
		t.Error("namespace missing or wrong — crawlers will ignore the file silently")
	}
	if strings.Count(doc, "<loc>") != 2 {
		t.Errorf("expected 2 <loc>, got %d", strings.Count(doc, "<loc>"))
	}
	if strings.Count(doc, "<lastmod>") != 1 {
		t.Errorf("lastmod must be omitted when empty, got %d", strings.Count(doc, "<lastmod>"))
	}

	var back sitemapURLSet
	if err := xml.Unmarshal([]byte(doc), &back); err != nil {
		t.Fatalf("emitted document does not parse: %v", err)
	}
	if len(back.URLs) != 2 || back.URLs[0].Loc != "https://example.co.uk/" {
		t.Errorf("round trip lost content: %+v", back.URLs)
	}
}

// TestSitemapQueryFiltersEveryVisibilityColumn is a source-scanning test, and it
// is deliberate. SEO-002's generalisable lesson is that a generator reading
// `pages` is a CONSUMER of every column that decides whether a page should be
// FOUND, and when a new one appears it will not fail — it will silently keep
// answering the old question. `noindex` arrived after the script and contradicted
// it for weeks.
//
// So this asserts the query still names each clause. It will fail loudly if
// someone simplifies the query, which is the point: removing one of these is a
// decision that should require deleting a test, not an oversight.
func TestSitemapQueryFiltersEveryVisibilityColumn(t *testing.T) {
	// ⚠ SCAN THE QUERY, NOT THE FILE. The first version of this test scanned the
	// whole source, and when the noindex filter was mutated out of the SQL the
	// test still PASSED — because the same words appear in this file's header
	// comment, and the scan found them there. That is the documented
	// "a source-scanning test makes your COMMENTS load-bearing" trap, walked into
	// while quoting it. Extracting the query first is what makes the assertion
	// discriminating; proven by re-running the same mutation, which now fails.
	src := extractPagesQuery(t, readSourceFile(t, "render_sitemap_action.go"))
	for _, clause := range []string{
		"status = 'active'",
		"noindex IS NOT TRUE",
		"deployed_at IS NOT NULL",
		"expires_at IS NULL OR expires_at > now()",
	} {
		if !strings.Contains(src, clause) {
			t.Errorf("visibility filter %q is gone from the pages query — if that was deliberate, "+
				"say so here and delete the case; if not, a page that should be hidden is now listed", clause)
		}
	}
}

// TestEmptySitemapIsRefused pins the decision that an empty sitemap is worse than
// none: it tells a crawler the site has no pages. The action reports
// rendered=false so a conditional skips the commit rather than publishing a file
// that actively misinforms.
func TestEmptySitemapIsRefused(t *testing.T) {
	src := readSourceFile(t, "render_sitemap_action.go")
	if !strings.Contains(src, "refusing to publish an empty sitemap") {
		t.Error("the empty-sitemap refusal is gone — an empty sitemap misinforms a crawler " +
			"rather than merely being useless")
	}
}

// extractPagesQuery returns just the SQL of the pages SELECT — the backtick
// string beginning with "SELECT url" — so the assertions above cannot be
// satisfied by prose elsewhere in the file.
func extractPagesQuery(t *testing.T, src string) string {
	t.Helper()
	start := strings.Index(src, "SELECT url")
	if start < 0 {
		t.Fatal("the pages query is gone: no `SELECT url` in render_sitemap_action.go")
	}
	end := strings.Index(src[start:], "`")
	if end < 0 {
		t.Fatal("could not find the end of the pages query literal")
	}
	q := src[start : start+end]
	if !strings.Contains(q, "FROM pages") {
		t.Fatalf("extracted text is not the pages query: %q", q)
	}
	return q
}
