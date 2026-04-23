-- =====================================================================
-- LAYOUT SEED: ecommerce-storefront
-- =====================================================================
-- Character: retail-clean, product-forward. Commercial trust signals.
-- Mapped themes: none.
-- Default typography: sans-modern (Inter).
-- Default header/footer: header-with-cart-or-nav (new), footer-4-column.
--
-- DIVERGENCE from affiliate-hub specifically:
--   - affiliate-hub: review site picking products to recommend — image
--     uses `object-fit: contain` with padding (product on white),
--     card body has rating + tagline + affiliate CTA + disclosure
--   - THIS (ecommerce-storefront): actual catalogue — image uses
--     `object-fit: cover` (photography lifestyle shots), card body is
--     name + price + "Add to Cart", sale-strike pricing, quick-view
--     button overlay on hover
--   - Category tiles are LARGE image cards with overlaid text, not
--     the compact icon+name tiles of affiliate-hub
--   - Trust bar strip (thin horizontal band: free shipping, returns,
--     secure checkout, satisfaction guarantee)
--   - Mini-cart dropdown structure (CSS only; behaviour is JS)
--   - Newsletter signup gets a prominent marketing section
--   - Payment method icons in footer
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
    'ecommerce-storefront',
    'E-commerce — Storefront',
    'Retail-clean storefront layout with promo hero, large image-overlay category tiles, product grid with lifestyle photography (cover-fit), add-to-cart CTAs, strike-through sale pricing, trust bar strip, cart-count header badge, newsletter section, and payment-icon footer. Suits independent shops, small-catalogue retailers, marketplace sellers.',
    'commerce',
    ARRAY['retail', 'ecommerce', 'shop', 'marketplace', 'product-catalogue', 'storefront'],
    '{
        "container_max_width": "1280px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "4rem",
        "section_padding_y_mobile": "2.5rem",
        "hero_min_height": "60vh",
        "product_card_min_width": "240px",
        "product_image_aspect": "1 / 1",
        "category_tile_aspect": "4 / 5",
        "border_radius": "0.5rem",
        "border_radius_sm": "0.25rem",
        "border_radius_lg": "1rem",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.04)",
        "shadow_md": "0 8px 24px rgba(0,0,0,0.1)",
        "transition_base": "250ms ease",
        "card_padding": "1rem",
        "header_height": "72px",
        "trust_bar_height": "48px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: ecommerce-storefront
 *
 * Grammar: catalogue + trust. Image-forward product cards,
 * lifestyle-photo category tiles, trust signals above the fold and
 * in the footer.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#111827"}};
  --color-primary-hover:  {{palette "primary_hover"  "#1f2937"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#6b7280"}};
  --color-accent:         {{palette "accent"         "#dc2626"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f9fafb"}};
  --color-text:           {{palette "text"           "#111827"}};
  --color-text-muted:     {{palette "text_muted"     "#6b7280"}};
  --color-border:         {{palette "border"         "#e5e7eb"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#111827"}};
  --color-cta-bg:         {{palette "cta_bg"         "#111827"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#111827"}};
  --color-footer-text:    {{palette "footer_text"    "rgba(255,255,255,0.75)"}};
  --color-sale:           {{palette "sale"           "#dc2626"}};
  --color-sold-out:       {{palette "sold_out"       "#6b7280"}};
  --color-trust-bar-bg:   {{palette "trust_bar_bg"   "#111827"}};
  --color-trust-bar-text: {{palette "trust_bar_text" "rgba(255,255,255,0.9)"}};

  /* Product cards stay neutral regardless of palette — product images
     demand a clean backdrop. If a palette wants coloured cards it can
     override card_bg, but this fallback keeps defaults clean. */
  --color-product-bg:     {{palette "product_bg"     "#ffffff"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-size-base:   {{typo "base_size"    "15px"}};
  --line-height-base: {{typo "line_height"  "1.55"}};

  /* ── Structure ── */
  --container-max:         {{token "container_max_width"      "1280px"}};
  --container-pad-x:       {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:         {{token "section_padding_y"        "4rem"}};
  --section-pad-y-sm:      {{token "section_padding_y_mobile" "2.5rem"}};
  --hero-min-h:            {{token "hero_min_height"          "60vh"}};
  --product-min:           {{token "product_card_min_width"   "240px"}};
  --product-aspect:        {{token "product_image_aspect"     "1 / 1"}};
  --category-aspect:       {{token "category_tile_aspect"     "4 / 5"}};
  --radius:                {{token "border_radius"            "0.5rem"}};
  --radius-sm:             {{token "border_radius_sm"         "0.25rem"}};
  --radius-lg:             {{token "border_radius_lg"         "1rem"}};
  --shadow-sm:             {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.04)"}};
  --shadow-md:             {{token "shadow_md"                "0 8px 24px rgba(0,0,0,0.1)"}};
  --transition:            {{token "transition_base"          "250ms ease"}};
  --card-pad:              {{token "card_padding"             "1rem"}};
  --header-h:              {{token "header_height"            "72px"}};
  --trust-h:               {{token "trust_bar_height"         "48px"}};

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
  line-height: 1.2;
  font-weight: 600;
  letter-spacing: -0.015em;
}
h1 { font-size: clamp(2rem, 4vw, 2.75rem); }
h2 { font-size: clamp(1.5rem, 2.5vw, 2rem); }
h3 { font-size: 1.125rem; font-weight: 500; }
h4 { font-size: 1rem; font-weight: 500; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
a {
  color: var(--color-text);
  text-decoration: none;
  transition: color var(--transition);
}
a:hover { color: var(--color-primary-hover); }

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── Trust bar — thin horizontal strip above header ── */
.trust-bar {
  background: var(--color-trust-bar-bg);
  color: var(--color-trust-bar-text);
  min-height: var(--trust-h);
  padding: 0.5rem var(--container-pad-x);
  font-size: 0.8125rem;
  display: flex;
  align-items: center;
  justify-content: center;
}
.trust-bar__inner {
  max-width: var(--container-max);
  width: 100%;
  display: flex;
  justify-content: space-around;
  gap: 1rem;
  flex-wrap: wrap;
  text-align: center;
}
.trust-item {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}
.trust-item__icon {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
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
  gap: 2rem;
  width: 100%;
}
.logo {
  font-size: 1.375rem;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--color-header-text);
  text-decoration: none;
  flex: 0 0 auto;
}
.logo-img { max-height: 40px; width: auto; }
.main-nav {
  margin-left: auto;
}
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
}
.main-nav a:hover,
.main-nav a.active { color: var(--color-primary); }

/* Header actions: search toggle, account, cart */
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-left: 1rem;
}
.icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  min-width: 44px;
  min-height: 44px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-header-text);
  cursor: pointer;
  position: relative;
}
.icon-btn:hover { background: var(--color-surface); }
.cart-badge {
  position: absolute;
  top: 6px;
  right: 6px;
  min-width: 18px;
  height: 18px;
  padding: 0 4px;
  background: var(--color-accent);
  color: #ffffff;
  border-radius: 9px;
  font-size: 0.6875rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  font-variant-numeric: tabular-nums;
}

/* Mini-cart dropdown (structure only; JS toggles .is-open) */
.mini-cart {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  width: 360px;
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-md);
  padding: 1rem;
  opacity: 0;
  visibility: hidden;
  transition: opacity var(--transition), visibility var(--transition);
}
.mini-cart.is-open { opacity: 1; visibility: visible; }

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

/* ── Hero — promotional banner with full-bleed image ── */
.hero-section {
  min-height: var(--hero-min-h);
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  background: var(--color-surface);
}
.hero-section__image {
  position: absolute;
  inset: 0;
}
.hero-section__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.hero-section .container {
  position: relative;
  z-index: 1;
  max-width: 720px;
}
.hero-section__body {
  padding: 2rem;
  background: rgba(255,255,255,0.92);
  border-radius: var(--radius);
  backdrop-filter: blur(4px);
}
.hero-section h1 { margin-bottom: 0.75rem; }
.hero-subtitle, .hero-section .lead {
  font-size: 1.125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 0 1.5rem;
}
.hero-actions {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
}

/* ── Category tiles — large image cards with overlay text ── */
.category-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}
.category-tile {
  position: relative;
  aspect-ratio: var(--category-aspect);
  border-radius: var(--radius);
  overflow: hidden;
  text-decoration: none;
  background: var(--color-surface);
  display: block;
}
.category-tile img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition);
}
.category-tile:hover img { transform: scale(1.05); }
.category-tile__overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.7), rgba(0,0,0,0) 50%);
  display: flex;
  align-items: flex-end;
  padding: 1.25rem;
}
.category-tile__name {
  color: #ffffff;
  font-size: 1.125rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}

