#!/bin/bash
# ============================================================================
# COMPONENT CREATOR — Test Trigger
# ============================================================================
# Tests the component-creator agent by asking it to generate a simple
# info-card component. Uses the same spawn+call pattern as other triggers.
#
# After running, verify with:
#   SELECT id, function, section_type, display_name, category, created_from,
#          LENGTH(html_template) as template_len
#   FROM content_components
#   WHERE section_type = 'info-card' AND created_from = 'generated';
#
# Clean up test data:
#   UPDATE content_components SET is_active = false
#   WHERE section_type = 'info-card' AND created_from = 'generated';
# ============================================================================

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Component Creator Test"
echo "========================================="
echo "  Section type: info-card"
echo "  Correlation:  ${CORRELATION_ID}"
echo "========================================="
echo ""

kubectl -n kafka run -i --rm kcat-comp-creator-$(date +%s) \
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
  -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_creator","processing_mode":"orchestrator","timeout_seconds":120,"steps":{"spawn_creator":{"action":"spawn_agent","config":{"role":"component_creator","agent_type":"component-creator"},"output_field":"creator_agent","next_step":"call_creator","description":"Spawn component-creator agent"},"call_creator":{"action":"call_agent","config":{"agent_type":"component-creator","target_role":"component_creator","input_mapping":{"section_type":"input_data.section_type","site_type":"input_data.site_type","page_context":"input_data.page_context","description":"input_data.description","design_direction":"input_data.design_direction"},"timeout_seconds":90},"output_field":"creator_result","next_step":"complete","description":"Generate and store component"},"complete":{"action":"complete_workflow","config":{"output_fields":["creator_result"]},"description":"Component creation complete"}}}},"input_data":{"section_type":"info-card","site_type":"brochure","page_context":"about","description":"A simple card displaying a heading, a short paragraph of text, and an optional icon. Used for feature highlights or key information points. Light background with subtle border. Should support 2-4 cards in a responsive grid.","design_direction":"Clean, professional, minimal. Cards should have subtle shadows and rounded corners."}}
JSON

echo ""
echo "========================================="
echo "Submitted"
echo "========================================="
echo ""
echo "SAVE: CORRELATION_ID=${CORRELATION_ID}"
echo ""
echo "Monitor:"
echo "  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=100 | grep '${CORRELATION_ID}'"
echo ""
echo "Check result:"
echo "  SELECT id, function, section_type, display_name, category, created_from,"
echo "         LENGTH(html_template) as template_len, input_schema::text"
echo "  FROM content_components"
echo "  WHERE section_type = 'info-card' AND created_from = 'generated';"
echo ""
echo "Verify contracts:"
echo "  SELECT function,"
echo "    html_template LIKE '%data-component=\"' || function || '\"%' as has_data_component,"
echo "    html_template LIKE '%' || function || '-section%' as has_section_class,"
echo "    html_template LIKE '%--color-%' as uses_css_vars,"
echo "    html_template LIKE '%@media%' as has_responsive"
echo "  FROM content_components"
echo "  WHERE section_type = 'info-card' AND created_from = 'generated';"
echo ""
echo "Test selector finds it:"
echo "  SELECT function, section_type,"
echo "    (CASE WHEN suitable_site_types @> to_jsonb('brochure'::text) THEN 0.35 ELSE 0.05 END"
echo "     + COALESCE(avg_quality_score, 0.3) * 0.3"
echo "     + LEAST(COALESCE(usage_count, 0)::float / 50.0, 1.0) * 0.1) as score"
echo "  FROM content_components"
echo "  WHERE section_type = 'info-card' AND is_active = true AND forked_from IS NULL;"
echo ""
echo "Clean up (if needed):"
echo "  UPDATE content_components SET is_active = false"
echo "  WHERE section_type = 'info-card' AND created_from = 'generated';"




------------

{
  "headers": {
    "correlation_id": "${CORRELATION_ID}",
    "orchestration_id": "${ORCHESTRATION_ID}",
    "request_id": "${REQUEST_ID}",
    "message_id": "${MESSAGE_ID}",
    "message_type": "request",
    "client_id": "${CLIENT_ID}",
    "action": "process",
    "sender": {"agent_id": "cli-user", "agent_type": "cli", "pod_name": "cli"},
    "timestamp": "${TIMESTAMP}"
  },
  "config": {
    "workflow": {
      "start_step": "spawn_creator",
      "processing_mode": "orchestrator",
      "timeout_seconds": 120,
      "steps": {
        "spawn_creator": {
          "action": "spawn_agent",
          "config": {
            "role": "component_creator",
            "agent_type": "component-creator"
          },
          "output_field": "creator_agent",
          "next_step": "call_creator",
          "description": "Spawn component-creator agent"
        },
        "call_creator": {
          "action": "call_agent",
          "config": {
            "agent_type": "component-creator",
            "target_role": "component_creator",
            "input_mapping": {
              "section_type": "input_data.section_type",
              "site_type": "input_data.site_type",
              "page_context": "input_data.page_context",
              "description": "input_data.description",
              "design_direction": "input_data.design_direction"
            },
            "timeout_seconds": 90
          },
          "output_field": "creator_result",
          "next_step": "complete",
          "description": "Generate and store component"
        },
        "complete": {
          "action": "complete_workflow",
          "config": {"output_fields": ["creator_result"]},
          "description": "Component creation complete"
        }
      }
    }
  },
  "input_data": {
    "section_type": "info-card",
    "site_type": "brochure",
    "page_context": "about",
    "description": "A simple card displaying a heading, a short paragraph of text, and an optional icon. Used for feature highlights or key information points. Light background with subtle border. Should support 2-4 cards in a responsive grid.",
    "design_direction": "Clean, professional, minimal. Cards should have subtle shadows and rounded corners."
  }
}
