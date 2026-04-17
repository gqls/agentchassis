-- =====================================================================
-- LAYOUT SEED: media-grid
-- =====================================================================
-- Character: thumbnail-dominant, continuous scroll, discovery through
--            browsing. The grid IS the content.
-- Mapped themes: none.
-- Default typography: sans-modern (Inter).
-- Default header/footer: header-with-search (new), existing minimal.
--
-- STRUCTURAL DIVERGENCE:
--   - No hero in the brochure sense. An optional featured/pinned item
--     spans the container width; otherwise the filter bar sits directly
--     below the header and the grid starts immediately
--   - Grid is auto-fill with minimum thumbnail width; columns scale
--     fluidly (no fixed 3-col)
--   - Thumbnails dominate the card (image is ~75% of card height);
--     title, channel/author, views, duration overlay live outside
--     and inside the image frame respectively
--   - Horizontal chip filter bar at top; scrollable on mobile
--   - "Featured row" variant (single wide card) + "row" variant
--     (horizontal-scroll shelf) alongside the default grid
--   - Fixed aspect ratio tokens: 16:9 for video, 1:1 for audio/podcast
--   - Hover is subtle — scale + info reveal, not lift
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
    'media-grid',
    'Media — Discovery Grid',
    'Thumbnail-dominant layout for media libraries. Auto-fill grid with fluid column count, aspect-locked thumbnails, duration overlays, horizontal-scroll shelves, chip-filter bar, and optional featured-item hero. Suits video platforms, audio libraries, image galleries, podcast catalogues.',
    'media',
    ARRAY['video', 'audio', 'podcast', 'gallery', 'media-library', 'streaming'],
    '{
        "container_max_width": "1440px",
        "container_padding_x": "1.5rem",
        "section_padding_y": "2rem",
        "section_padding_y_mobile": "1.5rem",
        "thumbnail_min_width": "260px",
        "thumbnail_aspect_video": "16 / 9",
        "thumbnail_aspect_square": "1 / 1",
        "grid_gap_x": "1rem",
        "grid_gap_y": "2rem",
        "border_radius": "0.75rem",
        "border_radius_sm": "0.375rem",
        "overlay_scrim": "linear-gradient(to top, rgba(0,0,0,0.85), transparent 60%)",
        "transition_base": "200ms ease",
        "chip_height": "36px",
        "header_height": "64px"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: media-grid
 *
 * Discovery grammar: grid first, chrome minimal. The user is scanning,
 * not reading.
 * ===================================================================== */

