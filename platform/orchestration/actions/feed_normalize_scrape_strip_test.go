// bugs_open/332: the second in-repo truncation of raw markdown.
//
// normalizeScrapeResults' Strategy-2 branch takes firecrawl's whole
// markdown_content as one item's summary and cuts it at 500 bytes. A cut through
// a link leaves "[text](https://…" — the half-pattern the display strip was
// blind to until this bug, and one that nothing downstream can repair once it is
// written to content_feed_items.
//
// This branch is INERT today (source_type='scrape' is 472 rows in 30 days with
// EMPTY summaries [MEASURED 2026-09-03], because Strategy 1 sets summary:"").
// It is guarded anyway: inert-today is the cheapest moment to close a
// manufacturing site, and a test is the only way anyone will know it is closed.

package actions

import (
	"strings"
	"testing"
	"unicode/utf8"

	"go.uber.org/zap"
)

func scrapePayload(markdown string) map[string]interface{} {
	return map[string]interface{}{
		"scrape_results": map[string]interface{}{
			"response_status": "complete",
			"response": map[string]interface{}{
				"data": map[string]interface{}{
					"url":              "https://example.com/news",
					"title":            "Example news index",
					"markdown_content": markdown,
				},
			},
		},
	}
}

func TestNormalizeScrapeStripsBeforeTheCut(t *testing.T) {
	// The fixture is OVER the 500-byte cut before stripping and UNDER it after,
	// which is the whole argument in one input: the URL was eating the budget.
	// Measured fleet-wide, 35.4% of a snippet's characters are URL characters —
	// an average of 69 of 197 — so stripping first does not merely avoid a
	// half-pattern, it hands the reader back the words the URL displaced.
	md := strings.Repeat("Noticias del sector. ", 20) +
		"[el informe completo](https://example.com/a/very/long/path/that/keeps/going/past/the/boundary) y mas texto."
	if len(md) <= 500 {
		t.Fatalf("premise wrong: fixture is %d bytes, must EXCEED the 500 cut before stripping", len(md))
	}

	items := normalizeScrapeResults(scrapePayload(md), "scrape_results", 10, zap.NewNop())
	if len(items) != 1 {
		t.Fatalf("expected the single-item Strategy-2 branch, got %d items", len(items))
	}
	summary, _ := items[0]["summary"].(string)

	if strings.Contains(summary, "](http") {
		t.Errorf("a half-pattern was written to the summary: %q", summary)
	}
	if strings.Contains(summary, "](") {
		t.Errorf("a bracket-paren sequence survived: %q", summary)
	}
	// The link TEXT must survive whole — a strip that emptied the value would
	// pass the two assertions above perfectly.
	if !strings.Contains(summary, "el informe completo y mas texto.") {
		t.Errorf("the strip removed visible text, not just markers: %q", summary)
	}
	// And because the URL is gone, the value now fits: no truncation at all.
	if strings.HasSuffix(summary, "...") {
		t.Errorf("still truncated after stripping — the URL should have freed the budget: %q", summary)
	}
	if !utf8.ValidString(summary) {
		t.Errorf("the cut produced invalid UTF-8: %q", summary)
	}
}

// The cut is now rune-safe. The original sliced bytes at exactly 500, which
// splits a multi-byte rune landing on the boundary.
func TestNormalizeScrapeCutIsRuneSafe(t *testing.T) {
	// "é" is two bytes; a run of them puts a rune across almost every offset.
	md := strings.Repeat("café ", 200)
	items := normalizeScrapeResults(scrapePayload(md), "scrape_results", 10, zap.NewNop())
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	summary, _ := items[0]["summary"].(string)
	if !utf8.ValidString(summary) {
		t.Errorf("invalid UTF-8 from the 500-byte cut: %q", summary)
	}
	if !strings.HasSuffix(summary, "...") {
		t.Errorf("expected the truncation marker, got %q", summary[len(summary)-8:])
	}
}

// Short content must pass through untouched — the guard must not become a
// transform that fires on everything.
func TestNormalizeScrapeLeavesShortCleanContentAlone(t *testing.T) {
	md := "Un resumen corto y limpio, sin markdown."
	items := normalizeScrapeResults(scrapePayload(md), "scrape_results", 10, zap.NewNop())
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if got, _ := items[0]["summary"].(string); got != md {
		t.Errorf("short clean content was changed:\n in:  %q\n out: %q", md, got)
	}
}
