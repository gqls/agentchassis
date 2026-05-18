#!/usr/bin/env bash
# Trigger design-discovery-agent for robot-hands.com.
# Run AFTER the new plan exists (build-site-planner has run).
# This emits needs_imagery work items for plan rows without matching assets.

set -u

AGENT_TYPE="design-discovery-agent"
SITE_ID="00ff3af5-dad8-4770-9f70-3edc267a3c92"
DOMAIN="robot-hands.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "=== Triggering: ${AGENT_TYPE} ==="
echo "  Site: ${DOMAIN} (${SITE_ID})"
echo "  Correlation: ${CORRELATION_ID}"
echo ""
echo "Expected: discovery will compare site_plan_imagery (current plan) against"
echo "assets and emit a needs_imagery work item for each plan row without a"
echo "matching active asset. Based on cross-reference, this should emit 8 items:"
echo "  - hero_home, brand_hero_canonical (2 heroes)"
echo "  - icon_catalog, icon_matchmatrix, icon_payload_calc,"
echo "    icon_cycle_time, icon_selection_guide, icon_learning (6 icons)"
echo ""

kubectl -n kafka run -i --rm "kcat-discovery-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
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
{"action":"orchestrate","config":{"agent_type":"${AGENT_TYPE}"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "Check what was emitted (run ~30 seconds after triggering):"
echo "  SELECT spec::jsonb->>'asset_key' AS asset_key,"
echo "         spec::jsonb->>'kind' AS kind,"
echo "         status, pipeline, priority, created_at"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND item_type = 'needs_imagery'"
echo "    AND created_at > now() - interval '5 minutes'"
echo "  ORDER BY priority, created_at;"
echo ""
echo "Expected: 8 rows, status='detected' (will become 'triaged' after audit runs),"
echo "pipeline='build' (assuming the unfulfilled_imagery_plan emit-default fix landed)."
echo "If pipeline='design' on these, the bug from item 25 is still in place and you'll"
echo "need to UPDATE them to 'build' before dispatch can claim them."
echo ""
echo "Next step after this completes: trigger design-audit-agent to triage them."