-- ============================================================================
-- Migration: Enhance pages table for pageflow build workflow
-- Database: clients_db
-- Run with: psql -h <host> -U clients_user -d clients_db -f 001_enhance_pages_for_pageflow.sql
-- ============================================================================

BEGIN;

-- 1. Add build workflow status (separate from display status)
-- Values: planned, content_draft, pending_review, approved, deployed, needs_rebuild
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS build_status VARCHAR(50) DEFAULT 'planned';

-- 2. Add sections from site plan (what sections this page should have)
-- Structure: ["Hero Section", "Features Grid", "Social Proof", "Call to Action"]
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS sections JSONB DEFAULT '[]'::jsonb;

-- 3. Add content_data for generated content per section
-- Structure: {"hero": {"headline": "...", "subheadline": "..."}, "features": {...}}
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS content_data JSONB DEFAULT '{}'::jsonb;

-- 4. Add rendered HTML (final output)
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS html TEXT;

-- 5. Add version tracking for edit history
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

-- 6. Add who/what last modified (for HITL tracking)
-- Values: "agent:content-writer", "user:email@example.com", "hitl:reviewer"
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS last_modified_by VARCHAR(100);

-- 7. Add timestamps for build workflow stages
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS content_generated_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS deployed_at TIMESTAMP WITH TIME ZONE;

-- 8. Create indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_pages_build_status
    ON pages(site_id, build_status);

CREATE INDEX IF NOT EXISTS idx_pages_needs_build
    ON pages(site_id)
    WHERE build_status IN ('planned', 'needs_rebuild');

-- 9. Set initial build_status for existing pages
UPDATE pages
SET build_status = CASE
                       WHEN html IS NOT NULL AND html != '' THEN 'deployed'
                       ELSE 'planned'
    END
WHERE build_status IS NULL;

-- 10. Add comments for documentation
COMMENT ON COLUMN pages.build_status IS 'Workflow state: planned, content_draft, pending_review, approved, deployed, needs_rebuild';
COMMENT ON COLUMN pages.sections IS 'Array of section names this page should contain, from site plan';
COMMENT ON COLUMN pages.content_data IS 'Generated content for each section, keyed by section name';
COMMENT ON COLUMN pages.html IS 'Final rendered HTML for the page';
COMMENT ON COLUMN pages.version IS 'Content version number, incremented on each edit';
COMMENT ON COLUMN pages.last_modified_by IS 'Agent or user who last modified: agent:content-writer, user:email, hitl:reviewer';

COMMIT;

-- Verify the changes
\d pages

----

need to reverse out some of that and do just this:

   -- ============================================================================
-- Migration: Add build workflow tracking to pages table
-- Database: clients_db
-- Run with: psql -h <host> -U clients_user -d clients_db -f 001_pages_build_status.sql
-- ============================================================================
--
-- NOTE: Content (rendered_html, content_data) lives in page_components table
-- Pages table just needs workflow tracking fields
-- ============================================================================

BEGIN;

-- 1. Add build workflow status for tracking page generation progress
-- Values: planned, writing_content, content_ready, building_html, deployed, needs_rebuild
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS build_status VARCHAR(50) DEFAULT 'planned';

-- 2. Add sections from site plan (what sections this page should have)
-- This is planning reference - actual content goes in page_components
-- Structure: ["Hero Section", "Features Grid", "Social Proof", "Call to Action"]
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS sections JSONB DEFAULT '[]'::jsonb;

-- 3. Add version tracking for rebuild detection
ALTER TABLE pages
    ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1;

-- 4. Create index on build_status for efficient querying of pages to build
CREATE INDEX IF NOT EXISTS idx_pages_build_status
    ON pages(site_id, build_status);

-- 5. Set initial build_status for existing pages
UPDATE pages
SET build_status = CASE
                       WHEN last_built_at IS NOT NULL THEN 'deployed'
                       ELSE 'planned'
    END
WHERE build_status IS NULL;

-- 6. Add comments for documentation
COMMENT ON COLUMN pages.build_status IS 'Workflow state: planned, writing_content, content_ready, building_html, deployed, needs_rebuild';
COMMENT ON COLUMN pages.sections IS 'Array of section names from site plan - actual content in page_components table';
COMMENT ON COLUMN pages.version IS 'Page version for rebuild detection';

COMMIT;

-- Verify the changes
SELECT column_name, data_type, column_default
FROM information_schema.columns
WHERE table_name = 'pages'
  AND column_name IN ('build_status', 'sections', 'version');


----


-- ============================================================================
-- Migration: Revert pages table to minimal pageflow structure
-- Database: clients_db
-- ============================================================================
--
-- Removes extra columns that were added but aren't needed
-- (content lives in page_components table, not directly on pages)
--
-- Target structure keeps only:
--   - Original columns from 002_links_clients_networks_etc_tables.sql
--   - build_status (workflow tracking)
--   - sections (planning reference)
--   - version (rebuild detection)
-- ============================================================================

BEGIN;

