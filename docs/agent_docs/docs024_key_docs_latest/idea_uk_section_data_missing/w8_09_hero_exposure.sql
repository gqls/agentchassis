-- W8 step 9 (read-only): does the presigned-expiry issue already affect HEROES on the
-- current renders? (8.2 showed hero_home's assets.url in the same B2 form as the
-- illustrations; if presigned it expired 06-28, and W6's rebuild re-resolved heroes
-- through the same a.url join.)

-- 9.1 The current hero renders: local paths, B2 URLs, presigned markers:
SELECT p.name, pc.slot_name,
       (pc.rendered_html LIKE '%/assets/images/%')            AS local_path,
       (pc.rendered_html LIKE '%personae-prod-uk001-images%') AS b2_url,
       (pc.rendered_html LIKE '%X-Amz-Expires%')              AS presigned,
       pc.updated_at
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND pc.slot_name IN ('hero', 'hero-about', 'hero-contact')
ORDER BY p.name;

-- 9.2 Are the stored hero URLs presigned (and therefore expired since ~06-28)?
SELECT asset_key,
       (url LIKE '%X-Amz-Expires%') AS presigned,
       substr(url, greatest(position('X-Amz-Date=' in url), 1), 26) AS signed_at_or_head,
       created_at
FROM assets
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND asset_key IN ('hero_home', 'hero_about', 'hero_contact', 'hero_report', 'hero_tools')
ORDER BY asset_key;
