-- W9 step 6 (read-only): watch the w9_05 rebuild pair drain, then confirm the renders
-- localise. Repeat until both items are complete AND both pages show t / f / f.
SELECT item_key, status, error IS NOT NULL AS has_error
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND created_by = 'w9_localise_rebuild'
ORDER BY item_key;

SELECT p.name, pc.updated_at,
       (pc.rendered_html LIKE '%/assets/images/illustration-%') AS local_illustration,  -- expect t
       (pc.rendered_html LIKE '%X-Amz-Expires%')                AS presigned_left,      -- expect f
       (pc.rendered_html LIKE '%personae-prod-uk001-images%')   AS b2_left              -- expect f
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name IN ('index','tools') AND pc.slot_name = 'brief-explanation'
ORDER BY p.name;
