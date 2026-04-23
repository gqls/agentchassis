-- 007_seed_layouts_tool_portal_and_social_lobby.sql
--
-- Seed two new layouts to close the library gap that caused gamesdesign.co.uk
-- (and every interactive-platform or social-network site that follows) to
-- fall back to brochure-formal.
--
-- Root cause reminder: resolveLayoutByTags matches
--   tagSet = {classification.category} ∪ {classification.industry_tags}
-- against layout.industry_tags. The classifier doesn't currently emit those
-- two fields, so tagSet is empty → fallback path "no classification tags".
-- This migration makes the LIBRARY ready. Migration 008 will update the
-- classifier prompt to emit category + industry_tags so these layouts can
-- actually get picked. Neither migration alone is sufficient — both are
-- needed for end-to-end resolution. 007 first because the tag values in
-- 008 need to match what's landed here.
--
-- Layouts:
--   1. tool-portal-dark — dark developer-utility portal with index of tools
--      + guides + games sections, narrow reading column for articles, flat
--      technical aesthetic. Target: gamesdesign.co.uk and similar.
--   2. social-lobby — light, colour-forward social platform layout with
--      room/lobby metaphor, provocation cards as primary UI unit, reaction
--      bars, archetype displays. Target: vonc.com and similar.
--
-- Both templates follow the renderer contract:
--   - Use {{palette}} / {{typo}} / {{token}} template helpers
--   - Do NOT declare --section-* defaults (renderer appends)
--   - The 5 renderer-managed surface classes (.features-section,
--     .services-section, .differentiators-section, .about-section,
--     .faq-section) are surface-coloured
--   - Element rules use var(--section-*, var(--color-*)) pattern
--
-- ----------------------------------------------------------------------------
-- Rollback (undo this migration):
--   DELETE FROM layouts WHERE name IN ('tool-portal-dark','social-lobby');
-- Any sites that had composition installed against these layouts would need
-- needs_composition re-queued after rollback.
-- ----------------------------------------------------------------------------

BEGIN;

-- ---------------------------------------------------------------
-- Pre-check: make sure these names aren't already taken
-- ---------------------------------------------------------------

SELECT 'BEFORE' AS phase, name
FROM layouts
WHERE name IN ('tool-portal-dark','social-lobby');


-- ============================================================================
-- LAYOUT 1: tool-portal-dark
-- ============================================================================

