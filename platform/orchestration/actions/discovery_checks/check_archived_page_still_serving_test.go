// FILE: platform/orchestration/actions/discovery_checks/check_archived_page_still_serving_test.go
//
// bugs_open/359 — pins archived_page_still_serving.
//
// THERE IS NO STABLE LIVE POSITIVE, and that is measured rather than assumed:
// the population MOVED between 2026-08-22 and 2026-08-26 (both loancalculator.co.uk
// pages the bug filed as serving now 404; five of the seven found on the 26th were
// not in the 22nd's sample). So production can neither demonstrate that this check
// bites nor be relied on to keep demonstrating it, and "0 findings" is also what a
// silently broken check reports. Every branch is therefore proven by an induced
// fault, and every guard by breaking it and watching a NAMED test fail:
//
//	mutation                                             test that catches it
//	---------------------------------------------------  ------------------------------------
//	move the resolve pass below the no-archived return    UnarchivedPageStillResolvesWithNoProbe
//	probe pages before the controls                       ControlsRunBeforeAnyPageIsJudged
//	treat a 200 on the invented control as a pass         PermissiveRouterRefusesToJudge
//	treat a dead sibling control as a pass                DeadOriginRefusesAndCannotRetract
//	drop the confirming probe before filing               Serving200IsConfirmedBeforeFiling
//	drop the confirming probe before resolving            Absence404IsConfirmedBeforeResolving
//	file on a 2xx that landed somewhere else              RedirectedRetirementResolvesAndFilesNothing
//	drop the active-page file-path collision guard        ActiveCollisionPathIsNeverProbed
//	sanitise a fragment url instead of declining it       FragmentURLIsNeverProbed
//	treat 403/500 as an absence (i.e. resolve on it)      InconclusiveStatusesFileNothingAndResolveNothing
//	give the item a handler_agent                         ArchivedItemNeverRoutesToAHandler
//	skip the pool-site status gate                        PoolSiteIsFullNoOp
//	let the const and the filed literal drift             ArchivedItemTypeConstMatchesTheLiteral
//
// The verdict table (judgeArchivedProbe) is pure and covered row by row.
//
// ⚠ TWO OF THESE ROWS WERE FALSE WHEN FIRST WRITTEN, and the correction is worth
// more than the table. On the first mutation run, M2 (treat a 200 on the invented
// control as a pass) and M3 (treat a dead sibling as a pass) BOTH SURVIVED — the
// named tests kept passing with the guard deleted. Neither test was wrong about
// what it asserted; both were passing on a DIFFERENT failure. Their fixtures did
// not script the machinery downstream of the control, so removing the control just
// tripped the next thing in series (an unscripted probe, an unanswered query) and
// Run returned an error anyway. The assertions saw an error and were content.
//
// That is the recorded trap "a mutation that PASSES usually hit a guard in
// SERIES", and the fix is two-part and is in both tests now: script everything
// downstream HEALTHY, so that a check ignoring the control would run to completion
// and do the damage; and assert on WHICH control refused, not merely that
// something did. A refusal test that does not name its refusal cannot tell its own
// guard from the next one along.
//
// The two load-bearing tests are DisconfirmingPair — the bug's §7 acceptance
// criterion, which a detector that flags every archived page fails on its second
// arm — and DeadOriginRefusesAndCannotRetract, which is the property that makes
// this check's zero readable at all.

package discovery_checks

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// --- fixtures -------------------------------------------------------------

const archivedTestDomain = "robot-hands.test"

var (
	pgServing  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	pgAbsent   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	pgActive   = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	pgIndex    = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	pgFragment = uuid.MustParse("55555555-5555-5555-5555-555555555555")
)

// stubArchivedProbe answers from a queue keyed on the url with its cache-buster
// stripped, and records every call in order — so a test can assert not only what
// was said but WHEN, which is how the control-ordering property is pinned.
//
// The invented-URL control carries a fresh uuid per run, so it is recognised by
// its path prefix rather than by an exact url. That is deliberate: a stub that
// needed the exact control url could only be written by copying the generator,
// and would then pass even if the generator stopped producing a unique path.
type stubArchivedProbe struct {
	mu       sync.Mutex
	seq      map[string][]archivedProbeResult
	invented []archivedProbeResult
	calls    []string
}

