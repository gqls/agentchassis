-- =====================================================================
-- LAYOUT SEED: soft-editorial
-- =====================================================================
-- Character: warm, reading-first, organic. Tinted background (not pure
--            white), serif display headings, pill buttons, gentle
--            gradients, generous line-heights.
-- Mapped themes: bakery, warm-friendly, calm-minimal, soft-editorial.
-- Default typography: serif-editorial (Merriweather + Lato body).
-- Default header/footer: existing site-header / site-footer.
--
-- STRUCTURAL DIVERGENCE:
--   - Container narrower (1000px) than brochure layouts — reading space
--     is the point
--   - Background is tinted (not #ffffff) — warmth is structural
--   - Hero has gentle wash overlay, not hard background — a 3% primary
--     opacity tint fades to background
--   - Buttons are pill-shaped (border-radius: 50px) not rectangular
--   - Cards have barely-there borders + soft shadows, not hard edges
--   - Section breaks are padding, not rules — no hard dividers
--   - Serif display on headings with letter-spacing: -0.02em
--   - Header transparent by default, floats on background
--   - Line-heights bumped to 1.75 for reading sections
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
    'soft-editorial',
    'Soft — Editorial',
    'Warm, reading-first layout with tinted background, serif display headings, pill buttons, gentle gradient washes, and generous padding. Section transitions rely on spacing rather than rules. Suits wellness blogs, lifestyle sites, personal essays, bakeries, artisan businesses.',
    'editorial',
    ARRAY['wellness', 'lifestyle', 'bakery', 'artisan', 'personal-brand', 'long-form'],
    '{
        "container_max_width": "1000px",
        "container_max_width_wide": "1200px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "6rem",
        "section_padding_y_mobile": "3.5rem",
        "line_height_reading": "1.75",
        "border_radius": "0.75rem",
        "border_radius_pill": "50px",
        "border_radius_sm": "0.5rem",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.03)",
        "shadow_md": "0 4px 12px rgba(0,0,0,0.05)",
        "shadow_soft": "0 10px 40px rgba(0,0,0,0.05)",
        "transition_base": "300ms ease",
        "card_padding": "2rem",
        "hero_wash_opacity": "0.03"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: soft-editorial
 *
 * Grammar: whitespace does the dividing, not rules. Serif on headings,
 * pill buttons, tinted ground, soft shadows.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#9b4020"}};
  --color-primary-hover:  {{palette "primary_hover"  "#7a3218"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#d4a574"}};
  --color-accent:         {{palette "accent"         "#9b4020"}};
  --color-background:     {{palette "background"     "#fffbeb"}};
  --color-surface:        {{palette "surface"        "#fef3c7"}};
  --color-text:           {{palette "text"           "#44403c"}};
  --color-text-muted:     {{palette "text_muted"     "#78716c"}};
  --color-border:         {{palette "border"         "#e7e5e4"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "transparent"}};
  --color-header-text:    {{palette "header_text"    "#44403c"}};
  --color-cta-bg:         {{palette "cta_bg"         "#9b4020"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#fef3c7"}};
  --color-footer-text:    {{palette "footer_text"    "#44403c"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Lato', Georgia, 'Times New Roman', serif"}};
  --font-heading:     {{typo "heading_font" "'Merriweather', Georgia, 'Times New Roman', serif"}};
  --font-size-base:   {{typo "base_size"    "17px"}};
  --line-height-base: {{typo "line_height"  "1.75"}};

  /* ── Structure ── */
  --container-max:         {{token "container_max_width"      "1000px"}};
  --container-max-wide:    {{token "container_max_width_wide" "1200px"}};
  --container-pad-x:       {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:         {{token "section_padding_y"        "6rem"}};
  --section-pad-y-sm:      {{token "section_padding_y_mobile" "3.5rem"}};
  --lh-reading:            {{token "line_height_reading"      "1.75"}};
  --radius:                {{token "border_radius"            "0.75rem"}};
  --radius-pill:           {{token "border_radius_pill"       "50px"}};
  --radius-sm:             {{token "border_radius_sm"         "0.5rem"}};
  --shadow-sm:             {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.03)"}};
  --shadow-md:             {{token "shadow_md"                "0 4px 12px rgba(0,0,0,0.05)"}};
  --shadow-soft:           {{token "shadow_soft"              "0 10px 40px rgba(0,0,0,0.05)"}};
  --transition:            {{token "transition_base"          "300ms ease"}};
  --card-pad:              {{token "card_padding"             "2rem"}};
  --hero-wash:             {{token "hero_wash_opacity"        "0.03"}};

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
  line-height: 1.25;
  font-weight: 700;
  letter-spacing: -0.02em;
}
h1 { font-size: clamp(2.25rem, 4.5vw, 3.25rem); line-height: 1.15; }
h2 { font-size: clamp(1.75rem, 3vw, 2.25rem); }
h3 { font-size: 1.5rem; }
h4 { font-size: 1.25rem; }

p, li, blockquote {
  color: var(--section-text, inherit);
  margin: 0 0 1.25rem;
}
.prose p, .prose li { font-size: 1.0625rem; line-height: var(--lh-reading); max-width: 65ch; }

blockquote {
  font-family: var(--font-heading);
  font-style: italic;
  font-size: 1.5rem;
  line-height: 1.4;
  border-left: none;
  padding: 1rem 0;
  margin: 2.5rem 0;
  text-align: center;
  position: relative;
  max-width: 700px;
  margin-inline: auto;
}
blockquote::before {
  content: "\201C";
  display: block;
  font-size: 4rem;
  line-height: 1;
  color: var(--color-secondary);
  opacity: 0.5;
  margin-bottom: 0.5rem;
}
a {
  color: var(--color-primary);
  text-decoration: none;
  background-image: linear-gradient(var(--color-primary), var(--color-primary));
  background-size: 100% 1px;
  background-position: 0 100%;
  background-repeat: no-repeat;
  transition: background-size var(--transition), color var(--transition);
}
a:hover {
  color: var(--color-primary-hover);
  background-size: 100% 2px;
}

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

/* ── Site header ── transparent, floating ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  position: sticky;
  top: 0;
  z-index: 1000;
  backdrop-filter: blur(6px);
}
.header-container {
  max-width: var(--container-max-wide);
  margin-inline: auto;
  padding: 1.5rem var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
}
.logo {
  font-family: var(--font-heading);
  font-size: 1.375rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--color-header-text);
  background: none;
}
.logo:hover { background: none; }
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
  font-size: 0.9375rem;
  padding: 0.5rem 0;
  background-size: 0% 1px;
}
.main-nav a:hover,
.main-nav a.active { background-size: 100% 1px; }
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
  height: 1.5px;
  background: currentColor;
  margin: 5px 0;
  border-radius: 1px;
}

/* ── Hero — gentle wash overlay ── */
.hero-section {
  padding-block: calc(var(--section-pad-y) * 1.2) var(--section-pad-y);
  position: relative;
  overflow: hidden;
  text-align: center;
}
.hero-section::before {
  content: "";
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at top,
    color-mix(in srgb, var(--color-primary) calc(var(--hero-wash) * 100%), transparent),
    transparent 70%);
  pointer-events: none;
}
.hero-section .container { position: relative; z-index: 1; }
.hero-section h1 {
  margin-bottom: 1.25rem;
  max-width: 18ch;
  margin-inline: auto;
}
.hero-subtitle, .hero-section .lead {
  font-size: 1.25rem;
  color: var(--section-text-muted, var(--color-text-muted));
  max-width: 600px;
  margin: 0 auto 2.5rem;
  line-height: 1.55;
}
.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Section grids — generous, gentle */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2.5rem;
  margin-top: 3rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid rgba(0,0,0,0.03);
  border-radius: var(--radius);
  padding: var(--card-pad);
  box-shadow: var(--shadow-soft);
  transition: transform var(--transition), box-shadow var(--transition);
  text-align: center;
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 20px 50px rgba(0,0,0,0.07);
}
.feature-icon,
.service-icon,
.differentiator-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 1.25rem;
  color: var(--color-primary);
}
.feature-card h3,
.service-card h3,
.differentiator-card h3 { margin-bottom: 0.75rem; }
.feature-card p,
.service-card p,
.differentiator-card p {
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 1rem;
  margin: 0;
}

/* About — centred, narrow reading width */
.about-section .container {
  max-width: 780px;
  text-align: center;
}
.about-section p { max-width: 65ch; margin-inline: auto; }

/* FAQ — soft accordion */
.faq-section .container { max-width: 780px; }
.faq-item {
  background: var(--color-card-bg);
  border-radius: var(--radius);
  padding: 1.25rem 1.5rem;
  margin-bottom: 0.75rem;
  box-shadow: var(--shadow-sm);
  transition: box-shadow var(--transition);
}
.faq-item:hover { box-shadow: var(--shadow-md); }
.faq-item summary {
  cursor: pointer;
  font-family: var(--font-heading);
  font-weight: 700;
  font-size: 1.0625rem;
  list-style: none;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-height: 44px;
  align-items: center;
  color: var(--section-heading, var(--color-primary));
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after {
  content: "+";
  font-weight: 300;
  font-size: 1.5rem;
  color: var(--color-primary);
  transition: transform var(--transition);
}
.faq-item[open] summary::after { transform: rotate(45deg); }
.faq-item p {
  padding-top: 0.5rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* CTA / testimonials — component-coloured */
.call-to-action-section { text-align: center; }
.call-to-action-section .container { max-width: 780px; }

.testimonials-section .testimonial {
  max-width: 720px;
  margin-inline: auto;
  text-align: center;
  font-family: var(--font-heading);
  font-size: clamp(1.25rem, 2vw, 1.625rem);
  font-style: italic;
  line-height: 1.4;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 1.5rem;
  font-family: var(--font-body);
  font-style: normal;
  font-weight: 700;
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* Contact */
.contact-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
  align-items: start;
}

/* Forms */
.form-field { margin-bottom: 1.5rem; }
.form-field label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.5rem;
  font-size: 0.9375rem;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.875rem 1rem;
  font: inherit;
  font-size: 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-card-bg);
  color: var(--color-text);
  min-height: 48px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 15%, transparent);
}

/* ── Buttons — pill-shaped, the defining moment ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.875rem 2rem;
  font: inherit;
  font-family: var(--font-body);
  font-weight: 600;
  font-size: 0.9375rem;
  border-radius: var(--radius-pill);
  border: 1.5px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 48px;
  background-image: none;
  transition: background var(--transition), color var(--transition),
              border-color var(--transition), transform var(--transition);
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover {
  background: var(--color-primary-hover);
  color: var(--color-primary-text);
  transform: translateY(-1px);
  background-size: 0;
}
.btn-secondary {
  background: transparent;
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-primary);
  background-size: 0;
}
.btn-secondary:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
  transform: translateY(-1px);
}
.btn-large { padding: 1rem 2.5rem; font-size: 1rem; min-height: 52px; }

/* ── Site footer — light, warm, not dark ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding: 4rem 0 0;
  margin-top: auto;
  font-family: var(--font-body);
}
.footer-container {
  max-width: var(--container-max-wide);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr;
  gap: 2rem;
}
.site-footer h3, .site-footer h4 {
  color: var(--color-primary);
  font-family: var(--font-heading);
  font-size: 1.0625rem;
}
.site-footer a { color: var(--color-footer-text); background-size: 0% 1px; }
.site-footer a:hover { color: var(--color-primary); background-size: 100% 1px; }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; font-size: 0.9375rem; }
.footer-bottom {
  margin-top: 3rem;
  padding: 2rem 0 1.5rem;
  border-top: 1px solid var(--color-border);
  text-align: center;
  font-size: 0.875rem;
  color: var(--section-text-muted, var(--color-text-muted));
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
  .features-grid,
  .services-grid,
  .differentiators-grid,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-background); padding: 1rem var(--container-pad-x); box-shadow: var(--shadow-md); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .main-nav.is-open a { display: block; padding: 0.75rem 0; }
  .mobile-menu-toggle { display: inline-flex; }
}

/* Accessibility */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 3px;
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
