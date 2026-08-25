// FILE: platform/orchestration/actions/discovery_checks/check_structure_floor_test.go
//
// Pins structure_floor, the STRUCTURE seat of the site acceptance council
// (RFC_056). The origin is an httptest TLS server standing in for
// https://<domain>, so URL resolution, the parked-domain control and the 2xx
// rule all run for real; only the http.Client is swapped. The DB is sqlmock, in
// this package's convention, with sites.settings handed in as JSON text.
//
// Every guard below was proven load-bearing by an induced fault with a NAMED
// test that catches it:
//
//	mutation                                             test that catches it
//	---------------------------------------------------  -------------------------------------
//	count month names inside <a> / inside <nav>          MonthNamesInNavLinksAreNotACalendar
//	match class by substring instead of whole token      BEMSubstringIsNotAClassToken
//	drop the heading half of the comparison rule         WideTableWithoutComparisonVocabulary…
//	skip the parked-domain control                       ParkedDomainBlindsTheSeat
//	file on zero fetched pages                           NoServedPageIsNoVerdict
//	count hidden inputs as tool fields                   ToolNeedsOperableFieldsAndAButton
//	give the item a handler / promote it                 BelowTheFloorFilesOneVerdictRow
//	ignore settings.structure_floor.n                    NOverriddenInSettings
//	file despite a recorded refusal                      RefusalIsRecordedNotFiled
//	fetch past the cap silently                          PageCapBoundsOutboundFetches

package discovery_checks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// structureTestSite is an httptest TLS origin. It serves the pages it is given,
// answers 404 to anything else (so the invented control path 404s, as a real
// site's would), or — when parked — 200s every unknown path the way a registrar
// parking page does. It records every path so a test can assert what was NOT
// requested.
type structureTestSite struct {
	srv   *httptest.Server
	mu    sync.Mutex
	paths []string
}

func structureNewTestSite(t *testing.T, pages map[string]string, parked bool) *structureTestSite {
	t.Helper()
	s := &structureTestSite{}
	s.srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.paths = append(s.paths, r.URL.Path)
		s.mu.Unlock()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if body, ok := pages[r.URL.Path]; ok {
			_, _ = io.WriteString(w, body)
			return
		}
		if parked {
			_, _ = io.WriteString(w, `<html><body><h1>This domain is for sale</h1></body></html>`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(s.srv.Close)

	prev := structureFloorHTTPClient
	structureFloorHTTPClient = s.srv.Client() // trusts the test certificate
	t.Cleanup(func() { structureFloorHTTPClient = prev })
	return s
}

// domain is what sites.domain would hold: host:port, so https://<domain>/…
// resolves to this server.
func (s *structureTestSite) domain() string { return s.srv.Listener.Addr().String() }

func (s *structureTestSite) pageRequests() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for _, p := range s.paths {
		if !strings.HasPrefix(p, "/__acceptance-control-") {
			out = append(out, p)
		}
	}
	return out
}

func (s *structureTestSite) controlRequests() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, p := range s.paths {
		if strings.HasPrefix(p, "/__acceptance-control-") {
			n++
		}
	}
	return n
}

// runStructureFloor drives the real check. Query order mirrors Run: sites
// (domain, settings), then pages. settings is JSON text or nil for NULL.
func runStructureFloor(t *testing.T, site *structureTestSite, settings interface{}, pageURLs []string) (*CheckResult, uuid.UUID, error) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain", "settings"}).AddRow(site.domain(), settings))

	pageRows := sqlmock.NewRows([]string{"id", "url"})
	for _, u := range pageURLs {
		pageRows.AddRow(uuid.New().String(), u)
	}
	mock.ExpectQuery("FROM pages").WillReturnRows(pageRows)

	siteID := uuid.New()
	res, err := (&StructureFloorCheck{}).Run(DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    siteID,
		Pipeline:  "design",
		AgentType: "test",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	})
	return res, siteID, err
}

func structurePageHTML(parts ...string) string {
	return `<html><head><title>t</title><style>.x{}</style></head><body>` +
		strings.Join(parts, "\n") + `</body></html>`
}

