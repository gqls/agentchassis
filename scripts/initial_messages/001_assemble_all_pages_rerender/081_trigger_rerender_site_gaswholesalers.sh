#!/usr/bin/env bash
# Full site rerender for gaswholesalers.com.
# DOESN'T REWRITE CONTENT
# Why this:
#   After the single-page test confirms the skip path works for
#   wholesale-pricing-explained, this triggers rerender-pages across the
#   whole site to verify behaviour on the other 22 pages. rerender-pages
#   inserts one page_rerender work item per deployed page; build-dispatch-loop
#   claims and processes them one at a time.
#
# Expected outcome:
#   - x work items of type page_rerender created with status='triaged'
#     (one per page where build_status='deployed').
#   - Build-dispatch-loop claims them over the next several minutes.
#   - x produce git commits ("Rerender: <page>.html").
#   - x (wholesale-pricing-explained) goes via complete_skipped, no commit.
#   - chassis logs for the x should show how many sections each page kept
#     vs skipped (new logging from getPageSections empty-section filter).
#
# Notes:
#   refresh_site_components=false  — header/footer/head are x days old but
#                                    not stale enough to need a re-render.
#                                    We only want pages re-assembled from
#                                    their existing components.
#
#   This will produce ~x commits to gqls/sites, each firing the b2 sync
#   workflow and a Cloudflare cache purge. Plan for ~15 min of GitHub
#   Actions runtime.
-------------------------------------------------
# reset page components true or false in the JSON

set -euo pipefail

AGENT_TYPE="rerender-pages"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

AGENT_TYPE="rerender-pages"
SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
DOMAIN="ai-agent-orchestration.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site:           ${DOMAIN} (${SITE_ID})"
echo "  Correlation:    ${CORRELATION_ID}"
echo "  Orchestration:  ${ORCHESTRATION_ID}"
echo "  Request:        ${REQUEST_ID}"
echo "  Time:           ${TIMESTAMP}"
echo ""

kubectl -n kafka run -i --rm kcat-rerender-site-$(date +%s) \
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
{"action":"orchestrate","config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN","refresh_site_components":true}}
JSON

echo ""
echo "=== Monitor ==="
echo ""
echo "Within ~30s, expect ~22 fresh page_rerender items in status='triaged':"
echo "  psql -c \"SELECT status, COUNT(*) FROM site_work_items"
echo "           WHERE site_id = '${SITE_ID}'"
echo "             AND item_type = 'page_rerender'"
echo "             AND created_at > NOW() - INTERVAL '5 minutes'"
echo "           GROUP BY status ORDER BY status;\""
echo ""
echo "Watch them transition to claimed/complete (refresh every 5s):"
echo "  watch -n 5 \"psql -c 'SELECT status, COUNT(*) FROM site_work_items"
echo "                       WHERE site_id = '\\\"${SITE_ID}\\\"'"
echo "                         AND item_type = ''\\\"page_rerender\\\"''"
echo "                         AND created_at > NOW() - INTERVAL '\\\"10 minutes\\\"'"
echo "                       GROUP BY status ORDER BY status;'\""
echo ""
echo "Tail chassis logs for the empty-section filter and skip outputs:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 -f | \\"
echo "    grep -E 'assemblePage|getPageSections|wholesale-pricing-explained'"
echo ""
echo "Watch git for commit progression (in a separate terminal):"
echo "  cd ~/projects/sites && while true; do"
echo "    git fetch origin --quiet"
echo "    git log -10 --pretty='%h %ad %s' --date=iso-strict origin/master -- domain.com/ | head -10"
echo "    echo '---'; sleep 15"
echo "  done"
echo ""
echo "=== After completion ==="
echo ""
echo "Final check — should see 21 success, 1 skip in the result JSON:"
echo "  psql -c \"SELECT"
echo "             COUNT(*) FILTER (WHERE result->'rendered_page'->>'skipped' = 'true') AS skipped,"
echo "             COUNT(*) FILTER (WHERE result->'rendered_page'->>'skipped' IS NULL OR"
echo "                                    result->'rendered_page'->>'skipped' = 'false') AS rendered,"
echo "             COUNT(*) AS total"
echo "           FROM site_work_items"
echo "           WHERE site_id = '${SITE_ID}'"
echo "             AND item_type = 'page_rerender'"
echo "             AND status = 'complete'"
echo "             AND created_at > NOW() - INTERVAL '30 minutes';\""
echo ""
