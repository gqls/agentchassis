pre validation
-- =====================================================================
-- Deliverable 2: Spec schemas for site-design-planner
-- =====================================================================
--
-- Purpose: document the JSONB shape of three new site_specs aspects
-- that site-design-planner will write.
--
--   navigation            — nav architecture, items, CTA, mobile pattern
--   layout                — page-level layout choices, header/footer style
--   resolved_composition  — machine-readable pointers to palette/layout/
--                            typography rows + reasoning
--
-- These are separate aspects because they serve separate readers:
--   - populate_nav_tables, InjectHeader, GetNavItems     → navigation
--   - AssembleMultipageSiteAction, header/footer templates → layout
--   - render_css_from_spec, webdesign-agent, audit agents → resolved_composition
--
-- This file:
--   1. Documents each shape with example JSON
--   2. Creates a validation function per aspect (called by write_site_spec
--      or by site-design-planner directly before writing)
--   3. Does NOT create a schema-enforcing table constraint — site_specs
--      stays as open JSONB; validation is best-effort at write time.
-- =====================================================================


-- ---------------------------------------------------------------------
-- ASPECT: navigation
-- ---------------------------------------------------------------------
-- Written by:  site-design-planner
-- Read by:     populate_nav_tables, InjectHeader, GetNavItems,
--              render_site_components (for nav container class choice)
-- Evolves by:  strategist updates (major nav restructure)
--              nav-agent updates (adding individual items)
-- ---------------------------------------------------------------------
--
-- Example:
-- {
--   "architecture":        "horizontal-top",
--   "primary_items":       ["Home", "Services", "Case Studies", "About", "Contact"],
--   "tools_strategy":      "grouped_under_tools_page",
--   "cta": {
--     "label":  "Book a consultation",
--     "url":    "/contact.html",
--     "style":  "solid"
--   },
--   "mobile":              "hamburger-slide",
--   "max_visible_items":   6,
--   "sticky":              true,
--   "logo_position":       "left",
--   "legal_in_footer_only": true,
--   "reasoning":           "Chosen because classification=corporate and identity.tone=professional"
-- }
--
-- Field reference:
--
-- architecture          (required)  enum:
--   - "horizontal-top"        — standard top-bar nav
--   - "horizontal-sticky"     — top-bar that sticks on scroll
--   - "vertical-sidebar"      — fixed left or right sidebar nav
--   - "split-nav"             — brand one side, items other side
--   - "transparent-overlay"   — hero-overlaid nav (hero_nav_merged=true in layout)
--   - "megamenu"              — multi-column dropdown
--   - "bottom-tab"            — mobile-first bottom-tab nav
--
-- primary_items         (required)  array of strings
--   Top-level nav labels. Ordering matters. Rendered verbatim.
--   populate_nav_tables translates these into site_nav_items with URL resolution.
--
-- tools_strategy        (optional)  enum:
--   - "individual_in_nav"         — each tool gets its own nav entry
--   - "grouped_under_tools_page"  — single "Tools" entry, landing page lists them
--   - "footer_only"               — tools not in main nav, footer links only
--   - "none"                      — site has no tools
--
-- cta                   (optional)  object
--   label, url, style ("solid" | "ghost" | "link")
--   Null means no CTA in header. Empty object is invalid.
--
-- mobile                (required)  enum:
--   - "hamburger-slide"       — icon toggles slide-out drawer
--   - "hamburger-fullscreen"  — icon toggles fullscreen overlay
--   - "bottom-tab"            — bottom-tab stays on mobile
--   - "simplified-inline"     — fewer items, still horizontal
--
-- max_visible_items     (optional)  int, default 6
--   Items above this collapse into "More" dropdown (desktop only).
--
-- sticky                (optional)  bool, default false
--
-- logo_position         (optional)  enum: "left" | "center" | "right", default "left"
--
-- legal_in_footer_only  (optional)  bool, default true
--   If true, Privacy/Terms links go in footer not header.
--
-- reasoning             (optional)  string
--   Human-readable explanation of why these choices were made.


