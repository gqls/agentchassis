-- diagnostic_hub_sections.sql  (READ-ONLY — no writes)
-- ----------------------------------------------------------------------------
-- THREAD B — the empty guides-index hub. List sections come from the plan (no
-- deterministic code path attaches them), so this shows how the WORKING
-- tools-index is structured vs the empty guides-index, in both the realised
-- pages and the current plan. With this I can mirror tools-index onto
-- guides-index (add the right list section, in the right place, with the right
-- component name) instead of guessing.
-- ----------------------------------------------------------------------------

-- (1) Realised pages.sections for the hubs + homepage.
SELECT name, page_type, build_status,
       jsonb_typeof(sections) AS sections_type,
       sections
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND name IN ('index', 'tools-index', 'games-index', 'guides-index')
ORDER BY name;

-- (2) Current-plan sections per hub page (ordering + component, NULL if none).
SELECT spp.name AS page,
       sps.ordering,
       sps.component_name
FROM site_plan_pages spp
JOIN site_plans sp ON sp.id = spp.plan_id AND sp.is_current = true
LEFT JOIN site_plan_sections sps
       ON sps.plan_id = spp.plan_id AND sps.page_name = spp.name
WHERE sp.site_id = (SELECT id FROM sites WHERE domain = 'gamesdesign.co.uk')
  AND spp.name IN ('index', 'tools-index', 'games-index', 'guides-index')
ORDER BY spp.name, sps.ordering;

-- (3) Confirm the guide-list component's identity + tier (the component_name
--     we would reference for guides-index, and that it is query-sourced).
SELECT name, component_level, render_mode, is_active,
       (input_schema::text LIKE '%query.pages_where_type:guide%') AS sources_guides
FROM content_components
WHERE name LIKE 'guide-list%' OR name LIKE 'tool-list%'
ORDER BY name;
