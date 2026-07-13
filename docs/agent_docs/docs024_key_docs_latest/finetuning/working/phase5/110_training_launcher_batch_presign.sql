-- Migration 110 — training-launcher: replace the per-checkpoint presign LOOP
-- with a single batch prepare_object_urls dispatch, and drop flatten.
--
-- WHY: the 2026-06-09 run confirmed the send-before-register race fix works, but
-- the per-checkpoint loop slows O(K^2) — every awaited substep re-persists the
-- whole expanded ~2K-substep workflow + growing collected_data/history, so at
-- K=40 it crawled to a halt by iter_9 and never reached write_manifest. This
-- collapses the loop (presign_checkpoints) + flatten_checkpoint_urls into ONE
-- awaited round-trip: dispatch_thunder_prepare_object_urls hands the adapter the
-- whole key array and gets back an ordered presigned_urls[] in one reply.
--
-- Net workflow change (training-launcher.default_config.workflow.steps):
--   compute_keys (unchanged) -> presign_checkpoints (NOW the batch dispatch)
--     -> presign_final (unchanged) -> assemble_manifest (repointed) -> ...
--   flatten_checkpoint_urls: REMOVED.
--
-- Path facts (verified against the live def 2026-06-09):
--   compute_keys.output_field            = "ckpt_keys"  (keys at ckpt_keys.checkpoint_keys,
--                                                         final at ckpt_keys.final_key)
--   presign_checkpoints (loop).next_step = "flatten_checkpoint_urls"
--   flatten_checkpoint_urls.next_step    = "presign_final"  (so the batch step's next_step = presign_final)
--   assemble_manifest.config.checkpoint_keys = "flat_ckpts.checkpoint_keys"  (repoint -> ckpt_keys.checkpoint_keys)
--   assemble_manifest.config.checkpoint_urls = "flat_ckpts.checkpoint_urls"  (repoint -> ckpt_presign_batch.presigned_urls)
--
-- DEPLOY ORDERING (not part of this SQL):
--   * adapter image must carry handlePrepareObjectURLs ("prepare_object_urls" case).
--   * chassis image must carry DispatchThunderPrepareObjectURLsAction + its registry entry.
--   * AFTER both are built+deployed, bump this def's image_tag to the new chassis tag
--     (step 4 below) so the launcher spawns on an image that knows the new action.
--     Applying this migration BEFORE that image is live would make the launcher
--     fail the step with an unknown-action error.

BEGIN;

-- 0. Guard: exactly one live training-launcher def should match.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION 'migration 110: expected exactly 1 live training-launcher def, found %', n;
  END IF;
END $$;

-- 1. Replace presign_checkpoints (loop) with the batch dispatch step.
--    next_step jumps straight to presign_final (flatten is removed in step 3).
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,presign_checkpoints}',
    '{
        "action": "dispatch_thunder_prepare_object_urls",
        "config": {
            "keys": "ckpt_keys.checkpoint_keys",
            "method": "PUT",
            "expiry_minutes": 3000
        },
        "next_step": "presign_final",
        "output_field": "ckpt_presign_batch",
        "description": "Presign ALL K checkpoint PUTs in ONE batch adapter call (prepare_object_urls). Hands the adapter ckpt_keys.checkpoint_keys and gets back ordered presigned_urls[] aligned 1:1. Replaces the per-checkpoint loop (O(K^2) state re-persist) and flatten_checkpoint_urls. expiry_minutes ~50h covers the 48h cap."
    }'::jsonb,
    true)
WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

-- 2. Repoint assemble_manifest inputs:
--      checkpoint_keys -> the canonical compute output (ckpt_keys.checkpoint_keys)
--      checkpoint_urls -> the batch reply (ckpt_presign_batch.presigned_urls)
--    Both are the same ordered set; assemble_upload_manifest's length check guards alignment.
UPDATE agent_definitions
SET default_config = jsonb_set(
    jsonb_set(
        default_config,
        '{workflow,steps,assemble_manifest,config,checkpoint_urls}',
        '"ckpt_presign_batch.presigned_urls"'::jsonb, true),
    '{workflow,steps,assemble_manifest,config,checkpoint_keys}',
    '"ckpt_keys.checkpoint_keys"'::jsonb, true)
WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

-- 3. Remove the now-unused flatten_checkpoint_urls step.
UPDATE agent_definitions
SET default_config = default_config #- '{workflow,steps,flatten_checkpoint_urls}'
WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

-- 4. (DEPLOY — fill in after the chassis is rebuilt with the new dispatch+registry)
--    Uncomment and set the tag the new chassis image was pushed under:
-- UPDATE agent_definitions SET image_tag='<NEW_CHASSIS_TAG>'
--  WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

COMMIT;

-- Verify (run after commit):
--   compute_keys -> presign_checkpoints(batch) -> presign_final -> assemble_manifest -> write_manifest -> ssh_exec_launch -> mark_running -> complete
-- SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_checkpoints}') AS presign_checkpoints,
--        jsonb_pretty(default_config #> '{workflow,steps,assemble_manifest,config}') AS assemble_config,
--        (default_config #> '{workflow,steps,flatten_checkpoint_urls}') IS NULL AS flatten_removed
--   FROM agent_definitions
--  WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
