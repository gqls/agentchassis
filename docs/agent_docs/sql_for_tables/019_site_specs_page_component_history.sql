-- ============================================================================
-- BLOCK A: DB migrations for the work-item build pipeline (Phase 0)
--
-- No Go changes needed. Can be applied immediately.
-- Run against clients_db.
--
-- Contents:
--   1. site_specs table
--   2. build_queue table
--   3. page_component_history table
--   4. approval_mode column on site_work_items
--   5. page_spec column on pages
--   6. Drop content_snapshot / schema_snapshot from page_components
--
-- NOTE on item 6: The Go struct PageComponentInstance has a SchemaSnapshot
-- field (omitempty). After applying this migration, remove that struct field
-- from the Go code. It's not actively used by any action — just a leftover
-- from an earlier design that this migration replaces.
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. site_specs
--
-- Stores all site specification data (identity, strategy, tone, visual_direction,
-- design, image_guidance, structure, marketing, adoption_source).
-- Each row is a complete record for one aspect. write_site_spec deep-merges
-- partial updates before inserting, so every row is self-contained.
-- History via is_current flag — old versions remain for rollback/audit.
-- ============================================================================

CREATE TABLE IF NOT EXISTS site_specs (
                                          id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id         uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    aspect          text NOT NULL,
    data            jsonb NOT NULL,

    -- Provenance
    source          text NOT NULL,       -- 'classifier', 'adoption', 'hitl',
-- 'planner', 'improvement', 'seed', 'manual',
-- 'rollback', 'fork', 'recovery'
    source_agent    text,                -- agent type that wrote this
    source_item_id  uuid,                -- work item that caused this write
    notes           text,                -- human-readable reason

-- Currency
    is_current      boolean NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    superseded_at   timestamptz,
    created_by      text NOT NULL
    );

-- One current spec per site per aspect (enforced)
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_specs_current
    ON site_specs (site_id, aspect)
    WHERE is_current = true;

-- Fast lookup: all current specs for a site
CREATE INDEX IF NOT EXISTS idx_site_specs_lookup
    ON site_specs (site_id)
    WHERE is_current = true;

-- History queries: all versions of a given aspect, newest first
CREATE INDEX IF NOT EXISTS idx_site_specs_history
    ON site_specs (site_id, aspect, created_at DESC);

-- Find specs written by a particular work item
CREATE INDEX IF NOT EXISTS idx_site_specs_source_item
    ON site_specs (source_item_id)
    WHERE source_item_id IS NOT NULL;


-- ============================================================================
-- 2. build_queue
--
-- Domain queue for seeding new sites into the work-item pipeline.
-- seed_build_queue action reads from here, creates site records and
-- initial work items based on the direction field.
-- ============================================================================

CREATE TABLE IF NOT EXISTS build_queue (
                                           id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    domain          text UNIQUE NOT NULL,
    direction       jsonb,               -- null, {objective}, {adopt_from}, {fork_from}, {brief_complete: true, ...}
    status          text NOT NULL DEFAULT 'queued',
    batch_id        uuid,
    priority        integer DEFAULT 100,
    created_at      timestamptz DEFAULT now(),
    updated_at      timestamptz DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_build_queue_status
    ON build_queue (status, priority)
    WHERE status = 'queued';


-- ============================================================================
-- 3. page_component_history
--
-- Tracks previous content_data values for page_components.
-- Before any content_data write, the current value is copied here.
-- Each row is a complete snapshot (not a diff).
-- Replaces the unused content_snapshot/schema_snapshot columns.
-- ============================================================================

CREATE TABLE IF NOT EXISTS page_component_history (
                                                      id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    component_id    uuid REFERENCES page_components(id) ON DELETE SET NULL,
    page_id         uuid NOT NULL REFERENCES pages(id),
    site_id         uuid NOT NULL REFERENCES sites(id),
    content_data    jsonb NOT NULL,

    source          text NOT NULL,       -- 'content-writer', 'section-editor', 'rollback', etc.
    source_item_id  uuid,                -- work item that triggered the change
    created_at      timestamptz NOT NULL DEFAULT now()
    );

-- Recent history per component (for rollback)
CREATE INDEX IF NOT EXISTS idx_pch_component
    ON page_component_history (component_id, created_at DESC);

-- Recent history per site (for site-wide audit)
CREATE INDEX IF NOT EXISTS idx_pch_site
    ON page_component_history (site_id, created_at DESC);

-- Find history entries from a specific work item
CREATE INDEX IF NOT EXISTS idx_pch_source_item
    ON page_component_history (source_item_id)
    WHERE source_item_id IS NOT NULL;


-- ============================================================================
-- 4. approval_mode on site_work_items
--
-- Controls whether items auto-dispatch or require human/eval approval.
-- Values: 'auto' (default), 'hitl', 'eval'
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'site_work_items' AND column_name = 'approval_mode'
    ) THEN
ALTER TABLE site_work_items ADD COLUMN approval_mode text DEFAULT 'auto';
END IF;
END $$;


-- ============================================================================
-- 5. page_spec on pages
--
-- JSONB column for page-level specification data: content hints, existing
-- content (for adoption), content_direction for rewrites. Nullable — most
-- pages won't have one until adoption or improvement agents write specs.
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'pages' AND column_name = 'page_spec'
    ) THEN
ALTER TABLE pages ADD COLUMN page_spec jsonb;
END IF;
END $$;


-- ============================================================================
-- 6. Drop content_snapshot / schema_snapshot from page_components
--
-- These are unused. page_component_history replaces them.
-- The Go struct has SchemaSnapshot (omitempty) — remove it after this runs.
--
-- v_section_schema_status references both columns. Not used in Go code
-- but useful as a diagnostic view, so recreate it without the snapshot columns.
-- ============================================================================

DROP VIEW IF EXISTS v_section_schema_status;
ALTER TABLE page_components DROP COLUMN IF EXISTS content_snapshot;
ALTER TABLE page_components DROP COLUMN IF EXISTS schema_snapshot;

CREATE VIEW v_section_schema_status AS
SELECT pc.id AS page_component_id,
       pc.page_id,
       p.name AS page_name,
       s.domain,
       cc.name AS component_name,
       cc.function AS component_function,
       COALESCE(pc.schema_mode, s.schema_mode, 'flexible'::text) AS effective_schema_mode,
       pc.locked_at,
       pc.locked_by,
       pc.build_status,
       pc.reviewed_at
FROM page_components pc
         JOIN pages p ON pc.page_id = p.id
         JOIN sites s ON p.site_id = s.id
         LEFT JOIN content_components cc ON pc.component_id = cc.id;


COMMIT;

-- ============================================================================
-- Verification queries (run after migration to confirm)
-- ============================================================================

-- SELECT count(*) FROM information_schema.tables WHERE table_name IN ('site_specs', 'build_queue', 'page_component_history');
-- Expected: 3

-- SELECT column_name FROM information_schema.columns WHERE table_name = 'site_work_items' AND column_name = 'approval_mode';
-- Expected: 1 row

-- SELECT column_name FROM information_schema.columns WHERE table_name = 'pages' AND column_name = 'page_spec';
-- Expected: 1 row

-- SELECT column_name FROM information_schema.columns WHERE table_name = 'page_components' AND column_name IN ('content_snapshot', 'schema_snapshot');
-- Expected: 0 rows