-- ---------------------------------------------------------------------
-- ASPECT: layout
-- ---------------------------------------------------------------------
-- Written by:  site-design-planner
-- Read by:     AssembleMultipageSiteAction, render_site_components,
--              InjectHeader/InjectFooter, header template, footer template
-- Evolves by:  strategist (major restructure), design-audit (recommendations)
-- ---------------------------------------------------------------------
--
-- Example:
-- {
--   "default_page_layout":  "full-width-stacked",
--   "header_style":         "dark-professional",
--   "footer_style":         "4-column",
--   "hero_nav_merged":      false,
--   "sidebar_pages":        [],
--   "page_overrides": {
--     "docs":  {"layout": "sidebar-left",  "nav": "sidebar-toc"},
--     "blog":  {"layout": "full-width-stacked", "nav": "none"}
--   },
--   "section_density":      "comfortable",
--   "reasoning":            "Corporate classification suggests formal structure"
-- }
--
-- Field reference:
--
-- default_page_layout   (required)  enum:
--   - "full-width-stacked"   — sections stack vertically, edge-to-edge
--   - "contained-stacked"    — sections stack, max-width container
--   - "sidebar-left"         — left nav/TOC + main content
--   - "sidebar-right"        — main content + right aside
--   - "two-column"           — 50/50 main and aside
--   - "asymmetric"           — layout-specific grid (defined by layouts row)
--
-- header_style          (required)  string
--   References a style key; resolved by header template against palette.
--   Common values: "dark-professional", "light-minimal", "bold-gradient",
--                  "transparent", "warm-friendly", "docs"
--
-- footer_style          (required)  string
--   Like header_style. Common values:
--   "minimal", "4-column", "compact-1-line", "with-disclaimer", "dark-rich"
--
-- hero_nav_merged       (optional)  bool, default false
--   True when the hero section visually incorporates the nav
--   (e.g. transparent nav overlaid on hero background image).
--
-- sidebar_pages         (optional)  array of page names
--   Pages that use sidebar layout even when default is stacked.
--   Page-specific override takes precedence if present in page_overrides.
--
-- page_overrides        (optional)  object, keyed by page name
--   Each value: {layout, nav, ...} — same fields as top-level where applicable.
--
-- section_density       (optional)  enum:
--   "compact" | "comfortable" | "spacious"
--   Hint for component rendering; default "comfortable".
--
-- reasoning             (optional)  string


-- ---------------------------------------------------------------------
-- ASPECT: resolved_composition
-- ---------------------------------------------------------------------
-- Written by:  site-design-planner (primary)
--              fork_theme_from_site (when forking to library)
-- Read by:     render_css_from_spec (via site_context.css_theme_id),
--              audit agents (to know what composition is expected),
--              admin dashboard (display current composition)
-- Evolves by:  re-composition is deferred; updates go through HITL initially
-- ---------------------------------------------------------------------
--
-- Example:
-- {
--   "css_theme_id":      "8a2e4b90-...-uuid",
--   "css_theme_name":    "adopted-gamedesign-uk",
--   "palette_id":        "4f1c8e33-...-uuid",
--   "palette_name":      "gamedesign-uk-dark",
--   "layout_id":         "7b9d2a61-...-uuid",
--   "layout_name":       "utility-tool",
--   "typography_set_id": "2e6f8b44-...-uuid",
--   "typography_name":   "sans-modern",
--   "lineage": {
--     "palette_source":     "fingerprint",
--     "layout_source":      "library_match",
--     "typography_source":  "fingerprint_font_family_match",
--     "layout_match_score": 0.82,
--     "layout_candidates":  ["utility-tool", "docs-sidebar", "brochure-formal"]
--   },
--   "reasoning": "Classification 'developer-tools' scored utility-tool=0.82 above threshold 0.5",
--   "resolved_by":     "site-design-planner",
--   "resolved_at":     "2026-04-18T19:00:00Z"
-- }
--
-- Field reference:
--
-- css_theme_id / css_theme_name           (required)  the composition row
-- palette_id / palette_name               (required)
-- layout_id / layout_name                 (required)
-- typography_set_id / typography_name     (required)
--
-- lineage                                 (required)  object
--   palette_source     enum: "fingerprint" | "library_reuse" | "mission_hint" |
--                            "design_intent_values" | "archetype_default"
--   layout_source      enum: "library_match" | "library_fallback" | "mission_hint" |
--                            "needs_new_layout_candidate"
--   typography_source  enum: "fingerprint_font_family_match" |
--                            "archetype_default" | "layout_default" |
--                            "mission_hint" | "fallback_sans_modern"
--   layout_match_score   (float 0-1) — tag-overlap score for chosen layout
--   layout_candidates    (array) — top-5 layouts considered, sorted by score
--
-- reasoning         (required)  string — why this composition was chosen
-- resolved_by       (required)  agent type name
-- resolved_at       (required)  ISO 8601 timestamp


