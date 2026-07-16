-- SQL_2026-07-16_asset_deployer_content_card_mode.sql
--
-- Phase I3.2: add a 'content_card' branch to asset-deployer's mode chain.
-- Routes after this migration:
--   mode == 'brand_head'   → derive_head_assets     (favicon + OG)
--   mode == 'sprite_css'   → emit_sprite_css        (sprites.css)
--   mode == 'content_card' → derive_card_asset_step (entity card crop)  ← new
--   otherwise              → deploy_asset           (default image deploy)
--
-- derive_card_asset needs the S3 client (reads the hero bytes) + Kafka
-- producer (git commit) — asset-deployer is the storage-enabled home for
-- exactly this shape (cf. brand_head). Reusable fleet-wide via a
-- needs_content_image work item (spec.mode='content_card', emitted by the
-- content_image_missing discovery check).
--
-- Explicit Strategy-0 input paths on the new step (the I2.1 lesson:
-- ExtractActionInputs' recursive search once resolved purpose→hero; explicit
-- input_data.* paths prevent it).
--
-- SAFE TO RUN PRE-DEPLOY: nothing emits mode='content_card' items until the
-- content_image_missing check is registered AND the new binary is out; if one
-- were hand-inserted early, the workflow fails visibly on the unknown action.
--
-- Backup + snapshot. Idempotent in effect.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS agent_def_asset_deployer_backup_20260716_content_card AS
SELECT * FROM agent_definitions
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

BEGIN;

SELECT snapshot_agent('asset-deployer', 'add content_card mode branch (derive_card_asset)');

-- Rewire check_sprite_mode's else path through the new conditional.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,check_sprite_mode,config,else_step}',
        '"check_card_mode"'::jsonb)
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,check_card_mode}',
        '{
           "action": "conditional",
           "config": {
             "condition": "input_data.spec.mode == \"content_card\" OR input_data.mode == \"content_card\"",
             "then_step": "derive_card_asset_step",
             "else_step": "deploy_asset"
           },
           "description": "Route content-card derivations away from the default image deploy"
         }'::jsonb)
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,derive_card_asset_step}',
        '{
           "action": "derive_card_asset",
           "config": {
             "site_id":     "input_data.site_id",
             "domain":      "input_data.domain",
             "entity_id":   "input_data.spec.entity_id",
             "entity_type": "input_data.spec.entity_type",
             "page_name":   "input_data.spec.page_name"
           },
           "next_step": "complete",
           "description": "Derive the entity card crop from its hero and commit"
         }'::jsonb),
    updated_at = now()
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

DO $verify$
BEGIN
    IF (SELECT default_config #>> '{workflow,steps,check_sprite_mode,config,else_step}'
        FROM agent_definitions WHERE type='asset-deployer' AND is_active=true
          AND (is_snapshot IS NULL OR is_snapshot=false)) <> 'check_card_mode' THEN
        RAISE EXCEPTION 'check_sprite_mode else_step not rewired';
    END IF;
    IF (SELECT default_config #>> '{workflow,steps,check_card_mode,config,else_step}'
        FROM agent_definitions WHERE type='asset-deployer' AND is_active=true
          AND (is_snapshot IS NULL OR is_snapshot=false)) <> 'deploy_asset' THEN
        RAISE EXCEPTION 'check_card_mode else_step must fall through to deploy_asset';
    END IF;
    IF NOT (SELECT (default_config #> '{workflow,steps}') ? 'derive_card_asset_step'
            FROM agent_definitions WHERE type='asset-deployer' AND is_active=true
              AND (is_snapshot IS NULL OR is_snapshot=false)) THEN
        RAISE EXCEPTION 'derive_card_asset_step missing';
    END IF;
    RAISE NOTICE 'asset-deployer content_card mode added';
END
$verify$;

COMMIT;
