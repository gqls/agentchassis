#!/usr/bin/env bash
# Hand-fire ONE `copy-editor` (stage 2) run against a named page.
#
# WHY BY HAND: nothing dispatches copy-editor, BY CHOICE (owner decision D2, 2026-08-12) —
# no item_type routes to it and it is absent from the improvement sweep. Every run to date
# has been a hand-fired canary. Its output PARKS at `copy_edit_proposed` /
# `needs_human_review`; no step in it can write to a page and migration 447 RAISEs if one
# is added, so a run cannot change a live site on its own.
#
# WHY THIS SCRIPT EXISTS: runs 1-3 were fired ad hoc and their orchestration rows have
# since been deleted (orchestration_states is NOT an archive), so the envelope was
# unrecoverable and the next session would have re-derived it. Envelope shape copied from
# scripts/fire-internal-linker.sh (same class of job).
#
# INPUT CONTRACT (the trap, and it differs from the linker's):
#   * ensure_site_record reads input_data.domain (site_db_actions.go extractDomainFromInput)
#   * load_page_target binds input_data.page_id ONLY — a `query_database` step whose
#     params are ["input_data.page_id"]. It does NOT take a page NAME. Passing spec.page_name
#     the way the linker does leaves page_target empty and the run judges nothing.
#
# ⚠ EXIT 0 PROVES NOTHING. kcat -P publishes and exits 0 having sent nothing. The
# orchestration row is the only proof of dispatch. Verify by correlation, and budget for
# queue latency (measured up to ~29 min under normal fleet load) before suspecting a drop.
set -euo pipefail

DOMAIN="${1:?usage: $0 <domain> <page_name>}"
PAGE="${2:?usage: $0 <domain> <page_name>}"
NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

read -r SITE_ID PAGE_ID <<<"$($PSQL -tAF' ' -c \
  "SELECT s.id, p.id FROM pages p JOIN sites s ON s.id=p.site_id
    WHERE s.domain='$DOMAIN' AND p.name='$PAGE';" | tr -d '\r')"
[ -n "${PAGE_ID:-}" ] || { echo "no page '$PAGE' on '$DOMAIN'" >&2; exit 2; }

# Dispatch readiness is the DEPLOYMENT, not pod churn: 70 of ~75 pods running this image
# are dynamic-agent pods spawned per orchestration, so "no pod started in 300s" is a
# condition that can never be true. (LANDMINES, 2026-08-20.)
kubectl -n $NS rollout status deploy/agent-chassis --timeout=20s >/dev/null || {
  echo "agent-chassis is mid-rollout — a spawn within ~300s of a restart is dropped" >&2; exit 3; }

HEALTHY=$($PSQL -tAc "SELECT healthy FROM ai_endpoint_health WHERE name='claude';" | tr -d '[:space:]')
[ "$HEALTHY" = "t" ] || { echo "claude endpoint is not healthy — refusing to spend a run" >&2; exit 4; }

# Refuse if another session already has copy-editor work parked on this page: a second
# proposal against the same components would race on review, and the first one's
# page_component_ids may already dangle.
# ⚠ The page lives in the page_id COLUMN, not in spec. checkpoint_for_review_action writes
# spec keys {domain, checkpoint, on_approve, review_data, source_agent, correlation_id} and
# NO page_id, so a guard reading spec->>'page_id' returns 0 on every real row — armed and
# inert, which is the shape this lane keeps catching in its own tools. Verified against the
# two live proposals: both carry page_id in the column and null in spec.
PARKED=$($PSQL -tAc "SELECT count(*) FROM site_work_items
  WHERE item_type='copy_edit_proposed' AND status NOT IN ('complete','cancelled','rejected')
    AND page_id='$PAGE_ID';" | tr -d '[:space:]')
[ "$PARKED" = "0" ] || { echo "$PARKED copy_edit_proposed item(s) already parked on this page — read them first" >&2; exit 5; }

WF=$(mktemp); MSG_F=$(mktemp); trap 'rm -f "$WF" "$MSG_F"' EXIT
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='copy-editor' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
[ -s "$WF" ] || { echo "empty workflow read — refusing to dispatch" >&2; exit 6; }

CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

python3 - "$WF" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE_ID" "$DOMAIN" "$PAGE_ID" > "$MSG_F" <<'PY'
import json,sys
wf=json.load(open(sys.argv[1],encoding='utf-8'))
corr,orch,req,msg,ts,site,dom,page_id=sys.argv[2:10]
print(json.dumps({"headers":{"correlation_id":corr,"orchestration_id":orch,"request_id":req,
  "message_id":msg,"message_type":"request","client_id":"cli-copyedit-canary","action":"process",
  "sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":ts},
  "config":{"workflow":wf},
  "input_data":{"site_id":site,"domain":dom,"page_id":page_id}},separators=(',',':')))
PY

# one line, or kcat publishes N invalid fragments
[ "$(wc -l < "$MSG_F")" -eq 1 ] || { echo "envelope is not one line — refusing" >&2; exit 7; }

echo "domain=$DOMAIN page=$PAGE page_id=$PAGE_ID"
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
  --header "client_id=cli-copyedit-canary" \
  --header "action=process" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TS" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no copy edit will run (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi

echo "CORRELATION_ID=$CORR"

cat <<MSG

dispatched — and exit 0 proves nothing. The durable record is the only proof:

  SELECT current_step, status FROM orchestration_states WHERE correlation_id='$CORR';

A missing row is almost always queue latency, not a drop (measured up to ~29 min under
normal load). Do NOT re-fire on an absent row — that costs a duplicate proposal.

When it completes, the proposal parks here:

  SELECT id, status, spec->>'page_id' FROM site_work_items
   WHERE item_type='copy_edit_proposed' ORDER BY created_at DESC LIMIT 1;

Then grade it BEFORE acting on it:
  docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/gate_stage2_edit.py --item <id>
MSG
