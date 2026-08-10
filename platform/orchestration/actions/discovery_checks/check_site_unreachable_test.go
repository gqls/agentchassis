// FILE: platform/orchestration/actions/discovery_checks/check_site_unreachable_test.go
//
// bugs_open/236 (522 half) — pins site_unreachable.
//
// Like asset_reference_404, this check has NO live positive: measured
// 2026-08-10, all 21 deployed sites serve 200 + HTML, so production cannot
// demonstrate the check bites and "0 findings" is also what a silently broken
// check reports. The substitute is the same: an induced fault per verdict row,
// plus a recorded mutation for each guard, each proven load-bearing by breaking
// it and watching a NAMED test fail:
//
//	mutation                                          test that catches it
//	------------------------------------------------  ----------------------------------
//	drop the confirming second probe                  BlipOnFirstProbeFilesNothing
//	treat an off-domain redirect as unreachable       DeliberateDelegationFilesNothing
//	file on a missing stored title                    TitleAbsentIsAFindingNotAnItem
//	skip the pool-site status gate                    PoolSiteIsNeverProbed
//	give the item a handler_agent                     NeverRoutesToAHandler
//	retract on an unreachable verdict too             UnreachableDoesNotRetract
//	drop AllOfType from the retraction                HealthyRetractsAllOfType
//	treat 522 as reachable (2xx check widened)        CloudflareFiveTwoTwoFiles
//
// The verdict table itself (judgeSiteProbe) is pure and covered row by row.

package discovery_checks

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// stubSiteProbe answers from a queue per domain and records calls, so a test can
// assert how many probes happened as well as what they said.
type stubSiteProbe struct {
	mu    sync.Mutex
	seq   map[string][]siteProbeResult
	calls []string
}

func (s *stubSiteProbe) probe(_ context.Context, domain string) siteProbeResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, domain)
	q := s.seq[domain]
	if len(q) == 0 {
		return siteProbeResult{TransportErr: "stubSiteProbe: no scripted answer for " + domain}
	}
	r := q[0]
	if len(q) > 1 {
		s.seq[domain] = q[1:]
	}
	return r
}

func installSiteProbe(t *testing.T, s *stubSiteProbe) {
	t.Helper()
	prevProbe := probeSiteOrigin
	prevWait := siteProbeRetryWait
	probeSiteOrigin = s.probe
	siteProbeRetryWait = time.Millisecond
	t.Cleanup(func() {
		probeSiteOrigin = prevProbe
		siteProbeRetryWait = prevWait
	})
}

func newSiteUnreachableCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
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

func expectSiteRow(mock sqlmock.Sqlmock, siteID uuid.UUID, domain, status string) {
	mock.ExpectQuery(`FROM sites WHERE id`).WithArgs(siteID).WillReturnRows(
		sqlmock.NewRows([]string{"domain", "status"}).AddRow(domain, status))
}

func expectTitleRow(mock sqlmock.Sqlmock, siteID uuid.UUID, title string) {
	mock.ExpectQuery(`FROM pages p`).WithArgs(siteID).WillReturnRows(
		sqlmock.NewRows([]string{"title"}).AddRow(title))
}

const healthyBody = `<!doctype html><html><head><title>Lendzy — Know the Rules Before You Borrow</title></head><body>x</body></html>`

// --- judgeSiteProbe: every verdict row of the PLAN table ---

