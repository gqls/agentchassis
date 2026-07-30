package actions

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Covers the byte-preserving adoption path (fidelity=locked). The functions
// under test are the ones whose failures are SILENT: a wrong deploy path yields
// a file committed under the wrong name (or no name at all), and a wrong
// identity collides two pages into one through ON CONFLICT (site_id, name).

// ---------------------------------------------------------------------------
// urlToDeployPath
// ---------------------------------------------------------------------------

func TestURLToDeployPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		// The homepage is the case that silently breaks the deploy: pages.url
		// of "/" makes getPageInfo derive an EMPTY filename.
		{"bare root", "/", "/index.html", true},
		{"absolute root", "https://loancalculator.co.uk/", "/index.html", true},
		{"absolute root no slash", "https://loancalculator.co.uk", "/index.html", true},

		{"flat tool page", "https://loancalculator.co.uk/tools/standard-calc.html", "/tools/standard-calc.html", true},
		{"flat guide page", "/guides/jargon-buster.html", "/guides/jargon-buster.html", true},
		{"path only, no scheme", "/legal.html", "/legal.html", true},

		// Directory and extensionless URLs serve their index.html.
		{"trailing slash dir", "https://example.com/guides/", "/guides/index.html", true},
		{"extensionless", "https://example.com/guides", "/guides/index.html", true},
		{"nested dir", "https://example.com/a/b/", "/a/b/index.html", true},

		// Query and fragment are not part of the file path.
		{"query stripped", "https://example.com/tools/x.html?utm=1", "/tools/x.html", true},
		{"fragment stripped", "/tools/x.html#top", "/tools/x.html", true},
		{"query on scheme-less", "/tools/x.html?a=b", "/tools/x.html", true},

		{"empty", "", "", false},
		{"whitespace", "   ", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := urlToDeployPath(tc.in)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v (got path %q)", ok, tc.ok, got)
			}
			if !tc.ok {
				return
			}
			if got != tc.want {
				t.Errorf("urlToDeployPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
			// The invariant that actually protects the deploy: a usable path is
			// absolute and, once the leading slash is trimmed to form the
			// filename, non-empty.
			if !strings.HasPrefix(got, "/") {
				t.Errorf("path %q is not absolute", got)
			}
			if strings.TrimPrefix(got, "/") == "" {
				t.Errorf("path %q yields an EMPTY deploy filename", got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// verbatimPageIdentity
// ---------------------------------------------------------------------------

func TestVerbatimPageIdentity(t *testing.T) {
	cases := []struct {
		path     string
		wantName string
		wantType string
		ok       bool
	}{
		{"/index.html", "index", "landing", true},
		{"/legal.html", "legal", "content", true},
		{"/404.html", "404", "content", true},

		// Prefixes mirror CanonicalisePage so downstream consumers keyed on the
		// name prefix (tool eligibility) see adopted tools as tools.
		{"/tools/standard-calc.html", "tool-standard-calc", "tool", true},
		{"/tools/overpayment-calculator.html", "tool-overpayment-calculator", "tool", true},
		{"/guides/jargon-buster.html", "guide-jargon-buster", "guide", true},

		// Directory-style URLs reduce to the same identity shape as flat ones.
		{"/tools/standard-calc/index.html", "tool-standard-calc", "tool", true},

		// Section indexes.
		{"/tools/index.html", "tools-index", "section-index", true},
		{"/guides/index.html", "guides-index", "section-index", true},

		// Deeper nesting keeps every segment, so two distinct pages cannot
		// collapse to one name.
		{"/a/b/c.html", "a-b-c", "content", true},

		{"/", "", "", false},
		{"", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			name, pageType, ok := verbatimPageIdentity(tc.path)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if !tc.ok {
				return
			}
			if name != tc.wantName {
				t.Errorf("name = %q, want %q", name, tc.wantName)
			}
			if pageType != tc.wantType {
				t.Errorf("page_type = %q, want %q", pageType, tc.wantType)
			}
			// pages.page_type carries a kebab-case CHECK constraint, and
			// pages.name goes into work-item keys — neither may contain a slash.
			if strings.ContainsAny(name, "/ ") {
				t.Errorf("name %q contains a slash or space", name)
			}
			if strings.ContainsAny(pageType, "/ _") {
				t.Errorf("page_type %q is not kebab-case", pageType)
			}
		})
	}
}

// The flat and directory forms of the same tool must NOT be treated as two
// pages — but they are legitimately different URLs, so identity collapsing them
// is only safe because crawlPathIndex has already deduplicated by content.
func TestVerbatimPageIdentityDistinguishesRealSiblings(t *testing.T) {
	seen := map[string]string{}
	for _, p := range []string{
		"/index.html",
		"/legal.html",
		"/tools/standard-calc.html",
		"/tools/overpayment-calculator.html",
		"/tools/compare-loans.html",
		"/guides/jargon-buster.html",
		"/guides/hidden-loan-fees.html",
		"/guides/can-i-overpay.html",
	} {
		name, _, ok := verbatimPageIdentity(p)
		if !ok {
			t.Fatalf("%s failed to yield an identity", p)
		}
		if prev, clash := seen[name]; clash {
			t.Errorf("name collision: %q and %q both -> %q", prev, p, name)
		}
		seen[name] = p
	}
}

// ---------------------------------------------------------------------------
// crawlPathIndex
// ---------------------------------------------------------------------------

// buildCrawlPageIndex stores the SAME *crawlPageContent under several aliases
// per page. Deduplicating by URL string would create one page per alias; the
// dedup must therefore be by content pointer.
func TestCrawlPathIndexDeduplicatesAliasesOfOnePage(t *testing.T) {
	home := &crawlPageContent{Markdown: "# home", RawHTML: "<!DOCTYPE html><title>Home</title>"}
	tool := &crawlPageContent{Markdown: "# calc", RawHTML: "<!DOCTYPE html><title>Calc</title>"}

	index := map[string]*crawlPageContent{
		"https://loancalculator.co.uk/": home,
		"/":                             home,
		"https://loancalculator.co.uk/index.html":               home,
		"https://loancalculator.co.uk/tools/standard-calc.html": tool,
		"/tools/standard-calc.html":                             tool,
	}

	byPath, skipped := crawlPathIndex(index)
	if len(skipped) != 0 {
		t.Errorf("skipped = %v, want none", skipped)
	}
	if len(byPath) != 2 {
		t.Fatalf("got %d pages, want 2: %v", len(byPath), keysOf(byPath))
	}
	if _, ok := byPath["/index.html"]; !ok {
		t.Errorf("homepage missing; got %v", keysOf(byPath))
	}
	if _, ok := byPath["/tools/standard-calc.html"]; !ok {
		t.Errorf("tool page missing; got %v", keysOf(byPath))
	}
}

// A markdown-only page cannot be preserved verbatim. Adopting it as an empty
// page would be worse than reporting it, so it must be reported as skipped.
func TestCrawlPathIndexSkipsPagesWithoutRawHTML(t *testing.T) {
	withHTML := &crawlPageContent{Markdown: "# a", RawHTML: "<!DOCTYPE html><title>A</title>"}
	noHTML := &crawlPageContent{Markdown: "# b"}

	byPath, skipped := crawlPathIndex(map[string]*crawlPageContent{
		"/a.html": withHTML,
		"/b.html": noHTML,
	})

	if len(byPath) != 1 {
		t.Fatalf("got %d preserved pages, want 1: %v", len(byPath), keysOf(byPath))
	}
	if _, ok := byPath["/a.html"]; !ok {
		t.Errorf("page with rawHtml was dropped")
	}
	if len(skipped) != 1 || skipped[0] != "/b.html" {
		t.Errorf("skipped = %v, want [/b.html]", skipped)
	}
}

// Map iteration order is random; the chosen alias must not be. Re-running
// adoption has to produce the same pages.url every time.
func TestCrawlPathIndexIsDeterministic(t *testing.T) {
	build := func() map[string]string {
		c := &crawlPageContent{RawHTML: "<!DOCTYPE html><title>T</title>"}
		byPath, _ := crawlPathIndex(map[string]*crawlPageContent{
			"https://example.com/tools/x.html":       c,
			"/tools/x.html":                          c,
			"https://example.com/tools/x.html?utm=a": c,
		})
		out := map[string]string{}
		for p := range byPath {
			out[p] = ""
		}
		return out
	}
	first := build()
	for i := 0; i < 20; i++ {
		got := build()
		if len(got) != len(first) {
			t.Fatalf("run %d produced %d paths, first produced %d", i, len(got), len(first))
		}
		for p := range first {
			if _, ok := got[p]; !ok {
				t.Fatalf("run %d lost path %q", i, p)
			}
		}
	}
	if len(first) != 1 {
		t.Errorf("three aliases of one page yielded %d paths: %v", len(first), first)
	}
}

func TestCrawlPathIndexEmpty(t *testing.T) {
	byPath, skipped := crawlPathIndex(map[string]*crawlPageContent{})
	if len(byPath) != 0 || len(skipped) != 0 {
		t.Errorf("empty index should yield nothing, got %v / %v", keysOf(byPath), skipped)
	}
}

// ---------------------------------------------------------------------------
// fidelity + metadata extraction
// ---------------------------------------------------------------------------

func TestAdoptionFidelity(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]interface{}
		want string
	}{
		{"absent", map[string]interface{}{}, ""},
		{"locked", map[string]interface{}{
			"input_data": map[string]interface{}{"fidelity": "locked"}}, "locked"},
		{"locked mixed case and padded", map[string]interface{}{
			"input_data": map[string]interface{}{"fidelity": "  LOCKED "}}, "locked"},
		{"high is not locked", map[string]interface{}{
			"input_data": map[string]interface{}{"fidelity": "high"}}, "high"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := adoptionFidelity(tc.in); got != tc.want {
				t.Errorf("adoptionFidelity = %q, want %q", got, tc.want)
			}
		})
	}

	// The gate itself: only the exact locked value may divert adoption.
	for _, v := range []string{"", "high", "medium", "low", "new", "lock", "locked-ish"} {
		in := map[string]interface{}{"input_data": map[string]interface{}{"fidelity": v}}
		if adoptionFidelity(in) == fidelityLocked {
			t.Errorf("fidelity %q must NOT select the verbatim path", v)
		}
	}
}