func stripBuster(target string) string {
	u, err := url.Parse(target)
	if err != nil {
		return target
	}
	u.RawQuery = ""
	return u.String()
}

func (s *stubArchivedProbe) probe(_ context.Context, target string) archivedProbeResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	clean := stripBuster(target)
	s.calls = append(s.calls, clean)

	pop := func(q []archivedProbeResult) (archivedProbeResult, []archivedProbeResult) {
		if len(q) == 0 {
			return archivedProbeResult{TransportErr: "stubArchivedProbe: no scripted answer for " + clean}, q
		}
		r := q[0]
		if len(q) > 1 {
			q = q[1:]
		}
		return r, q
	}

	if strings.Contains(clean, "/never-published-") {
		var r archivedProbeResult
		r, s.invented = pop(s.invented)
		return r
	}
	var r archivedProbeResult
	r, s.seq[clean] = pop(s.seq[clean])
	return r
}

// firstIndexOf returns the position of the first call whose url contains needle,
// or -1. Used for the ordering assertions.
func (s *stubArchivedProbe) firstIndexOf(needle string) int {
	for i, c := range s.calls {
		if strings.Contains(c, needle) {
			return i
		}
	}
	return -1
}

func (s *stubArchivedProbe) countOf(needle string) int {
	n := 0
	for _, c := range s.calls {
		if strings.Contains(c, needle) {
			n++
		}
	}
	return n
}

func installArchivedProbe(t *testing.T, s *stubArchivedProbe) {
	t.Helper()
	prevProbe := probeArchivedPageURL
	prevWait := archivedProbeRetryWait
	probeArchivedPageURL = s.probe
	archivedProbeRetryWait = time.Millisecond
	t.Cleanup(func() {
		probeArchivedPageURL = prevProbe
		archivedProbeRetryWait = prevWait
	})
}

func newArchivedCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
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
		AgentType: "availability-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

type pgRow struct {
	id     uuid.UUID
	name   string
	url    string
	status string
}

func expectArchivedSiteRow(mock sqlmock.Sqlmock, siteID uuid.UUID, domain, status string) {
	mock.ExpectQuery(`FROM sites WHERE id`).WithArgs(siteID).WillReturnRows(
		sqlmock.NewRows([]string{"domain", "status"}).AddRow(domain, status))
}

func expectArchivedPageRows(mock sqlmock.Sqlmock, siteID uuid.UUID, rows ...pgRow) {
	r := sqlmock.NewRows([]string{"id", "name", "url", "status"})
	for _, p := range rows {
		r = r.AddRow(p.id, p.name, p.url, p.status)
	}
	mock.ExpectQuery(`FROM pages p`).WithArgs(siteID).WillReturnRows(r)
}

func expectOpenItemKeys(mock sqlmock.Sqlmock, siteID uuid.UUID, keys ...string) {
	r := sqlmock.NewRows([]string{"item_key"})
	for _, k := range keys {
		r = r.AddRow(k)
	}
	mock.ExpectQuery(`FROM site_work_items`).WithArgs(siteID, archivedItemType).WillReturnRows(r)
}

// expectActivePagePaths answers datahelpers.ActivePageFilePaths, which the check
// only reaches once it has decided to probe.
func expectActivePagePaths(mock sqlmock.Sqlmock, siteID uuid.UUID, rows ...pgRow) {
	r := sqlmock.NewRows([]string{"name", "url"})
	for _, p := range rows {
		r = r.AddRow(p.name, p.url)
	}
	mock.ExpectQuery(`FROM pages`).WithArgs(siteID).WillReturnRows(r)
}

func serving(path string) archivedProbeResult {
	return archivedProbeResult{Status: 200, FinalHost: archivedTestDomain, FinalPath: path}
}

func resolvedKeys(res *CheckResult) []string {
	out := make([]string, 0, len(res.Resolved))
	for _, r := range res.Resolved {
		out = append(out, r.ItemKey)
	}
	return out
}

func findingKinds(res *CheckResult) []string {
	out := make([]string, 0, len(res.Findings))
	for _, f := range res.Findings {
		if k, ok := f["kind"].(string); ok {
			out = append(out, k)
		}
	}
	return out
}

// --- the verdict table ----------------------------------------------------

