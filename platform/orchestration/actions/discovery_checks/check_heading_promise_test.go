// FILE: platform/orchestration/actions/discovery_checks/check_heading_promise_test.go
//
// RFC_056 promise seat — pins heading_promise.
//
// The check ports a bash harness that was wrong twice before it was right, so the
// two corrections are NAMED tests here, not comments:
//
//	mutation                                              test that catches it
//	----------------------------------------------------  ----------------------------------
//	count months in anchor text / menus                   MenuMonthsDoNotKeepTheCalendarPromise
//	strip EVERY <header>, not just the landmark one       SectionHeaderContentIsNotChrome
//	match class tokens by substring (card__title = card)  TopNCountsWholeClassTokensOnly
//	file or resolve on a parked domain                    ParkedDomainControlBlindsTheCheck
//	treat a non-2xx page as observed                      ServerErrorPageIsSkipped / TransportError…
//	turn an unreachable origin into an empty clean result  UnreachableOriginIsAnErrorNotACleanBill
//	retract on an unmet page, or widen to AllOfType        MenuMonths… / NoPromiseHeadingIsAPositive…
//	give the item a handler_agent                         NeverRoutesToAHandler
//	fetch past the page cap                               PageCapBoundsFetches
//
// Real HTTP throughout (httptest), with only the site's base URL swapped — so the
// fetch, redirect and timeout wiring is the code that runs in production.

package discovery_checks

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// promiseServed is one path's scripted answer.
type promiseServed struct {
	status int
	html   string
	hangup bool // hijack and close: the client sees a transport error, not a status
}

// promiseTestSite is an origin under test: scripted pages, a 404 for everything
// else (including the invented control path), or — when parked — 200 for every
// path, which is exactly what a registrar's catch-all does.
type promiseTestSite struct {
	t      *testing.T
	srv    *httptest.Server
	mu     sync.Mutex
	pages  map[string]promiseServed
	parked bool
	hits   []string
}

