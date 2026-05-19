#!/usr/bin/env bash
# page-rebuild trigger for gaswholesalers.com.
#  REWRITES CONTENT
# This regenerates content for pages flagged as needs_rebuild. The agent
# workflow:
#   ensure_site_record → load_rebuild_context → select_style_collection
#   → get_pages_to_rebuild (build_statuses=["needs_rebuild"])
#   → loop: content-writer → content-reviewer → assemble → deploy
#   → trigger_site_deploy
#
# Cost note: LLM-heavy. Each page invokes page-content-writer (full
# article draft) and content-reviewer (review pass). For 7 pages that's
# 14 LLM calls minimum, plus deploy roundtrips. Expect ~10–15 minutes
# wall time and meaningful API spend.
#
# Pages targeted (as of 2026-05-12):
#   faq                                    (4 components, empty FAQ items — needs regeneration)
#   news                                   (3 components)
#   supply-terms-and-eligibility           (6 components)
#   tool-breakeven-volume-calculator-guide (0 components — fresh content)
#   tool-fuel-budget-forecaster-guide      (0 components — fresh content)
#   tool-gas-unit-converter-guide          (3 components)
#   fuel-pricing-framework                 (0 components, currently 'planned' — promoted to needs_rebuild first)

set -euo pipefail

AGENT_TYPE="page-rebuild"
SITE_ID="5fe15466-4e2e-4ff2-981e-98c1b7074002"
DOMAIN="gaswholesalers.com"

# Pre-step: promote fuel-pricing-framework from 'planned' to 'needs_rebuild'
# so the page-rebuild workflow picks it up. Without this, the get_pages_to_rebuild
# step (which filters on build_statuses=["needs_rebuild"]) skips it.
#
# Run this against clients_db before firing the trigger:
#
#   psql -c "UPDATE pages
#            SET build_status = 'needs_rebuild',
#                updated_at = NOW()
#            WHERE site_id = '5fe15466-4e2e-4ff2-981e-98c1b7074002'
#              AND name = 'fuel-pricing-framework'
#              AND build_status = 'planned'
#            RETURNING id, name, build_status;"
#
# Confirm 1 row returned. If 0, the page is already in needs_rebuild or
# a different state — check and adjust before proceeding.

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
echo "Reminder: promote fuel-pricing-framework to needs_rebuild BEFORE this trigger."
echo "Press Ctrl+C now if you haven't."
sleep 3
echo ""

kubectl -n kafka run -i --rm kcat-page-rebuild-$(date +%s) \
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
{"action":"orchestrate","config":{"agent_type":"page-rebuild"},"input_data":{"site_id":"5fe15466-4e2e-4ff2-981e-98c1b7074002","domain":"gaswholesalers.com"}}
JSON

echo ""
echo "=== Monitor ==="
echo ""
echo "1. Confirm orchestration picked up:"
echo "  psql -c \"SELECT orchestration_id, status, current_step, last_activity"
echo "           FROM orchestration_states"
echo "           WHERE owner_agent_type = 'page-rebuild'"
echo "             AND site_id = '${SITE_ID}'"
echo "             AND created_at > NOW() - INTERVAL '5 minutes'"
echo "           ORDER BY created_at DESC LIMIT 1;\""
echo ""
echo "2. Track pages as they transition needs_rebuild → built → deployed:"
echo "  watch -n 30 \"psql -c \\\"SELECT name, build_status, deployed_at"
echo "                          FROM pages WHERE site_id = '${SITE_ID}'"
echo "                            AND build_status IN ('needs_rebuild','built','deploying','deployed')"
echo "                            AND name IN ('faq','news','supply-terms-and-eligibility',"
echo "                                         'tool-breakeven-volume-calculator-guide',"
echo "                                         'tool-fuel-budget-forecaster-guide',"
echo "                                         'tool-gas-unit-converter-guide',"
echo "                                         'fuel-pricing-framework')"
echo "                          ORDER BY name;\\\"\""
echo ""
echo "3. Chassis logs scoped to this orchestration:"
echo "  kubectl -n ai-persona-system logs -l app=agent-chassis --tail=200 -f | \\"
echo "    grep -E 'page-rebuild|page-content-writer|content-reviewer|${CORRELATION_ID:0:8}'"
echo ""
echo "4. Watch git for the rebuild commits:"
echo "  cd ~/projects/sites && while true; do git fetch origin --quiet; \\"
echo "    git log -10 --pretty='%h %ad %s' --date=iso-strict \\"
echo "    origin/master -- gaswholesalers.com/ | head -10; \\"
echo "    echo '---'; sleep 30; done"
echo ""
echo "Expected end state:"
echo "  - All 7 pages: build_status='deployed', recent deployed_at"
echo "  - 7 git commits with subject 'Rebuild: <page>.html' (or similar)"
echo "  - faq.html has actual FAQ Q&As populated"
echo "  - fuel-pricing-framework.html exists in git (footer link no longer 404s)"