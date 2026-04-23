-- =====================================================================
-- LAYOUT SEED: tool-first-landing
-- =====================================================================
-- Character: the tool IS the page. Not contained — dominant. First
--            fold is 80% tool, minimal heading, split-pane input/output
--            common. Everything else is supporting material below.
-- Mapped themes: none.
-- Default typography: sans-modern (Inter).
-- Default header/footer: header-minimal-tool (new), existing minimal.
--
-- DIVERGENCE from utility-tool (layout 5):
--   - utility-tool: 800px narrow column; the tool is a contained
--     card in a page with a small hero above it
--   - THIS: full-container width (up to 1400px); tool dominates 80%+
--     of the viewport above the fold; supporting content is below
--     via scroll
--   - Defining primitive: .split-pane with left input / right output,
--     50/50 default (configurable via CSS vars to 40/60 etc.)
--   - Dark-mode-friendly — code-heavy tools often want dark output
--   - Optional tabbed interface for tools with modes
--   - Minimal heading (one line or none) — the UI affords the purpose
-- =====================================================================

INSERT INTO layouts (
    name,
    display_name,
    description,
    category,
    industry_tags,
    structure_tokens,
    css_template,
    origin,
    is_active
) VALUES (
    'tool-first-landing',
    'Tool-First — Landing',
    'Full-container tool-dominated layout where the page IS the tool. Above-the-fold is 80%+ tool interface with optional split-pane input/output. Supporting sections (how-it-works, docs) sit below the fold. Dark-mode-friendly for code output. Suits single-purpose calculators, API playgrounds, demo tools, configuration generators.',
    'tool',
    ARRAY['calculator', 'playground', 'demo-tool', 'configurator', 'developer-tool', 'saas-tool'],
    '{
        "container_max_width": "1400px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "3rem",
        "section_padding_y_mobile": "2rem",
        "tool_viewport_height": "min(80vh, 900px)",
        "split_pane_left": "50%",
        "split_pane_right": "50%",
        "split_pane_gap": "1px",
        "border_radius": "0.5rem",
        "border_radius_sm": "0.25rem",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.05)",
        "shadow_md": "0 4px 12px rgba(0,0,0,0.08)",
        "transition_base": "150ms ease",
        "pane_padding": "1.25rem",
        "header_height": "56px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: tool-first-landing
 *
 * Grammar: tool first, then everything else. Container-wide split
 * pane is the hero. Supporting sections are deliberately quiet.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#7c3aed"}};
  --color-primary-hover:  {{palette "primary_hover"  "#6d28d9"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#475569"}};
  --color-accent:         {{palette "accent"         "#7c3aed"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f8fafc"}};
  --color-text:           {{palette "text"           "#0f172a"}};
  --color-text-muted:     {{palette "text_muted"     "#64748b"}};
  --color-border:         {{palette "border"         "#e2e8f0"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#0f172a"}};
  --color-cta-bg:         {{palette "cta_bg"         "#7c3aed"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#f8fafc"}};
  --color-footer-text:    {{palette "footer_text"    "#64748b"}};

  /* Tool-specific palette — input pane light, output pane dark by
     default (common in developer tools). Themes can override to make
     both light or both dark as needed. */
  --color-input-pane-bg:    {{palette "input_pane_bg"    "#ffffff"}};
  --color-input-pane-text:  {{palette "input_pane_text"  "#0f172a"}};
  --color-output-pane-bg:   {{palette "output_pane_bg"   "#0f172a"}};
  --color-output-pane-text: {{palette "output_pane_text" "#e2e8f0"}};
  --color-pane-divider:     {{palette "pane_divider"     "#e2e8f0"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-mono:        {{typo "mono_font"    "'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"}};
  --font-size-base:   {{typo "base_size"    "15px"}};
  --line-height-base: {{typo "line_height"  "1.55"}};

  /* ── Structure ── */
  --container-max:         {{token "container_max_width"      "1400px"}};
  --container-pad-x:       {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:         {{token "section_padding_y"        "3rem"}};
  --section-pad-y-sm:      {{token "section_padding_y_mobile" "2rem"}};
  --tool-vh:               {{token "tool_viewport_height"     "min(80vh, 900px)"}};
  --split-left:            {{token "split_pane_left"          "50%"}};
  --split-right:           {{token "split_pane_right"         "50%"}};
  --split-gap:             {{token "split_pane_gap"           "1px"}};
  --radius:                {{token "border_radius"            "0.5rem"}};
  --radius-sm:             {{token "border_radius_sm"         "0.25rem"}};
  --shadow-sm:             {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.05)"}};
  --shadow-md:             {{token "shadow_md"                "0 4px 12px rgba(0,0,0,0.08)"}};
  --transition:            {{token "transition_base"          "150ms ease"}};
  --pane-pad:              {{token "pane_padding"             "1.25rem"}};
  --header-h:              {{token "header_height"            "56px"}};

  {{with palette "heading" ""}}--section-heading: {{.}};{{end}}
}

/* ── Base reset ── */
*, *::before, *::after { box-sizing: border-box; }
html { -webkit-text-size-adjust: 100%; }
body {
  margin: 0;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  font-family: var(--font-body);
  font-size: var(--font-size-base);
  line-height: var(--line-height-base);
  color: var(--color-text);
  background: var(--color-background);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }

/* ── Colour Inheritance Model ── */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-primary));
  margin: 0 0 0.75rem;
  line-height: 1.25;
  font-weight: 600;
  letter-spacing: -0.015em;
}
h1 { font-size: clamp(1.5rem, 2.5vw, 2rem); font-weight: 700; }
h2 { font-size: 1.375rem; }
h3 { font-size: 1.125rem; }
h4 { font-size: 1rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
code, pre { font-family: var(--font-mono); }
a {
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary-hover); text-decoration: underline; }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── Site header — compact, minimal ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 1000;
  height: var(--header-h);
  display: flex;
  align-items: center;
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  width: 100%;
}
.logo {
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--color-header-text);
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.logo-img { max-height: 28px; width: auto; }
.main-nav ul {
  display: flex;
  gap: 1.25rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 500;
  font-size: 0.875rem;
  padding: 0.25rem 0;
}
.main-nav a:hover,
.main-nav a.active { color: var(--color-primary); text-decoration: none; }
.mobile-menu-toggle {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
  min-width: 44px;
  min-height: 44px;
  color: var(--color-header-text);
}
.mobile-menu-toggle span {
  display: block;
  width: 20px;
  height: 2px;
  background: currentColor;
  margin: 4px 0;
}

/* ── The tool IS the hero ── */
.tool-hero {
  padding-top: 1rem;
}
.tool-heading {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0.75rem var(--container-pad-x) 1.25rem;
}
.tool-heading h1 {
  margin: 0;
  font-size: clamp(1.375rem, 2vw, 1.75rem);
}
.tool-heading p {
  margin: 0.375rem 0 0;
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* The tool region — container-wide, tall */
.tool-region {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
}

/* ── Split pane — the defining primitive ── */
.split-pane {
  display: grid;
  grid-template-columns: var(--split-left) var(--split-right);
  gap: var(--split-gap);
  background: var(--color-pane-divider);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
  min-height: var(--tool-vh);
  box-shadow: var(--shadow-md);
}
/* Configurable split ratios via modifier classes */
.split-pane--40-60 { grid-template-columns: 40% 60%; }
.split-pane--60-40 { grid-template-columns: 60% 40%; }
.split-pane--30-70 { grid-template-columns: 30% 70%; }

/* Single-pane variant when there's no output pane */
.tool-region.single-pane .split-pane {
  grid-template-columns: 1fr;
  min-height: calc(var(--tool-vh) * 0.9);
}

.pane {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}
.pane-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.625rem var(--pane-pad);
  font-size: 0.8125rem;
  font-weight: 500;
  border-bottom: 1px solid var(--color-border);
  flex: 0 0 auto;
}
.pane-header__title {
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  font-size: 0.6875rem;
  font-weight: 700;
}
.pane-header__actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}
.pane-body {
  flex: 1;
  overflow: auto;
  padding: var(--pane-pad);
  min-height: 0;
}

/* Input pane */
.pane--input {
  background: var(--color-input-pane-bg);
  color: var(--color-input-pane-text);
}
.pane--input .pane-header { border-bottom-color: var(--color-border); }

/* Output pane — dark-default */
.pane--output {
  background: var(--color-output-pane-bg);
  color: var(--color-output-pane-text);
}
.pane--output .pane-header {
  background: color-mix(in srgb, var(--color-output-pane-bg) 90%, #ffffff);
  border-bottom-color: rgba(255,255,255,0.08);
}
.pane--output .pane-header__title { color: rgba(255,255,255,0.6); }
.pane--output .pane-body {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}
.pane--output .icon-btn { color: rgba(255,255,255,0.7); }
.pane--output .icon-btn:hover { color: #ffffff; background: rgba(255,255,255,0.08); }

/* Light-output variant */
.pane--output.pane--light {
  background: var(--color-surface);
  color: var(--color-text);
}

/* Icon-only button affordance (copy, clear, download, settings) */
.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  min-height: 32px;
  padding: 0 0.375rem;
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: inherit;
  cursor: pointer;
  font-size: 0.75rem;
  transition: background var(--transition), border-color var(--transition);
}
.icon-btn:hover {
  background: var(--color-surface);
  border-color: var(--color-border);
}

/* ── Tabs affordance (for tools with modes) ── */
.tool-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 1rem;
}
.tool-tab {
  padding: 0.625rem 1rem;
  font-weight: 500;
  font-size: 0.875rem;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  min-height: 44px;
  transition: color var(--transition), border-color var(--transition);
}
.tool-tab:hover { color: var(--color-text); }
.tool-tab.is-active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

/* ── Forms inside panes ── */
.pane input:not([type="checkbox"]):not([type="radio"]),
.pane textarea,
.pane select,
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.5rem 0.75rem;
  font: inherit;
  font-size: 0.9375rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 40px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.pane textarea { font-family: var(--font-mono); font-size: 0.875rem; min-height: 200px; resize: vertical; }
.pane input:focus,
.pane textarea:focus,
.pane select:focus,
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.form-field { margin-bottom: 1rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.375rem;
  font-size: 0.8125rem;
}

/* ── Below-fold sections — supporting, quiet ── */
.hero-section {
  /* Re-used as a below-tool intro strip when present. Small. */
  padding: 2rem 0;
  border-top: 1px solid var(--color-border);
  margin-top: var(--section-pad-y);
  text-align: center;
}
.hero-section .container { max-width: 720px; }
.hero-subtitle, .hero-section .lead {
  font-size: 1rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Features (rendered as "How it works" typically) — horizontal steps */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
  margin-top: 1.5rem;
  counter-reset: step;
}
.feature-card,
.service-card,
.differentiator-card {
  counter-increment: step;
  padding: 1.5rem 1.25rem;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  position: relative;
}
.feature-card::before,
.service-card::before,
.differentiator-card::before {
  content: counter(step, decimal-leading-zero);
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--color-primary);
  letter-spacing: 0.1em;
  margin-bottom: 0.75rem;
}
.feature-card h3,
.service-card h3,
.differentiator-card h3 {
  font-size: 1rem;
  margin-bottom: 0.375rem;
}
.feature-card p,
.service-card p,
.differentiator-card p {
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0;
}

/* About */
.about-section .container { max-width: 720px; }

/* FAQ */
.faq-section .container { max-width: 780px; }
.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1rem 0;
}
.faq-item summary {
  cursor: pointer;
  font-weight: 600;
  list-style: none;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-height: 44px;
  align-items: center;
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; color: var(--color-text-muted); }
.faq-item[open] summary::after { content: "−"; }

/* CTA / testimonials */
.call-to-action-section { text-align: center; padding-block: 2.5rem; }
.testimonials-section { padding-block: 2.5rem; }
.testimonials-section .testimonial {
  max-width: 640px;
  margin-inline: auto;
  text-align: center;
  font-size: 1rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* Contact */
.contact-section .container { max-width: 640px; }

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.125rem;
  font: inherit;
  font-weight: 500;
  font-size: 0.9375rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 40px;
  transition: background var(--transition), color var(--transition), border-color var(--transition);
}
.btn:hover { text-decoration: none; }
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover {
  background: var(--color-primary-hover);
  color: var(--color-primary-text);
}
.btn-secondary {
  background: transparent;
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-border);
}
.btn-secondary:hover {
  border-color: var(--color-primary);
  background: var(--color-surface);
}
.btn-run {
  background: var(--color-primary);
  color: var(--color-primary-text);
  font-weight: 600;
  min-height: 44px;
}

/* ── Site footer — minimal one-line ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  border-top: 1px solid var(--color-border);
  padding: 1.5rem 0;
  margin-top: auto;
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  font-size: 0.8125rem;
}
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-primary); text-decoration: none; }
.site-footer ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .split-pane,
  .split-pane--40-60,
  .split-pane--60-40,
  .split-pane--30-70 {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto;
    min-height: auto;
  }
  .pane { min-height: 400px; }
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .split-pane { border-radius: var(--radius-sm); }
  .pane { min-height: 320px; padding-left: 0; padding-right: 0; }
  .pane-body { padding: 0.75rem; }
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: 1fr; }
  .tool-tabs { overflow-x: auto; -webkit-overflow-scrolling: touch; scrollbar-width: none; }
  .tool-tabs::-webkit-scrollbar { display: none; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-header-bg); border-bottom: 1px solid var(--color-border); padding: 0.5rem var(--container-pad-x); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .mobile-menu-toggle { display: inline-flex; }
  .footer-container { flex-direction: column; text-align: center; }
}

/* Accessibility */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
$LAYOUT$,
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name     = EXCLUDED.display_name,
    description      = EXCLUDED.description,
    category         = EXCLUDED.category,
    industry_tags    = EXCLUDED.industry_tags,
    structure_tokens = EXCLUDED.structure_tokens,
    css_template     = EXCLUDED.css_template,
    is_active        = EXCLUDED.is_active,
    updated_at       = NOW();
