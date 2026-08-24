// FILE: platform/orchestration/actions/render_seam_no_fallback_test.go
//
// The regression set for bugs_open/260: a component render either EXECUTED or
// ERRORED, and there is no third state.
//
// The third state was a regex renderer written for handlebars syntax that the
// seam dropped to whenever text/template failed. It substituted {{.field}}
// values and left every {{if}}, {{range}} and {{end}} it could not execute in
// the output — well-formed HTML, values resolved, control directives intact —
// so nothing downstream could tell a rendered page from an unrendered one, and
// 26 page builds across 7 domains were refused for it in the eight days to
// 2026-08-18.
//
// THE POSITIVE CONTROLS ARE NOT DECORATION. The owner has ruled every site
// should be able to have tool pages, and a page ABOUT templates legitimately
// carries {{ }} in its copy — one of the 26 recorded occurrences was exactly
// that, harmless content that merely looked like the bug. A fix that made both
// the broken page and the tool page fail would not be a fix, so the brace-
// bearing render below must PASS, and it is the test that would catch anyone
// "improving" this seam into an output brace-scan.

package actions

import (
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// The live shape: mechanism-flow ranges over steps, and each step may range
// over branches. The nested range is where the recorded failures happened.
const mechanismFlowLikeTemplate = `<section class="mechanism">
  <h2>{{.section_title}}</h2>
  {{range .steps}}<div class="step">
    <h3>{{.title}}</h3><p>{{.body}}</p>
    {{range .branches}}<span class="branch">{{.label}}: {{.body}}</span>{{end}}
  </div>{{end}}
</section>`

func mechanismFlowCtx(branches interface{}) *RenderContext {
	return &RenderContext{
		Domain: "example.com",
		ContentData: map[string]interface{}{
			"section_title": "How it works",
			"steps": []interface{}{
				map[string]interface{}{"title": "Assess", "body": "We look."},
				map[string]interface{}{"title": "Decide", "body": "We choose.", "branches": branches},
			},
		},
	}
}

// A — the negative case, and the exact live defect: a sentence where the schema
// declares a list.
func TestRenderFailsOnAMistypedNestedField(t *testing.T) {
	out, _, _, err := RenderTemplate(mechanismFlowLikeTemplate,
		mechanismFlowCtx("Either we file, or we appeal."), zap.NewNop())

	if err == nil {
		t.Fatalf("a {{range}} over a string must fail the render; got output:\n%s", out)
	}
	if out != "" {
		t.Errorf("a failed render must produce NO output; got:\n%s", out)
	}
	if !strings.Contains(err.Error(), "range can't iterate over") {
		t.Errorf("the error must carry text/template's own diagnosis, got: %v", err)
	}
}

// B — the control that makes A evidence rather than assertion. THE SAME
// template and the same data with `branches` in its declared shape must render
// cleanly, with no directive left behind. A change that makes both A and B fail
// is not a fix.
func TestRenderSucceedsOnCorrectlyShapedContent(t *testing.T) {
	out, _, _, err := RenderTemplate(mechanismFlowLikeTemplate,
		mechanismFlowCtx([]interface{}{
			map[string]interface{}{"label": "File", "body": "We file."},
			map[string]interface{}{"label": "Appeal", "body": "We appeal."},
		}), zap.NewNop())

	if err != nil {
		t.Fatalf("correctly shaped content must render: %v", err)
	}
	if strings.Contains(out, "{{") || strings.Contains(out, "}}") {
		t.Errorf("a successful render left template directives in the output:\n%s", out)
	}
	for _, want := range []string{"How it works", "We file.", "Appeal"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered output is missing %q:\n%s", want, out)
		}
	}
}

