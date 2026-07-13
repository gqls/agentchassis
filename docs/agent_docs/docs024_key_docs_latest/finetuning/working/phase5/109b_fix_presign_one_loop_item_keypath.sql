-- 109b — fix presign_checkpoints loop item binding (key_path, not input_mapping)
-- Apply to clients_db. Use this because 109 is ALREADY applied; this patches the live row.
-- (Canonical 109 corrected too, for a clean re-apply.)
--
-- Why:
--   presign_one (the per-checkpoint loop substep) presigned the DATASET key on EVERY
--   iteration instead of its own finetuning/checkpoints/<run>/ckpt-<N>.tar.gz key.
--   Confirmed in the launcher pod: every presign_checkpoints_iter_N_presign_one dispatch
--   logged object_key="finetuning/datasets/146a9a12…/training.jsonl".
--
--   Cause: presign_one used  input_mapping {key: ckpt_key}.  input_mapping did NOT populate
--   input_data.key for this loop substep, so dispatch_thunder_prepare_object_url ran its
--   key-resolution chain — explicit key → key_path → fall back to input_data.dataset_uri —
--   and hit the dataset_uri fallback (the launcher's standard single-dataset case),
--   deriving the dataset key 40 times. Net: no real per-key PUT URLs, which breaks
--   flatten_checkpoint_urls → assemble_upload_manifest → the manifest's checkpoints[].
--
--   Proven production loops (vet-batch-processor, the rebuild/fix-items loops, etc.) read
--   their item via a CONFIG DOT-PATH the action resolves itself, e.g.
--   "work_item_id":"current_item.id", "component_id":"current_component.component_id" —
--   NOT via input_mapping. presign_final in THIS def already uses key_path. So the fix is
--   to make presign_one match: read the loop item via key_path "ckpt_key", which reads
--   CollectedData["ckpt_key"] (where setLoopVariable puts each iteration's item) and
--   resolves BEFORE the dataset_uri fallback. No Go change.
--
-- Surgical: one jsonb_set replacing presign_one.config (drops input_mapping, adds key_path).
-- The loop re-expands from this template at runtime, so every iteration's presign_one
-- picks it up. Idempotent.

-- VERIFY BEFORE (expect input_mapping {key: ckpt_key}, no key_path):
--   SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_checkpoints,config,sub_workflow,steps,presign_one,config}')
--     FROM agent_definitions WHERE type='training-launcher'
--      AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,presign_checkpoints,config,sub_workflow,steps,presign_one,config}',
    '{"method":"PUT","expiry_minutes":3000,"key_path":"ckpt_key"}'::jsonb,
    false
)
WHERE type='training-launcher'
  AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;

COMMIT;

-- VERIFY AFTER (expect key_path "ckpt_key", no input_mapping):
--   SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_checkpoints,config,sub_workflow,steps,presign_one,config}')
--     FROM agent_definitions WHERE type='training-launcher'
--      AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
