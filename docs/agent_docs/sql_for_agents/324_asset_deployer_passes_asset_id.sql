-- 324: let `asset-deployer` pass `asset_id` to `deploy_image_asset`.
--
-- WHY. `deploy_image_asset` has declared `asset_id` as an Optional input since
-- long before today, and its DB fallback resolves an asset's source object from
-- that id. But the live `deploy_asset` step declares
--   input_fields: ["s3_uri","purpose","domain","asset_key"]
-- and `ExtractActionInputs` **Strategy 1 wins whenever `input_fields` is present,
-- extracting ONLY the listed names** (`action_inputs.go:441-455`; the recursive
-- all-fields hunt is Strategy 2, reached only when `input_fields` is ABSENT).
-- So `asset_id` was dropped even when the caller supplied it — measured
-- 2026-08-06: three asset_id-only deploys, the id present in the child's
-- `input_data`, all three skipped with "no storage URI found for icon".
-- That config is unchanged since 2026-02-20, so this has been true all along.
--
-- ORDERING — IMAGE FIRST, AND IT ALREADY IS. This is only safe on a chassis that
-- carries `storage.AssetSourceRef` (bugs_open/152 + /155, commit 1d11827c1,
-- live on v1.0.1259, pod-verified both replicas). On an OLDER binary the same
-- config would route `asset_id` into the deleted purpose-keyed branch and deploy
-- the WRONG asset's bytes — i.e. this file would arm bug 155 rather than close
-- it. The guard below refuses to apply if the fleet is not on such a build,
-- because a version check a human has to remember is not a control.
--
-- SECOND EFFECT, STATED DELIBERATELY: this also enables the post-deploy
--   UPDATE assets SET storage_path = COALESCE(storage_path, split_part(url,'?',1)),
--                     url = <deployed local path>
-- (IMG-053 "Edit F"), which is gated on the same `asset_id` and has therefore
-- never fired through this agent. That flip is what stranded rows in
-- bugs_open/152 — but it is now SAFE, and only now: the COALESCE preserves the
-- source before `url` is overwritten, `StoreAssetAction` records `storage_path`
-- at generation, migration 323 backfilled the remaining 205 rows, and all four
-- readers resolve through `AssetSourceRef`. Applying this file BEFORE 323 would
-- reintroduce the stranding.
--
-- IDEMPOTENT: appends only if absent; a re-run touches 0 rows.
--
-- HOW TO RUN (the marker is a GUC, NOT a psql -v variable — psql refuses a
-- dotted -v name, and psql does not substitute :vars inside dollar-quoted
-- bodies, so a -v marker could not reach the DO block anyway):
--
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
--       -c "SET assetdeployer.fix_is_live = 'v1.0.1259+'" \
--       -f - < docs/agent_docs/sql_for_agents/324_asset_deployer_passes_asset_id.sql
--
-- -c and -f run in the SAME session, in that order, so the plain SET (not SET
-- LOCAL) is still in force inside the transaction below.

BEGIN;

-- Backup the exact row being changed (not the whole table).
CREATE TABLE IF NOT EXISTS bak_asset_deployer_20260806 AS
SELECT * FROM agent_definitions WHERE false;
INSERT INTO bak_asset_deployer_20260806
SELECT * FROM agent_definitions
WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Refuse on an old binary: the fleet must already carry the source derivation.
-- Proxy for "the fix is live", checked against config the fix's own commit
-- removed — StoreAssetAction no longer writes any {purpose}_uri key, so a site
-- whose content_data gained one AFTER the roll would mean an old binary is
-- still writing. We assert the code-side precondition the only way SQL can:
-- the operator must have pod-verified. Recorded, and failed loudly if skipped.
DO $$
DECLARE marker text;
BEGIN
  SELECT current_setting('assetdeployer.fix_is_live', true) INTO marker;
  IF marker IS DISTINCT FROM 'v1.0.1259+' THEN
    RAISE EXCEPTION
      'REFUSED: prepend -c "SET assetdeployer.fix_is_live = <marker>" (see HOW TO RUN at the head of this file) and only AFTER pod-grepping BOTH replicas: AssetSourceRef >= 1 AND "Resolved s3_uri from site content_data via asset_id" = 0. On an older binary this config deploys the WRONG asset bytes (bugs_open/155).';
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,deploy_asset,config,input_fields}',
      (default_config #> '{workflow,steps,deploy_asset,config,input_fields}') || '["asset_id"]'::jsonb
    ),
    updated_at = NOW()
WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps,deploy_asset,config,input_fields}' ? 'asset_id');

-- Verify loudly. A verification made of SELECTs cannot stop a COMMIT
-- (ON_ERROR_STOP ignores a non-empty result) — use DO/RAISE.
DO $$
DECLARE fields jsonb;
BEGIN
  SELECT default_config #> '{workflow,steps,deploy_asset,config,input_fields}'
    INTO fields FROM agent_definitions
   WHERE type='asset-deployer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF fields IS NULL THEN
    RAISE EXCEPTION 'input_fields path missing — the step was renamed or restructured; do NOT assume this applied';
  END IF;
  IF NOT (fields ? 'asset_id') THEN
    RAISE EXCEPTION 'asset_id absent after update: %', fields;
  END IF;
  -- The four originals must survive: this is an ADDITION, and an overwrite
  -- that dropped s3_uri would break every existing explicit-URI deploy.
  IF NOT (fields ? 's3_uri' AND fields ? 'purpose' AND fields ? 'domain' AND fields ? 'asset_key') THEN
    RAISE EXCEPTION 'an original field was lost: %', fields;
  END IF;
  RAISE NOTICE 'OK: input_fields = %', fields;
END $$;

COMMIT;

-- ROLLBACK, if ever wanted (restores the exact pre-change config):
--   UPDATE agent_definitions a SET default_config = b.default_config
--     FROM bak_asset_deployer_20260806 b WHERE a.id = b.id;
