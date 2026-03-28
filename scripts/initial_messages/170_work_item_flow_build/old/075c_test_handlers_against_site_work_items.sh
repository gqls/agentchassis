#!/bin/bash
# ============================================================================
# test_handlers.sh — Test handler agents against site_work_items
# ============================================================================
# Run each handler independently before wiring up the orchestrator.
# Each test reads the spec from a work item and sends it to the handler.
#
# Prerequisites:
#   - Discovery has run (site_work_items has rows)
#   - handler_agent = 'asset-deployer' (not 'asset-deploy-agent')
#   - S3 + git adapter running (for asset-deployer)
# ============================================================================

set -euo pipefail

# ============================================================================
# TEST 1: asset-deployer (undeployed_asset items)
# ============================================================================
# The asset-deployer expects: domain, s3_uri, purpose
# These come from the work item's spec field.
#
# Query to see what we'd send:
#   SELECT spec->>'purpose' as purpose,
#          spec->>'url' as s3_uri,
#          s.domain
#   FROM site_work_items wi
#   JOIN sites s ON s.id = wi.site_id
#   WHERE wi.item_type = 'undeployed_asset'
#   ORDER BY wi.created_at;
#
# The asset-deployer's workflow is:
#   deploy_asset (deploy_image_asset with input_fields) → complete
# It uses input_fields: ["s3_uri", "deploy_path", "purpose", "domain"]
# So we pass these in input_data.

test_asset_deployer() {
    local DOMAIN="$1"
    local S3_URI="$2"
    local PURPOSE="$3"

    CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
    ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
    MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
    REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    echo "========================================="
    echo "Testing asset-deployer"
    echo "  Domain:  $DOMAIN"
    echo "  Purpose: $PURPOSE"
    echo "  S3 URI:  ${S3_URI:0:80}..."
    echo "  CorrID:  $CORRELATION_ID"
    echo "========================================="

    kubectl -n kafka run -i --rm kcat-asset-$(date +%s) \
      --image=edenhill/kcat:1.7.1 \
      --restart=Never -- \
      kcat -P \
      -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
      -t system.agent.generic.requests \
      -H correlation_id=$CORRELATION_ID \
      -H orchestration_id=$ORCHESTRATION_ID \
      -H request_id=$REQUEST_ID \
      -H message_id=$MESSAGE_ID \
      -H message_type=request \
      -H client_id=demo_client \
      -H action=process \
      -H sender_agent_type=cli \
      -H sender_agent_id=cli-user \
      -H responses_topic=system.agent.generic.responses \
      -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_deployer","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_deployer":{"action":"spawn_agent","config":{"role":"deployer","agent_type":"asset-deployer"},"output_field":"deployer","next_step":"call_deployer","description":"Spawn asset deployer"},"call_deployer":{"action":"call_agent","config":{"agent_type":"asset-deployer","target_role":"deployer","input_mapping":{"domain":"input_data.domain","s3_uri":"input_data.s3_uri","purpose":"input_data.purpose"},"timeout_seconds":120},"output_field":"deploy_result","next_step":"complete","description":"Deploy single asset"},"complete":{"action":"complete_workflow","config":{"output_fields":["deploy_result"]},"description":"Done"}}}},"input_data":{"domain":"${DOMAIN}","s3_uri":"${S3_URI}","purpose":"${PURPOSE}"}}
JSON

    echo ""
    echo "Monitor: kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
    echo "CORRELATION_ID=$CORRELATION_ID"
}

# ============================================================================
# TEST 2: webdesign-agent (missing_css item)
# ============================================================================
# The webdesign-agent loads site context itself via its workflow.
# It just needs site_id (or it can load from domain via ensure_site_record).
#
# Its workflow:
#   load_site_for_design → analyze_design (LLM) → generate_css (LLM) → deploy_css (git) → complete

test_webdesign_agent() {
    local DOMAIN="$1"

    CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
    ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
    MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
    REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    echo "========================================="
    echo "Testing webdesign-agent"
    echo "  Domain: $DOMAIN"
    echo "  CorrID: $CORRELATION_ID"
    echo "========================================="

    kubectl -n kafka run -i --rm kcat-webdesign-$(date +%s) \
      --image=edenhill/kcat:1.7.1 \
      --restart=Never -- \
      kcat -P \
      -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
      -t system.agent.generic.requests \
      -H correlation_id=$CORRELATION_ID \
      -H orchestration_id=$ORCHESTRATION_ID \
      -H request_id=$REQUEST_ID \
      -H message_id=$MESSAGE_ID \
      -H message_type=request \
      -H client_id=demo_client \
      -H action=process \
      -H sender_agent_type=cli \
      -H sender_agent_id=cli-user \
      -H responses_topic=system.agent.generic.responses \
      -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_designer","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"spawn_designer":{"action":"spawn_agent","config":{"role":"webdesigner","agent_type":"webdesign-agent"},"output_field":"designer","next_step":"call_designer","description":"Spawn webdesign agent"},"call_designer":{"action":"call_agent","config":{"agent_type":"webdesign-agent","target_role":"webdesigner","input_mapping":{"domain":"input_data.domain"},"timeout_seconds":300},"output_field":"design_result","next_step":"complete","description":"Generate and deploy CSS"},"complete":{"action":"complete_workflow","config":{"output_fields":["design_result"]},"description":"Done"}}}},"input_data":{"domain":"${DOMAIN}"}}
JSON

    echo ""
    echo "Monitor: kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
    echo "CORRELATION_ID=$CORRELATION_ID"
}

# ============================================================================
# USAGE
# ============================================================================

case "${1:-help}" in
    asset)
        # Get the S3 URIs from the database first:
        # kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db -c \
        #   "SELECT spec->>'purpose', spec->>'url' FROM site_work_items WHERE item_type='undeployed_asset';"
        #
        # Then run with actual values:
        DOMAIN="${2:?Usage: $0 asset <domain> <s3_uri> <purpose>}"
        S3_URI="${3:?Provide S3 URI}"
        PURPOSE="${4:?Provide purpose (hero/logo)}"
        test_asset_deployer "$DOMAIN" "$S3_URI" "$PURPOSE"
        ;;
    css)
        DOMAIN="${2:?Usage: $0 css <domain>}"
        test_webdesign_agent "$DOMAIN"
        ;;
    help|*)
        echo "Usage:"
        echo "  $0 asset <domain> <s3_uri> <purpose>   — test asset-deployer"
        echo "  $0 css <domain>                         — test webdesign-agent (CSS generation)"
        echo ""
        echo "Example:"
        echo "  $0 asset finetuning.uk 's3://personae-prod-uk001-images/images/...' hero"
        echo "  $0 css finetuning.uk"
        ;;
esac