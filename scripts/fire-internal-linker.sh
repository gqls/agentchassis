#!/usr/bin/env bash
# Hand-fire ONE internal-linker run to supply bugs_open/313 + 298's artefact proof.
#
# WHY BY HAND: the 20 queued needs_internal_links items are status='unresolved',
# which is TERMINAL (work_items_common.go:40-46) and invisible to the promoter
# (reads 'detected' only), the selector and the atomic claim
# (workItemDispatchableStatuses = {triaged, approved}, work_items_common.go:172-175).
# They will never dispatch. Fresh items arrive only on a site-discovery-rotation
# tick (hourly, one site, 7-day per-site stamp) that finds new orphans — days, not
# hours. So "wait for natural traffic" is not a plan.
#
# Envelope shape copied from scripts/backfill-meta-descriptions.sh (same class of
# job): reads the LIVE workflow out of agent_definitions so what runs is what is
# seeded, emits ONE line of JSON (kcat -P publishes one message per line — a
# pretty-printed envelope arrives as N invalid fragments), publishes to
# system.agent.generic.requests.
#
# INPUT CONTRACT (the trap): ensure_site_record reads input_data.domain
# (site_db_actions.go:617 extractDomainFromInput); load_target_page binds
# site_record.site_id + input_data.spec.page_name. Without spec.page_name the run
# exits at complete_not_found and proves nothing.
#
# ⚠ EXIT 0 PROVES NOTHING. Verify in orchestration_states by correlation.
set -euo pipefail

DOMAIN="${1:?usage: $0 <domain> <page_name>}"
PAGE="${2:?usage: $0 <domain> <page_name>}"
NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

SITE_ID=$($PSQL -tAc "SELECT id FROM sites WHERE domain='$DOMAIN';" | tr -d '[:space:]')
[ -n "$SITE_ID" ] || { echo "no site row for $DOMAIN" >&2; exit 2; }

# The "before" arm, re-read at fire time so the proof cannot rest on a stale zero.
BEFORE=$($PSQL -tAc "SELECT count(*) FROM llm_call_log WHERE agent_type='internal-linker';" | tr -d '[:space:]')
echo "$DOMAIN site_id=$SITE_ID page=$PAGE  llm_call_log_before=$BEFORE"

WF=$(mktemp); MSG_F=$(mktemp); trap 'rm -f "$WF" "$MSG_F"' EXIT
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='internal-linker' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
[ -s "$WF" ] || { echo "empty workflow read — refusing to dispatch" >&2; exit 3; }

CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

python3 - "$WF" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE_ID" "$DOMAIN" "$PAGE" > "$MSG_F" <<'PY'
import json,sys
wf=json.load(open(sys.argv[1],encoding='utf-8'))
corr,orch,req,msg,ts,site,dom,page=sys.argv[2:10]
print(json.dumps({"headers":{"correlation_id":corr,"orchestration_id":orch,"request_id":req,
  "message_id":msg,"message_type":"request","client_id":"cli-linker-canary","action":"process",
  "sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":ts},
  "config":{"workflow":wf},
  "input_data":{"site_id":site,"domain":dom,"spec":{"page_name":page}}},separators=(',',':')))
PY

# one line, or kcat will publish fragments
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
  --header "client_id=cli-linker-canary" \
  --header "action=process" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TS" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no internal linking will run (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi

echo "CORRELATION_ID=$CORR"

echo "dispatched (receipt asserted; now check the durable record):"
echo "  CORR=$CORR"
