-- 402 — build-dispatch-loop: the undeployed_asset repair item's asset_id is
-- nested under spec, so assetRowIdentity recovers it via the aggressive
-- recursive search instead of an explicit path, exactly like 380's purpose fix
--
-- WHY (bugs_open/248, council round 2, editquality's HIGH "missing" objection):
-- migration 401 closed this same class of gap for image-build-handler's own
-- call_asset_deployer dispatch. editquality correctly pointed out that the
-- REPAIR path (check_undeployed_assets.go's undeployed_asset item, dispatched
-- through build-dispatch-loop -> asset-deployer) was asserted safe in the R1/R2
-- submission on the strength of asset_id being PRESENT somewhere in the data,
-- not on having traced HOW it resolves — the same shortcut this bug's own
-- diagnosis warns against elsewhere.
--
-- THE DEFECT, traced (same shape as 380, different field): build-dispatch-loop's
-- call_handler maps the item spec NESTED ("spec": "current_item.spec"), with no
-- top-level asset_id. asset-deployer's deploy_asset step declares asset_id in
-- its input_fields, so ExtractActionInputs looks for a top-level "asset_id" key
-- (Strategy 1/3 both check the TOP of input_data, not one level into spec) and,
-- finding none there, falls through to Strategy 4's aggressive findFieldRecursive
-- search across the whole collected_data tree -- the exact landmined mechanism
-- editquality's R1 objection named for the OTHER caller. asset_id carries no
-- spec.Default (unlike purpose), so 380's specific failure mode (a default
-- silently winning) does not apply here -- but the aggressive-search risk 380's
-- own migration exists to avoid does.
--
-- MEASURED, not assumed: fleet-wide, exactly ONE (item_type, handler_agent) pair
-- carries asset_id in its spec -- (undeployed_asset, asset-deployer), 267 rows.
-- So this key activates for exactly the repair path bugs_open/248 diagnoses and
-- is a no-op for every other item build-dispatch-loop ever touches.
--
-- THE FIX. Map asset_id top-level in the dispatch mapping, optional, mirroring
-- 380's own idiom exactly:
--     "asset_id?": "current_item.spec.asset_id"
-- The `?` suffix (input_contracts/input_mapping.go:102) silently skips items
-- whose spec has no asset_id -- a no-op for the other 636+ dispatched item
-- types. This is the SAME idiom 380 already established on this SAME mapping
-- for `purpose?`, so this migration is a second, narrower instance of an
-- already-reviewed pattern, not a new kind of change to the seam.
--
-- Idempotent: the UPDATE is fenced on the key being absent, so a replay matches
-- no row. DB-only; snapshot-prefixed (380 itself omits the snapshot_agent call
-- that run-migrations.sh's own header says should open every migration
-- touching agent_definitions -- not repeated here).

BEGIN;

SELECT snapshot_agent('build-dispatch-loop',
    '402_build_dispatch_loop_maps_asset_id_top_level.sql: pre-update');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,asset_id?}',
        '"current_item.spec.asset_id"'::jsonb,
        true
    ),
    updated_at = now()
WHERE type = 'build-dispatch-loop'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT (default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}' ? 'asset_id?');

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline', 'build',
$note$## 402: the undeployed_asset repair path's asset_id was nested under spec, reachable only by the aggressive recursive search
Observed (bugs_open/248, council R2, editquality HIGH): the R1/R2 submission traced and fixed image-build-handler's call_asset_deployer omitting asset_id from its input_mapping, but asserted the check_undeployed_assets repair path was "safe" because asset_id is present in the item's spec, without tracing HOW it resolves. It resolves the same way image-build-handler's did before migration 401: build-dispatch-loop's call_handler maps spec as a whole NESTED object, so asset-deployer's own input_fields lookup for a top-level asset_id key misses, and ExtractActionInputs falls through to findFieldRecursive's aggressive search.
Root cause: identical shape to migration 380's purpose fix (bugs_open/231), different field. asset_id carries no spec.Default, so 380's specific silent-default failure does not apply, but the aggressive-search risk 380 itself was written to avoid does.
Fix: one optional top-level key added to the SAME mapping 380 already touched, asset_id? -> current_item.spec.asset_id, mirroring 380's own idiom. Measured blast radius: exactly one (item_type, handler_agent) pair fleet-wide carries asset_id in its spec -- (undeployed_asset, asset-deployer), 267 rows -- so this is a no-op for every other dispatched item type.
Categories: migration$note$,
'["migration"]'::jsonb, 'human', 'run-migrations.sh');

