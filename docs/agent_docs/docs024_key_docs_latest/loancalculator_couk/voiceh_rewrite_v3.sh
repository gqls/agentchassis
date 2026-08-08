#!/usr/bin/env bash
# Fire ONE framework content_rewrite (voice H) at ONE loancalculator page.
#
# Why this shape (2026-08-07): the owner approved voice H and asked for the rest
# of the site's copy to be rerun THROUGH THE FRAMEWORK — not hand-written through
# this CLI (standing ruling 2026-08-04). So every page goes through
# page-build-handler exactly as the 204 canary did.
#
# The prompt is COPIED BY SQL from the original canary item 2517bc4b, never
# retyped, so it cannot drift from the spec the owner reviewed. Only
# page/page_id/page_name are substituted.
#
# Two things this script deliberately does NOT do:
#   * it does not file the item as 'triaged'/'approved' — those are the
#     dispatcher's own statuses and would earn the page a SECOND, unrequested
#     build later (049c's caveat). It files 'detected', which nothing polls,
#     and the direct publish below is the only dispatch.
#   * it does not mark the item complete. Verify first, then complete — the
#     handler only writes status on its no-op/failure paths, so a 'detected'
#     row at the end means "not yet graded", which is the honest state.
#
# Usage: voiceh_rewrite.sh <page-name>          e.g. voiceh_rewrite.sh guide-jargon-buster
set -euo pipefail

SITE_ID='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
DOMAIN='loancalculator.co.uk'
CANARY_ITEM='6d52beaf-9555-4178-a274-4d70d5019af5'   # guidance v3: varied openings + preserve <style> blocks byte-for-byte
PAGE_NAME="$1"

PSQL() { kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db "$@"; }

PAGE_ID=$(PSQL -t -A -c "SELECT id FROM pages WHERE site_id='$SITE_ID' AND name='$PAGE_NAME' AND status='active';")
if [ -z "${PAGE_ID:-}" ]; then echo "no ACTIVE page named $PAGE_NAME" >&2; exit 1; fi

# Insert the work item, spec built from the canary's spec with the page fields swapped.
# NOTE: the INSERT is wrapped in a CTE deliberately — a bare `INSERT … RETURNING`
# prints its "INSERT 0 1" command tag on stdout even under -t -A, and that tag
# rides along into the variable and poisons the next query with a bad uuid.
ITEM_ID=$(PSQL -t -A -v ON_ERROR_STOP=1 -c "
WITH ins AS (
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, page_id, status, created_by,
   handler_agent, item_key, priority)
SELECT '$SITE_ID', 'voiceh-rollout', 'content_rewrite', 'medium',
       'Voice H rewrite - $PAGE_NAME',
       (swi.spec
         || jsonb_build_object('page','$PAGE_NAME','page_name','$PAGE_NAME','page_id','$PAGE_ID')),
       '$PAGE_ID', 'detected', 'voiceh-rollout',
       'page-build-handler', 'voiceh-v3:$PAGE_NAME', 50
FROM site_work_items swi WHERE swi.id='$CANARY_ITEM'
RETURNING id)
SELECT id FROM ins;")
if [ -z "${ITEM_ID:-}" ]; then echo "insert produced no id for $PAGE_NAME" >&2; exit 1; fi

SPEC=$(PSQL -t -A -c "SELECT spec::text FROM site_work_items WHERE id='$ITEM_ID';")

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "page=$PAGE_NAME item=$ITEM_ID corr=$CORR orch=$ORCH"

kubectl -n kafka run -i --rm "kcat-voiceh-$(date +%s%N | tail -c 7)" \
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
{"action":"orchestrate","config":{"agent_type":"page-build-handler"},"input_data":{"domain":"$DOMAIN","site_id":"$SITE_ID","work_item_id":"$ITEM_ID","item_type":"content_rewrite","page_name":"$PAGE_NAME","source":"voiceh-rollout","spec":$SPEC,"current_page":$SPEC}}
JSON

# kcat -P exits 0 having sent nothing (LANDMINE). The only proof of dispatch is
# the orchestration row appearing — check it, do not trust this script's exit.
echo "ITEM=$ITEM_ID ORCH=$ORCH"
