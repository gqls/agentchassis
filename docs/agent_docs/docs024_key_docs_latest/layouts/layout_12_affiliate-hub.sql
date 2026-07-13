-- =====================================================================
-- LAYOUT SEED: affiliate-hub
-- =====================================================================
-- Character: commercial but trustworthy. Product-focused content with
--            prominent but honest affiliate disclosure.
-- Mapped themes: none.
-- Default typography: sans-modern (Inter).
-- Default header/footer: header-with-cart-or-nav (new), footer-4-column.
--
-- DIVERGENCE from the other commerce-adjacent layouts:
--   - comparison-aggregator: horizontal row cards, headline metric
--   - affiliate-hub (THIS): VERTICAL product cards with image-top,
--     rating stars, price, affiliate CTA. Distinct from comparison —
--     these are "picks" with reviews, not raw price comparison
--   - ecommerce-storefront: image-dominant product grid for a catalogue
--   - Persistent top disclosure strip (visible on every page)
--   - Long-form review blocks with a pros/cons two-column grid
--   - Comparison tables (responsive horizontal scroll) — different
--     primitive from the result-card in comparison-aggregator
--   - Sticky "Top Picks" sidebar is optional in the body grid
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
    'affiliate-hub',
    'Affiliate — Review Hub',
    'Product review and buyer-guide layout with persistent disclosure strip, vertical product cards (image-top + rating + price + affiliate CTA), long-form review blocks with pros/cons grid, comparison tables, and optional sticky Top Picks sidebar. Suits product review sites, "best X for Y" guides, deal aggregators, comparison editorial.',
    'affiliate',
    ARRAY['product-reviews', 'buyers-guide', 'deals', 'affiliate', 'editorial-commerce'],
    '{
        "container_max_width": "1200px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "3.5rem",
        "section_padding_y_mobile": "2rem",
        "disclosure_strip_height": "40px",
        "product_card_min_width": "280px",
        "product_image_aspect": "4 / 3",
        "rating_star_size": "14px",
        "border_radius": "0.5rem",
        "border_radius_sm": "0.25rem",
        "shadow_sm": "0 1px 3px rgba(0,0,0,0.05)",
        "shadow_md": "0 6px 16px rgba(0,0,0,0.08)",
        "transition_base": "200ms ease",
        "card_padding": "1.5rem",
        "header_height": "64px",
        "article_max_width": "780px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: affiliate-hub
 *
 * Grammar: picks + reviews + disclosure. Cards are vertical with
 * image-top, star ratings, a prominent "check price" CTA, and an
 * explicit affiliate label. Review sections are long-form with a
 * pros/cons split.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#ea580c"}};
  --color-primary-hover:  {{palette "primary_hover"  "#c2410c"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#0f172a"}};
  --color-accent:         {{palette "accent"         "#ea580c"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#fafafa"}};
  --color-text:           {{palette "text"           "#18181b"}};
  --color-text-muted:     {{palette "text_muted"     "#71717a"}};
  --color-border:         {{palette "border"         "#e4e4e7"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#18181b"}};
  --color-cta-bg:         {{palette "cta_bg"         "#ea580c"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#18181b"}};
  --color-footer-text:    {{palette "footer_text"    "rgba(255,255,255,0.7)"}};
  --color-rating:         {{palette "rating"         "#f59e0b"}};
  --color-rating-empty:   {{palette "rating_empty"   "#e4e4e7"}};
  --color-pro:            {{palette "pro"            "#16a34a"}};
  --color-con:            {{palette "con"            "#dc2626"}};
  --color-disclosure-bg:  {{palette "disclosure_bg"  "#fef3c7"}};
  --color-disclosure-text: {{palette "disclosure_text" "#78350f"}};
  --color-badge-editor:   {{palette "badge_editor"   "#7c3aed"}};
  --color-badge-deal:     {{palette "badge_deal"     "#dc2626"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.6"}};

  /* ── Structure ── */
  --container-max:      {{token "container_max_width"      "1200px"}};
  --container-pad-x:    {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:      {{token "section_padding_y"        "3.5rem"}};
  --section-pad-y-sm:   {{token "section_padding_y_mobile" "2rem"}};
  --disclosure-h:       {{token "disclosure_strip_height"  "40px"}};
  --product-min:        {{token "product_card_min_width"   "280px"}};
  --product-aspect:     {{token "product_image_aspect"     "4 / 3"}};
  --star-size:          {{token "rating_star_size"         "14px"}};
  --radius:             {{token "border_radius"            "0.5rem"}};
  --radius-sm:          {{token "border_radius_sm"         "0.25rem"}};
  --shadow-sm:          {{token "shadow_sm"                "0 1px 3px rgba(0,0,0,0.05)"}};
  --shadow-md:          {{token "shadow_md"                "0 6px 16px rgba(0,0,0,0.08)"}};
  --transition:         {{token "transition_base"          "200ms ease"}};
  --card-pad:           {{token "card_padding"             "1.5rem"}};
  --header-h:           {{token "header_height"            "64px"}};
  --article-max:        {{token "article_max_width"        "780px"}};

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
  margin: 0 0 0.75rem;
  line-height: 1.25;
  font-weight: 700;
  letter-spacing: -0.01em;
}
h1 { font-size: clamp(1.875rem, 3.5vw, 2.5rem); line-height: 1.15; }
h2 { font-size: 1.625rem; }
h3 { font-size: 1.25rem; }
h4 { font-size: 1.0625rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
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

/* ── Persistent disclosure strip — top of page, always visible ── */
.disclosure-strip {
  background: var(--color-disclosure-bg);
  color: var(--color-disclosure-text);
  text-align: center;
  padding: 0.5rem var(--container-pad-x);
  font-size: 0.8125rem;
  min-height: var(--disclosure-h);
  display: flex;
  align-items: center;
  justify-content: center;
}
.disclosure-strip a {
  color: inherit;
  text-decoration: underline;
  font-weight: 600;
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
  font-weight: 800;
  letter-spacing: -0.01em;
  color: var(--color-header-text);
  text-decoration: none;
  flex: 0 0 auto;
}
.logo-img { max-height: 36px; width: auto; }
.main-nav { margin-left: auto; }
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

/* ── Hero — category-led ── */
.hero-section {
  padding-block: 3rem;
  border-bottom: 1px solid var(--color-border);
}
.hero-section .container { max-width: 960px; text-align: center; }
.hero-eyebrow {
  display: inline-block;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--color-primary);
  margin-bottom: 0.75rem;
}
.hero-section h1 { margin-bottom: 1rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.0625rem;
  color: var(--section-text-muted, var(--color-text-muted));
  max-width: 720px;
  margin: 0 auto 0.75rem;
}
.hero-meta {
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-top: 0.5rem;
}

/* ── Category index ── */
.category-index {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
  margin: 2rem 0;
}
.category-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  text-decoration: none;
  color: var(--color-text);
  transition: border-color var(--transition), box-shadow var(--transition);
}
.category-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
  text-decoration: none;
}
.category-card__icon {
  width: 32px;
  height: 32px;
  color: var(--color-primary);
  flex: 0 0 auto;
}
.category-card__name {
  font-weight: 600;
  font-size: 0.9375rem;
}
.category-card__count {
  display: block;
  font-size: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
  font-weight: 400;
}

/* ── Product cards (vertical, image-top) ── */
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--product-min), 1fr));
  gap: 1.5rem;
}
.product-card {
  display: flex;
  flex-direction: column;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
  transition: border-color var(--transition), box-shadow var(--transition), transform var(--transition);
  position: relative;
}
.product-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

/* Badges overlay */
.product-card__badges {
  position: absolute;
  top: 0.75rem;
  left: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  z-index: 2;
}
.badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  border-radius: var(--radius-sm);
  color: #ffffff;
}
.badge--editor { background: var(--color-badge-editor); }
.badge--deal { background: var(--color-badge-deal); }
.badge--ad { background: var(--color-text-muted); }

.product-card__image {
  aspect-ratio: var(--product-aspect);
  background: var(--color-surface);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}
.product-card__image img {
  max-width: 100%;
  max-height: 100%;
  width: auto;
  height: auto;
  object-fit: contain;
  transition: transform var(--transition);
}
.product-card:hover .product-card__image img { transform: scale(1.03); }

.product-card__body {
  padding: var(--card-pad);
  display: flex;
  flex-direction: column;
  flex: 1;
}
.product-card__name {
  font-size: 1.0625rem;
  font-weight: 600;
  line-height: 1.3;
  margin-bottom: 0.5rem;
  color: var(--section-heading, var(--color-primary));
}
.product-card__name a { color: inherit; text-decoration: none; }
.product-card__name a:hover { color: var(--color-primary); }
.product-card__tagline {
  font-size: 0.875rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-bottom: 0.75rem;
}

/* Rating stars */
.rating {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin-bottom: 0.75rem;
  font-size: 0.8125rem;
}
.rating__stars {
  color: var(--color-rating);
  letter-spacing: 1px;
  font-family: Georgia, serif;
  font-size: var(--star-size);
}
.rating__stars .empty { color: var(--color-rating-empty); }
.rating__count { color: var(--section-text-muted, var(--color-text-muted)); }

.product-card__price {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--color-primary);
  margin-bottom: 0.75rem;
  font-variant-numeric: tabular-nums;
}
.product-card__price-original {
  font-size: 0.875rem;
  font-weight: 400;
  color: var(--section-text-muted, var(--color-text-muted));
  text-decoration: line-through;
  margin-right: 0.375rem;
}
.product-card__cta {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.product-card__cta .btn { width: 100%; }
.product-card__review-link {
  text-align: center;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-primary);
  text-decoration: none;
}
.product-card__review-link:hover { text-decoration: underline; }
.product-card__disclosure {
  font-size: 0.6875rem;
  color: var(--section-text-muted, var(--color-text-muted));
  text-align: center;
  margin-top: 0.375rem;
}

/* ── Long-form review block ── */
.review-block {
  max-width: var(--article-max);
  margin-inline: auto;
}
.review-block h2 {
  border-bottom: 2px solid var(--color-border);
  padding-bottom: 0.75rem;
  margin-bottom: 1.25rem;
}
.review-header {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 2rem;
  padding: 1.5rem;
  background: var(--color-surface);
  border-radius: var(--radius);
  margin-bottom: 2rem;
  align-items: center;
}
.review-header__image {
  aspect-ratio: var(--product-aspect);
  display: flex;
  align-items: center;
  justify-content: center;
}
.review-header__verdict h3 { margin: 0 0 0.5rem; }

/* Pros/cons two-column grid — signature primitive */
.pros-cons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
  margin: 1.5rem 0;
  padding: 1.25rem;
  background: var(--color-surface);
  border-radius: var(--radius);
}
.pros, .cons {
  margin: 0;
  padding: 0;
  list-style: none;
}
.pros__title,
.cons__title {
  font-weight: 700;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}
.pros__title { color: var(--color-pro); }
.pros__title::before { content: "✓"; font-weight: 700; }
.cons__title { color: var(--color-con); }
.cons__title::before { content: "✕"; font-weight: 700; }
.pros li,
.cons li {
  padding-left: 1.25rem;
  position: relative;
  font-size: 0.9375rem;
  margin-bottom: 0.5rem;
  line-height: 1.5;
}
.pros li::before {
  content: "+";
  position: absolute;
  left: 0;
  color: var(--color-pro);
  font-weight: 700;
}
.cons li::before {
  content: "−";
  position: absolute;
  left: 0;
  color: var(--color-con);
  font-weight: 700;
}

/* ── Comparison table ── */
.comparison-table-wrapper {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  margin: 1.5rem 0;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}
.comparison-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9375rem;
  background: var(--color-card-bg);
  min-width: 640px;
}
.comparison-table th,
.comparison-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--color-border);
  vertical-align: top;
}
.comparison-table thead th {
  background: var(--color-surface);
  font-weight: 700;
  font-size: 0.875rem;
  position: sticky;
  top: 0;
}
.comparison-table tbody tr:hover { background: var(--color-surface); }
.comparison-table td.price {
  font-weight: 700;
  color: var(--color-primary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

/* ── Optional body grid with sticky sidebar ── */
.content-grid {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 3rem;
  max-width: var(--container-max);
  margin-inline: auto;
  padding: var(--section-pad-y) var(--container-pad-x);
}
.content-main { min-width: 0; }
.content-sidebar {
  position: sticky;
  top: calc(var(--header-h) + 1.5rem);
  align-self: start;
  max-height: calc(100vh - var(--header-h) - 2rem);
  overflow-y: auto;
}
.top-picks {
  padding: 1.25rem;
  background: var(--color-surface);
  border-radius: var(--radius);
}
.top-picks__title {
  font-size: 0.875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-primary);
}
.top-picks ol {
  counter-reset: pick;
  list-style: none;
  padding: 0;
  margin: 0;
}
.top-picks li {
  counter-increment: pick;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
  display: grid;
  grid-template-columns: 1.5rem 1fr;
  gap: 0.5rem;
  font-size: 0.9375rem;
}
.top-picks li:last-child { border-bottom: none; }
.top-picks li::before {
  content: counter(pick);
  font-weight: 700;
  color: var(--color-primary);
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Features / services — category-styled cards */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 1.25rem;
  margin-top: 1.5rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.25rem;
  transition: border-color var(--transition);
}
.feature-card:hover,
.service-card:hover,
.differentiator-card:hover { border-color: var(--color-primary); }

/* About / Contact / FAQ */
.about-section .container,
.contact-section .container,
.faq-section .container { max-width: var(--article-max); }

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
.faq-item summary::after { content: "+"; color: var(--color-text-muted); }
.faq-item[open] summary::after { content: "−"; }

/* CTA / testimonials */
.call-to-action-section { text-align: center; padding-block: 2.5rem; }
.testimonials-section { padding-block: 2.5rem; }
.testimonials-section .testimonial {
  max-width: 680px;
  margin-inline: auto;
  text-align: center;
  font-size: 1.0625rem;
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

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  font: inherit;
  font-weight: 600;
  font-size: 0.9375rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), border-color var(--transition), color var(--transition);
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
  background: var(--color-card-bg);
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-border);
}
.btn-secondary:hover { border-color: var(--color-primary); }

/* ── Site footer — 4-column with prominent disclosure column ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding: 3rem 0 0;
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
.site-footer h3, .site-footer h4 {
  color: #ffffff;
  font-size: 0.875rem;
  font-weight: 600;
  margin-bottom: 0.75rem;
}
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: #ffffff; }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; font-size: 0.875rem; }

.footer-disclosure-block {
  grid-column: 1 / -1;
  padding: 1.5rem 0 0;
  margin-top: 2rem;
  border-top: 1px solid rgba(255,255,255,0.1);
  font-size: 0.8125rem;
  line-height: 1.6;
}
.footer-disclosure-block strong { color: #ffffff; }
.footer-bottom {
  padding: 1.5rem var(--container-pad-x);
  border-top: 1px solid rgba(255,255,255,0.1);
  margin-top: 1rem;
  text-align: center;
  font-size: 0.8125rem;
  color: rgba(255,255,255,0.5);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .content-grid { grid-template-columns: 1fr; }
  .content-sidebar { position: static; max-height: none; }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
  .review-header { grid-template-columns: 1fr; }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .pros-cons { grid-template-columns: 1fr; }
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
