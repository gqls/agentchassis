-- 401 — image-build-handler: call_asset_deployer never mapped asset_id, so it
-- resolved (or failed to resolve) via the aggressive recursive search
--
-- WHAT: adds one optional key to the live image-build-handler definition's
-- call_asset_deployer step config: "asset_id?": "asset_stored.asset_id",
-- alongside the domain/s3_uri/purpose/asset_key? keys that step already maps.
--
-- WHY (bugs_open/248, council round 1, editquality's gating HIGH objection):
-- 248's shipped fix (930ace3bd, platform/orchestration/actions/
-- deploy_image_asset_action.go) makes the asset ROW the authority for
-- purpose/asset_key when a run states neither, via inputs.Get("asset_id").
-- ExtractActionInputs resolves a field explicitly (Strategy 0/1, deterministic)
-- only when the CALLER's input_mapping names it; otherwise it falls through to
-- findFieldRecursive — an aggressive, map-iteration-order-dependent search
-- across the whole collected_data tree (platform/orchestration/datahelpers/
-- unified_extractor.go). image-build-handler's call_asset_deployer step
-- (action: call_agent) mapped only {domain, s3_uri, purpose, asset_key?} —
-- asset_id was absent, so it reached deploy_asset via that aggressive search on
-- this path, exactly the risk editquality named. check_undeployed_assets'
-- OWN dispatch already carries asset_id in the work item's spec and was never
-- at risk; this migration closes the other caller.
--
-- store_imagery_asset (the step immediately before this one, output_field
-- asset_stored, action: store_asset) already returns asset_id in its result
-- (StoreAssetAction, platform/orchestration/actions/v3_site_actions.go) right
-- beside asset_stored.image_uri and asset_stored.purpose, which this same
-- mapping already reads — so the fix is one explicit dot-path, not a Go change.
--
-- MEASURED, not assumed: 5 real image-build-handler-parented asset-deployer
-- completions (2026-08-12/13, before this migration) all show asset_id empty
-- in the child's resolved input_data — the safe-but-incomplete failure mode
-- observed in production, not a wrong-asset pull. No historical row was found
-- where asset_key was ALSO empty (the shape where assetRowIdentity's fallback
-- for purpose actually engages), so that branch's behaviour under the exact
-- shape the objection is about remains unexercised in the sample examined —
-- recorded honestly in the council resubmission rather than overclaimed here.
--
-- asset_id? (optional, matching the asset_key? convention already live in the
-- same mapping): store_asset's "locked" and "no asset URL found" refusal
-- branches return no asset_id at all, and next_step/error_step both proceed to
-- spawn_asset_deployer regardless — an optional marker is required so those
-- branches do not turn into a hard input_mapping failure.
--
-- Idempotent: the UPDATE is fenced on the key being absent, so a replay
-- (or, as here, recording after an out-of-band apply) matches no row.
-- DB-only; snapshot-prefixed.

BEGIN;

