-- ============================================================================
-- 020_model_trainer_orchestrator.sql
-- ============================================================================
-- The model-trainer orchestrator owns the KICKOFF phase of a training run:
--   1. Worker `training-data-preparer` exports JSONL to S3 + INSERTs the
--      training_runs row in 'pending' state.
--   2. Worker `gpu-provisioner` calls Thunder API, sets up the VM,
--      installs deps, returns SSH connection info. Updates row to 'running'.
--   3. Worker `training-launcher` SCPs scripts/dataset, SSH-execs the
--      training script with nohup, returns immediately with the pid.
--   Workflow exits. Run is now 'running' on the VM.
--
-- The COMPLETION phase (poll for status, collect adapter, decommission VM)
-- is handled separately by the training-monitor scheduled task (migration
-- 027/028) so the orchestrator doesn't have to hold an open workflow for
-- the ~9 hours a training run takes.
--
-- Per chassis convention, all spawn_agent steps precede all call_agent steps
-- so target_role lookups succeed.
--
-- Workflow inputs (CollectedData["input_data"]):
--   export_id          UUID    — required, references training_exports.runs(id)
--   hyperparameters    JSONB   — required, full reproducibility set
--                                  {base_model, epochs, batch, grad_accum, lr,
--                                   lora_r, lora_alpha, max_seq, seed}
--   triggered_by       UUID    — optional, agent that initiated
--   orchestration_id   UUID    — optional, parent orchestration
--
-- Workflow outputs (final CollectedData):
--   preparation_result.training_run_id    UUID — the new model_lifecycle.training_runs row
--   preparation_result.dataset_uri        TEXT — s3://finetuning/datasets/{export_id}/training.jsonl
--   provisioning_result.thunder_instance_id   TEXT
--   provisioning_result.instance_ip           TEXT
--   provisioning_result.ssh_user              TEXT
--   provisioning_result.ssh_key_secret_name   TEXT — k8s secret holding the private key
--   launch_result.pid                     INT
--   launch_result.started_at              TIMESTAMPTZ
-- ============================================================================

INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config
)
VALUES (
           gen_random_uuid(),
           'model-trainer',
           'Model Trainer Orchestrator',
           'Orchestrates the QLoRA training kickoff: prepare data, provision GPU, launch training. The training-monitor scheduled task handles polling and completion.',
           'orchestrator',
           '{
             "workflow": {
               "start_step": "spawn_data_preparer",
               "steps": {
                 "spawn_data_preparer": {
                   "action": "spawn_agent",
                   "config": {
                     "role": "data_preparer",
                     "agent_type": "training-data-preparer"
                   },
                   "output_field": "data_preparer_agent",
                   "next_step": "spawn_provisioner",
                   "description": "Spawn worker that exports training JSONL to S3 and inserts the training_runs row"
                 },
                 "spawn_provisioner": {
                   "action": "spawn_agent",
                   "config": {
                     "role": "provisioner",
                     "agent_type": "gpu-provisioner"
                   },
                   "output_field": "provisioner_agent",
                   "next_step": "spawn_launcher",
                   "description": "Spawn worker that provisions a Thunder Compute A100 instance"
                 },
                 "spawn_launcher": {
                   "action": "spawn_agent",
                   "config": {
                     "role": "launcher",
                     "agent_type": "training-launcher"
                   },
                   "output_field": "launcher_agent",
                   "next_step": "call_data_preparer",
                   "description": "Spawn worker that launches the training script over SSH"
                 },
                 "call_data_preparer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "training-data-preparer",
                     "target_role": "data_preparer",
                     "input_mapping": {
                       "export_id": "input_data.export_id",
                       "hyperparameters": "input_data.hyperparameters",
                       "triggered_by": "input_data.triggered_by",
                       "orchestration_id": "input_data.orchestration_id"
                     },
                     "timeout_seconds": 300
                   },
                   "output_field": "preparation_result",
                   "next_step": "call_provisioner",
                   "description": "Stream training_exports.rows to S3 JSONL; INSERT model_lifecycle.training_runs in pending"
                 },
                 "call_provisioner": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "gpu-provisioner",
                     "target_role": "provisioner",
                     "input_mapping": {
                       "training_run_id": "preparation_result.training_run_id",
                       "hyperparameters": "input_data.hyperparameters"
                     },
                     "timeout_seconds": 900
                   },
                   "output_field": "provisioning_result",
                   "next_step": "call_launcher",
                   "description": "Provision Thunder A100 instance; update training_runs.thunder_instance_id"
                 },
                 "call_launcher": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "training-launcher",
                     "target_role": "launcher",
                     "input_mapping": {
                       "training_run_id": "preparation_result.training_run_id",
                       "dataset_uri": "preparation_result.dataset_uri",
                       "instance_ip": "provisioning_result.instance_ip",
                       "ssh_user": "provisioning_result.ssh_user",
                       "ssh_key_secret_name": "provisioning_result.ssh_key_secret_name",
                       "hyperparameters": "input_data.hyperparameters"
                     },
                     "timeout_seconds": 600
                   },
                   "output_field": "launch_result",
                   "next_step": "complete",
                   "description": "SCP scripts and dataset; SSH-exec backgrounded training; update training_runs to running"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "config": {
                     "output_fields": [
                       "preparation_result",
                       "provisioning_result",
                       "launch_result"
                     ]
                   },
                   "description": "Kickoff complete; training_runs.status = running. Polling phase takes over."
                 }
               }
             }
           }'::jsonb,
           true,
           '["training", "orchestrator", "qlora", "thunder-compute"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'latest',
           '{"requests": {"cpu": "100m", "memory": "256Mi"}, "limits": {"cpu": "500m", "memory": "512Mi"}}'::jsonb,
           '{"requests": "system.agent.model-trainer.requests", "responses": "system.agent.model-trainer.responses"}'::jsonb,
           '{"liveness_initial_delay": 30, "liveness_period": 30}'::jsonb
       );


