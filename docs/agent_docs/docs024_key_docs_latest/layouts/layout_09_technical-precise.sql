-- =====================================================================
-- LAYOUT SEED: technical-precise
-- =====================================================================
-- Character: engineered. Tight radii, glass-effect header, clean
--            sans-serif, subtle shadows, alternating very-subtle
--            section backgrounds, stat displays, flat solid CTAs.
-- Mapped themes: premium-elegant (with serif override), modern-
--                engineering-clean.
-- Default typography: sans-modern (Inter).
-- Default header/footer: existing site-header / site-footer.
--
-- STRUCTURAL DIVERGENCE:
--   - Header uses backdrop-filter: blur (glass effect), semi-transparent
--     background — this is the signature moment
--   - Tight 6px border-radius everywhere (not 0, not large)
--   - Cards bordered (1px solid border), minimal/no shadow by default;
--     hover adds box-shadow + secondary border colour
--   - Typography uses letter-spacing: -0.01em on headings,
--     -webkit-font-smoothing: antialiased
--   - .stats-row component — large tabular numbers with small labels
--   - CTA sections solid primary (no gradient, flat)
--   - Section backgrounds alternate white / very-subtle grey
--   - Footer is light background, muted text (not dark) — contrasts
--     with brochure-* dark footers
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
    'technical-precise',
    'Technical — Precise',
    'Engineering-clean layout with glass-effect header, tight radii, minimal shadows, antialiased sans typography, stat displays, and flat solid CTAs. Section backgrounds alternate white and very-subtle grey. Footer is light. Suits SaaS platforms, infrastructure products, engineering consultancies.',
    'technical',
    ARRAY['saas', 'infrastructure', 'engineering', 'developer-tools', 'b2b-tech'],
    '{
        "container_max_width": "1200px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "5rem",
        "section_padding_y_mobile": "3rem",
        "border_radius": "0.375rem",
        "border_radius_sm": "0.25rem",
        "border_radius_lg": "0.5rem",
        "header_blur": "12px",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.04)",
        "shadow_md": "0 4px 8px rgba(0,0,0,0.06)",
        "shadow_lg": "0 12px 24px rgba(0,0,0,0.08)",
        "transition_base": "180ms cubic-bezier(0.4, 0, 0.2, 1)",
        "card_padding": "1.75rem",
        "stats_gap": "3rem"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: technical-precise
 *
 * Grammar: engineered. Glass. Tight radii. Bordered cards. Tabular
 * stats. Subtle alternating backgrounds. Flat solid CTAs.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#0f172a"}};
  --color-primary-hover:  {{palette "primary_hover"  "#1e293b"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#0ea5e9"}};
  --color-accent:         {{palette "accent"         "#0ea5e9"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f8fafc"}};
  --color-background-alt: {{palette "background_alt" "#f8fafc"}};
  --color-text:           {{palette "text"           "#0f172a"}};
  --color-text-muted:     {{palette "text_muted"     "#64748b"}};
  --color-border:         {{palette "border"         "#e2e8f0"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "rgba(255,255,255,0.75)"}};
  --color-header-text:    {{palette "header_text"    "#0f172a"}};
  --color-cta-bg:         {{palette "cta_bg"         "#0f172a"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#f8fafc"}};
  --color-footer-text:    {{palette "footer_text"    "#475569"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-size-base:   {{typo "base_size"    "15px"}};
  --line-height-base: {{typo "line_height"  "1.6"}};

  /* ── Structure ── */
  --container-max:      {{token "container_max_width"      "1200px"}};
  --container-pad-x:    {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:      {{token "section_padding_y"        "5rem"}};
  --section-pad-y-sm:   {{token "section_padding_y_mobile" "3rem"}};
  --radius:             {{token "border_radius"            "0.375rem"}};
  --radius-sm:          {{token "border_radius_sm"         "0.25rem"}};
  --radius-lg:          {{token "border_radius_lg"         "0.5rem"}};
  --header-blur:        {{token "header_blur"              "12px"}};
  --shadow-sm:          {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.04)"}};
  --shadow-md:          {{token "shadow_md"                "0 4px 8px rgba(0,0,0,0.06)"}};
  --shadow-lg:          {{token "shadow_lg"                "0 12px 24px rgba(0,0,0,0.08)"}};
  --transition:         {{token "transition_base"          "180ms cubic-bezier(0.4, 0, 0.2, 1)"}};
  --card-pad:           {{token "card_padding"             "1.75rem"}};
  --stats-gap:          {{token "stats_gap"                "3rem"}};

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
  text-rendering: optimizeLegibility;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }

/* ── Colour Inheritance Model ── */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--color-primary));
  margin: 0 0 0.75rem;
  line-height: 1.2;
  font-weight: 600;
  letter-spacing: -0.02em;
}
h1 { font-size: clamp(2rem, 3.5vw, 2.75rem); font-weight: 700; }
h2 { font-size: clamp(1.625rem, 2.5vw, 2rem); }
h3 { font-size: 1.25rem; font-weight: 600; letter-spacing: -0.01em; }
h4 { font-size: 1.0625rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
a {
  color: var(--color-accent);
  text-decoration: none;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary); }

/* Eyebrow label — small uppercase over headings */
.eyebrow {
  display: block;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--color-accent);
  margin-bottom: 0.75rem;
}

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── Site header — the signature glass effect ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 1000;
  backdrop-filter: blur(var(--header-blur)) saturate(180%);
  -webkit-backdrop-filter: blur(var(--header-blur)) saturate(180%);
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
  font-size: 1.0625rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--color-header-text);
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
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
  font-size: 0.9375rem;
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
  width: 20px;
  height: 1.5px;
  background: currentColor;
  margin: 5px 0;
}

/* ── Hero — clean, no gradients ── */
.hero-section {
  padding-block: calc(var(--section-pad-y) * 1.2);
  text-align: center;
}
.hero-section .container { max-width: 820px; }
.hero-section h1 { margin-bottom: 1rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 auto 2rem;
  max-width: 620px;
  line-height: 1.5;
}
.hero-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Stats row — a defining component for this layout ── */
.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: var(--stats-gap);
  margin: 3rem 0;
  padding: 2.5rem 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
  text-align: left;
}
.stat-item {}
.stat-value {
  font-size: clamp(2rem, 3.5vw, 2.75rem);
  font-weight: 700;
  letter-spacing: -0.03em;
  color: var(--section-heading, var(--color-primary));
  font-variant-numeric: tabular-nums;
  line-height: 1;
  margin-bottom: 0.375rem;
}
.stat-label {
  font-size: 0.875rem;
  color: var(--section-text-muted, var(--color-text-muted));
  letter-spacing: 0;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Section grids — clean 3-col */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
  margin-top: 2.5rem;
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
  width: 40px;
  height: 40px;
  margin-bottom: 1rem;
  padding: 0.5rem;
  background: color-mix(in srgb, var(--color-accent) 10%, transparent);
  color: var(--color-accent);
  border-radius: var(--radius-sm);
}
.feature-card h3,
.service-card h3,
.differentiator-card h3 { margin-bottom: 0.5rem; }
.feature-card p,
.service-card p,
.differentiator-card p {
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.9375rem;
  margin: 0;
}

/* About */
.about-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
  align-items: center;
}

