package browserrunner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

// oneShotGateFixture is the shape bugs_open/126 was filed from: a consent gate
// that hides ITSELF on the first click, which is what the disclaimer doctrine
// asks for and what makes the tool unusable to a second interaction check on a
// shared page. The click handler is the bug file's own three lines, verbatim.
//
// The accept button is sized explicitly so has_visible_area's 24x24 floor is a
// statement about the gate being shown or hidden, not about default button
// metrics on whatever Chromium build is installed.
const oneShotGateFixture = `<!doctype html>
<html lang="en-GB"><head><meta charset="utf-8"><title>gated tool</title>
<style>#rw-accept{width:200px;height:44px}</style></head>
<body>
<div class="tool-container">
  <div id="rw-gate" class="rw-gate">
    <p>This tool can be wrong. Check the figures before you rely on them.</p>
    <button type="button" id="rw-accept" class="rw-accept">I understand</button>
  </div>
  <div id="tool-body" hidden>
    <label for="hours">Hours</label>
    <input type="number" id="hours" value="0">
    <button type="button" id="go">Calculate</button>
    <p id="result">&mdash;</p>
  </div>
</div>
<script>
(function () {
  var gate = document.getElementById('rw-gate');
  var accept = document.getElementById('rw-accept');
  var body = document.getElementById('tool-body');
  accept.addEventListener('click', function () {
    gate.hidden = true;
    gate.style.display = 'none';   // one-shot, by design
    body.hidden = false;
  });
  document.getElementById('go').addEventListener('click', function () {
    var h = parseFloat(document.getElementById('hours').value) || 0;
    document.getElementById('result').textContent = String(h * 2);
  });
})();
</script>
</body></html>`

// TestIntegrationOneShotGateReload is the half the fakePage tests cannot reach:
// a `reload` step really does restore a one-shot consent gate to a clickable
// state in a real browser (bugs_open/126, fix candidate 1).
//
// Unlike TestIntegrationRealBrowser above it needs no network — the fixture is
// served from an httptest server — so it is deterministic and takes seconds. It
// shares the same BROWSER_RUNNER_IT gate because it still needs Chromium:
//
//	BROWSER_RUNNER_IT=1 go test ./internal/adapters/browserrunner/ -run TestIntegrationOneShotGateReload -v
//
// The fence carries its own control. `gate-hidden-after-accept` MUST FAIL: it
// measures the gate at 0x0 once accepted, which is precisely the state
// Playwright reports as "element is not visible", and it is what stops
// `computes-after-reload` from being a test that could not have come out
// otherwise. A fixture whose gate did not hide itself would pass the reload
// check either way — and fail this one. It is measured rather than re-proved by
// clicking the hidden gate, because a real click on it costs Playwright's full
// 30s actionability timeout and reports the same fact.
func TestIntegrationOneShotGateReload(t *testing.T) {
	if os.Getenv("BROWSER_RUNNER_IT") != "1" {
		t.Skip("set BROWSER_RUNNER_IT=1 (with browsers installed) to run")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(oneShotGateFixture))
	}))
	defer srv.Close()

	criteria := `{
	  "profiles": ["desktop"],
	  "checks": [
	    {"id":"gate-visible-at-landing","type":"has_visible_area","selector":"#rw-accept"},
	    {"id":"accept-once","type":"interaction",
	      "steps":[{"action":"click","selector":"#rw-accept"}],
	      "expect":{"selector":"#tool-body"}},
	    {"id":"gate-hidden-after-accept","type":"has_visible_area","selector":"#rw-accept"},
	    {"id":"computes-after-reload","type":"interaction",
	      "steps":[{"action":"reload"},
	               {"action":"click","selector":"#rw-accept"},
	               {"action":"fill","selector":"#hours","value":"10"},
	               {"action":"click","selector":"#go"}],
	      "expect":{"selector":"#result","text_matches":"^20$"}}
	  ]}`

	logger, _ := zap.NewDevelopment()
	a := NewRunChecksAction(logger, nil) // no store — the IT exercises the browser, not S3
	out, err := a.Execute(context.Background(), RunChecksRequest{
		RunID: "it-126", URLs: []string{srv.URL},
		Profiles: []string{"desktop"}, CriteriaJSON: criteria,
		Function: "tool-gated-fixture",
	})
	if err != nil {
		t.Fatalf("real browser run failed: %v", err)
	}
	for _, r := range out.Results {
		t.Logf("%-24s pass=%-5v %s", r.CheckID, r.Pass, r.Detail)
	}
	if r := resultByID(out.Results, "gate-visible-at-landing", "desktop"); r == nil || !r.Pass {
		t.Errorf("the gate must be visible on arrival, else the fixture is not a gate: %+v", r)
	}
	if r := resultByID(out.Results, "accept-once", "desktop"); r == nil || !r.Pass {
		t.Errorf("the first click must accept the gate: %+v", r)
	}
	if r := resultByID(out.Results, "gate-hidden-after-accept", "desktop"); r == nil || r.Pass {
		t.Errorf("the gate must be UNCLICKABLE after accepting — without that this test proves nothing: %+v", r)
	}
	if r := resultByID(out.Results, "computes-after-reload", "desktop"); r == nil || !r.Pass {
		t.Errorf("a reload step must restore the one-shot gate so a later check can drive the tool (bugs_open/126): %+v", r)
	}
}

