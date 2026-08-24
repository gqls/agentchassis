// Tests for the contrast_ratio check's VERDICT half (the JS only measures;
// judgement lives in Go precisely so these can run without a browser). Each
// canned scan is a value the probe could genuinely return, and each test's
// fixture could come out the other way — a fixture that cannot fail proves
// nothing (WRONG_CALLS 2026-08-03).
package browserrunner

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
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
	return map[string]interface{}{
		"probe": contrastProbeMarker, "located": located,
		"scanned": 42.0, "failures": fs,
	}
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
	if !strings.Contains(detail, "meet their contrast threshold") || !strings.Contains(detail, "42 measured") {
		t.Errorf("pass detail should say what was asserted and how much was measured, got: %s", detail)
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

// A gradient-bounded failure IS firm — it fails the check — but the detail
// must say the number is a best case over a range, not a flat measurement.
// The distinction is load-bearing: a reader who treats a bound as an exact
// reading will "fix" the wrong token.
func TestContrastRatio_GradientBoundedFailsAndSaysSo(t *testing.T) {
	h := hit("div.gi-rules-label", 2.08, false, true)
	h["gradientBounded"] = true
	page := &fakePage{evalResult: scanResult(true, h)}
	pass, detail, r := runOn(t, page, 0)
	if pass {
		t.Fatal("a bounded-backdrop failure is firm and must fail the check")
	}
	if !strings.Contains(detail, "best case over a gradient backdrop") {
		t.Errorf("detail must mark the worst offender's number as a bound: %s", detail)
	}
	if !strings.Contains(detail, "1 of these sit on a bounded gradient backdrop") {
		t.Errorf("detail must count the bounded failures: %s", detail)
	}
	if r.CulpritSelector != "div.gi-rules-label" {
		t.Errorf("bounded failures must still attribute: %q", r.CulpritSelector)
	}
}

// The unbounded wording must name WHY it was not judged, so a reader does not
// read silence as "measured and fine" (the whole PASSES-WHILE-BLIND family).
func TestContrastRatio_UnboundedDetailNamesTheReason(t *testing.T) {
	page := &fakePage{evalResult: scanResult(true,
		hit("p.on-photo", 1.02, true, true),
		hit("div.flat", 2.00, false, true))}
	_, detail, _ := runOn(t, page, 0)
	if !strings.Contains(detail, "UNBOUNDED backdrop") || !strings.Contains(detail, "url() image or no opaque base") {
		t.Errorf("unjudged elements must say why they were unjudged: %s", detail)
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

// The gating objection of council round 7e2391ec: an Evaluate that returns
// NOTHING must never grade as a clean pass. nil is exactly what a page that
// silently swallowed the probe hands back, and before the probe marker it
// decoded to a zero-value scan and PASSED.
func TestContrastRatio_NilResultFailsClosed(t *testing.T) {
	page := &fakePage{evalResult: nil}
	pass, detail, _ := runOn(t, page, 0)
	if pass || !strings.Contains(detail, "probe did not run") {
		t.Errorf("nil Evaluate result must fail closed, got pass=%v detail=%s", pass, detail)
	}
}

func TestContrastRatio_ForeignPayloadFailsClosed(t *testing.T) {
	page := &fakePage{evalResult: map[string]interface{}{"located": true, "failures": []interface{}{}}}
	pass, detail, _ := runOn(t, page, 0)
	if pass || !strings.Contains(detail, "probe did not run") {
		t.Errorf("a payload without the probe marker must fail closed, got pass=%v detail=%s", pass, detail)
	}
}

func TestContrastRatio_ZeroMeasuredFailsClosed(t *testing.T) {
	page := &fakePage{evalResult: map[string]interface{}{
		"probe": contrastProbeMarker, "located": true, "scanned": 0.0, "failures": []interface{}{}}}
	pass, detail, _ := runOn(t, page, 0)
	if pass || !strings.Contains(detail, "0 text-bearing elements") {
		t.Errorf("a scan that measured nothing must not pass vacuously, got pass=%v detail=%s", pass, detail)
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

// TestAuditJSComposition guards the audit's live JS against the literal that was
// running in production before the contrastMathsJS split. It began life as a
// BYTE-IDENTITY test; bugs_open/352 changed the selector composition on purpose,
// so the guarantee is now CONFINEMENT — the JS differs from that literal in ONE
// contiguous region, and that region is pinned by content on both sides.
//
// WHY NOT SIMPLY REPOINT AT A NEW GOLDEN. The digest below was taken from the
// COMPILER'S view of the PRE-REFACTOR const — go/parser + strconv.Unquote over
// b32aa9cd9~1 — precisely so that "a golden silently regenerated from the
// post-refactor code" could not prove the change against itself (council
// 7e2391ec round 2, editquality). A new golden plus a new digest computed from
// MY OWN new code would discard exactly that property: two artefacts derived
// from one source are one artefact. So the 2026-08-22 golden and its historical
// digest STAY, anchored outside this commit, and what changes is what is
// asserted about them.
//
// ⚠ A naive `!strings.Contains(auditJS, "||el.tagName")` tripwire would be WRONG
// here: the legacy `cls` echo deliberately KEEPS that expression, because an
// un-rolled agent-chassis still reads it. Confinement is what distinguishes
// "the echo still exists" (correct) from "the selector still derives from it"
// (the bug).
func TestAuditJSComposition(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "audit_js_golden_2026-08-22.txt"))
	if err != nil {
		t.Fatalf("golden missing: %v", err)
	}
	golden := strings.TrimSuffix(string(raw), "\n")

	const preRefactorSHA = "4ec6cb73d2481686087c636820ffe3198025e796841a9ed8e845f53757258da7"
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(golden))); got != preRefactorSHA {
		t.Fatalf("golden no longer matches the pre-refactor const (sha %s, want %s) — it has been regenerated, not verified", got, preRefactorSHA)
	}

	// The shared maths kernel is NOT part of the 352 change, so it keeps the
	// original identity guarantee in full: byte-present in both strings.
	if !strings.Contains(golden, contrastMathsJS) {
		t.Fatal("the pre-refactor golden no longer contains contrastMathsJS — the kernel or the golden has moved")
	}
	if !strings.Contains(auditJS, contrastMathsJS) {
		t.Fatal("auditJS no longer embeds contrastMathsJS byte-for-byte — a join-point defect is possible again")
	}

	// One contiguous divergence: common prefix, common suffix, one region between.
	p := 0
	for p < len(auditJS) && p < len(golden) && auditJS[p] == golden[p] {
		p++
	}
	sfx := 0
	for sfx < len(auditJS)-p && sfx < len(golden)-p &&
		auditJS[len(auditJS)-1-sfx] == golden[len(golden)-1-sfx] {
		sfx++
	}
	oldRegion := golden[p : len(golden)-sfx]
	newRegion := auditJS[p : len(auditJS)-sfx]

	// The divergence must sit where 352 changed the JS. If a later edit changes
	// the probe somewhere else as well, the two edits merge into one region and
	// these assertions fail — which is the point: an unreviewed second change
	// cannot hide inside the reviewed one.
	if !strings.Contains(oldRegion, "||el.tagName") {
		t.Fatalf("the divergence from the pre-refactor literal is NOT the selector block —\n"+
			"something else in the probe has changed and must be reviewed on its own merits.\nold region: %q", oldRegion)
	}
	if !strings.Contains(newRegion, "indexOf.call(nodes,el)") {
		t.Fatalf("the in-page selector verification is missing from the changed region — a filed\n"+
			"selector would no longer be proven to select the element that was measured (bugs_open/352).\nnew region: %q", newRegion)
	}
	if !strings.Contains(newRegion, "selector:sel") {
		t.Fatalf("the composed selector is no longer emitted on the finding; the consumer would fall\n"+
			"back to the TAG.TAG composition for every element.\nnew region: %q", newRegion)
	}
	// The legacy echo must survive: removing it makes an un-rolled chassis
	// compose bare-tag keys and falsely retract every legacy row.
	// ⚠ Assert the whole EXPRESSION, not the substring "||el.tagName": the
	// comment immediately above that line in the source says "do not clean up the
	// ||el.tagName fallback", so a substring check matches the COMMENT and passes
	// while the CODE is gone. Caught by mutating the line out and watching this
	// test still pass — see WRONG_CALLS 2026-08-24.
	const clsEcho = `var cls=(typeof el.className==='string'?el.className:'')||el.tagName;`
	if !strings.Contains(newRegion, clsEcho) {
		t.Fatalf("the legacy cls echo has been removed or altered — an un-rolled agent-chassis reads\n"+
			"that field, and without it retraction closes live rows on a key-shape mismatch\n"+
			"(bugs_open/352 D1).\nnew region: %q", newRegion)
	}

	if strings.Count(auditJS, "function effBG") != 1 {
		t.Errorf("effBG must appear exactly once in the composed probe")
	}
}

// errFake is a trivial error type so tests need no fmt.Errorf import dance.
type errFake string

func (e errFake) Error() string { return string(e) }
