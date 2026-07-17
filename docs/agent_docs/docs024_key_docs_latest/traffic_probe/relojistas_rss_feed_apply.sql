-- ============================================================================
-- relojistas_rss_feed_apply.sql — P4 activation (apply ONLY AFTER the chassis
-- image carrying render_rss_feed is live — image first, then seeds: a workflow
-- naming an unregistered action fails at runtime).
--
-- Two parts:
--   1. content-feed-orchestrator workflow: commit_news -> render_rss_xml ->
--      check_has_rss -> commit_rss -> complete. Fleet-safe: the action returns
--      item_count=0 for sites without deploy_config.rss_feed.enabled, and
--      check_has_rss routes 0 -> complete (mirrors check_has_news).
--   2. relojistas.com: enable the per-site rss_feed gate with Spanish channel
--      metadata; atom:self = the LEGACY vBulletin feed URL so surviving
--      subscribers keep their original address.
-- ============================================================================

-- 1a. New steps
UPDATE agent_definitions SET default_config =
  jsonb_set(jsonb_set(jsonb_set(default_config,
    '{workflow,steps,render_rss_xml}',
    '{"action":"render_rss_feed","config":{"site_id":"input_data.site_id"},
      "description":"Render outbound RSS 2.0 feed.xml (per-site gated by deploy_config.rss_feed)",
      "output_field":"rss_render_result","next_step":"check_has_rss"}'::jsonb),
    '{workflow,steps,check_has_rss}',
    '{"action":"evaluate_condition","config":{"condition_field":"rss_render_result.item_count",
      "conditions":{"0":"complete"},"default":"commit_rss"},
      "description":"Skip RSS commit when disabled for the site or no items"}'::jsonb),
    '{workflow,steps,commit_rss}',
    '{"action":"git_commit","config":{"files_field":"rss_render_result.files",
      "domain_field":"rss_render_result.domain","commit_message":"Update RSS feed"},
      "description":"Commit feed.xml (repo resolved per-site via resolveGitRepoNameDB)",
      "output_field":"rss_commit_result","next_step":"complete"}'::jsonb),
  updated_at = NOW()
WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;

-- 1b. Rewire commit_news to flow into the RSS step instead of complete
UPDATE agent_definitions SET default_config =
  jsonb_set(default_config, '{workflow,steps,commit_news,next_step}', '"render_rss_xml"'),
  updated_at = NOW()
WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL;

-- 1c. Also surface the rss results in the completion output (optional, tidy)
UPDATE agent_definitions SET default_config =
  jsonb_set(default_config, '{workflow,steps,complete,config,output_fields}',
    (default_config->'workflow'->'steps'->'complete'->'config'->'output_fields')
      || '["rss_render_result","rss_commit_result"]'::jsonb),
  updated_at = NOW()
WHERE type = 'content-feed-orchestrator' AND deleted_at IS NULL
  AND NOT (default_config->'workflow'->'steps'->'complete'->'config'->'output_fields' ? 'rss_render_result');

-- CAVEAT: check_has_news routes item_count=0 straight to complete, so a cycle
-- with no news items also skips the RSS refresh. Acceptable: no new items ->
-- feed unchanged.

-- 2. relojistas.com per-site gate + Spanish channel metadata
UPDATE sites SET deploy_config = deploy_config || jsonb_build_object(
  'rss_feed', jsonb_build_object(
    'enabled', true,
    'channel_title', 'Relojistas — Noticias de relojería',
    'channel_link', 'https://relojistas.com/',
    'channel_description', 'Últimas noticias de relojería en español: novedades de marcas, ferias, subastas y alta relojería. Cada entrada enlaza a la fuente original.',
    'language', 'es',
    'self_url', 'https://relojistas.com/external.php?type=RSS2'
  ))
WHERE domain = 'relojistas.com';

-- Verify:
--   SELECT default_config->'workflow'->'steps'->'commit_news'->>'next_step',
--          default_config->'workflow'->'steps' ? 'render_rss_xml'
--   FROM agent_definitions WHERE type='content-feed-orchestrator' AND deleted_at IS NULL;
--   SELECT deploy_config->'rss_feed' FROM sites WHERE domain='relojistas.com';
-- Then trigger a feed pass and expect:
--   rss_render_result.item_count > 0, commit repo=vm-sites, and
--   curl -sS https://relojistas.com/feed.xml | xmllint --noout -   (well-formed)
-- ============================================================================
