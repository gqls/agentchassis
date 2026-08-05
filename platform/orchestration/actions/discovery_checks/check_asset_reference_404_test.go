// FILE: platform/orchestration/actions/discovery_checks/check_asset_reference_404_test.go
//
// bugs_open/084 — pins asset_reference_404.
//
// THIS FILE CARRIES AN UNUSUAL BURDEN AND IT IS WORTH SAYING WHY. The check it
// tests has NO live positive: measured 2026-08-05 across all 541 deployed pages,
// 96 of 96 distinct referenced assets return 200. So nothing in production can
// demonstrate that the check bites, and "it found nothing" is exactly what a
// check that is silently broken also reports (016b §9: a gate's 0 findings has
// two causes with opposite fixes).
//
// The substitute is an induced fault per branch, plus a recorded mutation for
// each guard. Every guard below was proven load-bearing by breaking it and
// watching a NAMED test fail:
//
//	mutation                                          test that catches it
//	------------------------------------------------  ----------------------------------
//	delete the 404-confirmation second probe          CandidateNotFoundNotReproduced…
//	treat any non-2xx as a finding                    InconclusiveStatusesFileNothing
//	treat a transport error as a status               TransportErrorFilesNothing
//	probe an empty src instead of reporting it        EmptyReferenceIsNeverProbed
//	drop the URL from the ItemKey (basename only)     SameBasenameDifferentDirectories…
//	resolve relative refs against the site root       RelativeReferenceResolvesAgainstThePage
//	regex the HTML instead of parsing the DOM         ScriptTagInsideJSCommentIsNotAReference
//	resolve chrome page-relative refs anyway          ChromePageRelativeReferenceIsSkipped
//	give the item a handler_agent                     NeverRoutesToAHandler
//	widen the selector to <img> / rel=icon            ImgAndIconAreNotOurs
//
// A guard no test can be made to fail against is not verified, it is decoration.

package discovery_checks

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// stubProbe replaces the network with a table of statuses and records every call,
// so a test can assert what was NOT requested as well as what was.
type stubProbe struct {
	mu     sync.Mutex
	status map[string]int   // url -> status
	err    map[string]error // url -> transport failure
	// seq lets a url answer differently on successive calls, which is how the
	// 404-confirmation branch is exercised.
	seq   map[string][]int
	calls []string
}

func newStubProbe() *stubProbe {
	return &stubProbe{
		status: map[string]int{},
		err:    map[string]error{},
		seq:    map[string][]int{},
	}
}

func (s *stubProbe) install(t *testing.T) {
	t.Helper()
	prev := probeAssetURL
	probeAssetURL = func(_ context.Context, target string) (int, error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.calls = append(s.calls, target)
		if e, ok := s.err[target]; ok {
			return 0, e
		}
		if q, ok := s.seq[target]; ok && len(q) > 0 {
			v := q[0]
			s.seq[target] = q[1:]
			return v, nil
		}
		if c, ok := s.status[target]; ok {
			return c, nil
		}
		return 200, nil
	}
	t.Cleanup(func() { probeAssetURL = prev })
}

func (s *stubProbe) callCount(target string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, c := range s.calls {
		if c == target {
			n++
		}
	}
	return n
}

func (s *stubProbe) totalCalls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

type refPage struct {
	url  string
	html string
}

// runAssetRef drives the real check. Query order matters and mirrors Run:
// sites (domain), then page_components, then site_components.
func runAssetRef(t *testing.T, domain string, pages []refPage, chromeHTML string) *CheckResult {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow(domain))

	pageRows := sqlmock.NewRows([]string{"rendered_html", "url", "id"})
	for _, p := range pages {
		pageRows.AddRow(p.html, p.url, uuid.New())
	}
	// The expectation deliberately requires the SHARED liveness builders, not
	// just the table name. sqlmock matches the query by regexp, so if anyone
	// re-hand-rolls this predicate as `deployed_at IS NOT NULL` (the first
	// version of this check did exactly that, and the council's edit-quality seat
	// caught it) the mock stops matching and EVERY test in this file fails.
	//
	// That is the point: 28 pages shipped under a status other than 'deployed'
	// (bugs_open/185) and 35 of 46 needs_rebuild rows are still being served, so
	// a fresh spelling silently shrinks the probe list — and a check that omits
	// pages reports clean for the wrong reason. Pinning it here rather than in a
	// comment keeps the guard from depending on someone reading the comment.
	mock.ExpectQuery(regexp.QuoteMeta(datahelpers.PageHasShippedPredicateFor("p"))).
		WillReturnRows(pageRows)

	chromeRows := sqlmock.NewRows([]string{"rendered_html"})
	if chromeHTML != "" {
		chromeRows.AddRow(chromeHTML)
	}
	mock.ExpectQuery("FROM site_components").WillReturnRows(chromeRows)

	res, err := (&AssetReference404Check{}).Run(DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.New(),
		Pipeline:  "design",
		AgentType: "test",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("AssetReference404Check.Run: %v", err)
	}
	return res
}

