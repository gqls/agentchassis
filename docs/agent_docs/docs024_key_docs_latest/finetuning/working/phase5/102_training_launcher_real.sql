-- FILE: 102_training_launcher_real.sql
--
-- Phase 5: replace the training-launcher STUB with the real multi-step
-- workflow, and thread provisioning_id from the provisioner to the launcher
-- in model-trainer.
--
-- Pattern (Option A): proven orchestrator-workflow + dispatch-action + await,
-- exactly like gpu-provisioner / thunder-reaper. Each round-trip step uses a
-- dispatch action that publishes to system.adapter.thunder.requests and returns
-- await_response:true; the saga coordinator pauses between steps.
--
-- New chassis actions this migration depends on (must be deployed FIRST):
--   dispatch_thunder_prepare_object_url  (presign a B2 object by key or s3_uri)
--   dispatch_thunder_ssh_exec            (run a command on the instance by provisioning_id)
--   mark_training_run_running            (training_runs pending -> running)
--
-- Launch logic itself lives in run.sh inside the scripts bundle at
--   <bucket>/finetuning/scripts/bundle.tar.gz
-- so this workflow stays declarative: presign two URLs, run one script.
--
-- Surgical UPDATEs (not a re-INSERT): the stub row already has the right
-- type/category/image_tag/output_contract; we only replace default_config and
-- extend input_contract (add provisioning_id). Mirrors 101_switch_to_haiku.sql
-- and swap_agent_model's targeted default_config updates.
--
-- Idempotent: re-running overwrites default_config / input_contract with the
-- same values and re-applies the model-trainer mapping addition.

BEGIN;

