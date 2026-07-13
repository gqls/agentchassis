#!/usr/bin/env bash
# 086_section_edit_provocation-card_vonc.sh
#
# Re-render the vonc.com index's provocation-card section FROM ITS TEMPLATE and
# redeploy, via the section-editor agent. Run this AFTER trim_minilobby.sql has
# updated content_components.html_template.
#
# WHY THIS AGENT (and not rerender-index-vonc.sh):
#   rerender_single_page is ASSEMBLE-ONLY — it redeploys the stored rendered_html
#   and would ship the untrimmed markup. rerender_page_sections (the "light path")
#   would work but re-renders ALL SIX sections, and brief-explanation's template is
#   newer than its instance, so it would silently change too. apply_section_edit
#   re-renders exactly ONE page_component from its template, updates rendered_html
#   + content_data, reassembles the page from the other five stored sections, and
#   commits. Blast radius: one section. This is the route
#   fix_component_template_action's header names ("content changes go through the
#   section-editor workflow").
#
# field_updates is a genuine NO-OP merge: _built_at is re-supplied with its exact
# current value, so content_data is bit-for-bit unchanged. A non-empty map is used
# because applyContentEdit errors when neither field_updates nor
# replacement_content_data is present, and an empty {} may not survive
# ExtractActionInputs.
#
# NOTE: section-editor has never run in production (0 rows in orchestration_states).
# This is its first exercise. Backups are taken by trim_minilobby.sql; recovery is
# to restore page_components.rendered_html and run 083_rerender-index-vonc.sh.
# ────────────────────────────────────────────────────────────────────────

set -euo pipefail

SITE_ID="9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"
PAGE_COMPONENT_ID="a757434e-ab8a-4d2d-bfee-0fb6932f140e"   # provocation-card @ index
DOMAIN="vonc.com"
BUILT_AT="2026-07-03T12:56:49Z"                             # current content_data._built_at

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

echo "========================================="
echo "Section edit: provocation-card  ($DOMAIN /index.html)"
echo "  Correlation: $CORRELATION_ID"
echo "========================================="
echo "SAVE: CORRELATION_ID=$CORRELATION_ID"
echo ""

kubectl -n kafka run -i --rm "kcat-section-edit-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_editor","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_editor":{"action":"spawn_agent","config":{"role":"section_editor","agent_type":"section-editor"},"output_field":"editor_agent","next_step":"call_editor","description":"Spawn section-editor"},"call_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"section_editor","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","page_component_id":"input_data.page_component_id","edit_type":"input_data.edit_type","field_updates":"input_data.field_updates"},"timeout_seconds":260},"output_field":"edit_result","next_step":"complete","description":"Re-render provocation-card from its trimmed template, reassemble, deploy"},"complete":{"action":"complete_workflow","config":{"output_fields":["edit_result"]},"description":"Section edit complete"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","page_component_id":"${PAGE_COMPONENT_ID}","edit_type":"content_edit","field_updates":{"_built_at":"${BUILT_AT}"}}}
JSON

echo ""
echo "Watch:"
echo "  SELECT status, current_step FROM orchestration_states WHERE correlation_id = '$CORRELATION_ID'::uuid;"
echo ""
echo "Verify (artifact, never item status):"
cat <<'SQL'
  SELECT length(rendered_html)                      AS rendered_len_expect_6488,
         (rendered_html LIKE '%pc-card%')           AS pc_card_gone_expect_f,
         (rendered_html LIKE '%<script%')           AS script_gone_expect_f,
         (rendered_html LIKE '%data-runtime-fill%') AS marker_kept_expect_t
  FROM page_components WHERE id = 'a757434e-ab8a-4d2d-bfee-0fb6932f140e'::uuid;
SQL
echo ""
echo "  curl -s https://vonc.com/index.html | grep -c 'pc-card'            # expect 0"
echo "  curl -s https://vonc.com/index.html | grep -o 'data-component=\"[^\"]*\"'  # expect the same six"