func assetRefKeys(res *CheckResult) []string {
	out := make([]string, 0, len(res.WorkItems))
	for _, w := range res.WorkItems {
		out = append(out, w.ItemKey)
	}
	return out
}

// ── the finding branch ──────────────────────────────────────────────────────

// THE DEFECT THIS CHECK EXISTS FOR, induced. A deployed tool page loads a script
// that is not there. Every other signal says the page is fine.
func TestAssetReference404_ConfirmedNotFoundFilesExactlyOneItem(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://webdesign.co.uk/tools/head-architect/app.js"] = 404
	sp.install(t)

	res := runAssetRef(t, "webdesign.co.uk", []refPage{{
		url:  "/tools/head-architect/index.html",
		html: `<html><body><h1>Head Architect</h1><script src="app.js"></script></body></html>`,
	}}, "")

	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d: %v", len(res.WorkItems), assetRefKeys(res))
	}
	w := res.WorkItems[0]
	if w.ItemType != "asset_reference_404" {
		t.Errorf("item type = %q", w.ItemType)
	}
	if w.Severity != "high" {
		t.Errorf("severity = %q, want high", w.Severity)
	}
	if !strings.Contains(w.ItemKey, "https://webdesign.co.uk/tools/head-architect/app.js") {
		t.Errorf("item key must carry the resolved URL, got %q", w.ItemKey)
	}
	if !strings.Contains(w.Summary, "404") {
		t.Errorf("summary should name the status, got %q", w.Summary)
	}
	if w.PageID == nil {
		t.Error("a page-surface finding must carry the page id")
	}
	if len(res.Findings) != 1 {
		t.Errorf("want 1 finding, got %d", len(res.Findings))
	}
}

// A 410 is as definitive as a 404 — the origin is telling us the file is gone.
func TestAssetReference404_GoneIsAlsoAFinding(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://x.com/a.js"] = 410
	sp.install(t)

	res := runAssetRef(t, "x.com", []refPage{{
		url: "/index.html", html: `<script src="/a.js"></script>`,
	}}, "")
	if len(res.WorkItems) != 1 {
		t.Fatalf("410 should file, got %d items", len(res.WorkItems))
	}
}

// ── the skip taxonomy: the property that makes the Cloudflare landmine harmless ──

// MUTATION GUARD: change the `default:` arm to file a finding and this fails.
//
// A prior Python sweep of this exact population had every request refused by
// Cloudflare — 63 refusals on webdesign.co.uk, whose scripts all serve 200. A
// "non-200 is a finding" rule turns that into 63 false 404s on the very site this
// bug is about. The check may be BLINDED by a refusal; it must never be made to
// LIE by one.
func TestAssetReference404_InconclusiveStatusesFileNothing(t *testing.T) {
	for _, code := range []int{401, 403, 405, 429, 500, 502, 503} {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			sp := newStubProbe()
			sp.status["https://x.com/a.js"] = code
			sp.install(t)

			res := runAssetRef(t, "x.com", []refPage{{
				url: "/index.html", html: `<script src="/a.js"></script>`,
			}}, "")

			if len(res.WorkItems) != 0 {
				t.Errorf("status %d filed %d work item(s); only 404/410 may file",
					code, len(res.WorkItems))
			}
			if len(res.Resolved) != 0 {
				t.Errorf("status %d retracted %d item(s); an inconclusive status proves nothing",
					code, len(res.Resolved))
			}
		})
	}
}

