-- W8 step 5 (read-only): provenance with the filter WIDENED (4.2's needs_page-only
-- scope was too narrow to support "nothing ran"). All item types since 15:00:
SELECT item_type, item_key, created_by, handler_agent, status, created_at
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND created_at > '2026-07-03 15:00'
ORDER BY created_at;
