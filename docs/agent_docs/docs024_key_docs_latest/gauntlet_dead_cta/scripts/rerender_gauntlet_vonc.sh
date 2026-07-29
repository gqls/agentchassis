#!/usr/bin/env bash
# rerender_gauntlet_vonc.sh — assemble-only rerender + deploy of vonc.com/tools/gauntlet.
#
# Mirrors page-build-handler's deploy_page via the direct orchestrator envelope
# (086 pattern: spawn_agent + call_agent, action=process). page-rerender then runs
#   check_rerender_mode -> rerender_single_page (assemble stored page_components)
#   -> check_skipped -> deploy_page (git_commit to 'sites') -> update_status.
# NO content rebuild: this re-assembles page_components.rendered_html and deploys.
#
# WHY THIS EXISTS RATHER THAN REUSING 210_vonc_trigger/085_*:
#   085 publishes with `kubectl run -i --rm ... -- kcat -P <<JSON`. If the
#   container starts before stdin is attached, kcat sees EOF, produces NOTHING and
#   exits 0 — the wrapper then prints a correlation id as if it had worked.
#   Measured 2026-07-26: 1 publish in 5 landed. This uses the hardened form —
#   payload in the container COMMAND, --command to beat the kcat ENTRYPOINT, and
#   `&& echo PUBLISH_OK` so a silent drop is visible.
#
# PRE-REQS:
#   1. page_components.rendered_html already holds the new markup (rerender
#      assembles from THAT column, not from content_components.html_template —
#      rerender_single_page_action.go:163,232,511). Write both or they diverge.
#   2. Not within ~300s of a chassis pod restart, or the spawn is silently dropped.
set -u

SITE_ID="9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"
PAGE_ID="ecb637c1-845f-46bf-b174-9c92a43f9586"   # tool-gauntlet, /tools/gauntlet/index.html
DOMAIN="vonc.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

BODY="{\"headers\":{\"correlation_id\":\"${CORRELATION_ID}\",\"orchestration_id\":\"${ORCHESTRATION_ID}\",\"request_id\":\"${REQUEST_ID}\",\"message_id\":\"${MESSAGE_ID}\",\"message_type\":\"request\",\"client_id\":\"${CLIENT_ID}\",\"action\":\"process\",\"sender\":{\"agent_id\":\"cli-user\",\"agent_type\":\"cli\",\"pod_name\":\"cli\"},\"timestamp\":\"${TIMESTAMP}\"},\"config\":{\"workflow\":{\"start_step\":\"spawn_rerender\",\"processing_mode\":\"orchestrator\",\"timeout_seconds\":240,\"steps\":{\"spawn_rerender\":{\"action\":\"spawn_agent\",\"config\":{\"role\":\"page_renderer\",\"agent_type\":\"page-rerender\"},\"output_field\":\"rerender_agent\",\"next_step\":\"call_rerender\",\"description\":\"Spawn page-rerender\"},\"call_rerender\":{\"action\":\"call_agent\",\"config\":{\"agent_type\":\"page-rerender\",\"target_role\":\"page_renderer\",\"input_mapping\":{\"domain\":\"input_data.domain\",\"page_id\":\"input_data.page_id\",\"site_id\":\"input_data.site_id\"},\"timeout_seconds\":200},\"output_field\":\"rerender_result\",\"next_step\":\"complete\",\"description\":\"Assemble stored components and deploy\"},\"complete\":{\"action\":\"complete_workflow\",\"config\":{\"output_fields\":[\"rerender_result\"]},\"description\":\"Rerender complete\"}}}},\"input_data\":{\"domain\":\"${DOMAIN}\",\"page_id\":\"${PAGE_ID}\",\"site_id\":\"${SITE_ID}\"}}"

echo "========================================="
echo "Rerender tool-gauntlet (assemble-only + deploy)  ${DOMAIN}"
echo "  SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo "========================================="

kubectl -n kafka run "kcat-gauntlet-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '${BODY}' | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
  -H correlation_id=${CORRELATION_ID} \
  -H orchestration_id=${ORCHESTRATION_ID} \
  -H request_id=${REQUEST_ID} \
  -H message_id=${MESSAGE_ID} \
  -H message_type=request -H client_id=${CLIENT_ID} -H action=process \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=${TIMESTAMP} && echo PUBLISH_OK"

echo ""
echo "Find the run BY PAYLOAD, not by the printed id:"
echo "  SELECT current_step, status FROM orchestration_states"
echo "   WHERE collected_data->'input_data'->>'page_id' = '${PAGE_ID}'"
echo "   ORDER BY created_at DESC LIMIT 3;"