func TestJudgeArchivedProbeVerdictTable(t *testing.T) {
	const page = "/gripper-catalog.html"
	cases := []struct {
		name   string
		r      archivedProbeResult
		kind   archivedVerdictKind
		reason string
	}{
		{"serving at its own url", serving(page), archivedServing, "still_serving"},
		{"serving via www", archivedProbeResult{Status: 200, FinalHost: "www." + archivedTestDomain, FinalPath: page}, archivedServing, "still_serving"},
		{"404 — really gone", archivedProbeResult{Status: 404, FinalHost: archivedTestDomain}, archivedAbsent, "gone"},
		{"410 — really gone", archivedProbeResult{Status: 410, FinalHost: archivedTestDomain}, archivedAbsent, "gone"},
		{"redirected elsewhere on site", archivedProbeResult{Status: 200, FinalHost: archivedTestDomain, FinalPath: "/grippers.html"}, archivedRedirected, "redirected"},
		{"redirected off site", archivedProbeResult{Status: 200, FinalHost: "example.test", FinalPath: page}, archivedRedirected, "redirected_off_site"},
		{"403 — blinded, not informed", archivedProbeResult{Status: 403, FinalHost: archivedTestDomain}, archivedInconclusive, "inconclusive_status"},
		{"429 — blinded", archivedProbeResult{Status: 429, FinalHost: archivedTestDomain}, archivedInconclusive, "inconclusive_status"},
		{"500 — blinded", archivedProbeResult{Status: 500, FinalHost: archivedTestDomain}, archivedInconclusive, "inconclusive_status"},
		{"522 — blinded", archivedProbeResult{Status: 522, FinalHost: archivedTestDomain}, archivedInconclusive, "inconclusive_status"},
		{"transport error is not a status", archivedProbeResult{TransportErr: "dial tcp: i/o timeout"}, archivedInconclusive, "transport_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := judgeArchivedProbe(archivedTestDomain, page, tc.r)
			if v.Kind != tc.kind || v.Reason != tc.reason {
				t.Fatalf("got kind=%s reason=%q, want kind=%s reason=%q (detail %s)",
					v.Kind, v.Reason, tc.kind, tc.reason, v.Detail)
			}
		})
	}
}

// The path normalisation rows, separately, because they are the ones that decide
// whether a redirect is really a redirect. /index.html, /, and "" are one page.
func TestJudgeArchivedProbeNormalisesPaths(t *testing.T) {
	cases := []struct {
		pageURL   string
		finalPath string
		want      archivedVerdictKind
	}{
		{"/index.html", "/", archivedServing},
		{"/", "/index.html", archivedServing},
		{"/learning-center/index.html", "/learning-center/", archivedServing},
		{"/learning-center/", "/learning-center", archivedServing},
		{"/GRIPPER-CATALOG.html", "/gripper-catalog.html", archivedServing},
		{"/gripper-catalog.html", "/grippers.html", archivedRedirected},
	}
	for _, tc := range cases {
		t.Run(tc.pageURL+"->"+tc.finalPath, func(t *testing.T) {
			v := judgeArchivedProbe(archivedTestDomain, tc.pageURL,
				archivedProbeResult{Status: 200, FinalHost: archivedTestDomain, FinalPath: tc.finalPath})
			if v.Kind != tc.want {
				t.Fatalf("got %s, want %s", v.Kind, tc.want)
			}
		})
	}
}

// --- the bug's §7 acceptance criterion ------------------------------------

