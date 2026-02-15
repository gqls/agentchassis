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