DO $$
DECLARE mapping jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}' INTO mapping
    FROM agent_definitions
    WHERE type = 'build-dispatch-loop'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF mapping IS NULL THEN
        RAISE EXCEPTION '402: no live build-dispatch-loop config at {...call_handler,config,input_mapping}';
    END IF;
    IF mapping ->> 'asset_id?' IS DISTINCT FROM 'current_item.spec.asset_id' THEN
        RAISE EXCEPTION '402: asset_id? does not map to current_item.spec.asset_id (%)', mapping;
    END IF;
END $$;

-- A verify block that only checks the new key cannot tell an explicit add from
-- a write that replaced the object or the step. Assert 380's own key, the
-- top-level workflow steps, AND the sub_workflow's own sibling steps (the real
-- neighbours of call_handler, which sits one level deeper than {workflow,steps}).
DO $$
DECLARE mapping jsonb; top_steps jsonb; sub_steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}',
           default_config #> '{workflow,steps}',
           default_config #> '{workflow,steps,process_item,config,sub_workflow,steps}'
      INTO mapping, top_steps, sub_steps
    FROM agent_definitions
    WHERE type = 'build-dispatch-loop'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF mapping ->> 'purpose?' IS DISTINCT FROM 'current_item.spec.purpose'
       OR mapping ->> 'spec' IS DISTINCT FROM 'current_item.spec'
       OR mapping ->> 'work_item_id' IS DISTINCT FROM 'current_item.id' THEN
        RAISE EXCEPTION '402: a neighbouring mapping key did not survive (%)', mapping;
    END IF;
    IF top_steps -> 'process_item' IS NULL OR top_steps -> 'load_items' IS NULL THEN
        RAISE EXCEPTION '402: a top-level sibling step vanished (%)', top_steps;
    END IF;
    IF sub_steps -> 'claim' IS NULL OR sub_steps -> 'mark_complete' IS NULL OR sub_steps -> 'mark_failed' IS NULL THEN
        RAISE EXCEPTION '402: a sub_workflow sibling step vanished — the write flattened the loop (%)', sub_steps;
    END IF;
END $$;

COMMIT;

-- Verify:
--   SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}'
--   FROM agent_definitions WHERE type='build-dispatch-loop'
--     AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- expect asset_id? and purpose? both present, mapped to current_item.spec.*
-- Rollback: restore the snapshot taken above (agent_definitions, is_snapshot=true
-- row for build-dispatch-loop, one-arg snapshot_agent overload -- the row lives
-- in agent_definitions itself, not agent_definitions_backup; confirm which
-- overload with pg_get_functiondef before relying on it, per the bugfix_134
-- council round's finding that the two overloads write to different tables).

-- ─────────────────────────────────────────────────────────────────────────────
-- CORRECTION 2026-08-15 (RFC_029 §6/§9 D4 — the Strategy-number collision this
-- file itself became the worked example of). The header above says resolution
-- "falls through to Strategy 4's aggressive findFieldRecursive search". That
-- citation is WRONG in a way that sends a reader to the wrong file:
-- action_inputs.go's own "Strategy 4" is a DIFFERENT arm ("resolve remaining
-- config value references"). The aggressive search was unified_extractor.go's
-- extractSingleField arm 4 — a second, independently-numbered chain one call
-- away. As of RFC_029 Phase 1 the inner chain's arms carry DESCRIPTIVE names
-- (direct-path, input-data-prefix, input-data-map, whole-tree-search, alias) so
-- this class of miscitation cannot recur: the arm this file means is
-- "whole-tree-search". The MECHANISM described above is otherwise correct, and
-- nothing about the applied UPDATE changes. This note corrects the record only.
--
-- (Also noted, same date: this mapping was NAMED as the `!` strict marker's
-- second adopter by RFC_029 §9 D3 and was NOT flipped — the mapping is shared
-- by every dispatched item type, so its `?` is per-item-type optionality and a
-- strict marker here would hard-fail every non-asset dispatch. See migration
-- 417_HOLD's header and RFC_029 §9's dated correction.)
