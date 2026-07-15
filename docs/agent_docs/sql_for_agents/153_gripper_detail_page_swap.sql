-- 153: swap category-wrong e-commerce components for gripper-spec-sheet
--
-- Context: robot-hands.com's gripper-detail and product-detail pages carried
-- Add-to-Cart/Buy-Now furniture (product-hero, product-specs, product-
-- details, product-card-with-cta) regardless of data — category-wrong for a
-- spec/comparison site. product-card-with-cta additionally had an LLM
-- apology about missing query.affiliate_products persisted as its content
-- (handoff §3d) — removed as part of the same cleanup, same root cause.
-- system-stats, features, generic-text-block, call-to-action are untouched
-- (not e-commerce furniture; generic-text-block's missing content is a
-- separate, already-flagged required_fields_missing item).
--
-- Requires 151 (component) and 152 (product rows) applied first.
--
-- Two different section-list sources, per page-build-handler's
-- load_page_sections_from_spec (site_specs.site_plan is authoritative,
-- pages.sections is the fallback):
--   - gripper-detail IS in the current site_plan (role entity-page,
--     sections: null) — both site_plan AND pages.sections are updated so a
--     future rebuild doesn't see stale/conflicting section lists.
--   - product-detail is NOT in the current site_plan at all (a legacy page
--     outside the current planning system) — only pages.sections applies.
--
-- Verify after applying:
--   SELECT pc.slot_name, cc.function FROM page_components pc
--   JOIN content_components cc ON cc.id=pc.component_id
--   WHERE pc.page_id IN ('11364960-d16d-4a3e-bb52-98916fa4edbe','6e43ae75-487f-40c0-960b-4b3310f516ea')
--   ORDER BY pc.page_id, pc.position;
--   -- expect no product-hero/product-specs/product-details/product-card-with-cta rows

BEGIN;

-- --- gripper-detail: remove the e-commerce components ---
DELETE FROM page_components
WHERE page_id = '11364960-d16d-4a3e-bb52-98916fa4edbe'
  AND slot_name IN ('product-hero', 'product-specs', 'product-details', 'product-card-with-cta');

-- --- gripper-detail: renumber survivors, insert gripper-spec-sheet at position 1 ---
UPDATE page_components SET position = 2 WHERE page_id = '11364960-d16d-4a3e-bb52-98916fa4edbe' AND slot_name = 'system-stats';
UPDATE page_components SET position = 3 WHERE page_id = '11364960-d16d-4a3e-bb52-98916fa4edbe' AND slot_name = 'features';
UPDATE page_components SET position = 4 WHERE page_id = '11364960-d16d-4a3e-bb52-98916fa4edbe' AND slot_name = 'generic-text-block';
UPDATE page_components SET position = 5 WHERE page_id = '11364960-d16d-4a3e-bb52-98916fa4edbe' AND slot_name = 'call-to-action';

INSERT INTO page_components (page_id, component_id, slot_name, position, build_status)
VALUES (
    '11364960-d16d-4a3e-bb52-98916fa4edbe',
    (SELECT id FROM content_components WHERE name = 'gripper-spec-sheet'),
    'gripper-spec-sheet',
    1,
    'pending'
);

-- --- gripper-detail: pages.sections (fallback) ---
UPDATE pages
SET sections = '["gripper-spec-sheet","system-stats","features","generic-text-block","call-to-action"]'::jsonb
WHERE id = '11364960-d16d-4a3e-bb52-98916fa4edbe';

-- --- gripper-detail: site_plan (authoritative) — locate its array index and set sections there ---
UPDATE site_specs
SET data = jsonb_set(
    data,
    ARRAY['pages', (
        SELECT (ordinality - 1)::text
        FROM jsonb_array_elements(data->'pages') WITH ORDINALITY AS t(elem, ordinality)
        WHERE elem->>'name' = 'gripper-detail'
    ), 'sections'],
    '["gripper-spec-sheet","system-stats","features","generic-text-block","call-to-action"]'::jsonb
)
WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
  AND aspect = 'site_plan'
  AND is_current = true;

-- --- product-detail: remove the e-commerce components ---
DELETE FROM page_components
WHERE page_id = '6e43ae75-487f-40c0-960b-4b3310f516ea'
  AND slot_name IN ('product-hero', 'product-specs');

UPDATE page_components SET position = 2 WHERE page_id = '6e43ae75-487f-40c0-960b-4b3310f516ea' AND slot_name = 'call-to-action';

INSERT INTO page_components (page_id, component_id, slot_name, position, build_status)
VALUES (
    '6e43ae75-487f-40c0-960b-4b3310f516ea',
    (SELECT id FROM content_components WHERE name = 'gripper-spec-sheet'),
    'gripper-spec-sheet',
    1,
    'pending'
);

-- --- product-detail: pages.sections (its only section-list source) ---
UPDATE pages
SET sections = '["gripper-spec-sheet","call-to-action"]'::jsonb
WHERE id = '6e43ae75-487f-40c0-960b-4b3310f516ea';

COMMIT;