func TestExtractHTMLTitleAndMeta(t *testing.T) {
	doc := `<!DOCTYPE html><html lang="en-GB"><head>
	<title>Loan Repayment Calculator | loancalculator.co.uk</title>
	<meta name="description" content="Work out monthly repayments &amp; total cost.">
	</head><body><h1>Hi</h1></body></html>`

	if got, want := extractHTMLTitle(doc), "Loan Repayment Calculator | loancalculator.co.uk"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
	if got, want := extractHTMLMetaDescription(doc), "Work out monthly repayments & total cost."; got != want {
		t.Errorf("meta = %q, want %q", got, want)
	}

	// Multi-line titles collapse; absent values return empty, not garbage.
	if got, want := extractHTMLTitle("<title>\n  Spread\n  Over Lines\n</title>"), "Spread Over Lines"; got != want {
		t.Errorf("collapsed title = %q, want %q", got, want)
	}
	if got := extractHTMLTitle("<html><body>no title here</body></html>"); got != "" {
		t.Errorf("missing title should be empty, got %q", got)
	}
	if got := extractHTMLMetaDescription("<html><head><title>x</title></head></html>"); got != "" {
		t.Errorf("missing meta should be empty, got %q", got)
	}
}

func keysOf(m map[string]*crawlPageContent) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// ---------------------------------------------------------------------------
// loadVerbatimPageHTML — the guard that decides whether a page bypasses
// assembly entirely. A false positive means a page silently stops being
// maintained by the platform, so each refusal path is asserted.
// ---------------------------------------------------------------------------

