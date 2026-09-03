-- 736_layout_content_hub_tools_archetype.sql
--
-- bugs_open/445 Phase 5 — the missing archetype: an editorial CONTENT HUB whose
-- core offering is a set of EMBEDDED INTERACTIVE TOOLS.
--
-- WHY THIS LAYOUT EXISTS. Seven live sites resolved to `magazine-grid` on ONE
-- shared tag (`editorial-publication`) at 7-10% coverage of their own declared
-- identity; the terms they emit that NO layout could match were `content-hub`
-- (3 of 7) and `interactive-tools` (2 of 7) — "content hub with embedded tools",
-- the same answer 445 §3 reached from an owner's eye, found independently in
-- the sites' own vocabulary. [MEASURED 2026-09-03], bugs_open/445 §8g.
--
-- WHY THESE TAGS AND NOT OTHERS — CHOSEN BY SIMULATION, NOT TASTE. Four
-- candidate tag sets were scored against all 33 composed sites with a scorer
-- that reproduces the matcher's own recorded score on 29 of 30 (445 §8c), and
-- against a marked-[ASSUMED] proxy for the 17 unbuilt remakes
-- (portfolio_positioning/CONTRIB_2026-09-03_seventeen_remaining_remakes…).
--   A  (445's title words only)  rescued 4 of 7; designblog.co.uk — the site the
--      owner complained about — and apis.uk stayed at 6-8% on a single tag.
--   B  (A + the FORM words the sites already emit: editorial-guides,
--      long-form-content, research-publication, content-platform, guides)
--      rescued 6 of 7: designblog 7%→16%, apis.uk 9%→19%. Pulled in two
--      non-cluster sites, both judged correct (gamedesign.uk: a design-practice
--      publication with interactive illustration; farmerinsurance.uk: self-
--      described "insurance-guidance, editorial-guides, content-platform,
--      interactive-calculators", sitting on industry-hub matching ZERO tags).
--      Its two industry-hub siblings do not move — it is not a wholesale steal.
--   C, D (B + calculator/tools, B + buyers-guide) added nothing over B live.
--   → B. oufe.com is NOT rescued under any candidate: its own tags lead with
--      `interactive-platform`, so tool-portal wins it, defensibly.
--
-- 9 OF THE 18 EXISTING LAYOUTS HAVE ZERO SITES EMITTING ANY OF THEIR TAGS —
-- their tags name an INDUSTRY (wellness, bakery; law, consultancy) while the
-- classifier emits FORM and platform words. So a new layout is only real if the
-- classifier already speaks its tags. The guard below refuses to seed a layout
-- nobody can reach. [MEASURED 2026-09-03] 14 current classification specs
-- emit at least one of these tags (raw, un-canonicalised).
--
-- WHAT MAKES IT VISIBLY NOT ITS SIBLINGS (structural, not palette — the design
-- overlay owns colour):
--   magazine-grid     — permanent 2/3+1/3 sidebar, widget rail, serif body,
--                       badge-heavy, 1280px. A magazine.
--   tool-portal-light — flat 2px hard edges, ink-on-paper, no sidebar, index
--                       grids of bordered cards, 1180px. A tool shelf that reads
--                       like a publication.
--   content-hub-tools — a SINGLE EDITORIAL SPINE (1120px) with a NARROW reading
--                       column (680px, narrower than both), broken by full-width
--                       TOOL SHELVES: tinted bands with a label rail where the
--                       tools sit inside the reading flow rather than on a
--                       separate index. A tool can also be EMBEDDED inline in an
--                       article (.tool-embed) at reading width. Quiet category
--                       strip under the header (editorial signal), 6px radius
--                       (between the siblings), hairline + lift, no heavy
--                       shadow, short accent rule under section titles.
--
-- RENDERER CONTRACT (identical to tool-portal-light / brochure-formal):
--   - MUST NOT declare --section-* defaults; the renderer appends them after
--     rendering based on palette luminance.
--   - Renderer-managed surface classes MUST be surface-coloured here:
--       .features-section, .services-section, .differentiators-section,
--       .about-section, .faq-section
--   - Element rules use var(--section-*, var(--color-*)).
--
-- CLASS VOCABULARY. Most live components carry their own <style> and only need
-- the frame; three do not (faq, featured-content, category-listing) and are
-- styled fully here. Both the sibling-layout contract classes (.tool-card,
-- .tools-grid, .tool-workspace, .article-body, .guide-card) and the classes the
-- live tool-list / guide-list / hero components actually emit (tl-*, guide-*,
-- hero*) are covered, so neither vocabulary renders unframed.
--
-- Data only. Live on COMMIT. No roll. The classifier sees the new tags in its
-- next run via read_layout_taxonomy (now rendered — migration 734).

BEGIN;

-- ── Guard 1: reachability. A layout no site's tags can reach is the tenth
--    unreachable layout, and this migration refuses to be it. ──
DO $$
DECLARE
    v_reach int;
    v_existing_origin text;
BEGIN
    SELECT count(DISTINCT ss.site_id) INTO v_reach
      FROM site_specs ss,
           jsonb_array_elements_text(ss.data->'industry_tags') e(tag)
     WHERE ss.aspect = 'classification' AND ss.is_current
       AND e.tag = ANY (ARRAY[
           'content-hub','interactive-tools','editorial-publication','editorial',
           'long-form','long-form-content','editorial-guides','guides',
           'research-publication','content-platform']);
    IF v_reach < 2 THEN
        RAISE EXCEPTION '736: only % current site(s) emit any of this layout''s tags — refusing to seed an unreachable layout (need >= 2). Re-derive the tag set from live classification specs before re-running.', v_reach;
    END IF;
    RAISE NOTICE '736: reachability OK — % current sites emit at least one of the new layout''s tags', v_reach;

    -- Guard 2: never clobber a forked/adopted row of the same name.
    SELECT origin INTO v_existing_origin FROM layouts WHERE name = 'content-hub-tools';
    IF v_existing_origin IS NOT NULL AND v_existing_origin <> 'seed' THEN
        RAISE EXCEPTION '736: layouts.content-hub-tools already exists with origin=% — not a seed row; refusing to overwrite.', v_existing_origin;
    END IF;
END $$;

INSERT INTO layouts (
    name, display_name, description, category, industry_tags, scheme,
    structure_tokens, css_template, origin, is_active
) VALUES (
    'content-hub-tools',
    'Content Hub — Tools in Context',
    'Editorial content hub whose core offering is a set of embedded interactive tools. A single reading spine with a narrow long-form column, broken by full-width tool shelves where the tools sit inside the reading flow; tools can also be embedded inline in an article at reading width. Quiet category strip, serif display over a humanist sans body, hairline rules with a gentle lift. Suits guide-and-tool sites, practitioner content platforms, research publications with calculators, buyer''s-guide hubs with finders, and any editorial site where the tools are the point rather than a sidebar.',
    'editorial',
    ARRAY['content-hub','interactive-tools','editorial-publication','long-form','long-form-content','editorial-guides','guides','research-publication','content-platform'],
    'light',
    '{
        "container_max_width": "1120px",
        "container_padding_x": "1.5rem",
        "reading_max_width": "680px",
        "tool_embed_max_width": "820px",
        "section_padding_y": "4.5rem",
        "section_padding_y_mobile": "2.75rem",
        "shelf_padding_y": "2.5rem",
        "category_strip_height": "44px",
        "border_radius": "6px",
        "border_radius_sm": "4px",
        "grid_gap": "1.75rem",
        "card_padding": "1.5rem",
        "card_min_width": "260px",
        "shadow_sm": "0 1px 2px rgba(0,0,0,0.04)",
        "shadow_md": "0 8px 22px rgba(0,0,0,0.07)",
        "transition_base": "180ms ease"
    }'::jsonb,
    $LAYOUT$
/* =====================================================================
 * LAYOUT: content-hub-tools
 * An editorial content hub whose core offering is embedded interactive
 * tools. Grammar: one reading spine, narrow long-form column, full-width
 * TOOL SHELVES that interrupt the prose rhythm, tools embeddable inline
 * at reading width. Quiet category strip, serif display, humanist sans
 * body, hairlines + gentle lift, 6px radius.
 *
 * Renderer contract (identical to tool-portal-light / brochure-formal):
 *   - MUST NOT declare --section-* defaults; the renderer appends them
 *     after rendering based on palette luminance.
 *   - Renderer-managed surface classes MUST be surface-coloured here:
 *       .features-section, .services-section, .differentiators-section,
 *       .about-section, .faq-section
 *   - Element rules use var(--section-*, var(--color-*)).
 * ===================================================================== */

:root {
  /* ── Palette — LIGHT fallbacks (palette vars override) ── */
  --color-primary:        {{palette "primary"        "#1f2a37"}};
  --color-primary-hover:  {{palette "primary_hover"  "#111827"}};
  --color-primary-text:   {{palette "primary_text"   "#ffffff"}};
  --color-secondary:      {{palette "secondary"      "#4b5563"}};
  --color-accent:         {{palette "accent"         "#0f766e"}};
  --color-background:     {{palette "background"     "#fcfbf9"}};
  --color-surface:        {{palette "surface"        "#f3f1ec"}};
  --color-surface-alt:    {{palette "surface_alt"    "#eae6de"}};
  --color-text:           {{palette "text"           "#1f2937"}};
  --color-text-muted:     {{palette "text_muted"     "#6b7280"}};
  --color-border:         {{palette "border"         "#d6d1c7"}};
  --color-hairline:       {{palette "hairline"       "#e6e2da"}};
  --color-card-bg:        {{palette "card_bg"        "#ffffff"}};
  --color-header-bg:      {{palette "header_bg"      "#fcfbf9"}};
  --color-header-text:    {{palette "header_text"    "#1f2937"}};
  --color-footer-bg:      {{palette "footer_bg"      "#f3f1ec"}};
  --color-footer-text:    {{palette "footer_text"    "#374151"}};
  --color-cta-bg:         {{palette "cta_bg"         "#0f766e"}};
  --color-cta-text:       {{palette "cta_text"       "#ffffff"}};
  --color-code-bg:        {{palette "code_bg"        "#f3f1ec"}};
  --color-callout-bg:     {{palette "callout_bg"     "rgba(15,118,110,0.07)"}};
  --color-callout-border: {{palette "callout_border" "#0f766e"}};
  --color-badge-bg:       {{palette "badge_bg"       "#1f2a37"}};
  --color-badge-text:     {{palette "badge_text"     "#ffffff"}};

  /* ── Typography — serif display, humanist sans body ── */
  --font-body:        {{typo "font_family"  "'Source Sans 3', 'Segoe UI', system-ui, sans-serif"}};
  --font-heading:     {{typo "heading_font" "'Source Serif 4', Georgia, 'Times New Roman', serif"}};
  --font-mono:        {{typo "mono_font"    "'JetBrains Mono', 'SFMono-Regular', Consolas, Menlo, monospace"}};
  --font-size-base:   {{typo "base_size"    "17px"}};
  --line-height-base: {{typo "line_height"  "1.65"}};

  /* ── Structure — one spine, narrow reading, shelves ── */
  --container-max:    {{token "container_max_width"      "1120px"}};
  --container-pad-x:  {{token "container_padding_x"      "1.5rem"}};
  --reading-max:      {{token "reading_max_width"        "680px"}};
  --tool-embed-max:   {{token "tool_embed_max_width"     "820px"}};
  --section-pad-y:    {{token "section_padding_y"        "4.5rem"}};
  --section-pad-y-sm: {{token "section_padding_y_mobile" "2.75rem"}};
  --shelf-pad-y:      {{token "shelf_padding_y"          "2.5rem"}};
  --strip-h:          {{token "category_strip_height"    "44px"}};
  --radius:           {{token "border_radius"            "6px"}};
  --radius-sm:        {{token "border_radius_sm"         "4px"}};
  --grid-gap:         {{token "grid_gap"                 "1.75rem"}};
  --card-pad:         {{token "card_padding"             "1.5rem"}};
  --card-min:         {{token "card_min_width"           "260px"}};
  --shadow-sm:        {{token "shadow_sm"                "0 1px 2px rgba(0,0,0,0.04)"}};
  --shadow-md:        {{token "shadow_md"                "0 8px 22px rgba(0,0,0,0.07)"}};
  --transition:       {{token "transition_base"          "180ms ease"}};

  {{with palette "heading" ""}}--section-heading: {{.}};{{end}}
}

/* ── Base ── */
*, *::before, *::after { box-sizing: border-box; }
html { -webkit-text-size-adjust: 100%; scroll-behavior: smooth; }
body {
  margin: 0;
  font-family: var(--font-body);
  font-size: var(--font-size-base);
  line-height: var(--line-height-base);
  color: var(--color-text);
  background: var(--color-background);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  -webkit-font-smoothing: antialiased;
}
main { flex: 1; }
img { max-width: 100%; height: auto; display: block; }
code, pre, kbd, samp { font-family: var(--font-mono); }

/* ── Headings + text inheritance (Colour Inheritance Model) ── */
h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-heading);
  color: var(--section-heading, var(--section-text, inherit));
  line-height: 1.2;
  margin: 0 0 0.75rem;
  font-weight: 600;
  letter-spacing: -0.01em;
}
h1 { font-size: clamp(2.1rem, 4.2vw, 3rem); line-height: 1.12; }
h2 { font-size: clamp(1.55rem, 2.8vw, 2rem); }
h3 { font-size: 1.3rem; }
h4 { font-size: 1.05rem; }
p, li, blockquote { color: var(--section-text, inherit); margin: 0 0 1rem; }
blockquote {
  margin: 1.5rem 0; padding: 0.25rem 0 0.25rem 1.25rem;
  border-left: 3px solid var(--color-accent);
  color: var(--section-text-muted, var(--color-text-muted));
  font-family: var(--font-heading); font-size: 1.1rem; font-style: italic;
}
/* strong/em/cite/span: do NOT set color */
a { color: var(--color-accent-ink, var(--color-accent)); text-decoration: none; transition: color var(--transition); }
a:hover { color: var(--color-primary); text-decoration: underline; text-underline-offset: 3px; }

