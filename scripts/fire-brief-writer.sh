#!/bin/bash
# Dispatch the brief-writer at one domain (migration 510).
#
# Usage:
#   ./scripts/fire-brief-writer.sh <domain> [ "a few words of direction" ]
#
# It reads the LIVE workflow out of agent_definitions and ships it in the envelope,
# so the run uses exactly what is seeded rather than a copy that can drift.
#
# WHAT IT PRODUCES: a `mission_brief` spec for the domain, plus a
# `needs_brief_review` work item held at `status='needs_human_review'` — which the
# dispatcher cannot pick up (it selects `status IN ('triaged','approved')`). So this
# writes a brief and BUILDS NOTHING. Releasing the build is a separate, human act.
#
# ⚠ `kcat -P` EXITS 0 HAVING SENT NOTHING (LANDMINES). The publish is therefore not
# evidence. This script prints the correlation id and the query that reads the
# durable record; a run that produced no orchestration row was not dispatched,
# however cleanly the command exited.
#
# ⚠ The site row must exist — `write_site_spec` needs a site_id. For a test domain,
# create it LOCKED (`locked_at = now()`), so that even if something else tries to
# dispatch a build for it, `find_dispatchable_site` excludes it.
set -uo pipefail

DOMAIN="${1:-}"
DIRECTION="${2:-}"
[ -n "$DOMAIN" ] || { echo "usage: $0 <domain> [direction]" >&2; exit 2; }

PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

SITE_ID=$($PSQL -tAc "SELECT id FROM sites WHERE domain='${DOMAIN//\'/\'\'}';" | tr -d '[:space:]')
[ -n "$SITE_ID" ] || { echo "no sites row for $DOMAIN — create one first (LOCKED, if it is a test)" >&2; exit 3; }

LOCKED=$($PSQL -tAc "SELECT locked_at IS NOT NULL FROM sites WHERE id='$SITE_ID';" | tr -d '[:space:]')
echo "site:   $DOMAIN ($SITE_ID)  locked=$LOCKED"
[ "$LOCKED" = "t" ] || echo "  ⚠ this site is NOT locked — the brief-writer builds nothing, but any OTHER dispatcher can pick the site up"

WF=$(mktemp); MSG_F=$(mktemp)
trap 'rm -f "$WF" "$MSG_F"' EXIT
$PSQL -tAc "SELECT default_config->'workflow' FROM agent_definitions
  WHERE type='brief-writer' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$WF"
[ -s "$WF" ] || { echo "empty workflow read — refusing to dispatch (is 510 applied?)" >&2; exit 4; }

# The research query. Deliberately derived from the domain rather than asked for:
# a brief-writer run should need nothing but a domain name, which is the whole
# point of it. Kept under 200 bytes — web_search DROPS a >=200-char query and the
# failure names config keys, not length (LANDMINES).
LABEL=$(printf '%s' "$DOMAIN" | sed 's/\..*$//; s/-/ /g')
QUERY="$LABEL: what the subject covers, who looks for it, the main brands or providers, and the tools or data people use"
QUERY=${QUERY:0:190}

CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

python3 - "$WF" "$CORR" "$ORCH" "$REQ" "$MSG" "$TS" "$SITE_ID" "$DOMAIN" "$QUERY" "$DIRECTION" > "$MSG_F" <<'PY'
import json, sys
wf = json.load(open(sys.argv[1], encoding='utf-8'))
corr, orch, req, msg, ts, site, dom, query, direction = sys.argv[2:11]
print(json.dumps({
    "headers": {"correlation_id": corr, "orchestration_id": orch, "request_id": req,
                "message_id": msg, "message_type": "request", "client_id": "cli-brief-writer",
                "action": "process",
                "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"},
                "timestamp": ts},
    "config": {"workflow": wf},
    "input_data": {"site_id": site, "domain": dom,
                   "research_query": query, "direction": direction}},
    separators=(',', ':')))
PY

[ "$(wc -l < "$MSG_F")" -eq 1 ] || { echo "envelope is not one line — refusing (kcat publishes per line)" >&2; exit 5; }
echo "query:  $QUERY"
[ -n "$DIRECTION" ] && echo "direction: $DIRECTION"

kubectl -n kafka run -i --rm "kcat-brief-$(date +%s)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H orchestration_id=$ORCH -H request_id=$REQ -H message_id=$MSG \
  -H message_type=request -H client_id=cli-brief-writer -H action=process \
  -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$TS \
  < "$MSG_F" >/dev/null 2>&1

cat <<EOF

CORRELATION_ID=$CORR
ORCHESTRATION_ID=$ORCH

Exit 0 proves nothing — kcat -P can send nothing and still exit 0. Read the durable record:

  # did it start?
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \\
    "SELECT current_step, status FROM orchestration_states WHERE orchestration_id='$ORCH';"

  # did the brief land?
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \\
    "SELECT jsonb_pretty(data) FROM site_specs ss JOIN sites s ON s.id=ss.site_id
      WHERE s.domain='$DOMAIN' AND ss.aspect='mission_brief' AND ss.is_current;"

  # is the build held? (expect one row, needs_human_review)
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \\
    "SELECT item_type, status, summary FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
      WHERE s.domain='$DOMAIN' AND wi.item_type='needs_brief_review';"
EOF
