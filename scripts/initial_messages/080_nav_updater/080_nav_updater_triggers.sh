#!/bin/bash
# Trigger nav-updater for sites with orphaned pages that have nav flags set

DOMAINS=("robot-hands.com" "finetuning.uk" "leopardessconsulting.co.uk" "gaswholesalers.com")

for DOMAIN in "${DOMAINS[@]}"; do
    CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
    ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
    MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
    REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    CLIENT_ID="demo_client"

    INPUT_DATA="{\"domain\":\"${DOMAIN}\"}"

    echo "========================================="
    echo "Nav Updater: $DOMAIN"
    echo "  Correlation ID: $CORRELATION_ID"
    echo "========================================="

    kubectl -n kafka run -i --rm kcat-nav-updater-$(date +%s) \
      --image=edenhill/kcat:1.7.1 \
      --restart=Never -- \
      kcat -P \
      -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
      -t system.agent.generic.requests \
      -H correlation_id=$CORRELATION_ID \
      -H orchestration_id=$ORCHESTRATION_ID \
      -H request_id=$REQUEST_ID \
      -H message_id=$MESSAGE_ID \
      -H message_type=request \
      -H client_id=$CLIENT_ID \
      -H action=process \
      -H sender_agent_type=cli \
      -H sender_agent_id=cli-user \
      -H responses_topic=system.agent.generic.responses \
      -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_nav_updater","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_nav_updater":{"action":"spawn_agent","config":{"role":"nav_updater","agent_type":"nav-updater"},"output_field":"nav_updater_agent","next_step":"call_nav_updater","description":"Spawn nav-updater agent"},"call_nav_updater":{"action":"call_agent","config":{"agent_type":"nav-updater","target_role":"nav_updater","input_mapping":{"domain":"input_data.domain"},"timeout_seconds":600},"output_field":"nav_result","next_step":"complete","description":"Run nav update"},"complete":{"action":"complete_workflow","config":{"output_fields":["nav_result"]},"description":"Nav update complete"}}}},"input_data":${INPUT_DATA}}
JSON

    echo "  Triggered: $DOMAIN (CORRELATION_ID=$CORRELATION_ID)"
    echo ""
    sleep 2
done

echo "All 4 sites triggered. Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep nav-updater"