package browserrunner

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"go.uber.org/zap"
)

// probeJSON builds what the in-page probe returns, so the tests exercise the
// real decode path rather than a hand-built Go struct.
func probeJSON(t *testing.T, body string) interface{} {
	t.Helper()
	var v interface{}
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		t.Fatalf("bad fixture: %v", err)
	}
	return v
}

func auditWith(page *fakePage) *RenderAuditAction {
	a := NewRenderAuditAction(zap.NewNop(), nil)
	a.open = func(context.Context, string, string, *zap.Logger) (browserPage, error) {
		return page, nil
	}
	return a
}

func TestRenderAuditReportsContrastBrokenImagesAndOverflow(t *testing.T) {
	page := &fakePage{status: 200, evalResult: probeJSON(t, `{
	  "contrast":[{"cls":"card-link","tag":"A","text":"Read it","fg":"rgb(27,42,59)",
	               "bg":"rgb(27,42,59)","ratio":1,"need":4.5,"overImage":false,"px":16}],
	  "images":[{"src":"/assets/images/missing.png","alt":"a hero"}],
	  "overflow":{"scrollWidth":1400,"viewport":1280}}`)}

	res, err := auditWith(page).Execute(context.Background(),
		RenderAuditRequest{RunID: "r1", URLs: []string{"https://example.com/"}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got := res.Summary.Contrast; got != 1 {
		t.Fatalf("want 1 contrast finding, got %d", got)
	}
	if res.Contrast[0].URL != "https://example.com/" {
		t.Errorf("finding must carry its URL, got %q", res.Contrast[0].URL)
	}
	if res.Contrast[0].Ratio != 1 || res.Contrast[0].Class != "card-link" {
		t.Errorf("finding not decoded: %+v", res.Contrast[0])
	}
	if res.Summary.BrokenImages != 1 || res.Images[0].Src != "/assets/images/missing.png" {
		t.Errorf("broken image not reported: %+v", res.Images)
	}
	if res.Summary.OverflowPages != 1 || res.Overflow[0].ScrollWidth != 1400 {
		t.Errorf("overflow not reported: %+v", res.Overflow)
	}
	if res.Summary.PagesFailed != 1 {
		t.Errorf("page with a firm failure must count as failed, got %d", res.Summary.PagesFailed)
	}
}

// The load-bearing distinction: an over-image reading is an approximation and
// must NOT on its own turn a page red. oufe.com's last finding was exactly this
// — a white button over a near-black hero — and a screenshot confirmed it fine.
func TestOverImageFindingIsReportedButDoesNotFailThePage(t *testing.T) {
	page := &fakePage{status: 200, evalResult: probeJSON(t, `{
	  "contrast":[{"cls":"btn","tag":"A","text":"How this site works","fg":"rgb(255,255,255)",
	               "bg":"rgb(128,128,128)","ratio":3.95,"need":4.5,"overImage":true,"px":16}],
	  "images":[],"overflow":null}`)}

	res, err := auditWith(page).Execute(context.Background(),
		RenderAuditRequest{URLs: []string{"https://example.com/"}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Summary.Contrast != 1 {
		t.Fatal("the finding must still be REPORTED so a human can judge it")
	}
	if !res.Contrast[0].OverImage {
		t.Error("the over_image flag must survive to the caller")
	}
	if res.Summary.ContrastFirm != 0 {
		t.Errorf("an approximation is not a firm failure, got %d", res.Summary.ContrastFirm)
	}
	if res.Summary.PagesFailed != 0 {
		t.Errorf("an over-image reading alone must not fail the page, got %d", res.Summary.PagesFailed)
	}
}

func TestCleanPageProducesNoFindings(t *testing.T) {
	page := &fakePage{status: 200, evalResult: probeJSON(t,
		`{"contrast":[],"images":[],"overflow":null}`)}
	res, err := auditWith(page).Execute(context.Background(),
		RenderAuditRequest{URLs: []string{"https://example.com/"}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Summary.Contrast != 0 || res.Summary.PagesFailed != 0 || res.Summary.Pages != 1 {
		t.Fatalf("clean page misreported: %+v", res.Summary)
	}
}

// A page that will not load must be REPORTED, never silently dropped — a dead
// page passing as clean is the worst possible outcome for a checker.
func TestUnreachablePageIsReportedNotSkipped(t *testing.T) {
	a := NewRenderAuditAction(zap.NewNop(), nil)
	a.open = func(_ context.Context, url string, _ string, _ *zap.Logger) (browserPage, error) {
		if url == "https://dead.example/" {
			return nil, errors.New("dial tcp: connection refused")
		}
		return &fakePage{status: 200, evalResult: probeJSON(t,
			`{"contrast":[],"images":[],"overflow":null}`)}, nil
	}
	res, err := a.Execute(context.Background(), RenderAuditRequest{
		URLs: []string{"https://dead.example/", "https://ok.example/"}})
	if err != nil {
		t.Fatalf("one dead page must not abort the run: %v", err)
	}
	if len(res.Unreachable) != 1 || res.Unreachable[0] != "https://dead.example/" {
		t.Fatalf("unreachable page not reported: %+v", res.Unreachable)
	}
	if res.Summary.Pages != 2 {
		t.Errorf("both pages should be counted, got %d", res.Summary.Pages)
	}
	if res.Summary.PagesFailed != 1 {
		t.Errorf("an unreachable page is a failed page, got %d", res.Summary.PagesFailed)
	}
}

func TestNavigationFailureIsUnreachable(t *testing.T) {
	page := &fakePage{navErr: "net::ERR_NAME_NOT_RESOLVED"}
	res, err := auditWith(page).Execute(context.Background(),
		RenderAuditRequest{URLs: []string{"https://nope.example/"}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(res.Unreachable) != 1 {
		t.Fatalf("a navigation error must be reported as unreachable: %+v", res)
	}
}

func TestNoURLsIsAnError(t *testing.T) {
	if _, err := auditWith(&fakePage{}).Execute(context.Background(),
		RenderAuditRequest{}); err == nil {
		t.Fatal("an empty request must error rather than report a clean run")
	}
}

// ── A1.1: whole-site renders (desktop + mobile) ────────────────────────────

func TestCaptureRendersSavesDesktopAndMobilePerPage(t *testing.T) {
	pages := map[string]*fakePage{
		"desktop": {status: 200, evalResult: probeJSON(t, `{"contrast":[],"images":[],"overflow":null}`)},
		"mobile":  {status: 200, evalResult: probeJSON(t, `{"contrast":[],"images":[],"overflow":null}`)},
	}
	store := &fakeStore{}
	a := NewRenderAuditAction(zap.NewNop(), store)
	a.open = func(_ context.Context, _ string, profile string, _ *zap.Logger) (browserPage, error) {
		return pages[profile], nil
	}

	res, err := a.Execute(context.Background(), RenderAuditRequest{
		RunID: "run-9", SiteID: "site-9", URLs: []string{"https://example.com/"},
		CaptureRenders: true})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(res.Renders) != 2 {
		t.Fatalf("want 2 renders (desktop+mobile), got %d: %+v", len(res.Renders), res.Renders)
	}
	wantKeys := map[string]bool{
		"render-sweep/site-9/run-9/0_desktop.png": true,
		"render-sweep/site-9/run-9/0_mobile.png":  true,
	}
	for _, k := range store.keys {
		if !wantKeys[k] {
			t.Errorf("unexpected render key %q", k)
		}
		delete(wantKeys, k)
	}
	if len(wantKeys) != 0 {
		t.Errorf("missing render keys: %v (saved: %v)", wantKeys, store.keys)
	}
	// Renders are pixels to look at, never failure evidence.
	for _, r := range res.Renders {
		if len(r.FailingChecks) != 0 {
			t.Errorf("a sweep render must carry no FailingChecks: %+v", r)
		}
	}
}

func TestCaptureRendersWithNoStoreDegradesToMeasurementOnly(t *testing.T) {
	page := &fakePage{status: 200, evalResult: probeJSON(t,
		`{"contrast":[{"cls":"x","tag":"P","fg":"rgb(1,1,1)","bg":"rgb(0,0,0)","ratio":1,"need":4.5,"overImage":false,"px":16}],"images":[],"overflow":null}`)}
	res, err := auditWith(page).Execute(context.Background(), RenderAuditRequest{
		URLs: []string{"https://example.com/"}, CaptureRenders: true})
	if err != nil {
		t.Fatalf("no store must not error: %v", err)
	}
	if len(res.Renders) != 0 {
		t.Fatalf("nil store must yield no renders, got %+v", res.Renders)
	}
	if res.Summary.ContrastFirm != 1 {
		t.Fatalf("measurement must be untouched by the degrade, got %+v", res.Summary)
	}
}

func TestRenderCaptureFailureNeverLosesTheMeasurement(t *testing.T) {
	pages := map[string]*fakePage{
		"desktop": {status: 200, shotErr: errors.New("chromium OOM"),
			evalResult: probeJSON(t, `{"contrast":[{"cls":"y","tag":"P","fg":"rgb(1,1,1)","bg":"rgb(0,0,0)","ratio":1.5,"need":4.5,"overImage":false,"px":16}],"images":[],"overflow":null}`)},
		"mobile": {status: 200, evalResult: probeJSON(t, `{"contrast":[],"images":[],"overflow":null}`)},
	}
	store := &fakeStore{}
	a := NewRenderAuditAction(zap.NewNop(), store)
	a.open = func(_ context.Context, _ string, profile string, _ *zap.Logger) (browserPage, error) {
		return pages[profile], nil
	}
	res, err := a.Execute(context.Background(), RenderAuditRequest{
		RunID: "run-10", SiteID: "site-10", URLs: []string{"https://example.com/"},
		CaptureRenders: true})
	if err != nil {
		t.Fatalf("a failed capture must not fail the audit: %v", err)
	}
	if res.Summary.ContrastFirm != 1 {
		t.Fatalf("the measurement was lost with the picture: %+v", res.Summary)
	}
	if len(res.Renders) != 1 || res.Renders[0].Profile != "mobile" {
		t.Fatalf("want exactly the mobile render, got %+v", res.Renders)
	}
	// The degrade must be COUNTED, not only logged (council 46640fe2,
	// bug_historian): a consumer must be able to tell a partial sweep from a
	// complete one without reading pod logs.
	if res.RendersFailed != 1 {
		t.Fatalf("want renders_failed=1 for the lost desktop shot, got %d", res.RendersFailed)
	}
}

// bugs_open/242: the requester's max_pages cap bite must be echoed into the
// summary — the reply is the only part of an awaited chassis step that reaches
// the stored artefact, so `pages: 25` with no total beside it reads as a
// complete sweep.
func TestSummaryEchoesRequesterTruncationClaim(t *testing.T) {
	page := &fakePage{status: 200, evalResult: probeJSON(t, `{"contrast":[],"images":[],"overflow":null}`)}

	res, err := auditWith(page).Execute(context.Background(),
		RenderAuditRequest{RunID: "r1", URLs: []string{"https://example.com/"},
			PagesTotal: 27, Truncated: true})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if res.Summary.PagesTotal != 27 || !res.Summary.Truncated {
		t.Fatalf("summary must echo the requester's cap bite, got total=%d truncated=%v",
			res.Summary.PagesTotal, res.Summary.Truncated)
	}
	buf, err := json.Marshal(res.Summary)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, key := range []string{`"pages_total":27`, `"truncated":true`} {
		if !json.Valid(buf) || !containsSubstring(string(buf), key) {
			t.Errorf("marshalled summary must carry %s, got %s", key, buf)
		}
	}
}

// The version-skew guarantee: an old-shape request (no cap fields) must produce
// the old summary shape byte-for-byte — omitempty pinned, so a chassis/adapter
// skew in either direction degrades to today's behaviour, never to a zero that
// reads like a measured total.
func TestOldShapeRequestKeepsOldSummaryShape(t *testing.T) {
	page := &fakePage{status: 200, evalResult: probeJSON(t, `{"contrast":[],"images":[],"overflow":null}`)}

	res, err := auditWith(page).Execute(context.Background(),
		RenderAuditRequest{RunID: "r1", URLs: []string{"https://example.com/"}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	buf, err := json.Marshal(res.Summary)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, key := range []string{"pages_total", "truncated"} {
		if containsSubstring(string(buf), key) {
			t.Errorf("old-shape request must not grow a %q key, got %s", key, buf)
		}
	}
}

// PagesAudited is the identity half of Pages, and the two DELIBERATELY
// disagree: Pages counts pages ATTEMPTED, PagesAudited names those
// successfully MEASURED. The gap is exactly the unreachable page — and a
// consumer that retracts work items against the wrong one of these closes
// tickets on a page that never loaded, which is the error Unreachable was
// added to prevent ("it would let a dead page pass as clean").
func TestPagesAuditedNamesOnlyPagesActuallyMeasured(t *testing.T) {
	a := NewRenderAuditAction(zap.NewNop(), nil)
	a.open = func(_ context.Context, url string, _ string, _ *zap.Logger) (browserPage, error) {
		switch url {
		case "https://dead.example/gone.html":
			return nil, errors.New("dial tcp: connection refused")
		case "https://ok.example/failing.html":
			return &fakePage{status: 200, evalResult: probeJSON(t, `{
			  "contrast":[{"cls":"card-title","tag":"H2","text":"Plans","fg":"rgb(17,17,17)",
			               "bg":"rgb(15,15,15)","ratio":1.05,"need":4.5,"overImage":false,"px":20}],
			  "images":[],"overflow":null}`)}, nil
		default:
			return &fakePage{status: 200, evalResult: probeJSON(t,
				`{"contrast":[],"images":[],"overflow":null}`)}, nil
		}
	}

	res, err := a.Execute(context.Background(), RenderAuditRequest{URLs: []string{
		"https://ok.example/repaired.html",
		"https://dead.example/gone.html",
		"https://ok.example/failing.html",
	}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	// The repaired page reports NOTHING and must still be named — that is the
	// entire reason this field exists, since a clean page is otherwise
	// indistinguishable from one never visited. The failing page must be named
	// too: it WAS measured, and per-selector retraction depends on saying so.
	want := []string{"https://ok.example/repaired.html", "https://ok.example/failing.html"}
	if len(res.Summary.PagesAudited) != len(want) {
		t.Fatalf("want %d audited pages, got %v", len(want), res.Summary.PagesAudited)
	}
	for i, w := range want {
		if res.Summary.PagesAudited[i] != w {
			t.Errorf("audited[%d]: want %q, got %q", i, w, res.Summary.PagesAudited[i])
		}
	}
	for _, got := range res.Summary.PagesAudited {
		if got == "https://dead.example/gone.html" {
			t.Fatal("an unreachable page must NEVER appear in pages_audited")
		}
	}
	// The disagreement itself is the guarantee, so pin both numbers.
	if res.Summary.Pages != 3 {
		t.Errorf("Pages counts every attempt, want 3, got %d", res.Summary.Pages)
	}
	if len(res.Unreachable) != 1 {
		t.Errorf("the dead page must still be reported unreachable, got %v", res.Unreachable)
	}
}

// A run in which nothing loaded must produce an EMPTY audited set, not an
// absent one that a consumer could read as "no scoping needed".
func TestPagesAuditedIsEmptyWhenNothingLoaded(t *testing.T) {
	a := NewRenderAuditAction(zap.NewNop(), nil)
	a.open = func(context.Context, string, string, *zap.Logger) (browserPage, error) {
		return nil, errors.New("dial tcp: connection refused")
	}
	res, err := a.Execute(context.Background(),
		RenderAuditRequest{URLs: []string{"https://dead.example/a.html", "https://dead.example/b.html"}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(res.Summary.PagesAudited) != 0 {
		t.Fatalf("nothing loaded, so nothing was audited: %v", res.Summary.PagesAudited)
	}
	buf, err := json.Marshal(res.Summary)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if containsSubstring(string(buf), "pages_audited") {
		t.Errorf("an empty audited set must omit the key (omitempty), got %s", buf)
	}
}

func containsSubstring(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
