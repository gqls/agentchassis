#!/usr/bin/env bash
# Single-page rerender test for wholesale-pricing-explained on gaswholesalers.com.
#
# Why this test:
#   wholesale-pricing-explained has 0 page_components in the database (Q3
#   earlier showed components=0, rendered_components=0). Before the fix,
#   assemblePage would build a stub (header + empty <main> + footer) and
#   ship it to git. After the fix, the new predicate
#       if len(sections) == 0 { return "", nil }
#   should fire, RerenderSinglePageAction returns skipped=true, the
#   page-rerender workflow routes to complete_skipped, and neither git_commit
#   nor update_page_status runs.
#
# Expected outcome:
#   1. Chassis logs contain: "assemblePage: page has no sections, skipping"
#                            with page_name=wholesale-pricing-explained
#   2. No new commit on origin/master under gaswholesalers.com/
#   3. pages row for wholesale-pricing-explained: deployed_at unchanged,
#      build_status unchanged (still 'deployed').
#   4. A new site_work_items row of type 'page_rerender' WAS created and
#      reached status='complete' via the skipped branch. result JSON contains
#      "skipped": true.

set -euo pipefail

AGENT_TYPE="page-rerender"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"
PAGE_ID="08e34599-a130-417d-90d9-76370484d03a"
PAGE_NAME="wholesale-pricing-explained"
FILENAME="wholesale-pricing-explained.html"

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
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"site_id":"5fe15466-4e2e-4ff2-981e-98c1b7074002","domain":"gaswholesalers.com","page_id":"08e34599-a130-417d-90d9-76370484d03a","page_name":"wholesale-pricing-explained","filename":"wholesale-pricing-explained.html"}}
JSON

echo ""
echo "=== Monitor ==="
echo ""
echo "Watch for the skip log (should appear within ~10s):"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 -f | \\"
echo "    grep -E 'assemblePage|page-rerender|${CORRELATION_ID:0:8}'"
echo ""
echo "Verify NO new git commit (last commit on this file should be the restore):"
echo "  cd ~/projects/sites && git fetch origin --quiet && \\"
echo "    git log -3 --pretty='%h %ad %s' --date=iso-strict origin/master -- gaswholesalers.com/${FILENAME}"
echo ""
echo "Verify page DB row unchanged (deployed_at should be 2026-05-06 17:54:56):"
echo "  psql -c \"SELECT name, build_status, deployed_at, updated_at"
echo "           FROM pages"
echo "           WHERE id = '${PAGE_ID}';\""
echo ""
echo "Verify the work item completed via the skipped branch:"
echo "  psql -c \"SELECT id, status, claimed_by,"
echo "                  result->'rendered_page'->'skipped' AS skipped,"
echo "                  result->'rendered_page'->'reason' AS reason,"
echo "                  created_at, updated_at"
echo "           FROM site_work_items"
echo "           WHERE site_id = '${SITE_ID}'"
echo "             AND item_type = 'page_rerender'"
echo "             AND created_at > NOW() - INTERVAL '5 minutes'"
echo "           ORDER BY created_at DESC LIMIT 1;\""
echo ""
echo "If 'skipped' = true and no new git commit, the fix works."
echo "If a new commit appeared, the chassis image rollout didn't take —"
echo "  check pod image SHAs:"
echo "    kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name}{\"\\t\"}{.spec.containers[0].image}{\"\\n\"}{end}'"