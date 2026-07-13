-- =====================================================================
-- MIGRATION: 025 Phase 2 — Palette/Layout/Typography tables
-- =====================================================================
-- Splits the concerns currently conflated in css_themes.css_template
-- into three independently-versioned tables. A css_themes row becomes
-- a composition: palette_id + layout_id + typography_set_id.
--
-- Additive and reversible:
--   - Three new tables, empty after this migration
--   - Three new nullable FK columns on css_themes, all NULL after this
--   - No existing css_themes row is modified
--   - The renderer still uses css_themes.css_template (the legacy path);
--     the new columns are read only once Phase 4 ships
--
-- Phase 3 seeds the new tables and populates the FK columns.
-- Phase 4 rewrites the renderer to read via the FK columns.
-- Phase 7 drops the legacy columns (css_template, css_content,
--          color_palette, typography) once confidence is established.
--
-- Rollback:
--   ALTER TABLE style_collections DROP COLUMN IF EXISTS forked_at;
--   ALTER TABLE style_collections DROP COLUMN IF EXISTS source_domain;
--   ALTER TABLE style_collections DROP COLUMN IF EXISTS source_site_id;
--   ALTER TABLE style_collections DROP COLUMN IF EXISTS forked_from_collection_id;
--   ALTER TABLE style_collections DROP COLUMN IF EXISTS needs_review;
--   ALTER TABLE style_collections DROP COLUMN IF EXISTS origin;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS forked_at;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS source_domain;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS source_site_id;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS forked_from_theme_id;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS needs_review;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS origin;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS typography_set_id;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS layout_id;
--   ALTER TABLE css_themes DROP COLUMN IF EXISTS palette_id;
--   DROP TABLE IF EXISTS typography_sets CASCADE;
--   DROP TABLE IF EXISTS layouts CASCADE;
--   DROP TABLE IF EXISTS palettes CASCADE;
--
-- Run:
--   psql -d clients_db -f migration_025_phase2.sql
--
-- Safe to re-run: every statement uses IF NOT EXISTS.
-- =====================================================================

BEGIN;

-- =====================================================================
-- PART 1 — Create the three new tables
-- =====================================================================

