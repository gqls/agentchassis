-- I3.4 — register the content_image_missing discovery check on
-- design-discovery-agent.
--
-- The check (platform/orchestration/actions/discovery_checks/check_content_image_missing.go)
-- emits needs_content_image (handler asset-deployer, mode content_card) for
-- each blog-post page that is LISTED somewhere (a component consumes
-- query.blog_posts) but has no entity-linked card asset, provided the page is
-- derivable (own plan hero, or the site-scope brand hero fallback).
-- derive_card_asset crops the hero to the 800×450 card and writes the
-- entity-linked assets row; the sweep's anti-join then goes quiet — the
-- entity link is the fulfilment stamp.
--
-- NOTE: the Go check must be DEPLOYED before it does anything — an enabled
-- but unregistered check name is warn-and-skip (harmless), so this is safe
-- to apply pre-deploy; it self-activates on the next binary.

\set ON_ERROR_STOP on
BEGIN;

-- Backup (house practice: every migration has backup + verify).
CREATE TABLE IF NOT EXISTS agent_def_design_discovery_backup_20260716_content_image AS
SELECT * FROM agent_definitions WHERE type = 'design-discovery-agent';

-- Append the check to run_checks' enabled list, idempotently.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
        || '["content_image_missing"]'::jsonb,
      false
    ),
    updated_at = now()
WHERE type = 'design-discovery-agent'
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["content_image_missing"]'::jsonb);

-- Verify: the check is present exactly once.
SELECT
  jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS total_checks,
  (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
    @> '["content_image_missing"]'::jsonb AS content_image_missing_enabled
FROM agent_definitions
WHERE type = 'design-discovery-agent';

COMMIT;

-- Rollback:
-- UPDATE agent_definitions SET default_config = (
--   SELECT default_config FROM agent_def_design_discovery_backup_20260716_content_image
-- ) WHERE type = 'design-discovery-agent';