-- Verification
SELECT type, category, is_active,
       jsonb_array_length(jsonb_object_keys(default_config->'workflow'->'steps')::jsonb[])
           AS step_count_should_be_7
FROM agent_definitions
WHERE type = 'model-trainer'
  AND deleted_at IS NULL;

-- Show the workflow shape
SELECT jsonb_pretty(default_config->'workflow') AS workflow
FROM agent_definitions
WHERE type = 'model-trainer'
  AND deleted_at IS NULL;
--

UPDATE agent_definitions SET agent_category = 'coordinator', domain_tags = '["training", "orchestrator", "qlora", "thunder-compute"]'::jsonb WHERE type = 'model-trainer' AND deleted_at IS NULL;

--
-- 023_training_agents_health_config_fix.sql
--
-- Patches health_config on training agents to include the required `port`
-- field. Without it, k8s Job creation fails with:
--   spec.template.spec.containers[0].ports[0].containerPort: Required value
--
-- The chassis spawn_agent code reads health_config.port and uses it as
-- the container's containerPort. Existing agents (e.g. defaults) use 8080.

UPDATE agent_definitions
SET health_config = '{
    "port": 8080,
    "liveness_path": "/health",
    "readiness_path": "/ready",
    "initial_delay_seconds": 30
}'::jsonb,
    updated_at = NOW()
WHERE type IN ('model-trainer', 'training-data-preparer')
  AND deleted_at IS NULL;

-- Verify
SELECT type, health_config
FROM agent_definitions
WHERE type IN ('model-trainer', 'training-data-preparer')
  AND deleted_at IS NULL
ORDER BY type;

---
-- update model trainer model-trainer
-- backup
clients_db=# SELECT snapshot_agent('model-trainer',
                                   'thunder orchestration id input mappings optional');
NOTICE:  Snapshot captured: type=model-trainer, source_version=1, source_id=94f5a069-6fb5-4aba-81e5-4fcc9220ed30, reason=thunder orchestration id input mappings optional
            snapshot_agent
--------------------------------------
 94f5a069-6fb5-4aba-81e5-4fcc9220ed30
