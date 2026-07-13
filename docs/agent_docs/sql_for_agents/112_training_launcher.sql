-- ============================================================================
-- 023_training_launcher_stub.sql
-- ============================================================================
-- STUB agent_definition for `training-launcher` to unblock end-to-end testing
-- of the model-trainer → training-data-preparer canonical spawn-and-call
-- path. The model-trainer's 7-step workflow spawns launcher BEFORE calling
-- data-preparer (spawn-before-call ordering), so without this row,
-- spawn_launcher errors with "agent definition: sql: no rows in result set"
-- and the orchestrator never reaches call_data_preparer.
--
-- This stub will be replaced by a real implementation in a future migration
-- that:
--   - SCPs training scripts to the GPU instance
--   - SCPs (or signals VM to download) the training dataset
--   - SSH-execs the training script in nohup background mode
--   - Updates training_runs.status to 'running'
--   - Returns {launched_at, launch_pid}
--
-- Step Zero check (per dev guide §0):
--   Searched agent_definitions for: launch, ssh, exec, training, run
--     - no existing agent does remote SSH execution. New agent required.
--   Searched actions for: ssh, exec, launch, train
--     - no existing action is a fit. Will write Go action in real impl.
--   Decision: stub for now, real impl later.
-- ============================================================================

INSERT INTO agent_definitions (
    type,
    display_name,
    description,
    category,
    agent_category,
    domain_tags,
    status,
    version,
    image_tag,
    default_config,
    health_config,
    env_vars,
    topics,
    input_contract,
    output_contract
) VALUES (
             'training-launcher',
             'Training Launcher',
             'STUB — SCPs scripts and dataset to the GPU instance, SSH-execs training in background, updates training_runs.status. Currently a no-op pass-through to unblock orchestrator testing; real implementation pending.',
             'specialist',
             'specialist',
             '["training", "ssh", "launcher", "stub"]'::jsonb,
             'experimental',
             1,
             'latest',
             '{
                 "workflow": {
                     "start_step": "complete",
                     "processing_mode": "task",
                     "timeout_seconds": 60,
                     "steps": {
                         "complete": {
                             "name": "",
                             "action": "complete_workflow",
                             "config": {"output_fields": []},
                             "description": "STUB — returns empty launch_result. Real implementation will SCP+SSH-exec the training script and return launched_at.",
                             "target_agent_type": ""
                         }
                     }
                 }
             }'::jsonb,
             '{
                 "port": 8080,
                 "liveness_path": "/health",
                 "readiness_path": "/ready",
                 "initial_delay_seconds": 30
             }'::jsonb,
             '[]'::jsonb,
             '{"error": "system.errors.training-launcher", "process": "system.agent.training-launcher.process", "response": "system.responses.training-launcher"}'::jsonb,
             '{"required": ["training_run_id"], "optional": ["instance_ip", "ssh_user", "ssh_key_secret_name", "dataset_uri", "hyperparameters"]}'::jsonb,
             '{"produces": ["launched_at", "launch_pid"]}'::jsonb
         );

-- Verify
SELECT type, category, agent_category, status,
       default_config->'workflow'->>'start_step' AS start_step,
    jsonb_object_keys(default_config->'workflow'->'steps') AS steps
FROM agent_definitions
WHERE type = 'training-launcher'
  AND deleted_at IS NULL;

