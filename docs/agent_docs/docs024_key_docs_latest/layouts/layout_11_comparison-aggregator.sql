-- =====================================================================
-- LAYOUT SEED: comparison-aggregator
-- =====================================================================
-- Character: search-first, data-dense, trustworthy. Users arrive to
--            compare options and leave with a decision.
-- Mapped themes: none.
-- Default typography: sans-modern (Inter).
-- Default header/footer: header-with-search (new), footer-with-
--                        disclaimer (new).
--
-- STRUCTURAL DIVERGENCE from the four other "commerce-adjacent"
-- layouts (affiliate-hub, ecommerce-storefront, tool-first-landing,
-- industry-hub) — each has a different core grammar:
--
--   - comparison-aggregator: hero IS a search input; sticky filter
--     bar below header; results as dense cards with a prominent
--     "headline metric" per card; banner callouts for regulatory
--     info; heavy disclaimer footer.
--
--   This layout's defining primitive: .result-card — a horizontal
--   row-style card (not the vertical product card of affiliate-hub
--   or the image-dominant card of ecommerce-storefront) with
--   name + headline metric + secondary metric + location + CTA.
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
    'comparison-aggregator',
    'Comparison — Aggregator',
    'Search-first comparison layout with hero search input, sticky filter bar, info banners, dense result cards with headline metric, calculator region, guide card pairs, and heavy-disclosure footer. Suits price comparison, insurance comparison, broadband comparison, trade directories, regulatory information hubs.',
    'comparison',
    ARRAY['comparison', 'aggregator', 'directory', 'insurance', 'finance', 'price-comparison'],
    '{
        "container_max_width": "1280px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "3rem",
        "section_padding_y_mobile": "2rem",
        "hero_padding_y": "4rem",
        "search_input_height": "56px",
        "filter_bar_height": "64px",
        "result_card_min_width": "320px",
        "border_radius": "0.5rem",
        "border_radius_sm": "0.25rem",
        "border_radius_pill": "9999px",
        "shadow_sm": "0 1px 3px rgba(0,0,0,0.06)",
        "shadow_md": "0 4px 12px rgba(0,0,0,0.08)",
        "transition_base": "180ms ease",
        "card_padding": "1.5rem",
        "metric_size": "2rem",
        "header_height": "64px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: comparison-aggregator
 *
 * Grammar: search → filter → compare. Every piece of chrome serves
 * the decision task. Disclaimers are visible, not buried.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#0369a1"}};
  --color-primary-hover:  {{palette "primary_hover"  "#075985"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#0c4a6e"}};
  --color-accent:         {{palette "accent"         "#0369a1"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f1f5f9"}};
  --color-text:           {{palette "text"           "#0f172a"}};
  --color-text-muted:     {{palette "text_muted"     "#64748b"}};
  --color-border:         {{palette "border"         "#e2e8f0"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#0f172a"}};
  --color-cta-bg:         {{palette "cta_bg"         "#0369a1"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#0f172a"}};
  --color-footer-text:    {{palette "footer_text"    "rgba(255,255,255,0.75)"}};
  --color-info-bg:        {{palette "info_bg"        "#eff6ff"}};
  --color-info-border:    {{palette "info_border"    "#3b82f6"}};
  --color-warn-bg:        {{palette "warn_bg"        "#fef3c7"}};
  --color-warn-border:    {{palette "warn_border"    "#d97706"}};
  --color-metric:         {{palette "metric"         "#0369a1"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-size-base:   {{typo "base_size"    "15px"}};
  --line-height-base: {{typo "line_height"  "1.55"}};

  /* ── Structure ── */
  --container-max:      {{token "container_max_width"      "1280px"}};
  --container-pad-x:    {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:      {{token "section_padding_y"        "3rem"}};
  --section-pad-y-sm:   {{token "section_padding_y_mobile" "2rem"}};
  --hero-pad-y:         {{token "hero_padding_y"           "4rem"}};
  --search-h:           {{token "search_input_height"      "56px"}};
  --filter-h:           {{token "filter_bar_height"        "64px"}};
  --result-min:         {{token "result_card_min_width"    "320px"}};
  --radius:             {{token "border_radius"            "0.5rem"}};
  --radius-sm:          {{token "border_radius_sm"         "0.25rem"}};
  --radius-pill:        {{token "border_radius_pill"       "9999px"}};
  --shadow-sm:          {{token "shadow_sm"                "0 1px 3px rgba(0,0,0,0.06)"}};
  --shadow-md:          {{token "shadow_md"                "0 4px 12px rgba(0,0,0,0.08)"}};
  --transition:         {{token "transition_base"          "180ms ease"}};
  --card-pad:           {{token "card_padding"             "1.5rem"}};
  --metric-size:        {{token "metric_size"              "2rem"}};
  --header-h:           {{token "header_height"            "64px"}};

  {{with palette "heading" ""}}--section-heading: {{.}};{{end}}
}

/* ── Base reset ── */
*, *::before, *::after { box-sizing: border-box; }
html { -webkit-text-size-adjust: 100%; scroll-padding-top: calc(var(--header-h) + var(--filter-h) + 1rem); }
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
h1 { font-size: clamp(1.875rem, 3.5vw, 2.5rem); }
h2 { font-size: 1.5rem; }
h3 { font-size: 1.125rem; }
h4 { font-size: 1rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 0.75rem; }
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
}
.header-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  align-items: center;
  gap: 1.5rem;
  width: 100%;
}
.logo {
  font-size: 1.125rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--color-header-text);
  text-decoration: none;
  flex: 0 0 auto;
}
.logo-img { max-height: 36px; width: auto; }
.main-nav {
  margin-left: auto;
}
.main-nav ul {
  display: flex;
  gap: 1.5rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 500;
  font-size: 0.9375rem;
  padding: 0.5rem 0;
  text-decoration: none;
}
.main-nav a:hover,
.main-nav a.active { color: var(--color-primary); }
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

/* ── Hero — the search IS the hero ── */
.hero-section {
  padding-block: var(--hero-pad-y);
  background: var(--color-surface);
  text-align: center;
}
.hero-section .container { max-width: 820px; }
.hero-section h1 { margin-bottom: 0.75rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.0625rem;
  color: var(--section-text-muted, var(--color-text-muted));
  max-width: 620px;
  margin: 0 auto 2rem;
}

.hero-search {
  display: flex;
  gap: 0.5rem;
  max-width: 680px;
  margin: 0 auto;
  padding: 0.5rem;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
}
.hero-search input {
  flex: 1;
  height: var(--search-h);
  padding: 0 1rem;
  font: inherit;
  font-size: 1.0625rem;
  border: none;
  background: transparent;
  color: var(--color-text);
  min-width: 0;
}
.hero-search input:focus { outline: none; }
.hero-search .btn {
  height: var(--search-h);
  min-height: var(--search-h);
  padding: 0 1.5rem;
  flex: 0 0 auto;
}
.hero-disclosure {
  display: inline-block;
  margin-top: 1rem;
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Info / regulatory callout banners ── */
.info-banner,
.warn-banner {
  display: flex;
  gap: 0.875rem;
  align-items: flex-start;
  padding: 1rem 1.25rem;
  border-left: 4px solid;
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  margin: 1.5rem 0;
  font-size: 0.9375rem;
}
.info-banner {
  background: var(--color-info-bg);
  border-left-color: var(--color-info-border);
  color: var(--color-text);
}
.warn-banner {
  background: var(--color-warn-bg);
  border-left-color: var(--color-warn-border);
  color: var(--color-text);
}
.info-banner__title,
.warn-banner__title {
  font-weight: 600;
  margin-bottom: 0.125rem;
}
.info-banner p,
.warn-banner p { margin: 0; }

/* ── Sticky filter bar ── */
.filter-bar {
  position: sticky;
  top: var(--header-h);
  z-index: 900;
  background: var(--color-background);
  border-bottom: 1px solid var(--color-border);
  padding: 0.75rem 0;
  box-shadow: var(--shadow-sm);
}
.filter-bar__inner {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}
.filter-group {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.875rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-pill);
  font-size: 0.875rem;
  min-height: 40px;
  cursor: pointer;
  transition: border-color var(--transition), background var(--transition);
}
.filter-group:hover { border-color: var(--color-primary); }
.filter-group label {
  font-weight: 500;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0;
}
.filter-group select,
.filter-group input {
  border: none;
  background: transparent;
  font: inherit;
  font-size: 0.875rem;
  color: var(--color-text);
  padding: 0;
  min-height: auto;
}
.filter-group select:focus,
.filter-group input:focus { outline: none; }

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  height: 32px;
  padding: 0 0.75rem;
  font-size: 0.8125rem;
  font-weight: 500;
  background: var(--color-surface);
  color: var(--color-text);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-pill);
  cursor: pointer;
  text-decoration: none;
  transition: background var(--transition), color var(--transition);
}
.filter-chip:hover { background: var(--color-border); }
.filter-chip.is-active {
  background: var(--color-primary);
  color: var(--color-primary-text);
  border-color: var(--color-primary);
}
.filter-chip .remove { font-weight: 400; opacity: 0.7; }

.results-meta {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 1rem var(--container-pad-x);
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
}
.results-meta__count { font-weight: 600; color: var(--color-text); }

/* ── Results grid — the defining primitive ── */
.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--result-min), 1fr));
  gap: 1rem;
  padding-bottom: var(--section-pad-y);
}

/* Result card — horizontal/vertical hybrid */
.result-card {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 1rem 1.5rem;
  padding: var(--card-pad);
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  transition: border-color var(--transition), box-shadow var(--transition);
}
.result-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}
.result-card__primary { min-width: 0; }
.result-card__name {
  font-size: 1.0625rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
  line-height: 1.3;
}
.result-card__name a { color: var(--color-text); text-decoration: none; }
.result-card__name a:hover { color: var(--color-primary); }
.result-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-bottom: 0.5rem;
}
.result-card__meta .separator { opacity: 0.5; }
.result-card__tag {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  background: var(--color-surface);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: 0.02em;
}
.result-card__description {
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0.5rem 0 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.result-card__secondary {
  text-align: right;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 1rem;
  min-width: 140px;
}
.result-card__metric {
  font-size: var(--metric-size);
  font-weight: 700;
  color: var(--color-metric);
  line-height: 1;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.02em;
}
.result-card__metric-label {
  font-size: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-top: 0.25rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.result-card__cta .btn { width: 100%; }

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* ── Calculator / tool section ── */
.calculator-section {
  padding: 2rem;
  background: var(--color-surface);
  border-radius: var(--radius);
  margin: var(--section-pad-y) 0;
}
.calculator-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: start;
}

/* ── Guide cards ── */
.guide-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
  margin-top: 1.5rem;
}
.guide-card {
  padding: var(--card-pad);
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  transition: border-color var(--transition);
}
.guide-card:hover { border-color: var(--color-primary); }
.guide-card h3 {
  margin-bottom: 0.5rem;
  font-size: 1.0625rem;
}
.guide-card p {
  margin: 0 0 0.75rem;
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
}
.guide-card__link {
  font-weight: 600;
  font-size: 0.9rem;
}