/* ── Primitives ── */
.container { max-width: var(--container-max); margin-inline: auto; padding-inline: var(--container-pad-x); width: 100%; }
.container--narrow, .container--reading { max-width: var(--reading-max); }
.section { padding-block: var(--section-pad-y); }

/* Section title with the SHORT ACCENT RULE — the layout's signature mark */
.section__title, .section-title, .section__header h2, .tl-heading, .guide-list-heading, .features-container > h2, .cta-section h2 {
  position: relative; padding-bottom: 0.75rem; margin-bottom: 1.25rem;
}
.section__title::after, .section-title::after, .section__header h2::after, .tl-heading::after, .guide-list-heading::after {
  content: ""; position: absolute; left: 0; bottom: 0;
  width: 2.75rem; height: 3px; border-radius: 2px;
  background: var(--color-accent);
}
.section__title--center::after { left: 50%; transform: translateX(-50%); }
.section__header { display: flex; align-items: flex-end; justify-content: space-between; gap: 1rem; margin-bottom: 1.5rem; }
.section__header--with-link .section__link { font-size: 0.9rem; font-weight: 600; white-space: nowrap; padding-bottom: 0.9rem; }
.section__link::after { content: " \2192"; }
.section__subtitle, .features-subtitle, .cta-subtitle, .tl-intro, .guide-list-intro {
  color: var(--section-text-muted, var(--color-text-muted)); max-width: 60ch; margin: 0 0 1.5rem;
}
/* eyebrow labels (guide-list / tool-list emit these) */
.tl-eyebrow, .guide-list-eyebrow, .eyebrow, .featured-article__category, .article-card__category {
  display: inline-block; font-family: var(--font-mono); font-size: 0.7rem; font-weight: 600;
  text-transform: uppercase; letter-spacing: 0.09em; color: var(--color-accent); margin-bottom: 0.6rem;
}

