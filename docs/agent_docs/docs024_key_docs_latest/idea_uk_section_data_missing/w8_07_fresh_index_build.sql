-- W8 step 7: the decisive forward test — ONE fresh index BUILD, no recovery or rerender
-- in flight, then read both the render AND the stored data. Schema first for the
-- content_data column (page_components was never \d'd in this thread).
\d page_components

-- 7.1 The build:
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             spec, priority, handler_agent, status, created_by, item_key, page_id)
SELECT p.site_id, 'manual', 'build', 'needs_page', 'medium',
       'Edit-B forward probe: fresh index build',
       jsonb_build_object('reason', 'editb_forward_probe', 'page_name', p.name),
       99, 'page-build-handler', 'triaged', 'w8_editb_probe',
       'page_rerender:' || p.name, p.id
FROM pages p
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk') AND p.name = 'index'
  AND NOT EXISTS (SELECT 1 FROM site_work_items w
    WHERE w.site_id = p.site_id AND w.item_key = 'page_rerender:' || p.name
      AND w.status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved'))
RETURNING item_key, status;

-- 7.2 After it completes (repeat until updated_at moves past the insert time):
SELECT p.name, pc.updated_at,
       (pc.rendered_html LIKE '%illustration_%')            AS has_image,
       pc.content_data ->> 'illustration_url'               AS stored_illustration_url
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name = 'index' AND pc.slot_name = 'brief-explanation';
-- Interpretation:
--   has_image t                          → resolver + template both fine; earlier f's were
--                                          rerender/recovery output over stale data; CLOSE.
--   has_image f + stored url PRESENT     → resolver fine, template/gate at render time — read the render.
--   has_image f + stored url ABSENT      → resolver missed on a clean build despite verified
--                                          code — next: temporary Info log in Edit B's loop.
