#!/bin/bash
# ============================================================================
# trigger_training_export_v3.sh
# ============================================================================
# Triggers a training data export via the orchestrator wrapper.
#
# The orchestrator spawns training-data-exporter in a dedicated pod, which
# writes the export to training_exports.runs + training_exports.rows.
#
# Flow:
#   Kafka → training-data-export-orchestrator (runs in agent-chassis pod)
#        → spawn_agent → training-data-exporter (dedicated pod spawned)
#        → call_agent  → does the DB work, returns export_id
#        → complete
#
# After this runs, query the result:
#
#   SELECT status, jsonb_pretty(collected_data->'export_result') as summary
#   FROM orchestration_states
#   WHERE orchestration_id = '<ORCH_ID>'::uuid;
#
# The export_id (UUID) in the summary is the primary reference to the dataset
# in training_exports.runs. At training time, use \copy from training_exports.rows
# to stream out NDJSON.
# ============================================================================

set -e

# ── Edit these per run ──────────────────────────────────────────────────────
AGENT_TYPE="page-content-writer"
STEP_NAME="process_sections_loop_iter_0_generate_content"
MODEL_FILTER="claude-sonnet-4-6"
# ────────────────────────────────────────────────────────────────────────────

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

echo "============================================================"
echo "Training Data Export v3"
echo "============================================================"
echo "  agent_type:   $AGENT_TYPE"
echo "  step_name:    $STEP_NAME"
echo "  model_filter: $MODEL_FILTER"
echo ""
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo "============================================================"

# Build payload flat via jq — safer than heredoc interpolation
# (doc 016 §9: use here-string with single-quoted flat JSON)
PAYLOAD=$(jq -nc \
    --arg agent "$AGENT_TYPE" \
    --arg step  "$STEP_NAME" \
    --arg model "$MODEL_FILTER" \
    '{
      action: "orchestrate",
      config: {agent_type: "training-data-export-orchestrator"},
      input_data: {
        agent_type:   $agent,
        step_name:    $step,
        model_filter: $model
      }
    }')

kubectl -n kafka run -i --rm "kcat-export-v3-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=training-export-v3-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=demo_client" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" \
  <<<"$PAYLOAD"

echo ""
echo "Check orchestration status (wait ~30s for the spawn + DB write):"
echo ""
echo "kubectl -n ai-persona-system exec -i pod/postgres-clients-0 -- \\"
echo "    psql -U clients_user -d clients_db -c \""
echo "SELECT status, current_step, LEFT(error, 300) as error_preview,"
echo "       jsonb_pretty(collected_data->'export_result') as summary"
echo "FROM orchestration_states"
echo "WHERE orchestration_id = '$ORCHESTRATION_ID'::uuid;\""
echo ""
echo "Find the spawned worker pod:"
echo ""
echo "  kubectl -n ai-persona-system get pods -l agent_type=training-data-exporter \\"
echo "      --sort-by=.metadata.creationTimestamp | tail -3"
echo ""
echo "Once an export_id comes back, query the data:"
echo ""
echo "  SELECT * FROM training_exports.runs ORDER BY created_at DESC LIMIT 1;"
echo "  SELECT COUNT(*) FROM training_exports.rows WHERE export_id = '<EXPORT_ID>';"