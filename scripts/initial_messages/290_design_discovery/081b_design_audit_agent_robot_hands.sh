#!/usr/bin/env bash
# Trigger design-audit-agent for robot-hands.com.
# Promotes the 8 currently-detected needs_imagery work items to status='triaged'
# so dispatch can claim them.

set -u

AGENT_TYPE="design-audit-agent"
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
echo "Pre-flight — confirm what's in 'detected' on this site:"
echo "  SELECT spec::jsonb->>'asset_key' AS asset_key, status, pipeline, priority"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND item_type = 'needs_imagery'"
echo "    AND status = 'detected'"
echo "  ORDER BY priority;"
echo "  -- Expected: 8 rows (2 heroes + 6 icons), pipeline='build' (after SQL migration)"
echo ""
echo "Pre-flight — confirm no stuck claimed items blocking the site:"
echo "  SELECT id, item_type, status, claimed_at,"
echo "         EXTRACT(EPOCH FROM (now() - claimed_at))::int AS seconds_claimed"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}' AND status = 'claimed';"
echo "  -- If rows return with seconds_claimed > 900, reset them first"
echo ""

kubectl -n kafka run -i --rm "kcat-audit-$(date +%s)" \
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
echo "Expected sequence:"
echo "  1. design-audit-agent ensure_site_record"
echo "  2. spawns visual-design-auditor → runs LLM audit → COMPLETED"
echo "  3. spawns content-quality-auditor → runs LLM audit → COMPLETED"
echo "  4. triage_detected_items action runs → promotes detected → triaged"
echo "  5. design-audit-agent COMPLETED"
echo ""
echo "Verify (run ~60s after trigger):"
echo "  SELECT spec::jsonb->>'asset_key' AS asset_key, status, pipeline, priority"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND item_type = 'needs_imagery'"
echo "    AND created_at > '2026-05-17'"
echo "  ORDER BY status, priority;"
echo "  -- Expected: 8 rows at status='triaged'"
echo ""
echo "After triage, dispatch picks them up automatically on the next 30s tick of"
echo "build-pipeline-trigger. Watch:"
echo ""
echo "  SELECT spec::jsonb->>'asset_key' AS asset_key, status, claimed_at,"
echo "         EXTRACT(EPOCH FROM (now() - claimed_at))::int AS seconds_claimed"
echo "  FROM site_work_items"
echo "  WHERE site_id = '${SITE_ID}'"
echo "    AND item_type = 'needs_imagery'"
echo "    AND status IN ('claimed','complete')"
echo "  ORDER BY status, claimed_at;"
echo ""
echo "Then look at the resulting icon assets — the actual test of the planner prompt rewrite."