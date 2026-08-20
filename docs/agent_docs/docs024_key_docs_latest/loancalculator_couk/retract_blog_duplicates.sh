#!/usr/bin/env bash
# Retract the 14 duplicate /blog/ pages minted by the 2026-08-17 recompose fire.
# OWNER DECISION 2026-08-18: keep /guides/. The guides are untouched by this.
#
# ORDER MATTERS: the rows must be ARCHIVED FIRST — retract_page_deployment
# refuses an active page ("retracting a live page is not what archiving means",
# retract_page_deployment_action.go:169). Archived here 2026-08-20.
#
# ⚠ EXPLICIT page_ids, NOT the default selection. The action's default is
#   "every non-active page with a deployed_at stamp", and this site has one
#   OTHER such page — tool-standard-calc (archived 08-03, holds one of the 12
#   locked calculator rows). Its file already 404s so the delete would be a
#   no-op, but sweeping it in would be an undecided change riding along.
#
# Pre-flight already done read-only (the dry run this agent cannot express,
# since dry_run is step-config only and the shared page-retraction definition
# does not set it):
#   inbound  nav rows -> /blog/  : 0
#   inbound  chrome refs         : 0
#   inbound  other page bodies   : 0      => no refusal expected
#   outbound newly stranded      : /tools/interest-rate-stress-test.html loses
#            its only in-body inbound link, but stays in the footer nav and the
#            sitemap, so it is reported, not orphaned.
#
# ⚠ kcat -P exits 0 having sent nothing — proof is the orchestration row BY
#   CORRELATION. ⚠ No dispatch within ~300s of a chassis pod (re)start.
set -euo pipefail

SITE_ID='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
DOMAIN='loancalculator.co.uk'
PAGE_IDS='["08af3c09-c909-4441-8001-3c39c6bf540a","22c98942-3c60-4dd8-876e-ac5fd00fd83a","1e3758e1-186f-4154-8073-a9a4f05be42d","a9c9f450-1075-43eb-a6ad-3bb272d50c55","64179ba1-6661-49b1-8bdf-09f66493afdc","dff01390-2754-4fe5-bd0f-4d285a5a00f0","4d9acbc9-8e03-435e-a03d-e3a08f52fc55","0d68b24e-4027-40f2-9b71-13e39fafdc9a","f61f9829-0c65-44cc-b06f-3bfb9696edd6","864cda2c-4871-457b-9f48-de8d47e55b0a","f7c94853-03c1-428e-83c9-ef80a6675f68","9f4ff6e5-cf39-4e37-bcc1-b6e6a6d91e18","45b6670a-755f-498b-814b-a59700f5af47","fcda9b66-4900-463c-a5f9-ede36a7dd0b2"]'

CORR=$(cat /proc/sys/kernel/random/uuid)
ORCH=$(cat /proc/sys/kernel/random/uuid)
REQ=$(cat /proc/sys/kernel/random/uuid)
MSG=$(cat /proc/sys/kernel/random/uuid)
TS=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

echo "== FIRING page-retraction corr=$CORR =="
kubectl -n kafka run -i --rm "kcat-retract-$(date +%s)" \
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
