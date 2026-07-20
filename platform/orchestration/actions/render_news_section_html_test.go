// FILE: platform/orchestration/actions/render_news_section_html_test.go
//
// Tests for server-side news rendering (bugs_open/027).
//
// The load-bearing test here is TestServerMarkupMatchesJSGuardSelector. The fix
// has two halves that must agree: migration 178 makes news-listing.js preserve
// the container when it already holds `article.news-list-item`, and this Go code
// is what puts that element there. If either side changes the class name, the
// guard silently stops guarding and server-rendered news gets wiped on the next
// empty feed — with nothing failing. That test pins the contract.

package actions

import (
	"strings"
	"testing"
)

func listingTemplate(inner string) string {
	return `<section data-component="news-listing" class="news-listing-section" id="news-listing">
  <div class="news-listing-container">
    <div class="news-listing-header">
      <h1 class="news-listing-title">Noticias</h1>
    </div>
    <div class="news-listing-items" id="news-listing-items">
      ` + inner + `
    </div>
    <div class="news-listing-footer" id="news-listing-footer" style="display:none;">
      <p class="news-listing-count" id="news-listing-count"></p>
    </div>
  </div>
  <script src="/tools/assets/news-listing.js"></script>
</section>`
}

func sampleItems() []newsJSONItem {
	return []newsJSONItem{
		{
			Title:   "Panerai Submersible PAM01756",
			Summary: "Reloj de buceo de 44 mm.",
			URL:     "https://trmagazine.es/panerai",
			Source:  "TR Magazine",
			Date:    "3d ago",
			Topics:  "Panerai, buceo",
		},
		{
			Title: "Zenith DEFY Extreme",
			URL:   "https://example.com/zenith",
			Date:  "1h ago",
		},
	}
}

// The contract between migration 178's JS guard and this renderer.
func TestServerMarkupMatchesJSGuardSelector(t *testing.T) {
	// The selector migration 178 uses: container.querySelector("article.news-list-item")
	const guardTag = `<article class="news-list-item"`

	got := renderNewsListingItemsHTML(sampleItems())
	if !strings.Contains(got, guardTag) {
		t.Fatalf("server markup must contain %q or migration 178's guard cannot see it.\ngot: %s", guardTag, got)
	}
}

func TestRenderNewsListingItemsHTML(t *testing.T) {
	got := renderNewsListingItemsHTML(sampleItems())

	for _, want := range []string{
		`<article class="news-list-item">`,
		`<h3 class="news-list-item-title"><a href="https://trmagazine.es/panerai" target="_blank" rel="noopener noreferrer">Panerai Submersible PAM01756</a></h3>`,
		`<p class="news-list-item-summary">Reloj de buceo de 44 mm.</p>`,
		`<span class="news-list-item-source">TR Magazine</span>`,
		`<span class="news-list-item-date">3 days ago</span>`,
		`<span class="news-list-tag">Panerai</span>`,
		`<span class="news-list-tag">buceo</span>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}

	// Optional fields must be omitted, not emitted empty.
	if strings.Contains(got, `<p class="news-list-item-summary"></p>`) {
		t.Error("empty summary should be omitted entirely")
	}
	if strings.Contains(got, `news-list-item-topics"></div>`) {
		t.Error("empty topics should be omitted entirely")
	}
	if n := strings.Count(got, "<article"); n != 2 {
		t.Errorf("expected 2 articles, got %d", n)
	}
}

func TestRenderLatestNewsCardsHTML(t *testing.T) {
	got := renderLatestNewsCardsHTML(sampleItems())
	for _, want := range []string{
		`<article class="news-card">`,
		`<h3 class="news-card-title">`,
		`<span class="news-source">TR Magazine</span>`,
		`<time class="news-date">3 days ago</time>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// Feed text is third-party. The scripts do not escape; the server must.
func TestRenderEscapesThirdPartyFeedText(t *testing.T) {
	got := renderNewsListingItemsHTML([]newsJSONItem{{
		Title:   `Rolex <script>alert("x")</script>`,
		Summary: `A & B "quoted"`,
		URL:     `https://example.com/?a=1&b=2`,
	}})

	if strings.Contains(got, "<script>alert") {
		t.Errorf("raw script tag survived escaping:\n%s", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") {
		t.Errorf("expected escaped title, got:\n%s", got)
	}
	if !strings.Contains(got, "https://example.com/?a=1&amp;b=2") {
		t.Errorf("expected escaped URL ampersand, got:\n%s", got)
	}
}