// MUTATION GUARD: compare the returned code without checking err first and this
// fails — a transport failure returns 0, and 0 is not 404, but nor is it a status.
func TestAssetReference404_TransportErrorFilesNothing(t *testing.T) {
	sp := newStubProbe()
	sp.err["https://x.com/a.js"] = errors.New("dial tcp: i/o timeout")
	sp.install(t)

	res := runAssetRef(t, "x.com", []refPage{{
		url: "/index.html", html: `<script src="/a.js"></script>`,
	}}, "")

	if len(res.WorkItems) != 0 {
		t.Errorf("a transport error filed %d item(s)", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("a transport error retracted %d item(s)", len(res.Resolved))
	}
}

// MUTATION GUARD: delete the second probe in probeAll and this fails. An edge that
// answers 404 once and 200 on retry has told us the first answer was not about the
// file.
func TestAssetReference404_CandidateNotFoundNotReproducedIsDiscarded(t *testing.T) {
	sp := newStubProbe()
	sp.seq["https://x.com/a.js"] = []int{404, 200}
	sp.install(t)

	res := runAssetRef(t, "x.com", []refPage{{
		url: "/index.html", html: `<script src="/a.js"></script>`,
	}}, "")

	if len(res.WorkItems) != 0 {
		t.Fatalf("an unreproduced 404 filed %d item(s)", len(res.WorkItems))
	}
	if sp.callCount("https://x.com/a.js") != 2 {
		t.Errorf("a candidate 404 must be confirmed by a second request; saw %d call(s)",
			sp.callCount("https://x.com/a.js"))
	}
	// The second answer was 200, which is a positive observation, so it retracts.
	if len(res.Resolved) != 1 {
		t.Errorf("want 1 retraction from the confirming 200, got %d", len(res.Resolved))
	}
}

// A reproduced 404 costs exactly two calls and then files.
func TestAssetReference404_ReproducedNotFoundIsConfirmedThenFiled(t *testing.T) {
	sp := newStubProbe()
	sp.seq["https://x.com/a.js"] = []int{404, 404}
	sp.install(t)

	res := runAssetRef(t, "x.com", []refPage{{
		url: "/index.html", html: `<script src="/a.js"></script>`,
	}}, "")

	if len(res.WorkItems) != 1 {
		t.Fatalf("a reproduced 404 must file, got %d", len(res.WorkItems))
	}
	if got := sp.callCount("https://x.com/a.js"); got != 2 {
		t.Errorf("want 2 probes (probe + confirm), got %d", got)
	}
}

// ── retraction ──────────────────────────────────────────────────────────────

// A healthy reference retracts by ItemKey and NEVER with AllOfType. The wide
// branch would close findings this run never observed.
func TestAssetReference404_HealthyReferenceRetractsNarrowly(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://x.com/a.js"] = 200
	sp.install(t)

	res := runAssetRef(t, "x.com", []refPage{{
		url: "/index.html", html: `<script src="/a.js"></script>`,
	}}, "")

	if len(res.WorkItems) != 0 {
		t.Errorf("a 200 filed %d item(s)", len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want 1 retraction, got %d", len(res.Resolved))
	}
	r := res.Resolved[0]
	if r.AllOfType {
		t.Error("retraction must be narrow: AllOfType would close findings this run never looked at")
	}
	if r.ItemKey != assetRefItemKey("unresolvable_reference", "https://x.com/a.js") {
		t.Errorf("retraction key = %q", r.ItemKey)
	}
	if r.ItemType != "asset_reference_404" {
		t.Errorf("retraction type = %q", r.ItemType)
	}
	if r.Reason == "" {
		t.Error("a self-closing item with no stated cause is indistinguishable from a hand-closed one")
	}
}

// ── the empty reference: reported, never probed ──────────────────────────────

// MUTATION GUARD: send an empty src to the probe and this fails. Per the HTML spec
// an empty src resolves against the current document, so the probe scores a broken
// reference 200 (bugfix_128's landmine, written for <img> and true verbatim here).
func TestAssetReference404_EmptyReferenceIsNeverProbed(t *testing.T) {
	sp := newStubProbe()
	sp.install(t)

	res := runAssetRef(t, "x.com", []refPage{{
		url: "/index.html",
		html: `<script src=""></script><script src="  "></script>` +
			`<link rel="stylesheet" href="#">`,
	}}, "")

	if sp.totalCalls() != 0 {
		t.Errorf("an unresolvable reference must never be requested; %d call(s) made", sp.totalCalls())
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 empty_src item, got %d: %v", len(res.WorkItems), assetRefKeys(res))
	}
	w := res.WorkItems[0]
	if w.ItemKey != "asset_reference_404:empty_src" {
		t.Errorf("empty_src key = %q", w.ItemKey)
	}
	if w.Severity != "medium" {
		t.Errorf("empty_src severity = %q, want medium", w.Severity)
	}
	if !strings.Contains(w.SpecJSON, `"count":3`) {
		t.Errorf("all three empty references should be counted, spec = %s", w.SpecJSON)
	}
}

// ── URL resolution ──────────────────────────────────────────────────────────

// MUTATION GUARD: resolve against "https://"+domain+"/" instead of the page URL and
// this fails. 17 of webdesign.co.uk's references are page-relative, and a verdict
// from a URL you built is a verdict about a page you invented (TL-032).
func TestAssetReference404_RelativeReferenceResolvesAgainstThePage(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://webdesign.co.uk/tools/micro-cms/js/app.js"] = 404
	sp.status["https://webdesign.co.uk/js/app.js"] = 200 // the wrong answer, if resolved at root
	sp.install(t)

	res := runAssetRef(t, "webdesign.co.uk", []refPage{{
		url:  "/tools/micro-cms/index.html",
		html: `<script src="js/app.js"></script>`,
	}}, "")

	if len(res.WorkItems) != 1 {
		t.Fatalf("want the page-relative resolution to be probed and filed, got %d items", len(res.WorkItems))
	}
	if got := res.WorkItems[0].SpecJSON; !strings.Contains(got, "/tools/micro-cms/js/app.js") {
		t.Errorf("resolved against the wrong base: %s", got)
	}
}

// MUTATION GUARD: key on the basename and this fails. `app.js` under two tool
// directories are two files with two HTTP results; a basename key lets
// idx_swi_dedup silently drop the second — bugs_open/091's failure mode.
func TestAssetReference404_SameBasenameDifferentDirectoriesAreTwoItems(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://w.com/tools/a/app.js"] = 404
	sp.status["https://w.com/tools/b/app.js"] = 404
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{
		{url: "/tools/a/index.html", html: `<script src="app.js"></script>`},
		{url: "/tools/b/index.html", html: `<script src="app.js"></script>`},
	}, "")

	if len(res.WorkItems) != 2 {
		t.Fatalf("want 2 distinct items, got %d: %v", len(res.WorkItems), assetRefKeys(res))
	}
	keys := assetRefKeys(res)
	if keys[0] == keys[1] {
		t.Errorf("both findings share the dedup key %q — the second would be dropped", keys[0])
	}
}

// The same URL referenced from two pages is ONE probe and ONE item.
func TestAssetReference404_IdenticalURLIsProbedOnce(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://w.com/shared.js"] = 404
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{
		{url: "/a.html", html: `<script src="/shared.js"></script>`},
		{url: "/b.html", html: `<script src="/shared.js"></script>`},
	}, "")

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item for 1 URL, got %d", len(res.WorkItems))
	}
	// two calls = probe + confirmation, not two probes of the same reference
	if got := sp.callCount("https://w.com/shared.js"); got != 2 {
		t.Errorf("want the URL probed once and confirmed once, saw %d calls", got)
	}
}

