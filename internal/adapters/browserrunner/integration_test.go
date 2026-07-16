package browserrunner

import (
	"context"
	"os"
	"testing"

	"go.uber.org/zap"
)

// TestIntegrationRealBrowser drives REAL headless Chromium against a live
// page — the in-process half of the 035 §2.15 smoke. Gated because it needs
// installed browsers and network:
//
//	go run github.com/mxschmitt/playwright-go/cmd/playwright install chromium
//	BROWSER_RUNNER_IT=1 go test ./internal/adapters/browserrunner/ -run Integration -v
//
// Optionally set BROWSER_RUNNER_IT_URL to point at a specific tool page;
// defaults to the gamesdesign xp-curve-designer (criteria per its live PLAN).
func TestIntegrationRealBrowser(t *testing.T) {
	if os.Getenv("BROWSER_RUNNER_IT") != "1" {
		t.Skip("set BROWSER_RUNNER_IT=1 (with browsers installed) to run")
	}
	url := os.Getenv("BROWSER_RUNNER_IT_URL")
	if url == "" {
		url = "https://gamesdesign.co.uk/tools/tool-xp-curve-designer.html"
	}

	criteria := `{
	  "profiles": ["desktop"],
	  "checks": [
	    {"id":"boots","type":"selector_exists","selector":".tool-container"},
	    {"id":"rows","type":"selector_exists","selector":"#tableWrap tr"},
	    {"id":"console","type":"no_console_errors"},
	    {"id":"status","type":"page_status_ok"},
	    {"id":"gone","type":"selector_exists","selector":"#definitely-not-here"}
	  ]}`

	logger, _ := zap.NewDevelopment()
	a := NewRunChecksAction(logger, nil) // no store — the IT exercises the browser, not S3
	out, err := a.Execute(context.Background(), RunChecksRequest{
		RunID: "it-1", URLs: []string{url},
		Profiles: []string{"desktop"}, CriteriaJSON: criteria,
		Function: "tool-xp-curve-designer",
	})
	if err != nil {
		t.Fatalf("real browser run failed: %v", err)
	}
	for _, r := range out.Results {
		t.Logf("%-8s pass=%-5v %s", r.CheckID, r.Pass, r.Detail)
	}
	if r := resultByID(out.Results, "status", "desktop"); r == nil || !r.Pass {
		t.Errorf("status should pass on the live page, got %+v", r)
	}
	if r := resultByID(out.Results, "rows", "desktop"); r == nil || !r.Pass {
		t.Errorf("rows (#tableWrap tr, JS-built) should pass in a REAL browser — this is the whole point of Tier 4; got %+v", r)
	}
	if r := resultByID(out.Results, "gone", "desktop"); r == nil || r.Pass {
		t.Errorf("a selector that exists nowhere must fail, got %+v", r)
	}
}
