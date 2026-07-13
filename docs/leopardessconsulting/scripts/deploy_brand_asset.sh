#!/bin/bash
# ============================================================================
# deploy_brand_asset.sh — commit a brand asset to a site's git repo via the
# platform's own asset-deployer agent. No external tooling.
#
# The asset must already exist in the image bucket. asset-deployer runs
# deploy_image_asset: download from S3 -> optimise by purpose -> commit to git.
#
#   purpose=logo  -> 400x400, quality 90, .png  (storage.ImagePurposes)
#   deploy_path   -> path within the site dir; git-adapter prefixes "{domain}/"
#
# Usage:
#   ./deploy_brand_asset.sh <domain> <s3_uri> <deploy_path> <purpose> <asset_key>
#
# Example:
#   ./deploy_brand_asset.sh leopardessconsulting.co.uk \
#     s3://personae-prod-uk001-images/images/demo_client/20260710/leopardess-logo-master-1024.png \
#     assets/images/logo.png logo logo
#
# NOTE: asset-deployer's input_fields do NOT include asset_id, so it will not
# rewrite assets.url. Write the row with url='/assets/images/<file>' yourself.
# (RUNBOOK landmine 6.)
# ============================================================================
set -euo pipefail

DOMAIN="${1:?domain}"
S3_URI="${2:?s3_uri}"
DEPLOY_PATH="${3:?deploy_path}"
PURPOSE="${4:-logo}"
ASSET_KEY="${5:-logo}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

# kcat splits on newlines — the payload MUST be a single line.
PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"asset-deployer"},"input_data":{"domain":"%s","s3_uri":"%s","deploy_path":"%s","purpose":"%s","asset_key":"%s"}}' \
  "$DOMAIN" "$S3_URI" "$DEPLOY_PATH" "$PURPOSE" "$ASSET_KEY")

echo "=== asset-deployer ==="
echo "  domain:      $DOMAIN"
echo "  s3_uri:      $S3_URI"
echo "  deploy_path: $DEPLOY_PATH  (committed as ${DOMAIN}/${DEPLOY_PATH})"
echo "  purpose:     $PURPOSE"
echo "  correlation: $CORRELATION_ID"
echo ""

echo "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-asset-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORRELATION_ID \
  -H request_id=$REQUEST_ID \
  -H message_id=$MESSAGE_ID \
  -H orchestration_id=$ORCHESTRATION_ID \
  -H orchestration_name=asset-${DOMAIN}-$(date +%Y%m%d-%H%M%S) \
  -H step_name=start \
  -H client_id=demo_client \
  -H message_type=request \
  -H action=orchestrate \
  -H from_agent_type=user \
  -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses

echo ""
echo "Verify (never trust 'complete' — check the artifact):"
echo "  kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -c \\"
echo "    \"SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;\""
echo "  curl -s -o /dev/null -w '%{http_code}\\n' https://${DOMAIN}/${DEPLOY_PATH}"
