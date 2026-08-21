// FILE: platform/orchestration/actions/discovery_checks/check_stylesheet_gutted_test.go
//
// bugs_open/198 + bugs_open/211 — pins stylesheet_gutted.
//
// UNLIKE ITS SIBLING, THIS CHECK HAS LIVE POSITIVES, and the two shapes below are
// both taken from real incidents rather than invented:
//
//	the 198 shape   remortgagecalculator.uk 2026-08-21: a 17,403-byte stylesheet
//	                replaced by 136 bytes of css-patch-agent contrast rules. Every
//	                --color-* definition gone; the file still serves 200.
//	the 211 shape   ai-agent-orchestration.com: a ~26KB stylesheet that is missing
//	                only the renderer's alias :root block — --color-heading defined
//	                ZERO times. A byte-floor predicate cannot see this one, which is
//	                why the predicate is definition COVERAGE.
//
// Every guard is proven load-bearing by breaking it and watching a NAMED test
// fail. A guard no test can be made to fail against is decoration:
//
//	mutation                                            test that catches it
//	--------------------------------------------------  ---------------------------------
//	count var(--x, fallback) as a hard reference        FallbackCarryingReferenceIsNotAGap
//	drop the css_snippets definition source             SnippetDefinedPropertyIsNotAGap
//	drop the in-page definition source                  InPageDefinedPropertyIsNotAGap
//	treat a non-2xx stylesheet as "defines nothing"     RefusedStylesheetFilesNothing
//	treat a transport error as an empty stylesheet      FetchFailureFilesNothing
//	retract when blinded                                RefusedStylesheetFilesNothing
//	regex the HTML instead of parsing the DOM           VarInsideScriptTextIsNotAReference
//	scan CSS comments as code                           VarInsideCSSCommentIsNotAReference
//	give the item a handler_agent                       NeverRoutesToAHandler
//	make the item key URL-shaped                        ItemKeyIsConstantAndNotURLShaped
//	fetch external stylesheets                          ExternalStylesheetIsNotFetched

package discovery_checks

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// ── fixtures: the two real incident shapes ──────────────────────────────────

// healthyStylesheet is the shape a webdesign-agent run produces: a :root block
// defining the vocabulary, plus the renderer's alias block.
const healthyStylesheet = `
:root {
  --color-primary:    #1E3A5F;
  --color-text:       #1A1A2E;
  --color-background: #F7F8F6;
  --color-border:     #DFE4EA;
}
body { color: var(--color-text); background: var(--color-background); }
:root { --color-heading: var(--color-text); --hero-ink: var(--color-text); }
`

// guttedStylesheet is remortgagecalculator.uk as served on 2026-08-21: the
// entire base replaced by two css-patch-agent appends. It is valid CSS, it
// returns 200, and it defines nothing.
const guttedStylesheet = `

/* css-patch-agent 2026-08-21: contrast */
p.P { color: #4b5563; }

/* css-patch-agent 2026-08-21: contrast */
p.P { color: #3d4451; }`

// alias-block-missing: the 211 shape. Large, plausible, and short exactly one
// block — --color-heading and --hero-ink are never defined.
const stylesheetMissingAliasBlock = `
:root {
  --color-primary:    #1E3A5F;
  --color-text:       #1A1A2E;
  --color-background: #F7F8F6;
  --color-border:     #DFE4EA;
}
body { color: var(--color-text); background: var(--color-background); }
.card { border: 1px solid var(--color-border); }
`

// pageUsingTokens is deployed markup that consumes the vocabulary the stylesheet
// is supposed to define, in the two places a component actually carries CSS.
const pageUsingTokens = `<html><head>
<link rel="stylesheet" href="/assets/css/styles.css">
</head><body>
<style>.hero { color: var(--color-text); background: var(--color-background); }
.hero h1 { color: var(--color-heading); }</style>
<div class="hero"><h1 style="border-color: var(--color-border)">Remortgage</h1></div>
</body></html>`

// ── harness ─────────────────────────────────────────────────────────────────

// stubSheetFetch replaces the network with a table of responses and records
// every call, so a test can assert what was NOT requested as well as what was.
type stubSheetFetch struct {
	mu     sync.Mutex
	body   map[string]string
	status map[string]int
	err    map[string]error
	calls  []string
}

func newStubSheetFetch() *stubSheetFetch {
	return &stubSheetFetch{
		body:   map[string]string{},
		status: map[string]int{},
		err:    map[string]error{},
	}
}

