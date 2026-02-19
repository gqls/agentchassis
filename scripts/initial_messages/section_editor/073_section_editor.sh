#!/bin/bash
# =============================================================================
# Section Editor Trigger
# =============================================================================
# Sends a section edit request via the generic agent, which spawns and calls
# the section-editor agent. The section-editor runs self-contained with its
# own workflow (load context → apply edit → git commit → deploy).
#
# Usage:
#   ./trigger_section_editor.sh content_edit   # field_updates merge
#   ./trigger_section_editor.sh replace        # full content_data replace
#   ./trigger_section_editor.sh swap           # component swap
# =============================================================================

DOMAIN="leopardessconsulting.co.uk"
EDIT_MODE="${1:-content_edit}"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

# --- Build input_data based on edit mode ---
case "$EDIT_MODE" in
  content_edit)
    # Merge specific fields into existing content_data
    INPUT_DATA=$(cat <<ENDJSON
{
  "domain": "${DOMAIN}",
  "page_name": "index",
  "slot_name": "hero",
  "edit_type": "content_edit",
  "field_updates": {
    "headline": "Strategic Consulting for Growth",
    "subheadline": "Helping businesses scale with clarity and confidence"
  }
}
ENDJSON
)
    ;;
  replace)
    # Full content_data replacement (e.g. rewrite a case study section)
    INPUT_DATA=$(cat <<ENDJSON
{
  "domain": "${DOMAIN}",
  "page_name": "use-cases",
  "slot_name": "case-studies-list",
  "edit_type": "content_edit",
  "content_data": {
    "section_title": "How We Help",
    "section_subtitle": "Real results for real businesses",
    "cases": [
      {
        "title": "Digital Transformation",
        "description": "Helping companies modernise their technology and processes",
        "outcome": "Streamlined operations across the organisation"
      },
      {
        "title": "Process Optimisation",
        "description": "Identifying bottlenecks and implementing efficient workflows",
        "outcome": "Reduced operational costs and improved delivery times"
      }
    ]
  }
}
ENDJSON
)
    ;;
  swap)
    # Component swap — different template, same content
    INPUT_DATA=$(cat <<ENDJSON
{
  "domain": "${DOMAIN}",
  "page_name": "index",
  "slot_name": "social-proof",
  "edit_type": "component_swap",
  "new_component_function": "testimonials-grid"
}
ENDJSON
)
    ;;
  *)
    echo "Unknown edit mode: $EDIT_MODE"
    echo "Usage: $0 [content_edit|replace|swap]"
    exit 1
    ;;
esac

# Compact the input data (remove newlines for Kafka)
INPUT_DATA_COMPACT=$(echo "$INPUT_DATA" | tr -d '\n' | sed 's/  */ /g')

echo "========================================="
echo "Section Editor ($EDIT_MODE)"
echo "========================================="
echo "  Domain:         $DOMAIN"
echo "  Edit mode:      $EDIT_MODE"
echo "  Correlation ID: $CORRELATION_ID"
echo "========================================="

# --- Build the inline workflow that spawns and calls section-editor ---
WORKFLOW=$(cat <<'ENDWF'
{
  "start_step": "spawn_section_editor",
  "processing_mode": "orchestrator",
  "timeout_seconds": 900,
  "steps": {
    "spawn_section_editor": {
      "action": "spawn_agent",
      "config": {
        "role": "section_editor",
        "agent_type": "section-editor"
      },
      "output_field": "section_editor_agent",
      "next_step": "call_section_editor",
      "description": "Spawn section-editor agent"
    },
    "call_section_editor": {
      "action": "call_agent",
      "config": {
        "agent_type": "section-editor",
        "target_role": "section_editor",
        "input_mapping": {
          "domain": "input_data.domain",
          "page_name": "input_data.page_name",
          "slot_name": "input_data.slot_name",
          "edit_type": "input_data.edit_type",
          "field_updates": "input_data.field_updates",
          "content_data": "input_data.content_data",
          "new_component_function": "input_data.new_component_function"
        },
        "timeout_seconds": 600
      },
      "output_field": "edit_result",
      "next_step": "complete",
      "description": "Run section edit"
    },
    "complete": {
      "action": "complete_workflow",
      "config": {
        "output_fields": ["edit_result"]
      },
      "description": "Section edit complete"
    }
  }
}
ENDWF
)

# Compact the workflow JSON
WORKFLOW_COMPACT=$(echo "$WORKFLOW" | tr -d '\n' | sed 's/  */ /g')

# --- Build the full message body ---
MESSAGE_BODY=$(cat <<ENDMSG
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":${WORKFLOW_COMPACT}},"input_data":${INPUT_DATA_COMPACT}}
ENDMSG
)

kubectl -n kafka run -i --rm kcat-section-editor-$(date +%s) \
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
  -H client_id=$CLIENT_ID \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<ENDKAFKA
${MESSAGE_BODY}
ENDKAFKA

echo ""
echo "========================================="
echo "Section editor triggered ($EDIT_MODE)"
echo "========================================="
echo ""
echo "Monitor with:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=50 | grep '$CORRELATION_ID'"
echo ""
echo "CORRELATION_ID=$CORRELATION_ID"


