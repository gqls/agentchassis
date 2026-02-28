#!/bin/bash
# ============================================================================
# IMPROVEMENT LOOP — Manual Trigger
# ============================================================================
# Runs post-build quality checks and dispatches fixes for a site.
#
# What it does:
#   1. Spawns quality-discovery-agent     (broken nav, placeholder contact, generic theme)
#   2. Spawns design-discovery-agent      (undeployed assets, missing CSS, duplicate palette)
#   3. Spawns completeness-discovery-agent (empty sections)
#   4. Triages findings (detected → triaged, domain → build)
#   5. If issues found: inserts needs_rerender, dispatches fixes via build-dispatch-loop
#
# Usage:
#   ./060_trigger_improvement_loop.sh                              # defaults to gaswholesalers.com
#   ./060_trigger_improvement_loop.sh <site_id> <domain>           # specific site
#   SITE_ID=xxx DOMAIN=yyy ./060_trigger_improvement_loop.sh       # env vars
# ============================================================================
set -euo pipefail

SITE_ID="${1:-${SITE_ID:-5fe15466-4e2e-4ff2-981e-98c1b7074002}}"
DOMAIN="${2:-${DOMAIN:-gaswholesalers.com}}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Improvement Loop"
echo "========================================="
echo "  Domain:         $DOMAIN"
echo "  Site ID:        $SITE_ID"
echo "  Correlation ID: $CORRELATION_ID"
echo "========================================="

kubectl -n kafka run -i --rm kcat-improve-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H message_type=request \
  -H client_id=$CLIENT_ID \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_improvement_loop","steps":{"spawn_improvement_loop":{"action":"spawn_agent","config":{"role":"improver","agent_type":"improvement-loop"},"description":"Spawn improvement loop agent","next_step":"call_improvement_loop","output_field":"improver_agent"},"call_improvement_loop":{"action":"call_agent","config":{"agent_type":"improvement-loop","target_role":"improver","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain"},"timeout_seconds":1800},"description":"Run improvement loop — discover issues, triage, dispatch fixes, rerender","next_step":"complete","output_field":"improvement_result"},"complete":{"action":"complete_workflow","config":{"output_fields":["improvement_result"]},"description":"Improvement loop complete"}}}},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}"}}
JSON

echo ""
echo "========================================="
echo "Improvement loop triggered"
echo "========================================="
echo ""
echo "Monitor discovery phase:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=quality-discovery-agent --tail=20"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=design-discovery-agent --tail=20"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=completeness-discovery-agent --tail=20"
echo ""
echo "Monitor fix dispatch:"
echo "  kubectl -n ai-persona-system logs -f -l agent-type=build-dispatch-loop --tail=30"
echo ""
echo "Check findings:"
echo "  psql -c \"SELECT item_type, status, severity, summary FROM site_work_items WHERE site_id='${SITE_ID}' AND source='discovery' ORDER BY created_at DESC LIMIT 20;\""
echo ""
echo "Overall progress:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '${CORRELATION_ID}'"
echo ""
echo "CORRELATION_ID=$CORRELATION_ID"

kubectl -n ai-persona-system logs --tail=300 -l agent-type=quality-discovery-agent -f | tee logs-quality-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=design-discovery-agent -f | tee logs-design-discovery-agent.json
kubectl -n ai-persona-system logs --tail=300 -l agent-type=completeness-discovery-agent -f | tee logs-completeness-discovery-agent.json



