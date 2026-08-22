// Tests for the contrast_ratio check's VERDICT half (the JS only measures;
// judgement lives in Go precisely so these can run without a browser). Each
// canned scan is a value the probe could genuinely return, and each test's
// fixture could come out the other way — a fixture that cannot fail proves
// nothing (WRONG_CALLS 2026-08-03).
package browserrunner

import (
	"strings"
	"testing"
)

// hit builds a probe failure record in the wire shape Evaluate returns
// (playwright hands back maps, not structs).
func hit(selector string, ratio float64, overImage, inTool bool) map[string]interface{} {
	return map[string]interface{}{
		"selector": selector, "text": "sample text", "fg": "rgb(252,92,125)",
		"bg": "rgb(124,60,255)", "ratio": ratio, "need": 4.5,
		"overImage": overImage, "px": 16.0, "inTool": inTool,
	}
}

func scanResult(located bool, hits ...map[string]interface{}) map[string]interface{} {
	fs := make([]interface{}, len(hits))
	for i, h := range hits {
		fs[i] = h
	}
	return map[string]interface{}{"located": located, "failures": fs}
}

func runOn(t *testing.T, page *fakePage, minRatio float64) (bool, string, CheckResult) {
	t.Helper()
	doc := criteriaDoc{Container: ".tool-container"}
	ch := criteriaCheck{ID: "legible", Type: "contrast_ratio", MinRatio: minRatio}
	return runContrastRatio(page, doc, ch, "mobile")
}

func TestContrastRatio_PassesWhenClean(t *testing.T) {
	pass, detail, _ := runOn(t, &fakePage{evalResult: scanResult(true)}, 0)
	if !pass {
		t.Fatalf("clean scan must pass, got fail: %s", detail)
	}
	if !strings.Contains(detail, "all text meets its contrast threshold") {
		t.Errorf("pass detail should say what was asserted, got: %s", detail)
	}
}

func TestContrastRatio_FailsWithAttribution(t *testing.T) {
	page := &fakePage{evalResult: scanResult(true, hit("div.gi-eyebrow", 1.66, false, true))}
	pass, detail, r := runOn(t, page, 0)
	if pass {
		t.Fatal("a firm in-tool failure must fail the check")
	}
	// The long marker is what the pod-grep deploy proof keys on — a rewording
	// silently breaks the RUNBOOK recipe.
	if !strings.Contains(detail, "text is painted below its contrast threshold") {
		t.Errorf("detail lost the deploy-proof marker: %s", detail)
	}
	if !strings.Contains(detail, "1.66:1") || !strings.Contains(detail, "div.gi-eyebrow") {
		t.Errorf("detail must name the offender and its ratio: %s", detail)
	}
	if r.Scope != ScopeTool {
		t.Errorf("in-tool offender must scope to tool, got %q", r.Scope)
	}
	if r.CulpritSelector != "div.gi-eyebrow" {
		t.Errorf("fixer handle wrong: %q", r.CulpritSelector)
	}
}

func TestContrastRatio_OverImageNeverFails(t *testing.T) {
	page := &fakePage{evalResult: scanResult(true,
		hit("p.hero-caption", 1.02, true, true),
		hit("span.overlay", 2.10, true, false))}
	pass, detail, _ := runOn(t, page, 0)
	if !pass {
		t.Fatalf("over-image readings are approximate and must never fail; got: %s", detail)
	}
	if !strings.Contains(detail, "2 element(s) over an image or gradient backdrop") {
		t.Errorf("the unjudged count must be visible, not silent: %s", detail)
	}
}

func TestContrastRatio_ChromeAttribution(t *testing.T) {
	page := &fakePage{evalResult: scanResult(true, hit("a.site-nav-link", 1.90, false, false))}
	pass, _, r := runOn(t, page, 0)
	if pass {
		t.Fatal("chrome text a visitor cannot read must still fail the check")
	}
	if r.Scope != ScopeChrome {
		t.Errorf("out-of-container offender must scope to chrome, got %q", r.Scope)
	}
}

func TestContrastRatio_ContainerNotFoundIsUnknownScope(t *testing.T) {
	page := &fakePage{evalResult: scanResult(false, hit("h2.title", 1.20, false, false))}
	pass, detail, r := runOn(t, page, 0)
	if pass {
		t.Fatal("a firm failure must fail even when attribution is unknown")
	}
	if r.Scope != ScopeUnknown || !strings.Contains(detail, "attribution unknown") {
		t.Errorf("missing container must not be blamed on chrome: scope=%q detail=%s", r.Scope, detail)
	}
}

