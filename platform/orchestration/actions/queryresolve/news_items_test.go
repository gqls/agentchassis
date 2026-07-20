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
	out := projectNewsItems(sampleRawItems(), true)
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
	out := projectNewsItems(sampleRawItems(), false)
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
	}}, true)

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
