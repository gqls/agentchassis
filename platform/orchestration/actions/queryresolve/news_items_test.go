// FILE: platform/orchestration/actions/queryresolve/news_items_test.go
//
// Tests for the query.latest_news / query.news_archive projection
// (bugs_open/027 rework). The load-bearing property: resolver output is
// rendered by text/template (which does NOT auto-escape) into live pages, and
// every text field originates in third-party RSS feeds — so projection MUST
// escape, and these tests pin that.

package queryresolve

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func sampleRawItems() []NewsItem {
	return []NewsItem{
		{
			Title:       "Panerai Submersible PAM01756",
			Summary:     "Reloj de buceo de 44 mm.",
			URL:         "https://trmagazine.es/panerai",
			Source:      "TR Magazine",
			PublishedAt: time.Now().Add(-72 * time.Hour),
			Topics:      []string{"Panerai", "buceo"},
		},
		{
			Title: "Zenith DEFY Extreme",
			URL:   "https://example.com/zenith",
			// zero PublishedAt: feed carried no date
		},
	}
}

func TestProjectNewsItemsShape(t *testing.T) {
	out := projectNewsItems(sampleRawItems(), true, zap.NewNop())
	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}

	first := out[0]
	if first["title"] != "Panerai Submersible PAM01756" {
		t.Errorf("title = %v", first["title"])
	}
	if first["date"] != "3 days ago" {
		t.Errorf("date = %v, want %q", first["date"], "3 days ago")
	}
	topics, ok := first["topics"].([]string)
	if !ok || len(topics) != 2 || topics[0] != "Panerai" {
		t.Errorf("topics = %v", first["topics"])
	}

	// Undated item renders an empty date so {{if .date}} omits it.
	if out[1]["date"] != "" {
		t.Errorf("zero PublishedAt should project empty date, got %v", out[1]["date"])
	}
}

func TestProjectNewsItemsOmitsTopicsForSnippet(t *testing.T) {
	out := projectNewsItems(sampleRawItems(), false, zap.NewNop())
	if _, present := out[0]["topics"]; present {
		t.Error("snippet projection must not carry topics — the homepage card has no tag row")
	}
}

// Feed text is third-party and the template engine does not escape.
func TestProjectNewsItemsEscapes(t *testing.T) {
	out := projectNewsItems([]NewsItem{{
		Title:   `Rolex <script>alert("x")</script>`,
		Summary: `A & B "quoted"`,
		URL:     `https://example.com/?a=1&b=2`,
		Source:  `<b>sneaky</b>`,
		Topics:  []string{`<i>tag</i>`},
	}}, true, zap.NewNop())

	m := out[0]
	for field, v := range map[string]string{
		"title":   m["title"].(string),
		"summary": m["summary"].(string),
		"url":     m["url"].(string),
		"source":  m["source"].(string),
	} {
		if strings.ContainsAny(v, "<>") {
			t.Errorf("%s not escaped: %q", field, v)
		}
	}
	if !strings.Contains(m["title"].(string), "&lt;script&gt;") {
		t.Errorf("title should carry escaped markup, got %q", m["title"])
	}
	if !strings.Contains(m["url"].(string), "&amp;") {
		t.Errorf("url ampersand should be escaped, got %q", m["url"])
	}
	if topics := m["topics"].([]string); len(topics) != 1 || strings.ContainsAny(topics[0], "<>") {
		t.Errorf("topics not escaped: %v", m["topics"])
	}
}

func TestNewsDisplayDate(t *testing.T) {
	now := time.Now()
	cases := []struct {
		in   time.Time
		want string
	}{
		{time.Time{}, ""},
		{now.Add(-30 * time.Minute), "Just now"},
		{now.Add(-1*time.Hour - time.Minute), "1 hour ago"},
		{now.Add(-5 * time.Hour), "5 hours ago"},
		{now.Add(-30 * time.Hour), "Yesterday"},
		{now.Add(-3 * 24 * time.Hour), "3 days ago"},
	}
	for _, c := range cases {
		if got := newsDisplayDate(c.in); got != c.want {
			t.Errorf("newsDisplayDate(%v) = %q, want %q", c.in, got, c.want)
		}
	}
	// Beyond a week: absolute date in the scripts' long form.
	old := now.Add(-30 * 24 * time.Hour)
	if got := newsDisplayDate(old); got != old.Format("2 Jan 2006") {
		t.Errorf("old date = %q, want %q", got, old.Format("2 Jan 2006"))
	}
}