:root {
  /* ── Palette ── */
  --color-primary:        {{palette "primary"        "#ef4444"}};
  --color-primary-hover:  {{palette "primary_hover"  "#dc2626"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#262626"}};
  --color-accent:         {{palette "accent"         "#ef4444"}};
  --color-background:     {{palette "background"     "#0f0f0f"}};
  --color-surface:        {{palette "surface"        "#1a1a1a"}};
  --color-text:           {{palette "text"           "#f5f5f5"}};
  --color-text-muted:     {{palette "text_muted"     "#a3a3a3"}};
  --color-border:         {{palette "border"         "#262626"}};
  --color-card-bg:        {{palette "card_bg"        "#1a1a1a"}};
  --color-header-bg:      {{palette "header_bg"      "#0f0f0f"}};
  --color-header-text:    {{palette "header_text"    "#f5f5f5"}};
  --color-cta-bg:         {{palette "cta_bg"         "#ef4444"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-footer-bg:      {{palette "footer_bg"      "#0f0f0f"}};
  --color-footer-text:    {{palette "footer_text"    "#a3a3a3"}};
  --color-chip-bg:        {{palette "chip_bg"        "#262626"}};
  --color-chip-bg-active: {{palette "chip_bg_active" "#f5f5f5"}};
  --color-chip-text:      {{palette "chip_text"      "#f5f5f5"}};
  --color-chip-text-active: {{palette "chip_text_active" "#0f0f0f"}};

  /* ── Typography ── */
  --font-body:        {{typo "font_family"  "'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"}};
  --font-heading:     {{typo "heading_font" "inherit"}};
  --font-size-base:   {{typo "base_size"    "15px"}};
  --line-height-base: {{typo "line_height"  "1.5"}};

  /* ── Structure ── */
  --container-max:         {{token "container_max_width"      "1440px"}};
  --container-pad-x:       {{token "container_padding_x"      "1.5rem"}};
  --section-pad-y:         {{token "section_padding_y"        "2rem"}};
  --section-pad-y-sm:      {{token "section_padding_y_mobile" "1.5rem"}};
  --thumb-min:             {{token "thumbnail_min_width"      "260px"}};
  --aspect-video:          {{token "thumbnail_aspect_video"   "16 / 9"}};
  --aspect-square:         {{token "thumbnail_aspect_square"  "1 / 1"}};
  --grid-gap-x:            {{token "grid_gap_x"               "1rem"}};
  --grid-gap-y:            {{token "grid_gap_y"               "2rem"}};
  --radius:                {{token "border_radius"            "0.75rem"}};
  --radius-sm:             {{token "border_radius_sm"         "0.375rem"}};
  --scrim:                 {{token "overlay_scrim"            "linear-gradient(to top, rgba(0,0,0,0.85), transparent 60%)"}};
  --transition:            {{token "transition_base"          "200ms ease"}};
  --chip-h:                {{token "chip_height"              "36px"}};
  --header-h:              {{token "header_height"            "64px"}};

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
  margin: 0 0 0.5rem;
  line-height: 1.25;
  font-weight: 600;
  letter-spacing: -0.01em;
}
h1 { font-size: 1.75rem; }
h2 { font-size: 1.375rem; }
h3 { font-size: 1rem; font-weight: 500; line-height: 1.35; }
h4 { font-size: 0.9375rem; font-weight: 500; }

p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 0.75rem; }
a {
  color: inherit;
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

/* ── Site header — search-prominent ── */
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
.logo-img { max-height: 32px; width: auto; }
.header-search {
  flex: 1;
  max-width: 600px;
  margin: 0 auto;
  position: relative;
}
.header-search input {
  width: 100%;
  height: 40px;
  padding: 0 2.5rem 0 1rem;
  font: inherit;
  font-size: 0.9375rem;
  border: 1px solid var(--color-border);
  border-radius: calc(var(--radius) * 3);
  background: var(--color-surface);
  color: var(--color-text);
}
.header-search input:focus {
  outline: none;
  border-color: var(--color-primary);
}
.header-search .search-submit {
  position: absolute;
  right: 4px;
  top: 4px;
  bottom: 4px;
  width: 32px;
  background: transparent;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  min-width: 44px;
  min-height: 32px;
}
.main-nav ul {
  display: flex;
  gap: 1rem;
  list-style: none;
  margin: 0;
  padding: 0;
}
.main-nav a {
  color: var(--color-header-text);
  font-weight: 500;
  font-size: 0.875rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-sm);
}
.main-nav a:hover,
.main-nav a.active { background: var(--color-surface); color: var(--color-text); }
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

/* ── Filter chip bar ── */
.filter-bar {
  position: sticky;
  top: var(--header-h);
  background: var(--color-background);
  z-index: 900;
  padding: 0.75rem 0;
  border-bottom: 1px solid var(--color-border);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}
.filter-bar::-webkit-scrollbar { display: none; }
.filter-bar ul {
  display: flex;
  gap: 0.5rem;
  list-style: none;
  padding: 0 var(--container-pad-x);
  margin: 0;
  max-width: var(--container-max);
  margin-inline: auto;
  white-space: nowrap;
}
.chip {
  display: inline-flex;
  align-items: center;
  height: var(--chip-h);
  padding: 0 1rem;
  font: inherit;
  font-size: 0.8125rem;
  font-weight: 500;
  background: var(--color-chip-bg);
  color: var(--color-chip-text);
  border: none;
  border-radius: calc(var(--chip-h) / 2);
  cursor: pointer;
  text-decoration: none;
  transition: background var(--transition);
  min-height: 44px;
}
.chip:hover { background: color-mix(in srgb, var(--color-chip-bg) 80%, var(--color-text-muted)); }
.chip.is-active {
  background: var(--color-chip-bg-active);
  color: var(--color-chip-text-active);
}

/* ── Hero — optional featured item, full-width card ── */
.hero-section {
  padding: 1.5rem 0 0;
}
.featured-item {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 2rem;
  background: var(--color-card-bg);
  border-radius: var(--radius);
  overflow: hidden;
}
.featured-item__thumb {
  aspect-ratio: var(--aspect-video);
  background: var(--color-surface);
  overflow: hidden;
  position: relative;
}
.featured-item__thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.featured-item__body {
  padding: 1.5rem 2rem;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
.featured-item__eyebrow {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--color-primary);
  margin-bottom: 0.5rem;
}
.featured-item h1 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}
.featured-item__meta {
  color: var(--section-text-muted, var(--color-text-muted));
  font-size: 0.875rem;
}

/* ── Renderer-managed surface sections ──
 *
 * TEMPORARY RENDERER COUPLING: Phase 4.5 pending. */
.features-section,
.services-section,
.differentiators-section,
.about-section,
.faq-section { background: var(--color-surface); }

/* ── The grid — auto-fill, fluid columns ── */
.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(var(--thumb-min), 1fr));
  column-gap: var(--grid-gap-x);
  row-gap: var(--grid-gap-y);
}

/* Horizontal scrolling shelf variant — for "Trending", "Recommended" */
.media-shelf {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: var(--thumb-min);
  gap: var(--grid-gap-x);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scroll-snap-type: x mandatory;
  padding-bottom: 0.5rem;
  scrollbar-width: thin;
}
.media-shelf > .media-card { scroll-snap-align: start; }
.shelf-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 1rem;
  gap: 1rem;
}
.shelf-head h2 { margin: 0; font-size: 1.125rem; }
.shelf-head .see-more { font-size: 0.8125rem; color: var(--color-text-muted); }