/* ── Header — quiet, then a category strip beneath it ── */
.site-header {
  background: var(--color-header-bg); color: var(--color-header-text);
  border-bottom: 1px solid var(--color-hairline);
  position: sticky; top: 0; z-index: 50;
}
.header-container, .header-inner {
  max-width: var(--container-max); margin-inline: auto; padding: 0 var(--container-pad-x);
  display: flex; align-items: center; justify-content: space-between; gap: 2rem; min-height: 68px;
}
.logo, .header-logo { font-family: var(--font-heading); font-size: 1.35rem; font-weight: 700; color: var(--color-header-text); text-decoration: none; letter-spacing: -0.02em; display: inline-flex; align-items: center; gap: 0.6rem; }
.logo:hover, .header-logo:hover { text-decoration: none; color: var(--color-accent); }
.logo-img, .header-logo-img { max-height: 36px; width: auto; }
.main-nav ul, .header-nav ul { display: flex; gap: 1.6rem; list-style: none; margin: 0; padding: 0; }
.main-nav a, .header-nav a, .nav-link { color: var(--color-header-text); font-weight: 500; font-size: 0.92rem; padding: 0.5rem 0; }
.main-nav a:hover, .main-nav a.active, .header-nav a:hover, .nav-link:hover, .nav-link.active { color: var(--color-accent); text-decoration: none; }
.header-cta, .header-actions .btn-primary { margin-left: 0.5rem; }
.mobile-menu-toggle, .hamburger { display: none; background: none; border: none; cursor: pointer; padding: 0.5rem; min-width: 44px; min-height: 44px; color: var(--color-header-text); }
.mobile-menu-toggle span, .hamburger span { display: block; width: 22px; height: 2px; background: currentColor; margin: 5px 0; }

