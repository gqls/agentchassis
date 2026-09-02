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
# ⚠ `kcat -P` EXITS 0 HAVING SENT NOTHING (LANDMINES) — FIXED 2026-08-24, bugs_open/327.
# This script used to carry that warning and publish through the racing form anyway,
# with `>/dev/null 2>&1` discarding both streams so there was no receipt in either
# direction. It now publishes through `scripts/kafka-publish-lib.sh`, which asserts the
# receipt and EXITS NON-ZERO when nothing was sent. The correlation id below is printed
# only after a confirmed publish, so it names something.
# The durable-record queries still stand: a receipt proves the broker took the bytes,
# never that the work happened.
#
# ⚠ The site row must exist — `write_site_spec` needs a site_id. For a test domain,
# create it LOCKED (`locked_at = now()`), so that even if something else tries to
# dispatch a build for it, `find_dispatchable_site` excludes it.
#
# ⚠ AND the row must carry `name` and `network_id` (2026-09-02, first release):
#   INSERT INTO sites (domain, name, network_id, status, locked_at)
#   VALUES ('<domain>', '<domain>', '00000000-0000-0000-0000-000000000002', 'test', now());
# A minimal row (domain+status only) satisfies THIS script and the whole brief flow,
# then fails at RELEASE — weeks later — when the build's ensure_site_record scans
# name/network_id without COALESCE: "Scan error on column \"name\": converting NULL
# to string". Measured on advertise.co.uk: brief written 08-26, release failed 09-02;
# all six brief-lane rows carried the same NULLs and were backfilled that day.
set -uo pipefail

DOMAIN="${1:-}"
DIRECTION="${2:-}"
[ -n "$DOMAIN" ] || { echo "usage: $0 <domain> [direction]" >&2; exit 2; }

PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

SITE_ID=$($PSQL -tAc "SELECT id FROM sites WHERE domain='${DOMAIN//\'/\'\'}';" | tr -d '[:space:]')
[ -n "$SITE_ID" ] || { echo "no sites row for $DOMAIN — create one first with name+network_id set (see header; LOCKED, if it is a test)" >&2; exit 3; }

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
  --header "orchestration_id=$ORCH" --header "request_id=$REQ" --header "message_id=$MSG" \
  --header "message_type=request" --header "client_id=cli-brief-writer" --header "action=process" \
  --header "sender_agent_type=cli" --header "sender_agent_id=cli-user" \
  --header "responses_topic=system.agent.generic.responses" --header "timestamp=$TS" || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — no brief will be written for ${DOMAIN}." >&2
  exit "$PUBLISH_RC"
fi

cat <<EOF

CORRELATION_ID=$CORR
ORCHESTRATION_ID=$ORCH

Published (receipt asserted). That proves the broker took it, not that the work happened —
read the durable record:

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