// THE TOOL-PAGE CONTROL, required by the owner's ruling that every site should
// be able to have tools. Content that CONTAINS braces is not content that
// failed to render, and this seam must never confuse the two. It passes by
// construction — the design fires on a template engine error and never scans
// output — and this test is what fails if anyone adds such a scan.
func TestBraceBearingCopyRendersAndKeepsItsBraces(t *testing.T) {
	ctx := &RenderContext{ContentData: map[string]interface{}{
		"heading": "Prompt library",
		"body":    "Use {{ variable }} in your prompt, then {{ another }} after it.",
		"snippet": "{{#each items}}<li>{{this}}</li>{{/each}}",
	}}

	out, _, _, err := RenderTemplate(`<h2>{{.heading}}</h2><p>{{.body}}</p><pre>{{.snippet}}</pre>`, ctx, zap.NewNop())
	if err != nil {
		t.Fatalf("copy that merely contains braces must render: %v", err)
	}
	if !strings.Contains(out, "{{ variable }}") {
		t.Errorf("the page's own braces were altered — a tool page must survive verbatim:\n%s", out)
	}
	if !strings.Contains(out, "{{#each items}}") {
		t.Errorf("a handlebars EXAMPLE in copy must survive verbatim — the retired dialect is not "+
			"substituted any more, and it must not be mangled either:\n%s", out)
	}
}

// A PARSE error is refused exactly like an execute error. This is the case the
// LANDMINE names: `{{if $.x}}` written inside a CSS comment, where the author
// believed comments were inert to the template engine.
func TestParseErrorIsRefusedNotDegraded(t *testing.T) {
	const broken = `<style>/* {{if $.dark}} */ .a{color:#000}</style><p>{{.body}}</p>`
	out, _, _, err := RenderTemplate(broken, &RenderContext{
		ContentData: map[string]interface{}{"body": "hello"}}, zap.NewNop())
	if err == nil {
		t.Fatalf("an unterminated {{if}} must fail to parse, not degrade; got:\n%s", out)
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("want a parse error, got: %v", err)
	}
	if out != "" {
		t.Errorf("a failed parse must produce no output; got:\n%s", out)
	}
}

// THE RETIRED DIALECT IS NOW UNRENDERABLE, not merely unused. A template naming
// {{nav_items_html}} used to be regex-patched by the fallback; it cannot even
// parse now ("function not defined"), so it hard-fails. Pinned so nobody
// re-authors one and wonders why it renders empty.
func TestRetiredHandlebarsDialectHardFails(t *testing.T) {
	out, _, _, err := RenderTemplate(`<nav>{{nav_items_html}}</nav>`, &RenderContext{
		NavItems: []NavItem{{Label: "Home", URL: "/index.html"}}}, zap.NewNop())
	if err == nil {
		t.Fatalf("the retired {{nav_items_html}} dialect must fail loudly, got:\n%s", out)
	}
	if !strings.Contains(out, "") || out != "" {
		t.Errorf("no output may be emitted for a template the engine cannot parse: %q", out)
	}
}

// An empty template is not an error: "executed and produced nothing" is a
// different fact from "could not execute", and rerender_page_sections carries a
// section on the first while failing on the second.
func TestEmptyTemplateIsNotAnError(t *testing.T) {
	out, _, _, err := RenderTemplate("", &RenderContext{}, zap.NewNop())
	if err != nil || out != "" {
		t.Fatalf("empty template: got (%q, %v), want (\"\", nil)", out, err)
	}
}

// GateConvertedTemplate is the RFC_034 acceptance gate. Before the seam had an
// error channel it rendered through the fallback, so it could green-light a
// template the REAL renderer cannot execute.
func TestConversionGateRefusesATemplateTheRendererCannotExecute(t *testing.T) {
	const converted = `<div id="{{.InstanceID}}-wrap">{{range .items}}<b>{{.}}</b>{{end}}</div>`
	// items is absent, so missingkey=zero makes the range a no-op — that must
	// still PASS, or the gate would refuse every data-less template.
	if _, err := GateConvertedTemplate("tool-x", converted, zap.NewNop()); err != nil {
		t.Fatalf("a template whose range has no data must still gate cleanly: %v", err)
	}

	const unparseable = `<div id="{{.InstanceID}}-wrap">{{if .x}}<b>y</b></div>`
	needsJudged, err := GateConvertedTemplate("tool-x", unparseable, zap.NewNop())
	if err == nil {
		t.Fatal("the gate must refuse a converted template the real renderer cannot parse")
	}
	if needsJudged {
		t.Error("an unrenderable template is a transform defect, never a judged-pool case")
	}
}

