// FILE: platform/orchestration/actions/chrome_note_and_cta_override_test.go
//
// Two more per-site chrome config carriers on the STY-050 mechanism (gated
// input_schema field → sourceResolver → config.* → site_specs aspect
// site_config), following footer_compliance_lines_test.go (STY-051), whose
// old/new constants this file chains from.
//
// Why these exist (bugs_open/146 §"the trap that shaped the oufe fix", and the
// 2026-08-08 finding in oufe/HANDOFF_2026-07-30_continue_here.md): oufe.com
// carried two hand-patches that lived ONLY in the stored site_components
// artefact — the footer honesty note (fallibility disclosure, mig 268's
// protected object) and the header CTA rewrite ("Get Started"→/contact.html
// replaced by "Read the cases"→/cases/index.html, FIX_2026-07-26). The chrome
// re-render of 2026-07-31 19:21 rebuilt both slots from the shared templates
// and silently reverted BOTH. A stored artefact is one legitimate rebuild from
// reset; these two fields move the content into the template+config path so a
// rebuild REPRODUCES it instead of deleting it.
//
//   - footer_note (STY-052): a per-site plain-text disclosure band rendered
//     between the footer link columns and the footer-bottom bar.
//     config.chrome.footer_note.
//   - header_cta_url + header_cta_label (STY-053): a per-site override of the
//     header CTA. The fixed vocabulary always supplies non-empty cta_text
//     ("Get Started") and a resolved cta_url, and the schema gap-fill is
//     gap-fill ONLY (presence wins), so the override needs NEW field names the
//     fixed vocabulary never sets; the template prefers them when BOTH are
//     present. config.chrome.header_cta_url / config.chrome.header_cta_label.
//     NOTE the override bypasses chromeLinks.Allows (which vets the default
//     cta_url at vocabulary-build time) — correct-or-absent is the operator's
//     duty for this key, stated in the schema description.
//
// The templates live in the DB (content_components e6347680… footer, 16 sites;
// 58fde68f… header, 15 sites — counted 2026-08-08), so what this file pins is
// the one property that makes a shared-template edit safe:
//
//	a site that has NOT set the new config renders BYTE-IDENTICALLY
//	under the new template.
//
// Both gated blocks carry their CSS inside the gate (the footer note) or reuse
// the existing .header-cta class (the CTA override) — nothing is added to the
// shared <style> blocks, which would change every site's chrome on its next
// re-render.
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// footerThemeChromeWithNote is footerThemeChromeNew (the live row, md5
// eea3fb6911cacc97f56a98ba8d68bba6 as of 2026-08-08) plus the gated
// footer_note band. The {{if}} opens at the end of the footer-container's
// closing-div line and the {{end}} closes immediately before that line's
// original newline, so the gated-out render reuses the old bytes exactly.
// The band's CSS is mig 253's final .footer-note form (full container width,
// rule above), carried inside the gate.
const footerThemeChromeWithNote = `<footer class="site-footer">
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
    </div>{{if .footer_note}}
    <div class="footer-note">
        <p>{{.footer_note}}</p>
        <style>.footer-note { max-width: 1200px; margin: 2rem auto 0; padding: 1.5rem 2rem; border-top: 1px solid rgba(255,255,255,0.15); } .footer-note p { color: var(--color-footer-text, var(--color-text-muted)); font-size: 0.85rem; margin: 0; }</style>
    </div>{{end}}
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

// headerThemeChromeLive is the html_template of header-theme-chrome as stored
// live on 2026-08-08 (2,551 bytes, no trailing newline, md5
// 0aae8077d9be27df8fef428b54561396). The DB row is the authority; this copy
// exists so the byte-identity property stays provable in CI after the row is
// edited.
const headerThemeChromeLive = `{{if .gtm_container_id}}<!-- Google Tag Manager (noscript) -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id={{.gtm_container_id}}"
height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
<!-- End Google Tag Manager (noscript) -->{{end}}
<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">
            {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}
        </a>
        <nav class="main-nav">
            <ul>
                {{if .nav_items_html}}{{.nav_items_html}}{{end}}
            </ul>
        </nav>
        {{if .cta_url}}<a href="{{.cta_url}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>{{end}}
        <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">
            <span></span><span></span><span></span>
        </button>
    </div>
