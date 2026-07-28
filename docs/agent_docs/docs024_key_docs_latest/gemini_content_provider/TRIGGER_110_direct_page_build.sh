#!/usr/bin/env bash
# Fire page-build-handler DIRECTLY for one page, bypassing the build-dispatch queue.
#
# WHY THIS EXISTS. bugs_open/110's last open question needs one Gemini call through the
# orchestration path, and the only agent on Gemini is page-content-writer, which is only
# reached via page-build-handler. On 2026-07-28 the build queue was head-of-line blocked:
# every tick picked the lowest-UUID site with pending work (fundamentallyai.com), spawned
# build-dispatch-loop, and hung at AWAITING_RESPONSES with awaited=[] — bugs_open/029.
# Items behind it never got a tick.
#
# WHY NOT 049b_deploy_single_page.sh. That script exists for exactly this class of
# blockage, but it targets page-rerender, which is DESIGNED not to call an LLM: with no
# reason it assembles stored HTML, and with section_data_resolved it re-renders stored
# content_data through the current template — "with NO LLM call" per its own header. It
# escalates to the writer ONLY if a section has NULL content_data. So it cannot produce
# the Gemini row this needs. One level up is the right lever.
#
# The input_data below is copied from orchestration af2d066b, the run that built
# dartsonline/sale successfully on 2026-07-27 — not invented. Pass the real work_item_id
# so the handler completes the queued item rather than leaving it orphaned; the dispatch
# loop only selects status IN ('triaged','approved'), so a completed item cannot then be
# double-built if the queue unblocks later.
#
# Usage: ./TRIGGER_110_direct_page_build.sh <page_id> <site_id> <domain> <page_name> <page_role> <work_item_id>
set -euo pipefail
PAGE_ID="$1"; SITE_ID="$2"; DOMAIN="$3"; PAGE_NAME="$4"; PAGE_ROLE="$5"; WORK_ITEM_ID="$6"
REASON="gemini_110_candidate2_verification"

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "corr=$CORR orch=$ORCH page=$PAGE_NAME domain=$DOMAIN work_item=$WORK_ITEM_ID"

# kcat -P -c 1 + heredoc is the proven pattern (049b header): a plain
# `kubectl run -i --rm | kcat -P` silently publishes nothing and still exits 0.
# -c 1 reads exactly one message, sidestepping the kubectl-run stdin race.
kubectl -n kafka run -i --rm "kcat-110-$(date +%s)" \
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
{"action":"orchestrate","config":{"agent_type":"page-build-handler"},"input_data":{"domain":"$DOMAIN","source":"gemini-110-verification","site_id":"$SITE_ID","item_type":"needs_page","page_name":"$PAGE_NAME","work_item_id":"$WORK_ITEM_ID","spec":{"reason":"$REASON","page_name":"$PAGE_NAME","page_role":"$PAGE_ROLE"},"current_page":{"reason":"$REASON","page_name":"$PAGE_NAME","page_role":"$PAGE_ROLE"}}}
JSON
echo "CORR=$CORR"
echo
echo "Watch:  SELECT current_step, status FROM orchestration_states WHERE orchestration_id='$ORCH';"
echo "Verify: ./VERIFY_110_post_roll.sh"