func newPromiseTestSite(t *testing.T) *promiseTestSite {
	t.Helper()
	s := &promiseTestSite{t: t, pages: map[string]promiseServed{}}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.hits = append(s.hits, r.URL.Path)
		parked := s.parked
		p, ok := s.pages[r.URL.Path]
		s.mu.Unlock()
		if parked {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><head><meta http-equiv="refresh" content="0;url=https://parking.example/"></head><body></body></html>`))
			return
		}
		if !ok {
			http.NotFound(w, r)
			return
		}
		if p.hangup {
			hj, can := w.(http.Hijacker)
			if !can {
				t.Fatal("test server cannot hijack")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("hijack: %v", err)
			}
			_ = conn.Close()
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(p.status)
		_, _ = w.Write([]byte(p.html))
	}))
	t.Cleanup(s.srv.Close)

	prev := promiseSiteBaseURL
	promiseSiteBaseURL = func(string) string { return s.srv.URL }
	t.Cleanup(func() { promiseSiteBaseURL = prev })
	return s
}

func (s *promiseTestSite) serve(path string, status int, html string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pages[path] = promiseServed{status: status, html: html}
}

func (s *promiseTestSite) hangup(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pages[path] = promiseServed{hangup: true}
}

func (s *promiseTestSite) hitCount(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, h := range s.hits {
		if h == path {
			n++
		}
	}
	return n
}

// pageHits counts requests that were NOT the control probe.
func (s *promiseTestSite) pageHits() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, h := range s.hits {
		if !strings.HasPrefix(h, "/__acceptance-control-") {
			n++
		}
	}
	return n
}

func (s *promiseTestSite) controlHits() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, h := range s.hits {
		if strings.HasPrefix(h, "/__acceptance-control-") {
			n++
		}
	}
	return n
}

// runHeadingPromise drives the real check against the test origin. Query order
// mirrors Run: sites (domain), then pages (id, url). Returns the page ids keyed by
// url so a test can assert on ItemKey / Resolved without guessing.
func runHeadingPromise(t *testing.T, urls ...string) (*CheckResult, error, map[string]uuid.UUID) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	siteID := uuid.New()
	mock.ExpectQuery("FROM sites").WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("example.test"))

	ids := map[string]uuid.UUID{}
	rows := sqlmock.NewRows([]string{"id", "url"})
	for _, u := range urls {
		id := uuid.New()
		ids[u] = id
		rows.AddRow(id, u)
	}
	mock.ExpectQuery("FROM pages").WithArgs(siteID).WillReturnRows(rows)

	res, runErr := (&HeadingPromiseCheck{}).Run(DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    siteID,
		Pipeline:  "content",
		AgentType: "acceptance-council-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	})
	return res, runErr, ids
}

func mustRunHeadingPromise(t *testing.T, urls ...string) (*CheckResult, map[string]uuid.UUID) {
	t.Helper()
	res, err, ids := runHeadingPromise(t, urls...)
	if err != nil {
		t.Fatalf("HeadingPromiseCheck.Run: %v", err)
	}
	return res, ids
}

// promiseMonthMenu is homegarden.uk's chrome, reduced: every month as a menu link.
// It is what made the raw count read 12 on /contact.html.
func promiseMonthMenu() string {
	var b strings.Builder
	b.WriteString(`<header class="site-header"><a href="/">homegarden.uk</a><nav class="main-nav"><ul>`)
	for _, m := range []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"} {
		fmt.Fprintf(&b, `<li><a href="/%s/index.html">%s</a></li>`, strings.ToLower(m), m)
	}
	b.WriteString(`</ul></nav></header>`)
	return b.String()
}

// promisePageWith wraps content in the full served shape: head with a title that
// mentions months (must not count), the month menu, <main>, and a footer of links.
func promisePageWith(main string) string {
	return `<!doctype html><html><head><title>Garden calendar: January to December</title>` +
		`<style>.card { display: block } /* January */</style>` +
		`<script>var months = ["January","February","March"]; // <li> checklist</script>` +
		`</head><body>` + promiseMonthMenu() +
		`<main>` + main + `</main>` +
		`<footer class="site-footer"><ul><li><a href="/about.html">About</a></li>` +
		`<li><a href="/contact.html">Contact</a></li></ul><h2>homegarden.uk</h2></footer>` +
		`</body></html>`
}

func promiseFindingFor(res *CheckResult, pageID uuid.UUID) map[string]interface{} {
	for _, f := range res.Findings {
		if f["page_id"] == pageID.String() {
			return f
		}
	}
	return nil
}

func promiseResolvedKeys(res *CheckResult) []string {
	out := make([]string, 0, len(res.Resolved))
	for _, r := range res.Resolved {
		out = append(out, r.ItemKey)
	}
	return out
}

// ── the calendar rule, and the regression the harness was corrected for ─────

func TestHeadingPromise_CalendarWithTwelveMonthsInContentIsMet(t *testing.T) {
	site := newPromiseTestSite(t)
	var b strings.Builder
	b.WriteString(`<h1>Garden maintenance for UK gardens, month by month</h1>`)
	for _, m := range []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"} {
		fmt.Fprintf(&b, `<section><h3>%s</h3><p>What to do in %s.</p></section>`, m, m)
	}
	site.serve("/garden/index.html", 200, promisePageWith(b.String()))

	res, ids := mustRunHeadingPromise(t, "/garden/index.html")

	if len(res.WorkItems) != 0 {
		t.Fatalf("a kept promise filed %d item(s): %+v", len(res.WorkItems), res.WorkItems)
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("a kept promise is a positive observation and must retract; got %v", promiseResolvedKeys(res))
	}
	r := res.Resolved[0]
	if r.ItemType != "heading_promise_unmet" || r.AllOfType || r.Reason == "" ||
		r.ItemKey != "heading_promise_unmet:"+ids["/garden/index.html"].String() {
		t.Fatalf("retraction must be narrow, typed and reasoned: %+v", r)
	}
	if f := promiseFindingFor(res, ids["/garden/index.html"]); f == nil || f["months"] != 12 {
		t.Fatalf("finding should measure 12 non-anchor months: %+v", f)
	}
}