/* Category strip — the editorial signal, kept quiet: a rail, not a bar */
.category-strip, .hwc-categories-bar {
  background: var(--color-surface); border-bottom: 1px solid var(--color-hairline);
}
.category-strip ul, .hwc-categories-inner {
  max-width: var(--container-max); margin-inline: auto; padding: 0 var(--container-pad-x);
  display: flex; gap: 1.5rem; list-style: none; min-height: var(--strip-h); align-items: center; overflow-x: auto;
}
.category-strip a, .hwc-category-link {
  font-size: 0.82rem; font-weight: 600; letter-spacing: 0.02em; white-space: nowrap;
  color: var(--color-text-muted); padding: 0.35rem 0; border-bottom: 2px solid transparent;
}
.category-strip a:hover, .category-strip a.active, .hwc-category-link:hover { color: var(--color-accent); border-bottom-color: var(--color-accent); text-decoration: none; }
.hwc-categories-label { font-family: var(--font-mono); font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--color-text-muted); }

/* ── Hero — an editorial lede, not a billboard ── */
.hero-section, .hero { padding-block: calc(var(--section-pad-y) * 1.1); border-bottom: 1px solid var(--color-hairline); }
.hero-section .container, .hero-content { max-width: 860px; }
.hero-section h1, .hero h1 { font-size: clamp(2.3rem, 5vw, 3.4rem); margin-bottom: 1.1rem; letter-spacing: -0.02em; }
.hero-subtitle, .hero-subheadline, .hero-section .lead {
  font-size: 1.2rem; line-height: 1.55; max-width: 620px; margin: 0 0 1.75rem;
  color: var(--section-text-muted, var(--color-text-muted));
}
.hero-actions { display: flex; gap: 0.75rem; flex-wrap: wrap; }

/* Featured article (featured-content component — no inline style, styled here) */
.section--featured { padding-block: var(--section-pad-y); }
.featured-article { display: grid; grid-template-columns: 1.25fr 1fr; gap: 2.5rem; align-items: center; }
.featured-article__image { border-radius: var(--radius); overflow: hidden; aspect-ratio: 16 / 10; background: var(--color-surface); }
.featured-article__image img { width: 100%; height: 100%; object-fit: cover; transition: transform var(--transition); }
.featured-article:hover .featured-article__image img { transform: scale(1.02); }
.featured-article__title { font-size: clamp(1.6rem, 3vw, 2.2rem); margin-bottom: 0.75rem; }
.featured-article__title a { color: inherit; }
.featured-article__excerpt { color: var(--section-text-muted, var(--color-text-muted)); font-size: 1.05rem; }
.featured-article__meta { display: flex; flex-wrap: wrap; gap: 0.75rem; font-size: 0.82rem; color: var(--color-text-muted); margin-top: 0.75rem; }
.featured-article__meta > * + *::before { content: "\00b7"; margin-right: 0.75rem; }