// §13g — THE TWO SEAMS DO NOT ACCEPT THE SAME LANGUAGE, and nothing else says
// so. RenderTemplateWithMap (the contact-info re-render) has no FuncMap, so
// {{safe}}, {{default}} and {{isset}} — ordinary in every component template —
// are PARSE errors there. That divergence used to be invisible because both
// paths answered with "" or with mangled output; now the second one returns an
// error, and its caller leaves the live block alone rather than deleting it.
func TestContactInfoSeamRejectsTheComponentLibraryFuncMap(t *testing.T) {
	const tmpl = `<div class="contact"><p>{{safe .email}}</p></div>`
	data := map[string]interface{}{"email": "hello@example.com"}

	if _, err := RenderTemplateWithMap(tmpl, data, zap.NewNop()); err == nil {
		t.Fatal("CONTROL FAILED: RenderTemplateWithMap appears to have a FuncMap now — " +
			"if that is deliberate, this test and bugs_open/260 §13g need rewriting, " +
			"because the two seams' languages have converged")
	}

	// The same template through the component seam must SUCCEED — that
	// asymmetry is the finding.
	if _, _, _, err := RenderTemplate(tmpl, &RenderContext{
		ContentData: map[string]interface{}{"email": "hello@example.com"}}, zap.NewNop()); err != nil {
		t.Fatalf("{{safe}} must work on the component seam: %v", err)
	}

	// And the ordinary case still renders there.
	out, err := RenderTemplateWithMap(`<p>{{.email}}</p>`, data, zap.NewNop())
	if err != nil {
		t.Fatalf("a plain field render must succeed on the contact-info seam: %v", err)
	}
	if !strings.Contains(out, "hello@example.com") {
		t.Errorf("contact-info render lost its data: %s", out)
	}
}

// CONTAINED 2026-08-24, ROUND 2 (council 661bcf00, guardian HIGH). For one
// commit this seam REFUSED an {{.InstanceID}} template with no token bound. The
// guardian seat objected that converting a shared render seam's log-only defect
// path into a hard error is new authority shipped unconditional, licensed only
// by a census of today's callers — and RFC_044 is the estate's precedent for
// what happens next if that is re-argued rather than contained. So it PUBLISHES
// now: ctx.UnboundInstanceToken, the same shape as AbsentRequiredFields, which
// is the remedy the owner ruling the submission cited actually used on this
// same function.
//
// The test asserts the PUBLICATION, not the log. A log line is not readable by
// code, which is the whole reason the field exists; asserting a logger.Error
// here would pin the thing that was already insufficient.
func TestRenderTemplate_publishesUnboundInstanceToken(t *testing.T) {
	const needsToken = `<section id="{{.InstanceID}}"><button id="{{.InstanceID}}-go"></button></section>`

	rc := &RenderContext{}
	out, _, _, err := RenderTemplate(needsToken, rc, zap.NewNop())
	if err != nil {
		t.Fatalf("the seam REPORTS this rather than refusing (see the write site's "+
			"note, and RFC_050) — a render error here means the containment was "+
			"undone without updating this test: %v", err)
	}
	if !rc.UnboundInstanceToken {
		t.Fatal("an {{.InstanceID}} template rendered with no token bound must PUBLISH " +
			"the fact on the context — otherwise the only signal is a log line, and " +
			"this estate has an owner ruling that a named log is not escalation")
	}
	// And the damage the flag is reporting must actually be there, or the flag
	// is reporting something that is not happening.
	if !strings.Contains(out, `id=""`) {
		t.Errorf("CONTROL FAILED: the unbound render should have produced an EMPTY id "+
			"under missingkey=zero, got: %s", out)
	}
	if DetectInstanceCollisions(out + out).Clean() {
		t.Error("CONTROL FAILED: two copies of an unbound render must be reported by " +
			"the detector — otherwise nothing downstream sees this either")
	}

	// POSITIVE CONTROL 1 — bound, it renders and the flag stays FALSE. Without
	// this the assertion above would pass against a seam that sets the flag
	// unconditionally.
	bc := &RenderContext{}
	BindInstanceToken(bc, InstanceToken("faq", 0))
	bound, _, _, err := RenderTemplate(needsToken, bc, zap.NewNop())
	if err != nil {
		t.Fatalf("CONTROL FAILED: a BOUND render must succeed: %v", err)
	}
	if bc.UnboundInstanceToken {
		t.Error("CONTROL FAILED: a bound render must not report an unbound token")
	}
	if !strings.Contains(bound, `id="c-faq"`) {
		t.Errorf("CONTROL FAILED: bound render lost the token: %s", bound)
	}

	// POSITIVE CONTROL 2 — a template that does not use the token must render
	// with nothing bound and must not be reported. This is the majority of the
	// corpus (140 of 297 active templates spell {{.InstanceID}} as of
	// 2026-08-24; the rest do not), and reporting them all would make the field
	// meaningless.
	pc := &RenderContext{}
	if _, _, _, err := RenderTemplate(`<section id="static"><p>{{.body}}</p></section>`, pc, zap.NewNop()); err != nil {
		t.Fatalf("CONTROL FAILED: a template with no {{.InstanceID}} must render unbound: %v", err)
	}
	if pc.UnboundInstanceToken {
		t.Error("CONTROL FAILED: a template that never spells the token must not be reported")
	}
}

