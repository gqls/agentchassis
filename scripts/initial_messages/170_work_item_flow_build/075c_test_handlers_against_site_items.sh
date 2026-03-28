#!/bin/bash
# ============================================================================
# 075c_test_handlers_against_site_work_items.sh
# ============================================================================
# Test handler agents individually before wiring up the dispatch loop.
# Each handler is self-contained — receives raw identifiers and loads its own context.
#
# Prerequisites:
#   - Discovery has run (site_work_items has rows)
#   - Items triaged: status = 'triaged'
#   - handler_agent values match real agent types (asset-deployer, webdesign-agent)
#   - S3 + git adapter running (for asset-deployer and webdesign-agent)
#
# Usage:
#   ./075c_test_handlers_against_site_work_items.sh css finetuning.uk
#   ./075c_test_handlers_against_site_work_items.sh asset finetuning.uk hero
#   ./075c_test_handlers_against_site_work_items.sh asset finetuning.uk logo
#   ./075c_test_handlers_against_site_work_items.sh status finetuning.uk
# ============================================================================

set -euo pipefail

KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
PSQL_PREFIX="kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db"

gen_ids() {
    CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
    ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
    MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
    REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
}

send_to_generic() {
    local JSON_PAYLOAD="$1"
    kubectl -n kafka run -i --rm kcat-handler-$(date +%s) \
      --image=edenhill/kcat:1.7.1 \
      --restart=Never -- \
      kcat -P \
      -b $KAFKA_BOOTSTRAP \
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
      -H timestamp=$TIMESTAMP <<< "$JSON_PAYLOAD"
}

print_monitor() {
    echo ""
    echo "Monitor:"
    echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
    echo ""
    echo "CORRELATION_ID=$CORRELATION_ID"
}

# ============================================================================
# TEST 1: asset-deployer (undeployed_asset items)
# ============================================================================
# The asset-deployer is self-contained. It receives:
#   domain, asset_id, purpose
# and resolves the s3:// URI itself via resolveStorageURIFromAsset:
#   1. Checks site content_data->>'{purpose}_uri' (s3:// URI from StoreAssetAction)
#   2. Falls back to converting the asset's presigned URL via PresignedURLToS3URI
#
# No need to look up or paste s3_uri manually — the agent handles it.

test_asset_deployer() {
    local DOMAIN="$1"
    local PURPOSE="${2:-hero}"
    gen_ids

    # Auto-lookup asset_id from work items
    echo "Looking up asset_id for $DOMAIN purpose=$PURPOSE..."
    ASSET_ID=$($PSQL_PREFIX -t -A -c "
        SELECT spec->>'asset_id'
        FROM site_work_items
        WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN')
          AND item_type = 'undeployed_asset'
          AND spec->>'purpose' = '$PURPOSE'
        LIMIT 1;
    " 2>/dev/null | tr -d '[:space:]')

    if [ -z "$ASSET_ID" ]; then
        echo "No undeployed_asset work item found for $DOMAIN purpose=$PURPOSE"
        echo "Run discovery first: ./075_trigger_discovery.sh $DOMAIN design"
        exit 1
    fi

    echo "========================================="
    echo "Testing asset-deployer (self-resolving)"
    echo "  Domain:   $DOMAIN"
    echo "  Purpose:  $PURPOSE"
    echo "  Asset ID: $ASSET_ID"
    echo "  CorrID:   $CORRELATION_ID"
    echo "========================================="

    # Inline workflow: spawn asset-deployer → call with raw identifiers → complete
    # The agent resolves s3_uri from asset_id via DB lookup.
    send_to_generic "$(cat <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_deployer","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_deployer":{"action":"spawn_agent","config":{"role":"deployer","agent_type":"asset-deployer"},"output_field":"spawn_deployer","next_step":"call_deployer","description":"Spawn asset deployer"},"call_deployer":{"action":"call_agent","config":{"target_role":"deployer","input_mapping":{"domain":"input_data.domain","asset_id":"input_data.asset_id","purpose":"input_data.purpose"},"timeout_seconds":120},"output_field":"deploy_result","next_step":"complete","description":"Deploy single asset (agent resolves s3_uri from asset_id)"},"complete":{"action":"complete_workflow","config":{"output_fields":["deploy_result"]},"description":"Done"}}}},"input_data":{"domain":"${DOMAIN}","asset_id":"${ASSET_ID}","purpose":"${PURPOSE}"}}