// THE REGRESSION, named. Every page on homegarden.uk carries all twelve month
// names as menu links; the page the owner found served three in its content. A
// raw count reads 12 and passes it. MUTATION GUARD: stop removing <nav> or <a>
// before counting and this fails.
func TestHeadingPromise_MenuMonthsDoNotKeepTheCalendarPromise(t *testing.T) {
	site := newPromiseTestSite(t)
	site.serve("/garden/index.html", 200, promisePageWith(
		`<h1>Garden maintenance for UK gardens, month by month</h1>`+
			`<p>Lawns need their first cut in March. Deadhead through June. Lift dahlias in October.</p>`+
			`<p>See the <a href="/january/index.html">January</a> and <a href="/february/index.html">February</a> guides.</p>`))

	res, ids := mustRunHeadingPromise(t, "/garden/index.html")

	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 item for the unmet page, got %d", len(res.WorkItems))
	}
	w := res.WorkItems[0]
	if w.ItemType != "heading_promise_unmet" {
		t.Errorf("item type = %q", w.ItemType)
	}
	if w.ItemKey != "heading_promise_unmet:"+ids["/garden/index.html"].String() {
		t.Errorf("item key = %q", w.ItemKey)
	}
	if w.PageID == nil || *w.PageID != ids["/garden/index.html"] {
		t.Errorf("item must carry the page id, got %v", w.PageID)
	}
	if w.Severity != "medium" || w.Priority != 110 || w.Status != "detected" || w.Source != "discovery" {
		t.Errorf("item mis-shaped: severity=%q priority=%d status=%q source=%q", w.Severity, w.Priority, w.Status, w.Source)
	}
	if !strings.Contains(w.Summary, "'Garden maintenance for UK gardens, month by month'") ||
		!strings.Contains(w.Summary, "3 distinct month name(s)") {
		t.Errorf("summary must carry the heading verbatim and the non-anchor count: %q", w.Summary)
	}
	for _, want := range []string{`"months":3`, `"seat":"promise"`, `"rfc":"RFC_056"`,
		`"rule":"calendar"`, `"needed":6`, `"nominates_not_adjudicates"`, `"not_dispatchable"`} {
		if !strings.Contains(w.SpecJSON, want) {
			t.Errorf("spec lacks %s: %s", want, w.SpecJSON)
		}
	}
	if len(res.Resolved) != 0 {
		t.Errorf("an unmet page must not retract, got %v", promiseResolvedKeys(res))
	}
}

// ── checklist / comparison / top-N ──────────────────────────────────────────

func TestHeadingPromise_ChecklistNeedsThreeListItems(t *testing.T) {
	for _, tc := range []struct {
		name  string
		items int
		unmet bool
	}{
		{"two_items_unmet", 2, true},
		{"five_items_met", 5, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			site := newPromiseTestSite(t)
			var b strings.Builder
			b.WriteString(`<h2>Autumn lawn checklist</h2><ul>`)
			for i := 0; i < tc.items; i++ {
				fmt.Fprintf(&b, `<li>Job %d</li>`, i+1)
			}
			b.WriteString(`</ul>`)
			site.serve("/lawn.html", 200, promisePageWith(b.String()))

			res, _ := mustRunHeadingPromise(t, "/lawn.html")
			if got := len(res.WorkItems) == 1; got != tc.unmet {
				t.Fatalf("%d list items: unmet=%v, want %v (%+v)", tc.items, got, tc.unmet, res.WorkItems)
			}
			if tc.unmet {
				if !strings.Contains(res.WorkItems[0].Summary, fmt.Sprintf("%d non-anchor list item(s)", tc.items)) {
					t.Errorf("summary should carry the count: %q", res.WorkItems[0].Summary)
				}
				// The menu's twelve <li><a> must not have counted.
				if !strings.Contains(res.WorkItems[0].SpecJSON, fmt.Sprintf(`"list_items":%d`, tc.items)) {
					t.Errorf("menu list items leaked into the count: %s", res.WorkItems[0].SpecJSON)
				}
			} else if len(res.Resolved) != 1 {
				t.Errorf("met promise must retract, got %v", promiseResolvedKeys(res))
			}
		})
	}
}

