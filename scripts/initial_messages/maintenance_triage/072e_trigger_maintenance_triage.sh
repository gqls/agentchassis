#!/bin/bash
# trigger_maintenance_triage.sh
# Triggers the maintenance-triage agent via generic-agent.
#
# Usage:
#   ./trigger_maintenance_triage.sh                                    # scan all deployed sites
#   ./trigger_maintenance_triage.sh leopardessconsulting.co.uk         # scan one site
#   ./trigger_maintenance_triage.sh --dry-run                          # scan all, don't dispatch
#   ./trigger_maintenance_triage.sh leopardessconsulting.co.uk --dry-run
#   ./trigger_maintenance_triage.sh --threshold 14                     # custom stale threshold (days)

set -euo pipefail

DOMAIN=""
DRY_RUN="false"
THRESHOLD=""

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)
            DRY_RUN="true"
            shift
            ;;
        --threshold)
            THRESHOLD="$2"
            shift 2
            ;;
        *)
            DOMAIN="$1"
            shift
            ;;
    esac
done

-- v --
threshold is number of days old
-- v --



DOMAIN="leopardessconsulting.co.uk"
DRY_RUN="false"
THRESHOLD="4"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Maintenance Triage Scan"
echo "========================================="
echo "  Domain:         ${DOMAIN:-all deployed sites}"
echo "  Dry run:        $DRY_RUN"
echo "  Threshold:      ${THRESHOLD:-30} days"
echo "  Correlation ID: $CORRELATION_ID"
echo "========================================="

# Build input_data JSON
INPUT_DATA="{"
FIELDS=""

if [[ -n "$DOMAIN" ]]; then
    FIELDS="${FIELDS}\"domain\":\"${DOMAIN}\","
fi

if [[ "$DRY_RUN" == "true" ]]; then
    FIELDS="${FIELDS}\"dry_run\":true,"
fi

if [[ -n "$THRESHOLD" ]]; then
    FIELDS="${FIELDS}\"stale_threshold_days\":${THRESHOLD},"
fi

# Remove trailing comma and close
FIELDS="${FIELDS%,}"
# INPUT_DATA="${INPUT_DATA}${FIELDS}}"
# Build input_data JSON — always include all fields so input_mapping paths resolve
INPUT_DATA="{\"domain\":\"${DOMAIN}\",\"dry_run\":${DRY_RUN},\"stale_threshold_days\":${THRESHOLD:-30}}"

echo "  Input data:     $INPUT_DATA"
echo ""

kubectl -n kafka run -i --rm kcat-maintenance-triage-$(date +%s) \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_triage","processing_mode":"orchestrator","timeout_seconds":2400,"steps":{"spawn_triage":{"action":"spawn_agent","config":{"role":"triage","agent_type":"maintenance-triage"},"output_field":"triage_agent","next_step":"call_triage","description":"Spawn maintenance-triage agent"},"call_triage":{"action":"call_agent","config":{"agent_type":"maintenance-triage","target_role":"triage","input_mapping":{"domain":"input_data.domain","stale_threshold_days":"input_data.stale_threshold_days","dry_run":"input_data.dry_run"},"timeout_seconds":1800},"output_field":"triage_result","next_step":"complete","description":"Run triage scan and dispatch"},"complete":{"action":"complete_workflow","config":{"output_fields":["triage_result"]},"description":"Triage dispatch complete"}}}},"input_data":${INPUT_DATA}}
JSON

echo ""
echo "========================================="
echo "Triage scan triggered"
echo "========================================="
echo ""
echo "Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
echo ""
echo "Check queue:"
echo "  kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c \\"
echo "    \"SELECT id, site_id, task_type, status, priority, payload->>'pages' as pages, created_at FROM maintenance_queue ORDER BY created_at DESC LIMIT 20;\""
echo ""
echo "CORRELATION_ID=$CORRELATION_ID"





# --------------- 3 ---------------------

# ===== FILL IN THESE VALUES =====
CORRELATION_ID=
ORCHESTRATION_ID=
HITL_REQUEST_ID=
RESPONSES_TOPIC=
# ================================

