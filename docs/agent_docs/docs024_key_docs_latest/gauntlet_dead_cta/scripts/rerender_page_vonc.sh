#!/usr/bin/env bash
# rerender_page_vonc.sh <page_id> — assemble-only rerender + deploy of ANY vonc page.
#
# Generalised from rerender_arena_vonc.sh (which hardcodes the tool-arena page id).
# Same envelope, same hardened publish; the ONLY difference is that the page comes
# in as an argument, so this does not need cloning per page.
#
# page-rerender runs:
#   check_rerender_mode -> rerender_single_page (assemble stored page_components)
#   -> check_skipped -> deploy_page (git_commit to 'sites') -> update_status
# NO content rebuild: this re-assembles page_components.rendered_html and deploys.
#
# WHY THE PUBLISH LOOKS LIKE THIS: `kubectl run -i --rm ... -- kcat -P <<JSON` loses
# the payload if the container starts before stdin attaches — kcat sees EOF, sends
# NOTHING and exits 0, so the wrapper prints a correlation id as if it had worked
# (measured 2026-07-26: 1 publish in 5 landed). Payload goes in the container
# COMMAND, --command beats the kcat ENTRYPOINT, and `&& echo PUBLISH_OK` makes a
# silent drop visible.
#
# PRE-REQS:
#   1. page_components.rendered_html already holds the markup you want served —
#      rerender assembles from THAT column, not content_components.html_template
#      (rerender_single_page_action.go:163,232,511).
#   2. Not within ~300s of a chassis pod restart, or the spawn is silently dropped:
#      kubectl -n ai-persona-system get pods -l app=agent-chassis
#   3. rerender_single_page has NO rebuild_policy check, so this is safe on an
#      `owned` page — unlike a generic rerender, which is hard-refused (bugs_closed/024).
set -u

PAGE_ID="${1:-}"
if [ -z "$PAGE_ID" ]; then
  echo "usage: $0 <page_id>" >&2
  echo "  e.g. a28abcd7-186b-4a33-9b89-5d7bfd727012   # /about.html" >&2
  echo "       d2c8a925-1dca-44b6-a866-4561905a87a8   # /tools/arena/index.html" >&2
  exit 2
fi

SITE_ID="9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"
DOMAIN="vonc.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

BODY="{\"headers\":{\"correlation_id\":\"${CORRELATION_ID}\",\"orchestration_id\":\"${ORCHESTRATION_ID}\",\"request_id\":\"${REQUEST_ID}\",\"message_id\":\"${MESSAGE_ID}\",\"message_type\":\"request\",\"client_id\":\"${CLIENT_ID}\",\"action\":\"process\",\"sender\":{\"agent_id\":\"cli-user\",\"agent_type\":\"cli\",\"pod_name\":\"cli\"},\"timestamp\":\"${TIMESTAMP}\"},\"config\":{\"workflow\":{\"start_step\":\"spawn_rerender\",\"processing_mode\":\"orchestrator\",\"timeout_seconds\":240,\"steps\":{\"spawn_rerender\":{\"action\":\"spawn_agent\",\"config\":{\"role\":\"page_renderer\",\"agent_type\":\"page-rerender\"},\"output_field\":\"rerender_agent\",\"next_step\":\"call_rerender\",\"description\":\"Spawn page-rerender\"},\"call_rerender\":{\"action\":\"call_agent\",\"config\":{\"agent_type\":\"page-rerender\",\"target_role\":\"page_renderer\",\"input_mapping\":{\"domain\":\"input_data.domain\",\"page_id\":\"input_data.page_id\",\"site_id\":\"input_data.site_id\"},\"timeout_seconds\":200},\"output_field\":\"rerender_result\",\"next_step\":\"complete\",\"description\":\"Assemble stored components and deploy\"},\"complete\":{\"action\":\"complete_workflow\",\"config\":{\"output_fields\":[\"rerender_result\"]},\"description\":\"Rerender complete\"}}}},\"input_data\":{\"domain\":\"${DOMAIN}\",\"page_id\":\"${PAGE_ID}\",\"site_id\":\"${SITE_ID}\"}}"

echo "========================================="
echo "Rerender (assemble-only + deploy)  ${DOMAIN}"
echo "  PAGE_ID=${PAGE_ID}"
echo "  SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo "========================================="

kubectl -n kafka run "kcat-rerender-$(date +%s)-$RANDOM" --rm --restart=Never \
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
