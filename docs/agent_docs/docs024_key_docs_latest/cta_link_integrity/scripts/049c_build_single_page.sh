#!/usr/bin/env bash
# Direct page-build-handler orchestrate for ONE work item — bypasses the
# build-dispatch-loop queue when a site is STARVED rather than stalled.
#
# Why this exists (bugs_open/028 residual, 2026-07-25): the dispatcher's
# find_dispatchable_site does
#   SELECT DISTINCT ON (wi.site_id) ... ORDER BY wi.site_id, wi.priority ASC LIMIT 1
# — ONE site per 120s tick, ordered by site_id (a UUID). A site with a high
# site_id and a lower-id site holding a large backlog never gets reached.
# relojistas.com (ecf1…) sat behind webdesign.co.uk (6b49…) and its 107 triaged
# items. Priority does NOT help: it only breaks ties WITHIN a site.
#
# Unlike 049b (page-rerender, assemble-only) this runs the FULL build —
# content writer included — so it regenerates content_data. Use it when the
# content itself is what you need to change.
#
# The input_data below mirrors build-dispatch-loop's call_handler input_mapping
# exactly, so the handler sees what it would have seen via the queue.
#
# CAVEAT: firing directly bypasses the loop's `claim` and `mark_complete`
# steps, so the work item stays 'triaged' after a SUCCESSFUL build (the
# handler only writes status on its own no-op/failure paths). Mark it complete
# yourself once verified, or the dispatcher will rebuild the page later.
set -euo pipefail
WORK_ITEM_ID="$1"

read -r SITE_ID DOMAIN SPEC <<EOF
$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -F' ' -c \
  "SELECT swi.site_id, s.domain, swi.spec::text FROM site_work_items swi JOIN sites s ON s.id=swi.site_id WHERE swi.id='$WORK_ITEM_ID';")
EOF

if [ -z "${SITE_ID:-}" ]; then echo "no work item $WORK_ITEM_ID" >&2; exit 1; fi

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "corr=$CORR item=$WORK_ITEM_ID domain=$DOMAIN"
kubectl -n kafka run -i --rm "kcat-build-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR \
  -H orchestration_id=$ORCH \
  -H request_id=$REQ \
  -H message_id=$MSG \
  -H message_type=request \
  -H client_id=demo_client \
  -H action=orchestrate \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"page-build-handler"},"input_data":{"domain":"$DOMAIN","site_id":"$SITE_ID","work_item_id":"$WORK_ITEM_ID","item_type":"needs_page","source":"operator:bugfix_028","spec":$SPEC,"current_page":$SPEC}}
JSON
echo "CORR=$CORR"
