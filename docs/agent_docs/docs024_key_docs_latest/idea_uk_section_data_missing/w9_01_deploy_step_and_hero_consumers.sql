-- W9 step 1 (read-only): size the localisation fix.

-- 1.1 image-build-handler's step NAMES (the earlier window caught flag_rebuild/complete;
--     this lists them all so the deploy step's real name is known):
SELECT jsonb_object_keys(default_config -> 'workflow' -> 'steps') AS step_name
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true AND deleted_at IS NULL;

-- 1.2 The deploy-ish step's config (adjust the anchor to the name 1.1 shows if needed):
SELECT substr(default_config::text,
              greatest(position('deploy_asset' in default_config::text) - 60, 1), 700) AS deploy_step
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true AND deleted_at IS NULL;

-- 1.3 Side-observation check: does ANY component consume site_assets.hero (are the
--     per-page hero assets consumed by anything, or did they expire unused)?
SELECT function
FROM content_components
WHERE is_active = true AND forked_from IS NULL
  AND input_schema::text LIKE '%site_assets.hero%'
ORDER BY function;