// TestIntegrationOneShotGateWithoutReloadStillFails re-proves bugs_open/126
// itself against the SAME fixture: two interaction checks that both click the
// gate, neither resetting. The second must fail — and it must still fail after
// this change, because a `reload` action that made gated tools pass whether or
// not the author asked for one would be worse than the bug (acceptance is the
// only tier that tests behaviour rather than markup).
//
// It is a separate function because it is the slow one: the failing click pays
// Playwright's full 30s actionability timeout, so this runs ~35s where the
// reload test runs ~7s. Same BROWSER_RUNNER_IT gate; neither runs by default.
func TestIntegrationOneShotGateWithoutReloadStillFails(t *testing.T) {
	if os.Getenv("BROWSER_RUNNER_IT") != "1" {
		t.Skip("set BROWSER_RUNNER_IT=1 (with browsers installed) to run")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(oneShotGateFixture))
	}))
	defer srv.Close()

	criteria := `{
	  "profiles": ["desktop"],
	  "checks": [
	    {"id":"accept-once","type":"interaction",
	      "steps":[{"action":"click","selector":"#rw-accept"}],
	      "expect":{"selector":"#tool-body"}},
	    {"id":"accept-again","type":"interaction",
	      "steps":[{"action":"click","selector":"#rw-accept"},
	               {"action":"fill","selector":"#hours","value":"10"},
	               {"action":"click","selector":"#go"}],
	      "expect":{"selector":"#result","text_matches":"^20$"}}
	  ]}`

	logger, _ := zap.NewDevelopment()
	a := NewRunChecksAction(logger, nil)
	out, err := a.Execute(context.Background(), RunChecksRequest{
		RunID: "it-126-control", URLs: []string{srv.URL},
		Profiles: []string{"desktop"}, CriteriaJSON: criteria,
		Function: "tool-gated-fixture",
	})
	if err != nil {
		t.Fatalf("real browser run failed: %v", err)
	}
	for _, r := range out.Results {
		t.Logf("%-16s pass=%-5v %s", r.CheckID, r.Pass, r.Detail)
	}
	if r := resultByID(out.Results, "accept-once", "desktop"); r == nil || !r.Pass {
		t.Errorf("the first click must accept the gate: %+v", r)
	}
	r := resultByID(out.Results, "accept-again", "desktop")
	if r == nil || r.Pass {
		t.Fatalf("a second click on a one-shot gate with no reload must FAIL — this is the bug, and a fix that hid it would be worse: %+v", r)
	}
	if !strings.Contains(r.Detail, "step 1 (click #rw-accept)") {
		t.Errorf("the failure must be attributed to the gate click, got %q", r.Detail)
	}
}