func (s *stubSheetFetch) install(t *testing.T) {
	t.Helper()
	prev := fetchSiteStylesheet
	fetchSiteStylesheet = func(_ context.Context, target string) (int, string, error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.calls = append(s.calls, target)
		if e, ok := s.err[target]; ok {
			return 0, "", e
		}
		code := 200
		if c, ok := s.status[target]; ok {
			code = c
		}
		return code, s.body[target], nil
	}
	t.Cleanup(func() { fetchSiteStylesheet = prev })
}

func (s *stubSheetFetch) totalCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

func (s *stubSheetFetch) called(target string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.calls {
		if c == target {
			return true
		}
	}
	return false
}

type sheetPage struct {
	url  string
	html string
}

// runStylesheetGutted drives the real check. Query order mirrors Run:
// sites (domain), page_components, site_components, then css_snippets.
func runStylesheetGutted(t *testing.T, domain string, pages []sheetPage, chromeHTML string, snippets []string) *CheckResult {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow(domain))

	pageRows := sqlmock.NewRows([]string{"rendered_html", "url"})
	for _, p := range pages {
		pageRows.AddRow(p.html, p.url)
	}
	// As in the sibling check's tests, the expectation requires the SHARED
	// liveness builders rather than the table name alone: if anyone re-hand-rolls
	// the predicate as `deployed_at IS NOT NULL`, the mock stops matching and
	// every test here fails. A check that silently omits live pages reports clean
	// for the wrong reason (bugs_open/185).
	mock.ExpectQuery(regexp.QuoteMeta(datahelpers.PageHasShippedPredicateFor("p"))).
		WillReturnRows(pageRows)

	chromeRows := sqlmock.NewRows([]string{"rendered_html"})
	if chromeHTML != "" {
		chromeRows.AddRow(chromeHTML)
	}
	mock.ExpectQuery("FROM site_components").WillReturnRows(chromeRows)

	snippetRows := sqlmock.NewRows([]string{"css_content"})
	for _, s := range snippets {
		snippetRows.AddRow(s)
	}
	mock.ExpectQuery("FROM css_snippets").WillReturnRows(snippetRows)

	res, err := (&StylesheetGuttedCheck{}).Run(DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.New(),
		Pipeline:  "design",
		AgentType: "test",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("StylesheetGuttedCheck.Run: %v", err)
	}
	return res
}

func missingFromSpec(t *testing.T, res *CheckResult) []string {
	t.Helper()
	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d", len(res.WorkItems))
	}
	var out []string
	spec := res.WorkItems[0].SpecJSON
	for _, name := range []string{"--color-text", "--color-background", "--color-heading", "--color-border", "--hero-ink", "--color-primary"} {
		if strings.Contains(spec, `"`+name+`"`) {
			out = append(out, name)
		}
	}
	return out
}

// ── the finding branch: the two real incident shapes ────────────────────────

// THE DEFECT THIS CHECK EXISTS FOR, induced from the real artefact.
// remortgagecalculator.uk, 2026-08-21: styles.css serves 200 and defines nothing.
func TestStylesheetGutted_ClobberedStylesheetFilesExactlyOneItem(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://remortgagecalculator.uk/assets/css/styles.css"] = guttedStylesheet
	sf.install(t)

	res := runStylesheetGutted(t, "remortgagecalculator.uk",
		[]sheetPage{{url: "/index.html", html: pageUsingTokens}}, "", nil)

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item for a gutted stylesheet, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Fatalf("a gutted stylesheet must never retract, got %d", len(res.Resolved))
	}
	if res.WorkItems[0].ItemType != "stylesheet_gutted" {
		t.Fatalf("item type = %q", res.WorkItems[0].ItemType)
	}
	if res.WorkItems[0].Severity != "high" {
		t.Fatalf("a whole-site visual outage should be high severity, got %q", res.WorkItems[0].Severity)
	}
	got := missingFromSpec(t, res)
	// Every token the page consumes without a fallback is undefined.
	for _, want := range []string{"--color-text", "--color-background", "--color-heading", "--color-border"} {
		found := false
		for _, g := range got {
			if g == want {
				found = true
			}
		}
		if !found {
			t.Errorf("expected %s in the missing set, got %v", want, got)
		}
	}
}