-- 1. Drop columns that belong in page_components or aren't needed
ALTER TABLE pages DROP COLUMN IF EXISTS deploy_commit;
ALTER TABLE pages DROP COLUMN IF EXISTS deployed_at;
ALTER TABLE pages DROP COLUMN IF EXISTS content_data;
ALTER TABLE pages DROP COLUMN IF EXISTS html;
ALTER TABLE pages DROP COLUMN IF EXISTS last_modified_by;
ALTER TABLE pages DROP COLUMN IF EXISTS content_generated_at;
ALTER TABLE pages DROP COLUMN IF EXISTS reviewed_at;

-- 2. Fix build_status type and default
-- First drop the index that depends on it
DROP INDEX IF EXISTS idx_pages_needs_build;
DROP INDEX IF EXISTS idx_pages_build_status;

-- Change type from TEXT to VARCHAR(50) and fix default
ALTER TABLE pages
ALTER COLUMN build_status TYPE VARCHAR(50),
ALTER COLUMN build_status SET DEFAULT 'planned';

-- Update existing values if needed
UPDATE pages SET build_status = 'planned' WHERE build_status = 'pending';
UPDATE pages SET build_status = 'deployed' WHERE build_status IS NULL AND last_built_at IS NOT NULL;
UPDATE pages SET build_status = 'planned' WHERE build_status IS NULL;

-- 3. Recreate indexes with correct types
CREATE INDEX idx_pages_build_status ON pages(site_id, build_status);

CREATE INDEX idx_pages_needs_build ON pages(site_id)
    WHERE build_status IN ('planned', 'needs_rebuild');

-- 4. Add comments
COMMENT ON COLUMN pages.build_status IS 'Workflow state: planned, writing_content, content_ready, building_html, deployed, needs_rebuild';
COMMENT ON COLUMN pages.sections IS 'Array of section names from site plan - actual content in page_components table';
COMMENT ON COLUMN pages.version IS 'Page version for rebuild detection';

COMMIT;

-- Verify final structure
\d pages

---

   -- Assign default sections to pages with empty sections
UPDATE pages
SET sections = '["hero", "generic-text-block", "call_to_action"]'::jsonb
WHERE site_id = '4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND (sections = '[]'::jsonb OR sections IS NULL);


-- Option A: Add to pages table
ALTER TABLE pages ADD COLUMN rendered_header TEXT;
ALTER TABLE pages ADD COLUMN rendered_footer TEXT;
ALTER TABLE pages ADD COLUMN rendered_head TEXT;

-- Then rerender just does:
SELECT rendered_head, rendered_header, rendered_footer FROM pages WHERE id = $1;
-- + load sections from page_components
-- = minimal assembly


- ============================================================
-- 4. Add site_area_id to pages table
-- ============================================================
ALTER TABLE pages ADD COLUMN IF NOT EXISTS site_area_id UUID REFERENCES site_areas(id);

CREATE INDEX IF NOT EXISTS idx_pages_area ON pages(site_area_id) WHERE site_area_id IS NOT NULL;

COMMENT ON COLUMN pages.site_area_id IS 'Optional area this page belongs to (for area-specific styling)';

--

        -- ============================================================
-- 6. Helper function to get effective component for a page
-- ============================================================
CREATE OR REPLACE FUNCTION get_page_component(
    p_page_id UUID,
    p_slot_name VARCHAR(100)
) RETURNS TEXT AS $$
DECLARE
v_site_id UUID;
    v_area_id UUID;
    v_html TEXT;
BEGIN
    -- Get page's site and area
SELECT site_id, site_area_id INTO v_site_id, v_area_id
FROM pages WHERE id = p_page_id;

-- Try area-level first (if page is in an area)
IF v_area_id IS NOT NULL THEN
SELECT rendered_html INTO v_html
FROM area_components
WHERE area_id = v_area_id AND slot_name = p_slot_name;

IF v_html IS NOT NULL THEN
            RETURN v_html;
END IF;
END IF;

    -- Fall back to site-level
SELECT rendered_html INTO v_html
FROM site_components
WHERE site_id = v_site_id AND slot_name = p_slot_name;

RETURN v_html;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_page_component IS 'Gets the effective component HTML for a page slot, with area override';


---

-- Phase 2/5: Add suppressed_sections column to pages table.
-- Used by the component removal flow to prevent discovery checks
-- from recreating sections that were intentionally removed.
-- The DELETE component endpoint writes to this column.
-- Discovery checks will filter on it in Phase 5.

ALTER TABLE pages ADD COLUMN IF NOT EXISTS suppressed_sections JSONB DEFAULT '[]'::jsonb;

---
-- Fix blog page to be a proper blog-index
UPDATE pages
SET page_type = 'blog-index',
    sections = '["hero", "blog-listing", "call-to-action"]'::jsonb,
    build_status = 'planned'
WHERE id = 'ff56bcaf-cf3c-40bd-a6ee-18703bd3d656';

---

snake case kebab case

      -- ============================================================================
