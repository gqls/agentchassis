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
	a := NewRenderAuditAction(zap.NewNop())
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
	a := NewRenderAuditAction(zap.NewNop())
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
