#!/bin/bash
# ============================================================================
# rerender_pages.sh — re-render specific pages from stored content_data and deploy.
#
# Drives page-rerender directly, one page per call, so a single stuck page does
# not block the rest (the rerender-site sequential loop is fragile — it stalled
# mid-run during the leopardess rebuild).
#
# KEY: page-rerender only REGENERATES section HTML from content_data when
# input_data.spec.reason is 'section_data_resolved' (or 'image_landed'). Without
# it, check_rerender_mode takes the else branch and just re-assembles the STORED
# (old) section HTML. And rerender_page_sections requires input_data.spec.page_name
# (the page's `name`, not its url). Both are set below.
#
# Usage: ./rerender_pages.sh <site_id> <domain> <page_name> [<page_name> ...]
# ============================================================================
set -euo pipefail
S="${1:?site_id}"; DOMAIN="${2:?domain}"; shift 2
[ $# -gt 0 ] || { echo "give at least one page_name"; exit 2; }

for PAGE_NAME in "$@"; do
  CID=$(cat /proc/sys/kernel/random/uuid); OID=$(cat /proc/sys/kernel/random/uuid)
  PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"%s","domain":"%s","spec":{"reason":"section_data_resolved","page_name":"%s"}}}' "$S" "$DOMAIN" "$PAGE_NAME")
  echo "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-rr-${PAGE_NAME//[^a-z0-9]/}-$(date +%s)" \
    --image=edenhill/kcat:1.7.1 --restart=Never -- \
    kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
    -t system.agent.generic.requests \
    -H correlation_id=$CID -H orchestration_id=$OID -H request_id=$(cat /proc/sys/kernel/random/uuid) \
    -H message_id=$(cat /proc/sys/kernel/random/uuid) \
    -H orchestration_name=rr-${PAGE_NAME}-$(date +%H%M%S) -H step_name=start -H client_id=demo_client \
    -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
    -H responses_topic=system.agent.generic.responses >/dev/null 2>&1
  echo "$PAGE_NAME  ->  $CID"
done
