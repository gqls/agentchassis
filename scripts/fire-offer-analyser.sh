#!/usr/bin/env bash
# fire-offer-analyser.sh — hand-fire ONE offer-analyser (B4) run at ONE site.
#
# WHY THIS EXISTS AND WHY NOT run_improvement_sweep_once.sh. That script is the
# lane's normal manual mode, and it fires the WHOLE improvement loop: the full
# audit chain (completeness discovery + LLM auditors + the render-audit write
# tail), and — the part that matters — `triage_findings` PROMOTES every
# status='detected' work item on the site and dispatches it to its handler,
# which CHANGES LIVE PAGES. [MEASURED 2026-08-22] that is 111 items on
# webdesign.co.uk and 37 on leopardessconsulting.co.uk, including items other
# lanes filed. To prove one gate, that is the wrong instrument by two orders of
# magnitude.
#
# This publishes exactly one dispatch for `offer-analyser` alone: no
# improvement-loop, no triage, no promotion, no handler dispatch. One LLM call.
#
# INPUT CONTRACT: `ensure_site_record` resolves the site from input_data.domain
# / input_data.site_id; `load_premise` and `load_offer_surface` then bind
# site_record.site_id. Nothing else is needed — unlike the linker, there is no
# spec.page_name.
#
# ⚠ EXIT 0 PROVES NOTHING — kcat -P can publish nothing and exit 0. Verify at the
# durable record; this script prints the before-arm and the queries to run.
#
# ⚠ NO DISPATCH WITHIN ~300s OF A CHASSIS POD (RE)START — silently dropped.
#
# Usage:  ./scripts/fire-offer-analyser.sh <site-domain>
set -euo pipefail

DOMAIN="${1:?usage: $0 <site-domain>}"
NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

SITE_ID=$($PSQL -tAc "SELECT id FROM sites WHERE domain='$DOMAIN';" | tr -d '[:space:]')
[ -n "$SITE_ID" ] || { echo "no site row for $DOMAIN" >&2; exit 2; }

# Before-arms, re-read at fire time so the proof cannot rest on a stale number.
# ⚠ KEY ON step_name, NOT agent_type. agent_type carries the DISPATCH context:
# a hand-fired run lands under 'generic', so filtering on 'offer-analyser' reads
# UNCHANGED after a successful run and fails toward "it did not run" (measured
# 2026-08-22; LANDMINES, "llm_call_log.agent_type is NOT which agent's workflow
# made this call"). step_name is written by the step itself.
BEFORE_LLM=$($PSQL -tAc "SELECT count(*) FROM llm_call_log WHERE step_name='run_offer_analysis';" | tr -d '[:space:]')
BEFORE_ORD=$($PSQL -tAc "SELECT COALESCE(jsonb_array_length(data->'lead_with'),-1) FROM site_specs
  WHERE site_id='$SITE_ID' AND aspect='offer_ordering' AND is_current;" | tr -d '[:space:]')
BEFORE_ROW=$($PSQL -tAc "SELECT COALESCE(id::text,'none') FROM site_specs
  WHERE site_id='$SITE_ID' AND aspect='offer_ordering' AND is_current;" | tr -d '[:space:]')

echo "$DOMAIN  site_id=$SITE_ID"
echo "  before: llm_call_log(offer-analyser)=$BEFORE_LLM  lead_with_len=$BEFORE_ORD  spec_row=$BEFORE_ROW"

# Read the LIVE workflow, so what runs is what is seeded — never a repo copy.
WF=$(mktemp); MSG_F=$(mktemp); trap 'rm -f "$WF" "$MSG_F"' EXIT
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='offer-analyser' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
[ -s "$WF" ] || { echo "empty workflow read — refusing to dispatch" >&2; exit 3; }

# Fail loudly if the gate is not in the workflow we are about to run: this
# script's whole purpose is proving that step, and a silent pre-537 workflow
# would produce a clean-looking run that proves the opposite of what is claimed.
grep -q 'verify_ordering_cardinals' "$WF" || {
  echo "the live workflow has NO verify_ordering_cardinals step — refusing" >&2; exit 5; }

CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

python3 - "$WF" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE_ID" "$DOMAIN" > "$MSG_F" <<'PY'
import json,sys
wf=json.load(open(sys.argv[1],encoding='utf-8'))
corr,orch,req,msg,ts,site,dom=sys.argv[2:9]
print(json.dumps({"headers":{"correlation_id":corr,"orchestration_id":orch,"request_id":req,
  "message_id":msg,"message_type":"request","client_id":"cli-offer-analyser","action":"process",
  "sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":ts},
  "config":{"workflow":wf},
  "input_data":{"site_id":site,"domain":dom}},separators=(',',':')))
PY

# one line, or kcat publishes fragments
[ "$(wc -l < "$MSG_F")" -eq 1 ] || { echo "envelope is not one line — refusing" >&2; exit 4; }

REPO_ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT" ] || [ ! -f "$REPO_ROOT/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found — refusing to publish unverified (bugs_open/327)." >&2
  exit 1
fi
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

PUBLISH_RC=0
kafka_publish_checked \
  --topic system.agent.generic.requests \
  --correlation "$CORR" \
  --payload "$(cat "$MSG_F")" \
  --header "orchestration_id=$ORCH" \
  --header "request_id=$REQ" \
  --header "message_id=$MSG" \
  --header "message_type=request" \
  --header "client_id=cli-offer-analyser" \
  --header "action=process" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TS" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no offer analysis will run (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi

echo "CORRELATION_ID=$CORR"

cat <<EOF

dispatched (exit 0 proves NOTHING — check the durable record):
  CORR=$CORR

  -- did it run?  (llm_call_log outlives orchestration_states' ~24h reaping)
  -- keyed on step_name: agent_type reads 'generic' for a hand-fired run
  SELECT created_at, agent_type, step_name, success FROM llm_call_log
   WHERE correlation_id='$CORR';                          -- exact, this run only
  SELECT count(*) FROM llm_call_log WHERE step_name='run_offer_analysis';  -- was $BEFORE_LLM

  -- where is it?
  SELECT current_step, status FROM orchestration_states WHERE correlation_id='$CORR';

  -- what did the gate do?  (empty array = ran and dropped nothing)
  SELECT jsonb_array_length(data->'lead_with') AS kept,
         jsonb_array_length(COALESCE(data->'dropped_unsourced','[]'::jsonb)) AS dropped,
         data->'dropped_unsourced'
    FROM site_specs WHERE site_id='$SITE_ID' AND aspect='offer_ordering' AND is_current;
  -- was $BEFORE_ORD points, spec row $BEFORE_ROW (a NEW row id proves a write happened)
EOF