// ── the DOM, not the HTML ───────────────────────────────────────────────────

// MUTATION GUARD: regex the raw HTML instead of parsing it and this fails.
//
// This is the regression test for a real misstep. The first measurement taken for
// this bug regexed `<script[^>]+src="([^"]+)"` over page_components and returned
// `<script src="...">` on a live tool page, which curl confirmed 404. It was a
// COMMENT inside that tool's own JavaScript describing a regex. Tool pages — this
// bug's whole population — are the pages most likely to talk about HTML inside JS.
func TestAssetReference404_ScriptTagInsideJSCommentIsNotAReference(t *testing.T) {
	sp := newStubProbe()
	sp.install(t)

	html := `<html><body><script>
	  // We want to keep anything that looks like <script src="..."> or <link rel="stylesheet">
	  const scripts = raw.match(/<script[\s\S]*?>/);
	</script></body></html>`

	res := runAssetRef(t, "webdesign.co.uk", []refPage{{
		url: "/tools/head-architect/index.html", html: html,
	}}, "")

	if sp.totalCalls() != 0 {
		t.Errorf("a mention of a script tag inside JS is not a reference; %d probe(s) made: %v",
			sp.totalCalls(), sp.calls)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("want no findings, got %v", assetRefKeys(res))
	}
}