func TestTruncateSummaryWordBoundary(t *testing.T) {
	long := strings.Repeat("palabra ", 40) // ~320 chars
	got := truncateSummary(long, 200)
	if len(got) > 204 {
		t.Errorf("truncated summary too long: %d chars", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis suffix, got %q", got[len(got)-10:])
	}
	if short := "corto"; truncateSummary(short, 200) != short {
		t.Error("short summary must pass through unchanged")
	}
}

// bugs_open/184 (council 060bcc0a r5/r6): the projection strips literal
// markdown from title and summary BEFORE escaping and BEFORE truncation, so a
// `# heading` or `[text](url)` in a third-party feed never reaches the
// template as marker characters. The raw NewsItem is untouched — the ingest
// record stays a faithful record; only the display projection is cleaned.
func TestProjectNewsItemsStripsLiteralMarkdown(t *testing.T) {
	t.Setenv("DISABLE_NEWS_MARKDOWN_STRIP", "")
	raw := []NewsItem{{
		Title:   "## Luke Littler wins **again**",
		Summary: "Read the [full report](https://example.com/report) and `stats`.",
		URL:     "https://example.com/report",
	}}
	out := projectNewsItems(raw, false, zap.NewNop())
	if got := out[0]["title"]; got != "Luke Littler wins again" {
		t.Errorf("title = %q, want markers stripped", got)
	}
	if got := out[0]["summary"]; got != "Read the full report and stats." {
		t.Errorf("summary = %q, want markers stripped", got)
	}
	// Input unchanged: the projection copies, it does not mutate the record.
	if raw[0].Title != "## Luke Littler wins **again**" {
		t.Errorf("raw title mutated: %q", raw[0].Title)
	}
}

// The kill switch is redeploy-free and restores the exact pre-strip output:
// HTML-escaped raw feed text, marker characters intact.
func TestProjectNewsItemsKillSwitchRestoresRawText(t *testing.T) {
	t.Setenv("DISABLE_NEWS_MARKDOWN_STRIP", "1")
	out := projectNewsItems([]NewsItem{{
		Title:   "## Luke Littler wins **again**",
		Summary: "Read the [full report](https://example.com/report).",
		URL:     "https://example.com/report",
	}}, false, zap.NewNop())
	if got := out[0]["title"]; got != "## Luke Littler wins **again**" {
		t.Errorf("with the kill switch set the title must be raw (escaped only), got %q", got)
	}
	if got := out[0]["summary"]; !strings.Contains(got.(string), "[full report](https://example.com/report)") {
		t.Errorf("with the kill switch set the summary must keep the link markdown, got %q", got)
	}
}

// Strip-before-truncate is load-bearing: a link straddling the 200-char
// truncation point would otherwise be cut mid-URL into "[text](https://exa..."
// — a half-pattern neither the detector nor the stripper can match, so the
// visitor sees it for ever. With the strip first, the link text survives and
// the truncation cuts prose.
func TestProjectNewsItemsStripsBeforeTruncating(t *testing.T) {
	t.Setenv("DISABLE_NEWS_MARKDOWN_STRIP", "")
	prefix := strings.Repeat("word ", 36) // 180 chars of prose; the raw link would straddle 200, the stripped text does not
	summary := prefix + "[Littler report](https://example.com/a/very/long/path/that/keeps/going) and more prose after the link."
	out := projectNewsItems([]NewsItem{{Title: "t", Summary: summary, URL: "https://x"}}, false, zap.NewNop())
	got := out[0]["summary"].(string)
	if strings.Contains(got, "](") || strings.Contains(got, "[Littler") {
		t.Errorf("truncated summary carries a half link pattern: %q", got)
	}
	if !strings.Contains(got, "Littler report") {
		t.Errorf("link text should survive the strip and precede the cut, got %q", got)
	}
}

// Escaping still happens AFTER the strip: the strip only ever deletes marker
// characters, and whatever remains is escaped exactly as before. A feed value
// that is both markdown-shaped and markup-bearing keeps its markup escaped and
// — by the detector's own suppression rule — keeps its backticks/brackets
// (a markup-bearing value is not a text-typed field).
func TestProjectNewsItemsStripThenEscape(t *testing.T) {
	t.Setenv("DISABLE_NEWS_MARKDOWN_STRIP", "")
	out := projectNewsItems([]NewsItem{{
		Title:   "**Bold** & <b>loud</b>",
		Summary: "plain **bold** & done",
		URL:     "https://x",
	}}, false, zap.NewNop())
	if got := out[0]["title"]; got != "Bold &amp; &lt;b&gt;loud&lt;/b&gt;" {
		t.Errorf("title = %q", got)
	}
	if got := out[0]["summary"]; got != "plain bold &amp; done" {
		t.Errorf("summary = %q", got)
	}
}
