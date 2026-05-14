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

---
--

-- ============================================================================
-- 029_gpu_provisioner_real_impl.sql
--
-- Phase 3.6: replace the gpu-provisioner stub (migration 022) with a real
-- implementation that dispatches provision_instance to thunder-adapter.
--
-- The agent's CALLER CONTRACT is unchanged:
--   Inputs read from input_data.*:
--     - training_run_id (optional, propagated to the thunder_instances row)
--     - hyperparameters (passed through; future use)
--     - gpu / num_gpus / vcpus / disk_size_gb / mode / template (all optional)
--   Response body shape (read by model-trainer's call_launcher):
--     - instance_ip, ssh_user, ssh_key_secret_name, provisioning_id,
--       thunder_identifier, provisioned_at
--
-- Migration is idempotent — re-applying re-UPDATEs the same row.
-- Reverting to the stub would require a follow-up migration with the old
-- mock workflow JSON. The original is in migration 022_gpu_provisioner_stub.sql.
--
-- Preconditions:
--   1. Migration 028 applied (registers thunder-reaper and validates the
--      dispatch action pattern).
--   2. Chassis binary contains the new DispatchThunderProvisionAction
--      registered as "dispatch_thunder_provision".
--   3. thunder-adapter v1.0.1013+ deployed (handles provision_instance).
-- ============================================================================

BEGIN;

-- Replace the workflow inside default_config.
-- Step structure: dispatch_provision → complete_workflow.
-- timeout_seconds on the dispatch step is 600 (10 min) — covers worst-case
-- WaitForRunning (adapter's internal deadline is 5 min) plus buffer for
-- transient Thunder API slowness. Outer workflow timeout is 660.

UPDATE agent_definitions
SET
    default_config = jsonb_build_object(
            'processing_mode', 'task',
            'workflow', jsonb_build_object(
                    'start_step', 'dispatch_provision',
                    'processing_mode', 'task',
                    'timeout_seconds', 660,
                    'steps', jsonb_build_object(
                            'dispatch_provision', jsonb_build_object(
                                    'action', 'dispatch_thunder_provision',
                                    'description', 'Publishes provision_instance to thunder-adapter '
                                        || 'and awaits the response with the running '
                                        || 'instance details. Reads training_run_id and '
                                        || 'any optional GPU/sizing overrides from '
                                        || 'input_data; adapter applies defaults for '
                                        || 'unspecified fields.',
                                    'config', jsonb_build_object(
                                            'output_field', 'provision_response',
                                            'timeout_seconds', 600
                                              ),
                                    'next_step', 'complete'
                                                  ),
                            'complete', jsonb_build_object(
                                    'action', 'complete_workflow',
                                    'description', 'Return the adapter response as this agent''s '
                                        || 'final result. The auto-unwrap in chassis '
                                        || 'input_mapping v2 lets callers read fields '
                                        || 'either as provisioning_result.instance_ip or '
                                        || 'provisioning_result.response.instance_ip.',
                                    'config', jsonb_build_object(
                                            'output_field', 'provision_response'
                                              )
                                        )
                             )
                        )
                     ),
    description = 'Provisions a Thunder Compute instance by dispatching '
        || 'provision_instance to thunder-adapter. Awaits the response '
        || 'containing instance_ip, ssh_user, ssh_key_secret_name, '
        || 'provisioning_id, thunder_identifier, provisioned_at. '
        || 'Replaces migration-022 stub.',
    status = 'experimental',
    updated_at = NOW()
WHERE type = 'gpu-provisioner'
  AND is_active = true
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- Sanity check — abort if no row was updated. Means the gpu-provisioner
-- agent_definition is missing or has an unexpected shape, which would mean
-- migration 022 wasn't applied or has been deleted.
DO $$
DECLARE
row_count INT;
BEGIN
SELECT COUNT(*) INTO row_count
FROM agent_definitions
WHERE type = 'gpu-provisioner' AND is_active = true;

IF row_count = 0 THEN
        RAISE EXCEPTION 'No active gpu-provisioner agent_definition found. '
                        'Apply migration 022_gpu_provisioner_stub.sql first.';
END IF;
END
$$;

COMMIT;


-- ============================================================================
-- Verification queries
-- ============================================================================

-- Confirm the workflow swap
-- SELECT type, status,
--        default_config->'workflow'->'start_step' AS start_step,
--        default_config->'workflow'->'steps'->'dispatch_provision'->>'action' AS step1_action,
--        default_config->'workflow'->'steps'->'complete'->>'action' AS step2_action,
--        updated_at
-- FROM agent_definitions
-- WHERE type = 'gpu-provisioner' AND is_active = true;
--
-- Expect:
--   gpu-provisioner | experimental | "dispatch_provision"
--     | "dispatch_thunder_provision" | "complete_workflow" | <now>


-- ============================================================================
-- Manual smoke test (after deploy, with daily cap lowered for safety)
-- ============================================================================
--
-- 1. CAP THE SPEND FIRST
--    UPDATE thunder_config SET daily_cap_usd = 5;
--
-- 2. Send a provision_instance request via kcat directly to gpu-provisioner
--    (bypasses model-trainer to isolate the new action):
--
--    CORRELATION=$(uuidgen)
--    ORCH=$(uuidgen)
--    REQ=$(uuidgen)
--
--    kubectl -n kafka run kcat-prov-$(date +%s) --rm -i --restart=Never \
--      --image=edenhill/kcat:1.7.1 -- \
--      kcat -P -c 1 \
--        -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
--        -t system.agent.generic.requests \
--        -H correlation_id=$CORRELATION \
--        -H orchestration_id=$ORCH \
--        -H request_id=$REQ \
--        -H message_type=request \
--        -H client_id=demo_client \
--        -H step_name=manual_provision_test \
--        -H sender_agent_type=cli \
--        -H timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ") <<JSON
--    {"action":"orchestrate","config":{"agent_type":"gpu-provisioner"},
--     "input_data":{"gpu":"a100","mode":"prototyping","num_gpus":1}}
--    JSON
--
-- 3. Watch logs for ~5 minutes:
--    kubectl -n ai-persona-system logs deploy/agent-chassis -f --tail=0
--    kubectl -n ai-persona-system logs deploy/thunder-adapter -f --tail=0
--
--    Expect in chassis: "DispatchThunderProvisionAction starting",
--                       "Dispatched provision_instance to thunder-adapter"
--    Expect in adapter: "Received request" with action=provision_instance,
--                       polling logs, then "Provision complete"
--
-- 4. Verify the row landed:
--    SELECT id, thunder_instance_id, status, instance_ip, ssh_user,
--           ssh_key_secret_name, hourly_rate_usd, requested_by
--    FROM thunder_instances
--    ORDER BY created_at DESC LIMIT 1;
--    -- Expect: status='running', instance_ip populated.
--
-- 5. Confirm the Secret was persisted:
--    kubectl -n ai-persona-system get secret -l app.kubernetes.io/managed-by=thunder-adapter
--
-- 6. Decommission via kcat using provisioning_id from step 4:
--    [same kcat shape as step 2, body becomes:]
--    {"action":"orchestrate","config":{"agent_type":"thunder-reaper"},
--     "input_data":{"provisioning_id":"<uuid-from-step-4>",
--                   "reason":"manual_verify"}}
--
--    OR directly via the adapter topic with action=decommission_instance:
--    [target topic system.adapter.thunder.requests, body matches
--     DecommissionInstanceRequest]
--
-- 7. Verify row updated:
--    SELECT status, cost_usd, decommissioned_at FROM thunder_instances
--    WHERE id='<provisioning-id>';
--
-- 8. Restore cap:
--    UPDATE thunder_config SET daily_cap_usd = 100;