(1 row)
--
-- 103_call_data_preparer_optional_inputs.sql
--
-- WHAT THIS FIXES
-- The model-trainer orchestration chain dies at its first call step,
-- call_data_preparer, with:
--   input_mapping failed: source path 'input_data.orchestration_id' not found
--                          for field 'orchestration_id'
-- (coordinator.go:1534; content_search.go:96 part=orchestration_id
--  available_keys=[hyperparameters, export_id]). Confirmed live in orchestration
-- 23863e2e-3090-4231-980c-6d73746b60e3 (2026-06-02): all three spawns resolved,
-- then call_data_preparer (step 1cfdb238) failed BEFORE dispatching any work
-- request, so the preparer never ran and no training_runs row was inserted.
--
-- WHY
-- call_data_preparer is a call_agent step. call_agent builds the child's
-- input_data via input_contracts.ResolveInputMapping, which HARD-FAILS when a
-- mapped source path is absent UNLESS the destination field name ends in '?'
-- (input_mapping.go L101-128). The step maps four fields:
--     export_id        <- input_data.export_id        (present)
--     hyperparameters  <- input_data.hyperparameters  (present)
--     orchestration_id <- input_data.orchestration_id (ABSENT on manual/most triggers)
--     triggered_by     <- input_data.triggered_by     (ABSENT)
-- with NO '?' markers, so the two absent sources are treated as required and the
-- whole step fails before any work request reaches the data-preparer.
--
-- The preparer ITSELF already treats these two as OPTIONAL
-- (PrepareTrainingDataInputSpec.Optional = [triggered_by, orchestration_id];
-- prepare_training_data_action.go), reads them through NullableString, and
-- training_runs.triggered_by / orchestration_id are nullable UUIDs
-- (019_model_lifecycle_schema.sql). So the workflow contract over-declared
-- required-ness relative to the action contract. This aligns the two.
--
-- THE FIX
-- Mark the two mappings optional with the '?' destination-field suffix — the
-- same convention call_launcher already uses (instance_ip?, ssh_user?,
-- ssh_key_secret_name?). When the source is absent, ResolveInputMapping skips
-- the field; the '?' is stripped (TrimSuffix) before the field is placed in the
-- child input_data, so the preparer still reads them by their bare names and
-- writes NULL when not supplied. export_id + hyperparameters are unchanged and
-- remain required.
--
-- VARIABLE-NAME NOTE (intentional rename): this renames two input_mapping KEYS
-- (orchestration_id -> orchestration_id?, triggered_by -> triggered_by?). The
-- child-visible field names are UNCHANGED — '?' is a resolver-only marker that
-- is trimmed off. No Go variable names change; the preparer keeps reading
-- input_data.export_id / .hyperparameters / .triggered_by / .orchestration_id,
-- and ExtractActionInputs (spec-driven) treats the latter two as optional.
--
-- SCOPE / MECHANISM
-- The chain lives in agent_definitions.default_config -> 'workflow' (verified:
-- call_data_preparer appears only in default_config, column 6). In-place
-- jsonb_set on the exact definition that executed (id 94f5a069-...), no version
-- bump, so a stable definition id is picked up by the next orchestrate.
-- CAVEAT: if the chassis caches agent_definitions in memory, a rollout of
-- agent-chassis may be needed for the change to take effect — confirm on the
-- next run rather than assuming the edit is hot.
--
-- PROVENANCE NOTE: with these optional and not supplied by a manual trigger,
-- training_runs.orchestration_id / triggered_by will be NULL for such runs.
-- Acceptable for iter_0. A follow-up (NOT in this migration) could have the
-- preparer default orchestration_id from its parent_orchestration_id.

BEGIN;

-- Before (expect orchestration_id / triggered_by with no '?'):
SELECT id, type, version, is_active,
       default_config #> '{workflow,steps,call_data_preparer,config,input_mapping}'
         AS before_input_mapping
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

UPDATE public.agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_data_preparer,config,input_mapping}',
        jsonb_build_object(
                'export_id',         'input_data.export_id',
                'hyperparameters',   'input_data.hyperparameters',
                'orchestration_id?', 'input_data.orchestration_id',
                'triggered_by?',     'input_data.triggered_by'
        ),
        false  -- do not create the path if missing; fail loud instead
                     ),
    updated_at = NOW()
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30'
  AND type = 'model-trainer';

