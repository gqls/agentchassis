-- R9 (2026-07-24) — give the gripper-catalog page the real, query-backed gripper
-- grid. Completes the browsability residual: gripper-detail already renders all
-- 10 products via gripper-spec-sheet -> query.products:gripper (kept in sync on
-- every render); the catalog page itself was static prose with zero product
-- names. Reuses the EXACT component proven live on gripper-detail
-- (585e3c5c-1f0d-4d9c-81c1-a47d102b3c5f) rather than product-grid, which is an
-- e-commerce card (price/rating/badge/image) our grippers don't have — empty
-- price scaffolding is a fabrication invitation, spec cards are the honest fit.
--
-- Mechanics: page_components has NO unique (page_id, position) constraint, but
-- deterministic ordering matters, so later components shift down one. The
-- pages.sections array is mirrored (040 lesson: keep rows and section lists
-- consistent). The products array is resolved fresh from the products table on
-- render (query.products:gripper, queryresolve.go resolveProducts), so content
-- here carries only the headline; nothing hand-copied that could drift.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

-- 1. Shift info-card-grid (3->4) and call-to-action (4->5)
UPDATE page_components pc SET position = position + 1, updated_at = now()
 WHERE pc.page_id = (SELECT id FROM pages WHERE site_id = :'site' AND name='gripper-catalog')
   AND pc.position >= 3;

-- 2. Insert the spec-sheet grid at position 3
INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, build_status)
SELECT p.id, '585e3c5c-1f0d-4d9c-81c1-a47d102b3c5f', 3, 'gripper-spec-sheet',
       '{"headline": "The Gripper Index — Complete and Sourced"}'::jsonb,
       'pending'
FROM pages p WHERE p.site_id = :'site' AND p.name = 'gripper-catalog';

-- 3. Mirror pages.sections
UPDATE pages SET sections = '["hero","generic-text-block","gripper-spec-sheet","info-card-grid","call-to-action"]'::jsonb,
       updated_at = now()
 WHERE site_id = :'site' AND name = 'gripper-catalog';

\echo '--- components after insert ---'
SELECT pc.position, pc.slot_name, pc.build_status
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id = :'site' AND p.name='gripper-catalog' ORDER BY pc.position;

-- 4. Re-render (proven fields: handler_agent + item_key + reason that reaches
--    the real rerender_sections branch)
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT
  p.site_id, 'robot-hands-r9-catalog-grid', 'page_rerender', 'medium',
  'Rerender gripper-catalog — query-backed gripper spec grid added (R9)',
  'triaged', 'session-2026-07-24-robot-hands', 'build',
  20, now(), 'page-rerender',
  'page_rerender_' || p.name || '_r9grid_' || p.site_id::text,
  jsonb_build_object('domain','robot-hands.com','reason','cta_links_stale',
                     'page_id',p.id,'page_name',p.name,'filename',ltrim(p.url,'/'))
FROM pages p
WHERE p.site_id = :'site' AND p.name = 'gripper-catalog';

\echo '--- queued ---'
SELECT status, handler_agent, count(*) FROM site_work_items
WHERE source='robot-hands-r9-catalog-grid' GROUP BY 1,2;

COMMIT;
