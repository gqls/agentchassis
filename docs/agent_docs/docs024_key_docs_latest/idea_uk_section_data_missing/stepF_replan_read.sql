-- Read the exact needs_site_plan trigger shape + build-site-planner's workflow, so a
-- manual re-plan emit for idea.uk matches the tested route. Schema-first where it touches
-- tables. The planner re-run is safe by design: normaliseRealisedToPlanPage unions the
-- realised pages (carrying their sections) with the LLM proposal, so the 6 built pages
-- keep their composition while the 3 catalogued pages get composed.

-- F.1 build-site-planner definition: workflow steps + the item type it consumes:
SELECT type, version, is_active,
       jsonb_object_keys(default_config->'workflow'->'steps') AS step
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true AND deleted_at IS NULL;

-- F.2 Its load_existing_pages / write_site_plan step configs (how it reads realised
--     pages and what input_mapping write_site_plan expects):
SELECT left(default_config::text, 2000) AS config_head
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active = true AND deleted_at IS NULL;

-- F.3 The historical needs_site_plan item for idea.uk (the shape a manual emit copies) —
--     schema first:
\d site_work_items
SELECT item_type, source, pipeline, severity, handler_agent, status,
       left(spec::text, 300) AS spec_head, created_by, item_key, created_at
FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE s.domain = 'idea.uk'
  AND (item_type = 'needs_site_plan' OR handler_agent = 'build-site-planner')
ORDER BY created_at;

-- F.4 Confirm the 3 target pages are in the pages catalogue with nav intent (so the
--     planner's load_existing_pages will surface them for the union):
SELECT name, page_type, build_status, status, in_header, in_footer, nav_order, nav_label
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND name IN ('news-index','guides-index','tool-audience-check')
ORDER BY name;
