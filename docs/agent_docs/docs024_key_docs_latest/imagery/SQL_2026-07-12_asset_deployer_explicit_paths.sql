-- SQL_2026-07-12_asset_deployer_explicit_paths.sql
--
-- Fix for the sprite-sheet deploy landing as 900×900 sprite-sheet-main.JPG
-- ("Deploy hero image"): the asset-deployer child RECEIVED
-- input_data.purpose='sprite_sheet' (verified in initial_request_data), but
-- deploy_image_asset's ExtractActionInputs fell to the aggressive recursive
-- ExtractFields search for `purpose`, which matched a stale value elsewhere
-- in collected_data → empty/hero → the action's "hero" default → hero
-- dimensions (1600×900 → 900² thumbnail) + jpg. The codebase's own remedy is
-- Strategy 0: EXPLICIT dot-path config values are resolved first and win.
--
-- This patch adds explicit input paths to asset-deployer's deploy_asset step
-- (input_fields kept for backward compat; Strategy 0 simply wins when the
-- paths resolve). Workflow-only — effective immediately, no chassis deploy.
--
-- NOTE (recorded, not fixed here): items dispatched to asset-deployer by
-- build-dispatch-loop carry the payload under input_data.spec.* — those
-- explicit paths miss and Strategy 1 search still applies for that shape
-- (the undeployed_asset flow). Same latent hazard; fix when that flow next
-- misbehaves or in a dedicated pass.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS agent_def_asset_deployer_backup_20260712_explicit_paths AS
SELECT * FROM agent_definitions
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

BEGIN;

SELECT snapshot_agent('asset-deployer',
                      'explicit Strategy-0 input paths on deploy_asset (sprite purpose fell to aggressive search)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_asset,config}',
        (default_config #> '{workflow,steps,deploy_asset,config}')
        || '{
              "s3_uri":   "input_data.s3_uri",
              "purpose":  "input_data.purpose",
              "domain":   "input_data.domain",
              "asset_key":"input_data.asset_key"
            }'::jsonb),
    updated_at = now()
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

DO $verify$
DECLARE
    v text;
BEGIN
    SELECT default_config #>> '{workflow,steps,deploy_asset,config,purpose}' INTO v
    FROM agent_definitions
    WHERE type = 'asset-deployer' AND is_active = true
      AND (is_snapshot IS NULL OR is_snapshot = false);
    IF v IS DISTINCT FROM 'input_data.purpose' THEN
        RAISE EXCEPTION 'explicit purpose path not set (got %)', v;
    END IF;
    RAISE NOTICE 'asset-deployer deploy_asset now has explicit Strategy-0 input paths';
END
$verify$;

COMMIT;
