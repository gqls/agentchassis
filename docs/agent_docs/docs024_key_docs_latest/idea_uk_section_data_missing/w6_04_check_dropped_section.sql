-- W6 step 4 (read-only): why is brief-explanation absent from the rebuilt index?
-- Three candidate mechanisms — this read distinguishes them before any response.
-- 4.1 Does the page itself still ask for it? (pages.sections is the names array.)
SELECT name, sections
FROM pages
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND name IN ('index','tools');

-- 4.2 The full rebuilt component sets for tools + contact (dropped there too?):
SELECT p.name, pc.slot_name, cc.function
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name IN ('tools','contact')
ORDER BY p.name, pc.slot_name;

-- 4.3 Anything the rebuild escalated or created during its window (19:11 onwards)?
SELECT item_type, handler_agent, status, item_key, left(summary, 80) AS summary, created_at
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND created_at > '2026-07-02 19:11'
  AND created_by <> 'w6_scheme_rebuild'
ORDER BY created_at;