func TestHeadingPromise_ComparisonNeedsATable(t *testing.T) {
	for _, tc := range []struct {
		name  string
		body  string
		unmet bool
	}{
		{"with_table_met", `<h2>Decking vs paving: a comparison</h2><table><tr><th>Decking</th><th>Paving</th></tr></table>`, false},
		{"without_table_unmet", `<h2>Decking vs paving: a comparison</h2><p>Decking is warmer underfoot; paving lasts longer.</p>`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			site := newPromiseTestSite(t)
			site.serve("/compare.html", 200, promisePageWith(tc.body))
			res, _ := mustRunHeadingPromise(t, "/compare.html")
			if got := len(res.WorkItems) == 1; got != tc.unmet {
				t.Fatalf("unmet=%v, want %v (%+v)", got, tc.unmet, res.WorkItems)
			}
			if tc.unmet && !strings.Contains(res.WorkItems[0].Summary, "0 table(s)") {
				t.Errorf("summary should say no table: %q", res.WorkItems[0].Summary)
			}
		})
	}
}

// MUTATION GUARD: match class tokens by substring and the BEM case passes — ten
// card__title elements are ten titles, not ten cards.
func TestHeadingPromise_TopNCountsWholeClassTokensOnly(t *testing.T) {
	cards := func(class string) string {
		var b strings.Builder
		b.WriteString(`<h1>Top 10 tools for small gardens</h1><div class="tool-grid">`)
		for i := 0; i < 10; i++ {
			fmt.Fprintf(&b, `<div class="%s"><h3>Tool %d</h3><p>Why it earns its place.</p></div>`, class, i+1)
		}
		b.WriteString(`</div>`)
		return b.String()
	}
	t.Run("ten_card_tokens_met", func(t *testing.T) {
		site := newPromiseTestSite(t)
		site.serve("/tools.html", 200, promisePageWith(cards("tool card featured")))
		res, _ := mustRunHeadingPromise(t, "/tools.html")
		if len(res.WorkItems) != 0 || len(res.Resolved) != 1 {
			t.Fatalf("ten whole-token cards keep a top-10 promise: items=%+v resolved=%v", res.WorkItems, promiseResolvedKeys(res))
		}
	})
	t.Run("bem_substrings_not_counted", func(t *testing.T) {
		site := newPromiseTestSite(t)
		site.serve("/tools.html", 200, promisePageWith(cards("card__title card-grid")))
		res, _ := mustRunHeadingPromise(t, "/tools.html")
		if len(res.WorkItems) != 1 {
			t.Fatalf("BEM substrings must not count as cards: %+v", res.WorkItems)
		}
		if !strings.Contains(res.WorkItems[0].SpecJSON, `"cards":0`) {
			t.Errorf("cards should measure 0: %s", res.WorkItems[0].SpecJSON)
		}
		if !strings.Contains(res.WorkItems[0].SpecJSON, `"rule":"top_n"`) || !strings.Contains(res.WorkItems[0].SpecJSON, `"needed":10`) {
			t.Errorf("rule should be top_n needing 10: %s", res.WorkItems[0].SpecJSON)
		}
	})
	t.Run("ten_list_items_also_met", func(t *testing.T) {
		site := newPromiseTestSite(t)
		var b strings.Builder
		b.WriteString(`<h1>Top 10 tools for small gardens</h1><ol>`)
		for i := 0; i < 10; i++ {
			fmt.Fprintf(&b, `<li>Tool %d</li>`, i+1)
		}
		b.WriteString(`</ol>`)
		site.serve("/tools.html", 200, promisePageWith(b.String()))
		res, _ := mustRunHeadingPromise(t, "/tools.html")
		if len(res.WorkItems) != 0 {
			t.Fatalf("ten list items keep a top-10 promise: %+v", res.WorkItems)
		}
	})
}

// ── chrome scoping: the landmark rule, both ways ─────────────────────────────