func TestInjectNewsItems_FirstRunDisplacesPlaceholder(t *testing.T) {
	html := listingTemplate(`<p class="news-listing-loading">Loading latest news...</p>`)
	inner := renderNewsListingItemsHTML(sampleItems())

	got, ok := injectNewsItems(html, newsListingAnchors, inner)
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	if strings.Contains(got, "news-listing-loading") {
		t.Error("loading placeholder should have been displaced")
	}
	if !strings.Contains(got, `<article class="news-list-item"`) {
		t.Error("items should be present after injection")
	}
	// Surrounding structure must survive intact.
	for _, want := range []string{
		`<h1 class="news-listing-title">Noticias</h1>`,
		`<div class="news-listing-footer"`,
		`<script src="/tools/assets/news-listing.js"></script>`,
		`</section>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("injection damaged surrounding markup, missing %q", want)
		}
	}
}

// Running the feed render repeatedly must not stack items.
func TestInjectNewsItems_Idempotent(t *testing.T) {
	html := listingTemplate(`<p class="news-listing-loading">Loading latest news...</p>`)
	inner := renderNewsListingItemsHTML(sampleItems())

	once, ok := injectNewsItems(html, newsListingAnchors, inner)
	if !ok {
		t.Fatal("first injection failed")
	}
	twice, ok := injectNewsItems(once, newsListingAnchors, inner)
	if !ok {
		t.Fatal("second injection failed")
	}
	if once != twice {
		t.Error("injection is not idempotent — repeated renders would stack items")
	}
	if n := strings.Count(twice, "<article"); n != 2 {
		t.Errorf("expected 2 articles after two injections, got %d", n)
	}
}

func TestInjectNewsItems_LatestNewsDisplacesNoscript(t *testing.T) {
	html := `<section data-component="latest-news"><div class="container">
    <div class="news-grid" id="news-container">
      <noscript><p class="news-empty">Enable JavaScript to see the latest news.</p></noscript>
    </div>
    <div id="news-footer"></div>
  </div></section>`

	got, ok := injectNewsItems(html, latestNewsAnchors, renderLatestNewsCardsHTML(sampleItems()))
	if !ok {
		t.Fatal("expected injection to succeed")
	}
	if strings.Contains(got, "<noscript>") {
		t.Error("noscript fallback should be displaced once real content is server-rendered")
	}
	if !strings.Contains(got, `<div id="news-footer"></div>`) {
		t.Error("sibling footer must survive")
	}
}

// A component whose template has drifted must be left alone, not corrupted.
func TestInjectNewsItems_RefusesOnMissingAnchors(t *testing.T) {
	cases := map[string]string{
		"no start anchor": `<section><div class="whatever"></div><div class="news-listing-footer"></div></section>`,
		"no end anchor":   `<section><div id="news-listing-items"><p>x</p></div></section>`,
		"empty html":      ``,
	}
	inner := renderNewsListingItemsHTML(sampleItems())
	for name, html := range cases {
		got, ok := injectNewsItems(html, newsListingAnchors, inner)
		if ok {
			t.Errorf("%s: expected refusal, got success", name)
		}
		if got != html {
			t.Errorf("%s: html was modified despite refusal", name)
		}
	}
}

func TestInjectNewsItems_EmptyInnerIsNoOp(t *testing.T) {
	html := listingTemplate(`<p class="news-listing-loading">Loading latest news...</p>`)
	got, ok := injectNewsItems(html, newsListingAnchors, "")
	if ok {
		t.Error("empty inner should not report success")
	}
	if got != html {
		t.Error("empty inner must leave the placeholder in place")
	}
}

func TestExpandRelativeNewsDate(t *testing.T) {
	cases := map[string]string{
		"3d ago":     "3 days ago",
		"1d ago":     "1 day ago",
		"5h ago":     "5 hours ago",
		"1h ago":     "1 hour ago",
		"2w ago":     "2 weeks ago",
		"10m ago":    "10 minutes ago",
		"Just now":   "Just now",
		"Yesterday":  "Yesterday",
		"2 Jan 2026": "2 Jan 2026",
		"":           "",
		"weird ago":  "weird ago",
	}
	for in, want := range cases {
		if got := expandRelativeNewsDate(in); got != want {
			t.Errorf("expandRelativeNewsDate(%q) = %q, want %q", in, got, want)
		}
	}
}