-- Migration: pages.page_type to kebab-case
-- ============================================================================
-- Two structural fixes in one migration:
--
-- 1. Kebab normalisation. Today the same conceptual page_type appears in
--    two forms: 42 rows of 'blog-post' vs 3 rows of 'blog_post', 3 rows of
--    'blog-index' vs 2 of 'blog_index', etc. Most reader code (line 14833,
--    30916, 38036, 42374, 42376 of the chassis) checks against kebab forms,
--    so the snake rows have been silently missed by those queries. The
--    canonicaliser was the only writer producing snake; aligning it with
--    the rest of the codebase + the standards-doc convention makes both
--    sides agree.
--
-- 2. Landing fix (Phase 1.5). The homepage's page_type has been written as
--    'index' by CanonicalisePage, conflating the page's NAME with its
--    TYPE. Name=index is correct (storage convention for the homepage);
--    Type=landing is the correct type. No code currently checks
--    page_type='index' so this is a safe rename — only a comment in
--    actions/load_page_record (line 71908) references the literal value
--    as a possible example, and that comment is being updated alongside.
--
-- After the migration, ALL of pages.page_type matches the regex
-- ^[a-z][a-z0-9]*(-[a-z0-9]+)*$
-- which is enforced by the new CHECK constraint at the bottom.
--
-- Apply with: \i 051_pages_page_type_kebab.sql
-- ============================================================================

BEGIN;

-- ──────────────────────────────────────────────────────────────────────
-- 1. Pre-flight: snapshot current distribution
-- ──────────────────────────────────────────────────────────────────────
SELECT 'before_migration' AS phase, page_type, COUNT(*) AS n
FROM pages
GROUP BY page_type
ORDER BY n DESC;

-- ──────────────────────────────────────────────────────────────────────
-- 2. Kebab data migration
--    Five distinct snake → kebab updates, plus the landing rename.
--    Using individual UPDATEs (rather than one big CASE) so each
--    affected row count is logged separately in the migration output.
-- ──────────────────────────────────────────────────────────────────────
UPDATE pages SET page_type = 'blog-post'        WHERE page_type = 'blog_post';
UPDATE pages SET page_type = 'blog-index'       WHERE page_type = 'blog_index';
UPDATE pages SET page_type = 'section-index'    WHERE page_type = 'section_index';
UPDATE pages SET page_type = 'entity-directory' WHERE page_type = 'entity_directory';
UPDATE pages SET page_type = 'entity-page'      WHERE page_type = 'entity_page';

-- Phase 1.5: landing/index canonical separation. Existing 'index' values
-- become 'landing'. The page's NAME (separate column) remains 'index'
-- where it was — only the TYPE moves.
UPDATE pages SET page_type = 'landing' WHERE page_type = 'index';

-- ──────────────────────────────────────────────────────────────────────
-- 3. Verify nothing snake-shaped remains
-- ──────────────────────────────────────────────────────────────────────
SELECT 'after_migration' AS phase, page_type, COUNT(*) AS n
FROM pages
GROUP BY page_type
ORDER BY n DESC;

-- This should return zero rows. If it returns anything, abort and
-- investigate — the migration is incomplete.
SELECT 'rows_still_violating_kebab' AS metric, COUNT(*) AS value
FROM pages
WHERE page_type !~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$';

-- ──────────────────────────────────────────────────────────────────────
-- 4. Lock it in with a CHECK constraint
--    Matches the regex from the standards doc's Component Naming
--    Contract. Allows empty string (matching the spirit of
--    chk_function_kebab_case on content_components) so legacy rows
--    with NULL/empty page_type aren't broken. Tightening to NOT NULL
--    is a separate decision.
-- ──────────────────────────────────────────────────────────────────────
ALTER TABLE pages
    ADD CONSTRAINT chk_page_type_kebab_case
        CHECK (page_type IS NULL
            OR page_type = ''
            OR page_type ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$');

COMMENT ON CONSTRAINT chk_page_type_kebab_case ON pages IS
    'Page types are kebab-case per the standards doc (data-shaped values). See contracts_and_standards.md for the identifier-shaped vs data-shaped distinction.';

COMMIT;

-- ──────────────────────────────────────────────────────────────────────
-- Post-commit smoke test
-- ──────────────────────────────────────────────────────────────────────
-- Verify the constraint rejects snake-form inserts:
--   INSERT INTO pages (site_id, name, page_type) VALUES (gen_random_uuid(), 'test', 'blog_post');
--   -- Expect: ERROR: new row for relation "pages" violates check constraint "chk_page_type_kebab_case"
--
-- Confirm the readers at line 42374-42376 of the chassis now hit all
-- the data they expected to:
--   SELECT page_type, COUNT(*) FROM pages
--   WHERE page_type IN ('blog-post', 'news-index', 'blog-index', 'sitemap',
--                       'privacy', 'terms', 'entity-directory')
--   GROUP BY page_type;
--   -- Compare with the budget-calculation query at line 42375 — every
--   -- row that was meant to be counted is now counted.