// MUTATION GUARD in both directions. Strip EVERY <header> and the six months in
// the section's own header vanish (0, unmet); strip NO header and the six in the
// site banner leak in (12). The right answer is exactly the six inside <main>.
// Measured 2026-08-25: homegarden.uk's / keeps its promise-bearing h2 inside
// <header class="period-cal__header"> within <main>.
func TestHeadingPromise_SectionHeaderContentIsNotChrome(t *testing.T) {
	site := newPromiseTestSite(t)
	page := `<!doctype html><html><body>` +
		`<header class="site-header"><p>Banner: July August September October November December</p>` + promiseMonthMenu() + `</header>` +
		`<main><section class="period-cal">` +
		`<header class="period-cal__header"><h2>The garden year, month by month</h2>` +
		`<p>January February March April May June</p></header>` +
		`<p>Read down from the current month.</p></section></main>` +
		`<footer class="site-footer"><p>January February March April May June July August</p></footer>` +
		`</body></html>`
	site.serve("/index.html", 200, page)

	res, ids := mustRunHeadingPromise(t, "/index.html")

	f := promiseFindingFor(res, ids["/index.html"])
	if f == nil || f["months"] != 6 {
		t.Fatalf("want exactly the 6 months inside the section header, got %+v", f)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 1 {
		t.Fatalf("six months meet the calendar rule: items=%+v resolved=%v", res.WorkItems, promiseResolvedKeys(res))
	}
}

// ── a page that no longer promises anything is a positive observation ────────

func TestHeadingPromise_NoPromiseHeadingIsAPositiveObservation(t *testing.T) {
	site := newPromiseTestSite(t)
	site.serve("/contact.html", 200, promisePageWith(`<h1>Get in touch</h1><h2>Email</h2><p>hello@example.test</p>`))

	res, ids := mustRunHeadingPromise(t, "/contact.html")

	if len(res.WorkItems) != 0 {
		t.Fatalf("no promise, no item: %+v", res.WorkItems)
	}
	if len(res.Resolved) != 1 || res.Resolved[0].ItemKey != "heading_promise_unmet:"+ids["/contact.html"].String() {
		t.Fatalf("a served page with no promise-bearing heading retracts its own key: %v", promiseResolvedKeys(res))
	}
	if res.Resolved[0].AllOfType {
		t.Error("retraction must never widen to AllOfType: this run observed one page")
	}
	if f := promiseFindingFor(res, ids["/contact.html"]); f == nil || f["months"] != 0 {
		t.Errorf("the menu's twelve months must read 0 on /contact.html (the harness's own regression): %+v", f)
	}
}

// ── the parked-domain control ────────────────────────────────────────────────

// MUTATION GUARD: skip the control and this fails — a catch-all serves the page
// 200 with no headings, which reads as "no promise, resolved". On a parked domain
// the check must refuse to judge: error, nothing filed, nothing resolved, and no
// page even fetched.
func TestHeadingPromise_ParkedDomainControlBlindsTheCheck(t *testing.T) {
	site := newPromiseTestSite(t)
	site.parked = true
	site.serve("/garden/index.html", 200, promisePageWith(`<h1>Garden maintenance, month by month</h1>`))

	res, err, _ := runHeadingPromise(t, "/garden/index.html")

	if err == nil {
		t.Fatal("a parked domain must be an error, not a verdict")
	}
	if !strings.Contains(err.Error(), "blinded") {
		t.Errorf("error should say the check was blinded: %v", err)
	}
	if res != nil && (len(res.WorkItems) != 0 || len(res.Resolved) != 0) {
		t.Errorf("parked domain filed/resolved: %+v", res)
	}
	if site.controlHits() != 1 {
		t.Errorf("want exactly one control probe, got %d", site.controlHits())
	}
	if site.pageHits() != 0 {
		t.Errorf("no page may be fetched once the control has failed; %d fetched", site.pageHits())
	}
}

// The control is an INVENTED path: it must never collide with a real page, and a
// healthy origin answers it 404, after which pages are read normally.
func TestHeadingPromise_ControlPathIsInventedAndRandom(t *testing.T) {
	a, err := promiseControlPath()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := promiseControlPath()
	if !regexp.MustCompile(`^__acceptance-control-[0-9a-f]{8}\.html$`).MatchString(a) {
		t.Errorf("control path shape: %q", a)
	}
	if a == b {
		t.Errorf("two control paths must differ: %q", a)
	}
}

// ── pages that could not be observed are skipped, never judged ───────────────

func TestHeadingPromise_ServerErrorPageIsSkipped(t *testing.T) {
	site := newPromiseTestSite(t)
	site.serve("/broken.html", 500, `<h1>Garden maintenance, month by month</h1>`)
	site.serve("/fine.html", 200, promisePageWith(`<h1>Get in touch</h1>`))

	res, ids := mustRunHeadingPromise(t, "/broken.html", "/fine.html")

	if len(res.WorkItems) != 0 {
		t.Fatalf("a 500 is not an observation of the page; filed %+v", res.WorkItems)
	}
	if keys := promiseResolvedKeys(res); len(keys) != 1 || keys[0] != "heading_promise_unmet:"+ids["/fine.html"].String() {
		t.Fatalf("only the served page may retract, got %v", keys)
	}
	if f := promiseFindingFor(res, ids["/broken.html"]); f == nil || f["skipped"] != "http_500" {
		t.Errorf("the skip must be visible in findings with its reason: %+v", f)
	}
}

func TestHeadingPromise_TransportErrorPageIsSkipped(t *testing.T) {
	site := newPromiseTestSite(t)
	site.hangup("/dropped.html")

	res, ids := mustRunHeadingPromise(t, "/dropped.html")

	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("a transport error is neither filed nor resolved: items=%+v resolved=%v", res.WorkItems, promiseResolvedKeys(res))
	}
	f := promiseFindingFor(res, ids["/dropped.html"])
	if f == nil || !strings.HasPrefix(fmt.Sprint(f["skipped"]), "transport_error") {
		t.Errorf("skip reason should name the transport error: %+v", f)
	}
}

