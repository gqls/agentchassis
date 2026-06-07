-- FILE: 109_launcher_upload_manifest_wiring.sql
--
-- Phase B: wire the checkpoint/final-adapter presign + upload-manifest path into
-- the training-launcher workflow.
--
-- SURGICAL by design: this ADDS six new steps and rewires ONE existing edge
-- (presign_scripts.next_step). It does NOT rewrite default_config and does NOT
-- touch ssh_exec_launch — so it cannot clobber the live launch command (which
-- differs from the uploaded 102; see VERIFY note below). The manifest is written
-- onto the VM by a separate write_manifest ssh step BEFORE ssh_exec_launch; the
-- file persists on /workspace between the two ssh sessions.
--
-- New chassis actions this depends on (DEPLOY THE CHASSIS FIRST, then apply this):
--   compute_checkpoint_keys     (pure)
--   flatten_presign_results     (pure)
--   assemble_upload_manifest    (pure)
--   dispatch_thunder_prepare_object_url  (key_path source ADDED — additive)
-- If the workflow references an action the deployed chassis lacks, those steps
-- fail with "unknown action" — order matters.
--
-- New sub-sequence (inserted after presign_scripts, rejoining at ssh_exec_launch):
--   compute_keys -> presign_checkpoints (loop) -> flatten_checkpoint_urls
--     -> presign_final -> assemble_manifest -> write_manifest -> ssh_exec_launch
--
-- Idempotent: re-running overwrites the six steps with identical values and
-- re-sets presign_scripts.next_step.
--
-- ───────────────────────────────────────────────────────────────────────────
-- VERIFY BEFORE APPLYING (the uploaded 102 may not match the live def):
--   The live launcher works, but resolveTemplateToken reads a {token}'s source
--   from TOP-LEVEL config[token] (a dot-path), never from config.input_mapping.
--   102 (uploaded) put scripts_url/dataset_url UNDER input_mapping, which would
--   resolve empty — so the live def must already have them at top-level config.
--   This migration does NOT touch ssh_exec_launch, so that discrepancy doesn't
--   block it; but confirm the live shape and that presign_scripts exists:
--
--     SELECT jsonb_pretty(default_config #> '{workflow,steps,presign_scripts}')
--       FROM agent_definitions WHERE type='training-launcher'
--        AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
--     SELECT jsonb_pretty(default_config #> '{workflow,steps,ssh_exec_launch,config}')
--       FROM agent_definitions WHERE type='training-launcher'
--        AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
--
--   (The second is informational for the upcoming Phase C run.sh wiring.)
-- ───────────────────────────────────────────────────────────────────────────

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
    -- inner: add the six new steps to workflow.steps
    jsonb_set(
        default_config,
        '{workflow,steps}',
        (default_config #> '{workflow,steps}') || $newsteps$
        {
          "compute_keys": {
            "action": "compute_checkpoint_keys",
            "description": "Build K checkpoint keys + final-adapter key from training_run_id. K = config checkpoint_count fallback (max_steps not known cluster-side yet).",
            "config": { "checkpoint_count": 40, "buffer": 4 },
            "output_field": "ckpt_keys",
            "next_step": "presign_checkpoints"
          },
          "presign_checkpoints": {
            "action": "loop",
            "description": "Presign one write-only PUT per checkpoint key via the existing single-key prepare_object_url. max_iterations must be >= the 512 key clamp or the loop silently truncates.",
            "config": {
              "items_field": "ckpt_keys.checkpoint_keys",
              "item_variable": "ckpt_key",
              "max_iterations": 512,
              "continue_on_error": false,
              "sub_workflow": {
                "start_step": "presign_one",
                "steps": {
                  "presign_one": {
                    "action": "dispatch_thunder_prepare_object_url",
                    "description": "Presign a PUT for this iteration key. Reads the loop item via key_path (a config dot-path resolved from CollectedData, where setLoopVariable puts ckpt_key each iteration) — the proven loop-substep pattern. input_mapping does NOT thread the var here, so the dispatch would otherwise fall back to dataset_uri. key_path resolves before that fallback. expiry_minutes overrides the 24h PUT default (~50h covers the 48h cap).",
                    "config": { "method": "PUT", "expiry_minutes": 3000, "key_path": "ckpt_key" },
                    "output_field": "ckpt_presign",
                    "next_step": "presign_done"
                  },
                  "presign_done": {
                    "action": "loop_complete",
                    "description": "Terminal substep. Every production loop ends on loop_complete and the async work substep chains here rather than being the terminal itself — the iteration boundary fires on this step's empty next_step."
                  }
                }
              }
            },
            "output_field": "ckpt_presigns",
            "next_step": "flatten_checkpoint_urls"
          },
          "flatten_checkpoint_urls": {
            "action": "flatten_presign_results",
            "description": "Reshape the loop results[] into flat, ordered checkpoint_urls[] + checkpoint_keys[] for assemble.",
            "config": {
              "presign_results": "ckpt_presigns.results",
              "url_field": "ckpt_presign.presigned_url",
              "key_field": "ckpt_presign.key"
            },
            "output_field": "flat_ckpts",
            "next_step": "presign_final"
          },
          "presign_final": {
            "action": "dispatch_thunder_prepare_object_url",
            "description": "Presign a PUT for the final adapter key. Plain local step: reads the dynamic key via key_path (input_mapping is dead for local steps). Same expiry override.",
            "config": { "method": "PUT", "expiry_minutes": 3000, "key_path": "ckpt_keys.final_key" },
            "output_field": "final_presign",
            "next_step": "assemble_manifest"
          },
          "assemble_manifest": {
            "action": "assemble_upload_manifest",
            "description": "Build upload_manifest.json content (+ base64) from the flattened checkpoint url/key lists, the final key+url, and training_run_id.",
            "config": {
              "training_run_id": "input_data.training_run_id",
              "checkpoint_urls": "flat_ckpts.checkpoint_urls",
              "checkpoint_keys": "flat_ckpts.checkpoint_keys",
              "final_key": "ckpt_keys.final_key",
              "final_url": "final_presign.presigned_url"
            },
            "output_field": "manifest_result",
            "next_step": "write_manifest"
          },
          "write_manifest": {
            "action": "dispatch_thunder_ssh_exec",
            "description": "Write upload_manifest.json onto the VM via base64 -d (dodges shell quoting of the URL-laden JSON). Separate ssh so ssh_exec_launch is untouched; provisioning_id comes from input_data automatically.",
            "config": {
              "command_template": "sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace && echo '{manifest_b64}' | base64 -d > /workspace/upload_manifest.json && echo MANIFEST_WRITTEN",
              "manifest_b64": "manifest_result.manifest_b64"
            },
            "output_field": "manifest_write_result",
            "next_step": "ssh_exec_launch"
          }
        }
        $newsteps$::jsonb
    ),
    -- outer: rewire the existing presign_scripts step to enter the new sub-sequence.
    -- (No-op if presign_scripts is absent — jsonb_set won't create the missing parent.)
    '{workflow,steps,presign_scripts,next_step}',
    '"compute_keys"'::jsonb
)
WHERE type = 'training-launcher'
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

COMMIT;

-- ───────────────────────────────────────────────────────────────────────────
-- Post-apply verification (run manually):
-- ───────────────────────────────────────────────────────────────────────────
-- \echo presign_scripts now enters the manifest path (expect "compute_keys"):
-- SELECT default_config #>> '{workflow,steps,presign_scripts,next_step}'
--   FROM agent_definitions WHERE type='training-launcher'
--    AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
-- \echo the six new steps exist and chain to ssh_exec_launch:
-- SELECT s.key AS step, s.value->>'action' AS action, s.value->>'next_step' AS next_step
--   FROM agent_definitions a,
--        jsonb_each(a.default_config #> '{workflow,steps}') s
--  WHERE a.type='training-launcher'
--    AND (a.is_snapshot IS NULL OR a.is_snapshot=false) AND a.deleted_at IS NULL
--    AND s.key IN ('compute_keys','presign_checkpoints','flatten_checkpoint_urls',
--                  'presign_final','assemble_manifest','write_manifest')
--  ORDER BY s.key;
