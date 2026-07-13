#!/usr/bin/env bash
# 083_regenerate_brief-explanation_vonc.sh
#
# Supersedes 082. Two runs pinned the requirement:
#   081 (all fields TOP-LEVEL): call_agent contract PASSED, but component-creator's
#       workflow reads input_data.spec.* → empty → generated a generic 'general-hero'.
#   082 (all fields NESTED under spec): call_agent contract FAILED:
#       "contract violation ... missing required fields: [section_type].
#        Provided fields: [domain site_id spec]" — component-creator never ran.
#
# So call_agent extracts via input_mapping, validates the target's input_contract
# (which requires section_type TOP-LEVEL), THEN invokes. But the workflow reads
# input_data.spec.* (the work-item convention spec_data -> input_data.spec). The two
# read fields in DIFFERENT places, so we must provide BOTH:
#   - section_type TOP-LEVEL  → satisfies the input_contract
#   - the full `spec` object  → satisfies the workflow (prompt + store step)
# All input_mapping SOURCES are one-level (input_data.X), which the 082 log proved
# resolve (component-creator received input_data.spec intact).
#
# Expected: component-creator receives section_type (top-level) + spec.* (full),
# the LLM gets the real section_type + description, emits function='brief-explanation',
# and store_generated_component UPDATES the existing row (58363894-...) in place
# (status 'regenerated'). Verify afterward.
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Regenerate brief-explanation (contract+spec)  site: vonc.com"
echo "  Correlation: $CORRELATION_ID"
echo "========================================="
echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-regen-be3-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_creator","processing_mode":"orchestrator","timeout_seconds":180,"steps":{"spawn_creator":{"action":"spawn_agent","config":{"role":"component_creator","agent_type":"component-creator"},"output_field":"creator_agent","next_step":"call_creator","description":"Spawn component-creator agent"},"call_creator":{"action":"call_agent","config":{"agent_type":"component-creator","target_role":"component_creator","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain","section_type":"input_data.section_type","spec":"input_data.spec"},"timeout_seconds":150},"output_field":"creator_result","next_step":"complete","description":"Generate and store (regenerate in place) the component"},"complete":{"action":"complete_workflow","config":{"output_fields":["creator_result"]},"description":"Component regeneration complete"}}}},"input_data":{"site_id":"9ec3b9ee-5b08-461b-b4f8-9e1e03579c74","domain":"vonc.com","section_type":"brief-explanation","spec":{"section_type":"brief-explanation","site_type":"content","page_context":"index — the Spark landing page; explains how the daily provocation game works","description":"A concise how-it-works explainer section for the Spark landing page that tells a first-time visitor what the daily game is and how to take part. Two-column layout: a text column beside an illustrative image. The text column has a small eyebrow label; a short heading with ONE emphasised word or phrase; a one or two sentence description; then an ordered list of EXACTLY THREE numbered steps, each step a short bold title plus a single explanatory sentence describing the daily flow (a provocation appears, you take your position, the Gauntlet scores the room); then a row of EXACTLY THREE small stats (a value and a label each); then two call-to-action buttons (a primary and a secondary). The image column shows the site illustration with a small badge label overlaid. This is a DARK section. Every copy field carrying the site voice (eyebrow, heading, description, the three step titles and their sentences, the badge, and both CTA labels) must be a content-writer placeholder; the image comes from the site illustration asset; the stat values and labels are tunable labels with sensible fallbacks. The component function name MUST be brief-explanation.","design_direction":"Dark, high-energy, game-like — an arena explaining its rules, not a corporate features block. Confident and punchy. Use the site CSS variables only (--color-primary, --color-accent, --color-text, --color-heading, etc.); this is a dark section so set the section --section-* variables and use them for text. Generous vertical spacing; clear numbered-step rhythm; strong but uncluttered."}}}
JSON

echo ""
echo "Monitor:  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=300 | grep \"$CORRELATION_ID\""
echo "Expect:   NO contract violation; component-creator generate_template runs with"
echo "          SECTION TYPE: brief-explanation; store_generated_component: regeneration"
echo "          — existing component found; component regenerated (function=brief-explanation)."
echo ""
echo "Verify: brief-explanation row 58363894-... updated in place (status regenerated,"
echo "        active_rows=1, schema_field_count>0, has_no_value=f, has_placeholders=t),"
echo "        plus any rebuild items raised (query 3)."

SELECT id, function, section_type, is_active, quality_score,
         template_variable_count, schema_field_count,
         LENGTH(html_template) AS tmpl_len,
          LENGTH(COALESCE(js_content,'')) AS js_len,
          (html_template LIKE '%<no value>%')     AS has_no_value,
          (html_template LIKE '%{{placeholder %') AS has_placeholders,
          created_at, updated_at
   FROM content_components WHERE function = 'brief-explanation'
   ORDER BY is_active DESC, updated_at DESC;

   SELECT COUNT(*) AS active_rows FROM content_components
   WHERE function='brief-explanation' AND is_active=true AND forked_from IS NULL;

   SELECT item_type, item_key, status, created_at FROM site_work_items
   WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid
    AND created_at > NOW() - INTERVAL '15 minutes'
   ORDER BY created_at DESC;