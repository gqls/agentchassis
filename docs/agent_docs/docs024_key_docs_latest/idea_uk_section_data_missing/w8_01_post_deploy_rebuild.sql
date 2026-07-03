-- W8: post-deploy rebuild of the two pages (run ONLY after slice 1 deploys).
-- Same real needs_page shape as W6; fresh created_by; dedup mirrors idx_swi_dedup.
INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary,
  spec, priority, handler_agent, status, created_by, item_key, page_id
)
SELECT p.site_id, 'manual', 'build', 'needs_page', 'medium',
       'Post-deploy rebuild (plan_sections slice 1): ' || p.name,
       jsonb_build_object('reason', 'plan_sections_skip_field_deploy', 'page_name', p.name),
       99, 'page-build-handler', 'triaged', 'w8_post_deploy_rebuild',
       'page_rerender:' || p.name, p.id
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND p.name IN ('index', 'tools')
  AND NOT EXISTS (
    SELECT 1 FROM site_work_items w
    WHERE w.site_id = p.site_id
      AND w.item_key = 'page_rerender:' || p.name
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved')
  )
RETURNING item_key, status;

-- Verify afterwards (repeat):
SELECT p.name, pc.slot_name,
       (pc.rendered_html LIKE '%data-component="brief-explanation"%') AS section_back,
       (pc.rendered_html LIKE '%illustration_home%' OR pc.rendered_html LIKE '%illustration_tools%') AS has_image
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name IN ('index','tools') AND pc.slot_name = 'brief-explanation'
ORDER BY p.name;

SELECT item_key, status FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND item_key LIKE 'section_data_%brief-explanation%';   -- expect both → complete (self-closed)