/* ── Product grid — lifestyle photography, cover-fit ── */
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--product-min), 1fr));
  gap: 1.5rem 1rem;
}

/* Product card — image dominates, body tight below */
.product-card {
  background: transparent;
  border: none;
  border-radius: 0;
  overflow: visible;
  text-decoration: none;
  color: inherit;
  position: relative;
  display: block;
}
.product-card__image {
  aspect-ratio: var(--product-aspect);
  overflow: hidden;
  border-radius: var(--radius);
  background: var(--color-product-bg);
  margin-bottom: 0.75rem;
  position: relative;
}
.product-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition), opacity var(--transition);
}
.product-card:hover .product-card__image img { transform: scale(1.04); }

/* Second image overlay (shown on hover if provided) */
.product-card__image--hover-swap::before {
  content: "";
  position: absolute;
  inset: 0;
  background-image: var(--hover-image);
  background-size: cover;
  background-position: center;
  opacity: 0;
  transition: opacity var(--transition);
  z-index: 1;
}
.product-card:hover .product-card__image--hover-swap::before { opacity: 1; }

/* Badges */
.product-card__badges {
  position: absolute;
  top: 0.625rem;
  left: 0.625rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  z-index: 2;
}
.product-card__badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  font-size: 0.6875rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: 2px;
}
.product-card__badge--sale { background: var(--color-sale); color: #ffffff; }
.product-card__badge--new { background: var(--color-primary); color: var(--color-primary-text); }
.product-card__badge--sold-out {
  background: var(--color-sold-out);
  color: #ffffff;
}

/* Quick-actions overlay on hover */
.product-card__quick-actions {
  position: absolute;
  left: 0.5rem;
  right: 0.5rem;
  bottom: 0.5rem;
  z-index: 2;
  transform: translateY(8px);
  opacity: 0;
  transition: opacity var(--transition), transform var(--transition);
}
.product-card:hover .product-card__quick-actions {
  opacity: 1;
  transform: translateY(0);
}
.product-card__quick-actions .btn {
  width: 100%;
  font-size: 0.8125rem;
  padding: 0.5rem;
  min-height: 40px;
  background: rgba(255,255,255,0.97);
  color: var(--color-primary);
  backdrop-filter: blur(4px);
}
.product-card__quick-actions .btn:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
}

