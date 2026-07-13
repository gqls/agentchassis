-- Migration: layouts.scheme + a light tools-editorial portal layout
--
-- WHY
--   The layout matcher (resolveLayoutByTags) was scheme-blind and the library
--   had no LIGHT multi-tool portal: the only tool-portal/interactive-platform
--   layout is tool-portal-dark (dark). A warm/light tools+editorial site like
--   idea.uk therefore matched the dark layout. This adds (1) a curated `scheme`
--   property the new matcher treats as a near-hard constraint, and (2) a light
--   counterpart to tool-portal-dark so light tools/maker/founder sites have a
--   real structural home.
--
-- SAFETY
--   - scheme is nullable; NULL means "no scheme constraint" so the matcher
--     degrades gracefully on layouts not yet tagged. Rollout is incremental:
--     populate more schemes over time to sharpen matching.
--   - Only the two layouts whose CSS we have actually seen are set here
--     (tool-portal-dark -> dark, soft-editorial -> light). The rest are left
--     NULL on purpose — use the query at the foot to set them accurately from
--     each layout's real background, rather than guessing from the name.
--   - The INSERT is a new seed row (origin='seed'); it does not touch any site.
--
-- ROLLBACK: footer.

\set ON_ERROR_STOP on
BEGIN;

-- 1) scheme column: 'light' | 'dark' | 'neutral' | NULL(unknown)
ALTER TABLE layouts ADD COLUMN IF NOT EXISTS scheme text;
ALTER TABLE layouts DROP CONSTRAINT IF EXISTS layouts_scheme_check;
ALTER TABLE layouts ADD CONSTRAINT layouts_scheme_check
  CHECK (scheme IS NULL OR scheme IN ('light','dark','neutral'));

-- 2) the two we have actually inspected (CSS backgrounds: #121212 / #fffbeb)
UPDATE layouts SET scheme = 'dark'  WHERE name = 'tool-portal-dark';
UPDATE layouts SET scheme = 'light' WHERE name = 'soft-editorial';

-- 3) new layout: a LIGHT, flat, editorial multi-tool portal.
--    Same structural class contract as tool-portal-dark (portal grids, tool
--    workspace, article/reading column, renderer-managed surface sections) so
--    the renderer and components work unchanged — but light fallbacks, flat
--    edges, 1px rules, serif headings. Reads palette/typography vars, so a
--    site palette (e.g. idea.uk parchment) drives the actual colours.
--
--    Tags: the shared shape tags (interactive-platform, tool-portal, tools,
--    interactive) PLUS the specific tags these sites emit (founder-tools,
--    maker-tools, product-development, idea-validation, editorial-publication,
--    practitioner-platform). The specific tags are rare in the library, so the
--    new weighted matcher scores them high; combined with scheme=light this
--    beats tool-portal-dark for a light tools/editorial site, while a dark
--    game-design site still prefers tool-portal-dark (it owns game-design,
--    game-development, developer-tools + scheme=dark).
INSERT INTO layouts (name, display_name, description, css_template, structure_tokens,
                     category, industry_tags, is_active, origin, needs_review, scheme)
