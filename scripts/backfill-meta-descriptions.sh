#!/usr/bin/env bash
# Dispatch the meta-description-backfiller at ONE site (bugs_open/320, SEO-004).
#
# Reads the LIVE agent config and sends it inline, so what runs is exactly what is
# seeded — a copy pasted into this script would drift from the row the moment
# anyone migrates it.
#
# Usage:  ./scripts/backfill-meta-descriptions.sh <domain>
#
# SAFE TO RE-RUN. The action's overwrite_existing defaults false, so a page that
# already has a description is left alone (proven at row level on loanzy.uk: a
# second run touched only the one still-blank page). It fills blanks, nothing else.
#
# ⚠ VERIFY AT THE ARTEFACT, NOT THE STATUS. The first canary run reported COMPLETED
# and wrote nothing (bugs_open/313's silent-skip). The check is the pages table, and
# then the served page — `rerender_single_page_action.go` strips an EMPTY description
# tag rather than serving it, so a page with none shows an ABSENT tag, not an empty
# one, and a DB row updated before the page is rerendered will disagree with what a
# visitor gets.
set -euo pipefail

DOMAIN="${1:?usage: $0 <domain>}"
NS=ai-persona-system
PSQL="kubectl -n $NS exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

SITE_ID=$($PSQL -tAc "SELECT id FROM sites WHERE domain='$DOMAIN';" | tr -d '[:space:]')
[ -n "$SITE_ID" ] || { echo "no site row for $DOMAIN" >&2; exit 2; }

BEFORE=$($PSQL -tAc "SELECT count(*) FROM pages WHERE site_id='$SITE_ID' AND status='active' AND COALESCE(meta_description,'')='';" | tr -d '[:space:]')
echo "$DOMAIN  site_id=$SITE_ID  empty_before=$BEFORE"
[ "$BEFORE" -gt 0 ] || { echo "nothing to do"; exit 0; }

WF=$(mktemp); trap 'rm -f "$WF" "$MSG_F"' EXIT
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='meta-description-backfiller' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"

CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
MSG_F=$(mktemp)
python3 - "$WF" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE_ID" "$DOMAIN" > "$MSG_F" <<'PY'
import json,sys
wf=json.load(open(sys.argv[1],encoding='utf-8'))
corr,orch,req,msg,ts,site,dom=sys.argv[2:9]
print(json.dumps({"headers":{"correlation_id":corr,"orchestration_id":orch,"request_id":req,
  "message_id":msg,"message_type":"request","client_id":"cli-backfill","action":"process",
  "sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":ts},
  "config":{"workflow":wf},"input_data":{"site_id":site,"domain":dom}},separators=(',',':')))
PY

echo "CORRELATION_ID=$CORR"
kubectl -n kafka run -i --rm "kcat-mdb-$(date +%s)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H orchestration_id=$ORCH -H request_id=$REQ -H message_id=$MSG \
  -H message_type=request -H client_id=cli-backfill -H action=process \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$TS \
  < "$MSG_F" >/dev/null 2>&1

echo "dispatched. check with:"
echo "  $PSQL -c \"SELECT current_step,status FROM orchestration_states WHERE correlation_id::text='$CORR';\""
echo "  $PSQL -c \"SELECT name, length(COALESCE(meta_description,'')) FROM pages WHERE site_id='$SITE_ID' AND status='active' ORDER BY name;\""