// The empty-id refusal must be routable as a VALUE, not by matching its message.
// Two council seats objected to the first cut classifying it with
// strings.Contains(err.Error(), "rendered EMPTY"), both citing this file's own
// landmine: a converter's refusal REASON is a routing signal callers grep for,
// and rewording it makes a defect ride the renamed reason. This test is what
// makes the sentinel load-bearing rather than decorative — it reworders the
// message and requires the routing to survive.
func TestGate_emptyIdRefusalIsRoutableAsAValue(t *testing.T) {
	const residue = `<section id="{{.InstanceID}}-wrap"><div id="{{.category_slug}}">x</div></section>`

	_, err := GateConvertedTemplate("category-listing", residue, zap.NewNop())
	if err == nil {
		t.Fatal("expected the empty-id refusal")
	}
	if !errors.Is(err, ErrEmptyElementID) {
		t.Fatalf("the empty-id refusal must be identifiable with errors.Is, not by "+
			"message text — a caller routing on it breaks silently the next time "+
			"the wording changes. got: %v", err)
	}
	// The sentinel must DISCRIMINATE, or errors.Is would be worthless: another
	// hard refusal from the same function must NOT match it.
	const mangled = `<div id="{{.InstanceID}}-wrap">{{if .x}}<b>y</b></div>`
	if _, err := GateConvertedTemplate("tool-x", mangled, zap.NewNop()); err == nil {
		t.Fatal("CONTROL FAILED: an unparseable template must still be refused")
	} else if errors.Is(err, ErrEmptyElementID) {
		t.Fatalf("CONTROL FAILED: a parse failure must not classify as an empty id — "+
			"the sentinel is matching everything: %v", err)
	}
}

// The gate's own `id="-` check sees only the token-empty shape
// id="{{.InstanceID}}-suffix". It cannot see an id whose WHOLE value resolved
// to nothing — either id="{{.InstanceID}}" spelled on its own (6 of the 140
// active InstanceID templates as of 2026-08-24, generic-text-block among them)
// or some other field that rendered empty, which is the live dartsonline case
// (category-listing's id="{{.category_slug}}").
func TestGate_refusesEmptyIdResidue(t *testing.T) {
	// Tokens ARE bound here — that is what makes this a transform defect rather
	// than the binding failure the render seam now refuses.
	const residue = `<section id="{{.InstanceID}}-wrap"><div id="{{.category_slug}}">x</div></section>`

	needsJudged, err := GateConvertedTemplate("category-listing", residue, zap.NewNop())
	if err == nil {
		t.Fatal("the gate must refuse a converted template that renders an EMPTY id " +
			"under real tokens — shipping it puts an unaddressable element into the " +
			"corpus through the gate that exists to keep it out")
	}
	if needsJudged {
		t.Error("an empty id is a transform defect, never a judged-pool case: the " +
			"judged pool is for scripts that need rewriting, not for an element " +
			"nothing can address")
	}

	// CONTROL — the same shape with the id supplied must gate cleanly, or this
	// test would pass against a gate that refuses everything.
	const sound = `<section id="{{.InstanceID}}-wrap"><div id="{{.InstanceID}}-inner">x</div></section>`
	if needsJudged, err := GateConvertedTemplate("category-listing", sound, zap.NewNop()); err != nil || needsJudged {
		t.Fatalf("CONTROL FAILED: a soundly converted template must pass "+
			"(err=%v needsJudged=%v)", err, needsJudged)
	}
}