/* ── THE TOOL SHELF — a full-width band that interrupts the reading spine ── */
.tool-list-section, .tools-section, .tool-shelf {
  background: var(--color-surface);
  border-top: 1px solid var(--color-hairline); border-bottom: 1px solid var(--color-hairline);
  padding-block: var(--shelf-pad-y);
}
.tl-inner, .tool-shelf__inner { max-width: var(--container-max); margin-inline: auto; padding-inline: var(--container-pad-x); }
.tl-header, .tool-shelf__header { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 1rem 2rem; align-items: end; margin-bottom: 1.5rem; }
/* the label rail: a mono label that runs down the left of the shelf on wide screens */
.tool-shelf__rail {
  font-family: var(--font-mono); font-size: 0.7rem; text-transform: uppercase; letter-spacing: 0.1em;
  color: var(--color-text-muted); writing-mode: vertical-rl; transform: rotate(180deg);
  border-right: 1px solid var(--color-border); padding-right: 0.5rem; align-self: stretch;
}
.tl-grid, .tools-grid, .tool-shelf__grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(var(--card-min), 1fr)); gap: var(--grid-gap);
}
.tl-card, .tool-card, .tool-shelf__card {
  background: var(--color-card-bg); border: 1px solid var(--color-hairline); border-radius: var(--radius);
  padding: var(--card-pad); display: flex; flex-direction: column; gap: 0.5rem;
  box-shadow: var(--shadow-sm); transition: transform var(--transition), box-shadow var(--transition), border-color var(--transition);
}
.tl-card:hover, .tool-card:hover, .tool-shelf__card:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); border-color: var(--color-border); }
.tl-card-icon, .tool-card__icon { width: 30px; height: 30px; color: var(--color-accent); margin-bottom: 0.25rem; }
.tl-card-media { border-radius: var(--radius-sm); overflow: hidden; margin-bottom: 0.5rem; }
.tl-card-title, .tool-card__title { font-family: var(--font-heading); font-size: 1.15rem; margin: 0; }
.tl-card-desc, .tool-card__description { font-size: 0.95rem; color: var(--color-text-muted); margin: 0; flex: 1; }
.tl-card-link, .tool-card__link { font-weight: 600; font-size: 0.9rem; margin-top: 0.25rem; }
.tl-card-link::after, .tool-card__link::after { content: " \2192"; }
.tl-cta, .tool-shelf__cta { margin-top: 1.5rem; display: flex; align-items: center; gap: 1rem; flex-wrap: wrap; }
.tl-cta-text { color: var(--color-text-muted); margin: 0; }
.tool-card__label, .tool-label { display: inline-block; font-family: var(--font-mono); font-size: 0.68rem; text-transform: uppercase; letter-spacing: 0.06em; padding: 0.15rem 0.45rem; border: 1px solid var(--color-hairline); border-radius: var(--radius-sm); color: var(--color-text-muted); }

/* ── Guide cards / article cards — the editorial grid ── */
.guide-list-section, .guides-section, .section--category { padding-block: var(--section-pad-y); }
.guide-list-inner { max-width: var(--container-max); margin-inline: auto; padding-inline: var(--container-pad-x); }
.guide-list-grid, .guides-grid, .article-grid, .grid--4 {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(var(--card-min), 1fr)); gap: var(--grid-gap);
}
.guide-card, .article-card {
  background: var(--color-card-bg); border: 1px solid var(--color-hairline); border-radius: var(--radius);
  overflow: hidden; display: flex; flex-direction: column; transition: transform var(--transition), box-shadow var(--transition);
}
.guide-card:hover, .article-card:hover, .hover-lift:hover { transform: translateY(-2px); box-shadow: var(--shadow-md); }
.article-card__image { aspect-ratio: 16 / 9; background: var(--color-surface); overflow: hidden; }
.article-card__image img { width: 100%; height: 100%; object-fit: cover; }
.article-card__content, .guide-card > :not(.article-card__image) { padding: var(--card-pad); }
.guide-card { padding: var(--card-pad); gap: 0.4rem; }
.guide-card-badge { align-self: flex-start; font-family: var(--font-mono); font-size: 0.68rem; text-transform: uppercase; letter-spacing: 0.06em; padding: 0.15rem 0.5rem; border-radius: var(--radius-sm); background: var(--color-badge-bg); color: var(--color-badge-text); }
.guide-card-title, .article-card__title { font-family: var(--font-heading); font-size: 1.15rem; margin: 0.25rem 0 0.35rem; }
.guide-card-title a, .article-card__title a { color: inherit; }
.guide-card-desc, .article-card__excerpt { font-size: 0.95rem; color: var(--color-text-muted); margin: 0; flex: 1; }
.guide-card-link-label, .article-card__date { font-size: 0.82rem; font-weight: 600; margin-top: 0.6rem; color: var(--color-text-muted); }
.article-card--compact { flex-direction: row; gap: 1rem; align-items: center; }
.article-card--compact .article-card__image { width: 96px; flex: 0 0 96px; aspect-ratio: 1; border-radius: var(--radius-sm); }
.article-card--compact .article-card__content { padding: 0.5rem 0.75rem 0.5rem 0; }
.article-card--compact .article-card__excerpt { display: none; }
.guide-list-cta-wrap { margin-top: 2.5rem; padding: 1.5rem; border: 1px dashed var(--color-border); border-radius: var(--radius); display: flex; gap: 1rem; align-items: center; justify-content: space-between; flex-wrap: wrap; }
.guide-list-cta-heading { font-family: var(--font-heading); margin: 0; }
.guide-list-cta-sub { color: var(--color-text-muted); margin: 0.25rem 0 0; }

