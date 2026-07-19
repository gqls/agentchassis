#!/usr/bin/env bash
set -euo pipefail
#
# Reassemble + redeploy all idea.uk pages with refreshed site chrome.
#
# WHY. bugs_open/018: idea.uk's header/footer rendered every link as href="" because the two
# per-site fork components declared field names absent from the chrome renderer's fixed
# vocabulary, and their templates had no {{if}} gates. Templates rewritten and gated by
# docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/sql/p3_01_chrome_templates_gated.sql.
#
# refresh_site_components=true is LOAD-BEARING here: without it the run reassembles pages from
# the STORED site_components.rendered_html, which still holds the old broken chrome
# (renderAndStoreSiteComponent skips a slot that already has rendered_html unless forced —
# render_site_components_action.go:468-483). With it, the chrome is re-rendered from the new
# templates first, then injected.
#
# EXPECT A WAIT. system.agent.generic.requests is single-partition/single-consumer; measured
# publish->orchestration latency on 2026-07-19 was 25-36 minutes under concurrent load. A missing
# orchestration row is almost certainly a queue, not a drop — see bugs_open/030 before concluding
# anything, and do NOT re-fire (that duplicates the work).
#
# VERIFY AGAINST THE DEPLOYED ARTEFACT, never the work-item status (this site has produced a
# 'failed' item for a page that rendered and deployed correctly):
#   curl -s https://idea.uk/ | grep -oE '<a href="[^"]*"' | sort | uniq -c | sort -rn
# Allow up to 5 further minutes after deploy for the box's sitesync.timer to pull.

AGENT_TYPE="rerender-pages"
SITE_ID="1244516d-014d-421c-88c6-090bb1e9552a"
DOMAIN="idea.uk"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-rerender-ideauk-$(date +%s) \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID -H message_id=$MESSAGE_ID \
  -H message_type=request -H client_id=demo_client \
  -H action=orchestrate -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","refresh_site_components":true}}
JSON

echo "CORRELATION_ID=$CORRELATION_ID  ORCHESTRATION_ID=$ORCHESTRATION_ID  REQUEST_ID=$REQUEST_ID TIMESTAMP=$TIMESTAMP"
