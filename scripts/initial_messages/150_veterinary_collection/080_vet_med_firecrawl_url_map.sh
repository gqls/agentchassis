#!/bin/bash
# Trigger med-url-mapper (Firecrawl /map based discovery).
# Optional: pass retailer_id and/or search term.
#
# Usage:
#   bash trigger_med_url_map.sh                          # all retailers
#   bash trigger_med_url_map.sh viovet                   # single retailer
#   bash trigger_med_url_map.sh viovet "prescription"    # with search filter

RETAILER_ID="${1:-}"
SEARCH_TERM="${2:-}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

# Build input_data JSON
INPUT_DATA="{}"
if [ -n "$RETAILER_ID" ] && [ -n "$SEARCH_TERM" ]; then
    INPUT_DATA="{\"retailer_id\":\"$RETAILER_ID\",\"search\":\"$SEARCH_TERM\"}"
elif [ -n "$RETAILER_ID" ]; then
    INPUT_DATA="{\"retailer_id\":\"$RETAILER_ID\"}"
elif [ -n "$SEARCH_TERM" ]; then
    INPUT_DATA="{\"search\":\"$SEARCH_TERM\"}"
fi

echo "========================================="
echo "Med URL Map Trigger (Firecrawl /map)"
echo "========================================="
echo "  Retailer:         ${RETAILER_ID:-all}"
echo "  Search:           ${SEARCH_TERM:-none}"
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "========================================="

kubectl -n kafka run -i --rm kcat-med-map-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.business-intel.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=med-url-map-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.business-intel.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"med-url-mapper"},"input_data":$INPUT_DATA}
JSON

echo ""
echo "========================================="
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f deploy/business-intel --tail=100 | grep MedMap"
echo "========================================="