// An origin that cannot even answer the control is not a clean site; it is a
// check that could not look. Error, not an empty result.
func TestHeadingPromise_UnreachableOriginIsAnErrorNotACleanBill(t *testing.T) {
	site := newPromiseTestSite(t)
	site.serve("/index.html", 200, promisePageWith(`<h1>Get in touch</h1>`))
	site.srv.Close()

	res, err, _ := runHeadingPromise(t, "/index.html")
	if err == nil {
		t.Fatalf("closed origin must error; got result %+v", res)
	}
	if !strings.Contains(err.Error(), "blinded") {
		t.Errorf("error should say the check was blinded: %v", err)
	}
}

// ── routing and the cap ──────────────────────────────────────────────────────

// MUTATION GUARD: give the item a handler and this fails. The repair — write the
// calendar, build the table, or cut the heading — is a planner/writer judgement.
func TestHeadingPromise_NeverRoutesToAHandler(t *testing.T) {
	site := newPromiseTestSite(t)
	site.serve("/lawn.html", 200, promisePageWith(`<h2>Autumn lawn checklist</h2><p>Scarify, then feed.</p>`))

	res, _ := mustRunHeadingPromise(t, "/lawn.html")
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item, got %d", len(res.WorkItems))
	}
	w := res.WorkItems[0]
	if w.HandlerAgent != "" {
		t.Errorf("flag-only: routed to %q", w.HandlerAgent)
	}
	if w.Status != "detected" {
		t.Errorf("status = %q, want detected", w.Status)
	}
}

// One item per page, however many headings fail on it; the spec lists them all.
func TestHeadingPromise_OneItemPerPageListsEveryUnmetHeading(t *testing.T) {
	site := newPromiseTestSite(t)
	site.serve("/guide.html", 200, promisePageWith(
		`<h1>The garden year, month by month</h1><p>March and October matter most.</p>`+
			`<h2>Autumn lawn checklist</h2><p>Scarify, then feed.</p>`+
			`<h2>Get in touch</h2>`))

	res, _ := mustRunHeadingPromise(t, "/guide.html")
	if len(res.WorkItems) != 1 {
		t.Fatalf("want ONE item for the page, got %d", len(res.WorkItems))
	}
	spec := res.WorkItems[0].SpecJSON
	if strings.Count(spec, `"rule_text"`) != 3 { // 1 top-level + 2 in unmet_headings
		t.Errorf("spec should list both unmet headings: %s", spec)
	}
	if !strings.Contains(res.WorkItems[0].Summary, "'The garden year, month by month'") {
		t.Errorf("summary names the first unmet heading: %q", res.WorkItems[0].Summary)
	}
}