</header>
<style>
/* Theme-owned chrome: every colour is a CSS variable resolved by the site
   stylesheet. Only gaps the layout does not style are covered here. */
.header-cta {
    background: var(--color-cta-bg, var(--color-accent));
    color: var(--color-cta-text, var(--color-primary-text));
    padding: 0.5rem 1.1rem;
    border-radius: var(--radius, 4px);
    text-decoration: none;
    font-weight: 600;
    font-size: 0.9rem;
    white-space: nowrap;
}
.header-cta:hover { filter: brightness(1.1); }
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: var(--color-header-text, var(--color-text));
    margin: 5px 0;
}
@media (max-width: 768px) {
    .main-nav.is-open {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: var(--color-header-bg, var(--color-surface));
        padding: 1rem;
        border-bottom: 1px solid var(--color-border);
    }
    .main-nav.is-open ul { flex-direction: column; }
    .main-nav.is-open a { display: block; padding: 0.75rem 0; }
    .header-cta { display: none; }
}
</style>
<script>
document.addEventListener("DOMContentLoaded", function() {
    var toggle = document.querySelector(".mobile-menu-toggle");
    var nav = document.querySelector(".main-nav");
    if (toggle && nav) {
        toggle.addEventListener("click", function() {
            var open = nav.classList.toggle("is-open");
            toggle.setAttribute("aria-expanded", open ? "true" : "false");
        });
    }
});
</script>`

// headerThemeChromeCTAOverride is headerThemeChromeLive with the CTA line
// replaced by an override-preferring form. The {{else}} branch carries the old
// CTA action text byte-for-byte, so a site with neither override key renders
// exactly the old bytes; the override branch fires only when BOTH url and
// label are present (an anchor with an empty label, or an unlabelled URL, is
// worse than the default).
const headerThemeChromeCTAOverride = `{{if .gtm_container_id}}<!-- Google Tag Manager (noscript) -->
<noscript><iframe src="https://www.googletagmanager.com/ns.html?id={{.gtm_container_id}}"
height="0" width="0" style="display:none;visibility:hidden"></iframe></noscript>
<!-- End Google Tag Manager (noscript) -->{{end}}
<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">
            {{if .logo_url}}<img src="{{.logo_url}}" alt="{{.logo_text}}" class="logo-img">{{else}}<span class="logo-text">{{.logo_text}}</span>{{end}}
        </a>
        <nav class="main-nav">
            <ul>
                {{if .nav_items_html}}{{.nav_items_html}}{{end}}
            </ul>
        </nav>
        {{if and .header_cta_url .header_cta_label}}<a href="{{.header_cta_url}}" class="header-cta">{{.header_cta_label}}</a>{{else}}{{if .cta_url}}<a href="{{.cta_url}}" class="header-cta">{{if .cta_text}}{{.cta_text}}{{else}}Get Started{{end}}</a>{{end}}{{end}}
        <button class="mobile-menu-toggle" aria-label="Toggle menu" aria-expanded="false">
            <span></span><span></span><span></span>
        </button>
    </div>
</header>
<style>
/* Theme-owned chrome: every colour is a CSS variable resolved by the site
   stylesheet. Only gaps the layout does not style are covered here. */