func verbatimCols() []string {
	return []string{"count", "verbatim_count", "rendered_html"}
}

func TestLoadVerbatimPageHTMLShipsStoredBytes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	pageID := uuid.New()
	doc := `<!DOCTYPE html><html lang="en-GB"><head><title>Calc</title></head><body>x</body></html>`

	mock.ExpectQuery("FROM page_components").
		WithArgs(pageID, deployModeVerbatim).
		WillReturnRows(sqlmock.NewRows(verbatimCols()).AddRow(1, 1, doc))

	got, ok := loadVerbatimPageHTML(context.Background(), db, pageID, zap.NewNop())
	if !ok {
		t.Fatal("expected the page to be recognised as verbatim")
	}
	if got != doc {
		t.Errorf("returned HTML differs from stored bytes:\n got %q\nwant %q", got, doc)
	}
}

func TestLoadVerbatimPageHTMLRefusals(t *testing.T) {
	doc := `<!DOCTYPE html><title>x</title>`

	cases := []struct {
		name          string
		componentRows int
		verbatimRows  int
		html          interface{}
		why           string
	}{
		{"not a verbatim page", 3, 0, nil,
			"a normal page must assemble as before"},
		{"verbatim but extra components attached", 2, 1, doc,
			"the page is no longer a single stored document"},
		{"verbatim but empty html", 1, 1, "",
			"deploying an empty page is worse than assembling"},
		{"verbatim but null html", 1, 1, nil,
			"a NULL rendered_html must not ship"},
		{"verbatim but whitespace only", 1, 1, "   \n\t ",
			"whitespace is not a document"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			pageID := uuid.New()
			mock.ExpectQuery("FROM page_components").
				WithArgs(pageID, deployModeVerbatim).
				WillReturnRows(sqlmock.NewRows(verbatimCols()).
					AddRow(tc.componentRows, tc.verbatimRows, tc.html))

			if _, ok := loadVerbatimPageHTML(context.Background(), db, pageID, zap.NewNop()); ok {
				t.Errorf("expected refusal (%s), but the page was shipped verbatim", tc.why)
			}
		})
	}
}

