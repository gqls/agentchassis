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