#!/usr/bin/env bash
# Hand-fire ONE `content-gap-planner` run against a named domain.
#
# ⚠⚠ THIS RUN WRITES TO A LIVE SITE. The workflow ends
#   plan_gaps (execute_llm_prompt) -> apply_plan (action: apply_gap_plan) -> complete
# and `apply_gap_plan` APPLIES the plan — it is not a proposal step and there is no
# human-review park in this workflow, unlike copy-editor's `copy_edit_proposed`. Do not
# fire it to "see what it says". Fire it when you intend the plan to land.
#
# WHY THIS SCRIPT EXISTS — the trap it removes (bugs_closed/327b family, 016b §9 case 3):
# three canary fires on 2026-09-02/03 were built by hand-copying the ENVELOPE out of
# scripts/fire-copy-editor.sh, and all three were REJECTED AT INTAKE within seconds:
#
#   agent_error_log: INCOMING_MESSAGE_REJECTED
#     "incoming message rejected: missing required header(s): client_id, orchestration_id"
#
# `client_id` and `orchestration_id` are required as **Kafka headers**. The envelope's own
# `headers` OBJECT is payload and intake does not read it — so an envelope that looks
# byte-perfect next to a working one still dies, and it dies where nothing the firer watches
# will show it. `orchestration_states` stays empty (retains ~2 days); the refusal is durable in
# `agent_error_log` (~30 days). The three fires read as an agent-family-specific dispatch bug
# for a day. They were script-specific.
#
# ⚠ EXIT 0 PROVES NOTHING, TWICE OVER. kcat -P publishes and exits 0 having sent nothing
# (327b), AND a genuinely published message can still be refused at intake. This script uses
# scripts/kafka-publish-lib.sh (OPP-009) for the first and CHECKS agent_error_log for the
# second. The orchestration row is the only proof of dispatch.
set -euo pipefail

DOMAIN="${1:?usage: $0 <domain>}"
NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
CLIENT_ID=cli-gapplanner-canary

SITE_ID=$($PSQL -tAc "SELECT id FROM sites WHERE domain='$DOMAIN';" | tr -d '[:space:]')
[ -n "${SITE_ID:-}" ] || { echo "no site row for '$DOMAIN'" >&2; exit 2; }

# Dispatch readiness is the DEPLOYMENT, not pod churn: most pods running this image are
# dynamic-agent pods spawned per orchestration, so "no pod started in 300s" can never be true.
kubectl -n $NS rollout status deploy/agent-chassis --timeout=20s >/dev/null || {
  echo "agent-chassis is mid-rollout — a spawn within ~300s of a restart is dropped" >&2; exit 3; }

HEALTHY=$($PSQL -tAc "SELECT healthy FROM ai_endpoint_health WHERE name='claude';" | tr -d '[:space:]')
[ "$HEALTHY" = "t" ] || { echo "claude endpoint is not healthy — refusing to spend a run" >&2; exit 4; }

