#!/usr/bin/env bash
# rerender-index-vonc.sh — assemble-only rerender + deploy of the vonc.com index.
#
# Mirrors page-build-handler's deploy_page: spawn page-rerender, call it with
# {domain, page_id, site_id}. page-rerender then runs its own workflow:
#   check_rerender_mode (no spec.reason → NOT image_landed/section_data_resolved
#   → render_page) → rerender_single_page (assemble stored components) →
#   check_skipped → deploy_page (git_commit to 'sites') → update_status.
# NO content rebuild — this only re-assembles the existing page_components and deploys.
#
# PRE-REQS (in order):
#   1. Patched rerender_single_page_action.go deployed (data-runtime-fill exemption).
#   2. Marker SQL run (data-runtime-fill on provocation-card template + index rendered_html).
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Rerender index (assemble-only + deploy)  vonc.com"
echo "  Correlation: $CORRELATION_ID"
echo "========================================="
echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo "Timestamp =$TIMESTAMP"
echo ""

kubectl -n kafka run -i --rm "kcat-rerender-index-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" \
  -H "client_id=$CLIENT_ID" \
  -H "action=process" \
  -H "sender_agent_type=cli" \
  -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" \
  -H "timestamp=$TIMESTAMP" <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerender","processing_mode":"orchestrator","timeout_seconds":240,"steps":{"spawn_rerender":{"action":"spawn_agent","config":{"role":"page_renderer","agent_type":"page-rerender"},"output_field":"rerender_agent","next_step":"call_rerender","description":"Spawn page-rerender"},"call_rerender":{"action":"call_agent","config":{"agent_type":"page-rerender","target_role":"page_renderer","input_mapping":{"domain":"input_data.domain","page_id":"input_data.page_id","site_id":"input_data.site_id"},"timeout_seconds":200},"output_field":"rerender_result","next_step":"complete","description":"Assemble stored components + deploy (mirrors deploy_page)"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result"]},"description":"Rerender complete"}}}},"input_data":{"domain":"vonc.com","page_id":"e4b3b195-919f-45ad-854e-201d3e846ea8","site_id":"9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"}}
JSON

echo ""
echo "Monitor:  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=300 | grep \"$CORRELATION_ID\""
echo "Expect in logs:"
echo "  getPageSections: filtered empty sections  skipped=1 kept=5   (lobby-grid still skipped; provocation-card now KEPT)"
echo "  RerenderSinglePageAction: Complete"
echo "  git_commit / deploy: Rerender: provocations/index.html  (commit to 'sites')"
