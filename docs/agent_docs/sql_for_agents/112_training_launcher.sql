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
