-- SQL_2026-07-11_asset_deployer_brand_head_mode.sql
--
-- Adds a "brand_head" mode branch to asset-deployer so it can run
-- derive_brand_head_assets (favicon + OG card from the logo). asset-deployer
-- is the natural home: it is the storage-enabled agent (spawned with S3 env
-- via spawn_actions.go isStorageEnabledAgent) and already owns git-commit of
-- image assets. A conditional start step routes on input_data.mode:
--   mode == 'brand_head' → derive_head_assets → complete
--   otherwise            → deploy_asset (unchanged default) → complete
--
-- Reusable fleet-wide: any caller that spawns asset-deployer with
-- {site_id, domain, mode:'brand_head'} gets its favicon + og-card committed.
--
-- Backup + snapshot per convention. Idempotent in effect (overwrites the
-- workflow with the same shape).

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS agent_def_asset_deployer_backup_20260711_brand_head AS
SELECT * FROM agent_definitions
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

BEGIN;

SELECT snapshot_agent('asset-deployer',
                      'add brand_head mode branch (derive favicon + OG card)');

-- New start step: route on mode.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,start_step}', to_jsonb('check_mode'::text))
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- check_mode conditional.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,check_mode}',
        '{
           "action": "conditional",
           "config": {
             "condition": "input_data.mode == \"brand_head\"",
             "then_step": "derive_head_assets",
             "else_step": "deploy_asset"
           },
           "description": "Route brand_head derivations away from the default image deploy"
         }'::jsonb)
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- derive_head_assets step.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,derive_head_assets}',
        '{
           "action": "derive_brand_head_assets",
           "config": {
             "site_id": "input_data.site_id",
             "domain": "input_data.domain"
           },
           "next_step": "complete",
           "description": "Derive favicon + OG card from the site logo and commit to git",
           "output_field": "brand_head_result"
         }'::jsonb),
    updated_at = now()
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- Verify
DO $verify$
DECLARE
    v_start text;
    v_has_derive boolean;
BEGIN
    SELECT default_config #>> '{workflow,start_step}',
           (default_config #> '{workflow,steps}') ? 'derive_head_assets'
      INTO v_start, v_has_derive
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active = true
       AND (is_snapshot IS NULL OR is_snapshot = false);
    IF v_start <> 'check_mode' THEN
        RAISE EXCEPTION 'start_step not check_mode (got %)', v_start;
    END IF;
    IF NOT v_has_derive THEN
        RAISE EXCEPTION 'derive_head_assets step missing';
    END IF;
    RAISE NOTICE 'asset-deployer brand_head mode added';
END
$verify$;

COMMIT;