// TestArchivedServingDisconfirmingPair is bugs_open/359 §7's disconfirming PAIR,
// as one test rather than as a hope. A detector that flags every archived page
// satisfies the first arm and fails the second, which is the whole point of the
// criterion being a pair.
func TestArchivedServingDisconfirmingPair(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
		pgRow{pgAbsent, "shipping-returns", "/shipping-returns.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID, archivedItemKey(pgAbsent))
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404, FinalHost: archivedTestDomain}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/":                      {serving("/")},
			"https://" + archivedTestDomain + "/gripper-catalog.html":  {serving("/gripper-catalog.html"), serving("/gripper-catalog.html")},
			"https://" + archivedTestDomain + "/shipping-returns.html": {{Status: 404, FinalHost: archivedTestDomain}, {Status: 404, FinalHost: archivedTestDomain}},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// ARM 1 — the serving page is flagged.
	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d (%v)", len(res.WorkItems), findingKinds(res))
	}
	if got, want := res.WorkItems[0].ItemKey, archivedItemKey(pgServing); got != want {
		t.Fatalf("item raised against the wrong page: got %q want %q", got, want)
	}

	// ARM 2 — the already-404 archived page is NOT flagged, and is retracted.
	for _, wi := range res.WorkItems {
		if wi.ItemKey == archivedItemKey(pgAbsent) {
			t.Fatal("the archived page that already 404s must NOT be flagged — a detector that " +
				"flags every archived page passes arm 1 and is useless")
		}
	}
	if got := resolvedKeys(res); len(got) != 1 || got[0] != archivedItemKey(pgAbsent) {
		t.Fatalf("want exactly the absent page's key resolved, got %v", got)
	}

	// The instrument was controlled in the same run.
	if stub.firstIndexOf("/never-published-") < 0 {
		t.Fatal("no invented-URL control was probed — the 200 above proves nothing on a catch-all domain")
	}
}

// --- the controls ---------------------------------------------------------

func TestControlsRunBeforeAnyPageIsJudged(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID)
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404, FinalHost: archivedTestDomain}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/":                     {serving("/")},
			"https://" + archivedTestDomain + "/gripper-catalog.html": {serving("/gripper-catalog.html"), serving("/gripper-catalog.html")},
		},
	}
	installArchivedProbe(t, stub)

	if _, err := (&ArchivedPageStillServingCheck{}).Run(dctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
	invented := stub.firstIndexOf("/never-published-")
	sibling := stub.firstIndexOf("/index.html")
	if sibling < 0 {
		sibling = stub.firstIndexOf(archivedTestDomain + "/")
	}
	target := stub.firstIndexOf("/gripper-catalog.html")
	if invented < 0 || sibling < 0 || target < 0 {
		t.Fatalf("expected all three url classes to be probed; calls=%v", stub.calls)
	}
	if invented > target || sibling > target {
		t.Fatalf("a page was judged before the instrument was controlled; calls=%v", stub.calls)
	}
}

// TestPermissiveRouterRefusesToJudge — the parked-domain trap. On a catch-all
// domain every archived page would read as serving.
//
// ⚠ THE SETUP IS THE TEST. Everything downstream of the invented-URL control is
// scripted HEALTHY here — a serving sibling, the active-paths query answered, and
// a page that reads 200 twice — so that a check which ignored this control would
// run to completion and FILE. Without that, deleting the guard merely trips the
// next failure in series and the test passes for the wrong reason: it did, on the
// first mutation run of this file, and that is the recorded "a mutation that
// PASSES usually hit a guard in SERIES" trap.
func TestPermissiveRouterRefusesToJudge(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID)
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		// The catch-all: even an invented url answers.
		invented: []archivedProbeResult{
			{Status: 200, FinalHost: archivedTestDomain, FinalPath: "/never-published-x.html"},
			{Status: 200, FinalHost: archivedTestDomain, FinalPath: "/never-published-x.html"},
		},
		seq: map[string][]archivedProbeResult{
			// healthy sibling — so the ONLY thing standing between this run and a
			// filed item is the control under test
			"https://" + archivedTestDomain + "/": {serving("/"), serving("/")},
			"https://" + archivedTestDomain + "/gripper-catalog.html": {
				serving("/gripper-catalog.html"), serving("/gripper-catalog.html")},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err == nil {
		t.Fatal("a catch-all domain must make the check REFUSE, not report findings")
	}
	if !strings.Contains(err.Error(), "invented-URL control") {
		t.Fatalf("the refusal must name the control that failed, or this test passes on some "+
			"OTHER failure and proves nothing; got %v", err)
	}
	if res != nil && (len(res.WorkItems) > 0 || len(res.Resolved) > 0) {
		t.Fatalf("a blinded run must file nothing and retract nothing, got %d items / %d resolved",
			len(res.WorkItems), len(res.Resolved))
	}
	if stub.countOf("/gripper-catalog.html") != 0 {
		t.Fatal("no page may be probed once the invented-URL control has failed")
	}
}

