-- ===========================================================================
-- MIGRATION: Schema Mode Infrastructure for Flexible/Strict Rendering
-- File: 043_schema_mode_infrastructure.sql
-- ===========================================================================
-- Adds support for:
--   - schema_mode on sites (flexible/strict default behavior)
--   - schema_snapshot on page_components (locks schema at approval)
--   - content_snapshot on page_components (stores approved content)
--   - component_version tracking
-- ===========================================================================

BEGIN;

-- ===========================================================================
-- PART 1: SITES TABLE - Default schema mode for new sections
-- ===========================================================================

ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    schema_mode TEXT DEFAULT 'flexible';
COMMENT ON COLUMN sites.schema_mode IS
    'Default rendering mode for new sections: flexible (best-effort, warn on missing) or strict (fail on schema mismatch)';

-- When to transition to strict mode
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    strict_mode_trigger TEXT DEFAULT 'first_deploy';
COMMENT ON COLUMN sites.strict_mode_trigger IS
    'When to lock sections to strict mode: hitl (on human approval), first_deploy (on first successful deploy), manual (never auto-transition)';

-- ===========================================================================
-- PART 2: PAGE_COMPONENTS TABLE - Per-section schema tracking
-- ===========================================================================

-- The locked schema for this section (set at approval time)
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    schema_snapshot JSONB;
COMMENT ON COLUMN page_components.schema_snapshot IS
    'Locked input_schema from component at approval time. Edits must match this schema in strict mode.';

-- The content values that were approved
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    content_snapshot JSONB;
COMMENT ON COLUMN page_components.content_snapshot IS
    'The actual content values used when approved. Used for edit comparison, rollback, and form pre-population.';

-- Which component version this was built with
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    component_version_id UUID;
COMMENT ON COLUMN page_components.component_version_id IS
    'Reference to specific component version (if versioning enabled). Ensures template consistency.';

-- Per-section schema mode (overrides site default)
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    schema_mode TEXT;
COMMENT ON COLUMN page_components.schema_mode IS
    'Section-specific schema mode. NULL = inherit from site. Set to flexible/strict to override.';

-- When the section was locked to strict mode
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    locked_at TIMESTAMPTZ;
COMMENT ON COLUMN page_components.locked_at IS
    'Timestamp when section was locked to strict mode';

-- Who/what locked it
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    locked_by TEXT;
COMMENT ON COLUMN page_components.locked_by IS
    'What triggered strict mode lock: hitl, auto_eval, first_deploy, manual';

-- ===========================================================================
-- PART 3: COMPONENT VERSIONING TABLE
-- ===========================================================================

CREATE TABLE IF NOT EXISTS component_versions (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Which component this is a version of
    component_id UUID NOT NULL REFERENCES content_components(id) ON DELETE CASCADE,

    -- Version number (increments on each change)
    version_number INTEGER NOT NULL DEFAULT 1,

    -- Snapshot of the component at this version
    html_template TEXT NOT NULL,
    css_template TEXT,
    input_schema JSONB,

    -- Change tracking
    change_description TEXT,
    changed_by TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT now(),

    -- Unique constraint
    UNIQUE(component_id, version_number)
    );

COMMENT ON TABLE component_versions IS
    'Versioned snapshots of component templates. Allows strict mode pages to use specific versions.';

-- Index for lookups
CREATE INDEX IF NOT EXISTS idx_component_versions_component
    ON component_versions(component_id, version_number DESC);

-- ===========================================================================
-- PART 4: FUNCTION TO LOCK SECTION TO STRICT MODE
-- ===========================================================================

CREATE OR REPLACE FUNCTION lock_section_to_strict(
    p_page_component_id UUID,
    p_content_data JSONB,
    p_locked_by TEXT DEFAULT 'manual'
) RETURNS VOID AS $$
DECLARE
v_component_id UUID;
    v_input_schema JSONB;
BEGIN
    -- Get the component's current input_schema
SELECT pc.component_id, cc.input_schema
INTO v_component_id, v_input_schema
FROM page_components pc
         JOIN content_components cc ON pc.component_id = cc.id
WHERE pc.id = p_page_component_id;

IF NOT FOUND THEN
        RAISE EXCEPTION 'Page component not found: %', p_page_component_id;
END IF;

    -- Lock the section
UPDATE page_components SET
                           schema_mode = 'strict',
                           schema_snapshot = v_input_schema,
                           content_snapshot = p_content_data,
                           locked_at = now(),
                           locked_by = p_locked_by
WHERE id = p_page_component_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION lock_section_to_strict IS
    'Locks a page section to strict schema mode, capturing the current schema and content.';

