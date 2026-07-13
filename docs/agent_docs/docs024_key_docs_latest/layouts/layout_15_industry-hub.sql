-- =====================================================================
-- LAYOUT SEED: industry-hub
-- =====================================================================
-- Character: independent authority on a vertical. Not a participant,
--            an information resource ABOUT the industry. Directory
--            as primary element, topic-organised guide index,
--            regulatory news, glossary, heavy disclaimer footer.
-- Mapped themes: none.
-- Default typography: serif-editorial (Merriweather headings + Lato
--                     body) — authority without being corporate.
-- Default header/footer: header-with-categories (new), footer-with-
--                        disclaimer (new).
--
-- DIVERGENCE from the four other commerce-adjacent layouts:
--   - comparison-aggregator: result-card with headline metric;
--     decision-to-transact user intent
--   - affiliate-hub: product card with affiliate CTA; editorial
--     commerce user intent
--   - ecommerce-storefront: product card for catalogue; buy user
--     intent
--   - tool-first-landing: page IS a tool
--   - THIS (industry-hub): .directory-card (provider/supplier row)
--     + .guide-card (topic-grouped, not chronological) +
--     .news-card (date-prominent, compact) + .glossary-list
--     (term + definition). Information architecture is the grammar.
--   - "About this site" banner early on — the independence claim
--     is structural, not buried in a footer link.
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
    'industry-hub',
    'Industry — Information Hub',
    'Vertical information hub positioning the site as an independent resource, with directory of providers as primary content, topic-organised guide index, regulatory news section, glossary, and heavy disclaimer footer. Distinguished from affiliate and comparison layouts by not being commercial. Suits regulatory information hubs, industry explainer sites, trade directories for verticals the operator does not participate in.',
    'hub',
    ARRAY['industry-hub', 'directory', 'regulatory', 'information-resource', 'vertical-guide'],
    '{
        "container_max_width": "1200px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "3.5rem",
        "section_padding_y_mobile": "2rem",
        "directory_card_min_width": "280px",
        "news_card_min_width": "260px",
        "border_radius": "0.375rem",
        "border_radius_sm": "0.25rem",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.04)",
        "shadow_md": "0 4px 10px rgba(0,0,0,0.06)",
        "transition_base": "180ms ease",
        "card_padding": "1.5rem",
        "header_height": "64px",
        "category_strip_height": "44px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: industry-hub
 *
 * Grammar: authority + navigation. Directory first (primary value),
 * then guides (topic-grouped), then news, then reference — ordered by
 * user intent frequency. Independence claim is visible, not buried.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#1e40af"}};
  --color-primary-hover:  {{palette "primary_hover"  "#1e3a8a"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#0f172a"}};
  --color-accent:         {{palette "accent"         "#1e40af"}};
  --color-background:     {{palette "background"     "#ffffff"}};
  --color-surface:        {{palette "surface"        "#f8fafc"}};
  --color-text:           {{palette "text"           "#1e293b"}};
  --color-text-muted:     {{palette "text_muted"     "#64748b"}};
  --color-border:         {{palette "border"         "#e2e8f0"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#ffffff"}};
  --color-header-text:    {{palette "header_text"    "#1e293b"}};
  --color-cta-bg:         {{palette "cta_bg"         "#1e40af"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#0f172a"}};
  --color-footer-text:    {{palette "footer_text"    "rgba(255,255,255,0.75)"}};
  --color-independence-bg:     {{palette "independence_bg"     "#f0f9ff"}};
  --color-independence-border: {{palette "independence_border" "#0284c7"}};
  --color-news-date:           {{palette "news_date"           "#b91c1c"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Lato', Georgia, 'Times New Roman', serif"}};
  --font-heading:     {{typo "heading_font" "'Merriweather', Georgia, 'Times New Roman', serif"}};
  --font-size-base:   {{typo "base_size"    "16px"}};
  --line-height-base: {{typo "line_height"  "1.65"}};

  /* ── Structure ── */
  --container-max:      {{token "container_max_width"      "1200px"}};
  --container-pad-x:    {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:      {{token "section_padding_y"        "3.5rem"}};
  --section-pad-y-sm:   {{token "section_padding_y_mobile" "2rem"}};
  --directory-min:      {{token "directory_card_min_width" "280px"}};
  --news-min:           {{token "news_card_min_width"      "260px"}};
  --radius:             {{token "border_radius"            "0.375rem"}};
  --radius-sm:          {{token "border_radius_sm"         "0.25rem"}};
  --shadow-sm:          {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.04)"}};
  --shadow-md:          {{token "shadow_md"                "0 4px 10px rgba(0,0,0,0.06)"}};
  --transition:         {{token "transition_base"          "180ms ease"}};
  --card-pad:           {{token "card_padding"             "1.5rem"}};
  --header-h:           {{token "header_height"            "64px"}};
  --cat-strip-h:        {{token "category_strip_height"    "44px"}};

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
  letter-spacing: -0.005em;
}
h1 { font-size: clamp(2rem, 3.5vw, 2.5rem); line-height: 1.15; }
h2 { font-size: 1.625rem; }
h3 { font-size: 1.25rem; }
h4 { font-size: 1.0625rem; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
blockquote {
  font-style: italic;
  border-left: 3px solid var(--color-primary);
  padding-left: 1.25rem;
  margin: 1.5rem 0;
  color: var(--section-text-muted, var(--color-text-muted));
}
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
.section__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-primary);
}
.section__head h2 { margin: 0; font-size: 1.375rem; }
.section__head .view-all {
  font-size: 0.875rem;
  font-family: var(--font-body);
  font-weight: 600;
  border-bottom: none;
}

/* ── Site header ── */
.site-header {
  background: var(--color-header-bg);
  color: var(--color-header-text);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 1000;
}
.site-header__main {
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
  justify-content: space-between;
  gap: 2rem;
  width: 100%;
}
.logo {
  font-family: var(--font-heading);
  font-size: 1.5rem;
  font-weight: 900;
  letter-spacing: -0.01em;
  color: var(--color-primary);
  text-decoration: none;
  border-bottom: none;
}
.logo-img { max-height: 44px; width: auto; }

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
  font-size: 0.9375rem;
  border-bottom: none;
  padding: 0.5rem 0;
}
.main-nav a:hover,
.main-nav a.active { color: var(--color-primary); }

/* Category strip — below main header */
.category-strip {
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
  height: var(--cat-strip-h);
  display: flex;
  align-items: center;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}
.category-strip::-webkit-scrollbar { display: none; }
.category-strip ul {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  gap: 1.5rem;
  list-style: none;
  white-space: nowrap;
}
.category-strip a {
  color: var(--color-text);
  font-size: 0.8125rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: none;
  padding: 0.25rem 0;
}
.category-strip a:hover,
.category-strip a.active { color: var(--color-primary); }

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

/* ── Hero — positioning statement, prominent search/browse ── */
.hero-section {
  padding-block: 4rem;
  background: var(--color-surface);
  text-align: center;
}
.hero-section .container { max-width: 820px; }
.hero-section h1 {
  margin-bottom: 1rem;
  font-size: clamp(2rem, 4vw, 2.75rem);
}
.hero-subtitle, .hero-section .lead {
  font-size: 1.125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  max-width: 640px;
  margin: 0 auto 1.75rem;
  line-height: 1.55;
}
.hero-cta-group {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  flex-wrap: wrap;
}

/* ── Independence / "about this site" callout banner ── */
.independence-banner {
  background: var(--color-independence-bg);
  border-left: 4px solid var(--color-independence-border);
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  padding: 1rem 1.25rem;
  margin: 2rem auto;
  max-width: var(--container-max);
  margin-inline: auto;
  font-size: 0.9375rem;
}
.independence-banner__title {
  font-weight: 700;
  margin-bottom: 0.25rem;
  font-family: var(--font-heading);
  font-size: 1rem;
}
.independence-banner p { margin: 0; color: var(--color-text); }

/* ── Directory listing — the primary content element ── */
.directory-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--directory-min), 1fr));
  gap: 1rem;
}
.directory-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--card-pad);
  transition: border-color var(--transition), box-shadow var(--transition);
}
.directory-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}
.directory-card__name {
  font-family: var(--font-heading);
  font-size: 1.0625rem;
  font-weight: 700;
  margin-bottom: 0.375rem;
  line-height: 1.3;
}
.directory-card__name a {
  color: var(--color-text);
  border-bottom: none;
}
.directory-card__name a:hover { color: var(--color-primary); }
.directory-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-bottom: 0.75rem;
}
.directory-card__tag {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  background: var(--color-surface);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  font-weight: 500;
}
.directory-card__description {
  font-size: 0.9375rem;
  line-height: 1.5;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0 0 0.75rem;
}
.directory-card__details-link {
  font-size: 0.875rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border-bottom: none;
}
.directory-card__details-link:hover { text-decoration: underline; }

