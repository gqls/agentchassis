#!/bin/bash
# ============================================================================
# 075d_trigger_maintenance.sh — Orchestrator-level maintenance commands
# ============================================================================
#
# Commands:
#   maintain  — trigger the full dispatch loop (site-work-orchestrator)
#   triage    — move detected items to triaged (approve for processing)
#   status    — check work item status for a site
#
# For individual handler testing, use 075c_test_handlers_against_site_work_items.sh
#
# Usage:
#   ./075d_trigger_maintenance.sh maintain finetuning.uk
#   ./075d_trigger_maintenance.sh triage finetuning.uk
#   ./075d_trigger_maintenance.sh status finetuning.uk
# ============================================================================

set -euo pipefail

KAFKA_BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
PSQL_CMD="kubectl -n ai-persona-system exec -it deploy/api-server -- psql -U clients_user -d clients_db"

gen_ids() {
    CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
    ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
    MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
    REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
}

send_to_generic() {
    local JSON_PAYLOAD="$1"
    kubectl -n kafka run -i --rm kcat-maint-$(date +%s) \
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

# ============================================================================
# MAINTAIN: Trigger the full dispatch loop
# ============================================================================
# Spawns site-work-orchestrator in maintenance mode.
# The orchestrator's own workflow (from agent_definitions) handles:
#   1. select_style → set_defaults → render_components (site chrome)
#   2. load content items (likely empty for maintenance)
#   3. load ALL triaged fix items (no handler filter)
#   4. fix_items_loop: for each item, spawn→call handler dynamically
#      - handler_agent from work item → spawn_agent (agent_type_field)
#      - call_agent finds by target_role → sends raw identifiers
#      - handler loads its own context, does its work, responds
#   5. apply_site_design → complete

do_maintain() {
    local DOMAIN="$1"
    gen_ids

    echo "========================================="
    echo "Maintenance Dispatch"
    echo "  Domain:  $DOMAIN"
    echo "  Mode:    maintenance (skip planning)"
    echo "  CorrID:  $CORRELATION_ID"
    echo "========================================="

    send_to_generic "$(cat <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"demo_client","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_orchestrator","processing_mode":"orchestrator","timeout_seconds":900,"steps":{"spawn_orchestrator":{"action":"spawn_agent","config":{"role":"site_orchestrator","agent_type":"site-work-orchestrator"},"output_field":"spawn_orchestrator","next_step":"call_orchestrator","description":"Spawn site work orchestrator"},"call_orchestrator":{"action":"call_agent","config":{"target_role":"site_orchestrator","input_mapping":{"domain":"input_data.domain","mode":"input_data.mode"},"timeout_seconds":600},"output_field":"orchestrator_result","next_step":"complete","description":"Run maintenance dispatch loop"},"complete":{"action":"complete_workflow","config":{"output_fields":["orchestrator_result"]},"description":"Maintenance complete"}}}},"input_data":{"domain":"${DOMAIN}","mode":"maintenance"}}
JSON
    )"

    echo ""
    echo "Monitor:"
    echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
    echo ""
    echo "Check progress:"
    echo "  $0 status $DOMAIN"
    echo ""
    echo "CORRELATION_ID=$CORRELATION_ID"
}

# ============================================================================
# TRIAGE: Move detected items to triaged
# ============================================================================
# Discovery agents write items with status='detected'.
# The dispatch loop only processes items with status='triaged'.
# This step approves them for processing.
#
# In the fully automated system, triage runs as a step within the
# orchestrator itself (auto-triage based on rules + severity).
# For now it's manual.

do_triage() {
    local DOMAIN="$1"

    echo "Current detected items for $DOMAIN:"
    $PSQL_CMD -c "
        SELECT item_type, handler_agent, severity, priority,
               spec->>'purpose' as purpose,
               summary
        FROM site_work_items
        WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN')
          AND status = 'detected'
        ORDER BY priority, created_at;
    "

    echo ""
    echo "Triaging..."
    $PSQL_CMD -c "
        UPDATE site_work_items
        SET status = 'triaged', updated_at = NOW()
        WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN')
          AND status = 'detected'
        RETURNING item_type, handler_agent, priority, spec->>'purpose' as purpose;
    "
}

# ============================================================================
# STATUS: Check work item status for a site
# ============================================================================

do_status() {
    local DOMAIN="$1"

    echo "Work items for $DOMAIN:"
    $PSQL_CMD -c "
        SELECT item_type, handler_agent, status, priority,
               spec->>'purpose' as purpose,
               spec->>'asset_id' as asset_id,
               result->>'commit_sha' as commit_sha,
               updated_at::timestamp(0) as updated
        FROM site_work_items
        WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN')
        ORDER BY priority, created_at;
    "

    echo ""
    echo "Summary:"
    $PSQL_CMD -c "
        SELECT status, COUNT(*) as count
        FROM site_work_items
        WHERE site_id = (SELECT id FROM sites WHERE domain = '$DOMAIN')
        GROUP BY status
        ORDER BY status;
    "
}

# ============================================================================
# USAGE
# ============================================================================

case "${1:-help}" in
    maintain)
        do_maintain "${2:?Usage: $0 maintain <domain>}"
        ;;
    triage)
        do_triage "${2:?Usage: $0 triage <domain>}"
        ;;
    status)
        do_status "${2:?Usage: $0 status <domain>}"
        ;;
    help|*)
        echo "Usage:"
        echo "  $0 maintain <domain>   — trigger full maintenance dispatch loop"
        echo "  $0 triage <domain>     — approve detected items for processing"
        echo "  $0 status <domain>     — check work item status"
        echo ""
        echo "Typical flow:"
        echo "  ./075_trigger_discovery.sh finetuning.uk design          # 1. find issues"
        echo "  $0 status finetuning.uk                                  # 2. see what was found"
        echo "  $0 triage finetuning.uk                                  # 3. approve items"
        echo "  ./075c_test_handlers... css finetuning.uk                # 4. test handlers individually"
        echo "  ./075c_test_handlers... asset finetuning.uk hero         # 5. test another handler"
        echo "  $0 maintain finetuning.uk                                # 6. or run the full loop"
        echo ""
        echo "Future (automated):"
        echo "  K8s CronJob → maintenance-batch-scheduler"
        echo "    → finds sites with due maintenance or triaged items"
        echo "    → spawns site-work-orchestrator per site"
        echo "    → orchestrator runs: discovery → triage → dispatch → verify"
        ;;
esac