# ------- 3 3 3 --------------------

CORRELATION_ID=5a09ca74-a604-4549-9c48-fb47cd50eedd
ORCHESTRATION_ID=66a49e34-a16e-4cc2-8a15-4d4c93ca6b77
HITL_REQUEST_ID=518be7d3-63e1-49f4-a7c5-926dcc39fc14
RESPONSES_TOPIC=job.5a09ca74-2542e33e-content-reviewer-spawn_reviewer.responses

# Auto-generated values
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

# Validate required fields
if [ -z "$CORRELATION_ID" ] || [ -z "$ORCHESTRATION_ID" ] || [ -z "$HITL_REQUEST_ID" ] || [ -z "$RESPONSES_TOPIC" ]; then
    echo "ERROR: Please fill in all required values at the top of the script"
    echo ""
    echo "Run this query to find them:"
    echo "  SELECT request_id, orchestration_id, correlation_id, responses_topic"
    echo "  FROM awaited_requests"
    echo "  WHERE step_name = 'escalate_to_human' AND status = 'waiting'"
    echo "  ORDER BY sent_at DESC LIMIT 1;"
    exit 1
fi

echo "========================================="
echo "Sending HITL Response: 3 Content Review Approval"
echo "========================================="
echo "  Correlation ID:      $CORRELATION_ID"
echo "  Orchestration ID:    $ORCHESTRATION_ID"
echo "  HITL Request ID:     $HITL_REQUEST_ID"
echo "  Message ID:          $MESSAGE_ID"
echo "  Step:                escalate_to_human"
echo "  Topic:               $RESPONSES_TOPIC"
echo "  Time:                $TIMESTAMP"
echo "========================================="
echo ""

kubectl -n kafka run -i --rm kcat-producer-content-review-$(date +%s) \
--image=edenhill/kcat:1.7.1 \
--restart=Never -- \
kcat -P \
-b personae-kafka-cluster-kafka-bootstrap:9092 \
-t $RESPONSES_TOPIC \
-H correlation_id=$CORRELATION_ID \
-H orchestration_id=$ORCHESTRATION_ID \
-H message_id=$MESSAGE_ID \
-H message_type=response \
-H client_id=demo_client \
-H in_response_to_request_id=$HITL_REQUEST_ID \
-H in_response_to_step_name=escalate_to_human \
-H status=complete \
-H is_complete=true \
-H sender_agent_type=human \
-H sender_agent_id=cli-user \
-H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","message_id":"${MESSAGE_ID}","message_type":"response","client_id":"demo_client","in_response_to_request_id":"${HITL_REQUEST_ID}","in_response_to_step_name":"escalate_to_human","in_response_to_action":"request_human_input","status":"complete","is_complete":true,"is_error":false,"sender":{"agent_id":"cli-user","agent_type":"human","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"body":{"success":true,"approved":true,"status":"approved","responded_by":"cli-user@example.com","responded_at":"${TIMESTAMP}","review_mode":"escalated","edits":{},"comments":"Content reviewed and approved via CLI. Auto-eval issues have been addressed."}}
JSON


echo ""
echo "========================================="
echo "HITL Response sent!"
echo "========================================="
echo ""
echo "Check if workflow continued:"
echo "  kubectl logs -n ai-persona-system -l app=content-reviewer --tail=50 | grep \"$CORRELATION_ID\" | grep -E 'process_escalation|finalize|complete'"
echo ""
echo "Monitor content-reviewer progress:"
echo "  kubectl logs -n ai-persona-system -l app=content-reviewer --tail=100 | grep \"$CORRELATION_ID\""
echo ""
echo "Check pageflow-builder (parent) progress:"
echo "  kubectl logs -n ai-persona-system -l app=pageflow-builder --tail=100 | grep \"$CORRELATION_ID\""
echo ""







UPDATE maintenance_queue
SET status = 'pending', claimed_at = NULL, claimed_by = NULL
WHERE id = '5a20f87f-b780-457e-bd5f-0ef17bbba3f4';