/* ── Article / guide page — the narrow reading column ── */
.article-section, .article-body-section { padding-block: var(--section-pad-y); }
.article-section .container, .article-body-section .container { max-width: var(--reading-max); }
.article-body, .article-body__content { font-size: 1.1rem; line-height: 1.75; color: var(--color-text); }
.article-body h2, .article-body__content h2 { font-size: 1.55rem; margin-top: 2.6rem; }
.article-body h3, .article-body__content h3 { font-size: 1.2rem; margin-top: 2rem; }
.article-body p, .article-body__content p { margin-bottom: 1.3rem; }
.article-body ul, .article-body ol, .article-body__content ul, .article-body__content ol { margin: 0 0 1.3rem 1.4rem; }
.article-body li, .article-body__content li { margin-bottom: 0.45rem; }
.article-body img, .article-body__content img { border-radius: var(--radius); margin: 1.75rem 0; }
.article-body code, .article-body__content code { background: var(--color-code-bg); border: 1px solid var(--color-hairline); border-radius: var(--radius-sm); padding: 0.1em 0.35em; font-size: 0.88em; }
.article-body pre, .article-body__content pre { background: var(--color-code-bg); border: 1px solid var(--color-hairline); border-radius: var(--radius); padding: 1rem 1.25rem; overflow-x: auto; margin: 1.5rem 0; font-size: 0.9rem; line-height: 1.5; }
.article-body pre code { background: transparent; border: none; padding: 0; }
.article-body table, .article-body__content table { width: 100%; border-collapse: collapse; margin: 1.5rem 0; font-size: 0.95rem; }
.article-body th, .article-body td, .article-body__content th, .article-body__content td { padding: 0.6rem 0.875rem; text-align: left; border-bottom: 1px solid var(--color-hairline); }
.article-body th, .article-body__content th { font-family: var(--font-mono); font-weight: 600; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.06em; color: var(--color-text-muted); }

/* Callout */
.callout, .design-rule { background: var(--color-callout-bg); border-left: 3px solid var(--color-callout-border); padding: 1rem 1.25rem; margin: 1.5rem 0; border-radius: 0 var(--radius) var(--radius) 0; }
.callout__label, .design-rule__label { display: block; font-family: var(--font-mono); font-size: 0.7rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--color-callout-border); margin-bottom: 0.35rem; }

/* ── THE INLINE TOOL EMBED — a tool inside the article, at (slightly more than) reading width ── */
.tool-embed, .inline-tool {
  max-width: var(--tool-embed-max); margin: 2.5rem auto;
  background: var(--color-card-bg); border: 1px solid var(--color-border); border-radius: var(--radius);
  box-shadow: var(--shadow-sm); overflow: hidden;
}
.tool-embed__header, .inline-tool__header {
  display: flex; align-items: baseline; justify-content: space-between; gap: 1rem;
  padding: 0.85rem 1.25rem; background: var(--color-surface); border-bottom: 1px solid var(--color-hairline);
}
.tool-embed__title, .inline-tool__title { font-family: var(--font-heading); font-size: 1.1rem; margin: 0; }
.tool-embed__label { font-family: var(--font-mono); font-size: 0.68rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--color-accent); }
.tool-embed__body, .inline-tool__body { padding: 1.25rem; }
.tool-embed__footer { padding: 0.75rem 1.25rem; border-top: 1px solid var(--color-hairline); font-size: 0.85rem; color: var(--color-text-muted); }

/* ── Tool page — the workspace, framed by an editorial intro ── */
.tool-section { padding-block: var(--section-pad-y); background: var(--color-background); }
.tool-section .container { max-width: 960px; }
.tool-intro { max-width: var(--reading-max); margin-bottom: 2rem; }
.tool-intro p { font-size: 1.1rem; color: var(--section-text-muted, var(--color-text-muted)); }
.tool-workspace { background: var(--color-card-bg); border: 1px solid var(--color-border); border-radius: var(--radius); padding: clamp(1.25rem, 3vw, 2rem); margin-bottom: 2.5rem; box-shadow: var(--shadow-sm); }
.tool-workspace__header { display: flex; justify-content: space-between; align-items: baseline; gap: 1rem; margin-bottom: 1.25rem; padding-bottom: 0.75rem; border-bottom: 1px solid var(--color-hairline); }
.tool-workspace__title { font-family: var(--font-heading); font-size: 1.3rem; margin: 0; }
.tool-workspace__meta { font-family: var(--font-mono); font-size: 0.78rem; color: var(--color-text-muted); }
.tool-inputs { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 1.5rem; }
.tool-output { background: var(--color-code-bg); border: 1px solid var(--color-hairline); border-radius: var(--radius); padding: 1.25rem; font-family: var(--font-mono); font-size: 0.95rem; line-height: 1.5; color: var(--color-text); overflow-x: auto; }
.tool-output__label { display: block; font-family: var(--font-body); font-size: 0.72rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: var(--color-text-muted); margin-bottom: 0.5rem; }
.tool-output__value { font-family: var(--font-heading); font-size: 1.7rem; font-weight: 600; color: var(--color-accent); }
/* "read the guide" cross-link under a tool — hub sites pair every tool with prose */
.tool-guide-link { display: inline-flex; align-items: center; gap: 0.5rem; margin-top: 1rem; font-weight: 600; }
.tool-guide-link::before { content: "\2192"; color: var(--color-accent); }

