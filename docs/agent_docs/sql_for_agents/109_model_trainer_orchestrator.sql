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

