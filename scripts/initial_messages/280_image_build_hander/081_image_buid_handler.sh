WORK_ITEM_ID="19a4b7cd-8bd8-4d52-a6a3-e8d094ad2536"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

# Pull spec into a shell variable
SELECT spec::text FROM site_work_items WHERE id = '$WORK_ITEM_ID';

echo "Spec retrieved ($(echo -n "$SPEC" | wc -c) bytes):"
echo "$SPEC" | jq .

/tmp/trigger.json
{"action":"orchestrate","config":{"agent_type":"image-build-handler"},"input_data":{"site_id":"00ff3af5-dad8-4770-9f70-3edc267a3c92","domain":"robot-hands.com","work_item_id":"19a4b7cd-8bd8-4d52-a6a3-e8d094ad2536","item_type":"needs_imagery","spec":{"key":"hero_home","kind":"hero","check":"unfulfilled_imagery_plan","scope":"page","prompt":"A dramatic, high-contrast close-up photograph of an industrial robotic gripper mounted on a robot arm, gripping a machined metal component. Soft directional lighting from the left, dark neutral background with subtle blue ambient light. Shallow depth of field. The gripper shows visible pneumatic lines and machined aluminium surfaces. Mood: precision, power, technical excellence. No text overlays, no logos, no people.","purpose":"hero","asset_key":"hero_home","scope_ref":"index","brand_update":true}}}


----------------------------------

set -u            # fail on any unset variable
set -o pipefail   # propagate pipe failures

WORK_ITEM_ID="19a4b7cd-8bd8-4d52-a6a3-e8d094ad2536"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

# Headers
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "=== Triggering image-build-handler for work item $WORK_ITEM_ID ==="
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Work Item ID:     $WORK_ITEM_ID"
echo "  domain:           $DOMAIN"
echo "  Site ID:          $SITE_ID"
echo "  Time:             $TIMESTAMP"
echo "========================================="
echo ""

kubectl -n kafka run -i --rm kcat-prep-$(date +%s) \
--image=edenhill/kcat:1.7.1 --restart=Never -- \
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
{"action":"orchestrate","config":{"agent_type":"image-build-handler"},"input_data":{"site_id":"00ff3af5-dad8-4770-9f70-3edc267a3c92","domain":"robot-hands.com","work_item_id":"19a4b7cd-8bd8-4d52-a6a3-e8d094ad2536","item_type":"needs_imagery","spec":{"key":"hero_home","kind":"hero","check":"unfulfilled_imagery_plan","scope":"page","prompt":"A dramatic, high-contrast close-up photograph of an industrial robotic gripper mounted on a robot arm, gripping a machined metal component. Soft directional lighting from the left, dark neutral background with subtle blue ambient light. Shallow depth of field. The gripper shows visible pneumatic lines and machined aluminium surfaces. Mood: precision, power, technical excellence. No text overlays, no logos, no people.","purpose":"hero","asset_key":"hero_home","scope_ref":"index","brand_update":true}}}
JSON




{"action":"orchestrate","config":{"agent_type":"image-build-handler"},"input_data":{"site_id":"00ff3af5-dad8-4770-9f70-3edc267a3c92","domain":"robot-hands.com","work_item_id":"19a4b7cd-8bd8-4d52-a6a3-e8d094ad2536","item_type":"needs_imagery","spec":{"key":"hero_home","kind":"hero","check":"unfulfilled_imagery_plan","scope":"page","prompt":"A dramatic, high-contrast close-up photograph of an industrial robotic gripper mounted on a robot arm, gripping a machined metal component. Soft directional lighting from the left, dark neutral background with subtle blue ambient light. Shallow depth of field. The gripper shows visible pneumatic lines and machined aluminium surfaces. Mood: precision, power, technical excellence. No text overlays, no logos, no people.","purpose":"hero","asset_key":"hero_home","scope_ref":"index","brand_update":true}}}


-- Find the most recent page-content-writer invocation
SELECT id, agent_type, status, created_at
FROM workflow_states
WHERE agent_type = 'page-build-handler'
ORDER BY created_at DESC
LIMIT 5;