# Another session's gap work on this site would race on apply.
OPEN=$($PSQL -tAc "SELECT count(*) FROM site_work_items
  WHERE site_id='$SITE_ID' AND item_type LIKE '%gap%'
    AND status NOT IN ('complete','cancelled','rejected');" | tr -d '[:space:]')
[ "$OPEN" = "0" ] || { echo "$OPEN open gap work item(s) on $DOMAIN — read them first" >&2; exit 5; }

WF=$(mktemp); MSG_F=$(mktemp); trap 'rm -f "$WF" "$MSG_F"' EXIT
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='content-gap-planner' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
[ -s "$WF" ] || { echo "empty workflow read — refusing to dispatch" >&2; exit 6; }

CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

python3 - "$WF" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE_ID" "$DOMAIN" "$CLIENT_ID" > "$MSG_F" <<'PY'
import json,sys
wf=json.load(open(sys.argv[1],encoding='utf-8'))
corr,orch,req,msg,ts,site,dom,client=sys.argv[2:10]
print(json.dumps({"headers":{"correlation_id":corr,"orchestration_id":orch,"request_id":req,
  "message_id":msg,"message_type":"request","client_id":client,"action":"process",
  "sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":ts},
  "config":{"workflow":wf},
  "input_data":{"site_id":site,"domain":dom}},separators=(',',':')))
PY

# one line, or kcat publishes N invalid fragments
[ "$(wc -l < "$MSG_F")" -eq 1 ] || { echo "envelope is not one line — refusing" >&2; exit 7; }

echo "domain=$DOMAIN site_id=$SITE_ID  ⚠ apply_gap_plan WILL write to this site"
REPO_ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT" ] || [ ! -f "$REPO_ROOT/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found — refusing to publish unverified (bugs_closed/327b)." >&2
  exit 1
fi
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

# THE HEADERS ARE THE POINT. Every one of these is sent as a Kafka header, NOT as payload;
# omitting client_id or orchestration_id gets the message refused at intake with no
# orchestration row and no signal to the firer. Do not "simplify" this list.
PUBLISH_RC=0
kafka_publish_checked \
  --topic system.agent.generic.requests \
  --correlation "$CORR" \
  --payload "$(cat "$MSG_F")" \
  --header "orchestration_id=$ORCH" \
  --header "request_id=$REQ" \
  --header "message_id=$MSG" \
  --header "message_type=request" \
  --header "client_id=$CLIENT_ID" \
  --header "action=process" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TS" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no gap plan will run (bugs_closed/327b)." >&2
  exit "$PUBLISH_RC"
fi

echo "CORRELATION_ID=$CORR"
echo "ORCHESTRATION_ID=$ORCH"

# A receipted publish can still be REFUSED at intake. Check the durable record before the
# caller walks away believing this landed. ~10s is generous: the three 2026-09-03 refusals
# were logged 6-70s after publish.
sleep 12
REJ=$($PSQL -tAc "SELECT count(*) FROM agent_error_log
  WHERE error_code='INCOMING_MESSAGE_REJECTED' AND occurred_at > now() - interval '2 minutes';" | tr -d '[:space:]')
if [ "${REJ:-0}" != "0" ]; then
  echo "⚠ $REJ INCOMING_MESSAGE_REJECTED row(s) in the last 2 minutes — very likely THIS message." >&2
  $PSQL -c "SELECT occurred_at, error_message FROM agent_error_log
     WHERE error_code='INCOMING_MESSAGE_REJECTED' AND occurred_at > now() - interval '2 minutes'
     ORDER BY occurred_at DESC LIMIT 3;" >&2
  echo "Refused at intake: NOT dispatched. Fix the headers; do not resend unchanged (016b §9 case 3)." >&2
  exit 12
fi

cat <<MSG

published and not refused within 12s. The durable record is still the only proof:

  SELECT current_step, status FROM orchestration_states WHERE correlation_id='$CORR';

A missing row is usually queue latency (measured up to ~29 min under load), not a drop.
Do NOT re-fire on an absent row — check agent_error_log first:

  SELECT occurred_at, error_code, error_message FROM agent_error_log
   WHERE occurred_at > now() - interval '1 hour' ORDER BY occurred_at DESC LIMIT 10;

The canary this exists for — did {{.build_standard}} render? (opt-ins 677/678/679):

  SELECT agent_type, created_at,
         position('BUILD STANDARD (applies to every site, regardless of inputs). Aim'
                  IN prompt_rendered)>0 AS has_standard,
         position('{{.build_standard}}' IN prompt_rendered)>0 AS unrendered
    FROM llm_call_log WHERE correlation_id='$CORR';

⚠ Do NOT use "stands comparison with the strongest sites" as that needle — it also matches
domain-research-classifier's own hard-coded copy of the block and proves nothing about injection.
MSG