SELECT snapshot_agent('image-build-handler',
    '401_image_build_handler_explicit_asset_id_mapping.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,call_asset_deployer,config,input_mapping,asset_id?}',
         '"asset_stored.asset_id"'::jsonb
       ),
       updated_at = now()
 WHERE type = 'image-build-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND NOT (default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}' ? 'asset_id?');

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 401: image-build-handler's deploy call never named asset_id, so it reached the deployer via the aggressive recursive search
Observed (bugs_open/248, council R1): image-build-handler's call_asset_deployer step mapped only {domain, s3_uri, purpose, asset_key?} to the spawned asset-deployer agent. 248's own fix makes the asset row the authority for purpose/asset_key when a run states neither, keyed on asset_id — but asset_id was never explicitly supplied on this path, so it fell through to findFieldRecursive's aggressive, map-order-dependent search instead of a deterministic explicit resolution.
Root cause: store_imagery_asset (the step immediately before this one) already returns asset_id in its result (StoreAssetAction, v3_site_actions.go), alongside image_uri and purpose which the SAME mapping already reads — the field was available and simply not named.
Fix: one optional key added to the live mapping, asset_id? -> asset_stored.asset_id, mirroring the asset_key? convention already in the same mapping. No Go change, no roll.
Verified: 5 real completions pre-fix all show asset_id empty (not wrong) in the resolved input_data — the safe-but-incomplete case, not the wrong-asset-pull the objection feared, though no historical row exercised the asset_key-empty shape where the fallback for purpose actually matters. The DO blocks below assert the new key resolves to the right literal and that the neighbouring keys and sibling steps survived the write.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE mapping jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}' INTO mapping
    FROM agent_definitions
    WHERE type = 'image-build-handler'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF mapping IS NULL THEN
        RAISE EXCEPTION '401: no live image-build-handler config at {workflow,steps,call_asset_deployer,config,input_mapping}';
    END IF;
    IF mapping ->> 'asset_id?' IS DISTINCT FROM 'asset_stored.asset_id' THEN
        RAISE EXCEPTION '401: asset_id? does not map to asset_stored.asset_id (%)', mapping;
    END IF;
END $$;

-- A verify block that only checks the new key cannot tell an explicit add from
-- a write that replaced the object, the step or the workflow. Assert the three
-- neighbouring keys and the surrounding steps too.
DO $$
DECLARE mapping jsonb; steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}',
           default_config #> '{workflow,steps}'
      INTO mapping, steps
    FROM agent_definitions
    WHERE type = 'image-build-handler'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF mapping ->> 'domain' IS DISTINCT FROM 'site_record.domain'
       OR mapping ->> 's3_uri' IS DISTINCT FROM 'asset_stored.image_uri'
       OR mapping ->> 'purpose' IS DISTINCT FROM 'asset_stored.purpose'
       OR mapping ->> 'asset_key?' IS DISTINCT FROM 'input_data.spec.asset_key' THEN
        RAISE EXCEPTION '401: a neighbouring mapping key did not survive (%)', mapping;
    END IF;
    IF steps -> 'store_imagery_asset' IS NULL OR steps -> 'mark_work_item_complete' IS NULL THEN
        RAISE EXCEPTION '401: a sibling step vanished — the write flattened the workflow (%)', steps;
    END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,call_asset_deployer,config,input_mapping}'
--   FROM agent_definitions WHERE type='image-build-handler'
--     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- expect exactly: {"domain": "site_record.domain", "s3_uri": "asset_stored.image_uri",
--   --   "purpose": "asset_stored.purpose", "asset_id?": "asset_stored.asset_id",
--   --   "asset_key?": "input_data.spec.asset_key"}
-- Rollback: restore the snapshot taken above (agent_definitions, is_snapshot=true
-- row for image-build-handler, source_version=1) — the two-arg snapshot_agent
-- overload used elsewhere in this estate writes to agent_definitions_backup
-- instead; THIS call is the one-arg form, so the snapshot is in agent_definitions
-- itself. Confirm which before relying on it: `pg_get_functiondef` distinguishes
-- the two overloads (bugfix_134's council round, same trap).
--
-- APPLIED OUT OF BAND, 2026-08-13, before this file existed: the same
-- snapshot_agent + jsonb_set was run by hand via psql (staged_component_build
-- lane, bugs_open/248 R1 response), confirmed committed inside that
-- transaction. This file exists so the ledger and any future --apply agree
-- with what is already live, and so a fresh database gets the fix on seed
-- replay. Register with:
--   ./scripts/migration/run-migrations.sh --record-only \
--     401_image_build_handler_explicit_asset_id_mapping.sql \
--     --note "applied by hand 2026-08-13 ~17:10Z; asset_id? confirmed present in the live mapping inside the committing transaction before this file was written; this file's own guard (UPDATE fenced on the key's absence) makes a future --apply a safe no-op"