// ── chrome ──────────────────────────────────────────────────────────────────

// Chrome is site-wide, so a rooted reference there has one answer for the whole
// site and is judged.
func TestAssetReference404_ChromeRootedReferenceIsJudged(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://w.com/assets/js/site.js"] = 404
	sp.install(t)

	res := runAssetRef(t, "w.com", nil, `<head><script src="/assets/js/site.js"></script></head>`)

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 chrome finding, got %d", len(res.WorkItems))
	}
	w := res.WorkItems[0]
	if !strings.Contains(w.SpecJSON, `"surface":"chrome"`) {
		t.Errorf("surface should be chrome: %s", w.SpecJSON)
	}
	if !strings.Contains(w.Summary, "every page") {
		t.Errorf("a chrome finding should say it is on every page: %q", w.Summary)
	}
	if w.PageID != nil {
		t.Error("chrome is not one page; PageID must be nil")
	}
}

// MUTATION GUARD: resolve a chrome page-relative reference against the site root
// and this fails. Shared chrome renders on every page, so `app.js` there resolves
// to a DIFFERENT URL on each one. There is no single verdict to give, so the check
// must decline rather than pick one.
func TestAssetReference404_ChromePageRelativeReferenceIsSkipped(t *testing.T) {
	sp := newStubProbe()
	sp.install(t)

	res := runAssetRef(t, "w.com", nil, `<head><script src="app.js"></script></head>`)

	if sp.totalCalls() != 0 {
		t.Errorf("a page-relative chrome reference has no single URL; %d probe(s) made: %v",
			sp.totalCalls(), sp.calls)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("want no findings, got %v", assetRefKeys(res))
	}
}

// ── boundaries with the checks that already own their asset class ───────────

// MUTATION GUARD: widen the selector and this fails. <img> is check_image_url_404's
// (answered from the DB with no HTTP at all) and rel="icon" is
// check_undeployed_assets' brand-head half. Four checks share the asset space and
// the landmine on check_image_url_404.go warns that widening one silently competes
// with another.
func TestAssetReference404_ImgAndIconAreNotOurs(t *testing.T) {
	sp := newStubProbe()
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{{
		url: "/index.html",
		html: `<img src="/assets/images/hero.jpg">` +
			`<link rel="icon" href="/favicon.ico">` +
			`<link rel="preload" href="/x.woff2" as="font">`,
	}}, "")

	if sp.totalCalls() != 0 {
		t.Errorf("only script[src] and rel~=stylesheet are ours; %d probe(s) made: %v",
			sp.totalCalls(), sp.calls)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("want no findings, got %v", assetRefKeys(res))
	}
}

// A multi-token rel is still a stylesheet. An == comparison misses these.
func TestAssetReference404_MultiTokenRelIsAStylesheet(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://w.com/a.css"] = 404
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{{
		url: "/index.html", html: `<link rel="alternate stylesheet" href="/a.css">`,
	}}, "")

	if len(res.WorkItems) != 1 {
		t.Fatalf("a token-matched rel should be judged, got %d items", len(res.WorkItems))
	}
	if !strings.Contains(res.WorkItems[0].SpecJSON, `"element":"stylesheet"`) {
		t.Errorf("element should be stylesheet: %s", res.WorkItems[0].SpecJSON)
	}
}

// A scheme we cannot GET is not a broken reference.
func TestAssetReference404_NonHTTPSchemesAreSkipped(t *testing.T) {
	sp := newStubProbe()
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{{
		url: "/index.html",
		html: `<script src="data:text/javascript,void%200"></script>` +
			`<script src="javascript:void(0)"></script>`,
	}}, "")

	if sp.totalCalls() != 0 {
		t.Errorf("non-http schemes must not be probed; %d call(s): %v", sp.totalCalls(), sp.calls)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("want no findings, got %v", assetRefKeys(res))
	}
}