// THE 211 SHAPE, and the reason the predicate is coverage rather than a byte
// floor: this stylesheet is large and plausible and still breaks the pages.
func TestStylesheetGutted_LargeStylesheetMissingOnlyTheAliasBlockStillFires(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://ai-agent-orchestration.com/assets/css/styles.css"] = stylesheetMissingAliasBlock
	sf.install(t)

	res := runStylesheetGutted(t, "ai-agent-orchestration.com",
		[]sheetPage{{url: "/index.html", html: pageUsingTokens}}, "", nil)

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item for the missing alias block, got %d", len(res.WorkItems))
	}
	got := missingFromSpec(t, res)
	if len(got) != 1 || got[0] != "--color-heading" {
		t.Fatalf("only --color-heading should be missing, got %v", got)
	}
	// A byte floor would have passed this file — that is the whole argument.
	if len(stylesheetMissingAliasBlock) < 100 {
		t.Fatalf("fixture no longer represents a large stylesheet")
	}
}

// ── the healthy branch and its retraction ───────────────────────────────────

func TestStylesheetGutted_HealthyStylesheetRetractsWithTheExactKey(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://remortgagecalculator.uk/assets/css/styles.css"] = healthyStylesheet
	sf.install(t)

	res := runStylesheetGutted(t, "remortgagecalculator.uk",
		[]sheetPage{{url: "/index.html", html: pageUsingTokens}}, "", nil)

	if len(res.WorkItems) != 0 {
		t.Fatalf("a healthy stylesheet must file nothing, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want exactly 1 retraction, got %d", len(res.Resolved))
	}
	if res.Resolved[0].ItemType != "stylesheet_gutted" {
		t.Fatalf("retraction item type = %q", res.Resolved[0].ItemType)
	}
	// The key must match what the finding branch files, or a repaired site
	// carries its item for ever.
	if res.Resolved[0].ItemKey != stylesheetGuttedItemKey() {
		t.Fatalf("retraction key %q does not match the filed key %q",
			res.Resolved[0].ItemKey, stylesheetGuttedItemKey())
	}
	if res.Resolved[0].AllOfType {
		t.Fatalf("retraction must be narrow, not AllOfType")
	}
	if res.Resolved[0].Reason == "" {
		t.Fatalf("a retraction with no stated reason is indistinguishable from a hand-close")
	}
}

// ── the definition sources: each one, proven load-bearing ───────────────────

// A component that defines its own token in an inline <style> is healthy, and a
// check that ignored in-page definitions would fire on it.
func TestStylesheetGutted_InPageDefinedPropertyIsNotAGap(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://example.uk/assets/css/styles.css"] = `:root { --color-text: #000; }`
	sf.install(t)

	html := `<html><head><link rel="stylesheet" href="/assets/css/styles.css"></head><body>
<style>.card { --card-ink: #333; color: var(--card-ink); }
body { color: var(--color-text); }</style></body></html>`

	res := runStylesheetGutted(t, "example.uk", []sheetPage{{url: "/index.html", html: html}}, "", nil)

	if len(res.WorkItems) != 0 {
		t.Fatalf("--card-ink is defined in the same <style> block; want no finding, got %d: %s",
			len(res.WorkItems), res.WorkItems[0].SpecJSON)
	}
}

// css_snippets are injected into the served head at ASSEMBLY time, so they are
// invisible in rendered_html. Omitting this source false-positives on every
// component that defines its tokens through a snippet.
func TestStylesheetGutted_SnippetDefinedPropertyIsNotAGap(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://example.uk/assets/css/styles.css"] = `:root { --color-text: #000; }`
	sf.install(t)

	html := `<html><head><link rel="stylesheet" href="/assets/css/styles.css"></head><body>
<style>.promo { background: var(--promo-bg); color: var(--color-text); }</style></body></html>`

	res := runStylesheetGutted(t, "example.uk",
		[]sheetPage{{url: "/index.html", html: html}}, "",
		[]string{`.promo-wrap { --promo-bg: #f0f0f0; }`})

	if len(res.WorkItems) != 0 {
		t.Fatalf("--promo-bg is defined by a css_snippet; want no finding, got %d: %s",
			len(res.WorkItems), res.WorkItems[0].SpecJSON)
	}
}

// ── the fallback rule ───────────────────────────────────────────────────────

// var(--x, #fff) keeps the declaration valid and is the defensive form the
// renderer itself emits. Counting it would fire on the recommended remedy.
func TestStylesheetGutted_FallbackCarryingReferenceIsNotAGap(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://example.uk/assets/css/styles.css"] = `:root { --color-text: #000; }`
	sf.install(t)

	html := `<html><head><link rel="stylesheet" href="/assets/css/styles.css"></head><body>
<style>.hero { background: var(--never-defined, #ffffff);
  color: var(--also-never, var(--color-text)); }</style></body></html>`

	res := runStylesheetGutted(t, "example.uk", []sheetPage{{url: "/index.html", html: html}}, "", nil)

	if len(res.WorkItems) != 0 {
		t.Fatalf("fallback-carrying references must not be gaps, got %d: %s",
			len(res.WorkItems), res.WorkItems[0].SpecJSON)
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("the site is healthy, so it should retract; got %d", len(res.Resolved))
	}
}