INSERT INTO layouts (
    name,
    display_name,
    description,
    css_template,
    structure_tokens,
    category,
    industry_tags,
    is_active,
    origin,
    needs_review
) VALUES (
             'tool-portal-dark',
             'Tool Portal — Dark',
             'Dark developer-utility portal with index pages for tools, guides, and playable prototypes. The tool page layout gives the interactive element full breathing room; the article layout provides a narrow reading column with code-block and callout support. Flat technical aesthetic, dense information layout, monospace-friendly data presentation. Suits game design tool platforms, developer utility hubs, calculator libraries, technical reference sites, and practitioner-focused interactive platforms that sit alongside an IDE rather than replacing a marketing site.',
             $template$
/* =====================================================================
 * LAYOUT: tool-portal-dark
 *
 * Target shape: interactive platform whose primary value is a library of
 * tools plus a library of theory guides plus (optionally) playable
 * prototypes. Dark-mode-first because the audience is developers and
 * designers working in dark IDEs. The layout supports THREE page shapes:
 *
 *   1. Portal/index pages  — hero + grid cards for tools/guides/games
 *   2. Tool pages          — tool area dominates; supporting prose sits
 *                            below. Form controls are first-class.
 *   3. Article/guide pages — narrow reading column (720px), generous
 *                            line-height, table + code + callout support.
 *
 * Renderer contract (see brochure-formal for the same notes):
 *   - This template MUST NOT declare --section-* defaults; the renderer
 *     appends them after rendering based on palette luminance.
 *   - Renderer-managed surface section classes:
 *       .features-section, .services-section, .differentiators-section,
 *       .about-section, .faq-section
 *     These MUST be surface-coloured here.
 *   - Element rules use var(--section-*, var(--color-*)) so dark-section
 *     components override per container without restating rules.
 * ===================================================================== */

                 :root {
  /* ── Palette — dark-first fallbacks ── */
  --color-primary:        {{palette "primary"      "#00bcd4"}};
  --color-primary-hover:  {{palette "primary_hover" "#26d0e2"}};
  --color-primary-text:   {{palette "primary_text"  "#0b0b0b"}};
  --color-secondary:      {{palette "secondary"     "#1e1e1e"}};
  --color-accent:         {{palette "accent"        "#00bcd4"}};
  --color-background:     {{palette "background"    "#121212"}};
  --color-surface:        {{palette "surface"       "#1e1e1e"}};
  --color-surface-alt:    {{palette "surface_alt"   "#1a1a1a"}};
  --color-text:           {{palette "text"          "#e0e0e0"}};
  --color-text-muted:     {{palette "text_muted"    "#a0a0a0"}};
  --color-border:         {{palette "border"        "#2a2a2a"}};
  --color-card-bg:        {{palette "card_bg"       "#1e1e1e"}};
  --color-header-bg:      {{palette "header_bg"     "#121212"}};
  --color-header-text:    {{palette "header_text"   "#e0e0e0"}};
  --color-cta-bg:         {{palette "cta_bg"        "#1e1e1e"}};
  --color-cta-text:       {{palette "cta_text"      "#e0e0e0"}};
  --color-footer-bg:      {{palette "footer_bg"     "#0d0d0d"}};
  --color-footer-text:    {{palette "footer_text"   "rgba(224,224,224,0.8)"}};
  --color-code-bg:        {{palette "code_bg"       "#0d0d0d"}};
  --color-callout-bg:     {{palette "callout_bg"    "rgba(0,188,212,0.08)"}};
  --color-callout-border: {{palette "callout_border" "#00bcd4"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Segoe UI', Roboto, Helvetica, Arial, sans-serif"}};
  --font-heading:     {{typo "heading_font" "'Segoe UI', Roboto, Helvetica, Arial, sans-serif"}};
  --font-mono:        {{typo "mono_font"    "'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.6"}};

  /* ── Structure — tuned for dense utility feel ── */
  --container-max:    {{token "container_max_width"      "1200px"}};
  --container-pad-x:  {{token "container_padding_x"      "1.25rem"}};
  --reading-max:      {{token "reading_max_width"        "720px"}};
  --section-pad-y:    {{token "section_padding_y"        "4rem"}};
  --section-pad-y-sm: {{token "section_padding_y_mobile" "2.5rem"}};
  --radius:           {{token "border_radius"            "0.25rem"}};
  --radius-sm:        {{token "border_radius_sm"         "2px"}};
  --radius-lg:        {{token "border_radius_lg"         "0.375rem"}};
  --shadow-sm:        {{token "shadow_sm"                "0 1px 0 rgba(0,0,0,0.25)"}};
  --shadow-md:        {{token "shadow_md"                "0 2px 4px rgba(0,0,0,0.35)"}};
  --shadow-lg:        {{token "shadow_lg"                "0 4px 16px rgba(0,0,0,0.5)"}};
  --transition:       {{token "transition_base"          "150ms ease"}};
  --card-pad:         {{token "card_padding"             "1.5rem"}};
  --grid-gap:         {{token "grid_gap"                 "1.5rem"}};

  /* Optional palette-driven heading override (same pattern as other layouts) */
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
code, pre, kbd, samp { font-family: var(--font-mono); }

/* ── Colour Inheritance Model ── */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-text));
  margin: 0 0 1rem;
  line-height: 1.25;
  font-weight: 600;
  letter-spacing: -0.01em;
}
h1 { font-size: 2.25rem; font-weight: 700; }
h2 { font-size: 1.625rem; }
h3 { font-size: 1.25rem; }
h4 { font-size: 1.05rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
a {
  color: var(--color-accent);
  text-decoration: none;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary); }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.container--reading { max-width: var(--reading-max); }
.section { padding-block: var(--section-pad-y); }

/* ── Site header — low chrome, same background as body ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 1000;
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0.875rem var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}
.logo {
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--color-header-text);
  text-decoration: none;
  letter-spacing: -0.01em;
}
.logo-img { max-height: 32px; width: auto; }
.main-nav ul {
  display: flex;
  gap: 1.75rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 500;
  font-size: 0.9rem;
  padding: 0.5rem 0;
}
.main-nav a:hover,
.main-nav a.active { color: var(--color-accent); }
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
  width: 22px;
  height: 2px;
  background: currentColor;
  margin: 5px 0;
}

/* ── Hero (component-coloured) ──
 * Tool-portal heroes are typically punchy statement typography, not
 * imagery-driven. Keep the layout minimal; the component supplies bg. */
.hero-section { padding-block: calc(var(--section-pad-y) * 1.25); }
.hero-section .container { text-align: center; max-width: 880px; }
.hero-section h1 {
  font-size: clamp(2rem, 4.5vw, 3rem);
  margin-bottom: 1rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}
.hero-subtitle, .hero-section .lead {
  font-size: 1.15rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 auto 2rem;
  max-width: 640px;
}
.hero-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Renderer-managed surface sections ──
 * Same contract as brochure-formal: these 5 classes MUST be surface-coloured.
 * The renderer will append --section-* defaults on top of this background. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* ── Generic section grids — denser than brochure (3-col → 2 → 1) ── */
.features-grid,
.services-grid,
.differentiators-grid,
.tools-grid,
.guides-grid,
.games-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: var(--grid-gap);
  margin-top: 2rem;
}
.feature-card,
.service-card,
.differentiator-card,
.tool-card,
.guide-card,
.game-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--card-pad);
  transition: border-color var(--transition), background var(--transition);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover,
