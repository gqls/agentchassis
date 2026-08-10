#!/bin/bash
# ============================================================================
# DEPLOY_asset.sh — commit a generated/uploaded image asset to a site's repo
# through the platform's own asset-deployer agent.
#
# SUPERSEDES docs/leopardessconsulting/scripts/deploy_brand_asset.sh, which
# passes `deploy_path`. That key was REMOVED 2026-08-04 (bugs_open/179 finding
# A) and an explicit value now draws a REFUSAL result — the older script cannot
# succeed as written. The deploy path is DERIVED from (asset_key, purpose) by
# storage.DeployedAssetPath; you choose it by choosing those two, never directly.
#
#   ./DEPLOY_asset.sh <domain> <s3_uri> <purpose> <asset_key> [asset_id]
#
#   purpose=logo, asset_key=logo  ->  /assets/images/logo.png   (400px, aspect
#                                     ratio preserved — a wide wordmark stays
#                                     wide; measured 400x218 on vetcomparison)
#
# WHY asset_key MUST BE PASSED EXPLICITLY (the bug this script exists to avoid).
# The live asset-deployer row declares config.asset_key as the dotted PATH
# "input_data.asset_key" and lists asset_key in input_fields. When a dispatch
# OMITS input_data.asset_key, the action's resolution ladder falls through to
# rung 2, which reads config["asset_key"] as a LITERAL string — so the file is
# committed as `input-data.asset-key.<ext>` (AssetKeyFilename maps _ to -) while
# every reader still resolves the reference through DeployedAssetPath and gets
# the correct name. The page then 404s against a file that is present under a
# placeholder name. Filed for diagnosis 2026-08-10, correlation
# 8cb3778d-c3e6-4dd8-9e80-09c0d1b0e594.
#
# s3_uri MUST come from the asset row's `storage_path`, NOT its `url` column:
# `url` holds an expiring presigned link, or — post-deploy — the local web path
# that overwrote it (bugs_open/152). Some storage_path values are bare keys and
# some are full https URLs; normalise to s3://<bucket>/<key> yourself.
#
# NEVER trust `complete`. Verify at the served artefact (this script prints the
# curl) and LOOK at the bytes — deploy_image_asset has resolved a source by
# purpose rather than asset_id in the past (bugs_open/155, 209).
# ============================================================================
set -euo pipefail

DOMAIN="${1:?domain}"
S3_URI="${2:?s3_uri (s3://bucket/key, derived from assets.storage_path)}"
PURPOSE="${3:?purpose (logo|hero|icon|...)}"
ASSET_KEY="${4:?asset_key — pass it EXPLICITLY, see header}"
ASSET_ID="${5:-}"

case "$PURPOSE" in
  favicon|og_card)
    echo "REFUSED: $PURPOSE is a brand-head purpose." >&2
    echo "derive_brand_head_assets composes those from the site logo; this" >&2
    echo "action refuses them so an arbitrary image cannot overwrite a live" >&2
    echo "favicon or social card (bugs_open/179 finding B)." >&2
    exit 2 ;;
esac

case "$S3_URI" in
  s3://*) ;;
  *) echo "REFUSED: s3_uri must be s3://bucket/key, got '$S3_URI'" >&2; exit 2 ;;
esac

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)

# kcat splits on newlines — the payload MUST be a single line.
if [ -n "$ASSET_ID" ]; then
  PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"asset-deployer"},"input_data":{"domain":"%s","s3_uri":"%s","purpose":"%s","asset_key":"%s","asset_id":"%s"}}' \
    "$DOMAIN" "$S3_URI" "$PURPOSE" "$ASSET_KEY" "$ASSET_ID")
else
  PAYLOAD=$(printf '{"action":"orchestrate","config":{"agent_type":"asset-deployer"},"input_data":{"domain":"%s","s3_uri":"%s","purpose":"%s","asset_key":"%s"}}' \
    "$DOMAIN" "$S3_URI" "$PURPOSE" "$ASSET_KEY")
fi

echo "=== asset-deployer ==="
echo "  domain:      $DOMAIN"
echo "  s3_uri:      $S3_URI"
echo "  purpose:     $PURPOSE"
echo "  asset_key:   $ASSET_KEY   (path is DERIVED from these two)"
echo "  asset_id:    ${ASSET_ID:-<omitted — assets.url will NOT be rewritten>}"
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

cat <<EOF

Verify — never trust 'complete', and note kcat can exit 0 having sent NOTHING:

  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \\
    "SELECT status, current_step, error FROM orchestration_states WHERE correlation_id='${CORRELATION_ID}'::uuid;"

  curl -s -o /tmp/deployed -w '%{http_code} %{size_download}\\n' https://${DOMAIN}\$(
    echo "/assets/images/${ASSET_KEY}.<ext>")   # ext from storage.ImagePurposes
  file /tmp/deployed    # confirm it is the image you MEANT to ship
EOF
