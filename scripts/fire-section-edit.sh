#!/usr/bin/env bash
# Dispatch ONE approved `section_edit` to section-editor. Sequential by design: several
# edits on one page race on the render and the deploy.
#
# This is the apply half of stage 2 (copy-editor). The proposal parks at
# `copy_edit_proposed`/`needs_human_review`, a human approves, you file `section_edit`
# items, then this ships them one at a time. Proven end to end 2026-08-23 on
# ai-agent-orchestration.com/index: 3 edits, DB -> render -> deploy -> served page.
#
# TWO TRAPS THIS SCRIPT EXISTS TO AVOID, both of which cost a failed attempt on 2026-08-21:
#
#  1. `client_id` BECOMES A SCHEMA NAME. `spawn_actions.go:2315` builds
#     `INSERT INTO client_%s.agent_instances` with the client_id interpolated UNQUOTED, and
#     section-editor's SECOND step is spawn_deployer. A hyphenated client_id
#     (`cli-copyedit-apply`) therefore dies as `syntax error at or near "-"` (SQLSTATE
#     42601) before any edit is attempted — which reads like a platform outage and is your
#     own payload. Use a real client id; the fleet uses `demo_client` / `system`.
#
#  2. THE COMPONENT ID IS RESOLVED BY SLOT AT DISPATCH TIME, never read from the item.
#     A rerender REPLACES the page_components row, and pages are re-rendered on a schedule
#     (ai-agent-orchestration.com/index: daily, ~14:43). An id captured when the item was
#     filed is dead within a day — measured twice on one proposal: filed 2026-08-21 with
#     the then-live ids, all three dead by the 2026-08-22 14:43 rerender.
#     (page_id, slot_name) is the identity; a stored id is a hint.
#
# ⚠ AND 'complete' IS NOT PROOF. check_edit_skipped routes a lock- or decision-gated
# REFUSAL to 'complete' too. Verify that page_components.content_data actually CHANGED.
# ⚠ Watch YOUR correlation, never "the most recent section-editor row" — with several
# edits in flight that returns the previous one and reports it as yours.
ITEM="$1"; NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
# ⚠ The component id is resolved BY SLOT at dispatch time, NOT read from the item.
# This page is re-rendered daily (~14:43) and a rerender REPLACES the page_components row,
# so an id captured when the item was filed is dead within a day — measured twice on this
# proposal: filed 08-21 with the then-live ids, all three dead by 08-22 14:43.
# (page_id, slot_name) is the identity; the stored id is a hint.
read -r SITE DOMAIN PAGE_ID PAGE_NAME SLOT COMP <<<"$($PSQL -tAF' ' -c \
 "SELECT wi.site_id, s.domain, wi.spec->>'page_id', wi.spec->>'page_name',
         wi.spec->>'slot_name', pc.id
    FROM site_work_items wi
    JOIN sites s ON s.id=wi.site_id
    JOIN page_components pc ON pc.page_id = (wi.spec->>'page_id')::uuid
                           AND pc.slot_name = wi.spec->>'slot_name'
   WHERE wi.id='$ITEM';" | tr -d '\r')"
[ -n "${COMP:-}" ] || { echo "no live component in slot for item $ITEM — the slot is gone, not merely re-rendered" >&2; exit 6; }
WF=$(mktemp); MSG_F=$(mktemp); FU=$(mktemp); trap 'rm -f "$WF" "$MSG_F" "$FU"' EXIT
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='section-editor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
[ -s "$WF" ] || { echo "empty workflow — refusing" >&2; exit 3; }
$PSQL -tAc "SELECT spec->'field_updates' FROM site_work_items WHERE id='$ITEM';" > "$FU"
[ -s "$FU" ] || { echo "no field_updates — refusing" >&2; exit 4; }
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
python3 - "$WF" "$FU" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE" "$DOMAIN" "$PAGE_ID" "$PAGE_NAME" "$SLOT" "$COMP" "$ITEM" > "$MSG_F" <<'PY'
import json,sys
wf=json.load(open(sys.argv[1],encoding='utf-8')); fu=json.load(open(sys.argv[2],encoding='utf-8'))
corr,orch,req,msg,ts,site,dom,pid,pname,slot,comp,item=sys.argv[3:15]
print(json.dumps({"headers":{"correlation_id":corr,"orchestration_id":orch,"request_id":req,
 "message_id":msg,"message_type":"request","client_id":"demo_client","action":"process",
 "sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":ts},
 "config":{"workflow":wf},
 "input_data":{"site_id":site,"domain":dom,"page_id":pid,"page_name":pname,"slot_name":slot,
  "page_component_id":comp,"component_id":comp,"edit_type":"content_edit",
  "field_updates":fu,"work_item_id":item}},separators=(',',':')))
PY
[ "$(wc -l < "$MSG_F")" -eq 1 ] || { echo "envelope not one line — refusing" >&2; exit 5; }
echo "$SLOT  CORR=$CORR"
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
  --header "client_id=demo_client" \
  --header "action=process" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TS" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no section edit will run (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi
echo "$CORR" > "/tmp/last_corr_$ITEM" 2>/dev/null || true
