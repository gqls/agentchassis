#!/usr/bin/env bash
# Direct section re-render for ONE page: every section re-rendered from its
# stored content_data through the CURRENT html_template, with NO LLM call.
# Use this to pick up a component template change or a content_data edit.
#
# WHY THIS EXISTS ALONGSIDE 049b_deploy_single_page.sh
# ----------------------------------------------------
# 049b builds its spec as {"reason": "..."} only. The live page-rerender agent
# wires the step like this:
#
#   "rerender_sections": {"action": "rerender_page_sections", "config": {
#       "reason":         "input_data.spec.reason",
#       "page_name":      "input_data.spec.page_name",     <-- REQUIRED
#       "target_site_id": "input_data.site_id"}}
#
# and rerender_page_sections declares Required: [target_site_id, page_name]
# (rerender_page_sections_action.go:80). So a dispatch carrying spec.reason but
# no spec.page_name fails immediately with:
#
#   step rerender_sections failed: ... input extraction failed:
#   missing required fields: [page_name]
#
# Measured 2026-07-26: both of this workstream's dispatches failed that way in
# under a second, and so did two from a different session on webdesign.co.uk —
# so it is the envelope, not the site. Following 049b's own documented recipe
# ("pass section_data_resolved to take the rerender_sections pre-pass") fails
# 100% of the time. This script passes page_name.
#
# GOTCHA carried over from 049b and still true: if ANY section has NULL
# content_data the whole page escalates to the content writer and the copy IS
# regenerated. Check before firing:
#   SELECT slot_name, content_data IS NULL FROM page_components WHERE page_id='...';
#
# Usage: rerender_page_sections_direct.sh <page_id> <site_id> <domain> <page_name> [reason]
#   reason defaults to section_data_resolved (the no-LLM re-render pre-pass).
set -euo pipefail

PAGE_ID="$1"; SITE_ID="$2"; DOMAIN="$3"; PAGE_NAME="$4"; REASON="${5:-section_data_resolved}"
CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "corr=$CORR page=$PAGE_NAME ($PAGE_ID) domain=$DOMAIN reason=$REASON"

# kcat -P -c 1 + heredoc: -c 1 reads exactly one message, which sidesteps the
# kubectl-run stdin race (the proven 049_TRIGGER pattern).
kubectl -n kafka run -i --rm "kcat-rerender-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_id=$MSG \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"page-rerender"},"input_data":{"page_id":"$PAGE_ID","site_id":"$SITE_ID","domain":"$DOMAIN","spec":{"reason":"$REASON","page_name":"$PAGE_NAME"}}}
JSON

echo "CORR=$CORR"
echo
echo "Find the run by PAYLOAD, not by this id, and use a window that starts BEFORE now:"
echo "  SELECT status, current_step, error FROM orchestration_states"
echo "   WHERE initial_request_data->'input_data'->>'page_id' = '$PAGE_ID'"
echo "     AND created_at > now() - interval '15 minutes' ORDER BY created_at DESC LIMIT 3;"
