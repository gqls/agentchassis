-- FILE: docs/agent_docs/sql_for_agents/307_asset_deployer_stops_declaring_deploy_path.sql
--
-- bugs_open/179 finding A — the config half.
--
-- WHAT AND WHY
-- ------------
-- `deploy_image_asset` no longer honours a `deploy_path`: the override is deleted
-- and an EXPLICIT value is refused before the download and the git commit, so
-- every path the deployer commits is `storage.DeployedAssetPath(asset_key,
-- purpose)` — the same derivation all six readers resolve through.
--
-- The live `asset-deployer` row still DECLARES the input, in two places:
--
--   default_config.workflow.steps.deploy_asset.config.input_fields
--     ["s3_uri", "deploy_path", "purpose", "domain", "asset_key"]
--   input_contract.optional
--     ["deploy_path", "purpose", "asset_key"]
--
-- A declaration is not merely untidy here. `ExtractActionInputs` resolves each
-- DECLARED field by a depth-20 recursive search of the whole of `collected_data`,
-- so while `deploy_path` is declared, a stray `deploy_path` key anywhere in a
-- deploy orchestration is hunted out and bound. That is how the override used to
-- be armed by a caller who never asked for it. The Go change makes such a value
-- inert (it is ignored, and the derived path wins); this migration stops the
-- extractor looking for it at all, and stops the contract advertising to a config
-- author a capability that no longer exists.
--
-- ORDER RELATIVE TO THE IMAGE: FREE, and that is measured, not assumed.
-- Nothing anywhere sets a deploy_path VALUE — 0 in open work-item specs, 0 in
-- active agent definitions, 0 in all orchestration history (matched on the JSON
-- shape '%"deploy_path":"%', not the bare word). So neither order can change any
-- behaviour: applied before the roll, the old binary simply stops receiving a key
-- nobody sends; applied after, nothing was sending one either. No ordering
-- constraint is claimed (the ordering exemption's condition (1) is retired).
--
-- RE-RUNNABLE. Both updates are idempotent: they write a literal array, and the
-- WHERE clause matches only the live non-snapshot row.
--
-- REVERT: the snapshot taken below. `SELECT snapshot_agent(...)` returns its id;
-- restore with the lane's usual restore path if a caller out-of-tree turns out to
-- have depended on the declaration (none exists in this estate — see the census).

BEGIN;

-- Snapshot first, per 107's pattern: an is_snapshot=true copy of the current row.
SELECT snapshot_agent('asset-deployer', 'bugs_open/179 finding A: stop declaring deploy_path') AS asset_deployer_snapshot_id;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_asset,config,input_fields}',
        '["s3_uri", "purpose", "domain", "asset_key"]'::jsonb,
        true),
    input_contract = jsonb_set(
        jsonb_set(
            input_contract,
            '{optional}',
            '["purpose", "asset_key"]'::jsonb,
            true),
        '{description}',
        '"Provide domain + s3_uri. Optional purpose (default: hero) controls resize dimensions; optional asset_key fixes the filename when it differs from purpose. The deploy PATH is derived from (asset_key, purpose) and cannot be chosen — an explicit deploy_path is refused (bugs_open/179 finding A)."'::jsonb,
        true),
    updated_at = now()
WHERE type = 'asset-deployer'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- VERIFY INSIDE THE TRANSACTION, and RAISE rather than SELECT.
--
-- A verify block made of SELECTs cannot stop the COMMIT: ON_ERROR_STOP ignores a
-- non-empty result, so a "verification" that prints the wrong rows still commits
-- them. This one aborts.
DO $$
DECLARE
    bad_fields int;
    bad_contract int;
BEGIN
    SELECT count(*) INTO bad_fields
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,deploy_asset,config,input_fields}' LIKE '%deploy_path%';

    SELECT count(*) INTO bad_contract
      FROM agent_definitions
     WHERE type = 'asset-deployer' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND input_contract::text LIKE '%deploy_path%'
       -- the description deliberately NAMES the removed input to explain the refusal;
       -- only a declaration in `optional` is a failure.
       AND (input_contract->>'optional') LIKE '%deploy_path%';

    IF bad_fields <> 0 OR bad_contract <> 0 THEN
        RAISE EXCEPTION
            'asset-deployer still declares deploy_path (input_fields hits=%, optional hits=%) — '
            'while it is declared, ExtractActionInputs hunts it recursively through collected_data',
            bad_fields, bad_contract;
    END IF;
END $$;

COMMIT;

-- POST-APPLY CHECK (expect exactly one row, and no deploy_path in either column):
--
--   SELECT default_config #>> '{workflow,steps,deploy_asset,config,input_fields}' AS input_fields,
--          input_contract->'optional' AS contract_optional
--     FROM agent_definitions
--    WHERE type='asset-deployer' AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--
-- And the fleet-wide census that must stay at zero VALUES (bare word is a trap —
-- it matches this row's own description and the council submissions):
--
--   SELECT count(*) FROM orchestration_states WHERE collected_data::text LIKE '%"deploy_path":"%';
