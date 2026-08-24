#!/usr/bin/env bash
# ONE-OFF regeneration of meta descriptions for a site (bugs_open/320 §15).
# Owner-authorised 2026-08-21 ("redo the descriptions through the framework").
#
# ⚠ THIS REPLACES LIVE PUBLIC COPY. It sends overwrite_existing=true INLINE.
# The SEEDED agent is deliberately NOT armed, so the hourly scheduled task stays
# fill-blanks-only. Check that has not drifted:
#   SELECT default_config#>'{workflow,steps,backfill_loop,config,sub_workflow,steps,save_description,config}' ? 'overwrite_existing'
#   FROM agent_definitions WHERE type='meta-description-backfiller' AND is_active;   -- must be f
#
# Reversible: every pre-regeneration description is in meta_description_pre_regen_20260821.
# Usage: ./scripts/regen-meta-descriptions.sh <domain> <path-to-wf_regen.json>
set -euo pipefail
DOMAIN="${1:?usage: $0 <domain> <wf.json>}"; WF="${2:?}"
NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"
SITE=$($PSQL -tAc "SELECT id FROM sites WHERE domain='$DOMAIN';" | tr -d '[:space:]')
[ -n "$SITE" ] || { echo "no site $DOMAIN" >&2; exit 2; }
N=$($PSQL -tAc "SELECT count(*) FROM pages WHERE site_id='$SITE' AND status='active' AND COALESCE(meta_description,'')<>'';" | tr -d '[:space:]')
echo "$DOMAIN  with_description=$N"
[ "$N" -gt 0 ] || { echo "nothing to redo"; exit 0; }
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ"); F=$(mktemp); trap 'rm -f "$F"' EXIT
python3 - "$WF" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE" "$DOMAIN" > "$F" <<'PY'
import json,sys
wf=json.load(open(sys.argv[1],encoding='utf-8'))
corr,orch,req,msg,ts,site,dom=sys.argv[2:9]
print(json.dumps({"headers":{"correlation_id":corr,"orchestration_id":orch,"request_id":req,
 "message_id":msg,"message_type":"request","client_id":"cli-regen","action":"process",
 "sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":ts},
 "config":{"workflow":wf},"input_data":{"site_id":site,"domain":dom}},separators=(',',':')))
PY
echo "CORRELATION_ID=$CORR"
REPO_ROOT_PUB="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT_PUB" ] || [ ! -f "$REPO_ROOT_PUB/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found — refusing to publish unverified (bugs_open/327)." >&2
  exit 1
fi
. "$REPO_ROOT_PUB/scripts/kafka-publish-lib.sh"

PUBLISH_RC=0
kafka_publish_checked \
  --topic system.agent.generic.requests \
  --correlation "$CORR" \
  --payload "$(cat "$F")" \
  --header "orchestration_id=$ORCH" \
  --header "request_id=$REQ" \
  --header "message_id=$MSG" \
  --header "message_type=request" \
  --header "client_id=cli-regen" \
  --header "action=process" \
  --header "sender_agent_type=cli" \
  --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" \
  --header "timestamp=$TS" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no meta descriptions will be regenerated (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi
echo "dispatched"