/* Card body */
.product-card__name {
  font-size: 0.9375rem;
  font-weight: 500;
  margin: 0 0 0.25rem;
  line-height: 1.3;
  color: var(--color-text);
}
.product-card__category {
  font-size: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.375rem;
}
.product-card__price {
  font-size: 0.9375rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.product-card__price-sale { color: var(--color-sale); }
.product-card__price-original {
  color: var(--section-text-muted, var(--color-text-muted));
  font-weight: 400;
  text-decoration: line-through;
  margin-left: 0.375rem;
  font-size: 0.875rem;
}
.product-card__swatches {
  display: flex;
  gap: 0.25rem;
  margin-top: 0.5rem;
}
.swatch {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 1px solid rgba(0,0,0,0.15);
  display: inline-block;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Features — concise USPs */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
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
  text-align: center;
}
.feature-icon,
.service-icon,
.differentiator-icon {
  width: 40px;
  height: 40px;
  margin: 0 auto 0.75rem;
  color: var(--color-primary);
}

/* ── Newsletter marketing section ── */
.newsletter-section {
  padding: 3rem var(--container-pad-x);
  background: var(--color-surface);
  text-align: center;
}
.newsletter-section .container { max-width: 560px; }
.newsletter-section h2 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}
.newsletter-section p {
  color: var(--section-text-muted, var(--color-text-muted));
  margin-bottom: 1.5rem;
}
.newsletter-form {
  display: flex;
  gap: 0.5rem;
  max-width: 440px;
  margin-inline: auto;
}
.newsletter-form input[type="email"] {
  flex: 1;
  padding: 0.75rem 1rem;
  font: inherit;
  font-size: 0.9375rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-card-bg);
  min-height: 44px;
  min-width: 0;
}

