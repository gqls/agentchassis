#!/usr/bin/env bash
# Single-page rerender test — index.html
# Tests the deploy_page fix (files_field) by rendering a page known to have
# a component with js_content (latest-news), so the deploy should produce
# both /index.html AND /tools/assets/latest-news.js in one commit.

set -euo pipefail

AGENT_TYPE="page-rerender"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"
PAGE_ID="4ff0e0ff-fab2-423e-a59c-b9de4674a84f"
PAGE_NAME="index"
FILENAME="index.html"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site:          ${DOMAIN} (${SITE_ID})"
echo "  Page:          ${PAGE_NAME} (${PAGE_ID})"
echo "  Filename:      ${FILENAME}"
echo "  Correlation:   ${CORRELATION_ID}"
echo "  Orchestration: ${ORCHESTRATION_ID}"
echo "  Request:       ${REQUEST_ID}"
echo "  Time:          ${TIMESTAMP}"
echo ""

kubectl -n kafka run -i --rm kcat-rerender-single-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=$CLIENT_ID \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","page_id":"${PAGE_ID}"}}
JSON

echo ""
echo "=== Verify ==="
echo "  # Within ~30s, the orchestration should complete"
echo "  psql -c \"SELECT orchestration_id, status, current_step,"
echo "             jsonb_pretty(collected_data->'deploy_result'->'data'->'files') AS deployed,"
echo "             collected_data->'deploy_result'->'data'->>'files_count' AS file_count"
echo "           FROM orchestration_states"
echo "           WHERE orchestration_id = '${ORCHESTRATION_ID}';\""
echo ""
echo "  # And check git for the new commit"
echo "  cd ~/projects/sites && git fetch origin --quiet && \\"
echo "    git log -3 --pretty='%h %ad %s' --date=iso-strict origin/master -- gaswholesalers.com/"