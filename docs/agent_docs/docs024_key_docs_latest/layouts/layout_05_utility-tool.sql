-- =====================================================================
-- LAYOUT SEED: utility-tool
-- =====================================================================
-- Character: the tool is the reason. Minimal chrome, centered narrow
--            column, large form controls, distinct output region,
--            secondary content below the fold.
-- Mapped themes: none — exists for selector/adoption matching.
-- Default typography: sans-modern (Inter).
-- Default header/footer: header-minimal-tool (new), existing minimal footer.
--
-- STRUCTURAL DIVERGENCE:
--   - Main content area capped at 800px — narrower than every other
--     layout. The tool gets attention; nothing else competes.
--   - Header is compact (52px tall) with just brand + sparse links
--   - Hero is almost nothing: one-line title, one-line description,
--     straight into the tool area
--   - Tool area is a single card with form controls and a distinct
--     output region beneath
--   - No card-grids, no 3-col sections — this layout refuses to be
--     a brochure
--   - "Supporting" sections below the fold (how-to, about, related
--     tools) are smaller and lower-contrast than the tool itself
--   - Form controls are LARGER than other layouts (48px min input
--     height) because this is where the user actually works
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
    'utility-tool',
    'Utility — Tool-First',
    'Minimal-chrome layout for single-purpose utilities. The tool area dominates a narrow centred column (800px) with large form controls, a distinct output region, and deliberately understated supporting sections below. Suits online calculators, unit converters, developer utilities, data validators.',
    'utility',
    ARRAY['calculator', 'converter', 'validator', 'developer-tool', 'saas-utility'],
    '{
        "container_max_width": "800px",
        "container_max_width_wide": "1040px",
        "container_padding_x": "1.25rem",
        "section_padding_y": "2.5rem",
        "section_padding_y_mobile": "1.5rem",
        "tool_padding": "2rem",
        "tool_output_padding": "1.5rem",
        "input_min_height": "48px",
        "border_radius": "0.5rem",
        "border_radius_sm": "0.25rem",
        "shadow_sm": "0 1px 3px rgba(0,0,0,0.06)",
        "shadow_md": "0 4px 12px rgba(0,0,0,0.08)",
        "transition_base": "150ms ease",
        "mono_font": "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace",
        "header_height": "52px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: utility-tool
 *
 * Grammar: narrow centred column. The tool card is the body.
 * Everything else is quiet.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#2563eb"}};
  --color-primary-hover:  {{palette "primary_hover"  "#1d4ed8"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#475569"}};
  --color-accent:         {{palette "accent"         "#2563eb"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f8fafc"}};
  --color-text:           {{palette "text"           "#1e293b"}};
  --color-text-muted:     {{palette "text_muted"     "#64748b"}};
  --color-border:         {{palette "border"         "#e2e8f0"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#1e293b"}};
  --color-cta-bg:         {{palette "cta_bg"         "#2563eb"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#f8fafc"}};
  --color-footer-text:    {{palette "footer_text"    "#64748b"}};
  --color-output-bg:      {{palette "output_bg"      "#f1f5f9"}};
  --color-success:        {{palette "success"        "#16a34a"}};
  --color-error:          {{palette "error"          "#dc2626"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-mono:        {{typo "mono_font"    "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.5"}};

  /* ── Structure ── */
  --container-max:         {{token "container_max_width"      "800px"}};
  --container-max-wide:    {{token "container_max_width_wide" "1040px"}};
  --container-pad-x:       {{token "container_padding_x"      "1.25rem"}};
  --section-pad-y:         {{token "section_padding_y"        "2.5rem"}};
  --section-pad-y-sm:      {{token "section_padding_y_mobile" "1.5rem"}};
  --tool-pad:              {{token "tool_padding"             "2rem"}};
  --tool-output-pad:       {{token "tool_output_padding"      "1.5rem"}};
  --input-min-h:           {{token "input_min_height"         "48px"}};
  --radius:                {{token "border_radius"            "0.5rem"}};
  --radius-sm:             {{token "border_radius_sm"         "0.25rem"}};
  --shadow-sm:             {{token "shadow_sm"                "0 1px 3px rgba(0,0,0,0.06)"}};
  --shadow-md:             {{token "shadow_md"                "0 4px 12px rgba(0,0,0,0.08)"}};
  --transition:            {{token "transition_base"          "150ms ease"}};
  --header-h:              {{token "header_height"            "52px"}};

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
  font-variant-numeric: tabular-nums;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }

/* Monospace numbers inside the tool — deliberate, not body-wide. */
.tool output,
.tool-output,
.result-value,
.mono { font-family: var(--font-mono); }

/* ── Colour Inheritance Model ── */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-primary));
  margin: 0 0 0.75rem;
  line-height: 1.25;
  font-weight: 600;
  letter-spacing: -0.01em;
}
h1 { font-size: 1.75rem; font-weight: 700; }
h2 { font-size: 1.375rem; }
h3 { font-size: 1.125rem; }
h4 { font-size: 1rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
code, pre {
  font-family: var(--font-mono);
  font-size: 0.9em;
}
pre {
  padding: 1rem;
  background: var(--color-output-bg);
  border-radius: var(--radius-sm);
  overflow-x: auto;
}
code:not(pre code) {
  padding: 0.15em 0.4em;
  background: var(--color-output-bg);
  border-radius: 3px;
  font-size: 0.875em;
}
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
.container-wide {
  max-width: var(--container-max-wide);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── Site header — compact, single line ── */
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
  max-width: var(--container-max-wide);
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
.logo:hover { text-decoration: none; }
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

/* ── Hero — tiny, just a title + one line ── */
.hero-section {
  padding: 2.5rem 0 1rem;
}
.hero-section .container { text-align: center; }
.hero-section h1 {
  margin-bottom: 0.5rem;
}
.hero-subtitle, .hero-section .lead {
  font-size: 1rem;
  color: var(--section-text-muted, var(--color-text-muted));
  max-width: 560px;
  margin: 0 auto 0;
}
.hero-actions { display: none; }

/* ── Tool area — the main event ──
 * Opinionated: a single card with form controls + output region.
 * Components named .tool or .utility-tool should target these classes. */
.tool-section {
  padding-block: 1.5rem 3rem;
}
.tool {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--tool-pad);
  box-shadow: var(--shadow-sm);
}
.tool h2 {
  margin-bottom: 1.5rem;
  font-size: 1.25rem;
}
.tool__inputs {
  display: grid;
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.tool__inputs.cols-2 { grid-template-columns: 1fr 1fr; }
.tool__inputs.cols-3 { grid-template-columns: repeat(3, 1fr); }
.tool__actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  padding-top: 0.5rem;
  border-top: 1px solid var(--color-border);
}
.tool__output {
  margin-top: 1.5rem;
  padding: var(--tool-output-pad);
  background: var(--color-output-bg);
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  font-size: 1rem;
  min-height: 120px;
}
.tool__output--code {
  font-family: var(--font-mono);
  font-size: 0.9rem;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}
.tool__output--empty {
  color: var(--section-text-muted, var(--color-text-muted));
  display: flex;
  align-items: center;
  justify-content: center;
}
.result-label {
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.25rem;
}
.result-value {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--section-heading, var(--color-primary));
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}
.result-positive { color: var(--color-success); }
.result-negative { color: var(--color-error); }

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 relocates these to components. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* All these "below the fold" sections are restrained — lower contrast,
   tighter padding, smaller type. The tool above got attention; these
   supplement. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section {
  padding-block: 2.5rem;
}

/* How-to / related-tools / features use a simple two-column at most */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
  margin-top: 1.5rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 1.25rem;
}
.feature-card h3,
.service-card h3,
.differentiator-card h3 {
  font-size: 1rem;
  margin-bottom: 0.5rem;
}
.feature-card p,
.service-card p,
.differentiator-card p {
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0;
}

/* ── About — compact */
.about-section .container p { max-width: 65ch; }

/* ── FAQ — compact accordion */
.faq-section .container { max-width: var(--container-max); }
.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1rem 0;
}
.faq-item summary {
  cursor: pointer;
  font-weight: 600;
  font-size: 0.9375rem;
  list-style: none;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-height: 44px;
  align-items: center;
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after {
  content: "+";
  color: var(--section-text-muted, var(--color-text-muted));
  font-weight: 400;
}
.faq-item[open] summary::after { content: "−"; }
.faq-item p {
  padding-top: 0.5rem;
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── CTA and testimonials omitted intentionally — not idiomatic for
   utility tools. Components rendering these classes still work but
   layout gives them no special positioning. ── */
.call-to-action-section { text-align: center; padding-block: 2rem; }
.testimonials-section { padding-block: 2rem; }
.testimonials-section .testimonial {
  max-width: 600px;
  margin-inline: auto;
  text-align: center;
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 0.5rem;
  font-style: normal;
}

/* ── Contact ── */
.contact-section .container { max-width: var(--container-max); }

/* ── Forms — large, clear, focus-obvious ── */
.form-field { margin-bottom: 1.25rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.375rem;
  font-size: 0.875rem;
}
.form-field .hint {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
}
.form-field input,
.form-field textarea,
.form-field select,
.tool input,
.tool textarea,
.tool select {
  width: 100%;
  padding: 0.75rem 0.875rem;
  font: inherit;
  font-size: 0.9375rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-background);
  color: var(--color-text);
  min-height: var(--input-min-h);
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input[type="number"],
.tool input[type="number"] {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus,
.tool input:focus,
.tool textarea:focus,
.tool select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
}
.tool fieldset {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 1rem;
  margin: 0 0 1rem;
}
.tool legend {
  padding: 0 0.5rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--section-text-muted, var(--color-text-muted));
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* ── Buttons ── */
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
  min-height: 44px;
  transition: background var(--transition), border-color var(--transition),
              color var(--transition);
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover {
  background: var(--color-primary-hover);
  color: var(--color-primary-text);
  text-decoration: none;
}
.btn-secondary {
  background: transparent;
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-border);
}
.btn-secondary:hover {
  background: var(--color-surface);
  border-color: var(--color-primary);
  text-decoration: none;
}

/* ── Site footer — minimal, single line ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  border-top: 1px solid var(--color-border);
  padding: 1.5rem 0;
  margin-top: auto;
}
.footer-container {
  max-width: var(--container-max-wide);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  justify-content: space-between;
  align-items: center;
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
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .tool { padding: 1.25rem; }
  .tool__inputs.cols-2,
  .tool__inputs.cols-3 { grid-template-columns: 1fr; }
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open {
    display: block;
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    background: var(--color-header-bg);
    border-bottom: 1px solid var(--color-border);
    padding: 1rem var(--container-pad-x);
  }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .main-nav.is-open a { display: block; padding: 0.75rem 0; border-bottom: 1px solid var(--color-border); }
  .mobile-menu-toggle { display: inline-flex; }
  .footer-container { flex-direction: column; text-align: center; }
}

/* ── Accessibility ── */
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