func TestHeadingPromise_PageCapBoundsFetches(t *testing.T) {
	site := newPromiseTestSite(t)
	urls := make([]string, 0, promiseMaxPages+5)
	for i := 0; i < promiseMaxPages+5; i++ {
		u := fmt.Sprintf("/p%03d.html", i)
		urls = append(urls, u)
		site.serve(u, 200, promisePageWith(`<h1>Get in touch</h1>`))
	}

	res, _ := mustRunHeadingPromise(t, urls...)

	if got := site.pageHits(); got != promiseMaxPages {
		t.Errorf("fetched %d pages; the cap is %d", got, promiseMaxPages)
	}
	if site.hitCount(urls[len(urls)-1]) != 0 {
		t.Errorf("%s is past the cap and must not be fetched", urls[len(urls)-1])
	}
	if len(res.Resolved) != promiseMaxPages {
		t.Errorf("only the pages actually read may retract: %d", len(res.Resolved))
	}
}

// ── the rule table: nomination, not adjudication ─────────────────────────────

func TestHeadingPromise_RuleTable(t *testing.T) {
	cases := []struct {
		heading string
		rule    string // "" for no rule
		needed  int
	}{
		{"Garden maintenance for UK gardens, month by month", "calendar", 6},
		{"The garden and home year, month-by-month", "calendar", 6},
		// NOMINATED, not adjudicated — both live on homegarden.uk 2026-08-25 and
		// neither promises the structure on that page. The check flags them; a
		// reader decides. That is CORRECTION 2, and it is deliberate.
		{"Why one calendar does not fit the whole country", "calendar", 6},
		{"What each comparison covers", "comparison", 1},
		{"Autumn lawn checklist", "checklist", 3},
		{"Overwintering, step by step", "checklist", 3},
		{"10 steps to a perfect lawn", "checklist", 3}, // first match wins: not top_n
		{"Decking vs paving", "comparison", 1},
		{"Fence panels versus gravel boards", "comparison", 1},
		{"Compare the two side by side", "comparison", 1},
		{"Top 10 tools for small gardens", "top_n", 10},
		{"The 7 mistakes new gardeners make", "top_n", 7},
		{"Best 5 mowers under £300", "top_n", 5},
		{"5 ways to keep slugs off hostas", "top_n", 5},
		{"12 ideas for a shady corner", "top_n", 12},
		{"Top 100 plants", "", 0},               // N > 50
		{"The 2024 guide to composting", "", 0}, // N > 50, and not "202"
		{"The 2 jobs people confuse", "", 0},    // N < 3
		{"Compost", "", 0},                      // "comp" is not "compare"
		{"Get in touch", "", 0},
		{"Versatile shrubs", "", 0}, // "versatile" is not " versus "
	}
	for _, tc := range cases {
		t.Run(tc.heading, func(t *testing.T) {
			r, ok := promiseRuleFor(strings.ToLower(tc.heading))
			if ok != (tc.rule != "") || (ok && (r.Name != tc.rule || r.Needed != tc.needed)) {
				t.Fatalf("got ok=%v rule=%q needed=%d, want rule=%q needed=%d", ok, r.Name, r.Needed, tc.rule, tc.needed)
			}
		})
	}
}

// Headings are read from the served document minus nav/style/script; the
// harness's 3–90 character bounds apply.
func TestHeadingPromise_HeadingsExcludeNavAndRespectBounds(t *testing.T) {
	headings, _, err := promiseReadPage([]byte(`<html><body>` +
		`<nav><h2>Browse month by month</h2></nav>` +
		`<main><h1>OK</h1><h2>Autumn lawn checklist</h2>` +
		`<h3>` + strings.Repeat("x", 91) + ` checklist</h3></main></body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(headings) != 1 || headings[0].Text != "Autumn lawn checklist" {
		t.Fatalf("want only the in-bounds content heading, got %+v", headings)
	}
}

func TestHeadingPromise_IsRegisteredUnderItsConfigName(t *testing.T) {
	if (&HeadingPromiseCheck{}).Name() != "heading_promise" {
		t.Fatalf("name = %q", (&HeadingPromiseCheck{}).Name())
	}
	if Get("heading_promise") == nil {
		t.Fatal("heading_promise is not registered; a checks-array entry naming it would hard-fail the runner")
	}
}