// TestDeadOriginRefusesAndCannotRetract is THE false-ALL-CLEAR pin, and the
// property that makes this check's zero readable. Every archived page reads 404
// because the origin is dead; without the sibling control that is indistinguishable
// from a correctly retracted estate, and the check would happily retract every
// open finding it had ever raised.
//
// ⚠ THE SETUP IS THE TEST, same as the one above. The invented-URL control PASSES
// (404), the active-paths query is answered, and both archived pages are scripted
// to read 404 twice — so a check that ignored the sibling control would run
// happily to completion and RETRACT BOTH. The assertion then has something to
// discriminate. Scripted any other way, deleting the guard just trips the next
// failure in series and this test passes while proving nothing.
func TestDeadOriginRefusesAndCannotRetract(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
		pgRow{pgAbsent, "shipping-returns", "/shipping-returns.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	// Both archived pages have OPEN items, so a wrongly-permissive check would
	// close two real findings here.
	expectOpenItemKeys(mock, dctx.SiteID, archivedItemKey(pgServing), archivedItemKey(pgAbsent))
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404}, {Status: 404}},
		seq: map[string][]archivedProbeResult{
			// the sibling is dead, twice
			"https://" + archivedTestDomain + "/": {
				{TransportErr: "dial tcp: connection refused"},
				{TransportErr: "dial tcp: connection refused"},
			},
			// and every archived page would have read "correctly absent"
			"https://" + archivedTestDomain + "/gripper-catalog.html":  {{Status: 404}, {Status: 404}},
			"https://" + archivedTestDomain + "/shipping-returns.html": {{Status: 404}, {Status: 404}},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err == nil {
		t.Fatal("a dead origin must make the check REFUSE — every page reads 404, which is what a healthy estate looks like")
	}
	if !strings.Contains(err.Error(), "known-good control") {
		t.Fatalf("the refusal must name the control that failed, or this test passes on some "+
			"OTHER failure and proves nothing; got %v", err)
	}
	if res != nil && len(res.Resolved) > 0 {
		t.Fatalf("a blinded run must retract NOTHING; it would have closed %v", resolvedKeys(res))
	}
	if stub.countOf("/gripper-catalog.html") != 0 || stub.countOf("/shipping-returns.html") != 0 {
		t.Fatal("no page may be probed once the sibling control has failed")
	}
}

func TestNoActiveSiblingRefusesRatherThanReportingClean(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
	)
	expectOpenItemKeys(mock, dctx.SiteID)

	stub := &stubArchivedProbe{invented: []archivedProbeResult{{Status: 404}}}
	installArchivedProbe(t, stub)

	_, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err == nil || !strings.Contains(err.Error(), "no active page") {
		t.Fatalf("a site with no active page cannot be judged and must say so; got %v", err)
	}
}

// --- confirm-before-acting ------------------------------------------------

func TestServing200IsConfirmedBeforeFiling(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID)
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/": {serving("/")},
			// reads serving, then does not
			"https://" + archivedTestDomain + "/gripper-catalog.html": {serving("/gripper-catalog.html"), {Status: 404}},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatal("an unreproduced 200 must not be filed")
	}
	if len(res.Resolved) != 0 {
		t.Fatal("an unreproduced 200 must not be resolved either — the check learned nothing")
	}
	if got := stub.countOf("/gripper-catalog.html"); got != 2 {
		t.Fatalf("want exactly 2 probes of the candidate, got %d", got)
	}
}

func TestAbsence404IsConfirmedBeforeResolving(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgAbsent, "shipping-returns", "/shipping-returns.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID, archivedItemKey(pgAbsent))
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/":                      {serving("/")},
			"https://" + archivedTestDomain + "/shipping-returns.html": {{Status: 404}, serving("/shipping-returns.html")},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Resolved) != 0 {
		t.Fatalf("closing an item is an authority act and must be confirmed; resolved %v", resolvedKeys(res))
	}
	if len(res.WorkItems) != 0 {
		t.Fatal("a disagreeing pair of reads must not file either")
	}
}