-- After (expect orchestration_id? and triggered_by? keys):
SELECT id, type, version, is_active,
       default_config #> '{workflow,steps,call_data_preparer,config,input_mapping}'
         AS after_input_mapping
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

COMMIT;

---

-- 104_provisioner_output_fields_and_launcher_mapping.sql
--
-- APPLY TO: clients_db (NOT templates_db). The flywheel-C agent_definitions
-- (rich schema with version/is_snapshot) are read by the chassis from clients_db;
-- templates_db holds only the old website-builder catalog.
--   <your clients_db psql> -f 104_provisioner_output_fields_and_launcher_mapping.sql
--
-- WHAT / WHY
-- Two in-place edits, no version bump (chassis loads definitions per-orchestrate,
-- so no restart needed):
--
--   1) gpu-provisioner (0bf9fa8a) complete step: replace the NON-STANDARD singular
--      `output_field: provision_response` with the standard `output_fields:
--      ["dispatch_provision"]`. extractWorkflowResult only honours the plural
--      `output_fields`; the singular key was silently ignored, dropping the agent
--      into the fallback dump ({dispatch_provision, input_data, ...}). The provision
--      result lands under the STEP NAME `dispatch_provision` (await storage keys by
--      step name; `provision_response` is never a collected key), so that is the
--      field we surface. This also drops the stray input_data echo.
--
--   2) model-trainer (94f5a069) call_launcher input_mapping: re-point the four
--      provisioning fields from `provisioning_result.<field>` to
--      `provisioning_result.dispatch_provision.<field>`. These resolve via the same
--      `.response` auto-unwrap that already makes `preparation_result.dataset_uri`
--      work (dispatch_provision is the immediate child once provisioning_result.response
--      is unwrapped). dataset_uri / training_run_id / hyperparameters are unchanged.
--
-- NOTE: the gpu-provisioner dispatch step keeps `output_field: provision_response`
-- (cosmetic for await results; the result keys by step name regardless). Left as-is
-- to keep this change minimal.

BEGIN;

-- ── BEFORE ──
SELECT 'gpu-provisioner complete.config BEFORE' AS label,
       default_config #> '{workflow,steps,complete,config}' AS value
FROM public.agent_definitions
WHERE id = '0bf9fa8a-925c-4ab5-9287-2c8e5d7b9451';

SELECT 'model-trainer call_launcher.input_mapping BEFORE' AS label,
       default_config #> '{workflow,steps,call_launcher,config,input_mapping}' AS value
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

-- ── 1) gpu-provisioner: singular output_field → standard output_fields ──
UPDATE public.agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,complete,config}',
        '{"output_fields": ["dispatch_provision"]}'::jsonb,
        false
                     ),
    updated_at = now()
WHERE id = '0bf9fa8a-925c-4ab5-9287-2c8e5d7b9451'
  AND default_config #> '{workflow,steps,complete,config}' IS NOT NULL;

-- ── 2) model-trainer: re-point call_launcher provisioning fields ──
UPDATE public.agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_launcher,config,input_mapping}',
        '{
            "dataset_uri": "preparation_result.dataset_uri",
            "training_run_id": "preparation_result.training_run_id",
            "hyperparameters": "input_data.hyperparameters",
            "provisioning_id": "provisioning_result.dispatch_provision.provisioning_id",
            "instance_ip?": "provisioning_result.dispatch_provision.instance_ip",
            "ssh_user?": "provisioning_result.dispatch_provision.ssh_user",
            "ssh_key_secret_name?": "provisioning_result.dispatch_provision.ssh_key_secret_name"
        }'::jsonb,
        false
                     ),
    updated_at = now()
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30'
  AND default_config #> '{workflow,steps,call_launcher,config,input_mapping}' IS NOT NULL;

-- ── AFTER ──
SELECT 'gpu-provisioner complete.config AFTER' AS label,
       default_config #> '{workflow,steps,complete,config}' AS value
FROM public.agent_definitions
WHERE id = '0bf9fa8a-925c-4ab5-9287-2c8e5d7b9451';