/* Filter bar above directory (simpler than comparison-aggregator's) */
.directory-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
  margin-bottom: 1.5rem;
  padding: 0.875rem 1rem;
  background: var(--color-surface);
  border-radius: var(--radius);
}
.directory-filter label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--section-text-muted, var(--color-text-muted));
}
.directory-filter select {
  padding: 0.375rem 0.625rem;
  font: inherit;
  font-size: 0.875rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-background);
  min-height: 36px;
}

/* ── Guide index — topic-organised, NOT chronological ── */
.guide-topics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 2rem;
  margin-top: 1.5rem;
}
.guide-topic {
  background: var(--color-card-bg);
  padding: var(--card-pad);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
}
.guide-topic__name {
  font-family: var(--font-heading);
  font-size: 1.125rem;
  font-weight: 700;
  margin-bottom: 0.75rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid var(--color-primary);
}
.guide-topic ul {
  list-style: none;
  padding: 0;
  margin: 0;
}
.guide-topic li {
  padding: 0.375rem 0;
  border-bottom: 1px solid var(--color-border);
}
.guide-topic li:last-child { border-bottom: none; }
.guide-topic a {
  display: block;
  font-size: 0.9375rem;
  line-height: 1.4;
  padding: 0.25rem 0;
  border-bottom: none;
  min-height: 32px;
}
.guide-topic__description {
  display: block;
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  font-weight: 400;
  margin-top: 0.125rem;
}

