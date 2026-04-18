-- =====================================================================
-- LAYOUT SEED: brochure-formal
-- =====================================================================
-- Phase 1 of 025_palette_layout_typography_migration.
--
-- DEPENDENCY: requires `layouts` table (created in Phase 2 migration).
-- This file is part of Phase 3 seeding; do NOT run before Phase 2.
--
-- Character: structured, understated, CTA-driven. Corporate restraint.
-- Mapped themes: default, standard-brochure, professional-dark.
-- Default header/footer: existing site-header / site-footer components
-- (FK linkage left null for now — header/footer rows not enumerated here).
--
-- ----- CONTRACT CHECKS this template satisfies -----
-- 1. Colour Inheritance Model: element rules use
--    var(--section-*, var(--color-*)) so dark-section components can
--    override per container.
-- 2. Dark Section Variable Contract: NO --section-* defaults declared
--    on section containers. The renderer (buildSectionDefaults) appends
--    them after rendering, picking colours by palette luminance against
--    the merged palette. Dark-section COMPONENTS override on their
--    containers per the Dark Section Variable Contract.
--    The one exception: at :root level, if the palette declares a
--    'heading' slot, we set --section-heading so the palette's heading
--    colour applies in light sections. This is contract-compatible —
--    h1-h6 uses var(--section-heading, var(--color-primary)) exactly
--    as the Colour Inheritance Model specifies.
-- 3. Renderer-managed sections: .differentiators-section,
--    .features-section, .faq-section, .services-section, .about-section
--    are surface-coloured (matches buildSectionDefaults assumption).
--    Hero / CTA / testimonials / contact have no background here —
--    components own them.
-- 4. Template helper fallback: every {{palette ...}}, {{typo ...}},
--    {{token ...}} call has an explicit fallback.
-- 5. Responsive: @media (max-width: 1024px) and (max-width: 768px);
--    touch targets >= 44px.
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
    'brochure-formal',
    'Brochure — Formal',
    'Structured, understated, CTA-driven. Corporate restraint with subtle borders, alternating sections, medium-weight typography. Suits consultancies, law, finance, and B2B professional services.',
    'brochure',
    ARRAY['consultancy', 'law', 'finance', 'b2b', 'professional-services'],
    '{
        "container_max_width": "1200px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "5rem",
        "section_padding_y_mobile": "3rem",
        "border_radius": "0.375rem",
        "border_radius_sm": "0.25rem",
        "border_radius_lg": "0.5rem",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.05)",
        "shadow_md": "0 4px 6px rgba(0,0,0,0.07)",
        "shadow_lg": "0 10px 15px rgba(0,0,0,0.1)",
        "transition_base": "200ms ease",
        "card_padding": "1.5rem",
        "grid_gap": "2rem"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: brochure-formal
 * Variables consumed via map-based template helpers:
 *   {{palette "key" "#hex"}}     palette slot lookup with fallback
 *   {{typo    "key" "default"}}  typography slot lookup with fallback
 *   {{token   "key" "default"}}  structure token lookup with fallback
 *
 * Renderer contract:
 *   - This template MUST NOT declare --section-* defaults; the renderer
 *     appends them after rendering based on palette luminance.
 *   - Renderer-managed surface section classes:
 *       .differentiators-section, .features-section, .faq-section,
 *       .services-section, .about-section
 *     These MUST be surface-coloured here.
 *   - Element rules use var(--section-*, var(--color-*)) so dark-section
 *     components can override per-container without restating rules.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#1a365d"}};
  --color-primary-hover:  {{palette "primary_hover"  "#1e4480"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#2c5282"}};
  --color-accent:         {{palette "accent"         "#3182ce"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f7fafc"}};
  --color-text:           {{palette "text"           "#2d3748"}};
  --color-text-muted:     {{palette "text_muted"     "#718096"}};
  --color-border:         {{palette "border"         "#e2e8f0"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#1a365d"}};
  --color-cta-bg:         {{palette "cta_bg"         "#1a365d"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#1a365d"}};
  --color-footer-text:    {{palette "footer_text"    "rgba(255,255,255,0.85)"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.6"}};

  /* ── Structure ── */
  --container-max:    {{token "container_max_width"      "1200px"}};
  --container-pad-x:  {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:    {{token "section_padding_y"        "5rem"}};
  --section-pad-y-sm: {{token "section_padding_y_mobile" "3rem"}};
  --radius:           {{token "border_radius"            "0.375rem"}};
  --radius-sm:        {{token "border_radius_sm"         "0.25rem"}};
  --radius-lg:        {{token "border_radius_lg"         "0.5rem"}};
  --shadow-sm:        {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.05)"}};
  --shadow-md:        {{token "shadow_md"                "0 4px 6px rgba(0,0,0,0.07)"}};
  --shadow-lg:        {{token "shadow_lg"                "0 10px 15px rgba(0,0,0,0.1)"}};
  --transition:       {{token "transition_base"          "200ms ease"}};
  --card-pad:         {{token "card_padding"             "1.5rem"}};
  --grid-gap:         {{token "grid_gap"                 "2rem"}};

  /* ── Optional palette-driven section overrides ──
   * If the palette declares a 'heading' slot, we emit --section-heading at
   * :root level. Per the Colour Inheritance Model contract, h1-h6 resolve
   * via var(--section-heading, var(--color-primary)). Setting it here makes
   * the palette's heading choice apply in light sections; dark-section
   * components still override it on their container per the Dark Section
   * Variable Contract. Palettes without a 'heading' slot fall through to
   * --color-primary as the contract specifies. */
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

/* ── Colour Inheritance Model ──
 * Element rules use var(--section-*, var(--color-*)) so dark-section
 * components can override per container without restating rules. */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-primary));
  margin: 0 0 1rem;
  line-height: 1.25;
  font-weight: 700;
}
h1 { font-size: 2.5rem; }
h2 { font-size: 2rem; }
h3 { font-size: 1.5rem; font-weight: 600; }
h4 { font-size: 1.25rem; font-weight: 600; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
/* strong/em/cite/span: do NOT set color — they inherit from parent */
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
.section { padding-block: var(--section-pad-y); }

/* ── Site header ── */
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
  font-weight: 700;
  color: var(--color-header-text);
  text-decoration: none;
}
.logo-img { max-height: 40px; width: auto; }
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

/* ── Hero (component-coloured) ── */
.hero-section { padding-block: calc(var(--section-pad-y) * 1.2); }
.hero-section .container { text-align: center; max-width: 880px; }
.hero-section h1 { font-size: clamp(2rem, 4vw, 3rem); margin-bottom: 1rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.25rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 auto 2rem;
  max-width: 640px;
}
.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync
 * with buildSectionDefaults in render_css_from_spec_action.go, which
 * hardcodes the same list to emit dark-surface --section-* overrides
 * when surface is dark.
 *
 * LONG-TERM DIRECTION: move surface painting into the relevant
 * components (features, services, differentiators, about, faq), change
 * the renderer to emit overrides keyed on a data-section-bg attribute
 * instead of hardcoded class names, and remove this block from every
 * layout. Tracked as Phase 4.5 in 025_palette_layout_typography_migration.
 *
 * Until then, these 5 classes MUST be surface-coloured in every layout
 * so the renderer's assumption holds. Hero/CTA/testimonials/contact are
 * NOT in this list — their background is component-owned per the Dark
 * Section Variable Contract. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Generic section grids (3-col → 2 → 1) */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--grid-gap);
  margin-top: 3rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--card-pad);
  transition: border-color var(--transition), box-shadow var(--transition);
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover {
  border-color: var(--color-accent);
  box-shadow: var(--shadow-md);
}
.feature-icon,
.service-icon,
.differentiator-icon {
  width: 48px;
  height: 48px;
  margin-bottom: 1rem;
  color: var(--color-accent);
}

/* ── About ── */
.about-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
  align-items: center;
}

/* ── FAQ ── */
.faq-section .container { max-width: 820px; }
.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1.25rem 0;
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
.faq-item summary::after { content: "+"; font-weight: 400; font-size: 1.25rem; }
.faq-item[open] summary::after { content: "−"; }
.faq-item p {
  padding-top: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Call to action (component-coloured) ── */
.call-to-action-section { text-align: center; }

/* ── Testimonials (component-coloured) ── */
.testimonials-section .testimonial {
  max-width: 720px;
  margin-inline: auto;
  text-align: center;
  font-size: 1.125rem;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 1rem;
  font-style: normal;
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.95rem;
}

/* ── Contact ── */
.contact-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
}

/* ── Forms ── */
.form-field { margin-bottom: 1.25rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.375rem;
  font-size: 0.9rem;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.75rem;
  font: inherit;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 44px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent) 20%, transparent);
}

/* ── Buttons ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font: inherit;
  font-weight: 500;
  border-radius: var(--radius);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), border-color var(--transition),
              color var(--transition), box-shadow var(--transition);
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
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-primary);
}
.btn-secondary:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-large { padding: 1rem 2rem; font-size: 1.05rem; }

/* ── Site footer ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding-top: 4rem;
  margin-top: auto;
}
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
.footer-bottom {
  margin-top: 3rem;
  padding: 1.5rem 0;
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.9rem;
  color: rgba(255,255,255,0.6);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: repeat(2, 1fr); }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  h1 { font-size: 2rem; }
  h2 { font-size: 1.625rem; }
  .features-grid,
  .services-grid,
  .differentiators-grid,
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; }
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