func TestJudgeSiteProbeVerdictTable(t *testing.T) {
	title := "Lendzy — Know the Rules Before You Borrow"
	cases := []struct {
		name        string
		title       string
		r           siteProbeResult
		unreachable bool
		reason      string
	}{
		{"transport error", title, siteProbeResult{TransportErr: "dial tcp: i/o timeout"}, true, "transport_error"},
		{"cloudflare 522", title, siteProbeResult{Status: 522, FinalHost: "lendzy.co.uk", Body: []byte("error")}, true, "http_522"},
		{"apex 404", title, siteProbeResult{Status: 404, FinalHost: "lendzy.co.uk", Body: []byte("nope")}, true, "http_404"},
		{"origin 500", title, siteProbeResult{Status: 500, FinalHost: "lendzy.co.uk", Body: []byte("boom")}, true, "http_500"},
		{"2xx empty body", title, siteProbeResult{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte("  \n")}, true, "empty_body"},
		{"2xx non-html", title, siteProbeResult{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte(`{"ok":true}`)}, true, "not_html"},
		{"off-domain redirect", title, siteProbeResult{Status: 200, FinalHost: "webdesign.co.uk", Body: []byte(healthyBody)}, false, "delegated"},
		{"title absent (parked or stale)", title, siteProbeResult{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte(`<html><title>Domain for sale</title></html>`)}, false, "title_absent"},
		{"healthy", title, siteProbeResult{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte(healthyBody)}, false, "healthy"},
		{"healthy via www", title, siteProbeResult{Status: 200, FinalHost: "www.lendzy.co.uk", Body: []byte(healthyBody)}, false, "healthy"},
		{"healthy with no stored title", "", siteProbeResult{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte(`<html><body>anything</body></html>`)}, false, "healthy"},
		{"entity-escaped title still matches", "Fish & Chips", siteProbeResult{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte(`<html><title>Fish &amp; Chips</title></html>`)}, false, "healthy"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := judgeSiteProbe("lendzy.co.uk", tc.title, tc.r)
			if v.Unreachable != tc.unreachable || v.Reason != tc.reason {
				t.Fatalf("got unreachable=%v reason=%q, want unreachable=%v reason=%q (detail: %s)",
					v.Unreachable, v.Reason, tc.unreachable, tc.reason, v.Detail)
			}
		})
	}
}

// --- Run: filing, retraction, gating ---

