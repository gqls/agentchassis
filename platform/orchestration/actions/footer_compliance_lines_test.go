// FILE: platform/orchestration/actions/footer_compliance_lines_test.go
//
// The chrome-level carrier for every-page invariants (portfolio_positioning
// seam 1, HANDOFF_2026-08-02). The shared footer-theme-chrome component gains
// a {{if .compliance_lines}}-gated block whose value arrives through the
// input_schema gap-fill (render_site_components → sourceResolver →
// config.chrome.compliance_lines, i.e. site_specs aspect site_config). This
// is the SECOND consumer of the registered STY-050 mechanism (per-site chrome
// config via a gated input_schema field) — the mechanism is prior art, the
// vocabulary key is new. The template lives in the DB (content_components
// e6347680-4c7c-448b-8cfc-1cea509159d1, shared by all 14 live sites' footer
// slots as of 2026-08-02), so this test pins the one property that makes the
// shared edit safe rather than the template itself:
//
//	a site that has NOT set compliance_lines renders BYTE-IDENTICALLY
//	under the new template.
//
// The block is deliberately inline on the copyright line and carries its CSS
// inside the gate — a rule added to the template's shared <style> block would
// change every site's rendered chrome on its next re-render, which is exactly
// what the gate exists to prevent.
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// footerThemeChromeOld is the html_template of footer-theme-chrome as stored
// live on 2026-08-02 (1,577 bytes, no trailing newline). The DB row is the
// authority; this copy exists so the byte-identity property stays provable in
// CI after the row is edited.
const footerThemeChromeOld = `<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand">
            <h3>{{.logo_text}}</h3>
            {{if .tagline}}<p>{{.tagline}}</p>{{end}}
        </div>
        {{if .quick_links_html}}<div class="footer-links">
            <h4>Quick Links</h4>
            <ul>
                {{.quick_links_html}}
            </ul>
        </div>{{end}}
        {{if .services_html}}<div class="footer-services">
            <h4>Explore</h4>
            <ul>
                {{.services_html}}
            </ul>
        </div>{{end}}
        {{if or .email .phone}}<div class="footer-contact">
            <h4>Contact</h4>
            {{if .email}}<p><a href="mailto:{{.email}}">{{.email}}</a></p>{{end}}
            {{if .phone}}<p>{{.phone}}</p>{{end}}
        </div>{{end}}
    </div>
    <div class="footer-bottom">
        <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>
        {{if .legal_links}}<div class="footer-legal">
            {{range .legal_links}}<a href="{{.url}}">{{.name}}</a>{{end}}
        </div>{{end}}
    </div>
</footer>
<style>
/* Theme-owned chrome — var()-based gaps only; the layout styles .site-footer,
   .footer-container and .footer-bottom. */
.footer-brand p { color: var(--color-footer-text, var(--color-text-muted)); margin: 0.5rem 0 0; }
.footer-legal { display: flex; gap: 1rem; flex-wrap: wrap; justify-content: center; margin-top: 0.5rem; }
.footer-legal a { color: var(--color-footer-text, var(--color-text-muted)); }
.footer-legal a:hover { color: var(--color-accent); }
</style>`

// footerThemeChromeNew is footerThemeChromeOld plus the gated compliance
// block. The {{if}} opens at the end of the copyright line and the {{end}}
// closes immediately before that line's original newline, so the gated-out
// render reuses the old bytes exactly — no residue line, no trim markers
// (trim markers would be misparsed by the regex fallback renderer).
const footerThemeChromeNew = `<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand">
            <h3>{{.logo_text}}</h3>
            {{if .tagline}}<p>{{.tagline}}</p>{{end}}
        </div>
        {{if .quick_links_html}}<div class="footer-links">
            <h4>Quick Links</h4>
            <ul>
                {{.quick_links_html}}
            </ul>
        </div>{{end}}
        {{if .services_html}}<div class="footer-services">
            <h4>Explore</h4>
            <ul>
                {{.services_html}}
            </ul>
        </div>{{end}}
        {{if or .email .phone}}<div class="footer-contact">
            <h4>Contact</h4>
            {{if .email}}<p><a href="mailto:{{.email}}">{{.email}}</a></p>{{end}}
            {{if .phone}}<p>{{.phone}}</p>{{end}}
        </div>{{end}}
    </div>
    <div class="footer-bottom">
        <p>&copy; {{.year}} {{.company_name}}. All rights reserved.</p>{{if .compliance_lines}}
        <div class="footer-compliance">{{range .compliance_lines}}<p>{{.}}</p>{{end}}</div>
        <style>.footer-compliance { margin-top: 0.5rem; } .footer-compliance p { color: var(--color-footer-text, var(--color-text-muted)); font-size: 0.85rem; margin: 0.25rem 0; }</style>{{end}}
        {{if .legal_links}}<div class="footer-legal">
            {{range .legal_links}}<a href="{{.url}}">{{.name}}</a>{{end}}
        </div>{{end}}
    </div>
</footer>
<style>
/* Theme-owned chrome — var()-based gaps only; the layout styles .site-footer,
   .footer-container and .footer-bottom. */
.footer-brand p { color: var(--color-footer-text, var(--color-text-muted)); margin: 0.5rem 0 0; }
.footer-legal { display: flex; gap: 1rem; flex-wrap: wrap; justify-content: center; margin-top: 0.5rem; }
.footer-legal a { color: var(--color-footer-text, var(--color-text-muted)); }
.footer-legal a:hover { color: var(--color-accent); }
</style>`