---
--
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
--
-- ── CHANGES vs the first draft (each verified, not assumed) ──────────────────
--  (A) ssh_exec_launch command_template HARDENED. The earlier one-liner
--      backgrounded the whole && chain with a single trailing '&' but left the
--      backgrounded subshell's stdout/stderr attached to the SSH channel and
--      only redirected run.sh. Because thunder-adapter's ssh_exec captures
--      exit_code + stdout (i.e. it reads to channel EOF), the channel would NOT
--      close until run.sh exited 30-90 min later — so the "return quickly with
--      LAUNCH_PID" the design assumes would instead block past the adapter
--      command timeout AND the 600s launcher timeout, and a torn-down session
--      could SIGHUP the un-nohup'd curls. New form runs the whole chain under
--      `setsid` in a new session with stdin</dev/null and stdout/stderr to
--      /workspace/launch.log, so the channel hits EOF right after the `echo`
--      and the chain is immune to session teardown. run.sh keeps its own
--      train.log. {scripts_url}/{dataset_url} tokens are UNCHANGED, so the
--      action's interpolation is identical — only the shell structure changed.
--  (B) complete step uses output_fields (plural array) — the key
--      complete_workflow actually reads (matches model-trainer's own complete
--      step and the stub's empty "output_fields": []). The earlier draft used
--      output_field (singular), which complete_workflow ignores, surfacing
--      nothing.
--  (C) Both UPDATEs scoped with `is_active = true` in addition to the
--      snapshot/deleted guards. There is currently exactly one live
--      training-launcher row (version 1, is_active=t), so this is a no-op
--      today; it stops a future re-run from clobbering the config of an
--      inactive prior version, since (type, version) is unique and multiple
--      versions can coexist.

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
--   ssh_exec_launch.launch_result       (curl both URLs, setsid run.sh, echo PID)
--   mark_running.mark_result            (training_runs -> running)
--
-- NOTE on the ssh_exec command: it is a SINGLE LINE (no literal newlines) so it
-- survives JSON + Kafka transport cleanly (see debugging guide §9 heredoc trap).
-- The complex chain is in run.sh, not here.

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
                    -- input_mapping. Key MUST match UPLOAD_bundle.sh's upload target.
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

                    -- ── Step 3: ssh_exec — fetch both, detach run.sh, capture PID ──
                    -- command_template is a CONSTANT (config). The action substitutes the
                    -- {scripts_url}/{dataset_url} tokens from input_data.scripts_url /
                    -- input_data.dataset_url, which input_mapping fills from the two presign
                    -- steps' presigned_url outputs. provisioning_id (runtime) comes from the
                    -- parent (model-trainer threads provisioning_result.provisioning_id).
                    --
                    -- HARDENED (change A): the whole fetch+launch chain runs under setsid in
                    -- a new session, stdin</dev/null, stdout/stderr -> /workspace/launch.log.
                    -- This detaches it from the SSH channel so the adapter's ssh_exec returns
                    -- as soon as `echo` prints LAUNCH_PID, instead of blocking until run.sh
                    -- finishes. curl/tar failures are captured in launch.log (the old form
                    -- discarded them). A bucket/key mismatch now fails fast and visibly at the
                    -- curl (-f → non-zero → chain stops) and is diagnosable in launch.log.
                    -- Tokens are double-quoted; B2 V4 presigned URLs are percent-encoded
                    -- (no raw single quotes), so they sit safely inside the single-quoted
                    -- `bash -c` body.
                        'ssh_exec_launch', jsonb_build_object(
                                'action', 'dispatch_thunder_ssh_exec',
                                'description', 'SSH to the instance (by provisioning_id): curl the presigned scripts bundle and dataset, then setsid run.sh (setup->smoke->full) fully detached. Returns quickly with LAUNCH_PID in stdout.',
                                'config', jsonb_build_object(
                                        'command_template', 'mkdir -p /workspace; setsid bash -c ''curl -fsSL "{scripts_url}" -o /workspace/bundle.tar.gz && tar -xzf /workspace/bundle.tar.gz -C /workspace && curl -fsSL "{dataset_url}" -o /workspace/training_iter0.jsonl && chmod +x /workspace/run.sh && /workspace/run.sh > /workspace/train.log 2>&1'' < /dev/null > /workspace/launch.log 2>&1 & echo "LAUNCH_PID=$!"',
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

                    -- ── Step 5: complete — surface launch_result ──
                    -- complete_workflow reads output_fields (PLURAL ARRAY), not output_field.
                    -- (change B) — matches model-trainer's own complete step.
                        'complete', jsonb_build_object(
                                'action', 'complete_workflow',
                                'description', 'Return launch_result so the parent sees the ssh_exec reply (exit_code/stdout incl. LAUNCH_PID).',
                                'config', jsonb_build_object(
                                        'output_fields', jsonb_build_array('launch_result')
                                          )
                                    )

                         )
                    )
                     )
WHERE type = 'training-launcher'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

-- ───────────────────────────────────────────────────────────────────────────
-- 1b. Extend the launcher input_contract: add provisioning_id (required now).
--     Keep training_run_id required; dataset_uri moves to required (the launcher
--     needs it to presign). model-trainer's call_launcher already maps all three
--     (training_run_id, dataset_uri from preparation_result; provisioning_id
--     added in section 2 below), plus the optionals.
-- ───────────────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET input_contract = jsonb_build_object(
        'required', jsonb_build_array('training_run_id', 'provisioning_id', 'dataset_uri'),
        'optional', jsonb_build_array('instance_ip', 'ssh_user', 'ssh_key_secret_name', 'hyperparameters')
                     )