.header-cta {
    background: var(--color-cta-bg, var(--color-accent));
    color: var(--color-cta-text, var(--color-primary-text));
    padding: 0.5rem 1.1rem;
    border-radius: var(--radius, 4px);
    text-decoration: none;
    font-weight: 600;
    font-size: 0.9rem;
    white-space: nowrap;
}
.header-cta:hover { filter: brightness(1.1); }
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: var(--color-header-text, var(--color-text));
    margin: 5px 0;
}
@media (max-width: 768px) {
    .main-nav.is-open {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        background: var(--color-header-bg, var(--color-surface));
        padding: 1rem;
        border-bottom: 1px solid var(--color-border);
    }
    .main-nav.is-open ul { flex-direction: column; }
    .main-nav.is-open a { display: block; padding: 0.75rem 0; }
    .header-cta { display: none; }
}
</style>
<script>
document.addEventListener("DOMContentLoaded", function() {
    var toggle = document.querySelector(".mobile-menu-toggle");
    var nav = document.querySelector(".main-nav");
    if (toggle && nav) {
        toggle.addEventListener("click", function() {
            var open = nav.classList.toggle("is-open");
            toggle.setAttribute("aria-expanded", open ? "true" : "false");
        });
    }
});
</script>`

// oufeFooterNote is the owner-approved wording (oufe
// DRAFT_disclaimer_for_owner_approval.md §A), the exact value the migration
// seeds into oufe's config.chrome.footer_note.
const oufeFooterNote = "OUFE publishes educational analysis of financial and legal mechanism. We make mistakes, and some of what is here is assembled with AI assistance that can invent convincing detail. Check anything that matters against the primary source. Nothing here is investment advice or a recommendation."

// headerRenderCtx mirrors the ContentData vocabulary render_site_components
// builds for the header slot: cta_text is ALWAYS the non-empty default and
// cta_url is the resolved contact page — which is exactly why the override
// needs new field names (gap-fill skips present, non-empty values).
func headerRenderCtx(extra map[string]interface{}) *RenderContext {
	cd := map[string]interface{}{
		"company_name":   "Example Ltd",
		"logo_text":      "Example",
		"logo_url":       "/assets/images/logo.jpg",
		"nav_items_html": `<li><a href="/index.html">Home</a></li><li><a href="/contact.html">Contact</a></li>`,
		"cta_text":       "Get Started",
		"cta_url":        "/contact.html",
		"year":           "2026",
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

func TestFooterNoteUnsetRendersByteIdentical(t *testing.T) {
	logger := zap.NewNop()

	oldOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeNew, footerRenderCtx(nil), logger)
	newOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeWithNote, footerRenderCtx(nil), logger)

	if oldOut != newOut {
		t.Fatalf("unset footer_note must render byte-identically.\nold (%d bytes):\n%s\nnew (%d bytes):\n%s",
			len(oldOut), oldOut, len(newOut), newOut)
	}
	if strings.Contains(newOut, "footer-note") {
		t.Fatalf("gated block leaked into an unset render:\n%s", newOut)
	}
}

// An empty string can reach ContentData even though navigateMap refuses to
// resolve one — a fixed-vocabulary writer or a future caller could put it
// there. {{if}} on "" is false, so it gates out exactly like unset.
func TestFooterNoteEmptyStringRendersByteIdentical(t *testing.T) {
	logger := zap.NewNop()

	oldOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeNew, footerRenderCtx(nil), logger)
	newOut, _, _ := RenderTemplateReportingMissing(footerThemeChromeWithNote,
		footerRenderCtx(map[string]interface{}{"footer_note": ""}), logger)

	if oldOut != newOut {
		t.Fatalf("empty footer_note must gate out exactly like unset.\nold:\n%s\nnew:\n%s", oldOut, newOut)
	}
}

// A set site renders the note band — and it must coexist with a set
// compliance_lines (both blocks live in footer-theme-chrome; a site may
// legitimately carry both).
func TestFooterNoteRendersOnSetSites(t *testing.T) {
	logger := zap.NewNop()

	lines := []interface{}{"This site does not lend."}
	out, _, _ := RenderTemplateReportingMissing(footerThemeChromeWithNote,
		footerRenderCtx(map[string]interface{}{
			"footer_note":      oufeFooterNote,
			"compliance_lines": lines,
		}), logger)

	if !strings.Contains(out, `<div class="footer-note">`) {
		t.Fatalf("footer-note block missing from a set render:\n%s", out)
	}
	if !strings.Contains(out, "<p>"+oufeFooterNote+"</p>") {
		t.Fatalf("note text missing from render:\n%s", out)
	}
	if !strings.Contains(out, ".footer-note {") {
		t.Fatalf("gated CSS missing from a set render:\n%s", out)
	}
	// Neighbours intact: the band sits between the link columns and the
	// footer-bottom bar, displacing neither, and the compliance block still
	// renders.
	if !strings.Contains(out, "All rights reserved.</p>") || !strings.Contains(out, `class="footer-compliance"`) {
		t.Fatalf("neighbouring footer content damaged:\n%s", out)
	}
}

func TestHeaderCTAOverrideUnsetRendersByteIdentical(t *testing.T) {
	logger := zap.NewNop()

	oldOut, _, _ := RenderTemplateReportingMissing(headerThemeChromeLive, headerRenderCtx(nil), logger)
	newOut, _, _ := RenderTemplateReportingMissing(headerThemeChromeCTAOverride, headerRenderCtx(nil), logger)

	if oldOut != newOut {
		t.Fatalf("unset header CTA override must render byte-identically.\nold (%d bytes):\n%s\nnew (%d bytes):\n%s",
			len(oldOut), oldOut, len(newOut), newOut)
	}
	if !strings.Contains(newOut, `class="header-cta">Get Started</a>`) {
		t.Fatalf("default CTA lost from an unset render:\n%s", newOut)
	}
}

// URL without label (or label without URL) must NOT half-fire: the override
// branch is gated on both, so a partial config renders the default exactly.
func TestHeaderCTAOverridePartialConfigRendersByteIdentical(t *testing.T) {
	logger := zap.NewNop()

	oldOut, _, _ := RenderTemplateReportingMissing(headerThemeChromeLive, headerRenderCtx(nil), logger)
	for _, partial := range []map[string]interface{}{
		{"header_cta_url": "/cases/index.html"},
		{"header_cta_label": "Read the cases"},
	} {
		newOut, _, _ := RenderTemplateReportingMissing(headerThemeChromeCTAOverride, headerRenderCtx(partial), logger)
		if oldOut != newOut {
			t.Fatalf("partial override %v must render byte-identically.\nold:\n%s\nnew:\n%s", partial, oldOut, newOut)
		}
	}
}

func TestHeaderCTAOverrideRendersOnSetSites(t *testing.T) {
	logger := zap.NewNop()

	out, _, _ := RenderTemplateReportingMissing(headerThemeChromeCTAOverride,
		headerRenderCtx(map[string]interface{}{
			"header_cta_url":   "/cases/index.html",
			"header_cta_label": "Read the cases",
		}), logger)

	if !strings.Contains(out, `<a href="/cases/index.html" class="header-cta">Read the cases</a>`) {
		t.Fatalf("override CTA missing from a set render:\n%s", out)
	}
	// The default CTA must be fully displaced — not rendered alongside.
	if strings.Contains(out, `class="header-cta">Get Started</a>`) {
		t.Fatalf("default CTA rendered alongside the override:\n%s", out)
	}
	// The nav's own /contact.html link is untouched — only the CTA moved.
	if !strings.Contains(out, `<li><a href="/contact.html">Contact</a></li>`) {
		t.Fatalf("nav damaged by the CTA override:\n%s", out)
	}
}

// The override must work on a site whose default CTA is ABSENT (no contact
// page ⇒ empty cta_url ⇒ the old template renders no CTA at all). The
// override is independent of the fixed vocabulary's resolution.
func TestHeaderCTAOverrideRendersWhenDefaultCTAAbsent(t *testing.T) {
	logger := zap.NewNop()

	out, _, _ := RenderTemplateReportingMissing(headerThemeChromeCTAOverride,
		headerRenderCtx(map[string]interface{}{
			"cta_url":          "",
			"header_cta_url":   "/cases/index.html",
			"header_cta_label": "Read the cases",
		}), logger)

	if !strings.Contains(out, `<a href="/cases/index.html" class="header-cta">Read the cases</a>`) {
		t.Fatalf("override CTA must render even when the default CTA is absent:\n%s", out)
	}
}
