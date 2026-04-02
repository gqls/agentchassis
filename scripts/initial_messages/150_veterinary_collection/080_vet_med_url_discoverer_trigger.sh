#!/bin/bash
# Trigger med-url-discoverer on the business-intel pod.
# Test with a single retailer first.

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Med URL Discovery Trigger"
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

kubectl -n kafka run -i --rm kcat-med-discover-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.business-intel.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=med-url-discover-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.business-intel.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"med-url-discoverer"},"input_data":{}}
JSON

echo ""
echo "========================================="
echo "URL discovery triggered"
echo "========================================="
echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f deploy/business-intel --tail=100 | grep '$CORRELATION_ID'"
echo ""
echo "Check results:"
echo "  SELECT retailer_id, count(*) as listings, count(product_id) as matched"
echo "  FROM business_intel.med_retailer_listings"
echo "  GROUP BY retailer_id;"
echo ""
echo "  SELECT retailer_product_name, retailer_url"
echo "  FROM business_intel.med_retailer_listings"
echo "  WHERE match_method = 'url_discovery'"
echo "  ORDER BY created_at DESC LIMIT 20;"



for just one retailer:
{"action":"orchestrate","config":{"agent_type":"med-url-discoverer"},"input_data":{"retailer_id":"pet_drugs_online"}}


-- Sample of discovered products
SELECT retailer_product_name, retailer_url
FROM business_intel.med_retailer_listings
WHERE match_method = 'url_discovery' AND retailer_id = 'pet_drugs_online'
ORDER BY retailer_product_name
LIMIT 20;

-- Check the HTTP log for the discovery call
SELECT method, url, status_code, latency_ms
FROM http_request_log
WHERE action_name = 'med_discover_urls'
ORDER BY created_at DESC LIMIT 5;

-- How many category pages were scraped?
SELECT category_urls FROM business_intel.med_retailers WHERE id = 'pet_drugs_online';

----

SELECT retailer_id, count(*) as total,
       count(CASE WHEN match_method = 'url_discovery' THEN 1 END) as discovered
FROM business_intel.med_retailer_listings
GROUP BY retailer_id;

-- Sample of discovered products
SELECT retailer_product_name, retailer_url
FROM business_intel.med_retailer_listings
WHERE match_method = 'url_discovery' AND retailer_id = 'pet_drugs_online'
ORDER BY retailer_product_name
LIMIT 20;

-- Check the HTTP log for the discovery call
SELECT method, url, status_code, latency_ms
FROM http_request_log
WHERE action_name = 'med_discover_urls'
ORDER BY created_at DESC LIMIT 5;

-- How many category pages were scraped?
SELECT category_urls FROM business_intel.med_retailers WHERE id = 'pet_drugs_online';