.tool-card:hover,
.guide-card:hover,
.game-card:hover {
  border-color: var(--color-accent);
  background: var(--color-surface-alt);
}
.feature-icon,
.service-icon,
.tool-card__icon,
.guide-card__icon,
.game-card__icon {
  width: 32px;
  height: 32px;
  margin-bottom: 0.5rem;
  color: var(--color-accent);
}
.tool-card__title,
.guide-card__title,
.game-card__title { font-size: 1.05rem; font-weight: 600; margin: 0; }
.tool-card__description,
.guide-card__excerpt,
.game-card__description {
  font-size: 0.9rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0;
  flex: 1;
}

/* ── Tool page layout ──
 * The tool itself (form + output) dominates. Supporting prose sits below. */
.tool-section {
  padding-block: var(--section-pad-y);
  background: var(--color-background);
}
.tool-section .container { max-width: 960px; }
.tool-workspace {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: clamp(1.25rem, 3vw, 2rem);
  margin-bottom: 2.5rem;
}
.tool-workspace__header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 1rem;
  margin-bottom: 1.25rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--color-border);
}
.tool-workspace__title { font-size: 1.25rem; margin: 0; }
.tool-workspace__meta {
  font-family: var(--font-mono);
  font-size: 0.8rem;
  color: var(--color-text-muted);
}
.tool-inputs {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.tool-output {
  background: var(--color-code-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.25rem;
  font-family: var(--font-mono);
  font-size: 0.95rem;
  line-height: 1.5;
  color: var(--color-text);
  overflow-x: auto;
}
.tool-output__label {
  display: block;
  font-family: var(--font-body);
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted);
  margin-bottom: 0.5rem;
}
.tool-output__value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-accent);
}

/* ── Article / guide page layout ──
 * Narrow reading column, generous line-height, supports code, tables,
 * callouts, and the Design Rule pattern from the content_direction. */
.article-section {
  padding-block: var(--section-pad-y);
}
.article-section .container { max-width: var(--reading-max); }
.article-body {
  font-size: 1.0625rem;
  line-height: 1.75;
  color: var(--color-text);
}
.article-body h2 {
  font-size: 1.5rem;
  margin-top: 2.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border);
}
.article-body h3 { font-size: 1.15rem; margin-top: 2rem; }
.article-body p { margin-bottom: 1.25rem; }
.article-body ul, .article-body ol { margin: 0 0 1.25rem 1.5rem; }
.article-body li { margin-bottom: 0.4rem; }
.article-body blockquote {
  margin: 1.5rem 0;
  padding: 0.75rem 1.25rem;
  border-left: 3px solid var(--color-border);
  color: var(--color-text-muted);
  font-style: italic;
}
.article-body code {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 0.1em 0.35em;
  font-size: 0.9em;
}
.article-body pre {
  background: var(--color-code-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1rem 1.25rem;
  overflow-x: auto;
  margin: 1.5rem 0;
  font-size: 0.9rem;
  line-height: 1.5;
}
.article-body pre code {
  background: transparent;
  border: none;
  padding: 0;
  font-size: inherit;
}
.article-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5rem 0;
  font-size: 0.95rem;
}
.article-body th, .article-body td {
  padding: 0.625rem 0.875rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
}
.article-body th {
  font-weight: 600;
  color: var(--color-text);
  background: var(--color-surface);
  border-bottom-color: var(--color-border);
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.article-body tbody td { color: var(--color-text-muted); }
.article-body td:first-child, .article-body th:first-child { padding-left: 0; }

/* Design Rule callout — the axiom pattern */
.design-rule {
  background: var(--color-callout-bg);
  border-left: 3px solid var(--color-callout-border);
  padding: 1rem 1.25rem;
  margin: 1.5rem 0;
  border-radius: 0 var(--radius) var(--radius) 0;
}
.design-rule__label {
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-callout-border);
  margin-bottom: 0.35rem;
}
.design-rule__body { color: var(--color-text); margin: 0; font-weight: 500; }

/* ── Forms (tool controls are first-class) ── */
.form-field { margin-bottom: 1rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.3rem;
  font-size: 0.85rem;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.625rem 0.75rem;
  font: inherit;
  font-size: 0.95rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 40px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-accent) 30%, transparent);
}
.form-field--inline input { font-family: var(--font-mono); }