WHERE type = 'training-launcher'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

-- ───────────────────────────────────────────────────────────────────────────
-- 2. model-trainer: thread provisioning_id into call_launcher's input_mapping.
--    Verified against the live model-trainer row: call_launcher.config.input_mapping
--    already exists as an object and already maps training_run_id + dataset_uri
--    (from preparation_result) and the ssh_* optionals (from provisioning_result),
--    so jsonb_set with create_missing lands the new provisioning_id key.
--    gpu-provisioner returns provisioning_id in provisioning_result
--    (ProvisionInstanceResult), so the reference resolves.
-- ───────────────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_launcher,config,input_mapping,provisioning_id}',
        '"provisioning_result.provisioning_id"'::jsonb,
        true  -- create_missing (parent input_mapping object confirmed present)
                     )
WHERE type = 'model-trainer'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

COMMIT;

-- ───────────────────────────────────────────────────────────────────────────
-- Post-apply verification (run manually).
-- ───────────────────────────────────────────────────────────────────────────
-- A. Exactly one live launcher row was touched (expect 1):
-- SELECT count(*) FROM agent_definitions
--  WHERE type='training-launcher' AND is_active=true
--    AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
--
-- B. Launcher workflow + contract:
-- SELECT jsonb_pretty(default_config) FROM agent_definitions WHERE type='training-launcher' AND is_active=true AND deleted_at IS NULL;
-- SELECT jsonb_pretty(input_contract)  FROM agent_definitions WHERE type='training-launcher' AND is_active=true AND deleted_at IS NULL;
--
-- C. model-trainer call_launcher mapping now contains provisioning_id (expect the line):
-- SELECT jsonb_pretty(default_config->'workflow'->'steps'->'call_launcher'->'config'->'input_mapping')
--   FROM agent_definitions WHERE type='model-trainer' AND is_active=true AND deleted_at IS NULL;
--
-- ── Two adapter/preparer-side preconditions this workflow assumes (confirm
--    BEFORE the first run; neither is fixable in this migration):
--
-- D. prepare_object_url reply key. input_mapping reads
--    scripts_url_result.presigned_url / dataset_url_result.presigned_url, so the
--    adapter's handlePrepareObjectURL MUST reply with the exact key "presigned_url".
--    Confirm in internal/adapters/thunder/data_url_actions.go (the JSON field), or
--    by a one-off prepare_object_url round-trip and inspecting collected_data.
--
-- E. Same-bucket invariant. keyFromS3URI strips s3://<bucket>/ and the adapter
--    presigns the bare key against ITS configured bucket. So the preparer's
--    dataset upload, the bundle upload (personae-model-training, per
--    UPLOAD_bundle.sh), and the adapter's presign bucket must all be the SAME
--    bucket. Check the preparer's S3 upload bucket vs the adapter's presign
--    bucket config. (If they differ, the VM curl 404s and stops — now visible in
--    /workspace/launch.log rather than silent, thanks to change A + curl -f.)