/* ── Renderer-managed surface sections (MUST be surface-coloured) ── */
.features-section, .services-section, .differentiators-section,
.about-section, .faq-section { background: var(--color-surface); }
.features-grid, .services-grid, .differentiators-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: var(--grid-gap); }
.feature-item, .feature-card, .service-card, .differentiator-card { background: var(--color-card-bg); border: 1px solid var(--color-hairline); border-radius: var(--radius); padding: var(--card-pad); }
.feature-icon, .service-icon { width: 28px; height: 28px; margin-bottom: 0.5rem; color: var(--color-accent); }

/* ── Forms (tool controls first-class) ── */
.form-field, .form-group { margin-bottom: 1rem; }
.form-field label, .form-group label { display: block; font-weight: 600; margin-bottom: 0.3rem; font-size: 0.8rem; color: var(--color-text-muted); text-transform: uppercase; letter-spacing: 0.04em; }
.form-field input, .form-field textarea, .form-field select,
.form-group input, .form-group textarea, .form-group select {
  width: 100%; padding: 0.65rem 0.8rem; font: inherit; font-size: 0.95rem; border: 1px solid var(--color-border);
  border-radius: var(--radius-sm); background: var(--color-card-bg); color: var(--color-text); min-height: 44px;
  transition: border-color var(--transition), box-shadow var(--transition);
}
.form-field input:focus, .form-field textarea:focus, .form-field select:focus,
.form-group input:focus, .form-group textarea:focus, .form-group select:focus {
  outline: none; border-color: var(--color-accent); box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent) 20%, transparent);
}

/* ── Buttons — soft rectangles: neither the pill nor the hard edge ── */
.btn, .button, .cta-btn, .tl-cta-btn, .guide-list-cta-btn, .hwc-cta-btn {
  display: inline-flex; align-items: center; justify-content: center; gap: 0.5rem; padding: 0.7rem 1.35rem;
  font: inherit; font-weight: 600; font-size: 0.92rem; border-radius: var(--radius); border: 1px solid transparent;
  cursor: pointer; text-decoration: none; min-height: 44px;
  transition: background var(--transition), border-color var(--transition), color var(--transition), transform var(--transition);
}
.btn:hover, .button:hover, .cta-btn:hover { text-decoration: none; transform: translateY(-1px); }
.btn-primary, .button--primary, .cta-btn-primary, .tl-cta-btn, .guide-list-cta-btn, .hwc-cta-btn { background: var(--color-cta-bg); color: var(--color-cta-bg-ink, var(--color-cta-text)); border-color: var(--color-cta-bg); }
.btn-primary:hover, .button--primary:hover, .cta-btn-primary:hover { background: var(--color-primary); border-color: var(--color-primary); color: var(--color-primary-text); }
.btn-secondary, .cta-btn-secondary { background: transparent; color: var(--color-text); border-color: var(--color-border); }
.btn-secondary:hover, .cta-btn-secondary:hover { background: var(--color-surface-alt); }
.btn-ghost { background: transparent; color: var(--color-text-muted); border-color: transparent; }
.btn-ghost:hover { color: var(--color-text); background: var(--color-surface); }
.btn-large { padding: 0.9rem 1.8rem; font-size: 0.98rem; min-height: 50px; }

/* ── CTA band ── */
.cta-section { background: var(--color-surface); border-top: 1px solid var(--color-hairline); border-bottom: 1px solid var(--color-hairline); padding-block: var(--section-pad-y); text-align: center; }
.cta-container { max-width: 720px; margin-inline: auto; padding-inline: var(--container-pad-x); }
.cta-buttons { display: flex; gap: 0.75rem; justify-content: center; flex-wrap: wrap; margin-top: 1.5rem; }

/* ── FAQ (component has no inline style — fully framed here) ── */
.section--faq .container, .faq-section .container { max-width: var(--reading-max); }
.faq-list { margin: 0; padding: 0; list-style: none; }
.faq-item { border-bottom: 1px solid var(--color-hairline); padding: 1rem 0; }
.faq-item summary, .faq-item__question { cursor: pointer; font-family: var(--font-heading); font-weight: 600; font-size: 1.05rem; list-style: none; display: flex; justify-content: space-between; gap: 1rem; min-height: 44px; align-items: center; }
.faq-item summary::-webkit-details-marker { display: none; }
.faq-item summary::after { content: "+"; color: var(--color-text-muted); font-size: 1.25rem; }
.faq-item[open] summary::after { content: "\2212"; color: var(--color-accent); }
.faq-item p, .faq-item__answer { padding-top: 0.6rem; color: var(--section-text-muted, var(--color-text-muted)); }

/* ── About / Contact ── */
.about-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; align-items: start; }
.contact-section .container, .contact-form-section .container { display: grid; grid-template-columns: 1fr 1fr; gap: 3rem; }