// A lookup failure must fall back to assembly — the behaviour every page had
// before this path existed — rather than fail the rerender.
func TestLoadVerbatimPageHTMLFallsBackOnQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	pageID := uuid.New()
	mock.ExpectQuery("FROM page_components").
		WithArgs(pageID, deployModeVerbatim).
		WillReturnError(errors.New("connection reset"))

	if _, ok := loadVerbatimPageHTML(context.Background(), db, pageID, zap.NewNop()); ok {
		t.Error("a failed lookup must not report the page as verbatim")
	}
}

// ---------------------------------------------------------------------------
// portedPageComponentID — returns a COUNT so the caller can refuse ambiguity
// rather than silently pick. cmd/webdesignport resolves ties by the NEWEST row
// and this path by the OLDEST, so a second active row means the two porters
// write different components; applyVerbatimAdoption fails on found > 1.
// ---------------------------------------------------------------------------

func TestPortedPageComponentID(t *testing.T) {
	newID := func() string { return uuid.New().String() }

	t.Run("exactly one active row is the healthy case", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		want := uuid.New()
		mock.ExpectBegin()
		mock.ExpectQuery("FROM content_components").
			WithArgs(portedPageSlot).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(want))
		tx, _ := db.Begin()
		got, found, err := portedPageComponentID(context.Background(), tx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found != 1 {
			t.Errorf("found = %d, want 1", found)
		}
		if got != want {
			t.Errorf("id = %s, want %s", got, want)
		}
	})

	t.Run("no active row fails rather than seeding a second", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		mock.ExpectBegin()
		mock.ExpectQuery("FROM content_components").
			WithArgs(portedPageSlot).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		tx, _ := db.Begin()
		if _, _, err := portedPageComponentID(context.Background(), tx); err == nil {
			t.Error("expected an error when no ported-page component exists")
		}
	})

	t.Run("two active rows are REPORTED, not silently resolved", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		oldest, newest := newID(), newID()
		mock.ExpectBegin()
		mock.ExpectQuery("FROM content_components").
			WithArgs(portedPageSlot).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(oldest).AddRow(newest))
		tx, _ := db.Begin()
		got, found, err := portedPageComponentID(context.Background(), tx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// The count is the signal the caller refuses on; without it the ambiguity
		// would ship as a warning at most.
		if found != 2 {
			t.Errorf("found = %d, want 2 — the caller cannot refuse what it is not told", found)
		}
		if got.String() != oldest {
			t.Errorf("id = %s, want the OLDEST (%s) — ordering must be deterministic", got, oldest)
		}
	})
}