/* Features/services kept quiet on this layout */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 1rem;
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

/* About / FAQ */
.about-section .container,
.faq-section .container { max-width: 820px; }

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
.faq-item summary::after {
  content: "+";
  color: var(--color-text-muted);
  font-weight: 400;
}
.faq-item[open] summary::after { content: "−"; }

/* CTA/testimonials — muted */
.call-to-action-section { text-align: center; padding-block: 2rem; }
.testimonials-section { padding-block: 2rem; }
.testimonials-section .testimonial {
  max-width: 640px;
  margin-inline: auto;
  text-align: center;
  font-size: 1rem;
}

/* Contact */
.contact-section .container { max-width: 720px; }

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
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  font: inherit;
  font-weight: 600;
  font-size: 0.9375rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  white-space: nowrap;
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
  background: var(--color-background);
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-border);
}
.btn-secondary:hover { border-color: var(--color-primary); }

/* ── Site footer — dark with heavy disclaimer block ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  margin-top: auto;
  padding-top: 3rem;
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
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
}
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: #ffffff; text-decoration: underline; }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.375rem; font-size: 0.875rem; }

/* Heavy disclaimer block — distinctive to this layout */
.footer-disclaimer {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 2.5rem var(--container-pad-x);
  margin-top: 2.5rem;
  border-top: 1px solid rgba(255,255,255,0.1);
  font-size: 0.8125rem;
  line-height: 1.6;
  color: rgba(255,255,255,0.6);
}
.footer-disclaimer__title {
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 0.75rem;
  font-size: 0.875rem;
}
.footer-disclaimer p { margin: 0 0 0.75rem; }
.footer-disclaimer p:last-child { margin-bottom: 0; }

.footer-bottom {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 1.5rem var(--container-pad-x);
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.8125rem;
  color: rgba(255,255,255,0.5);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .calculator-grid,
  .guide-grid { grid-template-columns: 1fr; }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .hero-section { padding-block: calc(var(--hero-pad-y) * 0.7); }
  .hero-search { flex-direction: column; padding: 0.75rem; }
  .hero-search .btn { width: 100%; }
  .result-card {
    grid-template-columns: 1fr;
  }
  .result-card__secondary {
    text-align: left;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    min-width: 0;
  }
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: 1fr; }
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-header-bg); border-bottom: 1px solid var(--color-border); padding: 0.75rem var(--container-pad-x); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .mobile-menu-toggle { display: inline-flex; }
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
