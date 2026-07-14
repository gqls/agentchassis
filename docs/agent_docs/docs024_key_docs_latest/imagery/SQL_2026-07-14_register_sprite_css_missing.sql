-- I2.4 — register the sprite_css_missing discovery check on design-discovery-agent.
--
-- The check (platform/orchestration/actions/discovery_checks/check_sprite_css_missing.go)
-- emits needs_sprite_css when a site has a VERIFIED, deployed sprite sheet but
-- /assets/css/sprites.css was never emitted, or was emitted from a grid that no
-- longer matches the sheet (cell names re-verified / sheet regenerated).
-- asset-deployer's sprite_css mode handles the item.
--
-- The other half of the I2.4 gap ("sprite_sheet planned but no asset") needs no
-- new check: unfulfilled_imagery_plan already emits needs_imagery for ANY
-- unfulfilled plan row regardless of kind.
--
-- NOTE: the Go check must be DEPLOYED before this runs, else the check name is
-- enabled in config but unregistered — RunDiscoveryChecksAction logs it as
-- unknown and skips it (harmless, but it does nothing).
\set ON_ERROR_STOP on
BEGIN;

-- Backup (house practice: every migration has backup + verify).
CREATE TABLE IF NOT EXISTS agent_def_design_discovery_backup_20260714_sprite_css AS
SELECT * FROM agent_definitions WHERE type = 'design-discovery-agent';

-- Append the check to run_checks' enabled list, idempotently.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,run_checks,config,checks}',
      (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
        || '["sprite_css_missing"]'::jsonb,
      false
    ),
    updated_at = now()
WHERE type = 'design-discovery-agent'
  AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks'
           @> '["sprite_css_missing"]'::jsonb);

-- Verify: the check is present exactly once.
SELECT
  jsonb_array_length(default_config->'workflow'->'steps'->'run_checks'->'config'->'checks') AS total_checks,
  (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
    @> '["sprite_css_missing"]'::jsonb AS sprite_css_missing_enabled
FROM agent_definitions
WHERE type = 'design-discovery-agent';

COMMIT;

-- Rollback:
-- UPDATE agent_definitions SET default_config = (
--   SELECT default_config FROM agent_def_design_discovery_backup_20260714_sprite_css
-- ) WHERE type = 'design-discovery-agent';
