// FILE: platform/orchestration/actions/discovery_checks/check_site_structural_validity_test.go
//
// bugs_open/251-adjacent — pins the four checks in check_site_structural_validity.go.
//
// Like asset_reference_404 and site_unreachable, none of these four checks has
// a live positive to point at yet (they are not enabled on any discovery agent
// in this pass — see the file header). The substitute is the same: an induced
// fault per verdict branch, plus the probe discipline proven by breaking it and
// watching a NAMED test fail.
//
//	mutation                                          test that catches it
//	------------------------------------------------  ----------------------------------
//	hand-roll "https://"+domain+url again              TestPreferredStructuralURL_*
//	drop the confirming second probe                   TestProbeInternalLinkTargets_ConfirmsCandidate404
//	treat any non-2xx link status as a finding          TestProbeInternalLinkTargets_OnlyFourOhFourAndGoneFile
//	probe an empty/self href instead of skipping it     TestDeadInternalLinkLive_EmptyHrefNeverProbed
//	compare the canonical against ANY url, not THIS one TestJudgeCanonical_WrongTarget
//	skip the on-domain check                            TestJudgeCanonical_OffDomain
//	regex ld+json instead of parsing the DOM            TestExtractAndValidateLDJSON_ScriptMentionInsideCodeSample
//	give any item a handler_agent                       TestAllFourChecksNeverRouteToAHandler

package discovery_checks