/* ── Buttons — flat, tool-specific verb framing ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  font: inherit;
  font-weight: 600;
  font-size: 0.9rem;
  border-radius: var(--radius);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 40px;
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
}
.btn-secondary {
  background: transparent;
  color: var(--color-accent);
  border-color: var(--color-accent);
}
.btn-secondary:hover {
  background: var(--color-accent);
  color: var(--color-primary-text);
}
.btn-ghost {
  background: transparent;
  color: var(--color-text-muted);
  border-color: transparent;
}
.btn-ghost:hover { color: var(--color-text); background: var(--color-surface); }
.btn-large { padding: 0.875rem 1.75rem; font-size: 0.95rem; }

/* ── FAQ ── */
.faq-section .container { max-width: var(--reading-max); }
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
.faq-item summary::after { content: "+"; color: var(--color-text-muted); font-size: 1.25rem; }
.faq-item[open] summary::after { content: "−"; color: var(--color-accent); }
.faq-item p {
  padding-top: 0.625rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── About / Contact (component-coloured) ── */
.about-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; align-items: start; }
.contact-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; }

/* ── Site footer — even darker than body for visual separation ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding-top: 3rem;
  margin-top: auto;
  border-top: 1px solid var(--color-border);
  font-size: 0.9rem;
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 2rem;
}
.site-footer h3, .site-footer h4 { color: var(--color-text); font-size: 0.95rem; }
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-accent); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.4rem; }
.footer-bottom {
  margin-top: 2.5rem;
  padding: 1.25rem 0;
  border-top: 1px solid var(--color-border);
  text-align: center;
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  h1 { font-size: 1.875rem; }
  h2 { font-size: 1.375rem; }
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; }
  .tool-workspace__header { flex-direction: column; align-items: flex-start; }
}

/* ── Accessibility ── */
:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
$template$::text,
    jsonb_build_object(
        'grid_gap',                  '1.5rem',
        'shadow_lg',                 '0 4px 16px rgba(0,0,0,0.5)',
        'shadow_md',                 '0 2px 4px rgba(0,0,0,0.35)',
        'shadow_sm',                 '0 1px 0 rgba(0,0,0,0.25)',
        'card_padding',              '1.5rem',
        'reading_max_width',         '720px',
        'border_radius',             '0.25rem',
        'transition_base',           '150ms ease',
        'border_radius_lg',          '0.375rem',
        'border_radius_sm',          '2px',
        'section_padding_y',         '4rem',
        'container_max_width',       '1200px',
        'container_padding_x',       '1.25rem',
        'section_padding_y_mobile',  '2.5rem'
    ),
    'interactive',
    ARRAY[
        -- Site-type signals (what classifier.site_type produces)
        'interactive-platform',
        'tools',
        'tool-portal',
        -- Industry / vertical signals (what classifier.industry_tags produces)
        'developer-tools',
        'calculators',
        'utility-platform',
        'practitioner-platform',
        'game-design',
        'game-development',
        'technical-reference',
        'design-tools',
        -- Aesthetic signals (what classifier.suggested_style produces)
        'dark-utility',
        'professional-dark',
        -- Category duplicate (matcher adds classification.category into tagSet)
        'interactive'
    ]::text[],
    true,
    'seed',
    false
);


-- ============================================================================
-- LAYOUT 2: social-lobby
-- ============================================================================

