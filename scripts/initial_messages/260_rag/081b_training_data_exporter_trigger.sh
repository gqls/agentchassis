#!/bin/bash
# ============================================================================
# Trigger training-data-exporter for page-content-writer iter_0
# ============================================================================
# Sends a Kafka message to spawn the training-data-exporter orchestration.
# Mirrors the adoption / rag-test-agent trigger pattern.
#
# After this runs, the exported JSONL file will be at:
#   /tmp/training_exports/page_content_writer_iter0.jsonl
# inside one of the agent-chassis pods (the one that picked up the message).
# ============================================================================

set -e

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

echo "CORRELATION_ID=$CORRELATION_ID"
echo "ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run -i --rm "kcat-export-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=training-export-iter0-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=demo_client" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{
  "action": "orchestrate",
  "config": {"agent_type": "training-data-exporter"},
  "input_data": {
    "agent_type": "page-content-writer",
    "step_name": "process_sections_loop_iter_0_generate_content",
    "model_filter": "claude-sonnet-4-6",
    "output_path": "/tmp/training_exports/page_content_writer_iter0.jsonl"
  }
}
JSON

echo ""
echo "Saved IDs for checking:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""
echo "Verify progress with:"
echo ""
echo "kubectl -n ai-persona-system exec -i pod/postgres-clients-0 -- psql -U clients_user -d clients_db -c \""
echo "  SELECT status, current_step, jsonb_pretty(collected_data->'export_result') as export_result"
echo "  FROM orchestration_states"
echo "  WHERE orchestration_id = '\$ORCHESTRATION_ID'::uuid;\""
echo ""
echo "Retrieve the file from whichever pod handled it:"
echo ""
echo "  # Find the pod:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=500 | grep '\$CORRELATION_ID' | head -1"
echo ""
echo "  # Copy it out (adjust pod name):"
echo "  kubectl -n ai-persona-system cp <pod>:/tmp/training_exports/page_content_writer_iter0.jsonl ./page_content_writer_iter0.jsonl"