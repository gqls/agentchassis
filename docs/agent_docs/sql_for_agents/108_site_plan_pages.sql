-- ============================================================================
-- Migration 033: Reconcile site_plan_pages columns + drop orphan
--                site_plan_partials (Phase 1 schema repair)
-- ============================================================================
-- Background: an earlier draft of migration 031 created site_plan_pages
-- with a `page_data jsonb` column and a separate `site_plan_partials`
-- table. The current 031 (and the write_site_plan action) instead use:
--
--   - title, meta_description, nav_label as direct columns on
--     site_plan_pages
--   - row-per-directive in site_plan_directives (not partials)
--
-- The earlier 031 was applied to production before the rewrite. The
-- rewritten 031, applied later, hit `CREATE TABLE IF NOT EXISTS` and
-- silently skipped. Result: site_plan_pages is missing 3 columns the
-- action writes; site_plan_partials exists but is unused.
--
-- This migration brings the schema in line with the action contract:
--
--   1. Add title, meta_description, nav_label to site_plan_pages.
--   2. Drop the now-orphan page_data column on site_plan_pages.
--   3. Drop the orphan site_plan_partials table (no rows, no readers).
--
-- All operations are safe: the new tables hold zero rows because every
-- write_site_plan call to date has failed at the title-column error.
-- This migration is idempotent (uses IF EXISTS / IF NOT EXISTS).
-- ============================================================================

BEGIN;

-- 1. Add missing columns. IF NOT EXISTS makes the operation safe to
--    re-run if any of the columns happen to already exist on a given
--    environment.
ALTER TABLE site_plan_pages
    ADD COLUMN IF NOT EXISTS title            TEXT,
    ADD COLUMN IF NOT EXISTS meta_description TEXT,
    ADD COLUMN IF NOT EXISTS nav_label        TEXT;

-- 2. Drop the now-orphan page_data column. The current write_site_plan
--    action does not insert into this column; the data it used to
--    carry (title, meta_description, nav_label) is in dedicated columns
--    above and section data is in site_plan_sections.
ALTER TABLE site_plan_pages
DROP COLUMN IF EXISTS page_data;

-- 3. Drop the orphan site_plan_partials table. The current Phase 1
--    design uses site_plan_directives (row-per-directive) for what
--    partials previously held in JSONB blobs. No code writes to or
--    reads from site_plan_partials.
DROP TABLE IF EXISTS site_plan_partials;

-- 4. Verification — show the final shape of site_plan_pages.
--    Should have 14 columns total (id, plan_id, name, role, slug, url,
--    parent_section, title, meta_description, nav_label, in_header,
--    in_footer, nav_order, created_at). page_data is gone.
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'site_plan_pages'
ORDER BY ordinal_position;

COMMIT;