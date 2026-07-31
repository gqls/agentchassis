package datahelpers

import "testing"

// The shapes below are not invented: every "live shape" case is a real
// pages.url value measured on 2026-07-31, and every rejection case is one the
// four pre-existing copies of this derivation got wrong (bugs_open/125).
func TestPageFilePathFromURL(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		want   string
		wantOK bool
	}{
		// ── Live shapes. 471 of 472 live urls are one of these two forms. ──
		{"root page", "/index.html", "index.html", true},
		{"top-level page", "/about.html", "about.html", true},
		{"subdirectory page — the whole point of the bug",
			"/tools/password-entropy.html", "tools/password-entropy.html", true},
		{"nested index", "/guides/building-it/index.html", "guides/building-it/index.html", true},
		{"blog post", "/blog/multi-agent-state-management.html", "blog/multi-agent-state-management.html", true},

		// ── The leading slash must go: the git adapter does domain + "/" + path,
		// so a kept slash yields "example.com//about.html". ──
		{"never returns a leading slash", "/a/b/c.html", "a/b/c.html", true},

		// ── Roots and directories ──
		{"site root", "/", "index.html", true},
		{"bare empty", "", "", false},
		{"whitespace only", "   ", "", false},
		{"directory style", "/guides/", "guides/index.html", true},

		// ── Extension handling: append only when there is none. The url is
		// authoritative — rewriting a declared extension 404s the canonical url. ──
		{"no extension gets .html", "/pricing", "pricing.html", true},
		{"nested no extension", "/guides/intro", "guides/intro.html", true},
		{"declared extension is preserved, NOT rewritten", "/legacy/report.php", "legacy/report.php", true},
		{"non-html asset preserved", "/feed.xml", "feed.xml", true},
		{"dot in a directory only", "/v1.2/intro", "v1.2/intro.html", true},

		// ── REJECTIONS. Each one falls back to the caller's own chain. ──
		{
			// The live instance. idea.uk's tool-audience-check has this url
			// while a DIFFERENT page (idea.uk/tools) owns /tools.html, so
			// stripping the fragment would overwrite that other page's file.
			name: "fragment url designates no file of its own",
			in:   "/tools.html#audience-check", want: "", wantOK: false,
		},
		{"query string is a view, not a file", "/search.html?q=x", "", false},
		{"absolute url is another origin", "https://example.com/a.html", "", false},
		{"protocol-relative url", "//example.com/a.html", "", false},
		{"backslash cannot be a tree path", `/a\b.html`, "", false},
		{"parent traversal", "/../etc/passwd", "", false},
		{"embedded traversal", "/a/../../b.html", "", false},
		{"double slash inside", "/a//b.html", "", false},
		{"single dot segment", "/./a.html", "", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := PageFilePathFromURL(c.in)
			if ok != c.wantOK {
				t.Fatalf("PageFilePathFromURL(%q) ok = %v, want %v (got path %q)", c.in, ok, c.wantOK, got)
			}
			if got != c.want {
				t.Errorf("PageFilePathFromURL(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// A rejected url must never yield a path — otherwise a caller that ignores ok
// would write to it. Belt and braces on the contract, since the whole fix
// depends on callers being able to trust the pair.
func TestPageFilePathFromURLRejectionYieldsNoPath(t *testing.T) {
	for _, in := range []string{
		"/tools.html#audience-check", "/x?y=1", "https://a/b", "//a/b", `/a\b`, "/../x", "/a//b",
	} {
		if got, ok := PageFilePathFromURL(in); ok || got != "" {
			t.Errorf("PageFilePathFromURL(%q) = (%q, %v); a rejection must yield an empty path", in, got, ok)
		}
	}
}

func TestPageDeployFilename(t *testing.T) {
	cases := []struct {
		name, url, page, want string
	}{
		{"url wins", "/tools/x.html", "x", "tools/x.html"},
		{"url wins over a disagreeing name", "/guides/a/index.html", "guide-a", "guides/a/index.html"},
		{"unusable url falls back to the name", "/tools.html#audience-check", "tool-audience-check", "tool-audience-check.html"},
		{"no url falls back to the name", "", "about", "about.html"},
		{"name already has the suffix", "", "about.html", "about.html"},
		{"index name", "", "index", "index.html"},
		{"home name", "", "home", "index.html"},
		{"nothing at all still yields a file", "", "", "index.html"},
		{"root url", "/", "index", "index.html"},

		// The three rerender copies special-cased name=="index" ahead of the
		// url. Dropping that is inert — measured 2026-07-31, 0 pages are named
		// index/home with a non-/index.html url — but pin the new precedence.
		{"url beats a page named index", "/landing.html", "index", "landing.html"},

		// ── The LANDMINES.md entry for getPageInfo: "a page whose url is /
		// deploys a file with NO NAME", because the filename was
		// TrimPrefix(url, "/") and the guard "keys on the page being NAMED
		// index, so an adopted or hand-made homepage under any other name
		// (home, landing, main) falls straight through it". Pinned here
		// because the council gate (corr 758f6e62) asked whether that warning
		// is absorbed by this call rather than merely absent today. It is: the
		// empty filename is now unreachable from any input — see also
		// TestPageDeployFilenameNeverEmpty. ──
		{"landmine: homepage named 'landing' at the root", "/", "landing", "index.html"},
		{"landmine: homepage named 'main' at the root", "/", "main", "index.html"},
		{"landmine: homepage named 'home' with no url", "", "home", "index.html"},
		{"landmine: an unnamed root page still gets a file", "/", "", "index.html"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := PageDeployFilename(c.url, c.page); got != c.want {
				t.Errorf("PageDeployFilename(%q, %q) = %q, want %q", c.url, c.page, got, c.want)
			}
		})
	}
}

// PageDeployFilename must never return "" — every caller assigns it straight
// into a files map key, where an empty key commits to the repo root.
func TestPageDeployFilenameNeverEmpty(t *testing.T) {
	for _, u := range []string{"", "/", "   ", "#x", "?q", "//", "/../", `\`} {
		for _, n := range []string{"", "  ", "index"} {
			if got := PageDeployFilename(u, n); got == "" {
				t.Errorf("PageDeployFilename(%q, %q) returned an empty filename", u, n)
			}
		}
	}
}
