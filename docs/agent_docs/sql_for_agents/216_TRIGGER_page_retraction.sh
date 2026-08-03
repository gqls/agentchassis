#!/bin/bash
# 216_TRIGGER_page_retraction.sh — dispatch the `page-retraction` agent
# (bugs_open/098, concept register DGH-006, RFC 011 option B).
#
# WHAT IT DOES: removes the deployed artefact of pages the platform no longer
# wants served, retires their nav rows, and reports anything left stranded.
#
# THE GUARDS ARE IN THE ACTION, NOT HERE, and cannot be switched off from this
# script: paths are derived only from `pages.url` via the shared
# PageFilePathFromURL; a page whose url names no file of its own is declined; a
# path an ACTIVE page also derives is refused; a page still linked from live
# content (rendered markup OR stored content_data) is refused and the referrers
# are named. Every refusal is reported rather than swallowed.
#
# ALWAYS PASS PAGE_IDS for a real retraction. Omitting it makes the action
# consider EVERY non-active page on the site that still carries a deploy stamp,
# which is the right default for a sweep and the wrong one for a first run.
#
# ACCEPTANCE IS TWO-PART, and the second half is the one that tests anything:
#   1. the url 404s immediately;
#   2. it STILL 404s after the next news refresh (~08:0x / ~20:0x).
# Part 1 passed even before the resurrection fix, because an archived page was
# being re-rendered and re-published twice a day (bugs_open/098's correction).
#
# Usage:
#   SITE_ID=<uuid> PAGE_IDS='["<uuid>"]' ./216_TRIGGER_page_retraction.sh
set -euo pipefail

SITE_ID="${SITE_ID:?set SITE_ID}"
PAGE_IDS="${PAGE_IDS:-}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

if [ -n "$PAGE_IDS" ]; then
  INPUT=$(jq -c -n --arg s "$SITE_ID" --argjson p "$PAGE_IDS" '{site_id:$s, page_ids:$p}')
else
  echo "WARNING: no PAGE_IDS — every non-active page on the site with a deploy stamp is in scope." >&2
  INPUT=$(jq -c -n --arg s "$SITE_ID" '{site_id:$s}')
fi

echo "========================================="
echo "page-retraction  correlation $CORRELATION_ID"
echo "  site_id:    $SITE_ID"
echo "  input_data: $INPUT"
echo "========================================="

# ONE LINE: kcat -P splits stdin on newlines into separate messages, so a
# pretty-printed payload becomes N broken messages, each failing silently.
PAYLOAD=$(jq -c -n \
  --arg c "$CORRELATION_ID" --arg o "$ORCHESTRATION_ID" --arg r "$REQUEST_ID" \
  --arg m "$MESSAGE_ID" --arg t "$TIMESTAMP" --arg cl "$CLIENT_ID" \
  --argjson input "$INPUT" \
  '{headers:{correlation_id:$c,orchestration_id:$o,request_id:$r,message_id:$m,
             message_type:"request",client_id:$cl,action:"orchestrate",
             sender:{agent_id:"cli-user",agent_type:"cli",pod_name:"cli"},timestamp:$t},
    config:{agent_type:"page-retraction"},
    input_data:$input}')

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-retract-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" \
  -H "client_id=$CLIENT_ID" \
  -H "action=orchestrate" \
  -H "sender_agent_type=cli" \
  -H "sender_agent_id=cli-user" \
  -H "timestamp=$TIMESTAMP"

echo
echo "Watch it:"
echo "  SELECT current_step, status, jsonb_pretty(collected_data->'retraction')"
echo "    FROM orchestration_states WHERE orchestration_id='$ORCHESTRATION_ID';"
