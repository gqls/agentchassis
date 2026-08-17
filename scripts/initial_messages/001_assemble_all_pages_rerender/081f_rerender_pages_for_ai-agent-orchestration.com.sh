#!/usr/bin/env bash
set -euo pipefail
#
# AGENT_TYPE="rerender-pages"
#
# Propagates migration 456 (foreground --color-primary -> --color-primary-ink) into
# ai-agent-orchestration.com's rendered pages. The migration changed 12
# content_components.html_template rows; rendered placements keep the OLD html until
# re-rendered, so this is the step that makes the fix visible.
#
# ⚠ TWO TRAPS BEFORE YOU RUN IT:
#   1. No orchestration dispatch within ~300s of an agent-chassis (re)start — the
#      spawn is silently dropped. Check:
#        kubectl -n ai-persona-system get pods -l app=agent-chassis \
#          -o jsonpath='{range .items[*]}{.status.startTime}{"\n"}{end}'
#   2. `kcat -P` can exit 0 having published NOTHING. Do not treat a clean exit as
#      evidence; confirm with the orchestration_states query printed at the end.
#
# ⚠ Pages whose components have NULL content_data CANNOT be re-rendered — there is
# nothing to rebuild the section from (bugs_closed/194). On this site that is
# `pricing` (5/5) and 7 others; they need a framework rebuild, not this.

SITE_ID="2a8ebf9c-20a2-4c39-b191-840b012371da"
DOMAIN="ai-agent-orchestration.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

kubectl -n kafka run -i --rm kcat-rerender-aiao-$(date +%s) \
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
{"action":"orchestrate","config":{"agent_type":"rerender-pages"},"input_data":{"site_id":"${SITE_ID}","domain":"${DOMAIN}","refresh_site_components":true}}
JSON

echo "CORRELATION_ID=$CORRELATION_ID  ORCHESTRATION_ID=$ORCHESTRATION_ID  REQUEST_ID=$REQUEST_ID TIMESTAMP=$TIMESTAMP"
echo
echo "CONFIRM IT LANDED (a clean kcat exit is NOT evidence):"
echo "  SELECT current_step, status, created_at FROM orchestration_states"
echo "   WHERE orchestration_id = '${ORCHESTRATION_ID}';"
echo "VERIFY AT THE ARTEFACT afterwards, never at the status:"
echo "  python3 scripts/render_audit.py https://${DOMAIN}/index.html https://${DOMAIN}/about.html"
