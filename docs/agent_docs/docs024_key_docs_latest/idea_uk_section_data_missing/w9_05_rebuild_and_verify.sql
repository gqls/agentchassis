-- W9 step 5: rebuild the two pages that render affected assets (index + tools —
-- the illustrations; hero renders are shadowed by the legacy site-level hero_url,
-- so other pages are render-neutral to the flip). Then verify the renders localise.
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             spec, priority, handler_agent, status, created_by, item_key, page_id)
SELECT p.site_id, 'manual', 'build', 'needs_page', 'medium',
       'Rebuild after asset URL localisation: ' || p.name,
       jsonb_build_object('reason', 'asset_url_localisation', 'page_name', p.name),
       99, 'page-build-handler', 'triaged', 'w9_localise_rebuild',
       'page_rerender:' || p.name, p.id
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
  AND p.name IN ('index', 'tools')
  AND NOT EXISTS (SELECT 1 FROM site_work_items w
    WHERE w.site_id = p.site_id AND w.item_key = 'page_rerender:' || p.name
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved'))
RETURNING item_key, status;

-- After the pair completes (repeat):
SELECT p.name, pc.updated_at,
       (pc.rendered_html LIKE '%/assets/images/illustration-%') AS local_illustration,  -- expect t
       (pc.rendered_html LIKE '%X-Amz-Expires%')                AS presigned_left,      -- expect f
       (pc.rendered_html LIKE '%personae-prod-uk001-images%')   AS b2_left              -- expect f
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name IN ('index','tools') AND pc.slot_name = 'brief-explanation'
ORDER BY p.name;
