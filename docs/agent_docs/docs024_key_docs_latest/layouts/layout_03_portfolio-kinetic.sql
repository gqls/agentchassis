-- =====================================================================
-- LAYOUT SEED: portfolio-kinetic
-- =====================================================================
-- Phase 3 of 025_palette_layout_typography_migration.
--
-- Character: asymmetric, motion-forward, display-type-led. Creative
--            studio energy. Negative space does the work that borders
--            and cards do in brochure layouts.
-- Mapped themes: none currently — exists for adoption/new-build matching.
-- Default typography: sans-modern (Inter — clean canvas for type).
-- Default header/footer: existing minimal header/footer.
--
-- STRUCTURAL DIVERGENCE from brochure-* layouts:
--   - No hero-centering, no CTA buttons in hero — text offset, animated
--     underline links instead of button-block CTAs
--   - Asymmetric two-column sections (40/60 splits, alternating)
--     instead of centred 3-col grids
--   - Work/project showcase uses CSS grid dense packing, not a uniform
--     3-col grid
--   - Cards have no borders, no background, no shadows — whitespace is
--     the separator; titles reveal on hover
--   - Forms use minimal underline-only inputs
--   - Container is narrower (1140px vs 1200px) to push breathing room
--
-- ----- CONTRACT CHECKS -----
-- 1. Colour Inheritance Model honoured — h1-h6 and p/li/blockquote
--    use var(--section-*, var(--color-*)).
-- 2. No --section-* defaults on section containers; conditional
--    root-level --section-heading only.
-- 3. Renderer-managed surface classes coloured (Phase 4.5 pending).
-- 4. Every helper call has a fallback.
-- 5. Responsive: 1024/768 breakpoints, touch targets >= 44px.
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
    'portfolio-kinetic',
    'Portfolio — Kinetic',
    'Asymmetric, motion-forward layout with oversized display type, negative space, staggered work grids, and animated text-link CTAs instead of buttons. Suits creative studios, design agencies, photographer portfolios, and anywhere the type itself is the visual.',
    'portfolio',
    ARRAY['design-studio', 'creative-agency', 'portfolio', 'photography', 'art-direction'],
    '{
        "container_max_width": "1140px",
        "container_padding_x": "2rem",
        "section_padding_y": "7rem",
        "section_padding_y_mobile": "4rem",
        "hero_min_height": "90vh",
        "display_heading_size": "clamp(3rem, 10vw, 8rem)",
        "asymmetric_split_a": "40%",
        "asymmetric_split_b": "60%",
        "border_radius": "0",
        "transition_base": "500ms cubic-bezier(0.25, 1, 0.5, 1)",
        "transition_fast": "250ms cubic-bezier(0.4, 0, 0.2, 1)",
        "grid_gap": "4rem",
        "work_grid_gap": "1.5rem",
        "underline_thickness": "1px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: portfolio-kinetic
 *
 * Whitespace IS the separator. Type IS the decoration. Motion IS the
 * affordance. Structurally this is not a brochure with extra padding —
 * it's a different grammar: asymmetric splits, dense work grids, text
 * links where other layouts have buttons.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#0f0f0f"}};
  --color-primary-hover:  {{palette "primary_hover"  "#333333"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#666666"}};
  --color-accent:         {{palette "accent"         "#ff3d00"}};
  --color-background:     {{palette "background"     "#fafafa"}};
  --color-surface:        {{palette "surface"        "#f0f0f0"}};
  --color-text:           {{palette "text"           "#1a1a1a"}};
  --color-text-muted:     {{palette "text_muted"     "#7a7a7a"}};
  --color-border:         {{palette "border"         "#1a1a1a"}};
  --color-card-bg:        {{palette "card_bg"        "transparent"}};
  --color-header-bg:      {{palette "header_bg"      "transparent"}};
  --color-header-text:    {{palette "header_text"    "#1a1a1a"}};
  --color-cta-bg:         {{palette "cta_bg"         "#0f0f0f"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#fafafa"}};
  --color-footer-text:    {{palette "footer_text"    "#1a1a1a"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-heading:     {{typo "heading_font" "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.6"}};

  /* ── Structure ── */
  --container-max:     {{token "container_max_width"      "1140px"}};
  --container-pad-x:   {{token "container_padding_x"      "2rem"}};
  --section-pad-y:     {{token "section_padding_y"        "7rem"}};
  --section-pad-y-sm:  {{token "section_padding_y_mobile" "4rem"}};
  --hero-min-h:        {{token "hero_min_height"          "90vh"}};
  --display-size:      {{token "display_heading_size"     "clamp(3rem, 10vw, 8rem)"}};
  --split-a:           {{token "asymmetric_split_a"       "40%"}};
  --split-b:           {{token "asymmetric_split_b"       "60%"}};
  --radius:            {{token "border_radius"            "0"}};
  --transition:        {{token "transition_base"          "500ms cubic-bezier(0.25, 1, 0.5, 1)"}};
  --transition-fast:   {{token "transition_fast"          "250ms cubic-bezier(0.4, 0, 0.2, 1)"}};
  --grid-gap:          {{token "grid_gap"                 "4rem"}};
  --work-grid-gap:     {{token "work_grid_gap"            "1.5rem"}};
  --underline-thick:   {{token "underline_thickness"      "1px"}};

  /* Per Colour Inheritance Model: emit --section-heading at :root only
     if the palette declares it. Otherwise h1-h6 fall through to
     --color-primary via var(--section-heading, var(--color-primary)). */
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
  margin: 0 0 1rem;
  line-height: 0.95;
  font-weight: 700;
  letter-spacing: -0.04em;
}
h1 { font-size: var(--display-size); font-weight: 800; }
h2 { font-size: clamp(2.25rem, 5vw, 4rem); }
h3 { font-size: clamp(1.5rem, 2.5vw, 2rem); line-height: 1.1; }
h4 { font-size: 1.25rem; line-height: 1.2; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
/* strong/em/cite/span: do NOT set color — they inherit from parent */

/* Text links get animated underline, not colour change — central to
   the character of this layout. Buttons-in-hero are replaced by these. */
a {
  color: inherit;
  text-decoration: none;
  background-image: linear-gradient(currentColor, currentColor);
  background-size: 100% var(--underline-thick);
  background-position: 0 100%;
  background-repeat: no-repeat;
  transition: background-size var(--transition);
  padding-bottom: 2px;
}
a:hover {
  background-size: 100% 3px;
}
/* The inline "decorative" link class — used for hero/section CTAs in
   place of buttons. Arrow affordance on hover. */
.link-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
  font-size: 1.125rem;
  color: var(--section-heading, var(--color-primary));
  background-size: 100% var(--underline-thick);
  min-height: 44px;
  line-height: 1.4;
}
.link-cta::after {
  content: "→";
  display: inline-block;
  transition: transform var(--transition-fast);
}
.link-cta:hover::after { transform: translateX(8px); }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── Site header ── transparent, floating, minimal ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 2rem var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}
.logo {
  font-family: var(--font-heading);
  font-size: 1rem;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--color-header-text);
  text-decoration: none;
  background: none;
}
.logo:hover { background: none; }
.logo-img { max-height: 32px; width: auto; }
.main-nav ul {
  display: flex;
  gap: 2.5rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 400;
  font-size: 0.95rem;
  padding: 0.5rem 0;
  background-size: 0% var(--underline-thick);
}
.main-nav a:hover,
.main-nav a.active { background-size: 100% var(--underline-thick); }
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
  width: 28px;
  height: 1px;
  background: currentColor;
  margin: 7px 0;
}

/* ── Hero — display type dominates, text offset-left, no centering ── */
.hero-section {
  min-height: var(--hero-min-h);
  display: flex;
  align-items: flex-end;
  padding-block: 0;
  padding-bottom: 6rem;
  position: relative;
}
.hero-section .container {
  /* Intentionally asymmetric — content sits in left 66% of container */
  max-width: var(--container-max);
}
.hero-section h1 {
  max-width: 12ch;
  margin-bottom: 2rem;
}
.hero-subtitle, .hero-section .lead {
  font-size: clamp(1.125rem, 1.5vw, 1.5rem);
  max-width: 40ch;
  color: var(--section-text-muted, var(--color-text-muted));
  font-weight: 400;
  margin: 0 0 2.5rem;
}
.hero-actions {
  display: flex;
  gap: 2.5rem;
  flex-wrap: wrap;
  align-items: center;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: these 5 class names must stay in sync
 * with buildSectionDefaults in render_css_from_spec_action.go, which
 * hardcodes the same list to emit dark-surface --section-* overrides
 * when surface is dark. Phase 4.5 moves this into components and
 * switches the renderer to a data-section-bg attribute selector. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* ── Asymmetric section pattern ──
 *
 * Alternates between left-heavy and right-heavy two-column splits.
 * Odd-numbered sections: 40% heading column, 60% content column.
 * Even-numbered: 60% content, 40% heading — reading rhythm shifts. */
.features-section .container,
.services-section .container,
.about-section .container,
.differentiators-section .container {
  display: grid;
  grid-template-columns: var(--split-a) var(--split-b);
  gap: var(--grid-gap);
  align-items: start;
}
.section:nth-of-type(even) .container {
  grid-template-columns: var(--split-b) var(--split-a);
}
.section-heading-column {
  position: sticky;
  top: 6rem;
}
.section-heading-column h2 {
  margin-bottom: 1.5rem;
}

/* ── Work / project showcase grid — dense packing, staggered ── */
.work-grid,
.portfolio-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  grid-auto-rows: minmax(200px, auto);
  gap: var(--work-grid-gap);
  margin-top: 3rem;
}
.work-item {
  grid-column: span 6;
  position: relative;
  overflow: hidden;
  background: transparent;
  cursor: pointer;
}
/* Staggered sizing — every third item spans wider, every seventh tall */
.work-item:nth-child(3n) { grid-column: span 8; }
.work-item:nth-child(3n+1) { grid-column: span 4; }
.work-item:nth-child(7n) { grid-row: span 2; }
.work-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition);
}
.work-item:hover img { transform: scale(1.03); }
.work-item-meta {
  padding: 1rem 0 0;
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  opacity: 0;
  transform: translateY(8px);
  transition: opacity var(--transition-fast), transform var(--transition-fast);
}
.work-item:hover .work-item-meta {
  opacity: 1;
  transform: translateY(0);
}
.work-item-title {
  font-size: 1.125rem;
  font-weight: 500;
  letter-spacing: -0.01em;
}
.work-item-category {
  font-size: 0.875rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Feature / service blocks (when not using work-grid) ──
 * Stacked list form, not a card grid. Large numerals, thin rules. */
.feature-list,
.service-list,
.differentiator-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.feature-list > li,
.service-list > li,
.differentiator-list > li {
  display: grid;
  grid-template-columns: 4rem 1fr;
  gap: 2rem;
  padding: 2.5rem 0;
  border-top: 1px solid var(--color-border);
}
.feature-list > li:last-child,
.service-list > li:last-child,
.differentiator-list > li:last-child {
  border-bottom: 1px solid var(--color-border);
}
.feature-list .counter,
.service-list .counter,
.differentiator-list .counter {
  font-family: var(--font-heading);
  font-size: 1.5rem;
  font-weight: 300;
  color: var(--section-text-muted, var(--color-text-muted));
  font-variant-numeric: tabular-nums;
}
.feature-list h3,
.service-list h3,
.differentiator-list h3 { margin-bottom: 0.5rem; }

/* ── FAQ — thin-rule list, no cards ── */
.faq-section .container { max-width: 820px; display: block; }
.faq-item {
  border-top: 1px solid var(--color-border);
  padding: 1.75rem 0;
}
.faq-item:last-child { border-bottom: 1px solid var(--color-border); }
.faq-item summary {
  cursor: pointer;
  font-weight: 500;
  font-size: 1.25rem;
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
  font-weight: 300;
  font-size: 1.5rem;
  transition: transform var(--transition-fast);
}
.faq-item[open] summary::after {
  content: "+";
  transform: rotate(45deg);
}
.faq-item p {
  padding-top: 1rem;
  color: var(--section-text-muted, var(--color-text-muted));
  max-width: 62ch;
}

/* ── About ── asymmetric like other sections */
.about-section .container { align-items: start; }

/* ── CTA / testimonials (component-coloured) ── */
.call-to-action-section { text-align: left; }
.call-to-action-section .container {
  display: block;
  max-width: 960px;
}
.call-to-action-section h2 {
  max-width: 14ch;
  margin-bottom: 2rem;
}
.testimonials-section .testimonial {
  max-width: 840px;
  font-size: clamp(1.5rem, 2.5vw, 2rem);
  line-height: 1.3;
  font-weight: 400;
  letter-spacing: -0.02em;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 2rem;
  font-style: normal;
  font-weight: 500;
  font-size: 0.95rem;
  letter-spacing: 0;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Contact ── asymmetric split */
.contact-section .container {
  display: grid;
  grid-template-columns: var(--split-a) var(--split-b);
  gap: var(--grid-gap);
  align-items: start;
}

/* ── Forms — underline inputs only, no boxes ── */
.form-field { margin-bottom: 2rem; }
.form-field label {
  display: block;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-bottom: 0.5rem;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.75rem 0;
  font: inherit;
  font-size: 1.125rem;
  border: none;
  border-bottom: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text);
  min-height: 44px;
  border-radius: 0;
  transition: border-color var(--transition-fast);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-bottom-color: var(--color-accent);
}

/* ── Buttons — used sparingly; most CTAs are .link-cta ──
 * A button here is a deliberate solid-block moment, e.g. the final
 * submit at the end of a contact form. Rectangular, no radius, no
 * gradient — structural honesty. */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1.25rem 2.5rem;
  font: inherit;
  font-weight: 500;
  font-size: 1rem;
  border-radius: var(--radius);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  background: none;
  min-height: 44px;
  transition: background var(--transition-fast), color var(--transition-fast),
              border-color var(--transition-fast);
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
  border-color: var(--color-primary);
}
.btn-primary:hover {
  background: transparent;
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-primary);
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