import (
	"context"
	"strings"
	"sync"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// preferredStructuralURL — mirrors preferred_page_url_test.go's cases exactly,
// so the two files can only drift by someone deliberately breaking the mirror.
// ---------------------------------------------------------------------------

func TestPreferredStructuralURL_RootIsNormalised(t *testing.T) {
	got := preferredStructuralURL("example.com", "/index.html")
	if got != "https://example.com/" {
		t.Errorf("root: got %q, want https://example.com/", got)
	}
}

func TestPreferredStructuralURL_NonRootKeepsItsPath(t *testing.T) {
	for _, url := range []string{
		"/legal.html",
		"/guides/index.html", // section index: /guides/ 404s live — must NOT normalise
		"/loans/index.html",
		"/indexes.html", // suffix trap: contains "index.html" but is not one
	} {
		got := preferredStructuralURL("example.com", url)
		want := "https://example.com" + url
		if got != want {
			t.Errorf("non-root %q: got %q, want %q", url, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// judgeCanonical — pure, so every row is a table test (judgeSiteProbe's own
// discipline).
// ---------------------------------------------------------------------------

func TestJudgeCanonical_Missing(t *testing.T) {
	v := judgeCanonical("example.com", "https://example.com/", `<head><title>x</title></head>`, nil)
	if v.OK || v.Reason != "missing" {
		t.Errorf("got %+v, want Reason=missing", v)
	}
}

// injectCanonicalLink always emits an absolute https:// URL; a relative one is
// a distinct defect from off_domain (it names no host at all, not a WRONG one).
func TestJudgeCanonical_NotAbsolute(t *testing.T) {
	body := `<head><link rel="canonical" href="/index.html"></head>`
	v := judgeCanonical("example.com", "https://example.com/", body, nil)
	if v.OK || v.Reason != "not_absolute" {
		t.Errorf("got %+v, want Reason=not_absolute", v)
	}
}

func TestJudgeCanonical_OffDomain(t *testing.T) {
	body := `<head><link rel="canonical" href="https://not-this-site.com/"></head>`
	v := judgeCanonical("example.com", "https://example.com/", body, nil)
	if v.OK || v.Reason != "off_domain" {
		t.Errorf("got %+v, want Reason=off_domain", v)
	}
}

// The exact bugs_open/251 shape: the served canonical names /index.html, which
// is NOT this page's preferred URL ("/"). A check that compared against ANY
// on-domain URL rather than THIS page's expected one would pass this case —
// that is precisely what this test guards against.
func TestJudgeCanonical_WrongTarget(t *testing.T) {
	body := `<head><link rel="canonical" href="https://example.com/index.html"></head>`
	v := judgeCanonical("example.com", "https://example.com/", body, nil)
	if v.OK || v.Reason != "wrong_target" {
		t.Errorf("got %+v, want Reason=wrong_target", v)
	}
	if v.Actual != "https://example.com/index.html" {
		t.Errorf("Actual = %q", v.Actual)
	}
}

func TestJudgeCanonical_NotLive(t *testing.T) {
	body := `<head><link rel="canonical" href="https://example.com/"></head>`
	resolvesLive := func(href string) (bool, string) { return false, "canonical target returns HTTP 404" }
	v := judgeCanonical("example.com", "https://example.com/", body, resolvesLive)
	if v.OK || v.Reason != "not_live" {
		t.Errorf("got %+v, want Reason=not_live", v)
	}
}

func TestJudgeCanonical_OK(t *testing.T) {
	body := `<head><link rel="canonical" href="https://example.com/legal.html"></head>`
	resolvesLive := func(href string) (bool, string) { return true, "" }
	v := judgeCanonical("example.com", "https://example.com/legal.html", body, resolvesLive)
	if !v.OK {
		t.Errorf("got %+v, want OK", v)
	}
}

// www.<domain> is the same site (sameHost's own contract, reused here) — must
// not false-positive as off_domain.
func TestJudgeCanonical_WWWIsSameHost(t *testing.T) {
	body := `<head><link rel="canonical" href="https://www.example.com/"></head>`
	v := judgeCanonical("example.com", "https://www.example.com/", body, func(string) (bool, string) { return true, "" })
	if !v.OK {
		t.Errorf("got %+v, want OK (www is the same host)", v)
	}
}

// ---------------------------------------------------------------------------
// extractAndValidateLDJSON
// ---------------------------------------------------------------------------

func TestExtractAndValidateLDJSON_ValidAndInvalidBlocks(t *testing.T) {
	body := `<head>
		<script type="application/ld+json">{"@type":"WebPage","name":"ok"}</script>
		<script type="application/ld+json">{not valid json</script>
	</head>`
	blocks := extractAndValidateLDJSON(body)
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2", len(blocks))
	}
	if blocks[0].Err != "" {
		t.Errorf("block 0 should parse, got err %q", blocks[0].Err)
	}
	if blocks[1].Err == "" {
		t.Errorf("block 1 should NOT parse")
	}
}

func TestExtractAndValidateLDJSON_ZeroBlocksIsNotAnError(t *testing.T) {
	blocks := extractAndValidateLDJSON(`<head><title>x</title></head>`)
	if len(blocks) != 0 {
		t.Errorf("got %d blocks, want 0", len(blocks))
	}
}

// The class of false positive check_asset_reference_404.go's header records for
// <script src>: a MENTION of the tag inside another script's text content is not
// an occurrence of it. A regex over the raw HTML would match the ld+json-looking
// text inside this <script>-as-code-sample; goquery, parsing the DOM, must not.
func TestExtractAndValidateLDJSON_ScriptMentionInsideCodeSample(t *testing.T) {
	body := `<body><script>
		var example = '<script type="application/ld+json">{not json}</script>';
	</script></body>`
	blocks := extractAndValidateLDJSON(body)
	if len(blocks) != 0 {
		t.Errorf("got %d blocks from a MENTION inside a code sample, want 0: %+v", len(blocks), blocks)
	}
}

// ---------------------------------------------------------------------------
// headEssentials
// ---------------------------------------------------------------------------

func TestHeadEssentials_AllPresent(t *testing.T) {
	body := `<html><head><title>A real title</title></head>
		<body><a href="#content">Skip to content</a><footer>c 2026</footer></body></html>`
	title, skip, footer := headEssentials(body)
	if !title || !skip || !footer {
		t.Errorf("title=%v skip=%v footer=%v, want all true", title, skip, footer)
	}
}

func TestHeadEssentials_SkipLinkByClassAlone(t *testing.T) {
	body := `<html><head><title>t</title></head>
		<body><a href="#main" class="skip-link">Skip</a><footer>c</footer></body></html>`
	_, skip, _ := headEssentials(body)
	if !skip {
		t.Errorf("class=\"skip-link\" alone should satisfy the skip-link signal")
	}
}

func TestHeadEssentials_EmptyTitleCountsAsMissing(t *testing.T) {
	body := `<html><head><title></title></head><body><footer>c</footer></body></html>`
	title, _, _ := headEssentials(body)
	if title {
		t.Errorf("an empty <title></title> must count as missing, not present")
	}
}

func TestHeadEssentials_AllMissing(t *testing.T) {
	body := `<html><head></head><body>no essentials here</body></html>`
	title, skip, footer := headEssentials(body)
	if title || skip || footer {
		t.Errorf("title=%v skip=%v footer=%v, want all false", title, skip, footer)
	}
}

// ---------------------------------------------------------------------------
// probeInternalLinkTargets — the shared link-target prober used by
// dead_internal_link_live and canonical_mismatch's resolvesLive.
// ---------------------------------------------------------------------------

type stubLinkProbe struct {
	mu    sync.Mutex
	seq   map[string][]int
	err   map[string]error
	calls []string
}

func newStubLinkProbe() *stubLinkProbe {
	return &stubLinkProbe{seq: map[string][]int{}, err: map[string]error{}}
}

func (s *stubLinkProbe) install(t *testing.T) {
	t.Helper()
	prev := probeInternalLinkTarget
	probeInternalLinkTarget = func(_ context.Context, target string) (int, error) {
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
		return 200, nil
	}
	t.Cleanup(func() { probeInternalLinkTarget = prev })
}

func TestProbeInternalLinkTargets_ConfirmsCandidate404(t *testing.T) {
	s := newStubLinkProbe()
	s.seq["https://example.com/flaky.html"] = []int{404, 200} // confirm reverses it
	s.install(t)

	dctx := DiscoveryCheckContext{Ctx: context.Background(), Logger: zap.NewNop()}
	out := probeInternalLinkTargets(dctx, []string{"https://example.com/flaky.html"})

	if out["https://example.com/flaky.html"].code != 200 {
		t.Errorf("got %+v, want the CONFIRMED 200 to win, not the first 404", out)
	}
	if len(s.calls) != 2 {
		t.Errorf("got %d calls, want exactly 2 (probe + confirm)", len(s.calls))
	}
}

func TestProbeInternalLinkTargets_OnlyFourOhFourAndGoneFile(t *testing.T) {
	s := newStubLinkProbe()
	s.seq["https://example.com/a.html"] = []int{403}
	s.seq["https://example.com/b.html"] = []int{500}
	s.seq["https://example.com/c.html"] = []int{404, 404} // reproduces
	s.seq["https://example.com/d.html"] = []int{410, 410} // reproduces
	s.install(t)

	dctx := DiscoveryCheckContext{Ctx: context.Background(), Logger: zap.NewNop()}
	out := probeInternalLinkTargets(dctx, []string{
		"https://example.com/a.html", "https://example.com/b.html",
		"https://example.com/c.html", "https://example.com/d.html",
	})

	if out["https://example.com/a.html"].code != 403 || out["https://example.com/b.html"].code != 500 {
		t.Errorf("403/500 must pass through unfiltered for the caller's switch to skip them: %+v", out)
	}
	if out["https://example.com/c.html"].code != 404 || out["https://example.com/d.html"].code != 410 {
		t.Errorf("reproduced 404/410 must stand: %+v", out)
	}
}

// ---------------------------------------------------------------------------
// DeadInternalLinkLiveCheck.Run — DB + fetch wired end to end.
// ---------------------------------------------------------------------------

func newStructuralCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
		Pipeline:  "build",
		AgentType: "site-structural-validity-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

func expectStructuralDomain(mock sqlmock.Sqlmock, siteID uuid.UUID, domain string) {
	mock.ExpectQuery(`FROM sites WHERE id`).WithArgs(siteID).WillReturnRows(
		sqlmock.NewRows([]string{"domain"}).AddRow(domain))
}

func expectStructuralPages(mock sqlmock.Sqlmock, siteID uuid.UUID, pages ...structuralPage) {
	rows := sqlmock.NewRows([]string{"id", "name", "url"})
	for _, p := range pages {
		rows.AddRow(p.ID, p.Name, p.URL)
	}
	mock.ExpectQuery(`FROM pages p`).WithArgs(siteID).WillReturnRows(rows)
}

func installStructuralPageFetch(t *testing.T, byURL map[string]struct {
	status int
	body   string
	err    error
}) {
	t.Helper()
	prev := fetchStructuralPage
	prevWait := structuralRetryWait
	structuralRetryWait = 0
	fetchStructuralPage = func(_ context.Context, absoluteURL string) (int, string, error) {
		r, ok := byURL[absoluteURL]
		if !ok {
			return 0, "", context.DeadlineExceeded
		}
		return r.status, r.body, r.err
	}
	t.Cleanup(func() {
		fetchStructuralPage = prev
		structuralRetryWait = prevWait
	})
}

// The empty-href rule (c), end to end: a page whose only internal <a> is
// href="" must never reach the network prober at all.
func TestDeadInternalLinkLive_EmptyHrefNeverProbed(t *testing.T) {
	dctx, mock := newStructuralCtx(t)
	pageID := uuid.New()
	expectStructuralDomain(mock, dctx.SiteID, "example.com")
	expectStructuralPages(mock, dctx.SiteID, structuralPage{ID: pageID, Name: "index", URL: "/index.html"})

	installStructuralPageFetch(t, map[string]struct {
		status int
		body   string
		err    error
	}{
		"https://example.com/": {status: 200, body: `<a href="">nowhere</a><a href="#">also nowhere</a>`},
	})

	probe := newStubLinkProbe()
	probe.install(t)

	check := &DeadInternalLinkLiveCheck{}
	result, err := check.Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(probe.calls) != 0 {
		t.Errorf("empty/self href reached the network prober: %v", probe.calls)
	}
	if len(result.Findings) != 0 || len(result.WorkItems) != 0 {
		t.Errorf("got findings/work items for an empty href, want none: %+v / %+v", result.Findings, result.WorkItems)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// A confirmed dead internal link files, carries no handler_agent, and its
// re-observed-healthy sibling retracts.
func TestDeadInternalLinkLive_FindsConfirmedDeadLink_HandlerAgentEmpty(t *testing.T) {
	dctx, mock := newStructuralCtx(t)
	pageID := uuid.New()
	expectStructuralDomain(mock, dctx.SiteID, "example.com")
	expectStructuralPages(mock, dctx.SiteID, structuralPage{ID: pageID, Name: "index", URL: "/index.html"})

	installStructuralPageFetch(t, map[string]struct {
		status int
		body   string
		err    error
	}{
		"https://example.com/": {status: 200, body: `<a href="/dead.html">dead</a><a href="/alive.html">alive</a>`},
	})

	probe := newStubLinkProbe()
	probe.seq["https://example.com/dead.html"] = []int{404, 404}
	probe.seq["https://example.com/alive.html"] = []int{200}
	probe.install(t)

	check := &DeadInternalLinkLiveCheck{}
	result, err := check.Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.WorkItems) != 1 {
		t.Fatalf("got %d work items, want 1: %+v", len(result.WorkItems), result.WorkItems)
	}
	wi := result.WorkItems[0]
	if wi.ItemType != "dead_internal_link_live" {
		t.Errorf("ItemType = %q", wi.ItemType)
	}
	if wi.HandlerAgent != "" {
		t.Errorf("HandlerAgent = %q, want empty (flag-only)", wi.HandlerAgent)
	}
	if wi.ItemKey != "dead_internal_link_live:https://example.com/dead.html" {
		t.Errorf("ItemKey = %q", wi.ItemKey)
	}
	if len(result.Resolved) != 1 || result.Resolved[0].ItemKey != "dead_internal_link_live:https://example.com/alive.html" {
		t.Errorf("Resolved = %+v, want one entry for the healthy alive.html link", result.Resolved)
	}
}

// ---------------------------------------------------------------------------
// CanonicalMismatchCheck.Run — the exact bugs_open/251 shape, end to end.
// ---------------------------------------------------------------------------

func TestCanonicalMismatchCheck_251Shape_FilesWrongTarget(t *testing.T) {
	dctx, mock := newStructuralCtx(t)
	pageID := uuid.New()
	expectStructuralDomain(mock, dctx.SiteID, "example.com")
	expectStructuralPages(mock, dctx.SiteID, structuralPage{ID: pageID, Name: "index", URL: "/index.html"})

	installStructuralPageFetch(t, map[string]struct {
		status int
		body   string
		err    error
	}{
		// The exact bugs_open/251 defect: the served head names /index.html.
		"https://example.com/": {status: 200, body: `<head><link rel="canonical" href="https://example.com/index.html"></head>`},
	})

	check := &CanonicalMismatchCheck{}
	result, err := check.Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.WorkItems) != 1 {
		t.Fatalf("got %d work items, want 1: %+v", len(result.WorkItems), result.WorkItems)
	}
	wi := result.WorkItems[0]
	if wi.HandlerAgent != "" {
		t.Errorf("HandlerAgent = %q, want empty (flag-only)", wi.HandlerAgent)
	}
	if !strings.Contains(wi.Summary, "wrong_target") {
		t.Errorf("Summary = %q, want it to name wrong_target", wi.Summary)
	}
}

func TestCanonicalMismatchCheck_FixedHomepage_Retracts(t *testing.T) {
	dctx, mock := newStructuralCtx(t)
	pageID := uuid.New()
	expectStructuralDomain(mock, dctx.SiteID, "example.com")
	expectStructuralPages(mock, dctx.SiteID, structuralPage{ID: pageID, Name: "index", URL: "/index.html"})

	installStructuralPageFetch(t, map[string]struct {
		status int
		body   string
		err    error
	}{
		"https://example.com/": {status: 200, body: `<head><link rel="canonical" href="https://example.com/"></head>`},
	})
	probe := newStubLinkProbe() // resolvesLive's confirm path — 200 for everything
	probe.install(t)

	check := &CanonicalMismatchCheck{}
	result, err := check.Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.WorkItems) != 0 {
		t.Errorf("got %d work items for a CORRECT canonical, want 0: %+v", len(result.WorkItems), result.WorkItems)
	}
	if len(result.Resolved) != 1 {
		t.Fatalf("got %d Resolved entries, want 1", len(result.Resolved))
	}
	if result.Resolved[0].ItemKey != "canonical_mismatch:"+pageID.String() {
		t.Errorf("Resolved ItemKey = %q", result.Resolved[0].ItemKey)
	}
}

// ---------------------------------------------------------------------------
// HandlerAgent is empty on every one of the four checks — belt and braces
// alongside handler_coverage_test.go's source scan, at the point of
// construction rather than the point of source-grep.
// ---------------------------------------------------------------------------

func TestAllFourChecksNeverRouteToAHandler(t *testing.T) {
	dctx := DiscoveryCheckContext{SiteID: uuid.New(), AgentType: "t", BatchID: uuid.New(), Pipeline: "build"}
	page := structuralPage{ID: uuid.New(), Name: "x", URL: "/x.html"}

	items := []struct {
		name string
		wi   WorkItemSpec
	}{
		{"dead_internal_link_live", buildDeadLinkWorkItem(dctx, deadLinkTarget{URL: "https://e.com/x.html", PageID: page.ID, PageURL: page.URL}, 404)},
		{"canonical_mismatch", buildCanonicalWorkItem(dctx, page, canonicalVerdict{Reason: "missing"}, "https://e.com/x.html")},
		{"structured_data_invalid", buildStructuredDataWorkItem(dctx, page, []map[string]interface{}{{"block_index": 0}}, 1)},
		{"head_essentials_missing", buildHeadEssentialsWorkItem(dctx, page, []string{"title"}, false)},
	}
	for _, tc := range items {
		if tc.wi.HandlerAgent != "" {
			t.Errorf("%s: HandlerAgent = %q, want empty (flag-only this pass)", tc.name, tc.wi.HandlerAgent)
		}
		if tc.wi.Status != "detected" {
			t.Errorf("%s: Status = %q, want \"detected\"", tc.name, tc.wi.Status)
		}
	}
}

// ---------------------------------------------------------------------------
// Name() sanity — matches what verifier_coverage_test.go's source scan and
// discovery_checks_registration_test.go's registry lookup key on.
// ---------------------------------------------------------------------------

func TestFourChecksAreRegisteredUnderTheirDocumentedNames(t *testing.T) {
	for _, name := range []string{
		"dead_internal_link_live", "canonical_mismatch",
		"structured_data_invalid", "head_essentials_missing",
	} {
		if Get(name) == nil {
			t.Errorf("check %q is not registered", name)
		}
	}
}
