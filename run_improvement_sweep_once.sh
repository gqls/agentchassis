#!/bin/bash
# run_improvement_sweep_once.sh — hand-fire ONE improvement-loop sweep at ONE site.
#
# vigilant_designer_offer_analysis programme: the owner's 2026-08-02 ruling is
# MANUAL PER-SITE TRIGGERS — the improvement-sweep scheduled task stays
# enabled=false (flipping it is G1, the owner's call, not yours). This script is
# the manual mode: it publishes exactly the spawn message the scheduler would,
# for one site, on the scheduler lane (migration 290).
#
# BLAST RADIUS — read before firing:
#   - The loop fingerprints the site (migration 291). If the fingerprint changed
#     or 14 days passed since last audit, the FULL AUDIT CHAIN runs: completeness
#     discovery + LLM auditors + (since 301) the render-audit write tail. That is
#     real LLM spend and a render sweep against the LIVE site.
#   - triage_findings then PROMOTES the site's detected work items on EVERY path
#     (single promoter, migration 286) — including items OTHER lanes filed at
#     this site. Promoted items get dispatched to their handlers, which CHANGE
#     LIVE PAGES. Before firing: read the site's detected queue and cancel
#     anything provably stale — a stale rerender row becomes a live page churn.
#       SELECT id, item_type, item_key, summary FROM site_work_items
#       WHERE site_id='<id>' AND status='detected';
#   - Three audits at an unchanged fingerprint file ONE capability_gap roadmap
#     row instead of reporting clean (bugs_open/171's fix — that is correct
#     behaviour, not a failure).
#   - No dispatch within ~300s of a chassis pod (re)start — silently dropped.
#
# Usage:
#   ./run_improvement_sweep_once.sh <site-domain>
#   ./run_improvement_sweep_once.sh relojistas.com
#
# Watch it (RUNBOOK has the full queries):
#   SELECT current_step, status FROM orchestration_states
#   WHERE correlation_id='<printed>' ORDER BY created_at DESC;
set -euo pipefail

DOMAIN="${1:?usage: $0 <site-domain>}"

SITE_ID=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -t -A \
  -c "SELECT id FROM sites WHERE domain='${DOMAIN}';")
if [ -z "$SITE_ID" ]; then
  echo "ERROR: no site row for domain '${DOMAIN}'" >&2; exit 1
fi

OPEN=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -t -A \
  -c "SELECT count(*) FROM site_work_items WHERE site_id='${SITE_ID}' AND status='detected';")
echo "NOTE: ${OPEN} detected item(s) at ${DOMAIN} will be promoted by this sweep."
echo "      If you have not reviewed them for staleness, Ctrl-C now (header says how)."

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

INPUT=$(jq -c -n --arg s "$SITE_ID" --arg d "$DOMAIN" '{site_id:$s, domain:$d}')

echo "========================================="
echo "improvement-loop sweep  correlation $CORRELATION_ID"
echo "  site:       $DOMAIN ($SITE_ID)"
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
    config:{agent_type:"improvement-loop"},
    input_data:$input}')

printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-sweep-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.scheduled.requests \
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
echo "SAVE: SWEEP_CORR=$CORRELATION_ID"
echo "Watch: SELECT current_step, status FROM orchestration_states WHERE correlation_id='$CORRELATION_ID';"
echo "Gate:  collected_data->'audit_state' on that row separates audited / skipped / not-converging (171)."
