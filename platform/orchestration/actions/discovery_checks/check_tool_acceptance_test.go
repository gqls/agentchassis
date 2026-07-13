package discovery_checks

import (
	"encoding/json"
	"testing"

	"github.com/gqls/agentchassis/platform/content"
)

// The anchor rule's founding cases (RUNBOOK 2026-07-09, tool-xp-curve-designer):
// #tableWrap tr passes on its anchor even though rows are JS-built;
// #xpTableBody tr fails because the anchor exists nowhere.
const samplePage = `<html><body>
<div class="tool-page">
  <div class="tool-container">
    <select id="curveType"><option>linear</option></select>
    <div id="tableWrap"></div>
    <canvas id="xpChart"></canvas>
  </div>
</div>
<script src="/tools/assets/tool-xp-curve-designer.js"></script>
</body></html>`

func TestSelectorAnchor(t *testing.T) {
	cases := map[string]string{
		"#tableWrap tr":      "#tableWrap",
		"#xpTableBody tr":    "#xpTableBody",
		".tool-container":    ".tool-container",
		".a > li":            ".a",
		"div.stat":           "div",
		"  #spaced  ":        "#spaced",
		"":                   "",
		"[data-x='1']":       "",
	}
	for sel, want := range cases {
		if got := selectorAnchor(sel); got != want {
			t.Errorf("selectorAnchor(%q) = %q, want %q", sel, got, want)
		}
	}
}

func TestAnchorPresent(t *testing.T) {
	if !anchorPresent(samplePage, "#tableWrap") {
		t.Error("#tableWrap should be present")
	}
	if anchorPresent(samplePage, "#xpTableBody") {
		t.Error("#xpTableBody should be absent (the invented selector)")
	}
	if !anchorPresent(samplePage, ".tool-container") {
		t.Error(".tool-container should be present")
	}
	if !anchorPresent(samplePage, "canvas") {
		t.Error("bare tag anchor should match")
	}
	if anchorPresent(samplePage, ".tool") {
		t.Error("class match must be word-bounded: .tool must not match tool-page/tool-container")
	}
}

func TestEvaluateStaticCriteria(t *testing.T) {
	raw := `{
	  "profiles": ["desktop","mobile"],
	  "checks": [
	    {"id":"boots","type":"selector_exists","selector":".tool-container"},
	    {"id":"console","type":"no_console_errors"},
	    {"id":"asset","type":"asset_loads","path":"/tools/assets/tool-xp-curve-designer.js"},
	    {"id":"status","type":"page_status_ok"},
	    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
	    {"id":"curve-switch","type":"interaction",
	      "steps":[{"action":"select","selector":"#curveType","value":"exponential"}],
	      "expect":{"selector":"#tableWrap tr"}},
	    {"id":"invented","type":"selector_exists","selector":"#xpTableBody tr"},
	    {"id":"later-EDIT","type":"selector_exists","selector":"#placeholder"}
	  ]}`
	var crit criteriaDoc
	if err := json.Unmarshal([]byte(raw), &crit); err != nil {
		t.Fatalf("criteria should parse: %v", err)
	}

	ev := evaluateStaticCriteria(crit, 200, samplePage)

	wantPassed := map[string]bool{"boots": true, "asset": true, "status": true, "curve-switch": true}
	for _, p := range ev.passed {
		if !wantPassed[p.id] {
			t.Errorf("unexpected pass: %s (%s)", p.id, p.detail)
		}
		delete(wantPassed, p.id)
	}
	for id := range wantPassed {
		t.Errorf("expected pass missing: %s", id)
	}

	if len(ev.failed) != 1 || ev.failed[0].id != "invented" {
		t.Errorf("expected exactly the invented selector to fail, got %+v", ev.failed)
	}

	skipped := map[string]bool{}
	for _, s := range ev.skipped {
		skipped[s.id] = true
	}
	for _, id := range []string{"console", "mobile-fit", "later-EDIT"} {
		if !skipped[id] {
			t.Errorf("expected %s to be skipped", id)
		}
	}
}

func TestEvaluateShellChecks(t *testing.T) {
	// Build the leaked page from the real sentinel so the test stays in
	// lockstep with content.ToolDocOpen.
	leaked := samplePage + "\n<script>" + content.ToolDocOpen + "\npurpose: x\n=== end tool-doc === */</script>"
	ev := evaluateStaticCriteria(criteriaDoc{}, 200, leaked)
	found := false
	for _, f := range ev.failed {
		if f.id == "shell-doc-header" {
			found = true
		}
	}
	if !found {
		t.Error("leaked tool-doc header should fail the shell check")
	}

	ev = evaluateStaticCriteria(criteriaDoc{}, 200, samplePage+"<p><no value></p>")
	found = false
	for _, f := range ev.failed {
		if f.id == "shell-template-residue" {
			found = true
		}
	}
	if !found {
		t.Error("'<no value>' residue should fail the shell check")
	}
}

func TestEvaluatePageStatusFail(t *testing.T) {
	var crit criteriaDoc
	_ = json.Unmarshal([]byte(`{"checks":[{"id":"status","type":"page_status_ok"}]}`), &crit)
	ev := evaluateStaticCriteria(crit, 404, "<html></html>")
	if len(ev.failed) == 0 || ev.failed[0].id != "status" {
		t.Errorf("HTTP 404 should fail page_status_ok, got %+v", ev.failed)
	}
}

func TestExtractCriteriaFence(t *testing.T) {
	body := "# PLAN\n\n## Acceptance criteria\n```criteria\n{\"checks\":[]}\n```\ntail"
	if got := extractCriteriaFence(body); got != `{"checks":[]}` {
		t.Errorf("fence extraction got %q", got)
	}
	if extractCriteriaFence("no fence here") != "" {
		t.Error("missing fence should return empty")
	}
	if extractCriteriaFence("```criteria\nunclosed") != "" {
		t.Error("unclosed fence should return empty")
	}
}