SELECT 'model-trainer call_launcher.input_mapping AFTER' AS label,
       default_config #> '{workflow,steps,call_launcher,config,input_mapping}' AS value
FROM public.agent_definitions
WHERE id = '94f5a069-6fb5-4aba-81e5-4fcc9220ed30';

COMMIT;

--

-- 105_launcher_workspace_sudo_mkdir.sql
-- DB: clients_db  (the live flywheel-C agent_definitions — NOT templates_db)
--
-- Problem (observed live 2026-06-03, orch cd906623 / launcher orch b002b359):
--   The full iter_0 chain ran end-to-end for the first time after 104. The
--   training-launcher's ssh_exec_launch connected, returned exit_code 0 and
--   LAUNCH_PID=193 — but training never started. stderr was:
--       mkdir: cannot create directory '/workspace': Permission denied
--       bash: line 1: /workspace/launch.log: No such file or directory
--
--   The command_template did a *plain* `mkdir -p /workspace`. The `ubuntu` ssh
--   user cannot create a dir at `/`, so the mkdir failed; the curls/run.sh had
--   nowhere to land, and the `&`-backgrounded setsid job died immediately. The
--   exit_code is 0 only because the command's last token is `echo` (the known
--   detached-ssh_exec false-success: a VM-side failure looks like success).
--
-- Why this fix (and why it is NOT a patch):
--   The bundle's OWN setup script already establishes the convention —
--   00_vm_setup.sh L51-52:
--       sudo mkdir -p "${WORKSPACE}"
--       sudo chown "$(id -u):$(id -g)" "${WORKSPACE}"
--   i.e. /workspace is meant to be created with sudo and chowned to the caller.
--   The launcher's command_template diverged (plain mkdir) AND runs the curls
--   before 00_vm_setup.sh executes. This migration makes the command_template
--   mirror the script: create + chown /workspace up front, as the running user.
--   sudo is known-good on these Thunder instances (the whole setup script uses
--   it) and /workspace on the root volume has the 100GB the prior manual run
--   used — so no re-bundle is needed: run.sh and 00_vm_setup.sh keep /workspace
--   (its sudo-mkdir simply becomes idempotent).
--
-- Shape change: in-place edit of one string. The chassis loads the def per
--   orchestrate (no restart needed). No version bump (consistent with 104).
--
-- Target def: training-launcher (active, non-snapshot). Verify the BEFORE/AFTER
--   SELECT shows exactly one row and the expected string.

BEGIN;

SELECT 'training-launcher ssh_exec_launch.command_template BEFORE' AS label,
       default_config #> '{workflow,steps,ssh_exec_launch,config,command_template}' AS value
FROM public.agent_definitions
WHERE type = 'training-launcher'
  AND is_active = true
  AND COALESCE(is_snapshot, false) = false;

UPDATE public.agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ssh_exec_launch,config,command_template}',
        to_jsonb($cmd$sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace; setsid bash -c 'curl -fsSL "{scripts_url}" -o /workspace/bundle.tar.gz && tar -xzf /workspace/bundle.tar.gz -C /workspace && curl -fsSL "{dataset_url}" -o /workspace/training_iter0.jsonl && chmod +x /workspace/run.sh && /workspace/run.sh > /workspace/train.log 2>&1' < /dev/null > /workspace/launch.log 2>&1 & echo "LAUNCH_PID=$!"$cmd$::text),
       false
   )
 WHERE type = 'training-launcher'
   AND is_active = true
   AND COALESCE(is_snapshot, false) = false;

SELECT 'training-launcher ssh_exec_launch.command_template AFTER' AS label,
       default_config #> '{workflow,steps,ssh_exec_launch,config,command_template}' AS value
FROM public.agent_definitions
WHERE type = 'training-launcher'
  AND is_active = true
  AND COALESCE(is_snapshot, false) = false;

-- Expect: UPDATE 1, and the AFTER value beginning with
--   sudo mkdir -p /workspace && sudo chown $(id -u):$(id -g) /workspace; setsid ...
-- If satisfied:
COMMIT;
-- else: ROLLBACK;