/* About / FAQ */
.about-section .container { max-width: 820px; }
.faq-section .container { max-width: 820px; }

.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1rem 0;
}
.faq-item summary {
  cursor: pointer;
  font-weight: 500;
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
  font-size: 1.25rem;
  font-weight: 300;
}
.faq-item[open] summary::after { content: "−"; }

/* CTA / testimonials */
.call-to-action-section { text-align: center; padding-block: 3rem; }
.testimonials-section { padding-block: 3rem; }
.testimonials-section .testimonial {
  max-width: 680px;
  margin-inline: auto;
  text-align: center;
  font-size: 1.0625rem;
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
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 15%, transparent);
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font: inherit;
  font-weight: 500;
  font-size: 0.9375rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), color var(--transition), border-color var(--transition);
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
  background: var(--color-card-bg);
  color: var(--section-heading, var(--color-primary));
  border-color: var(--color-primary);
}
.btn-secondary:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-large { padding: 0.875rem 2rem; font-size: 1rem; min-height: 48px; }

/* ── Site footer ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  padding: 3.5rem 0 0;
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
  margin-bottom: 1rem;
}
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: #ffffff; text-decoration: underline; }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; font-size: 0.875rem; }

/* Payment methods row */
.payment-methods {
  grid-column: 1 / -1;
  padding: 1.5rem 0;
  margin-top: 1.5rem;
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}
.payment-icons {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  list-style: none;
  padding: 0;
  margin: 0;
}
.payment-icons li {
  width: 36px;
  height: 24px;
  background: rgba(255,255,255,0.1);
  border-radius: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.payment-icons img {
  max-width: 100%;
  max-height: 100%;
  height: auto;
}

.footer-bottom {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 1.25rem var(--container-pad-x);
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.8125rem;
  color: rgba(255,255,255,0.5);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .category-grid { grid-template-columns: repeat(2, 1fr); }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .trust-bar__inner { gap: 0.75rem; font-size: 0.75rem; }
  .trust-item { flex: 1 1 calc(50% - 0.5rem); justify-content: center; }
  .hero-section__body { padding: 1.5rem; }
  :root { --product-min: 160px; }
  .category-grid { grid-template-columns: repeat(2, 1fr); }
  .newsletter-form { flex-direction: column; }
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-header-bg); border-bottom: 1px solid var(--color-border); padding: 0.75rem var(--container-pad-x); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .mobile-menu-toggle { display: inline-flex; }
  .mini-cart { width: calc(100vw - 2rem); right: 1rem; left: 1rem; }
  .product-card__quick-actions {
    opacity: 1;
    transform: none;
    position: static;
    margin-top: 0.5rem;
  }
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
