#!/bin/bash
# Trigger rerender-pages agent (uses agent_definitions workflow) multiple pages

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

echo "========================================="
echo "Triggering rerender-pages agent"
echo "========================================="
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "  Site ID:             $SITE_ID"
echo "  Domain:              $DOMAIN"
echo "  Time:                $TIMESTAMP"
echo "========================================="

kubectl -n kafka run -i --rm kcat-rerender-agent \
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
-H action=orchestrate \
-H sender_agent_type=cli \
-H sender_agent_id=cli-user \
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"orchestrate","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Message sent. Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100"


-------------------------------------

# rerender single page
#!/bin/bash
# Trigger rerender-pages agent for a SINGLE PAGE
# Change PAGE_NAME to target a specific page

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"
SITE_ID="4851f6fc-71cf-4160-a270-e03d6d3e0732"
DOMAIN="leopardessconsulting.co.uk"

# Single page filter - use one of these:
PAGE_NAME="Home"  # Match by page name in DB
# PAGE_ID="uuid-of-page"  # Or match by page ID

echo "========================================="
echo "Triggering rerender-pages agent (SINGLE PAGE)"
echo "========================================="
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "  Site ID:             $SITE_ID"
echo "  Domain:              $DOMAIN"
echo "  Page Name:           $PAGE_NAME"
echo "  Time:                $TIMESTAMP"
echo "========================================="

kubectl -n kafka run -i --rm kcat-rerender-single \
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
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerender_agent","steps":{"spawn_rerender_agent":{"action":"spawn_agent","config":{"role":"rerenderer","agent_type":"rerender-pages"},"description":"Spawn rerender-pages agent","next_step":"call_rerender_agent","output_field":"rerender_agent_info"},"call_rerender_agent":{"action":"call_agent","config":{"agent_type":"rerender-pages","target_role":"rerenderer","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain","page_name":"input_data.page_name"},"timeout_seconds":300},"description":"Call rerender-pages agent for single page","next_step":"complete","output_field":"rerender_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result"]}}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","page_name":"${PAGE_NAME}"}}
JSON

echo ""
echo "Message sent. Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100"