// ── blinded is not healthy ──────────────────────────────────────────────────

// A policy refusal (403 from a CDN) must file nothing AND retract nothing: the
// check cannot tell a missing definition from one it failed to read.
func TestStylesheetGutted_RefusedStylesheetFilesNothing(t *testing.T) {
	sf := newStubSheetFetch()
	sf.status["https://example.uk/assets/css/styles.css"] = 403
	sf.install(t)

	res := runStylesheetGutted(t, "example.uk",
		[]sheetPage{{url: "/index.html", html: pageUsingTokens}}, "", nil)

	if len(res.WorkItems) != 0 {
		t.Fatalf("a 403 blinds the check; want no finding, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Fatalf("a blinded run must never retract, got %d", len(res.Resolved))
	}
}

func TestStylesheetGutted_FetchFailureFilesNothing(t *testing.T) {
	sf := newStubSheetFetch()
	sf.err["https://example.uk/assets/css/styles.css"] = errors.New("dial tcp: timeout")
	sf.install(t)

	res := runStylesheetGutted(t, "example.uk",
		[]sheetPage{{url: "/index.html", html: pageUsingTokens}}, "", nil)

	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("a transport failure must file and retract nothing, got %d/%d",
			len(res.WorkItems), len(res.Resolved))
	}
}

// A 404 stylesheet is asset_reference_404's finding, not ours — and filing both
// would double-report one defect.
func TestStylesheetGutted_MissingStylesheetIsNotOurFinding(t *testing.T) {
	sf := newStubSheetFetch()
	sf.status["https://example.uk/assets/css/styles.css"] = 404
	sf.install(t)

	res := runStylesheetGutted(t, "example.uk",
		[]sheetPage{{url: "/index.html", html: pageUsingTokens}}, "", nil)

	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("a 404 belongs to asset_reference_404, got %d/%d",
			len(res.WorkItems), len(res.Resolved))
	}
}

// ── the DOM/comment rules ───────────────────────────────────────────────────

// A tool page's JavaScript may TALK about CSS. The html parser treats <script>
// content as raw text, so a mention is never reached — the same class of false
// positive asset_reference_404 records having measured.
func TestStylesheetGutted_VarInsideScriptTextIsNotAReference(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://example.uk/assets/css/styles.css"] = `:root { --color-text: #000; }`
	sf.install(t)

	html := `<html><head><link rel="stylesheet" href="/assets/css/styles.css"></head><body>
<style>body { color: var(--color-text); }</style>
<script>// we set el.style.color = "var(--totally-invented)" at runtime
const s = "var(--another-invention)";</script></body></html>`

	res := runStylesheetGutted(t, "example.uk", []sheetPage{{url: "/index.html", html: html}}, "", nil)

	if len(res.WorkItems) != 0 {
		t.Fatalf("a var() inside script text is not a reference, got %d: %s",
			len(res.WorkItems), res.WorkItems[0].SpecJSON)
	}
}

// A commented-out rule references nothing. The alias block this check most wants
// to read is surrounded by prose naming the very tokens it defines.
func TestStylesheetGutted_VarInsideCSSCommentIsNotAReference(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://example.uk/assets/css/styles.css"] = `:root { --color-text: #000; }`
	sf.install(t)

	html := `<html><head><link rel="stylesheet" href="/assets/css/styles.css"></head><body>
<style>/* opt in with var(--color-primary-ink) — never bare */
body { color: var(--color-text); }</style></body></html>`

	res := runStylesheetGutted(t, "example.uk", []sheetPage{{url: "/index.html", html: html}}, "", nil)

	if len(res.WorkItems) != 0 {
		t.Fatalf("a var() inside a CSS comment is not a reference, got %d: %s",
			len(res.WorkItems), res.WorkItems[0].SpecJSON)
	}
}

// ── scope and routing ───────────────────────────────────────────────────────

// Probing a third-party host to decide a local site is broken is a dependency
// this check does not take.
func TestStylesheetGutted_ExternalStylesheetIsNotFetched(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://example.uk/assets/css/styles.css"] = healthyStylesheet
	sf.install(t)

	html := `<html><head>
<link rel="stylesheet" href="/assets/css/styles.css">
<link rel="stylesheet" href="https://cdn.example.com/vendor.css">
</head><body><style>body { color: var(--color-text); }</style></body></html>`

	runStylesheetGutted(t, "example.uk", []sheetPage{{url: "/index.html", html: html}}, "", nil)

	if sf.called("https://cdn.example.com/vendor.css") {
		t.Fatalf("an external stylesheet must not be fetched")
	}
	if sf.totalCalls() != 1 {
		t.Fatalf("want exactly the one same-host fetch, got %d", sf.totalCalls())
	}
}