func TestConfirmedOutageFilesExactlyOneAlertItem(t *testing.T) {
	dctx, mock := newSiteUnreachableCtx(t)
	expectSiteRow(mock, dctx.SiteID, "lendzy.co.uk", "deployed")
	expectTitleRow(mock, dctx.SiteID, "Lendzy — Know the Rules Before You Borrow")
	stub := &stubSiteProbe{seq: map[string][]siteProbeResult{
		"lendzy.co.uk": {{Status: 522, FinalHost: "lendzy.co.uk", Body: []byte("cf error")}},
	}}
	installSiteProbe(t, stub)

	res, err := (&SiteUnreachableCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(stub.calls) != 2 {
		t.Fatalf("a filed outage must be confirmed by a second probe; got %d probe(s)", len(stub.calls))
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want exactly 1 work item, got %d", len(res.WorkItems))
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "site_unreachable" || wi.Severity != "high" || wi.Status != "detected" {
		t.Fatalf("item mis-shaped: %+v", wi)
	}
	if wi.ItemKey != "site_unreachable:"+dctx.SiteID.String() {
		t.Fatalf("item key must dedup per site: %q", wi.ItemKey)
	}
	if !strings.Contains(wi.Summary, "http_522") {
		t.Fatalf("summary must carry the reason: %q", wi.Summary)
	}
	if len(res.Resolved) != 0 {
		t.Fatalf("UnreachableDoesNotRetract: an outage must not self-clear, got %+v", res.Resolved)
	}
}

func TestCloudflareFiveTwoTwoFiles(t *testing.T) {
	// Pins the motivating case narrowly: 522 is not a 2xx, whatever the body.
	v := judgeSiteProbe("lendzy.co.uk", "", siteProbeResult{Status: 522, FinalHost: "lendzy.co.uk", Body: []byte(healthyBody)})
	if !v.Unreachable {
		t.Fatal("522 with an HTML error body must still be unreachable")
	}
}

func TestBlipOnFirstProbeFilesNothing(t *testing.T) {
	dctx, mock := newSiteUnreachableCtx(t)
	expectSiteRow(mock, dctx.SiteID, "lendzy.co.uk", "deployed")
	expectTitleRow(mock, dctx.SiteID, "Lendzy — Know the Rules Before You Borrow")
	stub := &stubSiteProbe{seq: map[string][]siteProbeResult{
		"lendzy.co.uk": {
			{TransportErr: "read: connection reset by peer"},
			{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte(healthyBody)},
		},
	}}
	installSiteProbe(t, stub)

	res, err := (&SiteUnreachableCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("a transient blip must not file; got %+v", res.WorkItems)
	}
	if len(stub.calls) != 2 {
		t.Fatalf("want the confirming re-probe, got %d call(s)", len(stub.calls))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("recovered probe must retract, got %d", len(res.Resolved))
	}
}

func TestHealthyRetractsAllOfType(t *testing.T) {
	dctx, mock := newSiteUnreachableCtx(t)
	expectSiteRow(mock, dctx.SiteID, "lendzy.co.uk", "deployed")
	expectTitleRow(mock, dctx.SiteID, "Lendzy — Know the Rules Before You Borrow")
	stub := &stubSiteProbe{seq: map[string][]siteProbeResult{
		"lendzy.co.uk": {{Status: 200, FinalHost: "lendzy.co.uk", Body: []byte(healthyBody)}},
	}}
	installSiteProbe(t, stub)

	res, err := (&SiteUnreachableCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(stub.calls) != 1 {
		t.Fatalf("a healthy first probe needs no confirmation, got %d call(s)", len(stub.calls))
	}
	if len(res.WorkItems) != 0 || len(res.Findings) != 0 {
		t.Fatalf("healthy site must file nothing: items=%v findings=%v", res.WorkItems, res.Findings)
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want 1 retraction, got %d", len(res.Resolved))
	}
	r := res.Resolved[0]
	if r.ItemType != "site_unreachable" || !r.AllOfType || r.ItemKey != "" || r.Reason == "" {
		t.Fatalf("retraction must be the wide, stated, reasoned one: %+v", r)
	}
}

func TestDeliberateDelegationFilesNothing(t *testing.T) {
	// webdesign.uk 302 → webdesign.co.uk, measured 2026-08-10. Reachable.
	dctx, mock := newSiteUnreachableCtx(t)
	expectSiteRow(mock, dctx.SiteID, "webdesign.uk", "deployed")
	expectTitleRow(mock, dctx.SiteID, "webdesign.uk: A complete website for your business, one fixed price")
	stub := &stubSiteProbe{seq: map[string][]siteProbeResult{
		"webdesign.uk": {{Status: 200, FinalHost: "webdesign.co.uk", Body: []byte(healthyBody)}},
	}}
	installSiteProbe(t, stub)

	res, err := (&SiteUnreachableCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("delegation must not file: %+v", res.WorkItems)
	}
	if len(res.Findings) != 1 || res.Findings[0]["reason"] != "delegated" {
		t.Fatalf("delegation must be visible as a finding: %+v", res.Findings)
	}
	if len(res.Resolved) != 1 {
		t.Fatal("a serving (delegated) site still answers every open unreachable item")
	}
}

func TestTitleAbsentIsAFindingNotAnItem(t *testing.T) {
	// mortgagecalculator.co.uk today: on-host 200 HTML with a divergent title.
	dctx, mock := newSiteUnreachableCtx(t)
	expectSiteRow(mock, dctx.SiteID, "mortgagecalculator.co.uk", "deployed")
	expectTitleRow(mock, dctx.SiteID, "Mortgage Calculator UK — Free Tools & Insider Guides")
	stub := &stubSiteProbe{seq: map[string][]siteProbeResult{
		"mortgagecalculator.co.uk": {{Status: 200, FinalHost: "mortgagecalculator.co.uk",
			Body: []byte(`<html><title>The UK's Authority on Mortgage Finance</title></html>`)}},
	}}
	installSiteProbe(t, stub)

	res, err := (&SiteUnreachableCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("title mismatch must not file (measured 1/21 false-positive rate): %+v", res.WorkItems)
	}
	if len(res.Findings) != 1 || res.Findings[0]["reason"] != "title_absent" {
		t.Fatalf("title mismatch must be visible as a finding: %+v", res.Findings)
	}
}

func TestPoolSiteIsNeverProbed(t *testing.T) {
	dctx, mock := newSiteUnreachableCtx(t)
	expectSiteRow(mock, dctx.SiteID, "somepoolsite.uk", "pool")
	stub := &stubSiteProbe{seq: map[string][]siteProbeResult{}}
	installSiteProbe(t, stub)

	res, err := (&SiteUnreachableCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(stub.calls) != 0 {
		t.Fatalf("a pool site's domain is unrouted BY DESIGN; probing it fabricates an outage. probed: %v", stub.calls)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Fatalf("pool site must be a full no-op: %+v", res)
	}
}

func TestNeverRoutesToAHandler(t *testing.T) {
	dctx, mock := newSiteUnreachableCtx(t)
	expectSiteRow(mock, dctx.SiteID, "lendzy.co.uk", "deployed")
	expectTitleRow(mock, dctx.SiteID, "")
	stub := &stubSiteProbe{seq: map[string][]siteProbeResult{
		"lendzy.co.uk": {{TransportErr: "dial tcp: no route to host"}},
	}}
	installSiteProbe(t, stub)

	res, err := (&SiteUnreachableCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 item, got %d", len(res.WorkItems))
	}
	if res.WorkItems[0].HandlerAgent != "" {
		t.Fatalf("alert-only: no agent can repair routing today, got handler %q", res.WorkItems[0].HandlerAgent)
	}
}