/* ── News section — date-prominent, compact ── */
.news-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--news-min), 1fr));
  gap: 1.25rem;
  margin-top: 1.5rem;
}
.news-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.125rem 1.25rem;
  transition: border-color var(--transition);
}
.news-card:hover { border-color: var(--color-primary); }
.news-card__date {
  display: block;
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--color-news-date);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.5rem;
  font-variant-numeric: tabular-nums;
}
.news-card__title {
  font-family: var(--font-heading);
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.3;
  margin-bottom: 0.375rem;
}
.news-card__title a { color: var(--color-text); border-bottom: none; }
.news-card__title a:hover { color: var(--color-primary); }
.news-card__excerpt {
  font-size: 0.875rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin: 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ── Glossary / reference list ── */
.glossary-section .container { max-width: 920px; }
.glossary-list {
  list-style: none;
  padding: 0;
  margin: 0;
}
.glossary-list > li {
  padding: 1rem 0;
  border-bottom: 1px solid var(--color-border);
}
.glossary-list > li:last-child { border-bottom: none; }
.glossary-term {
  font-family: var(--font-heading);
  font-weight: 700;
  font-size: 1.0625rem;
  margin-bottom: 0.25rem;
  display: block;
}
.glossary-definition {
  margin: 0;
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.9375rem;
  line-height: 1.55;
}

/* Alphabetic jumplist option */
.glossary-jumplist {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-bottom: 1.5rem;
}
.glossary-jumplist a {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  min-width: 32px;
  min-height: 32px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-weight: 600;
  font-size: 0.875rem;
  text-transform: uppercase;
  border-bottom: 1px solid var(--color-border);
}
.glossary-jumplist a:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
  border-color: var(--color-primary);
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* Features — when present, simple three-col list */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
  margin-top: 1.5rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-card-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: var(--card-pad);
}

/* About / Contact / FAQ */
.about-section .container,
.contact-section .container { max-width: 820px; }
.faq-section .container { max-width: 820px; }

.faq-item {
  border-bottom: 1px solid var(--color-border);
  padding: 1rem 0;
}
.faq-item summary {
  cursor: pointer;
  font-family: var(--font-heading);
  font-weight: 700;
  list-style: none;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  min-height: 44px;
  align-items: center;
}
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; font-weight: 400; font-size: 1.25rem; color: var(--color-primary); }
.faq-item[open] summary::after { content: "−"; }

/* CTA / testimonials */
.call-to-action-section { text-align: center; padding-block: 2.5rem; }
.testimonials-section { padding-block: 2.5rem; }
.testimonials-section .testimonial {
  max-width: 720px;
  margin-inline: auto;
  text-align: center;
  font-family: var(--font-heading);
  font-style: italic;
  font-size: 1.125rem;
  line-height: 1.4;
}

/* Forms */
.form-field { margin-bottom: 1rem; }
.form-field label {
  display: block;
  font-weight: 600;
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
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary) 20%, transparent);
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font: inherit;
  font-family: var(--font-body);
  font-weight: 700;
  font-size: 0.9375rem;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  border-bottom: none;
  transition: background var(--transition), color var(--transition), border-color var(--transition);
}
.btn:hover { text-decoration: none; border-bottom: none; }
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
  color: var(--color-primary);
  border-color: var(--color-primary);
}
.btn-secondary:hover {
  background: var(--color-primary);
  color: var(--color-primary-text);
}

/* ── Site footer — dark, with heavy disclaimer block ── */
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
  font-family: var(--font-heading);
  font-size: 0.9375rem;
  margin-bottom: 0.75rem;
}
.site-footer a { color: var(--color-footer-text); border-bottom: none; }
.site-footer a:hover { color: #ffffff; text-decoration: underline; }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.5rem; font-size: 0.9rem; }

/* The defining feature: prominent heavy disclaimer block */
.footer-disclaimer {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 2rem var(--container-pad-x);
  margin-top: 2.5rem;
  border-top: 1px solid rgba(255,255,255,0.15);
  font-size: 0.8125rem;
  line-height: 1.65;
  color: rgba(255,255,255,0.65);
}
.footer-disclaimer__title {
  color: #ffffff;
  font-weight: 700;
  font-family: var(--font-heading);
  margin-bottom: 0.75rem;
  font-size: 1rem;
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
  .guide-topics { grid-template-columns: repeat(2, 1fr); }
  .footer-container { grid-template-columns: repeat(2, 1fr); }
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  .hero-section { padding-block: 2.5rem; }
  .directory-grid,
  .guide-topics,
  .news-grid,
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
