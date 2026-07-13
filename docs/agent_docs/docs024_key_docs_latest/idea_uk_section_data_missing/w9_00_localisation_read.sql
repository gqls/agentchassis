-- W9 step 0 (read-only): the localisation mechanism — how do heroes get LOCAL paths?

-- 0.1 The hero component's hero_url field (its declared source is the mechanism's name):
SELECT function,
       substr(input_schema::text,
              greatest(position('hero_url' in input_schema::text) - 40, 1), 360) AS hero_url_field
FROM content_components
WHERE function = 'hero' AND is_active = true AND forked_from IS NULL;

-- 0.2 image-build-handler's workflow around its deploy step (what already writes files
--     into the site repo, and what it records where):
SELECT type, version,
       substr(default_config::text,
              greatest(position('deploy' in default_config::text) - 150, 1), 700) AS around_deploy
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true AND deleted_at IS NULL;