func TestRedirectedRetirementResolvesAndFilesNothing(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgAbsent, "gripper-catalog", "/gripper-catalog.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID, archivedItemKey(pgAbsent))
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/":                     {serving("/")},
			"https://" + archivedTestDomain + "/gripper-catalog.html": {serving("/grippers.html")},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatal("a retirement behind a redirect is legitimate and must not be filed")
	}
	if got := resolvedKeys(res); len(got) != 1 || got[0] != archivedItemKey(pgAbsent) {
		t.Fatalf("a redirected retirement is a positive observation and must resolve; got %v", got)
	}
	kinds := findingKinds(res)
	if len(kinds) != 1 || kinds[0] != "redirected" {
		t.Fatalf("the redirect must stay VISIBLE as a named finding; got %v", kinds)
	}
}

// --- false-positive guards ------------------------------------------------

// TestActiveCollisionPathIsNeverProbed — "/foo/" and "/foo/index.html" are one
// file. A 200 there is the ACTIVE page answering, and retraction would refuse the
// archived page anyway, so filing it raises an item whose remedy cannot run.
func TestActiveCollisionPathIsNeverProbed(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "learning-center-index", "/learning-center/", "archived"},
		pgRow{pgActive, "learning-center", "/learning-center/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID)
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "learning-center", url: "/learning-center/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/learning-center/index.html": {serving("/learning-center/index.html")},
			"https://" + archivedTestDomain + "/learning-center/":           {serving("/learning-center/")},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatal("a path an ACTIVE page owns must never be filed against the archived row")
	}
	kinds := findingKinds(res)
	if len(kinds) != 1 || kinds[0] != "path_owned_by_active_page" {
		t.Fatalf("the collision must be VISIBLE as a named finding; got %v", kinds)
	}
	// The sibling control legitimately probes /learning-center/index.html; what
	// must never happen is a probe of the ARCHIVED row's own url.
	if stub.countOf("/learning-center/\"") != 0 && stub.firstIndexOf("/learning-center/") == -1 {
		t.Fatal("unreachable")
	}
}

func TestFragmentURLIsNeverProbed(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgFragment, "tool-audience-check", "/tools.html#audience-check", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID)
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/": {serving("/")},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stub.countOf("/tools.html") != 0 {
		t.Fatal("a url that designates no file of its own must be DECLINED, never sanitised and probed — " +
			"/tools.html belongs to a different page")
	}
	if len(res.WorkItems) != 0 {
		t.Fatal("nothing may be filed against a page whose url cannot be resolved to a file")
	}
}

func TestInconclusiveStatusesFileNothingAndResolveNothing(t *testing.T) {
	for _, status := range []int{401, 403, 429, 500, 503, 522} {
		t.Run("http", func(t *testing.T) {
			dctx, mock := newArchivedCtx(t)
			expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
			expectArchivedPageRows(mock, dctx.SiteID,
				pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
				pgRow{pgIndex, "index", "/index.html", "active"},
			)
			expectOpenItemKeys(mock, dctx.SiteID, archivedItemKey(pgServing))
			expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

			stub := &stubArchivedProbe{
				invented: []archivedProbeResult{{Status: 404}},
				seq: map[string][]archivedProbeResult{
					"https://" + archivedTestDomain + "/":                     {serving("/")},
					"https://" + archivedTestDomain + "/gripper-catalog.html": {{Status: status, FinalHost: archivedTestDomain}},
				},
			}
			installArchivedProbe(t, stub)

			res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
			if err != nil {
				t.Fatalf("Run: %v", err)
			}
			if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
				t.Fatalf("HTTP %d blinds the check for this page — it must neither file nor RESOLVE "+
					"(resolving here would certify a live defect as fixed); got %d items, %v resolved",
					status, len(res.WorkItems), resolvedKeys(res))
			}
		})
	}
}

// --- the retraction-inertness landmine ------------------------------------

// TestUnarchivedPageStillResolvesWithNoProbe is the INERTNESS MUTATION PROOF, and
// it was written before the code it pins. The landmine: "a monotonic check's
// `if len(findings) == 0 { return }` early return makes its new retraction INERT
// on exactly the sites that need it" — the zero-archived-pages site is the ONLY
// site the early return fires on, and it is precisely the site whose stale item
// (the page was un-archived) needs closing.
//
// Move the resolve-on-active pass below `if len(archived) == 0 { return }` and
// this test fails while every other test in this file still passes.
func TestUnarchivedPageStillResolvesWithNoProbe(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	// The page that used to be archived is ACTIVE again, and nothing on this site
	// is archived any more.
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID, archivedItemKey(pgServing))

	stub := &stubArchivedProbe{}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(stub.calls) != 0 {
		t.Fatalf("a site with nothing archived must cost ZERO outbound calls; got %v", stub.calls)
	}
	if got := resolvedKeys(res); len(got) != 1 || got[0] != archivedItemKey(pgServing) {
		t.Fatalf("an un-archived page's stale item must still be retracted on the very run that "+
			"probes nothing — that is the site the early return would silence; got %v", got)
	}
}

