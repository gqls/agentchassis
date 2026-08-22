#!/usr/bin/env bash
# 235 close-out: retract the stale /assets/images/logo.jpg from one site via
# retract_asset_files (DGH-010, asset-retraction agent, seed 446). Generalised
# from staged_component_build/scripts/RETRACT_gaswholesalers_logo_jpg.sh with
# the robust kcat pattern (that script's own pipe form is the
# kcat-publish-silently-drops trap it warns about at its foot).
#
# Usage:          ./fire_235_retract_logo_jpg.sh <site_id> <domain>   # DRY RUN
#                 ARM=1 ./fire_235_retract_logo_jpg.sh <site_id> <domain>
# Verify (armed): control PAIR at the wire — logo.jpg 404 AND logo.png 200 —
# plus the ASSET_RETRACTION_* rows in agent_error_log.
# ⚠ step_overrides reaching the step config is [UNVERIFIED] for this agent
# shape; an armed run still reporting dry_run:true fails SAFE — fallback is a
# snapshot-first one-off config UPDATE.
set -euo pipefail

SITE_ID="${1:?site_id}"; DOMAIN="${2:?domain}"
BROKER="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
TOPIC="system.agent.generic.requests"
CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

if [ "${ARM:-0}" = "1" ]; then
  CONFIG='{"agent_type":"asset-retraction","step_overrides":{"retract":{"dry_run":false}}}'
  echo "ARMED — this run deletes on $DOMAIN."
else
  CONFIG='{"agent_type":"asset-retraction"}'
  # ⚠ Whether this is REALLY a dry run is decided by the LIVE step config, not
  # this banner — on 2026-08-22 ten "dry runs" deleted their files because an
  # operator edit had baked dry_run:false into the live row (LANDMINES:
  # "asset-retraction dry run by default was a LIE"; disarmed by migration 554).
  # Read it first, every time:
  #   SELECT default_config #> '{workflow,steps,retract,config}' FROM agent_definitions
  #   WHERE type='asset-retraction' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  echo "No override sent on $DOMAIN — the LIVE step config decides dry-run vs armed. READ IT FIRST (see comment). ARM=1 sends dry_run:false."
fi

JSON=$(printf '{"action":"orchestrate","config":%s,"input_data":{"site_id":"%s","paths":["/assets/images/logo.jpg"]}}' "$CONFIG" "$SITE_ID")

CMD="printf '%s' '${JSON}' | kcat -P -b ${BROKER} -t ${TOPIC} \
 -H correlation_id=${CORRELATION_ID} -H request_id=$(cat /proc/sys/kernel/random/uuid) \
 -H message_id=$(cat /proc/sys/kernel/random/uuid) -H orchestration_id=$(cat /proc/sys/kernel/random/uuid) \
 -H orchestration_name=asset-retract-${DOMAIN}-$(date +%Y%m%d-%H%M%S) \
 -H step_name=start -H client_id=demo_client -H message_type=request \
 -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
 -H responses_topic=system.agent.generic.responses -H timestamp=${TIMESTAMP} && echo PUBLISH_OK"

echo "DOMAIN=$DOMAIN  CORRELATION_ID=$CORRELATION_ID"
kubectl -n kafka run "kcat-retract-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "$CMD"
