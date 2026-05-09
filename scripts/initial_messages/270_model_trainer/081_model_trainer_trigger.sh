#!/bin/bash
# trigger_model_trainer.sh — kicks off a flywheel C phase 2 training run

AGENT_TYPE="model-trainer"
EXPORT_ID="146a9a12-c953-48eb-bf1f-c1856e5f13b7"   # iter_0 export from phase 1

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

# Hyperparameters — same shape as iter_0 manifest, used for reproducibility
# and stored in model_lifecycle.training_runs.hyperparameters JSONB.
read -r -d '' HYPERPARAMETERS <<'HP'
{
  "base_model": "unsloth/Llama-3.3-70B-Instruct-bnb-4bit",
  "epochs": 3,
  "batch": 1,
  "grad_accum": 8,
  "lr": 2e-4,
  "lora_r": 16,
  "lora_alpha": 32,
  "max_seq": 4096,
  "seed": 3407
}
HP

# Strip newlines so the JSONB embeds cleanly in the kafka message body
HYPERPARAMETERS_INLINE=$(echo "$HYPERPARAMETERS" | jq -c .)

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Export:         ${EXPORT_ID}"
echo "  Correlation:    ${CORRELATION_ID}"
echo "  Orchestration:  ${ORCHESTRATION_ID}"
echo "  Time:           ${TIMESTAMP}"
echo ""

kubectl -n kafka run -i --rm kcat-train-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P -c 1 \
    -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CORRELATION_ID \
    -H orchestration_id=$ORCHESTRATION_ID \
    -H request_id=$REQUEST_ID \
    -H message_id=$MESSAGE_ID \
    -H message_type=request \
    -H client_id=$CLIENT_ID \
    -H action=orchestrate \
    -H sender_agent_type=cli \
    -H sender_agent_id=cli-user \
    -H responses_topic=system.agent.generic.responses \
    -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"export_id":"${EXPORT_ID}","hyperparameters":${HYPERPARAMETERS_INLINE},"triggered_by":null,"orchestration_id":null}}
JSON

echo ""
echo "Save these for monitoring:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""
echo "Watch logs:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100 | grep \"$CORRELATION_ID\""
echo ""



--------------------

# All chassis activity for this run
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 \
  | grep "$CORRELATION_ID"

# Just the data-preparer pod (when it spawns)
kubectl -n ai-persona-system get pods -l agent-type=training-data-preparer
kubectl -n ai-persona-system logs -f <pod-name> | grep "$CORRELATION_ID"

# Verify env var landed in the spawned pod (the IMAGE_BUCKET=finetuning thing)
kubectl -n ai-persona-system exec <data-preparer-pod> -- env | grep -E 'BUCKET|B2_'


------------------------

-- ============================================================================
-- Step 1: Confirm the trigger reached the chassis and an orchestration started
-- ============================================================================
SELECT orchestration_id, owner_agent_type, status, current_step,
       created_at, updated_at,
       NOW() - updated_at AS stale_for
FROM orchestration_states
WHERE created_at > NOW() - INTERVAL '5 minutes'
ORDER BY created_at DESC
LIMIT 5;
-- Expect: 1+ rows, owner_agent_type='model-trainer' or 'training-data-preparer',
-- status progresses through 'RUNNING' → 'AWAITING_RESPONSES' → 'COMPLETE' or 'FAILED'.

-- ============================================================================
-- Step 2: Did the data-preparer INSERT a training_runs row?
-- ============================================================================
SELECT id, export_id, status, hyperparameters,
       created_at, started_at, completed_at, error_message
FROM model_lifecycle.training_runs
ORDER BY created_at DESC
LIMIT 5;
-- Expect: 1 new row, status='pending' (we don't update to 'running' yet —
-- gpu-provisioner does that and we haven't built it).

-- ============================================================================
-- Step 3: Check for errors on the run
-- ============================================================================
SELECT id, status, error_message, created_at
FROM model_lifecycle.training_runs
WHERE status = 'failed'
  AND created_at > NOW() - INTERVAL '5 minutes';
-- Expect: empty UNLESS something went wrong in data prep itself
-- (e.g. S3 upload failed, export_id has zero rows).

-- ============================================================================
-- Step 4: Did any artefact rows land? (none expected — we only prep data)
-- ============================================================================
SELECT * FROM model_lifecycle.artefacts
WHERE created_at > NOW() - INTERVAL '5 minutes';
-- Expect: empty. artefacts come from collect-adapter (025), not yet built.

-- ============================================================================
-- Step 5: Look at the spawned Job pod activity
-- ============================================================================
SELECT correlation_id, orchestration_id, owner_agent_type,
       current_step, status, last_activity,
       collected_data ? 'preparation_result' AS has_prep_result
FROM orchestration_states
WHERE owner_agent_type IN ('model-trainer', 'training-data-preparer')
  AND created_at > NOW() - INTERVAL '5 minutes'
ORDER BY created_at;
-- Expect: orchestrator row + 1 worker row.
-- has_prep_result=true on either if data-preparer completed successfully.

-- ============================================================================
-- Step 6: Inspect the actual preparation_result if it exists
-- ============================================================================
SELECT current_step,
       collected_data->'preparation_result' AS prep_result,
       collected_data->'preparation_result'->>'training_run_id' AS training_run_id,
       collected_data->'preparation_result'->>'dataset_uri' AS dataset_uri,
       collected_data->'preparation_result'->>'row_count' AS row_count,
       collected_data->'preparation_result'->>'size_bytes' AS size_bytes
FROM orchestration_states
WHERE collected_data ? 'preparation_result'
  AND created_at > NOW() - INTERVAL '10 minutes'
ORDER BY created_at DESC
LIMIT 3;
-- Expect: dataset_uri starts with 's3://finetuning/datasets/...',
-- row_count matches training_exports.runs.rows_exported (~1958 for iter_0),
-- size_bytes around 20-25 MB.

-- ============================================================================
-- Step 7: Stuck/stale orchestrations (run if nothing happened in 2 min)
-- ============================================================================
SELECT orchestration_id, owner_agent_type, status, current_step,
       NOW() - last_activity AS stuck_for,
       error_message
FROM orchestration_states
WHERE status = 'AWAITING_RESPONSES'
  AND last_activity < NOW() - INTERVAL '2 minutes'
  AND created_at > NOW() - INTERVAL '10 minutes';
-- Expect: maybe the orchestrator stuck on 'call_provisioner' since
-- gpu-provisioner doesn't exist yet — that's the expected stopping point.

-- ============================================================================
-- Step 8: Check awaited requests (what the orchestrator is waiting on)
-- ============================================================================
SELECT ar.request_id, ar.step_name, ar.target_agent_type, ar.status,
       ar.timeout_at, NOW() - ar.created_at AS waiting_for
FROM awaited_requests ar
WHERE ar.created_at > NOW() - INTERVAL '10 minutes'
ORDER BY ar.created_at DESC;
-- Expect: rows for each call_agent step. After data-preparer completes:
-- one 'completed' row for 'call_data_preparer', one 'waiting' for 'call_provisioner'.

-- ============================================================================
-- Step 9: Verify S3 object actually exists (separate — check Backblaze UI/CLI)
-- ============================================================================
-- From your laptop, requires b2 CLI configured:
--   b2 ls finetuning datasets/146a9a12-c953-48eb-bf1f-c1856e5f13b7/
-- Or via aws-cli if pointed at B2 endpoint.