func TestContrastRatio_WorstOffenderWins(t *testing.T) {
	page := &fakePage{evalResult: scanResult(true,
		hit("p.dim", 3.10, false, true),
		hit("span.invisible", 1.01, false, true))}
	_, _, r := runOn(t, page, 0)
	if r.CulpritSelector != "span.invisible" {
		t.Errorf("the lowest ratio is the culprit, got %q", r.CulpritSelector)
	}
}

func TestContrastRatio_EvaluateErrorFails(t *testing.T) {
	page := &fakePage{evalErr: errFake("boom")}
	pass, detail, _ := runOn(t, page, 0)
	if pass || !strings.Contains(detail, "could not measure contrast") {
		t.Errorf("a probe that cannot run must fail loudly, got pass=%v detail=%s", pass, detail)
	}
}

func TestContrastRatio_BadShapeFails(t *testing.T) {
	page := &fakePage{evalResult: map[string]interface{}{"located": true,
		"failures": []interface{}{map[string]interface{}{"ratio": "2.48"}}}}
	pass, detail, _ := runOn(t, page, 0)
	if pass || !strings.Contains(detail, "could not measure contrast") {
		t.Errorf("a ratio arriving as a string means the payload shape changed — never a pass: pass=%v detail=%s", pass, detail)
	}
}

func TestContrastProbe_EmbedsThresholdAndContainer(t *testing.T) {
	probe := contrastProbe(`.tool-container, [class*="x"]`, 2.5)
	if !strings.Contains(probe, "var minRatio = 2.5;") {
		t.Errorf("min_ratio not embedded: %s", probe[:200])
	}
	if !strings.Contains(probe, `[class*=\"x\"]`) && !strings.Contains(probe, `[class*="x"]`) {
		t.Errorf("container selector not embedded safely")
	}
	if strings.Contains(contrastProbe(".t", 0), "var minRatio = 0;") == false {
		t.Errorf("zero must mean the per-element WCAG default, embedded as 0")
	}
}

// TestContrastRatio_WiredIntoTheLadder proves the case arm exists end to end:
// a criteria fence naming the type reaches the probe and its verdict — the
// exact seam LANDMINES:512 is about (an unknown type is SKIPPED, silently).
func TestContrastRatio_WiredIntoTheLadder(t *testing.T) {
	doc := criteriaDoc{Checks: []criteriaCheck{{ID: "legible", Type: "contrast_ratio"}}}
	applicable, skipped := splitByProfile(doc, "mobile", "https://x.test/")
	if len(applicable) != 1 || len(skipped) != 0 {
		t.Fatalf("contrast_ratio must be applicable, not skipped: %d/%d", len(applicable), len(skipped))
	}
	page := &fakePage{status: 200, evalResult: scanResult(true, hit("div.gi-eyebrow", 1.66, false, true))}
	results := evaluateOnPage(page, doc, applicable, "mobile", "https://x.test/")
	r := resultByID(results, "legible", "mobile")
	if r == nil || r.Pass || r.CulpritSelector != "div.gi-eyebrow" {
		t.Fatalf("wired verdict wrong: %+v", r)
	}
	if len(page.evalScripts) != 1 || !strings.Contains(page.evalScripts[0], "function effBG") {
		t.Errorf("the probe handed to Evaluate must carry the shared maths kernel")
	}
}

// TestAuditJSComposition guards the refactor that moved the WCAG maths into
// contrastMathsJS: the render audit's composed probe must still carry every
// helper and its own scan body. The audit is live fleet machinery — a lost
// fragment here is a silent behaviour change on every weekly sweep.
func TestAuditJSComposition(t *testing.T) {
	for _, frag := range []string{
		"function parseRGB", "function lum", "function ratio", "function over(",
		"function effBG", "var out={contrast:[],images:[],overflow:null}",
		"r:128,g:128,b:128", "size>=18.66&&weight>=700", "return out;",
	} {
		if !strings.Contains(auditJS, frag) {
			t.Errorf("auditJS lost %q in the maths extraction", frag)
		}
	}
	if strings.Count(auditJS, "function effBG") != 1 {
		t.Errorf("effBG must appear exactly once in the composed probe")
	}
}

// errFake is a trivial error type so tests need no fmt.Errorf import dance.
type errFake string

func (e errFake) Error() string { return string(e) }
