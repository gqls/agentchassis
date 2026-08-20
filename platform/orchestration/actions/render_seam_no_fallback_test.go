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
	"strings"
	"testing"

	"go.uber.org/zap"
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
	out, err := RenderTemplate(mechanismFlowLikeTemplate,
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
	out, err := RenderTemplate(mechanismFlowLikeTemplate,
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

	out, err := RenderTemplate(`<h2>{{.heading}}</h2><p>{{.body}}</p><pre>{{.snippet}}</pre>`, ctx, zap.NewNop())
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
	out, err := RenderTemplate(broken, &RenderContext{
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
	out, err := RenderTemplate(`<nav>{{nav_items_html}}</nav>`, &RenderContext{
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
	out, err := RenderTemplate("", &RenderContext{}, zap.NewNop())
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
	if _, err := RenderTemplate(tmpl, &RenderContext{
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