// footerRenderCtx mirrors the ContentData vocabulary render_site_components
// builds for the footer slot. compliance_lines is added per-case, the way the
// input_schema gap-fill would add it.
func footerRenderCtx(extra map[string]interface{}) *RenderContext {
	cd := map[string]interface{}{
		"company_name":     "Example Ltd",
		"logo_text":        "Example",
		"tagline":          "Plain answers first",
		"year":             "2026",
		"email":            "hello@example.co.uk",
		"phone":            "",
		"quick_links_html": `<li><a href="/about.html">About</a></li>`,
		"services_html":    "",
		"legal_links": []map[string]interface{}{
			{"name": "Privacy", "url": "/privacy.html"},
		},
	}
	for k, v := range extra {
		cd[k] = v
	}
	return &RenderContext{
		CompanyName: "Example Ltd",
		Domain:      "example.co.uk",
		Year:        "2026",
		ContentData: cd,
	}
}

func TestFooterComplianceUnsetRendersByteIdentical(t *testing.T) {
	logger := zap.NewNop()

	oldOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeOld, footerRenderCtx(nil), logger)
	newOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeNew, footerRenderCtx(nil), logger)

	if oldOut != newOut {
		t.Fatalf("unset compliance_lines must render byte-identically.\nold (%d bytes):\n%s\nnew (%d bytes):\n%s",
			len(oldOut), oldOut, len(newOut), newOut)
	}
	if strings.Contains(newOut, "footer-compliance") {
		t.Fatalf("gated block leaked into an unset render:\n%s", newOut)
	}
}

func TestFooterComplianceEmptyArrayRendersByteIdentical(t *testing.T) {
	logger := zap.NewNop()

	oldOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeOld, footerRenderCtx(nil), logger)
	newOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeNew,
		footerRenderCtx(map[string]interface{}{"compliance_lines": []interface{}{}}), logger)

	if oldOut != newOut {
		t.Fatalf("empty compliance_lines must gate out exactly like unset.\nold:\n%s\nnew:\n%s", oldOut, newOut)
	}
}

func TestFooterComplianceLinesRenderOnSetSites(t *testing.T) {
	logger := zap.NewNop()

	lines := []interface{}{
		"This site does not lend, never will, takes no applications and does no lead generation.",
		"We are independent of, and not affiliated with, the Financial Conduct Authority.",
	}
	out, _, _ := RenderTemplateReportingMissing(footerThemeChromeNew,
		footerRenderCtx(map[string]interface{}{"compliance_lines": lines}), logger)

	if !strings.Contains(out, `<div class="footer-compliance">`) {
		t.Fatalf("compliance block missing from a set render:\n%s", out)
	}
	for _, l := range lines {
		if !strings.Contains(out, "<p>"+l.(string)+"</p>") {
			t.Fatalf("compliance line %q missing from render:\n%s", l, out)
		}
	}
	if !strings.Contains(out, ".footer-compliance {") {
		t.Fatalf("gated CSS missing from a set render:\n%s", out)
	}
	// The block must not have displaced its neighbours.
	if !strings.Contains(out, "All rights reserved.</p>") || !strings.Contains(out, `class="footer-legal"`) {
		t.Fatalf("neighbouring footer-bottom content damaged:\n%s", out)
	}
}

// A wrong-typed value (string instead of array) must not take the footer down.
// text/template errors on {{range}} over a string, which drops the whole
// template to the regex fallback renderer — degraded output is acceptable,
// an empty or panicking footer is not. The schema documents array-of-strings;
// this pins the blast radius of getting that wrong.
func TestFooterComplianceWrongTypeDoesNotDestroyFooter(t *testing.T) {
	logger := zap.NewNop()

	out, _, _ := RenderTemplateReportingMissing(footerThemeChromeNew,
		footerRenderCtx(map[string]interface{}{"compliance_lines": "not-an-array"}), logger)

	if !strings.Contains(out, "All rights reserved.") {
		t.Fatalf("footer destroyed by a wrong-typed compliance_lines value:\n%s", out)
	}
}

// The declared-type guard (council 56ab6e23, bug_historian advisory): a
// non-array under {{range}} errors the whole template into the silent regex
// fallback, so the fill is refused instead — the gated block renders absent
// and the rest of the chrome renders normally. Only array/list are enforced:
// measured 2026-08-02, every array/list-declared schema field fleet-wide is
// {{range}}-consumed (53) or unreferenced (16), zero bare-output.
func TestResolvedValueSatisfiesDeclaredType(t *testing.T) {
	cases := []struct {
		declared string
		value    interface{}
		want     bool
	}{
		{"array", []interface{}{"a", "b"}, true},
		{"array", []interface{}{}, true},
		{"array", "not-an-array", false},
		{"array", map[string]interface{}{"k": "v"}, false},
		{"array", 3.14, false},
		{"list", "scalar", false},
		{"list", []interface{}{"x"}, true},
		{"text", "anything", true},
		{"text", []interface{}{"even this"}, true},
		{"url", "/x.html", true},
		{"", "untyped fields pass", true},
		{"unknown-type", 42, true},
	}
	for _, c := range cases {
		if got := resolvedValueSatisfiesDeclaredType(c.declared, c.value); got != c.want {
			t.Fatalf("declared=%q value=%T: got %v, want %v", c.declared, c.value, got, c.want)
		}
	}
}
