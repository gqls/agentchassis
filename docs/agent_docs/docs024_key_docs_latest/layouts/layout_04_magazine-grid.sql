-- =====================================================================
-- LAYOUT SEED: magazine-grid
-- =====================================================================
-- Phase 3 of 025_palette_layout_typography_migration.
--
-- Character: publication feel. Article cards, sidebar with widgets,
--            featured article spanning full content width, newsletter
--            capture. Content density is the character.
-- Mapped themes: content-modern.
-- Default typography: serif-editorial (Merriweather headings + Lato/
--                     Georgia body).
-- Default header/footer: header-with-categories (new), footer-4-column.
--
-- STRUCTURAL DIVERGENCE from brochure-* and portfolio-kinetic:
--   - Main content is a 2/3 + 1/3 grid (main column + sidebar) at
--     desktop; this is a top-level layout primitive, not a section
--     variant
--   - Article cards with image-top, category badge, title, excerpt
--     (2-line clamp), author/date meta — a component shape
--   - "Featured article" variant spans full main-column width with
--     image-and-text side-by-side
--   - Sidebar widgets: popular-posts numbered list, categories,
--     newsletter signup, ad zone placeholder
--   - Serif editorial typography assumption bakes into heading sizes
--     and line-heights
--   - Section headers carry a "View all →" affordance aligned right
--
-- ----- CONTRACT CHECKS -----
-- 1. Colour Inheritance Model honoured
-- 2. No --section-* defaults on section containers (except conditional
--    root-level --section-heading)
-- 3. Renderer-managed surface classes painted (Phase 4.5 pending)
-- 4. Every helper call has a fallback
-- 5. Responsive: 1024/768 breakpoints, touch targets >= 44px
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
    'magazine-grid',
    'Magazine — Editorial Grid',
    'Publication layout with featured article, main 2/3 + 1/3 sidebar grid, article cards with image-top + category badge + excerpt + author meta, sidebar widgets (popular, categories, newsletter, ad zone), and serif editorial typography. Suits news, opinion, long-form blogs, and curated content networks.',
    'editorial',
    ARRAY['publication', 'news', 'blog', 'opinion', 'long-form', 'editorial'],
    '{
        "container_max_width": "1280px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "3.5rem",
        "section_padding_y_mobile": "2rem",
        "main_column_ratio": "2fr",
        "sidebar_column_ratio": "1fr",
        "main_sidebar_gap": "3rem",
        "card_grid_gap": "2rem",
        "card_image_aspect": "16 / 9",
        "border_radius": "0.375rem",
        "border_radius_sm": "0.25rem",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.05)",
        "shadow_md": "0 4px 12px rgba(0,0,0,0.08)",
        "transition_base": "200ms ease",
        "excerpt_line_clamp": "2",
        "article_max_width": "720px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: magazine-grid
 *
 * Publication grammar:
 *   body > main = [main-column 2fr] [sidebar 1fr] @ desktop
 *   main-column contains stacked section shapes — featured article,
 *     article card grids, section dividers
 *   sidebar contains stacked widgets — popular list, category list,
 *     newsletter form, ad zone
 * Mobile collapses both to single column with sidebar appended below.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#8b0000"}};
  --color-primary-hover:  {{palette "primary_hover"  "#6b0000"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#1a1a1a"}};
  --color-accent:         {{palette "accent"         "#8b0000"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f7f5f2"}};
  --color-text:           {{palette "text"           "#1f2937"}};
  --color-text-muted:     {{palette "text_muted"     "#6b7280"}};
  --color-border:         {{palette "border"         "#e5e7eb"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#1a1a1a"}};
  --color-cta-bg:         {{palette "cta_bg"         "#8b0000"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#1a1a1a"}};
  --color-footer-text:    {{palette "footer_text"    "rgba(255,255,255,0.8)"}};
  --color-badge-bg:       {{palette "badge_bg"       "#8b0000"}};
  --color-badge-text:     {{palette "badge_text"     "#ffffff"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Lato', Georgia, 'Times New Roman', serif"}};
  --font-heading:     {{typo "heading_font" "'Merriweather', Georgia, 'Times New Roman', serif"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.7"}};

  /* ── Structure ── */
  --container-max:     {{token "container_max_width"      "1280px"}};
  --container-pad-x:   {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:     {{token "section_padding_y"        "3.5rem"}};
  --section-pad-y-sm:  {{token "section_padding_y_mobile" "2rem"}};
  --main-col:          {{token "main_column_ratio"        "2fr"}};
  --sidebar-col:       {{token "sidebar_column_ratio"     "1fr"}};
  --main-sidebar-gap:  {{token "main_sidebar_gap"         "3rem"}};
  --card-gap:          {{token "card_grid_gap"            "2rem"}};
  --card-image-aspect: {{token "card_image_aspect"        "16 / 9"}};
  --radius:            {{token "border_radius"            "0.375rem"}};
  --radius-sm:         {{token "border_radius_sm"         "0.25rem"}};
  --shadow-sm:         {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.05)"}};
  --shadow-md:         {{token "shadow_md"                "0 4px 12px rgba(0,0,0,0.08)"}};
  --transition:        {{token "transition_base"          "200ms ease"}};
  --excerpt-clamp:     {{token "excerpt_line_clamp"       "2"}};
  --article-max:       {{token "article_max_width"        "720px"}};

  /* Per Colour Inheritance Model. */
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
  font-weight: 700;
  letter-spacing: -0.005em;
}
h1 { font-size: clamp(2rem, 4vw, 2.75rem); line-height: 1.15; }
h2 { font-size: 1.75rem; }
h3 { font-size: 1.375rem; }
h4 { font-size: 1.125rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
blockquote {
  font-style: italic;
  font-size: 1.25rem;
  line-height: 1.5;
  border-left: 3px solid var(--color-primary);
  padding: 0 0 0 1.5rem;
  margin: 2rem 0;
}
/* strong/em/cite/span: do NOT set color */
a {
  color: var(--color-primary);
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color var(--transition), color var(--transition);
}
a:hover {
  color: var(--color-primary-hover);
  border-bottom-color: currentColor;
}

/* ── Layout primitives ── */
.container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  width: 100%;
}
.section { padding-block: var(--section-pad-y); }

/* ── The defining two-column body shape ── */
.page-grid {
  display: grid;
  grid-template-columns: var(--main-col) var(--sidebar-col);
  gap: var(--main-sidebar-gap);
  max-width: var(--container-max);
  margin-inline: auto;
  padding-inline: var(--container-pad-x);
  padding-block: var(--section-pad-y);
}
.main-column { min-width: 0; }
.sidebar-column { min-width: 0; }

/* ── Section headers (title + "View all →" affordance) ── */
.section-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 1rem;
  padding-bottom: 0.75rem;
  margin-bottom: 1.5rem;
  border-bottom: 2px solid var(--color-primary);
}
.section-head h2 {
  margin: 0;
  font-size: 1.375rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 700;
}
.section-head .view-all {
  font-size: 0.875rem;
  font-family: var(--font-body);
  color: var(--section-text-muted, var(--color-text-muted));
  font-weight: 600;
  border-bottom: none;
}
.section-head .view-all:hover { color: var(--color-primary); }

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
  font-family: var(--font-heading);
  font-size: 1.75rem;
  font-weight: 900;
  letter-spacing: -0.01em;
  color: var(--color-header-text);
  text-decoration: none;
  border-bottom: none;
}
.logo-img { max-height: 44px; width: auto; }

/* Category strip — extends header visually */
.category-strip {
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}
.category-strip ul {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0.75rem var(--container-pad-x);
  display: flex;
  gap: 1.5rem;
  list-style: none;
  white-space: nowrap;
}
.category-strip a {
  color: var(--color-header-text);
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  border-bottom: none;
  padding: 0.25rem 0;
}
.category-strip a:hover,
.category-strip a.active { color: var(--color-primary); }

.main-nav ul {
  display: flex;
  gap: 1.5rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 600;
  font-size: 0.9rem;
  border-bottom: none;
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

/* ── Hero — a featured article, not a tagline ── */
.hero-section {
  padding-block: 2.5rem;
  border-bottom: 1px solid var(--color-border);
}
.hero-section .container {
  max-width: var(--container-max);
}
.featured-article {
  display: grid;
  grid-template-columns: 3fr 2fr;
  gap: 2.5rem;
  align-items: center;
}
.featured-article__image {
  aspect-ratio: var(--card-image-aspect);
  background: var(--color-surface);
  overflow: hidden;
  border-radius: var(--radius);
}
.featured-article__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition);
}
.featured-article:hover .featured-article__image img { transform: scale(1.02); }
.featured-article__meta {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  font-size: 0.8125rem;
  margin-bottom: 1rem;
}
.featured-article h1 { margin-bottom: 1rem; }
.featured-article__excerpt {
  font-size: 1.0625rem;
  color: var(--section-text-muted, var(--color-text-muted));
  line-height: 1.6;
  margin-bottom: 1rem;
}
.featured-article__byline {
  font-size: 0.875rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Article cards ── */
.article-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--card-gap);
}
.article-grid.cols-3 { grid-template-columns: repeat(3, 1fr); }
.article-grid.cols-1 { grid-template-columns: 1fr; }

.article-card {
  background: var(--color-card-bg);
  border-radius: var(--radius);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: transform var(--transition), box-shadow var(--transition);
}
.article-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}
.article-card__image {
  aspect-ratio: var(--card-image-aspect);
  background: var(--color-surface);
  overflow: hidden;
}
.article-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.article-card__body {
  padding: 1.25rem 1rem 1.5rem;
  display: flex;
  flex-direction: column;
  flex: 1;
}
.article-card__category {
  display: inline-block;
  background: var(--color-badge-bg);
  color: var(--color-badge-text);
  font-size: 0.6875rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-sm);
  margin-bottom: 0.75rem;
  align-self: flex-start;
  text-decoration: none;
  border-bottom: none;
}
.article-card__title {
  font-family: var(--font-heading);
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1.3;
  margin-bottom: 0.5rem;
  color: var(--section-heading, var(--color-primary));
}
.article-card__title a {
  color: inherit;
  border-bottom: none;
}
.article-card__title a:hover { color: var(--color-primary); }
.article-card__excerpt {
  font-size: 0.9375rem;
  color: var(--section-text-muted, var(--color-text-muted));
  line-height: 1.5;
  margin-bottom: 1rem;
  display: -webkit-box;
  -webkit-line-clamp: var(--excerpt-clamp);
  line-clamp: var(--excerpt-clamp);
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.article-card__meta {
  margin-top: auto;
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

/* Compact card variant for "more stories" */
.article-card--compact {
  flex-direction: row;
  align-items: flex-start;
  gap: 1rem;
  background: transparent;
  border-radius: 0;
}
.article-card--compact:hover { transform: none; box-shadow: none; }
.article-card--compact .article-card__image {
  flex: 0 0 120px;
  aspect-ratio: 1 / 1;
}
.article-card--compact .article-card__body {
  padding: 0;
  flex: 1;
}
.article-card--compact .article-card__title { font-size: 1rem; }
.article-card--compact .article-card__excerpt { display: none; }

/* ── Sidebar widgets ── */
.widget {
  margin-bottom: 2.5rem;
}
.widget__title {
  font-size: 0.9375rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding-bottom: 0.5rem;
  margin-bottom: 1rem;
  border-bottom: 2px solid var(--color-primary);
}

/* Popular posts — numbered list */
.popular-list {
  list-style: none;
  padding: 0;
  margin: 0;
  counter-reset: popular;
}
.popular-list li {
  counter-increment: popular;
  display: grid;
  grid-template-columns: 2rem 1fr;
  gap: 0.75rem;
  padding: 1rem 0;
  border-bottom: 1px solid var(--color-border);
}
.popular-list li:last-child { border-bottom: none; }
.popular-list li::before {
  content: counter(popular, decimal-leading-zero);
  font-family: var(--font-heading);
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-primary);
  line-height: 1;
  opacity: 0.4;
}
.popular-list a {
  font-family: var(--font-heading);
  font-weight: 600;
  line-height: 1.3;
  color: var(--color-text);
  border-bottom: none;
}
.popular-list a:hover { color: var(--color-primary); }

/* Category list */
.category-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.category-list li {
  border-bottom: 1px solid var(--color-border);
}
.category-list a {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0;
  font-weight: 600;
  color: var(--color-text);
  min-height: 44px;
  border-bottom: none;
}
.category-list a:hover { color: var(--color-primary); }
.category-list .count {
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  font-weight: 400;
}

/* Newsletter widget */
.newsletter-widget {
  background: var(--color-surface);
  padding: 1.75rem 1.5rem;
  border-radius: var(--radius);
}
.newsletter-widget h3 {
  font-size: 1.125rem;
  margin-bottom: 0.5rem;
}
.newsletter-widget p {
  font-size: 0.9rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-bottom: 1rem;
}
.newsletter-widget input[type="email"] {
  width: 100%;
  padding: 0.75rem;
  font: inherit;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-background);
  min-height: 44px;
  margin-bottom: 0.5rem;
}
.newsletter-widget input[type="email"]:focus {
  outline: none;
  border-color: var(--color-primary);
}
.newsletter-widget .btn { width: 100%; }

/* Ad zone placeholder */
.ad-zone {
  margin-bottom: 2rem;
  padding: 1rem;
  border: 1px dashed var(--color-border);
  text-align: center;
  font-size: 0.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
  text-transform: uppercase;
  letter-spacing: 0.1em;
}
.ad-zone img { margin-inline: auto; }

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 will move these to components
 * with a data-section-bg attribute. Until then, layouts paint. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* About section uses article-body column width */
.about-section .container {
  max-width: var(--article-max);
}

/* FAQ uses magazine accordions */
.faq-section .container { max-width: var(--article-max); }
.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1rem 0;
}
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
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; font-weight: 300; font-size: 1.25rem; }
.faq-item[open] summary::after { content: "−"; }
.faq-item p {
  padding-top: 0.5rem;
  color: var(--section-text-muted, var(--color-text-muted));
}

/* ── Contact ── single column, article-width */
.contact-section .container {
  max-width: var(--article-max);
}

/* ── Forms ── */
.form-field { margin-bottom: 1.25rem; }
.form-field label {
  display: block;
  font-weight: 600;
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
  border-radius: var(--radius-sm);
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

/* ── Buttons ── */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font: inherit;
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 0.875rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), color var(--transition);
  border-bottom: none;
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover {
  background: var(--color-primary-hover);
  color: var(--color-primary-text);
  border-bottom: none;
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

/* ── Pagination / load more ── */
.pagination {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 2.5rem;
}
.pagination a,
.pagination .current {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 44px;
  padding: 0 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-weight: 600;
  text-decoration: none;
  border-bottom: 1px solid var(--color-border);
}
.pagination .current,
.pagination a:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
  border-color: var(--color-primary);
}

/* ── Site footer ── */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  margin-top: auto;
  padding: 3.5rem 0 0;
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
  font-family: var(--font-heading);
  font-size: 1rem;
}
.site-footer a {
  color: var(--color-footer-text);
  border-bottom: none;
}
.site-footer a:hover { color: #ffffff; }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; font-size: 0.9rem; }
.footer-bottom {
  margin-top: 3rem;
  padding: 1.5rem 0;
  border-top: 1px solid rgba(255,255,255,0.1);
  text-align: center;
  font-size: 0.85rem;
  color: rgba(255,255,255,0.6);
}

/* ── Responsive ── */
@media (max-width: 1024px) {
  .page-grid {
    grid-template-columns: 1fr;
    gap: 2.5rem;
  }
  .article-grid.cols-3 { grid-template-columns: repeat(2, 1fr); }
  .featured-article { grid-template-columns: 1fr; }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .page-grid {
    padding-block: var(--section-pad-y-sm);
    gap: 2rem;
  }
  .article-grid,
  .article-grid.cols-3 { grid-template-columns: 1fr; }
  .footer-container { grid-template-columns: 1fr; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; }
  .mobile-menu-toggle { display: inline-flex; }
  .article-card--compact .article-card__image { flex-basis: 100px; }
}

/* ── Accessibility ── */
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
