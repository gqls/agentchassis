-- W4b step 2 (read-only): choose the forced re-render trigger.
-- 2.1 Each agent's render_site_components step config (does it pass force_rerender: true,
--     and how does site_id arrive?):
SELECT type, version,
       substr(default_config::text,
              greatest(position('render_site_components' in default_config::text) - 150, 1),
              520) AS around_render_step
FROM agent_definitions
WHERE default_config::text LIKE '%render_site_components%'
  AND deleted_at IS NULL AND is_active = true
ORDER BY type;

-- 2.2 The shape of real rerender work items (so ours is crafted exactly, not guessed):
SELECT item_type, handler_agent, status, item_key,
       left(spec::text, 300) AS spec_head, created_at
FROM site_work_items
WHERE handler_agent IN ('rerender-pages', 'rerender-site')
ORDER BY created_at DESC
LIMIT 5;