/* ── Site footer — minimal, single-line shape ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  margin-top: auto;
  padding: 3rem 0 2rem;
  border-top: 1px solid var(--color-border);
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 2rem;
  flex-wrap: wrap;
}
.footer-brand {
  font-family: var(--font-heading);
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 700;
  letter-spacing: -0.03em;
  line-height: 1;
}
.footer-meta {
  display: flex;
  gap: 2rem;
  font-size: 0.875rem;
  color: var(--section-text-muted, var(--color-text-muted));
}
.footer-meta a { padding-bottom: 0; }
.footer-bottom {
  margin-top: 2.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--section-text-muted, var(--color-text-muted));
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .work-item,
  .work-item:nth-child(3n),
  .work-item:nth-child(3n+1) { grid-column: span 6; }
  .work-item:nth-child(7n) { grid-row: span 1; }
  .features-section .container,
  .services-section .container,
  .about-section .container,
  .differentiators-section .container,
  .contact-section .container,
  .section:nth-of-type(even) .container {
    grid-template-columns: 1fr;
  }
  .section-heading-column { position: static; }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .hero-section { min-height: 70vh; padding-bottom: 3rem; }
  .work-item,
  .work-item:nth-child(3n),
  .work-item:nth-child(3n+1) { grid-column: span 12; }
  .feature-list > li,
  .service-list > li,
  .differentiator-list > li {
    grid-template-columns: 1fr;
    gap: 1rem;
    padding: 2rem 0;
  }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; }
  .footer-container { flex-direction: column; align-items: flex-start; }
  .work-item-meta { opacity: 1; transform: none; }
}

/* ── Accessibility ── */
:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 4px;
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
    transform: none !important;
  }
  .work-item-meta { opacity: 1; transform: none; }
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