INSERT INTO layouts (
    name,
    display_name,
    description,
    css_template,
    structure_tokens,
    category,
    industry_tags,
    is_active,
    origin,
    needs_review
) VALUES (
    'social-lobby',
    'Social — Lobby',
    'Light, colour-forward social platform layout with room/lobby metaphor as the entry point. Provocation cards are the primary UI unit — prompt + framing + "your take" input + reaction bar. Arena and Stage rooms are visually differentiated via palette slots so the two modes read differently on the lobby. Supports provocation-detail pages, room/topic index pages, and archetype/profile pages. Shareable-card-friendly proportions. Suits AI-seeded social platforms, provocation-driven communities, challenge-based social games, and creator platforms where the content unit is the interaction rather than the post.',
    $template$
/* =====================================================================
 * LAYOUT: social-lobby
 *
 * Target shape: social platform where rooms replace feeds and the core
 * content unit is a provocation card (prompt + AI take + "your take"
 * input + reaction bar). Explicitly NOT dark — light, colour-forward,
 * energetic. The brand primary dominates visually; a secondary accent
 * carries the Arena/Stage differentiation.
 *
 * Pages this layout supports:
 *   1. Lobby (homepage)       — hero framing + grid of active rooms
 *   2. Room / topic index     — grid of rooms or provocations in a topic
 *   3. Provocation detail     — central card + response stack + reactions
 *   4. Archetype / profile    — identity card + history strip
 *
 * Renderer contract:
 *   - This template MUST NOT declare --section-* defaults; the renderer
 *     appends them after rendering based on palette luminance.
 *   - Renderer-managed surface section classes:
 *       .features-section, .services-section, .differentiators-section,
 *       .about-section, .faq-section
 *     These MUST be surface-coloured here.
 *   - Hero / lobby / provocation / cta / testimonials are component-
 *     coloured; the layout supplies spacing and typography only.
 * ===================================================================== */

:root {
  /* ── Palette — light, colour-forward fallbacks ──
   * Fallback hue is a saturated violet-magenta to express "colour-forward
   * brand dominates" when a site hasn't yet supplied its own palette.
   * In production the adopted/inferred palette will replace every slot. */
  --color-primary:         {{palette "primary"         "#7c3aed"}};
  --color-primary-hover:   {{palette "primary_hover"   "#6d28d9"}};
  --color-primary-text:    {{palette "primary_text"    "#ffffff"}};
  --color-secondary:       {{palette "secondary"       "#ec4899"}};
  --color-accent:          {{palette "accent"          "#ec4899"}};
  --color-arena:           {{palette "arena"           "#ec4899"}};   /* pulsing competitive */
  --color-stage:           {{palette "stage"           "#f59e0b"}};   /* glowing creative */
  --color-background:      {{palette "background"      "#fafaf7"}};
  --color-surface:         {{palette "surface"         "#ffffff"}};
  --color-surface-alt:     {{palette "surface_alt"     "#f4f4f0"}};
  --color-text:            {{palette "text"            "#18181b"}};
  --color-text-muted:      {{palette "text_muted"      "#52525b"}};
  --color-border:          {{palette "border"          "#e4e4e7"}};
  --color-card-bg:         {{palette "card_bg"         "#ffffff"}};
  --color-header-bg:       {{palette "header_bg"       "#ffffff"}};
  --color-header-text:     {{palette "header_text"     "#18181b"}};
  --color-cta-bg:          {{palette "cta_bg"          "#7c3aed"}};
  --color-cta-text:        {{palette "cta_text"        "#ffffff"}};
  --color-footer-bg:       {{palette "footer_bg"       "#18181b"}};
  --color-footer-text:     {{palette "footer_text"     "rgba(255,255,255,0.85)"}};
  --color-reaction-pos:    {{palette "reaction_positive" "#10b981"}};   /* Genius, Based */
  --color-reaction-neg:    {{palette "reaction_negative" "#ef4444"}};   /* Delusional, Cursed */
  --color-reaction-meta:   {{palette "reaction_meta"     "#6366f1"}};   /* Suspicious */

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"}};
  --font-heading:     {{typo "heading_font" "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.6"}};

  /* ── Structure — friendly, roomy, shareable-card-friendly ── */
  --container-max:    {{token "container_max_width"      "1200px"}};
  --container-pad-x:  {{token "container_padding_x"      "1.25rem"}};
  --card-max:         {{token "card_max_width"           "680px"}};
  --section-pad-y:    {{token "section_padding_y"        "4rem"}};
  --section-pad-y-sm: {{token "section_padding_y_mobile" "2.5rem"}};
  --radius:           {{token "border_radius"            "0.75rem"}};
  --radius-sm:        {{token "border_radius_sm"         "0.375rem"}};
  --radius-lg:        {{token "border_radius_lg"         "1.25rem"}};
  --shadow-sm:        {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.06)"}};
  --shadow-md:        {{token "shadow_md"                "0 4px 12px rgba(0,0,0,0.08)"}};
  --shadow-lg:        {{token "shadow_lg"                "0 12px 32px rgba(0,0,0,0.12)"}};
  --transition:       {{token "transition_base"          "200ms cubic-bezier(0.4, 0, 0.2, 1)"}};
  --card-pad:         {{token "card_padding"             "1.75rem"}};
  --grid-gap:         {{token "grid_gap"                 "1.5rem"}};

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
  color: var(--section-heading, var(--color-text));
  margin: 0 0 1rem;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: -0.02em;
}
h1 { font-size: clamp(2rem, 5vw, 3rem); }
h2 { font-size: clamp(1.5rem, 3vw, 2rem); }
h3 { font-size: 1.25rem; font-weight: 600; }
h4 { font-size: 1.05rem; font-weight: 600; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
a {
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--transition);
  font-weight: 500;
}
a:hover { color: var(--color-primary-hover); }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.container--card { max-width: var(--card-max); }
.section { padding-block: var(--section-pad-y); }

/* ── Site header — clean white with primary brand accent ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 1000;
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 1rem var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}
.logo {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--color-primary);
  text-decoration: none;
  letter-spacing: -0.02em;
}
.logo-img { max-height: 36px; width: auto; }
.main-nav ul {
  display: flex;
  gap: 2rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 500;
  font-size: 0.95rem;
  padding: 0.5rem 0;
  border-bottom: 2px solid transparent;
}
.main-nav a:hover,
.main-nav a.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
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
  width: 22px;
  height: 2px;
  background: currentColor;
  margin: 5px 0;
}

/* ── Hero (component-coloured) ──
 * Lobby-landing hero: big headline, "what are you playing today?" energy.
 * Component supplies the gradient or colour field; layout provides spacing
 * and typography scale. */
.hero-section { padding-block: calc(var(--section-pad-y) * 1.25); }
.hero-section .container { text-align: center; max-width: 880px; }
.hero-section h1 {
  font-weight: 800;
  margin-bottom: 1.25rem;
}
.hero-subtitle, .hero-section .lead {
  font-size: 1.15rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 auto 2rem;
  max-width: 600px;
  line-height: 1.55;
}
.hero-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Renderer-managed surface sections ── */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface-alt); }

