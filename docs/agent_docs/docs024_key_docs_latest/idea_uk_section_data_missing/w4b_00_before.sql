-- W4b step 0 (read-only): chrome repoint prep. Schema first, then ids, rows, and the
-- operational trigger for render_site_components.
\d site_components

-- 0.1 The active fixed chrome components (the repoint targets):
SELECT id, function, is_active, updated_at
FROM content_components
WHERE function IN ('site-header','site-footer') AND is_active = true AND forked_from IS NULL;

-- 0.2 idea.uk's current site_components rows (what gets repointed; head stays pinned):
SELECT sc.slot_name, sc.component_id, cc.function, cc.is_active,
       length(sc.rendered_html) AS rendered_len, sc.build_status, sc.updated_at
FROM site_components sc
LEFT JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
ORDER BY sc.slot_name;

-- 0.3 Which agent workflow invokes render_site_components (the operational trigger for
--     the force_rerender run):
SELECT type, version, is_active
FROM agent_definitions
WHERE default_config::text LIKE '%render_site_components%' AND deleted_at IS NULL
ORDER BY type, version;