// structureDeliveredOf reads the delivered set from the run's one finding.
func structureDeliveredOf(t *testing.T, res *CheckResult) []string {
	t.Helper()
	if res == nil || len(res.Findings) != 1 {
		t.Fatalf("want exactly 1 finding, got %+v", res)
	}
	d, ok := res.Findings[0]["delivered"].([]string)
	if !ok {
		t.Fatalf("finding.delivered is %T, want []string", res.Findings[0]["delivered"])
	}
	return d
}

func structureHas(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// ── fixtures: one structure each, and nothing else ──────────────────────────

const (
	structureFixList = `<h2>Five tools</h2><ul>
		<li>Spade — for digging</li><li>Fork — for turning</li><li>Hoe — for weeding</li>
		<li>Rake — for levelling</li><li>Trowel — for planting</li></ul>`

	// 3 rows, 2 columns: a table, and NOT wide enough for a comparison.
	structureFixTable = `<table><tr><th>Plant</th><th>Bed</th></tr>
		<tr><td>Beans</td><td>North</td></tr><tr><td>Peas</td><td>South</td></tr></table>`

	structureFixMonthsText = `<p>Sow in January, February and March; plant out in April, May and June;
		harvest in July, August and September; clear in October, November and December.</p>`

	// Checkboxes outside any <form>, so they cannot double as a tool.
	structureFixCheckboxes = `<div class="prep"><label><input type="checkbox"> Clear beds</label>
		<label><input type="checkbox"> Order seed</label><label><input type="checkbox"> Sharpen tools</label></div>`

	// 3 rows, 3 columns: wide enough for a comparison, given the heading.
	structureFixWideTable = `<table><tr><th>Feature</th><th>Lender A</th><th>Lender B</th></tr>
		<tr><td>Rate</td><td>4.1%</td><td>4.3%</td></tr><tr><td>Fee</td><td>£999</td><td>£0</td></tr></table>`

	structureFixComparisonHeading = `<h2>Lender A vs Lender B</h2>`
	structureFixPlainHeading      = `<h2>Our lenders</h2>`

	structureFixDetails = `<details><summary>Q1</summary><p>A1</p></details>
		<details><summary>Q2</summary><p>A2</p></details><details><summary>Q3</summary><p>A3</p></details>`
)

// ── the verdict row ─────────────────────────────────────────────────────────

// THE DEFECT THIS SEAT EXISTS FOR, induced: a site that delivers two shapes of
// thing. MUTATION GUARD for routing: give the item a handler, or a status the
// promoter takes, and this fails.
func TestStructureFloor_BelowTheFloorFilesOneVerdictRow(t *testing.T) {
	site := structureNewTestSite(t, map[string]string{
		"/index.html":  structurePageHTML(structureFixList),
		"/plants.html": structurePageHTML(structureFixTable),
	}, false)

	res, siteID, err := runStructureFloor(t, site, nil, []string{"/index.html", "/plants.html"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if site.controlRequests() != 1 {
		t.Errorf("want exactly one parked-domain control probe, saw %d", site.controlRequests())
	}

	delivered := structureDeliveredOf(t, res)
	if strings.Join(delivered, ",") != "list,table" {
		t.Errorf("delivered = %v, want [list table]", delivered)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 verdict row, got %d", len(res.WorkItems))
	}
	if len(res.Resolved) != 0 {
		t.Errorf("below the floor must retract nothing, got %d", len(res.Resolved))
	}
	w := res.WorkItems[0]
	if w.ItemType != "structure_floor_unmet" {
		t.Errorf("item type = %q", w.ItemType)
	}
	if w.ItemKey != structureFloorItemKey(siteID) {
		t.Errorf("item key = %q", w.ItemKey)
	}
	if w.Severity != "medium" || w.Priority != 115 {
		t.Errorf("severity/priority = %q/%d, want medium/115", w.Severity, w.Priority)
	}
	if w.HandlerAgent != "" {
		t.Errorf("routed to %q; this seat is flag-only — a refusal is a verdict, not a fix", w.HandlerAgent)
	}
	if w.Status != "detected" {
		t.Errorf("status = %q, want detected", w.Status)
	}
	if !strings.HasPrefix(w.Summary, "2 of 6 reader-facing structures delivered across 2 pages: list, table") {
		t.Errorf("summary = %q", w.Summary)
	}
	for _, needle := range []string{
		`"count":2`, `"n":6`, `"delivered":["list","table"]`, `"seat":"structure"`, `"rfc":"RFC_056"`,
		`"refusal_path":"sites.settings->'maintenance_profile'->'structure_floor'->>'refusal'"`,
		`"not_dispatchable":"empty handler_agent`, `"pages_fetched":2`, `"pages_skipped":0`,
		`"evidence":{"list":"https://` + site.domain() + `/index.html","table":"https://` + site.domain() + `/plants.html"}`,
	} {
		if !strings.Contains(w.SpecJSON, needle) {
			t.Errorf("spec lacks %s:\n%s", needle, w.SpecJSON)
		}
	}
}

// Six distinct shapes across the site meet the owner's N, and the seat retracts
// its own row by key — narrowly, never AllOfType.
func TestStructureFloor_SixDeliveredMeetsTheFloorAndRetracts(t *testing.T) {
	site := structureNewTestSite(t, map[string]string{
		"/index.html":    structurePageHTML(structureFixList, structureFixTable),
		"/calendar.html": structurePageHTML(structureFixMonthsText, structureFixCheckboxes),
		"/compare.html":  structurePageHTML(structureFixComparisonHeading, structureFixWideTable),
		"/faq.html":      structurePageHTML(structureFixDetails),
	}, false)

	res, siteID, err := runStructureFloor(t, site, nil,
		[]string{"/index.html", "/calendar.html", "/compare.html", "/faq.html"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	delivered := structureDeliveredOf(t, res)
	if got := strings.Join(delivered, ","); got != "calendar,checklist,comparison,faq,list,table" {
		t.Errorf("delivered = %v", delivered)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("at the floor nothing may be filed, got %d item(s)", len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want 1 retraction, got %d", len(res.Resolved))
	}
	r := res.Resolved[0]
	if r.ItemType != "structure_floor_unmet" || r.ItemKey != structureFloorItemKey(siteID) || r.AllOfType {
		t.Errorf("retraction = %+v; want narrow, by this site's key", r)
	}
	if r.Reason != "6 of 6 delivered" {
		t.Errorf("reason = %q", r.Reason)
	}
}

// MUTATION GUARD: ignore settings->'maintenance_profile'->'structure_floor'->>'n'
// and this fails. Both spellings jsonb can hand back are honoured.
func TestStructureFloor_NOverriddenInSettings(t *testing.T) {
	for name, settings := range map[string]string{
		"number": `{"maintenance_profile":{"structure_floor":{"n":2}}}`,
		"string": `{"maintenance_profile":{"structure_floor":{"n":"2"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			site := structureNewTestSite(t, map[string]string{
				"/index.html": structurePageHTML(structureFixList, structureFixTable),
			}, false)
			res, _, err := runStructureFloor(t, site, settings, []string{"/index.html"})
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if len(res.WorkItems) != 0 {
				t.Errorf("N=2 met by 2, yet %d item(s) filed", len(res.WorkItems))
			}
			if len(res.Resolved) != 1 || res.Resolved[0].Reason != "2 of 2 delivered" {
				t.Errorf("resolved = %+v", res.Resolved)
			}
			if res.Findings[0]["n"] != 2 || res.Findings[0]["n_source"] != "settings" {
				t.Errorf("finding n/n_source = %v/%v", res.Findings[0]["n"], res.Findings[0]["n_source"])
			}
		})
	}

	// An unusable override falls back to the default and SAYS so.
	t.Run("invalid_falls_back", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(structureFixList, structureFixTable),
		}, false)
		res, _, err := runStructureFloor(t, site,
			`{"maintenance_profile":{"structure_floor":{"n":"lots"}}}`, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if res.Findings[0]["n"] != structureFloorDefault || res.Findings[0]["n_source"] != "invalid" {
			t.Errorf("finding n/n_source = %v/%v", res.Findings[0]["n"], res.Findings[0]["n_source"])
		}
		if len(res.WorkItems) != 1 {
			t.Errorf("default N=6 unmet by 2, want 1 item, got %d", len(res.WorkItems))
		}
	})
}

// MUTATION GUARD: file despite a recorded refusal and this fails. The refusal is
// the ruling's own escape hatch — on the record, with the count beside it, and
// it retracts an open row.
func TestStructureFloor_RefusalIsRecordedNotFiled(t *testing.T) {
	site := structureNewTestSite(t, map[string]string{
		"/index.html": structurePageHTML(structureFixList, structureFixTable),
	}, false)
	refusal := "brief is anti-commercial: directory and comparison ruled out; 2 structures is the site"
	res, siteID, err := runStructureFloor(t, site,
		`{"maintenance_profile":{"structure_floor":{"refusal":"`+refusal+`"}}}`, []string{"/index.html"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Errorf("a recorded refusal filed %d item(s)", len(res.WorkItems))
	}
	if len(res.Findings) != 1 || res.Findings[0]["refused"] != refusal || res.Findings[0]["count"] != 2 {
		t.Errorf("finding must carry the refusal text AND the count it refuses: %+v", res.Findings)
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want 1 retraction, got %d", len(res.Resolved))
	}
	r := res.Resolved[0]
	if r.ItemKey != structureFloorItemKey(siteID) || r.Reason != "refusal recorded: "+refusal {
		t.Errorf("retraction = %+v", r)
	}
}

// ── the parked-domain control ───────────────────────────────────────────────

// MUTATION GUARD: skip the invented-path control and this fails. A registrar
// parking page 200s EVERY path (LANDMINES), so every page would "serve" and the
// count would be a verdict about the parking page. Blinded means an ERROR —
// nothing filed, nothing retracted, and no page ever requested.
func TestStructureFloor_ParkedDomainBlindsTheSeat(t *testing.T) {
	site := structureNewTestSite(t, map[string]string{
		"/index.html": structurePageHTML(structureFixList, structureFixTable),
	}, true)
	res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
	if err == nil {
		t.Fatalf("a parked domain must be an error, got result %+v", res)
	}
	if !strings.Contains(err.Error(), "blinded") {
		t.Errorf("error should say the seat is blinded: %v", err)
	}
	if res != nil {
		t.Errorf("no verdict may be returned alongside the error, got %+v", res)
	}
	if site.controlRequests() != 1 {
		t.Errorf("want 1 control probe, saw %d", site.controlRequests())
	}
	if got := site.pageRequests(); len(got) != 0 {
		t.Errorf("no page may be fetched once blinded, saw %v", got)
	}
}

// Fewer than one page serving 2xx is not "zero structures" — it is no verdict.
// MUTATION GUARD: file on fetched == 0 and this fails.
func TestStructureFloor_NoServedPageIsNoVerdict(t *testing.T) {
	t.Run("every_page_404s", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/gone.html", "/also-gone.html"})
		if err == nil || res != nil {
			t.Fatalf("want an error and no result, got err=%v res=%+v", err, res)
		}
		if !strings.Contains(err.Error(), "cannot judge") {
			t.Errorf("error = %v", err)
		}
	})
	t.Run("no_active_pages_with_a_url", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{}, false)
		res, _, err := runStructureFloor(t, site, nil, nil)
		if err == nil || res != nil {
			t.Fatalf("want an error and no result, got err=%v res=%+v", err, res)
		}
	})
}

// ── the rubric's edges: named regressions ───────────────────────────────────

// MUTATION GUARD: count month names inside <a>, or stop stripping <nav>, and
// this fails. Twelve month names that are all links are twelve links. The same
// twelve names as plain text ARE a calendar — the positive control that keeps
// the negative from passing vacuously.
func TestStructureFloor_MonthNamesInNavLinksAreNotACalendar(t *testing.T) {
	months := []string{"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}
	var nav, bodyLinks, plain strings.Builder
	nav.WriteString(`<nav><ul>`)
	bodyLinks.WriteString(`<ul class="months">`)
	for _, m := range months {
		fmt.Fprintf(&nav, `<li><a href="/%s.html">%s</a></li>`, strings.ToLower(m), m)
		fmt.Fprintf(&bodyLinks, `<li><a href="/%s.html">%s</a></li>`, strings.ToLower(m), m)
	}
	nav.WriteString(`</ul></nav>`)
	bodyLinks.WriteString(`</ul>`)
	plain.WriteString(`<p>` + strings.Join(months, ", ") + `</p>`)

	t.Run("nav_links_only", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(nav.String(), structureFixList),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); structureHas(d, "calendar") {
			t.Errorf("twelve nav links counted as a calendar: %v", d)
		}
	})
	t.Run("body_links_only", func(t *testing.T) {
		// Not in <nav>, still inside <a>: the non-anchor rule alone must hold.
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(bodyLinks.String()),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		d := structureDeliveredOf(t, res)
		if structureHas(d, "calendar") {
			t.Errorf("twelve body links counted as a calendar: %v", d)
		}
		if structureHas(d, "list") {
			t.Errorf("a list of links has no non-anchor text and is not a list: %v", d)
		}
	})
	t.Run("plain_text_is_a_calendar", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(plain.String()),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); !structureHas(d, "calendar") {
			t.Errorf("twelve months in reader text must be a calendar: %v", d)
		}
	})
}

// MUTATION GUARD: match classes by substring and this fails. `card__title` is a
// BEM element, not a `card`; six of them pointing off-site are not a directory.
// Six real `card` siblings are — the positive control.
func TestStructureFloor_BEMSubstringIsNotAClassToken(t *testing.T) {
	cards := func(class string) string {
		var b strings.Builder
		b.WriteString(`<div class="grid">`)
		for i := 0; i < 6; i++ {
			fmt.Fprintf(&b, `<div class="%s"><a href="https://supplier-%d.example.org/">Supplier %d</a></div>`, class, i, i)
		}
		b.WriteString(`</div>`)
		return b.String()
	}
	t.Run("bem_element_only", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(cards("card__title")),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); structureHas(d, "directory") {
			t.Errorf("card__title counted as card: %v", d)
		}
	})
	t.Run("whole_token", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(cards("card card--wide")),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); !structureHas(d, "directory") {
			t.Errorf("six card siblings pointing off-site must be a directory: %v", d)
		}
	})
	t.Run("internal_links_are_not_a_directory", func(t *testing.T) {
		var b strings.Builder
		b.WriteString(`<div class="grid">`)
		for i := 0; i < 6; i++ {
			fmt.Fprintf(&b, `<div class="card"><a href="/guides/%d.html">Guide %d</a></div>`, i, i)
		}
		b.WriteString(`</div>`)
		site := structureNewTestSite(t, map[string]string{"/index.html": structurePageHTML(b.String())}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); structureHas(d, "directory") {
			t.Errorf("cards linking to our own pages are navigation, not a directory: %v", d)
		}
	})
}

// MUTATION GUARD: drop the heading half of the comparison rule and this fails.
// A wide table under a heading that does not say compare/vs is a table.
func TestStructureFloor_WideTableWithoutComparisonVocabularyIsATable(t *testing.T) {
	t.Run("plain_heading", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(structureFixPlainHeading, structureFixWideTable),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		d := structureDeliveredOf(t, res)
		if !structureHas(d, "table") || structureHas(d, "comparison") {
			t.Errorf("want [table] only, got %v", d)
		}
	})
	t.Run("vs_heading", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(structureFixComparisonHeading, structureFixWideTable),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); !structureHas(d, "comparison") {
			t.Errorf("wide table under 'A vs B' must be a comparison: %v", d)
		}
	})
}

// MUTATION GUARD: count hidden inputs as fields and this fails. A newsletter box
// with a CSRF token and a button is not a tool; a contact form with two operable
// fields and a button is.
func TestStructureFloor_ToolNeedsOperableFieldsAndAButton(t *testing.T) {
	t.Run("newsletter_box", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(`<form><input type="hidden" name="csrf" value="x">
				<input type="email" name="e"><button>Subscribe</button></form>`),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); structureHas(d, "tool") {
			t.Errorf("one operable field plus a hidden token is not a tool: %v", d)
		}
	})
	t.Run("two_fields_and_a_button", func(t *testing.T) {
		site := structureNewTestSite(t, map[string]string{
			"/index.html": structurePageHTML(`<form><input name="loan"><select name="term"><option>2</option></select>
				<button type="submit">Calculate</button></form>`),
		}, false)
		res, _, err := runStructureFloor(t, site, nil, []string{"/index.html"})
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if d := structureDeliveredOf(t, res); !structureHas(d, "tool") {
			t.Errorf("two operable fields and a button must be a tool: %v", d)
		}
	})
}

// The remaining rubric shapes, each on its own page, so a regression in any one
// rule is named rather than folded into a count.
func TestStructureFloor_FeedGuideAndClassTokenShapes(t *testing.T) {
	var guide strings.Builder
	for i := 0; i < 4; i++ {
		fmt.Fprintf(&guide, "<h2>Step %d</h2><p>%s</p>", i, strings.Repeat("soil light water ", 55))
	}
	feed := `<section class="news"><article><time datetime="2026-08-01">1 Aug</time> a</article>
		<article><time datetime="2026-08-02">2 Aug</time> b</article>
		<article><a href="https://www.rhs.org.uk/x">c</a></article></section>`
	faqList := `<div class="faq"><h3>Q1</h3><p>a</p><h3>Q2</h3><p>b</p><h3>Q3</h3><p>c</p></div>`
	checklistClass := `<div class="checklist"><ul><li>a</li><li>b</li><li>c</li></ul></div>`
	calendarClass := `<div class="period-cal"><p>this week</p></div>`

	site := structureNewTestSite(t, map[string]string{
		"/guide.html":    structurePageHTML(guide.String()),
		"/news.html":     structurePageHTML(feed),
		"/faq.html":      structurePageHTML(faqList),
		"/prep.html":     structurePageHTML(checklistClass),
		"/calendar.html": structurePageHTML(calendarClass),
	}, false)
	res, _, err := runStructureFloor(t, site, nil,
		[]string{"/guide.html", "/news.html", "/faq.html", "/prep.html", "/calendar.html"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	d := structureDeliveredOf(t, res)
	for _, want := range []string{"guide", "feed", "faq", "checklist", "calendar"} {
		if !structureHas(d, want) {
			t.Errorf("%s not detected; delivered = %v", want, d)
		}
	}
	// Evidence names the page each shape was FIRST seen on.
	ev, _ := res.Findings[0]["evidence"].(map[string]string)
	if !strings.HasSuffix(ev["feed"], "/news.html") || !strings.HasSuffix(ev["guide"], "/guide.html") {
		t.Errorf("evidence = %v", ev)
	}
}

// ── the cap ─────────────────────────────────────────────────────────────────

// The cap bounds outbound GETs; what it drops is logged, and the tail past it
// must never be requested.
func TestStructureFloor_PageCapBoundsOutboundFetches(t *testing.T) {
	pages := map[string]string{}
	var urls []string
	for i := 0; i < structureFloorPageCap+5; i++ {
		u := fmt.Sprintf("/p%03d.html", i)
		pages[u] = structurePageHTML(structureFixList)
		urls = append(urls, u)
	}
	site := structureNewTestSite(t, pages, false)
	res, _, err := runStructureFloor(t, site, nil, urls)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := len(site.pageRequests()); got != structureFloorPageCap {
		t.Errorf("fetched %d pages, cap is %d", got, structureFloorPageCap)
	}
	if res.Findings[0]["pages_over_cap"] != 5 {
		t.Errorf("finding must say how many pages were not looked at: %v", res.Findings[0]["pages_over_cap"])
	}
	last := fmt.Sprintf("/p%03d.html", structureFloorPageCap+4)
	for _, p := range site.pageRequests() {
		if p == last {
			t.Errorf("%s is past the cap and was requested", last)
		}
	}
}

// A name the binary does not register FAILS the discovery step: this is the pin
// between the code and the SQL that enables it.
func TestStructureFloor_IsRegisteredUnderItsConfigName(t *testing.T) {
	if (&StructureFloorCheck{}).Name() != "structure_floor" {
		t.Fatalf("name = %q", (&StructureFloorCheck{}).Name())
	}
	if Get("structure_floor") == nil {
		t.Fatal("structure_floor is not registered")
	}
}
