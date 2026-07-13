#!/usr/bin/env bash
# 084_create_provocations-archive-list_vonc.sh
#
# Creates the provocations-archive-list component via component-creator, using the
# PROVEN 083 dual-placement pattern: section_type BOTH top-level (satisfies the
# call_agent input_contract check) AND inside the spec object (satisfies the workflow's
# input_data.spec.* reads). Works whether or not the ValidateInputContract patch is live.
# Spec: SPEC_provocations-archive-list.md. Function name pinned in the description.
# Generation bakes in: data-runtime-fill on the section tag; NO inline script;
# llm header fields; ONE hidden clone-template archive item; visible empty state.
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Create provocations-archive-list  site: vonc.com"
echo "  Correlation: $CORRELATION_ID"
echo "========================================="
echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-create-pal-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H "correlation_id=$CORRELATION_ID" \
  -H "orchestration_id=$ORCHESTRATION_ID" \
  -H "request_id=$REQUEST_ID" \
  -H "message_id=$MESSAGE_ID" \
  -H "message_type=request" \
  -H "client_id=$CLIENT_ID" \
  -H "action=process" \
  -H "sender_agent_type=cli" \
  -H "sender_agent_id=cli-user" \
  -H "responses_topic=system.agent.generic.responses" \
  -H "timestamp=$TIMESTAMP" <<JSON
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_creator","processing_mode":"orchestrator","timeout_seconds":180,"steps":{"spawn_creator":{"action":"spawn_agent","config":{"role":"component_creator","agent_type":"component-creator"},"output_field":"creator_agent","next_step":"call_creator","description":"Spawn component-creator agent"},"call_creator":{"action":"call_agent","config":{"agent_type":"component-creator","target_role":"component_creator","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain","section_type":"input_data.section_type","spec":"input_data.spec"},"timeout_seconds":150},"output_field":"creator_result","next_step":"complete","description":"Generate and store the archive-list component"},"complete":{"action":"complete_workflow","config":{"output_fields":["creator_result"]},"description":"Component creation complete"}}}},"input_data":{"site_id":"9ec3b9ee-5b08-461b-b4f8-9e1e03579c74","domain":"vonc.com","section_type":"provocations-archive-list","spec":{"section_type":"provocations-archive-list","site_type":"content","page_context":"provocations-index — the Provocations Archive page at /provocations/index.html; the destination of the site's primary CTAs","description":"A self-contained archive section for the Provocations Archive page. The top-level section element MUST have the class provocations-archive, the attribute data-component with the value provocations-archive-list, and the attribute data-runtime-fill with the value true. Inside the section, a header: a small eyebrow label with class provocations-archive__eyebrow; a heading with class provocations-archive__title containing ONE emphasised word or phrase inside an em element; a one or two sentence subtitle with class provocations-archive__subtitle. Below the header, a list container with class provocations-archive__list containing EXACTLY ONE hidden clone-template item: an anchor element with class provocations-archive__item, the attribute data-archive-template, and the hidden attribute, containing a date element with class provocations-archive__item-date, a title element with class provocations-archive__item-title, a one-line teaser with class provocations-archive__item-teaser, and a stat element with class provocations-archive__item-stat that holds a small dot span with class provocations-archive__item-dot followed by a text span for the stat value. After the list, an empty-state line with class provocations-archive__empty that is visible by default. Then one call-to-action button linking back to the day's provocation. Every copy field carrying the site voice, namely the eyebrow, the heading, the subtitle, the empty-state line, and the CTA label, must be a content-writer placeholder; the CTA url is a tunable with fallback /index.html. This is a DARK section using the site CSS variables only. Do NOT include any script element anywhere in this component — the list is populated at runtime by an external loader; the hidden template item and the empty state are the entire built-in behaviour. The component function name MUST be provocations-archive-list.","design_direction":"A record room adjacent to the arena — dark, composed, legible. Generous vertical rhythm; archive items as quiet rows that light on hover using CSS only; the featured energy belongs to the index, not here. Use the site CSS variables only, for example --color-primary, --color-accent, --color-heading, --color-text, and set the section-level --section-* variables for the dark surface."}}}
JSON

echo ""
echo "Monitor:  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=300 | grep \"$CORRELATION_ID\""
echo "Expect:   NO contract violation; generate_template with SECTION TYPE provocations-archive-list;"
echo "          store_generated_component: component created (function=provocations-archive-list, status created)."
echo ""
echo "Verify (runbook Step 1 query): one active row, has_marker=t, has_inline_script=f, schema_field_count>0."