-- =====================================================================
-- VALIDATION FUNCTIONS
-- =====================================================================
-- Called by site-design-planner before writing each spec aspect.
-- Returns NULL on success, a text error message on failure.
-- site-design-planner should raise (or log.Error + fail the step) on any
-- non-null return.
-- =====================================================================

-- Validate `navigation` aspect
CREATE OR REPLACE FUNCTION validate_navigation_spec(spec jsonb)
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
arch text;
    mobile text;
    items jsonb;
BEGIN
    IF spec IS NULL OR jsonb_typeof(spec) != 'object' THEN
        RETURN 'navigation spec must be a JSON object';
END IF;

    -- architecture required, enum-checked
    arch := spec ->> 'architecture';
    IF arch IS NULL THEN
        RETURN 'navigation.architecture is required';
END IF;
    IF arch NOT IN ('horizontal-top','horizontal-sticky','vertical-sidebar',
                    'split-nav','transparent-overlay','megamenu','bottom-tab') THEN
        RETURN format('navigation.architecture %L is not a known value', arch);
END IF;

    -- primary_items required, array of strings, non-empty
    items := spec -> 'primary_items';
    IF items IS NULL OR jsonb_typeof(items) != 'array' THEN
        RETURN 'navigation.primary_items must be an array';
END IF;
    IF jsonb_array_length(items) = 0 THEN
        RETURN 'navigation.primary_items must be non-empty';
END IF;

    -- mobile required, enum-checked
    mobile := spec ->> 'mobile';
    IF mobile IS NULL THEN
        RETURN 'navigation.mobile is required';
END IF;
    IF mobile NOT IN ('hamburger-slide','hamburger-fullscreen',
                      'bottom-tab','simplified-inline') THEN
        RETURN format('navigation.mobile %L is not a known value', mobile);
END IF;

    -- logo_position if present must be valid
    IF spec ? 'logo_position' AND (spec ->> 'logo_position') NOT IN ('left','center','right') THEN
        RETURN format('navigation.logo_position %L is not a known value',
                      spec ->> 'logo_position');
END IF;

RETURN NULL;
END;
$$;


-- Validate `layout` aspect
CREATE OR REPLACE FUNCTION validate_layout_spec(spec jsonb)
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
page_layout text;
BEGIN
    IF spec IS NULL OR jsonb_typeof(spec) != 'object' THEN
        RETURN 'layout spec must be a JSON object';
END IF;

    page_layout := spec ->> 'default_page_layout';
    IF page_layout IS NULL THEN
        RETURN 'layout.default_page_layout is required';
END IF;
    IF page_layout NOT IN ('full-width-stacked','contained-stacked',
                           'sidebar-left','sidebar-right',
                           'two-column','asymmetric') THEN
        RETURN format('layout.default_page_layout %L is not a known value', page_layout);
END IF;

    IF (spec ->> 'header_style') IS NULL THEN
        RETURN 'layout.header_style is required';
END IF;

    IF (spec ->> 'footer_style') IS NULL THEN
        RETURN 'layout.footer_style is required';
END IF;

    -- section_density if present must be valid
    IF spec ? 'section_density' AND (spec ->> 'section_density') NOT IN
       ('compact','comfortable','spacious') THEN
        RETURN format('layout.section_density %L is not a known value',
                      spec ->> 'section_density');
END IF;

RETURN NULL;
END;
$$;


-- Validate `resolved_composition` aspect
CREATE OR REPLACE FUNCTION validate_resolved_composition_spec(spec jsonb)
RETURNS text
LANGUAGE plpgsql
AS $$
DECLARE
required_id_fields text[] := ARRAY[
        'css_theme_id','palette_id','layout_id','typography_set_id'
    ];
    required_name_fields text[] := ARRAY[
        'css_theme_name','palette_name','layout_name','typography_name'
    ];
    f text;
    lineage jsonb;
    palette_src text;
    layout_src text;
    typography_src text;
BEGIN
    IF spec IS NULL OR jsonb_typeof(spec) != 'object' THEN
        RETURN 'resolved_composition spec must be a JSON object';
END IF;

    FOREACH f IN ARRAY required_id_fields LOOP
        IF (spec ->> f) IS NULL THEN
            RETURN format('resolved_composition.%s is required', f);
END IF;
        -- Very loose UUID check (36 chars with hyphens); full validation is
        -- Go's job via uuid.Parse. This just catches obviously wrong values.
        IF length(spec ->> f) != 36 THEN
            RETURN format('resolved_composition.%s does not look like a UUID (length=%s)',
                          f, length(spec ->> f));
