-- 154: fix product-detail's section list in the AUTHORITATIVE site_plan_sections
--      table (corrects an omission in 153)
--
-- Context (2026-07-15): migration 153 swapped product-detail's components in
-- page_components and updated pages.sections + the site_specs.site_plan
-- aspect. But page-build-handler's load_page_sections_from_spec reads section
-- lists in PRIORITY ORDER (load_page_sections_from_spec_action.go:1-56):
--   1. site_plan_sections table (site_plans family) — AUTHORITATIVE
--   2. site_specs.site_plan aspect
--   3. pages.sections
-- and syncs the winning source DOWN to pages.sections. gripper-detail is NOT
-- in the site_plan_sections table, so its fix via the site_specs aspect
-- (source 2) held. product-detail IS in the table with the OLD components, so
-- on rebuild source 1 served [product-hero, product-specs, call-to-action],
-- re-synced it over pages.sections, and resurrected the deleted components.
-- 153's pages.sections/aspect edits for product-detail were therefore
-- overwritten. This migration fixes source 1.
--
-- After applying, re-drive product-detail's empty_section item so
-- page-build-handler rebuilds from the corrected section list (RUNBOOK §5b:
-- use the dispatch path, not direct kcat).
--
-- Verify after applying:
--   SELECT sps.page_name, sps.ordering, sps.component_name
--   FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id
--   WHERE sp.site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND sp.is_current
--     AND sps.page_name='product-detail' ORDER BY sps.ordering;
--   -- expect: 0 gripper-spec-sheet | 1 call-to-action

BEGIN;

-- Work against the current plan for this site.
WITH cur AS (
    SELECT id AS plan_id FROM site_plans
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true
)
DELETE FROM site_plan_sections
WHERE plan_id IN (SELECT plan_id FROM cur)
  AND page_name = 'product-detail'
  AND component_name IN ('product-hero', 'product-specs');

-- Renumber the surviving call-to-action to ordering 1.
WITH cur AS (
    SELECT id AS plan_id FROM site_plans
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true
)
UPDATE site_plan_sections
SET ordering = 1
WHERE plan_id IN (SELECT plan_id FROM cur)
  AND page_name = 'product-detail'
  AND component_name = 'call-to-action';

-- Insert gripper-spec-sheet at ordering 0.
WITH cur AS (
    SELECT id AS plan_id FROM site_plans
    WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true
)
INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
SELECT plan_id, 'product-detail', 0, 'gripper-spec-sheet' FROM cur;

-- Re-apply the page_components swap (the prior rebuild resurrected the old
-- components from the stale table before this fix).
DELETE FROM page_components
WHERE page_id = '6e43ae75-487f-40c0-960b-4b3310f516ea'
  AND slot_name IN ('product-hero', 'product-specs');

UPDATE page_components SET position = 2
WHERE page_id = '6e43ae75-487f-40c0-960b-4b3310f516ea' AND slot_name = 'call-to-action';

INSERT INTO page_components (page_id, component_id, slot_name, position, build_status)
SELECT '6e43ae75-487f-40c0-960b-4b3310f516ea',
       (SELECT id FROM content_components WHERE name = 'gripper-spec-sheet'),
       'gripper-spec-sheet', 1, 'pending'
WHERE NOT EXISTS (
    SELECT 1 FROM page_components
    WHERE page_id = '6e43ae75-487f-40c0-960b-4b3310f516ea' AND slot_name = 'gripper-spec-sheet'
);

-- Realign pages.sections (materialised cache) to match.
UPDATE pages
SET sections = '["gripper-spec-sheet","call-to-action"]'::jsonb
WHERE id = '6e43ae75-487f-40c0-960b-4b3310f516ea';

COMMIT;
