-- 266_asset_deployer_ingest_mode.sql
--
-- Adds the FOURTH mode branch to asset-deployer: ingest_upload — the
-- operator's asset-amend path (bugs_open/131 og-card slug). Mirrors the three
-- prior mode migrations exactly (SQL_2026-07-11 brand_head, SQL_2026-07-13
-- sprite_css, SQL_2026-07-16 content_card in docs024 imagery/):
--
--   check_mode(brand_head) → check_sprite_mode(sprite_css)
--     → check_card_mode(content_card) → check_ingest_mode(ingest_upload)  ← NEW
--       → else deploy_asset (default, unchanged)
--
-- The rewire touches ONLY check_card_mode.else_step; the three existing mode
-- conditions are disjoint string-equality checks and are not modified.
--
-- ⚠ ORDERING — IMAGE BEFORE SEED. Apply ONLY after the chassis image carrying
-- ingest_staged_asset (registry.go) is rolled and pod-grepped:
--   kubectl exec -n ai-persona-system <pod> -- sh -c \
--     'strings /app/agent-chassis | grep -c "ingest_staged_asset"'
-- A seed naming an unregistered action fails at runtime. 265 (the staging
-- table) is inert and may be applied any time; this one is the live half.
--
-- ROLLBACK: restore from the snapshot taken below, or surgically:
--   UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--     '{workflow,steps,check_card_mode,config,else_step}', to_jsonb('deploy_asset'::text))
--   WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false;
--   (then optionally delete the two new steps)

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS agent_def_asset_deployer_backup_20260729_ingest AS
SELECT * FROM agent_definitions
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

BEGIN;

SELECT snapshot_agent('asset-deployer',
                      'add ingest_upload mode branch (operator asset-amend path, bugs_open/131 og-card)');

-- Guard: the chain must look the way we think it does before we rewire it.
DO $guard$
DECLARE
    v_else text;
BEGIN
    SELECT default_config #>> '{workflow,steps,check_card_mode,config,else_step}'
      INTO v_else
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active = true
       AND (is_snapshot IS NULL OR is_snapshot = false);
    IF v_else IS DISTINCT FROM 'deploy_asset' THEN
        RAISE EXCEPTION 'check_card_mode.else_step is % (expected deploy_asset) — the chain has changed since this migration was written; re-read it before applying', v_else;
    END IF;
END
$guard$;

-- New conditional, slotted after check_card_mode.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,check_ingest_mode}',
        '{
           "action": "conditional",
           "config": {
             "condition": "input_data.spec.mode == \"ingest_upload\" OR input_data.mode == \"ingest_upload\"",
             "then_step": "ingest_staged_asset_step",
             "else_step": "deploy_asset"
           },
           "description": "Route operator asset-amend ingests away from the default image deploy"
         }'::jsonb)
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- The ingest step itself.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,ingest_staged_asset_step}',
        '{
           "action": "ingest_staged_asset",
           "config": {
             "staging_id": "input_data.spec.staging_id",
             "site_id": "input_data.site_id"
           },
           "next_step": "complete",
           "description": "Validate staged operator bytes, upload to S3 at a new key, amend the assets row",
           "output_field": "ingest_result"
         }'::jsonb)
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- Rewire the fall-through.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config, '{workflow,steps,check_card_mode,config,else_step}',
        to_jsonb('check_ingest_mode'::text)),
    updated_at = now()
WHERE type = 'asset-deployer' AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

DO $verify$
DECLARE
    v_card_else  text;
    v_ingest_else text;
    v_has_step   boolean;
BEGIN
    SELECT default_config #>> '{workflow,steps,check_card_mode,config,else_step}',
           default_config #>> '{workflow,steps,check_ingest_mode,config,else_step}',
           (default_config #> '{workflow,steps}') ? 'ingest_staged_asset_step'
      INTO v_card_else, v_ingest_else, v_has_step
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active = true
       AND (is_snapshot IS NULL OR is_snapshot = false);
    IF v_card_else <> 'check_ingest_mode' THEN
        RAISE EXCEPTION 'check_card_mode.else_step not rewired (got %)', v_card_else;
    END IF;
    IF v_ingest_else <> 'deploy_asset' THEN
        RAISE EXCEPTION 'check_ingest_mode.else_step is % (default path broken!)', v_ingest_else;
    END IF;
    IF NOT v_has_step THEN
        RAISE EXCEPTION 'ingest_staged_asset_step missing';
    END IF;
    RAISE NOTICE 'asset-deployer ingest_upload mode added; default deploy_asset fall-through preserved';
END
$verify$;

COMMIT;