-- ───────────────────────────────────────────────────────────────────────────
-- 1. training-launcher: replace stub default_config with the real workflow.
-- ───────────────────────────────────────────────────────────────────────────
--
-- Step flow:
--   presign_dataset  -> presign_scripts -> ssh_exec_launch -> mark_running -> complete
--
-- Data threading (input_mapping references prior steps' output_field):
--   presign_dataset.dataset_url_result  (from input_data.dataset_uri via s3_uri)
--   presign_scripts.scripts_url_result  (from the literal bundle key in config)
--   ssh_exec_launch.launch_result       (curl both URLs, nohup run.sh, echo PID)
--   mark_running.mark_result            (training_runs -> running)
--
-- NOTE on the ssh_exec command: it is a SINGLE LINE (no literal newlines) so it
-- survives JSON + Kafka transport cleanly (see debugging guide §9 heredoc trap).
-- The complex chain is in run.sh, not here. The command:
--   - makes /workspace, cd's in
--   - curls the presigned scripts bundle, untars it
--   - curls the presigned dataset to the path run.sh expects
--   - nohup's run.sh fully detached, stdout->train.log, prints LAUNCH_PID=<pid>
--
-- The presigned URLs are injected by input_mapping as
--   input_data.scripts_url / input_data.dataset_url at the ssh_exec step. The
--   command references them via ${...} placeholders the dispatch action does
--   NOT expand — so we instead pass the assembled command through input_mapping.
--   Because input_mapping does field references (not string interpolation), the
--   command template uses the literal presigned URLs resolved at run time by the
--   step reading them from collected_data. See the command_template note below.

UPDATE agent_definitions
SET default_config = jsonb_build_object(
        'workflow', jsonb_build_object(
                'processing_mode', 'task',
                'timeout_seconds', 600,
                'start_step', 'presign_dataset',
                'steps', jsonb_build_object(

                    -- ── Step 1: presign the dataset (GET) from its s3:// dataset_uri ──
                    -- method is a constant (config). s3_uri is a runtime reference resolved
                    -- by input_mapping from the parent's dataset_uri.
                        'presign_dataset', jsonb_build_object(
                                'action', 'dispatch_thunder_prepare_object_url',
                                'description', 'Presign a GET URL for the training dataset. s3_uri (runtime) is the preparer''s s3:// dataset_uri; the action strips it to a bucket-relative key. method is a config constant.',
                                'config', jsonb_build_object(
                                        'method', 'GET',
                                        'input_mapping', jsonb_build_object(
                                                's3_uri', 'input_data.dataset_uri'
                                                         )
                                          ),
                                'output_field', 'dataset_url_result',
                                'next_step', 'presign_scripts'
                                           ),

                    -- ── Step 2: presign the scripts bundle (GET) by literal key ──
                    -- Both key and method are constants → config. No runtime refs, so no
                    -- input_mapping.
                        'presign_scripts', jsonb_build_object(
                                'action', 'dispatch_thunder_prepare_object_url',
                                'description', 'Presign a GET URL for the scripts bundle (run.sh + setup + train). Literal key constant in config; not ID-derived.',
                                'config', jsonb_build_object(
                                        'key', 'finetuning/scripts/bundle.tar.gz',
                                        'method', 'GET'
                                          ),
                                'output_field', 'scripts_url_result',
                                'next_step', 'ssh_exec_launch'
                                           ),

                    -- ── Step 3: ssh_exec — fetch both, background run.sh, capture PID ──
                    -- command_template is a CONSTANT (config). The action substitutes the
                    -- {scripts_url}/{dataset_url} tokens from input_data.scripts_url /
                    -- input_data.dataset_url, which input_mapping fills from the two presign
                    -- steps' presigned_url outputs. provisioning_id (runtime) comes from the
                    -- parent (model-trainer threads provisioning_result.provisioning_id).
                    -- Single-line template; the real chain is in run.sh.
                        'ssh_exec_launch', jsonb_build_object(
                                'action', 'dispatch_thunder_ssh_exec',
                                'description', 'SSH to the instance (by provisioning_id): curl the presigned scripts bundle and dataset, then nohup run.sh (setup->smoke->full). Returns quickly with LAUNCH_PID in stdout.',
                                'config', jsonb_build_object(
                                        'command_template', 'mkdir -p /workspace && cd /workspace && curl -fsSL "{scripts_url}" -o bundle.tar.gz && tar -xzf bundle.tar.gz && curl -fsSL "{dataset_url}" -o /workspace/training_iter0.jsonl && chmod +x run.sh && nohup ./run.sh > /workspace/train.log 2>&1 & echo "LAUNCH_PID=$!"',
                                        'input_mapping', jsonb_build_object(
                                                'provisioning_id', 'input_data.provisioning_id',
                                                'scripts_url', 'scripts_url_result.presigned_url',
                                                'dataset_url', 'dataset_url_result.presigned_url'
                                                         )
                                          ),
                                'output_field', 'launch_result',
                                'next_step', 'mark_running'
                                           ),

                    -- ── Step 4: flip training_runs pending -> running ──
                        'mark_running', jsonb_build_object(
                                'action', 'mark_training_run_running',
                                'description', 'Transition training_runs to running and stamp started_at now that the process is launched.',
                                'config', jsonb_build_object(
                                        'input_mapping', jsonb_build_object(
                                                'training_run_id', 'input_data.training_run_id'
                                                         )
                                          ),
                                'output_field', 'mark_result',
                                'next_step', 'complete'
                                        ),

                    -- ── Step 5: complete — surface launched_at + launch_pid ──
                        'complete', jsonb_build_object(
                                'action', 'complete_workflow',
                                'description', 'Return launch_result so the parent sees launched_at / launch_pid.',
                                'config', jsonb_build_object(
                                        'output_field', 'launch_result'
                                          )
                                    )

                         )
                    )
                     )
WHERE type = 'training-launcher'
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

-- ───────────────────────────────────────────────────────────────────────────
-- 1b. Extend the launcher input_contract: add provisioning_id (required now).
--     Keep training_run_id required; keep the existing optionals. dataset_uri
--     moves to required (the launcher needs it to presign).
-- ───────────────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET input_contract = jsonb_build_object(
        'required', jsonb_build_array('training_run_id', 'provisioning_id', 'dataset_uri'),
        'optional', jsonb_build_array('instance_ip', 'ssh_user', 'ssh_key_secret_name', 'hyperparameters')
                     )
WHERE type = 'training-launcher'
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

-- ───────────────────────────────────────────────────────────────────────────
-- 2. model-trainer: thread provisioning_id into call_launcher's input_mapping.
--    The provisioner returns provisioning_id in provisioning_result (verified
--    against provision_action.go ProvisionInstanceResult). Add the one mapping;
--    leave the existing fields (the adapter ignores ip/user/secret and resolves
--    them from provisioning_id, but they're harmless to pass).
-- ───────────────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_launcher,config,input_mapping,provisioning_id}',
        '"provisioning_result.provisioning_id"'::jsonb,
        true  -- create_missing
                     )
WHERE type = 'model-trainer'
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

COMMIT;

-- ───────────────────────────────────────────────────────────────────────────
-- Post-apply verification (run manually):
-- ───────────────────────────────────────────────────────────────────────────
-- \echo launcher workflow:
-- SELECT jsonb_pretty(default_config) FROM agent_definitions WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
-- \echo launcher contract:
-- SELECT jsonb_pretty(input_contract) FROM agent_definitions WHERE type='training-launcher' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
-- \echo model-trainer call_launcher mapping (expect provisioning_id present):
-- SELECT jsonb_pretty(default_config->'workflow'->'steps'->'call_launcher'->'config'->'input_mapping') FROM agent_definitions WHERE type='model-trainer' AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