VALUES (
  'tool-portal-light',
  'Tool Portal — Light',
  'Warm, flat, editorial multi-tool portal. Index pages present tools, guides, and reports as bordered cards; tool pages give the interactive element a clean workspace with first-class form controls; article/guide pages use a narrow reading column with code, table, and callout support. Ink-on-paper aesthetic: light tinted ground, serif display headings, single restrained accent, hard rectangular edges and 1px rules rather than shadows. The light counterpart to tool-portal-dark. Suits idea/innovation workshops, founder and maker tool sets, practitioner tool libraries, and calculator/guide hubs that want to read as a considered publication rather than a SaaS dashboard.',
  $css$
/* =====================================================================
 * LAYOUT: tool-portal-light
 * Light counterpart to tool-portal-dark. Same structural class contract.
 * Grammar: ink on warm paper, flat, 1px rules, serif headings, generous
 * whitespace. Three page shapes: portal/index grids, tool workspace,
 * article/guide reading column.
 *
 * Renderer contract (identical to tool-portal-dark / brochure-formal):
 *   - MUST NOT declare --section-* defaults; the renderer appends them
 *     after rendering based on palette luminance.
 *   - Renderer-managed surface classes MUST be surface-coloured here:
 *       .features-section, .services-section, .differentiators-section,
 *       .about-section, .faq-section
 *   - Element rules use var(--section-*, var(--color-*)) so dark callout
 *     sections can override per container without restating rules.
 * ===================================================================== */

:root {
  /* ── Palette — LIGHT fallbacks (palette vars override) ── */
  --color-primary:        {{palette "primary"        "#1a1a1a"}};
  --color-primary-hover:  {{palette "primary_hover"  "#000000"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#4a4540"}};
  --color-accent:         {{palette "accent"         "#a8391a"}};
  --color-background:     {{palette "background"     "#faf8f3"}};
  --color-surface:        {{palette "surface"        "#f1ede4"}};
  --color-surface-alt:    {{palette "surface_alt"    "#e9e4d8"}};
  --color-text:           {{palette "text"           "#1a1a1a"}};
  --color-text-muted:     {{palette "text_muted"     "#6b655c"}};
  --color-border:         {{palette "border"         "#1a1a1a"}};
  --color-hairline:       {{palette "hairline"       "#d9d3c7"}};
  --color-card-bg:        {{palette "card_bg"        "#fffdf8"}};
  --color-header-bg:      {{palette "header_bg"      "#faf8f3"}};
  --color-header-text:    {{palette "header_text"    "#1a1a1a"}};
  --color-footer-bg:      {{palette "footer_bg"      "#f1ede4"}};
  --color-footer-text:    {{palette "footer_text"    "#1a1a1a"}};
  --color-code-bg:        {{palette "code_bg"        "#f1ede4"}};
  --color-callout-bg:     {{palette "callout_bg"     "rgba(168,57,26,0.06)"}};
  --color-callout-border: {{palette "callout_border" "#a8391a"}};

  /* ── Typography — serif display, clean sans body ── */
  --font-body:        {{typo "font_family"  "'IBM Plex Sans', system-ui, -apple-system, sans-serif"}};
  --font-heading:     {{typo "heading_font" "'Fraunces', Georgia, 'Times New Roman', serif"}};
  --font-mono:        {{typo "mono_font"    "'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace"}};
  --font-size-base:   {{typo "base_size"    "17px"}};
  --line-height-base: {{typo "line_height"  "1.65"}};

  /* ── Structure — flat, editorial, generous ── */
  --container-max:    {{token "container_max_width"      "1180px"}};
  --container-pad-x:  {{token "container_padding_x"      "1.5rem"}};
  --reading-max:      {{token "reading_max_width"        "720px"}};
  --section-pad-y:    {{token "section_padding_y"        "5rem"}};
  --section-pad-y-sm: {{token "section_padding_y_mobile" "3rem"}};
  --radius:           {{token "border_radius"            "2px"}};
  --radius-sm:        {{token "border_radius_sm"         "2px"}};
  --grid-gap:         {{token "grid_gap"                 "1.5rem"}};
  --card-pad:         {{token "card_padding"             "1.5rem"}};
  --transition:       {{token "transition_base"          "150ms ease"}};

  {{with palette "heading" ""}}--section-heading: {{.}};{{end}}
}

/* ── Base ── */
*, *::before, *::after { box-sizing: border-box; }
html { -webkit-text-size-adjust: 100%; }
body {
  margin: 0; min-height: 100vh; display: flex; flex-direction: column;
  font-family: var(--font-body); font-size: var(--font-size-base);
  line-height: var(--line-height-base); color: var(--color-text);
  background: var(--color-background);
  -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }
code, pre, kbd, samp { font-family: var(--font-mono); }

/* ── Headings (serif) + text inheritance ── */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-text));
  margin: 0 0 1rem; line-height: 1.2; font-weight: 600; letter-spacing: -0.01em;
}
h1 { font-size: clamp(2.25rem, 4.5vw, 3.25rem); font-weight: 600; }
h2 { font-size: clamp(1.6rem, 3vw, 2.1rem); }
h3 { font-size: 1.3rem; }
h4 { font-size: 1.05rem; }
p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
a { color: var(--color-accent); text-decoration: none; transition: color var(--transition); }
a:hover { color: var(--color-primary); text-decoration: underline; text-underline-offset: 2px; }

/* ── Primitives ── */
.container { max-width: var(--container-max); margin-inline: auto; padding-inline: var(--container-pad-x); width: 100%; }
.container--reading { max-width: var(--reading-max); }
.section { padding-block: var(--section-pad-y); }
.section + .section { border-top: 1px solid var(--color-hairline); }

