#!/bin/bash
# Trigger med-price-collector on the business-intel pod.
# The business-intel pod routes by config.agent_type → selectWorkflow → FindBestGroup.

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Med Price Scrape Trigger"
echo "========================================="
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "  Time:             $TIMESTAMP"
echo "========================================="
echo ""
echo "SAVE THESE IDs:"
echo "  CORRELATION_ID=$CORRELATION_ID"
echo "  ORCHESTRATION_ID=$ORCHESTRATION_ID"
echo ""

kubectl -n kafka run -i --rm kcat-med-price-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.business-intel.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=med-price-scrape-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.business-intel.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"med-price-collector"},"input_data":{"batch_size":5,"retailer_id":"pet_drugs_online"}}
JSON

echo ""
echo "========================================="
echo "Med price scrape triggered"
echo "========================================="
echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f deploy/business-intel --tail=100 | grep '$CORRELATION_ID'"
echo ""
echo "Check orchestration state:"
echo "  SELECT status, current_step, error FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID';"
echo ""
echo "Check price snapshots:"
echo "  SELECT ps.retailer_id, l.retailer_product_name, ps.size_variant, ps.price, ps.typical_vet_price, ps.collected_at"
echo "  FROM business_intel.med_price_snapshots ps"
echo "  JOIN business_intel.med_retailer_listings l ON l.id = ps.listing_id"
echo "  ORDER BY ps.collected_at DESC LIMIT 20;"
echo ""
echo "Check HTTP request log:"
echo "  SELECT method, url, status_code, latency_ms, success, error_message"
echo "  FROM http_request_log"
echo "  WHERE action_name = 'med_scrape_prices'"
echo "  ORDER BY created_at DESC LIMIT 10;"