---

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
-- ── INPUT RESOLUTION (important — this is NOT the call_agent pattern) ────────
-- These are LOCAL action steps, not call_agent. The coordinator resolves
-- output_mapping for local steps but NOT input_mapping (input_mapping is honoured
-- only by call_agent, which builds a CHILD's input_data, and by loop fan-out).
-- So local steps must NOT use input_mapping here — it would be dead config.
-- Instead:
--   * literals (method, key, command_template) are plain config values, read
--     directly from config by the action;
--   * a value already in input_data (provisioning_id, dataset_uri — threaded in
--     by model-trainer's call_launcher) is read by the action from input_data,
--     no config needed;
--   * a CROSS-STEP value (a prior step's output, e.g. a presigned_url) is a
--     plain config key whose value is a dot-path; the action resolves it from
--     collected_data (ssh_exec command tokens via resolveTemplateToken;
--     mark_running via ExtractActionInputs Strategy-0). Same way every other
--     local action pulls from collected_data — no chassis change, no input_mapping.
-- Requires the patched dispatch actions:
--   dispatch_thunder_prepare_object_url — falls back to input_data.dataset_uri
--   dispatch_thunder_ssh_exec           — resolveTemplateToken resolves {tokens}
--                                         from config dot-paths then input_data
--
-- Surgical UPDATEs (not a re-INSERT): the stub row already has the right
-- type/category/image_tag/output_contract; we only replace default_config and
-- extend input_contract (add provisioning_id). Mirrors 101_switch_to_haiku.sql.
-- Idempotent. Scoped to the single active row (is_active = true).

BEGIN;

-- ───────────────────────────────────────────────────────────────────────────
-- 1. training-launcher: replace stub default_config with the real workflow.
-- ───────────────────────────────────────────────────────────────────────────
--   presign_dataset -> presign_scripts -> ssh_exec_launch -> mark_running -> complete
--
-- ssh_exec command is a SINGLE LINE (no literal newlines) so it survives JSON +
-- Kafka transport cleanly (debugging guide §9 heredoc trap). The whole
-- fetch+launch chain runs under setsid in a new session, stdin</dev/null and
-- stdout/stderr -> /workspace/launch.log, so the SSH channel hits EOF right
-- after `echo` (the adapter's ssh_exec returns quickly instead of blocking for
-- the whole train) and the chain survives session teardown. run.sh keeps its
-- own train.log. {scripts_url}/{dataset_url} are command_template tokens, double
-- quoted; B2 V4 presigned URLs are percent-encoded so they sit safely inside the
-- single-quoted bash -c body.

UPDATE agent_definitions
SET default_config = jsonb_build_object(
        'workflow', jsonb_build_object(
                'processing_mode', 'task',
                'timeout_seconds', 600,
                'start_step', 'presign_dataset',
                'steps', jsonb_build_object(

                    -- ── Step 1: presign the dataset (GET) ──
                    -- No key/s3_uri given → the action falls back to the
                    -- preparer's dataset_uri in input_data (threaded by
                    -- call_launcher) and strips it to a bucket-relative key.
                    -- method is the only config value.
                        'presign_dataset', jsonb_build_object(
                                'action', 'dispatch_thunder_prepare_object_url',
                                'description', 'Presign a GET URL for the training dataset. The action derives the key from input_data.dataset_uri (the preparer''s s3:// URI). method is a config literal.',
                                'config', jsonb_build_object(
                                        'method', 'GET'
                                          ),
                                'output_field', 'dataset_url_result',
                                'next_step', 'presign_scripts'
                                           ),

                    -- ── Step 2: presign the scripts bundle (GET) by literal key ──
                    -- Both key and method are config literals.
                        'presign_scripts', jsonb_build_object(
                                'action', 'dispatch_thunder_prepare_object_url',
                                'description', 'Presign a GET URL for the scripts bundle (run.sh + setup + train). Literal key constant in config.',
                                'config', jsonb_build_object(
                                        'key', 'finetuning/scripts/bundle.tar.gz',
                                        'method', 'GET'
                                          ),
                                'output_field', 'scripts_url_result',
                                'next_step', 'ssh_exec_launch'
                                           ),

                    -- ── Step 3: ssh_exec — fetch both, detach run.sh, capture PID ──
                    -- provisioning_id is read by the action from input_data (the
                    -- adapter resolves ip/port/user/key from it). The command's
                    -- {scripts_url}/{dataset_url} tokens are resolved by
                    -- resolveTemplateToken from the config dot-paths below
                    -- (prior steps' presigned_url outputs in collected_data).
                        'ssh_exec_launch', jsonb_build_object(
                                'action', 'dispatch_thunder_ssh_exec',
                                'description', 'SSH to the instance (by provisioning_id from input_data): curl the presigned scripts bundle and dataset, then setsid run.sh (setup->smoke->full) fully detached. Returns quickly with LAUNCH_PID in stdout.',
                                'config', jsonb_build_object(
                                        'command_template', 'mkdir -p /workspace; setsid bash -c ''curl -fsSL "{scripts_url}" -o /workspace/bundle.tar.gz && tar -xzf /workspace/bundle.tar.gz -C /workspace && curl -fsSL "{dataset_url}" -o /workspace/training_iter0.jsonl && chmod +x /workspace/run.sh && /workspace/run.sh > /workspace/train.log 2>&1'' < /dev/null > /workspace/launch.log 2>&1 & echo "LAUNCH_PID=$!"',
                                        'scripts_url', 'scripts_url_result.presigned_url',
                                        'dataset_url', 'dataset_url_result.presigned_url'
                                          ),
                                'output_field', 'launch_result',
                                'next_step', 'mark_running'
                                           ),

                    -- ── Step 4: flip training_runs pending -> running ──
                    -- training_run_id is a config dot-path resolved by
                    -- ExtractActionInputs Strategy-0 (the canonical local-action
                    -- way), not input_mapping.
                        'mark_running', jsonb_build_object(
                                'action', 'mark_training_run_running',
                                'description', 'Transition training_runs to running and stamp started_at now that the process is launched.',
                                'config', jsonb_build_object(
                                        'training_run_id', 'input_data.training_run_id'
                                          ),
                                'output_field', 'mark_result',
                                'next_step', 'complete'
                                        ),

                    -- ── Step 5: complete — surface launch_result ──
                    -- complete_workflow reads output_fields (PLURAL ARRAY).
                        'complete', jsonb_build_object(
                                'action', 'complete_workflow',
                                'description', 'Return launch_result so the parent sees the ssh_exec reply (exit_code/stdout incl. LAUNCH_PID).',
                                'config', jsonb_build_object(
                                        'output_fields', jsonb_build_array('launch_result')
                                          )
                                    )

                         )
                    )
                     )
WHERE type = 'training-launcher'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

-- ───────────────────────────────────────────────────────────────────────────
-- 1b. Extend the launcher input_contract: add provisioning_id (required now).
--     dataset_uri moves to required (the launcher needs it to presign).
--     model-trainer's call_launcher already maps training_run_id + dataset_uri
--     (from preparation_result); provisioning_id added in section 2.
-- ───────────────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET input_contract = jsonb_build_object(
        'required', jsonb_build_array('training_run_id', 'provisioning_id', 'dataset_uri'),
        'optional', jsonb_build_array('instance_ip', 'ssh_user', 'ssh_key_secret_name', 'hyperparameters')
                     )
WHERE type = 'training-launcher'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

-- ───────────────────────────────────────────────────────────────────────────
-- 2. model-trainer: thread provisioning_id into call_launcher's input_mapping.
--    This IS a call_agent step, so input_mapping is the correct, working
--    mechanism (call_agent resolves it to build the launcher's input_data).
--    call_launcher.config.input_mapping already exists and already maps
--    training_run_id + dataset_uri (from preparation_result), so jsonb_set with
--    create_missing lands the new provisioning_id key. gpu-provisioner returns
--    provisioning_id in provisioning_result (ProvisionInstanceResult).
-- ───────────────────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_launcher,config,input_mapping,provisioning_id}',
        '"provisioning_result.provisioning_id"'::jsonb,
        true  -- create_missing (parent input_mapping object confirmed present)
                     )
WHERE type = 'model-trainer'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND deleted_at IS NULL;

COMMIT;

-- ───────────────────────────────────────────────────────────────────────────
-- Post-apply verification (run manually).
-- ───────────────────────────────────────────────────────────────────────────
-- A. One live launcher row touched (expect 1):
-- SELECT count(*) FROM agent_definitions
--  WHERE type='training-launcher' AND is_active=true
--    AND (is_snapshot IS NULL OR is_snapshot=false) AND deleted_at IS NULL;
--
-- B. Launcher workflow + contract:
-- SELECT jsonb_pretty(default_config) FROM agent_definitions WHERE type='training-launcher' AND is_active=true AND deleted_at IS NULL;
-- SELECT jsonb_pretty(input_contract)  FROM agent_definitions WHERE type='training-launcher' AND is_active=true AND deleted_at IS NULL;
--
-- C. model-trainer call_launcher mapping now contains provisioning_id:
-- SELECT jsonb_pretty(default_config->'workflow'->'steps'->'call_launcher'->'config'->'input_mapping')
--   FROM agent_definitions WHERE type='model-trainer' AND is_active=true AND deleted_at IS NULL;
--
-- ── Adapter/preparer-side preconditions (confirm BEFORE the first run; not
--    fixable here):
-- D. The DEPLOYED thunder-adapter must implement the prepare_object_url action
--    (the Phase-4 adapter.go copy seen so far has prepare_dataset_url /
--    prepare_artefact_url but no prepare_object_url → would return
--    not_implemented). Confirm the Phase-5 handler + dispatch case are deployed.
-- E. handlePrepareObjectURL's reply MUST use the key "presigned_url" (the
--    ssh_exec_launch tokens resolve scripts_url_result.presigned_url /
--    dataset_url_result.presigned_url). Confirm in data_url_actions.go.
-- F. Same-bucket: adapter presigns against personae-model-training (confirmed,
--    adapter.go default). Check that input_data.dataset_uri's key (after
--    keyFromS3URI strips s3://<bucket>/) is the real object key under that
--    bucket — the preparer's s3_bucket="finetuning" is noted stale/logical.