/* ── Header — light, low chrome, hairline rule ── */
.site-header {
  background: var(--color-header-bg); color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border); position: sticky; top: 0; z-index: 1000;
}
.header-container {
  max-width: var(--container-max); margin-inline: auto;
  padding: 1rem var(--container-pad-x);
  display: flex; align-items: center; justify-content: space-between; gap: 2rem;
}
.logo { font-family: var(--font-heading); font-size: 1.25rem; font-weight: 600; color: var(--color-header-text); text-decoration: none; letter-spacing: -0.01em; }
.logo:hover { text-decoration: none; }
.logo-img { max-height: 34px; width: auto; }
.main-nav ul { display: flex; gap: 1.75rem; list-style: none; margin: 0; padding: 0; }
.main-nav a { color: var(--color-header-text); font-weight: 500; font-size: 0.9rem; padding: 0.5rem 0; }
.main-nav a:hover, .main-nav a.active { color: var(--color-accent); text-decoration: none; }
.mobile-menu-toggle { display: none; background: none; border: none; cursor: pointer; padding: 0.5rem; min-width: 44px; min-height: 44px; color: var(--color-header-text); }
.mobile-menu-toggle span { display: block; width: 22px; height: 2px; background: currentColor; margin: 5px 0; }

/* ── Hero — typographic, not imagery-driven ── */
.hero-section { padding-block: calc(var(--section-pad-y) * 1.15); }
.hero-section .container { max-width: 880px; }
.hero-section h1 { font-size: clamp(2.4rem, 5vw, 3.5rem); margin-bottom: 1.25rem; letter-spacing: -0.02em; }
.hero-subtitle, .hero-section .lead { font-size: 1.2rem; color: var(--section-text-muted, var(--color-text-muted)); margin: 0 0 2rem; max-width: 640px; line-height: 1.5; }
.hero-actions { display: flex; gap: 0.75rem; flex-wrap: wrap; }

/* ── Renderer-managed surface sections (MUST be surface-coloured) ── */
.features-section, .services-section, .differentiators-section,
.about-section, .faq-section { background: var(--color-surface); }

/* ── Portal/index grids — bordered cards, flat ── */
.features-grid, .services-grid, .differentiators-grid,
.tools-grid, .guides-grid, .games-grid, .reports-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--grid-gap); margin-top: 2rem;
}
.feature-card, .service-card, .differentiator-card,
.tool-card, .guide-card, .game-card, .report-card {
  background: var(--color-card-bg); border: 1px solid var(--color-border);
  border-radius: var(--radius); padding: var(--card-pad);
  display: flex; flex-direction: column; gap: 0.5rem;
  transition: background var(--transition);
}
.tool-card:hover, .guide-card:hover, .game-card:hover, .report-card:hover,
.feature-card:hover, .service-card:hover, .differentiator-card:hover { background: var(--color-surface-alt); }
.feature-icon, .service-icon, .tool-card__icon, .guide-card__icon, .game-card__icon { width: 28px; height: 28px; margin-bottom: 0.4rem; color: var(--color-accent); }
.tool-card__title, .guide-card__title, .game-card__title, .report-card__title { font-family: var(--font-heading); font-size: 1.15rem; font-weight: 600; margin: 0; }
.tool-card__description, .guide-card__excerpt, .game-card__description, .report-card__description { font-size: 0.95rem; color: var(--section-text-muted, var(--color-text-muted)); margin: 0; flex: 1; }
/* privacy/host labels (private vs AI/hosted) — small, factual */
.tool-card__label, .tool-label { display: inline-block; font-family: var(--font-mono); font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.06em; padding: 0.15rem 0.45rem; border: 1px solid var(--color-hairline); border-radius: var(--radius-sm); color: var(--color-text-muted); }