// An external host is judged, and the finding says so, because the repair differs:
// pin a version rather than republish a file.
func TestAssetReference404_ExternalHostIsMarkedExternal(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://cdn.jsdelivr.net/npm/chart.js"] = 404
	sp.install(t)

	res := runAssetRef(t, "gamesdesign.co.uk", []refPage{{
		url:  "/index.html",
		html: `<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>`,
	}}, "")

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item, got %d", len(res.WorkItems))
	}
	if !strings.Contains(res.WorkItems[0].SpecJSON, `"external":true`) {
		t.Errorf("external host should be marked: %s", res.WorkItems[0].SpecJSON)
	}
}

// A same-origin reference is not external, including the www. form.
func TestAssetReference404_WWWFormIsSameOrigin(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://www.w.com/a.js"] = 404
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{{
		url: "/index.html", html: `<script src="https://www.w.com/a.js"></script>`,
	}}, "")

	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item, got %d", len(res.WorkItems))
	}
	if !strings.Contains(res.WorkItems[0].SpecJSON, `"external":false`) {
		t.Errorf("www. of the site's own domain is not external: %s", res.WorkItems[0].SpecJSON)
	}
}

// ── routing, and the cap ────────────────────────────────────────────────────

// MUTATION GUARD: give the item a handler and this fails. Repairing a 404ing
// reference means removing it, repointing it, or republishing the file it names —
// a judgement about intent that no generator can make. Same reasoning as
// check_image_url_404 and check_orphan_element_refs.
func TestAssetReference404_NeverRoutesToAHandler(t *testing.T) {
	sp := newStubProbe()
	sp.status["https://w.com/a.js"] = 404
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{{
		url: "/index.html", html: `<script src="/a.js"></script><script src=""></script>`,
	}}, "")

	if len(res.WorkItems) != 2 {
		t.Fatalf("want both kinds, got %d", len(res.WorkItems))
	}
	for _, w := range res.WorkItems {
		if w.HandlerAgent != "" {
			t.Errorf("%s routed to %q; this check is flag-only", w.ItemKey, w.HandlerAgent)
		}
		if w.Status != "detected" {
			t.Errorf("%s status = %q", w.ItemKey, w.Status)
		}
	}
}

// The cap bounds outbound calls, and what it drops it must be possible to see.
// A silent cap reads as "everything was checked" when it was not.
func TestAssetReference404_ProbeCapBoundsOutboundCalls(t *testing.T) {
	sp := newStubProbe()
	var b strings.Builder
	for i := 0; i < maxProbeURLs+5; i++ {
		u := fmt.Sprintf("https://w.com/s%03d.js", i)
		sp.status[u] = 404
		fmt.Fprintf(&b, `<script src="/s%03d.js"></script>`, i)
	}
	sp.install(t)

	res := runAssetRef(t, "w.com", []refPage{{url: "/index.html", html: b.String()}}, "")

	if len(res.WorkItems) > maxProbeURLs {
		t.Errorf("filed %d items past a cap of %d", len(res.WorkItems), maxProbeURLs)
	}
	// probe + confirmation for each of the capped set, and nothing beyond it.
	if got := sp.totalCalls(); got > maxProbeURLs*2 {
		t.Errorf("made %d calls; the cap should bound this at %d", got, maxProbeURLs*2)
	}
	// The dropped tail is the sorted remainder, so the last URL must be untouched.
	last := fmt.Sprintf("https://w.com/s%03d.js", maxProbeURLs+4)
	if sp.callCount(last) != 0 {
		t.Errorf("%s was probed despite being past the cap", last)
	}
}

// A site with no domain has nothing to resolve against and must say nothing
// rather than guess a host.
func TestAssetReference404_NoDomainIsSilent(t *testing.T) {
	sp := newStubProbe()
	sp.install(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow(""))

	res, err := (&AssetReference404Check{}).Run(DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: uuid.New(),
		Pipeline: "design", AgentType: "test", BatchID: uuid.New(), Logger: zap.NewNop(),
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 || sp.totalCalls() != 0 {
		t.Errorf("a domainless site must produce nothing; items=%d resolved=%d calls=%d",
			len(res.WorkItems), len(res.Resolved), sp.totalCalls())
	}
}

// The check registers itself under the name the workflow config uses. A name the
// binary does not register FAILS the discovery step, so this is the pin between
// the code and the SQL that enables it.
func TestAssetReference404_IsRegisteredUnderItsConfigName(t *testing.T) {
	if (&AssetReference404Check{}).Name() != "asset_reference_404" {
		t.Fatalf("name = %q", (&AssetReference404Check{}).Name())
	}
}