/* FAQ */
.faq-section .container { max-width: 820px; }
.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1.25rem 0;
}
.faq-item:first-child { border-top: 1px solid var(--color-border); }
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
  content: "";
  display: inline-block;
  width: 10px;
  height: 10px;
  border-right: 2px solid var(--color-text-muted);
  border-bottom: 2px solid var(--color-text-muted);
  transform: rotate(45deg);
  transition: transform var(--transition);
}
.faq-item[open] summary::after { transform: rotate(-135deg); }
.faq-item p {
  padding-top: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.9375rem;
}

/* CTA — solid primary, flat, no gradient */
.call-to-action-section { text-align: center; }
.call-to-action-section .container { max-width: 720px; }

/* Testimonials */
.testimonials-section .testimonial {
  max-width: 720px;
  margin-inline: auto;
  text-align: center;
  font-size: 1.125rem;
  line-height: 1.5;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 1rem;
  font-style: normal;
  font-weight: 600;
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.875rem;
}

/* Contact */
.contact-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
}

/* Forms */
.form-field { margin-bottom: 1rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.375rem;
  font-size: 0.875rem;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.625rem 0.875rem;
  font: inherit;
  font-size: 0.9375rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
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

/* Buttons — flat, solid, small radius */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  font: inherit;
  font-weight: 500;
  font-size: 0.9375rem;
  letter-spacing: -0.005em;
  border-radius: var(--radius);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), color var(--transition),
              border-color var(--transition), box-shadow var(--transition);
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.1);
}
.btn-primary:hover {
  background: var(--color-primary-hover);
  color: var(--color-primary-text);
  box-shadow: var(--shadow-sm);
}
.btn-secondary {
  background: var(--color-card-bg);
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-border);
}
.btn-secondary:hover {
  border-color: var(--color-text-muted);
  background: var(--color-surface);
}
.btn-large { padding: 0.75rem 1.5rem; font-size: 1rem; min-height: 48px; }

/* ── Site footer — LIGHT (contra brochure-* which are dark) ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  border-top: 1px solid var(--color-border);
  padding: 3.5rem 0 0;
  margin-top: auto;
  font-size: 0.875rem;
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 2rem;
}
.site-footer h3, .site-footer h4 {
  color: var(--color-primary);
  font-size: 0.875rem;
  font-weight: 600;
  letter-spacing: 0;
  margin-bottom: 1rem;
}
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-primary); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; }
.footer-bottom {
  margin-top: 2.5rem;
  padding: 1.5rem 0;
  border-top: 1px solid var(--color-border);
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
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
  .site-header { backdrop-filter: none; -webkit-backdrop-filter: none; background: var(--color-background); }
  .features-grid,
  .services-grid,
  .differentiators-grid,
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-background); border-bottom: 1px solid var(--color-border); padding: 0.5rem var(--container-pad-x); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .mobile-menu-toggle { display: inline-flex; }
  .footer-bottom { flex-direction: column; text-align: center; }
}

/* Accessibility */
:focus-visible {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
    transform: none !important;
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
