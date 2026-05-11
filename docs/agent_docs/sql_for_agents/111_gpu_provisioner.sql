-- ============================================================================
-- 022_gpu_provisioner_stub.sql
-- ============================================================================
-- STUB agent_definition for `gpu-provisioner` to unblock end-to-end testing
-- of the model-trainer → training-data-preparer canonical spawn-and-call
-- path. The model-trainer's 7-step workflow spawns provisioner BEFORE calling
-- data-preparer (spawn-before-call ordering), so without this row,
-- spawn_provisioner errors with "agent definition: sql: no rows in result
-- set" and the orchestrator never reaches call_data_preparer.
--
-- This stub will be replaced by a real implementation in a future migration
-- that:
--   - calls the Thunder Compute API to provision an A100 instance
--   - generates an SSH keypair, stores private key as k8s secret
--   - waits for instance ready
--   - returns {instance_ip, ssh_user, ssh_key_secret_name, thunder_instance_id}
--
-- For now this row exists purely so spawn_agent can find it and create a
-- Job pod. The pod runs an empty workflow, returns {}, and the orchestrator
-- moves on. The downstream call_launcher step will see an empty
-- provisioning_result map; missing fields resolve to nil in input_mapping.
--
-- Step Zero check (per dev guide §0):
--   Searched agent_definitions for: provision, gpu, vm, instance, thunder
--     - no existing agent provisions external compute. New agent required.
--   Searched actions for: provision, vm, ssh, gpu, instance
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
             'gpu-provisioner',
             'GPU Provisioner',
             'STUB — provisions a Thunder Compute A100 instance for training. Currently a no-op pass-through to unblock orchestrator testing; real implementation pending.',
             'specialist',
             'specialist',
             '["training", "gpu", "thunder-compute", "stub"]'::jsonb,
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
                             "description": "STUB — returns empty provisioning_result. Real implementation will return instance_ip, ssh_user, ssh_key_secret_name.",
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
             '{"error": "system.errors.gpu-provisioner", "process": "system.agent.gpu-provisioner.process", "response": "system.responses.gpu-provisioner"}'::jsonb,
             '{"required": ["training_run_id"], "optional": ["hyperparameters"]}'::jsonb,
             '{"produces": ["instance_ip", "ssh_user", "ssh_key_secret_name", "thunder_instance_id"]}'::jsonb
         );

-- Verify
SELECT type, category, agent_category, status,
       default_config->'workflow'->>'start_step' AS start_step,
    jsonb_object_keys(default_config->'workflow'->'steps') AS steps
FROM agent_definitions
WHERE type = 'gpu-provisioner'
  AND deleted_at IS NULL;

