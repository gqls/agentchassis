#!/usr/bin/env bash
# Retract tool-credit-roadmap — the DUPLICATE credit-health-check tool page.
# OWNER DECISION 2026-08-23: "we only need one of them". Kept:
# tool-credit-health-check, whose instance of the shared component is the LOCKED,
# golden-verified one; this page held an unlocked second instance.
#
# ORDER (all satisfied before firing, 2026-08-23):
#   1. page ARCHIVED (retract_page_deployment refuses an active page, :169)
#   2. inbound links CLEARED to 0 — 16 instances across 15 pages at the start;
#      archiving first meant the tool-page rebuilds and the template-fix rerender
#      wave regenerated their cross-link sections without it. MEASURED 0 pages,
#      0 chrome refs before firing. The action would REFUSE otherwise, and that
#      refusal is the guard working, not an obstacle.
#   3. the nav row is left to the action — it deactivates nav rows itself.
#
# ⚠ EXPLICIT page id, NOT the default sweep: the default is "every non-active page
#   with a deployed_at stamp", which on this site also matches tool-standard-calc
#   (archived 08-03, holds one of the 12 locked rows).
set -euo pipefail

SITE_ID='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
DOMAIN='loancalculator.co.uk'
PAGE_IDS='["e4e94578-89df-4656-b37a-95a0122d7e8c"]'

CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid);  MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "== FIRING page-retraction corr=$CORR =="
kubectl -n kafka run -i --rm "kcat-retract-cr-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -c 1 \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H orchestration_id=$ORCH -H request_id=$REQ \
  -H message_id=$MSG -H message_type=request -H client_id=demo_client \
  -H action=orchestrate -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses -H timestamp=$TS <<JSON
{"action":"orchestrate","config":{"agent_type":"page-retraction"},"input_data":{"site_id":"$SITE_ID","domain":"$DOMAIN","page_ids":$PAGE_IDS}}
JSON
echo "CORR=$CORR"