JSON
    )"

    print_monitor
}

# ============================================================================
# TEST 2: webdesign-agent (missing_css item)
# ============================================================================
# The webdesign-agent is self-contained. It receives:
#   domain (and optionally site_id)
# and loads everything else via its own workflow:
#   check_site_context → load_site_for_design → analyze → generate_css → deploy
#
# No work-item-specific inputs needed.

test_webdesign_agent() {
    local DOMAIN="$1"
    gen_ids

    echo "========================================="
    echo "Testing webdesign-agent"
    echo "  Domain: $DOMAIN"
    echo "  CorrID: $CORRELATION_ID"
    echo "========================================="

    send_to_generic "$(cat <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_designer","processing_mode":"orchestrator","timeout_seconds":600,"steps":{"spawn_designer":{"action":"spawn_agent","config":{"role":"webdesigner","agent_type":"webdesign-agent"},"output_field":"spawn_designer","next_step":"call_designer","description":"Spawn webdesign agent"},"call_designer":{"action":"call_agent","config":{"target_role":"webdesigner","input_mapping":{"domain":"input_data.domain"},"timeout_seconds":300},"output_field":"design_result","next_step":"complete","description":"Generate and deploy CSS"},"complete":{"action":"complete_workflow","config":{"output_fields":["design_result"]},"description":"Done"}}}},"input_data":{"domain":"${DOMAIN}"}}
JSON
    )"

    print_monitor
}

# ============================================================================
# STATUS: Check work item status for a site
# ============================================================================

do_status() {
    local DOMAIN="$1"

    echo "Work items for $DOMAIN:"
    $PSQL_PREFIX -c "
        SELECT item_type, handler_agent, status, priority,
               spec->>'purpose' as purpose,
               spec->>'asset_id' as asset_id,
               result->>'commit_sha' as commit_sha,
               updated_at::timestamp(0) as updated
        FROM site_work_items
        WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN')
        ORDER BY priority, created_at;
    "
}

# ============================================================================
# USAGE
# ============================================================================

case "${1:-help}" in
    asset)
        DOMAIN="${2:?Usage: $0 asset <domain> [purpose]}"
        PURPOSE="${3:-hero}"
        test_asset_deployer "$DOMAIN" "$PURPOSE"
        ;;
    css)
        DOMAIN="${2:?Usage: $0 css <domain>}"
        test_webdesign_agent "$DOMAIN"
        ;;
    status)
        DOMAIN="${2:?Usage: $0 status <domain>}"
        do_status "$DOMAIN"
        ;;
    help|*)
        echo "Usage:"
        echo "  $0 asset <domain> [purpose]  — test asset-deployer (resolves s3_uri from asset_id)"
        echo "  $0 css <domain>               — test webdesign-agent (CSS generation)"
        echo "  $0 status <domain>            — check work item status"
        echo ""
        echo "Examples:"
        echo "  $0 asset finetuning.uk hero   — deploy hero image"
        echo "  $0 asset finetuning.uk logo   — deploy logo"
        echo "  $0 css finetuning.uk          — generate and deploy CSS"
        echo ""
        echo "Test order:"
        echo "  1. ./075_trigger_discovery.sh finetuning.uk design   # find issues"
        echo "  2. $0 status finetuning.uk                           # see what was found"
        echo "  3. Triage items (see 075a step 5)"
        echo "  4. $0 css finetuning.uk                              # test one handler"
        echo "  5. $0 asset finetuning.uk hero                       # test another"
        echo "  6. ./075d_trigger_maintenance.sh maintain finetuning.uk  # or full loop"
        ;;
esac