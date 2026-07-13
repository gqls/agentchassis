#!/usr/bin/env bash
# 081_regenerate_brief-explanation_vonc.sh
#
# Regenerate the `brief-explanation` SECTION component IN PLACE via
# component-creator, in the context of vonc.com (Spark).
#
# WHY: brief-explanation is a Mode-B empty shell (html_template full of bare
# `<no value>`, empty input_schema, 0 slots, quality 50). It is STATIC content,
# so the fix is REGENERATION (fresh template + real schema the content writer
# can fill) — NOT a JS loader. component-creator is the only generation path
# (calls store_generated_component → separateInlineJS). A row already exists for
# function='brief-explanation', so store_generated_component takes its
# UPDATE-in-place branch: same component_id (no FK relink), snapshots the old
# row to component_versions, reactivates, raises rerender work item(s).
#
# Pattern: spawn_agent + call_agent via the generic entry point — same as the
# 080 component-creator test trigger, embedding the JSON literally (no jq dep).
# DELIBERATE ADDITIONS over 080 (flagged): site_id + domain in input_data AND
# the input_mapping so ensure_site_record/read_site_spec resolve vonc; plus a
# structure-focused description + Spark design_direction.
#
# DETERMINISM NOTE: store_generated_component looks up the existing row by the
# LLM's emitted `function`, not by any spec field. section_type='brief-explanation'
# makes the LLM mirror it to function='brief-explanation' (reliable for a known
# name) but it is NOT pinned — VERIFY afterward (status 'regenerated', same
# component_id, exactly one active row). A second row = duplicate INSERT; deactivate
# and re-run.
#
# AFTER REGEN — CONTENT FILL: the new template has new placeholders with no
# content_data yet. The auto-raised needs_rerender only ASSEMBLES stored HTML, so
# the section may show fallbacks until the index is rebuilt by the content writer.
# Expect a follow-up needs_page rebuild for the index; we check what items were
# created before deciding.
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Regenerate component: brief-explanation  (site: vonc.com)"
echo "  Correlation: $CORRELATION_ID"
echo "========================================="
echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-regen-be-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_creator","processing_mode":"orchestrator","timeout_seconds":180,"steps":{"spawn_creator":{"action":"spawn_agent","config":{"role":"component_creator","agent_type":"component-creator"},"output_field":"creator_agent","next_step":"call_creator","description":"Spawn component-creator agent"},"call_creator":{"action":"call_agent","config":{"agent_type":"component-creator","target_role":"component_creator","input_mapping":{"site_id":"input_data.site_id","domain":"input_data.domain","section_type":"input_data.section_type","site_type":"input_data.site_type","page_context":"input_data.page_context","description":"input_data.description","design_direction":"input_data.design_direction"},"timeout_seconds":150},"output_field":"creator_result","next_step":"complete","description":"Generate and store (regenerate in place) the component"},"complete":{"action":"complete_workflow","config":{"output_fields":["creator_result"]},"description":"Component regeneration complete"}}}},"input_data":{"site_id":"9ec3b9ee-5b08-461b-b4f8-9e1e03579c74","domain":"vonc.com","section_type":"brief-explanation","site_type":"content","page_context":"index — the Spark landing page; explains how the daily provocation game works","description":"A concise how-it-works explainer section for the Spark landing page that tells a first-time visitor what the daily game is and how to take part. Two-column layout: a text column beside an illustrative image. The text column has a small eyebrow label; a short heading with ONE emphasised word or phrase; a one or two sentence description; then an ordered list of EXACTLY THREE numbered steps, each step a short bold title plus a single explanatory sentence describing the daily flow (a provocation appears, you take your position, the Gauntlet scores the room); then a row of EXACTLY THREE small stats (a value and a label each); then two call-to-action buttons (a primary and a secondary). The image column shows the site illustration with a small badge label overlaid. This is a DARK section. Every copy field carrying the site voice (eyebrow, heading, description, the three step titles and their sentences, the badge, and both CTA labels) must be a content-writer placeholder; the image comes from the site illustration asset; the stat values and labels are tunable labels with sensible fallbacks.","design_direction":"Dark, high-energy, game-like — an arena explaining its rules, not a corporate features block. Confident and punchy. Use the site CSS variables only (--color-primary, --color-accent, --color-text, --color-heading, etc.); this is a dark section so set the section --section-* variables and use them for text. Generous vertical spacing; clear numbered-step rhythm; strong but uncluttered."}}
JSON

echo ""
echo "========================================="
echo "Submitted. Monitor + verify (queries printed below)."
echo "========================================="
cat <<'VERIFY'

# 1) Watch the run:
#   kubectl -n ai-persona-system logs -f -l app=agent-chassis --tail=300 | grep "<CORRELATION_ID>"
# Look for:
#   store_generated_component: pre-store validation passed
#   store_generated_component: regeneration — existing component found
#   store_generated_component: component regenerated     (NOT "component created")
# A "rejecting low-quality template" with a <no value> blocker => the LLM produced
# a broken template; inspect and re-run (do NOT hand-edit).

# 2) Confirm IN-PLACE update (same id as baseline, exactly one active row,
#    real schema, no <no value>):
#   SELECT id, function, section_type, is_active, quality_score,
#          template_variable_count, schema_field_count,
#          LENGTH(html_template) AS tmpl_len,
#          LENGTH(COALESCE(js_content,'')) AS js_len,
#          (html_template LIKE '%<no value>%')     AS has_no_value,
#          (html_template LIKE '%{{placeholder %') AS has_placeholders
#   FROM content_components WHERE function = 'brief-explanation'
#   ORDER BY is_active DESC, updated_at DESC;
#
#   SELECT COUNT(*) AS active_rows FROM content_components
#   WHERE function='brief-explanation' AND is_active=true AND forked_from IS NULL;
#   -- MUST be 1. If 2, a duplicate was INSERTed; deactivate the new dupe + re-run.

# 3) See what rebuild items the regen raised (decides whether we also need a
#    needs_page content-fill for the index):
#   SELECT item_type, item_key, status, created_at FROM site_work_items
#   WHERE site_id = '9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid
#     AND created_at > NOW() - INTERVAL '15 minutes'
#   ORDER BY created_at DESC;
VERIFY
echo ""