// F2 of the 2026-08-24 Fable review: the second-door report was the ONLY new
// mechanism nothing could kill — deleting its whole block left the package
// green, because a log line is asserted by nothing. This pins it with an
// observed logger (the refused_link_targets_test.go idiom).
//
// The path is currently LINKER-DEAD (RerenderSitePagesAction is registered
// nowhere — see the REACHABILITY note at the call site), so this test is the
// only executor the report has: it guards the revival case, and without it the
// report could rot to nothing before the path ever came back.
func TestRenderTemplateWithMap_reportsUnboundInstanceToken(t *testing.T) {
	const needsToken = `<section id="{{.InstanceID}}"><p>{{.title}}</p></section>`
	report := "no per-instance token was bound"
	count := func(logs *observer.ObservedLogs) int {
		n := 0
		for _, e := range logs.All() {
			if strings.Contains(e.Message, report) {
				n++
			}
		}
		return n
	}

	core, logs := observer.New(zapcore.ErrorLevel)
	out, err := RenderTemplateWithMap(needsToken, map[string]interface{}{"title": "x"}, zap.New(core))
	if err != nil {
		t.Fatalf("this seam renders a missing map key as <no value> and strips it — no error expected: %v", err)
	}
	if got := count(logs); got != 1 {
		t.Fatalf("an unbound {{.InstanceID}} through the SECOND render path must be "+
			"reported exactly once, got %d — this path has no RenderContext to "+
			"publish onto, so the log IS its whole error surface (bugs_open/283, "+
			"council 661bcf00 round 2 edit 2)", got)
	}
	// And the damage the report names must actually be in the output.
	if !strings.Contains(out, `id=""`) {
		t.Errorf("CONTROL FAILED: the unbound render should carry an EMPTY id after "+
			"<no value> stripping, got: %s", out)
	}

	// CONTROL 1 — token supplied: renders with the token, no report. Without
	// this the assertion above would pass against a report that fires always.
	core2, logs2 := observer.New(zapcore.ErrorLevel)
	bound, err := RenderTemplateWithMap(needsToken,
		map[string]interface{}{"title": "x", InstanceContentKey: InstanceToken("faq", 0)}, zap.New(core2))
	if err != nil {
		t.Fatalf("CONTROL FAILED: a bound render must succeed: %v", err)
	}
	if !strings.Contains(bound, `id="c-faq"`) {
		t.Errorf("CONTROL FAILED: bound render lost the token: %s", bound)
	}
	if got := count(logs2); got != 0 {
		t.Errorf("CONTROL FAILED: a bound render must not be reported, got %d", got)
	}

	// CONTROL 2 — a template that never spells the token must not be reported,
	// or every contact-info render logs an error nobody can act on.
	core3, logs3 := observer.New(zapcore.ErrorLevel)
	if _, err := RenderTemplateWithMap(`<div class="contact"><p>{{.email}}</p></div>`,
		map[string]interface{}{"email": "x@y.z"}, zap.New(core3)); err != nil {
		t.Fatalf("CONTROL FAILED: plain template must render: %v", err)
	}
	if got := count(logs3); got != 0 {
		t.Errorf("CONTROL FAILED: a token-free template must not be reported, got %d", got)
	}
}
