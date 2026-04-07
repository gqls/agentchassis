#!/bin/bash
# Trigger med-price-scrape-orchestrator — spawns a temporary pod to scrape prices.

BATCH_SIZE="${1:-20}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)

BATCH_SIZE="5"

echo "========================================="
echo "Med Price Scrape (spawned pod)"
echo "========================================="
echo "  Batch size:       $BATCH_SIZE"
echo "  Correlation ID:   $CORRELATION_ID"
echo "  Orchestration ID: $ORCHESTRATION_ID"
echo "========================================="

kubectl -n kafka run -i --rm kcat-med-scrape-$(date +%s) \
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
  -H client_id=demo_client \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.business-intel.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":$BATCH_SIZE}}
JSON

echo ""
echo "Monitor business-intel (spawn step):"
echo "  kubectl -n ai-persona-system logs -f deploy/business-intel --tail=20 | grep -E 'spawn|med-price'"
echo ""
echo "Monitor spawned pod (actual work):"
echo "  kubectl -n ai-persona-system logs -f -l app=dynamic-agent --tail=50 | grep MedScrape"


# check overall progress
SELECT r.id,
       count(l.id) as listings,
       count(l.id) FILTER (WHERE l.last_scraped_at IS NOT NULL) as scraped,
       count(l.id) FILTER (WHERE l.last_scraped_at IS NULL) as pending
FROM business_intel.med_retailers r
LEFT JOIN business_intel.med_retailer_listings l ON l.retailer_id = r.id
WHERE r.is_active = true
GROUP BY r.id;
        id        | listings | scraped | pending
------------------+----------+---------+---------
 pet_drugs_online |       44 |      44 |       0
 animed_direct    |      203 |     154 |      49
 hyperdrug        |       54 |      46 |       8
 viovet           |        4 |       2 |       2
(4 rows)