-- ---------------------------------------------------------------------
-- palettes: flat jsonb colour map + lineage
-- ---------------------------------------------------------------------
-- colours is a free-shape jsonb map. Slot names (primary, hero_title,
-- cta_bg, etc.) are not constrained at the DB level — they're consumed
-- via the {{palette "key" "fallback"}} template helper which is tolerant
-- of missing keys. Adding a new slot requires no schema change.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS palettes (
                                        id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name                    text UNIQUE NOT NULL,
    display_name            text NOT NULL,
    description             text,
    colours                 jsonb NOT NULL,
    category                text,
    industry_tags           text[] DEFAULT ARRAY[]::text[],
    is_active               boolean NOT NULL DEFAULT true,

    -- Lineage (mirrors css_themes / style_collections)
    origin                  text NOT NULL DEFAULT 'seed',
    needs_review            boolean NOT NULL DEFAULT false,
    forked_from_palette_id  uuid REFERENCES palettes(id) ON DELETE SET NULL,
    source_site_id          uuid REFERENCES sites(id) ON DELETE SET NULL,
    source_domain           text,
    forked_at               timestamptz,

    created_at              timestamptz NOT NULL DEFAULT NOW(),
    updated_at              timestamptz NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_palettes_name
    ON palettes(name);
CREATE INDEX IF NOT EXISTS idx_palettes_category
    ON palettes(category);
CREATE INDEX IF NOT EXISTS idx_palettes_active_reviewed
    ON palettes(is_active, needs_review)
    WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_palettes_source_site
    ON palettes(source_site_id)
    WHERE source_site_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_palettes_industry
    ON palettes USING gin (industry_tags);

-- ---------------------------------------------------------------------
-- layouts: Go CSS template + structure tokens + default header/footer
-- ---------------------------------------------------------------------
-- css_template is the Go-templated CSS parsed with palette/typo/token
-- helper funcs (defined in the renderer). structure_tokens holds
-- container widths, padding scales, border radii, shadows etc. —
-- visible to the template via {{token ...}}.
--
-- default_header_component_id / default_footer_component_id point at
-- content_components rows that ship as defaults when a style_collection
-- selects this layout. Both are optional; style_collections may pick
-- different header/footer components.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS layouts (
                                       id                           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name                         text UNIQUE NOT NULL,
    display_name                 text NOT NULL,
    description                  text,
    css_template                 text NOT NULL,
    structure_tokens             jsonb NOT NULL DEFAULT '{}'::jsonb,
    category                     text,
    industry_tags                text[] DEFAULT ARRAY[]::text[],

    -- Default header/footer components (optional)
    default_header_component_id  uuid REFERENCES content_components(id) ON DELETE SET NULL,
    default_footer_component_id  uuid REFERENCES content_components(id) ON DELETE SET NULL,

    is_active                    boolean NOT NULL DEFAULT true,

    -- Lineage
    origin                       text NOT NULL DEFAULT 'seed',
    needs_review                 boolean NOT NULL DEFAULT false,
    forked_from_layout_id        uuid REFERENCES layouts(id) ON DELETE SET NULL,
    source_site_id               uuid REFERENCES sites(id) ON DELETE SET NULL,
    source_domain                text,
    forked_at                    timestamptz,

    created_at                   timestamptz NOT NULL DEFAULT NOW(),
    updated_at                   timestamptz NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_layouts_name
    ON layouts(name);
CREATE INDEX IF NOT EXISTS idx_layouts_category
    ON layouts(category);
CREATE INDEX IF NOT EXISTS idx_layouts_active_reviewed
    ON layouts(is_active, needs_review)
    WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_layouts_source_site
    ON layouts(source_site_id)
    WHERE source_site_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_layouts_industry
    ON layouts USING gin (industry_tags);

-- ---------------------------------------------------------------------
-- typography_sets: font stacks + scale
-- ---------------------------------------------------------------------
-- fonts holds font_family, heading_font, mono_font, etc. as a jsonb map.
-- scale holds base_size, line_height, h1_ratio, etc. Both are consumed
-- via the {{typo "key" "fallback"}} template helper.
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS typography_sets (
                                               id                             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name                           text UNIQUE NOT NULL,
    display_name                   text NOT NULL,
    description                    text,
    fonts                          jsonb NOT NULL,
    scale                          jsonb NOT NULL DEFAULT '{}'::jsonb,
    category                       text,
    industry_tags                  text[] DEFAULT ARRAY[]::text[],
    is_active                      boolean NOT NULL DEFAULT true,

    -- Lineage
    origin                         text NOT NULL DEFAULT 'seed',
    needs_review                   boolean NOT NULL DEFAULT false,
    forked_from_typography_set_id  uuid REFERENCES typography_sets(id) ON DELETE SET NULL,
    source_site_id                 uuid REFERENCES sites(id) ON DELETE SET NULL,
    source_domain                  text,
    forked_at                      timestamptz,

    created_at                     timestamptz NOT NULL DEFAULT NOW(),
    updated_at                     timestamptz NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_typography_sets_name
    ON typography_sets(name);
CREATE INDEX IF NOT EXISTS idx_typography_sets_category
    ON typography_sets(category);
CREATE INDEX IF NOT EXISTS idx_typography_sets_active_reviewed
    ON typography_sets(is_active, needs_review)
    WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_typography_sets_industry
    ON typography_sets USING gin (industry_tags);
CREATE INDEX IF NOT EXISTS idx_typography_sets_source_site
    ON typography_sets(source_site_id)
    WHERE source_site_id IS NOT NULL;

-- =====================================================================
-- PART 2 — Alter existing tables
-- ---------------------------------------------------------------------
-- Two things happen here:
--
--   (a) Add three nullable FK columns to css_themes (palette_id,
--       layout_id, typography_set_id). These are what the Phase 4
--       renderer reads. Nullable so existing rows stay valid; Phase 3
--       populates them via UPDATE statements keyed on css_themes.name.
--
--   (b) Add lineage columns to css_themes AND style_collections
--       (origin, needs_review, forked_from_*, source_site_id,
--       source_domain, forked_at). These are required by the Phase 5
--       fork_theme_from_site action. A prior session reported them as
--       already added but the current schema shows them absent.
--       All existing rows default to origin='seed', needs_review=false,
--       which correctly reflects that nothing in the library was
--       adopted via the fork action yet.
--
-- ON DELETE SET NULL everywhere — same rationale as Part 1.
-- =====================================================================

-- ---------------------------------------------------------------------
-- css_themes: new FK columns + lineage
-- ---------------------------------------------------------------------

ALTER TABLE css_themes
    ADD COLUMN IF NOT EXISTS palette_id        uuid REFERENCES palettes(id)        ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS layout_id         uuid REFERENCES layouts(id)         ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS typography_set_id uuid REFERENCES typography_sets(id) ON DELETE SET NULL;

ALTER TABLE css_themes
    ADD COLUMN IF NOT EXISTS origin               text NOT NULL DEFAULT 'seed',
    ADD COLUMN IF NOT EXISTS needs_review         boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS forked_from_theme_id uuid REFERENCES css_themes(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS source_site_id       uuid REFERENCES sites(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS source_domain        text,
    ADD COLUMN IF NOT EXISTS forked_at            timestamptz;

CREATE INDEX IF NOT EXISTS idx_css_themes_palette_id
    ON css_themes(palette_id)
    WHERE palette_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_css_themes_layout_id
    ON css_themes(layout_id)
    WHERE layout_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_css_themes_typography_set_id
    ON css_themes(typography_set_id)
    WHERE typography_set_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_css_themes_active_reviewed
    ON css_themes(is_active, needs_review)
    WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_css_themes_source_site
    ON css_themes(source_site_id)
    WHERE source_site_id IS NOT NULL;

-- ---------------------------------------------------------------------
-- style_collections: lineage columns only
-- ---------------------------------------------------------------------
-- Same lineage shape as css_themes. A style_collection is a composition
-- row (header component + footer component + theme + palette + typo);
-- when the fork action produces one, it needs the same provenance fields.
-- ---------------------------------------------------------------------

ALTER TABLE style_collections
    ADD COLUMN IF NOT EXISTS origin                    text NOT NULL DEFAULT 'seed',
    ADD COLUMN IF NOT EXISTS needs_review              boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS forked_from_collection_id uuid REFERENCES style_collections(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS source_site_id            uuid REFERENCES sites(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS source_domain             text,
    ADD COLUMN IF NOT EXISTS forked_at                 timestamptz;

CREATE INDEX IF NOT EXISTS idx_style_collections_active_reviewed
    ON style_collections(is_active, needs_review)
    WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_style_collections_source_site
    ON style_collections(source_site_id)
    WHERE source_site_id IS NOT NULL;

COMMIT;

-- =====================================================================
-- Post-run verification
-- =====================================================================
-- Run these in psql to confirm the migration landed:
--
--   \d palettes
--   \d layouts
--   \d typography_sets
--   \d css_themes
--   \d style_collections
--
--   SELECT count(*) FROM palettes;        -- expect 0
--   SELECT count(*) FROM layouts;         -- expect 0
--   SELECT count(*) FROM typography_sets; -- expect 0
--
--   -- Phase 3 will fill the FK columns; for now all should be 0
--   SELECT count(*) AS themes_total,
--          count(palette_id) AS with_palette,
--          count(layout_id) AS with_layout,
--          count(typography_set_id) AS with_typography
--     FROM css_themes;
--
--   -- All existing rows should default to origin='seed'
--   SELECT origin, count(*) FROM css_themes GROUP BY origin;
--   SELECT origin, count(*) FROM style_collections GROUP BY origin;
--
--   -- Nothing should need review (fork action hasn't shipped yet)
--   SELECT count(*) FROM css_themes WHERE needs_review = true;         -- expect 0
--   SELECT count(*) FROM style_collections WHERE needs_review = true;  -- expect 0
-- =====================================================================