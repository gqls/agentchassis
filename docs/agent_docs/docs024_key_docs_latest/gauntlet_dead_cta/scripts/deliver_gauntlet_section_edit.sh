#!/usr/bin/env bash
# deliver_gauntlet_section_edit.sh
#
# Re-render vonc.com /tools/gauntlet/index.html's gauntlet-interface section FROM
# ITS (now honest) TEMPLATE and redeploy, via the section-editor agent.
# Run AFTER update_component.sql has rewritten content_components.html_template +
# js_content (dead CTAs removed, fabricated stats/leaderboard stripped, rules card
# added, primary CTA wired to start the challenge).
#
# WHY section-editor: the page is rebuild_policy='owned' (every tool page is).
# The generic rerender->save_page_sections path is HARD-REFUSED for owned pages
# (bugs_closed/024). apply_section_edit (content_edit) is the sanctioned path: it
# re-renders exactly ONE page_component from its current template, merges
# field_updates into content_data, reassembles the page from the other stored
# sections, and commits. Blast radius: one section.
#
# field_updates carries the honest copy (softened hero, the rules list) AND is
# the required non-empty merge applyContentEdit needs.
# Backups: backup_template.html / backup_js.js in the scratchpad; rollback =
# restore content_components + re-run this script.
set -euo pipefail

SITE_ID="9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"
PAGE_COMPONENT_ID="1048b344-f1fa-44ea-b936-951bc7eafc59"   # gauntlet-interface @ tool-gauntlet
DOMAIN="vonc.com"

CORRELATION_ID=$(cat /proc/sys/kernel/random/uuid)
ORCHESTRATION_ID=$(cat /proc/sys/kernel/random/uuid)
REQUEST_ID=$(cat /proc/sys/kernel/random/uuid)
MESSAGE_ID=$(cat /proc/sys/kernel/random/uuid)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
CLIENT_ID="demo_client"

FIELD_UPDATES='{"eyebrow_label":"TODAY'"'"'S GAUNTLET","hero_subtitle":"One provocation. One clock. A test of whether your take holds up under pressure.","rules_title":"How the Gauntlet works","rule_1":"Read today'"'"'s challenge and its three objectives.","rule_2":"Start the clock, then work each objective, ticking them off as you go.","rule_3":"Meet all three before the timer runs out to complete the Gauntlet.","rules_footnote":"A self-paced test of conviction - no sign-up, nothing is scored or shared."}'

echo "SAVE: CORRELATION_ID=$CORRELATION_ID"

kubectl -n kafka run -i --rm "kcat-gauntlet-edit-$(date +%s)" \
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
{"headers":{"correlation_id":"${CORRELATION_ID}","orchestration_id":"${ORCHESTRATION_ID}","request_id":"${REQUEST_ID}","message_id":"${MESSAGE_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_editor","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_editor":{"action":"spawn_agent","config":{"role":"section_editor","agent_type":"section-editor"},"output_field":"editor_agent","next_step":"call_editor","description":"Spawn section-editor"},"call_editor":{"action":"call_agent","config":{"agent_type":"section-editor","target_role":"section_editor","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","page_component_id":"input_data.page_component_id","edit_type":"input_data.edit_type","field_updates":"input_data.field_updates"},"timeout_seconds":260},"output_field":"edit_result","next_step":"complete","description":"Re-render gauntlet-interface from its honest template, reassemble, deploy"},"complete":{"action":"complete_workflow","config":{"output_fields":["edit_result"]},"description":"Section edit complete"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","page_component_id":"${PAGE_COMPONENT_ID}","edit_type":"content_edit","field_updates":${FIELD_UPDATES}}}
JSON

echo ""
echo "Watch:  SELECT status,current_step FROM orchestration_states WHERE correlation_id='$CORRELATION_ID'::uuid;"
