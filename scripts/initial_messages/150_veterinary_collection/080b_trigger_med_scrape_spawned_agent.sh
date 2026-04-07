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
{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":5,"retailer_id":"viovet"}}
JSON

echo ""
echo "Monitor business-intel (spawn step):"
echo "  kubectl -n ai-persona-system logs -f deploy/business-intel --tail=20 | grep -E 'spawn|med-price'"
echo ""
echo "Monitor spawned pod (actual work):"
echo "  kubectl -n ai-persona-system logs -f -l app=dynamic-agent --tail=50 | grep MedScrape"


------------------------------------------------------------------------------------------

# all retailers
{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":$BATCH_SIZE}}

# individual retailers
{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":50,"retailer_id":"animed_direct"}}
{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":10,"retailer_id":"hyperdrug"}}
{"action":"orchestrate","config":{"agent_type":"med-price-scrape-orchestrator"},"input_data":{"batch_size":5,"retailer_id":"viovet"}}


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

---------------

-- med_pricing_cleanup_and_status.sql
-- Run this to clean junk URLs and check pipeline health

-- ============================================================================
-- 1. Clean known junk patterns across ALL retailers
-- ============================================================================

-- First delete evidence referencing junk listings
DELETE FROM business_intel.med_scrape_evidence
WHERE listing_id IN (
    SELECT id FROM business_intel.med_retailer_listings
    WHERE retailer_url LIKE '%/knowledgebase%'
       OR retailer_url LIKE '%/breed_information%'
       OR retailer_url LIKE '%/pages.html%'
       OR retailer_url LIKE '%/shopping_basket%'
       OR retailer_url LIKE '%/advanced_search%'
       OR retailer_url LIKE '%/modern-slavery%'
       OR retailer_url LIKE '%/sitemap%'
       OR retailer_url LIKE '%/bslp/%'
       OR retailer_url LIKE '%/flp/%'
       OR retailer_url LIKE '%/flp%'
       OR retailer_url LIKE '%clearance-sale%'
       OR retailer_url LIKE '%/saddlery%'
       OR retailer_url LIKE '%/stable-yard%'
       OR retailer_url LIKE '%/reflective-%'
       OR retailer_url LIKE '%/rider-%'
       OR retailer_url LIKE '%/horse-%'
       OR retailer_url LIKE '%newsletter%'
       OR retailer_url LIKE '%/modals/%'
);

-- Then delete the listings
DELETE FROM business_intel.med_retailer_listings
WHERE retailer_url LIKE '%/knowledgebase%'
   OR retailer_url LIKE '%/breed_information%'
   OR retailer_url LIKE '%/pages.html%'
   OR retailer_url LIKE '%/shopping_basket%'
   OR retailer_url LIKE '%/advanced_search%'
   OR retailer_url LIKE '%/modern-slavery%'
   OR retailer_url LIKE '%/sitemap%'
   OR retailer_url LIKE '%/bslp/%'
   OR retailer_url LIKE '%/flp/%'
   OR retailer_url LIKE '%/flp%'
   OR retailer_url LIKE '%clearance-sale%'
   OR retailer_url LIKE '%/saddlery%'
   OR retailer_url LIKE '%/stable-yard%'
   OR retailer_url LIKE '%/reflective-%'
   OR retailer_url LIKE '%/rider-%'
   OR retailer_url LIKE '%/horse-%'
   OR retailer_url LIKE '%newsletter%'
   OR retailer_url LIKE '%/modals/%';

-- ============================================================================
-- 2. Status check
-- ============================================================================

-- Listings status
SELECT r.id as retailer,
       count(l.id) as listings,
       count(l.id) FILTER (WHERE l.last_scraped_at IS NOT NULL) as scraped,
       count(l.id) FILTER (WHERE l.last_scraped_at IS NULL) as pending
FROM business_intel.med_retailers r
LEFT JOIN business_intel.med_retailer_listings l ON l.retailer_id = r.id
WHERE r.is_active = true
GROUP BY r.id
ORDER BY r.id;

-- Price coverage
SELECT retailer_id,
       count(DISTINCT listing_id) as products_with_prices,
       count(*) as total_variants,
       round(avg(price)::numeric, 2) as avg_price,
       min(price) as min_price,
       max(price) as max_price
FROM business_intel.med_price_snapshots
GROUP BY retailer_id
ORDER BY retailer_id;

-- Scrape success rate
SELECT retailer_id,
       count(*) as total_scraped,
       count(*) FILTER (WHERE variants_found > 0) as had_prices,
       count(*) FILTER (WHERE variants_found = 0) as no_prices,
       round(100.0 * count(*) FILTER (WHERE variants_found > 0) / count(*), 1) as success_pct
FROM business_intel.med_scrape_evidence
GROUP BY retailer_id
ORDER BY retailer_id;

-- LLM fallback stats
SELECT success, count(*), round(avg(latency_ms)) as avg_latency_ms
FROM llm_call_log
WHERE provider = 'ollama'
AND step_name = 'scrape_prices'
GROUP BY success;