END IF;
END LOOP;

    FOREACH f IN ARRAY required_name_fields LOOP
        IF (spec ->> f) IS NULL OR (spec ->> f) = '' THEN
            RETURN format('resolved_composition.%s is required and non-empty', f);
END IF;
END LOOP;

    -- lineage required with its three source fields
    lineage := spec -> 'lineage';
    IF lineage IS NULL OR jsonb_typeof(lineage) != 'object' THEN
        RETURN 'resolved_composition.lineage must be a JSON object';
END IF;

    palette_src := lineage ->> 'palette_source';
    IF palette_src IS NULL THEN
        RETURN 'resolved_composition.lineage.palette_source is required';
END IF;
    IF palette_src NOT IN ('fingerprint','library_reuse','mission_hint',
                           'design_intent_values','archetype_default') THEN
        RETURN format('resolved_composition.lineage.palette_source %L is not a known value', palette_src);
END IF;

    layout_src := lineage ->> 'layout_source';
    IF layout_src IS NULL THEN
        RETURN 'resolved_composition.lineage.layout_source is required';
END IF;
    IF layout_src NOT IN ('library_match','library_fallback','mission_hint',
                          'needs_new_layout_candidate') THEN
        RETURN format('resolved_composition.lineage.layout_source %L is not a known value', layout_src);
END IF;

    typography_src := lineage ->> 'typography_source';
    IF typography_src IS NULL THEN
        RETURN 'resolved_composition.lineage.typography_source is required';
END IF;
    IF typography_src NOT IN ('fingerprint_font_family_match','archetype_default',
                              'layout_default','mission_hint','fallback_sans_modern') THEN
        RETURN format('resolved_composition.lineage.typography_source %L is not a known value', typography_src);
END IF;

    IF (spec ->> 'reasoning') IS NULL OR (spec ->> 'reasoning') = '' THEN
        RETURN 'resolved_composition.reasoning is required';
END IF;
    IF (spec ->> 'resolved_by') IS NULL THEN
        RETURN 'resolved_composition.resolved_by is required';
END IF;
    IF (spec ->> 'resolved_at') IS NULL THEN
        RETURN 'resolved_composition.resolved_at is required';
END IF;

RETURN NULL;
END;
$$;


-- =====================================================================
-- Self-test: sanity-check validators with expected inputs and outputs
-- =====================================================================
DO $$
DECLARE
err text;
BEGIN
    -- navigation: happy path
    err := validate_navigation_spec('{
        "architecture":   "horizontal-top",
        "primary_items":  ["Home","About"],
        "mobile":         "hamburger-slide"
    }'::jsonb);
    IF err IS NOT NULL THEN
        RAISE EXCEPTION 'Navigation happy-path failed: %', err;
END IF;

    -- navigation: missing required
    err := validate_navigation_spec('{"primary_items":["Home"]}'::jsonb);
    IF err IS NULL THEN
        RAISE EXCEPTION 'Navigation missing-field should have failed';
END IF;

    -- navigation: bad enum
    err := validate_navigation_spec('{
        "architecture":"spaceship",
        "primary_items":["Home"],
        "mobile":"hamburger-slide"
    }'::jsonb);
    IF err IS NULL THEN
        RAISE EXCEPTION 'Navigation bad-enum should have failed';
END IF;

    -- layout: happy path
    err := validate_layout_spec('{
        "default_page_layout":"full-width-stacked",
        "header_style":"dark-professional",
        "footer_style":"4-column"
    }'::jsonb);
    IF err IS NOT NULL THEN
        RAISE EXCEPTION 'Layout happy-path failed: %', err;
END IF;

    -- resolved_composition: happy path
    err := validate_resolved_composition_spec('{
        "css_theme_id":"00000000-0000-0000-0000-000000000001",
        "css_theme_name":"t",
        "palette_id":"00000000-0000-0000-0000-000000000002",
        "palette_name":"p",
        "layout_id":"00000000-0000-0000-0000-000000000003",
        "layout_name":"l",
        "typography_set_id":"00000000-0000-0000-0000-000000000004",
        "typography_name":"tp",
        "lineage":{
            "palette_source":"fingerprint",
            "layout_source":"library_match",
            "typography_source":"archetype_default"
        },
        "reasoning":"test",
        "resolved_by":"site-design-planner",
        "resolved_at":"2026-04-18T00:00:00Z"
    }'::jsonb);
    IF err IS NOT NULL THEN
        RAISE EXCEPTION 'Resolved-composition happy-path failed: %', err;
END IF;

    RAISE NOTICE 'Validator self-tests pass';
END $$;