/* ── Lobby: active-rooms grid ──
 * The homepage (after hero) shows active rooms. Arena and Stage use
 * different palette slots to read as different energies. */
.lobby-section { padding-block: var(--section-pad-y); }
.lobby-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--grid-gap);
  margin-top: 2rem;
}
.room-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--card-pad);
  transition: transform var(--transition), box-shadow var(--transition), border-color var(--transition);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  text-decoration: none;
  color: inherit;
  position: relative;
  overflow: hidden;
}
.room-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  color: inherit;
}
.room-card__badge {
  display: inline-block;
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  align-self: flex-start;
}
.room-card--arena .room-card__badge {
  background: color-mix(in srgb, var(--color-arena) 15%, transparent);
  color: var(--color-arena);
}
.room-card--stage .room-card__badge {
  background: color-mix(in srgb, var(--color-stage) 15%, transparent);
  color: var(--color-stage);
}
.room-card--arena::before,
.room-card--stage::before {
  content: "";
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
}
.room-card--arena::before { background: var(--color-arena); }
.room-card--stage::before { background: var(--color-stage); }
.room-card__title { font-size: 1.15rem; font-weight: 700; margin: 0; color: var(--color-text); }
.room-card__prompt {
  font-size: 0.95rem;
  color: var(--color-text-muted);
  line-height: 1.5;
  margin: 0;
  flex: 1;
}
.room-card__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
  color: var(--color-text-muted);
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border);
  margin-top: auto;
}
.room-card__timer {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  color: var(--color-primary);
}

/* ── Provocation card ──
 * Core content unit. Can appear standalone on a detail page (via
 * .provocation-section) or in a list. Shareable-card-friendly proportions. */
.provocation-section { padding-block: var(--section-pad-y); }
.provocation-section .container { max-width: var(--card-max); }
.provocation-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: clamp(1.5rem, 4vw, 2.5rem);
  box-shadow: var(--shadow-sm);
}
.provocation-card__framing {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-primary);
  margin-bottom: 0.75rem;
}
.provocation-card__prompt {
  font-size: clamp(1.25rem, 3vw, 1.75rem);
  font-weight: 700;
  line-height: 1.3;
  color: var(--color-text);
  margin: 0 0 1.5rem;
  letter-spacing: -0.01em;
}
.provocation-card__ai-take {
  background: var(--color-surface-alt);
  border-left: 3px solid var(--color-primary);
  border-radius: 0 var(--radius) var(--radius) 0;
  padding: 1rem 1.25rem;
  margin-bottom: 1.5rem;
}
.provocation-card__ai-take__label {
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-primary);
  margin-bottom: 0.35rem;
}
.provocation-card__ai-take__body { margin: 0; color: var(--color-text); line-height: 1.55; }
.provocation-card__input {
  margin-top: 1.25rem;
}
.provocation-card__input textarea {
  width: 100%;
  padding: 0.875rem 1rem;
  font: inherit;
  font-size: 1rem;
  border: 2px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 96px;
  resize: vertical;
  transition: border-color var(--transition);
}
.provocation-card__input textarea:focus {
  outline: none;
  border-color: var(--color-primary);
}
.provocation-card__footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1rem;
  gap: 1rem;
  flex-wrap: wrap;
}
.provocation-card__timer {
  font-variant-numeric: tabular-nums;
  font-weight: 700;
  color: var(--color-primary);
  font-size: 1rem;
}

