-- =====================================================================
-- LAYOUT SEED: docs-sidebar
-- =====================================================================
-- Character: reference-grade. Fixed left nav, anchored main content,
--            right-side table-of-contents, code-friendly.
-- Mapped themes: none.
-- Default typography: mono-technical (IBM Plex Mono for code, system-ui
--                     for body).
-- Default header/footer: header-docs (new), existing minimal.
--
-- STRUCTURAL DIVERGENCE:
--   - 3-zone CSS grid: [sidebar-left 260px] [main flex] [toc-right 200px]
--     This is the top-level shape of <main>, not a section-local grid.
--   - Sidebar is independently scrollable and sticky
--   - Right TOC disappears below 1280px; sidebar collapses below 1024px
--     behind a hamburger
--   - Main content has its own max-width (780px) inside the flex zone
--     with generous line-height (1.75)
--   - Code blocks: surface background, left accent border, horizontal
--     scroll, copy button positioning, language badge
--   - Admonitions: .callout / .callout--info / .callout--warn /
--     .callout--danger with left accent border + icon affordance
--   - Headings are anchored (scroll-padding-top offsets the sticky
--     header); hover reveals a ¶ anchor link
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
    'docs-sidebar',
    'Docs — Reference',
    'Three-zone documentation layout with fixed left sidebar nav, centred reading column, right-side table-of-contents, styled code blocks, admonition callouts, and anchored headings. Suits developer documentation, API references, knowledge bases, technical guides.',
    'docs',
    ARRAY['developer-docs', 'api-reference', 'knowledge-base', 'technical-guide'],
    '{
        "container_max_width": "1440px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "3rem",
        "section_padding_y_mobile": "2rem",
        "sidebar_width": "260px",
        "toc_width": "220px",
        "main_max_width": "780px",
        "zone_gap": "2.5rem",
        "header_height": "60px",
        "line_height_reading": "1.75",
        "border_radius": "0.375rem",
        "border_radius_sm": "0.25rem",
        "code_line_height": "1.6",
        "transition_base": "150ms ease"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: docs-sidebar
 *
 * Grammar: fixed-width left nav, reading column, right-side TOC.
 * Optimised for long-form technical reading with inline code.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#0284c7"}};
  --color-primary-hover:  {{palette "primary_hover"  "#0369a1"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#475569"}};
  --color-accent:         {{palette "accent"         "#0284c7"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f8fafc"}};
  --color-text:           {{palette "text"           "#1e293b"}};
  --color-text-muted:     {{palette "text_muted"     "#64748b"}};
  --color-border:         {{palette "border"         "#e2e8f0"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#1e293b"}};
  --color-cta-bg:         {{palette "cta_bg"         "#0284c7"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#f8fafc"}};
  --color-footer-text:    {{palette "footer_text"    "#64748b"}};
  --color-code-bg:        {{palette "code_bg"        "#f1f5f9"}};
  --color-code-text:      {{palette "code_text"      "#0f172a"}};
  --color-code-border:    {{palette "code_border"    "#cbd5e1"}};
  --color-info:           {{palette "info"           "#0284c7"}};
  --color-warn:           {{palette "warn"           "#d97706"}};
  --color-danger:         {{palette "danger"         "#dc2626"}};
  --color-success:        {{palette "success"        "#16a34a"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-mono:        {{typo "mono_font"    "'IBM Plex Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"}};
  --font-size-base:   {{typo "base_size"    "15px"}};
  --line-height-base: {{typo "line_height"  "1.75"}};

  /* ── Structure ── */
  --container-max:      {{token "container_max_width"      "1440px"}};
  --container-pad-x:    {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:      {{token "section_padding_y"        "3rem"}};
  --section-pad-y-sm:   {{token "section_padding_y_mobile" "2rem"}};
  --sidebar-w:          {{token "sidebar_width"            "260px"}};
  --toc-w:              {{token "toc_width"                "220px"}};
  --main-max:           {{token "main_max_width"           "780px"}};
  --zone-gap:           {{token "zone_gap"                 "2.5rem"}};
  --header-h:           {{token "header_height"            "60px"}};
  --lh-reading:         {{token "line_height_reading"      "1.75"}};
  --radius:             {{token "border_radius"            "0.375rem"}};
  --radius-sm:          {{token "border_radius_sm"         "0.25rem"}};
  --code-lh:            {{token "code_line_height"         "1.6"}};
  --transition:         {{token "transition_base"          "150ms ease"}};

  {{with palette "heading" ""}}--section-heading: {{.}};{{end}}
}

/* ── Base reset ── */
*, *::before, *::after { box-sizing: border-box; }
html {
  -webkit-text-size-adjust: 100%;
  scroll-padding-top: calc(var(--header-h) + 1rem);
  scroll-behavior: smooth;
}
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
  line-height: 1.3;
  font-weight: 600;
  letter-spacing: -0.01em;
  scroll-margin-top: calc(var(--header-h) + 1rem);
  position: relative;
}
h1 { font-size: 2rem; margin: 0 0 1rem; }
h2 { font-size: 1.5rem; margin: 2.5rem 0 1rem; padding-top: 0.5rem; border-top: 1px solid var(--color-border); }
h3 { font-size: 1.25rem; margin: 2rem 0 0.75rem; }
h4 { font-size: 1.0625rem; margin: 1.5rem 0 0.5rem; }

/* Anchor link affordance — ¶ appears on heading hover */
.heading-anchor {
  position: absolute;
  left: -1.5rem;
  top: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  font-size: 0.85em;
  color: var(--color-text-muted);
  opacity: 0;
  transition: opacity var(--transition);
  text-decoration: none;
}
h1:hover .heading-anchor,
h2:hover .heading-anchor,
h3:hover .heading-anchor,
h4:hover .heading-anchor { opacity: 1; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
ul, ol { padding-left: 1.5rem; margin: 0 0 1rem; }
li { margin-bottom: 0.375rem; }

a {
  color: var(--color-primary);
  text-decoration: none;
  border-bottom: 1px solid color-mix(in srgb, var(--color-primary) 30%, transparent);
  transition: border-color var(--transition), color var(--transition);
}
a:hover {
  color: var(--color-primary-hover);
  border-bottom-color: var(--color-primary);
}

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── The defining 3-zone shape ── */
.docs-grid {
  display: grid;
  grid-template-columns: var(--sidebar-w) minmax(0, 1fr) var(--toc-w);
  gap: var(--zone-gap);
  max-width: var(--container-max);
  margin-inline: auto;
  padding: var(--section-pad-y) var(--container-pad-x);
  align-items: start;
}

/* ── Sidebar navigation (left zone) ── */
.docs-sidebar {
  position: sticky;
  top: calc(var(--header-h) + 1.5rem);
  max-height: calc(100vh - var(--header-h) - 3rem);
  overflow-y: auto;
  padding-right: 0.5rem;
  font-size: 0.9375rem;
}
.docs-sidebar__section {
  margin-bottom: 1.5rem;
}
.docs-sidebar__heading {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
  margin-bottom: 0.5rem;
  padding: 0 0.5rem;
}
.docs-sidebar ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
.docs-sidebar li { margin: 0; }
.docs-sidebar a {
  display: block;
  padding: 0.375rem 0.5rem;
  color: var(--color-text);
  border: none;
  border-left: 2px solid transparent;
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  font-weight: 400;
  min-height: 36px;
  line-height: 1.4;
}
.docs-sidebar a:hover {
  background: var(--color-surface);
  color: var(--color-primary);
}
.docs-sidebar a.is-active {
  color: var(--color-primary);
  border-left-color: var(--color-primary);
  font-weight: 500;
  background: var(--color-surface);
}
.docs-sidebar .nested {
  padding-left: 0.75rem;
  margin-top: 0.125rem;
}

/* ── Main content (centre zone) ── */
.docs-main {
  min-width: 0;
  max-width: var(--main-max);
  line-height: var(--lh-reading);
}

/* ── Right-side table of contents ── */
.docs-toc {
  position: sticky;
  top: calc(var(--header-h) + 1.5rem);
  max-height: calc(100vh - var(--header-h) - 3rem);
  overflow-y: auto;
  font-size: 0.8125rem;
  padding-left: 1rem;
  border-left: 1px solid var(--color-border);
}
.docs-toc__heading {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
  margin-bottom: 0.5rem;
}
.docs-toc ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
.docs-toc a {
  display: block;
  padding: 0.25rem 0;
  color: var(--color-text-muted);
  border: none;
  line-height: 1.4;
}
.docs-toc a:hover { color: var(--color-text); }
.docs-toc a.is-active {
  color: var(--color-primary);
  font-weight: 500;
}
.docs-toc .nested {
  padding-left: 0.75rem;
}

/* ── Site header ── */
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
  backdrop-filter: blur(8px);
  background: color-mix(in srgb, var(--color-header-bg) 90%, transparent);
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
  border-bottom: none;
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
  border: none;
  padding: 0.25rem 0;
}
.main-nav a:hover { color: var(--color-primary); }
.header-search {
  max-width: 280px;
  flex: 1;
  position: relative;
}
.header-search input {
  width: 100%;
  height: 36px;
  padding: 0 0.75rem;
  font: inherit;
  font-size: 0.875rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
}
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

/* ── Hero (doc landing only — a summary band, not a sales hero) ── */
.hero-section {
  padding: 2.5rem 0 1.5rem;
  border-bottom: 1px solid var(--color-border);
}
.hero-section .container { max-width: 960px; }
.hero-section h1 { margin-bottom: 0.5rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  max-width: 640px;
  margin: 0 0 1.25rem;
  line-height: 1.5;
}
.hero-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

/* ── Code blocks ── */
pre {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  line-height: var(--code-lh);
  background: var(--color-code-bg);
  color: var(--color-code-text);
  border: 1px solid var(--color-code-border);
  border-left: 3px solid var(--color-primary);
  border-radius: var(--radius);
  padding: 1rem 1.25rem;
  overflow-x: auto;
  margin: 1rem 0 1.5rem;
  position: relative;
}
pre[data-lang]::before {
  content: attr(data-lang);
  position: absolute;
  top: 0.5rem;
  right: 0.75rem;
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-text-muted);
  font-family: var(--font-body);
}
code {
  font-family: var(--font-mono);
  font-size: 0.875em;
}
:not(pre) > code {
  padding: 0.15em 0.4em;
  background: var(--color-code-bg);
  border: 1px solid var(--color-code-border);
  border-radius: 3px;
  font-size: 0.85em;
  word-break: break-word;
}

/* ── Tables ── */
.docs-main table {
  width: 100%;
  margin: 1rem 0 1.5rem;
  border-collapse: collapse;
  font-size: 0.9375rem;
}
.docs-main table th,
.docs-main table td {
  padding: 0.625rem 0.875rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  vertical-align: top;
}
.docs-main table th {
  background: var(--color-surface);
  font-weight: 600;
  font-size: 0.875rem;
}
.docs-main table tr:nth-child(even) { background: var(--color-surface); }
.table-responsive {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  margin: 1rem -0.5rem 1.5rem;
  padding: 0 0.5rem;
}

/* ── Admonitions / callouts ── */
.callout {
  border-left: 4px solid var(--color-info);
  background: color-mix(in srgb, var(--color-info) 5%, var(--color-background));
  padding: 1rem 1.25rem;
  border-radius: 0 var(--radius) var(--radius) 0;
  margin: 1.5rem 0;
}
.callout__title {
  font-weight: 700;
  font-size: 0.9375rem;
  margin-bottom: 0.375rem;
  display: flex;
  align-items: center;
  gap: 0.375rem;
  color: var(--color-info);
}
.callout__title::before { content: "ⓘ"; }
.callout p:last-child { margin-bottom: 0; }
.callout--warn {
  border-left-color: var(--color-warn);
  background: color-mix(in srgb, var(--color-warn) 7%, var(--color-background));
}
.callout--warn .callout__title { color: var(--color-warn); }
.callout--warn .callout__title::before { content: "⚠"; }
.callout--danger {
  border-left-color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 7%, var(--color-background));
}
.callout--danger .callout__title { color: var(--color-danger); }
.callout--danger .callout__title::before { content: "✕"; }
.callout--success {
  border-left-color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 7%, var(--color-background));
}
.callout--success .callout__title { color: var(--color-success); }
.callout--success .callout__title::before { content: "✓"; }

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Section grids — quieter on docs */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1.25rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.25rem;
}
.feature-card h3,
.service-card h3,
.differentiator-card h3 { margin-top: 0; font-size: 1rem; }

/* About / Contact / FAQ  */
.about-section .container,
.contact-section .container,
.faq-section .container { max-width: var(--main-max); }

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
.faq-item summary::after { content: "+"; color: var(--color-text-muted); font-weight: 400; }
.faq-item[open] summary::after { content: "−"; }

/* CTA / testimonials */
.call-to-action-section { text-align: center; padding-block: 2rem; }
.testimonials-section { padding-block: 2rem; }

/* Forms */
.form-field { margin-bottom: 1rem; }
.form-field label {
  display: block;
  font-weight: 600;
  margin-bottom: 0.375rem;
  font-size: 0.875rem;
}
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
  min-height: 44px;
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
}

/* Buttons — restrained, link-like by default */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font: inherit;
  font-weight: 500;
  font-size: 0.875rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), border-color var(--transition), color var(--transition);
  border-bottom: none;
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover { background: var(--color-primary-hover); color: var(--color-primary-text); border-bottom: none; }
.btn-secondary {
  background: var(--color-background);
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-border);
}
.btn-secondary:hover { border-color: var(--color-primary); }

/* Site footer */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  border-top: 1px solid var(--color-border);
  padding: 2rem 0;
  margin-top: auto;
  font-size: 0.875rem;
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 1rem;
}
.site-footer a { color: var(--color-footer-text); border-bottom: none; }
.site-footer a:hover { color: var(--color-text); }
.site-footer ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

/* ── Responsive ── */
@media (max-width: 1280px) {
  .docs-grid {
    grid-template-columns: var(--sidebar-w) minmax(0, 1fr);
  }
  .docs-toc { display: none; }
}
@media (max-width: 1024px) {
  .docs-grid {
    grid-template-columns: minmax(0, 1fr);
    gap: 1rem;
  }
  .docs-sidebar {
    position: static;
    max-height: none;
    border-bottom: 1px solid var(--color-border);
    padding-bottom: 1rem;
    margin-bottom: 1rem;
  }
  .docs-sidebar.is-collapsed { display: none; }
  .header-search { display: none; }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-header-bg); border-bottom: 1px solid var(--color-border); padding: 0.5rem var(--container-pad-x); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .mobile-menu-toggle { display: inline-flex; }
}

/* Accessibility */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
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
