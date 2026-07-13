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

-------------------------------------------------------------

#!/bin/bash
# Trigger image-build-handler for the icon work item on robot-hands.com.
#
# Uses the psql-jsonb-builder pattern: psql constructs the payload directly
# from the work item's spec column, no shell escaping required, no
# placeholder substitution needed. See debugging guide section 9
# "kcat trigger sent literal placeholder strings" for rationale.

set -u            # fail on any unset variable
set -o pipefail   # propagate pipe failures

WORK_ITEM_ID="3cacc0dd-4bb0-44a1-a3ca-87a3911423a8"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

# Step 1: Pull spec straight out of the DB and assemble the trigger payload.
# psql handles all JSON escaping — newlines, quotes, special characters.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
    psql -U clients_user -d clients_db -t -A -c "
SELECT jsonb_build_object(
    'action', 'orchestrate',
    'config', jsonb_build_object('agent_type', 'image-build-handler'),
    'input_data', jsonb_build_object(
        'site_id', '${SITE_ID}',
        'domain', '${DOMAIN}',
        'work_item_id', '${WORK_ITEM_ID}',
        'item_type', 'needs_imagery',
        'spec', spec::jsonb
    )
)::text
FROM site_work_items
WHERE id = '${WORK_ITEM_ID}';
" > /tmp/trigger.json

# Step 2: Guards before sending. Catch the two most common silent failures.
if [ ! -s /tmp/trigger.json ]; then
    echo "ERROR: trigger payload empty — work item not found or psql failed"
    exit 1
fi
if ! cat /tmp/trigger.json | jq . >/dev/null 2>&1; then
    echo "ERROR: trigger payload is malformed JSON"
    cat /tmp/trigger.json
    exit 1
fi
if grep -qE '<[A-Z_]+>' /tmp/trigger.json; then
    echo "ERROR: trigger payload contains unsubstituted placeholder"
    grep -E '<[A-Z_]+>' /tmp/trigger.json
    exit 1
fi

# Step 3: Headers
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
CLIENT_ID="demo_client"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "=== Triggering image-build-handler for icon work item ==="
echo "========================================="
echo "  Correlation ID:   ${CORRELATION_ID}"
echo "  Orchestration ID: ${ORCHESTRATION_ID}"
echo "  Work Item ID:     ${WORK_ITEM_ID}"
echo "  Domain:           ${DOMAIN}"
echo "  Site ID:          ${SITE_ID}"
echo "  Time:             ${TIMESTAMP}"
echo "========================================="
echo "Payload preview (kind, prompt snippet):"
cat /tmp/trigger.json | jq '{
    work_item_id: .input_data.work_item_id,
    item_type: .input_data.item_type,
    spec_kind: .input_data.spec.kind,
    spec_purpose: .input_data.spec.purpose,
    spec_asset_key: .input_data.spec.asset_key,
    spec_prompt_preview: (.input_data.spec.prompt[0:120] + "...")
}'
echo "========================================="
echo ""

# Step 4: Send to kafka.
cat /tmp/trigger.json | kubectl -n kafka run -i --rm "kcat-icon-$(date +%s)" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -c 1 \
        -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
        -t system.agent.generic.requests \
        -H correlation_id=${CORRELATION_ID} \
        -H orchestration_id=${ORCHESTRATION_ID} \
        -H request_id=${REQUEST_ID} \
        -H message_id=${MESSAGE_ID} \
        -H message_type=request \
        -H client_id=${CLIENT_ID} \
        -H action=orchestrate \
        -H sender_agent_type=cli \
        -H sender_agent_id=cli-user \
        -H responses_topic=system.agent.generic.responses \
        -H timestamp=${TIMESTAMP}

echo ""
echo "Sent. Wait ~60-90s, then verify with:"
echo ""
echo "  -- Orchestration outcome:"
echo "  SELECT orchestration_id, status, current_step, "
echo "         LEFT(COALESCE(error, ''), 300) AS error_preview"
echo "  FROM orchestration_states"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND created_at > now() - interval '5 minutes'"
echo "  ORDER BY created_at DESC;"
echo ""
echo "  -- Asset row:"
echo "  SELECT asset_key, purpose, origin_model, created_at"
echo "  FROM assets"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND asset_key = 'icon_cross_technology';"
echo ""
echo "  -- Work item state:"
echo "  SELECT id, status, completed_at, attempt_count,"
echo "         result->>'completed_by_orchestration_id' AS completed_by"
echo "  FROM site_work_items"
echo "  WHERE id = '${WORK_ITEM_ID}';"

