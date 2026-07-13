-- W4b step 3 (read-only): the exact shapes for the trigger item — no guessing.
-- 3.1 The full rerender-pages v6 workflow (how spec.function / spec.component_id /
--     spec.refresh_site_components are consumed; the input_contract; the per-page
--     item creation step's filter, if any):
SELECT jsonb_pretty(default_config) AS rerender_pages_config
FROM agent_definitions
WHERE type = 'rerender-pages' AND version = 6 AND deleted_at IS NULL;

-- 3.2 Full metadata of the two most recent real needs_rerender items (every column the
--     insert must populate, taken from reality):
SELECT pipeline, source, item_type, severity, priority, handler_agent, status,
       created_by, item_key, spec
FROM site_work_items
WHERE item_type = 'needs_rerender'
ORDER BY created_at DESC
LIMIT 2;