// Flag-only is a decision, not an omission: routing this at webdesign-agent
// would let an automatic "repair" re-roll the palette (check_generic_theme
// records four CSS rewrites in one day, one of which lightened a dark site).
func TestStylesheetGutted_NeverRoutesToAHandler(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://remortgagecalculator.uk/assets/css/styles.css"] = guttedStylesheet
	sf.install(t)

	res := runStylesheetGutted(t, "remortgagecalculator.uk",
		[]sheetPage{{url: "/index.html", html: pageUsingTokens}}, "", nil)

	for _, w := range res.WorkItems {
		if w.HandlerAgent != "" {
			t.Fatalf("stylesheet_gutted must be flag-only, got handler %q", w.HandlerAgent)
		}
		if w.Status != "detected" {
			t.Fatalf("want status detected, got %q", w.Status)
		}
	}
}

// asset_reference_404 RETRACTS on any 2xx for its URL-shaped keys. A key that
// collided with that namespace would have this check's findings closed by its
// sibling while the site is still broken.
func TestStylesheetGutted_ItemKeyIsConstantAndNotURLShaped(t *testing.T) {
	key := stylesheetGuttedItemKey()
	if strings.Contains(key, "http") || strings.Contains(key, "/assets/") {
		t.Fatalf("item key must not be URL-shaped (asset_reference_404 retracts those): %q", key)
	}
	if !strings.HasPrefix(key, "stylesheet_gutted:") {
		t.Fatalf("item key must be namespaced to this check: %q", key)
	}
	if key == assetRefItemKey("unresolvable_reference", "https://x/assets/css/styles.css") {
		t.Fatalf("item key collides with asset_reference_404")
	}
}

// ── population and blindness edges ──────────────────────────────────────────

// No stylesheet link at all means no definition source to read; the check must
// decline rather than declare every reference missing.
func TestStylesheetGutted_NoStylesheetLinkFilesNothing(t *testing.T) {
	sf := newStubSheetFetch()
	sf.install(t)

	html := `<html><body><style>body { color: var(--color-text); }</style></body></html>`
	res := runStylesheetGutted(t, "example.uk", []sheetPage{{url: "/index.html", html: html}}, "", nil)

	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("no stylesheet to read means no verdict, got %d/%d",
			len(res.WorkItems), len(res.Resolved))
	}
	if sf.totalCalls() != 0 {
		t.Fatalf("nothing should have been fetched, got %d calls", sf.totalCalls())
	}
}

// A site with no domain has no URL to resolve against.
func TestStylesheetGutted_NoDomainFilesNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow(""))

	res, err := (&StylesheetGuttedCheck{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(),
		Pipeline: "design", AgentType: "test", BatchID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("no domain means no verdict, got %d/%d", len(res.WorkItems), len(res.Resolved))
	}
}

// Chrome is where the stylesheet link usually lives, and one gap there is on
// every page of the site.
func TestStylesheetGutted_ChromeSuppliesTheStylesheetLink(t *testing.T) {
	sf := newStubSheetFetch()
	sf.body["https://example.uk/assets/css/styles.css"] = guttedStylesheet
	sf.install(t)

	chrome := `<head><link rel="stylesheet" href="/assets/css/styles.css"></head>`
	page := `<html><body><style>body { color: var(--color-text); }</style></body></html>`

	res := runStylesheetGutted(t, "example.uk", []sheetPage{{url: "/index.html", html: page}}, chrome, nil)

	if len(res.WorkItems) != 1 {
		t.Fatalf("the chrome link should have been resolved and fetched, got %d items", len(res.WorkItems))
	}
	if !sf.called("https://example.uk/assets/css/styles.css") {
		t.Fatalf("the chrome-supplied stylesheet was never fetched")
	}
}

// ── the check is registered under the name the config will use ──────────────

func TestStylesheetGutted_IsRegisteredUnderItsConfigName(t *testing.T) {
	c := Get("stylesheet_gutted")
	if c == nil {
		t.Fatalf("stylesheet_gutted is not in the registry; an unregistered configured name fails the whole discovery step")
	}
	if c.Name() != "stylesheet_gutted" {
		t.Fatalf("registered under %q", c.Name())
	}
}
