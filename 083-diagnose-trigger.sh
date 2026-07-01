set -euo pipefail

TARGET_AGENT_TYPE='diagnose-orchestrator'
OWNER='gqls'; REPO='agentchassis'; REF='HEAD'
SYMPTOM='index page completed but content is a stub'
RUNTIME_SITE='gamesdesign.co.uk'
SITE_ID='e33263f4-74f8-494f-b191-546845dbbddf'

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID='demo_client'

echo "SAVE: CORRELATION_ID=$CORRELATION_ID  ORCHESTRATION_ID=$ORCHESTRATION_ID"

kubectl -n kafka run -i --rm "kcat-diagnose-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "orchestration_name=manual-diagnose-$(date +%Y%m%d-%H%M%S)" \
  -H "step_name=start" \
  -H "client_id=$CLIENT_ID" \
  -H "message_type=request" \
  -H "action=orchestrate" \
  -H "from_agent_type=user" \
  -H "from_agent_id=cli" \
  -H "responses_topic=system.agent.generic.responses" <<JSON
{"action":"orchestrate","config":{"agent_type":"$TARGET_AGENT_TYPE"},"input_data":{"owner":"$OWNER","repo":"$REPO","ref":"$REF","symptom":"$SYMPTOM","runtime_site":"$RUNTIME_SITE","site_id":"$SITE_ID"}}
JSON

echo "Tail by correlation:"
echo "  kubectl -n ai-persona-system logs -f -l agent_type=diagnose-agent --tail=500 | grep '$CORRELATION_ID'"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis    --tail=500 | grep '$CORRELATION_ID'"