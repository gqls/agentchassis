-- W7 step 0c (read-only): the three remaining unknowns. PASTE AS TEXT.

-- 0.1 (rerun — not yet received): brief-explanation's input_schema (the declared
--     site_assets.<KEY> path = what spi.key must be) + the illustration markup:
SELECT function,
       jsonb_pretty(input_schema) AS input_schema,
       substr(html_template, greatest(position('illustration' in html_template) - 120, 1), 480) AS illustration_region
FROM content_components
WHERE function = 'brief-explanation' AND is_active = true AND forked_from IS NULL;

-- 0.2 (rerun — not yet received): the escalations' advice columns:
SELECT item_key, jsonb_pretty(spec) AS spec, suggested_action, resolution_path
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND item_key LIKE 'section_data_%brief-explanation%'
ORDER BY created_at;

-- 0.5 CORRECTED (my column-name bug owned: it is component_name, and there is no
--     description column): brief-explanation's ORDINAL on index and tools — the
--     scope_ref convention is page:ordinal (about:2, index:4 in the existing rows):
SELECT page_name, ordering, component_name
FROM site_plan_sections sps
JOIN site_plans sp ON sp.id = sps.plan_id AND sp.is_current = true
WHERE sp.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND page_name IN ('index','tools')
ORDER BY page_name, ordering;

-- 0.6 One real section-illustration item's FULL spec (the shape our items copy):
SELECT jsonb_pretty(spec) AS full_spec
FROM site_work_items
WHERE item_key = 'needs_imagery:section:about:2:illustration_game_master';