/* ── Reaction bar ── */
.reaction-bar {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 1.25rem;
  padding-top: 1rem;
  border-top: 1px solid var(--color-border);
}
.reaction-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.4rem 0.75rem;
  background: var(--color-surface-alt);
  border: 1px solid var(--color-border);
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition);
  min-height: 36px;
}
.reaction-btn:hover { transform: translateY(-1px); box-shadow: var(--shadow-sm); }
.reaction-btn[aria-pressed="true"] { background: var(--color-primary); color: var(--color-primary-text); border-color: var(--color-primary); }
.reaction-btn__count { font-variant-numeric: tabular-nums; opacity: 0.75; font-weight: 500; }
.reaction-btn--positive[aria-pressed="true"] { background: var(--color-reaction-pos); border-color: var(--color-reaction-pos); }
.reaction-btn--negative[aria-pressed="true"] { background: var(--color-reaction-neg); border-color: var(--color-reaction-neg); }
.reaction-btn--meta[aria-pressed="true"]     { background: var(--color-reaction-meta); border-color: var(--color-reaction-meta); }

/* ── Response stack (user takes under a provocation) ── */
.response-stack { margin-top: 2rem; display: flex; flex-direction: column; gap: 1rem; }
.response-item {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.25rem;
}
.response-item__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.625rem;
  font-size: 0.8rem;
  color: var(--color-text-muted);
}
.response-item__author { font-weight: 600; color: var(--color-text); }
.response-item__body { margin: 0; color: var(--color-text); line-height: 1.55; }
.response-item__reactions { margin-top: 0.75rem; display: flex; gap: 0.5rem; flex-wrap: wrap; font-size: 0.8rem; }

/* ── Archetype / profile page ── */
.archetype-section { padding-block: var(--section-pad-y); }
.archetype-section .container { max-width: 820px; }
.archetype-card {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-accent) 100%);
  color: #ffffff;
  border-radius: var(--radius-lg);
  padding: clamp(1.75rem, 4vw, 3rem);
  text-align: center;
  box-shadow: var(--shadow-md);
}
.archetype-card__label {
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  opacity: 0.85;
}
.archetype-card__name {
  font-size: clamp(1.75rem, 5vw, 2.5rem);
  font-weight: 800;
  margin: 0.5rem 0 1rem;
  letter-spacing: -0.02em;
  color: #ffffff;
}
.archetype-card__description { margin: 0; opacity: 0.9; font-size: 1.05rem; line-height: 1.5; }
.archetype-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 1rem;
  margin-top: 2rem;
}
.archetype-stat {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.25rem;
  text-align: center;
}
.archetype-stat__value { font-size: 1.75rem; font-weight: 800; color: var(--color-primary); font-variant-numeric: tabular-nums; letter-spacing: -0.02em; }
.archetype-stat__label { font-size: 0.8rem; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.06em; font-weight: 600; }

/* ── Generic grids (compat with standard components) ── */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--grid-gap);
  margin-top: 2.5rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--card-pad);
  transition: transform var(--transition), box-shadow var(--transition);
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); }
.feature-icon,
.service-icon { width: 40px; height: 40px; margin-bottom: 1rem; color: var(--color-primary); }

/* ── Forms (including text-input surfaces on provocation pages) ── */
.form-field { margin-bottom: 1.25rem; }
.form-field label { display: block; font-weight: 500; margin-bottom: 0.35rem; font-size: 0.9rem; }
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.75rem 1rem;
  font: inherit;
  border: 2px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 44px;
  transition: border-color var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-primary);
}

/* ── Buttons — pill-shaped CTAs reinforce the friendly energetic feel ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font: inherit;
  font-weight: 600;
  font-size: 0.95rem;
  border-radius: var(--radius-lg);
  border: 2px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: all var(--transition);
}
.btn-primary { background: var(--color-primary); color: var(--color-primary-text); }
.btn-primary:hover { background: var(--color-primary-hover); transform: translateY(-1px); box-shadow: var(--shadow-md); color: var(--color-primary-text); }
.btn-secondary { background: transparent; color: var(--color-primary); border-color: var(--color-primary); }
.btn-secondary:hover { background: var(--color-primary); color: var(--color-primary-text); }
.btn-ghost { background: transparent; color: var(--color-text); border-color: transparent; }
.btn-ghost:hover { background: var(--color-surface-alt); }
.btn-large { padding: 1rem 2rem; font-size: 1.05rem; }

/* ── FAQ ── */
.faq-section .container { max-width: 820px; }
.faq-item { border-bottom: 1px solid var(--color-border); padding: 1.25rem 0; }
.faq-item summary { cursor: pointer; font-weight: 600; list-style: none; display: flex; justify-content: space-between; gap: 1rem; min-height: 44px; align-items: center; }
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; color: var(--color-primary); font-size: 1.25rem; font-weight: 400; }
.faq-item[open] summary::after { content: "−"; }
.faq-item p { padding-top: 0.75rem; color: var(--section-text-muted, var(--color-text-muted)); }

/* ── About / Contact ── */
.about-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; align-items: center; }
.contact-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; }

