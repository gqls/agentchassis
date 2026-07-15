#!/usr/bin/env bash
# 082_regenerate_brief-explanation_vonc.sh
#
# Supersedes 081. The 081 run produced a STRAY generic 'general-hero' component
# instead of regenerating brief-explanation, because the inputs were delivered at
# input_data.section_type / input_data.description (top-level), but component-creator
# reads input_data.spec.section_type / input_data.spec.description (the work-item
# convention: spec_data -> input_data.spec). So the generate_template LLM prompt got
# an EMPTY section_type/description and defaulted to a generic hero.
#
# FIX: nest the component-creator inputs under a `spec` object and map
# "spec" -> input_data.spec. site_id/domain stay TOP-LEVEL (ensure_site_record reads
# input_data.site_id/domain, not input_data.spec).
#
# Expected result this time: the LLM receives SECTION TYPE: brief-explanation +
# the description, emits function='brief-explanation', and store_generated_component
# takes the UPDATE-IN-PLACE branch on the existing row (id 58363894-...), status
# 'regenerated'. Verify afterward (status, same id, active_rows=1, real schema,
# no <no value>).
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Regenerate brief-explanation (spec-nested)  site: vonc.com"
echo "  Correlation: $CORRELATION_ID"
echo "========================================="
echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-regen-be2-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_creator","processing_mode":"orchestrator","timeout_seconds":180,"steps":{"spawn_creator":{"action":"spawn_agent","config":{"role":"component_creator","agent_type":"component-creator"},"output_field":"creator_agent","next_step":"call_creator","description":"Spawn component-creator agent"},"call_creator":{"action":"call_agent","config":{"agent_type":"component-creator","target_role":"component_creator","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain","spec":"input_data.spec"},"timeout_seconds":150},"output_field":"creator_result","next_step":"complete","description":"Generate and store (regenerate in place) the component"},"complete":{"action":"complete_workflow","config":{"output_fields":["creator_result"]},"description":"Component regeneration complete"}}}},"input_data":{"site_id":"9ec3b9ee-5b08-461b-b4f8-9e1e03579c74","domain":"vonc.com","spec":{"section_type":"brief-explanation","site_type":"content","page_context":"index — the Spark landing page; explains how the daily provocation game works","description":"A concise how-it-works explainer section for the Spark landing page that tells a first-time visitor what the daily game is and how to take part. Two-column layout: a text column beside an illustrative image. The text column has a small eyebrow label; a short heading with ONE emphasised word or phrase; a one or two sentence description; then an ordered list of EXACTLY THREE numbered steps, each step a short bold title plus a single explanatory sentence describing the daily flow (a provocation appears, you take your position, the Gauntlet scores the room); then a row of EXACTLY THREE small stats (a value and a label each); then two call-to-action buttons (a primary and a secondary). The image column shows the site illustration with a small badge label overlaid. This is a DARK section. Every copy field carrying the site voice (eyebrow, heading, description, the three step titles and their sentences, the badge, and both CTA labels) must be a content-writer placeholder; the image comes from the site illustration asset; the stat values and labels are tunable labels with sensible fallbacks. The component function name MUST be brief-explanation.","design_direction":"Dark, high-energy, game-like — an arena explaining its rules, not a corporate features block. Confident and punchy. Use the site CSS variables only (--color-primary, --color-accent, --color-text, --color-heading, etc.); this is a dark section so set the section --section-* variables and use them for text. Generous vertical spacing; clear numbered-step rhythm; strong but uncluttered."}}}
JSON

echo ""
echo "Monitor:  kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=300 | grep \"$CORRELATION_ID\""
echo "Expect:   store_generated_component: regeneration — existing component found"
echo "          store_generated_component: component regenerated   (function=brief-explanation)"
echo ""
echo "Verify (same as before): function='brief-explanation' row updated in place"
echo "(same id 58363894-..., status regenerated, active_rows=1, schema populated,"
echo " has_no_value=f, has_placeholders=t), plus any rebuild items raised."
