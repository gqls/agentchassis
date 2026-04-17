-- =====================================================================
-- LAYOUT SEED: high-energy
-- =====================================================================
-- Character: aggressive, kinetic, bold. Impact over subtlety.
-- Mapped themes: boxing.
-- Default typography: display-bold (Impact / Archivo Black).
-- Default header/footer: existing minimal header/footer.
--
-- STRUCTURAL DIVERGENCE:
--   - Uppercase headings with 2px letter-spacing — typography-forward
--   - Hero 80vh minimum with dark background as the default
--   - Diagonal section separators using clip-path: polygon(...)
--     Content is counter-rotated to stay level
--   - Hard edges (0 radius) except where intentional
--   - Hard shadows (offset x+y, no blur) — if any
--   - Alternating dark/light section scheme creates drumbeat rhythm
--   - Large numerals on feature cards (e.g. "01. SPEED")
--   - Footer dark, dense, minimal
--   - Header uppercase too — this layout commits
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
    'high-energy',
    'High Energy — Kinetic',
    'Aggressive, kinetic layout with uppercase display headings, diagonal clip-path section separators, alternating dark/light rhythm, hard edges, and large numerals. Suits boxing gyms, combat sports, fitness events, extreme sports, martial arts clubs.',
    'energy',
    ARRAY['combat-sports', 'fitness', 'events', 'extreme-sports', 'martial-arts', 'boxing'],
    '{
        "container_max_width": "1280px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "6rem",
        "section_padding_y_mobile": "3.5rem",
        "hero_min_height": "80vh",
        "border_radius": "0",
        "border_radius_sm": "0",
        "diagonal_angle": "-3deg",
        "diagonal_slope_top": "polygon(0 3vw, 100% 0, 100% 100%, 0 calc(100% - 3vw))",
        "shadow_hard": "4px 4px 0 rgba(0,0,0,0.9)",
        "shadow_hard_accent": "4px 4px 0 var(--color-accent)",
        "transition_base": "150ms ease",
        "card_padding": "2rem",
        "header_height": "72px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: high-energy
 *
 * Grammar: impact. Uppercase. Diagonal. Hard edges. Alternating dark/
 * light. No softness, no gradients by default, no rounded corners.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#0a0a0a"}};
  --color-primary-hover:  {{palette "primary_hover"  "#1a1a1a"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#ffffff"}};
  --color-accent:         {{palette "accent"         "#ff1744"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#0a0a0a"}};
  --color-text:           {{palette "text"           "#0a0a0a"}};
  --color-text-muted:     {{palette "text_muted"     "#4a4a4a"}};
  --color-border:         {{palette "border"         "#0a0a0a"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#0a0a0a"}};
  --color-header-text:    {{palette "header_text"    "#ffffff"}};
  --color-cta-bg:         {{palette "cta_bg"         "#ff1744"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#0a0a0a"}};
  --color-footer-text:    {{palette "footer_text"    "rgba(255,255,255,0.65)"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, sans-serif"}};
  --font-heading:     {{typo "heading_font" "'Archivo Black', 'Arial Black', Impact, sans-serif"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.5"}};

  /* ── Structure ── */
  --container-max:         {{token "container_max_width"      "1280px"}};
  --container-pad-x:       {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:         {{token "section_padding_y"        "6rem"}};
  --section-pad-y-sm:      {{token "section_padding_y_mobile" "3.5rem"}};
  --hero-min-h:            {{token "hero_min_height"          "80vh"}};
  --radius:                {{token "border_radius"            "0"}};
  --radius-sm:             {{token "border_radius_sm"         "0"}};
  --diag-angle:            {{token "diagonal_angle"           "-3deg"}};
  --diag-clip:             {{token "diagonal_slope_top"       "polygon(0 3vw, 100% 0, 100% 100%, 0 calc(100% - 3vw))"}};
  --shadow-hard:           {{token "shadow_hard"              "4px 4px 0 rgba(0,0,0,0.9)"}};
  --shadow-hard-accent:    {{token "shadow_hard_accent"       "4px 4px 0 var(--color-accent)"}};
  --transition:            {{token "transition_base"          "150ms ease"}};
  --card-pad:              {{token "card_padding"             "2rem"}};
  --header-h:              {{token "header_height"            "72px"}};

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
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}
h1 { font-size: clamp(3rem, 7vw, 5.5rem); letter-spacing: 0.03em; }
h2 { font-size: clamp(2.25rem, 4.5vw, 3.5rem); }
h3 { font-size: 1.5rem; letter-spacing: 0.04em; }
h4 { font-size: 1.125rem; letter-spacing: 0.05em; }

p, li, blockquote {
  color: var(--section-text, inherit);
  margin: 0 0 1rem;
  text-transform: none;
  letter-spacing: 0;
  line-height: 1.55;
}
a {
  color: var(--color-accent);
  text-decoration: none;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary); text-decoration: underline; }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); position: relative; }

/* ── Site header — dark, uppercase, bold ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  position: sticky;
  top: 0;
  z-index: 1000;
  height: var(--header-h);
  display: flex;
  align-items: center;
  border-bottom: 2px solid var(--color-accent);
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2rem;
  width: 100%;
}
.logo {
  font-family: var(--font-heading);
  font-size: 1.25rem;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.05em;
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
  font-family: var(--font-heading);
  font-weight: 700;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 0.5rem 0;
  position: relative;
}
.main-nav a::after {
  content: "";
  position: absolute;
  left: 0;
  right: 100%;
  bottom: -2px;
  height: 2px;
  background: var(--color-accent);
  transition: right var(--transition);
}
.main-nav a:hover::after,
.main-nav a.active::after { right: 0; }
.main-nav a:hover { color: var(--color-accent); text-decoration: none; }
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
  width: 24px;
  height: 3px;
  background: currentColor;
  margin: 5px 0;
}

/* ── Hero — structure only; component owns background ──
 *
 * Hero background is NOT painted by this layout. Per the Dark Section
 * Variable Contract, section backgrounds outside the 5 renderer-
 * managed surface classes (features/services/differentiators/about/
 * faq) belong to the component. A high-energy hero component
 * typically paints itself dark and declares the full --section-*
 * block; a light-hero variant skips both.
 *
 * The layout provides size, flex-centering, and padding — structure,
 * not colour. Text colour on dark hero arrives via --section-* from
 * the component's override block (contract-compliant inheritance). */
.hero-section {
  min-height: var(--hero-min-h);
  display: flex;
  align-items: center;
  padding-block: calc(var(--section-pad-y) * 1.2);
  position: relative;
  overflow: hidden;
}
.hero-section .container {
  text-align: center;
  max-width: 1100px;
  position: relative;
  z-index: 1;
}
.hero-section h1 { margin-bottom: 1.5rem; }
.hero-subtitle, .hero-section .lead {
  font-size: clamp(1.0625rem, 1.5vw, 1.375rem);
  color: var(--section-text-muted, var(--color-text-muted));
  text-transform: uppercase;
  letter-spacing: 0.05em;
  max-width: 780px;
  margin: 0 auto 2.5rem;
  font-weight: 500;
  line-height: 1.4;
}
.hero-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* Accent corner bar in hero (opt-in via component) */
.hero-accent-corner {
  position: absolute;
  top: 0;
  right: 0;
  width: 200px;
  height: 8px;
  background: var(--color-accent);
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending.
 *
 * NOTE: this layout's surface palette is typically DARK
 * (#0a0a0a default). The renderer detects surface luminance and emits
 * appropriate --section-* overrides for text/heading — same contract
 * as every other layout. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* ── Diagonal section separator (opt-in via modifier class) ──
 * Components add .section--diagonal to get a cut-angle top/bottom.
 * Applied selectively because full-layout diagonals become noisy. */
.section--diagonal {
  clip-path: var(--diag-clip);
  padding-block: calc(var(--section-pad-y) + 4vw);
}

/* Section heading with accent underline */
.section__title,
.section h2 {
  position: relative;
  padding-bottom: 1rem;
  margin-bottom: 2rem;
}
.section__title::after,
.section h2::after {
  content: "";
  display: block;
  width: 60px;
  height: 4px;
  background: var(--color-accent);
  margin-top: 0.75rem;
}

/* ── Features / services / differentiators — numbered counters ── */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
  margin-top: 3rem;
  counter-reset: item;
}
.feature-card,
.service-card,
.differentiator-card {
  counter-increment: item;
  background: var(--color-card-bg);
  border: 2px solid var(--color-border);
  padding: var(--card-pad);
  border-radius: var(--radius);
  position: relative;
  transition: transform var(--transition), box-shadow var(--transition);
}
.feature-card::before,
.service-card::before,
.differentiator-card::before {
  content: counter(item, decimal-leading-zero);
  display: block;
  font-family: var(--font-heading);
  font-size: 3.5rem;
  font-weight: 900;
  line-height: 1;
  color: var(--color-accent);
  margin-bottom: 1rem;
  letter-spacing: -0.02em;
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover {
  transform: translate(-2px, -2px);
  box-shadow: var(--shadow-hard-accent);
}
.feature-card h3,
.service-card h3,
.differentiator-card h3 {
  font-size: 1.25rem;
  margin-bottom: 0.75rem;
}
.feature-card p,
.service-card p,
.differentiator-card p {
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.9375rem;
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
  border-top: 2px solid var(--color-border);
  padding: 1.5rem 0;
}
.faq-item:last-child { border-bottom: 2px solid var(--color-border); }
.faq-item summary {
  cursor: pointer;
  font-family: var(--font-heading);
  font-weight: 700;
  font-size: 1.125rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
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
  font-size: 1.75rem;
  color: var(--color-accent);
  transition: transform var(--transition);
}
.faq-item[open] summary::after { transform: rotate(45deg); }
.faq-item p {
  padding-top: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
  text-transform: none;
}

/* CTA / testimonials — component-coloured */
.call-to-action-section { text-align: center; }
.testimonials-section .testimonial {
  max-width: 820px;
  margin-inline: auto;
  text-align: center;
  font-family: var(--font-heading);
  font-size: clamp(1.5rem, 2.5vw, 2rem);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.02em;
  line-height: 1.25;
}
.testimonials-section .testimonial cite {
  display: block;
  margin-top: 1.5rem;
  font-family: var(--font-body);
  font-style: normal;
  font-weight: 700;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--color-accent);
}

/* Contact */
.contact-section .container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 3rem;
}

/* Forms */
.form-field { margin-bottom: 1.25rem; }
.form-field label {
  display: block;
  font-family: var(--font-heading);
  font-weight: 700;
  margin-bottom: 0.5rem;
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.form-field input,
.form-field textarea,
.form-field select {
  width: 100%;
  padding: 0.875rem 1rem;
  font: inherit;
  font-size: 1rem;
  border: 2px solid var(--color-border);
  border-radius: var(--radius);
  background: var(--color-background);
  color: var(--color-text);
  min-height: 48px;
  transition: border-color var(--transition);
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-accent);
}

/* ── Buttons — hard, uppercase, chunky ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  font: inherit;
  font-family: var(--font-heading);
  font-weight: 700;
  font-size: 0.9375rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  border-radius: var(--radius);
  border: 2px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 48px;
  transition: background var(--transition), color var(--transition),
              border-color var(--transition), transform var(--transition),
              box-shadow var(--transition);
}
.btn:hover { text-decoration: none; }
.btn-primary {
  background: var(--color-accent);
  color: var(--color-cta-text);
  border-color: var(--color-accent);
}
.btn-primary:hover {
  transform: translate(-2px, -2px);
  box-shadow: var(--shadow-hard);
  color: var(--color-cta-text);
}
.btn-secondary {
  background: transparent;
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-primary);
}
.btn-secondary:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
  transform: translate(-2px, -2px);
  box-shadow: var(--shadow-hard-accent);
}
.btn-large { padding: 1.25rem 2.5rem; font-size: 1.0625rem; min-height: 56px; }

/* ── Site footer — dark, dense, minimal decoration ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding: 3.5rem 0 0;
  margin-top: auto;
  border-top: 4px solid var(--color-accent);
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
  color: #ffffff;
  font-size: 0.9375rem;
  letter-spacing: 0.08em;
}
.site-footer a { color: var(--color-footer-text); text-transform: none; letter-spacing: 0; font-weight: 500; }
.site-footer a:hover { color: var(--color-accent); text-decoration: none; }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; font-size: 0.9rem; }
.footer-bottom {
  margin-top: 3rem;
  padding: 1.5rem 0;
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.8125rem;
  color: rgba(255,255,255,0.5);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: repeat(2, 1fr); }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
  .section--diagonal { clip-path: none; padding-block: var(--section-pad-y); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .hero-section { min-height: 60vh; padding-block: var(--section-pad-y-sm); }
  .features-grid,
  .services-grid,
  .differentiators-grid,
  .about-section .container,
  .contact-section .container,
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-header-bg); padding: 1rem var(--container-pad-x); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .mobile-menu-toggle { display: inline-flex; }
  .btn { width: 100%; }
  .hero-actions { flex-direction: column; }
}

/* Accessibility */
:focus-visible {
  outline: 3px solid var(--color-accent);
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