/* ── Tool page — workspace dominates ── */
.tool-section { padding-block: var(--section-pad-y); background: var(--color-background); }
.tool-section .container { max-width: 960px; }
.tool-workspace { background: var(--color-card-bg); border: 1px solid var(--color-border); border-radius: var(--radius); padding: clamp(1.25rem, 3vw, 2rem); margin-bottom: 2.5rem; }
.tool-workspace__header { display: flex; justify-content: space-between; align-items: baseline; gap: 1rem; margin-bottom: 1.25rem; padding-bottom: 0.75rem; border-bottom: 1px solid var(--color-hairline); }
.tool-workspace__title { font-family: var(--font-heading); font-size: 1.3rem; margin: 0; }
.tool-workspace__meta { font-family: var(--font-mono); font-size: 0.8rem; color: var(--color-text-muted); }
.tool-inputs { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 1.5rem; }
.tool-output { background: var(--color-code-bg); border: 1px solid var(--color-border); border-radius: var(--radius); padding: 1.25rem; font-family: var(--font-mono); font-size: 0.95rem; line-height: 1.5; color: var(--color-text); overflow-x: auto; }
.tool-output__label { display: block; font-family: var(--font-body); font-size: 0.72rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--color-text-muted); margin-bottom: 0.5rem; }
.tool-output__value { font-family: var(--font-heading); font-size: 1.6rem; font-weight: 600; color: var(--color-accent); }

/* ── Article / guide page — narrow reading column ── */
.article-section { padding-block: var(--section-pad-y); }
.article-section .container { max-width: var(--reading-max); }
.article-body { font-size: 1.075rem; line-height: 1.75; color: var(--color-text); }
.article-body h2 { font-size: 1.5rem; margin-top: 2.5rem; padding-top: 1.5rem; border-top: 1px solid var(--color-hairline); }
.article-body h3 { font-size: 1.2rem; margin-top: 2rem; }
.article-body p { margin-bottom: 1.25rem; }
.article-body ul, .article-body ol { margin: 0 0 1.25rem 1.5rem; }
.article-body li { margin-bottom: 0.4rem; }
.article-body blockquote { margin: 1.5rem 0; padding: 0.5rem 1.25rem; border-left: 3px solid var(--color-accent); color: var(--color-text-muted); font-style: italic; }
.article-body code { background: var(--color-surface); border: 1px solid var(--color-hairline); border-radius: var(--radius-sm); padding: 0.1em 0.35em; font-size: 0.9em; }
.article-body pre { background: var(--color-code-bg); border: 1px solid var(--color-border); border-radius: var(--radius); padding: 1rem 1.25rem; overflow-x: auto; margin: 1.5rem 0; font-size: 0.9rem; line-height: 1.5; }
.article-body pre code { background: transparent; border: none; padding: 0; }
.article-body table { width: 100%; border-collapse: collapse; margin: 1.5rem 0; font-size: 0.95rem; }
.article-body th, .article-body td { padding: 0.6rem 0.875rem; text-align: left; border-bottom: 1px solid var(--color-hairline); }
.article-body th { font-weight: 600; background: var(--color-surface); font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.04em; }
.article-body tbody td { color: var(--color-text-muted); }

/* Callout / "design rule" axiom pattern */
.design-rule, .callout { background: var(--color-callout-bg); border-left: 3px solid var(--color-callout-border); padding: 1rem 1.25rem; margin: 1.5rem 0; border-radius: 0 var(--radius) var(--radius) 0; }
.design-rule__label { display: block; font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--color-callout-border); margin-bottom: 0.35rem; }
.design-rule__body { color: var(--color-text); margin: 0; font-weight: 500; }

/* ── Forms (tool controls first-class) ── */
.form-field { margin-bottom: 1rem; }
.form-field label { display: block; font-weight: 600; margin-bottom: 0.3rem; font-size: 0.82rem; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.04em; }
.form-field input, .form-field textarea, .form-field select { width: 100%; padding: 0.625rem 0.75rem; font: inherit; font-size: 0.95rem; border: 1px solid var(--color-border); border-radius: var(--radius); background: var(--color-card-bg); color: var(--color-text); min-height: 42px; transition: border-color var(--transition), box-shadow var(--transition); }
.form-field input:focus, .form-field textarea:focus, .form-field select:focus { outline: none; border-color: var(--color-accent); box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-accent) 25%, transparent); }
.form-field--inline input { font-family: var(--font-mono); }

