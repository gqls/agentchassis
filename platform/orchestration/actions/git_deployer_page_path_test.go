package actions

import (
	"testing"

	"go.uber.org/zap"
)

// determinePageFilename resolves where a single-page commit is written.
// bugs_open/125: it consulted slug/name/page_name/filename/id and never the
// page's canonical `url`, so a page at /tools/x.html was published to /x.html —
// a second, orphaned, fetchable copy rather than a moved page. These cases pin
// the new precedence AND the whole of the old fallback chain, because the
// fallback is what a url-less caller still depends on.
func TestDeterminePageFilenamePrefersCanonicalURL(t *testing.T) {
	log := zap.NewNop()
	cfg := map[string]interface{}{"page_field": "current_page"}

	withPage := func(p map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{"current_page": p}
	}

	cases := []struct {
		name string
		page map[string]interface{}
		want string
	}{
		{
			// The exact live instance from the bug report: page name
			// "ai-agent-roi-estimator", canonical url "/tools/…". Before the
			// fix this returned "ai-agent-roi-estimator.html" and published a
			// duplicate at the site root.
			name: "url beats name for a subdirectory page",
			page: map[string]interface{}{
				"name": "ai-agent-roi-estimator",
				"url":  "/tools/ai-agent-roi-estimator.html",
			},
			want: "tools/ai-agent-roi-estimator.html",
		},
		{
			name: "url beats slug too",
			page: map[string]interface{}{
				"slug": "intro",
				"name": "intro",
				"url":  "/guides/building-it/index.html",
			},
			want: "guides/building-it/index.html",
		},
		{
			name: "url and name agreeing is unchanged behaviour",
			page: map[string]interface{}{"name": "about", "url": "/about.html"},
			want: "about.html",
		},
		{
			// A url that names no file of its own must NOT be sanitised into
			// one: /tools.html belongs to a different page, and stripping the
			// fragment would aim this commit at that page's file.
			name: "fragment url falls back to the name",
			page: map[string]interface{}{
				"name": "tool-audience-check",
				"url":  "/tools.html#audience-check",
			},
			want: "tool-audience-check.html",
		},
		{
			name: "empty url falls back to the name",
			page: map[string]interface{}{"name": "about", "url": ""},
			want: "about.html",
		},

		// ── The pre-existing chain, unchanged, for callers with no url ──
		{"slug", map[string]interface{}{"slug": "pricing"}, "pricing.html"},
		{"name", map[string]interface{}{"name": "pricing"}, "pricing.html"},
		{"page_name", map[string]interface{}{"page_name": "pricing"}, "pricing.html"},
		{"filename verbatim", map[string]interface{}{"filename": "custom.htm"}, "custom.htm"},
		{"id", map[string]interface{}{"id": "abc123"}, "abc123.html"},
		{"slug outranks name", map[string]interface{}{"slug": "a", "name": "b"}, "a.html"},
		{"nothing usable", map[string]interface{}{"title": "Only a title"}, "index.html"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := determinePageFilename(withPage(c.page), cfg, log); got != c.want {
				t.Errorf("determinePageFilename(%v) = %q, want %q", c.page, got, c.want)
			}
		})
	}
}

// The url branch must not disturb the higher-priority config overrides, which
// exist for non-page files (CSS, JS) and for callers that computed the name
// themselves.
func TestDeterminePageFilenameConfigOverridesStillWin(t *testing.T) {
	log := zap.NewNop()
	data := map[string]interface{}{
		"current_page": map[string]interface{}{"name": "about", "url": "/about.html"},
		"rendered":     map[string]interface{}{"filename": "explicit.html"},
	}

	if got := determinePageFilename(data, map[string]interface{}{
		"filename_field": "rendered.filename", "page_field": "current_page",
	}, log); got != "explicit.html" {
		t.Errorf("filename_field must outrank the page url; got %q", got)
	}

	if got := determinePageFilename(data, map[string]interface{}{
		"file_path": "assets/css/styles.css", "page_field": "current_page",
	}, log); got != "assets/css/styles.css" {
		t.Errorf("file_path must outrank the page url; got %q", got)
	}

	// No page_field at all — the historical default.
	if got := determinePageFilename(data, map[string]interface{}{}, log); got != "index.html" {
		t.Errorf("no page_field should default to index.html; got %q", got)
	}
}

// A leading slash would become "example.com//tools/x.html" once the git adapter
// prefixes the domain (github_client.go CommitToRepo). Nothing downstream
// normalises it, so the resolver is the only place this can be enforced.
func TestDeterminePageFilenameNeverReturnsALeadingSlash(t *testing.T) {
	log := zap.NewNop()
	for _, u := range []string{"/index.html", "/about.html", "/tools/x.html", "/guides/a/index.html", "/"} {
		got := determinePageFilename(
			map[string]interface{}{"current_page": map[string]interface{}{"url": u, "name": "n"}},
			map[string]interface{}{"page_field": "current_page"}, log)
		if got == "" || got[0] == '/' {
			t.Errorf("url %q resolved to %q — a leading slash produces a double slash in the git tree path", u, got)
		}
	}
}