/* ── Media card ── */
.media-card {
  background: transparent;
  border-radius: 0;
  overflow: visible;
  cursor: pointer;
  min-width: 0;
}
.media-card__thumb {
  aspect-ratio: var(--aspect-video);
  background: var(--color-surface);
  overflow: hidden;
  border-radius: var(--radius);
  position: relative;
  margin-bottom: 0.625rem;
}
.media-card__thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition);
}
.media-card:hover .media-card__thumb img { transform: scale(1.03); }

/* Duration badge overlaid on thumb */
.media-card__duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0,0,0,0.85);
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 500;
  padding: 2px 6px;
  border-radius: 3px;
  font-variant-numeric: tabular-nums;
}
/* Scrim overlay (optional, e.g. for play icon affordance) */
.media-card__scrim {
  position: absolute;
  inset: 0;
  background: var(--scrim);
  opacity: 0;
  transition: opacity var(--transition);
}
.media-card:hover .media-card__scrim { opacity: 1; }

/* Square aspect variant (audio/podcast) */
.media-card--square .media-card__thumb { aspect-ratio: var(--aspect-square); }

/* Card body — tight, outside the frame */
.media-card__title {
  font-size: 0.9375rem;
  font-weight: 500;
  line-height: 1.35;
  color: var(--section-heading, var(--color-primary));
  margin-bottom: 0.25rem;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.media-card__channel {
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  margin-bottom: 0.125rem;
}
.media-card__channel a { color: inherit; }
.media-card__channel a:hover { color: var(--color-text); }
.media-card__stats {
  font-size: 0.8125rem;
  color: var(--section-text-muted, var(--color-text-muted));
  font-variant-numeric: tabular-nums;
}

/* Features/services rendered as a content-oriented grid on this layout
   — quieter than in brochure-* layouts */
.features-grid,
.services-grid,
.differentiators-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 1.25rem;
}
.feature-card,
.service-card,
.differentiator-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 1.25rem;
}
.feature-card h3,
.service-card h3,
.differentiator-card h3 { margin-bottom: 0.5rem; }

/* About */
.about-section .container { max-width: 720px; }

/* FAQ */
.faq-section .container { max-width: 720px; }
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
.faq-item summary::after { content: "+"; color: var(--color-text-muted); }
.faq-item[open] summary::after { content: "−"; }

/* CTA / testimonials — quiet on this layout */
.call-to-action-section { padding-block: 2rem; text-align: center; }
.testimonials-section { padding-block: 2rem; }

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
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-surface);
  color: var(--color-text);
  min-height: 44px;
}
.form-field input:focus,
.form-field textarea:focus,
.form-field select:focus {
  outline: none;
  border-color: var(--color-primary);
}

/* Buttons — pill-shaped like chips, to fit the grid aesthetic */
.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  font: inherit;
  font-weight: 500;
  font-size: 0.9375rem;
  border-radius: calc(var(--radius) * 3);
  border: 1px solid transparent;
  cursor: pointer;
  text-decoration: none;
  min-height: 44px;
  transition: background var(--transition), border-color var(--transition), color var(--transition);
}
.btn-primary {
  background: var(--color-primary);
  color: var(--color-primary-text);
}
.btn-primary:hover { background: var(--color-primary-hover); }
.btn-secondary {
  background: var(--color-surface);
  color: var(--color-text);
  border-color: var(--color-border);
}
.btn-secondary:hover { background: var(--color-chip-bg); }

/* Site footer — minimal, matches header weight */
.site-footer {
  background: var(--color-footer-bg);
  color: var(--color-footer-text);
  border-top: 1px solid var(--color-border);
  padding: 2rem 0;
  margin-top: auto;
  font-size: 0.8125rem;
}
.footer-container {
  max-width: var(--container-max);
  margin-inline: auto;
  padding: 0 var(--container-pad-x);
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-text); }
.site-footer ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

/* Responsive */
@media (max-width: 1024px) {
  .featured-item { grid-template-columns: 1fr; }
  .featured-item__body { padding: 1.25rem; }
}
@media (max-width: 768px) {
  .section { padding-block: var(--section-pad-y-sm); }
  :root { --thumb-min: 160px; --grid-gap-y: 1.25rem; }
  .header-search { display: none; }
  .main-nav { display: none; }
  .main-nav.is-open { display: block; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-header-bg); border-bottom: 1px solid var(--color-border); padding: 0.5rem var(--container-pad-x); }
  .main-nav.is-open ul { flex-direction: column; gap: 0; }
  .mobile-menu-toggle { display: inline-flex; }
  .footer-container { flex-direction: column; text-align: center; }
  .features-grid,
  .services-grid,
  .differentiators-grid { grid-template-columns: 1fr; }
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
    scroll-behavior: auto !important;
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
