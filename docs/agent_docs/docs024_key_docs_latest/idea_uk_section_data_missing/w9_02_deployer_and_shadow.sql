-- W9 step 2 (read-only): the asset-deployer's write behaviour + the hero shadow check.

-- 2.1 The deployer agent (exact type name first — my 'asset-deployer' is a guess):
SELECT type, version, is_active
FROM agent_definitions
WHERE type LIKE '%asset%' AND deleted_at IS NULL
ORDER BY type, version;

SELECT jsonb_object_keys(default_config -> 'workflow' -> 'steps') AS step_name
FROM agent_definitions
WHERE type = 'asset-deployer' AND is_active = true AND deleted_at IS NULL;

SELECT substr(default_config::text,
              greatest(position('commit' in default_config::text) - 120, 1), 700) AS around_commit
FROM agent_definitions
WHERE type = 'asset-deployer' AND is_active = true AND deleted_at IS NULL;

-- 2.2 The hero's actual site_assets.hero field (name + shape — background_image expected):
SELECT function,
       substr(input_schema::text,
              greatest(position('site_assets.hero' in input_schema::text) - 120, 1), 260) AS hero_source_field
FROM content_components
WHERE function = 'hero' AND is_active = true AND forked_from IS NULL;

-- 2.3 The shadow: does index's stored hero content_data hold BOTH keys — a presigned
--     background_image (resolved, shadowed) and the legacy hero_url (rendered)?
SELECT p.name,
       (pc.content_data::text LIKE '%background_image%') AS has_bg_image_key,
       (pc.content_data::text LIKE '%X-Amz-Expires%')    AS presigned_in_data,
       (pc.content_data::text LIKE '%hero_url%')         AS has_hero_url_key
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND p.name = 'index' AND pc.slot_name = 'hero';
