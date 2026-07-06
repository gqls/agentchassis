-- W8 step 8 (read-only): the CORRECTED probes. The old LIKE '%illustration_%' tested
-- the asset KEY string, but asset objects are UUID-named — the probe was blind.

-- 8.1 What the renders actually contain (both pages):
SELECT p.name, pc.updated_at,
       (pc.rendered_html LIKE '%brief-explanation__image-wrapper%') AS wrapper_rendered,
       (pc.rendered_html LIKE '%personae-prod-uk001-images%')       AS b2_url_in_render,
       substr(pc.rendered_html,
              greatest(position('__image' in pc.rendered_html) - 40, 1), 320) AS image_region
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name IN ('index','tools') AND pc.slot_name = 'brief-explanation'
ORDER BY p.name;

-- 8.2 URL forms by kind — how do heroes avoid the presigned problem? (deployed heroes
--     reference local /assets/images/ paths; these rows show whether assets.url differs
--     by kind or localisation happens downstream):
SELECT asset_key, status, left(url, 95) AS url_head, created_at
FROM assets
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND asset_key IN ('hero_home', 'illustration_home', 'illustration_tools')
ORDER BY asset_key;
