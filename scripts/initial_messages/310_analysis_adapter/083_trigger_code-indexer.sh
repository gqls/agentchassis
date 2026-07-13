set -euo pipefail

TARGET_AGENT_TYPE='code-indexer'
OWNER='gqls'
REPO='agentchassis'
REF='HEAD'
LANGUAGE='go'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "========================================="
echo "Manual code-indexer trigger"
echo "========================================="
echo "  Target Agent Type: $TARGET_AGENT_TYPE"
echo "  Owner: $OWNER"
echo "  Repo: $REPO"
echo "  Ref: $REF"
echo "  Timestamp: $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-classifier-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-classifier-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":{"owner":"$OWNER","repo":"$REPO","ref":"$REF","language":"$LANGUAGE"}}
JSON

echo ""
echo "========================================="
echo "Classifier triggered. Running in a chassis pod (generic entry point)."
echo "========================================="
echo ""
echo "Tail chassis logs for this correlation:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=500 | grep '$CORRELATION_ID'"
echo "  kubectl -n ai-persona-system logs -f -l agent_type=code-indexer --tail=500 | grep '$CORRELATION_ID'"
echo ""
echo "Watch key workflow steps:"
echo "  kubectl -n ai-persona-system logs -l agent_type=code-indexer --tail=500 | grep '$CORRELATION_ID' | grep -E 'search_domain|scrape_site|read_site_specs|read_layout_taxonomy|classify_and_extract|write_classification_spec'"
echo ""
echo "Check orchestration state:"
echo "  psql -c \"SELECT status, current_step, EXTRACT(EPOCH FROM (NOW() - last_activity))::int AS since_s, substring(COALESCE(error,''), 1, 300) AS err FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid ORDER BY created_at;\""
echo ""
echo "kubectl -n ai-persona-system logs --tail=500 -l app=analyser-adapter -f --max-log-requests 20 | tee logs-analyser-adapter.json"
echo "kubectl -n ai-persona-system logs --tail=500 -l agent_type=code-indexer -f --max-log-requests 20 | tee logs-code-indexer.json"
echo ""