/* ── Footer — light, hairline, editorial colophon ── */
.site-footer, .site-footer-section { background: var(--color-footer-bg); color: var(--color-footer-text); padding-top: 3rem; margin-top: auto; border-top: 1px solid var(--color-border); font-size: 0.9rem; }
.footer-container, .footer-inner, .footer-top { max-width: var(--container-max); margin-inline: auto; padding: 0 var(--container-pad-x); display: grid; grid-template-columns: 2fr 1fr 1fr 1fr; gap: 2rem; }
.site-footer h3, .site-footer h4, .footer-col h4 { font-family: var(--font-heading); color: var(--color-footer-text); font-size: 1rem; }
.footer-brand, .footer-logo { font-family: var(--font-heading); font-size: 1.2rem; font-weight: 700; }
.footer-tagline { color: var(--color-text-muted); max-width: 40ch; }
.site-footer a { color: var(--color-footer-text); }
.site-footer a:hover { color: var(--color-accent); }
.site-footer ul { list-style: none; padding: 0; margin: 0; }
.site-footer li { margin-bottom: 0.4rem; }
.footer-bottom { margin-top: 2.5rem; padding: 1.25rem 0; border-top: 1px solid var(--color-hairline); text-align: center; font-size: 0.8rem; color: var(--color-text-muted); }
.footer-newsletter .newsletter-form { display: flex; gap: 0.5rem; }

/* ── Responsive ── */
@media (max-width: 1024px) {
  .footer-container, .footer-inner, .footer-top { grid-template-columns: repeat(2, 1fr); }
  .featured-article { grid-template-columns: 1fr; }
}
@media (max-width: 768px) {
  .section, .tool-list-section, .tools-section, .tool-shelf, .guide-list-section, .cta-section { padding-block: var(--section-pad-y-sm); }
  .about-section .container, .contact-section .container, .contact-form-section .container,
  .footer-container, .footer-inner, .footer-top { grid-template-columns: 1fr; }
  .main-nav, .header-nav { display: none; }
  .main-nav.is-open, .header-nav.is-open { display: block; }
  .mobile-menu-toggle, .hamburger { display: inline-flex; flex-direction: column; justify-content: center; align-items: center; }
  .tl-header, .tool-shelf__header, .tool-workspace__header { grid-template-columns: 1fr; flex-direction: column; align-items: flex-start; }
  .tool-shelf__rail { writing-mode: horizontal-tb; transform: none; border-right: 0; border-bottom: 1px solid var(--color-border); padding: 0 0 0.35rem; }
  .article-card--compact { flex-direction: column; align-items: stretch; }
  .article-card--compact .article-card__image { width: 100%; aspect-ratio: 16 / 9; }
}

/* ── Accessibility ── */
:focus-visible { outline: 2px solid var(--color-accent); outline-offset: 2px; }
@media (prefers-reduced-motion: reduce) { *, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; transform: none !important; } }
$LAYOUT$,
    'seed',
    true
)
ON CONFLICT (name) DO UPDATE SET
    display_name     = EXCLUDED.display_name,
    description      = EXCLUDED.description,
    category         = EXCLUDED.category,
    industry_tags    = EXCLUDED.industry_tags,
    scheme           = EXCLUDED.scheme,
    structure_tokens = EXCLUDED.structure_tokens,
    css_template     = EXCLUDED.css_template,
    is_active        = EXCLUDED.is_active,
    updated_at       = NOW();

-- ── Verify: DO/RAISE, not SELECTs (ON_ERROR_STOP ignores a non-empty result). ──
DO $$
DECLARE
    v_active boolean; v_scheme text; v_ntags int; v_csslen int; v_cat text; v_reach int;
BEGIN
    SELECT is_active, scheme, cardinality(industry_tags), length(css_template), category
      INTO v_active, v_scheme, v_ntags, v_csslen, v_cat
      FROM layouts WHERE name = 'content-hub-tools';
    IF v_active IS NULL THEN RAISE EXCEPTION '736 VERIFY: row not present after insert.'; END IF;
    IF NOT v_active OR v_scheme <> 'light' OR v_cat <> 'editorial' THEN
        RAISE EXCEPTION '736 VERIFY: active=% scheme=% category=% — expected true/light/editorial', v_active, v_scheme, v_cat;
    END IF;
    IF v_ntags <> 9 THEN RAISE EXCEPTION '736 VERIFY: % tags, expected 9', v_ntags; END IF;
    IF v_csslen < 9000 THEN RAISE EXCEPTION '736 VERIFY: css_template is % chars — truncated?', v_csslen; END IF;
    IF position('--section-text:' in (SELECT css_template FROM layouts WHERE name='content-hub-tools')) > 0 THEN
        RAISE EXCEPTION '736 VERIFY: template declares a --section-* default — renderer contract violated.';
    END IF;
    -- the layout must be reachable by the vocabulary the classifier ACTUALLY emits today
    SELECT count(DISTINCT ss.site_id) INTO v_reach
      FROM site_specs ss, jsonb_array_elements_text(ss.data->'industry_tags') e(tag), layouts l
     WHERE ss.aspect='classification' AND ss.is_current AND l.name='content-hub-tools'
       AND (e.tag = ANY(l.industry_tags) OR e.tag = 'editorial');
    IF v_reach < 2 THEN RAISE EXCEPTION '736 VERIFY: seeded layout reachable by only % site(s).', v_reach; END IF;
    RAISE NOTICE '736 VERIFY: OK — content-hub-tools active, light, editorial, 9 tags, % chars CSS, reachable by % current sites', v_csslen, v_reach;
END $$;

COMMIT;
