#!/bin/bash
# ============================================================================
# submit-domain.sh — Submit a new domain to the build pipeline
#
# Usage:
#   ./submit-domain.sh <domain> [email] [phone]
#
# Examples:
#   ./submit-domain.sh dartsonline.com darts@contactforsales.com "+44 (0) 7934 524 911"
#   ./submit-domain.sh gaswholesalers.com info@gaswholesalers.com
#   ./submit-domain.sh mybusiness.co.uk
#
# Sends a Kafka message to domain-submitter which:
#   1. Creates the site record
#   2. Stores email/phone on the site
#   3. Creates needs_domain_research work item
#   4. build-pipeline-trigger picks up the work item on next cycle
#
# No SQL, no objective text needed. The classifier figures out what the site
# should be from the domain name and web research.
# ============================================================================

DOMAIN="${1:?Usage: $0 <domain> [email] [phone]}"
EMAIL="${2:-}"
PHONE="${3:-}"
---


DOMAIN="dartsonline.com"
EMAIL="darts@contactforsales.com"
PHONE="+44 (0) 7934 524 911"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Submitting Domain: ${DOMAIN}"
echo "========================================="
echo "  Email: ${EMAIL:-not set}"
echo "  Phone: ${PHONE:-not set}"
echo "  Correlation: ${CORRELATION_ID}"
echo "========================================="
echo ""

# Build input_data JSON — only include non-empty fields
INPUT_DATA="{\"domain\": \"${DOMAIN}\""
if [ -n "$EMAIL" ]; then
    INPUT_DATA="${INPUT_DATA}, \"email\": \"${EMAIL}\""
fi
if [ -n "$PHONE" ]; then
    INPUT_DATA="${INPUT_DATA}, \"phone\": \"${PHONE}\""
fi
INPUT_DATA="${INPUT_DATA}}"

kubectl -n kafka run -i --rm kcat-submit-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=submit-${DOMAIN}-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=$CLIENT_ID \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.generic.responses <<JSON
{"action":"orchestrate","config":{"agent_type":"domain-submitter"},"input_data":${INPUT_DATA}}
JSON

echo ""
echo "========================================="
echo "Submitted: ${DOMAIN}"
echo "========================================="
echo ""
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""
echo "What happens next:"
echo "  1. domain-submitter creates site record + needs_domain_research item"
echo "  2. build-pipeline-trigger finds it and dispatches"
echo "  3. domain-research-classifier researches the domain"
echo "  4. build-site-planner creates the site plan"
echo "  5. dispatch loop builds pages"
echo ""
echo "Check progress:"
echo "  SELECT id, domain, status FROM sites WHERE domain = '${DOMAIN}';"
echo ""
echo "  SELECT wi.item_type, wi.handler_agent, wi.status"
echo "  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id"
echo "  WHERE s.domain = '${DOMAIN}' ORDER BY wi.priority;"
echo ""
echo "  SELECT aspect, LEFT(data::text, 80)"
echo "  FROM site_specs ss JOIN sites s ON s.id = ss.site_id"
echo "  WHERE s.domain = '${DOMAIN}' AND ss.is_current = true;"
echo ""
echo "  SELECT wi.item_type, wi.handler_agent, LEFT(wi.error, 60)"
echo "  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id"
echo "  WHERE s.domain = '${DOMAIN}' AND wi.status = 'blocked';"