-- ===========================================================================
-- PART 5: FUNCTION TO UNLOCK SECTION (for redesign)
-- ===========================================================================

CREATE OR REPLACE FUNCTION unlock_section_for_redesign(
    p_page_component_id UUID,
    p_preserve_content BOOLEAN DEFAULT true
) RETURNS VOID AS $$
BEGIN
UPDATE page_components SET
                           schema_mode = 'flexible',
                           -- Optionally preserve snapshots for reference
                           schema_snapshot = CASE WHEN p_preserve_content THEN schema_snapshot ELSE NULL END,
                           content_snapshot = CASE WHEN p_preserve_content THEN content_snapshot ELSE NULL END,
                           locked_at = NULL,
                           locked_by = NULL
WHERE id = p_page_component_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION unlock_section_for_redesign IS
    'Unlocks a section from strict mode, allowing flexible rendering during redesign.';

-- ===========================================================================
-- PART 6: VIEW FOR SECTION SCHEMA STATUS
-- ===========================================================================

CREATE OR REPLACE VIEW v_section_schema_status AS
SELECT
    pc.id AS page_component_id,
    pc.page_id,
    p.name AS page_name,
    s.domain,
    cc.name AS component_name,
    cc.function AS component_function,
    COALESCE(pc.schema_mode, s.schema_mode, 'flexible') AS effective_schema_mode,
    pc.schema_snapshot IS NOT NULL AS has_schema_snapshot,
    pc.content_snapshot IS NOT NULL AS has_content_snapshot,
    pc.locked_at,
    pc.locked_by,
    pc.build_status,
    pc.reviewed_at
FROM page_components pc
         JOIN pages p ON pc.page_id = p.id
         JOIN sites s ON p.site_id = s.id
         LEFT JOIN content_components cc ON pc.component_id = cc.id;

COMMENT ON VIEW v_section_schema_status IS
    'Shows effective schema mode and lock status for all page sections.';

-- ===========================================================================
-- PART 7: TRIGGER TO AUTO-LOCK ON FIRST DEPLOY (if site configured)
-- ===========================================================================

CREATE OR REPLACE FUNCTION auto_lock_on_deploy() RETURNS TRIGGER AS $$
BEGIN
    -- Only act if changing to 'deployed' status
    IF NEW.build_status = 'deployed' AND OLD.build_status != 'deployed' THEN
        -- Check if site is configured for first_deploy locking
        IF EXISTS (
            SELECT 1 FROM pages p
            JOIN sites s ON p.site_id = s.id
            WHERE p.id = NEW.page_id
            AND s.strict_mode_trigger = 'first_deploy'
        ) THEN
            -- Lock to strict mode if not already locked
            IF NEW.schema_mode IS NULL OR NEW.schema_mode = 'flexible' THEN
                NEW.schema_mode := 'strict';
                NEW.locked_at := now();
                NEW.locked_by := 'first_deploy';
                -- Note: schema_snapshot and content_snapshot should be set before deploy
END IF;
END IF;
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger if it doesn't exist
DROP TRIGGER IF EXISTS trigger_auto_lock_on_deploy ON page_components;
CREATE TRIGGER trigger_auto_lock_on_deploy
    BEFORE UPDATE ON page_components
    FOR EACH ROW
    EXECUTE FUNCTION auto_lock_on_deploy();

-- ===========================================================================
-- PART 8: INDEXES
-- ===========================================================================

CREATE INDEX IF NOT EXISTS idx_page_components_schema_mode
    ON page_components(schema_mode) WHERE schema_mode IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_page_components_locked
    ON page_components(locked_at) WHERE locked_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sites_schema_mode
    ON sites(schema_mode);

COMMIT;

-- ===========================================================================
-- VERIFICATION QUERIES
-- ===========================================================================

-- Check new columns on sites
-- SELECT column_name, data_type, column_default
-- FROM information_schema.columns
-- WHERE table_name = 'sites' AND column_name IN ('schema_mode', 'strict_mode_trigger');

-- Check new columns on page_components
-- SELECT column_name, data_type
-- FROM information_schema.columns
-- WHERE table_name = 'page_components'
-- AND column_name IN ('schema_mode', 'schema_snapshot', 'content_snapshot', 'locked_at', 'locked_by');

-- Check component_versions table
-- SELECT * FROM component_versions LIMIT 5;

-- Check section schema status view
-- SELECT * FROM v_section_schema_status LIMIT 10;



UPDATE pages
SET nav_label = 'Case Studies'
WHERE name = 'case-studies'
  AND site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk');