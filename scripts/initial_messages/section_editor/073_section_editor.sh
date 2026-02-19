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
EDIT_MODE="${1:-content_edit}"




DOMAIN="leopardessconsulting.co.uk"
#EDIT_MODE="content_edit"
EDIT_MODE="replace"

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
          "edit_type": "input_data.edit_type",
          "page_name?": "input_data.page_name",
          "slot_name?": "input_data.slot_name",
          "field_updates?": "input_data.field_updates",
          "replacement_content_data?": "input_data.content_data",
          "new_component_function?": "input_data.new_component_function",
          "page_component_id?": "input_data.page_component_id"
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




---


-- 1. Parent orchestration state - what did spawn return?
SELECT orchestration_id, status, current_step, error,
       jsonb_pretty(collected_data->'section_editor_agent') as spawn_result,
       jsonb_pretty(collected_data->'input_data') as input_data
FROM orchestration_states
WHERE correlation_id = '0d137594-1376-494b-853b-c0467f2a80df'::uuid;

-- 2. Any child orchestrations?
SELECT orchestration_id, owner_agent_type, status, current_step, error,
       requests_topic, responses_topic
FROM orchestration_states
WHERE parent_orchestration_id = 'ea89edd1-86ef-4f21-bc00-181266daeeae';

-- 3. All awaited requests for this correlation
SELECT request_id, step_name, target_agent_type,
       requests_topic, responses_topic, status,
       sent_at, timeout_at, processed_at
FROM awaited_requests
WHERE correlation_id = '0d137594-1376-494b-853b-c0467f2a80df';


-- 1. What component is linked to index/hero? Is it the right one?
SELECT pc.id, pc.slot_name, pc.component_id,
       cc.function, cc.name,
       LEFT(cc.html_template, 300) as template_start
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
LEFT JOIN content_components cc ON pc.component_id = cc.id
WHERE p.name = 'index'
  AND p.site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
  AND pc.slot_name = 'hero';

-- 2. What field names does this hero template actually use?
--    (look for {{.something}} patterns)
SELECT cc.function,
       regexp_matches(cc.html_template, '\{\{\.([a-zA-Z_]+)\}\}', 'g') as template_fields
FROM content_components cc
JOIN page_components pc ON pc.component_id = cc.id
JOIN pages p ON pc.page_id = p.id
WHERE p.name = 'index'
  AND p.site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
  AND pc.slot_name = 'hero';

-- 3. What content_data does the hero have now (after the edit)?
SELECT pc.content_data
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
WHERE p.name = 'index'
  AND p.site_id = (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk')
  AND pc.slot_name = 'hero';

-- 4. What was the ORIGINAL home page hero component?
--    (check if there's a different hero for index vs other pages)
SELECT cc.id, cc.function, cc.name, LEFT(cc.html_template, 200) as preview
FROM content_components cc
WHERE cc.function LIKE 'hero%'
  AND cc.is_active = true
ORDER BY cc.function;