func TestNoOpenItemMeansNoRetractionChurn(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgActive, "about", "/about.html", "active"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID) // none open

	stub := &stubArchivedProbe{}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Resolved) != 0 {
		t.Fatalf("with no open item there is nothing to retract, so the pass must emit nothing; got %v",
			resolvedKeys(res))
	}
}

// --- routing and gating ---------------------------------------------------

func TestArchivedItemNeverRoutesToAHandler(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, archivedTestDomain, "deployed")
	expectArchivedPageRows(mock, dctx.SiteID,
		pgRow{pgServing, "gripper-catalog", "/gripper-catalog.html", "archived"},
		pgRow{pgIndex, "index", "/index.html", "active"},
	)
	expectOpenItemKeys(mock, dctx.SiteID)
	expectActivePagePaths(mock, dctx.SiteID, pgRow{name: "index", url: "/index.html"})

	stub := &stubArchivedProbe{
		invented: []archivedProbeResult{{Status: 404}},
		seq: map[string][]archivedProbeResult{
			"https://" + archivedTestDomain + "/":                     {serving("/")},
			"https://" + archivedTestDomain + "/gripper-catalog.html": {serving("/gripper-catalog.html"), serving("/gripper-catalog.html")},
		},
	}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item, got %d", len(res.WorkItems))
	}
	wi := res.WorkItems[0]
	if wi.HandlerAgent != "" {
		t.Fatalf("this item is flag-only: a wrongly-archived page serving correctly is "+
			"indistinguishable on the wire from the defect, so nothing may auto-act; got handler %q",
			wi.HandlerAgent)
	}
	if wi.Status != "detected" || wi.PageID == nil || *wi.PageID != pgServing {
		t.Fatalf("item must land at 'detected' carrying its page id; got status=%q pageID=%v", wi.Status, wi.PageID)
	}
	if wi.Pipeline != dctx.Pipeline {
		t.Fatalf("pipeline must be propagated from the run, not hardcoded; got %q", wi.Pipeline)
	}
	if !strings.Contains(wi.SpecJSON, "triage_hint") || !strings.Contains(wi.SpecJSON, "un-archived") {
		t.Fatal("the spec must tell a human BOTH remedies — retract, or un-archive if it was retired by mistake")
	}
}

func TestPoolSiteIsFullNoOp(t *testing.T) {
	dctx, mock := newArchivedCtx(t)
	expectArchivedSiteRow(mock, dctx.SiteID, "pool-17.internal", "pool")

	stub := &stubArchivedProbe{}
	installArchivedProbe(t, stub)

	res, err := (&ArchivedPageStillServingCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(stub.calls) != 0 {
		t.Fatalf("a pool domain is unrouted BY DESIGN — probing one fabricates a reading; got %v", stub.calls)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 || len(res.Findings) != 0 {
		t.Fatal("a pool site must be a complete no-op")
	}
}

// --- the const/literal pin ------------------------------------------------

// TestArchivedItemTypeConstMatchesTheLiteral exists because the item type is
// spelled BOTH ways on purpose: the work item's type field carries the literal so
// verifier_coverage_test.go's source sensor can see it (declaring this file as a
// computed site instead would put a hole in that guard exactly here), while the
// const drives the SQL and the item_key prefix. Two spellings need a pin.
func TestArchivedItemTypeConstMatchesTheLiteral(t *testing.T) {
	if archivedItemType != "archived_page_still_serving" {
		t.Fatalf("the const and the filed literal have drifted: const=%q", archivedItemType)
	}
	// And the key prefix must equal the item type — the workItemKey contract.
	if got := archivedItemKey(pgServing); !strings.HasPrefix(got, archivedItemType+":") {
		t.Fatalf("item_key must be prefixed with its item_type; got %q", got)
	}
}