/* ── Buttons — flat, hard edges ── */
.btn { display: inline-flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 0.7rem 1.4rem; font: inherit; font-weight: 600; font-size: 0.9rem; border-radius: var(--radius); border: 1px solid transparent; cursor: pointer; text-decoration: none; min-height: 44px; transition: background var(--transition), border-color var(--transition), color var(--transition); }
.btn:hover { text-decoration: none; }
.btn-primary { background: var(--color-accent); color: var(--color-primary-text); border-color: var(--color-accent); }
.btn-primary:hover { background: var(--color-primary); border-color: var(--color-primary); color: var(--color-primary-text); }
.btn-secondary { background: transparent; color: var(--color-text); border-color: var(--color-border); }
.btn-secondary:hover { background: var(--color-text); color: var(--color-background); }
.btn-ghost { background: transparent; color: var(--color-text-muted); border-color: transparent; }
.btn-ghost:hover { color: var(--color-text); background: var(--color-surface); }
.btn-large { padding: 0.9rem 1.9rem; font-size: 0.95rem; min-height: 50px; }

/* ── FAQ ── */
.faq-section .container { max-width: var(--reading-max); }
.faq-item { border-bottom: 1px solid var(--color-hairline); padding: 1rem 0; }
.faq-item summary { cursor: pointer; font-family: var(--font-heading); font-weight: 600; list-style: none; display: flex; justify-content: space-between; gap: 1rem; min-height: 44px; align-items: center; }
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; color: var(--color-text-muted); font-size: 1.25rem; }
.faq-item[open] summary::after { content: "\2212"; color: var(--color-accent); }
.faq-item p { padding-top: 0.6rem; color: var(--section-text-muted, var(--color-text-muted)); }

/* ── About / Contact ── */
.about-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; align-items: start; }
.contact-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; }

/* ── Footer — light, hairline ── */
.site-footer { background: var(--color-footer-bg); color: var(--color-footer-text); padding-top: 3rem; margin-top: auto; border-top: 1px solid var(--color-border); font-size: 0.9rem; }
.footer-container { max-width: var(--container-max); margin-inline: auto; padding: 0 var(--container-pad-x); display: grid; grid-template-columns: 2fr 1fr 1fr 1fr; gap: 2rem; }
.site-footer h3, .site-footer h4 { font-family: var(--font-heading); color: var(--color-text); font-size: 1rem; }
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-accent); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.4rem; }
.footer-bottom { margin-top: 2.5rem; padding: 1.25rem 0; border-top: 1px solid var(--color-hairline); text-align: center; font-size: 0.8rem; color: var(--color-text-muted); }

/* ── Responsive ── */
@media (max-width: 1024px) { .footer-container { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .about-section .container, .contact-section .container, .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; }
  .tool-workspace__header { flex-direction: column; align-items: flex-start; }
}

/* ── Accessibility ── */
:focus-visible { outline: 2px solid var(--color-accent); outline-offset: 2px; }
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; } }
$css$,
  '{"grid_gap":"1.5rem","card_padding":"1.5rem","border_radius":"2px","border_radius_sm":"2px","transition_base":"150ms ease","reading_max_width":"720px","section_padding_y":"5rem","container_max_width":"1180px","container_padding_x":"1.5rem","section_padding_y_mobile":"3rem"}'::jsonb,
  'interactive',
  ARRAY['interactive-platform','tool-portal','tools','interactive','founder-tools','maker-tools','product-development','idea-validation','editorial-publication','practitioner-platform','light-utility']::text[],
  true, 'seed', false, 'light'
)
ON CONFLICT (name) DO NOTHING;

COMMIT;

-- ---------------------------------------------------------------------------
-- POPULATE THE REMAINING SCHEMES (run, eyeball, then UPDATE) — do not guess:
--   this prints each layout's background fallback so you can set scheme from
--   the real colour rather than the name.
--
--   SELECT name, scheme,
--          substring(css_template from '--color-background:[^"]*"([^"]*)"') AS bg_default
--   FROM layouts ORDER BY scheme NULLS LAST, name;
--
--   -- then, per layout, e.g.:
--   -- UPDATE layouts SET scheme='light' WHERE name='magazine-grid';
--   -- UPDATE layouts SET scheme='dark'  WHERE name='media-grid';   -- if its bg is dark
--   (Layouts left NULL still work — the matcher just won't apply the scheme
--    constraint to them until set.)
--
-- VERIFY this migration:
--   SELECT name, scheme, category FROM layouts WHERE name IN ('tool-portal-light','tool-portal-dark','soft-editorial');
--   SELECT name, industry_tags FROM layouts WHERE name='tool-portal-light';
--
-- ROLLBACK:
--   DELETE FROM layouts WHERE name='tool-portal-light';
--   -- (leave the scheme column; dropping it would lose the curated values)
-- ---------------------------------------------------------------------------
