-- W4b step 5 (read-only): verify after the workflow runs (item → complete, then per-page
-- items drain). Run repeatedly until settled.
-- 5.1 The item's lifecycle:
SELECT id, status, updated_at
FROM site_work_items
WHERE item_key = 'chrome_refresh_rerender:' || (SELECT id::text FROM sites WHERE domain='idea.uk')
ORDER BY created_at DESC LIMIT 1;

-- 5.2 The stored chrome — the point of the whole exercise:
SELECT sc.slot_name, cc.function, cc.is_active, sc.build_status, sc.updated_at,
       length(sc.rendered_html)                          AS rendered_len,
       (sc.rendered_html LIKE '%site-header-section%')   AS is_new_header,
       (sc.rendered_html LIKE '%site-footer-section%')   AS is_new_footer,
       (sc.rendered_html LIKE '%color-mix%')             AS footer_has_mix,
       (sc.rendered_html LIKE '%site-header--gradient%') AS still_old_gradient
FROM site_components sc
LEFT JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = (SELECT id FROM sites WHERE domain = 'idea.uk')
ORDER BY sc.slot_name;
-- Expect: header is_new_header=t / still_old_gradient=f; footer is_new_footer=t +
--         footer_has_mix=t; head re-rendered (updated_at bumped) from its pinned template.

-- 5.3 The per-page rerender items draining:
SELECT status, count(*)
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND created_at > now() - interval '2 hours'
GROUP BY status ORDER BY status;
