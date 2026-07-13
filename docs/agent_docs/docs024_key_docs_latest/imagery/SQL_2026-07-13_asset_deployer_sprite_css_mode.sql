-- SQL_2026-07-13_asset_deployer_sprite_css_mode.sql
--
-- Phase I2.2: add a 'sprite_css' branch to asset-deployer's check_mode
-- (added by the brand_head migration). Routes:
--   mode == 'brand_head' → derive_head_assets   (favicon + OG)
--   mode == 'sprite_css' → emit_sprite_css       (sprites.css)   ← new
--   otherwise            → deploy_asset          (default image deploy)
--
-- emit_sprite_css needs only DB + Kafka producer (no S3), but asset-deployer
-- is the natural "commit to the site repo" home and already carries the mode
-- switch, so it lives here alongside brand_head. Reusable fleet-wide via a
-- needs_sprite_css work item (spec.mode='sprite_css').
--
-- Backup + snapshot. Idempotent in effect.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS agent_def_asset_deployer_backup_20260713_sprite_css AS
SELECT * FROM agent_definitions
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

BEGIN;

SELECT snapshot_agent('asset-deployer', 'add sprite_css mode branch (emit_sprite_css)');

-- Widen check_mode: sprite_css routes to emit_sprite_css_step; brand_head
-- keeps its route; else stays deploy_asset. A conditional has one then/else,
-- so add a second conditional (check_sprite_mode) reached from check_mode's
-- else path.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,check_mode,config}',
        '{
           "condition": "input_data.spec.mode == \"brand_head\" OR input_data.mode == \"brand_head\"",
           "then_step": "derive_head_assets",
           "else_step": "check_sprite_mode"
         }'::jsonb)
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,check_sprite_mode}',
        '{
           "action": "conditional",
           "config": {
             "condition": "input_data.spec.mode == \"sprite_css\" OR input_data.mode == \"sprite_css\"",
             "then_step": "emit_sprite_css_step",
             "else_step": "deploy_asset"
           },
           "description": "Route sprite_css derivations away from the default image deploy"
         }'::jsonb)
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,emit_sprite_css_step}',
        '{
           "action": "emit_sprite_css",
           "config": {
             "site_id": "input_data.site_id",
             "domain": "input_data.domain"
           },
           "next_step": "complete",
           "description": "Generate + commit sprites.css from the verified sprite-sheet plan"
         }'::jsonb),
    updated_at = now()
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

DO $verify$
BEGIN
    IF (SELECT default_config #>> '{workflow,steps,check_mode,config,else_step}'
        FROM agent_definitions WHERE type='asset-deployer' AND is_active=true
          AND (is_snapshot IS NULL OR is_snapshot=false)) <> 'check_sprite_mode' THEN
        RAISE EXCEPTION 'check_mode else_step not rewired';
    END IF;
    IF NOT (SELECT (default_config #> '{workflow,steps}') ? 'emit_sprite_css_step'
            FROM agent_definitions WHERE type='asset-deployer' AND is_active=true
              AND (is_snapshot IS NULL OR is_snapshot=false)) THEN
        RAISE EXCEPTION 'emit_sprite_css_step missing';
    END IF;
    RAISE NOTICE 'asset-deployer sprite_css mode added';
END
$verify$;

COMMIT;