/* ── Site footer — dark contrast foot to separate from light body ── */
.site-footer { background: var(--color-footer-bg); color: var(--color-footer-text); padding-top: 3.5rem; margin-top: auto; }
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 2rem;
}
.site-footer h3, .site-footer h4 { color: #ffffff; }
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-accent); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; }
.footer-bottom { margin-top: 3rem; padding: 1.5rem 0; border-top: 1px solid rgba(255,255,255,0.1); text-align: center; font-size: 0.9rem; color: rgba(255,255,255,0.6); }

/* ── Responsive ── */
@media (max-width: 1024px) {
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; }
  .provocation-card__footer { flex-direction: column; align-items: stretch; }
}

/* ── Accessibility ── */
:focus-visible { outline: 3px solid var(--color-primary); outline-offset: 2px; }
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; }
}
$template$::text,
    jsonb_build_object(
        'grid_gap',                  '1.5rem',
        'shadow_lg',                 '0 12px 32px rgba(0,0,0,0.12)',
        'shadow_md',                 '0 4px 12px rgba(0,0,0,0.08)',
        'shadow_sm',                 '0 1px 2px rgba(0,0,0,0.06)',
        'card_padding',              '1.75rem',
        'card_max_width',            '680px',
        'border_radius',             '0.75rem',
        'transition_base',           '200ms cubic-bezier(0.4, 0, 0.2, 1)',
        'border_radius_lg',          '1.25rem',
        'border_radius_sm',          '0.375rem',
        'section_padding_y',         '4rem',
        'container_max_width',       '1200px',
        'container_padding_x',       '1.25rem',
        'section_padding_y_mobile',  '2.5rem'
    ),
    'social',
    ARRAY[
        -- Site-type signals
        'social',
        'social-network',
        'social-platform',
        'community-platform',
        -- Modality signals
        'provocation-platform',
        'challenge-platform',
        'game-social',
        'room-based-platform',
        'content-first-social',
        -- AI-seeded signals
        'ai-platform',
        'ai-curated-content',
        'ai-seeded-community',
        -- Creator / identity signals
        'creator-platform',
        'archetype-platform',
        -- Category duplicate
        'social'
    ]::text[],
    true,
    'seed',
    false
);


-- ---------------------------------------------------------------
-- AFTER — verify both inserted and show what the matcher will see
-- ---------------------------------------------------------------

SELECT
    'AFTER' AS phase,
    name,
    category,
    array_length(industry_tags, 1) AS n_tags,
    length(css_template)            AS css_len,
    (SELECT count(*) FROM jsonb_object_keys(structure_tokens)) AS n_tokens
FROM layouts
WHERE name IN ('tool-portal-dark','social-lobby')
ORDER BY name;


-- Sanity check: confirm the 5 renderer-managed surface classes are declared
-- in each new CSS template. If any are missing the renderer's dark-surface
-- defaults won't apply correctly.
SELECT
    name,
    (css_template LIKE '%.features-section%')        AS has_features,
    (css_template LIKE '%.services-section%')        AS has_services,
    (css_template LIKE '%.differentiators-section%') AS has_differentiators,
    (css_template LIKE '%.about-section%')           AS has_about,
    (css_template LIKE '%.faq-section%')             AS has_faq
FROM layouts
WHERE name IN ('tool-portal-dark','social-lobby')
ORDER BY name;

COMMIT;


-- ----------------------------------------------------------------------------
-- What happens next
-- ----------------------------------------------------------------------------
-- These layouts are DORMANT until the classifier emits the matching tag
-- fields. Current gamesdesign.co.uk classification has:
--   site_type = 'interactive-platform'
--   suggested_style = 'professional-dark'
--   (no 'category' field, no 'industry_tags' field)
--
-- So resolveLayoutByTags still takes the "no classification tags" fallback
-- and we'd still land on brochure-formal. Migration 008 addresses this by
-- updating the domain-research-classifier prompt to emit:
--   "category":      "<one of: interactive|social|brochure|editorial|tool|hub|portfolio|storefront>"
--   "industry_tags": ["game-design","game-development","developer-tools",...]
-- chosen from a taxonomy that intersects with the tag sets declared on the
-- new layouts above.
--
-- After 008 lands, for any site whose classification gets regenerated:
--   gamesdesign.co.uk → classifier emits category='interactive' +
--                       industry_tags=['game-design','game-development',
--                       'developer-tools', ...]
--                     → overlap with tool-portal-dark: 4+ matches
--                     → overlap with brochure-formal:  0 matches
--                     → tool-portal-dark wins cleanly
--
-- To test: after 008 lands, mark gamesdesign's resolved_composition
-- is_current=false AND queue needs_composition (the install side-effect
-- lesson from today — see doc 028's failure modes).
-- ----------------------------------------------------------------------------